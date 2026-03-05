# SPEC-SLE-001: Statusline Enhancement

## Metadata

| Field    | Value                                                                      |
| -------- | -------------------------------------------------------------------------- |
| SPEC ID  | SPEC-SLE-001                                                               |
| Title    | Statusline Enhancement: Themes, Multi-line, Profile Sync, API Monitoring   |
| Status   | Planned                                                                    |
| Priority | P1 (High)                                                                  |
| Created  | 2026-03-04                                                                 |
| Domain   | internal/statusline, internal/profile, internal/cli, internal/config       |
| Depends  | SPEC-STATUSLINE-001 (Implemented)                                          |

---

## 1. Environment

### Current System State

MoAI-ADK renders a statusline for Claude Code via the `moai statusline` command, invoked as a Claude Code hook. The statusline was enhanced in SPEC-STATUSLINE-001 to support segment configuration via wizard and YAML config. The current system has the following characteristics:

| Component               | State                                                                        |
| ----------------------- | ---------------------------------------------------------------------------- |
| `internal/statusline/`  | builder.go (235L), renderer.go (254L), types.go (187L)                       |
| Theme support           | `ThemeName` field exists in `Options` (builder.go:39) but renderer ignores it |
| Verbose mode            | `renderVerbose()` at renderer.go:174 just calls `renderCompact()` -- no differentiation |
| Cost rendering          | `CostUSD` is collected in `MetricsData` but never rendered in any mode       |
| Preset runtime loading  | `preset` field exists in statusline.yaml but is not read at runtime           |
| Config layer            | `StatuslineConfig` struct does NOT exist in the config layer                  |
| Profile sync            | `SyncToProjectConfig()` syncs user/language but NOT statusline settings      |
| Catppuccin dependency   | `catppuccin/go` v0.3.0 already in go.mod as indirect dependency (via glamour) |
| Testability patterns    | Interface-based: `GitDataProvider`, `UpdateProvider`                         |
| Caching pattern         | TTL-based caching in update.go (`UpdateChecker`)                             |
| Execution model         | Statusline runs as subprocess per hook trigger -- in-memory cache won't work  |

### Key Source Files

| File                                     | Purpose                                               |
| ---------------------------------------- | ----------------------------------------------------- |
| `internal/statusline/renderer.go`        | Renders segments via `renderCompact()`, `renderMinimal()`, `renderVerbose()` |
| `internal/statusline/builder.go`         | Orchestrates data collection; `Options` struct; `New()` constructor |
| `internal/statusline/types.go`           | `StatuslineMode`, `StatusData`, `Builder` interface, segment constants |
| `internal/cli/statusline.go`             | CLI entry; reads `MOAI_STATUSLINE_MODE` env var        |
| `internal/profile/preferences.go`        | `ProfilePreferences` struct for user preferences       |
| `internal/profile/sync.go`              | `SyncToProjectConfig()` for profile-to-project sync    |
| `internal/cli/profile_setup.go`          | Profile wizard setup                                   |
| `internal/cli/profile_setup_translations.go` | Multi-language translations for profile wizard     |

---

## 2. Assumptions

- **A1**: The `catppuccin/go` v0.3.0 indirect dependency (already in go.mod via glamour) can be promoted to a direct dependency without introducing conflicts.
- **A2**: The existing `ThemeName` field in `Options` (builder.go:39) can be repurposed to select the theme implementation without breaking existing callers (currently unused by renderer).
- **A3**: Multi-line verbose output (newline-separated) is compatible with Claude Code's statusline hook rendering. If Claude Code truncates at the first newline, verbose mode will be documented as experimental.
- **A4**: The `lipgloss.AdaptiveColor` function provides adequate fallback for terminals that do not support 256-color or truecolor, ensuring cross-platform compatibility.
- **A5**: File-based caching at `~/.moai/cache/usage.json` with TTL is the appropriate caching strategy since the statusline runs as a subprocess per hook trigger (in-memory cache would not persist).
- **A6**: The Anthropic API usage endpoint accepts `ANTHROPIC_OAUTH_TOKEN` for authentication without requiring keyring integration or CGO dependencies.
- **A7**: The `SyncToProjectConfig()` function in `sync.go` can be extended to sync statusline settings without breaking existing user/language sync behavior.
- **A8**: The profile wizard in `profile_setup.go` supports adding new "Display" category questions following the same pattern as existing questions.

---

## 3. Requirements

### 3.1 Config Layer Foundation (Phase 1)

**REQ-SLE-001** (Ubiquitous)
The system shall define a `StatuslineConfig` struct in `pkg/models/` with fields: `Preset` (string), `Segments` (map[string]bool), and `Theme` (string).

**REQ-SLE-002** (Ubiquitous)
The system shall register `StatuslineConfig` in the Config struct and `sectionNames` map in `internal/config/types.go` so that statusline configuration is loaded and managed alongside other config sections.

**REQ-SLE-003** (Event-Driven)
WHEN `statusline.yaml` is absent, THEN all segments shall default to enabled and the theme shall default to `"default"`.

### 3.2 Theme System (Phase 2)

**REQ-SLE-010** (Ubiquitous)
The system shall define a `Theme` interface in `internal/statusline/theme.go` with methods: `Primary()`, `Muted()`, `Success()`, `Warning()`, `Danger()`, `Text()`, and `BarGradient(percentage float64)`, each returning `lipgloss.Color`.

**REQ-SLE-011** (Ubiquitous)
The system shall support three theme implementations: `DefaultTheme`, `CatppuccinMocha`, and `CatppuccinLatte`.

**REQ-SLE-012** (Event-Driven)
WHEN the theme name is `"default"`, THEN the `DefaultTheme` implementation shall be used, preserving the current hard-coded color values.

**REQ-SLE-013** (Event-Driven)
WHEN the theme name is `"catppuccin-mocha"`, THEN the `CatppuccinMocha` implementation shall be used with Catppuccin Mocha palette colors:
- Primary: `#CDD6F4` (Text)
- Muted: `#6C7086` (Overlay0)
- Success: `#A6E3A1` (Green)
- Warning: `#F9E2AF` (Yellow)
- Danger: `#F38BA8` (Red)
- Text: `#CDD6F4` (Text)

**REQ-SLE-014** (Event-Driven)
WHEN the theme name is `"catppuccin-latte"`, THEN the `CatppuccinLatte` implementation shall be used with Catppuccin Latte palette colors:
- Primary: `#4C4F69` (Text)
- Muted: `#9CA0B0` (Overlay0)
- Success: `#40A02B` (Green)
- Warning: `#DF8E1D` (Yellow)
- Danger: `#D20F39` (Red)
- Text: `#4C4F69` (Text)

**REQ-SLE-015** (Event-Driven)
WHEN context usage percentage is 0-25%, THEN `BarGradient()` shall return Stage 1 color (green: Mocha `#A6E3A1`, Latte `#40A02B`).
WHEN context usage percentage is 26-50%, THEN `BarGradient()` shall return Stage 2 color (yellow: Mocha `#F9E2AF`, Latte `#DF8E1D`).
WHEN context usage percentage is 51-75%, THEN `BarGradient()` shall return Stage 3 color (peach: Mocha `#FAB387`, Latte `#FE640B`).
WHEN context usage percentage is 76-100%, THEN `BarGradient()` shall return Stage 4 color (red: Mocha `#F38BA8`, Latte `#D20F39`).

**REQ-SLE-016** (State-Driven)
IF `NoColor` is true, THEN the system shall ignore the selected theme and render without ANSI escape codes.

**REQ-SLE-017** (Event-Driven)
WHEN `NewRenderer()` is called, THEN the `themeName` parameter shall be used to select and instantiate the appropriate `Theme` implementation, and the theme's colors shall be wired into the renderer's lipgloss styles.

**REQ-SLE-018** (Ubiquitous)
The `DefaultTheme` `BarGradient()` shall use the same 4-stage color progression as Catppuccin Mocha but with the existing hard-coded color values (green/yellow/orange/red).

### 3.3 Profile Sync Fix (Phase 3)

**REQ-SLE-020** (Ubiquitous)
The `ProfilePreferences` struct in `preferences.go` shall include a `StatuslineTheme` field of type `string`.

**REQ-SLE-021** (Event-Driven)
WHEN `SyncToProjectConfig()` is called, THEN the system shall persist `StatuslinePreset`, `StatuslineSegments`, and `StatuslineTheme` to `.moai/config/sections/statusline.yaml`.

**REQ-SLE-022** (Event-Driven)
WHEN `statusline.yaml` is absent at the time of sync, THEN all segments shall default to enabled and `preset` shall default to `"full"`.

### 3.4 Multi-Line Verbose Mode (Phase 4)

**REQ-SLE-030** (Ubiquitous)
The system shall differentiate `ModeVerbose` from `ModeDefault` by rendering multi-line output.

**REQ-SLE-031** (Ubiquitous)
`ModeMinimal` output shall consist of 1 line: Model + Context Graph + Output Style.

**REQ-SLE-032** (Ubiquitous)
`ModeDefault` output shall consist of 1 line with all enabled segments (current behavior, backward compatible).

**REQ-SLE-033** (Event-Driven)
WHEN `ModeVerbose` is active, THEN the output shall consist of up to 3 newline-separated lines:
- Line 1: Model | Context Graph | Output Style
- Line 2: Directory | Branch | Git Changes
- Line 3: Claude Version | MoAI Version | Cost

**REQ-SLE-034** (Event-Driven)
WHEN all segments in a verbose line are unavailable or disabled, THEN that line shall be omitted entirely (no empty rows).

**REQ-SLE-035** (Event-Driven)
WHEN `ModeVerbose` is active, THEN the `CostUSD` field from `MetricsData` shall be rendered on Line 3 as a formatted cost string (e.g., `$0.42`).

### 3.5 Profile Wizard Enhancement (Phase 5)

**REQ-SLE-040** (Event-Driven)
WHEN the user runs the profile setup wizard, THEN a theme selector step shall be presented in the "Display" section with three options: `default`, `catppuccin-mocha`, `catppuccin-latte`.

**REQ-SLE-041** (Event-Driven)
WHEN the user selects a theme in the wizard, THEN the selection shall be saved to `preferences.yaml` and synced via `SyncToProjectConfig()`.

**REQ-SLE-042** (Event-Driven)
WHEN an existing theme value is present in `preferences.yaml`, THEN the wizard shall pre-fill the theme selector with the existing value.

**REQ-SLE-043** (Ubiquitous)
The wizard shall support translations for the theme selector question in all 4 supported languages: English (en), Korean (ko), Japanese (ja), Chinese (zh).

### 3.6 API Rate Limit Monitoring (Phase 6 -- Optional/Deferred)

**REQ-SLE-050** (Ubiquitous)
The system shall define a `RateLimitProvider` interface with method `FetchUsage(ctx context.Context) (*RateLimitData, error)`, matching the existing testability patterns (`GitDataProvider`, `UpdateProvider`).

**REQ-SLE-051** (Ubiquitous)
The `RateLimitData` struct shall contain fields for 5-hour and 7-day usage percentages.

**REQ-SLE-052** (Event-Driven)
WHEN `FetchUsage()` is called, THEN responses shall be cached in a file at `~/.moai/cache/usage.json` with a minimum TTL of 60 seconds.

**REQ-SLE-053** (State-Driven)
IF the `ANTHROPIC_OAUTH_TOKEN` environment variable is not set or the API call fails, THEN the rate limit segment shall be omitted silently without error messages.

**REQ-SLE-054** (Event-Driven)
WHEN usage data is available, THEN the 5-hour and 7-day usage percentages shall be rendered with color coding using the theme's `BarGradient()` method.

**REQ-SLE-055** (Unwanted)
The API fetch operation shall NOT block statusline rendering beyond 300ms. If the fetch exceeds this duration, the cached or empty result shall be used.

---

## 4. Non-Functional Requirements

**REQ-SLE-NF-001** (Ubiquitous)
All new code shall achieve at least 85% test coverage with table-driven tests following TDD (quality.yaml: `test_first_required: true`).

**REQ-SLE-NF-002** (Ubiquitous)
All new code shall pass `go vet`, `go test -race ./...`, and `golangci-lint run` without errors.

**REQ-SLE-NF-003** (Unwanted)
The system shall NOT introduce any regressions in existing statusline rendering behavior.

**REQ-SLE-NF-004** (Ubiquitous)
Theme switching shall NOT increase render time by more than 5ms compared to the current implementation.

**REQ-SLE-NF-005** (Unwanted)
The API rate limit fetch shall NOT block statusline render beyond 300ms.

**REQ-SLE-NF-006** (Ubiquitous)
The system shall work correctly on macOS, Linux, and Windows.

**REQ-SLE-NF-007** (Ubiquitous)
The system shall maintain backward compatibility with existing `statusline.yaml` configs (configs without `theme` field shall default to `"default"` theme).

**REQ-SLE-NF-008** (Ubiquitous)
The `make build` command shall succeed after all changes, regenerating embedded template files that include the updated `statusline.yaml`.

---

## 5. Technical Design Summary

### 5.1 Theme Architecture

```
Theme (interface)
  |
  +-- DefaultTheme       (preserves current hard-coded colors)
  +-- CatppuccinMocha    (dark theme, warm palette)
  +-- CatppuccinLatte    (light theme, cool palette)

Each implements:
  Primary() lipgloss.Color
  Muted() lipgloss.Color
  Success() lipgloss.Color
  Warning() lipgloss.Color
  Danger() lipgloss.Color
  Text() lipgloss.Color
  BarGradient(pct float64) lipgloss.Color  // 4-stage gradient
```

### 5.2 Multi-Line Layout

```
ModeMinimal (1 line):
  Model | Context Graph | Output Style

ModeDefault (1 line -- current behavior):
  Model | Context Graph | Style | Directory | Git Status | Claude Ver | MoAI Ver | Branch

ModeVerbose (up to 3 lines):
  Line 1: Model | Context Graph | Output Style
  Line 2: Directory | Branch | Git Changes
  Line 3: Claude Version | MoAI Version | Cost
```

### 5.3 Profile Sync Flow

```
ProfilePreferences.StatuslineTheme
  --> SyncToProjectConfig()
    --> .moai/config/sections/statusline.yaml
      --> statusline.theme: "catppuccin-mocha"
```

### 5.4 Rate Limit Cache Architecture (Phase 6)

```
moai statusline (subprocess)
  --> RateLimitProvider.FetchUsage(ctx)
    --> Check ~/.moai/cache/usage.json (TTL: 60s min)
      --> If fresh: return cached data
      --> If stale: HTTP GET to Anthropic API (300ms timeout)
        --> Update cache file
        --> Return data
      --> If error: return nil (segment omitted)
```

### 5.5 statusline.yaml Extended Format

```yaml
# .moai/config/sections/statusline.yaml
statusline:
  preset: "full"
  theme: "default"           # NEW: "default" | "catppuccin-mocha" | "catppuccin-latte"
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

---

## 6. File Change Matrix

### New Files

| File                                      | Purpose                                        | Phase |
| ----------------------------------------- | ---------------------------------------------- | ----- |
| `internal/statusline/theme.go`            | Theme interface + 3 implementations            | 2     |
| `internal/statusline/theme_test.go`       | Theme unit tests (gradient, colors)            | 2     |
| `internal/statusline/ratelimit.go`        | RateLimitProvider interface + implementation   | 6     |
| `internal/statusline/ratelimit_test.go`   | Rate limit provider tests                      | 6     |

### Modified Files

| File                                                        | Changes                                                     | Phase |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----- |
| `pkg/models/` or `internal/config/types.go`                 | Add `StatuslineConfig` struct, register in config            | 1     |
| `internal/statusline/renderer.go`                           | Wire theme colors into lipgloss styles, implement multi-line verbose, render cost | 2, 4  |
| `internal/statusline/types.go`                              | Add `ThemeData`, `RateLimitData` structs                     | 2, 6  |
| `internal/statusline/builder.go`                            | Pass theme to renderer, add `RateLimitProvider` field        | 2, 6  |
| `internal/profile/preferences.go`                           | Add `StatuslineTheme string` field                           | 3     |
| `internal/profile/sync.go`                                  | Add statusline sync block to `SyncToProjectConfig()`         | 3     |
| `internal/cli/profile_setup.go`                             | Add theme selector step in Display section                   | 5     |
| `internal/cli/profile_setup_translations.go`                | Add translations for theme selector (en, ko, ja, zh)        | 5     |
| `internal/cli/statusline.go`                                | Read theme from config, read preset, pass to builder         | 2, 4  |
| `internal/template/templates/.moai/config/sections/statusline.yaml` | Add `theme` field to template                       | 1     |

### Test Files (New or Modified)

| File                                          | Purpose                                          | Phase |
| --------------------------------------------- | ------------------------------------------------ | ----- |
| `internal/statusline/renderer_test.go`        | Test theme-colored rendering, multi-line verbose  | 2, 4  |
| `internal/statusline/builder_test.go`         | Test theme passthrough, RateLimitProvider wiring  | 2, 6  |
| `internal/cli/statusline_test.go`             | Test theme config loading                         | 2     |
| `internal/profile/preferences_test.go`        | Test StatuslineTheme field                        | 3     |
| `internal/profile/sync_test.go`               | Test statusline sync                              | 3     |
| `internal/cli/profile_setup_test.go`          | Test theme selector wizard step                   | 5     |

---

## 7. Implementation Order and Dependencies

### Dependency Graph

```
Phase 1 (Config Layer)
  |
  v
Phase 2 (Theme System) -----> Phase 3 (Profile Sync)
  |                                |
  v                                v
Phase 4 (Multi-line Verbose)   Phase 5 (Wizard Enhancement)
                                   |
                               Phase 6 (API Rate Limit -- Optional)
```

### Phase Breakdown

| Phase | Name                     | Priority     | Dependencies | Scope                          |
| ----- | ------------------------ | ------------ | ------------ | ------------------------------ |
| 1     | Config Layer Foundation  | Primary Goal | None         | Config struct, YAML integration |
| 2     | Theme System             | Primary Goal | Phase 1      | Theme interface, 3 implementations, renderer wiring |
| 3     | Profile Sync Fix         | Secondary Goal | Phase 1    | Preferences field, sync function |
| 4     | Multi-Line Verbose Mode  | Secondary Goal | Phase 2    | Verbose renderer, cost display  |
| 5     | Wizard Enhancement       | Final Goal   | Phase 3      | Theme selector, translations    |
| 6     | API Rate Limit           | Optional Goal | Phase 2     | Provider interface, caching, UI |

### Implementation Ordering Notes

- Phases 1 and 2 must be completed first as they provide the foundation for all subsequent work.
- Phases 3 and 4 can be developed in parallel after Phase 2 is complete, as they modify different files.
- Phase 5 depends on Phase 3 (ProfilePreferences must have the StatuslineTheme field before the wizard can set it).
- Phase 6 is explicitly deferred and optional. It introduces external API dependency and should be developed only after all other phases are stable.

---

## 8. Risk Register

| Risk ID | Risk                                       | Likelihood | Impact | Mitigation                                          |
| ------- | ------------------------------------------ | ---------- | ------ | --------------------------------------------------- |
| R1      | Multi-line output truncated by Claude Code | Medium     | High   | Document verbose mode as experimental; test with Claude Code hook; graceful fallback to single-line |
| R2      | Anthropic API auth endpoint changes        | Low        | Medium | Phase 6 is deferred and opt-in; provider interface allows swapping implementation |
| R3      | 256-color terminal compatibility issues    | Medium     | Low    | Use `lipgloss.AdaptiveColor` for all theme colors; test on common terminal emulators |
| R4      | statusline.yaml migration from existing configs | Low   | Medium | Default all fields to backward-compatible values; missing `theme` field defaults to `"default"` |
| R5      | catppuccin/go dependency version conflicts | Low        | Low    | Already an indirect dependency via glamour; promote to direct with same version |
| R6      | Render performance regression with theme switching | Low | Medium | Benchmark before/after; theme selection is a static dispatch (interface method call) |
| R7      | File-based cache race conditions (concurrent statusline calls) | Medium | Low | Use atomic write (write to temp file, then rename) for cache file operations |

---

## 9. Traceability

| Requirement          | Implementation File(s)                                              | Test File(s)                              |
| -------------------- | ------------------------------------------------------------------- | ----------------------------------------- |
| REQ-SLE-001-003      | `internal/config/types.go`, template `statusline.yaml`              | `internal/config/types_test.go`           |
| REQ-SLE-010-018      | `internal/statusline/theme.go`, `renderer.go`                       | `theme_test.go`, `renderer_test.go`       |
| REQ-SLE-020-022      | `internal/profile/preferences.go`, `sync.go`                        | `preferences_test.go`, `sync_test.go`     |
| REQ-SLE-030-035      | `internal/statusline/renderer.go`                                   | `renderer_test.go`                        |
| REQ-SLE-040-043      | `internal/cli/profile_setup.go`, `profile_setup_translations.go`    | `profile_setup_test.go`                   |
| REQ-SLE-050-055      | `internal/statusline/ratelimit.go`, `builder.go`                    | `ratelimit_test.go`, `builder_test.go`    |
| REQ-SLE-NF-001-008   | All implementation files                                            | All test files                            |
