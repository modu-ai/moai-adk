#!/bin/sh
# scripts/ci-mirror/lib/java.sh — Java smoke pipeline
# Runs mvn test (pom.xml) or gradle test; exits 0 silently when neither is present.
set -eu
log() { printf '[ci-mirror][java] %s\n' "$1" >&2; }
REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"
if command -v mvn >/dev/null 2>&1 && [ -f "$REPO_ROOT/pom.xml" ]; then
    log "running mvn -q test"
    mvn -q test || exit 2
elif command -v gradle >/dev/null 2>&1; then
    log "running gradle test --quiet"
    gradle test --quiet || exit 2
else
    log "mvn/gradle not installed — skipping"
    exit 0
fi
log "OK"
exit 0
