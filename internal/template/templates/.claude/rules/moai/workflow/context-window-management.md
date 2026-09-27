# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Context Window Targets

[ZONE:Evolvable] [HARD] Operational threshold is **model-specific**. Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom:

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 5.5 on the Anthropic API (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Opus 4.8 on the Anthropic API (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| GLM-5.3 via `moai glm`/`moai cg` (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Fable (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Sonnet 5 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| 200K sessions — Sonnet 4.6 / Opus 4.6 without `[1m]`; Opus 4.8+ running with a 200K window (e.g. on Bedrock / Google Cloud / Foundry); any native-1M model under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`; `sonnet` behind an LLM gateway (non-Anthropic `ANTHROPIC_BASE_URL`) unless `sonnet[1m]` is selected; Sonnet 4.5 / Opus 4.5 and earlier | 200,000 tokens | **90%** | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |

A session that matches both a 1M row and the 200K-sessions row takes the 200K row: the window the session actually runs with sets the threshold, not the model name. The model-specific threshold is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action. Both this rule and `session-handoff.md` Trigger #1 read from this same table.

## User Responsibilities

User monitors via Claude Code statusline / `/cost` and intervenes at threshold (50% on 1M / GLM-5.3, 90% on 200K).

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

[ZONE:Evolvable] [HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (50% on 1M / GLM-5.3, 90% on 200K), it MUST:
1. Stop initiating new large tool calls or `Agent()` delegations
2. Persist all in-flight progress to `.moai/specs/<SPEC-ID>/progress.md`
3. Emit a structured "resume message" the user can paste verbatim after `/clear`
4. Recommend `/clear` via natural-language guidance (status announcement, not a question — `AskUserQuestion` not required). This is the `push`-mode phrasing. While `interview.recommendation_mode` is `pull`, the announcement **states the observation and the available action without recommending**: it reports the measured usage against the model-specific threshold and names `/clear` as the action that resets it, leaving the decision to the user. The announcement still fires — the mode changes its phrasing, never whether the threshold is surfaced — and it remains a status announcement rather than a question in both modes. SSOT: `.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles

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

**Where the number comes from, and how far to trust it.** The snapshot does not measure the window
itself — the statusline writes through the percentage Claude Code already computed, so an upstream
metering change reaches every snapshot on disk without passing through any recomputation of ours.
That makes the snapshot a *relay*, and its confidence the runtime's confidence. Measurement puts it
within roughly a percentage point of occupancy rebuilt from the session transcript, with no
double-counting signature in either direction. So treat it as trustworthy to about a point, and
**re-derive rather than cite it when a verdict needs a number** — a relayed figure carries the
upstream's defects silently. The sample size, the command, the distribution, and the dated baseline:
`context-window-management-detail.md` § Snapshot confidence.

The statusline's two-stage `/clear` marker is a signal, not a guarantee: the hard stage is
frequently pre-empted by the runtime's auto-compact and rarely fires. Snapshot field list and the
guide-gated advisory: `context-window-management-detail.md` § Detection Heuristics.

## Applies To

All MoAI workflows: `/moai plan|run|sync`, multi-SPEC Epics, iterative loops (`/moai loop`, GAN loop).

## Cross-references

- `.claude/rules/moai/workflow/cache-aware-execution.md` — prompt-cache-aware `/clear` timing (its directive 4 permits an earlier `/clear` before a large multi-spawn batch, below the thresholds above) + gate placement and stagger-spawn ordering.
- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format + auto-memory integration. Trigger #1 consumes the model-specific threshold table from this file (1M = 50%, 200K = 90%); `/clear` recommendation and paste-ready emission both fire at the same boundary.
- `context-window-management-detail.md` — the lazy companion. Load it for § Why This Matters · § Claude Code's Graduated-Compaction Layers · § Reduction Ladder — cheaper moves before `/clear` (the four cheaper rungs and the checkpoint mechanics) · § GLM-5.3 context window · § Multi-session work — resume rather than re-establish · § Detection Heuristics · § Snapshot confidence

---

Status: HARD operational rule, applies to all sessions
