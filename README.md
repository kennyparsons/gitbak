# GitBak
Get your dotfiles/config files backed up and managed with git.

## GO ON, GIT!
![GIT!](https://gifs.kennyparsons.com/git.gif)

GitBak is a simple tool to help you back up your dotfiles and configuration files using Git. It reads a configuration file (`gitbak.json`) to determine which files and directories to back up.

## Features

- Backup files and directories to a Git repository
- Restore files to their original locations
- Preserve file permissions and ownership (some cases might require root/sudo)
- Support for extended attributes (xattrs)
- Dry-run mode to preview changes

## Usage
For testing:
```
go run main.go <command> [--dry-run]
```

For building:
```
go build -o gitbak main.go
chmod +x gitbak
```

### Commands

| Command | Description |
|---------|-------------|
| backup  | Copy configured files to the backup directory and commit to Git |
| restore | Restore files from the backup directory |

### Flags

| Flag | Description |
|------|-------------|
| `--dry-run`  | Print actions without making changes |
| `--config`   | Path to config file (default "./gitbak.json") |
| `--app`      | Restore only a specific app |
| `--no-commit` | Skip git add/commit/push after backup |

## Configuration
Edit `gitbak.json` to set:
- `backup_dir`: Destination directory (should be a Git repo)
- `custom_apps`: Map of app names to their configuration, including paths and optional pre-backup scripts.

### Custom App Configuration
Each entry in `custom_apps` is an object with the following fields:
- `paths`: An array of absolute file or directory paths to back up.
- `pre_backup_script` (optional): An absolute path to a script to execute before backing up the app's paths. This script will be run using `bash -c`.

## Example
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
  }
}
```

## Requirements
- Go 1.23+
- Git

## Notes
- The backup directory must be an existing Git repository.
- Absolute paths are required in `gitbak.json`. Do not use `~` or relative paths.

## Inspiration
I built GitBak because I wanted a more flexible way to manage my dotfiles. I wanted more control over what gets backed up and how. For example, I wanted to be able to simply copy files instead of symlinking them, which can be problematic in some cases. GitBak also requires a Git repository for version control, which is a key feature.

Thanks to some awesome coworkers for the inspiration and starting point to build on. 