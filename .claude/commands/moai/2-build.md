---
name: moai:2-build
description: TDD 구현 (Red-Green-Refactor)
argument-hint: "SPEC-ID - 구현할 SPEC ID (예: SPEC-001) 또는 all로 모든 SPEC 구현"
tools: Read, Write, Edit, MultiEdit, Bash(python3:*), Bash(pytest:*), Task, WebFetch, Grep, Glob, TodoWrite
---

# MoAI-ADK 2단계: TDD 구현 (모드별 Git 통합)

**TDD 구현 대상**: $ARGUMENTS

- ULTRATHINK: code-builder 에이전트가 SPEC을 기반으로 Red-Green-Refactor 사이클과 TRUST 원칙 검증을 지원합니다.

## 🚀 최적화된 에이전트 협업 구조

- **Phase 1**: `code-builder` 에이전트가 전체 TDD 사이클(Red-Green-Refactor)을 일괄 처리합니다.
- **Phase 2**: `git-manager` 에이전트가 TDD 완료 후 모든 커밋을 한 번에 처리합니다.
- **단일 책임 원칙**: code-builder는 전체 TDD 구현, git-manager는 Git 작업 일괄 처리
- **배치 처리**: 단계별 중단 없이 연속적인 TDD 사이클 실행
- **에이전트 간 호출 금지**: 각 에이전트는 독립적으로 실행, 커맨드 레벨에서만 순차 호출

## 🔄 최적화된 워크플로우 실행 순서

### Phase 1: TDD 구현 (code-builder)

`code-builder` 에이전트가 다음을 **연속적으로** 수행:

1. **RED**: 실패하는 테스트 작성 및 확인
2. **GREEN**: 최소 구현으로 테스트 통과 확인
3. **REFACTOR**: 코드 품질 개선 및 TRUST 원칙 검증
4. **품질 검증**: 린터, 테스트 커버리지, 보안 검사 일괄 실행

### Phase 2: Git 작업 (git-manager)

`git-manager` 에이전트가 TDD 완료 후 **한 번에** 수행:

1. **체크포인트 생성**: TDD 시작 전 백업 포인트
2. **구조화된 커밋**: RED→GREEN→REFACTOR 단계별 커밋 생성
3. **최종 동기화**: 모드별 Git 전략 적용 및 원격 동기화


## TDD 단계별 가이드

1. **RED**: Given/When/Then 구조로 실패 테스트 작성. 언어별 테스트 파일 규칙을 따르고, 실패 로그를 간단히 기록합니다.
2. **GREEN**: 테스트를 통과시키는 최소한의 구현만 추가합니다. 최적화는 REFACTOR 단계로 미룹니다.
3. **REFACTOR**: 중복 제거, 명시적 네이밍, 구조화 로깅/예외 처리 보강. 필요 시 추가 커밋으로 분리합니다.

> 헌법 Article I은 기본 권장치만 제공하므로, `simplicity_threshold`를 초과하는 구조가 필요하다면 SPEC 또는 ADR에 근거를 남기고 진행하세요.

## 에이전트 역할 분리

### code-builder 전담 영역

- TDD Red-Green-Refactor 코드 구현
- 테스트 작성 및 실행
- TRUST 5원칙 검증
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
