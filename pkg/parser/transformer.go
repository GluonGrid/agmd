package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

// ItemMeta represents frontmatter metadata
type ItemMeta struct {
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
	Severity string `yaml:"severity"`
}

// DirectiveTransformer expands directive blocks
type DirectiveTransformer struct {
	RegistryPath string
}

// NewDirectiveTransformer creates a new transformer
func NewDirectiveTransformer(registryPath string) parser.ASTTransformer {
	return &DirectiveTransformer{
		RegistryPath: registryPath,
	}
}

// Transform expands directives in the AST
func (t *DirectiveTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch block := n.(type) {
		case *ListBlock:
			t.expandListBlock(block, node, reader)
		case *NewItemBlock:
			// Keep as-is, content already parsed as children
		}

		return ast.WalkContinue, nil
	})
}

// expandListBlock expands a :::list block by loading registry files
func (t *DirectiveTransformer) expandListBlock(listBlock *ListBlock, doc *ast.Document, reader text.Reader) {
	// Use the ItemType to determine which registry folder to use
	registryPath := filepath.Join(t.RegistryPath, listBlock.ItemType)

	// Load each item file and insert content
	for _, itemName := range listBlock.Names {
		content, err := t.loadItemContent(registryPath, itemName)
		if err != nil {
			// Silently skip missing items - validation can catch this later
			continue
		}

		// Add a heading with the item name
		heading := ast.NewHeading(3)
		heading.AppendChild(heading, ast.NewString([]byte(itemName)))
		listBlock.AppendChild(listBlock, heading)

		// Add the content as a paragraph
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte(content)))
		listBlock.AppendChild(listBlock, para)
	}
}

// loadItemContent loads an item file from the registry
func (t *DirectiveTransformer) loadItemContent(registryPath, name string) (string, error) {
	itemPath := filepath.Join(registryPath, name+".md")

	data, err := os.ReadFile(itemPath)
	if err != nil {
		return "", err
	}

	// Extract frontmatter and content
	_, content := extractFrontmatter(data)
	return strings.TrimSpace(string(content)), nil
}

// extractFrontmatter separates YAML frontmatter from markdown content
func extractFrontmatter(data []byte) (*ItemMeta, []byte) {
	lines := strings.Split(string(data), "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return nil, data
	}

	// Find closing ---
	var endIdx int
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == 0 {
		return nil, data
	}

	// Parse frontmatter
	frontmatterYAML := strings.Join(lines[1:endIdx], "\n")
	var meta ItemMeta
	yaml.Unmarshal([]byte(frontmatterYAML), &meta)

	// Return content after frontmatter
	content := strings.Join(lines[endIdx+1:], "\n")
	return &meta, []byte(content)
}
