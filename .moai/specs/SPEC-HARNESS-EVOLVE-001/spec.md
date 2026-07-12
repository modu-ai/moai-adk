---
id: SPEC-HARNESS-EVOLVE-001
title: "Routing Observation Ledger — Loop 0 (Generator) of the self-evolving harness"
version: "0.1.0"
status: draft
created: 2026-07-12
updated: 2026-07-12
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/harness/routing, internal/cli, .claude/skills/moai"
lifecycle: spec-anchored
tags: "harness-evolve-epic, routing-ledger, loop0, observation, stop-hook, jsonl, self-evolving-harness"
era: V3R6
tier: M
---

# SPEC-HARNESS-EVOLVE-001 — Routing Observation Ledger (Loop 0 / Generator)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-12 | 0.1.0 | Initial plan-phase draft (Tier M, 21 REQ / 26 AC). First SPEC (M1) of the HARNESS-EVOLVE Epic (5 SPECs + 2 horizons) per the approved design SSOT `.moai/reports/harness-self-evolving-redesign-final-20260712.html` (§4 3-Zone contract, §5 Loop 0 ledger schema incl. deltas A2/A4, §7 M1 milestone). Creates the `routing-ledger.jsonl` observation surface (schema v1), the `internal/harness/routing/` Go writer/reader, Stop-hook outcome capture riding the existing `harness-observe-stop` handler, and the workflow-skill recording obligations. No Curator writes, no CLAUDE.md/local.md mutation, no tier promotion (EVOLVE-002/003 territory). 2 open clarifications tracked in plan.md. | manager-spec |

## §A. Context and Intent

Today, `/moai` routing decisions — natural-language request → subcommand branch →
Phase 0.95 mode → tier/harness level → eventual outcome — are **never observed**.
The existing harness observation surface (`internal/harness/observer.go` →
`.moai/harness/usage-log.jsonl`) watches generated `harness-*` artifact usage, a
different subject. Design gap G1: "routing decisions are not observed — no raw
material for harness learning" (design SSOT §3), plus new gap G8: loop/goal
convergence trajectories and subagent delegation trajectories are not structurally
preserved on any surface (§3, identified via infographic cards 3·5).

This SPEC implements **Loop 0 (observation / Generator)** of the 3-Loop
self-evolving harness architecture (design SSOT §5, grounded in Lilian Weng's
"Harness Engineering for Self-Improvement"): a per-decision **routing ledger**
(`.moai/state/routing-ledger.jsonl`, append-only JSONL, schema v1) that later
Epic SPECs — EVOLVE-002 (Curator editable surfaces) and EVOLVE-003 (tier-surface
mapping + gates + re-proposal suppression) — will consume as Loop 1/Loop 2 input.

Three anti-fabrication principles bind the design (design SSOT §5 Loop 0 +
`verification-claim-integrity.md` §1.1):

1. **Machine signals only** — `outcome` derives from exit codes, audit scores,
   and gate results; never from model self-report. Finalization authority lives
   in the Stop hook (a mechanical actor), not in the orchestrator's prose.
2. **Privacy / template neutrality** — `request_digest` never carries verbatim
   user text (hash or coarse class only).
3. **Evidence-or-null** — where no machine signal exists for a field (e.g.
   `goal_converged`), the ledger records `null` rather than an inferred value.

**Boundary principle.** This SPEC is the Epic's foundation observation layer.
It writes ONLY the ledger + pending-row state under `.moai/state/` and the
recording obligations in the `/moai` workflow skill docs. It performs NO
learning, NO promotion, NO managed-block writes — the Curator write layer is
EVOLVE-002, the gates/registries are EVOLVE-003 (see §E Exclusions).

## §B. Scope Summary

**In scope**:
- `routing-ledger.jsonl` **schema v1** at `.moai/state/routing-ledger.jsonl`
  (append-only JSONL, runtime state, gitignored via the existing `.moai/state/`
  rule — same family as `context-usage.json`). Fields per §D.1, including
  design delta A2 (loop/goal convergence trajectory) and A4 (subagent
  delegation trajectory).
- Go writer/reader package `internal/harness/routing/`: append-only writer
  (O_APPEND single-line append, concurrent-session safe), reader with filters
  (subcommand / outcome / time window), request digest utility, deterministic
  `DeriveOutcome` machine-signal precedence, pending-row store
  (`.moai/state/routing-pending-<session>.json`).
- Stop-hook outcome capture: extend the EXISTING `moai hook harness-observe-stop`
  Go handler (already registered in settings.json Stop hooks, async, 5s) with a
  self-gated routing finalizer — finalize only when a pending routing row exists
  AND its evidence derives a terminal outcome; no-op otherwise. No new hook
  wrapper, no settings.json registration change.
- CLI recording surface: `moai harness ledger record | evidence | list` verbs
  (the orchestrator's dispatch-time recording + evidence-append + read entry
  points).
- Workflow skill wiring: `.claude/skills/moai/SKILL.md` router + `workflows/plan.md`
  / `workflows/run.md` / `workflows/sync.md` bodies gain the recording
  obligation (dispatch-time `ledger record`, pipeline evidence appends).
  Template-First: template mirrors edited first, then live copies, then
  `make build`.

**Preserve**:
- `internal/harness/observer.go` + `.moai/harness/usage-log.jsonl` untouched —
  the routing ledger is a SEPARATE surface (different subject: `/moai` body
  routing vs generated `harness-*` usage).
- The existing `harness-observe-stop` behavior (usage-log Stop event,
  auto-classify, auto-propose chains) — the routing finalizer is additive and
  fail-open.
- The `learning.enabled` gate semantics of the harness-observe family
  (default enabled; `false` → no-op).
- The 5s MoAI hook timeout policy; hook failures never block session end.
- Template neutrality (§25): the MECHANISM ships to templates (Go binary +
  skill-doc mirrors); the ledger DATA never does.

**Out of scope** — see §E.

## §C. Requirements (GEARS notation)

### C.1 Ledger schema v1 (Loop 0 observation record)

- **REQ-HEV-001** (Ubiquitous): The routing ledger shall persist at
  `.moai/state/routing-ledger.jsonl` as append-only JSONL with an explicit
  `schema_version: 1` field per row, in the gitignored runtime-state family
  (covered by the existing `.moai/state/` gitignore rule).
- **REQ-HEV-002** (Ubiquitous): Each ledger row shall carry the core routing
  fields: `ts` (RFC3339 UTC), `session_id`, `model_class`
  (opus|fable|sonnet|glm|haiku|unknown), `request_digest`,
  `matched_subcommand`, `mode_selected` (Phase 0.95 mode), `tier` (S|M|L|null),
  `harness_level` (minimal|standard|thorough|null), `clarify_rounds` (int ≥ 0),
  and `outcome` (success|fail|abort|reroute).
- **REQ-HEV-003** (Ubiquitous — design delta A2): Each ledger row shall carry
  the loop/goal convergence trajectory fields: `loop_iterations` (int ≥ 0),
  `goal_converged` (bool|null), `convergence_class`
  (converged|ceiling-exit|diverged|null).
- **REQ-HEV-004** (Ubiquitous — design delta A4): Each ledger row shall carry a
  `delegations` array whose entries are
  `{agent, cycle_type?, outcome, blocker?}` — the subagent delegation
  trajectory of the routed pipeline.
- **REQ-HEV-005** (Unwanted behavior — privacy): The ledger writer shall not
  persist verbatim user request text anywhere (`request_digest` is a truncated
  SHA-256 hex of the request, computed in-process; the raw text is never
  written to disk by this package).
- **REQ-HEV-006** (Unwanted behavior — machine-signal-only): The ledger shall
  not accept model self-reported prose as the basis for `outcome`;
  `evidence_refs` entries shall carry machine signals only (`audit_score`,
  `gate_exit`, `verify_path` kinds). When no machine signal exists for a
  convergence field, the writer shall record `null` for that field rather than
  an inferred value.

### C.2 Writer / reader package (`internal/harness/routing/`)

- **REQ-HEV-007** (Ubiquitous): The `internal/harness/routing` package shall
  provide an append-only writer that appends exactly one JSONL line per
  finalized row using `O_APPEND|O_CREATE|O_WRONLY` single-write semantics, safe
  under concurrent sessions (no read-modify-write of the ledger file; pending
  state is per-session isolated).
- **REQ-HEV-008** (Ubiquitous): The package shall provide a reader that streams
  ledger rows with composable filters: by `matched_subcommand`, by `outcome`,
  and by time window (since/until), skipping malformed lines fail-open (count
  reported, never a panic).
- **REQ-HEV-009** (Unwanted behavior — separation): The routing package shall
  not write to `.moai/harness/usage-log.jsonl` and shall not import or reuse
  the usage-log `Event` schema types — the two observation surfaces stay
  separate (different subjects).
- **REQ-HEV-010** (Compound — reroute): **While** a pending routing row exists
  for the current session, **When** a new `ledger record` call arrives for the
  same session, the writer shall finalize the earlier pending row with
  `outcome: reroute` before creating the new pending row.

### C.3 Outcome capture (pending-row lifecycle + Stop hook, Phase Ω)

- **REQ-HEV-011** (Event-driven): **When** the orchestrator dispatches a
  `/moai` routing decision, the `moai harness ledger record` CLI verb shall
  create a pending row at `.moai/state/routing-pending-<session_id>.json`
  carrying the dispatch-time fields (request digest, subcommand, mode, tier,
  harness level, clarify rounds, model class).
- **REQ-HEV-012** (Compound — self-gated Stop finalize): **While** a pending
  routing row exists for the session, **When** the Stop hook fires
  (`moai hook harness-observe-stop`), the handler shall evaluate
  `DeriveOutcome(evidence_refs)`; **When** the derivation yields a terminal
  outcome, the handler shall finalize the row (append to the ledger, delete the
  pending file); **While** the evidence is non-terminal (pipeline still in
  flight across turns), the handler shall leave the row pending. **While** no
  pending row exists, the handler shall no-op for routing (self-gate — the Stop
  hook fires every turn-end).
- **REQ-HEV-013** (Ubiquitous — deterministic derivation): The
  `DeriveOutcome` function shall derive `outcome` deterministically from
  machine evidence per fixed precedence: (1) explicit abort marker → `abort`;
  (2) any `gate_exit` evidence with non-zero exit → `fail`; (3) at least one
  terminal passing machine signal (declared-final `gate_exit` 0, audit verdict
  evidence, or verify-path evidence marked terminal) → `success`;
  (4) otherwise → non-terminal (stay pending). The function shall accept no
  free-text outcome override.
- **REQ-HEV-014** (Event-driven — stale sweep): **When** `ledger record` runs
  and finds a pending row belonging to a different session (or a stale pending
  row older than the staleness threshold), it shall finalize that row with
  `outcome: abort` before creating the new pending row (lazy abort sweep — no
  SessionEnd hook extension in this SPEC).
- **REQ-HEV-015** (Unwanted behavior — fail-open + budget): The routing
  finalize path shall not block session end: all errors are logged to stderr
  and swallowed (matching the existing `harness-observe-stop` fail-open
  pattern), and the added work shall stay within the 5s MoAI hook timeout
  policy (single pending-file read + single O_APPEND write; no network, no
  subprocess).
- **REQ-HEV-016** (Capability gate — learning gate inheritance): **Where**
  `.moai/config/sections/harness.yaml` sets `learning.enabled: false`, the
  routing capture (record / evidence / Stop finalize) shall no-op, inheriting
  the existing harness-observe family gate (default: enabled).

### C.4 Workflow skill wiring (recording obligation)

- **REQ-HEV-017** (Ubiquitous — router registration): The
  `.claude/skills/moai/SKILL.md` router shall carry the routing-ledger
  recording obligation: when dispatching a subcommand/workflow, the
  orchestrator records the routing decision via `moai harness ledger record`
  (dispatch-time), and appends machine evidence via `moai harness ledger
  evidence` at pipeline gate points.
- **REQ-HEV-018** (Ubiquitous — workflow-body wiring): The
  `workflows/plan.md`, `workflows/run.md`, and `workflows/sync.md` bodies shall
  each carry the recording obligation at their phase boundaries (dispatch
  record reference + at least one evidence-append point naming a machine
  signal: plan-audit verdict, gate exit, or verify path).
- **REQ-HEV-019** (Ubiquitous — Template-First): Every skill-doc edit shall be
  made Template-First (`internal/template/templates/.claude/skills/moai/...`
  FIRST, then the live `.claude/` copy, then `make build`); template
  neutrality (no internal SPEC IDs / dates / commit SHAs introduced into
  `internal/template/templates/**`) shall be preserved.
- **REQ-HEV-020** (Unwanted behavior — no data in templates): The template
  tree shall never ship ledger DATA: no `routing-ledger.jsonl`, no pending-row
  files, and no seeded learning rows shall exist under
  `internal/template/templates/**` — the MECHANISM (Go binary behavior +
  skill-doc recording obligations) ships; the observation DATA never does.

### C.5 Go quality invariants

- **REQ-HEV-021** (Ubiquitous): The `internal/harness/routing` package shall
  reach ≥ 90% statement coverage (hook-adjacent code), use table-driven tests
  with `t.TempDir()` isolation, wrap errors with `%w`, and set no OTEL
  environment variables via `t.Setenv` in parallel tests.

## §D. Reference — ledger schema v1 (SSOT)

### D.1 Row schema (JSONL, one object per line)

```json
{
  "schema_version": 1,
  "ts": "2026-07-12T03:04:05Z",
  "session_id": "<uuid-or-empty>",
  "model_class": "opus|fable|sonnet|glm|haiku|unknown",
  "request_digest": "sha256:a1b2c3d4e5f6",
  "request_class": "feature|bugfix|refactor|docs|question|pipeline|other",
  "matched_subcommand": "plan|run|sync|project|fix|loop|clean|mx|review|codemaps|gate|feedback|harness|moai",
  "mode_selected": "trivial|background|parallel|sub-agent|workflow|null",
  "tier": "S|M|L|null",
  "harness_level": "minimal|standard|thorough|null",
  "clarify_rounds": 0,
  "outcome": "success|fail|abort|reroute",
  "loop_iterations": 0,
  "goal_converged": null,
  "convergence_class": "converged|ceiling-exit|diverged|null",
  "delegations": [
    { "agent": "manager-develop", "cycle_type": "tdd", "outcome": "success|fail|abort", "blocker": null }
  ],
  "evidence_refs": [
    { "kind": "gate_exit",   "value": "0",    "ref": "go test ./...", "terminal": true },
    { "kind": "audit_score", "value": "0.91", "ref": ".moai/reports/plan-audit/<file>" },
    { "kind": "verify_path", "value": "",     "ref": ".moai/state/verify/<session>/1-go-test.log" }
  ]
}
```

Field notes:

- `request_digest` — truncated SHA-256 of the raw request, computed in-process;
  the raw text is never persisted (REQ-HEV-005). `request_class` is a coarse
  keyword-derived enum (non-verbatim) included so the Loop 1 pattern key
  (`request_class + subcommand + mode + outcome`, design SSOT §5) is derivable;
  its v1 inclusion is tracked as an open clarification in plan.md.
- `model_class` — STOP-guard weighting input (design SSOT §2: GLM sessions are
  observation-only in later loops; observation itself is always accepted).
  `unknown` is the fallback when resolution fails.
- `evidence_refs[].kind` — closed enum {`gate_exit`, `audit_score`,
  `verify_path`} in v1 (REQ-HEV-006). `terminal: true` marks a signal eligible
  to close the row via `DeriveOutcome` (REQ-HEV-013).
- `delegations[].outcome` — derived from post-delegation machine evidence where
  available; `blocker` carries the structured blocker category when the
  delegation returned a blocker report (a structurally observable artifact).
- Nullability — every field whose machine signal may be absent is nullable and
  defaults to `null`/zero, never to an inferred value (REQ-HEV-006).

### D.2 Pending-row lifecycle

```
dispatch ──► ledger record ──► .moai/state/routing-pending-<session>.json  (open)
                 │  (same session, new record)          │
                 └──► earlier row finalized: reroute    │ ledger evidence (append machine refs)
                                                        ▼
Stop hook (harness-observe-stop, self-gated) ──► DeriveOutcome(evidence)
     ├─ terminal (abort|fail|success) ──► append routing-ledger.jsonl + delete pending
     └─ non-terminal ──► leave pending (multi-turn pipeline)
next-session ledger record ──► stale/foreign pending row ──► finalized: abort (lazy sweep)
```

The full machine-verifiable AC matrix (AC-HEV-001 … AC-HEV-026) lives in
`acceptance.md` (SSOT). Every REQ maps to at least one AC; cross-file
registrations (SKILL.md router, workflow bodies, hook handler, CLI verb,
template mirrors) are pinned as SEPARATE baseline-0 ACs per the
reachability discipline.

## §E. Exclusions

The following are explicitly out of scope for this SPEC.

### Out of Scope — Curator writes and Learned surfaces (EVOLVE-002)

- NO writes to CLAUDE.md, CLAUDE.local.md, or any `MOAI:LEARNED-WORKFLOW`
  managed block. The typed managed-block writer, the append-only Learned
  section writer, the 2-layer Recall contract, and snapshot/rollback/lineage
  extension are `SPEC-HARNESS-EVOLVE-002` territory.
- NO digest-layer emission: this SPEC produces the ledger (원장 layer) only;
  nothing is loaded into always-on context.

### Out of Scope — Gates, registries, and tier mechanics (EVOLVE-003)

- NO tier promotion / demotion, NO tier-count changes, NO harness.yaml v2
  schema, NO `auto_detection` surface registration.
- NO negative-evidence registry, NO re-proposal suppression / cooldown, NO L2
  Canary or L3 Contradiction activation, NO permission-surface Frozen-guard
  expansion (design deltas A1, A6, A7 land in `SPEC-HARNESS-EVOLVE-003`).

### Out of Scope — Console verbs and Recall wiring (EVOLVE-004 / EVOLVE-005)

- NO `/moai harness evolve | promote | demote | freeze | unfreeze` verbs and
  NO `status` / `doctor` extension (EVOLVE-004).
- NO Phase −1 Harness Recall wiring, NO Phase Ω routing-bias consumption, NO
  `harness-spec.yaml` typed parser, NO template deployment of managed-block
  markers (EVOLVE-005). This SPEC only PRODUCES the ledger rows that Recall
  will later search.

### Out of Scope — Loop 1 / Loop 2 aggregation

- NO pattern aggregation, NO learner.go extension, NO lessons-inbox drain
  changes, NO auto-memory writes. The Reflector consumes this ledger in later
  SPECs; this SPEC does not read its own ledger for learning purposes (the
  reader exists for filtering/inspection only).

### Out of Scope — New hook surfaces and SessionEnd extension

- NO new hook wrapper script, NO settings.json / settings.json.tmpl hook
  registration change, NO SessionEnd handler extension. Outcome capture rides
  the EXISTING `harness-observe-stop` registration; stale-row abort is handled
  by the lazy sweep in `ledger record` (REQ-HEV-014). If a SessionEnd-based
  sweep proves necessary, it is a follow-up SPEC.

### Out of Scope — usage-log observer changes

- `internal/harness/observer.go`, `.moai/harness/usage-log.jsonl`, retention,
  auto-classify, and auto-propose chains are untouched except for the additive
  routing-finalizer call inside the `harness-observe-stop` CLI handler.

### Out of Scope — CHANGELOG / README / docs-site

- CHANGELOG.md is owned by manager-docs (sync-phase); README and docs-site
  4-locale updates are a follow-up sync/docs concern.

## §F. Cross-References

- `.moai/reports/harness-self-evolving-redesign-final-20260712.html` — design
  SSOT (§3 gaps G1/G8, §4 3-Zone surface contract, §5 Loop 0 ledger schema +
  deltas A2/A4, §6 Phase Ω, §7 M1 milestone + risk grid).
- `internal/harness/observer.go` — the EXISTING usage-log observer this SPEC
  stays separate from (REQ-HEV-009).
- `internal/cli/hook.go` — `runHarnessObserveStop` handler (Stop-hook host for
  the additive routing finalizer, REQ-HEV-012) + `isHarnessLearningEnabled`
  gate (REQ-HEV-016).
- `.claude/hooks/moai/handle-harness-observe-stop.sh` + settings.json Stop
  registration — the existing transport reused verbatim (no changes).
- `.claude/skills/moai/SKILL.md` + `workflows/{plan,run,sync}.md` — recording
  obligation hosts (REQ-HEV-017/018) with template mirrors under
  `internal/template/templates/.claude/skills/moai/`.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 — the
  no-unobserved-claim invariant the machine-signal-only rule (REQ-HEV-006)
  mechanizes at the observation layer.
- CLAUDE.local.md §2 (Template-First) + §25 (Template Internal-Content
  Isolation) — mirror + neutrality discipline (REQ-HEV-019/020).
- `SPEC-HARNESS-EVOLVE-002..005` (unauthored) — Epic successors consuming this
  ledger. This SPEC has no `depends_on` (Epic entry point).
- `plan.md` / `acceptance.md` — implementation plan + AC matrix (SSOT).
