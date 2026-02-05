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

REM Set platform
set PLATFORM=windows_amd64
echo [INFO] Detected platform: %PLATFORM%

REM Get version
if "%VERSION%"=="" (
    echo [INFO] Fetching latest version from GitHub...

    REM Use PowerShell to get latest version
    for /f "tokens=*" %%i in ('powershell -Command "Invoke-RestMethod -Uri https://api.github.com/repos/modu-ai/moai-adk/releases/latest | Select-Object -ExpandProperty tag_name" 2^>nul') do (
        set "VERSION=%%i"
    )

    REM Remove 'v' prefix if present
    set "VERSION=!VERSION:v=!"

    if "!VERSION!"=="" (
        echo [ERROR] Failed to fetch latest version
        exit /b 1
    )
)
echo [SUCCESS] Latest version: !VERSION!

REM Create temp directory
set "TEMP_DIR=%TEMP%\moai-install"
if not exist "%TEMP_DIR%" mkdir "%TEMP_DIR%"

REM Download URL
set "DOWNLOAD_URL=https://github.com/modu-ai/moai-adk/releases/download/v!VERSION!/moai-%PLATFORM%.exe"
set "DOWNLOAD_FILE=%TEMP_DIR%\moai.exe"

echo [INFO] Downloading from: !DOWNLOAD_URL!

REM Download using PowerShell
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri '!DOWNLOAD_URL!' -OutFile '!DOWNLOAD_FILE!' -UseBasicParsing" >nul 2>&1

if not exist "%DOWNLOAD_FILE%" (
    echo [ERROR] Download failed
    rmdir "%TEMP_DIR%" 2>nul
    exit /b 1
)
echo [SUCCESS] Download completed

REM Determine install location
if "%INSTALL_DIR%"=="" (
    set "INSTALL_DIR=%USERPROFILE%"
)

REM Install
set "TARGET_PATH=%INSTALL_DIR%\moai.exe"

echo [INFO] Installing to: %TARGET_PATH%

move /Y "%DOWNLOAD_FILE%" "%TARGET_PATH%" >nul
if errorlevel 1 (
    echo [ERROR] Failed to install
    rmdir "%TEMP_DIR%" 2>nul
    exit /b 1
)
echo [SUCCESS] Installed to: %TARGET_PATH%

REM Clean up
rmdir "%TEMP_DIR%" 2>nul

REM Add to PATH warning
echo.
echo [WARNING] Please add the following to your PATH:
echo.
echo     set PATH=%%PATH%%;%INSTALL_DIR%
echo.
echo Or use System Properties ^> Environment Variables ^> Path
echo.

REM Verify installation
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
