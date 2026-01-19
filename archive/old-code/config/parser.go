package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses an AGENTS.md file
func ParseFile(path string) (*AgmdFile, error) {
	// Read file contents
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract frontmatter and markdown content
	frontmatterBytes, markdownBytes, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	// Parse frontmatter
	var frontmatter AgmdFrontmatter
	if len(frontmatterBytes) > 0 {
		// Wrap in "agmd:" key if needed
		var yamlData map[string]any
		if err := yaml.Unmarshal(frontmatterBytes, &yamlData); err != nil {
			return nil, fmt.Errorf("invalid YAML frontmatter: %w", err)
		}

		// Check if "agmd" key exists
		if agmdData, ok := yamlData["agmd"]; ok {
			// Re-marshal just the agmd section
			agmdBytes, err := yaml.Marshal(agmdData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal agmd section: %w", err)
			}
			if err := yaml.Unmarshal(agmdBytes, &frontmatter); err != nil {
				return nil, fmt.Errorf("invalid agmd frontmatter: %w", err)
			}
		} else {
			// Try to parse directly (for backwards compatibility)
			if err := yaml.Unmarshal(frontmatterBytes, &frontmatter); err != nil {
				return nil, fmt.Errorf("invalid frontmatter: %w", err)
			}
		}
	}

	// Parse markdown sections
	markdown := string(markdownBytes)
	sections, err := parseSections(markdown)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sections: %w", err)
	}

	return &AgmdFile{
		Frontmatter: frontmatter,
		Content:     markdown,
		Sections:    sections,
		Path:        path,
	}, nil
}

// extractFrontmatter splits a file into YAML frontmatter and markdown content
func extractFrontmatter(content []byte) ([]byte, []byte, error) {
	// Look for YAML frontmatter: ---\n...\n---\n
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		// No frontmatter - return empty frontmatter and full content
		return nil, content, nil
	}

	// Find the end delimiter
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Scan() // Skip first ---

	var frontmatterLines []string
	lineNum := 1

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if line == "---" {
			// Found end delimiter
			frontmatterBytes := []byte(strings.Join(frontmatterLines, "\n"))

			// Everything after this line is markdown content
			remaining := content[scanner.Bytes()[0]:]
			// Skip past the closing --- and any following newlines
			afterDelimiter := bytes.SplitN(remaining, []byte("---"), 2)
			if len(afterDelimiter) > 1 {
				markdownBytes := bytes.TrimLeft(afterDelimiter[1], "\n\r")
				return frontmatterBytes, markdownBytes, nil
			}

			return frontmatterBytes, []byte{}, nil
		}

		frontmatterLines = append(frontmatterLines, line)
	}

	return nil, nil, fmt.Errorf("unclosed frontmatter (missing closing ---)")
}

// parseSections extracts markdown sections from content
func parseSections(markdown string) ([]Section, error) {
	var sections []Section

	// Regular expression to match markdown headings (## or ###)
	headingRegex := regexp.MustCompile(`(?m)^(#{2,3})\s+(.+?)$`)

	// Find all headings
	matches := headingRegex.FindAllStringSubmatchIndex(markdown, -1)
	if len(matches) == 0 {
		// No sections found - treat entire content as one section
		if strings.TrimSpace(markdown) != "" {
			sections = append(sections, Section{
				Title:   "",
				Level:   1,
				Content: markdown,
				Key:     "",
			})
		}
		return sections, nil
	}

	// Extract sections between headings
	for i, match := range matches {
		// match[0] = start of entire match
		// match[1] = end of entire match
		// match[2] = start of heading markers (##)
		// match[3] = end of heading markers
		// match[4] = start of title
		// match[5] = end of title

		level := match[3] - match[2] // Number of # characters
		title := markdown[match[4]:match[5]]

		// Find content: from end of this heading to start of next heading (or EOF)
		contentStart := match[1]
		contentEnd := len(markdown)
		if i < len(matches)-1 {
			contentEnd = matches[i+1][0]
		}

		content := strings.TrimRight(markdown[contentStart:contentEnd], "\n\r ")

		sections = append(sections, Section{
			Title:   strings.TrimSpace(title),
			Level:   level,
			Content: content,
			Key:     normalizeKey(title),
		})
	}

	return sections, nil
}

// normalizeKey converts a section title to a normalized key
// Example: "Code Quality Principles" -> "code-quality-principles"
func normalizeKey(title string) string {
	// Convert to lowercase
	key := strings.ToLower(title)

	// Remove special characters and replace spaces with hyphens
	key = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(key, "")
	key = regexp.MustCompile(`\s+`).ReplaceAllString(key, "-")
	key = regexp.MustCompile(`-+`).ReplaceAllString(key, "-")
	key = strings.Trim(key, "-")

	return key
}
