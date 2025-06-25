#!/bin/bash

# GopherStrike Build Script
# This script builds the GopherStrike binary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building GopherStrike...${NC}"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Clean up old binaries
rm -f GopherStrike gopherstrike

# Build the binary
echo "Running go build..."
go build -o GopherStrike main.go

if [ $? -eq 0 ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo "Binary created: $SCRIPT_DIR/GopherStrike"
    
    # Make it executable
    chmod +x GopherStrike
    
    echo ""
    echo "Next steps:"
    echo "1. Run './install.sh' to install GopherStrike to your system"
    echo "2. Or run './GopherStrike' directly from this directory"
else
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi