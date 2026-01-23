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
