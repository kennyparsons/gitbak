package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PathOverride represents a regex pattern and its replacement
type PathOverride struct {
	Pattern     *regexp.Regexp
	Replacement string
}

// ParsePathOverride parses a string in the format "regex=replacement" into a PathOverride
func ParsePathOverride(s string) (PathOverride, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return PathOverride{}, fmt.Errorf("invalid override format %q, must be pattern=replacement", s)
	}
	re, err := regexp.Compile(parts[0])
	if err != nil {
		return PathOverride{}, fmt.Errorf("invalid regex %q: %v", parts[0], err)
	}
	return PathOverride{Pattern: re, Replacement: parts[1]}, nil
}

// ApplyOverrides applies a list of path overrides to a path
func ApplyOverrides(path string, overrides []PathOverride) string {
	for _, o := range overrides {
		if o.Pattern.MatchString(path) {
			path = o.Pattern.ReplaceAllString(path, o.Replacement)
		}
	}
	return path
}

// ExpandPath expands ~/ and handles relative paths relative to CWD, applying overrides first
func ExpandPath(path string, overrides []PathOverride) string {
	path = ApplyOverrides(path, overrides)
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


