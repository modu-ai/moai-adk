#!/bin/sh
# t1534 round-14 GREEN observation: the REPAIRED validator on the three
# findings plus the whole existing probe suite.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/apos-tail" && sh "$VALIDATOR" >/dev/null 2>&1
echo "apos-tail GREEN(expect 0)=$?"
cd "$HERE/bracket-expr" && sh "$VALIDATOR" >/dev/null 2>&1
echo "bracket-expr GREEN(expect 0)=$?"
cd "$HERE/object-axis" && sh "$VALIDATOR" >/dev/null 2>&1
echo "object-axis GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$HERE/three-by-two" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round12 three-by-two(expect 0)=$?"
cd "$HERE/name-squote-escape" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round13 squote-escape(expect 0)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
