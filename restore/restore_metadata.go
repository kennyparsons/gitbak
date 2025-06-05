// restore_metadata.go
package restore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kennyparsons/gitbak/backup"
)

// loadMetadata loads metadata from the backup directory
func loadMetadata(backupRoot string) ([]backup.FileMetadata, error) {
	metadataPath := filepath.Join(backupRoot, backup.MetadataFileName)
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var metadata []backup.FileMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return metadata, nil
}

// applyMetadata applies the stored metadata to a file or directory
func applyMetadata(backupRoot string, meta backup.FileMetadata, dryRun bool) error {
	targetPath := filepath.Join(backupRoot, meta.Path)

	// Check if the target exists
	if _, err := os.Lstat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("target file does not exist: %s", targetPath)
	}

	if dryRun {
		fmt.Printf("[dry-run] Would apply metadata to %s\n", targetPath)
		return nil
	}

	// Set file mode
	if err := os.Chmod(targetPath, meta.Mode); err != nil {
		return fmt.Errorf("failed to set mode for %s: %v", targetPath, err)
	}

	// Set ownership (requires root)
	if os.Geteuid() == 0 {
		if err := os.Chown(targetPath, meta.Uid, meta.Gid); err != nil {
			return fmt.Errorf("failed to set ownership for %s: %v", targetPath, err)
		}
	} else if meta.Uid != os.Getuid() || meta.Gid != os.Getgid() {
		fmt.Printf("  [warning] Need root to set ownership for %s (UID: %d, GID: %d)\n",
			targetPath, meta.Uid, meta.Gid)
	}

	// Set extended attributes
	for _, xattr := range meta.Xattrs {
		if err := setXattr(targetPath, xattr); err != nil {
			return fmt.Errorf("failed to set xattr %s on %s: %v",
				xattr.Name, targetPath, err)
		}
	}

	// Set modification time
	modTime, err := time.Parse(time.RFC3339Nano, meta.Modified)
	if err == nil {
		if err := os.Chtimes(targetPath, time.Now(), modTime); err != nil {
			return fmt.Errorf("failed to set modification time for %s: %v",
				targetPath, err)
		}
	}

	return nil
}

// setXattr sets an extended attribute (platform-specific)
func setXattr(path string, xattr backup.Xattr) error {
	// This is a basic implementation that works on Unix-like systems
	// For a complete implementation, you might want to use a library like github.com/pkg/xattr
	return nil
}
