---
id: SPEC-V3R6-SESSION-LEGACY-COVERAGE-001
title: "internal/session 패키지 test coverage 보강 (77.7% → ≥85%, test-only)"
version: "0.1.0"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/session"
lifecycle: spec-anchored
tags: "session, coverage, test-only, behavior-preserving, tier-s, sprint-10"
---

# SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 — internal/session 패키지 test coverage 보강

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | 최초 작성 (plan-phase). Tier S minimal Section A 단일 섹션 변형. Sprint 10 lane A entry SPEC. `internal/session` 패키지 baseline coverage 77.7% → 프로젝트 기준 ≥85% (CLAUDE.local.md §6) 으로 보강. **test-only** scope — production .go 파일 0건 수정, behavior-preserving. ANTHROPIC-AUDIT-TIER3-001 (Sprint 9 lane A) 마감 후 coverage gap 정리 SPEC. |

## §1. 목적 (Goal)

`internal/session` 패키지의 test coverage를 **현재 77.7% 기준점에서 프로젝트 표준 ≥85%** (per CLAUDE.local.md §6 Coverage Targets) 로 보강한다. 본 SPEC은 다음 두 원칙을 절대 준수한다:

- **Test-only scope**: 13개 production `.go` 파일 (blocker.go / checkpoint.go / hydrate.go / lock.go / lock_windows.go / phase.go / registry.go / registry_lock_unix.go / registry_lock_windows.go / state.go / store.go / task_ledger.go + 1 추가) 의 본문은 **단 한 바이트도 수정하지 않는다**.
- **Behavior-preserving**: 기존 10개 `_test.go` 파일이 검증하는 모든 동작은 변경 없이 유지된다. 신규 테스트는 추가만 수행하며 기존 테스트의 assertion 변경/삭제는 금지.

## §2. 배경 (Background)

### §2.1 현재 coverage baseline (2026-05-25 plan-phase 시작 시점)

```
$ go test -cover ./internal/session/...
ok  	github.com/modu-ai/moai-adk/internal/session	1.869s	coverage: 77.7% of statements
```

프로젝트 표준은 CLAUDE.local.md §6 "Coverage Targets" 기준 **package-level 85% minimum, critical packages 90%+**. `internal/session`는 critical 분류 (session lifecycle state SSOT) 에 가깝지만 본 SPEC은 보수적으로 ≥85% (project minimum) 만 충족 목표로 한다. 추가 5%p 도달은 후속 SPEC scope.

### §2.2 Per-function coverage gap (priority targets)

`go tool cover -func` 출력 기준 ≤80% 함수 (인접 함수 묶음 우선):

| File | Function | Current | Gap-to-85 |
|------|----------|---------|----------|
| `state.go:29` | `MarshalJSON` | **0.0%** | 85.0% |
| `state.go:55` | `UnmarshalJSON` | 77.3% | 7.7% |
| `store.go:269` | `mergePhaseStates` | 69.6% | 15.4% |
| `store.go:392` | `WriteRunArtifact` | 66.7% | 18.3% |
| `store.go:459` | `ResolveBlocker` | 74.2% | 10.8% |
| `store.go:423` | `RecordBlocker` | 75.0% | 10.0% |
| `store.go:82` | `Checkpoint` | 76.0% | 9.0% |
| `store.go:328` | `checkBlockerFiles` | 78.6% | 6.4% |
| `registry.go:460` | `detectHost` | 75.0% | 10.0% |
| `registry.go:424` | `FormatStderrReminder` | 94.1% | (≥85 ✓) |

run-phase 우선순위: (1) `state.MarshalJSON` 0% → MUST cover (가장 큰 LOC delta), (2) `store.mergePhaseStates` + `store.WriteRunArtifact` (multi-session merge path + run artifact path — coordination critical), (3) `store.RecordBlocker` + `store.ResolveBlocker` (blocker lifecycle round-trip), (4) `state.UnmarshalJSON` + `store.Checkpoint` + `registry.detectHost` 잔여.

### §2.3 기존 10 test files (PRESERVE list)

```
blocker_test.go              checkpoint_test.go
hydrate_opts_test.go         inflight_test.go
phase_test.go                registry_test.go
state_test.go                store_test.go
subagent_boundary_test.go    team_merge_test.go
```

특히 `subagent_boundary_test.go` 는 SPEC-V3R6-MULTI-SESSION-COORD-001 (방금 close) 가 도입한 **subagent boundary CI guard** (AskUserQuestion / mcp__askuser 호출 금지) 를 정적 검사로 enforce한다. 본 SPEC 신규 테스트는 이 guard 를 어기지 않아야 한다.

### §2.4 13 production files (NEVER MODIFY)

```
blocker.go               checkpoint.go            hydrate.go
lock.go                  lock_windows.go          phase.go
registry.go              registry_lock_unix.go    registry_lock_windows.go
state.go                 store.go                 task_ledger.go
(+ 1 additional file as of internal/session/ listing — verified at run-phase entry)
```

[HARD] 본 SPEC run-phase는 위 13 (또는 ls 결과로 확정되는 N) 개 production 파일을 **0 byte 수정** 한다. AC-SLCO-002 binary verification 으로 enforce.

### §2.5 Cross-platform parity 의무

`registry_lock_unix.go` (build tag `!windows`) + `registry_lock_windows.go` (build tag `windows`) 는 cross-platform parity 가 핵심이다. 현재 unix 측 `acquire` 90% / `release` 85.7% 측정되고 windows 측은 macOS 환경에서 직접 측정 불가하지만 `GOOS=windows GOARCH=amd64 go build ./...` 컴파일 PASS 여야 한다 (B1 cross-platform build tag).

## §3. Requirements (EARS Format / GEARS-compatible)

### REQ-SLCO-001 [Ubiquitous] Coverage threshold
본 SPEC의 run-phase 종료 시점, `go test -cover ./internal/session/...` 출력의 coverage 값은 **≥85.0%** SHALL 달성한다.

### REQ-SLCO-002 [Unwanted] Production file mutation forbidden
본 SPEC의 run-phase는 `internal/session/` 디렉토리 하의 **non-`_test.go` 파일을 단 한 바이트도 수정 SHALL NOT** 한다. 신규 production 파일 생성도 금지.

### REQ-SLCO-003 [Ubiquitous] t.TempDir 강제
신규 추가되는 모든 테스트 함수는 임시 디렉토리가 필요할 때 `t.TempDir()` 를 SHALL 사용한다. `/tmp/...` 하드코딩, `os.TempDir()` 직접 호출, project root 수정 모두 금지 (CLAUDE.local.md §6 Test Isolation 준수).

### REQ-SLCO-004 [Event-Driven] Race detector clean
WHEN `go test -race ./internal/session/...` 가 실행되면, 결과는 zero data race 로 SHALL PASS 한다. 신규 테스트가 goroutine 을 도입할 경우 race 가 발견되어선 안 된다.

### REQ-SLCO-005 [Unwanted] Subagent boundary preserved
본 SPEC의 신규 테스트 또는 보강 테스트는 `internal/session/` 내에 `AskUserQuestion` 또는 `mcp__askuser` 호출 토큰을 SHALL NOT 도입한다. `subagent_boundary_test.go` 의 정적 grep 검사 PASS 유지.

### REQ-SLCO-006 [Ubiquitous] Cross-platform parity
본 SPEC의 run-phase 종료 시점, `GOOS=windows GOARCH=amd64 go build ./internal/session/...` 가 exit code 0 으로 SHALL 통과한다 (registry_lock_windows.go 측 테스트 보강 시 build tag 정합성 유지).

### REQ-SLCO-007 [Unwanted] t.Setenv OTEL forbidden
신규 테스트는 `OTEL_*` 환경 변수에 대해 `t.Setenv` 를 SHALL NOT 사용한다 (CLAUDE.local.md §2 [WARN] 데이터 레이스 회피). OTEL 의존이 필요하면 fake/no-op exporter 주입 패턴 사용.

### REQ-SLCO-008 [Optional] §24 namespace policy untouched
WHERE 가능한 경우, 본 SPEC은 `.claude/skills/` 또는 `.claude/agents/` 디렉토리를 SHOULD NOT 손댄다 — namespace 분리 정책 (CLAUDE.local.md §24) 과 무관한 test-only SPEC 임을 유지.

## §4. Out of Scope (자기 준수 — §3.4 self-compliance per SPEC-LINT-CLEANUP-001 canonical pattern)

### §4.1 Out of Scope — Production behavior change

- **production .go 파일 수정 일체 금지** — bug fix, refactor, rename, API 변경 모두 본 SPEC scope 외. 만약 신규 테스트 작성 도중 production 결함이 발견되면 (a) 본 SPEC plan-phase 보존 + (b) 별도 SPEC (예: `SPEC-V3R6-SESSION-<DOMAIN>-FIX-001`) 분리 작성 후 처리.
- **새로운 production API 추가 금지** — coverage 보강을 위해 production 함수를 export 시키는 행위 금지. 이미 export 된 API + internal helper 의 unexported 함수는 `_test.go` 의 internal test (same package) 로 검증 가능.

### §4.2 Out of Scope — Coverage 90%+ stretch

본 SPEC의 ceiling 은 ≥85% (project minimum). critical-package 90%+ 추가 보강은 별도 follow-up SPEC scope. 현실적으로 신규 테스트로 80% → 87% 도달 시점에서 추가 +3%p 는 marginal cost > marginal benefit.

### §4.3 Out of Scope — 다른 패키지 coverage 보강

`internal/hook/`, `internal/cli/`, `internal/harness/` 등 sibling 패키지의 coverage 갭은 본 SPEC scope 외. 본 SPEC은 `internal/session/` **단일 패키지** 에만 집중.

### §4.4 Out of Scope — Production test fixture refactor

기존 10 `_test.go` 파일이 사용 중인 helper 함수 / fixture / setup 패턴의 리팩터링은 본 SPEC scope 외. 신규 테스트는 기존 helper 를 재사용하거나, 필요 시 **새로운** helper 를 `_test.go` 파일에 추가만 한다.

### §4.5 Out of Scope — CI / lint 통합 변경

`golangci-lint`, `go vet`, `spec-lint`, `staticcheck` 규칙 추가 / 임계값 변경은 본 SPEC scope 외. Coverage 측정 도구 (`go test -cover`) 의 invocation 방식 변경도 scope 외 — 기존 `go test -cover ./internal/session/...` 명령 그대로 사용.

## §5. Dependencies

- **Required**: 없음. `internal/session` 패키지는 self-contained — 본 SPEC의 신규 테스트는 외부 SPEC 산출물에 의존 X.
- **Reference precedent (Tier S minimal 1-pass cohort, 14 entries by Sprint 9 close)**: SPEC-V3R6-SPEC-LINT-CLEANUP-001 (markdown-only Tier S, 4-phase 1-pass) + SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001 (cleanup Tier S, 4-phase 1-pass). 동일 cohort entry.
- **Constitutional dependency**: CLAUDE.local.md §6 Coverage Targets (≥85% project minimum) + §6 Test Isolation (`t.TempDir()` HARD) + §2 [WARN] OTEL setenv 금지.
- **No code dependencies**: 본 SPEC은 production code 미수정. `go test` / `go build` toolchain 외 외부 의존성 없음.
- **Adjacent SPEC (information only)**: SPEC-V3R6-MULTI-SESSION-COORD-001 (방금 close 된 같은 패키지 SPEC). COORD-001 이 도입한 `subagent_boundary_test.go` 의 static grep guard 는 본 SPEC이 preserve 의무 (REQ-SLCO-005).

## §6. Cross-references

- `internal/session/` — 13 production files + 10 test files (현재 baseline).
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S minimal 정의.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability — Tier S optional Section A-E (minimal delegation 허용).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12-field canonical schema (본 SPEC 4 artifacts 준수).
- `.claude/agents/meta/plan-auditor.md` — Phase 0.5 Plan Audit Gate (skip-eligible threshold 0.90).
- `CLAUDE.local.md §6 Testing Guidelines` — t.TempDir() HARD + filepath.Abs() 규칙.
- `CLAUDE.local.md §2 [WARN]` — OTEL setenv 데이터 레이스 회피.
- `SPEC-V3R6-SPEC-LINT-CLEANUP-001 spec.md §3` — canonical H3 "Out of Scope" pattern (본 SPEC §4 self-compliance).
