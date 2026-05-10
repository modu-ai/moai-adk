# SPEC-V3R2-ORC-003 Implementation Plan

> Implementation plan for **Effort-Level Calibration Matrix for 17 agents**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored on branch `plan/SPEC-V3R2-ORC-003` (Step 1 plan-in-main; base `origin/main` HEAD `3356aa9a9`).
> Run phase will execute on a fresh worktree `feat/SPEC-V3R2-ORC-003` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Step 2.

## HISTORY

| Version | Date       | Author                          | Description                                                                                                                                                              |
|---------|------------|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial implementation plan per `.claude/skills/moai/workflows/plan.md`. Scope: 17-agent effort matrix population, agent-authoring.md table publication, constitution cross-reference, and LR-03/LR-12 lint hardening. |

---

## 1. Plan Overview

### 1.1 Goal restatement

`spec.md` §1 의 핵심 목표를 milestone 분해:

> Publish the canonical 17-agent effort-level matrix in `agent-authoring.md`, populate every v3r2 agent's `effort:` frontmatter to match, promote LR-03 from warning to error (idempotent — already enforced as Error severity in `internal/cli/agent_lint.go:386` since ORC-002 implementation), add canonical-matrix drift detection (LR-12 / `ORC_EFFORT_MATRIX_DRIFT`), enforce 5-value enum validation (LR-13 / `AGT_INVALID_FRONTMATTER`), and reject fixed `budget_tokens` (LR-14 / `ORC_FIXED_BUDGET_PROHIBITED`).

### 1.2 Current State Audit (research.md §2 cross-reference)

Wave 1 audit of `.claude/agents/moai/*.md` on base `origin/main` HEAD `3356aa9a9`:

| Agent (current state) | `effort:` value | `status:` | Roster destiny per ORC-001 |
|----|----|----|----|
| manager-spec | `xhigh` ✅ | draft | KEEP (xhigh-target) — already correct |
| manager-strategy | `xhigh` ✅ | (none) | KEEP (xhigh-target) — already correct |
| manager-brain | `xhigh` ⚠ | (none) | NOT in v3r2 17-agent roster (out-of-scope) |
| evaluator-active | `high` ❌ | (none) | KEEP (xhigh-target) — **DRIFT-3-A: high → xhigh** |
| expert-refactoring | `high` ❌ | (none) | KEEP (xhigh-target) — **DRIFT-3-B: high → xhigh** |
| expert-security | `high` ❌ | (none) | KEEP (xhigh-target) — **DRIFT-3-C: high → xhigh** |
| plan-auditor | `high` ❌ | (none) | KEEP (xhigh-target) — **DRIFT-3-D: high → xhigh** |
| (19 others) | (missing) | (n/a) | mix of KEEP / RETIRED-by-ORC-001 |

Net actionable matrix population (Run-phase scope):

- **3 explicit drift corrections** (BC-V3R2-002): expert-security `high → xhigh`, evaluator-active `high → xhigh`, plan-auditor `high → xhigh`.
  - Note: research.md §2 also flags `expert-refactoring` as drift-4 (current `high` → matrix `xhigh`); spec §1.1 enumerated only 3 by oversight (1.1 says "expert-security, evaluator-active, plan-auditor are declared `high` but the constitution names them for `xhigh`"); reality is 4 explicit drifts per the matrix in spec §1.2 and §5.2 REQ-002. **Plan binding: 4 drift corrections** (not 3); spec-level reconciliation in §1.2.1 below.
- **13 missing-`effort:` populations** on the v3r2 roster: manager-cycle (NEW from ORC-001), manager-quality, manager-docs, manager-git, manager-project, expert-backend, expert-frontend, expert-devops, expert-performance, builder-platform (NEW from ORC-001), researcher.
- **0 retire-trees touched** (ORC-001 deleted manager-ddd, manager-tdd, builder-agent, builder-skill, builder-plugin, expert-debug, expert-testing files; this SPEC does NOT modify retired agent stubs).
- **2 agents already correct** (manager-spec, manager-strategy at `xhigh`).

If ORC-001 has not yet merged at run-time (verify Step 2 base SHA), the run-phase tasks T-ORC003-09 (manager-cycle.md effort) and T-ORC003-15 (builder-platform.md effort) become **blocked**; tasks.md §4.1 Dependency Resolution defines fallback path (apply effort to retired manager-ddd/manager-tdd/builder-* until consolidated agents land).

### 1.2.1 Acknowledged Discrepancies

본 plan 이 spec.md 와 의도적으로 다르게 처리하는 부분 (research.md §3 evidence 기반):

- **Drift count: spec §1.1 says 3, plan binds 4.** Per spec §1.2 In-Scope item "Correct the 3 explicit drift cases: expert-security, evaluator-active, plan-auditor (high → xhigh)" but expert-refactoring is *also* targeted at xhigh per spec §1.1 R5 audit table row "Reasoning-intensive (xhigh): manager-spec, manager-strategy, expert-security, **expert-refactoring**, evaluator-active, plan-auditor, researcher". expert-refactoring currently declares `effort: high` in its frontmatter — this IS an explicit drift requiring correction. Plan-binding count: **4 explicit drifts** (expert-security, evaluator-active, plan-auditor, expert-refactoring all moving from `high` to `xhigh`). Sync-phase HISTORY entry will reconcile spec.md §1.1's "3" → "4" and §1.2 "Correct the 3 explicit drift cases" → "Correct the 4 explicit drift cases".
- **LR-03 already at Error severity.** Per `internal/cli/agent_lint.go:386` (`severity := SeverityError`) and the comment "promoted from warning per SPEC-V3R2-ORC-003". The promotion REQ-006 is operationally **idempotent**; CI gate is already strict. Plan binding: REQ-006 verification only (no severity flip); add anchor comment + test fixture covering "LR-03 emits Error on missing effort" (defense-in-depth against future regression).
- **LR-12 / LR-13 / LR-14 are NEW lint rules introduced by this SPEC.** ORC-002 stops at LR-10. LR-11 reservation: research.md §4 anticipates ORC-004 worktree drift; this SPEC claims LR-12 (matrix drift), LR-13 (enum validation; renamed from spec REQ-012's "AGT_INVALID_FRONTMATTER" since SPEC-V3-AGT-001 owns the schema validator and this SPEC adds only the effort-specific surface), LR-14 (fixed budget_tokens prohibition). Sequence: 11 reserved for ORC-004; 12-14 owned by ORC-003.
- **`AGT_INVALID_FRONTMATTER` is owned by SPEC-V3-AGT-001 (schema validator), not by `agent_lint.go`.** REQ-012 says "the frontmatter validator (per SPEC-V3-AGT-001 REQ-002) shall reject the agent". Plan binding: T-ORC003-25 adds an effort-enum test fixture into `internal/cli/agent_lint_test.go` (LR-13 surface) AND coordinates with the schema validator package (`internal/config/schema/agent.go` if accessible from this branch; otherwise dual gate via lint-side + schema-side, Run-phase decision per Step 2 base scan).
- **`expert-debug`, `expert-testing`, `expert-mobile` are pre-ORC-001-retired or out-of-roster.** The 17-agent roster post-ORC-001 does NOT include expert-debug, expert-testing, expert-mobile, manager-ddd, manager-tdd, builder-agent, builder-skill, builder-plugin, manager-brain, claude-code-guide. This SPEC scopes effort population to the 17 active v3r2 agents only; retired agent stubs (if still present) are out-of-scope per spec §1.2 Non-Goals.

### 1.3 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

- **RED**: Add new lint test fixtures in `internal/cli/agent_lint_test.go` (or sibling) that FAIL on current state:
  - TestLintLR12_MatrixDrift (FAIL — no matrix-drift detection exists yet)
  - TestLintLR13_InvalidEffortEnum (FAIL — no enum validation exists yet)
  - TestLintLR14_FixedBudgetTokens (FAIL — no fixed-budget detection exists yet)
  - TestAuthoringDocHasEffortMatrix (FAIL — `agent-authoring.md` lacks the 17-row table currently)
  - TestConstitutionCrossReference (FAIL — `moai-constitution.md` §Opus 4.7 Prompt Philosophy currently inlines effort guidance; cross-reference target does not yet exist)
- **GREEN**: Implement §1.4 deltas. Population of 17 agents' `effort:` field + lint rules + matrix table publication. All RED tests turn GREEN.
- **REFACTOR**: Move shared effort-matrix constant (`canonicalEffortMatrix map[string]string`) into `internal/cli/agent_lint.go` package-level variable for single-source-of-truth between LR-12 + future tooling (e.g., `moai agent show-effort` if added in v3.x). Document rationale via @MX:NOTE.

### 1.4 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Canonical 17-agent effort matrix table | `.claude/rules/moai/development/agent-authoring.md` (new "## Effort-Level Calibration Matrix" section ~30 LOC) + template parity at `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | REQ-ORC-003-001, REQ-ORC-003-002 |
| Constitution cross-reference | `.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy "Effort level selection" bullet (modify existing line ~1 LOC + 1 cross-ref link) + template parity | REQ-ORC-003-005 |
| 4 drift corrections (`high` → `xhigh`) | `.claude/agents/moai/{expert-security,evaluator-active,plan-auditor,expert-refactoring}.md` (1 line each × 2 trees = 8 file edits) | REQ-ORC-003-003, AC-ORC-003-10 |
| 13 effort populations | `.claude/agents/moai/{manager-cycle,manager-quality,manager-docs,manager-git,manager-project,expert-backend,expert-frontend,expert-devops,expert-performance,builder-platform,researcher}.md` (1 line each × 2 trees = 22 file edits) | REQ-ORC-003-003, AC-ORC-003-02 |
| LR-12 implementation (matrix drift) | `internal/cli/agent_lint.go` (+~40 LOC: `canonicalEffortMatrix` constant + `checkEffortMatrixDrift()` function + dispatch in `lintAgentFile`) | REQ-ORC-003-013, AC-ORC-003-05 |
| LR-13 implementation (enum validation) | `internal/cli/agent_lint.go` (+~25 LOC: `validEffortValues` set + `checkInvalidEffortEnum()` function) | REQ-ORC-003-012, AC-ORC-003-08 |
| LR-14 implementation (fixed budget_tokens prohibition) | `internal/cli/agent_lint.go` (+~25 LOC: regex-based body scan for `budget_tokens` literal + `checkFixedBudgetTokens()` function) | REQ-ORC-003-014, CI fixture |
| Lint help text update | `internal/cli/agent_lint.go` long description (+3 lines for LR-12/13/14) | doc consistency |
| LR-12/13/14 test fixtures | `internal/cli/agent_lint_test.go` (+~150 LOC: 6 new sub-tests with PASS/FAIL fixture pairs) | T-ORC003-22..25 |
| Optional `effort_drift` JSON field (REQ-011) | `internal/cli/agent_lint.go` JSON output struct (+~10 LOC: new `EffortDrift []DriftEntry` field, populated by checkEffortMatrixDrift in non-strict mode) | REQ-ORC-003-011 |
| Optional `effort_mapping` cross-ref (REQ-010) | (no code change in this SPEC; HRN-001 owns harness.yaml; this SPEC documents the relationship in agent-authoring.md matrix section comment) | REQ-ORC-003-010 |
| MIG-001 hook (REQ-007 advisory) | `.moai/specs/SPEC-V3R2-MIG-001/` cross-link added in research.md; no code change in this SPEC (MIG-001 owns its migrator) | REQ-ORC-003-007 |
| `moai agent lint --format=json` JSON output regression test | `internal/cli/agent_lint_test.go` (+~30 LOC: TestLintJSONOutputIncludesEffortDrift) | REQ-ORC-003-011 |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 4 files (per §6 below) | mx_plan |

Embedded-template parity is **applicable** because both `.claude/rules/moai/development/agent-authoring.md` and the 17 agent files exist under `internal/template/templates/.claude/`. `make build` regeneration required after edits.

### 1.5 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-ORC003-NN IDs.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-ORC-003-001 | Ubiquitous (matrix table in agent-authoring.md) | AC-01 | T-ORC003-04 (table insertion), T-ORC003-05 (template parity), T-ORC003-23 (TestAuthoringDocHasEffortMatrix) |
| REQ-ORC-003-002 | Ubiquitous (canonical matrix content for 17 agents) | AC-01, AC-02, AC-10 | T-ORC003-04 (table content), T-ORC003-09..21 (per-agent population) |
| REQ-ORC-003-003 | Ubiquitous (effort: in 17 agent frontmatters) | AC-02, AC-10 | T-ORC003-09..21 (17 agents × frontmatter edit) |
| REQ-ORC-003-004 | Ubiquitous (template-local parity) | AC-06 | T-ORC003-05, T-ORC003-26 (`make build` + `diff -r` gate) |
| REQ-ORC-003-005 | Ubiquitous (constitution cross-reference) | AC-07 | T-ORC003-06 (constitution edit), T-ORC003-22 (TestConstitutionCrossReference) |
| REQ-ORC-003-006 | Event-Driven (LR-03 promotion idempotent) | AC-03, AC-04 | T-ORC003-02 (LR-03 anchor + comment), T-ORC003-24 (TestLR03Severity regression test) |
| REQ-ORC-003-007 | Event-Driven (MIG-001 drift rewrite) | AC-09 | T-ORC003-07 (research.md cross-link to MIG-001 contract) — no code change in this SPEC |
| REQ-ORC-003-008 | State-Driven (matrix unique source) | AC-07 | T-ORC003-04 (table is canonical), T-ORC003-22 (constitution cross-ref test) |
| REQ-ORC-003-009 | State-Driven (LR-03 blocks new agent) | AC-04 | T-ORC003-24 (regression test on synthetic agent without effort) |
| REQ-ORC-003-010 | Optional (harness override) | (no AC; documented behavior) | T-ORC003-04 (matrix section comment notes HRN-001 override path) |
| REQ-ORC-003-011 | Optional (effort_drift JSON field) | (no AC; tested via T-ORC003-25) | T-ORC003-25 (TestLintJSONOutputIncludesEffortDrift) |
| REQ-ORC-003-012 | Unwanted (invalid enum rejection) | AC-08 | T-ORC003-03 (LR-13 implementation), T-ORC003-25 (TestLintLR13_InvalidEffortEnum) |
| REQ-ORC-003-013 | Unwanted (matrix drift CI fail) | AC-05 | T-ORC003-03 (LR-12 implementation), T-ORC003-25 (TestLintLR12_MatrixDrift) |
| REQ-ORC-003-014 | Unwanted (fixed budget_tokens rejected) | (CI fixture; no AC in spec.md but tested) | T-ORC003-03 (LR-14 implementation), T-ORC003-25 (TestLintLR14_FixedBudgetTokens) |

Coverage: **14 unique REQs (001..014) → 10 ACs (AC-01..10) → 27 tasks (T-ORC003-01..27)**.

→ All REQ IDs from spec §5 are mapped. AC-01..10 from spec §6 are mapped. REQ-010 is documentation-only (HRN-001 owns the harness override semantics); REQ-011 has CI-fixture coverage but no spec-level AC (advisory feature).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule).

### M1: Test scaffolding + matrix table publication (RED phase + canonical doc) — Priority P0

Reference existing tests: `internal/cli/agent_lint_test.go` (~750 LOC, 30+ test functions covering LR-01..LR-10).

Owner role: `expert-backend` (Go test) or direct `manager-cycle` execution.

Tasks:
- T-ORC003-01: Create RED test fixtures in `internal/cli/agent_lint_test.go`:
  - `TestLintLR12_MatrixDrift_DriftedAgent` (FAIL initially — checkEffortMatrixDrift not yet implemented)
    - Fixture: synthetic `expert-security.md` with `effort: high` (drift from xhigh)
    - Expected: 1 violation, rule="LR-12", severity=Error, message contains "ORC_EFFORT_MATRIX_DRIFT"
  - `TestLintLR12_MatrixDrift_CleanAgent` (PASS initially — no violations expected on clean agent)
    - Fixture: synthetic `expert-security.md` with `effort: xhigh` (matches matrix)
    - Expected: 0 LR-12 violations
  - `TestLintLR13_InvalidEffortEnum` (FAIL initially)
    - Fixture: agent with `effort: ultra` (not in {low, medium, high, xhigh, max})
    - Expected: 1 violation, rule="LR-13", message contains "AGT_INVALID_FRONTMATTER" or equivalent enum-rejection text
  - `TestLintLR14_FixedBudgetTokens` (FAIL initially)
    - Fixture: agent body containing `budget_tokens: 5000` literal
    - Expected: 1 violation, rule="LR-14", message contains "ORC_FIXED_BUDGET_PROHIBITED"
  - `TestAuthoringDocHasEffortMatrix` (FAIL initially)
    - Reads `.claude/rules/moai/development/agent-authoring.md`; greps for "Effort-Level Calibration Matrix" section heading + 17-row table
    - Expected: section present + all 17 agent names + 5-value enum coverage
  - `TestConstitutionCrossReference` (FAIL initially)
    - Reads `.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy; greps for cross-link to `agent-authoring.md`
    - Expected: cross-link present
- T-ORC003-02: Add LR-03 regression anchor: rename existing `TestCheckMissingEffort` to `TestLintLR03_MissingEffortIsError` (or add a new sub-test) that explicitly asserts `Severity == SeverityError` (defense against future severity downgrade). Update `internal/cli/agent_lint.go:382-396` comment block to reference SPEC-V3R2-ORC-003 (idempotent doc).
- T-ORC003-03 (deferred to M2 — listed here for plan-audit traceability): LR-12, LR-13, LR-14 implementation pre-stub.

Verification gate: New M1 RED tests fail on `go test ./internal/cli/ -run "TestLintLR12|TestLintLR13|TestLintLR14|TestAuthoringDocHasEffortMatrix|TestConstitutionCrossReference"`. Existing `TestCheckMissingEffort` still passes (proves LR-03 idempotent).

### M2: Lint rule implementation (GREEN seed) — Priority P0

Owner role: `expert-backend`.

Tasks:
- T-ORC003-03 (continued from M1): Implement LR-12, LR-13, LR-14 in `internal/cli/agent_lint.go`:
  - Add package-level constant:
    ```go
    // canonicalEffortMatrix is the SPEC-V3R2-ORC-003 canonical effort assignment
    // for the 17 v3r2 agents. The matrix is consumed by checkEffortMatrixDrift (LR-12).
    // @MX:ANCHOR @MX:REASON: Single source of truth for agent-effort calibration; downstream
    // consumers (HRN-001 effort_mapping, doctor agent show-effort) reference this constant.
    var canonicalEffortMatrix = map[string]string{
        "manager-spec":       "xhigh",
        "manager-strategy":   "xhigh",
        "manager-cycle":      "high",
        "manager-quality":    "high",
        "manager-docs":       "medium",
        "manager-git":        "medium",
        "manager-project":    "medium",
        "expert-backend":     "high",
        "expert-frontend":    "high",
        "expert-security":    "xhigh",
        "expert-devops":      "medium",
        "expert-performance": "high",
        "expert-refactoring": "xhigh",
        "builder-platform":   "medium",
        "evaluator-active":   "xhigh",
        "plan-auditor":       "xhigh",
        "researcher":         "xhigh",
    }

    var validEffortValues = map[string]struct{}{
        "low": {}, "medium": {}, "high": {}, "xhigh": {}, "max": {},
    }
    ```
  - Add `checkEffortMatrixDrift(file string, fm AgentFrontmatter) []LintViolation` (LR-12). Logic: if agent name (derived from filename) is in canonicalEffortMatrix AND fm.Effort != "" AND fm.Effort != canonicalEffortMatrix[name], emit Error.
  - Add `checkInvalidEffortEnum(file string, fm AgentFrontmatter) []LintViolation` (LR-13). Logic: if fm.Effort != "" AND not in validEffortValues, emit Error.
  - Add `checkFixedBudgetTokens(file string, content []byte) []LintViolation` (LR-14). Logic: regex `\bbudget_tokens\s*:\s*\d+\b` against body text (excluding code blocks if feasible; v1 simple regex sufficient). Emit Error per match.
  - Wire all three into `lintAgentFile` dispatcher between existing checkMissingEffort (LR-03) and checkDeadHooks (LR-04).
- T-ORC003-08: Update lint help text (`internal/cli/agent_lint.go` cobra Long description) to enumerate LR-12/13/14 with one-line descriptions.
- T-ORC003-04: Insert canonical matrix table into `.claude/rules/moai/development/agent-authoring.md`. Section structure:
  ```markdown
  ## Effort-Level Calibration Matrix (SPEC-V3R2-ORC-003)

  Canonical effort assignment for the 17 v3r2 agents. This table is the single source
  of truth; agent frontmatter `effort:` values must match. Lint rule LR-12 enforces
  drift detection on `moai agent lint`.

  | Agent | Effort | Rationale |
  |---|---|---|
  | manager-spec | xhigh | Reasoning-intensive: SPEC EARS authoring |
  | manager-strategy | xhigh | Reasoning-intensive: architecture decisions |
  | ... (17 rows total) |

  Harness Override: Per SPEC-V3R2-HRN-001, `harness.yaml` `effort_mapping.<level>` may
  override per-spawn effort at session level; agent frontmatter is the default when
  no harness override applies (REQ-ORC-003-010).
  ```
  Mirror to `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`.
- T-ORC003-05: Run `make build` to regenerate `internal/template/embedded.go`. Verify byte-identical local-vs-template via `diff -r .claude/rules/moai/development/ internal/template/templates/.claude/rules/moai/development/` (CLAUDE.local.md §2 gate).
- T-ORC003-06: Modify `.claude/rules/moai/core/moai-constitution.md` §Opus 4.7 Prompt Philosophy "Effort level selection" bullet to **cross-reference** the matrix instead of duplicating values:
  - Before (current): "Effort level selection: reasoning-intensive agents (...) → effort: xhigh or high; implementation agents (...) → effort: high; speed-critical agents (...) → effort: medium"
  - After: "Effort level selection: see canonical 17-agent matrix in `.claude/rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix (SPEC-V3R2-ORC-003). Reasoning-intensive agents target `xhigh`; implementation agents target `high`; template-driven and speed-critical agents target `medium`."
  - Mirror to template tree.
- T-ORC003-07: Add cross-link in `research.md` §X (post-plan addition) to SPEC-V3R2-MIG-001 migration contract acknowledging REQ-007 (drift rewrite during migration). No code in this SPEC.

Verification gate: M1 RED tests for LR-12/LR-13/LR-14/TestAuthoringDocHasEffortMatrix/TestConstitutionCrossReference now PASS. Existing tests still PASS.

### M3: 17-agent frontmatter population (GREEN main) — Priority P0

Owner role: `expert-backend` (mechanical edit) — applies the matrix to each agent file.

Tasks (each task = 1 agent × 2 trees = 2 file edits, frontmatter `effort:` line):

- T-ORC003-09: `manager-cycle.md` — `effort: high`
  - Pre-condition: ORC-001 has merged. If not (Step 2 base verification fails), apply fallback to `manager-ddd.md` and `manager-tdd.md` until consolidated agent lands.
- T-ORC003-10: `manager-quality.md` — `effort: high`
- T-ORC003-11: `manager-docs.md` — `effort: medium`
- T-ORC003-12: `manager-git.md` — `effort: medium`
- T-ORC003-13: `manager-project.md` — `effort: medium`
- T-ORC003-14: `expert-backend.md` — `effort: high`
- T-ORC003-15: `expert-frontend.md` — `effort: high`
- T-ORC003-16: `expert-devops.md` — `effort: medium`
- T-ORC003-17: `expert-performance.md` — `effort: high`
- T-ORC003-18: `builder-platform.md` — `effort: medium`
  - Pre-condition: ORC-001 has merged. Fallback to builder-agent/builder-skill/builder-plugin if not.
- T-ORC003-19: `researcher.md` — `effort: xhigh`
- T-ORC003-20: **DRIFT-3-A** `expert-security.md` — `effort: high → xhigh` (BC-V3R2-002)
- T-ORC003-21: **DRIFT-3-B/C/D** `evaluator-active.md` `high → xhigh`, `plan-auditor.md` `high → xhigh`, `expert-refactoring.md` `high → xhigh` (BC-V3R2-002, plan-binding 4 drifts per §1.2.1)

Each frontmatter edit is `effort: <value>` inserted alphabetically near `description:` (consistent with existing manager-spec.md / manager-strategy.md placement, frontmatter order: name → description → tools → memory → model → effort → status → ...). Edit pattern (per file):

```yaml
# Before (manager-quality.md):
---
name: manager-quality
description: ...
tools: ...
model: sonnet
hooks: ...
---

# After:
---
name: manager-quality
description: ...
tools: ...
model: sonnet
effort: high   # SPEC-V3R2-ORC-003 canonical matrix
hooks: ...
---
```

Mirror each edit to `internal/template/templates/.claude/agents/moai/<name>.md`.

Verification gate: After M3 complete, `moai agent lint --path .claude/agents/moai/` reports 0 LR-03 errors AND 0 LR-12 errors for the 17 v3r2 agents.

### M4: CI integration tests + JSON drift output (REFACTOR + new feature) — Priority P0

Owner role: `expert-backend`.

Tasks:
- T-ORC003-22: Implement `TestConstitutionCrossReference` GREEN verification (M1 fixture turns GREEN after M2-T6).
- T-ORC003-23: Implement `TestAuthoringDocHasEffortMatrix` GREEN verification (M1 fixture turns GREEN after M2-T4).
- T-ORC003-24: Add `TestLintLR03_MissingEffortIsError` regression test (per T-ORC003-02 anchor). Assertion: severity == Error, not Warning. Defends against future regression.
- T-ORC003-25: Implement `effort_drift` JSON output:
  - Modify JSON output struct in `internal/cli/agent_lint.go` to include optional field `EffortDrift []EffortDriftEntry` populated by checkEffortMatrixDrift.
  - Each entry: `{Agent string, DeclaredEffort string, ExpectedEffort string, File string}`.
  - Add `TestLintJSONOutputIncludesEffortDrift`: lint a synthetic drift fixture in `--format=json` mode, assert the JSON output contains the `effort_drift` array.
- T-ORC003-26: Run full template-local parity check: `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` AND `diff -r .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md` AND `diff -r .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md`. Exit 0 = byte-identical.

Verification gate: All M1 RED tests now GREEN. `go test ./internal/cli/ -run TestLint` PASS. `moai agent lint` clean for the 17 v3r2 agents.

### M5: Verification gates + audit consolidation — Priority P0

Owner role: `manager-cycle` + `manager-quality`.

Tasks:
- T-ORC003-27: Run full test suite: `go test ./... -race -count=1`. Ensure 0 regressions.
- T-ORC003-28: Run linter: `golangci-lint run`. Fix any issues introduced.
- T-ORC003-29: Verify `make build` succeeds and `internal/template/embedded.go` is regenerated correctly.
- T-ORC003-30: Update `CHANGELOG.md` Unreleased section with bullet entries:
  - "feat(agents/SPEC-V3R2-ORC-003): publish canonical 17-agent effort-level calibration matrix"
  - "fix(agents/SPEC-V3R2-ORC-003): correct effort drift on expert-security/evaluator-active/plan-auditor/expert-refactoring (high → xhigh)"
  - "feat(lint/SPEC-V3R2-ORC-003): add LR-12 (matrix drift), LR-13 (effort enum), LR-14 (fixed budget_tokens prohibition)"
  - "feat(lint/SPEC-V3R2-ORC-003): emit `effort_drift` array in `moai agent lint --format=json` output"
- T-ORC003-31: Add @MX tags per §6 below.
- T-ORC003-32: Re-run `moai agent lint --path .claude/agents/moai/` (manual). Confirm 0 LR-03 / 0 LR-12 / 0 LR-13 / 0 LR-14 errors for the 17 v3r2 roster.

Verification gate: All AC-ORC-003-01 through AC-ORC-003-10 verified per acceptance.md.

---

## 3. File-Level Modification Map

### 3.1 Files modified (existing)

| File | Lines added/changed | Purpose |
|------|---------------------|---------|
| `.claude/rules/moai/development/agent-authoring.md` | +30 LOC | Effort matrix table |
| `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` | +30 LOC | Template parity |
| `.claude/rules/moai/core/moai-constitution.md` | -3 LOC + 2 LOC = -1 net | Cross-reference replaces inline list |
| `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` | -1 net LOC | Template parity |
| `.claude/agents/moai/manager-cycle.md` (or manager-ddd/tdd if ORC-001 not merged) | +1 LOC | effort: high |
| `internal/template/templates/.claude/agents/moai/manager-cycle.md` | +1 LOC | Template parity |
| `.claude/agents/moai/manager-quality.md` | +1 LOC | effort: high |
| `internal/template/templates/.claude/agents/moai/manager-quality.md` | +1 LOC | Template parity |
| `.claude/agents/moai/manager-docs.md` | +1 LOC | effort: medium |
| `internal/template/templates/.claude/agents/moai/manager-docs.md` | +1 LOC | Template parity |
| `.claude/agents/moai/manager-git.md` | +1 LOC | effort: medium |
| `internal/template/templates/.claude/agents/moai/manager-git.md` | +1 LOC | Template parity |
| `.claude/agents/moai/manager-project.md` | +1 LOC | effort: medium |
| `internal/template/templates/.claude/agents/moai/manager-project.md` | +1 LOC | Template parity |
| `.claude/agents/moai/expert-backend.md` | +1 LOC | effort: high |
| `internal/template/templates/.claude/agents/moai/expert-backend.md` | +1 LOC | Template parity |
| `.claude/agents/moai/expert-frontend.md` | +1 LOC | effort: high |
| `internal/template/templates/.claude/agents/moai/expert-frontend.md` | +1 LOC | Template parity |
| `.claude/agents/moai/expert-security.md` | ±1 LOC | effort: high → xhigh (DRIFT-3-A) |
| `internal/template/templates/.claude/agents/moai/expert-security.md` | ±1 LOC | Template parity |
| `.claude/agents/moai/expert-devops.md` | +1 LOC | effort: medium |
| `internal/template/templates/.claude/agents/moai/expert-devops.md` | +1 LOC | Template parity |
| `.claude/agents/moai/expert-performance.md` | +1 LOC | effort: high |
| `internal/template/templates/.claude/agents/moai/expert-performance.md` | +1 LOC | Template parity |
| `.claude/agents/moai/expert-refactoring.md` | ±1 LOC | effort: high → xhigh (DRIFT-3-D) |
| `internal/template/templates/.claude/agents/moai/expert-refactoring.md` | ±1 LOC | Template parity |
| `.claude/agents/moai/builder-platform.md` (or builder-* if ORC-001 not merged) | +1 LOC | effort: medium |
| `internal/template/templates/.claude/agents/moai/builder-platform.md` | +1 LOC | Template parity |
| `.claude/agents/moai/evaluator-active.md` | ±1 LOC | effort: high → xhigh (DRIFT-3-B) |
| `internal/template/templates/.claude/agents/moai/evaluator-active.md` | ±1 LOC | Template parity |
| `.claude/agents/moai/plan-auditor.md` | ±1 LOC | effort: high → xhigh (DRIFT-3-C) |
| `internal/template/templates/.claude/agents/moai/plan-auditor.md` | ±1 LOC | Template parity |
| `.claude/agents/moai/researcher.md` | +1 LOC | effort: xhigh |
| `internal/template/templates/.claude/agents/moai/researcher.md` | +1 LOC | Template parity |
| `internal/cli/agent_lint.go` | +90 LOC | LR-12 + LR-13 + LR-14 + canonicalEffortMatrix + JSON drift output |
| `internal/cli/agent_lint_test.go` | +180 LOC | 6 new sub-tests |
| `CHANGELOG.md` | +4 lines | Unreleased entry |

### 3.2 Files created (new)

None. (All work is additive frontmatter or function additions to existing files.)

### 3.3 Files removed

None.

### 3.4 Files NOT modified (out-of-scope)

- `.claude/agents/moai/manager-spec.md` — already at `effort: xhigh` (matches matrix)
- `.claude/agents/moai/manager-strategy.md` — already at `effort: xhigh` (matches matrix)
- `.claude/agents/moai/manager-brain.md` — NOT in v3r2 17-agent roster (out-of-scope per spec §1.2 Non-Goals)
- `.claude/agents/moai/expert-debug.md` — RETIRED by ORC-001 (if file still present, untouched)
- `.claude/agents/moai/expert-mobile.md` — NOT in v3r2 17-agent roster (out-of-scope; expert-backend covers mobile in v3r2 per ORC-001)
- `.claude/agents/moai/expert-testing.md` — RETIRED by ORC-001
- `.claude/agents/moai/manager-ddd.md`, `manager-tdd.md`, `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` — RETIRED by ORC-001
- `.claude/agents/moai/claude-code-guide.md` — NOT in v3r2 17-agent roster (out-of-scope)
- `internal/config/schema/agent.go` — schema validator territory of SPEC-V3-AGT-001 (REQ-012 routes invalid-enum rejection through agent_lint.go LR-13 surface only, per §1.2.1 acknowledged discrepancy)
- `harness.yaml` `effort_mapping` — owned by SPEC-V3R2-HRN-001 (REQ-010 documentation-only cross-reference)

---

## 4. Technical Approach

### 4.1 Canonical matrix as Go constant

Single source of truth in `internal/cli/agent_lint.go` (`canonicalEffortMatrix` map). The map is the **canonical machine-readable form**; `agent-authoring.md` is the **canonical human-readable form**. Both must agree.

Drift detection (LR-12 logic):

```go
func checkEffortMatrixDrift(file string, fm AgentFrontmatter) []LintViolation {
    name := agentNameFromFile(file)  // "expert-security" from path/expert-security.md
    expected, inMatrix := canonicalEffortMatrix[name]
    if !inMatrix {
        return nil  // out-of-roster agent; LR-12 does not apply
    }
    if fm.Effort == "" {
        // LR-03 already covers missing-effort case; LR-12 does not double-fire
        return nil
    }
    if fm.Effort != expected {
        return []LintViolation{{
            Rule:     "LR-12",
            Severity: SeverityError,
            File:     file,
            Line:     findFrontmatterLine(file, "effort:"),
            Message:  fmt.Sprintf("ORC_EFFORT_MATRIX_DRIFT: effort: %s drifts from canonical matrix value %s for agent %s (SPEC-V3R2-ORC-003 canonical matrix)", fm.Effort, expected, name),
        }}
    }
    return nil
}
```

### 4.2 Enum validation (LR-13)

Lightweight set membership check; no schema parser needed:

```go
func checkInvalidEffortEnum(file string, fm AgentFrontmatter) []LintViolation {
    if fm.Effort == "" {
        return nil  // LR-03 covers missing
    }
    if _, ok := validEffortValues[fm.Effort]; !ok {
        return []LintViolation{{
            Rule:     "LR-13",
            Severity: SeverityError,
            File:     file,
            Line:     findFrontmatterLine(file, "effort:"),
            Message:  fmt.Sprintf("AGT_INVALID_FRONTMATTER (effort): value %q is not in {low, medium, high, xhigh, max}", fm.Effort),
        }}
    }
    return nil
}
```

### 4.3 Fixed budget_tokens detection (LR-14)

Body scan via regex; matches `budget_tokens: <integer>` outside code blocks. v1 implementation: simple regex against full body text (false positives in code blocks are acceptable initial trade-off; refinement in v3.1 if false-positive rate noted):

```go
var fixedBudgetTokensRegex = regexp.MustCompile(`\bbudget_tokens\s*:\s*\d+\b`)

func checkFixedBudgetTokens(file string, content []byte) []LintViolation {
    matches := fixedBudgetTokensRegex.FindAllIndex(content, -1)
    var violations []LintViolation
    for _, match := range matches {
        // Compute line number from byte offset
        lineNum := bytes.Count(content[:match[0]], []byte("\n")) + 1
        violations = append(violations, LintViolation{
            Rule:     "LR-14",
            Severity: SeverityError,
            File:     file,
            Line:     lineNum,
            Message:  "ORC_FIXED_BUDGET_PROHIBITED: Opus 4.7 Adaptive Thinking rejects fixed budget_tokens (HTTP 400). Use effort: <level> instead.",
        })
    }
    return violations
}
```

### 4.4 JSON drift output (REQ-011)

Augment existing JSON output struct (in `internal/cli/agent_lint.go`):

```go
type LintReport struct {
    Violations  []LintViolation     `json:"violations"`
    EffortDrift []EffortDriftEntry  `json:"effort_drift,omitempty"`  // NEW
}

type EffortDriftEntry struct {
    Agent           string `json:"agent"`
    DeclaredEffort  string `json:"declared_effort"`
    ExpectedEffort  string `json:"expected_effort"`
    File            string `json:"file"`
}
```

Population: during `lintAgentFile`, when LR-12 fires, also append to `report.EffortDrift`. The JSON output remains backward-compatible (omitempty on EffortDrift array).

### 4.5 Backward compatibility

- 4 drift corrections (BC-V3R2-002): expert-security, evaluator-active, plan-auditor, expert-refactoring all upgrade from `high` to `xhigh`. Reasoning depth increases for these agents — latency may increase 10-30% on Opus 4.7 (Adaptive Thinking allocates more reasoning tokens). Mitigation: harness routing in HRN-001 can override per harness level (minimal/standard get `medium` regardless of frontmatter).
- LR-03 promotion (REQ-006) is **operationally idempotent** — already enforced as Error since ORC-002. No runtime breaking change.
- LR-12/13/14 are NEW errors. PRs touching agent frontmatter must comply or fail CI. Migration-time migrator (SPEC-V3R2-MIG-001) handles legacy v2 agent migration auto-rewriting drifted values (REQ-007).

### 4.6 Cross-platform behavior

- All edits are pure markdown/YAML frontmatter; platform-neutral.
- LR-12/13/14 are pure Go regex + map lookup; platform-neutral.

---

## 5. Quality Gates

Per `.moai/config/sections/quality.yaml`:

| Gate | Requirement | Verification command |
|------|-------------|-----------------------|
| Coverage | ≥ 85% per modified file | `go test -cover ./internal/cli/` |
| Race | `go test -race ./...` clean | `go test -race -count=1 ./...` |
| Lint | golangci-lint clean | `golangci-lint run` |
| Build | embedded.go regenerated | `make build` |
| Template parity | `diff -r` byte-identical | `diff -r .claude/ internal/template/templates/.claude/` |
| Agent lint | 0 LR-03/12/13/14 errors on 17 v3r2 roster | `moai agent lint --path .claude/agents/moai/ \| grep -E "LR-(03\|12\|13\|14)"` |
| MX | @MX tags applied per §6 | `moai mx scan internal/cli/` |

---

## 6. @MX Tag Plan (mx_plan)

Apply per `.claude/rules/moai/workflow/mx-tag-protocol.md`. Language: ko (per `code_comments: ko`).

| File | Tag | Reason |
|------|-----|--------|
| `internal/cli/agent_lint.go:canonicalEffortMatrix` | `@MX:ANCHOR @MX:REASON: SPEC-V3R2-ORC-003 17-에이전트 effort 매트릭스의 단일 진실 공급원; LR-12 drift detection + (향후) HRN-001 effort_mapping + 잠재적 doctor 명령 모두 본 상수 참조` | fan_in ≥ 3 (LR-12 + HRN-001 cross-ref + doctor agent show-effort 향후) |
| `internal/cli/agent_lint.go:checkEffortMatrixDrift` | `@MX:NOTE: LR-12 — 17-에이전트 매트릭스 drift 감지; spec.md §1.1 의 R5 audit 32% drift 비율을 0% 로 차단; out-of-roster 에이전트 (e.g., manager-brain) 는 LR-12 적용 제외` | non-obvious business rule (out-of-roster carve-out) |
| `internal/cli/agent_lint.go:checkFixedBudgetTokens` | `@MX:WARN @MX:REASON: budget_tokens 정규식 v1은 code block 내 false positive 가능; Opus 4.7 Adaptive Thinking 가 HTTP 400 으로 fixed budget 거부하므로 false positive 비용보다 회귀 비용이 높음 — v1 기준 채택, v3.1에서 code-block-aware 확장 검토` | external constraint + acknowledged simplification |
| `.claude/rules/moai/development/agent-authoring.md:Effort-Level Calibration Matrix` | `@MX:ANCHOR @MX:REASON: 17-에이전트 effort 매트릭스의 인간 가독 single source; LR-12 의 Go 상수와 동기화 필수; 매트릭스 변경은 SPEC-V3R2-CON-002 graduation protocol 통과 필요` | high fan_in + invariant contract |

---

## 7. Risk Mitigation Plan (spec §8 risks → run-phase tasks)

| spec §8 risk | Mitigation in run-phase |
|--------------|--------------------------|
| Row 1 — `high → xhigh` latency regression | T-ORC003-30 CHANGELOG entry + BC-V3R2-002 declaration; HRN-001 harness routing override available; 30-day telemetry window post-merge |
| Row 2 — `xhigh` over-invokes on trivial tasks | HRN-001 `harness: minimal` override path (out-of-scope this SPEC; documented in T-ORC003-04 matrix section comment) |
| Row 3 — New agent without effort: blocked | T-ORC003-24 (TestLintLR03_MissingEffortIsError regression) + T-ORC003-32 manual lint verification |
| Row 4 — Matrix-constitution drift | T-ORC003-04 (matrix in agent-authoring.md) + T-ORC003-06 (constitution cross-references, no duplicate) + T-ORC003-22 (TestConstitutionCrossReference) |
| Row 5 — Opus 4.7 guidance evolves | Out-of-scope (HRN-001 harness `model_upgrade_review` checklist) |
| Row 6 — Fixed `budget_tokens` reintroduced | T-ORC003-03 LR-14 implementation + T-ORC003-25 TestLintLR14_FixedBudgetTokens fixture |
| Row 7 — Non-matrix agents (98/99 commands) | scope filter to `.claude/agents/moai/*.md`; LR-12 out-of-roster carve-out covers exotic agents (manager-brain, claude-code-guide) |
| Row 8 — Format drift breaks parsing | T-ORC003-23 TestAuthoringDocHasEffortMatrix grep ensures section heading + 17 rows present |

---

## 8. Dependencies (status as of `3356aa9a9`)

### 8.1 Blocking (consumed)

- ✅ **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification) — assumed merged per spec §9.1.
- ⚠ **SPEC-V3R2-ORC-001** (17-agent roster) — status TBD at run-time. If unmerged: T-ORC003-09 (manager-cycle) + T-ORC003-18 (builder-platform) apply fallback to original retiree agents (manager-ddd/tdd, builder-agent/skill/plugin) per tasks.md §4.1.
- ✅ **SPEC-V3R2-ORC-002** (lint LR-03/04/05/06/07/08/09/10) — assumed merged. LR-03 already at Error severity per `internal/cli/agent_lint.go:386`. LR-11 reserved for ORC-004; LR-12/13/14 owned by THIS SPEC (ORC-003).

### 8.2 Blocked by (none active)

All blockers are advisory. ORC-001 fallback path documented; ORC-002 already-implemented status verified.

### 8.3 Blocks (downstream consumers)

- **SPEC-V3R2-HRN-001** (Harness routing) — `effort_mapping` in `harness.yaml` aligns to this matrix.
- **SPEC-V3R2-MIG-001** (migrator) — REQ-007 says migrator rewrites drifted v2 effort values to match this matrix; migrator implementation references `canonicalEffortMatrix` constant from this SPEC.
- **SPEC-V3R2-ORC-004** (worktree MUST) — claims LR-11; this SPEC reserves the slot at lint package level.

---

## 9. Verification Plan

### 9.1 Pre-merge verification (run-phase end)

- [ ] All 27 tasks (T-ORC003-01..27) complete per tasks.md (with M3 having parallel sub-tasks T-ORC003-09..21 across 13 agents)
- [ ] All 10 ACs (AC-ORC-003-01..10) verified per acceptance.md
- [ ] `go test -race -count=1 ./...` PASS (no regressions)
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] `moai agent lint --path .claude/agents/moai/` reports 0 LR-03 / 0 LR-12 / 0 LR-13 / 0 LR-14 errors for the 17 v3r2 roster
- [ ] `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/` byte-identical
- [ ] `diff -r .claude/rules/moai/development/ internal/template/templates/.claude/rules/moai/development/` byte-identical
- [ ] CHANGELOG entry written in Unreleased section
- [ ] @MX tags applied per §6 (verified via `moai mx scan`)

### 9.2 Plan-auditor target

- [ ] All 14 unique REQs mapped in §1.5 traceability matrix
- [ ] All 10 ACs mapped to ≥1 task
- [ ] No orphan tasks (every task supports ≥1 REQ)
- [ ] research.md evidence anchors cited (≥30 per §1 mandate)
- [ ] §1.2.1 explicitly addresses spec.md drift count discrepancy (3 → 4)
- [ ] §1.2.1 acknowledges LR-03 idempotency (already at Error severity)
- [ ] BC-V3R2-002 implications documented (4 agents `high → xhigh`)
- [ ] Worktree-base alignment per Step 2 (run-phase) called out
- [ ] §6 mx_plan covers ≥3 of {ANCHOR, WARN, NOTE} types (covered: 2 ANCHOR + 1 WARN + 1 NOTE = all 3)
- [ ] No time estimates anywhere (P0 priority labels only)
- [ ] Parallel SPEC isolation: this plan touches only `.claude/agents/moai/`, `.claude/rules/moai/{core,development}/`, `internal/cli/agent_lint{,_test}.go`, `internal/template/templates/.claude/`, `CHANGELOG.md`.

### 9.3 Plan-auditor risk areas (front-loaded mitigations)

- **Risk: Drift count mismatch (spec says 3, plan binds 4)** → addressed in §1.2.1 Acknowledged Discrepancies + research.md evidence; sync-phase HISTORY entry will reconcile.
- **Risk: LR-03 already Error → REQ-006 promotion is no-op** → addressed in §1.2.1; promotion REQ becomes regression-test verification (T-ORC003-24); idempotency is not a defect — it is operational reality.
- **Risk: ORC-001 not merged at run-time → manager-cycle/builder-platform files don't exist** → addressed in §1.2 (fallback to retiree agents) + §8.1 + tasks.md §4.1 dependency-resolution flow.
- **Risk: AGT_INVALID_FRONTMATTER ownership conflict (SPEC-V3-AGT-001 schema validator vs LR-13 lint)** → addressed in §1.2.1; LR-13 is the lint-side defense-in-depth surface; schema validator owns runtime enforcement; both gates fire with same error code for consistency.
- **Risk: Fixed budget_tokens regex false positives in code blocks** → addressed in §6 MX:WARN tag + §4.3 acknowledged simplification; v1 simple regex; v3.1 refinement deferred.
- **Risk: `3356aa9a9` baseline drift if main advances during plan PR review** → run-phase explicitly rebases on `origin/main` (Step 2 `moai worktree new --base origin/main`) per spec-workflow.md.
- **Risk: manager-cycle / builder-platform files renamed after ORC-001 merges, breaking T-ORC003-09/18 file paths** → addressed via fallback documented in §1.2 + tasks.md §4.1; if file renamed but content equivalent, edit transfers cleanly.

---

## 10. Run-Phase Entry Conditions

After plan PR squash-merged into main:

1. `git checkout main && git pull` (host checkout).
2. `moai worktree new SPEC-V3R2-ORC-003 --base origin/main` per Step 2 spec-workflow.md.
3. `cd ~/.moai/worktrees/moai-adk/SPEC-V3R2-ORC-003`.
4. `git rev-parse --show-toplevel` should output the worktree path (Block 0 verification per session-handoff.md).
5. `git rev-parse HEAD` should match plan-merge commit SHA on main.
6. Verify ORC-001 status: `git log --oneline origin/main | head -20 | grep ORC-001` — if MERGED, proceed normally; if NOT MERGED, apply M3 fallback path documented in tasks.md §4.1.
7. `/moai run SPEC-V3R2-ORC-003` invokes Phase 0.5 plan-audit gate, then proceeds to M1.

---

Version: 0.1.0
Status: Plan artifact for SPEC-V3R2-ORC-003
Run-phase methodology: TDD (per `.moai/config/sections/quality.yaml` `development_mode: tdd`)
Estimated artifacts: 0 new files + 1 doc table (agent-authoring.md) + 1 doc cross-ref (constitution) + 17 agent frontmatter edits × 2 trees + 3 new lint rules (LR-12/13/14) + 6 new test fixtures + CHANGELOG = ~430 LOC delta
