---
id: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001
title: "Progress — Harness Namespace 누출 검증 및 정리"
version: "0.1.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: Low
phase: "v3.0.0 cleanup"
module: ".claude/agents/, .claude/skills/, internal/template/"
lifecycle: spec-anchored
tags: "harness, namespace, cleanup, progress, tier-s"
issue_number: null
tier: S
plan_commit_sha: ""
run_commit_sha: ""
sync_commit_sha: ""
mx_commit_sha: ""
---

# Progress — SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001

## §E. Phase Tracking

### §E.1 Plan-phase Audit-Ready Signal

**Date**: 2026-05-25
**Agent**: manager-spec
**Session**: Sprint 8 P4 follow-up — concurrent with SPEC-V3R6-MULTI-SESSION-COORD-001 run-phase (disjoint scope verified)
**Tier**: S (minimal — Section A only)
**Plan commit SHA**: TBD (orchestrator는 본 plan-phase 산출물 commit 후 backfill)

#### Audit Findings

| 항목 | 상태 | Evidence |
|------|------|----------|
| Template `agents/{core,expert,meta}/` only | PASS | `find internal/template/templates/.claude/agents -type d` → `{., core, expert, meta}` |
| Template `moai-harness-learner` only | PASS | `ls -d internal/template/templates/.claude/skills/moai-harness-*` → `moai-harness-learner` 단일 |
| Local `.claude/agents/harness/` leaked | VIOLATION | 4 specialist `.md` files (9,267 bytes total) |
| Local `moai-harness-cli-template` leaked | VIOLATION | `SKILL.md` 4,683 bytes |
| Local `moai-harness-patterns` leaked | VIOLATION | `SKILL.md` 10,052 bytes |
| `internal/cli/update.go` isUserOwnedNamespace | PASS | line 1186 `agents/harness/` 보호 명시, line 1181 `my-harness-*` 보호 명시 |
| `internal/cli/update_namespace_protect.go` Contract | PASS | REQ-UNP-006 sentinel + backup pattern 정상 구현 |
| `internal/cli/update.go` isMoaiManaged exclusion note | PASS | line 1240-1244 `agents/harness/` 의도적 제외 명시 |
| SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 머지 상태 | PASS | commit `767bc04a4` (PR #1048) on main |
| Pre-spawn git sync check | PASS | `git rev-list --count --left-right origin/main...HEAD` → `0 0` |
| Multi-session race scope disjoint | PASS | COORD-001 = `internal/governance/, internal/session/registry*`; 본 SPEC = `.claude/agents/harness/, .claude/skills/moai-harness-*, internal/template/` |

#### Plan-auditor Verdict

**plan-auditor iter-1 verdict: PASS 0.93** (Tier S threshold **0.75** per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L), +0.18 margin; skip-eligible threshold 0.90, +0.03 margin). 4 minor fix-forward defects identified (D1 progress.md Tier S threshold annotation / D2 acceptance.md §D.5 plan.md section refs / D3 plan.md §A.2 deferred cross-refs / D4 acceptance.md §D.4 AC-HNC-003 severity note) — all non-blocking, addressed inline during run-phase.

Self-assessment (advisory, manager-spec internal estimate):

- Trace: 0.95 (7 REQ ↔ 7 AC ↔ 3 milestone 매핑 완비)
- Faith: 0.95 (§24 SSOT contract verbatim 인용, UPDATE-NAMESPACE-PROTECT-001 dependency 명시)
- Soundness: 0.95 (template 무수정 invariant + backup-first + Go test regression dual-gate)
- Clarity: 0.92 (Korean prose + English code/identifier separation, 7 REQ 명확 분리)
- Aggregate (harmonic mean estimate): ~0.94 (skip-eligible 가능, 단 본 평가는 plan-auditor가 독립적으로 재산정)

#### Plan-phase Self-Check Conclusion

- 4 artifacts (spec.md, plan.md, acceptance.md, progress.md) 작성 완료
- 12 canonical frontmatter fields 모두 4 파일에 일관 적용
- §7 Exclusions 7 sub-sections (h3 sub-headings) 포함
- §3 Non-requirements + §4 Constraints + §5 Dependencies + §6 Impact 모두 명시
- SPEC ID Pre-Write Self-Check decomposition: `SPEC ✓ | V3R6 ✓ | HARNESS ✓ | NAMESPACE ✓ | CLEANUP ✓ | 001 ✓ → PASS`
- Multi-session race disjoint scope 유지 (CRITICAL CONSTRAINTS 준수)
- NO commit / NO push (manager-spec 권한 범위 준수)
- 다음 단계: orchestrator의 plan-auditor 호출 + 사용자 검토 + commit 결정

### §E.2 Run-phase Evidence (manager-develop 책임)

**Date**: 2026-05-25
**Agent**: manager-develop (Tier S LEAN delegation)
**Methodology**: ANALYZE-PRESERVE-IMPROVE (DDD per quality.yaml development_mode)

#### AC Roster (binary matrix)

| AC | Status | Verification | Actual Output |
|----|--------|--------------|---------------|
| AC-HNC-001 | PASS | `find templates/.claude/agents -type d -name harness` + `ls -d templates/.claude/skills/moai-harness-*` + `find templates/.claude/agents -mindepth 1 -maxdepth 1 -type d` | 0 / 1 (moai-harness-learner only) / 3 lines (core, expert, meta) |
| AC-HNC-002 | PASS | `test ! -d .claude/agents/harness && test ! -d .claude/skills/moai-harness-{cli-template,patterns} && test -d .claude/skills/moai-harness-learner` | all 4 conditions hold |
| AC-HNC-003 | PASS | M3 read of skill-authoring.md §285-307 + agent-authoring.md §13-36 + 4 plan-phase verified locations (update.go:1166, update.go:1240-1244, update_namespace_protect.go:7, moai-meta-harness/SKILL.md) | all 6 cross-refs cite §24 without weakening; HARD enforcement language preserved across all sources |
| AC-HNC-004 | PASS | 3-grep post-cleanup parallel batch | 0 / 0 / 0 (all three commands) |
| AC-HNC-005 | PASS | `go test -v -run "TestTemplateAgentsStructure\|TestTemplateMoaiHarnessSkillsAllowlist" ./internal/template/...` | `--- PASS: TestTemplateAgentsStructure (0.00s)` + `--- PASS: TestTemplateMoaiHarnessSkillsAllowlist (0.00s)` + `PASS ok github.com/modu-ai/moai-adk/internal/template 0.442s` |
| AC-HNC-006 | PASS | `git diff --stat -- internal/template/templates/` | empty (no template body modifications) |
| AC-HNC-007 | PASS | Backup integrity 3-check | 1 backup dir / 1 .complete marker / 6 backed-up files |

#### M1 Backup + Cleanup Evidence

- Backup directory: `.moai/backups/harness-namespace-cleanup-2026-05-24T18-53-53Z/` (ISO-8601 hyphenated per UPDATE-NAMESPACE-PROTECT-001 REQ-UNP-010 pattern)
- Backup verification: `diff -r` confirmed byte-identical for all 3 source directories before any `rm` operation
- Files removed: 6 (`.claude/agents/harness/{cli-template,hook-ci,quality,workflow}-specialist.md` + `.claude/skills/moai-harness-{cli-template,patterns}/SKILL.md`)
- Empty parent directories removed: 3 (`.claude/agents/harness/`, `.claude/skills/moai-harness-cli-template/`, `.claude/skills/moai-harness-patterns/`)
- `.claude/skills/moai-harness-learner/` preserved (valid `moai-harness-*` per §24.1 allowlist)

#### M2 Go Integration Test Evidence

- New file: `internal/template/embedded_namespace_test.go` (~140 lines including godoc + 2 test functions)
- Test pattern: `package template` (internal package) using exported `EmbeddedTemplates()` — preserves DEFECT-5 INVARIANT (no direct `embeddedRaw` access)
- Both tests use `t.Parallel()` for consistency with project test conventions
- `TestTemplateAgentsStructure`: PASS — bidirectional check (missing AND unexpected subdirs both detected, future regression hardening)
- `TestTemplateMoaiHarnessSkillsAllowlist`: PASS — strict allowlist `{moai-harness-learner}` enforced
- Sentinel: `HARNESS_NAMESPACE_LEAK` documented in godoc + error messages

#### M3 Cross-Reference Doc Verification Evidence

- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy (lines 285-307): contract correctly cites `moai-harness-*` builder-only + `my-harness-*` user-owned + CI guard requirement; cross-references CLAUDE.local.md §24 + moai-meta-harness/SKILL.md + agent-authoring.md (3 of 6 cross-refs in this doc alone)
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention (lines 13-36): HARD enforcement of `internal/template/templates/.claude/agents/harness/` 존재 금지 + `moai update` `.claude/agents/harness/` sync 제외 + `moai-meta-harness` emit constraint; cross-references skill-authoring.md + moai-meta-harness/SKILL.md (2 cross-refs)
- All 6 §24 cross-references are consistent — no policy-weakening language detected

#### Pre-existing Baseline Test Failures (L46 Attribution)

The following 3 test failures pre-date this SPEC and are confirmed via `git stash` baseline verification — NOT regressions from this SPEC:

| Test | Failure | Attributed SPEC |
|------|---------|-----------------|
| `TestLoadEmbeddedCatalog_Success` | `AllEntries() = 50, want 60` | catalog count drift, separate cleanup SPEC needed |
| `TestRetirementCompletenessAssertion` | `manager-tdd/manager-ddd replacement manager-develop must exist` | SPEC-V3R3-RETIRED-AGENT-001 + SPEC-V3R3-RETIRED-DDD-001 follow-up |
| `TestManifestHashFormat` | `CATALOG_HASH_UNSTABLE: manager-develop/docs/spec` | catalog hash regen cascade from prior SPEC body edits |

Baseline verification protocol: stashed all working tree changes, ran the failing tests on clean HEAD `878801e88`, confirmed identical failure output, restored stash. Conclusion: these failures are L46 attribution to prior SPECs, out of scope for HNC-001.

#### M3.5 Optional Fix-Forward Applied

- D1 (progress.md §E.1 line 52 Tier S threshold): corrected `0.80` → `0.75` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (S/M/L) SSOT; also added iter-1 PASS 0.93 verdict + skip-eligible margin notation
- D2 (acceptance.md §D.5 plan.md section references): corrected `M1.5` → `M1 step 4-5` and `M1.1-1.3` → `M1 steps 1-3` for accurate cross-reference
- D3 (plan.md §A.2 deferred cross-refs): replaced DEFER markers with M3 PASS evidence + line numbers (285-307 for skill-authoring.md, 13-36 for agent-authoring.md)
- D4 (acceptance.md §D.4 AC-HNC-003 severity): NOT applied — AC-HNC-003 already marked SHOULD in §D matrix; redundant annotation in §D.4 would be cosmetic only

### §E.3 Run-phase Audit-Ready Signal (manager-develop 책임)

```yaml
run_complete_at: 2026-05-25T03:58:00Z
run_commit_sha: TBD (orchestrator backfills after M1 commit + push)
run_status: PASS
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 9
l44_pre_commit_fetch: "0 0"  # pre-spawn fetch verified clean before M1
l44_post_push_fetch: TBD (orchestrator backfills post-push)
new_warnings_or_lints_introduced: 0  # golangci-lint ./internal/template/... → 0 issues
cross_platform_build:
  linux_amd64: not_verified  # docs-only Go test, no platform-sensitive code
  darwin_arm64: PASS  # implicit via test execution
  windows_amd64: not_verified  # not in test scope per plan
total_run_phase_files: 5  # 6 deletions + 1 NEW Go test + 3 progress/acceptance/plan edits + 4 frontmatter status edits (consolidated via single commit)
m1_to_mN_commit_strategy: single-commit  # M1+M2+M3+frontmatter+§E.2/.3 fill consolidated per Tier S minimal scope
pre_existing_baseline_failures_attributed:
  - TestLoadEmbeddedCatalog_Success (catalog drift, not HNC-001)
  - TestRetirementCompletenessAssertion (V3R3 retirement follow-up, not HNC-001)
  - TestManifestHashFormat (catalog hash regen, not HNC-001)
mx_scan_required: false  # Mx Step C SKIP-judge likely — only Go test scaffold (pure declarative, no goroutines/complexity≥15/fan_in≥3)
```

### §E.4 Sync-phase Audit-Ready Signal (manager-docs 책임)

(empty)

### §E.5 Mx-phase Audit-Ready Signal (orchestrator + manager-docs 책임)

(empty — 본 SPEC은 plan/run/sync phase에서 .go 변경 ~50 LOC 예상, Mx Step C SKIP-judge 가능성 높음. manager-develop 결정.)

## §F. Cross-References

- spec.md §1 (배경)
- plan.md §A (Tier S minimal plan)
- acceptance.md §D (AC matrix)
- CLAUDE.local.md §24 (SSOT)
- SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (dependency)
- chore commit `4f1135684` (2026-05-23 template cleanup 선례)

## §G. Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Multi-session race (COORD-001 동시 진행) | Medium | Low | Pre-spawn fetch + disjoint scope (검증됨) |
| Backup 실패 → 복구 불가 | Very Low | High | `.complete` marker + 6-file byte-identical 검증 후 삭제 |
| Go test 신규 추가 시 컴파일 실패 | Low | Low | run-phase manager-develop이 `go test ./...` post-add 즉시 검증 |
| Cross-ref doc 비일관 발견 | Low | Low | run-phase M3에서 read-only verify; 발견 시 별도 hotfix SPEC |
| `moai-harness-learner` 누락 발견 | Very Low | Medium | M2 test가 실패 — 즉시 차단 신호 |
