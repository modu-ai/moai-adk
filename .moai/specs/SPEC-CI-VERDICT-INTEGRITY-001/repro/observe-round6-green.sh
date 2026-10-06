#!/bin/sh
# t1534 round-6 GREEN observation: the REPAIRED copies on the four findings
# (expect 0/0/0/1-abort) plus the whole existing validator probe suite.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)"
# GATE-9: observation log at an unpredictable mktemp path + cleanup trap.
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -f "$LOG"' EXIT INT TERM
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
WATCH="$REPO_ROOT/scripts/ci-watch/run.sh"
cd "$HERE/matrix-block" && sh "$VALIDATOR" >/dev/null 2>&1
echo "block GREEN(expect 0)=$?"
cd "$HERE/matrix-merge" && sh "$VALIDATOR" >/dev/null 2>&1
echo "merge GREEN(expect 0)=$?"
cd "$HERE/job-id" && sh "$VALIDATOR" >/dev/null 2>&1
echo "job-id GREEN(expect 0)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-base-develop/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "unkeyed-base GREEN(expect 1)=$?"
grep -c "no key for base branch" "$LOG"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -2
exit 0
