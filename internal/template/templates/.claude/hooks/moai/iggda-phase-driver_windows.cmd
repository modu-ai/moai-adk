@echo off
REM IGGDA Phase Driver - Windows stub.
REM
REM The IGGDA phase driver is a bash shell script that relies on jq + moai CLI +
REM progress.md parsing. On Windows, these dependencies are not reliably present
REM in the hook execution context. This stub emits a graceful-degradation log
REM and exits 0 (non-blocking) so the Windows build is not broken. The
REM orchestrator's verification-batch re-checks phase transitions deterministically
REM regardless of whether the Stop hook fired.
REM
REM See .claude/rules/moai/workflow/orchestration-mode-selection.md §I.3 for the
REM auto-advance mechanism contract. The full bash driver runs on macOS/Linux.

echo {"continue": true, "stopReason": "iggda_driver_windows_stub", "details": {"reason": "IGGDA phase driver bash script not invoked on Windows; orchestrator verification-batch handles phase transitions deterministically"}} 1>&2
exit /b 0
