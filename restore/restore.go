package restore

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennyparsons/gitbak/config"
)

// Restore restores files from the backup directory to their original locations
func Restore(cfg *config.Config, dryRun bool) error {
	// Process custom apps
	for appName, srcPaths := range cfg.CustomApps {
		fmt.Printf("● Restoring app: %s\n", appName)
		backupAppDir := filepath.Join(cfg.BackupDir, appName)

		// For each source path in the app's configuration
		for _, srcPath := range srcPaths {
			// Expand the source path (handle ~/ and relative paths)
			expandedSrc := expandPath(srcPath)
			
			// The backup path is in the app's directory
			// For files: backupAppDir/filename
			// For directories: backupAppDir/dirname/...
			srcBase := filepath.Base(expandedSrc)
			backupPath := filepath.Join(backupAppDir, srcBase)

			// Check if the backup exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				// If not found, try the old format (for backward compatibility)
				backupPath = filepath.Join(backupAppDir, filepath.Base(expandedSrc))
			}

			// Handle the restore
			if err := restorePath(backupPath, srcPath, dryRun); err != nil {
				fmt.Printf("  [error] restoring %s: %v\n", srcPath, err)
			}
		}
	}

	return nil
}

// expandPath expands ~/ and handles relative paths
func expandPath(path string) string {
	// Expand ~/
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	// Handle relative paths (assume relative to home)
	if !filepath.IsAbs(path) {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path)
	}
	return path
}

func restorePath(backupPath, originalPath string, dryRun bool) error {
	// Expand ~ in the original path
	expandedOriginal := originalPath
	if strings.HasPrefix(originalPath, "~/") {
		home, _ := os.UserHomeDir()
		expandedOriginal = filepath.Join(home, originalPath[2:])
	} else if !filepath.IsAbs(originalPath) {
		// Handle relative paths (assume relative to home)
		home, _ := os.UserHomeDir()
		expandedOriginal = filepath.Join(home, originalPath)
	}

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
