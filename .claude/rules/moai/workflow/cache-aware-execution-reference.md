---
description: "Reference companion for cache-aware-execution.md — cited cache-cost numbers, per-directive rationale, and worked examples for directives 6-10"
paths: "**/cache-aware-execution.md,**/.claude/agents/moai/*.md,**/.claude/rules/moai/workflow/*.md"
---

# Cache-Aware Execution — Reference Companion

> This is the reference companion of `cache-aware-execution.md`. The always-loaded rule owns the directive bodies (numbered 1-10) and nothing else from this file. This file owns the cited cache numbers, the per-directive rationale, and the worked examples behind directives 6-10. Load this file when authoring or reviewing one of those directives, or when a directive's grounding is challenged.

## Cited cache numbers

All figures in this section are **quoted source values** — quoted from the source article, **not re-measured** in the session that authored directives 6-10, and never to be restated as measurements of this project. Date context: quoted from the source article as current at this file's authoring and cross-checked against Anthropic's prompt-caching documentation at the same time — performance figures the provider may change; re-quote before relying on one.

| Quantity | Quoted value |
|---|---|
| Cache read | 0.1x the base input price |
| Cache write (worst case) | up to 2x |
| Output tokens | roughly 5x the input price — output pricing is separate from input pricing and model-dependent; not a cache-cost figure |
| Cache TTL (main conversation, Claude subscription within plan included usage) | 1 hour |
| Cache TTL (main conversation on usage credits, API key, or cloud provider) | 5 minutes |
| Cache TTL (subagents, workflows, in-process teammates, forks, compaction, titles) | 5 minutes |

The TTL is decided per request bucket and billing, not by authentication mode (official prompt-caching docs, § Which TTL each request gets; quoted, not measured in this project): the main conversation gets one hour only on a Claude subscription within plan included usage, every other bucket gets five minutes, and the one exception is server-controlled helper requests on subscriptions, which get one hour. These rows are defaults, not fixed properties: each cache hit refreshes the window (sliding TTL), and at the API level the 5-minute TTL is the default while the 1-hour TTL is an explicit `cache_control` override (`ttl: "1h"`). Before Claude Code 2.1.243 a session selected neither — the runtime placed cache breakpoints itself (parent rule, Non-goals). Since 2.1.243 (official docs floor: 2.1.242) the `promptCacheTtl` and `subagentPromptCacheTtl` settings let API-key and cloud-provider users opt the main conversation into the 1-hour TTL while subagents stay at 5 minutes, with per-bucket env-var twins `CLAUDE_CODE_PROMPT_CACHE_TTL` / `CLAUDE_CODE_SUBAGENT_PROMPT_CACHE_TTL` (Claude Code 2.1.243 changelog; official prompt-caching docs); the runtime still places the breakpoints. A gateway must forward the `anthropic-beta` header for the one-hour TTL to apply (same docs page). Whether a third-party gateway (`ANTHROPIC_BASE_URL`, e.g. z.ai) honors the setting is unmeasured — keep the 5-minute default as the operating assumption there until measured.

The parent rule's intro quotes different figures for the write multiplier and TTL (reads ~0.1x, writes 1.25x, 5-minute TTL) — both sets are quoted from their respective source passages; neither was measured here, and the difference is a scope-of-quote artifact, not a contradiction this file resolves. Neither figure is universal: the write multiplier ranges up to 2x and the TTL follows the request bucket and billing (see above), so the parent's 1.25x / 5-minute pairing holds only for the 5-minute-TTL passage it was quoted from.

## Directive rationale

### Directive 6 — `@`-mention, and `/context` as a one-shot audit

Citing a filename in prose and relying on the model to fetch it spends a turn on the fetch and risks a miss or a stale read; an `@`-mention loads the file once, deterministically, at the point of need. `/context` re-renders the loaded-context summary on every invocation, which is an audit action (what is actually loaded, how close is the ceiling), not a routine status check.

### Directive 7 — bounded command output

Command output lands in the context and stays there for every later turn of the session; one verbose command is paid again on each subsequent request. The runtime truncates at `BASH_MAX_OUTPUT_LENGTH`, but the bound should be chosen by the caller — quiet flags, targeted queries, or redirect-to-file with the exit code and a bounded tail (the file-redirect contract of `agent-common-protocol.md` § Parallel Execution, generalized from verification batches to all commands).

### Directive 8 — quiet forms of routine commands

Routine commands in their default form emit progress spinners, banners, tables, and ANSI color — none of it decision-relevant. The quiet form returns the same decision bytes for a fraction of the context cost: `--no-progress`, `-q`/`--quiet`, machine-readable output (`--json`, `--porcelain`) piped through a targeted filter.

### Directive 9 — session length as a cost axis

Every fresh session re-pays the always-loaded prefix at write price; a continuing session reads it from cache — but only while that cache is warm. A gap longer than the idle TTL (a blocking user gate, an unattended wait) or a prefix invalidation (a session-loaded file edited mid-session — directive 3) reverts the next request to write price, collapsing the cheaper-continuation claim. For the same body of work, N short sessions multiply the write; one long session pays it once. This is the same axis directive 4 weighs for `/clear` — the paste-ready resume exists precisely so a context reset buys something.

#### Idle compaction on the one-hour TTL

Quoted source values from the official docs as current when this section was authored; not measured in this project. Re-quote before relying on one.

- **What it does.** Claude Code 2.1.299 added terminal idle compaction: a conversation of 200K+ tokens idle for about 54 minutes is compacted before its one-hour prompt cache expires, so the next message sends a summary instead of reprocessing the whole conversation.
- **Default-on conditions (all must hold).** An interactive terminal, VS Code chat panel, or desktop Code tab, running locally or over SSH; `autoCompactEnabled` on; the conversation on the one-hour prompt-cache TTL; Claude Code able to fetch feature flags from Anthropic; context at or above the token threshold; about an hour since the last request with the computer awake.
- **Controls.** `idleCompaction` setting (boolean, default `true`, any settings file, v2.1.290+); the `/config` row "Idle compaction" (v2.1.302+); `CLAUDE_CODE_IDLE_COMPACT_MIN_TOKENS` sets the threshold (default 200000, positive integer, values below 100000 raised to 100000, v2.1.290+). `DISABLE_AUTO_COMPACT` turns idle compaction off together with auto-compaction.
- **History.** 2.1.299 introduced it; 2.1.302 added the `/config` row; 2.1.314 extended it to VS Code and the Desktop Code tab, and made it run for subscribers with telemetry turned off and on Bedrock, Vertex, Foundry, and LLM gateways once the one-hour prompt cache is on.
- **MoAI policy.** MoAI keeps the default and sets no key; it ships no `idleCompaction` setting, environment variable, or code path. Turning it off for a session is the operator's choice, via `/config` or `"idleCompaction": false` in an uncommitted settings file.
- **Sources.** https://code.claude.com/docs/en/settings-reference#idlecompaction ; https://code.claude.com/docs/en/prompt-caching#which-ttl-each-request-gets ; https://code.claude.com/docs/en/env-vars (`CLAUDE_CODE_IDLE_COMPACT_MIN_TOKENS` row); https://code.claude.com/docs/en/changelog (2.1.299, 2.1.302, 2.1.314).
- **Not measured.** Whether GLM or other gateway sessions are on the one-hour TTL or run idle compaction; the `DISABLE_AUTO_COMPACT` env-vars row itself (only the settings-reference statement was read).

### Directive 10 — mid-session model or effort switch

Caches are model-scoped, and an effort or thinking-budget change (`MAX_THINKING_TOKENS`) keys a fresh cache as well: the request after the switch re-writes the prefix at full price. Subagents inherit the session's model and effort, so a switch reaches them as well.

---

Version: 1.0.0
Classification: Reference companion — paths-scoped, never part of the always-loaded surface.

## Migrated from the core body


### Cache-Aware Execution


Prompt-caching-aware ordering rules for orchestrator execution. Caching is a **prefix match** over the rendered request: reads are far cheaper than writes, the TTL is **idle-based**, and a blocking wait longer than the window makes the next turn re-write the whole accumulated prefix. The directives below are designed against the 5-minute TTL, while a conversation on the one-hour TTL additionally gets idle compaction (§ Directive 9) — MoAI sets no TTL override, because cache spend is a user decision. Cited numbers, the TTL-setting caveat, and per-directive rationale: `cache-aware-execution-reference.md`. These rules govern WHEN and IN WHAT ORDER the orchestrator acts; they change no gate semantics and never bypass any approval gate.


> **Loading scope**: Intentionally always-loaded — the directives bind ordering decisions the orchestrator makes on any non-trivial turn (gate placement, agent spawns, rule edits, `/clear` timing).


### Cross-references


- `.claude/rules/moai/workflow/orchestration-mode-selection.md` — fanout parallel fan-out (stagger-spawn composes with its concurrency ceiling)
- `.claude/rules/moai/workflow/context-window-management.md` — model-specific `/clear` thresholds (directive 4 is an additional, earlier trigger)
- `cache-aware-execution-reference.md` — the lazy companion. Load it for § Cited cache numbers · § Directive rationale (directives 6-10).


Version: 1.2.0 (directive 5 — the opt-in `llm.agent_overrides_consume` key named as the single sanctioned override surface)
Classification: Evolvable operational rule — execution ordering only; gate semantics unchanged.

