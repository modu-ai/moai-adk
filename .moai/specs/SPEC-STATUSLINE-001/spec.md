# SPEC-STATUSLINE-001: Statusline Segment Configuration

## Metadata

| Field    | Value                                                              |
| -------- | ------------------------------------------------------------------ |
| SPEC ID  | SPEC-STATUSLINE-001                                                |
| Title    | Statusline Segment Configuration via Wizard and YAML Config        |
| Status   | Implemented                                                        |
| Priority | P1 (High)                                                          |
| Created  | 2026-02-14                                                         |
| Domain   | internal/statusline, internal/cli/wizard, internal/cli, internal/template |

---

## 1. Environment

### Current System State

MoAI-ADK renders a statusline for Claude Code via the `moai statusline` command, invoked as a Claude Code hook. The statusline displays up to 8 segments in a fixed order:

| # | Segment Key      | Format                                    | Example Output       |
|---|------------------|-------------------------------------------|----------------------|
| 1 | model            | `model-name`                              | Opus 4.6             |
| 2 | context          | `bar-graph percentage%`                   | 41%                  |
| 3 | output_style     | `style-name`                              | MoAI                 |
| 4 | directory        | `directory-name`                          | moai-adk-go          |
| 5 | git_status       | `+staged Mmodified ?untracked`            | +0 M3 ?1             |
| 6 | claude_version   | `vversion`                                | v1.0.80              |
| 7 | moai_version     | `vversion`                                | v2.3.1               |
| 8 | git_branch       | `branch-name`                             | feat/statusline      |

### Current Mode System

The `MOAI_STATUSLINE_MODE` environment variable controls display:

- **minimal**: model + context graph (+ git_status if fits in 40 chars)
- **default**: All 8 segments
- **verbose**: Same as default (alias)

### Limitations

1. **No user choice over individual segments**: Users cannot select which specific segments to display. They are limited to three hard-coded presets.
2. **No wizard integration**: Statusline configuration is not part of the `moai init` or `moai update -c` wizard flow.
3. **No persistent configuration file**: There is no `statusline.yaml` config file. The mode is determined solely from an environment variable.
4. **No 3-way merge support**: Without a config file in `internal/template/templates/.moai/config/sections/`, the statusline config is not preserved during `moai update`.

### Key Source Files

| File | Purpose |
|------|---------|
| `internal/statusline/renderer.go` | Renders segments via `renderCompact()`, `renderMinimal()`, `renderVerbose()` |
| `internal/statusline/builder.go` | Orchestrates data collection; `Options` struct; `New()` constructor |
| `internal/statusline/types.go` | `StatuslineMode`, `StatusData`, `Builder` interface |
| `internal/cli/statusline.go` | CLI entry; reads `MOAI_STATUSLINE_MODE` env var |
| `internal/cli/wizard/questions.go` | `DefaultQuestions()` with 13 questions |
| `internal/cli/wizard/types.go` | `WizardResult` struct (21 fields), `Question`/`Option` types |
| `internal/cli/wizard/translations.go` | Multi-language translations (ko, ja, zh) |
| `internal/cli/update.go` | `applyWizardConfig()` writes wizard results to YAML config files |
| `internal/template/templates/.moai/config/sections/` | Template source for config YAML files |

---

## 2. Assumptions

- **A1**: The BubbleTea wizard system supports adding new `QuestionTypeSelect` questions without structural changes to the wizard framework.
- **A2**: The `applyWizardConfig()` function in `update.go` can be extended to write a new `statusline.yaml` config file following the same pattern used for `workflow.yaml` and `user.yaml`.
- **A3**: A new `statusline.yaml` template file in `internal/template/templates/.moai/config/sections/` will be automatically picked up by the embedded template system and deployed during `moai init` and `moai update`.
- **A4**: The 3-way YAML merge system (`mergeYAML3Way` in `update.go`) will work correctly for `statusline.yaml` without modification, since it follows the same structure as other section files.
- **A5**: The `WizardResult` struct can be extended with new fields without breaking backward compatibility with existing wizard logic.
- **A6**: For the "custom" preset flow, a series of individual yes/no select questions (one per segment) is preferred over a multi-select question type, since the BubbleTea wizard only supports `QuestionTypeSelect` and `QuestionTypeInput`.
- **A7**: When `statusline.yaml` does not exist (pre-upgrade projects), the system defaults to all segments enabled, preserving backward compatibility with existing behavior.

---

## 3. Requirements

### 3.1 Configuration Storage

**REQ-SL-001** (Ubiquitous)
The system shall store statusline configuration in `.moai/config/sections/statusline.yaml` with a `statusline` root key containing `preset` (string) and `segments` (map of segment names to booleans).

**REQ-SL-002** (Ubiquitous)
The system shall recognize the following preset values: `full`, `compact`, `minimal`, `custom`.

**REQ-SL-003** (Ubiquitous)
The system shall recognize the following 8 segment keys in the `segments` map: `model`, `context`, `output_style`, `directory`, `git_status`, `claude_version`, `moai_version`, `git_branch`.

**REQ-SL-004** (Event-Driven)
When the preset is `full`, the system shall enable all 8 segments.

**REQ-SL-005** (Event-Driven)
When the preset is `compact`, the system shall enable only: `model`, `context`, `git_status`, `git_branch`.

**REQ-SL-006** (Event-Driven)
When the preset is `minimal`, the system shall enable only: `model`, `context`.

**REQ-SL-007** (Event-Driven)
When the preset is `custom`, the system shall use the individual `segments` map values to determine which segments are enabled.

### 3.2 Wizard Integration

**REQ-SL-010** (Event-Driven)
When the user runs `moai init` or `moai update -c`, the wizard shall present a "Statusline preset" question with options: Full (default), Compact, Minimal, Custom.

**REQ-SL-011** (Event-Driven)
When the user selects "Custom" as the statusline preset, the wizard shall present 8 individual yes/no questions, one for each segment, allowing the user to enable or disable each segment.

**REQ-SL-012** (Event-Driven)
When the user selects a preset other than "Custom", the wizard shall skip the individual segment questions.

**REQ-SL-013** (Ubiquitous)
The wizard shall support translations for the statusline questions in all 4 supported languages: English (en), Korean (ko), Japanese (ja), Chinese (zh).

**REQ-SL-014** (Ubiquitous)
The `WizardResult` struct shall include fields `StatuslinePreset` (string) and `StatuslineSegments` (map[string]bool) to store the user's statusline configuration choices.

### 3.3 Configuration Application

**REQ-SL-020** (Event-Driven)
When `applyWizardConfig()` is called with a `WizardResult` containing a non-empty `StatuslinePreset`, the system shall write the statusline configuration to `.moai/config/sections/statusline.yaml`.

**REQ-SL-021** (Event-Driven)
When a preset other than "custom" is selected, `applyWizardConfig()` shall compute the `segments` map from the preset definition and write both `preset` and `segments` to the YAML file.

**REQ-SL-022** (Event-Driven)
When the "custom" preset is selected, `applyWizardConfig()` shall write the user-selected `segments` map from `WizardResult.StatuslineSegments` to the YAML file.

### 3.4 Renderer Integration

**REQ-SL-030** (Event-Driven)
When the statusline is rendered, the system shall read segment configuration from `.moai/config/sections/statusline.yaml` and filter out disabled segments before rendering.

**REQ-SL-031** (Ubiquitous)
The `Options` struct in `builder.go` shall accept a `SegmentConfig` field (map[string]bool) that specifies which segments are enabled.

**REQ-SL-032** (Event-Driven)
When `SegmentConfig` is provided to the builder, the renderer shall skip any segment whose key maps to `false` in the config.

**REQ-SL-033** (Event-Driven)
When `SegmentConfig` is nil or empty (no config loaded), the renderer shall display all segments (backward-compatible default behavior).

### 3.5 CLI Integration

**REQ-SL-040** (Event-Driven)
When the `moai statusline` command executes, the CLI shall attempt to load `.moai/config/sections/statusline.yaml` and pass the segment configuration to the builder.

**REQ-SL-041** (Event-Driven)
When `statusline.yaml` does not exist or cannot be parsed, the CLI shall fall back to all segments enabled (backward compatibility).

**REQ-SL-042** (State-Driven)
While the `MOAI_STATUSLINE_MODE` environment variable is set to `minimal`, the mode-based filtering shall take precedence over the segment config. The segment config shall apply only when the mode is `default` or `verbose`.

### 3.6 Template

**REQ-SL-050** (Ubiquitous)
The system shall provide a default `statusline.yaml` template at `internal/template/templates/.moai/config/sections/statusline.yaml` with preset set to `full` and all 8 segments set to `true`.

**REQ-SL-051** (Ubiquitous)
The statusline template shall be compatible with the existing 3-way YAML merge system used during `moai update`.

---

## 4. Non-Functional Requirements

**REQ-SL-NF-001** (Ubiquitous)
All new code shall achieve at least 85% test coverage with table-driven tests.

**REQ-SL-NF-002** (Ubiquitous)
All new code shall pass `go vet`, `go test -race`, and `golangci-lint run` without errors.

**REQ-SL-NF-003** (Unwanted)
The system shall not introduce any regressions in existing statusline rendering behavior.

**REQ-SL-NF-004** (Unwanted)
The system shall not introduce any regressions in existing wizard question ordering or behavior.

**REQ-SL-NF-005** (Ubiquitous)
Reading and parsing `statusline.yaml` shall complete within 10ms to avoid impacting statusline rendering latency (statusline renders on every Claude Code update).

**REQ-SL-NF-006** (Ubiquitous)
The `make build` command shall succeed after all changes, regenerating embedded template files that include the new `statusline.yaml`.

**REQ-SL-NF-007** (Ubiquitous)
The system shall maintain backward compatibility: projects without `statusline.yaml` shall render all segments as before.

---

## 5. Specifications

### 5.1 Configuration File Format

```yaml
# .moai/config/sections/statusline.yaml
statusline:
  # Preset name for reference (full, compact, minimal, custom)
  preset: "full"

  # Individual segment toggles
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

### 5.2 Preset Definitions

| Preset  | model | context | output_style | directory | git_status | claude_version | moai_version | git_branch |
|---------|-------|---------|--------------|-----------|------------|----------------|--------------|------------|
| full    | true  | true    | true         | true      | true       | true           | true         | true       |
| compact | true  | true    | false        | false     | true       | false          | false        | true       |
| minimal | true  | true    | false        | false     | false      | false          | false        | false      |
| custom  | (user-defined per segment)                                                                           |

### 5.3 WizardResult Extension

```go
// In internal/cli/wizard/types.go
type WizardResult struct {
    // ... existing fields ...

    // Statusline settings
    StatuslinePreset  string          // Statusline preset: full, compact, minimal, custom
    StatuslineSegments map[string]bool // Per-segment toggles (only for custom preset)
}
```

### 5.4 Wizard Questions

New questions to add at the end of `DefaultQuestions()` in `questions.go`:

**Question: statusline_preset** (Select)
- Options: Full (default), Compact, Minimal, Custom
- Always visible (no condition)

**Questions: statusline_seg_{name}** (Select, 8 questions, conditional)
- One per segment: `statusline_seg_model`, `statusline_seg_context`, etc.
- Options: Enabled (default), Disabled
- Condition: `WizardResult.StatuslinePreset == "custom"`

### 5.5 Builder Options Extension

```go
// In internal/statusline/builder.go
type Options struct {
    // ... existing fields ...

    // SegmentConfig maps segment keys to enabled state.
    // When nil or empty, all segments are displayed (backward compatible).
    SegmentConfig map[string]bool
}
```

### 5.6 Renderer Segment Filtering

The `renderCompact()` method shall check `SegmentConfig` before appending each segment. The filtering logic:

```go
func (r *Renderer) isSegmentEnabled(key string) bool {
    if r.segmentConfig == nil || len(r.segmentConfig) == 0 {
        return true // backward compatible: show all when no config
    }
    enabled, exists := r.segmentConfig[key]
    if !exists {
        return true // unknown segments default to enabled
    }
    return enabled
}
```

Each segment block in `renderCompact()` shall be wrapped:

```go
// 1. Model with emoji
if r.isSegmentEnabled("model") && data.Metrics.Available && data.Metrics.Model != "" {
    sections = append(sections, fmt.Sprintf("model-name %s", data.Metrics.Model))
}
```

### 5.7 CLI Config Loading

```go
// In internal/cli/statusline.go
func loadSegmentConfig(projectRoot string) map[string]bool {
    configPath := filepath.Join(projectRoot, ".moai", "config", "sections", "statusline.yaml")

    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil // file missing = all segments enabled
    }

    var config struct {
        Statusline struct {
            Segments map[string]bool `yaml:"segments"`
        } `yaml:"statusline"`
    }

    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil // parse error = all segments enabled
    }

    return config.Statusline.Segments
}
```

### 5.8 Mode vs. Config Interaction

| Mode    | Segment Config | Behavior |
|---------|---------------|----------|
| minimal | (ignored)     | Hard-coded minimal rendering (model + context + optional git) |
| default | Applied       | Config-driven segment filtering via `SegmentConfig` |
| verbose | Applied       | Config-driven segment filtering via `SegmentConfig` |
| (empty) | Applied       | Same as default: config-driven segment filtering |

The `MOAI_STATUSLINE_MODE=minimal` mode retains its existing hard-coded behavior for maximum compactness. The segment config applies to `default` and `verbose` modes.

---

## 6. Affected Files

### New Files

| File | Purpose |
|------|---------|
| `internal/template/templates/.moai/config/sections/statusline.yaml` | Default template with all segments enabled |

### Modified Files

| File | Changes |
|------|---------|
| `internal/statusline/renderer.go` | Add `segmentConfig` field, `isSegmentEnabled()` method, wrap segment blocks |
| `internal/statusline/builder.go` | Add `SegmentConfig` field to `Options`, pass to `Renderer` |
| `internal/statusline/types.go` | (Minimal: possibly add segment key constants) |
| `internal/cli/statusline.go` | Load `statusline.yaml`, pass `SegmentConfig` to builder `Options` |
| `internal/cli/wizard/questions.go` | Add statusline_preset + 8 conditional segment questions |
| `internal/cli/wizard/types.go` | Add `StatuslinePreset` and `StatuslineSegments` fields to `WizardResult` |
| `internal/cli/wizard/translations.go` | Add ko, ja, zh translations for new questions |
| `internal/cli/update.go` | Extend `applyWizardConfig()` to write `statusline.yaml` |

### Test Files (New or Modified)

| File | Purpose |
|------|---------|
| `internal/statusline/renderer_test.go` | Test segment filtering with various configs |
| `internal/statusline/builder_test.go` | Test SegmentConfig passthrough |
| `internal/cli/statusline_test.go` | Test config loading and fallback |
| `internal/cli/wizard/questions_test.go` | Test new questions and conditional visibility |
| `internal/cli/update_test.go` | Test applyWizardConfig for statusline.yaml |

---

## 7. Traceability

| Requirement    | Implementation File(s)                                       | Test File(s)                           |
| -------------- | ------------------------------------------------------------ | -------------------------------------- |
| REQ-SL-001-007 | `statusline.yaml` template, `update.go`                     | `update_test.go`                       |
| REQ-SL-010-014 | `wizard/questions.go`, `wizard/types.go`, `wizard/translations.go` | `wizard/questions_test.go`       |
| REQ-SL-020-022 | `update.go` (`applyWizardConfig`)                            | `update_test.go`                       |
| REQ-SL-030-033 | `renderer.go`, `builder.go`                                  | `renderer_test.go`, `builder_test.go`  |
| REQ-SL-040-042 | `cli/statusline.go`                                          | `cli/statusline_test.go`               |
| REQ-SL-050-051 | `templates/.moai/config/sections/statusline.yaml`            | (template deployment tests)            |
| REQ-SL-NF-001-007 | All implementation files                                 | All test files                         |
