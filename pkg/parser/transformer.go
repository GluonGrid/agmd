package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agmd/pkg/skills"

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
	RegistryPath string // global ~/.agmd
	LocalPath    string // project-local .agmd (optional, checked first)
}

// NewDirectiveTransformer creates a new transformer
func NewDirectiveTransformer(registryPath, localPath string) parser.ASTTransformer {
	return &DirectiveTransformer{
		RegistryPath: registryPath,
		LocalPath:    localPath,
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
		case *SkillsBlock:
			t.expandSkillsBlock(block)
		case *NewItemBlock:
			// Keep as-is, content already parsed as children
		}

		return ast.WalkContinue, nil
	})
}

// expandListBlock expands a :::list block by loading registry files
func (t *DirectiveTransformer) expandListBlock(listBlock *ListBlock, doc *ast.Document, reader text.Reader) {
	// Load each item file and insert content
	for _, item := range listBlock.Items {
		content, err := t.loadItemContent(item.ItemType, item.Name)
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

// loadItemContent loads an item file, checking local registry first then global
func (t *DirectiveTransformer) loadItemContent(itemType, name string) (string, error) {
	readItem := func(base string) (string, error) {
		itemPath := filepath.Join(base, itemType, name+".md")
		data, err := os.ReadFile(itemPath)
		if err != nil {
			return "", err
		}
		_, content := extractFrontmatter(data)
		return strings.TrimSpace(string(content)), nil
	}

	// Try local first
	if t.LocalPath != "" {
		if content, err := readItem(t.LocalPath); err == nil {
			return content, nil
		}
	}

	// Fall back to global
	return readItem(t.RegistryPath)
}

// expandDocsBlock expands :::docs ... :::end into a formatted doc index.
// Names listed in the block are resolved as symlinks under ./docs/.
func (t *DirectiveTransformer) expandDocsBlock(docsBlock *DocsBlock) {
	const searchPath = "docs"

	if len(docsBlock.Names) == 0 {
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte("*No documentation linked. Use `agmd doc link <name>` to add documentation.*")))
		docsBlock.AppendChild(docsBlock, para)
		return
	}

	var docs []docInfo
	for _, name := range docsBlock.Names {
		entryPath := filepath.Join(searchPath, name)
		fileInfo, err := os.Lstat(entryPath)
		if err != nil {
			continue
		}

		var target string
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			target, err = filepath.EvalSymlinks(entryPath)
			if err != nil {
				continue
			}
		} else if fileInfo.IsDir() {
			target = entryPath
		} else {
			continue
		}

		targetInfo, err := os.Stat(target)
		if err != nil || !targetInfo.IsDir() {
			continue
		}

		docs = append(docs, docInfo{
			Name:      name,
			Path:      entryPath,
			Target:    target,
			FileCount: countFilesInDir(target),
		})
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

	pathNote := ast.NewParagraph()
	pathNote.AppendChild(pathNote, ast.NewString([]byte("*Located in `"+searchPath+"/`*")))
	docsBlock.AppendChild(docsBlock, pathNote)

	for _, doc := range docs {
		description := getDocDescription(doc.Target)

		var headerText string
		if description != "" {
			headerText = "**" + doc.Name + "** - " + description
		} else {
			headerText = "**" + doc.Name + "** (`" + doc.Path + "`, " + itoa(doc.FileCount) + " files)"
		}

		headerPara := ast.NewParagraph()
		headerPara.AppendChild(headerPara, ast.NewString([]byte(headerText)))
		docsBlock.AppendChild(docsBlock, headerPara)

		tree := buildCompactTree(doc.Target)
		if tree != "" {
			treePara := ast.NewParagraph()
			treePara.AppendChild(treePara, ast.NewString([]byte(tree)))
			docsBlock.AppendChild(docsBlock, treePara)
		}
	}
}

// expandSkillsBlock expands :::skills ... :::end into an <available_skills> XML block.
// Names listed in the block are resolved from the linked skills directory.
func (t *DirectiveTransformer) expandSkillsBlock(skillsBlock *SkillsBlock) {
	if len(skillsBlock.Names) == 0 {
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte("*No skills linked. Use `agmd skill link <name>` to add a skill.*")))
		skillsBlock.AppendChild(skillsBlock, para)
		return
	}

	// Auto-detect skills directory
	var searchPath string
	if _, err := os.Stat(".claude/skills"); err == nil {
		searchPath = ".claude/skills"
	} else if _, err := os.Stat(".agents/skills"); err == nil {
		searchPath = ".agents/skills"
	}
	// Build a map of all linked skills for lookup
	allLinked, err := skills.FindLinkedSkills(searchPath)
	_ = err
	linkedByName := make(map[string]*skills.LinkedSkill, len(allLinked))
	for _, s := range allLinked {
		linkedByName[s.Name] = s
	}

	// Expand only the names listed in the block, in order
	var linked []*skills.LinkedSkill
	for _, name := range skillsBlock.Names {
		if s, ok := linkedByName[name]; ok {
			linked = append(linked, s)
		}
	}
	if len(linked) == 0 {
		para := ast.NewParagraph()
		para.AppendChild(para, ast.NewString([]byte("*No skills linked. Use `agmd skill link <name>` to add a skill.*")))
		skillsBlock.AppendChild(skillsBlock, para)
		return
	}

	// Build <available_skills> XML block
	var sb strings.Builder
	sb.WriteString("<available_skills>\n")
	for _, s := range linked {
		// Resolve absolute path for location field
		absPath := s.SkillMDPath
		if !filepath.IsAbs(absPath) {
			if abs, err := filepath.Abs(absPath); err == nil {
				absPath = abs
			}
		}
		sb.WriteString(fmt.Sprintf("  <skill>\n"))
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", xmlEscape(s.Name)))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", xmlEscape(s.Description)))
		sb.WriteString(fmt.Sprintf("    <location>%s</location>\n", absPath))
		sb.WriteString(fmt.Sprintf("  </skill>\n"))
	}
	sb.WriteString("</available_skills>")

	para := ast.NewParagraph()
	para.AppendChild(para, ast.NewString([]byte(sb.String())))
	skillsBlock.AppendChild(skillsBlock, para)
}

// xmlEscape escapes special XML characters
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
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

// getDocDescription reads the first line of description.md if it exists
func getDocDescription(docPath string) string {
	descPath := filepath.Join(docPath, "description.md")
	data, err := os.ReadFile(descPath)
	if err != nil {
		return ""
	}
	// Get first non-empty line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

// buildCompactTree builds a compact tree representation of a directory
// Format: folder/*.ext: {file1,file2,file3} or folder/: {sub1,sub2}/*.ext
func buildCompactTree(rootPath string) string {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return ""
	}

	var lines []string
	for _, entry := range entries {
		// Skip description.md
		if entry.Name() == "description.md" {
			continue
		}

		if entry.IsDir() {
			subPath := filepath.Join(rootPath, entry.Name())
			subTree := buildDirCompact(subPath, entry.Name())
			if subTree != "" {
				lines = append(lines, "  "+subTree)
			}
		} else {
			// Root-level file - just add it
			lines = append(lines, "  "+entry.Name())
		}
	}

	return strings.Join(lines, "\n")
}

// buildDirCompact builds a compact representation of a directory
func buildDirCompact(dirPath string, dirName string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	// Separate files and subdirectories
	var files []string
	var subdirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}

	// Group files by extension
	extGroups := make(map[string][]string)
	for _, f := range files {
		ext := filepath.Ext(f)
		name := strings.TrimSuffix(f, ext)
		if ext == "" {
			ext = "(no ext)"
		}
		extGroups[ext] = append(extGroups[ext], name)
	}

	var parts []string

	// Build compact file list grouped by extension
	for ext, names := range extGroups {
		if ext == "(no ext)" {
			parts = append(parts, strings.Join(names, ", "))
		} else if len(names) == 1 {
			parts = append(parts, names[0]+ext)
		} else {
			parts = append(parts, "{"+strings.Join(names, ",")+"}"+ext)
		}
	}

	// Build subdirectory representations
	for _, subdir := range subdirs {
		subPath := filepath.Join(dirPath, subdir)
		subCompact := buildSubdirCompact(subPath, subdir)
		if subCompact != "" {
			parts = append(parts, subCompact)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return dirName + "/: " + strings.Join(parts, ", ")
}

// buildSubdirCompact builds compact representation of a subdirectory (one level)
func buildSubdirCompact(dirPath string, dirName string) string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return ""
	}

	var files []string
	var subdirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			subdirs = append(subdirs, entry.Name()+"/")
		} else {
			files = append(files, entry.Name())
		}
	}

	// Group files by extension
	extGroups := make(map[string][]string)
	for _, f := range files {
		ext := filepath.Ext(f)
		name := strings.TrimSuffix(f, ext)
		if ext == "" {
			ext = "(no ext)"
		}
		extGroups[ext] = append(extGroups[ext], name)
	}

	var parts []string

	// Build compact file list
	for ext, names := range extGroups {
		if ext == "(no ext)" {
			parts = append(parts, strings.Join(names, ", "))
		} else if len(names) == 1 {
			parts = append(parts, names[0]+ext)
		} else {
			parts = append(parts, "{"+strings.Join(names, ",")+"}"+ext)
		}
	}

	// Add subdirectory markers
	parts = append(parts, subdirs...)

	if len(parts) == 0 {
		return ""
	}

	return dirName + "/{" + strings.Join(parts, ", ") + "}"
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
