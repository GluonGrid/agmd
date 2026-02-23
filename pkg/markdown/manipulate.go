package markdown

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// AddToDirective adds an item to directives.md by either:
// 1. Appending to existing :::list block as type:name line
// 2. Appending a :::use type:name line at the end
func AddToDirective(content []byte, itemType, name string) ([]byte, error) {
	entry := fmt.Sprintf("%s:%s", itemType, name)

	// Try to find existing :::list block
	listPattern := regexp.MustCompile(`(?m)^:::list\s*$`)
	endPattern := regexp.MustCompile(`(?m)^:::end$`)

	listMatch := listPattern.FindIndex(content)
	if listMatch != nil {
		// Found a list block - find its :::end
		searchStart := listMatch[1]
		endMatch := endPattern.FindIndex(content[searchStart:])
		if endMatch == nil {
			return nil, fmt.Errorf("found :::list but no matching :::end")
		}

		// Insert entry before :::end
		insertPos := searchStart + endMatch[0]

		// Check if entry already exists in the list
		listContent := content[listMatch[1]:insertPos]
		if containsLine(listContent, entry) {
			return nil, fmt.Errorf("%s '%s' already exists in list", itemType, name)
		}

		var buf bytes.Buffer
		buf.Write(content[:insertPos])
		buf.WriteString(entry + "\n")
		buf.Write(content[insertPos:])
		return buf.Bytes(), nil
	}

	// No list found - check if :::use already exists
	usePattern := regexp.MustCompile(`(?m)^:::use\s+` + regexp.QuoteMeta(entry) + `\s*$`)
	if usePattern.Match(content) {
		return nil, fmt.Errorf("%s '%s' already included", itemType, name)
	}

	// Append :::use line at end
	var buf bytes.Buffer
	buf.Write(content)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		buf.WriteByte('\n')
	}
	buf.WriteString(":::use " + entry + "\n")

	return buf.Bytes(), nil
}

// RemoveFromDirective removes an item from directives.md by:
// 1. Removing type:name from :::list block if present
// 2. Removing :::use type:name line if present
func RemoveFromDirective(content []byte, itemType, name string) ([]byte, error) {
	entry := fmt.Sprintf("%s:%s", itemType, name)

	// Try to remove from :::list block first
	listPattern := regexp.MustCompile(`(?m)^:::list\s*$`)
	endPattern := regexp.MustCompile(`(?m)^:::end$`)

	listMatch := listPattern.FindIndex(content)
	if listMatch != nil {
		searchStart := listMatch[1]
		endMatch := endPattern.FindIndex(content[searchStart:])
		if endMatch != nil {
			listStart := listMatch[1]
			listEnd := searchStart + endMatch[0]

			removed, found := removeLine(content[listStart:listEnd], entry)
			if found {
				var buf bytes.Buffer
				buf.Write(content[:listStart])
				buf.Write(removed)
				buf.Write(content[listEnd:])
				return buf.Bytes(), nil
			}
		}
	}

	// Try to remove :::use type:name line
	usePattern := regexp.MustCompile(`(?m)^:::use\s+` + regexp.QuoteMeta(entry) + `\s*\n?`)
	if usePattern.Match(content) {
		return usePattern.ReplaceAll(content, nil), nil
	}

	return nil, fmt.Errorf("%s '%s' not found in directives.md", itemType, name)
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
