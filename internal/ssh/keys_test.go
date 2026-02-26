package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanPublicKeys_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create id_rsa.pub and id_rsa
	pubPath := filepath.Join(tmpDir, "id_rsa.pub")
	privPath := filepath.Join(tmpDir, "id_rsa")

	if err := os.WriteFile(pubPath, []byte("public key content"), 0644); err != nil {
		t.Fatalf("failed to write pub file: %v", err)
	}
	if err := os.WriteFile(privPath, []byte("private key content"), 0600); err != nil {
		t.Fatalf("failed to write private key file: %v", err)
	}

	keys, err := ScanPublicKeys(tmpDir)
	if err != nil {
		t.Fatalf("ScanPublicKeys failed: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
	if len(keys) > 0 && keys[0] != privPath {
		t.Errorf("expected key path %q, got %q", privPath, keys[0])
	}
}

func TestScanPublicKeys_ExcludesMissingPrivateKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Create orphan.pub without a corresponding private key
	pubPath := filepath.Join(tmpDir, "orphan.pub")
	if err := os.WriteFile(pubPath, []byte("public key content"), 0644); err != nil {
		t.Fatalf("failed to write pub file: %v", err)
	}

	keys, err := ScanPublicKeys(tmpDir)
	if err != nil {
		t.Fatalf("ScanPublicKeys failed: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}

func TestKeyLabel(t *testing.T) {
	label := KeyLabel("/home/user/.ssh/id_ed25519.pub")
	expected := "id_ed25519"
	if label != expected {
		t.Errorf("expected label %q, got %q", expected, label)
	}
}
