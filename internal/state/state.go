package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/platform"
)

// State represents the persistent state of SwiftSSH, tracking connection history.
type State struct {
	Connections map[string]int `json:"connections"` // key: host alias, value: count
	FirstRun    bool           `json:"first_run"`
}

// Load loads the state from the given path.
// If the file does not exist, it returns a new State with FirstRun: true.
// Any other error is returned.
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{FirstRun: true, Connections: make(map[string]int)}, nil
		}
		return nil, err
	}

	s := &State{}
	if err := json.Unmarshal(data, s); err != nil {
		// Corrupted state file — treat as a fresh install rather than erroring.
		return &State{Connections: make(map[string]int)}, nil
	}

	// Guard: if Connections is nil after unmarshal, initialize to empty map.
	if s.Connections == nil {
		s.Connections = make(map[string]int)
	}

	return s, nil
}

// Save saves the state to the given path.
// It writes to a temporary file first, then atomically replaces the original.
// The parent directory is created if it does not exist.
func Save(path string, s *State) error {
	// Ensure parent directory exists.
	if err := platform.EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	// Marshal state to JSON with indentation.
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	// Write to temporary file.
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	// Atomically replace the original file.
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on failure.
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}

// RecordConnection increments the connection count for the given host alias.
func RecordConnection(s *State, alias string) {
	s.Connections[alias]++
}

// FrequentHosts returns the top n most frequently connected hosts from the given list,
// sorted by connection count in descending order.
// If n <= 0 or n >= len(candidates), all candidates are returned.
// Hosts with 0 connections are excluded.
func FrequentHosts(s *State, hosts []config.Host, n int) []config.Host {
	// Build candidates: only hosts with at least one connection.
	candidates := []config.Host{}
	for _, h := range hosts {
		if s.Connections[h.Alias] > 0 {
			candidates = append(candidates, h)
		}
	}

	// Sort by connection count (descending) using stable sort to preserve order for ties.
	sort.SliceStable(candidates, func(i, j int) bool {
		return s.Connections[candidates[i].Alias] > s.Connections[candidates[j].Alias]
	})

	// Return top n.
	if n <= 0 || n >= len(candidates) {
		return candidates
	}
	return candidates[:n]
}
