#!/bin/bash
# Build CLI version for Linux

set -e

cd "$(dirname "$0")/.."

echo "[1/2] Building CLI binary for Linux..."
go build -trimpath -ldflags "-s -w -buildid=" -o dong ./cmd

echo "[2/2] Success: dong"
echo ""
echo "Run with: ./dong -all -pretty"
