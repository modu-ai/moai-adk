#!/bin/sh
# t1534 round-12 GREEN observation: the REPAIRED validator on the
# three-by-two finding plus the whole existing probe suite.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/three-by-two" && sh "$VALIDATOR" >/dev/null 2>&1
echo "three-by-two GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$HERE/matrix-include-suffix" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round6 include-suffix(expect 0)=$?"
cd "$HERE/suffix-extra" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round11 suffix-extra(expect 0)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
