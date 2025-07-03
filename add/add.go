

package add

import (
	"fmt"
	"path/filepath"

	"github.com/kennyparsons/gitbak/config"
)

// Add adds a new path to a specified app in the configuration.
// If the app doesn't exist, it will be created.
func Add(cfg *config.Config, appName string, pathToAdd string) error {
	// Ensure the path is absolute
	absPath, err := filepath.Abs(pathToAdd)
	if err != nil {
		return fmt.Errorf("could not get absolute path for %s: %v", pathToAdd, err)
	}

	// Check if the app already exists
	appCfg, ok := cfg.CustomApps[appName]
	if !ok {
		// If the app doesn't exist, create it
		appCfg = config.AppConfig{
			Paths: []string{},
		}
	}

	// Check if the path already exists in the app's paths
	for _, p := range appCfg.Paths {
		if p == absPath {
			fmt.Printf("Path %s already exists in app %s. Nothing to do.\n", absPath, appName)
			return nil // Path already exists, do nothing
		}
	}

	// Add the new path
	appCfg.Paths = append(appCfg.Paths, absPath)
	cfg.CustomApps[appName] = appCfg

	fmt.Printf("Added path %s to app %s.\n", absPath, appName)

	return nil
}

