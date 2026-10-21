---
id: SPEC-CARD-METADATA-VERBS-001
title: "Card-metadata editing verbs — set-spec and reclassify with an actor+reason audit trail"
version: "0.4.0"
status: in-progress
created: 2026-10-17
updated: 2026-10-22
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
- An audit surface recording actor, timestamp, field, old value, new value, and reason for every successful metadata mutation: a NEW card-keyed `card_metadata_events` table in the homestate database (decision-index Q3, Default applied 2026-10-17) reusing the shape of the existing `events` mechanism (`internal/homestate/factory.go:79-84`, `appendEvent` at `internal/homestate/card_transition.go:901`).
- A config flag, default OFF, gating both verbs.
- Actor identity capture following the existing patterns: lane label via `MOAI_FACTORY_WORKER` (`internal/cli/factory_audit_decide.go:30-34`), operator default, session UUID via the existing capture (`integrationSessionID`, `internal/cli/integration.go:138-150`).

Two stores are in play and both are write targets for `set-spec`:

- The queue item (`BacklogItem.SpecID`, `internal/factory/backlog_store.go:89`) — the store `cardToSpecID` reads.
- The factory card record (`cards.spec_id`, written today through `RecordPicked`/`CardFields`, `internal/homestate/card_picked.go:63-71,114`).

## §C Requirements (GEARS)

### C.1 Capability gate

- REQ-CMV-001: **Where** `workflow.card_metadata_verbs.enabled` is not true, the `set-spec` and `reclassify` verbs shall refuse with a refusal naming the flag key, writing neither store and no audit row.
- REQ-CMV-002: The distributed configuration default for `workflow.card_metadata_verbs.enabled` shall be false (flag default OFF), following the existing `workflow.<feature>.enabled` key conventions in `.moai/config/sections/workflow.yaml`.

### C.2 set-spec — validation gate

- REQ-CMV-003: **When** `set-spec` names a card or spec_id that fails any gate of its admission family, the verb shall refuse, writing neither store and no audit row.
    - Gate (a) — admission state: the queue item does not exist or is not in the `picked` state (the `requireQueuePicked` family, `internal/cli/factory_card.go:2792-2806`).
    - Gate (b) — no factory card record: the card is queue-picked but never assigned; the `cards` table is keyed `(run_id, card_id)` (`internal/homestate/factory.go:70-78`), and this SPEC creates no card row and invents no run identity for a card that has none — the operator assigns first, then sets the spec.
    - Gate (c) — grammar: the spec_id does not match the SPEC-ID grammar (literal `SPEC` prefix, one or more letter-leading uppercase-alphanumeric middle segments, digit-only three-digit tail); the refusal quotes the rejected value.
    - Gate (d) — no-change: the spec_id equals the value the queue item already carries AND the field's LATEST audit row carries that value as its new value (a converged-and-audited duplicate — the LATEST row is the most recent row the audit surface records for that card and field; a duplicate requires the LATEST row's new value to equal the commanded value, and the existence of an OLDER row whose new value equals it does not) — the convergence comparison reads the QUEUE item (the store `cardToSpecID` reads), never the factory record alone, so a direction-5 divergence is admitted for convergence rather than refused; a converged-but-unaudited item — including the return-to-prior-value corner, where an audited X→A then a later audited A→B precede a B→A crash in the direction-6 window, so an older row's new value equals the commanded A while the latest row's does not — is NOT refused here — it is admitted as the REQ-CMV-004 direction-6 converged-state audit completion.

### C.3 set-spec — write semantics

- REQ-CMV-004: **When** `set-spec` succeeds, the verb shall write the new spec_id to BOTH the queue item's `spec_id` and the factory card record's spec_id column — factory-record write first, queue write second, both inside the queue lock, and the audit row appended only after both stores have committed — under the cross-store failure matrix below. `WithLock` is a lock, not a transaction (`internal/factory/backlog_store.go:938-952`); the sequencing follows the `factory assign` cross-store pattern (`internal/cli/factory_card.go:2919-2923`), and this requirement states the rule for every direction in which the two stores — or the stores and the audit surface — can diverge.
    - Direction 1 — factory-side failure (the record write fails): nothing commits, the queue write is never attempted, the verb refuses; both stores unchanged, no audit row.
    - Direction 2 — queue-side failure after the record write committed: the verb shall run a compensating homestate write restoring the prior spec_id through the same version-fenced mutation function and exit non-zero with a divergence sentinel; no audit row is emitted (the append never ran).
    - Direction 3 — the compensating write itself fails: the verb shall exit non-zero with the divergence sentinel naming the card, the prior value, and the attempted value; the record-new/queue-old divergence stands, surfaced for operator repair — this SPEC builds and claims no cross-database transactionality.
    - Direction 4 — audit-append failure after both stores committed: the verb shall retry the append once, then exit non-zero with an audit-loss sentinel naming the card and the mutation, leaving the completed store writes in place — a surfaced missing row, never a silent success.
    - Direction 5 — crash between the two commits: the divergence is healed by re-running the verb — the no-change gate (REQ-CMV-003 gate (d)) reads the queue item, never the factory record alone, so the re-run is admitted and converges both stores on the commanded value; the heal's audit row takes its old value from the QUEUE item's pre-heal value — the value the unconverged store still carried — and its new value from the commanded one; the queue-read pin is load-bearing here and in direction 6 — a gate comparing the factory record instead would misread this state as converged and leave the queue divergence unhealed.
    - Direction 6 — crash after both commits and before the audit append: both stores carry the commanded value, no audit row exists, and no sentinel was emitted (the process died); the re-run shall be admitted by gate (d) as a converged-state audit completion that emits exactly the row the crashed run owed — the re-run's actor and caller-supplied reason, the old value taken from the field's LATEST audit row where one exists and empty otherwise — while re-mutating neither store (both stores row-unchanged), exiting zero. The catch-up row is an attestation: its old value is the prior-row value where the audit surface records one — the LATEST row's new value, which in the return-to-prior-value corner IS the true prior — and empty where no row exists, so a value that arrived at pick time and was never audited is honestly recorded as unknown-old, never fabricated. The catch-up is an admission, not a gate refusal, so REQ-CMV-013's before-write refusal scope is not engaged; REQ-CMV-010's exactly-one counts the value transition once, so a crash-plus-catch-up pair leaves exactly one row, never two; and the catch-up is the recovery route for a direction-4 audit-loss sentinel once the operator re-runs. The gate cannot distinguish this crash state from a value that arrived at pick time with no audit row — both are admitted and both land an attestation row recording the commanded value and reason; no store transition is fabricated, because neither store is re-mutated.
- REQ-CMV-005: The `set-spec` record-layer mutation function (the homestate function the verb calls — not the CLI verb itself) shall land the spec_id write at every card state from `picked` onward — every state `CardStates()` enumerates (`internal/homestate/card_record.go`: 21 states, `picked` plus the 20 post-picked states; `terminalStates = [done, failed, abandoned]`) — superseding the `RecordPicked` while-picked-only field rule for this mutation function alone; the write shall bump the card row's version and refuse on a stale expected version (`ErrStaleVersion`). Verb-level admission stays queue-picked-only per REQ-CMV-003: `todo done` archives the queue item out of the queue (`internal/cli/todo.go:1319-1320`), so terminal-state rows are reachable through the record-layer function only, never through the verb.
- REQ-CMV-006: **When** `set-spec` succeeds, the queue-side consumer path (`cardToSpecID`, `internal/cli/contract_report.go:44`) shall resolve the new spec_id on its next read.

### C.4 reclassify — validation gate

- REQ-CMV-007: **When** `reclassify` is invoked on a card that fails any gate of its admission family, the verb shall refuse, writing nothing, enumerating admitted states positively (no negated default that a future state could fall through).
    - Gate (a) — admission state: the queue item is in a state other than `queued` or `picked`, or is missing.
    - Gate (b) — closed sets: a mode outside {serial, parallelizable} or a priority outside {low, normal, high}, validated through the existing closed-set validator (`ValidateCardClassification`, `internal/factory/classification.go:109-124`) rather than a re-spelled copy.
    - Gate (c) — no axis: neither `--mode` nor `--priority` supplied (usage error).

### C.5 reclassify — write semantics

- REQ-CMV-008: **When** `reclassify` supplies one axis, the verb shall preserve the other axis's current value and the `blocked` flag unchanged, deriving any absent classification through the read-side default seam (`EffectiveCardClassification`, `internal/factory/classification.go`) — the reclassify write shall never alter the `blocked` axis.
- REQ-CMV-009: **When** `reclassify` succeeds, the verb shall write the supplied axes onto the queue item's classification record with decider `human`, the caller-supplied reason, and a fresh classified-at stamp, preserving the additive nullable-field discipline (absence marshals byte-identically; a card classified for the first time by reclassify acquires the field, never a placeholder value). The REQ-CMV-004 direction-6 catch-up is an admitted converged-state audit completion, not a fresh write — it is outside this requirement's fresh-write scope: the catch-up re-mutates no store (REQ-CMV-010 — the queue item row-unchanged, its decider, reason, and classified-at stamp exactly as the crashed run left them), so this requirement is not read to demand a fresh write (a new decider, reason, or classified-at stamp) where the direction-6 catch-up attests without re-mutating.

### C.6 Audit trail

- REQ-CMV-010: Every successful `set-spec` or `reclassify` mutation shall append exactly one audit row carrying actor, timestamp, field, old value, new value, and the caller-supplied reason, to a NEW card-keyed `card_metadata_events` table in the homestate database (decision-index Q3, Default applied 2026-10-17) that reuses the `events` table's shape — autoincrement seq, kind, JSON payload, created-at (`internal/homestate/factory.go:79-84`) — keyed by the card id the data actually has. The exactly-one invariant counts completed value transitions: where a run crashed after both stores committed and before the append (REQ-CMV-004 direction 6), the admitted re-run's catch-up append IS that row — a crash-plus-catch-up pair leaves exactly one row, never two. The same invariant and the same converged-state audit completion shape bind reclassify's single-store window, keyed to the queue item's classification field: where a run crashed after the queue classification write committed and before the append, a re-run commanding the value the queue item already carries is admitted when the field's LATEST audit row does not record that value as its new value (the same latest-row duplicate test as REQ-CMV-003 gate (d)) — it emits exactly the row the crashed run owed (the re-run's actor and reason, the old value taken from the field's LATEST audit row where one exists and empty otherwise), re-mutates no store (the queue item row-unchanged, its classified-at stamp and reason exactly as the crashed run left them), and exits zero; the catch-up is an admission, not a gate refusal, so REQ-CMV-013's before-write refusal scope is not engaged, and a crash-plus-catch-up pair leaves exactly one row, never two — never a degenerate old-equals-new self-transition row. A re-run whose LATEST audit row already records the commanded value as its new value (a converged-and-audited duplicate) is refused having appended nothing, so the completed transition is counted once. An append failure after the queue write commits retries the append once, then exits non-zero with the audit-loss sentinel naming the card and the mutation, leaving the completed queue write in place — the single-store mirror of REQ-CMV-004 direction 4, whose surfaced missing row this same catch-up recovers once the operator re-runs.
- REQ-CMV-011: The audit row's actor shall be the lane label read from the lane marker environment variable when set, otherwise `operator` (the capture pattern of `factory decide`, `internal/cli/factory_audit_decide.go:30-34`), and the row shall additionally carry the session UUID resolved through the existing session-identity capture when one is resolvable.
- REQ-CMV-012: The verbs shall refuse a reason that is empty, whitespace-only, or longer than 500 runes, writing neither store nor audit row.
- REQ-CMV-013: **When** a mutation is refused at any gate before its write, the verb shall write no audit row and leave both stores unchanged, judged at row level (per-row content comparison on a quiesced store — the form of the migration criterion REQ-CMV-014; file-byte identity is not the criterion). The REQ-CMV-004 direction-6 catch-up is an admitted converged-state audit completion, not a refusal — it is outside this requirement's before-write refusal scope.

### C.7 Migration discipline

- REQ-CMV-014: Adding the `card_metadata_events` surface shall be row-preserving — additive `CREATE TABLE IF NOT EXISTS` only — and shall rewrite, migrate, or delete no existing row of any existing table. (This discipline also forecloses the third audit-landing option — adding a nullable card column to the existing run-keyed `events` table — per decision-index Q3.)

### C.8 Actor boundary

- REQ-CMV-015: **When** either verb is invoked from a lane session (the lane marker environment variable set), the verb shall refuse with the lane-boundary sentinel of the existing lane-mutation boundary (`factoryLaneRefusal`, `internal/cli/factory_card.go:120-124`); the metadata verbs are leader/operator acts. (Decision-index Q2, Default applied 2026-10-17: no lane carve-out — t1715's blank-fill affordance routes through the leader; this requirement encodes the settled default.)

## §D Constraints

- RED tests first: every requirement lands as a failing test before its implementation (TDD mode per `.moai/config/sections/quality.yaml` `constitution.development_mode: tdd`).
- Config flag default OFF: `workflow.card_metadata_verbs.enabled`, distributed default false (design doc §6 "Each stage: RED tests first, config flag default off, row-preserving migration").
- Row-preserving migration: no destructive schema change (design doc §6).
- Cross-store write sequence: factory-record write first, queue write second, audit append only after both commit, all inside the queue lock — the failure matrix of REQ-CMV-004 is the honest boundary of `WithLock` (a lock, not a transaction); no cross-database transactionality is claimed or built.
- The queue store is the home SQLite database (the `<home>/db/<key>` layout), never the in-tree `backlog.db`; factory-database resolution for a queue store follows the sibling-resolution seam (`gtdFactoryDBForStore`, `internal/factory/gtd_operation.go:577+`, the site flagged P2 in a prior review).
- Reason bound: 500 runes maximum, named constant in the implementation.
- SPEC-ID grammar for `set-spec`: `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` (letter-leading uppercase middle segments, digit-only three-digit tail — the card delegation's named rule). Note: pick-time acceptance remains the looser historical pattern (`specIDPattern`, `internal/homestate/card_picked.go:15`); tightening pick-time is NOT part of this SPEC (see §F).
- English throughout (project instruction-doc language policy).

## §E Coordination

- **t1641** (picked, in queue): its addendum ("동사 부세 — pick 이후 spec_id 기록 불가" / "verb lacking — spec_id unrecordable after pick") named this exact gap. The implementation lives HERE. This SPEC absorbs that addendum: t1641 does not carry it, and t1641's remaining scope is untouched by this SPEC.
- **t1715** (queued): its lane self-recovery item (4) — "link a SPEC id when the record has none" — REUSES the `set-spec` verb this SPEC defines; t1715 must not re-implement it. t1715 keeps its own recovery verbs (re-acquire, re-stamp, adopt-tree). Decision-index Q2 (Default applied): the verbs are leader/operator-only, so t1715's item (4) routes through the leader; and the leader's dispatch of t1715 annotates its card text to reuse this verb rather than re-implement it (leader-side act — plan-audit iteration-1 D10; this SPEC cannot amend t1715's queue text from this tree).
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
- `internal/cli/factory_card.go` — `factory assign` `--spec` path (2810-2995), `requireQueuePicked` (2792), `factoryLaneRefusal` (120), the `WithLock`-wrapping-`RecordPicked` cross-store pattern (2919-2923).
- `internal/homestate/card_picked.go` — `CardFields` (63-71), spec-ID pattern (15), while-picked-only refusal (184-185), `card.fields` event (197).
- `internal/homestate/factory.go` — `cards` PK `(run_id, card_id)` (70-78), `events` table shape (79-84; `run_id TEXT NOT NULL DEFAULT ''` at 81).
- `internal/homestate/card_record.go` — `CardStates()` (21 states), `terminalStates` (done/failed/abandoned).
- `internal/homestate/card_transition.go` — `TransitionRequest` (34+), `appendEvent` (901), version fencing.
- `internal/factory/backlog_store.go` — `BacklogItem.SpecID` (89), `clearBacklogLease` (1175-1181), `WithLock`/`Mutate` (921-1000).
- `internal/factory/classification.go` — closed sets (33-68), `ValidateCardClassification` (109), `EffectiveCardClassification`.
- `internal/factory/gtd_operation.go` — `gtdFactoryDBForStore` sibling resolution (577+).
- `internal/cli/contract_report.go` — `cardToSpecID` (43-60).
- `internal/cli/factory_audit_decide.go` — actor capture pattern (30-34).
- `internal/cli/integration.go` — session-identity capture (138-150).
- `internal/cli/todo.go` — pick/unpick spec write sites (1671, 1749), `todo done` archive semantics (1319-1320), `recordFactoryCardState` (1765-1778).
- Status Transition Ownership Matrix: plan-phase artifacts authored by manager-spec (this agent); `draft → in-progress` owned by manager-develop; `in-progress → implemented → completed` by manager-docs (`.claude/rules/moai/development/spec-frontmatter-schema.md`).

## HISTORY

- 2026-10-17 v0.1.0 — Initial draft (manager-spec, card t1840 plan phase, stage E of the card-policy redesign). Absorbs the t1641 addendum naming the post-pick spec_id gap.
- 2026-10-17 v0.2.0 — Iteration-2 repair after plan-audit FAIL 0.81 (D1-D6): Q3 ruled — new card-keyed `card_metadata_events` table recorded as the Default and applied (D1, D8); REQ-CMV-004 rewritten as a bidirectional five-direction cross-store failure matrix with a compensating action and injected-failure AC (D2); consolidated 19 REQ / 21 AC to 15 REQ / 16 AC within the Tier M ceilings of 16/16 (D3); actor AC moved to the capture-layer function (D4); all-state sweep derived from `CardStates()` at the record layer with the verb-vs-record boundary stated (D5); set-spec on a card with no factory record refused at the admission gate with the `cards` PK `(run_id, card_id)` run-identity rule (D6); row-level store-identity wording (D7); t1715 leader-annotation note (D10).
- 2026-10-18 v0.3.0 — Iteration-3 repair after plan-audit iteration-2 FAIL 0.81 (N1-N3 blocking, N4-N5 optional — all applied): REQ-CMV-004 gained direction 6 (crash after both commits, before the audit append) with the catch-up recovery rule — gate (d) rewritten to refuse only a converged-AND-audited duplicate, its convergence comparison pinned to the QUEUE item (N5 folded: direction 5's heal rests on the same pin), and the catch-up defined as an admitted converged-state audit completion reconciled with REQ-CMV-010's exactly-one wording and REQ-CMV-013's refusal scope (N1); plan M1's GREEN rewritten to split the version-fenced set-spec card-row mutation function (no audit append inside the homestate transaction; the `card.fields` path cited for versioned-write mechanics only, never append timing) from the separate after-both-commits audit-append entry points (N2); reclassify's write target decoupled from the Card-row family — a queue-layer `WithLock`/`Mutate` classification write per REQ-CMV-009, stated in plan M1 and M3 (N3); the M2 refusal-atomicity sweep reworded to the gate-refusal set with the failure directions routed to the failure-matrix test (N4); AC-CMV-005 extended with the two seeded crash states and the duplicate-refusal case, and AC-CMV-003's no-change row re-seeded to the converged-AND-audited state. REQ count 15, AC count 16 — unchanged, within the Tier M ceilings.
- 2026-10-22 v0.4.0 — Sync-phase wording pass: plan-audit iteration-6 PASS-WITH-DEBT debts DB-2/DB-3/DB-4 (iteration-4's P2-P4, disposed to sync) applied, wording alignment only. DB-2: AC-CMV-003's no-change row re-worded to REQ-CMV-003 gate (d)'s LATEST-row duplicate test (the convergence comparison reads the QUEUE item; a duplicate requires the LATEST audit row's new value to equal the commanded value, an older row's match does not). DB-3: acceptance.md's §D matrix row for AC-CMV-005, §C crash-case cells, and AC-CMV-005's trailing map aligned to the implemented failure-matrix scope of M2's `TestSetSpec_FailureMatrix` — direction-5 heal, direction-6 catch-up, the ABA return-to-prior-value seed, and reclassify's single-store catch-up mirror (REQ-CMV-010 added to the row's map). DB-4: REQ-CMV-009 given the direction-6 catch-up carve-out its siblings (REQ-CMV-013, REQ-CMV-003 gate (d)) carry, mirroring REQ-CMV-013's deferring clause form. REQ count 15, AC count 16 — unchanged, within the Tier M ceilings.
