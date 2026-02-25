package tui

import tea "github.com/charmbracelet/bubbletea"

// handleKey processes key events and updates the model accordingly.
func handleKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	// Handle mode-specific keys
	switch m.mode {
	case modeNormal:
		return handleNormalMode(m, msg)
	case modeSearch:
		return m, nil // Placeholder for Phase 5
	case modeIdentityPicker:
		return m, nil // Placeholder for Phase 6
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
		// Placeholder for Phase 5 (SSH connection)
		return m, nil

	case "/":
		// Placeholder for Phase 5 (search mode)
		return m, nil

	case "i":
		// Placeholder for Phase 6 (identity picker)
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
