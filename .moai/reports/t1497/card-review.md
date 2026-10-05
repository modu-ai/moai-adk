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
