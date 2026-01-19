package config

import (
	"os"
)

// ExistingFile represents a detected configuration file
type ExistingFile struct {
	Filename string
	Size     int64
	IsSymlink bool
}

// DetectExistingFiles checks for existing AI agent config files
func DetectExistingFiles() []ExistingFile {
	var found []ExistingFile

	for _, filename := range LegacyFilenames {
		info, err := os.Lstat(filename)
		if err == nil {
			existing := ExistingFile{
				Filename: filename,
				Size:     info.Size(),
				IsSymlink: info.Mode()&os.ModeSymlink != 0,
			}
			found = append(found, existing)
		}
	}

	return found
}

// FindBestImportCandidate returns the most likely file to import
func FindBestImportCandidate(files []ExistingFile) *ExistingFile {
	// Priority: CLAUDE.md > AGENT.md > others
	priority := map[string]int{
		"CLAUDE.md":      10,
		"AGENT.md":       9,
		"agent.md":       8,
		".cursorrules":   7,
		"GEMINI.md":      6,
		".windsurfrules": 5,
		".continuerules": 4,
	}

	var best *ExistingFile
	bestScore := -1

	for i := range files {
		// Skip symlinks
		if files[i].IsSymlink {
			continue
		}

		score := priority[files[i].Filename]
		if score > bestScore {
			bestScore = score
			best = &files[i]
		}
	}

	return best
}
