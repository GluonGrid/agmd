# agmd v2 - Enhanced Profiles, Files, Docs & Tasks

## Overview

Enhance agmd to support:
1. **Profiles with file copying** - Bootstrap projects with templates
2. **Files as-is** - Store raw files (not .md wrapped)
3. **Docs linking** - Symlink documentation folders for AI context
4. **Dynamic doc directive** - Auto-expand linked docs in AGENTS.md
5. **Task management** - Project-based task tracking with dependencies
6. **Agent Skills support** - Future integration with agentskills.io

## Reserved Types

The following types have special behavior and cannot be overridden:
- `task/` - Task management (per-project)
- `doc/` - Documentation folders (symlinked)
- `file/` - Raw files (no .md wrapper, no frontmatter)
- `profile/` - Project templates (with `files:` support)
- `guide/` - Guides (default: guide:agmd)

---

## 1. Profile Enhancement

### Current State
- Profiles are single `.md` files in `~/.agmd/profile/`
- Only copy directives.md content

### New Design

**Structure:**
```
~/.agmd/
├── profile/
│   └── python-dev.md
├── file/                       # Files stored AS-IS (no .md wrapper)
│   ├── gitignore/
│   │   ├── python              # Actual .gitignore content
│   │   └── node
│   ├── readme/
│   │   └── minimal.md          # Actual README.md
│   └── python/                 # Folder of config files
│       ├── pyproject.toml
│       ├── ruff.toml
│       └── .python-version
├── doc/                        # Documentation folders for AI context
│   ├── svelte-kit/
│   │   ├── README.md
│   │   └── api.md
│   └── fastapi/
│       └── guide.md
├── rule/
├── workflow/
└── guide/
```

**Profile frontmatter with files:**
```yaml
---
name: python-dev
description: "Python development setup with uv and ruff"
files:
  - file:gitignore/python > .gitignore      # Single file with rename
  - file:readme/minimal.md > README.md
  - file:python/*                            # Copy all files from folder
---

# Agent Instructions
:::include guide:agmd
:::include rule:python
```

### Tasks

- [ ] Update profile frontmatter parsing to support `files:` field
- [ ] Implement file copy syntax: `file:name > destination`
- [ ] Implement folder copy syntax: `file:folder/*` (copies all, keeps names)
- [ ] Update `agmd init profile:name` to copy files
- [ ] Add `agmd new file:name --from ./path` to import files
- [ ] Add `agmd new file:folder --from ./dir/` to import folders
- [ ] Files stored as-is (no frontmatter, no .md conversion)

---

## 2. Docs Linking

### Concept
Store documentation folders in registry, symlink into projects for AI context.

**Use cases:**
- Framework docs (svelte-kit, fastapi, etc.)
- Internal team docs
- API references

### Commands

```bash
# Create doc folder
agmd new doc:svelte-kit                    # Creates folder, opens to add files
agmd new doc:svelte-kit --from ./docs/     # Import existing folder

# Link into project
agmd link doc:svelte-kit                   # Symlinks to ./docs/svelte-kit
agmd link doc:svelte-kit > ./references    # Custom destination

# Auto-add to .gitignore (optional)
agmd link doc:svelte-kit --gitignore

# Unlink
agmd unlink doc:svelte-kit

# List linked docs in current project
agmd link --list
```

### Tasks

- [ ] Support `doc/` as folder-based type (not single .md files)
- [ ] Add `agmd link doc:name` command
- [ ] Add `agmd unlink doc:name` command
- [ ] Add `--gitignore` flag to auto-add to .gitignore
- [ ] Add `--from` flag to import existing folder
- [ ] Track linked docs in `.agmd.json` or similar

---

## 3. Dynamic Doc Directive

### Concept
A special directive that auto-expands to show which docs are linked in the current project.
This tells AI assistants where to find documentation.

### Option A: `:::docs` directive

In `directives.md`:
```markdown
:::docs
```

Expands during `agmd sync` to:
```markdown
## Available Documentation

The following documentation is available in this project:

- `./docs/svelte-kit/` - SvelteKit framework documentation
- `./docs/fastapi/` - FastAPI framework guide

Refer to these docs for framework-specific guidance.
```

### Option B: `:::include guide:docs` with dynamic resolution

Create a special `guide:docs` that uses a placeholder:

```markdown
---
name: docs
description: "Dynamic documentation reference"
---

## Available Documentation

{{linked_docs}}

Refer to these docs for framework-specific guidance.
```

The `{{linked_docs}}` placeholder is resolved during sync based on:
1. Symlinks in project pointing to `~/.agmd/doc/`
2. Or entries in `.agmd.json`

### Recommendation

**Option A (`:::docs`)** is cleaner:
- No special guide file needed
- Clear intent in directives.md
- Auto-discovers linked docs

### Tasks

- [ ] Add `:::docs` directive to parser
- [ ] Detect linked doc symlinks in project
- [ ] Expand to list of available docs with paths
- [ ] Include doc descriptions if available (from doc folder's README or metadata)

---

## 4. Enhanced List Command

### Current State
- `agmd list` shows all items
- `agmd list type` shows items of that type

### New Features

```bash
agmd list                    # List all types and counts
agmd list rule               # List rules
agmd list file               # List files (shows folder structure)
agmd list doc                # List doc folders
agmd list profile            # List profiles (show files: count)
agmd list doc:svelte-kit     # List contents of specific doc folder
agmd list file:python        # List contents of file folder
```

### Tasks

- [ ] Add folder content listing for `doc:` and `file:` types
- [ ] Show `files:` count in profile listing
- [ ] Tree view for nested structures

---

## 5. File Type Handling

### Design Decisions

**Storage:** Files stored as-is (raw files, no wrapper)

```
~/.agmd/file/
├── gitignore/
│   ├── python           # Contains actual .gitignore content
│   └── node
└── python/
    ├── pyproject.toml   # Actual pyproject.toml
    └── ruff.toml
```

**Copy behavior:**
- `file:gitignore/python > .gitignore` - Copy file, rename to `.gitignore`
- `file:python/*` - Copy all files from folder, keep original names

**Commands:**
```bash
# Create from content
agmd new file:gitignore/python --content "__pycache__/\n*.pyc"

# Import existing file
agmd new file:gitignore/python --from ./.gitignore

# Import folder
agmd new file:python --from ./templates/

# Show file content
agmd show file:gitignore/python
```

### Tasks

- [ ] Handle `file:` type differently (no .md extension, no frontmatter)
- [ ] `agmd new file:name --from` to import files
- [ ] `agmd new file:folder --from ./dir/` to import folders
- [ ] `agmd show file:name` to display raw content

---

## 6. Agent Skills Support (Future)

### Alignment with agentskills.io

Our structure aligns well:
- `profile/` → Similar to `SKILL.md` (instructions + metadata)
- `file/` → Similar to `assets/` (templates, configs)
- `doc/` → Similar to `references/` (documentation)

### Potential Integration

```yaml
---
name: python-dev
description: "Python development setup"
files:
  - file:gitignore/python > .gitignore
skills:
  - skill:code-review          # Include external skill
  - skill:testing/pytest
---
```

### Tasks (Future)

- [ ] Research skills-ref validation
- [ ] Consider `skill/` type for local skills
- [ ] Import/export skills from agentskills.io

---

## Implementation Order

### Phase 1: File Type
1. Support `file/` as raw file storage (no .md wrapper)
2. `agmd new file:name --content` / `--from`
3. `agmd show file:name` for raw content
4. `agmd list file` / `agmd list file:folder`

### Phase 2: Profile Enhancement
1. Profile frontmatter `files:` parsing
2. `>` rename syntax
3. `/*` folder copy syntax
4. `agmd init profile:name` copies files

### Phase 3: Docs & Linking
1. `doc/` as folder type
2. `agmd link` / `agmd unlink` commands
3. `--gitignore` integration
4. `.agmd.json` tracking (optional)

### Phase 4: Dynamic Docs Directive
1. `:::docs` directive in parser
2. Auto-detect linked docs
3. Expand to documentation list

### Phase 5: Task Management
1. `task/` type with project-based organization
2. `agmd task` subcommand for status/dependencies
3. Auto-sorted list by dependency readiness
4. Delete with `agmd delete task:name`

### Phase 6: Skills (Future)
1. Agent Skills spec alignment
2. Skill import/export

---

## 7. Task Management

### Concept

Project-based task tracking with dependencies. Similar to Claude Code's task system but stored in the agmd registry, organized by project (cwd name).

### Structure

```
~/.agmd/
├── task/                           # Reserved: Tasks by project
│   ├── agent-md/                   # Project name (basename of cwd)
│   │   ├── setup-database.md
│   │   ├── create-api.md
│   │   └── write-tests.md
│   └── other-project/
│       └── init.md
├── doc/
├── file/
├── profile/
└── rule/
```

### Task File Format

```yaml
---
subject: Set up database schema
status: pending                     # pending | in_progress | completed
depends_on: []                      # Task IDs this depends on
---

Create the initial database schema with users and posts tables.
This is the foundation that other tasks depend on.
```

**Example with dependencies:**

```yaml
---
subject: Create API endpoints
status: pending
depends_on: [setup-database]
---

Create REST API endpoints for users and posts.
```

### Auto-Computed Status

Status is computed at display time (no file changes needed):
- `[ready]` - All dependencies completed (or no dependencies)
- `[blocked]` - Has pending dependencies
- `[in_progress]` - Manually set to in_progress
- `[completed]` - Done

### Auto-Sorted Display

When listing tasks, they are automatically sorted:
1. **Ready tasks** (can start now)
2. **Blocked tasks** (waiting on dependencies)
3. **Completed tasks** (shown with `--all`, hidden by default)

### Commands

```bash
# Create task (project = basename of cwd)
agmd new task:setup-database --content "Create database schema..."
agmd new task:create-api --content "Create endpoints" --blocked-by "setup-database"

# Create for specific project
agmd new task:setup-database --project other-name --content "..."

# List tasks (auto-sorted: ready → blocked → completed)
agmd list task                      # Current project
agmd list task --project foo        # Specific project
agmd list task --all                # Include completed tasks

# Show task
agmd show task:setup-database                   # Current project
agmd show task:other-project/setup-database     # Specific project

# Show all tasks with content
agmd show task --all                # All tasks for current project

# Delete task (consistent with other types)
agmd delete task:setup-database
agmd delete task:setup-database --force

# Task subcommand for status/dependencies
agmd task status setup-database pending
agmd task status setup-database in_progress
agmd task status setup-database completed

agmd task blocked-by create-api setup-database    # Add: create-api depends on setup-database
agmd task unblock create-api setup-database       # Remove dependency
```

### Example Flow

**Initial state:**
```
$ agmd list task

Tasks for: agent-md (3 tasks)

[ready] setup-database
  Set up database schema

[blocked] create-api (waiting: setup-database)
  Create API endpoints

[blocked] write-tests (waiting: setup-database, create-api)
  Write tests
```

**After completing setup-database:**
```
$ agmd task status setup-database completed
$ agmd list task

Tasks for: agent-md (3 tasks)

[ready] create-api
  Create API endpoints

[blocked] write-tests (waiting: create-api)
  Write tests

[completed] setup-database ✓
  Set up database schema
```

### Tasks

- [ ] Add `task/` as reserved type with project-based folders
- [ ] Task frontmatter: `subject`, `status`, `depends_on`
- [ ] `agmd new task:name` with `--project`, `--blocked-by` flags
- [ ] `agmd list task` with auto-sorting (ready → blocked → completed)
- [ ] `agmd list task --all` to include completed
- [ ] `agmd show task:name` and `agmd show task --all`
- [ ] `agmd delete task:name` (consistent with other types)
- [ ] `agmd task status <name> <status>` subcommand
- [ ] `agmd task blocked-by <task> <dependency>` subcommand
- [ ] `agmd task unblock <task> <dependency>` subcommand
- [ ] Compute ready/blocked status at display time
- [ ] Dependency tree visualization in list output
