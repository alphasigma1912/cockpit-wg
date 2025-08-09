# Cockpit WireGuard Manager

Cockpit WireGuard Manager is a [Cockpit](https://cockpit-project.org/) plugin that simplifies WireGuard setup and administration on Linux servers.

[![Build Status](https://github.com/alphasigma1912/cockpit-wg/workflows/Build%20Cockpit%20WireGuard%20Manager/badge.svg)](https://github.com/alphasigma1912/cockpit-wg/actions)
[![GitHub release](https://img.shields.io/github/release/alphasigma1912/cockpit-wg.svg)](https://github.com/alphasigma1912/cockpit-wg/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Scope

- Automatically installs WireGuard on first run using least-privilege escalation.
- Manages WireGuard interfaces and peers.
- Safely edits configuration files with validation and atomic writes.
- Displays live traffic graphs.
- Supports a secure configuration exchange mechanism between nodes without opening additional network services.

## Supported Platforms

### Primary Targets (Linux)
- **x86_64** - Standard 64-bit Intel/AMD systems
- **ARM64** - 64-bit ARM systems (Raspberry Pi 4+, AWS Graviton, Apple Silicon)
- **ARMv7** - 32-bit ARM systems (Raspberry Pi 3, embedded devices)

### Development/Testing
- **Windows x86_64** - Cross-compilation and development
- **macOS x86_64/ARM64** - Development environment

## Installation

### Pre-built Packages (Recommended)

1. **Download** the latest release for your platform:
   ```bash
   # Linux x86_64
   wget https://github.com/alphasigma1912/cockpit-wg/releases/latest/download/cockpit-wg-linux-amd64.tar.gz
   
   # Linux ARM64 (Raspberry Pi 4, etc.)
   wget https://github.com/alphasigma1912/cockpit-wg/releases/latest/download/cockpit-wg-linux-arm64.tar.gz
   
   # Linux ARMv7 (Raspberry Pi 3, etc.)
   wget https://github.com/alphasigma1912/cockpit-wg/releases/latest/download/cockpit-wg-linux-armv7.tar.gz
   ```

2. **Extract and install**:
   ```bash
   tar -xzf cockpit-wg-linux-*.tar.gz
   sudo cp -r cockpit-wg-linux-*/ /usr/share/cockpit/cockpit-wg/
   sudo systemctl restart cockpit
   ```

3. **Access** via Cockpit web interface at `https://your-server:9090`

### Building from Source

See [docs/BUILD.md](docs/BUILD.md) for detailed build instructions.

**Quick build:**
```bash
# For current platform
make dist

# For ARM64 (Raspberry Pi 4, etc.)
make dist GOOS=linux GOARCH=arm64

# For ARMv7 (Raspberry Pi 3, etc.)  
make dist GOOS=linux GOARCH=arm GOARM=7

# For all platforms
make multi-arch
```

## Repository Layout

- **ui/** – React + PatternFly frontend built with Vite
- **bridge/** – Go backend (`wg-bridge`)
- **packaging/** – Distribution packaging files
- **ansible/** – Ansible role for mass deployment
- **schemas/** – JSON schemas for API and `.wgx` format
- **docs/** – Project documentation
- **scripts/** – Development helpers and build tools
- **.github/workflows/** – CI/CD automation

## Development

### Prerequisites
- **Node.js 18+** and npm
- **Go 1.22+**
- **Make** (optional)

### Local Development
```bash
# Clone repository
git clone https://github.com/alphasigma1912/cockpit-wg.git
cd cockpit-wg

# Build for development
make dist

# Run tests
make test

# Lint code
make lint

# Multi-architecture build
./scripts/build-multi-arch.sh        # Linux/macOS
.\scripts\build-multi-arch.ps1       # Windows
```

### CI/CD

This project uses GitHub Actions for automated building and testing:
- **Continuous Integration** - Tests and lints on every push/PR
- **Multi-platform Builds** - Builds for all supported architectures
- **Automated Releases** - Creates releases when tags are pushed
- **Security Scanning** - Automated dependency and security checks

## Security

See [SECURITY.md](SECURITY.md) for the threat model and security baseline.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test && make lint`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
