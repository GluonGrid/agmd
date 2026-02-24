package parser

import (
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

// DocsBlock represents a :::docs ... :::end block.
// Each line inside is a bare doc name. Expands to a formatted doc index.
//
//	:::docs
//	svelte-kit
//	typescript
//	:::end
type DocsBlock struct {
	ast.BaseBlock
	Names []string // doc names listed inside the block; empty = auto-discover
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

// NewDocsBlock creates an empty DocsBlock ready to collect names.
func NewDocsBlock() *DocsBlock {
	return &DocsBlock{}
}

// SkillsBlock represents a :::skills ... :::end block.
// Each line inside is a bare skill name. Expands to <available_skills> XML.
//
//	:::skills
//	managing-agmd
//	pdf-tools
//	:::end
type SkillsBlock struct {
	ast.BaseBlock
	Names []string // skill names listed inside the block; empty = auto-discover
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

// NewSkillsBlock creates an empty SkillsBlock ready to collect names.
func NewSkillsBlock() *SkillsBlock {
	return &SkillsBlock{}
}
