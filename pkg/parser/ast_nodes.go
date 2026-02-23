package parser

import (
	"github.com/yuin/goldmark/ast"
)

// ListBlock represents :::list TYPE ... :::end or :::include:TYPE name
type ListBlock struct {
	ast.BaseBlock
	ItemType     string   // "rules", "workflows", "guidelines"
	Names        []string // Item names to load
	IsSingleItem bool     // True for :::include (no :::end needed)
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

// NewListBlock creates a new ListBlock
func NewListBlock(itemType string) *ListBlock {
	return &ListBlock{
		ItemType: itemType,
		Names:    []string{},
	}
}

// NewItemBlock represents :::new:TYPE name=foo ... :::end
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
	SearchPath string // Optional: custom path to search for docs (default: "docs")
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

// NewDocsBlock creates a new DocsBlock
func NewDocsBlock(searchPath string) *DocsBlock {
	if searchPath == "" {
		searchPath = "docs"
	}
	return &DocsBlock{
		SearchPath: searchPath,
	}
}

// SkillsBlock represents :::skills [path] directive
// Expands to <available_skills> XML block for Agent Skills integration
type SkillsBlock struct {
	ast.BaseBlock
	SearchPath string // Path to search for linked skills (default: auto-detect)
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

// NewSkillsBlock creates a new SkillsBlock
func NewSkillsBlock(searchPath string) *SkillsBlock {
	return &SkillsBlock{
		SearchPath: searchPath,
	}
}
