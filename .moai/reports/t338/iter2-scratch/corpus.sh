#!/usr/bin/env bash
# iter-2 audit scratch: run the adjacency counter over the depth-1 corpus and
# report every file that would HALT (ambiguous) or record an excluded id today.
n=0; halt=0; excl=0
for f in .moai/specs/*/acceptance.md; do
  n=$((n+1))
  out=$(python3 .moai/reports/t338/iter2-scratch/counter.py "$f" adj 2>&1)
  rc=$?
  if [ $rc -ne 0 ]; then halt=$((halt+1)); echo "HALT $f :: $out"; fi
  case "$out" in *"excluded=0"*) ;; *) if [ $rc -eq 0 ]; then excl=$((excl+1)); echo "EXCL $f :: $out"; fi;; esac
done
echo "files=$n  halting=$halt  files-with-excluded=$excl"
