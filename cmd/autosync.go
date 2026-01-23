package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"gopkg.in/yaml.v3"
)

// autoSyncRegistryFilenames scans all registry files and fixes filename mismatches silently
func autoSyncRegistryFilenames(reg *registry.Registry) {
	paths := reg.Paths()

	// Scan each directory
	syncDirectory(paths.Rules)
	syncDirectory(paths.Workflows)
	syncDirectory(paths.Guidelines)
	syncDirectory(paths.Profiles)
}

func syncDirectory(baseDir string) {
	// Walk the directory recursively
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		// Try to sync this file silently
		syncFileSilently(path, baseDir)
		return nil
	})
}

func syncFileSilently(filePath string, baseDir string) {
	// Read and parse the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	// Extract frontmatter
	frontmatter, _, err := extractFrontmatterBytes(content)
	if err != nil || len(frontmatter) == 0 {
		return
	}

	// Parse frontmatter to get name
	var meta struct {
		Name string `yaml:"name"`
	}

	if err := yaml.Unmarshal(frontmatter, &meta); err != nil || meta.Name == "" {
		return
	}

	// Get relative path from base directory
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return
	}

	// Current name includes subdirectory if present (e.g., "auth/custom-auth")
	currentName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

	// Check if name matches
	if currentName == meta.Name {
		// Already in sync
		return
	}

	// Names don't match, need to rename
	newFilePath := filepath.Join(baseDir, meta.Name+filepath.Ext(filePath))

	// Check if target already exists
	if _, err := os.Stat(newFilePath); err == nil {
		// Can't rename, target exists
		return
	}

	// Create subdirectories if needed
	newDir := filepath.Dir(newFilePath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return
	}

	// Rename the file silently
	if err := os.Rename(filePath, newFilePath); err != nil {
		return
	}

	// Clean up empty directories
	oldDir := filepath.Dir(filePath)
	if oldDir != baseDir {
		// Try to remove old directory (will only succeed if empty)
		os.Remove(oldDir)
	}
}

// Helper to extract frontmatter as bytes
func extractFrontmatterBytes(content []byte) ([]byte, []byte, error) {
	if len(content) < 4 || string(content[:4]) != "---\n" {
		return nil, content, nil
	}

	// Find the closing ---
	end := -1
	for i := 4; i < len(content)-3; i++ {
		if content[i] == '\n' && string(content[i+1:i+4]) == "---" {
			end = i + 1
			break
		}
	}

	if end == -1 {
		return nil, content, fmt.Errorf("unclosed frontmatter")
	}

	frontmatter := content[4:end]
	markdown := content[end+3:]

	// Trim leading newlines from markdown
	for len(markdown) > 0 && (markdown[0] == '\n' || markdown[0] == '\r') {
		markdown = markdown[1:]
	}

	return frontmatter, markdown, nil
}
