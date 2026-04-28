---
spec_id: SPEC-V3R3-ARCH-007
title: Task Decomposition — Token Circuit Breaker
version: "1.0.0"
status: draft
created: 2026-04-25
related_plan: .moai/specs/SPEC-V3R3-ARCH-007/plan.md
related_spec: .moai/specs/SPEC-V3R3-ARCH-007/spec.md
---

# 작업 분해 — SPEC-V3R3-ARCH-007

> **범례**:
> - **File owner**: 단독 소유 파일 경로
> - **Depends on**: 선행 task ID
> - **Wave**: A.1 / A.2 / A.3 / A.4
> - **Parallel OK**: 동일 Wave 내 병렬 가능 여부

---

## 전체 Task 개요

| Wave | Task 수 | Parallel 가능 | Sequential 필수 |
|------|---------|---------------|-----------------|
| A.1 runtime.yaml schema | 2 | T-A1-1 → T-A1-2 sequential (template은 local 복사) | — |
| A.2 Go runtime 모듈 | 4 | T-A2-1 → 2/3/4 (구조 의존) | T-A2-1 first |
| A.3 SessionStart 훅 통합 | 1 | — | T-A3-1 |
| A.4 Verification | 3 | — | T-A4-1 ~ 3 |

**총 task 수: 10**

---

## Wave A.1 — runtime.yaml Schema (2 tasks)

### T-A1-1: `.moai/config/sections/runtime.yaml` 작성
- **File owner**: `.moai/config/sections/runtime.yaml`
- **Depends on**: 없음
- **Parallel OK**: —
- **Action**: plan.md §2.1 schema 그대로 작성 (per_agent_budget 14개 agent + circuit_breaker + progress_persistence)
- **Verification**: AC-ARCH007-01 부분 (key grep 통과)
- **Rollback**: 파일 삭제

### T-A1-2: `internal/template/templates/.moai/config/sections/runtime.yaml` 동기화
- **File owner**: `internal/template/templates/.moai/config/sections/runtime.yaml`
- **Depends on**: T-A1-1
- **Parallel OK**: —
- **Action**: T-A1-1 결과 파일 그대로 복사
- **Verification**: `diff -q` 두 파일 동일

### Wave A.1 Checkpoint
- AC-ARCH007-01 통과

---

## Wave A.2 — Go Runtime 모듈 (4 tasks)

### T-A2-1: `internal/runtime/budget.go` Tracker 골격
- **File owner**: `internal/runtime/budget.go`
- **Depends on**: T-A1-1 (schema 참조)
- **Parallel OK**: First
- **Action**:
  1. Tracker struct 정의 (sync.RWMutex, map[string]int, map[string]time.Time, map[string]int)
  2. NewTracker(*RuntimeConfig) 생성자
  3. RecordCall, Usage, IsApproachingLimit, IsAtHardLimit, DetectStall 5 메서드 구현
  4. Goroutine-safe (mutex 사용)
  5. // SPEC: SPEC-V3R3-ARCH-007 헤더 주석
- **Verification**: `go build ./internal/runtime/...` 성공, 메서드 grep 6개 모두 매치
- **Rollback**: 파일 삭제

### T-A2-2: `internal/runtime/config.go` runtime.yaml 파싱
- **File owner**: `internal/runtime/config.go`
- **Depends on**: T-A2-1
- **Parallel OK**: Yes (T-A2-3과 병렬)
- **Action**:
  1. RuntimeConfig struct 정의 (plan.md §2.2 schema 매핑)
  2. LoadRuntime(path string) (*RuntimeConfig, error) 함수
  3. DefaultRuntimeConfig() *RuntimeConfig (REQ-ARCH007-011)
  4. yaml.Unmarshal 사용 (`gopkg.in/yaml.v3` 권장)
- **Verification**: `go build` 성공, default fallback test (T-A2-4 EC-1)
- **Rollback**: 파일 삭제

### T-A2-3: `internal/runtime/persist.go` progress.md 저장 + resume message
- **File owner**: `internal/runtime/persist.go`
- **Depends on**: T-A2-1
- **Parallel OK**: Yes (T-A2-2와 병렬)
- **Action**:
  1. PersistProgress(specID, waveLabel, approach, nextStep string) (string, error) 함수
  2. SPEC 디렉터리 존재 확인, 부재 시 silent-skip (debug log)
  3. progress.md atomic write (`os.Rename` 패턴)
  4. resume message format per .claude/rules/moai/workflow/context-window-management.md §Resume message format
- **Verification**: 단위 테스트 (T-A2-4 includes)
- **Rollback**: 파일 삭제

### T-A2-4: `internal/runtime/budget_test.go` unit tests
- **File owner**: `internal/runtime/budget_test.go`
- **Depends on**: T-A2-1, T-A2-2, T-A2-3
- **Parallel OK**: Last
- **Action**: 테스트 케이스:
  - TestRecordCallBasic
  - TestUsageRatio
  - TestIsApproachingLimitAt75Pct
  - TestHardLimitAt90Pct
  - TestDetectStall
  - TestRetryMaxFallback
  - TestPersistProgressAt75Pct
  - TestDefaultsWhenConfigMissing (EC-1)
  - TestSilentSkipOnMissingSPECDir (EC-2)
  - TestUnknownAgentUsesDefaultBudget (EC-3)
  - TestConcurrentRecordCall (EC-4, race-safe)
  - TestNoAutoClearInvocation (AC-ARCH007-06)
  - TestPerAgentBudgetOverWarning (AC-ARCH007-07)
- **Verification**: `go test -count=1 -race -v ./internal/runtime/...` 전체 PASS

### Wave A.2 Checkpoint
- AC-ARCH007-02 통과
- AC-ARCH007-04, 05, 06, 07 unit-test 부분 통과

---

## Wave A.3 — SessionStart 훅 통합 (1 task)

### T-A3-1: SessionStart 훅에서 runtime.yaml 로드 + Tracker 초기화
- **File owner**: `internal/cli/hook_session_start.go` (또는 hook 진입점; 시작 전 grep으로 식별)
- **Depends on**: T-A2-1, T-A2-2 (config + tracker 모두 사용 가능 상태)
- **Parallel OK**: —
- **Action**:
  1. `grep -rn "session-start\\|SessionStart" internal/cli/` 로 hook 등록 위치 식별
  2. 해당 함수 내 또는 진입점에 다음 추가:
     ```go
     cfg, err := runtime.LoadRuntime(filepath.Join(projectRoot, ".moai/config/sections/runtime.yaml"))
     if err != nil {
         cfg = runtime.DefaultRuntimeConfig()
     }
     tracker := runtime.NewTracker(cfg)
     // session context에 tracker 노출 (구현 환경에 맞게)
     ```
  3. `runtime` package import 추가
- **Verification**:
  - `grep -rn "runtime.NewTracker\\|runtime.LoadRuntime" internal/cli/` 매치
  - 빌드 성공 (`make build`)
- **Rollback**: 추가 라인 + import 제거

### Wave A.3 Checkpoint
- AC-ARCH007-03 통과

---

## Wave A.4 — Verification (3 tasks)

### T-A4-1: AC verification 일괄 실행
- **Depends on**: T-A3-1
- **Parallel OK**: First
- **Action**: acceptance.md AC-01 ~ 07 모든 verification 스크립트 실행
- **Verification**: 모든 AC 통과

### T-A4-2: make build && make install + dogfood smoke test
- **Depends on**: T-A4-1
- **Parallel OK**: No
- **Action**:
  - `make build` (exit 0)
  - `make install` (binary 갱신)
  - Claude Code 재시작 안내 (HARD constraint per MEMORY.md)
  - 수동 smoke test: 새 세션에서 /moai run 실행, log에 Tracker init 확인
- **Verification**: 빌드 성공, smoke test 시 Tracker init log 확인

### T-A4-3: CHANGELOG에 BC-V3R3-006 entry 추가
- **File owner**: `CHANGELOG.md`
- **Depends on**: T-A4-2
- **Parallel OK**: No
- **Action**: CHANGELOG.md `## v3.0.0-R3 (Pending)` 섹션에 추가:
  ```markdown
  ### Breaking Changes (Warning-First)

  - **BC-V3R3-006**: Token Circuit Breaker added (`.moai/config/sections/runtime.yaml`). Per-agent token budget tracking with warning emission at 75%/90% thresholds. /clear remains MANUAL (never auto-triggered). Hard-fail policy deferred to Phase 5. (SPEC-V3R3-ARCH-007)
  ```
- **Verification**: AC-ARCH007-07 CHANGELOG grep 통과

### Wave A.4 Checkpoint
- AC-ARCH007-01 ~ 07 모두 통과
- DoD 모두 충족

---

## Edge Case Tasks (조건부)

### T-EC-1: SessionStart 훅 진입점 미식별
- **Trigger**: T-A3-1 시작 전 grep 결과 매치 없음
- **Action**: STOP, 진입점 분석 보고 (별도 SPEC 또는 사용자 안내)
- **Verification**: hook registration 위치 확인

### T-EC-2: yaml 라이브러리 미존재
- **Trigger**: T-A2-2 빌드 시 import 실패
- **Action**: `go.mod`에 `gopkg.in/yaml.v3` 추가 (`go get`)
- **Verification**: `go build` 성공

### T-EC-3: progress.md atomic write 실패 (디스크 풀)
- **Trigger**: T-A2-3 PersistProgress 호출 시 IO error
- **Action**: error 반환, log warning, agent 실행 차단 금지
- **Verification**: error 시나리오 테스트

### T-EC-4: race detection 실패
- **Trigger**: T-A2-4 `go test -race` 실패
- **Action**: mutex 잠금 위치 재검토, 필요 시 RWMutex 강화
- **Verification**: `go test -race` PASS
