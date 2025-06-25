#!/bin/bash

# GopherStrike Installation Script
# This script installs GopherStrike globally so you can run it from anywhere

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root for system-wide installation
if [[ $EUID -eq 0 ]]; then
    INSTALL_DIR="/usr/local/bin"
    echo -e "${YELLOW}Installing GopherStrike system-wide...${NC}"
else
    # Create local bin directory if it doesn't exist
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
    echo -e "${YELLOW}Installing GopherStrike for current user...${NC}"
fi

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY_PATH="$SCRIPT_DIR/GopherStrike"

# Check if binary exists (try both uppercase and lowercase)
if [[ -f "$BINARY_PATH" ]]; then
    # Found uppercase version
    :
elif [[ -f "$SCRIPT_DIR/gopherstrike" ]]; then
    # Found lowercase version
    BINARY_PATH="$SCRIPT_DIR/gopherstrike"
else
    echo -e "${RED}Error: GopherStrike binary not found${NC}"
    echo "Please build the binary first with: ./build.sh or go build -o GopherStrike"
    exit 1
fi

# Copy binary to install directory
echo "Copying GopherStrike to $INSTALL_DIR..."
cp "$BINARY_PATH" "$INSTALL_DIR/gopherstrike"
chmod +x "$INSTALL_DIR/gopherstrike"

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: $INSTALL_DIR is not in your PATH${NC}"
    
    if [[ $EUID -ne 0 ]]; then
        echo "Add the following line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
        echo "Or run: echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.$(basename $SHELL)rc"
    fi
fi

echo -e "${GREEN}GopherStrike installed successfully!${NC}"
echo "You can now run 'gopherstrike' from anywhere in your terminal."

# Test the installation
if command -v gopherstrike &> /dev/null; then
    echo -e "${GREEN}Installation verified - 'gopherstrike' command is available${NC}"
else
    echo -e "${YELLOW}Note: You may need to restart your terminal or reload your shell profile${NC}"
fi