#!/bin/sh
# t1534 repro E26 — deadline-absence probe, REAL-BODY OBSERVATION form
# (AC-CI-005; leader micro-card t1543 P2-M supersedes all term-map/string
# approaches, and t1543 P2-L fixes the exit semantics).
#
# Extracts the merge step's REAL body (env map + run block) from the LIVE
# auto-merge.yml at run time, redirects gh at a call-recording stub, provides
# the DECLARED CLOCK INTERFACE (env: current_time, merge_deadline — epoch
# seconds), and executes the body twice:
#
#   pre-deadline  (current_time=1000, merge_deadline=1140)
#   post-deadline (current_time=1200, merge_deadline=1140)
#
# PASS = the observed pair: `gh pr merge` CALLED pre-deadline + NOT called
# post-deadline. Calling without evaluating the clock (both called) is the
# defect. The step's `if:` guard (three outputs, all true at this point in the
# real flow) is pre-satisfied and NOT re-evaluated here — the clock interface
# does the gating.
#
# Exit semantics (t1543 P2-L):
#   0  RED   — gh pr merge called in BOTH clock runs (deadline never evaluated;
#              the current tree's state)
#   2  FLIP  — called pre-deadline, NOT called post-deadline (M1 contract holds)
#   3  BAD   — gh pr merge NOT called pre-deadline (guard blocks a legitimate
#              merge; any non-proceeds-first shape)
#   4  HARNESS — the body itself failed to execute cleanly in either clock run
#              (probe-infrastructure failure; NEVER conflated with the repair
#              signal)
#   9  ERROR — extraction failure (loud, never silent)
#
# DECLARED CLOCK INTERFACE (keep-set contract note, plan §F M2): M1's
# implementation must read the env vars `current_time` and `merge_deadline`
# (epoch seconds) in the merge step for this probe to observe its gating.
# Run from the repository root.
set -eu
WF=".github/workflows/auto-merge.yml"
ANCHOR="name: Auto merge PR"

MATCHES="$(grep -c "$ANCHOR" "$WF" 2>/dev/null || true)"
[ -n "$MATCHES" ] || MATCHES=0
STEP="$(awk -v ANCHOR="$ANCHOR" '
  index($0, ANCHOR) { grab = 1; next }
  grab && match($0, /^ */) && RLENGTH < 6 && $0 !~ /^ *$/ { exit }
  grab { print }
' "$WF")"

# Extraction guard: renamed step or restructured YAML fails LOUDLY.
if [ "$MATCHES" -ne 1 ] || [ -z "$(printf '%s' "$STEP" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: anchor '$ANCHOR' matched $MATCHES line(s) in $WF; step empty. The workflow step name or structure changed — update this probe's anchor." >&2
  exit 9
fi

# Split the step into its env map and run block.
ENV_EXPORTS="$(printf '%s\n' "$STEP" | awk '
  /^          [A-Za-z_][A-Za-z_0-9]*: / {
    line = $0
    sub(/^ */, "", line)
    name = line; sub(/:.*/, "", name)
    val = line; sub(/^[^:]*: */, "", val)
    gsub(/'\''/, "", val)
    print "export " name "='"'"'" val "'"'"'"
  }
')"
RUNBLOCK="$(printf '%s\n' "$STEP" | awk '
  index($0, "run: |") { rb = 1; next }
  rb && match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/ { exit }
  rb { sub(/^ {10}/, ""); print }
')"

if [ -z "$(printf '%s' "$ENV_EXPORTS" | tr -d '[:space:]')" ] || [ -z "$(printf '%s' "$RUNBLOCK" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: env map or run block empty in $WF merge step — the workflow structure changed; update this probe." >&2
  exit 9
fi

# Compose the executable body ONCE: clock inputs are prepended per run; env
# exports carry substituted template expressions; then the real run block.
BODY_SCRIPT="$(mktemp /tmp/t1534-e26-XXXXXX)"
{
  printf '%s\n' "$ENV_EXPORTS" \
    | sed -e 's/\${{ secrets.GITHUB_TOKEN }}/dummy-t1534-token/' \
          -e 's/\${{ steps.pr.outputs.branch }}/release\/v3.3.0-probe/' \
          -e 's/\${{ steps.pr.outputs.number }}/1/'
  printf '%s\n' "$RUNBLOCK" \
    | sed -e 's/\${{ steps.pr.outputs.number }}/1/'
} > "$BODY_SCRIPT"

HERE="$(cd "$(dirname "$0")" && pwd)"

# run_clock <current_time> — executes the real body under one clock input.
# Prints "body_exit=<rc> calls=<n>" on stdout; never aborts the probe.
run_clock() {
  current_time="$1"
  LOG="$(mktemp /tmp/t1534-e26-log-XXXXXX)"
  # Pre-flight (t1543): an unwritable call log would make the stub fail
  # silently, the body would take its else branch, and the harness failure
  # would masquerade as "guard blocks the merge" — refuse loudly instead.
  if [ ! -w "$LOG" ]; then
    echo "HARNESS FAILURE: call log $LOG is not writable — probe infrastructure, never a repair signal" >&2
    rm -f "$LOG"
    exit 4
  fi
  CLOCK_SCRIPT="$(mktemp /tmp/t1534-e26-XXXXXX)"
  {
    printf 'current_time=%s\nmerge_deadline=1140\n' "$current_time"
    cat "$BODY_SCRIPT"
  } > "$CLOCK_SCRIPT"
  body_rc=0
  PATH="$HERE/stub-record:$PATH" MOCK_CALL_LOG="$LOG" \
    bash -e -o pipefail "$CLOCK_SCRIPT" > /dev/null 2>&1 || body_rc=$?
  calls=0
  if [ -f "$LOG" ]; then
    calls="$(grep -c 'pr merge' "$LOG" 2>/dev/null || true)"
    [ -n "$calls" ] || calls=0
  fi
  echo "body_exit=$body_rc calls=$calls"
  rm -f "$CLOCK_SCRIPT" "$LOG"
}

RESULTS="$(run_clock 1000; run_clock 1200)"
echo "$RESULTS"
rm -f "$BODY_SCRIPT"

pre_line="$(printf '%s\n' "$RESULTS" | sed -n 1p)"
post_line="$(printf '%s\n' "$RESULTS" | sed -n 2p)"
pre_body="$(printf '%s' "$pre_line" | sed 's/.*body_exit=\([0-9]*\).*/\1/')"
pre_calls="$(printf '%s' "$pre_line" | sed 's/.*calls=\([0-9]*\).*/\1/')"
post_body="$(printf '%s' "$post_line" | sed 's/.*body_exit=\([0-9]*\).*/\1/')"
post_calls="$(printf '%s' "$post_line" | sed 's/.*calls=\([0-9]*\).*/\1/')"

# P2-L: harness/execution failure gets its own code — never conflated with a
# repair signal, and the pair must both be clean before any verdict.
if [ "$pre_body" -ne 0 ] || [ "$post_body" -ne 0 ]; then
  echo "HARNESS FAILURE: the merge-step body did not execute cleanly under the probe (pre_exit=$pre_body post_exit=$post_body) — probe-infrastructure failure, never a repair signal" >&2
  exit 4
fi

if [ "$pre_calls" -eq 0 ]; then
  echo "GUARD MISBEHAVES: gh pr merge was not called at current_time=1000 (pre-deadline) — a merge must proceed before the deadline before any post-deadline withholding can count" >&2
  exit 3
fi

if [ "$post_calls" -eq 0 ]; then
  echo "M1 CONTRACT HOLDS: gh pr merge called pre-deadline ($pre_calls call(s)) and withheld post-deadline — both clock decisions observed"
  exit 2
fi

echo "DEFECT: gh pr merge called in BOTH clock runs ($pre_calls pre / $post_calls post, merge_at=1141 > declared_deadline=1140) — deadline never evaluated"
exit 0
