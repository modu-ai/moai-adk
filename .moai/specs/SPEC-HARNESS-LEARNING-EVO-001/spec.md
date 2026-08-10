---
id: SPEC-HARNESS-LEARNING-EVO-001
title: "Routing-ledger instrumentation repair (L1) — mechanical emission of delegation observations"
version: "0.2.0"
status: in-progress
created: 2026-08-09
updated: 2026-08-10
author: manager-spec
priority: P3
phase: "v3.2 target"
module: "internal/harness/routing"
lifecycle: spec-anchored
tags: "harness-learning, routing-ledger, instrumentation, mechanical-emission, hook-seams, meta-context-engineering, autonomy-epic"
tier: M
related_specs: [SPEC-HARNESS-LEARNING-EVO-002, SPEC-HARNESS-EVOLVE-001, SPEC-V3R6-HARNESS-PROPOSAL-GEN-001, SPEC-HARNESS-LOOP-REPAIR-001, SPEC-LSEL-LOCAL-EVOLUTION-001]
---

# SPEC-HARNESS-LEARNING-EVO-001 — Routing-ledger instrumentation repair (L1)

## HISTORY

- 2026-08-09 — Initial draft. P3 of the MoAI autonomy/workflow redesign roadmap (`.moai/reports/moai-autonomy-workflow-redesign-20260803.html` §5), applying §2.5 "Meta Context Engineering" (OpenReview P1jHroBS5E): the harness learning subsystem is MoAI's lower-level optimizer, evolving the delegation map from observed run outcomes instead of hand-tuned YAML.
- 2026-08-09 — v0.2.0. **Split**: the original single SPEC carried 33 requirements and 36 acceptance criteria, over the Tier ceiling at both M (16/16) and L (25/25). Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier ("Exceeding either ceiling is a signal to tier up or to split the SPEC, not to relax the budget"), the SPEC is split along the L1/L2 line it already drew in §C. This SPEC retains **L1 only** (instrumentation). The L2 analyzer moves to `SPEC-HARNESS-LEARNING-EVO-002`. Two requirements changed materially on new measurement: agent identity (§A.5) and the terminal-signal source (§A.6).

## §A. Context — the measured starting state

Six measurements establish the baseline. Every command is named in `plan.md` §A.1 so it can be re-run; measurements were taken against the **primary checkout** at `/Users/goos/MoAI/moai-adk-go` (the runtime state files below are gitignored and absent in the worktree where this SPEC is authored).

- **A.1 — the observation channel is dry.** `.moai/state/routing-ledger.jsonl` holds 4 rows. Every row has `delegations: []` (length 0); outcomes are `abort` (3) and `reroute` (1). Not one successfully-routed delegation has ever been recorded. The sibling generated-artifact observer log `.moai/harness/usage-log.jsonl` holds 109,236 rows over the same period. The asymmetry is specific to the routing ledger, not to the harness.
- **A.2 — the emission is instruction-dependent.** The only prescribed producer is the orchestrator invoking `moai harness ledger record` / `... evidence`, per `.claude/skills/moai/SKILL.md`. No Go code path emits either call. An instruction that must be remembered every dispatch is not emitted reliably, and the ledger's own contents are the evidence.
- **A.3 — nothing writes the target file.** `grep -rn 'delegation.yaml' internal/ --include='*.go' | grep -v _test` returns zero matches. The learning loop declared in `.moai/config/sections/delegation.yaml`'s own header (`observe: routing-ledger`, `propose_via: harness-tier-ladder`, `auto_apply: false`) has a dry observe channel.
- **A.4 — the proposal ladder itself works.** `internal/harness/proposalgen/` generates drafts; 73 proposal directories exist and `learning-history/tier-promotions.jsonl` carries 175 rows. Its input is `usage-log.jsonl` patterns, which carry no notion of delegation routing — which is why the consumer is a separate SPEC (`SPEC-HARNESS-LEARNING-EVO-002`) rather than an extension of the mapper.
- **A.5 — agent identity IS observable, with two caveats.** Of 2,783 `subagent_stop` rows in `usage-log.jsonl`, 1,941 (69.7%) carry a non-empty `agent_type`, and its values include exactly the identities `delegation.yaml` designates: `manager-spec` 281, `manager-develop` 277, `plan-auditor` 201, `Explore` 129, `manager-docs` 122, `manager-git` 74, `sync-auditor` 31, `super-advisor` 7. The derived `subject` field is `unknown` on all 2,783 rows and is the wrong field to read. **Caveat 1**: 842 rows (30.3%) carry no `agent_type` at all. **Caveat 2**: a *named* spawn puts the NAME in `agent_type`, not the agent type — directly observed in this session, where a spawn of `subagent_type: plan-auditor` with `name: audit-hle` produced `{"event_type":"subagent_stop","subject":"unknown","agent_type":"audit-hle","agent_id":"aaudit-hle-19009f7b26a36428"}`. The long tail confirms it: `general-purpose` 204, `workflow-subagent` 186, plus ~140 distinct one- and two-occurrence values that are spawn names (`lens-*`, `rev-*`, `humanize-*`, `audit-*`) or user-owned harness specialists (`hns-*-specialist`, `cli-template-specialist`).
- **A.6 — the terminal signal population is empty, and the root cause is hook registration.** Across every file in `.moai/evolution/telemetry/`, the count of records with `is_test_pass` or `is_test_fail` set is **0**. The cause is not classifier failure, decode failure, or a gate: `buildEvidenceRecord` (`internal/hook/evidence_writer.go:290`) is reached only from `logEvidence` (`internal/hook/post_tool.go:224`), which runs in the `moai hook post-tool` handler behind `handle-post-tool.sh` — and that wrapper is registered for `matcher: "Write|Edit|MultiEdit"` only. `buildBashRecord` (line 309) is therefore unreachable from the shipped hook wiring; execution never arrives. Confirmed by direct experiment: running `go test ./internal/harness/routing/...` left the day's telemetry file unchanged at 4,025 bytes, and every one of its records carries `tool_name: null` with `path_kind: docs-only` (file-edit records only).

The blocker is therefore **data starvation, not a missing algorithm**, and it has two distinct causes: routing rows are never created mechanically (A.2), and the terminal signal that would close them is never produced (A.6).

## §B. User Story

As the maintainer of MoAI's harness, I want the routing ledger to accumulate rows that record which agents were actually spawned for which subcommand and how the run ended — derived mechanically from hook input rather than from an instruction the orchestrator must remember — so that a downstream analyzer has real observations to reason over instead of an empty file.

## §C. Scope boundary

- **In scope (L1 — instrumentation repair).** Make routing-ledger rows accumulate with useful content: a pending row created per routed session, `delegations[]` carrying real entries keyed on the observed agent identity, a terminal machine signal closing rows as `success`/`fail` rather than only `reroute`/`abort`, and the Stop-hook finalizer verified end to end.
- **Out of scope (L2 — the analyzer).** Reading finalized rows, aggregating delegation patterns, and emitting delegation-map amendment proposals belong to `SPEC-HARNESS-LEARNING-EVO-002`. That SPEC consumes the rows this one produces, but is testable against fixtures without this SPEC being complete, so it is a sibling and not a dependency.
- **Out of scope (L3 — application).** See §G.

## §D. Requirements

Notation: GEARS. `REQ-HLE-0xx`. Budget: 16 requirements (Tier M ceiling 16).

### D.1 — Emission mechanics

- **REQ-HLE-001** — The routing observation subsystem shall derive every ledger field it can from hook input rather than from an orchestrator instruction. An instruction-dependent emission path shall be treated as a fallback, never as the primary producer.
- **REQ-HLE-002** — **When** a user prompt is submitted and both harness observation gates are open, the UserPromptSubmit hook shall ensure a pending routing row exists for the session, deriving `request_digest` and `request_class` from the prompt text through the existing `routing.RequestDigest` / `routing.ClassifyRequest` functions, and persisting no verbatim prompt text.
- **REQ-HLE-003** — **While** a pending routing row already exists for the session, the UserPromptSubmit hook shall leave that row in place and shall not finalize it as `reroute`. A multi-turn pipeline spans many prompts, and a per-prompt reroute would close a row that has not finished being observed.
- **REQ-HLE-004** — The routing store shall expose a create-if-absent recording operation distinct from the existing reroute-on-record operation, and that operation shall inherit both existing safety guards of the stale sweep unchanged: the 24-hour age guard (`routing.StalenessThreshold`) and the `active-sessions.json` liveness guard (`isSessionLive`). Together these are what protect a concurrent session's in-flight pending row from a false abort; pending files are per-session so they never contend, and the ledger is opened O_APPEND-only.
- **REQ-HLE-005** — The routing store shall expose an annotation operation that patches routing metadata (`matched_subcommand`, `mode_selected`, `tier`, `harness_level`, `clarify_rounds`, `model_class`) onto an existing pending row without creating a row and without finalizing one; a supplied empty value shall leave the existing field unchanged. The operation shall also be reachable as a CLI verb alongside the existing `record` / `evidence` verbs.
- **REQ-HLE-006** — **When** a submitted prompt begins with a literal `/moai <subcommand>` form, the UserPromptSubmit hook shall set `matched_subcommand` from that literal token **only if the field is currently empty** (first-writer-wins); a prompt carrying a different literal subcommand on a row already labelled shall leave the existing label unchanged. Otherwise the hook shall leave `matched_subcommand` empty for later annotation, and shall not guess a subcommand from natural-language text.

### D.2 — Delegation and outcome observation

- **REQ-HLE-007** — **When** a subagent stops and both harness observation gates are open, the SubagentStop hook shall append one `Delegation` entry to the session's pending routing row whose `agent` value is taken from the hook input's `agent_type` field verbatim, without normalization, mapping, or substitution of the derived `subject` field.
- **REQ-HLE-008** — **When** a stopping subagent supplies no `agent_type` value, the SubagentStop hook shall still append the delegation entry, recording the absent identity under an explicit distinguishable marker that is not the empty string, so that an unattributed delegation is countable and never silently indistinguishable from an attributed one.
- **REQ-HLE-009** — **When** no observable outcome signal is available for a stopping subagent, the SubagentStop hook shall record the delegation with an explicit unknown-outcome marker and a null blocker rather than inferring success.
- **REQ-HLE-010** — **When** the Stop hook observes a terminal machine signal already classified by the existing session evidence path (an observed test pass or test failure for the session), it shall append a corresponding terminal `gate_exit` evidence reference to the pending routing row before running the existing finalizer, so that rows can close as `success` or `fail`. Absence of a signal shall append nothing; absence is not failure.
- **REQ-HLE-011** — The system shall restore production reachability of the Bash evidence-record path (`buildBashRecord`), which §A.6 measured as unreachable from the shipped hook wiring, without which REQ-HLE-010 observes an empty signal population. The restoration shall be scoped so that no evidence record is written twice for a single tool call.

### D.3 — Safety, budget, and boundaries

- **REQ-HLE-012** — Every L1 emission path shall be fail-open: an error in recording, annotating, or appending shall be written to the hook's diagnostic sink and swallowed, and shall never block prompt submission, subagent shutdown, or session end.
- **REQ-HLE-013** — **While** either harness observation gate is closed (the fail-closed hook opt-in, or the fail-open learning switch), no L1 emission path shall write to the ledger, to a pending row, or to any other file.
- **REQ-HLE-014** — The routing ledger shall not gain a schema-breaking field in this SPEC; `schema_version` shall remain 1 and every existing consumer shall continue to parse existing rows unchanged.
- **REQ-HLE-015** — Each L1 seam shall keep its total filesystem read count bounded and declared, covering every read on the seam path — the pending-row read and write, the stale sweep's directory scan and its per-foreign-file reads, and the session evidence path's two whole-file telemetry reads — so that the touched hook stays within the 5-second MoAI hook budget. The create-if-absent operation shall not run the stale sweep on every invocation; the sweep shall remain bounded to the dispatch path so that per-prompt emission does not convert a once-per-dispatch directory scan into a once-per-prompt one.
- **REQ-HLE-016** — No raw user prompt text shall be persisted by any path introduced here; only the existing digest and coarse class shall be stored.

## §E. Central assumption and its falsification

The design rests on one assumption: **`delegations[]` measures delegation-map adherence.** The two caveats in §A.5 are the sharp edges.

The assumption is falsified by either of the following, observed after the ledger has accumulated at least 50 finalized rows:

- **F1 — the observation is an artifact.** The observed `agent_type` set is dominated by values that are not retained-catalog agent names (spawn names, `general-purpose`, `workflow-subagent`, user-owned harness specialists), so `delegations[]` measures spawn labelling rather than delegation-map adherence. Bound: fewer than half of the recorded delegation entries carry a retained-catalog name.
- **F2 — attribution is too sparse.** The share of delegation entries recorded under the absent-identity marker (REQ-HLE-008) stays at or above the 30.3% measured in §A.5, meaning nearly a third of every subcommand's observation is unattributable and no per-agent count can be trusted.

Recording these here is deliberate: both are outcomes this instrumentation can produce while working exactly as specified, and neither is a bug in the code.

## §F. Downstream consumer

`SPEC-HARNESS-LEARNING-EVO-002` reads the rows this SPEC produces and emits delegation-map amendment proposals through the existing Tier-4 approval gate. The coupling is one-directional and non-blocking: 002 is testable against synthetic fixtures generated from `routing.PendingRow.Finalize()`, so it neither waits on this SPEC nor gates it. The row shape is the contract between them, and REQ-HLE-014 freezes it.

## §G. Exclusions

### Out of Scope — the delegation analyzer (L2)

- No aggregation of finalized rows, no threshold logic, no proposal emission, and no read of `.moai/config/sections/delegation.yaml`. All of it belongs to `SPEC-HARNESS-LEARNING-EVO-002`.

### Out of Scope — automated application to delegation.yaml (L3)

- No automated writer for `.moai/config/sections/delegation.yaml`. Three independent surfaces agree that this file requires human approval: `.claude/lsel/frozen-allowlist.json` lists `^\.moai/config/sections/` among its frozen patterns; the file's own `auto_apply: false` key; and the Tier-4 gate language of the roadmap report. This is a deliberate design boundary, not an oversight.
- No modification of `.claude/lsel/frozen-allowlist.json` to make the file writable.

### Out of Scope — ledger schema migration

- No `schema_version: 2`. Every change here is additive within the existing row shape. Extending the row to record injected skills — which would unlock skill-level proposals — is recorded as a follow-up in `plan.md` §H.

### Out of Scope — backfill of historical data

- The four existing rows are left untouched. They predate the repaired instrumentation and carry no delegation content to recover.

### Out of Scope — per-subcommand row splitting

- A session that runs `/moai plan` then `/moai run` produces one row labelled by the first subcommand (REQ-HLE-006). Splitting such a session into one row per subcommand is a schema and lifecycle change deferred to a follow-up; the mislabel consequence is recorded in `acceptance.md` §F.
