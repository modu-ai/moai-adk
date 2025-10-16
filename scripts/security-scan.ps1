# @CODE:SECURITY-001 | Windows PowerShell security scan script
# MoAI-ADK Security Scan (PowerShell)

Write-Host ""
Write-Host "🔍 MoAI-ADK Security Scan" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan
Write-Host ""

# Function to check if a Python module is installed
function Test-PythonModule {
    param($ModuleName)

    try {
        & python -m $ModuleName --version 2>&1 | Out-Null
        return $?
    }
    catch {
        return $false
    }
}

# Function to install a Python module
function Install-PythonModule {
    param($ModuleName)

    Write-Host "Installing $ModuleName..." -ForegroundColor Yellow
    try {
        & python -m pip install $ModuleName
        if ($LASTEXITCODE -eq 0) {
            return $true
        }
    }
    catch {
        Write-Host "❌ Failed to install $ModuleName" -ForegroundColor Red
        return $false
    }
    return $false
}

# Check security tools
Write-Host "📦 Checking security tools..." -ForegroundColor Cyan

$tools = @("pip_audit", "bandit")
foreach ($tool in $tools) {
    if (-not (Test-PythonModule $tool)) {
        if (-not (Install-PythonModule $tool)) {
            exit 1
        }
    }
}

Write-Host ""

# Step 1: pip-audit
Write-Host "🔍 Step 1: Running pip-audit (dependency vulnerability scan)..." -ForegroundColor Cyan
Write-Host "-------------------------------------------------------------------"

$pipAuditFailed = $false
try {
    & python -m pip_audit
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ No vulnerabilities found" -ForegroundColor Green
    }
    else {
        Write-Host "⚠️ Vulnerabilities detected. Please review above." -ForegroundColor Yellow
        $pipAuditFailed = $true
    }
}
catch {
    Write-Host "⚠️ Vulnerabilities detected. Please review above." -ForegroundColor Yellow
    $pipAuditFailed = $true
}

Write-Host ""

# Step 2: bandit
Write-Host "🔍 Step 2: Running bandit (code security scan)..." -ForegroundColor Cyan
Write-Host "-------------------------------------------------------------------"

# Find src directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$srcDir = Join-Path $projectRoot "src"

if (-not (Test-Path $srcDir)) {
    Write-Host "❌ Source directory not found: $srcDir" -ForegroundColor Red
    exit 1
}

$banditFailed = $false
try {
    & python -m bandit -r $srcDir -ll
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ No high/medium security issues found" -ForegroundColor Green
    }
    else {
        Write-Host "❌ Security issues detected. Please review above." -ForegroundColor Red
        $banditFailed = $true
    }
}
catch {
    Write-Host "❌ Security issues detected. Please review above." -ForegroundColor Red
    $banditFailed = $true
}

Write-Host ""
Write-Host "==========================" -ForegroundColor Cyan

# Summary
if ($pipAuditFailed -or $banditFailed) {
    Write-Host "⚠️ Security scan completed with warnings/errors" -ForegroundColor Yellow
    Write-Host "   Please review the issues above and fix them." -ForegroundColor Yellow
    exit 1
}
else {
    Write-Host "✅ All security scans passed!" -ForegroundColor Green
    exit 0
}
