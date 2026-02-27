package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listTree bool

var listCmd = &cobra.Command{
	Use:     "list [type]",
	Aliases: []string{"ls"},
	Short:   "List all registry items",
	Long: `List all items in the registry organized by type.

This command lists user-defined types (rule, workflow, guideline, etc.).
Reserved types (task, file, doc, profile) have their own subcommands.

Examples:
  agmd list           # List all items
  agmd list rule      # List only rules
  agmd ls             # Same (alias)
  agmd list --tree    # Show as ASCII tree

Reserved types with subcommands:
  agmd task list      # List project tasks
  agmd file list      # List raw files
  agmd doc list       # List documentation folders`,
	ValidArgsFunction: completeTypeOnly,
	RunE:              runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&listTree, "tree", "t", false, "Display as ASCII tree")
}

func runList(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found\nRun 'agmd setup' first")
	}

	// Handle type:name format for doc/file folder listing
	if len(args) > 0 && strings.Contains(args[0], ":") {
		parts := strings.SplitN(args[0], ":", 2)
		itemType := parts[0]
		name := parts[1]

		if itemType == "doc" {
			return listDocContents(reg, name)
		}
		if itemType == "file" {
			// file: items are single files, not folders
			fmt.Printf("Use 'agmd file show %s' to view file contents\n", name)
			return nil
		}
	}

	// Hint for reserved types
	if len(args) > 0 && args[0] == "task" {
		fmt.Printf("Use 'agmd task list' to list tasks\n")
		return nil
	}
	if len(args) > 0 && args[0] == "doc" {
		fmt.Printf("Use 'agmd doc list' to list docs\n")
		return nil
	}
	if len(args) > 0 && args[0] == "file" {
		fmt.Printf("Use 'agmd file list' to list files\n")
		return nil
	}

	// Tree view
	if listTree {
		return runListTree(reg)
	}

	// Filter by specific type if provided
	var filterType string
	if len(args) > 0 {
		filterType = strings.ToLower(args[0])
	}

	// List by type
	types, err := reg.ListTypes()
	if err != nil {
		return err
	}

	if len(types) == 0 {
		fmt.Printf("%s Registry is empty\n", yellow("!"))
		fmt.Println("\nCreate your first item:")
		fmt.Println("  agmd new rule:my-rule")
		fmt.Println("  agmd new framework:my-framework")
		return nil
	}

	// If filtering by type, check it exists
	if filterType != "" {
		found := false
		for _, t := range types {
			if t == filterType {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("type '%s' not found in registry", filterType)
		}
	}

	fmt.Printf("%s\n\n", cyan(tildeHome(reg.BasePath)))

	for _, typeName := range types {
		// Skip reserved types in general listing (they have their own subcommands)
		if typeName == "task" || typeName == "doc" || typeName == "file" {
			continue
		}

		// Filter by type if specified
		if filterType != "" && typeName != filterType {
			continue
		}

		// Special handling for profiles to show files count
		if typeName == "profile" {
			profiles, err := reg.ListProfiles()
			if err != nil || len(profiles) == 0 {
				continue
			}

			fmt.Printf("%s/ (%d)\n", typeName, len(profiles))
			for _, profile := range profiles {
				filesInfo := ""
				if len(profile.Files) > 0 {
					filesInfo = fmt.Sprintf(" [%d files]", len(profile.Files))
				}
				if profile.Description != "" {
					fmt.Printf("  %s - %s%s\n", profile.Name, profile.Description, filesInfo)
				} else {
					fmt.Printf("  %s%s\n", profile.Name, filesInfo)
				}
			}
			fmt.Println()
			continue
		}

		items, err := reg.ListItems(typeName)
		if err != nil || len(items) == 0 {
			continue
		}

		fmt.Printf("%s/ (%d)\n", typeName, len(items))
		for _, item := range items {
			if item.Description != "" {
				fmt.Printf("  %s - %s\n", item.Name, item.Description)
			} else {
				fmt.Printf("  %s\n", item.Name)
			}
		}
		fmt.Println()
	}

	return nil
}

// runListTree displays registry as an ASCII tree
func runListTree(reg *registry.Registry) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	fmt.Printf("%s\n", cyan(tildeHome(reg.BasePath)))

	tree := buildRegistryTree(reg.BasePath)
	printTree(tree, "", true, dim)

	return nil
}

// TreeNode represents a node in the file tree
type TreeNode struct {
	Name     string
	IsDir    bool
	Children []*TreeNode
}

// buildRegistryTree builds a tree structure from the registry directory
func buildRegistryTree(basePath string) []*TreeNode {
	var nodes []*TreeNode

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nodes
	}

	// Sort: directories first, then alphabetically
	sort.Slice(entries, func(i, j int) bool {
		iDir := entries[i].IsDir()
		jDir := entries[j].IsDir()
		if iDir != jDir {
			return iDir
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		// Skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		node := &TreeNode{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		}

		if entry.IsDir() {
			childPath := filepath.Join(basePath, entry.Name())
			node.Children = buildRegistryTree(childPath)
		}

		nodes = append(nodes, node)
	}

	return nodes
}

// printTree prints the tree with ASCII art
func printTree(nodes []*TreeNode, prefix string, isRoot bool, dim func(a ...interface{}) string) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if node.IsDir {
			fmt.Printf("%s%s%s/\n", dim(prefix), dim(connector), node.Name)
		} else {
			fmt.Printf("%s%s%s\n", dim(prefix), dim(connector), node.Name)
		}

		if len(node.Children) > 0 {
			childPrefix := prefix + "│   "
			if isLast {
				childPrefix = prefix + "    "
			}
			printTree(node.Children, childPrefix, false, dim)
		}
	}
}

// listDocContents shows the contents of a doc folder
func listDocContents(reg *registry.Registry, name string) error {
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	docPath := filepath.Join(reg.BasePath, "doc", name)
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		return fmt.Errorf("doc:%s not found", name)
	}

	fmt.Printf("%s\n\n", cyan(docPath))

	// Build and print tree
	nodes := buildRegistryTree(docPath)
	printTree(nodes, "", true, dim)

	// Count files
	fileCount := 0
	filepath.Walk(docPath, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			fileCount++
		}
		return nil
	})

	fmt.Printf("\n%d files\n", fileCount)
	return nil
}
