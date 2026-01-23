package cmd

import "regexp"

// NewBlocksContent represents detected :::new blocks in directives.md
type NewBlocksContent struct {
	Rules      []string
	Workflows  []string
	Guidelines []string
}

// detectNewBlocks scans directives.md for :::new markers
func detectNewBlocks(content string) NewBlocksContent {
	result := NewBlocksContent{
		Rules:      []string{},
		Workflows:  []string{},
		Guidelines: []string{},
	}

	// Regex to match :::new:TYPE name=value blocks (parser syntax)
	// Example: :::new:rule name=simple-test
	re := regexp.MustCompile(`(?m)^:::new:(rule|workflow|guideline)\s+name=([a-z0-9/_-]+)\s*$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		itemType := match[1]
		name := match[2]

		switch itemType {
		case "rule":
			if !contains(result.Rules, name) {
				result.Rules = append(result.Rules, name)
			}
		case "workflow":
			if !contains(result.Workflows, name) {
				result.Workflows = append(result.Workflows, name)
			}
		case "guideline":
			if !contains(result.Guidelines, name) {
				result.Guidelines = append(result.Guidelines, name)
			}
		}
	}

	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
