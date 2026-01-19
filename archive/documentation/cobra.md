# Cobra - CLI Framework Documentation

## Package Information

**Import Path**: `github.com/spf13/cobra`  
**Latest Version**: v1.10.2  
**License**: Apache-2.0  
**Go Get**: `go get github.com/spf13/cobra@latest`

## Overview

Cobra is a library for creating powerful modern CLI applications. It's used in major projects like Kubernetes, Hugo, and GitHub CLI.

### Key Features

- Easy subcommand-based CLIs: `app server`, `app fetch`, etc.
- Fully POSIX-compliant flags (including short & long versions)
- Nested subcommands
- Global, local and cascading flags
- Intelligent suggestions (`app srver`... did you mean `app server`?)
- Automatic help generation for commands and flags
- Grouping help for subcommands
- Automatic help flag recognition of `-h`, `--help`, etc.
- Automatically generated shell autocomplete (bash, zsh, fish, powershell)
- Automatically generated man pages
- Command aliases
- Optional integration with [viper](https://github.com/spf13/viper) for 12-factor apps

## Core Concepts

### Commands

Commands represent actions. Each interaction the application supports is contained in a Command. A command can have children commands and optionally run an action.

```go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A brief description",
    Long:  `A longer description...`,
    Run: func(cmd *cobra.Command, args []string) {
        // Do stuff here
    },
}
```

### Flags

Flags are modifiers for commands. Cobra supports fully POSIX-compliant flags and the Go flag package.

```go
// Persistent flag (available to command and all subcommands)
rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

// Local flag (only available to this command)
rootCmd.Flags().BoolP("verbose", "v", false, "verbose output")
```

## Installation

```bash
go get github.com/spf13/cobra@latest
```

## Basic Usage

### Creating a Root Command

```go
package main

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "myapp",
        Short: "My application",
        Long:  `A longer description of my application.`,
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Hello from myapp!")
        },
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### Adding Subcommands

```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print the version number",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("myapp v1.0.0")
    },
}

func init() {
    rootCmd.AddCommand(versionCmd)
}
```

### Working with Flags

```go
var name string
var age int

func init() {
    // String flag with shorthand
    rootCmd.Flags().StringVarP(&name, "name", "n", "", "Your name")
    
    // Int flag
    rootCmd.Flags().IntVar(&age, "age", 0, "Your age")
    
    // Mark flag as required
    rootCmd.MarkFlagRequired("name")
}
```

## Command Structure

```go
type Command struct {
    // Use is the one-line usage message
    Use string
    
    // Aliases is an array of aliases
    Aliases []string
    
    // Short description (shown in 'help' output)
    Short string
    
    // Long description (shown in 'help <command>')
    Long string
    
    // Example usage
    Example string
    
    // Valid arguments for shell completion
    ValidArgs []Completion
    ValidArgsFunction CompletionFunc
    
    // Expected arguments
    Args PositionalArgs
    
    // Version for this command
    Version string
    
    // Run functions
    PersistentPreRun func(cmd *Command, args []string)
    PersistentPreRunE func(cmd *Command, args []string) error
    PreRun func(cmd *Command, args []string)
    PreRunE func(cmd *Command, args []string) error
    Run func(cmd *Command, args []string)
    RunE func(cmd *Command, args []string) error
    PostRun func(cmd *Command, args []string)
    PostRunE func(cmd *Command, args []string) error
    PersistentPostRun func(cmd *Command, args []string)
    PersistentPostRunE func(cmd *Command, args []string) error
    
    // Configuration
    TraverseChildren bool
    Hidden bool
    SilenceErrors bool
    SilenceUsage bool
}
```

## Run Function Order

The `*Run` functions are executed in this order:
1. `PersistentPreRun()`
2. `PreRun()`
3. `Run()`
4. `PostRun()`
5. `PersistentPostRun()`

```go
var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Runs before any command in the tree
    },
    PreRun: func(cmd *cobra.Command, args []string) {
        // Runs before Run
    },
    Run: func(cmd *cobra.Command, args []string) {
        // Main command logic
    },
    PostRun: func(cmd *cobra.Command, args []string) {
        // Runs after Run
    },
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        // Runs after any command in the tree
    },
}
```

## Argument Validation

Cobra provides built-in validators:

```go
// NoArgs - command will report error if any args
NoArgs

// ArbitraryArgs - command will accept any args
ArbitraryArgs

// OnlyValidArgs - command will report error if args not in ValidArgs
OnlyValidArgs

// MinimumNArgs(n) - command will report error if less than n args
cobra.MinimumNArgs(1)

// MaximumNArgs(n) - command will report error if more than n args
cobra.MaximumNArgs(3)

// ExactArgs(n) - command will report error if not exactly n args
cobra.ExactArgs(2)

// RangeArgs(min, max) - command will report error if args not in range
cobra.RangeArgs(1, 3)
```

Example:
```go
var cmd = &cobra.Command{
    Use:   "print [string to print]",
    Short: "Print anything to the screen",
    Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(strings.Join(args, " "))
    },
}
```

## Flag Groups

### Required Together

```go
rootCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
rootCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
rootCmd.MarkFlagsRequiredTogether("username", "password")
```

### Mutually Exclusive

```go
rootCmd.Flags().BoolVar(&json, "json", false, "Output in JSON")
rootCmd.Flags().BoolVar(&yaml, "yaml", false, "Output in YAML")
rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")
```

### One Required

```go
rootCmd.Flags().StringVar(&inputFile, "file", "", "Input file")
rootCmd.Flags().StringVar(&inputURL, "url", "", "Input URL")
rootCmd.MarkFlagsOneRequired("file", "url")
```

## Shell Completions

### Register Completion Functions

```go
cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]Completion, ShellCompDirective) {
    return []Completion{
        {Choice: "json", Description: "JSON output"},
        {Choice: "yaml", Description: "YAML output"},
        {Choice: "text", Description: "Plain text"},
    }, ShellCompDirectiveNoFileComp
})
```

### Generate Completion Scripts

```go
// Bash
cmd.GenBashCompletionFileV2("completion.bash", true)

// Zsh
cmd.GenZshCompletionFile("completion.zsh")

// Fish
cmd.GenFishCompletionFile("completion.fish", true)

// PowerShell
cmd.GenPowerShellCompletionFile("completion.ps1")
```

## Error Handling

### Using RunE

```go
var cmd = &cobra.Command{
    Use:   "example",
    Short: "An example command",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := doSomething(); err != nil {
            return err
        }
        return nil
    },
}
```

### Silencing Errors

```go
cmd.SilenceErrors = true  // Don't print errors
cmd.SilenceUsage = true   // Don't print usage on error
```

## Help and Usage

### Custom Help Template

```go
cmd.SetHelpTemplate(`{{.Long}}

Usage:
  {{.UseLine}}

{{if .HasAvailableSubCommands}}Available Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}

{{if .HasAvailableLocalFlags}}Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`)
```

### Custom Usage Function

```go
cmd.SetUsageFunc(func(c *cobra.Command) error {
    fmt.Fprintf(c.OutOrStderr(), "Usage: %s\n", c.UseLine())
    return nil
})
```

## Hooks and Initialization

```go
import "github.com/spf13/cobra"

func init() {
    // Called before Execute()
    cobra.OnInitialize(initConfig)
}

func initConfig() {
    // Read config file, setup logging, etc.
}
```

## Context Support

```go
ctx := context.WithTimeout(context.Background(), 5*time.Second)
if err := cmd.ExecuteContext(ctx); err != nil {
    log.Fatal(err)
}

// Access context in Run functions
cmd.Run = func(cmd *cobra.Command, args []string) {
    ctx := cmd.Context()
    // Use context
}
```

## Common Patterns for agent-config

### Basic Command Structure

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "agent-config",
    Short: "Manage AI agent configuration files",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.CompletionOptions.DisableDefaultCmd = true
}
```

### Subcommand Pattern

```go
package cmd

import (
    "github.com/spf13/cobra"
)

var (
    flagSection string
)

var addCmd = &cobra.Command{
    Use:   "add [rule]",
    Short: "Add a rule to agent.md",
    Args:  cobra.MinimumNArgs(1),
    RunE:  runAdd,
}

func init() {
    rootCmd.AddCommand(addCmd)
    addCmd.Flags().StringVarP(&flagSection, "section", "s", "General Rules", "Section to add the rule to")
}

func runAdd(cmd *cobra.Command, args []string) error {
    rule := strings.Join(args, " ")
    // Implementation
    return nil
}
```

### Flag Validation

```go
cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
    if flagName == "" {
        return fmt.Errorf("--name is required")
    }
    return nil
}
```

## Best Practices

1. **Use RunE instead of Run** - Allows better error handling
2. **Validate flags in PreRunE** - Catch errors before Run executes
3. **Keep commands focused** - One command, one responsibility
4. **Use persistent flags wisely** - Only for flags truly needed by all subcommands
5. **Provide good help text** - Users read Short and Long descriptions
6. **Test with cobra.Command.Execute()** - Don't call Run() directly
7. **Use completion functions** - Enhance UX with shell completions

## References

- **GitHub**: https://github.com/spf13/cobra
- **Documentation**: https://cobra.dev
- **Go Package**: https://pkg.go.dev/github.com/spf13/cobra
- **User Guide**: https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
