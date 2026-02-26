package config

// Host represents a single SSH host entry from the config.
type Host struct {
	Alias        string   // The host alias (e.g., "dev" from "Host dev")
	Hostname     string   // The actual hostname or IP to connect to
	User         string   // The SSH user (defaults to current user if not specified)
	Port         string   // The SSH port (defaults to "22" if not specified)
	IdentityFile string   // Path to the private key file (IdentityFile directive)
	Groups       []string // Group tags parsed from magic comment "# @group Work, Personal"
	SourceFile   string   // The config file this host was parsed from (for Include support)
}

// ParsedConfig represents the complete parsed SSH configuration.
type ParsedConfig struct {
	Hosts      []Host // All hosts from the config file(s)
	SourceFile string // The primary config file path
}
