#!/bin/bash
# per-phrase presence check against a target file
target="$1"
while IFS= read -r p; do
  n=$(grep -cF -- "$p" "$target")
  printf '%-45s %s\n' "$n" "$p"
done < "$(dirname "$0")/phrases.txt"
