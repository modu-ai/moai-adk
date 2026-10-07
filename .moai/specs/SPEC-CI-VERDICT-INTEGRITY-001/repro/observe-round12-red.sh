#!/bin/sh
# t1534 round-12 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (058a7fe6a); prep failure exits 9. Expect exit 1
# (the global-n clobber drops combinations AND the space separator misses).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="058a7fe6a"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r12-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/three-by-two" && sh "$VAL" >/dev/null 2>&1
echo "three-by-two RED(expect 1)=$?"
exit 0
