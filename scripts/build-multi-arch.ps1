# Multi-Architecture Build Script for Windows PowerShell
param(
    [string]$Version = "",
    [string]$BuildDir = "dist",
    [switch]$Clean = $false,
    [switch]$Help = $false
)

if ($Help) {
    Write-Host @"
Multi-Architecture Build Script for Cockpit WireGuard Manager

Usage: .\scripts\build-multi-arch.ps1 [options]

Options:
    -Version <string>   Version string (default: git describe or 'dev')
    -BuildDir <string>  Build output directory (default: 'dist')
    -Clean              Clean build directory before building
    -Help               Show this help message

Examples:
    .\scripts\build-multi-arch.ps1
    .\scripts\build-multi-arch.ps1 -Version "1.0.0" -Clean
    .\scripts\build-multi-arch.ps1 -BuildDir "releases"
"@
    exit 0
}

# Configuration
$Platforms = @(
    @{ OS = "linux"; Arch = "amd64"; Name = "linux-amd64" },
    @{ OS = "linux"; Arch = "arm64"; Name = "linux-arm64" },
    @{ OS = "linux"; Arch = "arm"; ARM = "7"; Name = "linux-armv7" },
    @{ OS = "windows"; Arch = "amd64"; Name = "windows-amd64"; Ext = ".exe" },
    @{ OS = "darwin"; Arch = "amd64"; Name = "darwin-amd64" },
    @{ OS = "darwin"; Arch = "arm64"; Name = "darwin-arm64" }
)

if (-not $Version) {
    try {
        $Version = git describe --tags --always --dirty 2>$null
        if (-not $Version) { $Version = "dev" }
    } catch {
        $Version = "dev"
    }
}

Write-Host "Multi-Architecture Build Script" -ForegroundColor Blue
Write-Host "Version: $Version" -ForegroundColor Blue
Write-Host "Build directory: $BuildDir" -ForegroundColor Blue
Write-Host ""

# Check prerequisites
function Test-Prerequisites {
    Write-Host "Checking prerequisites..." -ForegroundColor Yellow
    
    $missing = @()
    
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        $missing += "Go"
    }
    
    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        $missing += "Node.js"
    }
    
    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        $missing += "npm"
    }
    
    if ($missing.Count -gt 0) {
        Write-Host "Missing prerequisites: $($missing -join ', ')" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "All prerequisites found" -ForegroundColor Green
    Write-Host ""
}

# Build frontend
function Build-Frontend {
    Write-Host "Building frontend..." -ForegroundColor Yellow
    
    Push-Location ui
    try {
        npm install
        if ($LASTEXITCODE -ne 0) { throw "npm install failed" }
        
        npm run build
        if ($LASTEXITCODE -ne 0) { throw "npm run build failed" }
        
        Write-Host "Frontend built successfully" -ForegroundColor Green
    } catch {
        Write-Host "Frontend build failed: $_" -ForegroundColor Red
        exit 1
    } finally {
        Pop-Location
    }
    Write-Host ""
}

# Build backend for specific platform
function Build-Backend {
    param($Platform)
    
    $targetName = $Platform.Name
    Write-Host "Building backend for $targetName..." -ForegroundColor Yellow
    
    $buildPath = Join-Path $BuildDir "cockpit-wg-$targetName"
    New-Item -ItemType Directory -Path $buildPath -Force | Out-Null
    
    $binaryName = "wg-bridge"
    if ($Platform.Ext) { $binaryName += $Platform.Ext }
    
    $env:CGO_ENABLED = "0"
    $env:GOOS = $Platform.OS
    $env:GOARCH = $Platform.Arch
    if ($Platform.ARM) { $env:GOARM = $Platform.ARM }
    
    Push-Location bridge
    try {
        $outputPath = "..\$buildPath\$binaryName"
        & go build -ldflags "-s -w -X main.version=$Version" -o $outputPath .
        
        if ($LASTEXITCODE -ne 0) {
            throw "Go build failed for $targetName"
        }
        
        # Verify binary exists
        if (-not (Test-Path $outputPath)) {
            throw "Binary not created for $targetName"
        }
        
        Write-Host "Backend built for $targetName" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "Failed to build $targetName : $_" -ForegroundColor Red
        return $false
    } finally {
        Pop-Location
        # Clean up environment variables
        Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
        Remove-Item Env:GOOS -ErrorAction SilentlyContinue
        Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
        Remove-Item Env:GOARM -ErrorAction SilentlyContinue
    }
}

# Package complete plugin
function New-PluginPackage {
    param($Platform)
    
    $targetName = $Platform.Name
    $packageDir = Join-Path $BuildDir "cockpit-wg-$targetName"
    
    Write-Host "Packaging plugin for $targetName..." -ForegroundColor Yellow
    
    try {
        # Copy frontend files
        Copy-Item "ui\manifest.json" $packageDir
        Copy-Item "ui\dist\index.html" $packageDir
        Copy-Item "ui\dist\assets" $packageDir -Recurse
        
        # Create zip archive (Windows equivalent of tar.gz)
        $archivePath = Join-Path $BuildDir "cockpit-wg-$targetName.zip"
        if (Test-Path $archivePath) { Remove-Item $archivePath }
        
        # Use .NET compression
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::CreateFromDirectory($packageDir, $archivePath)
        
        $size = [math]::Round((Get-Item $archivePath).Length / 1MB, 2)
        Write-Host "Package created: cockpit-wg-$targetName.zip ($size MB)" -ForegroundColor Green
    } catch {
        Write-Host "Failed to package $targetName : $_" -ForegroundColor Red
    }
}

# Generate checksums
function New-Checksums {
    Write-Host "Generating checksums..." -ForegroundColor Yellow
    
    $packages = Get-ChildItem -Path $BuildDir -Filter "*.zip"
    if ($packages.Count -eq 0) {
        Write-Host "No packages found for checksum generation" -ForegroundColor Yellow
        return
    }
    
    $checksumPath = Join-Path $BuildDir "checksums.txt"
    $checksums = @()
    
    foreach ($package in $packages) {
        $hash = Get-FileHash $package.FullName -Algorithm SHA256
        $checksums += "$($hash.Hash.ToLower())  $($package.Name)"
    }
    
    $checksums | Out-File -FilePath $checksumPath -Encoding UTF8
    Write-Host "Checksums generated" -ForegroundColor Green
}

# Main build process
function Main {
    Test-Prerequisites
    
    # Clean previous builds
    if ($Clean -or (Test-Path $BuildDir)) {
        Write-Host "Cleaning previous builds..." -ForegroundColor Yellow
        if (Test-Path $BuildDir) { Remove-Item $BuildDir -Recurse -Force }
    }
    New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    
    Build-Frontend
    
    $successfulBuilds = @()
    $failedBuilds = @()
    
    foreach ($platform in $Platforms) {
        if (Build-Backend $platform) {
            New-PluginPackage $platform
            $successfulBuilds += $platform.Name
        } else {
            $failedBuilds += $platform.Name
        }
        Write-Host ""
    }
    
    New-Checksums
    
    Write-Host "Build Summary" -ForegroundColor Blue
    Write-Host "Successful builds ($($successfulBuilds.Count)):" -ForegroundColor Green
    foreach ($build in $successfulBuilds) {
        Write-Host "   $build"
    }
    
    if ($failedBuilds.Count -gt 0) {
        Write-Host "Failed builds ($($failedBuilds.Count)):" -ForegroundColor Red
        foreach ($build in $failedBuilds) {
            Write-Host "   $build"
        }
    }
    
    Write-Host ""
    Write-Host "Build artifacts in: $BuildDir" -ForegroundColor Blue
    Get-ChildItem -Path $BuildDir -Filter "*.zip" | ForEach-Object {
        $size = [math]::Round($_.Length / 1MB, 2)
        Write-Host "   $($_.Name) ($size MB)"
    }
}

# Run main function
Main
