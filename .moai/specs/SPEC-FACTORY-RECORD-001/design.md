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
```

Layers 2-5 (controller, agent sessions, notifications, Decider) are F2/F3.

## Schema — `cards` columns after migration (schema version 4)

Existing columns keep their names and meaning: `run_id`, `card_id`, `owner_label`, `state`,
`version`, `evidence_path`, `updated_at`.

| Column (proposed) | Type | Default | Meaning |
|---|---|---|---|
| `stage` | TEXT | `''` | last lease-holding working state reached (`plan` … `merge-ready`) |
| `lease_holder` | TEXT | `''` | worker label holding the lease; empty when none |
| `lease_expires_at` | TEXT | `''` | RFC 3339; empty when no lease |
| `heartbeat_at` | TEXT | `''` | last renewal by the holder |
| `decision_gate` | TEXT | `''` | `kickoff`, `push`, or `question` while a decision is pending |
| `decision_question` | TEXT | `''` | operator-facing text for `question` |
| `decision_resume` | TEXT | `''` | state to resume after `needs-decision` |
| `decider` | TEXT | `''` | `human` in F1; F3 adds `llm`, `jev` |
| `decided_at` | TEXT | `''` | time of the last decision |
| `hint_prefer` | TEXT | `''` | verbatim `key=value`, e.g. `backend=codex` |
| `hint_after` | TEXT | `''` | predecessor card id |
| `spec_id` | TEXT | `''` | SPEC identifier once one exists |
| `worktree_path` | TEXT | `''` | absolute path of the card worktree |
| `evidence_sha` | TEXT | `''` | commit SHA recorded at the last audit entry |
| `merge_sha` | TEXT | `''` | local merge commit |
| `merge_tree` | TEXT | `''` | tree hash of the merge commit |
| `remeasure_path` | TEXT | `''` | re-measure evidence file |
| `contract_spec_id` | TEXT | `''` | A1 pointer field |
| `contract_sha256` | TEXT | `''` | A1 pointer field |
| `contract_signed_at` | TEXT | `''` | A1 pointer field |

All new columns are `TEXT NOT NULL DEFAULT ''`, so `ALTER TABLE ... ADD COLUMN` needs no backfill and
every v3 row stays valid. No CHECK constraint is added to `state` (SQLite cannot alter one later, the
same reason the queue schema is frozen — research.md R5); the enum is enforced by the transition API.

`workers` gains no column. Heartbeat renewal writes the existing `workers.heartbeat_at`.

Migration: `migrateFactoryV3ToV4`, following `migrateFactoryV2ToV3` (research.md R2) — read
`PRAGMA table_info(cards)`, add only missing columns, `UPDATE meta SET value='4' WHERE
key='schema_version' AND value='3'`, all in one transaction, appended to the existing chain in
`OpenFactoryPath`.

## States

| Group | States |
|---|---|
| Admission | `picked`, `assigned` |
| Lease-holding (working) | `leased`, `plan`, `plan-audit`, `kickoff`, `run`, `sync`, `sync-audit`, `merge-ready`, `merging` |
| Post-merge | `merged-local`, `pushed`, `ci-green` |
| Paused | `needs-decision`, `blocked` |
| Terminal | `done`, `failed`, `abandoned` |

A row whose `state` is not one of these (a pre-v4 row) is a **legacy** row: `status` shows it, and the
only accepted transition is an operator `decide --choice abandon` (REQ-FR-003).

## Transition Table

Edges not listed are illegal (REQ-FR-007). "Gate" names what Go verifies before committing.

| # | From | To | Actor | Gate |
|---|---|---|---|---|
| T1 | (none) | `picked` | `assign` | queue item for the card id is `picked` (REQ-FR-022) |
| T2 | `picked` | `assigned` | `assign --to` | `after` predecessor merged (REQ-FR-016) |
| T3 | `assigned` | `leased` | worker (F2 verb / API) | holder registered in `workers` and equal to `owner_label` (REQ-FR-012) |
| T4 | `leased` | `plan`, `run`, or the recorded `stage` | holder | lease valid |
| T5 | `plan` | `plan-audit` | holder | E-ENTRY (REQ-FR-008) |
| T6 | `plan-audit` | `plan` | holder | E-VERDICT any verdict, audited SHA matches |
| T7 | `plan-audit` | `kickoff` | holder | E-VERDICT PASS / PASS-WITH-DEBT (REQ-FR-009); sets `decision_gate=kickoff` |
| T8 | `kickoff` | `run` | `decide --gate kickoff --choice approve` | decider `human` (REQ-FR-018) |
| T9 | `kickoff` | `blocked` | `decide --gate kickoff --choice reject` | decider `human` |
| T10 | `run` | `sync` | holder | supplied commit SHA resolves and is ancestor-or-equal of worktree HEAD |
| T11 | `sync` | `sync-audit` | holder | E-ENTRY |
| T12 | `sync-audit` | `sync` | holder | E-VERDICT any verdict |
| T13 | `sync-audit` | `merge-ready` | holder | E-VERDICT PASS / PASS-WITH-DEBT |
| T14 | `merge-ready` | `merging` | holder (F3 controller in autonomous mode) | lease valid |
| T15 | `merging` | `merge-ready` | holder | none (merge abandoned before commit) |
| T16 | `merging` | `merged-local` | holder / controller | E-MERGE (REQ-FR-010) |
| T17 | `merged-local` | `pushed` | `decide --gate push` (lead batch) | merge SHA ancestor of remote-tracking integration ref (REQ-FR-020) |
| T18 | `merged-local` | `done` | `decide --gate push` | repository has no remote; event note `no remote — no CI verdict` (REQ-FR-021) |
| T19 | `pushed` | `ci-green` | API (producer is F3) | CI evidence file names the pushed merge SHA with a success conclusion |
| T20 | `ci-green` | `done` | API / lead | none |
| T21 | any non-terminal, non-paused | `needs-decision` | any actor | question text required; lease cleared |
| T22 | `needs-decision` | resume state, or `assigned` (stage kept) when the resume state is lease-holding | `decide --choice resume` | human |
| T23 | `needs-decision` | `blocked` | `decide --choice block` | human |
| T24 | `blocked` | `assigned` | `decide --choice unblock` | human; stage kept |
| T25 | any non-terminal | `abandoned` | `decide --choice abandon` | human |
| T26 | any lease-holding | `failed` | holder / API | reason text required |
| T27 | lease-holding except `merging`, lease expired | `assigned` | automatic, before any other transition (REQ-FR-014) | expiry passed |
| T28 | `merging`, lease expired | `blocked` | automatic (REQ-FR-015) | expiry passed |

`leased → <stage>` (T4) is what makes a reclaimed card resume where it stopped: T27 keeps `stage`,
so the next lease holder re-enters that stage instead of restarting at `plan`.

## Evidence readers (Go reads; callers supply only paths and SHAs)

- **E-ENTRY** (T5, T11): inputs `sha`, `artifact_path`. Go runs, in the card's `worktree_path`:
  `git cat-file -e <sha>^{commit}`; `git merge-base --is-ancestor <sha> HEAD` (exit 0);
  `git cat-file -e <sha>:<artifact_path>`. On success it records `evidence_sha = sha`,
  `evidence_path = artifact_path`.
- **E-VERDICT** (T6, T7, T12, T13): Go reads the newest file matching
  `.moai/reports/<card-id>/{plan,sync}-audit*.md` and extracts two machine-readable lines (format per
  plan.md §C.4): `verdict: <PASS|PASS-WITH-DEBT|FAIL>` and `audited_sha: <sha>`. The audited SHA must
  equal `evidence_sha` (prefix of at least 12 hex characters accepted). A missing file, a missing line,
  or a mismatch refuses the transition.
- **E-MERGE** (T16): inputs `merge_sha`, `remeasure_path`. Go verifies: the commit has two parents;
  `git rev-parse <merge>^{tree}` equals `git rev-parse <merge>^2^{tree}`; `<merge>` is reachable from
  the local integration branch (`git merge-base --is-ancestor <merge> <integration-branch>`); the
  re-measure file exists and contains the merge SHA (full or ≥12-character prefix). Records
  `merge_sha`, `merge_tree`, `remeasure_path`.
- **Push ancestry** (T17): `git merge-base --is-ancestor <merge_sha> <remote>/<integration-branch>`
  against the remote-tracking ref as it stands; F1 does not fetch (the lead fetches before deciding).
- **No-remote test** (T18): `git remote` prints nothing.

No reader accepts a verdict, tree hash, or boolean from the caller (REQ-FR-011). The integration branch
name comes from the existing git-strategy configuration (`develop` in this repository); resolving it is
a run-phase detail.

## Atomic transition contract

```
BEGIN IMMEDIATE
  SELECT state, version, lease_*, … FROM cards WHERE run_id=? AND card_id=?
  if lease expired and state lease-holding → apply T27/T28, append lease.expired, COMMIT, return ErrLeaseExpired
  if state is legacy and target != abandoned → ErrLegacyState
  if (state → target) ∉ table → ErrIllegalTransition
  run gate readers (outside SQL; git subprocesses) → ErrEvidence on failure
  UPDATE cards SET state=?, version=version+1, … WHERE run_id=? AND card_id=? AND version=?expected
  rows affected == 0 → ErrStaleVersion
  INSERT INTO events(run_id, kind='card.transition', payload_json={card_id, from, to, version, actor, evidence})
COMMIT
```

The `WHERE version = expected` clause is the concurrency control; the immediate transaction lock
(`_txlock=immediate`, already set in `OpenFactoryPath`) serializes writers across processes. Git reads
run before the `UPDATE`; if the row changed meanwhile, the version compare rejects the write and no
state moved on stale evidence.

Returning `ErrLeaseExpired` after committing T27/T28 is deliberate: the caller asked for a transition
from a state the card is no longer in, and must re-read.

## Commands

| Command | Effect | Writes |
|---|---|---|
| `moai factory assign <card> [--to <label>] [--prefer k=v] [--after <card>] [--spec <SPEC-ID>] [--worktree <path>] [--contract-ref spec,sha256,signed_at] [--run <id>]` | T1 (create at `picked` or update hints/fields while `picked`), then T2 when `--to` | yes |
| `moai factory status [--run <id>] [--json]` | report rows; mark expired leases `expired` | **no** |
| `moai factory decide <card>... --gate kickoff --choice approve\|reject` | T8 / T9 per card | yes |
| `moai factory decide <card>... --gate push` | T17 or T18 per card | yes |
| `moai factory decide <card> --choice resume\|block\|unblock\|abandon` | T22-T25 | yes |

`decide` refuses `--decider` values other than `human` (the flag exists so F3 extends it rather than
adding a second flag). Each card in a batch is its own transaction; a refusal on one card is reported
and does not roll back the others. `status --json` output is the machine interface F2/F3 read.

## Wiring the existing dispatch paths (REQ-FR-025)

`gtd.go:249` and `goal.go:864` call `kanban.RecordFactoryCardAssignment` (queue runtime report). After
that call succeeds, the same dispatch records T1+T2 in factory.db for `(runID, cardID, lane)`. The
queue runtime write, the idempotency readback, and the message delivery are unchanged. A factory.db
failure behaves per plan.md §C.5.

`todo.go:1008` (`recordFactoryCardState`, e.g. unpick) stays queue-report-only in F1: lane stage
reporting moves in F2.

## Retiring `RecordCard`

`RecordCard` (`runtime.go:53-84`) writes `cards` without a version compare and has no production caller
(research.md R3). Leaving it would keep a second, unguarded writer available. Proposed disposition in
plan.md §C.6.
