---
id: SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001
title: skills_audit_test.go solo run.md path update — TEMPLATE-MIRROR-DRIFT fix for SPEC-V3R4-WORKFLOW-SPLIT-001 cascade
version: "0.1.0"
status: completed
created: 2026-05-24
updated: 2026-06-02
author: manager-spec
priority: P3
phase: v3.0.0
module: internal/template
lifecycle: spec-anchored
tags: "template-mirror-drift, test-fix, workflow-split"
issue_number: null
depends_on: []
---

# SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001 — skills_audit_test.go solo run.md path update

## A. Context

### A.1 Trigger

Pre-existing baseline failure in `internal/template/skills_audit_test.go:96`:

```
$ go test ./internal/template/ -run TestSkillsContainPlanAuditGateMarkers -v
--- FAIL: TestSkillsContainPlanAuditGateMarkers (0.00s)
    --- FAIL: TestSkillsContainPlanAuditGateMarkers/solo_run.md_—_plan_audit_gate_markers (0.00s)
        skills_audit_test.go:96: file ".claude/skills/moai/workflows/run.md" missing required pattern: "Phase 0.5: Plan Audit Gate"
        skills_audit_test.go:96: file ".claude/skills/moai/workflows/run.md" missing required pattern: "plan-auditor"
        skills_audit_test.go:96: file ".claude/skills/moai/workflows/run.md" missing required pattern: "--skip-audit"
        skills_audit_test.go:96: file ".claude/skills/moai/workflows/run.md" missing required pattern: "INCONCLUSIVE"
        skills_audit_test.go:96: file ".claude/skills/moai/workflows/run.md" missing required pattern: ".moai/reports/plan-audit/"
    --- PASS: TestSkillsContainPlanAuditGateMarkers/plan.md_—_audit-ready_signal
    --- PASS: TestSkillsContainPlanAuditGateMarkers/spec-workflow.md_—_Phase_0.5_documentation
    --- PASS: TestSkillsContainPlanAuditGateMarkers/team_run.md_—_plan_audit_gate_markers
```

This is the LCL-001 AC-LCL-005-adjacent residual (LCL-001 cleared the i18n-validator budget side; this is the unrelated TEMPLATE-MIRROR-DRIFT-001-family case discovered during the same Sprint 2 sweep).

### A.2 Root Cause

SPEC-V3R4-WORKFLOW-SPLIT-001 (commit `986418598`, "Wave 1 — run.md phase-scoped sub-skill split (4 sub-skills + thin router)") split the monolithic `internal/template/templates/.claude/skills/moai/workflows/run.md` (398+ lines) into a thin router (105 lines) by relocating Phase 0.5 documentation into sub-skill `internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md` (449 lines).

Verification (current main `d3ed4727d`):

| Pattern | run.md (router) | run/phase-execution.md (sub-skill) |
|---------|-----------------|------------------------------------|
| `Phase 0.5: Plan Audit Gate` | 0 | 1 |
| `plan-auditor` | 0 | 8 |
| `--skip-audit` | 0 | 3 |
| `INCONCLUSIVE` | 0 | 13 |
| `.moai/reports/plan-audit/` | 0 | 7 |

The test `TestSkillsContainPlanAuditGateMarkers` (added by SPEC-WF-AUDIT-GATE-001 T-06) was NOT updated as part of the WORKFLOW-SPLIT-001 split — it still asserts the 5 patterns against `.claude/skills/moai/workflows/run.md` (now the router) when the patterns were intentionally moved to `workflows/run/phase-execution.md` (the sub-skill).

### A.3 Attribution Correction

The prior session's L46 attribution memo cited commits `8ec11dee7` / `3a06f82e7` / `0d7debf19` (CODE-COMMENTS-EN-001 / HARNESS-RENAME-001 / CORE-SLIM-B-001) as TEMPLATE-MIRROR-DRIFT culprits. Orchestrator independent verification via `git show --stat <commit>` confirms NONE of those three commits touched the 4 target template files (`run.md` / `run/phase-execution.md` / `team/run.md` / `plan.md` / `spec-workflow.md`). The actual culprit is commit `986418598` (SPEC-V3R4-WORKFLOW-SPLIT-001 Wave 1). This SPEC supersedes the L46 attribution for this specific test failure case.

### A.4 Scope Decision

Tier S minimal scope: single test file update (2-4 lines), zero production code changes, zero template content changes. The WORKFLOW-SPLIT-001 production split is correct as-is — only the test assertion needs to track the new sub-skill location.

### A.5 Tier Justification

Tier S per `.claude/rules/moai/development/spec-frontmatter-schema.md`:

- Single file edit footprint: 2-4 lines in `internal/template/skills_audit_test.go`
- Zero production code edits
- Zero template/skill content edits
- 1 milestone (M1 = test path update + verification)
- Within ≤300 LOC envelope (actual: 2-4 lines)
- No cross-package coupling

### A.6 Out of Scope

The following are explicitly NOT in scope for this SPEC:

- **Production template files unchanged**: `internal/template/templates/.claude/skills/moai/workflows/run.md` (router, 105 lines) and `internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md` (sub-skill, 449 lines) MUST remain byte-identical. The WORKFLOW-SPLIT-001 architectural split is canonical.
- **Other test entries unchanged**: `team/run.md` sub-test (~lines 49-60), `plan.md` sub-test (~lines 61-71), `spec-workflow.md` sub-test (~lines 72-82), and `TestReportsDirGitkeepExists` (lines 109-128) MUST remain byte-identical.
- **Other workflow split SPECs not retroactively audited**: This SPEC fixes only the specific failure shown in A.1. Other potential test drift cases from other SPEC splits are not in scope (deferred to potential follow-up TEMPLATE-MIRROR-DRIFT-001 if discovered).
- **Skill content semantics not validated**: The test asserts only string presence, not semantic correctness of the Phase 0.5 documentation in the sub-skill.

## B. Requirements (EARS Format)

### B.1 REQ-SARM-001 — solo run.md test entry path update (Event-Driven)

**WHEN** the test suite `TestSkillsContainPlanAuditGateMarkers` is executed, **THEN** the `solo run.md — plan audit gate markers` sub-test SHALL load and verify patterns against `.claude/skills/moai/workflows/run/phase-execution.md` (the sub-skill containing Phase 0.5 documentation post-SPEC-V3R4-WORKFLOW-SPLIT-001), NOT `.claude/skills/moai/workflows/run.md` (the thin router).

### B.2 REQ-SARM-002 — All 5 required patterns SHALL pass (Ubiquitous)

The updated test entry SHALL assert presence of all 5 patterns (`Phase 0.5: Plan Audit Gate`, `plan-auditor`, `--skip-audit`, `INCONCLUSIVE`, `.moai/reports/plan-audit/`) without modification — the patterns themselves are correct; only the target file path changed.

### B.3 REQ-SARM-003 — Other test entries SHALL remain byte-identical (Unwanted)

The system SHALL NOT modify any other test entry in `skills_audit_test.go`. Specifically: `team run.md` (~lines 49-60), `plan.md` (~lines 61-71), `spec-workflow.md` (~lines 72-82) sub-tests, and `TestReportsDirGitkeepExists` test (lines 109-128) MUST remain unchanged.

### B.4 REQ-SARM-004 — Production templates SHALL remain unchanged (Unwanted)

The system SHALL NOT modify `internal/template/templates/.claude/skills/moai/workflows/run.md` (router) or `internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md` (sub-skill). The WORKFLOW-SPLIT-001 split is canonical.

### B.5 REQ-SARM-005 — Sub-test display name SHALL reflect new path (Event-Driven)

**WHEN** the test sub-test name is rendered by `t.Run(tt.name, ...)`, **THEN** the name SHALL reflect the new sub-skill location (e.g., `solo run/phase-execution.md — plan audit gate markers`) so that test output is unambiguous about which file is being verified.

### B.6 REQ-SARM-006 — Baseline quality SHALL be preserved (Ubiquitous)

The `go vet ./internal/template/...` and `golangci-lint run` baselines SHALL remain at 0 findings post-edit. No new lint or vet issues introduced.

### B.7 REQ-SARM-007 — Comment block SHALL document split context (Optional)

**WHERE** code documentation aids future maintainers, the system MAY update the comment block at lines 37-38 to reference the post-SPEC-V3R4-WORKFLOW-SPLIT-001 sub-skill path migration context. This is optional and aesthetic — primary correctness is in the `filePath` field.

## C. Decision Rule

Per spec.md SSOT (this section): The test entry `filePath` field is the canonical assertion target. `name` field SHOULD reflect the path for diagnostic clarity but is not assertion-critical. plan.md HOW-derivation must conform to this REQ rule.

## D. Validation Constraints

- **D.1 Tier S 3-artifact + progress requirement**: spec.md + plan.md + acceptance.md + progress.md (4 artifacts).
- **D.2 Frontmatter canonical schema**: 12 required fields per `.claude/rules/moai/development/spec-frontmatter-schema.md` — `created:`/`updated:`/`tags:` (NOT snake_case `created_at:`/`updated_at:`/`labels:`).
- **D.3 plan-auditor threshold**: Tier S = **0.75** minimum score (canonical SSOT — `.claude/agents/meta/plan-auditor.md` § Tier-differentiated PASS threshold + `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier; corrected from earlier 0.70 typo per plan-auditor iter-1 B-3 finding).
- **D.4 spec-lint compliance**: Two-part requirement — (a) `### A.6 Out of Scope` h3 sub-section present (above) to satisfy `MissingExclusions` rule on **spec.md only** (canonical artifact); (b) frontmatter `tags:` MUST use **CSV-string form** (e.g., `tags: "a, b, c"`) NOT YAML array form (`tags: [a, b, c]`) — array form causes `ParseFailure` in `internal/spec/lint.go` `SPECFrontmatter.Tags string yaml:"tags"` binding. Post-fix verified by `go run ./cmd/moai spec lint .moai/specs/SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001/spec.md` exit 0. **Note**: plan.md/acceptance.md/progress.md as derived artifacts return `MissingExclusions` ERROR when lint-checked individually (same as comparator SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 which merged 2026-05-24); CI does not block on derived-artifact lint per `git-strategy.yaml` pattern. The canonical lint surface is spec.md only.
- **D.5 [HARD] Working tree hygiene**: Do NOT touch dirty/untracked files (`.moai/harness/usage-log.jsonl`, `.moai/research/v3.0-redesign-2026-05-23.md`, `.moai/harness/observations.yaml`, `.moai/config/sections/{git-convention,language,quality}.yaml`).
