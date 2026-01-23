package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// New creates a new Registry instance
func New() (*Registry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".agmd")
	return &Registry{
		BasePath: basePath,
	}, nil
}

// GetBasePath returns the base path of the registry
func (r *Registry) GetBasePath() string {
	return r.BasePath
}

// Paths returns all registry subdirectory paths
func (r *Registry) Paths() RegistryPaths {
	return RegistryPaths{
		Base:       r.BasePath,
		Shared:     filepath.Join(r.BasePath, "shared"),
		Rules:      filepath.Join(r.BasePath, "rules"),
		Workflows:  filepath.Join(r.BasePath, "workflows"),
		Guidelines: filepath.Join(r.BasePath, "guidelines"),
		Profiles:   filepath.Join(r.BasePath, "profiles"),
	}
}

// Exists checks if the registry exists
func (r *Registry) Exists() bool {
	info, err := os.Stat(r.BasePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Initialize creates the registry directory structure
func (r *Registry) Initialize() error {
	paths := r.Paths()

	// Create all directories
	dirs := []string{
		paths.Base,
		paths.Shared,
		paths.Rules,
		paths.Workflows,
		paths.Guidelines,
		paths.Profiles,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// ListRules returns all rules in the registry
func (r *Registry) ListRules() ([]Rule, error) {
	paths := r.Paths()
	return r.loadRules(paths.Rules)
}

// GetRule retrieves a specific rule by name
func (r *Registry) GetRule(name string) (*Rule, error) {
	paths := r.Paths()
	path := filepath.Join(paths.Rules, name+".md")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("rule '%s' not found", name)
	}

	return loadRule(path)
}

// SaveRule saves a rule to the registry
func (r *Registry) SaveRule(rule Rule) error {
	paths := r.Paths()
	path := filepath.Join(paths.Rules, rule.Name+".md")

	content, err := marshalWithFrontmatter(rule, rule.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal rule: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write rule file: %w", err)
	}

	return nil
}

// DeleteRule removes a rule from the registry
func (r *Registry) DeleteRule(name string) error {
	paths := r.Paths()
	path := filepath.Join(paths.Rules, name+".md")

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	return nil
}

// ListWorkflows returns all workflows in the registry
func (r *Registry) ListWorkflows() ([]Workflow, error) {
	paths := r.Paths()
	return r.loadWorkflows(paths.Workflows)
}

// GetWorkflow retrieves a specific workflow by name
func (r *Registry) GetWorkflow(name string) (*Workflow, error) {
	paths := r.Paths()
	path := filepath.Join(paths.Workflows, name+".md")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("workflow '%s' not found", name)
	}

	return loadWorkflow(path)
}

// SaveWorkflow saves a workflow to the registry
func (r *Registry) SaveWorkflow(workflow Workflow) error {
	paths := r.Paths()
	path := filepath.Join(paths.Workflows, workflow.Name+".md")

	content, err := marshalWithFrontmatter(workflow, workflow.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write workflow file: %w", err)
	}

	return nil
}

// ListGuidelines returns all guidelines in the registry
func (r *Registry) ListGuidelines() ([]Guideline, error) {
	paths := r.Paths()
	return r.loadGuidelines(paths.Guidelines)
}

// GetGuideline retrieves a specific guideline by name
func (r *Registry) GetGuideline(name string) (*Guideline, error) {
	paths := r.Paths()
	path := filepath.Join(paths.Guidelines, name+".md")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("guideline '%s' not found", name)
	}

	return loadGuideline(path)
}

// SaveGuideline saves a guideline to the registry
func (r *Registry) SaveGuideline(guideline Guideline) error {
	paths := r.Paths()
	path := filepath.Join(paths.Guidelines, guideline.Name+".md")

	content, err := marshalWithFrontmatter(guideline, guideline.Content)
	if err != nil {
		return fmt.Errorf("failed to marshal guideline: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write guideline file: %w", err)
	}

	return nil
}

// ListProfiles returns all profiles in the registry
func (r *Registry) ListProfiles() ([]Profile, error) {
	paths := r.Paths()
	return r.loadProfiles(paths.Profiles)
}

// GetProfile retrieves a specific profile by name
func (r *Registry) GetProfile(name string) (*Profile, error) {
	paths := r.Paths()
	path := filepath.Join(paths.Profiles, name+".yaml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return loadProfile(path)
}

// SaveProfile saves a profile to the registry
func (r *Registry) SaveProfile(profile Profile) error {
	paths := r.Paths()
	path := filepath.Join(paths.Profiles, profile.Name+".yaml")

	return saveProfile(path, profile)
}
