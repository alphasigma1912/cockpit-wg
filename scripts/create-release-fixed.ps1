# GitHub Release Creation Script for Windows
# Run this script to create a GitHub release with all build artifacts

param(
    [string]$Version = "v1.0.0",
    [string]$Tag = "v1.0.0"
)

$ReleaseTitle = "Cockpit WireGuard Manager v1.0.0 - Multi-Platform Release"
$DistDir = "dist"

Write-Host "🚀 Creating GitHub Release: $Version" -ForegroundColor Blue
Write-Host "📁 Using artifacts from: $DistDir" -ForegroundColor Blue

# Check if we're in a git repository
try {
    git rev-parse --git-dir | Out-Null
} catch {
    Write-Host "❌ Error: Not in a git repository" -ForegroundColor Red
    exit 1
}

# Check if dist directory exists
if (-not (Test-Path $DistDir)) {
    Write-Host "❌ Error: Build directory $DistDir not found" -ForegroundColor Red
    Write-Host "Run the build script first: .\scripts\build-multi-arch.ps1" -ForegroundColor Yellow
    exit 1
}

# Create and push tag
Write-Host "🏷️  Creating tag: $Tag" -ForegroundColor Yellow
try {
    git tag -a $Tag -m $ReleaseTitle
    git push origin $Tag
    Write-Host "✅ Tag created and pushed successfully" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Tag may already exist or push failed" -ForegroundColor Yellow
    Write-Host "Error: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "📋 Manual Release Steps:" -ForegroundColor Cyan
Write-Host "1. Go to: https://github.com/alphasigma1912/cockpit-wg/releases/new" -ForegroundColor White
Write-Host "2. Select tag: $Tag" -ForegroundColor White
Write-Host "3. Set title: $ReleaseTitle" -ForegroundColor White
Write-Host "4. Copy the release body from RELEASE_NOTES.md" -ForegroundColor White
Write-Host "5. Upload these files from $DistDir/:" -ForegroundColor White
Write-Host "   - cockpit-wg-linux-arm64.zip (Ubuntu ARM64 - PRIMARY TARGET)" -ForegroundColor Green
Write-Host "   - cockpit-wg-linux-amd64.zip" -ForegroundColor White
Write-Host "   - cockpit-wg-linux-armv7.zip" -ForegroundColor White
Write-Host "   - cockpit-wg-windows-amd64.zip" -ForegroundColor White
Write-Host "   - cockpit-wg-darwin-amd64.zip" -ForegroundColor White
Write-Host "   - cockpit-wg-darwin-arm64.zip" -ForegroundColor White
Write-Host "   - checksums.txt" -ForegroundColor White
Write-Host ""
Write-Host "🎯 Ubuntu ARM64 package ready: $DistDir\cockpit-wg-linux-arm64.zip" -ForegroundColor Green

# Show file sizes
Write-Host ""
Write-Host "📊 Package Sizes:" -ForegroundColor Cyan
Get-ChildItem "$DistDir\*.zip" | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    if ($_.Name -like "*linux-arm64*") {
        Write-Host "   $($_.Name): $size MB" -ForegroundColor Green
    } else {
        Write-Host "   $($_.Name): $size MB" -ForegroundColor White
    }
}
