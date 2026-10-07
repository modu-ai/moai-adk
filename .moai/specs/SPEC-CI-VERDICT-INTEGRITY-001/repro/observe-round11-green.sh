#!/bin/sh
# t1534 round-11 GREEN observation: the REPAIRED validator on the three
# findings plus the whole existing probe suite (including the round-6
# include-suffix and round-7 merge probes that share the touched code).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/suffix-extra" && sh "$VALIDATOR" >/dev/null 2>&1
echo "suffix-extra GREEN(expect 0)=$?"
cd "$HERE/quoted-colon" && sh "$VALIDATOR" >/dev/null 2>&1
echo "quoted-colon GREEN(expect 0)=$?"
cd "$HERE/amp-value" && sh "$VALIDATOR" >/dev/null 2>&1
echo "amp-value GREEN(expect 0)=$?"
cd "$HERE/matrix-include-suffix" && sh "$VALIDATOR" >/dev/null 2>&1
echo "round6 include-suffix(expect 0)=$?"
cd "$HERE/merge-all-legs" && sh "$VALIDATOR" >/dev/null 2>&1
echo "round7 merge-all-legs(expect 0)=$?"
cd "$HERE/merge-overwrite" && sh "$VALIDATOR" >/dev/null 2>&1
echo "round7 merge-overwrite(expect 0)=$?"
cd "$HERE/matrix-colon-value" && sh "$VALIDATOR" >/dev/null 2>&1
echo "round10 colon-value(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
