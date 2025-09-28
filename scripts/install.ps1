# MoAI-ADK Windows Installation Script
# @TASK:WINDOWS-INSTALL-001
#
# This script automatically installs uv and MoAI-ADK on Windows systems.
# Usage: iwr https://raw.githubusercontent.com/modu-ai/moai-adk/main/scripts/install.ps1 -UseBasicParsing | iex

param(
    [switch]$SkipUvInstall,
    [string]$Version = "latest"
)

# Set error action preference
$ErrorActionPreference = "Stop"

# Colors for output
$Green = "`e[32m"
$Yellow = "`e[33m"
$Red = "`e[31m"
$Blue = "`e[34m"
$Reset = "`e[0m"

function Write-ColoredOutput {
    param(
        [string]$Message,
        [string]$Color = $Reset
    )
    Write-Host "${Color}${Message}${Reset}"
}

function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

Write-ColoredOutput "🗿 MoAI-ADK Windows Installation Script" $Blue
Write-ColoredOutput "=====================================" $Blue

# Check if running in elevated mode (optional, but recommended)
$isElevated = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isElevated) {
    Write-ColoredOutput "⚠️  Running without administrator privileges. This is usually fine, but some operations might fail." $Yellow
}

# Step 1: Install uv if not present
if (-not $SkipUvInstall) {
    if (Test-Command "uv") {
        Write-ColoredOutput "✅ uv is already installed" $Green
    }
    else {
        Write-ColoredOutput "📦 Installing uv (Astral Python package manager)..." $Blue
        try {
            Invoke-RestMethod https://astral.sh/uv/install.ps1 | Invoke-Expression
            Write-ColoredOutput "✅ uv installed successfully" $Green
        }
        catch {
            Write-ColoredOutput "❌ Failed to install uv: $_" $Red
            Write-ColoredOutput "Please install uv manually from https://astral.sh/uv/" $Yellow
            exit 1
        }
    }
}

# Step 2: Install/run MoAI-ADK
Write-ColoredOutput "🗿 Installing MoAI-ADK..." $Blue

try {
    if ($Version -eq "latest") {
        # Install latest version from PyPI
        & uv tool install moai-adk
        Write-ColoredOutput "✅ MoAI-ADK installed successfully" $Green
    }
    else {
        # Install specific version
        & uv tool install "moai-adk==$Version"
        Write-ColoredOutput "✅ MoAI-ADK $Version installed successfully" $Green
    }
}
catch {
    Write-ColoredOutput "❌ Failed to install MoAI-ADK: $_" $Red
    Write-ColoredOutput "Trying alternative installation with uvx..." $Yellow

    try {
        # Fallback to uvx (run without installing)
        & uvx --from moai-adk moai-adk doctor
        Write-ColoredOutput "✅ MoAI-ADK is working via uvx" $Green
        Write-ColoredOutput "💡 You can use 'uvx --from moai-adk moai-adk [command]' to run commands" $Blue
    }
    catch {
        Write-ColoredOutput "❌ Both installation methods failed: $_" $Red
        exit 1
    }
}

# Step 3: Verify installation
Write-ColoredOutput "🔍 Verifying installation..." $Blue

try {
    # Try uv tool run first
    if (Test-Command "moai-adk") {
        $version = & moai-adk --version
        Write-ColoredOutput "✅ MoAI-ADK is ready: $version" $Green
    }
    else {
        # Fallback to uvx
        $version = & uvx --from moai-adk moai-adk --version
        Write-ColoredOutput "✅ MoAI-ADK is ready via uvx: $version" $Green
    }
}
catch {
    Write-ColoredOutput "⚠️  Installation completed but verification failed: $_" $Yellow
    Write-ColoredOutput "Please try running 'moai-adk --version' or 'uvx --from moai-adk moai-adk --version'" $Blue
}

# Step 4: Show next steps
Write-ColoredOutput "" ""
Write-ColoredOutput "🎉 Installation completed successfully!" $Green
Write-ColoredOutput "" ""
Write-ColoredOutput "Next steps:" $Blue
Write-ColoredOutput "1. Create a new project: moai-adk init my-project" $Blue
Write-ColoredOutput "2. Or navigate to existing project and run: moai-adk init" $Blue
Write-ColoredOutput "3. For help: moai-adk --help" $Blue
Write-ColoredOutput "" ""

if (-not (Test-Command "moai-adk")) {
    Write-ColoredOutput "💡 If 'moai-adk' command is not found, use:" $Yellow
    Write-ColoredOutput "   uvx --from moai-adk moai-adk [command]" $Yellow
    Write-ColoredOutput "" ""
}

Write-ColoredOutput "🗿 Welcome to Spec-First TDD development with MoAI-ADK!" $Green