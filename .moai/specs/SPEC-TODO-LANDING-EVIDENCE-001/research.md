# Research — SPEC-TODO-LANDING-EVIDENCE-001

Every measurement this SPEC rests on: the command, its verbatim output, and the tree it ran against.
Nothing here is carried over from another tree or another time. What was **not** measured is in
`spec.md` §G, not omitted.

**Tree**: worktree `.claude/worktrees/t359`, branch `WT-landing-evidence`, HEAD `e50964ad3`.
**Measured**: 2026-09-03.

---

## §R.1 Half A is landed (the stated dependency precondition)

```
$ grep -n "^status:" .moai/specs/SPEC-TODO-LANDING-STATE-001/spec.md
5:status: completed

$ git merge-base --is-ancestor c9f712232 develop; echo "rc=$?"
rc=0
```

Both re-run in this worktree. The ancestry check decays as `develop` moves; it is re-read at
run-phase entry per `plan.md` §C rather than trusted from here.

## §R.2 `REQ-TODO-013` permits additive change (D1, first half — CONFIRMED)

```
$ grep -n "REQ-TODO-013" .moai/specs/SPEC-KANBAN-TODO-CLI-001/spec.md
59:- **REQ-TODO-013** (Ubiquitous) The backlog store shall preserve the existing version-1 record
shape — `{"version":1,"items":[{"id","text","added_at","spec_id","state"}]}` with
`state ∈ {queued, picked, dropped}` — changing it only additively (the high-water mark, per
REQ-TODO-009).
```

(Line wrapped for display; the source is one line.) The requirement constrains the *manner* of
change and names its own precedent. It does not freeze the field set.

## §R.3 The two completed SPECs agree (D1, second half — the card's premise is FALSIFIED)

```
$ awk 'NR==51 {print}' .moai/specs/SPEC-TODO-ANALYSIS-001/spec.md
- 항목당 5필드(`id`, `text`, `added_at`, `spec_id`, `state`)는 SPEC-KANBAN-TODO-CLI-001 REQ-TODO-013
이 고정한 계약이다. 다만 그 SPEC의 §E는 **자기 범위**의 out-of-scope 선언이지 영구 동결이 아니고,
같은 REQ가 "additively" 변경을 허용한다 — `last_seq` 최상위 필드가 그 선례다.
```

The clause after 다만 is the whole point: §E is a scope-local declaration, not a permanent freeze,
and the same REQ permits additive change. `SPEC-TODO-ANALYSIS-001` does **not** judge the opposite of
§R.2 — it states the same reading. No completed SPEC was edited, and none needs to be.

The precedent it names is present in the tree:

```
$ grep -n "type BacklogRecord" -A6 internal/kanban/backlog_store.go
187:type BacklogRecord struct {
188:	Version  int                   `json:"version"`
189:	LastSeq  int                   `json:"last_seq"`
190:	Items    []BacklogItem         `json:"items"`
...
```

## §R.4 The constraints that shape the design (D2 and the third constraint)

```
$ grep -n "REQ-1.10" -A5 .moai/specs/SPEC-KANBAN-QUEUE-PR-SYNC-001/spec.md
251:**REQ-1.10** — The resolver shall not name, return, or otherwise claim which
252:commit delivered a card. The `landed` outcome is a boolean fact about
253:`origin/main` and nothing more. (Grounds: §C.2 — a card's first matching commit
254:may be another card's report commit that merely mentions it, so any
255:"first match is the delivering commit" reading attributes wrongly.)

259:**REQ-2.1** — [HARD] The read surface shall leave
260:`.moai/state/kanban/backlog.json` byte-identical across an invocation, and shall
261:write no field, no `findings[]` entry, and no timestamp.
```

REQ-1.10's **grounds** are the load-bearing part, and are why §B.3 resolves on provenance rather
than asking for permission. REQ-2.1 is why the writer cannot be `todo pr` (§B.2).

## §R.5 The schema-freeze guard is column-blind (measured, not inferred)

A temporary probe (`internal/kanban/zz_t359_probe_test.go`) planted
`ALTER TABLE items ADD COLUMN bogus TEXT` and re-ran all four assertions
`TestTodoHistoryAddsNoSchemaChange` makes. Full output:
`.moai/reports/t359/measure-guard-column-blind.txt`.

```
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
=== RUN   TestT359ProbeGuardIsColumnBlind
    A1 table set     = "archived_findings archived_items findings items meta" -> match=true
    A2 CHECK present = true
    A2 stored items SQL now = CREATE TABLE items (
      seq      INTEGER PRIMARY KEY,
      ...
      state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
    , bogus TEXT)
    A3 index set     = "idx_items_state" -> match=true
    A4 schema_version = "1" -> match=true
    COLUMN SET (unasserted by the guard) = "seq id text added_at spec_id state bogus"
--- PASS: TestT359ProbeGuardIsColumnBlind (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.408s
```

Two facts follow. The guard makes **no** column-set assertion, so this SPEC's own column would land
silently (⇒ REQ-TLE-019). And the `strings.Contains` CHECK test survives an added column because
SQLite appends it *after* the closing constraint — visible in the A2 output above.

The probe was deleted immediately after the measurement:

```
$ rm internal/kanban/zz_t359_probe_test.go
$ git status --short
?? .moai/reports/t359/
```

No source file under `internal/` is modified.

## §R.6 The items SQL surface, and why the downgrade claim is buildable

```
$ grep -rn "SELECT .*FROM items\|INSERT INTO items" internal/kanban/*.go | grep -v _test
internal/kanban/backlog_migrate.go:60:		`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq`)
internal/kanban/backlog_migrate.go:276:			`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, ?, ?)`,
internal/kanban/backlog_migrate.go:340:		`SELECT COUNT(*) FROM items WHERE state = ?`, string(BacklogStateQueued)).Scan(&n); err != nil {
internal/kanban/backlog_migrate.go:351:	rows, err := e.db.QueryContext(ctx, `SELECT state, COUNT(*) FROM items GROUP BY state`)

$ grep -rn "SELECT \*" internal/kanban/ internal/cli/todo_pr.go; echo "rc=$?"
rc=1
```

Every statement names its columns; there is no `SELECT *`. This is the basis for REQ-TLE-018 — and
it is an argument, not a demonstration: no pre-change binary was actually run against a post-change
database (`spec.md` §G).

## §R.7 `ALTER TABLE` is new to this codebase

```
$ grep -rn "ALTER TABLE" internal/kanban/ internal/cli/; echo "rc=$?"
rc=1
```

No output. There is no in-tree idempotence pattern to copy, which is why `design.md` §2 states one
rather than assuming the reader supplies it.

## §R.8 The version gate rejects any bump

```
$ awk 'NR>=292 && NR<=298' internal/kanban/backlog_sqlite.go
	case backlogSchemaVersion:
		// current layout
	default:
		return fmt.Errorf("schema %s: unsupported schema_version %q (want %q): %w",
			e.dbPath, version, backlogSchemaVersion, ErrBacklogCorrupt)
```

A `schema_version` bump is a downgrade break for every operator queue in the field, not a feature —
hence the [HARD] no-bump constraint in `plan.md` §D and the §D exclusion in `spec.md`.

## §R.9 Baseline test state

```
$ go test ./internal/kanban/ -run 'TestTodoHistoryAddsNoSchemaChange' -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.408s
```

Scoped to the guard under discussion. A full pre-edit baseline over the two touched packages is
`plan.md` §C's pre-flight obligation and belongs to run-phase entry, not to this document — running
it here would produce a figure that has decayed by the time it is used.
