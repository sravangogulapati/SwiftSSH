This refined PRD focuses on **Minimum Viable Product (MVP)**. It strips away complexity to ensure the app is fast, safe, and easy to build while delivering the "quick-launch" value.

---

# PRD: SwiftSSH (MVP Specification)

## 1. Core Vision
A single-binary Go tool that turns your SSH config into a searchable, interactive menu. It prioritizes the **"Main Config"** as the source of truth and uses a simple TUI for navigation.

## 2. Functional Specifications

### 2.1. Configuration & Parsing
*   **Parsing:** Support `~/.ssh/config` and recursive `Include` directives.
*   **Duplicate Hosts:** Treated as separate entries. If a user defined `Host dev` twice, both appear in the list to avoid complex merging logic.
*   **Writing (Auto-Entry):** 
    *   **Logic:** Use **Append-only**. When a user connects to a new IP, SwiftSSH appends a new block to the end of the *main* `~/.ssh/config` file.
    *   **Default User:** Uses the current system user (`$USER`).
*   **Safety:** Before any append operation, the app creates a single `config.bak` (overwriting the previous backup).
*   **Magic Comments:** Supports a single line `# @group Work, Personal` per host.

### 2.2. Connection Logic
*   **Suspension:** Use `tea.ExecProcess` (Bubble Tea standard). This ensures the app cleans up the TUI, hands over the terminal to the SSH process, and returns to the menu upon exit.
*   **Identity Cycling:** When hitting `i`, the app scans `~/.ssh/*.pub` files to find available private keys and allows the user to select one for the **current session only** (does not write to config).

### 2.3. Persistence (State)
*   **State File:** A simple JSON file at `~/.config/swiftssh/state.json`.
*   **Logic:** Stores a simple **frequency count** of successful connections to power a "Frequent" section at the top of the TUI.

---

## 3. User Interface (TUI)

### 3.1. Layout & Interaction
*   **Unified List:** A single list of all hosts. Groups are treated as metadata tags that appear next to the Host name.
*   **Fuzzy Search:** Matches strictly against **Host (Alias)**, **Hostname (IP/Domain)**, and **Groups**.
*   **Vim Keys:** Supports `j`/`k` for navigation and `Enter` to connect.
*   **Paging:** Uses infinite scrolling (Viewport) to handle configs with 1,000+ entries effortlessly.
*   **Colors:** Inherits the terminal's **ANSI palette**. This ensures the app looks native to the user's existing theme (Catppuccin, Gruvbox, etc.).

### 3.2. Health Checks
*   **Mechanism:** TCP Dial to **Port 22** (not ICMP ping). This avoids root permission issues on Linux.
*   **Scope:** Only visible hosts in the current viewport are checked to keep the app responsive.
*   **Default:** Off (Toggled on via `p`).

---

## 4. Technical & Platform Specs

### 4.1. Implementation Details
*   **Language:** Go (Standard Library + Bubble Tea).
*   **CLI Args:** Minimalist `os.Args` or `flag` package (No Cobra/Viper to keep the binary small).
*   **Windows Support:** Target **Windows Terminal** (Modern). Legacy `cmd.exe` support is explicitly out of scope for the MVP.
*   **Alias Setup:** On first run, the app prints a suggested command for the user to copy-paste into their `.zshrc`, `.bashrc`, or PowerShell profile (e.g., `alias s='swiftssh'`). It does **not** modify these files automatically.

### 4.2. File Locations (MVP)
*   **Unix:**
    *   Config: `~/.ssh/config`
    *   State: `~/.config/swiftssh/state.json`
*   **Windows:**
    *   Config: `%USERPROFILE%\.ssh\config`
    *   State: `%LOCALAPPDATA%\swiftssh\state.json`

---

## 5. Security Summary
1.  **Argument Safety:** Passes host/user as separate elements in `exec.Command` to prevent shell injection.
2.  **Read-Only Focus:** The app treats the config as read-only, except for explicit "Auto-Entry" append operations.
3.  **No Cloud:** Zero telemetry or cloud syncing.

---

## 6. MVP Out of Scope (For Future Versions)
*   Native `ssh-copy-id` implementation.
*   Permanent writing of toggled identities to config.
*   Complex "Frecency" algorithms.
*   Multi-pane/Sidebar UI layouts.
*   Automatic shell profile modification.