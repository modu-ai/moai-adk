---
id: SPEC-SESSION-MIDMOVE-001
title: "Card sessions do not move between worktrees mid-session"
version: "0.1.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, session-start, skill-listing, context-budget, kanban, dispatch"
tier: M
related_specs: [SPEC-SESSION-DOUBLELOAD-001]
---

# SPEC-SESSION-MIDMOVE-001 — Card sessions do not move between worktrees mid-session

## HISTORY

- **2026-09-27** — v0.1.0 plan-phase draft for card t1279. Source of truth: `.moai/reports/t1279/verdict.md` §①–§④ (measured on Claude Code 2.1.283; every load judgment taken from session transcripts, never from debug logs). Scope is the lead's §④ decision: path D only. SPEC-SESSION-DOUBLELOAD-001 stays on HOLD and is not modified; it is listed in `related_specs` because its premise (entry form fixes instruction-file double load) was refuted by verdict §① Claim 2, which is why this SPEC targets the skill listing only.

## §A — Problem

The verdict separates two duplications that had been treated as one:

1. **Instruction files (CLAUDE.local.md).** A session walks up from its start directory and loads every `CLAUDE.md` / `CLAUDE.local.md` it passes, crossing git-repository boundaries (verdict §① Claims 1–2, probes `de2dbb1a`, `45a671b6`, `f592469f`). A session started inside an L1 card worktree under `.claude/worktrees/` therefore also loads the primary checkout's CLAUDE.local.md. The entry form does not change this. The fix is relocation (option A, held) or slimming (option B, owned by t1243/t1259). **Not addressed here.**
2. **Skill listing.** A session that starts in the primary checkout and then moves into a card worktree with `EnterWorktree` receives a second skill listing, scoped to the worktree (about 40 skills), while the primary listing stays in context (verdict §① (2), t1219 Evidence 4). The duplicate is tied to the **mid-session move**, not to where the tree lives. **Addressed here.**

The duplicate is observable in a transcript. A non-initial `skill_listing` attachment in session `bb145fe7` (timestamp `2026-09-27T03:57:49.175Z`, 46,541 B) carries 84 occurrences of the `worktrees/t1219:` prefix across its `names` and `content` fields (observed by reading the transcript's `skill_listing` attachments with `python3`; `isInitial`, `skillCount`, `names`, `content` are the attachment's keys).

Two facts are **not** measured (verdict §① Gaps, §② row D):
- the token cost of the duplicate listing;
- whether `/clear` issued after a move leaves a single listing;
- (also) whether a session started inside the worktree through the launcher carries exactly one listing — observed only through a debug log, which this SPEC does not accept as evidence.

The current doctrine prescribes the move. The kanban dispatch `wt` field, the "new card starts in a new worktree" clause, the dispatch example in its detail companion, the root agent contract §3, and the local lane protocol all direct an in-session `ExitWorktree` → `EnterWorktree` move for a new card, and order the `/clear` **before** that move (affected lines: §D).

## §B — Goal

A card's session starts inside the card worktree through the launcher and does not move to another worktree mid-session. When a move cannot be avoided, the next step after it is `/clear`, followed by the lead re-sending the dispatch pointer. The doctrine says this in every surface that currently prescribes the move, in the local and template copies. The cost of the path being retired is measured, not assumed.

## §C — Requirements (GEARS)

### C.1 Measurement of the mid-session-move path

- **REQ-SMM-001** (Ubiquitous) — The measurement shall commit its caps before the first probe runs, in a commit that carries no probe output. The caps are: at most 6 probe sessions (operator-run sessions included), at most 4 turns per probe, each headless probe bounded by an external `timeout -k 10 300`, and an aggregate wall-clock cap of 45 minutes.
- **REQ-SMM-002** (Ubiquitous) — The measurement shall record, for each of three paths in a scratch fixture repository that carries its own project skills in both the fixture primary and a fixture worktree under `.claude/worktrees/`:
  - **Paths:**
    - (P1) a session started inside the fixture worktree;
    - (P2) a session started in the fixture primary, moved into the fixture worktree by `EnterWorktree`, and continued for at least two further turns;
    - (P3) the P2 session after `/clear`, continued for one turn.
  - **Recorded for each path:**
    - the set of skill-source trees named by the transcript's `skill_listing` attachments (a tree is named by its scope prefix, or by the absence of one for the start tree);
    - the count of `skill_listing` attachments and the byte size of each;
    - per assistant turn, the input-side token total (`input_tokens` + `cache_creation_input_tokens` + `cache_read_input_tokens`) from the transcript usage fields.
- **REQ-SMM-003** (Unwanted) — The measurement shall not take any load or listing judgment from a debug log; every judgment shall be read from a session transcript or its usage fields, and the extraction command shall be recorded beside its output.
- **REQ-SMM-004** (Event-driven) — When a path records a single skill-source tree, or records no worktree-scoped listing, the evidence shall carry a positive control: the same extraction command, run on a committed extract of a transcript known to contain a worktree-scoped `skill_listing`, returns at least one `.claude/worktrees/`-scoped tree.
- **REQ-SMM-005** (Event-driven) — When a headless session cannot perform the move or the `/clear` of a path, or a cap is reached before a path is observed, the measurement shall record that path as a Gap with its reason, and may substitute an operator-run session in the same fixture whose transcript is then read by the same extraction command.
- **REQ-SMM-006** (Unwanted) — The measurement probes shall not:
  - start in, move into, or write to the primary checkout of this repository or any card worktree other than this card's;
  - change branch state anywhere outside the scratch fixture;
  - use `--setting-sources` (it suppresses the CLAUDE.local.md load the fixture depends on);
  - run without the `MOAI_KANBAN*` variables unset in the same compound invocation;
  - commit raw transcripts or debug output; only the extracted lines and the commands that produced them are committed.

### C.2 Doctrine — the entry form and the unavoidable move

- **REQ-SMM-007** (Ubiquitous) — The kanban dispatch rule, its detail companion, the root agent contract §3, the worktree integration rule, and the local lane protocol shall name starting the card's session inside the card worktree through the launcher (`moai cc -w <card-id>`, or `moai cc -w <card-id> --spawn` for a new window) as the entry form for card work.
- **REQ-SMM-008** (Event-driven) — When a card session has to move into another worktree mid-session, the doctrine shall make `/clear` the next step after the move and the lead's re-send of the dispatch pointer the step after that. The kanban dispatch rule's "The `/clear` handoff between phases" clause and its "new card starts in a new worktree" clause shall state this order (move → `/clear` → re-send).
- **REQ-SMM-009** (Unwanted) — The doctrine shall not offer an in-session `EnterWorktree` into a card worktree as an equal alternative to the launcher for starting card work.
- **REQ-SMM-010** (Event-driven) — When the doctrine states what the entry form or `/clear` achieves, the statement shall be bounded by the measurement:
  - it shall claim that the launcher entry yields a single skill listing only when P1 recorded one skill-source tree;
  - it shall claim that `/clear` removes the duplicate listing only when P3 recorded one skill-source tree;
  - otherwise (a second tree, or a Gap) it shall keep the move → `/clear` → re-send order and name ending the session and relaunching it through the launcher as the way to obtain a single listing.
- **REQ-SMM-011** (Where) — Where decision point DP-2 exempts the integration-window move, the kanban dispatch rule shall name the brief move into the integration or release worktree to merge, and the return move to the card worktree, as the only exempt moves, and the lane protocol and CLAUDE.local.md §4.1 shall carry the same exemption. Where DP-2 does not exempt it, the same surfaces shall apply the move → `/clear` → re-send order to those moves as well.

### C.3 Where the doctrine lives

- **REQ-SMM-012** (Ubiquitous) — Template copies shall be edited first. Each local copy that is byte-identical to its template today shall remain byte-identical: kanban dispatch rule, its detail companion, worktree integration rule. The §3 section of `AGENTS.md` shall remain identical to the §3 section of `AGENTS.md.tmpl`. The Codex contract byte ceiling and the always-loaded token budget guards shall still pass.
- **REQ-SMM-013** (Unwanted) — The lines this SPEC adds to any template file shall not carry card ids, SPEC ids, dates, or commit hashes.
- **REQ-SMM-014** (Unwanted) — The doctrine change shall not remove or weaken any existing `[HARD]` clause of the touched files. The `[HARD]` count of each touched file shall not decrease, and these clauses shall survive: launcher-only worktree creation, the `WT-<slug>` rename, "The `/clear` handoff between phases", exit-first before a new card, and "Completion is read, never trusted".
- **REQ-SMM-015** (Ubiquitous) — The local-only lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) and CLAUDE.local.md §4.1 shall state the same entry form and order, and shall remain without a template mirror.

### C.4 Optional guard (conditional on decision point DP-1)

- **REQ-SMM-016** (Where) — Where DP-1 selects a hook guard, the hook layer shall act on every `EnterWorktree` whose target is a card worktree and not an exempt integration worktree:
  - **warn option:** after the move, a notice visible to the session's model shall state that the next step is `/clear` followed by the lead re-sending the pointer; the move itself is not blocked, and the hook exits 0.
  - **block option:** before the move, the hook shall deny it with a reason prefixed by a fixed sentinel, only while a configuration flag is enabled (default disabled), and shall allow the move on any uncertainty (fail-open).

  Where DP-1 selects docs-only, no hook behavior changes.

## §D — Affected Surfaces

Enumerated with `grep -n -E "EnterWorktree|ExitWorktree|moai cc -w|/clear"` on each file at HEAD `2370c5b31`. Template and local copies of the kanban dispatch rule, its detail companion, and the worktree integration rule are byte-identical at that HEAD (`cmp` rc=0 for each pair), so their line numbers match. `AGENTS.md` §3 is identical to `AGENTS.md.tmpl` §3 (`diff` of the extracted sections rc=0); line numbers differ.

| File (template path first; local mirror in parentheses) | Line(s) | Current text (abridged) | Change |
|---|---|---|---|
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (`.claude/rules/moai/workflow/kanban-dispatch.md`) | 93 | `wt` field: exit-first instruction `ExitWorktree` → `EnterWorktree(<card-id>)` → `git branch -m WT-<slug>` | Launcher entry form; move path moved to the unavoidable-move sentence |
| same | 151–157 | "The `/clear` handoff between phases" clause | Add the move → `/clear` → re-send order for card changes; keep every existing sentence |
| same | 166–170 | isolation table: `moai cc -w <name>`, `--spawn`, `EnterWorktree(<path>)` "re-enter from the current session", `ExitWorktree` | Qualify the `EnterWorktree` row: not for starting card work; next step after a move is `/clear` |
| same | 179 | [HARD] "A new card starts in a new worktree — exit any previous one first" (`EnterWorktree(<card-id>)` …) | Launcher start for the new card; exit-first kept; unavoidable move → `/clear` → re-send |
| same | 181 | [HARD] `WT-` rename clause (mentions `EnterWorktree(<name>)` auto-naming) | Unchanged in substance; wording only if needed for the launcher-created branch name |
| same | 250–251 | integration: `EnterWorktree(<release-worktree-path>)`, return `EnterWorktree(<own-path>)` | Per DP-2: exemption sentence, or move → `/clear` → re-send |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` (local mirror) | 27 | glossary "worktree": entered through `moai cc -w` / `moai codex -w` / `EnterWorktree` | Launcher first; `EnterWorktree` not for starting card work |
| same | 28 | glossary "dispatch" example: `wt: EnterWorktree(t0)` | Example uses the launcher form |
| same | 188 | factory: "`/clear` boundary is between cards" | Add: the next card starts through the launcher, or move → `/clear` → re-send |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (local mirror) | 147, 218–220, 224 | `EnterWorktree(<path>)` for current-session re-entry; "Use this when the orchestrator is mid-turn and needs to move the active session" | Card-session caveat: a mid-session move carries the start tree's skill listing alongside; card work starts through the launcher |
| `internal/template/templates/AGENTS.md.tmpl` / `AGENTS.md` | tmpl 118–121, 136–139 / root 109–112, 127–130 | §3 "Work inside a worktree, entered through the launcher … `EnterWorktree(<path>)` to re-enter"; "Start a new card in a new worktree" | Launcher entry for card work; move → `/clear` → re-send |
| `.claude/rules/local/gitflow-lane-protocol.md` (local-only) | 20, 22, 37, 83, 155 | `moai cc -w <card-id>` **또는** `EnterWorktree(<card-id>)`; exit-first; return `EnterWorktree(<card-id>)`; wait at primary for next dispatch; develop refresh `EnterWorktree(<develop>)` | Launcher only for card entry; move → `/clear` → re-send; integration per DP-2 |
| `CLAUDE.local.md` (local-only) | 340 | lane window steps: `EnterWorktree(.claude/worktrees/develop)` … `ExitWorktree keep` | Per DP-2 only |

Baseline counts at `2370c5b31` (positive controls for the acceptance checks): `EnterWorktree(<card-id>)` appears 2 times in the template kanban dispatch rule and 2 times in the lane protocol, 0 times elsewhere in the set; `EnterWorktree(t0)` appears once in the template detail companion. `[HARD]` counts: kanban dispatch 38, detail 8, worktree integration 18, lane protocol 21, CLAUDE.local.md 48, AGENTS.md 0.

Evidence surface: `.moai/reports/t1279/` (gitignored by `.moai/reports/*`; committed files are force-added, as `verdict.md` already is).

## §E — Exclusions (What NOT to Build)

### Out of Scope — Option A, L2 relocation of card worktrees

- Held by the lead (verdict §④). Three blockers were read from code and tool contracts: `moai worktree new` accepts only an L1 leaf name (`internal/cli/worktree/new.go:50-58`), the WorktreeCreate hook is fixed to `.claude/worktrees/` (`internal/hook/worktree_create.go:41`), and `EnterWorktree` from inside a worktree only switches to targets under `.claude/worktrees/`.
- Revival requires both verdict §④ conditions: an authorized L2 creation path (including WorktreeCreate hook support), and an authorized way for a session already inside a worktree to move to another L2 tree or the integration worktree.

### Out of Scope — Option B, slimming the primary CLAUDE.local.md

- Owned by t1243/t1259 (CLAUDE.local.md → AGENTS.local.md, under 40,000 characters). Its landing reduces the ~20.9k-token instruction double load (verdict §① Claim 4). Cross-reference only.

### Out of Scope — Instruction-file double load

- The upward-traversal load of the primary CLAUDE.local.md into L1 worktree sessions is not fixed by the entry form (verdict §① Claim 2). This SPEC neither measures nor claims to reduce it.

### Out of Scope — Claude Code runtime behavior

- No change to how Claude Code discovers skills or instruction files, to `EnterWorktree` / `ExitWorktree`, or to `/clear`. The fix is the entry procedure.
- No change to the `moai cc` launcher. The optional guard (DP-1) acts only through MoAI's own hook handlers.

### Out of Scope — Generic resume and non-card sessions

- `session-handoff.md` Block 0 and `session-handoff-examples.md` keep `EnterWorktree(<path>)` as a current-session re-entry form for generic resumes. They govern non-card sessions; a follow-up card may align them once the measurement lands.

### Out of Scope — Subagent inheritance and exact token budgets

- Which instruction set a subagent carries (verdict §① Gaps) is not measured.
- Token figures are recorded for reference; no threshold is set and no requirement other than REQ-SMM-010's wording depends on them.

## §F — Decision Points for the Orchestrator

Resolved by the orchestrator (with the operator, per the Implementation Kickoff Approval gate) before run-phase entry. Each carries a recommendation; none is left open inside the SPEC body.

### DP-1 — Guard: block, warn, or docs-only

| Option | What it does | Cost | Risk |
|---|---|---|---|
| **Block** (PreToolUse deny, opt-in flag) | Refuses `EnterWorktree` into a card worktree | New PreToolUse matcher in the settings template, a new handler, a config key and default, tests; Tier moves to L (design.md + research.md added before run) | Breaks every sanctioned move: the integration window, resume Block 0 re-entry, and non-card sessions unless each is exempted; a stale exemption list turns into a silent block. A denied move is also recovered by the model trying another route |
| **Warn** (PostToolUse notice) | After the move, tells the model the next step is `/clear` + re-send | Extends the existing PostToolUse `EnterWorktree`/`ExitWorktree` handler (`internal/hook/post_tool_worktree.go`), which already emits a message after every move; one handler plus its test | A notice can be ignored; it cannot prevent the duplicate listing, only shorten how long it stays |
| **Docs-only** | Doctrine changes only | None beyond the doctrine | Nothing reminds a lane at the moment of the move; relies on the lead's dispatch text |

**Recommendation: warn.** The duplicate is created by the move and removed (if at all) by `/clear`; the useful moment to say so is right after the move, and the handler that fires there already exists. Blocking would contradict the sanctioned integration-window and resume moves, and exempting them by path makes the guard's correctness depend on a list that has to track worktree naming. Docs-only is acceptable if the orchestrator wants no Go change in this card; the warn can follow in a separate card.

### DP-2 — Integration-window moves: exempt or covered

The kanban dispatch rule (lines 250–251) and CLAUDE.local.md §4.1 prescribe a mid-session move into the integration/release worktree to merge, and a move back. The lead's scope bans mid-session moves but does not name this flow.

- **Exempt (recommended):** the move is short, does no card work in the moved-into tree, and ends with a return; requiring `/clear` + re-send twice per integration would discard the lane's context mid-window. The exemption is stated once and limited to those two moves.
- **Covered:** the same move → `/clear` → re-send order applies; the integration steps are re-sent after each `/clear`.

### DP-3 — Ordering against the rules-diet card t1175

`WT-rules-diet` (tip `3a48485af`) is not an ancestor of this branch (`git merge-base --is-ancestor WT-rules-diet HEAD` exit 1), and its diff against this branch rewrites the kanban dispatch rule (78 lines) and `AGENTS.md` (76 lines).

- **Wait (recommended):** M1 (measurement) runs now; the doctrine milestones start only after t1175 lands on develop and develop is absorbed here, with the landed commit's SHA recorded in the evidence.
- **Proceed:** edit now and resolve the conflict when t1175 merges; the conflict falls on the same `[HARD]` clauses REQ-SMM-014 protects.
