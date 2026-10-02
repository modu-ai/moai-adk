# t1425 decision records

card: t1425 (urgent, class B, leader-dispatched, no SPEC; `/moai run` with plan skipped, card text is the scope)
deadline: develop landing before 2026-10-04T09:18Z (KST 18:18); after that the oldest remaining event passes 30 days and the rewrite storm can recur

decision record: decided_by=lane-6 orchestrator (card t1425) with Jev jev-1.13.0 under the operator's standing delegation of design judgments to Jev evidence_refs=scratchpad t1425/state.txt + questions.json + jev-answer.json (answers: where_prune_runs=gate_in_hook conf 0.96, lock_behavior=block_then_recheck conf 1.00), internal/harness/retention.go:59-99, internal/harness/observer.go:93 and :141, internal/cli/hook.go:822/919/1056/1220, internal/lockfile/lockfile_unix.go ladder_path=leader dispatch rule "in-flight decisions go to Jev, confidence under 0.5 / push / PR / deletion go to the leader" (both answers above the 0.5 gate; neither is a completion, merge or queue judgment)

## What was asked and what Jev was told

The state text given to Jev lists seven facts the lane measured by reading the tree (stat of the live log size: 65,506,619 bytes; no figure carried over from the card text), not the card's prose. Questions: (1) keep pruning on the hook path behind an on-disk stamp and lock, or move it off the hook path; (2) a process that finds the lock held blocks then re-checks the stamp, or returns at once (which needs a new non-blocking try-lock in `internal/lockfile`).

## Choice taken

- Keep the lazy prune in `RecordEvent` / `RecordExtendedEvent`, gated by an on-disk stamp that stores the time returned by `nowFn` (tests inject a mock clock, so file mtimes are unusable) and an exclusive `lockfile.Lock` so one process prunes per interval.
- Lock waiters block, then re-read the stamp and return without pruning. No change to `internal/lockfile`.

## Known limits recorded up front

- On Windows `internal/lockfile` is an in-process mutex, not a cross-process lock; the single-pruner guarantee holds on Unix only. The stamp still bounds repeated rewrites to the interval there.
- A single pruner still reads the whole log and replaces it by rename: events appended between its read and its rename are lost. That race exists today (it carries an `@MX:WARN`); this card reduces it from N concurrent rewriters to one per interval and does not remove it.
- `partitionEvents` drops lines that fail JSON parsing although its comment says they are kept (data loss for malformed lines). Out of scope here; reported to the leader.
