#!/bin/bash
# Build script for Linux/macOS

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$SCRIPT_DIR/bin"

echo "Building Dong for skill..."

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux)
        BINARY_NAME="dong"
        ;;
    Darwin)
        BINARY_NAME="dong"
        ;;
    *)
        echo "Warning: Unsupported OS: $OS"
        BINARY_NAME="dong"
        ;;
esac

# Create bin directory
mkdir -p "$BIN_DIR"

# Build
cd "$PROJECT_ROOT"
echo "Building $BINARY_NAME..."
go build -trimpath -ldflags "-s -w -buildid=" -o "$BIN_DIR/$BINARY_NAME" ./cmd

# Make executable
chmod +x "$BIN_DIR/$BINARY_NAME"

echo ""
echo "✓ Build successful!"
echo "  Binary: $BIN_DIR/$BINARY_NAME"
echo "  Size: $(du -h "$BIN_DIR/$BINARY_NAME" | cut -f1)"
echo ""
echo "Test: $BIN_DIR/$BINARY_NAME -v"
