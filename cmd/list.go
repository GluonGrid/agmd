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
	Use:   "list",
	Short: "List all registry items",
	Long: `List all items in the registry organized by type.

Examples:
  agmd list           # List all items
  agmd list --tree    # Show as ASCII tree`,
	RunE: runList,
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

	// Tree view
	if listTree {
		return runListTree(reg)
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

	fmt.Printf("%s\n\n", cyan(reg.BasePath))

	for _, typeName := range types {
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

	fmt.Printf("%s\n", cyan(reg.BasePath))

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
