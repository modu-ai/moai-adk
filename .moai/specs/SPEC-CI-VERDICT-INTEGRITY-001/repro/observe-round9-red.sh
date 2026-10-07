#!/bin/sh
# t1534 round-9 RED observation: the PRE-REPAIR validator is extracted
# FROM THE COMMITTED TREE (d98812e39) — a /tmp leftover is not an
# attributed baseline; prep failure exits 9. Expect all three exit 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="d98812e39"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r9-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/matrix-hyphen-key" && sh "$VAL" >/dev/null 2>&1
echo "hyphen-key RED(expect 1)=$?"
cd "$HERE/matrix-flow-form" && sh "$VAL" >/dev/null 2>&1
echo "flow-form RED(expect 1)=$?"
cd "$HERE/matrix-space-value" && sh "$VAL" >/dev/null 2>&1
echo "space-value RED(expect 1)=$?"
exit 0
