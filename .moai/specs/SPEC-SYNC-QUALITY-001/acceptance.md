---
id: SPEC-SYNC-QUALITY-001
version: "1.0.0"
status: "draft"
created: "2025-12-22"
updated: "2025-12-22"
author: "GOOS"
priority: "high"
---

# 인수 기준: Phase 0.5 품질 검증 통합

## 개요

`/moai:3-sync` 커맨드의 Phase 0.5 품질 검증 통합에 대한 인수 기준입니다.

## 테스트 시나리오

### 시나리오 1: 모든 품질 검증 통과

```gherkin
Feature: Phase 0.5 품질 검증 - 전체 통과

  Scenario: 모든 도구가 설치되어 있고 모든 검증이 통과하는 경우
    Given pytest, ruff, mypy가 설치되어 있다
    And 모든 테스트가 통과한다
    And 린팅 이슈가 없다
    And 타입 오류가 없다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then Phase 0.5 품질 검증이 시작된다
    And pytest가 성공적으로 실행된다
    And ruff가 성공적으로 실행된다
    And mypy가 성공적으로 실행된다
    And code-review 에이전트가 호출된다
    And 품질 검증 결과 보고서가 표시된다
    And 문서 동기화가 진행된다
```

### 시나리오 2: pytest 실패 시 사용자 선택 - 계속 진행

```gherkin
Feature: Phase 0.5 품질 검증 - pytest 실패 후 계속

  Scenario: pytest가 실패하고 사용자가 계속 진행을 선택하는 경우
    Given pytest가 설치되어 있다
    And 일부 테스트가 실패한다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then pytest가 실행되고 실패 결과가 표시된다
    And 사용자에게 "계속 진행" 또는 "중단" 선택지가 제공된다
    When 사용자가 "계속 진행"을 선택한다
    Then ruff 검증이 계속 진행된다
    And mypy 검증이 진행된다
    And code-review 에이전트가 호출된다
    And 문서 동기화가 진행된다
    And 최종 보고서에 pytest 실패가 경고로 표시된다
```

### 시나리오 3: pytest 실패 시 사용자 선택 - 중단

```gherkin
Feature: Phase 0.5 품질 검증 - pytest 실패 후 중단

  Scenario: pytest가 실패하고 사용자가 중단을 선택하는 경우
    Given pytest가 설치되어 있다
    And 일부 테스트가 실패한다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then pytest가 실행되고 실패 결과가 표시된다
    And 사용자에게 "계속 진행" 또는 "중단" 선택지가 제공된다
    When 사용자가 "중단"을 선택한다
    Then sync 프로세스가 중단된다
    And "테스트 수정 후 다시 시도하세요" 메시지가 표시된다
    And 문서 동기화가 실행되지 않는다
```

### 시나리오 4: 도구 미설치 시 경고 후 계속 진행

```gherkin
Feature: Phase 0.5 품질 검증 - 도구 미설치

  Scenario: mypy가 설치되어 있지 않은 경우
    Given pytest와 ruff는 설치되어 있다
    And mypy는 설치되어 있지 않다
    And 모든 테스트가 통과한다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then pytest가 성공적으로 실행된다
    And ruff가 성공적으로 실행된다
    And mypy 단계에서 "mypy가 설치되어 있지 않습니다" 경고가 표시된다
    And mypy 검증이 스킵된다
    And code-review 에이전트가 호출된다
    And 문서 동기화가 진행된다
    And 최종 보고서에 mypy가 "WARN (not installed)"로 표시된다
```

### 시나리오 5: Python 프로젝트가 아닌 경우

```gherkin
Feature: Phase 0.5 품질 검증 - 비 Python 프로젝트

  Scenario: TypeScript 프로젝트에서 sync 실행
    Given 프로젝트에 pyproject.toml이나 requirements.txt가 없다
    And 프로젝트가 TypeScript 기반이다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then pytest, ruff, mypy 검증이 스킵된다
    And code-review 에이전트만 호출된다
    And 문서 동기화가 진행된다
    And 최종 보고서에 Python 도구들이 "SKIP (not a Python project)"로 표시된다
```

### 시나리오 6: ruff 린팅 이슈 발견

```gherkin
Feature: Phase 0.5 품질 검증 - ruff 이슈 발견

  Scenario: ruff가 린팅 이슈를 발견하는 경우
    Given 모든 도구가 설치되어 있다
    And pytest가 통과한다
    And ruff가 린팅 이슈를 발견한다
    When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
    Then pytest가 성공적으로 실행된다
    And ruff가 실행되고 이슈 목록이 표시된다
    And ruff 결과가 "WARN"으로 표시된다
    And mypy 검증이 계속 진행된다
    And code-review 에이전트가 호출된다
    And 문서 동기화가 진행된다
```

## 품질 게이트 기준

### 필수 통과 조건

| 기준 | 설명 | 통과 조건 |
|------|------|-----------|
| Phase 0.5 실행 | 품질 검증 파이프라인 실행 | 항상 실행됨 |
| 도구 감지 | 설치된 도구 정확히 감지 | 설치/미설치 상태 정확히 판별 |
| 사용자 선택 | pytest 실패 시 선택지 제공 | AskUserQuestion으로 2개 옵션 제공 |
| 결과 보고 | 통합 결과 보고서 출력 | 모든 도구 상태 표시 |

### 권장 통과 조건

| 기준 | 설명 | 권장 수준 |
|------|------|-----------|
| pytest 통과율 | 테스트 통과율 | 100% (0 failures) |
| ruff 이슈 | 린팅 이슈 수 | 0 issues |
| mypy 오류 | 타입 오류 수 | 0 errors |

## 엣지 케이스

### 케이스 1: 모든 도구 미설치

```gherkin
Scenario: 모든 Python 도구가 설치되어 있지 않은 경우
  Given pytest, ruff, mypy 모두 설치되어 있지 않다
  When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
  Then 각 도구에 대해 "미설치" 경고가 표시된다
  And code-review 에이전트만 실행된다
  And 문서 동기화가 진행된다
```

### 케이스 2: code-review 에이전트 미정의

```gherkin
Scenario: code-review 에이전트가 정의되어 있지 않은 경우
  Given code-review.md 에이전트 파일이 존재하지 않는다
  When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
  Then code-review 단계가 스킵된다
  And "code-review 에이전트가 정의되어 있지 않습니다" 경고가 표시된다
  And 문서 동기화가 진행된다
```

### 케이스 3: pytest 타임아웃

```gherkin
Scenario: pytest 실행이 타임아웃되는 경우
  Given pytest가 설치되어 있다
  And 테스트 실행이 5분 이상 소요된다
  When 사용자가 "/moai:3-sync SPEC-001"을 실행한다
  Then pytest가 타임아웃으로 중단된다
  And "pytest 타임아웃 (5분)" 경고가 표시된다
  And 사용자에게 "계속 진행" 또는 "중단" 선택지가 제공된다
```

## Definition of Done

### 기능 완료 조건

- [ ] Phase 0.5 품질 검증이 `/moai:3-sync` 실행 시 자동으로 시작됨
- [ ] pytest, ruff, mypy, code-review 순차 실행됨
- [ ] 도구 미설치 시 경고 메시지 표시 후 다음 단계로 진행됨
- [ ] pytest 실패 시 사용자에게 계속/중단 선택지가 제공됨
- [ ] 품질 검증 결과 보고서가 명확히 표시됨
- [ ] 기존 문서 동기화 로직이 정상 작동함

### 문서 완료 조건

- [ ] 3-sync.md 커맨드 파일이 업데이트됨
- [ ] code-review 에이전트가 정의되어 있음 (필요시)
- [ ] 사용자 가이드에 Phase 0.5 설명이 추가됨

### 테스트 완료 조건

- [ ] 모든 Given-When-Then 시나리오가 수동으로 검증됨
- [ ] 엣지 케이스가 적절히 처리됨
- [ ] 다양한 프로젝트 유형에서 테스트됨

## 추적성 태그

- TAG: SPEC-SYNC-QUALITY-001
- Related Spec: spec.md
- Related Plan: plan.md
