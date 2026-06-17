# Module Catalog

Detailed reference for all packages in moai-adk-go (v3.0.0-rc2), grouped by architectural layer. Deterministic facts (package existence, public symbols) come from `go list` + `go doc`; the "Architectural Insight" subsections carry judgment, clearly labeled per the insight-vs-extraction convention.

---

## Layer 1: Entry Point

### `cmd/moai/`

**Public Surface**: `main.go` — composition root entry point that calls `cli.Execute()`.

**Dependencies**: `internal/cli`.

---

## Layer 2: Presentation

### `internal/cli/`

**Public Surface**: Cobra command tree spanning 40 files with 119 `rootCmd.AddCommand` calls. Human-facing terminal verbs (`init`, `doctor`, `update`, `version`, `cc`, `glm`, `cg`, `web`, `session`, `spec`, `harness`, `worktree`, `hook`, `agent`, `research`, `workflow`, `migrate`, `constitution`, `statusline`, `telemetry`, `profile`, `lsp`, `github`, `brain`, `loop`, `mx`, `clean`, `feedback`, `review`, `coverage`, `e2e`, `codemaps`, `design`, `project`, `plan`, `run`, `sync`). Composition Root in `deps.go` (`InitDependencies()`).

**Dependencies**: nearly every `internal/*` package (it is the wiring layer).

**Architectural Insight**: The CLI is intentionally the only place concrete types are instantiated. Every command accesses domain services through interfaces, which is what makes the 48-package codebase free of circular imports.

### `internal/tui/`

**Public Surface**: Bubbletea TUI components (statusline preview, interactive selectors, `moai --help` rendered groups).

**Dependencies**: `bubbletea`, `bubbles`, `lipgloss`, `glamour`.

---

## Layer 3: Interface / Protocol

### `internal/hook/` (named capability-anchor)

**Public Surface**: Handler registry + 130 hook source files covering 27 Claude Code event types (`SessionStart`, `Stop`, `PostToolUse`, `SubagentStop`, `TaskCompleted`, `UserPromptSubmit`, `PreCompact`, `PostCompact`, `Notification`, `WorktreeCreate`, `WorktreeRemove`, etc.). Public handler constructors (`NewSessionStartHandler`, …) registered in `internal/cli/deps.go`.

**Dependencies**: `internal/config`, `internal/lsp`, `internal/update`, `internal/session`.

**Architectural Insight**: Handlers follow a typed `Handler` interface contract; each event type maps to exactly one handler. The async-expand pattern debounces diff-aware reloads in a background goroutine with a 5-second deadline so the main handler returns within ≤100ms.

### `internal/config/`

**Public Surface**: YAML config loader + `MergedSettings` + the GLM tier-models constants table (`DefaultGLMBaseURL`, `DefaultGLMHigh`, `DefaultGLMMedium`, `DefaultGLMLow`, `DefaultGLMSonnet`, `DefaultGLMHaiku`, `DefaultGLMOpus`). Environment-key constants (`EnvConfigDir`, `EnvClaudeProjectDir`, `EnvAnthropicBaseURL`, …).

**Dependencies**: `pkg/models`, `internal/defs`.

### `internal/tmux/`

**Public Surface**: Split-pane detector + session lifecycle helpers for `moai cg` (Claude leader + GLM teammates).

**Dependencies**: `internal/shell`.

### `internal/web/` (named capability-anchor)

**Public Surface**: Loopback-only HTTP server (templ + HTMX) exposing browser-based editing of the named MoAI settings (profile preferences + `user.yaml` / `language.yaml` / `statusline.yaml` sections). Templ files: `root.templ`, `page.templ`, `fieldsets.templ`, `icons.templ`. Handlers in `handlers.go` + `projectconfig.go`.

**Dependencies**: `internal/config`, `net/http`, `a-h/templ`.

**Architectural Insight**: The Web Console is a thin browser equivalent of the terminal profile wizard — it reuses the same config-section validation, so the two surfaces cannot drift.

---

## Layer 4: Lifecycle

### `internal/spec/` (named capability-anchor)

**Public Surface**: `ValidStatuses` (8-value status enum), `FrontmatterSchemaRule` (12-field required schema), `ClassifyEra` (V2.x / V3R2-R4 / V3R5 / V3R6 / unclassified), `Audit` (drift detection engine), `Close` (atomic `moai spec close` transaction with precondition matrix + lock + dry-run), lint rules (`FrontmatterSchemaRule`, `OwnershipTransitionRule`, `StatusGitConsistencyRule`, …), `Acceptance` / `AuditResult` / `Era` types. Error sentinels: `ErrPreconditionMissing`, `ErrSpecCloseLockHeld`.

**Dependencies**: `internal/git`, `internal/state`, `internal/manifest`, `gopkg.in/yaml.v3`.

**Architectural Insight**: This is the single source of truth for SPEC lifecycle facts (status enum, frontmatter schema, era classification). Every docs claim about "8 statuses" or "12 frontmatter fields" cites this package. The grandfather clause (`Era.IsModern() == true` only for V3R6) prevents retroactive drift findings on legacy SPECs.

### `internal/state/`

**Public Surface**: Session checkpoint primitives (`state.go`, `store.go`, `lock.go`) + the multi-session coordination registry (`active-sessions.json`).

**Dependencies**: `internal/session`.

### `internal/workflow/`

**Public Surface**: Phase-state machine for the plan→run→sync lifecycle.

---

## Layer 5: Harness

### `internal/harness/` (named capability-anchor)

**Public Surface**: Frontmatter modification applier (`Apply`), cleanup tracker (`CleanupTracker`, `CleanupOnFailure`), frozen guard (`frozen_guard.go` prevents writes to moai-managed directories during meta-harness invocation), chaining-rules reader/writer, evaluator validation utilities.

**Dependencies**: `internal/evaluator`, `internal/evolution`, `internal/measure`, `internal/astgrep`.

**Architectural Insight**: The harness is a GAN (Generative Adversarial Network) loop applier — it enriches skill/agent frontmatter based on observed usage. The frozen guard is the safety boundary that keeps meta-harness output from clobbering hand-authored templates.

### `internal/evaluator/`

**Public Surface**: Prior-judgment leak detection for evaluator spawn prompts + 4-dimension quality scoring (Functionality / Security / Craft / Consistency).

### `internal/evolution/` / `internal/measure/`

**Public Surface**: 4-tier evolution proposals + measurement telemetry that feeds the harness learning subsystem.

---

## Layer 6: Templates

### `internal/template/` (named capability-anchor)

**Public Surface**: Deployer + renderer + `embedded.go` (generated `go:embed` of `templates/` + `catalog.yaml`). Typed `Catalog` loader (`catalog_loader.go`). `TemplateContext` (`GoBinPath`, `HomeDir`).

**Dependencies**: `pkg/models`, `internal/manifest`, `internal/merge`, `embed.FS`.

**Architectural Insight**: The `go:embed` directive compiles the entire `.claude/` + `.moai/` template tree into the binary, which is why `moai init` is zero-dependency. The catalog (`catalog.yaml`) is the template-tracking manifest that `moai update` diffs against for 3-way merges.

### `internal/manifest/`

**Public Surface**: Tracks which template files were deployed and at which version.

### `internal/merge/`

**Public Surface**: 3-way file merge engine used by `moai update`.

### `internal/migrate/` / `internal/migration/`

**Public Surface**: `.agency/` → `.moai/` migration utilities + legacy-archive helpers.

---

## Layer 7: Quality

### `internal/lsp/`

**Public Surface**: LSP diagnostics client with CLI fallbacks (`ruff`, `golangci-lint`, `eslint`, etc.) for the phase-specific quality gates.

### `internal/loop/`

**Public Surface**: Iterative fix loop (Ralph Engine) — LSP/AST-grep/test/coverage diagnostic cycle.

**Dependencies**: `internal/lsp`, `internal/resilience`, `internal/ralph`.

### `internal/resilience/`

**Public Surface**: Retry + recovery helpers used by the loop and harness subsystems.

### `internal/ralph/`

**Public Surface**: Reporting primitives for the Ralph Engine.

---

## Layer 8: Infrastructure

### `internal/statusline/` (named capability-anchor)

**Public Surface**: Status-line renderer for tmux/vim/zsh/bash. `CanonicalSegments` (`model`, `git`, `context`, `cost`, `version`, …), `ResolveGLMContextWindow`, `ShortenModelName`, `BuildGradientBar`, `BatteryIcon`, model cache read/write.

**Dependencies**: `internal/config`, `internal/git`.

### `internal/session/` (named capability-anchor)

**Public Surface**: Multi-session coordination registry. `RegisterSession` / `DeregisterSession` / `Heartbeat` / `PurgeStale`, `Entry` struct, `BlockerReport` type, `Checkpoint` / `Clock` interfaces, `HydrateForPrompt`. Registry file: `.moai/state/active-sessions.json` (per-machine, gitignored, advisory not a strong lock).

**Dependencies**: `internal/state`, `internal/config`.

**Architectural Insight**: The registry is advisory — it does NOT enforce mutual exclusion. The orchestrator queries it pre-spawn and decides whether to proceed, defer, or escalate via `AskUserQuestion`. This is the defense against multi-session races on the shared working tree.

### `internal/github/` / `internal/git/` / `internal/update/` / `internal/shell/`

**Public Surface**: GitHub API client (`gh` wrapper), Git operations (3-interface segregation: `Repository` / `BranchManager` / `WorktreeManager`), version-check updater, environment detection.

---

## Layer 9: Governance

### `internal/constitution/`

**Public Surface**: `CONST-*` zone registry (Frozen / Evolvable HARD clauses) + `moai constitution list` engine.

### `internal/mx/`

**Public Surface**: `@MX` code annotation engine (`@MX:NOTE`, `@MX:WARN`, `@MX:ANCHOR`, `@MX:TODO`, `@MX:LEGACY`) + graph queries.

### `internal/telemetry/` / `internal/permission/`

**Public Surface**: Opt-in usage stats + permission policy primitives.

---

## Layer 10: Foundation

### `internal/foundation/` / `internal/defs/` / `internal/i18n/` / `internal/astgrep/` / `internal/sandbox/` / `internal/runtime/` / `internal/bodp/` / `internal/ciwatch/`

**Public Surface**: Cross-cutting utilities — string helpers, constants (`defs`), internationalization (16-language), AST-grep wrapper, sandbox executor, runtime helpers, Branch Origin Decision Protocol (`bodp`), CI watch loop (`ciwatch`).

---

## Layer 11: Public API

### `pkg/models/` / `pkg/version/`

**Public Surface**: `Config`, `Project`, `Language` types (`models`) + build-time version injection (`version` reads from git tags).

---

## Cross-cutting Concerns

- **No circular imports**: dependency flow is strictly presentation → interface → domain → infrastructure.
- **Interface-first contracts**: every domain module exposes Go interfaces; concrete implementations are package-private.
- **Embedded template filesystem**: the entire `.claude/` + `.moai/` template tree is compiled into the binary via `go:embed`.
- **Tier-based PR routing**: Tier S/M work flows directly to `main` (Hybrid Trunk 1-person OSS); Tier L or explicit `--pr` opens a PR.
- **Era-classified SPEC lifecycle**: V3R6 SPECs are subject to drift detection; V2.x / V3R2-R4 / V3R5 are grandfather-clause-protected.
