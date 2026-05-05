#!/bin/sh
# scripts/ci-mirror/lib/kotlin.sh — Kotlin smoke pipeline
# Runs gradle test; exits 0 silently when gradle is absent.
set -eu
log() { printf '[ci-mirror][kotlin] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v gradle >/dev/null 2>&1; then
    log "gradle not installed — skipping"
    exit 0
fi
log "running gradle test --quiet"
gradle test --quiet || exit 2
log "OK"
exit 0
