---
id: SPEC-SYNC-QUALITY-001
version: "1.0.0"
status: "draft"
created: "2025-12-22"
updated: "2025-12-22"
author: "GOOS"
priority: "high"
---

# 구현 계획: Phase 0.5 품질 검증 통합

## 개요

`/moai:3-sync` 커맨드에 Phase 0.5 품질 검증 단계를 통합하는 구현 계획입니다.

## 마일스톤

### Primary Goal: 핵심 품질 검증 파이프라인

**목표**: pytest, ruff, mypy, code-review 순차 실행 구조 구현

태스크:
1. 3-sync.md 커맨드 파일 수정
   - Phase 0.5 품질 검증 단계 추가
   - 도구 설치 감지 로직 구현
   - 순차 실행 파이프라인 구현

2. 도구 실행 및 결과 처리
   - pytest 실행 및 결과 파싱
   - ruff 린팅 실행 및 결과 파싱
   - mypy 타입 검사 실행 및 결과 파싱

3. 사용자 상호작용 구현
   - pytest 실패 시 AskUserQuestion으로 계속/중단 선택
   - 도구 미설치 시 경고 메시지 표시

### Secondary Goal: code-review 에이전트 통합

**목표**: AI 기반 코드 리뷰 기능 통합

태스크:
1. code-review 에이전트 정의 확인/생성
   - `.claude/agents/moai/code-review.md` 파일 확인
   - 필요시 새 에이전트 정의 생성

2. 3-sync에서 code-review 에이전트 호출
   - Task() 패턴으로 에이전트 호출
   - 리뷰 결과 수집 및 보고

### Final Goal: 결과 보고 및 문서화

**목표**: 품질 검증 결과를 명확히 보고하고 문서화

태스크:
1. 통합 결과 보고서 생성
   - 각 도구별 상태 (PASS/FAIL/WARN/SKIP)
   - 전체 상태 요약
   - 다음 단계 안내

2. 기존 sync 로직과 통합
   - Phase 0.5 완료 후 기존 문서 동기화 실행
   - 실패 시 적절한 오류 메시지 제공

## 수정 대상 파일

### 필수 수정

| 파일 경로 | 변경 유형 | 설명 |
|-----------|-----------|------|
| `.claude/commands/moai/3-sync.md` | 수정 | Phase 0.5 품질 검증 단계 추가 |

### 신규 생성 (필요시)

| 파일 경로 | 변경 유형 | 설명 |
|-----------|-----------|------|
| `.claude/agents/moai/code-review.md` | 생성 | code-review 에이전트 정의 |

### 참조 파일

| 파일 경로 | 용도 |
|-----------|------|
| `.claude/skills/moai-foundation-core/SKILL.md` | TRUST 5 프레임워크 참조 |
| `.claude/agents/moai/` | 기존 에이전트 패턴 참조 |

## 기술적 접근 방식

### 도구 설치 감지

```bash
# pytest 감지
which pytest || python -m pytest --version

# ruff 감지
which ruff || ruff --version

# mypy 감지
which mypy || mypy --version
```

### 실행 흐름

1. SPEC ID 파라미터 파싱
2. Phase 0.5 품질 검증 시작
   - 도구 설치 확인
   - pytest 실행 (설치된 경우)
   - pytest 실패 시 사용자 선택 요청
   - ruff 실행 (Python 프로젝트, 설치된 경우)
   - mypy 실행 (Python 프로젝트, 설치된 경우)
   - code-review 에이전트 호출
3. 결과 보고서 출력
4. 기존 문서 동기화 로직 실행

### 사용자 선택 구현

pytest 실패 시 AskUserQuestion 사용:
- 옵션 1: "계속 진행" - 테스트 실패를 무시하고 sync 계속
- 옵션 2: "중단" - sync 중단, 테스트 수정 권장

## 리스크 및 대응

### 리스크 1: 도구 미설치로 인한 사용자 혼란

- 대응: 명확한 경고 메시지와 설치 안내 제공
- 예시: "mypy가 설치되어 있지 않습니다. 타입 검사를 건너뜁니다. 설치: pip install mypy"

### 리스크 2: 품질 검증 시간 증가

- 대응: 각 도구에 타임아웃 설정 (최대 5분)
- 대응: 진행 상황 실시간 표시

### 리스크 3: code-review 에이전트 부재

- 대응: 에이전트 미존재 시 생성 또는 스킵 처리
- 대응: 기존 유사 에이전트 패턴 참조

## 테스트 전략

### 단위 테스트

- 도구 설치 감지 로직 테스트
- 결과 파싱 로직 테스트
- 사용자 선택 분기 테스트

### 통합 테스트

- 전체 Phase 0.5 파이프라인 실행
- 도구 미설치 시나리오
- pytest 실패 시나리오

### 수동 테스트

- 실제 프로젝트에서 `/moai:3-sync` 실행
- 다양한 도구 설치 조합 테스트

## 추적성 태그

- TAG: SPEC-SYNC-QUALITY-001
- Related Spec: spec.md
- Related Acceptance: acceptance.md
