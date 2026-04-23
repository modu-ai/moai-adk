# MoAI-ADK v3 — Priority Roadmap

Generated: 2026-04-22
Source: `gap-matrix.md` (194 items) organized into tiered delivery plan.

## Purpose

This roadmap groups the 194 gaps identified in Wave 1 into four delivery tiers. Tier
boundaries are deliberately firm: Tier 1 items MUST ship in v3.0.0 for the release to
credibly claim parity with Claude Code's core capabilities. Tier 4 items are explicitly
out of scope with documented rationale to prevent re-litigation.

## Tier Definitions

- **Tier 1 — Parity Foundations**: MUST be in v3.0.0. Absence blocks the "CC-parity" brand claim.
- **Tier 2 — Strategic Differentiators**: SHOULD be in v3.0.0. These are moai-unique value-adds or high-UX-impact additions that justify the major version bump.
- **Tier 3 — Polish & Cleanup**: TARGETED for v3.0.x patches or v3.1.0. Internal tech debt and low-priority UX polish.
- **Tier 4 — Defer / Reject**: EXPLICITLY out of scope. Includes CC-specific infrastructure (bridge) and features moai doesn't need (SDK, REPL).

## ID Convention

`{TIER}-{DOMAIN}-{NN}`. Domains: HOOK, MEM, MIG, SCH, AGT, SKL, CLI, PLG, OUT, CLN (cleanup), TEAM, DIFF (differentiator).

Every item references gap-matrix.md gap numbers (`gm#NN`) for traceability.

---

## Tier 1 — Parity Foundations (v3.0.0 required)

### T1-HOOK-01 — Hook Protocol v2 (rich JSON output)
Rationale: Enables `additionalContext`, `permissionDecision`, `updatedInput`, `watchPaths`, `stopReason`, `systemMessage`. Without this, moai hooks cannot participate in the CC ecosystem.
Covers: gm#4, gm#5, gm#6, gm#7
Dependencies: blocked-by none; blocks T1-HOOK-02..05, T2-HOOK-10..14
Risk: Breaking change — existing exit-code-only hooks need a grace period. Plan: v3.0 dual-parse (new JSON first, fall back to exit-code), v3.2 warn, v4.0 remove old path.

### T1-HOOK-02 — Hook `if` condition filter
Rationale: Prevent unnecessary subprocess spawns (e.g., only run format-hook when tool is `Write` or `Edit`). Reduces idle-session overhead.
Covers: gm#8
Dependencies: none
Risk: Low; additive.

### T1-HOOK-03 — Hook `async` / `asyncRewake` support
Rationale: Unlocks long-running Ralph-style loops and non-blocking notifications. Required for future event-driven extensions.
Covers: gm#9, gm#10
Dependencies: T1-HOOK-01 (JSON output includes `{"async":true,"asyncTimeout":N}` first-line protocol)
Risk: Medium; requires `AsyncHookRegistry` equivalent in Go.

### T1-HOOK-04 — Hook `once: true` flag
Rationale: Enables session-scoped one-shot hooks (e.g., first-run setup audit).
Covers: gm#11
Dependencies: T1-HOOK-05 (source precedence enables session-scoped hooks)
Risk: Low.

### T1-HOOK-05 — Hook source precedence (3-tier minimum)
Rationale: Without source precedence, plugin hooks and user hooks cannot coexist. v3.0 minimum: user / project / local. Policy/plugin/skill/session layers added incrementally.
Covers: gm#15 (scope-reduced)
Dependencies: T1-SCH-02 (schemas must exist for each source)
Risk: Breaking change for users relying on override-resolution order. Document precedence clearly.

### T1-MEM-01 — MEMORY.md truncation + freshness alignment with CC memdir
Rationale: moai already uses CC's 4-type taxonomy (user/feedback/project/reference). Adding `MAX_ENTRYPOINT_LINES=200` / `MAX_ENTRYPOINT_BYTES=25000` truncation with warning appendix aligns with CC's memory contract.
Covers: gm#44, gm#45, gm#50
Dependencies: none
Risk: Low; additive — existing MEMORY.md files may need grooming (user-facing notice on first truncation).

### T1-MIG-01 — Versioned migration framework
Rationale: moai has 104 SPECs, 50 skills, 22 agents — the config surface is too large to rely on one-shot migrations forever. Establishes upgrade safety net for all future releases.
Covers: gm#149, gm#150, gm#151
Dependencies: T1-SCH-01
Risk: Breaking change in upgrade semantics — migrations become implicit on every command run (breaks airgapped expectations). Opt-out flag `MOAI_DISABLE_MIGRATIONS=1` for CI.

### T1-MIG-02 — Initial migration set (5 baseline migrations)
Rationale: Establish migration framework credibility by shipping with real migrations:
- M01: `project.yaml.template_version` auto-sync from `system.yaml.moai.version`
- M02: `.agency/` vestigial removal (stub redirects, constitution stub)
- M03: Hook wrapper drift resolver (ensures local matches template)
- M04: Skill drift resolver (deploy 3 template-only skills locally)
- M05: `.moai-backups/` archival to `~/.moai/history/`

Covers: gm#183, gm#184, gm#185, gm#190, gm#191
Dependencies: T1-MIG-01
Risk: Medium; if a migration misbehaves, users lose confidence. Each migration logs before/after state; dry-run mode.

### T1-SCH-01 — Formal config schemas for `.moai/config/sections/*.yaml`
Rationale: Zero schema = zero validation. Adopt Go struct tags with `go-playground/validator/v10` + JSON Schema export for editor integration.
Covers: gm#156, gm#157, gm#163
Dependencies: none
Risk: Breaking change — existing configs with typos may now fail. Ship with `moai doctor config` auto-repair mode.

### T1-SCH-02 — Settings source layering (3-tier)
Rationale: Enterprise and multi-user scenarios require layered config. Minimum v3.0: user (`~/.moai/config/`), project (`.moai/config/`), local (`.moai/config/**/*.local.yaml`).
Covers: gm#138, gm#164 (scope-reduced from 6 sources)
Dependencies: T1-SCH-01
Risk: Breaking for users who expect flat override. Precedence order documented in CLAUDE.md.

### T1-AGT-01 — Agent frontmatter v2 bundle
Rationale: Ships 6 new frontmatter fields in one release (small individually, massive together):
- `memory: user|project|local` (persistent agent memory scope)
- `initialPrompt` (first-turn injection)
- `requiredMcpServers` (availability gating)
- `omitClaudeMd` (token savings for read-only agents)
- `maxTurns` (hard cap)
- `criticalSystemReminder_EXPERIMENTAL` (per-turn reminder)

Covers: gm#56, gm#57, gm#58, gm#59, gm#60, gm#62
Dependencies: T1-SCH-01 (schema for agent frontmatter)
Risk: Low; all fields are additive. `omitClaudeMd` is a major token win for Explore-style agents.

### T1-SKL-01 — Skill frontmatter v2 bundle
Rationale: Ships 4 skill frontmatter additions:
- `paths:` conditional activation (glob + gitignore-style filter)
- `effort:` per-skill Opus 4.7 override
- `$ARGUMENTS[N]` / `$name` argument substitution
- `${CLAUDE_SKILL_DIR}` / `${CLAUDE_SESSION_ID}` body substitution

Covers: gm#89, gm#90, gm#99, gm#100, gm#101, gm#104, gm#105
Dependencies: T1-SCH-01
Risk: Low; all additive. Existing skills unaffected.

### T1-CLN-01 — Template drift cleanup
Rationale: Resolve the 3 template-only skills not deployed locally (`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`) and `handle-permission-denied.sh` wrapper.
Covers: gm#184, gm#185
Dependencies: T1-MIG-02 (implements as migration M03/M04)
Risk: Low; moai update enhancement.

### T1-CLN-02 — template_version sync fix (critical bug)
Rationale: `project.yaml.template_version: v2.7.22` is ~12 minor versions stale while `system.yaml.moai.version: v2.12.0`. Indicates `moai update` silently skips this field.
Covers: gm#183
Dependencies: T1-MIG-02 (implements as migration M01)
Risk: Low; one-line fix + migration.

### T1-CLN-03 — Legacy code removal
Rationale: Remove `internal/cli/glm.go.bak` (28K), `internal/cli/worktree/new_test.go.bak` (13K), stale `coverage.out`/`coverage.html` (dated 2026-03-11), fix ADR-011 comment drift in `internal/template/embed.go:8-12`.
Covers: gm#186, gm#187, gm#188, gm#189
Dependencies: none
Risk: Zero; pure cleanup.

### T1-AGT-02 — `skills:` frontmatter preload (already partial)
Rationale: moai agents already use `skills:` as YAML array. Formalize the loading behavior (tool-pool merge, dedup).
Covers: gm#61
Dependencies: T1-SCH-01
Risk: Low; already in use — mostly formalization.

### T1-AGT-03 — `background: true` as frontmatter field (not just spawn param)
Rationale: CLAUDE.md already documents `background: true` at agent level; promote to formal frontmatter field parsed by deployer.
Covers: gm#63
Dependencies: T1-SCH-01
Risk: Low.

### T1-HOOK-06 — PermissionRequest/PermissionDenied handler upgrade
Rationale: moai's 27 event types are defined but handlers are observational. Upgrade `PermissionRequest` to return `decision: {behavior, updatedInput?, updatedPermissions?}` and `PermissionDenied` to return `{retry: boolean}`. Required to participate in permission ecosystem.
Covers: gm#27, gm#28
Dependencies: T1-HOOK-01
Risk: Medium; requires careful coordination with existing permission rules.

### T1-HOOK-07 — Handler richness upgrades for 6 existing event types
Rationale: Upgrade the 6 events whose handlers are observational to return structured outputs:
- PreCompact: return `{newCustomInstructions?, userDisplayMessage?}`
- ConfigChange: support exit 2 to BLOCK setting changes
- InstructionsLoaded: wire to load audit log
- Elicitation/ElicitationResult: action/content schemas
- WorktreeCreate: provider hook contract (stdout = absolute path)

Covers: gm#24, gm#29, gm#30, gm#31, gm#32
Dependencies: T1-HOOK-01
Risk: Medium; WorktreeCreate provider contract changes semantics for existing users.

---

## Tier 2 — Strategic Differentiators (v3.0.0 preferred)

### T2-HOOK-10 — Hook `type: 'prompt'` (LLM-evaluated)
Rationale: Near-zero-cost LLM-gated quality checks (Haiku-class, returns `{ok, reason?}`). Enables SPEC-validation-as-hook patterns.
Covers: gm#1
Dependencies: T1-HOOK-01
Risk: Medium; adds LLM call dependency to hook path. Configurable timeout + cost budget.

### T2-HOOK-11 — Hook `type: 'agent'` (full subagent verifier)
Rationale: Multi-turn verifier with tools (e.g., "Verify tests ran and passed"). Powerful for evaluator-active-style gates.
Covers: gm#2
Dependencies: T1-HOOK-01, T2-HOOK-10
Risk: Medium; requires subagent runtime. Respect `ALL_AGENT_DISALLOWED_TOOLS`.

### T2-HOOK-12 — Hook `type: 'http'` (webhook with SSRF guard)
Rationale: Notify Slack/Discord/observability service on commit or SPEC completion. SSRF guard + URL/env allowlist required.
Covers: gm#3
Dependencies: T1-HOOK-01
Risk: Medium; security-critical. Port CC's `ssrfGuard.ts` logic faithfully.

### T2-HOOK-13 — CLAUDE_ENV_FILE mechanism
Rationale: Hooks write bash exports that apply to subsequent BashTool commands (SessionStart/Setup/CwdChanged/FileChanged). Enables .envrc / direnv integration.
Covers: gm#20
Dependencies: T1-HOOK-01
Risk: Low; scoped to 4 specific events.

### T2-HOOK-14 — Hook source precedence (full 8-tier)
Rationale: Extend T1-HOOK-05 from 3-tier to full 8 (+ policySettings, pluginHook, skillHook, sessionHook, builtinHook). Required for plugin ecosystem (T2-PLG-*).
Covers: gm#15 (full scope), gm#16, gm#17
Dependencies: T1-HOOK-05, T2-PLG-01
Risk: Medium; late-binding of session hooks requires careful lifecycle.

### T2-PLG-01 — moai Plugin System (minimal)
Rationale: Enable third-party distribution of skills/agents/commands. Scope REDUCED vs CC: skills + agents + commands only (no hooks/MCP/outputStyles in v3.0). Manifest at `.moai-plugin/plugin.json`.
Covers: gm#140, gm#141, gm#142 (scope-reduced), gm#143, gm#144, gm#145
Dependencies: T1-SCH-01
Risk: High effort (XL); scope creep risk. Explicit scope-lock: no hooks in plugin v1.

### T2-PLG-02 — Plugin marketplace (GitHub + local)
Rationale: Lightweight marketplace for discovering plugins. v3.0 scope: github + local directory sources only. Defer git/url/file.
Covers: gm#141 (scope-reduced)
Dependencies: T2-PLG-01
Risk: Medium; handoff to existing GitHub workflow infrastructure.

### T2-MEM-02 — LLM-based memory relevance selection (opt-in)
Rationale: Sonnet side-query selects top 5 relevant memories per turn. Requires opt-in via `.moai/config/sections/memory.yaml` due to cost (~$0.01-0.03/turn).
Covers: gm#48, gm#49, gm#52
Dependencies: T1-MEM-01
Risk: Medium; API cost is user-visible. Default off; telemetry-driven graduation decision for v3.2.

### T2-AGT-04 — Fork subagent primitive (simplified)
Rationale: Omit `subagent_type` in Agent() call → child inherits parent's system prompt (NOT full cache-identical prefix; that's v3.1). Enables cheap "verifier sub-agents".
Covers: gm#67 (scope-reduced)
Dependencies: T1-AGT-01
Risk: High; careful impl to avoid subagent recursion loops.

### T2-AGT-05 — Built-in moai agent definitions
Rationale: Ship Explore / Plan / general-purpose agent definitions alongside existing 22 moai agents. Users still get CC runtime's built-ins — these are moai's AUGMENTED versions with SPEC awareness.
Covers: gm#66
Dependencies: T1-AGT-01
Risk: Low; additive.

### T2-TEAM-01 — Teammate mailbox Zod-equivalent schemas
Rationale: 10 structured message types for team coordination (shutdown, plan_approval, permission_request/response, etc.). Replaces ad-hoc JSON.
Covers: gm#71, gm#72, gm#75, gm#76
Dependencies: T1-SCH-01
Risk: Breaking for existing team mode users. Grace period: accept both shapes, warn on legacy.

### T2-DIFF-01 — SPEC-to-SPEC chaining (moai-unique)
Rationale: Not from CC research; moai-unique leverage of existing SPEC foundation. Enables SPEC inheritance (base + extension), SPEC templates (reusable feature patterns), dependency graph validation (circular ref detection), lifecycle transitions (spec-first → spec-anchored → spec-as-source).
Covers: (not in CC gap matrix; extends moai SPEC workflow)
Dependencies: T1-SCH-01
Risk: Medium; must not over-complicate — 104 existing SPECs work today.

### T2-SKL-02 — `context: 'fork'` skill execution
Rationale: Skill runs as sub-agent with fresh context instead of inline expansion. Pairs with Fork Subagent (T2-AGT-04).
Covers: gm#87, gm#88
Dependencies: T1-SKL-01, T2-AGT-04
Risk: High effort; similar to fork subagent but at skill granularity.

### T2-SKL-03 — Dynamic skill discovery (walk-up)
Rationale: Walk UP from touched file paths to find nested `.claude/skills/` dirs; deepest-first precedence. Enables code-proximity skill activation.
Covers: gm#106, gm#107
Dependencies: T1-SKL-01
Risk: Medium; realpath canonicalization + symlink handling.

### T2-CLI-01 — Startup profiler (profileCheckpoint markers)
Rationale: Measure cold-start latency of init/update/hook commands. Required to iterate on `moai init` performance.
Covers: gm#128
Dependencies: none
Risk: Low; internal instrumentation only.

### T2-OUT-01 — Output contract v2 (diff + validation errors + progress)
Rationale: Improve CC rendering fidelity: emit `diff --git` format, structured `{severity, path, line, message, suggestion}` errors, `Progress: N/M` lines during long tasks.
Covers: gm#175, gm#176, gm#177, gm#178
Dependencies: none
Risk: Low; output-only changes.

---

## Tier 3 — Polish & Cleanup (v3.0.x / v3.1.0)

### T3-CLI-02 — Fast-path CLI shim for `--version`
Rationale: Lift version check before full router init. Matches CC's cli.tsx fast-path pattern.
Covers: gm#127
Risk: Low; pure perf win.

### T3-CLI-03 — preAction hook wiring
Rationale: cobra PersistentPreRunE to auto-run migrations on every command. Formalizes T1-MIG-01 invocation.
Covers: gm#133
Risk: Low.

### T3-CLI-04 — Shell completion via `moai completion <shell>`
Rationale: cobra built-in; small UX win.
Covers: gm#135
Risk: Zero.

### T3-CLI-05 — `--bare` minimal mode
Rationale: Skip hooks/LSP/plugins for CI/perf testing.
Covers: gm#136
Risk: Low.

### T3-MEM-03 — Memory path security validation
Rationale: Port CC's `validateMemoryPath` rules (reject non-absolute, tilde-only, UNC, null byte, NFKC attacks).
Covers: gm#53, gm#54
Risk: Low.

### T3-HOOK-08 — Hook dedup across sources
Rationale: `{shell}\0{command}\0{if}` dedup key prevents duplicate firings in multi-source configs.
Covers: gm#14
Dependencies: T1-HOOK-05
Risk: Low.

### T3-HOOK-09 — Hook matcher patterns upgrade (regex + per-event matchQuery)
Rationale: Support regex matchers and per-event match-query resolution (tool_name for PreToolUse, reason for SessionEnd, etc.).
Covers: gm#13
Risk: Medium.

### T3-TEAM-02 — In-process teammate via goroutine context (v3.1)
Rationale: Go-native alternative to tmux. Uses `context.Context` for identity isolation instead of AsyncLocalStorage.
Covers: gm#68
Dependencies: T2-TEAM-01
Risk: High; requires careful goroutine lifecycle management. Explicitly deferred to v3.1.

### T3-TEAM-03 — Plan-approval flow with structured messages
Rationale: Team lead approves/rejects teammate plans via structured message; rejection carries feedback.
Covers: gm#72
Dependencies: T2-TEAM-01
Risk: Low.

### T3-CMD-01 — `/diff`, `/memory`, `/permissions` slash commands
Rationale: CC-compat slash commands for common browsing. Lower priority since `moai` CLI covers most.
Covers: gm#111, gm#114, gm#115
Risk: Low effort each.

### T3-CMD-02 — `/doctor` slash command wrapper
Rationale: Thin wrapper over `moai doctor` CLI.
Covers: gm#122
Risk: Zero.

### T3-OUT-02 — StatusIcon colors + eighth-block progress bars
Rationale: ✓✗⚠ℹ○… with ANSI colors; `▏▎▍▌▋▊▉█` progress.
Covers: gm#179, gm#181
Risk: Zero.

### T3-OUT-03 — Code block language hints + OSC-8 file hyperlinks
Rationale: Always emit ` ```go ` fence; `file:///absolute/path:LINE` for clickability.
Covers: gm#178, gm#180
Risk: Zero.

### T3-OUT-04 — Output-style CC-compat schema
Rationale: Use CC keys (`name`, `description`, `keep-coding-instructions`, `force-for-plugin`) + `moai:` prefix for extensions. Prevent collision.
Covers: gm#182, gm#167, gm#169
Risk: Low.

### T3-DOC-01 — docs-site 4-locale sync (en/ja/zh catching up to ko)
Rationale: `docs-site/content/en/` lacks `contributing/` and `multi-llm/` sections. Requires expert-docs delegation.
Covers: gm#194
Risk: Low (translator workflow already exists at `docs-site/scripts/translate.mjs`).

### T3-AGT-06 — agentNameRegistry auto-resume of stopped agents
Rationale: SendMessage to evicted agent triggers transcript-based resume in background.
Covers: gm#82
Dependencies: T2-TEAM-01
Risk: Medium.

### T3-HOOK-10 — Prompt elicitation from hook stdout (bidirectional)
Rationale: Hook writes `{"prompt":"id","message":"...","options":[...]}` → CC calls `requestPrompt` → response piped back to hook stdin.
Covers: gm#23
Dependencies: T1-HOOK-01
Risk: Medium; stdin round-trip protocol.

---

## Tier 4 — Defer / Reject

### T4-BRIDGE-01 — Remote Control Bridge — REJECT
Rationale: CC's bridge subsystem (33 files, 500KB+) is tied to `api.anthropic.com` OAuth and cannot be reused for moai's CG Mode (Claude+GLM). A moai-equivalent would require parallel infrastructure investment with no clear value-add. moai's existing tmux-based CG Mode is sufficient.
Source: gm#86; W1.3 §1.7 ("Bridge is specifically for Anthropic's cloud")

### T4-BUDDY-01 — Buddy Sprite — REJECT
Rationale: Gamified companion sprite (duck/goose/blob). Zero business value for a professional dev tool. Not an agent system.
Source: W1.3 §3 ("buddy/ is NOT a pair-agent system")

### T4-SDK-01 — Agent SDK re-export barrel — REJECT
Rationale: moai is a CLI, not a library. No SDK consumers exist. Exposing internal Go APIs would lock in implementation details.
Source: gm#171; W1.5 §3.4; W1.5 §9.9

### T4-MCP-01 — MCP server entrypoint (`moai mcp serve`) — DEFER
Rationale: moai's tools are Go binary subcommands, not MCP-exposed. If future use case emerges (e.g., exposing `moai hook`-style event dispatch via MCP), revisit. Scope creep if pursued pre-demand.
Source: gm#170; W1.5 §9.3

### T4-REPL-01 — REPL TUI — REJECT
Rationale: moai is non-interactive by design. All interactive flows go through CC's AskUserQuestion.
Source: gm#173; W1.4 §8.1 ("moai-adk is a non-interactive Go CLI")

### T4-PRINT-01 — Headless mode `--print` — REJECT
Rationale: moai is always headless. Concept doesn't apply.
Source: gm#172

### T4-AUTH-01 — OAuth subcommands — REJECT
Rationale: moai relies on CC's auth. `auth login`/`logout` would duplicate.
Source: W1.5 §9.9

### T4-POLICY-01 — Policy limits / managed remote settings — DEFER
Rationale: Enterprise-gated feature. moai's current user base is individual developers. Revisit when enterprise contracts materialize.
Source: W1.5 §9.9

### T4-AUTOMODE-01 — auto-mode classifier command — REJECT
Rationale: moai has its own permissionMode config via `.moai/config/sections/quality.yaml`. auto-mode is CC-specific.
Source: W1.5 §2.1 (auto-mode defaults/config/critique)

### T4-REMOTE-01 — `isolation: 'remote'` (ant-only CCR teleport) — REJECT
Rationale: Anthropic-internal CCR infrastructure. Not available to external users.
Source: gm#84; W1.3 §2.3

### T4-PLG-FULL — Full CC-parity plugin system — DEFER
Rationale: CC's plugin system includes hooks, mcpServers, outputStyles, force-for-plugin auto-apply, hot-reload, cowork plugins, policy-gated plugins. Tier 2 T2-PLG-01 ships a REDUCED scope (skills+agents+commands only). Full parity deferred to v3.2+ pending ecosystem demand.
Source: gm#142 (full scope), gm#146, gm#147, gm#148

### T4-COMPACT-01 — 5-layer compaction pipeline — REJECT (CC-runtime)
Rationale: snip → microcompact → context-collapse → autocompact → reactive compact is CC-runtime's QueryEngine concern. moai delegates to CC; reimplementing would duplicate effort without benefit.
Source: gm#36, gm#37; W1.2 §9.7

### T4-COST-01 — Cost tracker with OpenTelemetry counters — DEFER
Rationale: moai has no per-session cost model today. CC already tracks this. Adding OTEL deps conflicts with moai's minimal-deps philosophy (9 direct deps). Revisit if moai-specific cost attribution becomes a requirement.
Source: gm#42, gm#43; W1.2 §9.4

### T4-COORD-01 — Coordinator Mode — DEFER
Rationale: CC's 380-line coordinator prompt forces single-worker delegation. moai already has manager-strategy + agent orchestration for this pattern.
Source: gm#85; W1.3 §2.1

---

## Delivery Sequencing

Dependencies form a DAG. Critical path for v3.0.0:

1. **Foundation week(s)**: T1-SCH-01 → T1-SCH-02 → T1-MIG-01 → T1-MIG-02
2. **Hook core**: T1-HOOK-01 → T1-HOOK-02/03/04/05 (parallel) → T1-HOOK-06/07
3. **Memory core**: T1-MEM-01 (independent; can run parallel with Hook core)
4. **Agent/Skill core**: T1-AGT-01/02/03 → T1-SKL-01 (after T1-SCH-01)
5. **Cleanup**: T1-CLN-01/02/03 (parallel; mostly gated on T1-MIG-02)
6. **Tier 2 differentiators**: T2-HOOK-10/11/12/13/14 (after T1-HOOK-01) → T2-PLG-01/02 → T2-MEM-02 → T2-AGT-04/05 → T2-TEAM-01 → T2-DIFF-01 → T2-SKL-02/03 → T2-CLI-01 → T2-OUT-01

Parallel tracks where no dependencies exist (e.g., T2-CLI-01 startup profiler, T2-OUT-01 output contract) should be executed concurrently via worktree isolation.

## Open Questions for Wave 3 Architect

Flagged in synthesis for explicit design resolution:

1. Hook JSON output protocol backward compat window: v3.0 dual-parse, v3.2 warn, v4.0 remove. Confirm.
2. Plugin system v1 scope: skills + agents + commands only (no hooks/MCP/outputStyles). Confirm or expand.
3. Memory LLM relevance: opt-in default-off for v3.0. Telemetry-driven graduation to default-on in v3.2. Confirm.
4. Settings source layering v1: 3-tier (user/project/local) in v3.0; full 6-tier in v3.2. Confirm.
5. Fork subagent scope: simplified (inherit system prompt) in v3.0, full cache-identical prefix in v3.1. Confirm.
6. Schema technology: validator/v10 struct tags + JSON Schema export (Go-native) vs CUE (declarative). Opinion needed.
7. In-process teammate: v3.1 (not v3.0). Confirm deferral.
8. CG Mode vs CC Bridge: reject Bridge transport, but adopt `SDKControlRequest/Response` message NAMING for internal CG channel schemas. Opinion needed.

End of roadmap. See `v3-themes.md` for architectural organization of Tier 1 + Tier 2 items into top-level design themes.
