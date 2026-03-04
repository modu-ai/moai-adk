# SPEC-PSR-001: Profile Setup Redesign -- StatuslineMode Exposure

## Metadata

| Field    | Value                                                                       |
| -------- | --------------------------------------------------------------------------- |
| SPEC ID  | SPEC-PSR-001                                                                |
| Title    | Profile Setup Redesign: Expose StatuslineMode in Wizard and Config Layer    |
| Status   | Planned                                                                     |
| Priority | P1 (High)                                                                   |
| Created  | 2026-03-04                                                                  |
| Domain   | internal/cli, internal/profile, internal/statusline, internal/config        |
| Depends  | SPEC-SLE-001 (Implemented -- verbose/compact/minimal modes exist in renderer) |

---

## 1. Environment

### Current System State

SPEC-SLE-001 implemented three display modes in the statusline renderer (`ModeVerbose`, `ModeDefault`, `ModeMinimal`), but the profile setup wizard has no way to select them. The mode is determined solely by the `MOAI_STATUSLINE_MODE` environment variable, with no YAML persistence.

| Component                              | State                                                                 |
| -------------------------------------- | --------------------------------------------------------------------- |
| `internal/statusline/types.go`         | `StatuslineMode` type with `ModeMinimal`, `ModeDefault`, `ModeVerbose` constants |
| `internal/cli/statusline.go`           | Mode read ONLY from `MOAI_STATUSLINE_MODE` env var (line 30-33)       |
| `statusline.yaml`                      | Has `preset`, `theme`, `segments` -- NO `mode` field                  |
| `statuslineFileConfig` struct          | Has `Preset`, `Theme`, `Segments` -- NO `Mode` field                  |
| `statuslineData` struct (sync.go)      | Has `Preset`, `Segments`, `Theme` -- NO `Mode` field                  |
| `ProfilePreferences` struct            | Has `StatuslinePreset`, `StatuslineTheme`, `StatuslineSegments` -- NO `StatuslineMode` |
| `preset` field in statusline.yaml      | Read from YAML but NEVER USED in rendering logic (dead code path)     |
| Profile wizard Display section         | Shows Preset selector and Theme selector only -- NO Mode selector     |

### Terminology Disambiguation

| Term      | Definition                                             | Values                                 | Where Used                   |
| --------- | ------------------------------------------------------ | -------------------------------------- | ---------------------------- |
| **Mode**  | Display verbosity: how many lines the statusline uses  | `minimal`, `default`, `verbose`        | `StatuslineMode` in types.go |
| **Preset** | Segment selection shortcut: which segments are visible | `full`, `compact`, `minimal`, `custom` | statusline.yaml, wizard      |
| **Theme** | Color palette for rendering                            | `default`, `catppuccin-mocha`, `catppuccin-latte` | statusline.yaml, wizard |

**CRITICAL**: `Preset "minimal"` and `Mode "minimal"` are NOT the same thing.
- **Preset "minimal"** controls *which segments* are enabled (model + context only).
- **Mode "minimal"** controls *how many lines* are rendered (1-line output).
- A user can have Preset "full" (all segments) with Mode "minimal" (1-line layout).

### Key Source Files

| File                                        | Purpose                                                |
| ------------------------------------------- | ------------------------------------------------------ |
| `internal/statusline/types.go`              | `StatuslineMode` type definition and constants         |
| `internal/statusline/renderer.go`           | `Render()` dispatches to verbose/default/minimal       |
| `internal/cli/statusline.go`                | CLI entry: reads env var, loads config, creates builder |
| `internal/profile/preferences.go`           | `ProfilePreferences` struct for user settings          |
| `internal/profile/sync.go`                  | `SyncToProjectConfig()` for profile-to-project sync    |
| `internal/cli/profile_setup.go`             | Profile wizard UI with huh forms                       |
| `internal/cli/profile_setup_translations.go`| Multi-language translations for wizard                 |
| `.moai/config/sections/statusline.yaml`     | Template for statusline configuration                  |

---

## 2. Assumptions

- **A1**: The `MOAI_STATUSLINE_MODE` env var will continue to work as an override, but YAML config becomes the primary source when env var is not set.
- **A2**: The wizard flow order (Mode -> Theme -> Preset -> Custom Segments) is more intuitive than the current order (Preset -> Theme) because mode is the highest-level display decision.
- **A3**: The `preset` field in `statusline.yaml` remains for segment selection shortcuts and does NOT control display mode (disambiguation from Mode).
- **A4**: Adding a `StatuslineMode` field to `ProfilePreferences` is backward compatible -- existing `preferences.yaml` files without this field will default to `"default"` mode.
- **A5**: Adding a `mode` field to `statusline.yaml` is backward compatible -- existing YAML files without this field will fall back to the `MOAI_STATUSLINE_MODE` env var, then to `ModeDefault`.
- **A6**: The `statuslineData` struct in `sync.go` and `statuslineFileConfig` struct in `statusline.go` should have their `Mode` field added in parallel to maintain consistency.

---

## 3. Requirements

### 3.1 Mode Selector in Wizard (NEW)

**REQ-PWR-001** (Event-Driven)
WHEN the user runs `moai profile setup`, THEN the Display section SHALL present a Mode selector as the FIRST question in the Display group, with three options: `verbose` (3-line), `default` (1-line full), `minimal` (1-line compact).

**REQ-PWR-002** (Event-Driven)
WHEN an existing `StatuslineMode` value is present in `preferences.yaml`, THEN the wizard SHALL pre-fill the Mode selector with the existing value.

**REQ-PWR-003** (Event-Driven)
WHEN no `StatuslineMode` value exists in `preferences.yaml`, THEN the wizard SHALL default to `"default"`.

**REQ-PWR-004** (Ubiquitous)
The Mode selector SHALL use translated labels in all 4 supported languages (en, ko, ja, zh).

**REQ-PWR-005** (Ubiquitous)
The Display section question ordering SHALL be: Mode -> Theme -> Preset -> Custom Segments.

**REQ-PWR-006** (Ubiquitous)
The Mode selector option labels SHALL clearly describe the display behavior:
- `verbose`: "Verbose - 3-line detailed display" / equivalent translations
- `default`: "Default - 1-line standard display" / equivalent translations
- `minimal`: "Minimal - 1-line compact display" / equivalent translations

### 3.2 Mode Persistence in Config Layer

**REQ-PWR-010** (Ubiquitous)
The `ProfilePreferences` struct in `preferences.go` SHALL include a `StatuslineMode` field of type `string` with YAML tag `statusline_mode,omitempty`.

**REQ-PWR-011** (Event-Driven)
WHEN `SyncToProjectConfig()` is called with a non-empty `StatuslineMode`, THEN the system SHALL persist the mode to `.moai/config/sections/statusline.yaml` under the `statusline.mode` key.

**REQ-PWR-012** (Event-Driven)
WHEN `statusline.yaml` is absent at the time of sync, THEN the mode SHALL default to `"default"`.

**REQ-PWR-013** (Ubiquitous)
The `statuslineData` struct in `sync.go` SHALL include a `Mode` field of type `string` with YAML tag `mode,omitempty`.

**REQ-PWR-014** (Ubiquitous)
The `statuslineFileConfig` struct in `statusline.go` SHALL include a `Mode` field of type `string`.

**REQ-PWR-015** (Event-Driven)
WHEN `loadStatuslineFileConfig()` is called, THEN the `Mode` field SHALL be populated from the `statusline.mode` key in the YAML file.

**REQ-PWR-016** (Event-Driven)
WHEN `runStatusline()` executes, THEN it SHALL determine the mode using the following fallback chain:
1. `MOAI_STATUSLINE_MODE` env var (highest priority, preserves backward compatibility)
2. `statusline.yaml` `mode` field (config-based persistence)
3. `ModeDefault` (final fallback)

**REQ-PWR-017** (Ubiquitous)
The `statusline.yaml` template SHALL include a `mode` field with default value `"default"`.

### 3.3 Backward Compatibility

**REQ-PWR-020** (Ubiquitous)
The system SHALL NOT break existing `preferences.yaml` files that lack the `statusline_mode` field. Missing field SHALL default to `"default"`.

**REQ-PWR-021** (Ubiquitous)
The system SHALL NOT break existing `statusline.yaml` files that lack the `mode` field. Missing field SHALL fall back to `MOAI_STATUSLINE_MODE` env var, then to `ModeDefault`.

**REQ-PWR-022** (Unwanted)
The system SHALL NOT remove or change the behavior of the `MOAI_STATUSLINE_MODE` environment variable.

**REQ-PWR-023** (Unwanted)
The system SHALL NOT change the meaning of existing `preset` values (`full`, `compact`, `minimal`, `custom`).

### 3.4 UX Flow Ordering

**REQ-PWR-030** (Ubiquitous)
The Display section in the wizard SHALL present questions in this order: Mode -> Theme -> Preset -> Custom Segments.

**REQ-PWR-031** (Ubiquitous)
The Custom Segments group SHALL remain hidden unless the Preset is set to `"custom"` (existing behavior preserved).

### 3.5 Mode-Preset Interaction Semantics

**REQ-PWR-040** (Ubiquitous)
Mode and Preset SHALL be orthogonal -- any combination of Mode and Preset is valid.

**REQ-PWR-041** (Ubiquitous)
Mode controls the layout (number of lines); Preset controls segment visibility. They do not affect each other.

**REQ-PWR-042** (State-Driven)
IF Mode is `"verbose"` AND Preset is `"minimal"`, THEN the system SHALL render a verbose (3-line) layout showing only minimal segments (model + context on line 1, empty lines 2-3 omitted per REQ-SLE-034).

**REQ-PWR-043** (State-Driven)
IF Mode is `"minimal"` AND Preset is `"full"`, THEN the system SHALL render a minimal (1-line) layout with all segments enabled (segment visibility is limited by line capacity).

### 3.6 Naming Disambiguation

**REQ-PWR-050** (Ubiquitous)
The wizard Mode selector SHALL use labels that distinguish it from the Preset selector:
- Mode labels focus on "lines" and "display layout"
- Preset labels focus on "segments" and "information shown"

**REQ-PWR-051** (Ubiquitous)
The Mode selector description text SHALL explicitly mention "number of lines" to prevent confusion with Preset.

---

## 4. Non-Functional Requirements

**REQ-PWR-NF-001** (Ubiquitous)
All new code SHALL achieve at least 85% test coverage with table-driven tests following TDD (`quality.yaml: test_first_required: true`).

**REQ-PWR-NF-002** (Ubiquitous)
All new code SHALL pass `go vet`, `go test -race ./...`, and `golangci-lint run` without errors.

**REQ-PWR-NF-003** (Unwanted)
The system SHALL NOT introduce any regressions in existing wizard behavior (Identity, Languages, Model Settings sections unchanged).

**REQ-PWR-NF-004** (Ubiquitous)
The `make build` command SHALL succeed after all changes, regenerating embedded template files that include the updated `statusline.yaml`.

---

## 5. Acceptance Criteria Summary

| AC ID    | Criterion                                                                              | Requirement(s)    |
| -------- | -------------------------------------------------------------------------------------- | ----------------- |
| AC-01    | Wizard Display section shows Mode selector as first question                           | REQ-PWR-001, 005  |
| AC-02    | Mode selector has 3 options: verbose, default, minimal                                 | REQ-PWR-001, 006  |
| AC-03    | Selected mode is saved to `preferences.yaml` as `statusline_mode`                      | REQ-PWR-010       |
| AC-04    | Selected mode is synced to `statusline.yaml` as `statusline.mode`                      | REQ-PWR-011       |
| AC-05    | `runStatusline()` uses config-based mode when env var is not set                       | REQ-PWR-016       |
| AC-06    | `MOAI_STATUSLINE_MODE` env var overrides config-based mode                             | REQ-PWR-016, 022  |
| AC-07    | Existing `preferences.yaml` without `statusline_mode` defaults to `"default"`          | REQ-PWR-020       |
| AC-08    | Existing `statusline.yaml` without `mode` field falls back correctly                   | REQ-PWR-021       |
| AC-09    | All 4 languages have Mode selector translations                                        | REQ-PWR-004       |
| AC-10    | Question order is Mode -> Theme -> Preset -> Custom Segments                           | REQ-PWR-030       |

---

## 6. Risk Register

| Risk ID | Risk                                                          | Likelihood | Impact | Mitigation                                                       |
| ------- | ------------------------------------------------------------- | ---------- | ------ | ---------------------------------------------------------------- |
| R1      | User confuses Mode "minimal" with Preset "minimal"            | High       | Medium | Use distinct labels: Mode = "layout/lines", Preset = "segments"  |
| R2      | Existing env var users unaware of config-based mode           | Low        | Low    | Env var takes priority (backward compatible); document in changelog |
| R3      | `preferences.yaml` migration breaks on malformed files        | Low        | Medium | `omitempty` ensures missing field is safely ignored               |
| R4      | YAML marshaling changes field order in statusline.yaml        | Low        | Low    | Functional impact none; cosmetic only                             |
| R5      | Wizard UX regression from reordered Display section           | Medium     | Medium | Maintain all existing questions; only add Mode and reorder       |
| R6      | `runStatusline` config loading adds latency                   | Low        | Low    | Config is already loaded for segments/theme; mode is one more field |

---

## 7. File Impact Matrix

### Modified Files

| File                                        | Changes                                                          | Phase |
| ------------------------------------------- | ---------------------------------------------------------------- | ----- |
| `internal/profile/preferences.go`           | Add `StatuslineMode string` field to `ProfilePreferences`        | 1     |
| `internal/profile/sync.go`                  | Add `Mode` field to `statuslineData`, merge in `syncStatusline`  | 1     |
| `internal/cli/statusline.go`                | Add `Mode` to `statuslineFileConfig`, config fallback in `runStatusline` | 2 |
| `internal/cli/profile_setup.go`             | Add Mode selector, reorder Display section (Mode -> Theme -> Preset -> Segments) | 3 |
| `internal/cli/profile_setup_translations.go`| Add 5 new translation fields for Mode selector (4 languages)    | 3     |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | Add `mode: "default"` field | 4 |

### Test Files (New or Modified)

| File                                         | Purpose                                                    | Phase |
| -------------------------------------------- | ---------------------------------------------------------- | ----- |
| `internal/profile/preferences_test.go`       | Test `StatuslineMode` field serialization/deserialization   | 1     |
| `internal/profile/sync_test.go`              | Test mode sync to statusline.yaml                          | 1     |
| `internal/cli/statusline_test.go`            | Test mode config loading and fallback chain                 | 2     |
| `internal/cli/profile_setup_test.go`         | Test Mode selector presence and ordering                   | 3     |

---

## 8. Implementation Phases

### Phase 1: Data Layer (preferences.go + sync.go)

**Primary Goal** | No dependencies

Add `StatuslineMode` field to `ProfilePreferences` struct and `Mode` field to `statuslineData` struct. Update `syncStatusline()` to merge mode preference.

Files: `internal/profile/preferences.go`, `internal/profile/sync.go`
Tests: `internal/profile/preferences_test.go`, `internal/profile/sync_test.go`

### Phase 2: Config Loading (statusline.go)

**Primary Goal** | Depends on Phase 1

Add `Mode` field to `statuslineFileConfig` struct. Update `runStatusline()` to implement the config fallback chain: env var -> YAML config -> ModeDefault.

Files: `internal/cli/statusline.go`
Tests: `internal/cli/statusline_test.go`

### Phase 3: Wizard UI (profile_setup.go + translations.go)

**Secondary Goal** | Depends on Phase 1

Add Mode selector to the Display section. Reorder Display questions to Mode -> Theme -> Preset -> Custom Segments. Add 5 new translation fields to all 4 languages.

Files: `internal/cli/profile_setup.go`, `internal/cli/profile_setup_translations.go`
Tests: `internal/cli/profile_setup_test.go`

### Phase 4: Template (statusline.yaml)

**Final Goal** | Depends on Phase 2

Add `mode: "default"` field to the statusline.yaml template. Run `make build` to regenerate embedded files.

Files: `internal/template/templates/.moai/config/sections/statusline.yaml`
Tests: N/A (template validation via `make build`)

### Phase 5: Tests (all test files)

**Final Goal** | Depends on Phases 1-4

Ensure all test files achieve 85%+ coverage. Run full test suite with `go test -race ./...`.

### Dependency Graph

```
Phase 1 (Data Layer)
  |
  +---> Phase 2 (Config Loading)
  |       |
  |       +---> Phase 4 (Template)
  |
  +---> Phase 3 (Wizard UI)

Phase 5 (Tests) -- runs after all phases
```

---

## 9. Traceability

| Requirement        | Implementation File(s)                                             | Test File(s)                           |
| ------------------ | ------------------------------------------------------------------ | -------------------------------------- |
| REQ-PWR-001-006    | `profile_setup.go`, `profile_setup_translations.go`                | `profile_setup_test.go`                |
| REQ-PWR-010-017    | `preferences.go`, `sync.go`, `statusline.go`, template `statusline.yaml` | `preferences_test.go`, `sync_test.go`, `statusline_test.go` |
| REQ-PWR-020-023    | `preferences.go`, `statusline.go`                                  | `preferences_test.go`, `statusline_test.go` |
| REQ-PWR-030-031    | `profile_setup.go`                                                 | `profile_setup_test.go`                |
| REQ-PWR-040-043    | `statusline.go` (passthrough only, renderer handles layout)        | `statusline_test.go`                   |
| REQ-PWR-050-051    | `profile_setup_translations.go`                                    | `profile_setup_test.go`                |
| REQ-PWR-NF-001-004 | All implementation files                                           | All test files                         |
