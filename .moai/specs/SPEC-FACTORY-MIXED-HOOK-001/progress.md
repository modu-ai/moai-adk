---
id: SPEC-FACTORY-MIXED-HOOK-001
document: progress
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Progress — SPEC-FACTORY-MIXED-HOOK-001

## §A Status

- Current status: `in-progress`.
- Card: `t1074` (`picked`).
- Worktree: `WT-factory-mixed-hook`.
- Plan baseline: `758314007` (local develop matched when authoring began).
- Implementation baseline: `8d8e101bf`; M1 implementation commit: `cb099897a`.
- Next gate: finish M2/M3/M4 without absorbing the t1082 worktree lifecycle, then independent sync audit.

## §B Plan artifacts

| Artifact | Status |
|---|---|
| spec.md | authored |
| plan.md | authored |
| acceptance.md | authored |
| research.md | authored |
| design.md | authored |
| progress.md | authored |

## §C Plan audit

### Iteration 1

- Verdict: FAIL, merge-blocking.
- Findings: non-canonical GEARS clauses, forbidden sibling status fields, abbreviated REQ references, missing executable RED ledger, live `NOT_RUN` false-pass, and missing exclusion structure.
- Resolution: all findings corrected; strict lint and mutant gates added.

### Iteration 2 — final

- Independent auditor verdict: `PASS`.
- Score: `0.94 / 1.00` (Tier L threshold `0.85`).
- Findings: empty; merge-blocking findings: none.
- Auditor-observed evidence:
  - `go run ./cmd/moai spec lint SPEC-FACTORY-MIXED-HOOK-001 --strict --json` → `[]`, exit 0 (with isolated `GOCACHE`).
  - `git diff --check` → stdout empty, exit 0.
  - Fifteen RED-now named-test presence probes → stdout empty, exit 1 at tree `758314007d8c696ff1af377dc8cdc46d76368314`.
  - Synthetic live events: positive passes, `skip` and `NOT_RUN` fail.

## §D Run-phase evidence

- M1 committed as `cb099897a` (`feat(t1074): M1 bind canonical factory runs and peers`).
- Observed scoped tests at that commit: `internal/cli` → `ok ... 4.491s`; `internal/factorymsg` → `ok ... 1.310s`.
- M2/M3 changes are in progress and uncommitted; their earlier scoped package passes are provisional and must be rerun after the design-boundary edit.
- Live criteria and benchmark remain `NOT_RUN`.

## §D.1 Codex cwd redesign decision

- Latest official source/docs and a local Codex `0.155.1` probe establish that interactive `/cd` can keep the visible conversation flow while rotating the physical session/thread UUID.
- `t1074` therefore owns the stable logical lane/current-endpoint broker seam only.
- `t1082` owns card worktree creation and `/cd` or headless `cwd` handoff through `BOUND`.
- `t1075` depends on t1082 and may wake only the current bound endpoint.

## §E Audit-ready signals

- Plan-phase: audit-ready; independent `PASS` at score `0.94`.
- Run-phase: M1 complete; M2/M3/M4 pending, followed by independent sync audit.
