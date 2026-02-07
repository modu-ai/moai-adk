---
spec_id: SPEC-LOOP-001
title: Ralph Feedback Loop Engine - Implementation Plan
status: Planned
priority: Medium-High
tags:
  - loop
  - ralph
  - feedback
  - state-machine
  - convergence
  - automation
---

# SPEC-LOOP-001: Implementation Plan

## 1. 개요

Ralph Feedback Loop Engine은 MoAI-ADK의 자율 개발 루프를 제어하는 핵심 모듈이다. 본 계획은 `internal/loop/` (4개 파일)과 `internal/ralph/` (1개 파일)의 구현 전략과 마일스톤을 정의한다.

---

## 2. 마일스톤 (우선순위 기반)

### Primary Goal: 상태 머신 및 타입 시스템 구축

**파일**: `internal/loop/state.go`

**구현 내용**:
- `LoopPhase` 타입 및 상수 정의 (analyze, implement, test, review)
- `LoopState` 구조체 정의 (JSON 직렬화 태그 포함)
- `LoopStatus` 읽기 전용 스냅샷 구조체 정의
- `Feedback` 구조체 정의
- `Decision` 구조체 정의
- 상태 전이 유효성 검증 함수 (`ValidTransition(current, next LoopPhase) bool`)
- 다음 단계 결정 함수 (`NextPhase(current LoopPhase) LoopPhase`)
- 상태 복사 함수 (`ToStatus() *LoopStatus`) -- 스냅샷 생성

**완료 기준**:
- 모든 타입이 정의되고 JSON 직렬화/역직렬화 테스트 통과
- 상태 전이 규칙에 대한 테이블 기반 테스트 작성
- 유효하지 않은 전이에 대한 오류 반환 검증

**의존성**: 없음 (독립 구현 가능)

---

### Secondary Goal: 상태 영속화 구현

**파일**: `internal/loop/storage.go`

**구현 내용**:
- `Storage` 인터페이스 정의
- `FileStorage` 구조체 -- 파일 기반 구현
  - 생성자: `NewFileStorage(baseDir string) *FileStorage`
  - `SaveState(state *LoopState) error` -- `json.MarshalIndent()` 사용, 원자적 파일 쓰기
  - `LoadState(specID string) (*LoopState, error)` -- JSON 파싱 + 유효성 검증
  - `DeleteState(specID string) error` -- 파일 삭제
- 저장 경로: `.moai/loop/{specID}.json`
- 디렉토리 자동 생성 (`os.MkdirAll`)
- 원자적 쓰기 패턴: 임시 파일 -> `os.Rename`

**완료 기준**:
- 저장/로드/삭제 라운드트립 테스트 통과
- 손상된 JSON 파일에 대한 오류 처리 검증
- 존재하지 않는 상태 파일 로드 시 적절한 오류 반환
- 동시 접근 안전성 테스트 (goroutine)

**의존성**: Primary Goal 완료 (타입 정의 필요)

---

### Tertiary Goal: 피드백 수집 시스템 구현

**파일**: `internal/loop/feedback.go`

**구현 내용**:
- `FeedbackGenerator` 인터페이스 정의
- `DefaultFeedbackGenerator` 구조체
  - `quality.TrustChecker` 의존성 주입 (TRUST 5 결과 수집)
  - `Collect(ctx context.Context) (*Feedback, error)` -- 빌드, 테스트, 린트 결과 집계
- 수렴 비교 유틸리티 함수:
  - `IsImproved(prev, curr *Feedback) bool` -- 개선 여부 판단
  - `IsStagnant(prev, curr *Feedback) bool` -- 정체 여부 판단
  - `MeetsQualityGate(fb *Feedback) bool` -- 품질 게이트 충족 여부

**완료 기준**:
- Mock을 사용한 피드백 수집 테스트 통과
- context 취소 시 적절한 중단 검증
- 수렴 비교 함수에 대한 테이블 기반 테스트 작성

**의존성**: Primary Goal 완료 (타입 정의 필요), SPEC-QUALITY-001 (인터페이스 참조)

---

### Quaternary Goal: 의사결정 엔진 구현

**파일**: `internal/ralph/engine.go`

**구현 내용**:
- `DecisionEngine` 인터페이스 정의
- `RalphEngine` 구조체
  - `config.RalphConfig` 의존성 주입
  - `Decide(ctx context.Context, state *LoopState, feedback *Feedback) (*Decision, error)`
- 의사결정 로직:
  1. 최대 반복 횟수 도달 검사 -> `abort` (사유: max iterations reached)
  2. 완전 성공 검사 (실패 0, 오류 0, 빌드 성공, 커버리지 >= 85%) -> `converge`
  3. 자동 수렴 검사 (연속 두 반복 정체) -> `converge` (auto_converge가 true인 경우)
  4. Human review 검사 (human_review가 true이고 review 단계) -> `request_review`
  5. 기본 동작 -> `continue`

**완료 기준**:
- 모든 의사결정 분기에 대한 테이블 기반 테스트 작성
- auto_converge on/off 시나리오 검증
- human_review on/off 시나리오 검증
- 경계 조건 테스트 (반복 횟수 정확히 max인 경우)

**의존성**: Primary Goal 완료 (타입 정의), SPEC-CONFIG-001 (설정 로드)

---

### Final Goal: 루프 컨트롤러 통합 구현

**파일**: `internal/loop/controller.go`

**구현 내용**:
- `Controller` 인터페이스 정의
- `LoopController` 구조체
  - 생성자: `NewLoopController(storage Storage, engine DecisionEngine, feedback FeedbackGenerator, cfg config.RalphConfig) *LoopController`
  - 의존성 주입: Storage, DecisionEngine, FeedbackGenerator, RalphConfig
  - `sync.Mutex`를 사용한 동시 접근 보호
- `Start(ctx context.Context, specID string) error`:
  - 기존 루프 실행 여부 확인 (REQ-UW-002)
  - 초기 LoopState 생성 (Phase: analyze, Iteration: 1)
  - 상태 영속화
  - 루프 실행 goroutine 시작
- `Pause() error`: Running false 설정, context 취소, 상태 저장
- `Resume(ctx context.Context) error`: 상태 로드, Running true 설정, 루프 실행 재개
- `Cancel() error`: context 취소, 상태 삭제
- `Status() *LoopStatus`: 현재 상태의 읽기 전용 스냅샷 반환
- `RecordFeedback(feedback Feedback) error`: 피드백 추가, 상태 저장
- 내부 루프 실행 로직:
  ```
  for iteration <= maxIter {
      for each phase in [analyze, implement, test, review] {
          // execute phase work
          feedback := feedbackGen.Collect(ctx)
          state.Feedback = append(state.Feedback, feedback)
          storage.SaveState(state)

          if phase == review {
              decision := engine.Decide(ctx, state, feedback)
              switch decision.Action {
              case "converge": return (success)
              case "abort": return (error)
              case "request_review": pause and wait
              case "continue": next iteration
              }
          }
          state.Phase = NextPhase(state.Phase)
      }
      state.Iteration++
  }
  ```

**완료 기준**:
- Start -> Pause -> Resume -> 완료 전체 라이프사이클 테스트
- Start -> Cancel 시나리오 테스트
- 수렴 감지 통합 테스트
- 최대 반복 도달 테스트
- 동시 Start 호출 차단 테스트 (REQ-UW-002)
- Human review 일시 정지 테스트

**의존성**: 모든 이전 마일스톤 완료

---

## 3. 기술적 접근

### 3.1 패키지 구조 원칙

- `internal/loop/`: 루프 라이프사이클의 모든 타입, 인터페이스, 구현 포함
- `internal/ralph/`: 의사결정 엔진만 독립 패키지로 분리 (관심사 분리)
- 인터페이스 기반 의존성 주입 (ADR-004 준수)
- Mock 생성: mockery를 사용하여 모든 인터페이스의 Mock 자동 생성

### 3.2 동시성 전략

- `context.Context`: 모든 장기 실행 메서드의 첫 번째 파라미터
- `context.WithCancel`: Pause/Cancel 시 goroutine 안전 종료
- `sync.Mutex`: Controller 내부 상태 보호
- goroutine leak 방지: Controller에 `done` 채널 패턴 적용

### 3.3 오류 처리 전략

- 모든 오류는 `fmt.Errorf("loop: context: %w", err)` 형식으로 래핑
- 센티널 오류 정의:
  - `ErrLoopAlreadyRunning` -- 동일 SPEC에 대한 중복 실행
  - `ErrLoopNotRunning` -- 실행 중이 아닌 루프에 Pause/Resume 시도
  - `ErrLoopNotPaused` -- 일시 정지 상태가 아닌 루프에 Resume 시도
  - `ErrInvalidTransition` -- 유효하지 않은 상태 전이
  - `ErrCorruptedState` -- 손상된 상태 파일 감지
  - `ErrMaxIterationsReached` -- 최대 반복 도달

### 3.4 테스트 전략

- **단위 테스트**: 각 파일별 `*_test.go` 작성 (동일 패키지)
- **테이블 기반 테스트**: Go 관용적 패턴으로 상태 전이, 수렴 감지, 의사결정 분기 검증
- **병렬 테스트**: `t.Parallel()` 활용하여 독립 테스트 병렬 실행
- **Mock 기반 테스트**: mockery로 Storage, DecisionEngine, FeedbackGenerator Mock 생성
- **통합 테스트**: 임시 디렉토리에서 전체 라이프사이클 검증
- **Fuzz 테스트**: JSON 상태 파싱에 대한 fuzz 테스트

### 3.5 아키텍처 설계 방향

```
                  +------------------+
                  |  LoopController  |
                  |  (controller.go) |
                  +--------+---------+
                           |
              +------------+------------+
              |            |            |
      +-------v---+  +----v-----+  +---v---------+
      | Storage   |  | Decision |  | Feedback    |
      | (storage  |  | Engine   |  | Generator   |
      |  .go)     |  | (ralph/  |  | (feedback   |
      |           |  |  engine  |  |  .go)       |
      +-----------+  |  .go)    |  +-------------+
                     +----------+
                           |
                     +-----v------+
                     | RalphConfig|
                     | (config/)  |
                     +------------+
```

---

## 4. 리스크 및 대응 방안

| 리스크 | 영향도 | 대응 방안 |
|-------|--------|---------|
| SPEC-CONFIG-001 미완성으로 RalphConfig 사용 불가 | 높음 | 하드코딩된 기본값으로 초기 구현, config 완성 후 교체 |
| SPEC-QUALITY-001 미완성으로 피드백 수집 불가 | 높음 | FeedbackGenerator 인터페이스로 격리, Mock으로 테스트 |
| 상태 파일 동시 접근 경합 | 중간 | sync.Mutex 적용, 원자적 파일 쓰기 패턴 사용 |
| 루프 goroutine 누수 | 중간 | context.WithCancel + done 채널로 확정적 종료 보장 |
| JSON 상태 파일 손상 | 낮음 | 임시 파일 쓰기 후 rename 패턴, 유효성 검증 |

---

## 5. 의존성 그래프

```
SPEC-CONFIG-001 (config/) ----+
                               |
SPEC-QUALITY-001 (quality/) --+---> SPEC-LOOP-001 (loop/ + ralph/)
                               |
pkg/utils/ (logger, file) ----+
```

---

## 6. 파일별 예상 코드 규모

| 파일 | 예상 LOC | 테스트 LOC |
|------|---------|-----------|
| internal/loop/state.go | ~150 | ~200 |
| internal/loop/storage.go | ~120 | ~180 |
| internal/loop/feedback.go | ~100 | ~150 |
| internal/loop/controller.go | ~350 | ~400 |
| internal/ralph/engine.go | ~200 | ~250 |
| **합계** | **~920** | **~1,180** |

---

## 7. 전문가 상담 권장 사항

### expert-backend 상담 권장

본 SPEC은 상태 머신 설계, 동시성 제어, 파일 기반 영속화를 포함하고 있다. 구현 단계(`/moai run SPEC-LOOP-001`)에서 다음 항목에 대해 expert-backend 상담을 권장한다:

- goroutine 라이프사이클 관리 및 leak 방지 패턴
- 원자적 파일 쓰기 패턴 (임시 파일 + rename)
- context.Context 전파 최적화

### expert-testing 상담 권장

- 상태 머신 전이에 대한 fuzz 테스트 설계
- 동시성 테스트 (`-race` 플래그) 전략
- Mock 생성 및 인터페이스 테스트 패턴
