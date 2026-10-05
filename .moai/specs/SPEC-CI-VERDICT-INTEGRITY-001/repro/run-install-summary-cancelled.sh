#!/bin/sh
# t1534 repro E7 — extracts the test-install summary "Check test results" run
# block from the LIVE workflow file at run time (never a frozen copy) and runs
# the THREE-VARIANT input matrix (P2-I): flipping every result to `cancelled`
# simultaneously passes even a partial-dependency mutant (one that checks only
# test-sh), so the matrix varies ONE dependency at a time with the rest at
# success, plus a full-success control.
#
# Actions executes a shell:-less step's run block as `bash --noprofile --norc
# -e -o pipefail` on Linux runners — replicate that, not bare sh (P2-E); the
# body's [[ ]] conditions are bash-only and silently mis-evaluate under dash.
#
# RED expectation (pre-repair): ALL THREE variants exit 0 — variant A (healthy)
# legitimately, variants B/C because only "failure" is rejected (B: a cancelled
# dependency is invisible; C: parity is not even a dependency yet).
# GREEN expectation (post-M1): A stays exit 0; B and C exit non-zero.
# Run from the repository root.
set -eu
WF=".github/workflows/test-install.yml"
ANCHOR="name: Check test results"

MATCHES="$(grep -c "$ANCHOR" "$WF" 2>/dev/null || true)"
[ -n "$MATCHES" ] || MATCHES=0
BODY="$(awk -v ANCHOR="$ANCHOR" '
  index($0, ANCHOR) { grab = 1; next }
  grab && index($0, "run: |") { runblock = 1; next }
  grab && runblock && match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/ { exit }
  grab && runblock { print }
' "$WF")"

# Extraction guard (amendment 4): a renamed step or changed YAML structure must
# fail LOUDLY — never silently run an empty body and report exit 0.
if [ "$MATCHES" -ne 1 ] || [ -z "$(printf '%s' "$BODY" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: anchor '$ANCHOR' matched $MATCHES line(s) in $WF; body empty=$([ -z "$(printf '%s' "$BODY" | tr -d '[:space:]')" ] && echo yes || echo no). The workflow step name or structure changed — update this wrapper's anchor." >&2
  exit 9
fi

run_variant() {
  label="$1"
  shift
  SUB="$BODY"
  for pair in "$@"; do
    job="$(printf '%s' "$pair" | sed 's/=.*//')"
    val="$(printf '%s' "$pair" | sed 's/.*=//')"
    SUB="$(printf '%s\n' "$SUB" | sed "s/\${{ needs.${job}.result }}/${val}/g")"
  done
  VARIANT_SCRIPT="$(mktemp /tmp/t1534-e7-XXXXXX)"
  printf '%s\n' "$SUB" > "$VARIANT_SCRIPT"
  rc=0
  bash -e -o pipefail "$VARIANT_SCRIPT" || rc=$?
  rm -f "$VARIANT_SCRIPT"
  echo "variant $label: exit $rc"
}

echo "E7 input matrix (Actions default shell: bash -e -o pipefail):"
run_variant "A-full-success" \
  "test-sh=success" "test-ps1-pwsh=success" "test-ps1-powershell=success" \
  "test-bat=success" "compatibility-check=success" "install-script-parity=success"
run_variant "B-test-sh-cancelled-parity-success" \
  "test-sh=cancelled" "test-ps1-pwsh=success" "test-ps1-powershell=success" \
  "test-bat=success" "compatibility-check=success" "install-script-parity=success"
run_variant "C-parity-cancelled-test-sh-success" \
  "test-sh=success" "test-ps1-pwsh=success" "test-ps1-powershell=success" \
  "test-bat=success" "compatibility-check=success" "install-script-parity=cancelled"
