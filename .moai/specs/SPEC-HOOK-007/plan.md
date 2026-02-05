---
spec_id: SPEC-HOOK-007
title: Session Lifecycle Enhancements - Implementation Plan
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, implementation-plan, session-lifecycle, cleanup"
---

# SPEC-HOOK-007: Implementation Plan

## 1. Overview

### 1.1 Scope

세션 라이프사이클 향상 기능을 구현하는 `internal/hook/lifecycle/` 패키지. SessionStart 향상, SessionEnd 자동 정리, 메트릭 수집, 상태 지속성을 포함하며, Phase 3 Advanced Features의 P2 Medium 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: Persistence -> Metrics -> Cleanup -> SessionEnhanced 순서로 구현한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-HOOK-001    | Completed | Yes     | Hook system integration                    |
| SPEC-HOOK-005    | Planned | Yes      | Git Operations Manager                       |
| Go 1.22+         | Available | No     | encoding/json, os, time, filepath          |

---

## 2. Task Decomposition

### Milestone 1: Core Infrastructure (Primary Goal)

핵심 인프라 2개 파일 구현. 상태 지속성과 메트릭.

#### Task 1.1: Persistence (`persistence.go`)

**Priority**: High

**Description**: 작업 상태 저장 및 복구.

**Implementation Details**:
- `Save()`: JSON 파일로 상태 저장
- `Load()`: 상태 복구
- `UpdatePosition()`: 위치 업데이트
- `.moai/memory/last-session-state.json`

**Testing**:
- 상태 저장/로드 테스트
- 파일 없을 때 처리 테스트
- 손상된 JSON 복구 테스트

**Covered Requirements**: REQ-HOOK-380, REQ-HOOK-381, REQ-HOOK-382

#### Task 1.2: Metrics (`metrics.go`)

**Priority**: High

**Description**: 세션 메트릭 수집 및 저장.

**Implementation Details**:
- `RecordToolUse()`: 도구 사용 기록
- `RecordFileModification()`: 파일 수정 기록
- `RecordError()`: 오류 기록
- `Save()`: `.moai/memory/session-stats.json`에 저장

**Testing**:
- 메트릭 수집 테스트
- 저장/로드 테스트
- 집계 계산 테스트

**Covered Requirements**: REQ-HOOK-370, REQ-HOOK-371, REQ-HOOK-372

---

### Milestone 2: Lifecycle Enhancements (Secondary Goal)

세션 라이프사이클 향상.

#### Task 2.1: Cleanup (`cleanup.go`)

**Priority**: High

**Description**: 자동 정리 기능.

**Implementation Details**:
- `CleanTempFiles()`: .moai/temp/ 삭제
- `ClearCaches()`: 캐시 정리
- `GenerateCleanupReport()`: 정리 보고서 생성

**Testing**:
- 정리 작업 테스트
- 부분 실패 시 graceful degradation 테스트
- 보고서 생성 테스트

**Covered Requirements**: REQ-HOOK-360, REQ-HOOK-361, REQ-HOOK-362

#### Task 2.2: Session Enhanced (`session_enhanced.go`)

**Priority**: Medium

**Description**: 향상된 SessionStart.

**Implementation Details**:
- `GetProjectInfo()`: Git 정보 조회 (SPEC-HOOK-005 활용)
- `FormatWelcomeMessage()`: ASCII art 배너 생성
- config에서 프로젝트 정보 로드

**Testing**:
| Git repository | Git repository without Git |
| 프로젝트 정보 표시 | 프로젝트 이름만 표시 |
| 배너 포맷 테스트 | | 인터페이스 구현 확인 |

**Covered Requirements**: REQ-HOOK-350, REQ-HOOK-351, REQ-HOOK-352

---

### Milestone 3: Integration (Final Goal)

Hook 통합.

#### Task 3.1: Hook Integration

**Priority**: Medium

**Description**: SessionStart/SessionEnd 훅 통합.

**Implementation Details**:
- `internal/hook/session_start.go` 확장
- `internal/hook/session_end.go` 확장
- 메트릭 수집 통합
- 정리 작업 통합

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/hook/lifecycle` |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `encoding/json`   | State persistence                     |
| `os/filepath`      | File path operations                   |
| `os`               | File operations                        |
| `time`             | Timestamps, duration                   |

### 3.3 Internal Dependencies

| Package         | Interface Used        | Purpose                        |
|-----------------|----------------------|--------------------------------|
| `internal/config`| `ConfigManager`      | Configuration loading          |
| `internal/git/ops`| `GitOperationsManager`| Git project info             |
| `internal/hook`  | `Handler`             | Hook system integration      |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **파일 시스템 권한**                    | Medium   | Low        | 오류 무시하고 계속 진행                                 |
| **JSON 복구 실패**                     | Low      | Low        | 빈 상태로 초기화                                             |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **UI 메시지 형식**                     | Low      | Low        | 단순한 텍스트로 시작, UI 개선은 별도 SPEC        |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: settings.json의 hook command를 Python 스크립트에서 Go 서브커맨드로 변경한다.

### 5.2 Migration Steps

| Step | Action                                    | Verification                              |
|------|-------------------------------------------|-------------------------------------------|
| 1    | Go 세션 라이프사이클 모듈 구현 및 단위 테스트 통과 | `go test ./internal/hook/lifecycle/...`   |
| 2    | 상태 지속성 테스트                         | 저장/복구 확인                            |
| 3    | Hook 통합 테스트                            | SessionStart/SessionEnd 동작 확인         |
| 4    | settings.json 생성기에서 hook command 변경   | 수동 실행 확인                              |
| 5    | Python 훅 제거                              | `.claude/hooks/` 정리                      |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/hook/lifecycle/
    session_enhanced.go     # Enhanced SessionStart
    session_enhanced_test.go
    cleanup.go              # Auto-cleanup for SessionEnd
    cleanup_test.go
    metrics.go              # Session metrics collection
    metrics_test.go
    persistence.go          # Work state persistence
    persistence_test.go
    types.go                # Shared types
    banner.go               # ASCII art banners
```

### 6.2 Dependency Flow

```
internal/hook/session_start.go -- uses --> lifecycle/session_enhanced.go
    |
    v
session_enhanced.go -- uses --> internal/git/ops/manager.go

internal/hook/session_end.go -- uses --> lifecycle/cleanup.go
    |                                 |
    v                                 v
lifecycle/metrics.go             persistence.go
```

### 6.3 Constructor Pattern

```go
// NewSessionEnhancer creates an enhancer with dependencies.
func NewSessionEnhancer(
    cfg config.ConfigManager,
    gitMgr *gitops.GitOperationsManager,
) *SessionEnhancer {
    return &SessionEnhancer{
        cfg:    cfg,
        gitMgr: gitMgr,
    }
}

// NewSessionCleanup creates a cleanup handler.
func NewSessionCleanup() *SessionCleanup {
    return &SessionCleanup{
        tempDir: filepath.Join(".moai", "temp"),
    }
}
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                        | Target | Rationale                              |
|------------------------------|--------|----------------------------------------|
| `internal/hook/lifecycle/` 전체 | 85%    | 라이프사이클, 상태 저장              |
| SessionEnhanced             | 85%    | 프로젝트 정보 표시                   |
| Cleanup                      | 90%    | 정리 로직 완전 검증                  |
| Metrics                      | 85%    | 메트릭 수집, 저장                    |
| Persistence                  | 90%    | 상태 저장, 복구                       |

### 7.2 TRUST 5 Compliance

| Principle   | Lifecycle Module Application                                   |
|-------------|----------------------------------------------------------------|
| Tested      | 85%+ coverage, table-driven tests                                  |
| Readable    | Go naming conventions, godoc comments                            |
| Unified     | gofumpt formatting, golangci-lint compliance                       |
| Secured     | Path validation, temp file cleanup                                 |
| Trackable   | Conventional commits, SPEC-HOOK-007 reference in all commits    |
