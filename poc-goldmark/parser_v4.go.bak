package main

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type directiveParserV4 struct{}

var defaultDirectiveParserV4 = &directiveParserV4{}

func NewDirectiveParserV4() parser.BlockParser {
	return defaultDirectiveParserV4
}

type directiveDataV4 struct {
	node ast.Node
}

var directiveDataKeyV4 = parser.NewContextKey()

func (b *directiveParserV4) Trigger() []byte {
	return []byte{':'}
}

func (b *directiveParserV4) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()

	// Must start with ::: at beginning of line
	if !bytes.HasPrefix(line, []byte(":::")) {
		return nil, parser.NoChildren
	}

	// Match :::include:TYPE name (single item, no :::end needed)
	// Example: :::include:rule typescript
	includeRe := regexp.MustCompile(`^:::include:([a-z0-9-]+)\s+([a-z0-9-]+)`)
	if match := includeRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rule", "workflow"
		name := string(match[2])      // "typescript"

		// Pluralize type for folder name
		folderName := pluralize(itemType)

		// Create a single-item list block
		node := NewListBlock(folderName)
		node.Names = []string{name}
		node.IsSingleItem = true

		pc.Set(directiveDataKeyV4, &directiveDataV4{node})

		// Advance past the entire line (no :::end needed)
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return node, parser.NoChildren
	}

	// Match :::list:TYPE (multi-line, needs :::end)
	// Example: :::list:rules
	listRe := regexp.MustCompile(`^:::list:([a-z0-9-]+)`)
	if match := listRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rules", "workflows"

		node := NewListBlock(itemType)
		node.IsSingleItem = false

		pc.Set(directiveDataKeyV4, &directiveDataV4{node})

		// Advance past the :::list:TYPE line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return node, parser.NoChildren
	}

	// Match :::new:TYPE name=foo (inline definition, needs :::end)
	// Example: :::new:rule name=my-auth-rule
	newRe := regexp.MustCompile(`^:::new:([a-z0-9-]+)\s+name=([a-z0-9-]+)`)
	if match := newRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rule", "workflow"
		name := string(match[2])      // "my-auth-rule"

		node := NewNewItemBlock(itemType, name)

		pc.Set(directiveDataKeyV4, &directiveDataV4{node})

		// Advance past the :::new:TYPE line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return node, parser.HasChildren
	}

	return nil, parser.NoChildren
}

func (b *directiveParserV4) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)

	// Check for :::end
	if bytes.Equal(trimmed, []byte(":::end")) {
		// Advance past the :::end line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return parser.Close
	}

	// Handle ListBlock - collect item names
	if listBlock, ok := node.(*ListBlock); ok {
		name := string(trimmed)
		if name != "" && !bytes.HasPrefix(trimmed, []byte(":::")) {
			listBlock.Names = append(listBlock.Names, name)
		}
		// Advance to next line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return parser.Continue | parser.NoChildren
	}

	// Handle NewItemBlock - let Goldmark parse content as children
	// Don't advance here - let child parsers handle it
	return parser.Continue | parser.HasChildren
}

func (b *directiveParserV4) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	data := pc.Get(directiveDataKeyV4)
	if data != nil {
		ddata := data.(*directiveDataV4)
		if ddata.node == node {
			pc.Set(directiveDataKeyV4, nil)
		}
	}
}

func (b *directiveParserV4) CanInterruptParagraph() bool {
	return true
}

func (b *directiveParserV4) CanAcceptIndentedLine() bool {
	return false
}

// pluralize converts singular type to plural folder name
// rule -> rules, workflow -> workflows
func pluralize(s string) string {
	if len(s) > 0 && s[len(s)-1] != 's' {
		return s + "s"
	}
	return s
}
