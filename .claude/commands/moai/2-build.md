---
name: moai:2-build
description: TDD 구현 (Red-Green-Refactor)
argument-hint: [SPEC-ID] - 구현할 SPEC ID (예: SPEC-001) 또는 'all'로 모든 SPEC 구현
allowed-tools: Read, Write, Edit, MultiEdit, Bash(python3:*), Bash(pytest:*), Task, WebFetch, Grep, Glob, TodoWrite
---

# MoAI-ADK 2단계: TDD 구현 (모드별 Git 통합)

**TDD 구현 대상**: $ARGUMENTS

code-builder 에이전트가 SPEC을 기반으로 Red-Green-Refactor 사이클과 Constitution 검증을 지원합니다.

## 에이전트 협업 구조

- **1단계**: `code-builder` 에이전트가 TDD 구현 (Red-Green-Refactor)을 전담합니다.
- **2단계**: `git-manager` 에이전트가 단계별 커밋과 체크포인트를 전담합니다.
- **단일 책임 원칙**: code-builder는 코드 구현만, git-manager는 Git 작업만 수행합니다.
- **순차 실행**: 각 TDD 단계마다 code-builder → git-manager 순서로 실행합니다.
- **에이전트 간 호출 금지**: 각 에이전트는 다른 에이전트를 직접 호출하지 않고, 커맨드 레벨에서만 순차 실행합니다.

## 브레인스토밍/디버깅 지원 (선택)

- `.moai/config.json.brainstorming.enabled` 가 `true` 인 경우, 구현 전후로 다음 절차를 추가합니다.
  - 설계 검토 단계에서 `codex-bridge` 와 `gemini-bridge` 를 호출해 대안 아키텍처나 디버깅 아이디어를 수집합니다. (예: `Task: use codex-bridge to run "codex exec --model gpt-5-codex ..."`)
  - 외부 제안은 Claude Code에서 종합·검증한 뒤 실제 구현 결정에 반영합니다.
- 설정이 비활성화되어 있으면 Claude Code만 사용합니다.

## 워크플로우 실행 순서

당신은 다음 순서로 에이전트들을 **TDD 단계별로 순차 호출**해야 합니다:

### RED 단계

1. `code-builder` 에이전트: 실패하는 테스트 작성
2. `git-manager` 에이전트: RED 커밋 및 체크포인트 생성

### GREEN 단계

1. `code-builder` 에이전트: 최소 구현으로 테스트 통과
2. `git-manager` 에이전트: GREEN 커밋 및 체크포인트 생성

### REFACTOR 단계

1. `code-builder` 에이전트: 코드 품질 개선 및 리팩터링
2. `git-manager` 에이전트: REFACTOR 커밋 및 최종 동기화

**중요**: 각 에이전트는 독립적으로 실행되며, 에이전트 간 직접 호출은 금지됩니다.

## 실행 흐름 요약

| 단계     | code-builder 역할      | git-manager 역할           |
| -------- | ---------------------- | -------------------------- |
| RED      | 실패 테스트 작성       | 체크포인트 + RED 커밋      |
| GREEN    | 최소 구현, 테스트 통과 | 체크포인트 + GREEN 커밋    |
| REFACTOR | 품질 개선, 린터 실행   | 체크포인트 + REFACTOR 커밋 |
| 마무리   | Constitution 검증      | 최종 동기화 및 커밋        |

## TDD 단계별 가이드

1. **RED**: Given/When/Then 구조로 실패 테스트 작성. 언어별 테스트 파일 규칙을 따르고, 실패 로그를 간단히 기록합니다.
2. **GREEN**: 테스트를 통과시키는 최소한의 구현만 추가합니다. 최적화는 REFACTOR 단계로 미룹니다.
3. **REFACTOR**: 중복 제거, 명시적 네이밍, 구조화 로깅/예외 처리 보강. 필요 시 추가 커밋으로 분리합니다.

> 헌법 Article I은 기본 권장치만 제공하므로, `simplicity_threshold`를 초과하는 구조가 필요하다면 SPEC 또는 ADR에 근거를 남기고 진행하세요.

## 에이전트 역할 분리

### code-builder 전담 영역

- TDD Red-Green-Refactor 코드 구현
- 테스트 작성 및 실행
- Constitution 5원칙 검증
- 코드 품질 체크
- 언어별 린터/포매터 실행

### git-manager 전담 영역

- 모든 Git 커밋 작업 (add, commit, push)
- TDD 단계별 체크포인트 생성
- 모드별 커밋 전략 적용
- 깃 브랜치/태그 관리
- 원격 동기화 처리

## 품질 게이트 체크리스트

- 테스트 커버리지 ≥ `.moai/config.json.test_coverage_target` (기본 85%)
- 린터/포매터 통과 (`ruff`, `eslint --fix`, `gofmt` 등)
- 구조화 로깅 또는 관측 도구 호출 존재 확인
- 16-Core @TAG 업데이트 필요 변경 사항 메모 (다음 단계에서 doc-syncer가 사용)

## 다음 단계

- TDD 구현 완료 후 `/moai:3-sync`로 문서 동기화 진행
- 모든 Git 작업은 git-manager 에이전트가 전담하여 일관성 보장
- 에이전트 간 직접 호출 없이 커맨드 레벨 오케스트레이션만 사용
