---
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 — Implementation Plan (Tier S minimal Section A)"
version: "0.2.0"
status: implemented
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "session, coverage, test-only, behavior-preserving, tier-s, sprint-10, plan"
sync_commit_sha: "<pending>"
---

# SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — Implementation Plan (Tier S minimal Section A)

## §A. Context

### §A.1 Tier classification + 1-pass justification

**Tier S (Simple)** — `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier 기준:

| Tier 판정 항목 | 본 SPEC 값 | Tier S 임계값 | 충족 여부 |
|---------------|----------|-------------|---------|
| LOC scope (run-phase test additions) | ~200-280 LOC 신규 test code 예상 | < 300 | ✓ |
| Files affected (run-phase) | 1-3 `_test.go` 파일 신규 또는 기존 보강 | < 5 | ✓ |
| Production file mutation | 0 (강제) | 0 | ✓ (REQ-SLCO-002 binary) |
| Risk profile | test-only, behavior-preserving, package self-contained | Low | ✓ |
| Cross-cutting impact | `internal/session` 단일 패키지, 외부 API 변경 없음 | Low | ✓ |
| AC 개수 | 7 ACs (manageable) | n/a | ✓ |

**1-pass plan-phase 정당화**:
- 본 SPEC plan-phase 는 어떤 test code 도 작성하지 않으며, coverage gap target enumeration (§A.4) + run-phase delegation contract (§A.5) + risk 분석 (§A.6) 에 한정. Section B-E (Known Issues / Pre-flight / Constraints / Self-Verification deliverables) 는 run-phase delegation 시점에 적용되며 plan-phase 자체에는 불필요.
- 선례 정합성: SPEC-LINT-CLEANUP-001 + HARNESS-NAMESPACE-CLEANUP-001 (둘 다 Tier S minimal) 모두 plan-phase 1-pass 성공. 본 SPEC도 Sprint 10 entry SPEC 로 동일 cohort 진입 (Tier S minimal 1-pass cohort 14 → 15 sustain 목표).
- plan-auditor PASS 임계값: Tier S 최소 **0.75** (workflow doctrine), skip-eligible 0.90 (Phase 0.5). 본 SPEC self-audit 목표 ≥ 0.90 으로 skip-eligible 노림.

### §A.2 SPEC 산출물 경로 + 라인 카운트 (plan-phase 종료 시점 예상)

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main` (Hybrid Trunk per CLAUDE.local.md §23.7 — all-tier main 직진 허용)
- HEAD SHA at plan-phase start: `ebe49267059343fd2e6a75781da42373f950d8c9`
- SPEC artifacts (4 files exactly — L59 pre-commit staging assertion):
  - `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md` — canonical SSOT (≈110 lines)
  - `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/plan.md` — this file (Tier S minimal Section A only, ≈155 lines)
  - `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/acceptance.md` — 7 ACs (REQ-SLCO-001..007) + §G B12 self-test (≈115 lines)
  - `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md` — lifecycle + audit-ready signal (≈90 lines)
- plan-auditor verdict: self-audit 목표 PASS ≥ 0.90 (skip-eligible). 본 plan.md 작성 직후 자체 평가 후 progress.md §B 에 기록.

### §A.3 Multi-session race context (2026-05-25 시점)

본 plan-phase 시작 시점 `git status --porcelain | wc -l` = **10** dirty entries (M + ??). pre-spawn fetch `0 0` (clean) verified. 진행 중인 parallel session 작업 가능성:
- `.moai/harness/usage-log.jsonl` (M) — runtime-managed, 손대지 않음.
- `.moai/harness/learning-history/`, `.moai/harness/observations.yaml`, `.moai/research/anthropic-best-practices-2026-05-24.md`, `.moai/research/v3.0-redesign-2026-05-23.md`, `i18n-validator` (??) — parallel research / runtime, 손대지 않음.
- `.moai/specs/.moai/` (??) — 부정 path, 본 SPEC 무관.

본 SPEC plan-phase write target 은 `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/**` 4 파일 **만**. 다른 dirty / untracked 항목과 disjoint.

### §A.4 Existing infrastructure (PRESERVE vs EXTEND)

**PRESERVE (plan-phase 수정 금지) — 10 dirty/untracked entries** (verified via `git status --porcelain` at plan-phase start):

| # | Path | Type | Source |
|---|------|------|--------|
| 1 | `.moai/config/sections/git-convention.yaml` | M | dev settings (§22 maintainer-local intent) |
| 2 | `.moai/config/sections/language.yaml` | M | dev settings |
| 3 | `.moai/config/sections/quality.yaml` | M | dev settings |
| 4 | `.moai/harness/usage-log.jsonl` | M | runtime-managed (harness telemetry) |
| 5 | `.moai/harness/learning-history/` | ?? | runtime-managed (harness self-evolution) |
| 6 | `.moai/harness/observations.yaml` | ?? | runtime-managed |
| 7 | `.moai/research/anthropic-best-practices-2026-05-24.md` | ?? | parallel session audit artifact |
| 8 | `.moai/research/v3.0-redesign-2026-05-23.md` | ?? | parallel session research artifact |
| 9 | `.moai/specs/.moai/` | ?? | spurious nested directory (parallel session typo or test artifact — 본 SPEC scope 외) |
| 10 | `i18n-validator` | ?? | parallel session artifact (CLI subproject root) |

→ 본 plan-phase commit 은 위 10 개 모두 staged 영역에 포함 금지. `git add .moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/` **path-specific add 만** 사용. `git add -A` / `git add .` 절대 금지 (L46 + L59 pre-commit staging assertion).

**EXTEND (plan-phase) — 4 entries exactly (L59 staging assertion target)**:

| # | Path | Type | Operation |
|---|------|------|-----------|
| 1 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md` | NEW | plan-phase 에서 manager-spec 가 신규 작성 (canonical SSOT) |
| 2 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/plan.md` | NEW | this file (Tier S minimal Section A only) |
| 3 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/acceptance.md` | NEW | 7 ACs matrix + §G B12 self-test |
| 4 | `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md` | NEW | lifecycle + audit-ready signal |

**EXTEND (future run-phase) — coverage gap targets enumeration**:

run-phase 가 보강할 함수 우선순위 (per spec.md §2.2 per-function coverage table):

| Priority | File | Function | Current | Gap | Run-phase strategy |
|---------|------|----------|---------|-----|-------------------|
| P1 | `state.go` | `MarshalJSON` | **0.0%** | 85.0% | 신규 `state_marshal_test.go` 또는 `state_test.go` 보강 — round-trip test (Marshal → Unmarshal → DeepEqual) + edge case (nil fields, empty strings) |
| P2 | `store.go` | `mergePhaseStates` | 69.6% | 15.4% | `team_merge_test.go` 보강 — multi-session merge scenarios (3+ session conflict resolution) |
| P2 | `store.go` | `WriteRunArtifact` | 66.7% | 18.3% | `store_test.go` 보강 — run artifact write success / IO error / 권한 부족 case 추가 |
| P3 | `store.go` | `RecordBlocker` | 75.0% | 10.0% | `blocker_test.go` 보강 — blocker record round-trip + invalid input case |
| P3 | `store.go` | `ResolveBlocker` | 74.2% | 10.8% | `blocker_test.go` 보강 — resolve before record / double resolve / concurrent resolve case |
| P4 | `state.go` | `UnmarshalJSON` | 77.3% | 7.7% | `state_test.go` 보강 — malformed JSON / missing required fields / extra fields case |
| P4 | `store.go` | `Checkpoint` | 76.0% | 9.0% | `checkpoint_test.go` 보강 — checkpoint with no prior state / 대용량 task ledger case |
| P5 | `store.go` | `checkBlockerFiles` | 78.6% | 6.4% | `blocker_test.go` 또는 새 `blocker_files_test.go` — blocker file 부재 / 권한 거부 case |
| P5 | `registry.go` | `detectHost` | 75.0% | 10.0% | `registry_test.go` 보강 — hostname env override / hostname error fallback |

run-phase 가 위 우선순위 P1~P3 만 보강해도 coverage 가 약 77.7% → 85.5% 이상 도달 예상 (rough estimate: gap-LOC × density 0.6 = 약 +8%p). P4-P5 는 marginal — run-phase 진행 중 measured coverage 가 ≥85% 도달 즉시 stop 권장.

**Test files extend / new pattern (run-phase 예상)**:

| Operation | File | Rationale |
|-----------|------|-----------|
| EXTEND | `state_test.go` | MarshalJSON 신규 테스트 추가 (P1 가장 큰 gap) |
| EXTEND | `store_test.go` | WriteRunArtifact / Checkpoint / mergePhaseStates 보강 |
| EXTEND | `blocker_test.go` | RecordBlocker / ResolveBlocker / checkBlockerFiles round-trip 보강 |
| EXTEND or NEW | `team_merge_test.go` (extend) OR `merge_phase_test.go` (new) | mergePhaseStates multi-session scenarios |
| (preserve) | `registry_test.go` | run-phase가 detectHost 만 추가하면 보강 가능 — 작은 delta |

run-phase manager-develop 가 (a) EXTEND-only 전략, (b) NEW file 전략 둘 중 선택. 권장은 EXTEND (테스트 응집성 + 기존 helper 재사용). REQ-SLCO-005 (subagent boundary) 정합성 위해 신규 helper 도 `AskUserQuestion` 호출 금지.

### §A.5 Run-phase delegation contract

future `/moai run SPEC-V3R6-SESSION-LEGACY-COVERAGE-001` 실행 시 manager-develop 위임 prompt 는 다음을 포함해야 한다 (Section A-E minimal form per Tier S applicability):

- **Section A (Context)**: spec.md §2 baseline coverage 출력 + §A.4 EXTEND table 우선순위 P1~P3 (P4-P5 는 "coverage ≥85% 도달 시 stop" 가이드) + production 0-byte 변경 강제 (REQ-SLCO-002).
- **Section B (Known Issues — selective subset)**: B1 cross-platform build tag (registry_lock_unix.go / registry_lock_windows.go parity), B3 subagent boundary (REQ-SLCO-005 + 기존 `subagent_boundary_test.go` 정합성 유지), B8 working tree hygiene (PRESERVE 10 entries 손대지 않음). B2/B4/B5/B6/B7 는 본 SPEC scope 외 — 명시적으로 N/A 표기.
- **Section D (Constraints)**: REQ-SLCO-002 production 0-byte (binary `git diff --name-only | grep -v _test.go | grep '^internal/session/' | wc -l` = `0`) + REQ-SLCO-003 t.TempDir() 강제 + REQ-SLCO-007 no t.Setenv OTEL.
- **Section E (Self-Verification deliverables)**: AC matrix (acceptance.md §D) 모든 PASS/FAIL 출력 + coverage 측정 명령 출력 + race detector 결과 + windows cross-build 결과 + subagent boundary grep 결과.

### §A.6 Risks

| # | Risk | Severity | Mitigation |
|---|------|---------|-----------|
| R1 | run-phase 가 신규 테스트 작성 중 production code 결함을 발견하고 자체 수정 시도 | Critical | REQ-SLCO-002 [Unwanted] binary verification (AC-SLCO-002 `git diff --name-only` 필터 `_test.go` 만) + Section D Constraints 명시. 결함 발견 시 (a) 테스트 작성 skip, (b) blocker report return, (c) 별도 SPEC 분리 권장 |
| R2 | t.TempDir() 사용 실수로 project root 에 임시 파일 생성 | High | REQ-SLCO-003 + AC-SLCO-003 grep verification (`grep -rn 't\.TempDir()' new_tests`). CLAUDE.local.md §6 filepath.Abs() 규칙 reference |
| R3 | 신규 테스트가 goroutine 도입하여 data race 발생 | Medium | REQ-SLCO-004 + AC-SLCO-004 `go test -race` 명령 PASS 필수. 가능하면 goroutine-free 테스트 우선 |
| R4 | OTEL env var setenv 시도 (CLAUDE.local.md §2 [WARN]) | Medium | REQ-SLCO-007 + AC-SLCO-007 grep verification (`grep -rn 't\.Setenv.*OTEL_' new_tests` = 0). Fake exporter 패턴 강제 |
| R5 | `registry_lock_windows.go` 측 보강 시 build tag 누락 → `GOOS=windows go build` 실패 | Medium | REQ-SLCO-006 + AC-SLCO-006 cross-platform build verification 필수. windows-only 테스트는 `//go:build windows` 헤더 강제 |
| R6 | coverage 측정이 85% 도달했으나 새로운 미커버 분기 introduced | Low | AC-SLCO-001 binary verification (`go test -cover` 출력 grep) 으로 일관 검증. 측정 명령 변경 금지 |
| R7 | parallel session 이 `internal/session/` production 파일 동시 수정 | Low | pre-spawn fetch `0 0` 의무 (L44 HARD) + run-phase entry 시점 `git rev-list --count --left-right origin/main...HEAD` 재확인. divergence 발견 시 STOP |
| R8 | `subagent_boundary_test.go` grep guard 가 신규 테스트 함수명 (예: `TestAskUserQuestion...`) 을 false-positive trigger | Low | REQ-SLCO-005 + 신규 테스트 함수명에 `AskUserQuestion` 또는 `mcp__askuser` 토큰 사용 금지. (이름 우회 우회 — 정직한 boundary 준수) |

## §B-§E (Tier S minimal — intentionally omitted)

본 SPEC 은 Tier S minimal 변형이므로 plan.md Section B (Known Issues 8 카테고리) / Section C (Pre-flight checklist) / Section D (Constraints) / Section E (Self-Verification deliverables) 는 plan.md 본문에서 생략한다. 이들 섹션의 내용은 run-phase delegation prompt 구성 시 §A.5 에 따라 minimal form 으로 inline 주입된다. 선례: SPEC-LINT-CLEANUP-001 + HARNESS-NAMESPACE-CLEANUP-001 모두 동일 변형 적용.

## §F. Cross-references

- `internal/session/` — 13 production files + 10 test files (현재 baseline, plan-phase 진단 시점).
- spec.md §2.2 — per-function coverage gap table (run-phase 우선순위 ground truth).
- acceptance.md — 7 ACs matrix (REQ ↔ AC bidirectional traceability).
- progress.md — plan-phase audit-ready signal + Phase 0.5 plan-auditor verdict 기록.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S minimal 정의 + 1-pass cohort doctrine.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S optional Section A-E (minimal delegation ~500-800 tokens 허용).
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate authority + skip-eligible 0.90 정책.
- `CLAUDE.local.md §6` Testing Guidelines — `t.TempDir()` HARD + `filepath.Abs()` cross-platform 규칙 + 85% coverage minimum.
- `CLAUDE.local.md §2 [WARN]` OTEL t.Setenv 데이터 레이스 회피 (REQ-SLCO-007 origin).
- SPEC-LINT-CLEANUP-001 plan.md / HARNESS-NAMESPACE-CLEANUP-001 plan.md — Tier S minimal Section A 변형 선례.
- MEMORY.md `Sprint 9 lane A AAT-001 4-phase FULLY CLOSED + Sprint 10 entry pending` — 본 SPEC 이 Sprint 10 entry 5 후보 (A SESSION-LEGACY-COVERAGE-001 권장) 임을 명시.
