# Card Review — t1497 (card-scope codex review)

- Reviewer: codex (adversarial, via the session MCP codex_audit; card scope,
  base d05d1d5f0..HEAD), reviewed at `WT-harness-prune-detached@d0a153cdc`
- Verdict: fail — 1 finding (P1, high confidence), reproduced by the
  reviewer's own probe script (`/tmp/t1497-review-owned-probe.py`); the
  repository was not modified by the probe.

## P1 — the detached child has no exit deadline; waiting processes can accumulate

`internal/harness/retention_spawn_unix.go:24`: the child runs a blocking
flock plus archive reads with no execution deadline. Reviewer probe: while
the queue lock was held, all four spawned children were still alive after 6s;
against an archive that is a FIFO, the wait persisted even after the stamp
write; four gate calls on a stale stamp produced four spawn requests.

Recommended repair (reviewer): non-blocking or time-bounded lock acquisition
in the child; an overall execution deadline; reject special files for the
archive.

## Disposition (lane)

Recorded as a NAMED RESIDUAL + follow-up card candidate for the leader, not
a repair round in this card: (a) the independent sync-audit recorded zero
blocking code findings and did not raise it; (b) under normal conditions the
children terminate once the lock holder finishes and stamps (the fresh stamp
then suppresses later gates); (c) the unbounded cases require exotic local
conditions (archive replaced by a FIFO) or a lock held for the entire prune
of a very large log — a robustness hardening (LOCK_NB in the child path +
special-file rejection + optional deadline) rather than a correctness defect
in the card's contract. The SPEC's REQ-DP-005 "bounded by the prune itself"
should be read with this qualification.

🗿 MoAI

## Addendum — turn-end gate re-flag (P2, HEAD 176578494, high confidence)

The turn-end codex gate reproduced the accumulation at the final HEAD: with
the lock held, 3 spawned children survived 6s (`detached_utility_children_
after_6s 3`); all exited normally after release.

Sharpened analysis (lane): the production accumulation window is narrower
than the repro suggests — the lock holder is a real pruner and the retention
machinery stamps BEFORE work (stamp-before-work, retention.go), so once the
first child acquires the lock it stamps immediately and every later gate call
reads a fresh stamp (no further spawns). The repro held a lock WITHOUT a
stamping holder, which is not the production shape. The genuinely unbounded
case remains the archive-as-FIFO hazard. Disposition unchanged: follow-up
card candidate (child LOCK_NB + special-file rejection + optional deadline),
leader adjudicates whether it lands as an in-place amendment of this
completed SPEC or a new card.

🗿 MoAI
