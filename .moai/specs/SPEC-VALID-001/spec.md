---
id: VALID-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @Goos
priority: high
category: feature
labels:
  - validation
  - spec
  - metadata
  - automation
depends_on: []
related_specs:
  - INIT-003
scope:
  packages:
    - moai-adk-ts/src/core/spec
  files:
    - spec-validator.ts
    - metadata-parser.ts
---

# @SPEC:VALID-001: SPEC 메타데이터 검증 자동화

## HISTORY

### v0.0.1 (2025-10-16)
- **INITIAL**: SPEC 메타데이터 검증 자동화 명세 작성
- **AUTHOR**: @Goos
- **REASON**: SPEC 문서 품질 보장 및 일관성 유지를 위한 자동 검증 시스템 필요

---

## 개요

MoAI-ADK의 모든 SPEC 문서가 `.moai/memory/spec-metadata.md`에 정의된 메타데이터 표준을 준수하도록 자동 검증 시스템을 구축합니다. `/alfred:1-spec` 실행 시 7개 필수 필드와 HISTORY 섹션을 검증하고, 위반 시 구체적인 오류 메시지를 제공합니다.

---

## EARS 요구사항 명세

### Ubiquitous Requirements (기본 요구사항)

1. **메타데이터 필수 필드 검증**
   - 시스템은 SPEC 문서의 7개 필수 필드를 검증해야 한다
   - 필수 필드: id, version, status, created, updated, author, priority

2. **HISTORY 섹션 검증**
   - 시스템은 SPEC 문서에 HISTORY 섹션이 존재하는지 검증해야 한다
   - HISTORY 섹션은 최소 1개 이상의 버전 항목을 포함해야 한다

3. **디렉토리 명명 규칙 검증**
   - 시스템은 SPEC 디렉토리가 `SPEC-{ID}/` 형식을 준수하는지 검증해야 한다
   - 예: `SPEC-VALID-001/`, `SPEC-AUTH-001/`

### Event-driven Requirements (이벤트 기반 요구사항)

4. **SPEC 생성 시 자동 검증**
   - WHEN `/alfred:1-spec` 명령이 실행되면, 시스템은 SPEC 문서 생성 직후 메타데이터 검증을 수행해야 한다

5. **중복 ID 감지**
   - WHEN SPEC ID가 생성되면, 시스템은 기존 SPEC 디렉토리에서 중복 ID를 검색해야 한다
   - WHEN 중복 ID가 발견되면, 시스템은 오류 메시지와 함께 SPEC 생성을 중단해야 한다

6. **필수 필드 누락 감지**
   - WHEN SPEC 문서가 작성되면, 시스템은 7개 필수 필드 각각의 존재 여부를 확인해야 한다
   - WHEN 필수 필드가 누락되면, 시스템은 누락된 필드명과 예제를 포함한 오류 메시지를 반환해야 한다

### State-driven Requirements (상태 기반 요구사항)

7. **실시간 피드백 제공**
   - WHILE SPEC 작성 중일 때, 시스템은 검증 결과를 즉시 피드백으로 제공해야 한다

8. **검증 통과 시 확인 메시지**
   - WHILE 모든 검증이 통과했을 때, 시스템은 성공 메시지와 함께 다음 단계를 안내해야 한다

### Constraints (제약사항)

9. **필드 형식 검증**
   - IF `version` 필드가 존재하면, Semantic Versioning 형식(MAJOR.MINOR.PATCH)을 준수해야 한다
   - IF `author` 필드가 존재하면, `@{GitHub ID}` 형식을 준수해야 한다
   - IF `created/updated` 필드가 존재하면, `YYYY-MM-DD` 형식을 준수해야 한다

10. **Enum 값 검증**
    - IF `status` 필드가 존재하면, `draft|active|completed|deprecated` 중 하나여야 한다
    - IF `priority` 필드가 존재하면, `critical|high|medium|low` 중 하나여야 한다

11. **순환 의존성 감지**
    - IF `depends_on` 필드가 존재하면, 시스템은 순환 의존성을 감지하고 경고해야 한다

12. **구체적 오류 메시지**
    - IF 검증이 실패하면, 시스템은 실패한 필드명, 현재 값, 올바른 형식, 수정 예제를 포함한 오류 메시지를 제공해야 한다

### Optional Features (선택적 기능)

13. **선택 필드 검증**
    - WHERE 선택 필드(category, labels, scope 등)가 존재하면, 시스템은 해당 필드의 형식을 검증할 수 있다

---

## 수락 기준 (Acceptance Criteria)

### 필수 검증 항목 (12개)

1. ✅ **AC-001**: 7개 필수 필드 존재 확인
   - id, version, status, created, updated, author, priority

2. ✅ **AC-002**: HISTORY 섹션 존재 및 최소 1개 버전 항목 확인

3. ✅ **AC-003**: 디렉토리 명명 규칙 검증 (`SPEC-{ID}/`)

4. ✅ **AC-004**: version 필드 형식 검증 (Semantic Versioning)

5. ✅ **AC-005**: author 필드 형식 검증 (`@{GitHub ID}`)

6. ✅ **AC-006**: created/updated 날짜 형식 검증 (`YYYY-MM-DD`)

7. ✅ **AC-007**: status enum 값 검증 (draft|active|completed|deprecated)

8. ✅ **AC-008**: priority enum 값 검증 (critical|high|medium|low)

9. ✅ **AC-009**: depends_on 순환 의존성 감지

10. ✅ **AC-010**: 중복 SPEC ID 감지 (Grep 도구 활용)

11. ✅ **AC-011**: 필드별 구체적 오류 메시지 제공
    - 형식: `❌ {필드명} 검증 실패: {현재값} → 올바른 형식: {형식} (예: {예제})`

12. ✅ **AC-012**: 검증 통과 시 성공 메시지 반환
    - 형식: `✅ SPEC-{ID} 메타데이터 검증 통과`

---

## 기술 제약사항

### 검증 도구

- **Grep 도구**: 중복 ID 검색 및 TAG 체인 검증
- **정규 표현식**: YAML Front Matter 파싱 및 필드 형식 검증
- **파일 시스템**: 디렉토리 명명 규칙 검증

### 성능 요구사항

- 검증 수행 시간: ≤ 500ms (단일 SPEC 기준)
- 중복 ID 검색 시간: ≤ 1초 (.moai/specs/ 전체 스캔)

### 에러 핸들링

- 검증 실패 시 **작업 중단** 및 오류 메시지 출력
- 검증 통과 시 **다음 단계 자동 진행** (Git 커밋, Draft PR 생성)

---

## 테스트 전략

### 단위 테스트 (Unit Tests)

- 각 검증 함수별 독립 테스트
- 정상 케이스 6개, 오류 케이스 6개

### 통합 테스트 (Integration Tests)

- `/alfred:1-spec` 실행 시 전체 워크플로우 테스트
- 실제 SPEC 문서 생성 → 검증 → Git 커밋 → PR 생성

---

## 보안 고려사항

- YAML 파싱 시 **임의 코드 실행 방지** (safe parsing only)
- 파일 시스템 접근 시 **경로 검증** (.moai/specs/ 내부로 제한)

---

## 참조 문서

- **SPEC 메타데이터 표준**: `.moai/memory/spec-metadata.md`
- **개발 가이드**: `.moai/memory/development-guide.md`
- **CLAUDE.md**: SPEC-First TDD 워크플로우

---

## 다음 단계

1. `/alfred:2-build VALID-001` - TDD 구현 (RED-GREEN-REFACTOR)
2. `/alfred:3-sync` - 문서 동기화 및 TAG 검증

---

**문서 상태**: Draft (v0.0.1)
**마지막 업데이트**: 2025-10-16
