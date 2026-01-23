package registry

import "time"

// Rule represents a rule in the registry
type Rule struct {
	Name        string    `yaml:"name"`
	Category    string    `yaml:"category,omitempty"`
	Description string    `yaml:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at,omitempty"`
	Content     string    // Markdown content (below frontmatter)
	FilePath    string    // Path to the .md file
}

// Workflow represents a workflow in the registry
type Workflow struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at,omitempty"`
	Content     string    // Markdown content
	FilePath    string
}

// Guideline represents a guideline in the registry
type Guideline struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at,omitempty"`
	Content     string    // Markdown content
	FilePath    string
}

// Profile represents a directives.md template in the registry
type Profile struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Content     string // Full directives.md template content
	FilePath    string // Path to the .md file
}

// Registry manages the ~/.agmd/ directory structure
type Registry struct {
	BasePath string // ~/.agmd
}

// RegistryPaths contains all registry subdirectory paths
type RegistryPaths struct {
	Base       string // ~/.agmd
	Shared     string // ~/.agmd/shared
	Rules      string // ~/.agmd/rules
	Workflows  string // ~/.agmd/workflows
	Guidelines string // ~/.agmd/guidelines
	Profiles   string // ~/.agmd/profiles
}
