# AC-CSRB-002 — RED-now record (t562 M1)

- Command: `go test ./internal/cli/ -run 'TestJudgeCodexSkillEntry_SeparatorConversion' -count=1 -timeout 1800s -v`
- Exit code: `1`
- Tree SHA: `69cfdce6f` (branch `WT-codex-read-inverse`, worktree `.claude/worktrees/t562`,
  measured 2026-09-08 — the absorbed pre-implementation tree; the M2 conversion does not exist yet)
- Final test name pinned (acceptance.md AC-CSRB-002's evidence record):
  **`TestJudgeCodexSkillEntry_SeparatorConversion`** — the name acceptance.md carries as a
  placeholder is the name that shipped; no rename occurred.
- Verbatim output: `red-csrb-002.log` (same directory).

## Why it is red (the stated reason, per verification-completeness §2)

The recorder observes the DECLARED form — `statPath = e.Path` verbatim in
`judgeCodexSkillEntry`'s `codexPathAbsolute` branch — while the test wants
`fromConfigPath(declared, '\\')`. The two strings differ under the injected `\\` separator, so the
conversion absence is exactly what the failure shows:

```
    codex_skills_prune_readback_test.go:86: stat target = "/var/folders/.../gone/SKILL.md", want the converted form "\\var\\folders\\...\\gone\\SKILL.md" (declared "/var/folders/.../gone/SKILL.md")
--- FAIL: TestJudgeCodexSkillEntry_SeparatorConversion (0.00s)
```

Not an impossible red (M2's one-line branch conversion flips it), not a wrong-reason red (the
failing assertion is the stat-target assertion on the conversion the card will add; the Eligible
assertion passed — the recorder returned fs.ErrNotExist, so the gate itself is intact).

M2 owns the fix; this milestone deliberately stops at the RED.
