---
id: SPEC-WEB-003
version: "1.0.0"
status: "draft"
created: "2026-01-10"
updated: "2026-01-10"
author: "GOOS행"
tags: ["web", "ui", "settings", "configuration", "dialog"]
---

# Implementation Plan: SPEC-WEB-003

## 1. Overview

### 1.1 Objective
Web UI에서 프로젝트 설정(0-Project)을 시각적으로 구성하고 수정할 수 있는 Dialog 컴포넌트를 개발한다.

### 1.2 Scope
- **Frontend**: React/TypeScript 기반 설정 Dialog 컴포넌트 개발
- **Backend**: 설정 파일 CRUD API 엔드포인트 개발
- **Integration**: 기존 ConfigurationManager와 YAML 파일 시스템 연동

### 1.3 Success Metrics
- 모달 열기 시간 < 500ms
- 저장 응답 시간 < 2s
- 테스트 커버리지 >= 85%
- TRUST 5 프레임워크 준수

---

## 2. Technical Stack

### 2.1 Frontend
| Component | Technology | Version |
|-----------|------------|---------|
| Framework | React | 19.x |
| Language | TypeScript | 5.9+ |
| UI Library | shadcn/ui | Latest |
| State Management | Zustand | 5.x |
| HTTP Client | Fetch API | Native |
| Validation | Zod | 3.x |

### 2.2 Backend
| Component | Technology | Version |
|-----------|------------|---------|
| Framework | FastAPI | 0.115+ |
| Validation | Pydantic | 2.9+ |
| Config Management | ConfigurationManager | Existing |
| YAML Processing | PyYAML | Existing |

---

## 3. Implementation Phases

### Phase 1: Backend API (Primary Goal)

#### 3.1.1 Create Config Router
**File**: `src/moai_adk/web/routers/config.py`

```python
# API Endpoints
GET    /api/project/config    # 설정 조회
PUT    /api/project/config    # 설정 저장
POST   /api/project/config/validate   # 설정 유효성 검증
GET    /api/project/config/backup     # 백업 목록 조회
```

#### 3.1.2 Pydantic Schemas
**File**: `src/moai_adk/web/schemas/config.py`

```python
class ProjectConfigResponse(BaseModel):
    exists: bool
    config: Optional[ProjectConfig]
    defaults: ProjectConfig

class ConfigUpdateRequest(BaseModel):
    config: ProjectConfig
    create_backup: bool = True

class ConfigUpdateResponse(BaseModel):
    success: bool
    backup_path: Optional[str]
    message: str
```

#### 3.1.3 Integration with ConfigurationManager
- 기존 `ConfigurationManager` 클래스 재사용
- 백업 파일 생성 로직 활용
- YAML 파일 직렬화/역직렬화

#### 3.1.4 Deliverables
- [x] Config router 구현
- [x] Pydantic schemas 정의
- [x] 단위 테스트 (pytest)
- [x] API 통합 테스트

---

### Phase 2: Frontend Components (Primary Goal)

#### 3.2.1 Main Dialog Component
**File**: `src/moai_adk/web-ui/src/components/settings/project-settings-dialog.tsx`

```typescript
interface ProjectSettingsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// Features:
- Dialog layout (shadcn/ui Dialog)
- Tab navigation (shadcn/ui Tabs)
- Form state management
- Save/Cancel handlers
```

#### 3.2.2 Tab Components
**Files**:
- `src/moai_adk/web-ui/src/components/settings/tab-quick-start.tsx`
- `src/moai_adk/web-ui/src/components/settings/tab-documentation.tsx`
- `src/moai_adk/web-ui/src/components/settings/tab-git-automation.tsx`

```typescript
// Each tab contains:
- Batch groups based on tab_schema
- Form fields with validation
- Conditional rendering based on state
```

#### 3.2.3 Form Field Components
**File**: `src/moai_adk/web-ui/src/components/settings/form-fields.tsx`

```typescript
// Reusable form components:
- TextField (string input)
- SelectField (dropdown)
- ToggleField (boolean)
- RadioGroupField (single choice)
```

#### 3.2.4 Deliverables
- [x] ProjectSettingsDialog 컴포넌트
- [x] 3개 Tab 컴포넌트
- [x] 재사용 가능한 Form Field 컴포넌트
- [x] 컴포넌트 단위 테스트 (Vitest)

---

### Phase 3: State Management (Secondary Goal)

#### 3.3.1 Zustand Store
**File**: `src/moai_adk/web-ui/src/stores/config-store.ts`

```typescript
interface ConfigStore {
  // State
  config: ProjectConfig | null;
  isDirty: boolean;
  isValid: boolean;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchConfig: () => Promise<void>;
  updateConfig: (updates: Partial<ProjectConfig>) => void;
  saveConfig: () => Promise<boolean>;
  reset: () => void;
}
```

#### 3.3.2 API Client
**File**: `src/moai_adk/web-ui/src/lib/api/config.ts`

```typescript
export const configApi = {
  getConfig: () => Promise<ProjectConfigResponse>,
  updateConfig: (req: ConfigUpdateRequest) => Promise<ConfigUpdateResponse>,
  validateConfig: (config: ProjectConfig) => Promise<ValidationResult>,
};
```

#### 3.3.3 Deliverables
- [x] Config Zustand store
- [x] API client 모듈
- [x] Store 테스트

---

### Phase 4: Header Integration (Secondary Goal)

#### 3.4.1 Header Component Modification
**File**: `src/moai_adk/web-ui/src/components/header.tsx`

```typescript
// Add Settings button
<Button onClick={() => setSettingsDialogOpen(true)}>
  <SettingsIcon />
</Button>

// Add Dialog
<ProjectSettingsDialog
  open={settingsDialogOpen}
  onOpenChange={setSettingsDialogOpen}
/>
```

#### 3.4.2 Deliverables
- [x] Header에 Settings 버튼 추가
- [x] Dialog open/close 상태 연결
- [x] 통합 테스트

---

### Phase 5: Polish & Testing (Final Goal)

#### 3.5.1 Error Handling
- API 에러 표시 (Toast/Alert)
- 유효성 검증 에러 표시
- 파일 시스템 에러 처리

#### 3.5.2 Accessibility
- Keyboard navigation (Tab, Enter, ESC)
- Screen reader support
- Focus management
- ARIA labels

#### 3.5.3 Responsive Design
- Mobile (768px+)
- Tablet (1024px+)
- Desktop (1280px+)

#### 3.5.4 Deliverables
- [x] 에러 처리 완료
- [x] 접근성 audit 통과
- [x] 반응형 디자인 확인
- [x] E2E 테스트 (Playwright)

---

## 4. Dependencies

### 4.1 Internal Dependencies
| Module | Purpose | Status |
|--------|---------|--------|
| `src/moai_adk/web/schema.py` | Tab schema definitions | Existing |
| `src/moai_adk/core/configuration.py` | ConfigurationManager | Existing |
| `src/moai_adk/web-ui/src/components/header.tsx` | Settings button integration | Existing |

### 4.2 External Dependencies
| Package | Purpose | Version |
|---------|---------|---------|
| @radix-ui/react-dialog | Dialog primitives | Latest |
| @radix-ui/react-tabs | Tab primitives | Latest |
| zustand | State management | ^5.0.0 |
| zod | Schema validation | ^3.0.0 |

---

## 5. Architecture Design

### 5.1 Component Hierarchy
```
App
└── Header
    ├── SettingsButton
    └── ProjectSettingsDialog
        ├── Dialog (shadcn/ui)
        ├── Tabs (shadcn/ui)
        │   ├── TabList
        │   └── TabPanels
        │       ├── QuickStartTab
        │       ├── DocumentationTab
        │       └── GitAutomationTab
        └── DialogFooter
            ├── CancelButton
            └── SaveButton
```

### 5.2 Data Flow
```
User Input
    ↓
Form Component (Zustand Store)
    ↓
API Client (fetch)
    ↓
Backend Router (FastAPI)
    ↓
ConfigurationManager
    ↓
YAML File (config.yaml)
```

---

## 6. Risk Management

### 6.1 Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| YAML 파싱 오류 | HIGH | MEDIUM | Pydantic 스키마 검증 + 상세 에러 메시지 |
| 동시 설정 수정 | MEDIUM | LOW | 파일 Lock 또는 Last-Write-Wins 전략 |
| State 동기화 문제 | MEDIUM | MEDIUM | Zustand store 단일 출처 유지 |
| 백업 파일 디스크 부족 | MEDIUM | LOW | 저장 전 디스크 공간 확인 |

### 6.2 UX Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| 설정 복잡도로 인한 혼란 | MEDIUM | HIGH | Tab 구조로 그룹화 + 도움말 툴팁 |
| 변경사항 실수로 분실 | HIGH | MEDIUM | 저장 전 확인 + 자동 저장 옵션 |
| 모바일 환경 UI 문제 | MEDIUM | MEDIUM | 반응형 디자인 테스트 |

---

## 7. Testing Strategy

### 7.1 Unit Tests
- Backend: Pytest (FastAPI 테스트)
- Frontend: Vitest (컴포넌트 테스트)

### 7.2 Integration Tests
- API 엔드포인트 테스트
- Frontend-Backend 통합 테스트

### 7.3 E2E Tests
- Playwright로 전체 사용자 시나리오 테스트

### 7.4 Coverage Target
- **Backend**: >= 90%
- **Frontend**: >= 85%

---

## 8. Deployment Plan

### 8.1 Deployment Steps
1. Backend API 배포 (FastAPI 재시작)
2. Frontend 빌드 및 배포
3. 설정 파일 백업 확인
4. smoke test 실행

### 8.2 Rollback Plan
- 이전 config.yaml 백업 복원
- 이전 버전 코드로 revert

---

## 9. Milestones

| Priority | Milestone | Description |
|----------|-----------|-------------|
| Primary | Phase 1 Complete | Backend API 개발 완료 |
| Primary | Phase 2 Complete | Frontend Dialog 컴포넌트 개발 완료 |
| Secondary | Phase 3 Complete | State management 완료 |
| Secondary | Phase 4 Complete | Header 통합 완료 |
| Optional | Phase 5 Complete | Polish & E2E 테스트 완료 |

---

**Document Version**: 1.0.0
**Last Updated**: 2026-01-10
**Status**: DRAFT
