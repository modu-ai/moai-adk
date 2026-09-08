# M1 Evidence — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

Verbatim command + output pairs for Milestone M1 (the stored shape). Exported to this tracked path
per `agent-common-protocol.md` § Parallel Execution — evidence export obligation: `/tmp` and
`.moai/state/verify/` are machine-local scratch and gitignored, so nothing cited from there reaches
a clone, a CI runner, or an auditor.

**Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359` (confirmed by
`git rev-parse --show-toplevel` at the head of each batch).
**Branch**: `WT-landing-evidence`.
**Pre-change baseline HEAD**: `903bcc03c`.
**M1 commit**: `3bcb0c33a`. **Follow-up commit**: `2dacb1d83`.

> Tree-identity note. `/Users/goos/moai/moai-adk-go` is NOT a sibling checkout — it is the SAME
> tree under a second spelling (`stat -f '%d:%i'` → `16777231:253706617` for both;
> `git -C /Users/goos/moai/moai-adk-go rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go`).
> The real hazard is worktree → primary drift: a command run from the primary checkout succeeds and
> returns plausible bytes. The discriminant used throughout is that the two trees differ in HEAD and
> branch (`3bcb0c33a` / `WT-landing-evidence` vs `7ad9f8534` / `main`).

---

## Step 0 — re-measuring the decayed baseline

`acceptance.md` AC-TLE-019 carries a decay note: the guard is owned by
`SPEC-TODO-ARCHIVE-QUERY-001` and may have been extended on develop. The recorded RED was therefore
NOT cited; it was re-measured in this tree at `903bcc03c`.

Baseline hashes before any edit:

```
$ shasum internal/kanban/backlog_sqlite.go internal/kanban/backlog_schema_freeze_test.go
964cae182ed58310e3706b3c222cf4d8aba58508  internal/kanban/backlog_sqlite.go
32efad029ac3e98e5e7bed3b11228e856aae6d38  internal/kanban/backlog_schema_freeze_test.go
```

### Step 0a — `items` column planted, CURRENT (unextended) guard

```
$ go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v
planted items bogus via DDL
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.400s
```

### Step 0b — `archived_items` column planted alone, CURRENT guard

```
$ go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v
planted archived_items bogus2 via DDL
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.404s
```

Revert confirmed:

```
$ shasum internal/kanban/backlog_sqlite.go
964cae182ed58310e3706b3c222cf4d8aba58508  internal/kanban/backlog_sqlite.go
```

**Result**: both plants stayed GREEN → the guard was still column-blind on BOTH tables at
`903bcc03c` in this tree. The SPEC's premise holds; no decay. The lane was clear to proceed.

**Plant mechanism note.** Step 0's plants were made by adding the column to the `backlogDDL` const
rather than by an `ALTER`. An unconditional `ALTER` at that point failed the second open with
`duplicate column name: bogus` — a failure for the wrong reason, which would not have measured
column-blindness. The DDL-const plant produces the same observable (the column present on a fresh
fixture database) without that confound. The `ALTER`-shaped plants required by §D.2 were run in
Step 5, once the idempotent helper existed to carry them.

### Step 0c — the current tuples, probed before writing any expectation

```
$ go test ./internal/kanban/ -run TestT359Probe -count=1 -v   # throwaway probe, removed after
    items: seq|INTEGER|0|NULL
    items: id|TEXT|1|NULL
    items: text|TEXT|1|NULL
    items: added_at|TEXT|1|NULL
    items: spec_id|TEXT|0|NULL
    items: state|TEXT|1|NULL
    archived_items: seq|INTEGER|0|NULL
    archived_items: id|TEXT|1|NULL
    archived_items: text|TEXT|1|NULL
    archived_items: added_at|TEXT|1|NULL
    archived_items: spec_id|TEXT|0|NULL
    archived_items: state|TEXT|1|NULL
    archived_items: position|INTEGER|1|NULL
--- PASS: TestT359Probe (0.01s)
```

---

## Step 1 — guard extended against the PRE-change tuples, observed GREEN

`TestTodoHistoryAddsNoSchemaChange` gained exact ordered
`(name, type, notnull, dflt_value)` tuple assertions, per table, written against the column set as
it was — **without** `landing`. Expected strings at this moment ended `state:TEXT:1:NULL` (items)
and `position:INTEGER:1:NULL` (archived_items).

```
$ go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.239s
```

This ordering is `plan.md` §B.1's requirement: the new assertion is written against the OLD column
set, so it cannot have been written to match whatever was already built.

---

## Step 2 — the ALTER lands; the extended guard FAILS on BOTH tables

The intermediate state: `landing TEXT` added by `ensureLandingColumn`, expectations NOT yet updated.

```
$ go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v
=== RUN   TestTodoHistoryAddsNoSchemaChange
    backlog_schema_freeze_test.go:107: items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:0:NULL
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL
    backlog_schema_freeze_test.go:118: archived_items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:0:NULL
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL
--- FAIL: TestTodoHistoryAddsNoSchemaChange (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.461s
```

This is the strongest available demonstration that the guard is no longer column-blind: the same
plant that was invisible in Step 0 is now loud, on both tables.

**Known limit of this record.** Steps 2 and 3 land in the SAME commit by design (`plan.md` §F M1.3),
so this intermediate FAILURE exists in no commit — it is a transcript observation, captured live and
in order, not reconstructed. This file is the only durable carrier of it. Line numbers `:107` /
`:118` are the file at that intermediate state; after `gofmt` and the Step-3 update the same
assertions sit at `:108` / `:120`.

---

## Step 3 — expectations updated in the SAME commit as the ALTER, GREEN again

```
$ go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1 -v
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.397s
```

---

## Step 4 — the new criteria tests

```
$ go test ./internal/kanban/ -run 'TestBacklogLanding' -count=1 -v
=== RUN   TestBacklogLanding_ItemsColumnShape
--- PASS: TestBacklogLanding_ItemsColumnShape (0.01s)
=== RUN   TestBacklogLanding_ArchivedItemsColumnShape
--- PASS: TestBacklogLanding_ArchivedItemsColumnShape (0.00s)
=== RUN   TestBacklogLanding_MigrationIsIdempotent
--- PASS: TestBacklogLanding_MigrationIsIdempotent (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.472s
```

| Test | Criterion |
|---|---|
| `TestBacklogLanding_ItemsColumnShape` | AC-TLE-001 — items sequence ends in `landing`; that column is `TEXT`, `notnull=0`, `dflt_value` NULL |
| `TestBacklogLanding_ArchivedItemsColumnShape` | AC-TLE-002 — archived_items sequence, asserted independently |
| `TestBacklogLanding_MigrationIsIdempotent` | AC-TLE-003 — a pre-change database opened twice gains the column exactly once, every row's `(seq, id, text, added_at, spec_id, state)` tuple unchanged, row count 3 |

---

## Step 5 — the three §D.2 plants, each ALONE against the EXTENDED guard

Base hash of `internal/kanban/backlog_sqlite.go` before the plant series:

```
$ shasum internal/kanban/backlog_sqlite.go
01872b7bc1050b33aa1ee9eb3c326c726d4207c6  internal/kanban/backlog_sqlite.go
```

### AC-TLE-019a — `ALTER TABLE items ADD COLUMN bogus TEXT`

```
--- hash AFTER plant ---
ab925560bc9c3bcf648f4f20e98c78e4a6c381e4  internal/kanban/backlog_sqlite.go
--- FAIL: TestTodoHistoryAddsNoSchemaChange (0.01s)
    backlog_schema_freeze_test.go:108: items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:0:NULL bogus:TEXT:0:NULL
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:0:NULL
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.403s
--- hash AFTER revert ---
01872b7bc1050b33aa1ee9eb3c326c726d4207c6  internal/kanban/backlog_sqlite.go
```

### AC-TLE-019b — `ALTER TABLE archived_items ADD COLUMN bogus2 TEXT`, planted ALONE

No `items` plant present. The archived-table assertion is what fires, which is the point of
asserting it independently.

```
--- hash AFTER plant ---
28ef6218cf041532713f384e612ee3d940c2d830  internal/kanban/backlog_sqlite.go
--- FAIL: TestTodoHistoryAddsNoSchemaChange (0.01s)
    backlog_schema_freeze_test.go:120: archived_items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:0:NULL bogus2:TEXT:0:NULL
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:0:NULL
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.400s
--- hash AFTER revert ---
01872b7bc1050b33aa1ee9eb3c326c726d4207c6  internal/kanban/backlog_sqlite.go
```

### AC-TLE-019c — `landing` created as `TEXT NOT NULL DEFAULT ''`

No extra column on either table. A name-set assertion would PASS this plant; the failure is what
demonstrates the assertion compares `(name, type, notnull, dflt_value)` tuples.

```
--- hash AFTER plant ---
82ffa500276783d8451c0059ea42389edc458a71  internal/kanban/backlog_sqlite.go
--- FAIL: TestTodoHistoryAddsNoSchemaChange (0.01s)
    backlog_schema_freeze_test.go:108: items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:1:''''''
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL landing:TEXT:0:NULL
    backlog_schema_freeze_test.go:120: archived_items column tuples =
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:1:''''''
        want
         seq:INTEGER:0:NULL id:TEXT:1:NULL text:TEXT:1:NULL added_at:TEXT:1:NULL spec_id:TEXT:0:NULL state:TEXT:1:NULL position:INTEGER:1:NULL landing:TEXT:0:NULL
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.532s
--- hash AFTER revert ---
01872b7bc1050b33aa1ee9eb3c326c726d4207c6  internal/kanban/backlog_sqlite.go
```

`landing:TEXT:1:''''''` is the rendered tuple: notnull `1`, `dflt_value` `''''''` (SQL `quote()` of
the empty-string default), against a want of notnull `0`, `dflt_value` `NULL`.

### Plant-survival check

All three reverts return the file to `01872b7bc1050b33aa1ee9eb3c326c726d4207c6`, the pre-plant hash.
No mutant survived. Confirmed independently at the tree level:

```
$ git status --short
 M internal/kanban/backlog_schema_freeze_test.go
 M internal/kanban/backlog_sqlite.go
?? internal/kanban/backlog_landing_test.go
```

Only the three M1 deliverables differ from baseline — no plant residue.

---

## Extra — AC-TLE-003's own RED (not required by the dispatch)

Run to show the idempotence test is not vacuous. Mutant: make the `ALTER` unconditional rather than
deciding from `pragma_table_info`.

```
--- hash AFTER plant ---
c37c432252ca5dd7a7b9b90d783878a577a9a0e6  internal/kanban/backlog_sqlite.go
--- FAIL: TestBacklogLanding_MigrationIsIdempotent (0.01s)
    backlog_landing_test.go:70: open 2: add items.landing …/backlog.db: SQL logic error: duplicate column name: landing (1)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.407s
--- hash AFTER revert ---
01872b7bc1050b33aa1ee9eb3c326c726d4207c6  internal/kanban/backlog_sqlite.go
```

This is `acceptance.md` AC-TLE-003's stated RED, observed.

---

## Final scoped verification (at M1 commit `3bcb0c33a`)

```
$ gofmt -l internal/kanban/
(no output)

$ go vet ./internal/kanban/...
vet rc=0

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.728s

$ git status --short
 M internal/kanban/backlog_schema_freeze_test.go
 M internal/kanban/backlog_sqlite.go
?? internal/kanban/backlog_landing_test.go
```

Commit:

```
$ git show --stat --oneline HEAD
be8f36383 feat(kanban): add nullable landing column + column-tuple schema guard (t359)
 internal/kanban/backlog_landing_test.go       | 240 ++++++++++++++++++++++++++
 internal/kanban/backlog_schema_freeze_test.go |  66 +++++++
 internal/kanban/backlog_sqlite.go             |  71 +++++++-
 3 files changed, 374 insertions(+), 3 deletions(-)
```

`be8f36383` was amended to add the required `🗿 MoAI` trailer, becoming `3bcb0c33a` — same tree,
unpushed, no other change.

---

## Follow-up commit `2dacb1d83` — response to a PostToolUse SQL finding

A PostToolUse security check reported `sql-injection (high) ... SQL built by string concatenation
instead of parameters` against the in-flight work. Two interpolation sites were examined:

| Site | Disposition |
|---|---|
| `backlog_landing_test.go` `rowCount(t, eng, table)` — `` `SELECT count(*) FROM ` + table `` | **Removed.** One caller, always the literal `"items"`; the parameter earned nothing. Now `itemsRowCount(t, eng)` with the table fixed in the statement. |
| `backlog_sqlite.go:363` @ `2dacb1d83` — `fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT", table, backlogLandingColumn)` | **Kept, behavior unchanged**; a comment now states the constraint. SQLite cannot bind an identifier as a parameter, so interpolation is the only available shape. `table` ranges over the in-source `var landingCarryingTables = []string{"items", "archived_items"}`; `backlogLandingColumn` is `const = "landing"`. Nothing caller-derived reaches either, and neither may ever be fed from a runtime value. |

> Coordinate decay, recorded rather than silently fixed. This row first cited
> `backlog_sqlite.go:359`, the line at `3bcb0c33a`. The four comment lines this same commit
> (`2dacb1d83`) added above the statement moved it to `:363`, so the original citation was correct
> when written and stale one commit later. Verified now:
> `grep -n 'ALTER TABLE %s ADD COLUMN' internal/kanban/backlog_sqlite.go` → `363:`. A line citation
> decays exactly like a HEAD reading, which is why every coordinate in this file is anchored to a
> SHA.

The value-bearing probe is correctly parameterised and was not touched:

```go
`SELECT name FROM pragma_table_info(?) WHERE name = ?`
```

Post-change verification:

```
$ grep -rn '`+table|"+table|+ table' internal/kanban/backlog_landing_test.go internal/kanban/backlog_sqlite.go
NO string-concatenated SQL remains in the two files

$ gofmt -l internal/kanban/
(no output)

$ go vet ./internal/kanban/...
VET rc=0

$ go test ./internal/kanban/ -run 'TestBacklogLanding|TestTodoHistoryAddsNoSchemaChange' -count=1 -v
=== RUN   TestBacklogLanding_ItemsColumnShape
--- PASS: TestBacklogLanding_ItemsColumnShape (0.01s)
=== RUN   TestBacklogLanding_ArchivedItemsColumnShape
--- PASS: TestBacklogLanding_ArchivedItemsColumnShape (0.00s)
=== RUN   TestBacklogLanding_MigrationIsIdempotent
--- PASS: TestBacklogLanding_MigrationIsIdempotent (0.00s)
=== RUN   TestTodoHistoryAddsNoSchemaChange
--- PASS: TestTodoHistoryAddsNoSchemaChange (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.416s

$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	137.483s
```

### Disposition — the instruction, the deviation, and the review

Recorded in full because the outcome was good, and a good outcome is exactly the circumstance in
which the fact of a deviation gets smoothed away. If it is erased here, the next deviation — with a
worse outcome — has no baseline to be judged against.

| | |
|---|---|
| **What the lane was told** | The lead read `backlog_sqlite.go:359`, classed the finding a FALSE POSITIVE, and instructed: *"do not change it … Leave both as they are."* |
| **What the lane did** | Changed it anyway. The instruction arrived after `2dacb1d83` had landed; on re-reading the finding the lane judged that it pointed at a different site than the one the lead had inspected, and reported the divergence rather than silently conforming. |
| **How it was resolved** | The lead retracted the false-positive call on review: *"That was wrong — I had read `backlog_sqlite.go:359` and generalized 'this site is safe' into 'the finding is a false positive' without checking where the finding actually pointed. The word `concatenation` was the discriminator and `fmt.Sprintf` is not concatenation."* The deviation was **accepted on review**, not merely tolerated. |

The discriminator is worth keeping: the finding said **concatenation**. `fmt.Sprintf` is not
concatenation, so the production `ALTER` could not have been the site the checker named. The actual
site was `` `SELECT count(*) FROM ` + table `` in the test file — a `+` on a string, which is.

### Two claims that must not be collapsed

**(a) The finding was not a false positive.** It pointed at a real string-concatenated SQL
statement, which existed and has been removed.

**(b) That site was not dangerous.** `rowCount` was test-local; its `table` parameter had exactly
one caller, passing the literal `"items"`; no runtime value could reach it. There was no reachable
path by which caller-derived data entered the statement.

Both are true at once. (a) says the lead's retraction was correct about *where they had looked*;
(b) says nothing was at risk. A claim withdrawn for lack of evidence has not thereby been shown
false, and the converse holds too — a claim vindicated about its target is not thereby a near-miss.
Collapsing (a) into "a security hole was found and closed" would misrepresent a bookkeeping error
about where the reader's eye had been as a narrowly-averted incident.

**The removal's actual justification is simplification, not security.** The `table` parameter earned
nothing — one caller, one literal — so it went, and the concatenation went with it. Had the
parameter been carrying its weight, the correct response would have been to document the constraint
(as was done for the production `ALTER`), not to restructure the helper.

---

### The scanner re-fires on this file — that is noise on documentation, not a defect

After `2dacb1d83` landed, the SQL-injection check fired again at `m1-evidence.md:330` — the row in
the table above that **quotes** the removed `` `SELECT count(*) FROM ` + table `` as documentation.
The scanner matched a quoted example in markdown, not live code.

All three candidate sites were enumerated rather than reading one and generalizing:

| Site | Shape | Disposition |
|---|---|---|
| `m1-evidence.md:330` | this file quoting the removed concatenation | markdown, not code — scanner matched a quoted example |
| `backlog_schema_freeze_test.go:142` @ `2dacb1d83` | `` `SELECT … ` + `FROM pragma_table_info(?)` `` | two **adjacent string literals** joined for line wrapping; no data interpolated — `table` goes through the `?` bind |
| `backlog_sqlite.go:363` @ `2dacb1d83` | `fmt.Sprintf` over an in-source literal + a const | identifiers cannot be bound; the only available shape, already commented |

Zero live injection sites.

[HARD] **Do not rewrite this file to dodge the scanner.** Quoting the removed defect is the whole
point of the record; laundering the text to avoid a pattern match would erase what the document
exists to preserve. If the finding recurs on every read of this file, say so — it is scanner noise
on documentation.

---

## Corrections to this record

Two statements made earlier in this card's reporting are superseded. Recorded rather than silently
overwritten, so a reader of the transcript can reconcile it against this file.

1. **"`progress.md` §E.2 was not updated"** — true when written (the M1 dispatch scoped the work to
   three Go files and said to touch nothing else), **superseded by `d6420c1bd`**, which populated
   §E.2 with the M1 subsection. It is no longer a Gap, and a stale Gap reads as an open hole.
2. **"the sibling `/Users/goos/moai/…` was never touched"** — rests on the falsified sibling
   premise. The corrected form is the tree-identity note at the head of this file: it is the SAME
   tree under a second spelling, so the operative claim is the branch/HEAD discriminant
   (`WT-landing-evidence` vs `main`), not non-touching.

---

## What is NOT in this file (known losses — do not cite these later)

- **The throwaway probe's own source.** `zz_t359_probe_test.go` was created, run once, and deleted;
  only its output above survives. It was never committed.
- **Per-plant hashes of the two TEST files.** Only `backlog_sqlite.go` was hashed per plant, because
  no plant modified a test file. The substitute evidence is the tree-level `git status --short`
  above, which shows exactly three changed files and no residue.
- **Timing/latency measurements.** `ensureLandingColumn` adds two `pragma_table_info` queries per
  engine open; open latency was NOT measured before or after.
- **Any non-darwin build or test run.** Everything here is darwin/arm64. Windows and Linux are CI's
  verdict.
- **`internal/cli` after the M1 edits.** Measured at baseline `903bcc03c` (rc=0) by the lane, NOT
  re-run after the change. Scoped per the lane-local verification rule; CI owns the full verdict.
- **`golangci-lint`.** Not run. Only `gofmt -l` (clean) and `go vet` (rc=0).
- **Concurrent-open behavior.** `TestBacklogLanding_MigrationIsIdempotent` opens sequentially. Two
  processes opening a pre-change database simultaneously could both observe the column absent and
  both issue the `ALTER`; the loser would fail its open. Not exercised, not covered by any criterion.
