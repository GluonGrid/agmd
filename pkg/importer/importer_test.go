package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchDirectivesWithAgents(t *testing.T) {
	testCases := []struct {
		name          string
		testDir       string
		expectedItems map[string][]string // type -> names
	}{
		{
			name:    "Simple case - basic list matching",
			testDir: "case1_simple",
			expectedItems: map[string][]string{
				"rule":     {"typescript", "eslint"},
				"workflow": {"deploy"},
			},
		},
		{
			name:    "Custom text between sections",
			testDir: "case2_custom_text",
			expectedItems: map[string][]string{
				"rule":     {"typescript", "eslint"},
				"workflow": {"deploy", "test"},
			},
		},
		{
			name:    "Mixed includes and lists",
			testDir: "case3_mixed_includes",
			expectedItems: map[string][]string{
				"rule":     {"typescript", "eslint", "prettier"},
				"workflow": {"deploy"},
			},
		},
		{
			name:    "Nested headers inside rules",
			testDir: "case5_nested_headers",
			expectedItems: map[string][]string{
				"rule": {"typescript"},
			},
		},
		{
			name:    "Custom sections with items",
			testDir: "case_custom_sections",
			expectedItems: map[string][]string{
				"rule":     {"typescript", "eslint"},
				"workflow": {"deploy", "test"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testPath := filepath.Join("testdata", tc.testDir)

			directivesPath := filepath.Join(testPath, "directives.md")
			agentsPath := filepath.Join(testPath, "AGENTS.md")

			// Read files
			directivesContent, err := os.ReadFile(directivesPath)
			if err != nil {
				t.Fatalf("Failed to read directives.md: %v", err)
			}

			agentsContent, err := os.ReadFile(agentsPath)
			if err != nil {
				t.Fatalf("Failed to read AGENTS.md: %v", err)
			}

			// Parse and match
			result, warnings, err := MatchDirectivesWithAgents(
				string(directivesContent),
				string(agentsContent),
			)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(warnings) > 0 {
				t.Logf("Warnings: %v", warnings)
			}

			// Check extracted items match expected
			for itemType, expectedNames := range tc.expectedItems {
				actualItems, ok := result[itemType]
				if !ok {
					t.Errorf("Expected type %s not found in result", itemType)
					continue
				}

				if len(actualItems) != len(expectedNames) {
					t.Errorf("Type %s: expected %d items, got %d", itemType, len(expectedNames), len(actualItems))
					t.Logf("Expected: %v", expectedNames)
					t.Logf("Got: %v", extractNames(actualItems))
				}

				for _, expectedName := range expectedNames {
					found := false
					for _, item := range actualItems {
						if item.Name == expectedName {
							found = true
							// Verify content exists
							if item.Content == "" {
								t.Errorf("Item %s:%s has empty content", itemType, expectedName)
							} else {
								t.Logf("✓ Found %s:%s with content (%d bytes)", itemType, expectedName, len(item.Content))
							}
							break
						}
					}
					if !found {
						t.Errorf("Expected item %s:%s not found in result", itemType, expectedName)
					}
				}
			}
		})
	}
}

func TestContentBoundaryDetection(t *testing.T) {
	// This tests that custom ## headers INSIDE rules are included in content
	// but custom ## headers OUTSIDE (between rules) are not

	testPath := filepath.Join("testdata", "case_custom_sections")
	directivesPath := filepath.Join(testPath, "directives.md")
	agentsPath := filepath.Join(testPath, "AGENTS.md")

	directivesContent, _ := os.ReadFile(directivesPath)
	agentsContent, _ := os.ReadFile(agentsPath)

	result, _, err := MatchDirectivesWithAgents(
		string(directivesContent),
		string(agentsContent),
	)

	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	// Check typescript rule
	typescriptRule := findItem(result["rule"], "typescript")
	if typescriptRule == nil {
		t.Fatal("typescript rule not found")
	}

	// Content should include ## Purpose and ## Guidelines (they're INSIDE the rule)
	if !contains(typescriptRule.Content, "## Purpose") {
		t.Error("typescript rule should contain '## Purpose' header")
	}
	if !contains(typescriptRule.Content, "## Guidelines") {
		t.Error("typescript rule should contain '## Guidelines' header")
	}
	if !contains(typescriptRule.Content, "Enforce strict typing") {
		t.Error("typescript rule should contain purpose content")
	}

	// Content should NOT include "## Deployment Process" (that's a section boundary)
	if contains(typescriptRule.Content, "## Deployment Process") {
		t.Error("typescript rule should NOT contain '## Deployment Process'")
	}

	// Content should NOT include "Our deployment workflows:" (that's outside the rule)
	if contains(typescriptRule.Content, "Our deployment workflows") {
		t.Error("typescript rule should NOT contain text from next section")
	}

	t.Logf("TypeScript rule content:\n%s\n", typescriptRule.Content)
}

func TestParseDirectivesStructure(t *testing.T) {
	content := `# Project Directives

Welcome to my project!

## Code Quality

These rules ensure code quality:

:::list rule
typescript
eslint
:::end

## Deployment Process

Our deployment workflows:

:::list workflow
deploy
:::end

## Additional Notes

Some final notes.
`

	sections, err := parseDirectivesStructure(content)
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	if len(sections) != 3 {
		t.Errorf("Expected 3 sections, got %d", len(sections))
	}

	// Check first section
	if sections[0].HeaderText != "Code Quality" {
		t.Errorf("Expected 'Code Quality', got '%s'", sections[0].HeaderText)
	}
	if sections[0].ItemType != "rule" {
		t.Errorf("Expected type 'rule', got '%s'", sections[0].ItemType)
	}
	if len(sections[0].ItemNames) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sections[0].ItemNames))
	}

	// Check second section
	if sections[1].HeaderText != "Deployment Process" {
		t.Errorf("Expected 'Deployment Process', got '%s'", sections[1].HeaderText)
	}
	if sections[1].ItemType != "workflow" {
		t.Errorf("Expected type 'workflow', got '%s'", sections[1].ItemType)
	}

	// Check third section (no items)
	if sections[2].HeaderText != "Additional Notes" {
		t.Errorf("Expected 'Additional Notes', got '%s'", sections[2].HeaderText)
	}
	if sections[2].ItemType != "" {
		t.Errorf("Expected no item type, got '%s'", sections[2].ItemType)
	}

	t.Logf("Parsed %d sections successfully", len(sections))
	for i, sec := range sections {
		t.Logf("  [%d] %s (type: %s, items: %d)", i, sec.HeaderText, sec.ItemType, len(sec.ItemNames))
	}
}

// Helper functions

func extractNames(items []ImportedItem) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
	}
	return names
}

func findItem(items []ImportedItem, name string) *ImportedItem {
	for i := range items {
		if items[i].Name == name {
			return &items[i]
		}
	}
	return nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
