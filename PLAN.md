# SwiftSSH — Phased Project Plan

> **Source of truth:** `PRD.md`. Phases are independently buildable and testable.
> Check off items as you complete them. When asking Claude to build a phase, say:
> _"Build Phase N of the SwiftSSH plan"_
>
> **Testing:** See `TESTING.md` for step-by-step manual verification instructions for each phase.

---

## Final Directory Structure

```
swiftssh/
├── cmd/swiftssh/
│   └── main.go              # Entry point — wires all packages together
├── internal/
│   ├── platform/            # OS-aware file paths
│   │   ├── paths.go
│   │   └── paths_test.go
│   ├── config/              # SSH config parsing + writing
│   │   ├── types.go         # Host struct, Group type
│   │   ├── parser.go        # Parse ~/.ssh/config + Include
│   │   ├── writer.go        # Append-only writes + backup
│   │   └── parser_test.go
│   ├── state/               # Connection frequency persistence
│   │   ├── state.go
│   │   └── state_test.go
│   ├── ssh/                 # SSH key discovery + executor
│   │   ├── keys.go
│   │   ├── executor.go
│   │   └── keys_test.go
│   ├── health/              # TCP health checks
│   │   ├── check.go
│   │   └── check_test.go
│   └── tui/                 # Bubble Tea TUI
│       ├── model.go
│       ├── views.go
│       ├── keybindings.go
│       └── model_test.go
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── README.md
├── CLAUDE.md
└── PRD.md
```

---

## Phase 0 — Project Scaffold

**Goal:** Compilable Go module with a working Makefile, `.gitignore`, and a stub `main.go` that prints a version string. Establishes the full directory skeleton and dependency manifest.
**Depends on:** Nothing
**Verify:** `make build` succeeds and `./swiftssh --version` prints `swiftssh v0.1.0`

### Tasks

- [x] Create `go.mod` with module path `github.com/srava/swiftssh`, Go version `1.22`
- [x] Add dependencies to `go.mod`:
  - `github.com/charmbracelet/bubbletea` v0.25.0
  - `github.com/charmbracelet/lipgloss` v0.9.1 (for styled text rendering)
  - `github.com/sahilm/fuzzy` v0.1.1 (for fuzzy search)
- [x] Run `go mod tidy` to populate `go.sum` (40 lines)
- [x] Create `.gitignore` — exclude `swiftssh` binary, `swiftssh.exe`, `coverage.out`, `.DS_Store`
- [x] Create `Makefile` with targets:
  - [x] `build` — `go build -o swiftssh ./cmd/swiftssh`
  - [x] `run` — `go run ./cmd/swiftssh`
  - [x] `test` — `go test ./...`
  - [x] `test-cover` — `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
  - [x] `lint` — `go vet ./...`
  - [x] `fmt` — `go fmt ./...`
  - [x] `release` — `go build -ldflags="-s -w" -o swiftssh ./cmd/swiftssh`
  - [x] `clean` — remove `swiftssh`, `swiftssh.exe`, `coverage.out`
- [x] Create `cmd/swiftssh/main.go` stub:
  - [x] Parse `--version` / `-v` flags using the `flag` package
  - [x] Print `swiftssh v0.1.0` and exit if flag is set
  - [x] Otherwise print `"TUI coming soon"` (placeholder)
- [x] Create all `internal/` subdirectories with empty `.go` stub files (package declarations only) so the module compiles
- [x] Verify `go build` succeeds with zero errors (Note: Makefile requires `make` command; used `go build` directly)

---

## Phase 1 — Platform Paths + Core Types

**Goal:** OS-aware file path resolution and the central `Host` data structure used by all other packages.
**Depends on:** Phase 0
**Verify:** `go test ./internal/platform/` and `go test ./internal/config/` (types only) pass

### Tasks

#### `internal/platform/paths.go`
- [x] Implement `SSHConfigPath() string` — returns `~/.ssh/config` (Unix) or `%USERPROFILE%\.ssh\config` (Windows)
- [x] Implement `SSHConfigBackupPath() string` — same directory as config, filename `config.bak`
- [x] Implement `StateFilePath() string` — returns `~/.config/swiftssh/state.json` (Unix) or `%LOCALAPPDATA%\swiftssh\state.json` (Windows)
- [x] Implement `SSHKeyDir() string` — returns `~/.ssh` (Unix) or `%USERPROFILE%\.ssh` (Windows)
- [x] Use `os.UserHomeDir()`, `os.UserConfigDir()` from stdlib — no hardcoded paths (Note: `os.UserCacheDir()` not needed for Phase 1; used `UserConfigDir()` for state file as specified)
- [x] Export `EnsureDir(path string) error` — creates directory + parents if missing (for state file dir)

#### `internal/platform/paths_test.go`
- [x] Table-driven test: verify each path function returns a non-empty string (6 tests total)
- [x] Test `EnsureDir` creates a temp directory without error
- [x] Test `EnsureDir` is idempotent (no-op on subsequent calls)

#### `internal/config/types.go`
- [x] Define `Host` struct:
  ```go
  type Host struct {
      Alias      string   // The "Host" keyword value, e.g. "dev"
      Hostname   string   // The "Hostname" directive value, e.g. "192.168.1.10"
      User       string   // The "User" directive value (may be empty)
      Port       string   // The "Port" directive value (defaults to "22" if absent)
      Groups     []string // Parsed from "# @group Work, Personal"
      SourceFile string   // Which file this host was parsed from (for Include support)
  }
  ```
- [x] Define `ParsedConfig` struct: `Hosts []Host`, `SourceFile string`

---

## Phase 2 — SSH Config Parser

**Goal:** Correctly parse `~/.ssh/config` including recursive `Include` directives, duplicate hosts (preserved as-is), and magic group comments.
**Depends on:** Phase 1
**Verify:** `go test ./internal/config/` passes all table-driven tests

### Tasks

#### `internal/config/parser.go`
- [x] Implement `Parse(configPath string) ([]Host, error)`:
  - Read file line by line
  - Track current `Host` block; on each new `Host` keyword, save the previous and start a new one
  - Resolve `Include` directives: support glob patterns (e.g., `Include ~/.ssh/conf.d/*`), expand `~` to home dir, recurse into each matched file
  - Guard against circular includes (track visited paths)
  - When parsing a new `Host` block, look **back** for a magic comment on the line immediately preceding it: `# @group Tag1, Tag2`
  - Preserve duplicate `Host` blocks as separate `Host` entries (no merging)
  - Parse `Hostname`, `User`, `Port` directives within each block
  - Set `Host.SourceFile` to the file being parsed
  - Set default `Port` to `"22"` if not specified
  - Skip `Host *` (wildcard / defaults block) — do not add it to the list

#### `internal/config/parser_test.go`
- [x] Test basic single-host config file
- [x] Test multi-host config with all fields (Hostname, User, Port)
- [x] Test magic comment parsing: `# @group Work, Personal` → `Groups: ["Work", "Personal"]`
- [x] Test magic comment with extra whitespace around commas
- [x] Test duplicate `Host dev` blocks appear as two separate entries
- [x] Test `Include` directive with a relative path
- [x] Test `Include` directive with a glob pattern
- [x] Test recursive `Include` (A includes B which includes C)
- [x] Test circular `Include` protection (no infinite loop)
- [x] Test `Host *` block is excluded from results
- [x] Test graceful handling of a missing `Include`-d file (log warning, continue)

---

## Phase 3 — State Manager

**Goal:** Persist connection frequency to JSON and expose "Frequent" host ordering. Detect first run.
**Depends on:** Phase 1
**Verify:** `go test ./internal/state/` passes

### Tasks

#### `internal/state/state.go`
- [x] Define `State` struct:
  ```go
  type State struct {
      Connections map[string]int `json:"connections"` // key: host alias, value: count
      FirstRun    bool           `json:"first_run"`
  }
  ```
- [x] Implement `Load(path string) (*State, error)`:
  - If file doesn't exist, return a fresh `State{FirstRun: true, Connections: map[string]int{}}`
  - Unmarshal JSON; tolerate unknown fields (use `json.Decoder` with `DisallowUnknownFields(false)`)
- [x] Implement `Save(path string, s *State) error`:
  - Use `platform.EnsureDir` on the parent directory before writing
  - Write atomically: write to a temp file, then rename over the target
- [x] Implement `RecordConnection(s *State, alias string)` — increment counter for alias
- [x] Implement `FrequentHosts(s *State, hosts []config.Host, n int) []config.Host` — return top `n` hosts by connection count, preserving order for ties

#### `internal/state/state_test.go`
- [x] Test `Load` on non-existent file returns `FirstRun: true`
- [x] Test `Load` → `Save` → `Load` round-trip preserves data
- [x] Test `RecordConnection` increments count correctly
- [x] Test `FrequentHosts` returns correct top-N ordering
- [x] Test `FrequentHosts` with fewer hosts than `n` (returns all)
- [x] Test `Save` with missing parent directory (should create it)

---

## Phase 4 — TUI Foundation

**Goal:** A working Bubble Tea application that displays the host list with vim navigation and viewport scrolling. No SSH connection or search yet — just the core UI loop.
**Depends on:** Phases 1, 2, 3
**Verify:** `go run ./cmd/swiftssh` shows a scrollable list; `j`/`k` moves cursor; `q`/`Ctrl+C` quits

### Tasks

#### `internal/tui/model.go`
- [x] Define UI modes as `iota`:
  ```go
  type mode int
  const (
      modeNormal mode = iota
      modeSearch
      modeIdentityPicker
  )
  ```
- [x] Define `Model` struct:
  ```go
  type Model struct {
      allHosts     []config.Host  // Full unfiltered host list (frequent first, then rest)
      filtered     []config.Host  // Current displayed list (after search filtering)
      cursor       int            // Index into filtered
      viewport     int            // Index of top-visible item
      viewHeight   int            // Terminal rows available for the list
      mode         mode
      searchQuery  string
      state        *state.State
      statePath    string
      // Identity picker fields (Phase 5)
      // Health check fields (Phase 6)
  }
  ```
- [x] Implement `New(hosts []config.Host, st *state.State, statePath string) Model`:
  - Sort: frequent hosts first (from `state.FrequentHosts`), then remaining hosts alphabetically
  - Initialize `filtered` as a copy of `allHosts`
- [x] Implement `Init() tea.Cmd` — no-op for now
- [x] Implement `Update(msg tea.Msg) (tea.Model, tea.Cmd)` — handle `tea.WindowSizeMsg` to set `viewHeight`
- [x] Implement `View() string` — render the list (delegates to `views.go`)

#### `internal/tui/keybindings.go`
- [x] Handle `tea.KeyMsg` in `Update`:
  - [x] `q`, `ctrl+c` → `tea.Quit`
  - [x] `j`, `down` → move cursor down (with wrap-around at bottom)
  - [x] `k`, `up` → move cursor up (with wrap-around at top)
  - [x] Viewport scrolling: when cursor moves past bottom of viewport, advance viewport; when cursor moves above top, retreat viewport
  - [x] `Enter` → placeholder for Phase 5 (do nothing for now)
  - [x] `/` → placeholder for Phase 4b search (do nothing for now)
  - [x] `i` → placeholder for Phase 5 identity picker (do nothing for now)
  - [x] `p` → placeholder for Phase 6 health checks (do nothing for now)

#### `internal/tui/views.go`
- [x] Implement `renderList(m Model) string`:
  - [x] Show only `viewHeight` items starting from `m.viewport`
  - [x] Selected item: render with `>` prefix and bold/reverse style using `lipgloss`
  - [x] Each row format: `  [alias]  [hostname]  [user@]  [groups...]`
  - [x] Group tags rendered as `[tag]` in a muted color
  - [x] Status bar at the bottom: `[N hosts] | q: quit | /: search | Enter: connect`
- [x] Implement `renderHeader(m Model) string`:
  - [x] Show `SwiftSSH` title
  - [x] If `m.mode == modeSearch`, show the current search query as a prompt

#### `internal/tui/model_test.go`
- [x] Test cursor wraps correctly at list boundaries
- [x] Test viewport advances when cursor moves below visible area
- [x] Test viewport retreats when cursor moves above visible area
- [x] Test `New` sorts frequent hosts to top

#### `cmd/swiftssh/main.go` (update)
- [x] Load hosts from `platform.SSHConfigPath()` using `config.Parse`
- [x] Load state from `platform.StateFilePath()` using `state.Load`
- [x] Launch TUI with `tea.NewProgram(tui.New(hosts, st, statePath)).Run()`
- [x] Print error and exit 1 if config cannot be parsed

---

## Phase 5 — SSH Execution, Identity Picker & Auto-Entry

**Goal:** Actually connect to hosts via SSH using `tea.ExecProcess`. Identity picker overlay. Auto-entry for new IPs with `config.bak` backup.
**Depends on:** Phase 4
**Verify:** Selecting a host in the TUI launches SSH; `i` key shows key picker; connecting to a new IP appends to config

### Tasks

#### `internal/ssh/keys.go`
- [x] Implement `ScanPublicKeys(sshDir string) ([]string, error)`:
  - Glob `~/.ssh/*.pub` and return a list of matching key paths (strip `.pub` for the actual private key path)
  - Filter out entries where the corresponding private key file does not exist
- [x] Implement `KeyLabel(pubKeyPath string) string` — strip directory prefix and `.pub` suffix for display

#### `internal/ssh/keys_test.go`
- [x] Test `ScanPublicKeys` against a temp directory with mock `.pub` files
- [x] Test it excludes `.pub` files with no corresponding private key
- [x] Test `KeyLabel` strips path and extension correctly

#### `internal/ssh/executor.go`
- [x] Implement `BuildArgs(host config.Host, identity string) []string`:
  - Returns SSH arguments as a slice (never concatenated into a string)
  - Includes `-i <identity>` if identity is non-empty
  - Includes `-p <port>` if port is non-default
  - Includes `-l <user>` if user is set
  - Final argument is the host alias (or `user@hostname` format)
- [x] Implement `ConnectCmd(host config.Host, identity string) *exec.Cmd` — wraps `exec.Command("ssh", BuildArgs(...)...)`
- [x] Expose `ExecProcess(host config.Host, identity string) tea.ExecCommand` — returns a `tea.ExecProcess` compatible command (via `tea.ExecProcess` in keybindings)

#### `internal/config/writer.go`
- [x] Implement `IsKnownHost(hosts []config.Host, hostname string) bool` — returns true if any host has matching Hostname
- [x] Implement `AppendHost(configPath, backupPath string, h config.Host) error`:
  1. Copy `configPath` to `backupPath` (overwrite backup)
  2. Append a newline + formatted Host block to `configPath`:
     ```
     Host <alias>
         Hostname <hostname>
         User <user>
     ```
  3. Return error if file operations fail

#### TUI — Identity Picker (update `internal/tui/`)
- [x] Add `availableKeys []string` and `keyPickerCursor int` to `Model`
- [x] Add `selectedIdentity string` to `Model` (session-only, never persisted)
- [x] On `i` key press in `modeNormal`:
  - Call `ssh.ScanPublicKeys(platform.SSHKeyDir())`
  - If no keys found: show status message "No SSH keys found in ~/.ssh"
  - If keys found: set `mode = modeIdentityPicker`, populate `availableKeys`
- [x] In `modeIdentityPicker`:
  - `j`/`k` navigate the key list
  - `Enter` sets `selectedIdentity` and returns to `modeNormal`
  - `Esc` returns to `modeNormal` without changing identity
- [x] In `views.go`: implement `renderIdentityPicker(m Model) string` — renders as an overlay list

#### TUI — Connect on Enter (update `internal/tui/keybindings.go`)
- [x] On `Enter` in `modeNormal`:
  1. Get the selected `config.Host` from `m.filtered[m.cursor]`
  2. Call `state.RecordConnection` and save state
  3. Check `config.IsKnownHost` — if not known, call `config.AppendHost` to auto-entry
  4. Return `tea.ExecProcess(ssh.ConnectCmd(host, m.selectedIdentity), connectCallback)`
  5. On callback return, clear `selectedIdentity`

#### Tests
- [x] Test `ssh.BuildArgs` with identity set vs. empty
- [x] Test `ssh.BuildArgs` with non-default port
- [x] Test `config.AppendHost` writes correct block and creates backup
- [x] Test `config.IsKnownHost` returns correct results
- [x] Test identity picker mode transitions in TUI model

---

## Phase 6 — User Feedback & Iteration

**Goal:** Gather user feedback on the MVP (core features working: list, search, SSH execution, identity picker). Identify UX issues, missing features, or design improvements. Implement feedback and potentially introduce new development phases as needed.
**Depends on:** Phase 5
**Verify:** Feedback documented; new issues/features tracked; roadmap updated with any new phases

### Tasks

#### Feedback Collection
- [x] Document current MVP state (what works, what doesn't) — see `FEEDBACK.md`
- [x] Collect feedback on UI/UX, feature gaps, quality of life

#### Feedback Processing & Triage
- [x] Documented all feedback in `FEEDBACK.md`
- [x] Categorized and triaged all items (3 High Priority quick fixes)

#### Implementation Planning
- [x] F1 (High) — Column-aligned table layout: Quick Fix applied
- [x] F2 (High) — `i: identity` missing from status bar: Quick Fix applied
- [x] F3 (High) — CLI SSH passthrough (`swiftssh user@host`): Quick Fix applied

#### Documentation Updates
- [x] Updated `PLAN.md` checklist

---

## Phase 7 — Fuzzy Search

**Goal:** `/` activates search mode; typing filters the host list in real-time using fuzzy matching against alias, hostname, and groups.
**Depends on:** Phase 4, Phase 6 (feedback iteration if search changes required)
**Verify:** Pressing `/` shows search prompt; typing filters the list; `Esc`/`Enter` exits search

### Tasks

#### `internal/tui/model.go` (update)
- [x] Add `searchQuery string` to `Model`
- [x] Implement `applySearch(m *Model)`:
  - If `searchQuery == ""`, set `filtered = allHosts` and return
  - Build a single searchable string per host: `alias + " " + hostname + " " + strings.Join(groups, " ")`
  - Use `github.com/sahilm/fuzzy` to rank matches against the query
  - Update `m.filtered` with matched hosts in rank order
  - Reset `m.cursor = 0` and `m.viewport = 0` after filtering

#### `internal/tui/keybindings.go` (update)
- [x] On `/` in `modeNormal`: set `mode = modeSearch`, clear `searchQuery`
- [x] In `modeSearch`, on `tea.KeyMsg`:
  - Printable chars: append to `searchQuery`, call `applySearch`
  - `Backspace`: trim last rune from `searchQuery`, call `applySearch`
  - `Esc`: clear `searchQuery`, `applySearch`, set `mode = modeNormal`
  - `Enter`: set `mode = modeNormal` (keep current filtered results)
  - `ctrl+c`: quit

#### `internal/tui/views.go` (update)
- [x] In `renderHeader`: show `/ <query>█` search prompt when in `modeSearch`
- [ ] Highlight matched characters in rendered host rows (optional, nice-to-have for MVP)

#### Tests
- [x] Test `applySearch` with empty query returns full host list
- [x] Test `applySearch` filters by alias
- [x] Test `applySearch` filters by hostname
- [x] Test `applySearch` filters by group tag
- [x] Test search resets cursor and viewport to 0

---

## Phase 8 — Health Checks

**Goal:** `p` toggle fires async TCP dials to port 22 for viewport-visible hosts only. Results appear as status indicators without blocking the TUI.
**Depends on:** Phase 4, Phase 6 (feedback iteration if health check UX changes required)
**Verify:** Pressing `p` shows TCP status icons next to hosts in the viewport; off-screen hosts show nothing

### Tasks

#### `internal/health/check.go`
- [ ] Implement `CheckHost(hostname, port string, timeout time.Duration) bool`:
  - TCP dial to `hostname:port` (use `port` from `config.Host`, default `"22"`)
  - Return `true` if connection succeeds (close immediately after)
  - Return `false` on timeout or refusal
  - Use `net.DialTimeout`
- [ ] Define `Result` struct: `Alias string`, `Reachable bool`
- [ ] Implement `CheckHosts(hosts []config.Host, results chan<- Result)`:
  - Spawn one goroutine per host
  - Each goroutine calls `CheckHost` and sends to `results` channel
  - Use 2-second timeout per dial

#### `internal/health/check_test.go`
- [ ] Test `CheckHost` against `localhost:22` (skip if port not open with `t.Skip`)
- [ ] Test `CheckHost` against a deliberately closed port returns `false`
- [ ] Test `CheckHosts` sends one result per host

#### TUI — Health Check Integration (update `internal/tui/`)
- [ ] Add `pingEnabled bool` and `pingResults map[string]bool` to `Model`
- [ ] Define a `pingResultMsg` custom tea message type: `type pingResultMsg health.Result`
- [ ] On `p` key: toggle `pingEnabled`; if enabling, trigger `firePingChecks(m)`
- [ ] Implement `firePingChecks(m Model) tea.Cmd`:
  - Get visible hosts: `m.filtered[m.viewport : min(m.viewport+m.viewHeight, len(m.filtered))]`
  - Spawn goroutine calling `health.CheckHosts(visibleHosts, results)`
  - Return a `tea.Cmd` that listens for `pingResultMsg` and updates the model
- [ ] In `Update`, handle `pingResultMsg`: update `m.pingResults[msg.Alias] = msg.Reachable`
- [ ] Re-fire ping checks when viewport scrolls (only if `pingEnabled`)
- [ ] In `views.go`: render `●` (green/reachable) or `○` (grey/unreachable) before each host when `pingEnabled`

#### Tests
- [ ] Test ping toggle transitions model state correctly
- [ ] Test viewport change triggers new ping batch
- [ ] Test `pingResultMsg` updates `pingResults` map

---

## Phase 9 — First-Run UX & CLI Polish

**Goal:** Detect first run and print alias suggestion. Wire all CLI flags. Ensure the app feels complete.
**Depends on:** Phases 3, 5, Phase 6 (feedback may inform alias/flag improvements)
**Verify:** First `./swiftssh` run prints alias suggestion; subsequent runs go straight to TUI; `--help` and `--version` work

### Tasks

#### `cmd/swiftssh/main.go` (final)
- [ ] Define and parse flags using `flag` package:
  - `--version` / `-v` — print `swiftssh v0.1.0` and exit
  - `--help` / `-h` — print usage summary and exit
  - `--config <path>` — override SSH config path
  - `--no-frequent` — disable the "Frequent" section (show flat alphabetical list)
- [ ] Implement first-run detection:
  - If `state.FirstRun == true`: print alias suggestion block to stderr and set `state.FirstRun = false`, save state
  - Alias suggestions: `alias s='swiftssh'` for bash/zsh, `Set-Alias s swiftssh` for PowerShell
- [ ] Handle config parse error gracefully: print `"Error: could not parse SSH config: <err>"` to stderr, exit 1
- [ ] Handle empty host list: print `"No hosts found in <path>. Add entries to your SSH config."` and exit 0

#### First-Run Output Format
- [ ] Print to stderr (not stdout, so piping still works)
- [ ] Format:
  ```
  Welcome to SwiftSSH!
  To use the 's' alias, add one of the following to your shell profile:

    bash/zsh:   alias s='swiftssh'
    fish:       alias s swiftssh
    PowerShell: Set-Alias s swiftssh

  (SwiftSSH will not modify your shell profile automatically.)
  ```

#### Tests
- [ ] Test `--version` flag output format
- [ ] Test first-run detection correctly updates state
- [ ] Test empty host list exits with message (not error code)
- [ ] Test `--config` flag passes alternative path to parser

---

## Phase 10 — Testing & QA

**Goal:** Bring test coverage to a high standard across all packages. Fix edge cases discovered during feedback iteration.
**Depends on:** All prior phases (especially Phase 6 feedback findings)
**Verify:** `make test-cover` shows >80% coverage across all packages; `go vet ./...` is clean

### Tasks

#### Coverage Pass
- [ ] Run `make test-cover` and identify uncovered paths in each package
- [ ] `internal/config/parser_test.go` — add tests for:
  - Config file with Windows-style CRLF line endings
  - Multi-word group names (e.g., `# @group My Work, Client Projects`)
  - Host with no Hostname directive (should still parse, Hostname field empty)
  - Very large config file (1,000+ hosts) — verify parse completes in < 1 second
- [ ] `internal/tui/model_test.go` — add tests for:
  - `View()` does not panic with an empty host list
  - `View()` does not panic with cursor at the very last host
  - State is correctly recorded on successful connection (mock SSH exec)
- [ ] `internal/state/state_test.go` — add test for corrupted JSON file (returns fresh state, no crash)
- [ ] `internal/ssh/keys_test.go` — add test for directory with no `.pub` files

#### Code Quality
- [ ] Run `go vet ./...` — fix all warnings
- [ ] Run `go fmt ./...` — ensure consistent formatting
- [ ] Check all exported functions/types have doc comments
- [ ] Verify no `panic()` calls in production paths (only allowed in tests via `t.Fatal`)

---

## Phase 11 — Release Preparation

**Goal:** README, GitHub Actions CI, cross-platform release builds, and GitHub release artifacts.
**Depends on:** Phase 10
**Verify:** CI passes on push; `goreleaser` (or manual build matrix) produces binaries for all targets

### Tasks

#### `README.md`
- [ ] Write README with sections:
  - Installation (binary download + `go install`)
  - Usage (keybindings table)
  - SSH config magic comment syntax
  - First-run alias setup
  - Building from source

#### GitHub Actions (`.github/workflows/`)
- [ ] Create `ci.yml` — runs on every push and PR:
  - `go vet ./...`
  - `go test ./...`
  - Target OS matrix: `ubuntu-latest`, `macos-latest`, `windows-latest`
- [ ] Create `release.yml` — runs on tag push (`v*`):
  - Build binaries for: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
  - Upload as GitHub Release assets
  - Use `go build -ldflags="-s -w -X main.version=${{ github.ref_name }}"` to embed version

#### Version Embedding
- [ ] Add `var version = "dev"` to `cmd/swiftssh/main.go`
- [ ] Pass `-X main.version=<tag>` in release build ldflags so `--version` shows real tag

#### Final Checks
- [ ] `make release` produces a binary under 10MB
- [ ] Test binary on clean macOS, Linux, and Windows Terminal
- [ ] Verify `--version` output matches the git tag
- [ ] Tag `v0.1.0` and push to trigger release workflow

---

## Phase Summary

| # | Phase | Key Deliverable | Testable? |
|---|-------|-----------------|-----------|
| 0 | Scaffold | Compilable module + Makefile | `make build` works |
| 1 | Platform + Types | OS path resolution + Host struct | `go test ./internal/platform/` |
| 2 | Config Parser | Full SSH config parsing | `go test ./internal/config/` |
| 3 | State Manager | Frequency tracking JSON | `go test ./internal/state/` |
| 4 | TUI Foundation | Scrollable host list, vim nav | Run the app, navigate with j/k |
| 5 | SSH Exec + Identity | Real SSH connections, key picker | Connect to a real host |
| 6 | User Feedback & Iteration | Feedback documented, UX refined | Manual testing + feature triage |
| 7 | Fuzzy Search | Live filtered search | Press `/`, type to filter |
| 8 | Health Checks | Async TCP ping per visible host | Press `p`, see status icons |
| 9 | First-Run + CLI | Alias suggestion, all flags | `--version`, `--help`, first run |
| 10 | Testing & QA | >80% coverage, clean vet | `make test-cover` |
| 11 | Release | CI pipeline + GitHub Release | Tag v0.1.0 |
