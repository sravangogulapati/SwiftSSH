package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/srava/swiftssh/internal/config"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
	tagStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle   = lipgloss.NewStyle().Faint(true)
)

// padRight pads s with spaces on the right to exactly width characters.
// If s is already width or longer, it is returned as-is.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// truncateStr truncates s to at most maxW bytes, appending "…" if truncated.
func truncateStr(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if len(s) <= maxW {
		return s
	}
	if maxW == 1 {
		return s[:1]
	}
	return s[:maxW-1] + "~" // use ~ to stay single-byte safe
}

// colWidths computes per-column widths from the host list, floored at the
// header label widths and capped at reasonable maximums.
func colWidths(hosts []config.Host) (aliasW, hostW, userW int) {
	aliasW = len("ALIAS")
	hostW = len("HOSTNAME")
	userW = len("USER")
	for _, h := range hosts {
		if n := len(h.Alias); n > aliasW {
			aliasW = n
		}
		if n := len(h.Hostname); n > hostW {
			hostW = n
		}
		if n := len(h.User); n > userW {
			userW = n
		}
	}
	const maxAlias, maxHost, maxUser = 30, 40, 20
	if aliasW > maxAlias {
		aliasW = maxAlias
	}
	if hostW > maxHost {
		hostW = maxHost
	}
	if userW > maxUser {
		userW = maxUser
	}
	return
}

// renderHeader returns the header line for the TUI.
func renderHeader(m Model) string {
	header := titleStyle.Render("SwiftSSH")
	switch m.mode {
	case modeSearch:
		header += "  " + m.searchQuery + "█"
	case modeNormal:
		header += "  " + dimStyle.Render("Type to search")
	}
	return header
}

// renderList returns the column-aligned, scrollable list of hosts.
func renderList(m Model) string {
	if len(m.filtered) == 0 {
		return dimStyle.Render("  No hosts found.")
	}

	aliasW, hostW, userW := colWidths(m.filtered)

	// Column header row (always visible, above the scrolling viewport)
	headerStr := "  " +
		padRight("ALIAS", aliasW) + "  " +
		padRight("HOSTNAME", hostW) + "  " +
		padRight("USER", userW) + "  " +
		"GROUPS"
	rows := []string{dimStyle.Render(headerStr)}

	end := min(m.viewport+m.viewHeight, len(m.filtered))
	for i := m.viewport; i < end; i++ {
		rows = append(rows, renderRow(m, i, aliasW, hostW, userW))
	}

	return strings.Join(rows, "\n")
}

// renderRow returns the rendered display for a single host at index i.
// Column widths must be passed in so all rows share the same alignment.
func renderRow(m Model, i, aliasW, hostW, userW int) string {
	h := m.filtered[i]
	isSelected := i == m.cursor

	alias := padRight(truncateStr(h.Alias, aliasW), aliasW)
	hostname := padRight(truncateStr(h.Hostname, hostW), hostW)
	user := h.User
	if user == "" {
		user = "-"
	}
	userStr := padRight(truncateStr(user, userW), userW)

	var groupParts []string
	for _, g := range h.Groups {
		groupParts = append(groupParts, "["+g+"]")
	}
	groups := strings.Join(groupParts, " ")

	prefix := "  "
	if isSelected {
		prefix = "> "
	}

	if isSelected {
		// Render plain text so selectedStyle (reverse video) works cleanly
		row := prefix + alias + "  " + hostname + "  " + userStr
		if groups != "" {
			row += "  " + groups
		}
		return selectedStyle.Render(row)
	}

	// Non-selected: dim secondary columns, color group tags
	row := prefix + alias + "  " + dimStyle.Render(hostname) + "  " + dimStyle.Render(userStr)
	if groups != "" {
		row += "  " + tagStyle.Render(groups)
	}
	return row
}

// renderStatusBar returns the status bar display.
func renderStatusBar(m Model) string {
	return statusStyle.Render(fmt.Sprintf(
		"%d hosts | Enter: connect | esc: quit",
		len(m.filtered),
	))
}
