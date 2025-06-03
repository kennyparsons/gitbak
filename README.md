# GitBak
Get your dotfiles/config files backed up and managed with git.

## GO ON, GIT!
![GIT!](https://gifs.kennyparsons.com/git.gif)


GitBak is a simple tool to help you back up your dotfiles and configuration files using Git. It reads a configuration file (`gitbak.json`) to determine which files and directories to back up. 

> Note: It can also work with mackup-supported applications. This is a work in progress and might break. The easiest solution is to just use `mackup show <appname>` to list the files/directories you need to back up and configure them in the `gitbak.json`.

## Features
- Copies files and directories listed in `gitbak.json` to a backup directory
- Supports custom app paths and folders
- Commits and pushes changes to Git (must be a preconfigured git repo)
- Dry-run mode to preview actions

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

## Configuration
Edit `gitbak.json` to set:
- `backup_dir`: Destination directory (should be a Git repo)
- `whitelist_backup_apps`: List of Mackup-supported apps
- `custom_apps`: Map of app names to file/directory paths

## Example
```json
{
  "backup_dir": "/Users/kennyparsons/.dotfiles",
  "whitelist_backup_apps": [],
  "custom_apps": {
    "gitbak": [
      "/Users/kennyparsons/.config/gitbak.gitbak.json"
    ]
  }
}
```

## Requirements
- Go 1.23+
- Git
- (Optional) Mackup for listing supported apps

## Notes
- The backup directory must be an existing Git repository.
- Absolute paths are required in `gitbak.json`. Do not use `~` or relative paths.

## Inspiration
I built GitBak because I wanted a more flexible way to manage my dotfiles. I was inspired by the simplicity of tools like `mackup`, but I wanted more control over what gets backed up and how. For example, I wanted to be able to simply copy files instead of symlinking them, which can be problematic in some cases. GitBak also requires a Git repository for version control, which is a key feature.

Thanks to some awesome coworkers for the inspiration and starting point to build on. 