#!/bin/sh
# t1534 round-7 GREEN observation: the REPAIRED copies on the round-7
# findings plus the whole existing probe suite.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)"
# GATE-9: observation log at an unpredictable mktemp path + cleanup trap.
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -f "$LOG"' EXIT INT TERM
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
WATCH="$REPO_ROOT/scripts/ci-watch/run.sh"
cd "$HERE/merge-all-legs" && sh "$VALIDATOR" >/dev/null 2>&1
echo "merge-all-legs GREEN(expect 0)=$?"
cd "$HERE/merge-overwrite" && sh "$VALIDATOR" >/dev/null 2>&1
echo "merge-overwrite GREEN(expect 0)=$?"
cd "$HERE/comment-boundary" && sh "$VALIDATOR" >/dev/null 2>&1
echo "comment-boundary GREEN(expect 0)=$?"
cd "$HERE/merge-green-gone" && sh "$VALIDATOR" >/dev/null 2>&1
echo "green-gone GREEN(expect 1 = green correctly rejected)=$?"
cd "$HERE/quoted-key"
MOAI_CIWATCH_GH="$HERE/stub-skipping/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "quoted-key watch GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/observe-round6-green.sh" 2>/dev/null | grep -v "expect" | head -3
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
