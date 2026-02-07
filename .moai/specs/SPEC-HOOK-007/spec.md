---
id: SPEC-HOOK-007
title: Session Lifecycle Enhancements
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P2 Medium
phase: "Phase 3 - Advanced Features"
module: "internal/hook/lifecycle/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-HOOK-005
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, session-lifecycle, session-start, session-end, auto-cleanup, P2"
---

# SPEC-HOOK-007: Session Lifecycle Enhancements

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 Claude Code 세션 수명 주기 동안 향상된 사용자 경험을 제공한다. 이 SPEC은 SessionStart의 향상된 프로젝트 정보 표시와 SessionEnd의 자동 정리 및 메트릭 수집을 정의한다.

### 1.2 Problem Statement

Python 기반 MoAI-ADK의 세션 라이프사이클 훅은 다음과 같은 개선이 필요하다:

- **SessionStart**: 프로젝트 정보 표시가 제한적임
- **SessionEnd**: 정리가 자동화되지 않아 사용자가 수동으로 정리해야 함
- **메트릭 부족**: 세션 동안의 활동 추적이 불완전
- **상태 지속성**: 세션 간 상태가 유지되지 않음

### 1.3 Target Module

- **경로**: `internal/hook/lifecycle/`
- **파일 구성**: `session_enhanced.go`, `cleanup.go`, `metrics.go`, `persistence.go`
- **예상 LOC**: ~800

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| SPEC-HOOK-005    | Internal | Git Operations Manager                         |
| SPEC-CONFIG-001  | Internal | Configuration Manager                         |
| Go 1.22+         | Runtime  | encoding/json, os, time                      |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | SessionStart/SessionEnd 이벤트는 세션 시작/종료에 발생한다  | High       | 이벤트 미수신으로 기능 작동 불가           |
| A-002 | .moai/memory/ 디렉터리에 상태를 저장할 수 있다              | High       | 상태 지속 불가                         |
| A-003 | Git 정보는 프로젝트 컨텍스트 제공에 사용된다              | High       | 컨텍스트 부족                          |
| A-004 | 임시 파일은 세션 중 생성될 수 있다                        | High       | 정리 대상 식별 불가                     |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 사용자는 세션 시작 시 프로젝트 정보를 보길 원한다          | High       | 사용자 경험 저하                        |
| B-002 | 자동 정리는 사용자 편의성을 높여준다                      | High       | 수동 정리 부담 증가                     |

---

## 3. Requirements (R)

### Module 1: Enhanced SessionStart (향상된 세션 시작)

**REQ-HOOK-350** [Ubiquitous]
시스템은 **항상** SessionStart 이벤트에서 향상된 프로젝트 정보를 표시해야 한다.

- 프로젝트 이름, 버전, 설명
- Git 브랜치, 마지막 커밋
- 변경된 파일 수
- MoAI 버전 정보

**REQ-HOOK-351** [Event-Driven]
**WHEN** SessionStart가 발생하면 **THEN** 시스템은 GitOperationsManager를 통해 프로젝트 정보를 병렬로 조회해야 한다.

**REQ-HOOK-352** [State-Driven]
**IF** Git 저장소가 아니면 **THEN** 시스템은 Git 정보를 생략하고 프로젝트 이름과 버전만 표시해야 한다.

### Module 2: Auto-Cleanup (자동 정리)

**REQ-HOOK-360** [Ubiquitous]
시스템은 **항상** SessionEnd 이벤트에서 임시 리소스를 정리해야 한다.

- `.moai/temp/` 디렉터리 정리
- 세션 중 생성된 임시 파일 삭제
- 캐시 정리 (단, 세션 간 공유 캐시는 보존)

**REQ-HOOK-361** [Event-Driven]
**WHEN** SessionEnd가 발생하면 **THEN** 시스템은 정리 작업을 수행하고 정리된 항목 수를 보고해야 한다.

**REQ-HOOK-362** [State-Driven]
**IF** 정리 중 오류가 발생하면 **THEN** 시스템은 오류를 로그하고 나머지 정리를 계속해야 한다.

### Module 3: Metrics Collection (메트릭 수집)

**REQ-HOOK-370** [Ubiquitous]
시스템은 **항상** 세션 동안 활동 메트릭을 수집해야 한다.

- `tool_use_count`: 도구 사용 횟수
- `files_modified`: 수정된 파일 수
- `errors_committed`: 커밋된 오류 수
- `duration`: 세션 지속 시간

**REQ-HOOK-371** [Event-Driven]
**WHEN** SessionEnd가 발생하면 **THEN** 시스템은 세션 메트릭을 `.moai/memory/session-stats.json`에 저장해야 한다.

**REQ-HOOK-372** [State-Driven]
**IF** 메트릭 저장이 실패하면 **THEN** 시스템은 경고를 로그하고 세션을 정상 종료해야 한다.

### Module 4: Work State Persistence (작업 상태 지속성)

**REQ-HOOK-380** [Ubiquitous]
시스템은 **항상** 작업 상태를 `.moai/memory/last-session-state.json`에 저장해야 한다.

- `last_position`: 마지막 작업 위치
- `active_files`: 활성 파일 목록
- `clipboard`: 클립보드 내역
- `context_summary`: 컨텍스트 요약

**REQ-HOOK-381** [Event-Driven]
**WHEN** PostToolUse 또는 Stop 이벤트가 발생하면 **THEN** 시스템은 작업 상태를 업데이트해야 한다.

**REQ-HOOK-382** [State-Driven]
**IF** 세션이 시작되면 **THEN** 시스템은 저장된 상태를 복구하고 사용자에게 알려야 한다.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// SessionEnhancer handles SessionStart enhancements.
type SessionEnhancer interface {
    GetProjectInfo(ctx context.Context) (*ProjectContext, error)
    FormatWelcomeMessage(info *ProjectContext) string
}

// ProjectContext represents comprehensive project information.
type ProjectContext struct {
    ProjectName   string            `json:"projectName"`
    ProjectVersion string            `json:"projectVersion"`
    Description    string            `json:"description"`
    GitBranch     string            `json:"gitBranch"`
    LastCommit    string            `json:"lastCommit"`
    CommitTime    string            `json:"commitTime"`
    ModifiedFiles int               `json:"modifiedFiles"`
    MoaIVersion   string            `json:"moaiVersion"`
}

// SessionCleanup handles SessionEnd cleanup.
type SessionCleanup interface {
    CleanTempFiles() (*CleanupResult, error)
    ClearCaches() error
    GenerateCleanupReport() string
}

// CleanupResult represents the result of cleanup operations.
type CleanupResult struct {
    FilesDeleted    int    `json:"filesDeleted"`
    DirsDeleted     int    `json:"dirsDeleted"`
    BytesFreed      int64  `json:"bytesFreed"`
    Errors          []string `json:"errors"`
    Duration        time.Duration `json:"duration"`
}

// SessionMetrics tracks session activity.
type SessionMetrics struct {
    SessionID        string        `json:"sessionId"`
    StartTime        time.Time     `json:"startTime"`
    EndTime          time.Time     `json:"endTime"`
    Duration         time.Duration `json:"duration"`
    ToolUseCount     map[string]int `json:"toolUseCount"`
    FilesModified    int           `json:"filesModified"`
    ErrorsCommitted  int           `json:"errorsCommitted"`
}

// MetricsCollector collects and persists session metrics.
type MetricsCollector interface {
    RecordToolUse(toolName string)
    RecordFileModification(filePath string)
    RecordError()
    Save() error
}

// WorkState persists work state between sessions.
type WorkState interface {
    Save(state *WorkStateData) error
    Load() (*WorkStateData, error)
    UpdatePosition(filePath string, line int) error
}

// WorkStateData represents work session state.
type WorkStateData struct {
    LastPosition   *FilePosition    `json:"lastPosition"`
    ActiveFiles    []string         `json:"activeFiles"`
    ContextSummary  string           `json:"contextSummary"`
    Timestamp      time.Time         `json:"timestamp"`
}
```

### 4.2 SessionStart Message Format

```
╭─────────────────────────────────────────────────────────╮
│  MoAI Session Started                                      │
│  ╰──────────────────────────────────────────────────────╯  │
│                                                             │
│  Project: my-project (v1.0.0)                             │
│  Branch: main (2 commits ahead of origin)                  │
│  Last commit: feat: Add user authentication (5m ago)      │
│  Modified files: 3                                         │
│  MoAI: v0.8.0                                              │
╰─────────────────────────────────────────────────────────────╯
```

### 4.3 SessionEnd Message Format

```
╭─────────────────────────────────────────────────────────╮
│  Session Summary                                           │
│  ╰──────────────────────────────────────────────────────╯  │
│                                                             │
│  Duration: 2h 34m                                         │
│  Tool usage: Read(156), Write(42), Edit(23)               │
│  Files modified: 18                                        │
│  Cleanup: 3 files deleted, 2MB freed                      │
╰─────────────────────────────────────────────────────────────╯
```

### 4.4 Cleanup Patterns

| Pattern           | Location                          | Description                        |
|-------------------|-----------------------------------|------------------------------------|
| Temp files        | `.moai/temp/`                     | Session-specific temp files        |
| Cache files       | `.moai/cache/temp/`               | Session-specific caches            |
| Log files          | `.moai/logs/session-*.log`        | Session logs                       |
| State files       | `.moai/memory/*.json`            | Persistent state (keep)           |

### 4.5 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| SessionStart 처리              | < 500ms   | Benchmark test                         |
| SessionEnd 처리                | < 2s      | Benchmark test                         |
| 메트릭 저장                    | < 50ms    | Benchmark test                         |
| 상태 복구                      | < 100ms   | Benchmark test                         |
| 정리 작업                      | < 1s      | Benchmark test                         |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-350~352 | `session_enhanced.go`      |
| REQ-HOOK-360~362 | `cleanup.go`                |
| REQ-HOOK-370~372 | `metrics.go`                |
| REQ-HOOK-380~382 | `persistence.go`            |

### 5.2 Python Hook Mapping

| Python Script                       | Go Handler         | Status  |
|-------------------------------------|--------------------|---------|
| `session_start__show_project_info.py` | `session_enhanced.go` | Planned |
| `session_end__auto_cleanup.py`      | `cleanup.go`        | Planned |
| `session_end__rank_submit.py`       | `metrics.go`        | Planned |

### 5.3 Integration Points

- **SPEC-HOOK-001**: Hook system integration
- **SPEC-HOOK-005**: Git Operations Manager
- **internal/config/**: Configuration Manager
- **.moai/memory/**: State persistence

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 3 - Advanced Features

### Summary

Session lifecycle enhancements for improved user experience. SessionStart displays comprehensive project information, SessionEnd performs automatic cleanup and metrics collection. Work state persistence enables seamless session recovery. Integrates with hook system for automatic lifecycle management.

### Python Reference

- `session_start__show_project_info.py` (~200 LOC)
- `session_end__auto_cleanup.py` (~150 LOC)
- `session_end__rank_submit.py` (~100 LOC)
- `lib/memory_collector.py` (~200 LOC)
- `lib/context_manager.py` (~300 LOC)

### Estimated LOC

- `session_enhanced.go`: ~200 LOC
- `cleanup.go`: ~150 LOC
- `metrics.go`: ~200 LOC
- `persistence.go`: ~150 LOC
- Test files: ~300 LOC
- Total: ~1,000 LOC
