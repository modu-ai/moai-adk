#!/bin/sh
# scripts/ci-mirror/lib/cpp.sh — C++ smoke pipeline (best effort)
# Runs cmake build + ctest; exits 0 silently when neither is present.
set -eu
log() { printf '[ci-mirror][cpp] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if ! command -v cmake >/dev/null 2>&1; then
    log "cmake not installed — skipping"
    exit 0
fi
if [ -d "$REPO_ROOT/build" ]; then
    log "running cmake --build build --target test"
    cmake --build build --target test 2>/dev/null || ctest 2>/dev/null || exit 2
else
    log "build/ directory not found — skipping (run cmake -B build first)"
    exit 0
fi
log "OK"
exit 0
