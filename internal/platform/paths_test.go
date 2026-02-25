package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to assert a path is absolute
func assertIsAbsolute(t *testing.T, path string, desc string) {
	t.Helper()
	if !filepath.IsAbs(path) {
		t.Errorf("%s: expected absolute path, got: %s", desc, path)
	}
}

// Helper function to assert a path is non-empty
func assertNonEmpty(t *testing.T, path string, desc string) {
	t.Helper()
	if path == "" {
		t.Errorf("%s: expected non-empty path, got empty string", desc)
	}
}

// Helper function to assert a path is in a specific directory
func assertInDir(t *testing.T, path string, dir string, desc string) {
	t.Helper()
	if !filepath.HasPrefix(path, dir) {
		t.Errorf("%s: expected path to be in %s, got: %s", desc, dir, path)
	}
}

// TestPathFunctions validates all path functions return valid, absolute paths.
func TestPathFunctions(t *testing.T) {
	pathFuncs := map[string]func() string{
		"SSHConfigPath":       SSHConfigPath,
		"SSHConfigBackupPath": SSHConfigBackupPath,
		"StateFilePath":       StateFilePath,
		"SSHKeyDir":           SSHKeyDir,
	}

	for name, fn := range pathFuncs {
		t.Run(name, func(t *testing.T) {
			path := fn()
			assertNonEmpty(t, path, name)
			assertIsAbsolute(t, path, name)
		})
	}
}

// TestSSHConfigPath validates SSH config path resolution.
func TestSSHConfigPath(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		path := SSHConfigPath()
		assertNonEmpty(t, path, "SSHConfigPath")
	})

	t.Run("returns absolute path", func(t *testing.T) {
		path := SSHConfigPath()
		assertIsAbsolute(t, path, "SSHConfigPath")
	})

	t.Run("path ends with .ssh/config", func(t *testing.T) {
		path := SSHConfigPath()
		expected := filepath.Join(".ssh", "config")
		if !strings.HasSuffix(path, expected) {
			t.Errorf("expected path to end with %s, got: %s", expected, path)
		}
	})

	t.Run("path contains home directory", func(t *testing.T) {
		path := SSHConfigPath()
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get home dir")
		}
		if !strings.HasPrefix(path, homeDir) {
			t.Errorf("expected path to start with home dir %s, got: %s", homeDir, path)
		}
	})

	t.Run("consistent across multiple calls", func(t *testing.T) {
		path1 := SSHConfigPath()
		path2 := SSHConfigPath()
		if path1 != path2 {
			t.Errorf("expected consistent paths, got %s and %s", path1, path2)
		}
	})
}

// TestSSHConfigBackupPath validates SSH config backup path resolution.
func TestSSHConfigBackupPath(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		path := SSHConfigBackupPath()
		assertNonEmpty(t, path, "SSHConfigBackupPath")
	})

	t.Run("returns absolute path", func(t *testing.T) {
		path := SSHConfigBackupPath()
		assertIsAbsolute(t, path, "SSHConfigBackupPath")
	})

	t.Run("path ends with config.bak", func(t *testing.T) {
		path := SSHConfigBackupPath()
		if !strings.HasSuffix(path, "config.bak") {
			t.Errorf("expected path to end with config.bak, got: %s", path)
		}
	})

	t.Run("backup is in same directory as config", func(t *testing.T) {
		configPath := SSHConfigPath()
		backupPath := SSHConfigBackupPath()
		configDir := filepath.Dir(configPath)
		backupDir := filepath.Dir(backupPath)

		if configDir != backupDir {
			t.Errorf("backup dir %s != config dir %s", backupDir, configDir)
		}
	})

	t.Run("consistent across multiple calls", func(t *testing.T) {
		path1 := SSHConfigBackupPath()
		path2 := SSHConfigBackupPath()
		if path1 != path2 {
			t.Errorf("expected consistent paths, got %s and %s", path1, path2)
		}
	})
}

// TestStateFilePath validates state file path resolution.
func TestStateFilePath(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		path := StateFilePath()
		assertNonEmpty(t, path, "StateFilePath")
	})

	t.Run("returns absolute path", func(t *testing.T) {
		path := StateFilePath()
		assertIsAbsolute(t, path, "StateFilePath")
	})

	t.Run("path contains swiftssh", func(t *testing.T) {
		path := StateFilePath()
		if !strings.Contains(path, "swiftssh") {
			t.Errorf("expected path to contain 'swiftssh', got: %s", path)
		}
	})

	t.Run("path ends with state.json", func(t *testing.T) {
		path := StateFilePath()
		if !strings.HasSuffix(path, "state.json") {
			t.Errorf("expected path to end with state.json, got: %s", path)
		}
	})

	t.Run("path contains proper directory structure", func(t *testing.T) {
		path := StateFilePath()
		expected := filepath.Join("swiftssh", "state.json")
		if !strings.HasSuffix(path, expected) {
			t.Errorf("expected path to end with %s, got: %s", expected, path)
		}
	})

	t.Run("consistent across multiple calls", func(t *testing.T) {
		path1 := StateFilePath()
		path2 := StateFilePath()
		if path1 != path2 {
			t.Errorf("expected consistent paths, got %s and %s", path1, path2)
		}
	})
}

// TestSSHKeyDir validates SSH key directory path resolution.
func TestSSHKeyDir(t *testing.T) {
	t.Run("returns non-empty path", func(t *testing.T) {
		path := SSHKeyDir()
		assertNonEmpty(t, path, "SSHKeyDir")
	})

	t.Run("returns absolute path", func(t *testing.T) {
		path := SSHKeyDir()
		assertIsAbsolute(t, path, "SSHKeyDir")
	})

	t.Run("path ends with .ssh", func(t *testing.T) {
		path := SSHKeyDir()
		if !strings.HasSuffix(path, ".ssh") {
			t.Errorf("expected path to end with .ssh, got: %s", path)
		}
	})

	t.Run("path contains home directory", func(t *testing.T) {
		path := SSHKeyDir()
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Skip("cannot get home dir")
		}
		if !strings.HasPrefix(path, homeDir) {
			t.Errorf("expected path to start with home dir %s, got: %s", homeDir, path)
		}
	})

	t.Run("consistent across multiple calls", func(t *testing.T) {
		path1 := SSHKeyDir()
		path2 := SSHKeyDir()
		if path1 != path2 {
			t.Errorf("expected consistent paths, got %s and %s", path1, path2)
		}
	})
}

// TestEnsureDir validates directory creation with parent paths.
func TestEnsureDir(t *testing.T) {
	t.Run("creates single directory", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "single")

		err := EnsureDir(testPath)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path exists but is not a directory")
		}
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "a", "b", "c", "d")

		err := EnsureDir(testPath)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		info, err := os.Stat(testPath)
		if err != nil {
			t.Fatalf("nested directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("path exists but is not a directory")
		}
	})

	t.Run("is idempotent (safe to call multiple times)", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "idempotent")

		// Call multiple times
		for i := 0; i < 3; i++ {
			err := EnsureDir(testPath)
			if err != nil {
				t.Fatalf("EnsureDir call %d failed: %v", i+1, err)
			}
		}

		// Verify directory exists
		info, err := os.Stat(testPath)
		if err != nil || !info.IsDir() {
			t.Error("directory does not exist after multiple calls")
		}
	})

	t.Run("respects existing directories", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "existing")

		// Create directory first
		err := os.Mkdir(testPath, 0755)
		if err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		// Call EnsureDir on existing directory
		err = EnsureDir(testPath)
		if err != nil {
			t.Errorf("EnsureDir failed on existing directory: %v", err)
		}

		// Verify it still exists
		info, err := os.Stat(testPath)
		if err != nil || !info.IsDir() {
			t.Error("directory was removed or not found")
		}
	})

	t.Run("creates intermediate parent directories", func(t *testing.T) {
		tempDir := t.TempDir()
		testPath := filepath.Join(tempDir, "deep", "nested", "path")

		// Verify parents don't exist
		if _, err := os.Stat(filepath.Join(tempDir, "deep")); !os.IsNotExist(err) {
			t.Skip("parent directory already exists")
		}

		err := EnsureDir(testPath)
		if err != nil {
			t.Fatalf("EnsureDir failed: %v", err)
		}

		// Verify all levels were created
		for _, path := range []string{
			filepath.Join(tempDir, "deep"),
			filepath.Join(tempDir, "deep", "nested"),
			filepath.Join(tempDir, "deep", "nested", "path"),
		} {
			info, err := os.Stat(path)
			if err != nil {
				t.Errorf("parent directory not created: %s", path)
			}
			if !info.IsDir() {
				t.Errorf("path exists but is not a directory: %s", path)
			}
		}
	})
}

// BenchmarkPathFunctions provides performance baselines for path functions.
func BenchmarkPathFunctions(b *testing.B) {
	benches := map[string]func(){
		"SSHConfigPath": func() {
			_ = SSHConfigPath()
		},
		"SSHConfigBackupPath": func() {
			_ = SSHConfigBackupPath()
		},
		"StateFilePath": func() {
			_ = StateFilePath()
		},
		"SSHKeyDir": func() {
			_ = SSHKeyDir()
		},
	}

	for name, fn := range benches {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fn()
			}
		})
	}
}
