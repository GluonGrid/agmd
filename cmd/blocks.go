package cmd

import "regexp"

// NewBlock represents a single :::new block with its type and name
type NewBlock struct {
	Type string
	Name string
}

// NewBlocksContent represents detected :::new blocks in directives.md
type NewBlocksContent struct {
	// Items contains all detected :::new blocks with their type and name
	Items []NewBlock
}

// detectNewBlocks scans directives.md for :::new markers
func detectNewBlocks(content string) NewBlocksContent {
	result := NewBlocksContent{
		Items: []NewBlock{},
	}

	// Regex to match :::new TYPE:NAME blocks (any type allowed)
	// Example: :::new rule:simple-test, :::new prompt:code-review
	re := regexp.MustCompile(`(?m)^:::new\s+([a-z0-9_-]+):([a-z0-9/_-]+)\s*$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		itemType := match[1]
		name := match[2]

		// Check for duplicates
		found := false
		for _, item := range result.Items {
			if item.Type == itemType && item.Name == name {
				found = true
				break
			}
		}
		if !found {
			result.Items = append(result.Items, NewBlock{Type: itemType, Name: name})
		}
	}

	return result
}
