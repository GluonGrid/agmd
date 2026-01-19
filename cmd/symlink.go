package cmd

import (
	"fmt"
	"os"

	"agmd/internal/config"
	"agmd/internal/symlink"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	symlinkClaude   bool
	symlinkCursor   bool
	symlinkWindsurf bool
	symlinkCopilot  bool
	symlinkAider    bool
	symlinkAll      bool
)

var symlinkCmd = &cobra.Command{
	Use:   "symlink",
	Short: "Manage symlinks",
	Long:  `Manage symlinks to agent.md for different AI coding assistants.`,
}

var symlinkAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add symlinks",
	Long: `Create symlinks for AI coding assistants.

Examples:
  agmd symlink add --claude
  agmd symlink add --all
  agmd symlink add --claude --cursor`,
	RunE: runSymlinkAdd,
}

var symlinkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List symlinks",
	Long:  `List all symlinks and their status.`,
	RunE:  runSymlinkList,
}

var symlinkRemoveCmd = &cobra.Command{
	Use:   "remove [filename]",
	Short: "Remove a symlink",
	Long: `Remove a specific symlink.

Examples:
  agmd symlink remove CLAUDE.md
  agmd symlink remove .cursorrules`,
	Args: cobra.ExactArgs(1),
	RunE: runSymlinkRemove,
}

func init() {
	rootCmd.AddCommand(symlinkCmd)
	symlinkCmd.AddCommand(symlinkAddCmd)
	symlinkCmd.AddCommand(symlinkListCmd)
	symlinkCmd.AddCommand(symlinkRemoveCmd)

	// Add flags for symlink add
	symlinkAddCmd.Flags().BoolVar(&symlinkClaude, "claude", false, "Create CLAUDE.md symlink")
	symlinkAddCmd.Flags().BoolVar(&symlinkCursor, "cursor", false, "Create .cursorrules symlink")
	symlinkAddCmd.Flags().BoolVar(&symlinkWindsurf, "windsurf", false, "Create .windsurfrules symlink")
	symlinkAddCmd.Flags().BoolVar(&symlinkCopilot, "copilot", false, "Create .github/copilot-instructions.md symlink")
	symlinkAddCmd.Flags().BoolVar(&symlinkAider, "aider", false, "Create .aider.conf.yml symlink")
	symlinkAddCmd.Flags().BoolVar(&symlinkAll, "all", false, "Create symlinks for all tools")
}

func runSymlinkAdd(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	// Check if agent.md exists
	if _, err := os.Stat(config.AgentMdFilename); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist. Run 'agmd init' first", config.AgentMdFilename)
	}

	// Determine which tools to create symlinks for
	var toolsToCreate []config.ToolConfig

	if symlinkAll {
		toolsToCreate = config.AvailableTools()
	} else {
		if symlinkClaude {
			if tool := config.GetToolByName("claude"); tool != nil {
				toolsToCreate = append(toolsToCreate, *tool)
			}
		}
		if symlinkCursor {
			if tool := config.GetToolByName("cursor"); tool != nil {
				toolsToCreate = append(toolsToCreate, *tool)
			}
		}
		if symlinkWindsurf {
			if tool := config.GetToolByName("windsurf"); tool != nil {
				toolsToCreate = append(toolsToCreate, *tool)
			}
		}
		if symlinkCopilot {
			if tool := config.GetToolByName("copilot"); tool != nil {
				toolsToCreate = append(toolsToCreate, *tool)
			}
		}
		if symlinkAider {
			if tool := config.GetToolByName("aider"); tool != nil {
				toolsToCreate = append(toolsToCreate, *tool)
			}
		}
	}

	if len(toolsToCreate) == 0 {
		return fmt.Errorf("no tools specified. Use --claude, --cursor, etc., or --all")
	}

	// Create symlinks
	fmt.Printf("%s Creating symlinks...\n", blue("→"))
	manager := symlink.NewManager(config.AgentMdFilename)

	for _, tool := range toolsToCreate {
		if err := manager.Create(tool); err != nil {
			fmt.Printf("%s Failed to create %s: %v\n", yellow("⚠"), tool.Filename, err)
		} else {
			fmt.Printf("%s Created %s → %s\n", green("✓"), tool.Filename, config.AgentMdFilename)
		}
	}

	return nil
}

func runSymlinkList(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	manager := symlink.NewManager(config.AgentMdFilename)
	statuses := manager.List()

	fmt.Printf("%s Symlink Status:\n\n", cyan("ℹ"))

	for _, status := range statuses {
		if status.Exists {
			if status.IsValid {
				fmt.Printf("%s %s → %s\n", green("✓"), status.Tool.Filename, config.AgentMdFilename)
			} else {
				fmt.Printf("%s %s (exists but invalid: points to %s)\n", yellow("⚠"), status.Tool.Filename, status.Target)
			}
		} else {
			fmt.Printf("%s %s (not created)\n", red("✗"), status.Tool.Filename)
		}
	}

	return nil
}

func runSymlinkRemove(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	filename := args[0]

	fmt.Printf("%s Removing symlink %s...\n", blue("→"), filename)

	manager := symlink.NewManager(config.AgentMdFilename)
	if err := manager.Remove(filename); err != nil {
		return err
	}

	fmt.Printf("%s Removed %s\n", green("✓"), filename)
	return nil
}
