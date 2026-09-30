---
id: SPEC-LANE-STALL-WATCHDOG-001
title: "Lane stall watchdog — self-diagnosis, decision-ladder self-resume, unified --auto semantics, and autonomous gate transition across the Claude and Codex runners"
version: "0.2.0"
status: draft
created: 2026-09-30
updated: 2026-09-30
author: manager-spec (card t1370)
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow, .claude/skills/moai-lane-watchdog, internal/template/templates, .claude/rules/local"
lifecycle: spec-anchored
tags: "watchdog, stall, lane, decision-ladder, auto-semantics, harness-neutral, gate-inventory, jev-ask, update-plan, card-t1370"
tier: M
card: t1370
related_specs: [SPEC-RELATION-PICKUP-FILTER-001, SPEC-MANAGER-TODO-001, SPEC-INFINITE-GOAL-001, SPEC-CODEX-SESSION-MSG-001, SPEC-LSEL-DRAIN-STALL-001]
---

# SPEC: lane stall watchdog — self-diagnosis + decision-ladder resume + unified --auto semantics

## HISTORY

- 0.1.0 — 2026-09-30 — plan-phase artifact set authored (card t1370; worktree
  `.moai/worktrees/t1370`, branch `WT-lane-stall-watchdog`, HEAD `3dd5adf2f`).
  Tier M scope. Operator directive 2026-09-30, worker-70 배차. Audit lens:
  `--deep`.
- 0.2.0 — 2026-09-30 — plan-phase iter-2 revision (operator feedback, four
  directives relayed by the lead; full record in `progress.md` §D):
  (1) queue surface (`todo --auto` batch cycle) EXCLUDED from coverage — the
  30-minute unpick watchdog is design precedent only; (2) primary goal
  pivoted to REMOVING lead-response waiting as a stall cause via a five-step
  decision ladder (disk evidence → decision board → MCP audit cross → jev_ask
  → lead chat last resort); (3) harness neutrality REQ added (moai cc / glm /
  codex, per-step carrier table, codex substitutes explicit); (4) gate
  inventory + default-autonomous transition with a three-category keep-set,
  decision auditability, and a doc-code amendment list; Jev surface corrected
  to the `jev_ask` MCP tool (`scripts/jev/` measured ABSENT in this tree —
  the earlier ask.sh premise was stale); codex `update_plan` recorded as the
  native task view (lead-reported external measurement).

## §A Context

### A.1 Origin

Card t1370 (queued, home queue; operator directive 2026-09-30, worker-70
배차) — card text verbatim:

> 레인 무한 대기 차단 — 자가 진단+자율 재개 워치독+--auto 의미 통일. 실측(09-30
> 판): t1338 blocked-by 대기(블로커 해제 감지 없음)·t1344 plan-audit 확보
> 대기·t1311 리드 콜백 대기(답장 의존)·t1365 셸 오류 exit 128 정지(재시도
> 판정자 없음)·worker-69 /loop 재각성 후 재대기. 현황: todo --auto는 큐 표면
> 감시만(--auto-wait 30m unpick), 레인 세션 내부 스톨 4유형은 감시자 부재.

Iter-2 refines the card's intent under operator feedback: the t1311
lead-callback wait is not to be watched — it is to be ELIMINATED as a wait by
giving the lane a decision ladder that resolves judgments without a reply
(§B.2). The queue surface (`todo --auto`) is operator-unused and leaves
coverage; its watchdog pattern remains citable as precedent (§D).

### A.2 Measured stall inventory (2026-09-30, this tree HEAD `3dd5adf2f`)

| # | Observed stall (card) | What the lane waited on | Disposition in this SPEC |
|---|---|---|---|
| 1 | t1338 blocked-by wait | a predecessor card; no unblock detection | blocked-by remedy (REQ-LSW-004) |
| 2 | t1344 plan-audit availability wait | a chain actor becoming available | ladder step ① (disk verdict) + ③ (audit cross) |
| 3 | t1311 lead-callback wait | a reply message from the leader session | REMOVED as a stall cause — ladder REQ-LSW-003 |
| 4 | t1365 shell error (exit 128) halt | nothing — the lane stopped on the error | shell-error remedy (REQ-LSW-005) |
| 5 | worker-69 `/loop` re-awaken then re-stall | the awaken turn carried no diagnosis | awaken-path rule (REQ-LSW-002) |

The structural diagnosis (lead, code-confirmed; re-verified by this plan
phase): three `--auto` surfaces with divergent semantics, and the lane
surface — where sessions actually stall — had no watchdog at all. Iter-2
narrows the COVERAGE to the lane surface (§A.3, §D).

### A.3 Surface scoping after iter-2 (code-verified, this tree)

| Surface | Status in this SPEC | Evidence |
|---|---|---|
| lane (factory / kanban worker session, `moai cc` / `moai glm` / `moai codex`) | **COVERED** — the watchdog + ladder + doctrine | this SPEC; the open-ended wait posture is doctrine text (`gitflow-lane-protocol.md` §6) |
| `todo --auto` (queue batch cycle) | **OUT OF SCOPE** — operator-unused; inventoried as a boundary surface; its boundedness cited as design precedent only | `internal/cli/todo.go:313` (`--auto-wait` default `30*time.Minute`), `todo_auto.go:17-27, 276-328` (deadline + unpick), `todo_auto.go:142-161` (dead-owner rescue) |
| `goal --auto` (natural-language mission) | **OUT OF SCOPE** — already bounded; inventoried as a boundary surface | `internal/cli/goal.go:148-154, 194`; `internal/goal/evaluate.go:325, 337-347, 358-363` (ceiling / wall-clock / stagnation) |

The unification document (REQ-LSW-007) is therefore LANE-CENTERED: it defines
`--auto` for the lane surface and states the other two surfaces' boundaries —
not a rewrite of their mechanics.

### A.4 Doc-vs-code touch surface (the dispatch's stated hazard)

| Kind | Path | Mirror obligation |
|---|---|---|
| NEW doctrine rule | `.claude/rules/moai/workflow/auto-semantics.md` | template FIRST at `internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md`, then `make build` |
| NEW skill | `.claude/skills/moai-lane-watchdog/SKILL.md` | template FIRST at `internal/template/templates/.claude/skills/moai-lane-watchdog/SKILL.md`, then `make build` |
| EDIT doctrine rules | `.claude/rules/moai/workflow/kanban-dispatch.md` + the amendment list of §A.7 (each named file mirrored) | template mirror in the same change |
| EDIT local-only rule | `.claude/rules/local/gitflow-lane-protocol.md` §6 | **never mirrored** (local-only rules are outside the managed roots, `AGENTS.local.md` §2a) |
| NO change | Go code (`internal/`, `pkg/`, `cmd/`), `.moai/config/sections/`, `.claude/loop.md` (+ mirror) | — |

`.claude/loop.md` is template-managed (mirror exists, 1254 bytes measured) and
deliberately UNTOUCHED: the bare-`/loop` driver is the leader-side foreman
iteration; the lane-side awaken path is doctrine + skill, not a second loop
driver. The zero-Go boundary holds through iter-2: the autonomous transition
needs no code change because autonomous decisions record on the disk decision
board, NOT through `factory decide` — which stays a human-decider-only surface
(§A.7, §D).

### A.5 Overlap adjudication — SPEC-RELATION-PICKUP-FILTER-001 (card t1343)

t1343 (status: completed; landed — filter code present in this tree at
`internal/cli/todo_auto.go:162-187`, REQ-RPF-001/002/004; dead-owner rescue
ungated per REQ-RPF-006) owns the QUEUE-level pickup decision. This SPEC owns
the LANE-level, post-dispatch remedy. The iter-2 queue-surface exclusion
sharpens the boundary: this card now touches NO part of the queue cycle, so
the two SPECs share zero coverage — only data surfaces (the relation findings
and `.moai/reports/<card-id>/` evidence paths, which the blocked-by remedy
reads directly). The lane remedy performs zero queue mutation and never picks,
drops, edits, or unrelates.

### A.6 Coordination frame

Card t1370, Class C, procedure `plan → run → sync`, cycle_type=tdd
(`.moai/config/sections/quality.yaml:2` `development_mode: tdd`, measured).
Plan-phase authored in worktree `.moai/worktrees/t1370` on
`WT-lane-stall-watchdog` (0.1.0 at `3dd5adf2f`, 0.2.0 revision at the commit
carrying this change). No code changes at plan-phase; the live operator queue
(home surface `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` — the
repo-local `.moai/state/todo/backlog.db` is a 0-byte ghost, measured) is never
mutated by this card. CLI lifecycle verbs are cited per the current binary
(`todo` verbs: add / list / done / next / unpick — "advance" does not exist;
"lane self-advance" is a concept name only, never a verb claim).

### A.7 The amendment list (doc-code consistency, iter-2 directive 4)

Existing doctrine this landing SUPERSEDES or AMENDS — run-phase applies every
row; an unamended conflicting wording is a defect this card does not ship (the
t1343 §B.5 doctrine-text lesson):

| # | Surface | Current wording | Post-landing state |
|---|---|---|---|
| A | `orchestration-mode-selection.md` header + `askuser-protocol.md` (Implementation Kickoff Approval mandatory-restoration) | plan→run entry REQUIRES the operator's Kickoff answer, score-independent | Kickoff transitions to AUTONOMOUS by default: plan-auditor PASS (independent audit cross) + `plan_status: audit-ready` + artifact-hash integrity + no open blockers = entry, with a decision record (REQ-LSW-012). The operator keeps the explicit path for keep-set cases (§B.10) and the contract-signing path stays as an equivalent existing form |
| B | `spec-workflow.md` § Plan to Run + § Plan Audit Gate skip policy | skip-eligibility is an exception path (3 conditions) | the audit-cross path becomes the DEFAULT entry; the skip contract's mechanics (verdict PASS + per-tier score + hash unchanged) are reused AS the evidence criteria |
| C | `run.md` § Run-phase Autonomy + `goal-directive.md` § Hard Preconditions | Kickoff-first wording | entry follows the autonomous transition; `ac_converge` arming unchanged |
| D | `session-handoff.md` § Invariants ("Implementation Kickoff Approval unchanged") | gate named as invariant | invariant reworded to the autonomous transition + keep-set |
| E | `internal/cli/factory_card.go` decide surface (measured `:1469-1485`: `decide <card>...` with `--gate kickoff --choice approve\|reject`, `--gate push`, `--choice resume\|block\|unblock\|abandon`; `DeciderHuman` only; lane refusal REQ-SD-016) | records ONLY human decisions; lanes refused | **UNCHANGED code** — factory decide stays the human-override recorder; autonomous decisions record on the decision board (REQ-LSW-011), which lanes CAN write (progress records), keeping zero Go delta |
| F | `kanban-dispatch.md` § Entry into the board / `moai-kanban-foreman` surfaces | operator pick per card | batch-authorized promotion (the existing `--auto` invocation-as-approval semantics) named as the autonomous form; queue ADMISSION (production) stays the operator's — not a gate on work, out of scope |
| G | `AGENTS.local.md` §29 / `kickoff-autonomy.md` | kickoff autonomous EXCEPT operator gates | updated to the inventory + keep-set model; the three keep categories verbatim |
| H | `contract-autonomy.md` | signed contract is the gate-free entry | named as the existing precedent the default transition generalizes; signing path preserved |

The AskUserQuestion channel monopoly is NOT amended: keep-set gates and
operator dialogue still route through it.

## §B Decisions

### B.1 Progress definition

Lane progress is measured on exactly three channels: (a) the lane worktree's
HEAD commit SHA, (b) the mtime of the lane's evidence files — the card's
`.moai/reports/<card-id>/` evidence path and the active SPEC's `progress.md`,
(c) the integration-window state (`moai integration status`). A lane is STALL
only when all three channels show no change across an N-minute window. N
defaults to 15 minutes and is a per-invocation parameter of the watchdog
procedure — no new config section (a new `.moai/config` file is wiped by every
`moai update`; `AGENTS.local.md` §2.3).

### B.2 The decision ladder (iter-2 primary goal)

When a lane needs a judgment to proceed, it resolves through the ladder IN
ORDER — the lead chat is the LAST step, not the first:

| Step | Carrier | Autonomous? |
|---|---|---|
| ① disk evidence | read the deciding artifact directly: audit verdict files, `progress.md` §E, `plan-audit.md`, evidence paths | yes |
| ② decision board | poll the shared decision store — the lead/orchestrator records judgments there instead of replying; SSOT is a HOME-surface disk store (§B.8) | yes |
| ③ audit cross | `moai` MCP audit tools (`codex_audit` / `glm_audit` / `claude_audit` / `audit_multi` — registrations measured at `internal/cli/mcp_server.go:316, 493`) for an independent second opinion; fail-open: tool absence degrades to the next step | yes |
| ④ jev_ask | the MCP judgment tool (`internal/cli/mcp_server.go:650-658`, wraps `jev.Client.Ask`) for a bounded proceed / retry / wait call — ONLY while `workflow.jev.enabled` is true (shipped default false); repeatable, side-effect-free parameters; never for completion, merge, or queue decisions (§29 boundary persists) | yes, gated |
| ⑤ lead chat | the reply path — only when ①-④ are all unavailable; the wait is explicit: reason + whom + recheck recorded on disk | last resort |

t1311's failure mode — holding a card hostage to a reply address — is
structurally removed: steps ①-④ are readable or invocable without any reply.

### B.3 Execution vehicle — a skill, not a new CLI verb

The watchdog is model-mediated doctrine carried by ONE skill
(`moai-lane-watchdog`), mirroring the foreman precedent (`.claude/loop.md:8-13`
— "invoke Skill("moai-kanban-foreman") and follow it"). A new Go verb was
considered and rejected on the simplicity ladder: the remedies ARE model
behaviors; a binary detector would add a surface without removing the model
step it would still need.

### B.4 Awaken rule (the measured worker-69 fix)

The unified document carries the awaken rule — when a loop iteration
re-awakens a lane, the awaken turn runs the watchdog self-diagnosis (measure
progress → classify cause → apply remedy or ladder) BEFORE resuming card work
— plus the canonical awaken prompt (a paste-able `/loop` prompt invoking the
skill).

### B.5 Gate inventory and the default-autonomous transition (iter-2 directive 4)

The document enumerates the human gates (inventory measured where code-backed):

| Gate | Surface (measured) | Disposition |
|---|---|---|
| factory decide: kickoff approve/reject | `factory_card.go:1469-1485` | AUTONOMOUS (audit cross, §A.7 row A) |
| factory decide: push | same | KEEP — irreversible operation on an external shared system (origin push), keep-set ③ |
| factory decide: resume / block / unblock | same | AUTONOMOUS with decision record; operator `hold` state (card t1308 surface) forces KEEP, keep-set ② |
| factory decide: abandon | same | AUTONOMOUS only when the branch is already integrated (nothing lost); otherwise KEEP — irreversible disposal of the only copy, keep-set ③ |
| plan→run Kickoff (operator direct answer) | `orchestration-mode-selection.md` header, `askuser-protocol.md` | AUTONOMOUS (§A.7 row A) |
| sync blocking approval | `sync-phase-quality-gate.sh` (block under `MOAI_SYNC_GATE_BLOCKING=1`) | AUTONOMOUS — sync-auditor verdict + evidence thresholds adjudicate; external-shared operations in sync (release push/merge) KEEP, keep-set ③ |
| card pick (operator promotion) | `kanban-dispatch.md` § Entry into the board | AUTONOMOUS inside a batch authorization (the `--auto` invocation IS the approval — existing semantics); queue ADMISSION stays the operator's (production, not a gate — out of scope) |
| plan-audit bypass flags (`--skip-audit`, `--ignore-deps`) | `spec-workflow.md` § Gate skip policy | RETIRED into the default path (§A.7 row B) — the audit cross IS the entry evidence |
| Jev capability gate | `workflow.jev.enabled` (default false, measured `internal/config/types.go:522, 783`) | capability switch, not a work gate — listed for completeness; enabling is the operator's |

KEEP is allowed ONLY in three categories: (1) environment-impossible
(execution itself cannot proceed — missing credentials/runner), (2)
operator-held work, (3) irreversible operations touching external shared
systems. Any KEEP outside these three is justified with grounds in the
document.

### B.6 Harness neutrality (iter-2 directive 2)

The watchdog and every ladder step carry a per-runner path — the Claude runner
(`moai cc` / `moai glm` — same runner, different backend) and the codex runner
(`moai codex`):

| Ladder step | Claude runner | codex runner |
|---|---|---|
| ① disk evidence + progress detection (SHA / mtime / window state) | shell + git + stat | same (shell-based; harness-indifferent) |
| ② decision board | disk SSOT (HOME surface); native view: TaskCreate / TaskUpdate | disk SSOT (same store); native view: `update_plan` (lead-reported: OpenAI Codex default TODO tool; schema `update_plan(explanation?, plan: [{step, status}])`) |
| ③ audit cross | `moai` MCP audit tools | same MCP server wires into codex sessions; `codex_audit` is codex-native |
| ④ jev_ask | MCP tool — harness-indifferent | same |
| ⑤ reply path | SendMessage | NO SendMessage — cross-harness notice only via broker polling (`session_msg_register` / `session_msg_send` / poll — `mcp_server.go:550, 566`, SPEC-CODEX-SESSION-MSG-001, landed) + the queue-on-disk delegation channel; the wait reason is recorded on disk |

A step impossible on a runner names its substitute (never implicitly
impossible). The watchdog is awaken-tick based and HOOK-INDEPENDENT by design —
the dual-harness hook-parity precedent (t1099) binds only if a future iteration
adds a hook surface.

### B.7 View–SSOT rule (iter-2 directive 2 + 4)

The decision board's SSOT is a HOME-surface disk store — the primary
checkout's queue record / state directory (a tree-local copy is a ghost, the
same lesson as the queue home db: measured 0-byte `backlog.db` in the repo
tree vs the live home db). Each runner's native task tool is a rendering VIEW
only: Claude lanes render with TaskCreate / TaskUpdate, codex lanes with
`update_plan`. The doctrine codifies: render state with the native view, read
judgments from disk — and lane doctrine directs codex lanes to `update_plan`
(§A.7 row list carries the doc touch). ExecPlan / PLANS.md (long-running plan
pattern) is reference-only and is NOT the decision board.

### B.8 Decision records (auditability)

Every autonomous gate transition and every ladder-resolved judgment writes a
decision record on the disk SSOT naming: who decided (runner + role), on what
evidence (paths + verdict ids), and via which ladder path. The record format
is grep-able and defined in the unified document.

### B.9 Queue read-only boundary

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
  cause classification → remedy or ladder application) before resuming card
  work.

- **REQ-LSW-003** (When) — **When** a lane needs a judgment to proceed, the
  lane shall resolve it through the decision ladder in order — ① direct read
  of the on-disk evidence, ② the shared decision board (disk SSOT, polled),
  ③ an independent audit cross via the `moai` MCP audit tools (fail-open),
  ④ the `jev_ask` MCP tool for a bounded proceed / retry / wait judgment while
  `workflow.jev.enabled` is true, ⑤ the lead chat as last resort with the wait
  reason and recheck recorded on disk — and shall not treat the absence of a
  reply as the absence of progress; waiting on a reply shall thereby cease to
  be a stall cause.

- **REQ-LSW-004** (When) — **When** the watchdog classifies a stall as a
  blocked-by wait, the lane shall read the blocking predecessor's on-disk
  evidence directly, resume when that evidence shows the predecessor landed,
  and otherwise record an explicit wait (reason + predecessor id + recheck
  point); the lane shall not resolve a block by mutating the queue or by
  picking, dropping, or editing any card — composing with
  SPEC-RELATION-PICKUP-FILTER-001, which owns the pickup-side filter of the
  out-of-scope queue surface.

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

- **REQ-LSW-007** (Ubiquitous) — One LANE-CENTERED doctrine document shall
  define `--auto` for the lane surface — "minimize human intervention through
  autonomous adjudication" — and shall inventory `todo --auto` and `goal
  --auto` as boundary surfaces whose mechanics are not covered (the queue
  batch cycle is operator-unused; its 30-minute evidence-deadline unpick is
  cited as design precedent only), and shall carry the decision ladder, the
  harness-neutrality table, the gate inventory with dispositions, the
  decision-record format, and the Jev boundary.

- **REQ-LSW-008** (Ubiquitous) — The document shall enumerate the current
  human gates (the factory decide surface, the plan→run Kickoff, the sync
  blocking approval, card pick, the plan-audit bypass flags) and assign each
  a disposition: default AUTONOMOUS — the human judgment replaced by the
  independent audit cross + evidence criteria + the ladder — or KEEP with a
  stated reason; KEEP is admissible only for environment-impossible work,
  operator-held work, and irreversible operations on external shared systems,
  and any KEEP outside these three categories shall carry its justification
  in the document.

- **REQ-LSW-009** (Unwanted) — Judgment consumption via Jev shall use the
  `jev_ask` MCP tool as the single surface, gated by `workflow.jev.enabled`
  (shipped default false), with repeatable (side-effect-free) parameters, and
  shall be scoped to stall self-resume judgments; it shall not decide
  completion, merge, or queue outcomes, and no `ask.sh` surface shall be
  referenced (the `scripts/jev/` path does not exist in the current tree —
  measured).

- **REQ-LSW-010** (Ubiquitous) — The watchdog and every ladder step shall
  carry a per-runner path for the Claude runner (`moai cc` / `moai glm`) and
  the codex runner (`moai codex`) in the document's neutrality table; a step
  impossible on a runner shall name its substitute path or be declared
  unsupported explicitly — never implicitly impossible; and the watchdog shall
  be hook-independent by design (awaken-tick based), the dual-harness hook
  parity rule (t1099) applying to any future hook surface.

- **REQ-LSW-011** (Ubiquitous) — The decision board's SSOT shall be a
  harness-neutral HOME-surface disk store; each runner's native task tool
  (Claude TaskCreate / TaskUpdate; codex `update_plan`) shall be a rendering
  view only; the view–SSOT rule ("render state with the native view, read
  judgments from disk") shall be codified in the doctrine, and lane doctrine
  shall direct codex lanes to `update_plan`.

- **REQ-LSW-012** (Ubiquitous) — Every autonomous gate transition and every
  ladder-resolved judgment shall produce a decision record on the disk SSOT
  naming who decided, on what evidence, and via which ladder path, in the
  grep-able format the unified document defines.

- **REQ-LSW-013** (Ubiquitous) — The SPEC shall carry the amendment list of
  existing doctrine wordings this landing supersedes (§A.7), and run-phase
  shall apply every listed amendment — a conflicting wording left unamended
  after this card lands is a shipped defect.

### C.1 Traceability

REQ-LSW-001 → AC-LSW-005 · REQ-LSW-002 → AC-LSW-003 + AC-LSW-007b ·
REQ-LSW-003 → AC-LSW-004a + AC-LSW-013 · REQ-LSW-004 → AC-LSW-004b +
AC-LSW-006 · REQ-LSW-005 → AC-LSW-004c · REQ-LSW-006 → AC-LSW-004d ·
REQ-LSW-007 → AC-LSW-001a + AC-LSW-002 · REQ-LSW-008 → AC-LSW-007a +
AC-LSW-014 · REQ-LSW-009 → AC-LSW-002 + AC-LSW-016 · REQ-LSW-010 →
AC-LSW-013 · REQ-LSW-011 → AC-LSW-017 · REQ-LSW-012 → AC-LSW-015 ·
REQ-LSW-013 → AC-LSW-001b (AC bodies and evidence cells in `acceptance.md`).

## §D Out of Scope

### Out of Scope — the queue surface (`todo --auto` batch cycle)

- Operator-unused surface; EXCLUDED from coverage. The queue mechanics,
  pickup selection, and relation vocabulary are owned by
  SPEC-RELATION-PICKUP-FILTER-001 and unchanged. The 30-minute evidence
  deadline + unpick is citable ONLY as a design precedent for the watchdog's
  own timeout shape.

### Out of Scope — `goal --auto` mechanics

- The mission lifecycle (draft → approve → run; blocked → resume) and its
  turn ceiling / stagnation guard / wall-clock bound are bounded already and
  stay untouched; inventoried as a boundary surface only.

### Out of Scope — Go code, CLI verbs, and the factory decide verb

- No change to `internal/`, `pkg/`, `cmd/`; no new `moai` verb. "Lane
  self-advance" is a concept name — the CLI has no such verb (measured:
  `todo` verbs are add / list / done / next / unpick). `factory decide`
  stays the human-decider-only recorder (§A.7 row E); autonomous decisions
  record on the decision board instead, which is why the zero-Go boundary
  survives the autonomous transition.

### Out of Scope — the foreman loop and `.claude/loop.md`

- The leader-side bare-`/loop` driver keeps its foreman semantics; its
  queue-level 30-minute evidence deadline already covers worker death. The
  lane-side awaken path is doctrine + skill, not a second loop driver.

### Out of Scope — queue mutation and admission

- Pickup selection and relation bookkeeping are t1343's. Queue ADMISSION
  (production) is the operator's act, not a gate on work. Lanes never pick,
  drop, edit, unrelate, or admit.

### Out of Scope — keep-set gate mechanics

- The three KEEP categories (environment-impossible, operator-held,
  irreversible external-shared operations) keep their current mechanics;
  only the waiting posture and the inventory documentation change. The
  AskUserQuestion channel monopoly is unchanged.

### Out of Scope — Jev capability expansion and new config

- No new Jev capability beyond consuming the existing `jev_ask` tool; the
  `workflow.jev.enabled` gate stays default-false (enabling is the
  operator's act). No new `.moai/config/sections/` file or key.

### Out of Scope — ExecPlan / PLANS.md

- The Codex long-running ExecPlan pattern is reference-only for plan
  documentation shape; it is NOT the decision board and must not be
  conflated with the view–SSOT design (§B.7).

## §G Gaps and Residual Risks

- **R-1** — The watchdog is model-mediated doctrine, not a mechanical
  enforcer: a lane with no loop armed and no awaken turn still never
  self-diagnoses. The queue-level foreman unpick and the goal evaluator's
  ceiling remain the mechanical backstops. Honest limitation of a
  workflow-discipline card.
- **R-2** — The N default (15 minutes) is a judgment against observed loop
  cadences, not a measurement; it is a per-invocation parameter precisely so
  the first operational weeks can tune it without a code change.
- **R-3** — Template copies of the new doctrine must stay internally neutral
  (no card ids, no internal dates, no commit SHAs; composition citations with
  internal SPEC IDs live in the local layer and `.moai/specs/`, which are
  local-only). The template-neutrality CI guard is the net; run-phase runs it.
- **R-4** — Codex runner tool-surface facts (TaskList absence, `update_plan`
  shape and closure discipline) are LEAD-REPORTED external measurements from
  the OpenAI Codex Prompting Guide — this tree cannot verify them; the
  neutrality table states them as reported premises and M4's review pass
  re-checks them against the then-current Codex documentation. The broker
  path (§B.6 step ⑤) depends on SPEC-CODEX-SESSION-MSG-001 (landed —
  directory present, `session_msg_register`/`session_msg_send` measured at
  `mcp_server.go:550, 566`).
- **R-5** — The explicit-wait recheck needs an on-disk gate state surface;
  where none exists (a gate mid-AskUserQuestion), the recheck degrades to the
  ladder's ①-④ — the wait note is then the only record, and progress resumes
  only on evidence.
- **R-6** — The Kickoff autonomous transition is the highest-stakes doctrine
  change in this card: a defective audit cross would auto-approve an unready
  plan. Mitigation: only a plan-auditor PASS proceeds (FAIL / INCONCLUSIVE
  stay hard blocks — unchanged mechanics), the evidence criteria reuse the
  existing skip-eligibility hash contract, and every transition writes a
  decision record (REQ-LSW-012) that the sync audit can re-read.
- **R-7** — `workflow.jev.enabled` ships false: ladder step ④ is dormant
  until the operator enables it. The ladder degrades fail-open (④ unavailable
  → ⑤), which is the designed behavior, not a defect — but until the gate is
  on, judgments that only ④ could resolve fall to the lead chat.
