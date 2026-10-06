#!/bin/sh
# t1534 round-5 worse-order observation: the same {skipping, pass}
# duplicate name in both array orders. The PRE-REPAIR loop is extracted
# FROM THE COMMITTED TREE (ef8ebd702) with its lib — a /tmp leftover is
# not an attributed baseline; prep failure exits 9. RED = the pre-repair
# loop disagrees between orders (pending vs pass). GREEN = the repaired
# loop reports pending in BOTH orders.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)" || exit 9
BASE_COMMIT="ef8ebd702"
REPO_WATCH="$REPO_ROOT/scripts/ci-watch/run.sh"
STUB="$HERE/stub-skipping/gh"
chmod +x "$STUB" 2>/dev/null || true
PREP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/t1534-pre-r5w-XXXXXXXX")"
trap 'rm -rf "$PREP_DIR"' EXIT INT TERM
cp -R "$REPO_ROOT/scripts/ci-watch/lib" "$PREP_DIR/lib" || {
	echo "baseline prep FAILED (exit 9) — lib copy" >&2
	exit 9
}
git -C "$REPO_ROOT" show "$BASE_COMMIT:scripts/ci-watch/run.sh" > "$PREP_DIR/run.sh" || {
	echo "baseline prep FAILED (exit 9) — cannot attribute to $BASE_COMMIT" >&2
	exit 9
}
chmod +x "$PREP_DIR/run.sh"
cd "$HERE/matrix-exclude"
run_watch() {
    watch_sh="$1"; order="$2"; label="$3"
    out="$(SKIP_ORDER="$order" MOAI_CIWATCH_GH="$STUB" \
        MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
        MOAI_CIWATCH_NO_SLEEP=1 CIWATCH_TIMEOUT_SECONDS=10 \
        sh "$watch_sh" 99 main 2>&1)"
    pending_line="$(printf '%s\n' "$out" | grep 'required ' | head -1)"
    echo "$label order=$order: $pending_line"
}
run_watch "$PREP_DIR/run.sh" 1 "PRE-R5 "
run_watch "$PREP_DIR/run.sh" 2 "PRE-R5 "
run_watch "$REPO_WATCH" 1 "REPAIRED"
run_watch "$REPO_WATCH" 2 "REPAIRED"
exit 0
