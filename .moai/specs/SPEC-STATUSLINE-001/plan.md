# SPEC-STATUSLINE-001: Implementation Plan

## SPEC Reference

| Field   | Value               |
| ------- | ------------------- |
| SPEC ID | SPEC-STATUSLINE-001 |
| Title   | Statusline Segment Configuration |

---

## Milestone 1: Configuration Foundation (Primary Goal)

### Objective

Establish the configuration file format, template, and loading mechanism.

### Tasks

1. **Create statusline.yaml template**
   - File: `internal/template/templates/.moai/config/sections/statusline.yaml`
   - Content: Default config with preset=full and all segments=true
   - Verify: `make build` succeeds with new embedded template

2. **Implement config loader in CLI**
   - File: `internal/cli/statusline.go`
   - Add `loadSegmentConfig(projectRoot string) map[string]bool` function
   - Parse YAML, return nil on any error (backward-compatible fallback)
   - Wire into `runStatusline()`: load config, pass to builder

3. **Extend Builder Options**
   - File: `internal/statusline/builder.go`
   - Add `SegmentConfig map[string]bool` field to `Options` struct
   - Pass `SegmentConfig` from `Options` to `Renderer` in `New()`

4. **Write tests for config loading**
   - Table-driven tests: valid config, missing file, malformed YAML, empty segments map
   - Verify backward compatibility: nil config = all segments enabled

### Requirements Covered

REQ-SL-001, REQ-SL-003, REQ-SL-040, REQ-SL-041, REQ-SL-050, REQ-SL-051

---

## Milestone 2: Renderer Segment Filtering (Primary Goal)

### Objective

Modify the renderer to respect segment configuration when rendering the statusline.

### Tasks

1. **Add segment filtering to Renderer**
   - File: `internal/statusline/renderer.go`
   - Add `segmentConfig map[string]bool` field to `Renderer` struct
   - Add `isSegmentEnabled(key string) bool` method
   - Accept `segmentConfig` parameter in `NewRenderer()` (or add setter)

2. **Wrap each segment block in renderCompact()**
   - Guard each of the 8 segment blocks with `r.isSegmentEnabled("key")`
   - Preserve existing rendering logic within each guard
   - Apply same guards to `renderVerbose()` (which delegates to `renderCompact()`)

3. **Handle mode interaction**
   - `renderMinimal()` remains unchanged (hard-coded behavior)
   - Segment config applies only to `renderCompact()` and `renderVerbose()`

4. **Write renderer tests**
   - Table-driven tests for each segment disabled individually
   - Test all-disabled config (should return "MoAI" fallback)
   - Test nil config (backward compatibility: all segments shown)
   - Test partial config (missing keys default to enabled)

### Requirements Covered

REQ-SL-030, REQ-SL-031, REQ-SL-032, REQ-SL-033, REQ-SL-042

---

## Milestone 3: Wizard Integration (Secondary Goal)

### Objective

Add statusline configuration questions to the BubbleTea init/update wizard.

### Tasks

1. **Extend WizardResult**
   - File: `internal/cli/wizard/types.go`
   - Add `StatuslinePreset string` field
   - Add `StatuslineSegments map[string]bool` field

2. **Add wizard questions**
   - File: `internal/cli/wizard/questions.go`
   - Add `statusline_preset` select question (Full/Compact/Minimal/Custom)
   - Add 8 conditional `statusline_seg_{name}` select questions (Enabled/Disabled)
   - Condition on all 8 segment questions: `r.StatuslinePreset == "custom"`

3. **Add translations**
   - File: `internal/cli/wizard/translations.go`
   - Add Korean (ko), Japanese (ja), Chinese (zh) translations for:
     - `statusline_preset` question title, description, and options
     - All 8 `statusline_seg_{name}` question titles, descriptions, and options

4. **Wire applyWizardConfig()**
   - File: `internal/cli/update.go`
   - Extend `applyWizardConfig()` to handle `StatuslinePreset`
   - Compute segment map from preset definitions (full/compact/minimal)
   - For custom: use `StatuslineSegments` directly from WizardResult
   - Write result to `.moai/config/sections/statusline.yaml`

5. **Write wizard tests**
   - Test question filtering: segment questions hidden unless custom
   - Test applyWizardConfig: verify YAML output for each preset
   - Test applyWizardConfig: verify YAML output for custom with mixed toggles

### Requirements Covered

REQ-SL-002, REQ-SL-004, REQ-SL-005, REQ-SL-006, REQ-SL-007, REQ-SL-010, REQ-SL-011, REQ-SL-012, REQ-SL-013, REQ-SL-014, REQ-SL-020, REQ-SL-021, REQ-SL-022

---

## Milestone 4: Integration Testing and Polish (Final Goal)

### Objective

End-to-end validation, backward compatibility verification, and documentation.

### Tasks

1. **Integration tests**
   - Test full pipeline: wizard result -> applyWizardConfig -> statusline.yaml -> load config -> render
   - Test upgrade scenario: project without statusline.yaml renders all segments
   - Test 3-way merge: verify statusline.yaml is preserved during `moai update`

2. **Backward compatibility verification**
   - Run existing statusline tests, confirm no regressions
   - Run existing wizard tests, confirm no regressions
   - Verify `MOAI_STATUSLINE_MODE=minimal` still works independently

3. **Build verification**
   - `make build` succeeds (embedded templates regenerated)
   - `go test -race ./...` passes
   - `golangci-lint run` passes
   - Coverage >= 85% for all modified packages

### Requirements Covered

REQ-SL-NF-001 through REQ-SL-NF-007

---

## Technical Approach

### Architecture Design

The segment configuration feature follows a layered approach:

1. **Config Layer**: New `statusline.yaml` file with preset + segment toggles
2. **Wizard Layer**: New questions in BubbleTea wizard for user configuration
3. **Application Layer**: `applyWizardConfig()` writes config based on wizard results
4. **Loading Layer**: CLI reads config at statusline render time
5. **Rendering Layer**: Renderer filters segments based on loaded config

### Key Design Decisions

- **Config file over environment variable**: YAML config is persistent, mergeable, and part of the wizard flow. The env var `MOAI_STATUSLINE_MODE` is preserved for quick overrides.
- **Preset + custom hybrid**: Presets cover 90% of use cases with zero friction. Custom mode provides full control for power users.
- **Segment questions as individual yes/no**: Since the BubbleTea wizard does not support multi-select, 8 individual questions with conditional visibility provide the equivalent UX.
- **Nil config = all enabled**: Ensures zero-config backward compatibility for existing projects.
- **Mode takes precedence for minimal**: `MOAI_STATUSLINE_MODE=minimal` retains its hard-coded behavior since it is designed for constrained display contexts.

### Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Config loading adds latency to statusline rendering | Statusline feels sluggish | Config file is small (<500 bytes); parse is sub-millisecond. REQ-SL-NF-005 enforces 10ms budget. |
| 8 additional wizard questions overwhelm users | Poor wizard UX | Conditional visibility: segment questions only appear for "custom" preset. Most users pick a preset and never see them. |
| Existing statusline behavior changes unexpectedly | User confusion, bug reports | Nil config fallback guarantees identical behavior for projects without statusline.yaml. |
| 3-way merge fails for statusline.yaml during update | User config lost | statusline.yaml follows the same YAML structure as other section files; existing merge logic handles it. |
