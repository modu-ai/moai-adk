#!/usr/bin/env bash
# iter-2 audit scratch: per-file collision-line counts (same awk as collision-scan.sh)
for f in .moai/specs/*/acceptance.md; do
  n=$(awk '{
    delete p; s=$0; c=0
    while (match(s, /AC-([A-Z0-9]+-)*[0-9]+/)) {
      id=substr(s,RSTART,RLENGTH); s=substr(s,RSTART+RLENGTH)
      pre=id; sub(/[0-9]+$/,"",pre)
      if(!(pre in p)){p[pre]=1;c++}
    }
    if(c>=2) hit++
  } END{print hit+0}' "$f")
  if [ "$n" -gt 0 ]; then echo "$n $f"; fi
done | sort -rn
