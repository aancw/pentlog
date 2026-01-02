#!/bin/sh
set -e

REPO="aancw/pentlog"
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"

GITHUB_HEADER=""
if [ -n "$GH_TOKEN" ]; then
    GITHUB_HEADER="Authorization: token $GH_TOKEN"
fi

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
if [ -n "$GITHUB_HEADER" ]; then
    RELEASE_JSON=$(curl -s -H "$GITHUB_HEADER" "$LATEST_URL")
else
    RELEASE_JSON=$(curl -s "$LATEST_URL")
fi

# Check for rate limit or error
if echo "$RELEASE_JSON" | grep -q "API rate limit"; then
    echo "Error: GitHub API rate limit exceeded. Please try again later."
    exit 1
fi


if [ -n "$GITHUB_HEADER" ]; then

    ASSET_ID=$(echo "$RELEASE_JSON" | grep -C 5 "$BINARY" | grep "\"id\":" | head -n 1 | tr -d " \"," | cut -d: -f2)
    
    if [ -n "$ASSET_ID" ]; then
        DOWNLOAD_URL="https://api.github.com/repos/$REPO/releases/assets/$ASSET_ID"
        IS_API_URL=1
    fi
fi

if [ -z "$DOWNLOAD_URL" ]; then
    DOWNLOAD_URL=$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "$BINARY" | head -n 1 | cut -d '"' -f 4)
fi

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $BINARY in latest release."
    echo "Release Data Snippet: $(echo "$RELEASE_JSON" | head -n 10)"
    exit 1
fi

echo "Downloading $BINARY from $DOWNLOAD_URL..."

# Create temp directory
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

# Download binary
if [ -n "$GITHUB_HEADER" ] && [ "$IS_API_URL" = "1" ]; then
    # Use -H "Accept: application/octet-stream" specifically for API URL
    curl -L -H "$GITHUB_HEADER" -H "Accept: application/octet-stream" -o "$TMP_DIR/pentlog" "$DOWNLOAD_URL" --progress-bar
elif [ -n "$GITHUB_HEADER" ]; then
    # Fallback for browser_url with auth (might fail on private S3)
    curl -L -H "$GITHUB_HEADER" -o "$TMP_DIR/pentlog" "$DOWNLOAD_URL" --progress-bar
else
    curl -L -o "$TMP_DIR/pentlog" "$DOWNLOAD_URL" --progress-bar
fi

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
