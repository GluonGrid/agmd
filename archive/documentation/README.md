# agent-config Package Documentation

This directory contains comprehensive documentation for all Go packages used in the agent-config project.

## Documentation Files

### 1. cobra.md - CLI Framework
**Package**: `github.com/spf13/cobra` (v1.10.2)  
**Use in project**: Core CLI framework for all commands

**Contents**:
- ✅ Complete API reference
- ✅ Command structure and lifecycle
- ✅ Flag management (global, local, persistent)
- ✅ Argument validation
- ✅ Shell completions
- ✅ Error handling patterns
- ✅ Best practices for CLI apps
- ✅ Integration with agent-config examples

**When to read**: Before implementing any new commands or modifying existing ones.

### 2. color.md - Terminal Colors
**Package**: `github.com/fatih/color` (v1.18.0)  
**Use in project**: Colorized output for success/error messages

**Contents**:
- ✅ Simple color functions
- ✅ Custom color objects
- ✅ All color attributes (foreground, background, styles)
- ✅ Disable/enable colors programmatically
- ✅ Windows support
- ✅ RGB and 256-color support
- ✅ Practical examples for CLI status messages
- ✅ Best practices for readable output

**When to read**: When adding colored output to commands or improving UX.

### 3. viper.md - Configuration Management
**Package**: `github.com/spf13/viper` (v1.21.0)  
**Use in project**: Optional (for future enhancements)

**Contents**:
- ✅ Reading config files (YAML, JSON, TOML, etc.)
- ✅ Environment variables
- ✅ Integration with Cobra flags
- ✅ Unmarshaling to structs
- ✅ Configuration precedence
- ✅ Live config watching
- ✅ Remote config stores
- ✅ Complete application examples

**When to read**: If adding global configuration support (e.g., `~/.agent-config.yaml`).

**Note**: Not currently used in agent-config but documented for future reference.

### 4. go-stdlib.md - Standard Library
**Packages**: `os`, `path/filepath`, `strings`, `fmt`, `io`  
**Use in project**: File operations, path manipulation, string handling

**Contents**:
- ✅ File reading/writing (os.ReadFile, os.WriteFile)
- ✅ Symlink operations (os.Symlink, os.Readlink, os.Lstat)
- ✅ Path manipulation (filepath.Join, filepath.Dir, filepath.Ext)
- ✅ String operations (strings.Builder, strings.Join, etc.)
- ✅ Formatted output (fmt.Printf, fmt.Sprintf)
- ✅ Directory walking (filepath.Walk)
- ✅ Practical examples for agent-config

**When to read**: When working with file operations or path handling.

## Quick Reference

### Adding Dependencies

Remember the package management rules from GO_PACKAGE_MANAGEMENT.md:

```bash
# Add a package (like 'uv add' or 'cargo add')
go get github.com/package/name@latest

# Clean up
go mod tidy
```

### Common Patterns for agent-config

#### 1. Create a new Cobra command

See: **cobra.md** → "Subcommand Pattern"

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE:  runMyCommand,
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

#### 2. Add colored output

See: **color.md** → "Status Messages"

```go
green := color.New(color.FgGreen).SprintFunc()
fmt.Printf("%s %s\n", green("✓"), "Success message")
```

#### 3. Read/write agent.md

See: **go-stdlib.md** → "Common Patterns for agent-config"

```go
content, err := os.ReadFile("agent.md")
err := os.WriteFile("agent.md", []byte(content), 0644)
```

#### 4. Create symlinks

See: **go-stdlib.md** → "Symlinks"

```go
err := os.Symlink("agent.md", "CLAUDE.md")
target, err := os.Readlink("CLAUDE.md")
```

## Package Usage Matrix

| Package | Command | Purpose |
|---------|---------|---------|
| **cobra** | All | CLI structure |
| **color** | All | Colored output |
| **os** | init, add, list, edit, symlink | File operations |
| **path/filepath** | init, symlink | Path handling |
| **strings** | add, list | String manipulation |
| **fmt** | All | Output formatting |
| **viper** | - | (Future: global config) |

## Learning Path

### For Beginners
1. Start with **cobra.md** → Understand CLI structure
2. Read **color.md** → Learn output formatting
3. Study **go-stdlib.md** → Master file operations

### For Implementing Features
1. Check **cobra.md** for command patterns
2. Reference **go-stdlib.md** for file operations
3. Use **color.md** for user feedback

### For Advanced Features
1. Read **viper.md** if adding configuration
2. Study **cobra.md** advanced sections
3. Review **go-stdlib.md** advanced patterns

## External Resources

### Official Documentation
- **Cobra**: https://cobra.dev
- **Viper**: https://github.com/spf13/viper
- **Color**: https://github.com/fatih/color
- **Go Stdlib**: https://pkg.go.dev/std

### Package Registry
- **pkg.go.dev**: https://pkg.go.dev
- Search: `pkg.go.dev/package-name`

### Related Guides
- **GO_PACKAGE_MANAGEMENT.md**: Package management best practices
- **PROJECT_PLAN.md**: Implementation roadmap
- **QUICKSTART.md**: Getting started guide

## Tips

### 1. Use the Search Function
All these Markdown files are searchable. Use Ctrl+F (Cmd+F on Mac) to find specific functions or patterns.

### 2. Copy-Paste Ready
All code examples are tested and ready to use. Just adapt them to your needs.

### 3. Keep Updated
When upgrading packages (`go get -u`), check for breaking changes in official docs.

### 4. Offline Access
Download these files for offline reference during development.

## Quick Command Reference

### Cobra
```go
cobra.Command{Use, Short, Long, Run, RunE, Args}
cmd.Flags().StringP(), cmd.PersistentFlags()
rootCmd.AddCommand(subCmd)
```

### Color
```go
color.Green("text"), color.Red("text")
green := color.New(color.FgGreen).SprintFunc()
fmt.Printf("%s Success\n", green("✓"))
```

### File Operations
```go
os.ReadFile(), os.WriteFile()
os.Symlink(), os.Readlink(), os.Lstat()
filepath.Join(), filepath.Dir(), filepath.Ext()
```

### Strings
```go
strings.Join(), strings.Split()
strings.Contains(), strings.HasPrefix()
var builder strings.Builder; builder.WriteString()
```

---

**Last Updated**: January 2025  
**Project**: agent-config CLI tool  
**Go Version**: 1.21+
