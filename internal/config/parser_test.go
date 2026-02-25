package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/srava/swiftssh/internal/testutil"
)

// Helper: writeTempConfig creates a temporary config file in t.TempDir() with the given content.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return configPath
}

// Helper: writeTempConfigAt creates a file at a specific path within dir.
func writeTempConfigAt(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return path
}

// TestParse_BasicSingleHost verifies parsing a single host with default port.
func TestParse_BasicSingleHost(t *testing.T) {
	t.Run("single host basic", func(t *testing.T) {
		content := `Host myserver
Hostname example.com
User john
Port 2222
`
		configPath := writeTempConfig(t, content)
		hosts, err := Parse(configPath)

		testutil.AssertNoError(t, err, "Parse should not error")
		if len(hosts) != 1 {
			t.Fatalf("expected 1 host, got %d", len(hosts))
		}

		testutil.AssertStringEqual(t, hosts[0].Alias, "myserver", "Alias mismatch")
		testutil.AssertStringEqual(t, hosts[0].Hostname, "example.com", "Hostname mismatch")
		testutil.AssertStringEqual(t, hosts[0].User, "john", "User mismatch")
		testutil.AssertStringEqual(t, hosts[0].Port, "2222", "Port mismatch")
	})

	t.Run("single host with default port", func(t *testing.T) {
		content := `Host myserver
Hostname example.com
User john
`
		configPath := writeTempConfig(t, content)
		hosts, err := Parse(configPath)

		testutil.AssertNoError(t, err, "Parse should not error")
		if len(hosts) != 1 {
			t.Fatalf("expected 1 host, got %d", len(hosts))
		}

		testutil.AssertStringEqual(t, hosts[0].Port, "22", "Port should default to 22")
	})
}

// TestParse_MultiHostAllFields verifies parsing multiple hosts with all fields.
func TestParse_MultiHostAllFields(t *testing.T) {
	content := `Host dev
Hostname dev.example.com
User alice
Port 2222

Host prod
Hostname prod.example.com
User bob
Port 3333
`
	configPath := writeTempConfig(t, content)
	hosts, err := Parse(configPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Check first host
	testutil.AssertStringEqual(t, hosts[0].Alias, "dev", "First host alias mismatch")
	testutil.AssertStringEqual(t, hosts[0].Hostname, "dev.example.com", "First host hostname mismatch")
	testutil.AssertStringEqual(t, hosts[0].User, "alice", "First host user mismatch")
	testutil.AssertStringEqual(t, hosts[0].Port, "2222", "First host port mismatch")

	// Check second host
	testutil.AssertStringEqual(t, hosts[1].Alias, "prod", "Second host alias mismatch")
	testutil.AssertStringEqual(t, hosts[1].Hostname, "prod.example.com", "Second host hostname mismatch")
	testutil.AssertStringEqual(t, hosts[1].User, "bob", "Second host user mismatch")
	testutil.AssertStringEqual(t, hosts[1].Port, "3333", "Second host port mismatch")
}

// TestParse_MagicCommentBasic verifies magic comment parsing.
func TestParse_MagicCommentBasic(t *testing.T) {
	content := `# @group Work, Personal
Host myserver
Hostname example.com
`
	configPath := writeTempConfig(t, content)
	hosts, err := Parse(configPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}

	expected := []string{"Work", "Personal"}
	testutil.AssertSliceEqual(t, hosts[0].Groups, expected, "Groups mismatch")
}

// TestParse_MagicCommentWhitespace verifies whitespace handling in magic comments.
func TestParse_MagicCommentWhitespace(t *testing.T) {
	t.Run("extra spaces around commas", func(t *testing.T) {
		content := `# @group   Work  ,  Personal  ,  Finance
Host myserver
Hostname example.com
`
		configPath := writeTempConfig(t, content)
		hosts, err := Parse(configPath)

		testutil.AssertNoError(t, err, "Parse should not error")
		if len(hosts) != 1 {
			t.Fatalf("expected 1 host, got %d", len(hosts))
		}

		expected := []string{"Work", "Personal", "Finance"}
		testutil.AssertSliceEqual(t, hosts[0].Groups, expected, "Groups with extra whitespace")
	})

	t.Run("tabs in group list", func(t *testing.T) {
		content := "# @group\tWork\t,\tPersonal\nHost myserver\nHostname example.com\n"
		configPath := writeTempConfig(t, content)
		hosts, err := Parse(configPath)

		testutil.AssertNoError(t, err, "Parse should not error")
		if len(hosts) != 1 {
			t.Fatalf("expected 1 host, got %d", len(hosts))
		}

		expected := []string{"Work", "Personal"}
		testutil.AssertSliceEqual(t, hosts[0].Groups, expected, "Groups with tabs")
	})
}

// TestParse_DuplicateHostBlocks verifies duplicate host blocks are preserved.
func TestParse_DuplicateHostBlocks(t *testing.T) {
	content := `Host dev
Hostname dev1.example.com
User alice

Host dev
Hostname dev2.example.com
User bob
`
	configPath := writeTempConfig(t, content)
	hosts, err := Parse(configPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts (duplicates preserved), got %d", len(hosts))
	}

	// Both should have alias "dev"
	testutil.AssertStringEqual(t, hosts[0].Alias, "dev", "First duplicate alias")
	testutil.AssertStringEqual(t, hosts[1].Alias, "dev", "Second duplicate alias")

	// But different hostnames
	testutil.AssertStringEqual(t, hosts[0].Hostname, "dev1.example.com", "First duplicate hostname")
	testutil.AssertStringEqual(t, hosts[1].Hostname, "dev2.example.com", "Second duplicate hostname")
}

// TestParse_IncludeRelativePath verifies relative include path resolution.
func TestParse_IncludeRelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create main config
	mainConfigContent := `Host main
Hostname main.example.com

Include conf/included.conf
`
	mainConfigPath := writeTempConfigAt(t, tempDir, "config", mainConfigContent)

	// Create included config
	writeTempConfigAt(t, tempDir, "conf/included.conf", `Host included
Hostname included.example.com
`)

	hosts, err := Parse(mainConfigPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts (main + included), got %d", len(hosts))
	}

	testutil.AssertStringEqual(t, hosts[0].Alias, "main", "First host should be main")
	testutil.AssertStringEqual(t, hosts[1].Alias, "included", "Second host should be included")
}

// TestParse_IncludeGlobPattern verifies glob expansion in include directives.
func TestParse_IncludeGlobPattern(t *testing.T) {
	tempDir := t.TempDir()

	// Create main config
	mainConfigContent := `Host main
Hostname main.example.com

Include conf.d/*.conf
`
	mainConfigPath := writeTempConfigAt(t, tempDir, "config", mainConfigContent)

	// Create multiple config files
	writeTempConfigAt(t, tempDir, "conf.d/01-servers.conf", `Host server1
Hostname server1.example.com
`)
	writeTempConfigAt(t, tempDir, "conf.d/02-servers.conf", `Host server2
Hostname server2.example.com
`)

	hosts, err := Parse(mainConfigPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	// We should have 3 hosts: main + server1 + server2 (glob ordering may vary)
	if len(hosts) < 3 {
		t.Fatalf("expected at least 3 hosts from glob, got %d", len(hosts))
	}

	aliases := make(map[string]bool)
	for _, h := range hosts {
		aliases[h.Alias] = true
	}

	testutil.AssertTrue(t, aliases["main"], "main host should be present")
	testutil.AssertTrue(t, aliases["server1"], "server1 should be found via glob")
	testutil.AssertTrue(t, aliases["server2"], "server2 should be found via glob")
}

// TestParse_IncludeRecursive verifies recursive includes (A→B→C).
func TestParse_IncludeRecursive(t *testing.T) {
	tempDir := t.TempDir()

	// Create A (main config)
	mainConfigContent := `Host hostA
Hostname a.example.com

Include confB.conf
`
	mainConfigPath := writeTempConfigAt(t, tempDir, "config", mainConfigContent)

	// Create B (includes C)
	writeTempConfigAt(t, tempDir, "confB.conf", `Host hostB
Hostname b.example.com

Include confC.conf
`)

	// Create C
	writeTempConfigAt(t, tempDir, "confC.conf", `Host hostC
Hostname c.example.com
`)

	hosts, err := Parse(mainConfigPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 3 {
		t.Fatalf("expected 3 hosts from recursive includes, got %d", len(hosts))
	}

	expectedAliases := []string{"hostA", "hostB", "hostC"}
	if len(hosts) != len(expectedAliases) {
		t.Fatalf("expected %d hosts, got %d", len(expectedAliases), len(hosts))
	}

	for i, expected := range expectedAliases {
		testutil.AssertStringEqual(t, hosts[i].Alias, expected, fmt.Sprintf("Host %d alias", i))
	}
}

// TestParse_IncludeCircular verifies circular includes don't cause infinite loops.
func TestParse_IncludeCircular(t *testing.T) {
	tempDir := t.TempDir()

	// Create A (main config, includes B)
	mainConfigContent := `Host hostA
Hostname a.example.com

Include confB.conf
`
	mainConfigPath := writeTempConfigAt(t, tempDir, "config", mainConfigContent)

	// Create B (includes A, creating a cycle)
	writeTempConfigAt(t, tempDir, "confB.conf", fmt.Sprintf(`Host hostB
Hostname b.example.com

Include %s
`, mainConfigPath))

	// This should not hang or error, just return both hosts
	hosts, err := Parse(mainConfigPath)

	testutil.AssertNoError(t, err, "Parse should not error on circular includes")
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts (circular ref should skip duplicate), got %d", len(hosts))
	}

	testutil.AssertStringEqual(t, hosts[0].Alias, "hostA", "hostA should be present")
	testutil.AssertStringEqual(t, hosts[1].Alias, "hostB", "hostB should be present")
}

// TestParse_WildcardHostExcluded verifies Host * blocks are not included in results.
func TestParse_WildcardHostExcluded(t *testing.T) {
	content := `Host *
User defaultuser
Port 2222

Host myserver
Hostname example.com
User john
`
	configPath := writeTempConfig(t, content)
	hosts, err := Parse(configPath)

	testutil.AssertNoError(t, err, "Parse should not error")
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (wildcard excluded), got %d", len(hosts))
	}

	testutil.AssertStringEqual(t, hosts[0].Alias, "myserver", "Only myserver should be in results")
}

// TestParse_MissingIncludedFile verifies missing includes are handled gracefully.
func TestParse_MissingIncludedFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create main config that includes a non-existent file
	mainConfigContent := `Host main
Hostname main.example.com

Include /nonexistent/path/to/config.conf
`
	mainConfigPath := writeTempConfigAt(t, tempDir, "config", mainConfigContent)

	// Should not error, but should have only the main host
	hosts, err := Parse(mainConfigPath)

	testutil.AssertNoError(t, err, "Parse should not error on missing include")
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (missing include ignored), got %d", len(hosts))
	}

	testutil.AssertStringEqual(t, hosts[0].Alias, "main", "main host should still be parsed")
}
