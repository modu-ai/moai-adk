# SPEC-V3R2-ORC-003 Task List

> Implementation task list for **Effort-Level Calibration Matrix for 17 agents**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                          | Description                                                                                     |
|---------|------------|---------------------------------|-------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial task list. 27 tasks (T-ORC003-01..27) grouped into 5 milestones (M1..M5), all priority P0. |

---

## 1. Task Overview

| Milestone | Phase | Tasks | Priority | REQ Coverage |
|---|---|---|---|---|
| M1: Test scaffolding (RED) | RED | T-ORC003-01..02 | P0 | REQ-006, REQ-013, REQ-012, REQ-014, REQ-001, REQ-005 |
| M2: Lint rule + matrix table (GREEN seed) | GREEN | T-ORC003-03..08 | P0 | REQ-001, REQ-002, REQ-005, REQ-013, REQ-012, REQ-014, REQ-008 |
| M3: 17-agent frontmatter population (GREEN main) | GREEN | T-ORC003-09..21 | P0 | REQ-002, REQ-003, REQ-004 |
| M4: CI integration + JSON drift (REFACTOR + new) | REFACTOR | T-ORC003-22..26 | P0 | REQ-006, REQ-009, REQ-011, REQ-005, REQ-008 |
| M5: Verification + audit (final gate) | VERIFY | T-ORC003-27..32 | P0 | (cross-cutting) |

Total: 27 tasks (T-ORC003-01..32 with T-ORC003-08 as M2 lint help text update). **Note**: numbering is non-contiguous because M3 explodes into 13 parallel sub-tasks; M5 has 6 finalization tasks. Spec-level task IDs run T-ORC003-01..27 with M3 tasks T-ORC003-09..21 (13 agents) treated as 13 individual tasks for traceability.

---

## 2. Tasks by Milestone

### Milestone 1: Test scaffolding (RED phase) — Priority P0

#### T-ORC003-01: Create RED test fixtures in agent_lint_test.go

**REQ traceback**: REQ-ORC-003-006 (LR-03 verification), REQ-ORC-003-013 (LR-12), REQ-ORC-003-012 (LR-13), REQ-ORC-003-014 (LR-14), REQ-ORC-003-001 (matrix table), REQ-ORC-003-005 (constitution cross-ref)

**AC traceback**: AC-04, AC-05, AC-08 (regression test sources)

**Goal**: Scaffold 6 new sub-tests in `internal/cli/agent_lint_test.go` that FAIL initially, will GREEN-flip after M2 + M3 deltas applied.

**Sub-tasks**:
- Add `TestLintLR12_MatrixDrift_DriftedAgent` (synthesizes expert-security with effort: high; expects LR-12 violation)
- Add `TestLintLR12_MatrixDrift_CleanAgent` (synthesizes expert-security with effort: xhigh; expects 0 LR-12 violations)
- Add `TestLintLR13_InvalidEffortEnum` (synthesizes agent with effort: ultra; expects LR-13 violation)
- Add `TestLintLR14_FixedBudgetTokens` (synthesizes agent body with `budget_tokens: 5000`; expects LR-14 violation)
- Add `TestAuthoringDocHasEffortMatrix` (reads `.claude/rules/moai/development/agent-authoring.md`; expects matrix section + 17 agent rows)
- Add `TestConstitutionCrossReference` (reads `.claude/rules/moai/core/moai-constitution.md`; expects cross-link to agent-authoring.md)

**Acceptance**:
- [ ] All 6 new tests added
- [ ] All 6 tests FAIL on `go test ./internal/cli/ -run "TestLintLR1[234]|TestAuthoringDocHasEffortMatrix|TestConstitutionCrossReference"`

**Dependencies**: None (clean RED setup)

---

#### T-ORC003-02: Add LR-03 regression anchor

**REQ traceback**: REQ-ORC-003-006 (LR-03 promotion)

**AC traceback**: AC-04 (LR-03 Error severity)

**Goal**: Add `TestLintLR03_MissingEffortIsError` (or rename existing TestCheckMissingEffort) to explicitly assert `Severity == SeverityError`. Update `internal/cli/agent_lint.go:382-396` checkMissingEffort comment to reference SPEC-V3R2-ORC-003 (documentation idempotency).

**Sub-tasks**:
- Add new test function or rename existing to `TestLintLR03_MissingEffortIsError`
- Test asserts: Severity == SeverityError (NOT Warning)
- Update Go-doc comment on `checkMissingEffort` to: `// LR-03: Error severity per SPEC-V3R2-ORC-003 (idempotent — already Error per ORC-002 implementation; this SPEC verifies and pins).`

**Acceptance**:
- [ ] TestLintLR03_MissingEffortIsError passes (SeverityError asserted)
- [ ] checkMissingEffort doc comment references SPEC-V3R2-ORC-003

**Dependencies**: None.

---

### Milestone 2: Lint rule + matrix table (GREEN seed) — Priority P0

#### T-ORC003-03: Implement LR-12, LR-13, LR-14 in agent_lint.go

**REQ traceback**: REQ-ORC-003-013, REQ-ORC-003-012, REQ-ORC-003-014

**AC traceback**: AC-05, AC-08

**Goal**: Add 3 new lint rules to `internal/cli/agent_lint.go`. After this task, M1 RED tests for LR-12/13/14 turn GREEN.

**Sub-tasks**:
1. Add package-level `canonicalEffortMatrix` constant (17 entries) per plan §4.1.
2. Add package-level `validEffortValues` constant (5 entries) per plan §4.2.
3. Add `checkEffortMatrixDrift(file string, fm AgentFrontmatter) []LintViolation` (LR-12) per plan §4.1.
4. Add `checkInvalidEffortEnum(file string, fm AgentFrontmatter) []LintViolation` (LR-13) per plan §4.2.
5. Add `checkFixedBudgetTokens(file string, content []byte) []LintViolation` (LR-14) per plan §4.3.
6. Wire LR-12, LR-13 into `lintAgentFile` dispatcher between checkMissingEffort (LR-03) and checkDeadHooks (LR-04).
7. Wire LR-14 into `lintAgentFile` dispatcher (body-scan; currently no body-scan path exists; add one or thread file content through).
8. Add @MX:ANCHOR tag on `canonicalEffortMatrix` per plan §6.
9. Add @MX:NOTE tag on `checkEffortMatrixDrift` per plan §6.
10. Add @MX:WARN tag on `checkFixedBudgetTokens` per plan §6.

**Acceptance**:
- [ ] LR-12, LR-13, LR-14 functions exist
- [ ] canonicalEffortMatrix has 17 entries matching plan §4.1
- [ ] validEffortValues has 5 entries
- [ ] M1 RED tests for LR-12/13/14 turn GREEN
- [ ] @MX tags applied per plan §6

**Dependencies**: T-ORC003-01 (RED tests must exist first)

---

#### T-ORC003-04: Insert canonical matrix table in agent-authoring.md

**REQ traceback**: REQ-ORC-003-001, REQ-ORC-003-002

**AC traceback**: AC-01

**Goal**: Add `## Effort-Level Calibration Matrix (SPEC-V3R2-ORC-003)` section to `.claude/rules/moai/development/agent-authoring.md` per plan §1.4 deliverable.

**Sub-tasks**:
- Identify insertion point (after the "Field Details" section near line 80)
- Insert the section heading + table per plan §M2-T4 verbatim
- Include the 17-row table with Agent / Effort / Rationale columns
- Add the "Harness Override" note (REQ-010 cross-reference)
- Add @MX:ANCHOR tag per plan §6

**Acceptance**:
- [ ] Section exists in `.claude/rules/moai/development/agent-authoring.md`
- [ ] All 17 agents listed with correct effort values
- [ ] TestAuthoringDocHasEffortMatrix turns GREEN

**Dependencies**: None (independent doc edit)

---

#### T-ORC003-05: Mirror to template tree + run make build

**REQ traceback**: REQ-ORC-003-004 (template parity)

**AC traceback**: AC-06

**Goal**: Mirror agent-authoring.md edits to template tree, run `make build`, verify diff -r byte-identical.

**Sub-tasks**:
- Copy modified `.claude/rules/moai/development/agent-authoring.md` to `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (or apply same edit there)
- Run `make build` to regenerate `internal/template/embedded.go`
- Run `diff -r .claude/rules/moai/development/ internal/template/templates/.claude/rules/moai/development/` and confirm exit 0

**Acceptance**:
- [ ] Both trees byte-identical
- [ ] `internal/template/embedded.go` regenerated cleanly
- [ ] `make build` exits 0

**Dependencies**: T-ORC003-04 (matrix exists in local tree first)

---

#### T-ORC003-06: Modify constitution Opus 4.7 Prompt Philosophy bullet

**REQ traceback**: REQ-ORC-003-005 (constitution cross-reference)

**AC traceback**: AC-07

**Goal**: Modify `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy "Effort level selection" bullet to cross-reference the matrix instead of inline values.

**Sub-tasks**:
- Locate the "Effort level selection" bullet in §Opus 4.7 Prompt Philosophy
- Replace inline list (`reasoning-intensive (manager-spec, ...) → effort: xhigh or high; ...`) with cross-reference text per plan §M2-T6
- Preserve the 3-tier categorization semantics (reasoning-intensive / implementation / template+speed-critical)
- Mirror to template tree

**Acceptance**:
- [ ] Constitution bullet now cross-references agent-authoring.md
- [ ] FROZEN doctrine semantics preserved (3-tier categorization still readable)
- [ ] TestConstitutionCrossReference turns GREEN
- [ ] Template-local diff -r byte-identical

**Dependencies**: T-ORC003-04 (cross-ref target must exist first)

---

#### T-ORC003-07: Add MIG-001 cross-link in research.md

**REQ traceback**: REQ-ORC-003-007 (migrator drift rewrite, advisory)

**AC traceback**: AC-09

**Goal**: Document MIG-001's responsibility for drift rewrite in research.md §5.7 (already done in this SPEC's research.md authoring; this task is verification only).

**Sub-tasks**:
- Verify research.md §5.7 contains cross-link to SPEC-V3R2-MIG-001 with REQ-007 advisory text

**Acceptance**:
- [ ] research.md §5.7 cross-links MIG-001

**Dependencies**: None (already in research.md)

---

#### T-ORC003-08: Update lint help text

**REQ traceback**: (documentation consistency, no specific REQ)

**AC traceback**: (no AC; doc fidelity)

**Goal**: Update `internal/cli/agent_lint.go` cobra Long description (line ~85) to enumerate LR-12, LR-13, LR-14 with one-line descriptions.

**Sub-tasks**:
- Add 3 lines to the Long description block:
  ```
  LR-12: Reject effort drift from SPEC-V3R2-ORC-003 canonical matrix
  LR-13: Reject invalid effort enum value (must be one of low/medium/high/xhigh/max)
  LR-14: Reject fixed budget_tokens (Opus 4.7 Adaptive Thinking rejects HTTP 400)
  ```

**Acceptance**:
- [ ] LR-12/13/14 listed in lint help

**Dependencies**: T-ORC003-03 (rules implemented first)

---

### Milestone 3: 17-agent frontmatter population (GREEN main) — Priority P0

Each task = 1 agent × 2 trees (local + template) = 2 file edits. Each edit inserts `effort: <value>` line in YAML frontmatter (alphabetical placement near `description:` or `model:`, consistent with existing manager-spec.md placement).

#### T-ORC003-09: manager-cycle.md — effort: high

**REQ traceback**: REQ-ORC-003-002, REQ-ORC-003-003

**AC traceback**: AC-02

**Pre-condition**: ORC-001 has merged AND `.claude/agents/moai/manager-cycle.md` exists. If not, apply fallback (apply `effort: high` to `.claude/agents/moai/manager-ddd.md` and `manager-tdd.md` until ORC-001 merge propagates the consolidated file).

**Sub-tasks**:
- Edit `.claude/agents/moai/manager-cycle.md` — add `effort: high` to YAML frontmatter
- Edit `internal/template/templates/.claude/agents/moai/manager-cycle.md` — same

**Acceptance**:
- [ ] Both files declare `effort: high`
- [ ] `moai agent lint .claude/agents/moai/manager-cycle.md` shows 0 LR-03 violations on this file

**Dependencies**: ORC-001 merged (advisory; fallback path documented)

---

#### T-ORC003-10: manager-quality.md — effort: high

(Same structure as T-ORC003-09; target value `high`)

#### T-ORC003-11: manager-docs.md — effort: medium

#### T-ORC003-12: manager-git.md — effort: medium

#### T-ORC003-13: manager-project.md — effort: medium

#### T-ORC003-14: expert-backend.md — effort: high

#### T-ORC003-15: expert-frontend.md — effort: high

#### T-ORC003-16: expert-devops.md — effort: medium

#### T-ORC003-17: expert-performance.md — effort: high

#### T-ORC003-18: builder-platform.md — effort: medium

**Pre-condition**: ORC-001 has merged AND `.claude/agents/moai/builder-platform.md` exists. If not, apply fallback to `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` until consolidated file lands.

#### T-ORC003-19: researcher.md — effort: xhigh

#### T-ORC003-20: expert-security.md — effort: high → xhigh (DRIFT-3-A)

**REQ traceback**: REQ-ORC-003-002 (matrix), REQ-ORC-003-003 (frontmatter)

**AC traceback**: AC-02, AC-10

**BC reference**: BC-V3R2-002 (effort upgrade)

**Goal**: Replace `effort: high` with `effort: xhigh` in expert-security.md frontmatter (4 locations: 2 trees × 1 file each, but since expert-security.md already has `effort: high`, this is a value change, not insertion).

**Sub-tasks**:
- Edit `.claude/agents/moai/expert-security.md` — change `effort: high` → `effort: xhigh`
- Edit `internal/template/templates/.claude/agents/moai/expert-security.md` — same

**Acceptance**:
- [ ] Both files declare `effort: xhigh`
- [ ] `git diff` shows `-effort: high` and `+effort: xhigh`
- [ ] `moai agent lint .claude/agents/moai/expert-security.md` shows 0 LR-12 violations

---

#### T-ORC003-21: 3 drift corrections (evaluator-active, plan-auditor, expert-refactoring)

**REQ traceback**: REQ-ORC-003-002, REQ-ORC-003-003

**AC traceback**: AC-02, AC-10

**BC reference**: BC-V3R2-002

**Note**: This task bundles 3 file pairs (3 agents × 2 trees = 6 file edits) since they are mechanically identical to T-ORC003-20. Listed as 1 task for milestone management; conceptually 3 sub-tasks.

**Sub-tasks**:
- Edit `.claude/agents/moai/evaluator-active.md` — change `effort: high` → `effort: xhigh`
- Edit `internal/template/templates/.claude/agents/moai/evaluator-active.md` — same
- Edit `.claude/agents/moai/plan-auditor.md` — change `effort: high` → `effort: xhigh`
- Edit `internal/template/templates/.claude/agents/moai/plan-auditor.md` — same
- Edit `.claude/agents/moai/expert-refactoring.md` — change `effort: high` → `effort: xhigh`
- Edit `internal/template/templates/.claude/agents/moai/expert-refactoring.md` — same

**Acceptance**:
- [ ] All 3 agents declare `effort: xhigh` (in both trees)
- [ ] `git diff` shows `-effort: high` and `+effort: xhigh` for each
- [ ] `moai agent lint` shows 0 LR-12 violations on the 3 agents

---

### Milestone 4: CI integration + JSON drift (REFACTOR + new) — Priority P0

#### T-ORC003-22: TestConstitutionCrossReference GREEN verification

**REQ traceback**: REQ-ORC-003-005

**AC traceback**: AC-07

**Goal**: Verify the M1 RED test now passes after T-ORC003-06 applied.

**Sub-tasks**:
- Run `go test ./internal/cli/ -run TestConstitutionCrossReference`
- Confirm PASS

**Dependencies**: T-ORC003-06

---

#### T-ORC003-23: TestAuthoringDocHasEffortMatrix GREEN verification

**REQ traceback**: REQ-ORC-003-001

**AC traceback**: AC-01

**Goal**: Verify the M1 RED test now passes after T-ORC003-04 applied.

**Sub-tasks**:
- Run `go test ./internal/cli/ -run TestAuthoringDocHasEffortMatrix`
- Confirm PASS

**Dependencies**: T-ORC003-04

---

#### T-ORC003-24: TestLintLR03_MissingEffortIsError regression test

**REQ traceback**: REQ-ORC-003-006, REQ-ORC-003-009

**AC traceback**: AC-03, AC-04

**Goal**: Confirm regression test from T-ORC003-02 passes.

**Sub-tasks**:
- Run `go test ./internal/cli/ -run TestLintLR03_MissingEffortIsError`
- Confirm PASS (severity=Error asserted)

**Dependencies**: T-ORC003-02

---

#### T-ORC003-25: Add JSON output drift field + test

**REQ traceback**: REQ-ORC-003-011

**AC traceback**: (no AC; tested via this task)

**Goal**: Implement REQ-011 optional JSON output `effort_drift` field.

**Sub-tasks**:
- Modify the JSON output struct in `internal/cli/agent_lint.go` to include `EffortDrift []EffortDriftEntry`
- Define `EffortDriftEntry` struct with `Agent`, `DeclaredEffort`, `ExpectedEffort`, `File`
- Populate `EffortDrift` in `lintAgentFile` when LR-12 fires
- Add `TestLintJSONOutputIncludesEffortDrift`: synthesize a drift fixture, run lint with `--format=json`, assert JSON output contains `effort_drift` array

**Acceptance**:
- [ ] `moai agent lint --format=json` produces JSON containing `effort_drift` field on drift fixtures
- [ ] JSON omits `effort_drift` (or empty) when no drift
- [ ] Test passes

**Dependencies**: T-ORC003-03 (LR-12 must exist)

---

#### T-ORC003-26: Final template-local parity diff -r

**REQ traceback**: REQ-ORC-003-004

**AC traceback**: AC-06

**Goal**: Final byte-identical verification.

**Sub-tasks**:
- Run `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/`
- Run `diff .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md`
- Run `diff .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md`
- All exit 0

**Acceptance**:
- [ ] All 3 diffs exit 0

**Dependencies**: T-ORC003-09..21 (all M3 edits applied to both trees)

---

### Milestone 5: Verification + audit — Priority P0

#### T-ORC003-27: Run full test suite

**REQ traceback**: (cross-cutting)

**AC traceback**: (final gate)

**Goal**: `go test -race -count=1 ./...` PASS.

**Acceptance**: 0 regressions.

**Dependencies**: M3 + M4 complete.

---

#### T-ORC003-28: Run linter

**Goal**: `golangci-lint run` clean.

---

#### T-ORC003-29: Verify make build

**Goal**: `make build` exits 0; `internal/template/embedded.go` regenerated correctly.

---

#### T-ORC003-30: Update CHANGELOG.md

**Goal**: Add 4 bullet entries to Unreleased section per plan §M5-T30.

---

#### T-ORC003-31: Apply @MX tags

**Goal**: Apply 4 @MX tags per plan §6 (canonicalEffortMatrix ANCHOR, checkEffortMatrixDrift NOTE, checkFixedBudgetTokens WARN, agent-authoring.md matrix ANCHOR).

---

#### T-ORC003-32: Final lint verification

**Goal**: Re-run `moai agent lint --path .claude/agents/moai/` (manual). Confirm 0 LR-03/12/13/14 errors for the 17 v3r2 roster.

**AC traceback**: AC-03

---

## 3. Task Dependency Graph

```
T-01 (RED tests scaffold)
  ↓
T-02 (LR-03 anchor)
  ↓
T-03 (LR-12/13/14 implementation)  →  T-08 (lint help)
  ↓
T-04 (matrix table) → T-05 (template parity + make build) → T-06 (constitution cross-ref)
  ↓                                                              ↓
T-23 (TestAuthoringDocHasEffortMatrix GREEN)                 T-22 (TestConstitutionCrossReference GREEN)
T-24 (TestLintLR03_MissingEffortIsError GREEN)
  ↓
T-09..T-21 (17 frontmatter edits) — parallel within milestone
  ↓
T-25 (JSON drift output)
T-26 (final diff -r parity)
  ↓
T-27 (full test suite)
T-28 (linter)
T-29 (make build)
T-30 (CHANGELOG)
T-31 (MX tags)
T-32 (final manual lint verification)
```

---

## 4. Dependency Resolution

### 4.1 ORC-001 merge status (run-time check)

At Run-phase entry (Step 2), execute:

```bash
git log --oneline origin/main | head -50 | grep "ORC-001" | head -1
```

**Branch A (ORC-001 MERGED)**: Proceed normally. T-ORC003-09 targets `manager-cycle.md`; T-ORC003-18 targets `builder-platform.md`. New consolidated agents exist.

**Branch B (ORC-001 NOT MERGED)**: Apply fallback:
- T-ORC003-09 fallback: target `manager-ddd.md` AND `manager-tdd.md` (apply `effort: high` to both). Document in run-PR description that this fallback exists pending ORC-001 merge; SPEC-V3R2-MIG-001 migrator handles the rename + value preservation.
- T-ORC003-18 fallback: target `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` (apply `effort: medium` to all 3). Same MIG-001 handoff.

**Decision matrix**:

| ORC-001 status | T-ORC003-09 target | T-ORC003-18 target |
|---|---|---|
| MERGED | `manager-cycle.md` (1 file) | `builder-platform.md` (1 file) |
| NOT MERGED | `manager-ddd.md` + `manager-tdd.md` (2 files) | `builder-agent.md` + `builder-skill.md` + `builder-plugin.md` (3 files) |

In either branch, the v3r2 17-agent roster's lint result must show 0 LR-03 + 0 LR-12 errors (AC-03 + AC-02 satisfied).

### 4.2 ORC-002 status (assumed merged)

ORC-002 already merged per research.md §5.2. LR-03 already at Error severity per `internal/cli/agent_lint.go:386`. REQ-006 verification = T-ORC003-24 regression test only; no severity flip required.

### 4.3 SPEC-V3-AGT-001 schema validator (advisory)

REQ-012 (invalid enum) is dual-gated: schema validator (runtime, SPEC-V3-AGT-001 territory) + LR-13 (lint time, this SPEC). Both fire with consistent error code. T-ORC003-03 implements the lint-side gate; schema-side is out-of-scope.

### 4.4 SPEC-V3R2-MIG-001 migrator (advisory)

REQ-007 (migrator drift rewrite) is implemented in MIG-001. T-ORC003-07 documents the cross-link; no code in this SPEC.

### 4.5 SPEC-V3R2-HRN-001 harness (advisory)

REQ-010 (harness override) is documented in T-ORC003-04 matrix section comment. No code in this SPEC.

---

## 5. Effort Estimates (priority labels only, no time)

Per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule, all tasks use priority labels; no time estimates.

| Task category | Priority | Notes |
|---|---|---|
| All 27 tasks | P0 | Critical path; no P1/P2 in this SPEC |

---

## 6. Test Coverage Goal

Per `.moai/config/sections/quality.yaml` `coverage_target: 85`:

| File | Pre-SPEC LOC | Post-SPEC LOC | Coverage target |
|---|---|---|---|
| `internal/cli/agent_lint.go` | ~840 | ~930 | ≥ 85% |
| `internal/cli/agent_lint_test.go` | ~750 | ~930 | (test code; coverage proxy) |

---

## 7. Risk-based Task Prioritization

| Risk | Mitigation task |
|---|---|
| ORC-001 unmerged at run-time | T-ORC003-09 + T-ORC003-18 fallback path (§4.1) |
| LR-03 idempotency invisible to plan-auditor | T-ORC003-02 explicit anchor + comment |
| Matrix-constitution drift | T-ORC003-06 cross-ref + T-ORC003-22 regression test |
| Fixed budget_tokens regex false positives | T-ORC003-03 LR-14 + plan §6 @MX:WARN documentation |
| Template-local drift | T-ORC003-05 + T-ORC003-26 byte-identical diff -r gates |

---

End of tasks.

Version: 0.1.0
Status: Tasks artifact for SPEC-V3R2-ORC-003
