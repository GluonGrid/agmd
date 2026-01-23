package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// AddToDirective adds an item to AGENTS.md by either:
// 1. Appending to first existing :::list TYPE block
// 2. Creating a new ## Section with :::include TYPE:name
func AddToDirective(content []byte, itemType, name string) ([]byte, error) {
	// Try to find existing :::list TYPE block
	listPattern := regexp.MustCompile(`(?m)^:::list\s+` + regexp.QuoteMeta(itemType) + `$`)
	endPattern := regexp.MustCompile(`(?m)^:::end$`)

	listMatch := listPattern.FindIndex(content)
	if listMatch != nil {
		// Found a list block - find its :::end
		searchStart := listMatch[1]
		endMatch := endPattern.FindIndex(content[searchStart:])
		if endMatch == nil {
			return nil, fmt.Errorf("found :::list %s but no matching :::end", itemType)
		}

		// Insert name before :::end
		insertPos := searchStart + endMatch[0]

		// Check if name already exists in the list
		listContent := content[listMatch[1]:insertPos]
		if containsLine(listContent, name) {
			return nil, fmt.Errorf("%s '%s' already exists in list", itemType, name)
		}

		var buf bytes.Buffer
		buf.Write(content[:insertPos])
		buf.WriteString(name + "\n")
		buf.Write(content[insertPos:])
		return buf.Bytes(), nil
	}

	// No list found - create ## Section with :::include
	sectionTitle := strings.Title(itemType)
	includeDirective := fmt.Sprintf(":::include %s:%s", itemType, name)

	// Check if include already exists
	includePattern := regexp.MustCompile(`(?m)^:::include:` + regexp.QuoteMeta(itemType) + `\s+` + regexp.QuoteMeta(name) + `$`)
	if includePattern.Match(content) {
		return nil, fmt.Errorf("%s '%s' already included", itemType, name)
	}

	// Append section at end
	var buf bytes.Buffer
	buf.Write(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString("\n")
	buf.WriteString("## " + sectionTitle + "\n\n")
	buf.WriteString(includeDirective + "\n")

	return buf.Bytes(), nil
}

// RemoveFromDirective removes an item from AGENTS.md by:
// 1. Removing from :::list TYPE block if present
// 2. Removing :::include TYPE:name line if present
func RemoveFromDirective(content []byte, itemType, name string) ([]byte, error) {
	// Try to remove from :::list TYPE block first
	listPattern := regexp.MustCompile(`(?m)^:::list\s+` + regexp.QuoteMeta(itemType) + `$`)
	endPattern := regexp.MustCompile(`(?m)^:::end$`)

	listMatch := listPattern.FindIndex(content)
	if listMatch != nil {
		searchStart := listMatch[1]
		endMatch := endPattern.FindIndex(content[searchStart:])
		if endMatch != nil {
			// Found list block
			listStart := listMatch[1]
			listEnd := searchStart + endMatch[0]

			// Remove line with name
			removed, found := removeLine(content[listStart:listEnd], name)
			if found {
				var buf bytes.Buffer
				buf.Write(content[:listStart])
				buf.Write(removed)
				buf.Write(content[listEnd:])
				return buf.Bytes(), nil
			}
		}
	}

	// Try to remove :::include:TYPE name
	includePattern := regexp.MustCompile(`(?m)^:::include:` + regexp.QuoteMeta(itemType) + `\s+` + regexp.QuoteMeta(name) + `\n?`)
	if includePattern.Match(content) {
		return includePattern.ReplaceAll(content, nil), nil
	}

	return nil, fmt.Errorf("%s '%s' not found in AGENTS.md", itemType, name)
}

// containsLine checks if a line exists in content
func containsLine(content []byte, line string) bool {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == line {
			return true
		}
	}
	return false
}

// removeLine removes a line from content
func removeLine(content []byte, line string) ([]byte, bool) {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(content))
	found := false
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) == line && !found {
			found = true
			continue // Skip this line
		}
		buf.WriteString(text + "\n")
	}
	return buf.Bytes(), found
}
