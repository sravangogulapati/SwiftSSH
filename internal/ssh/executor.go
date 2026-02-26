package ssh

import (
	"os/exec"

	"github.com/srava/swiftssh/internal/config"
)

// BuildArgs constructs the SSH command-line arguments for a given host and identity.
func BuildArgs(host config.Host, identity string) []string {
	var args []string

	// Add identity flag if specified
	if identity != "" {
		args = append(args, "-i", identity)
	}

	// Add port flag if non-default and non-empty
	if host.Port != "" && host.Port != "22" {
		args = append(args, "-p", host.Port)
	}

	// Add user flag if specified
	if host.User != "" {
		args = append(args, "-l", host.User)
	}

	// Always add the host alias
	args = append(args, host.Alias)

	return args
}

// ConnectCmd returns an exec.Cmd for connecting to the host via SSH.
func ConnectCmd(host config.Host, identity string) *exec.Cmd {
	return exec.Command("ssh", BuildArgs(host, identity)...)
}
