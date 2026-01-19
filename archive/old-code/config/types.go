package config

// AgmdFrontmatter represents the YAML front matter in agent config files
type AgmdFrontmatter struct {
	Version   string            `yaml:"version"`            // agmd format version (e.g., "1.0.0")
	Shared    string            `yaml:"shared,omitempty"`   // Path to universal shared config
	Profiles  []string          `yaml:"profiles,omitempty"` // List of profile names to inherit
	Extends   string            `yaml:"extends,omitempty"`  // For profiles: what they extend
	Type      string            `yaml:"type,omitempty"`     // universal/profile/project
	Overrides map[string]any    `yaml:"overrides,omitempty"` // Project-specific overrides
}

// AgmdFile represents a complete agent configuration file
type AgmdFile struct {
	Frontmatter AgmdFrontmatter // Parsed YAML frontmatter
	Content     string          // Raw markdown content (without frontmatter)
	Sections    []Section       // Parsed markdown sections
	Path        string          // File path on disk
}

// Section represents a markdown section (## heading and its content)
type Section struct {
	Title   string // e.g., "Code Quality Principles"
	Level   int    // Heading level: 2 for ##, 3 for ###
	Content string // Everything until next same-level heading
	Key     string // Normalized key: "code-quality-principles"
}

// ResolvedConfig is the result of merging all inheritance layers
type ResolvedConfig struct {
	Universal *AgmdFile   // Universal shared config
	Profiles  []*AgmdFile // Profiles in order of application
	Project   *AgmdFile   // Project-specific config
	Merged    *AgmdFile   // Final merged result (effective config)
}

// ValidationError represents a config validation error
type ValidationError struct {
	Type    string // Type of error (e.g., "missing_file", "invalid_yaml")
	Message string // Human-readable error message
	Path    string // File path where error occurred
}

// ValidationResult contains the results of config validation
type ValidationResult struct {
	Valid    bool              // Whether config is valid
	Errors   []ValidationError // List of errors found
	Warnings []string          // List of warnings (non-fatal)
}
