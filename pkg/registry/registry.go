package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// New creates a new Registry instance
func New() (*Registry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(homeDir, ".agmd")
	return &Registry{
		BasePath: basePath,
	}, nil
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

// GetItem retrieves an item by type and name
func (r *Registry) GetItem(itemType, name string) (*Item, error) {
	path := filepath.Join(r.BasePath, itemType, name+".md")

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

	path := filepath.Join(typeDir, item.Name+".md")

	content := fmt.Sprintf(`---
name: %s
description: %s
---

%s`, item.Name, item.Description, item.Content)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ListTypes returns all type directories in the registry
func (r *Registry) ListTypes() ([]string, error) {
	entries, err := os.ReadDir(r.BasePath)
	if err != nil {
		return nil, err
	}

	var types []string
	for _, entry := range entries {
		if entry.IsDir() {
			types = append(types, entry.Name())
		}
	}
	return types, nil
}

// ListItems returns all items of a given type
func (r *Registry) ListItems(itemType string) ([]Item, error) {
	typeDir := filepath.Join(r.BasePath, itemType)

	if _, err := os.Stat(typeDir); os.IsNotExist(err) {
		return nil, nil // No items of this type
	}

	return r.loadItems(typeDir, itemType)
}

// loadItems loads all items from a directory
func (r *Registry) loadItems(dir, itemType string) ([]Item, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var items []Item
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
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
