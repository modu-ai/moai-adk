---
spec_id: SPEC-HOOK-006
title: Resilience Patterns - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, resilience"
---

# SPEC-HOOK-006: Acceptance Criteria

## 1. Circuit Breaker Scenarios

### Scenario 1: Circuit Breaker State Transitions

```gherkin
Feature: Circuit Breaker

  Scenario: Open circuit after threshold failures
    Given a circuit breaker with threshold=5
    And 5 consecutive failures have occurred
    When Call is invoked
    Then the circuit state is Open
    And the function is NOT executed
    And error is returned immediately

  Scenario: Transition to Half-Open after timeout
    Given a circuit in Open state
    And the open timeout has expired
    When Call is invoked
    Then the circuit state is Half-Open
    And a single request is allowed
    And subsequent requests wait

  Scenario: Close circuit after successful request
    Given a circuit in Half-Open state
    And the test request succeeds
    When Call is invoked
    Then the circuit state transitions to Closed
    And normal operation resumes

  Scenario: Reopen circuit on failure in Half-Open
    Given a circuit in Half-Open state
    And the test request fails
    When Call is invoked
    Then the circuit state returns to Open
```

---

## 2. Retry Policy Scenarios

### Scenario 2: Exponential Backoff

```gherkin
Feature: Retry with Exponential Backoff

  Scenario: Retry with exponential backoff
    Given a retry policy with baseDelay=100ms, maxRetries=3
    And a function that fails 2 times then succeeds
    When Retry is called
    Then the function is called 3 times total
    And delays between retries are approximately: 100ms, 200ms
    And the function returns nil (success)

  Scenario: Retry with jitter to prevent thundering herd
    Given a retry policy with useJitter=true
    And baseDelay=100ms
    When multiple goroutines retry simultaneously
    Then retry delays vary by +/- 20%
    And not all retries occur at the same time

  Scenario: Stop retrying after max retries
    Given a retry policy with maxRetries=2
    And a function that always fails
    When Retry is called
    Then the function is called 3 times (initial + 2 retries)
    And the final error is returned
```

---

## 3: Health Check Scenarios

### Scenario 3: Service Health Monitoring

```gherkin
Feature: Health Check

  Scenario: Detect unhealthy service
    Given a health checker for LSP server
    And the LSP server is not responding
    When Check is called
    Then Status returns Unhealthy
    And LastCheck timestamp is updated

  Scenario: Recover to healthy after check passes
    Given a service marked as Unhealthy
    When Check succeeds
    Then Status changes to Healthy
    And circuit breaker can be reset

  Scenario: Periodic health checks
    Given a health checker with interval=30s
    When StartPeriodic is called
    Then Check is executed every 30 seconds
    And service status is updated automatically
```

---

## 4: Resource Monitoring Scenarios

### Scenario 4: Resource Limits

```gherkin
Feature: Resource Monitoring

  Scenario: Monitor memory usage
    Given a resource monitor with memory threshold=80%
    When memory usage exceeds 80%
    Then a warning is logged
    And cache cleanup is triggered

  Scenario: Monitor goroutine count
    Given a resource monitor with goroutine threshold=1000
    When goroutine count exceeds 1000
    Then a warning is logged
    And new requests may be throttled

  Scenario: Get current resource stats
    When GetStats is called
    Then stats include memory, goroutines, and CPU
```

---

## 5: Integration Scenarios

### Scenario 5: LSP Integration

```gherkin
Feature: LSP Resilience

  Scenario: Protect LSP calls with circuit breaker
    Given an LSP server experiencing issues
    When 5 consecutive LSP calls fail
    Then circuit breaker opens
    And subsequent LSP calls fail immediately
    And fallback to CLI tools is used

  Scenario: Retry LSP initialization
    Given an LSP server that temporarily fails
    When LSP initialization is called with retry policy
    Then initialization is retried with exponential backoff
    And if retry succeeds, LSP becomes available
```

---

## 6: Performance Scenarios

### Scenario 6: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: Circuit breaker overhead under 1ms
    Given a closed circuit breaker
    When Call is invoked
    Then execution overhead is less than 1ms

  Scenario: Retry delay calculation under 1ms
    When retry delay is calculated
    Then calculation time is less than 1ms

  Scenario: Health check under 2 seconds
    Given a health check for a service
    When Check is executed
    Then execution time is less than 2 seconds
```

---

## 7: Definition of Done

### 7.1 Code Completion

- [ ] `internal/resilience/` 패키지 구현 완료
- [ ] Circuit breaker 구현
- [ ] Retry policy 구현
- [ ] Health check 구현
- [ ] Resource monitor 구현

### 7.2 Test Completion

- [ ] 단위 테스트: 90%+ coverage
- [ ] 서킷 브레이커 상태 전이 테스트
- [ ] 재시도 정책 테스트
- [ ] 헬스 체크 테스트
- [ ] 자원 모니터링 테스트

### 7.3 Quality Gate

- [ ] `golangci-lint run ./internal/resilience/...` 오류 0건
- [ ] `go vet ./internal/resilience/...` 오류 0건
- [ ] `go test -race ./internal/resilience/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 7.4 Integration

- [ ] LSP 클라이언트와 통합 확인
- [ ] Git operations와 통합 확인
- [ ] Hook system과 통합 확인

### 7.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 서킷 브레이커 사용 가이드
- [ ] 재시도 정책 설정 문서화
- [ ] 헬스 체크 설정 문서화
