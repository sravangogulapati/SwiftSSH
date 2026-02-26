package config

import (
	"fmt"
	"io"
	"os"
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

	// Write the new host block
	block := fmt.Sprintf("\nHost %s\n", h.Alias)
	block += fmt.Sprintf("    Hostname %s\n", h.Hostname)

	if h.User != "" {
		block += fmt.Sprintf("    User %s\n", h.User)
	}

	if h.Port != "" && h.Port != "22" {
		block += fmt.Sprintf("    Port %s\n", h.Port)
	}

	_, err = io.WriteString(f, block)
	if err != nil {
		return fmt.Errorf("failed to write host block: %w", err)
	}

	return nil
}
