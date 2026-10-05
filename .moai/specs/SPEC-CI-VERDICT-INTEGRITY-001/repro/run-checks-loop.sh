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
ANCHOR="name: Check all required CI checks passed"

MATCHES="$(grep -c "$ANCHOR" "$WF" 2>/dev/null || true)"
[ -n "$MATCHES" ] || MATCHES=0
BODY="$(awk -v ANCHOR="$ANCHOR" '
  index($0, ANCHOR) {
    if (!grab) { match($0, /^ */); si = RLENGTH; grab = 1; next }
  }
  grab && match($0, /^ */) && RLENGTH <= si && $0 !~ /^ *$/ { exit }
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

BODY="$(printf '%s\n' "$BODY" | sed 's/\${{ steps.pr.outputs.number }}/1/g')"

SCRIPT="$(mktemp /tmp/t1534-e4-XXXXXX)"
OUT="$(mktemp /tmp/t1534-e4-out-XXXXXX)"
trap 'rm -f "$SCRIPT" "$OUT"' EXIT
printf '%s\n' "$BODY" > "$SCRIPT"

HERE="$(cd "$(dirname "$0")" && pwd)"
# GITHUB_OUTPUT is a regular temp file (P2-F): /dev/stdout is refused in
# restricted environments, which silently drops the should_merge observation.
# Actions executes a shell:-less step as `bash -e` (P2-T: -e parity, NO
# pipefail — that flag is not in the runner default); the body's exit code is
# captured so the wrapper can print the emitted outputs before propagating it.
rc=0
PATH="$HERE/stubbin:$PATH" GITHUB_OUTPUT="$OUT" MAX_WAIT=1 bash -e "$SCRIPT" || rc=$?
echo "--- GITHUB_OUTPUT ---"
cat "$OUT"
exit "$rc"
