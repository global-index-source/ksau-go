# Variables
APP_NAME := ksau-go
VERSION ?= 1.0.0-r2
COMMIT := $(shell git rev-parse --short HEAD)
WINDOWS_SHENANIGANS :=

# Add .exe for bleeding on Windblows
ifeq ($(OS),Windows_NT)
	APP_NAME := $(APP_NAME).exe
	WINDOWS_SHENANIGANS := coreutils.exe
endif

# Append coreutils.exe (has to be installed) on windows,
# otherwise this will freeze make
DATE := $(shell $(WINDOWS_SHENANIGANS) date -u +%Y-%m-%d)
LDFLAGS := -X 'github.com/global-index-source/ksau-go/cmd.Version=$(VERSION)' -X 'github.com/global-index-source/ksau-go/cmd.Commit=$(COMMIT)' -X 'github.com/global-index-source/ksau-go/cmd.Date=$(DATE)'

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)

# Only to be used in GitHub actions
build_gh_actions:
	@echo "notice: this is meant to be used in workflows"
	go mod tidy
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-linux-arm64
	GOOS=android GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-android-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)-windows-amd64.exe

# Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run -ldflags "$(LDFLAGS)" main.go

# Clean up the binary
clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)

# Display version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Help command
help:
	@echo "Makefile Commands:"
	@echo "  all               - Build the application (default target)"
	@echo "  build             - Build the application"
	@echo "  build_gh_actions  - Build for Linux and Windows (intended for GitHub actions)"
	@echo "  run               - Run the application with dynamic values"
	@echo "  clean             - Remove the built binary"
	@echo "  version           - Show version, commit, and date info"
	@echo "  help              - Show this help message"
