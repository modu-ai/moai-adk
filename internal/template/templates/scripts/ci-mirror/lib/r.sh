#!/bin/sh
# scripts/ci-mirror/lib/r.sh — R smoke pipeline (best effort)
# Runs R CMD check; exits 0 silently when R is absent.
set -eu
log() { printf '[ci-mirror][r] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v R >/dev/null 2>&1; then
    log "R not installed — skipping"
    exit 0
fi
log "running R CMD check --no-manual"
R CMD check . --no-manual 2>/dev/null || exit 2
log "OK"
exit 0
