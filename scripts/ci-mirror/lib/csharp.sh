#!/bin/sh
# scripts/ci-mirror/lib/csharp.sh — C# smoke pipeline
# Runs dotnet test; exits 0 silently when dotnet is absent.
set -eu
log() { printf '[ci-mirror][csharp] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v dotnet >/dev/null 2>&1; then
    log "dotnet not installed — skipping"
    exit 0
fi
log "running dotnet test --nologo --verbosity quiet"
dotnet test --nologo --verbosity quiet || exit 2
log "OK"
exit 0
