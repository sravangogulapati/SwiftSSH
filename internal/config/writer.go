package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// IsKnownHost returns true if any host in the list has the given hostname.
func IsKnownHost(hosts []Host, hostname string) bool {
	for _, h := range hosts {
		if h.Hostname == hostname {
			return true
		}
	}
	return false
}

// buildHostBlock serializes a Host to its SSH config text block.
// If h has groups, a magic comment is prepended.
func buildHostBlock(h Host) string {
	var b strings.Builder

	if len(h.Groups) > 0 {
		fmt.Fprintf(&b, "# @group %s\n", strings.Join(h.Groups, ", "))
	}

	fmt.Fprintf(&b, "Host %s\n", h.Alias)
	fmt.Fprintf(&b, "    Hostname %s\n", h.Hostname)

	if h.User != "" {
		fmt.Fprintf(&b, "    User %s\n", h.User)
	}

	if h.Port != "" && h.Port != "22" {
		fmt.Fprintf(&b, "    Port %s\n", h.Port)
	}

	if h.IdentityFile != "" {
		fmt.Fprintf(&b, "    IdentityFile \"%s\"\n", h.IdentityFile)
	}

	return b.String()
}

// AppendHost appends a new host block to the SSH config file.
// It first backs up the config file, then appends the new host block.
func AppendHost(configPath, backupPath string, h Host) error {
	// Read the original config file
	original, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Write backup (even if original doesn't exist, backup will be empty)
	if err := os.WriteFile(backupPath, original, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Open config file for appending
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config for appending: %w", err)
	}
	defer f.Close()

	sep := "\n"
	if len(original) == 0 {
		sep = ""
	}
	_, err = fmt.Fprintf(f, "%s%s", sep, buildHostBlock(h))
	if err != nil {
		return fmt.Errorf("failed to write host block: %w", err)
	}

	return nil
}

// ReplaceHostBlock replaces the host block identified by h.LineStart and h.SourceFile
// with a freshly serialized block built from h.
// It writes a backup to h.SourceFile+".bak" before modifying the file.
// Returns (newLineStart, lineDelta, error):
//   - newLineStart: the new 1-based line number of the Host directive in the updated file.
//   - lineDelta: how many lines the block grew (+) or shrank (-) relative to the original.
func ReplaceHostBlock(h Host) (int, int, error) {
	if h.LineStart == 0 {
		return 0, 0, fmt.Errorf("ReplaceHostBlock: LineStart is 0, cannot locate host block")
	}

	// Read all lines
	raw, err := os.ReadFile(h.SourceFile)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read config: %w", err)
	}

	lines := splitLines(raw)

	// Write backup
	backupPath := h.SourceFile + ".bak"
	if err := os.WriteFile(backupPath, raw, 0600); err != nil {
		return 0, 0, fmt.Errorf("failed to write backup: %w", err)
	}

	blockStart := h.LineStart - 1 // convert to 0-based

	if blockStart < 0 || blockStart >= len(lines) {
		return 0, 0, fmt.Errorf("LineStart %d is out of range (file has %d lines)", h.LineStart, len(lines))
	}

	// Verify the line still has "Host <alias>".
	// Lenient: if LineStart points to a @group comment instead of the Host line
	// (e.g. parser off-by-one or drift after a previous save), look one line ahead.
	firstWord, _ := parseHostLine(lines[blockStart])
	if !strings.EqualFold(firstWord, "host") {
		if strings.Contains(lines[blockStart], "@group") && blockStart+1 < len(lines) {
			nextWord, _ := parseHostLine(lines[blockStart+1])
			if strings.EqualFold(nextWord, "host") {
				blockStart++ // advance past the mispointed magic comment
			} else {
				return 0, 0, fmt.Errorf("stale LineStart %d: expected 'Host' directive, got %q", h.LineStart, lines[blockStart])
			}
		} else {
			return 0, 0, fmt.Errorf("stale LineStart %d: expected 'Host' directive, got %q", h.LineStart, lines[blockStart])
		}
	}

	// Determine if there's a magic comment line just before the block
	magicStart := blockStart
	if blockStart > 0 && strings.Contains(lines[blockStart-1], "@group") {
		magicStart = blockStart - 1
	}

	// Find the end of this host block
	blockEnd := findBlockEnd(lines, blockStart)

	// Build new block lines
	newBlock := buildHostBlock(h)
	newBlockLines := splitLines([]byte(newBlock))

	// Reconstruct file: before + new block + after
	result := make([]string, 0, magicStart+len(newBlockLines)+(len(lines)-blockEnd))
	result = append(result, lines[:magicStart]...)
	result = append(result, newBlockLines...)
	result = append(result, lines[blockEnd:]...)

	// Join and write atomically
	output := strings.Join(result, "\n")
	// Preserve trailing newline: if original ended with newline, ensure result does too
	if len(raw) > 0 && raw[len(raw)-1] == '\n' && !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	tmpPath := h.SourceFile + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(output), 0600); err != nil {
		return 0, 0, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, h.SourceFile); err != nil {
		return 0, 0, fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Compute the new 1-based LineStart of the Host directive in the written file.
	// magicStart is the 0-based index of the block's first line in the result.
	newLineStart := magicStart + 1 // 1-based; Host line when no groups
	if len(h.Groups) > 0 {
		newLineStart++ // Host line is one below the magic comment
	}

	// lineDelta: positive means block grew, negative means block shrank.
	oldBlockSize := blockEnd - magicStart
	lineDelta := len(newBlockLines) - oldBlockSize

	return newLineStart, lineDelta, nil
}

// splitLines splits raw bytes into lines, stripping \r for Windows CRLF.
// Each element in the returned slice does NOT include the line terminator.
func splitLines(data []byte) []string {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // scanner.Text() already strips \r\n
	}
	return lines
}

// findBlockEnd returns the index of the first line (0-based) that belongs to the
// NEXT host block after the block starting at blockStart. Returns len(lines) at EOF.
// A line beginning a new block is one whose first non-blank, non-comment token is "host".
// Trailing blank lines between host blocks are NOT included in the current block; they
// fall into the lines[blockEnd:] "after" section so they are preserved on rewrite.
func findBlockEnd(lines []string, blockStart int) int {
	for i := blockStart + 1; i < len(lines); i++ {
		word, _ := parseHostLine(lines[i])
		if strings.EqualFold(word, "host") {
			end := i
			// magic comment belongs to the next block — back up over it first
			if end > blockStart+1 && strings.Contains(lines[end-1], "@group") {
				end--
			}
			// back up past trailing blank lines so they are preserved
			for end > blockStart+1 && strings.TrimSpace(lines[end-1]) == "" {
				end--
			}
			return end
		}
	}
	return len(lines)
}

// parseHostLine returns the first keyword and its value from a config line,
// or ("", "") if the line is blank or a comment.
func parseHostLine(line string) (keyword, value string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", ""
	}
	idx := strings.IndexAny(trimmed, " \t")
	if idx == -1 {
		return trimmed, ""
	}
	return trimmed[:idx], strings.TrimSpace(trimmed[idx+1:])
}
