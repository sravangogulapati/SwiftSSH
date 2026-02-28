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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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

	m := New(hosts, st, "/tmp/state.json", false)

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

// TestApplySearch_EmptyQuery tests that an empty query returns all hosts.
func TestApplySearch_EmptyQuery(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

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
	m := New(hosts, st, "/tmp/state.json", false)

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
	m := New(hosts, st, "/tmp/state.json", false)

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
	m := New(hosts, st, "/tmp/state.json", false)

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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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

// TestNormalMode_TypeToSearch tests that pressing a printable character in normal mode
// immediately enters search mode with that character as the initial query.
func TestNormalMode_TypeToSearch(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

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
	m := New(hosts, st, "/tmp/state.json", false)
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
	m := New(hosts, st, "/tmp/state.json", false)
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

// makeHostsWithLine builds hosts that have LineStart set (needed for edit mode).
func makeHostsWithLine(aliases ...string) []config.Host {
	hosts := make([]config.Host, len(aliases))
	for i, alias := range aliases {
		hosts[i] = config.Host{
			Alias:      alias,
			Hostname:   alias + ".example.com",
			User:       "user",
			Port:       "22",
			SourceFile: "/home/user/.ssh/config",
			Groups:     []string{},
			LineStart:  (i * 3) + 1, // non-zero, distinct per host
		}
	}
	return hosts
}

// pressCtrlE sends a ctrl+e key message and returns the updated model.
func pressCtrlE(m Model) Model {
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	return updatedModel.(Model)
}

// pressCtrlU sends a ctrl+u key message.
func pressCtrlU(m Model) Model {
	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	return updatedModel.(Model)
}

// TestEditMode_OpenClose tests that Ctrl+E opens edit mode and Esc closes it.
func TestEditMode_OpenClose(t *testing.T) {
	hosts := makeHostsWithLine("alpha", "beta")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)
	m.viewHeight = 10

	m = pressCtrlE(m)
	if m.mode != modeEdit {
		t.Errorf("expected modeEdit after Ctrl+E, got %d", m.mode)
	}
	if m.edit == nil {
		t.Fatal("expected edit form to be non-nil")
	}

	// Esc should cancel and return to normal mode
	m = pressSpecialKey(m, tea.KeyEsc)
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after Esc, got %d", m.mode)
	}
	if m.edit != nil {
		t.Error("expected edit form to be nil after Esc")
	}
}

// TestEditMode_NoOpenOnZeroLineStart tests that Ctrl+E on a host with LineStart=0
// does not enter edit mode but sets a statusMsg instead.
func TestEditMode_NoOpenOnZeroLineStart(t *testing.T) {
	hosts := []config.Host{
		{Alias: "notrack", Hostname: "notrack.example.com", Port: "22",
			SourceFile: "/tmp/config", Groups: []string{}, LineStart: 0},
	}
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	m = pressCtrlE(m)
	if m.mode == modeEdit {
		t.Error("should not enter edit mode for LineStart=0 host")
	}
	if m.statusMsg == "" {
		t.Error("expected a statusMsg explaining the failure")
	}
}

// TestEditMode_PrePopulatesFields tests that all 6 fields are pre-populated from the host.
func TestEditMode_PrePopulatesFields(t *testing.T) {
	hosts := []config.Host{
		{
			Alias:        "myhost",
			Hostname:     "my.example.com",
			User:         "alice",
			Port:         "2222",
			IdentityFile: "/home/alice/.ssh/id_rsa",
			Groups:       []string{"Work", "Personal"},
			SourceFile:   "/tmp/config",
			LineStart:    1,
		},
	}
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	m = pressCtrlE(m)
	if m.edit == nil {
		t.Fatal("expected edit form")
	}

	f := m.edit
	if f.fields[fieldAlias] != "myhost" {
		t.Errorf("expected Alias='myhost', got %q", f.fields[fieldAlias])
	}
	if f.fields[fieldHostname] != "my.example.com" {
		t.Errorf("expected Hostname='my.example.com', got %q", f.fields[fieldHostname])
	}
	if f.fields[fieldUser] != "alice" {
		t.Errorf("expected User='alice', got %q", f.fields[fieldUser])
	}
	if f.fields[fieldPort] != "2222" {
		t.Errorf("expected Port='2222', got %q", f.fields[fieldPort])
	}
	if f.fields[fieldIdentityFile] != "/home/alice/.ssh/id_rsa" {
		t.Errorf("expected IdentityFile pre-populated, got %q", f.fields[fieldIdentityFile])
	}
	if f.fields[fieldGroups] != "Work, Personal" {
		t.Errorf("expected Groups='Work, Personal', got %q", f.fields[fieldGroups])
	}
}

// TestEditMode_FieldNavigation tests that ↓/↑ cycles through fields.
func TestEditMode_FieldNavigation(t *testing.T) {
	hosts := makeHostsWithLine("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)
	m = pressCtrlE(m)

	if m.edit.activeField != fieldAlias {
		t.Errorf("expected initial activeField=fieldAlias, got %d", m.edit.activeField)
	}

	// ↓ cycles forward
	m = pressSpecialKey(m, tea.KeyDown)
	if m.edit.activeField != fieldHostname {
		t.Errorf("after ↓: expected fieldHostname, got %d", m.edit.activeField)
	}

	// ↑ cycles backward
	m = pressSpecialKey(m, tea.KeyUp)
	if m.edit.activeField != fieldAlias {
		t.Errorf("after ↑: expected fieldAlias, got %d", m.edit.activeField)
	}

	// ↓ wraps around past last field
	for i := 0; i < int(fieldCount); i++ {
		m = pressSpecialKey(m, tea.KeyDown)
	}
	if m.edit.activeField != fieldAlias {
		t.Errorf("after full ↓ cycle: expected fieldAlias, got %d", m.edit.activeField)
	}
}

// TestEditMode_TextInput tests printable appending, backspace delete, and Ctrl+U clear.
func TestEditMode_TextInput(t *testing.T) {
	hosts := makeHostsWithLine("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)
	m = pressCtrlE(m)

	// Clear the alias field first
	m = pressCtrlU(m)
	if m.edit.fields[fieldAlias] != "" {
		t.Errorf("after Ctrl+U: expected empty alias, got %q", m.edit.fields[fieldAlias])
	}

	// Append printable characters
	m = pressKey(m, "n")
	m = pressKey(m, "e")
	m = pressKey(m, "w")
	if m.edit.fields[fieldAlias] != "new" {
		t.Errorf("after typing 'new': expected alias='new', got %q", m.edit.fields[fieldAlias])
	}

	// Backspace deletes last rune
	m = pressSpecialKey(m, tea.KeyBackspace)
	if m.edit.fields[fieldAlias] != "ne" {
		t.Errorf("after backspace: expected alias='ne', got %q", m.edit.fields[fieldAlias])
	}

	// Ctrl+U clears the field
	m = pressCtrlU(m)
	if m.edit.fields[fieldAlias] != "" {
		t.Errorf("after Ctrl+U: expected empty alias, got %q", m.edit.fields[fieldAlias])
	}
}

// TestEditMode_ValidationEmptyAlias tests that saving with an empty alias shows an error.
func TestEditMode_ValidationEmptyAlias(t *testing.T) {
	hosts := makeHostsWithLine("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)
	m = pressCtrlE(m)

	// Clear alias field
	m = pressCtrlU(m)

	// Press Enter to save
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.mode != modeEdit {
		t.Errorf("expected to remain in modeEdit on validation failure, got %d", m.mode)
	}
	if m.edit.statusMsg == "" {
		t.Error("expected validation error message for empty alias")
	}
}

// TestEditMode_ValidationEmptyHostname tests that saving with an empty hostname shows an error.
func TestEditMode_ValidationEmptyHostname(t *testing.T) {
	hosts := makeHostsWithLine("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)
	m = pressCtrlE(m)

	// Navigate to hostname field and clear it
	m = pressSpecialKey(m, tea.KeyDown) // move to Hostname
	m = pressCtrlU(m)

	// Press Enter to save
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.mode != modeEdit {
		t.Errorf("expected to remain in modeEdit on validation failure, got %d", m.mode)
	}
	if m.edit.statusMsg == "" {
		t.Error("expected validation error message for empty hostname")
	}
}

// TestEditMode_SaveUpdatesAllHosts tests that receiving editSavedMsg updates allHosts.
func TestEditMode_SaveUpdatesAllHosts(t *testing.T) {
	hosts := makeHostsWithLine("alpha", "beta")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	updated := config.Host{
		Alias:      "alpha-updated",
		Hostname:   "updated.example.com",
		Port:       "22",
		SourceFile: "/home/user/.ssh/config",
		LineStart:  1,
		Groups:     []string{},
	}

	// Directly inject the editSavedMsg
	newModel, _ := m.Update(editSavedMsg{updated: updated, index: 0})
	m = newModel.(Model)

	if m.allHosts[0].Alias != "alpha-updated" {
		t.Errorf("expected allHosts[0].Alias='alpha-updated', got %q", m.allHosts[0].Alias)
	}
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after save, got %d", m.mode)
	}
	if m.statusMsg != "Saved." {
		t.Errorf("expected statusMsg='Saved.', got %q", m.statusMsg)
	}
	if m.edit != nil {
		t.Error("expected edit form to be cleared after save")
	}
}

// TestEditMode_SaveUpdatesLineStart tests that editSavedMsg propagates the new LineStart
// from the saved host into allHosts, so subsequent edits use the correct position.
func TestEditMode_SaveUpdatesLineStart(t *testing.T) {
	hosts := makeHostsWithLine("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	updated := config.Host{
		Alias:      "alpha",
		Hostname:   "alpha.example.com",
		Port:       "22",
		SourceFile: "/home/user/.ssh/config",
		LineStart:  99, // new LineStart returned by ReplaceHostBlock after save
		Groups:     []string{},
	}

	newModel, _ := m.Update(editSavedMsg{
		updated:           updated,
		index:             0,
		lineDelta:         0,
		originalLineStart: 1,
		sourceFile:        "/home/user/.ssh/config",
	})
	m = newModel.(Model)

	if m.allHosts[0].LineStart != 99 {
		t.Errorf("expected allHosts[0].LineStart=99, got %d", m.allHosts[0].LineStart)
	}
}

// TestEditMode_LineDeltaUpdatesSubsequentHosts tests that when a saved block grows by N
// lines, all hosts in the same file with LineStart > originalLineStart are shifted by N,
// while hosts in other files are unaffected.
func TestEditMode_LineDeltaUpdatesSubsequentHosts(t *testing.T) {
	// Two hosts in the same file; one host in a different file.
	hosts := []config.Host{
		{
			Alias:      "alpha",
			Hostname:   "alpha.example.com",
			Port:       "22",
			SourceFile: "/home/user/.ssh/config",
			LineStart:  1,
			Groups:     []string{},
		},
		{
			Alias:      "beta",
			Hostname:   "beta.example.com",
			Port:       "22",
			SourceFile: "/home/user/.ssh/config",
			LineStart:  5, // after alpha's block
			Groups:     []string{},
		},
		{
			Alias:      "gamma",
			Hostname:   "gamma.example.com",
			Port:       "22",
			SourceFile: "/home/user/.ssh/config2", // different file — must NOT be shifted
			LineStart:  1,
			Groups:     []string{},
		},
	}
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	// alpha (index 0 in allHosts) is saved; its block grew by 1 line (lineDelta=+1).
	updatedAlpha := config.Host{
		Alias:      "alpha",
		Hostname:   "alpha.example.com",
		Port:       "22",
		SourceFile: "/home/user/.ssh/config",
		LineStart:  1, // newLineStart unchanged (no group was added in this scenario)
		Groups:     []string{},
	}

	// Find alpha's index in allHosts (alphabetical, no frequent hosts → alpha first)
	alphaIdx := -1
	for i, h := range m.allHosts {
		if h.Alias == "alpha" {
			alphaIdx = i
			break
		}
	}
	if alphaIdx < 0 {
		t.Fatal("alpha not found in allHosts")
	}
	betaIdx := -1
	for i, h := range m.allHosts {
		if h.Alias == "beta" {
			betaIdx = i
			break
		}
	}
	if betaIdx < 0 {
		t.Fatal("beta not found in allHosts")
	}

	newModel, _ := m.Update(editSavedMsg{
		updated:           updatedAlpha,
		index:             alphaIdx,
		lineDelta:         1,             // block grew by 1
		originalLineStart: 1,             // alpha's LineStart before the save
		sourceFile:        "/home/user/.ssh/config",
	})
	m = newModel.(Model)

	// beta (same file, LineStart=5 > originalLineStart=1) must be shifted to 6
	if m.allHosts[betaIdx].LineStart != 6 {
		t.Errorf("expected beta.LineStart=6 after drift, got %d", m.allHosts[betaIdx].LineStart)
	}

	// gamma (different file) must be unchanged
	gammaIdx := -1
	for i, h := range m.allHosts {
		if h.Alias == "gamma" {
			gammaIdx = i
			break
		}
	}
	if gammaIdx < 0 {
		t.Fatal("gamma not found in allHosts")
	}
	if m.allHosts[gammaIdx].LineStart != 1 {
		t.Errorf("expected gamma.LineStart=1 (unchanged), got %d", m.allHosts[gammaIdx].LineStart)
	}
}

// TestNewNoFrequent_FlatAlphabeticalOrder tests that noFrequent=true ignores connection counts.
func TestNewNoFrequent_FlatAlphabeticalOrder(t *testing.T) {
	hosts := makeHosts("gamma", "alpha", "beta")
	st := makeState(map[string]int{"gamma": 10})
	m := New(hosts, st, "/tmp/state.json", true)

	if len(m.allHosts) != 3 {
		t.Fatalf("expected 3 hosts, got %d", len(m.allHosts))
	}
	if m.allHosts[0].Alias != "alpha" {
		t.Errorf("expected allHosts[0]=alpha, got %s", m.allHosts[0].Alias)
	}
	if m.allHosts[1].Alias != "beta" {
		t.Errorf("expected allHosts[1]=beta, got %s", m.allHosts[1].Alias)
	}
	if m.allHosts[2].Alias != "gamma" {
		t.Errorf("expected allHosts[2]=gamma, got %s", m.allHosts[2].Alias)
	}
}

// TestNewNoFrequent_StoresStateRef tests that state is stored even when noFrequent=true.
func TestNewNoFrequent_StoresStateRef(t *testing.T) {
	hosts := makeHosts("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", true)

	if m.state == nil {
		t.Error("expected m.state to be non-nil when noFrequent=true")
	}
}

// TestView_EmptyHostList_NoPanic verifies that View() does not panic when the host list is empty.
func TestView_EmptyHostList_NoPanic(t *testing.T) {
	m := New([]config.Host{}, makeState(make(map[string]int)), "/tmp/state.json", false)
	m.viewHeight = 10
	_ = m.View()
}

// TestView_CursorAtLastHost_NoPanic verifies that View() does not panic when the cursor is at the last host.
func TestView_CursorAtLastHost_NoPanic(t *testing.T) {
	hosts := makeHosts("alpha", "beta", "gamma")
	m := New(hosts, makeState(make(map[string]int)), "/tmp/state.json", false)
	m.viewHeight = 10
	m.cursor = len(m.filtered) - 1
	_ = m.View()
}

// TestNormalMode_EscQuits tests that pressing Esc in normal mode returns a quit command.
func TestNormalMode_EscQuits(t *testing.T) {
	hosts := makeHosts("alpha")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json", false)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
}
