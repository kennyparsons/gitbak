# GitBak

Get your dotfiles/config files backed up and managed with Git.

## GO ON, GIT!
![GIT!](https://gifs.kennyparsons.com/git.gif)

GitBak is a simple tool to help you back up your dotfiles and configuration files using Git.

## Features

- Backup files and directories to a Git repository
- Restore files to their original locations
- Preserve file permissions and ownership (some cases might require root/sudo)
- Parallel processing for faster backups
- Supports custom app paths and folders
- Commits and pushes changes to Git (must be a preconfigured Git repo)
- Dry-run mode to preview changes
- *(Pending)* Support for extended attributes (xattrs)

## Usage

For testing:
```sh
go run main.go <command> [--dry-run]
```

For building:
```sh
go build -o gitbak main.go
chmod +x gitbak
```

### Commands

| Command | Description |
|---------|-------------|
| `backup`  | Copy configured files to the backup directory and commit to Git |
| `restore` | Restore files from the backup directory |

### Flags

| Flag | Description |
|------|-------------|
| `--dry-run`   | Print actions without making changes |
| `--config`    | Path to config file (default `./gitbak.json`) |
| `--app`       | Restore only a specific app |
| `--no-commit` | Skip `git add/commit/push` after backup |

## Configuration

GitBak uses a `gitbak.json` file to define what to back up and how. The configuration must include:

- **`backup_dir`**: The destination directory for backups. This must be an existing Git repository.
- **`custom_apps`**: A map of app names with their backup details.
- **`global_ignores`** *(optional)*: An array of glob patterns for files or directories to exclude (applies globally).

### Example `gitbak.json`

```json
{
  "backup_dir": "/Users/kennyparsons/.dotfiles",
  "custom_apps": {
    "gitbak": {
      "paths": [
        "/Users/kennyparsons/.config/gitbak/gitbak.json"
      ]
    },
    "brew": {
      "paths": [
        "/Users/kennyparsons/.config/brew/Brewfile",
        "/Users/kennyparsons/.config/brew/brew_list.txt"
      ],
      "pre_backup_script": "/Users/kennyparsons/bin/brew_backup.sh backup"
    }
  },
  "global_ignores": [
    "/Users/kennyparsons/.config/some_app/cache",
    "/Users/kennyparsons/.local/share/some_other_app/*.log"
  ]
}
```

### Custom Apps

Each app under `custom_apps` must include:
- **`paths`**: An array of absolute file or directory paths to back up.
- **`pre_backup_script`** *(optional)*: An absolute path to a script to run before backing up that app’s files (runs with `bash -c`). This is useful for apps like `brew` or `pgdump` that can create snapshots or dumps before backing up.

### Global Ignores

Use `global_ignores` to skip caches, logs, or other files you don’t want to version. Patterns use [doublestar](https://github.com/bmatcuk/doublestar?tab=readme-ov-file#patterns) syntax for flexible matching.

> **Note:** Always use absolute paths — `~` and relative paths will not work.

## Requirements

- Go 1.23+
- Git

## Notes

- The backup directory must be an existing Git repository.
- Use absolute paths in `gitbak.json`. Do not use `~` or relative paths.

## Inspiration

I built GitBak because I wanted a flexible way to manage my dotfiles and configs without relying on symlinks (the primary method of the former Mackup utility), which can be problematic (see mackup issues [#2012](https://github.com/lra/mackup/pull/2012), [#2008](https://github.com/lra/mackup/issues/2008), [#2051](https://github.com/lra/mackup/issues/2051), and more). 

GitBak makes it easy to copy, version, and restore config files while giving you full control over what gets backed up, and not interfering with the original files.

Thanks to some awesome coworkers for the inspiration and a solid starting point to build on (that's you, James).