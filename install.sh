#!/bin/bash

# Script parameters
REPO_OWNER="ibigio"
REPO_NAME="shell-ai"
TOOL_NAME="shell-ai"
TOOL_SYMLINK="q"
INSTALL_DIR="$HOME/.local/bin"
HELP=false

# Check for help flag
if [ "$HELP" = true ]; then
    echo "shell-ai Installer Help!"
    echo "Usage: "
    echo "  -help <Shows this message>"
    echo "  -repoowner <Owner of the repo>"
    echo "  -reponame <Set the repository name>"
    echo "  -toolname <Set the tool name (inside the .zip build)>"
    echo "  -toolsymlink <Set name of the local executable>"
    exit 0
fi

# Detect if running as root
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
fi

# Ensure required tools are installed
if ! command -v curl &>/dev/null || ! command -v tar &>/dev/null; then
    echo "Error: curl and tar must be installed."
    exit 1
fi

# Detect system architecture
ARCH="$(uname -m)"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

# Adjust architecture name for compatibility
if [ "$ARCH" = "x86_64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
else
    ARCH="i386"
fi

# Fetch the latest release tag from GitHub API
API_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
LATEST_TAG=$(curl --silent $API_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Set the download URL based on the platform and latest release tag
DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/${TOOL_NAME}_${OS}_${ARCH}.tar.gz"

echo "Downloading from: $DOWNLOAD_URL"

# Download and extract the binary
curl -L "$DOWNLOAD_URL" -o "${TOOL_NAME}.tar.gz" || { echo "Download failed"; exit 1; }
mkdir -p "${TOOL_NAME}-temp"
tar xzf "${TOOL_NAME}.tar.gz" -C "${TOOL_NAME}-temp" || { echo "Extraction failed"; exit 1; }

# Move the binary to the installation directory
mkdir -p "$INSTALL_DIR"
mv "${TOOL_NAME}-temp/${TOOL_NAME}" "$INSTALL_DIR/$TOOL_SYMLINK"
chmod +x "$INSTALL_DIR/$TOOL_SYMLINK"

# Clean up
rm -rf "${TOOL_NAME}-temp"
rm -f "${TOOL_NAME}.tar.gz"

# Add the installation directory to PATH if needed
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "export PATH=\$PATH:$INSTALL_DIR" >> ~/.bashrc
    source ~/.bashrc
    echo "Added $INSTALL_DIR to PATH. Restart terminal or run 'source ~/.bashrc'."
fi

# Print success message
echo "$TOOL_NAME has been installed successfully (version: $LATEST_TAG)!"
