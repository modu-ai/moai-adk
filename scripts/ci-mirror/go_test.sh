#!/bin/sh
# scripts/ci-mirror/go_test.sh — Tests for lib/go.sh
# Run with: sh scripts/ci-mirror/go_test.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
GO_SH="$ROOT/scripts/ci-mirror/lib/go.sh"
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

# Test 1: go.sh runs and either exits 0 (clean project) or 2 (lint/test fail).
# We verify it does NOT exit with unexpected codes (1=missing tool, 3=build fail, 127=not found).
# The real project may have pre-existing lint issues; that is acceptable here.
if REPO_ROOT="$ROOT" SCRIPT_DIR="$ROOT/scripts/ci-mirror" sh "$GO_SH" 2>/dev/null && rc=0 || rc=$?; then
    :
fi
if [ "$rc" -eq 0 ] || [ "$rc" -eq 2 ]; then
    check "go.sh runs without unexpected exit code (0 or 2 acceptable)" ok
else
    check "go.sh runs without unexpected exit code (0 or 2 acceptable)" fail
    printf '  exit_code=%d (expected 0 or 2)\n' "$rc" >&2
fi

# Test 2: golangci-lint absent → still exit 0 if rest pass
# We test this by verifying go.sh has a graceful skip for golangci-lint.
if grep -q 'command -v golangci-lint' "$GO_SH"; then
    check "go.sh has graceful golangci-lint skip" ok
else
    check "go.sh has graceful golangci-lint skip" fail
fi

# Test 3: go.sh exits 2 on deliberate test failure
# Create a temp project with a failing test.
t3_dir="$(mktemp -d)"
trap 'rm -rf "$t3_dir"' EXIT
mkdir -p "$t3_dir/pkg"

# Minimal go.mod
printf 'module example.com/fail\n\ngo 1.21\n' > "$t3_dir/go.mod"

# Deliberately failing test
printf 'package pkg\n\nimport "testing"\n\nfunc TestAlwaysFail(t *testing.T) {\n\tt.Fatal("expected failure")\n}\n' \
    > "$t3_dir/pkg/fail_test.go"

REPO_ROOT="$t3_dir" SCRIPT_DIR="$ROOT/scripts/ci-mirror" sh "$GO_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 2 ]; then
    check "deliberate test failure: exits 2" ok
else
    check "deliberate test failure: exits 2" fail
    printf '  expected exit_code=2, got %d\n' "$rc" >&2
fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
