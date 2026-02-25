# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwiftSSH is a single-binary Go CLI tool that provides an interactive, searchable TUI for managing SSH connections from `~/.ssh/config`. See PRD.md for detailed specifications.

**Key Dependencies:**
- Bubble Tea (TUI framework) - github.com/charmbracelet/bubbletea
- ssh/config parsing - golang.org/x/crypto/ssh/config (standard) or custom parser
- Cross-platform file handling (Windows, Linux, macOS)

## Common Development Commands

### Setup & Dependencies
```bash
# Initialize Go module (if not already done)
go mod init github.com/yourusername/swiftssh
go mod tidy

# Install dependencies
go get github.com/charmbracelet/bubbletea
```

### Building
```bash
# Build the binary
go build -o swiftssh ./cmd/swiftssh

# Build with optimizations (release build)
go build -ldflags="-s -w" -o swiftssh ./cmd/swiftssh
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/config

# Run a single test (verbose output)
go test -v ./internal/config -run TestParseSSHConfig

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality
```bash
# Lint with golangci-lint (if installed)
golangci-lint run

# Format code (standard Go formatter)
go fmt ./...

# Run vet (static analysis)
go vet ./...

# Check for unused dependencies
go mod tidy -v
```

### Running Locally
```bash
# Run the tool directly from source
go run ./cmd/swiftssh [args]

# Build and run in one step
go run ./cmd/swiftssh
```

## Architecture & Code Organization

### Directory Structure
```
.
├── cmd/
│   └── swiftssh/           # Main entry point
│       └── main.go
├── internal/
│   ├── config/             # SSH config parsing & management
│   │   ├── parser.go       # Parse ~/.ssh/config with Include support
│   │   ├── types.go        # Host, Group, config data structures
│   │   └── parser_test.go
│   ├── state/              # Persistence layer
│   │   ├── state.go        # Connection frequency tracking
│   │   └── state_test.go
│   ├── ssh/                # SSH execution & identity management
│   │   ├── executor.go     # tea.ExecProcess wrapper for SSH
│   │   ├── keys.go         # Scan & select SSH keys
│   │   └── executor_test.go
│   ├── tui/                # Bubble Tea UI layer
│   │   ├── model.go        # Main TUI model (list, search, selection)
│   │   ├── views.go        # Render functions for different screens
│   │   ├── keybindings.go  # Key handler (j/k nav, Enter, i for identity, p for ping, etc)
│   │   └── model_test.go
│   ├── health/             # Health check (TCP dial to port 22)
│   │   ├── check.go
│   │   └── check_test.go
│   └── platform/           # Platform-specific paths & utilities
│       ├── paths.go        # Return correct config/state paths per OS
│       └── paths_test.go
└── go.mod, go.sum
```

### Core Components & Data Flow

1. **Config Parser (`internal/config`)**: Reads `~/.ssh/config`, parses Host blocks, respects Include directives. Returns a flat list of hosts with metadata (hostname, user, port, groups from magic comments).

2. **State Manager (`internal/state`)**: Maintains `~/.config/swiftssh/state.json` (or Windows equivalent). Tracks connection frequency and provides "Frequent" section at top of list.

3. **TUI Model (`internal/tui`)**: Bubble Tea model that:
   - Maintains current list state (filtered hosts, cursor position, search query)
   - Handles key events (vim keys: j/k, Enter to connect, i for identity picker, p for ping toggle, / for search)
   - Renders the viewport with visible hosts
   - Manages mode switching (search mode, identity picker, etc)

4. **SSH Executor (`internal/ssh`)**:
   - Uses `tea.ExecProcess` to hand off terminal to SSH subprocess
   - Scans `~/.ssh/*.pub` files to present available private keys
   - Constructs SSH command with selected host/user/identity (session-only)

5. **Health Checker (`internal/health`)**: TCP dial to port 22 on visible hosts. Toggled via `p` key. Only checks hosts currently in viewport.

6. **Platform Paths (`internal/platform`)**: Abstracts OS-specific paths:
   - Config: `~/.ssh/config` (Unix) or `%USERPROFILE%\.ssh\config` (Windows)
   - State: `~/.config/swiftssh/state.json` (Unix) or `%LOCALAPPDATA%\swiftssh\state.json` (Windows)

### Key Patterns & Constraints

- **Config as Source of Truth**: The SSH config file is append-only for new entries. Never modify existing entries except for auto-entry append.
- **Duplicate Hosts Preserved**: If `~/.ssh/config` has two `Host dev` blocks, both are treated as separate entries in the UI (no merging).
- **Backup Strategy**: Before any write operation, create `config.bak` (overwrites previous backup).
- **Minimalist CLI**: Use `os.Args` or `flag` package only—no Cobra/Viper to keep binary small.
- **Vim Keybindings**: Navigation is `j`/`k`, confirm with `Enter`, other bindings: `i` (identity), `p` (ping toggle), `/` (search).
- **ANSI Color Inheritance**: UI colors inherit from terminal theme; no custom theme override.
- **Windows Terminal Only**: Explicitly out of scope: legacy `cmd.exe` support. Target Windows Terminal (modern).

### Testing Strategy

- Unit test config parsing (including Include directives, duplicate hosts, magic comment parsing)
- Test state persistence (write/read JSON, upgrade logic)
- Test TUI model state transitions (search, selection, mode changes)
- Test SSH identity discovery
- Mock health check failures gracefully
- Platform path resolution tests

### String Handling & Magic Comments

- Magic comment format: `# @group Work, Personal` (single line per host)
- Parse groups as comma-separated tags for display/search
- Groups appear as metadata tags in UI next to hostname

## Git & Release Workflow

- Commit regularly with clear messages
- Tag releases as `v0.1.0`, `v0.2.0`, etc.
- Keep `main` branch as deployable release candidate
- Use feature branches for major work

## Notes for Future Claude Instances

- PRD.md is the single source of truth for functional requirements
- This project prioritizes **simplicity and speed** over feature richness
- Always confirm file paths are correct before read/write operations (cross-platform considerations)
- State file should be migrated gracefully if schema changes
- The TUI viewport should not block on health checks; checks happen async for visible hosts only
