# SPEC-PSR-001: Implementation Plan

## Metadata

| Field    | Value             |
| -------- | ----------------- |
| SPEC ID  | SPEC-PSR-001      |
| Title    | Profile Setup Redesign: Expose StatuslineMode |
| Mode     | TDD (RED-GREEN-REFACTOR) |

---

## Overview

This plan implements SPEC-PSR-001 across 5 phases, adding `StatuslineMode` selection to the profile wizard and persisting it through the config layer to the statusline renderer.

The implementation order follows data flow: Data Layer -> Config Loading -> Wizard UI -> Template -> Tests.

---

## Phase 1: Data Layer (preferences.go + sync.go)

**Priority**: Primary Goal
**Dependencies**: None

### 1.1 Add StatuslineMode to ProfilePreferences

**File**: `internal/profile/preferences.go`

Add field to `ProfilePreferences` struct:
```go
StatuslineMode string `yaml:"statusline_mode,omitempty"` // "verbose", "default", "minimal"
```

Place it in the Display settings group, after `StatuslineTheme`.

**RED Test**: Write test verifying `StatuslineMode` round-trips through YAML marshal/unmarshal.
**GREEN**: Add the field.
**REFACTOR**: Ensure field ordering is logical within the struct.

### 1.2 Add Mode to statuslineData and syncStatusline

**File**: `internal/profile/sync.go`

Add `Mode` field to `statuslineData` struct:
```go
Mode string `yaml:"mode,omitempty"`
```

Update `syncStatusline()` to merge mode preference:
```go
if prefs.StatuslineMode != "" {
    current.Statusline.Mode = prefs.StatuslineMode
}
```

Update the guard condition in `SyncToProjectConfig()` to include `StatuslineMode`:
```go
if prefs.StatuslinePreset != "" || prefs.StatuslineTheme != "" ||
   prefs.StatuslineSegments != nil || prefs.StatuslineMode != "" {
```

**RED Tests**:
- Test that `syncStatusline` writes `mode` to YAML when `StatuslineMode` is set.
- Test that existing YAML without `mode` field gets default `"default"` for mode.
- Test that `SyncToProjectConfig` triggers statusline sync when only `StatuslineMode` is set.

**GREEN**: Add the field and merge logic.
**REFACTOR**: Verify defaults apply correctly for missing mode.

### Phase 1 File Changes

| File                             | Change Type | Description                              |
| -------------------------------- | ----------- | ---------------------------------------- |
| `internal/profile/preferences.go`| Modify      | Add `StatuslineMode` field               |
| `internal/profile/sync.go`      | Modify      | Add `Mode` to struct, merge in sync      |
| `internal/profile/preferences_test.go` | Modify | Add StatuslineMode serialization tests   |
| `internal/profile/sync_test.go`  | Modify      | Add mode sync tests                      |

---

## Phase 2: Config Loading (statusline.go)

**Priority**: Primary Goal
**Dependencies**: Phase 1

### 2.1 Add Mode to statuslineFileConfig

**File**: `internal/cli/statusline.go`

Add `Mode` field to `statuslineFileConfig` struct:
```go
type statuslineFileConfig struct {
    Preset   string
    Theme    string
    Mode     string
    Segments map[string]bool
}
```

Update `loadStatuslineFileConfig()` to read mode from YAML:
```go
return &statuslineFileConfig{
    Preset:   raw.Statusline.Preset,
    Theme:    raw.Statusline.Theme,
    Mode:     raw.Statusline.Mode,
    Segments: raw.Statusline.Segments,
}
```

Update the anonymous struct's inner `Statusline` struct to include `Mode`:
```go
Mode string `yaml:"mode"`
```

### 2.2 Implement Config Fallback Chain in runStatusline

**File**: `internal/cli/statusline.go`

Replace the current mode determination logic (lines 30-33) with a fallback chain:

```go
// Determine display mode: env var > YAML config > default
mode := statusline.StatuslineMode(os.Getenv("MOAI_STATUSLINE_MODE"))
if mode == "" && statuslineCfg != nil && statuslineCfg.Mode != "" {
    mode = statusline.StatuslineMode(statuslineCfg.Mode)
}
if mode == "" {
    mode = statusline.ModeDefault
}
```

**Important**: The mode determination must happen AFTER `loadStatuslineFileConfig()` is called, not before. Current code loads config at line 39 but determines mode at line 30. Reorder so config is loaded first, then mode is determined with the fallback chain.

**RED Tests**:
- Test that `loadStatuslineFileConfig` reads `mode` from YAML.
- Test that `runStatusline` uses env var when set (highest priority).
- Test that `runStatusline` uses YAML config mode when env var is not set.
- Test that `runStatusline` defaults to `ModeDefault` when neither is set.

**GREEN**: Implement the changes.
**REFACTOR**: Ensure code reordering is clean and readable.

### Phase 2 File Changes

| File                              | Change Type | Description                                     |
| --------------------------------- | ----------- | ----------------------------------------------- |
| `internal/cli/statusline.go`      | Modify      | Add Mode field, reorder mode determination      |
| `internal/cli/statusline_test.go` | Modify      | Add config fallback chain tests                 |

---

## Phase 3: Wizard UI (profile_setup.go + translations.go)

**Priority**: Secondary Goal
**Dependencies**: Phase 1 (StatuslineMode field must exist in ProfilePreferences)

### 3.1 Add Translation Fields

**File**: `internal/cli/profile_setup_translations.go`

Add 5 new fields to `profileSetupText` struct:
```go
// Mode selector (Display section)
StatuslineModeTitle string
StatuslineModeDesc  string
ModeVerbose         string
ModeDefault         string
ModeMinimal         string
```

Add translations for all 4 languages:

**English (en)**:
- `StatuslineModeTitle`: "Statusline display mode"
- `StatuslineModeDesc`: "Controls the number of lines in the statusline output."
- `ModeVerbose`: "Verbose - 3-line detailed display"
- `ModeDefault`: "Default - 1-line standard display"
- `ModeMinimal`: "Minimal - 1-line compact display"

**Korean (ko)**:
- `StatuslineModeTitle`: "상태줄 표시 모드"
- `StatuslineModeDesc`: "상태줄 출력의 줄 수를 제어합니다."
- `ModeVerbose`: "Verbose - 3줄 상세 표시"
- `ModeDefault`: "Default - 1줄 표준 표시"
- `ModeMinimal`: "Minimal - 1줄 간략 표시"

**Japanese (ja)**:
- `StatuslineModeTitle`: "ステータスライン表示モード"
- `StatuslineModeDesc`: "ステータスライン出力の行数を制御します。"
- `ModeVerbose`: "Verbose - 3行の詳細表示"
- `ModeDefault`: "Default - 1行の標準表示"
- `ModeMinimal`: "Minimal - 1行のコンパクト表示"

**Chinese (zh)**:
- `StatuslineModeTitle`: "状态栏显示模式"
- `StatuslineModeDesc`: "控制状态栏输出的行数。"
- `ModeVerbose`: "Verbose - 3行详细显示"
- `ModeDefault`: "Default - 1行标准显示"
- `ModeMinimal`: "Minimal - 1行简洁显示"

### 3.2 Add Mode Selector and Reorder Display Section

**File**: `internal/cli/profile_setup.go`

Add `statuslineMode` variable initialization with default:
```go
statuslineMode := existingPrefs.StatuslineMode
if statuslineMode == "" {
    statuslineMode = "default"
}
```

Redesign the Display group with 4 subgroups in order: Mode -> Theme -> Preset -> Custom Segments.

The Display group becomes:
```go
// Section 5: Display
huh.NewGroup(
    huh.NewSelect[string]().
        Title(t.StatuslineModeTitle).
        Description(t.StatuslineModeDesc).
        Options(
            huh.NewOption(t.ModeVerbose, "verbose"),
            huh.NewOption(t.ModeDefault, "default"),
            huh.NewOption(t.ModeMinimal, "minimal"),
        ).
        Value(&statuslineMode),
    huh.NewSelect[string]().
        Title(t.StatuslineThemeTitle).
        Description(t.StatuslineThemeDesc).
        Options(
            huh.NewOption(t.ThemeDefault, "default"),
            huh.NewOption(t.ThemeCatppuccinMocha, "catppuccin-mocha"),
            huh.NewOption(t.ThemeCatppuccinLatte, "catppuccin-latte"),
        ).
        Value(&statuslineTheme),
    huh.NewSelect[string]().
        Title(t.StatuslineTitle).
        Description(t.StatuslineDesc).
        Options(
            huh.NewOption(t.StatuslineFull, "full"),
            huh.NewOption(t.StatuslineCompact, "compact"),
            huh.NewOption(t.StatuslineMinimal, "minimal"),
            huh.NewOption(t.StatuslineCustom, "custom"),
        ).
        Value(&statuslinePreset),
).Title(t.DisplayTitle),
```

Update preferences building to include `StatuslineMode`:
```go
prefs := profile.ProfilePreferences{
    ...
    StatuslineMode:   statuslineMode,
    ...
}
```

**RED Tests**:
- Test that translations exist for Mode selector in all 4 languages.
- Test that Display section ordering is Mode -> Theme -> Preset -> Segments.

**GREEN**: Implement the changes.
**REFACTOR**: Ensure wizard flow feels natural.

### Phase 3 File Changes

| File                                          | Change Type | Description                                    |
| --------------------------------------------- | ----------- | ---------------------------------------------- |
| `internal/cli/profile_setup_translations.go`  | Modify      | Add 5 new fields, 4 language translations      |
| `internal/cli/profile_setup.go`               | Modify      | Add Mode selector, reorder Display section     |
| `internal/cli/profile_setup_test.go`          | Modify      | Add Mode selector and ordering tests           |

---

## Phase 4: Template (statusline.yaml)

**Priority**: Final Goal
**Dependencies**: Phase 2

### 4.1 Update statusline.yaml Template

**File**: `internal/template/templates/.moai/config/sections/statusline.yaml`

Add `mode: "default"` field:
```yaml
statusline:
  mode: "default"
  preset: "full"
  theme: "default"
  segments:
    model: true
    context: true
    output_style: true
    directory: true
    git_status: true
    claude_version: true
    moai_version: true
    git_branch: true
```

### 4.2 Regenerate Embedded Files

Run `make build` to regenerate `internal/template/embedded.go`.

### Phase 4 File Changes

| File                                                                     | Change Type | Description            |
| ------------------------------------------------------------------------ | ----------- | ---------------------- |
| `internal/template/templates/.moai/config/sections/statusline.yaml`      | Modify      | Add `mode` field       |
| `internal/template/embedded.go`                                          | Generated   | Regenerated by `make build` |

---

## Phase 5: Full Test Suite

**Priority**: Final Goal
**Dependencies**: Phases 1-4

### 5.1 Run Complete Test Suite

```bash
go test -race ./internal/profile/... ./internal/cli/... ./internal/statusline/...
go vet ./...
golangci-lint run
```

### 5.2 Coverage Verification

Ensure 85%+ coverage across all modified packages:
```bash
go test -cover ./internal/profile/...
go test -cover ./internal/cli/...
```

### 5.3 Build Verification

```bash
make build
```

---

## Implementation Order Summary

| Step | Phase | Files Modified                                      | Priority       |
| ---- | ----- | --------------------------------------------------- | -------------- |
| 1    | 1     | preferences.go, sync.go + tests                     | Primary Goal   |
| 2    | 2     | statusline.go + tests                                | Primary Goal   |
| 3    | 3     | profile_setup.go, profile_setup_translations.go + tests | Secondary Goal |
| 4    | 4     | statusline.yaml template + make build                | Final Goal     |
| 5    | 5     | Full test suite, coverage, linting                   | Final Goal     |

---

## Risks and Mitigations

| Risk                                          | Mitigation                                                      |
| --------------------------------------------- | --------------------------------------------------------------- |
| Mode/Preset naming confusion for users        | Distinct labels: Mode = "lines/layout", Preset = "segments"     |
| Reordering Display section breaks existing UX | All existing questions preserved; only added Mode and reordered |
| Config loading order in statusline.go         | Explicit reorder: load config BEFORE mode determination         |
| Backward compatibility for existing YAML      | Fallback chain: env var -> config -> default                    |

---

## Notes

- This SPEC does NOT modify the renderer (renderer.go) -- it only exposes existing rendering modes through the wizard and config.
- The `preset` field naming overlap with `Mode "minimal"` vs `Preset "minimal"` is addressed through distinct UI labels and description text.
- Git operations (branch creation, commits, PR) are handled by manager-git agent, not this plan.
