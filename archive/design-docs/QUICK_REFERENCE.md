# agmd Quick Reference

**Status:** Design Complete, Ready to Implement

---

## 📚 Key Documents

| Document | Purpose | Size | Read Time |
|----------|---------|------|-----------|
| **AGMD_DESIGN.md** | Complete design & architecture | 15KB | 30-40 min |
| **UNIVERSAL_SHARED_AGENTS.md** | Generic shared config (ready to use) | 10KB | 20 min |
| **IMPLEMENTATION_PLAN.md** | Phase-by-phase build plan | 15KB | 30 min |
| **QUICK_REFERENCE.md** | This document | 5KB | 10 min |

---

## 🎯 The Vision

**Problem:** Agent config files are 60-80% duplicated across repos

**Solution:** 3-layer inheritance system managed by `agmd` CLI

```
Universal Shared (agmd managed)
    ↓ inherits
Language/Stack Profiles (user preferences)
    ↓ inherits
Project-Specific (unique to repo)
```

---

## 🏗️ Architecture

### Three Layers

**Layer 1: Universal Shared**
- Location: `~/agmd/shared/AGENTS.md`
- Content: Language-agnostic principles (all projects)
- Examples: "Refactor in place", "Ask before adding deps", "Conventional commits"

**Layer 2: Profiles**
- Location: `~/agmd/profiles/{name}.md`
- Content: Language/stack-specific rules
- Examples: `typescript.md`, `swift.md`, `node-cli.md`

**Layer 3: Project Config**
- Location: `./.agmd.md` or `./AGENTS.md` (in project)
- Content: Project structure, build commands, unique rules
- Example:
  ```yaml
  ---
  agmd:
    shared: ~/agmd/shared/AGENTS.md
    profiles: [typescript, node-cli]
  ---

  # Project Structure
  ...
  ```

---

## 🛠️ Built-in CLI Tools

### 1. Safe Command Runner
```bash
agmd run "pnpm build" --log
agmd run "pnpm dev" --bg --session myapp
agmd run "pnpm test" --timeout 5m --retry 3
```

**Replaces:** steipete's `./runner` pattern
**Features:** Logging, git checks, tmux integration, timeout, retry

### 2. Safe Git Committer
```bash
agmd commit "feat: add auth" src/auth/**
agmd commit "fix: leak" --interactive
agmd commit "chore: deps" --dry-run
```

**Replaces:** steipete's `scripts/committer` pattern
**Features:** Scoped staging, conventional commit validation, co-author

### 3. Profile Manager
```bash
agmd profile list
agmd profile show typescript
agmd profile create my-stack --extend typescript
```

### 4. Template Manager
```bash
agmd template list
agmd template use ci-github-actions
agmd template create my-deploy
```

### 5. Documentation Tools
```bash
agmd docs list
agmd docs index --format markdown
agmd docs validate --check-links
```

---

## 📦 Core Commands

### Project Management

```bash
# Initialize new project
agmd init --profile typescript --profile node-cli

# Show current config
agmd show

# Show fully merged config (all layers)
agmd show --merged

# Validate config
agmd validate

# Edit config
agmd edit
```

### Utilities

```bash
# First-time setup
agmd setup

# Diagnose issues
agmd doctor

# Update agmd
agmd upgrade

# Get/set config
agmd config get git.conventional-commits
agmd config set git.conventional-commits false
```

---

## 🎨 Example Workflow

### 1. Install agmd

```bash
# Homebrew (future)
brew install agmd

# Go install
go install github.com/yourusername/agmd@latest

# First-time setup
agmd setup
```

### 2. Initialize Project

```bash
cd my-new-project
agmd init --profile typescript --profile node-cli
```

Creates `.agmd.md`:
```yaml
---
agmd:
  version: 1.0.0
  shared: ~/agmd/shared/AGENTS.md
  profiles: [typescript, node-cli]
---

# Project: my-new-project

## Project Structure
...

## Build Commands
...
```

### 3. Use Tools

```bash
# Run build safely
agmd run "pnpm build" --log

# Commit changes
agmd commit "feat: add new endpoint" src/api/

# Validate config
agmd validate

# See effective config
agmd show --merged
```

---

## 📊 Implementation Phases

### Phase 1: MVP (2-3 weeks)
- [x] Design complete
- [x] Universal shared config created
- [ ] Config parser (YAML + markdown)
- [ ] Inheritance resolver
- [ ] Commands: `init`, `show`, `validate`
- [ ] Embed assets (shared + 3 profiles)
- [ ] `agmd setup` command

**Deliverable:** Can initialize projects with inherited config

### Phase 2: Safe Tooling (2-3 weeks)
- [ ] `agmd run` - Safe command runner
- [ ] `agmd commit` - Safe git committer
- [ ] Git operations, logging, tmux integration

**Deliverable:** Can run commands and commit safely

### Phase 3: Advanced (3-4 weeks)
- [ ] `agmd profile` - Profile management
- [ ] `agmd template` - Template system
- [ ] `agmd docs` - Documentation tools
- [ ] More profiles (Go, Python, Rust)
- [ ] Built-in templates (CI, deployment)

**Deliverable:** Full-featured config management

### Phase 4: Polish (2-3 weeks)
- [ ] `agmd upgrade` - Self-update
- [ ] `agmd doctor` - Diagnostics
- [ ] Documentation complete
- [ ] Homebrew formula
- [ ] Binary distribution

**Deliverable:** Production-ready, distributable

---

## 🔑 Key Differences from steipete's Approach

| Aspect | steipete | agmd |
|--------|----------|------|
| **Portability** | User-specific (`~/Projects/agent-scripts/`) | Generic (`~/agmd/`) |
| **Tooling** | Manual scripts per repo | Built-in CLI tools |
| **Profiles** | Flat, no separation | Layered (universal → profile → project) |
| **Distribution** | Copy-paste | Single binary with embedded defaults |
| **Customization** | Edit shared file | Override system + templates |
| **Discovery** | Read docs | `agmd template list`, `agmd profile list` |

---

## 🎁 Benefits

### For Individual Developers
- **No duplication** - Write rules once
- **Easy to customize** - Override system
- **Safe by default** - Built-in safety tools
- **Quick setup** - `agmd init` and go

### For Teams
- **Consistency** - Everyone uses same core rules
- **Easy updates** - Update shared file once
- **Shareable** - Export/import profiles
- **Discoverable** - Templates for common needs

### For the Ecosystem
- **Universal format** - Works for any language
- **Agent-agnostic** - Not tied to one AI tool
- **Community-driven** - Share profiles/templates
- **Future-proof** - Designed for evolution

---

## 🚀 Next Steps

### Immediate (This Week)
1. **Review designs** - AGMD_DESIGN.md, UNIVERSAL_SHARED_AGENTS.md
2. **Decide on approach** - Approve or suggest changes
3. **Create profiles** - TypeScript, Swift, Node CLI

### Short-term (Next 2-3 Weeks)
4. **Implement Phase 1** - Config system MVP
5. **Dogfood** - Use in agent-md project
6. **Iterate** - Fix issues, improve UX

### Medium-term (Next 2-3 Months)
7. **Complete Phases 2-4** - Full implementation
8. **Polish** - Documentation, distribution
9. **Release** - v1.0.0 with Homebrew formula

---

## 💬 Design Decisions to Confirm

1. **Profiles format:** Markdown with optional YAML frontmatter ✓
2. **Remote configs:** Support HTTPS URLs with caching ✓
3. **Profile conflicts:** Last profile wins, with warnings ✓
4. **tmux requirement:** Optional, detect and fall back ✓
5. **Versioning:** Semantic versioning with mismatch warnings ✓

---

## 📁 File Organization

### Deliverables Created

```
agent-md/
├── AGMD_DESIGN.md                  # Complete design document
├── UNIVERSAL_SHARED_AGENTS.md      # Ready-to-use shared config
├── IMPLEMENTATION_PLAN.md          # Phase-by-phase plan
├── QUICK_REFERENCE.md              # This file
│
├── CONSOLIDATION_PLAN.md           # Analysis-based plan
├── PROPOSED_SHARED_AGENTS.MD       # steipete-specific (for comparison)
├── EXECUTIVE_SUMMARY.md            # Analysis summary
├── README.md                       # Project overview
│
└── fetched_files/
    ├── ANALYSIS_REPORT.md          # Detailed analysis (55KB)
    └── [14 original files]         # Source material
```

---

## 🎯 Success Criteria

**MVP is successful when:**
- [ ] Can `agmd init` a project in <30 seconds
- [ ] Can `agmd show --merged` in <1 second
- [ ] Can `agmd validate` config files
- [ ] Works on macOS and Linux

**v1.0.0 is successful when:**
- [ ] Used in 5+ real projects
- [ ] >80% test coverage
- [ ] Homebrew installable
- [ ] Documentation complete
- [ ] Community has contributed 3+ profiles

---

## 🤔 FAQ

**Q: How is this different from .editorconfig or similar?**
A: Those configure tools. This configures AI agents' behavior and provides safety tooling.

**Q: Why not just use steipete's approach?**
A: It works but is user-specific. agmd is generic, distributable, and includes tooling.

**Q: Can I use my existing AGENTS.md?**
A: Yes! `agmd validate` works with existing files. Add YAML frontmatter to enable inheritance.

**Q: Do I need to use all the CLI tools?**
A: No, they're optional. Use what's helpful. The config system works standalone.

**Q: What if my project uses multiple languages?**
A: Use multiple profiles: `profiles: [typescript, swift, node-cli]`

**Q: Can I create private profiles?**
A: Yes, create in `~/.agmd/profiles/custom/`. They're never shared unless you export them.

---

## 📞 Questions?

- **Design questions** → Review AGMD_DESIGN.md
- **Implementation details** → Review IMPLEMENTATION_PLAN.md
- **Analysis data** → Review fetched_files/ANALYSIS_REPORT.md

**Ready to build?** Start with Phase 1.1 (data structures & parsing).

---

**Last Updated:** 2026-01-19
**Version:** Design v1.0
**Status:** ✅ Complete, Ready to Implement
