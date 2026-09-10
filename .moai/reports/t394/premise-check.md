# t394 — premise re-check

Card t394 was dispatched on the premise that `moai todo done` DESTROYS the card
record. The dispatching orchestrator had already found that premise false and
handed down five counter-measurements. This report re-runs all five, and adds
four further measurements the SPEC rests on.

Tree under measurement: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t394`,
branch `WT-todo-done-history`, HEAD `2c18091d127cbc723074124e1015353e077300ca`
(= `origin/develop`; `git rev-list --count --left-right origin/develop...HEAD` → `0	0`).

---

## Claim

1. `moai todo done` archives rather than deletes, at develop.
2. The archive tables `archived_items` / `archived_findings` exist in the DDL.
3. `moai todo undone <id>` restores from the archive.
4. `export-json` carries the archive and discloses downgrade loss on stderr.
5. The absence of a fourth `state` value is a documented decision, not an oversight.
6. **New** — no verb renders the archive; the read path nevertheless loads it on every read.
7. **New** — `why <id>` cannot distinguish an absent card from a live card with no findings.
8. **New** — `list` renders `dropped` rows; all three live states are visible.
9. **New** — the operator's live queue database has NO archive tables, because the
   installed binary predates the archive and its `done` deletes.

All nine reproduce. Item 9 is not in the dispatch and materially bounds what the
SPEC may claim.

---

## Evidence

### 1. `done` archives

```
$ grep -n "return rec.ArchiveCard(id)" internal/cli/todo.go
409:				return rec.ArchiveCard(id)

$ grep -n "func (r \*BacklogRecord) ArchiveCard" internal/kanban/backlog_store.go
223:func (r *BacklogRecord) ArchiveCard(id string) error {
```

`internal/cli/todo.go:409` @ `2c18091d1`. Not a delete.

### 2. Archive tables exist

```
$ grep -n "archived_items\|archived_findings\|CHECK (state IN" internal/kanban/backlog_sqlite.go
99:// archived_items deliberately carries NO state CHECK: an archived row is
113:  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
124:CREATE TABLE IF NOT EXISTS archived_items (
133:CREATE TABLE IF NOT EXISTS archived_findings (
```

### 3. `undone` restores

```
$ grep -n "return rec.RestoreCard(id)" internal/cli/todo_undone.go
68:				return rec.RestoreCard(id)
```

`internal/cli/todo_undone.go:68` @ `2c18091d1`.

### 4. `export-json` carries the archive and discloses the loss

```
$ sed -n '100,110p' internal/cli/todo_export.go
// printed there would corrupt a machine read to deliver a human message.
func discloseArchiveDowngradeCost(cmd *cobra.Command, rec *kanban.BacklogRecord) {
	if len(rec.Archived) == 0 {
		return
	}
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
		"note: this export carries %d archived card(s). A release predating the archive "+
			"reads only the fields it knows, so it will discard them on its first write to the queue. "+
			"Restore anything you still need with `moai todo undone <id>` before downgrading.\n",
		len(rec.Archived))
}
```

### 5. The three-value enum is a decision, not an oversight

```
$ sed -n '95,101p' internal/kanban/backlog_sqlite.go
// pair of tables rather than a fourth `state` value. SQLite cannot ALTER a
// CHECK constraint, so admitting a fourth state would need a table rebuild on
// every operator queue in the field.
//
// archived_items deliberately carries NO state CHECK: an archived row is
// history rather than a live lifecycle position, and leaving the constraint
// off keeps the live three-value enum the single constrained surface.
```

and in the SPEC that owns it:

```
$ grep -n "REQ-TDG-00[4567]" .moai/specs/SPEC-TODO-DESTRUCTIVE-GUARD-001/spec.md
170:- **REQ-TDG-004** — The archive shall be additive: new top-level record fields and new database tables, with the five-field per-item contract and the three-value `BacklogState` enum unchanged.
171:- **REQ-TDG-005** — The stored `schema_version` shall remain `"1"`, so that a binary predating the archive continues to open a database containing archived rows.
176:- **REQ-TDG-007** — Archived rows shall not appear in any listing, count, candidate set, or duplicate-analysis input that reports on the **live queue** — specifically `list`, `next`, `why`, `analyze`, and the state counts. This requirement does not constrain `export-json`, which is governed by REQ-TDG-015.
```

### 6. No verb renders the archive, yet every read loads it

Sixteen verbs are registered (the dispatch said 16; SPEC-TDG's HISTORY says 15,
measured before `pr` joined):

```
$ sed -n '148,152p' internal/cli/todo.go
	cmd.AddCommand(newTodoAddCmd(), newTodoListCmd(), newTodoDoneCmd(), newTodoUndoneCmd(), newTodoNextCmd(),
		newTodoUnpickCmd(), newTodoEditCmd(), newTodoMoveCmd(),
		newTodoDropCmd(), newTodoUndropCmd(),
		newTodoAnalyzeCmd(), newTodoRelateCmd(), newTodoUnrelateCmd(), newTodoWhyCmd(),
		newTodoPRCmd(), newTodoExportJSONCmd())
```

`rec.Archived` is nonetheless populated on every read:

```
$ sed -n '102,104p' internal/kanban/backlog_migrate.go
	if err := e.readArchive(ctx, rec); err != nil {
		return nil, err
	}
```

The call site is inside `readRecord` (`internal/kanban/backlog_migrate.go:49`).
The archive is already in memory at every `moai todo` invocation; only the
rendering is missing.

### 7. `why` conflates absent with finding-less

```
$ sed -n '34,37p' internal/cli/todo_why.go
			if len(findings) == 0 {
				_, _ = fmt.Fprintf(out, "%s: no findings\n", id)
				return nil
			}
```

The same bytes print for an id the queue has never seen and for a live card that
simply carries no findings.

### 8. `list` renders every live row including `dropped`

```
$ sed -n '310,312p' internal/cli/todo.go
	for _, it := range rec.Items {
		_, _ = fmt.Fprintf(out, "%s\t%s\t%s\n", it.ID, it.State, it.Text)
		for _, f := range rec.Findings {
```

No state filter. Measured against the live queue (installed binary — see
Baseline-attribution):

```
$ moai todo list | grep -oE '^t[0-9]+	[a-z]+' | awk '{print $2}' | sort | uniq -c
  19 dropped
  22 picked
  71 queued

$ moai todo list | wc -c
  156468
```

112 live cards, 156,468 bytes of default output. The dispatch's figures (109
cards, ~70KB) do not reproduce; the figures above are the ones the SPEC cites.

### 9. The operator's live queue has no archive at all

```
$ sqlite3 .moai/state/todo/backlog.db ".tables"
findings  items     meta

$ sqlite3 .moai/state/todo/backlog.db "SELECT count(*) FROM archived_items;"
Error: in prepare, no such table: archived_items
```

and the binary serving it deletes rather than archives:

```
$ git show origin/main:internal/cli/todo.go | sed -n '/newTodoDoneCmd/,/^}/p' | grep -n "Remove\|Items"
15:		Short: "Remove a card from the backlog queue by id",
23:						rec.Items = append(rec.Items[:i], rec.Items[i+1:]...)
27:						rec.RemoveFindingsNaming(id)

$ git show origin/main:internal/kanban/backlog_sqlite.go | grep -c "archived_items"
0

$ moai version | head -3
╭───────────────────╮
│                   │
│  moai-adk v3.1.2  │
```

Consequence: every card closed to date is already destroyed. The surface this
SPEC specifies answers only for cards closed after an archive-capable binary is
installed. It cannot recover the harness-selection card, nor t81/t83/t88/t89 —
the two incidents that motivated the card.

---

## Baseline-attribution

- Measurements 1-8's source reads are `grep` / `sed` against this worktree's
  working tree, which is `origin/develop@2c18091d1`
  (`git rev-list --count --left-right origin/develop...HEAD` → `0	0`, run this turn).
- The `origin/main` reads in measurement 9 are `git show origin/main:<path>` at
  `origin/main@48239c7dc`, run this turn from this worktree.
- The runtime measurements (`moai todo list`, `sqlite3`, `moai version`) were
  taken against the PRIMARY checkout's live queue at
  `/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`, served by the
  **installed** binary `/Users/goos/go/bin/moai`, `v3.1.2` — main-era, roughly
  686 commits behind develop. They measure the operator's queue as it stands
  today, NOT develop's behaviour. Every claim about develop's behaviour above
  rests on a source read, never on that binary.

---

## Gaps

- **[WITHDRAWN 2026-09-01, plan-audit iter1 D1] The runtime queue path differs
  from develop's.** This gap entry asserted that develop's `BacklogPathForRoot`
  (`internal/kanban/backlog_store.go:250` @ `2c18091d1`) returns
  `.moai/state/kanban/backlog.json`. **The claim is false and its citation does
  not resolve.** It was never measured — it was inferred from the sibling card's
  text and written here as though it had been. Re-measured at `2c18091d1`:

  ```
  $ grep -rn "func BacklogPathForRoot" internal/
  internal/kanban/state_dir.go:129:func BacklogPathForRoot(root string) string {

  $ grep -n 'const stateDirName\|const legacyStateDirName' internal/kanban/state_dir.go
  37:const stateDirName = "todo"
  43:const legacyStateDirName = "kanban"
  ```

  `BacklogPathForRoot` joins the directory `resolveStateDir` returns
  (`state_dir.go:79-104`), which is the **canonical** `.moai/state/todo`
  unconditionally once that directory exists. `kanban` is the retired name the
  rename moved away from — the inverse of the withdrawn claim. There is no path
  change, so there is nothing to migrate across and no risk of an upgrade
  starting an empty queue elsewhere. The SPEC's dependency note has been
  rewritten around what card t395 actually owns (the stale sibling
  `backlog.json`); see `spec.md` §E and `plan.md` §B.2.

  Recorded rather than deleted: the entry is the worked example of the failure
  its own § Gaps section exists to prevent — an unverified premise stated as a
  measured reason, in the direction nothing downstream contradicts.
- I did NOT build or run develop's binary. No behavioural measurement of
  develop's `done` / `undone` / `export-json` exists here — only source reads.
- I did NOT measure how many cards have been closed historically; the archive is
  empty and `main`'s delete left no record, so the number is unrecoverable.
- I did NOT measure `export-json`'s runtime output. The claim that it carries the
  archive rests on the source read at `internal/cli/todo_export.go:100-110`.

## Residual-risk

- The verb count (16) is read off one `AddCommand` call. A verb registered
  elsewhere would not appear in it; I did not sweep for a second registration site.
- `readRecord` loading the archive is measured on the SQLite path only. The
  legacy-JSON read path (`loadLegacyBacklogJSON`,
  `internal/kanban/backlog_migrate.go:411`) was not traced; a queue still on JSON
  may populate `rec.Archived` by a different route, or not at all, which would
  change the cost argument for the JSON-backed case. **Partly closed
  2026-09-01**: the legacy path is `loadLegacyBacklogJSON`, a plain
  `json.Unmarshal` into `BacklogRecord`, so `rec.Archived` is populated only if
  the file carries an `archived` key. The operator's own stale
  `.moai/state/todo/backlog.json` does NOT — measured top-level keys
  `['version','last_seq','items','findings']`, 109 items against the database's
  113. Every read path prefers the database when present
  (`backlog_store.go:411-427`, `:551-568`), so the stale file cannot answer
  today; where the database is absent it answers with no archive at all. The
  SPEC absorbs this as REQ-TAQ-013's second clause.
- Measurement 9's conclusion — "already-closed cards are unrecoverable" — assumes
  the operator has run no archive-capable binary against this queue at any point.
  The absence of the tables is strong evidence, since the DDL creates them on any
  open by a newer binary, but a table dropped by hand would present identically.
