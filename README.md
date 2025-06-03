# GitBak
Get your dotfiles/config files backed up and managed with git.

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
- `backup`   Copy configured files to the backup directory and commit to Git
- `restore`  (Coming soon) Restore files from the backup directory

### Flags
- `--dry-run`  Print actions without making changes

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
    "ssh": [
      "/Users/kennyparsons/.config/iterm2/AppSupport/DynamicProfiles"
    ]
  }
}
```

## Requirements
- Go 1.18+
- Git
- (Optional) Mackup for listing supported apps

## Notes
- The backup directory must be an existing Git repository.
- Only the `backup` command is implemented.
- Absolute paths are required in `gitbak.json`. Do not use `~` or relative paths.

## Inspiration
Wanting to make a more extensible and flexible tool for managing dotfiles, I created GitBak. It is inspired by the stupid simple usage of alias and git commands from coworkers, as well as the simplicity of tools like `mackup`. 

However, a core need for gitbak is more control over what gets backed up and how. Also, as of the time of release, `mackup` *only* symlinks files, which can be problematic for some use cases. `Gitbak` solves this by simply copying the configured directories/files to the specified repo. `Mackup` also does not assume the use of a git repo for version control, which is a key feature of GitBak.