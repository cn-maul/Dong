#!/bin/bash
# Install script for Linux/macOS

set -e

SKILL_NAME="dong"
SKILL_DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_DIR="$HOME/.claude/skills/$SKILL_NAME"
BIN_DIR="$SKILL_DIR/bin"

echo "====================================="
echo "  Dong Skill Installer"
echo "====================================="
echo ""

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

# Check if binary exists
if [ ! -f "$BIN_DIR/$BINARY_NAME" ]; then
    echo "✗ Binary not found: $BIN_DIR/$BINARY_NAME"
    echo ""
    echo "The skill package should include pre-compiled binaries."
    echo "Please download the complete skill package."
    echo ""
    echo "If you have the source code, you can build with:"
    echo "  ./build.sh"
    echo ""
    exit 1
fi

echo "✓ Binary found: $BIN_DIR/$BINARY_NAME"
echo ""

# Check if target already exists
if [ -e "$TARGET_DIR" ]; then
    echo "⚠ Target directory already exists: $TARGET_DIR"
    echo ""
    read -p "Remove and reinstall? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$TARGET_DIR"
        echo "✓ Removed existing installation"
    else
        echo "Installation aborted."
        exit 1
    fi
fi

# Create parent directory if needed
mkdir -p "$HOME/.claude/skills"

# Create symbolic link
echo "Creating symbolic link..."
ln -s "$SKILL_DIR" "$TARGET_DIR"

echo ""
echo "====================================="
echo "  ✓ Installation Successful!"
echo "====================================="
echo ""
echo "  Installed to: $TARGET_DIR"
echo "  Skill name: $SKILL_NAME"
echo "  Platform: $OS"
echo "  Binary: $BINARY_NAME (pre-compiled)"
echo ""
echo "The skill will be available after restarting Claude Code."
echo ""
echo "Usage in Claude Code:"
echo "  /dong -v"
echo "  /dong -cli -cpu -pretty"
echo "  /dong -cli -all -pretty"
echo ""
