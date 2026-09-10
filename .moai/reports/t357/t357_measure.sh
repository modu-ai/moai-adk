#!/bin/bash
# t357: Tier L 6-artifact status-transition contract gap — corpus measurement
# usage: t357_measure.sh <repo-root> <out.tsv>
# columns: id  spec_status  tier  design  research  plan  acceptance
cd "$1" || exit 1
out="$2"
: > "$out"
fm_of() {
  awk 'NR==1&&/^---/{f=1;next} f&&/^---/{exit} f' "$1"
}
for d in .moai/specs/SPEC-*/; do
  id=$(basename "$d")
  spec="${d}spec.md"
  if [ ! -f "$spec" ]; then
    printf '%s\tNOSPEC\t-\t-\t-\t-\t-\n' "$id" >> "$out"
    continue
  fi
  fm=$(fm_of "$spec")
  st=$(printf '%s\n' "$fm" | sed -n 's/^status:[[:space:]]*//p' | head -1 | tr -d '"' | tr -d "'" | tr -d '\r')
  tier=$(printf '%s\n' "$fm" | sed -n 's/^tier:[[:space:]]*//p' | head -1 | tr -d '"' | tr -d "'" | tr -d '\r')
  [ -z "$st" ] && st=NONE
  [ -z "$tier" ] && tier=NONE
  row=""
  for a in design research plan acceptance; do
    f="${d}${a}.md"
    if [ ! -f "$f" ]; then
      row="${row}	-"
      continue
    fi
    afm=$(fm_of "$f")
    if [ -z "$afm" ]; then
      row="${row}	NOFM"
      continue
    fi
    ast=$(printf '%s\n' "$afm" | sed -n 's/^status:[[:space:]]*//p' | head -1 | tr -d '"' | tr -d "'" | tr -d '\r')
    [ -z "$ast" ] && ast=NOSTATUS
    row="${row}	${ast}"
  done
  printf '%s\t%s\t%s%s\n' "$id" "$st" "$tier" "$row" >> "$out"
done
