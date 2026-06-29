package backup

import (
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name       string
		fullPath   string
		ignores    []string
		wantIgnore bool
		wantPattern string
		wantErr    bool
	}{
		{
			name:       "empty patterns",
			fullPath:   "/a/b/c.log",
			ignores:    []string{},
			wantIgnore: false,
		},
		{
			name:       "comment pattern",
			fullPath:   "/a/b/c.log",
			ignores:    []string{"# ignore logs", "*.log"},
			wantIgnore: true,
			wantPattern: "*.log",
		},
		{
			name:       "simple glob match",
			fullPath:   "/a/b/c.log",
			ignores:    []string{"*.log"},
			wantIgnore: true,
			wantPattern: "*.log",
		},
		{
			name:       "negated pattern",
			fullPath:   "/a/b/important.log",
			ignores:    []string{"*.log", "!important.log"},
			wantIgnore: false,
			wantPattern: "",
		},
		{
			name:       "absolute path match",
			fullPath:   "/Users/test/.config/app/cache",
			ignores:    []string{"/Users/test/.config/app/cache"},
			wantIgnore: true,
			wantPattern: "/Users/test/.config/app/cache",
		},
		{
			name:       "relative match anywhere",
			fullPath:   "/Users/test/.config/app/logs/error.log",
			ignores:    []string{"logs/*.log"},
			wantIgnore: true,
			wantPattern: "logs/*.log",
		},
		{
			name:       "no match",
			fullPath:   "/Users/test/safe_file.txt",
			ignores:    []string{"*.log", "cache/"},
			wantIgnore: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIgnore, gotPattern, err := shouldIgnore(tt.fullPath, tt.ignores)
			if (err != nil) != tt.wantErr {
				t.Fatalf("shouldIgnore() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotIgnore != tt.wantIgnore {
				t.Errorf("shouldIgnore() gotIgnore = %v, want %v", gotIgnore, tt.wantIgnore)
			}
			if gotPattern != tt.wantPattern {
				t.Errorf("shouldIgnore() gotPattern = %q, want %q", gotPattern, tt.wantPattern)
			}
		})
	}
}
