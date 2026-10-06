#!/bin/sh
# t1534 round-5 worse-order observation: the same {skipping, pass} duplicate
# name in both array orders. RED = the pre-repair loop (ef8ebd702 copy)
# disagrees between orders (pending vs pass). GREEN = the repaired loop
# reports pending in BOTH orders.
set -u
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(git -C "$HERE" rev-parse --show-toplevel)"
REPO_WATCH="$REPO_ROOT/scripts/ci-watch/run.sh"
STUB="$HERE/stub-skipping/gh"
chmod +x "$STUB" 2>/dev/null || true
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
run_watch /tmp/t1534-pre-r5-watch/run.sh 1 "PRE-R5 "
run_watch /tmp/t1534-pre-r5-watch/run.sh 2 "PRE-R5 "
run_watch "$REPO_WATCH" 1 "REPAIRED"
run_watch "$REPO_WATCH" 2 "REPAIRED"
exit 0
