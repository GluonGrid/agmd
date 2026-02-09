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

// Shared flags for doc subcommands
var docForce bool
var docGitignore bool

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Manage documentation folders in the registry",
	Long: `Manage documentation folders stored in the registry.

Docs are stored as folders containing documentation files.
Use this for framework docs, API references, or any multi-file documentation.

Unlike files (which are copied), docs are symlinked into projects.

Subcommands:
  list        List all docs
  add         Add a documentation folder to registry
  show        Show doc folder contents
  delete      Delete a doc from registry
  link        Symlink a doc into current project
  unlink      Remove a doc symlink from project

Examples:
  agmd doc list                              # List all docs
  agmd doc add ./docs svelte-kit             # Add docs folder as svelte-kit
  agmd doc show svelte-kit                   # Show doc contents
  agmd doc link svelte-kit                   # Symlink into project as docs/svelte-kit
  agmd doc link svelte-kit --gitignore       # Also add to .gitignore
  agmd doc unlink svelte-kit                 # Remove symlink
  agmd doc delete svelte-kit                 # Delete from registry`,
}

var docListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all docs in the registry",
	Long: `List all documentation folders stored in the registry.

Examples:
  agmd doc list    # List all docs
  agmd doc ls      # Same (alias)`,
	RunE: runDocList,
}

var docAddCmd = &cobra.Command{
	Use:   "add <path> <name>",
	Short: "Add a documentation folder to the registry",
	Long: `Copy a documentation folder into the registry.

Examples:
  agmd doc add ./docs svelte-kit           # Add docs folder as svelte-kit
  agmd doc add ./node_modules/lib/docs api # Add as api
  agmd doc add ./docs svelte-kit --force   # Overwrite if exists`,
	Args: cobra.ExactArgs(2),
	RunE: runDocAdd,
}

var docShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show doc folder contents",
	Long: `Display the contents of a documentation folder.

Examples:
  agmd doc show svelte-kit    # List files in svelte-kit docs`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDocNameOnly,
	RunE:              runDocShow,
}

var docDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a doc from the registry",
	Long: `Delete a documentation folder from the registry.

Examples:
  agmd doc delete svelte-kit         # Delete with confirmation
  agmd doc rm svelte-kit             # Same (alias)
  agmd doc delete svelte-kit --force # Skip confirmation`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDocNameOnly,
	RunE:              runDocDelete,
}

var docLinkCmd = &cobra.Command{
	Use:   "link <name> [dest]",
	Short: "Symlink a doc into current project",
	Long: `Create a symlink to a documentation folder in the current project.

By default, creates symlink at docs/<name>. Optionally specify a custom destination.

Examples:
  agmd doc link svelte-kit                  # Creates docs/svelte-kit -> ~/.agmd/doc/svelte-kit
  agmd doc link svelte-kit ./reference      # Creates ./reference -> ~/.agmd/doc/svelte-kit
  agmd doc link svelte-kit --gitignore      # Also add to .gitignore`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: completeDocNameOnly,
	RunE:              runDocLink,
}

var docUnlinkCmd = &cobra.Command{
	Use:   "unlink <name>",
	Short: "Remove a doc symlink from project",
	Long: `Remove a documentation symlink from the current project.

Examples:
  agmd doc unlink svelte-kit    # Removes docs/svelte-kit symlink`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: completeDocNameOnly,
	RunE:              runDocUnlink,
}

func init() {
	rootCmd.AddCommand(docCmd)
	docCmd.AddCommand(docListCmd)
	docCmd.AddCommand(docAddCmd)
	docCmd.AddCommand(docShowCmd)
	docCmd.AddCommand(docDeleteCmd)
	docCmd.AddCommand(docLinkCmd)
	docCmd.AddCommand(docUnlinkCmd)

	docAddCmd.Flags().BoolVarP(&docForce, "force", "f", false, "Overwrite if doc exists")
	docDeleteCmd.Flags().BoolVarP(&docForce, "force", "f", false, "Skip confirmation prompt")
	docLinkCmd.Flags().BoolVar(&docGitignore, "gitignore", false, "Add symlink path to .gitignore")
}

// completeDocNameOnly provides completion for doc names
func completeDocNameOnly(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	reg, err := registry.New()
	if err != nil || !reg.Exists() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	docDir := filepath.Join(reg.BasePath, "doc")
	entries, err := os.ReadDir(docDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), toComplete) {
			completions = append(completions, entry.Name())
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func runDocList(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	docDir := filepath.Join(reg.BasePath, "doc")
	entries, err := os.ReadDir(docDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%s No docs in registry\n", yellow("!"))
			fmt.Println("\nAdd a doc folder:")
			fmt.Println("  agmd doc add ./docs my-docs")
			return nil
		}
		return fmt.Errorf("failed to read doc directory: %w", err)
	}

	// Filter to directories only
	var docs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			docs = append(docs, entry)
		}
	}

	if len(docs) == 0 {
		fmt.Printf("%s No docs in registry\n", yellow("!"))
		fmt.Println("\nAdd a doc folder:")
		fmt.Println("  agmd doc add ./docs my-docs")
		return nil
	}

	fmt.Printf("%s\n\n", cyan(docDir))
	fmt.Printf("doc/ (%d)\n", len(docs))
	for _, entry := range docs {
		// Count files in doc folder
		docPath := filepath.Join(docDir, entry.Name())
		fileCount := countFiles(docPath)
		fmt.Printf("  %s %s\n", entry.Name(), dim(fmt.Sprintf("(%d files)", fileCount)))
	}

	return nil
}

func runDocAdd(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	sourcePath := args[0]
	docName := args[1]

	// Check source exists and is a directory
	sourceInfo, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("source folder not found: %s", sourcePath)
	}
	if err != nil {
		return fmt.Errorf("failed to access source folder: %w", err)
	}
	if !sourceInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s\nUse 'agmd file add' for single files", sourcePath)
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Build destination path
	docDir := filepath.Join(reg.BasePath, "doc")
	destPath := filepath.Join(docDir, docName)

	// Check if destination exists
	if _, err := os.Stat(destPath); err == nil {
		if !docForce {
			return fmt.Errorf("doc:%s already exists (use --force to overwrite)", docName)
		}
		// Remove existing
		if err := os.RemoveAll(destPath); err != nil {
			return fmt.Errorf("failed to remove existing doc: %w", err)
		}
	}

	// Create doc directory
	if err := os.MkdirAll(docDir, 0755); err != nil {
		return fmt.Errorf("failed to create doc directory: %w", err)
	}

	// Copy directory recursively
	fileCount, err := copyDir(sourcePath, destPath)
	if err != nil {
		return fmt.Errorf("failed to copy documentation: %w", err)
	}

	fmt.Printf("%s Added doc:%s (%d files)\n", green("ok"), docName, fileCount)
	fmt.Printf("%s %s\n", blue("->"), destPath)

	return nil
}

func runDocShow(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	name := args[0]

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	docPath := filepath.Join(reg.BasePath, "doc", name)
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("doc:%s not found", name)
	}

	fmt.Printf("%s\n\n", cyan(docPath))

	// List files in tree format
	err = filepath.Walk(docPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(docPath, path)
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			fmt.Printf("%s%s/\n", indent, info.Name())
		} else {
			fmt.Printf("%s%s %s\n", indent, info.Name(), dim(fmt.Sprintf("(%d bytes)", info.Size())))
		}

		return nil
	})

	return err
}

func runDocDelete(cmd *cobra.Command, args []string) error {
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

	docPath := filepath.Join(reg.BasePath, "doc", name)
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("doc:%s not found", name)
	}

	// Count files
	fileCount := countFiles(docPath)

	// Show what will be deleted
	fmt.Printf("%s Deleting doc:%s (%d files)\n", blue("→"), name, fileCount)
	fmt.Printf("  Path: %s\n", docPath)

	// Confirmation prompt (unless --force)
	if !docForce {
		fmt.Printf("\n%s This will permanently delete this documentation folder.\n", yellow("⚠"))
		fmt.Print("\nAre you sure? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			fmt.Println("\nCancelled.")
			return nil
		}
	}

	// Delete the folder
	if err := os.RemoveAll(docPath); err != nil {
		return fmt.Errorf("failed to delete doc: %w", err)
	}

	fmt.Printf("%s Deleted doc:%s\n", green("✓"), name)

	// Check if doc directory is empty
	docDir := filepath.Join(reg.BasePath, "doc")
	entries, err := os.ReadDir(docDir)
	if err == nil && len(entries) == 0 {
		os.Remove(docDir)
	}

	return nil
}

func runDocLink(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()

	name := args[0]

	// Determine destination path
	var destPath string
	if len(args) > 1 {
		destPath = args[1]
	} else {
		destPath = filepath.Join("docs", name)
	}

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Check doc exists in registry
	docPath := filepath.Join(reg.BasePath, "doc", name)
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("doc:%s not found in registry", name)
	}

	// Check if destination already exists
	if info, err := os.Lstat(destPath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			// Already a symlink - check if it points to our doc
			target, _ := os.Readlink(destPath)
			if target == docPath {
				fmt.Printf("%s doc:%s already linked at %s\n", blue("ℹ"), name, destPath)
				return nil
			}
			return fmt.Errorf("%s already exists as symlink to %s", destPath, target)
		}
		return fmt.Errorf("%s already exists", destPath)
	}

	// Create parent directory if needed
	destDir := filepath.Dir(destPath)
	if destDir != "." && destDir != "" {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destDir, err)
		}
	}

	// Create symlink
	if err := os.Symlink(docPath, destPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	fmt.Printf("%s Linked doc:%s → %s\n", green("✓"), name, destPath)

	// Add to .gitignore if requested
	if docGitignore {
		if err := addToGitignore(destPath); err != nil {
			fmt.Printf("%s Failed to update .gitignore: %v\n", blue("ℹ"), err)
		} else {
			fmt.Printf("%s Added %s to .gitignore\n", green("✓"), destPath)
		}
	}

	return nil
}

func runDocUnlink(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()

	name := args[0]

	// Check common locations for the symlink
	possiblePaths := []string{
		filepath.Join("docs", name),
		name,
	}

	var foundPath string
	for _, path := range possiblePaths {
		info, err := os.Lstat(path)
		if err == nil && info.Mode()&os.ModeSymlink != 0 {
			foundPath = path
			break
		}
	}

	if foundPath == "" {
		return fmt.Errorf("no symlink found for doc:%s\nChecked: %s", name, strings.Join(possiblePaths, ", "))
	}

	// Remove the symlink
	if err := os.Remove(foundPath); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	fmt.Printf("%s Unlinked doc:%s (removed %s)\n", green("✓"), name, foundPath)

	return nil
}

// Helper functions

func countFiles(path string) int {
	count := 0
	filepath.Walk(path, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func copyDir(src, dst string) (int, error) {
	fileCount := 0

	return fileCount, filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		fileCount++
		return os.Chmod(dstPath, info.Mode())
	})
}

func addToGitignore(path string) error {
	gitignorePath := ".gitignore"

	// Read existing content
	var content []byte
	if _, err := os.Stat(gitignorePath); err == nil {
		content, _ = os.ReadFile(gitignorePath)
	}

	// Check if already in gitignore
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == path {
			return nil // Already present
		}
	}

	// Append to gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		f.WriteString("\n")
	}

	_, err = f.WriteString(path + "\n")
	return err
}
