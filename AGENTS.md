# Agent Instructions

## agmd Quick Reference

This project uses **agmd** to manage AI instructions. The source file is `directives.md` and it generates `AGENTS.md`.

### Directive Syntax

```markdown
:::include type:name     # Include a single item from ~/.agmd/type/name.md
:::list                  # Include multiple items
item1
item2
:::end
:::new type:name         # Define inline content (promote to registry later)
content here...
:::end
:::docs                  # List symlinked documentation in ./docs/
:::docs ./reference      # List symlinked documentation in custom path
```

### Commands for AI Assistants

```bash
# Read content (no editor)
agmd show type:name              # Display item content
agmd show type:name --raw        # Include frontmatter
agmd file show script.sh         # Show raw file content
agmd doc show my-docs            # Show documentation folder contents

# Create items (no editor)
agmd new type:name --no-editor
agmd new type:name --content "# Content here"
agmd new type:name --local       # Create in project-local .agmd/ (team-shared)
echo "Content" | agmd new type:name
agmd file new script.sh --content "#!/bin/bash"

# Update items (no editor)
agmd edit type:name --content "# New content"
echo "Content" | agmd edit type:name

# Move items between registries
agmd mv rule:my-rule --to-global  # local → global
agmd mv rule:my-rule --to-local   # global → local

# Sync
agmd sync                        # Generate AGENTS.md from directives.md
agmd sync --stdout               # Print expanded content to stdout (for hooks)
agmd sync --diff                 # Print only changes since last sync (unified diff)
agmd list                        # List all registry items (local + global annotated)
agmd promote --all               # Promote all :::new blocks to registry

# Migrating existing projects
agmd migrate CLAUDE.md           # Migrate raw/unstructured file → directives.md
agmd migrate CLAUDE.md --force   # Overwrite existing directives.md
agmd collect                     # Collect rules from agmd project into ~/.agmd/
agmd collect -f CLAUDE.md        # Collect from a specific output file

# Task management
agmd task list                   # List tasks for current project
agmd task list --tree            # Tree view grouped by feature
agmd task list --type bug --priority 0   # Filter: critical bugs only
agmd task list --feature auth    # Filter by feature
agmd task list --status ready    # Filter by status
agmd task new setup-db --content "Set up database"
agmd task new fix-auth -t bug -p 0 --feature auth --content "Critical auth bug"
agmd task new api --blocked-by setup-db  # With dependency
agmd task status setup-db completed
agmd task edit setup-db --priority 1 --type feature
agmd task blocked-by api setup-db   # Add dependency
agmd task unblock api setup-db      # Remove dependency
agmd task clean                  # Delete all completed tasks
agmd task show --all             # Show all tasks with content

# File management (raw files without markdown processing)
agmd file list                   # List all files
agmd file add ./script.sh        # Add existing file to registry
agmd file delete script.sh       # Delete file

# Documentation folders (symlinked into projects)
agmd doc list                    # List all documentation
agmd doc add ./docs my-docs      # Add docs folder to registry
agmd doc link my-docs            # Symlink into project as docs/my-docs
agmd doc unlink my-docs          # Remove symlink

# Skills (multi-file capabilities for AI agents)
agmd skill list                  # List installed skills
agmd skill link my-skill         # Link skill into project
agmd skill unlink my-skill       # Remove skill link

# Registry git (runs git in ~/.agmd, from anywhere)
agmd git init
agmd git remote add origin <url>
agmd git add -A && agmd git commit -m "sync" && agmd git push
agmd git pull
agmd git status
agmd git log --oneline

# Local registry (project-scoped, team-shared via git)
agmd init --local                # Create .agmd/ in project root
agmd new rule:my-rule --local    # Create item in local registry (overrides global)
```

### Hook-Based Workflow (File-less)

Instead of generating `AGENTS.md` on disk, serve content via Claude Code hooks:

```json
{
  "hooks": {
    "SessionStart": [{
      "matcher": "",
      "hooks": [{
        "type": "command",
        "command": "agmd sync --stdout"
      }]
    }]
  }
}
```

Add to `.claude/settings.json`. The agent gets fresh context at every session start and after compaction.

- `--stdout`: full expanded content to stdout, updates cache
- `--diff`: unified diff against cache, updates cache
- Cache at `.agmd/cache/last-sync.md` (auto-gitignored)

### Migration: Existing Projects

| Command | Use when | Result |
|---------|----------|--------|
| `agmd migrate <file>` | Project has raw/unstructured AI instructions | Content → `directives.md` (wrap sections with `:::new`, then promote) |
| `agmd collect` | Project already uses agmd (has `directives.md`) | Registry items → `~/.agmd/` for reuse |

Quick migration flow:
```bash
agmd migrate CLAUDE.md   # creates directives.md, opens editor
# wrap sections with :::new ... :::end in editor, then:
agmd promote --all       # saves :::new blocks to ~/.agmd/
agmd sync                # generates AGENTS.md
```

The built-in `managing-agmd` skill provides detailed step-by-step guidance. Link it with:
```bash
agmd skill link managing-agmd
```

### Registry Resolution Order

When looking up items (e.g. `:::use rule:foo`):

1. `.agmd/rule/foo.md` — project-local (committed to project git, team-shared)
2. `~/.agmd/rule/foo.md` — global (personal, synced via `agmd git`)

### Personal Overrides

Create `directives.local.md` alongside `directives.md` for machine-specific or personal
directives that should not be committed. It is automatically appended to `directives.md`
during `agmd sync` and is added to `.gitignore` by `agmd init`.

### Reserved Types

Some types have special subcommands:

- `task` - Project tasks with dependencies, priority (P0-P4), type (bug/feature/task/chore)
- `file` - Raw files without frontmatter (`agmd file ...`)
- `doc` - Documentation folders, symlinked (`agmd doc ...`)
- `skill` - Multi-file AI skills (`agmd skill ...`)
- `profile` - Project templates (`agmd init profile:name`)

### Important

- Edit `directives.md`, not `AGENTS.md` (it gets overwritten on sync)
- Run `agmd sync` after changes to regenerate AGENTS.md
- Use `:::new` for project-specific content, then `agmd promote` to reuse elsewhere
- Types are flexible: use rule, workflow, prompt, guide, or any custom type
- Reserved types (task, file, doc, skill) use their own subcommands, not generic commands
- Commit `.agmd/` to share project rules with teammates
- Create `directives.local.md` for personal/machine-specific directives (gitignored, merged at sync)

## Core Agent Guardrails

## Code Quality Principles

### Refactoring Philosophy

- **Refactor in place - never create duplicate files**
  - ❌ NEVER: `utils-v2.js`, `helpers-new.ts`, `api-fixed.go`, `backup-service.swift`
  - ✅ INSTEAD: Update the canonical file and remove obsolete code entirely
  - If major restructuring is needed, do it incrementally in the same file
  - Delete old code, don't comment it out (git preserves history)

- **Keep files at manageable size**
  - Guideline: ~500-700 lines of code (not a hard limit)
  - When a file grows too large, extract cohesive modules
  - Prefer many small, focused files over few large ones
  - Extract helpers, utilities, or new modules proactively

- **Match existing code style**
  - Study the project's patterns before adding new ones
  - Follow established naming conventions
  - Maintain consistency in structure and organization
  - When in doubt, ask rather than introducing new patterns

### Code Clarity

- **Write self-documenting code**
  - Use descriptive names for variables, functions, and types
  - Prefer clarity over cleverness
  - Complex logic deserves comments explaining "why", not "what"

- **Avoid premature abstraction**
  - Don't create abstractions until patterns emerge
  - Three instances of duplication before abstracting
  - Keep it simple - add complexity only when needed

- **Type safety (when applicable)**
  - Use strong typing when the language supports it
  - Avoid dynamic types (`any`, `interface{}`, `object`) unless necessary
  - Leverage type systems to catch errors early

## Documentation

### When to Update Documentation

- **Update docs when behavior changes**
  - Keep README accurate
  - Update API documentation
  - Update configuration examples
  - Update diagrams if they exist

- **Only create new docs when requested**
  - Don't create redundant documentation
  - Edit existing docs rather than creating new ones
  - Ask before creating new doc files

### Changelog Guidelines

- **Add entries for user-facing changes**
  - New features
  - Bug fixes
  - Breaking changes
  - Deprecations

- **Don't add entries for**
  - Internal refactoring (unless it affects performance/behavior)
  - Test additions (unless they reveal a user-facing issue)
  - Dependency updates (unless they affect users)
  - Typo fixes in code comments

- **Format consistently**
  - Follow existing changelog format
  - Include issue/PR references when relevant
  - Thank contributors

## Operational Learnings Pattern

**Purpose:** Capture hard-won knowledge and troubleshooting insights

**When to add:**

- After solving a tricky bug
- After discovering a non-obvious solution
- After learning something valuable about the project

**Example:**

```markdown
## Operational Learnings

### Database Connection Timeouts

**Problem:** Intermittent connection timeouts in production
**Root Cause:** Connection pool exhaustion during traffic spikes
**Solution:** Increased pool size + connection timeout
**Lesson:** Always load test with production-like traffic patterns
```

**Benefits:**

- Prevents repeated mistakes
- Speeds up onboarding
- Documents tribal knowledge

## Development

# agmd CLI Integration Test Patterns

## Test Infrastructure

### In-Process Testing
Tests call `cmd.ExecuteArgs(args)` directly — no subprocess, no binary build required.
Output is captured via `os.Pipe()` redirect of stdout/stderr.

### Key Helpers (tests/integration/helpers_test.go)
```go
run(t, "cmd", "arg")          // run, return output (ignores error)
runE(t, "cmd", "arg")         // run, return (output, error)
setup(t)                       // fresh AGMD_HOME registry, returns path
chdir(t)                       // temp CWD, auto-restored on cleanup
fileExists(path)               // uses os.Lstat — detects symlinks too
readFile(t, path)              // fatal if missing
```

### Flag Reset (cmd/root.go: ExecuteArgs)
Cobra does NOT reset flag values between Execute() calls.
`resetFlags(rootCmd)` is called before each ExecuteArgs to reset all
flags to their DefValue and clear f.Changed — prevents cross-test pollution
when global vars are shared between subcommands (e.g. docForce, taskForce).

## Writing New Tests

### Pattern
```go
func TestX_YScenario(t *testing.T) {
    registry := setup(t)   // fresh AGMD_HOME
    dir := chdir(t)        // fresh CWD (required for path-sensitive commands)

    run(t, "cmd", "subcmd", "--flag")

    if !fileExists(filepath.Join(dir, ".agents", "skills", "foo")) {
        t.Error("expected foo to exist")
    }
}
```

### Commands Requiring --force
These commands prompt for confirmation; always pass --force in tests:
- `delete <type:name> --force`
- `doc delete <name> --force`
- `file delete <name> --force`
- `task delete <name> --force`
- `task clean --force`
- `skill delete <name> --force`

### Commands Requiring EDITOR=true
Commands that open an editor must set `t.Setenv("EDITOR", "true")` to avoid blocking:
- `migrate`
- `new` (without --no-editor or --content)

### Symlinks and fileExists
fileExists uses os.Lstat to detect symlinks even when the target doesn't exist.
This is critical for copilot (`.github/copilot-instructions.md → AGENTS.md`)
where the target is relative and doesn't exist in a fresh test directory.

## Bugs Found During Test Writing
- `collect.go writeItemToRegistry`: missing MkdirAll for known types — fixed
- `task.go`: priority omitted from YAML when == 2 (default), causing P2 tasks
  to be read back as P0. Fixed by always writing priority field.
- Cobra flag leakage between in-process test calls — fixed via resetFlags()
- `fileExists` using os.Stat follows symlinks — fixed to use os.Lstat
- `autoSyncRegistryFilenames` (cmd/autosync.go) used `filepath.Walk` recursively,
  causing it to rename `skill/managing-agmd/SKILL.md` → `skill/managing-agmd.md`
  because `relPath = "managing-agmd/SKILL" != meta.Name = "managing-agmd"`.
  Result: `sync` destroyed SKILL.md from skill packages on every run!
  Fixed by using os.ReadDir to process only top-level .md files (not subdirs).

## Workflow Tests (tests/integration/workflow_test.go)
Sequential multi-step tests that simulate real user flows:
- TestWorkflow_NewProject: setup → init → new rule → sync
- TestWorkflow_Migration: write CLAUDE.md → migrate → sync
- TestWorkflow_DocLinkAndSync: init → doc add → doc link → sync
- TestWorkflow_CollectAndPromote: collect → promote → verify registry
- TestWorkflow_SkillLinkAndSync: skill link → sync → verify AGENTS.md

### Key note: setup() does not save registry path
`setup(t)` sets AGMD_HOME but doesn't return it in workflow tests.
If you need to check registry contents, save the return value: `registry := setup(t)`

## Skills available

<available_skills>
  <skill>
    <name>managing-agmd</name>
    <description>Manages agmd projects: initializing registries, migrating existing AI instruction files (CLAUDE.md, AGENTS.md, .cursorrules) to directives.md format, managing reusable rules/workflows/guides, syncing output, and tracking project tasks. Use when the user mentions agmd, directives.md, wants to migrate AI instructions, organize agent rules, or manage project tasks with agmd.</description>
    <location>~/git/agent-md/.claude/skills/managing-agmd/SKILL.md</location>
  </skill>
</available_skills>

