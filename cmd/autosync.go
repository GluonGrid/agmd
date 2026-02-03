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
	// Get all type directories
	types, err := reg.ListTypes()
	if err != nil {
		return
	}

	for _, typeName := range types {
		syncDirectory(filepath.Join(reg.BasePath, typeName))
	}
}

func syncDirectory(baseDir string) {
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		syncFileSilently(path, baseDir)
		return nil
	})
}

func syncFileSilently(filePath string, baseDir string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	frontmatter, _, err := extractFrontmatterBytes(content)
	if err != nil || len(frontmatter) == 0 {
		return
	}

	var meta struct {
		Name string `yaml:"name"`
	}

	if err := yaml.Unmarshal(frontmatter, &meta); err != nil || meta.Name == "" {
		return
	}

	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return
	}

	currentName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

	if currentName == meta.Name {
		return
	}

	newFilePath := filepath.Join(baseDir, meta.Name+filepath.Ext(filePath))

	if _, err := os.Stat(newFilePath); err == nil {
		return
	}

	newDir := filepath.Dir(newFilePath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return
	}

	if err := os.Rename(filePath, newFilePath); err != nil {
		return
	}

	oldDir := filepath.Dir(filePath)
	if oldDir != baseDir {
		os.Remove(oldDir)
	}
}

func extractFrontmatterBytes(content []byte) ([]byte, []byte, error) {
	if len(content) < 4 || string(content[:4]) != "---\n" {
		return nil, content, nil
	}

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

	for len(markdown) > 0 && (markdown[0] == '\n' || markdown[0] == '\r') {
		markdown = markdown[1:]
	}

	return frontmatter, markdown, nil
}
