# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Why This Matters

Anthropic SSE streams stall (`stream_idle_partial`) near the context window ceiling — intermittent but predictable above the model-specific threshold. Reference: large-SPEC SSE-stall mitigation.

## Claude Code's Graduated-Compaction Layers (consumed, not implemented)

Before the context window reaches the ceiling, the Claude Code runtime applies a **graduated-compaction** mechanism — five escalating layers that progressively reduce the live input before each model call, in escalation order:

```
Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact
```

These five layer names are recorded here as a provenance cross-reference, sourced from the public paper "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (arXiv:2604.14228; companion repository github.com/VILA-Lab/Dive-into-Claude-Code).

The orchestrator CONSUMES Claude Code's graduated-compaction layers; it does NOT implement them. Budget Reduction, Snip, Microcompact, Context Collapse, and Auto-Compact are Claude Code runtime internals — the harness sits ON TOP of Claude Code and cannot modify the native compaction loop. The `/clear` discipline and the model-specific thresholds below are the orchestrator-side behaviors that interact with the runtime's graduated compaction; they are not a reimplementation of it. The vocabulary is recorded so the `/clear` thresholds can name the runtime mechanism they sit atop.

## Context Window Targets

[ZONE:Evolvable] [HARD] Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom:

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 4.8 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| GLM-5.2 via `moai glm`/`moai cg` (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Opus/Fable (256K) | 256,000 tokens | **90%** | ~230,000 tokens |
| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |

The model-specific threshold is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action. Both this rule and `session-handoff.md` Trigger #1 read from this same table.

### GLM-5.2 context window (Issue #653)

GLM-5.2 (z.ai, served via `moai glm` / `moai cg` GLM panes) is a genuine 1M-context model; operate it at the **50% (~500K)** handoff threshold, the same class as Opus 4.8 (1M). Do NOT treat a `moai glm` session as a 200K session.

Caveat (Issue #653): Claude Code reports `context_window_size` based on the Claude slot (Opus=1M, Sonnet/Haiku=200K) regardless of provider, so raw telemetry (`effectiveWindow`) may show ~180K under GLM. This is an upstream misreport. MoAI corrects it: the statusline gauge uses `MOAI_STATUSLINE_CONTEXT_SIZE` and Claude Code auto-compact uses `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, both resolved from the `glmContextWindows` table in `internal/statusline/memory.go` (glm-5.2 → 1,000,000) or the `llm.glm.context_windows` override. Trust the MoAI statusline CW%, not raw `effectiveWindow`.

## User Responsibilities

User monitors via Claude Code statusline / `/cost` and intervenes at threshold (50% on 1M / GLM-5.2, 90% on 200K/256K).

[ZONE:Evolvable] [HARD] When usage crosses the model-specific threshold:
1. Save in-flight state to `.moai/specs/<SPEC-ID>/progress.md` if not already saved (orchestrator does this automatically)
2. Run `/clear` to flush the conversation context
3. Paste the **resume message** (provided by the orchestrator before the clear) to continue

[ZONE:Evolvable] [HARD] When usage crosses 95% on any model:
- The next action MUST be `/clear` — no further large work in the current session
- Stall risk is severe; agent invocations may fail mid-stream
- This is the absolute hard stop regardless of model class

## Orchestrator Responsibilities

The orchestrator MUST proactively recognize the model-specific boundary and prepare the user for a clean handoff.

[ZONE:Evolvable] [HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (50% on 1M / GLM-5.2, 90% on 200K/256K), it MUST:
1. Stop initiating new large tool calls or `Agent()` delegations
2. Persist all in-flight progress to `.moai/specs/<SPEC-ID>/progress.md`
3. Emit a structured "resume message" the user can paste verbatim after `/clear`
4. Recommend `/clear` via natural-language guidance (status announcement, not a question — `AskUserQuestion` not required)

[ZONE:Evolvable] [HARD] Resume message format: include all of the following so the next session is self-sufficient (locale renderings per `session-handoff.md` § Localization Table — do not redefine a parallel format here):
```
ultrathink. Resume Epic <N>. SPEC-<ID> — <approach summary>.
applied lessons: <memory file names>.
progress.md path: .moai/specs/SPEC-<ID>/progress.md
Run: <one-line command>.
After merge: <next SPEC or /moai sync>.
```

Paste-ready, no editing required.

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

## Applies To

All MoAI workflows: `/moai plan|run|sync`, multi-SPEC Epics, iterative loops (`/moai loop`, GAN loop).

## Cross-references

- `.claude/rules/moai/workflow/cache-aware-execution.md` — prompt-cache-aware `/clear` timing (its directive 4 permits an earlier `/clear` before a large multi-spawn batch, below the thresholds above) + gate placement and stagger-spawn ordering.
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format + auto-memory integration. Trigger #1 consumes the model-specific threshold table from this file (1M = 50%, 200K = 90%); `/clear` recommendation and paste-ready emission both fire at the same boundary.
- large-SPEC split mitigation
- `.claude/skills/moai/references/file-reading-optimization.md` — token budget per file read
- `output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- CLAUDE.md §11 (Error Handling) — token-limit recovery flow

---

Status: HARD operational rule, applies to all sessions
