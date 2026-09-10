#!/bin/bash
# t487 M1 sweep — read-only shape check of .claude/settings.json across worktrees.
# Zero git commands: iterates the pre-captured tree-list.txt (produced separately
# by `git worktree list --porcelain` in the t487 own tree).
DIRTY=b669972dc738d1bf925281dcc90f152e
SWEEP_DIR=/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t487/.moai/reports/t487/sweep
OUT=$SWEEP_DIR/sweep.tsv
printf 'path\tmd5\task\tmatcher_narrow\tmatcher_extended\tkey_order_match\tmissing_keys\tsto_count\tfile_mtime\tdir_birthtime\tnotes\n' > "$OUT"
while IFS= read -r t; do
  f="$t/.claude/settings.json"
  if [ -f "$f" ]; then
    md5=$(md5 -q "$f")
    ask=$(jq -r 'if (.permissions.ask|type)=="array" then (.permissions.ask|length) else "null" end' "$f" 2>/dev/null || echo JQERR)
    narrow=$(grep -c '"matcher": "Write|Edit|MultiEdit"' "$f")
    extended=$(grep -c 'Write|Edit|MultiEdit|EnterWorktree|ExitWorktree' "$f")
    ko=$(jq -r 'keys_unsorted|join(",")' "$f" 2>/dev/null || echo JQERR)
    if [ "$ko" = '$schema,respectGitignore,cleanupPeriodDays,skillListingBudgetFraction,env,attribution,permissions,hooks,statusLine,outputStyle,showThinkingSummaries' ]; then kom=Y; else kom=N; fi
    mk=$(jq -r 'if (has("model") or has("includeGitInstructions") or has("plansDirectory")) then "present" else "absent" end' "$f" 2>/dev/null || echo JQERR)
    sto=$(grep -c 'status-transition-ownership' "$f")
    ft=$(stat -f '%Sm' -t '%Y-%m-%d %H:%M:%S' "$f")
    db=$(stat -f '%SB' -t '%Y-%m-%d %H:%M:%S' "$t" 2>/dev/null || echo NOSTAT)
    note=""
    [ "$md5" = "$DIRTY" ] && note="DIRTY_MD5_MATCH"
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$t" "$md5" "$ask" "$narrow" "$extended" "$kom" "$mk" "$sto" "$ft" "$db" "$note" >> "$OUT"
  else
    printf '%s\tNOFILE\n' "$t" >> "$OUT"
  fi
done < "$SWEEP_DIR/tree-list.txt"
wc -l < "$OUT"
