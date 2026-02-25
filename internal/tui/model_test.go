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

// TestCursorWraps tests that the cursor wraps around at the ends of the list.
func TestCursorWraps(t *testing.T) {
	t.Helper()

	hosts := makeHosts("alpha", "beta", "gamma")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 10

	// Test wrap forward: press j 3 times
	m = pressKey(m, "j")
	if m.cursor != 1 {
		t.Errorf("After first j: expected cursor=1, got %d", m.cursor)
	}
	m = pressKey(m, "j")
	if m.cursor != 2 {
		t.Errorf("After second j: expected cursor=2, got %d", m.cursor)
	}
	m = pressKey(m, "j")
	if m.cursor != 0 {
		t.Errorf("After third j (wrap): expected cursor=0, got %d", m.cursor)
	}
	if m.viewport != 0 {
		t.Errorf("After wrap forward: expected viewport=0, got %d", m.viewport)
	}

	// Test wrap backward: set cursor to 0, press k once
	m.cursor = 0
	m = pressKey(m, "k")
	if m.cursor != 2 {
		t.Errorf("After k from cursor=0 (wrap): expected cursor=2, got %d", m.cursor)
	}
}

// TestViewportAdvances tests that the viewport advances when the cursor scrolls past the bottom.
func TestViewportAdvances(t *testing.T) {
	t.Helper()

	hosts := makeHosts("a", "b", "c", "d", "e")
	st := makeState(make(map[string]int))
	m := New(hosts, st, "/tmp/state.json")
	m.viewHeight = 3

	// Press j 3 times to move cursor from 0 to 3
	m = pressKey(m, "j")
	m = pressKey(m, "j")
	m = pressKey(m, "j")

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

	// Press k three times to move cursor from 3 to 0
	m = pressKey(m, "k")
	m = pressKey(m, "k")
	m = pressKey(m, "k")

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
