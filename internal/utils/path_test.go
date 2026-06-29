package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "expand exact home tilde",
			path: "~",
			want: home,
		},
		{
			name: "expand home tilde with sub path",
			path: "~/foo/bar",
			want: filepath.Join(home, "foo/bar"),
		},
		{
			name: "keep absolute path",
			path: "/etc/passwd",
			want: "/etc/passwd",
		},
		{
			name: "relative path expands relative to cwd",
			path: "relative/path",
			want: filepath.Join(wd, "relative/path"),
		},
		{
			name: "relative path with dot expands relative to cwd",
			path: "./relative/path",
			want: filepath.Join(wd, "relative/path"), // filepath.Clean is applied by filepath.Abs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.path)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
