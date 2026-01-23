package registry

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:assets/defaults
var defaultAssets embed.FS

// Setup creates the ~/.agmd directory structure and copies default files
func (r *Registry) Setup(force bool) error {
	// Check if already exists
	if r.Exists() && !force {
		return fmt.Errorf("registry already exists at %s (use --force to reinitialize)", r.BasePath)
	}

	// Create directory structure
	if err := r.Initialize(); err != nil {
		return err
	}

	// Copy default files
	if err := r.copyDefaults(); err != nil {
		return fmt.Errorf("failed to copy default files: %w", err)
	}

	// Create default profile (used by agmd init)
	if err := r.createDefaultProfile(); err != nil {
		return fmt.Errorf("failed to create default profile: %w", err)
	}

	return nil
}

// createDefaultProfile creates the default.md profile
func (r *Registry) createDefaultProfile() error {
	profile := Profile{
		Name:        "default",
		Description: "Default directives.md template with basic structure",
		Content:     GetDefaultProfileTemplate(),
	}

	return r.SaveProfile(profile)
}

// copyDefaults copies default files from embedded assets to registry
func (r *Registry) copyDefaults() error {
	// Walk through embedded defaults
	return fs.WalkDir(defaultAssets, "assets/defaults", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root
		if path == "assets/defaults" {
			return nil
		}

		// Get relative path by removing "assets/defaults/" prefix
		relPath := path[16:] // len("assets/defaults/") = 16

		destPath := filepath.Join(r.BasePath, relPath)

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(destPath, 0755)
		}

		// Copy file
		content, err := defaultAssets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})
}
