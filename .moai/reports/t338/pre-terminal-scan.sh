#!/usr/bin/env bash
# How many multi-domain (citation-axis) acceptance.md files belong to SPECs
# that will still run B12 again? B12 fires only at that SPEC's own sync, so a
# `completed` SPEC never re-counts.
set -u
multi=0; pre=0; comp=0; nostatus=0
for f in .moai/specs/*/acceptance.md; do
  [ -f "$f" ] || continue
  n=$(grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' "$f" | sed -E 's/-[0-9]+$//' | sort -u | wc -l | tr -d ' ')
  [ "$n" -lt 2 ] && continue
  multi=$((multi + 1))
  s=$(sed -n '1,25p' "$(dirname "$f")/spec.md" 2>/dev/null | grep -m1 '^status:' | sed 's/^status:[[:space:]]*//')
  case "$s" in
    completed) comp=$((comp + 1)) ;;
    "")        nostatus=$((nostatus + 1)) ;;
    *)         pre=$((pre + 1)); echo "PRE-TERMINAL [$s] $f" ;;
  esac
done
echo "---"
echo "multi-domain files      : $multi"
echo "  status=completed      : $comp"
echo "  pre-terminal          : $pre"
echo "  no spec.md status     : $nostatus"
