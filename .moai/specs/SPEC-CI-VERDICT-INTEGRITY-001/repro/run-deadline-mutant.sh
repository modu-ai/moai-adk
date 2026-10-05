#!/bin/sh
# t1534 repro E26 — deadline-absence probe, REAL-BODY OBSERVATION form
# (AC-CI-005; micro-card t1546 final shape: P2-L full + P2-M final +
# P2-O single-execution mid-run clock progression + P2-R wait-crossing).
#
# Extracts the merge step's REAL body (env map + run block) from the LIVE
# auto-merge.yml with a BOUNDARY-LIMITED extractor (search scoped to the
# target step; block and single-line run: both handled; exit 9 on missing or
# ambiguous), redirects gh at the recording stub (repro/stub-record/gh) that
# answers the AC-CI-003 head re-query with the FIXTURE_SHA, and executes the
# body under the DECLARED CLOCK INTERFACE:
#
#   env merge_deadline           — epoch seconds (1140 in this probe)
#   env MOAI_CLOCK_FILE          — file holding the served time (epoch
#                                  seconds), RE-READABLE: the probe advances
#                                  it mid-run; M1's deadline implementation
#                                  must read (cat) this file at every
#                                  deadline evaluation — reading it once at
#                                  start and merging later after the served
#                                  time crossed the deadline is the P2-O
#                                  defect this probe exists to catch.
#
# ONE execution, clock starts pre-deadline (1100) and ADVANCES past the
# deadline (1200) mid-run; the recording stub logs each merge call's SERVED
# time. Verdicts:
#   pass  (exit 2) — every recorded merge call served <= deadline AND no
#                    merge call after the served time crossed the deadline
#   RED   (exit 0) — any merge call served AFTER the deadline crossed
#                    (includes: evaluated the clock once pre-crossing then
#                    merged post-crossing; and: never evaluated it at all)
#   3     — the body never attempted a merge at all (withholds everything)
#   4     — harness/execution failure (stub log-write failure marker, body
#            execution failure, mktemp failure) — never a repair signal
#   9     — extraction failure (loud, never silent)
#
# Run from the repository root.
set -eu
WF=".github/workflows/auto-merge.yml"
ANCHOR="name: Auto merge PR"
FIXTURE_SHA="a158b4b5f0000000000000000000000000000000"

harness_fail() {
  echo "HARNESS FAILURE: $*" >&2
  exit 4
}

MATCHES="$(grep -c "$ANCHOR" "$WF" 2>/dev/null)" || MATCHES=0
[ -n "$MATCHES" ] || MATCHES=0
if [ "$MATCHES" -ne 1 ]; then
  echo "repro extraction failed: anchor '$ANCHOR' matched $MATCHES line(s) in $WF (want exactly 1) — update this probe's anchor." >&2
  exit 9
fi

# Boundary-limited extraction (t1546 P2-Q): scope the search to the TARGET
# step. si = the anchor line's indentation (the step's own level). The scan
# stops at the first non-blank line indented above that level — the extractor
# can never leak into a later step's body, and a single-line `run:` on the
# target step is captured as the line remainder instead of being skipped.
STEP="$(awk -v ANCHOR="$ANCHOR" '
  index($0, ANCHOR) {
    if (!grab) { match($0, /^ */); si = RLENGTH; grab = 1; next }
  }
  grab && match($0, /^ */) && RLENGTH < si && $0 !~ /^ *$/ { exit }
  grab { print }
' "$WF")"

if [ -z "$(printf '%s' "$STEP" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: anchor '$ANCHOR' matched but the step body is empty in $WF — the workflow structure changed; update this probe." >&2
  exit 9
fi

# env map entries (indented NAME: value lines inside the step's env: block).
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

# The step's run: — block form (`run: |` until dedent) or single-line form
# (`run: <cmd>` — the remainder is the body). Boundary-limited to the step.
RUNBLOCK="$(printf '%s\n' "$STEP" | awk '
  /^ *run: */ {
    line = $0
    sub(/^ *run: */, "", line)
    if (line ~ /^[>|][+-]?[0-9]*$/) { block = 1; next }
    print line
    exit
  }
  block {
    if (match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/) { exit }
    sub(/^ {10}/, ""); print
  }
')"

if [ -z "$(printf '%s' "$RUNBLOCK" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: no run body found in $WF merge step (single-line run: and block run: | both unmatched) — the workflow structure changed; update this probe." >&2
  exit 9
fi

# Compose the executable body once. DECLARED CLOCK INTERFACE (t1546): the
# body reads the served time by `cat "$MOAI_CLOCK_FILE"` at every deadline
# evaluation, and the deadline from `$merge_deadline` — both EXPORTED so the
# stub child sees them too.
BODY_SCRIPT="$(mktemp /tmp/t1534-e26-XXXXXX)" || harness_fail "mktemp failed for body script"
{
  printf 'export merge_deadline=1140\n'
  printf 'export MOAI_CLOCK_FILE="%s"\n' "$(mktemp /tmp/t1534-e26-clock-XXXXXX)" || exit 4
  printf '%s\n' "$ENV_EXPORTS" \
    | sed -e 's/\${{ secrets.GITHUB_TOKEN }}/dummy-t1546-token/' \
          -e 's/\${{ steps.pr.outputs.branch }}/release\/v3.3.0-probe/' \
          -e 's/\${{ steps.pr.outputs.number }}/1/'
  printf '%s\n' "$RUNBLOCK" \
    | sed -e 's/\${{ steps.pr.outputs.number }}/1/'
} > "$BODY_SCRIPT"
CLOCK_FILE="$(sed -n 's/^export MOAI_CLOCK_FILE="\(.*\)"$/\1/p' "$BODY_SCRIPT")"
printf '1100\n' > "$CLOCK_FILE"

HERE="$(cd "$(dirname "$0")" && pwd)"

# ONE execution: the real body runs while the clock crosses the deadline at
# the FIRST gh contact (the stub advances the served clock then records — so
# any merge call is served post-crossing, and a correct AC-CI-003
# implementation re-checks the head, re-reads the clock (now past the
# deadline) and withholds without merging).
LOG="$(mktemp /tmp/t1534-e26-log-XXXXXX)" || harness_fail "mktemp failed for call log"
BODY_ERR="$(mktemp /tmp/t1534-e26-err-XXXXXX)" || harness_fail "mktemp failed for stderr capture"

body_rc=0
PATH="$HERE/stub-record:$PATH" MOCK_CALL_LOG="$LOG" \
  bash -e "$BODY_SCRIPT" > /dev/null 2>"$BODY_ERR" || body_rc=$?

# P2-L full: a stub log-write failure is a harness failure (exit 4), never an
# observed merge decision.
if [ -f "$BODY_ERR" ] && grep -q 't1546-stub-log-write-failed' "$BODY_ERR" 2>/dev/null; then
  rm -f "$BODY_ERR" "$LOG"
  harness_fail "recording stub could not write its call log"
fi
# P2-L full: bash exit 2 on the composed body is a probe translation bug —
# a harness failure, never a repair signal.
if [ "$body_rc" -eq 2 ]; then
  rm -f "$BODY_ERR" "$LOG"
  harness_fail "composed body failed to parse (exit 2 from bash) — probe translation bug"
fi

# Classify every recorded gh interaction.
total_calls=0
merge_calls=0
head_queries=0
late_calls=0
if [ -f "$LOG" ]; then
  total_calls="$(grep -c 'gh pr' "$LOG" 2>/dev/null)" || total_calls=0
  merge_calls="$(grep -c 'gh pr merge' "$LOG" 2>/dev/null)" || merge_calls=0
  head_queries="$(grep -c 'gh pr view' "$LOG" 2>/dev/null)" || head_queries=0
  late_calls="$(grep -c 'late=yes' "$LOG" 2>/dev/null)" || late_calls=0
fi
for v in total_calls merge_calls head_queries late_calls; do
  eval "[ -n \"\$$v\" ] || $v=0"
done

echo "body_exit=$body_rc gh_interactions=$total_calls head_queries=$head_queries merge_calls=$merge_calls late_calls=$late_calls (served clock crossed 1100 -> 1200 at first gh contact; deadline 1140)"
cat "$LOG" 2>/dev/null || true
rm -f "$BODY_SCRIPT" "$BODY_ERR" "$LOG" "$CLOCK_FILE"

if [ "$body_rc" -ne 0 ]; then
  echo "GUARD MISBEHAVES: the merge-step body exited $body_rc after clean execution — not a recognizable deadline contract shape" >&2
  exit 3
fi
if [ "$late_calls" -gt 0 ]; then
  echo "DEFECT: merge call(s) served AFTER the deadline crossed (late=$late_calls of $merge_calls merge call(s)) — the clock was not re-evaluated before merging"
  exit 0
fi
if [ "$merge_calls" -eq 0 ] && [ "$head_queries" -gt 0 ]; then
  echo "M1 CONTRACT HOLDS: head re-queried, clock observed past the deadline after re-evaluation, merge withheld (P2-R waiting case) — zero late calls"
  exit 2
fi
if [ "$merge_calls" -eq 0 ] && [ "$head_queries" -eq 0 ]; then
  echo "GUARD MISBEHAVES: no gh interaction observed at all — the body neither re-checked the head nor merged" >&2
  exit 3
fi
echo "UNRECOGNIZED SHAPE: merge_calls=$merge_calls late_calls=$late_calls head_queries=$head_queries" >&2
exit 3

if [ "$body_rc" -ne 0 ]; then
  echo "GUARD MISBEHAVES: the merge-step body exited $body_rc after clean execution — not a recognizable deadline contract shape" >&2
  exit 3
fi
if [ "$late_calls" -gt 0 ]; then
  echo "DEFECT: merge call(s) served AFTER the deadline crossed (late=$late_calls of $merge_calls merge call(s)) — the clock was not re-evaluated before merging"
  exit 0
fi
if [ "$merge_calls" -eq 0 ] && [ "$head_queries" -gt 0 ]; then
  echo "M1 CONTRACT HOLDS: head re-queried, clock observed past the deadline after re-evaluation, merge withheld (P2-R waiting case) — zero late calls"
  exit 2
fi
if [ "$merge_calls" -eq 0 ] && [ "$head_queries" -eq 0 ]; then
  echo "GUARD MISBEHAVES: no gh interaction observed at all — the body neither re-checked the head nor merged" >&2
  exit 3
fi
echo "UNRECOGNIZED SHAPE: merge_calls=$merge_calls late_calls=$late_calls head_queries=$head_queries" >&2
exit 3
