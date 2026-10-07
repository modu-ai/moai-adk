# Reference excerpt — t1500 seal record (SPEC-AUDIT-CEILING-REPAIR-001 AC-ACR-010)

Provenance:

- Source: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`
  (primary-checkout absolute path). The t1500 worktree was sweep-pending when
  this excerpt was pasted; plan-audit-1 finding D4 re-anchored AC-ACR-010's
  read-and-note obligation to this SPEC-local copy so the gate survives the
  source tree's disposal.
- Sections pasted verbatim: `## SEAL` (2026-10-06 23:4x KST) and `## PUSH`
  (2026-10-06 23:5x KST).
- Read by manager-spec on card t1560, 2026-10-07 (plan phase; re-read at the
  plan-audit-1 integration).
- Verification at read time: `git merge-base --is-ancestor 3d7215b72
  origin/main` → not an ancestor; `internal/runtime/audit_ceiling.go` present
  on `903ccd028` (the engine reached main by another landing).

--- verbatim excerpt below ---

## SEAL 2026-10-06 23:4x KST — operator reboot prep (leader drain order)

- **Seal point**: everything COMMITTED. Branch WT-audit-ceiling-counter HEAD `3d7215b72` (card-review repair round), tree clean, UNPUSHED — the worktree is the work's only copy, do NOT dispose.
- **Complete stages**: plan (5-round audit + leader conditional kickoff, decision record kickoff-decision.md) → run (M1-M4 + §E.2/§E.3, 22/22 AC, commits fb7e89c8e..bbbb30932) → sync (242ba7230 + 9184a52ff backfill, status: completed, §E.4) → card-review (codex advisory FAIL → 7×P2 RED-first repaired, 3 commits, card-review.md + §E.2 additive).
- **In flight at seal**: codex re-review (background task, read-only, does not survive session exit). Resume point ①: record its result into .moai/reports/t1500/card-review.md (dispatch result expected as task notification or re-run `codex_review scope=card`).
- **Resume points after re-review**: ② `factory stage t1500 merge-ready` path via `moai factory complete` (integration window acquire → merge into base → merge-ready; NOTE the stage field is pinned to plan — the stage verb refuses run/merge-ready transitions, bookkeeping limitation recorded, evidence paths carry truth) ③ leader completion report (merge SHA, unpushed count, evidence paths) ④ leader-side: develop/base push batch, CI verdict, t1453 release, t1531 lease, done marking.
- **Environment notes for next session**: MCP server env defect (claude backend) root-caused — see earlier UPDATE; [P1] infra card queued at leader. Cron dies with session — re-arm per lane rules at next session start.

## PUSH 2026-10-06 23:5x KST — leader-ordered backup push (reboot prep)

- `git push origin WT-audit-ceiling-counter` (non-force) exit 0; new remote branch; `origin/WT-audit-ceiling-counter` = `3d7215b72` = local HEAD (observed). Pre-push re-read: HEAD 3d7215b72, branch WT-audit-ceiling-counter, porcelain 0. The work is no longer single-copy. Seal resume points ②③④ unchanged.
