#!/bin/sh
# scripts/ci-autofix/test/classify_test.sh
# RED phase test harness for scripts/ci-autofix/classify.sh
# Covers all 9 regex pattern constants: 4 trivial, 1 mech-errcheck, 1 mech-lint,
# 1 mech-import, and 3 semantic patterns (race, panic, assert/deadlock).
# Run: bash scripts/ci-autofix/test/classify_test.sh

set -e

CLASSIFY_SH="${CLASSIFY_SH:-$(cd "$(dirname "$0")/.." && pwd)/classify.sh}"

pass_count=0
fail_count=0

# _run_test: feeds $2 string as stdin to classify.sh, checks output contains
# expected classification=$3 and sub_class=$4. Prints PASS/FAIL.
_run_test() {
    test_name="$1"
    log_input="$2"
    expected_class="$3"
    expected_sub="$4"

    output=$(printf '%s\n' "$log_input" | sh "$CLASSIFY_SH" 2>/dev/null) || true

    got_class=$(printf '%s\n' "$output" | grep '^classification=' | cut -d= -f2)
    got_sub=$(printf '%s\n' "$output" | grep '^sub_class=' | cut -d= -f2)

    if [ "$got_class" = "$expected_class" ] && [ "$got_sub" = "$expected_sub" ]; then
        printf 'PASS  %s\n' "$test_name"
        pass_count=$((pass_count + 1))
    else
        printf 'FAIL  %s\n' "$test_name"
        printf '      expected: classification=%s sub_class=%s\n' "$expected_class" "$expected_sub"
        printf '      got:      classification=%s sub_class=%s\n' "$got_class" "$got_sub"
        printf '      input:    %s\n' "$log_input"
        fail_count=$((fail_count + 1))
    fi
}

# ── Semantic patterns (checked first by classifier) ──────────────────────────

_run_test "RX_SEMANTIC_RACE: data race detected" \
    "==================== FAIL: data race detected in TestFoo (0.5s)" \
    "semantic" "none"

_run_test "RX_SEMANTIC_PANIC: goroutine panic" \
    "goroutine 47 [running]:
panic: runtime error: invalid memory address or nil pointer dereference" \
    "semantic" "none"

_run_test "RX_SEMANTIC_ASSERT: test assertion failure" \
    "--- FAIL: TestBuildRequiredPATH (0.01s)
    assert.go:35: expected foo got bar" \
    "semantic" "none"

_run_test "RX_SEMANTIC_DEADLOCK: all goroutines asleep" \
    "fatal error: all goroutines are asleep - deadlock!" \
    "semantic" "none"

# ── Mechanical trivial patterns ───────────────────────────────────────────────

_run_test "RX_TRIVIAL_GOFMT: gofmt diff" \
    "internal/cache/cache.go:45: gofmt: file does not end with expected newline" \
    "mechanical" "trivial"

_run_test "RX_TRIVIAL_WHITESPACE: trailing whitespace" \
    "scripts/ci-mirror/run.sh:12: trailing whitespace found" \
    "mechanical" "trivial"

_run_test "RX_TRIVIAL_GOIMPORTS: goimports needs update" \
    "internal/template/templates/CLAUDE.md: goimports needs update" \
    "mechanical" "trivial"

# ── Mechanical non-trivial patterns ──────────────────────────────────────────

_run_test "RX_MECH_ERRCHECK: err not checked" \
    "internal/cache/cache.go:45:17: Error return value of \`os.Setenv\` is not checked (errcheck)" \
    "mechanical" "non-trivial"

_run_test "RX_MECH_LINT_ST: staticcheck ST1005" \
    "pkg/version/version.go:12:3: error strings should not be capitalized (staticcheck: ST1005)" \
    "mechanical" "non-trivial"

# ── Summary ──────────────────────────────────────────────────────────────────

printf '\n%s tests run: %s PASS, %s FAIL\n' \
    "$((pass_count + fail_count))" "$pass_count" "$fail_count"

if [ "$fail_count" -gt 0 ]; then
    exit 1
fi
exit 0
