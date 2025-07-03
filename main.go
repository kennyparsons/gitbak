package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kennyparsons/gitbak/backup"
	"github.com/kennyparsons/gitbak/config"
	"github.com/kennyparsons/gitbak/git"
	"github.com/kennyparsons/gitbak/restore"
)

var version = "dev"

func printHelp() {
	fmt.Println(`
Usage: gitbak <command> [flags]

Commands:
  backup    Copy all configured files into the backup_dir and commit to Git.
  restore   Restore files from backup to their original locations.

Flags:
  --config string   Path to config file (default "./gitbak.json")
  --dry-run         Print actions without actually performing them.
  --no-commit       Skip git add/commit/push after backup
  --app string      When restoring, only restore this specific app
  --version         Print the version and exit

When restoring, if a file already exists, you'll be prompted to:
  (s)kip: Skip this file
  (o)verwrite: Replace the existing file
  (b)ackup: Create a backup of the existing file before restoring

Examples:
  gitbak restore --app ssh    # Only restore SSH configuration
  gitbak restore             # Restore all configured apps`)
}

func main() {
	versionFlag := flag.Bool("version", false, "Print the version and exit")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("gitbak version %s\n", version)
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		printHelp()
		os.Exit(1)
	}

	cmd := flag.Args()[0]
	dryRun := flag.Bool("dry-run", false, "Print steps without executing")
	noCommit := flag.Bool("no-commit", false, "Skip git add/commit/push after backup")
	configPath := flag.String("config", "./gitbak.json", "Path to config file (default: ./gitbak.json)")
	appName := flag.String("app", "", "Only restore this specific app")
	flag.CommandLine.Parse(os.Args[2:])

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "backup":
		if err := backup.PerformBackup(cfg, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			os.Exit(1)
		}
		if !*noCommit {
			if err := git.CommitAndPush(cfg.BackupDir, *dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Git step failed: %v\n", err)
				os.Exit(1)
			}
		} else if !*dryRun {
			fmt.Println("Skipping git commit/push as requested (--no-commit)")
		}
	case "restore":
		if err := restore.Restore(cfg, *dryRun, *appName); err != nil {
			fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
			os.Exit(1)
		}
	default:
		printHelp()
		os.Exit(1)
	}
}
