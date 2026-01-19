package state

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// ProjectState represents the agents.toml configuration file
type ProjectState struct {
	Rules      []string `toml:"rules"`
	Workflows  []string `toml:"workflows"`
	Guidelines []string `toml:"guidelines"`
	Profiles   []string `toml:"profiles"`
}

// DefaultState returns a new ProjectState with sensible defaults
func DefaultState() *ProjectState {
	return &ProjectState{
		Rules:      []string{"no-modify-agmd-sections"},
		Workflows:  []string{},
		Guidelines: []string{},
		Profiles:   []string{},
	}
}

// Load reads and parses an agents.toml file
func Load(path string) (*ProjectState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default state if file doesn't exist
			return DefaultState(), nil
		}
		return nil, fmt.Errorf("failed to read agents.toml: %w", err)
	}

	var state ProjectState
	if err := toml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse agents.toml: %w", err)
	}

	return &state, nil
}

// Save writes the state to an agents.toml file
func (s *ProjectState) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create agents.toml: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(s); err != nil {
		return fmt.Errorf("failed to encode agents.toml: %w", err)
	}

	return nil
}

// AddRule adds a rule to the rules list
func (s *ProjectState) AddRule(name string) {
	// Check if already exists
	for _, rule := range s.Rules {
		if rule == name {
			return // Already exists
		}
	}
	s.Rules = append(s.Rules, name)
}

// RemoveRule removes a rule from the rules list
func (s *ProjectState) RemoveRule(name string) {
	var filtered []string
	for _, rule := range s.Rules {
		if rule != name {
			filtered = append(filtered, rule)
		}
	}
	s.Rules = filtered
}

// AddWorkflow adds a workflow to the workflows list
func (s *ProjectState) AddWorkflow(name string) {
	for _, workflow := range s.Workflows {
		if workflow == name {
			return
		}
	}
	s.Workflows = append(s.Workflows, name)
}

// RemoveWorkflow removes a workflow from the workflows list
func (s *ProjectState) RemoveWorkflow(name string) {
	var filtered []string
	for _, workflow := range s.Workflows {
		if workflow != name {
			filtered = append(filtered, workflow)
		}
	}
	s.Workflows = filtered
}

// AddGuideline adds a guideline to the guidelines list
func (s *ProjectState) AddGuideline(name string) {
	for _, guideline := range s.Guidelines {
		if guideline == name {
			return
		}
	}
	s.Guidelines = append(s.Guidelines, name)
}

// RemoveGuideline removes a guideline from the guidelines list
func (s *ProjectState) RemoveGuideline(name string) {
	var filtered []string
	for _, guideline := range s.Guidelines {
		if guideline != name {
			filtered = append(filtered, guideline)
		}
	}
	s.Guidelines = filtered
}

// AddProfile adds a profile to the profiles list
func (s *ProjectState) AddProfile(name string) {
	for _, profile := range s.Profiles {
		if profile == name {
			return
		}
	}
	s.Profiles = append(s.Profiles, name)
}

// RemoveProfile removes a profile from the profiles list
func (s *ProjectState) RemoveProfile(name string) {
	var filtered []string
	for _, profile := range s.Profiles {
		if profile != name {
			filtered = append(filtered, profile)
		}
	}
	s.Profiles = filtered
}

// HasRule checks if a rule is active
func (s *ProjectState) HasRule(name string) bool {
	for _, rule := range s.Rules {
		if rule == name {
			return true
		}
	}
	return false
}

// HasWorkflow checks if a workflow is active
func (s *ProjectState) HasWorkflow(name string) bool {
	for _, workflow := range s.Workflows {
		if workflow == name {
			return true
		}
	}
	return false
}

// HasGuideline checks if a guideline is active
func (s *ProjectState) HasGuideline(name string) bool {
	for _, guideline := range s.Guidelines {
		if guideline == name {
			return true
		}
	}
	return false
}

// HasProfile checks if a profile is active
func (s *ProjectState) HasProfile(name string) bool {
	for _, profile := range s.Profiles {
		if profile == name {
			return true
		}
	}
	return false
}
