# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Why This Matters

Anthropic SSE streams have been observed to stall (`stream_idle_partial`) when very large prompts are produced near the upper end of the context window. Symptoms: model emits a few hundred bytes then the stream goes idle; orchestrator's outbound message hangs without a tool call. This is intermittent but predictable above the model-specific usage threshold.

Reference incident: 2026-04-25, SPEC-V3R2-WF-001 monolithic delegation. See feedback memory `feedback_large_spec_wave_split.md`.

## Context Window Targets

[HARD] Operational threshold is **model-specific** (revised 2026-05-09). Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom:

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 4.7 (1M) | 1,000,000 tokens | **75%** | ~750,000 tokens |
| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |

The model-specific threshold is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action. Both this rule and `session-handoff.md` Trigger #1 read from this same table.

## User Responsibilities

The user monitors context usage via the Claude Code statusline or `/cost` command and intervenes when usage crosses the model-specific threshold (75% on 1M models, 90% on 200K models).

[HARD] When usage crosses the model-specific threshold:
1. Save in-flight state to `.moai/specs/<SPEC-ID>/progress.md` if not already saved (orchestrator does this automatically)
2. Run `/clear` to flush the conversation context
3. Paste the **resume message** (provided by the orchestrator before the clear) to continue

[HARD] When usage crosses 95% on any model:
- The next action MUST be `/clear` — no further large work in the current session
- Stall risk is severe; agent invocations may fail mid-stream
- This is the absolute hard stop regardless of model class

## Orchestrator Responsibilities

The MoAI orchestrator MUST proactively recognize the model-specific boundary and prepare the user for a clean handoff.

[HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (75% on 1M, 90% on 200K), it MUST:
1. Stop initiating new large tool calls or `Agent()` delegations
2. Persist all in-flight progress to `.moai/specs/<SPEC-ID>/progress.md`
3. Emit a structured "resume message" the user can paste verbatim after `/clear`
4. Recommend `/clear` via natural-language guidance (this is a status announcement, not a question — `AskUserQuestion` not required)

[HARD] Resume message format: include all of the following so the next session is self-sufficient:
```
ultrathink. Wave <N> 이어서 진행. SPEC-<ID>부터 <approach 요약>.
applied lessons: <memory file names>.
progress.md 경로: .moai/specs/SPEC-<ID>/progress.md
다음 단계: <one-line command>.
완료 후: <next SPEC or /moai sync>.
```

The resume message is a verbatim hand-off — paste-ready, no editing required.

## Detection Heuristics

The orchestrator estimates context usage from these observable signals:

- Cumulative output bytes since session start (rough proxy)
- System reminder volume per turn (large rule-file injections inflate input rapidly)
- Number of large tool results received (each Read/Bash output >5 KB adds linear pressure)
- Number of Agent() invocations completed (each Agent context contributes to parent context on return)

When uncertain, prefer to under-estimate remaining capacity. A premature `/clear` recommendation costs one paste; a missed one costs a stalled stream and possibly lost work.

## Applies To

This rule applies to all MoAI workflows:
- `/moai plan`, `/moai run`, `/moai sync` — long phases that accumulate context
- Multi-SPEC waves (Wave 1 / Wave 2 multi-SPEC delegation) — most likely to hit the model-specific threshold
- Iterative loops (`/moai loop`, GAN loop) — context accumulates linearly per iteration

## Cross-references

- `.pi/generated/source/rules/moai/workflow/session-handoff.md` — paste-ready resume message canonical format and auto-memory integration. Trigger #1 of session-handoff.md consumes the model-specific threshold table from this file (1M = 75%, 200K = 90%). The two rules share the same threshold table; `/clear` recommendation and paste-ready emission both fire at the same boundary.
- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split mitigation for SPECs with 30+ tasks
- `.pi/generated/source/skills/moai/references/file-reading-optimization.md` — token budget per file read
- `output-styles/moai/moai.md` §6 (Persistence & Context Awareness) — orchestrator persistence pattern
- .pi/generated/source/CLAUDE.md §11 (Error Handling) — token-limit recovery flow

---

Source: 2026-04-25 stall incident analysis (debug log `[WARN] [Stall] stream_idle_partial`) + 2026-05-09 model-specific threshold revision (1M = 75%, 200K = 90%)
Status: HARD operational rule, applies to all sessions
