#!/bin/sh
# scripts/ci-mirror/swift_test.sh — Behavioral tests for lib/swift.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
SWIFT_SH="$ROOT/scripts/ci-mirror/lib/swift.sh"
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

# Test 1: No swift binary → exit 0 silently (use PATH with NO swift)
t1_dir="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t1_dir"
# Only include /bin to exclude /usr/bin/swift, but still have basic POSIX commands
PATH="$t1_dir:/bin" REPO_ROOT="$t1_dir" /bin/sh "$SWIFT_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then check "swift absent: exit 0" ok
else check "swift absent: exit 0" fail; printf '  exit_code=%d\n' "$rc" >&2; fi

# Test 2: Stub swift exits 0 → module exits 0
t2_bin="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t2_bin"
printf '#!/bin/sh\nexit 0\n' > "$t2_bin/swift"; chmod +x "$t2_bin/swift"
PATH="$t2_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$SWIFT_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ]; then check "stub swift exit 0: module exits 0" ok
else check "stub swift exit 0: module exits 0" fail; fi

# Test 3: Stub swift exits 1 → module exits 2
t3_bin="$(mktemp -d)"; cleanup_dirs="$cleanup_dirs $t3_bin"
printf '#!/bin/sh\nexit 1\n' > "$t3_bin/swift"; chmod +x "$t3_bin/swift"
PATH="$t3_bin:$PATH" REPO_ROOT="$(mktemp -d)" sh "$SWIFT_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 2 ]; then check "stub swift exit 1: module exits 2" ok
else check "stub swift exit 1: module exits 2" fail; printf '  expected 2, got %d\n' "$rc" >&2; fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
