@echo off
REM Build script for deej-winapp on Windows
REM Requires: Go 1.14+, gcc (mingw-w64)

echo === deej-winapp Windows Build ===

REM 1. Install rsrc for Windows resource embedding
echo [1/4] Installing rsrc...
go install github.com/akavel/rsrc@latest

REM 2. Generate .syso resource file in the cmd directory
echo [2/4] Generating Windows resources...
rsrc -manifest pkg\deej\assets\deej.manifest -ico pkg\deej\assets\logo.ico -arch amd64 -o pkg\deej\cmd\rsrc_windows_amd64.syso
if %ERRORLEVEL% neq 0 (
    echo WARNING: rsrc generation failed. Trying without icon...
    rsrc -manifest pkg\deej\assets\deej.manifest -arch amd64 -o pkg\deej\cmd\rsrc_windows_amd64.syso
)

REM 3. Clean previous build
echo [3/4] Cleaning...
go clean -cache

REM 4. Build
echo [4/4] Building deej.exe...
go build -ldflags="-H windowsgui" -o deej.exe .\pkg\deej\cmd\

if %ERRORLEVEL% equ 0 (
    echo.
    echo === Build successful! ===
    echo Output: deej.exe
    echo.
    echo Remember to place deej.exe and config.yaml in the same directory.
) else (
    echo.
    echo === Build failed ===
    echo Check the error messages above.
)

pause
