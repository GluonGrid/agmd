package registry

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// loadRules loads all rules from a directory (recursively scans subdirectories)
func (r *Registry) loadRules(dir string) ([]Rule, error) {
	return r.loadRulesRecursive(dir, dir)
}

// loadRulesRecursive recursively loads rules from a directory and subdirectories
func (r *Registry) loadRulesRecursive(baseDir, currentDir string) ([]Rule, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Rule{}, nil
		}
		return nil, err
	}

	var rules []Rule
	for _, entry := range entries {
		path := filepath.Join(currentDir, entry.Name())

		if entry.IsDir() {
			// Recursively scan subdirectories
			subRules, err := r.loadRulesRecursive(baseDir, path)
			if err != nil {
				continue
			}
			rules = append(rules, subRules...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		rule, err := loadRuleWithBase(path, baseDir)
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
	return loadRuleWithBase(path, filepath.Dir(path))
}

// loadRuleWithBase loads a rule and calculates name relative to baseDir
func loadRuleWithBase(path string, baseDir string) (*Rule, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Calculate relative path from baseDir for the name (e.g., "auth/custom-auth")
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		relPath = filepath.Base(path)
	}
	name := strings.TrimSuffix(relPath, ".md")

	rule := &Rule{
		FilePath: path,
		Name:     name,
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, rule); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		// Restore the calculated name (don't let frontmatter override it)
		rule.Name = name
	}

	rule.Content = string(markdown)
	return rule, nil
}

// loadWorkflows loads all workflows from a directory (recursively scans subdirectories)
func (r *Registry) loadWorkflows(dir string) ([]Workflow, error) {
	return r.loadWorkflowsRecursive(dir, dir)
}

// loadWorkflowsRecursive recursively loads workflows from a directory and subdirectories
func (r *Registry) loadWorkflowsRecursive(baseDir, currentDir string) ([]Workflow, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Workflow{}, nil
		}
		return nil, err
	}

	var workflows []Workflow
	for _, entry := range entries {
		path := filepath.Join(currentDir, entry.Name())

		if entry.IsDir() {
			subWorkflows, err := r.loadWorkflowsRecursive(baseDir, path)
			if err != nil {
				continue
			}
			workflows = append(workflows, subWorkflows...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		workflow, err := loadWorkflowWithBase(path, baseDir)
		if err != nil {
			continue
		}
		workflows = append(workflows, *workflow)
	}

	return workflows, nil
}

// loadWorkflow loads a single workflow from a file
func loadWorkflow(path string) (*Workflow, error) {
	return loadWorkflowWithBase(path, filepath.Dir(path))
}

// loadWorkflowWithBase loads a workflow and calculates name relative to baseDir
func loadWorkflowWithBase(path string, baseDir string) (*Workflow, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		relPath = filepath.Base(path)
	}
	name := strings.TrimSuffix(relPath, ".md")

	workflow := &Workflow{
		FilePath: path,
		Name:     name,
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, workflow); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		// Restore the calculated name (don't let frontmatter override it)
		workflow.Name = name
	}

	workflow.Content = string(markdown)
	return workflow, nil
}

// loadGuidelines loads all guidelines from a directory (recursively scans subdirectories)
func (r *Registry) loadGuidelines(dir string) ([]Guideline, error) {
	return r.loadGuidelinesRecursive(dir, dir)
}

// loadGuidelinesRecursive recursively loads guidelines from a directory and subdirectories
func (r *Registry) loadGuidelinesRecursive(baseDir, currentDir string) ([]Guideline, error) {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Guideline{}, nil
		}
		return nil, err
	}

	var guidelines []Guideline
	for _, entry := range entries {
		path := filepath.Join(currentDir, entry.Name())

		if entry.IsDir() {
			subGuidelines, err := r.loadGuidelinesRecursive(baseDir, path)
			if err != nil {
				continue
			}
			guidelines = append(guidelines, subGuidelines...)
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		guideline, err := loadGuidelineWithBase(path, baseDir)
		if err != nil {
			continue
		}
		guidelines = append(guidelines, *guideline)
	}

	return guidelines, nil
}

// loadGuideline loads a single guideline from a file
func loadGuideline(path string) (*Guideline, error) {
	return loadGuidelineWithBase(path, filepath.Dir(path))
}

// loadGuidelineWithBase loads a guideline and calculates name relative to baseDir
func loadGuidelineWithBase(path string, baseDir string) (*Guideline, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		relPath = filepath.Base(path)
	}
	name := strings.TrimSuffix(relPath, ".md")

	guideline := &Guideline{
		FilePath: path,
		Name:     name,
	}

	frontmatter, markdown, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	if len(frontmatter) > 0 {
		if err := yaml.Unmarshal(frontmatter, guideline); err != nil {
			return nil, fmt.Errorf("invalid frontmatter: %w", err)
		}
		// Restore the calculated name (don't let frontmatter override it)
		guideline.Name = name
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
