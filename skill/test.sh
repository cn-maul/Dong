#!/bin/bash
# Test script for Dong skill

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"

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
        BINARY_NAME="dong"
        ;;
esac

BINARY="$BIN_DIR/$BINARY_NAME"

if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found at $BINARY"
    echo "Run ./build.sh first"
    exit 1
fi

echo "====================================="
echo "  Dong Skill Test Suite"
echo "====================================="
echo ""

# Test version
echo "Test 1: Version check"
$BINARY -v
echo ""

# Test CPU
echo "Test 2: CPU detection"
echo "$ $BINARY -cli -cpu -pretty"
$BINARY -cli -cpu -pretty | head -20
echo ""

# Test memory
echo "Test 3: Memory detection"
echo "$ $BINARY -cli -memory -pretty"
$BINARY -cli -memory -pretty | head -15
echo ""

# Test software
echo "Test 4: Software detection"
echo "$ $BINARY -cli -software -pretty"
$BINARY -cli -software -pretty | head -25
echo ""

# Test fast scan
echo "Test 5: Fast scan"
echo "$ $BINARY -cli -all -fast -pretty"
$BINARY -cli -all -fast -pretty | head -30
echo ""

# Test full scan
echo "Test 6: Full scan with advanced diagnostics"
echo "$ $BINARY -cli -all -pretty"
$BINARY -cli -all -pretty | tail -50
echo ""

echo "====================================="
echo "  All tests completed!"
echo "====================================="
echo ""
echo "Binary: $BINARY"
echo "Size: $(du -h "$BINARY" | cut -f1)"
echo ""
