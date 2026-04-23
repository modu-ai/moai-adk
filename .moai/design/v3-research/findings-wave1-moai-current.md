# Wave 1.6 — moai-adk-go Current State Snapshot

Generated: 2026-04-22
Researcher: Wave 1.6 (moai-current snapshot)
Source tree: `/Users/goos/MoAI/moai-adk-go`
Git HEAD: `ad2019683` (main branch, clean)
Tag: `v2.12.0-15-gad2019683`
Go module: `github.com/modu-ai/moai-adk` (go.mod line 1)
Go toolchain: `go 1.26` (go.mod line 3)

This document is an evidence-based inventory of the current moai-adk-go (Go Edition) implementation. Every non-trivial claim is anchored to `file:line` or listed as a file existence claim. No content is interpreted beyond what is visible in the source.

---

## 1. CLI (cmd/moai/) — subcommand list + flags

### 1.1 Entry point

- `cmd/moai/main.go:12` — `main()` calls `cli.Execute()` and exits 1 on error.
- `cmd/moai/main.go` is the SOLE entry binary (@MX:ANCHOR tagged with fan_in=1).

### 1.2 Root command registration (cobra)

`internal/cli/root.go:13-76` registers the root command and three cobra groups.

- Root metadata: `Use: "moai"`, `Short: "MoAI-ADK: Agentic Development Kit for Claude Code"`, `Version: version.GetVersion()`.
- Command groups registered (`root.go:43-47`):
  - `launch` — "Launch Commands:"
  - `project` — "Project Commands:"
  - `tools` — "Tools:"
- Root subcommands registered in `root.go`: `worktree.WorktreeCmd` (lazy git init via PersistentPreRunE), `StatuslineCmd`, `NewAstGrepCmd()`, `telemetryCmd`.
- `Execute()` (`root.go:33-37`) calls `initConsole()` and `InitDependencies()` before delegating to cobra.

### 1.3 Top-level subcommand inventory

Grep `rootCmd.AddCommand` in `internal/cli/*.go` yields 20 distinct top-level commands (each defined in its own `<name>.go` file under `internal/cli/`):

| Command | File | Purpose |
|---------|------|---------|
| `cc` | `internal/cli/cc.go:13` (`Use: "cc [-p profile] [-- claude-args...]"`) | Launch Claude Code (default Anthropic API) |
| `cg` | `internal/cli/cg.go:8` (`Use: "cg [-p profile]"`) | Launch Claude with GLM in tmux (CG mode) |
| `glm` | `internal/cli/glm.go:20` (`Use: "glm [-p profile] [-- claude-args...]"`) | Launch Claude with GLM directly; subcommands `setup`, `status` at glm.go:61, 68 |
| `doctor` | `internal/cli/doctor.go:39` | Environment diagnostic |
| `github` | `internal/cli/github.go:66` (subcommands `parse-issue` at :80, `link-spec` at :157) | GitHub issue integration |
| `lsp` | `internal/cli/lsp_doctor.go:54` (`doctor` sub at :62, registered at :121) | LSP diagnostic |
| `init` | `internal/cli/init.go:26` | Initialize MoAI project |
| `hook` | `internal/cli/hook.go:17` (see §5 for sub inventory) | Claude Code hook dispatch |
| `loop` | `internal/cli/loop.go:11` (subs `start <SPEC-ID>` :39, `status` :47, `pause` :55, `resume <SPEC-ID>` :63, `cancel` :71) | Ralph loop engine |
| `mcp` | `internal/cli/mcp.go:15` (`lsp` sub at :21) | MCP lsp stub server |
| `migrate` | `internal/cli/migrate_agency.go:569` (`agency` sub at :576) | `/agency` → `/moai design` migration |
| `research` | `internal/cli/research.go:16` (subs `status` :30, `baseline [target]` :72, `list` :84) | Research observations/experiments |
| `status` | `internal/cli/status.go:22` | Show current project status |
| `update` | `internal/cli/update.go:66` | Update embedded templates to current binary version |
| `profile` | `internal/cli/profile.go:11` (subs `list`, `current`, `delete`) | Claude configuration profile manager |
| `worktree` | `internal/cli/worktree/root.go:16` (see §1.4) | MoAI git worktree manager |
| `statusline` | `internal/cli/statusline.go:18` | Render statusline for Claude Code |
| `ast-grep` | `internal/cli/astgrep.go:29` | ast-grep code scan |
| `telemetry` | `internal/cli/telemetry.go:14` (`report` sub at :25) | Skill usage telemetry |
| `version` | `internal/cli/version.go:12` | Version string |

Note: `profile_setup.go:104` defines `setup [name]` as a subcommand of `profile` (profile_setup wizard).

There is NO `cron` top-level subcommand; `CronCreate/CronDelete/CronList` are Claude Code-side deferred tools (not moai-adk-go responsibility).

### 1.4 `moai worktree` subcommand tree

`internal/cli/worktree/*.go` defines these files (each containing one or more cobra commands):

- `root.go:16` — `Use: "worktree"`, `Short: "Git worktree management"`
- `new.go` — create new worktree
- `list.go`, `remove.go`, `recover.go`, `sync.go`, `switch.go`, `go.go`, `done.go`, `clean.go`, `config.go`, `project.go`, `render.go`, `status.go`
- Companion infrastructure: `tmux_integration.go`, `errors.go`, `mock_extensions_test.go`

(Detailed flag inventory per subcommand omitted; file presence confirmed via `ls`.)

### 1.5 CLI wizard (`internal/cli/wizard/`)

- `wizard.go`, `questions.go`, `config_helpers.go`, `styles.go`, `translations.go`, `types.go`
- Used by `init` and `update` for interactive profile setup (charmbracelet/huh based).

---

## 2. internal/ packages — dependency map

29 top-level subpackages under `internal/`. Listed alphabetically with one-line purpose:

| Package | Purpose |
|---------|---------|
| `astgrep/` | ast-grep rule loader, scanner, SARIF emit (`analyzer.go`, `scanner.go`, `rules.go`, `sarif.go`) |
| `cli/` | All cobra commands + composition root (`deps.go`). 97 files. |
| `config/` | Config YAML loader + defaults + validation + envkeys (`manager.go`, `loader.go`, `types.go`, `defaults.go`, `envkeys.go`, `validation.go`) |
| `core/` | Cross-cutting domain: `git/`, `integration/`, `migration/`, `project/`, `quality/` |
| `defs/` | Path/file/directory constants (`dirs.go`, `files.go`, `paths.go`, `perms.go`) |
| `evolution/` | Design-system self-evolution (`apply.go`, `graduation.go`, `learning.go`, `safety.go`, `security.go`, `types.go`) |
| `foundation/` | TRUST5 + language definitions + error types + timeouts (`language.go`, `errors.go`, `trust/`) |
| `git/` | Branch detector + git convention (`branch_detector.go`, `convention/`) |
| `github/` | GitHub API integration (`gh.go`, `gh_client.go`, `issue_parser.go`, `issue_closer.go`, `pr_merger.go`, `pr_reviewer.go`, `spec_linker.go`) |
| `hook/` | Claude Code hook event handlers + registry (see §5) |
| `i18n/` | i18n templates (`templates.go`, `errors.go`) |
| `loop/` | Ralph loop controller (`controller.go`, `state.go`, `storage.go`, `feedback.go`, `go_feedback.go`, `feedback_channel.go`, `aggregator_feedback_test.go`) |
| `lsp/` | LSP subsystems: `aggregator`, `cache`, `config`, `core`, `gopls`, `hook`, `subprocess`, `transport` |
| `manifest/` | File provenance manifest (`manifest.go`, `hasher.go`, `types.go`) |
| `mcp/` | MoAI's own MCP server (stub implementation: `server.go`, `handler.go`, `tools.go`, `protocol.go`, `lsp_stub.go`) |
| `merge/` | 3-way merge during `moai update` (`confirm.go`, `differ.go`, `strategies.go`, `evolvable_zone.go`, `three_way.go`, `conflict.go`) |
| `profile/` | Claude profile preferences (`preferences.go`, `profile.go`, `sync.go`) |
| `ralph/` | Ralph engine (`engine.go`, `classify_lsp.go`) |
| `research/` | Research framework: `dashboard/`, `eval/`, `experiment/`, `observe/`, `safety/` |
| `resilience/` | Circuit breaker (`circuit.go`, `types.go`) |
| `shell/` | Shell env detection & config (`detect.go`, `config.go`, `env.go`, `types.go`) |
| `statusline/` | Statusline rendering (`renderer.go`, `builder.go`, `theme.go`, `metrics.go`, `git.go`, `memory.go`, `task.go`, `update.go`, `usage.go`, `version.go`, `gradient.go`, `types.go`) |
| `telemetry/` | Skill usage telemetry (`recorder.go`, `async_recorder.go`, `report.go`, `outcome.go`, `types.go`) |
| `template/` | go:embed template deployment (`embed.go`, `deployer.go`, `renderer.go`, `settings.go`, `context.go`, `model_policy.go`, `deployer_mode.go`, `templates/`) |
| `tmux/` | tmux detection + session management (`detector.go`, `session.go`) |
| `update/` | CLI self-update (`updater.go`, `checker.go`, `local.go`, `orchestrator.go`, `rollback.go`, `cache.go`) |
| `workflow/` | SPEC-ID parser + worktree orchestrator (`specid.go`, `worktree_orchestrator.go`) |

### 2.1 Composition Root (internal/cli/deps.go)

`internal/cli/deps.go:27-54` declares `Dependencies` struct binding:
- `Config *config.ConfigManager`
- `Git git.Repository`
- `GitBranch git.BranchManager`
- `GitWorktree git.WorktreeManager`
- `HookRegistry hook.Registry`
- `HookProtocol hook.Protocol`
- `UpdateChecker update.Checker`
- `UpdateOrch update.Orchestrator`
- `LoopController *loop.LoopController`
- `Logger *slog.Logger`

`InitDependencies()` at `deps.go:66+` is annotated `@MX:ANCHOR fan_in=5`. It wires every subsystem: ralph engine → loop storage → gopls bridge (optional, `deps.go:89-109`) → feedback generator → security scanner → circuit breaker → diagnostics collector → ast-grep analyzer → hook registry handler registration (see §5).

### 2.2 Notable cross-package dependencies

- `internal/cli/` depends on EVERY OTHER internal package (it is the composition root).
- `internal/hook/` depends on `internal/config/`, `internal/hook/trace`, `internal/hook/security`, `internal/hook/quality`, `internal/hook/mx`, `internal/hook/lifecycle`, `internal/hook/memo`, `internal/hook/dbsync`, `internal/hook/agents`.
- `internal/loop/` depends on `internal/ralph/`, `internal/lsp/gopls`, `internal/lsp/hook`.
- `internal/lsp/core/` uses `github.com/charmbracelet/x/powernap v0.1.4` (go.mod line 9); `internal/lsp/gopls/` is the legacy hand-rolled path (SPEC-GOPLS-BRIDGE-001).
- `internal/template/` has NO external internal/* dependencies (only imports `internal/config`, `internal/manifest`, `pkg/models`).

---

## 3. pkg/ public API surface

- `pkg/models/` — Public config/models shared with templates (`config.go`, `lang.go`, `project.go`, `doc.go`).
  - Exposes `ModeDDD`, `ModeTDD` constants (used by `internal/template/context.go:76`).
- `pkg/version/version.go` — `Version = "v2.12.0"` hard-coded default (`pkg/version/version.go:7`), overridden at build time via ldflags (see Makefile line 11). Exports `GetVersion()`, `GetCommit()`, `GetDate()`, `GetFullVersion()`.
- No other pkg/ packages exist.

---

## 4. Template system (internal/template/)

### 4.1 Embedded template filesystem

- `internal/template/embed.go:25` — `//go:embed all:templates` directive.
- `EmbeddedTemplates()` returns `fs.Sub(embeddedRaw, "templates")` (`embed.go:36-39`).
- @MX:ANCHOR fan_in=6.

### 4.2 Template asset count

- Total embedded files under `internal/template/templates/`: **502** (from `find -type f | wc -l`).
- Roots: `.claude/`, `.moai/`, `CLAUDE.md`, `.mcp.json.tmpl`, `.gitignore`.
- NO `.agency/` folder exists in templates (absorbed into `.moai/design/` per SPEC-AGENCY-ABSORB-001).

### 4.3 TemplateContext fields (`internal/template/context.go:10-59`)

Exposed fields: `ProjectName`, `ProjectRoot`, `UserName`, `ConversationLanguage`, `ConversationLanguageName`, `AgentPromptLanguage`, `GitCommitMessages`, `CodeComments`, `Documentation`, `ErrorMessages`, `GitMode`, `GitProvider`, `GitHubUsername`, `GitLabInstanceURL`, `DevelopmentMode`, `EnforceQuality`, `TestCoverageTarget`, `AutoClear`, `PlanTokens`, `RunTokens`, `SyncTokens`, `Version`, `Platform`, `InitializedAt`, `CreatedAt`, `GoBinPath`, `HomeDir`, `SmartPATH`, `ModelPolicy`.

Defaults set in `NewTemplateContext` (`context.go:64-96`).

### 4.4 Deployment mechanism (internal/template/deployer.go)

- Interface `Deployer` has `Deploy`, `ExtractTemplate`, `ListTemplates` methods (`deployer.go:21-30`).
- Factory constructors: `NewDeployer`, `NewDeployerWithRenderer`, `NewDeployerWithForceUpdate`, `NewDeployerWithRendererAndForceUpdate` (`deployer.go:43-62`).
- File provenance recorded via `manifest.Manager` (`deployer.go:65-169`). Provenance types: `TemplateManaged`, `UserCreated`, `UserModified`.
- `.tmpl` suffix files are rendered with TemplateContext and saved without `.tmpl` (`deployer.go:103-112`).
- Existing-file protection logic at `deployer.go:121-137`: `UserCreated`/`UserModified` files are NEVER overwritten (unless `forceUpdate=true`).
- Security: `validateDeployPath` blocks absolute paths and `..` traversal (`deployer.go:197-225`).

### 4.5 MergeZones / EvolvableZone

- `internal/merge/evolvable_zone.go` implements evolvable zone calculation.
- `internal/merge/strategies.go` implements merge strategies.
- Evolvable zones are part of the SPEC-AGENCY-ABSORB-001 migration; FROZEN zones (e.g., `.claude/rules/moai/design/constitution.md`) are guarded.

### 4.6 Settings.json generation (NOT template-expanded)

- `internal/template/settings.go` defines `SettingsGenerator` per ADR-011 (Zero Runtime Template Expansion).
- BUT the actual settings template at `internal/template/templates/.claude/settings.json.tmpl` IS expanded via Go template syntax (`{{- if eq .Platform "windows"}}`). Contradicts ADR-011 description in `embed.go:8-12` → investigate: the `.tmpl` suffix IS used, so the file DOES get rendered. ADR comment is stale.

### 4.7 Protected directories (rules per CLAUDE.local.md §2)

- `.claude/`, `.moai/project/`, `.moai/specs/` — never deleted during template sync.
- `.claude/settings.local.json`, `.claude/agent-memory/`, `.claude/commands/98-*.md` `99-*.md`, `.moai/cache/`, `.moai/logs/`, `.moai/state/`, `.moai/specs/`, `.moai/plans/`, `.moai/reports/`, `.moai/manifest.json`, `.moai/status_line.sh` — never in templates.

---

## 5. Hook implementations

### 5.1 Hook Registry (`internal/hook/registry.go`)

- `registry` struct (line 18): central registration + sequential dispatch.
- Default timeout: `DefaultHookTimeout = 30s` (`types.go:14`).
- Dispatch logic (`registry.go:65-167`): timeout via `context.WithTimeout`, block short-circuit via `isBlockDecision`, exit code 2 propagation for `TeammateIdle`/`TaskCompleted`.
- TraceWriter lazy initialization on first Dispatch (`registry.go:252-264`) — REQ-OBS-001.
- @MX:ANCHOR fan_in=20+.

### 5.2 Hook event types (internal/hook/types.go)

27 event constants defined (`types.go:19-114`):
`SessionStart, PreToolUse, PostToolUse, SessionEnd, Stop, SubagentStop, PreCompact, PostToolUseFailure, Notification, SubagentStart, UserPromptSubmit, PermissionRequest, TeammateIdle, TaskCompleted, WorktreeCreate, WorktreeRemove, PostCompact, InstructionsLoaded, StopFailure, Setup, ConfigChange, TaskCreated, CwdChanged, FileChanged, Elicitation, ElicitationResult, PermissionDenied`

Version gating in comments marks each event's Claude Code minimum version (v2.1.10 through v2.1.89).

### 5.3 Registered handlers in `InitDependencies()` (deps.go:151-186)

Handlers explicitly registered (28 calls):
- `NewSessionStartHandler(deps.Config)`
- `buildSessionEndHandler(cwd)` (observability-aware)
- `NewAutoUpdateHandler(buildAutoUpdateFunc())` (also SessionStart)
- `NewStopHandler`
- `NewPreToolHandlerWithScanner(deps.Config, secPolicy, securityScanner)` (PreToolUse)
- `NewPostToolHandlerWithMxValidatorAndTimeout(diagnosticsCollector, astAnalyzer, cwd, 500ms)` (PostToolUse)
- `NewCompactHandler`, `NewPostToolUseFailureHandler`, `NewNotificationHandler`, `NewSubagentStartHandler`, `NewUserPromptSubmitHandler(deps.Config)`, `NewPermissionRequestHandler`, `NewTeammateIdleHandler`, `NewTaskCompletedHandler`, `NewWorktreeCreateHandler`, `NewWorktreeRemoveHandler`, `NewPostCompactHandler`, `NewInstructionsLoadedHandler`, `NewStopFailureHandler`, `NewSubagentStopHandler`, `NewTaskCreatedHandler`, `NewPermissionDeniedHandler`, `NewConfigChangeHandler`, `NewCwdChangedHandler`, `NewFileChangedHandler`, `NewElicitationHandler`, `NewElicitationResultHandler`, `NewSetupHandler`.

28 handlers registered vs. 27 declared event types. The `AutoUpdateHandler` ALSO listens on SessionStart (shared event).

### 5.4 CLI hook dispatch (`internal/cli/hook.go`)

`hook.go:28-60` declares 27 hook subcommands mapped to `EventType`:
`session-start, pre-tool, post-tool, session-end, stop, compact, post-tool-failure, notification, subagent-start, user-prompt-submit, permission-request, teammate-idle, task-completed, subagent-stop, worktree-create, worktree-remove, post-compact, instructions-loaded, stop-failure, setup, config-change, task-created, cwd-changed, file-changed, elicitation, elicitation-result, permission-denied`.

Additional non-event subcommands: `hook list` (`:75`), `hook agent [action]` (`:82-88`), `hook db-schema-sync` (`:91-98`).

### 5.5 Hook wrapper shell scripts

Local hooks at `.claude/hooks/moai/`: **26** `.sh` scripts (one per event). Template source at `internal/template/templates/.claude/hooks/moai/`: **27** `.sh.tmpl` files (including `handle-permission-denied.sh.tmpl` which is missing from local — possibly by design since local was deployed before v2.1.89 support).

Standard wrapper pattern: each script reads stdin JSON and calls `moai hook <event>`.

### 5.6 Hook subpackages (internal/hook/)

Subpackages with specialized behavior:
- `agents/` — per-agent hook handlers (backend, ddd, debug, devops, docs, frontend, quality, spec, tdd, testing). Factory pattern (`factory.go`).
- `dbsync/` — SPEC-DB-SYNC-001 schema sync (`db_schema_sync.go` is 20KB, biggest non-registry handler).
- `lifecycle/` — session lifecycle persistence (`cleanup.go`, `persistence.go`, `persistent_mode.go`).
- `memo/` — priority-based memo reader/writer.
- `mx/` — MX tag validator (`validator.go` is 13KB; validates auto-generated tags).
- `quality/` — quality gates (`gate.go` 20KB, `astgrep_gate.go`, `linter.go`, `lint_instruction.go`, `change_detector.go`, `tool_registry.go`).
- `security/` — AST-based security scanner (`ast_grep.go`, `scanner.go`, `rules.go`).
- `trace/` — observability JSONL writer.

### 5.7 Hook output schema (types.go:167-311)

- `HookInput` carries 30+ possible fields depending on event (SessionID, ToolName, ToolInput, ToolOutput, ToolResponse, ToolUseID, Source, Model, AgentType, Reason, StopHookActive, AgentID, AgentTranscriptPath, LastAssistantMessage, Trigger, CustomInstructions, Error, IsInterrupt, ErrorType, ErrorMessage, Prompt, Message, Title, NotificationType, ProjectDir, TeamName, TeammateName, TaskID, TaskSubject, TaskDescription, WorktreePath, WorktreeBranch, AgentName, ConfigFilePath, ConfigurationSource, FilePath, ChangeType, OldCwd, NewCwd, ElicitationServerName, MCPToolName, ElicitationRequest, InstructionFilePath, MemoryType, LoadReason, Globs, TriggerFilePath, ParentFilePath, PermissionSuggestions).
- `HookOutput` (types.go:284-311) carries: `Continue, StopReason, SystemMessage, SuppressOutput, Decision, Reason, HookSpecificOutput, UpdatedInput, Retry, ExitCode`.
- Factory helpers (`types.go:313-471`): `NewAllowOutput, NewDenyOutput, NewAskOutput, NewBlockOutput, NewSuppressOutput, NewSessionOutput, NewPostToolOutput, NewStopBlockOutput, NewPostToolBlockOutput, NewPermissionRequestOutput, NewUserPromptBlockOutput, NewDeferOutput, NewPermissionDeniedRetryOutput, NewTeammateKeepWorkingOutput, NewTaskRejectedOutput`.

### 5.8 Missing events vs CLAUDE.md reference

CLAUDE.md §7/§14 references events that map to what's currently implemented. All 27 types defined in `types.go` appear to have wrappers. One discrepancy: local `.claude/hooks/moai/` is missing `handle-permission-denied.sh` — the template has it, so the local project is lagging.

---

## 6. Agent catalog (counts, categories)

### 6.1 Local vs template parity

- `.claude/agents/moai/` (local): **22** `.md` files.
- `internal/template/templates/.claude/agents/moai/` (template): **22** `.md` files.
- Parity is clean — every local agent has a template counterpart. Small byte-diff: `builder-agent.md` local 5460 / template 5473 (13 bytes), `expert-refactoring.md` local 3693 / template 3706, `expert-security.md` local 5118 / template 5131, `evaluator-active.md` local 4817 / template 4830, `manager-spec.md` local 5866 / template 5880, `manager-strategy.md` local 5564 / template 5578, `plan-auditor.md` local 15184 / template 15197. Other 15 files match byte-for-byte.

### 6.2 Catalog by category

**Managers (8):** manager-ddd, manager-docs, manager-git, manager-project, manager-quality, manager-spec, manager-strategy, manager-tdd

**Experts (8):** expert-backend, expert-debug, expert-devops, expert-frontend, expert-performance, expert-refactoring, expert-security, expert-testing

**Builders (3):** builder-agent, builder-plugin, builder-skill

**Evaluators (2):** evaluator-active, plan-auditor

**Researcher (1):** researcher

**Total: 22 agents** (matches CLAUDE.md §4: "Manager Agents (8) + Expert Agents (8) + Builder Agents (3) + Evaluator Agents (2) + Agency retained as fallback").

Note: agency agents (copywriter, designer, planner, builder, evaluator, learner) were REMOVED per SPEC-AGENCY-ABSORB-001 M5. Copywriter/designer are now skills (`moai-domain-copywriting`, `moai-domain-brand-design`). The `researcher` is new (v2.12 era).

NOT present: Claude Code's built-in agents (Explore, Plan, general-purpose, statusline-setup, claude-code-guide) — these come from Claude Code itself, not from moai-adk-go.

### 6.3 Agent frontmatter fields in use

Observed via Grep across 22 files:

- `name:` — all 22
- `description:` — all 22 (multi-language keyword blocks EN/KO/JA/ZH for MUST-INVOKE matching)
- `tools:` — all 22 (CSV string, includes MCP tool names)
- `model:` — all 22. Distribution:
  - `opus`: expert-security, manager-spec, manager-strategy, researcher (4 agents)
  - `sonnet`: 17 agents (bulk of managers/experts/builders + evaluator-active)
  - `haiku`: manager-docs, manager-git (2)
  - `inherit`: plan-auditor (1)
  - NOTE: `inherit` lets the caller decide; ideal for plan-auditor's adversarial role.
- `permissionMode:` — all 22. Distribution:
  - `bypassPermissions`: 20 agents (all write-capable agents)
  - `plan`: evaluator-active (read-only)
  - `default`: plan-auditor
- `memory:` — all 22. `user` for builders (3); `project` for rest (19).
- `skills:` — all 22 reference bundled skills as YAML array (one of the few YAML-array frontmatter fields).
- `hooks:` — present on domain-specific agents (expert-backend has PreToolUse/PostToolUse with `handle-agent-hook.sh backend-validation/verification`). Confirmed on expert-backend.md:23-35.

No agent uses `effort:`, `background:`, or `isolation:` in frontmatter (those are spawn-time parameters per CLAUDE.md §14 `Agent(isolation: "worktree")`).

### 6.4 Orphaned/deprecated agents

- None in current catalog.
- Legacy `.agency/` agent files (copywriter, designer, planner, builder, evaluator, learner) were removed per SPEC-AGENCY-ABSORB-001 M5; their absorption path: copywriter → `moai-domain-copywriting` skill; designer → `moai-domain-brand-design` skill.

---

## 7. Skill catalog (counts, categories)

### 7.1 Local vs template parity

- `.claude/skills/` (local): 47 skill dirs (including `moai/` entry skill).
- `internal/template/templates/.claude/skills/` (template): 50 skill dirs.

**Template-only (3 additions vs local):**
- `moai-domain-db-docs/` (template only)
- `moai-workflow-design-context/` (template only)
- `moai-workflow-pencil-integration/` (template only)

These are forward-looking skills not yet copied to local. The local project is running v2.12.0 templates; new db/design skills were added in the v2.12 cycle.

### 7.2 Skills by category (template, 50 total)

**`moai` (orchestrator skill, 1):** `moai/` (entry point — routes to workflows)

**`moai-foundation-*` (5):** foundation-cc, foundation-context, foundation-core, foundation-philosopher, foundation-quality, foundation-thinking

**`moai-workflow-*` (13):** workflow-ddd, workflow-design-context (template-only), workflow-design-import, workflow-gan-loop, workflow-jit-docs, workflow-loop, workflow-pencil-integration (template-only), workflow-project, workflow-research, workflow-spec, workflow-tdd, workflow-templates, workflow-testing, workflow-thinking, workflow-worktree

**`moai-domain-*` (7):** domain-backend, domain-brand-design, domain-copywriting, domain-database, domain-db-docs (template-only), domain-frontend, domain-uiux

**`moai-platform-*` (4):** platform-auth, platform-chrome-extension, platform-database-cloud, platform-deployment

**`moai-framework-*` (1):** framework-electron

**`moai-library-*` (3):** library-mermaid, library-nextra, library-shadcn

**`moai-formats-*` (1):** formats-data

**`moai-tool-*` (2):** tool-ast-grep, tool-svg

**`moai-ref-*` (5):** ref-api-patterns, ref-git-workflow, ref-owasp-checklist, ref-react-patterns, ref-testing-pyramid

**`moai-design-*` (2):** design-craft, design-tools

**`moai-docs-*` (1):** docs-generation

Total matches: 1+6+14+7+4+1+3+1+2+5+2+1 = 47 (if using the 47 local count) — but local lacks the 3 template-only.

### 7.3 Skill frontmatter fields

Observed via Grep:
- `name:` — all skills
- `description:` — all skills (in YAML frontmatter, triggers skill activation)
- `allowed-tools:` — all skills. CSV string format, often 8-20 Bash command-scoped tools like `Bash(npm:*), Bash(pytest:*)`.
- `paths:` — not widely used in current skills; CLAUDE.local.md notes it's optional.

Not observed (not in use): `skills:` on skill frontmatter (only on agents).

### 7.4 Progressive Disclosure Level 2 bundled resources

- Total skill markdown files: **320** (via `find`).
- Top-level `SKILL.md` files: ~50. So ~270 bundled module/reference files.
- Bundled folder structures observed:
  - `moai/references/` (mx-tag.md, reference.md) — 2 files
  - `moai/team/` (debug.md, glm.md, plan.md, review.md, run.md, sync.md) — 6 files
  - `moai/workflows/` (21 files: clean, codemaps, context, coverage, design, e2e, feedback, fix, gate, github, loop, moai, mx, plan, project, review, run, security, sync; plus more)
  - `moai-foundation-core/modules/` (18 files)
  - `moai-foundation-core/references/` (directory)
  - `moai-workflow-spec/references/` (examples.md 28KB, reference.md 17KB), `moai-workflow-spec/modules/`, `moai-workflow-spec/reference/`

Progressive Disclosure Level 2 content follows the pattern from CLAUDE.md §13: skill body ~5K tokens, bundled resources on-demand.

---

## 8. Command catalog

### 8.1 Local vs template

- `.claude/commands/moai/`: 13 `.md` files (clean, codemaps, coverage, e2e, feedback, fix, loop, mx, plan, project, review, run, sync). Each is a thin router.
- `internal/template/templates/.claude/commands/moai/`: 14 `.md` / `.md.tmpl` files (adds `db.md`, which is the v2.12 DB docs command).
- `.claude/commands/agency/`: 8 `.md` files (deprecated) — matches template.
- `.claude/commands/`: 2 extra ("98-github.md", "99-release.md" — developer-project commands, not in template per CLAUDE.local.md rule).

### 8.2 Thin command pattern compliance

Spot-check on 2 files:
- `plan.md`: 7 lines body, just `Use Skill("moai") with arguments: plan $ARGUMENTS`.
- `run.md`: 7 lines body, same router pattern.
- All under 20 LOC body.
- Pattern: YAML frontmatter `description`, `argument-hint`, `allowed-tools: Skill`.
- Enforced by `internal/template/commands_audit_test.go` per CLAUDE.local.md §3 reference.

### 8.3 Root commands

- `98-github.md` (21,950 bytes) — full GitHub workflow command (developer-only).
- `99-release.md` (22,815 bytes) — release workflow command (developer-only).

Per CLAUDE.local.md §2 they must NOT appear in templates (confirmed).

### 8.4 Agency commands (deprecated)

`.claude/commands/agency/`: `agency.md` (1053), `brief.md` (272), `build.md` (277), `evolve.md` (534), `learn.md` (524), `profile.md` (267), `resume.md` (275), `review.md` (260). All redirect to `/moai design` per SPEC-AGENCY-ABSORB-001 REQ-DEPRECATE-003.

---

## 9. Rules hierarchy (.claude/rules/moai/)

### 9.1 Hierarchy

Identical in local and template:

- `moai/core/` (6 files): `agent-common-protocol.md`, `agent-hooks.md`, `hooks-system.md`, `lsp-client.md`, `moai-constitution.md`, `settings-management.md`.
- `moai/workflow/` (7 files): `file-reading-optimization.md`, `moai-memory.md`, `mx-tag-protocol.md`, `spec-workflow.md`, `team-protocol.md`, `workflow-modes.md`, `worktree-integration.md`.
- `moai/development/` (4 files): `agent-authoring.md`, `coding-standards.md`, `model-policy.md`, `skill-authoring.md`.
- `moai/languages/` (16 files): cpp, csharp, elixir, flutter, go, java, javascript, kotlin, php, python, r, ruby, rust, scala, swift, typescript.
- `moai/design/` (1 file): `constitution.md` (17,626 bytes; FROZEN/EVOLVABLE zone spec).
- `agency/constitution.md` (stub redirect to `moai/design/constitution.md`).

Total rules: 35 files. (6 + 7 + 4 + 16 + 1 + 1 = 35).

### 9.2 One-liners for key rules

- `core/moai-constitution.md` — Core HARD rules, Agent Core Behaviors (6 cross-cutting), Parallel Execution, Opus 4.7 prompt philosophy, Lessons protocol.
- `core/agent-common-protocol.md` — Shared protocol for all agents (User Interaction Boundary, Language Handling, Output Format, MCP Fallback, Background Agent Execution restrictions).
- `core/agent-hooks.md` — Agent-level hook declaration pattern (`handle-agent-hook.sh backend-validation`).
- `core/hooks-system.md` — 15KB deep dive on hook system internals.
- `core/lsp-client.md` — Rationale for powernap v0.1.4 vs alternatives (charmbracelet/crush uses same library at 23k stars).
- `core/settings-management.md` — settings.json schema + .mcp.json format.
- `workflow/spec-workflow.md` — Plan-Run-Sync phase definitions, token budgets.
- `workflow/workflow-modes.md` — DDD/TDD methodology selection.
- `workflow/worktree-integration.md` — Worktree decision tree, Claude Native vs MoAI Worktree.
- `workflow/team-protocol.md` — Team coordination (TeamCreate, SendMessage, TaskCreate).
- `workflow/mx-tag-protocol.md` — @MX tag syntax + lifecycle.
- `workflow/file-reading-optimization.md` — Tiered file reading per size.
- `workflow/moai-memory.md` — Auto-memory (~/.claude/projects/.../memory/MEMORY.md).
- `development/agent-authoring.md` — Agent frontmatter schema + required fields.
- `development/skill-authoring.md` — Skill frontmatter schema.
- `development/coding-standards.md` — Language policy, file-size limits, thin command pattern.
- `development/model-policy.md` — opus/sonnet/haiku assignment rules.
- `design/constitution.md` — FROZEN/EVOLVABLE zones for agency-absorbed design system, v3.3.0.

---

## 10. Configuration schema (.moai/config/sections/)

### 10.1 Full inventory (22 YAML files)

All under `.moai/config/sections/`:

| File | Size | Key keys |
|------|------|---------|
| `constitution.yaml` | 876B | `approved_frameworks`, `approved_languages`, `forbidden_patterns`, `naming_conventions` |
| `context.yaml` | 531B | `context_search.enabled`, `search.max_results`, `memory_integration` |
| `design.yaml` | 2763B | `gan_loop.max_iterations=5`, `pass_threshold=0.75`, `sprint_contract`, `claude_design.supported_bundle_versions=["1.0"]`, `evolution`, `adaptation` |
| `git-convention.yaml` | 530B (0600) | git commit conventions |
| `git-strategy.yaml` | 2001B | git branch/PR strategy |
| `harness.yaml` | 5182B | `default_profile`, `mode_defaults`, `auto_detection`, `escalation`, `effort_mapping`, `levels` (minimal/standard/thorough), `model_upgrade_review` |
| `interview.yaml` | 264B | onboarding interview flags |
| `language.yaml` | 200B (0600) | `conversation_language: ko`, `code_comments: ko`, `error_messages: en` |
| `llm.yaml` | 476B (0600) | `mode`, `team_mode`, `glm.base_url`, `glm.models{high,medium,low,opus,sonnet,haiku}`, `default_model`, `quality_model`, `speed_model` |
| `lsp.yaml` | 8098B | 16-language server matrix, `lsp.enabled=false`, `client_impl=gopls_bridge` |
| `mx.yaml` | 8545B | `auto_tag=true`, `exclude` (vendor, node_modules, build, etc.), per-language `warn_patterns` |
| `project.yaml` | 153B | `name`, `mode=personal`, `template_version=v2.7.22`, `initialized=true` |
| `quality.yaml` | 2259B | `development_mode=tdd`, `test_coverage_target=85`, `lsp_quality_gates`, `ast_grep_gate`, `principles.simplicity.max_parallel_tasks=10`, `trust5_integration` |
| `ralph.yaml` | 1101B | Ralph loop engine config |
| `research.yaml` | 680B | research framework flags |
| `security.yaml` | 1023B | security scan rules |
| `state.yaml` | 34B | `state_dir: .moai/state` |
| `statusline.yaml` | 288B | `mode: default` |
| `sunset.yaml` | 820B | deprecation markers |
| `system.yaml` | 1067B | `moai.version=v2.12.0`, `template_version=v2.12.0`, `update_check_frequency=daily`, `document_management.directories`, `github` |
| `user.yaml` | 27B (0600) | `user.name: GOOS행님` |
| `workflow.yaml` | 3965B | `auto_clear`, `completion markers`, `execution_mode=auto`, `loop_prevention`, `team.enabled`, `team.max_teammates=10`, `team.default_model=opus[1m]`, `role_profiles` (researcher/analyst/architect/implementer/tester/designer/reviewer), `patterns`, `token_budget`, `worktree` |

### 10.2 Evaluator profiles (`.moai/config/evaluator-profiles/`)

4 profile Markdown files: `default.md`, `frontend.md`, `lenient.md`, `strict.md`.

### 10.3 ast-grep rules (`.moai/config/astgrep-rules/`)

20 rule files (directory listed). Language-scoped ast-grep patterns for security scanning.

### 10.4 Template source (`internal/template/templates/.moai/config/`)

`ls` confirms: `config/`, `design/`, `docs/`, `evolution/`, `learning/`, `project/`, `state/`, `status_line.sh.tmpl`. Mirrors production but with template variables.

### 10.5 Schema invariants (cross-file)

- `language.yaml.code_comments` governs @MX tag description language (see rules/moai/workflow/mx-tag-protocol.md §Language Settings).
- `quality.yaml.lsp_quality_gates.run.max_errors=0` is a HARD quality gate per-phase.
- `harness.yaml.effort_mapping` maps harness level → Claude Code `CLAUDE_CODE_EFFORT_LEVEL` (medium/high/xhigh).
- `workflow.yaml.team.role_profiles.*.isolation` controls agent worktree spawning (`none` vs `worktree`).

---

## 11. SPEC system state

### 11.1 Active SPEC count

- `.moai/specs/` contains **104** directories.
- Naming convention: `SPEC-<DOMAIN>-<NNN>` (zero-padded) or `SPEC-<DOMAIN>-<SUBDOMAIN>-<NNN>`.
- Notable meta-SPECs: `UPGRADE-HARNESS-DESIGN/` (not SPEC-prefixed; legacy folder).

### 11.2 Domain distribution (by prefix)

Sampled from the directory listing:

- `AGENCY` (1 active + 1 absorbed)
- `AGENT` (2)
- `DB-*` (4): DB-CMD, DB-SYNC, DB-SYNC-HARDEN, DB-SYNC-RELOC, DB-TEMPLATES, PROJECT-DB-HINT (5 total)
- `DESIGN-*` (4): DESIGN, DESIGN-ATTACH, DESIGN-CONST-AMEND, DESIGN-DOCS, DESIGN-PENCIL (5 total)
- `DOCS-*` (2): DOCS-SB-REMOVE, DOCS-SITE
- `HOOK-*` (9): HOOK-001 .. HOOK-009
- `LSP-*` (6): LSP-001, LSP-AGG-003, LSP-CORE-002, LSP-LOOP-005, LSP-MULTI-006, LSP-QGATE-004, LSPMCP-001
- `MX-*` (2), `SKILL-*` (4), `SPEC-SRS-*` (3), `SPEC-UI-*` (3), `SPEC-WORKTREE-*` (2)
- + many others: TEAM-001, TELEMETRY-001, TEMPLATE-001, THIN-CMDS-001, TOKEN-001, TOOL-001, UPDATE-001, UPDATE-002, CI/CD, MEMO, OBSERVE, ORCH, PERSIST, PLAYWRIGHT, PSR, QUALITY, REFACTOR, REFLECT, SDD, SEARCH, SECURITY, SEMAP, STATUSLINE, SUNSET, etc.

### 11.3 Sample SPEC structure (SPEC-CORE-001)

`.moai/specs/SPEC-CORE-001/` contains:
- `spec.md` (20,536 bytes)
- `plan.md` (9,819 bytes)
- `acceptance.md` (20,411 bytes)

Triple-file pattern: `spec.md` (what/why), `plan.md` (how), `acceptance.md` (verification criteria).

### 11.4 EARS format confirmation

SPEC-CORE-001/spec.md YAML frontmatter (`spec_id`, `title`, `status`, `priority`, `phase`, `module`, `estimated_loc`, `dependencies`, `assigned`, `lifecycle`, `tags`) — confirmed to follow the SPEC schema.

Body sections: `HISTORY`, `Environment`, `Assumptions`, `Requirements`, etc. — EARS structure observed.

### 11.5 SPEC lifecycle tags (from `tags:` frontmatter)

Examples observed: `foundation, ears, languages, trust5, domain-patterns` (SPEC-CORE-001), `go-embed, templates, bundling, binary, deployment, init` (SPEC-EMBED-001).

### 11.6 Completed vs in-progress

Random sample: SPEC-CORE-001 has `status: completed`. SPEC-EMBED-001 has `status: completed`. Most v2.x-era SPECs appear to be completed; active work is concentrated in newer SPECs (DB, DESIGN, DOCS-SITE, OPUS47-COMPAT).

---

## 12. Docs site state (docs-site/adk.mo.ai.kr)

### 12.1 Hugo + Hextra

- `docs-site/hugo.yaml` line 1: `baseURL: "https://adk.mo.ai.kr/"`.
- Theme: Hextra (via Hugo modules, `module.imports[0].path: github.com/imfing/hextra`).
- `defaultContentLanguage: ko` (line 9), `defaultContentLanguageInSubdir: true`.
- `enableGitInfo: true`, `enableRobotsTXT: true`, `hasCJKLanguage: true`.

### 12.2 4-locale state

`docs-site/content/` has 4 locale dirs: `ko/`, `en/`, `ja/`, `zh/` + a `.claude/` dir for Claude Code integration.

Locale structure (ko):
- 13 section dirs: `advanced/`, `contributing/`, `core-concepts/`, `db/`, `design/`, `getting-started/`, `multi-llm/`, `quality-commands/`, `utility-commands/`, `workflow-commands/`, `worktree/`.
- `_index.md`, `_meta.yaml` at root.

Locale structure (en):
- 11 section dirs (matches most of ko; lacks `contributing/`, `multi-llm/`).
- Possible translation lag.

Local structure (ja, zh): checked folder counts similar to en (not inspected in detail).

### 12.3 Vercel deployment

- `docs-site/vercel.json` present.
- Project ID per CLAUDE.local.md §17.6: `prj_EZaVdfE3gJeXVbizafBEECpniINP`.
- Domain: `adk.mo.ai.kr` (production branch).

### 12.4 Scripts

`docs-site/scripts/`:
- `fix-mdx-formatting.js` (3.5K)
- `generate-favicon-icons.md` (1.4K)
- `generate-favicons.js` (2.8K)
- `README_TRANSLATE.md` (4.8K)
- `translate.mjs` (16.4K) — AI-driven 4-locale translation

Top-level `scripts/`:
- `convert-nextra-to-hextra/`
- `docs-version-snapshot/` (manager-git release hook per CLAUDE.local.md §17.4)

### 12.5 URL standard (per CLAUDE.local.md §17.1)

Only `adk.mo.ai.kr` is allowed. Blacklisted URLs: `docs.moai-ai.dev` (old), `adk.moai.com` (typo), `adk.moai.kr` (typo).

---

## 13. Tests & coverage

### 13.1 Test function count

Total `func Test...` across `internal/` and `pkg/`: **3762**.

### 13.2 By major package

- `internal/cli/`: 1142 test functions
- `internal/hook/`: 733 test functions
- `internal/template/`: 94 test functions

Coverage data (stale, from 2026-03-11 coverage.out): `coverage.out` 16K, `coverage.html` 77K — not reliable for v2.12 state; need `make coverage` to refresh.

### 13.3 Makefile targets

- `make build` → `go build -ldflags ... -o bin/moai ./cmd/moai`
- `make test` → `go test -race -coverprofile=coverage.out -covermode=atomic ./...`
- `make test-verbose`, `make coverage`, `make lint` (golangci-lint), `make fix` (go fix modernizers, twice), `make vet`, `make fmt` (gofumpt), `make generate`, `make clean`, `make tidy`, `make run`, `make help`
- `make release-local` → copy to `$HOME/.moai/releases/moai-v<ver>-<os>-<arch>`.

### 13.4 CI hints

- `.github/workflows/` likely contains CI definitions (not inspected).
- `.goreleaser.yml` present (goreleaser-based releases).
- `install.sh`, `install.bat`, `install.ps1` → cross-platform installers.

---

## 14. External dependencies

### 14.1 Direct dependencies (go.mod lines 5-15)

1. `github.com/charmbracelet/bubbletea v1.3.10` — TUI framework
2. `github.com/charmbracelet/huh v0.8.0` — Interactive forms
3. `github.com/charmbracelet/lipgloss v1.1.1-0.20250404203927-76690c660834` — Styling
4. `github.com/charmbracelet/x/powernap v0.1.4` — LSP client (SPEC-LSP-CORE-002)
5. `github.com/mattn/go-isatty v0.0.21` — TTY detection
6. `github.com/spf13/cobra v1.10.2` — CLI framework
7. `golang.org/x/sync v0.20.0` — Concurrency primitives
8. `golang.org/x/text v0.36.0` — Text handling
9. `gopkg.in/yaml.v3 v3.0.1` — YAML parsing

Only 9 direct dependencies. Intentionally minimal.

### 14.2 Indirect dependencies (23 listed in go.mod lines 17-45)

Includes: `atotto/clipboard, aymanbagabas/go-osc52, catppuccin/go, charmbracelet/bubbles, charmbracelet/colorprofile, charmbracelet/x/ansi, charmbracelet/x/cellbuf, charmbracelet/x/term, clipperhouse/*, erikgeiser/coninput, inconshreveable/mousetrap, lucasb-eyer/go-colorful, mattn/go-localereader, mattn/go-runewidth, mitchellh/hashstructure/v2, muesli/ansi, muesli/cancelreader, muesli/termenv, rivo/uniseg, sourcegraph/jsonrpc2, spf13/pflag, xo/terminfo, golang.org/x/sys, dustin/go-humanize`.

### 14.3 Notable choices

- `github.com/sourcegraph/jsonrpc2 v0.2.1` (indirect) — Used by powernap internally; direct use is forbidden per `.claude/rules/moai/core/lsp-client.md` §Technical Constraints.
- No `logrus`, `zap` — uses `log/slog` stdlib.
- No `testify` — uses pure `testing` package.
- No HTTP framework — only uses stdlib `net/http` (not seen in imports).

---

## 15. Open weaknesses / known gaps self-identified

### 15.1 From CLAUDE.local.md (developer's own list)

§2: Protected directories rule — if templates and local drift, user data can be lost. Enforcement relies on `forceUpdate=false` default in deployer.

§3.1 Local vs template drift:
- Local skills: 47 dirs; template: 50 dirs (3 template-only: `moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`).
- Local hook wrapper scripts: 26; template: 27 (missing `handle-permission-denied.sh` locally).

### 15.2 Stale coverage artifacts

- `coverage.out` dated Mar 11 (2026-03-11) — predates many SPEC changes. Actual coverage unknown without re-run.
- `coverage.html` present but stale.

### 15.3 ADR-011 comment drift

`internal/template/embed.go:8-12` claims "Runtime-generated files (settings.json, .mcp.json, .lsp.json) are intentionally excluded from the embedded templates". BUT `.claude/settings.json.tmpl` IS in the embedded template tree and IS rendered at runtime. Comment appears to be stale since `.tmpl` rendering was added.

### 15.4 `moai-docs-generation` skill

Local template counts: 47 vs 50 suggests template has 3 more skills. The local project is running v2.12.0 templates per `project.yaml:template_version: v2.7.22` (STALE — system.yaml shows v2.12.0).

### 15.5 Hooks not yet registered?

- `InitDependencies()` registers 28 handler calls but only 27 EventType constants exist. `AutoUpdateHandler` is a second handler on SessionStart (compose pattern).
- Missing: no explicit `NewPreCompactHandler` registration — lookup: `NewCompactHandler` handles `PreCompact` (confirmed from `compact.go` existence).

### 15.6 `.agency/` vestigial folder

- `.moai-backups/` folder exists at project root (from `ls` at §1 preamble) — backup snapshot from April migration.
- `.claude/commands/agency/*.md` still deployed but redirect-only.
- `.claude/rules/agency/constitution.md` is a 695-byte stub redirect.

### 15.7 Hook script permission missing

Local `.claude/hooks/moai/` lacks `handle-permission-denied.sh`. Template source has `handle-permission-denied.sh.tmpl`. Local project was last deployed before the v2.1.89 PermissionDenied event support landed. `moai update` should resolve.

### 15.8 SPEC template_version drift

- `.moai/config/sections/project.yaml:template_version: v2.7.22` — very old (~12 minor versions behind).
- `.moai/config/sections/system.yaml:moai.version: v2.12.0` — current.
- Suggests `project.yaml.template_version` is not auto-updated by `moai update` — potential bug or design choice worth clarifying.

### 15.9 No `cron`, `EnterWorktree`, `ExitWorktree`, `Team*` commands in moai CLI

These are CLAUDE CODE-side tools exposed via deferred tool search (system-reminder listed them). moai-adk-go does NOT implement these; instead, it relies on Claude Code's runtime.

### 15.10 Legacy `.go.bak` files

- `internal/cli/glm.go.bak` (28,567 bytes) — pre-April refactor of glm.go. Should be removed.
- `internal/cli/worktree/new_test.go.bak` (13,700 bytes) — pre-April test backup.

---

## 16. Version & release state

- Current tag: `v2.12.0-15-gad2019683` (from `git describe --tags`, line shows 15 commits past v2.12.0).
- `pkg/version/version.go:7`: `Version = "v2.12.0"` (default, build-time overridden).
- `.moai/config/sections/system.yaml`: `moai.version: v2.12.0`, `moai.template_version: v2.12.0`.
- Makefile line 5: `VERSION ?= $(shell git describe --tags --abbrev=0 ...)`.
- Goreleaser configured (`.goreleaser.yml`) — but release cycle depends on git tags.
- Recent commits (git log --oneline -20):
  - `ad2019683` — chore: post-merge cleanup #695
  - `8682ce904` — docs(sync): #677 stale /moai context + #676 GLM concurrency
  - `de655f934` — chore(deps): bump powernap
  - `f56d70d4b` — docs(readme): v2.12.0 post-sync 4-language
  - `eaf3fe6c3` — docs(docs-site): db section 4-language
  - `6a9c5a2a7`, `283c21295` — CI(claude-review): max-turns tuning
  - `0ab040e5d` — fix(lsp): TypeScript + 15 languages detection
  - `99c3d93fe` — fix(cli): prevent `moai glm` from polluting settings.local.json
  - `4271fd8a8` — feat: SPEC-AGENCY-ABSORB-001 /agency → /moai design absorption
  - `5e9f61a3b` — feat(docs-site): Hextra-based monorepo integration
  - `07525c7ae` — SPEC-OPUS47-COMPAT-001 Opus 4.7 prompt philosophy
  - `a44998324` — fix #640 GLM team 401

### 16.1 CG / GLM integration

- `internal/cli/cg.go`, `internal/cli/glm.go`, `internal/cli/cc.go` — three launch modes.
- `internal/cli/launcher.go` — shared launcher infrastructure (23K, biggest launcher).
- `glm.go.bak` exists — indicates recent refactor; current `glm.go` is 31K.
- `glm_compat_test.go`, `glm_team_test.go`, `glm_model_override_test.go` — thorough test coverage.
- `oauth_token_preservation_test.go` — protects OAuth tokens during env injection.

### 16.2 Pencil MCP integration

- `expert-frontend.md` frontmatter declares `mcp__pencil__*` tools (batch_design, batch_get, get_editor_state, get_guidelines, get_screenshot, get_style_guide, get_style_guide_tags, get_variables, set_variables, open_document, snapshot_layout, find_empty_space_on_canvas, search_all_unique_properties, replace_all_matching_properties) — 14 pencil tools.
- `moai-design-tools` skill declares same set.
- `pencil/` directory at project root — exists, contents not inspected.
- `.mcp.json` does NOT list `pencil` (uses context7, sequential-thinking, chrome-devtools) — pencil likely configured at user level in `~/.claude/` and not project-scoped.

### 16.3 Registered MCP servers in `.mcp.json`

1. `context7` — `npx @upstash/context7-mcp@latest`
2. `sequential-thinking` — `npx @modelcontextprotocol/server-sequential-thinking`
3. `chrome-devtools` — `npx chrome-devtools-mcp@latest --headless` (for docs-site E2E testing)

`staggeredStartup.enabled=true`, `delayMs=500`, `connectionTimeout=15000`.

---

## 17. Source References (file:line)

Master reference list of all non-trivial file:line anchors cited above.

### Entry points
- `cmd/moai/main.go:12` — Sole binary entry
- `internal/cli/root.go:13-76` — Root cobra command
- `internal/cli/deps.go:27-54` — Dependencies struct
- `internal/cli/deps.go:66-190` — InitDependencies composition root

### CLI subcommands
- `internal/cli/init.go:26` — `moai init`
- `internal/cli/update.go:66` — `moai update`
- `internal/cli/doctor.go:39` — `moai doctor`
- `internal/cli/version.go:12` — `moai version`
- `internal/cli/hook.go:17, 28-60, 75, 82, 91` — `moai hook` + subs
- `internal/cli/glm.go:20, 61, 68, 82` — `moai glm` + setup/status
- `internal/cli/cg.go:8` — `moai cg`
- `internal/cli/cc.go:13` — `moai cc`
- `internal/cli/loop.go:11, 39, 47, 55, 63, 71` — `moai loop` + subs
- `internal/cli/github.go:66, 80, 157` — `moai github` + subs
- `internal/cli/lsp_doctor.go:54, 62, 121` — `moai lsp doctor`
- `internal/cli/mcp.go:15, 21, 32` — `moai mcp lsp`
- `internal/cli/migrate_agency.go:569, 576, 595` — `moai migrate agency`
- `internal/cli/research.go:16, 30, 72, 84, 145` — `moai research` + subs
- `internal/cli/profile.go:11, 25, 32, 38, 50` — `moai profile` + subs
- `internal/cli/profile_setup.go:104` — `moai profile setup`
- `internal/cli/worktree/root.go:16` — `moai worktree` subtree
- `internal/cli/statusline.go:18` — `moai statusline`
- `internal/cli/astgrep.go:29` — `moai ast-grep`
- `internal/cli/telemetry.go:14, 25` — `moai telemetry report`
- `internal/cli/status.go:22` — `moai status`

### Hooks
- `internal/hook/registry.go:18-325` — Registry implementation
- `internal/hook/types.go:19-114` — 27 EventType constants
- `internal/hook/types.go:167-311` — HookInput/HookOutput schema
- `internal/hook/types.go:313-471` — Factory helpers
- `internal/cli/deps.go:151-186` — Handler registrations

### Templates
- `internal/template/embed.go:25` — go:embed directive
- `internal/template/embed.go:36-39` — EmbeddedTemplates
- `internal/template/context.go:10-59` — TemplateContext fields
- `internal/template/context.go:64-96` — NewTemplateContext defaults
- `internal/template/deployer.go:21-30` — Deployer interface
- `internal/template/deployer.go:43-62` — Factory constructors
- `internal/template/deployer.go:65-169` — Deploy logic
- `internal/template/deployer.go:197-225` — validateDeployPath

### Versioning
- `pkg/version/version.go:7` — Version default
- `Makefile:11` — ldflags injection

### Config
- `.moai/config/sections/system.yaml` — version + template_version
- `.moai/config/sections/workflow.yaml:1-118` — Workflow and team profiles
- `.moai/config/sections/quality.yaml` — TRUST5 gates
- `.moai/config/sections/harness.yaml` — Harness levels & effort mapping
- `.moai/config/sections/design.yaml` — GAN loop + brand context
- `.moai/config/sections/language.yaml` — ko/en pref
- `.moai/config/sections/llm.yaml` — GLM API config
- `.moai/config/sections/lsp.yaml` — 16-language matrix
- `.moai/config/sections/mx.yaml` — @MX tag rules

### Agent samples
- `.claude/agents/moai/expert-backend.md:23-35` — Agent-level hooks
- `.claude/agents/moai/manager-spec.md:1-20` — opus model, spec skills
- `.claude/agents/moai/plan-auditor.md:1-20` — inherit model, adversarial stance

### Skills
- `.claude/skills/moai/SKILL.md:1-11` — moai orchestrator entry
- `.claude/skills/moai/workflows/` (21 files: clean, codemaps, context, coverage, design, e2e, feedback, fix, gate, github, loop, moai, mx, plan, project, review, run, security, sync)
- `.claude/skills/moai/team/` (6 files: debug, glm, plan, review, run, sync)
- `.claude/skills/moai/references/` (2 files: mx-tag.md, reference.md)
- `.claude/skills/moai-foundation-core/modules/` (17 files)

### CLAUDE.md / local
- `CLAUDE.md` (678 lines, 28,321 bytes) — 17 sections, 11 `[HARD]` rules in §1
- `CLAUDE.local.md` (761 lines, 23,015 bytes) — Developer-private rules
- `grep -c "^## " CLAUDE.md = 17` (section count)
- `grep -c "\[HARD\]" CLAUDE.md = 19` (HARD rule count)

### SPECs
- `.moai/specs/` (104 directories)
- Sample: `.moai/specs/SPEC-CORE-001/{spec,plan,acceptance}.md`
- `.moai/specs/SPEC-EMBED-001/spec.md:1-40` — go:embed SPEC example

### Docs site
- `docs-site/hugo.yaml:1` — baseURL adk.mo.ai.kr
- `docs-site/content/{ko,en,ja,zh}/` — 4 locales
- `docs-site/vercel.json` — deployment config
- `docs-site/scripts/translate.mjs` — AI translation

### MCP
- `.mcp.json:1-22` — 3 servers (context7, sequential-thinking, chrome-devtools)
- `.mcp.json` — NO pencil (user-level instead)
- `expert-frontend.md` tools line — 14 pencil tools declared

### Test counts (grep `^func Test`)
- `internal/` + `pkg/` total: 3762
- `internal/cli/`: 1142
- `internal/hook/`: 733
- `internal/template/`: 94

### Git state
- HEAD: `ad2019683`
- Tag: `v2.12.0-15-gad2019683`
- Branch: `main` (clean)

---

## Summary Numbers

| Metric | Count |
|--------|------:|
| CLI top-level subcommands | 20 |
| Hook event types (types.go) | 27 |
| Hook handlers registered | 28 |
| Hook wrapper scripts (template) | 27 |
| Hook wrapper scripts (local) | 26 |
| Internal packages | 29 |
| Agents (local/template) | 22 / 22 |
| Skills (local/template) | 47 / 50 |
| Commands (local moai/) | 13 |
| Commands (template moai/) | 14 |
| Agency commands (deprecated) | 8 |
| Rule files | 35 |
| Language rules | 16 |
| Config YAML sections | 22 |
| Evaluator profiles | 4 |
| ast-grep rules | 20 |
| SPECs | 104 |
| Templates embedded files | 502 |
| Skill markdown files total | 320 |
| go.mod direct deps | 9 |
| go.mod indirect deps | 23 |
| Test functions total | 3762 |
| Test functions cli/ | 1142 |
| Test functions hook/ | 733 |
| docs-site locales | 4 |
| CLAUDE.md sections | 17 |
| CLAUDE.md HARD rules (§1) | 19 |

End of Wave 1.6 snapshot.
