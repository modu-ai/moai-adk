---
id: SPEC-HOOK-004
title: LSP Diagnostics Integration
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P2 Medium
phase: "Phase 3 - Advanced Features"
module: "internal/lsp/hook/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-HOOK-002
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, lsp, diagnostics, quality-gate, regression-detection, P2"
---

# SPEC-HOOK-004: LSP Diagnostics Integration

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 LSP(Language Server Protocol) 진단 정보를 Claude Code와 통합하여 실시간 피드백을 제공한다. 이 SPEC은 PostToolUse 훅에서 LSP 진단을 수집하고 보고하는 시스템을 정의한다.

### 1.2 Problem Statement

Python 기반 MoAI-ADK의 `post_tool__lsp_diagnostic.py` 훅은 LSP 통합을 제공하지만 다음과 같은 문제가 있다:

- **LSP 서버 의존성**: 각 언어별 LSP 서버 실행 필요
- **비동기 실행 복잡성**: asyncio 이벤트 루프 관리
- **fallback 부족**: LSP 없을 때 대안 도구 미흡
- **상태 추적 부재**: 진단 이력 추적 및 회귀 탐지 미지원

### 1.3 Target Module

- **경로**: `internal/lsp/hook/`
- **파일 구성**: `diagnostics.go`, `fallback.go`, `tracker.go`, `gate.go`
- **예상 LOC**: ~1,400

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| internal/lsp     | Internal | LSP client and server manager                  |
| Go 1.22+         | Runtime  | context, encoding/json, sync                   |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 `moai hook lsp` 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | internal/lsp 패키지가 LSP 클라이언트를 제공한다                 | High       | LSP 기능 완전 불가                     |
| A-002 | LSP 서버는 비동기 실행을 필요로 한다                            | High       | 동기 래퍼 필요                         |
| A-003 | LSP 없을 때 CLI 도구로 fallback 가능하다                      | High       | fallback 없이 기능 제한                 |
| A-004 | 진단 상태를 파일 단위로 저장할 수 있다                        | High       | 상태 추적 불가                         |
| A-005 | 품질 게이트 설정은 YAML에서 로드할 수 있다                    | High       | 게이트 설정 불가                       |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 사용자는 실시간 진단 피드백을 원한다                           | High       | 수동 검사 요청 증가                     |
| B-002 | 회귀 탐지는 품질 유지에 중요하다                               | High       | 버전 간 품질 저하 위험                  |

---

## 3. Requirements (R)

### Module 1: LSP Diagnostics Collection (진단 수집)

**REQ-HOOK-150** [Ubiquitous]
시스템은 **항상** PostToolUse 이벤트에서 파일 수정 후 LSP 진단을 수집해야 한다.

- `GetDiagnostics(ctx, filePath) ([]Diagnostic, error)`: 진단 수집
- `GetSeverityCounts(diagnostics []Diagnostic) SeverityCounts`: 심각도 집계
- 비동기 실행 래퍼 (asyncio.run 또는 nest_asyncio)

**REQ-HOOK-151** [Event-Driven]
**WHEN** LSP 서버가 사용 가능하면 **THEN** 시스템은 LSP 진단을 우선 사용해야 한다.

**REQ-HOOK-152** [State-Driven]
**IF** LSP 서버를 사용할 수 없으면 **THEN** 시스템은 fallback CLI 도구를 사용해야 한다.

**REQ-HOOK-153** [Unwanted]
시스템은 LSP 진단 실패 시 파일 작성을 차단**하지 않아야 한다**. 진단은 관찰 전용으로 작동해야 한다.

### Module 2: Fallback CLI Tools (CLI 도구 대안)

**REQ-HOOK-160** [Ubiquitous]
시스템은 **항상** LSP 없을 때 사용할 fallback 도구를 지원해야 한다.

| Language    | Fallback Tools                              |
|-------------|---------------------------------------------|
| Python      | ruff check, flake8, pylint                   |
| TypeScript  | tsc, eslint                                 |
| JavaScript  | eslint                                      |
| Go          | go vet, golangci-lint                        |
| Rust        | cargo clippy                                |

**REQ-HOOK-161** [Event-Driven]
**WHEN** fallback 도구가 실행되면 **THEN** 시스템은 출력을 진단 형식으로 변환해야 한다.

**REQ-HOOK-162** [State-Driven]
**IF** fallback 도구도 사용할 수 없으면 **THEN** 시스템은 "diagnostics unavailable" 메시지를 반환해야 한다.

### Module 3: Regression Detection (회귀 탐지)

**REQ-HOOK-170** [Ubiquitous]
시스템은 **항상** 진단 이력을 저장하여 회귀를 탐지해야 한다.

- `SaveBaseline(filePath string, diagnostics []Diagnostic) error`: 기준선 저장
- `CompareWithBaseline(filePath string, diagnostics []Diagnostic) RegressionReport`: 회귀 비교
- 상태 파일: `.moai/memory/diagnostics-baseline.json`

**REQ-HOOK-171** [Event-Driven]
**WHEN** 새 진단이 기준선보다 더 많은 error를 포함하면 **THEN** 시스템은 "Regression detected" 경고를 보고해야 한다.

**REQ-HOOK-172** [State-Driven]
**IF** error 수가 감소했으면 **THEN** 시스템은 "Improvement" 메시지를 보고해야 한다.

### Module 4: Quality Gate Enforcement (품질 게이트)

**REQ-HOOK-180** [Ubiquitous]
시스템은 **항상** 품질 게이트 설정을 확인하여 진단에 따라 동작을 결정해야 한다.

- `ShouldBlock(diagnostics []Diagnostic, gate QualityGate) bool`: 차단 여부 결정
- 게이트 설정: `.moai/config/sections/quality.yaml`

**REQ-HOOK-181** [Event-Driven]
**WHEN** error 수가 게이트 임계값을 초과하면 **THEN** 시스템은 `exit code 2`를 반환해야 한다.

**REQ-HOOK-182** [State-Driven]
**IF** warning 수가 게이트 임계값을 초과하면 **THEN** 시스템은 경고를 로그하지만 계속 진행해야 한다.

### Module 5: Progress Tracking (진행 추적)

**REQ-HOOK-190** [Ubiquitous]
시스템은 **항상** 세션 동안의 진단 통계를 추적해야 한다.

- `SessionStats`: error/warning/info/hint 누적 count
- `FileStats`: 파일별 진단 이력

**REQ-HOOK-191** [Event-Driven]
**WHEN** SessionEnd 이벤트가 발생하면 **THEN** 시스템은 세션 진단 요약을 보고해야 한다.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// DiagnosticSeverity represents the severity of a diagnostic.
type DiagnosticSeverity string

const (
    SeverityError       DiagnosticSeverity = "error"
    SeverityWarning     DiagnosticSeverity = "warning"
    SeverityInformation DiagnosticSeverity = "information"
    SeverityHint        DiagnosticSeverity = "hint"
)

// Diagnostic represents a single LSP diagnostic.
type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity"`
    Code     string             `json:"code,omitempty"`
    Source   string             `json:"source,omitempty"`
    Message  string             `json:"message"`
}

// Range represents a range in a text document.
type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

// Position represents a position in a text document.
type Position struct {
    Line      int `json:"line"`
    Character int `json:"character"`
}

// SeverityCounts represents counts by severity.
type SeverityCounts struct {
    Errors       int `json:"errors"`
    Warnings     int `json:"warnings"`
    Information  int `json:"information"`
    Hints        int `json:"hints"`
}

// RegressionReport compares current diagnostics with baseline.
type RegressionReport struct {
    HasRegression bool   `json:"hasRegression"`
    HasImprovement bool  `json:"hasImprovement"`
    NewErrors     int    `json:"newErrors"`
    FixedErrors   int    `json:"fixedErrors"`
    NewWarnings   int    `json:"newWarnings"`
    FixedWarnings int    `json:"fixedWarnings"`
}

// QualityGate defines quality gate thresholds.
type QualityGate struct {
    MaxErrors     int    `json:"maxErrors"`
    MaxWarnings   int    `json:"maxWarnings"`
    BlockOnError  bool   `json:"blockOnError"`
    BlockOnWarning bool  `json:"blockOnWarning"`
}

// LSPDiagnosticsCollector collects LSP diagnostics.
type LSPDiagnosticsCollector interface {
    GetDiagnostics(ctx context.Context, filePath string) ([]Diagnostic, error)
    GetSeverityCounts(diagnostics []Diagnostic) SeverityCounts
}

// FallbackDiagnostics uses CLI tools when LSP unavailable.
type FallbackDiagnostics interface {
    RunFallback(filePath string) ([]Diagnostic, error)
    IsAvailable(language string) bool
}

// RegressionTracker tracks diagnostic baselines.
type RegressionTracker interface {
    SaveBaseline(filePath string, diagnostics []Diagnostic) error
    CompareWithBaseline(filePath string, diagnostics []Diagnostic) (RegressionReport, error)
}

// QualityGateEnforcer enforces quality gate rules.
type QualityGateEnforcer interface {
    ShouldBlock(counts SeverityCounts, gate QualityGate) bool
    LoadConfig() (QualityGate, error)
}
```

### 4.2 Supported Languages for Fallback

| Language    | Fallback Tools                    | JSON Output Support |
|-------------|-----------------------------------|---------------------|
| Python      | ruff check --output-format=json   | Yes                 |
| TypeScript  | tsc --pretty false               | No (parse stderr)   |
| JavaScript  | eslint --format json             | Yes                 |
| Go          | go vet                           | No (parse stdout)   |
| Rust        | cargo clippy --message-format=json | Yes               |

### 4.3 Diagnostic Output Format

```go
// Hook output format for LSP diagnostics
type LSPHookOutput struct {
    HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput"`
}

// Example additionalContext:
// "LSP: 2 error(s), 1 warning(s) in main.go
//   - [ERROR] Line 45: undeclared name 'foo' [gopls]
//   - [ERROR] Line 67: cannot use 'x' (type int) as type string [gopls]
//   - [WARNING] Line 12: unused variable 'temp' [gopls]"
```

### 4.4 Baseline Storage Format

```go
// .moai/memory/diagnostics-baseline.json
type DiagnosticsBaseline struct {
    Version   string                    `json:"version"`
    UpdatedAt time.Time                 `json:"updatedAt"`
    Files     map[string]FileBaseline   `json:"files"`
}

type FileBaseline struct {
    Path        string         `json:"path"`
    Hash        string         `json:"hash"`
    Diagnostics []Diagnostic   `json:"diagnostics"`
    UpdatedAt   time.Time      `json:"updatedAt"`
}
```

### 4.5 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| LSP 진단 수집                  | < 3s      | Benchmark test                         |
| Fallback 도구 실행             | < 5s      | Benchmark test                         |
| 기준선 저장/로드               | < 50ms    | Benchmark test                         |
| 회귀 비교                     | < 10ms    | Benchmark test                         |
| 메모리 사용량 (진단 중)        | < 50MB    | Runtime profiling                      |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-150~153 | `diagnostics.go`              |
| REQ-HOOK-160~162 | `fallback.go`                 |
| REQ-HOOK-170~172 | `tracker.go`                  |
| REQ-HOOK-180~182 | `gate.go`                     |
| REQ-HOOK-190~191 | `tracker.go` (session stats)  |

### 5.2 Python Hook Mapping

| Python Script                       | Go Handler         | Status  |
|-------------------------------------|--------------------|---------|
| `post_tool__lsp_diagnostic.py`      | `diagnostics.go`   | Planned |
| `quality_gate_with_lsp.py`          | `gate.go`          | Planned |

### 5.3 Integration Points

- **SPEC-HOOK-001**: Registry, Protocol, Contract 재사용
- **internal/lsp/**: LSP 클라이언트 및 서버 매니저
- **internal/config/**: 품질 게이트 설정 로드
- **.moai/memory/**: 진단 기준선 저장

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 3 - Advanced Features

### Summary

LSP Diagnostics Integration for real-time feedback after code modifications. Collects diagnostics from LSP servers with CLI tool fallback, tracks regression via baseline comparison, and enforces quality gates. Integrates with PostToolUse hook to provide actionable error information to Claude Code.

### Python Reference

- `post_tool__lsp_diagnostic.py` (516 LOC)
- `quality_gate_with_lsp.py` (~200 LOC)

### Estimated LOC

- `diagnostics.go`: ~400 LOC
- `fallback.go`: ~300 LOC
- `tracker.go`: ~400 LOC
- `gate.go`: ~200 LOC
- Test files: ~600 LOC
- Total: ~1,900 LOC
