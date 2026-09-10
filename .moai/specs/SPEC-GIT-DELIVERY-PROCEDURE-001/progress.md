# Progress — SPEC-GIT-DELIVERY-PROCEDURE-001

> Plan-phase skeleton. §E.2–§E.4 are populated by manager-develop (run-phase) and manager-docs (sync-phase); this agent emits only §E.1.

## §E.1 Plan-phase Audit-Ready Signal

- **Plan-phase artifacts emitted**: spec.md, plan.md, acceptance.md, progress.md (this file).
- **Tier**: M. **Era**: V3R6. **Status**: `draft`.
- **Card**: t622. **Tree measured**: `b412f8a33b9f82ec5f85ccb5eeb960ef125dd8c0` (worktree `.claude/worktrees/t622`, branch `WT-git-procedure-fixes`).
- **Pre-write self-checks executed**:
  - SPEC ID regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` on `SPEC-GIT-DELIVERY-PROCEDURE-001` → `PASS`.
  - SPEC ID dedup: `ls -d .moai/specs/SPEC-GIT-DELIVERY-PROCEDURE-001` → `No such file or directory`, exit 1 (before creation).
  - Frontmatter: 12 canonical fields present, plus optional `era`, `tier`, `related_specs`. `phase` is a release target, not a stage token.
- **Open decisions (not decided here)**: OD-1 (Late-branch handling), OD-2 (auto-merge default source) — spec.md §C.
- **Deferred requirements**: REQ-GDP-006 (OD-2); REQ-GDP-009, 010, 011, 012 (OD-1).
- **Positive controls measured at plan time** (scope files unchanged from the tree above; raw outputs in the session scratchpad, not exported — run-phase re-measures into `.moai/reports/t622/run/`):
  - AC-GDP-001: `grep -n -E 'fetch.*independent'` on `manager-git.md` → line 156, exit 0; on `.codex/agents/moai/manager-git.toml` → line 150, exit 0. Section markers `## Synchronization` 154, `## PR Auto-Merge` 164.
  - AC-GDP-002: joined fetch/rev-list regex on `agent-common-protocol.md` → count 1 (line 347), exit 0; standalone fetch → line 296; markers 290 / 328 / 362; session-list line 303; matrix rows 9 (308-320).
  - AC-GDP-004: `grep -c 'pr merge --squash --delete-branch'` on `delivery.md` → 2.
  - AC-GDP-005: `manager-git.md` lines 32 (default sentence) and 114 (hardcoded example).
  - AC-GDP-007: `git grep -c -E` with the audit pattern at HEAD on the four template scope files → 8 / 3 / 1 / 2 = 14.
  - AC-GDP-008: title-case extraction → `Late-Branch Invocation Pattern` at `spec-assembly.md:340` and `spec-workflow.md:62`; target heading at `manager-git.md:88`. The first draft regex swallowed trailing words and was corrected in acceptance.md.
  - AC-GDP-010: `worktree add` detector → `main-checkout-branch-guard.md:38`. AC-GDP-012: `main_late_branch` → `manager-git.md:137`; `default skips GitHub Issue creation` in `SKILL.md` → 2. AC-GDP-015: SPEC-ID/date detector on spec.md → 4.
  - AC-GDP-013: `diff` of the two `delivery.md` copies → exit 1; header-stripped body holds only the 275 / 278 / footer differences.
  - SPEC lifecycle audit (`mcp__moai__spec_audit`, project_root = this worktree, filter this SPEC; server build v3.2.0-rc.5 commit 84fa4ece4) → total 1, modern_era_clean 1, drift_findings [].
- **plan_status**: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
