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

func TestAppendHost_OmitsDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config")
	backupPath := filepath.Join(tmpDir, "config.bak")

	// Append host with default port
	newHost := Host{
		Alias:    "standard",
		Hostname: "standard.example.com",
		User:     "bob",
		Port:     "22",
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
	if strings.Contains(configStr, "Port 22") {
		t.Error("expected default port 22 to be omitted")
	}
	if !strings.Contains(configStr, "User bob") {
		t.Error("expected 'User bob' in config")
	}
}
