#!/bin/sh
# t1534 round-13 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (6aa294541); prep failure exits 9. Expect both
# exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="6aa294541"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r13-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/name-squote-escape" && sh "$VAL" >/dev/null 2>&1
echo "squote-escape RED(expect 1)=$?"
cd "$HERE/matrix-alias" && sh "$VAL" >/dev/null 2>&1
echo "matrix-alias RED(expect 1)=$?"
exit 0
