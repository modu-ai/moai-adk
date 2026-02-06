# MoAI-ADK Go Edition Installer for Windows
# Requires PowerShell 5.1 or later
# Supports piped execution: irm https://... | iex

# Error handling
$ErrorActionPreference = "Stop"

function Print-Info {
    Write-Host "[INFO] " -ForegroundColor Cyan -NoNewline
    Write-Host $args
}

function Print-Success {
    Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline
    Write-Host $args
}

function Print-Error {
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $args
}

function Print-Warning {
    Write-Host "[WARNING] " -ForegroundColor Yellow -NoNewline
    Write-Host $args
}

# Detect platform
function Get-Platform {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture

    $platform = switch ($arch) {
        "X64"   { "windows_amd64" }
        "Arm64" { "windows_arm64" }
        default {
            Print-Error "Unsupported architecture: $arch"
            Print-Info "Supported architectures: x64, arm64"
            exit 1
        }
    }

    Print-Success "Detected platform: $platform"
    return $platform
}

# Get latest Go edition version
function Get-LatestVersion {
    $versionUrl = "https://api.github.com/repos/modu-ai/moai-adk/releases"

    try {
        $response = Invoke-RestMethod -Uri $versionUrl -Method Get
        # Find the latest release (accept both v* and go-v* tags)
        $goRelease = $response | Where-Object { $_.tag_name -like "v*" -or $_.tag_name -like "go-v*" } | Select-Object -First 1

        if (-not $goRelease) {
            Print-Error "No releases found"
            Print-Info "You can:"
            Write-Host "  1. Install a specific version: .\install.ps1 -version 2.0.0"
            Write-Host "  2. Install from source: go install github.com/modu-ai/moai-adk/cmd/moai@latest"
            exit 1
        }

        $version = $goRelease.tag_name -replace '^go-v', '' -replace '^v', ''
        Print-Success "Latest Go edition version: $version"
        return $version
    }
    catch {
        Print-Error "Failed to fetch latest Go edition version from GitHub: $_"
        exit 1
    }
}

# Download binary
function Download-Binary {
    param(
        [string]$Version,
        [string]$Platform
    )

    # Extract OS and ARCH from platform (e.g., "windows_amd64")
    $parts = $Platform -split '_'
    $os = $parts[0]
    $arch = $parts[1]

    # Build archive filename matching goreleaser format
    $archiveName = "moai-adk_${Version}_${os}_${arch}.zip"
    $downloadUrl = "https://github.com/modu-ai/moai-adk/releases/download/v$Version/$archiveName"
    $checksumUrl = "https://github.com/modu-ai/moai-adk/releases/download/v$Version/checksums.txt"

    $tempDir = Join-Path $env:TEMP "moai-install-$(New-Guid)"
    $archiveFile = Join-Path $tempDir $archiveName
    $checksumFile = Join-Path $tempDir "checksums.txt"

    # Create temp directory
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    Print-Info "Downloading from: $downloadUrl"

    try {
        # Use TLS 1.2
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

        # Download archive
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archiveFile -UseBasicParsing
        Print-Success "Download completed"

        # Download checksums (optional)
        try {
            Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumFile -UseBasicParsing
            Print-Info "Verifying checksum..."

            $checksumContent = Get-Content $checksumFile
            $expectedLine = $checksumContent | Select-String -Pattern $archiveName | Select-Object -First 1

            if ($expectedLine) {
                $expectedChecksum = ($expectedLine -split '\s+')[0]
                $actualChecksum = (Get-FileHash -Path $archiveFile -Algorithm SHA256).Hash.ToLower()

                if ($expectedChecksum -eq $actualChecksum) {
                    Print-Success "Checksum verified"
                }
                else {
                    Print-Error "Checksum mismatch!"
                    Print-Error "Expected: $expectedChecksum"
                    Print-Error "Actual:   $actualChecksum"
                    Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
                    exit 1
                }
            }
        }
        catch {
            Print-Warning "Failed to verify checksum (continuing anyway)"
        }
    }
    catch {
        Print-Error "Download failed: $_"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Extract archive
    Print-Info "Extracting archive..."
    try {
        Expand-Archive -Path $archiveFile -DestinationPath $tempDir -Force
        Print-Success "Extraction completed"
    }
    catch {
        Print-Error "Failed to extract archive: $_"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Find the binary
    $binaryPath = Join-Path $tempDir "moai.exe"
    if (-not (Test-Path $binaryPath)) {
        Print-Error "Binary not found in archive"
        Remove-Item $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    return $binaryPath
}

# Install binary
function Install-Binary {
    param(
        [string]$BinaryPath
    )

    # Determine install location
    $targetDir = Join-Path $env:LOCALAPPDATA "Programs\moai"
    if ($env:MOAI_INSTALL_DIR) {
        $targetDir = $env:MOAI_INSTALL_DIR
    }

    # Create target directory
    if (-not (Test-Path $targetDir)) {
        Print-Info "Creating directory: $targetDir"
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    $targetPath = Join-Path $targetDir "moai.exe"

    Print-Info "Installing to: $targetPath"

    try {
        Copy-Item -Path $BinaryPath -Destination $targetPath -Force
        Print-Success "Installed to: $targetPath"
    }
    catch {
        Print-Error "Failed to install: $_"
        Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue
        exit 1
    }

    # Clean up temp directory
    Remove-Item (Split-Path $BinaryPath) -Recurse -Force -ErrorAction SilentlyContinue

    return $targetPath
}

# Add to PATH
function Add-ToPath {
    param([string]$TargetPath)

    $targetDir = Split-Path $TargetPath -Parent

    # Check if already in PATH
    $pathEnv = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($pathEnv -like "*$targetDir*") {
        Print-Info "Already in PATH"
        return
    }

    Print-Info "Adding to PATH..."
    try {
        $newPath = "$pathEnv;$targetDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$targetDir"
        Print-Success "Added to PATH"
    }
    catch {
        Print-Warning "Failed to add to PATH automatically"
        Write-Host ""
        Print-Info "Please add manually by running (Admin PowerShell):"
        Write-Host ""
        Write-Host "    [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$targetDir', 'User')" -ForegroundColor Yellow
        Write-Host ""
    }
}

# Verify installation
function Verify-Installation {
    param([string]$TargetPath)

    try {
        $output = & $TargetPath version 2>&1
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
        Print-Info "Run: $TargetPath version"
    }
}

# Main installation flow
function Main {
    param($Arguments)

    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════════════════════╗"
    Write-Host "║          MoAI-ADK Go Edition Installer v2.0                   ║"
    Write-Host "╚══════════════════════════════════════════════════════════════╝"
    Write-Host ""

    # Parse arguments
    $Version = ""
    $InstallDir = ""

    for ($i = 0; $i -lt $Arguments.Count; $i++) {
        switch ($Arguments[$i]) {
            { $_ -in "--version", "-version" } {
                $Version = $Arguments[$i + 1]
                $i++
            }
            { $_ -in "--install-dir", "-install-dir" } {
                $InstallDir = $Arguments[$i + 1]
                $i++
            }
            { $_ -in "-h", "--help", "-help" } {
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
                Write-Host ""
                Write-Host "Piped execution:"
                Write-Host "  irm https://raw.githubusercontent.com/.../install.ps1 | iex"
                exit 0
            }
            default {
                Print-Error "Unknown option: $($Arguments[$i])"
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
    $targetPath = Install-Binary -BinaryPath $binaryPath

    # Add to PATH
    Add-ToPath -TargetPath $targetPath

    # Verify installation
    Verify-Installation -TargetPath $targetPath

    Write-Host ""
    Print-Success "Installation complete!"
    Write-Host ""
    Print-Info "Documentation: https://github.com/modu-ai/moai-adk"
}

# Run main function with script arguments
Main -Arguments $args
