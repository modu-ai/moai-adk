#!/usr/bin/env bash
# t664 residual-risk measurement — walk cost of removing the depth cap.
#
# Times detect_languages end-to-end against a real tree. That is the operational
# unit: has_suffix is defined INSIDE detect_languages and does not exist at
# global scope, so a probe that sources the hook and calls has_suffix directly
# measures nothing at all (it calls a function that is not defined). One full
# detect_languages call is also what the gate actually pays per invocation.
#
# The expensive direction is a MISS: -quit stops the walk at the first match, so
# a language that is present costs a partial traversal, while every language
# absent from the tree costs a full pruned traversal. A repo triggers roughly one
# full traversal per absent language.
#
# Usage: measure-cost.sh <hook> <tree> <reps>
set -uo pipefail

HOOK="${1:?hook}"; TREE="${2:?tree}"; REPS="${3:-3}"

set -- ""
# shellcheck disable=SC1090
source "$HOOK"

command -v detect_languages >/dev/null 2>&1 || {
    printf 'FAIL: detect_languages not defined after sourcing %s\n' "$HOOK" >&2
    exit 1
}

start=$(date +%s.%N)
i=0
while [ "$i" -lt "$REPS" ]; do
    detect_languages "$TREE" >/dev/null 2>&1
    i=$((i + 1))
done
end=$(date +%s.%N)

printf 'hook\t%s\n' "$HOOK"
printf 'reps\t%s\n' "$REPS"
printf 'sec_per_detect_languages\t'
awk -v s="$start" -v e="$end" -v r="$REPS" 'BEGIN{printf "%.3f\n", (e-s)/r}'
printf 'candidates\t%s\n' "$(detect_languages "$TREE" | paste -sd, -)"
