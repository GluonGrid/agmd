package parser

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

// ListItem represents a single type:name pair in a :::list block
type ListItem struct {
	ItemType string // "rule", "workflow", "guideline", etc.
	Name     string // item name (may include path like "frontend/typescript")
}

// ListBlock represents a :::list ... :::end block or :::use type:name (single item)
type ListBlock struct {
	ast.BaseBlock
	Items       []ListItem // mixed type:name pairs
	isSingleUse bool       // true when created by :::use (single line, no :::end)
}

// KindListBlock is the kind of ListBlock
var KindListBlock = ast.NewNodeKind("ListBlock")

// Kind implements ast.Node
func (n *ListBlock) Kind() ast.NodeKind {
	return KindListBlock
}

// Dump implements ast.Node
func (n *ListBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// NewListBlock creates a new empty ListBlock
func NewListBlock() *ListBlock {
	return &ListBlock{
		Items: []ListItem{},
	}
}

// NewItemBlock represents :::new TYPE:name ... :::end
type NewItemBlock struct {
	ast.BaseBlock
	ItemType string // "rule", "workflow", "guideline"
	Name     string
}

// KindNewItemBlock is the kind of NewItemBlock
var KindNewItemBlock = ast.NewNodeKind("NewItemBlock")

// Kind implements ast.Node
func (n *NewItemBlock) Kind() ast.NodeKind {
	return KindNewItemBlock
}

// Dump implements ast.Node
func (n *NewItemBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// NewNewItemBlock creates a new NewItemBlock
func NewNewItemBlock(itemType, name string) *NewItemBlock {
	return &NewItemBlock{
		ItemType: itemType,
		Name:     name,
	}
}

// DocsBlock represents :::docs directive
// It expands to list available doc symlinks in the project
type DocsBlock struct {
	ast.BaseBlock
	SearchPath string   // Optional: custom path to search for docs (default: "docs")
	Names      []string // Optional: explicit doc names to include (filter); empty = all
}

// KindDocsBlock is the kind of DocsBlock
var KindDocsBlock = ast.NewNodeKind("DocsBlock")

// Kind implements ast.Node
func (n *DocsBlock) Kind() ast.NodeKind {
	return KindDocsBlock
}

// Dump implements ast.Node
func (n *DocsBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// NewDocsBlock creates a new DocsBlock.
// arg is either empty (auto-detect), a path (starts with . or /), or space-separated names.
func NewDocsBlock(arg string) *DocsBlock {
	if arg == "" {
		return &DocsBlock{SearchPath: "docs"}
	}
	// If it looks like a path, treat as search path override
	if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "../") {
		return &DocsBlock{SearchPath: arg}
	}
	// Otherwise treat as space-separated doc names (filter)
	names := strings.Fields(arg)
	return &DocsBlock{SearchPath: "docs", Names: names}
}

// SkillsBlock represents :::skills directive
// Expands to <available_skills> XML block for Agent Skills integration
type SkillsBlock struct {
	ast.BaseBlock
	SearchPath string   // Path to search for linked skills (default: auto-detect)
	Names      []string // Optional: explicit skill names to include (filter); empty = all
}

// KindSkillsBlock is the kind of SkillsBlock
var KindSkillsBlock = ast.NewNodeKind("SkillsBlock")

// Kind implements ast.Node
func (n *SkillsBlock) Kind() ast.NodeKind {
	return KindSkillsBlock
}

// Dump implements ast.Node
func (n *SkillsBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// NewSkillsBlock creates a new SkillsBlock.
// arg is either empty (auto-detect), a path (starts with . or /), or space-separated names.
func NewSkillsBlock(arg string) *SkillsBlock {
	if arg == "" {
		return &SkillsBlock{}
	}
	// If it looks like a path, treat as search path override
	if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "../") {
		return &SkillsBlock{SearchPath: arg}
	}
	// Otherwise treat as space-separated skill names (filter)
	names := strings.Fields(arg)
	return &SkillsBlock{Names: names}
}
