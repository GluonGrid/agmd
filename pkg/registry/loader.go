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

// loadItem loads a single item from a file
func loadItem(path string, itemType string) (*Item, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Get name from filename
	name := strings.TrimSuffix(filepath.Base(path), ".md")

	item := &Item{
		Type:     itemType,
		Name:     name,
		FilePath: path,
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
		FilePath: path,
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	// Extract frontmatter if present
	frontmatter, _, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		var meta ItemMeta
		if err := yaml.Unmarshal(frontmatter, &meta); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		profile.Description = meta.Description
	}

	// Store the full content (with frontmatter)
	profile.Content = string(content)

	return profile, nil
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
