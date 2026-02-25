package platform

import (
	"os"
	"path/filepath"
)

// SSHConfigPath returns the path to ~/.ssh/config (or Windows equivalent).
func SSHConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config")
}

// SSHConfigBackupPath returns the path to ~/.ssh/config.bak (or Windows equivalent).
func SSHConfigBackupPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "config.bak")
}

// StateFilePath returns the path to the state file.
// On Unix: ~/.config/swiftssh/state.json
// On Windows: %LOCALAPPDATA%\swiftssh\state.json
func StateFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "swiftssh", "state.json")
}

// SSHKeyDir returns the path to ~/.ssh (or Windows equivalent).
func SSHKeyDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh")
}

// EnsureDir creates a directory and all parent directories if they don't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}
