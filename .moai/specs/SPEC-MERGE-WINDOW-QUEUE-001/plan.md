# SPEC-MERGE-WINDOW-QUEUE-001 — Plan

> Card t1479 · Tier L · development mode per `.moai/config/sections/quality.yaml`. Milestones are
> ordered by decision reversibility: data model and new interfaces first, mechanical doc and
> regression work last. Priority labels only, no time estimates. Target release v3.2.0.

## §A Context

See spec.md §A and research.md §R1-§R5. The operator approved FIFO self-acquire and the abolition
of leader nomination on 2026-10-03 (decision-index Q1); the leader settled Q2-Q6 and the
plan-audit iteration 1 decisions Q8-Q13 (mission contract 07d28c4b).

## §B Known issues carried in

- The in-window re-measure duration figure is leader-measured, not re-measured here (spec.md §F).
- SPEC-CANDIDATE-CI-001 (t1478) is not landed; `workflow.candidate_ci.enabled` absent reads as
  false, so the local record form governs until it lands (REQ-MWQ-015).
- Heartbeat 15 s, heartbeat window 60 s, re-entry grace 120 s are leader values (Q14); M0 may only
  tighten them.

## §C Pre-flight (run phase)

1. Re-read `git rev-parse --short HEAD`, `git branch --show-current`; absorb local develop.
2. Re-run research.md §R1 cells on the run tree; any cell already GREEN means another card landed
   part of this scope — stop and report.
3. Check whether t1478 has landed and whether its key path is still `workflow.candidate_ci.enabled`;
   whichever card lands second verifies the shared path (research.md §R5).
4. No open decision remains (Q14-Q16 settled in v0.4.0); re-read decision-index for any later row.
5. The no-`--wait` baseline is already committed (`3bc274dac`,
   `.moai/reports/t1479/baseline-acquire-nowait/`); do not regenerate it after code changes.

## §D Constraints

- Every queue mutation inside the existing mutation section; no new lock file.
- Additive optional record fields only; outputs without new flags byte-identical to the fixture.
- Template-first for the distributed rule text; `make build` after template edits.
- Lane-local verification scoped to touched packages (`internal/kanban`, `internal/cli`,
  `internal/homestate`, `internal/factorylane`, `internal/hook`, `internal/config`), `-race` on the
  queue/promotion code; no local full suite.
- A Bash tool call is capped at 10 minutes, below the 60-minute default wait. Lanes run
  `acquire --wait` as a background command, or loop `acquire --wait --slice <d>` with slices under
  the cap; REQ-MWQ-003 keeps the position across slices. The lane doctrine (M4) states this.

## §E Self-verification

Each milestone exits on its acceptance.md ACs, measured on the milestone HEAD, recorded in
progress.md §E.2.

## §F Milestones

### M0 — In-window duration baseline (Priority High, measurement only)
- Manual simulation of the new in-window path on a fixture repository (a scratch clone with a card
  branch that already absorbed develop): time `git merge-tree --write-tree` identity check +
  `git merge --no-ff` + `git rev-parse <merge>^{tree}` comparison, N = 20 runs, reporting median and
  maximum wall time. Output `.moai/reports/t1479/m0-window-duration.md` with the command and raw
  timings, committed BEFORE the M1 code commit (verification-claim-integrity §2.3).
- The result may only LOWER the 30-minute lease default (REQ-MWQ-009).
- ACs: AC-MWQ-009 (M0 evidence clause).

### M1 — Window record data model: queue, ticket liveness, lease, policy (Priority High)
- Additive `queue[]` (state, waiter pid + start time, heartbeat), `lease_expires_at`, reserved
  ticket readiness bound; policy sibling record; legacy-record read compatibility.
- ACs: AC-MWQ-001, -009, -025.

### M2 — Re-measure record and the executing verb (Priority High)
- `moai integration remeasure -- <command>` (name final at run): clean-tree and unchanged-HEAD
  checks before and after, captures command + exit code, recognized-report count, empty-sweep and
  unstructured-run refusal, build identity, key = `HEAD^{tree}` + absorbed base; one verifier for
  the local and candidate-CI forms selected by `workflow.candidate_ci.enabled`.
- ACs: AC-MWQ-015, -016, -017.

### M3 — acquire --wait, slices, liveness drop, promotion, status (Priority High)
- FIFO enqueue, total bound (bare 60m), `--slice` with still-queued exit and same-position resume,
  heartbeat refresh, dead/stale/expired ticket drop, timeout-vs-promotion ordering, promotion on
  release / dead owner / lease expiry, lease renewal, status human + JSON.
- ACs: AC-MWQ-002 … -007, -010, -011, -012.

### M4 — Policy verb and the nomination doctrine (Priority High)
- `moai integration policy open|hold`; lane-role refusal; hold suspends every promotion.
- AGENTS.local.md §4.1 (line 221, the lane window procedure, the background/slice wait note) and
  gitflow-lane-protocol.md §3/§6 carry the REQ-MWQ-014 sentence.
- ACs: AC-MWQ-008, -013, -014.

### M5 — Substantive completion gate (Priority High)
- `factory complete` refuses without a keyed record; the stand-in never satisfies the gate;
  `verifyMerge` structural check; merge-readiness fourth condition with the command printed.
- ACs: AC-MWQ-021, -022, -023.

### M6 — In-window path: identity-only, base-move requeue, reserved tickets (Priority Medium)
- Lane procedure and `factory complete` in-window steps reduced to identity check + `--no-ff`
  (REQ-CCI-011 landing check under candidate CI); base moved → release + reserved ticket; readiness
  bound; front-once; second move → tail.
- ACs: AC-MWQ-018, -019, -020.

### M7 — Distributed and local doctrine text (Priority Medium)
- Template `kanban-dispatch-mechanics.md` § Integration (drop the announcement; describe queue +
  policy + `acquire --wait`) → `make build` → local mirror; `.moai/docs/gitflow-integration-chain.md`
  window bash re-ordered (re-measure before acquire).
- ACs: AC-MWQ-024.

### M8 — Regression sweep (Priority Low)
- Guard, session-end automerge, existing integration/factory tests unchanged; `-race` on M3.
- ACs: AC-MWQ-011, -025, quality gates in acceptance.md §G.

## §G Risks

| Risk | Mitigation |
|---|---|
| Waiter killed → orphan ticket promoted, window wedged | ticket liveness keyed to the waiter process (pid + start time) and heartbeat; dropped at next mutation (REQ-MWQ-004) |
| Promotion races the bound | both decided in the mutation; late-observed promotion releases at once (REQ-MWQ-006) |
| Bash 10-min cap below the 60-min wait | background run or slices that keep position (REQ-MWQ-003) |
| Lease evicts a merging holder | 30-min default far above the identity + merge path; renewals; M0 may only shorten |
| Requeue churn / starvation | reserved ticket that does not block, front-once, 30-min readiness bound; only own-re-measure moves count; three consecutive requeues → tail + log (REQ-MWQ-020); rule-model requeue measurement in M6 (research.md §R7) |
| Release during hold empties the window | intended: queue intact, promotion resumes on `open` (REQ-MWQ-008) |
| Empty or unstructured test runs pass the gate | refused by REQ-MWQ-016 |
| Dirty tree measured under HEAD's key | refused by REQ-MWQ-017 |
| Hook guard semantics drift | REQ-MWQ-025 regression on holder-only records |

## §H Cross-references

spec.md · acceptance.md · design.md · research.md · decision-index.md · progress.md ·
`.moai/reports/t1479/plan-audit-iter1.md` · SPEC-CANDIDATE-CI-001
