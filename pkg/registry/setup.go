package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// Setup creates the ~/.agmd directory structure
func (r *Registry) Setup(force bool) error {
	// Check if already exists
	if r.Exists() && !force {
		return fmt.Errorf("registry already exists at %s (use --force to reinitialize)", r.BasePath)
	}

	// Create base directory
	if err := os.MkdirAll(r.BasePath, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	// Create default profile
	if err := r.createDefaultProfile(); err != nil {
		return fmt.Errorf("failed to create default profile: %w", err)
	}

	// Create default agmd guide
	if err := r.createDefaultGuide(); err != nil {
		return fmt.Errorf("failed to create default guide: %w", err)
	}

	return nil
}

// createDefaultProfile creates the default.md profile
func (r *Registry) createDefaultProfile() error {
	// Ensure profile directory exists
	profileDir := filepath.Join(r.BasePath, "profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}

	profile := Profile{
		Name:        "default",
		Description: "Default directives.md template",
		Content:     GetDefaultProfileTemplate(),
	}

	return r.SaveProfile(profile)
}

// createDefaultGuide creates the guide/agmd.md file
func (r *Registry) createDefaultGuide() error {
	// Ensure guide directory exists
	guideDir := filepath.Join(r.BasePath, "guide")
	if err := os.MkdirAll(guideDir, 0755); err != nil {
		return err
	}

	guidePath := filepath.Join(guideDir, "agmd.md")
	return os.WriteFile(guidePath, []byte(GetAgmdGuideTemplate()), 0644)
}
