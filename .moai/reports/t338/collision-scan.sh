#!/usr/bin/env bash
# collision-scan.sh — count lines carrying >=2 DISTINCT AC prefixes across the
# depth-1 acceptance.md corpus. This is the mechanical shape of the
# live-and-non-live-identifier-on-one-line collision (D1).
#
# Run from the repository root of this worktree.
lines=0
files=0
for f in .moai/specs/*/acceptance.md; do
  n=$(awk '{
    delete p
    s = $0
    c = 0
    while (match(s, /AC-([A-Z0-9]+-)*[0-9]+/)) {
      id = substr(s, RSTART, RLENGTH)
      s  = substr(s, RSTART + RLENGTH)
      pre = id; sub(/[0-9]+$/, "", pre)
      if (!(pre in p)) { p[pre] = 1; c++ }
    }
    if (c >= 2) hit++
  } END { print hit + 0 }' "$f")
  if [ "$n" -gt 0 ]; then
    lines=$((lines + n))
    files=$((files + 1))
  fi
done
echo "lines carrying >=2 distinct AC prefixes: $lines"
echo "files containing such a line          : $files"
