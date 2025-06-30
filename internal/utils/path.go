package utils

import (
	"path/filepath"
	"os"
	"strings"
)

// expandPath expands ~/ and handles relative paths
func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	if !filepath.IsAbs(path) {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path)
	}
	return path
}
