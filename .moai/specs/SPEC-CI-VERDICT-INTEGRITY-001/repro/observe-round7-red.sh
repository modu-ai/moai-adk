#!/bin/sh
# t1534 round-7 RED observation: the PRE-REPAIR copies are extracted FROM
# THE COMMITTED TREE (c5f5ff229) — a /tmp leftover is not an attributed
# baseline; prep failure exits 9. Expect: merge-all-legs 1, merge-overwrite
# 1, comment-boundary 1, quoted-key watch 1 (spurious abort), and the
# green-gone negative probe 0 on the pre-repair (green wrongly allowed).
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="c5f5ff229"
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r7-XXXXXXXX")"
LOG="$(mktemp "${TMPDIR:-/tmp}/t1534-obs-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"; rm -f "$LOG"' EXIT INT TERM
VAL="$PREP_DIR/validator.sh"
WATCH="$PREP_DIR/watch.sh"
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-mirror/validate-required-checks.sh" > "$VAL" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$VAL"
cd "$HERE/merge-all-legs" && sh "$VAL" >/dev/null 2>&1
echo "merge-all-legs RED(expect 1)=$?"
cd "$HERE/merge-overwrite" && sh "$VAL" >/dev/null 2>&1
echo "merge-overwrite RED(expect 1)=$?"
cd "$HERE/comment-boundary" && sh "$VAL" >/dev/null 2>&1
echo "comment-boundary RED(expect 1)=$?"
cd "$HERE/merge-green-gone" && sh "$VAL" >/dev/null 2>&1
echo "green-gone RED(expect 0 = wrongly allowed)=$?"
cp -R "$REPO_ROOT/scripts/ci-watch/lib" "$PREP_DIR/lib" || {
	echo "baseline prep FAILED (exit 9) — lib copy" >&2
	exit 9
}
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-watch/run.sh" > "$WATCH" || {
	echo "baseline prep FAILED (exit 9) — watch copy" >&2
	exit 9
}
chmod +x "$WATCH"
cd "$HERE/quoted-key"
MOAI_CIWATCH_GH="$HERE/stub-skipping/gh" \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
sh "$WATCH" 99 > "$LOG" 2>&1
echo "quoted-key watch RED(expect 1)=$?"
grep -c "no key for base branch" "$LOG"
exit 0
