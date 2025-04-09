#!/bin/bash

# Exit on error
set -e

OUTPUT_NAME="gofi"
SOURCE_DIR="./cmd/*.go"

echo "Building release executable: ${OUTPUT_NAME}..."

# Build with flags to strip debug info and symbols, and remove paths
go build -ldflags="-s -w" -trimpath -o "${OUTPUT_NAME}" ${SOURCE_DIR}

# Make executable (optional, but good practice)
chmod +x "${OUTPUT_NAME}"

# Get final size
SIZE=$(du -h "${OUTPUT_NAME}" | cut -f1)

echo "Release build complete: ${OUTPUT_NAME} (${SIZE})"
