# Build script for deej-winapp on Windows (PowerShell)
# Requires: Go 1.14+, gcc (mingw-w64)

Write-Host "=== deej-winapp Windows Build ===" -ForegroundColor Cyan

# Step 1: Install rsrc for Windows resource embedding
Write-Host "[1/4] Installing rsrc..." -ForegroundColor Yellow
go install github.com/akavel/rsrc@latest
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to install rsrc" -ForegroundColor Red
    exit 1
}

# Step 2: Generate .syso resource file
Write-Host "[2/4] Generating Windows resources..." -ForegroundColor Yellow
$rsrcPath = Join-Path (go env GOPATH) "bin" "rsrc.exe"
& $rsrcPath -manifest "pkg\deej\assets\deej.manifest" -ico "pkg\deej\assets\logo.ico" -arch amd64 -o "pkg\deej\cmd\rsrc_windows_amd64.syso"
if ($LASTEXITCODE -ne 0) {
    Write-Host "WARNING: rsrc with icon failed. Trying without icon..." -ForegroundColor Yellow
    & $rsrcPath -manifest "pkg\deej\assets\deej.manifest" -arch amd64 -o "pkg\deej\cmd\rsrc_windows_amd64.syso"
}

# Step 3: Clean and build
Write-Host "[3/4] Building deej.exe..." -ForegroundColor Yellow
go build -ldflags="-H windowsgui" -o deej.exe .\pkg\deej\cmd\

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n=== Build successful! ===" -ForegroundColor Green
    Write-Host "Output: deej.exe" -ForegroundColor Green
    Write-Host "`nRemember to place deej.exe and config.yaml in the same directory." -ForegroundColor Gray
} else {
    Write-Host "`n=== Build failed ===" -ForegroundColor Red
    Write-Host "Check the error messages above." -ForegroundColor Red
    exit 1
}
