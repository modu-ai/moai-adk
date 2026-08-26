# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Why This Matters

Anthropic SSE streams stall (`stream_idle_partial`) near the context-window ceiling —
intermittently but predictably above the model-specific threshold. Claude Code's default-on stream
watchdog aborts and retries a stream idle for 5 minutes, which softens the consequence but not the
hazard: a stall near the ceiling still wastes a turn. The runtime additionally applies five
escalating compaction layers before the ceiling; the orchestrator **consumes** them and does not
reimplement them. Rationale and the layer vocabulary:
`context-window-management-detail.md` § Why This Matters and § Graduated-Compaction Layers.

## Context Window Targets

[ZONE:Evolvable] [HARD] Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom:

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Opus 4.8 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| GLM-5.3 via `moai glm`/`moai cg` (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Fable (256K) | 256,000 tokens | **90%** | ~230,000 tokens |
| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |

The model-specific threshold is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action. Both this rule and `session-handoff.md` Trigger #1 read from this same table.

### GLM-5.3 context window (Issue #653)

GLM-5.3 (z.ai, served via `moai glm` / `moai cg` GLM panes) is a genuine 1M-context model; operate it at the **50% (~500K)** handoff threshold, the same class as Opus 5 / Opus 4.8 (1M). Do NOT treat a `moai glm` session as a 200K session.

Caveat (Issue #653): Claude Code reports `context_window_size` based on the Claude slot (Opus=1M, Sonnet/Haiku=200K) regardless of provider, so raw telemetry (`effectiveWindow`) may show ~180K under GLM. This is an upstream misreport. MoAI corrects it: the statusline gauge uses `MOAI_STATUSLINE_CONTEXT_SIZE` and Claude Code auto-compact uses `CLAUDE_CODE_AUTO_COMPACT_WINDOW`, both resolved from the `glmContextWindows` table in `internal/statusline/memory.go` (glm-5.3 → 1,000,000) or the `llm.glm.context_windows` override. Trust the MoAI statusline CW%, not raw `effectiveWindow`.


## Reduction Ladder — cheaper moves before `/clear`

`/clear` is the heaviest reduction available: it discards the whole window, including the warm prompt cache, and forces the next turn to re-pay the entire always-loaded prefix. That cost is not incidental — it scales with the always-loaded footprint, so `/clear` gets more expensive as the rule tree grows. Reach for it when a full reset is genuinely what is wanted, not as the reflexive answer to a full window.

Four cheaper moves come first. Each targets a different cause of context growth, so pick by cause rather than working down the list:

| Move | Use when | Effect |
|------|----------|--------|
| `/btw <question>` | A side question would otherwise land in the transcript | Answer renders in a dismissible overlay and never enters conversation history — context does not grow at all |
| `/compact <instructions>` | The window is full but the *current* task must continue | Summarizes in place; the instructions steer what survives |
| `Esc Esc` / `/rewind` → **Summarize up to here** | Early exploration is spent but recent turns must stay verbatim | Compacts the old prefix, keeps the tail intact |
| `Esc Esc` / `/rewind` → restore a checkpoint | A line of attempts polluted the context, or the tree needs reverting | Restores conversation, files, or both, from a per-prompt snapshot |

Checkpoints are automatic (one per prompt) and persist across sessions, so an approach can be tried and abandoned rather than deliberated over. They track only Claude's own edits — external processes are invisible to them, and they are not a substitute for git.

`/clear` remains correct for a genuine task switch, and remains **mandatory** at the thresholds below. The ladder shortens how often those thresholds are reached; it does not move them.

### Multi-session work: resume rather than re-establish

`claude --continue` reopens the most recent session and `claude --resume` picks one from a list,
both with context intact; `/rename` gives a session a durable name so it stays findable. Resume
continues a session that still exists; the paste-ready handoff crosses a `/clear` or a machine
boundary, where the previous context is gone by construction — they compose rather than replace one
another. Detail: `context-window-management-detail.md`.

## User Responsibilities

User monitors via Claude Code statusline / `/cost` and intervenes at threshold (50% on 1M / GLM-5.3, 90% on 200K/256K).

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

[ZONE:Evolvable] [HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (50% on 1M / GLM-5.3, 90% on 200K/256K), it MUST:
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

The orchestrator estimates context usage **state-file-first**: it reads
`<projectDir>/.moai/state/context-usage/<session-id>.json`, the snapshot the statusline writes each
render, and prefers its `raw_pct` and `stage` fields over any proxy. The record is per session, so
the one named for the current session belongs to it by construction — no cross-session validity
check is needed. When it is absent or unparseable, usage is estimated from cumulative output bytes,
system-reminder volume, large tool results, and completed `Agent()` returns — under-estimating when
uncertain, since a premature `/clear` costs one paste and a missed one costs a stalled stream.

The statusline's two-stage `/clear` marker is a signal, not a guarantee: the hard stage is
frequently pre-empted by the runtime's auto-compact and rarely fires. Snapshot field list and the
guide-gated advisory: `context-window-management-detail.md` § Detection Heuristics.

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
