# ksau-go

[![Go Version](https://img.shields.io/badge/go-1.23.4-blue)](https://golang.org/doc/go1.23)

## Installation
For Windows, run:
```
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/global-index-source/ksau-go/master/install.ps1").Content
```

## Build Instruction
### To build this project, you need two important thing:
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
