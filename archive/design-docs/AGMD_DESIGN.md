# agmd CLI Tool Design

**Vision:** A universal agent configuration framework with inheritance, CLI tooling, and easy customization.

**Problem:** Agent instruction files contain 60-80% duplicated content. steipete's approach with `~/Projects/agent-scripts/AGENTS.MD` is good but too specific to one user's workflow.

**Solution:** `agmd` - A CLI tool that manages generic, inheritable agent configurations with built-in safe tooling.

---

## 🎯 Core Concept

### Three-Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Layer 1: UNIVERSAL SHARED (agmd managed)                    │
│ ~/agmd/shared/AGENTS.md                                     │
│ • Generic agent guardrails applicable to ALL projects       │
│ • No user-specific assumptions                              │
│ • Language-agnostic core principles                         │
└─────────────────────────────────────────────────────────────┘
                              ↓ inherits
┌─────────────────────────────────────────────────────────────┐
│ Layer 2: USER/STACK PROFILES (agmd managed)                 │
│ ~/agmd/profiles/{typescript,swift,go,python,...}.md         │
│ • Language/stack-specific extensions                        │
│ • User preferences (e.g., prefer Biome over Prettier)       │
│ • Common patterns across YOUR projects                      │
└─────────────────────────────────────────────────────────────┘
                              ↓ inherits
┌─────────────────────────────────────────────────────────────┐
│ Layer 3: PROJECT-SPECIFIC (in each repo)                    │
│ ./.agmd.md or ./AGENTS.md                                   │
│ • Project structure                                         │
│ • Build commands                                            │
│ • Unique workflows                                          │
└─────────────────────────────────────────────────────────────┘
```

### Inheritance Declaration

**In a project's AGENTS.md:**
```markdown
---
agmd:
  shared: ~/agmd/shared/AGENTS.md
  profiles: [typescript, node-cli]
  version: 1.0.0
---

# Project-Specific Config

## Project Structure
...
```

**Or using agmd CLI:**
```bash
# Initialize in a project
agmd init --profile typescript --profile node-cli

# Add custom section
agmd add section "Deployment" --template kubernetes

# Validate against shared rules
agmd validate

# Show effective config (merged)
agmd show --merged
```

---

## 📋 Universal Shared AGENTS.md Structure

### Core Sections (Required in Universal Shared)

These are **language-agnostic** and apply to **all** agent workflows:

```markdown
# Universal Agent Guardrails

## 1. Intake & Context
- Read local agent config before starting
- Understand project structure before modifying
- Ask questions when requirements unclear

## 2. Code Quality Principles
- Refactor in place (never create V2/backup files)
- Keep files manageable size (~500-700 LOC guideline)
- Match existing code style before adding patterns
- Write clear, self-documenting code

## 3. Dependency Management
- Ask before adding dependencies (provide rationale + URL)
- Never swap package managers without approval
- Use exact versions for patched/vendored dependencies

## 4. Testing Philosophy
- Add/extend tests for bug fixes
- Run full test suite before handoff
- Prioritize critical paths and edge cases

## 5. Version Control
- Only commit when explicitly requested
- Use conventional commit format: type(scope): summary
- One logical change per commit
- Never delete unfamiliar files without checking

## 6. Documentation
- Update docs when behavior changes
- Only create new docs when requested
- Keep changelog for user-facing changes only

## 7. Security
- Never commit secrets (use env vars/keychain)
- Don't edit dependency lockfiles manually
- Ask before changing project-wide config

## 8. Build & Verification
- Run full gate before handoff (lint → test → build)
- Surface failures clearly with exact output
- Don't stop running dev servers/watchers

## 9. Multi-Agent Safety
- Don't create/apply/drop git stash entries
- Don't switch branches unless requested
- Scope commits to your changes only
- List unrecognized files at end if relevant
```

---

## 🎨 Profile System

### Language/Stack Profiles

**~/agmd/profiles/typescript.md**
```markdown
# TypeScript Profile

Extends: universal

## Stack Specifics
- Strict typing (avoid `any`)
- Package manager: respect declared (pnpm/bun/npm)
- Testing: Vitest (preferred) or Jest
- Formatting: Biome (preferred) or Prettier

## File Naming
- Source: `*.ts`, `*.tsx`
- Tests: `*.test.ts`, `*.e2e.test.ts`
- Config: `tsconfig.json`, `biome.json`, etc.

## Build Commands (Template)
```bash
pnpm install   # Dependencies
pnpm dev       # Development
pnpm build     # Production build
pnpm test      # Tests
pnpm lint      # Lint & format
```

## Common Tools
- Type checking: `tsc --noEmit`
- Bundling: esbuild, Vite, or tsup
```

**~/agmd/profiles/swift.md**
```markdown
# Swift Profile

Extends: universal

## Stack Specifics
- Swift 6+ with strict concurrency
- Annotations: Maintain `Sendable`, `@MainActor`, actors
- Formatting: SwiftFormat (4-space indent, 120 col width)
- Linting: SwiftLint

## File Naming
- Source: `*.swift`
- Tests: `*Tests.swift`

## Build Commands (Template)
```bash
swift build              # Build
swift test               # Tests
swiftformat .            # Format
swiftlint lint           # Lint
```

## Testing
- Swift Testing (preferred for new code)
- XCTest (legacy support)
- Use `@Test`, `@Suite`, `#expect()` for Swift Testing
```

**~/agmd/profiles/node-cli.md**
```markdown
# Node CLI Tools Profile

Extends: [typescript]

## CLI Specifics
- Entry point: `src/cli.ts` or `bin/`
- Use commander, yargs, or similar for arg parsing
- Provide `--help` and `--version`
- Handle signals gracefully (SIGINT, SIGTERM)

## Distribution
- Package as executable: `#!/usr/bin/env node`
- Specify `bin` field in package.json
- Consider standalone binaries (pkg, bun compile)

## Error Handling
- Exit codes: 0 (success), 1 (error), 2 (usage error)
- Clear error messages to stderr
- Stack traces only with `--verbose`
```

### Custom Profiles (User-Defined)

Users can create their own:

```bash
# Create custom profile
agmd profile create my-fastify-api --extend typescript --extend node-cli

# Edit profile
agmd profile edit my-fastify-api

# List profiles
agmd profile list
```

---

## 🛠️ Built-in CLI Tools

Extract common patterns from agent-scripts as `agmd` subcommands:

### 1. Safe Command Runner (`agmd run`)

**Replaces:** steipete's `./runner` script pattern

**Purpose:** Run commands with safety checks, logging, and tmux integration

```bash
# Run with safety checks
agmd run "pnpm build" --check-git-clean --log

# Run in background (tmux)
agmd run "pnpm dev" --bg --session myapp-dev

# Run with timeout
agmd run "pnpm test" --timeout 5m

# Run with retry
agmd run "npm install" --retry 3

# List running sessions
agmd run list
```

**Features:**
- Check git status before running (prevent dirty state issues)
- Automatic logging to `~/.agmd/logs/`
- tmux integration for long-running tasks
- Timeout support
- Retry logic with exponential backoff
- Session management

### 2. Safe Git Committer (`agmd commit`)

**Replaces:** steipete's `scripts/committer` pattern

**Purpose:** Scoped, safe git commits with conventional format

```bash
# Commit specific files with scoped staging
agmd commit "feat: add user auth" src/auth/**

# Interactive selection
agmd commit "fix: resolve memory leak" --interactive

# Validate message format
agmd commit "invalid message" --validate

# Show what would be committed
agmd commit "feat: xyz" --dry-run

# Skip hooks (with confirmation)
agmd commit "feat: xyz" --no-verify --confirm
```

**Features:**
- Scoped staging (only specified files)
- Conventional commit validation
- Pre-commit hook integration
- Co-author support
- Dry-run mode
- Safety confirmations

### 3. Documentation Helper (`agmd docs`)

**Replaces:** steipete's `docs:list` pattern

**Purpose:** Generate and maintain documentation metadata

```bash
# List all docs with metadata
agmd docs list

# Generate docs index
agmd docs index --format markdown

# Update doc front-matter
agmd docs update README.md --set updated=2026-01-19

# Validate doc links
agmd docs validate --check-links

# Generate API docs from code
agmd docs generate --source src/ --output docs/api/
```

### 4. Config Validator (`agmd validate`)

**Purpose:** Validate agent config structure and inheritance

```bash
# Validate current project config
agmd validate

# Validate specific file
agmd validate ./AGENTS.md

# Check inheritance chain
agmd validate --check-inheritance

# Lint for common issues
agmd validate --lint
```

### 5. Template Manager (`agmd template`)

**Purpose:** Manage reusable config sections

```bash
# List available templates
agmd template list

# Use template
agmd template use ci-github-actions

# Create custom template
agmd template create my-deploy --interactive

# Share template
agmd template export my-deploy > template.json
agmd template import < template.json
```

### 6. Profile Manager (`agmd profile`)

**Purpose:** Manage language/stack profiles

```bash
# List profiles
agmd profile list

# Show profile content
agmd profile show typescript

# Create new profile
agmd profile create my-stack --extend typescript

# Edit profile
agmd profile edit my-stack

# Validate profile
agmd profile validate my-stack
```

---

## 📁 File Structure

### agmd Installation

```
~/.agmd/
├── shared/
│   └── AGENTS.md              # Universal shared config
├── profiles/
│   ├── typescript.md          # Built-in profiles
│   ├── swift.md
│   ├── go.md
│   ├── python.md
│   ├── rust.md
│   └── custom/
│       └── user-defined.md    # User-created profiles
├── templates/
│   ├── sections/
│   │   ├── ci-github.md
│   │   ├── deployment-k8s.md
│   │   └── security.md
│   └── projects/
│       ├── node-cli.md
│       └── web-api.md
├── logs/
│   └── runs/
│       └── 2026-01-19_build.log
└── config.yml                 # User preferences
```

### Project Structure

```
my-project/
├── .agmd.md                   # OR AGENTS.md (agmd recognizes both)
├── .agmd/
│   ├── cache/                 # Resolved config cache
│   └── local-templates/       # Project-specific templates
└── ... (rest of project)
```

---

## 🔧 Configuration Format

### Project Config (.agmd.md or AGENTS.md)

**YAML Front Matter:**
```yaml
---
agmd:
  version: 1.0.0
  shared: ~/agmd/shared/AGENTS.md
  profiles:
    - typescript
    - node-cli
  overrides:
    code-quality.file-size-limit: 800
    testing.framework: jest
  sections:
    - project-structure
    - build-commands
    - deployment
---
```

**Content:**
```markdown
# Project: MyApp

## Project Structure
[project-specific content]

## Build Commands
[project-specific commands]

## Deployment
[project-specific deployment info]
```

### agmd Config (~/.agmd/config.yml)

```yaml
# User preferences
defaults:
  profiles: [typescript]  # Default profiles for new projects
  editor: code           # Editor for `agmd edit`
  git:
    conventional-commits: true
    co-author: "Claude Sonnet 4.5 <noreply@anthropic.com>"

# Tool behavior
run:
  tmux-enabled: true
  log-dir: ~/.agmd/logs
  default-timeout: 10m

commit:
  validate-format: true
  require-scope: false
  allowed-types: [feat, fix, chore, docs, test, refactor, build, ci, style, perf]

# Profile sources
profile-sources:
  - ~/.agmd/profiles
  - ~/.agmd/profiles/custom

# Update behavior
updates:
  check-on-start: true
  auto-update-shared: false  # Prompt before updating shared config
```

---

## 🎨 CLI Interface Design

### Main Commands

```bash
agmd init [options]              # Initialize in current directory
agmd show [--merged] [--json]    # Show effective config
agmd validate [file]             # Validate config
agmd edit [section]              # Edit config in editor

# Subcommands
agmd run <command> [options]     # Safe command runner
agmd commit <message> [files]    # Safe git committer
agmd docs <subcommand>           # Documentation tools
agmd profile <subcommand>        # Profile management
agmd template <subcommand>       # Template management

# Utilities
agmd upgrade                     # Upgrade agmd itself
agmd doctor                      # Diagnose issues
agmd config [get|set] <key>      # Manage config
```

### Example Workflows

**Initialize new TypeScript project:**
```bash
cd my-new-project
agmd init --profile typescript --profile node-cli
# Creates .agmd.md with inheritance
```

**Add deployment section:**
```bash
agmd template use deployment-docker
# Adds section to .agmd.md
```

**Validate before commit:**
```bash
agmd validate
agmd run "pnpm lint && pnpm test" --check-git-clean
agmd commit "feat: add new endpoint" src/api/
```

**Show effective config (inheritance resolved):**
```bash
agmd show --merged
# Shows fully resolved config from all layers
```

---

## 🏗️ Implementation Architecture

### Core Components

```
agmd/
├── cmd/
│   ├── init.go          # Initialize project
│   ├── show.go          # Show config
│   ├── validate.go      # Validate config
│   ├── run.go           # Safe runner
│   ├── commit.go        # Safe committer
│   ├── docs.go          # Docs tools
│   ├── profile.go       # Profile management
│   └── template.go      # Template management
├── pkg/
│   ├── config/
│   │   ├── parser.go    # Parse YAML front matter + markdown
│   │   ├── resolver.go  # Resolve inheritance chain
│   │   ├── validator.go # Validate config structure
│   │   └── merger.go    # Merge layers
│   ├── runner/
│   │   ├── executor.go  # Execute commands safely
│   │   ├── tmux.go      # tmux integration
│   │   └── logger.go    # Logging
│   ├── git/
│   │   ├── committer.go # Git operations
│   │   └── status.go    # Git status checks
│   ├── profile/
│   │   ├── manager.go   # Profile CRUD
│   │   └── builtin.go   # Built-in profiles
│   └── template/
│       ├── manager.go   # Template CRUD
│       └── builtin.go   # Built-in templates
├── assets/
│   ├── shared/
│   │   └── AGENTS.md    # Universal shared (embedded)
│   ├── profiles/        # Built-in profiles (embedded)
│   └── templates/       # Built-in templates (embedded)
└── main.go
```

### Key Design Decisions

**1. YAML Front Matter + Markdown**
- Use front matter for machine-readable config (inheritance, versions)
- Use markdown body for human-readable content (actual guidelines)
- Compatible with existing AGENTS.md files

**2. Layered Inheritance**
- Universal (ships with agmd) → Profile (user/lang-specific) → Project (unique)
- Each layer can override previous layers
- Explicit > Implicit (project overrides beat profiles)

**3. Embedded Defaults**
- Ship with universal shared config embedded
- Ship with common profiles embedded
- Users can override by creating files in ~/.agmd/

**4. Safe by Default**
- All destructive operations require confirmation
- Dry-run mode for testing
- Git status checks before operations
- Validation before applying changes

---

## 📦 Distribution

### Installation

```bash
# Homebrew (future)
brew install agmd

# Go install
go install github.com/yourusername/agmd@latest

# Download binary
curl -fsSL https://agmd.dev/install.sh | sh

# First-time setup
agmd setup
# Creates ~/.agmd/ structure
# Installs shared config and profiles
```

### Bundled Assets

**Universal Shared Config** (embedded at build time):
- Based on analysis but made generic (no steipete-specific stuff)
- Language-agnostic principles
- Update via `agmd upgrade`

**Built-in Profiles** (embedded):
- TypeScript, Swift, Go, Python, Rust, etc.
- Common stacks (node-cli, web-api, etc.)
- Can be customized by creating file in ~/.agmd/profiles/custom/

**Built-in Templates** (embedded):
- CI/CD (GitHub Actions, GitLab CI, etc.)
- Deployment (Docker, K8s, etc.)
- Security checklists
- Testing strategies

---

## 🎯 Benefits Over steipete's Approach

| Aspect | steipete's Approach | agmd Approach |
|--------|---------------------|---------------|
| **Portability** | User-specific path | Generic, works for anyone |
| **Tooling** | Manual scripts per repo | Built-in CLI tools |
| **Profiles** | Flat, no language separation | Layered profiles (TS, Swift, etc.) |
| **Inheritance** | Single shared file | Multi-layer (universal → profile → project) |
| **Customization** | Edit shared file directly | Override system, templates |
| **Distribution** | Copy-paste | Single binary with embedded defaults |
| **Validation** | Manual | Built-in validation |
| **Discovery** | Read docs | `agmd template list`, `agmd profile list` |

---

## 🚀 Phased Implementation Plan

### Phase 1: Core Config System (MVP)
**Goal:** Basic config management and inheritance

- [ ] YAML front matter parser
- [ ] Markdown content parser
- [ ] Inheritance resolver (3-layer system)
- [ ] `agmd init` - Initialize project
- [ ] `agmd show` - Display config (merged)
- [ ] `agmd validate` - Validate config
- [ ] Universal shared config (generic version)
- [ ] 2-3 built-in profiles (TypeScript, Swift)

**Deliverable:** Can initialize projects with inherited config

### Phase 2: Safe Tooling
**Goal:** Replace common script patterns

- [ ] `agmd run` - Safe command runner
  - Basic execution
  - Git status checks
  - Logging
- [ ] `agmd commit` - Safe git committer
  - Scoped staging
  - Conventional commit validation
  - Co-author support

**Deliverable:** Can run commands and commit safely

### Phase 3: Advanced Features
**Goal:** Templates, profiles, docs

- [ ] `agmd profile` - Profile management
- [ ] `agmd template` - Template system
- [ ] `agmd docs` - Documentation tools
- [ ] More built-in profiles (Go, Python, Rust)
- [ ] Built-in templates (CI, deployment, security)

**Deliverable:** Full-featured config management

### Phase 4: Ecosystem
**Goal:** Distribution and community

- [ ] `agmd upgrade` - Self-update
- [ ] `agmd doctor` - Diagnostics
- [ ] Profile/template sharing (import/export)
- [ ] Web documentation
- [ ] Homebrew formula
- [ ] Community profile repository

**Deliverable:** Production-ready, distributable tool

---

## 📊 Success Metrics

**Adoption:**
- Can initialize new project in <30 seconds
- Can validate config in <1 second
- 90% of common patterns available as templates

**Quality:**
- Zero config duplication across projects
- Consistent agent behavior across repos
- Easy to customize without breaking inheritance

**Developer Experience:**
- Single command to set up new project
- Built-in safety prevents common mistakes
- Clear error messages guide users

---

## 💡 Future Possibilities

**1. IDE Integration**
- VSCode extension for editing configs
- Inline validation and autocomplete
- Quick actions (add section, use template)

**2. Agent Runtime Integration**
- Claude Code reads .agmd.md directly
- Other agents (Cursor, Aider) adopt format
- Standard config format emerges

**3. Community**
- Public profile/template registry
- Share custom profiles
- Rate and discover configurations

**4. Analytics**
- Track which sections are most used
- Identify common patterns
- Improve universal shared config

---

## 🎬 Next Steps

1. **Review this design** - Get feedback on architecture
2. **Create universal shared AGENTS.md** - Generic version (no steipete-specific content)
3. **Define core data structures** - Go structs for config, profiles, templates
4. **Implement Phase 1** - Basic config system
5. **Test with real projects** - Dogfood in this repo and others
6. **Iterate based on feedback**

---

**Questions to Resolve:**

1. Should profiles be markdown or YAML? (Leaning markdown for consistency)
2. Should we support JSON output for programmatic use? (Yes, via --json flag)
3. How to handle profile conflicts? (Last wins, with warnings)
4. Should tmux integration be optional/pluggable? (Yes, detect and use if available)
5. Version compatibility strategy? (Semantic versioning, warn on major version mismatch)
