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
