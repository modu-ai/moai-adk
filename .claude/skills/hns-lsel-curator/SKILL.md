---
name: hns-lsel-curator
description: >
  Local Self-Evolution Loop (LSEL) curator — the CLUSTER + drain engine for the
  GOOS-local PROPOSE→APPLY seam closure (SPEC-LSEL-LOCAL-EVOLUTION-001). Companion-offset
  drain of .moai/lessons-inbox.jsonl with a drain-side severity filter that drops the ~65%
  Bash-timeout/sandbox noise, event_key clustering with a frequency gate, and a
  Generative-Agents-style 1-10 importance score. Candidates stage at
  .moai/state/lsel/clusters.json. M1 = drain only (NO PROPOSE, NO APPLY, NO memory/ writes).
allowed-tools: Read, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.1.0"
  category: "harness"
  status: "active"
  updated: "2026-08-04"
  tags: "lsel,self-evolution,drain,cluster,harness,dogfood"
---

# hns-lsel-curator — LSEL CLUSTER + drain engine

> **Namespace:** `hns-lsel-*` is user-owned dogfood (CLAUDE.local.md §24). This skill is
> NOT mirrored into `internal/template/templates/` — it lives only in this repo. Graduation
> to `moai-lsel-*` + 16-language distribution is a separate SPEC (out of scope per spec.md §G).
>
> **M1 scope:** drain + cluster + stage candidates. NO PROPOSE (M2), NO APPROVE, NO APPLY (M3),
> and **NO writes to `memory/`** — candidates wait in `clusters.json` until M2 proposes them.

## What this skill does

The MoAI-ADK repo accumulates tool-failure stubs in `.moai/lessons-inbox.jsonl` (624 stubs at
M1 start, re-measured — a moving target). The constitution names the orchestrator as the drain
actor, but until this skill there was **zero mechanical drain code** — the drain existed only as
a doctrine paragraph (`moai-constitution.md:147`). This skill closes that gap in user-owned
surfaces, without touching the frozen Go applier (`internal/harness/applier.go:22` stays
`enableTriggerInjectionWrites=false` — REQ-LSEL-003: bypass, never unfreeze).

The drain is split into a **mechanical core** (`drain.sh`, deterministic, testable) and a
**model-mediated layer** (this SKILL.md + your judgment, invoked for M2+ importance refinement
and proposal drafting).

## The mechanical core — `drain.sh`

`drain.sh` is a portable bash + jq script that lives next to this SKILL.md. It performs the
deterministic half of the drain:

```
drain.sh --inbox <path-to-lessons-inbox.jsonl> --state-dir <path-to-lsel-state>
```

Pipeline (REQ-LSEL-009 + AC-LSEL-009 / AC-LSEL-010):

1. **Companion offset** — read `<state-dir>/drain-offset.json` (seed `{"offset":0}` if absent).
   The inbox is append-only and is NEVER mutated; the offset marks consumed stubs
   (SPEC-HARNESS-RATCHET-REWIRE-001 D3 companion-offset pattern).
2. **Slice** — read stubs from the offset onwards (`tail -n +<offset+1>`).
3. **Drain-side severity filter** (AC-LSEL-010) — discard noise BEFORE clustering:
   - `tool_failure:Bash:UnknownFailure` — the opaque ~65% timeout/sandbox bucket (the dominant
     noise share; report §2).
   - `tool_failure:Bash:SandboxViolation` — environment constraint, not a code defect.
   - any `*:TimeoutError` (Bash + MCP timeouts).
   The filter is drain-side because `internal/hook/failure_observer.go` (the inbox writer) is
   OUTSIDE the six loop-writable surfaces (plan.md §F.1 [DECISION RESOLVED]), so the loop cannot
   edit the writer — it filters on read instead.
4. **Cluster** by `event_key` with frequency count, first/last seen, and up to 3 sample summaries.
5. **Singleton gate** — discard clusters with `frequency < 2` (single-occurrence noise per the
   constitution Lessons Protocol drain paragraph).
6. **Importance** — score each survivor with a Generative-Agents-style 1-10 gate:
   `importance = min(10, frequency)` (frequency as proxy; the model augments this in M2+ with a
   severity hint and retrieval-weighted judgment).
7. **Emit** candidates to `<state-dir>/clusters.json`; advance the companion offset.

### `clusters.json` schema

```json
{
  "drained_at": "2026-08-04T08:41:00Z",
  "offset_before": 0,
  "offset_after": 624,
  "total_read": 624,
  "noise_discarded": 533,
  "singletons_discarded": 4,
  "candidates": [
    {
      "event_key": "tool_failure:Agent:UnknownFailure",
      "frequency": 41,
      "first_seen": "...",
      "last_seen": "...",
      "sample_summaries": ["...", "...", "..."],
      "source": "tool:Agent",
      "importance": 10
    }
  ]
}
```

### Empty-delta no-op

If the inbox has not grown past the offset, `drain.sh` writes an empty-candidate `clusters.json`
and leaves the offset unchanged. Not a failure (acceptance.md §E edge case).

## The model-mediated layer (you, when invoked)

`drain.sh` produces the deterministic candidate set. When this skill is invoked for a real
curation pass (M2+), your job on top of the mechanical output is:

- **Read `clusters.json`** and rank candidates by `importance` then `frequency`.
- **Augment importance** with a severity hint the mechanical core cannot see: a recurring
  `Bash:ExitError` cluster points at a real command-shape defect (high signal); a recurring
  `Agent:ContextCancelled` cluster may be session-teardown noise (lower signal). Record the
  rationale in the candidate's prose when you draft the M2 proposal — do NOT rewrite
  `clusters.json` (it is the mechanical artifact; your augmentation lives in the proposal).
- **Do NOT write to `memory/` in M1.** Candidates stage in `clusters.json` only. The first
  `feedback_*.md` topic file is produced by the M2 PROPOSE stage after retrieval-before-propose
  and self-critique (REQ-LSEL-010).

## What this skill does NOT do (M1 boundaries)

- **No PROPOSE** — shadow proposals land in M2 (`.moai/state/lsel/proposals/<id>/`).
- **No APPROVE / APPLY** — the parallel user-owned applier (`hns-lsel-applier`) is M3.
- **No edits to frozen doctrine** — `.claude/rules/moai/**`, `CLAUDE.md`,
  `internal/template/templates/**`, retained agents, `moai-*` skills, and the frozen Go
  applier / `curator_dispatch.go` are all byte-for-byte untouched (REQ-LSEL-001 / §B.3).
- **No new `.moai/config/sections/` file** — loop state lives under `.moai/state/lsel/`
  (a new section file would be wiped on `moai update`; plan.md §B.4 / AP-LSEL-005).
- **No `AskUserQuestion`** — this is a subagent-owned mechanism skill. On a missing input,
  return a structured blocker report; the orchestrator runs the user gate (CLAUDE.md §8).

## Verification (run before declaring a drain complete)

```bash
# 1. The drain mechanics (fixture-based characterization test — AC-LSEL-009/010):
.claude/skills/hns-lsel-curator/drain_test.sh

# 2. A real drain of the live backlog (re-measure the count first — it is a moving target):
LIVE_COUNT=$(wc -l < .moai/lessons-inbox.jsonl | tr -d ' ')
.claude/skills/hns-lsel-curator/drain.sh --inbox .moai/lessons-inbox.jsonl --state-dir .moai/state/lsel
jq '.offset_after == ($LIVE_COUNT|tonumber) and (.candidates | length) >= 1' .moai/state/lsel/clusters.json

# 3. M1 invariant — zero memory/ writes from the drain:
find memory -newer <drain-start-timestamp> -name 'feedback_*' 2>/dev/null | wc -l   # must be 0
```

## Characterization test

`drain_test.sh` (next to this SKILL.md) is the TDD RED→GREEN harness. It builds a synthetic
inbox with known noise + signal stubs, runs `drain.sh`, and asserts the drain semantics:
noise excluded pre-cluster, signal clustered with correct frequencies, singletons discarded,
offset advanced, candidates emitted, zero `memory/` writes, idempotent re-drain. Run it after
any edit to `drain.sh`.

## Cross-references

- **SPEC:** `.moai/specs/SPEC-LSEL-LOCAL-EVOLUTION-001/{spec,plan,acceptance,progress}.md`
- **Design report (SSOT):** `.moai/reports/moai-local-self-evolution-design-20260804.html`
  §6 stage 2 (CLUSTER), §10 P1, §11 mustFix B#1/B#3.
- **Frozen applier (reference only):** `internal/harness/applier.go:22`
  (`enableTriggerInjectionWrites=false`), `internal/harness/curator_dispatch.go`.
- **Constitution drain paragraph (the "0 Go code" stub this skill replaces):**
  `.claude/rules/moai/core/moai-constitution.md:147`.
- **Namespace guard:** `internal/template/split_namespace_test.go`,
  `internal/template/internal_content_leak_test.go` (extended in M2 — AC-LSEL-006).
