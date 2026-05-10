# Module Catalog

Detailed reference for all packages in moai-adk-go, grouped by architectural layer.

---

## Layer 1: Entry Point

### `cmd/moai/`

| Attribute | Detail |
|-----------|--------|
| Path | `cmd/moai/main.go` |
| Purpose | Binary entry point. Delegates immediately to `cli.Execute()`. |
| LOC | ~16 |
| Key Exports | `main()` |

---

## Layer 2: Presentation

### `internal/cli/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/cli/` |
| Purpose | Complete Cobra command tree, dependency injection wiring (Composition Root), and all top-level CLI command implementations. |
| LOC | ~26,000 (including tests) |
| Key Exports | `Execute()`, `InitDependencies()`, `Dependencies` struct |

**Notable files:**

| File | Responsibility |
|------|---------------|
| `root.go` | Root `moai` command, `Execute()` entry, registers worktree and statusline subcommands |
| `deps.go` | **Composition Root** — `InitDependencies()` wires all domain modules; `Dependencies` struct holds all injected services |
| `init.go` | `moai init` — Project initialization wizard |
| `update.go` | `moai update` — Template sync to existing project |
| `hook.go` | `moai hook` — Dispatches Claude Code hook events via `hook.Registry` |
| `cc.go` | `moai cc` — Switches settings to Claude-only mode |
| `glm.go` | `moai glm` — Switches settings to GLM-only mode |
| `cg.go` | `moai cg` — Switches settings to Claude+GLM hybrid mode |
| `doctor.go` | `moai doctor` — Project health diagnostics |
| `github.go` | `moai github` — GitHub workflow setup |
| `rank.go` | `moai rank` — Rank API integration |
| `statusline.go` | `moai statusline` — Shell status line management |
| `status.go` | `moai status` — Project status report |
| `version.go` | `moai version` — Version display |

**Subdirectory:**

| Subdirectory | Purpose |
|-------------|---------|
| `internal/cli/worktree/` | `moai worktree` subcommand tree (new, list, switch, sync, clean, remove) |
| `internal/cli/var/` | Shared CLI variable helpers |

### `internal/ui/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/ui/` |
| Purpose | Interactive TUI components built with Bubbletea and Huh. Provides prompts, selection menus, progress displays, and markdown-rendered output used during `moai init` and wizard flows. |
| LOC | ~1,500 |
| Key Exports | `RunInitWizard()`, `SelectionModel`, `ConfirmPrompt` |

### `internal/tui/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/tui/` |
| Purpose | Theme-driven TUI rendering layer (SPEC-V3R3-CLI-TUI-001). Provides 14 reusable primitives with auto-detection, i18n support, and NO_COLOR compliance. Single source of truth for brand-consistent styling across all CLI commands. |
| LOC | ~8,500 (including tests + i18n catalogs) |
| Key Exports | `Theme()`, `Box()`, `Pill()`, `StatusIcon()`, `Spinner()`, `Progress()`, `Stepper()`, `Table()`, `Form()`, `Prompt()`, `Help()`, `Term()`, `Resolve()`, `LoadCatalog()` |

**Core Primitives (14 surface APIs):**

| Component | Purpose | Golden Snapshots |
|-----------|---------|------------------|
| `theme.go` | Light/Dark token singleton, Catppuccin palette, adaptive color profiles | 28 light/dark |
| `box.go` | Panel/card borders via lipgloss (Box, ThickBox, rounded variants) | 8 + 8 |
| `pill.go` | Colored badges with variant support (Primary/Ok/Info/Warn/Error) | 24 |
| `status.go` | StatusIcon glyphs, Spinner (animated + reduced-motion fallback), Progress | 13 |
| `form.go` | RadioRow/CheckRow wrappers around huh, custom theming (◆/◇ prefix) | 5 |
| `table.go` | KV pairs, CheckLine rows, Section dividers with Unicode fallback | 43 |
| `prompt.go` | Text input + cursor control (pure functions, no lifecycle) | 7 |
| `help.go` | HelpBar + KeyHint elements for command usage | 4 |
| `term.go` | Screenshot capture mode (MOAI_SCREENSHOT=1 gate) | 2 |
| `detect.go` | Auto-detect Theme based on MOAI_THEME / TTY / NO_COLOR / TERM | 12 |
| `i18n.go` | 14-language message catalog loading (embed.FS, zero filesystem) | 6 |
| `profile.go` | ColorProfile detection via termenv, fallback levels (Truecolor→256→basic) | (integration) |

**Files:**

| File | Purpose | Lines |
|------|---------|-------|
| `theme.go` + `theme_test.go` | Token definitions, singleton instance, color palette | 100 + 140 |
| `box.go` + `box_test.go` | Panel/card rendering via lipgloss borders | 85 + 110 |
| `pill.go` + `pill_test.go` | Colored badge badges (Primary/Ok/Info/Warn/Error) | 120 + 45 |
| `status.go` + `status_test.go` | StatusIcon, Spinner, Progress — reduced-motion support | 145 + 170 |
| `form.go` + `form_test.go` | RadioRow/CheckRow huh wrappers with theming | 90 + 85 |
| `table.go` + `table_test.go` | KV, CheckLine, Section — Unicode escape handling | 130 + 180 |
| `prompt.go` + `prompt_test.go` | Text prompt + cursor rendering | 60 + 70 |
| `help.go` + `help_test.go` | HelpBar/KeyHint elements | 45 + 60 |
| `term.go` + `term_test.go` | Screenshot-mode capture (MOAI_SCREENSHOT=1) | 80 + 65 |
| `detect.go` + `detect_test.go` | Auto-detect theme logic + TTY/NO_COLOR/TERM | 120 + 140 |
| `i18n.go` + `i18n_test.go` | 14-language catalog loader via embed.FS | 130 + 110 |
| `profile.go` + `profile_test.go` | ColorProfile detection interface + production implementation | 95 + 80 |
| `doc.go` | Godoc + design source attribution | 30 |
| `catppuccin.go` | Catppuccin palette constants (Mocha/Latte) | 60 |
| `golden/index_test.go` | Golden snapshot validation suite (33 errchecks fixed) | 180 |

**Subdirectory:**

| Subdirectory | Purpose |
|-------------|---------|
| `internal/tui/internal/` | `runeguard.go` helper — East Asian width detection for Korean text (고정폭 character measurement) |
| `internal/tui/messages/` | 14-language YAML catalogs (ko.yaml complete + ja/zh/en + 10 stubs with @MX:TODO for future translation) |
| `internal/tui/testdata/` | 106 golden snapshot files (visual regression guards for all primitives) |

**Quality & Standards:**

- **TRUST 5 Compliance**: ✅ All five pillars (Tested 100%, Readable with @MX:ANCHOR/NOTE, Unified theming, Secured no-new-surface, Trackable commit trail)
- **AC Coverage**: ✅ 17/17 acceptance criteria complete (AC-CLI-TUI-001 through AC-CLI-TUI-017)
- **No Hex Literals**: ✅ zero hardcoded `#RRGGBB` outside theme.go (brand source of truth)
- **No Hand-Drawn Box Chars**: ✅ all borders via lipgloss (no U+2500 ─ / U+2502 │)
- **NO_COLOR Support**: ✅ detect.go applies throughout rendering layer
- **Reduced-Motion**: ✅ Spinner/Progress fallback to static characters when `MOAI_REDUCED_MOTION=1`
- **i18n Foundation**: ✅ 14-language message catalog with embed.FS (zero filesystem dependency at runtime)

---

## Layer 3: Interface / Protocol

### `internal/hook/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/hook/` |
| Purpose | Claude Code hook event system. Defines the `Registry` interface for event dispatch, the `Handler` interface for individual handlers, and all 24 handler implementations covering 17 event types. |
| LOC | ~10,500 (including tests) |
| Key Exports | `Registry`, `Handler`, `Protocol`, `NewRegistry()`, `NewProtocol()` |

**Handler inventory:**

| Handler File | Event Type | Purpose |
|-------------|-----------|---------|
| `session_start.go` | `SessionStart` | Checks for updates, banner display, config validation |
| `session_end.go` | `SessionEnd` | Session metrics, cleanup |
| `pre_tool.go` | `PreToolUse` | Security scanning via AST analysis before tool execution |
| `post_tool.go` | `PostToolUse` | LSP diagnostics collection after tool execution |
| `post_tool_metrics.go` | `PostToolUse` | Token and timing metrics |
| `post_tool_failure.go` | `PostToolUseFailure` | Error handling and reporting |
| `stop.go` | `Stop` | Agent stop signal handling |
| `teammate_idle.go` | `TeammateIdle` | LSP quality gate enforcement (exit 2 = keep working) |
| `task_completed.go` | `TaskCompleted` | Task completion validation |
| `worktree_create.go` | `WorktreeCreate` | Worktree lifecycle logging (v2.1.49+) |
| `worktree_remove.go` | `WorktreeRemove` | Worktree cleanup logging (v2.1.49+) |
| `user_prompt_submit.go` | `UserPromptSubmit` | Prompt preprocessing |
| `permission_request.go` | `PermissionRequest` | Permission gate control |
| `subagent_start.go` | `SubagentStart` | Subagent initialization |
| `notification.go` | `Notification` | Notification routing |
| `compact.go` | `PreCompact` | Context compaction handling |
| `auto_update.go` | `SessionStart` | Auto-update trigger logic |
| `rank_session.go` | `SessionStart/End` | Rank API session tracking |

**Supporting files:**

| File | Purpose |
|------|---------|
| `registry.go` | Event dispatcher — routes stdin JSON to typed handlers |
| `types.go` | All event payload types (14,800 LOC — largest file in the package) |
| `protocol.go` | JSON stdin/stdout protocol implementation |
| `contract.go` | `Handler` and `Registry` interface definitions |

**Subdirectory:**

| Subdirectory | Purpose |
|-------------|---------|
| `internal/hook/security/` | AST-based security scanner used by `PreToolUse` handler |

### `internal/config/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/config/` |
| Purpose | YAML configuration file management. Reads and writes `.moai/config/sections/*.yaml` files. Implements `ConfigManager` and `ConfigProvider` interfaces used by hooks and CLI commands. |
| LOC | ~900 |
| Key Exports | `ConfigManager`, `ConfigProvider`, `NewConfigManager()` |

### `internal/tmux/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/tmux/` |
| Purpose | Tmux session detection and split-pane integration. Used by `moai cg` for Claude+GLM hybrid mode visual split. |
| LOC | ~350 |
| Key Exports | `IsRunning()`, `NewPane()`, `SendKeys()` |

---

## Layer 4: Domain

### `internal/core/project/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/core/project/` |
| Purpose | Project lifecycle management. Handles detection of existing projects, initialization workflows, methodology auto-detection (TDD vs DDD), phase management, and TRUST 5 validation coordination. |
| LOC | ~2,500 (excluding tests) |
| Key Exports | `Initializer`, `Detector`, `Validator`, `MethodologyDetector`, `PhaseManager` |

**Files:**

| File | Purpose |
|------|---------|
| `initializer.go` | Full project init workflow (19K chars — complex orchestration) |
| `detector.go` | Project root detection and existing project identification |
| `methodology_detector.go` | Analyzes coverage to recommend TDD vs DDD |
| `phase.go` | Workflow phase state tracking (plan/run/sync) |
| `validator.go` | Validates project structure and configuration |
| `reporter.go` | Generates project health reports |

### `internal/core/git/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/core/git/` |
| Purpose | Git repository abstractions. Provides three distinct interfaces for repository queries, branch management, and worktree management. |
| LOC | ~2,000 (excluding tests) |
| Key Exports | `Repository`, `BranchManager`, `WorktreeManager`, `Manager` (implements all three) |

**Files:**

| File | Purpose |
|------|---------|
| `manager.go` | Concrete `Manager` struct implementing all three interfaces via `git` CLI |
| `types.go` | Domain types: `Branch`, `Worktree`, `CommitInfo` |
| `branch.go` | Branch creation, switching, listing, upstream tracking |
| `worktree.go` | Worktree create, remove, list, sync operations |
| `conflict.go` | Merge conflict detection and reporting |
| `event.go` | Git event types for hook integration |

### `internal/core/quality/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/core/quality/` |
| Purpose | TRUST 5 quality framework validation. Coordinates test coverage checks, LSP diagnostics, and methodology-specific quality gates. |
| LOC | ~600 |
| Key Exports | `Validator`, `Report`, `TrustScore` |

### `internal/template/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/template/` |
| Purpose | Template deployment system. Embeds all `.claude/`, `.moai/`, and `CLAUDE.md` templates into the binary using `go:embed`. Handles rendering (Go text/template), deployment (file extraction + merge), model policy application, and settings.json generation. |
| LOC | ~1,800 (excluding tests) |
| Key Exports | `Deployer`, `Renderer`, `TemplateContext`, `DeployerMode`, `ApplyModelPolicy()` |

**Files:**

| File | Purpose |
|------|---------|
| `embed.go` | `//go:embed templates` directive — mounts the embedded FS |
| `deployer.go` | Main deployment logic: extract, merge, write files |
| `deployer_mode.go` | Mode enum (init vs update) with mode-specific rules |
| `renderer.go` | Go text/template rendering with `TemplateContext` |
| `context.go` | `TemplateContext` builder (GoBinPath, HomeDir, platform vars) |
| `model_policy.go` | Per-agent model assignment based on policy level (high/medium/low) |
| `settings.go` | `settings.json` generation for Claude Code |
| `validator.go` | Validates template structure and required file presence |

### `internal/workflow/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/workflow/` |
| Purpose | Workflow state management. Tracks current SPEC phase (plan/run/sync), persists state to `.moai/` directory, and provides phase transition validation. |
| LOC | ~400 |
| Key Exports | `StateManager`, `Phase`, `WorkflowState` |

### `internal/loop/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/loop/` |
| Purpose | Iterative fix loop implementation (`/moai loop`). Drives automated error detection and fix cycles using LSP diagnostics as the quality signal. Implements retry logic with configurable maximum iterations. |
| LOC | ~600 |
| Key Exports | `Runner`, `LoopConfig`, `LoopResult` |

### `internal/merge/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/merge/` |
| Purpose | 3-way file merge for template updates. Compares the original template, user-modified version, and new template to produce a merged result. Used by `moai update` to preserve user customizations. |
| LOC | ~700 |
| Key Exports | `Merge3Way()`, `MergeResult`, `ConflictSection` |

### `internal/core/integration/` and `internal/core/migration/`

| Attribute | Detail |
|-----------|--------|
| Purpose | Integration test support and version migration utilities. |
| LOC | ~200 combined |

---

## Layer 5: Infrastructure

### `internal/lsp/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/lsp/` |
| Purpose | LSP (Language Server Protocol) diagnostic collection and caching. Provides both a direct LSP client and fallback CLI-based diagnostics (`go vet`, `golangci-lint`). Results feed quality gates in `TeammateIdle` and `TaskCompleted` hooks. |
| LOC | ~900 (excluding tests) |
| Key Exports | `DiagnosticsCollector`, `FallbackDiagnostics`, `DiagnosticResult`, `Cache` |

**Subdirectory:**

| Subdirectory | Purpose |
|-------------|---------|
| `internal/lsp/hook/` | Hook-specific LSP wrappers (`lsphook.NewDiagnosticsCollector()`) |

### `internal/github/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/github/` |
| Purpose | GitHub API integration. Handles repository setup, GitHub Actions workflow generation, issue creation (for `/moai feedback`), and PR operations. |
| LOC | ~1,200 (excluding tests) |
| Key Exports | `Client`, `WorkflowGenerator`, `IssueCreator` |

### `internal/rank/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/rank/` |
| Purpose | Rank API integration for session metrics and developer ranking. Tracks tool usage, session duration, and quality metrics. Provides credential storage and browser-based OAuth flow. |
| LOC | ~1,100 (excluding tests) |
| Key Exports | `Client`, `CredentialStore`, `BrowserOpener`, `SessionMetrics` |

### `internal/update/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/update/` |
| Purpose | Version check and auto-update system. Compares current version against latest GitHub release, downloads and installs new binary if available. Provides `Checker` and `Orchestrator` interfaces. |
| LOC | ~800 (excluding tests) |
| Key Exports | `Checker`, `Orchestrator`, `NewChecker()`, `NewOrchestrator()` |

### `internal/shell/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/shell/` |
| Purpose | Shell environment detection. Identifies current shell (zsh, bash, fish), detects Go binary paths, and determines platform-specific configuration locations. |
| LOC | ~400 |
| Key Exports | `Detector`, `ShellInfo`, `GoBinPath()` |

### `internal/statusline/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/statusline/` |
| Purpose | Shell prompt status line integration. Generates and manages a dynamic status indicator for zsh/bash prompts showing MoAI workflow phase, project info, and quality gate status. |
| LOC | ~900 (excluding tests) |
| Key Exports | `Manager`, `StatusInfo`, `RenderStatusLine()` |

### `internal/resilience/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/resilience/` |
| Purpose | Retry and error recovery patterns. Provides exponential backoff, circuit breaker, and fallback strategies used throughout the codebase for network calls and CLI operations. |
| LOC | ~500 |
| Key Exports | `Retry()`, `RetryConfig`, `CircuitBreaker` |

---

## Layer 6: Cross-cutting Support

### `internal/foundation/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/foundation/` |
| Purpose | Shared utility functions. File system helpers, string utilities, path normalization, and common error types used across packages. |
| LOC | ~600 |
| Key Exports | `SafeWrite()`, `EnsureDir()`, `NormalizePath()` |

### `internal/defs/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/defs/` |
| Purpose | Project-wide constant definitions. Canonical paths (`.moai/`, `.claude/`), event type names, file name constants, and configuration key names. |
| LOC | ~300 |
| Key Exports | `MoaiDir`, `ClaudeDir`, `SpecsDir`, event type constants |

### `internal/i18n/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/i18n/` |
| Purpose | Internationalization support for user-facing CLI messages. Supports English, Korean, Japanese, and Chinese based on `language.conversation_language` configuration. |
| LOC | ~200 |
| Key Exports | `T()`, `Locale`, `LoadLocale()` |

### `internal/manifest/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/manifest/` |
| Purpose | Template deployment manifest tracking. Records which files were deployed, at what template version, enabling `moai update` to detect user modifications and apply 3-way merges safely. |
| LOC | ~350 |
| Key Exports | `Manifest`, `FileRecord`, `Load()`, `Save()` |

### `internal/astgrep/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/astgrep/` |
| Purpose | AST-based code pattern matching. Wraps the `ast-grep` CLI tool for structural code search used by the security scanner in `hook/security/`. |
| LOC | ~300 |
| Key Exports | `Search()`, `Pattern`, `Match` |

### `internal/ralph/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/ralph/` |
| Purpose | Reporting utilities. Formats structured quality and session reports using Markdown and plain text renderers. |
| LOC | ~150 |
| Key Exports | `Reporter`, `Report`, `FormatMarkdown()` |

### `internal/git/`

| Attribute | Detail |
|-----------|--------|
| Path | `internal/git/` |
| Purpose | Low-level Git convention detection and branch naming utilities. Separate from `internal/core/git/` — provides standalone branch pattern detection without the full domain model. |
| LOC | ~400 |
| Key Exports | `BranchDetector`, `ConventionChecker`, `DetectBranchPattern()` |

---

## Layer 7: Public API

### `pkg/models/`

| Attribute | Detail |
|-----------|--------|
| Path | `pkg/models/` |
| Purpose | Shared data model types exported for external use. Includes `Config`, `Project`, `Language`, `User` structs that are used by both CLI and template packages. |
| LOC | ~300 |
| Key Exports | `Config`, `Project`, `Language`, `User`, `QualityConfig` |

### `pkg/version/`

| Attribute | Detail |
|-----------|--------|
| Path | `pkg/version/` |
| Purpose | Version information management. Reads version from build-time `ldflags` injection. Provides `GetVersion()` used throughout the CLI. |
| LOC | ~100 |
| Key Exports | `GetVersion()`, `Version` (set via ldflags) |

---

## Cross-cutting Concerns

| Concern | Handled By |
|---------|-----------|
| Configuration access | `internal/config.ConfigManager` (singleton via `deps.go`) |
| Error wrapping | `fmt.Errorf("context: %w", err)` — enforced project-wide |
| Logging | `log/slog` — text handler to discard in production; structured in debug |
| Testing isolation | `t.TempDir()` — all tests use temp directories, never the project root |
| Concurrency safety | `go test -race ./...` — enforced in CI |
| Embedded assets | `//go:embed` in `internal/template/embed.go` |
