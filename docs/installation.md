# Installation Guide

## Linux/macOS
Run this command to download and install ksau-go:
```bash
curl -sSL https://raw.githubusercontent.com/global-index-source/ksau-go/master/setup.sh | bash
```
The script will:
1. Automatically detect your OS and architecture
2. Download the appropriate binary from the latest release
3. Create the configuration directory (~/.ksau/.conf/ on Unix-like systems)
4. Offer to install either system-wide (requires sudo) or in your user directory

## Windows
Run this command in PowerShell:
```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/global-index-source/ksau-go/master/install.ps1").Content
```
> Note: You might encounter security-related messages because the executable is not signed.
> The script installs the tool as `ksau-go.exe` instead of `ksau-go-windows-amd64.exe`.


## Verification
After installation, verify the installation by running:
```bash
ksau-go version
```
You should see the version information of ksau-go.
