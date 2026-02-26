package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/srava/swiftssh/internal/ssh"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	tagStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle   = lipgloss.NewStyle().Faint(true)
)

// renderHeader returns the header line for the TUI.
func renderHeader(m Model) string {
	header := titleStyle.Render("SwiftSSH")
	if m.mode == modeSearch {
		header += "  / " + m.searchQuery + "█"
	}
	return header
}

// renderList returns the list of hosts with proper viewport.
func renderList(m Model) string {
	if len(m.filtered) == 0 {
		return dimStyle.Render("  No hosts found.")
	}

	end := min(m.viewport+m.viewHeight, len(m.filtered))
	var rows []string

	for i := m.viewport; i < end; i++ {
		rows = append(rows, renderRow(m, i))
	}

	return strings.Join(rows, "\n")
}

// renderRow returns the rendered display for a single host at index i.
func renderRow(m Model, i int) string {
	h := m.filtered[i]

	// Build the row text
	var parts []string
	parts = append(parts, h.Alias)

	if h.Hostname != "" {
		parts = append(parts, dimStyle.Render(h.Hostname))
	}

	if h.User != "" {
		parts = append(parts, dimStyle.Render(h.User+"@"))
	}

	// Add groups as tags
	for _, group := range h.Groups {
		parts = append(parts, tagStyle.Render("["+group+"]"))
	}

	rowText := strings.Join(parts, "  ")

	// Apply selection styling if this is the cursor position
	if i == m.cursor {
		return selectedStyle.Render("> " + rowText)
	}

	return "  " + rowText
}

// renderStatusBar returns the status bar display.
func renderStatusBar(m Model) string {
	if m.statusMsg != "" {
		return statusStyle.Render(m.statusMsg)
	}
	return statusStyle.Render(fmt.Sprintf("%d hosts | q: quit | /: search | Enter: connect", len(m.filtered)))
}

// renderIdentityPicker renders the identity selection overlay.
func renderIdentityPicker(m Model) string {
	var lines []string
	lines = append(lines, dimStyle.Render("Select identity (Esc to cancel):"))

	for i, keyPath := range m.availableKeys {
		label := ssh.KeyLabel(keyPath)
		if i == m.keyPickerCursor {
			lines = append(lines, selectedStyle.Render("> "+label))
		} else {
			lines = append(lines, "  "+label)
		}
	}

	return strings.Join(lines, "\n")
}
