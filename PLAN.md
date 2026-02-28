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
├── cmd/sssh/
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
**Verify:** `make build` succeeds and `./sssh --version` prints `sssh v0.1.0`

### Tasks

- [x] Create `go.mod` with module path `github.com/srava/swiftssh`, Go version `1.22`
- [x] Add dependencies to `go.mod`:
  - `github.com/charmbracelet/bubbletea` v0.25.0
  - `github.com/charmbracelet/lipgloss` v0.9.1 (for styled text rendering)
  - `github.com/sahilm/fuzzy` v0.1.1 (for fuzzy search)
- [x] Run `go mod tidy` to populate `go.sum` (40 lines)
- [x] Create `.gitignore` — exclude `sssh` binary, `sssh.exe`, `coverage.out`, `.DS_Store`
- [x] Create `Makefile` with targets:
  - [x] `build` — `go build -o sssh ./cmd/sssh`
  - [x] `run` — `go run ./cmd/sssh`
  - [x] `test` — `go test ./...`
  - [x] `test-cover` — `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
  - [x] `lint` — `go vet ./...`
  - [x] `fmt` — `go fmt ./...`
  - [x] `release` — `go build -ldflags="-s -w" -o sssh ./cmd/sssh`
  - [x] `clean` — remove `sssh`, `sssh.exe`, `coverage.out`
- [x] Create `cmd/sssh/main.go` stub:
  - [x] Parse `--version` / `-v` flags using the `flag` package
  - [x] Print `sssh v0.1.0` and exit if flag is set
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
**Verify:** `go run ./cmd/sssh` shows a scrollable list; `j`/`k` moves cursor; `q`/`Ctrl+C` quits

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

#### `cmd/sssh/main.go` (update)
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
  2. Append formatted Host block (skips leading newline when file is empty)
  3. Return error if file operations fail
- [x] Implement `ReplaceHostBlock(h Host) (newLineStart int, lineDelta int, error)`:
  1. Write backup to `h.SourceFile + ".bak"` before any modification
  2. Locate block by `h.LineStart`; lenient stale check handles @group off-by-one
  3. Replace block lines (including any preceding magic comment) with `buildHostBlock(h)`
  4. Write atomically via temp file + rename
  5. Return updated 1-based `LineStart` of Host directive and line count delta
- [x] Helper `buildHostBlock(h)` — serialises Host to SSH config text; prepends `# @group` if groups set
- [x] Helpers `splitLines`, `findBlockEnd`, `parseHostLine` for in-place editing

#### TUI — Connect on Enter (update `internal/tui/keybindings.go`)
- [x] On `Enter` in `modeNormal`:
  1. Get the selected `config.Host` from `m.filtered[m.cursor]`
  2. Call `state.RecordConnection` and save state
  3. Check `config.IsKnownHost` — if not known, call `config.AppendHost` to auto-entry
  4. Return `tea.ExecProcess(ssh.ConnectCmd(host, ""), connectCallback)`

> **Note:** `modeIdentityPicker` (i-key overlay) was scoped out after Phase 5 feedback — identity file is set via the Edit form (`Ctrl+E → IdentityFile field`) instead.

#### Tests
- [x] Test `ssh.BuildArgs` with identity set vs. empty
- [x] Test `ssh.BuildArgs` with non-default port
- [x] Test `config.AppendHost` writes correct block and creates backup
- [x] Test `config.AppendHost` on empty file produces no leading blank line
- [x] Test `config.IsKnownHost` returns correct results
- [x] Test `config.ReplaceHostBlock` basic replace, magic comment add/remove, stale-line detection, backup creation, newLineStart & lineDelta return values

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

#### Quick Fixes Applied
- [x] F1 (High) — Column-aligned table layout: dynamic `colWidths()`, padded rows
- [x] F2 (High) — Status bar updated: `Enter: connect | Ctrl+E: edit | esc: quit`
- [x] F3 (High) — CLI SSH passthrough (`sssh user@host`): `runPassthrough()` + `looksLikeSSHArgs()` + `parseSSHTarget()`

#### Edit Mode — In-Place Host Editor (`Ctrl+E`)

This was the largest feature delivered from Phase 6 feedback. Hosts can now be edited directly without touching the SSH config file manually.

##### `internal/config/writer.go`
- [x] `ReplaceHostBlock(h Host)` — atomic in-place rewrite of a host block (see Phase 5 tasks above)

##### `internal/config/parser.go` — Bug Fixes Applied
- [x] Fix: group leak — `# @group` before host 2 was being assigned to host 1 via direct `current.Groups` assignment; removed the direct assignment, prevLine mechanism is now the sole source
- [x] Fix: `AppendHost` leading blank line — empty file no longer gets a blank first line before the host block

##### `internal/tui/model.go`
- [x] Add `modeEdit` to `mode` iota (alongside `modeNormal`, `modeSearch`)
- [x] Add `editField` iota: `fieldAlias`, `fieldHostname`, `fieldUser`, `fieldPort`, `fieldIdentityFile`, `fieldGroups` (+ `fieldCount = 6`)
- [x] Add `editForm` struct: `original Host`, `fields [6]string`, `activeField editField`, `statusMsg string`
- [x] Add `editSavedMsg` struct: `updated Host`, `index int`, `lineDelta int`, `originalLineStart int`, `sourceFile string`
- [x] Add `edit *editForm` field to `Model` (nil unless in edit mode)
- [x] `Update()` handles `editSavedMsg`: patches `allHosts[index]`, shifts `LineStart` for all subsequent hosts in the same file if `lineDelta ≠ 0`, re-applies search, returns to `modeNormal`

##### `internal/tui/keybindings.go`
- [x] `Ctrl+E` in `modeNormal` and `modeSearch` → calls `openEditForm(m)`
- [x] `openEditForm(m)`: validates `LineStart ≠ 0`, copies all 6 fields, converts groups slice to comma-separated string, sets `activeField = fieldAlias`
- [x] `handleEditMode(m, msg)`: full key handler for edit form
  - `↓` / `↑`: cycle activeField with wraparound
  - `Backspace`: delete last rune in active field
  - `Ctrl+U`: clear entire active field
  - `Enter`: call `saveEditForm(m)`
  - `Esc`: discard, return to `modeNormal`
  - `Ctrl+C`: quit
  - Any printable rune: append to active field
- [x] `saveEditForm(m)`: trims all fields; validates Alias & Hostname non-empty; parses groups as comma-split; defaults empty Port to "22"; finds host by `(SourceFile, LineStart)`; calls `config.ReplaceHostBlock()`; emits `editSavedMsg` with `newLineStart` & `lineDelta`

##### `internal/tui/views.go`
- [x] `renderEditForm(m)`: 6-field form with label column (14-char padded), active field highlighted in reverse video, cursor `█` on active field value, footer key hints; shows validation error instead of hints on failure

##### Tests
- [x] `model_test.go`: `TestEditForm_FieldNavigation` — ↓/↑ cycles all 6 fields
- [x] `model_test.go`: `TestEditForm_SaveUpdatesHost` — save propagates to `allHosts` and resets mode
- [x] `model_test.go`: `TestEditForm_ValidationRejectsEmptyAlias` — empty Alias keeps edit mode open
- [x] `model_test.go`: `TestEditForm_LineStartPropagation` — after saving host that gains a group (lineDelta=+1), subsequent host's LineStart is incremented

---

## Phase 7 — Fuzzy Search

**Goal:** Any printable keypress activates search mode; typing filters the host list in real-time using fuzzy matching against alias, hostname, and groups.
**Depends on:** Phase 4, Phase 6 (feedback iteration if search changes required)
**Verify:** Typing any character shows search prompt; typing filters the list; `Esc`/`Backspace`-to-empty exits search; `Enter` connects to selected host

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
- [x] Any printable rune in `modeNormal`: enter `modeSearch` with that character as the first query character
- [x] In `modeSearch`, on `tea.KeyMsg`:
  - Printable chars: append to `searchQuery`, call `applySearch`
  - `Backspace`: trim last rune; if query becomes empty, return to `modeNormal`
  - `ctrl+w`: clear `searchQuery`, return to `modeNormal`
  - `Esc`: clear `searchQuery`, `applySearch`, set `mode = modeNormal`
  - `Enter`: connect to selected host (same as `Enter` in normal mode)
  - `down`/`up`: navigate list without exiting search
  - `ctrl+c`: quit

#### `internal/tui/views.go` (update)
- [x] In `renderHeader`: show `<query>█` prompt when in `modeSearch`; show `"Type to search"` hint in `modeNormal`

#### Tests
- [x] Test `applySearch` with empty query returns full host list
- [x] Test `applySearch` filters by alias
- [x] Test `applySearch` filters by hostname
- [x] Test `applySearch` filters by group tag
- [x] Test search resets cursor and viewport to 0

---

## Phase 8 — First-Run UX & CLI Polish

**Goal:** Wire remaining CLI flags. Ensure the app feels complete.
**Depends on:** Phases 3, 5, Phase 6
**Verify:** `./sssh --version` prints `sssh v0.1.0`; `--config` and `--no-frequent` work

### Already Done
- [x] `--version` / `-v` — prints `sssh v0.1.0` and exits
- [x] `--help` / `-h` — handled automatically by `flag.Parse()`
- [x] Config parse error: prints `"Error: could not parse SSH config: <err>"` to stderr, exits 1
- [x] SSH passthrough: `sssh user@host [flags]` — auto-saves unknown host to config, hands off to system `ssh`

### Tasks

#### `cmd/sssh/main.go` (update)
- [x] Add `--config <path>` flag — override SSH config path (pass to `config.Parse` and `config.AppendHost`)
- [x] Add `--no-frequent` flag — skip `state.FrequentHosts` ordering, show flat alphabetical list instead
- [x] Handle empty host list: print `"No hosts found in <path>. Add entries to your SSH config."` and exit 0

#### Tests
- [x] Test `--config` flag passes alternative path to parser (`cmd/sssh/main_test.go` — `TestExtractConfigFlag`)
- [x] Test `--no-frequent` flag returns flat alphabetical list (`TestNewNoFrequent_FlatAlphabeticalOrder`)
- [x] Test empty host list exits 0 with message (manual: `./sssh --config /tmp/empty_ssh`)

---

## Phase 9 — Testing & QA

**Goal:** Bring test coverage to a high standard across all packages. Fix edge cases discovered during feedback iteration.
**Depends on:** All prior phases
**Verify:** `go test ./...` passes; `go vet ./...` is clean; coverage is solid across core packages

### Tasks

#### Coverage Pass
- [x] Run `go test -cover ./...` and identify uncovered paths in each package
- [x] `internal/config/parser_test.go` — add tests for:
  - Config file with Windows-style CRLF line endings
  - Multi-word group names (e.g., `# @group My Work, Client Projects`)
  - Host with no Hostname directive (should still parse, Hostname field empty)
  - Very large config file (1,000+ hosts) — verify parse completes in < 1 second
- [x] `internal/config/writer_test.go` — add tests for:
  - `AppendHost` when config file does not yet exist
- [x] `internal/tui/model_test.go` — add tests for:
  - `View()` does not panic with an empty host list
  - `View()` does not panic with cursor at the very last host
  - Backspace-to-empty in search mode returns to `modeNormal` (already existed)
  - `ctrl+w` in search mode clears query and returns to `modeNormal` (already existed)
  - `ctrl+i` — skipped; identity picker was scoped out, no binding exists
- [x] `internal/state/state_test.go` — add test for corrupted JSON file (returns fresh state, no crash)
- [x] `internal/ssh/keys_test.go` — add test for directory with no `.pub` files

#### Code Quality
- [x] Run `go vet ./...` — clean
- [x] Run `go fmt ./...` — consistent formatting confirmed
- [x] Verify no `panic()` calls in production paths

---

## Phase 10 — Release Preparation

**Goal:** README, GitHub Actions CI, cross-platform release builds, and GitHub release artifacts.
**Depends on:** Phase 9
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
  - Upload as GitHub Release assets (named `sssh`, `sssh.exe`)
  - Use `go build -ldflags="-s -w -X main.version=${{ github.ref_name }}" -o sssh ./cmd/sssh` to embed version

#### Version Embedding
- [ ] Change `const Version = "0.1.0"` to `var version = "dev"` in `cmd/sssh/main.go` (ldflags requires a `var`, not a `const`)
- [ ] Update `--version` output to use `version` var instead of `Version` const
- [ ] Pass `-X main.version=<tag>` in release build ldflags so `--version` shows real tag

#### Final Checks
- [ ] `make release` produces `sssh` binary under 10MB
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
| 7 | Fuzzy Search | Live filtered search | Type any char to filter |
| 8 | First-Run + CLI | Alias suggestion, remaining flags | First run, `--config`, `--no-frequent` |
| 9 | Testing & QA | >80% coverage, clean vet | `make test-cover` |
| 10 | Release | CI pipeline + GitHub Release | Tag v0.1.0 |
