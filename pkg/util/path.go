package util

import (
	"path/filepath"
)

// ResolvePath turns the path into an absolute path
func ResolvePath(path string) string {
	if path == "" {
		return path
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return absPath
}
