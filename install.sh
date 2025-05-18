#!/bin/bash

# Replace these values with your tool's information:
REPO_OWNER="erfannf"
REPO_NAME="shell-ai"
TOOL_NAME="shell-ai"
TOOL_SYMLINK="q"

# Detect the platform (architecture and OS)
ARCH="$(uname -m)"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

# Fetch the latest release tag from GitHub API
API_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"
LATEST_TAG=$(curl --silent $API_URL | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

# Set the download URL based on the platform and latest release tag
DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/${TOOL_NAME}_${OS}_${ARCH}.tar.gz"

echo $DOWNLOAD_URL

# # Download and extract the binary
curl -L "$DOWNLOAD_URL" -o "${TOOL_NAME}.tar.gz"
mkdir -p "${TOOL_NAME}-temp"
tar xzf "${TOOL_NAME}.tar.gz" -C "${TOOL_NAME}-temp"

# Make the binary executable
mv "${TOOL_NAME}-temp/${TOOL_NAME}" "/usr/local/bin/${TOOL_SYMLINK}"
chmod +x /usr/local/bin/"${TOOL_SYMLINK}"

# # Clean up
rm -rf "${TOOL_NAME}-temp"
rm "${TOOL_NAME}.tar.gz"

# Print success message
echo "The $TOOL_NAME has been installed successfully (version: $LATEST_TAG)."