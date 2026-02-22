# Acceptance Criteria: SPEC-UPDATE-001

## Overview

Global/Local Installation Mode Redesign 기능에 대한 상세 수용 기준입니다.

---

## Test Scenarios

### Scenario 1: moai init - New Project with Global Mode

Given: 사용자가 .moai/ 디렉토리가 없는 새 프로젝트 디렉토리에 있음
When: 사용자가 moai init . 실행
Then:
- 시스템이 정상적으로 초기화 프로세스 진행
- 설치 모드 선택 질문이 표시됨 (Global 권장 / Local)
- Global 모드 선택 시:
  - ~/.claude/agents/moai/ 생성됨
  - ~/.claude/skills/moai*/ 생성됨
  - ~/.claude/rules/moai/ 생성됨
  - .claude/hooks/moai/ 생성됨
  - .claude/settings.json 생성됨
  - .moai/config/sections/system.yaml에 installation_mode: "global" 저장됨
- 프로젝트 .claude/agents/moai/는 생성되지 않음
- 프로젝트 .claude/skills/moai*/는 생성되지 않음

```gherkin
Feature: Global Mode Installation

  Scenario: Initialize new project with Global mode
    Given a directory without .moai/
    When user runs "moai init ."
    And user selects "Global (recommended)" mode
    Then ~/.claude/agents/moai/ should exist
    And ~/.claude/skills/moai-foundation-core/ should exist
    And ~/.claude/rules/moai/ should exist
    And .claude/hooks/moai/ should exist
    And .claude/settings.json should exist
    And .claude/agents/moai/ should not exist
    And .moai/config/sections/system.yaml should contain "installation_mode: global"
```

---

### Scenario 2: moai init - New Project with Local Mode

Given: 사용자가 .moai/ 디렉토리가 없는 새 프로젝트 디렉토리에 있음
When: 사용자가 moai init . 실행 후 Local 모드 선택
Then:
- .claude/agents/moai/ 생성됨
- .claude/skills/moai*/ 생성됨
- .claude/rules/moai/ 생성됨
- .claude/hooks/moai/ 생성됨
- .claude/settings.json 생성됨
- .moai/config/sections/system.yaml에 installation_mode: "local" 저장됨

```gherkin
Feature: Local Mode Installation

  Scenario: Initialize new project with Local mode
    Given a directory without .moai/
    When user runs "moai init ."
    And user selects "Local" mode
    Then .claude/agents/moai/ should exist
    And .claude/skills/moai-foundation-core/ should exist
    And .claude/rules/moai/ should exist
    And .claude/hooks/moai/ should exist
    And .claude/settings.json should exist
    And .moai/config/sections/system.yaml should contain "installation_mode: local"
```

---

### Scenario 3: moai init - Already Initialized Project

Given: 사용자가 이미 .moai/ 디렉토리가 있는 프로젝트에 있음
When: 사용자가 moai init . 실행
Then:
- 정상적으로 초기화 프로세스 진행
- 기존 installation_mode 설정 유지
- 사용자가 모드 변경 선택 시 전환 로직 실행

```gherkin
Feature: Re-initialization

  Scenario: Re-initialize existing project
    Given a directory with .moai/ and installation_mode: local
    When user runs "moai init ."
    Then the installation should proceed normally
    And installation_mode should remain "local" unless user changes it
```

---

### Scenario 4: moai update - Reset Configuration (Yes)

Given: 사용자가 초기화된 프로젝트에 있음
When: 사용자가 moai update 실행
Then:
- "설정을 재설정하시겠습니까? (y/n)" 질문 표시
- 'y' 응답 시 설정 wizard 재실행
- Global 모드 변경 질문 표시
- 설정 파일이 새 값으로 덮어씌워짐

```gherkin
Feature: Update with Configuration Reset

  Scenario: Reset configuration during update
    Given an initialized project
    When user runs "moai update"
    And user answers "y" to "Reset configuration?"
    Then configuration wizard should start
    And user can change installation mode
```

---

### Scenario 5: moai update - Keep Configuration (No)

Given: 사용자가 초기화된 프로젝트에 있음
When: 사용자가 moai update 실행 후 'n' 응답
Then:
- 기존 설정 백업 생성
- 템플릿 파일 병합 (3-way merge)
- Global 모드 변경 질문 표시

```gherkin
Feature: Update with Configuration Merge

  Scenario: Merge configuration during update
    Given an initialized project
    When user runs "moai update"
    And user answers "n" to "Reset configuration?"
    Then backup should be created in .moai-backups/
    And template files should be merged
    And user should be asked about changing installation mode
```

---

### Scenario 6: moai update - Remove -c Flag

Given: 사용자가 초기화된 프로젝트에 있음
When: 사용자가 moai update -c 실행
Then:
- -c 플래그가 더 이상 지원되지 않음을 알리는 메시지 또는 무시
- 또는 -c 플래그 없이 동일한 동작 수행 (설정 재설정 질문으로 대체)

```gherkin
Feature: Deprecated -c Flag

  Scenario: -c flag is removed
    Given an initialized project
    When user runs "moai update -c"
    Then the command should proceed without -c flag effect
    And user should be asked "Reset configuration?" instead
```

---

### Scenario 7: moai glm - Settings Local Only

Given: 사용자가 초기화된 프로젝트에 있음
When: 사용자가 moai glm 실행
Then:
- .claude/settings.local.json 생성 또는 수정됨
- .claude/settings.json은 수정되지 않음
- GLM 환경 변수가 settings.local.json의 env 섹션에 추가됨

```gherkin
Feature: GLM Settings Isolation

  Scenario: GLM modifies only settings.local.json
    Given an initialized project
    When user runs "moai glm"
    Then .claude/settings.local.json should be modified
    And .claude/settings.json should not be modified
    And ANTHROPIC_* variables should be in settings.local.json env section
```

---

### Scenario 8: Mode Transition - Local to Global

Given: Local 모드로 초기화된 프로젝트
When: 사용자가 moai update 실행 후 Global 모드로 변경 선택
Then:
- .claude/agents/moai/ 삭제됨
- .claude/skills/moai*/ 삭제됨
- .claude/rules/moai/ 삭제됨
- .claude/hooks/moai/ 유지됨
- .claude/settings.json 유지됨
- .claude/settings.local.json 유지됨
- .moai/ 유지됨
- ~/.claude/에 시스템 파일 설치됨

```gherkin
Feature: Local to Global Transition

  Scenario: Switch from Local to Global mode
    Given a project with installation_mode: local
    When user runs "moai update"
    And user selects to change to Global mode
    Then .claude/agents/moai/ should be deleted
    And .claude/skills/moai*/ should be deleted
    And .claude/rules/moai/ should be deleted
    And .claude/hooks/moai/ should exist
    And .claude/settings.json should exist
    And .moai/ should exist
    And ~/.claude/agents/moai/ should exist
```

---

### Scenario 9: Mode Transition - Global to Local

Given: Global 모드로 초기화된 프로젝트
When: 사용자가 moai update 실행 후 Local 모드로 변경 선택
Then:
- ~/.claude/에서 .claude/agents/moai/로 복사됨
- ~/.claude/에서 .claude/skills/moai*/로 복사됨
- ~/.claude/에서 .claude/rules/moai/로 복사됨
- 기존 hooks, settings 파일 유지됨

```gherkin
Feature: Global to Local Transition

  Scenario: Switch from Global to Local mode
    Given a project with installation_mode: global
    When user runs "moai update"
    And user selects to change to Local mode
    Then .claude/agents/moai/ should be created
    And .claude/skills/moai-foundation-core/ should be created
    And .claude/rules/moai/ should be created
    And existing hooks should be preserved
    And existing settings.json should be preserved
```

---

### Scenario 10: Global Mode - Multiple Projects

Given: Global 모드로 초기화된 두 개의 프로젝트
When: 한 프로젝트에서 moai update 실행
Then:
- ~/.claude/의 시스템 파일이 업데이트됨
- 두 프로젝트 모두 동일한 버전의 시스템 파일 사용
- 각 프로젝트의 settings.local.json은 독립적

```gherkin
Feature: Shared Global Installation

  Scenario: Multiple projects share Global installation
    Given project A with installation_mode: global
    And project B with installation_mode: global
    When user runs "moai update" in project A
    Then ~/.claude/agents/moai/ should be updated
    And project B should see the updated agents
    And project A settings should be independent from project B settings
```

---

## Quality Gate Criteria

### Functional Requirements

- [ ] 모든 Given-When-Then 시나리오 통과
- [ ] Global/Local 모드 전환 정상 동작
- [ ] GLM 설정 격리 정상 동작
- [ ] 기존 프로젝트 호환성 유지

### Non-Functional Requirements

- [ ] moai init 실행 시간 5초 이내
- [ ] moai update 실행 시간 10초 이내
- [ ] 메모리 사용량 100MB 이내

### Security Requirements

- [ ] GLM API 키가 settings.json에 저장되지 않음
- [ ] settings.local.json이 .gitignore에 권장됨
- [ ] Global 설치 파일이 적절한 권한으로 생성됨

### Compatibility Requirements

- [ ] macOS에서 정상 동작
- [ ] Linux에서 정상 동작
- [ ] Windows에서 정상 동작
- [ ] 기존 Local 모드 프로젝트와 호환

---

## Test Data

### Sample Project Structure (Global Mode)

test-project/
  .claude/
    hooks/
      moai/
        handle-session-start.sh
    settings.json
    settings.local.json (optional)
  .moai/
    config/
      sections/
        system.yaml (installation_mode: global)
        language.yaml
        ...
    specs/

### Sample system.yaml (Global Mode)

moai:
  version: "v2.5.0"
  template_version: "v2.5.0"
  installation_mode: "global"
  global_path: "~/.claude"

---

## Definition of Done

- [ ] 모든 테스트 시나리오 통과
- [ ] 코드 커버리지 85% 이상
- [ ] 코드 리뷰 완료
- [ ] TRUST 5 품질 게이트 통과
- [ ] 문서 업데이트 완료

---

Version: 1.0.0
Last Updated: 2026-02-21
