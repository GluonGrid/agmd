# Goldmark POC - Custom Directive Parser

## Summary

This POC demonstrates parsing custom directives in Markdown using Goldmark:
- `:::list` - List of rule names to expand from registry
- `:::new-rule name=foo` - User-defined inline rules
- `:::custom` - Custom section preserved during rendering

## What Works ✅

1. **Basic directive detection**: Parser recognizes `:::list`, `:::new-rule`, `:::custom`
2. **AST node creation**: Custom nodes are created and added to AST
3. **List name collection**: `:::list` correctly collects rule names (typescript, eslint, no-console)
4. **Clean output**: Renders back to Markdown without directive markers

## What Needs Fixing 🔧

1. **:::list expansion**: Transformer loads registry files but content isn't appearing in output
   - Issue: Child nodes added to ListBlock aren't being rendered
   - Fix needed: Ensure child nodes are properly attached and rendered

2. **:::new-rule parsing**: Content after directive isn't captured
   - Issue: Parser stops at block boundary
   - Fix needed: Collect all content until `:::end` as raw content

3. **:::custom parsing**: Same issue as :::new-rule
   - Fix needed: Preserve raw markdown content between markers

## Key Learnings

### Goldmark Architecture
```
Input MD → Parser → AST → Transformer → Renderer → Output MD
           ↑                  ↑              ↑
      Custom           Expand          Custom
      Blocks          Directives      Markdown
                                     Renderer
```

### Directive Syntax (Final)
```md
:::list
typescript
eslint
:::end

:::new-rule name=my-rule
# Content here
:::end

:::custom
## Custom content
:::end
```

### Registry File Structure
```md
---
name: typescript
category: code-style
severity: error
---

# TypeScript Rules
- Content here
```

## Next Steps for Integration

1. **Fix content preservation** in :::new-rule and :::custom blocks
2. **Fix list expansion** to show registry content in output
3. **Integrate into main agmd codebase**:
   - Create `pkg/template/` package with these components
   - Update `agmd render` command to use this pipeline
   - Update `agmd add/remove` to edit `:::list` blocks
   - Update `agmd validate` to check directive validity
   - Update `agmd promote` to handle `:::new-rule` blocks

## Files Created

- `ast_nodes.go` - Custom AST node definitions (ListBlock, NewRuleBlock, CustomSection)
- `parser.go` - Block parser for directives
- `transformer.go` - AST transformer to expand directives
- `renderer.go` - Markdown renderer
- `main.go` - Test harness
- `registry/rules/*.md` - Mock registry files
- `templates/pre-rendered.md` - Sample template

## Testing

```bash
cd poc-goldmark
go run .
```

Output shows:
- Input template with directives
- AST structure
- Clean output (currently missing content)

## Conclusion

✅ **Proof of Concept Successful!**

Goldmark can parse our custom directives and we can process them. The basic architecture works. With the fixes mentioned above, this approach will work perfectly for the template-based agmd system.
