---
description: "Detail companion for context-window-management.md — the SSE-stall rationale, the graduated-compaction layer vocabulary consumed from the runtime, and the four-part context-usage detection heuristics (state file, validity guard, fallbacks, two-stage marker)"
paths: "**/context-window-management*.md,**/session-handoff*.md,**/internal/statusline/**"
---

# Context Window Management — Detail Companion

> Detail companion of `context-window-management.md` (the always-loaded stub). The stub owns the
> per-model threshold table, the reduction ladder, and the user and orchestrator obligations. This
> file owns why the ceiling matters, the runtime's compaction vocabulary, and how context usage is
> actually detected. Load it when reading the usage state file, diagnosing a stalled stream, or
> changing what the statusline writes.

## Why This Matters

Anthropic SSE streams stall (`stream_idle_partial`) near the context window ceiling — intermittent but predictable above the model-specific threshold. Reference: large-SPEC SSE-stall mitigation.

> **CC 2.1.196 watchdog note**: The streaming idle watchdog is now default-on for all providers — it aborts and retries a response stream that produces no events for 5 minutes (`CLAUDE_ENABLE_STREAM_WATCHDOG=0` disables). This softens the mid-stream-hang *consequence* (auto-abort+retry) but not the stall *hazard* itself — a stall near the ceiling still wastes a turn. The `/clear` thresholds below are unchanged.


## Claude Code's Graduated-Compaction Layers (consumed, not implemented)

Before the context window reaches the ceiling, the Claude Code runtime applies a **graduated-compaction** mechanism — five escalating layers that progressively reduce the live input before each model call, in escalation order:

```
Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact
```

These five layer names are recorded here as a provenance cross-reference, sourced from the public paper "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (arXiv:2604.14228; companion repository github.com/VILA-Lab/Dive-into-Claude-Code).

The orchestrator CONSUMES Claude Code's graduated-compaction layers; it does NOT implement them. Budget Reduction, Snip, Microcompact, Context Collapse, and Auto-Compact are Claude Code runtime internals — the harness sits ON TOP of Claude Code and cannot modify the native compaction loop. The `/clear` discipline and the model-specific thresholds below are the orchestrator-side behaviors that interact with the runtime's graduated compaction; they are not a reimplementation of it. The vocabulary is recorded so the `/clear` thresholds can name the runtime mechanism they sit atop.


## Detection Heuristics

The orchestrator estimates context usage **state-file-first**: it reads the
authoritative snapshot the statusline writes each render, and falls back to the
byte / system-reminder heuristics only when that snapshot is absent, stale, or
unparseable.

### 1. Authoritative snapshot — `.moai/state/context-usage.json`

The statusline persists a best-effort snapshot of raw context usage to
`<projectDir>/.moai/state/context-usage.json` on every render. When present and
valid, this file is the authoritative signal — prefer it over the estimation
heuristics below. Its fields:

- `raw_pct` — raw context-window usage (tokens ÷ window); the direct handoff signal
- `stage` — the two-stage handoff classification: `none` / `soft` / `hard`
- `session_id` / `writer_pid` / `captured_at` — validity-guard inputs (see §2)
- `context_window_size` / `tokens_used` / `band` — supporting context

Read `stage` and `raw_pct` directly rather than re-deriving usage from proxies.

### 2. Validity guard (do not resume another session's snapshot)

Trust the snapshot only when it belongs to the current session:

- **Real session id on both sides**: valid only when the record's `session_id`
  equals the current session id (last-writer-wins). A differing id → treat as
  stale and fall back to the heuristics (avoids resuming another session's usage).
- **No real session id (empty) on both sides**: validate by `captured_at`
  freshness (a generous, session-scoped window) instead of id equality, so the
  common single-session case still uses the snapshot. When two same-checkout
  sessions both lack a real id and share one file, the `writer_pid` discriminator
  distinguishes them; a reader that cannot supply its own writer identity treats
  a concurrent same-checkout case conservatively and falls back to the heuristics.
- **Mixed (one real id, one empty), unparseable, or absent**: fall back to the
  heuristics.

### 3. Fallback heuristics (snapshot absent, stale, or unparseable)

When the snapshot cannot be trusted, estimate context usage from four signals:

- Cumulative output bytes since session start (rough proxy)
- System reminder volume per turn (rule-file injections inflate input)
- Number of large tool results (each Read/Bash output >5 KB adds linear pressure)
- Number of Agent() invocations completed (each contributes to parent context on return)

Under-estimate when uncertain — premature `/clear` costs one paste; missed one costs a stalled stream.

### 4. Two-stage handoff marker + reachability limitation

The statusline appends a `/clear` hint to the context bar in two stages: a soft
`(⚠️/clear)` marker at the band's soft threshold, and a hard `(🛑/clear!)` marker
at an auto-compact-aware ceiling (`min(cap, auto-compact-threshold + margin)`).

Because the runtime's auto-compact fires near the auto-compact threshold of the
raw window, the hard ceiling is **frequently pre-empted** by auto-compact and
the hard stage **rarely fires** in practice — an intentional, documented
tradeoff of the auto-compact-aware formula. The hard marker is a strong upper
signal, not a guarantee; the doctrine makes no claim that the hard stage will
trigger on every session.

### 5. Guide-gated advisory (optional)

When the handoff guide flag is enabled, the orchestrator MAY surface a
state-file-derived advisory (for example, "raw usage at the hard stage —
consider `/clear`") alongside the automatic pre-clear announcement. This
advisory is doctrine-level guidance only: it adds no new runtime hook and never
gates the statusline marker or the snapshot write, both of which stay
unconditional.


---

Classification: Lazy companion — rationale, vocabulary, and detection mechanism only. Every
threshold and every obligation stays in `context-window-management.md`.

## Multi-session work — resume rather than re-establish

Work that spans sessions does not have to be rebuilt from a paste each time. `claude --continue` reopens the most recent session and `claude --resume` picks one from a list, both with context intact; `/rename` gives a session a durable name (`oauth-migration`) so it stays findable. Treat named sessions as branches — one per work stream, each with its own accumulated context.

This composes with the paste-ready handoff rather than replacing it. Resume is for continuing a session that still exists; the handoff (`session-handoff.md`) is for crossing a `/clear` or a machine boundary, where the previous context is gone by construction.
