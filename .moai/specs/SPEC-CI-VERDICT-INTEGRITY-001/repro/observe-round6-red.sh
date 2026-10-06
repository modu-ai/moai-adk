#!/bin/sh
# t1534 round-6 RED observation: the PRE-REPAIR copies are extracted FROM
# THE COMMITTED TREE (a236ac8b3) — a /tmp leftover is not an attributed
# baseline; prep failure exits 9. Expect: validator 3 exit 1 (phantom),
# watch 1 exit 0 (false green).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="a236ac8b3"
STUB="$HERE/stub-base-develop/gh"
chmod +x "$STUB" 2>/dev/null || true
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r6-XXXXXXXX")"
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"; rm -f "$LOG"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
WATCH="$PREP_DIR/watch.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/matrix-block" && sh "$VAL" >/dev/null 2>&1
echo "block RED(expect 1)=$?"
cd "$HERE/matrix-merge" && sh "$VAL" >/dev/null 2>&1
echo "merge RED(expect 1)=$?"
cd "$HERE/job-id" && sh "$VAL" >/dev/null 2>&1
echo "job-id RED(expect 1)=$?"
cp -R "$REPO_ROOT/scripts/ci-watch/lib" "$PREP_DIR/lib" || {
	echo "baseline prep FAILED (exit 9) — lib copy" >&2
	exit 9
}
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-watch/run.sh" > "$WATCH" || {
	echo "baseline prep FAILED (exit 9) — watch copy" >&2
	exit 9
}
chmod +x "$WATCH"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$STUB" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "unkeyed-base RED(expect 0)=$?"
grep -c "ADVISORY" "$LOG"
exit 0
