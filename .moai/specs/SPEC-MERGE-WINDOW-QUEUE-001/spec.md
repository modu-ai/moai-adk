---
id: SPEC-MERGE-WINDOW-QUEUE-001
title: "Merge-window automation — FIFO acquire queue, leader nomination abolished, re-measure moved out of the window, substantive complete gate"
version: "0.4.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli (integration, factory complete/merge), internal/homestate, internal/factorylane, .claude/rules/local/gitflow-lane-protocol.md, AGENTS.local.md, internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md"
lifecycle: spec-anchored
tags: "integration-window, fifo-queue, acquire-wait, waiter-liveness, leader-nomination, window-policy, remeasure, candidate-tree, factory-complete, card-t1479"
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
| 0.3.0 | 2026-10-03 | manager-spec | Plan-audit iteration 1 (FAIL 0.75, `.moai/reports/t1479/plan-audit-iter1.md`, MP-1 + D1-D10 blocking) closed with leader decisions (mission contract 07d28c4b, decision-index Q8-Q13). Push verb (former REQ-040..044) removed to SPEC-CANDIDATE-CI-001 (card t1478) with a hand-off note in research.md §R6 (D1/D8). Lettered items folded; REQ and AC renumbered contiguously 001-025 (MP-1/D2). Hold suspends every promotion (REQ-008, D3). Reserved re-entry ticket that does not block the queue, with a 30-min readiness bound (REQ-020, D4). Waiter-process liveness + heartbeat, timeout-vs-promotion race, slice re-entry without position loss (REQ-003..006, D5). Empty sweeps and unstructured runs refused (REQ-016, D6). Clean tree before and after the re-measure (REQ-017, D7). No-`--wait` baseline fixture committed ahead of code (`.moai/reports/t1479/baseline-acquire-nowait/`, commit `3bc274dac`, D9). Exact doctrine grep (AC-014, D10). Optional D11-D14 addressed (M0 method, `workflow.candidate_ci.enabled` named, record provenance + residual risk, positive grep on the distributed text). |
| 0.4.0 | 2026-10-03 | manager-spec | Leader decisions (mission contract 07d28c4b): Q14 liveness values — heartbeat 15 s, window 60 s, re-entry grace 120 s, M0 tighten-only (REQ-MWQ-003/004); Q15 starvation — only a base move during the ticket's own re-measure counts toward "second move → tail", a move while a ready ticket waits for another holder keeps its front position, and three consecutive requeues of any kind send it to the tail with a logged event (REQ-MWQ-020, AC-MWQ-020 scenarios 4-6); Q16 — Q12 boundary accepted, non-test-command residual risk stated (§D). Counts unchanged 25/25. |

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
expensive re-measure happens outside the window against the exact tree that will land; and
completion requires a measurement keyed to that tree.

## §C Requirements (GEARS)

### C.1 Window queue

- **REQ-MWQ-001** (Ubiquitous) — The integration window record shall carry an ordered queue of
  tickets, each naming the waiting session, its lane name, its card, its enqueue instant, the
  waiter process's id and start time, its last heartbeat instant, and its state (`waiting`,
  `between-slices`, or `reserved`); a record written before the queue existed shall read as a
  record with an empty queue and an unchanged holder.
- **REQ-MWQ-002** (Event-driven) — **When** `moai integration acquire --wait[=<total>]` is invoked
  while the window is held by a live session other than the caller or while the window policy is
  `hold`, the acquire verb shall append one ticket for the caller at the tail of the queue, with
  the arrival order decided inside the same serialized record mutation that already guards acquire
  and release; the ticket's total wait bound shall be counted from its first enqueue, defaulting to
  60 minutes when no value is given, and the waiting process shall refresh the ticket's heartbeat
  while it blocks.
- **REQ-MWQ-003** (Event-driven) — **When** an invocation given a per-invocation slice
  (`--slice <duration>`) reaches the end of its slice without being promoted, the acquire verb
  shall keep the caller's ticket at its position, mark it `between-slices` with a re-entry deadline,
  and exit with a dedicated still-queued exit code naming the position; **When** the same session
  re-invokes `acquire --wait` before that deadline, the verb shall resume the existing ticket at the
  same position rather than enqueue a new one. The re-entry grace shall be 120 seconds.
- **REQ-MWQ-004** (Event-driven) — **When** any queue mutation runs, the window record shall drop
  every ticket whose waiter process (matched on id AND start time) is gone, whose heartbeat is
  older than the heartbeat window, or whose `between-slices` re-entry deadline has passed, and the
  mutating command shall name each dropped ticket and the reason in its output. The waiter shall
  heartbeat every 15 seconds and the heartbeat window shall be 60 seconds (four missed beats); run
  milestone M0 may only tighten the heartbeat, window, and re-entry grace values.
- **REQ-MWQ-005** (Event-driven) — **When** a ticket's total wait bound elapses before it is
  promoted, the acquire verb shall withdraw the ticket and exit non-zero, naming the holder, the
  ticket's last queue position, and the bound.
- **REQ-MWQ-006** (Event-driven) — **When** a ticket's promotion and its bound-elapse coincide, the
  window record shall decide both inside the record mutation and only the one recorded first shall
  take effect; **When** a waiter observes that it was promoted after its own bound had already
  elapsed, it shall release the window immediately (promoting onward) and exit non-zero naming that
  it released.
- **REQ-MWQ-007** (Event-driven) — **When** the policy is `open` and the holder releases the window,
  or the holder is stale because its owning process is gone or its lease has expired, the window
  record shall, in the same serialized mutation, promote the first ticket that is live and ready
  (state `waiting`, or `reserved` that has become ready under REQ-MWQ-020) and shall record any
  displaced holder exactly as today's stale takeover records it.
- **REQ-MWQ-008** (State-driven) — **While** the policy is `hold`, the window record shall perform
  no promotion of any kind: a release shall leave the window without a holder and the queue intact,
  a stale holder shall be cleared without a successor, and no new holder shall be granted; **When**
  the policy returns to `open`, promotion shall resume in queue order at that same mutation.
- **REQ-MWQ-009** (Ubiquitous) — The window lease shall be on by default with a 30-minute duration:
  acquire and promotion shall stamp the holder record with a lease expiry, every window verb the
  holder invokes shall renew it, and a configured duration of zero shall disable it (validity then
  decided by owning-session liveness alone, as before this SPEC); the default may be lowered only
  on the in-window duration measured in run milestone M0.
- **REQ-MWQ-010** (Ubiquitous) — `moai integration status` shall show the holder, the holder's lease
  expiry, the window policy, and every queued ticket in order with its position, state, liveness,
  and the recorded re-measure command of the holder and of each ready ticket, in both the human
  form and the `--json` form.
- **REQ-MWQ-011** (Capability gate) — **Where** `--wait` is absent, `moai integration acquire` shall
  keep its current refusal behavior and output byte-for-byte against the committed fixture
  `.moai/reports/t1479/baseline-acquire-nowait/` (after its two stated normalizations), and shall
  enqueue nothing.
- **REQ-MWQ-012** (Unwanted) — The window record shall not let a ticket overtake an earlier live and
  ready ticket except as REQ-MWQ-020 grants a reserved ticket its one front promotion, and a session
  that already holds a ticket shall not gain a second one; `--force` shall remain the only path that
  takes a live holder's window and shall keep recording what it displaced.

### C.2 Window policy — nomination abolished

- **REQ-MWQ-013** (Event-driven) — **When** `moai integration policy open` or
  `moai integration policy hold --reason <text>` is invoked, the policy verb shall persist the
  policy beside the window record (an absent policy reading as `open`), refusing the write for a
  session that declares the lane role; **While** the policy is `hold`, acquire without `--wait`
  shall refuse naming the reason and acquire with `--wait` shall enqueue.
- **REQ-MWQ-014** (Ubiquitous) — The local doctrine (`AGENTS.local.md` §4.1 and
  `.claude/rules/local/gitflow-lane-protocol.md`) shall each carry the sentence
  `open 정책에서 대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다`, and shall no
  longer carry the nomination clause (`지명만이 근거`) or the announcement-as-first-layer clause
  (`리더 공지가 여전히 첫 번째 층`).

### C.3 Re-measure on the candidate tree, outside the window

- **REQ-MWQ-015** (Ubiquitous) — The re-measure that gates a merge shall run before the lane joins
  the window queue, against the candidate tree (the card branch's tree after absorbing the
  integration branch tip), and shall produce a record keyed by that tree SHA carrying the absorbed
  integration-branch commit and the producing moai build's identity; one verifier shall read both
  record forms — the local form (REQ-MWQ-016) and the candidate-CI form (a SPEC-CANDIDATE-CI-001
  candidate run id whose verdict is green) — and shall require the candidate-CI form where
  `workflow.candidate_ci.enabled` is true and the local form where it is false or absent.
- **REQ-MWQ-016** (Event-driven) — **When** the re-measure verb runs a command, it shall itself
  capture the command and its exit code; **When** the command's tool emits a recognized structured
  test report (in this repository, `go test -json`), it shall record the reported test count, and a
  count of zero, a runner-reported empty sweep (`no tests to run`, `[no test files]`, and the
  equivalent markers of other recognized runners), or a run of a tool that supports structured
  output without it (in this repository, `go test` without `-json`) shall make the record invalid;
  a command whose tool has no recognized report shall be valid on exit code zero.
- **REQ-MWQ-017** (Event-driven) — **When** the re-measure verb starts or finishes, it shall verify
  that `git status --porcelain` is empty (tracked and untracked) and that `HEAD` is unchanged
  across the run, and shall refuse to write a record otherwise.
- **REQ-MWQ-018** (State-driven) — **While** a lane holds the window, the merge path shall perform
  only the tree-identity check against the record (under `workflow.candidate_ci.enabled`, the
  shared landing check of SPEC-CANDIDATE-CI-001 REQ-CCI-011 serves as that check), the `--no-ff`
  merge, and the existing merge-tree identity verification; it shall run no test suite inside the
  window.
- **REQ-MWQ-019** (Event-driven) — **When** the integration branch tip has moved since the record's
  absorbed commit, the merge path shall not merge; it shall release the window (promoting the next
  ready ticket), convert the caller's place into a `reserved` ticket under REQ-MWQ-020, and report
  that a re-absorb and re-measure are required, naming the recorded and the current SHAs.
- **REQ-MWQ-020** (State-driven) — **While** a `reserved` ticket's owner has not produced a record
  whose absorbed commit equals the current integration tip, the ticket shall be not ready and shall
  not block promotion of the next ready ticket; **When** it becomes ready, it shall receive the next
  promotion once and shall keep that front position while it waits for another holder's window;
  **When** it has not become ready within its readiness bound (default 30 minutes) it shall be
  dropped with that reason; **When** a ticket is requeued by a base move that happened during its
  own re-measure for the second consecutive time, it shall go to the tail as an ordinary `waiting`
  ticket, whereas a base move that happened while the ready ticket waited for another holder's
  window shall not count toward that limit; and **When** a ticket reaches three consecutive
  requeues of any kind, it shall go to the tail and the window record shall log the event.

### C.4 Substantive completion gate

- **REQ-MWQ-021** (Event-driven) — **When** `moai factory complete` runs and no valid re-measure
  record keyed by the card branch's candidate tree exists, the command shall refuse before any card
  state transition and before any merge.
- **REQ-MWQ-022** (Unwanted) — `moai factory complete` shall not write a record that stands in for
  the re-measure; the merge identity it records shall be stored separately from, and shall never
  satisfy, the re-measure requirement.
- **REQ-MWQ-023** (Ubiquitous) — The merge-gate reader shall accept a merge only when the record's
  tree equals the merge commit's tree and the record is valid under REQ-MWQ-015/016 (containing the
  merge SHA as text shall no longer be sufficient), and the merge-readiness pre-checks
  (`moai factory merge ready`) shall report record validity as a named fourth condition, printing
  the recorded command verbatim.

### C.5 Doctrine and compatibility

- **REQ-MWQ-024** (Ubiquitous) — The distributed window procedure
  (`kanban-dispatch-mechanics.md` § Integration into the release branch, template source first)
  shall describe serialization by the recorded hold, the queue, and the window policy, shall drop
  the announcement to the leader as a layer, and shall stay neutral across the supported
  programming languages.
- **REQ-MWQ-025** (Unwanted) — No change in this SPEC shall alter what the PreToolUse
  integration-lock guard decides for a record that carries no queue, lease, or policy data.

## §D Non-functional constraints

- Backward compatibility: every new record field is additive and optional; existing records and the
  acquire/release/status outputs without the new flags are unchanged (REQ-MWQ-011, -025).
- Serialization: every queue mutation runs inside the existing integration-lock mutation section
  (SPEC-INTEGRATION-LOCK-ATOMIC-001); no second lock is introduced.
- Liveness asymmetry is preserved for holders: an indeterminate owner reads as live
  (SPEC-INTEGRATION-LOCK-LIVENESS-001); a lease expiry is the only new route to holder staleness.
  Tickets use the opposite default — a waiter that cannot be shown alive is dropped (REQ-MWQ-004) —
  because a dropped ticket costs a re-enqueue while an orphan ticket costs a wedged window.
- Trust model: the record is written by the re-measure verb and carries its build identity, but a
  hand-written record file is not distinguishable by the verifier. Actors are assumed cooperative;
  forgery is a residual risk, not a defended boundary.
- Residual risk — non-test commands (leader decision Q12/Q16): a command whose tool emits no
  recognized test structure stays valid on exit code zero, so the verifier cannot tell a
  non-test command (for example `true`) from a real test run in that case. The candidate-CI form
  (SPEC-CANDIDATE-CI-001, `workflow.candidate_ci.enabled: true`) is the stronger record and
  removes this gap where it is enabled.

## §E Exclusions

### Out of Scope — Integration push

- `moai integration push` (threshold, red-tip hold, pinned-SHA push, landing report) is specified in
  SPEC-CANDIDATE-CI-001 REQ-CCI-012/013 (card t1478); research.md §R6 carries this SPEC's audit
  findings on it as a hand-off.

### Out of Scope — Codex lanes

- Codex lanes keep stopping at merge-ready (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-025); they do not
  join the window queue through this SPEC.

### Out of Scope — CI workflow and candidate commits

- Candidate commits, `ci/**` branches, the verdict reader, and the shared landing check are
  SPEC-CANDIDATE-CI-001's; this SPEC consumes a green candidate run id and the landing check and
  produces neither.

### Out of Scope — Lane remote pushes

- Lanes still never push their `WT-` branches or request CI directly through this SPEC.

## §F Gaps at plan time

- The in-window re-measure duration (up to the 30-minute timeout under load 120-255) is the
  leader's investigation figure on develop `42d8474de`; this plan did not re-measure it.
- Cited line numbers were re-read on this tree (HEAD `d7112d005`, tree `632f65b47aa5`) and by the
  plan audit on `1e1d0cc84`; see research.md §R1.
- Heartbeat interval (15 s), heartbeat window (60 s), and re-entry grace (120 s) are leader
  decisions (Q14), not measurements; M0 may only tighten them.
