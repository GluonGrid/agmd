package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agmd",
	Short: "Manage AI agent configuration files",
	Long: `agmd is a CLI tool for managing AI agent configuration files.

It helps you maintain a single source of truth (agent.md) and automatically
creates symlinks for different AI coding assistants (Claude, Cursor, Windsurf, etc.).`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
