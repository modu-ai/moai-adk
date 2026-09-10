#!/usr/bin/env bash
# Upper-bound scan: acceptance.md files carrying >=2 distinct AC domain prefixes.
# A multi-domain file is a CANDIDATE for foreign-identifier citation, not proof of it —
# a genuinely multi-domain SPEC (two REQ families in one SPEC) reads the same way.
set -u
n_multi=0
total=0
for f in .moai/specs/*/acceptance.md; do
  [ -f "$f" ] || continue
  total=$((total + 1))
  n=$(grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' "$f" | sed -E 's/-[0-9]+$//' | sort -u | wc -l | tr -d ' ')
  if [ "$n" -ge 2 ]; then
    n_multi=$((n_multi + 1))
    echo "MULTIDOMAIN $n $f"
  fi
done
echo "---"
echo "acceptance.md scanned      : $total"
echo "files with >=2 AC prefixes : $n_multi"
