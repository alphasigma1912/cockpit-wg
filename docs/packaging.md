# Packaging

## Prerequisites
- [nfpm](https://nfpm.goreleaser.com/) installed (`go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest`)
- `make` build tools

## Build packages
```bash
# Assemble the plugin artifacts
make dist

# Set version used in nfpm.yaml
export VERSION=$(git describe --tags --always --dirty)

# Build .deb package
nfpm package --config packaging/nfpm.yaml --packager deb \
  --target dist/cockpit-wg_${VERSION}_amd64.deb

# Build .rpm package
nfpm package --config packaging/nfpm.yaml --packager rpm \
  --target dist/cockpit-wg-${VERSION}-1.x86_64.rpm
```

## Install
```bash
# Debian/Ubuntu
dpkg -i dist/cockpit-wg_${VERSION}_amd64.deb

# RHEL/Fedora
rpm -i dist/cockpit-wg-${VERSION}-1.x86_64.rpm

# Reload Cockpit
sudo systemctl restart cockpit
```
