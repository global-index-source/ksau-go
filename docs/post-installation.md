# Post-Installation Guide

## Configuration
The tool stores its configuration in:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`

## Refresh Configuration
After installation, run the following command to refresh the rclone configuration:
```bash
ksau-go refresh
```

## Shell Completion
To enable command-line completion for your shell, run:
```bash
ksau-go completion [shell] >> ~/.bashrc && source ~/.bashrc  # For bash
ksau-go completion [shell] >> ~/.zshrc && source ~/.zshrc    # For zsh
ksau-go completion [shell] >> $PROFILE                       # For PowerShell
```
Replace `[shell]` with your shell type (e.g., bash, zsh, powershell).

## Renaming the Tool
To rename the tool from `ksau-go` to `ksau`, you can create an alias in your shell configuration file:
```bash
alias ksau='ksau-go'
```
Add this line to your shell configuration file (`~/.bashrc`, `~/.zshrc`, or `$PROFILE` for PowerShell) and reload the configuration:
```bash
source ~/.bashrc  # For bash
source ~/.zshrc   # For zsh
. $PROFILE        # For PowerShell
```

## Environment Variables
If you need to set any environment variables for the tool, add them to your shell configuration file. For example:
```bash
export KSAU_GO_ENV_VAR=value
```
Replace `KSAU_GO_ENV_VAR` with the actual environment variable name and `value` with the desired value.

## Setting Up Aliases
You can set up additional aliases for frequently used commands. For example:
```bash
alias ksau-upload='ksau-go upload'
alias ksau-refresh='ksau-go refresh'
```
Add these lines to your shell configuration file and reload the configuration.

## Additional Tips
- Ensure your configuration file is properly formatted and contains all necessary settings.
- Regularly update the tool to get the latest features and bug fixes.
- Refer to the [troubleshooting guide](troubleshooting.md) for common issues and solutions.
