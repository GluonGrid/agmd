# agmd Implementation - Getting Started

**Status:** Ready to implement
**Target:** MVP in 4 weeks

---

## Phase 1: Core Config System (Week 1-2)

### Week 1: Data Structures & Parsing

#### Day 1-2: Project Setup
- [x] Design complete
- [ ] Initialize Go modules
- [ ] Set up project structure
- [ ] Create basic CLI skeleton

#### Day 3-4: Config Parsing
- [ ] YAML frontmatter parser
- [ ] Markdown content parser
- [ ] Section parser (extract ## headings)
- [ ] Tests for parsing

#### Day 5-7: Inheritance Resolution
- [ ] Load universal config
- [ ] Load profiles
- [ ] Merge algorithm (by section title)
- [ ] Override application
- [ ] Tests for resolution

### Week 2: CLI Commands

#### Day 8-10: Core Commands
- [ ] `agmd init`
- [ ] `agmd show`
- [ ] `agmd show --merged`
- [ ] `agmd validate`

#### Day 11-12: Profile Management
- [ ] `agmd profile list`
- [ ] `agmd profile show`

#### Day 13-14: Setup & Utilities
- [ ] `agmd setup`
- [ ] `agmd doctor`
- [ ] `agmd version`

---

## Project Structure

```
agmd/
├── cmd/
│   └── agmd/
│       └── main.go           # CLI entry point
├── pkg/
│   ├── config/
│   │   ├── types.go          # Data structures
│   │   ├── parser.go         # YAML + Markdown parser
│   │   ├── resolver.go       # Inheritance resolution
│   │   ├── merger.go         # Section merging
│   │   ├── validator.go      # Validation
│   │   └── override.go       # Override application
│   ├── profile/
│   │   ├── manager.go        # Profile CRUD
│   │   └── builtin.go        # Built-in profiles
│   ├── cli/
│   │   ├── init.go           # agmd init
│   │   ├── show.go           # agmd show
│   │   ├── validate.go       # agmd validate
│   │   ├── profile.go        # agmd profile
│   │   └── setup.go          # agmd setup
│   └── util/
│       ├── paths.go          # Path resolution
│       └── file.go           # File operations
├── assets/
│   ├── shared/
│   │   └── base.md           # Universal shared
│   └── profiles/
│       ├── typescript.md
│       ├── swift.md
│       └── go.md
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Implementation Order

### 1. Data Structures (pkg/config/types.go)

```go
package config

// AgmdFrontmatter represents the YAML front matter
type AgmdFrontmatter struct {
    Version  string            `yaml:"version"`
    Shared   string            `yaml:"shared,omitempty"`
    Profiles []string          `yaml:"profiles,omitempty"`
    Extends  string            `yaml:"extends,omitempty"`
    Type     string            `yaml:"type,omitempty"` // universal/profile/project
    Overrides map[string]any   `yaml:"overrides,omitempty"`
}

// AgmdFile represents a complete agent config file
type AgmdFile struct {
    Frontmatter AgmdFrontmatter
    Content     string
    Sections    []Section
    Path        string
}

// Section represents a markdown section
type Section struct {
    Title   string  // e.g., "Code Quality Principles"
    Level   int     // 2 for ##, 3 for ###
    Content string  // Everything until next same-level heading
    Key     string  // Normalized: "code-quality-principles"
}

// ResolvedConfig is the result of merging all layers
type ResolvedConfig struct {
    Universal *AgmdFile
    Profiles  []*AgmdFile
    Project   *AgmdFile
    Merged    *AgmdFile
}
```

### 2. Parser (pkg/config/parser.go)

```go
package config

import (
    "bytes"
    "gopkg.in/yaml.v3"
)

// ParseFile reads and parses an AGENTS.md file
func ParseFile(path string) (*AgmdFile, error) {
    // Read file
    // Extract YAML frontmatter
    // Parse markdown sections
    // Return AgmdFile
}

// extractFrontmatter splits file into frontmatter and content
func extractFrontmatter(content []byte) ([]byte, []byte, error) {
    // Look for ---\n...---\n pattern
    // Return YAML bytes and remaining content
}

// parseSections extracts markdown sections
func parseSections(markdown string) ([]Section, error) {
    // Parse ## headings
    // Extract content between headings
    // Return sections array
}
```

### 3. Resolver (pkg/config/resolver.go)

```go
package config

// Resolve loads and merges all config layers
func Resolve(projectPath string) (*ResolvedConfig, error) {
    // 1. Load project AGENTS.md
    // 2. Load shared config (from frontmatter.Shared)
    // 3. Load each profile (from frontmatter.Profiles)
    // 4. Merge in order: universal → profiles → project
    // 5. Apply overrides
    // 6. Return ResolvedConfig
}

// LoadShared loads the universal shared config
func LoadShared(sharedPath string) (*AgmdFile, error) {
    // Resolve path (~ expansion, etc.)
    // Parse file
    // Validate it's type: universal
}

// LoadProfile loads a profile by name
func LoadProfile(name string) (*AgmdFile, error) {
    // Check ~/.agmd/profiles/{name}.md
    // Check ~/.agmd/profiles/custom/{name}.md
    // Check embedded assets
    // Parse and return
}
```

### 4. Merger (pkg/config/merger.go)

```go
package config

// MergeSections merges sections from multiple files
func MergeSections(base, overlay []Section) []Section {
    // For each section in overlay:
    //   Find matching section in base (by title)
    //   If found: replace base section with overlay section
    //   If not found: append to result
    // Return merged sections
}

// ApplyOverrides applies frontmatter overrides to sections
func ApplyOverrides(sections []Section, overrides map[string]any) []Section {
    // For each override:
    //   Parse key (e.g., "code-quality.file-size-limit")
    //   Find matching section
    //   Apply override (text replacement or metadata)
}
```

### 5. CLI Commands (pkg/cli/init.go)

```go
package cli

import (
    "github.com/spf13/cobra"
)

// InitCmd creates agmd init command
func InitCmd() *cobra.Command {
    var profiles []string

    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize AGENTS.md in current directory",
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Check if AGENTS.md already exists
            // 2. Create frontmatter with profiles
            // 3. Add template sections (structure, commands)
            // 4. Write AGENTS.md
            // 5. Print success message
        },
    }

    cmd.Flags().StringSliceVarP(&profiles, "profile", "p", nil, "Profiles to include")

    return cmd
}
```

---

## Quick Start Implementation

### Step 1: Initialize Project

```bash
cd /Users/sky/git/agent-md
mkdir -p cmd/agmd pkg/{config,profile,cli,util} assets/{shared,profiles}

# Initialize Go modules
go mod init github.com/yourusername/agmd
go get github.com/spf13/cobra
go get gopkg.in/yaml.v3
```

### Step 2: Create Main Entry Point

```go
// cmd/agmd/main.go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/yourusername/agmd/pkg/cli"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "agmd",
        Short: "Agent Markdown Manager",
        Long:  "Manage agent configuration files with inheritance",
    }

    // Add commands
    rootCmd.AddCommand(cli.InitCmd())
    rootCmd.AddCommand(cli.ShowCmd())
    rootCmd.AddCommand(cli.ValidateCmd())
    rootCmd.AddCommand(cli.SetupCmd())
    rootCmd.AddCommand(cli.ProfileCmd())
    rootCmd.AddCommand(cli.SymlinkCmd())

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### Step 3: Copy Assets

```bash
# Copy universal shared config
cp UNIVERSAL_SHARED_AGENTS.md assets/shared/base.md

# We'll create profile files next
```

---

## Testing Strategy

### Unit Tests

```go
// pkg/config/parser_test.go
func TestParseFrontmatter(t *testing.T) {
    input := `---
agmd:
  version: 1.0.0
  profiles: [typescript]
---

# Content
`
    file, err := ParseFile("test.md")
    // Assert frontmatter parsed correctly
}

func TestParseSections(t *testing.T) {
    markdown := `
## Section 1
Content 1

## Section 2
Content 2
`
    sections, err := parseSections(markdown)
    // Assert 2 sections found
}
```

### Integration Tests

```go
// pkg/config/resolver_test.go
func TestResolve(t *testing.T) {
    // Create temp dir with test files
    // Create universal, profile, project configs
    // Call Resolve()
    // Assert correct merging
}
```

---

## Build & Run

### Makefile

```makefile
.PHONY: build test install clean

build:
	go build -o bin/agmd ./cmd/agmd

test:
	go test -v ./...

install:
	go install ./cmd/agmd

clean:
	rm -rf bin/

run:
	go run ./cmd/agmd

# Development
dev-init:
	go run ./cmd/agmd init --profile typescript

dev-show:
	go run ./cmd/agmd show --merged
```

### Usage

```bash
# Build
make build

# Run
./bin/agmd init --profile typescript
./bin/agmd show --merged
./bin/agmd validate

# Install globally
make install
agmd --help
```

---

## Next Steps

### Immediate (Today)

1. **Set up project structure**
   ```bash
   mkdir -p cmd/agmd pkg/{config,profile,cli,util}
   go mod init github.com/yourusername/agmd
   ```

2. **Create types.go**
   - Define AgmdFrontmatter
   - Define AgmdFile
   - Define Section
   - Define ResolvedConfig

3. **Create parser.go skeleton**
   - ParseFile function
   - extractFrontmatter function
   - parseSections function

### This Week

4. **Implement parser** (Day 3-4)
5. **Implement resolver** (Day 5-6)
6. **Implement merger** (Day 7)

### Next Week

7. **Implement CLI commands** (Day 8-11)
8. **Add tests** (Day 12-13)
9. **Polish & document** (Day 14)

---

## Success Criteria

**Week 1 Complete:**
- [ ] Can parse YAML frontmatter
- [ ] Can parse markdown sections
- [ ] Can load universal + profiles
- [ ] Can merge configs correctly
- [ ] Tests pass

**Week 2 Complete:**
- [ ] `agmd init` creates valid AGENTS.md
- [ ] `agmd show --merged` displays resolved config
- [ ] `agmd validate` checks config validity
- [ ] `agmd setup` creates ~/.agmd/ structure
- [ ] Manual testing works end-to-end

**MVP Complete (Week 4):**
- [ ] All core commands work
- [ ] 3 profiles created (TS, Swift, Go)
- [ ] Assets embedded in binary
- [ ] Documentation complete
- [ ] Used in at least 1 real project (agent-md itself)

---

## Questions to Resolve

1. **Go module path?**
   - Suggestion: `github.com/yourusername/agmd`
   - Or your actual GitHub username

2. **Testing framework?**
   - Built-in `testing` package is fine for MVP
   - Consider `testify` for assertions later

3. **YAML library?**
   - `gopkg.in/yaml.v3` (recommended, most popular)

4. **Markdown parsing?**
   - Simple regex for `##` headings (good enough for MVP)
   - Consider `goldmark` for full parser later

---

Ready to start? Let's begin with project setup!
