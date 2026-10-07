# Offline Main-flow test: load only the installer functions, replacing network,
# binary placement and user PATH writes with in-memory stubs. No profile is used.
param([string]$Installer = (Join-Path $PSScriptRoot "../install.ps1"))
$ErrorActionPreference = "Stop"
$tokens = $null
$errors = $null
$ast = [System.Management.Automation.Language.Parser]::ParseFile(
    (Resolve-Path $Installer).Path, [ref]$tokens, [ref]$errors)
if ($errors.Count -ne 0) { throw "Installer parse failed: $errors" }
foreach ($name in @("Main", "Verify-Installation")) {
    $fn = $ast.Find({ param($node)
        $node -is [System.Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -eq $name
    }, $true)
    if (-not $fn) { throw "Installer function missing: $name" }
    Invoke-Expression $fn.Extent.Text
}
function Get-Platform { return "offline-test" }
function Download-Binary { return "scratch-binary" }
function Install-Binary { return "Test-InstalledBinary" }
function Add-ToPath { $script:PathStubCalls++ }
function Print-Info {}
function Print-Warning {}
function Print-Success {
    param([string]$Message)
    if ($Message -eq "Installation complete!") { $script:Completed = $true }
}
function Test-InstalledBinary {
    $script:BinaryCalls.Add(($args -join " "))
    if ($script:FailVersion) { throw "optional version probe unavailable" }
    $global:LASTEXITCODE = 0
    return "MoAI offline fixture"
}
foreach ($failVersion in @($false, $true)) {
    $script:FailVersion = $failVersion
    $script:BinaryCalls = [System.Collections.Generic.List[string]]::new()
    $script:PathStubCalls = 0
    $script:Completed = $false
    Main -Arguments @("--version", "9.9.9")
    if (-not $script:Completed -or $script:PathStubCalls -ne 1) {
        throw "Installer did not complete through isolated PATH stub"
    }
    if ($script:BinaryCalls.Count -ne 1 -or $script:BinaryCalls[0] -ne "version") {
        throw "Installer invoked a post-install command: $($script:BinaryCalls -join ', ')"
    }
    Write-Host "PASS binary-only Main (optional version failure=$failVersion)"
}
