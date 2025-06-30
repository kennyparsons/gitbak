package restore

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennyparsons/gitbak/backup"
	"github.com/kennyparsons/gitbak/config"
	"github.com/kennyparsons/gitbak/internal/utils"
)

// Restore restores files from the backup directory to their original locations.
// If appName is not empty, only restores the specified app.
func Restore(cfg *config.Config, dryRun bool, appName string) error {
	// Load metadata
	metadata, err := loadMetadata(cfg.BackupDir)
	if err != nil {
		fmt.Printf("  [warning] Failed to load metadata: %v\n", err)
	}

	// Create a map for faster lookups
	metadataMap := make(map[string]backup.FileMetadata)
	for _, meta := range metadata {
		metadataMap[meta.Path] = meta
	}

	// Process custom apps
	for currentAppName, appCfg := range cfg.CustomApps {
		if appName != "" && currentAppName != appName {
			continue
		}

		fmt.Printf("● Restoring app: %s\n", currentAppName)
		backupAppDir := filepath.Join(cfg.BackupDir, currentAppName)

		for _, srcPath := range appCfg.Paths {
			expandedSrc := utils.ExpandPath(srcPath)
			srcBase := filepath.Base(expandedSrc)
			backupPath := filepath.Join(backupAppDir, srcBase)

			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				backupPath = filepath.Join(backupAppDir, filepath.Base(expandedSrc))
			}

			// Handle the restore
			if err := restorePath(backupPath, expandedSrc, dryRun); err != nil {
				fmt.Printf("  [error] restoring %s: %v\n", srcPath, err)
			}

			// Apply metadata if available
			relPath := filepath.Join(currentAppName, filepath.Base(expandedSrc))
			if meta, exists := metadataMap[relPath]; exists {
				if err := applyMetadata(cfg.BackupDir, meta, dryRun); err != nil {
					fmt.Printf("  [warning] Failed to apply metadata to %s: %v\n",
						expandedSrc, err)
				}
			}
		}
	}

	return nil
}



func restorePath(backupPath, originalPath string, dryRun bool) error {
	// Expand ~ in the original path
	expandedOriginal := utils.ExpandPath(originalPath)

	// Check if the backup path exists
	backupInfo, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		// If not found, try to find the backup in the new structure
		// For directories, the backup might be in backupAppDir/dirname
		backupDir := filepath.Dir(backupPath)
		srcBase := filepath.Base(expandedOriginal)
		backupPath = filepath.Join(backupDir, srcBase)

		backupInfo, err = os.Stat(backupPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("backup not found: %s (tried %s)", filepath.Base(expandedOriginal), backupPath)
		}
	} else if err != nil {
		return fmt.Errorf("error checking backup path: %v", err)
	}

	// In dry-run mode, just show what would happen
	if dryRun {
		if backupInfo.IsDir() {
			fmt.Printf("[dry-run] Would restore directory %s → %s\n", backupPath, expandedOriginal)
			return nil
		}
		fmt.Printf("[dry-run] Would restore file %s → %s\n", backupPath, expandedOriginal)
		return nil
	}

	// Check if destination exists
	if _, err := os.Stat(expandedOriginal); err == nil {
		// Destination exists, prompt for action
		fmt.Printf("  [conflict] %s already exists. (s)kip, (o)verwrite, (b)ackup? ", expandedOriginal)
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		switch response {
		case "o": // Overwrite
			// Continue with restore
		case "b": // Backup
			backupPath := fmt.Sprintf("%s.gitbak-restore-state-%s", expandedOriginal, time.Now().Format("2006-01-02T15:04:05"))
			if err := os.Rename(expandedOriginal, backupPath); err != nil {
				return fmt.Errorf("failed to backup existing file: %v", err)
			}
			fmt.Printf("  [backup] created backup at %s\n", backupPath)
		case "s": // Skip
			fallthrough
		default:
			fmt.Println("  [skipped]")
			return nil
		}
	}

	if backupInfo.IsDir() {
		return restoreDirectory(backupPath, expandedOriginal, dryRun)
	}
	return restoreFile(backupPath, expandedOriginal, dryRun)
}

func restoreFile(src, dst string, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] Restore file %s → %s\n", src, dst)
		return nil
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	// Copy the file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// Preserve file mode
	if srcInfo, err := os.Stat(src); err == nil {
		os.Chmod(dst, srcInfo.Mode())
	}

	fmt.Printf("  [restored] %s\n", dst)
	return nil
}

func restoreDirectory(src, dst string, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] Restore directory %s → %s\n", src, dst)
		return nil
	}

	fmt.Printf("  [restoring directory] %s\n", dst)

	// First, ensure the source directory exists
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("failed to get source directory info: %v", err)
	}

	// Create the destination directory with the same permissions as source
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %v", err)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the source directory
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create the directory with the same permissions as source
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dstPath, err)
			}
			// Set directory permissions
			if err := os.Chmod(dstPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to set permissions for %s: %v", dstPath, err)
			}
			return nil
		}

		// For files, use restoreFile which handles permissions and copying
		return restoreFile(path, dstPath, dryRun)
	})
}
