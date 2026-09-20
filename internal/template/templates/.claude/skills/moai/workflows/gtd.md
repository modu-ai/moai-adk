# /moai gtd — Canonical GTD Entry Point and Backlog Queue

> The canonical workflow identity for GTD task management, and the operator's
> entry point into the kanban board. `backlog` has no owning session, so
> nothing dispatches work into it — admission is always an operator act, and
> this is the surface for it.
> Dispatch protocol: `.claude/rules/moai/workflow/kanban-dispatch.md`.

> Compatibility surface: `/moai todo` is the compat alias of `/moai gtd`, and
> `moai todo` is the compat alias of `moai gtd` on the CLI. Both names dispatch
> through this same entry point and reach one database, with the same card
> identities, ordering, archive, and restore path, so every contract below
> holds through either of them.
## What It Is

A plain queue of things to work on next. An item is one line of intent — not a
SPEC, not a plan, not an estimate. After the operator picks it, the lead follows
the card class: Class A direct close, Class B run → sync, Class C plan → run → sync.
Only Class C requires SPEC authoring.

The queue is deliberately thin. It records *what the operator wants next*, and
nothing that a SPEC, a git history, or a board would record better.

State lives at `~/.moai/db/<project-key>/todo/backlog.db`, keyed from the PRIMARY checkout
(home-scoped, not committed) — a SQLite database, not a JSON file. A
linked worktree resolves to the same primary queue — one repository, one
queue: a card worktree's `moai gtd` adds to and reads the store the lead
and the foreman loop see. A project without git metadata keeps its queue at
the same project-keyed home path — the first run there adopts
an existing project-local queue (same items, same states) rather than
starting an empty one.
A `backlog.json` at that same path is NOT the queue. It is an export
(`moai gtd export-json`) or a pre-migration leftover, its contents can be
arbitrarily stale, and reading it answers silently and wrongly — which is
why every `moai gtd` read verb discloses one on stderr when it finds one.
Do not read or write the store directly — run the `moai gtd` commands: they
hold a cross-process lock across every mutation, so concurrent sessions cannot
lose cards or collide ids.

## Commands

When the operator says `/moai gtd "<description>"`, run
`moai gtd add "<description>"`; a bare `/moai gtd` runs `moai gtd list`.

| Command | Effect |
|---|---|
| `moai gtd add "<text>"` | Append an item under the lock. Prints the issued id (`t<n>`) and its queue position. |
| `moai gtd list` | Render the queue, lock-free. The default view is the live load: `queued` and `picked` cards, with the dropped set collapsed into one count line naming `--dropped`. `moai gtd list --dropped` renders the discarded set with its markers — the surface `undrop` reads. The render is bounded at 20 rows (`--limit <n>` adjusts, `0` lifts the bound); a truncated listing states the withheld count on stderr, because a truncated read must never be mistaken for a complete one. `--json` emits the structured records — every card, dropped included, and never bounded — so a machine consumer filters by the `state` field rather than by absence. |
| `moai gtd done <n> [--expect <prefix>] [--require-landed]` | Take the addressed row out of the live queue under the lock. A bare `<n>` means `t<n>`; the explicit id (`moai gtd done t3`) is the preferred form because positions move. The card and every finding naming it are ARCHIVED rather than discarded, so `undone` restores both; archived rows are invisible to `list`, `next`, `why`, `analyze`, and the counts. `--expect <prefix>` refuses unless the card's text starts with the prefix — the guard against closing the wrong card. `--require-landed` refuses unless a commit on the landed ref names the card; see the note below for what it can and cannot answer. Every successful `done` prints exactly one landing verdict on stdout — `done <id> landing=landed|not-landed|unknown`. Absent the flag the verdict is `unknown` because no query ran, UNLESS the card already carries landing evidence recorded by `moai gtd landed`: a validated record naming a delivering commit IS an answer, obtained earlier and stored, so the verdict reads `landed` and the line appends `sha=<delivering commit> source=operator`. The provenance travels with the value, so a stored assertion is never read as an answer this run produced. A record holding only the observed ref position appends nothing — it asserts no delivering commit. |
| `moai gtd undone <n>` | Restore an archived card to the live queue at the position it held, together with every finding that named it, and empty the archive entry. `done` + `undone` returns the queue record to the same bytes. Refused when the id has since been reissued to a different live card — the collision is named and the live card is left alone. |
| `moai gtd next` | Print the queued items oldest-first — read-only candidates. |
| `moai gtd next <n> [--spec <SPEC-ID>]` | Mark the addressed item `picked` (attaching `spec_id` when given) as one locked write. |
| `moai gtd edit <n> "<text>" [--expect <prefix>]` | Rewrite the addressed card's text under the lock. `id`, `added_at`, `state`, and `spec_id` are preserved, so a correction never churns the card's identity the way `done` + re-add does. The confirmation carries the prior text as well as the new one, so a wrong edit is reversed by editing back. |
| `moai gtd move <n> (--top\|--bottom\|--before <m>\|--after <m>)` | Reposition the card within the queue order under the lock. Exactly one destination is required. The move permutes the order and nothing else — no card is dropped, duplicated, or altered — so a wrong move is reversed by another move. |

| `moai gtd drop <n> "<reason>" [--expect <prefix>]` | Move the addressed **queued** card to `dropped` under the lock, prefixing its text with `[DROPPED — <reason>] `. The card stays in the file — `done` removes a finished card, `drop` keeps a discarded one with its reason (`list --dropped` renders the discarded set; the default list hides it behind a count line) — and it is no longer a pick candidate. A picked card is unpicked first, so nothing `undrop` cannot restore is ever taken. |
| `moai gtd undrop <n> [--expect <prefix>]` | Return the addressed dropped card to `queued`, stripping the marker. The state is the authority, so a card marked dropped by hand (no marker in its text) undrops with its text untouched. `drop` + `undrop` returns the queue file to the same bytes. |
| `moai gtd add "<text>" --force` | Admit a card the analyser reads as an exact duplicate. The card is appended verbatim and the queue records that the duplicate was forced, so the collision stays visible instead of being argued about later. |
| `moai gtd analyze` | Re-read the whole queue and record what the analyser finds. Appends, removes, reorders, and edits nothing. Re-running records nothing new — the same relation is never stacked twice. |
| `moai gtd relate <a> <b> --relation (contains\|absorbs\|replaces\|conflicts) [--note <text>]` | Record one relation between two existing cards. The verb writes a record and touches neither card; `absorbs` does not absorb. |
| `moai gtd unrelate <index>` | Remove the addressed record. The index is the one `why` prints. No card changes. |
| `moai gtd why <n>` | Print every record naming the card, or an explicit no-findings line. A card the queue knows nothing about says so rather than printing nothing. |
| `moai gtd history [<id\|n>]` | Answer what became of a card — read-only, lock-free, writes nothing. One line per lookup: `live` with the card's current state (`queued`\|`picked`\|`dropped`), `archived` with the state it held when it was closed, or `absent` when the queue holds no record — an id at or below the issued-id mark qualifies its `absent` on stderr, because a card closed by a binary predating the archive leaves none. Each `live` and `archived` line carries a landing column before the card text — `landing=<delivering commit>` when the operator recorded one, `landing=ref-head` when the record holds only the observed ref position, `landing=-` when no record was made, and `landing=malformed` when a stored record fails validation. The SHA is rendered in full here, unlike the abbreviated cell `pr` renders into its aligned table, so a closed card's delivering commit is readable without a second lookup. The card text stays LAST, so a consumer reading the tail is unaffected by the added column. A bare lookup id accepts the bare `<n>` form too. With no id, the archive lists newest-first, bounded at 20 (`--limit <n>` adjusts, `--limit 0` unbounded; a truncated listing states the withheld count on stderr). A store that cannot vouch for an archive — a database predating the archive tables, or a legacy `backlog.json` serving with no `backlog.db` — names itself on stderr and says no archive is available, rather than letting `absent` read as authoritative. |
| `moai gtd pr [<id>]` | Report each card's open pull request or landed state — read-only, and it writes nothing. The landed question is asked about the branch this project INTEGRATES on: the ref resolves from the configured worktree base branch (`origin/<that branch>`) and falls back to `origin/main` when none is configured. A project that integrates elsewhere would otherwise read every card that shipped as not-landed — silently, because an empty commit set and a wrong ref look identical. Five outcomes: `linked` (one open PR carries the card id; confidence `exact` from the PR title, `inferred` from a single PR body), `ambiguous` (several PR bodies carry it — every candidate is listed and none is chosen), `landed` (no open PR, but the resolved ref's history names the card), `no-link` (nobody has started it), and `unknown` (the landing question could not be asked — no such ref, no git, a failed query). Two limits belong to THIS list, not to the opt-in guard alone. `landed` means SOMETHING naming the card landed on that ref — NOT that the card's LAST step landed; a card whose run commit shipped reads as landed while its sync commit is still unpushed. And `unknown` is NOT evidence of not-landed: it says the question went unasked, which is a different fact from an answer of no. The row carries seven tab-separated columns — card id, outcome, pull requests, confidence, queue state, landing evidence, card text — so a `picked` card with no commits and a `queued`, never-started one no longer render alike, and an operator's recorded landing reads beside the resolver's own verdict instead of nowhere. The evidence cell is empty for a card with no record, and the marker in it says which kind of SHA is being shown: `(operator)` for a delivering commit the operator asserted, `(ref-head)` for the machine's observation of where the ref stood. `--json` carries the same record under a `landing` key, present only on a card that has one. The card text stays the LAST field, so a consumer reading the tail is unaffected by either added column. One `gh` query per invocation, never one per card, which is why the link is a separate verb rather than a column on `list`: the queue's cheapest read stays free of the network. When `gh` is absent, unauthenticated, or offline the link column renders empty, the degradation is noted on stderr, and the exit code stays 0 — the landed check is local git and keeps running. |
| `moai gtd landed <n> [--sha <sha>] [--ref <ref>]` · `--clear` | Record — or clear — what the operator observed about a card's landing, as one locked write of one column. The record carries the ref the observation was made against, that ref's head at the observation instant, the instant itself, and the SPEC status read at record time (an explicit `unknown` when it could not be read, never a guess). `--sha` adds the delivering commit **on the operator's authority**, stored together with the fact that an operator asserted it — the machine never fills this in, because a commit that mentions a card is not a commit that delivered it. A `--sha` is checked for referential integrity before anything is stored (the object must exist, and be reachable from the ref) and the check names which half failed; a check that cannot be run at all refuses rather than degrades, because a read that cannot answer may stay permissive but a write that cannot validate may not. Recording changes nothing else — not the card's state, not its position, not its text. A second `landed` replaces the record; `--clear` removes it, and the two are mutually exclusive. |
| `moai gtd auto-done [--fetch] [--dry-run] [--json]` | The lead's post-push batch closer — never a lane surface. After a batch push is confirmed on the remote (`git fetch origin develop && git rev-parse origin/develop` shows the ref moved), the lead runs it `--dry-run` first to read the planned closes, then live. The scan closes cards whose SPEC reads `completed` and whose landing is proven by one of two evidence forms — a recorded delivering SHA, or a landing commit whose subject attributes the card id (a subject carrying two distinct card tokens attributes nothing; a reissued id skips `ambiguous-id`). Everything else skips with a stated reason from the closed four-token vocabulary (`skip <id> reason=<reason>`); a close prints `done <id> landing=landed source=auto-land ref=<ref> form=<form>`. Closes apply in one locked write, an append-only JSONL log under the runtime state dir records every execution, and `undone` reverses any close. `--fetch` runs exactly one `git fetch`; absent the flag, zero network fetches. `--dry-run` writes nothing. Exit 0 covers closes and skips alike; exit 1 only when the store is unreadable. |

[HARD] `edit`, `move`, `drop`, `undrop`, `done`, and `undone` are operator
acts, exactly like `add` and the pick. Correct a card's wording, move it, or discard it because
the operator said to — never on inferred priority, never as tidy-up, never to
fold one card into another, and never because a card looks stale. The queue
records the operator's intent; it does not curate it.

[HARD] `landed` is an operator act as well, and what it writes is EVIDENCE rather than a
transition. Recording that a card's work landed does not close the card, move it, or mark it
done — the operator still decides that, with `done`. The queue stores what the operator
observed; it never acts on it.

[HARD] `auto-done` is a LEAD surface, never a lane one. It closes cards the operator already picked, after the remote landing is confirmed; a lane never runs it, and it is never a substitute for reading the evidence — the printed close and skip lines plus the audit log are what the lead reads.

[HARD] `--require-landed` is OPT-IN and honestly limited. It asks whether ANY
commit on the landed ref names the card — not whether the card's LAST step
landed — so a card whose run commit shipped reads as landed while its sync
commit is still sitting unpushed in a lane's worktree, which is the exact case
that motivated the guard. It is therefore a second pair of eyes, never a
substitute for reading the evidence. Absent the flag no landing query runs at
all, and an unanswerable query (no git, no such ref) PROCEEDS rather than
refusing: refusing on the absence of evidence would block every machine that
cannot answer. Which of those happened is now SAID rather than left to be
inferred: the stdout verdict reads `landing=unknown` when the guard could not
run and `landing=landed` when it was satisfied, so a guard that passed and a
guard that never ran are no longer the same bytes. The ref the question was
asked about is named in the refusal and in the degradation note.

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

## GTD stages

`moai gtd` carries the five GTD stages in order — capture → clarify → organize →
reflect → engage — and one verb that is not a stage at all, `answer`. Captured
GTD items stay separate from the established development queue: only an
explicitly approved Engage operation may publish one into the same
`backlog.db` the queue verbs above already use, and the queue's existing
states, ids, ordering, archive, and restore behavior are unchanged by any
stage.

Each entry below gives the usage shape the verb's own `--help` prints, and
every flag that help lists apart from `--json` and `--help`, which every verb
carries.

### Capture

`moai gtd capture <text>` captures an inbox item without publishing a card.

| Flag | Effect |
|---|---|
| `--event` | Stable capture event identity. |
| `--sensitivity` | Public, private, or secret; private when unset. |
| `--source` | Capture source; user when unset. |
| `--source-authorized` | Confirm a non-user source is authorized. |

### Clarify

`moai gtd clarify <gtd-id>` clarifies outcome, evidence, trust, and authority.

| Flag | Effect |
|---|---|
| `--outcome` | Desired outcome. |
| `--evidence` | Completion evidence. |
| `--disposition` | Action, reference, someday, waiting, or trash. |
| `--authority` | Delegated authority. |
| `--trusted` | Confirm the source is trusted. |

### Organize

`moai gtd organize <gtd-id>` organizes a clarified item and its relationships.

| Flag | Effect |
|---|---|
| `--class` | Action, project, reference, waiting, scheduled, or someday. |
| `--context` | Action context. |
| `--part-of` | Project GTD item id. |
| `--depends-on` | Prerequisite GTD item id. |
| `--review-at` | Review time. |

### Reflect

`moai gtd reflect` reviews blockers, stale evidence, and missing next actions.
It takes no argument.

| Flag | Effect |
|---|---|
| `--rebuild-projection` | Rebuild the private GTD relation projection. |

### Engage

`moai gtd engage <gtd-id>` publishes, picks, and dispatches an approved
actionable item. This is the one stage that can reach the backlog queue, and it
reaches it only when the operator approves the queue effect explicitly.

| Flag | Effect |
|---|---|
| `--approve` | Explicitly approve queue effects. |
| `--pick` | Pick the published card. |
| `--dispatch` | Record dispatch after picking. |
| `--lane` | Available lane owner. |
| `--run-id` | Stable mission run identity. |
| `--fresh` | Confirm current evidence revision. |
| `--dependencies-ready` | Confirm dependencies are complete. |
| `--resources` | Confirm resource budget. |

### Answering a gate-blocked card

- `moai gtd answer <t-id> <text>` answers a gate-blocked card; it is not a stage.

The verb belongs to the dispatch loop rather than to the GTD progression: a
lead reads the response on its next poll. It takes no flags of its own.

## Reading the records

`moai gtd list --json` emits the file's records:

```json
{
  "project_uuid": "<uuid>",
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
  ],
  "archived": []
}
```

The top level is a single JSON OBJECT — the card array lives under `items`, so a
consumer must reach it through that key. A guess that the top level is an array
(`jq '.[0]'`, `jq 'length'`) fails with a jq type error (exit 5) (t696):

```bash
moai gtd list --json | jq -r '.items[] | select(.state == "queued") | .id'  # correct
moai gtd list --json | jq '.[0]'                                            # WRONG — Cannot index object with number
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
- `state` — `queued` | `picked` | `dropped`. Three values, and `done` is not a
  fourth: a finished card leaves the live queue entirely and moves to the
  archive. A picked item stays in the file so the operator can see what is in
  flight. A dropped item stays too, carrying its `[DROPPED — <reason>] ` marker,
  recoverable with `undrop`, and rendered by `list --dropped` (the default list
  hides it behind a count line) — a discard is a decision on the record, not an
  erasure.
- `archived` — the cards `done` took out, each carrying the findings that named
  it and the position both held. No live reader sees them; `undone` puts one
  back exactly where it was. The archive is not pruned: it grows, and a
  retention policy is the operator's decision rather than the queue's.

## Picking the next card

`moai gtd next` (and the lead's own post-`/clear` opening move) presents the
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
made at the moment the card appears instead of at the next `moai gtd next`.
What makes it a pick rather than a preselect is that the question is genuinely
open — starting is one branch among the others, chosen by the operator, and no
branch is taken on their behalf when they do not answer. A workflow that starts
work without that answer has preselected, whatever it calls the step.

Once picked:

1. Record it with `moai gtd next <n> [--spec <SPEC-ID>]` (one locked write).
   Attach the SPEC only when one exists and its identifier is known.
2. Follow `kanban-dispatch.md`'s card class: Class A direct close, Class B
   run → sync without a SPEC, Class C plan → run → sync with SPEC authoring
   in plan. A pick alone neither creates a SPEC nor requires one.

## Standing sources

A standing source is a workflow the operator authorized once to issue a card
when it finishes, rather than being asked for that card every time. The
authorization is still the operator's and still comes first; what the standing
source changes is *when* they give it, not *who* gives it. A card issued this
way is the workflow carrying out an instruction already on the record — never a
tool deciding on its own that work exists.

A workflow is a standing source when — and only when — all five properties
below hold. The properties are the whole of the test, so the set of standing
sources is **conditional rather than closed**: a workflow joins it by meeting
them, never by resembling one that already has. Two meet them today:

| Standing source | Issues on | Marked |
|---|---|---|
| `/moai project` | the run completing | `[PROJECT] ` |
| the codemaps-debt trigger in `moai integration release` | the codemaps freshness layer reaching its threshold | `[GRAPH] ` |

- **One card per occasion.** Not one per document, per feature, per finding, or
  per measured file. `/moai project` issues once per run; the codemaps trigger
  issues once per debt period, which is what stops several landings inside one
  period from each producing a card.
- **Derived, never invented.** The text restates what the source already
  measured or was already told. `/moai project` takes it from that run's own
  `.moai/project/harness-spec.yaml` — its `goal`, bounded by its `scope` — so
  the card restates what the operator said in the interview; the codemaps
  trigger takes it from the freshness report's own metric, threshold, and
  content anchor. A source with nothing to derive from issues nothing; an
  empty result is reported, not filled in.
- **Marked at the front.** The text carries the source's prefix, so the queue
  shows at a glance which cards a workflow issued and which a person typed.
  The prefix is the card's provenance — the record carries no other.
- **The issued id is reported.** The source names it (`t<n>`) as it issues, so
  the card is visible in the same breath as its creation. Without this the
  queue can grow unobserved, which is the failure mode automatic issuing
  introduces and this property is the only defence against.
- **Starting it is a separate pick.** The card is queued, not started. Whether
  work begins is a separate operator choice (§ Picking the next card) — what a
  standing source creates is a queue entry, never started work.

### Not stacking duplicates

The queue itself refuses a card whose normalized text equals that of a card
already **live** in it, naming the holder and leaving the queue file
byte-identical. A standing source relies on that refusal, and a source that
also reads the queue first (`moai gtd list --json`) reports the existing id
rather than provoking it.

[HARD] **The refusal is EXACT, so a standing source's text must be constant
for as long as its card should be.** A text carrying a measured value, a
running count, or a timestamp is a different text on every issue and is never
suppressed — several sources firing in one period would each be admitted, and
the queue would fill with restatements of one condition. Key the text on
something stable for the whole period the card describes (the codemaps trigger
uses the content anchor, which moves only when the artifact is regenerated),
and put the volatile figures on the reporting line, where they inform without
becoming part of the key.

Two consequences follow from "live", and both are intended. A **dropped** card
is not a comparison subject, so dropping one does not lock its text out of the
queue forever. A **done** card leaves the live set too, so a condition that is
closed and then recurs can be raised again rather than being silenced by its
own history.

Nothing becomes a standing source by precedent. TODO comments, open issues,
audit findings, and report milestones stay outside: they are surfaced to the
operator, who asks for a card when they want one.

## Outside Kanban Mode

`moai gtd` works in an ordinary session too — it is just a queue. What it will
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
