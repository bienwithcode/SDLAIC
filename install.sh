#!/bin/sh
set -e

# Detect OS and Architecture
OS_NAME=$(uname -s)
ARCH_NAME=$(uname -m)

case "$OS_NAME" in
    Darwin)
        OS="Darwin"
        ;;
    Linux)
        OS="Linux"
        ;;
    *)
        echo "Unsupported operating system: $OS_NAME"
        exit 1
        ;;
esac

case "$ARCH_NAME" in
    x86_64|amd64)
        ARCH="x86_64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH_NAME"
        exit 1
        ;;
esac

# Repository details
REPO="bienwithcode/SDLAIC"
GITHUB_API="https://api.github.com/repos/$REPO/releases/latest"

echo "Fetching latest release info..."
# Get latest release tag
TAG=$(curl -fsSL "$GITHUB_API" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
    echo "Error: Could not retrieve latest release tag."
    exit 1
fi

VERSION="${TAG#v}"
echo "Latest version: $VERSION"

FILENAME="sdlaic_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$FILENAME"

echo "Downloading $FILENAME..."
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -fsSL "$URL" -o "$TMP_DIR/$FILENAME"

echo "Extracting..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

# Install binary
INSTALL_DIR="/usr/local/bin"
# Check if /usr/local/bin is writeable, otherwise fallback to ~/.local/bin
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo "Warning: /usr/local/bin is not writeable. Installing to $INSTALL_DIR instead."
    echo "Please ensure $INSTALL_DIR is in your PATH."
fi

mv "$TMP_DIR/sdlaic" "$INSTALL_DIR/sdlaic"
chmod +x "$INSTALL_DIR/sdlaic"

echo "Successfully installed sdlaic to $INSTALL_DIR/sdlaic"
