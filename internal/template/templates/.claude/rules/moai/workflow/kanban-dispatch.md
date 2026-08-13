# Kanban Dispatch Protocol

How the **lead** session of Kanban Mode moves a card across the board: what admits work, who is told to do it, how completion is judged, and when the operator is asked to `/clear`.

> **Loading scope**: Intentionally always-loaded. A session learns it is the kanban lead from the SessionStart context, not from a file path, so a `paths:`-restricted rule would never reach the session that needs it.

## Scope — when this rule is live

This rule binds a session whose SessionStart context declares **Kanban Mode** with the `lead` role. In every other session it is inert: a companion session (`plan` / `run` / `review` / `sync`) receives instructions, it does not dispatch them, and a session outside Kanban Mode has no board to move.

Kanban Mode is entered with `moai cc -k` (or `moai glm -k`), which elects one lead and prints one launch command per companion role. Companion sessions are launched **by hand, one per terminal** — a session cannot launch another session, and no mechanism to spawn a peer exists or is wanted.

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

## The dispatch cycle

Each arrow below is one dispatch from the lead to one companion session:

```
[operator picks a card]  →  plan  →  run  →  review  →  sync  →  [lead marks done]
```

Dispatch is addressed by session name. Companion sessions are named `<role>-<run-id>` at launch (for example `plan-a1b2c3`), and `ListAgents` reports the live set. Send with `SendMessage({to: "<name>", message: "…"})`; when a bare name is ambiguous, use the short reference the listing prints.

Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command to run, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself rather than receiving them inline.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the lead reads the sync session's completion evidence and records the terminal transition itself.

## Completion is read, never trusted

[HARD] The lead advances a card on **evidence it read**, not on a companion's reply. Reply routing between sessions is not guaranteed to arrive, and a reply is a claim rather than an observation.

Before moving a card out of a working column, the lead reads the card's `progress.md` (and, where the phase declares one, the verification evidence path it cites) and confirms the phase actually closed. A missing, unreadable, or stale evidence file is a **gap** — the card stays where it is and the lead reports why. Absence of a failure signal is not a pass.

This applies equally to the operator: when the lead reports a column advanced, it names what it read.

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

## Boundaries — what this protocol does not do

- **No board state store.** The queue is a plain file; column position is held by the lead within a card's run and re-derived from SPEC status after a clear. Persistent six-column state, per-card worktree lifecycle, WIP limits, and card/frontmatter consistency reconciliation are separate work and are not assumed here.
- **No spawning.** The lead addresses sessions the operator launched. It never creates one.
- **No gate bypass.** Kickoff approval before run-phase entry, and every other approval gate, is unchanged by being inside a dispatch cycle.
- **No question delegation.** Companion sessions return blocker reports; the operator is asked by the lead, through `AskUserQuestion`.
- **A role with no live session is a fault, not a wait.** Where the role owning a dispatchable card's column has no occupant, the lead dispatches to nobody, and reports both the empty role and the waiting card. Silently holding the card presents as a hang, which is the most expensive failure shape to diagnose.

## Cross-references

- `.claude/rules/moai/core/askuser-protocol.md` — the question channel the lead uses for card selection and `/clear` prompts
- `.claude/rules/moai/core/verification-claim-integrity.md` — why completion is read rather than trusted
- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/skills/moai/workflows/todo.md` — the backlog queue surface
- `.claude/agents/moai/manager-kanban.md` — the coordination agent, including its kanban-lead role

---

Classification: Evolvable operational rule — applies to the lead session of Kanban Mode.
