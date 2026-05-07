#!/bin/sh
# scripts/ci-autofix/log-fetch.sh
# POSIX sh — Fetches failed CI run log + PR diff for the Wave 3 auto-fix loop.
#
# Usage: sh log-fetch.sh <run-id> <pr-number>
#   run-id    GitHub Actions workflow run ID (numeric string)
#   pr-number GitHub PR number (numeric, > 0)
#
# Output: combined log + diff written to stdout.
#   - Failed log: first 200KB from `gh run view <run-id> --log-failed`
#   - PR diff:    output of `gh pr diff <pr-number>`
#   If log exceeds 200KB, warning is emitted to stderr.
#   Exit 0 on success; exit 1 on fatal error (auth fail, empty run-id, etc.).
#
# Mock injection: MOAI_AUTOFIX_GH= env var overrides the 'gh' binary path.
#   Example: MOAI_AUTOFIX_GH=/path/to/mock_gh.sh sh log-fetch.sh 123 456
#
# Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 3 (W3-T03)

# ── Constants ─────────────────────────────────────────────────────────────────

# Maximum bytes to capture from a single CI run log (200KB).
readonly LOG_CAP_BYTES=200000

# Sentinel message emitted to stderr when log is truncated.
readonly LOG_TRUNCATED_MSG="[ci-autofix] log truncated to 200KB; manual review may be needed"

# Error prefix for fatal messages.
readonly ERR_PREFIX="[ci-autofix][FATAL]"

# ── Resolve gh binary ─────────────────────────────────────────────────────────

GH="${MOAI_AUTOFIX_GH:-gh}"

# ── Input validation ──────────────────────────────────────────────────────────

run_id="$1"
pr_number="$2"

if [ -z "$run_id" ]; then
    printf '%s run-id is required (first argument)\n' "$ERR_PREFIX" >&2
    exit 1
fi

if [ -z "$pr_number" ]; then
    printf '%s pr-number is required (second argument)\n' "$ERR_PREFIX" >&2
    exit 1
fi

# ── Fetch failed log ──────────────────────────────────────────────────────────

# Capture the failed log output; truncate at LOG_CAP_BYTES bytes.
log_output=""
log_fetch_exit=0
log_output=$("$GH" run view "$run_id" --log-failed 2>/dev/null) || log_fetch_exit=$?

if [ "$log_fetch_exit" -ne 0 ]; then
    printf '%s gh run view failed (exit %s). Check: gh auth status\n' \
        "$ERR_PREFIX" "$log_fetch_exit" >&2
    exit 1
fi

# Check byte length and truncate if needed.
log_len=$(printf '%s' "$log_output" | wc -c | tr -d ' ')
if [ "$log_len" -gt "$LOG_CAP_BYTES" ]; then
    log_output=$(printf '%s' "$log_output" | head -c "$LOG_CAP_BYTES")
    printf '%s\n' "$LOG_TRUNCATED_MSG" >&2
fi

# ── Fetch PR diff ─────────────────────────────────────────────────────────────

diff_output=""
diff_fetch_exit=0
diff_output=$("$GH" pr diff "$pr_number" 2>/dev/null) || diff_fetch_exit=$?

# PR diff failure is non-fatal: log and continue with empty diff.
if [ "$diff_fetch_exit" -ne 0 ]; then
    printf '[ci-autofix][WARN] gh pr diff failed (exit %s); continuing without diff\n' \
        "$diff_fetch_exit" >&2
    diff_output=""
fi

# ── Emit combined output ──────────────────────────────────────────────────────

printf '=== CI RUN LOG (run-id: %s) ===\n' "$run_id"
printf '%s\n' "$log_output"
printf '\n=== PR DIFF (pr: #%s) ===\n' "$pr_number"
printf '%s\n' "$diff_output"

exit 0
