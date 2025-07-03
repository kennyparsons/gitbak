package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kennyparsons/gitbak/add"
	"github.com/kennyparsons/gitbak/backup"
	"github.com/kennyparsons/gitbak/config"
	"github.com/kennyparsons/gitbak/git"
	"github.com/kennyparsons/gitbak/help"
	"github.com/kennyparsons/gitbak/internal/utils"
	"github.com/kennyparsons/gitbak/restore"
)

var version = "dev"

func main() {
	// Subcommands
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addApp := addCmd.String("app", "", "App name to add the path to (required)")
	addPath := addCmd.String("path", "", "Path to the file or folder to add (required)")
	addConfig := addCmd.String("config", "~/.config/gitbak/gitbak.json", "Path to config file")

	backupCmd := flag.NewFlagSet("backup", flag.ExitOnError)
	backupDryRun := backupCmd.Bool("dry-run", false, "Print steps without executing")
	backupNoCommit := backupCmd.Bool("no-commit", false, "Skip git add/commit/push after backup")
	backupConfig := backupCmd.String("config", "~/.config/gitbak/gitbak.json", "Path to config file")

	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	restoreDryRun := restoreCmd.Bool("dry-run", false, "Print steps without executing")
	restoreApp := restoreCmd.String("app", "", "Only restore this specific app")
	restoreConfig := restoreCmd.String("config", "~/.config/gitbak/gitbak.json", "Path to config file")

	if len(os.Args) < 2 {
		help.PrintGeneralHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])

		if *addApp == "" {
			fmt.Fprintln(os.Stderr, "Error: --app flag is required")
			addCmd.Usage()
			os.Exit(1)
		}
		if *addPath == "" {
			fmt.Fprintln(os.Stderr, "Error: --path flag is required")
			addCmd.Usage()
			os.Exit(1)
		}

		configPath := utils.ExpandPath(*addConfig)

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", configPath, err)
			os.Exit(1)
		}
		if err := add.Add(cfg, *addApp, *addPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding path: %v\n", err)
			os.Exit(1)
		}
		if err := cfg.SaveConfig(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully updated config at %s\n", configPath)

	case "backup":
		backupCmd.Parse(os.Args[2:])
		configPath := utils.ExpandPath(*backupConfig)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", configPath, err)
			os.Exit(1)
		}
		if err := backup.PerformBackup(cfg, *backupDryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Backup failed: %v\n", err)
			os.Exit(1)
		}
		if !*backupNoCommit {
			if err := git.CommitAndPush(cfg.BackupDir, *backupDryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Git step failed: %v\n", err)
				os.Exit(1)
			}
		}

	case "restore":
		restoreCmd.Parse(os.Args[2:])
		configPath := utils.ExpandPath(*restoreConfig)
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", configPath, err)
			os.Exit(1)
		}
		if err := restore.Restore(cfg, *restoreDryRun, *restoreApp); err != nil {
			fmt.Fprintf(os.Stderr, "Restore failed: %v\n", err)
			os.Exit(1)
		}
	case "--version", "-version":
		fmt.Printf("%s\n", version)
		os.Exit(0)
	case "help":
		help.PrintGeneralHelp()
		os.Exit(0)
	default:
		help.PrintGeneralHelp()
		os.Exit(1)
	}
}
