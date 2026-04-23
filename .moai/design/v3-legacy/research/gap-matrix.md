# MoAI-ADK v3 — Gap Matrix (CC-has / moai-LACKS)

Generated: 2026-04-22
Source: Wave 1 findings (6 documents) synthesized via Wave 2.

## Purpose

This matrix catalogs every capability where Claude Code's TypeScript runtime or CC ecosystem
has a feature that moai-adk-go v2.12 does NOT. Each row is anchored to at least one Wave 1
findings file:section.

## How to Read

- **Severity**: `Critical` (blocks v3 parity claim) · `High` (significant UX) · `Medium` (nice-to-have) · `Low` (marginal)
- **Effort**: `XS` (1-2 days) · `S` (3-5 days) · `M` (1-2 weeks) · `L` (3-4 weeks) · `XL` (>1 month)
- **Breaking?**: Yes = changes an existing contract or behavior; No = additive
- **Source**: Wave 1 file + section. Shorthand — `W1.1` = findings-wave1-hooks-commands.md, `W1.2` = findings-wave1-query-context.md, `W1.3` = findings-wave1-agent-team.md, `W1.4` = findings-wave1-ui-ux.md, `W1.5` = findings-wave1-bootstrap-cli.md, `W1.6` = findings-wave1-moai-current.md

Rows are grouped by domain. Items in Section 7 are moai-adk self-identified issues from Wave 1.6 §15.

---

## Section 1 — Hook System Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 1 | Hook Types | `type: 'prompt'` — LLM-evaluated hook (Haiku-class, returns `{ok, reason?}` JSON) | Only `type: 'command'` (shell) supported | High | M | No | W1.1 §2.2; W1.5 §6.2 |
| 2 | Hook Types | `type: 'agent'` — full subagent verifier (multi-turn query with tools, bounded by timeout) | Not implemented | High | M | No | W1.1 §2.3; W1.5 §6.2 |
| 3 | Hook Types | `type: 'http'` — POST JSON webhook with SSRF guard + env-var allowlist + CRLF sanitization | Not implemented | Medium | M | No | W1.1 §2.4; W1.1 §12 (ssrfGuard.ts) |
| 4 | Hook Output | Hook JSON output protocol: `decision`, `hookSpecificOutput`, `additionalContext`, `updatedInput`, `systemMessage`, `suppressOutput`, `stopReason`, `continue` | Only exit codes + minimal JSON | Critical | L | Yes | W1.1 §4.1; W1.1 §14.3 |
| 5 | Hook Output | `hookSpecificOutput.additionalContext` — injects content into model turn | Not implemented | Critical | M | No | W1.1 §14.3 |
| 6 | Hook Output | `hookSpecificOutput.updatedInput` — hook rewrites tool input before execution | Not implemented | High | M | No | W1.1 §14.3 |
| 7 | Hook Output | `hookSpecificOutput.updatedMCPToolOutput` — hook rewrites MCP tool output | Not implemented | Medium | S | No | W1.1 §14.3 |
| 8 | Hook Features | `if` condition with permission-rule syntax (`Bash(git *)`, `Read(*.ts)`) — filter spawns before subprocess start | Not implemented; hooks always spawn | High | S | No | W1.1 §3.8; W1.1 §12 |
| 9 | Hook Features | `async: true` — backgrounded hook execution | Not implemented | High | M | No | W1.1 §14.3 |
| 10 | Hook Features | `asyncRewake: true` — backgrounded + wake model on exit 2 via task-notification queue | Not implemented | High | M | No | W1.1 §14.3 |
| 11 | Hook Features | `once: true` — self-remove after first successful execution | Not implemented | Medium | XS | No | W1.1 §3.7; W1.1 §14.3 |
| 12 | Hook Features | `shell: 'powershell'` selection per hook | bash-only | Low | S | No | W1.1 §4.5; W1.1 §14.3 |
| 13 | Hook Features | Hook matcher patterns: exact / pipe-separated / regex / `*` / per-event `matchQuery` resolution | Basic matchers only | Medium | S | No | W1.1 §3.8 |
| 14 | Hook Features | Hook dedup across sources (plugin/skill/settings) via `{shell}\0{command}\0{if}` key | Not implemented | Medium | S | No | W1.1 §3.9; W1.1 §14.3 |
| 15 | Hook Sources | Source precedence pipeline (8 levels): policySettings → userSettings → projectSettings → localSettings → pluginHook → skillHook → sessionHook → builtinHook | Only project `.claude/settings.json` loaded | High | M | Yes | W1.1 §3.2; W1.1 §14.3 |
| 16 | Hook Sources | Session-scoped hooks via `addSessionHook` — ephemeral, skill/agent-frontmatter driven | Not implemented | Medium | S | No | W1.1 §3.6; W1.1 §14.3 |
| 17 | Hook Sources | Skill-frontmatter `hooks:` — register session hooks when skill invoked | Not implemented | Medium | S | No | W1.1 §3.6 |
| 18 | Hook Sources | Agent-frontmatter `hooks:` with Stop→SubagentStop rewrite | Agent-level hooks exist (e.g., expert-backend) but no Stop→SubagentStop rewrite | Medium | S | No | W1.1 §3.6; W1.6 §6.3 |
| 19 | Hook Sources | `policySettings.disableAllHooks` / `allowManagedHooksOnly` gating | Not implemented | Low | S | No | W1.1 §3.4; W1.1 §14.3 |
| 20 | Hook Env | `CLAUDE_ENV_FILE` mechanism — hook writes bash exports for subsequent BashTool commands (SessionStart/Setup/CwdChanged/FileChanged) | Not implemented | High | S | No | W1.1 §4.3; W1.1 §14.3 |
| 21 | Hook Env | `${CLAUDE_PLUGIN_ROOT}` / `${CLAUDE_PLUGIN_DATA}` / `${user_config.X}` substitution in command strings | Not implemented | Medium | S | No | W1.1 §4.4 |
| 22 | Hook Timeout | SessionEnd 1500ms tight bound (override via `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS`) | Uses generic `DefaultHookTimeout = 30s` | Low | XS | No | W1.1 §4.7; W1.6 §5.1 |
| 23 | Hook Bidir | Prompt elicitation from hook stdout — `{"prompt":"id","message":"...","options":[...]}` JSON line → `AskUserQuestion` bridge → stdin response | Not implemented | Low | M | No | W1.1 §4.10; W1.1 §14.3 |
| 24 | Hook Event | `WorktreeCreate` as PROVIDER hook: stdout MUST contain absolute path to created worktree | moai WorktreeCreate handler is observational only | Medium | S | Yes | W1.1 §14.4 |
| 25 | Hook Event | SessionStart `watchPaths` reply — registers paths with FileChanged watcher | Not implemented | Medium | S | No | W1.1 §2.1 (SessionStart) |
| 26 | Hook Event | CwdChanged `watchPaths` reply — dynamic watch list update | Not implemented | Low | S | No | W1.1 §2.1 (CwdChanged) |
| 27 | Hook Event | PermissionRequest hook reply: `decision: { behavior:'allow'\|'deny', updatedInput?, updatedPermissions? }` — programmatic permission decisions | Event type defined; handler is shell-based without decision schema | High | M | Yes | W1.1 §2.1 (PermissionRequest) |
| 28 | Hook Event | PermissionDenied hook reply: `retry?: boolean` — tells model it may retry | Event type defined; handler lacks retry semantics | Medium | S | No | W1.1 §2.1 (PermissionDenied) |
| 29 | Hook Event | PreCompact hook — return `{newCustomInstructions?, userDisplayMessage?}` | Event type defined; only observational | Medium | S | No | W1.1 §2.1 (PreCompact) |
| 30 | Hook Event | Elicitation / ElicitationResult hook reply — auto-respond to or override MCP elicitation dialogs | Event types defined; observational only | Low | M | No | W1.1 §2.1 (Elicitation) |
| 31 | Hook Event | ConfigChange hook — exit 2 can BLOCK settings change from being applied | Event type defined; observational only | Medium | S | Yes | W1.1 §2.1 (ConfigChange) |
| 32 | Hook Event | InstructionsLoaded hook — observability for CLAUDE.md/memory file loading with `load_reason` + globs | Event type defined; not wired to load audit | Medium | S | No | W1.1 §2.1 (InstructionsLoaded) |
| 33 | Hook Trust | Workspace-trust gating on hooks (`shouldSkipHookDueToTrust`) | Not implemented | Low | S | No | W1.1 §4.6 |
| 34 | Hook Progress | Hook progress streaming to SDK consumers via `HookStartedEvent`/`HookProgressEvent`/`HookResponseEvent` | Not applicable (moai has no SDK consumers) | Low | - | No (N/A) | W1.1 §4.9 |

---

## Section 2 — Query / Context / Memory / Compaction Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 35 | Query Engine | Stateful `QueryEngine` with mutable messages, file-read cache, permission denial log | None — moai delegates to CC runtime | Low | - | No (N/A) | W1.2 §1.1; W1.2 §9.1 |
| 36 | Compaction | 5-layer pipeline (snip → microcompact → context-collapse → autocompact → reactive) | None | Low | - | No (N/A, CC-runtime) | W1.2 §2.8; W1.2 §9.7 |
| 37 | Compaction | Autocompact thresholds (`AUTOCOMPACT_BUFFER=13K`, `MANUAL_COMPACT_BUFFER=3K`, circuit breaker at 3 failures) | None | Low | - | No (N/A) | W1.2 §2.1; W1.2 §9.7 |
| 38 | Context | System prompt assembly with cache-key prefix ordering (`systemPrompt`, `userContext`, `systemContext`) | None (CC-runtime) | Low | - | No (N/A) | W1.2 §3.1; W1.2 §9.2 |
| 39 | Context | Git status injection with `MAX_STATUS_CHARS=2000` truncation | Not in moai (CC injects automatically) | Low | - | No (N/A) | W1.2 §3.4; W1.2 §9.2 |
| 40 | History | Global prompt history `~/.claude/history.jsonl` locked, 100-entry window | N/A — CC-runtime feature | Low | - | No (N/A) | W1.2 §4.1; W1.2 §9.3 |
| 41 | Resume | Per-session transcript file `{projectDir}/{sessionId}.jsonl` with 50MB read guard | N/A — CC-runtime feature | Low | - | No (N/A) | W1.2 §4.7; W1.2 §9.3 |
| 42 | Cost | Per-model USD accounting (input/output/cacheRead/cacheCreation) | Not tracked | Low | S | No | W1.2 §5.5; W1.2 §9.4 |
| 43 | Cost | `maxBudgetUsd` hard budget with per-iteration check | Not implemented | Low | S | No | W1.2 §5.7; W1.2 §9.4 |
| 44 | Memory | MEMORY.md truncation: `MAX_ENTRYPOINT_LINES=200`, `MAX_ENTRYPOINT_BYTES=25000` with warning append | No truncation (CLAUDE.md has 40K char cap but not MEMORY.md) | Critical | XS | No | W1.2 §6.5; W1.2 §9.5 |
| 45 | Memory | 4-type memory taxonomy (user/feedback/project/reference) with full XML prompts | Schema ALREADY aligns; not formally enforced | High | XS | No | W1.2 §6.6; W1.2 §9.5 |
| 46 | Memory | Auto-directory resolution: git-root-canonical → env override chain (`CLAUDE_CODE_REMOTE_MEMORY_DIR`, `CLAUDE_COWORK_MEMORY_PATH_OVERRIDE`) | Fixed paths, no override | Medium | S | No | W1.2 §6.2; W1.2 §9.5 |
| 47 | Memory | Memory scan: cap 200 files, mtime-sort, 30-line frontmatter read | None — moai context search uses Grep on sessions | Medium | S | No | W1.2 §6.11; W1.2 §9.5 |
| 48 | Memory | LLM-based relevance selection (Sonnet sideQuery, max 5 memories returned) | None; moai uses Grep-based context search | High | S | No | W1.2 §6.10; W1.2 §9.5 |
| 49 | Memory | `alreadySurfaced` dedup within session | Not implemented | Medium | XS | No | W1.2 §6.10 |
| 50 | Memory | Memory age string + `memoryFreshnessNote` — `<system-reminder>`-wrapped staleness caveat for >1 day memories | Not implemented | High | XS | No | W1.2 §6.12; W1.2 §9.5 |
| 51 | Memory | Memory prefetch in parallel with model streaming (`startRelevantMemoryPrefetch`) | None — moai context search is blocking | Low | - | No (N/A, CC-runtime) | W1.2 §6.14; W1.2 §9.5 |
| 52 | Memory | `filterDuplicateMemoryAttachments` via `readFileState` cache | Not implemented | Low | S | No | W1.2 §6.14 |
| 53 | Memory | `validateMemoryPath` security rules: reject non-absolute, tilde-only, UNC paths, null byte, NFKC attacks | Not implemented | Medium | S | No | W1.2 §6.3; W1.2 §9.5 |
| 54 | Memory | Team memory sub-dir `{autoMemPath}/team/` with path-key sanitization | Not implemented (moai team memory is implicit in `.moai/project/brand/` without validation) | Medium | S | No | W1.2 §6.13 |
| 55 | Memory | KAIROS daily-log mode (`{autoMemPath}/logs/YYYY/MM/YYYY-MM-DD.md`) + nightly `/dream` distillation | Not applicable (feature-gated in CC) | Low | - | No (N/A) | W1.2 §6.4 |

---

## Section 3 — Agent / Team / Task Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 56 | Agent FM | `memory: user\|project\|local` — persistent agent-scoped memory with snapshot-based init | moai has auto-memory but no agent-type scope | High | S | No | W1.3 §4.2; W1.3 §8.1 |
| 57 | Agent FM | `initialPrompt` — prepended to first user turn (slash commands work) | Not implemented | High | XS | No | W1.3 §4.2; W1.3 §8.1 |
| 58 | Agent FM | `requiredMcpServers` — agent unavailable if listed MCP servers not connected | Not implemented | High | S | No | W1.3 §4.2; W1.3 §8.1 |
| 59 | Agent FM | `omitClaudeMd` — skip CLAUDE.md hierarchy (saves 5-15 Gtok/week per CC BQ analysis) | Not implemented (moai always loads CLAUDE.md) | High | XS | No | W1.3 §4.2; W1.3 §8.1 |
| 60 | Agent FM | `maxTurns` — hard cap on agentic turns | Not implemented | Medium | XS | No | W1.3 §4.2 |
| 61 | Agent FM | `skills:` frontmatter — preload skills into agent tool pool | Already partially used in moai agent frontmatter (YAML array) | Medium | XS | No | W1.3 §4.2; W1.6 §6.3 |
| 62 | Agent FM | `criticalSystemReminder_EXPERIMENTAL` — re-injected at every user turn | Not implemented | Medium | XS | No | W1.3 §4.2 |
| 63 | Agent FM | `background: true` in frontmatter — forces async execution | moai documents this in prompts but not a frontmatter field | Medium | XS | No | W1.3 §4.2 |
| 64 | Agent FM | `permissionMode: bubble` — fork-specific mode surfacing prompts to parent | Not implemented | Low | S | No | W1.3 §4.3 |
| 65 | Agent FM | `permissionMode: auto` — classifier-based auto-approval | Not implemented | Low | M | No | W1.3 §4.3 |
| 66 | Built-in Agents | Explore / Plan / Code Guide / Statusline Setup / Verification / worker / fork as first-class built-ins | moai has 22 custom agents; no CC-equivalent built-ins shipped (relies on CC runtime) | Medium | S | No | W1.3 §4.4; W1.6 §6.2 |
| 67 | Fork | Fork subagent primitive: omit `subagent_type` → child inherits parent's conversation + system prompt (cache-identical) | Not implemented | High | L | No | W1.3 §4.4; W1.3 §8.1 |
| 68 | Team Backends | In-process teammate backend (AsyncLocalStorage-based, same-process) | tmux-only | High | L | No | W1.3 §7.1; W1.3 §8.1 |
| 69 | Team Backends | iTerm2 native-split backend via `it2` CLI | tmux-only | Low | S | No | W1.3 §7.1; W1.3 §8.1 |
| 70 | Team Backends | Detection priority: tmux → iterm2 → tmux-external → throw | tmux-only | Low | S | No | W1.3 §7.1 |
| 71 | Team Mailbox | Zod-validated structured message types: `shutdown_request`, `shutdown_approved`, `shutdown_rejected`, `plan_approval_request`, `plan_approval_response`, `permission_request`, `permission_response`, `sandbox_permission_request`, `sandbox_permission_response`, `task_assignment` (10 types) | Ad-hoc JSON | Medium | M | Yes | W1.3 §6.8; W1.3 §8.2 |
| 72 | Team Flow | Plan-approval flow: team lead approves/rejects teammate plans via structured message with `feedback` | Not implemented | Medium | S | No | W1.3 §6.4; W1.3 §8.3 |
| 73 | Team Flow | Per-teammate permission mode (`TeamFile.members[].mode`) — independent cyclable modes | Not implemented | Low | S | No | W1.3 §8.1 |
| 74 | Team Flow | Team-allowed paths (`TeamFile.teamAllowedPaths`) — shared directory allowlist | Not implemented | Low | S | No | W1.3 §8.1 |
| 75 | Team Flow | `TEAMMATE_MESSAGES_UI_CAP = 50` — prevents 36GB whale sessions | Not implemented | Medium | XS | No | W1.3 §7.2; W1.3 §8.1 |
| 76 | Team Flow | Word-slug team-name collision via `generateWordSlug()` | Not implemented | Low | XS | No | W1.3 §6.2; W1.3 §8.1 |
| 77 | Team Flow | Session cleanup registry (`registerTeamForSessionCleanup`) — SIGINT auto-cleanup of team directories | Partial (moai handles tmux pane cleanup per feedback memory) | Low | S | No | W1.3 §6.3 |
| 78 | Task Framework | 7 task types with unified `kill(taskId, setAppState)` interface: local_bash, local_agent, remote_agent, in_process_teammate, local_workflow, monitor_mcp, dream | moai has per-domain handlers; no unified Task abstraction | Medium | M | Yes | W1.3 §6.7; W1.3 §8.1 |
| 79 | Task Framework | `TaskStop` as LLM-invokable tool (deprecated `KillShell` alias) | Not implemented; moai uses tmux pane-kill via admin paths | Medium | S | No | W1.3 §6.5; W1.3 §8.1 |
| 80 | Task Framework | `isConcurrencySafe` tool flag — marks tools as parallel-safe | Not implemented | Low | XS | No | W1.3 §8.1 |
| 81 | Task Framework | Auto-background timer (2min) for sync sub-agents | Not implemented | Low | S | No | W1.3 §7.4; W1.3 §8.1 |
| 82 | Agent Registry | `agentNameRegistry` maps name → agentId; SendMessage to evicted agent auto-resumes via `resumeAgentBackground` | Not implemented | Low | M | No | W1.3 §8.1 |
| 83 | Agent Worktree | `isolation: 'worktree'` with cleanup logic (hookBased keep, headCommit match remove) | Implemented; no hookBased concept | Low | S | No | W1.3 §7.5 |
| 84 | Agent Worktree | `isolation: 'remote'` (ant-only) — teleport to CCR container | Not applicable (Anthropic-internal) | Low | - | No (N/A) | W1.3 §2.3; W1.3 §8.1 |
| 85 | Coordinator | Coordinator mode: single-worker orchestrator with 380-line system prompt, `<task-notification>` XML envelope | Not implemented; moai orchestrator uses AskUserQuestion + Agent() | Low | L | No | W1.3 §2.1 |
| 86 | Bridge | Remote Control Bridge (`bridge/`) — 33 files, WebSocket to api.anthropic.com + OAuth | Not applicable (Anthropic-cloud specific; CC's bridge not reusable for moai CG mode) | Low | - | No (REJECT) | W1.3 §1; W1.3 §8.1 |

---

## Section 4 — Command / Skill Ecosystem Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 87 | Skill FM | `context: 'inline' \| 'fork'` — inline expands into current conv, fork runs as sub-agent | Inline only | High | L | No | W1.1 §11; W1.1 §14.5 |
| 88 | Skill FM | `agent:` frontmatter — sub-agent type when context:fork | Not implemented | High | S | No | W1.1 §11 |
| 89 | Skill FM | `paths:` conditional activation (glob patterns, gitignore-style filter) | Documented in coding-standards.md but not used | High | S | No | W1.1 §6.7; W1.1 §14.5; W1.6 §7.3 |
| 90 | Skill FM | `effort:` per-skill Opus 4.7 override (low/medium/high/xhigh/max/integer) | moai has effortLevel in settings.json only | Medium | XS | No | W1.1 §14.5 |
| 91 | Skill FM | `disableModelInvocation` — hide from SkillTool | Not implemented | Low | XS | No | W1.1 §14.5 |
| 92 | Skill FM | `user-invocable` — allow/disallow user typing /skill-name | Implicit (all moai skills user-invocable) | Low | XS | No | W1.1 §14.5 |
| 93 | Skill FM | `hide-from-slash-command-tool` — hide from SlashCommand tool | Not implemented | Low | XS | No | W1.1 §14.5 |
| 94 | Skill FM | `version:` skill versioning | Not implemented | Low | XS | No | W1.1 §14.5 |
| 95 | Skill FM | `isSensitive: true` — redact args from conversation history | Not implemented | Low | S | No | W1.1 §14.5 |
| 96 | Skill FM | `immediate: true` — execute without waiting for stop point, bypass queue | Not implemented | Low | S | No | W1.1 §14.5 |
| 97 | Skill FM | YAML special-char auto-quoting (glob patterns parse correctly) | Not implemented | Medium | S | No | W1.1 §7 (frontmatterParser.ts) |
| 98 | Skill FM | Brace expansion in `paths:` (e.g. `src/*.{ts,tsx}` → 2 patterns) | Not implemented | Low | XS | No | W1.1 §7 |
| 99 | Args | `$ARGUMENTS[N]` indexed substitution (0-indexed) | Not implemented (only `$ARGUMENTS`) | Medium | XS | No | W1.1 §8; W1.1 §14.5 |
| 100 | Args | `$N` shorthand for `$ARGUMENTS[N]` | Not implemented | Medium | XS | No | W1.1 §8 |
| 101 | Args | `$name` named arg substitution when `arguments: "name1 name2"` set | Not implemented | Medium | XS | No | W1.1 §8; W1.1 §14.5 |
| 102 | Args | Progressive argument hint as user types (remaining-args typeahead) | Partial (moai has argument-hint frontmatter, no progressive fill) | Low | M | No | W1.1 §8; W1.1 §14.5 |
| 103 | Skill Body | <code>!`cmd`</code> inline shell + <code>```!</code> block execution inside skill body | Not implemented | Medium | S | No | W1.1 §9; W1.1 §14.5 |
| 104 | Skill Body | `${CLAUDE_SKILL_DIR}` substitution in skill body (normalized to forward slashes on Windows) | Not implemented | Medium | XS | No | W1.1 §9; W1.1 §14.5 |
| 105 | Skill Body | `${CLAUDE_SESSION_ID}` substitution in skill body | Not implemented | Low | XS | No | W1.1 §9; W1.1 §14.5 |
| 106 | Skill Discovery | Dynamic skill discovery: walk UP from touched file paths, load nested `.claude/skills/` (deepest-first) | Not implemented; moai loads once at session start | Medium | M | No | W1.1 §6.6; W1.1 §14.5 |
| 107 | Skill Discovery | Realpath canonicalization dedup (handles symlinks + nested git repos) | Not implemented | Low | S | No | W1.1 §6.5 |
| 108 | Skill Discovery | Legacy `.claude/commands/` namespaced skills (path-prefix → `namespace:name`) | Not applicable (moai uses `.claude/commands/moai/` but as thin routers) | Low | - | No (N/A) | W1.1 §6.4 |
| 109 | Slash Commands | `/rewind` (checkpoint) — restore code + conversation to previous point | Not implemented | Low | M | No | W1.1 §14.6 |
| 110 | Slash Commands | `/branch` (fork) — create branch of current conversation | Not implemented | Low | M | No | W1.1 §14.6 |
| 111 | Slash Commands | `/diff` — view uncommitted changes + per-turn diffs | Not implemented | Medium | S | No | W1.1 §14.6 |
| 112 | Slash Commands | `/context` — visualize context usage as colored grid | Not implemented | Medium | M | No | W1.1 §14.6 |
| 113 | Slash Commands | `/compact [instructions]` — keep summary in context | Not applicable (CC-runtime feature) | Low | - | No (N/A) | W1.1 §14.6 |
| 114 | Slash Commands | `/memory` — edit Claude memory files | moai has auto-memory but no edit command | Medium | S | No | W1.1 §14.6 |
| 115 | Slash Commands | `/permissions` — manage allow/deny rules | Not implemented | Medium | M | No | W1.1 §14.6 |
| 116 | Slash Commands | `/skills` — list skills | Not implemented (moai uses Skill() programmatically) | Low | S | No | W1.1 §14.6 |
| 117 | Slash Commands | `/agents` — manage agent configurations | Not implemented | Low | S | No | W1.1 §14.6 |
| 118 | Slash Commands | `/hooks` — view hook configurations | Not implemented (moai has `moai hook` CLI) | Low | S | No | W1.1 §14.6 |
| 119 | Slash Commands | `/plugin` / `/marketplace` — plugin management | Not implemented (no plugin system) | Medium | M | No | W1.1 §14.6 |
| 120 | Slash Commands | `/mcp [enable\|disable server]` — MCP server management | Not implemented | Low | M | No | W1.1 §14.6 |
| 121 | Slash Commands | `/reload-plugins` — activate pending plugin changes mid-session | Not applicable (no plugin system yet) | Low | - | No | W1.1 §14.6 |
| 122 | Slash Commands | `/doctor` slash — diagnose installation (moai has `moai doctor` CLI but no slash wrapper) | Not implemented | Low | XS | No | W1.1 §14.6 |
| 123 | Slash Commands | `/model` per-session switch | Not applicable (CC-runtime) | Low | - | No (N/A) | W1.1 §14.6 |
| 124 | Slash Commands | `/effort` — set effort level | moai has effortLevel in settings.json only | Low | XS | No | W1.1 §14.6 |
| 125 | Slash Commands | `/output-style` picker | Not implemented | Low | S | No | W1.5 §7.4 |
| 126 | Slash Commands | `/tasks` — background task viewer (alias: bashes) | Not applicable (CC runtime TaskList is different) | Low | - | No (N/A) | W1.1 §14.6 |

---

## Section 5 — Bootstrap / CLI / Plugin / Migration / Schema Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 127 | CLI Shim | Fast-path shim with 13 zero-import paths (`--version`, `--dump-system-prompt`, `--daemon-worker`, etc.) | All subcommands load full binary | Medium | S | No | W1.5 §1.2; W1.5 §9.1 |
| 128 | CLI Shim | Startup profiler: `profileCheckpoint()` with 20+ named stages | Not implemented | Medium | S | No | W1.5 §1.7; W1.5 §9.1 |
| 129 | CLI Shim | Global env side-effects before main() (COREPACK, NODE_OPTIONS for CCR, ablation) | Not applicable (Go has different constraints) | Low | - | No | W1.5 §1.3 |
| 130 | CLI Init | Memoized `init()` with ordered steps (enableConfigs → safeEnvVars → CAcerts → gracefulShutdown → backgrounds → mTLS → agents → preconnect) | moai's `InitDependencies()` is similar but not memoized | Low | S | No | W1.5 §1.4 |
| 131 | CLI Init | `applyExtraCACertsFromConfig()` before first TLS handshake | Not applicable (moai doesn't make TLS calls) | Low | - | No (N/A) | W1.5 §1.4 |
| 132 | CLI Init | `ConfigParseError` branch with Ink dialog for interactive, stderr for non-interactive | Not implemented; moai errors are plain | Medium | S | No | W1.5 §1.6 |
| 133 | CLI Commander | `preAction` hook — runs migrations, loads remote settings, syncs settings on every invocation | Per-command manually | Medium | S | No | W1.5 §2.1; W1.5 §9.2 |
| 134 | CLI Commander | `createSortedHelpConfig()` enforces alphabetical help | cobra default ordering | Low | XS | No | W1.5 §2.3 |
| 135 | CLI Commander | Shell completion (`completion <shell>`) subcommand | Not shipped (cobra has built-in support) | Low | XS | No | W1.5 §2.1; W1.5 §9.2 |
| 136 | CLI Mode | `--bare` minimal mode (skips hooks, LSP, plugins, attribution, memory, keychain) | Not implemented | Low | S | No | W1.5 §2.2; W1.5 §9.2 |
| 137 | CLI Flag | `--add-dir <dirs...>` — additional tool-access directories | Not applicable (moai doesn't model tool-access) | Low | - | No (N/A) | W1.5 §2.2 |
| 138 | CLI Flag | `--setting-sources <sources>` filter (user/project/local) | Not implemented (single source tier) | High | M | Yes | W1.5 §2.2; W1.5 §9.2 |
| 139 | CLI Flag | `--settings <file-or-json>` — inline settings | Not implemented | Medium | S | No | W1.5 §2.2 |
| 140 | Plugin | Plugin manifest (`.claude-plugin/plugin.json`) + 3 plugin kinds (built-in, marketplace, session inline) | Not implemented (moai has templates only) | High | XL | No | W1.5 §4.1; W1.5 §9.4 |
| 141 | Plugin | Marketplace with 5 source types: github, git, url, directory, file | Not implemented | High | L | No | W1.5 §4.5; W1.5 §9.4 |
| 142 | Plugin | Plugin capabilities: agents, skills, commands, hooks, mcpServers, outputStyles | Not applicable until plugin system exists | High | XL | No | W1.5 §4.6; W1.5 §9.4 |
| 143 | Plugin | Install scopes (user/project/local) | Not applicable until plugin system exists | Medium | M | No | W1.5 §4.4; W1.5 §9.4 |
| 144 | Plugin | `plugin validate <path>` w/ manifest + content walk | Not implemented | Medium | S | No | W1.5 §4.4; W1.5 §9.4 |
| 145 | Plugin | `--plugin-dir <path>` repeatable session-only inline plugin | Not implemented | Medium | S | No | W1.5 §4.1 |
| 146 | Plugin | Plugin hook hot-reload on policySettings change | Not applicable until plugin system exists | Low | M | No | W1.1 §3.5 |
| 147 | Plugin | `clearPluginCache()` + cache TTL | Not applicable until plugin system exists | Low | S | No | W1.5 §4.7 |
| 148 | Plugin | Cowork plugins (`--cowork` flag triggers separate `cowork_plugins/` dir) | Not applicable | Low | - | No (N/A) | W1.5 §4.5 |
| 149 | Migration | `CURRENT_MIGRATION_VERSION` counter with ordered migration runner | Only `moai migrate agency` one-shot | Critical | L | Yes | W1.5 §5.1; W1.5 §9.5 |
| 150 | Migration | Idempotent migrations (flag-based + value-match patterns) | Partial (agency migration tracks completion) | High | M | No | W1.2 §7.2; W1.5 §9.5 |
| 151 | Migration | Migration invocation in preAction hook (auto-fires on every `claude` run) | Explicit `moai migrate` subcommand only | High | S | No | W1.5 §5.2; W1.5 §9.5 |
| 152 | Migration | Source specificity: `userSettings` only, never touches `projectSettings` | Not applicable (single tier) | Medium | - | No | W1.2 §7.4 |
| 153 | Migration | Async migration fire-and-forget (`migrateChangelogFromConfig`) | Not implemented | Low | XS | No | W1.5 §5.1 |
| 154 | Migration | Migration analytics events (`tengu_migrate_*` with has_error flag) | Not implemented | Low | XS | No | W1.2 §7.5; W1.5 §9.5 |
| 155 | Migration | Timestamp-based notification trigger (e.g. `sonnet45To46MigrationTimestamp`) | Not implemented | Low | XS | No | W1.2 §7.6 |
| 156 | Schema | Zod v4 schemas with `lazySchema()` for deferred construction | YAML via go-yaml; no schemas | Critical | L | Yes | W1.5 §6.1; W1.5 §9.6 |
| 157 | Schema | Settings schema (generated TypeScript types from Zod) | No generated types | High | M | No | W1.5 §6.4; W1.5 §9.6 |
| 158 | Schema | Hook schema (discriminated union of 4 types, no `.transform()`) | Partial (Go structs for hook events, no schema union) | High | M | No | W1.5 §6.2; W1.5 §9.6 |
| 159 | Schema | Sandbox schema with network + filesystem config | Not applicable (moai has no sandbox) | Low | - | No (N/A) | W1.5 §6.3 |
| 160 | Schema | MCP server schema (Stdio/SSE/Http/Sdk/ProcessTransport variants) | Not implemented | Medium | S | No | W1.5 §6.6 |
| 161 | Schema | Permission schema (PermissionUpdate, PermissionDecisionClassification, PermissionMode) | Not implemented | Medium | S | No | W1.5 §6.6 |
| 162 | Schema | SDK control protocol schemas (SDKControlRequest/Response/ElicitationResponse) | Not applicable (no SDK) | Low | - | No (N/A) | W1.5 §6.5 |
| 163 | Schema | Validation at settings load (`parseSettingsFile`) + write (`updateSettingsForSource`) round-trip safety | Not implemented | High | M | No | W1.5 §6.7; W1.5 §9.6 |
| 164 | Settings | 6-source layering: userSettings, projectSettings, localSettings, flagSettings, policySettings, managedSettings | Single `.moai/config/sections/*.yaml` tier | High | L | Yes | W1.5 §9.2 |
| 165 | Output Styles | 4 sources merged: project → user → plugin → built-in | moai ships one style (template) | Low | S | No | W1.5 §7.1; W1.5 §9.7 |
| 166 | Output Styles | Built-ins: default, Explanatory, Learning | Only MoAI style (template) | Low | XS | No | W1.5 §7.3 |
| 167 | Output Styles | `keep-coding-instructions` flag (boolean/string parsing) | Not documented; unclear if used | Low | XS | No | W1.5 §7.2; W1.5 §9.7 |
| 168 | Output Styles | `force-for-plugin` — plugin style auto-applies when plugin enabled | Not applicable (no plugin system) | Low | - | No (N/A) | W1.5 §7.5 |
| 169 | Output Styles | Frontmatter parsing: name, description, keep-coding-instructions, force-for-plugin | Not formalized | Low | XS | No | W1.5 §7.2 |
| 170 | Entry Points | MCP server entrypoint (`claude mcp serve` over stdio) | Not applicable (moai tools are Go binaries) | Low | - | No (DEFER) | W1.5 §3.3; W1.5 §9.3 |
| 171 | Entry Points | SDK re-export barrel (`entrypoints/agentSdkTypes.ts`) | Not applicable (moai is not a library) | Low | - | No (REJECT) | W1.5 §3.4; W1.5 §9.3 |
| 172 | Entry Points | Headless mode (`-p`/`--print`) with stream-json I/O | N/A — moai is always headless | Low | - | No (N/A) | W1.5 §3.5; W1.5 §9.3 |
| 173 | Entry Points | REPL entry (`replLauncher.tsx` mounting Ink TUI) | N/A — moai has no REPL | Low | - | No (N/A) | W1.5 §3.6; W1.5 §9.3 |
| 174 | Entry Points | Worktree fast-path (`--worktree --tmux` execs into tmux before main) | Not implemented (moai uses MoAI worktree layer) | Low | S | No | W1.5 §3.7 |

---

## Section 6 — UI/UX Output Contract Gaps

| # | Domain | Capability (CC) | moai-adk State | Gap Severity | Effort | Breaking? | Source |
|---|--------|-----------------|----------------|--------------|--------|-----------|--------|
| 175 | Output | Diff visualization via `diff --git` format → CC's StructuredDiff renders with syntax highlighting | moai emits plain text diffs | Medium | XS | No | W1.4 §8.4 |
| 176 | Output | Structured validation errors `{severity, path, line, message, suggestion}` → CC renders via ValidationErrorsList | moai errors are plain text | Medium | S | No | W1.4 §8.4 |
| 177 | Output | Progress indicators (`Status:`/`Task:`/`Progress:` prefixes during long tasks) | Silent during LSP scans, template rendering | Medium | S | No | W1.4 §8.4 |
| 178 | Output | Code block language hints (` ```go `, ` ```python `) for HighlightedCode syntax highlighting | Often emits raw code without fence language | Low | XS | No | W1.4 §8.4 |
| 179 | Output | StatusIcon semantic colors: ✓ (success), ✗ (error), ⚠ (warning), ℹ (info), ○ (pending), … (loading) | Not consistently used | Low | XS | No | W1.4 §8.2; W1.4 §2.10 |
| 180 | Output | File path OSC-8 hyperlinks (`file:///absolute/path:LINE`) | Not consistently used | Low | XS | No | W1.4 §8.2 |
| 181 | Output | Progress bar via eighth-block chars `▏▎▍▌▋▊▉█` | Not implemented | Low | XS | No | W1.4 §8.2; W1.4 §2.3 |
| 182 | Output Styles | moai output-style schema compat with CC keys (name, description, keep-coding-instructions, force-for-plugin) vs moai: prefix for extensions | Not verified — potential collision if moai uses custom frontmatter keys | Medium | XS | No | W1.4 §8.5 |

---

## Section 7 — moai-adk-go Self-Identified Issues (Wave 1.6 §15)

| # | Domain | Issue | Current State | Severity | Effort | Breaking? | Source |
|---|--------|-------|---------------|----------|--------|-----------|--------|
| 183 | Config Drift | `project.yaml:template_version: v2.7.22` — stale by ~12 minor versions (system.yaml says v2.12.0) | `moai update` does not auto-sync `project.yaml.template_version` | Critical | XS | No | W1.6 §15.4; W1.6 §15.8 |
| 184 | Template Drift | 3 template-only skills missing locally: `moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration` | Local 47 skills; template 50 | Medium | XS | No | W1.6 §15.1; W1.6 §7.1 |
| 185 | Template Drift | Hook wrapper `handle-permission-denied.sh` missing locally (template has `.tmpl`) | Local 26 wrappers; template 27 | Medium | XS | No | W1.6 §15.1; W1.6 §15.7; W1.6 §5.5 |
| 186 | Code Cruft | `internal/cli/glm.go.bak` (28,567 bytes) — pre-April refactor backup | File checked in; should be removed | Low | XS | No | W1.6 §15.10 |
| 187 | Code Cruft | `internal/cli/worktree/new_test.go.bak` (13,700 bytes) — pre-April test backup | File checked in; should be removed | Low | XS | No | W1.6 §15.10 |
| 188 | Coverage | `coverage.out` + `coverage.html` dated 2026-03-11 (stale) | Predates many SPEC changes; actual coverage unknown | Low | XS | No | W1.6 §13.2; W1.6 §15.2 |
| 189 | Doc Drift | `internal/template/embed.go:8-12` ADR-011 comment claims runtime-generated files are excluded, but `.claude/settings.json.tmpl` IS embedded and rendered | Stale comment | Low | XS | No | W1.6 §4.6; W1.6 §15.3 |
| 190 | Legacy | `.agency/` vestigial (absorbed per SPEC-AGENCY-ABSORB-001) — `.claude/rules/agency/constitution.md` is 695-byte redirect stub; `.claude/commands/agency/*.md` are redirect-only | 8 agency command redirects + constitution stub still deployed | Low | XS | No | W1.6 §15.6 |
| 191 | Legacy | `.moai-backups/` folder at project root (April migration snapshot) | Exists; cleanup candidate | Low | XS | No | W1.6 §15.6 |
| 192 | MCP Config | `pencil` MCP server declared in `expert-frontend.md` tools but NOT in `.mcp.json` (user-level only, not project-scoped) | Implicit user-level config; could be inconsistent across machines | Low | XS | No | W1.6 §16.2 |
| 193 | Handler Count | `InitDependencies()` registers 28 handler calls but only 27 EventType constants exist (AutoUpdateHandler is a second SessionStart handler via compose pattern) | Intentional but undocumented | Low | XS | No | W1.6 §15.5 |
| 194 | Docs Locale Lag | `docs-site/content/en/` lacks `contributing/` and `multi-llm/` sections present in `ko/` | Translation lag per CLAUDE.local.md §17.3 | Medium | S | No | W1.6 §12.2 |

---

## Gap Totals by Severity

| Severity | Count |
|----------|------:|
| Critical | 9 |
| High | 32 |
| Medium | 55 |
| Low | 98 |
| **Total** | **194** |

Note: many "Low" rows represent items where the CC capability is non-applicable to moai-adk (SDK, MCP server, bridge, REPL). These are retained for audit completeness but are classified Tier 4 in the roadmap.

## Source Legend

- `W1.1` = `.moai/design/v3-research/findings-wave1-hooks-commands.md` (1,148 lines)
- `W1.2` = `.moai/design/v3-research/findings-wave1-query-context.md` (1,083 lines)
- `W1.3` = `.moai/design/v3-research/findings-wave1-agent-team.md` (1,015 lines)
- `W1.4` = `.moai/design/v3-research/findings-wave1-ui-ux.md` (1,304 lines)
- `W1.5` = `.moai/design/v3-research/findings-wave1-bootstrap-cli.md` (1,202 lines)
- `W1.6` = `.moai/design/v3-research/findings-wave1-moai-current.md` (909 lines)

End of gap matrix. See `priority-roadmap.md` for tier assignments and `v3-themes.md` for architectural organization.
