package parser

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/util"
)

// DirectiveExtension is a Goldmark extension for directive parsing
type DirectiveExtension struct {
	RegistryPath string
}

// NewDirectiveExtension creates a new directive extension
func NewDirectiveExtension(registryPath string) *DirectiveExtension {
	return &DirectiveExtension{
		RegistryPath: registryPath,
	}
}

// Extend extends the Goldmark parser with directive support
func (e *DirectiveExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewDirectiveParser(), 100),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewDirectiveTransformer(e.RegistryPath), 100),
		),
	)
}
