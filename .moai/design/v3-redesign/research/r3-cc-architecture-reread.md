# R3 — Claude Code Architecture Re-read (Fresh Lens)

> Research team: R3
> Basis: v3-legacy/research/findings-wave1-*.md + source-map verification
> Date: 2026-04-23
> Purpose: Move past gap-counting; extract the architectural essence of Claude Code so moai v3 can choose what to adopt, diverge from, and deliberately ignore.

---

## Executive Summary

Claude Code (CC) is best understood not as a CLI, not as a TUI, and not as an agent — it is an **Orchestration Runtime for Model Turns**. Its core abstraction is the `QueryEngine` — a stateful per-conversation object whose job is to translate a user prompt into a sequence of (model stream → tool execution → context compaction → hook dispatch) cycles, under strict budget, permission, and safety envelopes. Every other subsystem (hooks, agents, teams, bridge, permissions, memory, UI) hangs off this runtime as a layer with clearly defined entry and exit contracts. The reason CC looks "sprawling" (803KB main.tsx, 250+ React components, 27 hook events, 52 subcommands, 3 plugin origins, 110+ Zod schemas) is that each layer is deliberately extensible through declarative configuration rather than code — the runtime itself is small; the **surface area of the extensibility boundary** is huge.

The second key insight is that CC's architecture is organized around **provenance-scoped configuration with cache-prefix determinism**. Settings have 6+ sources (policy → user → project → local → plugin → skill → session → builtin), hooks dedup by source, plugins install to 3 scopes, memory is gated on canonical git root, and system prompts layer `[defaultSystemPrompt, memoryMechanics?, appendSystemPrompt?]` in that exact order so that `(systemPrompt, userContext, systemContext)` forms a stable API cache-key prefix. Prompt caching is not a feature bolted on — it is a **load-bearing invariant** that shapes what can be rewritten mid-session and what cannot.

The third insight is that CC's "big ideas" are **permission bubbling** and **sub-agent context isolation as a primitive**. Every tool call passes through a multi-layer permission check that can be satisfied by any tier (built-in allow, session rule, project rule, user rule, policy rule, hook-returned allow). Every sub-agent runs with its own system prompt, tool pool, permission mode, CWD, MCP servers, file-state cache, and transcript — yet shares the same Node.js process. This fusion of **strict isolation without process overhead** is what enables CC's fan-out (20+ parallel agents), its worktree model, and its async/background execution model. moai-adk today lacks both — it has neither a permission bubbling model nor a first-class sub-agent isolation primitive. These two mechanisms (not "more hooks" or "more commands") are the structural choices moai v3 most needs to study.

---

## 1. CC Logical Architecture

### 1.1 Core abstractions and relationships

CC's architecture is organized around 11 fundamental abstractions. The ones that matter, in dependence order:

| Abstraction | Responsibility | Lifetime | Source |
|---|---|---|---|
| **AppState** | Global per-process state: permission context, teams map, inboxes, sessionHooks, agentNameRegistry, inline plugins, registered hooks, invoked skills, allowed channels, beta-header latches | Process | `bootstrap/state.ts:277-429, 1239-1749` |
| **Session** | Per-conversation identity: sessionId, cwd, projectRoot, parentSessionId | Conversation | `bootstrap/state.ts:431-533` |
| **QueryEngine** | Turn-owner: mutable messages, abortController, permissionDenials, totalUsage, fileStateCache, discoveredSkillNames, loadedNestedMemoryPaths | Conversation | `QueryEngine.ts:184-1177` |
| **Turn** (`submitMessage`) | One user→model→tools→result cycle: process user input, build system prompt, loop through model stream, dispatch tools, run hooks, persist transcript, yield result | ~seconds | `QueryEngine.ts:209-1156` |
| **Tool** | Pure function on `(input, toolUseContext) → ContentBlock[]`, with input schema, permission definition, and (optional) permission prompt UI component | Per-call | `tools/*` |
| **Hook** | Declarative runtime intercept: 27 events × 4 types (command/prompt/http/agent) × 7 matcher strategies × optional `if` condition × exit-code-OR-JSON protocol | Per-event | `utils/hooks.ts:747-5022` |
| **Permission** | Multi-source rule stack: `policySettings > userSettings > projectSettings > localSettings > pluginSettings > sessionRules > hookDecision` — resolves to `allow | ask | deny` with optional updatedInput | Per-tool-use | `utils/permissions/*` |
| **Agent** | Isolated sub-turn: own system prompt, tool pool, permission mode, MCP servers, CWD, transcript, file-state — spawned in-process via AsyncLocalStorage or in a tmux pane | Seconds-hours | `tools/AgentTool/AgentTool.tsx:239-1200` |
| **Team** | Coordination envelope: TeamFile + mailboxes + tasks + teammates map — 1:1 with a task list | Session-to-session | `utils/swarm/teamHelpers.ts`, `tools/TeamCreateTool/*` |
| **Memory** | Typed project-scoped auto-memory: MEMORY.md index + typed topic files (user/feedback/project/reference), with LLM-selected relevance and age-based freshness staleness | Long-lived | `memdir/*` |
| **Plugin** | Configuration bundle: manifest + agents + skills + commands + hooks + mcpServers + outputStyles, loaded from builtin/installed/inline origin at three scopes (user/project/local) | Install-time | `plugins/builtinPlugins.ts`, `utils/plugins/*` |

**Relationship shape** (UML-ish):

```
AppState (singleton)
  ├── owns Session (1..N, switchable via switchSession)
  │   └── owns QueryEngine (1 per active turn)
  │       └── runs Turn (submitMessage pipeline)
  │           ├── loads Memory (memdir.loadMemoryPrompt) ──────────►
  │           ├── loads CLAUDE.md tree (context.getClaudeMds) ────►  system prompt prefix
  │           ├── loads git status (context.getGitStatus) ─────────►
  │           ├── dispatches Hooks (27 events, declarative) 
  │           ├── calls Tool (via canUseTool + Permission stack)
  │           │   └── may spawn Agent (sync/async, isolated context)
  │           │       └── may spawn Team (TeamCreate → teammates → mailboxes)
  │           │           └── teammates SendMessage via mailbox OR bridge
  │           └── may trigger Compaction (snip/micro/collapse/auto/reactive)
  └── owns Plugins (builtin + installed + inline)
      └── contribute Agents, Skills, Commands, Hooks, MCP, OutputStyles
          (merged into scope-layered settings stack)
```

Evidence: agent-team.md §0 (three-layer activation), query-context.md §1.2 (QueryEngine state), hooks-commands.md §3 (source precedence), bootstrap-cli.md §1 (three-layer launch).

### 1.2 Control flow for a typical turn

ASCII trace of the hot path (user types a prompt, Claude runs 1 tool, responds):

```
 ┌──────────────────────────────────────────────────────────────┐
 │ 1. cli.tsx shim — fast-path check (--version returns here)    │
 └───────┬──────────────────────────────────────────────────────┘
         ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ 2. main.tsx commander preAction hook                           │
 │    - runMigrations() (versioned, 11 migrations)                │
 │    - loadRemoteManagedSettings() (async, fire-and-forget)      │
 │    - loadPolicyLimits() (async)                                │
 │    - uploadUserSettingsInBackground() (async, flag-gated)      │
 └───────┬──────────────────────────────────────────────────────┘
         ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ 3. init() memoized — configs, CA certs, signal handlers,      │
 │    preconnect, LSP cleanup registration                       │
 └───────┬──────────────────────────────────────────────────────┘
         ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ 4. showSetupScreens (onboarding, trust dialog, etc.)           │
 └───────┬──────────────────────────────────────────────────────┘
         ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ 5. REPL / --print / --resume path → QueryEngine.submitMessage │
 │    ┌────────────────────────────────────────────────────────┐│
 │    │ 5a. processUserInput (slash commands, @imports)        ││
 │    │ 5b. fetchSystemPromptParts (parallel):                 ││
 │    │     - defaultSystemPrompt                              ││
 │    │     - userContext (CLAUDE.md + date)                   ││
 │    │     - systemContext (git status)                       ││
 │    │ 5c. assemble: [default, memoryMechanics?, append?]      ││
 │    │     (ordering fixes API cache-prefix)                  ││
 │    │ 5d. persist transcript BEFORE API call (--resume-safe) ││
 │    │ 5e. yield 'system/init' message (tools/model/skills)    ││
 │    │ 5f. if local-command fast-path: yield result, return    ││
 │    │ 5g. enter queryLoop:                                    ││
 │    │     ┌──────────────────────────────────────────────┐   ││
 │    │     │ startRelevantMemoryPrefetch (async)          │   ││
 │    │     │ startSkillDiscoveryPrefetch (async)          │   ││
 │    │     │ getMessagesAfterCompactBoundary              │   ││
 │    │     │ applyToolResultBudget                        │   ││
 │    │     │ snipCompact → microcompact → contextCollapse │   ││
 │    │     │ autoCompactIfNeeded                          │   ││
 │    │     │ calculateTokenWarningState (blocking limit)  │   ││
 │    │     │ callModel (streaming):                       │   ││
 │    │     │   - StreamingToolExecutor if enabled         │   ││
 │    │     │   - FallbackTriggeredError handles overload  │   ││
 │    │     │   - PromptTooLong → reactive compact         │   ││
 │    │     │   - max_output_tokens → escalate+retry (3×)  │   ││
 │    │     │ executePostSamplingHooks (fire-and-forget)   │   ││
 │    │     │ if no follow-up:                             │   ││
 │    │     │   - handleStopHooks (with blocking nudge)    │   ││
 │    │     │   - checkTokenBudget (90% continuation)      │   ││
 │    │     │   - return {reason:'completed'}              │   ││
 │    │     │ else runTools (with canUseTool permission)   │   ││
 │    │     │   - permission stack resolves                │   ││
 │    │     │   - PermissionRequest hook fires             │   ││
 │    │     │   - PreToolUse hook fires (can block/rewrite)│   ││
 │    │     │   - tool executes                            │   ││
 │    │     │   - PostToolUse hook fires (can rewrite out) │   ││
 │    │     │ generateToolUseSummary (Haiku, async)        │   ││
 │    │     │ consume memory+skill prefetches              │   ││
 │    │     │ check maxTurns; increment turnCount          │   ││
 │    │     │ loop                                         │   ││
 │    │     └──────────────────────────────────────────────┘   ││
 │    │ 5h. emit terminal result with aggregated usage         ││
 │    └────────────────────────────────────────────────────────┘│
 └──────────────────────────────────────────────────────────────┘
```

Evidence: bootstrap-cli.md §1.1-1.7, query-context.md §1.3-1.4.

### 1.3 Data flow for configuration and state

```
Configuration sources (priority high→low, per hooksSettings.ts:15-228):
  1. policySettings        ~/.config/claude-code/managed-settings.json (MDM)
  2. userSettings          ~/.claude/settings.json
  3. projectSettings       <cwd>/.claude/settings.json
  4. localSettings         <cwd>/.claude/settings.local.json
  5. pluginHook            plugin.hooksConfig (per enabled plugin)
  6. skillHook             skill frontmatter (session-scoped)
  7. sessionHook           addSessionHook() (ephemeral)
  8. builtinHook           registered via registerHookCallbacks()

Each source can contribute:
  - hooks (27 events × 4 types)
  - permissions rules (allow/deny/ask)
  - MCP server definitions
  - env vars
  - extra known marketplaces
  - model pins
  - output-style selection
  - plugin enablement
  - memory directory override

Read path:
  getMergedSettings(cwd) → dynamic overlay
  getHooksConfigFromSnapshot() → frozen at startup + plugin hot-reload
  getRegisteredHooks() → runtime-added (SDK, plugin JS)
  getSessionHooks() → skill/agent-invocation scoped

Write path (migrations/runtime):
  updateSettingsForSource(source, fn)
  saveGlobalConfig(fn)    (GlobalConfig ~/.claude.json, NOT settings)
  addSessionHook()         (memory only, no disk)

Policy gates (hooksConfigSnapshot.ts):
  - policySettings.disableAllHooks=true → {}
  - policySettings.allowManagedHooksOnly=true → policy only
  - mergedSettings.disableAllHooks=true (non-managed) → managed only
  - strictPluginOnlyCustomization → policy + plugin only
```

Evidence: hooks-commands.md §3.2-3.4.

---

## 2. CC Design Decisions (what works)

The 15 most consequential decisions CC made, each with the alternative it rejected and why the chosen path wins:

1. **QueryEngine as a stateful class, not a reducer.**
   Alternative: pure functional `reduce(state, event)` pattern.
   Why this works: turns generate dozens of side effects (transcript writes, MCP calls, hook spawns, worktree creations) whose ordering must survive errors and aborts. A class with `abortController`, `mutableMessages.splice()` at compact boundary, and `using pendingMemoryPrefetch` disposal gives predictable cleanup without ceremony. Evidence: `QueryEngine.ts:925-933` compact boundary GC.

2. **Cache-prefix discipline on system prompts.**
   Alternative: build prompts fresh each turn for maximum flexibility.
   Why this works: Anthropic's prompt caching requires byte-stable prefixes. CC freezes `(systemPrompt, userContext, systemContext)` ordering and uses `memoize()` on each piece. Cache-hit-rate is a first-class quality metric; flexibility is a second-class one. Evidence: query-context.md §3.7.

3. **Layered compaction (5 strategies) instead of one.**
   Alternative: single summarization pass at 90% window.
   Why this works: different pressure regimes need different tools — zombie-message removal (snip), cache-trimming (microcompact), projection (collapse), proactive summary (autocompact), reactive error recovery (reactive). Combining them yields graceful degradation instead of a hard cliff. Evidence: query-context.md §2.8.

4. **3-failure circuit breaker on autocompact.**
   Alternative: unlimited retries.
   Why this works: BigQuery data showed 1,279 sessions had 50+ consecutive failures wasting ~250K API calls/day. A circuit breaker with `MAX_CONSECUTIVE_AUTOCOMPACT_FAILURES=3` bounds the tax. Failure modes must be capped, not trusted. Evidence: `autoCompact.ts:70`, query-context.md §2.1.

5. **Hook output protocol is JSON-OR-exitcode, not just exitcode.**
   Alternative: exit codes only (simpler, shell-friendly).
   Why this works: rich hook capabilities (`additionalContext`, `permissionDecision`, `updatedInput`, `watchPaths`) require structured output. But exit codes remain for scripts without JSON producers. Dual protocol is strictly more expressive without excluding bash-script authors. Evidence: hooks-commands.md §4.1-4.2.

6. **Four hook types with clear cost ladders.**
   Alternative: just command (shell) hooks.
   Why this works: `command` (cheap subprocess), `prompt` (cheap LLM yes/no), `agent` (expensive LLM-with-tools), `http` (SSRF-guarded webhook). Each is a tool for a different problem; the 4-way choice lets users pick the right grain. Evidence: hooks-commands.md §2.

7. **Hook matcher is string-OR-regex-OR-pipelist, not one syntax.**
   Alternative: always regex.
   Why this works: most matchers are trivial (`Write|Edit`) and users write regex only when needed. The `utils/hooks.ts:1346-1381` 3-way resolver handles the 95% case with 10× less ceremony. Evidence: hooks-commands.md §3.8.

8. **Sub-agent isolation via AsyncLocalStorage, not process boundary.**
   Alternative: spawn each agent in its own Node process.
   Why this works: AsyncLocalStorage provides per-agent `agentId`, `agentType`, `workerTools`, `permissionMode`, `cwd` without fork overhead. In-process teammates can share 36GB whale sessions with controlled per-agent caps (`TEAMMATE_MESSAGES_UI_CAP=50`). Evidence: agent-team.md §7.2.

9. **Worktree isolation as an orthogonal axis to agent isolation.**
   Alternative: agents that write files always get their own process.
   Why this works: context (system prompt, tools, permissions) is independent from filesystem (writes to isolated branch). A read-only agent needs context isolation but NOT worktree; a fork subagent needs fast shared context AND shared worktree. Decoupling the axes yields 4 deliberate combinations, not one blunt policy. Evidence: agent-team.md §5.3, §7.5.

10. **Versioned migrations with preAction auto-apply.**
    Alternative: explicit `claude migrate` command.
    Why this works: users never need to know migrations exist. `CURRENT_MIGRATION_VERSION = 11` + preAction hook + per-migration idempotency guards mean CC silently rolls forward every user without a single support ticket. Compare: moai-adk ships `moai migrate agency` as an explicit command users must remember. Evidence: bootstrap-cli.md §5.

11. **Provenance on every configuration element.**
    Alternative: flat merged config.
    Why this works: `settingSource`, `pluginRoot`, `skillRoot`, `policyPinned` flow through every hook, command, and permission rule. This enables: showing "this rule came from which file" in `/permissions`, blocking user edits when `strictPluginOnlyCustomization`, and migrating `userSettings` without touching `projectSettings` (which may be attacker-controlled). Evidence: hooks-commands.md §3.2, bootstrap-cli.md §5.4.

12. **Plugin system with 3 origins, each with different trust/lifetime.**
    Alternative: one plugin kind.
    Why this works: `builtin` (bundled with binary, always trusted), `installed` (from marketplace, persistent, user-toggleable), `inline` (`--plugin-dir`, session-only, dev loop). Each origin has a natural trust level and lifetime that matches its use case. Evidence: bootstrap-cli.md §4.1.

13. **Memory is LLM-selected, not grep-selected.**
    Alternative: regex-match CLAUDE.md against prompt.
    Why this works: `findRelevantMemories` runs a dedicated Sonnet sideQuery with `selectRelevantMemories` to pick up to 5 memories per turn. Manifest is a compact header-only scan (`scanMemoryFiles`, 30-line cap). This matches the scale and fuzziness of human knowledge better than pattern-matching. Evidence: query-context.md §6.10.

14. **Memory freshness is staleness-aware.**
    Alternative: always present memories as current fact.
    Why this works: `memoryFreshnessNote` wraps stale memories (>1 day) in `<system-reminder>` with explicit caveat. Rationale: "citation makes the stale claim sound more authoritative, not less." This is a behavioral guardrail baked into the memory payload itself. Evidence: query-context.md §6.12.

15. **Permission mode `bubble` for fork agents.**
    Alternative: fork agents run with parent's exact mode.
    Why this works: a fork that inherits context SHOULD ask the parent-terminal's user for permission (bubble), not the teammate's mailbox. `bubble` is a permission *mode* value, which means it's a first-class primitive, not a boolean special case. Evidence: agent-team.md §4.3.

Bonus: **`@commander-js/extra-typings` + declarative option chains in main.tsx.** The 803KB main.tsx is intimidating, but the subcommand registration is effectively a declarative config table with 52 entries + 60 options. Dispatch is data, not code. Moving it to a switch statement would not improve anything.

---

## 3. CC Limitations & Technical Debt

Real architectural weaknesses (not features-could-be-added):

1. **main.tsx is a single 803KB function-scoped registration block.**
   The preAction hook, root option chain, and 52 `.command()` calls live inline in one action handler. Testing requires mocking commander; dependency injection is absent. Breaking it up by subcommand would lose DCE benefits (`"external" === 'ant'` string literal comparison relies on the whole file being visible to the minifier). This is a **bundler-driven architecture decision**, not laziness — but it makes main.tsx a hot-spot for conflicts. Evidence: bootstrap-cli.md §8.5.

2. **Windows parity gaps are shell-abstraction debt.**
   Hooks use per-hook `shell: 'bash' | 'powershell'` (hooks-commands.md §4.5) but:
   - Shell-prefix conversion happens only for bash (`CLAUDE_ENV_FILE`, `windowsPathToPosixPath`).
   - Git-bash on Windows requires `setShellIfWindows()` detection.
   - `SUPPORTS_TERMINAL_VT_MODE` branch (Node 22.17 / Bun 1.2.23) means `shift+tab` fallback binding on older Windows.
   The platform abstraction is "bash with per-call fixups" — not a `Terminal` interface. Evidence: ui-ux.md §5.2, bootstrap-cli.md §1.7.

3. **Feature-flag gating produces a combinatorial documentation gap.**
   Of 52 subcommands, many are `feature('TEMPLATES')`, `feature('BG_SESSIONS')`, `feature('KAIROS')`-gated (bootstrap-cli.md §8.6). 16 more are `"external" === 'ant'` (ant-only). Published docs cannot accurately describe "what commands are available" because the answer depends on GrowthBook flags and build channel. Source:documentation mismatch is baked into the build system.

4. **27 hook events is almost certainly past the comprehension ceiling.**
   BaseHookInput + per-event payload + matcher + exit-code table + hookSpecificOutput reply shape × 27 = 1,000+ distinct rules. No single human fully holds this model. Evidence: hooks-commands.md §1.2 lists each event with 6-10 attributes, ~300 lines to enumerate. Some events (`StopFailure`, `InstructionsLoaded`, `FileChanged`) are used by <0.1% of users.

5. **SessionEnd 1500ms timeout is a leaky abstraction.**
   Most hooks get `TOOL_HOOK_EXECUTION_TIMEOUT_MS = 10 min`, but SessionEnd is capped at 1.5s with a dedicated env override `CLAUDE_CODE_SESSIONEND_HOOKS_TIMEOUT_MS`. This reflects the implementation reality (process is about to exit) but the contract change (10 min → 1.5s based solely on event type) is invisible to hook authors who wrote a slow cleanup script. Evidence: hooks-commands.md §4.7.

6. **Memory path resolution has 4+ fallbacks, each with different semantics.**
   `CLAUDE_COWORK_MEMORY_PATH_OVERRIDE` > `autoMemoryDirectory` setting > `CLAUDE_CODE_REMOTE_MEMORY_DIR` env > `{base}/projects/{sanitize(findCanonicalGitRoot())}/memory/`. Each fallback was added for a specific integration (Cowork, CCR, worktree dedup gh-24382). The chain is correct but un-documented-in-one-place. Evidence: query-context.md §6.2.

7. **Plugin system is scaffolded but 0 built-in plugins ship.**
   `plugins/builtinPlugins.ts` exposes `registerBuiltinPlugin()`; `plugins/bundled/index.ts:20-23` is the entire populated list:
   ```
   export function initBuiltinPlugins(): void {
     // No built-in plugins registered yet — this is the scaffolding for
     // migrating bundled skills that should be user-toggleable.
   }
   ```
   The architecture is ready; the migration from "bundled skill" to "built-in plugin" is unfinished. Evidence: bootstrap-cli.md §4.2.

8. **Claude-API-specific assumptions leak through the stack.**
   `calculateUSDCost`, `getStoredSessionCosts`, 1M context via `[1m]` suffix, effort-level beta header, Haiku-as-sideQuery, Fast Mode routing — all baked in. A non-Anthropic provider would need to reimplement cost table, context-window resolution, model alias migration, and beta-header injection. Evidence: query-context.md §5, §8.8.

9. **Bridge architecture is cloud-tied.**
   `bridge/` is 33 files, ~400KB, and every transport path terminates at `api.anthropic.com`. The env-based variant uses `/v1/environments/bridge`; the env-less uses `/v1/code/sessions`; remote uses `wss://api.anthropic.com/v1/sessions/ws/...`. OAuth + `X-Trusted-Device-Token` + `anthropic-version` headers are unavoidable. This is not a general peer-to-peer transport; it's a cloud-relay client. Evidence: agent-team.md §1.7.

10. **Ink fork is 50 files / ~750KB.**
    CC vendored its own Ink because upstream lacks DEC 2026 synchronized output, OSC-8, Kitty keyboard protocol, and the theme provider wrapping pattern. This is the right engineering call for CC's requirements — but the TUI is unrecoverable for anyone else. Attempting to reuse "CC's rendering" outside CC means reimplementing Ink. Evidence: ui-ux.md §1.

---

## 4. Adoption Candidates (structural patterns moai should adopt)

Five structural patterns worth adopting, ordered by leverage:

1. **Multi-layer settings with explicit precedence and provenance.**
   Adopt: `policy > user > project > local > plugin > skill > session > builtin` tiers. Every config value carries a `source` tag. Merge is deterministic and traceable. moai-adk today has only `~/.moai/config/` and `.moai/config/`; adding policy (for enterprise/team rollouts), explicit session-scoped overrides, and plugin provenance yields reliable conflict resolution and better debuggability of "which file set this?". Derived benefit: migrations touching only `userSettings` is safe; plugin hooks dedupe correctly across multiple installs with similar names. Evidence: hooks-commands.md §3.2-3.9.

2. **Permission bubble/escalation model.**
   Adopt: tools execute through `canUseTool(tool, input, context) → allow | ask | deny + updatedInput?`. Permission decisions have 5+ sources (builtin, session rules, 3 setting scopes, hook decisions). Permission modes are first-class values (`default`, `acceptEdits`, `bypassPermissions`, `plan`, `auto`, `bubble`). Fork agents use `bubble` mode to push decisions to parent terminal. moai today has implicit trust ("agent is running, let it work"); adding a permission envelope would enable: pre-allowlist for common dev ops, bubble-to-user for risky cases, and programmatic rule rewriting via hook responses. Evidence: agent-team.md §4.3, hooks-commands.md §1.2.

3. **Sub-agent context isolation as a primitive.**
   Adopt: every agent has `{systemPrompt, toolPool, permissionMode, cwd, mcpServers, fileStateCache, transcript}` — all independent. Isolation mechanism can be AsyncLocalStorage (in-process) OR git worktree (filesystem) OR remote CCR (full process). moai currently spawns subagents via Claude Code's `Agent()` but has no internal primitive for "moai CLI working on SPEC X with its own tool pool." The v3 run phase would benefit from explicit context objects, not just shell-out calls. Evidence: agent-team.md §5.3, §7.1-7.5.

4. **Hook output protocol: JSON-OR-exitcode.**
   Adopt: hooks can reply with `{"additionalContext": "...", "permissionDecision": "allow|deny|ask", "updatedInput": {...}, "systemMessage": "...", "continue": false}`. Backward compatible: exit code 0/2/nonzero still works for simple scripts. Unlocks: inject context into model turn, programmatic permission gating, mid-turn rewriting of tool input, user-visible system messages. moai hooks today signal only via exit code. A JSON protocol would make Sprint Contract negotiation and MX tag injection actually programmable. Evidence: hooks-commands.md §4.1-4.2.

5. **Output style as override contract, not global setting.**
   Adopt: `.claude/output-styles/*.md` with frontmatter `{name, description, keep-coding-instructions, force-for-plugin}`. Markdown body appends to (or replaces) default coding prompt. Project-level overrides user-level. Plugins can declare `forceForPlugin: true` to auto-apply when the plugin is enabled. This is a **versioned, layered, hot-reloadable system prompt modification mechanism** — exactly what `/moai` skill body is trying to be, but formalized. moai should adopt CC's frontmatter schema directly so the artifact serializes identically in both systems. Evidence: bootstrap-cli.md §7, ui-ux.md §6.

Runners-up worth mentioning:

6. **Versioned migration framework with preAction auto-apply** (bootstrap-cli.md §5). moai-adk's `moai migrate agency` is explicit; CC's `runMigrations()` at preAction is silent. Strongly increases upgrade reliability without documentation burden.

7. **Typed memory taxonomy (user/feedback/project/reference)** (query-context.md §6.6). moai MEMORY.md informally already uses this schema; formalizing it as a constitutional rule prevents drift.

8. **Startup profiler with named checkpoints** (bootstrap-cli.md §1.7). 20+ `profileCheckpoint()` markers make cold-start measurable. Adding these to `cmd/moai/main.go` would immediately surface the "moai init is slow" complaint as concrete wall-clock numbers per stage.

---

## 5. Divergence Candidates (what moai should NOT copy)

Five things CC does that moai should deliberately not replicate:

1. **Do not copy the Ink TUI fork.**
   CC's TUI is ~750KB of vendored source with DEC 2026 synchronized output and Kitty keyboard protocol. moai-adk runs *inside* CC (as a subprocess or hook callee); rebuilding a TUI in Go would duplicate work and create conflicts. Instead: emit structured markdown + ANSI that CC's renderer handles, plus JSONL for machine-readable outputs. Divergence: moai stays headless. Evidence: ui-ux.md §8.1.

2. **Do not copy Claude-only model assumptions.**
   CC's cost table, 1M context via `[1m]` suffix, effort-level beta, Haiku-as-sideQuery, and FallbackTriggeredError handling all assume Anthropic API surface. moai supports GLM in CG mode and could theoretically support other providers; hardwiring Claude would break this. Divergence: moai treats "model" as opaque name + env — no provider-specific branches in the run loop. Evidence: query-context.md §5, §8.8.

3. **Do not copy the bridge/remote-control architecture.**
   33 files, 4 transport layers, OAuth, trusted device token — all point at `api.anthropic.com`. moai's CG mode achieves similar cross-machine parallelism via tmux sessions (local) and uses Claude Code's existing bridge when needed. Building a moai-native bridge would require either Anthropic partnership or a parallel infra stack. Divergence: moai delegates cross-machine to CC's bridge; moai focuses on local dev. Evidence: agent-team.md §1.

4. **Do not copy GrowthBook / feature-flag gating.**
   CC uses runtime feature flags (`feature('UDS_INBOX')`, `feature('KAIROS')`, `feature('BG_SESSIONS')`) with build-time DCE. This requires an analytics backend (`logEvent('tengu_*')`), a flag service, and user segmentation. moai is open-source and deployed-as-binary; feature flags would mean lying to users about available commands. Divergence: moai enables/disables via explicit config keys (quality.yaml, workflow.yaml), not opaque build flags. Evidence: bootstrap-cli.md §8.6.

5. **Do not copy the centralized MCP server registry.**
   CC has `getMcpServerStatus`, enterprise-managed MCP, project approval via `mcp.json`, and a full `plugin marketplace` with 5 source types — all pointing at a central registry model. moai treats MCP servers as per-project opt-in (`.mcp.json` is user-owned) and does not need a marketplace. Divergence: moai keeps plugin distribution via git clone / template copy, not registry. Evidence: bootstrap-cli.md §4.5.

Runners-up:

6. **Do not ship a 52-subcommand CLI.** CC's scope (daemon, bg sessions, templates, self-hosted runner, ssh, environment-runner) reflects internal tooling + enterprise deployment. moai's ~11 subcommands are the right shape.

7. **Do not adopt a 3-failure circuit breaker on arbitrary operations.** It's load-bearing for autocompact (which calls paid APIs); most moai operations are local and should fail fast instead.

8. **Do not replicate the 200KB LogSelector or the 116KB BackgroundTasksDialog.** These are CC-specific UX surfaces that have no home in a headless CLI.

---

## 6. CC's "Big Idea" vs moai's "Big Idea"

**CC's big idea, in one paragraph:** Claude Code is a **safety-enveloped model turn runtime** — everything in the stack exists to make it safe and performant to run an open-ended model loop inside an end-user's terminal on their actual filesystem. The QueryEngine is the nucleus; permission envelopes (5 sources, 6 modes, bubble escalation), hooks (27 events, 4 types, JSON or exit protocol), compaction (5 strategies, circuit breaker), sub-agent isolation (in-process, worktree, or remote), and typed memory (with LLM selection and freshness caveats) are concentric rings of protection and performance optimization around that nucleus. Everything is declarative, everything has provenance, everything can be scoped to project/user/session/policy. The product bet is that *the runtime is the platform*: if the turn loop is robust enough, any capability (agents, teams, design, security, feedback) can be added as configuration.

**moai's big idea, in one paragraph (divergent):** moai-adk is a **SPEC-governed development workflow orchestrator** — its nucleus is not a turn runtime, it is the Plan→Run→Sync→Review pipeline anchored in SPEC documents with EARS-format requirements, TRUST 5 quality gates, and DDD/TDD methodology. Where CC asks "how do I safely run one more model turn?", moai asks "how do I ship the SPEC correctly?" CC's loop is stateless-per-turn with per-turn safety; moai's loop is stateful-per-SPEC with per-phase quality gates. CC treats agents as tools; moai treats the SPEC as the contract and uses agents to execute it. The product bet is that *the SPEC is the source of truth*: if the SPEC is well-formed and each phase verifies against it, the resulting codebase is correct by construction. moai rides CC's runtime as a hosting layer; it does not try to replace the runtime. This is why moai v3 should adopt CC's **settings precedence, permission bubbling, output style contract, and hook JSON protocol** (structural patterns that make workflows more reliable) and deliberately NOT adopt CC's **bridge, feature-flag system, or Ink TUI** (implementation details specific to being a terminal agent product).

---

## References (file:line from legacy findings + source-map)

Legacy findings (primary input):
- `/Users/goos/MoAI/moai-adk-go/.moai/design/v3-legacy/research/findings-wave1-hooks-commands.md` — sections §1 (27 events), §2 (4 hook types), §3 (source precedence), §4 (output protocol), §5 (commands).
- `/Users/goos/MoAI/moai-adk-go/.moai/design/v3-legacy/research/findings-wave1-query-context.md` — §1 (QueryEngine/queryLoop), §2 (autocompaction), §3 (context loading), §5 (cost tracker), §6 (memdir), §7 (migrations).
- `/Users/goos/MoAI/moai-adk-go/.moai/design/v3-legacy/research/findings-wave1-agent-team.md` — §0 (3 orthogonal systems), §1 (bridge), §4 (agent frontmatter), §5 (sub-agent dispatch), §6 (Team API), §7 (isolation & backends).
- `/Users/goos/MoAI/moai-adk-go/.moai/design/v3-legacy/research/findings-wave1-ui-ux.md` — §1 (Ink fork), §5 (keybindings), §7 (permission UI).
- `/Users/goos/MoAI/moai-adk-go/.moai/design/v3-legacy/research/findings-wave1-bootstrap-cli.md` — §1 (startup), §2 (CLI), §4 (plugin system), §5 (migrations), §6 (schemas), §7 (output styles), §8 (main.tsx probe).

Source-map anchors (verification):
- `entrypoints/cli.tsx:1-303` — fast-path shim, 13 early returns.
- `entrypoints/init.ts:57-238` — memoized init orchestrator.
- `bootstrap/state.ts:277-429` — AppState shape.
- `QueryEngine.ts:184-1177` — per-conversation engine class.
- `query.ts:241-1728` — the queryLoop (tool dispatch, compaction orchestration, hook firing).
- `services/compact/autoCompact.ts:30-351` — thresholds, formulas, circuit breaker.
- `utils/hooks.ts:747-5022` — the 5022-line hook runner (spawn, dedup, match, dispatch).
- `utils/hooks/hooksSettings.ts:15-228` — 8-tier source precedence.
- `entrypoints/sdk/coreTypes.ts:25-53` — HOOK_EVENTS constant (27 events).
- `entrypoints/sdk/coreSchemas.ts:355-949` — all Zod schemas for hook inputs/outputs.
- `schemas/hooks.ts:32-213` — 4 hook types (command/prompt/http/agent).
- `tools/AgentTool/AgentTool.tsx:82-1200` — input schema, dispatch tree, async/sync paths.
- `tools/AgentTool/loadAgentsDir.ts:73-755` — agent frontmatter schema.
- `tools/SendMessageTool/SendMessageTool.ts:67-917` — target routing (name/broadcast/uds/bridge + structured messages).
- `utils/swarm/backends/registry.ts:74-451` — 3 teammate backends (tmux/iterm2/in-process).
- `memdir/memdir.ts:34-507` — entrypoint truncation, `loadMemoryPrompt` dispatcher.
- `memdir/memoryTypes.ts:14-106` — 4 memory types with prompts.
- `memdir/findRelevantMemories.ts:18-141` — Sonnet sideQuery selector.
- `memdir/memoryAge.ts:6-53` — freshness staleness caveat.
- `main.tsx:325-352` — `CURRENT_MIGRATION_VERSION = 11` + runner.
- `main.tsx:903-1006` — commander preAction hook + root options.
- `plugins/builtinPlugins.ts:21-159` — built-in plugin registry.
- `outputStyles/loadOutputStylesDir.ts:26-98` — output style loader.
- `ink.ts:18-85` — Ink entry + re-exports (ThemedBox/Text).
- `keybindings/defaultBindings.ts:32-340` — 18 contexts × ~85 actions.

End of R3 re-read.
