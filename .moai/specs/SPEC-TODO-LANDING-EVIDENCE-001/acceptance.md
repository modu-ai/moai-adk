# Acceptance Criteria — SPEC-TODO-LANDING-EVIDENCE-001

Nineteen criteria, one per requirement (Tier L ceiling: 25). Every criterion is binary-testable and
carries an explicit **RED** clause naming the change that makes it fail. Where a criterion could
otherwise pass vacuously, the RED is a **planted mutant** the run-phase must observe failing and
then revert — the pattern half A adopted after its own AC-TLS-008 was found satisfiable by a mutant
(`SPEC-TODO-LANDING-STATE-001` HISTORY 0.2.0/0.3.0).

**Prefer a criterion a mutant can trip.** No criterion here is satisfied by the presence of a name,
a comment, or a doc sentence; each asserts an observable value.

---

## §D Acceptance matrix

### AC-TLE-001 — the `items` column exists, with exactly the declared shape (REQ-TLE-001)

**Given** a backlog database created by the post-change binary,
**When** the test reads `pragma_table_info('items')`,
**Then** the column-name sequence is exactly `seq id text added_at spec_id state landing`, and the
`landing` row reports type `TEXT`, `notnull = 0`, and `dflt_value` NULL.

**RED**: remove the `ALTER TABLE items ADD COLUMN landing TEXT` statement → the sequence is missing
its final element. Also red on giving the column a `DEFAULT` or `NOT NULL`, which the shape
assertion catches independently of the name assertion.

### AC-TLE-002 — `archived_items` carries the same column (REQ-TLE-002)

**Given** the same database,
**When** the test reads `pragma_table_info('archived_items')`,
**Then** the column-name sequence is exactly `seq id text added_at spec_id state position landing`.

**RED**: add the column to `items` only → this sequence lacks `landing` while AC-TLE-001 still
passes. The two criteria are deliberately separate so a half-applied migration is caught.

### AC-TLE-003 — the migration is idempotent and rewrites nothing (REQ-TLE-003)

**Given** (a) a database built by the **pre-change** DDL — constructed in-test by creating the
tables without `landing` and inserting three cards — and (b) that same database after a post-change
open,
**When** the post-change binary opens it twice in succession,
**Then** both opens return no error, `pragma_table_info` reports the column exactly once after each,
and every pre-existing row's `(seq, id, text, added_at, spec_id, state)` tuple is unchanged and the
row count is 3.

**RED**: issue the `ALTER` unconditionally rather than deciding from column metadata → the second
open fails with SQLite's `duplicate column name: landing`. Separately red on any implementation that
rebuilds the table, which changes `seq` assignment or row order.

### AC-TLE-004 — the state CHECK is untouched and no write moves a card (REQ-TLE-004)

**Given** a queue holding one `queued` and one `picked` card,
**When** `moai todo landed` records evidence for each,
**Then** the stored `items` DDL still contains `CHECK (state IN ('queued','picked','dropped'))`, and
each card's `state` reads the same value it held before the write.

**RED**: a mutant that sets the recorded card's `state` to `'dropped'` (or adds a fourth CHECK
value) → the per-card state assertion fails without any name matching. The run-phase MUST plant this
mutant, observe both cards' assertions fail, and revert.

### AC-TLE-005 — a record carries all six facts (REQ-TLE-005)

**Given** a card, a resolvable ref whose head SHA is known to the test, and an operator-supplied
`--sha`,
**When** the operator records a landing and the test decodes the stored `landing` value,
**Then** the decoded record's six fields — ref, ref head SHA, observation instant, SHA provenance,
operator-supplied SHA, SPEC status — are all present and equal to the supplied or observed values.

**RED**: drop any one field from the encoder → that field's assertion fails. The criterion asserts
six values, so it cannot be satisfied by an encoder that emits a subset.

### AC-TLE-006 — absence is NULL, and renders as absent (REQ-TLE-006)

**Given** a queue with one card that has never had a landing recorded,
**When** the test reads `SELECT landing IS NULL FROM items WHERE id = ?` and runs `moai todo pr`,
**Then** the SQL predicate returns 1, and the card's evidence column in the rendered row is empty
while its outcome column carries whatever the resolver returned.

**RED**: encode an empty record (`{}` or a zero-valued struct) instead of leaving NULL → the SQL
predicate returns 0. Separately red on rendering `not-landed` in the evidence column, which the
outcome-column assertion isolates from the resolver's own answer.

### AC-TLE-007 — the verb writes one record, under the lock, and asks nothing (REQ-TLE-007)

**Given** a queue,
**When** `moai todo landed <id> --sha <sha>` runs,
**Then** it exits 0, exactly one card's `landing` is non-NULL, and a concurrently issued second
`landed` on a different card serializes rather than interleaving (both records land intact); and
`grep -rn AskUserQuestion internal/cli/todo_landed*.go` returns no match.

**RED**: the verb absent → unknown command, non-zero exit. Writing outside the lock → the
concurrency assertion loses one of the two records.

### AC-TLE-008 — the write moves nothing else in the queue (REQ-TLE-008)

**Given** a queue of five cards in mixed states with one carrying a `spec_id`,
**When** evidence is recorded for one of them,
**Then** the ordered list of `(id, state, position, text, spec_id)` tuples for **all five** cards is
identical before and after, and the item count is 5.

**RED**: **planted mutant** — make the verb also set the recorded card's `state` to `'dropped'`, or
move it to the head of the queue. Either flips the tuple list. The run-phase MUST observe this RED
and revert. This is a **behavioural whole-queue** assertion, not a grep for a mutating call, because
half A recorded that a name-based sweep was satisfied by a mutant.

### AC-TLE-009 — re-record replaces; `--clear` removes (REQ-TLE-009)

**Given** a card with a recorded landing naming SHA `A`,
**When** the operator records again with SHA `B`, and then runs `moai todo landed <id> --clear`,
**Then** after the second record the stored value decodes to exactly one record naming `B` and none
naming `A`, and after `--clear` `SELECT landing IS NULL` returns 1.

**RED**: append rather than replace → the decode yields two records, or a value containing `A`.
`--clear` implemented as a no-op → the NULL predicate returns 0.

### AC-TLE-010 — the SPEC status is read, and an unreadable one is not invented (REQ-TLE-010)

**Given** two cards: one whose `spec_id` names a fixture SPEC whose frontmatter reads
`status: implemented`, and one whose `spec_id` names a SPEC document that does not exist,
**When** evidence is recorded for both,
**Then** the first record's `spec_status` is `implemented`, and the second's is the unknown marker —
neither empty-as-if-read, nor a default such as `completed` or `draft`.

**RED**: default the status on an unreadable SPEC → the second assertion fails. Hard-code the status
→ the first assertion fails when the fixture's frontmatter is changed to `completed`, which the test
does as its second phase.

### AC-TLE-011 — the resolver names no commit (REQ-TLE-011)

**Given** a fixture ref whose history contains three commits whose messages all name card `tX`, with
their SHAs known to the test,
**When** the landed resolver answers for `tX` and the test renders the full outcome (struct fields
and `--json`) to a string,
**Then** none of the three SHAs — full or abbreviated to 7 characters — appears anywhere in that
string, and the answer is one of the three landing values.

**RED**: **planted mutant** — have the resolver carry its first match into the outcome. One SHA then
appears and the containment assertion fails. This is the criterion that keeps REQ-1.10 verified
rather than merely restated; the run-phase MUST observe its RED.

### AC-TLE-012 — a stored SHA is operator-supplied or absent (REQ-TLE-012)

**Given** the same three-match fixture ref,
**When** the operator records a landing **without** `--sha`,
**Then** the stored record's delivering-SHA field is empty, its provenance field is not `operator`,
and none of the three known SHAs appears in the stored value (the ref head SHA is asserted
separately by AC-TLE-013 and is excluded from this containment set by construction — the fixture
places the three matching commits below the head).

**RED**: fill the field from the first grep match → the field is non-empty and one known SHA appears.

### AC-TLE-013 — the ref position is keyed and labelled as a ref position (REQ-TLE-013)

**Given** a record produced without `--sha`,
**When** the test decodes the stored value and renders it,
**Then** the ref head SHA appears under a key distinct from the delivering-SHA key, and the rendered
form labels it as a ref position; the two keys are never the same key and the delivering-SHA key is
absent rather than aliased to the head.

**RED**: collapse the two into one key, or alias the delivering SHA to the ref head → the
distinct-keys assertion fails.

### AC-TLE-014 — `moai todo pr` still writes nothing, project-wide (REQ-TLE-014)

**Given** a project root containing a queue in which two cards carry landing evidence,
**When** the test hashes **every file under the project root** (content plus relative path), runs
`moai todo pr`, and hashes again,
**Then** the two hashes are equal.

**RED**: **planted mutant** — write a cache file from the `todo pr` path. The scope is the **whole
project root, not the queue directory**, because half A's D1 delta fix recorded that a
queue-directory-scoped form stayed GREEN against a cache planted outside `StateDirForRoot`
(`SPEC-TODO-LANDING-STATE-001` HISTORY 0.3.0). The run-phase MUST plant the mutant *outside* the
queue directory, observe the RED, and revert.

### AC-TLE-015 — seven columns, card text last, JSON key present (REQ-TLE-015)

**Given** a queue of three cards, one carrying evidence,
**When** `moai todo pr` renders,
**Then** every row splits into exactly 7 tab-separated fields, field 7 equals that card's text
verbatim, field 6 holds the evidence (empty for the cards without it), and the `--json` output's
object for the evidence-carrying card contains the record under its own key.

**RED**: append the evidence *after* the text → field 7 is no longer the text. Emit six fields →
the field-count assertion fails. Omit the JSON key → the last assertion fails independently of the
text rendering.

### AC-TLE-016 — assertion and observation are machine-distinguishable (REQ-TLE-016)

**Given** two cards — one with an operator-asserted `--sha`, one recorded without,
**When** both rows are rendered and both JSON records decoded,
**Then** the two evidence cells differ in a marker that is present independently of the SHA value:
substituting the asserted card's SHA for the other card's ref head SHA in the test fixture leaves
the two cells still distinguishable.

**RED**: render both identically apart from the SHA text → after the substitution the two cells
compare equal. The substitution step is what stops this criterion passing on the SHA values alone.

### AC-TLE-017 — the round trip preserves evidence, and its parity check is not vacuous (REQ-TLE-017)

**Given** a record containing one live card and one archived card, both carrying landing evidence,
**When** the record is exported to legacy JSON and migrated back, and the migration's parity
verification runs,
**Then** the verification passes and both cards' decoded evidence equals the original.

**RED**: **planted mutant** — make the migration drop the `landing` value. The run-phase MUST
observe that the parity verification then **FAILS** (not merely that the equality assertion fails),
because the risk this criterion exists for is a parity check that reports success while dropping
rows — the exact hazard `internal/kanban/backlog_migrate.go:604-608` records for archived rows.

### AC-TLE-018 — an older binary still serves the database (REQ-TLE-018)

**Given** a database carrying the `landing` columns,
**When** the test executes the pre-change production statements verbatim —
`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq` and the matching
`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (…)` — and reads
`schema_version`,
**Then** both statements succeed, the SELECT returns the pre-change field values unchanged, and
`schema_version` reads exactly `"1"`.

**RED**: bump `backlogSchemaVersion` → the version assertion fails, and the pre-change open path
rejects the database as `ErrBacklogCorrupt`
(`internal/kanban/backlog_sqlite.go:293-297`). Give `landing` a `NOT NULL` without a default → the
verbatim INSERT fails.

### AC-TLE-019 — the schema-freeze guard now sees columns (REQ-TLE-019)

**Given** the extended `TestTodoHistoryAddsNoSchemaChange`,
**When** an extra column is planted (`ALTER TABLE items ADD COLUMN bogus TEXT`) and the guard runs,
**Then** the guard **FAILS**; and with no planted column it passes against the post-change column
sets.

**RED is already measured as the current state.** At HEAD `e50964ad3` all four of the guard's
existing assertions stay GREEN with `bogus` planted —
`.moai/reports/t359/measure-guard-column-blind.txt`, `COLUMN SET (unasserted by the guard) = "seq id
text added_at spec_id state bogus"`. This criterion therefore starts red by construction: the guard
as it stands cannot satisfy it. The run-phase must plant the column against the **extended** guard,
observe the failure, and revert.

---

## §D.1 Severity

| Class | Criteria | Consequence of failure |
|---|---|---|
| Blocking — data | AC-TLE-003, AC-TLE-017, AC-TLE-018 | evidence or existing rows lost, or an operator queue in the field becomes unopenable |
| Blocking — doctrine | AC-TLE-004, AC-TLE-008, AC-TLE-011, AC-TLE-012, AC-TLE-014 | the machine transitions a card, or claims a delivery it inferred |
| Blocking — contract | AC-TLE-001, AC-TLE-002, AC-TLE-015, AC-TLE-019 | the stored or rendered shape is not the one specified, or drifts silently |
| Standard | AC-TLE-005, AC-TLE-006, AC-TLE-007, AC-TLE-009, AC-TLE-010, AC-TLE-013, AC-TLE-016 | the feature is wrong but recoverable without data loss |

Every blocking criterion must pass before the SPEC may move to `implemented`.

## §D.2 Planted-mutant register

Five criteria require an observed RED via a planted mutant. The run-phase records, for each: the
mutant applied, the verbatim failure output, and the SHA-1 of every touched source file **before and
after** the revert, so it is demonstrable that no mutant survived.

| Criterion | Mutant |
|---|---|
| AC-TLE-004 | evidence write also sets `state = 'dropped'` |
| AC-TLE-008 | evidence write also moves the card to the head of the queue |
| AC-TLE-011 | resolver carries its first grep match into the outcome |
| AC-TLE-014 | `todo pr` writes a cache file **outside** the queue directory |
| AC-TLE-017 | migration drops the `landing` value |

AC-TLE-019's RED is not planted in run-phase from zero — it is measured already (above) and is
re-observed against the extended guard.

## §D.3 Definition of Done

- All 19 criteria pass; the five planted mutants have each been observed failing and reverted, with
  before/after file hashes recorded.
- `go test ./internal/kanban/... ./internal/cli/...` is green, run in this worktree, with the
  command and its verbatim tail recorded in `progress.md` §E.2.
- `go vet` and the project linter are clean on the touched packages.
- The two `todo.md` surfaces (local and template mirror) both state the new verb, the
  evidence-is-not-a-transition rule, and the seven-column contract; `make build` has been run so the
  template change is embedded.
- No source file outside the module list in `spec.md` frontmatter is modified.
