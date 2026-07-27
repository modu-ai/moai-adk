# SPEC-PRETOOL-GATE-MOVE-001 — progress.md

> Plan-phase emission. §E.1 populated by manager-spec; §E.2/§E.3/§E.4 are placeholder
> headings only (per the §E skeleton HARD obligation) and will be populated by their
> respective owners (manager-develop for §E.2/§E.3, manager-docs for §E.4).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: pending-plan-audit
- plan_complete_at: _(pending — awaiting plan-auditor verdict)_
- artifact_set: Tier M (spec.md + plan.md + acceptance.md + progress.md)
- plan_iter: 0 (initial draft, pre-audit)
- plan_artifact_hash: _(pending — computed by `internal/runtime/audit_cache.go` `ComputeHash` over the 4-file plan-artifact set: spec.md + plan.md + acceptance.md + tasks.md; for V3R6 Tier L the set would be spec.md + plan.md + acceptance.md + design.md + research.md, but this is Tier M so the 4-file set applies with the V3R4-era `tasks.md` name grandfathered in the hash subject list)_
- worktree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/p1b-gate-move`
- branch: `feat/SPEC-PRETOOL-GATE-MOVE-001`
- base: `origin/main @ 3e6c92ef7` (SPEC-FALSE-ALLCLEAR-GUARD-001 PR #1183 merge)
- census_anchor: `.moai/reports/census-2026-07-27-handoff.md` §C-2 (line 107), §7 P1-B (line 619)
- sibling_SPECs:
  - SPEC-GATE-001 (implemented) — original gate; REQ-GATE-011 no-bypass intent this SPEC operationalizes
  - SPEC-PRECOMMIT-001 (completed) — `PreCommitInstaller` this SPEC extends
  - SPEC-FALSE-ALLCLEAR-GUARD-001 (PR #1183, worktree base) — ast-grep scanner tuning this SPEC preserves
- chosen_direction: (e) relocate heavy gate OFF PreToolUse to native git pre-commit hook
- fallback_direction: (e-prime) standalone `moai gate` CLI if M1.a finds git pre-commit does NOT fire under Claude Code (REQ-PGM-012)
- tier: M
- req_count: 12 (REQ-PGM-001 .. REQ-PGM-012)
- ac_count: 15 (AC-PGM-001 .. AC-PGM-015, all MUST-PASS)

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs; carries sync_commit_sha>_
