#!/bin/bash
# Local multi-architecture build and test script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PLATFORMS=(
    "linux/amd64"
    "linux/arm64" 
    "linux/arm/v7"
    "windows/amd64"
    "darwin/amd64"
    "darwin/arm64"
)

VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
BUILD_DIR="dist"
TEMP_DIR=$(mktemp -d)

echo -e "${BLUE}🚀 Multi-Architecture Build Script${NC}"
echo -e "${BLUE}Version: ${VERSION}${NC}"
echo -e "${BLUE}Build directory: ${BUILD_DIR}${NC}"
echo ""

# Check prerequisites
check_prerequisites() {
    echo -e "${YELLOW}📋 Checking prerequisites...${NC}"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        exit 1
    fi
    
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js is not installed${NC}"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}❌ npm is not installed${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All prerequisites found${NC}"
    echo ""
}

# Build frontend
build_frontend() {
    echo -e "${YELLOW}🎨 Building frontend...${NC}"
    
    cd ui
    npm install
    npm run build
    cd ..
    
    echo -e "${GREEN}✅ Frontend built successfully${NC}"
    echo ""
}

# Build backend for specific platform
build_backend() {
    local platform=$1
    local goos=$(echo $platform | cut -d'/' -f1)
    local goarch=$(echo $platform | cut -d'/' -f2)
    local goarm=""
    
    if [[ $platform == *"/v7" ]]; then
        goarch="arm"
        goarm="7"
    elif [[ $platform == *"/v6" ]]; then
        goarch="arm"
        goarm="6"
    fi
    
    local binary_name="wg-bridge"
    if [[ $goos == "windows" ]]; then
        binary_name="wg-bridge.exe"
    fi
    
    local target_name="${goos}-${goarch}"
    if [[ -n $goarm ]]; then
        target_name="${goos}-${goarch}v${goarm}"
    fi
    
    echo -e "${YELLOW}🔨 Building backend for ${target_name}...${NC}"
    
    local build_dir="${BUILD_DIR}/cockpit-wg-${target_name}"
    mkdir -p "$build_dir"
    
    cd bridge
    CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch GOARM=$goarm \
        go build -ldflags "-s -w -X main.version=${VERSION}" \
        -o "../${build_dir}/${binary_name}" .
    cd ..
    
    # Verify binary was created
    if [[ ! -f "${build_dir}/${binary_name}" ]]; then
        echo -e "${RED}❌ Failed to build ${target_name}${NC}"
        return 1
    fi
    
    # Make executable (for non-Windows)
    if [[ $goos != "windows" ]]; then
        chmod +x "${build_dir}/${binary_name}"
    fi
    
    echo -e "${GREEN}✅ Backend built for ${target_name}${NC}"
    return 0
}

# Package complete plugin
package_plugin() {
    local platform=$1
    local goos=$(echo $platform | cut -d'/' -f1)
    local goarch=$(echo $platform | cut -d'/' -f2)
    local goarm=""
    
    if [[ $platform == *"/v7" ]]; then
        goarch="arm"
        goarm="7"
    fi
    
    local target_name="${goos}-${goarch}"
    if [[ -n $goarm ]]; then
        target_name="${goos}-${goarch}v${goarm}"
    fi
    
    local package_dir="${BUILD_DIR}/cockpit-wg-${target_name}"
    
    echo -e "${YELLOW}📦 Packaging plugin for ${target_name}...${NC}"
    
    # Copy frontend files
    cp ui/manifest.json "$package_dir/"
    cp ui/dist/index.html "$package_dir/"
    cp -r ui/dist/assets "$package_dir/"
    
    # Create tarball
    cd "$BUILD_DIR"
    tar -czf "cockpit-wg-${target_name}.tar.gz" "cockpit-wg-${target_name}/"
    cd ..
    
    # Get size info
    local size=$(du -h "${BUILD_DIR}/cockpit-wg-${target_name}.tar.gz" | cut -f1)
    echo -e "${GREEN}✅ Package created: cockpit-wg-${target_name}.tar.gz (${size})${NC}"
}

# Test binary (basic smoke test)
test_binary() {
    local platform=$1
    local goos=$(echo $platform | cut -d'/' -f1)
    local goarch=$(echo $platform | cut -d'/' -f2)
    
    # Skip testing for non-native platforms for now
    local native_os=$(go env GOOS)
    local native_arch=$(go env GOARCH)
    
    if [[ $goos != $native_os ]] || [[ $goarch != $native_arch ]]; then
        echo -e "${YELLOW}⏭️  Skipping test for ${goos}/${goarch} (cross-compiled)${NC}"
        return 0
    fi
    
    local target_name="${goos}-${goarch}"
    local binary_path="${BUILD_DIR}/cockpit-wg-${target_name}/wg-bridge"
    
    if [[ $goos == "windows" ]]; then
        binary_path="${binary_path}.exe"
    fi
    
    echo -e "${YELLOW}🧪 Testing binary for ${target_name}...${NC}"
    
    # Basic test - just try to run with --help
    if timeout 5 "$binary_path" --help &>/dev/null; then
        echo -e "${GREEN}✅ Binary test passed for ${target_name}${NC}"
    else
        echo -e "${YELLOW}⚠️  Binary test skipped or failed for ${target_name}${NC}"
    fi
}

# Generate checksums
generate_checksums() {
    echo -e "${YELLOW}🔐 Generating checksums...${NC}"
    
    cd "$BUILD_DIR"
    if ls *.tar.gz 1> /dev/null 2>&1; then
        sha256sum *.tar.gz > checksums.txt
        echo -e "${GREEN}✅ Checksums generated${NC}"
    else
        echo -e "${YELLOW}⚠️  No packages found for checksum generation${NC}"
    fi
    cd ..
}

# Clean up
cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Main build process
main() {
    trap cleanup EXIT
    
    check_prerequisites
    
    # Clean previous builds
    echo -e "${YELLOW}🧹 Cleaning previous builds...${NC}"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    build_frontend
    
    local failed_builds=()
    local successful_builds=()
    
    for platform in "${PLATFORMS[@]}"; do
        if build_backend "$platform"; then
            package_plugin "$platform"
            test_binary "$platform"
            successful_builds+=("$platform")
        else
            failed_builds+=("$platform")
        fi
        echo ""
    done
    
    generate_checksums
    
    echo -e "${BLUE}📊 Build Summary${NC}"
    echo -e "${GREEN}✅ Successful builds (${#successful_builds[@]}):${NC}"
    for platform in "${successful_builds[@]}"; do
        echo -e "   • $platform"
    done
    
    if [[ ${#failed_builds[@]} -gt 0 ]]; then
        echo -e "${RED}❌ Failed builds (${#failed_builds[@]}):${NC}"
        for platform in "${failed_builds[@]}"; do
            echo -e "   • $platform"
        done
    fi
    
    echo ""
    echo -e "${BLUE}📁 Build artifacts in: ${BUILD_DIR}${NC}"
    ls -la "$BUILD_DIR"/*.tar.gz 2>/dev/null || echo "No packages created"
}

# Run main function
main "$@"
