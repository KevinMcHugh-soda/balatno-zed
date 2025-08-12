package game

import (
	"fmt"
	"os"
	"path/filepath"
)

// openGameFile searches for a file by walking up the directory tree
// starting from the current working directory. It returns the first
// matching file found or an error if the file cannot be located.
func openGameFile(relPath string) (*os.File, error) {
	// Try relative to the current working directory first
	if f, err := os.Open(relPath); err == nil {
		return f, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		tryPath := filepath.Join(dir, relPath)
		if f, err := os.Open(tryPath); err == nil {
			return f, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("could not locate %s", relPath)
}
