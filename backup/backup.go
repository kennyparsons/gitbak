package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennyparsons/gitbak/config"
)

// expandPath expands ~/ and handles relative paths
func expandPath(path string) string {
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

// copyDir recursively copies a directory tree: srcDir → dstDir
// The destination directory will be created if it doesn't exist
// The source directory's basename will be preserved in the destination
func copyDir(srcDir, dstDir string, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] CopyDir %s → %s\n", srcDir, dstDir)
		return nil
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// Get source directory info to preserve permissions
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %v", err)
	}

	// Ensure destination directory has the same permissions as source
	if err := os.Chmod(dstDir, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set directory permissions: %v", err)
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from srcDir to the current path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		targetPath := filepath.Join(dstDir, relPath)

		// Skip the root directory
		if path == srcDir {
			return nil
		}

		if info.IsDir() {
			// Create the directory with the same permissions as source
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", targetPath, err)
			}
			// Set directory permissions
			return os.Chmod(targetPath, info.Mode())
		}

		// For files, copy directly to the target path
		return copyFile(path, targetPath, dryRun)
	})
}

// copyFile copies a single file to the specified destination path
// dstPath can be either:
// - A directory: file will be placed inside it with its original name
// - A file path: will be used as the exact destination path
func copyFile(srcFile, dstPath string, dryRun bool) error {
	// Get source file info to preserve permissions
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	// Check if dstPath is a directory (either exists as dir or ends with separator)
	dstIsDir := false
	if dstInfo, err := os.Stat(dstPath); err == nil {
		dstIsDir = dstInfo.IsDir()
	} else if strings.HasSuffix(dstPath, string(filepath.Separator)) {
		dstIsDir = true
	}

	// If destination is a directory, append the source filename
	if dstIsDir {
		dstPath = filepath.Join(dstPath, filepath.Base(srcFile))
	} else {
		// Ensure the parent directory exists
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %v", err)
		}
	}

	if dryRun {
		fmt.Printf("[dry-run] CopyFile %s → %s\n", srcFile, dstPath)
		return nil
	}

	// Open source file
	src, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer src.Close()

	// Create destination file with the same permissions as source
	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy the file contents
	if _, err = io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// Preserve file modification time
	if err := os.Chtimes(dstPath, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to preserve modification time: %v", err)
	}

	return nil
}

// PerformBackup copies all files for custom apps
func PerformBackup(cfg *config.Config, dryRun bool) error {
	var allMetadata []FileMetadata

	// Process custom apps
	for appName, rawPaths := range cfg.CustomApps {
		fmt.Printf("● Processing custom app: %s\n", appName)
		dstRoot := filepath.Join(cfg.BackupDir, appName)

		for _, rawPath := range rawPaths {
			srcPath := expandPath(rawPath)
			srcBase := filepath.Base(srcPath)
			dstPath := filepath.Join(dstRoot, srcBase)

			info, err := os.Stat(srcPath)
			if err != nil {
				fmt.Printf("  [skipped] %s (does not exist)\n", srcPath)
				continue
			}

			// Collect metadata
			meta, err := collectFileMetadata(srcPath, filepath.Dir(srcPath))
			if err != nil {
				fmt.Printf("  [warning] Failed to collect metadata for %s: %v\n", srcPath, err)
			} else {
				// Update path to be relative to backup root
				meta.Path = filepath.Join(appName, filepath.Base(srcPath))
				allMetadata = append(allMetadata, meta)
			}

			if info.IsDir() {
				if err := copyDir(srcPath, dstPath, dryRun); err != nil {
					fmt.Printf("  [error] copying directory %s: %v\n", srcPath, err)
				}
			} else {
				if err := os.MkdirAll(dstRoot, 0755); err != nil {
					fmt.Printf("  [error] creating directory %s: %v\n", dstRoot, err)
					continue
				}
				if err := copyFile(srcPath, dstPath, dryRun); err != nil {
					fmt.Printf("  [error] copying file %s: %v\n", srcPath, err)
				}
			}
		}
	}

	// Save metadata
	if !dryRun && len(allMetadata) > 0 {
		if err := saveMetadata(cfg.BackupDir, allMetadata); err != nil {
			return fmt.Errorf("failed to save metadata: %v", err)
		}
		fmt.Println("✓ Saved file metadata")
	}

	return nil
}
