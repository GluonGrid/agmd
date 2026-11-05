package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

// Registry locates the agmd data directory that holds task files.
type Registry struct {
	BasePath string // ~/.agmd, or $AGMD_HOME when set
}

// New resolves the data directory. It defaults to ~/.agmd and can be
// overridden with AGMD_HOME. The directory is created lazily by writers.
func New() (*Registry, error) {
	if override := os.Getenv("AGMD_HOME"); override != "" {
		return &Registry{BasePath: override}, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &Registry{BasePath: filepath.Join(homeDir, ".agmd")}, nil
}
