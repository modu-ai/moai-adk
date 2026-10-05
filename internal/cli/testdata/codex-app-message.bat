@echo off
rem SPEC-CODEX-LAUNCHER-001 AC-CL-016 fixture: a codex stand-in whose stdout
rem must survive the launcher byte-for-byte. Windows leg; the .sh twin covers
rem POSIX. CODEX_FIXTURE_EXIT makes the child exit non-zero (default 0).
echo install the desktop app from https://example.invalid/codex-desktop
if "%CODEX_FIXTURE_EXIT%"=="" exit /b 0
exit /b %CODEX_FIXTURE_EXIT%
