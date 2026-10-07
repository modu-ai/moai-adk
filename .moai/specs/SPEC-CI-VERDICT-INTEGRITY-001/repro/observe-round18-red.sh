#!/bin/sh
# t1534 round-18 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (2eb41ec31); prep failure exits 9.
# obj-exclude and first-empty-excl are INVERTED probes: the pre-repair
# validator wrongly APPROVES excluded combinations (expect 0).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="2eb41ec31"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r18-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/obj-exclude" && sh "$VAL" >/dev/null 2>&1
echo "obj-exclude RED(expect 0 = false approval)=$?"
cd "$HERE/first-empty-excl" && sh "$VAL" >/dev/null 2>&1
echo "first-empty-excl RED(expect 0 = false approval)=$?"
cd "$HERE/missing-include-field" && sh "$VAL" >/dev/null 2>&1
echo "missing-include-field RED(expect 1)=$?"
cd "$HERE/empty-array" && sh "$VAL" >/dev/null 2>&1
echo "empty-array RED(expect 0 = false approval)=$?"
exit 0
