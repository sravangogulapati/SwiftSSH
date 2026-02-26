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
		return m, nil // Placeholder for Phase 5
	case modeIdentityPicker:
		return handleIdentityPickerMode(m, msg)
	}
	return m, nil
}

// handleNormalMode processes keys in normal mode.
func handleNormalMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.cursor = (m.cursor + 1) % len(m.filtered)
		// Adjust viewport when wrapping to top or scrolling past bottom
		if m.cursor == 0 {
			m.viewport = 0
		} else if m.cursor >= m.viewport+m.viewHeight {
			m.viewport = m.cursor - m.viewHeight + 1
		}
		return m, nil

	case "k", "up":
		if len(m.filtered) == 0 {
			return m, nil
		}
		m.cursor = (m.cursor - 1 + len(m.filtered)) % len(m.filtered)
		// Adjust viewport when wrapping to bottom or scrolling past top
		if m.cursor == len(m.filtered)-1 {
			m.viewport = max(0, len(m.filtered)-m.viewHeight)
		} else if m.cursor < m.viewport {
			m.viewport = m.cursor
		}
		return m, nil

	case "enter":
		// SSH connection
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
		cmd := ssh.ConnectCmd(host, m.selectedIdentity)
		return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
			return sshExitMsg{err: err}
		})

	case "/":
		// Placeholder for Phase 5 (search mode)
		return m, nil

	case "i":
		// Identity picker
		keys, _ := ssh.ScanPublicKeys(platform.SSHKeyDir())
		if len(keys) == 0 {
			m.statusMsg = "No SSH keys found in ~/.ssh"
			return m, nil
		}
		m.availableKeys = keys
		m.keyPickerCursor = 0
		m.mode = modeIdentityPicker
		return m, nil

	case "p":
		// Placeholder for Phase 6 (ping toggle)
		return m, nil
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

// handleIdentityPickerMode processes keys in identity picker mode.
func handleIdentityPickerMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if len(m.availableKeys) > 0 {
			m.keyPickerCursor = (m.keyPickerCursor + 1) % len(m.availableKeys)
		}
		return m, nil

	case "k", "up":
		if len(m.availableKeys) > 0 {
			m.keyPickerCursor = (m.keyPickerCursor - 1 + len(m.availableKeys)) % len(m.availableKeys)
		}
		return m, nil

	case "enter":
		if len(m.availableKeys) > 0 {
			m.selectedIdentity = m.availableKeys[m.keyPickerCursor]
		}
		m.mode = modeNormal
		return m, nil

	case "esc":
		m.mode = modeNormal
		return m, nil

	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}
