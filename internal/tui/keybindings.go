package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/platform"
	"github.com/srava/swiftssh/internal/ssh"
	"github.com/srava/swiftssh/internal/state"
)

// handleKey processes key events and updates the model accordingly.
func handleKey(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch m.mode {
	case modeNormal:
		return handleNormalMode(m, msg)
	case modeSearch:
		return handleSearchMode(m, msg)
	case modeEdit:
		return handleEditMode(m, msg)
	}
	return m, nil
}

// moveCursorDown moves the cursor down by one, wrapping around to the top.
func moveCursorDown(m Model) Model {
	if len(m.filtered) == 0 {
		return m
	}
	m.cursor = (m.cursor + 1) % len(m.filtered)
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

	state.RecordConnection(m.state, host.Alias)
	_ = state.Save(m.statePath, m.state)

	if !config.IsKnownHost(m.allHosts, host.Hostname) {
		_ = config.AppendHost(platform.SSHConfigPath(), platform.SSHConfigBackupPath(), host)
	}

	cmd := ssh.ConnectCmd(host, "")
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return nil
	})
}

// openEditForm initialises an editForm for the currently selected host.
func openEditForm(m Model) Model {
	if len(m.filtered) == 0 {
		m.statusMsg = "No host selected."
		return m
	}
	host := m.filtered[m.cursor]
	if host.LineStart == 0 {
		m.statusMsg = "Cannot edit: host has no tracked line position."
		return m
	}

	form := &editForm{
		original:    host,
		activeField: fieldAlias,
	}
	form.fields[fieldAlias] = host.Alias
	form.fields[fieldHostname] = host.Hostname
	form.fields[fieldUser] = host.User
	form.fields[fieldPort] = host.Port
	form.fields[fieldIdentityFile] = host.IdentityFile
	form.fields[fieldGroups] = strings.Join(host.Groups, ", ")

	m.edit = form
	m.mode = modeEdit
	return m
}

// saveEditForm validates and saves the edit form, returning a cmd that emits editSavedMsg.
func saveEditForm(m Model) (Model, tea.Cmd) {
	form := m.edit

	alias := strings.TrimSpace(form.fields[fieldAlias])
	hostname := strings.TrimSpace(form.fields[fieldHostname])

	if alias == "" {
		form.statusMsg = "Alias cannot be empty."
		m.edit = form
		return m, nil
	}
	if hostname == "" {
		form.statusMsg = "Hostname cannot be empty."
		m.edit = form
		return m, nil
	}

	// Parse groups from comma-separated string
	var groups []string
	for _, g := range strings.Split(form.fields[fieldGroups], ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			groups = append(groups, g)
		}
	}

	port := strings.TrimSpace(form.fields[fieldPort])
	if port == "" {
		port = "22"
	}

	updated := form.original
	updated.Alias = alias
	updated.Hostname = hostname
	updated.User = strings.TrimSpace(form.fields[fieldUser])
	updated.Port = port
	updated.IdentityFile = strings.TrimSpace(form.fields[fieldIdentityFile])
	updated.Groups = groups

	// Find index in allHosts by SourceFile + LineStart
	idx := -1
	for i, h := range m.allHosts {
		if h.SourceFile == form.original.SourceFile && h.LineStart == form.original.LineStart {
			idx = i
			break
		}
	}

	originalLineStart := form.original.LineStart
	newLineStart, lineDelta, err := config.ReplaceHostBlock(updated)
	if err != nil {
		form.statusMsg = "Save failed: " + err.Error()
		m.edit = form
		return m, nil
	}
	updated.LineStart = newLineStart

	savedIdx := idx
	savedHost := updated
	return m, func() tea.Msg {
		return editSavedMsg{
			updated:           savedHost,
			index:             savedIdx,
			lineDelta:         lineDelta,
			originalLineStart: originalLineStart,
			sourceFile:        savedHost.SourceFile,
		}
	}
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

	case "ctrl+e":
		return openEditForm(m), nil
	}

	if msg.Type == tea.KeyRunes {
		m.mode = modeSearch
		m.searchQuery = string(msg.Runes)
		applySearch(&m)
	}
	return m, nil
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

	case "ctrl+e":
		return openEditForm(m), nil

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

// handleEditMode processes keys while the editor form is open.
func handleEditMode(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	form := m.edit

	switch msg.String() {
	case "esc":
		m.edit = nil
		m.mode = modeNormal
		return m, nil

	case "ctrl+c":
		return m, tea.Quit

	case "down":
		form.activeField = (form.activeField + 1) % fieldCount
		m.edit = form
		return m, nil

	case "up":
		form.activeField = (form.activeField - 1 + fieldCount) % fieldCount
		m.edit = form
		return m, nil

	case "backspace":
		runes := []rune(form.fields[form.activeField])
		if len(runes) > 0 {
			form.fields[form.activeField] = string(runes[:len(runes)-1])
		}
		form.statusMsg = ""
		m.edit = form
		return m, nil

	case "ctrl+u":
		form.fields[form.activeField] = ""
		form.statusMsg = ""
		m.edit = form
		return m, nil

	case "enter":
		return saveEditForm(m)

	default:
		if msg.Type == tea.KeyRunes {
			form.fields[form.activeField] += string(msg.Runes)
			form.statusMsg = ""
			m.edit = form
		}
		return m, nil
	}
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
