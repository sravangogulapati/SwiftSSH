package config

import (
	"testing"
)

// TestHostStructCreation validates Host struct creation and field access.
func TestHostStructCreation(t *testing.T) {
	t.Run("creates Host with all fields", func(t *testing.T) {
		h := Host{
			Alias:      "dev-server",
			Hostname:   "192.168.1.10",
			User:       "ubuntu",
			Port:       "2222",
			Groups:     []string{"Work", "Development"},
			SourceFile: "/home/user/.ssh/config",
		}

		if h.Alias != "dev-server" {
			t.Errorf("Alias: expected 'dev-server', got '%s'", h.Alias)
		}
		if h.Hostname != "192.168.1.10" {
			t.Errorf("Hostname: expected '192.168.1.10', got '%s'", h.Hostname)
		}
		if h.User != "ubuntu" {
			t.Errorf("User: expected 'ubuntu', got '%s'", h.User)
		}
		if h.Port != "2222" {
			t.Errorf("Port: expected '2222', got '%s'", h.Port)
		}
		if len(h.Groups) != 2 {
			t.Errorf("Groups: expected 2 groups, got %d", len(h.Groups))
		}
		if h.Groups[0] != "Work" || h.Groups[1] != "Development" {
			t.Errorf("Groups: expected [Work, Development], got %v", h.Groups)
		}
		if h.SourceFile != "/home/user/.ssh/config" {
			t.Errorf("SourceFile: expected '/home/user/.ssh/config', got '%s'", h.SourceFile)
		}
	})

	t.Run("creates Host with empty fields", func(t *testing.T) {
		h := Host{
			Alias: "test",
		}

		if h.Alias != "test" {
			t.Errorf("Alias should be set to 'test'")
		}
		if h.Hostname != "" {
			t.Errorf("Hostname should be empty string, got '%s'", h.Hostname)
		}
		if h.User != "" {
			t.Errorf("User should be empty string, got '%s'", h.User)
		}
		if h.Port != "" {
			t.Errorf("Port should be empty string, got '%s'", h.Port)
		}
		if h.Groups != nil && len(h.Groups) > 0 {
			t.Errorf("Groups should be empty, got %v", h.Groups)
		}
	})

	t.Run("creates Host with nil Groups", func(t *testing.T) {
		h := Host{
			Alias:    "test",
			Hostname: "example.com",
		}

		if h.Groups != nil && len(h.Groups) > 0 {
			t.Errorf("Groups should be empty or nil, got %v", h.Groups)
		}
	})

	t.Run("Host supports field mutation", func(t *testing.T) {
		h := Host{Alias: "original"}

		h.Alias = "modified"
		h.Port = "22"
		h.Groups = []string{"Tag1"}

		if h.Alias != "modified" {
			t.Error("Alias mutation failed")
		}
		if h.Port != "22" {
			t.Error("Port mutation failed")
		}
		if len(h.Groups) != 1 || h.Groups[0] != "Tag1" {
			t.Error("Groups mutation failed")
		}
	})
}

// TestHostGroups validates Group handling.
func TestHostGroups(t *testing.T) {
	t.Run("single group", func(t *testing.T) {
		h := Host{Groups: []string{"Work"}}
		if len(h.Groups) != 1 || h.Groups[0] != "Work" {
			t.Error("single group assignment failed")
		}
	})

	t.Run("multiple groups", func(t *testing.T) {
		groups := []string{"Work", "Client A", "Production"}
		h := Host{Groups: groups}

		if len(h.Groups) != 3 {
			t.Errorf("expected 3 groups, got %d", len(h.Groups))
		}
		for i, expected := range groups {
			if h.Groups[i] != expected {
				t.Errorf("group %d: expected '%s', got '%s'", i, expected, h.Groups[i])
			}
		}
	})

	t.Run("groups with spaces and special chars", func(t *testing.T) {
		groups := []string{"My Work", "Client-A", "Prod_2024"}
		h := Host{Groups: groups}

		if len(h.Groups) != 3 {
			t.Errorf("expected 3 groups, got %d", len(h.Groups))
		}
		for i, expected := range groups {
			if h.Groups[i] != expected {
				t.Errorf("group %d mismatch", i)
			}
		}
	})

	t.Run("empty Groups slice", func(t *testing.T) {
		h := Host{Groups: []string{}}
		if h.Groups == nil {
			t.Error("empty Groups slice became nil")
		}
		if len(h.Groups) != 0 {
			t.Error("empty Groups should have length 0")
		}
	})

	t.Run("can append to Groups", func(t *testing.T) {
		h := Host{Groups: []string{"Initial"}}
		h.Groups = append(h.Groups, "New")

		if len(h.Groups) != 2 {
			t.Errorf("expected 2 groups after append, got %d", len(h.Groups))
		}
		if h.Groups[1] != "New" {
			t.Error("appended group not found")
		}
	})
}

// TestHostPortField validates Port field handling.
func TestHostPortField(t *testing.T) {
	t.Run("standard ports", func(t *testing.T) {
		ports := []string{"22", "2222", "8022", "65535"}
		for _, port := range ports {
			h := Host{Port: port}
			if h.Port != port {
				t.Errorf("port %s not set correctly", port)
			}
		}
	})

	t.Run("empty port string", func(t *testing.T) {
		h := Host{Port: ""}
		if h.Port != "" {
			t.Error("empty port should remain empty")
		}
	})
}

// TestParsedConfigStructCreation validates ParsedConfig struct creation.
func TestParsedConfigStructCreation(t *testing.T) {
	t.Run("creates ParsedConfig with hosts", func(t *testing.T) {
		hosts := []Host{
			{Alias: "host1", Hostname: "server1.com"},
			{Alias: "host2", Hostname: "server2.com"},
		}
		cfg := ParsedConfig{
			Hosts:      hosts,
			SourceFile: "/home/user/.ssh/config",
		}

		if len(cfg.Hosts) != 2 {
			t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
		}
		if cfg.Hosts[0].Alias != "host1" {
			t.Error("first host not correct")
		}
		if cfg.SourceFile != "/home/user/.ssh/config" {
			t.Error("SourceFile not correct")
		}
	})

	t.Run("creates ParsedConfig with empty hosts", func(t *testing.T) {
		cfg := ParsedConfig{
			Hosts:      []Host{},
			SourceFile: "/home/user/.ssh/config",
		}

		if cfg.Hosts == nil {
			t.Error("empty Hosts slice became nil")
		}
		if len(cfg.Hosts) != 0 {
			t.Error("empty Hosts should have length 0")
		}
	})

	t.Run("creates ParsedConfig with nil hosts", func(t *testing.T) {
		cfg := ParsedConfig{
			SourceFile: "/home/user/.ssh/config",
		}

		if cfg.Hosts != nil && len(cfg.Hosts) > 0 {
			t.Error("Hosts should be nil or empty")
		}
	})

	t.Run("ParsedConfig supports mutation", func(t *testing.T) {
		cfg := ParsedConfig{
			Hosts:      []Host{},
			SourceFile: "original",
		}

		cfg.SourceFile = "modified"
		cfg.Hosts = append(cfg.Hosts, Host{Alias: "new"})

		if cfg.SourceFile != "modified" {
			t.Error("SourceFile mutation failed")
		}
		if len(cfg.Hosts) != 1 {
			t.Error("Hosts mutation failed")
		}
	})
}

// TestParsedConfigHostManipulation validates host manipulation.
func TestParsedConfigHostManipulation(t *testing.T) {
	t.Run("can add hosts to ParsedConfig", func(t *testing.T) {
		cfg := ParsedConfig{
			Hosts:      []Host{},
			SourceFile: "/home/user/.ssh/config",
		}

		h1 := Host{Alias: "prod", Hostname: "prod.example.com"}
		h2 := Host{Alias: "staging", Hostname: "staging.example.com"}

		cfg.Hosts = append(cfg.Hosts, h1, h2)

		if len(cfg.Hosts) != 2 {
			t.Errorf("expected 2 hosts, got %d", len(cfg.Hosts))
		}
		if cfg.Hosts[0].Alias != "prod" {
			t.Error("first host incorrect")
		}
		if cfg.Hosts[1].Alias != "staging" {
			t.Error("second host incorrect")
		}
	})

	t.Run("can iterate ParsedConfig hosts", func(t *testing.T) {
		hosts := []Host{
			{Alias: "h1"},
			{Alias: "h2"},
			{Alias: "h3"},
		}
		cfg := ParsedConfig{Hosts: hosts}

		count := 0
		for i, h := range cfg.Hosts {
			count++
			if h.Alias != hosts[i].Alias {
				t.Errorf("iteration %d: alias mismatch", i)
			}
		}

		if count != 3 {
			t.Errorf("expected 3 iterations, got %d", count)
		}
	})

	t.Run("can filter hosts by criteria", func(t *testing.T) {
		cfg := ParsedConfig{
			Hosts: []Host{
				{Alias: "prod", Groups: []string{"Production"}},
				{Alias: "staging", Groups: []string{"Staging"}},
				{Alias: "prod2", Groups: []string{"Production"}},
			},
		}

		var prodHosts []Host
		for _, h := range cfg.Hosts {
			for _, g := range h.Groups {
				if g == "Production" {
					prodHosts = append(prodHosts, h)
					break
				}
			}
		}

		if len(prodHosts) != 2 {
			t.Errorf("expected 2 prod hosts, got %d", len(prodHosts))
		}
	})

	t.Run("can access host by index safely", func(t *testing.T) {
		cfg := ParsedConfig{
			Hosts: []Host{
				{Alias: "first"},
				{Alias: "second"},
				{Alias: "third"},
			},
		}

		if cfg.Hosts[0].Alias != "first" {
			t.Error("index 0 incorrect")
		}
		if cfg.Hosts[2].Alias != "third" {
			t.Error("index 2 incorrect")
		}
	})
}

// TestHostComparison validates Host equality and comparison.
func TestHostComparison(t *testing.T) {
	t.Run("identical hosts are equal", func(t *testing.T) {
		h1 := Host{
			Alias:      "dev",
			Hostname:   "dev.example.com",
			User:       "user",
			Port:       "22",
			Groups:     []string{"Work"},
			SourceFile: "/config",
		}
		h2 := Host{
			Alias:      "dev",
			Hostname:   "dev.example.com",
			User:       "user",
			Port:       "22",
			Groups:     []string{"Work"},
			SourceFile: "/config",
		}

		// Manual field comparison
		if h1.Alias != h2.Alias || h1.Hostname != h2.Hostname || h1.User != h2.User || h1.Port != h2.Port {
			t.Error("hosts should be equal")
		}
	})

	t.Run("hosts with different aliases are not equal", func(t *testing.T) {
		h1 := Host{Alias: "dev", Hostname: "same.com"}
		h2 := Host{Alias: "prod", Hostname: "same.com"}

		if h1.Alias == h2.Alias {
			t.Error("hosts should not be equal")
		}
	})
}

// TestEdgeCases validates edge cases and boundary conditions.
func TestEdgeCases(t *testing.T) {
	t.Run("Host with very long strings", func(t *testing.T) {
		longString := ""
		for i := 0; i < 1000; i++ {
			longString += "a"
		}

		h := Host{
			Alias:      longString,
			Hostname:   longString,
			User:       longString,
			SourceFile: longString,
		}

		if len(h.Alias) != 1000 {
			t.Error("long alias not stored correctly")
		}
	})

	t.Run("Host with empty Group in slice", func(t *testing.T) {
		h := Host{
			Groups: []string{"Work", "", "Personal"},
		}

		if len(h.Groups) != 3 {
			t.Errorf("expected 3 groups including empty one, got %d", len(h.Groups))
		}
		if h.Groups[1] != "" {
			t.Error("empty group string not preserved")
		}
	})

	t.Run("ParsedConfig with large number of hosts", func(t *testing.T) {
		const numHosts = 10000
		hosts := make([]Host, numHosts)
		for i := 0; i < numHosts; i++ {
			hosts[i] = Host{Alias: "host-" + string(rune(i))}
		}

		cfg := ParsedConfig{
			Hosts:      hosts,
			SourceFile: "/config",
		}

		if len(cfg.Hosts) != numHosts {
			t.Errorf("expected %d hosts, got %d", numHosts, len(cfg.Hosts))
		}

		// Spot-check a few entries
		if cfg.Hosts[0].Alias != "host-\x00" {
			t.Error("first host incorrect")
		}
		if cfg.Hosts[numHosts-1].Alias != "host-\u270f" {
			t.Error("last host incorrect")
		}
	})

	t.Run("Host with Unicode characters", func(t *testing.T) {
		h := Host{
			Alias:      "dev-日本語",
			Hostname:   "example.中国",
			Groups:     []string{"Ελληνικά", "Русский"},
			SourceFile: "/home/user/.ssh/config-🔑",
		}

		if h.Alias != "dev-日本語" {
			t.Error("Unicode alias not preserved")
		}
		if len(h.Groups) != 2 {
			t.Error("Unicode groups not preserved")
		}
	})

	t.Run("ParsedConfig zero initialization", func(t *testing.T) {
		var cfg ParsedConfig

		// All fields should have zero values
		if cfg.Hosts != nil {
			t.Error("Hosts should be nil on zero initialization")
		}
		if cfg.SourceFile != "" {
			t.Error("SourceFile should be empty on zero initialization")
		}
	})
}

// TestTypeCompatibility validates that types work as expected with common patterns.
func TestTypeCompatibility(t *testing.T) {
	t.Run("Host can be used in maps by alias", func(t *testing.T) {
		hosts := []Host{
			{Alias: "prod", Hostname: "prod.example.com"},
			{Alias: "dev", Hostname: "dev.example.com"},
		}

		hostMap := make(map[string]Host)
		for _, h := range hosts {
			hostMap[h.Alias] = h
		}

		if hostMap["prod"].Hostname != "prod.example.com" {
			t.Error("map lookup failed")
		}
	})

	t.Run("Host can be used in function parameters", func(t *testing.T) {
		testFunc := func(h Host) string {
			return h.Alias + "@" + h.Hostname
		}

		h := Host{Alias: "user", Hostname: "example.com"}
		result := testFunc(h)

		if result != "user@example.com" {
			t.Errorf("expected 'user@example.com', got '%s'", result)
		}
	})

	t.Run("ParsedConfig can be passed by pointer", func(t *testing.T) {
		testFunc := func(cfg *ParsedConfig) int {
			return len(cfg.Hosts)
		}

		cfg := &ParsedConfig{
			Hosts: []Host{
				{Alias: "h1"},
				{Alias: "h2"},
			},
		}

		if testFunc(cfg) != 2 {
			t.Error("pointer passing failed")
		}
	})
}

// BenchmarkHostCreation provides performance baselines for Host creation.
func BenchmarkHostCreation(b *testing.B) {
	b.Run("create simple host", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Host{
				Alias:    "test",
				Hostname: "example.com",
			}
		}
	})

	b.Run("create host with groups", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Host{
				Alias:    "test",
				Hostname: "example.com",
				Groups:   []string{"Work", "Dev", "Production"},
			}
		}
	})

	b.Run("create host with all fields", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Host{
				Alias:      "test",
				Hostname:   "example.com",
				User:       "ubuntu",
				Port:       "2222",
				Groups:     []string{"Work", "Dev", "Production"},
				SourceFile: "/home/user/.ssh/config",
			}
		}
	})
}

// BenchmarkParsedConfigOperations provides performance baselines for ParsedConfig operations.
func BenchmarkParsedConfigOperations(b *testing.B) {
	b.Run("create empty config", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ParsedConfig{
				Hosts:      []Host{},
				SourceFile: "/config",
			}
		}
	})

	b.Run("create config with 100 hosts", func(b *testing.B) {
		hosts := make([]Host, 100)
		for i := 0; i < 100; i++ {
			hosts[i] = Host{Alias: "host"}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ParsedConfig{Hosts: hosts}
		}
	})

	b.Run("append to config hosts", func(b *testing.B) {
		cfg := ParsedConfig{Hosts: []Host{}}
		newHost := Host{Alias: "new"}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cfg.Hosts = append(cfg.Hosts, newHost)
		}
	})
}
