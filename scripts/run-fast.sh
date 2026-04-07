#!/bin/bash
# Fast scan and output JSON report

set -e

cd "$(dirname "$0")/.."

if [ ! -f "./dong" ]; then
    echo "Error: dong binary not found. Run build-cli.sh first."
    exit 1
fi

REPORT_NAME="report_fast_$(date +%Y%m%d_%H%M%S)"
./dong -cli -all -fast -o "$REPORT_NAME" -pretty
