---
spec_id: SPEC-HOOK-003
title: Security & Scanning - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, security, ast-grep"
---

# SPEC-HOOK-003: Acceptance Criteria

## 1. AST-Grep Integration Scenarios

### Scenario 1: Tool Availability Check

```gherkin
Feature: AST-Grep Tool Availability

  Scenario: Detect installed ast-grep
    Given ast-grep (sg) is installed in system PATH
    When IsAvailable is called
    Then true is returned
    And GetVersion returns version string like "0.20.0"

  Scenario: Graceful degradation when ast-grep not installed
    Given ast-grep is NOT installed
    When IsAvailable is called
    Then false is returned
    And Scan operation returns empty result with no error
```

### Scenario 2: Language Support

```gherkin
Feature: Multi-Language Support

  Scenario: Scan Python file
    Given a Python file "vulnerable.py" with security issues
    And ast-grep is available
    When Scan is called on "vulnerable.py"
    Then scan is executed successfully
    And findings are returned with Python-specific rules

  Scenario: Scan JavaScript file
    Given a JavaScript file "script.js" with XSS vulnerability
    And ast-grep is available
    When Scan is called on "script.js"
    Then scan is executed successfully
    And findings include XSS-related rule IDs

  Scenario: Skip unsupported file types
    Given a file "config.xml"
    When Scan is called on "config.xml"
    Then scan is skipped
    And ScanResult.Scanned is false
```

---

## 2. Rule Management Scenarios

### Scenario 3: Rule Discovery and Loading

```gherkin
Feature: Rule Configuration

  Scenario: Find project-specific rules
    Given a project with .ast-grep/sgconfig.yml
    When FindRulesConfig is called
    Then the path to sgconfig.yml is returned

  Scenario: Use default OWASP rules when no custom rules
    Given a project without sgconfig.yml
    And default OWASP rules are built-in
    When LoadRules is called
    Then default OWASP rules are returned
    And rules include SQL injection, XSS, hardcoded secrets patterns

  Scenario: Handle malformed rule file
    Given a sgconfig.yml with invalid YAML syntax
    When LoadRules is called
    Then an error is logged
    And default rules are returned as fallback
```

---

## 3. Scanner Scenarios

### Scenario 4: Vulnerability Detection

```gherkin
Feature: Security Scanning

  Scenario: Detect SQL injection vulnerability
    Given a Python file with execute(query) pattern
    And query contains user input
    When Scan is called
    Then a finding with ruleId "sql-injection" is returned
    And severity is "error"

  Scenario: Detect XSS vulnerability
    Given a JavaScript file with innerHTML = userInput
    When Scan is called
    Then a finding with ruleId "xss" is returned
    And severity is "error"

  Scenario: Detect hardcoded secrets
    Given a file with password = "hardcoded_secret_123"
    When Scan is called
    Then a finding with ruleId "hardcoded-secret" is returned
    And severity is "warning" or "error"

  Scenario: Multiple findings in single file
    Given a file with 5 security issues
    When Scan is called
    Then ScanResult.Findings contains 5 items
    And error/warning counts are accurate
```

### Scenario 5: Scan Timeout Handling

```gherkin
Feature: Scan Timeout

  Scenario: Handle large file scan timeout
    Given a file with 10,000 lines of code
    And scan timeout is set to 10 seconds
    When Scan is called
    And scan takes longer than timeout
    Then context timeout is triggered
    And ScanResult.Error contains "timeout"
    And execution continues without crash
```

---

## 4. Finding Reporting Scenarios

### Scenario 6: Result Formatting

```gherkin
Feature: Finding Reporting

  Scenario: Format single finding
    Given a scan result with 1 error finding
    When FormatResult is called
    Then output includes severity, rule ID, message, and line number
    And format is: "[ERROR] rule-id: message (line N)"

  Scenario: Format multiple findings with limit
    Given a scan result with 15 findings
    When FormatResult is called
    Then only top 10 findings are displayed
    And output includes "... and 5 more"

  Scenario: Aggregate results from multiple files
    Given scan results from 3 files
    When FormatMultiple is called
    Then output includes total error and warning counts
    And findings are grouped by file

  Scenario: Exit with code 2 on errors
    Given a scan result with errorCount > 0
    When ShouldExitWithError is called
    Then true is returned
    And Claude Code is prompted to address findings
```

---

## 5. Integration Scenarios

### Scenario 7: PostToolUse Hook Integration

```gherkin
Feature: PostToolUse Security Integration

  Scenario: Auto-scan after Write operation
    Given Claude Code triggers PostToolUse event
    And tool_name is "Write"
    And file_path is "vulnerable.py"
    And the file contains security issues
    When the security handler processes the event
    Then ast-grep scan is executed
    And findings are reported in additionalContext

  Scenario: Skip scan when disabled via environment variable
    Given MOAI_DISABLE_AST_GREP_SCAN is set to "true"
    When PostToolUse event is triggered
    Then scan is skipped
    And no security context is provided
```

---

## 6. Edge Case Scenarios

### Scenario 8: Error Handling

```gherkin
Feature: Error Handling

  Scenario: Handle JSON parsing failure
    Given ast-grep returns non-JSON output
    When Scan is called
    Then fallback regex parsing is attempted
    And scan continues if possible

  Scenario: Handle missing file
    Given a file path that does not exist
    When Scan is called
    Then ScanResult.Scanned is false
    And ScanResult.Error indicates file not found
    And execution continues without crash
```

---

## 7. Performance Scenarios

### Scenario 9: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: Scan small file under 500ms
    Given a file with 50 lines of code
    When Scan is executed
    Then execution time is less than 500ms

  Scenario: Scan medium file under 2 seconds
    Given a file with 500 lines of code
    When Scan is executed
    Then execution time is less than 2 seconds

  Scenario: Rule loading under 100ms
    When LoadRules is called
    Then execution time is less than 100ms
```

---

## 8. Definition of Done

### 8.1 Code Completion

- [ ] `internal/hook/security/` 패키지 구현 완료
- [ ] AST-Grep 도구 통합
- [ ] 규칙 관리자 구현
- [ ] 스캐너 구현
- [ ] 보고서 생성기 구현
- [ ] 14개 언어 지원

### 8.2 Test Completion

- [ ] 단위 테스트: 85%+ coverage
- [ ] 도구 가용성 테스트
- [ ] 취약점 탐지 테스트 (SQL injection, XSS, hardcoded secrets)
- [ ] 규칙 관리 테스트
- [ ] 결과 포맷팅 테스트
- [ ] 타임아웃 처리 테스트
- [ ] 크로스 플랫폼 테스트

### 8.3 Quality Gate

- [ ] `golangci-lint run ./internal/hook/security/...` 오류 0건
- [ ] `go vet ./internal/hook/security/...` 오류 0건
- [ ] `go test -race ./internal/hook/security/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 8.4 Integration

- [ ] PostToolUse 훅 통합 확인
- [ ] 환경 변수로 비활성화 기능 확인
- [ ] Claude Code와의 JSON 통신 확인

### 8.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 지원 언어 목록 문서화
- [ ] OWASP 규칙 문서화
- [ ] 환경 변수 문서화
