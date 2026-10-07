#!/bin/sh
# t1534 round-16 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (85ec666c8); prep failure exits 9.
# excl-apostrophe is an INVERTED probe: the pre-repair validator wrongly
# APPROVES the excluded combination (expect 0); empty-dim-val drops the
# empty first combination (expect 1).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="85ec666c8"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r16-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/excl-apostrophe" && sh "$VAL" >/dev/null 2>&1
echo "excl-apostrophe RED(expect 0 = false approval)=$?"
cd "$HERE/empty-dim-val" && sh "$VAL" >/dev/null 2>&1
echo "empty-dim-val RED(expect 1)=$?"
exit 0
