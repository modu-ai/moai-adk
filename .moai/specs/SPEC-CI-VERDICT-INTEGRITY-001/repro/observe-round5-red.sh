#!/bin/sh
# t1534 round-5 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (ef8ebd702) — a /tmp leftover is not an
# attributed baseline. Prep failure exits 9 (distinct from the observed
# exit codes). Expect exit 1 (phantom) on all three fixtures.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="ef8ebd702"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r5-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/matrix-squote" && sh "$VAL" >/dev/null 2>&1
echo "squote RED(expect 1)=$?"
cd "$HERE/matrix-readd" && sh "$VAL" >/dev/null 2>&1
echo "readd RED(expect 1)=$?"
cd "$HERE/matrix-include-suffix" && sh "$VAL" >/dev/null 2>&1
echo "include-suffix RED(expect 1)=$?"
exit 0
