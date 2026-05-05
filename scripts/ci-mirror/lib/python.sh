#!/bin/sh
# scripts/ci-mirror/lib/python.sh — Python smoke pipeline
# Runs pytest if available; exits 0 silently when pytest is absent.
set -eu
log() { printf '[ci-mirror][python] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v pytest >/dev/null 2>&1; then
    log "pytest not installed — skipping"
    exit 0
fi
log "running pytest -q"
pytest -q || exit 2
log "OK"
exit 0
