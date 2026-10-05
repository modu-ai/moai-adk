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
ANCHOR="name: Verify all OS legs passed"

MATCHES="$(grep -c "$ANCHOR" "$WF" 2>/dev/null || true)"
[ -n "$MATCHES" ] || MATCHES=0
BODY="$(awk -v ANCHOR="$ANCHOR" '
  index($0, ANCHOR) {
    if (!grab) { match($0, /^ */); si = RLENGTH; grab = 1; next }
  }
  grab && match($0, /^ */) && RLENGTH < si && $0 !~ /^ *$/ { exit }
  grab && !runfound && /^ *run: */ {
    line = $0; sub(/^ *run: */, "", line)
    if (line ~ /^[>|][+-]?[0-9]*$/) { runblock = 1; next }
    print line; exit
  }
  grab && runblock {
    if (match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/) { exit }
    sub(/^ {10}/, ""); print
  }
' "$WF")"

# Extraction guard (amendment 4): a renamed step or changed YAML structure must
# fail LOUDLY — never silently run an empty body and report exit 0.
if [ "$MATCHES" -ne 1 ] || [ -z "$(printf '%s' "$BODY" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: anchor '$ANCHOR' matched $MATCHES line(s) in $WF; body empty=$([ -z "$(printf '%s' "$BODY" | tr -d '[:space:]')" ] && echo yes || echo no). The workflow step name or structure changed — update this wrapper's anchor." >&2
  exit 9
fi

# Substitute the dependency results AND the detect outputs the repaired gate
# body may reference (P2-S); any remaining template expression is a harness
# error (exit 4), never a gate verdict.
BODY="$(printf '%s\n' "$BODY" \
  | sed -e 's/\${{ needs.detect-release.result }}/success/g' \
        -e 's/\${{ needs.full-matrix-test.result }}/cancelled/g' \
        -e 's/\${{ needs.detect-release.outputs.go_code }}/true/g' \
        -e 's/\${{ needs.detect-release.outputs.docs_only }}/false/g')"
if printf '%s' "$BODY" | grep -q '\${{'; then
  echo "repro substitution incomplete: unresolved \${{ ... }} expression remains in the extracted gate body — add it to this wrapper's substitution map" >&2
  exit 4
fi

SCRIPT="$(mktemp /tmp/t1534-e2-XXXXXX)" || { echo "harness failure: mktemp" >&2; exit 4; }
trap 'rm -f "$SCRIPT"' EXIT
printf '%s\n' "$BODY" > "$SCRIPT"
# Actions executes a shell:-less step's run block as `bash --noprofile --norc
# -e {0}` on Linux runners (P2-T: -e parity, NO pipefail — that flag is not in
# the runner default).
bash -e "$SCRIPT"
