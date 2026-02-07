---
id: SPEC-HOOK-005
title: Git Operations Manager
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P2 Medium
phase: "Phase 3 - Advanced Features"
module: "internal/git/ops/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, git, parallel-execution, caching, connection-pooling, P2"
---

# SPEC-HOOK-005: Git Operations Manager

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 Claude Code와 통합하여 Git 작업을 최적화한다. 이 SPEC은 병렬 Git 명령 실행, 결과 캐싱, 연결 풀링을 제공하는 Git Operations Manager를 정의한다.

### 1.2 Problem Statement

Python 기반 MoAI-ADK의 `lib/git_operations_manager.py` 모듈(~593 LOC)은 최적화된 Git 작업을 제공하지만 다음과 같은 문제가 있다:

- **직렬 실행**: Git 명령이 순차적으로 실행되어 성능 저하
- **중복 실행**: 동일한 명령이 반복 실행되어 자원 낭비
- **연결 비용**: 매번 새 프로세스 생성으로 오버헤드
- **제한된 통계**: 실행 횟수, 캐시 적중률 등의 통계 부족

### 1.3 Target Module

- **경로**: `internal/git/ops/`
- **파일 구성**: `manager.go`, `pool.go`, `cache.go`, `stats.go`
- **예상 LOC**: ~1,200

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| Go 1.22+         | Runtime  | context, os/exec, crypto/md5, sync, time      |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | git 바이너리는 시스템 PATH에 존재한다                        | High       | 기능 완전 불가                         |
| A-002 | Go 1.22+ goroutine과 WaitGroup로 병렬 실행 가능하다            | High       | 병렬 실행 불가로 성능 저하               |
| A-003 | 캐시 키는 명령어 인자와 작업 디렉터리로 생성할 수 있다        | High       | 캐시 충돌 가능                         |
| A-004 | Git 명령은 멱등적(idempotent)으로 간주할 수 있다              | Medium     | 중복 실행으로 부정확한 결과             |
| A-005 | 세마포어로 최대 동시 실행 수를 제어할 수 있다                 | High       | 리소스 고갈 가능                       |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 프로젝트는 Git으로 버전 관리된다                             | High       | Git 기능 사용 불가                     |
| B-002 | 병렬 실행으로 성능 향상이 기대된다                           | High       | 성능 기준 미달                         |

---

## 3. Requirements (R)

### Module 1: Command Execution (명령 실행)

**REQ-HOOK-200** [Ubiquitous]
시스템은 **항상** Git 명령을 `exec.Command`로 실행하고 결과를 반환해야 한다.

- `ExecuteCommand(cmd GitCommand) GitResult`: 단일 명령 실행
- `ExecuteParallel(cmds []GitCommand) []GitResult`: 병렬 명령 실행
- 타임아웃: 기본 10초, 설정 가능

**REQ-HOOK-201** [Event-Driven]
**WHEN** 명령이 실패하면 **THEN** 시스템은 재시도 정책에 따라 재시도해야 한다(기본 2회).

**REQ-HOOK-202** [State-Driven]
**IF** 모든 재시도가 실패하면 **THEN** 시스템은 마지막 오류를 반환하고 실패를 로그해야 한다.

### Module 2: Parallel Execution (병렬 실행)

**REQ-HOOK-210** [Ubiquitous]
시스템은 **항상** 세마포어로 동시 Git 작업 수를 제한해야 한다(기본 4개).

- `maxWorkers`: 최대 동시 작업 수
- `semaphore`: 실행 제어
- `worker pool`: goroutine 풀

**REQ-HOOK-211** [Event-Driven]
**WHEN** 다중 Git 명령이 제출되면 **THEN** 시스템은 최대 maxWorkers만큼 병렬로 실행해야 한다.

**REQ-HOOK-212** [State-Driven]
**IF** 명령이 완료되면 **THEN** 세마포어를 해제하고 다음 명령을 실행해야 한다.

### Module 3: Result Caching (결과 캐싱)

**REQ-HOOK-220** [Ubiquitous]
시스템은 **항상** Git 명령 결과를 캐시하여 중복 실행을 방지해야 한다.

- `cache`: map[string]CacheEntry
- `cacheKey`: operation_type + args + cwd + branch
- `ttl`: 기본 60초

**REQ-HOOK-221** [Event-Driven]
**WHEN** 캐시 히트가 발생하면 **THEN** 시스템은 명령을 실행하지 않고 캐시된 결과를 반환해야 한다.

**REQ-HOOK-222** [State-Driven]
**IF** 캐시가 만료되었으면 **THEN** 시스템은 만료 항목을 삭제하고 명령을 다시 실행해야 한다.

**REQ-HOOK-223** [Unwanted]
시스템은 캐시 크기를 캐시_size_limit(기본 100) 이상으로 증가시키지 않아야 한다. LRU 정책으로 오래된 항목을 제거해야 한다.

### Module 4: Connection Pooling (연결 풀링)

**REQ-HOOK-230** [Ubiquitous]
시스템은 **항상** Git 명령 실행을 위한 worker pool을 유지해야 한다.

- `executor`: ThreadPoolExecutor (고정 크기)
- `queue`: 대기 중인 명령 큐
- `shutdown`: 정상 종료 지원

**REQ-HOOK-231** [Event-Driven]
**WHEN** 명령이 큐에 제출되면 **THEN** 시스템은 사용 가능한 worker에 할당해야 한다.

**REQ-HOOK-232** [State-Driven]
**IF** 모든 worker가 사용 중이면 **THEN** 명령은 큐에서 대기해야 한다.

### Module 5: Statistics Tracking (통계 추적)

**REQ-HOOK-240** [Ubiquitous]
시스템은 **항상** 실행 통계를 추적해야 한다.

- `totalOperations`: 총 실행 횟수
- `cacheHits`: 캐시 적중 횟수
- `cacheMisses`: 캐시 미스 횟수
- `errors`: 오류 횟수
- `totalTime`: 총 실행 시간

**REQ-HOOK-241** [Event-Driven]
**WHEN** GetStatistics가 호출되면 **THEN** 시스템은 집계된 통계와 캐시 적중률을 반환해야 한다.

### Module 6: Project Info (프로젝트 정보)

**REQ-HOOK-250** [Ubiquitous]
시스템은 **항상** 프로젝트 Git 정보를 효율적으로 조회해야 한다.

- `GetProjectInfo() ProjectInfo`: 브랜치, 마지막 커밋, 상태 등
- 병렬로 4개 명령 실행(branch, log, status)

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// GitOperationType represents the type of Git operation.
type GitOperationType string

const (
    OpBranch   GitOperationType = "branch"
    OpCommit   GitOperationType = "commit"
    OpStatus   GitOperationType = "status"
    OpLog      GitOperationType = "log"
    OpDiff     GitOperationType = "diff"
    OpRemote   GitOperationType = "remote"
    OpConfig   GitOperationType = "config"
)

// GitCommand represents a Git command specification.
type GitCommand struct {
    OperationType    GitOperationType `json:"operationType"`
    Args             []string          `json:"args"`
    CacheTTLSeconds  int               `json:"cacheTTLSeconds"`
    RetryCount       int               `json:"retryCount"`
    TimeoutSeconds   int               `json:"timeoutSeconds"`
}

// GitResult represents the result of a Git operation.
type GitResult struct {
    Success       bool              `json:"success"`
    Stdout        string            `json:"stdout"`
    Stderr        string            `json:"stderr"`
    ReturnCode    int               `json:"returnCode"`
    ExecutionTime time.Duration     `json:"executionTime"`
    Cached        bool              `json:"cached"`
    CacheHit      bool              `json:"cacheHit"`
    OperationType GitOperationType  `json:"operationType"`
    Command       []string          `json:"command"`
}

// ProjectInfo represents comprehensive Git project information.
type ProjectInfo struct {
    Branch      string    `json:"branch"`
    LastCommit  string    `json:"lastCommit"`
    CommitTime  string    `json:"commitTime"`
    Changes     int       `json:"changes"`
    FetchTime   time.Time `json:"fetchTime"`
}

// GitOperationsManager manages optimized Git operations.
type GitOperationsManager interface {
    ExecuteCommand(cmd GitCommand) GitResult
    ExecuteParallel(cmds []GitCommand) []GitResult
    GetProjectInfo() ProjectInfo
    GetStatistics() Statistics
    ClearCache(opType GitOperationType) int
    Shutdown()
}

// Statistics represents performance and cache statistics.
type Statistics struct {
    Operations      OperationStats `json:"operations"`
    Cache           CacheStats      `json:"cache"`
    Queue           QueueStats      `json:"queue"`
}

type OperationStats struct {
    Total               int     `json:"total"`
    CacheHits           int     `json:"cacheHits"`
    CacheMisses         int     `json:"cacheMisses"`
    CacheHitRate        float64 `json:"cacheHitRate"`
    Errors              int     `json:"errors"`
    AvgExecutionTime    float64 `json:"avgExecutionTime"`
}

type CacheStats struct {
    Size        int     `json:"size"`
    SizeLimit   int     `json:"sizeLimit"`
    Utilization float64 `json:"utilization"`
}

type QueueStats struct {
    Pending int `json:"pending"`
}
```

### 4.2 Cache Key Generation

```go
// Generate cache key from operation context
func (m *GitManager) generateCacheKey(opType GitOperationType, args []string) string {
    data := fmt.Sprintf("%s:%v:%s:%s",
        opType,
        args,
        os.Getwd(),
        m.getCurrentBranch(), // For status/diff operations
    )
    hash := md5.Sum([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### 4.3 Parallel Execution Pattern

```go
// Execute multiple Git commands in parallel
func (m *GitManager) ExecuteParallel(cmds []GitCommand) []GitResult {
    results := make([]GitResult, len(cmds))
    var wg sync.WaitGroup
    sem := make(chan struct{}, m.maxWorkers)

    for i, cmd := range cmds {
        wg.Add(1)
        go func(idx int, c GitCommand) {
            defer wg.Done()
            sem <- struct{}{}        // Acquire
            defer func() { <-sem }()  // Release
            results[idx] = m.ExecuteCommand(c)
        }(i, cmd)
    }

    wg.Wait()
    return results
}
```

### 4.4 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| 단일 Git 명령 실행            | < 2s      | Benchmark test                         |
| 프로젝트 정보 조회(4개 병렬)   | < 3s      | Benchmark test                         |
| 캐시 적중률                   | > 50%     | Production metrics                      |
| 캐시 조회                     | < 1ms     | Benchmark test                         |
| 병렬 명령 4개 실행             | < 3s      | Benchmark test                         |
| 메모리 사용량                 | < 100MB   | Runtime profiling                      |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-200~202 | `manager.go`                  |
| REQ-HOOK-210~212 | `pool.go`                     |
| REQ-HOOK-220~223 | `cache.go`                    |
| REQ-HOOK-230~232 | `pool.go`                     |
| REQ-HOOK-240~241 | `stats.go`                    |
| REQ-HOOK-250     | `manager.go`                  |

### 5.2 Python Hook Mapping

| Python Script                       | Go Handler         | Status  |
|-------------------------------------|--------------------|---------|
| `lib/git_operations_manager.py`    | `manager.go`       | Planned |

### 5.3 Integration Points

- **SPEC-HOOK-001**: Registry, Protocol 재사용
- **internal/cli/hook.go**: SessionStart/Stop 훅에서 프로젝트 정보 조회
- **.moai/memory/**: 통계 저장

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 3 - Advanced Features

### Summary

Git Operations Manager for optimized Git command execution with parallel processing, result caching, and connection pooling. Reduces redundant Git operations through intelligent caching and provides performance statistics tracking. Integrates with SessionStart/SessionEnd hooks for project information display.

### Python Reference

- `lib/git_operations_manager.py` (593 LOC)

### Estimated LOC

- `manager.go`: ~400 LOC
- `pool.go`: ~300 LOC
- `cache.go`: ~200 LOC
- `stats.go`: ~100 LOC
- Test files: ~400 LOC
- Total: ~1,400 LOC
