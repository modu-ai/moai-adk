---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Card sessions start inside their worktree so instruction files and skills load once"
version: "0.2.0"
status: draft
created: 2026-09-26
updated: 2026-09-27
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, session-start, context-budget, kanban, dispatch, double-load"
tier: M
related_specs: [SPEC-SESSION-WORKTREE-001]
---

# SPEC-SESSION-DOUBLELOAD-001 — Card sessions start inside their worktree

## HISTORY

- **2026-09-26** — v0.1.0 plan-phase draft authored for card t1219. Premises taken from the card's measured verdict (`.moai/reports/t1219/verdict.md`, measured on Claude Code 2.1.283 against develop `e464fd5d0`). The CLAUDE.local.md contradiction half of the card is already closed in commit `27aef8547` and is excluded here.
- **2026-09-27** — v0.2.0 revision after plan-audit iter-1 (FAIL 0.64, `.moai/reports/t1219/plan-audit.md`, audited at `2fdd1f8c9`). All defects D1–D20 addressed. Operator decisions recorded (§F): two-tier entry structure, M3/M4 gated behind t1175 landing on develop, probes run in the t1219 worktree under isolation conditions. Premise updated with verdict Evidence 4 (skill-list double listing observed on the move path). Paths D and D′ (subagents) added to the measurement.

## §A — Problem

A Claude Code session loads project instruction files (CLAUDE.md, CLAUDE.local.md, `.claude/rules/**`) and the project skills from the tree it **starts** in.

Observed (verdict § Evidence, and plan-audit D11):

1. **Started inside the worktree (headless).** The session loaded exactly one project skills directory, the worktree's (`probe-skills-dirs.txt`). The worktree copy of CLAUDE.local.md was observed loading (`probe-memory-files.txt`); that file is a truncated hook message, so the absence of the primary copy is not established by it.
2. **Started in the primary checkout, then moved with `EnterWorktree`.**
   - The primary instruction files were loaded at start.
   - Reading a worktree file then attached the worktree copies of nested instruction files (verdict Evidence 2).
   - As turns continued, a second skill list was attached: about 40 skills prefixed `.claude/worktrees/t1219:`, while the primary list stayed in place (verdict Evidence 4).

   Instruction files and skills are therefore both carried twice on this path. Because the two CLAUDE.local.md copies differ (52,280 B vs 62,301 B), the session also holds two divergent instruction sets. The token cost is not measured.
3. **Subagents.** A plan-auditor subagent spawned by a primary-started parent, with its cwd in the worktree, carried instruction headers from primary paths only, while its agent-memory path resolved to the worktree (plan-audit D11). This is one observation, from context headers that may not list the full load set.

Not measured:
- the token cost of the double load;
- whether the `moai cc -w <name>` launcher itself loads one set;
- whether `/clear` after a mid-session move restores a single set;
- which set a subagent carries, for either parent.

The current dispatch doctrine prescribes the double-loading path. The `wt` field guidance and the "new card starts in a new worktree" clause in `kanban-dispatch.md`, the dispatch example in `kanban-dispatch-detail.md`, and the new-card paragraph in `AGENTS.md` all direct an in-session `ExitWorktree` → `EnterWorktree` move. The `/clear` handoff clause clears the lane **before** that move.

## §B — Goal

Card sessions carry one instruction set and one skill set — the card worktree's — for the whole card. The dispatch doctrine says how to achieve that on both kinds of session:
- sessions opened fresh for a card;
- long-lived kanban and factory lanes that carry many cards.

The doctrine changes only after the unmeasured premises are measured.

## §C — Requirements (GEARS)

### C.1 Measurement before doctrine

- **REQ-SDL-001** (Ubiquitous) — The measurement milestone shall complete and commit its evidence before any doctrine file named in §D is edited.
- **REQ-SDL-002** (Ubiquitous) — The measurement shall record the following for each of five session-start paths:
  - **Paths:**
    - (A) started inside the card worktree through the launcher form;
    - (B) started headless inside the card worktree;
    - (C) started in the primary checkout, moved by `EnterWorktree`, then reading a worktree file;
    - (D) a subagent spawned by a parent that started inside the card worktree;
    - (D′) a subagent spawned by a parent that started in the primary checkout, with the subagent's cwd set to the card worktree.
  - **Recorded for each path:**
    - the set of absolute paths of project skill sources;
    - the set of absolute paths of loaded instruction files;
    - a reference-only prompt-token figure for the first assistant turn after the session or subagent is inside the worktree.
- **REQ-SDL-003** (Ubiquitous) — The measurement shall commit its caps before the first probe runs, in a commit that carries no probe output. The caps are: at most 8 probe sessions, at most 4 turns per probe, every probe bounded by an external `timeout` of 300 seconds, and an aggregate wall-clock cap of 45 minutes.
- **REQ-SDL-004** (Event-driven) — When a cap is reached before an item of REQ-SDL-002 or REQ-SDL-005 is observed, the measurement shall stop and record that item as a Gap rather than as a result.
- **REQ-SDL-005** (Ubiquitous) — The measurement shall determine, by a stated method, whether `/clear` issued after a mid-session move into a card worktree leaves only the worktree's instruction files and skills loaded.
- **REQ-SDL-006** (Unwanted) — The measurement probes shall not:
  - change branch state in the primary checkout;
  - write to any worktree other than the t1219 worktree and the primary checkout's own `.moai/` runtime state;
  - leave a process of their own process group running;
  - run without the `MOAI_KANBAN*` variables scrubbed in the same compound invocation.
- **REQ-SDL-007** (Event-driven) — When a zero-count result is recorded, the evidence shall carry a positive control. The control runs the same extraction command on a known input that holds two sources of the same shape, and returns `count=2`.

### C.2 Doctrine decision (two-tier)

- **REQ-SDL-008** (Where) — Where the measurement shows a single worktree instruction set and a single worktree skill set, the dispatch doctrine shall name starting inside the card worktree through the launcher (`moai cc -w <name>`, or `moai cc -w <name> --spawn` for a new window) as the primary entry form for **sessions opened fresh for a card**. The measurement is taken on path A, or on path B when A is a Gap and B is declared as its proxy.
- **REQ-SDL-009** (Where) — Where the measurement shows that `/clear` after a mid-session move restores a single worktree set, the dispatch doctrine shall name this standard sequence for **long-lived lanes**: move into the card worktree → operator `/clear` → lead re-sends the dispatch pointer. Where it does not, or the result is a Gap, the doctrine shall name ending the lane session and relaunching it through the launcher.
- **REQ-SDL-010** (Event-driven) — When the measurement contradicts a premise of §A for a class (instruction files or skills) — path C shows that class loaded once — the doctrine shall not be edited on account of that class, and the contradiction shall be recorded in the evidence file.

### C.3 Where the doctrine lives

- **REQ-SDL-011** (Ubiquitous) — The kanban dispatch rule shall state the two tiers of REQ-SDL-008/009 in three places:
  - its `wt` field guidance;
  - its "new card starts in a new worktree" clause;
  - its "The `/clear` handoff between phases" clause.

  Where the lane standard applies, the `/clear` handoff clause shall order a card change as: move → `/clear` → re-send pointer.
- **REQ-SDL-012** (Ubiquitous) — The following shall state the same entry form: the dispatch example in the kanban dispatch detail companion, and the "Start a new card in a new worktree" paragraph of the root agent contract. Template copies shall be edited first. Each local copy shall remain byte-identical to its template, and the root agent contract paragraph shall be identical in `AGENTS.md` and `AGENTS.md.tmpl`.
- **REQ-SDL-013** (Unwanted) — The lines this SPEC adds to any template file shall not carry card ids, SPEC ids, dated incident references, or commit hashes.
- **REQ-SDL-014** (Ubiquitous) — The local-only lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) shall state the same two tiers at its card-worktree entry clause. It shall no longer offer an in-session `EnterWorktree(<card-id>)` as an equal alternative to the launcher, and it shall remain without a template mirror.
- **REQ-SDL-015** (Unwanted) — The doctrine change shall not remove or weaken any existing `[HARD]` clause of the kanban dispatch rule. This includes launcher-only worktree creation, the `WT-<slug>` rename, the `/clear` handoff between phases, and exit-first before a new card. Exit-first is read as ending the previous card's session in the primary entry form, and as `ExitWorktree` in the lane standard.
- **REQ-SDL-016** (Unwanted) — The doctrine edits (M3, M4) shall not begin before the rules-diet card t1175 has landed on develop and that develop has been absorbed into this branch.

## §D — Affected Surfaces

| Surface | Path | Kind |
|---|---|---|
| T | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` | template, edited first |
| L | `.claude/rules/moai/workflow/kanban-dispatch.md` | local mirror of T |
| DT | `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` | template, edited first |
| DL | `.claude/rules/moai/workflow/kanban-dispatch-detail.md` | local mirror of DT |
| AT | `internal/template/templates/AGENTS.md.tmpl` | template, edited first |
| A | `AGENTS.md` | root agent contract |
| G | `.claude/rules/local/gitflow-lane-protocol.md` | local-only |
| E | `.moai/reports/t1219/` | measurement evidence |

## §E — Exclusions (What NOT to Build)

This SPEC is a measurement plus a doctrine edit. The items below are out of scope.

### Out of Scope — CLAUDE.local.md behavior contradictions

- The four contradictions recorded in verdict §(3) are already fixed in commit `27aef8547`; this SPEC does not reopen them.

### Out of Scope — Lead-only injection of the kanban dispatch rule

- Moving `kanban-dispatch.md` to a `paths:` scope and injecting it only for the lead role through SessionStart is a proposal recorded in verdict §(4). It belongs to a separate card after the always-loaded rules diet lands.

### Out of Scope — worktree-integration.md

- `worktree-integration.md` § `EnterWorktree` / `ExitWorktree` Tools describes what the runtime tools do for any session; it does not prescribe the card-entry procedure. The card-entry procedure is owned by the dispatch rule and the root agent contract, which this SPEC edits. Left unchanged.

### Out of Scope — Integration-window worktree moves

- The brief `EnterWorktree` into the release or integration worktree to merge a card branch, and the `EnterWorktree` back into the card tree afterwards, keep their current wording. If the measurement observes their double-load cost, it records it, but their doctrine does not change here.

### Out of Scope — Runtime or CLI changes

- No change to the `moai cc` launcher, to `EnterWorktree` behavior, to hooks, or to any Go code. The fix is the entry procedure, not the loader.
- No new mechanical gate or hook that enforces the entry form.

### Out of Scope — Exact token budgets

- The SPEC records reference-only token figures; no decision depends on them and no token threshold is set.

## §F — Operator Decisions (recorded 2026-09-27)

- **Lane lifecycle: two tiers.** Sessions opened fresh for a card use launcher start. Long-lived kanban and factory lanes use move → `/clear` → re-send pointer, which moves the `/clear` lanes already do rather than adding one. That standard holds only if M1 measures that `/clear` restores a single load; otherwise lanes relaunch.
- **Ordering against t1175:** M1 and M2 run now. M3 and M4 wait until t1175 lands on develop and develop has been absorbed (REQ-SDL-016).
- **Probe tree:** probes run in the t1219 worktree under the isolation conditions in plan.md §D.
