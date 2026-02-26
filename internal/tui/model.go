package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/state"
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeIdentityPicker
)

// sshExitMsg is sent when an SSH session exits.
type sshExitMsg struct {
	err error
}

// Model represents the TUI state for the host list.
type Model struct {
	allHosts         []config.Host
	filtered         []config.Host
	cursor           int
	viewport         int
	viewHeight       int
	width            int
	mode             mode
	searchQuery      string
	state            *state.State
	statePath        string
	availableKeys    []string
	keyPickerCursor  int
	selectedIdentity string
	statusMsg        string
}

// New creates a new Model with hosts sorted by frequency and then alphabetically.
func New(hosts []config.Host, st *state.State, statePath string) Model {
	// Get frequent hosts sorted by connection count (descending)
	frequent := state.FrequentHosts(st, hosts, len(hosts))

	// Build a set of frequent host IDs to exclude from remaining hosts
	frequentSet := make(map[string]bool)
	for _, h := range frequent {
		frequentSet[h.Alias+"\x00"+h.SourceFile] = true
	}

	// Collect remaining hosts (not in frequent set)
	var remaining []config.Host
	for _, h := range hosts {
		if !frequentSet[h.Alias+"\x00"+h.SourceFile] {
			remaining = append(remaining, h)
		}
	}

	// Sort remaining alphabetically by alias (case-insensitive)
	sort.Slice(remaining, func(i, j int) bool {
		return strings.ToLower(remaining[i].Alias) < strings.ToLower(remaining[j].Alias)
	})

	// Combine frequent hosts first, then remaining
	allHosts := make([]config.Host, len(frequent)+len(remaining))
	copy(allHosts, frequent)
	copy(allHosts[len(frequent):], remaining)

	// Initialize filtered list as a copy of all hosts
	filtered := make([]config.Host, len(allHosts))
	copy(filtered, allHosts)

	return Model{
		allHosts:         allHosts,
		filtered:         filtered,
		cursor:           0,
		viewport:         0,
		viewHeight:       20,
		width:            80,
		mode:             modeNormal,
		searchQuery:      "",
		state:            st,
		statePath:        statePath,
		availableKeys:    []string{},
		keyPickerCursor:  0,
		selectedIdentity: "",
		statusMsg:        "",
	}
}

// Init returns nil (no initial command).
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.viewHeight = msg.Height - 3
		if m.viewHeight < 1 {
			m.viewHeight = 1
		}
		return m, nil
	case tea.KeyMsg:
		newModel, cmd := handleKey(m, msg)
		return newModel, cmd
	case sshExitMsg:
		m.selectedIdentity = ""
		m.statusMsg = ""
		return m, nil
	}
	return m, nil
}

// View renders the current TUI display.
func (m Model) View() string {
	header := renderHeader(m)
	list := renderList(m)
	if m.mode == modeIdentityPicker {
		list = renderIdentityPicker(m)
	}
	statusBar := renderStatusBar(m)
	return header + "\n" + list + "\n" + statusBar
}
