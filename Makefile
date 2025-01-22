# Variables
APP_NAME := ksau-go
VERSION ?= 1.0.0
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
LDFLAGS := -X 'github.com/ksauraj/ksau-oned-api/cmd.Version=$(VERSION)' -X 'github.com/ksauraj/ksau-oned-api/cmd.Commit=$(COMMIT)' -X 'github.com/ksauraj/ksau-oned-api/cmd.Date=$(DATE)'

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME)

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
	@echo "  all       - Build the application (default target)"
	@echo "  build     - Build the application"
	@echo "  run       - Run the application with dynamic values"
	@echo "  clean     - Remove the built binary"
	@echo "  version   - Show version, commit, and date info"
	@echo "  help      - Show this help message"
