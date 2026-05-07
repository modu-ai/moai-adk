#!/bin/sh
# scripts/ci-autofix/classify.sh
# POSIX sh — CI failure log classifier for the Wave 3 auto-fix loop.
# Reads CI log lines from stdin, emits classification verdict to stdout.
#
# Output format (two lines, always):
#   classification=<mechanical|semantic|unknown>
#   sub_class=<trivial|non-trivial|none>
#
# Ordering: semantic patterns checked first (conservative — prevents a log
# with both lint warning AND race condition being classified as mechanical).
# Unknown → treated as semantic by the orchestrator (escalate immediately).
#
# Mock injection: MOAI_AUTOFIX_FIXTURE_DIR=/path/to/fixtures reads from
# fixture files instead of stdin (for test isolation).
#
# Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 3 (W3-T02)
# Language-neutral: regex matches CI log text only, no AST dependency.

# ── Pattern constants (all readonly, no inline literals in logic) ─────────────

# Trivial mechanical patterns: whitespace / formatting fixes auto-applicable.
readonly RX_TRIVIAL_WHITESPACE='trailing whitespace|file does not end with newline|no newline at end'
readonly RX_TRIVIAL_GOFMT='gofmt|goimports needs'
readonly RX_TRIVIAL_GOIMPORTS='goimports needs'
readonly RX_TRIVIAL_IMPORT_ORDER='import.order|import-order|imports not sorted'

# Non-trivial mechanical patterns: lint/errcheck — need careful patch.
readonly RX_MECH_ERRCHECK='errcheck|[Ee]rr.*not checked|Error return value.*not checked|unused (variable|import)'
readonly RX_MECH_LINT_ST='staticcheck.*(ST1005|QF1003|SA[0-9]+)'
readonly RX_MECH_TYPO_IMPORT='undeclared name|undefined: |missing import|cannot find package'

# Semantic patterns: race conditions, panics, assertion failures — no auto-patch.
readonly RX_SEMANTIC_RACE='data race|race detected|--- FAIL: TestRace'
readonly RX_SEMANTIC_PANIC='panic: |goroutine [0-9]'
# Match both: "--- FAIL: Test<UpperCase>" (any test function failure line)
# AND "expected ... got" style assertion output appearing anywhere in the log.
# These are checked as two independent grep calls because multi-line matching
# in POSIX sh grep is unreliable across platforms.
readonly RX_SEMANTIC_ASSERT='--- FAIL: Test[A-Z]|FAIL\tTest[A-Z]|expected .* got |assertion (failed|error)'
readonly RX_SEMANTIC_DEADLOCK='fatal error: all goroutines are asleep|deadlock'

# ── Read input ────────────────────────────────────────────────────────────────

if [ -n "$MOAI_AUTOFIX_FIXTURE_DIR" ] && [ -d "$MOAI_AUTOFIX_FIXTURE_DIR" ]; then
    log_content="$(cat "$MOAI_AUTOFIX_FIXTURE_DIR/"*.log 2>/dev/null || true)"
else
    log_content="$(cat)"
fi

# ── Classification logic (semantic first, then mechanical, else unknown) ──────

# Helper: test if log_content matches a pattern via grep -E.
# Uses -e flag explicitly so patterns starting with '--' are not misinterpreted
# as options on BSD grep (macOS, FreeBSD).
_matches() {
    printf '%s\n' "$log_content" | grep -qE -e "$1" 2>/dev/null
}

# 1. Semantic patterns (highest priority — escalate immediately, no patch).
if _matches "$RX_SEMANTIC_RACE" || \
   _matches "$RX_SEMANTIC_PANIC" || \
   _matches "$RX_SEMANTIC_ASSERT" || \
   _matches "$RX_SEMANTIC_DEADLOCK"; then
    printf 'classification=semantic\n'
    printf 'sub_class=none\n'
    exit 0
fi

# 2. Trivial mechanical patterns (whitespace/formatting — silent apply after iter 1).
if _matches "$RX_TRIVIAL_WHITESPACE" || \
   _matches "$RX_TRIVIAL_GOFMT" || \
   _matches "$RX_TRIVIAL_GOIMPORTS" || \
   _matches "$RX_TRIVIAL_IMPORT_ORDER"; then
    printf 'classification=mechanical\n'
    printf 'sub_class=trivial\n'
    exit 0
fi

# 3. Non-trivial mechanical patterns (lint/errcheck/import — confirm before apply).
if _matches "$RX_MECH_ERRCHECK" || \
   _matches "$RX_MECH_LINT_ST" || \
   _matches "$RX_MECH_TYPO_IMPORT"; then
    printf 'classification=mechanical\n'
    printf 'sub_class=non-trivial\n'
    exit 0
fi

# 4. Unknown — orchestrator treats as semantic (escalate).
printf 'classification=unknown\n'
printf 'sub_class=none\n'
exit 0
