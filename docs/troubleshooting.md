# Troubleshooting Guide

## Common Error Messages and Solutions

### Error: "Failed to authenticate"
- **Solution**: Ensure your credentials are correct and properly configured in the configuration file. Refresh the configuration using:
  ```bash
  ksau-go refresh
  ```

### Error: "Network timeout"
- **Solution**: Check your internet connection. If the problem persists, try increasing the timeout settings:
  ```bash
  ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --timeout 60s
  ```

### Error: "Insufficient storage"
- **Solution**: Verify your OneDrive storage quota using:
  ```bash
  ksau-go quota
  ```
  Free up space or upgrade your storage plan if necessary.

## Configuration Issues

### Issue: "Configuration file not found"
- **Solution**: Ensure the configuration file is located in the correct directory:
  - Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
  - Windows: `%AppData%\ksau\.conf\rclone.conf`

### Issue: "Invalid configuration format"
- **Solution**: Verify the configuration file format and ensure all required fields are correctly filled.

## Network-Related Problems

### Problem: "Unable to connect to OneDrive"
- **Solution**: Check your network connection and firewall settings. Ensure that OneDrive's endpoints are not blocked.

### Problem: "Slow upload speeds"
- **Solution**: Optimize your network settings and try uploading during off-peak hours. You can also adjust the chunk size for uploads:
  ```bash
  ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760
  ```

## Updating/Refreshing Configurations

### How to update configurations
- **Solution**: To update your configurations, edit the configuration file and then refresh the settings using:
  ```bash
  ksau-go refresh
  ```

## Command Reference for Problem-Solving

### Refresh Configuration
```bash
ksau-go refresh
```

### Check Quota
```bash
ksau-go quota
```

### List Remotes
```bash
ksau-go list-remotes
```

### Upload with Increased Timeout
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --timeout 60s
```

### Upload with Custom Chunk Size
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760
```

## Additional Tips
- Regularly update the tool to get the latest features and bug fixes.
- Refer to the [post-installation guide](post-installation.md) for configuration and setup tips.
- Ensure your configuration file is properly formatted and contains all necessary settings.
