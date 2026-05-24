---
id: SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001
title: "Progress — Harness Namespace 누출 검증 및 정리"
version: "0.1.0"
status: draft
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

**Awaiting plan-auditor invocation** by orchestrator. Tier S threshold 0.80, skip-eligible 0.90.

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

(empty — populated by manager-develop during run-phase)

### §E.3 Run-phase Audit-Ready Signal (manager-develop 책임)

(empty)

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
