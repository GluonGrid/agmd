package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"agmd/internal/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit agent.md in your default editor",
	Long: `Open agent.md in your default editor (specified by $EDITOR environment variable).
Falls back to vim if $EDITOR is not set.

Examples:
  agmd edit
  EDITOR=nano agmd edit`,
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Check if agent.md exists
	if _, err := os.Stat(config.AgentMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist. Run 'agmd init' first", config.AgentMdFilename)
	}

	// Get editor from environment, default to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	fmt.Printf("%s Opening %s in %s...\n", blue("→"), config.AgentMdFilename, editor)

	// Create command to open editor
	editorCmd := exec.Command(editor, config.AgentMdFilename)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	// Run editor
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("%s Editing complete\n", green("✓"))
	return nil
}
