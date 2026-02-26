package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/state"
)

// makeHosts builds a Host slice from alias strings.
func makeHosts(aliases ...string) []config.Host {
	hosts := make([]config.Host, len(aliases))
	for i, alias := range aliases {
		hosts[i] = config.Host{
			Alias:      alias,
			Hostname:   alias + ".example.com",
			User:       "user",
			Port:       "22",
			SourceFile: "/home/user/.ssh/config",
			Groups:     []string{},
		}
	}
	return hosts
}

// makeState builds a State with given connection counts.
func makeState(counts map[string]int) *state.State {
	return &state.State{
		Connections: counts,
		FirstRun:    false,
	}
}

// pressKey sends a KeyRunes message and returns the updated model.
func pressKey(m Model, key string) Model {
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updatedModel.(Model)
}

// pressSpecialKey sends a special key message and returns the updated model.
func pressSpecialKey(m Model, keyType tea.KeyType) Model {
	updatedModel, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updatedModel.(Model)
}

// TestCursorWraps tests that the cursor wraps around at the ends of the list.
func TestCursorWraps(t *testing.T) {
	t.Helper()

	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 10

	// Test wrap forward: press ↓ 3 times
	m = pressSpecialKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("After first ↓: expected cursor=1, got %d", m.cursor)
	}
	m = pressSpecialKey(m, tea.KeyDown)
	if m.cursor != 2 {
		t.Errorf("After second ↓: expected cursor=2, got %d", m.cursor)
	}
	m = pressSpecialKey(m, tea.KeyDown)
	if m.cursor != 0 {
		t.Errorf("After third ↓ (wrap): expected cursor=0, got %d", m.cursor)
	}
	if m.viewport != 0 {
		t.Errorf("After wrap forward: expected viewport=0, got %d", m.viewport)
	}

	// Test wrap backward: set cursor to 0, press ↑ once
	m.cursor = 0
	m = pressSpecialKey(m, tea.KeyUp)
	if m.cursor != 2 {
		t.Errorf("After ↑ from cursor=0 (wrap): expected cursor=2, got %d", m.cursor)
	}
}

// TestViewportAdvances tests that the viewport advances when the cursor scrolls past the bottom.
func TestViewportAdvances(t *testing.T) {
	t.Helper()

	hosts := makeHosts("a", "b", "c", "d", "e")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 3

	// Press ↓ 3 times to move cursor from 0 to 3
	m = pressSpecialKey(m, tea.KeyDown)
	m = pressSpecialKey(m, tea.KeyDown)
	m = pressSpecialKey(m, tea.KeyDown)

	if m.cursor != 3 {
		t.Errorf("Expected cursor=3, got %d", m.cursor)
	}
	if m.viewport != 1 {
		t.Errorf("Expected viewport=1, got %d", m.viewport)
	}
}

// TestViewportRetreats tests that the viewport retreats when the cursor scrolls past the top.
func TestViewportRetreats(t *testing.T) {
	t.Helper()

	hosts := makeHosts("a", "b", "c", "d", "e")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 3

	// Manually set cursor=3, viewport=1
	m.cursor = 3
	m.viewport = 1

	// Press ↑ three times to move cursor from 3 to 0
	m = pressSpecialKey(m, tea.KeyUp)
	m = pressSpecialKey(m, tea.KeyUp)
	m = pressSpecialKey(m, tea.KeyUp)

	if m.cursor != 0 {
		t.Errorf("Expected cursor=0, got %d", m.cursor)
	}
	if m.viewport != 0 {
		t.Errorf("Expected viewport=0, got %d", m.viewport)
	}
}

// TestNewSortsFrequentHostsFirst tests that New() sorts hosts with frequent ones first.
func TestNewSortsFrequentHostsFirst(t *testing.T) {
	t.Helper()

	hosts := makeHosts("beta", "alpha", "gamma")
	st := makeState(map[string]int{
		"gamma": 5,
		"beta":  0,
		"alpha": 0,
	})

	m := New(hosts, st, "/tmp/state.json")

	if len(m.allHosts) != 3 {
		t.Errorf("Expected 3 hosts, got %d", len(m.allHosts))
	}

	if m.allHosts[0].Alias != "gamma" {
		t.Errorf("Expected first host=gamma, got %s", m.allHosts[0].Alias)
	}

	if m.allHosts[1].Alias != "alpha" {
		t.Errorf("Expected second host=alpha (sorted), got %s", m.allHosts[1].Alias)
	}

	if m.allHosts[2].Alias != "beta" {
		t.Errorf("Expected third host=beta (sorted), got %s", m.allHosts[2].Alias)
	}
}

// TestIdentityPickerNavigation tests cursor movement in identity picker mode.
func TestIdentityPickerNavigation(t *testing.T) {
	t.Helper()

	hosts := makeHosts("dev", "prod")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	// Manually set up identity picker mode
	m.mode = modeIdentityPicker
	m.availableKeys = []string{"/home/user/.ssh/id_a", "/home/user/.ssh/id_b", "/home/user/.ssh/id_c"}
	m.keyPickerCursor = 0

	// Press j twice
	m = pressKey(m, "j")
	if m.keyPickerCursor != 1 {
		t.Errorf("After first j: expected cursor=1, got %d", m.keyPickerCursor)
	}

	m = pressKey(m, "j")
	if m.keyPickerCursor != 2 {
		t.Errorf("After second j: expected cursor=2, got %d", m.keyPickerCursor)
	}
}

// TestIdentityPickerSelectsIdentity tests selecting an identity.
func TestIdentityPickerSelectsIdentity(t *testing.T) {
	t.Helper()

	hosts := makeHosts("dev")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	// Set up identity picker mode
	m.mode = modeIdentityPicker
	m.availableKeys = []string{"/home/user/.ssh/id_a", "/home/user/.ssh/id_b"}
	m.keyPickerCursor = 1

	// Press enter (KeyEnter type)
	m = pressSpecialKey(m, tea.KeyEnter)

	if m.selectedIdentity != "/home/user/.ssh/id_b" {
		t.Errorf("Expected selectedIdentity=/home/user/.ssh/id_b, got %q", m.selectedIdentity)
	}
	if m.mode != modeNormal {
		t.Errorf("Expected mode=modeNormal, got %d", m.mode)
	}
}

// TestIdentityPickerEscCancels tests escaping the identity picker.
func TestIdentityPickerEscCancels(t *testing.T) {
	t.Helper()

	hosts := makeHosts("dev")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	// Set up identity picker mode with a previously selected identity
	m.mode = modeIdentityPicker
	m.availableKeys = []string{"/home/user/.ssh/id_a"}
	m.selectedIdentity = "prev_identity"

	// Press esc
	m = pressSpecialKey(m, tea.KeyEsc)

	if m.mode != modeNormal {
		t.Errorf("Expected mode=modeNormal, got %d", m.mode)
	}
	if m.selectedIdentity != "prev_identity" {
		t.Errorf("Expected selectedIdentity=prev_identity (unchanged), got %q", m.selectedIdentity)
	}
}

// TestApplySearch_EmptyQuery tests that an empty query returns all hosts.
func TestApplySearch_EmptyQuery(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	m.searchQuery = ""
	applySearch(&m)

	if len(m.filtered) != 3 {
		t.Errorf("Expected 3 hosts, got %d", len(m.filtered))
	}
}

// TestApplySearch_ByAlias tests that fuzzy search filters by alias.
func TestApplySearch_ByAlias(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	m.searchQuery = "beta"
	applySearch(&m)

	if len(m.filtered) == 0 {
		t.Fatal("Expected at least one match for 'beta', got none")
	}
	if m.filtered[0].Alias != "beta" {
		t.Errorf("Expected first match alias='beta', got %q", m.filtered[0].Alias)
	}
}

// TestApplySearch_ByHostname tests that fuzzy search filters by hostname.
func TestApplySearch_ByHostname(t *testing.T) {
	hosts := []config.Host{
		{Alias: "dev", Hostname: "192.168.1.10", User: "alice", Port: "22", Groups: []string{}, SourceFile: "/tmp/config"},
		{Alias: "prod", Hostname: "10.0.0.5", User: "alice", Port: "22", Groups: []string{}, SourceFile: "/tmp/config"},
	}
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	m.searchQuery = "192"
	applySearch(&m)

	if len(m.filtered) == 0 {
		t.Fatal("Expected at least one match for '192', got none")
	}
	if m.filtered[0].Hostname != "192.168.1.10" {
		t.Errorf("Expected match hostname='192.168.1.10', got %q", m.filtered[0].Hostname)
	}
}

// TestApplySearch_ByGroup tests that fuzzy search filters by group tag.
func TestApplySearch_ByGroup(t *testing.T) {
	hosts := []config.Host{
		{Alias: "dev", Hostname: "dev.example.com", User: "alice", Port: "22", Groups: []string{"Work"}, SourceFile: "/tmp/config"},
		{Alias: "home", Hostname: "home.example.com", User: "alice", Port: "22", Groups: []string{"Personal"}, SourceFile: "/tmp/config"},
	}
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	m.searchQuery = "Work"
	applySearch(&m)

	if len(m.filtered) == 0 {
		t.Fatal("Expected at least one match for 'Work', got none")
	}
	if m.filtered[0].Alias != "dev" {
		t.Errorf("Expected match alias='dev', got %q", m.filtered[0].Alias)
	}
}

// TestApplySearch_ResetsCursorAndViewport tests that search resets cursor and viewport to 0.
func TestApplySearch_ResetsCursorAndViewport(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma", "delta", "epsilon")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 3
	m.cursor = 4
	m.viewport = 2

	m.searchQuery = "alpha"
	applySearch(&m)

	if m.cursor != 0 {
		t.Errorf("Expected cursor=0 after search, got %d", m.cursor)
	}
	if m.viewport != 0 {
		t.Errorf("Expected viewport=0 after search, got %d", m.viewport)
	}
}

// TestSearchMode_NavigateDown tests that ↓ moves cursor down while in search mode.
func TestSearchMode_NavigateDown(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 10
	m.mode = modeSearch

	m = pressSpecialKey(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("After ↓ in search mode: expected cursor=1, got %d", m.cursor)
	}
	if m.mode != modeSearch {
		t.Errorf("Mode should remain modeSearch after ↓, got %d", m.mode)
	}
}

// TestSearchMode_NavigateUp tests that ↑ moves cursor up while in search mode.
func TestSearchMode_NavigateUp(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 10
	m.mode = modeSearch
	m.cursor = 2

	m = pressSpecialKey(m, tea.KeyUp)
	if m.cursor != 1 {
		t.Errorf("After ↑ in search mode: expected cursor=1, got %d", m.cursor)
	}
	if m.mode != modeSearch {
		t.Errorf("Mode should remain modeSearch after ↑, got %d", m.mode)
	}
}

// TestSearchMode_BackspaceOnEmptyExitsSearch tests that backspace with an empty query exits search mode.
func TestSearchMode_BackspaceOnEmptyExitsSearch(t *testing.T) {
	hosts := makeHosts("alpha", "beta")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.mode = modeSearch
	m.searchQuery = ""

	m = pressSpecialKey(m, tea.KeyBackspace)
	if m.mode != modeNormal {
		t.Errorf("Expected mode=modeNormal after backspace on empty query, got %d", m.mode)
	}
}

// TestSearchMode_BackspaceOnLastCharExitsSearch tests that backspace on a single-char
// query deletes the char and immediately exits search mode.
func TestSearchMode_BackspaceOnLastCharExitsSearch(t *testing.T) {
	hosts := makeHosts("alpha", "beta")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.mode = modeSearch
	m.searchQuery = "a"

	m = pressSpecialKey(m, tea.KeyBackspace)
	if m.mode != modeNormal {
		t.Errorf("Expected mode=modeNormal after backspace on last char, got %d", m.mode)
	}
	if m.searchQuery != "" {
		t.Errorf("Expected searchQuery='', got %q", m.searchQuery)
	}
}

// TestSearchMode_CtrlWClearsQuery tests that ctrl+w clears the entire query and exits search mode.
func TestSearchMode_CtrlWClearsQuery(t *testing.T) {
	hosts := makeHosts("alpha", "beta")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.mode = modeSearch
	m.searchQuery = "hello"

	m = pressSpecialKey(m, tea.KeyCtrlW)
	if m.mode != modeNormal {
		t.Errorf("Expected mode=modeNormal after ctrl+w, got %d", m.mode)
	}
	if m.searchQuery != "" {
		t.Errorf("Expected searchQuery='', got %q", m.searchQuery)
	}
}

// TestIdentityPickerWrap tests cursor wrapping in identity picker mode.
func TestIdentityPickerWrap(t *testing.T) {
	t.Helper()

	hosts := makeHosts("dev")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	// Set up identity picker mode
	m.mode = modeIdentityPicker
	m.availableKeys = []string{"/home/user/.ssh/id_a", "/home/user/.ssh/id_b", "/home/user/.ssh/id_c"}
	m.keyPickerCursor = 2 // Last item

	// Press j (should wrap to 0)
	m = pressKey(m, "j")

	if m.keyPickerCursor != 0 {
		t.Errorf("Expected cursor=0 (wrap), got %d", m.keyPickerCursor)
	}
}

// TestNormalMode_TypeToSearch tests that pressing a printable character in normal mode
// immediately enters search mode with that character as the initial query.
func TestNormalMode_TypeToSearch(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	m = pressKey(m, "a")

	if m.mode != modeSearch {
		t.Errorf("Expected mode=modeSearch after typing 'a', got %d", m.mode)
	}
	if m.searchQuery != "a" {
		t.Errorf("Expected searchQuery='a', got %q", m.searchQuery)
	}
}

// TestNormalMode_JEntersSearch tests that pressing 'j' in normal mode enters search
// (does NOT navigate down).
func TestNormalMode_JEntersSearch(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 10

	m = pressKey(m, "j")

	if m.mode != modeSearch {
		t.Errorf("Expected mode=modeSearch after typing 'j', got %d", m.mode)
	}
	if m.searchQuery != "j" {
		t.Errorf("Expected searchQuery='j', got %q", m.searchQuery)
	}
	if m.cursor != 0 {
		t.Errorf("Cursor should not have moved; expected cursor=0, got %d", m.cursor)
	}
}

// TestSearchMode_JAppendsToQuery tests that pressing 'j' in search mode appends to
// the query instead of navigating.
func TestSearchMode_JAppendsToQuery(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.mode = modeSearch
	m.searchQuery = "al"

	m = pressKey(m, "j")

	if m.mode != modeSearch {
		t.Errorf("Expected mode=modeSearch after typing 'j', got %d", m.mode)
	}
	if m.searchQuery != "alj" {
		t.Errorf("Expected searchQuery='alj', got %q", m.searchQuery)
	}
}

// TestNormalMode_EscQuits tests that pressing Esc in normal mode returns a quit command.
func TestNormalMode_EscQuits(t *testing.T) {
	hosts := makeHosts("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
}
