---
spec_id: SPEC-HOOK-004
title: LSP Diagnostics Integration - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, lsp, diagnostics"
---

# SPEC-HOOK-004: Acceptance Criteria

## 1. LSP Diagnostics Scenarios

### Scenario 1: LSP Diagnostics Collection

```gherkin
Feature: LSP Diagnostics Collection

  Scenario: Collect diagnostics from LSP server
    Given an LSP server is running for the language
    And a file "main.go" with type errors
    When GetDiagnostics is called on "main.go"
    Then diagnostics are returned
    And at least one diagnostic has severity "error"
    And each diagnostic includes range, message, and source

  Scenario: Handle LSP server unavailable gracefully
    Given no LSP server is running
    When GetDiagnostics is called
    Then fallback CLI tools are used
    And diagnostics are returned from CLI tools
```

### Scenario 2: Fallback CLI Tools

```gherkin
Feature: Fallback CLI Tools

  Scenario: Use ruff for Python when LSP unavailable
    Given a Python file "utils.py" with linting issues
    And no LSP server is running
    When GetDiagnostics is called
    Then ruff check --output-format=json is executed
    And diagnostics are parsed from JSON output

  Scenario: Use tsc for TypeScript when LSP unavailable
    Given a TypeScript file "index.ts" with type errors
    And no LSP server is running
    When GetDiagnostics is called
    Then tsc --noEmit is executed
    And diagnostics are parsed from stderr

  Scenario: Handle all fallback tools unavailable
    Given a file with unsupported language
    And no LSP server is running
    And no CLI tools are available
    When GetDiagnostics is called
    Then "diagnostics unavailable" message is returned
```

---

## 2. Regression Detection Scenarios

### Scenario 3: Baseline Tracking

```gherkin
Feature: Regression Detection

  Scenario: Save diagnostic baseline
    Given a file "service.go" with 3 errors
    When SaveBaseline is called
    Then baseline is saved to .moai/memory/diagnostics-baseline.json
    And file hash is recorded for change detection

  Scenario: Detect regression when errors increase
    Given a baseline with 2 errors for "service.go"
    And current diagnostics have 4 errors
    When CompareWithBaseline is called
    Then RegressionReport.HasRegression is true
    And RegressionReport.NewErrors is 2
    And regression warning is included in output

  Scenario: Detect improvement when errors decrease
    Given a baseline with 5 errors for "service.go"
    And current diagnostics have 2 errors
    When CompareWithBaseline is called
    Then RegressionReport.HasImprovement is true
    And RegressionReport.FixedErrors is 3
    And improvement message is included in output
```

---

## 3. Quality Gate Scenarios

### Scenario 4: Quality Gate Enforcement

```gherkin
Feature: Quality Gate Enforcement

  Scenario: Block when errors exceed threshold
    Given quality gate with maxErrors=5
    And current diagnostics have 10 errors
    When ShouldBlock is called
    Then true is returned
    And exit code 2 is returned to Claude

  Scenario: Allow when errors within threshold
    Given quality gate with maxErrors=10
    And current diagnostics have 5 errors
    When ShouldBlock is called
    Then false is returned
    And exit code 0 is returned

  Scenario: Load quality gate from config
    Given .moai/config/sections/quality.yaml exists
    With content:
      """
      lsp_quality_gates:
        max_errors: 0
        max_warnings: 10
      """
    When LoadConfig is called
    Then QualityGate.MaxErrors is 0
    And QualityGate.MaxWarnings is 10
```

---

## 4. Integration Scenarios

### Scenario 5: PostToolUse Integration

```gherkin
Feature: PostToolUse LSP Integration

  Scenario: Provide diagnostics after Write operation
    Given Claude Code triggers PostToolUse event
    And tool_name is "Write"
    And file_path is "main.go"
    And the file has type errors
    When the LSP handler processes the event
    Then diagnostics are collected
    And additionalContext contains error summary
    And top 5 diagnostics are listed

  Scenario: Disable diagnostics via environment variable
    Given MOAI_DISABLE_LSP_DIAGNOSTIC is set to "true"
    When PostToolUse event is triggered
    Then diagnostics are skipped
    And no diagnostic context is provided
```

---

## 5. Session Tracking Scenarios

### Scenario 6: Session Statistics

```gherkin
Feature: Session Statistics Tracking

  Scenario: Track diagnostics across session
    Given multiple files have been modified during session
    When SessionEnd event is triggered
    Then session statistics are aggregated
    And summary includes total errors and warnings
    And summary includes files with most issues

  Scenario: Persist session statistics
    Given session with diagnostic activity
    When session ends
    Then statistics are saved to .moai/memory/
    And next session can access previous session data
```

---

## 6. Edge Case Scenarios

### Scenario 7: Error Handling

```gherkin
Feature: Error Handling

  Scenario: Handle LSP timeout
    Given LSP server takes longer than timeout
    When GetDiagnostics is called
    Then timeout error is logged
    And fallback CLI tools are attempted

  Scenario: Handle malformed CLI output
    Given CLI tool returns unexpected format
    When fallback diagnostics are collected
    Then parsing error is logged
    And empty diagnostics are returned
```

---

## 7. Performance Scenarios

### Scenario 8: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: LSP diagnostics under 3 seconds
    Given a file with 500 lines of code
    And LSP server is running
    When GetDiagnostics is executed
    Then execution time is less than 3 seconds

  Scenario: Fallback diagnostics under 5 seconds
    Given a file with 500 lines of code
    And LSP server is NOT running
    When GetDiagnostics is executed
    Then fallback tools complete in under 5 seconds

  Scenario: Baseline operations under 50ms
    Given a baseline file exists
    When SaveBaseline or CompareWithBaseline is executed
    Then execution time is less than 50ms
```

---

## 8. Definition of Done

### 8.1 Code Completion

- [ ] `internal/lsp/hook/` 패키지 구현 완료
- [ ] LSP 진단 수집기 구현
- [ ] Fallback CLI 도구 구현
- [ ] 회귀 트래커 구현
- [ ] 품질 게이트 강제자 구현

### 8.2 Test Completion

- [ ] 단위 테스트: 85%+ coverage
- [ ] LSP 진단 수집 테스트
- [ ] Fallback 도구 테스트 (ruff, tsc, go vet)
- [ ] 회귀 탐지 테스트
- [ ] 품질 게이트 테스트
- [ ] 세션 통계 테스트

### 8.3 Quality Gate

- [ ] `golangci-lint run ./internal/lsp/hook/...` 오류 0건
- [ ] `go vet ./internal/lsp/hook/...` 오류 0건
- [ ] `go test -race ./internal/lsp/hook/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 8.4 Integration

- [ ] PostToolUse 훅 통합 확인
- [ ] 환경 변수로 비활성화 기능 확인
- [ ] internal/lsp 패키지 통합 확인
- [ ] Claude Code와의 JSON 통신 확인

### 8.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 지원되는 언어 및 fallback 도구 문서화
- [ ] 품질 게이트 설정 형식 문서화
- [ ] 환경 변수 문서화
