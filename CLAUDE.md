# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwiftSSH is a single-binary Go CLI tool that provides an interactive, searchable TUI for managing SSH connections from `~/.ssh/config`. See PRD.md for detailed specifications and PLAN.md for the phased roadmap (Phases 0–7 are complete).

**Current feature set (as of Phase 7):**
- Scrollable, column-aligned host list with vim navigation (`j`/`k`)
- Live fuzzy search (any printable char enters search mode)
- In-place host editor (`Ctrl+E`) — edit all 6 host fields without touching the file manually
- SSH connection via `tea.ExecProcess` (proper TUI handoff)
- Connection frequency tracking ("frequent" hosts bubble to top)
- CLI SSH passthrough: `swiftssh user@host -p 2222` auto-saves unknown hosts to config and hands off to system `ssh`
- Magic comment groups: `# @group Work, Personal`
- Cross-platform paths (Unix + Windows Terminal)

**Key Dependencies:**
- `github.com/charmbracelet/bubbletea` v0.25.0 — TUI framework
- `github.com/charmbracelet/lipgloss` v0.9.1 — styled text rendering
- `github.com/sahilm/fuzzy` v0.1.1 — fuzzy search ranking

## Common Development Commands

### Building
```bash
go build -o swiftssh ./cmd/swiftssh
go build -ldflags="-s -w" -o swiftssh ./cmd/swiftssh   # release build
```

### Testing
```bash
go test ./...                                           # all packages
go test ./internal/config                               # single package
go test -v ./internal/tui -run TestEditForm             # single test
go test -cover ./...                                    # with coverage
```

### Code Quality
```bash
go vet ./...
go fmt ./...
go mod tidy -v
```

### Running
```bash
go run ./cmd/swiftssh
go run ./cmd/swiftssh --version
go run ./cmd/swiftssh user@host -p 2222    # SSH passthrough
```

## Architecture & Code Organization

### Directory Structure
```
.
├── cmd/
│   └── swiftssh/
│       └── main.go               # Entry point, flag parsing, SSH passthrough
├── internal/
│   ├── config/
│   │   ├── types.go              # Host struct, ParsedConfig
│   │   ├── types_test.go
│   │   ├── parser.go             # SSH config parser (Include, magic comments, CircularDetect)
│   │   ├── parser_test.go
│   │   ├── writer.go             # AppendHost, ReplaceHostBlock, IsKnownHost, buildHostBlock
│   │   └── writer_test.go
│   ├── state/
│   │   ├── state.go              # Load/Save (atomic), RecordConnection, FrequentHosts
│   │   └── state_test.go
│   ├── ssh/
│   │   ├── keys.go               # ScanPublicKeys, KeyLabel
│   │   ├── keys_test.go
│   │   ├── executor.go           # BuildArgs, ConnectCmd
│   │   └── executor_test.go
│   ├── tui/
│   │   ├── model.go              # Model struct, modes, editForm, applySearch, Update
│   │   ├── views.go              # renderList, renderEditForm, renderHeader, renderStatusBar
│   │   ├── keybindings.go        # handleNormalMode, handleSearchMode, handleEditMode
│   │   └── model_test.go
│   ├── health/
│   │   ├── check.go              # TCP health check stub (Phase 9+)
│   │   └── check_test.go
│   ├── platform/
│   │   ├── paths.go              # SSHConfigPath, StateFilePath, SSHKeyDir, EnsureDir
│   │   └── paths_test.go
│   └── testutil/
│       └── assert.go             # 18 shared assertion helpers (t.Helper-based)
├── go.mod                        # module github.com/srava/swiftssh, Go 1.22
├── go.sum
├── Makefile
├── .gitignore
├── PRD.md                        # Product requirements (source of truth)
├── PLAN.md                       # Phased roadmap with completion checklist
├── FEEDBACK.md                   # Phase 6 UX feedback log
├── TEST_BEST_PRACTICES.md        # Testing patterns and anti-patterns
└── TEST_SUITE_SUMMARY.md         # Test suite metrics and organization
```

### Host Struct (`internal/config/types.go`)

```go
type Host struct {
    Alias        string   // "Host" directive value, e.g. "dev"
    Hostname     string   // "Hostname" directive, e.g. "192.168.1.10"
    User         string   // "User" directive (may be empty)
    Port         string   // "Port" directive (defaults to "22" if absent)
    IdentityFile string   // "IdentityFile" directive, quotes stripped on parse
    Groups       []string // from magic comment "# @group Work, Personal"
    SourceFile   string   // which file this host was parsed from (Include support)
    LineStart    int      // 1-based line number of "Host <alias>" directive
}
```

### Core Components & Data Flow

#### 1. `cmd/swiftssh/main.go` — Entry Point
Before any flag parsing, raw args are checked with `looksLikeSSHArgs()`. If they look like an SSH invocation (contain `@` or recognized SSH flags), `runPassthrough()` is called instead of the TUI. This avoids `flag: provided but not defined` errors when users pass SSH-style args.

**Normal TUI flow:**
1. Parse `--version`/`-v` flag
2. `config.Parse(platform.SSHConfigPath())` → `[]Host`
3. `state.Load(platform.StateFilePath())` → `*State`
4. `tea.NewProgram(tui.New(hosts, st, statePath), tea.WithAltScreen()).Run()`

**SSH passthrough flow:**
1. `parseSSHTarget(args)` extracts destination, port, user, identity
2. If `user@host` form: split on `@`
3. `config.IsKnownHost()` → if unknown, `config.AppendHost()` + print to stderr
4. `exec.Command("ssh", args...)` with `cmd.Run()` — blocks until SSH exits

#### 2. `internal/config/parser.go` — SSH Config Parser
Line-by-line state machine. Key behaviours:
- `Host` keyword finalizes the previous block and starts a new one
- Groups assigned via `parseMagicComment(prevLine)` when `Host` keyword is encountered — `prevLine` is the mechanism; there is **no** direct `current.Groups` assignment inside the `#` branch (was a bug, now fixed)
- `Include` directives: tilde expansion → relative-to-configDir resolution → `filepath.Glob` → recursive `parseFile` with circular detection via `visited map[string]bool`
- `Host *` wildcard blocks are skipped
- Default Port `"22"` applied at finalization
- IdentityFile: surrounding quotes stripped on parse

#### 3. `internal/config/writer.go` — Config Writer
Two public write operations:

**`AppendHost(configPath, backupPath, h)`**: backup → open for append → write block. Skips the separator `\n` when the file is empty (avoids a leading blank line on first-ever write).

**`ReplaceHostBlock(h)`**: Used by the TUI edit form.
1. Read all lines, write backup
2. Locate block at `h.LineStart - 1` (0-based); lenient stale check: if that line is a `@group` comment rather than `Host`, look one ahead
3. Determine `magicStart` (includes preceding `@group` line if present)
4. `findBlockEnd` scans forward for next `Host` keyword, backs up past trailing blanks and magic comments
5. Splice in `buildHostBlock(h)` lines, join, atomic write via temp file + rename
6. Returns `(newLineStart int, lineDelta int, error)` — TUI uses these to update `LineStart` for all subsequent hosts in the same file

#### 4. `internal/tui/model.go` — TUI Model
Three modes:
- `modeNormal` — list navigation, search entry, edit entry, connect, quit
- `modeSearch` — live fuzzy filter, navigate within results, connect or edit while searching
- `modeEdit` — 6-field form editor for the selected host

`New(hosts, st, statePath)` sorts: frequent hosts (top N by connection count, descending) followed by remaining hosts (alphabetical). Deduplication for the frequent list uses composite key `alias + "\x00" + sourceFile`.

`applySearch(m *Model)` uses `github.com/sahilm/fuzzy` over `alias + " " + hostname + " " + groups`. Resets cursor and viewport to 0.

`Update()` handles `editSavedMsg` (returned async from `saveEditForm`): patches `allHosts[index]`, shifts `LineStart` for all subsequent hosts in the same SourceFile by `lineDelta`, re-applies current search filter.

#### 5. `internal/tui/keybindings.go` — Key Handlers

| Mode | Key | Action |
|------|-----|--------|
| Normal | `j` / `↓` | Move cursor down (wrap) |
| Normal | `k` / `↑` | Move cursor up (wrap) |
| Normal | `Enter` | Connect to selected host |
| Normal | `Ctrl+E` | Open edit form |
| Normal | any printable | Enter search mode |
| Normal | `Esc` / `Ctrl+C` | Quit |
| Search | printable | Append to query, re-filter |
| Search | `Backspace` | Delete last rune; empty → exit search |
| Search | `Ctrl+W` | Clear query, exit search |
| Search | `Esc` | Clear query, exit search |
| Search | `Enter` | Connect to selected |
| Search | `Ctrl+E` | Open edit form |
| Search | `↓` / `↑` | Navigate within filtered list |
| Edit | `↓` / `↑` | Cycle to next/previous field |
| Edit | printable | Append to active field |
| Edit | `Backspace` | Delete last rune in field |
| Edit | `Ctrl+U` | Clear entire field |
| Edit | `Enter` | Validate & save |
| Edit | `Esc` | Discard, return to normal |

#### 6. `internal/tui/views.go` — Rendering
- `renderHeader`: title + search query (`query█`) or dim `"Type to search"` hint
- `renderList`: column header (ALIAS / HOSTNAME / USER / GROUPS) + host rows; `colWidths()` computes dynamic column widths from content + minimums
- `renderRow`: selected row → reverse-video with `> ` prefix; non-selected → alias plain, hostname/user dim, groups colored
- `renderStatusBar`: transient `statusMsg` if set, otherwise key hint line
- `renderEditForm`: 6-row form, label (14-char padded, reverse if active) + value + `█` cursor; validation error replaces footer hints

#### 7. `internal/state/state.go` — Persistence
Atomic JSON writes: write to `path + ".tmp"` then `os.Rename`. `Load` returns `FirstRun: true` for new installs. `FrequentHosts` uses `sort.SliceStable` so tied-count hosts keep their original order.

#### 8. `internal/ssh/` — SSH Execution
`BuildArgs` constructs `[-i identity] [-p port] [-l user] alias`. `ConnectCmd` wraps `exec.Command("ssh", args...)`. Called via `tea.ExecProcess` in the TUI so the terminal is cleanly handed off.

#### 9. `internal/platform/paths.go` — OS Paths
| Function | Unix | Windows |
|----------|------|---------|
| `SSHConfigPath()` | `~/.ssh/config` | `%USERPROFILE%\.ssh\config` |
| `SSHConfigBackupPath()` | `~/.ssh/config.bak` | `%USERPROFILE%\.ssh\config.bak` |
| `StateFilePath()` | `~/.config/swiftssh/state.json` | `%LOCALAPPDATA%\swiftssh\state.json` |
| `SSHKeyDir()` | `~/.ssh` | `%USERPROFILE%\.ssh` |

## Key Patterns & Constraints

- **No Cobra/Viper**: `flag` package only — keeps binary small
- **Config append-only for new entries**: `AppendHost` appends; `ReplaceHostBlock` edits in-place with atomic writes and backup
- **Duplicate hosts preserved**: two `Host dev` blocks appear as two separate TUI entries (no merging)
- **Backup on every write**: `config.bak` written before any modification (overwrites previous backup)
- **Magic comments are the sole group mechanism**: `# @group Work, Personal` on the line immediately before `Host`. Parser assigns groups via `prevLine` only when a `Host` directive is encountered — never by direct assignment inside the comment branch
- **LineStart tracking**: every `Host` carries its 1-based line number. `ReplaceHostBlock` returns `(newLineStart, lineDelta)` and the TUI shifts all subsequent hosts' `LineStart` by `lineDelta` to keep them accurate without re-parsing
- **ANSI colors**: inherit from terminal theme via lipgloss — no custom theme override
- **Windows Terminal only**: legacy `cmd.exe` explicitly out of scope

## Testing Strategy & Patterns

- Subtests with `t.Run()` for clear organisation; `t.Helper()` in all helpers
- `internal/testutil/assert.go` provides 18 shared assertion helpers
- Config parser: edge cases include Include directives, circular includes, duplicate hosts, magic comment whitespace, IdentityFile quote stripping, LineStart accuracy
- Writer: `AppendHost` on empty vs. non-empty file; `ReplaceHostBlock` with add/remove groups, stale-line detection, lineDelta return values
- TUI model: cursor wrap, viewport advance/retreat, search filter, edit field navigation, save propagation, LineStart drift correction
- State: Load on missing file, round-trip, FrequentHosts ordering, atomic write with missing parent directory

## Notes for Future Claude Instances

- **PRD.md** is the functional source of truth; **PLAN.md** is the implementation roadmap
- **Phases 0–7 are complete.** Phases 8 (first-run UX, `--config`/`--no-frequent` flags), 9 (QA pass), and 10 (release) remain
- The identity picker overlay (`i` key, `modeIdentityPicker`) was scoped out — identity is set via the Edit form's `IdentityFile` field
- `ReplaceHostBlock` is the most complex function in the codebase — read `writer.go` carefully before touching it
- Always run `go vet ./...` and `go test ./...` after any change; both must be clean before committing
- When adding new Host fields: update `types.go`, `buildHostBlock`, `parseFile` keyword dispatch, `editForm` field constants, `renderEditForm` labels, and any tests that construct `Host` literals
- State file schema changes must handle missing fields gracefully (new `Load` returns safe defaults)
- Platform paths must be tested on both Unix and Windows — `paths.go` uses `os.UserHomeDir()` and `os.UserConfigDir()` from stdlib, never hardcoded paths
