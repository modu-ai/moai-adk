---
id: SPEC-V3R5-WORKFLOW-OPT-001
title: "Workflow Optimization — 8-Layer Improvement Plan"
version: "0.2.0"
status: implemented
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
module: ".claude/rules/moai + .moai/config/sections/workflow.yaml + internal/harness/capture + .claude/agents/moai/plan-auditor.md"
lifecycle: spec-anchored
tags: "workflow, optimization, 8-layer, manager-develop-prompt, agent-teams, ci-watch, lessons-autocapture, plan-auditor-d7-d8, mega-sprint, w-meta"
---

# SPEC-V3R5-WORKFLOW-OPT-001 — Workflow Optimization (8-Layer Improvement Plan)

## HISTORY

| Version | Date       | Author    | Change                                                                                                                                  |
|---------|------------|-----------|----------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-20 | GOOS Kim  | Initial draft — 8-layer workflow optimization plan derived from W3 HARNESS-AUTONOMY-001 meta-analysis. Source: `.moai/research/workflow-opt-vision-2026-05-20.md`. |
| 0.2.0   | 2026-05-20 | GOOS Kim  | run-phase complete — M1+M2+M4+M5+M6 done; M3 (Layer F capture) DEFERRED pending W3 PR #1024 merge. Status `draft → implemented`. Dogfooding wall-time: 25.0 min (target ≤ 30 min, 73% reduction vs W3 91 min). 12/14 ACs PASS, 2/14 DEFERRED (AC-WO-005, AC-WO-010 M3-scope). |

---

## 1. Overview

### 1.1 Background

W3 HARNESS-AUTONOMY-001 run-phase measurement reported wall-time **91 minutes** vs target **30 minutes (+200%)**. Root-cause meta-analysis (`feedback_w3_metaanalysis_lessons.md`) identified 4 defect patterns that triggered 3 manager-develop re-delegations plus 1 orchestrator-direct fix:

| Defect                                       | Direct loss | Re-delegation cycle    | Verification wait |
|----------------------------------------------|-------------|------------------------|-------------------|
| spec-lint h3 heading omission                | 4 min       | 0                      | 3 min             |
| AC-HRA-009 V3R4 retirement conflict          | 14 min      | 1 (manager-develop #2) | 10 min            |
| Windows syscall.Flock build-tag omission     | 10 min      | 0 (orchestrator fix)   | 5 min             |
| observer.go path resolution bug              | 5 min       | 0                      | 0                 |
| **Sum**                                      | **33 min**  | **+1 cycle**           | **18 min**        |

Wall-time decomposition (5,460 s total):
- Delegated implementation work: 46 min (50.5%)
- Verification + idle wait + decision wait: 45 min (49.5%)

→ Verification/wait equals delegation ≈ **1:1 ratio = bottleneck**.

### 1.2 Purpose

Generalize the four W3 defect patterns into reusable orchestrator workflow improvements across 8 layers (A–H). The SPEC itself serves as the first dogfooding subject (AC-WO-001).

### 1.3 Architecture (4 domains)

| Domain                | Layers covered          | Artifact type                                                       | Risk    | Dependencies   |
|-----------------------|-------------------------|---------------------------------------------------------------------|---------|----------------|
| R (rule-only)         | A (remainder), C, D, E, H | `.claude/rules/moai/` markdown + template mirror                  | Low     | Independent    |
| C (config)            | B                       | `.moai/config/sections/workflow.yaml` + experimental flag           | Medium  | R recommended  |
| F (Go code — capture) | F                       | `internal/harness/capture/` package extension (defect detection)    | High    | R, C, W3       |
| G (agent prompt)      | G                       | `.claude/agents/moai/plan-auditor.md` + `evaluator-profiles/`       | Medium  | R              |

### 1.4 Quantitative Goals

| Metric                                   | W3 baseline | Target           |
|------------------------------------------|-------------|------------------|
| Run-phase wall-time                      | 91 min      | ≤ 30 min         |
| manager-develop delegation count (1-pass) | 3 (33%)     | ≤ 1 (≥ 80%)      |
| CI serial wait time                      | 15 min      | ≤ 3 min          |
| Verification serial time                 | 10 min      | ≤ 3 min          |

### 1.5 Affected Modules (PRESERVE / EXTEND map)

| Path                                                             | Strategy | Note                                                                                              |
|------------------------------------------------------------------|----------|---------------------------------------------------------------------------------------------------|
| `.claude/rules/moai/development/manager-develop-prompt-template.md` | EXTEND | File exists (Layer A bootstrap done); add template mirror + cross-refs.                          |
| `.claude/rules/moai/workflow/ci-watch-protocol.md`               | EXTEND   | Layer C — append background-watch standardization.                                                |
| `.claude/rules/moai/core/agent-common-protocol.md`               | EXTEND   | Layer D + H — strengthen Parallel Execution; tool optimization patterns.                          |
| `.claude/rules/moai/workflow/spec-workflow.md`                   | EXTEND   | Layer E — Phase Transitions allow plan-PR overlap with run start; Plan Audit Gate skip-policy.   |
| `.claude/rules/moai/workflow/agent-teams-pattern.md`             | NEW      | Layer B — 5-teammate pattern documentation.                                                       |
| `.claude/rules/moai/workflow/verification-batch-pattern.md`      | NEW      | Layer D — verification grouping pattern.                                                          |
| `.moai/config/sections/workflow.yaml`                            | EXTEND   | Layer B — add role_profiles; preserve mode dispatch.                                              |
| `internal/harness/capture/`                                      | EXTEND   | Layer F — package created by W3 (PR #1024); add NEW files only (`defect_detector.go` etc.).      |
| `.claude/agents/moai/plan-auditor.md`                            | EXTEND   | Layer G — add D7 (cross-SPEC) + D8 (cross-platform) dimensions. PRESERVE D1–D6.                  |
| `.moai/config/evaluator-profiles/{default,frontend}.md`          | EXTEND   | Layer G — register D7/D8 dimension weights.                                                       |
| `internal/template/templates/.claude/rules/moai/**`              | MIRROR   | Layer A + M1 — every rule edit must be mirrored; `make build` mandatory.                          |

---

## 2. Requirements (EARS format)

### 2.1 Ubiquitous Requirements

- **REQ-WO-001**: The orchestrator **shall** include the 5-section structure defined in `manager-develop-prompt-template.md` (Context / Known Issues / Pre-flight / Constraints / Self-Verification Deliverables) in every `manager-develop` delegation prompt.
- **REQ-WO-002**: The orchestrator **shall** execute every read-only verification (test / coverage / grep / sentinel / CLI / benchmark / lint) as a single-turn multi-Bash batch, never serially across turns.
- **REQ-WO-003**: While CI checks are pending, the orchestrator **shall not** remain idle; it **shall** invoke `gh pr checks --watch` via `run_in_background: true` and continue other productive work concurrently.
- **REQ-WO-004**: Every edit to a file under `.claude/rules/moai/**` **shall** be mirrored to the corresponding path under `internal/template/templates/.claude/rules/moai/**`, and `make build` **shall** be re-run before the change is committed.

### 2.2 State-Driven Requirements

- **REQ-WO-010**: **While** `workflow.team.enabled: true` **and** the SPEC declares ≥5 independent implementation packages, the orchestrator **shall** be permitted to spawn 5 implementer teammates plus 1 tester plus 1 reviewer via the Agent Teams API.
- **REQ-WO-011**: **While** `plan-auditor` returns verdict PASS with overall score ≥ 0.90, the orchestrator **shall** skip the run-phase Plan Audit Gate re-execution and record the skip decision in the run-phase delegation prompt.
- **REQ-WO-012**: **While** the SPEC body declares dependency on a package whose status is `retired` or `superseded`, `plan-auditor` D7 **shall** raise a BLOCKING finding referencing the conflicting SPEC ID.
- **REQ-WO-013**: **While** the SPEC body references `syscall` package (any subpath) without a corresponding `//go:build` constraint declaration, `plan-auditor` D8 **shall** raise a BLOCKING finding.

### 2.3 Event-Driven Requirements

- **REQ-WO-020**: **When** `manager-develop` subagent reports a defect via the SubagentStop hook, the `internal/harness/capture/` defect-detector **shall** classify the defect against the 8 known-issue taxonomy (B1–B8) and append an observation entry to the lessons memory log.
- **REQ-WO-021**: **When** a plan-PR is opened (via `gh pr create` for a branch matching `plan/SPEC-*`), the `plan-auditor` D7 dimension **shall** scan the SPEC body for cross-SPEC references and verify each referenced SPEC's current status against `.moai/specs/`.
- **REQ-WO-022**: **When** a SPEC body adds or modifies a Go file that imports `syscall`, `plan-auditor` D8 **shall** require either a documented `//go:build` constraint or an explicit cross-platform justification in the SPEC.
- **REQ-WO-023**: **When** a new defect pattern is added to the lessons memory log by Layer F, subsequent `manager-develop` delegation prompt generation **shall** auto-prepend the matching lessons entries via keyword matching, without manual orchestrator intervention.

### 2.4 Optional Requirements

- **REQ-WO-030**: **Where** a SPEC is classified as high-stakes (priority `P0` or harness level `thorough`), the orchestrator **may** spawn `evaluator-active` in parallel with `manager-develop` so that quality assessment proceeds concurrently with implementation.

### 2.5 Unwanted Behavior Requirements

- **REQ-WO-040**: **If** a delegation prompt to `manager-develop` omits Section B (Known Issues B1–B8) auto-injection, **then** the orchestrator **shall not** dispatch the delegation; the omission **shall** be reported as a self-discipline failure.
- **REQ-WO-041**: **If** any rule file under `.claude/rules/moai/**` is committed without its template mirror in `internal/template/templates/**`, **then** CI **shall** fail with `RuleTemplateMirrorDrift` (introduced by Layer A's CI guard test).

---

## 3. Constraints

### 3.1 Functional Constraints

- **C-WO-001 (Subagent Boundary)**: `internal/harness/capture/` extensions **shall not** invoke `AskUserQuestion` or any user-interaction tool. Verification: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/capture/ | grep -v "_test.go" | grep -v "// "` must yield 0 matches.
- **C-WO-002 (Cross-platform discipline)**: Layer F defect-detector code **shall not** introduce any `syscall` imports. Pattern detection is heuristic-only (string and AST-shape matching).
- **C-WO-003 (No reversal of W3 PRESERVE list)**: Layer F additions **shall** be net-new files only (e.g., `defect_detector.go`, `defect_detector_test.go`). Existing W3-shipped capture files **shall not** be modified.
- **C-WO-004 (No reversal of plan-auditor D1–D6)**: Layer G **shall** add D7 and D8 only. Existing dimensions D1–D6 **shall** be preserved verbatim.
- **C-WO-005 (No reversal of workflow.yaml mode dispatch)**: Layer B **shall** add `team.role_profiles` and `team.role_profile_keys` keys only. Existing `mode dispatch` (from SPEC-V3R2-WF-003/WF-004) **shall** be preserved.

### 3.2 Operational Constraints

- **C-WO-010 (Template mirror)**: Every rule edit under `.claude/rules/moai/**` requires (1) mirror copy to `internal/template/templates/.claude/rules/moai/**`, (2) `make build`, (3) verification via `git diff internal/template/embedded.go` is non-empty.
- **C-WO-011 (Coverage threshold)**: New code in `internal/harness/capture/defect_detector.go` **shall** maintain ≥ 90% line coverage (Layer F invariant).
- **C-WO-012 (CI-Tier independence)**: SPEC AC verification **shall** test each of three CI tiers (spec-lint, golangci-lint, Test-per-OS) independently — passing one does not imply passing another (lessons #19).
- **C-WO-013 (Frontmatter canonical schema)**: This SPEC's frontmatter uses the canonical 12-field schema per `.claude/rules/moai/development/spec-frontmatter-schema.md` with field names `created:` / `updated:` / `tags:` (NOT snake_case aliases).

---

## 4. Out of Scope (Exclusions)

This SPEC explicitly excludes the following items. They are tracked as separate SPECs or follow-up work.

### 4.1 Out of Scope — Exclusion List

- **EXCL-WO-001 (W4 PROJECT-MEGA-001 harness self-improvement)**: This SPEC operates at the orchestrator workflow (meta) layer; harness model-layer self-improvement remains W4 scope. No overlap because the two layers are architecturally separated.
- **EXCL-WO-002 (PR #1024 W3 run-phase merge)**: User-natural merge area. This SPEC does not gate on PR #1024 status; M3 (capture extension) assumes W3 base files exist on `main` at run-phase start. If W3 has not yet merged, M3 starts from the W3 feature branch (rebase strategy).
- **EXCL-WO-003 (observer.go path resolution bug fix)**: Captured as separate SPEC `SPEC-V3R5-OBSERVER-PATH-001` (tentative). Layer F's defect-detector may pattern-match this defect class but does not modify `internal/harness/observer.go`.
- **EXCL-WO-004 (Lint baseline 2 issues cleanup)**: W1/W2 residual chicken-and-egg debt — captured as separate SPEC `SPEC-V3R5-LINT-DEBT-001` (tentative). This SPEC's AC verifies delta-only (NEW = 0); existing baseline is not touched.
- **EXCL-WO-005 (GitHub webhook integration for advanced CI)**: Out of scope. Layer C uses `gh pr checks --watch` only — no webhook subscription, no GitHub App, no API rate-limit risk.
- **EXCL-WO-006 (Other-domain meta-analysis)**: SPEC-creation phase, sync-phase, design-phase meta-analyses are not in scope. This SPEC's meta-analysis is bounded to run-phase orchestration.
- **EXCL-WO-007 (Cross-platform syscall introduction)**: This SPEC introduces no `syscall` imports anywhere. B1 (cross-platform build-tags) is addressed only by Layer G (plan-auditor D8 verification of OTHER SPECs).
- **EXCL-WO-008 (Untracked working-tree files)**: This SPEC's commits **shall not** include `.moai/harness/usage-log.jsonl`, `progress.md` files of other SPECs, or any `{}/` literal directory.

---

## 5. Acceptance Criteria Summary

The full Given-When-Then matrix is defined in [acceptance.md](./acceptance.md). Summary of binary ACs:

| AC ID       | Description                                                                                                                                                            |
|-------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| AC-WO-001   | Self-dogfooding wall-time of this SPEC's run-phase ≤ 30 minutes.                                                                                                       |
| AC-WO-002   | `manager-develop` delegation prompt template includes all 8 known-issue categories B1–B8 (grep verification).                                                          |
| AC-WO-003   | `internal/template/templates/` mirror is byte-equivalent for every modified rule; `make build` produces clean `embedded.go` regeneration.                              |
| AC-WO-004   | `plan-auditor` correctly detects a simulated V3R4-retirement cross-SPEC conflict and emits BLOCKING D7 finding in iteration 1.                                         |
| AC-WO-005   | `internal/harness/capture/defect_detector.go` unit-test coverage ≥ 90% (measured via `go test -coverprofile`).                                                         |
| AC-WO-006   | Layer C: `gh pr checks --watch` invocation in CI watch protocol uses `run_in_background: true` (rule body verification).                                               |
| AC-WO-007   | Layer D: At least one example in `agent-common-protocol.md` Parallel Execution section demonstrates 7-item read-only verification batch in one turn.                   |
| AC-WO-008   | Layer E: `spec-workflow.md` Phase Transitions section documents the Plan Audit Gate skip policy (PASS ≥ 0.90 threshold).                                               |
| AC-WO-009   | Layer B: `workflow.yaml` `team.role_profiles` map contains exactly 7 keys (5 implementer + tester + reviewer); existing keys (`auto_selection`, `default_model`) preserved. |
| AC-WO-010   | Layer F: Defect-detector classifies a synthetic `manager-develop` failure report into one of the 8 B-categories with ≥ 0.7 confidence (unit-test).                     |
| AC-WO-011   | Layer G: D7 dimension definition appears in `.claude/agents/moai/plan-auditor.md` and includes the cross-SPEC status check verb.                                       |
| AC-WO-012   | Layer G: D8 dimension definition appears in `.claude/agents/moai/plan-auditor.md` and references `syscall` + `//go:build` constraint requirement.                      |
| AC-WO-013   | Layer H: `agent-common-protocol.md` references `gh pr checks --json … \| jq` as the canonical CI status query pattern.                                                 |
| AC-WO-014   | spec-lint baseline regression check: NEW findings introduced by this SPEC's files = 0 (delta-only semantics).                                                          |

---

## 6. REQ ↔ AC Traceability

Detailed mapping appears in [acceptance.md](./acceptance.md) §6. Summary:

- REQ-WO-001 → AC-WO-002 (template content), AC-WO-003 (mirror)
- REQ-WO-002 → AC-WO-007 (batch example)
- REQ-WO-003 → AC-WO-006 (background watch)
- REQ-WO-004 → AC-WO-003, AC-WO-014
- REQ-WO-010 → AC-WO-009 (role_profiles config)
- REQ-WO-011 → AC-WO-008 (skip policy)
- REQ-WO-012 → AC-WO-004, AC-WO-011 (D7)
- REQ-WO-013 → AC-WO-012 (D8)
- REQ-WO-020 → AC-WO-005, AC-WO-010 (defect-detector)
- REQ-WO-021 → AC-WO-004, AC-WO-011
- REQ-WO-022 → AC-WO-012
- REQ-WO-023 → AC-WO-010 (auto-prepend)
- REQ-WO-030 → (manual verification, see acceptance.md §5)
- REQ-WO-040 → AC-WO-002 (self-discipline)
- REQ-WO-041 → AC-WO-014 (CI guard)

---

## 7. Cross-References

- Vision document: [.moai/research/workflow-opt-vision-2026-05-20.md](../../research/workflow-opt-vision-2026-05-20.md)
- Parent meta-analysis: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_w3_metaanalysis_lessons.md`
- Frontmatter schema SSOT: [.claude/rules/moai/development/spec-frontmatter-schema.md](../../../.claude/rules/moai/development/spec-frontmatter-schema.md)
- W3 (related, parallel): [SPEC-V3R5-HARNESS-AUTONOMY-001](../SPEC-V3R5-HARNESS-AUTONOMY-001/spec.md)
- Cross-SPEC reconciliation (workflow.yaml owner): SPEC-V3R2-WF-003, SPEC-V3R2-WF-004
- Lessons references: lessons #19 (CI 3-tier), lessons #21 (syscall build-tag)
