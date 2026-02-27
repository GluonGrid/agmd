package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initLocal bool

var initCmd = &cobra.Command{
	Use:   "init [profile:name]",
	Short: "Initialize a new project with directives.md",
	Long: `Initialize a new project by creating directives.md (source file).

Without a profile, creates directives.md with:
- A title and introduction
- Example sections showing :::use and :::list directives
- Instructions on how to use agmd

With a profile, creates directives.md from a saved template.

Use --local to also create a project-local .agmd/ registry that takes
priority over your global ~/.agmd for this project. Commit .agmd/ to
share project-specific rules with your team.

Run 'agmd sync' to create AGENTS.md from directives.md.

Examples:
  agmd init                    # Initialize with default profile
  agmd init profile:svelte-kit # Initialize with svelte-kit profile
  agmd init --local            # Also create .agmd/ local registry`,
	ValidArgsFunction: completeProfileName,
	RunE:              runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initLocal, "local", false, "Create a project-local .agmd/ registry alongside directives.md")
}

func runInit(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Check for profile:name argument
	var profileName string
	if len(args) > 0 && strings.HasPrefix(args[0], "profile:") {
		profileName = strings.TrimPrefix(args[0], "profile:")
	}

	if profileName != "" {
		fmt.Printf("%s Initializing agmd project with profile '%s'...\n", blue("→"), profileName)
	} else {
		fmt.Printf("%s Initializing agmd project...\n", blue("→"))
	}

	// Check if registry exists
	reg, err := registry.New()
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	if !reg.Exists() {
		return fmt.Errorf("registry not found at %s\n%s\nRun 'agmd setup' first to initialize the registry",
			red(tildeHome(reg.BasePath)),
			red("✗"))
	}

	// Check if directives.md already exists
	if _, err := os.Stat(directivesMdFilename); err == nil {
		return fmt.Errorf("directives.md already exists in current directory")
	}

	var templateContent string
	var profile *registry.Profile

	// Use profile if specified
	if profileName != "" {
		p, err := reg.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile '%s' not found\nRun 'agmd list' to see available profiles", profileName)
		}
		profile = p
		templateContent = profile.Content
		fmt.Printf("%s Using profile: %s\n", green("✓"), profileName)
		if profile.Description != "" {
			fmt.Printf("  %s\n", profile.Description)
		}
	} else {
		// Try to use 'default' profile if it exists, otherwise use minimal template
		defaultProfile, err := reg.GetProfile("default")
		if err == nil {
			// Default profile exists, use it
			profile = defaultProfile
			templateContent = profile.Content
			fmt.Printf("%s Using default profile\n", green("✓"))
			if profile.Description != "" {
				fmt.Printf("  %s\n", profile.Description)
			}
		} else {
			// Fallback: No default profile found (shouldn't happen after setup)
			fmt.Printf("%s Using fallback template\n", blue("ℹ"))
			fmt.Printf("  Tip: Run 'agmd setup' to create the default profile\n")
			templateContent = registry.GetDefaultDirectivesTemplate()
		}
	}

	// Create directives.md
	fmt.Printf("%s Creating directives.md...\n", blue("→"))
	if err := os.WriteFile(directivesMdFilename, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to write directives.md: %w", err)
	}

	fmt.Printf("%s Created directives.md\n", green("✓"))

	// Always add directives.local.md to .gitignore
	if err := addToGitignore(localDirectivesMdFilename); err == nil {
		fmt.Printf("%s Added %s to .gitignore\n", green("✓"), localDirectivesMdFilename)
	}

	// Copy files from profile if any
	var copiedFiles []string
	if profile != nil && len(profile.Files) > 0 {
		fmt.Printf("%s Copying profile files...\n", blue("→"))
		for _, pf := range profile.Files {
			srcPath := filepath.Join(reg.BasePath, "file", pf.Source)
			destPath := pf.Destination

			// Check if source file exists
			if _, err := os.Stat(srcPath); os.IsNotExist(err) {
				fmt.Printf("  %s file:%s not found, skipping\n", red("✗"), pf.Source)
				continue
			}

			// Check if destination already exists
			if _, err := os.Stat(destPath); err == nil {
				fmt.Printf("  %s %s already exists, skipping\n", blue("ℹ"), destPath)
				continue
			}

			// Create destination directory if needed
			destDir := filepath.Dir(destPath)
			if destDir != "." && destDir != "" {
				if err := os.MkdirAll(destDir, 0755); err != nil {
					fmt.Printf("  %s failed to create directory %s: %v\n", red("✗"), destDir, err)
					continue
				}
			}

			// Copy the file
			content, err := os.ReadFile(srcPath)
			if err != nil {
				fmt.Printf("  %s failed to read file:%s: %v\n", red("✗"), pf.Source, err)
				continue
			}

			if err := os.WriteFile(destPath, content, 0644); err != nil {
				fmt.Printf("  %s failed to write %s: %v\n", red("✗"), destPath, err)
				continue
			}

			fmt.Printf("  %s Copied file:%s → %s\n", green("✓"), pf.Source, destPath)
			copiedFiles = append(copiedFiles, destPath)
		}
	}

	// Create local registry if requested
	if initLocal {
		localPath := ".agmd"
		if _, err := os.Stat(localPath); err == nil {
			fmt.Printf("%s .agmd/ already exists, skipping\n", blue("ℹ"))
		} else {
			fmt.Printf("%s Creating .agmd/ local registry...\n", blue("→"))
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return fmt.Errorf("failed to create .agmd/: %w", err)
			}
			// Add a README so the directory purpose is clear
			readme := "# Local agmd Registry\n\nItems here override your global ~/.agmd registry for this project.\nCommit this directory to share rules with your team.\n\nCreate items with:\n  agmd new rule:my-rule --local\n"
			if err := os.WriteFile(filepath.Join(localPath, "README.md"), []byte(readme), 0644); err != nil {
				return fmt.Errorf("failed to write .agmd/README.md: %w", err)
			}
			fmt.Printf("%s Created .agmd/\n", green("✓"))
		}
	}

	fmt.Printf("\n%s Project initialized successfully!\n", green("✓"))
	fmt.Println("\nCreated:")
	fmt.Printf("  • %s - Source file with directives (edit this)\n", directivesMdFilename)
	if initLocal {
		fmt.Printf("  • .agmd/ - Project-local registry (commit this to share with team)\n")
	}
	for _, f := range copiedFiles {
		fmt.Printf("  • %s\n", f)
	}

	fmt.Println("\nNext steps:")
	fmt.Println("  • Edit directives.md to add directives")
	fmt.Println("  • Run 'agmd sync' to create AGENTS.md for AI agents")
	fmt.Printf("  • Create %s for personal/machine-specific directives (gitignored)\n", localDirectivesMdFilename)
	if initLocal {
		fmt.Println("  • Run 'agmd new rule:<name> --local' to create team-shared rules")
	} else {
		fmt.Println("  • Run 'agmd new rule:<name>' to create rules in your global registry")
	}

	return nil
}
