package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// New creates a new Registry instance, auto-detecting a project-local .agmd/ in CWD
func New() (*Registry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".agmd")

	// Auto-detect project-local .agmd/ in CWD
	var localPath string
	cwd, err := os.Getwd()
	if err == nil {
		candidate := filepath.Join(cwd, ".agmd")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			localPath = candidate
		}
	}

	return &Registry{
		BasePath:  basePath,
		LocalPath: localPath,
	}, nil
}

// HasLocal returns true if a project-local registry is active
func (r *Registry) HasLocal() bool {
	return r.LocalPath != ""
}

// TypePath returns the path for a given type
func (r *Registry) TypePath(itemType string) string {
	return filepath.Join(r.BasePath, itemType)
}

// Exists checks if the registry exists
func (r *Registry) Exists() bool {
	info, err := os.Stat(r.BasePath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetItem retrieves an item by type and name, checking local registry first
func (r *Registry) GetItem(itemType, name string) (*Item, error) {
	itemPath := func(base string) string {
		if IsRawType(itemType) {
			return filepath.Join(base, itemType, name)
		}
		return filepath.Join(base, itemType, name+".md")
	}

	// Try local first
	if r.LocalPath != "" {
		path := itemPath(r.LocalPath)
		if _, err := os.Stat(path); err == nil {
			return loadItem(path, itemType)
		}
	}

	// Fall back to global
	path := itemPath(r.BasePath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s '%s' not found", itemType, name)
	}
	return loadItem(path, itemType)
}

// SaveItem saves an item to the registry
func (r *Registry) SaveItem(item Item) error {
	// Ensure type directory exists
	typeDir := filepath.Join(r.BasePath, item.Type)
	if err := os.MkdirAll(typeDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var path string
	var content string

	if IsRawType(item.Type) {
		// Raw types (file:) store content as-is without .md extension or frontmatter
		path = filepath.Join(typeDir, item.Name)
		content = item.Content
	} else {
		path = filepath.Join(typeDir, item.Name+".md")
		content = fmt.Sprintf(`---
name: %s
description: %s
---

%s`, item.Name, item.Description, item.Content)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ListTypes returns all type directories, merging local and global (local takes priority)
func (r *Registry) ListTypes() ([]string, error) {
	seen := make(map[string]bool)
	var types []string

	// Local types first
	if r.LocalPath != "" {
		entries, err := os.ReadDir(r.LocalPath)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !seen[entry.Name()] {
					seen[entry.Name()] = true
					types = append(types, entry.Name())
				}
			}
		}
	}

	// Global types
	entries, err := os.ReadDir(r.BasePath)
	if err != nil {
		return types, nil // return local-only if global unreadable
	}
	for _, entry := range entries {
		if entry.IsDir() && !seen[entry.Name()] {
			seen[entry.Name()] = true
			types = append(types, entry.Name())
		}
	}

	return types, nil
}

// ListItems returns all items of a given type, merging local and global (local wins on name collision)
func (r *Registry) ListItems(itemType string) ([]Item, error) {
	seen := make(map[string]bool)
	var items []Item

	// Local items first
	if r.LocalPath != "" {
		localDir := filepath.Join(r.LocalPath, itemType)
		if _, err := os.Stat(localDir); err == nil {
			localItems, err := r.loadItems(localDir, itemType)
			if err == nil {
				for _, item := range localItems {
					seen[item.Name] = true
					items = append(items, item)
				}
			}
		}
	}

	// Global items (skip those overridden locally)
	globalDir := filepath.Join(r.BasePath, itemType)
	if _, err := os.Stat(globalDir); os.IsNotExist(err) {
		return items, nil
	}
	globalItems, err := r.loadItems(globalDir, itemType)
	if err != nil {
		return items, nil
	}
	for _, item := range globalItems {
		if !seen[item.Name] {
			items = append(items, item)
		}
	}

	return items, nil
}

// loadItems loads all items from a directory
func (r *Registry) loadItems(dir, itemType string) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Raw types (file:) include all files; other types only .md files
		if !IsRawType(itemType) && !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		item, err := loadItem(path, itemType)
		if err != nil {
			continue // Skip invalid files
		}
		items = append(items, *item)
	}

	return items, nil
}

// Profile functions (special case - templates for directives.md)

// GetProfile retrieves a profile by name
func (r *Registry) GetProfile(name string) (*Profile, error) {
	path := filepath.Join(r.BasePath, "profile", name+".md")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile '%s' not found", name)
	}

	return loadProfile(path)
}

// SaveProfile saves a profile to the registry
func (r *Registry) SaveProfile(profile Profile) error {
	profileDir := filepath.Join(r.BasePath, "profile")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(profileDir, profile.Name+".md")
	return saveProfile(path, profile)
}

// ListProfiles returns all profiles
func (r *Registry) ListProfiles() ([]Profile, error) {
	profileDir := filepath.Join(r.BasePath, "profile")

	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		return nil, nil
	}

	return r.loadProfiles(profileDir)
}
