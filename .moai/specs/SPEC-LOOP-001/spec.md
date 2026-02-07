---
spec_id: SPEC-LOOP-001
title: Ralph Feedback Loop Engine
status: Completed
priority: Medium-High
phase: "Phase 3 - Automation"
created: 2026-02-03
dependencies:
  - SPEC-QUALITY-001
  - SPEC-CONFIG-001
modules:
  - internal/loop/
  - internal/ralph/
estimated_loc: 1200
tags:
  - loop
  - ralph
  - feedback
  - state-machine
  - convergence
  - automation
lifecycle: spec-anchored
---

# SPEC-LOOP-001: Ralph Feedback Loop Engine

## HISTORY

| 날짜 | 버전 | 변경 내용 |
|------|------|----------|
| 2026-02-03 | 1.0.0 | 초기 SPEC 작성 |

---

## 1. Environment (환경)

### 1.1 프로젝트 컨텍스트

MoAI-ADK (Go Edition)는 Claude Code용 고성능 개발 도구킷이다. Ralph Feedback Loop Engine은 Phase 3 (Automation)에 해당하며, 반복적 개발 사이클(analyze, implement, test, review)을 자율적으로 관리하는 피드백 루프 컨트롤러를 구현한다.

### 1.2 기술 환경

- **언어**: Go 1.22+
- **모듈 경로**: `github.com/modu-ai/moai-adk-go`
- **대상 패키지**: `internal/loop/` (controller.go, feedback.go, state.go, storage.go), `internal/ralph/` (engine.go)
- **의존 패키지**:
  - `internal/config/` -- RalphConfig 설정값 로드 (SPEC-CONFIG-001)
  - `internal/core/quality/` -- TRUST 5 품질 게이트 결과 수집 (SPEC-QUALITY-001)
  - `internal/core/git/` -- Git 상태 확인
  - `pkg/utils/` -- 로깅, 타임아웃 유틸리티
- **동시성 모델**: goroutine + context.Context (취소/타임아웃)
- **직렬화**: `encoding/json` (상태 저장/복원)

### 1.3 운영 환경

- 루프 상태는 `.moai/loop/` 디렉토리에 JSON 파일로 영속화
- Claude Code 세션 재시작 시 상태 복원 지원
- 최대 반복 횟수: 5회 (기본값, RalphConfig에서 변경 가능)
- Human-in-the-loop 리뷰 브레이크포인트 지원

---

## 2. Assumptions (가정)

### 2.1 선행 조건

- [A-01] `internal/config/` 모듈이 구현되어 `RalphConfig` 구조체를 포함한 설정값을 로드할 수 있다 (SPEC-CONFIG-001 완료)
- [A-02] `internal/core/quality/` 모듈이 구현되어 TRUST 5 품질 게이트 검증 결과를 반환할 수 있다 (SPEC-QUALITY-001 완료)
- [A-03] SPEC 문서가 `.moai/specs/SPEC-{ID}/spec.md` 경로에 존재한다

### 2.2 기술 가정

- [A-04] JSON 직렬화/역직렬화는 Go 표준 라이브러리 `encoding/json`을 사용한다
- [A-05] 루프 상태 파일 크기는 최대 1MB를 초과하지 않는다
- [A-06] 단일 시점에 하나의 SPEC에 대해서만 루프가 실행된다 (향후 다중 루프는 확장 사항)
- [A-07] `context.Context`를 통해 모든 장기 실행 작업의 취소와 타임아웃을 제어한다

### 2.3 비즈니스 가정

- [A-08] 수렴(convergence)은 연속 두 번의 반복에서 테스트 실패 수, 린트 오류 수, 커버리지가 개선되지 않을 때 감지된다
- [A-09] Human review가 활성화된 경우, review 단계에서 루프가 일시 정지되어 사용자 승인을 기다린다

---

## 3. Requirements (요구사항)

### 3.1 루프 라이프사이클 관리

#### REQ-LOOP-001: 루프 시작 (Event-Driven)

**WHEN** 사용자가 유효한 SPEC ID로 루프 시작을 요청하면 **THEN** 시스템은 해당 SPEC에 대한 새로운 피드백 루프를 초기화하고, 초기 상태를 `PhaseAnalyze`로 설정하며, 상태를 영속화해야 한다.

#### REQ-LOOP-002: 루프 일시 정지 (Event-Driven)

**WHEN** 실행 중인 루프에 대해 일시 정지가 요청되면 **THEN** 시스템은 현재 단계 작업을 안전하게 중단하고, 현재 상태를 영속화한 후, `Running` 상태를 `false`로 변경해야 한다.

#### REQ-LOOP-003: 루프 재개 (Event-Driven)

**WHEN** 일시 정지된 루프에 대해 재개가 요청되면 **THEN** 시스템은 영속화된 상태를 복원하고, 마지막 저장된 단계부터 실행을 계속해야 한다.

#### REQ-LOOP-004: 루프 취소 (Event-Driven)

**WHEN** 루프 취소가 요청되면 **THEN** 시스템은 현재 실행을 중단하고, 해당 SPEC의 영속화된 상태를 삭제하며, 리소스를 정리해야 한다.

#### REQ-LOOP-005: 루프 상태 조회 (Event-Driven)

**WHEN** 루프 상태 조회가 요청되면 **THEN** 시스템은 현재 SPEC ID, 단계, 반복 횟수, 최대 반복 횟수, 수렴 여부, 실행 여부를 포함하는 읽기 전용 스냅샷을 반환해야 한다.

### 3.2 상태 머신 (State Machine)

#### REQ-SM-001: 단계 전이 (Ubiquitous)

시스템은 **항상** `analyze -> implement -> test -> review` 순서로 단계를 전이해야 한다. review 단계 완료 후에는 다음 반복의 `analyze` 단계로 돌아가야 한다.

#### REQ-SM-002: 단계 전이 유효성 검증 (Ubiquitous)

시스템은 **항상** 현재 단계에서 허용된 다음 단계로만 전이를 허용해야 한다. 정의되지 않은 전이 시도는 오류를 반환해야 한다.

#### REQ-SM-003: 반복 카운터 증가 (Event-Driven)

**WHEN** review 단계가 완료되고 수렴하지 않은 상태에서 다음 analyze 단계로 전이하면 **THEN** 시스템은 반복 카운터를 1 증가시켜야 한다.

### 3.3 수렴 감지 (Convergence Detection)

#### REQ-CONV-001: 자동 수렴 감지 (State-Driven)

**IF** `auto_converge`가 `true`이고 연속 두 번의 반복에서 (테스트 실패 수 변화 없음 AND 린트 오류 수 변화 없음 AND 커버리지 변화 없음) **THEN** 시스템은 루프를 수렴 완료로 표시하고 실행을 종료해야 한다.

#### REQ-CONV-002: 최대 반복 횟수 제한 (State-Driven)

**IF** 현재 반복 횟수가 `max_iterations` (기본값: 5)에 도달하면 **THEN** 시스템은 수렴 여부와 관계없이 루프를 종료하고, 최대 반복 도달 사유를 기록해야 한다.

#### REQ-CONV-003: 완전 성공 수렴 (Event-Driven)

**WHEN** 피드백에서 테스트 실패 수가 0이고, 린트 오류 수가 0이고, 빌드 성공이며, 커버리지가 85% 이상이면 **THEN** 시스템은 즉시 수렴 완료로 판단하고 루프를 종료해야 한다.

### 3.4 상태 영속화 (State Persistence)

#### REQ-PERSIST-001: 상태 저장 (Event-Driven)

**WHEN** 단계 전이가 발생하거나 피드백이 기록되면 **THEN** 시스템은 현재 루프 상태를 `.moai/loop/{specID}.json` 경로에 JSON 형식으로 저장해야 한다.

#### REQ-PERSIST-002: 상태 복원 (Event-Driven)

**WHEN** 루프 재개가 요청되면 **THEN** 시스템은 `.moai/loop/{specID}.json`에서 상태를 읽어 `LoopState` 구조체로 역직렬화하고, 유효성을 검증해야 한다.

#### REQ-PERSIST-003: 상태 삭제 (Event-Driven)

**WHEN** 루프가 정상 종료(수렴 완료 또는 취소)되면 **THEN** 시스템은 해당 SPEC의 영속화된 상태 파일을 삭제해야 한다.

#### REQ-PERSIST-004: 상태 파일 무결성 (Ubiquitous)

시스템은 **항상** `json.Marshal()`을 통해 상태 파일을 생성해야 하며, 문자열 연결을 통한 JSON 생성은 금지된다 (ADR-011 준수).

### 3.5 피드백 수집 및 기록

#### REQ-FB-001: 피드백 구조 (Ubiquitous)

시스템은 **항상** 다음 필드를 포함하는 피드백을 수집해야 한다: 단계(Phase), 반복 횟수(Iteration), 테스트 통과 수(TestsPassed), 테스트 실패 수(TestsFailed), 린트 오류 수(LintErrors), 빌드 성공 여부(BuildSuccess), 커버리지(Coverage), 소요 시간(Duration), 비고(Notes).

#### REQ-FB-002: 피드백 기록 (Event-Driven)

**WHEN** `RecordFeedback`가 호출되면 **THEN** 시스템은 해당 피드백을 현재 루프 상태의 피드백 히스토리에 추가하고, `UpdatedAt` 타임스탬프를 갱신하며, 상태를 영속화해야 한다.

### 3.6 의사결정 엔진 (Decision Engine)

#### REQ-DE-001: 의사결정 생성 (Event-Driven)

**WHEN** review 단계에서 피드백이 수집되면 **THEN** Ralph DecisionEngine은 현재 상태와 피드백을 분석하여 다음 행동(Action), 다음 단계(NextPhase), 수렴 여부(Converged), 사유(Reason)를 포함하는 Decision을 생성해야 한다.

#### REQ-DE-002: 의사결정 행동 유형 (Ubiquitous)

시스템은 **항상** 다음 행동 유형을 지원해야 한다:
- `continue`: 다음 반복을 계속한다
- `converge`: 수렴 완료로 루프를 종료한다
- `request_review`: Human review를 요청하고 루프를 일시 정지한다
- `abort`: 오류로 인해 루프를 중단한다

### 3.7 Human-in-the-Loop

#### REQ-HIL-001: Human Review 브레이크포인트 (State-Driven)

**IF** `human_review`가 `true`이고 현재 단계가 `review`이면 **THEN** 시스템은 루프를 자동으로 일시 정지하고, 사용자에게 검토를 요청해야 한다.

### 3.8 금지 사항 (Unwanted Behavior)

#### REQ-UW-001: 무한 루프 방지 (Unwanted)

시스템은 `max_iterations`를 초과하여 루프를 실행**하지 않아야 한다**.

#### REQ-UW-002: 동시 루프 방지 (Unwanted)

시스템은 동일 SPEC ID에 대해 동시에 두 개 이상의 루프를 실행**하지 않아야 한다**.

#### REQ-UW-003: 손상된 상태 복원 금지 (Unwanted)

시스템은 JSON 파싱에 실패하거나 유효성 검증에 실패한 상태 파일로부터 루프를 복원**하지 않아야 한다**.

---

## 4. Specifications (명세)

### 4.1 파일 구조

```
internal/loop/
    controller.go   -- Controller 인터페이스 구현, 루프 라이프사이클 관리
    feedback.go     -- FeedbackGenerator 인터페이스 구현, 빌드/테스트/린트 결과 수집
    state.go        -- LoopState, LoopStatus, LoopPhase, Feedback, Decision 타입 정의 및 상태 머신 로직
    storage.go      -- Storage 인터페이스 구현, JSON 파일 기반 상태 영속화

internal/ralph/
    engine.go       -- DecisionEngine 인터페이스 구현, 수렴 휴리스틱 및 의사결정 로직
```

### 4.2 인터페이스 정의

```go
// Controller orchestrates the Ralph feedback loop lifecycle.
type Controller interface {
    Start(ctx context.Context, specID string) error
    Pause() error
    Resume(ctx context.Context) error
    Cancel() error
    Status() *LoopStatus
    RecordFeedback(feedback Feedback) error
}

// Storage persists loop state for session resumption.
type Storage interface {
    SaveState(state *LoopState) error
    LoadState(specID string) (*LoopState, error)
    DeleteState(specID string) error
}

// FeedbackGenerator collects feedback from build, test, and lint results.
type FeedbackGenerator interface {
    Collect(ctx context.Context) (*Feedback, error)
}

// DecisionEngine determines the next loop action based on state and feedback.
type DecisionEngine interface {
    Decide(ctx context.Context, state *LoopState, feedback *Feedback) (*Decision, error)
}
```

### 4.3 타입 정의

```go
type LoopPhase string

const (
    PhaseAnalyze   LoopPhase = "analyze"
    PhaseImplement LoopPhase = "implement"
    PhaseTest      LoopPhase = "test"
    PhaseReview    LoopPhase = "review"
)

type LoopState struct {
    SpecID    string     `json:"spec_id"`
    Phase     LoopPhase  `json:"phase"`
    Iteration int        `json:"iteration"`
    MaxIter   int        `json:"max_iterations"`
    Feedback  []Feedback `json:"feedback"`
    StartedAt time.Time  `json:"started_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

type LoopStatus struct {
    SpecID    string
    Phase     LoopPhase
    Iteration int
    MaxIter   int
    Converged bool
    Running   bool
}

type Feedback struct {
    Phase        LoopPhase     `json:"phase"`
    Iteration    int           `json:"iteration"`
    TestsPassed  int           `json:"tests_passed"`
    TestsFailed  int           `json:"tests_failed"`
    LintErrors   int           `json:"lint_errors"`
    BuildSuccess bool          `json:"build_success"`
    Coverage     float64       `json:"coverage"`
    Duration     time.Duration `json:"duration"`
    Notes        string        `json:"notes"`
}

type Decision struct {
    Action    string    `json:"action"`
    NextPhase LoopPhase `json:"next_phase"`
    Converged bool      `json:"converged"`
    Reason    string    `json:"reason"`
}
```

### 4.4 설정 (RalphConfig)

```yaml
ralph:
  max_iterations: 5
  auto_converge: true
  human_review: true
```

### 4.5 상태 영속화 경로

- 저장 경로: `.moai/loop/{specID}.json`
- 파일 형식: `json.MarshalIndent()` 출력
- 파일 생성/삭제: `pkg/utils/file.go`의 안전한 파일 쓰기 유틸리티 사용

### 4.6 상태 전이 다이어그램

```
[Start] --> analyze --> implement --> test --> review --> [Decision]
                                                             |
                                         +-------------------+-------------------+
                                         |                   |                   |
                                    [continue]          [converge]          [request_review]
                                         |                   |                   |
                                    analyze              [Done]              [Pause]
                                  (iteration++)                             (wait for
                                                                            human input)
```

### 4.7 의존성 방향

```
internal/loop/ ----> internal/ralph/
                  -> internal/core/quality/
                  -> internal/core/git/
                  -> internal/config/
                  -> pkg/utils/

internal/ralph/ ---> internal/config/
                  -> pkg/models/
```

### 4.8 비기능 요구사항

| 항목 | 목표 |
|------|------|
| 상태 저장 지연 시간 | < 10ms |
| 상태 복원 지연 시간 | < 10ms |
| 루프 단계 전이 오버헤드 | < 5ms |
| 메모리 사용량 (루프 활성) | < 5MB 추가 |
| 테스트 커버리지 (loop 패키지) | >= 90% |
| 테스트 커버리지 (ralph 패키지) | >= 85% |

---

## 5. Traceability (추적성)

| 요구사항 ID | 파일 | 테스트 시나리오 |
|------------|------|---------------|
| REQ-LOOP-001 | controller.go | AC-START-* |
| REQ-LOOP-002 | controller.go | AC-PAUSE-* |
| REQ-LOOP-003 | controller.go | AC-RESUME-* |
| REQ-LOOP-004 | controller.go | AC-CANCEL-* |
| REQ-LOOP-005 | controller.go | AC-STATUS-* |
| REQ-SM-001 | state.go | AC-SM-* |
| REQ-SM-002 | state.go | AC-SM-* |
| REQ-SM-003 | state.go | AC-SM-* |
| REQ-CONV-001 | engine.go | AC-CONV-* |
| REQ-CONV-002 | controller.go | AC-CONV-* |
| REQ-CONV-003 | engine.go | AC-CONV-* |
| REQ-PERSIST-001 | storage.go | AC-PERSIST-* |
| REQ-PERSIST-002 | storage.go | AC-PERSIST-* |
| REQ-PERSIST-003 | storage.go | AC-PERSIST-* |
| REQ-PERSIST-004 | storage.go | AC-PERSIST-* |
| REQ-FB-001 | feedback.go, state.go | AC-FB-* |
| REQ-FB-002 | controller.go | AC-FB-* |
| REQ-DE-001 | engine.go | AC-DE-* |
| REQ-DE-002 | engine.go | AC-DE-* |
| REQ-HIL-001 | controller.go | AC-HIL-* |
| REQ-UW-001 | controller.go | AC-UW-* |
| REQ-UW-002 | controller.go | AC-UW-* |
| REQ-UW-003 | storage.go | AC-UW-* |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 92.7% (loop) / 100.0% (ralph)

### Summary

Ralph feedback loop system implemented with state machine-based loop controller, persistent storage for session feedback data, and the Ralph decision engine. Controller manages loop iterations with configurable max retries, human-in-the-loop intervention points, and unwinding capabilities for safe loop exit. Feedback module collects and structures user feedback for decision engine input. Ralph engine evaluates feedback patterns to determine loop continuation, retry, or termination decisions. Storage module provides JSON-based persistent state with atomic writes.

### Files Created

- `internal/loop/controller.go`
- `internal/loop/controller_test.go`
- `internal/loop/feedback.go`
- `internal/loop/feedback_test.go`
- `internal/loop/state.go`
- `internal/loop/state_test.go`
- `internal/loop/storage.go`
- `internal/loop/storage_test.go`
- `internal/ralph/engine.go`
- `internal/ralph/engine_test.go`
