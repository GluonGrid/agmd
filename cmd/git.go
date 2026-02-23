package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"agmd/pkg/registry"

	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:                "git [git args...]",
	Short:              "Run a git command in ~/.agmd",
	Long: `Forward any git command to the ~/.agmd registry directory.

Useful for syncing your global registry across devices without
having to cd into ~/.agmd first.

Examples:
  agmd git init
  agmd git remote add origin git@github.com:you/agmd-registry.git
  agmd git add -A
  agmd git commit -m "add typescript rule"
  agmd git push
  agmd git pull
  agmd git log --oneline
  agmd git status`,
	DisableFlagParsing: true, // pass all flags/args straight to git
	RunE:               runGitCmd,
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func runGitCmd(cmd *cobra.Command, args []string) error {
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}
	if !reg.Exists() {
		return fmt.Errorf("registry not found — run 'agmd setup' first")
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH")
	}

	c := exec.Command(gitPath, args...)
	c.Dir = reg.BasePath
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
