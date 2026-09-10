#!/usr/bin/env bash
# t492 C2 — apply the §2.1 relocation to a (stub, companion) pair. Extracts the blocks from the
# stub being processed, so a neutralized template mirror keeps its own neutralized wording.
# Usage: bash build-mirror.sh <stub> <companion> <header-file> <pointer-file>
set -eu
S="$1"; D="$2"; HDR="$3"; PTR="$4"
T="$(dirname "$S")/.t492.tmp"

# 1. extract the two blocks from THIS stub (before cutting)
sed -n '54,76p'   "$S" > "$T.blockA"
sed -n '120,152p' "$S" > "$T.blockB"

# 2. companion: body(1-105) + header + blocks + trailer(last 5)
head -n 105 "$D" >  "$T.d"
cat "$HDR"       >> "$T.d"
cat "$T.blockA"  >> "$T.d"
printf '\n'      >> "$T.d"
cat "$T.blockB"  >> "$T.d"
tail -n 5 "$D"   >> "$T.d"
mv "$T.d" "$D"
sed -i '' '106,$ s/^#### /### /' "$D"

# 3. stub: 1-53 + pointer + 77-119 + 153-EOF
sed -n '1,53p'   "$S" >  "$T.s"
cat "$PTR"            >> "$T.s"
sed -n '77,119p' "$S" >> "$T.s"
sed -n '153,$p'  "$S" >> "$T.s"
mv "$T.s" "$S"
sed -i '' 's/so the predicate below is written for coordinates generally/so the predicate is written for coordinates generally/' "$S"

rm -f "$T.blockA" "$T.blockB"
wc -c "$S" "$D"
