---
id: SPEC-WEB-003
version: "1.0.0"
status: "draft"
created: "2026-01-10"
updated: "2026-01-10"
author: "GOOS행"
tags: ["web", "ui", "settings", "configuration", "dialog"]
---

# Acceptance Criteria: SPEC-WEB-003

## 1. Test Scenarios (Given-When-Then Format)

### Scenario 1: 설정 모달 열기 (SC-001)

```
Given: 사용자가 웹 UI 메인 페이지에 있고
  And: 프로젝트 설정 파일이 존재或不존재

When: 사용자가 우측 상단 Settings 버튼 클릭

Then:
  - 설정 모달(Dialog)이 표시되고
  - 기존 설정값이 폼에 로드되며 (파일 존재 시)
  - 기본값이 폼에 표시된다 (파일 미존재 시)
  - 3개의 Tab(Quick Start, Documentation, Git Automation)이 표시된다
```

**Requirements Coverage**: E-001, S-004, S-005

---

### Scenario 2: 설정 저장 성공 (SC-002)

```
Given: 사용자가 설정 모달이 열려있고
  And: 모든 필수 필드가 유효한 값으로 채워져 있고
  And: 사용자가 일부 필드를 수정했을 때

When: 사용자가 저장 버튼 클릭

Then:
  - 변경사항이 .moai/config/config.yaml에 저장되고
  - 백업 파일이 생성되며
  - 성공 메시지(Toast)가 표시되고
  - 모달이 닫힌다
```

**Requirements Coverage**: E-003, E-005, U-002, U-003

---

### Scenario 3: 조건부 Batch 표시 (SC-003)

```
Given: 사용자가 설정 모달이 열려있고
  And: Git Automation Tab을 선택하고

When: git_strategy.mode를 'personal'로 변경

Then:
  - Tab 3에 Personal Git 관련 질문만 표시되고
  - Team 관련 질문은 숨겨진다

When: git_strategy.mode를 'team'으로 변경

Then:
  - Tab 3에 Team Git 관련 질문만 표시되고
  - Personal 관련 질문은 숨겨진다
```

**Requirements Coverage**: S-001, S-002, S-003

---

### Scenario 4: 유효성 검증 실패 (SC-004)

```
Given: 사용자가 설정 모달이 열려있고
  And: 필수 필드가 비어있거나 유효하지 않은 값일 때

When: 사용자가 저장 버튼 클릭

Then:
  - 저장이 차단되고
  - 에러 메시지가 필드 아래에 표시되며
  - 모달은 열려있는 상태를 유지한다
```

**Requirements Coverage**: U-002, X-001, S-006

---

### Scenario 5: 변경사항 없이 취소 (SC-005)

```
Given: 사용자가 설정 모달이 열려있고
  And: 아무 필드도 수정하지 않았을 때

When: 사용자가 취소 버튼 클릭 또는 ESC 키 입력

Then:
  - 확인 없이 모달이 즉시 닫힌다
```

**Requirements Coverage**: E-004, E-008

---

### Scenario 6: 변경사항 있음 취소 (SC-006)

```
Given: 사용자가 설정 모달이 열려있고
  And: 하나 이상의 필드를 수정했을 때

When: 사용자가 취소 버튼 클릭 또는 ESC 키 입력

Then:
  - 변경사항을 버릴지 확인하는 다이얼로그가 표시되고
  - "예" 선택 시 모달이 닫히고
  - "아니오" 선택 시 모달이 유지된다
```

**Requirements Coverage**: E-004, E-008, S-007

---

### Scenario 7: Tab 전환 (SC-007)

```
Given: 사용자가 설정 모달이 열려있고
  And: Quick Start Tab에 있고
  And: 일부 필드를 수정했을 때

When: 사용자가 Documentation Tab 클릭

Then:
  - Documentation Tab이 활성화되고
  - Quick Start Tab의 변경사항이 유지된다
```

**Requirements Coverage**: E-002, E-007

---

### Scenario 8: API 오류 처리 (SC-008)

```
Given: 사용자가 설정 모달이 열려있고
  And: 유효한 설정값으로 수정했을 때

When: 사용자가 저장 버튼 클릭하고 서버가 오류를 반환할 때

Then:
  - 에러 메시지가 표시되고
  - 모달은 열려있는 상태를 유지하고
  - 사용자는 재시도할 수 있다
```

**Requirements Coverage**: E-006, X-005

---

### Scenario 9: 백업 파일 생성 확인 (SC-009)

```
Given: 사용자가 설정 저장을 요청하고

When: 저장이 성공적으로 완료되면

Then:
  - .moai/config/backups/ 디렉토리에 백업 파일이 생성되고
  - 파일명 형식은 config.YYYYMMDD_HHMMSS.yaml 이다
```

**Requirements Coverage**: U-003, X-002

---

### Scenario 10: 접근성 (SC-010)

```
Given: 사용자가 키보드만 사용할 때

When: 설정 모달이 열려 있을 때

Then:
  - Tab 키로 폼 필드 간 순환 이동이 가능하고
  - Enter 키로 저장이 가능하고
  - ESC 키로 모달 닫기가 가능하고
  - Screen reader가 모든 요소를 올바르게 읽는다
```

**Requirements Coverage**: Accessibility requirements

---

## 2. Quality Gates

### 2.1 Functional Gates
- [ ] Settings 버튼 클릭 시 모달 정상 표시
- [ ] 3개 Tab이 모두 렌더링됨
- [ ] 조건부 Batch가 올바르게 표시/숨김
- [ ] 저장 시 config.yaml이 업데이트됨
- [ ] 백업 파일이 생성됨
- [ ] 유효성 검증이 동작함

### 2.2 Quality Gates (TRUST 5)

#### Test-first Pillar
- [ ] 단위 테스트 커버리지 >= 85%
- [ ] 모든 EARS 요구사항에 대응하는 테스트 존재
- [ ] Edge case 테스트 포함

#### Readable Pillar
- [ ] 명확한 컴포넌트/함수 네이밍
- [ ] JSDoc/Docstring 포함
- [ ] 복잡도 지수 적정 수준 유지

#### Unified Pillar
- [ ] 코드 스타일 일관성 (ESLint/Prettier, ruff)
- [ ] Import 구조 정리
- [ ] TypeScript strict mode 준수

#### Secured Pillar
- [ ] API 키 및 민감 정보 로그 미포함
- [ ] 입력값 sanitization
- [ ] XSS 방지

#### Trackable Pillar
- [ ] Git commit message 명확
- [ ] SPEC ID 참조 포함
- [ ] 변경사항 추적 가능

### 2.3 Performance Gates
- [ ] 모달 열기 시간 < 500ms
- [ ] 저장 응답 시간 < 2s
- [ ] Tab 전환 지연 < 100ms
- [ ] 초기 렌더링 < 1s

### 2.4 Accessibility Gates
- [ ] WCAG 2.1 AA 준수
- [ ] Keyboard navigation 완전 지원
- [ ] Screen reader 호환
- [ ] Focus 관리 올바름

---

## 3. Definition of Done

A SPEC-WEB-003 implementation is considered complete when:

1. **Code Complete**
   - [ ] All components implemented
   - [ ] All API endpoints functional
   - [ ] Code review approved

2. **Testing Complete**
   - [ ] Unit tests passing (>= 85% coverage)
   - [ ] Integration tests passing
   - [ ] E2E tests passing
   - [ ] Manual testing complete

3. **Documentation Complete**
   - [ ] API documentation updated
   - [ ] Component props documented
   - [ ] User guide updated

4. **Quality Gates Passed**
   - [ ] TRUST 5 validation passed
   - [ ] Accessibility audit passed
   - [ ] Performance benchmarks met

5. **Deployment Ready**
   - [ ] No critical bugs
   - [ ] Rollback plan documented
   - [ ] Monitoring configured

---

## 4. Test Data

### 4.1 Sample Config Data

```yaml
# Valid config for testing
user:
  name: "Test User"
  email: "test@example.com"

language:
  conversation_language: "ko"
  code_comments: "en"

project:
  name: "test-project"
  type: "web_application"

git_strategy:
  mode: "personal"
  auto_commit: true

documentation:
  mode: "full_now"
  depth: "detailed"
```

### 4.2 Invalid Config Examples

```yaml
# Missing required field
user:
  name: ""  # Empty name should fail validation

# Invalid enum value
language:
  conversation_language: "korean"  # Should be "ko"

# Invalid type
git_strategy:
  mode: null  # Should be string
```

---

## 5. Verification Methods

### 5.1 Automated Testing
- **Unit Tests**: Vitest (Frontend), Pytest (Backend)
- **Integration Tests**: API 테스트 suite
- **E2E Tests**: Playwright scenarios

### 5.2 Manual Testing
- **Smoke Test**: 기본 기능 quick check
- **Exploratory Test**: Edge case 발견
- **Accessibility Test**: Keyboard + Screen reader

### 5.3 Tools
- **Lighthouse**: Performance & Accessibility score
- **Axe DevTools**: Accessibility audit
- **Postman**: API endpoint testing

---

**Document Version**: 1.0.0
**Last Updated**: 2026-01-10
**Status**: DRAFT
