# Acceptance Criteria — SPEC-TODO-LANDING-EVIDENCE-001

Twenty-one criteria, one per requirement (Tier L ceiling: 25). AC-TLE-019 carries three
independently-asserted sub-criteria (019a/b/c) under one id, per the AC sub-ID convention; it is one
criterion against the budget. Every criterion is binary-testable and
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

**Given** a card **carrying a `spec_id` that resolves to a fixture SPEC**, a resolvable ref whose
head SHA is known to the test, and an operator-supplied `--sha` naming a commit reachable from that
ref,
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
`landed` on a different card serializes rather than interleaving (both records land intact).

**RED**: the verb absent → unknown command, non-zero exit. Writing outside the lock → the
concurrency assertion loses one of the two records, because `BacklogStore.Mutate`
(`internal/kanban/backlog_store.go:638-665`) is a whole-record read-modify-write
(`readRecord` → callback → `normalizeBacklogRecord` → `writeRecord`), so two unlocked concurrent
mutations on *different* cards genuinely drop one.

**No prompt-guard conjunct here.** The inherited `TestTodoCmd_NoAskUserQuestion`
(`internal/cli/todo_test.go:451-480`) globs `todo*.go`, so it covers `todo_landed.go` automatically
the moment the file exists, and it carries a `scanned < 2` positive control plus a synthetic-violation
negative control that a fresh grep conjunct would not. Restating it here would add a weaker duplicate.

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

**Given** a project root containing a git repository and a queue in which two cards carry landing
evidence,
**When** the test hashes every file under the project root **except everything under `.git/`**
(content plus relative path), runs `moai todo pr`, and hashes again,
**Then** the two hashes are equal; **and** the positive control holds — the hashed set is non-empty
and contains at least the queue database under `.moai/state/kanban/`.

**Why `.git/` is excluded, and why the exclusion is not a hollowing-out.** `moai todo pr` shells out
to `git` in the project working directory (`internal/kanban/prlink_landed.go:150-154` via the
`todoRunCommand` seam at `internal/cli/todo_pr.go:57-65`, which calls
`exec.CommandContext(name, args...)`), and the Given requires a git repository for the landed
question to be askable at all. Git may write inside `.git/` during a read — commit-graph writes,
`gc.log`, opportunistic ref packing — so an unexcluded `.git/` makes the two hashes differ for
reasons unrelated to the property being asserted. That is a flake, not a detection. The positive
control is what stops the exclusion from emptying the assertion: an exclusion broad enough to hash
nothing would fail the non-empty check, and one that dropped the queue directory would fail the
containment check — so the criterion cannot pass by hashing an irrelevant set.

**Scope history.** Half A's D1 delta fix widened this assertion from the queue directory to the
project root, because a queue-scoped form stayed GREEN against a cache planted outside
`StateDirForRoot` (`SPEC-TODO-LANDING-STATE-001` HISTORY 0.3.0). That widening is kept; only `.git/`
is carved back out, and the carve-out is exactly the subtree the subject under test is entitled to
touch.

**RED**: **planted mutant** — write a cache file from the `todo pr` path, planted **outside** the
queue directory and outside `.git/` (e.g. `.moai/cache/`), so it lands in the hashed set. The
run-phase MUST observe the RED and revert.

### AC-TLE-015 — seven columns, card text last, JSON key present (REQ-TLE-015)

**Given** a queue of three cards — one `linked` with a known PR number and confidence, one `landed`
and `picked`, one `no-link` and `queued` — with exactly one carrying evidence, and with the
**pre-change six-field rendering of the same fixture captured as the expected prefix**,
**When** `moai todo pr` renders,
**Then** every row splits into exactly 7 tab-separated fields; **fields 1-5 equal their pre-change
values for the same fixture** (card id, outcome, pull requests, confidence, queue state — each
asserted individually, not as a joined string); field 6 holds the evidence (empty for the cards
without it); field 7 equals that card's text verbatim; and the `--json` output's object for the
evidence-carrying card contains the record under its own key.

**Why fields 1-5 are pinned.** This is the surface's **second** contract change — half A went five
columns to six — and §G's stated residual is precisely that a consumer keys on a column position.
A criterion that pins only the count, field 6, and field 7 leaves five of seven positions free: an
implementation that reorders or relabels id / outcome / PRs / confidence / state while inserting
evidence at 6 would pass it unchanged. Pinning all seven makes the criterion cover the whole
contract it is disturbing.

**RED**: append the evidence *after* the text → field 7 is no longer the text. Emit six fields →
the field-count assertion fails. Swap the confidence and queue-state columns, or relabel an outcome
token → the fields 1-5 assertion fails while every other clause still passes. Omit the JSON key →
the last assertion fails independently of the text rendering.

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
**When** the test (a) executes the pre-change production statements verbatim —
`SELECT id, text, added_at, spec_id, state FROM items ORDER BY seq` and the matching
`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (…)` — and (b) exercises a
**reconstruction of the pre-change open path**: the `backlogDDL` const, the `schemaVersion` read, and
the version switch **exactly as they stand at HEAD `e50964ad3`**, held as a frozen test-local copy,
run against that same post-change database,
**Then** both statements succeed and the SELECT returns the pre-change field values unchanged; the
reconstructed open path returns no error and does not classify the database as `ErrBacklogCorrupt`;
and `schema_version` reads exactly `"1"`.

**What this does and does not demonstrate.** Clause (b) exercises the code path a pre-change binary
would take (`internal/kanban/backlog_sqlite.go:278-299`: DDL exec → `schemaVersion` → version
switch), which clause (a) alone never reaches. It is still a **reconstruction compiled from today's
source**, not a genuinely older build — a divergence between the frozen copy and a real released
binary would be invisible to it. `spec.md` §G keeps REQ-TLE-018 listed as an argued claim with a
partial demonstration for exactly this reason; this criterion narrows the gap, it does not close it.

**RED**: bump `backlogSchemaVersion` → the version assertion fails, **and** the reconstructed open
path rejects the database as `ErrBacklogCorrupt` at `backlog_sqlite.go:293-297` — the two clauses
fail independently, which is the point of adding (b). Give `landing` a `NOT NULL` without a default
→ the verbatim INSERT in (a) fails.

### AC-TLE-019 — the schema-freeze guard now sees column tuples, on BOTH tables (REQ-TLE-019)

Three sub-criteria, each with its **own independent plant**. They are separated for the same reason
AC-TLE-001 and AC-TLE-002 are separated: a guard extended for one table only, or for names only,
would otherwise satisfy the whole criterion while leaving the rest of REQ-TLE-019 unbuilt.

**AC-TLE-019a — the `items` plant.**
**Given** the extended `TestTodoHistoryAddsNoSchemaChange`,
**When** `ALTER TABLE items ADD COLUMN bogus TEXT` is planted and the guard runs,
**Then** the guard **FAILS**.

**AC-TLE-019b — the `archived_items` plant, asserted independently.**
**Given** the same extended guard, with **no** plant on `items`,
**When** `ALTER TABLE archived_items ADD COLUMN bogus2 TEXT` is planted alone and the guard runs,
**Then** the guard **FAILS**.

**AC-TLE-019c — the tuple plant (type / nullability / default drift).**
**Given** the same extended guard, with no extra column on either table,
**When** the `landing` column is created as `TEXT NOT NULL DEFAULT ''` instead of nullable `TEXT`,
**Then** the guard **FAILS** — a name-set assertion alone would pass this plant, so the failure
demonstrates the assertion compares `(name, type, notnull, dflt_value)` tuples.

**And with no plant at all**, the guard passes against the post-change column tuples of both tables.

**RED is already measured as the current state, on both tables.** At HEAD `e50964ad3` all four of
the guard's existing assertions stay GREEN with a column planted —
`.moai/reports/t359/measure-guard-column-blind.txt` records
`COLUMN SET (unasserted by the guard) = "seq id text added_at spec_id state bogus"` for `items`, and
the plan-audit's independent probe reproduced the same blindness on `archived_items`
(`COLUMN SET archived_items = "seq id text added_at spec_id state position bogus2"`, all four
assertions GREEN — `.moai/reports/t359/plan-audit.md` Hunt 2). The guard as it stands cannot satisfy
any of 019a / 019b / 019c. The run-phase must plant each against the **extended** guard, observe
each failure separately, and revert.

**Decay note.** The guard is owned by `SPEC-TODO-ARCHIVE-QUERY-001`. If that SPEC's owner extends it
independently on `develop`, this measurement decays and the run-phase MUST re-measure rather than
cite this criterion's recorded RED.

### AC-TLE-020 — a supplied SHA is validated for existence and reachability (REQ-TLE-020)

**Given** a fixture ref with a known history, and three `--sha` inputs: (a) a commit reachable from
that ref, (b) a syntactically valid but non-existent object id, (c) a commit that exists but is
**not** reachable from that ref (created on a detached side branch the fixture does not merge),
**When** the operator records a landing with each in turn,
**Then** (a) exits 0 and stores the record with the **full resolved SHA**, not the abbreviated input
form; (b) and (c) each exit 1, write nothing (`SELECT landing IS NULL` returns 1 for the card), and
each names on stderr which check failed — existence for (b), reachability for (c), distinguishably.

**And the attribution boundary holds**: the card id is not passed to either check. Asserted by
running (a) with the card renamed between two invocations and observing the same accept/reject
outcome — the validation result is independent of which card the record belongs to.

**RED**: skip validation entirely → (b) and (c) exit 0 and store a record, so both the exit-code and
the NULL assertions fail. Validate existence but not reachability → (c) passes when it must fail.
Store the supplied abbreviated form rather than the resolved SHA → (a)'s full-SHA assertion fails.
Collapse the two stderr messages into one → the distinguishability assertion fails.

**Why the card-id clause is not decorative.** It is the mechanical guarantee that REQ-TLE-020 stays
a referential-integrity check and never becomes attribution: an implementation that fed the card
token into the validation would produce a card-dependent result and fail the rename assertion.

### AC-TLE-021 — the two doctrine surfaces agree, and agree with the emitted row (REQ-TLE-021)

**Given** `.claude/skills/moai/workflows/todo.md` and its template mirror at
`internal/template/templates/.claude/skills/moai/workflows/todo.md`,
**When** the test extracts the `moai todo landed` verb-table row and the `moai todo pr` row from each
file and compares them, and separately parses the column count those `todo pr` rows state,
**Then** the extracted rows are byte-identical between the two files; **and** the stated column count
equals the field count AC-TLE-015 measures on the rendered row (7).

**RED**: edit the local `todo.md` without mirroring to the template (the drift this repository has
already paid for once, in the `.sh` / `.sh.tmpl` hook-wrapper pair) → the byte-identity assertion
fails. Leave either `todo pr` row saying six columns after the render emits seven → the
count-agreement assertion fails.

**This is not a doc-grep criterion.** It asserts no sentence is present; it asserts two files agree
with each other and that a number stated in prose matches a number measured from behaviour. Both
comparisons have a reachable red that no amount of pasting text can satisfy. The doctrine *prose*
deliberately has no criterion — see `spec.md` §C.7 and §G.

---

## §D.1 Severity

| Class | Criteria | Consequence of failure |
|---|---|---|
| Blocking — data | AC-TLE-003, AC-TLE-017, AC-TLE-018, AC-TLE-020 | evidence or existing rows lost, an operator queue in the field becomes unopenable, or an unvalidated SHA is stored permanently |
| Blocking — doctrine | AC-TLE-004, AC-TLE-008, AC-TLE-011, AC-TLE-012, AC-TLE-014 | the machine transitions a card, or claims a delivery it inferred |
| Blocking — contract | AC-TLE-001, AC-TLE-002, AC-TLE-015, AC-TLE-019 (a/b/c) | the stored or rendered shape is not the one specified, or drifts silently |
| Standard | AC-TLE-005, AC-TLE-006, AC-TLE-007, AC-TLE-009, AC-TLE-010, AC-TLE-013, AC-TLE-016, AC-TLE-021 | the feature is wrong but recoverable without data loss |

Every blocking criterion must pass before the SPEC may move to `implemented`.

## §D.2 Planted-mutant register

Five criteria require an observed RED via a planted mutant, and AC-TLE-019 requires three further
plants against an already-measured baseline. The run-phase records, for each: the mutant applied,
the verbatim failure output, and the SHA-1 of every touched source file **before and after** the
revert, so it is demonstrable that no mutant survived.

| Criterion | Mutant |
|---|---|
| AC-TLE-004 | evidence write also sets `state = 'dropped'` |
| AC-TLE-008 | evidence write also moves the card to the head of the queue |
| AC-TLE-011 | resolver carries its first grep match into the outcome |
| AC-TLE-014 | `todo pr` writes a cache file **outside** the queue directory and outside `.git/` |
| AC-TLE-017 | migration drops the `landing` value |
| AC-TLE-019a | `ALTER TABLE items ADD COLUMN bogus TEXT` |
| AC-TLE-019b | `ALTER TABLE archived_items ADD COLUMN bogus2 TEXT` — planted alone, with no `items` plant |
| AC-TLE-019c | `landing` created as `TEXT NOT NULL DEFAULT ''` instead of nullable `TEXT` |

AC-TLE-019's three REDs are not planted from zero: the column-blindness baseline is already measured
on BOTH tables (two independent probes — this SPEC's and the plan-audit's). The run-phase re-observes
each against the **extended** guard, separately, so that a guard extended for one table or for names
alone cannot pass all three.

AC-TLE-020 needs no planted mutant: its RED is reachable from ordinary inputs (a non-existent object
id, and an unreachable commit), which is a stronger position than a mutant — the failing input is
part of the criterion rather than a temporary edit to the source.

## §D.3 Definition of Done

- All 21 criteria pass (AC-TLE-019 counts as passed only when 019a, 019b, and 019c each pass); the
  five planted mutants plus AC-TLE-019's three plants have each been observed failing and reverted,
  with before/after file hashes recorded.
- `go test ./internal/kanban/... ./internal/cli/...` is green, run in this worktree, with the
  command and its verbatim tail recorded in `progress.md` §E.2.
- `go vet` and the project linter are clean on the touched packages.
- The two `todo.md` surfaces (local and template mirror) both state the new verb, the
  evidence-is-not-a-transition rule, and the seven-column contract; `make build` has been run so the
  template change is embedded. **Mirror parity and the stated column count are verified by
  AC-TLE-021; the prose itself is a DoD item with no criterion behind it** (`spec.md` §C.7) — this
  is the one DoD line that is deliberately not criterion-backed, and it is marked so no reader
  mistakes it for verified.
- No source file outside the module list in `spec.md` frontmatter is modified.
