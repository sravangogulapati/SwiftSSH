# SwiftSSH — Phase 6 Feedback & Iteration

## MVP State (Phase 5 complete, as of 2026-02-25)

### What Works
| Feature | Status | Notes |
|---------|--------|-------|
| `~/.ssh/config` parsing | ✅ Complete | Include, glob, circular-detect, magic comments |
| Vim navigation (j/k) | ✅ Complete | Wrap-around, viewport auto-scroll |
| Frequent hosts (top-N first) | ✅ Complete | Sorted by connection count desc |
| SSH connection via Enter | ✅ Complete | Uses `tea.ExecProcess`, state recorded |
| Identity picker (i key) | ✅ Complete | Lists `~/.ssh/*.pub` keys, overlay UI |
| Auto-entry for unknown hosts | ✅ Complete | Appends to `~/.ssh/config` with backup |
| State persistence | ✅ Complete | JSON at `~/.config/swiftssh/state.json` |
| Cross-platform paths | ✅ Complete | Unix + Windows Terminal |
| Fuzzy search (/ key) | 🚧 Placeholder | Mode switches but no filtering |
| Ping toggle (p key) | 🚧 Placeholder | No-op; health module empty |
| First-run UX | 🚧 Phase 9 | Not yet started |
| CLI flags (--config, --no-frequent) | 🚧 Phase 9 | Not yet started |

---

## Identified Issues (Pre-User Feedback)

### Critical
*(none blocking core usage)*

### High Priority
| # | Issue | Location | Impact |
|---|-------|----------|--------|
| H1 | State save error silenced | `keybindings.go:66` | Connection frequency may not persist silently |
| H2 | Config append error silenced | `keybindings.go:70` | Unknown host not added silently |
| H3 | Config writer missing trailing newline | `writer.go:41-50` | Second append corrupts SSH config formatting |

### Medium Priority
| # | Issue | Location | Impact |
|---|-------|----------|--------|
| M1 | Key scan error silenced | `keybindings.go:85` | Empty key list with no explanation |
| M2 | `IsKnownHost` checks `host.Hostname` which may be empty | `keybindings.go:68` | Appends duplicate host to config |
| M3 | State file written with `0644` permissions | `state.go` | Minor privacy: world-readable connection history |

### Low Priority
*(none identified pre-testing)*

---

## User Feedback (Round 1)

| # | Category | Feedback |
|---|----------|---------|
| F1 | UX | Host list is hard to read — all columns left-aligned with no spacing |
| F2 | UX | `i` (identity picker) missing from the status bar hints |
| F3 | Feature | No way to add hosts from the CLI; want `swiftssh user@host [-p port]` to auto-save + connect |

---

## Triage & Action Plan

| # | Priority | Resolution |
|---|----------|-----------|
| F1 | High | Quick Fix — computed column widths + header row |
| F2 | High | Quick Fix — add `i: identity` to status bar hint string |
| F3 | High | Quick Fix — SSH passthrough in `main.go` |

---

## Implementation Log

### Quick Fixes Applied (2026-02-25)

**F1 — Column-aligned table layout (`internal/tui/views.go`)**
- Added `padRight`, `truncateStr`, `colWidths` helpers
- `renderList` now computes per-column widths from actual host data, renders a dim header row (`ALIAS / HOSTNAME / USER / GROUPS`)
- `renderRow` signature changed to accept `(aliasW, hostW, userW int)` for consistent column padding
- Selected row: plain-text padding then `selectedStyle.Render` for clean reverse-video
- Non-selected row: individual cell dim/tag styling applied after padding
- Adjusted `viewHeight = Height - 4` in `model.go` to account for the header row

**F2 — Status bar hint (`internal/tui/views.go`)**
- Added `| i: identity` to `renderStatusBar` hint string

**F3 — CLI SSH passthrough (`cmd/swiftssh/main.go`)**
- `parseSSHTarget(args)` — minimal SSH arg parser: extracts destination, port (`-p`), user (`-l`), identity (`-i`); skips all other option-value pairs
- `runPassthrough(args)` — parses `user@hostname`, auto-appends to `~/.ssh/config` if not known (with backup), then hands off to `exec.Command("ssh", args...)`
- Usage: `swiftssh user@host`, `swiftssh host -p 2222`, etc.
- Prints `swiftssh: saved '<alias>' to SSH config` to stderr when a new host is added
