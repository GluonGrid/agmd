---
name: agmd
description: Workflow for updating AI instructions with agmd
---

# Updating AI Instructions

When modifying AI instructions for this project:

1. **Edit directives.md** - Add or modify directives
2. **Run `agmd sync`** - Regenerate AGENTS.md
3. **Review changes** - Check the generated output
4. **Commit both files** - Keep them in sync

## Adding New Content

```bash
# Add existing item from registry
agmd add rule:name

# Create new item inline (in directives.md)
# Then promote to registry
agmd promote
agmd sync
```

## Quick Reference

| Action | Command |
|--------|---------|
| Sync changes | `agmd sync` |
| Add item | `agmd add type:name` |
| Remove item | `agmd remove type:name` |
| List available | `agmd list` |
| Edit in registry | `agmd edit type:name` |
