# Examples and Usage

## Basic Usage
To upload a file to OneDrive:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder
```

## Advanced Usage
To upload a file with custom chunk size and retry settings:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760 --retries 5 --retry-delay 10s
```

## Progress Visualization
Uploading a file with progress visualization:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --progress modern
```

## Remote Management
Listing available remotes:
```bash
ksau-go list-remotes
```

## Quota Information
Displaying OneDrive quota information:
```bash
ksau-go quota
```

## Real-World Use Cases
### Backup Local Files
To backup a local directory to OneDrive:
```bash
ksau-go upload --file /path/to/local/directory --remote /path/to/remote/backup --recursive
```

### Sync Files
To sync files between local and remote directories:
```bash
ksau-go sync --source /path/to/local/directory --destination /path/to/remote/directory
```

### Download Files
To download files from OneDrive to a local directory:
```bash
ksau-go download --remote /path/to/remote/file --local /path/to/local/directory
```

### Share Files
To share a file with a link:
```bash
ksau-go share --file /path/to/remote/file
```

### Automate Uploads
To automate uploads using a cron job:
```bash
(crontab -l ; echo "0 2 * * * ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder") | crontab -
```

## Command Syntax Patterns
### Upload Command
```bash
ksau-go upload --file <local-file-path> --remote <remote-folder-path> [options]
```
### Sync Command
```bash
ksau-go sync --source <local-folder-path> --destination <remote-folder-path> [options]
```
### Download Command
```bash
ksau-go download --remote <remote-file-path> --local <local-folder-path> [options]
```
### Share Command
```bash
ksau-go share --file <remote-file-path> [options]
