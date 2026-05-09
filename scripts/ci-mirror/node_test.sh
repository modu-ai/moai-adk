#!/bin/sh
# scripts/ci-mirror/node_test.sh — Behavioral tests for lib/node.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
NODE_SH="$ROOT/scripts/ci-mirror/lib/node.sh"
PASS=0
FAIL=0

check() {
    desc="$1"; result="$2"
    if [ "$result" = "ok" ]; then printf '[PASS] %s\n' "$desc"; PASS=$((PASS + 1))
    else printf '[FAIL] %s\n' "$desc"; FAIL=$((FAIL + 1)); fi
}

orig_path="$PATH"
cleanup_dirs=""
cleanup() { rm -rf $cleanup_dirs; PATH="$orig_path"; }
trap cleanup EXIT

# Test 1: No npm → exit 0 silently
t1_dir="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t1_dir"
PATH="$t1_dir:/bin:/usr/bin" REPO_ROOT="$t1_dir" /bin/sh "$NODE_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then check "npm absent: exit 0" ok
else check "npm absent: exit 0" fail; printf '  exit_code=%d\n' "$rc" >&2; fi

# Test 2: Stub npm exits 0 → module exits 0
t2_bin="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t2_bin"
printf '#!/bin/sh\nexit 0\n' > "$t2_bin/npm"; chmod +x "$t2_bin/npm"
PATH="$t2_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$NODE_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then check "stub npm exit 0: module exits 0" ok
else check "stub npm exit 0: module exits 0" fail; fi

# Test 3: Stub npm exits 1 → module exits 2
t3_bin="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t3_bin"
printf '#!/bin/sh\nexit 1\n' > "$t3_bin/npm"; chmod +x "$t3_bin/npm"
PATH="$t3_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$NODE_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 2 ]; then check "stub npm exit 1: module exits 2" ok
else check "stub npm exit 1: module exits 2" fail; printf '  expected 2, got %d\n' "$rc" >&2; fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
