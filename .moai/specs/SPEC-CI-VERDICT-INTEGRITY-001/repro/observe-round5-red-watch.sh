#!/bin/sh
# t1534 round-5 RED observation (P1): the PRE-REPAIR watch loop is
# extracted FROM THE COMMITTED TREE (ef8ebd702) with its lib — a /tmp
# leftover is not an attributed baseline. Prep failure exits 9. Expect
# exit 0 — the false green (empty required list scored all-passed).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="ef8ebd702"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r5w-XXXXXXXX")"
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"; rm -f "$LOG"' EXIT INT TERM
mkdir -p "$PREP_DIR"
cp -R "$REPO_ROOT/scripts/ci-watch/lib" "$PREP_DIR/lib" || {
	echo "baseline prep FAILED (exit 9) — lib copy" >&2
	exit 9
}
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-watch/run.sh" > "$PREP_DIR/run.sh" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$PREP_DIR/run.sh" "$HERE/stub-base-fail/gh"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-base-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 \
CIWATCH_TIMEOUT_SECONDS=10 \
sh "$PREP_DIR/run.sh" 99 feature/not-an-ssot-key > "$LOG" 2>&1
echo "base-fail RED(expect 0)=$?"
grep -c "All required checks passed" "$LOG"
exit 0
