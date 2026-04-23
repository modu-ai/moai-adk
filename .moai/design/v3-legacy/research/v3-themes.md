# MoAI-ADK v3 — Architectural Themes

Generated: 2026-04-23
Source: `gap-matrix.md` + `priority-roadmap.md`

## Purpose

This document organizes Tier 1 and Tier 2 roadmap items into 9 top-level themes that will structure the master design document. Each theme bundles related capabilities into a coherent architectural story the Wave 3 architect can use as chapter headings.

Themes are NOT parallel workstreams — they are SYNTHESIS units. A single roadmap item (e.g., T1-HOOK-01 Hook Protocol v2) can anchor multiple themes. Themes 1-6 are primarily CC-parity; Theme 7 is moai-unique; Theme 8-9 are contract/hygiene.

---

## Theme 1 — Hook Protocol v2

Scope (5 bullets):
- Rich JSON output protocol: `decision`, `hookSpecificOutput`, `additionalContext`, `updatedInput`, `updatedMCPToolOutput`, `systemMessage`, `stopReason`, `continue`, `suppressOutput`, `watchPaths`
- 4 hook types: `command` (existing), `prompt` (new, Haiku-gated), `agent` (new, subagent verifier), `http` (new, SSRF-guarded)
- Per-hook features: `if` condition (permission-rule syntax), `async` / `asyncRewake`, `once`, shell selection, matcher patterns (exact / pipe / regex / `*`)
- Source precedence pipeline: user → project → local → policy → plugin → skill → session → builtin (3-tier v3.0, full 8-tier v3.2)
- CLAUDE_ENV_FILE mechanism: SessionStart/Setup/CwdChanged/FileChanged write bash exports for subsequent BashTool invocations

Tier 1 items: T1-HOOK-01 through T1-HOOK-07 (7 items)
Tier 2 items: T2-HOOK-10 through T2-HOOK-14 (5 items)

Rationale: Hooks are moai's single largest integration surface with Claude Code. The current protocol is capability-poor (exit codes only). Upgrading this one subsystem unlocks: plugin ecosystem (hooks discovered from plugins), permission-request participation (programmatic allow/deny before dialog), model-turn context injection (additionalContext), tool-input rewriting (updatedInput), environment injection (CLAUDE_ENV_FILE), LLM-gated quality checks (type: prompt), and async long-running checks (asyncRewake).

Risks:
- Breaking change: existing hook wrappers use exit codes + simple JSON. Dual-parse during v3.0→v3.2 grace period.
- Security: `type: http` requires faithful port of CC's SSRF guard (`utils/hooks/ssrfGuard.ts`).
- Performance: source precedence merge on every event dispatch must be sub-millisecond — use `sync.Once` + cached merge.
- `type: agent` recursion: agent-hooks triggering more agent-hooks; bound with `ALL_AGENT_DISALLOWED_TOOLS` set.

Open question: what's the deprecation timeline for exit-code-only hooks? Proposed v3.0 dual-parse → v3.2 warn → v4.0 remove.

---

## Theme 2 — Memory 2.0

Scope (5 bullets):
- MEMORY.md truncation: 200-line + 25K-byte cap with human-readable warning appendix (aligns with CC memdir `truncateEntrypointContent`)
- 4-type memory taxonomy ENFORCEMENT (user/feedback/project/reference) — moai already uses this schema per `~/.claude/projects/-*/memory/MEMORY.md` format; just add validation
- Memory freshness `<system-reminder>` wrapping for >1 day old memories (prevents stale-fact assertions)
- LLM-based relevance selection (opt-in) — Sonnet side-query returns top 5 relevant memories per turn
- Path security validation: reject non-absolute, tilde-only, UNC, null byte, NFKC Unicode attacks (from CC `validateMemoryPath`)

Tier 1 items: T1-MEM-01
Tier 2 items: T2-MEM-02
Tier 3 items: T3-MEM-03

Rationale: moai already aligns with CC's memory TAXONOMY (4 types) and PATH CONVENTION (canonical git root via sanitized project hash). The delta is enforcement and intelligence: truncation prevents MEMORY.md bloat; freshness warnings prevent stale fact propagation (a CC BigQuery-driven learning from real user reports per W1.2 §6.12); LLM relevance selection turns memory from a linear scan into an intelligent retrieval system.

Risks:
- LLM cost: side-query adds ~$0.01-0.03 per turn. Opt-in via `.moai/config/sections/memory.yaml` default off.
- Truncation UX: first-time truncation of an existing 30K+ MEMORY.md may surprise users. Log a clear migration event.
- Schema collision: moai's ad-hoc memory files may fail NFC normalization on Windows. Test matrix needs coverage.

Open question: should moai ship KAIROS-style daily log distillation (`~/memory/logs/YYYY/MM/YYYY-MM-DD.md` with nightly `/dream`) or defer to v3.1? Recommendation: defer.

---

## Theme 3 — Plugin Ecosystem (Minimal)

Scope (5 bullets):
- Plugin manifest at `.moai-plugin/plugin.json` with `name`, `description`, `version`, `capabilities` fields
- 3 plugin kinds: built-in (shipped with moai binary), marketplace (installed from github/local), session-inline (`--plugin-dir`)
- Plugin capabilities v1: agents + skills + commands (NO hooks/MCP/outputStyles in v3.0 — reduced scope)
- Install scopes: user (`~/.moai/plugins/`), project (`.moai/plugins/`), local (`.moai/plugins/local/`)
- CLI surface: `moai plugin install/uninstall/enable/disable/update/list`, `moai plugin marketplace add/list/remove/update`, `moai plugin validate <path>`

Tier 1 items: (none — full plugin system is Tier 2)
Tier 2 items: T2-PLG-01, T2-PLG-02
Tier 4 items: T4-PLG-FULL (full CC-parity plugin system, defer)

Rationale: moai's 22 agents + 50 skills + 13 commands are a tight, opinionated catalog. A plugin system unlocks community contribution without eroding that opinionatedness. Scope reduction (no hooks in plugins) is deliberate — hooks are global session-level contracts; plugin hooks would multiply source-precedence complexity. MCP servers are out of scope because moai tools are Go binaries, not MCP-exposed.

Risks:
- Scope creep: CC's plugin system is MASSIVE (agents/skills/commands/hooks/mcpServers/outputStyles/force-for-plugin/hot-reload/cowork variant/policy-gated). Keep v3.0 scope locked to 3 capabilities.
- Security: plugin code execution requires the same trust model as `.claude/commands/` and `.claude/skills/`. Reuse moai's existing workspace-trust pattern.
- Marketplace source types: v3.0 supports github + local directory only. Defer git/url/file to v3.1.
- Dependency conflicts: two plugins declaring the same agent `name` → error at install time.

Open question: should plugins be versioned in their manifest (semver) or pinned by commit hash? Recommendation: semver for marketplace, commit hash for github-direct.

---

## Theme 4 — Migration Framework

Scope (5 bullets):
- `CURRENT_MIGRATION_VERSION` counter in `.moai/config/sections/system.yaml`
- Ordered, idempotent migration runner firing in cobra PersistentPreRunE
- Per-migration patterns: flag-based completion (for one-shot transforms) + value-match idempotency (for re-runnable)
- Initial migration set (5): template_version sync, .agency/ removal, hook wrapper drift, skill drift, .moai-backups/ archival
- Opt-out via `MOAI_DISABLE_MIGRATIONS=1` (for CI/airgapped)

Tier 1 items: T1-MIG-01, T1-MIG-02

Rationale: moai has 104 SPECs, 50 skills, 22 agents, 502 embedded template files, 35 rule files. This config surface is too large to manage with ad-hoc one-shot migrations forever. Establishing a versioned migration framework NOW (while only `moai migrate agency` exists) is the correct time — adding it later with dozens of needed migrations is harder. The 5 initial migrations are DOGFOOD: they fix moai's own self-identified drift (W1.6 §15).

Risks:
- Breaking change: migrations firing on every command is a SEMANTIC shift. Users who expect "moai init is idempotent" may be surprised when an unrelated command runs a migration. Clear communication needed.
- Migration misbehavior: if a migration corrupts config, user loses confidence forever. Each migration MUST support dry-run, log before/after state, and have rollback capability via `.moai/backups/`.
- Ordering invariant: migrations run in strict order; skipping requires version-specific gating.
- Opt-out risk: `MOAI_DISABLE_MIGRATIONS=1` lets CI pipelines freeze at an old config version. Warn on stale config age.

Open question: should migrations run on every cobra command (aggressive) or only on `moai init`, `moai update`, `moai doctor`, `moai migrate` (conservative)? Recommendation: conservative — reduces blast radius.

---

## Theme 5 — Agent Runtime v2

Scope (5 bullets):
- Agent frontmatter bundle: `memory`, `initialPrompt`, `requiredMcpServers`, `omitClaudeMd`, `maxTurns`, `criticalSystemReminder_EXPERIMENTAL`, `skills` preload (formalize), `background` (as frontmatter)
- Fork subagent primitive (simplified): omit `subagent_type` → child inherits parent's system prompt (NOT cache-identical prefix; v3.1 goal)
- Built-in moai agents: Explore / Plan / general-purpose as moai-augmented versions (SPEC-aware)
- Teammate mailbox Zod-equivalent schemas: shutdown_request / shutdown_approved / shutdown_rejected / plan_approval_request / plan_approval_response / permission_request / permission_response / sandbox_permission_request / sandbox_permission_response / task_assignment (10 types)
- Plan-approval flow: team lead approves/rejects teammate plans via structured message with `feedback` payload

Tier 1 items: T1-AGT-01, T1-AGT-02, T1-AGT-03
Tier 2 items: T2-AGT-04, T2-AGT-05, T2-TEAM-01, T2-SKL-02 (context:fork)
Tier 3 items: T3-TEAM-02 (in-process teammate, v3.1), T3-AGT-06 (agentNameRegistry)

Rationale: Agent frontmatter bundle is a quick win — each field is small but together they give moai agents parity with CC's agent definition model. `omitClaudeMd` alone is a massive token win for Explore-style read-only agents (per W1.3 §4.2, CC saves 5-15 Gtok/week on 34M+ Explore spawns). Mailbox Zod schemas replace ad-hoc JSON with validated discriminated unions, preventing a whole class of team-mode bugs. Fork subagent enables cheap "verifier" sub-agents as a first-class pattern instead of requiring full Agent() spawns.

Risks:
- Breaking: existing team mode uses ad-hoc JSON. Grace period accepting both shapes.
- Fork subagent recursion: must detect and prevent (CC uses `FORK_BOILERPLATE_TAG`).
- In-process teammate (T3): goroutine + `context.Context` for identity isolation is different from AsyncLocalStorage. Test thoroughly.
- `requiredMcpServers` availability check: 30-second timeout for pending MCP connections matches CC. Configurable.

Open question: should `omitClaudeMd` be default-TRUE for moai's Explore-style research agents? Recommendation: default-FALSE, opt-in via per-agent frontmatter. Too-clever-by-default is risky.

---

## Theme 6 — Schema & Validation Layer

Scope (5 bullets):
- Formal schemas for all `.moai/config/sections/*.yaml` files via `go-playground/validator/v10` struct tags
- JSON Schema export (auto-generated from Go structs) for editor integration (`$schema` field)
- Settings source layering: 3-tier v3.0 (user / project / local), full 6-tier v3.2 (+ policy / flag / managed)
- Validation at settings load (`parseSettingsFile`) and write (`updateSettingsForSource`) with round-trip safety
- Schema for hook types (discriminated union), MCP server config, permission rules, migration metadata

Tier 1 items: T1-SCH-01, T1-SCH-02

Rationale: moai currently has ZERO formal schemas. YAML is parsed via go-yaml with struct tags but no validation. This is fragile — a typo in `.moai/config/sections/quality.yaml` silently fails at runtime. Formalizing schemas now is foundational for every other Tier 1 item (migrations need schemas to validate before/after; plugins need schemas to declare capabilities; hook source precedence needs schemas to merge). The technology choice matters: Go struct tags + validator/v10 is idiomatic Go; JSON Schema export enables VS Code / IntelliJ integration without a separate schema language.

Risks:
- Breaking: existing configs with typos may now fail validation. Ship with `moai doctor config --fix` auto-repair mode.
- Over-engineering: resist the temptation to validate EVERYTHING. Start with the 22 YAML files in `.moai/config/sections/` + hook schema + migration schema.
- Source precedence merge semantics: deep-merge vs replace-on-conflict per field. Document precedence rules in CLAUDE.md.
- JSON Schema drift: auto-gen from Go structs must be part of `make build`, not manual.

Open question: validator/v10 (Go-native) vs CUE (declarative, external)? Recommendation: validator/v10 for v3.0 — minimal new deps, fits moai's 9-dep philosophy. Revisit CUE if cross-language schema sharing becomes a requirement.

---

## Theme 7 — SPEC-to-SPEC Chaining (moai-unique)

Scope (5 bullets):
- SPEC inheritance: base SPEC + extension SPECs via YAML `inherits:` field in frontmatter
- SPEC templates: reusable patterns for common feature archetypes (CRUD, auth flow, API endpoint, etc.) in `.moai/spec-templates/`
- SPEC dependency graph validation: detect circular references before `/moai plan` commits
- SPEC lifecycle transitions: `spec-first` → `spec-anchored` → `spec-as-source` with explicit justification required for downgrades
- SPEC-to-REQ-to-TAG chain tracking: REQ-IDs in acceptance.md map to @MX:ANCHOR / @MX:NOTE tags in implementation

Tier 2 items: T2-DIFF-01 (the whole theme)

Rationale: moai's 104 active SPECs demonstrate the SPEC-First methodology works. The gap is RE-USE: currently each SPEC is authored from scratch even when 80% overlaps with a prior SPEC (common in CRUD / auth / API patterns). SPEC inheritance + templates extract this reuse. Lifecycle transitions (from SDD 2025 Standard already documented in `moai-workflow-spec` skill) formalize when a SPEC becomes a source of truth. This is NOT from CC research — it's a moai-unique differentiator leveraging the existing SPEC foundation.

Risks:
- Over-complication: 104 SPECs work today without chaining. Don't add complexity users don't want.
- Template quality: bad templates poison downstream SPECs. Curate a small initial set (3-5 templates) and iterate.
- Inheritance cycles: graph validator needs to run at SPEC save time, not just at plan time.
- Lifecycle downgrade: moving from spec-as-source back to spec-first is an anti-pattern that should require explicit user confirmation.

Open question: are SPEC templates user-editable markdown files or code-generated? Recommendation: markdown with variable placeholders (Go text/template style), since users already edit SPEC markdown.

---

## Theme 8 — Output Contract v2

Scope (5 bullets):
- Diff visualization: emit `diff --git` format so CC's StructuredDiff renders with syntax highlighting (replaces plain-text diffs)
- Structured validation errors: `{severity, path, line, message, suggestion}` YAML/markdown so CC's ValidationErrorsList can render
- Progress indicators during long tasks: `Status:` / `Task:` / `Progress: N/M` line prefixes that CC's HookProgressMessage picks up
- Code block language hints: always emit ` ```go `, ` ```python ` fences for CC's HighlightedCode syntax highlighting
- File path OSC-8 hyperlinks: `file:///absolute/path:LINE` for clickability via CC's FilePathLink

Tier 2 items: T2-OUT-01
Tier 3 items: T3-OUT-02, T3-OUT-03, T3-OUT-04

Rationale: moai's output currently passes through CC's Ink renderer as plain text. CC has 146 components for rendering structured content (Markdown, HighlightedCode, StructuredDiff, ValidationErrorsList, StatusIcon, ProgressBar). moai can leverage 100% of these WITHOUT implementing any TUI code — just by emitting the right output formats. This is a pure output-fidelity theme: zero net-new capabilities, but massive UX improvement when moai runs inside CC sessions.

Risks:
- Output-style collision: moai's `.claude/output-styles/` frontmatter may collide with CC's `name`, `description`, `keep-coding-instructions`, `force-for-plugin` keys. Use `moai:` prefix for extensions.
- ANSI in non-TTY: when moai runs in CI (no terminal), OSC-8 and eighth-block chars must gracefully degrade. Detect `isatty` + emit plain fallback.
- Version dependency: CC changes its rendering heuristics between versions. Keep output contract stable across CC v2.1.97+.

Open question: should moai add `moai:` prefixed frontmatter fields to output styles (for metadata like moai SPEC-ID) or stick to CC-compat-only? Recommendation: reserve `moai:` prefix, ship with empty-value fields for future use.

---

## Theme 9 — Internal Cleanup

Scope (5 bullets):
- Config drift fix: `project.yaml.template_version` auto-sync with `system.yaml.moai.version` (currently v2.7.22 vs v2.12.0 — 12 minor versions stale)
- Template drift fix: deploy 3 template-only skills locally (`moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`) and `handle-permission-denied.sh` wrapper
- Legacy removal: `.agency/` stub redirects, `.moai-backups/` folder archival, `.go.bak` files (glm.go.bak 28K, worktree/new_test.go.bak 13K), stale `coverage.out`/`coverage.html` from 2026-03-11
- Doc drift: fix ADR-011 comment in `internal/template/embed.go:8-12` (claims runtime-generated files excluded, but `.claude/settings.json.tmpl` IS embedded)
- docs-site 4-locale sync: `docs-site/content/en/` lacks `contributing/` and `multi-llm/` sections present in `ko/`

Tier 1 items: T1-CLN-01, T1-CLN-02, T1-CLN-03
Tier 3 items: T3-DOC-01 (docs-site sync)

Rationale: These are moai's SELF-IDENTIFIED issues from Wave 1.6 §15. They don't add capability — they restore integrity. Shipping them as part of v3.0 signals discipline and avoids accumulating tech debt into the next major version. All items are low-risk XS-effort; most are single-file changes or one-shot migrations (T1-MIG-02 M01-M05).

Risks:
- User-visible config changes: auto-syncing `template_version` may surprise users who manually pinned it. Add a one-time notice.
- Locale translation cost: docs-site sync requires expert-docs translation. Leverage existing `docs-site/scripts/translate.mjs`.
- Cleanup scope creep: resist adding "while we're at it" changes. Keep strictly to Wave 1.6 §15 identified items.

Open question: should `.moai-backups/` be deleted or moved to `~/.moai/history/{timestamp}/`? Recommendation: move (never delete user data without explicit consent).

---

## Theme Summary Matrix

| Theme | Tier 1 Items | Tier 2 Items | Tier 3 Items | Tier 4 (excluded) |
|-------|--------------|--------------|--------------|-------------------|
| 1. Hook Protocol v2 | 7 | 5 | 3 | bridge-transport |
| 2. Memory 2.0 | 1 | 1 | 1 | 5-layer compaction, KAIROS daily log |
| 3. Plugin Ecosystem | 0 | 2 | 0 | full CC-parity plugin system |
| 4. Migration Framework | 2 | 0 | 1 | (none) |
| 5. Agent Runtime v2 | 3 | 4 | 2 | coordinator mode, isolation:remote |
| 6. Schema & Validation | 2 | 0 | 0 | sandbox schema, SDK schemas |
| 7. SPEC-to-SPEC | 0 | 1 | 0 | (none) |
| 8. Output Contract v2 | 0 | 1 | 3 | (none) |
| 9. Internal Cleanup | 3 | 0 | 1 | (none) |

Total: 18 Tier-1 items, 14 Tier-2 items, 11 Tier-3 items, 13 Tier-4 exclusions.

## Cross-Theme Dependencies

Critical dependencies that cross theme boundaries:

- Theme 1 (Hook) depends on Theme 6 (Schema) for source precedence
- Theme 3 (Plugin) depends on Theme 6 (Schema) for manifest validation
- Theme 3 (Plugin) depends on Theme 1 (Hook) for plugin-registered hooks (deferred to T4-PLG-FULL)
- Theme 4 (Migration) depends on Theme 6 (Schema) for before/after state validation
- Theme 5 (Agent) depends on Theme 6 (Schema) for frontmatter validation
- Theme 9 (Cleanup) depends on Theme 4 (Migration) for implementation (M01-M05)
- Theme 2 (Memory) is independent of all other themes

Implication: Theme 6 (Schema & Validation) is the critical-path foundation. Ship it FIRST within v3.0 development.

## For the Wave 3 Architect

These 9 themes should serve as the top-level chapters of the v3 master design document. Suggested chapter structure:

- Chapter 1: Introduction (v3 vision, scope, non-goals)
- Chapter 2: Schema & Validation Layer (Theme 6) — the foundation
- Chapter 3: Migration Framework (Theme 4)
- Chapter 4: Hook Protocol v2 (Theme 1)
- Chapter 5: Agent Runtime v2 (Theme 5)
- Chapter 6: Memory 2.0 (Theme 2)
- Chapter 7: Plugin Ecosystem (Theme 3)
- Chapter 8: SPEC-to-SPEC Chaining (Theme 7)
- Chapter 9: Output Contract v2 (Theme 8)
- Chapter 10: Internal Cleanup (Theme 9)
- Chapter 11: Delivery Plan (sequencing from priority-roadmap.md)
- Chapter 12: Migration Guide for v2→v3 users

End of themes. Cross-references: `gap-matrix.md` (194 rows), `priority-roadmap.md` (56 tiered items).
