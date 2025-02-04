# Advanced Features

## Custom Configurations (With your own remote)
You can customize the tool's behavior by modifying the configuration file located at:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`
> NOTE: You must encrypt it with pgp private key and passphrase and upadte it in crypto/

## Performance Optimization
### Adjusting Chunk Size
To optimize upload performance, you can adjust the chunk size:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760
```

### Increasing Timeout
For large file uploads, increase the timeout setting:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --timeout 60s
```

## Security Best Practices
### Encrypting Configuration Files
Ensure your configuration files are encrypted and stored securely. Use tools like GPG to encrypt sensitive files.

### Using Environment Variables
Store sensitive information in environment variables instead of configuration files. For example:
```bash
export KSAU_GO_API_KEY=your_api_key
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
