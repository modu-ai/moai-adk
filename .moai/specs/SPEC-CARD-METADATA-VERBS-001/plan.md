---
id: SPEC-CARD-METADATA-VERBS-001
title: "Implementation plan — card-metadata editing verbs"
version: "0.1.0"
created: 2026-10-17
updated: 2026-10-17
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

Cross-store sequencing precedent: `factory assign` runs the record write inside `queueStore.WithLock(...)` so a factory-side failure aborts the queue write (`internal/cli/factory_card.go`, the `WithLock` body around `db.RecordPicked`).

Actor capture precedents: lane label from the lane marker env (`factory_audit_decide.go:30-34`, `todo.go:1770`), `operator` default, session UUID via `integrationSessionID` (`internal/cli/integration.go:138-150`).

Lane boundary precedent: `factoryLaneRefusal` (`internal/cli/factory_card.go:120-124`), sentinel `lane boundary` (`factoryLaneBoundarySentinel`), base rule from SPEC-FACTORY-SELF-DISPATCH-001 (completed) REQ-SD-015/-016.

### Tier classification rationale

**Tier M.** Single subsystem (the factory/todo card record and its two stores), multi-file (internal/homestate, internal/factory, internal/cli, internal/config, template config), no cross-domain fan-out (no web, no template-mirror rule edits, no docs-site). 19 requirements, acceptance matrix at 1:1 plus edge cases — within the Tier M envelope this project's recent Tier M SPECs occupy (15-25 REQ). Tier L artifacts (design.md, research.md) are not warranted: the design decisions are three and each is a decision-index row, not a design document.

## §B Known Issues (baseline, measured on this tree at e806cdd6c)

- No post-pick spec_id setter exists. `RecordPicked` refuses past-picked field changes (`internal/homestate/card_picked.go:184-185`); unpick clears spec (`internal/cli/todo.go:1749`), so the only route is unpick + re-pick. The t1631 stopgap was a direct SQL UPDATE.
- Pick-time spec acceptance is looser than this SPEC's grammar: `specIDPattern = ^SPEC-[A-Z0-9]+(?:-[A-Z0-9]+)*$` (`card_picked.go:15`) accepts digit-leading segments and non-3-digit tails; the queue side records as-is with no validation at all (`todo.go:1672-1674`, "Recorded as-is: the store is not a SPEC registry"). Consequence: a spec_id accepted at pick can be refused by set-spec's stricter grammar. Deliberate (see spec.md §F); not a defect of this SPEC.
- The classification reason field has no length validation today (`classification.go:109-124` validates value sets only; "the classified_at stamp and reason may be empty" per `CardClassificationValue`). The 500-rune reason bound is NEW validation, introduced here for the metadata verbs only.
- The homestate `events` table is run-keyed (`run_id`, factory.go:79-84). A queued card has no run, so reclassify of a queued card cannot write a run-keyed event without inventing a run — the reason the audit surface is card-keyed (decision-index Q3).
- `gtdFactoryDBForStore` (`internal/factory/gtd_operation.go:577+`) resolves the factory DB sibling for a home-layout queue — the P2-flagged site in a prior review. The new verbs MUST route through this same resolution (or its CLI equivalent), never assume the in-tree `backlog.db` (which is a ghost — the live queue is the home DB).

## §C Pre-flight

- [ ] decision-index Q1-Q3 dispositions confirmed (defaults applied or operator verdicts recorded) before M2 begins — Q1 (verb family) shapes every CLI file M2 touches.
- [ ] `go build ./...` green on the base tree.
- [ ] `go test ./internal/homestate/... ./internal/factory/... ./internal/cli/...` green on the base tree (baseline for RED contrast).
- [ ] `moai spec lint .moai/specs/SPEC-CARD-METADATA-VERBS-001/spec.md` clean at strict level.

## §D Constraints

1. RED tests first — each milestone's test file lands with failing tests before the implementation commit (TDD, `quality.yaml` `development_mode: tdd`).
2. Config flag `workflow.card_metadata_verbs.enabled`, distributed default false; the template default in `.moai/config/sections/workflow.yaml` and the Go default (`internal/config`) must agree.
3. Row-preserving migration — additive `CREATE TABLE IF NOT EXISTS` only; no existing row rewritten, deleted, or migrated.
4. Spec grammar `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` enforced by set-spec (named, exported-or-package-level pattern beside `specIDPattern`; do not mutate `specIDPattern` itself — pick-time behavior is out of scope).
5. Closed sets validated through `ValidateCardClassification` — no re-spelled copies (classification.go's own header rule: "no other file re-spells them").
6. Reason bound: 500 runes, named constant.
7. Queue-store writes go through `WithLock`/`Mutate` — never a direct engine write.
8. Actor capture reuses the env patterns; no new identity plumbing.
9. `moai spec lint` invocation takes the file path, not the SPEC ID (known gotcha).
10. Every git call in verification commands stays in plain pipeline position (worktree-session guard refuses compound forms).

## §E Self-Verification

Plan-phase: this document plus spec.md, acceptance.md, progress.md, decision-index.md authored by manager-spec; `moai spec lint` run on spec.md before the plan-audit; frontmatter validated against `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields, `phase` a release target, status `draft`).

## §F Milestones

Ordered by decision-reversibility: the data model lands first (hardest to reverse once audit rows exist), user-facing verbs second, mechanical integration last.

### M1 — Audit surface and record-layer mutation functions (Priority High)

Data model + homestate layer. RED tests first.

- RED: `TestCardMetadataEvent_RowShape` — one row per mutation carrying actor, timestamp, field, old, new, reason (REQ-CMV-014).
- RED: `TestCardMetadataSurface_RowPreserving` — pre-seeded rows of every existing table byte-identical after the migration runs (REQ-CMV-018).
- RED: `TestSetCardSpec_AnyState_Lands` — the new homestate mutation function writes spec_id at picked, assigned, run, and done-state rows, bumping version (REQ-CMV-007); contrast case documents that `RecordPicked` still refuses past-picked (characterization guard, so the while-picked rule is not silently widened for `assign`).
- RED: `TestSetCardSpec_StaleVersion_Refused` (REQ-CMV-007).
- GREEN: card-keyed metadata event surface (additive CREATE TABLE IF NOT EXISTS, decision-index Q3 default), the set-spec/reclassify mutation functions in `internal/homestate` beside `RecordPicked` (withCardTx + updateCardRow + event append; the `card.fields` event path is the precedent), strict spec-grammar pattern, version fencing.

### M2 — `set-spec` CLI verb + flag gate (Priority High)

User-facing surface. RED tests first.

- RED: `TestSetSpec_FlagOff_Refused` — refusal names the flag key; queue and factory stores untouched; no audit row (REQ-CMV-001, REQ-CMV-002 default).
- RED: `TestSetSpec_NotPicked_Refused` — missing item and each non-picked state (REQ-CMV-003).
- RED: `TestSetSpec_Grammar_Refused` — table: lowercase domain, digit-leading segment, missing 3-digit tail, `SPEC-` prefix missing, empty (REQ-CMV-004).
- RED: `TestSetSpec_NoChange_Refused` (REQ-CMV-005).
- RED: `TestSetSpec_TwoStoresWritten` — queue item AND factory record carry the new value; a factory-side failure rolls the queue write back (queue lock wrapping, the assign pattern) (REQ-CMV-006).
- RED: `TestSetSpec_ReasonValidation` — empty, whitespace-only, 501 runes refused; 500 accepted (REQ-CMV-016).
- RED: `TestSetSpec_RefusalAtomicity` — every refusal above leaves both stores byte-identical and writes no audit row (REQ-CMV-017).
- GREEN: verb under the family chosen in decision-index Q1 (default: `moai factory set-spec <card> --spec <id> --reason "..."`), config key + gate check, grammar validation, queue+record writes inside `WithLock`, audit append on success only.

### M3 — `reclassify` CLI verb (Priority Medium)

- RED: `TestReclassify_AdmissionStates` — queued and picked admitted; dropped, hold, missing refused, positive enumeration (REQ-CMV-009).
- RED: `TestReclassify_ClosedSets_Refused` — mode/priority tables through the real validator (REQ-CMV-010).
- RED: `TestReclassify_NoAxis_Refused` (REQ-CMV-011).
- RED: `TestReclassify_UnspecifiedAxisPreserved` + `TestReclassify_BlockedUntouched` (REQ-CMV-012).
- RED: `TestReclassify_WriteSemantics` — decider `human`, fresh stamp, reason landed; first-time classification acquires the nullable field additively (REQ-CMV-013).
- GREEN: verb, `--mode`/`--priority`/`--reason` flags, read-modify-write through `EffectiveCardClassification`, audit append.

### M4 — Actor resolution, lane boundary, consumer integration (Priority Medium)

- RED: `TestMetadataActor_Resolution` — lane env set → lane label; unset → `operator`; session UUID carried when resolvable (REQ-CMV-015).
- RED: `TestMetadataVerbs_LaneBoundary_Refused` — lane marker set → sentinel refusal for both verbs (REQ-CMV-019; Q2 default encoded).
- RED: `TestSetSpec_ConsumerResolves` — the `cardToSpecID` path resolves the new spec_id after set-spec (REQ-CMV-008).
- GREEN: capture wiring (reuse the two env patterns), boundary check ahead of the gate check order per the refusal-family conventions, consumer integration test against a seeded queue.

### M5 — Config surface and help text (Priority Low)

- Template `workflow.yaml` gains the commented-out default-off key; `internal/config` struct + defaults; help text for both verbs naming the flag; `moai doctor` untouched (flag absence is not a defect — default off).

### Anti-ordering note

Do not implement the CLI verbs before M1's mutation functions exist (M2/M3 RED tests exercise the record layer through the verbs and would otherwise double-implement the write path).

## §G Anti-Patterns

- **Do not** widen `RecordPicked`'s while-picked rule for `assign` — the new verb gets its own mutation function; `assign --spec` keeps its existing semantics (M1 contrast case pins this).
- **Do not** re-spell the closed sets or the spec grammar — reuse `ValidateCardClassification`; define the strict grammar as a NEW pattern beside `specIDPattern`, never a mutation of it.
- **Do not** write the queue store outside `WithLock`/`Mutate`.
- **Do not** key audit rows by run — a queued card has no run (Q3 rationale); do not invent a synthetic run id.
- **Do not** emit an audit row on a refused mutation (fail-atomic).
- **Do not** touch the in-tree `backlog.db` (ghost); the home DB is the live store, resolved through the sibling seam.
- **Do not** have t1715's recovery verbs re-implement set-spec (spec.md §E).

## §H Cross-References

- spec.md §C requirements ↔ acceptance.md §D matrix (1:1, declared "maps REQ-" form).
- decision-index.md Q1 (verb family) ← M2; Q2 (lane carve-out) ← M4/REQ-CMV-019; Q3 (audit landing) ← M1.
- SPEC-FACTORY-SELF-DISPATCH-001 (completed) — lane-boundary base rule.
- SPEC-TODO-CLASSIFY-DISPATCH-001 (completed) — the classification data model this SPEC's reclassify writes.
- Cards: t1840 (this), t1641 (addendum absorbed here), t1715 (reuses set-spec).
- Status Transition Ownership Matrix — plan artifacts: manager-spec; first run-phase commit flips `draft → in-progress`: manager-develop; sync close: manager-docs.
