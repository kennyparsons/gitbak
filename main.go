package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kennyparsons/gitbak/config"
	"github.com/kennyparsons/gitbak/backup"
	"github.com/kennyparsons/gitbak/git"
)

func printHelp() {
	fmt.Println(`
Usage: gitbak <command> [flags]

Commands:
  backup    Copy all configured files into the backup_dir and commit to Git.
  restore   Reverse backup for custom apps (not implemented yet).

Flags:
  --config string   Path to config file (default "./gitbak.json")
  --dry-run         Print actions without actually performing them.
`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]
	dryRun := flag.Bool("dry-run", false, "Print steps without executing")
	configPath := flag.String("config", "./gitbak.json", "Path to config file (default: ./gitbak.json)")
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
		if err := git.CommitAndPush(cfg.BackupDir, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Git step failed: %v\n", err)
			os.Exit(1)
		}
	case "restore":
		fmt.Println("Restore not implemented yet.")
	default:
		printHelp()
		os.Exit(1)
	}
}
