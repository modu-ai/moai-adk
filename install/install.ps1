# MoAI-ADK installer (Rust edition) for PowerShell — served at https://adk.mo.ai.kr/install.ps1 (#2935).
#
#   irm https://adk.mo.ai.kr/install.ps1 | iex
#
# Same contract as install.sh: download the `moai-cli` release archive for this machine,
# verify its SHA-256 before anything is written (a missing, malformed or different checksum
# stops the install), and place it in the version folder the binary itself names
# (`moai-cli install-inspect` -> `home_state_root`):
#
#   <state home>\versions\<version>\bin\moai-cli.exe   (+ moai.exe, a hard link to it)
#   <link dir>\moai-cli.exe, <link dir>\moai.exe       (default <state home>\path)
#
# The User PATH and shell profiles are never edited, and an existing Go-edition `moai` is never
# touched: the script prints what to run next instead. Requires PowerShell 5.1 or later; on
# macOS and Linux `pwsh` runs the same flow with the `.tar.xz` archive.

$ErrorActionPreference = "Stop"

$RepoUrl = "https://github.com/modu-ai/moai-adk"
$ReleaseBase = if ($env:MOAI_RELEASE_BASE) { $env:MOAI_RELEASE_BASE } else { "$RepoUrl/releases/download" }
$ReleaseFeed = if ($env:MOAI_RELEASE_FEED) { $env:MOAI_RELEASE_FEED } else { "$RepoUrl/releases.atom" }
$TagPrefix = "moai-cli-v"
$IsWin = ($IsWindows -or ($null -eq $IsWindows -and $env:OS -eq "Windows_NT"))
$Exe = if ($IsWin) { ".exe" } else { "" }

function Print-Info { Write-Host "[INFO] " -ForegroundColor Cyan -NoNewline; Write-Host $args }
function Print-Success { Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline; Write-Host $args }
function Print-Warning { Write-Host "[WARNING] " -ForegroundColor Yellow -NoNewline; Write-Host $args }
function Print-Error { Write-Host "[ERROR] " -ForegroundColor Red -NoNewline; Write-Host $args }

$script:TmpDir = $null
$script:StageDir = $null
function Remove-Temp {
    foreach ($d in @($script:StageDir, $script:TmpDir)) {
        if ($d -and (Test-Path -LiteralPath $d)) { Remove-Item -LiteralPath $d -Recurse -Force -ErrorAction SilentlyContinue }
    }
}
function Fail([string]$Message) {
    Print-Error $Message
    Remove-Temp
    exit 1
}

function Show-Usage {
    Write-Host "Usage: install.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  --version VERSION     Install a specific version (default: latest)"
    Write-Host "  --install-dir DIR     Put the moai and moai-cli commands in DIR (default: <state home>\path)"
    Write-Host "  --channel CHANNEL     stable (default) or rc: rc also takes release candidates"
    Write-Host "  -h, --help            Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host '  irm https://adk.mo.ai.kr/install.ps1 | iex'
    Write-Host '  & ([scriptblock]::Create((irm https://adk.mo.ai.kr/install.ps1))) --version 2026.11.0'
    Write-Host '  & ([scriptblock]::Create((irm https://adk.mo.ai.kr/install.ps1))) --channel rc'
}

# A version is SemVer-shaped text only: it becomes part of a URL and a folder name.
function Test-Version([string]$V) {
    return ($V -match '^[0-9][0-9A-Za-z.+-]*$') -and ($V -notmatch '\.\.')
}

# Windows architecture detection keeps the Go installer's layers (env vars, 64-bit OS check).
function Get-Target {
    if ($IsWin) {
        $arch = if ($env:ARCHITEW6432) { $env:ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
        if ($arch -eq "x86" -and [Environment]::Is64BitOperatingSystem) { $arch = "AMD64" }
        switch -Exact ("$arch".ToUpper()) {
            "AMD64" { $t = "x86_64-pc-windows-msvc" }
            "X64" { $t = "x86_64-pc-windows-msvc" }
            "ARM64" { $t = "aarch64-pc-windows-msvc" }
            default { Print-Info "Supported architectures: AMD64, ARM64"; Fail "Unsupported architecture: $arch" }
        }
    }
    else {
        $os = (& uname -s).Trim()
        $machine = (& uname -m).Trim()
        switch ($os) {
            "Darwin" { $osPart = "apple-darwin" }
            "Linux" { $osPart = "unknown-linux-gnu" }
            default { Fail "Unsupported OS: $os" }
        }
        switch ($machine) {
            { $_ -in "x86_64", "amd64" } { $archPart = "x86_64" }
            { $_ -in "arm64", "aarch64" } { $archPart = "aarch64" }
            default { Fail "Unsupported architecture: $machine" }
        }
        $t = "$archPart-$osPart"
    }
    Print-Success "Detected platform: $t"
    return $t
}

# Invoke-WebRequest does not read file:// URLs; tests serve a fake release from a folder.
function Get-Url([string]$Url, [string]$OutFile) {
    if ($Url.StartsWith("file://")) {
        Copy-Item -LiteralPath ([Uri]$Url).LocalPath -Destination $OutFile
        return
    }
    if (-not $IsWin -or $PSVersionTable.PSVersion.Major -ge 6) {
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing -TimeoutSec 300
    }
    else {
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing -TimeoutSec 300
    }
}

# The newest moai-cli-v<version> tag in the releases feed (not the rate-limited REST API;
# /releases/latest is the Go edition's: Rust final releases are published with --latest=false).
function Get-LatestVersion([string]$Channel) {
    $feed = Join-Path $script:TmpDir "releases.atom"
    try { Get-Url $ReleaseFeed $feed }
    catch {
        Print-Info "You can install a specific version: install.ps1 --version <version>"
        Fail "Could not read the release list from GitHub"
    }
    $text = Get-Content -LiteralPath $feed -Raw
    foreach ($m in [regex]::Matches($text, "/releases/tag/$([regex]::Escape($TagPrefix))([^""/?#<]+)")) {
        $v = $m.Groups[1].Value
        if (-not (Test-Version $v)) { continue }
        if ($v.Contains("-") -and $Channel -ne "rc") { continue }
        Print-Success "Latest version ($Channel): $v"
        return $v
    }
    if ($Channel -eq "stable") { Print-Info "No stable release yet. For a release candidate: install.ps1 --channel rc" }
    Fail "No $TagPrefix* release found for channel $Channel"
}

function Get-Sha256([string]$Path) {
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLower()
}

function Main {
    param($Arguments)

    $Version = ""
    $InstallDir = $env:MOAI_INSTALL_DIR
    $Channel = "stable"
    for ($i = 0; $i -lt $Arguments.Count; $i++) {
        $arg = $Arguments[$i]
        if ($arg -in "--version", "-version") {
            if ($i + 1 -ge $Arguments.Count) { Fail "--version needs a value" }
            $Version = [string]$Arguments[++$i]
            if (-not (Test-Version $Version)) { Fail "Not a version: '$Version'" }
        }
        elseif ($arg -in "--install-dir", "-install-dir") {
            if ($i + 1 -ge $Arguments.Count) { Fail "--install-dir needs a value" }
            $InstallDir = [string]$Arguments[++$i]
        }
        elseif ($arg -in "--channel", "-channel") {
            if ($i + 1 -ge $Arguments.Count) { Fail "--channel needs a value" }
            $Channel = [string]$Arguments[++$i]
            if ($Channel -notin "stable", "rc") { Fail "Unknown channel: $Channel (stable or rc)" }
        }
        elseif ($arg -in "-h", "--help", "-help") {
            Show-Usage
            exit 0
        }
        else {
            Print-Error "Unknown option: $arg"
            Write-Host "Use --help for usage information"
            exit 1
        }
    }

    Write-Host ""
    Write-Host "MoAI-ADK installer (Rust edition)"
    Write-Host ""
    $target = Get-Target
    $script:TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) "moai-install-$([System.Guid]::NewGuid().ToString())"
    New-Item -ItemType Directory -Path $script:TmpDir -Force | Out-Null

    if (-not $Version) { $Version = Get-LatestVersion $Channel }
    else { Print-Info "Installing version: $Version" }

    # Download and verify: nothing outside the temporary folder is written before this passes.
    $ext = if ($IsWin) { "zip" } else { "tar.xz" }
    $asset = "moai-cli-$target.$ext"
    $url = "$ReleaseBase/$TagPrefix$Version/$asset"
    $archive = Join-Path $script:TmpDir $asset
    Print-Info "Downloading from: $url"
    try { Get-Url $url $archive } catch { Fail "Download failed: $_" }
    try { Get-Url "$url.sha256" "$archive.sha256" }
    catch { Fail "Could not download the checksum $asset.sha256: nothing is installed without it" }
    $expected = ((Get-Content -LiteralPath "$archive.sha256" -TotalCount 1) -split '\s+')[0].ToLower()
    if ($expected -notmatch '^[0-9a-f]{64}$') { Fail "The checksum file is not a SHA-256 digest: nothing is installed" }
    $actual = Get-Sha256 $archive
    if ($expected -ne $actual) {
        Print-Error "Expected: $expected"
        Print-Error "Actual:   $actual"
        Fail "Checksum mismatch!"
    }
    Print-Success "Checksum verified"

    $x = Join-Path $script:TmpDir "x"
    New-Item -ItemType Directory -Path $x | Out-Null
    try {
        if ($IsWin) { Expand-Archive -LiteralPath $archive -DestinationPath $x -Force }
        else { & tar -xJf $archive -C $x; if ($LASTEXITCODE -ne 0) { throw "tar exit $LASTEXITCODE" } }
    }
    catch { Fail "Failed to extract the archive: $_" }
    $newBin = Get-ChildItem -LiteralPath $x -Recurse -Depth 1 -File -Filter "moai-cli$Exe" | Select-Object -First 1
    if (-not $newBin) { Fail "moai-cli$Exe not found in the archive" }
    $newBin = $newBin.FullName
    if (-not $IsWin) { & chmod 755 $newBin }

    # Ask the downloaded binary where its state lives and what `moai` on PATH is (read only).
    try {
        $inspect = (& $newBin install-inspect 2>$null | Out-String) | ConvertFrom-Json
    }
    catch { Fail "The downloaded moai-cli does not run on this machine" }
    $stateHome = [string]$inspect.result.home_state_root
    if (-not $stateHome -or -not [System.IO.Path]::IsPathRooted($stateHome)) {
        Fail "The downloaded moai-cli did not name an absolute state folder: nothing is installed"
    }
    $goOnPath = ("$($inspect.result.path_first_moai.generation)" -eq "go")

    # Place: a new version folder is staged beside its final name and renamed into place; an
    # existing one is kept when it holds these bytes and refused otherwise.
    $versions = Join-Path $stateHome "versions"
    $dest = Join-Path $versions $Version
    $destCli = Join-Path (Join-Path $dest "bin") "moai-cli$Exe"
    if (Test-Path -LiteralPath $dest) {
        if ((Test-Path -LiteralPath $destCli -PathType Leaf) -and ((Get-Sha256 $destCli) -eq (Get-Sha256 $newBin))) {
            Print-Success "$Version is already installed: $dest"
        }
        else {
            Print-Error "$dest already exists with different contents; it is left as it is."
            Fail "Remove or rename that folder yourself if you want this download in its place"
        }
    }
    else {
        New-Item -ItemType Directory -Path $versions -Force | Out-Null
        $script:StageDir = Join-Path $versions ".install-$Version.$([System.Guid]::NewGuid().ToString('N'))"
        $stageBin = Join-Path $script:StageDir "bin"
        New-Item -ItemType Directory -Path $stageBin -Force | Out-Null
        Copy-Item -LiteralPath $newBin -Destination (Join-Path $stageBin "moai-cli$Exe")
        if (-not $IsWin) { & chmod 755 (Join-Path $stageBin "moai-cli$Exe") }
        New-Item -ItemType HardLink -Path (Join-Path $stageBin "moai$Exe") -Target (Join-Path $stageBin "moai-cli$Exe") | Out-Null
        Move-Item -LiteralPath $script:StageDir -Destination $dest
        $script:StageDir = $null
        Print-Success "Installed to: $(Join-Path $dest 'bin')"
    }

    # The commands: links (copies on Windows, where a symbolic link needs extra rights) in the
    # link folder. A file there that this installer did not make is left as it is.
    $linkDir = if ($InstallDir) { $InstallDir } else { Join-Path $stateHome "path" }
    New-Item -ItemType Directory -Path $linkDir -Force | Out-Null
    foreach ($name in @("moai-cli$Exe", "moai$Exe")) {
        $link = Join-Path $linkDir $name
        $src = Join-Path (Join-Path $dest "bin") $name
        $item = Get-Item -LiteralPath $link -Force -ErrorAction SilentlyContinue
        if ($IsWin) {
            if ($item -and ((Get-Sha256 $link) -eq (Get-Sha256 $src))) { continue }
            if ($item) { Print-Warning "$link is not a copy of this version; left as it is"; continue }
            Copy-Item -LiteralPath $src -Destination $link
            Print-Success "Copied: $link"
            continue
        }
        if ($item -and -not $item.LinkType) { Print-Warning "$link is not a link made by this installer; left as it is"; continue }
        if ($item -and ("$($item.Target)" -eq $src)) { continue }
        if ($item) { Remove-Item -LiteralPath $link -Force }
        New-Item -ItemType SymbolicLink -Path $link -Target $src | Out-Null
        Print-Success "Linked: $link -> $src"
    }

    Remove-Temp
    $cli = Join-Path (Join-Path $dest "bin") "moai-cli$Exe"
    Write-Host (& $cli --version | Out-String).Trim()
    if ($goOnPath) {
        Write-Host ""
        Print-Warning "The moai first on your PATH is the Go edition; it was not changed."
        Print-Info "Preview what moving to the Rust edition changes (writes nothing):"
        Write-Host "  $cli update"
    }
    $sep = [System.IO.Path]::PathSeparator
    if (-not ("$sep$env:PATH$sep".Contains("$sep$linkDir$sep"))) {
        Write-Host ""
        Print-Info "$linkDir is not on your PATH. Add it (this installer does not edit PATH):"
        if ($IsWin) { Write-Host "  [Environment]::SetEnvironmentVariable('Path', '$linkDir;' + [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')" }
        else { Write-Host "  export PATH=""$linkDir`:`$PATH""" }
    }
    Write-Host ""
    Print-Success "Installation complete!"
    Print-Info "Documentation: $RepoUrl"
}

Main -Arguments $args
