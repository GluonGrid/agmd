package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func main() {
	fmt.Println("=== Goldmark POC: Directive Parsing ===\n")

	// Read template file (from args or default)
	templatePath := "templates/template-v4.md"
	if len(os.Args) > 1 {
		templatePath = os.Args[1]
	}
	source, err := os.ReadFile(templatePath)
	if err != nil {
		fmt.Printf("Error reading template: %v\n", err)
		return
	}

	fmt.Println("Input (pre-rendered.md):")
	fmt.Println("------------------------")
	fmt.Println(string(source))
	fmt.Println()

	// Create Goldmark parser with custom extension (v4 with :::include, :::list, :::new)
	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(NewDirectiveParserV4(), 100),
			),
			parser.WithASTTransformers(
				util.Prioritized(NewDirectiveTransformer("registry"), 100),
			),
		),
	)

	// Parse the markdown
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	fmt.Println("AST Structure:")
	fmt.Println("--------------")
	doc.Dump(source, 0)
	fmt.Println()

	// Render back to clean Markdown
	var buf bytes.Buffer
	renderer := NewMarkdownRenderer()
	if err := renderer.Render(&buf, source, doc); err != nil {
		fmt.Printf("Error rendering: %v\n", err)
		return
	}

	fmt.Println("Output (AGENTS.md):")
	fmt.Println("-------------------")
	fmt.Println(buf.String())

	// Write output
	outputPath := "output.md"
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error writing output: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Output written to %s\n", outputPath)
}
