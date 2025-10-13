#!/bin/bash

# Jot Installation Script
# This script builds and installs jot globally on your system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="jot"

echo -e "${GREEN}=== Jot Installation Script ===${NC}"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    echo "Please install Go from https://golang.org/dl/"
    exit 1
fi

# Display Go version
GO_VERSION=$(go version)
echo -e "${GREEN}Found Go:${NC} $GO_VERSION"
echo ""

# Check if running from the jot directory
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: go.mod not found${NC}"
    echo "Please run this script from the jot project root directory"
    exit 1
fi

# Check if the module is correct
MODULE_NAME=$(grep "^module" go.mod | awk '{print $2}')
if [[ "$MODULE_NAME" != *"jot"* ]]; then
    echo -e "${YELLOW}Warning: Unexpected module name: $MODULE_NAME${NC}"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}Building jot...${NC}"
# Build the binary
if go build -o "$BINARY_NAME" ./cmd/jot; then
    echo -e "${GREEN}✓ Build successful${NC}"
else
    echo -e "${RED}✗ Build failed${NC}"
    exit 1
fi

# Check if install directory exists
if [ ! -d "$INSTALL_DIR" ]; then
    echo -e "${YELLOW}Directory $INSTALL_DIR does not exist.${NC}"
    echo "Creating directory (may require sudo)..."
    sudo mkdir -p "$INSTALL_DIR"
fi

# Check if we need sudo for installation
if [ -w "$INSTALL_DIR" ]; then
    echo -e "${GREEN}Installing to $INSTALL_DIR...${NC}"
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    echo -e "${YELLOW}Installing to $INSTALL_DIR (requires sudo)...${NC}"
    sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify installation
if command -v jot &> /dev/null; then
    INSTALLED_VERSION=$(jot --version 2>&1 || echo "version unknown")
    echo ""
    echo -e "${GREEN}=== Installation Complete ===${NC}"
    echo -e "${GREEN}✓${NC} jot has been installed to: $INSTALL_DIR/$BINARY_NAME"
    echo -e "${GREEN}✓${NC} Version: $INSTALLED_VERSION"
    echo ""
    echo "You can now use 'jot' from anywhere in your terminal."
    echo "Try running: jot --help"
else
    echo ""
    echo -e "${YELLOW}Warning: jot was installed but is not in your PATH${NC}"
    echo "Add $INSTALL_DIR to your PATH by adding this line to your shell profile:"
    echo ""
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
    echo "For bash, add to ~/.bashrc or ~/.bash_profile"
    echo "For zsh, add to ~/.zshrc"
fi