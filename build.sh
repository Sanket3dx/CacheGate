#!/bin/bash

OUTPUT_DIR="builds"
mkdir -p $OUTPUT_DIR

# Windows 64-bit
echo "Building for Windows 64-bit..."
GOOS=windows GOARCH=amd64 go build -o $OUTPUT_DIR/CacheGate_x64.exe

# Windows 32-bit
echo "Building for Windows 32-bit..."
GOOS=windows GOARCH=386 go build -o $OUTPUT_DIR/CacheGate_x32.exe

# macOS 64-bit (Intel)
echo "Building for macOS (Intel 64-bit)..."
GOOS=darwin GOARCH=amd64 go build -o $OUTPUT_DIR/CacheGate_mac64_intel

# macOS 64-bit (ARM, Apple Silicon)
echo "Building for macOS (Apple Silicon 64-bit)..."
GOOS=darwin GOARCH=arm64 go build -o $OUTPUT_DIR/CacheGate_mac64_arm

# Linux 64-bit
echo "Building for Linux 64-bit..."
GOOS=linux GOARCH=amd64 go build -o $OUTPUT_DIR/CacheGate_linux64

# Linux 32-bit
echo "Building for Linux 32-bit..."
GOOS=linux GOARCH=386 go build -o $OUTPUT_DIR/CacheGate_linux32

echo "Build completed. Binaries can be found in the '$OUTPUT_DIR' directory."
