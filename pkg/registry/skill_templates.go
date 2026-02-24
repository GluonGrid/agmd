package registry

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// GetManagingAgmdSkillTemplate returns the SKILL.md for the bundled managing-agmd skill
func GetManagingAgmdSkillTemplate() string {
	return `---
name: managing-agmd
description: "Manages agmd projects: initializing registries, migrating existing AI instruction files (CLAUDE.md, AGENTS.md, .cursorrules) to directives.md format, managing reusable rules/workflows/guides, syncing output, and tracking project tasks. Use when the user mentions agmd, directives.md, wants to migrate AI instructions, organize agent rules, or manage project tasks with agmd."
license: MIT
compatibility: claude, cursor, windsurf, any
allowed-tools: Bash, Read, Write, Edit, Glob, Grep
---

## Managing agmd Projects

agmd manages AI agent instructions via a personal registry (` + "`~/.agmd/`" + `) and a per-project ` + "`directives.md`" + ` that expands to a full ` + "`AGENTS.md`" + `.

### Quick Orientation

| File | Role |
|------|------|
` + "| `directives.md` | Source — short, readable, contains `:::use` directives. **Edit this.** |" + `
` + "| `AGENTS.md` | Generated output — full expanded content. **Never edit directly.** |" + `
` + "| `~/.agmd/` | Global registry — personal reusable rules, workflows, guides |" + `
` + "| `.agmd/` | Local registry — project-specific, committed with team |" + `
` + "| `directives.local.md` | Personal overrides — gitignored, merged at sync time |" + `

### Directive Syntax

` + "```" + `markdown
:::use type:name         # single item from registry
:::list                  # multiple items, mixed types
rule:typescript
workflow:commit
:::end
:::new type:name         # define inline (promote to registry later)
content here...
:::end
:::skills                # inject <available_skills> XML block
:::docs                  # list symlinked documentation
` + "```" + `

### Most Common Commands

` + "```" + `bash
# Setup (once per machine)
agmd setup

# Per project
agmd init                # create directives.md
agmd sync                # generate AGENTS.md
agmd symlink --target claude  # link AGENTS.md → CLAUDE.md

# Registry items
agmd new type:name --content "..."   # create
agmd edit type:name --content "..."  # update
agmd show type:name                  # read
agmd list                            # browse all

# Migration (existing projects)
agmd migrate CLAUDE.md   # raw file → directives.md
agmd collect             # agmd project → pull items into ~/.agmd/
agmd promote --all       # :::new blocks → registry

# Tasks
agmd task list           # list tasks
agmd task new <name>     # create task
agmd task status <name> completed
` + "```" + `

### Verify Setup

Run the check script to confirm agmd is ready:

` + "```" + `bash
bash scripts/check-setup.sh
` + "```" + `

### Detailed References

Load these on demand — only when the task requires deeper detail:

- **Migration workflow** → [references/migration.md](references/migration.md)
- **Task management** → [references/tasks.md](references/tasks.md)
- **Directive syntax deep-dive** → [references/directives.md](references/directives.md)
`
}

// GetManagingAgmdMigrationRef returns references/migration.md content
func GetManagingAgmdMigrationRef() string {
	return `# Migration Reference

Complete guide for migrating existing AI instruction files to agmd format.

## When to Use Which Command

| Situation | Command |
|-----------|---------|
| Raw/unstructured file (CLAUDE.md, .cursorrules, etc.) | ` + "`agmd migrate <file>`" + ` |
| Project already uses agmd (has directives.md) | ` + "`agmd collect`" + ` |
| Fresh project, no existing instructions | ` + "`agmd init`" + ` |

## Full Migration Flow (Raw File)

### 1. Check agmd is installed

` + "```" + `bash
agmd --version
# If missing:
curl -fsSL https://gluongrid.dev/agmd/install.sh | bash
agmd setup
` + "```" + `

### 2. Identify the source file

Common locations:
- ` + "`CLAUDE.md`" + ` — Claude Code
- ` + "`.cursorrules`" + ` — Cursor
- ` + "`.windsurfrules`" + ` — Windsurf
- ` + "`.github/copilot-instructions.md`" + ` — GitHub Copilot
- ` + "`AGENTS.md`" + ` — generic agents

### 3. Run migrate

` + "```" + `bash
agmd migrate CLAUDE.md
# or force if directives.md already exists:
agmd migrate CLAUDE.md --force
` + "```" + `

What it does:
1. Creates ` + "`CLAUDE.md.backup`" + `
2. Copies content to ` + "`directives.md`" + ` with ` + "`:::use guide:agmd`" + ` header
3. Opens editor for you to wrap sections

### 4. Wrap reusable sections with :::new

` + "```" + `markdown
:::new rule:typescript
# TypeScript Standards
Use strict mode. Prefer ` + "`unknown`" + ` over ` + "`any`" + `.
:::end

:::new workflow:commit
# Commit Process
Run tests → stage specific files → conventional commit message.
:::end
` + "```" + `

Leave project-specific content unwrapped — it stays in ` + "`directives.md`" + ` as plain text.

### 5. Promote to registry

` + "```" + `bash
agmd promote --all
` + "```" + `

Each ` + "`:::new`" + ` block is saved to ` + "`~/.agmd/type/name.md`" + ` and replaced with ` + "`:::use type:name`" + `.

### 6. Sync

` + "```" + `bash
agmd sync
` + "```" + `

### 7. Symlink to AI tool's expected filename

` + "```" + `bash
agmd symlink --target claude   # → CLAUDE.md
agmd symlink --target cursor   # → .cursorrules
agmd symlink --target windsurf # → .windsurfrules
agmd symlink --list            # see all targets
` + "```" + `

## Collecting from an Existing agmd Project

` + "```" + `bash
agmd collect              # reads AGENTS.md, saves items to ~/.agmd/
agmd collect -f CLAUDE.md # use a different output file
` + "```" + `

## After Migration

` + "```" + `bash
agmd list                         # inspect registry
agmd show rule:typescript         # read a rule

# Reuse in another project
cd ~/other-project && agmd init && agmd sync

# Sync registry across machines
agmd git remote add origin <url>
agmd git add -A && agmd git commit -m "sync" && agmd git push
` + "```" + `

## Local vs Global Registry

` + "```" + `bash
agmd new rule:my-rule              # global (personal)
agmd new rule:team-style --local   # local (team-shared, committed)
agmd mv rule:my-rule --to-local    # move global → local
agmd mv rule:team-rule --to-global # move local → global
` + "```" + `

Resolution order: ` + "`.agmd/`" + ` (local) wins over ` + "`~/.agmd/`" + ` (global).

## Personal Overrides (directives.local.md)

` + "```" + `bash
echo ":::use rule:my-local-tool" >> directives.local.md
agmd sync   # automatically merges directives.local.md
` + "```" + `

` + "`agmd init`" + ` adds ` + "`directives.local.md`" + ` to ` + "`.gitignore`" + ` automatically.
`
}

// GetManagingAgmdTasksRef returns references/tasks.md content
func GetManagingAgmdTasksRef() string {
	return `# Task Management Reference

agmd includes project-level task tracking with dependencies, priorities, types, and feature grouping. Tasks are stored globally in ` + "`~/.agmd/task/`" + ` — personal, not committed.

## Task Anatomy

| Field | Values | Description |
|-------|--------|-------------|
| ` + "`status`" + ` | ` + "`ready`, `blocked`, `in_progress`, `completed`" + ` | Current state |
| ` + "`priority`" + ` | ` + "`P0`–`P4`" + ` | P0 = critical, P4 = nice-to-have |
| ` + "`type`" + ` | ` + "`bug`, `feature`, `task`, `chore`" + ` | Nature of work |
| ` + "`feature`" + ` | any string | Groups related tasks (like a folder) |
| ` + "`blocked-by`" + ` | task name(s) | Dependencies — auto-sets status to ` + "`blocked`" + ` |

## Core Commands

` + "```" + `bash
# List
agmd task list                        # all active tasks
agmd task list --tree                 # dependency tree, grouped by feature
agmd task list --all                  # include completed
agmd task list --feature auth         # filter by feature
agmd task list --status ready         # filter by status
agmd task list --type bug --priority 0 # critical bugs only

# Create
agmd task new setup-db \
  --content "Set up PostgreSQL schema" \
  --feature backend --priority 1 --type task

agmd task new fix-login \
  --feature auth --priority 0 --type bug \
  --blocked-by setup-db

# Update
agmd task status setup-db completed
agmd task edit setup-db --priority 0 --type bug

# Dependencies
agmd task blocked-by api setup-db   # api blocked by setup-db
agmd task unblock api setup-db      # remove dependency

# Cleanup
agmd task clean                     # delete all completed
agmd task delete setup-db --force   # delete specific task
` + "```" + `

## Priority Guide

| Priority | Use for |
|----------|---------|
| P0 | Production down, data loss, security issue |
| P1 | Major feature broken, blocking teammates |
| P2 | Normal feature work, standard bugs |
| P3 | Minor improvements, non-blocking |
| P4 | Nice-to-have, backlog |

## Typical Workflow

` + "```" + `bash
# Plan
agmd task new design-schema --feature payments --priority 1 --type task
agmd task new implement-api --feature payments --priority 1 --type feature \
  --blocked-by design-schema

# Work
agmd task status design-schema in_progress
agmd task status design-schema completed
agmd task unblock implement-api design-schema
agmd task status implement-api in_progress

# Review
agmd task list --feature payments --tree

# Cleanup
agmd task clean
` + "```" + `
`
}

// GetManagingAgmdDirectivesRef returns references/directives.md content
func GetManagingAgmdDirectivesRef() string {
	return `# Directive Syntax Reference

Complete reference for all directives available in ` + "`directives.md`" + `.

## :::use — Single Item

` + "```" + `markdown
:::use rule:typescript
:::use workflow:commit
:::use rule:frontend/typescript   # subdirectory support
` + "```" + `

Type is the folder name in ` + "`~/.agmd/`" + ` — completely user-defined.
Common conventions: ` + "`rule`, `workflow`, `guide`, `guardrail`, `prompt`, `pattern`" + `.

## :::list — Multiple Items (Mixed Types)

` + "```" + `markdown
:::list
rule:typescript
rule:eslint
workflow:commit
guideline:code-style
:::end
` + "```" + `

No type on the ` + "`:::list`" + ` line — each line declares its own type freely.

## :::new — Inline Definition

` + "```" + `markdown
:::new rule:custom-auth
# Custom Auth Rules
Always validate JWT on every request.
:::end
` + "```" + `

Requires promotion before sync:
` + "```" + `bash
agmd promote --all
` + "```" + `

## :::skills — Agent Skills XML

` + "```" + `markdown
:::skills                    # auto-detect .claude/skills or .agents/skills
:::skills .claude/skills     # explicit path
` + "```" + `

Output: ` + "`<available_skills>`" + ` XML block with name, description, location per skill.

` + "```" + `bash
agmd skill install owner/repo
agmd skill link managing-agmd --target both
` + "```" + `

## :::docs — Documentation List

` + "```" + `markdown
:::docs              # default: ./docs/
:::docs ./reference  # custom path
` + "```" + `

## Registry Resolution

For ` + "`:::use type:name`" + `:
1. ` + "`.agmd/type/name.md`" + ` — project-local (team-shared)
2. ` + "`~/.agmd/type/name.md`" + ` — global (personal)

` + "```" + `bash
agmd mv rule:my-rule --to-local   # share with team
agmd mv rule:my-rule --to-global  # keep personal
` + "```" + `

## directives.local.md

Gitignored companion, same syntax, merged at sync:

` + "```" + `markdown
:::use rule:my-local-tool
` + "```" + `

` + "`agmd init`" + ` adds it to ` + "`.gitignore`" + ` automatically.
`
}

// GetManagingAgmdCheckScript returns scripts/check-setup.sh content
func GetManagingAgmdCheckScript() string {
	return `#!/usr/bin/env bash
# check-setup.sh — verify agmd is installed and registry is ready
# Usage: bash scripts/check-setup.sh

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

ok()   { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; FAILED=1; }
warn() { echo -e "${YELLOW}!${NC} $1"; }
info() { echo -e "${BLUE}→${NC} $1"; }

FAILED=0

echo ""
info "Checking agmd setup..."
echo ""

if command -v agmd &>/dev/null; then
  VERSION=$(agmd --version 2>&1 | head -1)
  ok "agmd installed: $VERSION"
else
  fail "agmd not found in PATH"
  echo "   Install: curl -fsSL https://gluongrid.dev/agmd/install.sh | bash"
  exit 1
fi

REGISTRY="$HOME/.agmd"
if [ -d "$REGISTRY" ]; then
  ok "Global registry exists: $REGISTRY"
else
  fail "Global registry not found at $REGISTRY"
  echo "   Run: agmd setup"
  FAILED=1
fi

if [ -f "$REGISTRY/guide/agmd.md" ]; then
  ok "guide:agmd present"
else
  warn "guide:agmd missing — run: agmd setup --force"
fi

if [ -d "$REGISTRY/skill/managing-agmd" ]; then
  ok "managing-agmd skill present"
else
  warn "managing-agmd skill missing — run: agmd setup --force"
fi

if [ -f "directives.md" ]; then
  ok "directives.md found in current directory"
else
  warn "No directives.md in current directory"
  echo "   Run: agmd init           (fresh project)"
  echo "   Run: agmd migrate <file> (existing instructions)"
fi

if [ -d ".agmd" ]; then
  ok "Local registry (.agmd/) found"
else
  info "No local registry — use 'agmd init --local' to create one (optional)"
fi

if [ -f "AGENTS.md" ]; then
  ok "AGENTS.md found (run 'agmd sync' to regenerate)"
else
  warn "No AGENTS.md — run: agmd sync"
fi

if [ -f ".gitignore" ] && grep -q "directives.local.md" .gitignore 2>/dev/null; then
  ok "directives.local.md is gitignored"
elif [ -f ".gitignore" ]; then
  warn "directives.local.md not in .gitignore — run: agmd init (or add manually)"
fi

echo ""
if [ "$FAILED" -eq 0 ]; then
  ok "All checks passed."
else
  fail "Some checks failed — see above."
  exit 1
fi
`
}

// GetManagingAgmdIcon generates a 128x128 PNG icon for the managing-agmd skill.
// Returns nil if generation fails (icon is optional).
func GetManagingAgmdIcon() []byte {
	const size = 128

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	bg := color.RGBA{26, 31, 46, 255}
	card := color.RGBA{32, 40, 60, 255}
	accent := color.RGBA{79, 158, 255, 255}
	secondary := color.RGBA{123, 184, 255, 255}
	white := color.RGBA{235, 240, 255, 255}

	inRoundRect := func(x, y, rx, ry, rw, rh, r int) bool {
		if x < rx+r && y < ry+r {
			dx, dy := rx+r-x, ry+r-y
			return dx*dx+dy*dy <= r*r
		}
		if x > rx+rw-r && y < ry+r {
			dx, dy := x-(rx+rw-r), ry+r-y
			return dx*dx+dy*dy <= r*r
		}
		if x < rx+r && y > ry+rh-r {
			dx, dy := rx+r-x, y-(ry+rh-r)
			return dx*dx+dy*dy <= r*r
		}
		if x > rx+rw-r && y > ry+rh-r {
			dx, dy := x-(rx+rw-r), y-(ry+rh-r)
			return dx*dx+dy*dy <= r*r
		}
		return x >= rx && x <= rx+rw && y >= ry && y <= ry+rh
	}

	// Fill background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bg)
		}
	}

	// Draw rounded card
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if inRoundRect(x, y, 8, 8, size-16, size-16, 16) {
				img.Set(x, y, card)
			}
		}
	}

	// :::use line — accent keyword + white name
	for y := 28; y <= 34; y++ {
		for x := 20; x <= 52; x++ {
			img.Set(x, y, accent)
		}
		for x := 56; x <= 100; x++ {
			img.Set(x, y, white)
		}
	}

	// :::list line — secondary keyword
	for y := 46; y <= 52; y++ {
		for x := 20; x <= 48; x++ {
			img.Set(x, y, secondary)
		}
	}

	// indented item lines
	for y := 60; y <= 66; y++ {
		for x := 30; x <= 98; x++ {
			img.Set(x, y, white)
		}
	}
	for y := 72; y <= 78; y++ {
		for x := 30; x <= 82; x++ {
			img.Set(x, y, white)
		}
	}

	// :::end line — accent
	for y := 88; y <= 94; y++ {
		for x := 20; x <= 50; x++ {
			img.Set(x, y, accent)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}
