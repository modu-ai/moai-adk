---
spec_id: SPEC-HOOK-002
title: Code Quality Automation - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, formatter, linter"
---

# SPEC-HOOK-002: Acceptance Criteria

## 1. Tool Registry Scenarios

### Scenario 1: Tool Registration and Discovery

```gherkin
Feature: Tool Registry

  Scenario: Register and retrieve tools for a language
    Given a tool registry is initialized
    When registering Python tools (ruff-format, black, mypy)
    And querying tools for Python files with ToolTypeFormatter
    Then tools are returned sorted by priority (ruff-format first, then black)
    And each tool has correct command, args, and extensions

  Scenario: Check tool availability on system
    Given a tool registry with ruff and black registered
    And ruff is installed on the system
    And black is NOT installed on the system
    When checking tool availability
    Then ruff returns available=true
    And black returns available=false
```

### Scenario 2: Cross-Language Tool Support

```gherkin
Feature: Multi-Language Tool Support

  Scenario: Get tools for all supported languages
    Given a tool registry with 16 languages registered
    When querying tools for each language
    Then each language returns at least one formatter or linter
    And supported languages include:
      | python     | .py, .pyi    |
      | javascript | .js, .jsx    |
      | typescript | .ts, .tsx    |
      | go         | .go          |
      | rust       | .rs          |
      | java       | .java        |
      | kotlin     | .kt, .kts     |
      | swift      | .swift       |
      | cpp        | .c, .cpp, .h |
      | ruby       | .rb, .rake   |
      | php        | .php         |
      | elixir     | .ex, .exs     |
      | scala      | .scala, .sc  |
      | r          | .r, .R, .Rmd  |
      | dart       | .dart        |
      | csharp     | .cs          |
      | markdown   | .md, .mdx     |
```

---

## 2. Change Detection Scenarios

### Scenario 3: File Hash-Based Change Detection

```gherkin
Feature: Change Detection

  Scenario: Detect file modification after formatting
    Given a file "main.go" with initial content
    When computing file hash
    And storing the hash
    And running gofmt which modifies the file
    And computing file hash again
    Then the hashes are different
    And HasChanged returns true

  Scenario: Cache hash for performance
    Given a file "utils.py" with content
    When computing file hash twice
    Then the first computation reads the file
    And the second computation uses cached hash
    And execution time is significantly reduced

  Scenario: Handle non-existent files gracefully
    Given a file path "missing.py" that does not exist
    When computing file hash
    Then an error is returned
    And the error message is descriptive
```

---

## 3. Formatter Scenarios

### Scenario 4: Automatic Code Formatting

```gherkin
Feature: Code Formatter

  Scenario: Format Python file with ruff
    Given a tool registry with ruff-format available
    And a Python file "example.py" with unformatted code
    When FormatFile is called on "example.py"
    Then ruff-format is executed
    And the file is modified
    And ToolResult.FileModified is true
    And ToolResult.ToolName is "ruff-format"

  Scenario: Format Go file with gofmt
    Given a tool registry with gofmt available
    And a Go file "main.go" with unformatted code
    When FormatFile is called on "main.go"
    Then gofmt is executed with "-w" flag
    And the file is modified

  Scenario: Skip unsupported file types
    Given a file "config.json"
    When FormatFile is called on "config.json"
    Then ShouldFormat returns false
    And no formatter is executed

  Scenario: Skip files in build directories
    Given a file "node_modules/package/index.js"
    When FormatFile is called on this path
    Then ShouldFormat returns false
    And no formatter is executed
```

### Scenario 5: Graceful Degradation

```gherkin
Feature: Formatter Graceful Degradation

  Scenario: Handle missing formatter
    Given a file "script.rb"
    And no Ruby formatter is installed
    When FormatFile is called on "script.rb"
    Then ToolResult.Success is false
    And ToolResult.Error contains "formatter not available"
    And execution continues without crash

  Scenario: Handle formatter timeout
    Given a file "large.py"
    And the formatter takes longer than timeout
    When FormatFile is called with 30s timeout
    Then context timeout is triggered
    And ToolResult.Success is false
    And ToolResult.Error contains "timeout"
```

---

## 4. Linter Scenarios

### Scenario 6: Automatic Code Linting

```gherkin
Feature: Code Linter

  Scenario: Lint Python file and find issues
    Given a tool registry with ruff available
    And a Python file with unused imports
    When LintFile is called on the file
    Then ruff check is executed
    And ToolResult.IssuesFound is greater than 0
    And ToolResult.Output contains issue details

  Scenario: Auto-fix linting issues
    Given a tool registry with ruff --fix available
    And a Python file with fixable issues
    When AutoFix is called on the file
    Then ruff check --fix is executed
    And issues are automatically fixed
    And ToolResult.IssuesFixed is greater than 0

  Scenario: Provide issue summary to Claude
    Given a linter found 10 issues
    When formatting the result for Claude
    Then only top 5 issues are included
    And issues are sorted by severity (errors first)
    And summary includes total error and warning counts
```

---

## 5. Integration Scenarios

### Scenario 7: PostToolUse Hook Integration

```gherkin
Feature: PostToolUse Quality Integration

  Scenario: Format after Write operation
    Given Claude Code triggers PostToolUse event
    And tool_name is "Write"
    And tool_input contains file_path="main.go"
    And the file was written with unformatted code
    When the quality handler processes the event
    Then the file is auto-formatted
    And hook output contains "Auto-formatted with gofmt"

  Scenario: Format after Edit operation
    Given Claude Code triggers PostToolUse event
    And tool_name is "Edit"
    And tool_input contains file_path="utils.ts"
    And the edit added unformatted code
    When the quality handler processes the event
    Then the file is auto-formatted
    And hook output contains "Auto-formatted with prettier"

  Scenario: No formatting for unchanged files
    Given Claude Code triggers PostToolUse event
    And tool_name is "Write"
    And the file was already formatted
    When the quality handler processes the event
    Then suppressOutput is true
    And no additional context is provided
```

---

## 6. Cross-Platform Scenarios

### Scenario 8: Platform Compatibility

```gherkin
Feature: Cross-Platform Tool Execution

  Scenario: Execute formatter on Windows
    Given the system is Windows
    And gofmt is available in PATH
    When FormatFile is called on "main.go"
    Then gofmt is executed successfully
    And file paths use Windows path separators

  Scenario: Execute formatter on macOS
    Given the system is macOS
    And ruff is installed via Homebrew
    When FormatFile is called on "main.py"
    Then ruff is executed successfully
    And /usr/local/bin/ruff is found

  Scenario: Execute formatter on Linux
    Given the system is Linux
    And clang-format is available
    When FormatFile is called on "code.cpp"
    Then clang-format is executed successfully
```

---

## 7. Performance Scenarios

### Scenario 9: Performance Requirements

```gherkin
Feature: Performance Benchmarks

  Scenario: Format single file under 2 seconds
    Given a file with 1000 lines of code
    When FormatFile is executed
    Then execution time is less than 2 seconds
    And ToolResult.ExecutionTime is under 2000ms

  Scenario: Lint single file under 5 seconds
    Given a file with 1000 lines of code
    When LintFile is executed
    Then execution time is less than 5 seconds

  Scenario: Tool availability check under 10ms
    Given a tool in registry
    When IsToolAvailable is called
    Then execution time is less than 10ms for cached result

  Scenario: Hash computation under 5ms for 1MB file
    Given a file with 1MB of content
    When ComputeHash is executed
    Then execution time is less than 5ms
```

---

## 8. Security Scenarios

### Scenario 10: Security Validation

```gherkin
Feature: Security - Path Validation

  Scenario: Reject paths with null bytes
    Given a file path contains null byte: "test\x00.py"
    When FormatFile is called with this path
    Then an error is returned
    And no tool is executed

  Scenario: Reject paths with shell metacharacters
    Given a file path with pipe: "file.py | rm -rf"
    When FormatFile is called with this path
    Then the path is validated
    And shell metacharacters are rejected or escaped

  Scenario: Use exec.Command to prevent injection
    Given a tool configuration with args
    When RunTool is called
    Then exec.Command is used with separate args
    And shell=True is never used
```

---

## 9. Definition of Done

### 9.1 Code Completion

- [ ] `internal/hook/quality/` 패키지 구현 완료
- [ ] ToolRegistry 인터페이스 및 구현
- [ ] ChangeDetector 인터페이스 및 구현
- [ ] Formatter 핸들러 구현
- [ ] Linter 핸들러 구현
- [ ] 16개 언어 도구 등록

### 9.2 Test Completion

- [ ] 단위 테스트: 90%+ coverage
- [ ] 도구 등록/조회 테스트
- [ ] 포맷팅 기능 테스트 (주요 5개 언어)
- [ ] 린팅 기능 테스트 (주요 5개 언어)
- [ ] 변경 감지 테스트
- [ ] 건너뛰기 패턴 테스트
- [ ] 크로스 플랫폼 테스트

### 9.3 Quality Gate

- [ ] `golangci-lint run ./internal/hook/quality/...` 오류 0건
- [ ] `go vet ./internal/hook/quality/...` 오류 0건
- [ ] `go test -race ./internal/hook/quality/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 9.4 Integration

- [ ] PostToolUse 훅 통합 확인
- [ ] `moai hook post-tool` CLI 실행 확인
- [ ] Claude Code와의 JSON 통신 확인
- [ ] settings.json 생성기 통합 확인

### 9.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 지원 언어 목록 문서화
- [ ] 건너뛰기 패턴 문서화
- [ ] 성능 기준 문서화
