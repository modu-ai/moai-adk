---
spec_id: SPEC-HOOK-007
title: Session Lifecycle Enhancements - Acceptance Criteria
version: "0.1.0"
created: 2026-02-04
updated: 2026-02-04
author: GOOS
tags: "hook, acceptance-criteria, given-when-then, session-lifecycle"
---

# SPEC-HOOK-007: Acceptance Criteria

## 1. SessionStart Scenarios

### Scenario 1: Enhanced Project Information Display

```gherkin
Feature: Enhanced SessionStart

  Scenario: Display comprehensive project info for Git repository
    Given a project with Git repository
    And .moai/config contains project metadata
    When SessionStart event is triggered
    Then project information is displayed with:
      | Project name  | my-project       |
      | Version       | v1.0.0           |
      | Git branch    | main             |
      | Last commit   | feat: Add auth   |
      | Modified files| 3                |
    And MoAI version is shown
    And a formatted welcome banner is displayed

  Scenario: Handle project without Git repository
    Given a project without Git repository
    And .moai/config contains project metadata
    When SessionStart event is triggered
    Then project name and version are displayed
    And Git information is omitted
    And welcome banner is still displayed

  Scenario: Restore previous work state
    Given a previous session state exists
    And last-session-state.json contains:
      | lastPosition | src/main.go:45 |
      | activeFiles  | src/main.go, src/utils.go |
    When SessionStart event is triggered
    Then work state is loaded
    And a message shows "Resuming from previous session"
    And active files are listed
```

---

## 2: SessionEnd Scenarios

### Scenario 2: Auto-Cleanup

```gherkin
Feature: Automatic SessionEnd Cleanup

  Scenario: Clean temporary files
    Given a session with temporary files in .moai/temp/
    When SessionEnd event is triggered
    Then all files in .moai/temp/ are deleted
    And .moai/temp/ directory is removed if empty
    And deleted file count is reported

  Scenario: Clean session-specific caches
    Given a session with cache in .moai/cache/temp/
    When SessionEnd event is triggered
    Then session cache is cleared
    And shared cache is preserved

  Scenario: Generate cleanup report
    Given cleanup operations completed
    When cleanup report is generated
    Then report includes:
      | Files deleted | 5            |
      | Dirs deleted  | 2            |
      | Bytes freed  | > 0          |
    And any errors during cleanup are listed

  Scenario: Handle cleanup errors gracefully
    Given a file that cannot be deleted
    When cleanup is performed
    Then deletion error is logged
    And remaining cleanup continues
    And session still ends normally
```

---

## 3: Metrics Collection Scenarios

### Scenario 3: Session Metrics

```gherkin
Feature: Session Metrics Collection

  Scenario: Track tool usage during session
    Given a session with multiple tool invocations
    When various tools are used (Read, Write, Edit, Bash)
    Then each tool use is counted
    And counts are stored in session metrics

  Scenario: Save session metrics on end
    Given a session that has been active for 2 hours
    With 45 Read operations, 12 Write operations, 8 Edit operations
    When SessionEnd is triggered
    Then session-stats.json is created with:
      | toolUseCount  | Read:45, Write:12, Edit:8 |
      | duration     | approximately 2h            |
      | filesModified | 20                          |

  Scenario: Load previous session metrics for comparison
    Given a previous session with 100 tool uses
    When current session reaches 50 tool uses
    Then metrics can be compared across sessions
```

---

## 4: Work State Persistence Scenarios

### Scenario 4: Work State Save and Restore

```gherkin
Feature: Work State Persistence

  Scenario: Save work state during session
    Given a user is editing src/main.go at line 45
    When PostToolUse event is triggered
    Then work state is saved with:
      | lastPosition | src/main.go:45  |
      | activeFiles  | src/main.go      |
      | timestamp    | current time     |

  Scenario: Restore work state on new session
    Given a previous session state exists
    With lastPosition = src/main.go:45
    When a new session starts
    Then a message indicates restored position
    And cursor position can be restored if supported

  Scenario: Handle missing state file gracefully
    Given no previous session state exists
    When session starts
    Then no restoration message is shown
    And session starts normally
```

---

## 5: Integration Scenarios

### Scenario 5: Hook Integration

```gherkin
Feature: Hook Integration

  Scenario: SessionStart integration
    Given moai hook session-start is configured
    When Claude Code starts a session
    Then SessionEnhancer is invoked
    And project info is collected via GitOperationsManager
    And welcome message is displayed to user

  Scenario: SessionEnd integration
    Given moai hook session-end is configured
    When Claude Code ends a session
    Then SessionCleanup is invoked
    And MetricsCollector saves metrics
    And WorkState is persisted
```

---

## 6: Edge Case Scenarios

### Scenario 6: Error Handling

```gherkin
Feature: Error Handling

  Scenario: Handle missing .moai directory
    Given a project without .moai directory
    When SessionStart is triggered
    Then .moai/ directory is created
    And default config is initialized

  Scenario: Handle permission denied during cleanup
    Given a temporary file without delete permission
    When cleanup is attempted
    Then error is logged
    And cleanup continues with other files
```

---

## 7: Performance Scenarios

### Scenario 7: Performance Benchmarks

```gherkin
Feature: Performance Requirements

  Scenario: SessionStart processing under 500ms
    Given a project with Git repository
    When SessionStart hook is executed
    Then total processing time is less than 500ms

  Scenario: SessionEnd processing under 2 seconds
    Given a session with temporary files
    When SessionEnd hook is executed
    Then cleanup and metrics save complete in under 2 seconds

  Scenario: State save and load under 100ms combined
    Given a work state with several active files
    When saving and loading state
    Then combined operation takes less than 100ms
```

---

## 8: Definition of Done

### 8.1 Code Completion

- [ ] `internal/hook/lifecycle/` 패키지 구현 완료
- [ ] SessionEnhancer 구현
- [ ] SessionCleanup 구현
- [ ] MetricsCollector 구현
- [ ] WorkState persistence 구현

### 8.2 Test Completion

- [ ] 단위 테스트: 85%+ coverage
- [ ] SessionStart 향상 테스트
- [ ] SessionEnd 정리 테스트
- [ ] 메트릭 수집 테스트
- [ ] 상태 저장/복구 테스트

### 8.3 Quality Gate

- [ ] `golangci-lint run ./internal/hook/lifecycle/...` 오류 0건
- [ ] `go vet ./internal/hook/lifecycle/...` 오류 0건
- [ ] `go test -race ./internal/hook/lifecycle/...` 경쟁 조건 0건
- [ ] godoc 주석: 모든 exported 타입, 함수, 메서드

### 8.4 Integration

- [ ] SessionStart 훅 통합 확인
- [ ] SessionEnd 훅 통합 확인
- [ ] Git Operations Manager 통합 확인
- [ ] Config Manager 통합 확인

### 8.5 Documentation

- [ ] 각 파일에 package-level godoc 주석
- [ ] 세션 라이프사이클 흐름 문서화
- [ ] 메트릭 형식 문서화
- [ ] 상태 파일 형식 문서화
