# SPEC-V3R2-ORC-004 Progress Tracker

> Phase tracker for **Worktree MUST Rule for write-heavy role profiles**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Updated after each milestone completion. Used by `re-planning gate` (spec-workflow.md) for stagnation detection.

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec  | Initial progress tracker shell. plan_complete_at set; awaiting plan-auditor verdict. |

---

## Plan Phase Status

| Field | Value |
|-------|-------|
| `plan_complete_at` | 2026-05-10 (Step 1, plan-in-main) |
| `plan_status` | audit-ready |
| `plan_branch` | `plan/SPEC-V3R2-ORC-004` |
| `plan_base` | `origin/main` HEAD `dca57b14d` |
| `plan_pr` | (TBD — fill after `gh pr create`) |
| `plan_auditor_iteration` | 0 (not yet run) |
| `plan_auditor_verdict` | (pending) |
| `plan_auditor_score` | (pending) |

### Plan File Inventory

| File | Status | Bytes | Lines |
|------|--------|-------|-------|
| spec.md | EXISTING (untouched) | 21200 | 277 |
| research.md | NEW | (post-write) | (post-write) |
| plan.md | NEW | (post-write) | (post-write) |
| tasks.md | NEW | (post-write) | (post-write) |
| acceptance.md | NEW | (post-write) | (post-write) |
| progress.md | NEW (this file) | (post-write) | (post-write) |
| issue-body.md | NEW | (post-write) | (post-write) |

---

## Run Phase Status — COMPLETE (2026-05-18)

| Milestone | Status | Completed | Tasks | AC Met |
|-----------|--------|-----------|-------|--------|
| M1 — Frontmatter Additions (4 agents) | DONE | 2026-05-18 | T-04..T-08 | AC-01, AC-02 |
| M2 — manager-cycle (conditional) | SKIPPED (ORC_DEPENDENCY_MISSING) | — | T-09 | (deferred to ORC-001 merge) |
| M3 — Rule Text Update | DONE | 2026-05-18 | T-04, T-12 | AC-01 |
| M4 — Researcher Body | DONE | 2026-05-18 | T-08 | AC-08 |
| M5 — Lint Sentinel Messages | DONE | 2026-05-18 | T-10..T-13 | AC-06, AC-07 |
| M6 — `moai workflow lint` | DONE | 2026-05-18 | T-14..T-17 | AC-09 |
| M7 — Template Mirror & Build | DONE | 2026-05-18 | T-18..T-20 | AC-02 |
| M8 — Verification & Documentation | DONE | 2026-05-18 | T-21..T-31 | AC-05, AC-10 |

### Conditional Skip — T-ORC004-09
**ORC_DEPENDENCY_MISSING**: manager-cycle.md does not exist (SPEC-V3R2-ORC-001 not yet merged).
AC-02 is satisfied 4/5 — manager-cycle deferred pending ORC-001 merge.

### Iteration Log (TDD cycle tracking — populated by manager-cycle)

| Iter | Cycle Phase | Acceptance Criteria Met (count) | New Errors (delta) | Notes |
|------|-------------|---------------------------------|--------------------|-------|
| 1 | RED | 0 | (n/a) | Add T-10, T-11, T-14 RED tests |
| 2 | GREEN | (TBD) | (TBD) | (run-phase entry) |
| 3 | REFACTOR | (TBD) | (TBD) | (run-phase entry) |

---

## Sync Phase Status (placeholder — not started)

| Field | Value |
|-------|-------|
| `sync_branch` | (TBD: `sync/SPEC-V3R2-ORC-004` after run merges) |
| `sync_pr` | (TBD) |
| `codemap_regenerated` | (TBD) |
| `mx_tags_validated` | (TBD) |
| `changelog_entry_added` | (TBD) |

---

## Re-Planning Gate Counters

Per `spec-workflow.md` § Re-planning Gate, manager-cycle appends acceptance criteria completion count and error count delta to this file at the end of each iteration. Stagnation flagged when AC completion rate is zero for 3+ consecutive entries.

### Acceptance Criteria Completion Tracker

| AC | Status | Verified By |
|----|--------|-------------|
| AC-V3R2-ORC-004-01 | PASS | worktree-integration.md MUST clause + Sentinel Key Glossary added |
| AC-V3R2-ORC-004-02 | PARTIAL (4/5) | 4 agents verified; manager-cycle: ORC_DEPENDENCY_MISSING |
| AC-V3R2-ORC-004-03 | PASS | 4 read-only agents have no isolation:worktree |
| AC-V3R2-ORC-004-04 | PASS | yq verified: implementer/tester/designer = worktree |
| AC-V3R2-ORC-004-05 | PASS | `moai agent lint` — no LR-05 for ORC-004 target agents |
| AC-V3R2-ORC-004-06 | PASS | TestLintLR05_OrcWorktreeMissingSentinel GREEN |
| AC-V3R2-ORC-004-07 | PASS | TestLintLR09_OrcWorktreeOnReadonlySentinel GREEN |
| AC-V3R2-ORC-004-08 | PASS | researcher.md body: "mandatory per SPEC-V3R2-ORC-004" |
| AC-V3R2-ORC-004-09 | PASS | `moai workflow lint` exit 0 + 4 unit tests GREEN |
| AC-V3R2-ORC-004-10 | PASS | team/run.md cross-reference present |

---

## Stagnation Detection Reference

If, between any 3 consecutive iteration log entries, no new AC transitions from PENDING to PASS, manager-cycle MUST emit a stagnation report and the orchestrator runs the re-planning gate AskUserQuestion (continue / revise SPEC / alternative approach / pause).

Currently zero iterations completed; stagnation detection inactive.

---

## Blockers

(none currently — populated when discovered during run phase)

| Date | Blocker | Resolution Path |
|------|---------|-----------------|
| — | — | — |

---

## Notes

- ORC-001 dependency: T-09 may pause if manager-cycle file does not exist at run-phase start. Document the wait in §Blockers above when triggered.
- `moai workflow lint` is a new CLI; ensure cobra subcommand registration uses correct GroupID (`tools` to mirror `agent lint`).
- Sentinel constant placement: REFACTOR step (T-16) chooses between in-file `const` block vs new `internal/cli/sentinels.go`. Decision deferred to GREEN phase.

---

End of progress.
