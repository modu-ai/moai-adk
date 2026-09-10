# t661 slot step 5 - home fingerprint comparison (2026-09-11)

- sha256 of 7 files: diff home-before-sha.txt home-after-sha.txt -> exit 0 (identical)
- stat mtime/size: diff home-before-stat.txt home-after-stat.txt -> exit 1, sole hunk `8d7 < stat: /Users/goos/.claude/hooks/moai: stat: No such file or directory`
  - attribution: before-stat.txt has 8 lines (the before capture appended the hooks-absence stat error to it), after-stat.txt has 7
  - lines 1-7 of before vs after-stat.txt -> diff exit 0 (mtime and size of all 7 files identical)
  - line 8 of before vs home-after-hooks.txt -> diff exit 0 (same absence line)
- ~/.claude/hooks/moai: absent after (stat exit 1); ~/.claude/hooks: absent after (ls exit 1); control `ls -d ~/.claude` exit 0
- nothing was restored

Verdict: no content, mtime, or size change in the 7 files; hooks dir still absent. The literal stat diff exit 1 is a capture-format difference, reported to the lead as-is.
