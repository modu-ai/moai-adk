#!/usr/bin/env bash
# iter-2 audit scratch: D7 cross-SPEC reconciliation over the 4 artifacts
cd .moai/specs/SPEC-AC-COUNT-DISCRIMINATOR-001 || exit 1
grep -hoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md plan.md acceptance.md progress.md | sort -u | while read -r SID; do
  d="../../../.moai/specs/$SID/spec.md"
  d="../$SID/spec.md"
  if [ -f "$d" ]; then
    st=$(grep -m1 '^status:' "$d" | cut -d: -f2 | tr -d ' ')
    case "$st" in
      retired|superseded|archived) echo "BLOCKING $SID status=$st";;
      *) echo "ok       $SID status=$st";;
    esac
  else
    echo "SHOULD   $SID NOT FOUND"
  fi
done
