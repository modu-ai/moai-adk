#!/bin/sh
# t1534 round-14 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (ca15eceda); prep failure exits 9. Expect all
# three exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="ca15eceda"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r14-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/apos-tail" && sh "$VAL" >/dev/null 2>&1
echo "apos-tail RED(expect 1)=$?"
cd "$HERE/bracket-expr" && sh "$VAL" >/dev/null 2>&1
echo "bracket-expr RED(expect 1)=$?"
cd "$HERE/object-axis" && sh "$VAL" >/dev/null 2>&1
echo "object-axis RED(expect 1)=$?"
exit 0
