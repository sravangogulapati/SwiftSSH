package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Parse reads the SSH config file at configPath and returns all hosts.
// It handles Include directives with glob expansion and circular include detection.
func Parse(configPath string) ([]Host, error) {
	visited := make(map[string]bool)
	return parseFile(configPath, visited)
}

// parseFile is the recursive parser that handles a single config file.
func parseFile(path string, visited map[string]bool) ([]Host, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	// Get absolute cleaned path for circular detection
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path // fallback if Abs fails
	}
	absPath = filepath.Clean(absPath)

	// Check for circular include
	if visited[absPath] {
		return nil, nil // silently skip already visited files
	}
	visited[absPath] = true

	var hosts []Host
	var current *Host
	var prevLine string
	var lineNum int

	scanner := bufio.NewScanner(file)
	configDir := filepath.Dir(path)

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Find first whitespace to split keyword and value
		trimmed := strings.TrimSpace(line)

		// Handle empty lines and all comment lines (including magic comments).
		// Magic comments set prevLine so the next Host directive can pick up groups.
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			prevLine = line
			continue
		}

		// Parse keyword and value
		idx := strings.IndexAny(trimmed, " \t")
		if idx == -1 {
			// keyword only, no value
			prevLine = line
			continue
		}

		keyword := trimmed[:idx]
		value := strings.TrimSpace(trimmed[idx+1:])

		// Handle directives
		switch strings.ToLower(keyword) {
		case "host":
			// Finalize previous host if exists and not wildcard
			if current != nil && current.Alias != "*" {
				// Set default port if not specified
				if current.Port == "" {
					current.Port = "22"
				}
				hosts = append(hosts, *current)
			}
			// Start new host block
			current = &Host{
				Alias:      value,
				SourceFile: path,
				Groups:     parseMagicComment(prevLine),
				LineStart:  lineNum,
			}

		case "hostname":
			if current != nil {
				current.Hostname = value
			}

		case "user":
			if current != nil {
				current.User = value
			}

		case "port":
			if current != nil {
				current.Port = value
			}

		case "identityfile":
			if current != nil {
				current.IdentityFile = strings.Trim(value, `"`)
			}

		case "include":
			// Finalize current host if any before processing global directive
			if current != nil && current.Alias != "*" {
				if current.Port == "" {
					current.Port = "22"
				}
				hosts = append(hosts, *current)
				current = nil
			}

			// Process include directive
			expanded, err := expandTilde(value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "swiftssh: warning: include %q: %v\n", value, err)
				prevLine = line
				continue
			}

			// Resolve relative to config directory if not absolute
			if !filepath.IsAbs(expanded) {
				expanded = filepath.Join(configDir, expanded)
			}

			// Glob expansion
			matches, err := filepath.Glob(expanded)
			if err != nil {
				fmt.Fprintf(os.Stderr, "swiftssh: warning: include %q: glob error: %v\n", value, err)
				prevLine = line
				continue
			}

			if len(matches) == 0 {
				fmt.Fprintf(os.Stderr, "swiftssh: warning: include %q: no files matched\n", expanded)
				prevLine = line
				continue
			}

			// Recursively parse each matched file
			for _, match := range matches {
				// Get absolute cleaned path
				absMatch, cleanErr := filepath.Abs(match)
				if cleanErr != nil {
					absMatch = match
				}
				absMatch = filepath.Clean(absMatch)

				// Check if already visited (avoid infinite recursion)
				if visited[absMatch] {
					continue
				}

				// Recursively parse
				includedHosts, parseErr := parseFile(match, visited)
				if parseErr != nil {
					fmt.Fprintf(os.Stderr, "swiftssh: warning: include %q: %v\n", match, parseErr)
					continue
				}
				hosts = append(hosts, includedHosts...)
			}
		}

		prevLine = line
	}

	// Finalize last open host block
	if current != nil && current.Alias != "*" {
		if current.Port == "" {
			current.Port = "22"
		}
		hosts = append(hosts, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return hosts, nil
}

// parseMagicComment extracts groups from a magic comment line.
// Format: # @group Work, Personal
// Returns nil if the line is not a magic comment.
func parseMagicComment(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "#") {
		return nil
	}

	rest := strings.TrimSpace(trimmed[1:])
	if !strings.HasPrefix(rest, "@group") {
		return nil
	}

	// Extract the part after "@group"
	tagsPart := strings.TrimSpace(strings.TrimPrefix(rest, "@group"))
	if tagsPart == "" {
		return nil
	}

	// Split on comma and trim each tag
	parts := strings.Split(tagsPart, ",")
	var groups []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			groups = append(groups, trimmed)
		}
	}

	if len(groups) == 0 {
		return nil
	}

	return groups
}

// expandTilde expands ~ to home directory.
func expandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot get home directory: %w", err)
		}
		return home, nil
	}
	return path, nil
}
