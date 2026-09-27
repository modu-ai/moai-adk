# design.md — SPEC-FACTORY-RECORD-001 (card t1239)

Design for the F1 record layer. Names of Go identifiers below are proposals for the run phase, not
requirements; the requirements are in `spec.md` §B.

## Layer placement (lead design v7, layer 1)

```
queue backlog.db items        (operator-picked cards; schema frozen; F1 reads state only)
        │ card id, state = picked
        ▼
factory.db cards / workers / events   ◄── F1: this SPEC
        │ evidence path + commit SHA read by Go
        ▼
.moai/reports/<card>/ , card worktree git objects

$MOAI_HOME/db/<project-key>/            (ProjectDir, research.md R1)
  ├── factory/factory.db                ◄── F1 writes
  ├── todo/backlog.db                   ◄── F1 reads item state only
  └── contract/                         ◄── A2/A3 write; F1 stores a pointer, never writes
```

Layers 2-5 (controller, agent sessions, notifications, Decider) are F2/F3.

## Schema — `cards` columns after migration (schema version 4)

Existing columns keep their names and meaning: `run_id`, `card_id`, `owner_label`, `state`,
`version`, `evidence_path`, `updated_at`.

| Column (proposed) | Type | Default | Meaning |
|---|---|---|---|
| `stage` | TEXT | `''` | last resumable working state reached (`plan`, `plan-audit`, `run`, `sync`, `sync-audit`, `merge-ready`) |
| `lease_holder` | TEXT | `''` | worker label holding the lease; empty when none |
| `lease_expires_at` | TEXT | `''` | RFC 3339; empty when no lease |
| `heartbeat_at` | TEXT | `''` | last renewal by the holder |
| `decision_gate` | TEXT | `''` | `kickoff` or `question` while a decision is pending |
| `decision_question` | TEXT | `''` | operator-facing text for `question` |
| `decision_resume` | TEXT | `''` | state to resume after `needs-decision` |
| `decider` | TEXT | `''` | `human` in F1; F3 adds `llm`, `llm+jev` |
| `decided_at` | TEXT | `''` | time of the last decision |
| `failure_reason` | TEXT | `''` | reason recorded on `failed` |
| `hint_prefer` | TEXT | `''` | verbatim `key=value`, e.g. `backend=codex` |
| `hint_after` | TEXT | `''` | predecessor card id |
| `spec_id` | TEXT | `''` | SPEC identifier once one exists |
| `worktree_path` | TEXT | `''` | absolute path of the card worktree |
| `evidence_sha` | TEXT | `''` | commit SHA recorded at the last audit entry |
| `merge_sha` | TEXT | `''` | local merge commit |
| `merge_tree` | TEXT | `''` | tree hash of the merge commit |
| `remeasure_path` | TEXT | `''` | re-measure evidence file |
| `contract_spec_id` | TEXT | `''` | A1 pointer field `spec_id` |
| `contract_sha256` | TEXT | `''` | A1 pointer field `contract_sha256` (64 hex) |
| `contract_signed_at` | TEXT | `''` | A1 pointer field `signed_at` (RFC 3339) |
| `contract_event` | TEXT | `''` | locator of the signing event in `$MOAI_HOME/db/<project-key>/contract/`: the SHA-256 (64 hex) of the store line that recorded the signing; empty until the A3 store exists |

All new columns are `TEXT NOT NULL DEFAULT ''`, so `ALTER TABLE ... ADD COLUMN` needs no backfill and
every v3 row stays valid. No CHECK constraint is added to `state` (SQLite cannot alter one later, the
same reason the queue schema is frozen — research.md R5); the enum is enforced by the transition API.

`workers` gains no column. Heartbeat renewal writes the existing `workers.heartbeat_at`.

Migration: `migrateFactoryV3ToV4`, following `migrateFactoryV2ToV3` (research.md R2) — read
`PRAGMA table_info(cards)`, add only missing columns, `UPDATE meta SET value='4' WHERE
key='schema_version' AND value='3'`, all in one transaction, appended to the existing chain in
`OpenFactoryPath`.

## The contract pointer (lead decision R10)

The pointer carries A1's three stable fields (A1 design.md § F1 Reference Shape at `8a7cb0e22`, unchanged since `de8aee456`) plus
`contract_event`, which names the one line in the moai-owned contract store that recorded the signing.
The store directory resolves as `ProjectDir(root)/contract`, which is `$MOAI_HOME/db/<project-key>/contract/`
under the default layout — the sibling of `factory/` and `todo/`. F1:

- validates only the pointer's format (non-empty SPEC ID matching the SPEC-ID pattern, 64-hex digest,
  RFC 3339 time, and a 64-hex or empty event locator);
- never opens, creates, or writes the store directory;
- never verifies that the digest matches the contract file or that the event line exists.

A reader that wants to know whether a pointer is current runs `moai contract verify` (A1) and, once A3
lands, looks the event up in the store. The store's file name inside `contract/` is A3's choice
(at `710530d67` the signing events live in `events.jsonl`, receipts in `receipts.jsonl`, research.md R12); the locator names a line of `events.jsonl` and is independent of the file name.

**Forward note (plan-audit iteration 2, D12):** F1 only stores and format-checks `contract_event`; it
never computes it. Computing the line hash belongs to whoever writes the signing event and wires the
optional fourth `--contract-ref` value — the A3 signer (which appends the store line) or the F3
controller (which links a signed contract to a card). That SPEC must state the exact bytes hashed
(the stored line without its trailing newline is the natural choice) so readers can recompute it.

## States

| Group | States |
|---|---|
| Admission | `picked`, `assigned` |
| Lease-holding (working) | `leased`, `plan`, `plan-audit`, `run`, `sync`, `sync-audit`, `merge-ready`, `merging` |
| Decision-pending (no lease) | `kickoff`, `needs-decision` |
| Post-merge | `merged-local`, `pushed`, `ci-green` |
| Paused | `blocked` |
| Terminal | `done`, `failed`, `abandoned` |

`kickoff` is decision-pending, not lease-holding (D1 of plan-audit iteration 1): the worker's lease is
released on entry, so a human who takes longer than the lease duration cannot lose the pending
decision to an expiry. Approval returns the card to `assigned` with `stage = run`; the owner re-leases
and resumes into `run` through T4. The card-level order `kickoff → run` is preserved through the
stage, not through a direct edge.

A row whose `state` is not one of the 19 (a pre-v4 row) is a **legacy** row: `status` shows it, and
the only accepted transition is an operator `decide --choice abandon` (REQ-FR-003).

## Transition Table

Edges not listed are illegal (REQ-FR-004). "Guard" names what Go verifies before committing. Rows
T19 and T20 are reserved: listed so the machine is complete, refused in F1 (REQ-FR-006).

| # | From | To | Actor | Guard |
|---|---|---|---|---|
| T1 | (none) | `picked` | `assign` | queue item for the card id is `picked` (REQ-FR-022) |
| T2 | `picked` | `assigned` | `assign --to` | `after` predecessor merged (REQ-FR-016) |
| T3 | `assigned` | `leased` | worker (F2 verb / API) | holder registered in `workers` and equal to `owner_label` (REQ-FR-012) |
| T4a | `leased` | `plan` | holder | `stage` is `plan` or empty |
| T4b | `leased` | `run` | holder | `stage` is `run`, or `stage` is empty (Class A/B cards skip plan) |
| T4c | `leased` | `plan-audit` | holder | `stage` is `plan-audit` |
| T4d | `leased` | `sync` | holder | `stage` is `sync` |
| T4e | `leased` | `sync-audit` | holder | `stage` is `sync-audit` |
| T4f | `leased` | `merge-ready` | holder | `stage` is `merge-ready` |
| T5 | `plan` | `plan-audit` | holder | E-ENTRY (REQ-FR-007) |
| T6 | `plan-audit` | `plan` | holder | E-VERDICT any verdict, audited SHA matches |
| T7 | `plan-audit` | `kickoff` | holder | E-VERDICT PASS / PASS-WITH-DEBT (REQ-FR-008); sets `decision_gate=kickoff`, clears lease (REQ-FR-018) |
| T8 | `kickoff` | `assigned` | `decide --gate kickoff --choice approve` | decider `human`; sets `stage=run` (REQ-FR-019) |
| T9 | `kickoff` | `blocked` | `decide --gate kickoff --choice reject` | decider `human` |
| T10 | `run` | `sync` | holder | supplied commit SHA resolves and is ancestor-or-equal of worktree HEAD |
| T11 | `sync` | `sync-audit` | holder | E-ENTRY |
| T12 | `sync-audit` | `sync` | holder | E-VERDICT any verdict |
| T13 | `sync-audit` | `merge-ready` | holder | E-VERDICT PASS / PASS-WITH-DEBT |
| T14 | `merge-ready` | `merging` | holder (F3 controller in autonomous mode) | lease valid |
| T15 | `merging` | `merge-ready` | holder | none (merge abandoned before commit) |
| T16 | `merging` | `merged-local` | holder / controller | E-MERGE (REQ-FR-009) |
| T17 | `merged-local` | `pushed` | `decide --gate push` (lead batch) | a remote exists and merge SHA is an ancestor of the remote-tracking integration ref (REQ-FR-021) |
| T18 | `merged-local` | `done` | `decide --gate push` | repository has no remote; event note `no remote — no CI verdict` (REQ-FR-021) |
| T19 | `pushed` | `ci-green` | — | **reserved** — refused in F1; CI verdict reader is F3 (REQ-FR-006) |
| T20 | `ci-green` | `done` | — | **reserved** — refused in F1 (REQ-FR-006) |
| T21 | each of `picked`, `assigned`, the 8 lease-holding states, `merged-local`, `pushed`, `ci-green` | `needs-decision` | any actor | question text required; lease cleared (REQ-FR-018) |
| T22a | `needs-decision` | `assigned` | `decide --choice resume` | `decision_resume` is `assigned` or a lease-holding state (then `stage := decision_resume` when resumable) |
| T22b | `needs-decision` | `picked` | `decide --choice resume` | `decision_resume` is `picked` |
| T22c | `needs-decision` | `merged-local` | `decide --choice resume` | `decision_resume` is `merged-local` |
| T22d | `needs-decision` | `pushed` | `decide --choice resume` | `decision_resume` is `pushed` |
| T22e | `needs-decision` | `ci-green` | `decide --choice resume` | `decision_resume` is `ci-green` |
| T23 | `needs-decision` | `blocked` | `decide --choice block` | human |
| T24 | `blocked` | `assigned` | `decide --choice unblock` | human; stage kept (REQ-FR-020) |
| T25 | each of the 16 non-terminal states | `abandoned` | `decide --choice abandon` | human |
| T26 | each of the 8 lease-holding states | `failed` | holder / API | non-empty reason; lease cleared (REQ-FR-015) |
| T27 | lease-holding except `merging`, lease expired | `assigned` | automatic, before any other transition (REQ-FR-013) | expiry passed |
| T28 | `merging`, lease expired | `blocked` | automatic (REQ-FR-014) | expiry passed |

T21 excludes `blocked` (as well as `kickoff` and `needs-decision`) from its sources on purpose: a
`blocked` card is already waiting on an operator decision (`unblock` or `abandon`), so routing it to a
second decision-pending state would create two pending decisions for one card with no rule for which
one wins.

### Requested-edge count (the number AC-005 asserts)

T1 creates a row and T27/T28 are automatic, so none is a requested (from, to) pair among the 19
states. With a remote configured (AC-005's fixture), T17 is acceptable and T18 is not; T19 and T20 are
refused. The accepted requested pairs are:

| Rows | Pairs |
|---|---|
| T2, T3, T5-T17 (15 single edges; T4 counted below) | 15 |
| T4a-T4f | 6 |
| T21 | 13 |
| T22a-T22e | 5 |
| T23, T24 | 2 |
| T25 | 16 |
| T26 | 8 |
| **Total** | **65** |

No pair appears in two rows (checked by the AC-005 test itself, which fails on a duplicate). The
remaining 361 − 65 = 296 pairs are refused. Each guarded row is exercised with a fixture that
satisfies its guard; the guard's negative side is exercised separately (AC-005 second clause).

## Evidence readers (Go reads; callers supply only paths and SHAs)

- **E-ENTRY** (T5, T11): inputs `sha`, `artifact_path`. Go runs, in the card's `worktree_path`:
  `git cat-file -e <sha>^{commit}`; `git merge-base --is-ancestor <sha> HEAD` (exit 0);
  `git cat-file -e <sha>:<artifact_path>`. On success it records `evidence_sha = sha`,
  `evidence_path = artifact_path`.
- **E-VERDICT** (T6, T7, T12, T13): Go reads the newest file matching
  `.moai/reports/<card-id>/{plan,sync}-audit*.md` (newest by the iteration number in the name, then by
  modification time) and extracts the two machine-readable lines REQ-FR-011 makes auditors emit:
  `^verdict: (PASS|PASS-WITH-DEBT|FAIL)$` and `^audited_sha: ([0-9a-f]{12,40})$`. The audited SHA must
  equal `evidence_sha` (a prefix of at least 12 hex characters accepted). A missing file, a missing
  line, two conflicting lines, or a mismatch refuses the transition.
- **E-MERGE** (T16): inputs `merge_sha`, `remeasure_path`. Go verifies: the commit has two parents;
  `git rev-parse <merge>^{tree}` equals `git rev-parse <merge>^2^{tree}`; `<merge>` is reachable from
  the local integration branch (`git merge-base --is-ancestor <merge> <integration-branch>`); the
  re-measure file exists and contains the merge SHA (full or ≥12-character prefix). Records
  `merge_sha`, `merge_tree`, `remeasure_path`.
- **Push ancestry** (T17): `git merge-base --is-ancestor <merge_sha> <remote>/<integration-branch>`
  against the remote-tracking ref as it stands; F1 does not fetch (the lead fetches before deciding).
- **No-remote test** (T17 vs T18): `git remote` prints nothing.
- **CI verdict** (T19): **no F1 reader.** The source of a CI verdict (the hosting provider's check
  runs for the pushed merge commit) needs network access, provider credentials, and a policy for
  required versus optional checks — all F3 controller concerns. F1 refuses the edge rather than accept
  a caller-supplied file, which would be exactly the unverified claim REQ-FR-010 forbids.

No reader accepts a verdict, tree hash, or boolean from the caller (REQ-FR-010). The integration branch
name comes from the existing git-strategy configuration; resolving it is a run-phase detail.

## Verdict-line producer (REQ-FR-011)

The E-VERDICT gate is inert unless auditors write the two lines, so the producer is part of F1:

| Surface | Path | Change |
|---|---|---|
| plan-auditor, local | `.claude/agents/moai/plan-auditor.md` | output-format section instructs the two lines |
| plan-auditor, template | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | same (template-neutral wording) |
| sync-auditor, local | `.claude/agents/moai/sync-auditor.md` | same |
| sync-auditor, template | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | same |
| Codex emitted definitions | `internal/template/templates/.codex/agents/moai/{plan,sync}-auditor.toml` | regenerated by `make agents-emit`, never hand-edited |
| convention, local + template | `.moai/docs/audit-artifact-convention.md` and its template mirror | § What lists the two lines as required |

The line format carries no project-specific content (no SPEC ID, date, or SHA literal), so it satisfies
template neutrality.

**Guardrail (research.md R16):** the `verdict:` / `audited_sha:` lines are verdict-FILE content only —
their instruction goes in the file-export part of each auditor's output format, never near or after the
existing `AUDIT-VERDICT:` chat-message instruction (`### [HARD] Cite your audit receipt`), which must
remain the auditor's last instruction so `ParseVerdictLine` keeps finding it as the last non-empty line.

## Atomic transition contract

```
BEGIN IMMEDIATE
  SELECT state, version, lease_*, … FROM cards WHERE run_id=? AND card_id=?
  if state is lease-holding and lease expired → apply T27/T28, append lease.expired, COMMIT, return ErrLeaseExpired
  if state is legacy and target != abandoned → ErrLegacyState
  if (state → target) is T19 or T20 → ErrReservedEdge
  if (state → target) ∉ table, or its guard fails → ErrIllegalTransition
  run evidence readers (outside SQL; git subprocesses) → ErrEvidence on failure
  UPDATE cards SET state=?, version=version+1, … WHERE run_id=? AND card_id=? AND version=?expected
  rows affected == 0 → ErrStaleVersion
  INSERT INTO events(run_id, kind='card.transition', payload_json={card_id, from, to, version, actor, evidence})
COMMIT
```

Lease expiry applies only to lease-holding states, so a card in `kickoff` or `needs-decision` never
reaches the first branch. The `WHERE version = expected` clause is the concurrency control; the
immediate transaction lock (`_txlock=immediate`, already set in `OpenFactoryPath`) serializes writers
across processes. Git reads run before the `UPDATE`; if the row changed meanwhile, the version compare
rejects the write and no state moved on stale evidence.

## Commands

| Command | Effect | Writes |
|---|---|---|
| `moai factory assign <card> [--to <label>] [--prefer k=v] [--after <card>] [--spec <SPEC-ID>] [--worktree <path>] [--contract-ref <spec>,<sha256>,<signed_at>[,<event>]] [--run <id>]` | T1 (create at `picked` or update fields while `picked`), then T2 when `--to` | yes |
| `moai factory status [--run <id>] [--json]` | report rows; mark expired leases `expired` | **no** |
| `moai factory decide <card>... --gate kickoff --choice approve\|reject` | T8 / T9 per card | yes |
| `moai factory decide <card>... --gate push` | T17 or T18 per card | yes |
| `moai factory decide <card> --choice resume\|block\|unblock\|abandon` | T22, T23, T24, T25 | yes |

`decide` refuses `--decider` values other than `human`. Each card in a batch is its own transaction; a
refusal on one card is reported and does not roll back the others. `status --json` output is the
machine interface F2/F3 read.

## Wiring the existing dispatch paths (REQ-FR-025)

`gtd.go:249` (queue dispatch) and `goal.go:864` (auto-mission dispatch) each call
`kanban.RecordFactoryCardAssignment` (queue runtime report). After that call succeeds, the same dispatch
records T1+T2 in factory.db for `(runID, cardID, lane)`. The queue runtime write, the idempotency
readback, and the message delivery are unchanged. A factory.db failure does not fail the dispatch
(plan.md §C.5, lead condition) and leaves three traces:

1. a `FACTORY_RECORD_UNAVAILABLE` line on stderr;
2. one JSON line appended to `ProjectDir/factory/record-unavailable.jsonl` (the same directory as
   factory.db, outside the database so it survives a database write failure):
   `{at, run_id, card_id, lane, error, reconciled: false}`;
3. `moai factory status` reads that file (read-only) and reports every entry with
   `reconciled: false`, in text and in JSON.

Reconciliation: at the start of the next successful transition-API write for the same run, the store
reads the unreconciled entries for that run and, inside the same transaction as that write, appends one
`record.drift` event per entry with payload `{card_id, lane, dispatched_at, factory_state}` —
`factory_state` is the card's current state in factory.db, or `absent`. After the transaction commits,
each entry is rewritten with `reconciled: true` (write-to-temp and rename). An entry already reconciled
never produces a second event. If the directory itself is unwritable, only trace 1 exists — a residual
risk stated in `spec.md` §E.

`todo.go:1008` (`recordFactoryCardState`, e.g. unpick) stays queue-report-only in F1: lane stage
reporting moves in F2.

## Retiring `RecordCard`

`RecordCard` (`runtime.go:53-84`) writes `cards` without a version compare and has no production caller
(research.md R3). Leaving it would keep a second, unguarded writer available. Proposed disposition in
plan.md §C.6.
