# /moai todo — Backlog Queue

> The operator's entry point into the kanban board. `backlog` has no owning
> session, so nothing dispatches work into it — admission is always an operator
> act, and this is the surface for it.
> Dispatch protocol: `.claude/rules/moai/workflow/kanban-dispatch.md`.

## What It Is

A plain queue of things to work on next. An item is one line of intent — not a
SPEC, not a plan, not an estimate. It becomes a SPEC only when the operator picks
it and the lead dispatches it to the `plan` session.

The queue is deliberately thin. It records *what the operator wants next*, and
nothing that a SPEC, a git history, or a board would record better.

State lives at `.moai/state/kanban/backlog.json`. It is project-local and is not
committed.

## Verbs

| Invocation | Effect |
|---|---|
| `/moai todo "<description>"` | Append an item to the queue. Prints the item and its position. |
| `/moai todo` | List the queue, in order, with positions. |
| `/moai todo done <n>` | Remove item `n` (it was completed elsewhere, or is no longer wanted). Prints what was removed. |
| `/moai todo next` | Present the queue through `AskUserQuestion` so the operator picks the next card. |

Any other argument form is treated as a description — `/moai todo fix the flaky
CI cache` adds an item rather than erroring, because the cost of a wrong guess
here is one line the operator can delete.

## Record shape

```json
{
  "version": 1,
  "items": [
    {
      "id": "t1",
      "text": "Rework the auth middleware error paths",
      "added_at": "<RFC3339 timestamp>",
      "spec_id": null,
      "state": "queued"
    }
  ]
}
```

- `id` — short, stable, assigned on append. Never reused after removal.
- `spec_id` — filled in when the item is picked and a SPEC is authored for it.
  Until then it is `null`, which is what distinguishes a backlog item from a card
  already on the board.
- `state` — `queued` | `picked` | `dropped`. A picked item stays in the file so
  the operator can see what is in flight; it is removed when its card reaches
  `done`.

Write the file atomically (write a sibling temp file, then rename) so a crash
mid-write cannot truncate the queue. A missing file is an empty queue, not an
error. A malformed file is reported and left untouched — never silently reset,
because the operator's intent is the one thing here that cannot be regenerated.

## Picking the next card

`/moai todo next` (and the lead's own post-`/clear` opening move) presents the
queue through `AskUserQuestion` — one option per queued item, capped at the four
the tool allows, oldest first, with the remainder summarized in the response body
so nothing is hidden behind the cap.

[HARD] The pick is the operator's. Do not preselect, do not reorder by inferred
priority, and do not append a "start the top one" default. Where the queue is
empty, say so and stop — an empty backlog is a legitimate state, not a prompt to
invent work.

An operator may authorize several cards at once — naming them, or saying to work
the queue in order until it empties. That is still their pick, made once instead
of one at a time, and the lead then admits those cards in the authorized order
without asking again. It grants nothing else: no additions to the queue, no
reordering, and no cover for a card that turns out to need a decision the
authorization never covered. See `kanban-dispatch.md` § Batch admission.

Once picked:

1. Mark the item `picked`.
2. Dispatch to the `plan` session per `kanban-dispatch.md` — the card enters the
   `plan` column, and SPEC authoring happens there, not here.
3. Record the SPEC ID onto the item as soon as the plan session reports one.

## Outside Kanban Mode

`/moai todo` works in an ordinary session too — it is just a queue. What it will
not do is dispatch: with no companion sessions there is nobody to instruct, so
the queue is read and written and the operator drives the work themselves.

Say this plainly when it applies rather than implying a board exists.

## Boundaries

- **Not a task tracker.** No priorities, no assignees, no due dates, no
  dependencies. Anything needing those belongs in an issue tracker or a SPEC.
- **Not a board.** Column position lives with the lead and the SPEC status, not
  in this file.
- **Not a source of truth for work in flight.** Once a card has a SPEC, the SPEC
  artifacts are authoritative; the backlog item is only a pointer to it.
- **Never auto-populated.** The queue is not filled from TODO comments, open
  issues, or audit findings on the tool's initiative. An operator adds items.

## Cross-references

- `.claude/rules/moai/workflow/kanban-dispatch.md` — the dispatch cycle this feeds
- `.claude/rules/moai/core/askuser-protocol.md` — the channel the pick runs through
- `.claude/agents/moai/manager-kanban.md` — the coordination agent
