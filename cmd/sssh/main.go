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

var version = "dev"

// extractConfigFlag pre-scans args for --config <path> or --config=<path>
// without calling flag.Parse(), so it works before the SSH passthrough check.
func extractConfigFlag(args []string) string {
	for i, arg := range args {
		if (arg == "--config" || arg == "-config") && i+1 < len(args) {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config=")
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config=")
		}
	}
	return ""
}

func main() {
	rawArgs := os.Args[1:]
	configOverride := extractConfigFlag(rawArgs) // pre-scan before flag.Parse

	// Detect SSH passthrough invocations before flag.Parse() so that
	// SSH flags like -i, -p, -l don't trigger "flag provided but not defined".
	// A passthrough call contains at least one argument that is either
	// user@host syntax or an SSH option flag (-i, -p, -l, etc.).
	if looksLikeSSHArgs(rawArgs) {
		runPassthrough(rawArgs, configOverride)
		return
	}

	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.BoolVar(showVersion, "v", false, "Print version and exit (shorthand)")
	configFlag := flag.String("config", "", "Path to SSH config file")
	noFrequent := flag.Bool("no-frequent", false, "Flat alphabetical order (skip frequency sort)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("sssh %s\n", version)
		os.Exit(0)
	}

	configPath := platform.SSHConfigPath()
	if *configFlag != "" {
		configPath = *configFlag
	}

	hosts, err := config.Parse(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not parse SSH config: %v\n", err)
		os.Exit(1)
	}

	if len(hosts) == 0 {
		fmt.Printf("No hosts found in %s. Add entries to your SSH config.\n", configPath)
		os.Exit(0)
	}

	statePath := platform.StateFilePath()
	st, err := state.Load(statePath)
	if err != nil {
		st = &state.State{Connections: make(map[string]int)}
	}

	p := tea.NewProgram(tui.New(hosts, st, statePath, *noFrequent), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// runPassthrough parses SSH-style arguments, auto-saves unknown hosts to
// the SSH config, then hands off to the system ssh binary.
func runPassthrough(args []string, configOverride string) {
	dest, port, user, identity := parseSSHTarget(args)
	if dest == "" {
		fmt.Fprintln(os.Stderr, "sssh: no destination found in arguments")
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

	// Auto-append to config if not already known
	configPath := platform.SSHConfigPath()
	if configOverride != "" {
		configPath = configOverride
	}
	backupPath := filepath.Join(filepath.Dir(configPath), "config.bak")
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
			fmt.Fprintf(os.Stderr, "sssh: warning: could not save host to config: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "sssh: saved '%s' to SSH config\n", alias)
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

// looksLikeSSHArgs reports whether args appear to be an SSH passthrough
// invocation rather than sssh-native flags. It returns true when any
// argument contains "@" (user@host) or is a recognized SSH option flag.
func looksLikeSSHArgs(args []string) bool {
	sshFlags := map[string]bool{
		"-i": true, "-p": true, "-l": true, "-b": true, "-c": true,
		"-D": true, "-E": true, "-e": true, "-F": true, "-I": true,
		"-J": true, "-L": true, "-m": true, "-o": true, "-Q": true,
		"-R": true, "-S": true, "-w": true, "-W": true,
		// boolean SSH flags
		"-4": true, "-6": true, "-A": true, "-a": true, "-C": true,
		"-f": true, "-G": true, "-g": true, "-K": true, "-k": true,
		"-M": true, "-N": true, "-n": true, "-q": true, "-s": true,
		"-T": true, "-t": true, "-V": true, "-X": true, "-x": true,
		"-Y": true, "-y": true,
	}
	for _, arg := range args {
		if strings.Contains(arg, "@") {
			return true
		}
		if sshFlags[arg] {
			return true
		}
	}
	return false
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
