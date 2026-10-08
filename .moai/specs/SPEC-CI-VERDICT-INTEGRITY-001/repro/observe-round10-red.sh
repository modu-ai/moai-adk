#!/bin/sh
# t1534 round-10 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (82d5a6e66); prep failure exits 9. Expect
# colon-value RED exit 1 (the valid string value judged a mapping).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="82d5a6e66"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r10-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/matrix-colon-value" && sh "$VAL" >/dev/null 2>&1
echo "colon-value RED(expect 1)=$?"
exit 0
