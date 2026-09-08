#!/bin/bash
# t487 sweep verification: for each worktree, compare working-copy settings.json
# md5 against the tracked blob md5 at that tree's HEAD. All git invocations are
# read-only object reads (`git show <sha>:<path>`) executed from the t487 tree.
cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t487 || exit 1
OUT=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t487/.moai/reports/t487/sweep/blob-check.tsv
printf 'path\twork_md5\thead_blob_md5\tverdict\n' > "$OUT"
while IFS=$'\t' read -r t sha; do
  f="$t/.claude/settings.json"
  [ -f "$f" ] || continue
  wm=$(md5 -q "$f")
  bm=$(git show "$sha:.claude/settings.json" 2>/dev/null | md5 -q)
  if [ "$wm" = "$bm" ]; then v=CLEAN_MATCHES_HEAD; else v="DIRTY_VS_HEAD"; fi
  printf '%s\t%s\t%s\t%s\n' "$t" "$wm" "$bm" "$v" >> "$OUT"
done < /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t487/.moai/reports/t487/sweep/tree-heads.txt
cut -f4 "$OUT" | sort | uniq -c
