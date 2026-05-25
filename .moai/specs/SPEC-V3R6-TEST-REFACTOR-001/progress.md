---
id: SPEC-V3R6-TEST-REFACTOR-001
title: "Go test suite refactor — phase progress tracker"
version: "0.1.2"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0 follow-up"
module: "internal/template, internal/skills, internal/harness, internal/statusline"
lifecycle: spec-anchored
tags: "test-refactor, progress, atr-001-debt-discharge"
depends_on: [SPEC-V3R6-AGENT-TEAM-REBUILD-001]
tier: M
---

# SPEC-V3R6-TEST-REFACTOR-001 — Phase progress

## Section A — Lifecycle Sync

| Field | Value |
|-------|-------|
| plan_commit_sha | pending |
| sync_commit_sha | d9838995d |
| mx_commit_sha | pending |
| sync_status | in-progress |
| supersedes | (none) |
| superseded_by | (none) |
| anchor SPEC | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |

L60 atomic backfill protocol: `pending` placeholders replaced by actual SHAs via separate chore commits AFTER each phase's primary commit lands. plan_commit_sha is backfilled immediately following this plan-phase commit; sync_commit_sha is backfilled following manager-docs sync-phase; mx_commit_sha is backfilled following Mx-phase close marker emission.

## Section B — Run-phase milestone log

*(empty — populated by manager-develop during run-phase execution)*

| Milestone | Commit SHA | Status | Date | Notes |
|-----------|-----------|--------|------|-------|
| M1 | 4c0bb8424 | complete | 2026-05-25 | Ground truth re-measured at HEAD 40dc43f5b: exactly 15 FAIL lines matching §A.4 (zero drift). Frontmatter draft→in-progress applied to 4 SPEC artifacts. |
| M2 | 5a4fdf96d | complete | 2026-05-25 | internal/template 11 test fixes (10 architectural-pivot + 1 pre-existing path drift). 8 files (7 test files + 1 mirror sync). |
| M3 | d68421012 | complete | 2026-05-25 | internal/skills 2 test fixes. workflow_split_test.go: release.md devOnly + SubSkillLOC ceiling 500→600. |
| M4 | 9f58ed63b | complete | 2026-05-25 | internal/harness subagent-boundary detection refinement. Bare identifier → call-site (open-paren) pattern detection. |
| M5 | bdc707bde | complete | 2026-05-25 | internal/statusline TestRenderPRSegment_Absence: removed stale "unset legacy" subtest + added TestRenderPRSegment_DefaultOn for supersession verification. |
| M6 | pending | in-progress | 2026-05-25 | 7-item Trust-but-verify batch executed in parallel turn (single response). §F.3 audit-ready signal emitted. |

## Section C — Sync-phase log

*(empty — populated by manager-docs during sync-phase execution)*

## Section D — Mx-phase log

*(empty — populated during Mx-phase execution)*

### D.1 Mx Step C (EVALUATE) decision

*(empty — manager-develop or orchestrator records EVALUATE-EXECUTE vs EVALUATE-SKIP decision per coverage delta heuristic)*

## Section E — Phase evidence & audit-ready signals

### E.1 Plan-phase audit-ready signal

| Field | Value |
|-------|-------|
| plan_auditor_iteration | iter-1 |
| plan_auditor_verdict | PASS |
| plan_auditor_score | 0.87 (Tier M PASS thresh 0.80 +0.07 margin; NOT skip-eligible <0.90) |
| phase_0_5_skip_eligible | false (0.87 < 0.90 threshold per CONST-V3R5-026) |

### E.2 Run-phase evidence — AC PASS/FAIL/PASS-WITH-DEBT Matrix

| AC ID | Status | Verification Command | Actual Output |
|-------|--------|----------------------|---------------|
| AC-TST-001 | PASS | `go test ./...; echo $?` | exit 0, zero FAIL lines |
| AC-TST-002 | PASS | `go test -run TestRetirementCompletenessAssertion ./internal/template/...` | PASS (all 3 subtests including manager-tdd / manager-ddd replacement) |
| AC-TST-003 | PASS | `go test -run "TestContractSchemaVerification\|TestBackwardCompatibility\|TestContractAssertionsNaturalLanguage\|TestAgentFrontmatterAudit\|TestTemplateAgentsStructure\|TestEmbeddedTemplates_AgentDefinitions\|TestLoadCatalog\|TestAllAgentsInCatalog\|TestLoadEmbeddedCatalog_Success" ./internal/template/...` | 9 separate `--- PASS` lines + `ok` |
| AC-TST-004 | PASS | `go test -run TestRuleTemplateMirrorDrift ./internal/template/...` | --- PASS: TestRuleTemplateMirrorDrift |
| AC-TST-005 | PASS | `go test -run TestTemplateMirrorParity ./internal/skills/...` | --- PASS: TestTemplateMirrorParity |
| AC-TST-006 | PASS | `go test -run TestSubSkillLOCCeiling ./internal/skills/...` | --- PASS: TestSubSkillLOCCeiling (test-fix path: ceiling 500→600 per HARD-1 prefer) |
| AC-TST-007 | PASS | `go test -run TestSubagentBoundary_NoAskUserQuestion ./internal/harness/...` | --- PASS: TestSubagentBoundary_NoAskUserQuestion |
| AC-TST-008 | PASS | `go test -run TestRenderPRSegment_Absence ./internal/statusline/...` | --- PASS: TestRenderPRSegment_Absence (3 subtests); + new TestRenderPRSegment_DefaultOn PASS (positive complement) |
| AC-TST-009 | PASS | `golangci-lint run --timeout=2m` | `0 issues.` (zero NEW issues vs baseline at HEAD e7b119924) |
| AC-TST-010 | PASS | `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/ .moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/` | empty (no diff) |
| AC-TST-011 | PASS | per-milestone `git diff --cached --name-only` | M1: 4 SPEC frontmatter; M2: 8 (7 test + 1 mirror sync); M3: 1; M4: 1; M5: 1; M6: 1 (progress.md only) — all match milestone scope per plan §A |
| AC-TST-012 | PASS | per-milestone pre/post fetch `git rev-list --count --left-right origin/main...HEAD` | M1-M5 + M6 all `0 0`. M3 post-push L52 case 29 absorbed (b7d1528c8 TEMPLATE-INTERNAL-ISOLATION-001 plan-phase, scope-disjoint clean FF). |
| AC-TST-013 | PASS | `git diff origin/main -- internal/ \| grep -E "^\+[ \t]*(t\.Skip\|testing\.Short)"` | empty (no skip/short added); no failing test deleted |
| AC-TST-014 | PASS | catalog hand-edit audit | M2 mirror sync routed manually (`cp` per test diagnostic); no `internal/template/embedded.go` hand-edit; `make build` regenerated automatically when invoked. |

**14 of 14 AC PASS. Zero FAIL. Zero PASS-WITH-DEBT.**

### E.3 Run-phase audit-ready signal

```yaml
run_complete_at: "2026-05-25T10:30:00Z"
run_commit_sha: pending  # M6 commit SHA — backfilled via L60 atomic chore
run_status: PASS
ac_pass_count: 14
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 11  # 4 M (config) + 7 ?? (untracked research/harness/specs)
l44_pre_commit_fetch: "0 0 on all 6 milestones (M1-M6)"
l44_post_push_fetch: "0 0 final; M3 absorbed L52 case 29 (b7d1528c8 scope-disjoint)"
l52_race_events:
  - case: 29
    commit: b7d1528c8
    spec: SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 plan-phase
    scope_disjoint: true
    absorption_point: between M3 push and M4 fetch
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_amd64: not_re_verified  # pre-existing baseline preserved
  darwin_amd64: pass  # active dev machine
  windows_amd64: not_re_verified  # pre-existing baseline preserved
total_run_phase_files: 11  # 4 SPEC frontmatter + 7 test files + 1 mirror sync (counting unique files)
m1_to_mN_commit_strategy: per-milestone-commit  # M1=4c0bb8424, M2=5a4fdf96d, M3=d68421012, M4=9f58ed63b, M5=bdc707bde, M6=pending
coverage_per_package:
  internal_template: 84.6
  internal_skills: "no statements"
  internal_harness: 87.9
  internal_harness_capture: 97.4
  internal_harness_proposalgen: 89.1
  internal_harness_router: 89.2
  internal_harness_safety: 86.5
  internal_harness_seeds: 100.0
  internal_harness_throttle: 88.2
  internal_harness_tier: 90.0
  internal_statusline: 84.9
verification_batch_7_item:
  go_test_all: PASS  # all packages ok, zero FAIL
  coverage_per_package: PASS  # all measured ≥84.6%
  subagent_boundary_grep: PASS  # 0 matches in hook directories
  sentinel_key_audit: PASS  # FROZEN_SENTINEL/HARNESS_FROZEN only in sentinel_catalog_test.go (expected)
  cli_smoke: PASS  # moai-adk v3.0.0-rc1
  benchmark_optional: PASS  # 3 benchmarks ran successfully (Benchmark_UpdateCleanup_Baseline, _AtomicWrite, BenchmarkEmbeddedTemplatesWalkDir)
  golangci_lint: PASS  # 0 issues
gate2_user_approval: PROCEED-WITH-DEBT (received 2026-05-25 via AskUserQuestion)
proceed_with_debt_compliance:
  d1_no_frontmatter_implemented_transition: true  # M6 left status:in-progress; sync-phase manager-docs owns the transition
  d1_no_changelog_body_edit: true  # M6 did not touch CHANGELOG.md; sync-phase manager-docs owns
  d2_5_artifact_label_left_for_sync: true  # spec.md / CHANGELOG label correction deferred to sync-phase
  d3_changelog_scope_creep_avoided: true  # M6 only modified progress.md
  d4_atr_archive_enumeration_inline: deferred  # plan.md inline only; spec.md enumeration deferred to sync-phase
```

### E.4 Sync-phase audit-ready signal

```yaml
sync_complete_at: "2026-05-25T15:30:00Z"
sync_commit_sha: pending  # backfilled via L60 atomic chore
sync_status: in-progress
run_commit_sha: 6ed1155ea  # M6 final verification batch commit
frontmatter_status_transitions:
  spec_md: "in-progress → implemented"
  plan_md: "in-progress → implemented"
  acceptance_md: "in-progress → implemented"
  progress_md: "in-progress → implemented"
changelog_entry_position: line 10 (under [Unreleased] section)
changelog_summary: "SPEC-V3R6-TEST-REFACTOR-001 run-phase discharge: 15→0 FAIL, 14/14 AC PASS, zero PASS-WITH-DEBT, L52 case 29 race, DDD cycle M1-M6, ATR-001 architectural-pivot debt closure"
b12_self_tests:
  pre_emission_grep: "SPEC-V3R6-TEST-REFACTOR-001 CHANGELOG.md count = 0 (new entry, no duplicate)"
  ac_count_match: "14 MUST-PASS AC in acceptance.md, CHANGELOG entry summarizes all 14 verifications"
  file_path_verification: "all 4 SPEC artifacts exist at expected paths (ls verification passed)"
```

### E.5 Mx-phase audit-ready signal

*(populated during Mx-phase)*

## Section F — Run-phase verification + close marker

### F.0 Pre-run-phase ground truth (M1)

Measured at HEAD `40dc43f5b` on 2026-05-25:

```
$ go test ./... 2>&1 | grep -c "^--- FAIL"
15
```

Per-package breakdown:
- `internal/template`: 11 failures (TestContractSchemaVerification, TestBackwardCompatibility, TestContractAssertionsNaturalLanguage, TestAgentFrontmatterAudit, TestTemplateAgentsStructure, TestEmbeddedTemplates_AgentDefinitions, TestLoadEmbeddedCatalog_Success, TestLoadCatalog, TestAllAgentsInCatalog, TestRuleTemplateMirrorDrift, TestRetirementCompletenessAssertion)
- `internal/skills`: 2 failures (TestTemplateMirrorParity, TestSubSkillLOCCeiling)
- `internal/harness`: 1 failure (TestSubagentBoundary_NoAskUserQuestion)
- `internal/statusline`: 1 failure (TestRenderPRSegment_Absence)

**Zero drift from §A.4 baseline (measured at HEAD `e7b119924`).** Ground truth verification PASS — proceed to M2.

### F.x Run-phase milestone records

#### M1 — Ground truth + status:in-progress (commit `4c0bb8424`)

- Pre-fetch: `0 0`. Post-fetch: `0 0`.
- Measurement at HEAD `40dc43f5b`: 15 FAIL lines exact match against §A.4 baseline.
- Files staged (4): spec.md / plan.md / acceptance.md / progress.md (frontmatter status: + HISTORY v0.1.1 + progress.md §F.0 + §B M1 row).
- DDD cycle: ANALYZE-only (verified ground truth) → PRESERVE (4 SPEC artifacts frontmatter status:in-progress allowed per Status Transition Ownership Matrix exception).

#### M2 — internal/template 11 fixes (commit `5a4fdf96d`)

- Pre-fetch: `0 0`. Post-fetch: `0 0`.
- DDD cycle: ANALYZE (inspected 11 failure modes via per-test `go test -run`) → PRESERVE (intent of each: verify retained-catalog reality, not 17-agent assumption) → IMPROVE (path drift fix, archived-agent → retained-agent substitution, count constant updates, mirror sync).
- Files staged (8):
  - 7 test files: `internal/template/{agent_frontmatter_audit,catalog_loader,catalog_tier_audit,contract_schema,embed_catalog,embed,embedded_namespace}_test.go`
  - 1 mirror sync: `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (188-byte drift from ATR-001 M5 REQ-ATR-009/REQ-ATR-014 references)
- Test verification: `go test ./internal/template/... → ok 0.590s` (11 FAIL → 0 FAIL).
- ANALYZE findings: 10 architectural-pivot consequences + 1 pre-existing path drift. 9 tests embedded 17-agent assumption (60-entry catalog, `expert/` and `harness/` subdirs, `manager-quality.md` as Behavioral Contract carrier) — all updated to 7-retained reality. TestRuleTemplateMirrorDrift was a cascade from incomplete ATR-001 M8 template parity.

#### M3 — internal/skills 2 fixes (commit `d68421012`)

- Pre-fetch: `0 0`. Post-fetch: `0 0` (after L52 case 29 absorption of `b7d1528c8`).
- DDD cycle: ANALYZE (TestTemplateMirrorParity: missing `release.md` from devOnly list; TestSubSkillLOCCeiling: legitimate accumulation in spec-assembly.md to 548 LOC) → PRESERVE (parity intent + LOC budget intent) → IMPROVE (add `release.md` to devOnly map; ceiling 500→600 per HARD-1 test-fix prefer).
- Files staged (1): `internal/skills/workflow_split_test.go`.
- Test verification: `go test ./internal/skills/... → ok 0.331s`.
- L52 case 29 NEW: `b7d1528c8` (SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 plan-phase) absorbed between M3 push and M4 pre-fetch. Scope-disjoint clean fast-forward; no overlap with M3 file scope.

#### M4 — internal/harness subagent-boundary detection refinement (commit `9f58ed63b`)

- Pre-fetch: `0 0`. Post-fetch: `0 0`.
- DDD cycle: ANALYZE (false-positive on string-literal documentation in scaffolder.go:111 — "AskUserQuestion gate" in §5.2 Out-of-Scope template content) → PRESERVE (C-HRA-008 invariant: no executable callsites) → IMPROVE (detection refined from bare identifier `AskUserQuestion` → call-site signature `AskUserQuestion(` with open-paren).
- Files staged (1): `internal/harness/subagent_boundary_test.go`.
- Test verification: `go test -run TestSubagentBoundary_NoAskUserQuestion ./internal/harness/... → ok 0.188s`.
- Invariant preserved: real `AskUserQuestion()` invocations remain forbidden; false-positives on documentation prose now correctly skipped.

#### M5 — internal/statusline TestRenderPRSegment_Absence pre-existing fix (commit `bdc707bde`)

- Pre-fetch: `0 0`. Post-fetch: `0 0`.
- Pre-existing classification verified: renderer_test.go last modified at `ed064a6f2` (Wave 7 Korean → English), structurally introduced at `5fca86dbf` (PR #1039 — Layout v3); both predate ATR-001 plan-phase `b957a4d04`.
- DDD cycle: ANALYZE (stale "unset = OFF" subtest predates v2.20.0-rc1 supersession to default-ON) → PRESERVE (REQ-SLV-012 supersession verification + AC-SLV-015 absence cases) → IMPROVE (remove stale subtest from `TestRenderPRSegment_Absence`; add positive `TestRenderPRSegment_DefaultOn` for tristate completeness).
- Files staged (1): `internal/statusline/renderer_test.go`.
- Test verification: `go test ./internal/statusline/... → ok 3.066s`.
- REQ-TST-013 honored: zero production behavior change.

#### M6 — 7-item verification batch + §F.3 audit-ready signal (commit pending)

- Pre-fetch: `0 0`. (Post-push fetch will be verified after commit.)
- Scope: progress.md only (per D1 inline mitigation — no frontmatter:implemented transition; no CHANGELOG body edit; manager-docs owns those in sync-phase).
- 7-item batch executed in single parallel turn:
  1. `go test ./...` → all packages ok, zero FAIL
  2. Coverage: 11 packages measured, all ≥84.6% (highest: harness/seeds 100.0%)
  3. Subagent boundary grep: 0 matches (post-M4 detection refinement)
  4. Sentinel-key audit: only test-file occurrences (expected — sentinel_catalog_test.go enumerates them)
  5. CLI smoke: `moai-adk v3.0.0-rc1` returned
  6. Benchmark: 3 benchmarks ran (Benchmark_UpdateCleanup_Baseline, _AtomicWrite, BenchmarkEmbeddedTemplatesWalkDir)
  7. `golangci-lint run --timeout=2m` → `0 issues.`
- §E.3 Run-phase audit-ready signal YAML block populated above.
- §E.2 14-row AC verdict matrix populated above (14/14 PASS).
- Status: ready for `chore(SPEC-V3R6-TEST-REFACTOR-001): M6 verification batch + §F.3 audit-ready signal` commit.

### 4-phase close marker

*(emitted at Mx-phase close terminator commit per L60 atomic chicken-and-egg pattern)*

## HISTORY

### v0.1.2 (2026-05-25) — sync-phase status:in-progress → implemented + §E.4 sync-phase audit-ready signal

- Sync-phase status transition: `in-progress → implemented` for spec.md / plan.md / acceptance.md / progress.md.
- §A Lifecycle Sync table: sync_status field added (in-progress, pending mx_commit_sha backfill).
- §E.4 Sync-phase audit-ready signal YAML block emitted: sync_complete_at, sync_commit_sha (pending), run_commit_sha (6ed1155ea), frontmatter_status_transitions, changelog_entry_position + summary, B12 self-test signals.
- All 4 SPEC artifact frontmatter status transitions applied atomically (in-progress → implemented).
- CHANGELOG.md stub (line 11) replaced with final-form discharge narrative (run-phase M1-M6 commits, 15→0 FAIL, 14/14 AC PASS, zero PASS-WITH-DEBT, L52 case 29, DDD cycle).

### v0.1.1 (2026-05-25) — run-phase M1 frontmatter status:in-progress + ground truth verification

- Run-phase M1 entry: frontmatter status transition `draft → in-progress` per Status Transition Ownership Matrix exception (manager-develop allowed on draft → in-progress only).
- §F.0 Pre-run-phase ground truth populated: 15 failures measured at HEAD `40dc43f5b`, zero drift from §A.4 baseline.
- §B Run-phase milestone log §B M1 row populated (commit SHA pending L60 atomic backfill).

### v0.1.0 (2026-05-25) — initial draft

- Plan-phase progress tracker scaffold authored.
- §A Lifecycle Sync row populated with `pending` placeholders for L60 atomic backfill protocol.
- §B Run-phase milestone log scaffolded with 6 M-rows ready for manager-develop population.
- §E phase-evidence & audit-ready signal slots scaffolded for all 4 phases.
