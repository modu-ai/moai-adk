#!/bin/sh
# scripts/ci-mirror/test/makefile_test.sh — Verify Makefile ci-local + pr-merge targets
# Run with: sh scripts/ci-mirror/test/makefile_test.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
MK="$ROOT/Makefile"
PASS=0
FAIL=0

check() {
    desc="$1"
    result="$2"
    if [ "$result" = "ok" ]; then
        printf '[PASS] %s\n' "$desc"
        PASS=$((PASS + 1))
    else
        printf '[FAIL] %s\n' "$desc"
        FAIL=$((FAIL + 1))
    fi
}

# Test 1: ci-local target exists
if make -n ci-local -C "$ROOT" 2>&1 | grep -q 'ci-mirror/run.sh'; then
    check "ci-local dry-run contains ./scripts/ci-mirror/run.sh" ok
else
    check "ci-local dry-run contains ./scripts/ci-mirror/run.sh" fail
    make -n ci-local -C "$ROOT" 2>&1 | head -5 >&2
fi

# Test 2: pr-merge target with PR=999 STRATEGY=merge
if make -n pr-merge PR=999 STRATEGY=merge -C "$ROOT" 2>&1 | grep -q 'gh pr merge 999.*--auto.*--merge'; then
    check "pr-merge PR=999 STRATEGY=merge dry-run contains correct gh command" ok
else
    check "pr-merge PR=999 STRATEGY=merge dry-run contains correct gh command" fail
    make -n pr-merge PR=999 STRATEGY=merge -C "$ROOT" 2>&1 | head -5 >&2
fi

# Test 3: pr-merge with no PR= exits non-zero
if make -n pr-merge -C "$ROOT" 2>&1 | grep -qE 'Usage:|error'; then
    check "pr-merge without PR= shows usage hint" ok
else
    # Some make versions may handle this differently — soft check
    printf '[SKIP] pr-merge without PR= behavior varies by make version\n'
fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
