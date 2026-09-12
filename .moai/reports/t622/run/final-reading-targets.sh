#!/bin/sh
# Compares the reading-record target lines taken at M1 with the final (post-M3) targets, ignoring
# the file:line prefix. Read-only; run from .moai/reports/t622/run.
for side in local template; do
  sed -E 's/^[^:]+:[0-9]+: //' "ac028-$side-mode-lines.txt" > "final-$side-ml-m1.body"
  sed -E 's/^[^:]+:[0-9]+: //' "final-$side-028-mode-lines.txt" > "final-$side-ml-final.body"
  diff "final-$side-ml-m1.body" "final-$side-ml-final.body" > "final-$side-mode-lines-vs-m1.diff"
  echo "$side mode-lines body diff exit=$?"
  sed -E 's/^[^:]+:[0-9]+: //' "ac026-$side-merge-lines.txt" > "final-$side-mg-m1.body"
  sed -E 's/^[^:]+:[0-9]+: //' "final-$side-026-merge-lines.txt" > "final-$side-mg-final.body"
  diff "final-$side-mg-m1.body" "final-$side-mg-final.body" > "final-$side-merge-lines-vs-m1.diff"
  echo "$side merge-lines body diff exit=$?"
  echo "$side final mode-line coordinates: $(cut -d: -f1,2 "final-$side-028-mode-lines.txt" | tr '\n' ' ')"
  echo "$side final merge-line coordinates: $(cut -d: -f1,2 "final-$side-026-merge-lines.txt" | tr '\n' ' ')"
done
