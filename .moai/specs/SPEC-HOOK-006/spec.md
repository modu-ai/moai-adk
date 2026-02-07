---
id: SPEC-HOOK-006
title: Resilience Patterns
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P2 Medium
phase: "Phase 3 - Advanced Features"
module: "internal/resilience/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, resilience, circuit-breaker, retry, health-check, P2"
---

# SPEC-HOOK-006: Resilience Patterns

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 외부 서비스 및 도구 호출 시 회복탄력(resilience)을 갖�어야 한다. 이 SPEC은 서킷 브레이커, 재시도 정책, 헬스 체크 패턴을 정의한다.

### 1.2 Problem Statement

외부 서비스 호출(LSP, AST-Grep, CLI 도구 등)은 다음과 같은 실패 모드가 있다:

- **일시적 오류**: 네트워크 타임아웃, 서버 과부하
- **서비스 장애**: LSP 서버 다운, 도구 설치 안됨
- **성능 저하**: 느린 응답 시간
- **자원 고갈**: 너무 많은 동시 요청

### 1.3 Target Module

- **경로**: `internal/resilience/`
- **파일 구성**: `circuit.go`, `retry.go`, `health.go`, `monitor.go`
- **예상 LOC**: ~1,000

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| Go 1.22+         | Runtime  | context, time, sync                           |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | 외부 서비스는 일시적 장애가 발생할 수 있다                  | High       | 재시도 정책 필요                       |
| A-002 | 서킷 브레이커로 연�쇄적 실패를 방지할 수 있다               | High       | 계단적 장애 가능                       |
| A-003 | 헬스 체크로 서비스 상태를 감지할 수 있다                  | Medium     | 감지 불능시 false positive 가능        |
| A-004 | 지수 백오프는 서버 과부하를 악화시킬 수 있다              | High       | 백오프 간격 조정 필요                  |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 사용자는 빠른 실패(fail-fast)보다는 graceful degradation 선호 | High       | 장애 시 전체 중단 가능                  |
| B-002 | 자동 복구는 사용자 개입 필요성을 줄여준다                 | High       | 수동 개입 빈도 증가                     |

---

## 3. Requirements (R)

### Module 1: Circuit Breaker (서킷 브레이커)

**REQ-HOOK-300** [Ubiquitous]
시스템은 **항상** 외부 서비스 호출에 서킷 브레이커를 적용해야 한다.

- `CircuitBreaker`: Open, Half-Open, Closed 상태
- `CallWithBreaker()`: 서킷 브레이커를 통한 호출
- `threshold`: 실패 임계값 (기본 5회 연속 실패)
- `timeout`: Open 상태 지속 시간 (기본 30초)

**REQ-HOOK-301** [Event-Driven]
**WHEN** 서비스가 threshold 이상 연속 실패하면 **THEN** 서킷 브레이커는 Open 상태로 전환하고 즉시 실패를 반환해야 한다.

**REQ-HOOK-302** [State-Driven]
**IF** Open 상태에서 timeout이 경과하면 **THEN** 서킷 브레이커는 Half-Open 상태로 전환하고 단일 요청을 허용해야 한다.

**REQ-HOOK-303** [Unwanted]
시스템은 Open 상태일 때 외부 서비스를 호출**하지 않아야 한다**. 모든 요청은 즉시 실패해야 한다.

### Module 2: Retry Policy (재시도 정책)

**REQ-HOOK-310** [Ubiquitous]
시스템은 **항상** 일시적 오류에 대해 재시도를 수행해야 한다.

- `RetryPolicy`: maxRetries, backoff, timeout
- `ExponentialBackoff`: 지수 백오프 (기본 100ms * 2^attempt)
- `Jitter`: 지터 추가로 썬더 현상 방지

**REQ-HOOK-311** [Event-Driven]
**WHEN** 요청이 실패하고 재시도 가능한 오류면 **THEN** 시스템은 백오프 후 재시도해야 한다.

**REQ-HOOK-312** [State-Driven]
**IF** 모든 재시도가 실패하면 **THEN** 시스템은 최종 오류를 반환해야 한다.

**REQ-HOOK-313** [Unwanted]
시스템은 클라이언트 오류(4xx, invalid input)에 대해 재시도**하지 않아야 한다**.

### Module 3: Health Check (헬스 체크)

**REQ-HOOK-320** [Ubiquitous]
시스템은 **항상** 외부 서비스의 건전성을 주기적으로 확인해야 한다.

- `HealthChecker`: 서비스별 헬스 체크
- `Check()`: 헬스 상태 조회
- `interval`: 헬스 체크 주기 (기본 30초)

**REQ-HOOK-321** [Event-Driven]
**WHEN** 헬스 체크가 실패하면 **THEN** 시스템은 해당 서비스를 unhealthy로 표시하고 서킷 브레이커를 열어야 한다.

**REQ-HOOK-322** [State-Driven]
**IF** 헬스 체크가 성공하면 **THEN** 시스템은 해당 서비스를 healthy로 표시하고 서킷 브레이커를 닫을 수 있어야 한다.

### Module 4: Resource Monitoring (자원 모니터링)

**REQ-HOOK-330** [Ubiquitous]
시스템은 **항상** 시스템 자원 사용량을 모니터링해야 한다.

- `ResourceMonitor`: CPU, 메모리, goroutine 수
- `GetStats()`: 현재 자원 상태
- `threshold`: 경고 임계값

**REQ-HOOK-331** [Event-Driven]
**WHEN** 메모리 사용량이 임계값(80%)을 초과하면 **THEN** 시스템은 경고를 로그하고 캐시를 정리해야 한다.

**REQ-HOOK-332** [State-Driven]
**IF** goroutine 수가 임계값(1000)을 초과하면 **THEN** 시스템은 새 요청을 제한해야 한다.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// CircuitState represents the state of a circuit breaker.
type CircuitState string

const (
    StateClosed   CircuitState = "closed"   // Normal operation
    StateOpen     CircuitState = "open"     // Failing fast
    StateHalfOpen CircuitState = "half-open" // Testing recovery
)

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker interface {
    Call(ctx context.Context, fn func() error) error
    State() CircuitState
    Reset()
}

// RetryPolicy defines retry behavior.
type RetryPolicy struct {
    MaxRetries    int           `json:"maxRetries"`
    BaseDelay     time.Duration `json:"baseDelay"`
    MaxDelay      time.Duration `json:"maxDelay"`
    UseJitter     bool          `json:"useJitter"`
    RetryableErrors []error     `json:"retryableErrors"`
}

// Retry executes a function with retry policy.
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error

// HealthStatus represents service health.
type HealthStatus string

const (
    StatusHealthy   HealthStatus = "healthy"
    StatusUnhealthy HealthStatus = "unhealthy"
    StatusUnknown   HealthStatus = "unknown"
)

// HealthChecker checks service health.
type HealthChecker interface {
    Check(ctx context.Context) HealthStatus
    LastCheck() time.Time
    Status() HealthStatus
}

// ResourceStats represents system resource usage.
type ResourceStats struct {
    MemoryUsedMB  uint64  `json:"memoryUsedMB"`
    MemoryTotalMB uint64  `json:"memoryTotalMB"`
    GoroutineCount int    `json:"goroutineCount"`
    CPUPercent    float64 `json:"cpuPercent"`
}

// ResourceMonitor tracks system resources.
type ResourceMonitor interface {
    GetStats() ResourceStats
    StartMonitoring(ctx context.Context, interval time.Duration)
}
```

### 4.2 Circuit Breaker State Machine

```
Closed --(failures >= threshold)--> Open
  ^                                |
  |                                v
Half-Open <--(timeout expired)---- Open
  |                                |
  +---(success)---------------------+
```

### 4.3 Retry Policy Configuration

| Service Type      | MaxRetries | BaseDelay | MaxDelay |
|-------------------|------------|-----------|----------|
| LSP Server        | 2          | 100ms     | 5s       |
| AST-Grep          | 1          | 200ms     | 5s       |
| CLI Tools         | 1          | 50ms      | 1s       |
| Git Operations    | 2          | 100ms     | 3s       |

### 4.4 Health Check Endpoints

| Service Type      | Check Method                | Interval |
|-------------------|----------------------------|----------|
| LSP Server        | initialize request          | 30s      |
| AST-Grep          | sg --version               | 60s      |
| Git               | git rev-parse --git-dir    | N/A      |

### 4.5 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| 서킷 브레이커 오버헤드           | < 1ms     | Benchmark test                         |
| 재시도 지연 계산                 | < 1ms     | Benchmark test                         |
| 헬스 체크 실행                  | < 2s      | End-to-end test                         |
| 자원 모니터링 오버헤드           | < 5%      | Runtime profiling                      |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-300~303 | `circuit.go`                  |
| REQ-HOOK-310~313 | `retry.go`                    |
| REQ-HOOK-320~322 | `health.go`                   |
| REQ-HOOK-330~332 | `monitor.go`                  |

### 5.2 Integration Points

- **SPEC-HOOK-001**: Hook system integration
- **SPEC-HOOK-004**: LSP diagnostics (circuit breaker, retry)
- **SPEC-HOOK-005**: Git operations (retry)
- **internal/lsp/**: LSP client (health check)

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 3 - Advanced Features

### Summary

Resilience patterns for external service integration. Circuit breaker prevents cascading failures, exponential backoff retry handles transient errors, health checks enable automatic recovery detection, and resource monitoring prevents system overload. Integrates with all hook modules that call external services.

### Python Reference

- `lib/unified_timeout_manager.py` (659 LOC - partial overlap)

### Estimated LOC

- `circuit.go`: ~300 LOC
- `retry.go`: ~200 LOC
- `health.go`: ~200 LOC
- `monitor.go`: ~150 LOC
- Test files: ~300 LOC
- Total: ~1,150 LOC
