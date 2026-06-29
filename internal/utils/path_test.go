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
			got := ExpandPath(tt.path, nil)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestExpandPath_Overrides(t *testing.T) {
	o1, _ := ParsePathOverride("^/Users/olduser/(.*)=/Users/newuser/$1")
	o2, _ := ParsePathOverride("foo=bar")
	overrides := []PathOverride{o1, o2}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "simple regex override",
			path: "/Users/olduser/documents/file.txt",
			want: "/Users/newuser/documents/file.txt",
		},
		{
			name: "substring override",
			path: "/path/to/foo/file",
			want: "/path/to/bar/file",
		},
		{
			name: "no match",
			path: "/Users/otheruser/file",
			want: "/Users/otheruser/file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPath(tt.path, overrides)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) with overrides = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestParsePathOverride(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple override",
			input:   "foo=bar",
			wantErr: false,
		},
		{
			name:    "valid regex override",
			input:   "^/Users/([^/]+)/=/home/$1/",
			wantErr: false,
		},
		{
			name:    "invalid format (no equals)",
			input:   "foobar",
			wantErr: true,
		},
		{
			name:    "invalid regex pattern",
			input:   "*(invalid=bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePathOverride(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePathOverride(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

