#!/bin/sh
# t1534 repro E26 — deadline-absence probe (AC-CI-005 RED addendum; auditor
# iteration-2 mutant shape: declared_deadline=1140, merge_at=1141 -> exit 0).
#
# Extracts the merge step's `if:` condition from the LIVE auto-merge.yml at run
# time, proves the condition references no deadline term, then evaluates the
# guard conjunction with all three outputs true at merge_at = deadline + 1 —
# the merge proceeds because no deadline comparison exists anywhere in the
# current guard.
#
# RED expectation (pre-repair): exit 0, "merge proceeds … deadline never
# evaluated". GREEN expectation (post-M1): the extracted condition carries a
# deadline term and the probe's evaluation at deadline+1 withholds the merge.
set -eu
WF=".github/workflows/auto-merge.yml"

# The merge step's condition is a folded scalar (`if: >` with the conjunction
# on the following, more-indented lines) — the extraction must follow the fold,
# not stop at the `if:` line (first draft grabbed only ">" and mis-observed).
COND="$(awk '
  index($0, "name: Auto merge PR") { grab = 1; next }
  grab && /^ *if: */ {
    line = $0
    sub(/^ *if: */, "", line)
    gsub(/^ +| +$/, "", line)
    # Empty remainder, or a block/folded scalar indicator (>, |, >-, |-, >2 …)
    # means the condition lives on the following, more-indented lines.
    if (line == "" || line ~ /^[>|][+-]?[0-9]*$/) { fold = 1; next }
    print line; exit
  }
  grab && fold {
    if (match($0, /^ */) && RLENGTH < 10 && $0 !~ /^ *$/) { exit }
    sub(/^ */, ""); print
  }
' "$WF")"

if [ -z "$(printf '%s' "$COND" | tr -d '[:space:]')" ]; then
  echo "repro extraction failed: merge-step if: condition not found in $WF — the workflow structure changed; update this probe's anchor." >&2
  exit 9
fi

echo "merge-step condition (extracted):"
printf '%s\n' "$COND"

case "$COND" in
  *deadline*)
    echo "condition references a deadline: YES"
    exit 1
    ;;
  *)
    echo "condition references a deadline: NO"
    ;;
esac

# Tie the guard evaluation to the extracted text: the known three-output
# conjunction shape must be present, or this probe's evaluation is invalid.
case "$COND" in
  *steps.checks.outputs.should_merge*\&\&*steps.head.outputs.ok*\&\&*steps.coderabbit.outputs.ok*) ;;
  *) echo "unexpected condition shape — update this probe's guard evaluation" >&2; exit 9 ;;
esac

declared_deadline=1140
merge_at=1141
checks_should_merge=true
head_ok=true
coderabbit_ok=true

# Evaluate the extracted conjunction's guard shape with all outputs true.
if [ "$checks_should_merge" = "true" ] && [ "$head_ok" = "true" ] && [ "$coderabbit_ok" = "true" ]; then
  echo "merge proceeds at merge_at=$merge_at (declared_deadline=$declared_deadline) — deadline never evaluated"
  exit 0
fi
echo "merge withheld"
exit 1
