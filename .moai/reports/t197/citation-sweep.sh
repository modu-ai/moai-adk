#!/bin/bash
# Citation sweep for t197.
#
#   bash .moai/reports/t197/citation-sweep.sh
#
# Every claim that cites something outside its own file is a citation, and a
# citation goes stale the moment its target is regenerated. Twice now a fix
# touched one artifact and left the files that cite it behind — the baseline pin
# on the launcher, and INIT-001's plan.md. This script names every citation so
# they can be swept in the same commit as the change that invalidates them.
#
# Read-only.

set -uo pipefail
cd "$(git rev-parse --show-toplevel)" || exit 1

SPECS=.moai/specs/SPEC-CODEX-LAUNCHER-001
SPECS2=.moai/specs/SPEC-CODEX-INIT-001
REPORTS=.moai/reports/t197
SCOPE="$SPECS $SPECS2 $REPORTS"

hdr() { echo; echo "=== $1 ==="; }

# A citation is a LIVE claim about the current state. A history row or a record
# of what was corrected names an OLD value on purpose — flagging those would
# train the reader to ignore the sweep, so they are excluded by pattern:
#   - HISTORY table rows (start with a version cell)
#   - split-record's correction ledger (names old -> new)
#   - verdict files (each pins the tree it judged, by design)
NOT_LIVE='verdict-|^[^:]*:[0-9]+:\| [0-9]+\.[0-9]+\.[0-9]+ \||낡은 핀'

hdr "1. baseline pins (must all equal the transcript's own HEAD)"
echo "-- transcript records:"
grep -A1 'git rev-parse --short HEAD' "$REPORTS/probe-output.txt" | sed -n '2p'
echo "-- LIVE pins cited in prose (full lines, so history rows can be excluded):"
# 7b217da7c is the t88 ancestor the probe checks, not a baseline pin — expected.
grep -rnE '\b[0-9a-f]{9}\b' $SCOPE --include='*.md' \
  | grep -vE "$NOT_LIVE" \
  | grep -v '7b217da7c' \
  | sed -E 's/^([^:]+:[0-9]+):.*(\b[0-9a-f]{9}\b).*/\1  ->  \2/'

hdr "2. figures that must trace to the transcript"
grep -rnE '[0-9]+초|[0-9]+s\b|[0-9]+,[0-9]{3}' $SCOPE --include='*.md' | grep -vE "$NOT_LIVE" | grep -vE 'timeout|600s|^.*\| 0\.[0-9]'

hdr "3. transcript line citations (L<n>-<n>) — verify against current transcript length"
echo "-- transcript is $(wc -l < "$REPORTS/probe-output.txt") lines"
grep -rnoE 'L[0-9]+-[0-9]+' $SCOPE --include='*.md' | grep -vE "$NOT_LIVE" | head -40

hdr "4. REQ/AC ids cited outside their own SPEC"
grep -rn 'REQ-CL-\|AC-CL-' "$SPECS2" --include='*.md' || echo "(none: INIT cites no launcher ids)"
grep -rn 'REQ-CI-\|AC-CI-' "$SPECS" --include='*.md' || echo "(none: launcher cites no INIT ids)"

hdr "5. cross-artifact design claims inside each SPEC (spec <-> plan <-> acceptance)"
for d in "$SPECS" "$SPECS2"; do
  echo "-- $d"
  grep -rnoE '(AC|REQ)-C[LI]-[0-9]+' "$d/plan.md" | sort -t: -k3 -u | head -20
done

hdr "6. counts asserted in prose (cells, states, fixtures)"
grep -rnE '[0-9]+칸|[0-9]+상태|[0-9]종|[0-9]+ REQ|[0-9]+ AC' $SCOPE --include='*.md' | grep -vE "$NOT_LIVE" | head -30

echo
echo "=== sweep done — every line above is a citation that a change may invalidate ==="
