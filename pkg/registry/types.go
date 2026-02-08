package registry

// Item represents a generic registry item (rule, workflow, guideline, or custom type)
type Item struct {
	Type        string // e.g., "rule", "workflow", "framework"
	Name        string
	Description string
	Content     string // Markdown content (below frontmatter)
	FilePath    string // Path to the .md file
}

// ProfileFile represents a file to copy when initializing with a profile
type ProfileFile struct {
	Source      string // file:name reference in registry
	Destination string // destination path in project (optional, defaults to source name)
}

// Profile represents a directives.md template
type Profile struct {
	Name        string
	Description string
	Files       []ProfileFile // Files to copy on init
	Content     string        // Full directives.md template content (below frontmatter)
	RawContent  string        // Full content including frontmatter
	FilePath    string
}

// Registry manages the ~/.agmd/ directory
type Registry struct {
	BasePath string // ~/.agmd
}

// IsRawType returns true for types that store files without .md extension and frontmatter
func IsRawType(itemType string) bool {
	return itemType == "file"
}
