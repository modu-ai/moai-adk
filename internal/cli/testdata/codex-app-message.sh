#!/bin/sh
# SPEC-CODEX-LAUNCHER-001 AC-CL-016 fixture: a codex stand-in whose stdout
# must survive the launcher byte-for-byte (no filtering, rewriting, or
# reinterpretation). POSIX leg; the .bat twin covers Windows. The optional
# CODEX_FIXTURE_EXIT env var makes the child exit non-zero for the
# exit-code-propagation cell (defaults to 0 on both platforms).
echo 'install the desktop app from https://example.invalid/codex-desktop'
exit "${CODEX_FIXTURE_EXIT:-0}"
