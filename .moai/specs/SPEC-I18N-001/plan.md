# SPEC-I18N-001: Implementation Plan

> **SPEC ID**: SPEC-I18N-001
> **Title**: WebUI Internationalization System (Full Coverage)
> **Phase**: Plan
> **Updated**: 2026-01-10

---

## TAG BLOCK

```yaml
spec_id: SPEC-I18N-001
plan_version: 2.0.0
estimated_files: 30+
complexity: High
total_phases: 6
total_components: 23
total_strings: 49+
```

---

## Implementation Strategy

### Approach: Phased Rollout with Custom Lightweight i18n

**Why Custom Solution**:
1. **Zero Dependencies**: No additional npm packages required
2. **Type Safety**: Full TypeScript support with compile-time key validation
3. **Minimal Bundle Impact**: ~5KB total overhead (expanded scope)
4. **Full Control**: Easy to customize for MoAI-ADK specific needs
5. **Config Integration**: Direct integration with existing config-api.ts
6. **Phased Approach**: Incremental delivery with testable milestones

---

## Phased Implementation Plan

### Phase 1: Core Infrastructure + Spec Create Dialog (Primary Goal)

**Scope**: Build foundation for multilingual support + Initial proof of concept

**Priority**: High (Must Have)

**Tasks**:
1. Create i18n type definitions (`src/lib/i18n/types.ts`)
2. Implement translation context provider (`src/lib/i18n/context.tsx`)
3. Create useTranslation hook (`src/lib/i18n/useTranslation.ts`)
4. Add UI store language state enhancement (`stores/ui-store.ts`)
5. Create all 4 translation files (en, ko, ja, zh) - initial keys
6. Refactor spec-create-dialog.tsx to use translations
7. Write unit tests for core i18n module

**Deliverables**:
- Type-safe translation system
- React context for language state
- Custom hook for component translations
- Fully localized spec-create-dialog
- Test coverage >= 85%

**Requirements Covered**: REQ-001, REQ-002, REQ-003, REQ-004, REQ-005

---

### Phase 2: Navigation and Layout Components (Secondary Goal)

**Scope**: Apply i18n to header and sidebar

**Priority**: High (Must Have)

**Tasks**:
1. Add layout namespace translations to all 4 language files
2. Refactor `header.tsx` to use translations
3. Refactor `sidebar.tsx` to use translations
4. Update tests for layout components

**Target Files**:
- `src/components/layout/header.tsx`
- `src/components/layout/sidebar.tsx`

**Translation Keys** (~10 strings):
- `layout.header.toggleSidebar`
- `layout.header.toggleTheme`
- `layout.header.connected` / `layout.header.disconnected`
- `layout.header.settings`
- `layout.sidebar.chat` / `specs` / `terminal` / `costs`
- `layout.sidebar.sessions` / `newSession` / `deleteSession` / `noSessions`

**Deliverables**:
- Localized header and sidebar
- Updated translation files

**Requirements Covered**: REQ-006

---

### Phase 3: Chat Interface Components (Tertiary Goal)

**Scope**: Apply i18n to chat-related components

**Priority**: Medium (Should Have)

**Tasks**:
1. Add chat namespace translations to all 4 language files
2. Refactor `chat-view.tsx` to use translations
3. Refactor `chat-input.tsx` to use translations
4. Refactor `message-list.tsx` to use translations
5. Update tests for chat components

**Target Files**:
- `src/components/chat/chat-view.tsx`
- `src/components/chat/chat-input.tsx`
- `src/components/chat/message-list.tsx`

**Translation Keys** (~10 strings):
- `chat.noSessionTitle`
- `chat.noSessionDescription`
- `chat.inputPlaceholder`
- `chat.sendButton`
- `chat.startConversation`

**Deliverables**:
- Localized chat interface
- Updated translation files

**Requirements Covered**: REQ-007

---

### Phase 4: SPEC Management Components (Quaternary Goal)

**Scope**: Apply i18n to remaining SPEC components (excluding dialog)

**Priority**: Medium (Should Have)

**Tasks**:
1. Add spec namespace translations (list, detail, status) to all 4 language files
2. Refactor `spec-list.tsx` to use translations
3. Refactor `spec-detail-viewer.tsx` to use translations
4. Refactor `spec-card.tsx` to use translations
5. Update tests for SPEC components

**Target Files**:
- `src/components/spec/spec-list.tsx`
- `src/components/spec/spec-detail-viewer.tsx`
- `src/components/spec/spec-card.tsx`

**Translation Keys** (~15 strings):
- `spec.monitor`
- `spec.list.filterAll` / `filterPlanned` / `filterInProgress` / `filterCompleted`
- `spec.list.newSpec` / `refresh` / `searchPlaceholder` / `emptyState`
- `spec.detail.tabSpec` / `tabPlan` / `tabAcceptance` / `loading`
- `spec.status.planned` / `inProgress` / `completed` / `blocked`

**Deliverables**:
- Localized SPEC management UI
- Updated translation files

**Requirements Covered**: REQ-008

---

### Phase 5: Config and Cost Components (Quinary Goal)

**Scope**: Apply i18n to configuration and cost analytics

**Priority**: Low (Nice to Have)

**Tasks**:
1. Add config and cost namespace translations to all 4 language files
2. Refactor `config-dialog.tsx` to use translations
3. Refactor `config-form.tsx` to use translations
4. Refactor `cost-view.tsx` to use translations
5. Refactor `cost-by-provider.tsx` to use translations
6. Update tests

**Target Files**:
- `src/components/config/config-dialog.tsx`
- `src/components/config/config-form.tsx`
- `src/components/cost/cost-view.tsx`
- `src/components/cost/cost-by-provider.tsx`

**Translation Keys** (~11 strings):
- `config.title` / `loading` / `save` / `cancel` / `reset`
- `cost.title` / `description` / `byProvider`
- `common.loading` / `error` / `retry`

**Deliverables**:
- Localized config and cost UI
- Updated translation files

**Requirements Covered**: REQ-009, REQ-010

---

### Phase 6: Terminal View Unification (Final Goal)

**Scope**: Fix mixed language issues in terminal view

**Priority**: Low (Nice to Have)

**Tasks**:
1. Add terminal namespace translations to all 4 language files
2. Refactor `terminal-view.tsx` to use translations
3. Remove hardcoded Korean error messages
4. Ensure language consistency across all states
5. Update tests

**Target Files**:
- `src/components/terminal/terminal-view.tsx`

**Translation Keys** (~3 strings):
- `terminal.error`
- `terminal.connectionError`
- `terminal.reconnecting`

**Deliverables**:
- Unified language in terminal view
- No mixed language content

**Requirements Covered**: REQ-011

---

## Technical Approach

### Architecture Design

```
┌─────────────────────────────────────────────────────────────┐
│                        App Layout                           │
│  ┌───────────────────────────────────────────────────────┐ │
│  │                   I18nProvider                         │ │
│  │  ┌─────────────────────────────────────────────────┐  │ │
│  │  │              Component Tree                      │  │ │
│  │  │                                                  │  │ │
│  │  │  ┌──────────────────────────────────────────┐   │  │ │
│  │  │  │         spec-create-dialog               │   │  │ │
│  │  │  │                                          │   │  │ │
│  │  │  │  const { t } = useTranslation('spec');   │   │  │ │
│  │  │  │  <Title>{t('createDialog.title')}</Title>│   │  │ │
│  │  │  └──────────────────────────────────────────┘   │  │ │
│  │  └─────────────────────────────────────────────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                      UI Store (Zustand)                     │
│  language: 'ko' | 'en' | 'ja' | 'zh'                       │
│  setLanguage: (lang) => void                                │
└─────────────────────────────────────────────────────────────┘
          ▲
          │ fetchConfig() on app init
          │
┌─────────────────────────────────────────────────────────────┐
│                    Config API (/api/config)                 │
│  { language: { conversation_language: 'ko' } }             │
└─────────────────────────────────────────────────────────────┘
```

### File Structure

```
src/moai_adk/web-ui/src/
├── lib/
│   └── i18n/
│       ├── index.ts              # Public exports
│       ├── types.ts              # TypeScript types (expanded)
│       ├── context.tsx           # I18nProvider component
│       ├── useTranslation.ts     # Translation hook
│       └── translations/
│           ├── index.ts          # Translation loader
│           ├── en.json           # English (49+ keys)
│           ├── ko.json           # Korean (49+ keys)
│           ├── ja.json           # Japanese (49+ keys)
│           └── zh.json           # Chinese (49+ keys)
├── stores/
│   └── ui-store.ts               # Add language state (modify)
├── app/
│   └── layout.tsx                # Wrap with I18nProvider (modify)
└── components/
    ├── layout/
    │   ├── header.tsx            # Phase 2: Apply translations
    │   └── sidebar.tsx           # Phase 2: Apply translations
    ├── chat/
    │   ├── chat-view.tsx         # Phase 3: Apply translations
    │   ├── chat-input.tsx        # Phase 3: Apply translations
    │   └── message-list.tsx      # Phase 3: Apply translations
    ├── spec/
    │   ├── spec-create-dialog.tsx # Phase 1: Apply translations
    │   ├── spec-list.tsx          # Phase 4: Apply translations
    │   ├── spec-detail-viewer.tsx # Phase 4: Apply translations
    │   └── spec-card.tsx          # Phase 4: Apply translations
    ├── config/
    │   ├── config-dialog.tsx      # Phase 5: Apply translations
    │   └── config-form.tsx        # Phase 5: Apply translations
    ├── cost/
    │   ├── cost-view.tsx          # Phase 5: Apply translations
    │   └── cost-by-provider.tsx   # Phase 5: Apply translations
    └── terminal/
        └── terminal-view.tsx      # Phase 6: Apply translations
```

### Key Implementation Details

**1. Type Definitions**:
```typescript
// Strong typing for translation keys
type TranslationKey = keyof typeof translations['en'];
type NestedKeyOf<T> = /* recursive type for nested keys */;
```

**2. Context Provider**:
```typescript
// Provides language context to component tree
interface I18nContextValue {
  language: SupportedLanguage;
  t: (key: string) => string;
  setLanguage: (lang: SupportedLanguage) => void;
}
```

**3. Translation Hook**:
```typescript
// Namespace-scoped translations
function useTranslation(namespace?: string) {
  const { language, setLanguage } = useI18nContext();
  const t = (key: string) => getTranslation(language, namespace, key);
  return { t, language, setLanguage };
}
```

---

## Risks and Mitigation

### Risk 1: Bundle Size Increase

**Risk Level**: Low
**Mitigation**:
- JSON translation files are small (~1KB each)
- Lazy load translations if needed
- Monitor bundle size after implementation

### Risk 2: Type Safety Complexity

**Risk Level**: Medium
**Mitigation**:
- Start with simple string-based keys
- Add strict typing incrementally
- Use TypeScript template literal types for key validation

### Risk 3: Missing Translations

**Risk Level**: Medium
**Mitigation**:
- Implement fallback to English for missing keys
- Add console warnings in development mode
- Create translation key validation script

### Risk 4: SSR Compatibility

**Risk Level**: Low (Current scope is client-side only)
**Mitigation**:
- Design provider to support SSR if needed later
- Use client-side language detection initially
- Document SSR extension path

---

## Dependencies

### Internal Dependencies

| Component | Dependency | Status |
|-----------|------------|--------|
| Config API | `/api/config` endpoint | Exists |
| UI Store | Zustand store | Exists, needs modification |
| App Layout | Root layout.tsx | Exists, needs wrapper |

### External Dependencies

| Package | Required | Note |
|---------|----------|------|
| None | - | Custom solution, no new deps |

---

## Definition of Done

### Phase 1 Completion Criteria (Core + Dialog)
- [ ] i18n types defined and exported
- [ ] I18nProvider implemented and integrated
- [ ] useTranslation hook working with namespaces
- [ ] All 4 translation files created (en, ko, ja, zh) with spec namespace
- [ ] spec-create-dialog fully localized
- [ ] Language detection from Config API working
- [ ] Fallback to English implemented
- [ ] Unit tests passing (>85% coverage)
- [ ] No TypeScript errors

### Phase 2 Completion Criteria (Layout)
- [ ] Layout namespace translations added to all 4 language files
- [ ] header.tsx fully localized
- [ ] sidebar.tsx fully localized
- [ ] Layout component tests updated

### Phase 3 Completion Criteria (Chat)
- [ ] Chat namespace translations added to all 4 language files
- [ ] chat-view.tsx fully localized
- [ ] chat-input.tsx fully localized
- [ ] message-list.tsx fully localized
- [ ] Chat component tests updated

### Phase 4 Completion Criteria (SPEC Management)
- [ ] SPEC list/detail/status namespace translations added
- [ ] spec-list.tsx fully localized
- [ ] spec-detail-viewer.tsx fully localized
- [ ] spec-card.tsx fully localized
- [ ] SPEC component tests updated

### Phase 5 Completion Criteria (Config & Cost)
- [ ] Config and cost namespace translations added
- [ ] config-dialog.tsx fully localized
- [ ] config-form.tsx fully localized
- [ ] cost-view.tsx fully localized
- [ ] cost-by-provider.tsx fully localized

### Phase 6 Completion Criteria (Terminal)
- [ ] Terminal namespace translations added
- [ ] terminal-view.tsx fully localized
- [ ] No mixed language content in terminal
- [ ] All hardcoded Korean removed

### Final Completion Criteria
- [ ] All 23 components localized
- [ ] All 49+ strings translated in 4 languages
- [ ] Total bundle size increase < 10KB
- [ ] Test coverage >= 85% for i18n module
- [ ] No TypeScript errors
- [ ] Documentation updated
