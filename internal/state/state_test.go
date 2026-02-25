package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/srava/swiftssh/internal/config"
	"github.com/srava/swiftssh/internal/testutil"
)

// tempStatePath creates a temporary file path for testing.
func tempStatePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

// TestLoad_NonExistentFile verifies that loading a non-existent file returns a new state with FirstRun: true.
func TestLoad_NonExistentFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "state.json")
	s, err := Load(path)

	testutil.AssertNoError(t, err, "Load should not error on non-existent file")
	testutil.AssertNotNil(t, s, "Load should return non-nil state")
	testutil.AssertTrue(t, s.FirstRun, "FirstRun should be true for new state")
	testutil.AssertEqual(t, len(s.Connections), 0, "Connections should be empty for new state")
}

// TestLoadSave_RoundTrip verifies that state is preserved through a save/load cycle.
func TestLoadSave_RoundTrip(t *testing.T) {
	path := tempStatePath(t)

	// Create initial state with some data.
	original := &State{
		FirstRun: false,
		Connections: map[string]int{
			"dev":     10,
			"prod":    5,
			"staging": 2,
		},
	}

	// Save to disk.
	err := Save(path, original)
	testutil.AssertNoError(t, err, "Save should not error")

	// Load from disk.
	loaded, err := Load(path)
	testutil.AssertNoError(t, err, "Load should not error")
	testutil.AssertNotNil(t, loaded, "Load should return non-nil state")

	// Verify data integrity.
	testutil.AssertFalse(t, loaded.FirstRun, "FirstRun should match after round-trip")
	testutil.AssertEqual(t, len(loaded.Connections), 3, "Connections length should match")
	testutil.AssertEqual(t, loaded.Connections["dev"], 10, "dev count should match")
	testutil.AssertEqual(t, loaded.Connections["prod"], 5, "prod count should match")
	testutil.AssertEqual(t, loaded.Connections["staging"], 2, "staging count should match")
}

// TestRecordConnection verifies that recording connections increments the count.
func TestRecordConnection(t *testing.T) {
	s := &State{
		FirstRun:    false,
		Connections: make(map[string]int),
	}

	// First call should set count to 1.
	RecordConnection(s, "myhost")
	testutil.AssertEqual(t, s.Connections["myhost"], 1, "First connection should set count to 1")

	// Second call should increment to 2.
	RecordConnection(s, "myhost")
	testutil.AssertEqual(t, s.Connections["myhost"], 2, "Second connection should increment to 2")

	// Recording a different host should not affect the first.
	RecordConnection(s, "other")
	testutil.AssertEqual(t, s.Connections["myhost"], 2, "myhost count should still be 2")
	testutil.AssertEqual(t, s.Connections["other"], 1, "other count should be 1")
}

// TestFrequentHosts_TopN verifies that the top n hosts are returned sorted by count.
func TestFrequentHosts_TopN(t *testing.T) {
	hosts := []config.Host{
		{Alias: "first", Hostname: "host1.com"},
		{Alias: "second", Hostname: "host2.com"},
		{Alias: "third", Hostname: "host3.com"},
	}

	s := &State{
		FirstRun: false,
		Connections: map[string]int{
			"first":  5,
			"second": 3,
			"third":  1,
		},
	}

	// Get top 2.
	frequent := FrequentHosts(s, hosts, 2)

	testutil.AssertEqual(t, len(frequent), 2, "Should return top 2 hosts")
	testutil.AssertEqual(t, frequent[0].Alias, "first", "First should be 'first' (count 5)")
	testutil.AssertEqual(t, frequent[1].Alias, "second", "Second should be 'second' (count 3)")
}

// TestFrequentHosts_FewerThanN verifies that all hosts are returned when there are fewer than n.
func TestFrequentHosts_FewerThanN(t *testing.T) {
	hosts := []config.Host{
		{Alias: "alpha", Hostname: "alpha.com"},
		{Alias: "beta", Hostname: "beta.com"},
	}

	s := &State{
		FirstRun: false,
		Connections: map[string]int{
			"alpha": 5,
			"beta":  2,
		},
	}

	// Request top 10 but only 2 available.
	frequent := FrequentHosts(s, hosts, 10)

	testutil.AssertEqual(t, len(frequent), 2, "Should return all 2 hosts when n > available")
	testutil.AssertEqual(t, frequent[0].Alias, "alpha", "First should be 'alpha' (count 5)")
	testutil.AssertEqual(t, frequent[1].Alias, "beta", "Second should be 'beta' (count 2)")
}

// TestSave_MissingParentDirectory verifies that Save creates parent directories as needed.
func TestSave_MissingParentDirectory(t *testing.T) {
	// Create a path with nested non-existent directories.
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "level1", "level2", "level3", "state.json")

	s := &State{
		FirstRun: true,
		Connections: map[string]int{
			"test": 1,
		},
	}

	err := Save(path, s)
	testutil.AssertNoError(t, err, "Save should succeed even with missing parent directories")

	// Verify the file was created and contains the expected data.
	data, err := os.ReadFile(path)
	testutil.AssertNoError(t, err, "Should be able to read the saved file")

	var loaded State
	err = json.Unmarshal(data, &loaded)
	testutil.AssertNoError(t, err, "Saved JSON should be valid")
	testutil.AssertTrue(t, loaded.FirstRun, "FirstRun should be preserved")
	testutil.AssertEqual(t, loaded.Connections["test"], 1, "Connections should be preserved")
}
