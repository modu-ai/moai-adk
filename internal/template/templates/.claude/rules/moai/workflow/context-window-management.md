# Context Window Management

Long-horizon session continuity guidance for both users and the MoAI orchestrator.

## Why This Matters

Anthropic SSE streams have been observed to stall (`stream_idle_partial`) when very large prompts are produced near the upper end of the context window. Symptoms: model emits a few hundred bytes then the stream goes idle; orchestrator's outbound message hangs without a tool call. This is intermittent but predictable above the 75% usage threshold.

Reference incident: 2026-04-25, SPEC-V3R2-WF-001 monolithic delegation. See feedback memory `feedback_large_spec_wave_split.md`.

## Context Window Targets

| Model class | Window | 75% threshold |
|-------------|--------|---------------|
| Opus 4.7 (1M) | 1,000,000 tokens | ~750,000 |
| Sonnet/Opus standard (200K) | 200,000 tokens | ~150,000 |
| Haiku (200K) | 200,000 tokens | ~150,000 |

The 75% line is the operational ceiling — beyond it, plan for a `/clear` before the next non-trivial action.

## User Responsibilities

The user monitors context usage via the Claude Code statusline or `/cost` command and intervenes when usage crosses 75%.

[HARD] When usage crosses 75%:
1. Save in-flight state to `.moai/specs/<SPEC-ID>/progress.md` if not already saved (orchestrator does this automatically)
2. Run `/clear` to flush the conversation context
3. Paste the **resume message** (provided by the orchestrator before the clear) to continue

[HARD] When usage crosses 90%:
- The next action MUST be `/clear` — no further large work in the current session
- Stall risk is severe; agent invocations may fail mid-stream

## Orchestrator Responsibilities

The MoAI orchestrator MUST proactively recognize the 75% boundary and prepare the user for a clean handoff.

[HARD] Pre-clear announcement: When the orchestrator detects accumulated context (input + output) approaching 75%, it MUST:
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
- System reminder volume per turn (large rule-file injections inflate input rapidly — typically 10-50K tokens per `<system-reminder>` block injection)
- Number of large tool results received (each Read/Bash output >5 KB adds linear pressure)
- Number of Agent() invocations completed (only the **final response message** of each sub-agent contributes to parent context — typically 0.5-3K tokens per agent return, NOT the agent's internal total_tokens)

When uncertain, prefer to under-estimate remaining capacity. A premature `/clear` recommendation costs one paste; a missed one costs a stalled stream and possibly lost work.

## Sub-Agent Token Accounting [HARD]

**Common mistake**: treating an Agent()'s reported `total_tokens` as parent context consumption. This produces wildly inflated estimates and triggers premature `/clear` recommendations. Resolved 2026-04-27.

The truth:
- Each Agent() spawns a **separate context window**. The sub-agent consumes its own input + output + thinking + tool-call tokens internally. That internal consumption is reported as `total_tokens` in the agent return message.
- Only the sub-agent's **final user-facing response message** crosses back into the parent context. That is typically a structured report of 500-3000 tokens, NOT the multi-hundred-K internal total.
- Therefore: **do NOT sum `total_tokens` from Agent returns into the parent context estimate**.

[HARD] Correct estimation formula for parent context:
```
parent_context ≈
  initial_system_prompt (~30-60K, includes CLAUDE.md + autoloaded rules)
  + sum(system_reminders this session)        # each ~5-50K
  + sum(user_messages this session)           # each ~0.5-5K
  + sum(orchestrator responses this session)  # each ~1-10K
  + sum(tool_results received)                # Bash/Read each ~0.5-30K
  + sum(agent_FINAL_responses)                # NOT total_tokens; each ~0.5-3K
```

For Opus 4.7 (1M context), even an aggressive session with 5 agent delegations + 20 large tool calls + 30 system reminders typically lands at 200-400K — well under the 750K threshold. Premature `/clear` advice based on summing agent total_tokens is a HARD rule violation.

[HARD] Authoritative measurement: when in doubt, the user runs `/cost` or checks the Claude Code statusline. Orchestrator's heuristic is for proactive warnings only; it MUST NOT override the actual measurement.

## Applies To

This rule applies to all MoAI workflows:
- `/moai plan`, `/moai run`, `/moai sync` — long phases that accumulate context
- Multi-SPEC waves (Wave 1 / Wave 2 multi-SPEC delegation) — most likely to hit 75%
- Iterative loops (`/moai loop`, GAN loop) — context accumulates linearly per iteration

## Cross-references

- `feedback_large_spec_wave_split.md` (auto-memory) — wave-split mitigation for SPECs with 30+ tasks
- `.claude/rules/moai/workflow/file-reading-optimization.md` — token budget per file read
- `output-styles/moai/moai.md` §6 (Persistence & Context Awareness) — orchestrator persistence pattern
- CLAUDE.md §11 (Error Handling) — token-limit recovery flow

---

Source: 2026-04-25 stall incident analysis (debug log `[WARN] [Stall] stream_idle_partial`)
Status: HARD operational rule, applies to all sessions
