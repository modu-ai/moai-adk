#!/bin/bash
# t383 iteration-3 independent re-measurement of the N1 line-population relation.
# Read-only. Writes nothing. Run from anywhere.
D="$HOME/.moai/claude-profiles/moai-adk/projects/-Users-goos-MoAI-moai-adk-go/memory"
M="$D/MEMORY.md"

echo "bytes: $(wc -c < "$M" | tr -d ' ')   entry-lines: $(grep -c '^- \[' "$M")"

allu=$(grep -o '](\([^)]*\.md\))' "$M" | sed 's/](//;s/)$//' | sort -u)
entu=$(grep '^- \[' "$M" | grep -o '](\([^)]*\.md\))' | sed 's/](//;s/)$//' | sort -u)

echo "whole-file occurrences: $(grep -o '](\([^)]*\.md\))' "$M" | wc -l | tr -d ' ')"
echo "whole-file unique     : $(printf '%s\n' "$allu" | grep -c .)"
echo "entry-line occurrences: $(grep '^- \[' "$M" | grep -o '](\([^)]*\.md\))' | wc -l | tr -d ' ')"
echo "entry-line unique     : $(printf '%s\n' "$entu" | grep -c .)"

missall=""
for f in $allu; do [ -f "$D/$f" ] || missall="$missall$f"$'\n'; done
missent=""
for f in $entu; do [ -f "$D/$f" ] || missent="$missent$f"$'\n'; done

na=$(printf '%s' "$missall" | grep -c .)
ne=$(printf '%s' "$missent" | grep -c .)
echo "missing unique, whole-file: $na"
echo "missing unique, entry-line: $ne"
echo "missing ONLY outside entry lines: $((na - ne))"

echo "--- entry lines carrying >=1 missing target ---"
grep -n '^- \[' "$M" | while IFS= read -r line; do
  n=${line%%:*}
  body=${line#*:}
  hit=0
  for t in $(printf '%s' "$body" | grep -o '](\([^)]*\.md\))' | sed 's/](//;s/)$//'); do
    [ -f "$D/$t" ] || hit=1
  done
  [ "$hit" = 1 ] && echo "$n"
done | wc -l | tr -d ' '

echo "--- the outside-only set ---"
printf '%s' "$missall" | grep -v -x -F -f <(printf '%s' "$missent")
