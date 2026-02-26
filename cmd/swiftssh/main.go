package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/platform"
	"github.com/srava/swiftssh/internal/state"
	"github.com/srava/swiftssh/internal/tui"
)

const Version = "0.1.0"

func main() {
	version := flag.Bool("version", false, "Print version and exit")
	flag.BoolVar(version, "v", false, "Print version and exit (shorthand)")
	flag.Parse()

	if *version {
		fmt.Printf("swiftssh v%s\n", Version)
		os.Exit(0)
	}

	// If extra arguments are given, treat them as SSH passthrough arguments:
	//   swiftssh user@hostname [-p port] [-i key] ...
	if args := flag.Args(); len(args) > 0 {
		runPassthrough(args)
		return
	}

	// No arguments — launch interactive TUI
	configPath := platform.SSHConfigPath()
	hosts, err := config.Parse(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not parse SSH config: %v\n", err)
		os.Exit(1)
	}

	statePath := platform.StateFilePath()
	st, err := state.Load(statePath)
	if err != nil {
		st = &state.State{Connections: make(map[string]int)}
	}

	p := tea.NewProgram(tui.New(hosts, st, statePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// runPassthrough parses SSH-style arguments, auto-saves unknown hosts to
// ~/.ssh/config, then hands off to the system ssh binary.
func runPassthrough(args []string) {
	dest, port, user, identity := parseSSHTarget(args)
	if dest == "" {
		fmt.Fprintln(os.Stderr, "swiftssh: no destination found in arguments")
		os.Exit(1)
	}

	// Separate user from hostname if provided as user@hostname
	hostname := dest
	if idx := strings.Index(dest, "@"); idx >= 0 {
		if user == "" {
			user = dest[:idx]
		}
		hostname = dest[idx+1:]
	}

	if port == "" {
		port = "22"
	}

	// Auto-append to ~/.ssh/config if not already known
	configPath := platform.SSHConfigPath()
	backupPath := platform.SSHConfigBackupPath()
	hosts, _ := config.Parse(configPath)
	if !config.IsKnownHost(hosts, hostname) {
		alias := hostname
		if user != "" {
			alias = user + "-" + hostname
		}
		absIdentity := identity
		if identity != "" {
			if abs, err := filepath.Abs(identity); err == nil {
				absIdentity = abs
			}
		}
		h := config.Host{
			Alias:        alias,
			Hostname:     hostname,
			User:         user,
			Port:         port,
			IdentityFile: absIdentity,
		}
		if err := config.AppendHost(configPath, backupPath, h); err != nil {
			fmt.Fprintf(os.Stderr, "swiftssh: warning: could not save host to config: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "swiftssh: saved '%s' to SSH config\n", alias)
		}
	}

	// Hand off to ssh with the original arguments unchanged
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

// parseSSHTarget scans SSH-style arguments and extracts the destination,
// port (-p), user (-l), and identity (-i). The destination is the first
// positional argument (not preceded by an option that takes a value).
func parseSSHTarget(args []string) (dest, port, user, identity string) {
	// SSH options that consume the next argument as their value
	optWithValue := map[string]bool{
		"-b": true, "-c": true, "-D": true, "-E": true, "-e": true,
		"-F": true, "-I": true, "-i": true, "-J": true, "-L": true,
		"-l": true, "-m": true, "-o": true, "-p": true, "-Q": true,
		"-R": true, "-S": true, "-w": true, "-W": true,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch arg {
		case "-p":
			if i+1 < len(args) {
				port = args[i+1]
				i += 2
				continue
			}
		case "-l":
			if i+1 < len(args) {
				user = args[i+1]
				i += 2
				continue
			}
		case "-i":
			if i+1 < len(args) {
				identity = args[i+1]
				i += 2
				continue
			}
		default:
			if optWithValue[arg] && i+1 < len(args) {
				i += 2 // skip option + value we don't care about
				continue
			}
			if !strings.HasPrefix(arg, "-") && dest == "" {
				dest = arg
			}
		}
		i++
	}
	return
}
