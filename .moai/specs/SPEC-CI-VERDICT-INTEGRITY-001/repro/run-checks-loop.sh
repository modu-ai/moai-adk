#!/bin/sh
# t1534 repro E4 — extracts the auto-merge "Check all required CI checks passed"
# run block from the LIVE workflow file at run time (never a frozen copy),
# substitutes the template expression, binds GITHUB_OUTPUT to stdout so the
# emitted should_merge lands in the visible stream, and redirects gh at the
# committed stub (repro/stubbin/gh).
#
# RED expectation (pre-repair): stub lookup failure/empty output ->
# "should_merge=true" (the loop treats an incomplete observation as complete).
# GREEN expectation (post-M1): should_merge=false.
# Run from the repository root.
set -eu
WF=".github/workflows/auto-merge.yml"

BODY="$(awk '
  index($0, "name: Check all required CI checks passed") { grab = 1; next }
  grab && index($0, "run: |") { runblock = 1; next }
  grab && runblock && match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/ { exit }
  grab && runblock { print }
' "$WF")"

BODY="$(printf '%s\n' "$BODY" | sed 's/\${{ steps.pr.outputs.number }}/1/g')"

SCRIPT="$(mktemp /tmp/t1534-e4-XXXXXX)"
trap 'rm -f "$SCRIPT"' EXIT
printf '%s\n' "$BODY" > "$SCRIPT"

HERE="$(cd "$(dirname "$0")" && pwd)"
PATH="$HERE/stubbin:$PATH" GITHUB_OUTPUT=/dev/stdout MAX_WAIT=1 sh "$SCRIPT"
