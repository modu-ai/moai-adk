#!/bin/sh
# t1534 round-17 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (284c7d1dd); prep failure exits 9. Expect both
# exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="284c7d1dd"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r17-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/obj-two-items" && sh "$VAL" >/dev/null 2>&1
echo "obj-two-items RED(expect 1)=$?"
cd "$HERE/empty-only-val" && sh "$VAL" >/dev/null 2>&1
echo "empty-only-val RED(expect 1)=$?"
exit 0
