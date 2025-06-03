package backup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kennyparsons/gitbak/config"
)

// copyDir recursively copies a directory tree: srcDir → dstDir/<basename(srcDir)>
func copyDir(srcDir, dstDirRoot string, dryRun bool) error {
	base := filepath.Base(srcDir)
	dstDir := filepath.Join(dstDirRoot, base)
	if dryRun {
		fmt.Printf("[dry-run] CopyDir %s → %s\n", srcDir, dstDir)
		return nil
	}
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, dryRun)
	})
}

// copyFile copies a single file to the specified destination path
// If dstDirRoot is a directory, the file will be placed inside it with its original name
// If dstDirRoot is a file path, it will be used as the exact destination path
func copyFile(srcFile, dstPath string, dryRun bool) error {
	// Check if dstPath is a directory (ends with a separator or is an existing directory)
	if dstInfo, err := os.Stat(dstPath); err == nil && dstInfo.IsDir() {
		// If it's a directory, append the source filename
		dstPath = filepath.Join(dstPath, filepath.Base(srcFile))
	}

	if dryRun {
		fmt.Printf("[dry-run] CopyFile %s → %s\n", srcFile, dstPath)
		return nil
	}

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// getMackupPaths runs "mackup show <app>" and parses the output paths
func getMackupPaths(app string) ([]string, error) {
	cmd := exec.Command("mackup", "show", app)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				paths = append(paths, parts[1])
			}
		}
	}
	return paths, nil
}

// PerformBackup copies all files for supported and custom apps
func PerformBackup(cfg *config.Config, dryRun bool) error {
	// 1. Process Mackup-supported apps (preserve full path under BackupDir)
	for _, app := range cfg.WhitelistBackupApps {
		paths, err := getMackupPaths(app)
		if err != nil {
			fmt.Printf("Warning: cannot get mackup paths for %s: %v\n", app, err)
			continue
		}
		for _, relPath := range paths {
			src := filepath.Join(os.Getenv("HOME"), relPath)
			dst := filepath.Join(cfg.BackupDir, relPath)
			if dryRun {
				fmt.Printf("[dry-run] Copy %s → %s\n", src, dst)
				continue
			}
			info, err := os.Stat(src)
			if err != nil {
				fmt.Printf("  [skipped] %s (does not exist)\n", src)
				continue
			}
			if info.IsDir() {
				if err := os.MkdirAll(dst, info.Mode()); err != nil {
					fmt.Printf("  [error] mkdir %s: %v\n", dst, err)
					continue
				}
				if err := filepath.Walk(src, func(path string, fi os.FileInfo, e error) error {
					if e != nil {
						return e
					}
					rp, _ := filepath.Rel(src, path)
					target := filepath.Join(dst, rp)
					if fi.IsDir() {
						return os.MkdirAll(target, fi.Mode())
					}
					return copyFile(path, filepath.Dir(target), dryRun)
				}); err != nil {
					fmt.Printf("  [error] copying dir %s: %v\n", src, err)
				}
			} else {
				if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
					fmt.Printf("  [error] mkdir %s: %v\n", filepath.Dir(dst), err)
					continue
				}
				if err := copyFile(src, filepath.Dir(dst), dryRun); err != nil {
					fmt.Printf("  [error] copying file %s: %v\n", src, err)
				}
			}
		}
	}

	// 2. Process custom apps (each under BackupDir/<appName>/)
	for appName, rawPaths := range cfg.CustomApps {
		fmt.Printf("● Processing custom app: %s\n", appName)
		dstRoot := filepath.Join(cfg.BackupDir, appName)
		for _, rawPath := range rawPaths {
			// Expand "~/" if present
			src := rawPath
			if strings.HasPrefix(rawPath, "~/") {
				src = filepath.Join(os.Getenv("HOME"), rawPath[2:])
			}
			// If not absolute and not "~/", assume relative to HOME
			if !filepath.IsAbs(rawPath) && !strings.HasPrefix(rawPath, "~/") {
				src = filepath.Join(os.Getenv("HOME"), rawPath)
			}
			info, err := os.Stat(src)
			if err != nil {
				fmt.Printf("  [skipped] %s (does not exist)\n", src)
				continue
			}
			if info.IsDir() {
				if err := copyDir(src, dstRoot, dryRun); err != nil {
					fmt.Printf("  [error] copyDir %s → %s: %v\n", src, dstRoot, err)
				}
			} else {
				if err := copyFile(src, dstRoot, dryRun); err != nil {
					fmt.Printf("  [error] copyFile %s → %s: %v\n", src, dstRoot, err)
				}
			}
		}
		fmt.Println()
	}
	return nil
}
