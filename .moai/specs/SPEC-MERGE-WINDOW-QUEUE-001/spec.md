---
id: SPEC-MERGE-WINDOW-QUEUE-001
title: "Merge-window automation — FIFO acquire queue, leader nomination abolished, re-measure outside the window, lane merge verb, substantive complete gate"
version: "0.6.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli (integration, factory complete/merge), internal/homestate, internal/factorylane, .claude/rules/local/gitflow-lane-protocol.md, AGENTS.local.md, internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md"
lifecycle: spec-anchored
tags: "integration-window, fifo-queue, acquire-wait, leader-nomination, window-policy, remeasure, candidate-tree, integration-merge, factory-complete, card-t1479"
tier: L
card: t1479
related_specs: [SPEC-CANDIDATE-CI-001, SPEC-INTEGRATION-LOCK-ATOMIC-001, SPEC-INTEGRATION-LOCK-LIVENESS-001, SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001, SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-LEAD-AUTOPUSH-001]
---

# SPEC-MERGE-WINDOW-QUEUE-001 — Merge-window automation (card t1479)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-03 | manager-spec | Initial draft — card t1479 (Tier L). Operator approval 2026-10-03 (leader session df44e022, AskUserQuestion): "병합 창 선착순 대기열" — leader window nomination abolished, immediate implementation. Recorded in decision-index.md Q1. |
| 0.2.0 | 2026-10-03 | manager-spec | Leader decisions Q2-Q6 (mission contract 07d28c4b; operator: implement now, include in v3.2.0): 30-min lease on by default, bare `--wait` bound 60 min, requeue front-once, candidate-CI vs local record form, tool-reported test count. |
| 0.3.0 | 2026-10-03 | manager-spec | Plan-audit iteration 1 (FAIL 0.75, `.moai/reports/t1479/plan-audit-iter1.md`) closed with leader decisions Q8-Q13: push verb moved to SPEC-CANDIDATE-CI-001; contiguous renumbering 25/25; hold suspends promotion; reserved tickets; waiter liveness + slices; empty-sweep refusal; clean tree; committed no-`--wait` baseline (`3bc274dac`). |
| 0.4.0 | 2026-10-03 | manager-spec | Leader decisions Q14-Q16: heartbeat 15 s / window 60 s / re-entry grace 120 s; requeue counting and three-requeue bound; non-test-command residual risk. |
| 0.5.0 | 2026-10-03 | manager-spec | Plan-audit iteration 2 (FAIL 0.69, regressed — STOP; `.moai/reports/t1479/plan-audit-iter2.md`, N1-N5 blocking). Leader decision Q17 (mission contract 07d28c4b): SCOPE REDUCTION. The window is now held for seconds (identity check + `--no-ff` merge), so fairness machinery buys nothing: reserved tickets, `--slice` / between-slices, front-once, the requeue counter and the three-requeue rule are removed (N2/N3/N4 dissolve with them). The queue is a plain FIFO; the ticket carries the owning-session pid that promotion stamps on the holder, restoring the `Stale` / `releasableBy` semantics (N1). A moved develop makes the merge verb release and exit with a re-measure-and-re-acquire code; the lane re-enters at the back. Leader decision Q18 folded in: the lane merge verb `moai integration merge --card <id>` with SPEC-CANDIDATE-CI-001's shared landing check (REQ-MWQ-017); the doctrine forbids manual lane `git merge` into develop. N5 hand-off item (e) added to research §R6; N6 acknowledged (REQ-MWQ-018, §R5); N7 fixture README reference fixed; N8 AC-013 checks both files per clause. Counts 23 REQ / 23 AC. |
| 0.6.0 | 2026-10-03 | manager-spec | Plan-audit iteration 3 (FAIL 0.75, `.moai/reports/t1479/plan-audit-iter3.md`, B1-B3 blocking). Leader decision Q19 (mission contract 07d28c4b; one delta round inside the auditor's fix_scope, no scope change). B1: one merge path — `moai factory complete` merges only by calling the REQ-MWQ-017 step (its own merge at `factory_card.go:1409` replaced, gates before develop moves) and adopts a landing already made by `integration merge` without re-measuring (REQ-MWQ-019); doctrine sentence covers both verbs and the self-dispatch clauses are aligned (REQ-MWQ-013). B2: SHA pinned across check, landing check and `git merge --no-ff <sha>`; every in-window failure releases with a distinct exit code; merge failure aborts and verifies a clean worktree, else sets `hold`; non-holder calls refused without touching the lock (REQ-MWQ-017/018). B3: tickets record `branch`/`branch_source`/`worktree` at enqueue and promotion copies them (REQ-MWQ-001/006). O1: `status` named a mutation (REQ-MWQ-009). O2: a waiter whose ticket was dropped exits non-zero (REQ-MWQ-003). O3/O4: research §R5 notes. Counts unchanged 23/23. |

## §A Background

Card t1479 is the second of two factory-autonomy cards issued on 2026-10-03 (the first, t1478, is
the pre-landing candidate CI, SPEC-CANDIDATE-CI-001). The integration window — the record that
serializes lane merges into local `develop` — works today as a refusal-only mutex with a human
scheduler on top:

1. **No queue.** `moai integration acquire` refuses when a live holder exists
   (`internal/kanban/integration_lock.go:273-283`; the factory merge path surfaces it as
   `merge-readiness: WAITING` at `internal/cli/factory_merge.go:157-176`; `moai factory complete`
   refuses at `internal/cli/factory_card.go:1340-1343`). A refused lane re-polls by hand; arrival
   order is not recorded and not honoured.
2. **The leader is the scheduler.** `AGENTS.local.md` §4.1 states that a `free` status is not
   approval and that only the leader's window nomination counts. The distributed doctrine already
   reads the other way (`kanban-dispatch.md` § Integration into the release branch is self-served),
   and the operator approved on 2026-10-03 replacing nomination with a first-come queue.
3. **The re-measure runs inside the window.** The documented order is acquire → absorb `develop` →
   re-measure → merge, so every lane waits behind the slowest re-measure (leader-reported: up to the
   30-minute package timeout under load 120-255 — not re-measured in this plan; see §F).
4. **The completion gate is hollow.** The merge-gate reader requires only that the re-measure file
   contain the merge SHA prefix (`internal/homestate/card_evidence_readers.go:243-257`), and when no
   file is passed `moai factory complete` writes `merge-record.txt` itself naming that SHA
   (`internal/cli/factory_card.go:1416-1421`, `1612-1628`) — the gate is satisfied by a file the
   gated command wrote. The pre-merge triple in `internal/factorylane/merge.go` runs no tests.

The batch push that was item 5 of the 0.1.0 background now belongs to SPEC-CANDIDATE-CI-001
(REQ-CCI-012/013); see §E.

## §B Goal

Make the window self-scheduling and its evidence substantive: lanes queue in arrival order and are
promoted automatically; the leader governs by policy (open / hold) instead of by nomination; the
expensive re-measure happens outside the window against the exact tree that will land; a single
merge verb performs the seconds-long in-window step; and completion requires a measurement keyed
to that tree.

## §C Requirements (GEARS)

### C.1 Window queue

- **REQ-MWQ-001** (Ubiquitous) — The integration window record shall carry a first-in-first-out
  queue of tickets, each naming the waiting session, its lane name, its card, its enqueue instant,
  the owning-session process id resolved the same way `acquire` resolves it today together with
  the `session-owner` pid source, the integration target (`branch`, `branch_source`, `worktree`)
  resolved at enqueue exactly as `acquire` resolves it today, the waiter process's id and start
  time, and its last heartbeat instant; a record written before the queue existed shall read as a
  record with an empty queue and an unchanged holder.
- **REQ-MWQ-002** (Event-driven) — **When** `moai integration acquire --wait[=<bound>]` is invoked
  while the window is held by a live session other than the caller or while the window policy is
  `hold`, the acquire verb shall append one ticket for the caller at the tail of the queue, with
  the order decided inside the same serialized record mutation that already guards acquire and
  release; the bound shall be counted from the enqueue instant and default to 60 minutes, and the
  waiting process shall refresh its ticket's heartbeat every 15 seconds while it blocks.
- **REQ-MWQ-003** (Event-driven) — **When** any queue mutation runs, the window record shall drop
  every ticket whose owning session is gone, whose waiter process (matched on id AND start time) is
  gone, or whose heartbeat is older than 60 seconds, and the mutating command shall name each
  dropped ticket and the reason in its output, and a still-running waiter that finds its own ticket
  dropped shall exit non-zero naming that reason without re-enqueueing itself; run milestone M0 may
  only tighten the 15-second and 60-second values.
- **REQ-MWQ-004** (Event-driven) — **When** a ticket's bound elapses before it is promoted, the
  acquire verb shall withdraw the ticket and exit non-zero, naming the holder, the ticket's last
  queue position, and the bound.
- **REQ-MWQ-005** (Event-driven) — **When** a ticket's promotion and its bound-elapse coincide, the
  window record shall decide both inside the record mutation and only the one recorded first shall
  take effect; **When** a waiter observes that it was promoted after its own bound had already
  elapsed, it shall release the window immediately (promoting onward) and exit non-zero naming that
  it released.
- **REQ-MWQ-006** (Event-driven) — **When** the policy is `open` and the holder releases the window,
  or the holder is stale because its owning-session process is gone or its lease has expired, the
  window record shall, in the same serialized mutation, promote the first live ticket, copying the
  ticket's session id, lane name, card, owning-session process id with its `session-owner` pid
  source, and integration target (`branch`, `branch_source`, `worktree`) onto the holder record so
  that the promoted holder's liveness, self-release, and target-ownership checks behave exactly as
  a directly acquired holder's do, and shall record any displaced holder as today's stale takeover
  does.
- **REQ-MWQ-007** (State-driven) — **While** the policy is `hold`, the window record shall perform
  no promotion of any kind: a release shall leave the window without a holder and the queue intact,
  a stale holder shall be cleared without a successor, and no new holder shall be granted; **When**
  the policy returns to `open`, promotion shall resume in queue order at that same mutation.
- **REQ-MWQ-008** (Ubiquitous) — The window lease shall be on by default with a 30-minute duration:
  acquire and promotion shall stamp the holder record with a lease expiry, every window verb the
  holder invokes shall renew it, and a configured duration of zero shall disable it (validity then
  decided by owning-session liveness alone, as before this SPEC); the default may be lowered only
  on the in-window duration measured in run milestone M0.
- **REQ-MWQ-009** (Ubiquitous) — `moai integration status` shall show the holder, the holder's lease
  expiry, the window policy, and every queued ticket in order with its position and liveness, in
  both the human form and the `--json` form; like `acquire`, `release`, `policy`, and `merge`, it is
  a queue mutation that applies REQ-MWQ-003 drops and REQ-MWQ-006 promotion before it prints.
- **REQ-MWQ-010** (Capability gate) — **Where** `--wait` is absent, `moai integration acquire` shall
  keep its current refusal behavior and output byte-for-byte against the committed fixture
  `.moai/reports/t1479/baseline-acquire-nowait/` (after its two stated normalizations), and shall
  enqueue nothing.
- **REQ-MWQ-011** (Unwanted) — The window record shall not let a ticket overtake an earlier live
  ticket, shall not grant the window to an acquire without `--wait` while a live ticket is queued,
  and shall not hold two tickets for one session; `--force` shall remain the only path that takes a
  live holder's window and shall keep recording what it displaced.

### C.2 Window policy and doctrine — nomination abolished

- **REQ-MWQ-012** (Event-driven) — **When** `moai integration policy open` or
  `moai integration policy hold --reason <text>` is invoked, the policy verb shall persist the
  policy beside the window record (an absent policy reading as `open`), refusing the write for a
  session that declares the lane role; **While** the policy is `hold`, acquire without `--wait`
  shall refuse naming the reason and acquire with `--wait` shall enqueue.
- **REQ-MWQ-013** (Ubiquitous) — The local doctrine (`AGENTS.local.md` §4.1 and
  `.claude/rules/local/gitflow-lane-protocol.md`) shall each carry the sentence
  `open 정책에서 대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다` and the sentence
  `레인은 develop에 손으로 git merge 하지 않고 moai integration merge --card 또는 그것을 부르는 moai factory complete 로만 병합한다`,
  the self-dispatch merge clause of each file (today `AGENTS.local.md:219`,
  `gitflow-lane-protocol.md:99`) shall say that `moai factory complete` merges through
  `moai integration merge`, and neither file shall carry the nomination clause (`지명만이 근거`) or
  the announcement-as-first-layer clause (`리더 공지가 여전히 첫 번째 층`).

### C.3 Re-measure outside the window, merge verb inside it

- **REQ-MWQ-014** (Ubiquitous) — The re-measure that gates a merge shall run before the lane joins
  the window queue, against the candidate tree (the card branch's tree after absorbing the
  integration branch tip), and shall produce a record keyed by that tree SHA carrying the absorbed
  integration-branch commit and the producing moai build's identity; one verifier shall read both
  record forms — the local form (REQ-MWQ-015) and the candidate-CI form (a SPEC-CANDIDATE-CI-001
  candidate run id whose verdict is green) — and shall require the candidate-CI form where
  `workflow.candidate_ci.enabled` is true and the local form where it is false or absent.
- **REQ-MWQ-015** (Event-driven) — **When** the re-measure verb runs a command, it shall itself
  capture the command and its exit code; **When** the command's tool emits a recognized structured
  test report (in this repository, `go test -json`), it shall record the reported test count, and a
  count of zero, a runner-reported empty sweep (`no tests to run`, `[no test files]`, and the
  equivalent markers of other recognized runners), or a run of a tool that supports structured
  output without it (in this repository, `go test` without `-json`) shall make the record invalid;
  a command whose tool has no recognized report shall be valid on exit code zero.
- **REQ-MWQ-016** (Event-driven) — **When** the re-measure verb starts or finishes, it shall verify
  that `git status --porcelain` is empty (tracked and untracked) and that `HEAD` is unchanged
  across the run, and shall refuse to write a record otherwise.
- **REQ-MWQ-017** (State-driven) — **While** the caller holds the window,
  `moai integration merge --card <id>` — the one in-window merge step, which `moai factory complete`
  also calls — shall resolve the card's `WT-` branch the same way SPEC-CANDIDATE-CI-001 REQ-CCI-004
  resolves it, pin the branch tip to one commit SHA, check that SHA's tree against the card's
  re-measure record and the record's absorbed commit against the integration branch tip, call
  SPEC-CANDIDATE-CI-001's shared landing check (REQ-CCI-011; a no-op while
  `workflow.candidate_ci.enabled` is false) for that SHA, run `git merge --no-ff <pinned SHA>` into
  the integration branch (never the branch name), verify the merge commit's tree equals the
  record's tree, and release the window; it shall run no test suite; **When** the caller does not
  hold the window, it shall refuse without reading or writing the window record's holder or queue.
- **REQ-MWQ-018** (Event-driven) — **When** any in-window step fails, the merge step shall release
  the window (promoting the next live ticket) and exit with a code distinct per cause — integration
  tip moved since the record's absorbed commit (the re-measure-and-re-acquire code, naming both
  SHAs; the lane then re-absorbs, re-measures, and re-acquires at the tail, the only point at which
  a stale candidate is rebuilt — never eagerly at queue entry, as SPEC-CANDIDATE-CI-001 assigns to
  this SPEC), pinned tree differing from the record's tree, landing-check refusal, merge failure,
  and any other error; on a merge failure it shall run `git merge --abort` and verify the
  integration worktree is clean before releasing, and **When** the worktree is still not clean, it
  shall set the window policy to `hold` with a reason naming the dirty worktree before releasing,
  so no later holder is promoted onto it.

### C.4 Substantive completion gate

- **REQ-MWQ-019** (Event-driven) — **When** `moai factory complete` runs, it shall perform its merge
  only by calling the REQ-MWQ-017 merge step (replacing its own merge, so every gate runs before
  develop moves) and shall then record the card as merged-local; **When** no valid re-measure record
  keyed by the card branch's candidate tree exists, it shall refuse before any card state transition
  and before any merge; **When** the step fails, the card state shall not change; and **When** a
  merge commit of that card's branch is already reachable from the integration branch (the lane
  merged through `moai integration merge` first), it shall record merged-local from that commit
  without calling the merge step and without requiring a fresh re-measure.
- **REQ-MWQ-020** (Unwanted) — `moai factory complete` shall not write a record that stands in for
  the re-measure; the merge identity it records shall be stored separately from, and shall never
  satisfy, the re-measure requirement.
- **REQ-MWQ-021** (Ubiquitous) — The merge-gate reader shall accept a merge only when the record's
  tree equals the merge commit's tree and the record is valid under REQ-MWQ-014/015 (containing the
  merge SHA as text shall no longer be sufficient), and the merge-readiness pre-checks
  (`moai factory merge ready`) shall report record validity as a named fourth condition, printing
  the recorded command verbatim.

### C.5 Distribution and compatibility

- **REQ-MWQ-022** (Ubiquitous) — The distributed window procedure
  (`kanban-dispatch-mechanics.md` § Integration into the release branch, template source first)
  shall describe serialization by the recorded hold, the queue, the window policy, and the merge
  verb, shall drop the announcement to the leader as a layer, and shall stay neutral across the
  supported programming languages.
- **REQ-MWQ-023** (Unwanted) — No change in this SPEC shall alter what the PreToolUse
  integration-lock guard decides for a record that carries no queue, lease, or policy data.

## §D Non-functional constraints

- Backward compatibility: every new record field is additive and optional; existing records and the
  acquire/release/status outputs without the new flags are unchanged (REQ-MWQ-010, -023).
- Serialization: every queue mutation runs inside the existing integration-lock mutation section
  (SPEC-INTEGRATION-LOCK-ATOMIC-001); no second lock is introduced.
- Holder liveness is unchanged in kind: owning-session pid with the live-when-unknown asymmetry
  (SPEC-INTEGRATION-LOCK-LIVENESS-001), now carried through promotion (REQ-MWQ-006); a lease expiry
  is the only new route to holder staleness. Tickets additionally require their waiter process,
  because a dropped ticket costs a re-enqueue while an orphan ticket costs a wedged window.
- A Bash tool call is capped at 10 minutes, below the 60-minute default wait; lanes run
  `acquire --wait` as a background command. A waiter that exits loses its ticket and re-enqueues at
  the tail — accepted, because the window is held for seconds.
- Trust model: the record is written by the re-measure verb and carries its build identity, but a
  hand-written record file is not distinguishable by the verifier. Actors are assumed cooperative;
  forgery is a residual risk, not a defended boundary.
- Residual risk — non-test commands (leader decision Q12/Q16): a command whose tool emits no
  recognized test structure stays valid on exit code zero, so the verifier cannot tell a non-test
  command (for example `true`) from a real test run in that case. The candidate-CI form
  (SPEC-CANDIDATE-CI-001, `workflow.candidate_ci.enabled: true`) is the stronger record and removes
  this gap where it is enabled.
- Residual risk — no fairness beyond FIFO: a lane whose re-measure is repeatedly invalidated by
  other merges re-enters at the tail each time (research.md §R7).

## §E Exclusions

### Out of Scope — Integration push

- `moai integration push` (threshold, red-tip hold, pinned-SHA push, landing report) is specified in
  SPEC-CANDIDATE-CI-001 REQ-CCI-012/013 (card t1478); research.md §R6 carries this SPEC's audit
  findings on it as a hand-off.

### Out of Scope — Queue fairness machinery

- Reserved re-entry tickets, `--slice` waits, front-once promotion, and requeue counting were
  removed in v0.5.0 (decision-index Q17); this SPEC provides plain FIFO only.

### Out of Scope — Codex lanes

- Codex lanes keep stopping at merge-ready (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-025); they do not
  join the window queue through this SPEC.

### Out of Scope — CI workflow and candidate commits

- Candidate commits, `ci/**` branches, the verdict reader, branch resolution, and the shared landing
  check are SPEC-CANDIDATE-CI-001's; this SPEC calls them and produces none of them.

### Out of Scope — Lane remote pushes

- Lanes still never push their `WT-` branches or request CI directly through this SPEC.

## §F Gaps at plan time

- The in-window re-measure duration (up to the 30-minute timeout under load 120-255) is the
  leader's investigation figure on develop `42d8474de`; this plan did not re-measure it.
- Cited line numbers were re-read on this tree (HEAD `d7112d005`, tree `632f65b47aa5`) and by the
  plan audits on `1e1d0cc84` and `6ab3dbcb2`; see research.md §R1.
- Heartbeat 15 s and heartbeat window 60 s are leader values (Q14), not measurements; M0 may only
  tighten them.
- SPEC-CANDIDATE-CI-001 is not landed; REQ-MWQ-017 depends on its REQ-CCI-004 branch resolution and
  REQ-CCI-011 landing check. If this SPEC's run precedes t1478's landing, the run implements the
  branch resolution locally behind the same contract and treats the landing check as absent (no-op),
  and whichever card lands second reconciles the two (research.md §R5).
