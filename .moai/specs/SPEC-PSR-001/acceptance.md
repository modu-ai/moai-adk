# SPEC-PSR-001: Acceptance Criteria

## Metadata

| Field    | Value             |
| -------- | ----------------- |
| SPEC ID  | SPEC-PSR-001      |
| Title    | Profile Setup Redesign: Expose StatuslineMode |

---

## AC-01: Mode Selector Appears as First Question in Display Section

**Requirements**: REQ-PWR-001, REQ-PWR-005

```gherkin
Scenario: Mode selector is the first question in Display section
  Given the user runs "moai profile setup"
  And the user completes Language, Identity, Languages, and Model Settings sections
  When the Display section is presented
  Then the first question SHALL be the Mode selector
  And the Mode selector SHALL appear before the Theme selector
  And the Theme selector SHALL appear before the Preset selector
```

---

## AC-02: Mode Selector Has 3 Options

**Requirements**: REQ-PWR-001, REQ-PWR-006

```gherkin
Scenario: Mode selector provides verbose, default, and minimal options
  Given the Display section is shown
  When the user views the Mode selector options
  Then option 1 SHALL be "Verbose - 3-line detailed display"
  And option 2 SHALL be "Default - 1-line standard display"
  And option 3 SHALL be "Minimal - 1-line compact display"

Scenario: Mode selector labels distinguish from Preset labels
  Given the Display section is shown
  When the user views both Mode and Preset selectors
  Then Mode labels SHALL reference "lines" and "display layout"
  And Preset labels SHALL reference "segments" and "information shown"
```

---

## AC-03: Selected Mode Saved to preferences.yaml

**Requirements**: REQ-PWR-010

```gherkin
Scenario: Mode selection persists to preferences.yaml
  Given the user runs "moai profile setup"
  And the user selects "verbose" as the statusline mode
  When the wizard completes
  Then preferences.yaml SHALL contain "statusline_mode: verbose"

Scenario: Mode field uses omitempty for backward compatibility
  Given a ProfilePreferences struct with empty StatuslineMode
  When the struct is marshaled to YAML
  Then the output SHALL NOT contain "statusline_mode" key
```

---

## AC-04: Selected Mode Synced to statusline.yaml

**Requirements**: REQ-PWR-011, REQ-PWR-013

```gherkin
Scenario: Mode is synced to project statusline.yaml
  Given a ProfilePreferences with StatuslineMode = "verbose"
  And the current directory is inside a MoAI project
  When SyncToProjectConfig() is called
  Then .moai/config/sections/statusline.yaml SHALL contain "mode: verbose"

Scenario: Mode sync preserves existing statusline.yaml fields
  Given an existing statusline.yaml with preset = "full" and theme = "catppuccin-mocha"
  When SyncToProjectConfig() is called with StatuslineMode = "minimal"
  Then statusline.yaml SHALL contain mode = "minimal"
  And statusline.yaml SHALL still contain preset = "full"
  And statusline.yaml SHALL still contain theme = "catppuccin-mocha"

Scenario: SyncToProjectConfig triggers when only StatuslineMode is set
  Given a ProfilePreferences with only StatuslineMode = "verbose" (all other statusline fields empty)
  When SyncToProjectConfig() is called
  Then syncStatusline SHALL be invoked
  And statusline.yaml SHALL contain "mode: verbose"
```

---

## AC-05: runStatusline Uses Config-Based Mode

**Requirements**: REQ-PWR-016

```gherkin
Scenario: Config-based mode used when env var is not set
  Given MOAI_STATUSLINE_MODE env var is NOT set
  And statusline.yaml contains mode = "verbose"
  When runStatusline() executes
  Then the statusline SHALL render in verbose mode

Scenario: Default mode used when neither env var nor config is set
  Given MOAI_STATUSLINE_MODE env var is NOT set
  And statusline.yaml does NOT contain a mode field
  When runStatusline() executes
  Then the statusline SHALL render in default mode
```

---

## AC-06: Environment Variable Overrides Config

**Requirements**: REQ-PWR-016, REQ-PWR-022

```gherkin
Scenario: Env var takes priority over config
  Given MOAI_STATUSLINE_MODE env var is set to "minimal"
  And statusline.yaml contains mode = "verbose"
  When runStatusline() executes
  Then the statusline SHALL render in minimal mode (env var wins)

Scenario: Env var behavior is preserved from pre-PSR-001
  Given MOAI_STATUSLINE_MODE env var is set to "verbose"
  And no statusline.yaml exists
  When runStatusline() executes
  Then the statusline SHALL render in verbose mode
```

---

## AC-07: Backward Compatibility -- preferences.yaml

**Requirements**: REQ-PWR-020

```gherkin
Scenario: Legacy preferences.yaml without statusline_mode
  Given a preferences.yaml file that was created before SPEC-PSR-001
  And the file does NOT contain "statusline_mode" key
  When ReadPreferences() is called
  Then ProfilePreferences.StatuslineMode SHALL be "" (empty string)

Scenario: Wizard defaults empty StatuslineMode to "default"
  Given ReadPreferences() returns a ProfilePreferences with empty StatuslineMode
  When the wizard initializes
  Then the Mode selector SHALL be pre-filled with "default"
```

---

## AC-08: Backward Compatibility -- statusline.yaml

**Requirements**: REQ-PWR-021

```gherkin
Scenario: Legacy statusline.yaml without mode field
  Given a statusline.yaml that was created before SPEC-PSR-001
  And the file does NOT contain "mode" key
  When loadStatuslineFileConfig() is called
  Then statuslineFileConfig.Mode SHALL be "" (empty string)

Scenario: Fallback chain with missing mode field
  Given statusline.yaml does NOT contain "mode" key
  And MOAI_STATUSLINE_MODE env var is NOT set
  When runStatusline() executes
  Then mode SHALL fall back to ModeDefault
```

---

## AC-09: Translations for All 4 Languages

**Requirements**: REQ-PWR-004

```gherkin
Scenario Outline: Mode selector has translations in <language>
  Given the user selects <language> as the conversation language
  When the Display section is presented
  Then the Mode selector title SHALL be in <language>
  And the Mode selector description SHALL be in <language>
  And all three mode options SHALL have labels in <language>

  Examples:
    | language |
    | en       |
    | ko       |
    | ja       |
    | zh       |

Scenario: Translation completeness check
  Given the profileSetupTexts map
  When all 4 language entries are inspected
  Then each entry SHALL contain non-empty values for:
    | Field                |
    | StatuslineModeTitle  |
    | StatuslineModeDesc   |
    | ModeVerbose          |
    | ModeDefault          |
    | ModeMinimal          |
```

---

## AC-10: Question Order Validation

**Requirements**: REQ-PWR-030, REQ-PWR-031

```gherkin
Scenario: Display section question order
  Given the user enters the Display section of the wizard
  When all form fields are rendered
  Then the order SHALL be:
    1. Mode selector (StatuslineModeTitle)
    2. Theme selector (StatuslineThemeTitle)
    3. Preset selector (StatuslineTitle)
    4. Custom Segments group (hidden unless preset = "custom")

Scenario: Custom Segments remain hidden for non-custom presets
  Given the user selects Preset = "full"
  When the Display section renders
  Then the Custom Segments group SHALL be hidden

Scenario: Custom Segments visible for custom preset
  Given the user selects Preset = "custom"
  When the Display section renders
  Then the Custom Segments group SHALL be visible with 8 toggle options
```

---

## Definition of Done

- [ ] All 10 acceptance criteria pass
- [ ] `go test -race ./internal/profile/... ./internal/cli/...` passes
- [ ] `go vet ./...` passes with zero errors
- [ ] `golangci-lint run` passes with zero errors
- [ ] Test coverage >= 85% for modified packages
- [ ] `make build` succeeds (embedded templates regenerated)
- [ ] No regressions in existing wizard sections (Identity, Languages, Model Settings)
- [ ] `MOAI_STATUSLINE_MODE` env var still works as override
