@echo off
REM MoAI-ADK Go Edition Installer for Windows CMD
REM Requires Windows 7 or later

setlocal enabledelayedexpansion

REM Parse arguments
set "VERSION="
set "INSTALL_DIR="
set "SHOW_HELP=0"

:parse_args
if "%~1"=="" goto done_parsing
if "%~1"=="--version" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_args
)
if "%~1"=="--install-dir" (
    set "INSTALL_DIR=%~2"
    shift
    shift
    goto parse_args
)
if "%~1"=="-h" goto show_help
if "%~1"=="--help" goto show_help
if "%~1"=="/?" goto show_help
echo [ERROR] Unknown option: %~1
echo Use --help for usage information
exit /b 1

:done_parsing

REM Main installation
echo.
echo ╔══════════════════════════════════════════════════════════════╗
echo ║          MoAI-ADK Go Edition Installer v2.0                   ║
echo ╚══════════════════════════════════════════════════════════════╝
echo.

REM Detect platform (OS and architecture)
set "OS=windows"
set "ARCH=amd64"

REM Detect ARM64 using PROCESSOR_ARCHITECTURE
if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64" set "ARCH=arm64"
if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64" set "ARCH=arm64"

set "PLATFORM=%OS%_%ARCH%"
echo [INFO] Detected platform: %PLATFORM%

REM Get version
if "%VERSION%"=="" (
    echo [INFO] Fetching latest Go edition version from GitHub...

    REM Use PowerShell to get latest version (accept both v* and go-v* tags)
    for /f "tokens=*" %%i in ('powershell -Command "$releases = Invoke-RestMethod -Uri https://api.github.com/repos/modu-ai/moai-adk/releases; $goRelease = $releases ^| Where-Object { $_.tag_name -like 'v*' -or $_.tag_name -like 'go-v*' } ^| Select-Object -First 1; if ($goRelease) { $goRelease.tag_name -replace '^go-v', '' -replace '^v', '' } else { '' }" 2^>nul') do (
        set "VERSION=%%i"
    )

    if "!VERSION!"=="" (
        echo [ERROR] Failed to fetch latest version
        echo [INFO] No releases found. You can:
        echo   1. Install a specific version: install.bat --version 2.0.0
        echo   2. Install from source: go install github.com/modu-ai/moai-adk/cmd/moai@latest
        exit /b 1
    )
)
echo [SUCCESS] Latest Go edition version: !VERSION!

REM Create temp directory
set "TEMP_DIR=%TEMP%\moai-install-%RANDOM%"
if not exist "%TEMP_DIR%" mkdir "%TEMP_DIR%"

REM Build archive filename matching goreleaser format
set "ARCHIVE_NAME=moai-adk_!VERSION!_%OS%_%ARCH%.zip"
set "DOWNLOAD_URL=https://github.com/modu-ai/moai-adk/releases/download/v!VERSION!/!ARCHIVE_NAME!"
set "ARCHIVE_FILE=%TEMP_DIR%\!ARCHIVE_NAME!"
set "CHECKSUM_URL=https://github.com/modu-ai/moai-adk/releases/download/v!VERSION!/checksums.txt"
set "CHECKSUM_FILE=%TEMP_DIR%\checksums.txt"

echo [INFO] Downloading from: !DOWNLOAD_URL!

REM Download using PowerShell
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '!DOWNLOAD_URL!' -OutFile '!ARCHIVE_FILE!' -UseBasicParsing" >nul 2>&1

if not exist "!ARCHIVE_FILE!" (
    echo [ERROR] Download failed
    rmdir /s /q "%TEMP_DIR%" 2>nul
    exit /b 1
)
echo [SUCCESS] Download completed

REM Download and verify checksums (optional)
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; try { Invoke-WebRequest -Uri '!CHECKSUM_URL!' -OutFile '!CHECKSUM_FILE!' -UseBasicParsing } catch { }" >nul 2>&1

if exist "!CHECKSUM_FILE!" (
    echo [INFO] Verifying checksum...
    REM PowerShell checksum verification
    powershell -Command "$checksums = Get-Content '!CHECKSUM_FILE!'; $line = $checksums | Select-String -Pattern '!ARCHIVE_NAME!' | Select-Object -First 1; if ($line) { $expected = ($line -split '\s+')[0]; $actual = (Get-FileHash -Path '!ARCHIVE_FILE!' -Algorithm SHA256).Hash.ToLower(); if ($expected -eq $actual) { exit 0 } else { Write-Host '[ERROR] Checksum mismatch!'; Write-Host '[ERROR] Expected: ' + $expected; Write-Host '[ERROR] Actual: ' + $actual; exit 1 } }" >nul 2>&1
    if errorlevel 1 (
        echo [ERROR] Checksum verification failed
        rmdir /s /q "%TEMP_DIR%" 2>nul
        exit /b 1
    )
    echo [SUCCESS] Checksum verified
) else (
    echo [WARNING] Checksum verification skipped
)

REM Extract archive
echo [INFO] Extracting archive...
powershell -Command "Expand-Archive -Path '!ARCHIVE_FILE!' -DestinationPath '%TEMP_DIR%' -Force" >nul 2>&1

if not exist "%TEMP_DIR%\moai.exe" (
    echo [ERROR] Failed to extract archive or binary not found
    rmdir /s /q "%TEMP_DIR%" 2>nul
    exit /b 1
)
echo [SUCCESS] Extraction completed

REM Determine install location
if "%INSTALL_DIR%"=="" (
    set "INSTALL_DIR=%LOCALAPPDATA%\Programs\moai"
)

REM Create install directory if it doesn't exist
if not exist "%INSTALL_DIR%" (
    echo [INFO] Creating directory: %INSTALL_DIR%
    mkdir "%INSTALL_DIR%"
)

REM Install
set "TARGET_PATH=%INSTALL_DIR%\moai.exe"

echo [INFO] Installing to: %TARGET_PATH%

copy /Y "%TEMP_DIR%\moai.exe" "%TARGET_PATH%" >nul
if errorlevel 1 (
    echo [ERROR] Failed to install
    rmdir /s /q "%TEMP_DIR%" 2>nul
    exit /b 1
)
echo [SUCCESS] Installed to: %TARGET_PATH%

REM Clean up
rmdir /s /q "%TEMP_DIR%" 2>nul

REM Add to PATH
echo.
echo [INFO] Adding to PATH...
powershell -Command "$currentPath = [Environment]::GetEnvironmentVariable('Path', 'User'); if ($currentPath -notlike '*%INSTALL_DIR%*') { [Environment]::SetEnvironmentVariable('Path', $currentPath + ';%INSTALL_DIR%', 'User'); Write-Host '[SUCCESS] Added to PATH' } else { Write-Host '[INFO] Already in PATH' }"

REM Verify installation
echo.
echo [INFO] Verifying installation...
"%TARGET_PATH%" version
echo.
echo [SUCCESS] Installation complete!
echo.
echo To get started, run:
echo     moai init          # Initialize a new project
echo     moai doctor        # Check system health
echo     moai update --project # Update project templates
echo.
echo Documentation: https://github.com/modu-ai/moai-adk
goto :eof

:show_help
echo Usage: install.bat [OPTIONS]
echo.
echo Options:
echo   --version VERSION    Install specific version (default: latest)
echo   --install-dir DIR     Install to custom directory
echo   -h, --help            Show this help message
echo.
echo Examples:
echo   install.bat                              # Install latest version
echo   install.bat --version 2.0.0              # Install version 2.0.0
echo   install.bat --install-dir "C:\Tools"      # Install to custom directory
exit /b 0
