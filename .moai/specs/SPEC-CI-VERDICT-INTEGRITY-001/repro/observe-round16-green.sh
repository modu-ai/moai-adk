#!/bin/sh
# t1534 round-16 GREEN observation: the REPAIRED validator on the two
# findings plus the whole existing probe suite. excl-apostrophe is
# INVERTED: the repaired validator REJECTS the excluded combination
# (expect 1).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
cd "$HERE/excl-apostrophe" && sh "$VALIDATOR" >/dev/null 2>&1
echo "excl-apostrophe GREEN(expect 1 = excluded combo rejected)=$?"
cd "$HERE/empty-dim-val" && sh "$VALIDATOR" >/dev/null 2>&1
echo "empty-dim-val GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$HERE/obj-two-fields" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round15 obj-two-fields(expect 0)=$?"
cd "$HERE/squote-inner" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round15 squote-inner(expect 0)=$?"
cd "$HERE/empty-include-val" && sh "$VALIDATOR" >/dev/null 2>&1; echo "round15 empty-include-val(expect 0)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
bash "$REPO_ROOT/scripts/ci-watch/test/run_test.sh" 2>&1 | tail -1
exit 0
