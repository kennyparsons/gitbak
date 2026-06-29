package utils

import (
	"path/filepath"
	"os"
	"strings"
)

// ExpandPath expands ~/ and handles relative paths relative to CWD
func ExpandPath(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err == nil {
			return abs
		}
	}
	return path
}

