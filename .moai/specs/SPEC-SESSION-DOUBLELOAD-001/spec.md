---
id: SPEC-SESSION-DOUBLELOAD-001
title: "Card sessions start inside their worktree so instruction files and skills load once"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
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

## §A — Problem

A Claude Code session loads project instruction files (CLAUDE.md, CLAUDE.local.md, `.claude/rules/**`) and the project skills directory from the tree it **starts** in.

Measured (verdict §(1)(2) Evidence):

1. A headless session started with its cwd inside a card worktree loaded exactly one project skills directory — the worktree's (`probe-skills-dirs.txt`) — and the worktree copy of CLAUDE.local.md.
2. A session started in the primary checkout that then moved with `EnterWorktree` had the primary copies loaded at start, and on reading a file inside the worktree additionally received the worktree copies of nested instruction files (observed: the worktree copy of `.claude/rules/local/gitflow-lane-protocol.md` attached in full).
3. The two copies of CLAUDE.local.md differ (52,280 B vs 62,301 B), so the second path carries two divergent instruction sets at once, plus the extra tokens.

Not measured (verdict § Gaps): whether the skill list is duplicated on the `EnterWorktree` path; the token cost of the double load; whether the `moai cc -w <name>` launcher itself loads one set; and whether `/clear` after a mid-session move restores a single set.

The current dispatch doctrine prescribes exactly the double-loading path: the `wt:` field guidance and the "new card starts in a new worktree" clause in `kanban-dispatch.md` tell a lane to `ExitWorktree` → `EnterWorktree(<card-id>)` inside its running session.

## §B — Goal

Card sessions carry one instruction set and one skill set — the card worktree's — for the whole card, and the doctrine that dispatches card work says how to achieve that. The doctrine change is made only after the unmeasured premises are measured.

## §C — Requirements (GEARS)

### C.1 Measurement before doctrine

- **REQ-SDL-001** (Ubiquitous) — The measurement milestone shall complete and commit its evidence before any doctrine file named in §D is edited.
- **REQ-SDL-002** (Ubiquitous) — The measurement shall record, for each session-start path — (A) started inside the card worktree through the launcher form, (B) started inside the card worktree headless, (C) started in the primary checkout then moved by `EnterWorktree` and then reading a worktree file — the set of project skill directories loaded, the set of instruction-file paths loaded, and the prompt-token total of one fixed probe turn.
- **REQ-SDL-003** (Ubiquitous) — The measurement shall declare its caps in the evidence file before the first probe runs: at most 8 probe sessions, at most 4 turns per probe, every probe bounded by an external `timeout` of 300 seconds, and an aggregate wall-clock cap of 45 minutes across all probes.
- **REQ-SDL-004** (Event-driven) — When a cap is reached before an item of REQ-SDL-002 or REQ-SDL-005 is observed, the measurement shall stop and record that item as a Gap rather than as a result.
- **REQ-SDL-005** (Ubiquitous) — The measurement shall determine whether `/clear` issued after a mid-session move into a card worktree leaves only the worktree's instruction files and skill directory loaded.
- **REQ-SDL-006** (Unwanted) — The measurement probes shall not change branch state in the primary checkout, shall not write to any card worktree other than the one named for the probe, and shall not leave a spawned process running without an external bound.
- **REQ-SDL-007** (Event-driven) — When a zero-count result is recorded (for example, no second skills directory on a path), the evidence shall carry a positive control run whose same query returns a non-zero count.

### C.2 Doctrine decision

- **REQ-SDL-008** (Where) — Where the measurement shows the launcher path (A) loads a single instruction set and a single project skills directory, the dispatch doctrine shall name starting the card session inside the card worktree through the launcher (`moai cc -w <name>`, or `moai cc -w <name> --spawn` for a new window) as the primary entry form for card work.
- **REQ-SDL-009** (Where) — Where the measurement shows `/clear` after a mid-session move restores a single set, the dispatch doctrine shall name `/clear` immediately after the move as the fallback entry form; where it does not, or the result is a Gap, the doctrine shall name ending the session and relaunching through the launcher as the fallback instead.
- **REQ-SDL-010** (Event-driven) — When the measurement contradicts a premise of §A (path C shows no second instruction set and no second skills directory), the doctrine files shall not be edited for that premise and the contradiction shall be reported in the evidence file.

### C.3 Where the doctrine lives

- **REQ-SDL-011** (Ubiquitous) — The `wt` field guidance and the "new card starts in a new worktree" clause of the kanban dispatch rule shall state the primary and fallback entry forms in place of the current in-session `ExitWorktree` → `EnterWorktree` sequence.
- **REQ-SDL-012** (Ubiquitous) — The distributed template copy of the kanban dispatch rule shall be edited first and the local copy shall remain byte-identical to it.
- **REQ-SDL-013** (Unwanted) — The distributed template copy shall not carry card ids, SPEC ids, dated incident references, or commit hashes.
- **REQ-SDL-014** (Ubiquitous) — The local-only lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) shall state the same primary and fallback entry forms at its card-worktree entry clause, and shall remain without a template mirror.
- **REQ-SDL-015** (Unwanted) — The doctrine change shall not remove or weaken any existing `[HARD]` clause of the kanban dispatch rule, including launcher-only worktree creation, exit-first before a new card, and the `WT-<slug>` branch rename.

## §D — Affected Surfaces

- `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (template, edited first)
- `.claude/rules/moai/workflow/kanban-dispatch.md` (local mirror, byte-identical)
- `.claude/rules/local/gitflow-lane-protocol.md` (local-only)
- `.moai/reports/t1219/` (measurement evidence)

## §E — Exclusions (What NOT to Build)

This SPEC is a measurement plus a doctrine edit. The items below are out of scope.

### Out of Scope — CLAUDE.local.md behavior contradictions

- The four contradictions recorded in verdict §(3) are already fixed in commit `27aef8547`; this SPEC does not reopen them.

### Out of Scope — Lead-only injection of the kanban dispatch rule

- Moving `kanban-dispatch.md` to a `paths:` scope and injecting it only for the lead role through SessionStart is a proposal recorded in verdict §(4). It belongs to a separate card after the always-loaded rules diet lands.

### Out of Scope — Integration-window worktree moves

- The brief `EnterWorktree` into the release or integration worktree to merge a card branch, and the `EnterWorktree` back into the card tree afterwards, keep their current wording. Their double-load cost is recorded by the measurement if observed, but their doctrine is not changed here.

### Out of Scope — Runtime or CLI changes

- No change to the `moai cc` launcher, to `EnterWorktree` behavior, to hooks, or to any Go code. The fix is the entry procedure, not the loader.
- No new mechanical gate or hook that enforces the entry form.

### Out of Scope — Exact token budgets

- The SPEC records measured token totals; it sets no token threshold that the doctrine must meet.
