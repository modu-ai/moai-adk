#!/bin/sh
# t1534 repro E7 — extracts the test-install summary "Check test results" run
# block from the LIVE workflow file at run time (never a frozen copy) and runs
# the FULL per-dependency input matrix (P2-I + leader addendum): every
# dependency in the summary's needs set flipped ONE at a time to EACH of
# cancelled / timed_out / skipped (6 x 3 = 18 single-failure cases), plus a
# full-success control (19 cases). Rationale: flipping every result to
# `cancelled` simultaneously passes even a partial-dependency mutant (one that
# checks only test-sh and parity), and a cancelled-only matrix misses the
# timed_out/skipped conclusions.
#
# Actions executes a shell:-less step's run block as `bash --noprofile --norc
# -e -o pipefail` on Linux runners — replicate that, not bare sh (P2-E); the
# body's [[ ]] conditions are bash-only and silently mis-evaluate under dash.
#
# RED expectation (pre-repair): the control (A-full-success) exits 0
# legitimately; ALL 18 single-failure variants exit 0 — only "failure" is
# rejected, so cancelled/timed_out/skipped are invisible (and parity is not
# even a dependency yet).
# GREEN expectation (post-M1): A-full-success stays exit 0; all 18
# single-failure variants exit non-zero.
# Run from the repository root.
set -eu
WF=".github/workflows/test-install.yml"
ANCHOR="name: Check test results"
JOBS="test-sh test-ps1-pwsh test-ps1-powershell test-bat compatibility-check install-script-parity"

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
  fail_job="$2"
  fail_conclusion="$3"
  SUB="$BODY"
  # `dep`, not `j` — POSIX sh has no locals; an inner loop named like the
  # caller's loop variable would clobber it and corrupt the matrix (measured).
  for dep in $JOBS; do
    if [ "$dep" = "$fail_job" ]; then
      val="$fail_conclusion"
    else
      val="success"
    fi
    SUB="$(printf '%s\n' "$SUB" | sed "s/\${{ needs.${dep}.result }}/${val}/g")"
  done
  VARIANT_SCRIPT="$(mktemp /tmp/t1534-e7-XXXXXX)"
  printf '%s\n' "$SUB" > "$VARIANT_SCRIPT"
  rc=0
  # Body stdout suppressed: the recorded observation is the per-variant exit.
  bash -e -o pipefail "$VARIANT_SCRIPT" >/dev/null || rc=$?
  rm -f "$VARIANT_SCRIPT"
  echo "variant $label: exit $rc"
}

echo "E7 input matrix (Actions default shell: bash -e -o pipefail; 1 control + 24 single-failure cases):"
run_variant "A-full-success" "" ""

for j in $JOBS; do
  for c in failure cancelled timed_out skipped; do
    run_variant "$j-$c" "$j" "$c"
  done
done
