#!/bin/sh
# scripts/ci-mirror/python_test.sh — Behavioral tests for lib/python.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
PYTHON_SH="$ROOT/scripts/ci-mirror/lib/python.sh"
PASS=0
FAIL=0

check() {
    desc="$1"; result="$2"
    if [ "$result" = "ok" ]; then printf '[PASS] %s\n' "$desc"; PASS=$((PASS + 1))
    else printf '[FAIL] %s\n' "$desc"; FAIL=$((FAIL + 1)); fi
}

# Test 1: No pytest → exit 0 silently
orig_path="$PATH"
t1_dir="$(mktemp -d)"
trap 'rm -rf "$t1_dir"; PATH="$orig_path"' EXIT
# Override PATH to a dir with no pytest (use /bin/sh explicitly to avoid PATH lookup)
PATH="$t1_dir:/bin:/usr/bin" REPO_ROOT="$t1_dir" /bin/sh "$PYTHON_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then
    check "pytest absent: exit 0" ok
else
    check "pytest absent: exit 0" fail
    printf '  exit_code=%d\n' "$rc" >&2
fi

# Test 2: Stub pytest that exits 0 → module exits 0
t2_bin="$(mktemp -d)"
printf '#!/bin/sh\nexit 0\n' > "$t2_bin/pytest"
chmod +x "$t2_bin/pytest"
PATH="$t2_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$PYTHON_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then
    check "stub pytest exit 0: module exits 0" ok
else
    check "stub pytest exit 0: module exits 0" fail
fi

# Test 3: Stub pytest that exits 1 → module exits 2
t3_bin="$(mktemp -d)"
printf '#!/bin/sh\nexit 1\n' > "$t3_bin/pytest"
chmod +x "$t3_bin/pytest"
PATH="$t3_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$PYTHON_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 2 ]; then
    check "stub pytest exit 1: module exits 2" ok
else
    check "stub pytest exit 1: module exits 2" fail
    printf '  expected 2, got %d\n' "$rc" >&2
fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
