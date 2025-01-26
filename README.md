# ksau-go

[![Go Version](https://img.shields.io/badge/go-1.23.4-blue)](https://golang.org/doc/go1.23)

## Installation

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

## Configuration
The tool stores its configuration in:
- Linux/macOS: `$HOME/.ksau/.conf/rclone.conf`
- Windows: `%AppData%\ksau\.conf\rclone.conf`

## Build Instructions
### To build this project, you need two important things:
1. Private PGP key used to decrypt rclone.conf
2. The passphrase of the PGP key

### They should be placed under **crypto/** like so:
```
└───crypto
        algo.go
        placeholder.go
        >> passphrase.txt  ⌉  -- These files
        >> privkey.pem     ⌋     they are not provided by the repo
```

### Finally, install the dependencies and you're ready to build the project!
```
go mod tidy  # install dependencies
make build   # build the project
```

Depending on the OS you're on, you'll see `ksau-go` or `ksau-go.exe` generated
in the current working directory.
