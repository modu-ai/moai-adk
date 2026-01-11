# SPEC-I18N-001: Acceptance Criteria

> **SPEC ID**: SPEC-I18N-001
> **Title**: WebUI Internationalization System (Full Coverage)
> **Phase**: Acceptance Criteria
> **Updated**: 2026-01-10

---

## TAG BLOCK

```yaml
spec_id: SPEC-I18N-001
test_framework: vitest
coverage_target: 85%
total_test_scenarios: 17
total_acceptance_criteria: 10
phases: 6
```

---

## Acceptance Criteria

### AC-001: Language Detection on Application Load

**Related Requirement**: REQ-001

**Given** the WebUI application is loading for the first time
**And** the user has configured `conversation_language: "ja"` in `.moai/config/sections/language.yaml`
**When** the application fetches configuration from `/api/config`
**Then** the UI language should be set to Japanese
**And** the Spec Create Dialog should display Japanese text

---

### AC-002: Translation System Type Safety

**Related Requirement**: REQ-002

**Given** a developer is using the translation system
**When** they call `t('spec.createDialog.invalidKey')`
**Then** TypeScript should show a compile-time error for invalid key
**And** valid keys like `t('spec.createDialog.title')` should compile without errors

---

### AC-003: Spec Create Dialog Full Localization

**Related Requirement**: REQ-003

**Given** the UI language is set to Korean ("ko")
**When** the user opens the Spec Create Dialog
**Then** the following elements should display in Korean:
| Element | Korean Text |
|---------|-------------|
| Title | 새 SPEC 생성 |
| Description | 구현하고자 하는 기능이나 요구사항을 설명해주세요... |
| Label | 상세 지시사항 |
| Cancel Button | 취소 |
| Create Button | SPEC 생성 |

---

### AC-004: Fallback to English for Unsupported Language

**Related Requirement**: REQ-004

**Given** the configuration contains an unsupported language code (e.g., "fr")
**When** the application loads
**Then** the UI should fallback to English
**And** no error should be displayed to the user

---

### AC-005: Navigation and Layout Localization

**Related Requirement**: REQ-006

**Given** the UI language is set to a supported language
**When** the user views the header and sidebar components
**Then** all navigation elements should display in the selected language:
- Header: Toggle buttons tooltips, connection status, settings label
- Sidebar: Navigation tabs (Chat, SPECs, Terminal, Costs), session management labels

---

### AC-006: Chat Interface Localization

**Related Requirement**: REQ-007

**Given** the UI language is set to a supported language
**When** the user views the chat interface
**Then** all chat-related text should display in the selected language:
- Empty state messages
- Input placeholders and button labels
- Conversation prompts

---

### AC-007: SPEC Management Localization

**Related Requirement**: REQ-008

**Given** the UI language is set to a supported language
**When** the user views the SPEC management components
**Then** all SPEC-related text should display in the selected language:
- SPEC Monitor title and filters
- Status badges (Planned, In Progress, Completed, Blocked)
- Tab labels and action buttons

---

### AC-008: Configuration Dialog Localization

**Related Requirement**: REQ-009

**Given** the UI language is set to a supported language
**When** the user opens the configuration dialog
**Then** all configuration text should display in the selected language:
- Dialog title and section labels
- Form labels and button text
- Loading and error states

---

### AC-009: Cost Analytics Localization

**Related Requirement**: REQ-010

**Given** the UI language is set to a supported language
**When** the user views the cost analytics section
**Then** all cost-related text should display in the selected language:
- Page title and descriptions
- Provider section headers
- Summary labels

---

### AC-010: Terminal View Language Consistency

**Related Requirement**: REQ-011

**Given** the UI language is set to a supported language
**When** the user views the terminal component
**Then** all terminal UI text should display in the selected language
**And** no mixed language content should appear (Korean/English mixed)
**And** error messages should be in the selected language

---

## Test Scenarios

### TC-001: Language Detection - Korean Configuration

```gherkin
Feature: Language Detection from Config API
  As a user with Korean language preference
  I want the WebUI to automatically display in Korean
  So that I can use the application in my preferred language

  Scenario: Detect Korean language on startup
    Given the config API returns conversation_language as "ko"
    When the WebUI application initializes
    Then the UI store language state should be "ko"
    And the I18nProvider should provide Korean translations
```

---

### TC-002: Language Detection - Default Fallback

```gherkin
Feature: Default Language Fallback
  As a user with no language preference set
  I want the WebUI to default to English
  So that I have a consistent experience

  Scenario: Fallback to English when config is missing
    Given the config API returns no language configuration
    When the WebUI application initializes
    Then the UI store language state should be "en"
    And English translations should be displayed
```

---

### TC-003: Translation System - Type Safe Keys

```gherkin
Feature: Type-Safe Translation Keys
  As a developer implementing i18n
  I want compile-time validation of translation keys
  So that typos are caught before runtime

  Scenario: Valid translation key compiles
    Given a component using useTranslation hook
    When calling t('createDialog.title')
    Then the TypeScript compiler should not report errors

  Scenario: Invalid translation key fails to compile
    Given a component using useTranslation hook
    When calling t('createDialog.nonExistentKey')
    Then the TypeScript compiler should report an error
```

---

### TC-004: Spec Create Dialog - English Display

```gherkin
Feature: Spec Create Dialog Localization
  As an English-speaking user
  I want the Spec Create Dialog in English
  So that I can understand all UI elements

  Scenario: Display English translations
    Given the UI language is set to "en"
    When I open the Spec Create Dialog
    Then the dialog title should be "New SPEC"
    And the description should contain "Describe the feature or requirement"
    And the label should be "Detailed Instructions"
    And the cancel button should say "Cancel"
    And the create button should say "Create SPEC"
```

---

### TC-005: Spec Create Dialog - Korean Display

```gherkin
Feature: Spec Create Dialog Korean Localization
  As a Korean-speaking user
  I want the Spec Create Dialog in Korean
  So that I can use the application in my native language

  Scenario: Display Korean translations
    Given the UI language is set to "ko"
    When I open the Spec Create Dialog
    Then the dialog title should be "새 SPEC 생성"
    And the label should be "상세 지시사항"
    And the cancel button should say "취소"
    And the create button should say "SPEC 생성"
```

---

### TC-006: Spec Create Dialog - Japanese Display

```gherkin
Feature: Spec Create Dialog Japanese Localization
  As a Japanese-speaking user
  I want the Spec Create Dialog in Japanese
  So that I can use the application in my native language

  Scenario: Display Japanese translations
    Given the UI language is set to "ja"
    When I open the Spec Create Dialog
    Then the dialog title should be "新規SPEC作成"
    And the label should be "詳細な指示"
    And the cancel button should say "キャンセル"
    And the create button should say "SPEC作成"
```

---

### TC-007: Spec Create Dialog - Chinese Display

```gherkin
Feature: Spec Create Dialog Chinese Localization
  As a Chinese-speaking user
  I want the Spec Create Dialog in Chinese
  So that I can use the application in my native language

  Scenario: Display Chinese translations
    Given the UI language is set to "zh"
    When I open the Spec Create Dialog
    Then the dialog title should be "新建SPEC"
    And the label should be "详细说明"
    And the cancel button should say "取消"
    And the create button should say "创建SPEC"
```

---

### TC-008: Fallback Behavior - Unsupported Language

```gherkin
Feature: Unsupported Language Fallback
  As a user with an unsupported language preference
  I want graceful fallback to English
  So that the application remains usable

  Scenario: Fallback for unsupported language code
    Given the config API returns conversation_language as "fr"
    When the WebUI application initializes
    Then the UI language should fallback to "en"
    And English translations should be displayed
    And no error should be shown to the user

  Scenario: Fallback for invalid language code
    Given the config API returns conversation_language as "invalid"
    When the WebUI application initializes
    Then the UI language should fallback to "en"
```

---

### TC-009: Header Component - English Display

```gherkin
Feature: Header Component Localization
  As an English-speaking user
  I want the header navigation in English
  So that I can understand all navigation elements

  Scenario: Display English translations in header
    Given the UI language is set to "en"
    When I view the header component
    Then the toggle sidebar tooltip should be "Toggle sidebar"
    And the toggle theme tooltip should be "Toggle theme"
    And the connection status should be "Connected" or "Disconnected"
    And the settings button should say "Settings"
```

---

### TC-010: Sidebar Component - Korean Display

```gherkin
Feature: Sidebar Component Korean Localization
  As a Korean-speaking user
  I want the sidebar navigation in Korean
  So that I can navigate the application in my native language

  Scenario: Display Korean translations in sidebar
    Given the UI language is set to "ko"
    When I view the sidebar component
    Then the navigation tabs should be "채팅", "SPEC", "터미널", "비용"
    And the sessions section should say "세션"
    And the new session button should say "새 세션"
    And the empty state should say "세션이 없습니다"
```

---

### TC-011: Chat View - Empty State Localization

```gherkin
Feature: Chat View Empty State Localization
  As a user with language preference
  I want the chat empty state in my language
  So that I understand how to start using the chat

  Scenario: Display localized empty state in Japanese
    Given the UI language is set to "ja"
    When I view the chat view with no session selected
    Then the title should be "セッション未選択"
    And the description should contain instructions in Japanese
```

---

### TC-012: Chat Input - Placeholder Localization

```gherkin
Feature: Chat Input Localization
  As a user with language preference
  I want the chat input placeholder in my language
  So that I know how to use the input field

  Scenario: Display localized placeholder in Chinese
    Given the UI language is set to "zh"
    When I view the chat input component
    Then the placeholder should be in Chinese
    And the send button aria-label should be in Chinese
```

---

### TC-013: SPEC List - Filter and Actions Localization

```gherkin
Feature: SPEC List Localization
  As a user managing SPECs
  I want the SPEC list interface in my language
  So that I can effectively manage specifications

  Scenario: Display localized SPEC list in Korean
    Given the UI language is set to "ko"
    When I view the SPEC list component
    Then the page title should be "SPEC 모니터"
    And the filter options should be "전체", "계획됨", "진행 중", "완료됨"
    And the new SPEC button should say "새 SPEC"
    And the search placeholder should be in Korean
```

---

### TC-014: SPEC Card - Status Badge Localization

```gherkin
Feature: SPEC Card Status Localization
  As a user viewing SPEC cards
  I want status badges in my language
  So that I understand the SPEC status at a glance

  Scenario: Display localized status badges in Japanese
    Given the UI language is set to "ja"
    When I view a SPEC card with status "in_progress"
    Then the status badge should display "進行中"

  Scenario: Display localized status badges in Chinese
    Given the UI language is set to "zh"
    When I view a SPEC card with status "completed"
    Then the status badge should display "已完成"
```

---

### TC-015: Config Dialog - Form Labels Localization

```gherkin
Feature: Config Dialog Localization
  As a user configuring settings
  I want the configuration dialog in my language
  So that I can understand all settings options

  Scenario: Display localized config dialog in Korean
    Given the UI language is set to "ko"
    When I open the configuration dialog
    Then the dialog title should be in Korean
    And the save button should say "저장"
    And the cancel button should say "취소"
```

---

### TC-016: Cost View - Analytics Labels Localization

```gherkin
Feature: Cost Analytics Localization
  As a user viewing cost analytics
  I want the cost information in my language
  So that I can understand usage costs

  Scenario: Display localized cost view in Japanese
    Given the UI language is set to "ja"
    When I view the cost analytics page
    Then the page title should be in Japanese
    And the provider section headers should be in Japanese
```

---

### TC-017: Terminal View - No Mixed Language

```gherkin
Feature: Terminal View Language Consistency
  As a user viewing the terminal
  I want consistent language without mixing
  So that I have a cohesive user experience

  Scenario: Display consistent language in terminal (English)
    Given the UI language is set to "en"
    When I view the terminal component with an error state
    Then the error message should be in English
    And no Korean text should be displayed

  Scenario: Display consistent language in terminal (Korean)
    Given the UI language is set to "ko"
    When I view the terminal component with an error state
    Then the error message should be in Korean
    And no English text should be displayed
```

---

## Quality Gates

### Code Quality

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| Test Coverage (i18n module) | >= 85% | `vitest --coverage` |
| Test Coverage (overall) | >= 80% | `vitest --coverage` |
| TypeScript Errors | 0 | `tsc --noEmit` |
| Lint Errors | 0 | `eslint src/lib/i18n/ src/components/` |

### Performance

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| Bundle Size Impact | < 10KB (expanded scope) | `npm run build` size comparison |
| Language Switch Time | < 50ms | Performance profiling |
| Initial Load Overhead | < 100ms | Lighthouse metrics |
| Translation Lookup | O(1) | Code review |

### Accessibility

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| RTL Support Ready | Planned | Architecture review |
| Screen Reader Compatible | Yes | ARIA labels maintained |
| Language Consistency | 100% | Manual QA + automated tests |

### Coverage by Phase

| Phase | Components | Strings | Coverage Target |
|-------|------------|---------|-----------------|
| Phase 1 | 1 (dialog) | ~8 | 100% |
| Phase 2 | 2 (layout) | ~10 | 100% |
| Phase 3 | 3 (chat) | ~10 | 100% |
| Phase 4 | 3 (spec) | ~15 | 100% |
| Phase 5 | 4 (config/cost) | ~11 | 100% |
| Phase 6 | 1 (terminal) | ~3 | 100% |
| **Total** | **23** | **49+** | **100%** |

---

## Verification Methods

### Unit Tests

```typescript
// Example test structure
describe('useTranslation', () => {
  it('should return correct translation for valid key', () => {
    const { result } = renderHook(() => useTranslation('spec'));
    expect(result.current.t('createDialog.title')).toBe('New SPEC');
  });

  it('should respect current language', () => {
    // Set language to Korean
    const { result } = renderHook(() => useTranslation('spec'), {
      wrapper: ({ children }) => (
        <I18nProvider language="ko">{children}</I18nProvider>
      ),
    });
    expect(result.current.t('createDialog.title')).toBe('새 SPEC 생성');
  });
});
```

### Integration Tests

```typescript
describe('SpecCreateDialog i18n', () => {
  it('should display Korean text when language is ko', async () => {
    render(<SpecCreateDialog open onOpenChange={() => {}} onCreateSpec={async () => {}} />, {
      wrapper: ({ children }) => (
        <I18nProvider language="ko">{children}</I18nProvider>
      ),
    });

    expect(screen.getByText('새 SPEC 생성')).toBeInTheDocument();
    expect(screen.getByText('취소')).toBeInTheDocument();
  });
});
```

---

## Definition of Done Checklist

### Phase 1 (Core + Dialog)
- [ ] AC-001 to AC-004 verified
- [ ] TC-001 to TC-008 passing
- [ ] i18n module test coverage >= 85%

### Phase 2 (Layout)
- [ ] AC-005 verified
- [ ] TC-009, TC-010 passing
- [ ] header.tsx and sidebar.tsx localized

### Phase 3 (Chat)
- [ ] AC-006 verified
- [ ] TC-011, TC-012 passing
- [ ] All chat components localized

### Phase 4 (SPEC Management)
- [ ] AC-007 verified
- [ ] TC-013, TC-014 passing
- [ ] All SPEC components (except dialog) localized

### Phase 5 (Config & Cost)
- [ ] AC-008, AC-009 verified
- [ ] TC-015, TC-016 passing
- [ ] Config and cost components localized

### Phase 6 (Terminal)
- [ ] AC-010 verified
- [ ] TC-017 passing
- [ ] No mixed language in terminal

### Final Verification
- [ ] All acceptance criteria (AC-001 to AC-010) verified
- [ ] All test scenarios (TC-001 to TC-017) passing
- [ ] Test coverage >= 85% for i18n module
- [ ] Test coverage >= 80% overall
- [ ] Zero TypeScript compiler errors
- [ ] Zero ESLint errors
- [ ] Bundle size increase < 10KB
- [ ] Language switching works without page reload
- [ ] All four languages display correctly in all 23 components
- [ ] Fallback behavior working as expected
- [ ] No mixed language content anywhere
- [ ] Cross-component consistency verified
- [ ] Code reviewed and approved

### Cross-Component Consistency Verification
- [ ] All translation keys follow naming convention
- [ ] All components use useTranslation hook
- [ ] No hardcoded strings in any component
- [ ] All 4 language files have identical key structure
- [ ] No missing translations in any language file
