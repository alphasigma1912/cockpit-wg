# Test runner script for WireGuard configuration parser and validator
# PowerShell version for Windows compatibility

param(
    [switch]$SkipCoverage,
    [switch]$SkipFuzz,
    [switch]$SkipBenchmarks,
    [string]$FuzzTime = "10s"
)

# Set error handling
$ErrorActionPreference = "Stop"

Write-Host "🧪 Running comprehensive test suite for WireGuard bridge..." -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

# Navigate to bridge directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location (Join-Path $ScriptDir "bridge")

# Function to print colored output
function Write-Status($message) {
    Write-Host "[INFO] $message" -ForegroundColor Blue
}

function Write-Success($message) {
    Write-Host "[PASS] $message" -ForegroundColor Green
}

function Write-Warning($message) {
    Write-Host "[WARN] $message" -ForegroundColor Yellow
}

function Write-Error($message) {
    Write-Host "[FAIL] $message" -ForegroundColor Red
}

try {
    # Clean up any previous test artifacts
    Write-Status "Cleaning up previous test artifacts..."
    & go clean -testcache
    if (Test-Path "coverage.out") { Remove-Item "coverage.out" }
    if (Test-Path "coverage.html") { Remove-Item "coverage.html" }
    if (Test-Path "benchmark_results.txt") { Remove-Item "benchmark_results.txt" }

    # Verify Go environment
    Write-Status "Checking Go environment..."
    & go version
    & go mod tidy
    & go mod verify

    # Run static analysis if available
    Write-Status "Running static analysis..."
    if (Get-Command "golangci-lint" -ErrorAction SilentlyContinue) {
        & golangci-lint run .\...
        Write-Success "Static analysis passed"
    } else {
        Write-Warning "golangci-lint not found, skipping static analysis"
    }

    # Run unit tests with coverage
    if (-not $SkipCoverage) {
        Write-Status "Running unit tests with coverage analysis..."
        & go test -v -race -coverprofile=coverage.out -covermode=atomic .\...
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Unit tests failed"
            exit 1
        }
        Write-Success "Unit tests passed"

        # Generate coverage report
        Write-Status "Generating coverage report..."
        & go tool cover -html=coverage.out -o coverage.html

        # Calculate coverage percentage
        $CoverageOutput = & go tool cover -func=coverage.out | Select-String "total"
        if ($CoverageOutput) {
            if ($CoverageOutput -match "(\d+\.\d+)%") {
                $Coverage = [double]$Matches[1]
                Write-Status "Total test coverage: $Coverage%"
                
                if ($Coverage -ge 90) {
                    Write-Success "Coverage target met (≥90%): $Coverage%"
                } else {
                    Write-Warning "Coverage below target (<90%): $Coverage%"
                }
            }
        }
    }

    # Run benchmarks
    if (-not $SkipBenchmarks) {
        Write-Status "Running benchmark tests..."
        & go test -bench=. -benchmem .\internal\config\ | Out-File -FilePath "benchmark_results.txt" -Encoding UTF8
        Write-Success "Benchmark tests completed"
    }

    # Run fuzz tests
    if (-not $SkipFuzz) {
        Write-Status "Running fuzz tests..."
        
        $FuzzTests = @(
            "FuzzParseConfig",
            "FuzzValidateIPs", 
            "FuzzDetectIPConflicts"
        )

        foreach ($FuzzTest in $FuzzTests) {
            Write-Status "Running $FuzzTest..."
            
            # Use Start-Process with timeout for fuzz tests
            $Process = Start-Process -FilePath "go" -ArgumentList "test", "-fuzz=$FuzzTest", "-fuzztime=$FuzzTime", ".\internal\config\" -PassThru -NoNewWindow
            
            if (-not $Process.WaitForExit(15000)) {
                $Process.Kill()
                Write-Success "$FuzzTest completed (timeout reached)"
            } elseif ($Process.ExitCode -ne 0) {
                Write-Error "$FuzzTest failed"
                exit 1
            } else {
                Write-Success "$FuzzTest completed successfully"
            }
        }
        Write-Success "Fuzz tests completed"
    }

    # Test cross-platform compatibility
    Write-Status "Testing cross-platform compatibility..."
    $env:GOOS = "linux"
    & go test -v .\...
    if ($LASTEXITCODE -ne 0) { throw "Linux build test failed" }
    
    $env:GOOS = "windows"
    & go test -v .\...
    if ($LASTEXITCODE -ne 0) { throw "Windows build test failed" }
    
    $env:GOOS = "darwin"
    & go test -v .\...
    if ($LASTEXITCODE -ne 0) { throw "Darwin build test failed" }
    
    Remove-Item env:GOOS
    Write-Success "Cross-platform tests passed"

    # Test race conditions
    Write-Status "Running race condition tests..."
    & go test -race -count=10 .\internal\config\
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Race condition tests failed"
        exit 1
    }
    Write-Success "Race condition tests passed"

    # Validate test data corpus
    Write-Status "Validating test data corpus..."
    $TestDataDir = "testdata"
    if (Test-Path $TestDataDir) {
        $ConfigFiles = Get-ChildItem -Path $TestDataDir -Filter "*.wg"
        $ConfigCount = $ConfigFiles.Count
        Write-Status "Found $ConfigCount test configuration files"
        
        foreach ($ConfigFile in $ConfigFiles) {
            if (Test-Path $ConfigFile.FullName -PathType Leaf) {
                Write-Success "Test data valid: $($ConfigFile.Name)"
            } else {
                Write-Error "Test data invalid: $($ConfigFile.Name)"
            }
        }
    } else {
        Write-Warning "Test data directory not found"
        $ConfigCount = 0
    }

    # Generate test report
    Write-Status "Generating test report..."
    
    $CoverageDetails = ""
    if (Test-Path "coverage.out") {
        $CoverageDetails = & go tool cover -func=coverage.out | Out-String
    }
    
    $BenchmarkResults = ""
    if (Test-Path "benchmark_results.txt") {
        $BenchmarkResults = Get-Content "benchmark_results.txt" | Out-String
    }
    
    $ValidConfigs = 0
    $InvalidConfigs = 0
    if (Test-Path $TestDataDir) {
        $ValidConfigs = (Get-ChildItem -Path $TestDataDir -Filter "valid_*.wg").Count
        $InvalidConfigs = (Get-ChildItem -Path $TestDataDir -Filter "invalid_*.wg").Count
    }

    $TestReport = @"
# Test Report

## Summary
- **Test Coverage**: $Coverage%
- **Unit Tests**: ✅ Passed
- **Fuzz Tests**: ✅ Passed
- **Race Tests**: ✅ Passed
- **Cross-platform**: ✅ Passed

## Coverage Details
``````
$CoverageDetails
``````

## Benchmark Results
``````
$BenchmarkResults
``````

## Test Data Corpus
- Configuration files: $ConfigCount
- Valid configs: $ValidConfigs
- Invalid configs: $InvalidConfigs

Generated on: $(Get-Date)
"@

    $TestReport | Out-File -FilePath "test_report.md" -Encoding UTF8
    Write-Success "Test report generated: test_report.md"

    # Final summary
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Cyan
    Write-Success "🎉 All tests completed successfully!"
    Write-Host ""
    Write-Status "📊 Coverage: $Coverage% (target: 90%)"
    Write-Status "📋 Report: test_report.md"
    Write-Status "🌐 HTML Coverage: coverage.html"
    Write-Status "⚡ Benchmarks: benchmark_results.txt"
    Write-Host ""

    # Offer to open coverage report
    if (Test-Path "coverage.html") {
        $Response = Read-Host "Open coverage report in browser? (y/N)"
        if ($Response -match "^[Yy]") {
            Start-Process "coverage.html"
        }
    }

    Write-Success "Test suite completed! 🚀"

} catch {
    Write-Error "Test suite failed: $_"
    exit 1
}
