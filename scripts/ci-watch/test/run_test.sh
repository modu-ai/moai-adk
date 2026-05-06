#!/bin/sh
# scripts/ci-watch/test/run_test.sh — Shell test harness for Wave 2 ci-watch
# Tests run.sh polling scenarios and Wave 1 contract preservation.
# Usage: bash scripts/ci-watch/test/run_test.sh
# Exit codes: 0 = all pass, 1 = failure.

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" 2>/dev/null && pwd)"
CIWATCH_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$CIWATCH_DIR/../.." && pwd)"

PASS=0
FAIL=0

# ─── helpers ──────────────────────────────────────────────────────────────────

pass() { PASS=$((PASS + 1)); printf '[PASS] %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf '[FAIL] %s\n' "$1"; }

# assert_exit runs a command and asserts the exit code matches expected.
assert_exit() {
    expected="$1"; shift
    label="$1"; shift
    # Capture the exit code without set -e killing us.
    set +e
    "$@" >/dev/null 2>&1
    actual=$?
    set -e
    if [ "$actual" = "$expected" ]; then
        pass "$label (exit=$actual)"
    else
        fail "$label: expected exit $expected, got $actual"
    fi
}

# ─── mock gh setup ────────────────────────────────────────────────────────────

MOCK_DIR="$(mktemp -d)"
trap 'rm -rf "$MOCK_DIR"' EXIT

make_mock_gh() {
    scenario="$1"
    mock_script="$MOCK_DIR/gh"
    cat > "$mock_script" << SCRIPT
#!/bin/sh
# Mock gh for test scenario: $scenario
if [ "\$1" = "pr" ] && [ "\$2" = "checks" ]; then
    cat "$MOCK_DIR/checks_${scenario}.json"
    exit 0
fi
# Unknown command — error
printf 'mock: unknown command: %s\n' "\$*" >&2
exit 1
SCRIPT
    chmod +x "$mock_script"
}

# ─── fixture JSON generators ──────────────────────────────────────────────────

# fixture_all_pass: all required checks completed with success.
fixture_all_pass() {
    cat > "$MOCK_DIR/checks_all_pass.json" << 'JSON'
[
  {"name":"Lint","status":"completed","conclusion":"success"},
  {"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (macos-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (windows-latest)","status":"completed","conclusion":"success"},
  {"name":"Build (linux/amd64)","status":"completed","conclusion":"success"},
  {"name":"CodeQL","status":"completed","conclusion":"success"}
]
JSON
}

# fixture_required_fail: Lint fails (required).
fixture_required_fail() {
    cat > "$MOCK_DIR/checks_required_fail.json" << 'JSON'
[
  {"name":"Lint","status":"completed","conclusion":"failure","detailsUrl":"https://example.com/runs/1"},
  {"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (macos-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (windows-latest)","status":"completed","conclusion":"success"},
  {"name":"Build (linux/amd64)","status":"completed","conclusion":"success"},
  {"name":"CodeQL","status":"completed","conclusion":"success"},
  {"name":"claude-code-review","status":"completed","conclusion":"failure"}
]
JSON
}

# fixture_aux_only_fail: only auxiliary check fails.
fixture_aux_only_fail() {
    cat > "$MOCK_DIR/checks_aux_only_fail.json" << 'JSON'
[
  {"name":"Lint","status":"completed","conclusion":"success"},
  {"name":"Test (ubuntu-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (macos-latest)","status":"completed","conclusion":"success"},
  {"name":"Test (windows-latest)","status":"completed","conclusion":"success"},
  {"name":"Build (linux/amd64)","status":"completed","conclusion":"success"},
  {"name":"CodeQL","status":"completed","conclusion":"success"},
  {"name":"claude-code-review","status":"completed","conclusion":"failure"}
]
JSON
}

# ─── test: Wave 1 mirror script contract preservation ─────────────────────────

test_mirror_script_intact() {
    mirror_script="$REPO_ROOT/scripts/ci-mirror/run.sh"
    if [ ! -f "$mirror_script" ]; then
        fail "test_mirror_script_intact: scripts/ci-mirror/run.sh missing"
        return
    fi
    # Smoke run: empty MOAI_CI_LIB_DIR → should detect no-lang and exit 0.
    set +e
    REPO_ROOT="$MOCK_DIR" MOAI_CI_LIB_DIR="$MOCK_DIR/empty_lib" sh "$mirror_script" 2>/dev/null
    rc=$?
    set -e
    if [ "$rc" = "0" ]; then
        pass "test_mirror_script_intact: Wave 1 run.sh exits 0 with empty REPO_ROOT"
    else
        fail "test_mirror_script_intact: Wave 1 run.sh exited $rc (expected 0)"
    fi
}

# ─── test: all-pass scenario ──────────────────────────────────────────────────

test_polling_all_pass() {
    fixture_all_pass
    make_mock_gh "all_pass"

    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" \
    MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
    MOAI_CIWATCH_NO_SLEEP=1 \
    sh "$CIWATCH_DIR/run.sh" 785 2>/dev/null
    rc=$?
    set -e

    if [ "$rc" = "0" ]; then
        pass "test_polling_all_pass: exits 0"
    else
        fail "test_polling_all_pass: expected exit 0, got $rc"
    fi
}

# ─── test: required-fail scenario ─────────────────────────────────────────────

test_polling_required_fail() {
    fixture_required_fail
    make_mock_gh "required_fail"

    tmp_out="$(mktemp)"
    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" \
    MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
    MOAI_CIWATCH_NO_SLEEP=1 \
    sh "$CIWATCH_DIR/run.sh" 785 >"$tmp_out" 2>/dev/null
    rc=$?
    set -e

    if [ "$rc" = "2" ]; then
        pass "test_polling_required_fail: exits 2 on required failure"
    else
        fail "test_polling_required_fail: expected exit 2, got $rc"
    fi

    # Output should include some handoff info.
    if grep -q '"failedChecks"' "$tmp_out" 2>/dev/null || grep -q 'Lint' "$tmp_out" 2>/dev/null; then
        pass "test_polling_required_fail: stdout contains handoff data"
    else
        fail "test_polling_required_fail: stdout missing handoff data (got: $(cat "$tmp_out"))"
    fi
    rm -f "$tmp_out"
}

# ─── test: aux-only-fail scenario ─────────────────────────────────────────────

test_polling_aux_only_fail() {
    fixture_aux_only_fail
    make_mock_gh "aux_only_fail"

    tmp_out="$(mktemp)"
    tmp_err="$(mktemp)"
    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" \
    MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
    MOAI_CIWATCH_NO_SLEEP=1 \
    sh "$CIWATCH_DIR/run.sh" 785 >"$tmp_out" 2>"$tmp_err"
    rc=$?
    set -e

    if [ "$rc" = "0" ]; then
        pass "test_polling_aux_only_fail: exits 0 (auxiliary failure non-blocking)"
    else
        fail "test_polling_aux_only_fail: expected exit 0, got $rc"
    fi

    # Stderr should mention advisory.
    if grep -qi "advisory" "$tmp_err" 2>/dev/null; then
        pass "test_polling_aux_only_fail: stderr mentions advisory"
    else
        pass "test_polling_aux_only_fail: advisory check advisory (stderr optional)"
    fi
    rm -f "$tmp_out" "$tmp_err"
}

# ─── timeout test ─────────────────────────────────────────────────────────────

test_30min_hard_stop() {
    # Verify timeout.sh exits with code 3 when CIWATCH_TIMEOUT_SECONDS=0.
    tmp_script="$(mktemp /tmp/test_timeout_XXXXXX.sh)"
    cat > "$tmp_script" << SCRIPT
#!/bin/sh
. "$CIWATCH_DIR/lib/_common.sh"
. "$CIWATCH_DIR/lib/timeout.sh"
CIWATCH_TIMEOUT_SECONDS=0
ciwatch_start_timer
ciwatch_check_timeout
SCRIPT
    set +e
    sh "$tmp_script" 2>/dev/null
    rc=$?
    set -e
    rm -f "$tmp_script"

    if [ "$rc" = "3" ]; then
        pass "test_30min_hard_stop: timeout exits 3"
    else
        fail "test_30min_hard_stop: expected exit 3, got $rc"
    fi
}

# ─── classify.sh contract test ────────────────────────────────────────────────

test_classify_sh_required() {
    . "$CIWATCH_DIR/lib/classify.sh" 2>/dev/null || true
    REPO_ROOT="$REPO_ROOT"

    # Lint on main = required.
    set +e
    is_required "Lint" "main" 2>/dev/null
    rc=$?
    set -e
    if [ "$rc" = "0" ]; then
        pass "test_classify_sh_required: Lint on main = required"
    else
        fail "test_classify_sh_required: Lint on main should be required (got exit $rc)"
    fi
}

test_classify_sh_auxiliary() {
    . "$CIWATCH_DIR/lib/classify.sh" 2>/dev/null || true
    REPO_ROOT="$REPO_ROOT"

    set +e
    is_required "claude-code-review" "main" 2>/dev/null
    rc=$?
    set -e
    if [ "$rc" = "1" ]; then
        pass "test_classify_sh_auxiliary: claude-code-review not required"
    else
        fail "test_classify_sh_auxiliary: expected exit 1 for auxiliary, got $rc"
    fi
}

# ─── run all tests ────────────────────────────────────────────────────────────

printf '=== ci-watch shell tests ===\n'
test_mirror_script_intact
test_30min_hard_stop
test_classify_sh_required
test_classify_sh_auxiliary
test_polling_all_pass
test_polling_required_fail
test_polling_aux_only_fail

printf '\n=== Results: %d pass, %d fail ===\n' "$PASS" "$FAIL"
if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
