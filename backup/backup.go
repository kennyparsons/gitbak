package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/kennyparsons/gitbak/config"
	"github.com/kennyparsons/gitbak/internal/utils"
)

// shouldIgnore checks if a path should be ignored based on a list of patterns.
// Patterns are similar to .gitignore:
// - Empty lines are ignored.
// - Lines starting with # are comments.
// - Trailing spaces are ignored unless escaped with a backslash.
// - An optional prefix "!" which negates the pattern.
// - If a pattern ends with a slash, it is treated as a directory pattern.
// - Otherwise, the pattern matches files and directories.
// - Patterns containing a slash match paths relative to the repository root.
// - Patterns not containing a slash match paths relative to the directory containing the pattern file.
func shouldIgnore(fullPath string, ignores []string) (bool, string, error) {
	// Normalize path to use forward slashes for consistent matching
	fullPath = filepath.ToSlash(fullPath)

	// Keep track of the last matching pattern (negated or not)
	matched := false

	for _, pattern := range ignores {
		pattern = strings.TrimSpace(pattern)

		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue // Skip empty lines and comments
		}

		negate := false
		if strings.HasPrefix(pattern, "!") {
			negate = true
			pattern = pattern[1:]
		}

		// Handle trailing spaces (escaped or not)
		if strings.HasSuffix(pattern, "\\ ") {
			pattern = strings.TrimSuffix(pattern, "\\ ") + " "
		} else {
			pattern = strings.TrimRight(pattern, " ")
		}

		// If pattern ends with a slash, it only matches directories
		isDirPattern := strings.HasSuffix(pattern, "/")
		if isDirPattern {
			pattern = strings.TrimSuffix(pattern, "/")
		}

		var match bool
		var err error

		if strings.HasPrefix(pattern, "/") {
			// Pattern starts with '/', match from the filesystem root
			match, err = doublestar.Match(pattern, fullPath)
		} else {
			// Pattern does not start with '/', match anywhere in the fullPath
			match, err = doublestar.Match("**/" + pattern, fullPath)
		}

		if err != nil {
			return false, "", fmt.Errorf("error matching pattern %s against path %s: %v", pattern, fullPath, err)
		}

		if match {
			if negate {
				matched = false // Negate previous match
			} else {
				matched = true // This pattern matches
				return matched, pattern, nil
			}
		}
	}
	return matched, "", nil
}

// copyDir recursively copies a directory tree: srcDir → dstDir
// The destination directory will be created if it doesn't exist
// The source directory's basename will be preserved in the destination
func copyDir(srcDir, dstDir string, dryRun bool, globalIgnores []string, appName string) error {
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

		// Skip the root directory itself from ignore checks
		if path == srcDir {
			return nil
		}

		// Check if the current path should be ignored
		ignore, matchedPattern, err := shouldIgnore(path, globalIgnores)
		if err != nil {
			return fmt.Errorf("error checking ignore for %s: %v", path, err)
		}
		if ignore {
			if info.IsDir() {
				fmt.Printf("  %s: Ignored directory %s (globally ignored with \"%s\")\n", appName, relPath, matchedPattern)
				return filepath.SkipDir // Skip this directory and its contents
			}
			fmt.Printf("  %s: Ignored file %s (globally ignored with \"%s\")\n", appName, relPath, matchedPattern)
			return nil // Skip this file
		}

		targetPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// Create the directory with the same permissions as source
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", targetPath, err)
			}
			// Set directory permissions
			return os.Chmod(targetPath, info.Mode())
		}

		// For files, copy directly to the target path
		return copyFile(path, targetPath, dryRun, globalIgnores, appName)
	})
}

// copyFile copies a single file to the specified destination path
// dstPath can be either:
// - A directory: file will be placed inside it with its original name
// - A file path: will be used as the exact destination path
func copyFile(srcFile, dstPath string, dryRun bool, globalIgnores []string, appName string) error {
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
	var wg sync.WaitGroup
	// Use a channel to collect errors from goroutines
	errChan := make(chan error, len(cfg.CustomApps))
	// Channel to send collected metadata from goroutines
	metadataChan := make(chan FileMetadata, len(cfg.CustomApps)*100) // Buffer size is an estimate

	// Process custom apps in parallel
	for appName, appCfg := range cfg.CustomApps {
		wg.Add(1)
		go func(appName string, appCfg config.AppConfig) {
			defer wg.Done()

			fmt.Printf("Processing custom app: %s\n", appName)

			// Execute pre-backup script if defined
			if appCfg.PreBackupScript != "" {
				scriptPath := utils.ExpandPath(appCfg.PreBackupScript)
				fmt.Printf("  %s: Running pre-backup script: %s\n", appName, scriptPath)
				if !dryRun {
					cmd := exec.Command("bash", "-c", scriptPath)
					var stdoutBuf, stderrBuf bytes.Buffer
					cmd.Stdout = &stdoutBuf
					cmd.Stderr = &stderrBuf
					cmdErr := cmd.Run() // Store the error

					// Print stdout
					if stdout := stdoutBuf.String(); stdout != "" {
						for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
							fmt.Printf("  %s: Pre-backup script stdout: %s\n", appName, line)
						}
					}
					// Print stderr
					if stderr := stderrBuf.String(); stderr != "" {
						for _, line := range strings.Split(strings.TrimSpace(stderr), "\n") {
							fmt.Printf("  %s: Pre-backup script stderr: %s\n", appName, line)
						}
					}

					if cmdErr != nil { // Now check the error and return if necessary
						errChan <- fmt.Errorf("%s: pre-backup script failed: %v", appName, cmdErr)
						return
					}
				}
			}

			dstRoot := filepath.Join(cfg.BackupDir, appName)

			for _, rawPath := range appCfg.Paths {

				srcPath := utils.ExpandPath(rawPath)
				srcBase := filepath.Base(srcPath)
				dstPath := filepath.Join(dstRoot, srcBase)

				info, err := os.Stat(srcPath)
				if err != nil {
					fmt.Printf("  %s: Skipped %s (does not exist)\n", appName, srcPath)
					continue
				}

				// Check if the root of the custom app path should be ignored
				ignore, matchedPattern, err := shouldIgnore(srcPath, cfg.GlobalIgnores)
				if err != nil {
					errChan <- fmt.Errorf("%s: checking ignore for %s: %v", appName, srcPath, err)
					continue
				}
				if ignore {
					fmt.Printf("  %s: Ignored %s (matched global ignore pattern \"%s\")\n", appName, srcPath, matchedPattern)
					continue // Skip this entire app path
				}

				// Collect metadata within the goroutine
				meta, err := collectFileMetadata(srcPath, filepath.Dir(srcPath))
				if err != nil {
					fmt.Printf("  %s: Failed to collect metadata for %s: %v\n", appName, srcPath, err)
				} else {
					meta.Path = filepath.Join(appName, filepath.Base(srcPath))
					metadataChan <- meta // Send metadata to channel
				}

				if info.IsDir() {
					if err := copyDir(srcPath, dstPath, dryRun, cfg.GlobalIgnores, appName); err != nil {
						errChan <- fmt.Errorf("%s: copying directory %s: %v", appName, srcPath, err)
						continue
					}
				} else {
					// Check if the file itself should be ignored
					// For single files, the pattern should match the full path relative to the source root
					// Here, we consider the file's path relative to its parent directory for ignore matching
					ignore, matchedPattern, err := shouldIgnore(srcPath, cfg.GlobalIgnores)
					if err != nil {
						errChan <- fmt.Errorf("%s: checking ignore for %s: %v", appName, srcPath, err)
						continue
					}
					if ignore {
						fmt.Printf("  %s: Ignored file %s (globally ignored with \"%s\")\n", appName, srcPath, matchedPattern)
						continue
					}

					if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
						errChan <- fmt.Errorf("%s: creating directory %s: %v", appName, filepath.Dir(dstPath), err)
						continue
					}
					if err := copyFile(srcPath, dstPath, dryRun, cfg.GlobalIgnores, appName); err != nil {
						errChan <- fmt.Errorf("%s: copying file %s: %v", appName, srcPath, err)
						continue
					}
				}
			}
			fmt.Printf("Finished processing custom app: %s\n", appName)
		}(appName, appCfg)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the metadata channel once all workers are done
	close(metadataChan)

	// Collect metadata from the channel in the main goroutine
	var allMetadata []FileMetadata
	for meta := range metadataChan {
		allMetadata = append(allMetadata, meta)
	}

	// Close the error channel and collect errors
	close(errChan)
	var allErrors []error
	for err := range errChan {
		allErrors = append(allErrors, err)
	}

	// If any errors occurred, return the first one (or a combined error)
	if len(allErrors) > 0 {
		return fmt.Errorf("errors during backup: %v", allErrors)
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
