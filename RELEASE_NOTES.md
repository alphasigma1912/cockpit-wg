# Cockpit WireGuard Manager Release

## Version: v1.0.0 (Build: 5e4e183)

### 📦 Build Artifacts

This release includes pre-built packages for multiple platforms:

#### **Ubuntu/Linux ARM64** (Your requested target):
- **File**: `cockpit-wg-linux-arm64.zip` (2.11 MB)
- **SHA256**: `24f7ead5a253017f7756378409d8671a51139d5e5b50c0404282db4aefc9c1bb`

#### All Platforms:
- **Linux AMD64**: `cockpit-wg-linux-amd64.zip` (2.23 MB)
- **Linux ARM64**: `cockpit-wg-linux-arm64.zip` (2.11 MB) ⭐
- **Linux ARMv7**: `cockpit-wg-linux-armv7.zip` (2.15 MB)
- **Windows AMD64**: `cockpit-wg-windows-amd64.zip` (2.22 MB)
- **macOS Intel**: `cockpit-wg-darwin-amd64.zip` (2.17 MB)
- **macOS Apple Silicon**: `cockpit-wg-darwin-arm64.zip` (2.09 MB)

### 📋 Installation Instructions for Ubuntu ARM64

1. Download `cockpit-wg-linux-arm64.zip`
2. Extract the archive:
   ```bash
   unzip cockpit-wg-linux-arm64.zip
   cd cockpit-wg-linux-arm64
   ```
3. Install to Cockpit plugins directory:
   ```bash
   sudo mkdir -p /usr/share/cockpit/wg
   sudo cp -r * /usr/share/cockpit/wg/
   sudo systemctl restart cockpit
   ```

### 🔐 Verification

Verify the download integrity using SHA256:
```bash
echo "24f7ead5a253017f7756378409d8671a51139d5e5b50c0404282db4aefc9c1bb  cockpit-wg-linux-arm64.zip" | sha256sum -c
```

### 📁 Package Contents

Each package contains:
- `wg-bridge` - Backend binary for WireGuard management
- `index.html` - Frontend web interface
- `manifest.json` - Cockpit plugin manifest
- `assets/` - Static assets (CSS, JS, fonts)

### 🚀 Features

- WireGuard interface management
- Peer configuration and monitoring
- Real-time metrics and traffic graphs
- Web-based administration interface
- Multi-platform support

### 🛠️ Build Information

- **Build Date**: August 9, 2025
- **Go Version**: Latest
- **Node.js/npm**: Latest
- **Commit**: 5e4e183
- **Build Script**: Multi-architecture automated build

---

**Note**: This build was specifically created for Ubuntu ARM64 as requested.
