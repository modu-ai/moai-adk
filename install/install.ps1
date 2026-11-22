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
#
# The language of the conversation with the model (#2935 L1) comes from --language, else
# MOAI_LANGUAGE, else (at a terminal) a numbered menu, else the locale, else English. After the
# commands are linked the script runs `moai-cli language-set --home --if-absent <code>` once;
# that records only when nothing is recorded yet, so a reinstall never overwrites an earlier
# choice, and a moai-cli without it only costs a warning. Everything this script prints stays
# English whatever the language.
#
# A release archive carries the skill package and the base agents beside the binary (`skills\`,
# `agents\`); they go into the version folder next to `bin\`, where `moai update` reads them
# (SPEC-20261117-installer-skill-agents). Into the host (Claude Code) they go only
# through the product's own preview and confirmed apply, `moai-cli update --json` and then
# `moai-cli update --confirm <plan_digest> --json`: this script copies nothing into a host
# folder. It shows the folders the preview names, asks once at a terminal (--yes answers for a
# run without one) and then names the next steps for each host it found.

$ErrorActionPreference = "Stop"

$RepoUrl = "https://github.com/modu-ai/moai-adk"
$ReleaseBase = if ($env:MOAI_RELEASE_BASE) { $env:MOAI_RELEASE_BASE } else { "$RepoUrl/releases/download" }
$ReleaseFeed = if ($env:MOAI_RELEASE_FEED) { $env:MOAI_RELEASE_FEED } else { "$RepoUrl/releases.atom" }
$ReleaseApi = if ($env:MOAI_RELEASE_API) { $env:MOAI_RELEASE_API } else { "https://api.github.com/repos/modu-ai/moai-adk/releases?per_page=100" }
$TagPrefix = "moai-cli-v"
$IsWin = ($IsWindows -or ($null -eq $IsWindows -and $env:OS -eq "Windows_NT"))
$Exe = if ($IsWin) { ".exe" } else { "" }

function Print-Info { Write-Host "[INFO] " -ForegroundColor Cyan -NoNewline; Write-Host $args }
function Print-Success { Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline; Write-Host $args }
function Print-Warning { Write-Host "[WARNING] " -ForegroundColor Yellow -NoNewline; Write-Host $args }
function Print-Error { Write-Host "[ERROR] " -ForegroundColor Red -NoNewline; Write-Host $args }

$script:TmpDir = $null
$script:StageDir = $null
$script:PartTmp = $null
function Remove-Temp {
    foreach ($d in @($script:PartTmp, $script:StageDir, $script:TmpDir)) {
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
    Write-Host "  --version VERSION     Install a specific version, for example 4.2647.12 (default: latest)"
    Write-Host "  --install-dir DIR     Put the moai and moai-cli commands in DIR (default: <state home>\path)"
    Write-Host "  --channel CHANNEL     stable (default) or rc: rc also takes releases marked as pre-releases"
    Write-Host "  --language CODE       Language of your conversation with the model, for example en, ko, ja or"
    Write-Host "                        zh (default: asked at a terminal, else MOAI_LANGUAGE, else your locale,"
    Write-Host "                        else en). MoAI's own messages stay English."
    Write-Host "  --yes                 Install the MoAI skill and agents into the hosts found without asking"
    Write-Host "                        (the preview is still shown); without a terminal and without --yes"
    Write-Host "                        nothing is written into your home folder's host settings"
    Write-Host "  -h, --help            Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host '  irm https://adk.mo.ai.kr/install.ps1 | iex'
    Write-Host '  & ([scriptblock]::Create((irm https://adk.mo.ai.kr/install.ps1))) --version 4.2647.12'
    Write-Host '  & ([scriptblock]::Create((irm https://adk.mo.ai.kr/install.ps1))) --channel rc'
    Write-Host '  & ([scriptblock]::Create((irm https://adk.mo.ai.kr/install.ps1))) --language ko'
}

# A version is SemVer-shaped text only: it becomes part of a URL and a folder name. A HeadVer
# version (`4.2647.12`, `{head}.{yearweek}.{build}`) is one; so is `2026.11.0-rc.1` of the earlier
# scheme.
function Test-Version([string]$V) {
    return ($V -match '^[0-9][0-9A-Za-z.+-]*$') -and ($V -notmatch '\.\.')
}

# The conversation language is one of the four codes the language module holds (#2935 L1, G6):
# en, ko, ja and zh, the same list the install menu offers and moai-cli accepts. Anything else —
# a well-shaped code like fr or pt-BR included — is refused here, so nothing outside the list
# reaches the binary or the recorded file.
function Test-Language([string]$V) {
    return (($V -ceq "en") -or ($V -ceq "ko") -or ($V -ceq "ja") -or ($V -ceq "zh"))
}

# The language of the first of LC_ALL, LC_MESSAGES, LANG that is set (`ko_KR.UTF-8` -> ko), else
# the culture of the session (Windows); as written, or "" when it is C, POSIX, invariant or no code.
function Get-LocaleLanguage {
    foreach ($name in @("LC_ALL", "LC_MESSAGES", "LANG")) {
        $loc = [Environment]::GetEnvironmentVariable($name)
        if (-not $loc) { continue }
        $part = ($loc -split '[_.@]')[0]
        if (($part -in "C", "POSIX") -or -not (Test-Language $part)) { return "" }
        return $part
    }
    try { $part = (Get-Culture).TwoLetterISOLanguageName } catch { return "" }
    if ($part -and $part -ne "iv" -and (Test-Language $part)) { return $part }
    return ""
}

# One line from the terminal after the prompt text. [Console]::In.ReadLine() gives $null when the
# input is closed, and a null answer takes the default. Read-Host would end the whole host there,
# and the install with it, on Linux. Whether Ctrl-D closes the input depends on the host: press
# Enter to take the default.
function Read-Answer([string]$Prompt) {
    Write-Host "${Prompt}: " -NoNewline
    try { $line = [Console]::In.ReadLine() } catch { $line = $null }
    if ($null -eq $line) { Write-Host "" }
    return $line
}

# Asks for the language at a terminal. A wrong answer is asked again, three answers at most, then
# the default is used; an empty answer or the end of input takes the default too.
function Read-LanguageChoice([string]$Default) {
    Write-Host ""
    Write-Host "Which language should the model use when it talks with you?"
    Write-Host "  1) English"
    Write-Host ([string]::Concat("  2) ", [char]0xD55C, [char]0xAD6D, [char]0xC5B4))
    Write-Host ([string]::Concat("  3) ", [char]0x65E5, [char]0x672C, [char]0x8A9E))
    Write-Host ([string]::Concat("  4) ", [char]0x4E2D, [char]0x6587))
    $known = @{ "1" = "en"; "2" = "ko"; "3" = "ja"; "4" = "zh" }
    for ($try = 0; $try -lt 3; $try++) {
        $answer = Read-Answer "Choose 1-4 [default: $Default]"
        if ($null -eq $answer) { break }
        $answer = $answer.Trim()
        if ($answer -eq "") { return $Default }
        if ($known.ContainsKey($answer)) { return $known[$answer] }
        Print-Warning "Please type a number from 1 to 4."
    }
    Print-Info "Using $Default."
    return $Default
}

# The supported platforms (user decision 2026-11-17, SPEC-20261117-supported-platforms): macOS on
# Apple silicon, Linux x86_64 and Windows x86_64. Anything else stops before a download or a write.
$SupportedPlatforms = "macOS arm64 (Apple silicon), Linux x86_64, Windows x86_64"

# Windows: the triple for the architecture Windows reports, or $null when it is not supported.
# A 32-bit PowerShell on 64-bit x86 Windows reports x86 (the Go installer's 64-bit OS check).
function Get-WindowsTarget([string]$Arch, [bool]$Is64BitOs) {
    if ($Arch -eq "x86" -and $Is64BitOs) { $Arch = "AMD64" }
    switch -Exact ($Arch.ToUpper()) {
        "AMD64" { return "x86_64-pc-windows-msvc" }
        "X64" { return "x86_64-pc-windows-msvc" }
    }
    return $null
}

# macOS and Linux (pwsh): the triple for `uname -s` / `uname -m`, or $null when not supported.
# An x86_64 shell under Rosetta 2 on Apple silicon ($Translated "1") runs on an arm64 machine.
function Get-UnixTarget([string]$Os, [string]$Machine, [string]$Translated) {
    if ($Os -eq "Darwin" -and $Machine -eq "x86_64" -and $Translated -eq "1") { $Machine = "arm64" }
    if ($Os -eq "Darwin" -and $Machine -in "arm64", "aarch64") { return "aarch64-apple-darwin" }
    if ($Os -eq "Linux" -and $Machine -in "x86_64", "amd64") { return "x86_64-unknown-linux-gnu" }
    return $null
}

# Windows architecture detection keeps the Go installer's layers (env vars, 64-bit OS check).
function Get-Target {
    if ($IsWin) {
        $arch = if ($env:ARCHITEW6432) { $env:ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
        $t = Get-WindowsTarget "$arch" ([Environment]::Is64BitOperatingSystem)
        $what = "Windows $arch"
    }
    else {
        $os = (& uname -s).Trim()
        $machine = (& uname -m).Trim()
        $translated = ""
        # The system's own sysctl, by absolute path: a `sysctl` earlier on PATH is not asked.
        if ($os -eq "Darwin" -and $machine -eq "x86_64") {
            try { $translated = "$(& /usr/sbin/sysctl -n sysctl.proc_translated 2>$null)".Trim() } catch { $translated = "" }
        }
        $t = Get-UnixTarget $os $machine $translated
        $what = "$os $machine"
    }
    if (-not $t) {
        Print-Info "Supported platforms: $SupportedPlatforms"
        Fail "Unsupported platform: $what"
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

# The first candidate (newest first) that is a moai-cli-v<version> release for the channel. A
# pre-release is told by the release's own mark; a `-` in the version marks one of the earlier
# scheme only, because a HeadVer version has no suffix. The feed has no mark, so a HeadVer
# pre-release cannot be told there and the newest tag is taken (docs/release-checklist.md).
function Select-Version($Candidates, [string]$Channel) {
    foreach ($c in $Candidates) {
        $tag = [string]$c.tag
        if (-not $tag.StartsWith($TagPrefix) -or $c.draft) { continue }
        $v = $tag.Substring($TagPrefix.Length)
        if (-not (Test-Version $v)) { continue }
        if ($Channel -ne "rc" -and ($c.prerelease -or $v.Contains("-"))) { continue }
        return $v
    }
    return $null
}

# The newest moai-cli-v<version> release. /releases/latest is the Go edition's (Rust final
# releases are published with --latest=false). First the REST list (100 newest, with the draft
# and prerelease fields); when it cannot be read (no network to the API, or the
# unauthenticated limit is spent), the releases feed, which holds only the 10 newest entries
# (audit #2943 B1).
function Get-LatestVersion([string]$Channel) {
    $v = $null
    $api = Join-Path $script:TmpDir "releases.json"
    try {
        Get-Url $ReleaseApi $api
        $list = Get-Content -LiteralPath $api -Raw | ConvertFrom-Json
        $v = Select-Version ($list | ForEach-Object { @{ tag = $_.tag_name; draft = [bool]$_.draft; prerelease = [bool]$_.prerelease } }) $Channel
    }
    catch { $v = $null }
    if (-not $v) {
        $feed = Join-Path $script:TmpDir "releases.atom"
        try { Get-Url $ReleaseFeed $feed }
        catch {
            Print-Info "You can install a specific version: install.ps1 --version <version>"
            Fail "Could not read the release list from GitHub"
        }
        $text = Get-Content -LiteralPath $feed -Raw
        $tags = [regex]::Matches($text, "/releases/tag/($([regex]::Escape($TagPrefix))[^""/?#<]+)") |
            ForEach-Object { @{ tag = $_.Groups[1].Value; draft = $false; prerelease = $false } }
        $v = Select-Version $tags $Channel
    }
    if ($v) {
        Print-Success "Latest version ($Channel): $v"
        return $v
    }
    if ($Channel -eq "stable") { Print-Info "No stable release yet. For a release candidate: install.ps1 --channel rc" }
    Fail "No $TagPrefix* release found for channel $Channel"
}

function Get-Sha256([string]$Path) {
    return (Get-FileHash -LiteralPath $Path -Algorithm SHA256).Hash.ToLower()
}

# Copies the release parts the archive carries beside the binary (`skills`, `agents`) into the
# version folder, each under a temporary name renamed into place. A part the folder already has,
# or one the archive does not carry (an older release), is left alone. Returns the parts added.
function Add-ReleaseParts([string]$ArchiveRoot, [string]$Folder) {
    $added = @()
    foreach ($part in @("skills", "agents")) {
        $src = Get-Item -LiteralPath (Join-Path $ArchiveRoot $part) -Force -ErrorAction SilentlyContinue
        if (-not $src -or -not $src.PSIsContainer -or $src.LinkType) { continue }
        if (Get-Item -LiteralPath (Join-Path $Folder $part) -Force -ErrorAction SilentlyContinue) { continue }
        $tmp = Join-Path $Folder ".$part.install-$PID"
        $script:PartTmp = $tmp
        try { Copy-Item -LiteralPath $src.FullName -Destination $tmp -Recurse }
        catch { Fail "Could not copy $part into ${Folder}: $_" }
        Move-Item -LiteralPath $tmp -Destination (Join-Path $Folder $part)
        $script:PartTmp = $null
        $added += "$part"
    }
    return $added
}

# The hosts this machine has: the host's command on PATH, or its folder in HOME.
function Get-Hosts {
    $found = @()
    foreach ($h in @("claude")) {
        if ((Get-Command $h -CommandType Application -ErrorAction SilentlyContinue) -or
            (Test-Path -LiteralPath (Join-Path $env:HOME ".$h") -PathType Container)) { $found += $h }
    }
    return $found
}

# Runs `moai-cli update` with these arguments from the version folder (so the hooks of the project
# the user stands in are not part of the plan). Returns the exit code, the parsed envelope ($null
# for an answer that is not JSON) and the first error line.
function Invoke-Update([string]$Cli, [string]$Folder, [string[]]$UpdateArgs) {
    $err = Join-Path $script:TmpDir "update.err"
    Push-Location -LiteralPath $Folder
    try {
        $raw = (& $Cli update @UpdateArgs 2>$err | Out-String)
        $code = $LASTEXITCODE
    }
    catch { $raw = ""; $code = 1 }
    finally { Pop-Location }
    $parsed = $null
    try { $parsed = $raw | ConvertFrom-Json } catch { $parsed = $null }
    $first = (Get-Content -LiteralPath $err -TotalCount 1 -ErrorAction SilentlyContinue)
    return [pscustomobject]@{ Code = $code; Json = $parsed; Err = "$first" }
}

# A failed `moai-cli update`: the binary stays installed, and the product's own message is shown.
function Write-UpdateFailed($Answer) {
    $msg = ""
    if ($Answer -and $Answer.Json -and $Answer.Json.error) { $msg = [string]$Answer.Json.error.message }
    if (-not $msg -and $Answer) { $msg = $Answer.Err }
    if (-not $msg) { $msg = "moai-cli update did not succeed" }
    Print-Warning "The MoAI skill and agents were not installed: $msg"
    Print-Info "The moai command is installed; try again later with: moai update"
}

# Installs the skill and the agents into the hosts found, through the product's preview and
# confirmed apply. Returns $true when they are installed and up to date.
function Install-Skill([string]$Cli, [string]$Folder, $Hosts, [bool]$Yes) {
    Write-Host ""
    if (-not $Hosts) {
        Print-Warning "No Claude Code found (no claude command, no ~/.claude)."
        Print-Info "Install it, then install the MoAI skill and agents with: moai update"
        return $false
    }
    Print-Info "Previewing the MoAI skill and agents for your hosts (writes nothing):"
    $preview = Invoke-Update $Cli $Folder @("--json")
    if ($preview.Code -ne 0 -or -not $preview.Json -or $preview.Json.ok -ne $true) { Write-UpdateFailed $preview; return $false }
    $result = $preview.Json.result
    if ($result.nothing_to_apply -eq $true) {
        Print-Success "The MoAI skill and agents are already installed and up to date"
        return $true
    }
    $digest = [string]$result.plan_digest
    if ($digest -notmatch '^sha256:[0-9a-f]{64}$') {
        Write-UpdateFailed ([pscustomobject]@{ Json = $null; Err = "the preview named no plan to confirm" })
        return $false
    }
    foreach ($h in @($result.plan.hosts)) { if ($h.projection_root) { Write-Host "  skill:  $($h.projection_root)" } }
    foreach ($h in @($result.agents.hosts)) { if ($h.dir) { Write-Host "  agents: $($h.dir)" } }
    Write-Host "  and MoAI's sections in those hosts' settings files (moai update lists every file)"
    if ($Yes) { Print-Info "--yes: installing them" }
    elseif (-not [Console]::IsInputRedirected -and -not [Console]::IsOutputRedirected) {
        $answer = Read-Answer "Install them now? [y/N]"
        if ("$answer".Trim() -notin "y", "yes") {
            Print-Info "Nothing was written to your home folder. Install them later with: moai update"
            return $false
        }
    }
    else {
        Print-Info "Nothing was written to your home folder (no terminal to ask at). Install them with: moai update (or run this installer with --yes)"
        return $false
    }
    $applied = Invoke-Update $Cli $Folder @("--confirm", $digest, "--json")
    if ($applied.Code -ne 0 -or -not $applied.Json -or $applied.Json.ok -ne $true -or $applied.Json.result.applied -ne $true) {
        Write-UpdateFailed $applied
        return $false
    }
    Print-Success "Installed the MoAI skill and agents"
    return $true
}

function Main {
    param($Arguments)

    $Version = ""
    $InstallDir = $env:MOAI_INSTALL_DIR
    $Channel = "stable"
    $Language = ""
    $Yes = $false
    for ($i = 0; $i -lt $Arguments.Count; $i++) {
        $arg = $Arguments[$i]
        if ($arg -in "--version", "-version") {
            if ($i + 1 -ge $Arguments.Count) { Fail "--version needs a value" }
            $given = [string]$Arguments[++$i]
            # `v2026.11.0` and the tag `moai-cli-v2026.11.0` name the same version.
            $Version = $given
            if ($Version.StartsWith($TagPrefix)) { $Version = $Version.Substring($TagPrefix.Length) }
            if ($Version.StartsWith("v")) { $Version = $Version.Substring(1) }
            if (-not (Test-Version $Version)) { Fail "Not a version: '$given'" }
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
        elseif ($arg -in "--language", "-language") {
            if ($i + 1 -ge $Arguments.Count) { Fail "--language needs a value" }
            $given = [string]$Arguments[++$i]
            if (-not (Test-Language $given)) { Fail "Not a language code: '$given' (for example en, ko, ja, zh)" }
            $Language = $given
        }
        elseif ($arg -in "--yes", "-yes") { $Yes = $true }
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

    # Releases are read over https only (file:// for the tests): the checksum comes from the
    # same place as the archive, so a plain-http source would make the check meaningless.
    foreach ($url in @($ReleaseBase, $ReleaseFeed, $ReleaseApi)) {
        if (-not ($url.StartsWith("https://") -or $url.StartsWith("file://"))) {
            Fail "Release sources must be https:// (or file://): $url"
        }
    }

    Write-Host ""
    Write-Host "MoAI-ADK installer (Rust edition)"
    Write-Host ""
    # The language of the conversation, settled before anything is downloaded or written:
    # --language, else MOAI_LANGUAGE, else the menu at a terminal, else the locale, else en.
    if (-not $Language -and $env:MOAI_LANGUAGE) {
        if (-not (Test-Language $env:MOAI_LANGUAGE)) { Fail "Not a language code: '$($env:MOAI_LANGUAGE)' (for example en, ko, ja, zh)" }
        $Language = $env:MOAI_LANGUAGE
    }
    if (-not $Language) {
        $Language = Get-LocaleLanguage
        if (-not $Language) { $Language = "en" }
        if (-not [Console]::IsInputRedirected -and -not [Console]::IsOutputRedirected) { $Language = Read-LanguageChoice $Language }
    }
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
    # moai-cli reads HOME; Windows often has only USERPROFILE (codex review P1 on #2943).
    if (-not $env:HOME -and $env:USERPROFILE) { $env:HOME = $env:USERPROFILE }
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
            # A folder an earlier installer placed holds bin only: add what this archive carries.
            $added = @(Add-ReleaseParts (Split-Path -Parent $newBin) $dest)
            if ($added.Count -gt 0) { Print-Success "Added to ${dest}: $($added -join ', ')" }
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
        $added = @(Add-ReleaseParts (Split-Path -Parent $newBin) $script:StageDir)
        Move-Item -LiteralPath $script:StageDir -Destination $dest
        $script:StageDir = $null
        Print-Success "Installed to: $(Join-Path $dest 'bin')"
        if ($added.Count -gt 0) { Print-Success "With the release parts: $($added -join ', ')" }
    }

    # The commands: links (copies on Windows, where a symbolic link needs extra rights) in the
    # link folder. A file there that this installer did not make is left as it is.
    $linkDir = if ($InstallDir) { $InstallDir } else { Join-Path $stateHome "path" }
    New-Item -ItemType Directory -Path $linkDir -Force | Out-Null
    $versionsPrefix = (Join-Path $stateHome "versions") + [System.IO.Path]::DirectorySeparatorChar
    foreach ($name in @("moai-cli$Exe", "moai$Exe")) {
        $link = Join-Path $linkDir $name
        $src = Join-Path (Join-Path $dest "bin") $name
        $item = Get-Item -LiteralPath $link -Force -ErrorAction SilentlyContinue
        $tmp = Join-Path $linkDir ".$name.install-$PID"
        if ($IsWin) {
            if ($item) {
                $have = Get-Sha256 $link
                if ($have -eq (Get-Sha256 $src)) { continue }
                # A copy of a version this installer placed is its own and is replaced on an
                # upgrade (codex review P2 on #2943); any other file is someone else's.
                $owned = Get-ChildItem -LiteralPath (Join-Path $stateHome "versions") -Directory -ErrorAction SilentlyContinue |
                    Where-Object { $f = Join-Path (Join-Path $_.FullName "bin") $name; (Test-Path -LiteralPath $f) -and ((Get-Sha256 $f) -eq $have) }
                if (-not $owned) { Print-Warning "$link is not a copy this installer placed; left as it is"; continue }
            }
            Copy-Item -LiteralPath $src -Destination $tmp
            Move-Item -LiteralPath $tmp -Destination $link -Force
            Print-Success "Copied: $link"
            continue
        }
        if ($item -and -not $item.LinkType) { Print-Warning "$link is not a link made by this installer; left as it is"; continue }
        if ($item) {
            $current = "$($item.Target)"
            if ($current -eq $src) { continue }
            # Only a link into a version folder this installer placed is replaced; a link to the
            # Go edition's moai, say, is someone else's (codex review P1 on #2943).
            if (-not $current.StartsWith($versionsPrefix)) {
                Print-Warning "$link points to $current, not a version this installer placed; left as it is"
                continue
            }
        }
        # Made under a temporary name and renamed over the old link, so moai never goes missing.
        New-Item -ItemType SymbolicLink -Path $tmp -Target $src | Out-Null
        Move-Item -LiteralPath $tmp -Destination $link -Force
        Print-Success "Linked: $link -> $src"
    }

    # Record the language once, with the installed moai-cli. It passes the record-once mode
    # (--if-absent), so a value that is already recorded is never overwritten: only an explicit
    # `moai language-set --home <code>` replaces it. A moai-cli that has no language-set (an
    # older pinned version) or refuses the code costs a warning, not the install.
    $cli = Join-Path (Join-Path $dest "bin") "moai-cli$Exe"
    $recorded = $null
    try {
        $answer = (& $cli language-set --home --if-absent $Language 2>$null | Out-String)
        if ($LASTEXITCODE -eq 0) { $recorded = $answer | ConvertFrom-Json }
    }
    catch { $recorded = $null }
    if (-not $recorded) {
        Print-Warning "Could not record the conversation language ($Language): this moai-cli has no language-set or does not accept the code. The install is complete."
    }
    elseif ($recorded.result.written -eq $true) { Print-Success "Conversation language recorded: $Language" }
    elseif ("$($recorded.result.reason)" -eq "exists") { Print-Info "A conversation language is already recorded; left as it is" }
    elseif ("$($recorded.result.reason)" -eq "no_home_state") { Print-Info "No MoAI state folder in your home yet; the conversation language was not recorded" }
    else { Print-Warning "The conversation language answer was not understood; it may not be recorded." }

    $hosts = @(Get-Hosts)
    $skillReady = Install-Skill $cli $dest $hosts $Yes

    Remove-Temp
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
    Write-Host ""
    Print-Info "Next steps:"
    $n = 1
    if (-not $skillReady) {
        Write-Host "  $n. Install the MoAI skill and agents (shows what it writes, then asks): moai update"
        $n++
    }
    Write-Host "  $n. Make a project folder and enter it: New-Item -ItemType Directory my-project; Set-Location my-project"
    $n++
    Write-Host "  $n. Start a session there, then set the project up inside it:"
    $entries = if ($hosts.Count -gt 0) { $hosts } else { @("claude") }
    foreach ($h in $entries) {
        switch ($h) {
            "claude" { Write-Host "     Claude Code: moai cc (or moai glm; -l leader, -a agent), then /moai project" }
        }
    }
    Write-Host ""
    Print-Info "Documentation: $RepoUrl"
}

Main -Arguments $args
