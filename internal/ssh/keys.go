package ssh

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanPublicKeys returns a list of private key paths from the SSH directory.
// It looks for .pub files and only includes them if the corresponding private key exists.
func ScanPublicKeys(sshDir string) ([]string, error) {
	if sshDir == "" {
		return []string{}, nil
	}

	// Find all .pub files
	pubFiles, err := filepath.Glob(filepath.Join(sshDir, "*.pub"))
	if err != nil {
		return []string{}, err
	}

	var keys []string
	for _, pubFile := range pubFiles {
		// Strip .pub suffix to get the private key path
		privateKeyPath := strings.TrimSuffix(pubFile, ".pub")

		// Check if the private key exists
		if _, err := os.Stat(privateKeyPath); err == nil {
			keys = append(keys, privateKeyPath)
		}
	}

	return keys, nil
}

// KeyLabel returns the filename without the .pub extension from a public key path.
func KeyLabel(pubKeyPath string) string {
	return strings.TrimSuffix(filepath.Base(pubKeyPath), ".pub")
}
