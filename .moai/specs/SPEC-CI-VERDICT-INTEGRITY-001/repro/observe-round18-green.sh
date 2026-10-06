#!/bin/sh
# t1534 round-18 GREEN observation: the REPAIRED validator on the four
# findings plus the whole existing probe suite. obj-exclude,
# first-empty-excl and empty-array are INVERTED probes: the repaired
# validator REJECTS the combinations GitHub never publishes (expect 1).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/obj-exclude" && sh "$VALIDATOR" >/dev/null 2>&1
echo "obj-exclude GREEN(expect 1 = excluded combo rejected)=$?"
cd "$HERE/first-empty-excl" && sh "$VALIDATOR" >/dev/null 2>&1
echo "first-empty-excl GREEN(expect 1 = excluded combo rejected)=$?"
cd "$HERE/missing-include-field" && sh "$VALIDATOR" >/dev/null 2>&1
echo "missing-include-field GREEN(expect 0)=$?"
cd "$HERE/empty-array" && sh "$VALIDATOR" >/dev/null 2>&1
echo "empty-array GREEN(expect 1 = explicit failure)=$?"
cd "$HERE"
cd "$HERE/expr-axis" && sh "$VALIDATOR" >/dev/null 2>&1
echo "expr-axis GREEN(expect 1 = expression axis never fabricates a name)=$?"
cd "$HERE/obj-partial-axis" && sh "$VALIDATOR" >/dev/null 2>&1
echo "obj-partial-axis GREEN(expect 1 = partial combos never publish)=$?"
cd "$HERE/inc-missing-field" && sh "$VALIDATOR" >/dev/null 2>&1
echo "inc-missing-field GREEN(expect 0 = unset tuple field fills empty)=$?"
cd "$HERE/inc-empty-suffix" && sh "$VALIDATOR" >/dev/null 2>&1
echo "inc-empty-suffix GREEN(expect 0 = empty tuple value keeps its slot)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$HERE/obj-two-items" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round17 obj-two-items(expect 0)=$?"
cd "$HERE/obj-mixed-negative" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round17 obj-mixed-negative(expect 1)=$?"
cd "$HERE/empty-only-val" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round17 empty-only-val(expect 0)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
