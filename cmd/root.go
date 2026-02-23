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
  :::use rule:typescript
  :::use workflow:commit
  :::list
  guideline:code-style
  guideline:documentation
  :::end
  :::skills                  # auto-detect .claude/skills or .agents/skills
  :::docs                    # list symlinked documentation in ./docs/

Benefits:
  • One source of truth for your coding standards
  • Update once, sync everywhere
  • Share rules across teams via profiles
  • Works with Claude, Cursor, Windsurf, and any AI coding assistant

Get started:
  agmd setup    # Initialize your registry
  agmd init     # Create directives.md in your project
  agmd sync     # Generate AGENTS.md

Special subcommands:
  agmd task     # Manage project tasks with dependencies
  agmd file     # Manage raw files (scripts, configs)
  agmd doc      # Manage documentation folders
  agmd skill    # Manage Agent Skills (install, link, list)

Agent Skills:
  agmd skill install owner/repo   # install from GitHub
  agmd skill link my-skill        # symlink into agent dirs
  agmd skill list                 # list installed skills

Sync across devices:
  agmd git init
  agmd git remote add origin <url>
  agmd git add -A && agmd git commit -m "sync" && agmd git push
  git clone <url> ~/.agmd   # on a new machine`,
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
	// Enable shell completion command (agmd completion bash/zsh/fish/powershell)
}
