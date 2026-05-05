#!/bin/sh
# scripts/ci-mirror/lib/ruby.sh — Ruby smoke pipeline
# Runs rspec or rake test; exits 0 silently when neither is present.
set -eu
log() { printf '[ci-mirror][ruby] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if command -v rspec >/dev/null 2>&1; then
    log "running rspec --format progress"
    rspec --format progress 2>/dev/null || exit 2
elif command -v rake >/dev/null 2>&1; then
    log "running rake test"
    rake test || exit 2
else
    log "rspec/rake not installed — skipping"
    exit 0
fi
log "OK"
exit 0
