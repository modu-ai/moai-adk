# F1 correction record + t432 integrity note — SPEC-CODEMAPS-ACCURACY-001 (t304)

Date: 2026-09-02 · Recording lane: t304 (`.claude/worktrees/t304`, branch `WT-codemaps-accuracy`)

## F1 — t432 report §3.1 count mismatch (26 vs 27)

**Claim**: the t432 verification report's §3.1 heading states "26 items" while the table it
introduces carries 27 rows, and the report's own §Gaps line (t432 report line 251) says
"27 items". The heading count is wrong; the table and the Gaps line are right.

**Evidence** (read-only, this run, from the frozen lane-7 tree):

```
sed -n '194p;251p' /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t432/.moai/reports/t432/codemaps-accuracy-verification.md
### §3.1 entry-points.md·data-flow.md citation identifier full-census hit/miss table (26 items)
**Gaps**: none — the full census of function·method·type identifiers in entry-points.md·data-flow.md (27 items) + the full bidirectional census of docs-truth §1 (12 rows).
```

**Disposition**: the t432 evidence file is lane-7's local-only artifact on a frozen tree — per
REQ-CMA-009 this card does NOT write there. The correction is recorded HERE and reported to the
lead; the t432 tree stays untouched.

## t432 worktree integrity (this card's writes: 0)

- Worktree HEAD ref: `refs/heads/WT-codemaps-refresh` (read from
  `.git/worktrees/t432/HEAD`; loose ref read `WT-codemaps-refresh` → `ceec857bb2af…` —
  recorded as this run's observed state; this lane holds no earlier same-session measurement,
  so no unchanged-claim is made about the tip).
- M2-gate ancestry measurement (this run, from the t304 worktree):
  `git merge-base --is-ancestor 7e1c4d94f origin/develop` → exit 0 (MERGED) — the t432
  regeneration the dispatch keyed on is integrated.
- Writes from this lane to the t432 tree: **none.** Every file write, edit, commit, and
  temp artifact of this card targeted `.claude/worktrees/t304/**` or `/tmp/**`; the only
  t432-path operations were Read/sed reads. Direct `git -C <t432> status` was refused by the
  worktree-session guard (documented behavior — cross-tree `git -C` from an isolated
  worktree session is not an authorized path), so the integrity observation rests on the
  ref-file reads above and this lane's operation history rather than a status listing.
