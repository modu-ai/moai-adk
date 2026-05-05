#!/bin/sh
# scripts/ci-mirror/lib/elixir.sh — Elixir smoke pipeline
# Runs mix test; exits 0 silently when mix is absent.
set -eu
log() { printf '[ci-mirror][elixir] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v mix >/dev/null 2>&1; then
    log "mix not installed — skipping"
    exit 0
fi
log "running mix test"
mix test || exit 2
log "OK"
exit 0
