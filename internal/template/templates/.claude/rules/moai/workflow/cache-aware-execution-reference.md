---
description: "Reference companion for cache-aware-execution.md — cited cache-cost numbers, per-directive rationale, and worked examples for directives 6-10"
paths: "**/cache-aware-execution.md"
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
| Cache TTL (subscription auth, default) | 1 hour |
| Cache TTL (API-key auth, default) | 5 minutes |

Both TTL rows are authentication-mode defaults, not fixed properties: each cache hit refreshes the window (sliding TTL), and at the API level the 5-minute TTL is the default while the 1-hour TTL is an explicit `cache_control` override (`ttl: "1h"`). A Claude Code session selects neither — the runtime places cache breakpoints itself (parent rule, Non-goals).

The parent rule's intro quotes different figures for the write multiplier and TTL (reads ~0.1x, writes 1.25x, 5-minute TTL) — both sets are quoted from their respective source passages; neither was measured here, and the difference is a scope-of-quote artifact, not a contradiction this file resolves. Neither figure is universal: the write multiplier ranges up to 2x and the TTL follows the authentication default (see above), so the parent's 1.25x / 5-minute pairing holds only for the API-key-default passage it was quoted from.

## Directive rationale

### Directive 6 — `@`-mention, and `/context` as a one-shot audit

Citing a filename in prose and relying on the model to fetch it spends a turn on the fetch and risks a miss or a stale read; an `@`-mention loads the file once, deterministically, at the point of need. `/context` re-renders the loaded-context summary on every invocation, which is an audit action (what is actually loaded, how close is the ceiling), not a routine status check.

### Directive 7 — bounded command output

Command output lands in the context and stays there for every later turn of the session; one verbose command is paid again on each subsequent request. The runtime truncates at `BASH_MAX_OUTPUT_LENGTH`, but the bound should be chosen by the caller — quiet flags, targeted queries, or redirect-to-file with the exit code and a bounded tail (the file-redirect contract of `agent-common-protocol.md` § Parallel Execution, generalized from verification batches to all commands).

### Directive 8 — quiet forms of routine commands

Routine commands in their default form emit progress spinners, banners, tables, and ANSI color — none of it decision-relevant. The quiet form returns the same decision bytes for a fraction of the context cost: `--no-progress`, `-q`/`--quiet`, machine-readable output (`--json`, `--porcelain`) piped through a targeted filter.

### Directive 9 — session length as a cost axis

Every fresh session re-pays the always-loaded prefix at write price; a continuing session reads it from cache — but only while that cache is warm. A gap longer than the idle TTL (a blocking user gate, an unattended wait) or a prefix invalidation (a session-loaded file edited mid-session — directive 3) reverts the next request to write price, collapsing the cheaper-continuation claim. For the same body of work, N short sessions multiply the write; one long session pays it once. This is the same axis directive 4 weighs for `/clear` — the paste-ready resume exists precisely so a context reset buys something.

### Directive 10 — mid-session model or effort switch

Caches are model-scoped, and an effort or thinking-budget change (`MAX_THINKING_TOKENS`) keys a fresh cache as well: the request after the switch re-writes the prefix at full price. The apparent conflict with `agent-common-protocol.md` § Per-Spawn Model Injection is an axis difference, stated in the directive itself: Per-Spawn Model Injection governs which model a *subagent context* runs on (most agent definitions declare `model: inherit`, and an unspecified spawn silently falls back to the parent model); directives 5 and 10 govern the *main session's* accumulated cache. Neither SSOT revises the other.

---

Version: 1.0.0
Classification: Reference companion — paths-scoped, never part of the always-loaded surface.
