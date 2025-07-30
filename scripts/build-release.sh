#!/bin/bash

# Script to build release binaries for all supported platforms

set -e

# Create dist directory if it doesn't exist
mkdir -p dist

# Build for each platform
echo "Building for darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -o dist/iamctl-darwin-amd64 .

echo "Building for darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -o dist/iamctl-darwin-arm64 .

echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o dist/iamctl-linux-amd64 .

echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 go build -o dist/iamctl-windows-amd64.exe .

echo "Build complete! Binaries are in the dist/ directory."