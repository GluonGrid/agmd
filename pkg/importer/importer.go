package importer

import (
	"fmt"
	"regexp"
	"strings"
)

// ImportedItem represents a rule/workflow/guideline extracted from AGENTS.md
type ImportedItem struct {
	Type    string // "rule", "workflow", "guideline"
	Name    string
	Content string // Full markdown content
}

// DirectivesSection represents a section in directives.md
type DirectivesSection struct {
	HeaderLine int    // Line number where ## Header appears
	HeaderText string // "Code Quality", "Rules", etc.
	ItemType   string // "rule", "workflow", "guideline" (if contains :::list/:::include)
	ItemNames  []string
}

// MatchDirectivesWithAgents parses both files and matches items
// Returns: map[type][]ImportedItem, warnings, error
func MatchDirectivesWithAgents(directivesContent, agentsContent string) (map[string][]ImportedItem, []string, error) {
	var warnings []string

	// Step 1: Parse directives.md structure to understand document layout
	sections, err := parseDirectivesStructure(directivesContent)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse directives.md structure: %w", err)
	}

	// Step 2: Parse AGENTS.md using directives structure as guide
	result, err := extractItemsFromAgents(agentsContent, sections)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract items from AGENTS.md: %w", err)
	}

	return result, warnings, nil
}

// parseDirectivesStructure extracts the document structure including section headers
// and what items they contain
func parseDirectivesStructure(content string) ([]DirectivesSection, error) {
	var sections []DirectivesSection

	lines := strings.Split(content, "\n")
	h2Re := regexp.MustCompile(`^## (.+)$`)
	listRe := regexp.MustCompile(`^:::list\s+([a-z0-9-]+)\s*$`)
	includeRe := regexp.MustCompile(`^:::include\s+([a-z0-9-]+):([a-z0-9/_-]+)\s*$`)
	endRe := regexp.MustCompile(`^:::end\s*$`)

	var currentSection *DirectivesSection
	var inList bool
	var listType string

	for i, line := range lines {
		// Detect ## Section headers
		if match := h2Re.FindStringSubmatch(line); match != nil {
			// Save previous section if exists
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}

			// Start new section
			currentSection = &DirectivesSection{
				HeaderLine: i,
				HeaderText: match[1],
				ItemNames:  []string{},
			}
			inList = false
			continue
		}

		// Detect :::list TYPE
		if match := listRe.FindStringSubmatch(line); match != nil {
			inList = true
			listType = match[1]
			if currentSection != nil {
				currentSection.ItemType = listType
			}
			continue
		}

		// Detect :::include TYPE:NAME
		if match := includeRe.FindStringSubmatch(line); match != nil {
			itemType := match[1]
			itemName := match[2]
			if currentSection != nil {
				currentSection.ItemType = itemType
				currentSection.ItemNames = append(currentSection.ItemNames, itemName)
			}
			continue
		}

		// Detect :::end
		if endRe.MatchString(line) {
			inList = false
			listType = ""
			continue
		}

		// If we're inside a :::list block, collect item names
		if inList && currentSection != nil {
			name := strings.TrimSpace(line)
			if name != "" {
				currentSection.ItemNames = append(currentSection.ItemNames, name)
			}
		}
	}

	// Save last section
	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	return sections, nil
}

// extractItemsFromAgents extracts item content from AGENTS.md using directives structure
func extractItemsFromAgents(agentsContent string, sections []DirectivesSection) (map[string][]ImportedItem, error) {
	result := make(map[string][]ImportedItem)

	lines := strings.Split(agentsContent, "\n")
	h2Re := regexp.MustCompile(`^## (.+)$`)
	h3Re := regexp.MustCompile(`^### ([a-z0-9-_]+)\s*$`)

	// Build section map by header text
	sectionMap := make(map[string]DirectivesSection)
	for _, section := range sections {
		sectionMap[section.HeaderText] = section
	}

	var currentSection *DirectivesSection
	var currentItemName string
	var currentItemContent strings.Builder
	var inItem bool

	for i, line := range lines {
		// Check for ## Section headers
		if match := h2Re.FindStringSubmatch(line); match != nil {
			headerText := match[1]

			// Check if this header matches a section from directives
			if sec, ok := sectionMap[headerText]; ok {
				// This IS a section boundary from directives.md

				// Save previous item if exists
				if inItem && currentItemName != "" && currentSection != nil {
					result[currentSection.ItemType] = append(result[currentSection.ItemType], ImportedItem{
						Type:    currentSection.ItemType,
						Name:    currentItemName,
						Content: strings.TrimSpace(currentItemContent.String()),
					})
				}

				// Enter new section
				if sec.ItemType != "" {
					currentSection = &sec
				} else {
					currentSection = nil
				}

				inItem = false
				currentItemName = ""
				currentItemContent.Reset()
				continue
			} else {
				// This is NOT a section boundary - it's content inside an item
				// (like ## Purpose, ## Guidelines, etc.)
				// Just accumulate it like any other line
				if inItem {
					currentItemContent.WriteString(line)
					currentItemContent.WriteString("\n")
				}
				continue
			}
		}

		// Check for ### item headers (only if we're in a section with items)
		if currentSection != nil && currentSection.ItemType != "" {
			if match := h3Re.FindStringSubmatch(line); match != nil {
				// Save previous item if exists
				if inItem && currentItemName != "" {
					result[currentSection.ItemType] = append(result[currentSection.ItemType], ImportedItem{
						Type:    currentSection.ItemType,
						Name:    currentItemName,
						Content: strings.TrimSpace(currentItemContent.String()),
					})
				}

				// Start new item
				currentItemName = match[1]
				currentItemContent.Reset()
				inItem = true
				continue
			}
		}

		// Accumulate content if we're inside an item
		if inItem {
			currentItemContent.WriteString(line)
			currentItemContent.WriteString("\n")
		}

		_ = i // prevent unused variable error
	}

	// Save last item
	if inItem && currentItemName != "" && currentSection != nil {
		result[currentSection.ItemType] = append(result[currentSection.ItemType], ImportedItem{
			Type:    currentSection.ItemType,
			Name:    currentItemName,
			Content: strings.TrimSpace(currentItemContent.String()),
		})
	}

	return result, nil
}

// Legacy functions for backward compatibility

// parseDirectivesManifest extracts the list of items from directives.md
// Returns: map[type][]name
func parseDirectivesManifest(content string) (map[string][]string, error) {
	manifest := make(map[string][]string)

	// Match :::list TYPE ... :::end blocks
	listRe := regexp.MustCompile(`(?s):::list\s+([a-z0-9-]+)\s*\n(.*?)\n:::end`)
	listMatches := listRe.FindAllStringSubmatch(content, -1)

	for _, match := range listMatches {
		if len(match) < 3 {
			continue
		}
		itemType := match[1]
		items := match[2]

		// Split items by newline and trim
		lines := strings.Split(items, "\n")
		for _, line := range lines {
			name := strings.TrimSpace(line)
			if name == "" {
				continue
			}
			manifest[itemType] = append(manifest[itemType], name)
		}
	}

	// Match :::include TYPE:NAME
	includeRe := regexp.MustCompile(`:::include\s+([a-z0-9-]+):([a-z0-9/_-]+)`)
	includeMatches := includeRe.FindAllStringSubmatch(content, -1)

	for _, match := range includeMatches {
		if len(match) < 3 {
			continue
		}
		itemType := match[1]
		name := match[2]
		manifest[itemType] = append(manifest[itemType], name)
	}

	return manifest, nil
}

// extractItemContent is a helper function for testing
func extractItemContent(agentsContent, sectionTitle, itemName string) string {
	// Parse simple structure for testing
	lines := strings.Split(agentsContent, "\n")
	h2Re := regexp.MustCompile(`^## (.+)$`)
	h3Re := regexp.MustCompile(`^### ([a-z0-9-_]+)\s*$`)

	inSection := false
	inItem := false
	var content strings.Builder

	for _, line := range lines {
		if match := h2Re.FindStringSubmatch(line); match != nil {
			if match[1] == sectionTitle {
				inSection = true
			} else {
				if inItem {
					break // End of item
				}
				inSection = false
			}
			continue
		}

		if inSection {
			if match := h3Re.FindStringSubmatch(line); match != nil {
				if inItem && match[1] != itemName {
					break // Hit next item
				}
				if match[1] == itemName {
					inItem = true
				}
				continue
			}
		}

		if inItem {
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	return strings.TrimSpace(content.String())
}
