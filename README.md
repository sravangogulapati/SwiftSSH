# sssh — Swift SSH

An interactive, searchable TUI for managing SSH connections from `~/.ssh/config`.

```
  sssh                    ALIAS          HOSTNAME            USER       GROUPS
─────────────────────── ─────────────── ─────────────────── ────────── ──────────
  > dev                  dev             192.168.1.10        ubuntu     Work
    prod                 prod.example.com                               Work
    pi                   10.0.0.5        pi                             Home
    bastion              bastion.co      deploy
```

## Features

- Scrollable, column-aligned host list
- Vim-style navigation (`j`/`k`)
- Live fuzzy search — just start typing
- In-place host editor (`Ctrl+E`) — edit all fields without touching the file
- Connection frequency tracking — frequent hosts bubble to the top
- Magic comment groups: `# @group Work, Personal`
- SSH passthrough: `sssh user@host -p 2222` auto-saves unknown hosts and hands off to system `ssh`
- `--config` flag to point at a non-default SSH config
- `--no-frequent` flag for flat alphabetical ordering
- Cross-platform: Unix and Windows Terminal

## Installation

### Download binary

**Linux / macOS:**
```sh
# Replace <version> and <os_arch> with e.g. v0.1.0 and linux_amd64
curl -L https://github.com/srava/swiftssh/releases/download/<version>/sssh_<os_arch> \
  -o /usr/local/bin/sssh && chmod +x /usr/local/bin/sssh
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/srava/swiftssh/releases/download/<version>/sssh_windows_amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Microsoft\WindowsApps\sssh.exe"
```

### go install

```sh
go install github.com/srava/swiftssh/cmd/sssh@latest
```

## Usage

```sh
sssh                         # open TUI
sssh --version               # print version
sssh --config ~/work/.ssh/config   # use a different SSH config
sssh --no-frequent           # alphabetical order, no frequency sort
sssh user@host               # SSH passthrough (saves unknown host, then connects)
sssh user@host -p 2222 -i ~/.ssh/id_ed25519
```

### Keybindings — Normal mode

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` | Connect to selected host |
| `Ctrl+E` | Open edit form |
| any printable char | Enter search mode |
| `Esc` / `Ctrl+C` | Quit |

### Keybindings — Search mode

| Key | Action |
|-----|--------|
| printable char | Append to search query |
| `Backspace` | Delete last character (empty query exits search) |
| `Ctrl+W` / `Esc` | Clear query, exit search |
| `↓` / `↑` | Navigate within filtered results |
| `Enter` | Connect to selected host |
| `Ctrl+E` | Open edit form for selected host |

### Keybindings — Edit form

| Key | Action |
|-----|--------|
| `↓` / `↑` | Next / previous field |
| printable char | Append to active field |
| `Backspace` | Delete last character |
| `Ctrl+U` | Clear entire field |
| `Enter` | Validate and save |
| `Esc` | Discard changes |

## Magic comment groups

Add a `# @group` comment on the line immediately before a `Host` directive to assign the host to one or more groups:

```
# @group Work, DevOps
Host prod
    Hostname prod.example.com
    User deploy
    IdentityFile ~/.ssh/id_ed25519

# @group Home
Host pi
    Hostname 10.0.0.5
    User pi
```

Groups are displayed in the TUI and searchable.

## SSH passthrough

When arguments look like an SSH invocation (contain `@` or SSH flags like `-p`, `-i`), `sssh` acts as a transparent wrapper:

```sh
sssh deploy@prod.example.com -p 2222
```

If the hostname is not already in your SSH config, `sssh` appends an entry automatically before connecting. Useful as a drop-in alias for `ssh`.

## CLI flags

| Flag | Description |
|------|-------------|
| `--version` / `-v` | Print version and exit |
| `--config <path>` | Use an alternative SSH config file |
| `--no-frequent` | Flat alphabetical order (skip frequency-based sorting) |

## First-run alias tip

Add to your shell profile for a one-letter shortcut:

```sh
alias s='sssh'
```

## Building from source

```sh
git clone https://github.com/srava/swiftssh.git
cd swiftssh

make build          # ./sssh
make test           # go test ./...
make release        # stripped binary, injects version from VERSION env var
make release-all    # linux + windows binaries

# Cross-compile manually:
GOOS=darwin GOARCH=arm64 go build -o sssh_darwin_arm64 ./cmd/sssh

# Inject version:
make release VERSION=v0.1.0
```

**Requirements:** Go 1.22+
