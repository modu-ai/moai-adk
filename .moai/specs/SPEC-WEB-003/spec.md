---
id: SPEC-WEB-003
version: "1.0.0"
status: "draft"
created: "2026-01-10"
updated: "2026-01-10"
author: "GOOS행"
priority: "HIGH"
tags: ["web", "ui", "settings", "configuration", "dialog"]
---

## HISTORY

| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|----------|--------|
| 1.0.0 | 2026-01-10 | 초기 SPEC 작성 | GOOS행 |

---

# SPEC-WEB-003: Web Settings UI - 0-Project Configuration Dialog

## 1. Overview

### 1.1 Purpose
Web UI에서 프로젝트 설정(0-Project)을 시각적으로 구성하고 수정할 수 있는 Dialog 컴포넌트를 개발한다.

### 1.2 Scope
- Frontend: React/TypeScript 기반 설정 Dialog 컴포넌트
- Backend: 설정 파일 CRUD API 엔드포인트
- Integration: 기존 ConfigurationManager와 YAML 파일 시스템 연동

### 1.3 Context
현재 Web UI는 설정 수정을 위해 CLI를 통해서만 가능하다. Web UI 내에서 직관적인 설정 관리가 필요하다.

---

## 2. Environment

### 2.1 Target Environment
- **Browser**: Modern browsers (Chrome 120+, Firefox 120+, Safari 17+, Edge 120+)
- **Runtime**: Node.js 20+ (Server), React 19 (Client)
- **Screen Size**: Desktop (1280px+), Tablet (768px+)

### 2.2 Technical Context
- **Frontend Framework**: React 19, TypeScript 5.9+
- **UI Library**: shadcn/ui (Dialog, Tabs, Form, Button components)
- **Backend Framework**: FastAPI 0.115+
- **State Management**: Zustand 5.x
- **Configuration Format**: YAML (config.yaml)

### 2.3 Dependencies
- 기존 `src/moai_adk/web/schema.py`의 tab_schema 활용
- 기존 `src/moai_adk/core/configuration.py`의 ConfigurationManager 재사용
- 기존 Web UI Header 컴포넌트와의 통합

---

## 3. Assumptions

### 3.1 Technical Assumptions
- 사용자 브라우저는 WebSocket을 지원한다
- 서버는 config.yaml 파일에 직접 접근 권한을 가진다
- 백업 파일 생성을 위한 디스크 공간이 충분하다

### 3.2 Business Assumptions
- 사용자는 CLI 명령어 없이 Web UI만으로 설정을 완료하기를 원한다
- 설정 변경 시 즉시 반영되기를 기대한다

### 3.3 User Assumptions
- 사용자는 YAML 문법을 알 필요가 없다
- 사용자는 기본적인 Web UI 조작 방법을 알고 있다

---

## 4. Requirements (EARS Format)

### 4.1 Ubiquitous Requirements (항상 활성)

| ID | Requirement |
|----|-------------|
| U-001 | 시스템은 항상 설정 파일 존재 여부를 확인해야 한다 |
| U-002 | 시스템은 항상 입력 값의 유효성을 검증해야 한다 |
| U-003 | 시스템은 항상 백업 파일을 생성 후 저장해야 한다 |
| U-004 | 시스템은 항상 변경사항 저장 여부를 사용자에게 명확히 표시해야 한다 |

### 4.2 Event-Driven Requirements (이벤트 기반)

| ID | WHEN Event | THEN Response |
|----|------------|---------------|
| E-001 | WHEN Settings 버튼 클릭 | THEN 설정 모달(Dialog)을 표시하고 기존 설정값을 로드한다 |
| E-002 | WHEN Tab 전환 | THEN 해당 Tab에 속하는 Batch(질문 그룹)를 표시한다 |
| E-003 | WHEN 저장 버튼 클릭 | THEN 유효성 검증 후 YAML 파일을 저장하고 성공 메시지를 표시한다 |
| E-004 | WHEN 취소 버튼 클릭 | THEN 변경사항 없이 모달을 닫는다 |
| E-005 | WHEN 저장 성공 | THEN 성공 토스트 메시지를 표시하고 모달을 닫는다 |
| E-006 | WHEN 저장 실패 | THEN 에러 메시지를 표시하고 모달을 유지한다 |
| E-007 | WHEN 필드 값 변경 | THEN 변경사항 추적 상태를 업데이트한다 |
| E-008 | WHEN ESC 키 입력 | THEN 변경사항이 없으면 모달을 닫고, 있으면 확인을 요청한다 |

### 4.3 State-Driven Requirements (상태 기반)

| ID | IF Condition | THEN Response |
|----|--------------|---------------|
| S-001 | IF git_strategy.mode == 'personal' | THEN Tab 3에 Personal 관련 Batch만 표시한다 |
| S-002 | IF git_strategy.mode == 'team' | THEN Tab 3에 Team 관련 Batch만 표시한다 |
| S-003 | IF documentation_mode == 'full_now' | THEN documentation_depth 관련 질문을 표시한다 |
| S-004 | IF 설정 파일이 존재하지 않으면 | THEN 기본값으로 폼을 초기화한다 |
| S-005 | IF 설정 파일이 존재하면 | THEN 기존 값으로 폼을 초기화한다 |
| S-006 | IF 필수 필드가 유효하지 않으면 | THEN 저장 버튼을 비활성화한다 |
| S-007 | IF 변경사항이 존재하면 | THEN 저장 전 확인 경고를 표시한다 |

### 4.4 Unwanted Requirements (금지 사항)

| ID | Prohibited Behavior |
|----|---------------------|
| X-001 | 시스템은 유효하지 않은 설정을 저장하면 안 된다 |
| X-002 | 시스템은 백업 없이 기존 설정을 덮어쓰면 안 된다 |
| X-003 | 시스템은 API 키나 민감 정보를 로그에 출력하면 안 된다 |
| X-004 | 시스템은 사용자 동의 없이 설정을 자동 저장하면 안 된다 |
| X-005 | 시스템은 파일 시스템 오류를 무시하면 안 된다 |

### 4.5 Optional Requirements (선택 사항)

| ID | Feature |
|----|---------|
| O-001 | 가능하면 설정 변경 내역 추적(Revision History)을 제공한다 |
| O-002 | 가능하면 설정값 실시간 검증 피드백을 제공한다 |
| O-003 | 가능하면 설정값을 가져오기/내보내기(Import/Export) 기능을 제공한다 |
| O-004 | 가능하면 다크 모드에 대응하는 스타일을 제공한다 |

---

## 5. Specifications

### 5.1 Functional Specifications

#### 5.1.1 Dialog Layout
```
+--------------------------------------------------+
| Project Configuration                    [X]     |
+--------------------------------------------------+
| +----------+ +----------------------------------+ |
| | Quick    | | Batch 1: User Information         | |
| | Start    | | - Name: [_____________]            | |
| +----------+ | - Email: [_____________]           | |
| | Docs     | +----------------------------------+ |
| +----------+ | Batch 2: Language Settings         | |
| | Git Auto | | - Conversation: [ko ▼]            | |
| +----------+ +----------------------------------+ |
|              |                                  | |
| +----------+ +----------------------------------+ |
| |          | | [Cancel]            [Save]       | |
| +----------+ +----------------------------------+ |
+--------------------------------------------------+
```

#### 5.1.2 Tab Structure
| Tab | Content | Batch Groups |
|-----|---------|--------------|
| Tab 1: Quick Start | 프로젝트 기본 정보 | user_info, project_type |
| Tab 2: Documentation | 문서화 설정 | documentation_mode, documentation_depth |
| Tab 3: Git Automation | Git 전략 설정 | git_strategy (personal/team) |

#### 5.1.3 Backend API Specification

**GET /api/project/config**
```yaml
Response 200:
  type: object
  properties:
    exists: boolean
    config:
      user: object
      language: object
      project: object
      git_strategy: object
    defaults: object  # exists=false일 때 기본값
```

**PUT /api/project/config**
```yaml
Request Body:
  type: object
  properties:
    config: object
    create_backup: boolean (default: true)

Response 200:
  type: object
  properties:
    success: boolean
    backup_path: string | null
    message: string

Response 422:
  type: object
  properties:
    detail: array  # Pydantic validation errors
```

### 5.2 Data Models

#### 5.2.1 Configuration Schema
```python
class ProjectConfigSchema(BaseModel):
    user: UserInfoSchema
    language: LanguageConfigSchema
    project: ProjectMetadataSchema
    git_strategy: GitStrategySchema
    documentation: DocumentationConfigSchema
```

#### 5.2.2 Frontend State Model
```typescript
interface ConfigState {
  config: ProjectConfig | null;
  isDirty: boolean;
  isValid: boolean;
  activeTab: string;
  isLoading: boolean;
  error: string | null;
}
```

### 5.3 UI/UX Specifications

#### 5.3.1 Visual Hierarchy
1. **Header**: Dialog 제목 + 닫기 버튼
2. **Navigation**: 왼쪽 Tab 메뉴 (수직)
3. **Content**: 오른쪽 Form 영역
4. **Footer**: 취소/저장 버튼 (오른쪽 정렬)

#### 5.3.2 Interaction Rules
- **Tab 전환**: 변경사항 유지, 경고 없음
- **저장 시도**: 유효성 검증 후 성공/실패 피드백
- **취소 시도**: 변경사항 있으면 확인 다이얼로그
- **외부 클릭**: 변경사항 있으면 무시, 없으면 닫기

---

## 6. Traceability

### 6.1 Tag Mapping

| Tag | File | Line |
|-----|------|------|
| SPEC-WEB-003 | All | All |

### 6.2 Requirement to Component Mapping

| Requirement ID | Component | File Path |
|----------------|-----------|-----------|
| U-001, U-002, E-001~E-008 | ProjectSettingsDialog | components/settings/project-settings-dialog.tsx |
| S-001~S-007 | Tab Form Components | components/settings/tab-*.tsx |
| E-003, E-005, E-006 | Config API | web/routers/config.py |
| U-003, X-002 | Config Manager | core/configuration.py |

### 6.3 Acceptance Criteria Mapping

| Scenario ID | Requirements |
|-------------|--------------|
| SC-001 (모달 열기) | E-001, S-004, S-005 |
| SC-002 (저장 성공) | E-003, E-005, U-002, U-003 |
| SC-003 (조건부 Batch) | S-001, S-002, S-003 |
| SC-004 (유효성 검증) | U-002, X-001, S-006 |

---

## 7. Success Criteria

### 7.1 Functional Success
- [ ] Settings 버튼 클릭 시 모달 정상 표시
- [ ] 3개 Tab이 모두 렌더링됨
- [ ] 조건부 Batch가 올바르게 표시/숨김
- [ ] 저장 시 config.yaml이 업데이트됨
- [ ] 백업 파일이 생성됨
- [ ] 유효성 검증이 동작함

### 7.2 Quality Success
- [ ] TRUST 5 프레임워크 준수
- [ ] 테스트 커버리지 85% 이상
- [ ] 접근성 기준 (WCAG 2.1 AA) 준수
- [ ] 로그에 민감 정보 미포함

### 7.3 Performance Success
- [ ] 모달 열기 시간 < 500ms
- [ ] 저장 응답 시간 < 2s
- [ ] Tab 전환 지연 < 100ms

---

## 8. Risks and Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| YAML 파싱 오류 | HIGH | MEDIUM | Pydantic 스키마 검증 + 에러 메시지 상세 표시 |
| 동시 설정 수정 | MEDIUM | LOW | 파일 Lock 또는 Last-Write-Wins 전략 |
| 백업 파일 디스크 부족 | MEDIUM | LOW | 저장 전 디스크 공간 확인 |
| 설정 파일 손상 | HIGH | LOW | 백업 자동 복구 기능 |

---

**Document Version**: 1.0.0
**Last Updated**: 2026-01-10
**Status**: DRAFT
