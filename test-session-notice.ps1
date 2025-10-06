# Windows PowerShell Test Script for session-notice
# @CODE:INIT-002:TEST | PowerShell version of test-session-notice.js

$projectRoot = Get-Location

Write-Host "Project Root: $projectRoot" -ForegroundColor Cyan
Write-Host ""

# Test 1: Check .moai directory
$moaiDir = Join-Path $projectRoot ".moai"
$moaiExists = Test-Path $moaiDir
if ($moaiExists) {
    Write-Host "✓ Check .moai directory: ✅ EXISTS" -ForegroundColor Green
} else {
    Write-Host "✓ Check .moai directory: ❌ NOT FOUND" -ForegroundColor Red
}
Write-Host "  Path: $moaiDir"
Write-Host ""

# Test 2: Check .claude/commands/alfred directory
$alfredCommands = Join-Path $projectRoot ".claude" | Join-Path -ChildPath "commands" | Join-Path -ChildPath "alfred"
$alfredExists = Test-Path $alfredCommands
if ($alfredExists) {
    Write-Host "✓ Check .claude/commands/alfred: ✅ EXISTS" -ForegroundColor Green
} else {
    Write-Host "✓ Check .claude/commands/alfred: ❌ NOT FOUND" -ForegroundColor Red
}
Write-Host "  Path: $alfredCommands"
Write-Host ""

# Test 3: Check old moai path (should not be used)
$moaiCommands = Join-Path $projectRoot ".claude" | Join-Path -ChildPath "commands" | Join-Path -ChildPath "moai"
$moaiCommandsExists = Test-Path $moaiCommands
if ($moaiCommandsExists) {
    Write-Host "✓ Check .claude/commands/moai (old path): ⚠️ EXISTS (not used)" -ForegroundColor Yellow
} else {
    Write-Host "✓ Check .claude/commands/moai (old path): ✅ NOT FOUND" -ForegroundColor Green
}
Write-Host "  Path: $moaiCommands"
Write-Host ""

# Test 4: isMoAIProject logic
$isMoAIProject = $moaiExists -and $alfredExists
Write-Host ("=" * 50)
if ($isMoAIProject) {
    Write-Host "isMoAIProject() result: ✅ TRUE (initialized)" -ForegroundColor Green
} else {
    Write-Host "isMoAIProject() result: ❌ FALSE (not initialized)" -ForegroundColor Red
}
Write-Host ("=" * 50)
Write-Host ""

if ($isMoAIProject) {
    Write-Host "✅ PASS: Project should NOT show initialization message" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ FAIL: Project WILL show initialization message" -ForegroundColor Red
    if (-not $moaiExists) { Write-Host "  Reason: .moai directory missing" -ForegroundColor Red }
    if (-not $alfredExists) { Write-Host "  Reason: .claude/commands/alfred directory missing" -ForegroundColor Red }
    exit 1
}
