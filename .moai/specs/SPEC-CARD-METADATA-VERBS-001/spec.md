---
id: SPEC-CARD-METADATA-VERBS-001
title: "Card-metadata editing verbs — set-spec and reclassify with an actor+reason audit trail"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/factory, internal/homestate, internal/config"
lifecycle: spec-anchored
tags: "card-metadata, set-spec, reclassify, audit-trail, factory, todo-queue, card-t1840"
tier: M
card: t1840
---

# SPEC: Card-Metadata Editing Verbs

Stage E of the card-policy redesign (design authority: `.moai/reports/card-policy-redesign-20261017/design.md` §2 principle 10, §3 "Metadata" row, §6 stage E — machine-local, read-only; cited as design intent, never as committed-tree authority). Card-metadata editing verbs for the factory/todo card record: `set-spec` (set a card's spec_id after pick), `reclassify` (re-designate mode and priority after issue), and the actor+reason audit trail every metadata mutation records.

## §A Problem

Today a card's `spec_id` is settable only at pick time (`todo pick --spec` / `factory next --spec`). After pick there is no setter path: the only route is unpick + re-pick with `--spec`, which has caused at least 9 SpecID-blank incidents; t1631 was handled by a direct SQL UPDATE stopgap because no verb existed. The record-layer guard actively refuses the change — `RecordPicked` accepts field changes only while the card is still `picked` (`internal/homestate/card_picked.go:184-185`), and the queue-side consumer `cardToSpecID` (`internal/cli/contract_report.go:44-60`) reads whatever the queue item carries, so a blank stays blank for every downstream consumer.

A card's classification (mode, priority) is likewise a creation-time-only judgment: `CardClassification` is "written only through the add path's decider seam, inside the locked write that appends the card" (`internal/factory/classification.go:83-84`). No verb re-designates it after issue, even though selection ordering depends on it and operator judgment legitimately changes.

Neither surface records an audit trail of who changed what and why. Design principle 10 ("Metadata is editable with an audit trail") requires explicit verbs with actor and reason.

## §B Scope

In scope:

- Two CLI verbs on the card record: `set-spec` (spec_id, after pick) and `reclassify` (mode + priority, after issue).
- Validation: SPEC-ID grammar for `set-spec`; closed-set values for `reclassify`; non-empty bounded reason for both.
- An audit surface recording actor, timestamp, field, old value, new value, and reason for every successful metadata mutation, reusing the shape of the existing homestate `events` mechanism (`internal/homestate/factory.go:79-84`, `appendEvent` at `internal/homestate/card_transition.go:901`).
- A config flag, default OFF, gating both verbs.
- Actor identity capture following the existing patterns: lane label via `MOAI_FACTORY_WORKER` (`internal/cli/factory_audit_decide.go:30-34`), operator default, session UUID via the existing capture (`integrationSessionID`, `internal/cli/integration.go:138-150`).

Two stores are in play and both are write targets for `set-spec`:

- The queue item (`BacklogItem.SpecID`, `internal/factory/backlog_store.go:89`) — the store `cardToSpecID` reads.
- The factory card record (`cards.spec_id`, written today through `RecordPicked`/`CardFields`, `internal/homestate/card_picked.go:63-71,114`).

## §C Requirements (GEARS)

### C.1 Capability gate

- REQ-CMV-001: **Where** `workflow.card_metadata_verbs.enabled` is not true, the `set-spec` and `reclassify` verbs shall refuse with a refusal naming the flag key, writing neither store and no audit row.
- REQ-CMV-002: The distributed configuration default for `workflow.card_metadata_verbs.enabled` shall be false (flag default OFF), following the existing `workflow.<feature>.enabled` key conventions in `.moai/config/sections/workflow.yaml`.

### C.2 set-spec — admission and validation

- REQ-CMV-003: **When** `set-spec` names a card whose queue item does not exist or is not in the `picked` state, the verb shall refuse through the same admission family `factory assign` uses (`requireQueuePicked`, `internal/cli/factory_card.go:2792-2806`), writing neither store.
- REQ-CMV-004: **When** `set-spec` receives a spec_id that does not match the SPEC-ID grammar — literal `SPEC` prefix, one or more letter-leading uppercase-alphanumeric middle segments, and a digit-only three-digit tail — the verb shall refuse with a refusal quoting the rejected value, writing neither store.
- REQ-CMV-005: **When** `set-spec` receives a spec_id equal to the value the queue item already carries, the verb shall refuse as a no-change mutation, writing neither store nor audit row.

### C.3 set-spec — write semantics

- REQ-CMV-006: **When** `set-spec` succeeds, the verb shall write the new spec_id to BOTH the queue item's `spec_id` and the factory card record's spec_id column, sequenced inside the queue lock so a factory-side failure rolls the queue write back (the cross-store pattern of `factory assign`, `internal/cli/factory_card.go` — `queueStore.WithLock` wrapping the record write).
- REQ-CMV-007: The `set-spec` factory-record write shall land at every card state from `picked` onward — including every post-picked state the transition machine defines — superseding the `RecordPicked` while-picked-only field rule for this verb; the write shall bump the card row's version and refuse on a stale expected version (`ErrStaleVersion`).
- REQ-CMV-008: **When** `set-spec` succeeds, the queue-side consumer path (`cardToSpecID`, `internal/cli/contract_report.go:44`) shall resolve the new spec_id on its next read.

### C.4 reclassify — admission and validation

- REQ-CMV-009: **When** `reclassify` names a card whose queue item is in a state other than `queued` or `picked`, the verb shall refuse, enumerating admitted states positively (no negated default that a future state could fall through).
- REQ-CMV-010: The `reclassify` verb shall refuse any mode outside the closed set {serial, parallelizable} and any priority outside the closed set {low, normal, high}, validating through the existing closed-set validator (`ValidateCardClassification`, `internal/factory/classification.go:109-124`) rather than a re-spelled copy.
- REQ-CMV-011: **When** `reclassify` is invoked with neither `--mode` nor `--priority`, the verb shall refuse with a usage error, writing nothing.
- REQ-CMV-012: **When** `reclassify` supplies one axis, the verb shall preserve the other axis's current value and the `blocked` flag unchanged, deriving any absent classification through the read-side default seam (`EffectiveCardClassification`, `internal/factory/classification.go`) — the reclassify write shall never alter the `blocked` axis.

### C.5 reclassify — write semantics

- REQ-CMV-013: **When** `reclassify` succeeds, the verb shall write the supplied axes onto the queue item's classification record with decider `human`, the caller-supplied reason, and a fresh classified-at stamp, preserving the additive nullable-field discipline (absence marshals byte-identically; a card classified for the first time by reclassify acquires the field, never a placeholder value).

### C.6 Audit trail

- REQ-CMV-014: Every successful `set-spec` or `reclassify` mutation shall append exactly one audit row carrying actor, timestamp, field, old value, new value, and the caller-supplied reason, in a card-keyed metadata event surface in the homestate database that reuses the `events` table's shape (autoincrement seq, kind, JSON payload, created-at) per `internal/homestate/factory.go:79-84`.
- REQ-CMV-015: The audit row's actor shall be the lane label read from the lane marker environment variable when set, otherwise `operator` (the capture pattern of `factory decide`, `internal/cli/factory_audit_decide.go:30-34`), and the row shall additionally carry the session UUID resolved through the existing session-identity capture when one is resolvable.
- REQ-CMV-016: The verbs shall refuse a reason that is empty, whitespace-only, or longer than 500 runes, writing neither store nor audit row.
- REQ-CMV-017: **When** a mutation is refused at any gate before its write, the verb shall write no audit row and leave both stores byte-identical (refusal atomicity).

### C.7 Migration discipline

- REQ-CMV-018: Adding the audit surface shall be row-preserving — additive `CREATE TABLE IF NOT EXISTS` only — and shall rewrite, migrate, or delete no existing row of any existing table.

### C.8 Actor boundary

- REQ-CMV-019: **When** either verb is invoked from a lane session (the lane marker environment variable set), the verb shall refuse with the lane-boundary sentinel of the existing lane-mutation boundary (`factoryLaneRefusal`, `internal/cli/factory_card.go:120-124`); the metadata verbs are leader/operator acts. (The t1715 lane-reuse carve-out — a lane filling a blank spec_id on its own card — is an open decision: `decision-index.md` Q2; this requirement encodes the no-carve-out default.)

## §D Constraints

- RED tests first: every requirement lands as a failing test before its implementation (TDD mode per `.moai/config/sections/quality.yaml` `constitution.development_mode: tdd`).
- Config flag default OFF: `workflow.card_metadata_verbs.enabled`, distributed default false (design doc §6 "Each stage: RED tests first, config flag default off, row-preserving migration").
- Row-preserving migration: no destructive schema change (design doc §6).
- The queue store is the home SQLite database (the `<home>/db/<key>` layout), never the in-tree `backlog.db`; factory-database resolution for a queue store follows the sibling-resolution seam (`gtdFactoryDBForStore`, `internal/factory/gtd_operation.go:577+`, the site flagged P2 in a prior review).
- Reason bound: 500 runes maximum, named constant in the implementation.
- SPEC-ID grammar for `set-spec`: `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (letter-leading uppercase middle segments, digit-only three-digit tail — the card delegation's named rule). Note: pick-time acceptance remains the looser historical pattern (`specIDPattern`, `internal/homestate/card_picked.go:15`); tightening pick-time is NOT part of this SPEC (see §F).
- English throughout (project instruction-doc language policy).

## §E Coordination

- **t1641** (picked, in queue): its addendum ("동사 부세 — pick 이후 spec_id 기록 불가" / "verb lacking — spec_id unrecordable after pick") named this exact gap. The implementation lives HERE. This SPEC absorbs that addendum: t1641 does not carry it, and t1641's remaining scope is untouched by this SPEC.
- **t1715** (queued): its lane self-recovery item (4) — "link a SPEC id when the record has none" — REUSES the `set-spec` verb this SPEC defines; t1715 must not re-implement it. t1715 keeps its own recovery verbs (re-acquire, re-stamp, adopt-tree). The lane-affordance question (may a lane invoke `set-spec` on its own card) is decision-index Q2; t1715's card work consumes whichever resolution lands.
- Stage mapping (design doc §6): this SPEC is stage E. Stage A (liveness/ownership), stage C (eligibility-filtered selection), and stage D (resource-keyed locks) are separate cards and are excluded below.

## §F Out of Scope

### Out of Scope — Evidence-SHA carry-over

- Evidence-SHA carry-over between commits (the third metadata binding of design principle 10). A later stage with its own card; owner paths/dependency/toolchain invariance rules must be decided first.

### Out of Scope — Adjacent redesign stages

- Stage A liveness/ownership (process-identity probes, epoch fencing) — separate card.
- Stage C eligibility-filtered selection — separate card.
- Stage D resource-keyed locks — separate card (t1835 rescoped).
- Stage B durable identity migration; the card-keyed audit surface designed here must not preclude it, but implementing it is excluded.

### Out of Scope — Lane recovery verbs

- Re-acquire, re-stamp, and adopt-tree lane recovery verbs — t1715's scope. This SPEC defines only the metadata verbs t1715 reuses.

### Out of Scope — Pick-time grammar tightening

- Tightening the pick-time `--spec` acceptance (the looser `specIDPattern` at `internal/homestate/card_picked.go:15`) to the stricter grammar `set-spec` enforces. `set-spec` validates its own input; existing picks are not retroactively re-validated.

### Out of Scope — blocked axis and non-metadata classification fields

- The `blocked` classification axis, and the decider-seam judgment at add time. `reclassify` covers mode and priority only, per the delegation's named axes.

### Out of Scope — direct SQL access

- Any documented or blessed path of direct SQL UPDATE against the queue or factory databases as a metadata-editing substitute (the t1631 stopgap pattern this SPEC replaces).

## §G References

- Design authority (machine-local, read-only, primary tree): `/Users/goos/MoAI/moai-adk-go/.moai/reports/card-policy-redesign-20261017/design.md` §2 principle 10, §3 Metadata row, §6 stage E, §7.
- `internal/cli/factory_card.go` — `factory assign` `--spec` path (2810-2995), `requireQueuePicked` (2792), `factoryLaneRefusal` (120).
- `internal/homestate/card_picked.go` — `CardFields` (63-71), spec-ID pattern (15), while-picked-only refusal (184-185), `card.fields` event (197).
- `internal/homestate/factory.go` — `events` table shape (79-84).
- `internal/homestate/card_transition.go` — `TransitionRequest` (34+), `appendEvent` (901), version fencing.
- `internal/factory/backlog_store.go` — `BacklogItem.SpecID` (89), `clearBacklogLease` (1175-1181), `WithLock`/`Mutate` (921-1000).
- `internal/factory/classification.go` — closed sets (33-68), `ValidateCardClassification` (109), `EffectiveCardClassification`.
- `internal/factory/gtd_operation.go` — `gtdFactoryDBForStore` sibling resolution (577+).
- `internal/cli/contract_report.go` — `cardToSpecID` (43-60).
- `internal/cli/factory_audit_decide.go` — actor capture pattern (30-34).
- `internal/cli/integration.go` — session-identity capture (138-150).
- `internal/cli/todo.go` — pick/unpick spec write sites (1671, 1749), `recordFactoryCardState` (1765-1778).
- Status Transition Ownership Matrix: plan-phase artifacts authored by manager-spec (this agent); `draft → in-progress` owned by manager-develop; `in-progress → implemented → completed` owned by manager-docs (`.claude/rules/moai/development/spec-frontmatter-schema.md`).

## HISTORY

- 2026-10-17 v0.1.0 — Initial draft (manager-spec, card t1840 plan phase, stage E of the card-policy redesign). Absorbs the t1641 addendum naming the post-pick spec_id gap.
