---
description: "Executable workflow trace evidence and performance summary contract"
paths:
  - ".claude/skills/moai/workflows/project*.md"
  - ".claude/skills/moai/workflows/project/**"
  - ".claude/skills/moai/workflows/plan*.md"
  - ".claude/skills/moai/workflows/plan/**"
  - ".claude/skills/moai/workflows/run*.md"
  - ".claude/skills/moai/workflows/run/**"
  - ".claude/skills/moai/workflows/sync*.md"
  - ".claude/skills/moai/workflows/sync/**"
  - ".claude/hooks/moai/trace-ledger.sh"
---

# Workflow trace ledger contract

`TRACE PROBE` comments in workflow documents are activation hints only. A
comment, a pseudocode line, or an empty file is not execution evidence. When
`MOAI_TRACE_PHASES=1` is enabled, the orchestrator MUST append the corresponding
event through `.claude/hooks/moai/trace-ledger.sh record` to
`.moai/state/workflow-trace.jsonl`; when the flag is off, no trace is claimed.

Each JSONL row is append-only and contains:

- `stage` and `event` (`stage_start`, `stage_end`, `tool_start`, or `tool_end`)
- `parent_id` for the session/agent parent that owns the event
- `input_hash` for the exact routing or verification input
- `cohort` for the comparable S/M/L, cold/warm, or injected-failure cohort
- `cache_hit` (`hit`, `miss`, or `unknown`)
- `retry_cause` (`none` when no retry occurred)
- `duration_ms` measured by the caller, never invented from a comment

The helper rejects unsafe field values and non-numeric durations. The
`summary` command counts `tool_end` calls and calculates p50/p95 from recorded
durations, plus cache, retry, and observed-cohort counts. Compare S/M/L and
cold/warm or injected-failure cohorts only when their rows carry the same input
and cohort identity; missing rows remain an observation gap, not a zero-cost
result. Routing
decisions remain in `.moai/state/routing-ledger.jsonl`; this ledger adds
execution timing and trace identity without changing that schema.
