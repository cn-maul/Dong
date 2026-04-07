#!/bin/bash
# Full scan and output JSON report

set -e

cd "$(dirname "$0")/.."

if [ ! -f "./dong" ]; then
    echo "Error: dong binary not found. Run build-cli.sh first."
    exit 1
fi

REPORT_NAME="report_full_$(date +%Y%m%d_%H%M%S)"
./dong -cli -all -o "$REPORT_NAME" -pretty
