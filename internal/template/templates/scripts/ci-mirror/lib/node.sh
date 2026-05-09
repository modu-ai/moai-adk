#!/bin/sh
# scripts/ci-mirror/lib/node.sh — Node.js smoke pipeline
# Runs npm test if available; exits 0 silently when npm is absent.
set -eu
log() { printf '[ci-mirror][node] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v npm >/dev/null 2>&1; then
    log "npm not installed — skipping"
    exit 0
fi
log "running npm test --silent"
npm test --silent || exit 2
log "OK"
exit 0
