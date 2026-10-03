# SPEC-MERGE-WINDOW-QUEUE-001 — Plan

> Card t1479 · Tier L · development mode per `.moai/config/sections/quality.yaml`. Milestones are
> ordered by decision reversibility: data model and new interfaces first, mechanical doc and
> regression work last. Priority labels only, no time estimates.

## §A Context

See spec.md §A and research.md §R1-§R3. The operator approved FIFO self-acquire and the abolition
of leader nomination on 2026-10-03 (decision-index.md Q1).

## §B Known issues carried in

- The re-measure duration figure is leader-measured, not re-measured here (spec.md §F).
- t1478 not landed; candidate CI evidence is an optional input (research.md §R5).

## §C Pre-flight (run phase)

1. Re-read `git rev-parse --short HEAD`, `git branch --show-current`; absorb local develop.
2. Re-run research.md §R1 cells E1-E10 on the run tree; any cell already GREEN means another card
   landed part of this scope — stop and report.
3. Check whether t1478 has landed (`git log --oneline develop | grep -i t1478`); record the answer
   in progress.md — it selects the candidate source for M2.
4. Decision-index rows Q2-Q6 carry the leader's disposition (spec.md v0.2.0); no open decision
   blocks M1/M2. Target release: v3.2.0 (operator, 2026-10-03).

## §D Constraints

- Every queue mutation inside the existing mutation section; no new lock file.
- Additive optional record fields only; outputs without new flags byte-identical.
- Template-first for the distributed rule text; `make build` after template edits.
- Lane-local verification scoped to touched packages (`internal/kanban`, `internal/cli`,
  `internal/homestate`, `internal/factorylane`, `internal/hook`, `internal/config`), `-race` on the
  queue/promotion code; no local full suite.

## §E Self-verification

Each milestone exits on its acceptance.md ACs, measured on the milestone HEAD, recorded in
progress.md §E.2.

## §F Milestones

### M0 — In-window duration baseline (Priority High, measurement only)
- Measure the in-window duration under the new order (identity check + `--no-ff` merge, no tests)
  on the run tree; record command + output in progress.md §E.2. The result may only LOWER the
  30-minute lease default (REQ-MWQ-009); it never raises it.
- Commit the baseline before the M1 code it informs (verification-claim-integrity §2.3).

### M1 — Window record data model: queue, lease, policy (Priority High)
- Additive `queue[]` (with the one-time front-reservation marker), `lease_expires_at`; lease on by
  default at 30 min, zero disables; policy sibling record; legacy-record read compatibility.
- ACs: AC-MWQ-001, -009, -009a, -010, -051.

### M2 — Re-measure record and the executing verb (Priority High)
- `moai integration remeasure -- <command>` (name final at run): runs the command in the caller's
  worktree, captures command + exit code, keys the record by `HEAD^{tree}` with the absorbed base;
  records a test count only from a recognized structured report (`go test -json` here), zero
  reported = invalid; one verifier for the local and candidate-CI forms, the form required being
  selected by t1478's candidate-CI setting (off/absent → local).
- ACs: AC-MWQ-020, -021, -021a, -021b.

### M3 — acquire --wait, promotion, status view (Priority High)
- FIFO enqueue, bounded wait (bare `--wait` = 60 min), promotion on release / stale PID / lease
  expiry, lease renewal on every holder window verb, dead-ticket pruning, status human + JSON.
- ACs: AC-MWQ-002 … -008, -005a.

### M4 — Policy verb and the nomination doctrine (Priority High)
- `moai integration policy open|hold`; lane-role refusal; acquire honours hold.
- AGENTS.local.md §4.1 (line 221 and the lane window procedure), gitflow-lane-protocol.md §3/§6.
- ACs: AC-MWQ-011, -012, -013.

### M5 — Substantive completion gate (Priority High)
- `factory complete` refuses without a keyed record; stand-in no longer satisfies the gate;
  `verifyMerge` structural check; merge-readiness fourth condition.
- ACs: AC-MWQ-030 … -033.

### M6 — In-window path: identity-only + requeue on base move (Priority Medium)
- Lane procedure and `factory complete` in-window steps reduced to identity check + `--no-ff`;
  base-moved → release + requeue report; requeued holder keeps a front slot once, a second
  consecutive move sends it to the tail.
- ACs: AC-MWQ-022, -023, -024.

### M7 — `moai integration push` (Priority Medium)
- Threshold, red-CI hold, fail-closed CI read, lane refusal, window-held refusal, no force,
  pre-push read, post-push landing report. Reuses the existing gh conclusion mapping.
- ACs: AC-MWQ-040 … -044.

### M8 — Distributed and local doctrine text (Priority Medium)
- Template `kanban-dispatch-mechanics.md` § Integration (drop the announcement layer; describe
  queue + policy) → `make build` → local mirror; gitflow-lane-protocol.md §4 points at the verb;
  `.moai/docs/gitflow-integration-chain.md` window bash re-ordered.
- ACs: AC-MWQ-050.

### M9 — Regression sweep (Priority Low)
- Guard, session-end automerge, existing integration/factory tests unchanged; `-race` on M3.
- ACs: AC-MWQ-007, -051, quality gates in acceptance.md §G.

## §G Risks

| Risk | Mitigation |
|---|---|
| A waiter dies after promotion → window wedged | promotion target is checked for liveness on every observation (D2) |
| Lease evicts a merging holder | 30-min default far above the identity+merge path; every holder window verb renews; M0 may only shorten (Q2) |
| Base moves often under load → requeue churn / starvation | one-time front reservation, second consecutive move → tail (Q4) |
| Test-count parsing is language-specific in a distributed verb | count only from a recognized structured report; other runners record command + exit (Q6) |
| t1478 setting not landed when M2 runs | absent setting reads as off → local form required (Q5) |
| Hook guard semantics drift | REQ-MWQ-051 + AC-MWQ-051 regression on holder-only records |

## §H Cross-references

spec.md · acceptance.md · design.md · research.md · decision-index.md · progress.md
