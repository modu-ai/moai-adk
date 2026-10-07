#!/bin/sh
# t1534 round-9 GREEN observation: the REPAIRED validator on the three
# findings plus the whole existing probe suite.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
VALIDATOR="$REPO_ROOT/scripts/ci-mirror/validate-required-checks.sh"
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -f "$LOG"' EXIT INT TERM
cd "$HERE/matrix-hyphen-key" && sh "$VALIDATOR" >/dev/null 2>&1
echo "hyphen-key GREEN(expect 0)=$?"
cd "$HERE/matrix-flow-form" && sh "$VALIDATOR" >/dev/null 2>&1
echo "flow-form GREEN(expect 0)=$?"
cd "$HERE/matrix-space-value" && sh "$VALIDATOR" >/dev/null 2>&1
echo "space-value GREEN(expect 0)=$?"
cd "$HERE"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom.sh" >/dev/null 2>&1; echo "phantom(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-phantom-control.sh" >/dev/null 2>&1; echo "control(expect 0)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-malformed.sh" >/dev/null 2>&1; echo "malformed(expect 1)=$?"
sh "$REPO_ROOT/.moai/specs/SPEC-CI-VERDICT-INTEGRITY-001/repro/run-matrix-exclude.sh" >/dev/null 2>&1; echo "matrix-exclude(expect 1)=$?"
cd "$REPO_ROOT"
sh "$VALIDATOR" >/dev/null 2>&1; echo "repo-root(expect 0)=$?"
exit 0
