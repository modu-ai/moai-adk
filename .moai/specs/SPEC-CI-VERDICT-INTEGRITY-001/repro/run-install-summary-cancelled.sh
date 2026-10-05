#!/bin/sh
# t1534 repro E7 — extracts the test-install summary "Check test results" run
# block from the LIVE workflow file at run time (never a frozen copy) and
# substitutes every needs.X.result as "cancelled", exactly as GitHub renders it
# when the whole run is cancelled.
#
# RED expectation (pre-repair): "All tests passed!" exit 0 (only "failure" is
# rejected). GREEN expectation (post-M1): non-zero.
# Run from the repository root.
set -eu
WF=".github/workflows/test-install.yml"

BODY="$(awk '
  index($0, "name: Check test results") { grab = 1; next }
  grab && index($0, "run: |") { runblock = 1; next }
  grab && runblock && match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/ { exit }
  grab && runblock { print }
' "$WF")"

BODY="$(printf '%s\n' "$BODY" | sed 's/\${{ needs.[a-z0-9-]*.result }}/cancelled/g')"

SCRIPT="$(mktemp /tmp/t1534-e7-XXXXXX)"
trap 'rm -f "$SCRIPT"' EXIT
printf '%s\n' "$BODY" > "$SCRIPT"
sh "$SCRIPT"
