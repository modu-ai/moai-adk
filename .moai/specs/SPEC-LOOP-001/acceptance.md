---
spec_id: SPEC-LOOP-001
title: Ralph Feedback Loop Engine - Acceptance Criteria
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

# SPEC-LOOP-001: Acceptance Criteria

## 1. 루프 시작/정지/재개/취소

### AC-START-001: 정상 루프 시작

```gherkin
Given 유효한 SPEC ID "SPEC-TEST-001"이 존재하고
  And 해당 SPEC에 대해 실행 중인 루프가 없을 때
When Controller.Start(ctx, "SPEC-TEST-001")을 호출하면
Then LoopState가 생성되어야 한다
  And Phase는 "analyze"여야 한다
  And Iteration은 1이어야 한다
  And MaxIter는 RalphConfig.MaxIterations (기본값 5)여야 한다
  And 상태가 ".moai/loop/SPEC-TEST-001.json"에 저장되어야 한다
  And Status().Running이 true여야 한다
```

### AC-START-002: 이미 실행 중인 루프에 대한 시작 시도

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 이미 루프가 실행 중일 때
When Controller.Start(ctx, "SPEC-TEST-001")을 호출하면
Then ErrLoopAlreadyRunning 오류가 반환되어야 한다
  And 기존 루프 상태가 변경되지 않아야 한다
```

### AC-PAUSE-001: 실행 중인 루프 일시 정지

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 실행 중이고
  And 현재 Phase가 "implement"일 때
When Controller.Pause()를 호출하면
Then Status().Running이 false여야 한다
  And 현재 상태가 영속화되어야 한다
  And Status().Phase는 "implement"여야 한다 (변경되지 않음)
```

### AC-PAUSE-002: 실행 중이 아닌 루프 일시 정지 시도

```gherkin
Given 실행 중인 루프가 없을 때
When Controller.Pause()를 호출하면
Then ErrLoopNotRunning 오류가 반환되어야 한다
```

### AC-RESUME-001: 일시 정지된 루프 재개

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 일시 정지 상태이고
  And Phase가 "implement"이고 Iteration이 2일 때
When Controller.Resume(ctx)를 호출하면
Then Status().Running이 true여야 한다
  And Status().Phase가 "implement"여야 한다
  And Status().Iteration이 2여야 한다
  And 루프가 "implement" 단계부터 실행을 계속해야 한다
```

### AC-RESUME-002: 세션 재시작 후 루프 재개

```gherkin
Given ".moai/loop/SPEC-TEST-001.json"에 유효한 상태 파일이 존재하고
  And Phase가 "test"이고 Iteration이 3일 때
When 새로운 Controller 인스턴스를 생성하고
  And Controller.Resume(ctx)를 호출하면
Then 상태 파일에서 LoopState가 복원되어야 한다
  And Status().Phase가 "test"여야 한다
  And Status().Iteration이 3이어야 한다
  And Feedback 히스토리가 보존되어야 한다
```

### AC-RESUME-003: 일시 정지 상태가 아닌 루프 재개 시도

```gherkin
Given 루프가 실행 중이거나 존재하지 않을 때
When Controller.Resume(ctx)를 호출하면
Then ErrLoopNotPaused 오류가 반환되어야 한다
```

### AC-CANCEL-001: 실행 중인 루프 취소

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 실행 중일 때
When Controller.Cancel()을 호출하면
Then Status().Running이 false여야 한다
  And ".moai/loop/SPEC-TEST-001.json" 파일이 삭제되어야 한다
  And 실행 중인 goroutine이 정리되어야 한다
```

### AC-CANCEL-002: 일시 정지된 루프 취소

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 일시 정지 상태일 때
When Controller.Cancel()을 호출하면
Then ".moai/loop/SPEC-TEST-001.json" 파일이 삭제되어야 한다
  And Status()가 빈 상태를 반환해야 한다
```

### AC-STATUS-001: 루프 상태 조회

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 실행 중이고
  And Phase가 "test"이고 Iteration이 2일 때
When Controller.Status()를 호출하면
Then LoopStatus.SpecID가 "SPEC-TEST-001"이어야 한다
  And LoopStatus.Phase가 "test"여야 한다
  And LoopStatus.Iteration이 2여야 한다
  And LoopStatus.MaxIter가 5여야 한다
  And LoopStatus.Converged가 false여야 한다
  And LoopStatus.Running이 true여야 한다
```

---

## 2. 상태 머신 전이

### AC-SM-001: 정상 단계 전이 순서

```gherkin
Given 현재 Phase가 "analyze"일 때
When NextPhase()를 호출하면
Then "implement"가 반환되어야 한다

Given 현재 Phase가 "implement"일 때
When NextPhase()를 호출하면
Then "test"가 반환되어야 한다

Given 현재 Phase가 "test"일 때
When NextPhase()를 호출하면
Then "review"가 반환되어야 한다

Given 현재 Phase가 "review"일 때
When NextPhase()를 호출하면
Then "analyze"가 반환되어야 한다 (다음 반복)
```

### AC-SM-002: 유효한 전이 검증

```gherkin
Given 현재 Phase가 "analyze"일 때
When ValidTransition("analyze", "implement")를 호출하면
Then true가 반환되어야 한다

Given 현재 Phase가 "analyze"일 때
When ValidTransition("analyze", "review")를 호출하면
Then false가 반환되어야 한다 (건너뛰기 금지)
```

### AC-SM-003: 반복 카운터 증가

```gherkin
Given Phase가 "review"이고 Iteration이 1이고
  And Decision.Action이 "continue"일 때
When 다음 "analyze" 단계로 전이하면
Then Iteration이 2로 증가해야 한다
  And Phase가 "analyze"여야 한다
```

---

## 3. 수렴 감지

### AC-CONV-001: 자동 수렴 감지 (정체 상태)

```gherkin
Given auto_converge가 true이고
  And 반복 1의 피드백이 {TestsFailed: 2, LintErrors: 1, Coverage: 78.5}이고
  And 반복 2의 피드백이 {TestsFailed: 2, LintErrors: 1, Coverage: 78.5}일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Converged가 true여야 한다
  And Decision.Action이 "converge"여야 한다
  And Decision.Reason에 "stagnant" 또는 "no improvement" 관련 메시지가 포함되어야 한다
```

### AC-CONV-002: 자동 수렴 비활성화 시 정체 무시

```gherkin
Given auto_converge가 false이고
  And 연속 두 반복의 피드백이 동일할 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Converged가 false여야 한다
  And Decision.Action이 "continue"여야 한다
```

### AC-CONV-003: 최대 반복 횟수 도달

```gherkin
Given max_iterations가 5이고
  And 현재 Iteration이 5이고
  And Phase가 "review"일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "abort"여야 한다
  And Decision.Reason에 "max iterations reached" 관련 메시지가 포함되어야 한다
```

### AC-CONV-004: 최대 반복 1회 전 정상 계속

```gherkin
Given max_iterations가 5이고
  And 현재 Iteration이 4이고
  And 개선이 감지되었을 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "continue"여야 한다
  And Decision.NextPhase가 "analyze"여야 한다
```

### AC-CONV-005: 완전 성공으로 즉시 수렴

```gherkin
Given 피드백이 {TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 92.3}일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Converged가 true여야 한다
  And Decision.Action이 "converge"여야 한다
  And Decision.Reason에 "quality gate satisfied" 관련 메시지가 포함되어야 한다
```

### AC-CONV-006: 커버리지 85% 미만 시 계속

```gherkin
Given 피드백이 {TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 82.0}일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Converged가 false여야 한다
  And Decision.Action이 "continue"여야 한다
```

---

## 4. 상태 영속화

### AC-PERSIST-001: 상태 저장 및 로드 라운드트립

```gherkin
Given LoopState가 {SpecID: "SPEC-TEST-001", Phase: "test", Iteration: 3}일 때
When Storage.SaveState(state)를 호출하고
  And Storage.LoadState("SPEC-TEST-001")을 호출하면
Then 로드된 상태의 SpecID가 "SPEC-TEST-001"이어야 한다
  And Phase가 "test"여야 한다
  And Iteration이 3이어야 한다
  And Feedback 히스토리가 보존되어야 한다
```

### AC-PERSIST-002: 상태 파일 경로

```gherkin
Given baseDir가 ".moai/loop/"일 때
When Storage.SaveState(state)를 호출하면
Then ".moai/loop/SPEC-TEST-001.json" 파일이 생성되어야 한다
  And 파일 내용이 유효한 JSON이어야 한다
  And json.Valid() 검증을 통과해야 한다
```

### AC-PERSIST-003: 존재하지 않는 상태 파일 로드

```gherkin
Given ".moai/loop/SPEC-NONEXIST-001.json" 파일이 존재하지 않을 때
When Storage.LoadState("SPEC-NONEXIST-001")을 호출하면
Then nil 상태와 적절한 오류가 반환되어야 한다
  And 오류가 os.ErrNotExist를 래핑해야 한다
```

### AC-PERSIST-004: 손상된 상태 파일 로드 거부

```gherkin
Given ".moai/loop/SPEC-TEST-001.json"에 유효하지 않은 JSON이 저장되어 있을 때
When Storage.LoadState("SPEC-TEST-001")을 호출하면
Then nil 상태와 ErrCorruptedState 오류가 반환되어야 한다
  And 루프가 복원되지 않아야 한다
```

### AC-PERSIST-005: 상태 삭제

```gherkin
Given ".moai/loop/SPEC-TEST-001.json" 파일이 존재할 때
When Storage.DeleteState("SPEC-TEST-001")을 호출하면
Then 파일이 삭제되어야 한다
  And Storage.LoadState("SPEC-TEST-001") 호출 시 오류가 반환되어야 한다
```

### AC-PERSIST-006: 원자적 파일 쓰기

```gherkin
Given 상태 저장 중 시스템 장애가 발생할 때
When Storage.SaveState(state)가 중간에 실패하면
Then 이전 상태 파일이 손상되지 않아야 한다
  And 임시 파일이 정리되어야 한다
```

### AC-PERSIST-007: 디렉토리 자동 생성

```gherkin
Given ".moai/loop/" 디렉토리가 존재하지 않을 때
When Storage.SaveState(state)를 호출하면
Then 디렉토리가 자동으로 생성되어야 한다
  And 상태 파일이 성공적으로 저장되어야 한다
```

---

## 5. 피드백 수집 및 기록

### AC-FB-001: 피드백 기록

```gherkin
Given 루프가 실행 중이고 Phase가 "test"일 때
When Controller.RecordFeedback(Feedback{
    Phase: "test", Iteration: 1,
    TestsPassed: 42, TestsFailed: 3,
    LintErrors: 1, BuildSuccess: true,
    Coverage: 81.5
  })를 호출하면
Then LoopState.Feedback 슬라이스에 해당 피드백이 추가되어야 한다
  And LoopState.UpdatedAt가 현재 시간으로 갱신되어야 한다
  And 상태가 영속화되어야 한다
```

### AC-FB-002: 피드백 히스토리 보존

```gherkin
Given 반복 1, 2, 3의 피드백이 기록되었을 때
When Status()를 조회하면
Then 모든 3개의 피드백이 순서대로 보존되어야 한다
  And 각 피드백의 Iteration이 1, 2, 3이어야 한다
```

### AC-FB-003: 개선 감지

```gherkin
Given 이전 피드백이 {TestsFailed: 5, LintErrors: 3, Coverage: 72.0}이고
  And 현재 피드백이 {TestsFailed: 2, LintErrors: 1, Coverage: 80.0}일 때
When IsImproved(prev, curr)를 호출하면
Then true가 반환되어야 한다
```

### AC-FB-004: 정체 감지

```gherkin
Given 이전 피드백이 {TestsFailed: 2, LintErrors: 1, Coverage: 80.0}이고
  And 현재 피드백이 {TestsFailed: 2, LintErrors: 1, Coverage: 80.0}일 때
When IsStagnant(prev, curr)를 호출하면
Then true가 반환되어야 한다
```

### AC-FB-005: 품질 게이트 충족

```gherkin
Given 피드백이 {TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 87.5}일 때
When MeetsQualityGate(feedback)를 호출하면
Then true가 반환되어야 한다

Given 피드백이 {TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 82.0}일 때
When MeetsQualityGate(feedback)를 호출하면
Then false가 반환되어야 한다 (커버리지 85% 미만)
```

---

## 6. 의사결정 엔진

### AC-DE-001: continue 결정

```gherkin
Given 반복 1이고 max_iterations가 5이고
  And 피드백에서 개선이 감지되었을 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "continue"여야 한다
  And Decision.NextPhase가 "analyze"여야 한다
  And Decision.Converged가 false여야 한다
```

### AC-DE-002: converge 결정

```gherkin
Given 피드백이 {TestsFailed: 0, LintErrors: 0, BuildSuccess: true, Coverage: 90.0}일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "converge"여야 한다
  And Decision.Converged가 true여야 한다
```

### AC-DE-003: request_review 결정

```gherkin
Given human_review가 true이고
  And 현재 Phase가 "review"이고
  And 품질 게이트가 충족되지 않았을 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "request_review"여야 한다
  And Decision.Converged가 false여야 한다
```

### AC-DE-004: abort 결정 (최대 반복)

```gherkin
Given max_iterations가 5이고
  And 현재 Iteration이 5일 때
When DecisionEngine.Decide()를 호출하면
Then Decision.Action이 "abort"여야 한다
  And Decision.Reason에 "max iterations" 관련 메시지가 포함되어야 한다
```

---

## 7. Human-in-the-Loop

### AC-HIL-001: Human Review 브레이크포인트

```gherkin
Given human_review가 true이고
  And 루프가 "review" 단계에 도달했을 때
When DecisionEngine이 "request_review"를 결정하면
Then 루프가 자동으로 일시 정지되어야 한다
  And Status().Running이 false여야 한다
  And Status().Phase가 "review"여야 한다
  And 상태가 영속화되어야 한다
```

### AC-HIL-002: Human Review 후 재개

```gherkin
Given human_review로 인해 루프가 일시 정지된 상태일 때
When 사용자가 Controller.Resume(ctx)를 호출하면
Then 루프가 다음 반복의 "analyze" 단계부터 계속되어야 한다
  And Iteration이 1 증가해야 한다
```

### AC-HIL-003: Human Review 비활성화

```gherkin
Given human_review가 false이고
  And 루프가 "review" 단계에 도달했을 때
When DecisionEngine이 결정을 내리면
Then "request_review" 결정이 생성되지 않아야 한다
  And 루프가 자동으로 다음 반복을 계속해야 한다
```

---

## 8. 금지 사항 (Unwanted Behavior)

### AC-UW-001: 무한 루프 방지

```gherkin
Given max_iterations가 3일 때
When 루프가 3번째 반복을 완료하면
Then 4번째 반복이 시작되지 않아야 한다
  And 루프가 종료되어야 한다
```

### AC-UW-002: 동시 루프 방지

```gherkin
Given SPEC ID "SPEC-TEST-001"에 대해 루프가 실행 중일 때
When 다른 goroutine에서 Controller.Start(ctx, "SPEC-TEST-001")을 호출하면
Then ErrLoopAlreadyRunning 오류가 반환되어야 한다
  And 기존 루프가 영향받지 않아야 한다
```

### AC-UW-003: 손상된 상태 복원 거부

```gherkin
Given ".moai/loop/SPEC-TEST-001.json"에 잘못된 JSON이 저장되어 있을 때
When Controller.Resume(ctx)를 호출하면
Then ErrCorruptedState 오류가 반환되어야 한다
  And 루프가 시작되지 않아야 한다
```

---

## 9. 비기능 검증

### AC-PERF-001: 상태 저장 성능

```gherkin
Given 5개의 피드백 항목을 포함하는 LoopState가 존재할 때
When Storage.SaveState(state)를 호출하면
Then 10ms 이내에 완료되어야 한다
```

### AC-PERF-002: 상태 복원 성능

```gherkin
Given ".moai/loop/SPEC-TEST-001.json"에 유효한 상태 파일이 존재할 때
When Storage.LoadState("SPEC-TEST-001")을 호출하면
Then 10ms 이내에 완료되어야 한다
```

### AC-PERF-003: Context 취소 응답성

```gherkin
Given 루프가 실행 중일 때
When context가 취소되면
Then 1초 이내에 루프가 정리되고 종료되어야 한다
  And goroutine이 누수되지 않아야 한다
```

---

## 10. Definition of Done

- [ ] `internal/loop/state.go` -- 모든 타입 정의 및 상태 전이 로직 구현
- [ ] `internal/loop/storage.go` -- Storage 인터페이스 및 FileStorage 구현
- [ ] `internal/loop/feedback.go` -- FeedbackGenerator 및 수렴 비교 유틸리티 구현
- [ ] `internal/loop/controller.go` -- Controller 인터페이스 및 LoopController 구현
- [ ] `internal/ralph/engine.go` -- DecisionEngine 인터페이스 및 RalphEngine 구현
- [ ] 모든 테스트 파일 작성 및 통과 (`go test -race ./internal/loop/... ./internal/ralph/...`)
- [ ] `internal/loop/` 패키지 테스트 커버리지 >= 90%
- [ ] `internal/ralph/` 패키지 테스트 커버리지 >= 85%
- [ ] golangci-lint 경고 0건
- [ ] go vet 오류 0건
- [ ] 전체 라이프사이클 통합 테스트 통과 (Start -> feedback -> converge)
- [ ] 세션 재시작 시 상태 복원 통합 테스트 통과
- [ ] goroutine 누수 테스트 통과 (goleak 또는 runtime.NumGoroutine 검증)
