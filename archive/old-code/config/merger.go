package config

import (
	"fmt"
	"regexp"
	"strings"
)

// MergeSections merges sections from overlay into base
// Overlay sections replace base sections with matching keys
func MergeSections(base, overlay []Section) []Section {
	// Create a map of base sections by key
	baseMap := make(map[string]Section)
	for _, section := range base {
		if section.Key != "" {
			baseMap[section.Key] = section
		}
	}

	// Track which keys we've seen in overlay
	seenKeys := make(map[string]bool)

	// Build result: start with overlay sections (they take precedence)
	var result []Section
	for _, section := range overlay {
		if section.Key != "" {
			seenKeys[section.Key] = true
		}
		result = append(result, section)
	}

	// Add base sections that weren't in overlay
	for _, section := range base {
		if section.Key != "" && !seenKeys[section.Key] {
			result = append(result, section)
		} else if section.Key == "" {
			// Keep sections without keys (like intro content)
			result = append(result, section)
		}
	}

	return result
}

// ApplyOverrides applies frontmatter overrides to sections
func ApplyOverrides(file *AgmdFile, overrides map[string]any) *AgmdFile {
	if len(overrides) == 0 {
		return file
	}

	// Create a copy to avoid modifying original
	result := &AgmdFile{
		Frontmatter: file.Frontmatter,
		Content:     file.Content,
		Path:        file.Path,
		Sections:    make([]Section, len(file.Sections)),
	}
	copy(result.Sections, file.Sections)

	// Apply each override
	for key, value := range overrides {
		result.Sections = applyOverride(result.Sections, key, value)
	}

	// Rebuild content
	result.Content = RebuildContent(result.Sections)

	return result
}

// applyOverride applies a single override to sections
// Override key format: "section-key.property-name"
func applyOverride(sections []Section, overrideKey string, value any) []Section {
	// Parse override key: "section-key.property-name"
	parts := strings.SplitN(overrideKey, ".", 2)
	if len(parts) != 2 {
		// Invalid override format - skip it
		return sections
	}

	sectionKey := parts[0]
	propertyName := parts[1]

	// Find matching section
	for i, section := range sections {
		if section.Key == sectionKey {
			// Apply override by replacing content
			sections[i].Content = applyPropertyOverride(section.Content, propertyName, value)
			break
		}
	}

	return sections
}

// applyPropertyOverride applies a property override to section content
func applyPropertyOverride(content, propertyName string, value any) string {
	// Convert value to string
	valueStr := fmt.Sprintf("%v", value)

	// Try to find and replace property in content
	// Pattern: "- PropertyName: oldvalue" or similar
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i)(-\s*)?%s[:\s]+[^\n]+`, regexp.QuoteMeta(propertyName)))

	if pattern.MatchString(content) {
		// Replace existing property
		replacement := fmt.Sprintf("${1}%s: %s", propertyName, valueStr)
		return pattern.ReplaceAllString(content, replacement)
	}

	// If property not found, append it to the section
	return content + fmt.Sprintf("\n- %s: %s\n", propertyName, valueStr)
}

// FindSectionByKey finds a section by its normalized key
func FindSectionByKey(sections []Section, key string) *Section {
	for _, section := range sections {
		if section.Key == key {
			return &section
		}
	}
	return nil
}

// GetSectionTitles returns a list of all section titles
func GetSectionTitles(sections []Section) []string {
	var titles []string
	for _, section := range sections {
		if section.Title != "" {
			titles = append(titles, section.Title)
		}
	}
	return titles
}
