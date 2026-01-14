#!/bin/sh
set -e

cat << 'EOF'
 ███████████                       █████    █████                        
░░███░░░░░███                     ░░███    ░░███                         
 ░███    ░███  ██████  ████████   ███████   ░███         ██████   ███████
 ░██████████  ███░░███░░███░░███ ░░░███░    ░███        ███░░███ ███░░███
 ░███░░░░░░  ░███████  ░███ ░███   ░███     ░███       ░███ ░███░███ ░███
 ░███        ░███░░░   ░███ ░███   ░███ ███ ░███      █░███ ░███░███ ░███
 █████       ░░██████  ████ █████  ░░█████  ███████████░░██████ ░░███████
░░░░░         ░░░░░░  ░░░░ ░░░░░    ░░░░░  ░░░░░░░░░░░  ░░░░░░   ░░░░░███
                                                                 ███ ░███
                                                                ░░██████ 
                                                                 ░░░░░░  
                 PentLog — Evidence-First Pentest Logging Tool
                                        created by Petruknisme

EOF

REPO="aancw/pentlog"
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"

# Detect OS and Architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

echo "Detected OS: $OS, Architecture: $ARCH"

if [ "$OS" = "linux" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        BINARY="pentlog-linux-amd64"
    else
        echo "Error: Unsupported architecture for Linux: $ARCH"
        exit 1
    fi
elif [ "$OS" = "darwin" ]; then
    if [ "$ARCH" = "x86_64" ]; then
        BINARY="pentlog-darwin-amd64"
    elif [ "$ARCH" = "arm64" ]; then
        BINARY="pentlog-darwin-arm64"
    else
        echo "Error: Unsupported architecture for macOS: $ARCH"
        exit 1
    fi
else
    echo "Error: Unsupported OS: $OS"
    exit 1
fi

echo "Fetching latest release for $BINARY..."

# Fetch release data
RELEASE_JSON=$(curl -s "$LATEST_URL")

# Extract version
VERSION=$(echo "$RELEASE_JSON" | grep "\"tag_name\":" | head -n 1 | cut -d '"' -f 4)

# Check for rate limit or error
if echo "$RELEASE_JSON" | grep -q "API rate limit"; then
    echo "Error: GitHub API rate limit exceeded. Please try again later."
    exit 1
fi

DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "$BINARY" | head -n 1 | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $BINARY in latest release."
    echo "Release Data Snippet: $(echo "$RELEASE_JSON" | head -n 10)"
    exit 1
fi

echo "Downloading $BINARY ($VERSION)..."

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download binary
curl -L --progress-bar -o "$TMP_DIR/pentlog" "$DOWNLOAD_URL"

# Make executable
chmod +x "$TMP_DIR/pentlog"

# Install location (User local bin, no sudo required)
INSTALL_DIR="$HOME/.local/bin"

if [ ! -d "$INSTALL_DIR" ]; then
    echo "Creating $INSTALL_DIR..."
    mkdir -p "$INSTALL_DIR"
fi

echo "Installing to $INSTALL_DIR..."

mv "$TMP_DIR/pentlog" "$INSTALL_DIR/pentlog"

echo ""
echo "✅ pentlog installed successfully to $INSTALL_DIR/pentlog"
echo ""

# Check if INSTALL_DIR is in PATH
if echo ":$PATH:" | grep -Fq ":$INSTALL_DIR:"; then
    echo "Get started:"
    echo "  pentlog setup    # Initialize configuration"
    echo "  pentlog create   # Start an engagement"
else
    echo "⚠️  Warning: $INSTALL_DIR is not in your PATH."
    echo "Add the following line to your shell configuration (.zshrc, .bashrc, etc.):"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    echo ""
    echo "Then restart your shell or run source <config_file>."
fi
