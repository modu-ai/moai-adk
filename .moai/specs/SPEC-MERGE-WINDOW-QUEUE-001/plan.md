# SPEC-MERGE-WINDOW-QUEUE-001 — Plan

> Card t1479 · Tier L · development mode per `.moai/config/sections/quality.yaml`. Milestones are
> ordered by decision reversibility: data model and new interfaces first, mechanical doc and
> regression work last. Priority labels only, no time estimates. Target release v3.2.0.

## §A Context

See spec.md §A and research.md §R1-§R5. The operator approved FIFO self-acquire and the abolition
of leader nomination on 2026-10-03 (decision-index Q1). Leader decisions (mission contract
07d28c4b): Q2-Q16 across iterations, then Q17 (v0.5.0 scope reduction to plain FIFO) and Q18 (lane
merge verb).

## §B Known issues carried in

- The in-window re-measure duration figure is leader-measured, not re-measured here (spec.md §F).
- SPEC-CANDIDATE-CI-001 (t1478) is not landed; `workflow.candidate_ci.enabled` absent reads as
  false, so the local record form governs and the landing check is a no-op until it lands. Its
  REQ-CCI-004 branch resolution is a dependency of the merge verb (spec.md §F).
- Heartbeat 15 s and window 60 s are leader values (Q14); M0 may only tighten them.

## §C Pre-flight (run phase)

1. Re-read `git rev-parse --short HEAD`, `git branch --show-current`; absorb local develop.
2. Re-run research.md §R1 cells on the run tree; any cell already GREEN means another card landed
   part of this scope — stop and report.
3. Check whether t1478 has landed; if so, consume its REQ-CCI-004 resolver and REQ-CCI-011 landing
   check; if not, implement the resolver behind the same contract and leave the landing check as
   an absent no-op seam (research.md §R5).
4. No open decision remains; re-read decision-index for any later row.
5. The no-`--wait` baseline is already committed (`3bc274dac`,
   `.moai/reports/t1479/baseline-acquire-nowait/`); do not regenerate it after code changes.

## §D Constraints

- Every queue mutation inside the existing mutation section; no new lock file.
- Additive optional record fields only; outputs without new flags byte-identical to the fixture.
- Template-first for the distributed rule text; `make build` after template edits.
- Lane-local verification scoped to touched packages (`internal/kanban`, `internal/cli`,
  `internal/homestate`, `internal/factorylane`, `internal/hook`, `internal/config`), `-race` on the
  queue/promotion code; no local full suite.
- A Bash tool call is capped at 10 minutes, below the 60-minute default wait: lanes run
  `acquire --wait` as a background command (stated in the M4 doctrine).

## §E Self-verification

Each milestone exits on its acceptance.md ACs, measured on the milestone HEAD, recorded in
progress.md §E.2.

## §F Milestones

### M0 — In-window duration baseline (Priority High, measurement only)
- Manual simulation of the new in-window path on a fixture repository (a scratch clone with a card
  branch that already absorbed develop): time the tree-identity check + `git merge --no-ff` +
  `git rev-parse <merge>^{tree}` comparison, N = 20 runs, reporting median and maximum wall time.
  Output `.moai/reports/t1479/m0-window-duration.md` with the command and raw timings, committed
  BEFORE the M1 code commit (verification-claim-integrity §2.3).
- The result may only LOWER the 30-minute lease default and tighten the 15 s / 60 s liveness values.
- ACs: AC-MWQ-008 (M0 evidence clause).

### M1 — Window record data model: FIFO queue, owner pid on tickets, lease, policy (Priority High)
- Additive `queue[]` (owner pid + `pid_source`, waiter pid + start time, heartbeat),
  `lease_expires_at`; policy sibling record; legacy-record read compatibility.
- ACs: AC-MWQ-001, -008, -023.

### M2 — Re-measure record and the executing verb (Priority High)
- `moai integration remeasure -- <command>` (name final at run): clean-tree and unchanged-HEAD
  checks before and after, captures command + exit code, recognized-report count, empty-sweep and
  unstructured-run refusal, build identity, key = `HEAD^{tree}` + absorbed base; one verifier for
  the local and candidate-CI forms selected by `workflow.candidate_ci.enabled`.
- ACs: AC-MWQ-014, -015, -016.

### M3 — acquire --wait, liveness drop, promotion, status (Priority High)
- FIFO enqueue, bound (bare 60m), heartbeat refresh, owner/waiter/heartbeat drop,
  timeout-vs-promotion ordering, promotion with owner-pid stamping on release / dead owner / lease
  expiry, lease renewal, no-`--wait` refusal while a ticket is queued, status human + JSON.
- ACs: AC-MWQ-002 … -006, -009, -010, -011.

### M4 — Policy verb and the nomination doctrine (Priority High)
- `moai integration policy open|hold`; lane-role refusal; hold suspends every promotion.
- AGENTS.local.md §4.1 (line 221, the lane window procedure, the background-wait note) and
  gitflow-lane-protocol.md §3/§6 carry the two REQ-MWQ-013 sentences.
- ACs: AC-MWQ-007, -012, -013.

### M5 — Lane merge verb (Priority High)
- `moai integration merge --card <id>`: holder check first (read only; a non-holder or an
  expired-lease holder is refused with the record unchanged), then lease renewal and drops; branch
  resolution, SHA pinned once; pre-merge checks in order — record validity, base equals tip, SHA
  descends from base, tree identity, landing check — then `git merge --no-ff <sha>`, merge-tree
  verification, release; nine causes with distinct exit codes; merge failure → `git merge --abort` +
  clean-worktree check, else `hold`; any failure after the merge commit exists leaves the commit,
  sets `hold` naming the SHA, releases.
- ACs: AC-MWQ-017, -018.

### M6 — Substantive completion gate on the one merge path (Priority High)
- `factory complete` in the REQ-MWQ-019 order, every gate before develop moves: (1) card gates read
  without transitioning (merge-ready, own unexpired card lease, version); (2) adoption only of a
  merge commit whose second parent is the branch's current tip and whose tree matches a valid
  record; (3) refusal without a valid record for the current candidate tree; (4) the M5 step
  replacing its own merge (`factory_card.go:1409`), with the merging → merged-local transitions only
  after it succeeds. The stand-in never satisfies the gate; `verifyMerge` structural check;
  merge-readiness fourth condition with the command printed.
- ACs: AC-MWQ-019, -020, -021.

### M7 — Distributed and local doctrine text (Priority Medium)
- Template `kanban-dispatch-mechanics.md` § Integration (drop the announcement; describe queue +
  policy + `acquire --wait` + `integration merge`) → `make build` → local mirror;
  `.moai/docs/gitflow-integration-chain.md` window bash re-ordered (re-measure before acquire,
  merge through the verb).
- ACs: AC-MWQ-022.

### M8 — Regression sweep (Priority Low)
- Guard, session-end automerge, existing integration/factory tests unchanged; `-race` on M3.
- ACs: AC-MWQ-010, -023, quality gates in acceptance.md §G.

## §G Risks

| Risk | Mitigation |
|---|---|
| Promoted holder read as stale once its waiter exits | owner-session pid stamped on promotion (REQ-MWQ-006, AC-MWQ-006 scenario 2) |
| Waiter killed → orphan ticket promoted, window wedged | ticket dropped when owner or waiter is gone or heartbeat stale (REQ-MWQ-003) |
| Promotion races the bound | both decided in the mutation; late-observed promotion releases at once (REQ-MWQ-005) |
| Bash 10-min cap below the 60-min wait | background `acquire --wait`; an exited waiter re-enqueues at the tail (accepted) |
| Lease evicts a merging holder | 30-min default far above a seconds-long in-window step; renewals; M0 may only shorten |
| Repeated invalidation sends a lane to the tail each time | accepted with the scope reduction; residual risk in research.md §R7 |
| Release during hold empties the window | intended: queue intact, promotion resumes on `open` (REQ-MWQ-007) |
| Empty or unstructured test runs pass the gate | refused by REQ-MWQ-015 |
| Dirty tree measured under HEAD's key | refused by REQ-MWQ-016 |
| A failed in-window step wedges the window or leaves a half-merge | every failure releases with a distinct code; abort + clean check, else `hold` (REQ-MWQ-018) |
| Complete and the merge verb diverge into two merge paths | complete calls the verb's step and adopts a prior landing (REQ-MWQ-019) |
| Promoted holder without an integration target | ticket carries `branch` / `branch_source` / `worktree`, copied on promotion (REQ-MWQ-001/006) |
| t1478 not landed when M5 runs | resolver behind the same contract, landing check absent no-op; reconciled by the second card |
| Hook guard semantics drift | REQ-MWQ-023 regression on holder-only records |

## §H Cross-references

spec.md · acceptance.md · design.md · research.md · decision-index.md · progress.md ·
`.moai/reports/t1479/plan-audit-iter1.md` · `.moai/reports/t1479/plan-audit-iter2.md` ·
SPEC-CANDIDATE-CI-001
