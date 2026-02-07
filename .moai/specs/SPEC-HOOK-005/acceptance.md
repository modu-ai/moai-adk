---
spec_id: SPEC-HOOK-005
title: Git Operations Manager - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, git, operations"
---

# SPEC-HOOK-005: Acceptance Criteria

## 1. Command Execution Scenarios

### Scenario 1: Single Command Execution

```gherkin
Feature: Git Command Execution

  Scenario: Execute successful Git command
    Given a Git command for branch: ["branch", "--show-current"]
    When ExecuteCommand is called
    Then GitResult.Success is true
    And GitResult.Stdout contains current branch name
    And GitResult.ReturnCode is 0

  Scenario: Retry failed command
    Given a Git command that fails on first attempt
    And retry count is 2
    When ExecuteCommand is called
    Then the command is retried
    And if retry succeeds, Success is true

  Scenario: Handle command timeout
    Given a Git command that takes longer than timeout
    When ExecuteCommand is called with 10s timeout
    Then context timeout is triggered
    And GitResult.Success is false
    And error message contains "timeout"
```

---

## 2. Parallel Execution Scenarios

### Scenario 2: Parallel Git Operations

```gherkin
Feature: Parallel Execution

  Scenario: Execute multiple commands in parallel
    Given 4 Git commands (branch, log, status, diff)
    And maxWorkers is 4
    When ExecuteParallel is called
    Then all 4 commands execute concurrently
    And total execution time is less than sum of individual times
    And results are returned in original order

  Scenario: Limit concurrent execution with semaphore
    Given maxWorkers is 2
    And 5 Git commands are submitted
    When ExecuteParallel is called
    Then only 2 commands execute at a time
    And remaining commands wait in queue
```

---

## 3. Caching Scenarios

### Scenario 3: Result Caching

```gherkin
Feature: Git Result Caching

  Scenario: Cache successful command result
    Given a Git command with cacheTTL=60
    When ExecuteCommand is called first time
    Then the command is executed
    And result is stored in cache
    And GitResult.Cached is false

  Scenario: Return cached result on subsequent call
    Given a cached result exists for a command
    And cache has not expired
    When ExecuteCommand is called again
    Then Git command is NOT executed
    And cached result is returned
    And GitResult.Cached is true
    And GitResult.CacheHit is true

  Scenario: Invalidate expired cache
    Given a cached result with TTL of 1 second
    And 2 seconds have passed
    When ExecuteCommand is called
    Then cache entry is invalidated
    And command is executed again
```

---

## 4: Cache Management Scenarios

### Scenario 4: Cache Size Management

```gherkin
Feature: Cache Size Limits

  Scenario: Enforce cache size limit
    Given cache size limit is 100
    And cache contains 100 entries
    When a new entry is added
    Then least recently used entry is removed
    And cache size remains at 100

  Scenario: Clear cache by operation type
    Given cache with entries for different operation types
    When ClearCache is called with OpStatus
    Then only status cache entries are removed
    And other entries remain
```

---

## 5. Statistics Scenarios

### Scenario 5: Performance Statistics

```gherkin
Feature: Statistics Tracking

  Scenario: Track operation statistics
    Given multiple Git commands have been executed
    When GetStatistics is called
    Then statistics include:
      | totalOperations | > 0 |
      | cacheHits       | >= 0 |
      | cacheMisses     | >= 0 |
      | errors          | >= 0 |

  Scenario: Calculate cache hit rate
    Given 10 total operations
    And 6 cache hits
    And 4 cache misses
    When GetStatistics is called
    Then cacheHitRate is 0.6 (60%)
```

---

## 6: Project Info Scenarios

### Scenario 6: Project Information

```gherkin
Feature: Project Information

  Scenario: Get comprehensive project info
    Given a Git repository
    When GetProjectInfo is called
    Then ProjectInfo contains:
      | branch      | current branch name |
      | lastCommit  | short commit hash   |
      | commitTime  | relative time       |
      | changes     | number of modified files |

  Scenario: Execute project info commands in parallel
    When GetProjectInfo is called
    Then 4 Git commands execute in parallel
    Commands are: branch, log, log (time), status
    And total time is less than sequential execution
```

---

## 7. Pool Management Scenarios

### Scenario 7: Worker Pool

```gherkin
Feature: Worker Pool Management

  Scenario: Shutdown worker pool gracefully
    Given active workers in the pool
    When Shutdown is called
    Then queued commands are completed
    Then workers are terminated
    And resources are cleaned up

  Scenario: Reuse workers for multiple commands
    Given a worker pool with maxWorkers=4
    When 20 commands are executed sequentially
    Then workers are reused
    And new goroutines are not created for each command
```

---

## 8. Performance Scenarios

### Scenario 8: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: Single command under 2 seconds
    Given a simple Git command (branch, status)
    When ExecuteCommand is executed
    Then execution time is less than 2 seconds

  Scenario: Parallel 4 commands under 3 seconds
    Given 4 Git commands (branch, log, status, diff)
    When ExecuteParallel is executed
    Then total execution time is less than 3 seconds

  Scenario: Cache lookup under 1ms
    Given a cached result exists
    When cache lookup is performed
    Then lookup time is less than 1ms
```

---

## 9. Definition of Done

### 9.1 Code Completion

- [ ] `internal/git/ops/` 패키지 구현 완료
- [ ] Git operations manager 구현
- [ ] Worker pool 구현
- [ ] Result cache 구현
- [ ] Statistics tracker 구현

### 9.2 Test Completion

- [ ] 단위 테스트: 90%+ coverage
- [ ] 명령 실행 테스트
- [ ] 병렬 실행 테스트
- [ ] 캐시 동작 테스트
- [ ] 통계 집계 테스트
- [ ] 경쟁 조건 테스트 (`go test -race`)

### 9.3 Quality Gate

- [ ] `golangci-lint run ./internal/git/ops/...` 오류 0건
- [ ] `go vet ./internal/git/ops/...` 오류 0건
- [ ] `go test -race ./internal/git/ops/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 9.4 Integration

- [ ] SessionStart 훅 통합 확인
- [ ] SessionEnd 훅 통합 확인
- [ ] 프로젝트 정보 표시 확인

### 9.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 캐시 전략 문서화
- [ ] 병렬 실행 전략 문서화
- [ ] 성능 기준 문서화
