# SPEC-I18N-001: WebUI Internationalization (i18n) System

> **SPEC ID**: SPEC-I18N-001
> **Title**: WebUI Internationalization System (Full Coverage)
> **Created**: 2026-01-10
> **Updated**: 2026-01-10
> **Status**: Planned
> **Priority**: High
> **Assigned**: expert-frontend

---

## TAG BLOCK

```yaml
spec_id: SPEC-I18N-001
title: WebUI Internationalization System (Full Coverage)
domain: frontend
scope: web-ui/src/components/
dependencies:
  - config-api (language detection)
  - .moai/config/sections/language.yaml
related_specs: []
components_count: 23
strings_count: 49+
```

---

## Environment

### Scope Overview

- **Total Components**: 23개 파일
- **Total Strings**: 약 49개 이상의 번역 대상 문자열
- **Languages**: 4개 (English, Korean, Japanese, Chinese)

### Component Categories

#### Layout Components (3 files, ~10 strings)

| File | Strings | Description |
|------|---------|-------------|
| `header.tsx` | "Toggle sidebar", "Toggle theme", "Connected"/"Disconnected", "Settings" | 헤더 네비게이션 요소 |
| `sidebar.tsx` | "Chat", "SPECs", "Terminal", "Costs", "Sessions", "New session", "Delete session", "No sessions yet" | 사이드바 네비게이션 |

#### Chat Components (4 files, ~10 strings)

| File | Strings | Description |
|------|---------|-------------|
| `chat-view.tsx` | "No Session Selected", "Create a new session..." | 채팅 뷰 상태 메시지 |
| `chat-input.tsx` | Placeholders, aria-labels, button labels | 입력 필드 요소 |
| `message-list.tsx` | "Start a Conversation" | 빈 상태 메시지 |

#### SPEC Management Components (4 files, ~15 strings)

| File | Strings | Description |
|------|---------|-------------|
| `spec-list.tsx` | "SPEC Monitor", status filters, buttons, search placeholder | SPEC 목록 UI |
| `spec-create-dialog.tsx` | 다이얼로그 전체 텍스트 (기존 SPEC 범위) | SPEC 생성 다이얼로그 |
| `spec-detail-viewer.tsx` | Tab labels, loading states | SPEC 상세 뷰어 |
| `spec-card.tsx` | Status labels | SPEC 카드 상태 표시 |

#### Configuration Components (4 files, ~6 strings)

| File | Strings | Description |
|------|---------|-------------|
| `config-dialog.tsx` | Title, badges, loading | 설정 다이얼로그 |
| `config-form.tsx` | Buttons | 설정 폼 버튼 |

#### Cost Analytics Components (4 files, ~5 strings)

| File | Strings | Description |
|------|---------|-------------|
| `cost-view.tsx` | Titles, descriptions | 비용 분석 뷰 |
| `cost-by-provider.tsx` | Section headers | 프로바이더별 비용 |

#### Terminal Components (1 file, ~3 strings)

| File | Strings | Description |
|------|---------|-------------|
| `terminal-view.tsx` | 오류 메시지 (현재 한국어/영어 혼재) | 터미널 뷰 - 언어 통일 필요 |

### Current State (spec-create-dialog.tsx)

- **Target File**: `src/moai_adk/web-ui/src/components/spec/spec-create-dialog.tsx`
- **Hardcoded Text (Korean)**:
  - Dialog Title: "새 SPEC 생성"
  - Dialog Description: "구현하고자 하는 기능이나 요구사항을 설명해주세요..."
  - Label: "상세 지시사항"
  - Placeholder: "만들고 싶은 기능을 설명해주세요..."
  - Helper Text: "생성하려면 Cmd+Enter..."
  - Buttons: "취소", "생성 중...", "SPEC 생성"

### Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Next.js | 16.x | App Router framework |
| React | 19.x | UI library |
| TypeScript | 5.9+ | Type-safe JavaScript |
| Tailwind CSS | 3.x | Styling |

### Language Configuration Source

- **Config File**: `.moai/config/sections/language.yaml`
- **Key Field**: `language.conversation_language` (ko, en, ja, zh)
- **API Endpoint**: `/api/config` (existing)

---

## Assumptions

### Technical Assumptions

| Assumption | Confidence | Evidence | Risk if Wrong |
|------------|------------|----------|---------------|
| Config API provides language setting | High | config-api.ts exists with fetchConfig() | Requires backend modification |
| Next.js App Router supports client-side i18n | High | Standard pattern in Next.js 16 | Alternative approach needed |
| Four languages sufficient (EN/KO/JA/ZH) | High | Matches MoAI-ADK core languages | Easy to extend later |

### Business Assumptions

| Assumption | Confidence | Evidence | Risk if Wrong |
|------------|------------|----------|---------------|
| Users prefer UI in conversation_language | High | Consistent UX pattern | Make configurable |
| No server-side rendering needed for i18n | Medium | Dialog is client-side | Add SSR support if needed |

---

## Requirements

### REQ-001: Language Detection (Event-Driven)

**EARS Pattern**: Event-Driven

**WHEN** the WebUI application initializes
**THEN** the system shall fetch the `conversation_language` setting from the Config API (`/api/config`)
**AND** store the detected language in the UI state (Zustand store)

**Rationale**: Ensures UI language matches user's configured preference.

---

### REQ-002: Translation System Architecture (Ubiquitous)

**EARS Pattern**: Ubiquitous

The system **shall** implement a lightweight i18n solution with the following characteristics:
- Type-safe translation keys using TypeScript
- Translation files organized by language (en.json, ko.json, ja.json, zh.json)
- Namespace support for component-level translations
- Runtime language switching without page reload

**Rationale**: Enables maintainable, scalable multilingual support.

---

### REQ-003: Spec Create Dialog Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the Spec Create Dialog shall display all text elements in the selected language:
- Dialog title
- Dialog description
- Label text
- Placeholder text
- Helper text
- Button labels (Cancel, Creating..., Create SPEC)

**Rationale**: Primary implementation target for i18n proof of concept.

---

### REQ-004: Fallback Language Handling (Unwanted/State-Driven)

**EARS Pattern**: Combined (Unwanted + State-Driven)

**IF** the configured language is not supported
**THEN** the system shall fallback to English (en) as the default language

The system **shall NOT**:
- Display missing translation keys to users
- Show mixed language content in a single component
- Crash or fail silently on missing translations

**Rationale**: Ensures graceful degradation and consistent UX.

---

### REQ-005: Translation Extensibility (Optional)

**EARS Pattern**: Optional

**WHERE** possible, the i18n architecture should:
- Support future addition of new languages without code changes
- Allow per-component translation namespaces
- Enable server-side translation loading for SSR components

**Rationale**: Future-proofs the implementation for broader WebUI i18n adoption.

---

### REQ-006: Navigation and Layout Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the Navigation and Layout components shall display all text elements in the selected language:

**header.tsx**:
- Toggle sidebar tooltip
- Toggle theme tooltip
- Connection status ("Connected" / "Disconnected")
- Settings button label

**sidebar.tsx**:
- Navigation tabs: "Chat", "SPECs", "Terminal", "Costs"
- Session section: "Sessions", "New session", "Delete session"
- Empty state: "No sessions yet"

**Rationale**: 네비게이션은 사용자가 가장 먼저 접하는 UI 요소로, 일관된 언어 경험 제공 필수.

---

### REQ-007: Chat Interface Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the Chat interface components shall display all text elements in the selected language:

**chat-view.tsx**:
- Empty state title: "No Session Selected"
- Empty state description: "Create a new session to start chatting with Claude"

**chat-input.tsx**:
- Input placeholder text
- Send button aria-label
- File attachment labels

**message-list.tsx**:
- Empty conversation message: "Start a Conversation"
- Loading indicators

**Rationale**: 채팅 인터페이스는 핵심 사용자 상호작용 영역으로, 완전한 로컬라이제이션 필요.

---

### REQ-008: SPEC Management Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the SPEC Management components shall display all text elements in the selected language:

**spec-list.tsx**:
- Page title: "SPEC Monitor"
- Status filter labels: "All", "Planned", "In Progress", "Completed"
- Action buttons: "New SPEC", "Refresh"
- Search placeholder
- Empty state message

**spec-detail-viewer.tsx**:
- Tab labels: "Specification", "Plan", "Acceptance"
- Loading state text
- Error messages

**spec-card.tsx**:
- Status badges: "Planned", "In Progress", "Completed", "Blocked"
- Action tooltips

**Rationale**: SPEC 관리는 MoAI-ADK의 핵심 워크플로우이며, 전체 로컬라이제이션 필요.

---

### REQ-009: Configuration Dialog Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the Configuration components shall display all text elements in the selected language:

**config-dialog.tsx**:
- Dialog title
- Section badges
- Loading indicators

**config-form.tsx**:
- Form labels
- Button labels: "Save", "Cancel", "Reset"
- Validation messages

**Rationale**: 설정 다이얼로그는 사용자 환경 설정에 중요한 역할을 하며, 명확한 이해 필요.

---

### REQ-010: Cost Analytics Localization (State-Driven)

**EARS Pattern**: State-Driven

**IF** the current UI language is set to a supported language (en, ko, ja, zh)
**THEN** the Cost Analytics components shall display all text elements in the selected language:

**cost-view.tsx**:
- Page title
- Description text
- Summary labels

**cost-by-provider.tsx**:
- Provider section headers
- Cost labels
- Chart legends

**Rationale**: 비용 분석 정보는 정확한 이해가 필요하며, 사용자의 언어로 제공되어야 함.

---

### REQ-011: Terminal View Language Unification (Unwanted + State-Driven)

**EARS Pattern**: Combined (Unwanted + State-Driven)

**IF** the current UI language is set to a supported language
**THEN** the Terminal View shall display all UI text in the selected language

The system **shall NOT**:
- Display mixed language content (현재 한국어/영어 혼재 상태 해결)
- Show hardcoded Korean error messages when other languages are selected
- Have inconsistent language in error states

**terminal-view.tsx**:
- Error messages
- Status indicators
- Connection status

**Rationale**: 터미널 뷰의 언어 혼재 문제를 해결하여 일관된 사용자 경험 제공.

---

## Specifications

### i18n Library Selection

**Recommended**: Custom lightweight solution (no external dependency)

| Option | Pros | Cons | Recommendation |
|--------|------|------|----------------|
| **Custom Solution** | Zero dependencies, full control, type-safe | Manual implementation | **Recommended** |
| next-intl | SSR support, Next.js integration | Adds dependency, complex setup | Consider for full SSR |
| react-i18next | Feature-rich, popular | Heavy for simple use case | Overkill |

**Custom Solution Architecture**:

```typescript
// src/lib/i18n/types.ts
type SupportedLanguage = 'en' | 'ko' | 'ja' | 'zh';

interface Translations {
  // Layout namespace
  layout: {
    header: {
      toggleSidebar: string;
      toggleTheme: string;
      connected: string;
      disconnected: string;
      settings: string;
    };
    sidebar: {
      chat: string;
      specs: string;
      terminal: string;
      costs: string;
      sessions: string;
      newSession: string;
      deleteSession: string;
      noSessions: string;
    };
  };
  // Chat namespace
  chat: {
    noSessionTitle: string;
    noSessionDescription: string;
    inputPlaceholder: string;
    sendButton: string;
    startConversation: string;
  };
  // SPEC namespace
  spec: {
    monitor: string;
    createDialog: {
      title: string;
      description: string;
      label: string;
      placeholder: string;
      helperText: string;
      cancelButton: string;
      creatingButton: string;
      createButton: string;
    };
    list: {
      filterAll: string;
      filterPlanned: string;
      filterInProgress: string;
      filterCompleted: string;
      newSpec: string;
      refresh: string;
      searchPlaceholder: string;
      emptyState: string;
    };
    detail: {
      tabSpec: string;
      tabPlan: string;
      tabAcceptance: string;
      loading: string;
    };
    status: {
      planned: string;
      inProgress: string;
      completed: string;
      blocked: string;
    };
  };
  // Config namespace
  config: {
    title: string;
    loading: string;
    save: string;
    cancel: string;
    reset: string;
  };
  // Cost namespace
  cost: {
    title: string;
    description: string;
    byProvider: string;
  };
  // Terminal namespace
  terminal: {
    error: string;
    connectionError: string;
    reconnecting: string;
  };
  // Common namespace
  common: {
    loading: string;
    error: string;
    retry: string;
    close: string;
  };
}
```

### Translation File Structure

```
src/moai_adk/web-ui/src/
├── lib/
│   └── i18n/
│       ├── index.ts           # i18n exports
│       ├── types.ts           # Type definitions
│       ├── context.tsx        # React context provider
│       ├── useTranslation.ts  # Translation hook
│       └── translations/
│           ├── en.json        # English
│           ├── ko.json        # Korean
│           ├── ja.json        # Japanese
│           └── zh.json        # Chinese
```

### Translation Content Example

```json
// en.json
{
  "spec": {
    "createDialog": {
      "title": "New SPEC",
      "description": "Describe the feature or requirement you want to implement. Claude will help you write detailed specifications interactively.",
      "label": "Detailed Instructions",
      "placeholder": "Describe the feature you want to build...\n\nExample: Implement a user authentication system using JWT tokens. It should support email/password login and Google OAuth social login.",
      "helperText": "Press Cmd+Enter (Mac) or Ctrl+Enter (Windows) to create",
      "cancelButton": "Cancel",
      "creatingButton": "Creating...",
      "createButton": "Create SPEC"
    }
  }
}
```

### Integration Pattern

```typescript
// Component usage
import { useTranslation } from '@/lib/i18n';

function SpecCreateDialog() {
  const { t } = useTranslation('spec');

  return (
    <DialogTitle>{t('createDialog.title')}</DialogTitle>
    // ...
  );
}
```

### State Management Integration

```typescript
// stores/ui-store.ts enhancement
interface UIState {
  language: SupportedLanguage;
  setLanguage: (lang: SupportedLanguage) => void;
}

// Initialize from config on app load
const initializeLanguage = async () => {
  const config = await fetchConfig();
  if (config.success) {
    useUIStore.getState().setLanguage(
      config.data.language?.conversation_language ?? 'en'
    );
  }
};
```

---

## Constraints

### Technical Constraints

| Constraint | Description |
|------------|-------------|
| Bundle Size | i18n solution must add < 5KB to bundle |
| Type Safety | All translation keys must be type-checked |
| No External Dependencies | Prefer custom solution over heavy libraries |

### Performance Constraints

| Metric | Target |
|--------|--------|
| Language Switch | < 50ms (no page reload) |
| Initial Load | i18n overhead < 100ms |
| Translation Lookup | O(1) lookup time |

---

## Traceability

| Requirement | Test Scenario | Acceptance Criteria | Phase |
|-------------|---------------|---------------------|-------|
| REQ-001 | TC-001, TC-002 | AC-001 | Phase 1 |
| REQ-002 | TC-003 | AC-002 | Phase 1 |
| REQ-003 | TC-004, TC-005, TC-006, TC-007 | AC-003 | Phase 1 |
| REQ-004 | TC-008 | AC-004 | Phase 1 |
| REQ-005 | TC-003 | AC-002 | Phase 1 |
| REQ-006 | TC-009, TC-010 | AC-005 | Phase 2 |
| REQ-007 | TC-011, TC-012 | AC-006 | Phase 3 |
| REQ-008 | TC-013, TC-014 | AC-007 | Phase 4 |
| REQ-009 | TC-015 | AC-008 | Phase 5 |
| REQ-010 | TC-016 | AC-009 | Phase 5 |
| REQ-011 | TC-017 | AC-010 | Phase 6 |

---

## References

- [MoAI-ADK Language Configuration](.moai/config/sections/language.yaml)
- [Config API Implementation](src/moai_adk/web-ui/src/lib/config-api.ts)
- [Target Component](src/moai_adk/web-ui/src/components/spec/spec-create-dialog.tsx)
