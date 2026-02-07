---
spec_id: SPEC-HOOK-005
title: Git Operations Manager - Implementation Plan
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, implementation-plan, git, parallel-execution, caching"
---

# SPEC-HOOK-005: Implementation Plan

## 1. Overview

### 1.1 Scope

Python 기반 Git Operations Manager(~593 LOC)을 Go로 포팅하여 병렬 Git 실행과 캐싱 기능을 제공한다. `internal/git/ops/` 패키지에 4개 파일을 구현하며, Phase 3 Advanced Features의 P2 Medium 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: Cache -> Pool -> Manager -> Stats 순서로 구현한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-HOOK-001    | Completed | Yes     | Hook system integration                    |
| Go 1.22+         | Available | No     | sync, context, os/exec, crypto/md5        |

---

## 2. Task Decomposition

### Milestone 1: Core Infrastructure (Primary Goal)

핵심 인프라 2개 파일 구현. 캐시와 worker pool.

#### Task 1.1: Cache (`cache.go`)

**Priority**: High

**Description**: Git 명령 결과 캐시 구현.

**Implementation Details**:
- `CacheEntry`: result, timestamp, ttl, hitCount
- `Get(key string) (GitResult, bool)`: 캐시 조회
- `Set(key string, result GitResult, ttl int)`: 캐시 저장
- `Cleanup()`: 만료 항목 제거
- `EnforceLimit()`: LRU 제거

**Testing**:
- 캐시 히트/미스 테스트
- TTL 만료 테스트
- 크기 제한 시행 테스트
- 경쟁 조건 테스트

**Covered Requirements**: REQ-HOOK-220, REQ-HOOK-221, REQ-HOOK-222, REQ-HOOK-223

#### Task 1.2: Worker Pool (`pool.go`)

**Priority**: High

**Description**: Git 명령 실행을 위한 worker pool 구현.

**Implementation Details**:
- `worker`: goroutine based worker
- `semaphore`: 최대 동시 실행 제어
- `queue`: 대기 중인 명령
- `shutdown`: 정상 종료

**Testing**:
- 병렬 실행 테스트
- 세마포어 제어 테스트
- worker 재사용 테스트
- shutdown 테스트

**Covered Requirements**: REQ-HOOK-210, REQ-HOOK-211, REQ-HOOK-212, REQ-HOOK-230, REQ-HOOK-231, REQ-HOOK-232

---

### Milestone 2: Manager and Statistics (Secondary Goal)

2개 핵심 파일 구현. 관리자와 통계.

#### Task 2.1: Manager (`manager.go`)

**Priority**: High

**Description**: Git 작업 관리자 구현.

**Implementation Details**:
- `ExecuteCommand()`: 단일 명령 실행 (캐시 체크)
- `ExecuteParallel()`: 병렬 명령 실행
- `GetProjectInfo()`: 프로젝트 정보 조회 (4개 병렬)
- 재시도 로직

**Testing**:
- 명령 실행 성공/실패 테스트
- 재시도 로직 테스트
- 병렬 실행 테스트
- 프로젝트 정보 조회 테스트

**Covered Requirements**: REQ-HOOK-200, REQ-HOOK-201, REQ-HOOK-202, REQ-HOOK-250

#### Task 2.2: Statistics (`stats.go`)

**Priority**: Medium

**Description**: 실행 통계 추적.

**Implementation Details**:
- `RecordOperation()`: 실행 기록
- `GetStatistics()`: 통계 조회
- thread-safe 카운터

**Testing**:
- 통계 집계 테스트
- 캐시 적중률 계산 테스트
- 평균 실행 시간 계산 테스트

**Covered Requirements**: REQ-HOOK-240, REQ-HOOK-241

---

### Milestone 3: Integration (Final Goal)

Hook 통합.

#### Task 3.1: Hook Integration

**Priority**: Medium

**Description**: SessionStart/Stop 훅에서 Git 정보 표시.

**Implementation Details**:
- SessionStart: 프로젝트 정보 조회 및 표시
- SessionEnd: 통계 저장

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/git/ops`      |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `context`         | Cancellation, timeouts                  |
| `crypto/md5`      | Cache key generation                   |
| `encoding/hex`    | Hash to string conversion               |
| `os/exec`         | Git command execution                   |
| `sync`            | WaitGroup, Mutex                       |
| `time`            | Timestamps, durations                   |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **Git 명령 실행 실패**                   | Medium   | Medium     | 재시도 정책, 명확한 오류 메시지                             |
| **캐시 충돌**                          | Low      | Low        | MD5 해시 사용, 고유 키 보장                             |
| **과도한 병렬 실행**                    | Medium   | Low        | 세마포어로 제어                                            |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **Git 버전 호환성**                     | Low      | Low        | 일반적인 Git 명령만 사용                                 |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: internal 함수 호출을 Go 패키지로 변경한다.

### 5.2 Migration Steps

| Step | Action                                    | Verification                              |
|------|-------------------------------------------|-------------------------------------------|
| 1    | Go Git operations 모듈 구현 및 단위 테스트 통과   | `go test ./internal/git/ops/...`           |
| 2    | 캐시 및 병렬 실행 테스트                     | 성능 기준 충족 확인                       |
| 3    | Hook 통합 테스트                            | 프로젝트 정보 표시 확인                     |
| 4    | Python 모듈 제거                            | `lib/git_operations_manager.py` 삭제        |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/git/ops/
    manager.go               # Git operations manager
    manager_test.go
    pool.go                  # Worker pool for parallel execution
    pool_test.go
    cache.go                 # Result caching
    cache_test.go
    stats.go                 # Statistics tracking
    stats_test.go
    types.go                 # Shared types
```

### 6.2 Dependency Flow

```
internal/hook/session_start.go -- uses --> internal/git/ops/manager.go
    |
    v
manager.GetProjectInfo() -- parallel execution --> pool.go
    |
    v
cache.go (result caching)
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                    | Target | Rationale                              |
|--------------------------|--------|----------------------------------------|
| `internal/git/ops/` 전체   | 90%    | 병렬 실행, 경쟁 조건 중요                 |
| Manager                  | 90%    | 핵심 경로 완전 검증                      |
| Cache                    | 95%    | 캐시 로직 완전 검증                       |
| Pool                     | 90%    | goroutine 관련 검증                      |

### 7.2 TRUST 5 Compliance

| Principle   | Git Module Application                                         |
|-------------|---------------------------------------------------------------|
| Tested      | 90%+ coverage, race condition testing                            |
| Readable    | Go naming conventions, godoc comments                          |
| Unified     | gofumpt formatting, golangci-lint compliance                    |
| Secured     | Path validation, command sanitization                          |
| Trackable   | Conventional commits, SPEC-HOOK-005 reference in all commits   |
