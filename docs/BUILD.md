# Build Instructions

## Automated Builds (GitHub Actions)

This project automatically builds for multiple platforms using GitHub Actions:

### Supported Platforms
- **Linux x86_64** - Standard 64-bit Intel/AMD systems
- **Linux ARM64** - 64-bit ARM systems (Raspberry Pi 4, Apple Silicon, etc.)
- **Linux ARMv7** - 32-bit ARM systems (Raspberry Pi 3, older ARM boards)
- **Windows x86_64** - For development/testing only
- **macOS x86_64/ARM64** - For development/testing only

### Automatic Builds
- **On every push** to `main` or `develop` branches
- **On pull requests** to `main`
- **On git tags** (creates releases)

### Download Pre-built Packages
1. Go to the [Releases](../../releases) page
2. Download the appropriate package for your platform
3. Extract and install (see installation instructions below)

## Local Development Builds

### Prerequisites
- **Node.js 18+** and npm
- **Go 1.22+**
- **Make** (optional, for convenience)

> The project uses these versions in `.github/workflows/test.yml` and `.github/workflows/release.yml`; install them locally for consistent building, linting, testing, and packaging.

### Quick Start
```bash
# Build for current platform (Linux recommended)
make dist

# Build for specific platform
make dist GOOS=linux GOARCH=arm64    # ARM64 Linux
make dist GOOS=linux GOARCH=arm      # ARMv7 Linux
make dist GOOS=windows GOARCH=amd64  # Windows (dev only)

# Build for all supported platforms
make multi-arch
```

### Manual Build (without Make)

#### 1. Build Frontend
```bash
cd ui
npm install
npm run build
```

#### 2. Build Backend

**For Linux x86_64:**
```bash
cd bridge
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags "-s -w" -o ../dist/cockpit-wg/wg-bridge
```

**For Linux ARM64 (Raspberry Pi 4, etc.):**
```bash
cd bridge
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags "-s -w" -o ../dist/cockpit-wg/wg-bridge
```

**For Linux ARMv7 (Raspberry Pi 3, etc.):**
```bash
cd bridge
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -ldflags "-s -w" -o ../dist/cockpit-wg/wg-bridge
```

#### 3. Assemble Package
```bash
# Create distribution directory
mkdir -p dist/cockpit-wg

# Copy frontend files
cp ui/manifest.json dist/cockpit-wg/
cp ui/dist/index.html dist/cockpit-wg/
cp -r ui/dist/assets dist/cockpit-wg/

# Make binary executable
chmod +x dist/cockpit-wg/wg-bridge
```

## Cross-Platform Development

### Windows Development
You can develop on Windows and cross-compile for Linux:

```powershell
# PowerShell
cd bridge
$env:CGO_ENABLED=0
$env:GOOS="linux"
$env:GOARCH="arm64"  # or "amd64", "arm"
go build -ldflags "-s -w" -o ..\dist\cockpit-wg\wg-bridge
```

### Docker-based Build
For consistent builds across platforms:

```bash
# Build in Docker container
docker run --rm -v $(pwd):/src -w /src \
  golang:1.22-alpine sh -c "
    apk add --no-cache nodejs npm make
    make multi-arch
  "
```

## Installation

### Manual Installation
```bash
# Extract package
tar -xzf cockpit-wg-linux-arm64.tar.gz

# Install plugin
sudo cp -r cockpit-wg-linux-arm64/ /usr/share/cockpit/cockpit-wg/

# Restart Cockpit
sudo systemctl restart cockpit
```

### Package Installation (Future)
```bash
# Debian/Ubuntu (when available)
sudo apt install cockpit-wg

# Red Hat/Fedora (when available)
sudo dnf install cockpit-wg
```

## Architecture-Specific Notes

### ARM64 (64-bit ARM)
- **Raspberry Pi 4** and newer
- **Apple Silicon Macs** (for development)
- **AWS Graviton** instances
- Most modern ARM servers

### ARMv7 (32-bit ARM)
- **Raspberry Pi 3** and older
- Many embedded ARM devices
- Note: Limited to 32-bit address space

### x86_64
- Standard Intel/AMD 64-bit systems
- Most common server architecture

## Troubleshooting

### Build Issues
```bash
# Clean and rebuild
make clean
make dist

# Check Go version
go version  # Should be 1.22+

# Check Node.js version
node --version  # Should be 18+
```

### Platform Detection
```bash
# Check your current platform
uname -m  # Architecture (x86_64, aarch64, armv7l)
uname -s  # OS (Linux, Darwin, etc.)

# Check what Go would build for
go env GOOS GOARCH
```

### Performance Notes
- ARM builds may be slower during compilation
- ARM64 generally performs better than ARMv7
- Use appropriate GOARM version for ARMv7 (6 or 7)

## Development Workflow

1. **Make changes** to code
2. **Test locally**: `make test && make lint`
3. **Build for target platform**: `make dist GOOS=linux GOARCH=arm64`
4. **Push to GitHub** - automatic CI/CD builds all platforms
5. **Create tag** for release: `git tag v1.0.0 && git push --tags`
