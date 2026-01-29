package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agmd",
	Short: "Manage AI agent instructions across all your projects",
	Long: `agmd - Agent Markdown

Stop copy-pasting AI instructions between projects. agmd lets you maintain
a personal registry of rules, workflows, and guidelines that you can mix
and match across any project.

How it works:
  1. Store reusable instructions in ~/.agmd/ (your personal registry)
  2. Reference them in directives.md with simple directives
  3. Run 'agmd sync' to generate AGENTS.md for AI agents

Example directives.md:
  :::include rule:typescript
  :::include workflow:commit
  :::list guideline
  code-style
  documentation
  :::end

Benefits:
  • One source of truth for your coding standards
  • Update once, sync everywhere
  • Share rules across teams via profiles
  • Works with Claude, Cursor, Windsurf, and any AI coding assistant

Get started:
  agmd setup    # Initialize your registry
  agmd init     # Create directives.md in your project
  agmd sync     # Generate AGENTS.md`,
	Version: "0.1.0",
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
