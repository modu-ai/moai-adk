#!/usr/bin/env bash
# shard-cover.sh — exact-cover listing for the internal/cli race-shard selectors.
#
# Lists every top-level internal/cli test once (go test -list), lists it again
# per selector, and proves the four selector listings partition the full list:
# same line count, no overlap, no gap. The selectors are read from a workflow
# file (default ci.yml) so the same script judges ci.yml today and the release
# workflow after the run phase.
#
# Usage: bash shard-cover.sh <worktree-root> <outdir> [workflow-file]
# Exit:  0 exact cover; 1 gap/overlap or empty sweep; 9 usage/extraction error.
set -u
root="${1:-}"; out="${2:-}"; wf="${3:-.github/workflows/ci.yml}"
[ -n "$root" ] && [ -n "$out" ] || { echo "usage: shard-cover.sh <root> <outdir> [workflow]" >&2; exit 9; }
cd "$root" || exit 9
mkdir -p "$out"

# Selectors in file order: the -run '<sel>' argument of every internal/cli leg
# (rosterguard uses ./internal/harness/rosterguard/ and is excluded here).
grep -- '-race' "$wf" | grep -o "\-run '[^']*' ./internal/cli/" | sed -e "s/^-run '//" -e "s/' .\/internal\/cli\/\$//" > "$out/selectors.txt"
n=$(wc -l < "$out/selectors.txt" | tr -d ' ')
if [ "$n" -eq 0 ] && command -v yq >/dev/null 2>&1; then
  # Matrix form: the release workflow carries one include entry per shard with
  # `pkg` and `run` keys; the internal/cli shards are the entries whose pkg is
  # ./internal/cli/ . An unparsable or key-less workflow yields zero lines and
  # falls through to the count check below (exit 9), never to a pass.
  yq -r '.jobs.full-matrix-test.strategy.matrix.include[]? | select(.pkg == "./internal/cli/") | .run' "$wf" 2>/dev/null > "$out/selectors.txt"
  n=$(wc -l < "$out/selectors.txt" | tr -d ' ')
fi
[ "$n" -eq 4 ] || { echo "FAIL: expected 4 internal/cli selectors in $wf, found $n" >&2; exit 9; }

go test -list '.*' ./internal/cli/ 2> "$out/full.err" | grep '^Test' | sort > "$out/full.sorted"
total=$(wc -l < "$out/full.sorted" | tr -d ' ')
[ "$total" -gt 0 ] || { echo "FAIL: empty full listing (empty sweep asserts nothing)"; exit 1; }

: > "$out/union.raw"
i=0
while IFS= read -r sel; do
  i=$((i+1))
  go test -list "$sel" ./internal/cli/ 2>> "$out/full.err" | grep '^Test' > "$out/s$i.txt"
  echo "selector $i  $sel  -> $(wc -l < "$out/s$i.txt" | tr -d ' ') names"
  cat "$out/s$i.txt" >> "$out/union.raw"
done < "$out/selectors.txt"

sort "$out/union.raw" > "$out/union.sorted"
echo "full=$total union=$(wc -l < "$out/union.sorted" | tr -d ' ') union_unique=$(sort -u "$out/union.sorted" | wc -l | tr -d ' ')"
if diff "$out/full.sorted" "$out/union.sorted" > "$out/diff.txt"; then
  echo "EXACT COVER: full list == union of the four selectors (no gap, no overlap)"
  exit 0
fi
echo "NOT AN EXACT COVER (first differences):"
head -20 "$out/diff.txt"
exit 1
