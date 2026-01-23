package cmd

import (
	"fmt"
	"os"
	"strings"

	"agmd/pkg/registry"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [profile:name]",
	Short: "Initialize a new project with directives.md",
	Long: `Initialize a new project by creating directives.md (source file).

Without a profile, creates directives.md with:
- A title and introduction
- Example sections showing :::list and :::include directives
- Instructions on how to use agmd

With a profile, creates directives.md from a saved template.

Run 'agmd sync' to create AGENTS.md from directives.md.

Examples:
  agmd init                    # Initialize with default profile
  agmd init profile:svelte-kit # Initialize with svelte-kit profile`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
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
			red(reg.BasePath),
			red("✗"))
	}

	// Check if directives.md already exists
	if _, err := os.Stat(directivesMdFilename); err == nil {
		return fmt.Errorf("directives.md already exists in current directory")
	}

	var templateContent string

	// Use profile if specified
	if profileName != "" {
		profile, err := reg.GetProfile(profileName)
		if err != nil {
			return fmt.Errorf("profile '%s' not found\nRun 'agmd list' to see available profiles", profileName)
		}

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
			templateContent = defaultProfile.Content
			fmt.Printf("%s Using default profile\n", green("✓"))
			if defaultProfile.Description != "" {
				fmt.Printf("  %s\n", defaultProfile.Description)
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

	fmt.Printf("\n%s Project initialized successfully!\n", green("✓"))
	fmt.Println("\nCreated:")
	fmt.Printf("  • %s - Source file with directives (edit this)\n", directivesMdFilename)

	fmt.Println("\nNext steps:")
	fmt.Println("  • Edit directives.md to add directives")
	fmt.Println("  • Run 'agmd add rule <name>' to add rules to directives.md")
	fmt.Println("  • Run 'agmd sync' to create AGENTS.md for AI agents")
	fmt.Println("  • Run 'agmd new rule <name>' to create custom rules")

	return nil
}
