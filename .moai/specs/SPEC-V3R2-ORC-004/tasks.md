# SPEC-V3R2-ORC-004 Implementation Tasks

> Task breakdown for **Worktree MUST Rule for write-heavy role profiles**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0.
> Each task references the milestone it belongs to (M1..M8 from `plan.md` §2) and the REQ/AC it serves.

## HISTORY

| Version | Date       | Author        | Description |
|---------|------------|---------------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec  | Initial task breakdown with REQ-ORC-004-XXX traceability and milestone grouping. |

---

## Legend

- **T-ORC004-NN**: Task ID (NN is 2-digit sequential).
- **REQ**: SPEC requirement IDs satisfied (from `spec.md` §5).
- **AC**: Acceptance criteria served (from `acceptance.md`).
- **M**: Milestone bucket (from `plan.md` §2).
- **Type**: `code`, `doc`, `test`, `verify`, `chore`.
- **Priority**: P0 (blocker), P1 (high), P2 (medium).

All tasks executed in run-phase worktree (`feat/SPEC-V3R2-ORC-004`). Template-first discipline mandatory: edits flow into `internal/template/templates/` first, then mirror via `make build && make install && moai update`.

---

## Phase 0: Preconditions (run-phase entry)

| Task | Action | Type | Priority |
|------|--------|------|----------|
| T-ORC004-00 | Verify `git rev-parse --show-toplevel` ends with `SPEC-V3R2-ORC-004` (worktree cwd anchoring per session-handoff.md Block 0) | verify | P0 |
| T-ORC004-01 | Verify `git rev-parse origin/main` is current (no rebase needed mid-run) | verify | P0 |
| T-ORC004-02 | Verify SPEC-V3R2-ORC-001 status (manager-cycle existence). If unmerged, set conditional flag for T-ORC004-09 | verify | P0 |
| T-ORC004-03 | Read all 5 plan files into context: spec.md, plan.md, research.md, acceptance.md, tasks.md | verify | P0 |

---

## Phase 1: M1 — Frontmatter Additions (4 existing agents)

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-04 | Update `worktree-integration.md` line 135 SHOULD → MUST + add 5-agent enumeration. Also add §"Sentinel Key Glossary" section near §HARD Rules listing `ORC_WORKTREE_REQUIRED`, `ORC_WORKTREE_MISSING`, `ORC_WORKTREE_ON_READONLY` | `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` | REQ-001, REQ-006, REQ-013, REQ-014 | AC-01 | M3 | doc | P0 |
| T-ORC004-05 | Add `isolation: worktree` to expert-backend frontmatter (after `permissionMode:` line) | `internal/template/templates/.claude/agents/moai/expert-backend.md` | REQ-002 | AC-02 | M1 | doc | P0 |
| T-ORC004-06 | Add `isolation: worktree` to expert-frontend frontmatter | `internal/template/templates/.claude/agents/moai/expert-frontend.md` | REQ-002 | AC-02 | M1 | doc | P0 |
| T-ORC004-07 | Add `isolation: worktree` to expert-refactoring frontmatter | `internal/template/templates/.claude/agents/moai/expert-refactoring.md` | REQ-002 | AC-02 | M1 | doc | P0 |
| T-ORC004-08 | Add `isolation: worktree` to researcher frontmatter + locate body line "All experiments in worktree isolation when possible" and replace with "All experiments in worktree isolation (mandatory per SPEC-V3R2-ORC-004)". Also add `@MX:NOTE` annotation explaining P-A22 reconciliation. | `internal/template/templates/.claude/agents/moai/researcher.md` | REQ-002, REQ-011 | AC-02, AC-08 | M1, M4 | doc | P0 |

---

## Phase 2: M2 — manager-cycle Frontmatter (conditional)

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-09 | **CONDITIONAL** on ORC-001 merge: Add `isolation: worktree` to manager-cycle frontmatter. If file does not exist, emit blocker report `ORC_DEPENDENCY_MISSING` to orchestrator with cause "SPEC-V3R2-ORC-001 not yet merged" | `internal/template/templates/.claude/agents/moai/manager-cycle.md` | REQ-002, REQ-012 | AC-02 | M2 | doc | P0 |

---

## Phase 3: M5 — Lint Sentinel Messages (RED → GREEN)

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-10 | (RED) Add `TestLintLR05_OrcWorktreeMissingSentinel` to `agent_lint_test.go` per plan.md §4.1.1. Verify FAIL initially | `internal/cli/agent_lint_test.go` | REQ-013 | AC-06 | M5 | test | P1 |
| T-ORC004-11 | (RED) Add `TestLintLR09_OrcWorktreeOnReadonlySentinel` to `agent_lint_test.go` per plan.md §4.1.1. Verify FAIL initially | `internal/cli/agent_lint_test.go` | REQ-014 | AC-07 | M5 | test | P1 |
| T-ORC004-12 | (GREEN) Add `SentinelWorktreeMissing` and `SentinelWorktreeOnReadonly` const block to `agent_lint.go`. Update LR-05 violation message to append ` (SPEC-V3R2-ORC-004 ORC_WORKTREE_MISSING)`. Update LR-09 violation message to append ` (SPEC-V3R2-ORC-004 ORC_WORKTREE_ON_READONLY)`. Re-run T-ORC004-10/11 — verify GREEN | `internal/cli/agent_lint.go` | REQ-006, REQ-013, REQ-014 | AC-06, AC-07 | M5 | code | P1 |
| T-ORC004-13 | Update `agent_lint.go` help text to mention sentinel keys in LR-05 and LR-09 line descriptions | `internal/cli/agent_lint.go` | REQ-013, REQ-014 | (info) | M5 | doc | P2 |

---

## Phase 4: M6 — `moai workflow lint` Implementation (RED → GREEN → REFACTOR)

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-14 | (RED) Create `internal/cli/workflow_lint_test.go` with 4 test cases per plan.md §4.1.2: implementer/tester/designer triggers + clean-config no-violation. Verify all FAIL initially | `internal/cli/workflow_lint_test.go` | REQ-008 | AC-09 | M6 | test | P1 |
| T-ORC004-15 | (GREEN) Implement `internal/cli/workflow_lint.go` per plan.md §1.5.1: `WorkflowLintViolation` struct, `runWorkflowLint`, `validateRoleProfiles`, `loadWorkflowYAML`. Wire `workflow lint` cobra subcommand under new `workflow` group with `GroupID: "tools"`. Const `SentinelWorktreeRequired = "ORC_WORKTREE_REQUIRED"`. Re-run T-ORC004-14 — verify GREEN | `internal/cli/workflow_lint.go` | REQ-008 | AC-09 | M6 | code | P1 |
| T-ORC004-16 | (REFACTOR) Extract sentinel keys into shared location (e.g. `internal/cli/sentinels.go` with const block). Re-run all lint tests — verify GREEN preserved | `internal/cli/sentinels.go`, `internal/cli/agent_lint.go`, `internal/cli/workflow_lint.go` | (refactor) | (refactor) | M6 | code | P2 |
| T-ORC004-17 | Add `moai workflow lint` to `moai doctor` integration (optional integration test): doctor.go can invoke `moai workflow lint` and reports failures alongside other lint output. **Out-of-scope guard: skip if doctor integration adds complexity; document as future work in §9 of plan.md** | `internal/cli/doctor.go` (if exists) | REQ-008 | (info) | M6 | code | P2 |

---

## Phase 5: M7 — Template Mirror & Build

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-18 | Run `make build` in worktree root. Verify `internal/template/embedded.go` regenerates with new agent frontmatter content | (regenerated) | REQ-005 | (CI gate) | M7 | chore | P0 |
| T-ORC004-19 | Run `make install` to install fresh `moai` binary | (binary) | REQ-005 | (CI gate) | M7 | chore | P0 |
| T-ORC004-20 | Run `moai update` from worktree root to mirror templates to `.claude/`. Verify `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` returns zero diff | `.claude/agents/moai/*.md` | REQ-005 | AC-02 | M7 | chore | P0 |

---

## Phase 6: M8 — Verification & Documentation

| Task | Action | Files | REQ | AC | M | Type | Priority |
|------|--------|-------|-----|----|---|------|----------|
| T-ORC004-21 | Verify AC-02: `for f in expert-backend expert-frontend expert-refactoring researcher manager-cycle; do grep "isolation: worktree" .claude/agents/moai/$f.md \|\| echo "MISSING: $f"; done` returns 5 matches (or 4 + 1 conditional skip if ORC-001 unmerged) | (verify) | REQ-002 | AC-02 | M8 | verify | P0 |
| T-ORC004-22 | Verify AC-03: `for f in manager-strategy manager-quality evaluator-active plan-auditor; do grep "isolation: worktree" .claude/agents/moai/$f.md && echo "UNEXPECTED: $f"; done` returns 0 matches | (verify) | REQ-003 | AC-03 | M8 | verify | P0 |
| T-ORC004-23 | Verify AC-04: `yq '.workflow.team.role_profiles \| to_entries[] \| select(.value.isolation == "worktree") \| .key' .moai/config/sections/workflow.yaml` returns `implementer`, `tester`, `designer` (3 names) | (verify) | REQ-004 | AC-04 | M8 | verify | P0 |
| T-ORC004-24 | Verify AC-05: `moai agent lint` exit code 0 against post-edit roster | (verify) | REQ-009 | AC-05 | M8 | verify | P0 |
| T-ORC004-25 | Verify AC-09: `moai workflow lint` exit code 0 against current workflow.yaml | (verify) | REQ-008 | AC-09 | M8 | verify | P0 |
| T-ORC004-26 | Verify AC-10: Read `.claude/skills/moai/team/run.md` and confirm "no absolute paths in spawn prompts for worktree-isolated agents" guidance is present (cross-reference, no new text required) | (verify) | REQ-010, REQ-015 | AC-10 | M8 | verify | P1 |
| T-ORC004-27 | Add MX tags per plan.md §6: 5 annotations across worktree-integration.md, workflow_lint.go, agent_lint.go, researcher.md | 5 files | mx_plan | (info) | M8 | doc | P1 |
| T-ORC004-28 | Run `make test ./...` — all tests GREEN. Special focus: `internal/cli/...` package | (verify) | (test gate) | (CI gate) | M8 | verify | P0 |
| T-ORC004-29 | Run `go vet ./...` and `golangci-lint run` — zero issues | (verify) | (lint gate) | (CI gate) | M8 | verify | P0 |
| T-ORC004-30 | Add CHANGELOG.md Unreleased entry: "feat(workflow): SPEC-V3R2-ORC-004 — Promote isolation: worktree from SHOULD to MUST for write-heavy v3r2 agents (#TBD)" | `CHANGELOG.md` | (Trackable) | (TRUST 5) | M8 | doc | P1 |
| T-ORC004-31 | Final pre-submission self-review: read full diff against acceptance.md scenarios. Confirm "Is there a simpler approach?" answer is no. | (review) | (TRUST 5) | (Self-review) | M8 | verify | P1 |

---

## Task Sequencing

Execution order (logical dependencies):

```
T-ORC004-00 → T-ORC004-01 → T-ORC004-02 → T-ORC004-03  (Phase 0 preconditions)
                ↓
              T-ORC004-04 (rule text — independent)
                ↓
              T-ORC004-05, -06, -07, -08 (4 frontmatter additions, parallelizable)
                ↓
              T-ORC004-09 (manager-cycle — conditional)
                ↓
              T-ORC004-10, -11 (RED tests, parallelizable)
                ↓
              T-ORC004-12, -13 (GREEN sentinel msg)
                ↓
              T-ORC004-14 (RED workflow_lint test)
                ↓
              T-ORC004-15, -16, -17 (GREEN workflow_lint impl + REFACTOR + optional doctor)
                ↓
              T-ORC004-18, -19, -20 (build + install + moai update — sequential)
                ↓
              T-ORC004-21..-31 (verification + docs, mostly parallelizable)
```

Phase 1 (M1) and Phase 3 (M5 RED) can interleave because they touch different files. Phase 5 (M7) MUST run AFTER all Phase 1-4 edits because `make build` re-generates embedded.go.

---

## Total Task Count

- **Phase 0** (preconditions): 4 tasks
- **Phase 1** (M1 frontmatter): 5 tasks (T-04 through T-08)
- **Phase 2** (M2 manager-cycle): 1 task (T-09, conditional)
- **Phase 3** (M5 sentinel msg): 4 tasks (T-10 through T-13)
- **Phase 4** (M6 workflow lint): 4 tasks (T-14 through T-17)
- **Phase 5** (M7 build & mirror): 3 tasks (T-18 through T-20)
- **Phase 6** (M8 verify): 11 tasks (T-21 through T-31)

**Total: 32 tasks** (T-ORC004-00 through T-ORC004-31).

Of these, 6 are P0 verifications (Phase 0 + final gates), 13 are P0/P1 implementation, 8 are P1 documentation/verify, and 5 are P2 chore/optional. The plan-auditor's drift-guard threshold of 30% maps to ~10 file modifications; ORC-004 is well under that ceiling.

---

End of tasks.
