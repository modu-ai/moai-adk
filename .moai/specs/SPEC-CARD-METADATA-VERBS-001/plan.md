---
id: SPEC-CARD-METADATA-VERBS-001
title: "Implementation plan — card-metadata editing verbs"
version: "0.3.0"
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
tags: "card-metadata, set-spec, reclassify, audit-trail, card-t1840"
---

# Implementation Plan — SPEC-CARD-METADATA-VERBS-001

## §A Context

Stage E of the card-policy redesign (card t1840, operator delegation 2026-10-17). Design authority: `.moai/reports/card-policy-redesign-20261017/design.md` §2 principle 10, §3 Metadata row, §6 stage E (machine-local in the primary tree — design intent, not committed-tree authority).

The two mutation surfaces today:

1. **Queue item** (`internal/factory/backlog_store.go:89` `SpecID *string`; classification at :116). Spec is written only at pick (`internal/cli/todo.go:1671`), cleared at unpick (:1749) and lease reclamation (`clearBacklogLease`, backlog_store.go:1175-1181). Classification is add-path-only (`internal/factory/classification.go:83-84`).
2. **Factory card record** (`internal/homestate` cards table). `RecordPicked` writes fields only while `picked` (`card_picked.go:184-185`) and bumps version with a `card.fields` event (:192-198) — the in-codebase precedent for a versioned field mutation with an event.

Consumer: `cardToSpecID` (`internal/cli/contract_report.go:44-60`) reads the QUEUE item — so a set-spec that wrote only the factory record would be invisible to the consumer. Both stores must move.

Cross-store sequencing precedent: `factory assign` runs the record write inside `queueStore.WithLock(...)` so a factory-side failure aborts the queue write (`internal/cli/factory_card.go:2919-2923`). `WithLock` is a lock, not a transaction (backlog_store.go:938-952) — REQ-CMV-004's failure matrix is the honest boundary of that mechanism, and this plan builds no cross-database transactionality.

Actor capture precedents: lane label from the lane marker env (`factory_audit_decide.go:30-34`, `todo.go:1770`), `operator` default, session UUID via `integrationSessionID` (`internal/cli/integration.go:138-150`).

Lane boundary precedent: `factoryLaneRefusal` (`internal/cli/factory_card.go:120-124`), sentinel `lane boundary` (`factoryLaneBoundarySentinel`), base rule from SPEC-FACTORY-SELF-DISPATCH-001 (completed) REQ-SD-015/-016.

### Tier classification rationale

**Tier M.** Single subsystem (the factory/todo card record and its two stores), multi-file (internal/homestate, internal/factory, internal/cli, internal/config, template config), no cross-domain fan-out (no web, no template-mirror rule edits, no docs-site). 15 requirements and 16 acceptance criteria — at the Tier M REQ/AC ceilings of 16/16 (`spec-workflow.md` § SPEC Complexity Tier). Iteration-1 audit D3 found the original 19 REQ / 21 AC over the ceilings; the repair merged the three set-spec validation REQs into REQ-CMV-003, merged the reclassify admission trio into REQ-CMV-007, and folded the paired ACs (validation tables into single table criteria; the failure directions into one injected-failure criterion). Tier L artifacts (design.md, research.md) are not warranted: the design decisions are three and each is a settled decision-index row, not a design document.

## §B Known Issues (baseline, measured on this tree)

- No post-pick spec_id setter exists. `RecordPicked` refuses past-picked field changes (`internal/homestate/card_picked.go:184-185`); unpick clears spec (`internal/cli/todo.go:1749`), so the only route is unpick + re-pick. The t1631 stopgap was a direct SQL UPDATE.
- Pick-time spec acceptance is looser than this SPEC's grammar: `specIDPattern = ^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$` (`card_picked.go:15`) accepts digit-leading segments and non-3-digit tails; the queue side records as-is with no validation at all (`todo.go:1672-1674`, "Recorded as-is: the store is not a SPEC registry"). Consequence: a spec_id accepted at pick can be refused by set-spec's stricter grammar. Deliberate (see spec.md §F); not a defect of this SPEC.
- The classification reason field has no length validation today (`classification.go:109-124` validates value sets only; "the classified_at stamp and reason may be empty" per `CardClassificationValue`). The 500-rune reason bound is NEW validation, introduced here for the metadata verbs only.
- The homestate `events` table is run-keyed (`run_id`, factory.go:79-84). An empty `run_id` IS schema-legal (`run_id TEXT NOT NULL DEFAULT ''`, factory.go:81) — iteration-1 plan-audit D8 correction. The card-keyed landing (decision-index Q3, Default applied) rests on consumer/keying discipline and design principle 5, not on schema legality.
- The `cards` table is keyed `(run_id, card_id)` (factory.go:70-78): a queue-picked, never-assigned card has NO factory row and no run identity to land under — set-spec refuses at the admission gate for such cards (REQ-CMV-003 gate (b); iteration-1 plan-audit D6).
- `gtdFactoryDBForStore` (`internal/factory/gtd_operation.go:577+`) resolves the factory DB sibling for a home-layout queue — the P2-flagged site in a prior review. The new verbs MUST route through this same resolution (or its CLI equivalent), never assume the in-tree `backlog.db` (which is a ghost — the live queue is the home DB).

## §C Pre-flight

- [ ] decision-index dispositions in force, gated per milestone (iteration-1 plan-audit D1 fix): Q3 REQUIRED before M1 begins — SATISFIED at plan time (Default applied 2026-10-17: new card-keyed `card_metadata_events` table); Q1 REQUIRED before M2 begins (Default: `moai factory` family only; product-level row — operator confirms at Kickoff); Q2 REQUIRED before M4 begins — SATISFIED at plan time (Default applied 2026-10-17: no lane carve-out).
- [ ] `go build ./...` green on the base tree.
- [ ] `go test ./internal/homestate/... ./internal/factory/... ./internal/cli/...` green on the base tree (baseline for RED contrast).
- [ ] `moai spec lint .moai/specs/SPEC-CARD-METADATA-VERBS-001/spec.md --strict` clean at strict level.

## §D Constraints

1. RED tests first — each milestone's test file lands with failing tests before the implementation commit (TDD, `quality.yaml` `development_mode: tdd`).
2. Config flag `workflow.card_metadata_verbs.enabled`, distributed default false; the Go config default (internal/config) lands in M2 with the gate check, the template key and help text land in M5 — one milestone owns each config surface (iteration-1 plan-audit D9 fix); the two must agree when both have landed.
3. Row-preserving migration — additive `CREATE TABLE IF NOT EXISTS card_metadata_events` only; no existing row rewritten, deleted, or migrated.
4. Spec grammar `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` enforced by set-spec (named, exported-or-package-level pattern beside `specIDPattern`; do not mutate `specIDPattern` itself — pick-time behavior is out of scope).
5. Closed sets validated through `ValidateCardClassification` — no re-spelled copies (classification.go's own header rule: "no other file re-spells them").
6. Reason bound: 500 runes, named constant.
7. Queue-store writes go through `WithLock`/`Mutate` — never a direct engine write. Write order: factory-record write first, queue write second, audit append only after both commit (REQ-CMV-004).
8. Actor capture reuses the env patterns; no new identity plumbing.
9. `moai spec lint` invocation takes the file path, not the SPEC ID (known gotcha).
10. Every git call in verification commands stays in plain pipeline position (worktree-session guard refuses compound forms).

## §E Self-Verification

Plan-phase: this document plus spec.md, acceptance.md, progress.md, decision-index.md authored by manager-spec; `moai spec lint` run on spec.md at strict level before the plan-audit; frontmatter validated against `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields, `phase` a release target, status `draft`). Iteration-2 repair (v0.2.0) applied the plan-audit iteration-1 fix routes D1-D6 (blocking) and D7-D10 (optional, applied where cheap). Iteration-3 repair (v0.3.0) applied the iteration-2 blockers N1-N3 and the optional N4-N5: the REQ-CMV-004 direction-6 catch-up with the gate-(d) converged-AND-audited semantics and QUEUE-read pin, the M1 commit-boundary split (audit append outside the homestate transaction), the reclassify queue-layer write-target decoupling, and the refusal-atomicity sweep rewording.

## §F Milestones

Ordered by decision-reversibility: the data model lands first (hardest to reverse once audit rows exist), user-facing verbs second, mechanical integration last.

### M1 — Audit surface and the set-spec record-layer mutation function (Priority High)

Data model + homestate layer. RED tests first. Pre-condition: decision-index Q3 disposition (satisfied — Default applied).

- RED: `TestCardMetadataEvent_RowShape` — one row per mutation carrying actor, timestamp, field, old, new, reason (REQ-CMV-010).
- RED: `TestCardMetadataSurface_RowPreserving` — pre-seeded rows of every existing table byte-identical after the migration runs (REQ-CMV-014).
- RED: `TestSetCardSpec_AnyState_Lands` — the record-layer mutation function writes spec_id at every state `CardStates()` enumerates (all 21: picked plus the 20 post-picked; the sweep derives from the source enumeration, not a hand-picked subset), bumping version (REQ-CMV-005); contrast case documents that `RecordPicked` still refuses past-picked (characterization guard, so the while-picked rule is not silently widened for `assign`).
- RED: `TestSetCardSpec_StaleVersion_Refused` (REQ-CMV-005).
- GREEN: card-keyed metadata event surface — additive `CREATE TABLE IF NOT EXISTS card_metadata_events`, the decision-index Q3 Default (applied 2026-10-17) — plus, scoped to `set-spec` ONLY, the record-layer card-row mutation function in `internal/homestate` beside `RecordPicked` (withCardTx + updateCardRow + version fencing; NO audit append inside this function and none inside the homestate transaction — the `card.fields` event path at `card_picked.go:197` is the precedent for the versioned-write MECHANICS only, never for append timing, because REQ-CMV-004 orders the audit row after both stores commit), the strict spec-grammar pattern, the compensating-restore entry point REQ-CMV-004 direction 2 requires, and the separate audit-append entry points — the after-both-commits append and the direction-6 catch-up append — both invoked outside the mutation function, after both stores have committed (the catch-up mutates neither store). `reclassify` gets NO card-row mutation function here: its write target is the queue item's classification record (REQ-CMV-009), landed in M3 as a queue-layer `WithLock`/`Mutate` write — only its audit append touches this surface.

### M2 — `set-spec` CLI verb + flag gate (Priority High)

User-facing surface. RED tests first. Pre-condition: decision-index Q1 disposition before this milestone begins (Default: `moai factory` family; operator confirms at Kickoff).

- RED: `TestSetSpec_FlagOff_Refused` — refusal names the flag key; queue and factory stores untouched (row level); no audit row (REQ-CMV-001, REQ-CMV-002 default).
- RED: `TestSetSpec_ValidationGate` — the refusal table of AC-CMV-003: admission states, no-factory-record (queue-picked, never-assigned), grammar cases, no-change — the converged-AND-audited duplicate of gate (d) (REQ-CMV-003; the converged-but-unaudited admission is asserted in `TestSetSpec_FailureMatrix`, direction 6).
- RED: `TestSetSpec_TwoStoresWritten` — queue item AND factory record carry the new value; exactly one audit row (REQ-CMV-004, REQ-CMV-010).
- RED: `TestSetSpec_FailureMatrix` — the four injected directions plus the two seeded crash states of AC-CMV-005: factory-side failure (nothing committed, queue write never attempted), queue-side failure after the record commit (compensating restore + divergence sentinel, no audit row), compensating-write failure (sentinel names card, prior value, attempted value), audit-append failure after both commits (one retry, audit-loss sentinel, stores left in place); seeded record-new/queue-old (direction 5) — the re-run is admitted and converges both stores with exactly one audit row; seeded both-converged/unaudited (direction 6) — the re-run is admitted as the converged-state audit completion, emits exactly the owed row with the re-run's actor and reason, mutates neither store, exits zero; and the duplicate re-run after that completion (converged AND audited) is refused with no second row (REQ-CMV-004).
- RED: `TestSetSpec_ReasonValidation` — empty, whitespace-only, 501 runes refused; 500 accepted (REQ-CMV-012).
- RED: `TestSetSpec_RefusalAtomicity` — the gate-refusal set only — flag-off, the REQ-CMV-003 admission/no-record/grammar/no-change gates, and the REQ-CMV-012 reason refusals (every refusal at a gate BEFORE the write, REQ-CMV-013's scope) — leaves both stores unchanged (row level) and writes no audit row; the failure directions are NOT in this sweep (their rules deliberately leave residue — the compensating restore, the surfaced divergence, the audit-loss sentinel) and are asserted per-direction by `TestSetSpec_FailureMatrix`, as is the direction-6 catch-up, which is an admitted completion, not a refusal (REQ-CMV-013).
- GREEN: verb under the decision-index Q1 Default family (`moai factory set-spec <card> --spec <id> --reason "..."`), config key + gate check + the Go config default (the template key and help text land in M5 — one milestone per config surface), grammar validation, record→queue write order inside `WithLock`, audit append only after both stores commit through M1's separate append entry point, and the no-change gate comparing against the QUEUE item — refusing only a converged-AND-audited duplicate and admitting the direction-6 catch-up.

### M3 — `reclassify` CLI verb (Priority Medium)

- RED: `TestReclassify_ValidationGate` — admission states (queued/picked admitted; dropped, hold, missing refused, positive enumeration), closed-set mode/priority tables through the real validator, no-axis usage refusal (REQ-CMV-007).
- RED: `TestReclassify_UnspecifiedAxisPreserved` + `TestReclassify_BlockedUntouched` (REQ-CMV-008).
- RED: `TestReclassify_WriteSemantics` — decider `human`, fresh stamp, reason landed; first-time classification acquires the nullable field additively (REQ-CMV-009).
- RED: `TestReclassify_AuditWindowCatchUp` — REQ-CMV-010's reclassify reconciliation for the queue-commit→append crash window: seeded converged-but-unaudited classification (queue write committed, append never ran) → the re-run is admitted as the converged-state audit completion keyed to the queue classification field — exactly one owed row (old value from the field's LATEST audit row, empty where none exists), no store mutation (classified-at stamp and reason as the crashed run left them), exit zero, never a degenerate self-transition row; converged-and-audited duplicate re-run (LATEST row already records the commanded value) → refused, no second row; append failure after the queue commit → one retry, then the audit-loss sentinel with the queue write left in place (REQ-CMV-010).
- GREEN: verb, `--mode`/`--priority`/`--reason` flags, the queue-layer `WithLock`/`Mutate` classification write — REQ-CMV-009's write target, the queue item's classification record; NO Card-row write and no run identity (a queued-only card has no factory row — REQ-CMV-007 admits `queued`), so a queued-only card takes the queue-only path and only the audit append touches homestate — read-modify-write through `EffectiveCardClassification`, audit append after the queue write commits — plus the audit-window wiring REQ-CMV-010 adds for reclassify: the latest-row duplicate admission test ahead of the write, the catch-up append through M1's separate catch-up entry point (re-mutating no store), and the single-retry append path with the audit-loss sentinel mirroring REQ-CMV-004 direction 4.

### M4 — Actor resolution, lane boundary, consumer integration (Priority Medium)

Pre-condition: decision-index Q2 disposition (satisfied — Default applied: no lane carve-out).

- RED: `TestMetadataActor_Resolution` — the CAPTURE-LAYER function: lane env set → lane label; unset → `operator`; session UUID carried when resolvable (REQ-CMV-011; iteration-1 plan-audit D4 fix — the actor table tests the capture-layer function, because a lane session cannot reach the verb at all under the Q2 default).
- RED: `TestMetadataVerbs_LaneBoundary_Refused` — lane marker set → sentinel refusal for both verbs (REQ-CMV-015; Q2 Default encoded).
- RED: `TestSetSpec_ConsumerResolves` — the `cardToSpecID` path resolves the new spec_id after set-spec (REQ-CMV-006).
- GREEN: capture wiring (reuse the two env patterns), boundary check ahead of the gate check order per the refusal-family conventions, consumer integration test against a seeded queue.

### M5 — Template config surface and help text (Priority Low)

- Template `workflow.yaml` gains the commented-out default-off key, and help text for both verbs naming the flag — ONLY those two surfaces (the Go config struct + default + gate check landed in M2; iteration-1 plan-audit D9 fix — no config surface is claimed by two milestones). `moai doctor` untouched (flag absence is not a defect — default off).

### Anti-ordering notes

- Do not implement the CLI verbs before M1's mutation functions exist (M2/M3 RED tests exercise the record layer through the verbs and would otherwise double-implement the write path).
- Do not implement the compensating restore ad hoc inside the verb — it is the M1 mutation-function entry point the M2 failure-matrix tests inject against (D2 residual-risk: an implementation that copies the `assign` precedent verbatim re-enters the one-directional gap through the milestone tests unless the matrix AC exists first — it now does, AC-CMV-005).

## §G Anti-Patterns

- **Do not** widen `RecordPicked`'s while-picked rule for `assign` — the new verb gets its own mutation function; `assign --spec` keeps its existing semantics (M1 contrast case pins this).
- **Do not** re-spell the closed sets or the spec grammar — reuse `ValidateCardClassification`; define the strict grammar as a NEW pattern beside `specIDPattern`, never a mutation of it.
- **Do not** write the queue store outside `WithLock`/`Mutate`.
- **Do not** claim or build cross-database transactionality — `WithLock` is a lock; the REQ-CMV-004 failure matrix defines every divergence direction and its rule (compensating restore, surfaced double-failure, audit-loss sentinel, re-run heal, converged-state catch-up).
- **Do not** key audit rows by run — a queued card has no run (Q3 Default rationale); do not invent a synthetic run id; do not add a nullable card column to the existing `events` table (foreclosed by REQ-CMV-014).
- **Do not** create a factory card row for a never-assigned card or invent a run identity for it — set-spec refuses (REQ-CMV-003 gate (b)).
- **Do not** emit an audit row for a mutation whose store writes did not both commit (fail-atomic; the append runs only after both stores commit — the direction-6 catch-up appends precisely because both stores DID commit in the crashed run; the row is owed, not fabricated).
- **Do not** touch the in-tree `backlog.db` (ghost); the home DB is the live store, resolved through the sibling seam.
- **Do not** have t1715's recovery verbs re-implement set-spec (spec.md §E; the leader's dispatch of t1715 annotates its card text to reuse this verb).

## §H Cross-References

- spec.md §C requirements ↔ acceptance.md §D matrix (15 REQ / 16 AC, declared "maps REQ-" form; REQ-CMV-004 and REQ-CMV-005 covered twice by design).
- decision-index.md Q1 (verb family) ← gate before M2; Q2 (lane carve-out) ← gate before M4, Default applied; Q3 (audit landing) ← gate before M1, Default applied (new card-keyed `card_metadata_events` table).
- SPEC-FACTORY-SELF-DISPATCH-001 (completed) — lane-boundary base rule.
- SPEC-TODO-CLASSIFY-DISPATCH-001 (completed) — the classification data model this SPEC's reclassify writes.
- Cards: t1840 (this), t1641 (addendum absorbed here), t1715 (reuses set-spec; leader-side annotation of its card text requested at landing — iteration-1 plan-audit D10).
- Status Transition Ownership Matrix — plan artifacts: manager-spec; first run-phase commit flips `draft → in-progress`: manager-develop; sync close: manager-docs.
