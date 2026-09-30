---
id: SPEC-LANE-STALL-WATCHDOG-001
title: "Lane stall watchdog — self-diagnosis, per-cause autonomous resume, and unified --auto semantics across the todo, lane, and goal surfaces"
version: "0.1.0"
status: draft
created: 2026-09-30
updated: 2026-09-30
author: manager-spec (card t1370)
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow, .claude/skills/moai-lane-watchdog, internal/template/templates, .claude/rules/local"
lifecycle: spec-anchored
tags: "watchdog, stall, lane, auto-semantics, todo-auto, goal-auto, resume, self-diagnosis, card-t1370"
tier: M
card: t1370
related_specs: [SPEC-RELATION-PICKUP-FILTER-001, SPEC-MANAGER-TODO-001, SPEC-INFINITE-GOAL-001, SPEC-LSEL-DRAIN-STALL-001]
---

# SPEC: lane stall watchdog — self-diagnosis + autonomous resume + unified --auto semantics

## HISTORY

- 0.1.0 — 2026-09-30 — plan-phase artifact set authored (card t1370; worktree
  `.moai/worktrees/t1370`, branch `WT-lane-stall-watchdog`, HEAD `3dd5adf2f`).
  Tier M scope. Operator directive 2026-09-30 ("레인 무한 대기 차단"), dispatched
  to worker-70. Audit lens: `--deep`.

## §A Context

### A.1 Origin

Card t1370 (queued, home queue; operator directive 2026-09-30, worker-70
배차) — card text verbatim:

> 레인 무한 대기 차단 — 자가 진단+자율 재개 워치독+--auto 의미 통일. 실측(09-30
> 판): t1338 blocked-by 대기(블로커 해제 감지 없음)·t1344 plan-audit 확보
> 대기·t1311 리드 콜백 대기(답장 의존)·t1365 셸 오류 exit 128 정지(재시도
> 판정자 없음)·worker-69 /loop 재각성 후 재대기. 현황: todo --auto는 큐 표면
> 감시만(--auto-wait 30m unpick), 레인 세션 내부 스톨 4유형은 감시자 부재.

The card mandates five design elements, restated as the requirement spine of
this SPEC: (1) a progress definition with N-minute no-progress detection, (2) a
per-cause autonomous remedy table, (3) one document unifying `--auto` semantics
across the three surfaces, (4) human gates kept as gates with explicit-wait
instead of infinite waiting, (5) Jev kept display-only with its auxiliary-signal
limit documented.

### A.2 Measured stall inventory (2026-09-30, this tree HEAD `3dd5adf2f`)

| # | Observed stall (card) | What the lane waited on | Class in §B.2 |
|---|---|---|---|
| 1 | t1338 blocked-by wait | a predecessor card; no unblock detection | blocked-by |
| 2 | t1344 plan-audit availability wait | a chain actor (plan-auditor) becoming available | awaited-actor |
| 3 | t1311 lead-callback wait | a reply message from the leader session | awaited-actor |
| 4 | t1365 shell error (exit 128) halt | nothing — the lane stopped on the error | shell-error |
| 5 | worker-69 `/loop` re-awaken then re-stall | the awaken turn carried no diagnosis, so the lane resumed waiting the same way | awaken-path |

The structural diagnosis (lead, code-confirmed; re-verified by this plan
phase): the THREE `--auto` surfaces carry divergent semantics and only the lane
surface has no watchdog at all.

### A.3 Three-surface `--auto` inventory (code-verified, this tree)

| Surface | What `--auto` means there | Stall watchdog present? | Evidence |
|---|---|---|---|
| `todo --auto` | queue serial cycle: pick → dispatch → judge by evidence file at a per-card deadline → done/unpick | YES — `--auto-wait` default 30 min evidence deadline, unpick on absence, dead-owner rescue | `internal/cli/todo.go:313` (default `30*time.Minute`), `internal/cli/todo_auto.go:17-27` (cycle contract), `todo_auto.go:276-283` (deadline poll), `todo_auto.go:314-328` (unpick), `todo_auto.go:142-161` (dead-owner rescue) |
| lane (factory / kanban worker session) | runs one dispatched card in its own worktree; integrates via the serial window | **NO** — the four stall classes of §A.2 have no detector and no remedy path | `gitflow-lane-protocol.md` §6 (lane waits for the leader's next dispatch — open-ended); no stall surface exists anywhere in the lane doctrine |
| `goal --auto` | natural-language autonomous mission (draft → approve → run; blocked → resume) | YES, bounded — turn ceiling, stagnation guard, wall-clock bound; NOT infinite; awakening depends on an explicit resume or a loop | `internal/cli/goal.go:148-154` (mission verbs incl. `resume`), `goal.go:194` (`--auto` flag), `internal/goal/evaluate.go:325` (MaxTurns ceiling), `evaluate.go:337-347` (wall-clock), `evaluate.go:358-363` (stagnation, REQ-GLE-017) |

The divergence is the defect: one word (`--auto`), three meanings, and the
surface where sessions actually stall (the lane) is the one with no watching
mechanism.

### A.4 Doc-vs-code touch surface (the dispatch's stated hazard)

This card edits doctrine documents AND adds one skill; it changes no Go code.
The full enumerated surface:

| Kind | Path | Mirror obligation |
|---|---|---|
| NEW doctrine rule | `.claude/rules/moai/workflow/auto-semantics.md` | template FIRST at `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`, then `make build` |
| NEW skill | `.claude/skills/moai-lane-watchdog/SKILL.md` | template FIRST at `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`, then `make build` |
| EDIT doctrine rule | `.claude/rules/moai/workflow/kanban-dispatch.md` | template mirror in the same change |
| EDIT local-only rule | `.claude/rules/local/gitflow-lane-protocol.md` §6 | **never mirrored** (local-only rules are outside the managed roots, `AGENTS.local.md` §2a) |
| NO change | Go code (`internal/`, `pkg/`, `cmd/`), `.moai/config/sections/`, `.claude/loop.md` | — |

`.claude/loop.md` is template-managed (mirror exists, byte-size 1254 measured)
and deliberately UNTOUCHED: the bare-`/loop` driver is the foreman's
(leader-side) iteration, whose queue-level 30-minute evidence deadline already
covers worker death; the lane-side awaken path is doctrine + skill, not a
second loop driver.

### A.5 Overlap adjudication — SPEC-RELATION-PICKUP-FILTER-001 (card t1343)

t1343 (status: completed; landed in this tree) consumes `blocks`/`depends`
findings as a **pickup filter**: a relation-blocked `queued` card is excluded
from `todo --auto` pickup while the finding exists in the live record
(`internal/cli/todo_auto.go:162-187`, REQ-RPF-001/002/004; the dead-owner
rescue arm is deliberately not gated, REQ-RPF-006). That is a QUEUE-level,
pre-dispatch mechanism.

This SPEC governs the LANE-level, post-dispatch case: a lane already carrying
a card that stalls because the card is blocked by a predecessor. The two
compose without overlap:

- different lifecycle moment — t1343 decides who gets picked; this SPEC decides
  what an in-flight lane does when its own card cannot proceed;
- shared data, no shared logic — both read the same on-disk relation findings
  and evidence paths, but the lane remedy never mutates the queue and never
  picks, drops, or edits cards (queue production and mutation stay the
  operator's and the leader's);
- the lane remedy's unblock signal is the predecessor's on-disk evidence
  (`.moai/reports/<card-id>/`), the same evidence surface the cycle reads.

### A.6 Coordination frame

Card t1370, Class C (design change across doctrine surfaces), procedure
`plan → run → sync`, cycle_type=tdd (`.moai/config/sections/quality.yaml:2`
`development_mode: tdd`, measured). Plan-phase authored in worktree
`.moai/worktrees/t1370` on `WT-lane-stall-watchdog`. No code changes at
plan-phase; the live operator queue (home surface
`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` — the repo-local
`.moai/state/todo/backlog.db` is a 0-byte ghost, measured 2026-09-30) is never
mutated by this card.

## §B Decisions

### B.1 Progress definition (mandated element 1)

Lane progress is measured on exactly three channels: (a) the lane worktree's
HEAD commit SHA, (b) the mtime of the lane's evidence files — the card's
`.moai/reports/<card-id>/` evidence path and the active SPEC's `progress.md`,
(c) the integration-window state (`moai integration status`). A lane is
STALL only when all three channels show no change across an N-minute window.
N defaults to 15 minutes and is a per-invocation parameter of the watchdog
procedure — no new config section (a new `.moai/config` file is wiped by every
`moai update`; `AGENTS.local.md` §2.3).

### B.2 Per-cause remedy table (mandated element 2)

| Cause (from §A.2) | Autonomous remedy |
|---|---|
| awaited-actor (lead callback, plan-audit availability) | re-read the awaited actor's on-disk evidence (`.moai/reports/<card-id>/`, SPEC `progress.md` §E, audit verdict files) and resume from what the evidence shows; a reply's absence is never read as the absence of progress |
| blocked-by | read the blocking predecessor's on-disk evidence directly; resume when it shows the predecessor landed; otherwise record an explicit wait (reason + predecessor id + re-check) and re-check at each awaken |
| shell-error | adjudicate: classify transient vs determinate, retry a transient error at most 3 times (the loop-prevention ceiling) with the error and retry count recorded, then escalate as a structured blocker — never an unbounded halt |
| accidental stop | checkpoint return: resume from the last checkpoint in the card's progress record via the session-handoff resume convention (worktree-anchored Block 0), never restart the card |

### B.3 Execution vehicle — a skill, not a new CLI verb

The watchdog is model-mediated doctrine carried by ONE skill
(`moai-lane-watchdog`), mirroring the proven foreman precedent: `.claude/loop.md`
is a bare-`/loop` iteration whose whole body is "invoke
Skill("moai-kanban-foreman") and follow it" (loop.md:8-13). A new Go verb was
considered and rejected on the simplicity ladder: the remedies ARE model
behaviors (re-read evidence, resume, record a wait) — a binary detector would
add a surface without removing the model step it would still need. No Go code
changes (§A.4).

### B.4 Awaken rule (the measured worker-69 fix)

The unified document carries the awaken rule: **when a loop iteration
re-awakens a lane, the awaken turn runs the watchdog self-diagnosis — measure
progress, classify cause, apply the remedy — BEFORE resuming card work** — plus
the canonical awaken prompt (a paste-able `/loop` prompt invoking the skill).
worker-69's re-stall is exactly an awaken turn without a diagnosis procedure;
the fix is the procedure existing canonically, not a new scheduler.

### B.5 Explicit-wait protocol for human gates (mandated element 4)

Human gates KEEP their semantics — Implementation Kickoff Approval, sync
blocking approval, and operator queue gates are never answered on the
operator's behalf and never bypassed. What changes is the waiting posture:
waiting on a gate is an EXPLICIT wait — a recorded note naming the reason and
the actor waited on, re-checked at each watchdog awaken against the gate's
on-disk state surface — replacing the open-ended stop the lane doctrine
carries today (e.g. gitflow-lane-protocol.md §6 "리더가 다음 카드를 dispatch
할 때까지 기다린다").

### B.6 One `--auto` document (mandated element 3)

`auto-semantics.md` defines `--auto` ONCE for all three surfaces — "minimize
human intervention through autonomous adjudication" — and carries the
per-surface inventory of §A.3 plus the invariants common to all three: human
gates remain gates; queue production stays the operator's; the watchdog never
answers a gate. The document does not rewrite any surface's mechanics; it is
the shared definition each surface's own doctrine links to.

### B.7 Jev stays display-only (mandated element 5)

The watchdog never consumes Jev output as a stall verdict, a remedy choice, or
any queue/gate decision. The unified document restates the display-only
contract (REQ-MT-014/015 lineage, `internal/cli/todo_auto.go:221-226`) and its
limit as an auxiliary decision signal.

### B.8 Queue read-only boundary

The watchdog and the remedies read evidence and state surfaces only. Queue
mutation verbs (`add`, `drop`, `done`, `edit`, `relate`/`unrelate`) and
`moai contract sign` stay prohibited for lanes (gitflow-lane-protocol.md §6);
a blocked-by remedy resolves by evidence, never by relation bookkeeping.

## §C Requirements

Verification layer: `acceptance.md` (RED-now/green pairs; release-blocking
criteria name command + verbatim output + exit code + tree pin). Requirement
layer below is GEARS.

- **REQ-LSW-001** (Ubiquitous) — The lane stall watchdog shall measure lane
  progress by exactly three channels: the lane worktree's HEAD commit SHA, the
  mtime of the lane's evidence files (the card's `.moai/reports/<card-id>/`
  evidence path and the active SPEC's `progress.md`), and the
  integration-window state; and shall classify the lane as stalled only when
  all three channels show no change across the N-minute window (default
  15 minutes, a per-invocation parameter).

- **REQ-LSW-002** (When) — **When** a loop iteration re-awakens a lane session,
  the awaken turn shall run the watchdog self-diagnosis (progress measurement →
  cause classification → remedy application) before resuming card work.

- **REQ-LSW-003** (When) — **When** the watchdog classifies a stall as a wait
  on another session (leader callback, plan-audit availability, chain actor),
  the lane shall re-read the awaited actor's on-disk evidence and resume from
  what the evidence shows; the lane shall not treat the absence of a reply
  message as the absence of progress, and shall not hold the card hostage to a
  reply address.

- **REQ-LSW-004** (When) — **When** the watchdog classifies a stall as a
  blocked-by wait, the lane shall read the blocking predecessor's on-disk
  evidence directly, resume when that evidence shows the predecessor landed,
  and otherwise record an explicit wait (reason + predecessor id + re-check
  point); the lane shall not resolve a block by mutating the queue or by
  picking, dropping, or editing any card — composing with
  SPEC-RELATION-PICKUP-FILTER-001, which owns the pickup-side filter.

- **REQ-LSW-005** (When) — **When** the watchdog classifies a stall as a
  shell-error halt, the lane shall adjudicate the error (transient vs
  determinate), retry a transient error at most 3 times with the error and the
  retry count recorded, and escalate as a structured blocker when the retries
  are exhausted or the error is determinate; the lane shall neither halt
  indefinitely on an unadjudicated error nor loop silently.

- **REQ-LSW-006** (When) — **When** the watchdog finds a lane session stopped
  mid-card, the resume shall follow the checkpoint-return convention — the
  session-handoff resume message with the worktree-anchored Block 0 — resuming
  from the last recorded checkpoint in the card's progress record; the card
  shall not be restarted from zero.

- **REQ-LSW-007** (Ubiquitous) — One doctrine document shall define `--auto`
  once for all three surfaces (`todo --auto`, the lane, `goal --auto`) —
  "minimize human intervention through autonomous adjudication" — and shall
  carry the per-surface inventory (each surface's meaning and its watchdog
  posture) and the invariants common to all three: human gates remain gates,
  queue production stays the operator's, and no watchdog consumes Jev output
  as a decision.

- **REQ-LSW-008** (Ubiquitous) — **Where** a lane waits on a human gate
  (Implementation Kickoff Approval, sync blocking approval, operator queue
  gates), the wait shall be explicit — a recorded wait note naming the reason
  and the actor waited on, re-checked at each watchdog awaken against the
  gate's on-disk state surface — and the gate's own semantics shall remain
  unchanged: the watchdog shall not answer, bypass, or accelerate any gate.

- **REQ-LSW-009** (Unwanted) — The watchdog shall not consume Jev output as a
  stall verdict, a cause classification, a remedy choice, or any queue or gate
  decision; the unified document shall state Jev's display-only contract and
  its limit as an auxiliary signal.

### C.1 Traceability

REQ-LSW-001 → AC-LSW-005 · REQ-LSW-002 → AC-LSW-003 + AC-LSW-007 ·
REQ-LSW-003 → AC-LSW-004a · REQ-LSW-004 → AC-LSW-004b + AC-LSW-006 ·
REQ-LSW-005 → AC-LSW-004c · REQ-LSW-006 → AC-LSW-004d · REQ-LSW-007 →
AC-LSW-001 + AC-LSW-002 · REQ-LSW-008 → AC-LSW-007 + AC-LSW-008 + AC-LSW-009 ·
REQ-LSW-009 → AC-LSW-002 (AC bodies and evidence cells in `acceptance.md`).

## §D Out of Scope

### Out of Scope — Go code and new CLI verbs

- No change to `internal/`, `pkg/`, `cmd/`. No new `moai` verb. The watchdog
  is doctrine + skill (§B.3); a binary detector is rejected on the simplicity
  ladder, and any future mechanical detector is a separate SPEC.

### Out of Scope — the foreman loop and `.claude/loop.md`

- The leader-side bare-`/loop` driver (`.claude/loop.md` + template mirror)
  keeps its foreman semantics; its queue-level 30-minute evidence deadline
  already covers worker death. The lane-side awaken path is doctrine + skill,
  not a second loop driver.

### Out of Scope — queue mutation and pickup semantics

- Pickup selection, relation vocabulary, and the t1343 filter's behavior are
  owned by SPEC-RELATION-PICKUP-FILTER-001. This SPEC adds no pickup logic,
  no relation bookkeeping, and no queue-mutation capability to lanes.

### Out of Scope — gate semantics

- Implementation Kickoff Approval, sync blocking approval, and operator queue
  gates are unchanged: no bypass, no auto-answer, no relaxation. Only the
  waiting posture becomes explicit (§B.5).

### Out of Scope — Jev capability and goal-ceiling tuning

- Jev keeps its display-only role; no decision consumption is added. The
  goal surface's turn ceiling, stagnation guard, and wall-clock bound are
  bounded already (§A.3) and stay untouched.

### Out of Scope — new config surfaces and Codex-lane parity

- No new `.moai/config/sections/` file or key (wiped by `moai update`;
  the watchdog's N is a per-invocation parameter). Codex lanes read
  `AGENTS.md`, not `.claude/rules/`; extending the doctrine to the Codex
  harness surface is a future card (§G R-4).

## §G Gaps and Residual Risks

- **R-1** — The watchdog is model-mediated doctrine, not a mechanical
  enforcer: a lane with no loop armed and no awaken turn still never
  self-diagnoses. The mechanical backstops remain what they are today (the
  foreman's evidence-deadline unpick at queue level, the goal evaluator's
  ceiling at session level). Honest limitation of a workflow-discipline card.
- **R-2** — The N default (15 minutes) is a judgment against observed loop
  cadences, not a measurement; it is a per-invocation parameter precisely so
  the first operational weeks can tune it without a code change.
- **R-3** — Template copies of the new doctrine must stay internally neutral
  (no card ids, no internal dates, no commit SHAs; composition citations with
  internal SPEC IDs live in the local layer and the SPEC artifacts, which are
  local-only). The template-neutrality CI guard is the net; run-phase runs it.
- **R-4** — Codex lanes do not read `.claude/rules/`; the unified semantics
  reach Claude surfaces first. Parity on the `AGENTS.md` side is deferred.
- **R-5** — The explicit-wait re-check needs an on-disk gate state surface;
  where none exists (a gate mid-`AskUserQuestion`), the re-check degrades to
  the reply-independence rule of REQ-LSW-003 — the wait note is then the only
  record, and progress resumes only on evidence.
