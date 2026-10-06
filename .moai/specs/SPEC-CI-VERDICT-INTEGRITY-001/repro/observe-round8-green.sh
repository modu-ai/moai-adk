#!/bin/sh
# t1534 round-8 GREEN observation: REPAIRED copies on the four findings
# plus the whole existing probe suite. flow-contexts runs BOTH ways: yq
# absent -> loud abort 1; yq present -> the Lint=fail is scored exit 2.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)"
# GATE-9: observation log at an unpredictable mktemp path + cleanup trap
# (the three watch runs reuse it sequentially — each redirects fresh).
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -f "$LOG"' EXIT INT TERM
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
WATCH="$REPO_ROOT/scripts/ci-watch/run.sh"
NOYQ_PATH="/usr/bin:/bin"
cd "$HERE/flow-contexts"
PATH="$NOYQ_PATH" MOAI_CIWATCH_GH="$HERE/stub-main-lint-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "flow-contexts GREEN no-yq (expect 1)=$?"
grep -c "yq not found\|no key for base branch" "$LOG"
MOAI_CIWATCH_GH="$HERE/stub-main-lint-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "flow-contexts GREEN with-yq (expect 2 = Lint fail scored)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-bad-link/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "bad-link GREEN exit(expect 2)=$?"
grep -o "http://example.invalid/[a-z]*" "$LOG" | head -1
cd "$HERE/name-comment" && sh "$VALIDATOR" >/dev/null 2>&1
echo "name-comment GREEN(expect 0)=$?"
cd "$HERE/no-space-matrix" && sh "$VALIDATOR" >/dev/null 2>&1
echo "no-space-matrix GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
