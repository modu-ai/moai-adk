#!/bin/sh
# scripts/ci-autofix/test/log_fetch_test.sh
# RED phase test harness for scripts/ci-autofix/log-fetch.sh
# 4 mock scenarios: success, gh-auth-fail, large-log-cap, missing-run-id
# Mock injection via MOAI_AUTOFIX_GH= env var pointing to a mock script.
# Run: bash scripts/ci-autofix/test/log_fetch_test.sh

set -e

LOG_FETCH_SH="${LOG_FETCH_SH:-$(cd "$(dirname "$0")/.." && pwd)/log-fetch.sh}"

# Temp dir for mock scripts and output capture.
TMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

pass_count=0
fail_count=0

_run_test() {
    test_name="$1"
    mock_gh="$2"         # path to mock gh script (or empty string to use real gh)
    run_id="$3"
    pr_number="$4"
    expected_exit="$5"   # expected exit code (0=success, 1=failure)
    expected_stdout_grep="$6"   # string that must appear in stdout (or empty)
    expected_stderr_grep="$7"   # string that must appear in stderr (or empty)

    stdout_file="$TMP_DIR/stdout_$test_name"
    stderr_file="$TMP_DIR/stderr_$test_name"

    gh_env=""
    if [ -n "$mock_gh" ]; then
        gh_env="MOAI_AUTOFIX_GH=$mock_gh"
    fi

    actual_exit=0
    # Use MOAI_AUTOFIX_GH env directly (not eval) to preserve empty-string args.
    MOAI_AUTOFIX_GH="$mock_gh" sh "$LOG_FETCH_SH" "$run_id" "$pr_number" \
        >"$stdout_file" 2>"$stderr_file" || actual_exit=$?

    ok=1

    if [ "$actual_exit" != "$expected_exit" ]; then
        ok=0
        printf 'FAIL  %s — exit code: expected %s, got %s\n' \
            "$test_name" "$expected_exit" "$actual_exit"
    fi

    if [ -n "$expected_stdout_grep" ]; then
        if ! grep -q "$expected_stdout_grep" "$stdout_file" 2>/dev/null; then
            ok=0
            printf 'FAIL  %s — stdout missing: %s\n' "$test_name" "$expected_stdout_grep"
            printf '      stdout: %s\n' "$(cat "$stdout_file")"
        fi
    fi

    if [ -n "$expected_stderr_grep" ]; then
        if ! grep -q "$expected_stderr_grep" "$stderr_file" 2>/dev/null; then
            ok=0
            printf 'FAIL  %s — stderr missing: %s\n' "$test_name" "$expected_stderr_grep"
            printf '      stderr: %s\n' "$(cat "$stderr_file")"
        fi
    fi

    if [ "$ok" = "1" ]; then
        printf 'PASS  %s\n' "$test_name"
        pass_count=$((pass_count + 1))
    else
        fail_count=$((fail_count + 1))
    fi
}

# ── Mock 1: success — gh returns log + diff ───────────────────────────────────
mock_success="$TMP_DIR/mock_gh_success.sh"
cat > "$mock_success" << 'MOCK_EOF'
#!/bin/sh
# Mock gh: success path
if [ "$1" = "run" ] && [ "$2" = "view" ]; then
    printf 'Error return value of `os.Setenv` is not checked (errcheck)\n'
    exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "diff" ]; then
    printf '+if err := os.Setenv("K","V"); err != nil { return err }\n'
    exit 0
fi
exit 1
MOCK_EOF
chmod +x "$mock_success"

_run_test "success_gh_returns_log_and_diff" \
    "$mock_success" "12345678" "739" \
    "0" "errcheck" ""

# ── Mock 2: gh-auth-fail — gh exits non-zero ─────────────────────────────────
mock_auth_fail="$TMP_DIR/mock_gh_auth_fail.sh"
cat > "$mock_auth_fail" << 'MOCK_EOF'
#!/bin/sh
# Mock gh: auth failure
printf 'Error: Not logged in. Run: gh auth login\n' >&2
exit 1
MOCK_EOF
chmod +x "$mock_auth_fail"

_run_test "gh_auth_fail_exits_nonzero" \
    "$mock_auth_fail" "12345678" "739" \
    "1" "" "gh"

# ── Mock 3: large-log-cap — >200KB output is truncated ───────────────────────
mock_large_log="$TMP_DIR/mock_gh_large.sh"
cat > "$mock_large_log" << 'MOCK_EOF'
#!/bin/sh
# Mock gh: emits more than 200KB
if [ "$1" = "run" ] && [ "$2" = "view" ]; then
    # Generate ~210KB of output (200000 chars + extras)
    python3 -c "import sys; sys.stdout.write('X' * 210000)" 2>/dev/null || \
        awk 'BEGIN{ for(i=0;i<210000;i++) printf "X"; print "" }'
    exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "diff" ]; then
    printf '+diff line\n'
    exit 0
fi
exit 0
MOCK_EOF
chmod +x "$mock_large_log"

_run_test "large_log_cap_200kb_truncation_warning" \
    "$mock_large_log" "99999999" "888" \
    "0" "" "log truncated"

# ── Mock 4: missing-run-id — gh returns empty output ─────────────────────────
mock_empty="$TMP_DIR/mock_gh_empty.sh"
cat > "$mock_empty" << 'MOCK_EOF'
#!/bin/sh
# Mock gh: returns empty (no run ID match)
if [ "$1" = "run" ] && [ "$2" = "view" ]; then
    exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "diff" ]; then
    exit 0
fi
exit 0
MOCK_EOF
chmod +x "$mock_empty"

_run_test "missing_run_id_gh_returns_empty" \
    "$mock_empty" "" "739" \
    "1" "" "run-id"

# ── Summary ──────────────────────────────────────────────────────────────────

printf '\n%s tests run: %s PASS, %s FAIL\n' \
    "$((pass_count + fail_count))" "$pass_count" "$fail_count"

if [ "$fail_count" -gt 0 ]; then
    exit 1
fi
exit 0
