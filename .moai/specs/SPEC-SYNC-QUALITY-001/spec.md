---
id: SPEC-SYNC-QUALITY-001
version: "1.0.0"
status: "completed"
created: "2025-12-22"
updated: "2025-12-22"
author: "GOOS"
priority: "high"
---

# SPEC-SYNC-QUALITY-001: Phase 0.5 품질 검증 통합 - 3-sync 커맨드

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2025-12-22 | GOOS | 초기 SPEC 작성 |

## 개요

`/moai:3-sync` 커맨드에 Phase 0.5 품질 검증 단계를 통합합니다. 문서 동기화 전에 pytest, ruff, mypy, code-review를 자동으로 실행하여 코드 품질을 보장합니다.

**핵심 결정사항**: 스킵 플래그 없이 품질 검증을 항상 자동 모드로 실행합니다.

## 요구사항

### Ubiquitous (항상 적용)

- [REQ-U-001] 시스템은 항상 `/moai:3-sync` 실행 시 Phase 0.5 품질 검증을 먼저 수행해야 한다
- [REQ-U-002] 시스템은 항상 품질 검증 결과를 사용자에게 명확히 보고해야 한다
- [REQ-U-003] 시스템은 항상 각 검증 도구의 실행 상태(성공/실패/미설치)를 추적해야 한다

### Event-Driven (이벤트 기반)

- [REQ-E-001] WHEN `/moai:3-sync SPEC-XXX` 명령이 실행되면 THEN Phase 0.5 품질 검증을 시작한다
- [REQ-E-002] WHEN pytest가 성공하면 THEN ruff 린팅 검사를 실행한다
- [REQ-E-003] WHEN ruff가 성공하면 THEN mypy 타입 검사를 실행한다
- [REQ-E-004] WHEN mypy가 성공하면 THEN code-review 에이전트를 호출한다
- [REQ-E-005] WHEN 모든 품질 검증이 완료되면 THEN 기존 문서 동기화 로직을 실행한다
- [REQ-E-006] WHEN pytest가 실패하면 THEN 사용자에게 계속 진행 여부를 질문한다
- [REQ-E-007] WHEN 검증 도구가 미설치 상태이면 THEN 경고 메시지를 표시하고 다음 단계로 진행한다

### State-Driven (상태 기반)

- [REQ-S-001] IF pytest가 설치되어 있으면 THEN 테스트를 실행한다
- [REQ-S-002] IF ruff가 설치되어 있으면 THEN 린팅 검사를 실행한다
- [REQ-S-003] IF mypy가 설치되어 있으면 THEN 타입 검사를 실행한다
- [REQ-S-004] IF Python 프로젝트가 아니면 THEN Python 관련 도구 검사를 스킵한다
- [REQ-S-005] IF 사용자가 pytest 실패 시 "계속"을 선택하면 THEN 다음 검증 단계로 진행한다
- [REQ-S-006] IF 사용자가 pytest 실패 시 "중단"을 선택하면 THEN sync를 중단하고 오류를 보고한다

### Unwanted (금지 사항)

- [REQ-N-001] 시스템은 품질 검증 스킵 플래그를 제공하지 않아야 한다
- [REQ-N-002] 시스템은 도구 미설치를 오류로 처리하지 않아야 한다 (경고만 표시)
- [REQ-N-003] 시스템은 사용자 확인 없이 pytest 실패 상태에서 sync를 진행하지 않아야 한다

### Optional (선택 사항)

- [REQ-O-001] 가능하면 병렬 실행으로 검증 속도를 최적화한다
- [REQ-O-002] 가능하면 이전 검증 결과를 캐싱하여 중복 실행을 방지한다

## 기술 사양

### Phase 0.5 품질 검증 파이프라인

검증 순서:
1. pytest - 단위 테스트 및 통합 테스트 실행
2. ruff - Python 코드 린팅 (Python 프로젝트에만 해당)
3. mypy - 정적 타입 검사 (Python 프로젝트에만 해당)
4. code-review - AI 기반 코드 리뷰 (code-review 에이전트 활용)

### 도구 설치 감지 로직

```
도구별 설치 확인:
- pytest: `which pytest` 또는 `python -m pytest --version`
- ruff: `which ruff` 또는 `ruff --version`
- mypy: `which mypy` 또는 `mypy --version`
```

### 사용자 상호작용 시나리오

pytest 실패 시 프롬프트:
```
pytest 실행 결과: 3개 테스트 실패

옵션:
1. 계속 진행 - 실패한 테스트를 무시하고 sync 진행
2. 중단 - sync를 중단하고 테스트 수정 후 재시도
```

### 출력 형식

품질 검증 결과 보고서:
```
Phase 0.5 품질 검증 결과
========================

pytest:      PASS (45 tests, 100% coverage)
ruff:        PASS (no issues)
mypy:        WARN (tool not installed)
code-review: PASS (no critical issues)

전체 상태: PASS - 문서 동기화를 진행합니다
```

## 제약사항

### 기술적 제약

- Python 프로젝트가 아닌 경우 pytest, ruff, mypy 검사 스킵
- code-review 에이전트는 항상 실행 (언어 무관)
- 각 도구의 타임아웃: 최대 5분

### 비즈니스 제약

- 스킵 플래그 미제공으로 품질 검증 강제
- 도구 미설치 시 오류가 아닌 경고 처리 (사용자 경험 우선)

## 관련 문서

- `/moai:3-sync` 커맨드: `.claude/commands/moai/3-sync.md`
- code-review 에이전트: `.claude/agents/moai/code-review.md` (신규 생성 또는 확장 필요)
- TRUST 5 Framework: `moai-foundation-core` 스킬

## 추적성 태그

- TAG: SPEC-SYNC-QUALITY-001
- Related: moai-foundation-core (TRUST 5 Framework)
- Related: moai-workflow-spec (SPEC-First TDD)
