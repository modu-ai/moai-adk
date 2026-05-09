#!/bin/sh
# scripts/ci-mirror/lib/go.sh — Go pipeline (vet + lint + test -race + cross-compile)
# Exit codes: 0 success, 2 lint/test failure, 3 build failure.
set -eu

log() { printf '[ci-mirror][go] %s\n' "$1" >&2; }

REPO_ROOT="${REPO_ROOT:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$REPO_ROOT"

log "step 1/4: go vet"
go vet ./... || exit 2

log "step 2/4: golangci-lint"
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run ./... || exit 2
else
    log "golangci-lint not installed — skipping"
fi

log "step 3/4: go test -race -short"
go test -race -count=1 -short ./... || exit 2

log "step 4/4: cross-compile"
CROSS="$SCRIPT_DIR/cross-compile.sh"
if [ -x "$CROSS" ]; then
    sh "$CROSS" || exit 3
else
    sh ./scripts/ci-mirror/cross-compile.sh || exit 3
fi

log "go pipeline OK"
exit 0
