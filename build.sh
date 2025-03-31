#!/bin/bash

# Exit on error
set -e

# Build Go program
echo "Building Go program..."
CGO_ENABLED=0 go build -ldflags="-w -s" -o gofi cmd/gofi.go

# Make binary executable
chmod +x gofi

echo "Build complete!"
echo "Binary: ./gofi" 