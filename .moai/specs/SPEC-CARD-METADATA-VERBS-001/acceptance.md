---
id: SPEC-CARD-METADATA-VERBS-001
title: "Acceptance criteria — card-metadata editing verbs"
version: "0.4.0"
created: 2026-10-17
updated: 2026-10-22
author: manager-spec
tags: "card-metadata, set-spec, reclassify, audit-trail, card-t1840"
---

# Acceptance Criteria — SPEC-CARD-METADATA-VERBS-001

## §A Command conventions (guard-safe forms)

- Every verification command is a plain single command, one git call per line, no shell variables, no process substitution, no compound pipes into `grep -E` with escaped alternation (worktree-session guard + known vacuous-green traps).
- Baseline anchors: pin SHAs at run start; never judge "unchanged" against a moving ref.
- Every "PASS = no match" predicate is exercised once against a positive control before being trusted.
- Store-identity assertions are judged at ROW level (per-row content comparison on a quiesced store — the AC-CMV-015 form); file-byte comparison against SQLite is not the criterion (iteration-1 audit D7).

## §B Given-When-Then scenarios

Labeled `AC-CMV-XXX`. Binary-testable; each names its verifying mechanism (Go test or CLI invocation).

### B.1 Capability gate

- **AC-CMV-001** — Given the distributed default config (flag key absent or false), When `set-spec`/`reclassify` is invoked, Then the verb exits non-zero with a refusal whose text contains `workflow.card_metadata_verbs.enabled`, and the queue store, the factory database, and the audit surface are unchanged, judged at row level. (maps REQ-CMV-001)
- **AC-CMV-002** — Given a fresh template deployment, When the shipped `workflow.yaml` is read, Then `workflow.card_metadata_verbs.enabled` is either absent or explicitly false — never true by default. (maps REQ-CMV-002)

### B.2 set-spec — validation gate

- **AC-CMV-003** — Given the refusal table below, When `set-spec` runs with each row, Then every case refuses, writing neither store nor audit row:
  - admission state: queue item `queued`, `dropped`, `hold`, or missing (refusal through the not-picked admission family);
  - no factory card record: a queue-picked, never-assigned card — the refusal names the missing record, no card row is created, and no run identity is invented (the `cards` table PK `(run_id, card_id)` is the reason a row cannot be created);
  - grammar: [`spec-auth-001` (lowercase), `SPEC-2FA-001` (digit-leading segment), `SPEC-AUTH-1` (tail not 3 digits), `AUTH-001` (prefix missing), `` (empty)] — each refusal quotes the rejected value;
  - no-change: `SPEC-FOO-001` equal to the value the queue item already carries AND the field's LATEST audit row carries that value as its new value (the converged-AND-audited duplicate — the convergence comparison reads the QUEUE item, never the factory record alone, and a duplicate requires the LATEST row's new value to equal the commanded value: the existence of an OLDER row whose new value equals it does not; the converged-but-unaudited state is the direction-6 admission, asserted in AC-CMV-005). (maps REQ-CMV-003)

### B.3 set-spec — writes

- **AC-CMV-004** — Given a picked card with blank spec_id, When `set-spec --spec SPEC-CMV-999 --reason "..."` succeeds, Then BOTH the queue item's `spec_id` and the factory card record's spec_id column read `SPEC-CMV-999`, and exactly one audit row exists for the mutation. (maps REQ-CMV-004, REQ-CMV-010)
- **AC-CMV-005** — Given an injected failure at each injectable point of the write sequence, When `set-spec` runs, Then each direction follows its REQ-CMV-004 rule: factory-side failure (before the record commit) → nothing committed, queue write never attempted, both stores unchanged; queue-side failure after the record commit → a compensating record write restores the prior spec_id and the verb exits non-zero with the divergence sentinel, no audit row; compensating-write failure → exit non-zero with the divergence sentinel naming card, prior value, and attempted value (divergence surfaced); audit-append failure after both commits → one retry, then exit non-zero with the audit-loss sentinel, store writes left in place. Given the seeded crash states, When `set-spec` is re-run, Then: record-new/queue-old (direction 5) → the re-run is ADMITTED — the no-change gate's convergence comparison reads the QUEUE item, never the factory record alone — and converges both stores on the commanded value with exactly one audit row; both-stores-new with no audit row (direction 6) → the re-run is admitted as the converged-state audit completion: it emits exactly the missing audit row (the re-run's actor and reason), leaves both stores row-unchanged, and exits zero; a further re-run after that completion (converged AND audited) → refused at the no-change gate with no second audit row. Given the ABA seed — audited X→A, then audited A→B, then a B→A run that crashes after both commits and before the append — When `set-spec` is re-run commanding A, Then the re-run is ADMITTED, because the duplicate test reads the field's LATEST audit row (whose new value is B, not A) and not any prior row, and it lands exactly one row with old=B and new=A, leaves both stores row-unchanged, and exits zero. And given the same crash window on `reclassify` — the queue classification write committed, the append never ran — When `reclassify` is re-run commanding the converged value, Then it is admitted as the converged-state audit completion keyed to the queue classification field: it emits exactly the owed row (the re-run's actor and reason, the old value from the field's LATEST audit row, empty where none exists), mutates no store, and exits zero; a further re-run after that completion (converged AND audited) → refused with no second row; and an append failure after the queue commit retries once, then exits non-zero with the audit-loss sentinel, the queue write left in place (REQ-CMV-010's reclassify reconciliation). (maps REQ-CMV-004, REQ-CMV-010)
- **AC-CMV-006** — Given factory card rows at every state `CardStates()` enumerates (all 21 states on this tree — `picked` plus the 20 post-picked states, derived from the source enumeration `internal/homestate/card_record.go`, not a hand-picked subset), When the RECORD-LAYER mutation function runs against each row, Then every row's spec_id updates and its version increments by exactly 1, and the pre-existing while-picked refusal of `assign --spec` for post-picked rows is preserved (contrast case). Verb-level admission remains queue-picked-only per REQ-CMV-003 — a done card's queue item is archived (`todo done`), so terminal-state rows are unreachable through the verb and this criterion exercises the record layer, named as such. (maps REQ-CMV-005)
- **AC-CMV-007** — Given a card row read at version N and concurrently advanced to N+1, When the record-layer mutation function presents expected version N, Then the write refuses with the stale-version error and writes nothing. (maps REQ-CMV-005)
- **AC-CMV-008** — Given a card whose spec_id was just set via `set-spec`, When the contract-report resolution path (`cardToSpecID`) reads the card, Then it returns the new spec_id with an empty cause. (maps REQ-CMV-006)

### B.4 reclassify

- **AC-CMV-009** — Given the refusal table below, When `reclassify` runs with each row, Then every case refuses and writes nothing: admission state — items `dropped`, `hold`, or absent refused, `queued` and `picked` admitted (positively enumerated); closed sets — mode=[fast, "", SERIAL, parallel] and priority=[urgent, "", HIGH, normal-] refused through the existing closed-set validator's error messages; no axis — neither `--mode` nor `--priority` supplied (usage error). (maps REQ-CMV-007)
- **AC-CMV-010** — Given a card classified {priority: high, blocked: true, mode: serial}, When `reclassify --priority low` runs, Then priority reads low, mode still reads serial, and blocked still reads true; and given a card with NO classification record, When `reclassify --mode parallelizable` runs, Then the item acquires the classification field additively (priority derived normal, mode parallelizable, decider human, fresh stamp) and the pre-change JSON encoding marshals byte-identically for every field except the acquired one. (maps REQ-CMV-008, REQ-CMV-009)

### B.5 Audit trail

- **AC-CMV-011** — Given one successful set-spec and one successful reclassify, When the card-keyed `card_metadata_events` surface is read, Then exactly two rows exist, each carrying actor, created-at timestamp, the mutated field, the old value, the new value, and the caller-supplied reason string verbatim. (maps REQ-CMV-010)
- **AC-CMV-012** — Given the ACTOR-RESOLUTION function of the metadata audit layer (the capture-layer function, not the CLI verb — a lane session cannot reach the verb, REQ-CMV-015), When it is invoked with the lane marker env set to `lane-9`, Then it returns `lane-9`; given the marker unset, it returns `operator`; and where a session UUID is resolvable through the existing capture, the resolved identity carries it as a supplementary field. (maps REQ-CMV-011)
- **AC-CMV-013** — Given the reason table [``, `   `, 501-rune string], When either verb runs with each, Then every case refuses and writes nothing; given a 500-rune reason, the mutation succeeds. (maps REQ-CMV-012)
- **AC-CMV-014** — Given any refused mutation from B.2-B.4, When the audit surface is read, Then its row count is unchanged and both stores are unchanged, judged at row level. (maps REQ-CMV-013)

### B.6 Migration and boundary

- **AC-CMV-015** — Given a populated factory database (cards, events, and every other existing table pre-seeded), When the migration adding the `card_metadata_events` surface runs, Then every pre-existing row of every pre-existing table is byte-identical after the migration (row-preserving), and the migration is idempotent on a second run. (maps REQ-CMV-014)
- **AC-CMV-016** — Given the lane marker env set (a lane session), When either metadata verb runs, Then the verb refuses with the lane-boundary sentinel and writes nothing. (maps REQ-CMV-015)

## §C Edge cases

- Set-spec on a card whose factory record does not exist (queue picked, never assigned): REFUSED at REQ-CMV-003 gate (b) — the `cards` table is keyed `(run_id, card_id)` (`internal/homestate/factory.go:70-78`); this SPEC creates no row and invents no run identity for a card that has none. Recovery: assign first, then set-spec. (Replaces the iteration-1 framing in which the queue write landed record-less — that relaxed the BOTH-store invariant of REQ-CMV-004 without a rule; plan-audit iteration-1 D6.)
- Verb-vs-record-layer boundary: a done/failed/abandoned card's queue item is archived (`todo done`, `internal/cli/todo.go:1319-1320`), so the verb cannot reach terminal-state rows; the all-state write (REQ-CMV-005) is a record-layer property exercised through the mutation function (AC-CMV-006). (plan-audit iteration-1 D5.)
- Crash between the record commit and the queue write (REQ-CMV-004 direction 5): record-new/queue-old divergence; healed by re-running the verb, which is admitted (the no-change gate's convergence comparison reads the QUEUE item, never the factory record alone) and converges both stores on the commanded value — the heal's audit row takes its old value from the QUEUE item's pre-heal value and its new value from the commanded one.
- Crash after both commits and before the audit append (REQ-CMV-004 direction 6): both stores converged, no audit row, no sentinel; the re-run is admitted as the converged-state audit completion — it emits the row the crashed run owed and mutates neither store. The ABA return-to-prior-value corner (audited X→A, then audited A→B, then a B→A run that crashes in this window) is admitted by the same test: the duplicate reads the field's LATEST audit row's new value (B), not any older row's (A), and the row lands exactly once with old=B and new=A. The reclassify single-store mirror (REQ-CMV-010) admits the same crash window on the queue classification field. The converged-AND-audited duplicate re-run is the only no-change refusal left standing; the catch-up is an admission, not a refusal, so REQ-CMV-013's refusal scope is not engaged.
- Reclassify racing a pick (queued → picked between read and write): the `WithLock` write re-reads under the lock; the classification write lands on the post-pick item.
- Unicode reason strings: the 500-rune bound counts runes, not bytes.
- A spec_id differing only by surrounding whitespace: trimmed before comparison; `SPEC-FOO-001 ` is a no-change refusal, not a write.

## §D AC matrix

| AC | Requirement | Severity | Verifies (mechanism) |
|----|-------------|----------|----------------------|
| AC-CMV-001 | REQ-CMV-001 | P0 | Go test (flag-off refusal + row-level store assertion) |
| AC-CMV-002 | REQ-CMV-002 | P0 | Template default read (shipped workflow.yaml) |
| AC-CMV-003 | REQ-CMV-003 | P0 | Go test (validation-gate table: state, no-record, grammar, no-change) |
| AC-CMV-004 | REQ-CMV-004, REQ-CMV-010 | P0 | Go test (two-store assertion + one audit row) |
| AC-CMV-005 | REQ-CMV-004, REQ-CMV-010 | P0 | Go test (four injected failure directions + direction-5 heal, direction-6 catch-up, and the ABA return-to-prior-value seed; reclassify's single-store catch-up mirror) |
| AC-CMV-006 | REQ-CMV-005 | P0 | Go test (CardStates()-derived record-layer sweep + assign contrast) |
| AC-CMV-007 | REQ-CMV-005 | P1 | Go test (stale version) |
| AC-CMV-008 | REQ-CMV-006 | P0 | Go test (cardToSpecID resolution) |
| AC-CMV-009 | REQ-CMV-007 | P0 | Go test (admission + closed-set + no-axis table) |
| AC-CMV-010 | REQ-CMV-008, REQ-CMV-009 | P0 | Go test (axis preservation + additive acquisition) |
| AC-CMV-011 | REQ-CMV-010 | P0 | Go test (row shape) |
| AC-CMV-012 | REQ-CMV-011 | P1 | Go test (capture-layer actor table) |
| AC-CMV-013 | REQ-CMV-012 | P0 | Go test (reason table) |
| AC-CMV-014 | REQ-CMV-013 | P0 | Go test (refusal atomicity sweep, row-level) |
| AC-CMV-015 | REQ-CMV-014 | P0 | Go test (row-preserving + idempotent migration) |
| AC-CMV-016 | REQ-CMV-015 | P0 | Go test (lane sentinel) |

Coverage declaration: this matrix maps REQ-CMV-001 through REQ-CMV-015 (every requirement named in the Requirement column at least once; REQ-CMV-004 twice by design — happy path and failure matrix; REQ-CMV-005 twice — sweep and stale version).

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
