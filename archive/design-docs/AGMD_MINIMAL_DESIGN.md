# agmd - Minimal Design (Config Management Only)

**Focus:** Agent configuration file management with inheritance
**Out of Scope:** Command runners, git tooling (separate CLI tools)

---

## Core Problem

**Agent instruction files are 60-80% duplicated across repositories.**

### The Solution: 3-Layer Inheritance

```
Layer 1: Universal Shared (language-agnostic rules)
    ↓ inherits
Layer 2: Profiles (language/stack-specific rules)
    ↓ inherits
Layer 3: Project Config (project-specific details)
```

---

## What agmd Does

### Primary Functions

1. **Initialize** project config with inheritance
2. **Show** effective config (merged from all layers)
3. **Validate** config structure and inheritance chain
4. **Manage profiles** (create, edit, list)

### What agmd Does NOT Do

❌ Run commands (use separate `safe-run` CLI)
❌ Git operations (use separate `safe-commit` CLI)
❌ Documentation tools (out of scope)
❌ Template system (keep it simple for now)

---

## Architecture

### File Structure

```
~/.agmd/
├── shared/
│   └── base.md              # Universal shared config
├── profiles/
│   ├── typescript.md
│   ├── swift.md
│   ├── go.md
│   └── custom/
│       └── my-profile.md
└── config.yml               # agmd settings
```

### Project Structure

```
my-project/
├── .agmd.md                 # Project config (OR AGENTS.md)
└── .agmd/
    └── cache.json           # Resolved config cache
```

---

## Universal Shared Config Sections

### Core Structure (What Every Project Needs)

Based on analysis, these sections have **75-90% duplication** and should be universal:

#### 1. **Intake & Context** (Duplication: 80%)
- Read project config before starting
- Understand structure before modifying
- Ask questions when unclear

#### 2. **Code Quality** (Duplication: 90%)
- Refactor in place (no V2 files)
- Keep files manageable size (~500-700 LOC)
- Match existing code style

#### 3. **Dependencies** (Duplication: 85%)
- Ask before adding dependencies
- Provide rationale + GitHub URL
- Never swap package managers

#### 4. **Testing** (Duplication: 75%)
- Add tests for bug fixes
- Run full test suite before handoff
- Focus on critical paths

#### 5. **Version Control** (Duplication: 90%)
- Only commit when asked
- Use conventional commit format
- One logical change per commit

#### 6. **Documentation** (Duplication: 80%)
- Update docs when behavior changes
- Only create new docs when requested
- Changelog for user-facing changes

#### 7. **Security** (Duplication: 75%)
- Never commit secrets
- Don't edit node_modules/vendor dirs
- Ask before changing project-wide config

#### 8. **Build & Verification** (Duplication: 70%)
- Run full gate before handoff (format → lint → test → build)
- Surface failures clearly
- Don't stop running dev servers

#### 9. **Multi-Agent Safety** (Duplication: 60%, from clawdis)
- Don't create/drop git stash
- Don't switch branches unless requested
- Scope commits to your changes

---

## Profile Sections

### What Profiles Contain (Language/Stack-Specific)

Profiles **extend** universal config with language-specific details:

#### TypeScript Profile Example

```markdown
# TypeScript Profile

Extends: universal

## Language Specifics
- Strict typing (avoid `any`)
- ESM modules
- Prefer utility types

## Common Tools
- Package manager: pnpm, bun, or npm (project specifies)
- Formatting: Biome (preferred) or Prettier
- Linting: Biome, oxlint, or ESLint
- Testing: Vitest (preferred) or Jest
- Build: tsc, esbuild, or Vite

## File Conventions
- Source: `*.ts`, `*.tsx`
- Tests: `*.test.ts`, `*.e2e.test.ts`
- Config: `tsconfig.json`, `biome.json`, etc.

## Build Pattern (Template)
```bash
pnpm install   # Dependencies
pnpm dev       # Development
pnpm build     # Production build
pnpm test      # Tests
pnpm lint      # Lint & format
```
```

#### Swift Profile Example

```markdown
# Swift Profile

Extends: universal

## Language Specifics
- Swift 6+ with strict concurrency
- Annotations: `Sendable`, `@MainActor`, actors
- Structured concurrency (no raw threads)

## Common Tools
- Formatting: SwiftFormat (4-space indent, 120 col)
- Linting: SwiftLint
- Testing: Swift Testing (preferred) or XCTest
- Build: SwiftPM or Xcode

## File Conventions
- Source: `*.swift`
- Tests: `*Tests.swift`

## Build Pattern (Template)
```bash
swift build              # Build
swift test               # Tests
swiftformat .            # Format
swiftlint lint           # Lint
```
```

---

## Project Config Sections

### What Projects Contain (Unique to Repository)

Projects contain **only project-specific** details:

#### Required Sections

1. **Project Structure** - Directory layout
2. **Build Commands** - Actual commands (not templates)
3. **Special Workflows** - Unique processes (if any)

#### Optional Sections

4. **Deployment** - How to deploy
5. **Special Rules** - Project-specific constraints
6. **Troubleshooting** - Known issues

#### Example Project Config

```yaml
---
agmd:
  version: 1.0.0
  shared: ~/.agmd/shared/base.md
  profiles: [typescript, node-cli]
  overrides:
    code-quality.file-size-limit: 800
---

# Project: my-api

## Project Structure
- `src/` - API implementation
  - `routes/` - Express routes
  - `models/` - Database models
  - `services/` - Business logic
- `tests/` - Vitest tests
- `docs/` - API documentation

## Build Commands
```bash
pnpm install          # Install deps
pnpm dev              # Start dev server (port 3000)
pnpm build            # Build for production
pnpm test             # Run tests
pnpm lint             # Lint & format
pnpm db:migrate       # Run migrations
```

## Deployment
- Platform: Railway
- CI: GitHub Actions (.github/workflows/deploy.yml)
- Staging: Push to `develop` branch
- Production: Push to `main` branch

## Special Rules
- Database migrations must be backward-compatible
- API versioning via `/v1/`, `/v2/` routes
- All endpoints require authentication except `/health`
```

---

## YAML Front Matter Format

### Project Config (.agmd.md)

```yaml
---
agmd:
  version: 1.0.0                    # agmd format version
  shared: ~/.agmd/shared/base.md    # Path to universal config
  profiles:                         # Profiles to inherit (in order)
    - typescript
    - node-cli
  overrides:                        # Override specific rules
    code-quality.file-size-limit: 800
    testing.framework: jest
---

# Project content below
```

### Profile Config

```yaml
---
agmd:
  version: 1.0.0
  extends: universal                # Profiles extend universal
  type: profile
---

# Profile content below
```

### Universal Config

```yaml
---
agmd:
  version: 1.0.0
  type: universal                   # Base config
---

# Universal content below
```

---

## Inheritance Resolution

### Resolution Order (Last Wins)

1. Load **universal** config
2. For each **profile** (in order):
   - Load profile
   - Merge with previous (profile sections override universal)
3. Load **project** config
4. Apply **overrides** from frontmatter
5. Result: **Effective config**

### Example

**Universal says:**
```markdown
## Code Quality
- Keep files under 500 LOC
```

**TypeScript profile says:**
```markdown
## Code Quality
- Keep files under 700 LOC (TypeScript guideline)
```

**Project overrides:**
```yaml
overrides:
  code-quality.file-size-limit: 800
```

**Effective config:**
```markdown
## Code Quality
- Keep files under 800 LOC
```

---

## CLI Commands

### Core Commands (Minimal)

```bash
# Initialize new project
agmd init [--profile typescript] [--profile node-cli]

# Show current config
agmd show

# Show fully merged config (all inheritance resolved)
agmd show --merged

# Validate config
agmd validate

# Edit config in $EDITOR
agmd edit
```

### Profile Management

```bash
# List available profiles
agmd profile list

# Show profile content
agmd profile show typescript

# Create custom profile
agmd profile create my-stack --extend typescript

# Edit profile
agmd profile edit my-stack

# Delete custom profile
agmd profile delete my-stack
```

### Utilities

```bash
# First-time setup
agmd setup

# Show version
agmd version

# Diagnose issues
agmd doctor
```

---

## Implementation Architecture

### Core Go Packages

```
pkg/
├── config/
│   ├── types.go          # Data structures
│   ├── parser.go         # YAML + Markdown parser
│   ├── resolver.go       # Inheritance resolution
│   ├── merger.go         # Merge configs
│   └── validator.go      # Validation
├── profile/
│   ├── manager.go        # Profile CRUD
│   └── builtin.go        # Built-in profiles
└── cli/
    ├── init.go           # Commands
    ├── show.go
    ├── validate.go
    └── profile.go
```

### Key Data Structures

```go
// Front matter
type AgmdFrontmatter struct {
    Version  string            `yaml:"version"`
    Shared   string            `yaml:"shared,omitempty"`    // Universal path
    Profiles []string          `yaml:"profiles,omitempty"`  // Profile names
    Extends  string            `yaml:"extends,omitempty"`   // For profiles
    Type     string            `yaml:"type,omitempty"`      // universal/profile/project
    Overrides map[string]any   `yaml:"overrides,omitempty"` // Override values
}

// Full file
type AgmdFile struct {
    Frontmatter AgmdFrontmatter
    Content     string          // Markdown content
    Path        string          // File location
    Sections    []Section       // Parsed sections
}

// Parsed section
type Section struct {
    Title   string
    Level   int       // Heading level (##, ###)
    Content string    // Section content
    Key     string    // Normalized key (e.g., "code-quality")
}

// Resolved config
type ResolvedConfig struct {
    Universal *AgmdFile
    Profiles  []*AgmdFile
    Project   *AgmdFile
    Merged    *AgmdFile       // Final merged result
}
```

---

## Minimal Universal Config Structure

### What to Include in ~/.agmd/shared/base.md

Based on analysis, include these **9 core sections**:

```markdown
---
agmd:
  version: 1.0.0
  type: universal
---

# Universal Agent Guardrails

## 1. Intake & Context Understanding
[Language-agnostic rules about reading context]

## 2. Code Quality Principles
[Refactor in place, file size, match style]

## 3. Dependency Management
[Ask before adding, provide rationale]

## 4. Testing Philosophy
[Add tests for bugs, run full suite]

## 5. Version Control & Git
[Only commit when asked, conventional commits]

## 6. Documentation
[Update when behavior changes]

## 7. Security & Secrets
[Never commit secrets, don't edit lockfiles]

## 8. Build & Verification
[Run full gate before handoff]

## 9. Multi-Agent Safety
[Don't stash, don't switch branches]
```

### Section Detail Level

**Keep universal sections:**
- ✅ Principles and "why"
- ✅ Rules that apply to ALL languages
- ✅ Behavior expectations
- ❌ NOT specific tools (Prettier, SwiftFormat, etc.)
- ❌ NOT commands (those go in profiles/projects)
- ❌ NOT file paths (those are project-specific)

---

## Profile Structure

### Minimal Profiles to Create

**1. TypeScript Profile**
- Extends universal
- TypeScript-specific rules (strict typing, ESM)
- Common tools (Biome, Vitest, etc.)
- Build pattern template

**2. Swift Profile**
- Extends universal
- Swift-specific rules (concurrency, Sendable)
- Common tools (SwiftFormat, SwiftLint)
- Build pattern template

**3. Go Profile**
- Extends universal
- Go-specific rules (error handling, interfaces)
- Common tools (gofmt, golangci-lint)
- Build pattern template

---

## CLI Workflows

### Initialize New Project

```bash
$ cd my-new-project
$ agmd init --profile typescript

Created .agmd.md with:
- Universal shared config: ~/.agmd/shared/base.md
- Profile: typescript

Next steps:
1. Edit .agmd.md to add project details
2. Run 'agmd show --merged' to see effective config
```

### Show Effective Config

```bash
$ agmd show --merged

# Resolving inheritance...
# ✓ Loaded universal: ~/.agmd/shared/base.md
# ✓ Loaded profile: typescript
# ✓ Loaded project: ./.agmd.md
# ✓ Applied overrides

[Shows fully merged markdown]
```

### Validate Config

```bash
$ agmd validate

✓ YAML frontmatter valid
✓ Universal config found
✓ Profile 'typescript' found
✗ Profile 'node-cli' not found
! Warning: Override 'testing.framework' doesn't match any section

Validation: PASSED with warnings
```

---

## Distribution

### Embedded Assets

```go
//go:embed assets/shared/base.md
var universalShared string

//go:embed assets/profiles/*.md
var profiles embed.FS
```

### Installation

```bash
# Go install
go install github.com/yourusername/agmd@latest

# Setup (first time)
agmd setup
# Creates ~/.agmd/ structure
# Extracts embedded assets
```

---

## MVP Scope

### Phase 1: Core Functionality

**Week 1-2: Config System**
- [ ] YAML frontmatter parser
- [ ] Markdown parser
- [ ] Inheritance resolver
- [ ] Section merger

**Week 2-3: CLI**
- [ ] `agmd init`
- [ ] `agmd show` / `agmd show --merged`
- [ ] `agmd validate`
- [ ] `agmd setup`

**Week 3-4: Profiles**
- [ ] Create 3 built-in profiles (TypeScript, Swift, Go)
- [ ] `agmd profile list/show`
- [ ] Embed assets in binary

**Week 4: Polish**
- [ ] Testing (>70% coverage)
- [ ] Documentation
- [ ] Examples

### Out of MVP Scope

- ❌ Custom profile creation (just manually create in ~/.agmd/profiles/custom/)
- ❌ Template system (too complex for MVP)
- ❌ Remote config fetching (URLs) (local files only)
- ❌ Documentation tools (separate concern)
- ❌ Command runners (separate CLI tools)

---

## Success Criteria

**MVP is successful when:**

1. ✅ Can `agmd init` a new project in <30 seconds
2. ✅ Can `agmd show --merged` to see effective config
3. ✅ Can `agmd validate` config files
4. ✅ Universal config + 3 profiles are embedded
5. ✅ Inheritance resolution works correctly
6. ✅ Installation is simple (`go install` + `agmd setup`)

**v1.0 is successful when:**

1. ✅ Used in 5+ real projects
2. ✅ >70% test coverage
3. ✅ Clear documentation with examples
4. ✅ Community has contributed 2+ custom profiles

---

## File Format Philosophy

### Keep It Simple

**Use Markdown:**
- ✅ Human-readable
- ✅ Familiar to developers
- ✅ Works in any editor
- ✅ Can include code examples
- ✅ Can render on GitHub

**Use YAML Frontmatter:**
- ✅ Machine-readable config
- ✅ Standard (Jekyll, Hugo, etc. use it)
- ✅ Easy to parse
- ✅ Keeps config separate from content

**Don't Use:**
- ❌ JSON (not human-friendly for long content)
- ❌ TOML (less familiar, harder to read)
- ❌ Custom format (reinventing the wheel)

---

## Questions Resolved

### 1. Should profiles be markdown or YAML?
**Answer:** Markdown with YAML frontmatter (consistency)

### 2. How to handle section conflicts?
**Answer:** Last wins (profile → project → overrides)
- Warn user in `agmd validate` if conflicts detected
- Show which config won in `agmd show --merged --verbose`

### 3. Should we support remote configs?
**Answer:** Not in MVP (local files only)
- Add in v1.1 if there's demand

### 4. Where should overrides be specified?
**Answer:** In project's YAML frontmatter only
- Profiles can't override (they extend)
- Only projects can override

### 5. How granular should overrides be?
**Answer:** Section-level keys
- Format: `section-name.property-name: value`
- Example: `code-quality.file-size-limit: 800`
- Keep it simple for MVP

---

## Next Steps

### Immediate (This Week)

1. **Finalize universal config sections** - Review and approve
2. **Create TypeScript profile** - First profile to validate approach
3. **Start implementation** - Begin with parser

### Short-term (Next 2-3 Weeks)

4. **Implement Phase 1** - Config system + CLI
5. **Create Swift & Go profiles**
6. **Write tests**

### Medium-term (Next Month)

7. **Dogfood** - Use in agent-md project
8. **Polish** - Fix bugs, improve UX
9. **Document** - README, examples, guides

---

## Summary

**agmd does ONE thing well: Manage agent config files with inheritance**

### Core Value Proposition

- ❌ NOT a command runner
- ❌ NOT a git tool
- ❌ NOT a template engine
- ✅ **Config management with inheritance**
- ✅ **Eliminate 60-80% duplication**
- ✅ **Simple, focused, reliable**

### Minimal Architecture

```
Universal Config (9 core sections)
    ↓
Profiles (language/stack-specific)
    ↓
Project Config (project-specific)
    ↓
agmd resolves → Effective Config
```

**Result:** Clean, maintainable, DRY agent configurations.
