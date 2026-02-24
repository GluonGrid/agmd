package parser

import (
	"bytes"
	"regexp"
	"strings"

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

// itemRe matches "type:name" entries; name may include path segments
var itemRe = regexp.MustCompile(`^([a-z0-9-]+):([a-z0-9/_-]+)$`)

// nameRe matches bare names (no colon) — used for :::docs and :::skills body lines
var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

func (b *directiveParser) Trigger() []byte {
	return []byte{':'}
}

func (b *directiveParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()

	// Must start with ::: at beginning of line
	if !bytes.HasPrefix(line, []byte(":::")) {
		return nil, parser.NoChildren
	}

	trimmed := bytes.TrimSpace(line)

	// :::use TYPE:NAME — single item, no :::end needed
	// Example: :::use rule:typescript
	useRe := regexp.MustCompile(`^:::use\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	if match := useRe.FindSubmatch(line); match != nil {
		itemType := string(match[1])
		name := string(match[2])

		node := NewListBlock()
		node.Items = []ListItem{{ItemType: itemType, Name: name}}
		node.isSingleUse = true

		pc.Set(directiveDataKey, &directiveData{node})
		advanceLine(reader, line, segment)
		return node, parser.NoChildren | parser.Continue
	}

	// :::list — mixed multi-line block, each line is type:name, needs :::end
	// Example:
	//   :::list
	//   rule:typescript
	//   workflow:commit
	//   :::end
	if bytes.Equal(trimmed, []byte(":::list")) {
		node := NewListBlock()
		pc.Set(directiveDataKey, &directiveData{node})
		advanceLine(reader, line, segment)
		return node, parser.NoChildren
	}

	// :::new TYPE:NAME — inline definition, needs :::end
	// Example: :::new rule:my-auth-rule
	newRe := regexp.MustCompile(`^:::new\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	if match := newRe.FindSubmatch(line); match != nil {
		itemType := string(match[1])
		name := string(match[2])

		node := NewNewItemBlock(itemType, name)
		pc.Set(directiveDataKey, &directiveData{node})
		advanceLine(reader, line, segment)
		return node, parser.HasChildren
	}

	// :::docs — multi-line block, each line is a bare doc name, needs :::end
	// Example:
	//   :::docs
	//   svelte-kit
	//   typescript
	//   :::end
	if bytes.Equal(trimmed, []byte(":::docs")) {
		node := NewDocsBlock()
		pc.Set(directiveDataKey, &directiveData{node})
		advanceLine(reader, line, segment)
		return node, parser.NoChildren
	}

	// :::skills — multi-line block, each line is a bare skill name, needs :::end
	// Example:
	//   :::skills
	//   managing-agmd
	//   pdf-tools
	//   :::end
	if bytes.Equal(trimmed, []byte(":::skills")) {
		node := NewSkillsBlock()
		pc.Set(directiveDataKey, &directiveData{node})
		advanceLine(reader, line, segment)
		return node, parser.NoChildren
	}

	return nil, parser.NoChildren
}

func (b *directiveParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	trimmed := bytes.TrimSpace(line)

	// :::use blocks (single item, pre-populated) close immediately
	if listBlock, ok := node.(*ListBlock); ok && listBlock.isSingleUse {
		return parser.Close
	}

	// Check for :::end — closes :::list, :::docs, :::skills, :::new
	if bytes.Equal(trimmed, []byte(":::end")) {
		advanceLine(reader, line, segment)
		return parser.Close
	}

	// Handle :::list block — collect type:name lines
	if listBlock, ok := node.(*ListBlock); ok {
		entry := string(trimmed)
		if entry != "" && !strings.HasPrefix(entry, ":::") {
			if match := itemRe.FindStringSubmatch(entry); match != nil {
				listBlock.Items = append(listBlock.Items, ListItem{
					ItemType: match[1],
					Name:     match[2],
				})
			}
			// silently skip lines that don't match type:name (comments, blanks, etc.)
		}
		advanceLine(reader, line, segment)
		return parser.Continue | parser.NoChildren
	}

	// Handle :::docs block — collect bare name lines
	if docsBlock, ok := node.(*DocsBlock); ok {
		name := string(trimmed)
		if name != "" && !strings.HasPrefix(name, ":::") {
			if nameRe.MatchString(name) {
				docsBlock.Names = append(docsBlock.Names, name)
			}
			// silently skip lines that don't look like names
		}
		advanceLine(reader, line, segment)
		return parser.Continue | parser.NoChildren
	}

	// Handle :::skills block — collect bare name lines
	if skillsBlock, ok := node.(*SkillsBlock); ok {
		name := string(trimmed)
		if name != "" && !strings.HasPrefix(name, ":::") {
			if nameRe.MatchString(name) {
				skillsBlock.Names = append(skillsBlock.Names, name)
			}
			// silently skip lines that don't look like names
		}
		advanceLine(reader, line, segment)
		return parser.Continue | parser.NoChildren
	}

	// Handle NewItemBlock — let Goldmark parse content as children
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

// advanceLine advances the reader past a single line
func advanceLine(reader text.Reader, line []byte, segment text.Segment) {
	newline := 1
	if len(line) > 0 && line[len(line)-1] != '\n' {
		newline = 0
	}
	reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
}
