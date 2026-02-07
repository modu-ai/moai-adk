---
spec_id: SPEC-HOOK-004
title: LSP Diagnostics Integration - Implementation Plan
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, implementation-plan, lsp, diagnostics, quality-gate"
---

# SPEC-HOOK-004: Implementation Plan

## 1. Overview

### 1.1 Scope

Python 기반 LSP 진단 훅(~716 LOC)을 Go로 포팅하여 실시간 진단 피드백과 회귀 탐지 기능을 제공한다. `internal/lsp/hook/` 패키지에 4개 파일을 구현하며, Phase 3 Advanced Features의 P2 Medium 모듈이다.

### 1.2 Implementation Strategy

**Bottom-Up 접근**: Fallback -> Diagnostics -> Tracker -> Gate 순서로 구현한다.

### 1.3 Dependencies

| Dependency       | Status  | Blocking | Impact                                     |
|------------------|---------|----------|--------------------------------------------|
| SPEC-HOOK-001    | Completed | Yes     | Registry, Protocol, Contract 재사용        |
| internal/lsp     | Planned | Yes      | LSP 클라이언트                               |
| Go 1.22+         | Available | No     | context, encoding/json, sync               |

---

## 2. Task Decomposition

### Milestone 1: Fallback Infrastructure (Primary Goal)

LSP 없을 때 사용할 fallback CLI 도구 인프라.

#### Task 1.1: Fallback Diagnostics (`fallback.go`)

**Priority**: High

**Description**: CLI 도구 기반 진단 수집.

**Implementation Details**:
- `RunFallback()`: 언어별 CLI 도구 실행
- JSON 파싱 (ruff, eslint)
- 텍스트 파싱 (tsc, go vet)
- 타임아웃: 30초

**Testing**:
- Python ruff fallback 테스트
- TypeScript tsc fallback 테스트
- Go go vet fallback 테스트
- 파싱 성공/실패 테스트

**Covered Requirements**: REQ-HOOK-160, REQ-HOOK-161, REQ-HOOK-162

---

### Milestone 2: LSP Integration (Secondary Goal)

LSP 클라이언트 통합.

#### Task 2.1: LSP Diagnostics (`diagnostics.go`)

**Priority**: High

**Description**: LSP 진단 수집 및 포맷팅.

**Implementation Details**:
- `GetDiagnostics()`: internal/lsp 클라이언트 호출
- 비동기 실행 래퍼 (sync.WaitGroup 또는 goroutine)
- `FormatDiagnostics()`: Claude 피드백 포맷팅
- fallback과 통합

**Testing**:
- LSP 사용 가능/불가능 테스트
- 진단 포맷팅 테스트
- 타임아웃 처리 테스트

**Covered Requirements**: REQ-HOOK-150, REQ-HOOK-151, REQ-HOOK-152, REQ-HOOK-153

---

### Milestone 3: Regression and Quality Gate (Final Goal)

회귀 탐지와 품질 게이트.

#### Task 3.1: Regression Tracker (`tracker.go`)

**Priority**: Medium

**Description**: 진단 기준선 관리 및 회귀 탐지.

**Implementation Details**:
- `SaveBaseline()`: JSON 파일로 기준선 저장
- `CompareWithBaseline()`: 이전 진단과 비교
- 세션 통계 추적

**Testing**:
- 기준선 저장/로드 테스트
- 회귀 탐지 테스트
- 개선 탐지 테스트

**Covered Requirements**: REQ-HOOK-170, REQ-HOOK-171, REQ-HOOK-172

#### Task 3.2: Quality Gate Enforcer (`gate.go`)

**Priority**: Medium

**Description**: 품질 게이트 규칙 강제.

**Implementation Details**:
- `ShouldBlock()`: 게이트 규칙 평가
- `LoadConfig()`: YAML에서 게이트 설정 로드
- exit code 결정

**Testing**:
- 게이트 규칙 평가 테스트
- 설정 로드 테스트
- exit code 결정 테스트

**Covered Requirements**: REQ-HOOK-180, REQ-HOOK-181, REQ-HOOK-182

---

## 3. Technology Specifications

### 3.1 Language and Runtime

| Component     | Specification          |
|---------------|------------------------|
| Language      | Go 1.22+               |
| Module        | `github.com/modu-ai/moai-adk-go` |
| Package       | `internal/lsp/hook`    |
| Build         | `CGO_ENABLED=0`        |

### 3.2 Standard Library Dependencies

| Package           | Purpose                                |
|-------------------|----------------------------------------|
| `context`         | Cancellation, timeouts                  |
| `encoding/json`   | Baseline persistence, diagnostic parsing |
| `os/exec`         | Fallback tool execution                 |
| `sync`            | Async wait for LSP                     |
| `time`            | Timestamps, TTL                        |

### 3.3 Internal Dependencies

| Package         | Interface Used        | Purpose                        |
|-----------------|----------------------|--------------------------------|
| `internal/lsp`  | `Client`              | LSP server communication      |
| `internal/config`| `ConfigManager`      | Quality gate configuration    |
| `internal/hook`  | `Handler`             | Hook system integration       |

---

## 4. Risk Analysis

### 4.1 Technical Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **LSP 서버 시작 시간**                    | Medium   | High       | Fallback CLI 도구로 우선 실행                             |
| **비동기 실행 복잡성**                   | Medium   | Medium     | goroutine + WaitGroup 패턴 사용                           |
| **JSON 파싱 실패**                       | Medium   | Medium     | 텍스트 기반 fallback 파싱                                 |

### 4.2 Process Risks

| Risk                                    | Severity | Likelihood | Mitigation                                                |
|-----------------------------------------|----------|------------|-----------------------------------------------------------|
| **internal/lsp 패키지 미구현**            | High     | Medium     | Fallback 우선 구현으로 독립적 개발 가능                    |

---

## 5. Migration Plan (Python -> Go)

### 5.1 Migration Strategy

**단계적 교체(Phased Replacement)**: settings.json의 hook command를 Python 스크립트에서 Go 서브커맨드로 변경한다.

### 5.2 Migration Steps

| Step | Action                                    | Verification                              |
|------|-------------------------------------------|-------------------------------------------|
| 1    | Go LSP 진단 모듈 구현 및 단위 테스트 통과      | `go test ./internal/lsp/hook/...`          |
| 2    | Fallback 도구 테스트                       | CLI 도구 실행 확인                          |
| 3    | PostToolUse 통합 테스트                     | 진단 수집 확인                              |
| 4    | settings.json 생성기에서 hook command 변경   | `moai hook lsp-diagnostics` 수동 실행 확인   |
| 5    | Python 훅 제거                              | `.claude/hooks/` 정리                      |

---

## 6. Architecture Design Direction

### 6.1 Package Structure

```
internal/lsp/hook/
    diagnostics.go           # LSP diagnostics collection
    diagnostics_test.go
    fallback.go              # CLI tool fallback
    fallback_test.go
    tracker.go               # Regression tracking
    tracker_test.go
    gate.go                  # Quality gate enforcement
    gate_test.go
    types.go                 # Shared types
```

### 6.2 Dependency Flow

```
internal/cli/hook.go (PostToolUse)
    |
    v
internal/lsp/hook/diagnostics.go -- uses --> internal/lsp/client.go
    |                                         |
    v                                         v
fallback.go (CLI tools)              LSP server (gopls, pyright, etc.)
    |
    v
tracker.go (baseline comparison)
    |
    v
gate.go (quality gate decision)
```

### 6.3 Constructor Pattern

```go
// NewDiagnosticsCollector creates a collector with fallback.
func NewDiagnosticsCollector(client *lsp.Client) *DiagnosticsCollector {
    return &DiagnosticsCollector{
        client:   client,
        fallback: NewFallbackDiagnostics(),
        tracker:  NewRegressionTracker(),
        gate:     NewQualityGateEnforcer(),
    }
}
```

---

## 7. Quality Criteria

### 7.1 Coverage Target

| Scope                    | Target | Rationale                              |
|--------------------------|--------|----------------------------------------|
| `internal/lsp/hook/` 전체   | 85%    | LSP 통합, 복잡한 로직                   |
| Diagnostics             | 90%    | 핵심 경로 완전 검증                      |
| Fallback                | 85%    | 다양한 도구 지원                         |
| Tracker                 | 85%    | 상태 추적 로직                           |

### 7.2 TRUST 5 Compliance

| Principle   | LSP Module Application                                       |
|-------------|--------------------------------------------------------------|
| Tested      | 85%+ coverage, table-driven tests                            |
| Readable    | Go naming conventions, godoc comments                       |
| Unified     | gofumpt formatting, golangci-lint compliance                 |
| Secured     | Path validation, input sanitization                         |
| Trackable   | Conventional commits, SPEC-HOOK-004 reference in all commits |
