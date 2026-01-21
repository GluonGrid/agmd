# Goldmark POC - SUCCESS! ✅

## Summary

The proof of concept is **working successfully**! The custom Goldmark parser can:

1. ✅ Parse `:::list` blocks and collect rule names
2. ✅ Parse `:::new-rule` blocks and preserve their markdown content
3. ✅ Expand `:::list` directives by loading content from registry files
4. ✅ Render clean Markdown output without directive markers
5. ✅ Preserve all content outside directive blocks automatically

## Key Learnings

### 1. Reader Advancement
**Critical**: When advancing the reader after processing a line, you must account for the newline character:

```go
newline := 1
if len(line) > 0 && line[len(line)-1] != '\n' {
    newline = 0
}
reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
```

This is what the fenced code block parser does, and it's essential for Continue() to be called for every line.

### 2. Context Storage
Use `parser.Context` to store parser state between Open() and Continue() calls:

```go
var directiveDataKey = parser.NewContextKey()

// In Open()
pc.Set(directiveDataKey, &directiveData{node})

// In Close()
data := pc.Get(directiveDataKey)
```

### 3. AST Node Source References
AST nodes store `text.Segment` references to source bytes, not actual text. When you parse external content (registry files) and attach those nodes to the main document, their segments point to the wrong source.

**Solution**: Use `ast.NewString()` nodes which store actual byte content:

```go
heading := ast.NewHeading(3)
heading.AppendChild(heading, ast.NewString([]byte(content)))
listBlock.AppendChild(listBlock, heading)
```

### 4. Parser vs Renderer
- **Parser**: Converts Markdown → AST
- **Transformer**: Modifies AST (expands directives)
- **Renderer**: Converts AST → Output

All three need to be coordinated.

## What Works

### Input (pre-rendered.md)
```markdown
## Core Rules

:::list
typescript
eslint
:::end

## Custom Inline Rule

:::new-rule name=my-auth-rule
# Authentication Rules
- Always use JWT
:::end

## Project Notes
Custom content is preserved!
```

### Output (AGENTS.md)
```markdown
## Core Rules

### typescript
# TypeScript Rules
- Always use strict typing
- Avoid using `any` type

### eslint
# ESLint Rules
- Run ESLint on all files

## Custom Inline Rule

# Authentication Rules
- Always use JWT

## Project Notes
Custom content is preserved!
```

## Next Steps for Integration

1. Remove debug output from parser_v3.go, transformer.go, renderer.go
2. Improve content rendering (currently wraps in heading+paragraph, should preserve markdown structure better)
3. Integrate into main agmd codebase:
   - Move parser_v3.go, transformer.go, renderer.go to `pkg/template/`
   - Update `agmd render` command to use this pipeline
   - Update `agmd init` to create `pre-rendered.md`
   - Update `agmd add/remove` to edit `:::list` blocks
4. Add proper error handling
5. Add tests

## Files

- `parser_v3.go` - Custom block parser for `:::list` and `:::new-rule` ✅
- `transformer.go` - AST transformer to expand directives ✅
- `renderer.go` - Markdown renderer ✅
- `ast_nodes.go` - Custom AST node definitions ✅
- `main.go` - Test harness ✅

## Conclusion

The Goldmark-based approach **works**! The key was understanding:
1. How to properly advance the reader
2. How Goldmark's AST uses segment references
3. How to use ast.String for external content

Ready to integrate into the main codebase!
