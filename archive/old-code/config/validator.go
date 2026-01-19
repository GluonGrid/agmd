package config

import (
	"fmt"
	"os"
)

// Validate validates a config file and its inheritance chain
func Validate(projectPath string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Parse project file
	projectFile, err := ParseFile(projectPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "parse_error",
			Message: fmt.Sprintf("Failed to parse project config: %v", err),
			Path:    projectPath,
		})
		return result
	}

	// Validate frontmatter
	if projectFile.Frontmatter.Version == "" {
		result.Warnings = append(result.Warnings, "No version specified in frontmatter")
	}

	// Check shared config exists (if specified)
	if projectFile.Frontmatter.Shared != "" {
		sharedPath := expandPath(projectFile.Frontmatter.Shared)
		if _, err := os.Stat(sharedPath); os.IsNotExist(err) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_file",
				Message: fmt.Sprintf("Shared config not found: %s", sharedPath),
				Path:    sharedPath,
			})
		} else {
			// Try to parse it
			if _, err := LoadShared(sharedPath); err != nil {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Type:    "invalid_shared",
					Message: fmt.Sprintf("Invalid shared config: %v", err),
					Path:    sharedPath,
				})
			}
		}
	}

	// Check each profile exists
	for _, profileName := range projectFile.Frontmatter.Profiles {
		if _, err := LoadProfile(profileName); err != nil {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_profile",
				Message: fmt.Sprintf("Profile '%s' not found", profileName),
				Path:    profileName,
			})
		}
	}

	// Validate overrides reference existing sections
	if len(projectFile.Frontmatter.Overrides) > 0 {
		// Try to resolve the full config to validate overrides
		resolved, err := Resolve(projectPath)
		if err == nil && resolved.Merged != nil {
			for overrideKey := range projectFile.Frontmatter.Overrides {
				// Parse section key from override
				if sectionKey := getSectionKeyFromOverride(overrideKey); sectionKey != "" {
					if FindSectionByKey(resolved.Merged.Sections, sectionKey) == nil {
						result.Warnings = append(result.Warnings,
							fmt.Sprintf("Override '%s' doesn't match any section", overrideKey))
					}
				}
			}
		}
	}

	// Check for common sections (warnings only)
	expectedSections := []string{
		"project-structure",
		"build-commands",
	}

	projectSectionKeys := make(map[string]bool)
	for _, section := range projectFile.Sections {
		projectSectionKeys[section.Key] = true
	}

	for _, expected := range expectedSections {
		if !projectSectionKeys[expected] {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Consider adding '%s' section", expected))
		}
	}

	return result
}

// getSectionKeyFromOverride extracts the section key from an override key
// Example: "code-quality.file-size-limit" -> "code-quality"
func getSectionKeyFromOverride(overrideKey string) string {
	if idx := len(overrideKey); idx > 0 {
		for i, c := range overrideKey {
			if c == '.' {
				return overrideKey[:i]
			}
		}
	}
	return ""
}

// ValidateProfile validates a profile file
func ValidateProfile(profilePath string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Parse profile file
	profileFile, err := ParseFile(profilePath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:    "parse_error",
			Message: fmt.Sprintf("Failed to parse profile: %v", err),
			Path:    profilePath,
		})
		return result
	}

	// Validate it's marked as a profile
	if profileFile.Frontmatter.Type != "" && profileFile.Frontmatter.Type != "profile" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Profile should have type: profile, got: %s", profileFile.Frontmatter.Type))
	}

	// Validate extends field
	if profileFile.Frontmatter.Extends == "" {
		result.Warnings = append(result.Warnings, "Profile should specify 'extends' field")
	}

	return result
}
