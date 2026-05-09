#!/bin/sh
# scripts/ci-mirror/lib/scala.sh — Scala smoke pipeline
# Runs sbt test; exits 0 silently when sbt is absent.
set -eu
log() { printf '[ci-mirror][scala] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v sbt >/dev/null 2>&1; then
    log "sbt not installed — skipping"
    exit 0
fi
log "running sbt test"
sbt test || exit 2
log "OK"
exit 0
