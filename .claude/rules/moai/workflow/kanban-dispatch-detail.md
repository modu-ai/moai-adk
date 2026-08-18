---
description: "Detail companion for kanban-dispatch.md — terminology, board table, card classes, dispatch-cycle naming, sync-gate review-lens table, /clear message structure, isolation rationale, verification-load incident record"
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-kanban.md,**/.claude/skills/moai/workflows/todo.md"
---

# Kanban Dispatch — Detail Companion

> Detail companion of `kanban-dispatch.md` (the always-loaded stub). The stub keeps every [HARD] rule, prohibition, and cross-reference; this file owns the long tables, the dispatch-cycle walkthrough, incident narratives, and worked rationale. Load when moving a card between columns, classifying a card, or choosing review lenses for a sync dispatch.

## Terminology — the board vocabulary

`kanban-dispatch.md`, `sprint-round-naming.md`, and the operating notes share a working vocabulary that previously had no definition anywhere. Each term gets one definition and one example; the sections below assume these meanings.

| Term | Definition | Example |
|---|---|---|
| **lane** | One parallel work stream that carries a card end to end: one session paired with one worktree. A lane is a swimlane — a band reserved for one stream of work so parallel streams never interleave, and never share a working tree. "Lane-local verification" = that lane runs only the tests its own change can affect. | The `run` session working in worktree `WT-t0` is one lane. |
| **card** | One unit of work on the board, entered by the operator via `/moai todo "<description>"` and referred to by a short id. A card owns one worktree, one progress record, and its completion evidence. | `t0` — a one-line fix card. |
| **column** | One stage of the board, in fixed order `backlog → plan → run → sync → done`. The three working columns each map to exactly one companion role; the review verdict lives inside the sync gate. | `/moai run <SPEC-ID>` happens in the `run` column. |
| **backlog** | The entry queue of the board. No session owns it by design — work enters only when the operator puts it there. | `/moai todo "rename hint is stale"` appends a card to the backlog. |
| **lead** | The single coordinating session (`moai cc -k`). Moves cards between columns on evidence it read itself, asks the operator to `/clear` companions between phases, never writes code. | The session that dispatched a card with its worktree instruction. |
| **companion** | A worker session launched by hand, one terminal at a time (`moai cc -k --name <role>`), owning one column's work at a time. Named by its bare role; a second live session claiming the same role takes the next free number. | `plan`, `run`, `sync`. |
| **run-id** | The short identifier the lead prints at launch. It names the LEAD alone (its session name, leader socket, `MOAI_KANBAN_ID`) — companions are named by role and never carry it. | `a1b2c3` — the lead's own session name. |
| **worktree** | The isolated checkout where a card's work happens, entered through the launcher (`moai cc -w <name>` / `EnterWorktree`), never raw `git worktree add`. Branch named `WT-<card-id>`. A worktree outlives a phase: one spans run through sync. | `.claude/worktrees/t0` on branch `WT-t0`. |
| **dispatch** | The lead's instruction to one companion: a pointer (card id, SPEC id, phase command, completion signal), never a copy of the work. Written in the operator's conversation_language. | "card: t0 — wt: EnterWorktree(t0) … evidence: .moai/reports/t0/". |

The pair most easily confused: a **column** names a phase of the work (`run`); a **lane** names who carries one card through those phases (the `run` session in `WT-t0`). One is a stage on the board, the other is a stream through the stages.

## The board

Five columns, fixed and ordered:

```
backlog → plan → run → sync → done
```

`backlog` and `done` have no owning session. The three working columns between them each map to exactly one companion role, which is what makes dispatch a lookup rather than a decision. There is no `review` column: the review verdict is absorbed by the sync gate, which runs the review lenses itself (§ Review lens selection).

| Column | Owning role | What happens there |
|---|---|---|
| `backlog` | *none* — a queue | Work waits. Entry is an operator act (see the stub). |
| `plan` | `plan` | SPEC authored (`/moai plan`), then plan-audited. |
| `run` | `run` | Implementation (`/moai run <SPEC-ID>`). |
| `sync` | `sync` | Review verdict (lenses per card), docs, CHANGELOG, PR (`/moai sync <SPEC-ID>`). |
| `done` | *none* — terminal | Card closed. Nothing is dispatched here. |

## Report milestones ↔ queue cards

The stub's [HARD] rule — a `## Card Cross-Check` section per milestone-bearing report, mapping claims verified against the queue — exists because report→card linkage used to live in one person's memory: a report declared milestones, cards were issued separately, and nothing reconciled the two. Milestones surfaced with no card, and cards claimed milestones that had already landed.

The mechanical check runs the same comparison the lead states by hand:

```
moai graph build && moai graph query --milestones-no-card
```

It writes the report-milestone and milestone-card edges from each report's Card Cross-Check table and lists every milestone whose claimed card is missing from the live queue (queued/picked; dropped does not qualify). "Not in live queue" covers both completed and never-issued cards — resolve each flag with `git log --oneline --grep 'merge: <card-id>'` before issuing a new card.

## Card classes — not every card needs every column

Most of what accumulates in the backlog is chores: a one-line fix, a stale reference, a renamed flag. Sending those through `plan → run → sync` costs more in ceremony than the change is worth. The lead classifies each card as it leaves `backlog` and names the class in the dispatch.

| Class | Shape | Path |
|---|---|---|
| A — direct close | The change is one file and one line, there is no design judgement in it, and CI catches the regression | One session carries the card through to a pull request; `plan` is skipped |
| B — defect, cause unknown | Something is wrong and the cause has not been established | `run → sync`; `plan` is skipped, so no SPEC exists |
| C — design change | The change contains a decision, or spans subsystems | All three working columns |

The Class-A evidence rule is the same shape as the CodeRabbit section of the stub: a class that skips review on a claim nobody checked is exactly the unobserved-claim hazard this rule forbids everywhere else; writing the justification down is not the same as verifying it.

For Class A this inverts where the parallelism comes from. Handing three sessions a whole card each puts three cards in flight; pipelining one card through three columns puts one. Pipelining repays its handoff cost only when each column does substantial work, which is the Class C case — and research fan-out during `plan` is reserved for Class C for the same reason.

## The dispatch cycle

Each arrow below is one dispatch from the lead to one companion session:

```
[operator picks a card]  →  plan  →  run  →  sync  →  [lead marks done]
```

Dispatch is addressed by session name. Companions are named by their bare role; a name held by a live session is bumped to the next free number, and no run id travels in companion names (one run per machine; the lead keeps the id). `ListAgents` reports the live set; send with `SendMessage({to: "<name>", message: "…"})`, using the short reference the listing prints when a bare name is ambiguous.

Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command to run, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself.

### Dispatch language — classification rationale

[HARD] A dispatch is written in the operator's `conversation_language` (normative statement lives in the stub). This is a classification, not an exemption from the language rules, and it needs no change to either of them:

- `agent-common-protocol.md` § Language Handling already opens with the opposite of an English default — agents receive and respond in the configured `conversation_language`. What it fixes to English is code, identifiers, and names. A dispatch is prose, so the rule was never against it; it simply did not name cross-session messages in its list.
- `moai-constitution.md` § Response Language reserves English for internal agent communication, but the axis that clause sits on is stated one line above it: user-facing responses go in the operator's language. A message a human reads is user-facing by that rule's own criterion, so putting a dispatch in the operator's language applies the constitution rather than carving an exception out of it.

### Dispatch format — rationale

The address-block format (normative statement lives in the stub) is the "pointer, not a copy" rule above made mechanical: every field is an address the companion resolves by reading what it names. It also settles the Dispatch language rule by construction — a block of pure addresses has nothing to translate, while the lead's reports to the operator (progress notes, `/clear` requests) remain in the operator's `conversation_language`.

## The verdict's home

The stub keeps the norm — the final PASS/FAIL verdict is the lead's, read from evidence on disk, never delegated to the lane that produced the work. The division is structural, not ceremonial: where the board's lanes run on a different backend than the lead, the lane sessions cannot commission judgment work onto the lead's backend, so the verdict has a home in the lead even when the execution has none.

## CodeRabbit endpoint measurement

The CodeRabbit predicate in the stub assumes the **combined** status endpoint, `/commits/{sha}/status`, which returns only the most recent status per context — measured on this repository, exactly one CodeRabbit entry per head. That assumption is the load-bearing part, so it is stated rather than left implicit: do not substitute the plural `/commits/{sha}/statuses`, which returns the full history newest-first. Measured on one head there: five CodeRabbit entries running from `Review queued` through `Review completed`, so a positional pick on that endpoint is wrong in one direction or the other — `last` selects the oldest. Where history is genuinely wanted, select by maximum `created_at` rather than by position.

Branch protection is not the lever here. The status state is `success` in precisely the failing case, so adding CodeRabbit to the required contexts would admit the unreviewed pull request just as readily. The distinction lives in the description, and only a read of the description surfaces it — which is why an automated merge gate closing this hole does not close it on the path a human merges by hand.

## Review lens selection

`review` is not one thing. The lead picks the lenses from what the card actually changed, and states the choice in the `sync` dispatch so the sync gate runs that review rather than re-deriving it:

| Card touched | Lenses to instruct |
|---|---|
| Auth, session handling, input parsing, external calls, secrets, file/path handling | `--security` (add `--deep` when the surface is reachable by untrusted input) |
| Non-trivial logic across several files | `--deep` (adversarially verified multi-phase scan) |
| Whole-tree sweep rather than a diff | add `--repo` |
| Suspected over-engineering | `--lean` (advisory only; applies no fixes) |
| UI or design-system surface | `--design`, and `--critique` after the build |
| Small, local, low-risk diff | no flag — the default 4-perspective pass is enough |

## The `/clear` handoff between phases — message structure

The lead's message to the operator states three things, in this order:

1. **What closed** — the card, the phase, and the evidence that was read.
2. **Which session to `/clear`** — by name, so the operator clears the right terminal.
3. **What happens next** — the column the card moves to, and which session will be instructed once the clear is done.

## Isolation rationale

Two properties make the shared checkout the wrong place for a card:

- Several sessions read it at once, so a branch switch, a `git stash`, or a `git add -A` there sweeps another session's uncommitted work into a commit that was never meant to carry it.
- A card outlives a phase. Its worktree spans run through sync, which is why disposal is triggered by the merge rather than by the phase finishing.

## Verification load incident record

Sessions share one machine as surely as they share one checkout, and verification is where that sharing goes wrong. Measured on a day when it did: load average reached 413, a neighbouring workspace's build took two and a half minutes, and its browser tests timed out — not because anything was wrong with them, but because four lanes were each running the full test suite at once and a full suite there takes five to ten minutes.

The never-spawn-background-load rule (normative statement lives in the stub) comes from the same incident's second cause: a verification recipe started eight spin loops to test behaviour under CPU contention and placed its kill line *after* the long test command. The agent finished before reaching it, and twelve spinners ran orphaned for thirty-seven minutes.

The same day supplied the reason contention and flakiness feed each other: a failing test left an unbounded spin-loop goroutine running, which burned a core for the remainder of that package's run and slowed every test after it. Load makes tests fail; failing tests can generate load.
