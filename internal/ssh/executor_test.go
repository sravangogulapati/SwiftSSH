package ssh

import (
	"testing"

	"github.com/srava/swiftssh/internal/config"
)

func TestBuildArgs_NoIdentityDefaultPort(t *testing.T) {
	host := config.Host{
		Alias:    "dev",
		Hostname: "192.168.1.100",
		User:     "alice",
		Port:     "22",
	}

	args := BuildArgs(host, "")
	expected := []string{"-l", "alice", "dev"}

	if len(args) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(args))
	}
	for i, arg := range args {
		if i < len(expected) && arg != expected[i] {
			t.Errorf("arg %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}

func TestBuildArgs_WithIdentity(t *testing.T) {
	host := config.Host{
		Alias:    "prod",
		Hostname: "10.0.0.1",
		User:     "bob",
		Port:     "22",
	}
	identity := "/home/user/.ssh/id_ed25519"

	args := BuildArgs(host, identity)
	expected := []string{"-i", identity, "-l", "bob", "prod"}

	if len(args) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(args))
	}
	for i, arg := range args {
		if i < len(expected) && arg != expected[i] {
			t.Errorf("arg %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}

func TestBuildArgs_NonDefaultPort(t *testing.T) {
	host := config.Host{
		Alias:    "staging",
		Hostname: "staging.example.com",
		User:     "ubuntu",
		Port:     "2222",
	}

	args := BuildArgs(host, "")
	expected := []string{"-p", "2222", "-l", "ubuntu", "staging"}

	if len(args) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(args))
	}
	for i, arg := range args {
		if i < len(expected) && arg != expected[i] {
			t.Errorf("arg %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}

func TestBuildArgs_EmptyUserAndIdentity(t *testing.T) {
	host := config.Host{
		Alias:    "simple",
		Hostname: "example.com",
		User:     "",
		Port:     "22",
	}

	args := BuildArgs(host, "")
	expected := []string{"simple"}

	if len(args) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(args))
	}
	for i, arg := range args {
		if i < len(expected) && arg != expected[i] {
			t.Errorf("arg %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}
