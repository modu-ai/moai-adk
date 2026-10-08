# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Context Window Targets

Operational threshold is **model-specific**. Match the row for the window the session actually runs with:

| Model class | Window | Handoff threshold | Absolute ceiling |
|---|---|---|---|
| Opus 5.5 / Opus 4.8 / GLM-5.3 via `moai glm`/`moai cg` / Fable / Sonnet 5.5, native 1M | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Any native-1M model under `CLAUDE_CODE_DISABLE_1M_CONTEXT=1`; Sonnet 4.6 / Opus 4.6 without `[1m]`; models below the 1M-default line on an LLM gateway (non-Anthropic `ANTHROPIC_BASE_URL`; CC 2.1.285+ defaults 1M there for Sonnet 5+ / Opus 4.7+ / Fable, and on Bedrock / Vertex / Foundry for Opus 4.7+ and Fable); a 200K-capped gateway — run `/autocompact 200k`; Sonnet 4.5 / Opus 4.5 and earlier; Haiku | 200,000 tokens | **90%** | ~180,000 tokens |

A session that matches both a 1M row and the 200K row takes the 200K row: the window the session actually runs with sets the threshold, not the model name. The threshold is the operational ceiling. Beyond it, plan a `/clear` before the next non-trivial action. `session-handoff.md` Trigger #1 reads this same table.

## User Responsibilities

Monitor usage via the Claude Code statusline or `/cost`. Intervene at threshold (50% on 1M / GLM-5.3, 90% on 200K).

[ZONE:Evolvable] [HARD] When usage crosses the model-specific threshold:
1. Save in-flight state to `.moai/specs/<SPEC-ID>/progress.md` (the orchestrator does this automatically)
2. Run `/clear`
3. Paste the resume message the orchestrator provided before the clear

[ZONE:Evolvable] [HARD] When usage crosses 95% on any model:
- Run `/clear` next — start no further large work in the session
- Stall risk is severe; agent invocations may fail mid-stream
- This is the absolute hard stop on every model class

## Orchestrator Responsibilities

The orchestrator MUST recognize the model-specific boundary early and prepare the user for a clean handoff.

[ZONE:Evolvable] [HARD] Pre-clear announcement: when accumulated context (input + output) nears the model-specific threshold (50% on 1M / GLM-5.3, 90% on 200K), the orchestrator MUST:
1. Stop new large tool calls and `Agent()` delegations
2. Persist all in-flight progress to `.moai/specs/<SPEC-ID>/progress.md`
3. Emit a structured resume message the user can paste verbatim after `/clear`
4. Recommend `/clear` in a status announcement, never a question (`AskUserQuestion` not required). Under `interview.recommendation_mode: pull`, state the measured usage against the threshold and name `/clear` as the action that resets it; leave the decision to the user. The announcement fires in both modes — the mode changes only its phrasing. SSOT: `.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles

[ZONE:Evolvable] [HARD] Resume message format — include all of the following so the next session is self-sufficient (locale renderings per `session-handoff.md` § Localization Table; do not define a parallel format):
```
Resume Epic <N>. SPEC-<ID> — <approach summary>.
applied lessons: <memory file names>.
progress.md path: .moai/specs/SPEC-<ID>/progress.md
Run: <one-line command>.
After merge: <next SPEC or /moai sync>.
```

Paste-ready, no editing required.

## Detection Heuristics

Estimate context usage state-file-first: read `<projectDir>/.moai/state/context-usage/<session-id>.json` (the statusline's per-session snapshot) and prefer its `raw_pct` and `stage` fields. If it is absent or unparseable, estimate from cumulative output, system-reminder volume, and completed `Agent()` returns. Under-estimate when uncertain: a premature `/clear` costs one paste; a missed one costs a stalled stream.

**Snapshot confidence.** The snapshot relays the percentage Claude Code already computed; the statusline writes it through without recomputation, so an upstream metering change reaches every snapshot on disk. Measurement puts it within about a percentage point of transcript-rebuilt occupancy, with no double-counting signature either way. Treat it as trustworthy to about a point. Re-derive the number rather than cite it when a verdict needs it — a relayed figure carries the upstream's defects silently. Sample size, command, distribution, and dated baseline: `context-window-management-detail.md` § Snapshot confidence.

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
