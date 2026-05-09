#!/bin/sh
# scripts/ci-mirror/lib/php.sh — PHP smoke pipeline
# Runs phpunit; exits 0 silently when phpunit is absent.
set -eu
log() { printf '[ci-mirror][php] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if command -v vendor/bin/phpunit >/dev/null 2>&1; then
    log "running vendor/bin/phpunit --no-coverage"
    vendor/bin/phpunit --no-coverage 2>/dev/null || exit 2
elif command -v phpunit >/dev/null 2>&1; then
    log "running phpunit"
    phpunit || exit 2
else
    log "phpunit not installed — skipping"
    exit 0
fi
log "OK"
exit 0
