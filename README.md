# agmd

**Stop copy-pasting AI instructions between projects.**

agmd lets you maintain a personal registry of coding rules, workflows, and guidelines that you can mix and match across any project. Write your standards once, use them everywhere.

## The Problem

Every project needs an `AGENTS.md` or `CLAUDE.md` file to guide AI coding assistants. But maintaining these files is painful:

- **Copy-paste hell** - Same rules duplicated across 10+ repos
- **Drift** - Update one file, forget the others
- **No reuse** - Can't easily share best practices between projects
- **Verbose** - Files grow to 500+ lines, hard to scan

## The Solution

agmd introduces a simple two-file system:

```
your-project/
├── directives.md   # What you edit (compact references)
└── AGENTS.md       # What AI reads (expanded content)
```

Your `directives.md` stays clean and scannable:

```markdown
# Project Instructions

## Code Quality
:::include rule:typescript
:::include rule:no-any

## Workflows
:::include workflow:commit
:::include workflow:pr-review

## Guidelines
:::list guideline
code-style
documentation
testing
:::end
```

Run `agmd sync` and it expands to a full `AGENTS.md` with all the content.

## Quick Start

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/GluonGrid/agmd/main/install.sh | bash

# Initialize your personal registry
agmd setup

# In any project
agmd init              # Create directives.md
agmd add rule:typescript
agmd sync              # Generate AGENTS.md
```

## Already Have a CLAUDE.md or AGENTS.md?

Don't start from scratch:

```bash
# For raw/unstructured files (not using agmd yet)
agmd migrate CLAUDE.md

# For agmd-compatible projects (has directives.md)
agmd collect
```

See [Migrating Existing Projects](#migrating-existing-projects) for details.

## How It Works

### 1. Your Personal Registry

agmd creates `~/.agmd/` to store your reusable content:

```
~/.agmd/
├── rules/           # Coding rules (typescript.md, no-console.md, ...)
├── workflows/       # Process workflows (commit.md, deploy.md, ...)
├── guidelines/      # Best practices (code-style.md, testing.md, ...)
└── profiles/        # Project templates (svelte-kit.md, fastapi.md, ...)
```

### 2. Simple Directive Syntax

Reference items with clean, readable directives:

```markdown
# Single item
:::include rule:typescript

# Multiple items
:::list workflow
commit
deploy
release
:::end

# Inline definition (for project-specific content)
:::new rule:custom-auth
Your custom content here
:::end
```

### 3. Sync Everywhere

```bash
agmd sync  # Expands directives.md → AGENTS.md
```

Update a rule in your registry, run `agmd sync` in each project, done.

## Commands

| Command | Description |
|---------|-------------|
| `agmd setup` | Initialize your `~/.agmd/` registry |
| `agmd init [profile:name]` | Create `directives.md` in current project |
| `agmd sync` | Generate `AGENTS.md` from `directives.md` |
| `agmd add type:name` | Add an item to `directives.md` |
| `agmd remove type:name` | Remove an item from `directives.md` |
| `agmd new type:name` | Create a new item in the registry |
| `agmd list` | List all registry items (shows active items) |
| `agmd promote` | Promote `:::new` blocks to the registry |
| `agmd migrate <file>` | Migrate a raw CLAUDE.md/AGENTS.md to agmd format |
| `agmd collect [-f file]` | Collect rules from an agmd project into your registry |

## Migrating Existing Projects

Two commands help you work with existing projects:

| Command | Use when... | Result |
|---------|-------------|--------|
| `migrate` | Project has raw/unstructured CLAUDE.md (not using agmd) | Content → `directives.md` (for organizing) |
| `collect` | Project already uses agmd (has `directives.md`) | Rules → `~/.agmd/` (for reuse) |

### Migrate: For Raw/Unstructured Files

Use `migrate` when a project has a freeform CLAUDE.md or AGENTS.md that doesn't use agmd yet:

```bash
agmd migrate CLAUDE.md
```

This creates a backup, appends content to `directives.md` with `:::new` markers, and opens your editor. Wrap sections with `:::new rule:name` / `:::end`, then run `agmd promote` to save them to your registry.

### Collect: For agmd-Compatible Projects

Use `collect` when a project already uses agmd (has `directives.md` with `:::include` directives):

```bash
agmd collect                    # Collect rules referenced in directives.md
agmd collect --file CLAUDE.md   # Use CLAUDE.md instead of AGENTS.md
```

This parses `directives.md` to find referenced rules and saves them to `~/.agmd/` so you can reuse them in other projects.

### Quick Start (No Organizing)

```bash
cp CLAUDE.md directives.md      # Use existing file as-is
agmd add rule:typescript        # Start adding registry items
agmd sync                       # Generate new AGENTS.md
```

## Profiles

Bootstrap new projects instantly with profiles:

```bash
# Save current project as a template
agmd new profile:svelte-kit

# Use it in a new project
agmd init profile:svelte-kit
```

Profiles are complete `directives.md` templates for specific tech stacks.

## Works With Any AI Assistant

agmd generates standard markdown that works with:

- **Claude Code** (`CLAUDE.md`)
- **Cursor** (`.cursorrules` or project instructions)
- **Windsurf** (project context)
- **GitHub Copilot** (repository context)
- **Any AI** that reads markdown instructions

Use `agmd symlink` to create the appropriate files for your toolchain.

## Example Workflow

```bash
# You have 5 TypeScript projects with similar rules

# 1. Create your master rule
agmd new rule:typescript
# (Opens editor - add your TypeScript standards)

# 2. In each project
cd ~/projects/app-1
agmd init
agmd add rule:typescript
agmd sync

# 3. Later, update the rule
agmd edit rule:typescript

# 4. Sync all projects
for dir in ~/projects/app-*; do
  (cd "$dir" && agmd sync)
done
```

## Registry Organization

Organize with subdirectories:

```bash
agmd new rule:frontend/typescript
agmd new rule:frontend/react
agmd new rule:backend/api-design
```

Reference with full path:

```markdown
:::include rule:frontend/typescript
:::include rule:backend/api-design
```

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/GluonGrid/agmd/main/install.sh | bash
```

### From Source

```bash
go install github.com/GluonGrid/agmd@latest
```

### Manual Download

Download binaries from [Releases](https://github.com/GluonGrid/agmd/releases).

## Configuration

agmd works out of the box with sensible defaults:

- Registry: `~/.agmd/`
- Source file: `directives.md`
- Output file: `AGENTS.md`

## Philosophy

1. **DRY for AI instructions** - Write once, reference everywhere
2. **Human-readable source** - `directives.md` is scannable at a glance
3. **Machine-readable output** - `AGENTS.md` has full context for AI
4. **Personal registry** - Your standards, your way
5. **Simple syntax** - Learn 3 directives: `:::include`, `:::list`, `:::new`

## Roadmap

Planned features for future releases:

- **Agent Skills support** - Integration with the [Agent Skills](https://agentskills.io) specification for portable, reusable AI agent capabilities
- **Multiple output targets** - Support for `directives.test.md` → `TEST.md` and other custom mappings
- **LLM-friendly CLI** - Non-interactive flags (`--no-editor`, `--content`) for AI agents to manage rules programmatically
- **Registry sharing** - Import/export registries, share rule packs with teams

## Contributing

Contributions welcome! Please read the contributing guidelines first.

## License

MIT

---

**agmd** - Your AI instructions, organized.
