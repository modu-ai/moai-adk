#!/bin/bash
# t383 run-phase: derive the missing-target population from the LIVE active index,
# and emit the deterministic M0 sample (every 5th of the sorted set).
#
# Read-only with respect to the memory stores. Writes only to the paths given as
# $1 (missing list) and $2 (sample list); both default under /tmp.
# Run from anywhere. Requires bash (uses $'\n' and arrays), not sh.
#
# Population rule (plan.md §F M0 step 2): the unique missing targets, WHOLE-FILE
# scope — every markdown .md link target on any line shape, not just `^- \[`
# entry lines. This is the population M3 actually copies, and it is the same
# count `moai memory doctor` reports as MEMORY_DANGLING_INDEX_LINK.

set -u

D="$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"
M="$D/MEMORY.md"
OUT_MISSING="${1:-/tmp/t383-missing.txt}"
OUT_SAMPLE="${2:-/tmp/t383-sample.txt}"

# Whole-file unique targets, sorted lexicographically (LC_ALL=C for a stable,
# locale-independent order — the sample indices depend on this order).
export LC_ALL=C

grep -o '](\([^)]*\.md\))' "$M" | sed 's/](//;s/)$//' | sort -u > /tmp/t383-all-targets.$$

: > "$OUT_MISSING"
while IFS= read -r f; do
  if [ ! -f "$D/$f" ]; then
    printf '%s\n' "$f" >> "$OUT_MISSING"
  fi
done < /tmp/t383-all-targets.$$

total=$(grep -c . < /tmp/t383-all-targets.$$)
missing=$(grep -c . < "$OUT_MISSING")

echo "index:                 $M"
echo "unique targets:        $total"
echo "unique missing:        $missing"

# Deterministic sample: every 5th, 1-indexed -> 1, 6, 11, ... 56
: > "$OUT_SAMPLE"
i=0
while IFS= read -r f; do
  i=$((i + 1))
  if [ $(( (i - 1) % 5 )) -eq 0 ]; then
    printf '%d\t%s\n' "$i" "$f" >> "$OUT_SAMPLE"
  fi
done < "$OUT_MISSING"

sampled=$(grep -c . < "$OUT_SAMPLE")
echo "sampled (every 5th):   $sampled"
echo "coverage limit:        indices not reached by the rule are listed below"

# Which indices the rule did not reach (stated as a limit, per AC-MSR-015 item 4).
last_sampled=$(awk -F'\t' 'END{print $1}' "$OUT_SAMPLE")
echo "highest sampled index: $last_sampled  (of $missing)"

rm -f /tmp/t383-all-targets.$$
