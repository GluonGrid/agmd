package registry

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ItemMeta represents the YAML frontmatter for an item
type ItemMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// ProfileMeta represents the YAML frontmatter for a profile
type ProfileMeta struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Files       []string `yaml:"files,omitempty"` // e.g., ["file:setup.sh > setup.sh", "file:config.json"]
}

// loadItem loads a single item from a file
func loadItem(path string, itemType string) (*Item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Get name from filename (strip .md for regular types, keep full name for raw types)
	name := filepath.Base(path)
	if !IsRawType(itemType) {
		name = strings.TrimSuffix(name, ".md")
	}

	item := &Item{
		Type:     itemType,
		Name:     name,
		FilePath: path,
	}

	// Raw types (file:) store content as-is without frontmatter
	if IsRawType(itemType) {
		item.Content = string(content)
		return item, nil
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		var meta ItemMeta
		if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		item.Description = meta.Description
	}

	item.Content = string(markdown)
	return item, nil
}

// loadProfiles loads all profiles from a directory
func (r *Registry) loadProfiles(dir string) ([]Profile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Profile{}, nil
		}
		return nil, err
	}

	var profiles []Profile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		profile, err := loadProfile(path)
		if err != nil {
			continue
		}
		profiles = append(profiles, *profile)
	}

	return profiles, nil
}

// loadProfile loads a single profile from a markdown file
func loadProfile(path string) (*Profile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	profile := &Profile{
		FilePath:   path,
		Name:       strings.TrimSuffix(filepath.Base(path), ".md"),
		RawContent: string(content),
	}

	// Extract frontmatter if present
	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		var meta ProfileMeta
		if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		profile.Description = meta.Description

		// Parse files field
		for _, fileSpec := range meta.Files {
			pf := parseProfileFile(fileSpec)
			if pf != nil {
				profile.Files = append(profile.Files, *pf)
			}
		}
	}

	// Store content below frontmatter
	profile.Content = string(markdown)

	return profile, nil
}

// parseProfileFile parses a file spec like "file:name" or "file:name > dest"
func parseProfileFile(spec string) *ProfileFile {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil
	}

	// Check for redirect syntax: "file:name > destination"
	if strings.Contains(spec, ">") {
		parts := strings.SplitN(spec, ">", 2)
		source := strings.TrimSpace(parts[0])
		dest := strings.TrimSpace(parts[1])

		// Extract file name from source (remove file: prefix if present)
		source = strings.TrimPrefix(source, "file:")

		return &ProfileFile{
			Source:      source,
			Destination: dest,
		}
	}

	// Simple syntax: "file:name" - destination is same as source name
	source := strings.TrimPrefix(spec, "file:")
	return &ProfileFile{
		Source:      source,
		Destination: source,
	}
}

// saveProfile saves a profile to a markdown file
func saveProfile(path string, profile Profile) error {
	// If Content already has frontmatter, use it as-is
	if profile.Content != "" && bytes.HasPrefix([]byte(profile.Content), []byte("---\n")) {
		if err := os.WriteFile(path, []byte(profile.Content), 0644); err != nil {
			return fmt.Errorf("failed to write profile file: %w", err)
		}
		return nil
	}

	// Create frontmatter from profile struct
	meta := ItemMeta{
		Name:        profile.Name,
		Description: profile.Description,
	}

	fullContent, err := marshalWithFrontmatter(meta, profile.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(path, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// extractFrontmatter extracts YAML frontmatter from markdown
func extractFrontmatter(content []byte) ([]byte, []byte, error) {
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, content, nil
	}

	lines := bytes.Split(content, []byte("\n"))
	if len(lines) < 3 {
		return nil, content, nil
	}

	var frontmatterEnd int
	for i := 1; i < len(lines); i++ {
		if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("---")) {
			frontmatterEnd = i
			break
		}
	}

	if frontmatterEnd == 0 {
		return nil, nil, fmt.Errorf("unclosed frontmatter")
	}

	frontmatter := bytes.Join(lines[1:frontmatterEnd], []byte("\n"))
	markdown := bytes.Join(lines[frontmatterEnd+1:], []byte("\n"))

	return frontmatter, bytes.TrimLeft(markdown, "\n\r"), nil
}

// marshalWithFrontmatter creates a markdown file with YAML frontmatter
func marshalWithFrontmatter(data interface{}, content string) (string, error) {
	frontmatter, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(frontmatter)
	buf.WriteString("---\n\n")
	buf.WriteString(content)

	return buf.String(), nil
}
