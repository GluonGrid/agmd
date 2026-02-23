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
:::use rule:typescript
:::use rule:no-any

## Workflows
:::use workflow:commit
:::use workflow:pr-review

## Guidelines
:::list
guideline:code-style
guideline:documentation
guideline:testing
:::end
```

Run `agmd sync` and it expands to a full `AGENTS.md` with all the content.

## Quick Start

```bash
# Install
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash

# Initialize your personal registry
agmd setup

# In any project
agmd init              # Create directives.md
agmd edit              # Add directives (opens editor)
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
├── profile/         # Project templates (default.md, svelte-kit.md, ...)
└── <your-types>/    # Create any structure you need
    ├── rule/        # e.g., typescript.md, no-console.md
    ├── workflow/    # e.g., commit.md, deploy.md
    ├── prompt/      # e.g., code-review.md
    └── anything/    # Your own categories
```

There are no predefined folders—you create whatever structure makes sense for your workflow. Just use `agmd new type:name` and the type folder is created automatically.

### 2. Simple Directive Syntax

Reference items with clean, readable directives:

```markdown
# Single item
:::use rule:typescript

# Multiple items (mixed types)
:::list
workflow:commit
workflow:deploy
rule:typescript
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

- `:::use` and `:::list` directives expand to full content from registry
- `:::new` blocks must be promoted first with `agmd promote`

Update a rule in your registry, run `agmd sync` in each project, done.

## Commands

| Command | Description |
|---------|-------------|
| `agmd setup` | Initialize your `~/.agmd/` registry |
| `agmd init [profile:name]` | Create `directives.md` in current project |
| `agmd sync` | Generate `AGENTS.md` from `directives.md` |
| `agmd edit [type:name]` | Edit `directives.md` (default) or a registry item |
| `agmd new type:name` | Create a new item in the registry |
| `agmd show type:name` | Display item content (useful for AI assistants) |
| `agmd list [type]` | List registry items (all types or specific type) |
| `agmd promote` | Promote `:::new` blocks to registry (required before sync) |
| `agmd migrate <file>` | Migrate a raw CLAUDE.md/AGENTS.md to agmd format |
| `agmd collect [-f file]` | Collect rules from an agmd project into your registry |
| `agmd task <action>` | Manage project tasks (list, new, show, delete, status, ...) |
| `agmd file <action>` | Manage raw files - scripts, configs (list, new, add, show, delete) |
| `agmd doc <action>` | Manage documentation folders (list, add, show, delete, link, unlink) |
| `agmd git <args>` | Run any git command in `~/.agmd` from anywhere |

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

Use `collect` when a project already uses agmd (has `directives.md` with `:::use` directives):

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
agmd edit              # Add :::use rule:typescript
agmd sync
# → AGENTS.md now has your TypeScript standards

# 3. Use the same rule in project B
cd ~/projects/api-server
agmd init
agmd edit              # Add :::use rule:typescript
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
:::use rule:frontend/typescript
:::use rule:backend/api-design
```

### Reserved Types

Some types have special behavior and dedicated subcommands:

| Type | Command | Description |
|------|---------|-------------|
| `task` | `agmd task` | Project tasks with dependencies |
| `file` | `agmd file` | Raw files without markdown processing |
| `doc` | `agmd doc` | Documentation folders (symlinked) |
| `profile` | `agmd init profile:name` | Project templates |

Use `agmd new type:name` for all other types (rule, workflow, guideline, prompt, etc.).

## Installation

### Quick Install

```bash
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash
```

### From Source

```bash
go install github.com/GluonGrid/agmd@latest
```

### Manual Download

Download binaries from [Releases](https://github.com/GluonGrid/agmd/releases).

### Shell Completion

Enable tab completion for commands, types, and item names:

```bash
# Bash (add to ~/.bashrc)
eval "$(agmd completion bash)"

# Zsh (add to ~/.zshrc)
eval "$(agmd completion zsh)"

# Fish (add to ~/.config/fish/config.fish)
agmd completion fish | source
```

Restart your shell or source the config file to enable completions.

## Configuration

agmd works out of the box with sensible defaults:

- Registry: `~/.agmd/`
- Source file: `directives.md`
- Output file: `AGENTS.md`

## Sync Across Devices

Use `agmd git` to run any git command in `~/.agmd` from anywhere — no need to `cd` first:

```bash
# One-time setup
agmd git init
agmd git remote add origin git@github.com:you/agmd-registry.git
agmd git add -A && agmd git commit -m "init" && agmd git push

# On another machine
git clone git@github.com:you/agmd-registry.git ~/.agmd

# Daily sync
agmd git add -A && agmd git commit -m "sync" && agmd git push
agmd git pull
agmd git status
agmd git log --oneline
```

All flags pass straight through to git — use any git command you know.

## Local Registry (Team Sharing)

Share project-specific rules with your team by committing a `.agmd/` folder:

```bash
# Set up a local registry alongside directives.md
agmd init --local

# Create team-shared rules (goes to .agmd/, not ~/.agmd/)
agmd new rule:our-coding-style --local
agmd new guide:architecture --local

# Commit it with the project
git add .agmd/ && git commit -m "add team rules"
```

**Resolution order** — local takes priority over global:

1. `.agmd/rule/foo.md` — project-local (team-shared, committed to git)
2. `~/.agmd/rule/foo.md` — global (personal, synced via `agmd git`)

Teammates clone the repo and `agmd sync` automatically picks up local rules without any extra setup.

## Personal Overrides (directives.local.md)

For machine-specific or personal directives you don't want committed, create `directives.local.md` alongside `directives.md`:

```bash
# Create your personal overrides (automatically gitignored by agmd init)
touch directives.local.md

# Add personal directives — only you see these
echo ":::use rule:my-local-tool" >> directives.local.md
```

`agmd sync` automatically merges `directives.local.md` into the output if it exists. It's appended after `directives.md`, so your overrides always win. The file is added to `.gitignore` automatically by `agmd init`.

**When to use it:**
- Machine-specific tools or paths only available on your machine
- Personal coding preferences your team hasn't agreed on
- Experimental rules you're testing before proposing to the team
- Rules from your global `~/.agmd` you want active only in this project

## Philosophy

1. **DRY for AI instructions** - Write once, reference everywhere
2. **Scannable source** - `directives.md` is short and readable at a glance
3. **Complete output** - `AGENTS.md` has full expanded context for AI
4. **Personal registry** - Your standards, your way
5. **Simple syntax** - Learn 3 directives: `:::use`, `:::list`, `:::new`

## Task Management

agmd includes project-based task tracking with dependencies:

```bash
# Create tasks with priority and type
agmd task new setup-db --content "Set up database schema"
agmd task new fix-auth -t bug -p 0 --content "Critical auth bug"
agmd task new add-dark-mode -t feature -p 1 --content "Add dark mode"
agmd task new create-api --content "Create endpoints" --blocked-by "setup-db"
agmd task new setup-db --feature auth     # Scope task to a feature

# List tasks (auto-sorted: P0→P4, then ready → blocked → completed)
agmd task list                            # All tasks for current project
agmd task list --type bug --priority 0    # Critical bugs only
agmd task list --feature auth             # Filter by feature
agmd task list --status ready             # Filter by computed status
agmd task list --tree                     # Show dependency tree
agmd task list --all                      # Include completed tasks

# Manage status and dependencies
agmd task status setup-db completed       # Update status
agmd task edit setup-db --priority 0 --type bug  # Change priority/type
agmd task blocked-by create-api setup-db  # Add dependency
agmd task unblock create-api setup-db     # Remove dependency

# View and delete
agmd task show setup-db                   # Show task content
agmd task show --all                      # Show all tasks with content
agmd task delete setup-db --force         # Delete task
agmd task clean                           # Delete all completed tasks
```

**Priority levels:** P0 (critical), P1 (high), P2 (medium/default, hidden), P3 (low), P4 (backlog)
**Task types:** bug, feature, task (default, hidden), chore

Tasks are stored in `.agmd.json` at project root and auto-sorted by priority then status. The `--tree` view groups tasks by feature and shows dependency chains. Use `--feature` to scope tasks to features. The `--status`, `--priority`, and `--type` flags filter the list. Dependencies via `--blocked-by` are validated on creation.

## File Management

Store raw files (scripts, configs, templates) without markdown processing:

```bash
# Add files to registry
agmd file new setup.sh --content "#!/bin/bash\necho hello"
agmd file add ./scripts/deploy.sh           # Copy existing file
agmd file add ./config.json settings.json   # Copy with new name

# List and view
agmd file list                              # List all files
agmd file show setup.sh                     # Display file content

# Delete
agmd file delete setup.sh                   # Delete with confirmation
agmd file delete setup.sh --force           # Skip confirmation
```

Files are stored as-is in `~/.agmd/file/` without frontmatter processing. Use profiles with a `files:` field to copy files into projects on `agmd init`.

## Documentation Folders

Manage multi-file documentation that can be symlinked into projects:

```bash
# Add documentation to registry
agmd doc add ./docs svelte-kit              # Copy docs folder
agmd doc add ./reference api --force        # Overwrite existing

# List and view
agmd doc list                               # List all docs
agmd doc show svelte-kit                    # Show folder contents

# Link into projects (creates symlink)
agmd doc link svelte-kit                    # Creates docs/svelte-kit → ~/.agmd/doc/svelte-kit
agmd doc link svelte-kit ./reference        # Custom destination
agmd doc link svelte-kit --gitignore        # Also add to .gitignore

# Unlink
agmd doc unlink svelte-kit                  # Remove symlink

# Delete from registry
agmd doc delete svelte-kit
```

Docs are stored in `~/.agmd/doc/` as folders. Unlike files (which are copied), docs are symlinked into projects for easy updates.

Use the `:::docs` directive in `directives.md` to automatically list linked documentation:

```markdown
:::docs           # Lists all symlinked docs in ./docs/
:::docs ./ref     # Lists symlinked docs in ./ref/
```

## AI Assistant Integration

agmd is designed to be used by AI coding assistants. All commands support non-interactive modes:

```bash
# Read content without opening editor
agmd show rule:typescript
agmd show rule:typescript --raw    # Include frontmatter
agmd file show script.sh           # Show raw file
agmd doc show my-docs              # Show doc folder contents

# Create items without editor
agmd new rule:test --no-editor
agmd new rule:test --content "# My Rule\nContent here"
echo "# My Rule" | agmd new rule:test
agmd file new script.sh --content "#!/bin/bash"

# Update items without editor
agmd edit rule:test --content "# Updated content"
echo "New content" | agmd edit rule:test

# Task management
agmd task new setup-db --content "Set up database"
agmd task status setup-db completed
agmd task show --all               # Show all tasks with content
```

## Roadmap

Planned features for future releases:

- **Agent Skills support** - Integration with the [Agent Skills](https://agentskills.io) specification for portable, reusable AI agent capabilities
- **Multiple output targets** - Support for `directives.test.md` → `TEST.md` and other custom mappings
- **Registry sharing** - Import/export registries, share rule packs with teams

## Contributing

Contributions welcome! Please read the contributing guidelines first.

## License

MIT

---

**agmd** - Your AI instructions, organized.
