#!/usr/bin/env bash
# Watch CodeRabbit's review status on a PR head until it reaches the TARGET state.
#
# Written after a negative-condition watch ("until description is not 'in progress'")
# exited immediately on "Review rate limited" and reported completion while having
# observed nothing. A negative condition is satisfied by every state except the one
# named — success, failure, and any string added later.
#
# So: assert the POSITIVE target, bound the wait, and let the exit code say which
# thing happened rather than merely that the loop ended.
#
#   exit 0  — reached "Review completed"        (the only success)
#   exit 3  — terminal non-target state observed (rate limited / skipped) — read it, act on it
#   exit 4  — timed out still pending            (unknown, not a pass)
#
# Usage: watch-review.sh <repo> <sha> [timeout-seconds]

set -o pipefail

REPO="${1:?repo required, e.g. modu-ai/moai-adk}"
SHA="${2:?head sha required}"
TIMEOUT="${3:-1800}"
# Guard the arithmetic below: an attacker-controlled or malformed arg would otherwise
# reach $(( )) as an array-subscript expansion (iter3 out-of-scope finding, fixed on adoption).
[[ "$TIMEOUT" =~ ^[0-9]+$ ]] || { echo "watch-review.sh: TIMEOUT must be a non-negative integer, got: $TIMEOUT" >&2; exit 5; }
INTERVAL=60

deadline=$(( SECONDS + TIMEOUT ))

while [ "$SECONDS" -lt "$deadline" ]; do
    desc=$(gh api "repos/$REPO/commits/$SHA/status" \
             --jq '.statuses[] | select(.context=="CodeRabbit") | .description' 2>/dev/null | tail -1)

    case "$desc" in
        "Review completed")
            echo "REACHED: $desc"
            exit 0
            ;;
        "Review rate limited"|"Review skipped")
            # Terminal for this attempt, and NOT a pass. Each needs a different next
            # action, so report the string rather than collapsing them into "done".
            echo "TERMINAL-NOT-TARGET: $desc"
            exit 3
            ;;
        *)
            # Empty (no row yet) or in progress — keep waiting.
            sleep "$INTERVAL"
            ;;
    esac
done

echo "TIMEOUT after ${TIMEOUT}s; last observed: ${desc:-<no CodeRabbit row>}"
exit 4
