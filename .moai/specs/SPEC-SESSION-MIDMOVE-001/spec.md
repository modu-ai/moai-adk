---
id: SPEC-SESSION-MIDMOVE-001
title: "Card sessions do not move between worktrees mid-session"
version: "0.2.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, session-start, skill-listing, context-budget, kanban, dispatch"
tier: L
related_specs: [SPEC-SESSION-DOUBLELOAD-001]
---

# SPEC-SESSION-MIDMOVE-001 — Card sessions do not move between worktrees mid-session

## HISTORY

- **2026-09-27** — v0.1.0 plan-phase draft for card t1279 (Tier M, 16 REQ / 16 AC). Source of truth: `.moai/reports/t1279/verdict.md` §①–§④. Scope is the lead's §④ decision: path D only. SPEC-SESSION-DOUBLELOAD-001 stays on HOLD and is not modified.
- **2026-09-27** — v0.2.0 revision after plan-audit iter-1 (FAIL 0.55, `.moai/reports/t1279/plan-audit.md`, audited at `a14fcf851`). Defects D1–D22 addressed. Re-tiered to **L** (19 REQ, 21 AC; above the Tier M ceiling of 16), so design.md and research.md are added and the plan-auditor PASS threshold becomes 0.85. Changes: ranges use a read-time `CARD_BASE`; an always-loaded net-zero rule; a REQ for the real-session cost of the duplicate; the standing-lane per-card flow; a symmetric DP-1 table.

## §A — Problem

The verdict separates two duplications:

1. **Instruction files (CLAUDE.local.md).** A session walks up from its start directory and loads every `CLAUDE.md` / `CLAUDE.local.md` it passes (verdict §① Claims 1–2). An L1 card worktree under `.claude/worktrees/` therefore also loads the primary checkout's CLAUDE.local.md, whatever the entry form. **Not addressed here** (option A held, option B owned by t1243/t1259).
2. **Skill listing.** A session that moves into a worktree with `EnterWorktree` receives an additional skill listing scoped to that worktree, while the listing it started with stays in context (verdict §① (2)). The duplicate is tied to the **mid-session move**. **Addressed here.**

Observed in the transcript of session `bb145fe7` (research.md §R2 gives the command and full output). Byte basis throughout this SPEC: the UTF-8 length of the `skill_listing` attachment's `content` field.

- The non-initial `skill_listing` at `2026-09-27T03:57:49.175Z` has `skillCount` 60, 42 names carrying the `.claude/worktrees/t1219:` prefix, and a `content` of 43,566 B.
- The input-side usage total (`input_tokens` + `cache_creation_input_tokens` + `cache_read_input_tokens`) of the assistant turns immediately before and after it is 278,878 → 299,707, a delta of 20,829 tokens. The 21 rows between those two turns also carry 20,526 B of other content, so the delta is an **upper bound** on the listing's cost, not the cost itself.
- Two later non-initial listings in the same transcript carry `.claude/worktrees/t1279:` names (55 skills, 45,703 B at `08:46:19.764Z`), so the duplicate recurs on every move.

Not measured:
- whether a session started inside a worktree carries exactly one listing (seen only in a debug log, which this SPEC does not accept);
- whether `/clear` after a move leaves a single listing;
- whether a `systemMessage` hook output reaches the model (the Go type documents it as "shown to user", `internal/hook/types.go:366`).

The current doctrine prescribes the move and orders the `/clear` **before** it (affected lines: §D).

## §B — Goal

- A card session opened fresh for a card starts inside the card worktree through the launcher.
- A standing Kanban or Factory session that carries many cards changes cards with one `/clear` placed **after** the move, and the lead re-sends the dispatch pointer after that `/clear`.
- The doctrine states this in every surface that now prescribes the move, in the local and template copies, without growing the always-loaded surface.
- The mechanism is checked in a fixture, and the cost is read from a real transcript.

## §C — Requirements (GEARS)

### C.1 Measurement of the move path (fixture mechanism check)

- **REQ-SMM-001** (Ubiquitous) — The measurement shall commit its caps before the first probe runs, in a commit that touches only the caps file. The caps are:
  - at most 6 probe sessions, counted as distinct transcript session ids (a `--resume` of the same id is one session; a new id after `/clear` is a new session);
  - at most 4 turns per probe;
  - each headless probe bounded by `timeout -k 10 300`;
  - at most 45 minutes of wall clock, measured from the first row of the first probe transcript to the last row of the last probe transcript.
- **REQ-SMM-002** (Ubiquitous) — The measurement shall record the following for three paths in a scratch fixture repository. The fixture primary carries project skills named `fx-primary-*`, and a fixture worktree under `.claude/worktrees/` carries skills named `fx-wt-*`.
  - **Paths:**
    - (P1) a session started inside the fixture worktree;
    - (P2) a session started at the fixture root, moved into the fixture worktree by `EnterWorktree`, then continued for at least two turns;
    - (P3) the P2 session after `/clear`, continued for one turn.
  - **Recorded for each path:**
    - the fixture-tree set. A listed skill belongs to the primary tree when its base name (after removing any `<path>:` prefix) starts with `fx-primary-`, and to the worktree tree when it starts with `fx-wt-`. Every other name (user, plugin, and command namespaces) is excluded. P3 counts only listings after the `/clear`;
    - the count of `skill_listing` attachments, and the UTF-8 byte length of each one's `content`;
    - per assistant turn, the input-side usage total;
    - the session id, and the `cwd` field of the transcript's rows.
- **REQ-SMM-003** (Unwanted) — The measurement shall not take any load or listing judgment from a debug log. Every judgment shall come from a session transcript through a committed extractor, and the command shall be recorded beside its output.
- **REQ-SMM-004** (Event-driven) — When a path's fixture-tree set holds one tree, or when no path records a worktree-scoped listing, the evidence shall carry two controls, both run through the same extractor:
  - a positive control on a committed JSONL file shaped like a transcript and holding a worktree-scoped `skill_listing`, which reports at least one worktree-scoped name;
  - a negative control on the same shape with no scoped names, which reports none.
- **REQ-SMM-005** (Event-driven) — When a headless session cannot perform a path's move or `/clear`, or a cap is reached first, the measurement shall either record the path as a Gap with its reason, or substitute one operator-run session in the same fixture. The substitute's transcript passes the same extractor, the same `cwd` check, and the same session-id recording.
- **REQ-SMM-006** (Unwanted) — The measurement probes shall not:
  - run with a transcript `cwd` outside the scratch fixture;
  - change branch state outside the fixture;
  - use `--setting-sources`, `-d`, `--debug`, or `--debug-file`;
  - run without the `MOAI_KANBAN*` variables unset in the same compound invocation;
  - commit raw transcripts or debug output.
- **REQ-SMM-007** (Ubiquitous) — The measurement shall keep probe invocations, extraction commands, and fixture construction in three separate committed files, so each can be checked against its own format.

### C.2 Cost of the duplicate listing (real session)

- **REQ-SMM-008** (Ubiquitous) — The evidence shall record the real-session cost of the duplicate listing from the transcript of session `bb145fe7`. For each non-initial `skill_listing` that carries a name with a `.claude/worktrees/<x>:` prefix, it records:
  - the timestamp, `skillCount`, the scoped-name count, and the `content` UTF-8 bytes;
  - the input-side usage totals of the assistant turns immediately before and after it, and their delta;
  - the byte size of the other rows between those two turns.

  The delta shall be labelled an upper bound. When that transcript is no longer readable, the evidence shall record `gap` with the reason, instead of any figure.

### C.3 Doctrine — entry and card change

- **REQ-SMM-009** (Ubiquitous) — The kanban dispatch rule, its detail companion, the root agent contract §3, the worktree integration rule, the local lane protocol, and CLAUDE.local.md §4.1 shall name starting the session inside the card worktree through the launcher (`moai cc -w <card-id>`, or `--spawn` for a new window) as the entry form for a session opened fresh for a card.
- **REQ-SMM-010** (While) — While a standing Kanban companion or Factory lane session carries successive cards, the doctrine shall state the per-card flow in this order:
  1. the lead sends the next card's pointer;
  2. the session moves into the new card worktree;
  3. the operator issues `/clear`;
  4. the lead re-sends the full pointer.

  This `/clear` is the single between-cards `/clear`: it replaces the earlier clear-then-move order and is not added to it. Relaunching the session per card shall not be required.
- **REQ-SMM-011** (Unwanted) — The doctrine shall not present an in-session move into a card worktree without the `/clear` and pointer re-send that follow it. For sessions opened fresh for a card, it shall not offer `EnterWorktree(<card-id>)` as an equal alternative to the launcher.
- **REQ-SMM-012** (Event-driven) — When the doctrine states what a move, the launcher entry, or `/clear` does to the skill listing, the statement shall be bounded by the measurement:
  - "a move adds the moved-into tree's listing" only when P2 recorded two fixture trees;
  - "the launcher entry yields one listing" only when P1 recorded one fixture tree;
  - "`/clear` removes the carried listing" only when P3 recorded one fixture tree.

  When the result is otherwise, including a Gap, the doctrine shall keep the flow of REQ-SMM-010 and name ending the session and relaunching it through the launcher as the way to a single listing, as an option rather than a requirement.
- **REQ-SMM-013** (Where) — Where decision point DP-2 exempts integration-window moves, the kanban dispatch rule, the lane protocol, and CLAUDE.local.md §4.1 shall name exactly two moves as exempt: the move into the integration or release worktree to merge, and the return move to the card worktree. Where DP-2 does not exempt them, those surfaces shall apply the move → `/clear` → re-send order to those moves as well.

### C.4 Where the doctrine lives

- **REQ-SMM-014** (Ubiquitous) — Template copies shall be edited first, and every commit that changes a local mirror shall change its template in the same commit. At HEAD, three local copies shall be byte-identical to their templates: the kanban dispatch rule, its detail companion, and the worktree integration rule. At HEAD, the §3 section of `AGENTS.md` shall be identical to the §3 section of `AGENTS.md.tmpl`.
- **REQ-SMM-015** (Unwanted) — The doctrine change shall not increase the always-loaded surface:
  - the `always-loaded surface = N tokens` figure logged by `TestAlwaysLoadedTokenBudget` after the change shall not exceed the figure logged immediately before the first doctrine commit;
  - the `AlwaysLoadedTokenBudget` constant shall not change;
  - explanatory text shall go to paths-scoped files (the detail companion, the worktree integration rule), and the always-loaded files shall carry only offset-balanced wording and pointers.
- **REQ-SMM-016** (Unwanted) — The lines this SPEC adds to any template file shall not carry card ids, SPEC ids, dates, or commit hashes.
- **REQ-SMM-017** (Unwanted) — The doctrine change shall not remove or weaken an existing `[HARD]` clause of a touched file. In particular:
  - every non-blank line of "The `/clear` handoff between phases" section shall remain verbatim;
  - the `WT-` rename paragraph shall remain verbatim;
  - the new-card paragraph shall keep, verbatim, its bold lead sentence, its fresh-tree sentence, and its merge-not-reuse sentence;
  - the `[HARD]` count of each touched file shall not decrease.
- **REQ-SMM-018** (Ubiquitous) — The local-only lane protocol (`.claude/rules/local/gitflow-lane-protocol.md`) and CLAUDE.local.md §4.1 shall state the entry form of REQ-SMM-009 and the flow of REQ-SMM-010, and shall remain without a template mirror.

### C.5 Optional hook (conditional on decision point DP-1)

- **REQ-SMM-019** (Where) — Where DP-1 selects a hook, the hook shall act only while the session runs in Kanban or Factory mode, and only on `EnterWorktree` targets that are card worktrees and not an exempt integration worktree.
  - **Warn option:** after the move, the hook shall carry the notice (next step `/clear`, then the lead re-sends the pointer) in the `additionalContext` output field, and shall exit 0.
  - **Block option:** before the move, the hook shall deny the move with a reason prefixed by a fixed sentinel, only while a configuration flag is enabled (default disabled), and shall allow the move on any uncertainty.

  Where DP-1 selects docs-only, the hook layer shall not change.

## §D — Affected Surfaces

Enumerated with `grep -n -E "EnterWorktree|ExitWorktree|moai cc -w|/clear"` at HEAD `2370c5b31`. **These line numbers predate the absorption of develop.** Milestone M3 re-greps after absorption and records the refreshed table in the evidence. The t1175 branch rewrites the kanban dispatch rule and `AGENTS.md`, and adds `kanban-dispatch-mechanics.md`. Load scope, measured from the frontmatter: the kanban dispatch rule and `AGENTS.md` are always-loaded (inside the budget surface); the detail companion (`paths:` `**/kanban-dispatch*.md,…`) and the worktree integration rule (`paths:` `**/.claude/worktrees/**,…`) are paths-scoped.

| File (template path; local mirror) | Line(s) at `2370c5b31` | Current text (abridged) | Change | Load |
|---|---|---|---|---|
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (`.claude/rules/moai/workflow/kanban-dispatch.md`) | 93 | `wt`: exit-first `ExitWorktree` → `EnterWorktree(<card-id>)` → `git branch -m WT-<slug>` | Launcher for fresh sessions; standing-session flow pointer | always |
| same | 151–157 | "The `/clear` handoff between phases" | One added sentence: card change = move → `/clear` → re-send; existing lines verbatim | always |
| same | 166–170 | isolation table (`EnterWorktree(<path>)` "re-enter from the current session") | Qualify: followed by `/clear` + re-send for card work | always |
| same | 179 | [HARD] new card starts in a new worktree | Launcher / standing-session flow; three sentences kept verbatim | always |
| same | 181 | [HARD] `WT-` rename | No change | always |
| same | 250–251 | integration `EnterWorktree(<release-worktree-path>)`, return `EnterWorktree(<own-path>)` | Per DP-2 | always |
| `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch-detail.md` (mirror) | 27, 28, 188, 214–220 | glossary "worktree"; example `wt: EnterWorktree(t0)`; factory `/clear` boundary; `/clear` message structure | Launcher example; standing-session per-card flow in full; rationale for the order | paths |
| `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` (mirror) | 147, 218–220, 224 | `EnterWorktree(<path>)` for current-session re-entry | Card-session caveat (REQ-SMM-012 bounded) | paths |
| `internal/template/templates/AGENTS.md.tmpl` / `AGENTS.md` | tmpl 118–121, 136–139 / root 109–112, 127–130 | §3 entry line; "Start a new card in a new worktree" | Launcher + flow, net ≤ 0 tokens | always (root) |
| `.claude/rules/local/gitflow-lane-protocol.md` (local-only) | 20, 22, 37, 83, 155 | `moai cc -w <card-id>` 또는 `EnterWorktree(<card-id>)`; exit-first; return; wait at primary; develop refresh | Launcher only for fresh entry; standing flow; DP-2 | local |
| `CLAUDE.local.md` (local-only) | §4.1 lane-duty bullets (340) | window steps with `EnterWorktree(.claude/worktrees/develop)` | Entry-form sentence + standing flow + DP-2 | local |

Baselines at `2370c5b31`:
- `EnterWorktree(<card-id>)`: 2 in the template kanban dispatch rule, 2 in the lane protocol, 0 elsewhere in the set.
- `wt: EnterWorktree(t0)`: 1 in the template detail companion.
- `[HARD]` counts: kanban dispatch 38, detail 8, worktree integration 18, lane protocol 21, CLAUDE.local.md 48, `AGENTS.md` 0.
- Always-loaded surface: 77,530 tokens against a budget of 77,600 (headroom 70), measured by `go test ./internal/config/ -run '^TestAlwaysLoadedTokenBudget$' -count=1 -v`.

## §E — Exclusions (What NOT to Build)

### Out of Scope — Option A, L2 relocation of card worktrees

- Held by the lead (verdict §④). Blockers were read from code and tool contracts: `internal/cli/worktree/new.go:50-58`, `internal/hook/worktree_create.go:41`, and the `EnterWorktree` switch contract.
- It can be revived only when both verdict §④ conditions hold: an authorized L2 creation path (the WorktreeCreate hook included), and an authorized move path for a session already inside a worktree.

### Out of Scope — Option B, slimming the primary CLAUDE.local.md

- Owned by t1243/t1259. Cross-reference only.

### Out of Scope — Instruction-file double load

- The upward-traversal load of the primary CLAUDE.local.md is not fixed by the entry form (verdict §① Claim 2), and is neither measured nor claimed here.

### Out of Scope — Claude Code runtime behavior

- No change to skill or instruction discovery, to `EnterWorktree` / `ExitWorktree`, to `/clear`, or to the `moai cc` launcher.

### Out of Scope — Generic resume and non-card sessions

- `session-handoff.md` Block 0 and `session-handoff-examples.md` keep `EnterWorktree(<path>)` for generic resumes.

### Out of Scope — Subagent inheritance and token thresholds

- Subagent instruction inheritance is not measured.
- The REQ-SMM-008 cost figures are recorded, but no threshold is set on them.

## §F — Decision Points for the Orchestrator

Each decision point is resolved at the Implementation Kickoff Approval gate and recorded in progress.md §E.2 before M1. For each, the facts come first and the recommendation follows as a separate label.

### DP-1 — Hook: block, warn, or docs-only

| | Block (PreToolUse deny) | Warn (PostToolUse notice) | Docs-only |
|---|---|---|---|
| Effect | Refuses the move | Move happens; the model is told the next step | No runtime effect |
| Exemption determination | Needed: integration/release worktree, resume re-entry, non-card sessions | Needed: the same set (a notice on an exempt move is wrong advice) | None |
| Template audience | Distributed to non-Kanban users (`claude -w`, `isolation: worktree` subagents); must be gated to Kanban/Factory mode | Same gating needed; otherwise non-Kanban users are told the lead will re-send a pointer | Not applicable |
| Output channel | Deny reason (the documented PreToolUse deny path returns it to the model; not probed here) | `additionalContext`. The existing handler (`internal/hook/post_tool_worktree.go`) emits only `systemMessage`, which the Go type documents as "shown to user" (`internal/hook/types.go:366`). Whether `systemMessage` reaches the model is **unverified**; a probe must show the notice text in the model's reply | None |
| Build cost | New PreToolUse matcher in the settings template, handler, config key + default, tests; Tier stays L | Extend one handler + test; mode gate + exemption logic | None |
| Failure mode | A wrong exemption blocks a sanctioned move; the model may route around the deny | A wrong exemption gives wrong advice; the notice can be ignored | A lane is reminded only by the dispatch text |

**Recommendation (label):** docs-only in this card. A follow-up card can add warn once a probe shows the `additionalContext` notice reaching the model. Both hook options need the same exemption determination and mode gate; warn's harm is lower on a wrong determination.

### DP-2 — Integration-window moves: exempt or covered

- **Exempt:** the move is short, does no card work in the tree it enters, and ends with a return. **Cost:** each entry into the integration worktree adds that tree's skill listing, which stays in the session until the next `/clear` (the same kind of duplicate this SPEC targets).
- **Covered:** the integration steps are re-sent after each `/clear`. **Cost:** two extra `/clear` rounds per integration window, and the lane loses its merge context mid-window.

**Recommendation (label):** exempt. The listing it leaves lasts at most until the between-cards `/clear` of REQ-SMM-010; whether that `/clear` removes it is what P3 measures.

### DP-3 — Ordering against the rules-diet card t1175

- `WT-rules-diet` is not an ancestor of this branch. Its tip was `3a48485af` at plan time and `4989ea6b0` at revision time, so the tip moves.
- Its diff rewrites the kanban dispatch rule and `AGENTS.md`, and adds `kanban-dispatch-mechanics.md`.

**Recommendation (label):** wait. M1 runs now. The doctrine milestones start after t1175 lands on develop and develop is absorbed here. Ranges then start at the read-time `CARD_BASE`, so t1175's commits fall outside them.
