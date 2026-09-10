#!/usr/bin/env bash
set -u
SWEEP='AC-([A-Z0-9]+-)*[0-9]+'
MARK='폐기|retired|RETIRED|superseded|SUPERSEDED|철회|withdrawn'
total=0; affected=0; overcount_ids=0
for f in .moai/specs/*/acceptance.md; do
  [ -f "$f" ] || continue
  total=$((total + 1))
  ids=$(grep -oE "$SWEEP" "$f" | sort -u)
  [ -z "$ids" ] && continue
  hit=0
  for id in $ids; do
    lines=$(grep -nE "(^|[^A-Z0-9-])${id}([^0-9-]|$)" "$f")
    [ -z "$lines" ] && continue
    clean=$(printf '%s\n' "$lines" | grep -vE "$MARK")
    if [ -z "$clean" ]; then
      hit=$((hit + 1)); overcount_ids=$((overcount_ids + 1))
      echo "OVERCOUNT $f  $id"
    fi
  done
  [ "$hit" -gt 0 ] && affected=$((affected + 1))
done
echo "---"
echo "acceptance.md scanned : $total"
echo "files over-counted    : $affected"
echo "phantom AC ids        : $overcount_ids"
