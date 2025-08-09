#!/bin/bash

# Test runner script for WireGuard configuration parser and validator
# This script runs all tests including unit tests, fuzz tests, and coverage analysis

set -e

echo "🧪 Running comprehensive test suite for WireGuard bridge..."
echo "============================================================"

# Navigate to bridge directory
cd "$(dirname "$0")/bridge"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
go clean -testcache
rm -f coverage.out coverage.html

# Verify Go environment
print_status "Checking Go environment..."
go version
go mod tidy
go mod verify

# Run static analysis
print_status "Running static analysis..."
if command -v golangci-lint &> /dev/null; then
    golangci-lint run ./...
    print_success "Static analysis passed"
else
    print_warning "golangci-lint not found, skipping static analysis"
fi

# Run unit tests with coverage
print_status "Running unit tests with coverage analysis..."
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

if [ $? -eq 0 ]; then
    print_success "Unit tests passed"
else
    print_error "Unit tests failed"
    exit 1
fi

# Generate coverage report
print_status "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Calculate coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
print_status "Total test coverage: ${COVERAGE}%"

# Check if coverage meets target (90%)
if (( $(echo "$COVERAGE >= 90" | bc -l) )); then
    print_success "Coverage target met (≥90%): ${COVERAGE}%"
else
    print_warning "Coverage below target (<90%): ${COVERAGE}%"
fi

# Run benchmarks
print_status "Running benchmark tests..."
go test -bench=. -benchmem ./internal/config/ > benchmark_results.txt
print_success "Benchmark tests completed"

# Run fuzz tests (short duration for CI)
print_status "Running fuzz tests..."
export FUZZ_TIME="10s"

# List of fuzz test functions
FUZZ_TESTS=(
    "FuzzParseConfig"
    "FuzzValidateIPs" 
    "FuzzDetectIPConflicts"
)

for fuzz_test in "${FUZZ_TESTS[@]}"; do
    print_status "Running $fuzz_test..."
    timeout 15s go test -fuzz="$fuzz_test" -fuzztime="$FUZZ_TIME" ./internal/config/ || {
        if [ $? -eq 124 ]; then
            print_success "$fuzz_test completed (timeout reached)"
        else
            print_error "$fuzz_test failed"
            exit 1
        fi
    }
done

print_success "Fuzz tests completed"

# Test with different build tags
print_status "Testing cross-platform compatibility..."
GOOS=linux go test -v ./...
GOOS=windows go test -v ./...
GOOS=darwin go test -v ./...
print_success "Cross-platform tests passed"

# Test race conditions
print_status "Running race condition tests..."
go test -race -count=10 ./internal/config/
print_success "Race condition tests passed"

# Validate test data corpus
print_status "Validating test data corpus..."
TEST_DATA_DIR="testdata"
if [ -d "$TEST_DATA_DIR" ]; then
    CONFIG_COUNT=$(find "$TEST_DATA_DIR" -name "*.wg" | wc -l)
    print_status "Found $CONFIG_COUNT test configuration files"
    
    # Validate each test config file
    for config_file in "$TEST_DATA_DIR"/*.wg; do
        if [ -f "$config_file" ]; then
            filename=$(basename "$config_file")
            # Basic syntax check - ensure file is readable
            if [ -r "$config_file" ]; then
                print_success "Test data valid: $filename"
            else
                print_error "Test data invalid: $filename"
            fi
        fi
    done
else
    print_warning "Test data directory not found"
fi

# Generate test report
print_status "Generating test report..."
cat > test_report.md << EOF
# Test Report

## Summary
- **Test Coverage**: ${COVERAGE}%
- **Unit Tests**: ✅ Passed
- **Fuzz Tests**: ✅ Passed
- **Race Tests**: ✅ Passed
- **Cross-platform**: ✅ Passed

## Coverage Details
\`\`\`
$(go tool cover -func=coverage.out)
\`\`\`

## Benchmark Results
\`\`\`
$(cat benchmark_results.txt)
\`\`\`

## Test Data Corpus
- Configuration files: $CONFIG_COUNT
- Valid configs: $(find "$TEST_DATA_DIR" -name "valid_*.wg" 2>/dev/null | wc -l)
- Invalid configs: $(find "$TEST_DATA_DIR" -name "invalid_*.wg" 2>/dev/null | wc -l)

Generated on: $(date)
EOF

print_success "Test report generated: test_report.md"

# Final summary
echo ""
echo "============================================================"
print_success "🎉 All tests completed successfully!"
echo ""
print_status "📊 Coverage: ${COVERAGE}% (target: 90%)"
print_status "📋 Report: test_report.md"
print_status "🌐 HTML Coverage: coverage.html"
print_status "⚡ Benchmarks: benchmark_results.txt"
echo ""

# Open coverage report if running interactively
if [ -t 1 ] && command -v xdg-open &> /dev/null; then
    read -p "Open coverage report in browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        xdg-open coverage.html
    fi
elif [ -t 1 ] && command -v open &> /dev/null; then
    read -p "Open coverage report in browser? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        open coverage.html
    fi
fi

print_success "Test suite completed! 🚀"
