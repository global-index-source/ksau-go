# ksau-go

[![Go Version](https://img.shields.io/badge/go-1.23.4-blue)](https://golang.org/doc/go1.23)

## Table of Contents
1. [Introduction](#introduction)
2. [Installation](#installation)
   - [Recommended Method](#recommended-method)
   - [Linux/macOS](#linuxmacos)
   - [Windows](#windows)
   - [From Source](#from-source)
3. [Configuration](#configuration)
4. [Post-Installation](#post-installation)
5. [Usage](#usage)
   - [Basic Usage](#basic-usage)
   - [Advanced Usage](#advanced-usage)
   - [Examples](#examples)
6. [Project Structure](#project-structure)
7. [Contribution Guidelines](#contribution-guidelines)
8. [Motivation](#motivation)
9. [License](#license)
10. [Contact Information](#contact-information)

## Introduction
ksau-go is a tool for uploading files to "our" OneDrive for "free" unitl we run out space, promoting open-source culture and helping developers as we are one of them.

## Installation

### Recommended Method
Run this command to download and install ksau-go:
```bash
curl -sSL https://raw.githubusercontent.com/global-index-source/ksau-go/master/setup.sh | bash
```
This method is recommended because it automatically detects your OS and architecture, downloads the appropriate binary, and sets up the configuration directory. Note: This method currently works for Linux/macOS only.

### Linux/macOS
Run this command to download and install ksau-go:
```bash
curl -sSL https://raw.githubusercontent.com/global-index-source/ksau-go/master/setup.sh | bash
```
The script will:
1. Automatically detect your OS and architecture
2. Download the appropriate binary from the latest release
3. Create the configuration directory (~/.ksau/.conf/ on Unix-like systems)
4. Offer to install either system-wide (requires sudo) or in your user directory

### Windows
Run this command in PowerShell:
```powershell
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/global-index-source/ksau-go/master/install.ps1").Content
```
Note: You might encounter security-related messages because the executable is not signed. The script installs the tool as `ksau-go` instead of `ksau-go-windows-amd64.exe`.

### From Source
To build this project from source, you need two important things:
1. Private PGP key used to decrypt rclone.conf
2. The passphrase of the PGP key

They should be placed under **crypto/** like so:
```
└───crypto
        algo.go
        placeholder.go
        >> passphrase.txt  ⌉  -- These files
        >> privkey.pem     ⌋     they are not provided by the repo
```

Finally, install the dependencies and you're ready to build the project!
```
go mod tidy  # install dependencies
make build   # build the project
```

Depending on the OS you're on, you'll see `ksau-go` or `ksau-go.exe` generated in the current working directory.

## Configuration
The tool stores its configuration in:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`

## Post-Installation
After installation, run the following command to refresh the rclone configuration:
```bash
ksau-go refresh
```

To enable command-line completion for your shell, run:
```bash
ksau-go completion [shell] >> ~/.bashrc && source ~/.bashrc  # For bash
ksau-go completion [shell] >> ~/.zshrc && source ~/.zshrc    # For zsh
ksau-go completion [shell] >> $PROFILE                       # For PowerShell
```
Replace `[shell]` with your shell type (e.g., bash, zsh, powershell).

## Usage

### Basic Usage
To upload a file to OneDrive:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder
```

### Advanced Usage
To upload a file with custom chunk size and retry settings:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --chunk-size 10485760 --retries 5 --retry-delay 10s
```

### Examples
Uploading a file with progress visualization:
```bash
ksau-go upload --file /path/to/local/file --remote /path/to/remote/folder --progress modern
```

Listing available remotes:
```bash
ksau-go list-remotes
```

Displaying OneDrive quota information:
```bash
ksau-go quota
```

## Project Structure
- `.gitattributes`, `.gitignore`: Git configuration files.
- `go.mod`, `go.sum`: Go module files.
- `install.ps1`, `setup.sh`: Installation scripts for different platforms.
- `LICENSE`: License file.
- `main.go`: Main entry point of the application.
- `Makefile`: Build instructions.
- `README.md`: Documentation file.
- `.git/`, `.github/`: Git-related directories.
- `azure/`: Contains Azure-related code.
- `cmd/`: Contains command-line related code.
- `crypto/`: Contains cryptographic-related code.

## Contribution Guidelines
We welcome contributions! Please follow these guidelines:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes.
4. Submit a pull request with a detailed description of your changes.

## Motivation
The motivation behind ksau-go is to provide a robust and efficient tool for interacting with Microsoft Azure services, specifically OneDrive and SharePoint. It aims to solve problems related to large file uploads, metadata management, and storage quota information.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact Information
For further inquiries or support, please contact us at [Telegram](https://t.me/ksau_update)
