#!/bin/sh
# t1534 round-15 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (166882a89); prep failure exits 9. Expect all
# three exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="166882a89"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r15-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/squote-inner" && sh "$VAL" >/dev/null 2>&1
echo "squote-inner RED(expect 1)=$?"
cd "$HERE/obj-two-fields" && sh "$VAL" >/dev/null 2>&1
echo "obj-two-fields RED(expect 1)=$?"
cd "$HERE/empty-include-val" && sh "$VAL" >/dev/null 2>&1
echo "empty-include-val RED(expect 1)=$?"
exit 0
