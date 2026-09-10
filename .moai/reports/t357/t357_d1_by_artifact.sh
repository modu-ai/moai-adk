#!/bin/bash
# t357: D1 per-artifact breakdown over ALL SPEC dirs (no lifecycle-status filter)
# usage: bash .moai/cache/t357_d1_by_artifact.sh <repo-root>
cd "$1" || exit 1
fm_of() { awk 'NR==1&&/^---/{f=1;next} f&&/^---/{exit} f' "$1"; }
d=0; r=0; p=0; a=0
for dir in .moai/specs/SPEC-*/; do
  for art in plan acceptance design research; do
    f="${dir}${art}.md"
    [ -f "$f" ] || continue
    if fm_of "$f" | grep -qE '^status:[[:space:]]'; then
      case "$art" in
        design) d=$((d+1)) ;;
        research) r=$((r+1)) ;;
        plan) p=$((p+1)) ;;
        acceptance) a=$((a+1)) ;;
      esac
    fi
  done
done
echo "HEAD=$(git rev-parse --short HEAD)"
echo "design $d"
echo "research $r"
echo "plan $p"
echo "acceptance $a"
echo "total $((d+r+p+a))"
