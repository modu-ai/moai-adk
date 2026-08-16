---
description: "Lead-only detail companion for kanban-dispatch.md — board, card classes, dispatch cycle, review lenses, /clear handoff"
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-kanban.md,**/.claude/skills/moai/workflows/todo.md"
---

# Kanban Dispatch — Detail Companion

> Detail companion of `kanban-dispatch.md`, owning its lead-only sections; the always-loaded stub keeps every section that binds non-lead sessions. Load when moving a card between columns, classifying a card, or choosing review lenses.

## The board

Six columns, fixed and ordered:

```
backlog → plan → run → review → sync → done
```

`backlog` and `done` have no owning session. The four columns between them each map to exactly one companion role, which is what makes dispatch a lookup rather than a decision.

| Column | Owning role | What happens there |
|---|---|---|
| `backlog` | *none* — a queue | Work waits. Entry is an operator act (see below). |
| `plan` | `plan` | SPEC authored (`/moai plan`), then plan-audited. |
| `run` | `run` | Implementation (`/moai run <SPEC-ID>`). |
| `review` | `review` | Code review (`/moai review …`) with lenses chosen per card. |
| `sync` | `sync` | Docs, CHANGELOG, PR (`/moai sync <SPEC-ID>`). |
| `done` | *none* — terminal | Card closed. Nothing is dispatched here. |

## Entry into the board is an operator act

`backlog` has no owning session, so no completion report exists for the lead to react to, and a lead that admitted cards on its own initiative would be **generating** work rather than scheduling it. Entry is therefore always the operator's decision.

The mechanism is `/moai todo`:

- `/moai todo "<description>"` appends an item to the backlog queue.
- `/moai todo` alone lists the queue.
- After a `/clear`, the lead presents the queue through `AskUserQuestion` and the operator picks the next card. Only then does the lead dispatch to `plan`.

The lead never picks for the operator, and never silently promotes a backlog item.

## Card classes — not every card needs four columns

Most of what accumulates in the backlog is chores: a one-line fix, a stale reference, a renamed flag. Sending those through `plan → run → review → sync` costs more in ceremony than the change is worth. The lead classifies each card as it leaves `backlog` and names the class in the dispatch.

| Class | Shape | Path |
|---|---|---|
| A — direct close | The change is one file and one line, there is no design judgement in it, and CI catches the regression | One session carries the card through to a pull request; `plan` and `review` are skipped |
| B — defect, cause unknown | Something is wrong and the cause has not been established | `run → review → sync`; `plan` is skipped, so no SPEC exists |
| C — design change | The change contains a decision, or spans subsystems | All four columns |

[HARD] **Class A is admitted on checked evidence, not on an assertion.** Two of its three properties are mechanically checkable, so they are checked and the check is cited: the diff is measured (`git diff --stat` against the base, showing the one file) and CI is observed green **on the head that will merge**. The third — no design judgement in it — is a judgement rather than a measurement, so it is stated in the dispatch where the operator can disagree with it. A card that cannot cite both measurements is not Class A, and it takes the `review` column.

This is the same shape as the CodeRabbit section below. A class that skips review on a claim nobody checked is exactly the unobserved-claim hazard this rule forbids everywhere else; writing the justification down is not the same as verifying it.

The justification is never "it is faster". Speed is the effect of skipping the columns, not the reason for it, and a card justified by speed alone is a Class C card being rushed.

**Work in progress: one card per column, and only when each card occupies a different worktree.** Two cards sharing one worktree run serially whatever columns they sit in, because they share a working tree and a branch.

For Class A this inverts where the parallelism comes from. Handing four sessions a whole card each puts four cards in flight; pipelining one card through four columns puts one. Pipelining repays its handoff cost only when each column does substantial work, which is the Class C case — and research fan-out during `plan` is reserved for Class C for the same reason.

## The dispatch cycle

Each arrow below is one dispatch from the lead to one companion session:

```
[operator picks a card]  →  plan  →  run  →  review  →  sync  →  [lead marks done]
```

Dispatch is addressed by session name. Companion sessions are named `<role>-<run-id>` at launch (for example `plan-a1b2c3`), and `ListAgents` reports the live set. Send with `SendMessage({to: "<name>", message: "…"})`; when a bare name is ambiguous, use the short reference the listing prints.

Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command to run, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself.

### Dispatch language

[HARD] A dispatch is written in the operator's `conversation_language`. The operator watches it scroll past, which makes it user-facing output rather than internal agent traffic.

This is a classification, not an exemption from the language rules, and it needs no change to either of them:

- `agent-common-protocol.md` § Language Handling already opens with the opposite of an English default — agents receive and respond in the configured `conversation_language`. What it fixes to English is code, identifiers, and names. A dispatch is prose, so the rule was never against it; it simply did not name cross-session messages in its list.
- `moai-constitution.md` § Response Language reserves English for internal agent communication, but the axis that clause sits on is stated one line above it: user-facing responses go in the operator's language. A message a human reads is user-facing by that rule's own criterion, so putting a dispatch in the operator's language applies the constitution rather than carving an exception out of it.

The carve-out is narrow, and the boundary is **who reads it**. An `Agent()` subagent prompt reaches no human and stays English.

What stays verbatim in every language: SPEC IDs, command names and their flags, file paths, session names, and technical identifiers. Those are addresses rather than prose, and a translated address does not resolve.

## Review lens selection

`review` is not one thing. The lead picks the lenses from what the card actually changed, and states the reason in the dispatch so the review session does not re-derive it:

| Card touched | Lenses to instruct |
|---|---|
| Auth, session handling, input parsing, external calls, secrets, file/path handling | `--security` (add `--deep` when the surface is reachable by untrusted input) |
| Non-trivial logic across several files | `--deep` (adversarially verified multi-phase scan) |
| Whole-tree sweep rather than a diff | add `--repo` |
| Suspected over-engineering | `--lean` (advisory only; applies no fixes) |
| UI or design-system surface | `--design`, and `--critique` after the build |
| Small, local, low-risk diff | no flag — the default 4-perspective pass is enough |

`--deep --patch` is opt-in twice over: `--patch` drafts a fix and is absent unless the operator asked for it. Do not add it on the lead's own initiative.

## The `/clear` handoff between phases

[HARD] A companion session does not carry one card's context into the next card. When a phase completes and the lead has read its evidence, the lead **asks the operator to `/clear` that session** — `/clear` is a user-typed command and cannot be sent as an instruction.

The lead's message to the operator states three things, in this order:

1. **What closed** — the card, the phase, and the evidence that was read.
2. **Which session to `/clear`** — by name, so the operator clears the right terminal.
3. **What happens next** — the column the card moves to, and which session will be instructed once the clear is done.

Where the next phase reuses a session that has just been cleared, the lead re-sends the full pointer instruction after the clear rather than assuming the session remembers the card.

The lead's own session is cleared the same way, between cards rather than between phases: once a card reaches `done`, the lead asks the operator to `/clear` the lead session, and the next turn begins by presenting the backlog queue again.

