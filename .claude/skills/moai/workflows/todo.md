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

State lives at `.moai/state/kanban/backlog.json` (project-local, not committed).
Do not read or write that file directly — run the `moai todo` commands: they
hold a cross-process lock across every mutation, so concurrent sessions cannot
lose cards or collide ids.

## Commands

When the operator says `/moai todo "<description>"`, run
`moai todo add "<description>"`; a bare `/moai todo` runs `moai todo list`.

| Command | Effect |
|---|---|
| `moai todo add "<text>"` | Append an item under the lock. Prints the issued id (`t<n>`) and its queue position. |
| `moai todo list` | Render the queue, lock-free. `--json` emits the structured records. |
| `moai todo done <n>` | Remove the addressed row under the lock. A bare `<n>` means `t<n>`; the explicit id (`moai todo done t3`) is the preferred form because positions move. |
| `moai todo next` | Print the queued items oldest-first — read-only candidates. |
| `moai todo next <n> [--spec <SPEC-ID>]` | Mark the addressed item `picked` (attaching `spec_id` when given) as one locked write. |

The queue is never mutated through any other surface. A missing backlog file is
an empty queue, never an error; a malformed file is reported and left untouched.

## Reading the records

`moai todo list --json` emits the file's records:

```json
{
  "version": 1,
  "last_seq": 12,
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

- `id` — assigned on append, never reused after removal (`last_seq` is the
  persisted high-water mark that guarantees it).
- `spec_id` — filled in when the item is picked; until then it is `null`, which
  is what distinguishes a backlog item from a card already on the board.
- `state` — `queued` | `picked` | `dropped`. A picked item stays in the file so
  the operator can see what is in flight; it is removed when its card reaches
  `done`.

## Picking the next card

`moai todo next` (and the lead's own post-`/clear` opening move) presents the
queued items through `AskUserQuestion` — one option per queued item, capped at
the four the tool allows, oldest first, with the remainder summarized in the
response body so nothing is hidden behind the cap.

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

1. Record it with `moai todo next <n> --spec <SPEC-ID>` (one locked write).
2. Dispatch to the `plan` session per `kanban-dispatch.md` — the card enters the
   `plan` column, and SPEC authoring happens there, not here.

## Outside Kanban Mode

`moai todo` works in an ordinary session too — it is just a queue. What it will
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
