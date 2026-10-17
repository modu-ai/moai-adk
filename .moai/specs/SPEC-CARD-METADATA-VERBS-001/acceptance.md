---
id: SPEC-CARD-METADATA-VERBS-001
title: "Acceptance criteria — card-metadata editing verbs"
version: "0.1.0"
created: 2026-10-17
updated: 2026-10-17
author: manager-spec
tags: "card-metadata, set-spec, reclassify, audit-trail, card-t1840"
---

# Acceptance Criteria — SPEC-CARD-METADATA-VERBS-001

## §A Command conventions (guard-safe forms)

- Every verification command is a plain single command, one git call per line, no shell variables, no process substitution, no compound pipes into `grep -E` with escaped alternation (worktree-session guard + known vacuous-green traps).
- Baseline anchors: pin SHAs at run start; never judge "unchanged" against a moving ref.
- Every "PASS = no match" predicate is exercised once against a positive control before being trusted.

## §B Given-When-Then scenarios

Labeled `AC-CMV-XXX`. Binary-testable; each names its verifying mechanism (Go test or CLI invocation).

### B.1 Capability gate

- **AC-CMV-001** — Given the distributed default config (flag key absent or false), When `set-spec`/`reclassify` is invoked, Then the verb exits non-zero with a refusal whose text contains `workflow.card_metadata_verbs.enabled`, and the queue store, the factory database, and the audit surface are byte-identical to before. (maps REQ-CMV-001)
- **AC-CMV-002** — Given a fresh template deployment, When the shipped `workflow.yaml` is read, Then `workflow.card_metadata_verbs.enabled` is either absent or explicitly false — never true by default. (maps REQ-CMV-002)

### B.2 set-spec — admission and validation

- **AC-CMV-003** — Given a card whose queue item is `queued` (or missing, or `dropped`/`hold`), When `set-spec` runs, Then the verb refuses with the not-picked admission family and writes nothing. (maps REQ-CMV-003)
- **AC-CMV-004** — Given a picked card and the spec-id table [`spec-auth-001` (lowercase), `SPEC-2FA-001` (digit-leading segment), `SPEC-AUTH-1` (tail not 3 digits), `AUTH-001` (prefix missing), `` (empty)], When `set-spec` runs with each, Then every case refuses, the refusal quotes the rejected value, and nothing is written. (maps REQ-CMV-004)
- **AC-CMV-005** — Given a picked card already carrying `SPEC-FOO-001` in the queue item, When `set-spec --spec SPEC-FOO-001` runs, Then the verb refuses as a no-change mutation with no store write and no audit row. (maps REQ-CMV-005)

### B.3 set-spec — writes

- **AC-CMV-006** — Given a picked card with blank spec_id, When `set-spec --spec SPEC-CMV-999 --reason "..."` succeeds, Then BOTH the queue item's `spec_id` and the factory card record's spec_id column read `SPEC-CMV-999`. (maps REQ-CMV-006)
- **AC-CMV-007** — Given a picked card and an injected factory-side write failure, When `set-spec` runs, Then the queue write is rolled back — the queue item still reads the old spec_id (cross-store atomicity under the queue lock). (maps REQ-CMV-006)
- **AC-CMV-008** — Given factory card rows at states picked, assigned, run-phase, and done, When `set-spec` runs against each, Then every row's spec_id updates, every row's version increments by exactly 1, and the pre-existing while-picked refusal of `assign --spec` for post-picked rows is preserved (contrast case). (maps REQ-CMV-007)
- **AC-CMV-009** — Given a card row read at version N and concurrently advanced to N+1, When `set-spec` presents expected version N, Then the write refuses with the stale-version error and writes nothing. (maps REQ-CMV-007)
- **AC-CMV-010** — Given a card whose spec_id was just set via `set-spec`, When the contract-report resolution path (`cardToSpecID`) reads the card, Then it returns the new spec_id with an empty cause. (maps REQ-CMV-008)

### B.4 reclassify

- **AC-CMV-011** — Given queue items in states queued and picked, When `reclassify` runs on each, Then both succeed; given items in dropped, hold, or absent, Then each refuses with the positively-enumerated admission error. (maps REQ-CMV-009)
- **AC-CMV-012** — Given the out-of-set tables mode=[fast, "", SERIAL, parallel] and priority=[urgent, "", HIGH, normal-], When `reclassify` runs with each, Then every case refuses through the existing closed-set validator's error messages. (maps REQ-CMV-010)
- **AC-CMV-013** — Given a queued card, When `reclassify` runs with neither `--mode` nor `--priority`, Then the verb refuses with a usage error and writes nothing. (maps REQ-CMV-011)
- **AC-CMV-014** — Given a card classified {priority: high, blocked: true, mode: serial}, When `reclassify --priority low` runs, Then priority reads low, mode still reads serial, and blocked still reads true. (maps REQ-CMV-012)
- **AC-CMV-015** — Given a card with NO classification record, When `reclassify --mode parallelizable` runs, Then the item acquires the classification field additively (priority derived normal, mode parallelizable, decider human) and the pre-change JSON encoding of the item marshals byte-identically to before for every field except the acquired one. (maps REQ-CMV-012, REQ-CMV-013)

### B.5 Audit trail

- **AC-CMV-016** — Given one successful set-spec and one successful reclassify, When the card-keyed metadata event surface is read, Then exactly two rows exist, each carrying actor, created-at timestamp, the mutated field, the old value, the new value, and the caller-supplied reason string verbatim. (maps REQ-CMV-014)
- **AC-CMV-017** — Given the lane marker env set to `lane-9`, When a metadata verb succeeds from that environment, Then the audit row's actor reads `lane-9`; given the marker unset, the actor reads `operator`; and where a session UUID is resolvable through the existing capture, the row carries it as a supplementary field. (maps REQ-CMV-015)
- **AC-CMV-018** — Given the reason table [``, `   `, 501-rune string], When either verb runs with each, Then every case refuses and writes nothing; given a 500-rune reason, the mutation succeeds. (maps REQ-CMV-016)
- **AC-CMV-019** — Given any refused mutation from B.2-B.4, When the audit surface is read, Then its row count is unchanged and both stores are byte-identical to before the attempt. (maps REQ-CMV-017)

### B.6 Migration and boundary

- **AC-CMV-020** — Given a populated factory database (cards, events, and every other existing table pre-seeded), When the migration adding the audit surface runs, Then every pre-existing row of every pre-existing table is byte-identical after the migration (row-preserving), and the migration is idempotent on a second run. (maps REQ-CMV-018)
- **AC-CMV-021** — Given the lane marker env set (a lane session), When either metadata verb runs, Then the verb refuses with the lane-boundary sentinel and writes nothing. (maps REQ-CMV-019)

## §C Edge cases

- Set-spec on a card whose factory record does not exist yet (queue picked, never assigned): the queue write lands; the record-side event is card-keyed and lands without a run; no synthetic run is invented.
- Reclassify racing a pick (queued → picked between read and write): the `WithLock` write re-reads under the lock; the classification write lands on the post-pick item.
- Unicode reason strings: the 500-rune bound counts runes, not bytes.
- A spec_id differing only by surrounding whitespace: trimmed before comparison; `SPEC-FOO-001 ` is a no-change refusal, not a write.

## §D AC matrix

| AC | Requirement | Severity | Verifies (mechanism) |
|----|-------------|----------|----------------------|
| AC-CMV-001 | REQ-CMV-001 | P0 | Go test (flag-off refusal + store-bytes assertion) |
| AC-CMV-002 | REQ-CMV-002 | P0 | Template default read (shipped workflow.yaml) |
| AC-CMV-003 | REQ-CMV-003 | P0 | Go test (state table) |
| AC-CMV-004 | REQ-CMV-004 | P0 | Go test (grammar table) |
| AC-CMV-005 | REQ-CMV-005 | P1 | Go test (no-change refusal) |
| AC-CMV-006 | REQ-CMV-006 | P0 | Go test (two-store assertion) |
| AC-CMV-007 | REQ-CMV-006 | P0 | Go test (injected failure, rollback) |
| AC-CMV-008 | REQ-CMV-007 | P0 | Go test (state matrix + assign contrast) |
| AC-CMV-009 | REQ-CMV-007 | P1 | Go test (stale version) |
| AC-CMV-010 | REQ-CMV-008 | P0 | Go test (cardToSpecID resolution) |
| AC-CMV-011 | REQ-CMV-009 | P0 | Go test (admission table) |
| AC-CMV-012 | REQ-CMV-010 | P0 | Go test (closed-set tables) |
| AC-CMV-013 | REQ-CMV-011 | P1 | Go test (usage refusal) |
| AC-CMV-014 | REQ-CMV-012 | P0 | Go test (axis preservation) |
| AC-CMV-015 | REQ-CMV-012, REQ-CMV-013 | P1 | Go test (additive acquisition) |
| AC-CMV-016 | REQ-CMV-014 | P0 | Go test (row shape) |
| AC-CMV-017 | REQ-CMV-015 | P1 | Go test (actor table) |
| AC-CMV-018 | REQ-CMV-016 | P0 | Go test (reason table) |
| AC-CMV-019 | REQ-CMV-017 | P0 | Go test (refusal atomicity sweep) |
| AC-CMV-020 | REQ-CMV-018 | P0 | Go test (row-preserving + idempotent migration) |
| AC-CMV-021 | REQ-CMV-019 | P0 | Go test (lane sentinel) |

Coverage declaration: this matrix maps REQ-CMV-001 through REQ-CMV-019 one-to-one (every requirement named in the Requirement column at least once; REQ-CMV-006 twice by design — happy path and rollback).

## §E Quality gates (Definition of Done)

- All AC Go tests green; `go test ./internal/homestate/... ./internal/factory/... ./internal/cli/...` green.
- RED evidence: each milestone's test list (plan.md §F) was observed failing against the pre-implementation tree (commit graph witnesses baseline-first ordering — the RED test commit precedes its implementation commit).
- `gofmt` clean; `go vet` clean on touched packages.
- Coverage of the new mutation and audit code ≥ 85%.
- `moai spec lint .moai/specs/SPEC-CARD-METADATA-VERBS-001/spec.md --strict` — zero errors.
- Config flag default OFF in template and Go defaults (AC-CMV-002).
- The t1641 addendum and t1715 reuse notes (spec.md §E) referenced in the sync-phase report so neither card re-implements this surface.

## §F Forward-looking checks

- Stage B (durable identity): the card-keyed audit surface must survive the identity migration without rewriting rows — the additive discipline here is its precondition.
- Evidence-SHA carry-over (the third principle-10 binding) will extend the audit row shape; the field/old/new/reason payload must admit a third field kind without schema change (JSON payload, not fixed columns).
