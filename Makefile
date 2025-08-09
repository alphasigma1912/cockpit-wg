UI_DIR := ui
BRIDGE_DIR := bridge
DIST_DIR := dist/cockpit-wg
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Default target architecture (can be overridden)
GOOS ?= linux
GOARCH ?= amd64
GOARM ?= 7

# Binary name with architecture suffix for multi-arch builds
ifeq ($(GOOS),windows)
	BINARY_EXT := .exe
else
	BINARY_EXT :=
endif

BINARY_NAME := wg-bridge$(BINARY_EXT)
BINARY_ARCH_NAME := wg-bridge-$(GOOS)-$(GOARCH)$(BINARY_EXT)

.PHONY: ui bridge dist clean help test lint multi-arch-build

# Default target
all: dist

# Help target
help:
	@echo "Available targets:"
	@echo "  ui              - Build frontend only"
	@echo "  bridge          - Build backend for current platform"
	@echo "  dist            - Build complete plugin for current platform"
	@echo "  multi-arch      - Build for all supported architectures"
	@echo "  test            - Run all tests"
	@echo "  lint            - Run linters"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS            - Target OS (linux, windows, darwin)"
	@echo "  GOARCH          - Target architecture (amd64, arm64, arm)"
	@echo "  GOARM           - ARM version (6, 7) when GOARCH=arm"
	@echo "  VERSION         - Version string (default: git describe)"
	@echo ""
	@echo "Examples:"
	@echo "  make dist GOOS=linux GOARCH=arm64    # ARM64 Linux build"
	@echo "  make dist GOOS=linux GOARCH=arm      # ARMv7 Linux build"
	@echo "  make multi-arch                      # All architectures"

# Build frontend
ui:
	@echo "Building frontend..."
	cd $(UI_DIR) && npm install
	cd $(UI_DIR) && npm run build

# Build backend for specified architecture
bridge:
	@echo "Building backend for $(GOOS)/$(GOARCH)..."
	mkdir -p $(DIST_DIR)
	cd $(BRIDGE_DIR) && CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GOARM=$(GOARM) \
		go build -ldflags "-s -w -X main.version=$(VERSION)" \
		-o ../$(DIST_DIR)/$(BINARY_NAME)

# Build complete plugin package
dist: ui bridge
	@echo "Assembling plugin package..."
	mkdir -p $(DIST_DIR)
	cp $(UI_DIR)/manifest.json $(DIST_DIR)/manifest.json
	cp $(UI_DIR)/dist/index.html $(DIST_DIR)/index.html
	cp -r $(UI_DIR)/dist/assets $(DIST_DIR)/assets
	@echo "Plugin package ready in $(DIST_DIR)"

# Build for all supported architectures
multi-arch: ui
	@echo "Building for multiple architectures..."
	# Linux builds
	$(MAKE) bridge GOOS=linux GOARCH=amd64 DIST_DIR=dist/cockpit-wg-linux-amd64
	$(MAKE) bridge GOOS=linux GOARCH=arm64 DIST_DIR=dist/cockpit-wg-linux-arm64
	$(MAKE) bridge GOOS=linux GOARCH=arm GOARM=7 DIST_DIR=dist/cockpit-wg-linux-armv7
	
	# Windows builds (for development)
	$(MAKE) bridge GOOS=windows GOARCH=amd64 DIST_DIR=dist/cockpit-wg-windows-amd64
	
	# macOS builds (for development)
	$(MAKE) bridge GOOS=darwin GOARCH=amd64 DIST_DIR=dist/cockpit-wg-darwin-amd64
	$(MAKE) bridge GOOS=darwin GOARCH=arm64 DIST_DIR=dist/cockpit-wg-darwin-arm64
	
	# Assemble packages
	for target in linux-amd64 linux-arm64 linux-armv7 windows-amd64 darwin-amd64 darwin-arm64; do \
		mkdir -p dist/cockpit-wg-$$target; \
		cp $(UI_DIR)/manifest.json dist/cockpit-wg-$$target/; \
		cp $(UI_DIR)/dist/index.html dist/cockpit-wg-$$target/; \
		cp -r $(UI_DIR)/dist/assets dist/cockpit-wg-$$target/; \
		if [ "$$target" = "windows-amd64" ]; then \
			cp dist/cockpit-wg-$$target/wg-bridge.exe dist/cockpit-wg-$$target/wg-bridge.exe; \
		else \
			chmod +x dist/cockpit-wg-$$target/wg-bridge; \
		fi; \
		tar -czf dist/cockpit-wg-$$target.tar.gz -C dist cockpit-wg-$$target/; \
	done
	@echo "Multi-architecture packages ready in dist/"

# Run tests
test:
	@echo "Running frontend tests..."
	cd $(UI_DIR) && npm install
	cd $(UI_DIR) && if grep -q "\"test\"" package.json; then npm test; else echo "No frontend tests configured"; fi
	
	@echo "Running backend tests..."
	cd $(BRIDGE_DIR) && go test -v ./...

# Run linters
lint:
	@echo "Linting frontend..."
	cd $(UI_DIR) && npm install
	cd $(UI_DIR) && if grep -q "\"lint\"" package.json; then npm run lint; else npx tsc --noEmit; fi
	
	@echo "Linting backend..."
	cd $(BRIDGE_DIR) && go vet ./...
	cd $(BRIDGE_DIR) && go fmt ./...

# Clean build artifacts
clean:
	rm -rf dist $(UI_DIR)/node_modules $(UI_DIR)/dist

# Development shortcuts
dev-linux-amd64:
	$(MAKE) dist GOOS=linux GOARCH=amd64

dev-linux-arm64:
	$(MAKE) dist GOOS=linux GOARCH=arm64

dev-windows:
	$(MAKE) dist GOOS=windows GOARCH=amd64
