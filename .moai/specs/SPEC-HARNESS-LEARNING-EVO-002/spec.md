---
id: SPEC-HARNESS-LEARNING-EVO-002
title: "Delegation-map analyzer (L2) — proposals from observed routing-ledger rows"
version: "0.1.0"
status: completed
created: 2026-08-09
updated: 2026-08-10
author: manager-spec
priority: P3
phase: "v3.2 target"
module: "internal/harness/delegationmap"
lifecycle: spec-anchored
tags: "harness-learning, delegation-map, analyzer, proposalgen, meta-context-engineering, autonomy-epic"
tier: M
related_specs: [SPEC-HARNESS-LEARNING-EVO-001, SPEC-V3R6-HARNESS-PROPOSAL-GEN-001, SPEC-HARNESS-EVOLVE-001, SPEC-LSEL-LOCAL-EVOLUTION-001]
---

# SPEC-HARNESS-LEARNING-EVO-002 — Delegation-map analyzer (L2)

## HISTORY

- 2026-08-09 — Initial draft. Split from `SPEC-HARNESS-LEARNING-EVO-001` v0.1.0, which carried 33 requirements and 36 acceptance criteria — over the ceiling at Tier M (16/16) and Tier L (25/25). Per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier ("Exceeding either ceiling is a signal to tier up or to split the SPEC, not to relax the budget"), the SPEC was split along the L1/L2 line it already drew. This SPEC carries **L2 only** — the analyzer. Two requirements are shaped by measurements taken during the split: agent identity is a mixed population (§A.2) and two designated agents are conditionally invoked (§A.3).

## §A. Context

### A.1 — the consumer of a channel that is being repaired

`.moai/config/sections/delegation.yaml` declares a learning loop in its own header: `observe: routing-ledger`, `propose_via: harness-tier-ladder`, `auto_apply: false`. The observe channel is empty (4 rows, every one with `delegations: []`), and `SPEC-HARNESS-LEARNING-EVO-001` repairs it. This SPEC builds the missing consumer: read finalized rows, aggregate delegation patterns per subcommand, and emit proposals in the existing `proposalgen` directory shape whose content is a concrete delegation-map amendment.

The two SPECs are siblings, not a dependency chain. This one is testable end to end against synthetic fixtures generated from `routing.PendingRow.Finalize()` — a function that exists unchanged today — so it neither waits on nor gates its sibling. The finalized row shape is the contract between them, and `SPEC-HARNESS-LEARNING-EVO-001` REQ-HLE-014 freezes it at `schema_version: 1`.

### A.2 — the observed agent-identity population is mixed

Measured against the primary checkout on 2026-08-09: of 2,783 `subagent_stop` rows in `.moai/harness/usage-log.jsonl`, 1,941 (69.7%) carry a non-empty `agent_type`. The distribution leads with exactly the identities `delegation.yaml` designates — `manager-spec` 281, `manager-develop` 277, `plan-auditor` 201, `Explore` 129, `manager-docs` 122, `manager-git` 74, `sync-auditor` 31, `super-advisor` 7 — but three other populations are mixed in:

- **Absent** — 842 rows (30.3%) carry no identity at all.
- **Non-catalog agent types** — `general-purpose` 204, `workflow-subagent` 186, plus user-owned harness specialists (`hns-*-specialist`, `cli-template-specialist`) and the runtime built-in `claude-code-guide`. These are real agents; none is a member of the 12-agent retained catalog the delegation map designates from.
- **Spawn names** — a named spawn puts the NAME in `agent_type`. Directly observed: a spawn of `subagent_type: plan-auditor` with `name: audit-hle` produced `agent_type: "audit-hle"`. The long tail holds roughly 140 such values (`lens-*`, `rev-*`, `humanize-*`, `audit-*`).

The consequence for this SPEC is a mechanical discrimination rule: only a value that exactly matches a retained-catalog name is comparable against the delegation map. Everything else must be recorded and classified, never treated as an agent the map failed to designate.

### A.3 — two designated agents are conditionally invoked

`delegation.yaml` designates `sync-auditor` for `sync` and `plan-auditor` for `plan`. Both are legitimately absent from many runs:

- `sync-auditor` runs at harness level `thorough` only (`CLAUDE.md` §6; `.claude/rules/moai/workflow/spec-workflow.md` § Mode Dispatch auto-selection).
- `plan-auditor` is skipped by the skip-eligibility path in `.claude/rules/moai/workflow/spec-workflow.md` § Plan Audit Gate skip policy whenever the cached verdict is PASS at or above the tier threshold with an unchanged artifact hash.

A `designated_never_spawned` proposal reasons from absence, so without an explicit exclusion it would fire on both — reporting as a defect the exact behavior the rules prescribe.

## §B. User Story

As the maintainer of MoAI's harness, I want the delegation map revised from what the pipeline actually did — which agents were really spawned for which subcommand, and how those runs ended — rather than from my recollection, so that `delegation.yaml` stops drifting from reality and every proposed amendment arrives with an observation count behind it.

## §C. Scope boundary

- **In scope (L2).** Reading finalized routing-ledger rows, aggregating per subcommand, applying thresholds, resolving the designated agent set from `delegation.yaml` read-only, and emitting proposals through the existing `proposalgen` writer and directory layout. A CLI verb with a dry-run mode.
- **Out of scope (L1).** Creating rows, appending delegations, and closing rows on a terminal signal belong to `SPEC-HARNESS-LEARNING-EVO-001`.
- **Out of scope (L3).** Applying a proposal to `delegation.yaml`. See §G.

## §D. Requirements

Notation: GEARS. `REQ-HLA-0xx`. Budget: 16 requirements (Tier M ceiling 16).

### D.1 — Reading and aggregation

- **REQ-HLA-001** — The analyzer shall consume ledger rows through the existing `routing.Reader` and shall not open the ledger by an independently-declared path literal. It shall retain no per-row state beyond the running aggregate. **Where** the ledger file exceeds a declared maximum size, the analyzer shall decline to read it and return the empty result with a machine-readable reason rather than performing an unbounded read.
- **REQ-HLA-002** — The analyzer shall aggregate, per matched subcommand, the set of agents observed in `delegations[]` together with each agent's observation count and the count of rows for that subcommand. It shall treat one finalized row as exactly one subcommand observation, matching the first-writer-wins labelling its sibling SPEC records.
- **REQ-HLA-003** — The analyzer shall count toward its aggregation only rows whose outcome is `success` or `fail`, and shall report `reroute` and `abort` counts separately as routing-instability context that does not by itself produce a proposal.

### D.2 — Identity discrimination

- **REQ-HLA-004** — The analyzer shall compare an observed agent value against the delegation map only **when** that value exactly matches a member of the retained agent catalog; every other value shall be recorded and classified as non-catalog, and shall never produce an *undesignated-agent* proposal. A value that names a spawn, a non-catalog agent type, or a user-owned harness specialist is not evidence that the map omitted a designation.
- **REQ-HLA-005** — The analyzer shall count delegation entries carrying the absent-identity marker as unattributed observations, shall exclude them from every per-agent count, and shall report the unattributed share per subcommand so a reviewer can judge how much of the population the finding rests on.

### D.3 — Thresholds and proposal kinds

- **REQ-HLA-006** — **While** a subcommand has fewer finalized qualifying rows than the minimum-observation threshold, or **while** an observed agent's support ratio for that subcommand is below the confidence threshold, the analyzer shall emit no proposal for that subcommand or agent respectively, so that a single stray row cannot produce a proposal.
- **REQ-HLA-007** — The analyzer shall emit two proposal kinds and no others: an *undesignated-agent* proposal, where a retained-catalog agent clears both thresholds for a subcommand but is absent from that subcommand's designated agent list; and a *designated-never-spawned* proposal, where a designated agent is observed in zero qualifying rows for a subcommand that itself clears the minimum-observation threshold.
- **REQ-HLA-008** — The analyzer shall exclude conditionally-invoked designations from *designated-never-spawned* proposals, using a declared exclusion set in which each entry cites the rule that makes its invocation conditional. `sync-auditor` and `plan-auditor` shall be members of that set.

### D.4 — Emission and isolation

- **REQ-HLA-009** — The analyzer shall read `.moai/config/sections/delegation.yaml` to resolve the designated agent list, shall open it read-only, and shall never write it.
- **REQ-HLA-010** — The analyzer shall key each proposal with a `pattern_key` in the reserved namespace `delegation_map:<subcommand>:<discriminator>`, which shall not collide with the existing `tool_failure:` / `agent_invocation:` and other event-type-prefixed namespaces.
- **REQ-HLA-011** — The analyzer shall not extend `harness.PatternBearingEventTypes()` and shall not route its candidates through `proposalgen.MapPromotions`; it shall construct its own candidates and emit them through the existing proposal writer and directory-layout accessor.
- **REQ-HLA-012** — Each emitted proposal shall carry the observation count, the support ratio, the qualifying-row count for its subcommand, the unattributed share, and the proposal kind, so a reviewer can judge the amendment without re-reading the ledger.
- **REQ-HLA-013** — The analyzer shall be deterministic: two runs over the same input shall produce the same candidate set in the same order, and re-running over an unchanged input shall not churn already-written draft directories.
- **REQ-HLA-014** — **When** the ledger is absent, empty, entirely malformed, or over the size limit, the analyzer shall return an empty result with a machine-readable reason and shall exit successfully.
- **REQ-HLA-015** — A CLI verb shall expose the analyzer with a dry-run mode that performs no filesystem write and a structured JSON result on standard output.
- **REQ-HLA-016** — The analyzer shall not modify `.claude/lsel/frozen-allowlist.json` and shall not remove `^\.moai/config/sections/` from its frozen patterns; shall not invoke `AskUserQuestion`; and shall not emit a skill-level proposal, because the schema-v1 ledger records no skill-injection field.

## §E. Central assumption and its falsification

The design rests on one assumption: **ledger-observed delegation patterns carry enough signal to justify a delegation-map amendment.** It is not yet testable, because the data does not exist — which is why `SPEC-HARNESS-LEARNING-EVO-001` precedes it in practice even though it does not gate it.

The assumption is falsified by any of the following, each bounded so it can actually be discharged. The review window opens once the ledger holds at least 50 qualifying rows and closes after the first 10 proposals reach the Tier-4 gate.

- **F1 — no stable pattern.** For every subcommand clearing the minimum-observation threshold, no retained-catalog agent reaches the support-ratio threshold; the observed agent distribution is flat. The map cannot be evolved from a signal that does not concentrate.
- **F2 — the map is already right.** The analyzer emits zero proposals of either kind across the whole catalogue, meaning the hand-tuned map already matches observed behavior and the automation buys nothing.
- **F3 — proposals are systematically rejected.** Of the first 10 proposals surfaced at the Tier-4 gate within the review window, at least 8 are rejected by the reviewer — the signal is real but is not a valid basis for amendment.
- **F4 — the observation is unattributable.** The unattributed share (REQ-HLA-005) stays at or above the 30.3% measured in §A.2 across every subcommand clearing the threshold, so no per-agent count can carry a proposal.

Recording the falsifiers here is deliberate: F2 in particular is a *successful* outcome for the harness and a *negative* one for this SPEC, and the two must not be confused at review time.

## §F. Upstream producer

`SPEC-HARNESS-LEARNING-EVO-001` produces the rows this analyzer reads. The coupling is one-directional and non-blocking; that SPEC is listed under `related_specs` rather than `depends_on` deliberately, because a `depends_on` entry would gate this SPEC's run-phase entry on the sibling reaching `status: completed` (`.claude/rules/moai/workflow/spec-workflow.md` § Depends_on Pre-flight Check), which the fixture-based test strategy makes unnecessary.

## §G. Exclusions

### Out of Scope — instrumentation (L1)

- No hook seams, no store API changes, no changes to how rows are created, annotated, or finalized. All of it belongs to `SPEC-HARNESS-LEARNING-EVO-001`.

### Out of Scope — automated application to delegation.yaml (L3)

- No automated writer for `.moai/config/sections/delegation.yaml`. Three independent surfaces agree that this file requires human approval: `.claude/lsel/frozen-allowlist.json` lists `^\.moai/config/sections/` among its frozen patterns; the file's own `auto_apply: false` key; and the Tier-4 gate language of the roadmap report. This is a deliberate design boundary, not an oversight or an unfinished edge.
- No modification of the frozen allowlist to make the file writable.
- No new approval gate: the existing Tier-4 `AskUserQuestion` flow applies unchanged, and no component introduced here invokes it.

### Out of Scope — skill-level delegation proposals

- No proposal of the form "subcommand X should add skill Y". The schema-v1 ledger row has no skill field, so a skill amendment cannot be grounded in observation. Extending the row schema to record injected skills is a candidate follow-up, recorded in `plan.md` §H.

### Out of Scope — the usage-log / proposalgen tier ladder

- `internal/harness/proposalgen/mapper.go`, the tier-promotion path, and `usage-log.jsonl` semantics are consumed as they are. The analyzer reuses the proposal *writer* and *layout*, not the promotion mapper.

### Out of Scope — threshold tuning from real data

- `MinQualifyingRows` and `MinSupportRatio` ship as declared constants chosen by analogy to the existing `proposalgen` precedent. Re-tuning them against real observations is deferred until the §E review window has run.
