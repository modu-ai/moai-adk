---
id: SPEC-INTERNAL-TEST-001
title: "internal/ 테스트·CI 베이스라인 복구 — go test ./internal/... green 회복"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/"
lifecycle: spec-anchored
tags: "test, ci, baseline, coverage, go, internal"
tier: S
era: V3R6
related_specs: [SPEC-INTERNAL-ARCH-001, SPEC-INTERNAL-PERF-001, SPEC-INTERNAL-SECURITY-001]
---

# SPEC-INTERNAL-TEST-001 — internal/ 테스트·CI 베이스라인 복구

## HISTORY

| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|-----------|--------|
| 0.1.0 | 2026-07-08 | 최초 작성 (draft). 2026-07-08 internal/ 독립 감사(audit)에서 검증된 5개 발견사항(F1-F5)을 GEARS 요구사항으로 변환. 이전 저작 시도는 파일 기록 전 server API error로 중단되어 본 문서가 최초 기록본. 동일 감사 코호트: SPEC-INTERNAL-ARCH-001 / SPEC-INTERNAL-PERF-001 / SPEC-INTERNAL-SECURITY-001. | manager-spec |

## §A 배경 및 목표 (Why)

`go test ./internal/...`가 현재 green이 아니다. `internal/cli` 테스트 바이너리는 SIGSEGV로 크래시하여 커버리지 측정 자체가 불가능하고, doc-content 게이트 2건이 FAIL이며, 2개 패키지는 커버리지 플로어(85%)를 크게 밑돈다. 이 상태에서는 모든 후속 SPEC의 binding gate(`go test ./...` exit 0)가 측정 불능이 되어 패키지 건강도(package health)를 판정할 수 없다.

본 SPEC의 목표는 **CI 베이스라인 복구**다: `go test ./internal/...`가 0 FAIL / 0 panic으로 완주하고, 커버리지 플로어 미달 2개 패키지를 기준선(85%)까지 끌어올려 패키지 건강도가 다시 측정 가능한 상태로 만든다.

### 검증된 발견사항 (2026-07-08 독립 감사 — 앵커는 본 SPEC 저작 시점 HEAD에서 read-only 재확인됨)

| ID | 심각도 | 위치 | 요약 |
|----|--------|------|------|
| F1 | P0 | `internal/cli/coverage_test.go:58-85` | `TestRunHookEvent_ReadInputError`가 구(舊) fail-closed 계약(`err != nil`)을 단언 — `internal/cli/hook.go:174-187` `runHookEvent`는 의도적으로 fail-open(malformed stdin JSON → default `HookOutput` 출력 + `nil` 반환)으로 변경됨. non-fatal `t.Error` 후 무조건 `err.Error()` 호출로 nil-pointer deref SIGSEGV → `internal/cli` 테스트 바이너리 전체 크래시, 커버리지 기록 불능 |
| F2 | P2 | `internal/cli/agentlint/agent_lint_test.go:1077-1083` | `TestAuthoringDocHasEffortMatrix`의 `expectedAgents`가 stale 17-agent 목록 — archived agent명(`manager-strategy`, `manager-quality`, `manager-project`, `expert-backend` 등 6종 `expert-*`, `researcher`) 및 현행 카탈로그에 없는 `manager-cycle`/`builder-platform` 포함. `agent-authoring.md` effort matrix와 드리프트 |
| F3 | P2 | `internal/spec/drift_doctrine_test.go:37-62` | `TestCloseSubjectDoctrineAmendment/lifecycle-sync-gate.md` FAIL — AC-DLC-011 doc-content 게이트. `.claude/rules/moai/workflow/lifecycle-sync-gate.md`에 combined/abbreviated scope 문구 + prohibition token의 ≤400자 co-location 부재 (기지(旣知) 부채: `project_ac_css_001_rescan_debt.md`) |
| F4 | P1 | `internal/migration/*_test.go` | 전 테스트(17개 `t.Skip`)가 placeholder — 패키지 커버리지 1.1%. `runner.go`/`log.go`/`registry.go`/`version.go`는 완전 구현되어 session-start hook silent path + `moai migration run` CLI에 배선됨. crash-recovery / idempotency / atomic version-file update 보증이 전부 미검증 |
| F5 | P2 | `internal/constitution/{canary,contradiction}.go` | 약 445 LOC 합산 0.0% function coverage — 패키지 41.3% vs 형제 파일 85-98%. `canary_test.go`/`contradiction_test.go` 부재 |

## §B GEARS 요구사항

### REQ-TEST-001 — headline: internal/ 전체 green (Event-driven)

**When** `go test ./internal/... -count=1` is executed at repository HEAD, the internal test suite **shall** complete with exit code 0, zero `FAIL` packages, and zero unrecovered panics.

### REQ-TEST-002 — internal/cli 테스트 바이너리 SIGSEGV 제거 (F1, Event-driven)

**When** `TestRunHookEvent_ReadInputError` exercises `runHookEvent` with a failing `ReadInput` mock, the test **shall** assert the current fail-open contract — `err == nil` and default `HookOutput` emission — and the `internal/cli` test binary **shall** complete without SIGSEGV, reporting a coverage percentage.

### REQ-TEST-003 — agentlint doc 드리프트 정합 (F2, Event-driven + Where)

**When** `go test -run TestAuthoringDocHasEffortMatrix ./internal/cli/agentlint/` is executed, the test **shall** PASS. The test's `expectedAgents` list **shall** contain no archived agent names and **shall** reflect the current retained agent catalog. **Where** the remediation edits `agent-authoring.md` (template-mirrored artifact), the change **shall** follow the Template-First Rule (`internal/template/templates/...` 원본 선편집 → `make build` → local `.claude/` mirror 동기화).

### REQ-TEST-004 — AC-DLC-011 doc-content 게이트 충족 (F3, Event-driven)

**When** `go test -run TestCloseSubjectDoctrineAmendment ./internal/spec/` is executed, both subtests **shall** PASS. `.claude/rules/moai/workflow/lifecycle-sync-gate.md` **shall** contain a close-subject full-ID mandate amendment in which a same-line `combined`/`abbreviated` + `scope` phrase is co-located within ≤400 chars (either order) with a prohibition token (`MUST use individual full-ID` / `prohibited` / `disallowed`), per the oracle regex in `drift_doctrine_test.go:47-48`.

### REQ-TEST-005 — migration 패키지 테스트 실체화 (F4, Ubiquitous)

The `internal/migration` package **shall** verify its implemented guarantees — crash-recovery (`detectInFlightState`/`cleanupInFlightState`), idempotency, and atomic version-file update — through executable tests: every `t.Skip("waiting for migration package implementation")` placeholder **shall** be replaced with a real test body, and package statement coverage **shall** be ≥ 85%.

### REQ-TEST-006 — constitution 미검증 파일 테스트 추가 (F5, Ubiquitous)

The `internal/constitution` package **shall** gain unit tests covering `canary.go` (`NewCanary`/`Evaluate`/`estimateScoreImpact`/`sortMostRecent`/`parseScoreFromProgress`) and `contradiction.go` (`NewContradictionDetector`/`Scan`/`containsSequence`/`containsWord`/`extractAction`/`extractModifier`/`isOppositeModifier`), raising package statement coverage to ≥ 85%.

## §C 제약사항 (Constraints)

- **C-1 (unwanted)**: The remediation **shall not** alter the production fail-open behavior of `runHookEvent` (`internal/cli/hook.go:174-187`) — 테스트를 현행 계약에 맞추는 방향만 허용.
- **C-2 (unwanted)**: The remediation **shall not** create a template mirror of `lifecycle-sync-gate.md` — 해당 파일은 dev-only local 파일(내부 SPEC ID 포함)로, 템플릿 미러 생성은 template neutrality(§25) 위반.
- **C-3 (unwanted)**: F4/F5 remediation **shall not** modify migration/constitution production code — tests-only. 테스트 작성 중 실제 production 결함이 발견되면 silent fix 대신 blocker report로 반환.
- **C-4**: **While** authoring migration tests, all filesystem interaction **shall** use `t.TempDir()` isolation — 실제 `~/.moai` 또는 프로젝트 루트 파일 접촉 금지 (CLAUDE.local.md §6).
- **C-5**: 커버리지 플로어는 CLAUDE.local.md 기준 준수 — package-level 85% 최소 (migration, constitution 적용 대상).

## Out of Scope (제외 범위)

다음 항목은 본 SPEC의 out of scope로 명시적으로 제외한다.

### Out of Scope — internal/statusline env-flaky test

- `TestBuild_WritesContextUsageWithSessionID`의 env/state 의존 flake(`context_window_size=1000000, want 256000`)는 별도 pre-existing 부채(`project_ac_css_001_rescan_debt.md` 기록)로, trivially co-fixable하지 않는 한 본 SPEC에서 수정하지 않는다. residual risk로만 문서화한다 (acceptance.md §D.3 재실행 프로토콜 참조).

### Out of Scope — runHookEvent production 동작 변경

- `internal/cli/hook.go`의 fail-open 계약은 의도된 설계("Malformed/truncated stdin JSON must NEVER fail the hook pipeline")이며 본 SPEC은 이를 변경하지 않는다. 수정 대상은 stale 테스트 단언뿐이다.

### Out of Scope — migration/constitution production 코드 리팩터링

- F4/F5는 테스트 추가만 다룬다. production 코드의 구조 개선·버그 수정은 발견 시 blocker report → 후속 SPEC 소관.

### Out of Scope — CI workflow 변경

- `.github/workflows/**` 편집 없음. 본 SPEC은 테스트 스위트 자체를 green으로 만드는 것이지 CI 파이프라인 구성을 바꾸지 않는다.

### Out of Scope — lifecycle-sync-gate.md 템플릿 미러 생성

- 해당 파일은 CLAUDE.local.md §2 Local-Only 목록에 등재된 dev-only 파일이다. 템플릿 트리로의 미러링은 금지(C-2)이며 본 SPEC 범위 밖이다.

## 참조 (Cross-References)

- `internal/cli/coverage_test.go` / `internal/cli/hook.go` — F1 앵커
- `internal/cli/agentlint/agent_lint_test.go` / `.claude/rules/moai/development/agent-authoring.md` (+ template mirror `internal/template/templates/.claude/rules/moai/development/agent-authoring.md`) — F2 앵커
- `internal/spec/drift_doctrine_test.go` / `.claude/rules/moai/workflow/lifecycle-sync-gate.md` / `.claude/rules/moai/development/spec-frontmatter-schema.md` § Close-subject full-ID mandate (미러링할 canonical prohibition prose) — F3 앵커
- `internal/migration/{runner,log,registry,version}.go` + `*_test.go` — F4 앵커
- `internal/constitution/{canary,contradiction}.go` — F5 앵커
- CLAUDE.local.md §2 (Template-First Rule, Local-Only Files), §6 (Test Isolation, Coverage Targets)
- 메모리 부채 기록: `project_ac_css_001_rescan_debt.md` (F3 + statusline flake)
