package generator

import (
	"strings"
)

const (
	customStartMarker = "<!-- agmd:custom-start -->"
	customEndMarker   = "<!-- agmd:custom-end -->"
)

// extractCustomSection extracts content between ACTUAL custom section markers
// (not example markers within the rule content)
func extractCustomSection(content string) string {
	// Find the LAST managed-end marker (actual section boundary, not examples in rules)
	managedEndIdx := strings.LastIndex(content, "<!-- agmd:managed-end -->")
	if managedEndIdx == -1 {
		return "" // No managed section, no custom section
	}

	// Look for custom markers AFTER the last managed-end marker
	remainingContent := content[managedEndIdx:]

	startIdx := strings.Index(remainingContent, customStartMarker)
	if startIdx == -1 {
		return "" // No custom section found
	}

	endIdx := strings.Index(remainingContent[startIdx:], customEndMarker)
	if endIdx == -1 {
		return "" // No ending marker found
	}

	// Extract everything from start marker to end marker (inclusive)
	endIdx += startIdx + len(customEndMarker)
	return remainingContent[startIdx:endIdx]
}
