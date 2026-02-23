package parser

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// ParseAndExpand reads markdown with directives, expands them from registry, and returns expanded markdown.
// localPath is optional (empty string if no local registry).
func ParseAndExpand(input []byte, registryPath, localPath string) ([]byte, error) {
	// Create Goldmark with GFM + our directive extension
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			NewDirectiveExtension(registryPath, localPath),
		),
	)

	// Parse the markdown
	reader := text.NewReader(input)
	doc := md.Parser().Parse(reader)

	// Render back to markdown
	var buf bytes.Buffer
	renderer := NewMarkdownRenderer()
	if err := renderer.Render(&buf, input, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
