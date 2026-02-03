package registry

// Item represents a generic registry item (rule, workflow, guideline, or custom type)
type Item struct {
	Type        string // e.g., "rule", "workflow", "framework"
	Name        string
	Description string
	Content     string // Markdown content (below frontmatter)
	FilePath    string // Path to the .md file
}

// Profile represents a directives.md template
type Profile struct {
	Name        string
	Description string
	Content     string // Full directives.md template content
	FilePath    string
}

// Registry manages the ~/.agmd/ directory
type Registry struct {
	BasePath string // ~/.agmd
}
