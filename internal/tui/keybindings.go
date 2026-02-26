package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/platform"
	"github.com/srava/swiftssh/internal/ssh"
	"github.com/srava/swiftssh/internal/state"
)

// handleKey processes key events and updates the model accordingly.
func handleKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	// Handle mode-specific keys
	switch m.mode {
	case modeNormal:
		return handleNormalMode(m, msg)
	case modeSearch:
		return handleSearchMode(m, msg)
	}
	return m, nil
}

// moveCursorDown moves the cursor down by one, wrapping around to the top.
func moveCursorDown(m Model) Model {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor = (m.cursor + 1) % len(m.filtered)
	// Adjust viewport when wrapping to top or scrolling past bottom
	if m.cursor == 0 {
		m.viewport = 0
	} else if m.cursor >= m.viewport+m.viewHeight {
		m.viewport = m.cursor - m.viewHeight + 1
	}
	return m
}

// moveCursorUp moves the cursor up by one, wrapping around to the bottom.
func moveCursorUp(m Model) Model {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor = (m.cursor - 1 + len(m.filtered)) % len(m.filtered)
	// Adjust viewport when wrapping to bottom or scrolling past top
	if m.cursor == len(m.filtered)-1 {
		m.viewport = max(0, len(m.filtered)-m.viewHeight)
	} else if m.cursor < m.viewport {
		m.viewport = m.cursor
	}
	return m
}

// connectToSelected records the connection and executes SSH for the selected host.
func connectToSelected(m Model) (Model, tea.Cmd) {
	if len(m.filtered) == 0 {
		return m, nil
	}
	host := m.filtered[m.cursor]

	// Record connection in state
	state.RecordConnection(m.state, host.Alias)
	_ = state.Save(m.statePath, m.state)

	// Check if host is known; if not, append it to config
	if !config.IsKnownHost(m.allHosts, host.Hostname) {
		_ = config.AppendHost(platform.SSHConfigPath(), platform.SSHConfigBackupPath(), host)
	}

	// Execute SSH connection
	cmd := ssh.ConnectCmd(host, "")
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return nil
	})
}

// handleNormalMode processes keys in normal mode.
func handleNormalMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		return m, tea.Quit

	case "down":
		return moveCursorDown(m), nil

	case "up":
		return moveCursorUp(m), nil

	case "enter":
		return connectToSelected(m)

	}

	// Any printable rune enters search mode immediately with that character.
	if msg.Type == tea.KeyRunes {
		m.mode = modeSearch
		m.searchQuery = string(msg.Runes)
		applySearch(&m)
	}
	return m, nil
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleSearchMode processes keys in search mode.
func handleSearchMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchQuery = ""
		applySearch(&m)
		m.mode = modeNormal
		return m, nil

	case "enter":
		return connectToSelected(m)

	case "down":
		return moveCursorDown(m), nil

	case "up":
		return moveCursorUp(m), nil

	case "ctrl+c":
		return m, tea.Quit

	case "ctrl+w":
		m.searchQuery = ""
		applySearch(&m)
		m.mode = modeNormal
		return m, nil

	case "backspace":
		runes := []rune(m.searchQuery)
		if len(runes) == 0 {
			m.mode = modeNormal
			return m, nil
		}
		m.searchQuery = string(runes[:len(runes)-1])
		applySearch(&m)
		if len(m.searchQuery) == 0 {
			m.mode = modeNormal
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.searchQuery += string(msg.Runes)
			applySearch(&m)
		}
		return m, nil
	}
}

