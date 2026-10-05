#!/bin/sh
# t1534 repro E2 — extracts the release-pr-gate "Verify all OS legs passed" run
# block from the LIVE workflow file at run time (never a frozen copy), applies
# the declared GitHub template substitution (needs.detect-release.result ->
# "success", needs.full-matrix-test.result -> "cancelled"), and runs the
# substituted body.
#
# RED expectation (pre-repair): exit 0 with the unconditional
# "verification PASSED" line — the gate treats a cancelled matrix as success.
# GREEN expectation (post-M1): non-zero with an error annotation.
# Run from the repository root.
set -eu
WF=".github/workflows/release-pr-multi-os.yml"

BODY="$(awk '
  index($0, "name: Verify all OS legs passed") { grab = 1; next }
  grab && index($0, "run: |") { runblock = 1; next }
  grab && runblock && match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/ { exit }
  grab && runblock { print }
' "$WF")"

# GitHub renders each ${{ needs.X.result }} as a bare double-quoted value.
BODY="$(printf '%s\n' "$BODY" \
  | sed 's/\${{ needs.detect-release.result }}/success/g; s/\${{ needs.full-matrix-test.result }}/cancelled/g')"

SCRIPT="$(mktemp /tmp/t1534-e2-XXXXXX)"
trap 'rm -f "$SCRIPT"' EXIT
printf '%s\n' "$BODY" > "$SCRIPT"
sh "$SCRIPT"
