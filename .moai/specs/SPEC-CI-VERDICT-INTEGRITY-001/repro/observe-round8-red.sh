#!/bin/sh
# t1534 round-8 RED observation: the PRE-REPAIR copies are extracted FROM
# THE COMMITTED TREE (8d1b66b51) — a /tmp leftover is not an attributed
# baseline; prep failure exits 9. flow-contexts runs under a yq-less PATH
# (homebrew stripped). bad-link captures the handoff logUrl. Expect:
# flow-contexts watch exit 0 (false green), logUrl pointing at the PASS
# run, name-comment 1, no-space-matrix 1.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="8d1b66b51"
NOYQ_PATH="/usr/bin:/bin"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r8-XXXXXXXX")"
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"; rm -f "$LOG"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
WATCH="$PREP_DIR/watch.sh"
cp -R "$REPO_ROOT/scripts/ci-watch/lib" "$PREP_DIR/lib" || {
	echo "baseline prep FAILED (exit 9) — lib copy" >&2
	exit 9
}
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-watch/run.sh" > "$WATCH" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$WATCH"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — validator copy" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/flow-contexts"
PATH="$NOYQ_PATH" MOAI_CIWATCH_GH="$HERE/stub-main-lint-fail/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "flow-contexts RED(expect 0)=$?"
cd "$HERE/matrix-exclude"
MOAI_CIWATCH_GH="$HERE/stub-bad-link/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "bad-link RED exit(expect 2)=$?"
grep -o "http://example.invalid/[a-z]*" "$LOG" | head -1
cd "$HERE/name-comment" && sh "$VAL" >/dev/null 2>&1
echo "name-comment RED(expect 1)=$?"
cd "$HERE/no-space-matrix" && sh "$VAL" >/dev/null 2>&1
echo "no-space-matrix RED(expect 1)=$?"
exit 0
