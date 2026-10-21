#!/usr/bin/env bash
# leg-match-probe.sh — does scripts/release.sh's per-OS leg lookup find a
# shard-suffixed Release Verify job?
#
# FIDELITY, stated plainly: this probe does NOT run scripts/release.sh. It reads
# the real file, extracts (a) the `for leg in ...` list line and (b) the awk
# program text of the lookup from the non-comment `awk -F'|' -v leg=` line, and
# runs THAT extracted program over fixture job lists. It therefore follows edits
# to the real lookup (a rewritten lookup that no longer has that shape makes the
# extraction fail, exit 9), but it is a probe of the extracted expression, not
# of the whole script. The whole-script RED/green pair is the sharded all-green
# scenario added to matrixScenarios() in internal/template/release_script_test.go
# (plan.md M2, AC-RVS-011): run through the existing gh double, it is refused by
# the current script and allowed by the repaired one.
#
# Usage: bash leg-match-probe.sh <worktree-root>
# Exit:  0 = RED state (exact leg found; shard-suffixed leg NOT found);
#        1 = the shard-suffixed leg WAS found (the lookup changed);
#        9 = extraction failure (the lookup no longer has the exact-name shape).
set -u
root="${1:-}"
[ -n "$root" ] || { echo "usage: leg-match-probe.sh <worktree-root>" >&2; exit 9; }
src="$root/scripts/release.sh"

legs_line="$(grep -n '^[[:space:]]*for leg in "Release Verify (' "$src" | head -1)"
awk_line="$(grep -v '^[[:space:]]*#' "$src" | grep -F "awk -F'|' -v leg=" | head -1)"
[ -n "$legs_line" ] && [ -n "$awk_line" ] || { echo "FAIL: exact-name lookup shape not found in $src (extraction)" >&2; exit 9; }
prog="$(printf '%s\n' "$awk_line" | sed -n "s/.*-v leg=\"\\\$leg\" '\\([^']*\\)'.*/\\1/p")"
[ -n "$prog" ] || { echo "FAIL: awk program text not extractable from: $awk_line" >&2; exit 9; }
echo "extracted legs line : ${legs_line%%$'\n'*}"
echo "extracted awk prog  : $prog"

lookup() { # lookup <leg> <jobs-text>
  printf '%s\n' "$2" | awk -F'|' -v leg="$1" "$prog"
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
