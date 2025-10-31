# Acceptance Criteria: SPEC-SESSION-CLEANUP-001

<!-- @ACCEPTANCE:SESSION-CLEANUP-001 -->

---

## Overview

이 문서는 SPEC-SESSION-CLEANUP-001의 수락 기준을 정의합니다. 모든 시나리오는 Given-When-Then 형식으로 작성되었으며, 각 Alfred 커맨드의 완료 패턴을 검증합니다.

---

## Acceptance Scenarios

### Scenario 1: `/alfred:0-project` 완료 후 AskUserQuestion 호출

**Feature**: Alfred는 프로젝트 초기화 완료 후 다음 단계를 묻는다

```gherkin
Given Alfred가 /alfred:0-project 커맨드를 실행 중이다
And 프로젝트 초기화가 성공적으로 완료되었다
And TodoWrite의 모든 작업이 "completed" 상태이다

When 커맨드가 종료되려고 할 때

Then AskUserQuestion tool이 정확히 1번 호출되어야 한다
And AskUserQuestion은 다음 3개 옵션을 포함해야 한다:
    | Label                  | Description                        |
    | ---------------------- | ---------------------------------- |
    | 📋 스펙 작성 진행      | /alfred:1-plan 실행                |
    | 🔍 프로젝트 구조 검토  | 현재 상태 확인                     |
    | 🔄 새 세션 시작        | /clear 실행                        |
And 질문 헤더는 "다음 단계"이어야 한다
And 질문 본문은 사용자의 conversation_language로 작성되어야 한다
And Prose 제안 (e.g., "You can now run...")이 출력되지 않아야 한다
```

**Test Case**: `@TEST:SESSION-001`

**Validation**:
- ✅ AskUserQuestion 호출 확인
- ✅ 옵션 개수 = 3
- ✅ 옵션 label 및 description 검증
- ✅ Prose 패턴 부재 확인

---

### Scenario 2: `/alfred:1-plan` 완료 후 AskUserQuestion 호출

**Feature**: Alfred는 SPEC 작성 완료 후 다음 단계를 묻는다

```gherkin
Given Alfred가 /alfred:1-plan 커맨드를 실행 중이다
And SPEC 문서(spec.md, plan.md, acceptance.md)가 생성되었다
And TodoWrite의 모든 작업이 "completed" 상태이다

When 커맨드가 종료되려고 할 때

Then AskUserQuestion tool이 정확히 1번 호출되어야 한다
And AskUserQuestion은 다음 3개 옵션을 포함해야 한다:
    | Label                  | Description                        |
    | ---------------------- | ---------------------------------- |
    | 🚀 구현 진행           | /alfred:2-run SPEC-XXX-001 실행    |
    | ✏️ SPEC 수정           | 현재 SPEC 재작업                   |
    | 🔄 새 세션 시작        | /clear 실행                        |
And SPEC ID가 옵션 1의 description에 포함되어야 한다
And Prose 제안이 출력되지 않아야 한다
```

**Test Case**: `@TEST:SESSION-002`

**Validation**:
- ✅ AskUserQuestion 호출 확인
- ✅ 옵션 개수 = 3
- ✅ SPEC ID 동적 삽입 확인
- ✅ Prose 패턴 부재 확인

---

### Scenario 3: `/alfred:2-run` 완료 후 AskUserQuestion 호출

**Feature**: Alfred는 TDD 구현 완료 후 다음 단계를 묻는다

```gherkin
Given Alfred가 /alfred:2-run SPEC-XXX-001 커맨드를 실행 중이다
And TDD 사이클 (RED → GREEN → REFACTOR)이 완료되었다
And 모든 테스트가 통과했다
And TodoWrite의 모든 작업이 "completed" 상태이다

When 커맨드가 종료되려고 할 때

Then AskUserQuestion tool이 정확히 1번 호출되어야 한다
And AskUserQuestion은 다음 3개 옵션을 포함해야 한다:
    | Label                  | Description                        |
    | ---------------------- | ---------------------------------- |
    | 📚 문서 동기화         | /alfred:3-sync 실행                |
    | 🧪 추가 테스트/검증    | 테스트 재실행                      |
    | 🔄 새 세션 시작        | /clear 실행                        |
And Prose 제안이 출력되지 않아야 한다
```

**Test Case**: `@TEST:SESSION-003`

**Validation**:
- ✅ AskUserQuestion 호출 확인
- ✅ 옵션 개수 = 3
- ✅ TDD 완료 상태 확인
- ✅ Prose 패턴 부재 확인

---

### Scenario 4: `/alfred:3-sync` 완료 후 AskUserQuestion 호출

**Feature**: Alfred는 문서 동기화 완료 후 다음 단계를 묻는다

```gherkin
Given Alfred가 /alfred:3-sync 커맨드를 실행 중이다
And 문서 동기화(README.md, CHANGELOG.md 등)가 완료되었다
And @TAG 체인이 검증되었다
And TodoWrite의 모든 작업이 "completed" 상태이다

When 커맨드가 종료되려고 할 때

Then AskUserQuestion tool이 정확히 1번 호출되어야 한다
And AskUserQuestion은 다음 3개 옵션을 포함해야 한다:
    | Label                  | Description                        |
    | ---------------------- | ---------------------------------- |
    | 📋 다음 기능 계획      | /alfred:1-plan 실행                |
    | 🔀 PR 병합             | main 브랜치로 병합                 |
    | ✅ 세션 완료           | 작업 종료                          |
And Prose 제안이 출력되지 않아야 한다
```

**Test Case**: `@TEST:SESSION-004`

**Validation**:
- ✅ AskUserQuestion 호출 확인
- ✅ 옵션 개수 = 3
- ✅ 문서 동기화 완료 상태 확인
- ✅ Prose 패턴 부재 확인

---

### Scenario 5: 세션 요약 생성 (사용자가 "새 세션" 선택)

**Feature**: 사용자가 "새 세션" 선택 시 세션 요약이 생성된다

```gherkin
Given Alfred가 /alfred:1-plan 커맨드를 완료했다
And TodoWrite에 3개의 "completed" 작업이 있다
And AskUserQuestion이 다음 단계를 물어봤다

When 사용자가 "🔄 새 세션 시작" 옵션을 선택했을 때

Then 세션 요약이 Markdown 형식으로 생성되어야 한다
And 세션 요약은 다음 섹션을 포함해야 한다:
    | Section              | Content                           |
    | -------------------- | --------------------------------- |
    | 완료된 작업          | TodoWrite의 모든 completed 작업   |
    | Git 통계             | 커밋 수, 변경된 파일, 라인 수     |
    | 다음 권장 작업       | 사용자에게 추천할 다음 단계       |
And 세션 요약은 사용자의 conversation_language로 작성되어야 한다
```

**Test Case**: `@TEST:SESSION-005`

**Validation**:
- ✅ 세션 요약 생성 확인
- ✅ 필수 섹션 3개 포함 확인
- ✅ Markdown 형식 검증
- ✅ Language consistency 확인

---

### Scenario 6: TodoWrite 정리 (커맨드 완료 시)

**Feature**: 커맨드 완료 시 TodoWrite 상태가 정리된다

```gherkin
Given Alfred가 /alfred:2-run 커맨드를 실행 중이다
And TodoWrite에 5개의 작업이 있다:
    | Task                  | Status      |
    | --------------------- | ----------- |
    | 테스트 작성           | completed   |
    | 코드 구현             | completed   |
    | 리팩토링              | completed   |
    | 문서 업데이트         | completed   |
    | 커밋 생성             | completed   |

When 커맨드가 AskUserQuestion을 호출하기 직전

Then 모든 "completed" 작업이 추출되어야 한다
And 추출된 작업 목록이 세션 컨텍스트에 저장되어야 한다
And TodoWrite 상태가 초기화되어야 한다 (optional)
```

**Test Case**: `@TEST:SESSION-006`

**Validation**:
- ✅ TodoWrite에서 completed 작업 추출
- ✅ 세션 컨텍스트에 저장 확인
- ✅ 모든 작업이 captured되었는지 검증

---

### Scenario 7: Prose 제안 금지 검증

**Feature**: 시스템은 어떤 상황에서도 prose로 다음 단계를 제안하지 않는다

```gherkin
Given Alfred가 /alfred:0-project 커맨드를 완료했다

When 커맨드 출력을 검사할 때

Then 다음 패턴이 출력에 포함되지 않아야 한다:
    | Prohibited Pattern                          |
    | ------------------------------------------- |
    | "You can now run..."                        |
    | "Next, you should..."                       |
    | "To continue, please run..."                |
    | "다음으로 /alfred:1-plan을 실행하세요"      |
And AskUserQuestion tool이 반드시 호출되어야 한다
```

**Test Case**: `@TEST:SESSION-007`

**Validation**:
- ✅ Prose 패턴 검색 (regex)
- ✅ AskUserQuestion 호출 필수 확인
- ✅ 모든 커맨드에 대해 검증 (4개)

---

### Scenario 8: Batched AskUserQuestion 디자인 검증

**Feature**: AskUserQuestion은 1개의 호출로 모든 질문을 처리한다

```gherkin
Given Alfred가 /alfred:1-plan 커맨드를 완료했다

When AskUserQuestion tool이 호출될 때

Then 호출 횟수는 정확히 1번이어야 한다
And 질문 개수는 1개여야 한다
And 옵션 개수는 3-4개 범위여야 한다
And 각 옵션은 label과 description을 포함해야 한다
```

**Test Case**: `@TEST:SESSION-008`

**Validation**:
- ✅ AskUserQuestion 호출 횟수 = 1
- ✅ 질문 구조 검증 (questions array)
- ✅ 옵션 개수 범위 검증 (3-4)

---

## Quality Metrics

### Functional Metrics

| Metric                                      | Target | Actual | Status |
| ------------------------------------------- | ------ | ------ | ------ |
| AskUserQuestion 호출 성공률                 | 100%   | TBD    | ⏳     |
| Prose 제안 검출 결과                        | 0건    | TBD    | ⏳     |
| 세션 요약 생성 성공률                       | 100%   | TBD    | ⏳     |
| TodoWrite 정리 성공률                       | 100%   | TBD    | ⏳     |
| 옵션 개수 일관성 (3-4개)                    | 100%   | TBD    | ⏳     |

### Non-Functional Metrics

| Metric                                      | Target   | Actual | Status |
| ------------------------------------------- | -------- | ------ | ------ |
| AskUserQuestion 응답 시간                   | <500ms   | TBD    | ⏳     |
| 세션 요약 생성 시간                         | <1000ms  | TBD    | ⏳     |
| TodoWrite 상태 변경 지연                    | <100ms   | TBD    | ⏳     |

### User Experience Metrics

| Metric                                      | Target | Actual | Status |
| ------------------------------------------- | ------ | ------ | ------ |
| 옵션 이해도 (사용자 피드백)                 | >90%   | TBD    | ⏳     |
| 다음 단계 선택 정확도                       | >95%   | TBD    | ⏳     |
| 세션 요약 유용성 (사용자 만족도)            | >85%   | TBD    | ⏳     |

---

## Test Execution Plan

### Phase 1: Unit Testing

**Target**: 개별 컴포넌트 검증

1. **AskUserQuestion 템플릿 검증**
   - 각 커맨드 파일에서 AskUserQuestion 코드 블록 추출
   - Python 문법 검증 (syntax check)
   - 옵션 개수 및 구조 검증

2. **Prose 패턴 검색**
   - Regex 패턴으로 4개 커맨드 파일 검색
   - 금지 문구 검출 확인
   - 결과: 0건 예상

3. **TodoWrite 정리 로직 검증**
   - Mock TodoWrite 상태 생성
   - Completed 작업 추출 테스트
   - 세션 컨텍스트 저장 확인

---

### Phase 2: Integration Testing

**Target**: 커맨드 실행 흐름 검증

1. **`/alfred:0-project` End-to-End**
   - 프로젝트 초기화 실행
   - AskUserQuestion 호출 확인
   - 사용자 옵션 선택 시뮬레이션
   - 다음 커맨드 실행 확인

2. **`/alfred:1-plan` End-to-End**
   - SPEC 생성 실행
   - AskUserQuestion 호출 확인
   - SPEC ID 동적 삽입 검증

3. **`/alfred:2-run` End-to-End**
   - TDD 구현 실행
   - AskUserQuestion 호출 확인
   - 테스트 통과 확인

4. **`/alfred:3-sync` End-to-End**
   - 문서 동기화 실행
   - AskUserQuestion 호출 확인
   - @TAG 체인 검증

---

### Phase 3: User Acceptance Testing

**Target**: 실제 사용자 시나리오 검증

1. **새 프로젝트 생성 시나리오**
   - `/alfred:0-project` → `/alfred:1-plan` → `/alfred:2-run` → `/alfred:3-sync` 전체 흐름
   - 각 단계에서 AskUserQuestion 정상 동작 확인
   - 세션 요약 생성 확인

2. **중간 단계 재시작 시나리오**
   - `/alfred:1-plan` 후 "SPEC 수정" 선택
   - SPEC 재작업 후 다시 `/alfred:2-run` 실행
   - TodoWrite 상태 일관성 확인

3. **세션 종료 시나리오**
   - `/alfred:3-sync` 후 "세션 완료" 선택
   - 세션 요약 출력 확인
   - 모든 작업 기록 검증

---

## Definition of Done

**이 SPEC은 다음 조건을 모두 만족할 때 완료된 것으로 간주됩니다:**

### Code Implementation
- [x] 4개 커맨드 파일에 AskUserQuestion 패턴 추가 완료
- [ ] TodoWrite 정리 로직 구현 완료
- [ ] 세션 요약 생성 함수 구현 완료

### Testing
- [ ] 8개 시나리오 모두 테스트 통과
- [ ] Prose 패턴 검색 결과 0건
- [ ] Unit test coverage ≥ 90%
- [ ] Integration test 통과율 100%

### Documentation
- [x] CLAUDE.md 문서 업데이트 완료
- [ ] CHANGELOG.md에 변경사항 기록
- [ ] `.moai/specs/SPEC-SESSION-CLEANUP-001/` 문서 3개 생성

### Quality Gates
- [ ] Linting 통과 (0 issues)
- [ ] Type checking 통과 (mypy, pyright)
- [ ] @TAG 체인 검증 완료
- [ ] TRUST 5 원칙 준수 확인

### User Acceptance
- [ ] 사용자 피드백 수집 (≥3명)
- [ ] 옵션 이해도 ≥90%
- [ ] 다음 단계 선택 정확도 ≥95%

---

## Edge Cases & Error Handling

### Edge Case 1: AskUserQuestion 호출 실패

**Scenario**:
```gherkin
Given Alfred가 /alfred:0-project 완료 후 AskUserQuestion을 호출했다
When AskUserQuestion tool이 실패했을 때 (timeout, error)
Then 시스템은 fallback 메시지를 출력해야 한다
And 에러 로그에 실패 원인을 기록해야 한다
And 사용자에게 수동으로 다음 커맨드를 실행하도록 안내해야 한다
```

**Fallback Message**:
```markdown
⚠️ 다음 단계 선택을 불러올 수 없습니다.

수동으로 다음 커맨드를 실행하세요:
- `/alfred:1-plan` - 스펙 작성 시작
- `/clear` - 새 세션 시작
```

---

### Edge Case 2: TodoWrite 상태 불일치

**Scenario**:
```gherkin
Given Alfred가 /alfred:2-run 실행 중이다
And TodoWrite에 1개의 "in_progress" 작업이 남아있다
When 커맨드가 종료되려고 할 때
Then 시스템은 경고 메시지를 출력해야 한다
And AskUserQuestion 호출을 차단해야 한다
And 사용자에게 미완료 작업을 처리하도록 요청해야 한다
```

**Warning Message**:
```markdown
⚠️ 미완료 작업이 있습니다:
- [ ] in_progress: 테스트 작성 중

먼저 모든 작업을 완료하거나 취소하세요.
```

---

### Edge Case 3: 사용자가 잘못된 옵션 선택

**Scenario**:
```gherkin
Given AskUserQuestion이 3개 옵션을 제시했다
When 사용자가 유효하지 않은 입력을 제공했을 때 (예: 숫자 4)
Then 시스템은 에러 메시지를 출력해야 한다
And AskUserQuestion을 다시 호출해야 한다
```

**Error Message**:
```markdown
❌ 잘못된 선택입니다. 1-3 중에서 선택해주세요.
```

---

## Traceability Matrix

| Requirement         | Test Case         | Status |
| ------------------- | ----------------- | ------ |
| REQ-SESSION-001     | TEST:SESSION-001  | ⏳     |
| REQ-SESSION-002     | TEST:SESSION-006  | ⏳     |
| REQ-SESSION-003     | TEST:SESSION-001  | ⏳     |
| REQ-SESSION-004     | TEST:SESSION-002  | ⏳     |
| REQ-SESSION-005     | TEST:SESSION-003  | ⏳     |
| REQ-SESSION-006     | TEST:SESSION-004  | ⏳     |
| REQ-SESSION-007     | TEST:SESSION-005  | ⏳     |
| REQ-SESSION-008     | TEST:SESSION-006  | ⏳     |
| REQ-SESSION-009     | TEST:SESSION-005  | ⏳     |
| REQ-SESSION-010     | TEST:SESSION-007  | ⏳     |
| REQ-SESSION-011     | TEST:SESSION-008  | ⏳     |

---

## Sign-off

**Acceptance Criteria Approved By**:

- [ ] **Product Owner**: @GoosLab
- [ ] **Tech Lead**: @GoosLab
- [ ] **QA Lead**: TBD

**Date**: 2025-10-30

---
