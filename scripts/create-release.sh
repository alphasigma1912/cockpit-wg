#!/bin/bash
# GitHub Release Creation Script
# Run this script to create a GitHub release with all build artifacts

set -e

VERSION="v1.0.0"
TAG="v1.0.0"
RELEASE_TITLE="Cockpit WireGuard Manager v1.0.0 - Multi-Platform Release"
RELEASE_BODY="# Cockpit WireGuard Manager Release

## Features
- WireGuard interface management through Cockpit web interface
- Real-time metrics and traffic monitoring
- Multi-platform support (Linux, Windows, macOS)
- **Special build for Ubuntu ARM64 included** 🎯

## Ubuntu ARM64 Installation
\`\`\`bash
# Download and extract
wget https://github.com/alphasigma1912/cockpit-wg/releases/download/${TAG}/cockpit-wg-linux-arm64.zip
unzip cockpit-wg-linux-arm64.zip
cd cockpit-wg-linux-arm64

# Install to Cockpit
sudo mkdir -p /usr/share/cockpit/wg
sudo cp -r * /usr/share/cockpit/wg/
sudo systemctl restart cockpit
\`\`\`

## Package Verification
All packages include SHA256 checksums for integrity verification.

## Build Information
- Build: 5e4e183
- Built: August 9, 2025
- Platforms: 6 architectures supported"

DIST_DIR="dist"

echo "🚀 Creating GitHub Release: $VERSION"
echo "📁 Using artifacts from: $DIST_DIR"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Check if dist directory exists
if [ ! -d "$DIST_DIR" ]; then
    echo "❌ Error: Build directory $DIST_DIR not found"
    echo "Run the build script first: ./scripts/build-multi-arch.sh"
    exit 1
fi

# Create and push tag
echo "🏷️  Creating tag: $TAG"
git tag -a "$TAG" -m "$RELEASE_TITLE"
git push origin "$TAG"

echo "✅ Tag created and pushed successfully"
echo ""
echo "📋 Manual Release Steps:"
echo "1. Go to: https://github.com/alphasigma1912/cockpit-wg/releases/new"
echo "2. Select tag: $TAG"
echo "3. Set title: $RELEASE_TITLE"
echo "4. Copy the release body from RELEASE_NOTES.md"
echo "5. Upload these files from $DIST_DIR/:"
echo "   - cockpit-wg-linux-arm64.zip (Ubuntu ARM64 - PRIMARY TARGET)"
echo "   - cockpit-wg-linux-amd64.zip"
echo "   - cockpit-wg-linux-armv7.zip"
echo "   - cockpit-wg-windows-amd64.zip"
echo "   - cockpit-wg-darwin-amd64.zip"
echo "   - cockpit-wg-darwin-arm64.zip"
echo "   - checksums.txt"
echo ""
echo "🎯 Ubuntu ARM64 package ready: $DIST_DIR/cockpit-wg-linux-arm64.zip"
