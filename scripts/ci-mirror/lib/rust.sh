#!/bin/sh
# scripts/ci-mirror/lib/rust.sh — Rust smoke pipeline
# Runs cargo test (and clippy if available); exits 0 silently when cargo is absent.
set -eu
log() { printf '[ci-mirror][rust] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v cargo >/dev/null 2>&1; then
    log "cargo not installed — skipping"
    exit 0
fi
log "running cargo test --quiet"
cargo test --quiet || exit 2
if cargo clippy --version >/dev/null 2>&1; then
    log "running cargo clippy"
    cargo clippy -- -D warnings || exit 2
fi
log "OK"
exit 0
