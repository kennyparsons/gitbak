package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config represents the structure of gitbak.json
type Config struct {
	BackupDir  string              `json:"backup_dir"`
	CustomApps map[string][]string `json:"custom_apps"`
}

// LoadConfig reads and parses gitbak.json into a Config struct
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures the config has necessary fields and paths exist
func (c *Config) Validate() error {
	// Ensure BackupDir exists
	info, err := os.Stat(c.BackupDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrInvalid
	}
	return nil
}
