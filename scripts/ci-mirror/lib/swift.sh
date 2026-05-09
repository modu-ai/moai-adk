#!/bin/sh
# scripts/ci-mirror/lib/swift.sh — Swift smoke pipeline
# Runs swift test; exits 0 silently when swift is absent.
set -eu
log() { printf '[ci-mirror][swift] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v swift >/dev/null 2>&1; then
    log "swift not installed — skipping"
    exit 0
fi
log "running swift test"
swift test || exit 2
log "OK"
exit 0
