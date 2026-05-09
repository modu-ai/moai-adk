#!/bin/sh
# scripts/ci-mirror/run_test.sh — Tests for run.sh dispatcher
# Run with: sh scripts/ci-mirror/run_test.sh
set -eu

ROOT="${1:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
RUN_SH="$ROOT/scripts/ci-mirror/run.sh"
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

# --- Test 1: No language markers → exit 0, stderr contains "skipping" ---
t1_dir="$(mktemp -d)"
trap 'rm -rf "$t1_dir"' EXIT
# No go.mod, package.json, etc. in t1_dir
err_out="$t1_dir/err.txt"
REPO_ROOT="$t1_dir" sh "$RUN_SH" 2>"$err_out" && rc=0 || rc=$?
if [ "$rc" -eq 0 ] && grep -q "skipping" "$err_out" 2>/dev/null; then
    check "no markers: exit 0 + skipping message" ok
else
    check "no markers: exit 0 + skipping message" fail
    printf '  exit_code=%d stderr=%s\n' "$rc" "$(cat "$err_out" 2>/dev/null)" >&2
fi

# --- Test 2: go.mod present + stub lib/go.sh → dispatched, marker file created ---
t2_dir="$(mktemp -d)"
t2_lib="$t2_dir/lib"
mkdir -p "$t2_lib"
touch "$t2_dir/go.mod"
marker="$t2_dir/go_module_ran"
# Stub go.sh that creates a marker file.
printf '#!/bin/sh\ntouch "%s"\n' "$marker" > "$t2_lib/go.sh"
chmod +x "$t2_lib/go.sh"
REPO_ROOT="$t2_dir" MOAI_CI_LIB_DIR="$t2_lib" sh "$RUN_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 0 ] && [ -f "$marker" ]; then
    check "go.mod: dispatches go module + marker created" ok
else
    check "go.mod: dispatches go module + marker created" fail
    printf '  exit_code=%d marker_exists=%s\n' "$rc" "$([ -f "$marker" ] && printf yes || printf no)" >&2
fi

# --- Test 3: stub lib/go.sh exits 2 → run.sh propagates exit 2 ---
t3_dir="$(mktemp -d)"
t3_lib="$t3_dir/lib"
mkdir -p "$t3_lib"
touch "$t3_dir/go.mod"
printf '#!/bin/sh\nexit 2\n' > "$t3_lib/go.sh"
chmod +x "$t3_lib/go.sh"
REPO_ROOT="$t3_dir" MOAI_CI_LIB_DIR="$t3_lib" sh "$RUN_SH" 2>/dev/null && rc=0 || rc=$?
if [ "$rc" -eq 2 ]; then
    check "stub exits 2: exit code propagated" ok
else
    check "stub exits 2: exit code propagated" fail
    printf '  expected exit_code=2, got %d\n' "$rc" >&2
fi

printf '\n%d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
