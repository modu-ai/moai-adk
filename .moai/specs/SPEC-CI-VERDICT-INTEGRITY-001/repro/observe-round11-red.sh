#!/bin/sh
# t1534 round-11 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (ba824499f); prep failure exits 9. Expect all
# three exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="ba824499f"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r11-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/suffix-extra" && sh "$VAL" >/dev/null 2>&1
echo "suffix-extra RED(expect 1)=$?"
cd "$HERE/quoted-colon" && sh "$VAL" >/dev/null 2>&1
echo "quoted-colon RED(expect 1)=$?"
cd "$HERE/amp-value" && sh "$VAL" >/dev/null 2>&1
echo "amp-value RED(expect 1)=$?"
exit 0
