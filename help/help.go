package help

import (
	"fmt"
)

// PrintGeneralHelp prints the general usage information for gitbak.
func PrintGeneralHelp() {
	fmt.Println(`
Usage: gitbak <command> [flags]

Commands:
  add             Add a file or folder to an app in the config.
  backup          Copy all configured files into the backup_dir and commit to Git.
  restore         Restore files from backup to their original locations.

Use "gitbak <command> --help" for more information about a command.

Global Flags:
  --version       Print the version and exit
  --help          Print help for all commmands or a specific command (e.g. "gitbak add --help")

Examples:
  gitbak add --path /path/to/file --app myapp
  gitbak restore --app ssh    # Only restore SSH configuration
  gitbak restore             # Restore all configured apps`)
}
