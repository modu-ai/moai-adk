# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Why This Matters

Anthropic SSE streams stall (`stream_idle_partial`) near the context window ceiling — intermittent but predictable above the model-specific threshold. Reference: 2026-04-25 incident (`feedback_large_spec_wave_split.md`).

## Claude Code's Graduated-Compaction Layers (consumed, not implemented)

Before the context window reaches the ceiling, the Claude Code runtime applies a **graduated-compaction** mechanism — five escalating layers that progressively reduce the live input before each model call, in escalation order:

```
Budget Reduction → Snip → Microcompact → Context Collapse → Auto-Compact
```

These five layer names are recorded here as a provenance cross-reference, sourced from the public paper "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" (arXiv:2604.14228; companion repository github.com/VILA-Lab/Dive-into-Claude-Code).

The orchestrator CONSUMES Claude Code's graduated-compaction layers; it does NOT implement them. Budget Reduction, Snip, Microcompact, Context Collapse, and Auto-Compact are Claude Code runtime internals — the harness sits ON TOP of Claude Code and cannot modify the native compaction loop. The `/clear` discipline and the model-specific thresholds below are the orchestrator-side behaviors that interact with the runtime's graduated compaction; they are not a reimplementation of it. The vocabulary is recorded so the `/clear` thresholds can name the runtime mechanism they sit atop.

## Context Window Targets

[ZONE:Evolvable] [HARD] Operational threshold is **model-specific** (revised 2026-05-09). Larger windows tolerate higher percentage utilization before stall risk dominates; smaller windows hit the operational ceiling later in percentage terms but with less absolute headroom:

| Model class | Window | Handoff threshold | Absolute ceiling |
|-------------|--------|-------------------|------------------|
| Opus 4.8 (1M) | 1,000,000 tokens | **50%** | ~500,000 tokens |
| Sonnet/Opus standard (200K) | 200,000 tokens | **90%** | ~180,000 tokens |
| Haiku (200K) | 200,000 tokens | **90%** | ~180,000 tokens |

The model-specific threshold is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action. Both this rule and `session-handoff.md` Trigger #1 read from this same table.

## User Responsibilities

User monitors via Claude Code statusline / `/cost` and intervenes at threshold (50% on 1M, 90% on 200K).

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

[ZONE:Evolvable] [HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching the model-specific threshold (50% on 1M, 90% on 200K), it MUST:
1. Stop initiating new large tool calls or `Agent()` delegations
2. Persist all in-flight progress to `.moai/specs/<SPEC-ID>/progress.md`
3. Emit a structured "resume message" the user can paste verbatim after `/clear`
4. Recommend `/clear` via natural-language guidance (status announcement, not a question — `AskUserQuestion` not required)

[ZONE:Evolvable] [HARD] Resume message format: include all of the following so the next session is self-sufficient:
```
ultrathink. Sprint <N> 이어서 진행. SPEC-<ID>부터 <approach 요약>.
applied lessons: <memory file names>.
progress.md 경로: .moai/specs/SPEC-<ID>/progress.md
다음 단계: <one-line command>.
완료 후: <next SPEC or /moai sync>.
```

Paste-ready, no editing required.

## Detection Heuristics

Orchestrator estimates context usage from four signals:

- Cumulative output bytes since session start (rough proxy)
- System reminder volume per turn (rule-file injections inflate input)
- Number of large tool results (each Read/Bash output >5 KB adds linear pressure)
- Number of Agent() invocations completed (each contributes to parent context on return)

Under-estimate when uncertain — premature `/clear` costs one paste; missed one costs a stalled stream.

## Applies To

All MoAI workflows: `/moai plan|run|sync`, multi-SPEC sprints, iterative loops (`/moai loop`, GAN loop).

## Cross-references

- `.claude/rules/moai/workflow/session-handoff.md` — paste-ready resume format + auto-memory integration. Trigger #1 consumes the model-specific threshold table from this file (1M = 50%, 200K = 90%); `/clear` recommendation and paste-ready emission both fire at the same boundary.
- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split mitigation for 30+ task SPECs
- `.claude/skills/moai/references/file-reading-optimization.md` — token budget per file read
- `output-styles/moai/moai.md` §6 (Persistence & Context Awareness)
- CLAUDE.md §11 (Error Handling) — token-limit recovery flow

---

Status: HARD operational rule, applies to all sessions
