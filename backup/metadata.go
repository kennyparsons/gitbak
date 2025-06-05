// metadata.go
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// FileMetadata contains all the metadata we want to preserve for a file
type FileMetadata struct {
	Path     string      `json:"path"`     // Relative path from backup root
	Mode     os.FileMode `json:"mode"`     // File mode including permissions
	Uid      int         `json:"uid"`      // User ID
	Gid      int         `json:"gid"`      // Group ID
	Xattrs   []Xattr     `json:"xattrs"`   // Extended attributes
	Modified string      `json:"modified"` // Modification time
}

// Xattr represents an extended attribute
type Xattr struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MetadataFileName is the name of the metadata file in the backup
const MetadataFileName = ".gitbak_metadata.json"

// collectFileMetadata collects metadata for a file or directory
func collectFileMetadata(path, basePath string) (FileMetadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return FileMetadata{}, err
	}

	// Get Unix-specific file info
	sys, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return FileMetadata{}, fmt.Errorf("could not get file stats")
	}

	// Get relative path
	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return FileMetadata{}, fmt.Errorf("failed to get relative path: %v", err)
	}

	// Collect extended attributes
	xattrs, err := getXattrs(path)
	if err != nil {
		return FileMetadata{}, fmt.Errorf("failed to get xattrs: %v", err)
	}

	return FileMetadata{
		Path:     relPath,
		Mode:     info.Mode(),
		Uid:      int(sys.Uid),
		Gid:      int(sys.Gid),
		Xattrs:   xattrs,
		Modified: info.ModTime().Format(time.RFC3339Nano),
	}, nil
}

// saveMetadata saves metadata to the backup directory
func saveMetadata(backupRoot string, metadata []FileMetadata) error {
	metadataPath := filepath.Join(backupRoot, MetadataFileName)
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}
	return os.WriteFile(metadataPath, data, 0644)
}

// loadMetadata loads metadata from the backup directory
func loadMetadata(backupRoot string) ([]FileMetadata, error) {
	metadataPath := filepath.Join(backupRoot, MetadataFileName)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metadata []FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return metadata, nil
}

// getXattrs gets extended attributes for a file (platform-specific)
func getXattrs(path string) ([]Xattr, error) {
	// This is a basic implementation that works on Unix-like systems
	// For a complete implementation, you might want to use a library like github.com/pkg/xattr
	return []Xattr{}, nil
}
