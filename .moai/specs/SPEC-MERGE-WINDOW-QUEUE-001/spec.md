---
id: SPEC-MERGE-WINDOW-QUEUE-001
title: "Merge-window automation — FIFO acquire queue, leader nomination abolished, re-measure moved out of the window, substantive complete gate, gated integration push"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli (integration, factory complete/merge), internal/homestate, internal/factorylane, .claude/rules/local/gitflow-lane-protocol.md, AGENTS.local.md, internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-mechanics.md"
lifecycle: spec-anchored
tags: "integration-window, fifo-queue, acquire-wait, leader-nomination, window-policy, remeasure, candidate-tree, factory-complete, integration-push, lead-push-threshold, card-t1479"
tier: L
card: t1479
related_specs: [SPEC-INTEGRATION-LOCK-ATOMIC-001, SPEC-INTEGRATION-LOCK-LIVENESS-001, SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001, SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-LEAD-AUTOPUSH-001, SPEC-LANE-PUSH-BATCH-001, SPEC-CI-VERDICT-PRODUCER-001]
---

# SPEC-MERGE-WINDOW-QUEUE-001 — Merge-window automation (card t1479)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-03 | manager-spec | Initial draft — card t1479 (Tier L). Operator approval 2026-10-03 (leader session df44e022, AskUserQuestion): "병합 창 선착순 대기열" — leader window nomination abolished, immediate implementation. Recorded in decision-index.md Q1. |

## §A Background

Card t1479 is the second of two factory-autonomy cards issued on 2026-10-03 (the first, t1478, is
the pre-landing candidate CI). The integration window — the record that serializes lane merges into
local `develop` — works today as a refusal-only mutex with a human scheduler on top:

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
   30-minute package timeout under load 120-255 — not re-measured in this plan; see §E Gaps).
4. **The completion gate is hollow.** The merge-gate reader requires only that the re-measure file
   contain the merge SHA prefix (`internal/homestate/card_evidence_readers.go:243-257`), and when no
   file is passed `moai factory complete` writes `merge-record.txt` itself naming that SHA
   (`internal/cli/factory_card.go:1416-1421`, `1612-1628`) — the gate is satisfied by a file the
   gated command wrote. The pre-merge triple in `internal/factorylane/merge.go` runs no tests.
5. **Batch push exists only in prose.** `git_strategy.manual.lead_push_threshold` is a typed config
   field (`internal/config/types.go:140`, default `internal/config/defaults.go:1050`) with zero
   consumers outside `internal/config`; the threshold trigger and the red-CI hold are executed by
   hand per `.claude/rules/local/gitflow-lane-protocol.md` §4.

## §B Goal

Make the window self-scheduling and its evidence substantive: lanes queue in arrival order and are
promoted automatically; the leader governs by policy (open / hold) instead of by nomination; the
expensive re-measure happens outside the window against the exact tree that will land; completion
requires a measurement keyed to that tree; and the leader's batch push becomes one gated verb.

## §C Requirements (GEARS)

### C.1 Window queue

- **REQ-MWQ-001** (Ubiquitous) — The integration window record shall carry an ordered waiter queue
  of tickets, each naming the waiting session, its owning-session process id, its lane name, its
  card, and its enqueue instant; a record written before the queue existed shall read as a record
  with an empty queue and an unchanged holder.
- **REQ-MWQ-002** (Event-driven) — **When** `moai integration acquire --wait[=<bound>]` is invoked
  while the window is held by a live session other than the caller or while the window policy is
  `hold`, the acquire verb shall append one ticket for the caller at the tail of the queue, with
  the arrival order decided inside the same serialized record mutation that already guards acquire
  and release, and shall block until the caller becomes the holder or the bound elapses.
- **REQ-MWQ-003** (Event-driven) — **When** the holder releases the window, the release shall, in
  the same serialized mutation, promote the first ticket whose owning session is live to holder and
  remove it from the queue; tickets whose owning session is gone shall be removed and named in the
  release output.
- **REQ-MWQ-004** (Event-driven) — **When** any acquire, wait poll, or status read observes that
  the holder's owning session is gone or that the holder's lease has expired, the window record
  shall promote the queue head under the same rules as REQ-MWQ-003 and shall record the displaced
  holder, exactly as today's stale takeover records it.
- **REQ-MWQ-005** (Event-driven) — **When** a waiting caller's bound elapses before it is promoted,
  the acquire verb shall remove the caller's own ticket and exit non-zero, naming the holder, the
  caller's last queue position, and the bound.
- **REQ-MWQ-006** (Ubiquitous) — `moai integration status` shall show the holder, the holder's
  lease expiry, the window policy, and every queued ticket in order with its position and liveness,
  in both the human form and the `--json` form.
- **REQ-MWQ-007** (Capability gate) — **Where** `--wait` is absent, `moai integration acquire`
  shall keep its current refusal behavior and output byte-for-byte, and shall enqueue nothing.
- **REQ-MWQ-008** (Unwanted) — The window record shall not let a ticket overtake an earlier live
  ticket, and a caller that already holds a ticket shall not gain a second one by re-invoking
  acquire; `--force` shall remain the only path that takes a live holder's window and shall keep
  recording what it displaced.
- **REQ-MWQ-009** (Capability gate) — **Where** a window lease duration is configured to a non-zero
  value, acquire shall stamp the holder record with a lease expiry and the holder shall be able to
  renew it; **Where** the duration is zero or absent, the holder's validity shall be decided by
  owning-session liveness alone, exactly as today.

### C.2 Window policy — nomination abolished

- **REQ-MWQ-010** (Ubiquitous) — The window shall carry a policy that is either `open` or
  `hold:<reason>`, persisted beside the window record; an absent policy shall read as `open`.
- **REQ-MWQ-011** (State-driven) — **While** the policy is `hold`, acquire shall grant no new
  holder: without `--wait` it shall refuse naming the hold reason, and with `--wait` it shall
  enqueue and keep waiting; the current holder shall not be displaced by a hold.
- **REQ-MWQ-012** (Event-driven) — **When** a session that declares the lane role attempts to set
  the window policy, the policy verb shall refuse; the leader session and a session that declares
  no factory role shall be able to set it.
- **REQ-MWQ-013** (Ubiquitous) — The local doctrine (`AGENTS.local.md` §4.1 and
  `.claude/rules/local/gitflow-lane-protocol.md`) shall state that a lane is authorized to merge
  when it is promoted to holder under an `open` policy, and shall no longer state that a leader
  nomination is required.

### C.3 Re-measure on the candidate tree, outside the window

- **REQ-MWQ-020** (Ubiquitous) — The re-measure that gates a merge shall run before the lane joins
  the window queue, against the candidate tree: the card branch's tree after it has absorbed the
  integration branch tip, or — where the pre-landing candidate path of card t1478 is available —
  the tree of that candidate commit.
- **REQ-MWQ-021** (Ubiquitous) — A re-measure record shall be keyed by the candidate tree SHA and
  shall carry the integration-branch commit that was absorbed, the command that was run, its exit
  code, and either a test count greater than zero or the id of a candidate CI run whose conclusion
  is success; the command and exit code shall be captured by the moai verb that executed the
  command rather than asserted by the caller.
- **REQ-MWQ-022** (State-driven) — **While** a lane holds the window, the merge path shall perform
  only the tree-identity check against the re-measure record and the `--no-ff` merge, followed by
  the existing merge-tree identity verification; it shall run no test suite inside the window.
- **REQ-MWQ-023** (Event-driven) — **When** the integration branch tip has moved since the
  re-measure record's absorbed commit, the merge path shall not merge; it shall release the window
  (promoting the next ticket) and report that a re-absorb and re-measure are required, naming the
  recorded and the current integration-branch SHAs.

### C.4 Substantive completion gate

- **REQ-MWQ-030** (Event-driven) — **When** `moai factory complete` runs and no valid re-measure
  record keyed by the card branch's candidate tree exists, the command shall refuse before any
  card state transition and before any merge.
- **REQ-MWQ-031** (Unwanted) — `moai factory complete` shall not write a record that stands in for
  the re-measure; the merge identity it records shall be stored separately from, and shall never
  satisfy, the re-measure requirement.
- **REQ-MWQ-032** (Ubiquitous) — The merge-gate reader shall accept a merge only when the
  re-measure record's tree equals the merge commit's tree, the recorded exit code is zero, and the
  record carries a positive test count or a successful candidate CI run id; containing the merge
  SHA as text shall no longer be sufficient.
- **REQ-MWQ-033** (Ubiquitous) — The merge-readiness pre-checks (`moai factory merge ready`) shall
  include the re-measure record's presence and validity as a named condition alongside the existing
  sync-audit, conflict-free, and tree-identity conditions.

### C.5 Gated integration push

- **REQ-MWQ-040** (Event-driven) — **When** `moai integration push` is invoked, the push verb shall
  count the commits of the local integration branch not yet on its remote-tracking ref and compare
  the count with `git_strategy.manual.lead_push_threshold`; below a non-zero threshold it shall not
  push and shall report the count and the threshold; a zero or absent threshold shall impose no
  count condition.
- **REQ-MWQ-041** (State-driven) — **While** the most recent completed CI run on the remote
  integration branch tip concluded in failure, the push verb shall refuse; a tip with no completed
  run yet shall not be treated as red.
- **REQ-MWQ-042** (Event-driven) — **When** the CI state cannot be read (the CI client is absent,
  unauthenticated, or the query fails), the push verb shall refuse naming the fault rather than
  treat the state as green.
- **REQ-MWQ-043** (Unwanted) — The push verb shall not run for a session that declares the lane
  role, shall not run while the integration window is held, and shall not force-push.
- **REQ-MWQ-044** (Event-driven) — **When** the push verb proceeds, it shall first re-read the
  integration branch's local tip and confirm no merge is in progress in the integration worktree
  (the pre-push final read), push that tip once, then fetch and report whether the remote-tracking
  ref now equals the pushed tip.

### C.6 Doctrine and distribution

- **REQ-MWQ-050** (Ubiquitous) — The distributed window procedure
  (`kanban-dispatch-mechanics.md` § Integration into the release branch, template source first)
  shall describe serialization by the recorded hold and the queue, without the announcement to the
  leader as a layer, and shall stay neutral across the supported programming languages.
- **REQ-MWQ-051** (Unwanted) — No change in this SPEC shall alter what the PreToolUse
  integration-lock guard decides for a record that carries no queue, lease, or policy data.

## §D Non-functional constraints

- Backward compatibility: every new record field is additive and optional; existing records and the
  existing acquire/release/status outputs without the new flags are unchanged (REQ-MWQ-007, -051).
- Serialization: every queue mutation runs inside the existing integration-lock mutation section
  (SPEC-INTEGRATION-LOCK-ATOMIC-001); no second lock is introduced.
- Liveness asymmetry is preserved: an indeterminate owner reads as live
  (SPEC-INTEGRATION-LOCK-LIVENESS-001); a lease expiry is the only new route to staleness.
- Push remains a leader act on shared external state (keep-set); the verb gates it, it does not
  automate it — no daemon, timer, or hook invokes it.

## §E Exclusions

### Out of Scope — Codex lanes

- Codex lanes keep stopping at merge-ready (SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-025); they do not
  join the window queue through this SPEC.

### Out of Scope — CI workflow and candidate commits

- The `ci/**` candidate branches, workflow triggers, and CI job restructuring belong to card t1478;
  this SPEC consumes a candidate CI run id where one exists and does not produce one.

### Out of Scope — Push automation

- No automatic push: no timer, hook, or daemon invokes `moai integration push`; the leader invokes it.
- Release branches, `main`, and pull requests are untouched.

### Out of Scope — Lane remote pushes

- Lanes still never push their `WT-` branches or request CI directly; this SPEC does not relax
  AGENTS.local.md §4.1 on that axis (t1478 owns any change there).

## §F Gaps at plan time

- The in-window re-measure duration (up to the 30-minute timeout under load 120-255) is the
  leader's investigation figure on develop `42d8474de`; this plan did not re-measure it.
- Cited line numbers were re-read on this tree (HEAD `d7112d005`, tree `632f65b47aa5`); see
  research.md §R1 for the baseline cells.
