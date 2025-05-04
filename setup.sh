#!/bin/bash

# Function to print error message and exit
error_exit() {
    echo "Error: $1" >&2
    exit 1
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required commands
for cmd in curl grep cut tr; do
    command_exists "$cmd" || error_exit "$cmd is required but not installed."
done

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS="linux";;
    Darwin*)    OS="darwin";;
    *)          error_exit "Unsupported operating system: ${OS}";;
esac

# Detect architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64*)    ARCH="amd64";;
    aarch64*)   ARCH="arm64";;
    arm64*)     ARCH="arm64";;
    *)          error_exit "Unsupported architecture: ${ARCH}";;
esac

echo "Detected system: $OS-$ARCH"

# Get latest release URL
echo "Fetching latest release information..."
API_URL="https://api.github.com/repos/global-index-source/ksau-go/releases/latest"
RELEASE_DATA=$(curl -s "$API_URL" || error_exit "Failed to fetch release information")

# Extract download URL for the appropriate binary
BINARY_NAME="ksau-go-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep "browser_download_url.*${BINARY_NAME}" | cut -d '"' -f 4)
if [ -z "$DOWNLOAD_URL" ]; then
    error_exit "Could not find download URL for ${BINARY_NAME}"
fi

# Create temporary directory
TMP_DIR=$(mktemp -d) || error_exit "Failed to create temporary directory"
trap 'rm -rf "$TMP_DIR"' EXIT

# Download binary
echo "Downloading ${BINARY_NAME}..."
curl -L "$DOWNLOAD_URL" -o "${TMP_DIR}/ksau-go" || error_exit "Failed to download binary"

# Make binary executable
chmod +x "${TMP_DIR}/ksau-go" || error_exit "Failed to make binary executable"

# Create configuration directory
if [ "$OS" = "linux" ] || [ "$OS" = "darwin" ]; then
    CONFIG_DIR="$HOME/.ksau/.conf"
else
    CONFIG_DIR="$HOME/AppData/Roaming/ksau/.conf"
fi

mkdir -p "$CONFIG_DIR" || error_exit "Failed to create config directory: $CONFIG_DIR"
echo "Created configuration directory: $CONFIG_DIR"

# Ask for installation preference
read -r -p "Do you want to install ksau-go system-wide? (requires sudo) [y/N] " response </dev/tty
response="${response,,}"  # Convert to lowercase

if [[ "$response" =~ ^[y]$ ]]; then
    # System-wide installation
    if ! command_exists sudo; then
        error_exit "sudo is required for system-wide installation but not installed"
    fi
    echo "Installing system-wide..."
    sudo mv "${TMP_DIR}/ksau-go" /usr/local/bin/ || error_exit "Failed to install binary"
    echo "ksau-go has been installed to /usr/local/bin/ksau-go"
else
    # Local installation
    mkdir -p "$HOME/.local/bin" || error_exit "Failed to create local bin directory"
    mv "${TMP_DIR}/ksau-go" "$HOME/.local/bin/" || error_exit "Failed to install binary"
    echo "ksau-go has been installed to $HOME/.local/bin/ksau-go"
    
    # Add to PATH if not already present
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo "Adding $HOME/.local/bin to PATH in shell configuration..."
        if [ -f "$HOME/.bashrc" ]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bashrc"
            echo "Please run 'source ~/.bashrc' to update your current shell"
        elif [ -f "$HOME/.zshrc" ]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
            echo "Please run 'source ~/.zshrc' to update your current shell"
        else
            echo "Warning: Could not find .bashrc or .zshrc to update PATH"
            echo "Please manually add $HOME/.local/bin to your PATH"
        fi
    fi
fi

echo "Installation complete!"
echo "Configuration directory is at: $CONFIG_DIR"
echo "Note: You will need to place your rclone.conf file in this directory."
