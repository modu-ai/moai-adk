# SPEC-ERA-H3-NARROWING-001 — sync-phase close verdict (t382)

Worktree `.claude/worktrees/t382`, branch `WT-era-plan-phase`. Sync close performed 2026-09-01 by manager-docs.

## Claim

- SPEC-ERA-H3-NARROWING-001 (Tier S, card t382) is closed: spec.md frontmatter `status: in-progress → completed` rides the single sync commit `123014952` (3-phase close; `completed` transition merged into the sync commit, no separate Mx commit).
- progress.md §E.4 Sync-phase Audit-Ready Signal written; §E.3 `run_commit_sha` backfilled with the measured tree `c0cdb2fd3`; §E.4 `sync_commit_sha` backfilled in follow-up commit `45a44b01d` (a commit cannot name its own hash).
- No CHANGELOG / README / docs-site / template-mirror emission — per the repo's card-flow precedent (t376: zero doc files; sync audit did not require them). This is a sweep result, not a skip: this change (era classification predicate) is described in none of those surfaces, and run-phase files (`internal/spec/**` ×3, `.claude/rules/local/**` ×1) have no template mirror.
- Final AC state is 7 PASS / 1 FAIL (AC-EH3-007, 명제 3 — one false-positive drift row whose cause is a pre-existing defect, per §E.2). The SPEC closes with that FAIL recorded, not reinterpreted.

## Evidence

```
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t382 log --oneline -3
45a44b01d docs(SPEC-ERA-H3-NARROWING-001): backfill sync_commit_sha in progress.md §E.4 (t382)
123014952 docs(SPEC-ERA-H3-NARROWING-001): sync-phase artifacts — 3-phase close (t382)
987cb8ae4 fix(SPEC-ERA-H3-NARROWING-001): decide REQ-004 by symbol boundary, not by file count (t382)
```

```
$ sed -n '1,10p' .moai/specs/SPEC-ERA-H3-NARROWING-001/spec.md   (after edits, at commit 123014952)
status: completed
updated: 2026-09-01
```

```
$ grep -n 'run_commit_sha\|sync_commit_sha' .moai/specs/SPEC-ERA-H3-NARROWING-001/progress.md   (after commit 45a44b01d)
101:run_commit_sha: c0cdb2fd3   # 측정 트리(§E.2 첫 줄) — 증거가 실제로 잰 트리
145:sync_commit_sha: 123014952  # backfilled in a follow-up commit (a commit cannot name its own hash)
```

```
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t382 status --porcelain   (after both commits)
(no output — clean)
```

```
$ git -C ... show --stat 123014952   (from commit output)
2 files changed, 47 insertions(+), 4 deletions(-)
   .moai/specs/SPEC-ERA-H3-NARROWING-001/{progress.md, spec.md}
```

## Baseline-attribution

All commands above were run in this session against this worktree (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t382`, branch `WT-era-plan-phase`) at HEAD `45a44b01d`. The run-phase evidence figures quoted in §E.4 are NOT this session's measurements — they are citations of run-phase measurements taken at tree `c0cdb2fd3` (recorded in progress.md §E.2, written by run-phase). The `run_commit_sha: c0cdb2fd3` backfill value was determined from §E.2's first line ("측정 트리 `c0cdb2fd3`"), not from HEAD.

## Gaps — what this session did NOT observe

- **No tests, build, vet, or lint re-run.** Sync-phase touched only markdown artifacts; §E.2/§E.3's figures (coverage 89.6%, lint 0, cross-platform rc 0) are run-phase measurements at `c0cdb2fd3`, not sync-phase re-measurements.
- **No CI verdict.** Nothing was pushed; no CI exists for these commits. Integration (merge window, absorption of the recorded `5 6` divergence vs `origin/develop=297a21ea7`, merged-tree re-measurement) remains the lead's domain.
- **No sync-auditor run.** The independent 4-dimension sync audit has not been executed against this close; this document is the manager-docs close verdict only.
- **AC-EH3-007 remains FAIL** (1 false-positive drift row, cause pre-existing per `m3-drift-row-judgment.md`). Follow-up card material, not fixed here.

## Residual-risk

- The `completed` status and §E.4 signals exist only on the unpushed branch `WT-era-plan-phase` — until the lead merges, drift detection elsewhere reads the SPEC as `in-progress`, and this worktree holds the only copy.
- Merged-tree re-measurement is outstanding: §E.2 explicitly records that all figures predate absorption of `origin/develop`; a semantic conflict could still surface at merge.
- Era-classification behavior of the population (V3R5 24→1 etc.) was measured before absorption; the merge could shift the corpus.
