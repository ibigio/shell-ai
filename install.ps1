# Script parameters
REPOOWNER="ibigio"
REPONAME="shell-ai"
TOOLNAME="shell-ai"
TOOLSYMLINK="q"
HELP=false

# Print help message if requested
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
if [ "$EUID" -ne 0 ]; then
    echo "Not running as root. Will attempt user-level installation..."
    INSTALL_DIR="$HOME/.local/bin"
else
    INSTALL_DIR="/usr/local/bin"
fi

# Check for required tools
command -v curl >/dev/null 2>&1 || { echo >&2 "curl is required but not installed. Aborting."; exit 1; }
command -v unzip >/dev/null 2>&1 || { echo >&2 "unzip is required but not installed. Aborting."; exit 1; }

# Detect the platform (architecture and OS)
ARCH=""
OS="linux"

if [ "$(uname -m)" = "x86_64" ]; then
    ARCH="x86_64"
elif [ "$(uname -m)" = "aarch64" ]; then
    ARCH="arm64"
else
    ARCH="i386"
fi

# Fetch the latest release tag from GitHub API
API_URL="https://api.github.com/repos/$REPOOWNER/$REPONAME/releases/latest"
LATEST_TAG=$(curl -s $API_URL | grep -oP '"tag_name": "\K(.*)(?=")')

# Set the download URL based on the platform and latest release tag
DOWNLOAD_URL="https://github.com/$REPOOWNER/$REPONAME/releases/download/$LATEST_TAG/${TOOLNAME}_${OS}_${ARCH}.zip"

echo "Downloading from: $DOWNLOAD_URL"

# Download the ZIP file
curl -L "$DOWNLOAD_URL" -o "${TOOLNAME}.zip"
if [ $? -ne 0 ]; then
    echo "Download failed. Please check the URL or your internet connection."
    exit 1
fi

# Extract the ZIP file
EXTRACTED_DIR="${TOOLNAME}-temp"
unzip -o "${TOOLNAME}.zip" -d "$EXTRACTED_DIR"
if [ $? -ne 0 ]; then
    echo "Failed to extract zip file."
    exit 1
fi

# Move the binary to the installation directory
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

mv "${EXTRACTED_DIR}/${TOOLNAME}" "$INSTALL_DIR/$TOOLSYMLINK"
chmod +x "$INSTALL_DIR/$TOOLSYMLINK"

# Clean up
rm -rf "${EXTRACTED_DIR}"
rm -f "${TOOLNAME}.zip"

# Add the installation directory to PATH if not already present
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    echo "export PATH=\$PATH:$INSTALL_DIR" >> ~/.bashrc
    source ~/.bashrc
    echo "Added $INSTALL_DIR to PATH. Restart your terminal or run 'source ~/.bashrc' to use the tool."
fi

# Print success message
echo "$TOOLNAME has been installed successfully (version: $LATEST_TAG)!"
