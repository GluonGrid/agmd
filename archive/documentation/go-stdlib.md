# Go Standard Library - File Operations & Utilities

This guide covers the essential Go standard library packages used in the agent-config project.

## os Package

**Import**: `import "os"`

### File Operations

#### Reading Files

```go
// Read entire file into memory
content, err := os.ReadFile("agent.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(content))

// Open file for reading
file, err := os.Open("agent.md")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Read with buffer
buf := make([]byte, 1024)
n, err := file.Read(buf)
```

#### Writing Files

```go
// Write entire file (overwrites existing)
data := []byte("# Agent Configuration\n\nRules here...")
err := os.WriteFile("agent.md", data, 0644)
if err != nil {
    log.Fatal(err)
}

// Create/overwrite file
file, err := os.Create("agent.md")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Write string
file.WriteString("# Agent Configuration\n")

// Write bytes
file.Write([]byte("content"))
```

#### Appending to Files

```go
// Open in append mode
file, err := os.OpenFile("agent.md", os.O_APPEND|os.O_WRONLY, 0644)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

file.WriteString("\n- New rule\n")
```

### File/Directory Checks

```go
// Check if file exists
if _, err := os.Stat("agent.md"); os.IsNotExist(err) {
    fmt.Println("File does not exist")
} else if err != nil {
    log.Fatal(err)
} else {
    fmt.Println("File exists")
}

// Get file info
info, err := os.Stat("agent.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Size:", info.Size())
fmt.Println("Mode:", info.Mode())
fmt.Println("Modified:", info.ModTime())
fmt.Println("Is Dir:", info.IsDir())

// Check if path is directory
info, err := os.Stat("path")
if err == nil && info.IsDir() {
    fmt.Println("It's a directory")
}
```

### Directory Operations

```go
// Create directory
err := os.Mkdir("docs", 0755)

// Create directory and parents
err := os.MkdirAll("docs/api/v1", 0755)

// Remove file
err := os.Remove("agent.md")

// Remove directory and contents
err := os.RemoveAll("docs")

// Read directory
entries, err := os.ReadDir(".")
if err != nil {
    log.Fatal(err)
}
for _, entry := range entries {
    fmt.Println(entry.Name(), entry.IsDir())
}
```

### Symlinks

```go
// Create symlink
err := os.Symlink("agent.md", "CLAUDE.md")
if err != nil {
    log.Fatal(err)
}

// Read symlink target
target, err := os.Readlink("CLAUDE.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Points to:", target)

// Get file info without following symlink
info, err := os.Lstat("CLAUDE.md")
if err != nil {
    log.Fatal(err)
}
if info.Mode()&os.ModeSymlink != 0 {
    fmt.Println("It's a symlink")
}
```

### Environment Variables

```go
// Get environment variable
home := os.Getenv("HOME")
path := os.Getenv("PATH")

// Set environment variable
os.Setenv("MY_VAR", "value")

// Get all environment variables
for _, env := range os.Environ() {
    fmt.Println(env)
}

// Expand environment variables in string
expanded := os.ExpandEnv("$HOME/.config")
```

### Working Directory

```go
// Get current directory
dir, err := os.Getwd()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Current dir:", dir)

// Change directory
err := os.Chdir("/tmp")
```

### Process Information

```go
// Get executable path
exe, err := os.Executable()

// Get process ID
pid := os.Getpid()

// Exit program
os.Exit(1)
```

### File Permissions

```go
// Change file permissions
err := os.Chmod("agent.md", 0644)

// Change file owner (Unix only)
err := os.Chown("agent.md", uid, gid)
```

---

## path/filepath Package

**Import**: `import "path/filepath"`

### Path Manipulation

```go
// Join path components
path := filepath.Join("home", "user", "agent.md")
// Result: home/user/agent.md (or home\user\agent.md on Windows)

// Get directory
dir := filepath.Dir("/home/user/agent.md")
// Result: /home/user

// Get filename
base := filepath.Base("/home/user/agent.md")
// Result: agent.md

// Get file extension
ext := filepath.Ext("agent.md")
// Result: .md

// Split extension
name := "agent.md"
ext := filepath.Ext(name)
nameWithoutExt := name[:len(name)-len(ext)]
// nameWithoutExt: agent

// Split path
dir, file := filepath.Split("/home/user/agent.md")
// dir: /home/user/, file: agent.md
```

### Path Cleaning

```go
// Clean path (remove .. and .)
clean := filepath.Clean("/home/user/../user/./agent.md")
// Result: /home/user/agent.md

// Get absolute path
abs, err := filepath.Abs("agent.md")
if err != nil {
    log.Fatal(err)
}
fmt.Println(abs) // /current/directory/agent.md
```

### Path Matching

```go
// Match glob pattern
matched, err := filepath.Match("*.md", "agent.md")
// matched: true

// Find all matching files
matches, err := filepath.Glob("*.md")
if err != nil {
    log.Fatal(err)
}
for _, match := range matches {
    fmt.Println(match)
}
```

### Walking Directory Trees

```go
err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    
    if !info.IsDir() && filepath.Ext(path) == ".md" {
        fmt.Println("Found:", path)
    }
    
    return nil
})

// Skip directories
err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
    if err != nil {
        return err
    }
    
    if info.IsDir() && info.Name() == "node_modules" {
        return filepath.SkipDir
    }
    
    fmt.Println(path)
    return nil
})
```

### WalkDir (Go 1.16+, faster)

```go
err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
    if err != nil {
        return err
    }
    
    if !d.IsDir() {
        fmt.Println(path)
    }
    
    return nil
})
```

### Relative Path

```go
// Get relative path
rel, err := filepath.Rel("/home/user", "/home/user/docs/agent.md")
// rel: docs/agent.md
```

### Volume Name (Windows)

```go
vol := filepath.VolumeName("C:\\Users\\user\\agent.md")
// vol: C:
```

---

## strings Package

**Import**: `import "strings"`

### String Manipulation

```go
// Join strings
joined := strings.Join([]string{"one", "two", "three"}, ", ")
// Result: "one, two, three"

// Split string
parts := strings.Split("one,two,three", ",")
// Result: ["one", "two", "three"]

// Split with limit
parts := strings.SplitN("one,two,three", ",", 2)
// Result: ["one", "two,three"]

// Contains
hasSubstr := strings.Contains("agent.md", ".md")
// Result: true

// Has prefix/suffix
hasPrefix := strings.HasPrefix("agent.md", "agent")
hasPrefix := strings.HasSuffix("agent.md", ".md")

// Replace
replaced := strings.Replace("foo foo foo", "foo", "bar", 2)
// Result: "bar bar foo"

// Replace all
replaced := strings.ReplaceAll("foo foo foo", "foo", "bar")
// Result: "bar bar bar"
```

### Case Conversion

```go
upper := strings.ToUpper("agent config")
// Result: "AGENT CONFIG"

lower := strings.ToLower("AGENT CONFIG")
// Result: "agent config"

title := strings.Title("agent config")
// Result: "Agent Config"
```

### Trimming

```go
// Trim whitespace
trimmed := strings.TrimSpace("  agent.md  ")
// Result: "agent.md"

// Trim specific characters
trimmed := strings.Trim("!!!agent!!!", "!")
// Result: "agent"

// Trim prefix/suffix
trimmed := strings.TrimPrefix("agent.md", "agent")
// Result: ".md"

trimmed := strings.TrimSuffix("agent.md", ".md")
// Result: "agent"
```

### Searching

```go
// Find index
index := strings.Index("agent config", "config")
// Result: 6

// Find last index
index := strings.LastIndex("foo bar foo", "foo")
// Result: 8

// Count occurrences
count := strings.Count("cheese", "e")
// Result: 3
```

### String Builder (Efficient)

```go
var builder strings.Builder

builder.WriteString("# Agent Configuration\n")
builder.WriteString("\n## General Rules\n")
builder.WriteString("- Rule 1\n")
builder.WriteString("- Rule 2\n")

result := builder.String()
```

---

## fmt Package

**Import**: `import "fmt"`

### Print Functions

```go
// Print to stdout
fmt.Print("Hello")
fmt.Println("Hello") // with newline
fmt.Printf("Hello %s\n", "World")

// Print to stderr
fmt.Fprint(os.Stderr, "Error: ")
fmt.Fprintln(os.Stderr, "something went wrong")
fmt.Fprintf(os.Stderr, "Error: %v\n", err)

// Format to string
s := fmt.Sprint("Hello")
s := fmt.Sprintln("Hello")
s := fmt.Sprintf("Hello %s", "World")
```

### Format Verbs

```go
// General
%v  // default format
%+v // with field names (structs)
%#v // Go-syntax representation
%T  // type

// Boolean
%t  // true or false

// Integer
%d  // decimal
%b  // binary
%o  // octal
%x  // hexadecimal lowercase
%X  // hexadecimal uppercase

// Float
%f  // decimal point, no exponent
%e  // scientific notation
%.2f // 2 decimal places

// String
%s  // string
%q  // quoted string

// Pointer
%p  // pointer address

// Width and alignment
%5d   // width 5, right aligned
%-5d  // width 5, left aligned
%05d  // width 5, zero padded
```

### Examples

```go
name := "agent-config"
version := "1.0.0"
port := 8080

fmt.Printf("Starting %s v%s on port %d\n", name, version, port)
// Starting agent-config v1.0.0 on port 8080

// Struct printing
type Config struct {
    Name string
    Port int
}
cfg := Config{"agent-config", 8080}

fmt.Printf("%v\n", cfg)   // {agent-config 8080}
fmt.Printf("%+v\n", cfg)  // {Name:agent-config Port:8080}
fmt.Printf("%#v\n", cfg)  // main.Config{Name:"agent-config", Port:8080}
```

---

## io and io/ioutil Packages

**Import**: `import "io"` and `import "io/ioutil"` (deprecated in Go 1.16+)

### io Package

```go
// Copy data
src, _ := os.Open("source.txt")
dst, _ := os.Create("dest.txt")
io.Copy(dst, src)

// Copy with limit
io.CopyN(dst, src, 1024) // Copy max 1024 bytes

// Read all
reader := strings.NewReader("Hello, World!")
data, err := io.ReadAll(reader)

// Write string
io.WriteString(file, "content")
```

### Reading Files (Modern Go)

```go
// Read entire file (Go 1.16+)
content, err := os.ReadFile("agent.md")

// Write entire file (Go 1.16+)
err := os.WriteFile("agent.md", []byte("content"), 0644)

// Read directory (Go 1.16+)
entries, err := os.ReadDir(".")
```

---

## Common Patterns for agent-config

### Read and Parse agent.md

```go
func readAgentMd() ([]byte, error) {
    content, err := os.ReadFile("agent.md")
    if err != nil {
        if os.IsNotExist(err) {
            return nil, fmt.Errorf("agent.md not found")
        }
        return nil, err
    }
    return content, nil
}
```

### Write agent.md

```go
func writeAgentMd(content string) error {
    return os.WriteFile("agent.md", []byte(content), 0644)
}
```

### Check if Symlink Exists

```go
func isSymlink(path string) bool {
    info, err := os.Lstat(path)
    if err != nil {
        return false
    }
    return info.Mode()&os.ModeSymlink != 0
}
```

### Create Symlinks for Tools

```go
func createSymlinks(tools []string) error {
    symlinkMap := map[string]string{
        "claude":   "CLAUDE.md",
        "cursor":   ".cursorrules",
        "windsurf": ".windsurfrules",
        "copilot":  ".github/copilot-instructions.md",
    }
    
    for _, tool := range tools {
        target, ok := symlinkMap[tool]
        if !ok {
            continue
        }
        
        // Create .github directory if needed
        if strings.Contains(target, "/") {
            dir := filepath.Dir(target)
            if err := os.MkdirAll(dir, 0755); err != nil {
                return err
            }
        }
        
        // Create symlink
        if err := os.Symlink("agent.md", target); err != nil {
            if !os.IsExist(err) {
                return err
            }
        }
    }
    
    return nil
}
```

### Build File Content with strings.Builder

```go
func buildAgentMd(sections map[string][]string) string {
    var builder strings.Builder
    
    builder.WriteString("# Agent Configuration\n\n")
    
    for section, rules := range sections {
        builder.WriteString(fmt.Sprintf("## %s\n\n", section))
        for _, rule := range rules {
            builder.WriteString(fmt.Sprintf("- %s\n", rule))
        }
        builder.WriteString("\n")
    }
    
    return builder.String()
}
```

### Safe File Write (Atomic)

```go
func safeWriteFile(path string, content []byte) error {
    // Write to temporary file first
    tmpPath := path + ".tmp"
    if err := os.WriteFile(tmpPath, content, 0644); err != nil {
        return err
    }
    
    // Rename (atomic on Unix)
    return os.Rename(tmpPath, path)
}
```

## Best Practices

1. **Always close files** - Use `defer file.Close()`
2. **Check errors** - Never ignore file operation errors
3. **Use filepath.Join** - For cross-platform paths
4. **Use os.ReadFile/WriteFile** - For small files
5. **Use bufio** - For large files or line-by-line reading
6. **Use strings.Builder** - For efficient string concatenation
7. **Check os.IsNotExist** - To distinguish missing files from other errors

## References

- **os**: https://pkg.go.dev/os
- **path/filepath**: https://pkg.go.dev/path/filepath
- **strings**: https://pkg.go.dev/strings
- **fmt**: https://pkg.go.dev/fmt
- **io**: https://pkg.go.dev/io
