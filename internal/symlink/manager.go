package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"agmd/internal/config"
)

// Manager handles symlink operations
type Manager struct {
	sourceFile string
}

// NewManager creates a new symlink manager
func NewManager(sourceFile string) *Manager {
	return &Manager{
		sourceFile: sourceFile,
	}
}

// Create creates a symlink for the specified tool
func (m *Manager) Create(tool config.ToolConfig) error {
	// If the tool needs a directory, create it
	if tool.NeedsDir {
		dir := filepath.Dir(tool.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Check if target already exists
	if _, err := os.Lstat(tool.Filename); err == nil {
		// File exists, check if it's already a symlink to agent.md
		target, err := os.Readlink(tool.Filename)
		if err == nil && target == m.sourceFile {
			// Already a valid symlink
			return nil
		}

		// File exists but is not our symlink
		return fmt.Errorf("file %s already exists (not a symlink to %s)", tool.Filename, m.sourceFile)
	}

	// Create the symlink
	if err := os.Symlink(m.sourceFile, tool.Filename); err != nil {
		return fmt.Errorf("failed to create symlink %s: %w", tool.Filename, err)
	}

	return nil
}

// Remove removes a symlink
func (m *Manager) Remove(filename string) error {
	// Check if it's a symlink
	info, err := os.Lstat(filename)
	if err != nil {
		return fmt.Errorf("file %s not found: %w", filename, err)
	}

	// Verify it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("%s is not a symlink", filename)
	}

	// Verify it points to agent.md
	target, err := os.Readlink(filename)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", filename, err)
	}

	if target != m.sourceFile {
		return fmt.Errorf("%s is a symlink but doesn't point to %s (points to %s)", filename, m.sourceFile, target)
	}

	// Remove the symlink
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to remove symlink %s: %w", filename, err)
	}

	return nil
}

// List returns all existing symlinks and their status
func (m *Manager) List() []SymlinkStatus {
	var statuses []SymlinkStatus

	for _, tool := range config.AvailableTools() {
		status := SymlinkStatus{
			Tool:     tool,
			Exists:   false,
			IsValid:  false,
			Target:   "",
		}

		info, err := os.Lstat(tool.Filename)
		if err == nil {
			status.Exists = true

			// Check if it's a symlink
			if info.Mode()&os.ModeSymlink != 0 {
				target, err := os.Readlink(tool.Filename)
				if err == nil {
					status.Target = target
					status.IsValid = (target == m.sourceFile)
				}
			}
		}

		statuses = append(statuses, status)
	}

	return statuses
}

// SymlinkStatus represents the status of a symlink
type SymlinkStatus struct {
	Tool    config.ToolConfig
	Exists  bool
	IsValid bool // true if it's a valid symlink to agent.md
	Target  string
}

// Verify checks if agent.md exists
func (m *Manager) Verify() error {
	if _, err := os.Stat(m.sourceFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", m.sourceFile)
		}
		return fmt.Errorf("failed to check %s: %w", m.sourceFile, err)
	}
	return nil
}
