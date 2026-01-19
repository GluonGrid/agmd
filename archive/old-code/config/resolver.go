package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolve loads and merges all config layers (universal → profiles → project)
func Resolve(projectPath string) (*ResolvedConfig, error) {
	resolved := &ResolvedConfig{}

	// 1. Load project AGENTS.md
	projectFile, err := ParseFile(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}
	resolved.Project = projectFile

	// 2. Load universal shared config (if specified)
	if projectFile.Frontmatter.Shared != "" {
		sharedPath := expandPath(projectFile.Frontmatter.Shared)
		universalFile, err := LoadShared(sharedPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load shared config: %w", err)
		}
		resolved.Universal = universalFile
	}

	// 3. Load each profile (in order)
	for _, profileName := range projectFile.Frontmatter.Profiles {
		profileFile, err := LoadProfile(profileName)
		if err != nil {
			return nil, fmt.Errorf("failed to load profile '%s': %w", profileName, err)
		}
		resolved.Profiles = append(resolved.Profiles, profileFile)
	}

	// 4. Merge all layers
	merged, err := MergeConfigs(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	// 5. Apply overrides
	if len(projectFile.Frontmatter.Overrides) > 0 {
		merged = ApplyOverrides(merged, projectFile.Frontmatter.Overrides)
	}

	resolved.Merged = merged
	return resolved, nil
}

// LoadShared loads the universal shared config
func LoadShared(sharedPath string) (*AgmdFile, error) {
	// Expand path (handle ~, etc.)
	expandedPath := expandPath(sharedPath)

	// Parse file
	file, err := ParseFile(expandedPath)
	if err != nil {
		return nil, err
	}

	// Validate it's a universal config
	if file.Frontmatter.Type != "" && file.Frontmatter.Type != "universal" {
		return nil, fmt.Errorf("shared config must be type 'universal', got '%s'", file.Frontmatter.Type)
	}

	return file, nil
}

// LoadProfile loads a profile by name
func LoadProfile(name string) (*AgmdFile, error) {
	// Search paths in order:
	// 1. ~/.agmd/profiles/custom/{name}.md
	// 2. ~/.agmd/profiles/{name}.md
	// 3. Embedded assets (TODO: implement later)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	searchPaths := []string{
		filepath.Join(homeDir, ".agmd", "profiles", "custom", name+".md"),
		filepath.Join(homeDir, ".agmd", "profiles", name+".md"),
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			// File exists - parse it
			file, err := ParseFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to parse profile at %s: %w", path, err)
			}

			// Validate it's a profile
			if file.Frontmatter.Type != "" && file.Frontmatter.Type != "profile" {
				return nil, fmt.Errorf("profile must be type 'profile', got '%s'", file.Frontmatter.Type)
			}

			return file, nil
		}
	}

	return nil, fmt.Errorf("profile '%s' not found (searched: %s)", name, strings.Join(searchPaths, ", "))
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	// Handle ~ for home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}

// MergeConfigs merges all config layers into a single effective config
func MergeConfigs(resolved *ResolvedConfig) (*AgmdFile, error) {
	merged := &AgmdFile{
		Sections: []Section{},
	}

	// Start with universal config
	if resolved.Universal != nil {
		merged.Sections = append(merged.Sections, resolved.Universal.Sections...)
		merged.Content = resolved.Universal.Content
	}

	// Merge each profile (in order)
	for _, profile := range resolved.Profiles {
		merged.Sections = MergeSections(merged.Sections, profile.Sections)
	}

	// Merge project config
	if resolved.Project != nil {
		merged.Sections = MergeSections(merged.Sections, resolved.Project.Sections)
	}

	// Rebuild content from merged sections
	merged.Content = RebuildContent(merged.Sections)

	return merged, nil
}

// RebuildContent reconstructs markdown content from sections
func RebuildContent(sections []Section) string {
	var builder strings.Builder

	for _, section := range sections {
		if section.Title != "" {
			// Add heading
			heading := strings.Repeat("#", section.Level)
			builder.WriteString(fmt.Sprintf("%s %s\n", heading, section.Title))
		}
		// Add content
		builder.WriteString(section.Content)
		if !strings.HasSuffix(section.Content, "\n\n") {
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}
