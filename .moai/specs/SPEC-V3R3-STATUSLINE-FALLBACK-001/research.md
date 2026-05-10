# Research: SPEC-V3R3-STATUSLINE-FALLBACK-001

## Statusline Architecture

### Call Chain

```
settings.json statusLine.command → bash .moai/status_line.sh → moai statusline (Go binary)
                                                               ↓
                                                         readStdinWithTimeout()
                                                               ↓
                                                         statusline.New(opts).Build(ctx, stdinData)
                                                               ↓
                                                         parseStdin(r) → *StdinData (nil on error)
                                                               ↓
                                                         collectAll(ctx, input) → *StatusData
                                                               ↓
                                                         renderer.Render(data, mode) → string
```

### Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `internal/cli/statusline.go` | CLI entry point, stdin reading, config loading | 169 |
| `internal/statusline/builder.go` | Build pipeline: parseStdin → collectAll → Render | 315 |
| `internal/statusline/renderer.go` | 3-line/5-line layout rendering | 537 |
| `internal/statusline/types.go` | StdinData JSON schema, type definitions | 315 |
| `internal/statusline/metrics.go` | Model/cost extraction from stdin | 193 |
| `internal/statusline/memory.go` | Context window extraction + GLM context window override | 286 |
| `internal/template/templates/.moai/status_line.sh.tmpl` | Shell wrapper (stdin relay to Go binary) | 36 |

### StdinData JSON Schema (types.go:57-71)

```go
type StdinData struct {
    HookEventName  string             `json:"hook_event_name"`
    SessionID      string             `json:"session_id"`
    TranscriptPath string             `json:"transcript_path"`
    CWD            string             `json:"cwd"`        // Legacy
    Model          *ModelInfo         `json:"model"`       // Can be object or string
    Workspace      *WorkspaceInfo     `json:"workspace"`
    Cost           *CostData          `json:"cost"`
    ContextWindow  *ContextWindowInfo `json:"context_window"`
    OutputStyle    *OutputStyleInfo   `json:"output_style"`
    RateLimits     *RateLimitInfo     `json:"rate_limits"`
    Effort         *EffortInfo        `json:"effort"`
    Thinking       *ThinkingInfo      `json:"thinking"`
    Version        string             `json:"version"`
}
```

## Current Fallback Behavior Analysis

### parseStdin() — builder.go:166-179

When stdin is empty, `{}`, `null`, or invalid JSON:
- `json.Decoder.Decode(&input)` returns error → `slog.Debug(...)` → returns `nil`
- No data is preserved; entire stdin is discarded

### collectAll() — builder.go:183-288 with `input == nil`

| Collector | With nil input | Result |
|-----------|---------------|--------|
| `CollectMemory(nil)` | `input == nil \|\| input.ContextWindow == nil` | `{Available: false}` — no CW bar |
| `CollectMetrics(nil)` | `input == nil` | `{Available: false}` — no model name |
| `CollectTask()` | No stdin dependency | Works — shows active task |
| `extractProjectDirectory(nil)` | `input == nil` | `""` — no directory shown |
| Worktree extraction | `input == nil` | `""` — no WT indicator |
| OutputStyle | `input == nil` | `""` — no output style |
| ClaudeCodeVersion | `input == nil` | `""` — no CC version |
| RateLimits | `input == nil` | `nil` — falls through to Usage API call |
| Effort/Thinking | `input == nil` | `nil` — no effort display |
| **Git collector** | Independent (filesystem) | **Works** — branch, status |
| **Version collector** | Independent (config) | **Works** — MoAI version |

### Summary: What breaks on empty stdin

- **Model name**: Missing entirely — most impactful user-visible gap
- **Project directory**: Missing — `renderDirGitLine` shows only branch
- **Context window bar**: Not shown — `Available: false`
- **Rate limits**: Missing — triggers blocking API call (5s timeout) which causes statusline disappearance
- **Claude Code version**: Missing
- **Output style**: Missing
- **Worktree indicator**: Missing
- **Effort/Thinking**: Missing

### What still works on empty stdin

- Git branch + ahead/behind
- Git status (staged/modified/untracked)
- MoAI-ADK version
- Active task display

## Shell Wrapper (status_line.sh.tmpl)

The shell wrapper:
1. Creates temp file, reads stdin into it
2. Sources GLM env vars
3. Tries `moai` in PATH, then hardcoded Go bin path, then `$HOME/.local/bin/moai`
4. Passes temp file as stdin to `moai statusline`

**Problem**: The cwd guard (from stash@{0}) was added to the *local* `.moai/status_line.sh` but NOT to the template. The template has NO cwd guard. This means:
- Template-generated statusline scripts have no cwd protection
- The cwd guard needs to move into the Go binary itself

## Test Coverage Gaps

### Existing tests (internal/cli/statusline_test.go)
- Command registration (exists, hidden, subcommand)
- `renderSimpleFallback()` — returns "moai"
- `loadSegmentConfig()` — YAML parsing
- `loadStatuslineFileConfig()` — YAML parsing

### Missing tests
- **parseStdin with empty reader** — what happens with `io.MultiReader()`
- **parseStdin with `{}`** — empty JSON object
- **parseStdin with `null`** — null literal
- **parseStdin with partial JSON** — e.g., `{"model": null}`
- **collectAll with nil input** — verify each collector's nil behavior
- **CWD-based fallback for project directory** — `getcwd()` when no workspace
- **Model name fallback** — env var or cache file
- **Renderer with partial StatusData** — which segments show/hide

### Existing tests in internal/statusline/
- `renderer_test.go`, `gradient_test.go`, `update_test.go`, `metrics_test.go` — cover rendering logic

## Recommendations for SPEC Scope

### In Scope (REQ-1 to REQ-4)

1. **REQ-1**: Empty/null stdin → fallback project directory from `os.Getwd()`, git branch from filesystem
2. **REQ-2**: Model name fallback via env var `MOAI_LAST_MODEL` or cache file `~/.moai/state/last-model.txt`
3. **REQ-3**: `Workspace.CurrentDir` fallback to `os.Getwd()` (already partially works via git collector)
4. **REQ-4**: Cwd guard in Go binary (absorb from shell wrapper)

### Out of Scope

- Settings.json `statusLine` key persistence (OQ-1 — requires SessionStart hook investigation)
- Branch output mismatch (OQ-2 — requires git worktree caching investigation)
- UI redesign of statusline layout
- New display modes

### Risk Assessment

- **Low risk**: Changes are additive fallbacks; existing happy path untouched
- **Test surface**: All fallback paths can be tested with nil/empty StdinData
- **Concurrency**: No new goroutines; git collector already thread-safe
- **Performance**: Fallbacks are instant (env var read, getcwd) — no I/O added to hot path

### Files to Modify

1. `internal/statusline/builder.go` — `parseStdin()` nil handling, `collectAll()` fallback logic
2. `internal/cli/statusline.go` — cwd guard before stdin read
3. `internal/statusline/metrics.go` — model name fallback (env/cache)
4. `internal/statusline/builder_test.go` — new test cases for nil/partial input
5. `internal/template/templates/.moai/status_line.sh.tmpl` — simplify (cwd guard absorbed by Go)
