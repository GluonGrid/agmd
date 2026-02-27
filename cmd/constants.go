package cmd

import (
	"os"
	"strings"
)

// File name constants used across commands
const (
	directivesMdFilename      = "directives.md"       // Source file with directives (committed)
	localDirectivesMdFilename = "directives.local.md" // Personal overrides (gitignored)
	agentsMdFilename          = "AGENTS.md"           // Generated output for AI agents
)

// tildeHome replaces the user's home directory prefix with ~ for display.
func tildeHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
