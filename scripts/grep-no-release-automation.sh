#!/bin/sh
# scripts/grep-no-release-automation.sh — Asserts no release/tag automation
# in pre-push hook and ci-mirror scripts (AC-CIAUT-013).
# Exit 0 = clean. Exit 1 = found prohibited pattern.
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
patterns='git tag\|gh release\|goreleaser'
HITS=0

# Files to check.
STATIC_FILES="$ROOT/internal/template/templates/.git_hooks/pre-push \
$ROOT/scripts/ci-mirror/run.sh \
$ROOT/scripts/ci-mirror/cross-compile.sh"

DYNAMIC_FILES=""
for f in "$ROOT/scripts/ci-mirror/lib/"*.sh \
         "$ROOT/internal/template/templates/scripts/ci-mirror/lib/"*.sh; do
    [ -f "$f" ] && DYNAMIC_FILES="$DYNAMIC_FILES $f"
done

for f in $STATIC_FILES $DYNAMIC_FILES; do
    [ -f "$f" ] || continue
    if grep -E "$patterns" "$f" 2>/dev/null; then
        printf '[grep-no-release-automation] FAIL: found release pattern in %s\n' "$f" >&2
        HITS=$((HITS + 1))
    fi
done

if [ "$HITS" -ne 0 ]; then
    printf '[grep-no-release-automation] FAIL: %d file(s) contain prohibited release automation\n' "$HITS" >&2
    exit 1
fi

printf '[grep-no-release-automation] OK: zero release automation patterns found\n' >&2
exit 0
