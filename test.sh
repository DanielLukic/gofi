#!/bin/bash

# Exit on error
set -e

echo "Running all tests..."

# Run tests in all packages verbosely
go test -v ./pkg/...

echo "\nAll tests passed!"
