# SPEC-SLE-001: Implementation Plan

## Metadata

| Field    | Value              |
| -------- | ------------------ |
| SPEC ID  | SPEC-SLE-001       |
| Title    | Statusline Enhancement |
| Status   | Planned            |

---

## 1. Milestones

### Primary Goal: Config Layer + Theme System (Phases 1-2)

Establish the config foundation and implement the theme system. This is the highest-value work that unblocks all subsequent phases.

**Phase 1: Config Layer Foundation**

Tasks:
1. Define `StatuslineConfig` struct in `pkg/models/` with fields: `Preset`, `Segments`, `Theme`
2. Register `StatuslineConfig` in the `Config` struct and `sectionNames` in `internal/config/types.go`
3. Update the statusline.yaml template to include the `theme` field
4. Run `make build` to regenerate embedded files
5. Write unit tests for config loading with and without `theme` field

Files Impacted:
- `pkg/models/` or `internal/config/types.go` (new struct + registration)
- `internal/template/templates/.moai/config/sections/statusline.yaml` (add `theme` field)
- `internal/config/types_test.go` (new tests)

**Phase 2: Theme System**

Tasks:
1. RED: Write tests for `Theme` interface and all 3 implementations
2. GREEN: Create `internal/statusline/theme.go` with:
   - `Theme` interface (7 methods)
   - `DefaultTheme` struct (preserves current colors)
   - `CatppuccinMocha` struct (Mocha palette)
   - `CatppuccinLatte` struct (Latte palette)
   - `NewTheme(name string) Theme` factory function
3. RED: Write tests for 4-stage gradient bar in each theme
4. GREEN: Implement `BarGradient(pct float64)` for all themes
5. REFACTOR: Wire theme into `NewRenderer()` -- replace hard-coded `mutedStyle` with theme colors
6. RED: Write tests for themed `buildBar()` and `renderContextGraph()`
7. GREEN: Update `buildBar()` to use theme colors via `BarGradient()`
8. Update `internal/cli/statusline.go` to read theme from config and pass to builder
9. Integration test: verify end-to-end theme selection

Files Impacted:
- `internal/statusline/theme.go` (NEW)
- `internal/statusline/theme_test.go` (NEW)
- `internal/statusline/renderer.go` (wire theme colors)
- `internal/statusline/renderer_test.go` (update tests)
- `internal/statusline/builder.go` (pass theme)
- `internal/cli/statusline.go` (read theme config)

### Secondary Goal: Profile Sync + Multi-Line Verbose (Phases 3-4)

These phases can be developed in parallel as they modify different files.

**Phase 3: Profile Sync Fix**

Tasks:
1. RED: Write test for `StatuslineTheme` field in `ProfilePreferences`
2. GREEN: Add `StatuslineTheme string` to `ProfilePreferences` struct
3. RED: Write test for statusline sync in `SyncToProjectConfig()`
4. GREEN: Add statusline sync block to `SyncToProjectConfig()`:
   - Sync `preset`, `segments`, and `theme` to `statusline.yaml`
5. REFACTOR: Handle edge cases (missing file, partial config)

Files Impacted:
- `internal/profile/preferences.go` (add field)
- `internal/profile/preferences_test.go` (new tests)
- `internal/profile/sync.go` (add sync block)
- `internal/profile/sync_test.go` (new tests)

**Phase 4: Multi-Line Verbose Mode**

Tasks:
1. RED: Write test for `renderVerbose()` returning multi-line output
2. GREEN: Implement `renderVerbose()` with 3-line layout:
   - Line 1: Model | Context Graph | Output Style
   - Line 2: Directory | Branch | Git Changes
   - Line 3: Claude Version | MoAI Version | Cost
3. RED: Write test for empty-line omission
4. GREEN: Implement line omission when all segments in a row are unavailable
5. RED: Write test for cost rendering format
6. GREEN: Implement cost rendering (e.g., `$0.42`) using `MetricsData.CostUSD`
7. REFACTOR: Ensure `ModeDefault` behavior is unchanged (backward compatible)

Files Impacted:
- `internal/statusline/renderer.go` (rewrite `renderVerbose()`, add cost rendering)
- `internal/statusline/renderer_test.go` (new verbose + cost tests)

### Final Goal: Wizard Enhancement (Phase 5)

**Phase 5: Profile Wizard Enhancement**

Tasks:
1. RED: Write test for theme selector question in profile setup
2. GREEN: Add theme selector step to `profile_setup.go` Display section:
   - 3 options: `default`, `catppuccin-mocha`, `catppuccin-latte`
3. RED: Write test for pre-fill with existing theme value
4. GREEN: Implement pre-fill logic
5. Add translations for all 4 languages in `profile_setup_translations.go`:
   - en: "Statusline Theme" / "Select a color theme for the statusline"
   - ko: "Statusline 테마" / "상태 표시줄의 색상 테마를 선택하세요"
   - ja: "Statusline テーマ" / "ステータスラインのカラーテーマを選択してください"
   - zh: "Statusline 主题" / "请选择状态栏的颜色主题"
6. Wire selection to `preferences.yaml` save and `SyncToProjectConfig()` call

Files Impacted:
- `internal/cli/profile_setup.go` (add question)
- `internal/cli/profile_setup_translations.go` (add translations)
- `internal/cli/profile_setup_test.go` (new tests)

### Optional Goal: API Rate Limit Monitoring (Phase 6 -- Deferred)

**Phase 6: API Rate Limit Monitoring**

Tasks:
1. RED: Write test for `RateLimitProvider` interface
2. GREEN: Define `RateLimitProvider` interface and `RateLimitData` struct
3. RED: Write test for file-based cache with TTL
4. GREEN: Implement cache at `~/.moai/cache/usage.json`:
   - Atomic write (temp file + rename) for safety
   - 60-second minimum TTL
5. RED: Write test for HTTP client with 300ms timeout
6. GREEN: Implement Anthropic API client:
   - Read `ANTHROPIC_OAUTH_TOKEN` from env
   - 300ms timeout for HTTP call
   - Parse 5h and 7d usage percentages
7. RED: Write test for graceful degradation
8. GREEN: Implement segment omission when no token or API error
9. Wire into builder and renderer with color-coded display
10. Integration test: verify segment appears/disappears based on env

Files Impacted:
- `internal/statusline/ratelimit.go` (NEW)
- `internal/statusline/ratelimit_test.go` (NEW)
- `internal/statusline/types.go` (add `RateLimitData`)
- `internal/statusline/builder.go` (add `RateLimitProvider`)
- `internal/statusline/builder_test.go` (update tests)

---

## 2. Technical Approach

### Theme System Design

The theme system uses a Go interface with static dispatch for zero-overhead theme switching:

```go
type Theme interface {
    Primary() lipgloss.Color
    Muted() lipgloss.Color
    Success() lipgloss.Color
    Warning() lipgloss.Color
    Danger() lipgloss.Color
    Text() lipgloss.Color
    BarGradient(pct float64) lipgloss.Color
}
```

The factory function `NewTheme(name string)` returns the appropriate implementation. The theme is injected into the `Renderer` at construction time and used for all lipgloss style creation. This replaces the current hard-coded `mutedStyle` with `theme.Muted()`.

### Multi-Line Verbose Strategy

The current `renderVerbose()` is a pass-through to `renderCompact()`. The new implementation will:

1. Group segments into 3 logical rows
2. Build each row independently using the same `isSegmentEnabled()` check
3. Filter out empty rows
4. Join rows with `\n`

This maintains backward compatibility because `ModeDefault` continues to call `renderCompact()`.

### Profile Sync Strategy

Extend the existing `SyncToProjectConfig()` pattern:

1. Add `StatuslineTheme` to `ProfilePreferences` struct
2. In `SyncToProjectConfig()`, after the existing user/language sync blocks, add a statusline sync block
3. The sync block reads the current `statusline.yaml`, updates the `theme` field, and writes back

### Cache Architecture for Rate Limits

Since the statusline runs as a subprocess per hook trigger, in-memory caching is not viable. The file-based cache strategy:

1. Check `~/.moai/cache/usage.json` for freshness (TTL check)
2. If fresh, return cached data (no I/O beyond file read)
3. If stale, fetch from API with 300ms timeout
4. Write new data atomically (temp file + os.Rename)
5. On any failure, return nil (segment silently omitted)

---

## 3. Architecture Design Direction

### Current Architecture

```
CLI (statusline.go)
  --> Builder (builder.go)
    --> GitDataProvider (git.go)
    --> UpdateProvider (update.go)
    --> Renderer (renderer.go)
      --> Hard-coded colors
```

### Target Architecture

```
CLI (statusline.go)
  --> Config loading (statusline.yaml)
  --> Builder (builder.go)
    --> GitDataProvider (git.go)
    --> UpdateProvider (update.go)
    --> RateLimitProvider (ratelimit.go) [Phase 6]
    --> Renderer (renderer.go)
      --> Theme (theme.go)
        --> DefaultTheme | CatppuccinMocha | CatppuccinLatte
      --> Multi-line layout (verbose mode)

Profile Wizard (profile_setup.go)
  --> ProfilePreferences (preferences.go)
    --> SyncToProjectConfig (sync.go)
      --> statusline.yaml
```

### Key Design Decisions

1. **Theme as interface, not config**: Themes are Go interfaces, not YAML configurations. This keeps the color definitions compile-time checked and avoids runtime parsing overhead.

2. **Gradient via method, not lookup table**: `BarGradient(pct)` uses method dispatch rather than a color lookup array. This allows themes to define custom gradient transitions without a fixed schema.

3. **File-based cache over in-memory**: The subprocess execution model of statusline hook necessitates file-based caching. This is consistent with the existing `update.go` caching pattern.

4. **Verbose mode as opt-in experiment**: Multi-line output may not be supported by all Claude Code versions. Document as experimental and provide single-line fallback.

---

## 4. Risks and Response Plans

| Risk                                     | Response Plan                                                    |
| ---------------------------------------- | ---------------------------------------------------------------- |
| Multi-line rendering issues in Claude Code | Document verbose mode as experimental; test compatibility; provide automatic fallback to single-line if newline is detected in output |
| catppuccin/go API changes                | Pin to v0.3.0; use only stable color constants                   |
| Performance regression from theme lookup | Benchmark `buildBar()` before/after; theme selection is interface dispatch (nanosecond overhead) |
| File cache race conditions               | Use atomic write pattern (write temp + rename); handle lock contention gracefully |
| Wizard backward compatibility            | New questions are appended (not inserted) to avoid reordering existing questions |
| statusline.yaml migration                | Default all new fields to backward-compatible values; explicit nil checks |

---

## 5. Development Methodology

This SPEC follows TDD (quality.yaml: `development_mode: tdd`):

- **RED**: Write failing test first
- **GREEN**: Write minimal code to pass
- **REFACTOR**: Clean up while tests remain green

Coverage target: 85% minimum per package.

All tests must use `t.TempDir()` for isolation to prevent modification of project files.

---

## 6. Expert Consultation Recommendations

This SPEC involves backend implementation (Go interfaces, file I/O, caching) and would benefit from:

- **expert-backend**: Architecture review for theme interface design, cache strategy, and provider patterns
- **expert-performance**: Benchmark validation for render time impact (REQ-SLE-NF-004)
