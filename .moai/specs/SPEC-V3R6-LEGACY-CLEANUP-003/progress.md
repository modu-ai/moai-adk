---
id: SPEC-V3R6-LEGACY-CLEANUP-003
title: "SPEC-V3R6-LEGACY-CLEANUP-003 — Progress"
version: "0.2.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-develop
priority: P3
tags: "cleanup, legacy, terminology, sprint-2, progress"
issue_number: null
tier: M
phase: "v3.0.0"
module: "internal/runtime"
lifecycle: spec-anchored
---

# SPEC-V3R6-LEGACY-CLEANUP-003 — Progress

## Status Matrix

| Phase | Status | Commit | Evidence |
|-------|--------|--------|----------|
| Pre-flight | DONE | 146a268e7 (plan-ready baseline) | 28 Wave hits / 16 files / 10 budget_test.go literals verified |
| Plan | DONE | 146a268e7 | manager-spec Tier M FULL Section A-E (4 artifacts: spec.md + plan.md + acceptance.md + progress.md) |
| Plan-audit | DONE (iter-2 PASS 0.87) | 146a268e7 | 5 BLOCKING resolved + MP-2 PASS via orchestrator-direct fix-forward; Tier M threshold 0.80 +0.07 margin |
| Run M1 — Comment renames (12 files) | PASS | 80414b8f5 | 17 lines renamed; 3 §a/§b exemptions preserved; go build + vet + test PASS |
| Run M2 — PersistProgress API rename | PASS | e432fc276 | 5 edits in persist.go; 0 wave residuals; 6 round occurrences |
| Run M3 — ResumeMessageFormat + DefaultFallback | PASS | afb4957f1 | 2 edits in config.go; production wave residuals 0 (exc. budget_test.go) |
| Run M4 — Test file alignment + verification | PASS | 9b18493fb | 10 edits in budget_test.go; runtime tests PASS; coverage 89.5% |
| Sync | TBD | TBD | TBD (populated by manager-docs `/moai sync`) |
| Mx Step C | SKIP justified | n/a | Comment-only renames; no new @MX:ANCHOR/WARN triggers (fan_in unchanged); @MX:SPEC unchanged. Sync phase may emit `@MX:NOTE [AUTO]` for budget.go fallback string history if desired. |

## Acceptance Criteria

| AC | REQ | Status | Evidence |
|----|-----|--------|----------|
| AC-LCL-001 | REQ-LCL-001 | PASS | `grep -rnE "\b[Ww]ave\b" <14 files>` excluding `SPEC-V3R3-CI-AUTONOMY-001 Wave 5` + `strategy-wave5` exemptions → 0 lines |
| AC-LCL-002 | REQ-LCL-002 | PASS | `grep "waveLabel|wave_label|- Wave:" persist.go` → 0 lines; `grep "roundLabel|round_label|- Round:" persist.go` → 6 occurrences |
| AC-LCL-003 | REQ-LCL-003 | PASS | `grep "{wave_label}" config.go` → 0 lines; line 136 default value uses `{round_label}` |
| AC-LCL-004 | REQ-LCL-004 | PASS | `grep -rn "split_into_waves\|smaller waves" internal/runtime/` excluding budget_test.go (M4 scope) → 0 lines; DefaultFallback = "split_into_rounds" |
| AC-LCL-005 | REQ-LCL-005 | PASS | `grep '"Wave [0-9]"\|split_into_waves\|{wave_label}' budget_test.go` → 0 lines; 1:1 swap to Round/split_into_rounds/{round_label} (10 occurrences) |
| AC-LCL-006 | REQ-LCL-006 | PASS (with documented pre-existing baseline) | `go test ./internal/runtime/...` → ok 1.927s coverage 89.5%; full `go test ./...` shows pre-existing `internal/template/` failures (TestBackwardCompatibility / TestAgentFrontmatterAudit / TestEmbeddedTemplates_AgentDefinitions / TestLateBranchTemplateMirror / TestAllAgentsInCatalog / TestLoadCatalog / TestRuleTemplateMirrorDrift / TestLoadEmbeddedCatalog_Success / TestSkillsContainPlanAuditGateMarkers / TestRetirementCompletenessAssertion) attributable to sibling SPECs CODE-COMMENTS-EN-001 8ec11dee7 + HARNESS-RENAME-001 3a06f82e7 + CORE-SLIM-B-001 0d7debf19 per L46 attribution rule (stash-and-rerun verified identical baseline failures pre-M4). Zero NEW failures introduced by this SPEC. |
| AC-LCL-007 | REQ-LCL-006 + LSP gate | PASS | `go vet ./...` → no output (exit 0); `golangci-lint run --timeout=2m` → 0 issues (identical to pre-flight baseline) |
| AC-LCL-008 | REQ-LCL-007 + REQ-LCL-008 | PASS | (1) `handle-harness-observe` references in internal/ Go files: 24 (≥9 baseline preserved); (2) `internal/cli/migrate_agency*` cluster untouched (B10 scope discipline); (3) `internal/config/types.go` Copywriter/Designer fields untouched; (4) state_guard.go:15 + doc.go:11 historical SPEC-IDs verbatim; (5) state_guard.go:39 strategy-wave5.md file reference verbatim |

## Sync-phase Evidence (TBD — populated by manager-docs)

TBD

## Mx-phase Evidence

Skip justified for `/moai mx` Step C per §D-3 reasoning:
- All edits are mechanical Wave→Round renames in comments + 1 API parameter + 1 default string + test fixtures
- No new functions added → no new @MX:ANCHOR candidates (fan_in unchanged)
- No new goroutines / complexity ≥15 / state mutation → no new @MX:WARN candidates
- No new business rules surfaced → no new @MX:NOTE candidates from this SPEC
- No public functions left untested → no new @MX:TODO candidates
- Existing @MX tags in state_guard.go (@MX:ANCHOR + @MX:REASON + @MX:SPEC) preserved verbatim per §a exemption

Optional follow-up (out of scope, may be batched with LCL-002 follow-up sweep): add `@MX:NOTE [AUTO]` to budget.go line 167 explaining the `split_into_rounds` value evolution from historical `split_into_waves`. Recommendation only; not blocking.
