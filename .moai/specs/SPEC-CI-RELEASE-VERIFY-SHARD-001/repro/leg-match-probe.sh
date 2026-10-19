#!/usr/bin/env bash
# leg-match-probe.sh — does scripts/release.sh's per-OS leg lookup find a
# shard-suffixed Release Verify job?
#
# Extracts the REAL awk lookup expression from scripts/release.sh (the line that
# tests `$1 == leg`) and feeds it job names in `name|conclusion` form, exactly as
# `gh run view --json jobs --jq` renders them in the script.
#
# Usage: bash leg-match-probe.sh <worktree-root>
# Output: one `case -> result` line per case, then a verdict line.
# Exit:  0 = every case behaved as the CURRENT script is expected to behave
#            (exact leg found; shard-suffixed leg NOT found -> the RED state);
#        1 = the shard-suffixed leg WAS found (the repaired state);
#        9 = extraction failure (the awk line is no longer present).
set -u
root="${1:-}"
[ -n "$root" ] || { echo "usage: leg-match-probe.sh <worktree-root>" >&2; exit 9; }
src="$root/scripts/release.sh"
grep -q "\$1 == leg" "$src" || { echo "FAIL: release.sh no longer contains the \$1 == leg lookup (extraction)" >&2; exit 9; }

lookup() { # lookup <leg> <jobs-text>
  printf '%s\n' "$2" | awk -F'|' -v leg="$1" '$1 == leg {print "found|" $2; exit}'
}

exact='Release Verify (ubuntu-latest)|success'
sharded='Release Verify (ubuntu-latest, cli-1)|success
Release Verify (ubuntu-latest, cli-2)|success'
leg='Release Verify (ubuntu-latest)'

r1="$(lookup "$leg" "$exact")"
r2="$(lookup "$leg" "$sharded")"
echo "exact job name        -> ${r1:-<no leg>}"
echo "shard-suffixed names  -> ${r2:-<no leg>}"
if [ -z "$r2" ]; then
  echo "RED: the lookup finds no leg among shard-suffixed job names"
  [ -n "$r1" ] && exit 0
  exit 9
fi
echo "REPAIRED: the lookup finds a leg among shard-suffixed job names"
exit 1
