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
		case *DocsBlock:
			t.expandDocsBlock(block)
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

		// Add the content directly without a heading
		// (user can include their own heading in the item content)
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

// expandDocsBlock expands :::docs by detecting doc symlinks in project
func (t *DirectiveTransformer) expandDocsBlock(docsBlock *DocsBlock) {
	searchPath := docsBlock.SearchPath

	// Check if the search path exists
	info, err := os.Stat(searchPath)
	if err != nil || !info.IsDir() {
		// No docs directory found, add a note
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte("*No documentation linked. Use `agmd doc link <name>` to add documentation.*")))
		docsBlock.AppendChild(docsBlock, para)
		return
	}

	// Find symlinks to doc folders
	entries, err := os.ReadDir(searchPath)
	if err != nil {
		return
	}

	var docs []docInfo
	for _, entry := range entries {
		entryPath := filepath.Join(searchPath, entry.Name())

		// Check if it's a symlink
		fileInfo, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			// It's a symlink - resolve it to absolute path
			target, err := filepath.EvalSymlinks(entryPath)
			if err != nil {
				continue
			}

			// Check if target is a directory
			targetInfo, err := os.Stat(target)
			if err != nil || !targetInfo.IsDir() {
				continue
			}

			// Count files in the doc (use resolved target path)
			fileCount := countFilesInDir(target)

			docs = append(docs, docInfo{
				Name:      entry.Name(),
				Path:      entryPath,
				Target:    target,
				FileCount: fileCount,
			})
		}
	}

	if len(docs) == 0 {
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte("*No documentation linked. Use `agmd doc link <name>` to add documentation.*")))
		docsBlock.AppendChild(docsBlock, para)
		return
	}

	// Build documentation list
	heading := ast.NewHeading(3)
	heading.AppendChild(heading, ast.NewString([]byte("Available Documentation")))
	docsBlock.AppendChild(docsBlock, heading)

	list := ast.NewList('-')
	for _, doc := range docs {
		item := ast.NewListItem(0)
		text := ast.NewString([]byte(
			"**" + doc.Name + "** - `" + doc.Path + "` (" + itoa(doc.FileCount) + " files)",
		))
		para := ast.NewParagraph()
		para.AppendChild(para, text)
		item.AppendChild(item, para)
		list.AppendChild(list, item)
	}
	docsBlock.AppendChild(docsBlock, list)
}

type docInfo struct {
	Name      string
	Path      string
	Target    string
	FileCount int
}

// countFilesInDir counts all files in a directory recursively
func countFilesInDir(path string) int {
	count := 0
	filepath.Walk(path, func(_ string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

// itoa converts int to string without importing strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
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
