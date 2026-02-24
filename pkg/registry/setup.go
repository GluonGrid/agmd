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

	// Create bundled managing-agmd skill
	if err := r.createDefaultManagingSkill(); err != nil {
		return fmt.Errorf("failed to create managing-agmd skill: %w", err)
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
	guideDir := filepath.Join(r.BasePath, "guide")
	if err := os.MkdirAll(guideDir, 0755); err != nil {
		return err
	}
	guidePath := filepath.Join(guideDir, "agmd.md")
	return os.WriteFile(guidePath, []byte(GetAgmdGuideTemplate()), 0644)
}

// createDefaultManagingSkill creates the bundled managing-agmd skill with full structure
func (r *Registry) createDefaultManagingSkill() error {
	skillDir := filepath.Join(r.BasePath, "skill", "managing-agmd")

	// Create all subdirectories
	for _, sub := range []string{"", "references", "assets", "scripts"} {
		if err := os.MkdirAll(filepath.Join(skillDir, sub), 0755); err != nil {
			return err
		}
	}

	// Write each file
	files := map[string]string{
		"SKILL.md":                    GetManagingAgmdSkillTemplate(),
		"references/migration.md":     GetManagingAgmdMigrationRef(),
		"references/tasks.md":         GetManagingAgmdTasksRef(),
		"references/directives.md":    GetManagingAgmdDirectivesRef(),
		"scripts/check-setup.sh":      GetManagingAgmdCheckScript(),
	}

	for rel, content := range files {
		path := filepath.Join(skillDir, rel)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", rel, err)
		}
	}

	// Make script executable
	scriptPath := filepath.Join(skillDir, "scripts", "check-setup.sh")
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return err
	}

	// Write icon
	if icon := GetManagingAgmdIcon(); icon != nil {
		iconPath := filepath.Join(skillDir, "assets", "icon.png")
		if err := os.WriteFile(iconPath, icon, 0644); err != nil {
			return fmt.Errorf("failed to write icon: %w", err)
		}
	}

	return nil
}
