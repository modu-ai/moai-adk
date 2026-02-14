# SPEC-STATUSLINE-001: Acceptance Criteria

## SPEC Reference

| Field   | Value               |
| ------- | ------------------- |
| SPEC ID | SPEC-STATUSLINE-001 |
| Title   | Statusline Segment Configuration |

---

## 1. Configuration File

### AC-SL-001: Default template deployed

**Given** a fresh `moai init` is executed
**When** template deployment completes
**Then** `.moai/config/sections/statusline.yaml` exists with `preset: "full"` and all 8 segments set to `true`

### AC-SL-002: Config preserved during update

**Given** the user has modified `statusline.yaml` to set `preset: "compact"` and `git_status: false`
**When** `moai update` runs and performs 3-way YAML merge
**Then** the user's `preset: "compact"` and `git_status: false` values are preserved in the merged output

### AC-SL-003: Config file format valid

**Given** a `statusline.yaml` file with the documented schema
**When** parsed by `yaml.Unmarshal`
**Then** the `statusline.preset` field is a string and `statusline.segments` is a map of 8 string keys to boolean values

---

## 2. Wizard Integration

### AC-SL-010: Preset question appears in wizard

**Given** the user runs `moai init` or `moai update -c`
**When** the wizard reaches the statusline configuration step
**Then** a "Statusline preset" select question is displayed with 4 options: Full, Compact, Minimal, Custom

### AC-SL-011: Custom triggers segment questions

**Given** the user selects "Custom" as the statusline preset
**When** the wizard advances to the next questions
**Then** 8 individual segment enable/disable questions are displayed, one for each segment

### AC-SL-012: Non-custom skips segment questions

**Given** the user selects "Full" (or "Compact" or "Minimal") as the statusline preset
**When** the wizard advances to the next questions
**Then** the 8 individual segment questions are NOT displayed (skipped by condition)

### AC-SL-013: Translations available

**Given** the user has set conversation language to Korean (ko)
**When** the statusline preset question is displayed
**Then** the question title, description, and option labels are displayed in Korean

**Given** the user has set conversation language to Japanese (ja)
**When** the statusline preset question is displayed
**Then** the question title, description, and option labels are displayed in Japanese

**Given** the user has set conversation language to Chinese (zh)
**When** the statusline preset question is displayed
**Then** the question title, description, and option labels are displayed in Chinese

### AC-SL-014: WizardResult populated

**Given** the user selects "Compact" as the statusline preset
**When** the wizard completes
**Then** `WizardResult.StatuslinePreset` equals `"compact"` and `WizardResult.StatuslineSegments` is nil or empty

**Given** the user selects "Custom" and enables model, context, git_branch; disables the rest
**When** the wizard completes
**Then** `WizardResult.StatuslinePreset` equals `"custom"` and `WizardResult.StatuslineSegments` contains `{model: true, context: true, output_style: false, directory: false, git_status: false, claude_version: false, moai_version: false, git_branch: true}`

---

## 3. Configuration Application

### AC-SL-020: Full preset writes correct YAML

**Given** `WizardResult.StatuslinePreset` is `"full"`
**When** `applyWizardConfig()` executes
**Then** `.moai/config/sections/statusline.yaml` contains `preset: "full"` and all 8 segments set to `true`

### AC-SL-021: Compact preset writes correct YAML

**Given** `WizardResult.StatuslinePreset` is `"compact"`
**When** `applyWizardConfig()` executes
**Then** `.moai/config/sections/statusline.yaml` contains `preset: "compact"` with `model: true`, `context: true`, `git_status: true`, `git_branch: true` and remaining segments set to `false`

### AC-SL-022: Minimal preset writes correct YAML

**Given** `WizardResult.StatuslinePreset` is `"minimal"`
**When** `applyWizardConfig()` executes
**Then** `.moai/config/sections/statusline.yaml` contains `preset: "minimal"` with `model: true`, `context: true` and remaining 6 segments set to `false`

### AC-SL-023: Custom preset writes user selections

**Given** `WizardResult.StatuslinePreset` is `"custom"` and `StatuslineSegments` has `model: true, context: true, directory: true` and all others `false`
**When** `applyWizardConfig()` executes
**Then** `.moai/config/sections/statusline.yaml` contains `preset: "custom"` with the exact segment toggle values from `StatuslineSegments`

### AC-SL-024: Empty preset skips write

**Given** `WizardResult.StatuslinePreset` is empty (user skipped or question not asked)
**When** `applyWizardConfig()` executes
**Then** `statusline.yaml` is NOT modified by the wizard (existing content preserved)

---

## 4. Renderer Segment Filtering

### AC-SL-030: All segments shown with full config

**Given** `SegmentConfig` has all 8 segments set to `true`
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the output contains all 8 segments separated by " | "

### AC-SL-031: Disabled segment hidden

**Given** `SegmentConfig` has `output_style: false` and all others `true`
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the output does NOT contain the output_style segment (no style name displayed)
**And** the output contains the remaining 7 segments

### AC-SL-032: Multiple segments disabled

**Given** `SegmentConfig` has only `model: true` and `context: true`, all others `false`
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the output contains only the model and context segments

### AC-SL-033: Nil config shows all segments (backward compatibility)

**Given** `SegmentConfig` is nil
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the output contains all 8 segments (identical to current behavior)

### AC-SL-034: Empty config shows all segments (backward compatibility)

**Given** `SegmentConfig` is an empty map
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the output contains all 8 segments

### AC-SL-035: Unknown segment key defaults to enabled

**Given** `SegmentConfig` contains `model: true, context: true` but is missing the `git_branch` key entirely
**And** `StatusData` has valid data for all segments
**When** the renderer renders in default mode
**Then** the `git_branch` segment is displayed (missing key defaults to enabled)

### AC-SL-036: All segments disabled shows fallback

**Given** `SegmentConfig` has all 8 segments set to `false`
**When** the renderer renders in default mode
**Then** the output is `"MoAI"` (fallback string)

---

## 5. Mode Interaction

### AC-SL-040: Minimal mode ignores segment config

**Given** `MOAI_STATUSLINE_MODE` is set to `"minimal"`
**And** `SegmentConfig` has `model: false` (model disabled)
**When** the statusline is rendered
**Then** the minimal mode hard-coded behavior is used (model + context + optional git, regardless of SegmentConfig)

### AC-SL-041: Default mode respects segment config

**Given** `MOAI_STATUSLINE_MODE` is set to `"default"` (or unset)
**And** `SegmentConfig` has `claude_version: false` and `moai_version: false`
**When** the statusline is rendered
**Then** the output does NOT contain claude_version or moai_version segments

### AC-SL-042: Verbose mode respects segment config

**Given** `MOAI_STATUSLINE_MODE` is set to `"verbose"`
**And** `SegmentConfig` has `directory: false`
**When** the statusline is rendered
**Then** the output does NOT contain the directory segment

---

## 6. CLI Config Loading

### AC-SL-050: Config loaded from file

**Given** `.moai/config/sections/statusline.yaml` exists with `git_status: false`
**When** `moai statusline` executes
**Then** the loaded `SegmentConfig` map has `git_status` set to `false`

### AC-SL-051: Missing file returns nil config

**Given** `.moai/config/sections/statusline.yaml` does not exist
**When** `moai statusline` executes
**Then** `loadSegmentConfig()` returns nil and all segments are rendered

### AC-SL-052: Malformed YAML returns nil config

**Given** `.moai/config/sections/statusline.yaml` contains invalid YAML (`{{{invalid`)
**When** `moai statusline` executes
**Then** `loadSegmentConfig()` returns nil and all segments are rendered (graceful fallback)

---

## 7. Non-Functional Acceptance

### AC-SL-NF-001: Test coverage

**Given** all implementation is complete
**When** `go test -cover ./internal/statusline/... ./internal/cli/... ./internal/cli/wizard/...` is executed
**Then** coverage for modified packages is at least 85%

### AC-SL-NF-002: Static analysis

**Given** all implementation is complete
**When** `go vet ./...` and `golangci-lint run` are executed
**Then** both pass with zero errors

### AC-SL-NF-003: Race detection

**Given** all implementation is complete
**When** `go test -race ./...` is executed
**Then** no race conditions are detected

### AC-SL-NF-004: Build verification

**Given** all implementation is complete
**When** `make build` is executed
**Then** the build succeeds with embedded templates including the new `statusline.yaml`

### AC-SL-NF-005: Config load performance

**Given** a valid `statusline.yaml` file on disk
**When** `loadSegmentConfig()` is called
**Then** the function completes within 10ms

### AC-SL-NF-006: No wizard regressions

**Given** all implementation is complete
**When** existing wizard tests are executed
**Then** all existing tests pass without modification

### AC-SL-NF-007: No statusline regressions

**Given** all implementation is complete
**When** existing statusline renderer tests are executed
**Then** all existing tests pass without modification

---

## Definition of Done

- [ ] All acceptance criteria (AC-SL-*) verified and passing
- [ ] `statusline.yaml` template deployed during `moai init`
- [ ] Wizard presents preset + conditional segment questions
- [ ] Translations complete for ko, ja, zh
- [ ] `applyWizardConfig()` writes correct YAML for all presets
- [ ] Renderer filters segments based on loaded config
- [ ] Backward compatibility verified (missing config = all segments)
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run` passes
- [ ] `make build` succeeds
- [ ] Test coverage >= 85% for all modified packages
