# SPEC-SLE-001: Acceptance Criteria

## Metadata

| Field    | Value              |
| -------- | ------------------ |
| SPEC ID  | SPEC-SLE-001       |
| Title    | Statusline Enhancement |
| Status   | Planned            |

---

## 1. Phase 1: Config Layer Foundation

### AC-1.1: StatuslineConfig struct exists and loads correctly

```gherkin
Given the config system is initialized
When .moai/config/sections/statusline.yaml contains preset, segments, and theme fields
Then StatuslineConfig is populated with the correct values

Given the config system is initialized
When .moai/config/sections/statusline.yaml does not exist
Then StatuslineConfig defaults to preset "full", all segments true, theme "default"

Given the config system is initialized
When statusline.yaml exists but has no theme field
Then StatuslineConfig defaults theme to "default" while preserving existing preset and segments
```

### AC-1.2: Template includes theme field

```gherkin
Given the template at internal/template/templates/.moai/config/sections/statusline.yaml
When moai init is executed
Then the deployed statusline.yaml contains a theme field set to "default"
```

---

## 2. Phase 2: Theme System

### AC-2.1: Theme interface implementations

```gherkin
Given the theme name is "default"
When NewTheme("default") is called
Then a DefaultTheme instance is returned

Given the theme name is "catppuccin-mocha"
When NewTheme("catppuccin-mocha") is called
Then a CatppuccinMocha instance is returned with Primary() returning "#CDD6F4"

Given the theme name is "catppuccin-latte"
When NewTheme("catppuccin-latte") is called
Then a CatppuccinLatte instance is returned with Primary() returning "#4C4F69"

Given an unknown theme name "nonexistent"
When NewTheme("nonexistent") is called
Then a DefaultTheme instance is returned as fallback
```

### AC-2.2: 4-stage gradient bar

```gherkin
Given the theme is CatppuccinMocha
When BarGradient(10) is called
Then the returned color is "#A6E3A1" (green, Stage 1)

Given the theme is CatppuccinMocha
When BarGradient(40) is called
Then the returned color is "#F9E2AF" (yellow, Stage 2)

Given the theme is CatppuccinMocha
When BarGradient(60) is called
Then the returned color is "#FAB387" (peach, Stage 3)

Given the theme is CatppuccinMocha
When BarGradient(90) is called
Then the returned color is "#F38BA8" (red, Stage 4)

Given the theme is CatppuccinLatte
When BarGradient(10) is called
Then the returned color is "#40A02B" (green, Stage 1)

Given the theme is CatppuccinLatte
When BarGradient(90) is called
Then the returned color is "#D20F39" (red, Stage 4)
```

### AC-2.3: Boundary values for gradient stages

```gherkin
Given the theme is CatppuccinMocha
When BarGradient(0) is called
Then the returned color is "#A6E3A1" (Stage 1)

Given the theme is CatppuccinMocha
When BarGradient(25) is called
Then the returned color is "#A6E3A1" (Stage 1, boundary)

Given the theme is CatppuccinMocha
When BarGradient(26) is called
Then the returned color is "#F9E2AF" (Stage 2, boundary)

Given the theme is CatppuccinMocha
When BarGradient(50) is called
Then the returned color is "#F9E2AF" (Stage 2, boundary)

Given the theme is CatppuccinMocha
When BarGradient(51) is called
Then the returned color is "#FAB387" (Stage 3, boundary)

Given the theme is CatppuccinMocha
When BarGradient(75) is called
Then the returned color is "#FAB387" (Stage 3, boundary)

Given the theme is CatppuccinMocha
When BarGradient(76) is called
Then the returned color is "#F38BA8" (Stage 4, boundary)

Given the theme is CatppuccinMocha
When BarGradient(100) is called
Then the returned color is "#F38BA8" (Stage 4)
```

### AC-2.4: Theme wired into renderer

```gherkin
Given a renderer is created with theme "catppuccin-mocha"
When renderContextGraph() is called with 40% usage
Then the bar graph fill color matches CatppuccinMocha.BarGradient(40)

Given a renderer is created with NoColor = true
When rendering any segment
Then no ANSI escape codes are present in the output regardless of theme
```

### AC-2.5: Theme selection from CLI config

```gherkin
Given statusline.yaml has theme: "catppuccin-mocha"
When moai statusline is executed
Then the renderer uses the CatppuccinMocha theme

Given statusline.yaml has no theme field
When moai statusline is executed
Then the renderer uses the DefaultTheme
```

---

## 3. Phase 3: Profile Sync Fix

### AC-3.1: ProfilePreferences includes StatuslineTheme

```gherkin
Given a ProfilePreferences struct
When StatuslineTheme is set to "catppuccin-mocha"
Then the field is persisted to preferences.yaml correctly

Given preferences.yaml has no StatuslineTheme field
When ProfilePreferences is loaded
Then StatuslineTheme defaults to empty string
```

### AC-3.2: SyncToProjectConfig syncs statusline settings

```gherkin
Given ProfilePreferences has StatuslineTheme = "catppuccin-latte"
When SyncToProjectConfig() is called
Then .moai/config/sections/statusline.yaml contains theme: "catppuccin-latte"

Given ProfilePreferences has StatuslineTheme = "catppuccin-mocha"
And statusline.yaml already exists with preset: "compact" and segments configured
When SyncToProjectConfig() is called
Then the theme field is updated to "catppuccin-mocha"
And the existing preset and segments are preserved

Given statusline.yaml does not exist
When SyncToProjectConfig() is called with StatuslineTheme = "default"
Then statusline.yaml is created with preset: "full", all segments: true, theme: "default"
```

---

## 4. Phase 4: Multi-Line Verbose Mode

### AC-4.1: ModeDefault is unchanged (backward compatibility)

```gherkin
Given a renderer with mode ModeDefault
When Render() is called with full StatusData
Then the output is a single line (no newline characters)
And the output contains all enabled segments separated by " | "
```

### AC-4.2: ModeVerbose produces multi-line output

```gherkin
Given a renderer with mode ModeVerbose
And StatusData has model, context, style, directory, branch, git changes, Claude version, MoAI version, and cost
When Render() is called
Then the output contains exactly 2 newline characters (3 lines)
And Line 1 contains Model, Context Graph, and Output Style
And Line 2 contains Directory, Branch, and Git Changes
And Line 3 contains Claude Version, MoAI Version, and Cost
```

### AC-4.3: Empty verbose lines are omitted

```gherkin
Given a renderer with mode ModeVerbose
And StatusData has model and context but no directory, branch, git changes, version, or cost
When Render() is called
Then the output contains 0 newline characters (only Line 1)
And Line 1 contains Model and Context Graph

Given a renderer with mode ModeVerbose
And StatusData has model, context, directory, branch but no version or cost
When Render() is called
Then the output contains exactly 1 newline character (2 lines)
And Line 1 contains Model and Context Graph
And Line 2 contains Directory and Branch
```

### AC-4.4: Cost rendering format

```gherkin
Given a renderer with mode ModeVerbose
And MetricsData.CostUSD = 0.42
When Render() is called
Then Line 3 contains "$0.42"

Given a renderer with mode ModeVerbose
And MetricsData.CostUSD = 0.0
When Render() is called
Then the cost segment is omitted from Line 3

Given a renderer with mode ModeVerbose
And MetricsData.CostUSD = 15.7891
When Render() is called
Then Line 3 contains "$15.79" (2 decimal places)
```

### AC-4.5: ModeMinimal is unchanged

```gherkin
Given a renderer with mode ModeMinimal
When Render() is called with full StatusData
Then the output is a single line containing only Model and Context Graph
```

---

## 5. Phase 5: Profile Wizard Enhancement

### AC-5.1: Theme selector appears in wizard

```gherkin
Given the user runs profile setup wizard
When the Display section is reached
Then a "Statusline Theme" question is presented
And three options are available: default, catppuccin-mocha, catppuccin-latte
```

### AC-5.2: Theme selection is saved

```gherkin
Given the user selects "catppuccin-mocha" in the theme selector
When the wizard completes
Then preferences.yaml contains StatuslineTheme: "catppuccin-mocha"
And SyncToProjectConfig() is called to update statusline.yaml
```

### AC-5.3: Theme selector pre-fills existing value

```gherkin
Given preferences.yaml contains StatuslineTheme: "catppuccin-latte"
When the profile setup wizard is opened
Then the theme selector shows "catppuccin-latte" as the pre-selected option
```

### AC-5.4: Translations are complete

```gherkin
Given the conversation language is "ko"
When the theme selector question is displayed
Then the question text is in Korean

Given the conversation language is "ja"
When the theme selector question is displayed
Then the question text is in Japanese

Given the conversation language is "zh"
When the theme selector question is displayed
Then the question text is in Chinese
```

---

## 6. Phase 6: API Rate Limit Monitoring (Optional)

### AC-6.1: RateLimitProvider interface

```gherkin
Given a RateLimitProvider implementation
When FetchUsage(ctx) is called
Then it returns a RateLimitData with Usage5h and Usage7d percentages
```

### AC-6.2: Cache behavior

```gherkin
Given ~/.moai/cache/usage.json has fresh data (less than 60 seconds old)
When FetchUsage() is called
Then cached data is returned without API call

Given ~/.moai/cache/usage.json has stale data (more than 60 seconds old)
When FetchUsage() is called
Then the API is called and the cache is updated

Given ~/.moai/cache/usage.json does not exist
When FetchUsage() is called
Then the API is called and the cache file is created
```

### AC-6.3: Graceful degradation

```gherkin
Given ANTHROPIC_OAUTH_TOKEN is not set
When the statusline is rendered
Then the rate limit segment is not displayed
And no error message is shown

Given the API call takes longer than 300ms
When FetchUsage() is called
Then the call is cancelled after 300ms
And the cached or nil result is returned

Given the API returns an error
When FetchUsage() is called
Then nil is returned
And the rate limit segment is not displayed
```

### AC-6.4: Color-coded display

```gherkin
Given usage data is available with Usage5h = 60%
When the rate limit segment is rendered with CatppuccinMocha theme
Then the 5h bar uses Stage 3 color (peach, #FAB387)

Given usage data is available with Usage7d = 85%
When the rate limit segment is rendered with CatppuccinMocha theme
Then the 7d bar uses Stage 4 color (red, #F38BA8)
```

---

## 7. Quality Gate Criteria

### Definition of Done

- [ ] All EARS requirements (REQ-SLE-*) have corresponding test cases
- [ ] Test coverage >= 85% per package for all modified/new files
- [ ] `go vet ./...` passes without errors
- [ ] `go test -race ./...` passes without errors
- [ ] `golangci-lint run` passes without errors
- [ ] `make build` succeeds (embedded templates regenerated)
- [ ] No regressions in existing statusline rendering (ModeDefault, ModeMinimal)
- [ ] Theme switching adds less than 5ms to render time
- [ ] Backward compatible: projects without `theme` field in statusline.yaml work correctly
- [ ] All 4 language translations verified for wizard questions (en, ko, ja, zh)

### Verification Methods

| Criterion                   | Method                                | Tool                    |
| --------------------------- | ------------------------------------- | ----------------------- |
| Test coverage               | `go test -cover ./internal/statusline/...` | go test           |
| Race conditions             | `go test -race ./...`                 | go test                 |
| Lint compliance             | `golangci-lint run`                   | golangci-lint           |
| Static analysis             | `go vet ./...`                        | go vet                  |
| Render performance          | Benchmark test for `buildBar()`       | go test -bench          |
| Backward compatibility      | Test with empty SegmentConfig + nil theme | table-driven tests |
| Template regeneration       | `make build`                          | make                    |
| Cross-platform              | CI matrix: macOS, Linux, Windows      | GitHub Actions          |
