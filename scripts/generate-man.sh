#!/bin/bash
set -euo pipefail

# Create man directory if it doesn't exist
mkdir -p man

# Check if we have a pre-built man page
if [ -f "man/mcp-server-dump.1" ]; then
    echo "Using existing man page"
    exit 0
fi

# Try to generate using help2man if available
if command -v help2man >/dev/null 2>&1; then
    echo "Generating man page with help2man"

    # Build binary for man page generation
    go build -o mcp-server-dump ./cmd/mcp-server-dump

    # Generate basic man page
    help2man --no-discard-stderr --output=man/mcp-server-dump.1 ./mcp-server-dump

    # Clean up
    rm -f mcp-server-dump

    echo "Man page generated with help2man"
else
    echo "help2man not found, man page generation skipped"
    echo "Manual man page should be present at man/mcp-server-dump.1"
    exit 1
fi