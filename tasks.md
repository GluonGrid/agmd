# agmd v2 - Enhanced Profiles & Skills

## Overview

Enhance agmd to support:
1. **Profiles with file copying** - Bootstrap projects with templates
2. **Docs linking** - Symlink documentation folders for AI context
3. **Agent Skills support** - Future integration with agentskills.io

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
├── file/                    # Shared files library
│   ├── gitignore/
│   │   ├── python.md
│   │   └── node.md
│   ├── readme/
│   │   └── minimal.md
│   └── python/              # Can copy entire folder
│       ├── pyproject.toml.md
│       ├── ruff.toml.md
│       └── .python-version.md
├── doc/                     # Documentation folders for AI context
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
  - file:readme/minimal > README.md
  - file:python/*                            # Copy all files from folder
---

# Agent Instructions
:::include guide:agmd
:::include rule:python
```

### Tasks

- [ ] Update profile frontmatter parsing to support `files:` field
- [ ] Implement file copy syntax: `file:type/name > destination`
- [ ] Implement folder copy syntax: `file:folder/*` (copies all, keeps names)
- [ ] Update `agmd init profile:name` to copy files
- [ ] Add `agmd new file:name` command (stores as `.md` with content)
- [ ] Handle file extension mapping (`.md` in registry → actual extension on copy)

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
```

### Tasks

- [ ] Support `doc/` as folder-based type (not single .md files)
- [ ] Add `agmd link doc:name` command
- [ ] Add `agmd unlink doc:name` command
- [ ] Add `--gitignore` flag to auto-add to .gitignore
- [ ] Add `--from` flag to import existing folder

---

## 3. Enhanced List Command

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

## 4. File Type Handling

### Design Decisions

**Storage:** Files stored as `.md` in registry with frontmatter
```yaml
---
name: python
description: "Python gitignore"
filename: .gitignore          # Original filename (optional, for reference)
---

__pycache__/
*.pyc
.venv/
dist/
```

**Copy behavior:**
- `file:gitignore/python > .gitignore` - Strips frontmatter, writes to `.gitignore`
- `file:python/*` - Copies all files, uses `filename` from frontmatter or strips `.md`

### Tasks

- [ ] Add `filename` frontmatter field for files
- [ ] Strip frontmatter when copying files
- [ ] Handle filename resolution (frontmatter > strip .md)

---

## 5. Agent Skills Support (Future)

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

### Phase 1: Core Profile Enhancement
1. File type and storage
2. Profile frontmatter `files:` parsing
3. `agmd init` file copying
4. `>` rename syntax
5. `/*` folder copy syntax

### Phase 2: Docs & Linking
1. Doc folder type
2. `agmd link` / `agmd unlink` commands
3. Gitignore integration

### Phase 3: Enhanced UX
1. List command improvements
2. Tree views
3. `--from` import flag

### Phase 4: Skills (Future)
1. Agent Skills spec alignment
2. Skill import/export
