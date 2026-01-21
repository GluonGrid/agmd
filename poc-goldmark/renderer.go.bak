package main

import (
	"io"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

// MarkdownRenderer renders AST back to clean Markdown
type MarkdownRenderer struct{}

// NewMarkdownRenderer creates a new MarkdownRenderer
func NewMarkdownRenderer() renderer.Renderer {
	return &MarkdownRenderer{}
}

// Render renders the AST to clean Markdown
func (r *MarkdownRenderer) Render(w io.Writer, source []byte, node ast.Node) error {
	return ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		return r.renderNode(w, source, n, entering)
	})
}

// renderNode renders a single node
func (r *MarkdownRenderer) renderNode(w io.Writer, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	switch n := node.(type) {
	case *ast.Document:
		// Just process children
		return ast.WalkContinue, nil

	case *ListBlock:
		// ListBlock has been expanded, just render its children
		return ast.WalkContinue, nil

	case *NewItemBlock:
		if entering {
			// Render new item block content
			w.Write([]byte("\n"))
			return ast.WalkContinue, nil
		}
		w.Write([]byte("\n"))
		return ast.WalkContinue, nil

	case *CustomSection:
		if entering {
			// Render custom section content
			w.Write([]byte("\n"))
			return ast.WalkContinue, nil
		}
		w.Write([]byte("\n"))
		return ast.WalkContinue, nil

	case *ast.Heading:
		if entering {
			for i := 0; i < n.Level; i++ {
				w.Write([]byte{'#'})
			}
			w.Write([]byte{' '})
		} else {
			w.Write([]byte("\n\n"))
		}
		return ast.WalkContinue, nil

	case *ast.Paragraph:
		if !entering {
			w.Write([]byte("\n\n"))
		}
		return ast.WalkContinue, nil

	case *ast.List:
		if !entering {
			w.Write([]byte("\n"))
		}
		return ast.WalkContinue, nil

	case *ast.ListItem:
		if entering {
			w.Write([]byte("- "))
		} else {
			w.Write([]byte("\n"))
		}
		return ast.WalkContinue, nil

	case *ast.Text:
		if entering {
			w.Write(n.Segment.Value(source))
			if n.SoftLineBreak() {
				w.Write([]byte{'\n'})
			}
		}
		return ast.WalkContinue, nil

	case *ast.String:
		if entering {
			w.Write(n.Value)
		}
		return ast.WalkContinue, nil

	case *ast.Emphasis:
		if entering {
			if n.Level == 1 {
				w.Write([]byte{'*'})
			} else {
				w.Write([]byte("**"))
			}
		} else {
			if n.Level == 1 {
				w.Write([]byte{'*'})
			} else {
				w.Write([]byte("**"))
			}
		}
		return ast.WalkContinue, nil

	case *ast.CodeSpan:
		if entering {
			w.Write([]byte{'`'})
			w.Write(n.Text(source))
			w.Write([]byte{'`'})
		}
		return ast.WalkSkipChildren, nil

	default:
		// Skip other node types for now
		return ast.WalkContinue, nil
	}
}

// AddOptions adds options to the renderer
func (r *MarkdownRenderer) AddOptions(...renderer.Option) {}
