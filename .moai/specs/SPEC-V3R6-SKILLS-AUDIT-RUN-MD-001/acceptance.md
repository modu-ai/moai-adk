---
id: SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001
title: Acceptance Criteria — skills_audit_test.go solo run.md path update
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
phase: v3.0.0
module: internal/template
lifecycle: spec-anchored
tags: "template-mirror-drift, test-fix, workflow-split"
---

# Acceptance Criteria — SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001

## SSOT for Acceptance Verification

5 ACs total. Each is binary PASS/FAIL with explicit verification command.

## AC Matrix

| AC ID | REQs Covered | Requirement | Verification Command | PASS Criterion |
|-------|--------------|-------------|----------------------|----------------|
| **AC-SARM-001** | REQ-SARM-001, REQ-SARM-002, REQ-SARM-005 | `TestSkillsContainPlanAuditGateMarkers` 4/4 sub-tests PASS | `go test ./internal/template/ -run "TestSkillsContainPlanAuditGateMarkers" -v` | All 4 sub-tests report `--- PASS`; no `--- FAIL` lines; exit code 0 |
| **AC-SARM-002** | REQ-SARM-003 | `TestReportsDirGitkeepExists` no regression | `go test ./internal/template/ -run "TestReportsDirGitkeepExists" -v` | `--- PASS: TestReportsDirGitkeepExists`; exit code 0 |
| **AC-SARM-003** | REQ-SARM-001, REQ-SARM-002, REQ-SARM-003, REQ-SARM-006 | Full `internal/template/...` package suite PASS | `go test ./internal/template/... -count=1` | Exit code 0; output ends with `ok\tgithub.com/modu-ai/moai-adk/internal/template\t<duration>` |
| **AC-SARM-004** | REQ-SARM-003, REQ-SARM-004 | Unrelated test entries byte-identical (team/run.md, plan.md, spec-workflow.md sub-tests + TestReportsDirGitkeepExists); production templates untouched | `git diff HEAD internal/template/skills_audit_test.go \| grep -E "^[+-]" \| grep -v "^[+-]{3}"` AND `git diff HEAD internal/template/templates/.claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md` | First diff contains ONLY changes to `solo run.md` test entry region (~lines 36-48); second diff empty (zero template content changes) |
| **AC-SARM-005** | REQ-SARM-006 | `go vet` + `golangci-lint` baseline preserved (0 new findings) | `go vet ./internal/template/... && golangci-lint run --timeout=2m ./internal/template/...` | Both commands exit 0; no new warnings vs pre-edit baseline (0 findings) |

> **REQ-SARM-007 (Optional)** — intentionally has no AC coverage. The Optional MAY-clause covers an aesthetic comment-block update at `skills_audit_test.go` lines 37-38; not assertion-critical. Decision rationale: per spec.md §C, `filePath` is the canonical assertion target; the comment block is documentation drift, not test correctness.

## Given-When-Then Scenarios

### Scenario 1: Solo run.md sub-test passes against new path

**Given** the test file `internal/template/skills_audit_test.go` has been updated per plan.md M1 (line 39 name + line 40 filePath edited),

**When** the developer runs `go test ./internal/template/ -run "TestSkillsContainPlanAuditGateMarkers/solo" -v`,

**Then** the output shows:
```
--- PASS: TestSkillsContainPlanAuditGateMarkers/solo_run/phase-execution.md_—_plan_audit_gate_markers
```
and exit code is 0.

### Scenario 2: Other sub-tests unchanged

**Given** the same edit applied,

**When** the developer runs `go test ./internal/template/ -run "TestSkillsContainPlanAuditGateMarkers" -v`,

**Then** the output shows 4 `--- PASS` lines:
```
--- PASS: TestSkillsContainPlanAuditGateMarkers/solo_run/phase-execution.md_—_plan_audit_gate_markers
--- PASS: TestSkillsContainPlanAuditGateMarkers/team_run.md_—_plan_audit_gate_markers
--- PASS: TestSkillsContainPlanAuditGateMarkers/plan.md_—_audit-ready_signal
--- PASS: TestSkillsContainPlanAuditGateMarkers/spec-workflow.md_—_Phase_0.5_documentation
```

### Scenario 3: Production templates untouched

**Given** the edit is complete,

**When** the developer runs `git diff HEAD internal/template/templates/.claude/skills/moai/workflows/run.md internal/template/templates/.claude/skills/moai/workflows/run/phase-execution.md`,

**Then** the diff is empty (zero changes to production template content).

### Scenario 4: Baseline quality preserved

**Given** the edit is complete,

**When** the developer runs `go vet ./internal/template/...` followed by `golangci-lint run --timeout=2m ./internal/template/...`,

**Then** both commands exit 0 with no new warnings vs the pre-edit baseline.

## Edge Cases

| Edge Case | Expected Behavior |
|-----------|-------------------|
| Sub-skill `phase-execution.md` deleted in concurrent SPEC | `TestSkillsContainPlanAuditGateMarkers/solo` would FAIL with `ReadFile` error — surfaces drift immediately (desired fail-loud behavior). |
| Required pattern (e.g., `INCONCLUSIVE`) renamed in sub-skill | Test FAILs with `missing required pattern` error — surfaces semantic drift (desired). |
| New SPEC adds 5th sub-skill that should also be tested | Out of scope for this SPEC; follow-up SPEC would add new test entry. |
| Go subtest path matching with `/` separator | Sub-test name `"solo run/phase-execution.md — plan audit gate markers"` produces rendered name `TestSkillsContainPlanAuditGateMarkers/solo_run/phase-execution.md_—_plan_audit_gate_markers` (Go converts spaces to underscores, preserves `/` as nested-subtest path segment). `-run "TestSkillsContainPlanAuditGateMarkers/solo"` filter MATCHES via Go's prefix-on-slash-segment semantics (`/solo` matches `/solo_run/...` because slash separates segments and `solo_run` starts with `solo`). Verified by run-phase Scenario 1 + Scenario 2 PASS outputs. |

## Definition of Done (DoD)

- [ ] AC-SARM-001 PASS verified
- [ ] AC-SARM-002 PASS verified
- [ ] AC-SARM-003 PASS verified
- [ ] AC-SARM-004 PASS verified
- [ ] AC-SARM-005 PASS verified
- [ ] 4 SPEC artifacts (spec.md/plan.md/acceptance.md/progress.md) frontmatter status `draft → implemented` (sync phase)
- [ ] CHANGELOG `[Unreleased]` `### Fixed` entry added (sync phase)
- [ ] `progress.md` §Sync-phase Evidence row populated with B12 self-test PASS evidence

## Quality Gate Thresholds

| Gate | Threshold | This SPEC |
|------|-----------|-----------|
| plan-auditor score | ≥ 0.75 (Tier S, canonical SSOT — corrected from 0.70 typo per iter-1 B-3) | Target |
| BLOCKING findings | 0 | Target |
| Code coverage | Baseline preserved (no new code) | Target |
| Lint/vet findings | 0 new | Target |
| Test pass rate | 100% (5/5 ACs) | Target |
