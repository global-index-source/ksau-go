# Advanced Features

## Custom Configurations (With your own remote)
In the future, you can customize the tool's behavior by modifying the configuration file located at:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`

However, currently the tool has no way of knowing if the file is not encrypted, therefore the decryption
process will fail and cause the program to panic.

## Performance Optimization
### Adjusting Chunk Size
To optimize upload performance, you can adjust the chunk size:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760
```

### Increasing Timeout
For large file uploads with large chunk size, increase the timeout setting:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --timeout 60s
```

## Integration with Other Tools
### Using with Cron Jobs
Automate uploads using cron jobs:
```bash
(crontab -l ; echo "0 2 * * * ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder") | crontab -
```

## Additional Tips
- Regularly update the tool to get the latest features and bug fixes.
- Refer to the [troubleshooting guide](troubleshooting.md) for common issues and solutions.
- Ensure your configuration file is properly formatted and contains all necessary settings.
