---
id: SPEC-TODO-ARCHIVE-QUERY-001
title: "A read surface for the backlog archive — was this card ever issued, and what became of it"
version: "0.2.2"
status: completed
created: 2026-09-01
updated: 2026-09-01
author: manager-spec (card t394)
priority: P2
phase: "v3.1.5 target"
module: "internal/cli, .claude/skills/moai/workflows/todo.md, internal/template/templates/.claude/skills/moai/workflows/todo.md"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, cli, archive, read-only, history, query"
tier: M
related_specs:
  - SPEC-TODO-DESTRUCTIVE-GUARD-001
  - SPEC-TODO-SQLITE-001
  - SPEC-KANBAN-TODO-CLI-001
  - SPEC-TODO-ANALYSIS-001
---

# SPEC: A read surface for the backlog archive

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.2.2 | 2026-09-01 | Lead-mandated premise repair (correction round 3) after a post-snapshot re-measurement of the primary checkout's live queue (`2026-09-01 16:20–16:25 KST`). The binary claim in §A.4 and the `[MEASUREMENT PROVENANCE]` block — a `v3.1.2` main-era install that deletes on `done` — was stale from `06:37 KST` that day, when a develop build (`v3.1.2-1033-g64bba61aa`, whose `done` **archives**) replaced it; root cause: the draft read only `moai version`'s banner and missed the descriptor line printed on the same output. The archive tables the SPEC's snapshots lacked now exist — created after that install, `archived_items` = 11 rows at the `16:20` snapshot, every one recoverable via `undone` — and are recorded as snapshot C. The destroyed-card figure is restated as a **pre-archive-era estimate** (~289, corrected arithmetic `408 − 108 − 11`), explicitly separated from current archive-capable behavior. "Every card closed to date is already destroyed" corrected to closures before the install only; the harness-selection card corrected from unrecoverable to **never lost** (`t393`, live in `items`, state `picked`); t81/t83/t88/t89 stand as measured-unrecoverable. The forward-coverage boundary re-scoped from this SPEC's implementation to the install instant. §A.4 figures remain non-load-bearing by the SPEC's own design; no REQ or AC text changed. |
| 0.2.1 | 2026-09-01 | Plan-audit iter2 repair (PASS-WITH-DEBT 0.875; four blocking-class defects closed). AC-TAQ-011 clause 1 gains an adding-commit singleness assertion and a `git diff --exit-code "$C"` byte-integrity assertion, closing the modify-path escape by which goldens edited after M0 kept clause 1 green; its prose no longer claims to foreclose more than it does, and it is now explicitly scoped to the pre-integration branch. The AC-TAQ-011 develop-drift residual and the merge-method dependency it creates are recorded (`plan.md` §D). Every runtime figure re-attributed to a `[MEASUREMENT PROVENANCE]` block in §A.4 naming checkout, binary and instant; the 112/113 apparent conflict resolved as two instants of one live counter, evidenced by a same-second snapshot in which `list` and `count(*)` both return 113. The acceptance fixture convention now names its two storage-surgery exceptions instead of asserting none exist. The constructor name `newTodoHistoryCmd`, load-bearing for two ACs, fixed as a contract in `plan.md` M1. |
| 0.2.0 | 2026-09-01 | Plan-audit iter1 repair. Tier corrected S → M (the REQ/AC ceiling of 8 is exceeded by 15, and the emitted artifact set is already M's). The §E dependency note on card t395 stated the inverse of develop's behaviour on a citation that did not resolve; corrected against `internal/kanban/state_dir.go` and rewritten around what t395 actually owns. REQ-TAQ-004's disclosure re-based on the queue's durable id-space accounting rather than the archive's current emptiness. REQ-TAQ-013 extended to the legacy-JSON store. AC-TAQ-011 given a capture provenance it cannot certify itself against; AC-TAQ-014 given an existence precondition. |
| 0.1.0 | 2026-09-01 | Initial plan-phase authoring (card t394). The dispatching card's premise — that `done` destroys the record — was re-measured and found false; see `.moai/reports/t394/premise-check.md`. The SPEC follows the measured defect (no query surface) rather than the dispatched one (destruction). |

## §A Context

### A.1 What the queue already does

At `origin/develop@2c18091d1`, `moai todo done <id>` **archives**: `internal/cli/todo.go:409`
calls `rec.ArchiveCard(id)` (`internal/kanban/backlog_store.go:223`), which moves
the card and every finding naming it into `archived_items` / `archived_findings`
(`internal/kanban/backlog_sqlite.go:124,133`). `moai todo undone <id>`
(`internal/cli/todo_undone.go:68`) puts both back. `export-json` carries the
archive and discloses the downgrade loss on stderr
(`internal/cli/todo_export.go:100-110`). Nothing is destroyed.

The archive is also **already loaded on every read**: `readRecord` calls
`readArchive` unconditionally (`internal/kanban/backlog_migrate.go:102`), so
`rec.Archived` is populated in memory at every `moai todo` invocation.

### A.2 The defect

Of the sixteen verbs registered at `internal/cli/todo.go:148-152`, exactly one
touches the archive — `export-json`, and it is a whole-record dump written to
disk as a downgrade artifact, not a query. `undone <id>` requires the operator to
already know the id. REQ-TDG-007 deliberately hides archived rows from `list`,
`next`, `why`, `analyze` and the counts.

So the queue holds the answer to "was this card ever issued, and what became of
it" and offers no way to ask it. Two consequences were observed on 2026-09-01:

- The kanban lead could not answer whether a `moai init` harness-selection card
  had ever been issued. "Not in the queue now" was the whole of what it could say.
  (2026-09-01 re-measurement: that card is `t393`, live in `items` — see §A.4.
  This bullet records the incident as observed; it asserts no destruction.)
- A lane verifying card t90's entry conditions tried to read the state of
  t81/t83/t88/t89, found no record, and fell back to reading SPEC frontmatter
  status instead — substituting a different subsystem's field for the queue's own.

`why <id>` looks like it should close the second case and does not: it prints
`<id>: no findings` for an id the queue has never seen **and** for a live card
that simply carries no findings (`internal/cli/todo_why.go:34-37`). The two facts
are the same bytes.

### A.3 The three live fates, and the fourth

`list` applies no state filter (`internal/cli/todo.go:310-311`), so all three
live states render with their state in column 2. A `dropped` card is a **live
row**, not an archived one. The four distinguishable fates are therefore:

| Fate | Storage | Visible to `list` |
|---|---|---|
| `queued` | `items` | yes |
| `picked` | `items` | yes |
| `dropped` | `items` | yes |
| archived | `archived_items` | no (REQ-TDG-007) |

A surface that conflates any of these four is a defect, not a simplification.

### A.4 What the surface cannot do

An archive-capable binary has been installed since **2026-09-01 06:37 KST** —
the develop build the `[MEASUREMENT PROVENANCE]` block below describes — and
under it `done` **archives** (§A.1). Under the binary installed before that
instant — a `v3.1.2` main-era build — `done` deleted, and every closure it
performed destroyed its record. Those closures are the part of the gap this
SPEC cannot close in content: ~289 cards are estimated destroyed across that
**pre-archive era** (the gap arithmetic below; a conditional estimate, not a
measured count). This SPEC's coverage boundary is therefore that install
instant — **not this SPEC's own implementation**: from the install forward it
answers for every card the archive holds; for the pre-archive era it can only
say that an id was issued and its record is gone.

Measured, the recovery surface splits in two. t81/t83/t88/t89 are
unrecoverable — absent from both `items` and `archived_items` at the
`2026-09-01T16:20+09:00` snapshot, pre-archive-era destruction. The
harness-selection card, which an earlier draft grouped with them, was **never
lost**: it is `t393`, live in `items` (state `picked`, added
`2026-08-31T15:43:00Z`, an operator decision dated 2026-09-01 in its text),
and the current binary's `list` renders it. The 11 cards closed since the
install are likewise recoverable today via `moai todo undone`. Stating these
limits is part of the deliverable — a surface that silently reports `absent`
for a destroyed card teaches the operator a false negative.

**Why `absent` can nevertheless be qualified per id.** The queue issues ids from
a persisted high-water mark, `meta.last_seq`, which is deliberately never
derived from the ids currently present — `internal/kanban/backlog_store.go:14-17`
@ `2c18091d1` states the reason: *"`done` removes rows and a derived mark would
reuse the removed card's id"*. `normalizeBacklogRecord`
(`internal/kanban/backlog_store.go:772-778`) only ever raises the mark, never
lowers it. The mark is therefore a durable record of how many ids were ever
issued, and it survives the destruction of the rows themselves.

<a id="measurement-provenance"></a>
**[MEASUREMENT PROVENANCE — this block defines the coordinates every runtime
figure in this SPEC and in `plan.md` carries. Cited elsewhere as
"§A.4 provenance".]**

- **Checkout**: the **primary checkout** `/Users/goos/MoAI/moai-adk-go`, *not*
  the worktree this SPEC was written in. `.moai/state/todo/` does not exist in
  `.claude/worktrees/t394` (`ls -d .moai/state/todo` → *No such file or
  directory*), so every `sqlite3`, `ls` and `moai todo` transcript below is
  unrunnable at the tree §A.1 names and must not be read as runnable there.
- **Binary** — corrected 2026-09-01 (lead-mandated premise repair). The binary
  installed at the plan-phase measurement turn was a `v3.1.2` main-era build
  that deleted on `done`; at `06:37 KST` the same day it was replaced by the
  current install (`which moai` → `/Users/goos/go/bin/moai`, mtime
  `2026-09-01 06:37 KST`), a **develop** build: `moai version` prints the
  banner `moai-adk v3.1.2` **plus** the descriptor line
  `v3.1.2-1033-g64bba61aa   built 2026-08-31T21:37:48Z` — 1033 commits past
  the tag — and its `todo done` **archives**. The earlier bullet quoted the
  banner alone and missed the descriptor line on the same output; the stale
  binary claim and the "that is why the queue has no archive tables" inference
  built on it are withdrawn. Snapshots A and B below predate the install and
  were taken under the deleting binary; snapshot C records what the new binary
  created.
- **These figures are live counters, not fixed measurements.** `meta.last_seq`
  and `count(*) FROM items` move whenever the operator adds or closes a card.
  Every figure below is a snapshot at a named instant and is expected to be
  stale by the time it is read. **No requirement and no acceptance criterion
  turns on any of them** — AC-TAQ-004 constructs its own `last_seq = 5` fixture
  — so staleness costs a reader nothing but a re-measure. Anyone re-deriving an
  argument from them must re-measure rather than carry them forward.

Snapshot **A**, `2026-09-01T04:11+09:00` (the plan-phase measurement turn):

```
$ sqlite3 .moai/state/todo/backlog.db "SELECT value FROM meta WHERE key='last_seq';"
401
$ sqlite3 .moai/state/todo/backlog.db "SELECT count(*) FROM items;"
113
$ sqlite3 .moai/state/todo/backlog.db ".tables"
findings  items     meta
```

Snapshot **B**, `2026-09-01T04:30:18+09:00` — every command below issued inside
that same second, same checkout, same binary — taken to demonstrate the movement
rather than to replace snapshot A:

```
$ sqlite3 .moai/state/todo/backlog.db "SELECT value FROM meta WHERE key='last_seq';"
402
$ sqlite3 .moai/state/todo/backlog.db "SELECT count(*) FROM items;"
113
$ sqlite3 .moai/state/todo/backlog.db "SELECT state, count(*) FROM items GROUP BY state;"
dropped|20
picked|22
queued|71
$ sqlite3 .moai/state/todo/backlog.db ".tables"
findings  items     meta
$ moai todo | grep -cE '^\s*t[0-9a-z]+'
113
```

**The `112` and `113` this SPEC quotes elsewhere are two instants of one moving
counter, not a disagreement about population.** `list` applies no state filter
(`internal/cli/todo.go:310-311` @ `2c18091d1`), so `moai todo`'s card count and
`count(*) FROM items` name the same set — and snapshot B is the evidence, both
commands returning `113` within one second. The plan-phase turn recorded `112`
from `list` and `113` from `count(*)`; the record does not say which was taken
first and neither is wrong. Each was correct when taken.

Two further movements are visible in the sequence and are worth naming, because
this SPEC exists because of the second one. Between snapshot A and `04:25` the
queue gained a card (`401 → 402`, `113 → 114`). Between `04:25` and snapshot B
it **lost** one (`queued` `72 → 71`, `count(*)` `114 → 113`) while `last_seq`
held at `402` — a `done` under the deleting binary then installed, destroying
one more record during the writing of the SPEC that documents the destruction
(a pre-install event; the archive-capable binary only arrived at `06:37 KST`). The `dropped|20`
line is included because `dropped` rows are live rows (§A.3); a reader who
assumed otherwise would expect `list` to render `93`, not `113`.

At snapshot B: 402 ids issued, 113 live rows, no archive tables at all — **289**
ids issued and held by neither store at that instant. (Snapshot A gave 288 by the same
subtraction; the figure grew by exactly the one card destroyed above.) So for an
id `t<k>`, `k <= last_seq` and present in
neither store means *issued and its record is gone*, while `k > last_seq` means
*never issued*. That distinction is durable: it does not weaken when the first
card is archived, and it does not require recovering anything. What stays
impossible is the record's **content** — this SPEC still cannot say what
t81/t83/t88/t89 contained, only that they were issued and their records were
destroyed. (The first draft grouped the harness-selection card with them; that
was wrong — it is `t393`, live in `items`. See §A.4.)

**Snapshot C**, `2026-09-01T16:20+09:00` — the lead's premise re-verification
against the same primary checkout and queue, taken after the `06:37` develop
install. Measured by the lead dispatch (`16:20–16:25 KST`); the figures are
live counters at that instant, under the same re-measure discipline as A and B:

- tables: `archived_findings`, `archived_items`, `findings`, `items`, `meta`
- `meta.last_seq` = 408; `count(*) FROM items` = 108
  (`queued` 74, `picked` 14, `dropped` 20)
- `archived_items` = 11 rows (seq 1–11: t400, t332, t333, t338, t343, t350,
  t356, t357, t358, t362, t399) — every card closed since the install, each
  recoverable via `moai todo undone`
- t81, t83, t88, t89: in neither `items` nor `archived_items` — pre-archive-era
  destruction, as §A.4 states

The archive tables snapshots A and B lacked exist here: absent at `04:11` and
`04:30`, present at `16:20`, with the `06:37` install between — the `IF NOT
EXISTS` DDL described below doing what it says.

One condition on the subtraction, which is easy to carry forward wrongly:
`last_seq − count(*)` equals the destroyed count **only while the queue has no
archive tables**. That condition held at snapshots A and B — both predate the
install — so the `288`/`289` figures above stand as **pre-archive-era**
counts. Once an archive-capable binary runs one `done`, the row moves to
`archived_items` and leaves `items`, so `count(*) FROM items` falls while
`last_seq` holds, and the same subtraction over-counts destroyed cards by
exactly the number of archived rows — no longer hypothetical, since snapshot C
is post-install: `408 − 108 = 300` issued ids absent from `items`, minus the
`11` archived, gives **289** in neither store — identical to snapshot B's
count. No card has been destroyed since the install; the estimate is frozen at
the pre-archive era. REQ-TAQ-004 is not exposed to this — it is
a per-id predicate that consults **both** stores, never a difference of counts —
but this paragraph's arithmetic is, and a reader re-deriving it today must
subtract the archive before trusting the result.

The signal has one residual, stated rather than hidden: `normalizeBacklogRecord`
raises `last_seq` to cover the highest id the record *holds*, so a record whose
mark was hand-edited low, or one imported with a higher mark than it ever issued,
can leave a gap that was never an issued card. The disclosure is therefore worded
as a qualification of `absent` — *this id is at or below the mark, so it may have
been issued and destroyed* — not as a claim that it certainly was.

**The upgrade creates the archive on first open, which is why the archive's own
emptiness is the wrong signal.** `internal/kanban/backlog_sqlite.go:92-94` @
`2c18091d1`: *"this whole DDL runs on every open and every statement is IF NOT
EXISTS, so a queue created by an earlier binary gains the tables the first time
a newer one opens it"*. An archive-emptiness test therefore stops firing after
the first `moai todo done` while the destroyed cards stay destroyed forever —
see REQ-TAQ-004, which is written against the id-space fact instead.

### A.5 Why a first-class verb rather than `export-json | jq`

`export-json` technically holds the data, and the simplicity ladder requires
answering why it is not the answer:

1. It **writes to disk**. It lands a `backlog.json` beside the database as an
   atomic temp-plus-rename. A read that mutates the working tree is not a read,
   and using a downgrade route as a query surface makes every lookup produce an
   artifact the operator must then reason about.
2. It **prints a downgrade warning** on stderr on every invocation carrying an
   archive. A lookup that emits a migration caution trains the operator to ignore
   that caution.
3. It **cannot answer the question asked**. The dump distinguishes live from
   archived by which array a row sits in, but reports nothing for an id in
   neither — and "this card was never issued" is precisely what both incidents
   needed to establish.

Against that, the new surface costs no query, no schema change, and no additional
read: `rec.Archived` is already in memory (§A.1). What is being added is
rendering, not retrieval.

## §B Requirements (GEARS)

### B.1 Lookup

- **REQ-TAQ-001** — **When** an operator runs `moai todo history <id>` for an id
  held by a live row, the CLI shall print one line naming the id, the literal
  word `live`, that row's state (`queued` | `picked` | `dropped`), and the card
  text.
- **REQ-TAQ-002** — **When** an operator runs `moai todo history <id>` for an id
  held by an archived entry, the CLI shall print one line naming the id, the
  literal word `archived`, the state the card held when it was archived, and the
  card text.
- **REQ-TAQ-003** — **When** an operator runs `moai todo history <id>` for an id
  present in neither the live rows nor the archive, the CLI shall print one line
  naming the id and the literal word `absent`, and shall exit 0 — an id the queue
  does not know is an answer, not an error.
- **REQ-TAQ-004** — **Where** the looked-up id's ordinal is at or below the
  queue's persisted `last_seq` high-water mark while the id is held by neither
  the live rows nor the archive, **when** the CLI reports `absent`, it shall
  additionally state on stderr that this id lies at or below the mark of ids the
  queue has issued, so it may have been issued and its record destroyed — a card
  closed by a binary predating the archive leaves none — and `absent` here
  therefore does not establish never-issued. **Where** the ordinal is above `last_seq`, or the
  id is not of the `t<n>` form, no such note shall be printed. The condition is
  a function of the durable id-space accounting (§A.4), not of the archive's
  current emptiness, and therefore does not stop firing once the first card is
  archived.
- **REQ-TAQ-005** — `moai todo history <id>` shall accept the same id forms
  `done`, `undone`, `why` and `next` accept — a bare `<n>` normalizing to `t<n>`,
  or the explicit id.

### B.2 Listing

- **REQ-TAQ-006** — **When** an operator runs `moai todo history` with no id, the
  CLI shall print the archived entries most-recently-archived first, one line per
  entry in the same column shape REQ-TAQ-002 defines.
- **REQ-TAQ-007** — The listing shall be bounded by default, printing at most 20
  entries, and `--limit <n>` shall raise or lower that bound. `--limit 0` shall
  mean unbounded.
- **REQ-TAQ-008** — **When** the listing is truncated by the bound, the CLI shall
  state on stderr how many entries were withheld, so a truncated read is never
  mistaken for a complete one.
- **REQ-TAQ-009** — **When** the archive holds no entries, the CLI shall print an
  explicit empty-archive line rather than nothing, and shall exit 0.

### B.3 Invariants the surface must not break

- **REQ-TAQ-010** — The verb shall be read-only: it shall take no write lock,
  shall not call `Mutate`, and shall leave every byte of the queue's storage
  artifacts unchanged.
- **REQ-TAQ-011** — The default output of `list`, `list --json`, `next`, `why`,
  `analyze` and the state counts shall be byte-identical before and after this
  change, so REQ-TDG-007's invisibility survives intact.
- **REQ-TAQ-012** — The run-phase change shall not add a table, alter a `CHECK`
  constraint, add a fourth `BacklogState` value, or change `schema_version`
  (REQ-TDG-004, REQ-TDG-005).
- **REQ-TAQ-013** — **Where** the opened store cannot vouch for an archive —
  either because the database predates the archive tables, or because the answer
  is served from a legacy `backlog.json` (no `backlog.db` present), which a
  pre-archive binary writes with no `archived` field at all — the CLI shall
  degrade to reporting the live rows, shall name on stderr which store answered,
  and shall say that no archive is available, rather than failing. A store that
  cannot carry an archive must not be allowed to answer `absent` as though it
  could.
- **REQ-TAQ-014** — The verb shall not prompt the user; a missing input is
  reported and the verb exits.

### B.4 Documentation

- **REQ-TAQ-015** — The verb shall be documented in
  `.claude/skills/moai/workflows/todo.md` and in its template mirror
  `internal/template/templates/.claude/skills/moai/workflows/todo.md`, and the
  mirror shall carry no SPEC ID, REQ token, internal date, or commit SHA
  (template neutrality).

## §C Acceptance criteria

Enumerated in `acceptance.md`, which is the Tier M artifact set's third file.
`acceptance.md` is the canonical enumeration — do not restate the AC here, or the
two surfaces will drift.

## §D Scope boundaries

### Out of Scope — storage

- Any new table, altered `CHECK` constraint, fourth `BacklogState` value, table
  rebuild, or `schema_version` bump. SQLite cannot `ALTER` a `CHECK` constraint,
  so a fourth state would force a rebuild of every operator queue in the field —
  the reason the archive is a table pair in the first place
  (`internal/kanban/backlog_sqlite.go:95-101`; REQ-TDG-004, REQ-TDG-005).
- Any change to what `ArchiveCard` / `RestoreCard` store.

### Out of Scope — the default queue read

- Any change to what `list`, `next`, `why`, `analyze` or the state counts print by
  default. The live queue measured **112 cards and 156,468 bytes of `list`
  output** — a live counter, plan-phase measurement turn `2026-09-01T~04:11+09:00`,
  §A.4 provenance (primary checkout, then-installed `v3.1.2` — replaced by the
  develop build at `06:37 KST` that day). The order of magnitude
  is what the argument rests on, not the digits: folding the archive into that
  read would grow the most-used surface to answer a rare question. REQ-TDG-007
  stands.
- Any `--archived` flag on `list`. A flag on the queue's cheapest read invites the
  invisibility invariant to be relaxed one caller at a time; a distinct verb keeps
  the boundary where it is checkable.

### Out of Scope — the stale sibling `backlog.json`

- Removing, repairing, or reconciling the stale `backlog.json` that currently
  sits beside `backlog.db` in `.moai/state/todo/`. Measured at the plan-phase
  measurement turn `2026-09-01T~04:11+09:00`, §A.4 provenance (primary checkout,
  then-installed `v3.1.2` — these paths do not exist in the worktree this SPEC was
  written in): `backlog.db` (2026-09-01 02:37), `backlog.json`
  (2026-08-31 21:13), `backlog.json.migrated` (2026-08-27 23:01) coexist, and
  the stale `backlog.json` parses with top-level keys
  `['version','last_seq','items','findings']` — **no `archived` key** — and 109
  items against the database's 113 at that instant. Sibling card **t395** owns
  that repair; this SPEC does not touch the file.

  The load-bearing fact here is the **absence of the `archived` key**, which is a
  property of the file's shape and does not move. The item counts and mtimes are
  live and will have moved — in particular, the `109` vs `113` gap is not a fixed
  quantity: `backlog.json` is frozen at its last write while the database keeps
  moving, so the gap widens on its own. What REQ-TAQ-013's second clause needs is
  only that the file exists, is stale, and carries no archive.
- **The consequence this SPEC does absorb** is a read-time one, not a repair:
  every read path prefers the database when it exists
  (`internal/kanban/backlog_store.go:411-427` `QueuedCount`, `:551-568`
  `LoadPure`), so while `backlog.db` is present the stale file never answers.
  When it is absent, that file answers authoritatively, carries no archive, and
  would let this verb report a confident `absent` for a card that exists —
  precisely the false negative the verb exists to prevent. REQ-TAQ-013 makes
  that case say which store answered instead of answering silently.

### Out of Scope — migration

- Whether an upgrade migrates an existing queue, and the one-time legacy-JSON →
  SQLite cutover itself. Sibling card **t395** owns migration. See §E.
- Recovering the **content** of cards destroyed by a pre-archive binary.
  Impossible by construction (§A.4); this SPEC only makes the destruction
  legible, per REQ-TAQ-004.

### Out of Scope — search

- Full-text search of card text, whether over the live rows or the archive. The
  first incident (§A.2) would be served by it, but the archived listing serves it
  too when piped to `grep`, and a query language is a much larger surface than
  either incident justifies.

## §E Cross-references and dependencies

- **SPEC-TODO-DESTRUCTIVE-GUARD-001** — owns the archive itself. REQ-TDG-003
  (archive on `done`), REQ-TDG-004/005 (additive, no schema bump), REQ-TDG-007
  (invisible to live readers), REQ-TDG-015 (`export-json` carries it). This SPEC
  adds a reader and changes none of them.
- **SPEC-TODO-SQLITE-001** — owns the storage engine and the `export-json`
  downgrade route.
- **SPEC-TODO-ANALYSIS-001** — owns `why` / `analyze` / `relate`, whose
  finding-less output this SPEC's lookup must not be confused with.
- **Dependency note — card t395.** An earlier draft of this note recorded a
  queue-path divergence between the installed binary and develop. That claim was
  wrong in both its citation and its direction, and is withdrawn. Re-measured at
  `2c18091d1`:

  - `BacklogPathForRoot` is at `internal/kanban/state_dir.go:129`, not
    `backlog_store.go:250` (`grep -rn "func BacklogPathForRoot" internal/`
    returns `state_dir.go:129` and `state_dir.go:137` and nothing else).
  - `state_dir.go:37` — `const stateDirName = "todo"`; `state_dir.go:43` —
    `const legacyStateDirName = "kanban"`. `.moai/state/todo` is develop's
    **canonical** directory and `.moai/state/kanban` is the **legacy** name the
    rename moved away from — the inverse of what the draft asserted.
  - `resolveStateDir` (`state_dir.go:79-104`) returns the canonical directory
    unconditionally once it exists, which it does. There is no path change to
    migrate across, and no risk that an upgrade starts an empty queue at a
    different location.

  What t395 actually owns is the **stale sibling `backlog.json`** left inside
  `.moai/state/todo/` — three coexisting files, one of them a pre-archive
  snapshot a lane already read and was misled by. That finding IS load-bearing
  here, and it is recorded in §D (Out of Scope — the stale sibling
  `backlog.json`) and absorbed as a read-time obligation by REQ-TAQ-013. This
  SPEC records the coupling and absorbs none of the repair.
