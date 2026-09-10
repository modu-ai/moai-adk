#!/usr/bin/env bash
# t492 — for each always-loaded rule file, list the sibling companion files that share its
# basename stem (-detail / -reference / -examples / -catalogue / any suffix), with sizes.
set -u
ROOT="$1"
while IFS=$'\t' read -r _b _t rel; do
  case "$rel" in
    .claude/rules/moai/*) ;;
    *) continue ;;
  esac
  f="$ROOT/$rel"
  dir=$(dirname "$f")
  stem=$(basename "$f" .md)
  own=$(wc -c < "$f" | tr -d ' ')
  comp_bytes=0
  comp_names=""
  for c in "$dir/$stem"-*.md; do
    [ -f "$c" ] || continue
    cb=$(wc -c < "$c" | tr -d ' ')
    comp_bytes=$((comp_bytes + cb))
    comp_names="$comp_names $(basename "$c")($cb)"
  done
  if [ -z "$comp_names" ]; then comp_names=" NONE"; fi
  printf '%s\tstub=%s\tcompanion_total=%s\t%s\n' "$(basename "$f")" "$own" "$comp_bytes" "$comp_names"
done < "$2"
