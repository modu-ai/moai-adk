# MoAI-ADK Go Edition Installer for Windows
# Requires PowerShell 5.1 or later

# Error handling
$ErrorActionPreference = "Stop"

# Colors
function Write-ColorOutput($ForegroundColor) {
    process {
        Write-Host $ForegroundColor -NoNewline
        $_
    }
}

function Print-Info {
    Write-Host "ℹ️  [INFO] " -ForegroundColor Cyan -NoNewline
    Write-Host $args
}

function Print-Success {
    Write-Host "✅ [SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $args
}

function Print-Error {
    Write-Host "❌ [ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $args
}

function Print-Warning {
    Write-Host "⚠️  [WARNING] " -ForegroundColor Yellow -NoNewline
    Write-Host $args
}

# Detect platform
function Get-Platform {
    $os = [System.Environment]::OSVersion.Platform
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture

    $platform = "windows_amd64"

    Print-Success "Detected platform: $platform"
    return $platform
}

# Get latest Go edition version
function Get-LatestVersion {
    $versionUrl = "https://api.github.com/repos/modu-ai/moai-adk/releases"

    try {
        $response = Invoke-RestMethod -Uri $versionUrl -Method Get
        # Find the latest Go edition release (tag starts with "go-v")
        $goRelease = $response | Where-Object { $_.tag_name -like "go-v*" } | Select-Object -First 1
        $version = $goRelease.tag_name -replace '^go-v', ''
        Print-Success "Latest Go edition version: $version"
        return $version
    }
    catch {
        Print-Error "Failed to fetch latest Go edition version from GitHub"
        exit 1
    }
}

# Download binary
function Download-Binary {
    param(
        [string]$Version,
        [string]$Platform
    )

    $downloadUrl = "https://github.com/modu-ai/moai-adk/releases/download/go-v$Version/moai-$Platform.exe"
    $tempDir = Join-Path $env:TEMP "moai-install"
    $downloadFile = Join-Path $tempDir "moai.exe"

    # Create temp directory
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    Print-Info "Downloading from: $downloadUrl"

    try {
        # Use TLS 1.2
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadFile -UseBasicParsing
        Print-Success "Download completed"
    }
    catch {
        Print-Error "Download failed: $_"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    return $downloadFile
}

# Install binary
function Install-Binary {
    param(
        [string]$BinaryPath
    )

    # Determine install location
    $targetDir = $env:USERPROFILE
    if ($env:MOAI_INSTALL_DIR) {
        $targetDir = $env:MOAI_INSTALL_DIR
    }

    # Create target directory
    $targetPath = Join-Path $targetDir "moai.exe"

    Print-Info "Installing to: $targetPath"

    try {
        Move-Item -Path $BinaryPath -Destination $targetPath -Force
        Print-Success "Installed to: $targetPath"
    }
    catch {
        Print-Error "Failed to install: $_"
        Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Clean up temp directory
    Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue
}

# Add to PATH
function Add-ToPath {
    $targetDir = Split-Path (Resolve-Path "$env:USERPROFILE\moai.exe") -Parent

    # Check if already in PATH
    $pathEnv = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($pathEnv -like "*$targetDir*") {
        Print-Info "Already in PATH"
        return
    }

    Print-Warning "Please add the following to your PATH:"
    Write-Host ""
    Write-Host "    `$env:PATH += `";$targetDir`"" -ForegroundColor Yellow
    Write-Host ""
    Print-Info "Or run in PowerShell (Admin):"
    Write-Host ""
    Write-Host "    `[Environment]::SetEnvironmentVariable(`"Path`", `$env:PATH + `";$targetDir`", `"User`")`" -ForegroundColor Yellow
    Write-Host ""
}

# Verify installation
function Verify-Installation {
    try {
        $output = & "$env:USERPROFILE\moai.exe" version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Print-Success "MoAI-ADK installed successfully!"
            Write-Host ""
            Write-Host $output
            Write-Host ""
            Print-Info "To get started, run:"
            Write-Host "    moai init          # Initialize a new project" -ForegroundColor Cyan
            Write-Host "    moai doctor        # Check system health" -ForegroundColor Cyan
            Write-Host "    moai update --project # Update project templates" -ForegroundColor Cyan
        }
    }
    catch {
        Print-Warning "Installation completed, verify manually"
        Print-Info "Run: $env:USERPROFILE\moai.exe version"
    }
}

# Main installation flow
function Main {
    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════════════════════╗"
    Write-Host "║          MoAI-ADK Go Edition Installer v2.0                   ║"
    Write-Host "╚══════════════════════════════════════════════════════════════╝"
    Write-Host ""

    # Parse arguments
    $Version = ""
    $InstallDir = ""

    for ($i = 0; $i -lt $args.Count; $i++) {
        switch ($args[$i]) {
            "--version" {
                $Version = $args[$i + 1]
                $i++
            }
            "--install-dir" {
                $InstallDir = $args[$i + 1]
                $i++
            }
            "-h" {
                Write-Host "Usage: .\install.ps1 [OPTIONS]"
                Write-Host ""
                Write-Host "Options:"
                Write-Host "  --version VERSION    Install specific version (default: latest)"
                Write-Host "  --install-dir DIR     Install to custom directory"
                Write-Host "  -h, --help            Show this help message"
                Write-Host ""
                Write-Host "Examples:"
                Write-Host "  .\install.ps1                           # Install latest version"
                Write-Host "  .\install.ps1 -version 2.0.0              # Install version 2.0.0"
                Write-Host "  .\install.ps1 -install-dir `"C:\Tools`"  # Install to custom directory"
                exit 0
            }
            default {
                Print-Error "Unknown option: $($args[$i])"
                Write-Host "Use -h for usage information"
                exit 1
            }
        }
    }

    # Detect platform
    $platform = Get-Platform

    # Get version
    if (-not $Version) {
        $Version = Get-LatestVersion
    }
    else {
        Print-Info "Installing version: $Version"
    }

    # Set install directory if specified
    if ($InstallDir) {
        $env:MOAI_INSTALL_DIR = $InstallDir
    }

    # Download and install
    $binaryPath = Download-Binary -Version $Version -Platform $platform
    Install-Binary -BinaryPath $binaryPath

    # Add to PATH
    Add-ToPath

    # Verify installation
    Verify-Installation

    Write-Host ""
    Print-Success "Installation complete!"
    Write-Host ""
    Print-Info "Documentation: https://github.com/modu-ai/moai-adk"
}

# Run main function
Main $args
