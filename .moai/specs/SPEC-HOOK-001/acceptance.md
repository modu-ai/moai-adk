---
spec_id: SPEC-HOOK-001
title: Compiled Hook System - Acceptance Criteria
version: "1.0.0"
created: 2026-02-03
updated: 2026-02-03
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, contract-test"
---

# SPEC-HOOK-001: Acceptance Criteria

## 1. Event Handler Scenarios

### Scenario 1: SessionStart -- Normal Session Initialization

```gherkin
Feature: SessionStart Event Handler

  Scenario: Normal session initialization with valid project
    Given moai binary is in the system PATH
    And a valid MoAI project exists at "/Users/goos/project"
    And .moai/config/sections/ contains valid YAML configuration
    When Claude Code triggers SessionStart event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "SessionStart",
        "project_dir": "/Users/goos/project"
      }
      """
    Then moai hook session-start exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And stdout JSON "data" field contains project name and version
    And stderr contains structured log entry with session_id "sess-abc-123"

  Scenario: SessionStart with missing configuration
    Given moai binary is in the system PATH
    And the working directory does NOT contain .moai/config/
    When Claude Code triggers SessionStart event
    Then moai hook session-start exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And stderr contains warning log about missing configuration
```

### Scenario 2: PreToolUse -- Security Policy Enforcement

```gherkin
Feature: PreToolUse Event Handler

  Scenario: Allowed tool execution passes security check
    Given moai binary is in the system PATH
    And security policy allows "Write" and "Edit" tools
    When Claude Code triggers PreToolUse event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "PreToolUse",
        "tool_name": "Write",
        "tool_input": {"file_path": "/Users/goos/project/src/main.go", "content": "package main"}
      }
      """
    Then moai hook pre-tool exits with code 0
    And stdout contains valid JSON with "decision": "allow"

  Scenario: Blocked tool execution is rejected
    Given moai binary is in the system PATH
    And security policy blocks "Bash" tool for destructive commands
    When Claude Code triggers PreToolUse event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "PreToolUse",
        "tool_name": "Bash",
        "tool_input": {"command": "rm -rf /"}
      }
      """
    Then moai hook pre-tool exits with code 2
    And stdout contains valid JSON with "decision": "block"
    And stdout JSON "reason" field explains the security violation
```

### Scenario 3: PostToolUse -- Metric Collection

```gherkin
Feature: PostToolUse Event Handler

  Scenario: Successful tool output metric collection
    Given moai binary is in the system PATH
    And a valid session is active
    When Claude Code triggers PostToolUse event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "PostToolUse",
        "tool_name": "Write",
        "tool_input": {"file_path": "main.go"},
        "tool_output": {"success": true, "path": "main.go"}
      }
      """
    Then moai hook post-tool exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And stdout JSON "data" field contains collected metrics
    And handler execution time is less than 100ms
```

### Scenario 4: SessionEnd -- Cleanup and Metric Persistence

```gherkin
Feature: SessionEnd Event Handler

  Scenario: Normal session end with metric persistence
    Given moai binary is in the system PATH
    And session metrics have been collected during the session
    When Claude Code triggers SessionEnd event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "SessionEnd",
        "project_dir": "/Users/goos/project"
      }
      """
    Then moai hook session-end exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And session metrics are persisted to .moai/memory/
    And temporary resources are cleaned up

  Scenario: SessionEnd with ranking submission enabled
    Given moai binary is in the system PATH
    And ranking submission is enabled in configuration
    When Claude Code triggers SessionEnd event
    Then moai hook session-end exits with code 0
    And stdout JSON "data" field indicates ranking submission status
```

### Scenario 5: Stop -- Graceful Shutdown

```gherkin
Feature: Stop Event Handler

  Scenario: Graceful shutdown with state preservation
    Given moai binary is in the system PATH
    And a Ralph loop controller has active state
    When Claude Code triggers Stop event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "Stop",
        "project_dir": "/Users/goos/project"
      }
      """
    Then moai hook stop exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And loop controller state is saved to disk
    And no data loss occurs during shutdown

  Scenario: Stop without active loop state
    Given moai binary is in the system PATH
    And no Ralph loop controller state exists
    When Claude Code triggers Stop event
    Then moai hook stop exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And stderr does NOT contain error-level logs
```

### Scenario 6: PreCompact -- Context Preservation

```gherkin
Feature: PreCompact Event Handler

  Scenario: Context snapshot before compaction
    Given moai binary is in the system PATH
    And a valid session with accumulated context exists
    When Claude Code triggers PreCompact event with stdin JSON:
      """
      {
        "session_id": "sess-abc-123",
        "cwd": "/Users/goos/project",
        "hook_event_name": "PreCompact",
        "project_dir": "/Users/goos/project"
      }
      """
    Then moai hook compact exits with code 0
    And stdout contains valid JSON with "decision": "allow"
    And context snapshot is saved to .moai/memory/
    And snapshot contains sufficient data for post-compaction recovery
```

---

## 2. Contract Test Scenarios (ADR-012)

### Scenario 7: Minimal PATH Contract

```gherkin
Feature: Hook Execution Contract -- Minimal PATH

  Scenario: Hook works with minimal PATH (non-interactive shell)
    Given moai binary is compiled and available
    And environment PATH is set to "/usr/bin:/bin" only
    And HOME is set to a temporary directory
    And no SHELL, USER, or LANG variables are set
    When executing "moai hook session-start" via exec.Command
    With stdin JSON: {"session_id":"test","cwd":"/tmp","hook_event_name":"SessionStart"}
    Then the process exits with code 0
    And stdout output passes json.Valid() check
    And no panic or runtime error occurs
```

### Scenario 8: JSON Round-Trip Contract

```gherkin
Feature: Hook Execution Contract -- JSON Round-Trip

  Scenario: JSON output is round-trip safe
    Given a valid HookOutput struct with Decision, Reason, and Data
    When the struct is serialized via json.Marshal()
    And the result is deserialized via json.Unmarshal()
    And re-serialized via json.Marshal()
    Then the first serialization and re-serialization produce identical JSON
    And json.Valid() returns true for all intermediate results
```

### Scenario 9: Non-Interactive Shell Contract

```gherkin
Feature: Hook Execution Contract -- Non-Interactive Shell

  Scenario: Hook works without shell initialization files
    Given moai binary is compiled and available
    And the process is started via exec.Command (no shell wrapper)
    And no .bashrc, .zshrc, or .profile is loaded
    And stdin contains a valid PostToolUse JSON payload:
      """
      {
        "session_id": "test",
        "cwd": "/tmp",
        "hook_event_name": "PostToolUse",
        "tool_name": "Write"
      }
      """
    When the hook command executes
    Then the process exits with code 0
    And stdout contains valid JSON
    And the hook does NOT depend on shell functions or aliases
```

### Scenario 10: Exit Code Contract

```gherkin
Feature: Hook Execution Contract -- Exit Codes

  Scenario: Exit code 0 for allowed actions
    Given a PreToolUse event for an allowed tool
    When the hook processes the event
    Then exit code is 0
    And stdout JSON contains "decision": "allow"

  Scenario: Exit code 2 for blocked actions
    Given a PreToolUse event for a blocked tool
    When the hook processes the event
    Then exit code is 2
    And stdout JSON contains "decision": "block"

  Scenario: Non-zero exit for runtime errors (non-blocking)
    Given a SessionStart event with corrupted configuration
    When the hook encounters a non-critical error
    Then exit code is 0 (non-blocking error)
    And stderr contains error log with details
    And stdout JSON contains "decision": "allow"
```

---

## 3. Edge Case Scenarios

### Scenario 11: Malformed JSON Input

```gherkin
Feature: Error Handling -- Malformed JSON

  Scenario: Completely invalid JSON on stdin
    Given moai binary is in the system PATH
    When stdin contains: "this is not json at all"
    And "moai hook session-start" is executed
    Then the process exits with code 1 (non-blocking error)
    And stderr contains "hook: invalid JSON input"
    And stdout is empty OR contains error JSON response

  Scenario: Valid JSON but missing required fields
    Given moai binary is in the system PATH
    When stdin contains: {"unknown_field": "value"}
    And "moai hook pre-tool" is executed
    Then the process exits with code 1
    And stderr contains validation error for missing session_id

  Scenario: JSON with extra unknown fields
    Given moai binary is in the system PATH
    When stdin contains valid hook JSON with additional unknown fields
    And "moai hook post-tool" is executed
    Then the process exits with code 0
    And unknown fields are silently ignored
    And known fields are correctly parsed
```

### Scenario 12: Timeout Handling

```gherkin
Feature: Error Handling -- Timeout

  Scenario: Handler exceeds configured timeout
    Given hook timeout is configured to 1 second
    And a handler is designed to take longer than 1 second
    When the hook dispatches the event
    Then the handler is cancelled via context
    And ErrHookTimeout error is returned
    And the process exits with code 1
    And stderr contains "hook: execution timed out"

  Scenario: Handler completes within timeout
    Given hook timeout is configured to 30 seconds (default)
    And all handlers complete within 100ms
    When the hook dispatches the event
    Then no timeout error occurs
    And the process exits with code 0
```

### Scenario 13: Empty stdin

```gherkin
Feature: Error Handling -- Empty stdin

  Scenario: Completely empty stdin (EOF immediately)
    Given moai binary is in the system PATH
    When stdin is empty (immediate EOF)
    And "moai hook session-start" is executed
    Then the process exits with non-zero code
    And stderr contains descriptive error message
    And the process does NOT hang or block indefinitely

  Scenario: stdin with only whitespace
    Given moai binary is in the system PATH
    When stdin contains only spaces and newlines
    And "moai hook pre-tool" is executed
    Then the process exits with non-zero code
    And stderr contains "hook: invalid JSON input"
```

### Scenario 14: Large Payload Handling

```gherkin
Feature: Error Handling -- Large Payloads

  Scenario: Very large tool_output in PostToolUse
    Given moai binary is in the system PATH
    And tool_output contains 1MB of JSON data
    When "moai hook post-tool" is executed
    Then the process exits with code 0
    And memory usage stays under 10MB
    And handler execution completes within 100ms

  Scenario: Deeply nested JSON input
    Given moai binary is in the system PATH
    And tool_input contains 100-level deep nested JSON
    When "moai hook pre-tool" is executed
    Then the process handles it without stack overflow
    And exits with code 0 or appropriate error
```

---

## 4. Performance Criteria

### Scenario 15: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: Single handler execution under 100ms
    Given benchmark test environment
    When BenchmarkSessionStartHandler is executed 1000 times
    Then average execution time is less than 100ms
    And P95 execution time is less than 150ms
    And no memory leaks detected (allocs per op stable)

  Scenario: Full dispatch cycle under 200ms
    Given benchmark test environment
    And all 6 handlers are registered
    When BenchmarkRegistryDispatch is executed 1000 times
    Then average full dispatch time is less than 200ms
    And P95 execution time is less than 300ms

  Scenario: JSON protocol under 1ms
    Given benchmark test environment
    When BenchmarkProtocolReadInput is executed 10000 times
    Then average parsing time is less than 1ms
    When BenchmarkProtocolWriteOutput is executed 10000 times
    Then average serialization time is less than 1ms

  Scenario: Contract validation under 1ms
    Given benchmark test environment
    When BenchmarkContractValidate is executed 10000 times
    Then average validation time is less than 1ms
    And overhead is negligible for every hook invocation

  Scenario: Memory usage under 10MB
    Given runtime profiling environment
    When a complete hook dispatch cycle executes
    Then peak memory allocation is less than 10MB
    And no goroutine leaks are detected
```

---

## 5. Cross-Platform Test Scenarios

### Scenario 16: Cross-Platform Compatibility

```gherkin
Feature: Cross-Platform Support

  Scenario: macOS arm64 (Apple Silicon) execution
    Given moai binary compiled for darwin/arm64
    And running on macOS with Apple Silicon
    When all 6 hook events are executed sequentially
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON
    And file paths use "/" separator correctly

  Scenario: macOS amd64 (Intel) execution
    Given moai binary compiled for darwin/amd64
    And running on macOS with Intel processor
    When all 6 hook events are executed sequentially
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON

  Scenario: Linux amd64 execution
    Given moai binary compiled for linux/amd64
    And running on Ubuntu/Debian Linux
    When all 6 hook events are executed sequentially
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON
    And file paths use "/" separator correctly

  Scenario: Linux arm64 execution
    Given moai binary compiled for linux/arm64
    And running on ARM-based Linux (AWS Graviton, Raspberry Pi)
    When all 6 hook events are executed sequentially
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON

  Scenario: Windows amd64 execution
    Given moai binary compiled for windows/amd64
    And running on Windows 10/11
    When all 6 hook events are executed via cmd.exe
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON
    And no SIGALRM-related errors occur (Windows does not support SIGALRM)
    And file paths handle both "/" and "\" correctly via filepath.Clean()

  Scenario: Windows arm64 execution
    Given moai binary compiled for windows/arm64
    And running on Windows ARM device
    When all 6 hook events are executed
    Then all hooks exit with expected codes
    And all stdout outputs are valid JSON
```

---

## 6. Registry Integration Scenarios

### Scenario 17: Registry Dispatch Behavior

```gherkin
Feature: Registry Dispatch Logic

  Scenario: Empty registry returns allow
    Given a registry with no handlers registered
    When Dispatch is called for EventSessionStart
    Then HookOutput with Decision "allow" is returned
    And no error occurs

  Scenario: Multiple handlers execute in order
    Given handlers A, B, C are registered for EventPostToolUse
    When Dispatch is called for EventPostToolUse
    Then handler A executes first
    And handler B executes second
    And handler C executes third
    And final result aggregates all handler outputs

  Scenario: Block short-circuits remaining handlers
    Given handlers A (allow), B (block), C (allow) are registered for EventPreToolUse
    When Dispatch is called for EventPreToolUse
    Then handler A executes and returns allow
    And handler B executes and returns block
    And handler C does NOT execute
    And Dispatch returns Decision "block" from handler B

  Scenario: Handler error stops dispatch chain
    Given handlers A (allow), B (error), C (allow) are registered
    When Dispatch is called
    Then handler A executes successfully
    And handler B returns an error
    And handler C does NOT execute
    And Dispatch returns the error from handler B
```

---

## 7. Definition of Done

### 7.1 Code Completion

- [ ] `internal/hook/` 패키지 내 9개 파일 구현 완료
- [ ] 모든 인터페이스(`Handler`, `Registry`, `Protocol`, `Contract`) 구현
- [ ] 6개 이벤트 핸들러 구현 및 Registry 등록
- [ ] Sentinel errors 정의 (`errors.go`)
- [ ] `internal/cli/hook.go` Cobra 서브커맨드 통합

### 7.2 Test Completion

- [ ] 단위 테스트: `internal/hook/` 전체 95% 이상 coverage
- [ ] Contract 테스트: Minimal PATH, JSON round-trip, non-interactive shell 모두 통과
- [ ] Benchmark 테스트: 모든 성능 기준 충족 (handler < 100ms, protocol < 1ms)
- [ ] Fuzz 테스트: JSON 입력 파싱 fuzz test 작성
- [ ] Edge case: malformed JSON, empty stdin, timeout, large payload 모두 처리
- [ ] 테이블 기반 테스트: Go 관용적 테스트 패턴 적용

### 7.3 Quality Gate

- [ ] `golangci-lint run ./internal/hook/...` 오류 0건
- [ ] `go vet ./internal/hook/...` 오류 0건
- [ ] `go test -race ./internal/hook/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드
- [ ] `gofumpt` 포매팅 적용

### 7.4 Cross-Platform

- [ ] CI 매트릭스 6 플랫폼(darwin/linux/windows x amd64/arm64) 전체 통과
- [ ] Windows: SIGALRM 미사용 확인, cmd.exe 실행 확인
- [ ] 경로 처리: `filepath.Clean()` + `filepath.Join()` 일관 사용

### 7.5 Integration

- [ ] `moai hook session-start` CLI 실행 확인
- [ ] `moai hook pre-tool` CLI 실행 확인
- [ ] `moai hook post-tool` CLI 실행 확인
- [ ] `moai hook session-end` CLI 실행 확인
- [ ] `moai hook stop` CLI 실행 확인
- [ ] `moai hook compact` CLI 실행 확인
- [ ] settings.json 생성기와의 통합 확인 (ADR-011)

### 7.6 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] Contract 보증/비보증 사항 코드 내 문서화
- [ ] 오류 코드 및 exit code 의미론 문서화
