<div align="center">

# agmd

**Stop copy-pasting AI instructions between projects.**

Maintain a personal registry of rules, workflows, and guidelines that you can mix and match across any project. Write your standards once, use them everywhere.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/GluonGrid/agmd.svg)](https://github.com/GluonGrid/agmd/releases)

</div>

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
├── rule/            # Coding rules (typescript.md, no-console.md, ...)
├── workflow/        # Process workflows (commit.md, deploy.md, ...)
├── guideline/       # Best practices (code-style.md, testing.md, ...)
├── profile/         # Project templates (svelte-kit.md, fastapi.md, ...)
└── <custom>/        # Your own categories (prompts/, templates/, ...)
```

You can create any folder structure that makes sense for your workflow.

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

- `:::include` and `:::list` directives expand to full content from registry
- `:::new` blocks must be promoted first with `agmd promote`

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
| `agmd promote` | Promote `:::new` blocks to registry (required before sync) |
| `agmd migrate <file>` | Migrate a raw CLAUDE.md/AGENTS.md to agmd format |
| `agmd collect [-f file]` | Collect rules from an agmd project into your registry |

## Migrating Existing Projects

Two commands help you work with existing projects:

| Command | Use when... | Result |
|---------|-------------|--------|
| `migrate` | Project has raw/unstructured AI instructions (not using agmd) | Content → `directives.md` (for organizing) |
| `collect` | Project already uses agmd (has `directives.md`) | Rules → `~/.agmd/` (for reuse) |

### Migrate: For Raw/Unstructured Files

Use `migrate` when a project has a freeform AI instruction file that doesn't use agmd yet:

```bash
agmd migrate CLAUDE.md              # Creates directives.md and opens editor
agmd migrate CLAUDE.md --force      # Overwrite existing directives.md
```

The command:
1. Creates a backup of your file (`CLAUDE.md.backup`)
2. Copies content to `directives.md` with a guide header
3. Opens your editor to organize with `:::new` markers

Wrap sections you want to reuse:

```markdown
:::new rule:typescript
# TypeScript Standards
Use strict mode. Avoid `any` type.
:::

:::new workflow:deploy
# Deploy Process
Steps for deployment...
:::
```

Then run `agmd promote` to save them to your registry, followed by `agmd sync` to generate AGENTS.md.

### Collect: For agmd-Compatible Projects

Use `collect` when a project already uses agmd (has `directives.md` with `:::include` directives):

```bash
agmd collect                    # Collect from AGENTS.md (default)
agmd collect -f CLAUDE.md       # Collect from CLAUDE.md instead
```

This parses `directives.md` to find referenced items and extracts their expanded content from the output file, saving them to `~/.agmd/` for reuse in other projects.

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
# 1. Create your master TypeScript rule once
agmd new rule:typescript
# (Opens editor - add your TypeScript standards)

# 2. Use it in project A
cd ~/projects/frontend-app
agmd init
agmd add rule:typescript
agmd sync
# → AGENTS.md now has your TypeScript standards

# 3. Use the same rule in project B
cd ~/projects/api-server
agmd init
agmd add rule:typescript
agmd sync
# → Same standards, zero copy-paste

# 4. Update the rule, sync wherever needed
agmd edit rule:typescript
cd ~/projects/frontend-app && agmd sync
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
2. **Scannable source** - `directives.md` is short and readable at a glance
3. **Complete output** - `AGENTS.md` has full expanded context for AI
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
