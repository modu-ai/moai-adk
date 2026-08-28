#!/usr/bin/env bash
# run-scratch generator (not distributed): emits the corpus baseline snapshot
# using the counter command extracted from the B12 clause.
set -u
CLAUSE=.claude/agents/moai/manager-docs.md
CMD=$(awk '/# MOAI-AC-COUNTER-BEGIN/{f=1;next}/# MOAI-AC-COUNTER-END/{f=0}f' "$CLAUSE")
{
  echo "# AC-count corpus baseline — depth-1 glob .moai/specs/*/acceptance.md (_archive excluded)"
  echo "# Regenerate with .moai/reports/t338/run-scratch/gen-baseline.sh; every line is a measurement."
  for f in .moai/specs/*/acceptance.md; do
    out=$(AC_FILE=$f sh -c "$CMD" 2>/tmp/acstderr.$$)
    rc=$?
    tally=$(cat /tmp/acstderr.$$)
    if [ $rc -ne 0 ]; then
      ids=$(printf '%s\n' "$out" | head -1 | sed 's/^AMBIGUOUS //')
      owner=$(basename "$(dirname "$f")")
      echo "$f  HALT $ids  owner=$owner  reason=convention-not-yet-applied-by-owning-SPEC"
    else
      echo "$f  COUNT $out  $tally"
    fi
  done
} > .moai/reports/t338/ac-count-baseline.txt
rm -f /tmp/acstderr.$$
wc -l < .moai/reports/t338/ac-count-baseline.txt
