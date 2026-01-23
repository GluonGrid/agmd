package parser

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type directiveParser struct{}

var defaultDirectiveParser = &directiveParser{}

// NewDirectiveParser creates a new directive parser for Goldmark
func NewDirectiveParser() parser.BlockParser {
	return defaultDirectiveParser
}

type directiveData struct {
	node ast.Node
}

var directiveDataKey = parser.NewContextKey()

func (b *directiveParser) Trigger() []byte {
	return []byte{':'}
}

func (b *directiveParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()

	// Must start with ::: at beginning of line
	if !bytes.HasPrefix(line, []byte(":::")) {
		return nil, parser.NoChildren
	}

	// Match :::include TYPE:NAME (treat as having children to force Continue to be called)
	// Example: :::include rule:typescript
	includeRe := regexp.MustCompile(`^:::include\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	if match := includeRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rule", "workflow"
		name := string(match[2])      // "typescript"

		// Create a single-item list block
		node := NewListBlock(itemType)
		node.Names = []string{name}
		node.IsSingleItem = true

		pc.Set(directiveDataKey, &directiveData{node})

		// Advance past the entire line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		
		// Return NoChildren with Continue - this might help
		return node, parser.NoChildren | parser.Continue
	}

	// Match :::list TYPE (multi-line, needs :::end)
	// Example: :::list rule
	listRe := regexp.MustCompile(`^:::list\s+([a-z0-9-]+)`)
	if match := listRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rule", "workflow"

		node := NewListBlock(itemType)
		node.IsSingleItem = false

		pc.Set(directiveDataKey, &directiveData{node})

		// Advance past the :::list TYPE line
		newline := 1
		if len(line) > 0 && line[len(line)-1] != '\n' {
			newline = 0
		}
		reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
		return node, parser.NoChildren
	}

	// Match :::new TYPE:NAME (inline definition, needs :::end)
	// Example: :::new rule:my-auth-rule
	newRe := regexp.MustCompile(`^:::new\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	if match := newRe.FindSubmatch(line); match != nil {
		itemType := string(match[1]) // "rule", "workflow"
		name := string(match[2])      // "my-auth-rule"

		node := NewNewItemBlock(itemType, name)

		pc.Set(directiveDataKey, &directiveData{node})

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

func (b *directiveParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)

	// For single-item includes, close immediately
	if listBlock, ok := node.(*ListBlock); ok && listBlock.IsSingleItem {
		return parser.Close
	}

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

func (b *directiveParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	data := pc.Get(directiveDataKey)
	if data != nil {
		ddata := data.(*directiveData)
		if ddata.node == node {
			pc.Set(directiveDataKey, nil)
		}
	}
}

func (b *directiveParser) CanInterruptParagraph() bool {
	return true
}

func (b *directiveParser) CanAcceptIndentedLine() bool {
	return false
}
