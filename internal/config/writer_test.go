package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsKnownHost_Found(t *testing.T) {
	hosts := []Host{
		{Alias: "dev", Hostname: "192.168.1.1"},
		{Alias: "prod", Hostname: "10.0.0.1"},
	}

	if !IsKnownHost(hosts, "192.168.1.1") {
		t.Error("expected IsKnownHost to return true for known hostname")
	}
}

func TestIsKnownHost_NotFound(t *testing.T) {
	hosts := []Host{
		{Alias: "dev", Hostname: "192.168.1.1"},
	}

	if IsKnownHost(hosts, "192.168.1.2") {
		t.Error("expected IsKnownHost to return false for unknown hostname")
	}
}

func TestIsKnownHost_EmptyList(t *testing.T) {
	if IsKnownHost([]Host{}, "192.168.1.1") {
		t.Error("expected IsKnownHost to return false for empty list")
	}
}

func TestAppendHost_WritesBlock(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	// Create initial config
	initialContent := "Host existing\n    Hostname old.example.com\n"
	if err := os.WriteFile(configPath, []byte(initialContent), 0600); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Append new host
	newHost := Host{
		Alias:    "newhost",
		Hostname: "new.example.com",
		User:     "alice",
		Port:     "2222",
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	// Read and verify config
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, "Host newhost") {
		t.Error("expected 'Host newhost' in config")
	}
	if !strings.Contains(configStr, "Hostname new.example.com") {
		t.Error("expected 'Hostname new.example.com' in config")
	}
	if !strings.Contains(configStr, "User alice") {
		t.Error("expected 'User alice' in config")
	}
	if !strings.Contains(configStr, "Port 2222") {
		t.Error("expected 'Port 2222' in config")
	}

	// Verify original content is preserved
	if !strings.Contains(configStr, "Host existing") {
		t.Error("expected original 'Host existing' to be preserved")
	}
}

func TestAppendHost_CreatesBackup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	// Create initial config
	initialContent := "Host existing\n    Hostname old.example.com\n"
	if err := os.WriteFile(configPath, []byte(initialContent), 0600); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Append new host
	newHost := Host{Alias: "test", Hostname: "test.example.com"}
	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	// Verify backup contains original content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup: %v", err)
	}

	if string(backupContent) != initialContent {
		t.Errorf("backup content mismatch: expected %q, got %q", initialContent, string(backupContent))
	}
}

func TestAppendHost_OmitsEmptyUserPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	// Append host without User and Port
	newHost := Host{
		Alias:    "simple",
		Hostname: "simple.example.com",
		User:     "",
		Port:     "",
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	// Read and verify config
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	configStr := string(content)
	if strings.Contains(configStr, "User") {
		t.Error("expected no 'User' line in config")
	}
	if strings.Contains(configStr, "Port") {
		t.Error("expected no 'Port' line in config")
	}
	if !strings.Contains(configStr, "Host simple") {
		t.Error("expected 'Host simple' in config")
	}
}

func TestAppendHost_QuotesIdentityFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	newHost := Host{
		Alias:        "azure",
		Hostname:     "4.227.83.154",
		User:         "azureuser",
		IdentityFile: "/home/user/my keys/ssh_key.pem",
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, `IdentityFile "/home/user/my keys/ssh_key.pem"`) {
		t.Errorf("expected quoted IdentityFile path, got:\n%s", configStr)
	}
}

func TestAppendHost_OmitsDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	newHost := Host{
		Alias:    "standard",
		Hostname: "standard.example.com",
		User:     "bob",
		Port:     "22",
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	configStr := string(content)
	if strings.Contains(configStr, "Port 22") {
		t.Error("expected default port 22 to be omitted")
	}
	if !strings.Contains(configStr, "User bob") {
		t.Error("expected 'User bob' in config")
	}
}

// TestAppendHost_WritesGroups is a regression test ensuring groups are written on append.
func TestAppendHost_WritesGroups(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	newHost := Host{
		Alias:    "grouped",
		Hostname: "grouped.example.com",
		Groups:   []string{"Work", "Personal"},
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, "# @group Work, Personal") {
		t.Errorf("expected magic comment in appended block, got:\n%s", configStr)
	}
}

// TestAppendHost_EmptyFile_NoLeadingBlankLine verifies that when AppendHost is called
// on an empty (or non-existent) config file, no leading blank line is written before
// the host block. The first line of the resulting file must be the "Host " directive
// (or a magic comment if groups are present), not an empty line.
func TestAppendHost_EmptyFile_NoLeadingBlankLine(t *testing.T) {
	tmpDir := t.TempDir()
	// Use a path that doesn't exist yet — AppendHost should create it.
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	newHost := Host{
		Alias:    "firsthost",
		Hostname: "first.example.com",
	}

	if err := AppendHost(configPath, backupPath, newHost); err != nil {
		t.Fatalf("AppendHost failed: %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		t.Fatal("config file is empty after AppendHost")
	}

	firstLine := lines[0]
	if firstLine == "" {
		t.Errorf("first line of config should not be blank; got empty string (leading blank line bug)")
	}
	if !strings.HasPrefix(firstLine, "Host ") {
		t.Errorf("expected first line to start with 'Host ', got %q", firstLine)
	}
}

// --- buildHostBlock tests ---

func TestBuildHostBlock_AllFields(t *testing.T) {
	h := Host{
		Alias:        "myhost",
		Hostname:     "my.example.com",
		User:         "alice",
		Port:         "2222",
		IdentityFile: "/home/alice/.ssh/id_rsa",
		Groups:       []string{"Work", "Personal"},
	}

	block := buildHostBlock(h)

	if !strings.Contains(block, "# @group Work, Personal") {
		t.Error("expected magic comment line")
	}
	if !strings.Contains(block, "Host myhost") {
		t.Error("expected Host line")
	}
	if !strings.Contains(block, "Hostname my.example.com") {
		t.Error("expected Hostname line")
	}
	if !strings.Contains(block, "User alice") {
		t.Error("expected User line")
	}
	if !strings.Contains(block, "Port 2222") {
		t.Error("expected Port line")
	}
	if !strings.Contains(block, `IdentityFile "/home/alice/.ssh/id_rsa"`) {
		t.Error("expected quoted IdentityFile line")
	}
}

func TestBuildHostBlock_NoGroups(t *testing.T) {
	h := Host{
		Alias:    "simple",
		Hostname: "simple.example.com",
	}

	block := buildHostBlock(h)

	if strings.Contains(block, "@group") {
		t.Error("expected no magic comment when Groups is empty")
	}
	if !strings.HasPrefix(block, "Host simple\n") {
		t.Errorf("expected block to start with 'Host simple', got:\n%s", block)
	}
}

func TestBuildHostBlock_OmitsDefaultPort(t *testing.T) {
	h := Host{
		Alias:    "svc",
		Hostname: "svc.example.com",
		Port:     "22",
	}

	block := buildHostBlock(h)
	if strings.Contains(block, "Port") {
		t.Error("expected Port 22 to be omitted")
	}
}

// --- ReplaceHostBlock tests ---

// writeHostConfig writes content to a temp file and returns the path.
func writeHostConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

func TestReplaceHostBlock_Basic(t *testing.T) {
	content := "Host first\n    Hostname first.example.com\n\nHost second\n    Hostname second.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "first-renamed",
		Hostname:   "new.example.com",
		Port:       "22",
		SourceFile: path,
		LineStart:  1,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	s := string(result)

	if !strings.Contains(s, "Host first-renamed") {
		t.Error("expected renamed Host line")
	}
	if !strings.Contains(s, "Hostname new.example.com") {
		t.Error("expected new Hostname")
	}
	// Second host must still be present
	if !strings.Contains(s, "Host second") {
		t.Error("expected second host to remain")
	}
}

func TestReplaceHostBlock_WithMagicComment(t *testing.T) {
	content := "# @group OldGroup\nHost myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "new.example.com",
		Groups:     []string{"NewGroup"},
		SourceFile: path,
		LineStart:  2,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	s := string(result)

	if strings.Contains(s, "OldGroup") {
		t.Error("expected OldGroup to be replaced")
	}
	if !strings.Contains(s, "# @group NewGroup") {
		t.Error("expected NewGroup magic comment")
	}
}

func TestReplaceHostBlock_AddGroups(t *testing.T) {
	content := "Host myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     []string{"Work"},
		SourceFile: path,
		LineStart:  1,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	if !strings.Contains(string(result), "# @group Work") {
		t.Error("expected magic comment to be added")
	}
}

func TestReplaceHostBlock_RemoveGroups(t *testing.T) {
	content := "# @group Work\nHost myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     nil, // remove groups
		SourceFile: path,
		LineStart:  2,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	if strings.Contains(string(result), "@group") {
		t.Error("expected magic comment to be removed")
	}
}

func TestReplaceHostBlock_LastHost(t *testing.T) {
	content := "Host first\n    Hostname first.example.com\n\nHost last\n    Hostname last.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "last",
		Hostname:   "updated.example.com",
		Port:       "22",
		SourceFile: path,
		LineStart:  4,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	s := string(result)

	if !strings.Contains(s, "Hostname updated.example.com") {
		t.Error("expected updated hostname for last host")
	}
	if !strings.Contains(s, "Host first") {
		t.Error("expected first host to remain")
	}
}

func TestReplaceHostBlock_StaleLine(t *testing.T) {
	content := "Host myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	// LineStart points to a non-Host line
	h := Host{
		Alias:      "myhost",
		Hostname:   "new.example.com",
		SourceFile: path,
		LineStart:  2, // line 2 is "    Hostname old.example.com"
	}

	_, _, err := ReplaceHostBlock(h)
	if err == nil {
		t.Error("expected error for stale LineStart pointing to non-Host line")
	}
}

func TestReplaceHostBlock_ZeroLineStart(t *testing.T) {
	h := Host{
		Alias:      "myhost",
		Hostname:   "example.com",
		SourceFile: "/tmp/nonexistent",
		LineStart:  0,
	}

	_, _, err := ReplaceHostBlock(h)
	if err == nil {
		t.Error("expected error when LineStart is 0")
	}
}

func TestReplaceHostBlock_CreatesBackup(t *testing.T) {
	content := "Host myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "new.example.com",
		Port:       "22",
		SourceFile: path,
		LineStart:  1,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	backupPath := path + ".bak"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("expected backup file at %s: %v", backupPath, err)
	}

	if string(backup) != content {
		t.Errorf("backup content mismatch:\nexpected: %q\ngot: %q", content, string(backup))
	}
}

// TestReplaceHostBlock_MagicCommentAtBlockStart tests the lenient stale check: when
// LineStart erroneously points to the @group comment rather than the Host line,
// the function should still succeed and return the correct new LineStart.
func TestReplaceHostBlock_MagicCommentAtBlockStart(t *testing.T) {
	content := "# @group Local\nHost myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "new.example.com",
		Groups:     []string{"Local"},
		SourceFile: path,
		LineStart:  1, // points to "# @group Local", not the Host line
	}

	newLineStart, _, err := ReplaceHostBlock(h)
	if err != nil {
		t.Fatalf("expected lenient stale check to succeed, got: %v", err)
	}

	result, _ := os.ReadFile(path)
	if !strings.Contains(string(result), "Hostname new.example.com") {
		t.Error("expected new hostname in result")
	}
	// With groups, Host line follows the magic comment, so newLineStart = magicStart+2 = 2
	if newLineStart != 2 {
		t.Errorf("expected newLineStart=2, got %d", newLineStart)
	}
}

// TestReplaceHostBlock_ReturnsNewLineStart_AddGroups tests that adding groups to a
// previously group-less host returns newLineStart=2 (Host line shifted down by magic comment).
func TestReplaceHostBlock_ReturnsNewLineStart_AddGroups(t *testing.T) {
	content := "Host myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     []string{"Work"}, // adding a group for the first time
		SourceFile: path,
		LineStart:  1, // Host line is at line 1 before the save
	}

	newLineStart, _, err := ReplaceHostBlock(h)
	if err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	// After save: line 1 = "# @group Work", line 2 = "Host myhost"
	if newLineStart != 2 {
		t.Errorf("expected newLineStart=2 after adding groups, got %d", newLineStart)
	}
}

// TestReplaceHostBlock_ReturnsNewLineStart_RemoveGroups tests that removing all groups
// from a host returns newLineStart=1 (Host line shifts up; no magic comment).
func TestReplaceHostBlock_ReturnsNewLineStart_RemoveGroups(t *testing.T) {
	content := "# @group Work\nHost myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     nil, // removing all groups
		SourceFile: path,
		LineStart:  2, // Host line is at line 2 (after @group comment)
	}

	newLineStart, _, err := ReplaceHostBlock(h)
	if err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	// After save: line 1 = "Host myhost" (magic comment removed)
	if newLineStart != 1 {
		t.Errorf("expected newLineStart=1 after removing groups, got %d", newLineStart)
	}
}

// TestReplaceHostBlock_PreservesBlankLine verifies that a blank line between two host
// blocks is not erased when the first block is rewritten.
func TestReplaceHostBlock_PreservesBlankLine(t *testing.T) {
	content := "Host first\n    Hostname first.example.com\n\nHost second\n    Hostname second.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "first",
		Hostname:   "first-updated.example.com",
		Port:       "22",
		SourceFile: path,
		LineStart:  1,
	}

	if _, _, err := ReplaceHostBlock(h); err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	result, _ := os.ReadFile(path)
	s := string(result)

	// Both hosts must still be present
	if !strings.Contains(s, "Host first") {
		t.Error("expected first host in result")
	}
	if !strings.Contains(s, "Host second") {
		t.Error("expected second host in result")
	}

	// The blank separator line must still be present between the two blocks
	if !strings.Contains(s, "first-updated.example.com\n\nHost second") {
		t.Errorf("expected blank line to be preserved between host blocks, got:\n%s", s)
	}
}

// TestReplaceHostBlock_ReturnsLineDelta_AddGroup verifies that adding a group to a
// previously group-less host returns lineDelta=+1.
func TestReplaceHostBlock_ReturnsLineDelta_AddGroup(t *testing.T) {
	content := "Host myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     []string{"Work"}, // adding group for first time
		SourceFile: path,
		LineStart:  1,
	}

	newLineStart, lineDelta, err := ReplaceHostBlock(h)
	if err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	if newLineStart != 2 {
		t.Errorf("expected newLineStart=2, got %d", newLineStart)
	}
	if lineDelta != 1 {
		t.Errorf("expected lineDelta=+1 when adding a group, got %d", lineDelta)
	}
}

// TestReplaceHostBlock_ReturnsLineDelta_RemoveGroup verifies that removing all groups
// from a host returns lineDelta=-1.
func TestReplaceHostBlock_ReturnsLineDelta_RemoveGroup(t *testing.T) {
	content := "# @group Work\nHost myhost\n    Hostname old.example.com\n"
	path := writeHostConfig(t, content)

	h := Host{
		Alias:      "myhost",
		Hostname:   "old.example.com",
		Groups:     nil, // removing all groups
		SourceFile: path,
		LineStart:  2,
	}

	newLineStart, lineDelta, err := ReplaceHostBlock(h)
	if err != nil {
		t.Fatalf("ReplaceHostBlock failed: %v", err)
	}

	if newLineStart != 1 {
		t.Errorf("expected newLineStart=1, got %d", newLineStart)
	}
	if lineDelta != -1 {
		t.Errorf("expected lineDelta=-1 when removing a group, got %d", lineDelta)
	}
}
