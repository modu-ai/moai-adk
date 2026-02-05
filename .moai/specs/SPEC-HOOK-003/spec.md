---
id: SPEC-HOOK-003
title: Security & Scanning
version: "0.1.0"
status: Draft
created: 2026-02-04
updated: 2026-02-04
author: GOOS
priority: P1 High
phase: "Phase 2 - Quality Integration"
module: "internal/hook/security/"
dependencies:
  - SPEC-HOOK-001
  - SPEC-HOOK-002
  - SPEC-CONFIG-001
adr_references:
  - ADR-006 (Hooks as Binary Subcommands)
  - ADR-012 (Hook Execution Contract)
resolves_issues: []
lifecycle: spec-anchored
tags: "hook, security, ast-grep, scanning, vulnerability-detection, P1"
---

# SPEC-HOOK-003: Security & Scanning

## HISTORY

| Version | Date       | Author | Description                            |
|---------|------------|--------|----------------------------------------|
| 0.1.0   | 2026-02-04 | GOOS   | Initial SPEC creation                  |

---

## 1. Environment (E)

### 1.1 Project Context

MoAI-ADK Go Edition은 Claude Code와 통합하여 AI 개발 워크플로우에 보안 스캐닝 기능을 제공한다. 이 SPEC은 AST-Grep(ast-grep)을 사용한 구조 기반 보안 취약점 탐지 시스템을 정의한다.

### 1.2 Problem Statement

Python 기반 MoAI-ADK의 `post_tool__ast_grep_scan.py` 훅은 실시간 보안 스캐닝을 제공하지만 다음과 같은 문제가 있다:

- **도구 의존성**: ast-grep(sg) 바이너리 필수 설치 필요
- **규칙 관리**: 보안 규칙(sgconfig.yml)의 발견과 로드가 복잡
- **결과 파싱**: JSON 출력 파싱 오류에 취약
- **성능**: 대형 파일 스캔 시 타임아웃 가능

### 1.3 Target Module

- **경로**: `internal/hook/security/`
- **파일 구성**: `ast_grep.go`, `scanner.go`, `rules.go`, `reporter.go`
- **예상 LOC**: ~1,200

### 1.4 Dependencies

| Dependency       | Type     | Description                                    |
|------------------|----------|------------------------------------------------|
| SPEC-HOOK-001    | Internal | Compiled Hook System                          |
| SPEC-HOOK-002    | Internal | Tool Registry (재사용)                         |
| ast-grep (sg)    | External | Structural code search and lint tool           |
| Go 1.22+         | Runtime  | encoding/json, os/exec, context               |

### 1.5 Architecture Reference

- **ADR-006**: Hooks as Binary Subcommands -- 훅을 `moai hook security-scan` 서브커맨드로 구현
- **ADR-012**: Hook Execution Contract -- 실행 환경 보증/비보증 사항 명세

---

## 2. Assumptions (A)

### 2.1 Technical Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| A-001 | ast-grep(sg) 바이너리가 시스템 PATH에 존재하거나 설치 가능하다      | High       | 기능 완전 불가, graceful degradation 필수 |
| A-002 | sgconfig.yml 규칙 파일은 프로젝트 루트 또는 .ast-grep/에 위치한다 | High       | 규칙 발견 실패로 기본 규칙 사용          |
| A-003 | ast-grep는 --json 출력을 지원한다                             | High       | 결과 파싱 불가로 정규식 fallback 필요    |
| A-004 | 40개 언어에 대한 AST 파싱이 가능하다                           | Medium     | 일부 언어 제한 가능                     |
| A-005 | 보안 규칙은 YML/JSON 형식으로 정의할 수 있다                   | High       | 규칙 파싱 불가                         |

### 2.2 Business Assumptions

| ID    | Assumption                                                    | Confidence | Risk if Wrong                          |
|-------|---------------------------------------------------------------|------------|----------------------------------------|
| B-001 | 사용자는 자동 보안 스캔을 원한다                               | High       | 수동 스캔 요청 증가                     |
| B-002 | OWASP Top 10 규칙으로 충분하다                                | Medium     | 추가 규칙 요청 가능                     |

---

## 3. Requirements (R)

### Module 1: AST-Grep Integration (도구 통합)

**REQ-HOOK-100** [Ubiquitous]
시스템은 **항상** ast-grep(sg) 바이너리의 가용성을 확인하고 실행 가능한지 검증해야 한다.

- `IsASTGrepAvailable() bool`: sg 바이너리 존재 확인
- `GetASTGrepVersion() string`: sg 버전 확인
- 미설치 시 graceful degradation으로 skip

**REQ-HOOK-101** [Event-Driven]
**WHEN** PostToolUse 이벤트가 발생하고 파일이 지원되는 언어이면 **THEN** 시스템은 ast-grep 스캔을 실행해야 한다.

**REQ-HOOK-102** [State-Driven]
**IF** 스캔이 취약점을 발견하면 **THEN** 시스템은 심각도(severity)별로 분류하고 Claude에게 요약을 제공해야 한다.

**REQ-HOOK-103** [Unwanted]
시스템은 스캔 실패 시 파일 작성을 차단**하지 않아야 한다**. 보안 스캔은 관찰 전용(observational)으로 작동해야 한다.

### Module 2: Rule Management (규칙 관리)

**REQ-HOOK-110** [Ubiquitous]
시스템은 **항상** 프로젝트별 보안 규칙(sgconfig.yml)을 탐지하고 로드해야 한다.

- `FindRulesConfig() string`: 규칙 파일 경로 탐지
- `LoadRules(configPath string) ([]Rule, error)`: 규칙 로드
- 기본 경로: `.claude/skills/moai-tool-ast-grep/rules/sgconfig.yml`
- 대체 경로: `sgconfig.yml`, `.ast-grep/sgconfig.yml`

**REQ-HOOK-111** [Event-Driven]
**WHEN** 규칙 파일이 발견되지 않으면 **THEN** 시스템은 내장된 OWASP 기본 규칙을 사용해야 한다.

**REQ-HOOK-112** [State-Driven]
**IF** 규칙 파일이 잘못되었으면 **THEN** 시스템은 경고를 로그하고 기본 규칙로 대체해야 한다.

### Module 3: Scanner Execution (스캐너 실행)

**REQ-HOOK-120** [Ubiquitous]
시스템은 **항상** 타임아웃 내에 스캔을 완료해야 한다(기본 30초).

**REQ-HOOK-121** [Event-Driven]
**WHEN** 스캔 명령이 실행되면 **THEN** 시스템은 `sg scan --json` 형식으로 실행하고 JSON 출력을 파싱해야 한다.

**REQ-HOOK-122** [State-Driven]
**IF** JSON 파싱이 실패하면 **THEN** 시스템은 정규식 기반 fallback으로 결과를 추출해야 한다.

**REQ-HOOK-123** [Optional]
**가능하면** 병렬 스캔을 지원하여 다중 파일을 동시에 처리한다.

### Module 4: Finding Reporting (결과 보고)

**REQ-HOOK-130** [Ubiquitous]
시스템은 **항상** 발견된 취약점을 structured format으로 Claude에게 보고해야 한다.

- `Severity`: error, warning, info
- `RuleId`: 규칙 식별자
- `Message`: 취약점 설명
- `Location`: 파일 경로 및 라인 번호

**REQ-HOOK-131** [Event-Driven]
**WHEN** error 심각도의 취약점이 발견되면 **THEN** 시스템은 `exit code 2`를 반환하여 Claude의 주의를 환기해야 한다.

**REQ-HOOK-132** [State-Driven]
**IF** 취약점이 10개 초과로 발견되면 **THEN** 시스템은 상위 10개만 보고하고 나머지는 "... and N more"로 요약해야 한다.

### Module 5: Language Support

**REQ-HOOK-140** [Ubiquitous]
시스템은 **항상** 다음 언어에 대한 AST-Grep 스캔을 지원해야 한다: Python, JavaScript, TypeScript, Go, Rust, Java, Kotlin, C/C++, Ruby, PHP, Swift, C#, Elixir, Scala.

**REQ-HOOK-141** [Unwanted]
시스템은 지원하지 않는 파일 확장자에 대해 스캔을 시도**하지 않아야 한다**.

---

## 4. Specifications (S)

### 4.1 Interface Definitions

```go
// Severity represents the severity level of a security finding.
type Severity string

const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
    SeverityHint    Severity = "hint"
)

// Finding represents a single security finding.
type Finding struct {
    RuleID    string   `json:"ruleId"`
    Severity  Severity `json:"severity"`
    Message   string   `json:"message"`
    File      string   `json:"file"`
    Line      int      `json:"line"`
    Column    int      `json:"column,omitempty"`
    EndLine   int      `json:"endLine,omitempty"`
    EndColumn int      `json:"endColumn,omitempty"`
    Code      string   `json:"code,omitempty"`
}

// ScanResult represents the result of a security scan.
type ScanResult struct {
    Scanned      bool       `json:"scanned"`
    ErrorCount   int        `json:"errorCount"`
    WarningCount int        `json:"warningCount"`
    InfoCount    int        `json:"infoCount"`
    Findings     []Finding  `json:"findings"`
    Error        string     `json:"error,omitempty"`
    Duration     time.Duration `json:"duration"`
}

// ASTGrepScanner handles AST-Grep security scanning.
type ASTGrepScanner interface {
    IsAvailable() bool
    Scan(ctx context.Context, filePath string, configPath string) (*ScanResult, error)
    ScanMultiple(ctx context.Context, filePaths []string, configPath string) ([]*ScanResult, error)
}

// RuleManager manages security rule configuration.
type RuleManager interface {
    FindRulesConfig(projectDir string) string
    LoadRules(configPath string) ([]string, error)
    GetDefaultRules() []string
}

// FindingReporter formats scan results for Claude.
type FindingReporter interface {
    FormatResult(result *ScanResult, filePath string) string
    FormatMultiple(results []*ScanResult) string
    ShouldExitWithError(result *ScanResult) bool
}
```

### 4.2 Supported Languages

| Language    | Extensions              | AST-Grep Support |
|-------------|-------------------------|------------------|
| Python      | .py, .pyi               | Yes              |
| JavaScript  | .js, .jsx, .mjs, .cjs   | Yes              |
| TypeScript  | .ts, .tsx, .mts, .cts   | Yes              |
| Go          | .go                     | Yes              |
| Rust        | .rs                     | Yes              |
| Java        | .java                   | Yes              |
| Kotlin      | .kt, .kts               | Yes              |
| C/C++       | .c, .cpp, .cc, .h, .hpp  | Yes              |
| Ruby        | .rb                     | Yes              |
| PHP         | .php                    | Yes              |
| Swift       | .swift                  | Yes              |
| C#          | .cs                     | Yes              |
| Elixir      | .ex, .exs               | Yes              |
| Scala       | .scala                  | Yes              |

### 4.3 OWASP Default Rules

```go
// DefaultOWASPRules provides built-in security rules.
const DefaultOWASPRules = `
id: owasp-security
message: OWASP Top 10 Security Vulnerabilities
severity: error
language: generic

rules:
  # SQL Injection
  - pattern: execute($$$QUERY)
    metavariables:
      QUERY: string containing user input

  # XSS
  - pattern: innerHTML = $INPUT
    metavariables:
      INPUT: user input

  # Hardcoded secrets
  - pattern: password = $PASSWORD
    metavariables:
      PASSWORD: string literal with length > 10

  # Insecure random
  - pattern: Math.random()
    message: Use crypto.randomBytes() instead

  # eval usage
  - pattern: eval($$$CODE)
    message: Avoid eval() for security reasons
`
```

### 4.4 Scan Output Format

```go
// Hook output format for security findings
type SecurityHookOutput struct {
    HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput"`
}

type HookSpecificOutput struct {
    HookEventName     string `json:"hookEventName"`
    AdditionalContext  string `json:"additionalContext"`
}

// Example additionalContext:
// "AST-Grep found 2 error(s), 1 warning(s) in main.py
//   - [ERROR] sql-injection: Potential SQL injection (line 45)
//   - [ERROR] hardcoded-secret: Hardcoded API key detected (line 12)
//   - [WARNING] weak-random: Math.random() used (line 78)"
```

### 4.5 Performance Requirements

| Metric                        | Target    | Measurement Method                     |
|-------------------------------|-----------|----------------------------------------|
| 단일 파일 스캔 (소형, <100줄)  | < 500ms   | Benchmark test                         |
| 단일 파일 스캔 (중형, <1000줄) | < 2s      | Benchmark test                         |
| 단일 파일 스캔 (대형, <5000줄) | < 10s     | Benchmark test                         |
| 규칙 로드                      | < 100ms   | Benchmark test                         |
| 결과 파싱                      | < 50ms    | Benchmark test                         |
| 메모리 사용량 (스캔 중)        | < 100MB   | Runtime profiling                      |

---

## 5. Traceability

### 5.1 Requirements to Files

| Requirement      | Implementation File            |
|------------------|-------------------------------|
| REQ-HOOK-100~103 | `ast_grep.go`                 |
| REQ-HOOK-110~112 | `rules.go`                    |
| REQ-HOOK-120~123 | `scanner.go`                  |
| REQ-HOOK-130~132 | `reporter.go`                 |
| REQ-HOOK-140~141 | `ast_grep.go` (language map)  |

### 5.2 Python Hook Mapping

| Python Script                       | Go Handler         | Status  |
|-------------------------------------|--------------------|---------|
| `post_tool__ast_grep_scan.py`       | `scanner.go`       | Planned |

### 5.3 Integration Points

- **SPEC-HOOK-001**: Registry, Protocol, Contract 재사용
- **SPEC-HOOK-002**: Tool Registry 패턴 재사용
- **internal/cli/hook.go**: PostToolUse 서브커맨드에서 스캐너 호출
- **.claude/skills/moai-tool-ast-grep/**: 보안 규칙 파일 위치

---

## Implementation Notes

**Status**: Draft
**Phase**: Phase 2 - Quality Integration

### Summary

AST-Grep integration for real-time security vulnerability detection in code. Supports 14+ programming languages with pattern-based scanning, OWASP Top 10 rules, and structured reporting to Claude Code. Runs as observational PostToolUse hook with graceful degradation when ast-grep is unavailable.

### Python Reference

- `post_tool__ast_grep_scan.py` (284 LOC)
- `moai_adk/astgrep/` package (~300 LOC)

### Estimated LOC

- `ast_grep.go`: ~300 LOC
- `scanner.go`: ~400 LOC
- `rules.go`: ~200 LOC
- `reporter.go`: ~200 LOC
- Test files: ~600 LOC
- Total: ~1,700 LOC
