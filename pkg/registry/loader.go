package registry

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadRules loads all rules from a directory
func (r *Registry) loadRules(dir string) ([]Rule, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Rule{}, nil
		}
		return nil, err
	}

	var rules []Rule
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		rule, err := loadRule(path)
		if err != nil {
			// Skip invalid files
			continue
		}
		rules = append(rules, *rule)
	}

	return rules, nil
}

// loadRule loads a single rule from a file
func loadRule(path string) (*Rule, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rule := &Rule{
		FilePath: path,
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, rule); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
	}

	rule.Content = string(markdown)
	return rule, nil
}

// loadWorkflows loads all workflows from a directory
func (r *Registry) loadWorkflows(dir string) ([]Workflow, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Workflow{}, nil
		}
		return nil, err
	}

	var workflows []Workflow
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		workflow, err := loadWorkflow(path)
		if err != nil {
			continue
		}
		workflows = append(workflows, *workflow)
	}

	return workflows, nil
}

// loadWorkflow loads a single workflow from a file
func loadWorkflow(path string) (*Workflow, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	workflow := &Workflow{
		FilePath: path,
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, workflow); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
	}

	workflow.Content = string(markdown)
	return workflow, nil
}

// loadGuidelines loads all guidelines from a directory
func (r *Registry) loadGuidelines(dir string) ([]Guideline, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Guideline{}, nil
		}
		return nil, err
	}

	var guidelines []Guideline
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		guideline, err := loadGuideline(path)
		if err != nil {
			continue
		}
		guidelines = append(guidelines, *guideline)
	}

	return guidelines, nil
}

// loadGuideline loads a single guideline from a file
func loadGuideline(path string) (*Guideline, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	guideline := &Guideline{
		FilePath: path,
		Name:     strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, guideline); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
	}

	guideline.Content = string(markdown)
	return guideline, nil
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
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

// loadProfile loads a single profile from a YAML file
func loadProfile(path string) (*Profile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var profile Profile
	if err := yaml.Unmarshal(content, &profile); err != nil {
		return nil, fmt.Errorf("invalid profile YAML: %w", err)
	}

	profile.FilePath = path
	if profile.Name == "" {
		profile.Name = strings.TrimSuffix(filepath.Base(path), ".yaml")
	}

	return &profile, nil
}

// saveProfile saves a profile to a YAML file
func saveProfile(path string, profile Profile) error {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	return nil
}

// extractFrontmatter extracts YAML frontmatter from markdown
func extractFrontmatter(content []byte) ([]byte, []byte, error) {
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		// No frontmatter
		return nil, content, nil
	}

	// Find closing ---
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

	// Extract frontmatter (skip first and last ---)
	frontmatter := bytes.Join(lines[1:frontmatterEnd], []byte("\n"))

	// Extract markdown (everything after closing ---)
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
