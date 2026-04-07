#!/bin/bash
# Build Web version for Linux

set -e

cd "$(dirname "$0")/.."

echo "[1/2] Building Web binary for Linux..."
go build -trimpath -ldflags "-s -w -buildid=" -tags webui -o dong-web ./cmd

echo "[2/2] Success: dong-web"
echo ""
echo "Run with: ./dong-web -all"
echo "Then open: http://127.0.0.1:18080"
