# fatih/color - Terminal Color Output

## Package Information

**Import Path**: `github.com/fatih/color`  
**Latest Version**: v1.18.0  
**License**: MIT  
**Go Get**: `go get github.com/fatih/color@latest`

## Overview

The `color` package provides an ANSI color package to output colorized or SGR defined output to the standard output. It has full Windows support and works across different terminals.

### Key Features

- Simple API for colorized output
- Support for 16 basic ANSI colors
- Support for 256 colors
- Support for RGB/True color (24-bit)
- Chainable color attributes (bold, underline, etc.)
- Disable color output programmatically
- Works on Windows, Linux, macOS
- NO_COLOR environment variable support
- Printf-style formatting

## Installation

```bash
go get github.com/fatih/color@latest
```

## Basic Usage

### Simple Helper Functions

```go
package main

import "github.com/fatih/color"

func main() {
    // Basic colors - newline appended automatically
    color.Red("Error: something went wrong")
    color.Green("Success!")
    color.Yellow("Warning: check this")
    color.Blue("Info: processing...")
    color.Magenta("Debug: value = %d", 42)
    color.Cyan("Status: ready")
    color.White("Normal text")
    
    // Hi-intensity (bright) colors
    color.HiRed("Bright red")
    color.HiGreen("Bright green")
    color.HiBlack("Bright black (gray)")
}
```

### Custom Color Objects

```go
// Create a new color object
c := color.New(color.FgCyan)
c.Println("Cyan text")

// Add multiple attributes
c = color.New(color.FgCyan, color.Bold, color.Underline)
c.Println("Cyan, bold, underlined text")

// Or chain them
color.New(color.FgWhite, color.BgBlack).Println("White on black")
```

## Color Attributes

### Foreground Colors

```go
color.FgBlack
color.FgRed
color.FgGreen
color.FgYellow
color.FgBlue
color.FgMagenta
color.FgCyan
color.FgWhite

// Hi-intensity (bright) colors
color.FgHiBlack    // Gray
color.FgHiRed
color.FgHiGreen
color.FgHiYellow
color.FgHiBlue
color.FgHiMagenta
color.FgHiCyan
color.FgHiWhite
```

### Background Colors

```go
color.BgBlack
color.BgRed
color.BgGreen
color.BgYellow
color.BgBlue
color.BgMagenta
color.BgCyan
color.BgWhite

// Hi-intensity backgrounds
color.BgHiBlack
color.BgHiRed
color.BgHiGreen
color.BgHiYellow
color.BgHiBlue
color.BgHiMagenta
color.BgHiCyan
color.BgHiWhite
```

### Text Attributes

```go
color.Reset        // Reset all attributes
color.Bold         // Bold/bright text
color.Faint        // Faint/dim text
color.Italic       // Italic text
color.Underline    // Underlined text
color.BlinkSlow    // Slow blink
color.BlinkRapid   // Rapid blink
color.ReverseVideo // Reverse video (swap fg/bg)
color.Concealed    // Concealed text
color.CrossedOut   // Crossed out text
```

## Creating Reusable Colors

```go
// Create color functions
red := color.New(color.FgRed).SprintFunc()
green := color.New(color.FgGreen).SprintFunc()
yellow := color.New(color.FgYellow).SprintFunc()

// Use them
fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red("error"))
fmt.Printf("Status: %s\n", green("OK"))
```

## String Functions

### Colored Strings

```go
// Returns colored string without printing
s := color.RedString("This is red")
s := color.BlueString("This is %s", "blue")

// Use in fmt.Printf
fmt.Printf("Error: %s\n", color.RedString("failed"))
```

### Sprint Functions

```go
// Create a color and get Sprint function
red := color.New(color.FgRed).SprintFunc()
boldRed := color.New(color.FgRed, color.Bold).SprintFunc()

text := red("This will be red")
bold := boldRed("This will be bold red")
```

## Methods on Color Objects

```go
c := color.New(color.FgCyan)

// Print methods
c.Print("Cyan text")         // No newline
c.Println("Cyan text")       // With newline
c.Printf("Cyan %s", "text")  // Formatted

// Sprint methods (return string)
s := c.Sprint("Cyan text")
s = c.Sprintf("Cyan %s", "text")

// Fprint methods (write to io.Writer)
c.Fprint(os.Stdout, "Cyan text")
c.Fprintf(os.Stdout, "Cyan %s", "text")
```

## Chaining Attributes

```go
// Add attributes to existing color
c := color.New(color.FgRed)
c.Add(color.Bold)
c.Add(color.Underline)
c.Println("Bold, underlined red text")

// Or create with multiple attributes
c = color.New(color.FgRed, color.Bold, color.Underline)
c.Println("Same result")
```

## Set/Unset Global Colors

```go
// Set foreground color for subsequent output
color.Set(color.FgYellow)
fmt.Println("This text will be yellow")
fmt.Println("This too")
color.Unset() // Reset to default

// With multiple attributes
color.Set(color.FgMagenta, color.Bold)
defer color.Unset() // Good practice
fmt.Println("Bold magenta text")
```

## Disabling/Enabling Colors

### Global Disable/Enable

```go
import "flag"

var noColor = flag.Bool("no-color", false, "Disable color output")

func main() {
    flag.Parse()
    
    if *noColor {
        color.NoColor = true // Disables all colors
    }
    
    color.Green("This will be green (or plain if --no-color)")
}
```

### Per-Color Disable/Enable

```go
c := color.New(color.FgCyan)
c.Println("Cyan text")

c.DisableColor()
c.Println("Plain text (no color)")

c.EnableColor()
c.Println("Cyan text again")
```

### Environment Variables

The package automatically checks for:
- `NO_COLOR` - If set to any non-empty value, colors are disabled
- `TERM=dumb` - Colors are disabled for dumb terminals

## Windows Support

```go
// For Windows, always use color.Output for Fprintf
fmt.Fprintf(color.Output, "Windows support: %s", color.GreenString("PASS"))

// Or create color with Sprint functions
info := color.New(color.FgWhite, color.BgGreen).SprintFunc()
fmt.Fprintf(color.Output, "This %s rocks!\n", info("package"))
```

## RGB Colors (24-bit / True Color)

```go
// If terminal supports 24-bit colors
c := color.RGB(255, 128, 0) // Orange
c.Println("This is orange")

// With background
c = color.RGB(255, 255, 255).BgRGB(0, 0, 0) // White on black
c.Println("White text on black background")
```

## 256 Color Support

```go
// Use color code from 256 color palette
c := color.New(color.FgColor256(208)) // Orange
c.Println("256 color mode")

// Background color
c = color.New(color.FgColor256(15), color.BgColor256(240))
c.Println("Custom 256 colors")
```

## Common Use Cases for agent-config

### Status Messages

```go
package main

import "github.com/fatih/color"

func printSuccess(msg string) {
    green := color.New(color.FgGreen).SprintFunc()
    fmt.Printf("%s %s\n", green("✓"), msg)
}

func printError(msg string) {
    red := color.New(color.FgRed).SprintFunc()
    fmt.Printf("%s %s\n", red("✗"), msg)
}

func printWarning(msg string) {
    yellow := color.New(color.FgYellow).SprintFunc()
    fmt.Printf("%s %s\n", yellow("⚠"), msg)
}

func printInfo(msg string) {
    blue := color.New(color.FgBlue).SprintFunc()
    fmt.Printf("%s %s\n", blue("ℹ"), msg)
}

func main() {
    printSuccess("Created agent.md")
    printError("File already exists")
    printWarning("Section not found")
    printInfo("Downloading dependencies...")
}
```

### Highlighting in Output

```go
// Highlight section names
cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
fmt.Printf("Added to [%s]: %s\n", cyan("Code Style"), rule)

// Emphasize important parts
yellow := color.New(color.FgYellow).SprintFunc()
fmt.Printf("Note: %s is not a standard section\n", yellow(section))
```

### Interactive Prompts

```go
green := color.New(color.FgGreen).SprintFunc()
fmt.Printf("%s ", green("✓"))
fmt.Print("Added rule successfully\n")
```

### Structured Output

```go
func listRules(rules []Rule) {
    cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
    green := color.New(color.FgGreen).SprintFunc()
    
    for _, rule := range rules {
        fmt.Printf("%s\n", cyan("## " + rule.Section))
        fmt.Printf("  %s %s\n", green("•"), rule.Content)
    }
}
```

## Best Practices

1. **Create reusable color functions** - Don't create new colors in loops
2. **Use Sprint functions for flexibility** - Compose colored strings with other text
3. **Respect NO_COLOR** - The package handles this automatically
4. **Use consistent colors** - Same color for same message types
5. **Test without colors** - Ensure output is readable when colors are disabled
6. **Use color.Output on Windows** - For Fprintf on Windows
7. **Don't overuse colors** - Too many colors can be distracting

## Color Combinations

```go
// Success with checkmark
success := color.New(color.FgGreen)
success.Print("✓ ")

// Error with X
failure := color.New(color.FgRed)
failure.Print("✗ ")

// Warning with triangle
warning := color.New(color.FgYellow)
warning.Print("⚠ ")

// Info with i
info := color.New(color.FgBlue)
info.Print("ℹ ")

// Highlighted text
highlight := color.New(color.FgCyan, color.Bold)
highlight.Print("[IMPORTANT] ")
```

## Testing Color Output

```go
// Disable colors in tests
func TestSomething(t *testing.T) {
    color.NoColor = true
    defer func() { color.NoColor = false }()
    
    // Test output without ANSI codes
}
```

## Complete Example for CLI

```go
package main

import (
    "fmt"
    "github.com/fatih/color"
)

var (
    success = color.New(color.FgGreen).SprintFunc()
    failure = color.New(color.FgRed).SprintFunc()
    warning = color.New(color.FgYellow).SprintFunc()
    info    = color.New(color.FgBlue).SprintFunc()
    cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
)

func main() {
    fmt.Printf("%s Created agent.md\n", success("✓"))
    fmt.Printf("%s Added to [%s]: Never use fixed versions\n", 
        success("✓"), cyan("General Rules"))
    
    fmt.Printf("%s File already exists\n", failure("✗"))
    
    fmt.Printf("%s %s is not a standard section\n", 
        warning("⚠"), "Custom Section")
    
    fmt.Printf("%s Downloading dependencies...\n", info("ℹ"))
}
```

## References

- **GitHub**: https://github.com/fatih/color
- **Go Package**: https://pkg.go.dev/github.com/fatih/color
- **License**: MIT
