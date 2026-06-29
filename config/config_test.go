package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gitbak.json")

	content := `{
		"backup_dir": "/tmp/backup",
		"custom_apps": {
			"app1": {
				"paths": ["/path/to/file1"]
			}
		}
	}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.BackupDir != "/tmp/backup" {
		t.Errorf("BackupDir = %q, want %q", cfg.BackupDir, "/tmp/backup")
	}

	app1, ok := cfg.CustomApps["app1"]
	if !ok {
		t.Fatalf("app1 not found in config")
	}

	if len(app1.Paths) != 1 || app1.Paths[0] != "/path/to/file1" {
		t.Errorf("app1 paths = %v, want [%q]", app1.Paths, "/path/to/file1")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	content := `{invalid json}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig succeeded on invalid JSON, expected error")
	}
}

func TestConfig_Validate(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Valid config (backup dir exists)
	cfg := &Config{
		BackupDir: tempDir,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate failed for valid config: %v", err)
	}

	// 2. Invalid config (backup dir does not exist)
	cfgNonExistent := &Config{
		BackupDir: filepath.Join(tempDir, "does-not-exist"),
	}
	if err := cfgNonExistent.Validate(); err == nil {
		t.Error("Validate succeeded for non-existent backup dir, expected error")
	}

	// 3. Invalid config (backup dir is a file)
	tempFile := filepath.Join(tempDir, "somefile.txt")
	if err := os.WriteFile(tempFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	cfgFile := &Config{
		BackupDir: tempFile,
	}
	if err := cfgFile.Validate(); err == nil {
		t.Error("Validate succeeded when backup dir is a file, expected error")
	}
}

func TestConfig_SaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "saved.json")

	cfg := &Config{
		BackupDir: "/some/backup/dir",
		CustomApps: map[string]AppConfig{
			"testapp": {
				Paths: []string{"/some/path"},
			},
		},
	}

	if err := cfg.SaveConfig(configPath); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Reload and verify
	reloaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to reload saved config: %v", err)
	}

	if reloaded.BackupDir != cfg.BackupDir {
		t.Errorf("reloaded BackupDir = %q, want %q", reloaded.BackupDir, cfg.BackupDir)
	}

	if _, ok := reloaded.CustomApps["testapp"]; !ok {
		t.Error("reloaded config is missing 'testapp'")
	}
}
