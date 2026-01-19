# agmd CLI Implementation Plan

**Goal:** Build `agmd` as a universal agent configuration manager with built-in safe tooling.

**Based on:** Analysis of 13 agent files across 11 repositories + design doc

---

## Phase 0: Project Setup ✅ (Partially Complete)

### Completed
- [x] Analysis of existing agent files
- [x] Design document (AGMD_DESIGN.md)
- [x] Universal shared AGENTS.md
- [x] Project structure exists (Go-based)

### TODO
- [ ] Review and finalize design
- [ ] Set up CI/CD pipeline
- [ ] Create project README
- [ ] Initialize go modules structure

---

## Phase 1: Core Config System (MVP)

**Timeline:** 2-3 weeks
**Goal:** Basic config management with inheritance

### 1.1 Data Structures & Parsing

**Files to create:**
```
pkg/config/
├── types.go          # Core data structures
├── parser.go         # YAML front matter + markdown parser
├── resolver.go       # Inheritance resolution
├── validator.go      # Config validation
└── merger.go         # Merge layers (universal → profile → project)
```

**Key Types:**
```go
type AgmdConfig struct {
    Version  string   `yaml:"version"`
    Shared   string   `yaml:"shared"`     // Path or URL to shared config
    Profiles []string `yaml:"profiles"`   // Profile names to inherit
    Overrides map[string]interface{} `yaml:"overrides"`
    Sections []string `yaml:"sections"`
}

type AgentFile struct {
    Frontmatter AgmdConfig
    Content     string    // Markdown content
    Path        string    // File location
}

type ResolvedConfig struct {
    Shared   *AgentFile
    Profiles []*AgentFile
    Project  *AgentFile
    Merged   string       // Final merged markdown
}
```

**Deliverables:**
- [ ] Parse YAML front matter from markdown
- [ ] Parse markdown content
- [ ] Resolve file paths (~/agmd/..., ./..., URLs)
- [ ] Load and cache shared config
- [ ] Load and cache profiles
- [ ] Merge configs in correct order (universal → profile → project)

### 1.2 CLI Commands (Basic)

**Commands to implement:**

```bash
agmd init [--profile typescript] [--profile node-cli]
```
- Create `.agmd.md` in current directory
- Add YAML frontmatter with specified profiles
- Generate project-specific sections (structure, build commands)
- Interactive mode if no profiles specified

```bash
agmd show [--merged] [--json]
```
- Show current project config
- With `--merged`: show fully resolved config from all layers
- With `--json`: output as JSON for programmatic use

```bash
agmd validate [file]
```
- Validate YAML front matter
- Check inheritance chain (can all files be resolved?)
- Lint for common issues (missing sections, invalid overrides)
- Exit 0 if valid, exit 1 if invalid

**Files to create:**
```
cmd/
├── init.go
├── show.go
└── validate.go
```

**Deliverables:**
- [ ] `agmd init` creates valid config file
- [ ] `agmd show` displays config correctly
- [ ] `agmd show --merged` resolves inheritance
- [ ] `agmd validate` catches common errors

### 1.3 Built-in Assets

**Embed at build time:**

```go
//go:embed assets/shared/AGENTS.md
var universalShared string

//go:embed assets/profiles/*.md
var profiles embed.FS
```

**Files to create:**
```
assets/
├── shared/
│   └── AGENTS.md       # UNIVERSAL_SHARED_AGENTS.md
└── profiles/
    ├── typescript.md
    ├── swift.md
    └── node-cli.md
```

**Initial Profiles to create:**

1. **typescript.md** - TypeScript-specific rules
2. **swift.md** - Swift-specific rules
3. **node-cli.md** - Node CLI tool patterns

**Deliverables:**
- [ ] Universal shared config embedded
- [ ] 3 initial profiles created and embedded
- [ ] Assets accessible via `agmd` binary

### 1.4 Installation & Setup

```bash
agmd setup
```
- Create `~/.agmd/` directory structure
- Extract embedded assets to `~/.agmd/`
- Create default `config.yml`
- Print success message with next steps

**Directory structure created:**
```
~/.agmd/
├── shared/
│   └── AGENTS.md
├── profiles/
│   ├── typescript.md
│   ├── swift.md
│   ├── node-cli.md
│   └── custom/      # User-created profiles
├── templates/
│   └── sections/    # Empty initially
├── logs/
│   └── runs/        # Empty initially
└── config.yml
```

**config.yml template:**
```yaml
version: 1.0.0

defaults:
  profiles: []
  editor: ${EDITOR:-vim}

git:
  conventional-commits: true
  co-author: "Claude Sonnet 4.5 <noreply@anthropic.com>"

run:
  tmux-enabled: true
  log-dir: ~/.agmd/logs
  default-timeout: 10m

commit:
  validate-format: true
  require-scope: false
  allowed-types: [feat, fix, chore, docs, test, refactor, build, ci, style, perf]
```

**Deliverables:**
- [ ] `agmd setup` creates directory structure
- [ ] Config file created with sensible defaults
- [ ] Assets extracted correctly

---

## Phase 2: Safe Tooling

**Timeline:** 2-3 weeks
**Goal:** Replace common script patterns with safe built-in tools

### 2.1 Safe Command Runner (`agmd run`)

**Purpose:** Run commands safely with logging, git checks, and tmux support

**Implementation:**
```
pkg/runner/
├── executor.go   # Core command execution
├── tmux.go       # tmux integration
├── logger.go     # Command logging
└── git.go        # Git status checks
```

**Features:**

```bash
# Basic execution with logging
agmd run "pnpm build" --log

# Check git is clean before running
agmd run "pnpm deploy" --check-git-clean

# Run in background (tmux)
agmd run "pnpm dev" --bg --session myapp-dev

# Timeout support
agmd run "pnpm test" --timeout 5m

# Retry logic
agmd run "npm install" --retry 3 --retry-delay 5s

# List running sessions
agmd run list

# Attach to session
agmd run attach myapp-dev

# Kill session
agmd run kill myapp-dev
```

**Deliverables:**
- [ ] Basic command execution
- [ ] Logging to `~/.agmd/logs/`
- [ ] Git status check (warn if dirty)
- [ ] tmux integration (detect, create session, attach)
- [ ] Timeout support
- [ ] Retry logic with backoff
- [ ] Session management (list, attach, kill)

### 2.2 Safe Git Committer (`agmd commit`)

**Purpose:** Scoped git commits with validation

**Implementation:**
```
pkg/git/
├── committer.go   # Main commit logic
├── status.go      # Git status parsing
├── staging.go     # Scoped staging
└── validator.go   # Conventional commit validation
```

**Features:**

```bash
# Commit specific files
agmd commit "feat: add user auth" src/auth/**

# Interactive selection
agmd commit "fix: resolve leak" --interactive

# Validate message format
agmd commit "feat(api): add endpoint" --validate-only

# Dry-run
agmd commit "feat: xyz" --dry-run

# Skip hooks (with confirmation)
agmd commit "chore: update deps" --no-verify

# Co-author
agmd commit "feat: xyz" --co-author "Name <email>"
```

**Commit Message Validation:**
- Type is one of allowed types
- Optional scope in parentheses
- Colon and space after type(scope)
- Summary is concise (<72 chars recommended)
- Optional body separated by blank line

**Deliverables:**
- [ ] Scoped staging (only specified files)
- [ ] Conventional commit validation
- [ ] Interactive file selection
- [ ] Dry-run mode
- [ ] Hook management
- [ ] Co-author support

---

## Phase 3: Advanced Features

**Timeline:** 3-4 weeks
**Goal:** Profile management, templates, documentation tools

### 3.1 Profile Management (`agmd profile`)

**Commands:**

```bash
# List all profiles
agmd profile list

# Show profile content
agmd profile show typescript

# Create new profile
agmd profile create my-stack --extend typescript --interactive

# Edit profile
agmd profile edit my-stack

# Validate profile
agmd profile validate my-stack

# Delete custom profile
agmd profile delete my-stack

# Export profile
agmd profile export my-stack > my-stack.md

# Import profile
agmd profile import my-stack < my-stack.md
```

**Implementation:**
```
pkg/profile/
├── manager.go    # CRUD operations
├── builtin.go    # Built-in profiles (embedded)
├── custom.go     # User profiles (~/.agmd/profiles/custom/)
└── validator.go  # Profile validation
```

**Deliverables:**
- [ ] List profiles (built-in + custom)
- [ ] Show profile content
- [ ] Create custom profiles
- [ ] Edit profiles in $EDITOR
- [ ] Validate profile structure
- [ ] Import/export profiles

### 3.2 Template System (`agmd template`)

**Commands:**

```bash
# List available templates
agmd template list

# Use template (adds section to .agmd.md)
agmd template use ci-github-actions

# Create custom template
agmd template create my-deploy --interactive

# Show template
agmd template show ci-github-actions

# Export template
agmd template export my-deploy > template.json

# Import template
agmd template import < template.json
```

**Templates to create:**

**Section Templates:**
- `ci-github-actions.md` - GitHub Actions CI config section
- `ci-gitlab-ci.md` - GitLab CI config section
- `deployment-docker.md` - Docker deployment notes
- `deployment-k8s.md` - Kubernetes deployment notes
- `security-checklist.md` - Security considerations

**Implementation:**
```
pkg/template/
├── manager.go    # CRUD operations
├── builtin.go    # Built-in templates (embedded)
├── applicator.go # Apply template to config file
└── custom.go     # User templates (~/.agmd/templates/)

assets/templates/
└── sections/
    ├── ci-github-actions.md
    ├── deployment-docker.md
    └── security-checklist.md
```

**Deliverables:**
- [ ] 5+ built-in section templates
- [ ] Template listing
- [ ] Apply template to project config
- [ ] Create custom templates
- [ ] Import/export templates

### 3.3 Documentation Tools (`agmd docs`)

**Commands:**

```bash
# List all docs with metadata
agmd docs list

# Generate docs index
agmd docs index --format markdown

# Update front-matter
agmd docs update README.md --set updated=2026-01-19

# Validate doc links
agmd docs validate --check-links

# Generate API docs from code (language-specific)
agmd docs generate --source src/ --output docs/api/
```

**Implementation:**
```
pkg/docs/
├── scanner.go     # Find markdown files
├── metadata.go    # Extract/update front matter
├── linker.go      # Validate links
└── generator.go   # Generate indexes
```

**Deliverables:**
- [ ] List docs with metadata
- [ ] Generate doc index
- [ ] Update front-matter
- [ ] Validate links (basic check)

---

## Phase 4: Polish & Distribution

**Timeline:** 2-3 weeks
**Goal:** Production-ready, distributable tool

### 4.1 Utilities & Diagnostics

**Commands:**

```bash
# Self-update
agmd upgrade [--check]

# Diagnose issues
agmd doctor

# Config management
agmd config get git.conventional-commits
agmd config set git.conventional-commits false

# Version info
agmd version [--verbose]
```

**`agmd doctor` checks:**
- agmd installation location
- Config file validity
- Shared config accessible
- Profiles loadable
- Git installed
- tmux available (optional)
- Editor configured

**Deliverables:**
- [ ] Self-update mechanism
- [ ] Comprehensive diagnostics
- [ ] Config get/set commands
- [ ] Detailed version info

### 4.2 Documentation

**Create:**
- [ ] README.md with quick start
- [ ] CONTRIBUTING.md
- [ ] Full command reference
- [ ] Profile creation guide
- [ ] Template creation guide
- [ ] Migration guide (from steipete-style configs)
- [ ] FAQ

**Website (future):**
- https://agmd.dev
- Getting started guide
- Command reference
- Profile/template gallery
- Community contributions

### 4.3 Distribution

**Homebrew Formula:**
```ruby
class Agmd < Formula
  desc "Universal agent configuration manager"
  homepage "https://agmd.dev"
  url "https://github.com/yourusername/agmd/archive/v1.0.0.tar.gz"
  sha256 "..."

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args
  end

  test do
    system "#{bin}/agmd", "version"
  end
end
```

**Installation methods:**
```bash
# Homebrew (macOS/Linux)
brew install agmd

# Go install
go install github.com/yourusername/agmd@latest

# Direct download
curl -fsSL https://agmd.dev/install.sh | sh

# From source
git clone https://github.com/yourusername/agmd
cd agmd
make install
```

**Deliverables:**
- [ ] Homebrew formula
- [ ] Installation script
- [ ] Release automation (GitHub Actions)
- [ ] Binary distribution for major platforms

### 4.4 Testing

**Unit Tests:**
- [ ] Config parsing
- [ ] Inheritance resolution
- [ ] Validation logic
- [ ] Git operations (mocked)
- [ ] Template application

**Integration Tests:**
- [ ] Full command flows
- [ ] Real git operations (in temp repos)
- [ ] Profile loading
- [ ] Template usage

**E2E Tests:**
- [ ] Initialize new project
- [ ] Show merged config
- [ ] Run commands safely
- [ ] Create commits

**Test Coverage Goal:** >80%

---

## Phase 5: Community & Ecosystem (Future)

**Timeline:** Ongoing

### 5.1 Community Profiles

**Create public registry:**
- Profiles contributed by community
- Rating/download counts
- Search and discovery

**Popular profiles to add:**
- `react` - React-specific patterns
- `vue` - Vue.js patterns
- `python-django` - Django web apps
- `rust-cli` - Rust CLI tools
- `go-web` - Go web services
- etc.

### 5.2 IDE Integration

**VSCode Extension:**
- Syntax highlighting for .agmd.md
- Inline validation
- Quick actions (add section, use template)
- Config preview

**Other IDEs:**
- JetBrains plugin
- Vim/Neovim plugin

### 5.3 Agent Runtime Integration

**Goal:** Make .agmd.md a standard that agents adopt

**Integrations:**
- Claude Code (built-in)
- Cursor AI
- Aider
- Continue.dev
- etc.

---

## Success Metrics

### Phase 1 Success
- [ ] Can initialize project in <30 seconds
- [ ] Can show merged config in <1 second
- [ ] Inheritance resolution works correctly

### Phase 2 Success
- [ ] Can run commands safely with logging
- [ ] Can commit with scoped staging
- [ ] Conventional commit validation works

### Phase 3 Success
- [ ] 10+ built-in profiles
- [ ] 10+ built-in templates
- [ ] Profile creation is intuitive

### Phase 4 Success
- [ ] Installable via Homebrew
- [ ] Documentation complete
- [ ] >80% test coverage
- [ ] Used in 10+ real projects

---

## Current agmd Repository Integration

### Existing Project Structure

The `agent-md` repo already exists with:
- Go project structure
- Some documentation
- Config management foundations

### Integration Strategy

**1. Keep existing code that's useful:**
- Build system
- Project structure
- Any config parsing logic

**2. Add new packages:**
```
pkg/
├── config/      # NEW: Config parsing & resolution
├── runner/      # NEW: Safe command runner
├── git/         # NEW: Git operations
├── profile/     # NEW: Profile management
├── template/    # NEW: Template system
├── docs/        # NEW: Doc tools
└── [existing packages]
```

**3. Add new commands:**
```
cmd/
├── init.go      # NEW
├── show.go      # NEW
├── validate.go  # NEW
├── run.go       # NEW
├── commit.go    # NEW
├── profile.go   # NEW
├── template.go  # NEW
├── docs.go      # NEW
└── [existing commands]
```

**4. Add assets:**
```
assets/
├── shared/
│   └── AGENTS.md        # NEW: UNIVERSAL_SHARED_AGENTS.md
├── profiles/            # NEW
│   ├── typescript.md
│   ├── swift.md
│   └── node-cli.md
└── templates/           # NEW
    └── sections/
        ├── ci-github-actions.md
        └── deployment-docker.md
```

**5. Update documentation:**
- README.md - Add new features
- Add DESIGN.md
- Add IMPLEMENTATION_PLAN.md (this file)

---

## Next Immediate Steps

### 1. Review & Approve Design (1-2 days)
- [ ] Review AGMD_DESIGN.md
- [ ] Review UNIVERSAL_SHARED_AGENTS.md
- [ ] Review this implementation plan
- [ ] Discuss any concerns/changes

### 2. Create Profiles (2-3 days)
- [ ] Create `assets/profiles/typescript.md`
- [ ] Create `assets/profiles/swift.md`
- [ ] Create `assets/profiles/node-cli.md`

### 3. Implement Phase 1.1 (1 week)
- [ ] Design data structures
- [ ] Implement YAML + markdown parser
- [ ] Implement inheritance resolver
- [ ] Add tests

### 4. Implement Phase 1.2 (1 week)
- [ ] Implement `agmd init`
- [ ] Implement `agmd show`
- [ ] Implement `agmd validate`
- [ ] Add tests

### 5. Implement Phase 1.3-1.4 (3-4 days)
- [ ] Embed assets
- [ ] Implement `agmd setup`
- [ ] Test installation flow

### 6. Dogfood (Ongoing)
- [ ] Use agmd in this project (agent-md)
- [ ] Create .agmd.md for agent-md
- [ ] Use `agmd run` for builds
- [ ] Use `agmd commit` for commits
- [ ] Iterate based on experience

---

## Questions to Resolve

1. **Should profiles be markdown or YAML?**
   - **Recommendation:** Markdown (with optional YAML frontmatter)
   - Keeps everything consistent
   - More readable
   - Can include examples and explanations

2. **Should we support remote shared configs (URLs)?**
   - **Recommendation:** Yes, but validate carefully
   - Allow `https://` URLs in `shared:` field
   - Cache remotely-fetched configs
   - Warn if remote config changes

3. **How to handle profile conflicts?**
   - **Recommendation:** Last profile wins
   - Warn user about conflicts
   - Show which profile's rule won in `--verbose` mode

4. **Should tmux be required?**
   - **Recommendation:** No, optional
   - Detect availability (`which tmux`)
   - Fall back to foreground execution
   - Warn if tmux not available for `--bg`

5. **Version compatibility strategy?**
   - **Recommendation:** Semantic versioning
   - Major version mismatch = error
   - Minor version mismatch = warning
   - Patch version mismatch = ignore

---

## Timeline Summary

- **Phase 1** (MVP): 2-3 weeks
- **Phase 2** (Tooling): 2-3 weeks
- **Phase 3** (Features): 3-4 weeks
- **Phase 4** (Polish): 2-3 weeks
- **Total**: ~10-13 weeks to v1.0.0

With focused work, could compress to ~6-8 weeks.

---

## Done When...

### Phase 1 MVP is done when:
- [ ] Can `agmd init` a new project
- [ ] Can `agmd show --merged` to see effective config
- [ ] Can `agmd validate` config files
- [ ] Universal shared config is embedded
- [ ] 3 profiles are embedded
- [ ] Installation works (`agmd setup`)

### v1.0.0 is done when:
- [ ] All Phase 1-4 deliverables complete
- [ ] Documentation complete
- [ ] Homebrew formula published
- [ ] Used in 5+ real projects
- [ ] Test coverage >80%
- [ ] No known critical bugs

---

**Ready to start?** Begin with profile creation, then Phase 1.1.
