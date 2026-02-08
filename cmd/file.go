package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Shared flags for file subcommands
var fileForce bool
var fileContent string
var fileNoEditor bool

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage raw files in the registry",
	Long: `Manage raw files stored in the registry.

Files are stored as-is without .md extension or frontmatter.
Use this for scripts, config files, or any non-markdown content.

Subcommands:
  list        List all files
  new         Add a new file
  show        Show file content
  delete      Delete a file

Examples:
  agmd file list                                  # List all files
  agmd file new setup.sh --content "#!/bin/bash" # Create file
  agmd file show setup.sh                         # Show file content
  agmd file delete setup.sh                       # Delete file`,
}

var fileListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all files in the registry",
	Long: `List all files stored in the registry.

Examples:
  agmd file list    # List all files
  agmd file ls      # Same (alias)`,
	RunE: runFileList,
}

var fileNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Add a new file to the registry",
	Long: `Add a new file to the registry.

Files are stored as-is without frontmatter.

Examples:
  agmd file new setup.sh --content "#!/bin/bash\necho hello"
  agmd file new config.json --content '{"key": "value"}'
  echo "content" | agmd file new myfile.txt
  agmd file new script.py    # Opens editor`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeFileName,
	RunE:              runFileNew,
}

var fileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show file content",
	Long: `Display the content of a file.

Examples:
  agmd file show setup.sh     # Show file content
  agmd file show config.json  # Show config file`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeFileNameOnly,
	RunE:              runFileShow,
}

var fileDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a file from the registry",
	Long: `Delete a file from the registry.

Examples:
  agmd file delete setup.sh         # Delete with confirmation
  agmd file rm setup.sh             # Same (alias)
  agmd file delete setup.sh --force # Skip confirmation`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeFileNameOnly,
	RunE:              runFileDelete,
}

func init() {
	rootCmd.AddCommand(fileCmd)
	fileCmd.AddCommand(fileListCmd)
	fileCmd.AddCommand(fileNewCmd)
	fileCmd.AddCommand(fileShowCmd)
	fileCmd.AddCommand(fileDeleteCmd)

	fileNewCmd.Flags().StringVar(&fileContent, "content", "", "File content")
	fileNewCmd.Flags().BoolVar(&fileNoEditor, "no-editor", false, "Don't open editor after creating")

	fileDeleteCmd.Flags().BoolVarP(&fileForce, "force", "f", false, "Skip confirmation prompt")
}

// completeFileNameOnly provides completion for file names (without file: prefix)
func completeFileNameOnly(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	fileDir := filepath.Join(reg.BasePath, "file")
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), toComplete) {
			completions = append(completions, entry.Name())
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func runFileList(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	fileDir := filepath.Join(reg.BasePath, "file")
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s No files in registry\n", yellow("!"))
			fmt.Println("\nCreate a file:")
			fmt.Println("  agmd file new script.sh --content '#!/bin/bash'")
			return nil
		}
		return fmt.Errorf("failed to read file directory: %w", err)
	}

	// Filter out directories
	var files []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry)
		}
	}

	if len(files) == 0 {
		fmt.Printf("%s No files in registry\n", yellow("!"))
		fmt.Println("\nCreate a file:")
		fmt.Println("  agmd file new script.sh --content '#!/bin/bash'")
		return nil
	}

	fmt.Printf("%s\n\n", cyan(fileDir))
	fmt.Printf("file/ (%d)\n", len(files))
	for _, entry := range files {
		info, err := entry.Info()
		if err != nil {
			fmt.Printf("  %s\n", entry.Name())
		} else {
			fmt.Printf("  %s (%d bytes)\n", entry.Name(), info.Size())
		}
	}

	return nil
}

func runFileNew(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Build path
	fileDir := filepath.Join(reg.BasePath, "file")
	filePath := filepath.Join(fileDir, name)

	// Check if exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file:%s already exists", name)
	}

	// Create directory
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return fmt.Errorf("failed to create file directory: %w", err)
	}

	// Determine content source
	var content string
	if fileContent != "" {
		content = fileContent
	} else if !isTerminal(os.Stdin) {
		stdinContent, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(stdinContent)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	fmt.Printf("%s Created file:%s\n", green("ok"), name)

	// Open editor unless --no-editor or content was provided
	if fileNoEditor || content != "" {
		fmt.Printf("%s %s\n", blue("->"), filePath)
		return nil
	}

	fmt.Printf("%s Opening editor...\n", blue("->"))
	return openInEditor(filePath)
}

func runFileShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	filePath := filepath.Join(reg.BasePath, "file", name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file:%s not found", name)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fmt.Print(string(content))
	return nil
}

func runFileDelete(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	filePath := filepath.Join(reg.BasePath, "file", name)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file:%s not found", name)
	}

	// Show what will be deleted
	fmt.Printf("%s Deleting file:%s\n", blue("→"), name)
	fmt.Printf("  Path: %s\n", filePath)

	// Confirmation prompt (unless --force)
	if !fileForce {
		fmt.Printf("\n%s This will permanently delete this file.\n", yellow("⚠"))
		fmt.Print("\nAre you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("\nCancelled.")
			return nil
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fmt.Printf("%s Deleted file:%s\n", green("✓"), name)

	// Check if file directory is empty
	fileDir := filepath.Join(reg.BasePath, "file")
	entries, err := os.ReadDir(fileDir)
	if err == nil && len(entries) == 0 {
		os.Remove(fileDir)
	}

	return nil
}
