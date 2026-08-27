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

State lives at `.moai/state/todo/backlog.json` of the PRIMARY checkout
(project-local, not committed). A linked worktree resolves to the same
primary queue — one repository, one queue: a card worktree's `moai todo`
adds to and reads the file the lead and the foreman loop see. A project
without git metadata keeps its queue at `~/.moai/todo/<project-key>/backlog.json`
instead — the first run there adopts an existing project-local queue (same
items, same states) rather than starting an empty one.
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
| `moai todo edit <n> "<text>" [--expect <prefix>]` | Rewrite the addressed card's text under the lock. `id`, `added_at`, `state`, and `spec_id` are preserved, so a correction never churns the card's identity the way `done` + re-add does. The confirmation carries the prior text as well as the new one, so a wrong edit is reversed by editing back. |
| `moai todo move <n> (--top\|--bottom\|--before <m>\|--after <m>)` | Reposition the card within the queue order under the lock. Exactly one destination is required. The move permutes the order and nothing else — no card is dropped, duplicated, or altered — so a wrong move is reversed by another move. |

| `moai todo drop <n> "<reason>" [--expect <prefix>]` | Move the addressed **queued** card to `dropped` under the lock, prefixing its text with `[DROPPED — <reason>] `. The card stays in the file — `done` removes a finished card, `drop` keeps a discarded one visible with its reason — and it is no longer a pick candidate. A picked card is unpicked first, so nothing `undrop` cannot restore is ever taken. |
| `moai todo undrop <n> [--expect <prefix>]` | Return the addressed dropped card to `queued`, stripping the marker. The state is the authority, so a card marked dropped by hand (no marker in its text) undrops with its text untouched. `drop` + `undrop` returns the queue file to the same bytes. |
| `moai todo add "<text>" --force` | Admit a card the analyser reads as an exact duplicate. The card is appended verbatim and the queue records that the duplicate was forced, so the collision stays visible instead of being argued about later. |
| `moai todo analyze` | Re-read the whole queue and record what the analyser finds. Appends, removes, reorders, and edits nothing. Re-running records nothing new — the same relation is never stacked twice. |
| `moai todo relate <a> <b> --relation (contains\|absorbs\|replaces\|conflicts) [--note <text>]` | Record one relation between two existing cards. The verb writes a record and touches neither card; `absorbs` does not absorb. |
| `moai todo unrelate <index>` | Remove the addressed record. The index is the one `why` prints. No card changes. |
| `moai todo why <n>` | Print every record naming the card, or an explicit no-findings line. A card the queue knows nothing about says so rather than printing nothing. |
| `moai todo pr [<id>]` | Report each card's open pull request or landed state — read-only, and it writes nothing. Four outcomes: `linked` (one open PR carries the card id; confidence `exact` from the PR title, `inferred` from a single PR body), `ambiguous` (several PR bodies carry it — every candidate is listed and none is chosen), `landed` (no open PR, but the integration branch's history names the card), and `no-link`. One `gh` query per invocation, never one per card, which is why the link is a separate verb rather than a column on `list`: the queue's cheapest read stays free of the network. When `gh` is absent, unauthenticated, or offline the link column renders empty, the degradation is noted on stderr, and the exit code stays 0 — the landed check is local git and keeps running. |

[HARD] `edit`, `move`, `drop`, and `undrop` are operator acts, exactly like
`add` and the pick. Correct a card's wording, move it, or discard it because
the operator said to — never on inferred priority, never as tidy-up, never to
fold one card into another, and never because a card looks stale. The queue
records the operator's intent; it does not curate it.

The queue is never mutated through any other surface. A missing backlog file is
an empty queue, never an error; a malformed file is reported and left untouched.

### What the analyser may do

Analysis runs automatically and records — on every `add`, and across the whole
queue on `analyze`. A record changes no card: not its text, not its position,
not its state.

The analysis performs exactly one transformation: it **refuses the admission**
of a card whose normalized text is identical to a card already queued or
picked. A refusal touches no existing card and leaves the queue file
byte-identical — it creates nothing rather than folding anything, and the
operator sees an error instead of an id. `--force` admits the card anyway and
records that it was forced.

[HARD] Analysis never folds one card into another, never reorders the queue,
never drops a card, and never edits one. The four semantic relations —
`contains`, `absorbs`, `replaces`, `conflicts` — cause nothing but a record.
Acting on a record is the operator's act, performed through `drop`, `edit`, or
`move`, exactly as the clause above requires.

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
  ],
  "findings": [
    {
      "subject_id": "t2",
      "related_id": "t1",
      "relation": "near-duplicate",
      "source": "mechanical",
      "score": 0.83,
      "note": "",
      "at": "<RFC3339 timestamp>"
    }
  ]
}
```

- `id` — assigned on append, never reused after removal (`last_seq` is the
  persisted high-water mark that guarantees it).
- `spec_id` — filled in when the item is picked; until then it is `null`, which
  is what distinguishes a backlog item from a card already on the board.
- `findings` — the records the analysis layer keeps ABOUT pairs of cards; a
  relation belongs to the pair rather than to either card, which is why it
  lives here and not in an item. Always an array: a file written before the
  field loads with an empty one, so "no findings" never has to be told apart
  from "no such feature". `source` is `mechanical` (measured text similarity)
  or `agent` (a judgement written through `relate`), and a mechanical finding
  with no agent finding on the same pair renders marked `machine-only` — which
  records that nothing agent-sourced was written, never that anyone reviewed
  it. A finding leaves the file when its card does.
- `state` — `queued` | `picked` | `dropped`. A picked item stays in the file so
  the operator can see what is in flight; it is removed when its card reaches
  `done`. A dropped item stays too, carrying its `[DROPPED — <reason>] ` marker
  and recoverable with `undrop` — a discard is a decision on the record, not an
  erasure.

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
authorization never covered. See `kanban-dispatch.md` § Entry into the board is
an operator act.

A workflow that ends by asking whether to start the card it just issued is the
same thing in a narrower form: the branch the operator chooses IS the pick,
made at the moment the card appears instead of at the next `moai todo next`.
What makes it a pick rather than a preselect is that the question is genuinely
open — starting is one branch among the others, chosen by the operator, and no
branch is taken on their behalf when they do not answer. A workflow that starts
work without that answer has preselected, whatever it calls the step.

Once picked:

1. Record it with `moai todo next <n> --spec <SPEC-ID>` (one locked write).
2. Dispatch to the `plan` session per `kanban-dispatch.md` — the card enters
   the `plan` column, and SPEC authoring happens there, not here.

## Standing sources

A standing source is a workflow the operator authorized once to issue a card
when it finishes, rather than being asked for that card every time. The
authorization is still the operator's and still comes first; what the standing
source changes is *when* they give it, not *who* gives it. A card issued this
way is the workflow carrying out an instruction already on the record — never a
tool deciding on its own that work exists.

`/moai project` is the only standing source. Five properties are what separate
it from invention, and all five bind:

- **One card per run.** Not one per document, per feature, or per finding.
- **Derived, never invented.** The text comes from that run's own
  `.moai/project/harness-spec.yaml` — its `goal`, bounded by its `scope` — so
  the card restates what the operator said in the interview. A run with
  nothing to derive from issues nothing; an empty result is reported, not
  filled in.
- **Marked at the front.** The text carries the `[PROJECT] ` prefix, so the
  queue shows at a glance which cards a workflow issued and which a person
  typed. The prefix is the card's provenance — the record carries no other.
- **The issued id is reported.** The completion report names it (`t<n>`), so
  the card is visible in the same breath as its creation.
- **Starting it is a separate pick.** The card is queued, not started. Whether
  work begins is asked in the same completion question, and that answer is the
  pick (§ Picking the next card).

Re-running the workflow does not stack duplicates: before adding, read the
queue (`moai todo list --json`) and skip the add when a queued card already
carries the same `[PROJECT] ` text, reporting the existing id instead of
issuing a second one.

Nothing else is a standing source. TODO comments, open issues, audit findings,
and report milestones stay outside: they are surfaced to the operator, who asks
for a card when they want one.

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
- **Never auto-populated on the tool's initiative.** The queue is not filled
  from TODO comments, open issues, or audit findings because a tool noticed
  them. An operator adds items — directly, or through a standing source they
  authorized in advance (§ Standing sources). Nothing else adds.

## Cross-references

- `.claude/rules/moai/workflow/kanban-dispatch.md` — the dispatch cycle this feeds
- `.claude/rules/moai/core/askuser-protocol.md` — the channel the pick runs through
- `.claude/agents/moai/manager-lead.md` — the coordination agent
