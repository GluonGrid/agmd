---
name: agmd
description: How to use agmd CLI for managing AI agent instructions
---

# Using agmd

This project uses **agmd** to manage AI instructions. The `directives.md` file is the source of truth - never edit `AGENTS.md` directly.

## Key Files

- `directives.md` - Source file you edit (contains directives)
- `AGENTS.md` - Generated output (do not edit manually)

Run `agmd list` or `agmd list --tree` to see available items in your registry.

## Directive Syntax

```markdown
:::include rule:name
:::include workflow:name

:::list rule
item1
item2
:::end

:::new rule:name
Content here...
:::end
```

## Common Commands

```bash
agmd sync              # Generate AGENTS.md from directives.md
agmd add rule:name     # Add a rule to directives.md
agmd remove rule:name  # Remove a rule from directives.md
agmd list              # Show available registry items
agmd edit rule:name    # Edit a rule in the registry
```

## Workflow

1. Edit `directives.md` to add/remove directives
2. Run `agmd sync` to regenerate `AGENTS.md`
3. If using `:::new` blocks, run `agmd promote` first

## Important

- Always run `agmd sync` after editing `directives.md`
- Never commit `AGENTS.md` changes without syncing from `directives.md`
- Use `:::new` for project-specific content, `:::include` for reusable content
