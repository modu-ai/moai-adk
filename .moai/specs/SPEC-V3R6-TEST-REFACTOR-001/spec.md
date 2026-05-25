---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — ATR-001 PROCEED-WITH-DEBT discharge"
version: "0.1.1"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0 follow-up"
module: "internal/template, internal/skills, internal/harness, internal/statusline"
lifecycle: spec-anchored
tags: "test-refactor, atr-001-debt-discharge, architectural-pivot, sprint-10"
depends_on: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
tier: M
---

# SPEC-V3R6-TEST-REFACTOR-001 — Go test suite refactor (ATR-001 PROCEED-WITH-DEBT discharge)

## Section A — Lifecycle Sync

| Field | Value |
|-------|-------|
| plan_commit_sha | pending |
| sync_commit_sha | pending |
| mx_commit_sha | pending |
| supersedes | (none) |
| superseded_by | (none) |
| anchor SPEC | SPEC-V3R6-AGENT-TEAM-REBUILD-001 (ATR-001 PROCEED-WITH-DEBT directive at progress.md §F.2.8) |

### A.1 Predecessor anchor

This SPEC discharges the PROCEED-WITH-DEBT directive issued at ATR-001 4-phase close (Mx commit `ccb4a14c0`). Verbatim from ATR-001 progress.md §F.2.8:

> "8 Go test fails (1 pre-existing path drift + 7 architectural-pivot consequences) deferred to follow-up SPEC-V3R6-TEST-REFACTOR-001 per user PROCEED-WITH-DEBT directive"

The ATR-001 spec.md / plan.md / acceptance.md / progress.md bodies are absolute SSOT (L48). This SPEC MUST NOT modify any predecessor SPEC artifact.

### A.2 Ground truth discovery method

Test baseline established by executing `go test ./...` at HEAD `e7b119924` (post-CATALOG-HASH-REGRESSION-CLEANUP-001 4-phase close). The 15 measured failures are NOT enumerated as immutable AC anchors in this body — instead, the §A.4 ground truth table is a regenerable snapshot. Run-phase manager-develop verifies ground truth re-measurement at its M1 entry and reports any drift via L66 working-tree race protocol.

### A.3 Baseline drift accounting

ATR-001 M8 measurement reported 8 failures. Current measurement reports 15 failures (+7 drift). The +7 drift accumulated across 2 sync-phase + 2 Mx-phase commits of CATALOG + ATR-001 cohort completion between the M8 baseline and this plan-phase commit. The classification of the 7 additional failures (architectural-pivot consequence vs cascade from catalog refresh vs pre-existing) is verified inline at run-phase M1 via package-level inspection.

### A.4 Ground truth — measured 2026-05-25 at HEAD `e7b119924`

| # | Test name | Package | Classification | First-discovery anchor |
|---|-----------|---------|----------------|------------------------|
| 1 | TestSubagentBoundary_NoAskUserQuestion | internal/harness | architectural-pivot consequence | ATR-001 hook boundary expansion |
| 2 | TestTemplateMirrorParity | internal/skills | cascade (catalog/template drift) | post-ATR-001 archive |
| 3 | TestSubSkillLOCCeiling | internal/skills | needs-investigation (LOC budget drift suspected) | cohort cascade |
| 4 | TestRenderPRSegment_Absence | internal/statusline | pre-existing (per ATR-001 §F.2.8) | predates ATR-001 |
| 5 | TestContractSchemaVerification | internal/template | architectural-pivot consequence | ATR-001 17→8 agent catalog |
| 6 | TestBackwardCompatibility | internal/template | architectural-pivot consequence | ATR-001 archived-agent rejection |
| 7 | TestContractAssertionsNaturalLanguage | internal/template | architectural-pivot consequence | ATR-001 catalog body refactor |
| 8 | TestAgentFrontmatterAudit | internal/template | architectural-pivot consequence | ATR-001 retained-agent frontmatter realignment |
| 9 | TestTemplateAgentsStructure | internal/template | architectural-pivot consequence | ATR-001 retained-agent structure |
| 10 | TestEmbeddedTemplates_AgentDefinitions | internal/template | architectural-pivot consequence | ATR-001 embedded.go regeneration |
| 11 | TestLoadCatalog | internal/template | architectural-pivot consequence | ATR-001 catalog.yaml 12-archived purge |
| 12 | TestAllAgentsInCatalog | internal/template | architectural-pivot consequence | ATR-001 catalog.yaml retained-list update |
| 13 | TestLoadEmbeddedCatalog_Success | internal/template | architectural-pivot consequence | ATR-001 embedded catalog refresh |
| 14 | TestRuleTemplateMirrorDrift | internal/template | cascade (NEW since ATR-001 M8) | post-ATR-001 rule template churn |
| 15 | TestRetirementCompletenessAssertion | internal/template | pre-existing path drift (per ATR-001 §F.2.8) | SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (`.claude/agents/moai/` → `.claude/agents/core/`) |

Aggregate classification:
- Pre-existing: 2 (rows 4, 15)
- Architectural-pivot consequence: 9 (rows 1, 5–13)
- Cascade (post-ATR-001 drift): 3 (rows 2, 14, and partially row 11 via catalog regen)
- Needs-investigation: 1 (row 3)

The "Needs-investigation" row 3 is resolved at run-phase M3 entry via single-test inspection before classification commits.

## Section B — Requirements (GEARS notation)

### B.1 Ubiquitous requirements (always active)

- **REQ-TST-001** (Ubiquitous): The Go test suite shall return exit code 0 from `go test ./...` upon SPEC completion.
- **REQ-TST-002** (Ubiquitous): The test refactor shall preserve observable production behavior of `cmd/moai` and all retained-agent / retained-hook contracts established by ATR-001.
- **REQ-TST-003** (Ubiquitous): The test refactor shall NOT modify predecessor SPEC artifact bodies (ATR-001 / CATALOG / LNCO-001 / UNP-001 / any pre-existing SPEC).
- **REQ-TST-004** (Ubiquitous): The test refactor shall NOT introduce new test cases beyond fixes to the 15 enumerated baseline failures.

### B.2 Event-driven requirements (When ... shall ...)

- **REQ-TST-005** (Event-driven): When `make build` is invoked during run-phase M2 or M3, the embedded template registry shall regenerate without manual editing of `internal/template/embedded.go`.
- **REQ-TST-006** (Event-driven): When the run-phase reaches M6 verification, the orchestrator shall execute the 7-item read-only verification batch (`go test`, `coverprofile`, `grep` subagent-boundary, sentinel scan, CLI smoke, benchmark optional, lint) in a single parallel turn per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution.
- **REQ-TST-007** (Event-driven): When a test fix touches catalog state (catalog.yaml or embedded equivalents), the fix shall route catalog regeneration through `make build` rather than direct edit.

### B.3 State-driven requirements (While ... shall ...)

- **REQ-TST-008** (State-driven): While run-phase milestones M1–M5 are in progress, the SPEC frontmatter status shall remain `in-progress` and the predecessor SPEC bodies shall remain unmodified.
- **REQ-TST-009** (State-driven): While the test refactor is active on the working tree, the L46 path-specific staging discipline shall apply — each milestone commit shall stage only files within the milestone scope.

### B.4 Where (capability gate) requirements

- **REQ-TST-010** (Where capability): Where the failing test is classified as `pre-existing` in §A.4 (rows 4 and 15), the fix shall annotate the commit message with `(pre-existing)` to preserve the classification trail.
- **REQ-TST-011** (Where capability): Where the failing test is classified as `architectural-pivot consequence`, the fix shall reference the ATR-001 anchor in its commit body so that the discharge trail back to ATR-001 PROCEED-WITH-DEBT directive remains traceable.

### B.5 Unwanted behavior requirements (shall not)

- **REQ-TST-012** (Unwanted): The test refactor shall not skip any failing test via `t.Skip()` or `testing.Short()` to mask a failure rather than fix it.
- **REQ-TST-013** (Unwanted): The test refactor shall not delete any failing test to bypass the failure; if a test is genuinely obsolete (e.g., asserts on an archived agent), the obsolescence shall be justified inline with reference to ATR-001 archive list and replaced with an equivalent retained-catalog assertion.

## Section C — HARD / SHOULD constraints

### C.1 HARD constraints

- **HARD-1**: The refactor is test-only where feasible — production code under `internal/cmd`, `internal/cli`, `internal/skills` (non-test files), `internal/template` (non-test files), `internal/harness` (non-test files), `internal/statusline` (non-test files) shall remain unchanged unless a test's expected-output literal references a stale path/catalog/structure that exists in the test fixture and not in production code. Production code edits, when unavoidable, shall be limited to: (a) catalog.yaml refresh via `make build`, (b) catalog.go generated artifacts via `make build`, (c) embedded.go regenerated artifacts via `make build`. Hand-edits of generated files are prohibited.
- **HARD-2**: Predecessor SPEC bodies (`.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/**`, `.moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/**`, `.moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/**`, and all other pre-existing SPECs) shall not be modified. L48 SSOT preservation is absolute.
- **HARD-3**: Catalog regeneration shall route exclusively through `make build`. Manual edits of `internal/template/embedded.go`, `internal/template/catalog.go`-generated entries, or hash-generated artifacts are prohibited per AC-CHR-008 ORPHAN purge post-condition established by CATALOG-HASH-REGRESSION-CLEANUP-001.
- **HARD-4**: Every milestone commit shall pass pre-commit and post-push `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` returning `0 0` (or document L52 race absorption). L44 38x discipline applies.
- **HARD-5**: Run-phase commit messages shall follow Conventional Commits format: `fix(SPEC-V3R6-TEST-REFACTOR-001): <milestone summary>` for fix milestones, with `🗿 MoAI <email@mo.ai.kr>` footer.

### C.2 SHOULD constraints

- **SHOULD-1**: Run-phase manager-develop should aggregate fixes within a single package into one milestone commit (M1 / M2 / M3 / M4 / M5 each one commit) to keep the commit graph readable. Cross-package fixes that share a root cause may be aggregated with explicit justification.
- **SHOULD-2**: The run-phase should verify ground truth re-measurement (`go test ./... 2>&1 | grep "^--- FAIL"`) at M1 entry and report any drift from the §A.4 enumeration via blocker report to the orchestrator before commit.
- **SHOULD-3**: The `golangci-lint run --timeout=2m` command should return zero issues at run-phase M6 verification. If pre-existing lint issues exist on the baseline, document them as out-of-scope deferred debt rather than fix them in this SPEC.

## Section D — Definition of Done

This SPEC is COMPLETE when all of the following hold simultaneously:

1. All 14 MUST-PASS AC rows (see acceptance.md §A) verified PASS.
2. `go test ./...` returns exit code 0 with zero `FAIL` lines.
3. `golangci-lint run --timeout=2m` returns no NEW lint issues compared to baseline at HEAD `e7b119924`.
4. CHANGELOG.md contains one entry under `[Unreleased]` summarizing the discharge.
5. SPEC frontmatter status transitions: `draft → in-progress` (run-phase entry) → `implemented` (sync-phase) → `completed` (Mx-phase, optional).
6. All milestone commits authored against the L46 path-specific staging discipline (each commit's `git diff --cached --name-only` matches the milestone scope).
7. Mx-phase Step C executes (EVALUATE per coverage delta) and emits the 4-phase close marker.
8. Predecessor SPEC bodies remain unmodified (verified via `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/` returns empty after this SPEC's full 4-phase close).

## Section E — Exclusions (OUT OF SCOPE)

The following are explicitly OUT OF SCOPE of this SPEC. Attempts to expand the scope into these areas shall be rejected at plan-auditor verdict or run-phase entry:

### E.1 No NEW test additions

This SPEC fixes the 15 enumerated baseline failures only. NEW test cases (e.g., additional coverage for under-tested packages, regression tests for newly discovered classes of bugs, integration tests for newly added features) are out of scope. If under-testing is discovered, a follow-up SPEC `SPEC-V3R6-TEST-COVERAGE-EXPANSION-XXX` shall be authored separately.

### E.2 No production behavior changes

This SPEC does not change the observable behavior of `moai cli`, retained agents, retained hooks, retained skills, or any user-facing surface. Production code edits are limited to `make build`-driven catalog regeneration per HARD-3.

### E.3 No architectural redesign

This SPEC does not revisit the ATR-001 17→8 agent catalog consolidation, the Anthropic 2026 alignment findings A1–A6, the archived-agent rejection rule, or any architectural decision finalized by ATR-001. The pivot is settled SSOT.

### E.4 No GEARS-EARS migration work

This SPEC does not address residual EARS legacy modality in the 88 pre-v3 SPECs. The GEARS migration backward-compatibility window (6 months from v3.0.0 release per SPEC-V3R6-GEARS-MIGRATION-001 v0.2.0) remains active and is tracked separately.

### E.5 No catalog-hash regression cleanup beyond test-fix scope

This SPEC does not re-author the catalog drift detection mechanism. The mechanism authored by CATALOG-HASH-REGRESSION-CLEANUP-001 is treated as a verified guardrail; tests that fail because the catalog drifted are fixed by routing catalog regen through `make build`, not by altering the drift-detection mechanism.

## HISTORY

### v0.1.1 (2026-05-25) — run-phase M1 frontmatter status:in-progress

- Run-phase M1 entry: frontmatter status transition `draft → in-progress` per Status Transition Ownership Matrix exception (manager-develop allowed on draft → in-progress only).
- Ground truth re-measurement at HEAD `40dc43f5b` confirms 15 failures matching §A.4 baseline exactly.

### v0.1.0 (2026-05-25) — initial draft

- Plan-phase 5-artifact set authored to discharge ATR-001 PROCEED-WITH-DEBT directive (8 reported failures at M8, 15 measured at this plan-phase entry per +7 drift across 2 sync + 2 Mx cohort commits).
- 13 GEARS requirements (4 Ubiquitous + 3 Event-driven + 2 State-driven + 2 Where-capability + 2 Unwanted), zero IF/THEN modality.
- 14 MUST-PASS AC + 100% bidirectional REQ↔AC traceability.
- 5 HARD constraints + 3 SHOULD constraints + 5 explicit OOS sections.
- Tier M classification per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (300–1000 LOC test refactor, 5–15 file scope, M1–M6 milestones, Hybrid Trunk main-direct default).
- L48 SSOT preservation: ATR-001 / CATALOG / LNCO-001 bodies absolutely preserved.
