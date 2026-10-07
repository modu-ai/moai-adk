#!/bin/sh
# t1534 round-17 GREEN observation: the REPAIRED validator on the two
# findings plus the whole existing probe suite. obj-two-items requires
# the REAL published set (GitHub evaluates a missing field as empty:
# windows publishes `Test (windows-latest/)`); obj-mixed-negative is the
# INVERTED probe — the mixed combination GitHub never publishes must be
# rejected (expect 1).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/obj-two-items" && sh "$VALIDATOR" >/dev/null 2>&1
echo "obj-two-items GREEN(expect 0)=$?"
cd "$HERE/obj-mixed-negative" && sh "$VALIDATOR" >/dev/null 2>&1
echo "obj-mixed-negative GREEN(expect 1 = mixed combo rejected)=$?"
cd "$HERE/empty-only-val" && sh "$VALIDATOR" >/dev/null 2>&1
echo "empty-only-val GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$HERE/object-axis" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round14 object-axis(expect 0)=$?"
cd "$HERE/obj-two-fields" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round15 obj-two-fields(expect 0)=$?"
cd "$HERE/empty-dim-val" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round16 empty-dim-val(expect 0)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
