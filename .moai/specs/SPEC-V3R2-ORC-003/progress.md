# SPEC-V3R2-ORC-003 Progress Tracker

> Phase tracker shell for **Effort-Level Calibration Matrix for 17 agents**.
> Updated by run-phase agents during M1..M5 execution.

## HISTORY

| Version | Date       | Author                          | Description                                                       |
|---------|------------|---------------------------------|-------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial progress.md shell. plan_complete_at populated post-PR-merge. |

---

## 1. Phase Status

| Phase   | Status       | Started | Completed | Notes |
|---------|--------------|---------|-----------|-------|
| Plan    | in-progress  | 2026-05-10 | (TBD)   | plan/SPEC-V3R2-ORC-003 branch open; PR pending squash-merge |
| Run     | pending      | -       | -         | Awaits plan PR merge + worktree setup |
| Sync    | pending      | -       | -         | Awaits run PR merge |
| Cleanup | pending      | -       | -         | Awaits sync PR merge |

Plan-phase status fields (populated by plan-auditor):
- `plan_complete_at`: (set when plan PR merged into main)
- `plan_status`: draft → audit-ready → audited → merged

---

## 2. Plan Phase Checkpoints

- [x] spec.md authored (already present at HEAD `3356aa9a9`)
- [x] research.md authored (Phase 1A — 30 evidence anchors)
- [x] plan.md authored (Phase 1B — 5 milestones, 27 tasks, 14 REQs traced)
- [x] acceptance.md authored (10 ACs with Given/When/Then)
- [x] tasks.md authored (27 tasks across M1..M5)
- [x] progress.md shell authored
- [x] issue-body.md authored (PR body draft)
- [ ] plan-auditor PASS verdict (target ≥ 0.85 first iteration)
- [ ] plan PR squash-merged into main
- [ ] plan_complete_at + plan_status = audit-ready persisted

---

## 3. Milestone Tracker (Run Phase)

### M1: Test scaffolding (RED) — Priority P0

- [ ] T-ORC003-01: 6 RED test fixtures added
- [ ] T-ORC003-02: LR-03 regression anchor added

Verification gate: `go test ./internal/cli/ -run "TestLintLR1[234]|TestAuthoringDocHasEffortMatrix|TestConstitutionCrossReference"` → all FAIL initially.

### M2: Lint rule + matrix table (GREEN seed) — Priority P0

- [ ] T-ORC003-03: LR-12, LR-13, LR-14 implemented
- [ ] T-ORC003-04: agent-authoring.md matrix table inserted
- [ ] T-ORC003-05: template parity + make build
- [ ] T-ORC003-06: constitution cross-reference applied
- [ ] T-ORC003-07: research.md cross-link to MIG-001 (already done in plan-phase)
- [ ] T-ORC003-08: lint help text updated

Verification gate: M1 RED tests now GREEN.

### M3: 17-agent frontmatter population (GREEN main) — Priority P0

- [ ] T-ORC003-09: manager-cycle.md → effort: high (or fallback ddd/tdd)
- [ ] T-ORC003-10: manager-quality.md → effort: high
- [ ] T-ORC003-11: manager-docs.md → effort: medium
- [ ] T-ORC003-12: manager-git.md → effort: medium
- [ ] T-ORC003-13: manager-project.md → effort: medium
- [ ] T-ORC003-14: expert-backend.md → effort: high
- [ ] T-ORC003-15: expert-frontend.md → effort: high
- [ ] T-ORC003-16: expert-devops.md → effort: medium
- [ ] T-ORC003-17: expert-performance.md → effort: high
- [ ] T-ORC003-18: builder-platform.md → effort: medium (or fallback agent/skill/plugin)
- [ ] T-ORC003-19: researcher.md → effort: xhigh
- [ ] T-ORC003-20: expert-security.md → effort: high → xhigh (DRIFT-3-A)
- [ ] T-ORC003-21: 3 drift corrections (evaluator-active, plan-auditor, expert-refactoring → xhigh)

Verification gate: `moai agent lint --path .claude/agents/moai/` reports 0 LR-03 + 0 LR-12 errors on the 17 v3r2 roster.

### M4: CI integration + JSON drift (REFACTOR + new) — Priority P0

- [ ] T-ORC003-22: TestConstitutionCrossReference GREEN
- [ ] T-ORC003-23: TestAuthoringDocHasEffortMatrix GREEN
- [ ] T-ORC003-24: TestLintLR03_MissingEffortIsError regression PASS
- [ ] T-ORC003-25: JSON drift field implemented + test passes
- [ ] T-ORC003-26: final template-local diff -r byte-identical

Verification gate: All 6 M1 RED fixtures GREEN.

### M5: Verification + audit — Priority P0

- [ ] T-ORC003-27: `go test -race -count=1 ./...` PASS
- [ ] T-ORC003-28: `golangci-lint run` clean
- [ ] T-ORC003-29: `make build` succeeds
- [ ] T-ORC003-30: CHANGELOG.md updated
- [ ] T-ORC003-31: @MX tags applied per plan §6
- [ ] T-ORC003-32: final manual lint verification (0 LR-03/12/13/14 errors on 17 v3r2 roster)

Verification gate: All AC-ORC-003-01..10 verified per acceptance.md.

---

## 4. AC Status Tracker

| AC ID | Status | Verified by | Verified at |
|---|---|---|---|
| AC-01 | pending | T-ORC003-23 | (TBD) |
| AC-02 | pending | T-ORC003-09..21 | (TBD) |
| AC-03 | pending | T-ORC003-32 | (TBD) |
| AC-04 | pending | T-ORC003-24 | (TBD) |
| AC-05 | pending | T-ORC003-25 (LR-12 fixture) | (TBD) |
| AC-06 | pending | T-ORC003-26 | (TBD) |
| AC-07 | pending | T-ORC003-22 | (TBD) |
| AC-08 | pending | T-ORC003-25 (LR-13 fixture) | (TBD) |
| AC-09 | SOFT-PASS | (advisory; documented cross-link) | 2026-05-10 (this plan-phase) |
| AC-10 | pending | T-ORC003-20 + T-ORC003-21 | (TBD) |

---

## 5. Iteration Tracker (Re-planning Gate)

Per `.claude/rules/moai/workflow/spec-workflow.md` § Re-planning Gate, append per-iteration acceptance criteria completion count + error count delta to detect stagnation (3+ iterations no AC progress → trigger AskUserQuestion gap analysis).

| Iteration | Date | AC completed | Error delta | Notes |
|---|---|---|---|---|
| (TBD)     | (TBD) | (TBD)       | (TBD)       | First run-phase iteration |

---

## 6. Risk Watch

Per spec §8 + plan §7 + research §7:

| Risk | Status | Mitigation |
|---|---|---|
| ORC-001 unmerged at run-time | UNKNOWN | tasks §4.1 fallback path |
| LR-03 idempotency confuses auditor | MITIGATED | plan §1.2.1 explicit acknowledgement |
| 4 (not 3) drift corrections | MITIGATED | plan §1.2.1 explicit acknowledgement |
| `high → xhigh` latency regression | MONITORED | post-merge 30-day telemetry; HRN-001 harness override available |
| Out-of-roster agents (manager-brain) | CARVED-OUT | LR-12 carve-out logic; not modified by this SPEC |

---

## 7. Notes

- Run-phase methodology: TDD per `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- Worktree base: `origin/main` HEAD `3356aa9a9` (plan-phase author baseline).
- Template-first discipline: all edits to `internal/template/templates/.claude/` first; `make build` regenerates; local tree byte-identical (diff -r gate).
- Estimated artifacts: 0 new files + 1 doc table + 1 doc cross-ref + 17 agent frontmatter edits × 2 trees + 3 new lint rules + 6 new test fixtures + CHANGELOG = ~430 LOC delta.

---

End of progress.

Version: 0.1.0
Status: Progress shell for SPEC-V3R2-ORC-003
