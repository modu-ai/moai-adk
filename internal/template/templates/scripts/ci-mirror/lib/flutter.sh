#!/bin/sh
# scripts/ci-mirror/lib/flutter.sh — Flutter smoke pipeline
# Runs flutter test; exits 0 silently when flutter is absent.
set -eu
log() { printf '[ci-mirror][flutter] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v flutter >/dev/null 2>&1; then
    log "flutter not installed — skipping"
    exit 0
fi
log "running flutter test"
flutter test || exit 2
log "OK"
exit 0
