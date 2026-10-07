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
    checks_exit="${2:-0}"
    mock_script="$MOCK_DIR/gh"
    cat > "$mock_script" << SCRIPT
#!/bin/sh
# Mock gh for test scenario: $scenario
# GATE-6: the watch resolves the PR base branch first — serve main.
if [ "\$1" = "pr" ] && [ "\$2" = "view" ]; then
    printf '%s\n' 'main'
    exit 0
fi
if [ "\$1" = "pr" ] && [ "\$2" = "checks" ]; then
    cat "$MOCK_DIR/checks_${scenario}.json"
    exit $checks_exit
fi
# Unknown command — error
printf 'mock: unknown command: %s\n' "\$*" >&2
exit 1
SCRIPT
    chmod +x "$mock_script"
}

# ─── fixture JSON generators ──────────────────────────────────────────────────
# t1534 M3: fixtures carry the REAL `gh pr checks --json` shape — the
# supported field set is exactly name/state/bucket/link (the pre-M3 fixtures
# used status/conclusion/detailsUrl, the field list that made every real
# poll abort with "Unknown JSON field: 'status'").

# fixture_all_pass: all nine SSoT main contexts bucket=pass.
fixture_all_pass() {
    cat > "$MOCK_DIR/checks_all_pass.json" << 'JSON'
[
  {"name":"Lint","state":"SUCCESS","bucket":"pass","link":"https://example.com/lint"},
  {"name":"Test (ubuntu-latest)","state":"SUCCESS","bucket":"pass","link":"https://example.com/tu"},
  {"name":"Build (linux/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b1"},
  {"name":"Build (linux/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b2"},
  {"name":"Build (darwin/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b3"},
  {"name":"Build (darwin/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b4"},
  {"name":"Build (windows/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b5"},
  {"name":"Analyze (Go) (go)","state":"SUCCESS","bucket":"pass","link":"https://example.com/cq"},
  {"name":"Release PR Multi-OS Gate","state":"SUCCESS","bucket":"pass","link":"https://example.com/mos"}
]
JSON
}

# fixture_required_fail: Lint bucket=fail (required), the rest pass, plus an
# auxiliary failure (claude-code-review) that must stay non-blocking.
fixture_required_fail() {
    cat > "$MOCK_DIR/checks_required_fail.json" << 'JSON'
[
  {"name":"Lint","state":"FAILURE","bucket":"fail","link":"https://example.com/lint-fail"},
  {"name":"Test (ubuntu-latest)","state":"SUCCESS","bucket":"pass","link":"https://example.com/tu"},
  {"name":"Build (linux/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b1"},
  {"name":"Build (linux/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b2"},
  {"name":"Build (darwin/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b3"},
  {"name":"Build (darwin/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b4"},
  {"name":"Build (windows/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b5"},
  {"name":"Analyze (Go) (go)","state":"SUCCESS","bucket":"pass","link":"https://example.com/cq"},
  {"name":"Release PR Multi-OS Gate","state":"SUCCESS","bucket":"pass","link":"https://example.com/mos"},
  {"name":"claude-code-review","state":"FAILURE","bucket":"fail","link":"https://example.com/ccr"}
]
JSON
}

# fixture_aux_only_fail: all nine required pass, only the auxiliary fails.
fixture_aux_only_fail() {
    cat > "$MOCK_DIR/checks_aux_only_fail.json" << 'JSON'
[
  {"name":"Lint","state":"SUCCESS","bucket":"pass","link":"https://example.com/lint"},
  {"name":"Test (ubuntu-latest)","state":"SUCCESS","bucket":"pass","link":"https://example.com/tu"},
  {"name":"Build (linux/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b1"},
  {"name":"Build (linux/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b2"},
  {"name":"Build (darwin/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b3"},
  {"name":"Build (darwin/arm64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b4"},
  {"name":"Build (windows/amd64)","state":"SUCCESS","bucket":"pass","link":"https://example.com/b5"},
  {"name":"Analyze (Go) (go)","state":"SUCCESS","bucket":"pass","link":"https://example.com/cq"},
  {"name":"Release PR Multi-OS Gate","state":"SUCCESS","bucket":"pass","link":"https://example.com/mos"},
  {"name":"docs-i18n-check","state":"FAILURE","bucket":"fail","link":"https://example.com/i18n"}
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

# ─── test: field-contract regression (t1534 M3) ───────────────────────────────

# A strict mock gh that SERVES the fixture only when the field list is
# exactly the supported set — any regression to the pre-M3 field list
# (status/conclusion/detailsUrl) makes gh fail "Unknown JSON field" the same
# way the real CLI does.
make_mock_gh_strict() {
    fixture="$1"
    mock_script="$MOCK_DIR/gh"
    cat > "$mock_script" << SCRIPT
#!/bin/sh
# argv: gh pr checks <PR_NUMBER> --json <fields> — \$3 is the PR number.
# GATE-6: the watch resolves the PR base branch first — serve main.
if [ "\$1" = "pr" ] && [ "\$2" = "view" ]; then
    printf '%s\n' 'main'
    exit 0
fi
if [ "\$1" = "pr" ] && [ "\$2" = "checks" ] && [ "\$4" = "--json" ] && [ "\$5" = "name,state,bucket,link" ]; then
    cat "$MOCK_DIR/$fixture"
    exit 0
fi
printf 'mock-strict: unsupported argument list: %s\n' "\$*" >&2
exit 1
SCRIPT
    chmod +x "$mock_script"
}

test_field_contract_regression() {
    # The pre-M3 field list must never return: with the strict mock, a
    # regression makes the poll abort (exit 1) instead of exit 0.
    make_mock_gh_strict "checks_all_pass.json"
    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" \
    MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
    MOAI_CIWATCH_NO_SLEEP=1 \
    sh "$CIWATCH_DIR/run.sh" 785 2>/dev/null
    rc=$?
    set -e
    if [ "$rc" = "0" ]; then
        pass "test_field_contract_regression: supported field list serves exit 0"
    else
        fail "test_field_contract_regression: expected exit 0 (field list regressed?), got $rc"
    fi
}

# ─── test: missing required check counts as pending (t1534 M3) ────────────────

test_missing_required_pending() {
    # A minimal 2-context SSoT with a response carrying only ONE of them —
    # the missing required check must count PENDING (pre-M3, iterating only
    # returned names, a not-yet-published required check silently counted
    # as passed).
    cat > "$MOCK_DIR/ssot_minimal.yml" << 'YML'
version: 1
branches:
  main:
    contexts:
      - Lint
      - "Build (linux/amd64)"
auxiliary: []
YML
    cat > "$MOCK_DIR/checks_minimal.json" << 'JSON'
[
  {"name":"Lint","state":"SUCCESS","bucket":"pass","link":"https://example.com/lint"}
]
JSON
    make_mock_gh_strict "checks_minimal.json"

    tmp_err="$(mktemp)"
    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" \
    MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$MOCK_DIR/ssot_minimal.yml" \
    MOAI_CIWATCH_NO_SLEEP=1 \
    sh "$CIWATCH_DIR/run.sh" 785 2>"$tmp_err"
    rc=$?
    set -e

    if [ "$rc" = "0" ]; then
        pass "test_missing_required_pending: NO_SLEEP single tick exits 0"
    else
        fail "test_missing_required_pending: expected exit 0, got $rc"
    fi
    if grep -q "required 1/2 pass, 1 pending" "$tmp_err" 2>/dev/null; then
        pass "test_missing_required_pending: missing required counted pending (1/2)"
    else
        fail "test_missing_required_pending: expected 'required 1/2 pass, 1 pending' in stderr (got: $(grep required "$tmp_err" | head -1))"
    fi
    rm -f "$tmp_err"
}

# A checks command can publish valid state with a documented nonzero
# verdict exit status. Transport failures must still abort.
test_checks_exit_statuses() {
    fixture_required_fail
    make_mock_gh "required_fail" 1
    assert_exit 2 "checks failure status preserves required-failure handoff" \
        env MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
        sh "$CIWATCH_DIR/run.sh" 785

    fixture_all_pass
    jq '.[0].bucket = "pending" | .[0].state = "PENDING"' \
        "$MOCK_DIR/checks_all_pass.json" > "$MOCK_DIR/checks_pending.json"
    make_mock_gh "pending" 8
    tmp_err="$(mktemp)"
    set +e
    MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
        sh "$CIWATCH_DIR/run.sh" 785 > /dev/null 2> "$tmp_err"
    rc=$?
    set -e
    if [ "$rc" = 0 ] && grep -q '1 pending' "$tmp_err"; then
        pass "checks pending status reaches pending classification"
    else
        fail "checks pending status aborted or lost pending state (exit=$rc)"
    fi
    rm -f "$tmp_err"

    printf 'authentication failed\n' > "$MOCK_DIR/checks_error.json"
    make_mock_gh "error" 1
    assert_exit 1 "checks transport failure remains fatal" \
        env MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
        sh "$CIWATCH_DIR/run.sh" 785

    printf '{}\n' > "$MOCK_DIR/checks_error.json"
    make_mock_gh "error" 0
    assert_exit 1 "checks wrong JSON shape remains fatal" \
        env MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
        sh "$CIWATCH_DIR/run.sh" 785
}

# State and bucket must agree with gh's aggregateChecks mapping.
test_check_state_bucket_contract() {
    for state in FAILURE PENDING NOT_A_STATE; do
        fixture_all_pass
        jq --arg state "$state" '.[0].state = $state' "$MOCK_DIR/checks_all_pass.json" > "$MOCK_DIR/checks_contradictory.json"
        make_mock_gh "contradictory"
        assert_exit 1 "contradictory $state/pass is rejected" \
            env MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
            MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
            sh "$CIWATCH_DIR/run.sh" 785
    done
    while read -r state bucket expected; do
        fixture_all_pass
        jq --arg state "$state" --arg bucket "$bucket" \
            '.[0].state = $state | .[0].bucket = $bucket' "$MOCK_DIR/checks_all_pass.json" > "$MOCK_DIR/checks_valid_state.json"
        make_mock_gh "valid_state"
        assert_exit "$expected" "legitimate $state/$bucket is classified" \
            env MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
            MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
            sh "$CIWATCH_DIR/run.sh" 785
    done <<'STATES'
SUCCESS pass 0
NEUTRAL skipping 0
SKIPPED skipping 0
ERROR fail 2
FAILURE fail 2
TIMED_OUT fail 2
ACTION_REQUIRED fail 2
CANCELLED cancel 2
EXPECTED pending 0
REQUESTED pending 0
WAITING pending 0
QUEUED pending 0
PENDING pending 0
IN_PROGRESS pending 0
STALE pending 0
STATES
}

# The real clock command crosses the deadline during a successful poll;
# a pre-poll-only timeout check must not authorize its late green verdict.
test_deadline_after_poll() {
    fixture_all_pass
    make_mock_gh "all_pass"
    mkdir -p "$MOCK_DIR/clock-bin"
    cat > "$MOCK_DIR/clock-bin/date" << 'CLOCK'
#!/bin/sh
n=0
[ ! -f "$CIWATCH_TEST_CLOCK" ] || n="$(cat "$CIWATCH_TEST_CLOCK")"
n=$((n + 1))
printf '%s\n' "$n" > "$CIWATCH_TEST_CLOCK"
if [ "$n" -le 2 ]; then printf '100\n'; else printf '110\n'; fi
CLOCK
    chmod +x "$MOCK_DIR/clock-bin/date"
    assert_exit 3 "deadline crossed during poll rejects late success" \
        env PATH="$MOCK_DIR/clock-bin:$PATH" CIWATCH_TEST_CLOCK="$MOCK_DIR/clock-state" \
        CIWATCH_TIMEOUT_SECONDS=10 MOAI_CIWATCH_GH="$MOCK_DIR/gh" MOAI_CIWATCH_NO_SLEEP=1 \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE="$REPO_ROOT/.github/required-checks.yml" \
        sh "$CIWATCH_DIR/run.sh" 785
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
test_field_contract_regression
test_missing_required_pending
test_checks_exit_statuses
test_check_state_bucket_contract
test_deadline_after_poll

printf '\n=== Results: %d pass, %d fail ===\n' "$PASS" "$FAIL"
if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
