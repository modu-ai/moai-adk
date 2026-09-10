# SPEC-TODO-SQLITE-001 — Design

Companion to spec.md (requirements) and plan.md (milestones). Decisions D1–D8 below
are binding for run-phase unless the orchestrator re-delegates a change to manager-spec
(D-NEW-1 pattern).

## 1. Driver (D1)

`modernc.org/sqlite` (pure Go). Decision inputs measured in research.md §7:

| Axis | `modernc.org/sqlite` | `mattn/go-sqlite3` | Verdict |
|---|---|---|---|
| CGO_ENABLED=0 CI builds | compiles | CANNOT compile (cgo required) | decisive; ci.yml GOOS matrix pins `CGO_ENABLED: "0"` |
| windows/amd64 (release-blocking) | supported | cgo toolchain friction | modernc |
| Binary size delta vs 62 MB baseline | ~MB-order (measured M1, budget C-4: >+12 MB triggers review) | smaller core but cgo overheads elsewhere | accept + measure |
| Query throughput at this scale (<10³ rows, per-mutation single txn) | 1.5–2× slower than C sqlite — microseconds against our millisecond budgets | faster | immaterial here |
| Maintenance/dep-tree | single pure-Go tree; large generated codebase, known heavy lint surface (mitigation §9) | classic, mature; carries cgo everywhere | modernc |
| Existing precedent in go.mod | none today (first) | none | neutral |

"Keep JSON" alternative considered and rejected as directed: the operator set SQLite as
the goal; nothing measured contradicts viability (single-file store, hundreds of rows,
sub-ms queries).

Integration posture: `database/sql` front door with `sql.Open("sqlite", dsn)` once per
process where needed (lazy on first store use), pragmas applied per connection at open:
`journal_mode=WAL`, `busy_timeout=5000`. No ORM, no migration framework — schema is
two `CREATE TABLE IF NOT EXISTS` statements plus a meta table and version marker,
idempotent on open.

## 2. Schema (D2)

```sql
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
-- rows: ('schema_version', '1'), ('last_seq', '<int>')

CREATE TABLE IF NOT EXISTS items (
  seq      INTEGER PRIMARY KEY,          -- mirrors t<N>; preserves insertion order
  id       TEXT    NOT NULL UNIQUE,      -- 't<N>'
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,             -- RFC3339 UTC, verbatim from legacy
  spec_id  TEXT,                         -- NULL ↔ JSON null (SpecID pointer semantics)
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
);

CREATE TABLE IF NOT EXISTS findings (
  subject_id TEXT  NOT NULL,
  related_id TEXT  NOT NULL,
  relation   TEXT  NOT NULL,
  source     TEXT  NOT NULL,
  score      REAL  NOT NULL,
  note       TEXT  NOT NULL DEFAULT '',
  at         TEXT  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_state ON items(state);
```

Rationale points:

- **No AUTOINCREMENT / sqlite_sequence**: `meta.last_seq` is the single explicit
  high-water mark, mirroring the legacy file's normalize semantics
  (`max(persisted, max-present-id)` applied during load AND after mutation), keeping
  parity tests trivially expressible.
- **Ordering contract (REQ-TOSQ-004)**: reads issue `ORDER BY seq`; both the items list
  and the findings "insertion order" (unrelate's index addressing must survive DELETE
  churn exactly as it did against the array rewrite) are round-trip tested.
- **Findings tuple uniqueness stays APPLICATION-level** (`AppendFindingOnce`),
  deliberately not a UNIQUE index: legacy data could in principle carry duplicate
  tuples and lossless doctrine forbids letting a constraint reject them at migration.
  Items' id uniqueness IS structural today → safe as a DB constraint.
- **Flat, foreign-key-free** as directed: findings reference item ids that may already
  be deleted (`RemoveFindingsNaming` runs on card removal, so orphans should not exist
  — but enforcement belongs to the ported application logic, not FK constraints that
  would turn a legacy quirk into an import failure).
- WAL mode: readers never block the writer; matches the lock-free `Load()` contract.

## 3. Layout resolution & directory rename (D3, t309 absorption)

One order of operations, enforced idempotently on the ADOPTING open path (the CLI verb
path, never the web read path — mirroring today's pure-vs-adopting split):

1. Resolve queue root (unchanged resolvers).
2. **Directory layer**: if `<root>/.moai/state/todo/` exists → it wins; old
   `.moai/state/kanban/`, if also present, is left strictly untouched (stale-copy
   policy). If todo dir is absent and kanban dir present → rename the DIRECTORY
   (same-volume atomic; all registry `<uuid>.json` files ride along).
3. **Storage layer**: inside `.moai/state/todo/`, apply lazy json→db migration (§4).

Fallback READ requirement: any open path that finds ONLY the old layout resolves it by
performing step 2 then proceeding; environments that cannot relocate (cross-device,
permissions) fail over to READING the old location best-effort rather than erroring —
the queue stays usable, exactly like `ResolveTodoQueueRoot`'s fail-open posture.
Read-only surfaces (web console render, statusline) resolve through PURE logic only —
they may observe either directory but never move bytes.

## 4. Migration state machine (D4)

States per queue root (single file lock `backlog.lock` held across ANY transition):

```
A {no db, no json}                       → open empty db (first run of new binary)
B {no db, json}                          → MIGRATE (REQ-TOSQ-011..014)
C {db, no json}                          → steady state
D {db, json}                             → db authoritative; re-quarantine best-effort (REQ-TOSQ-013)
E {db.migrated-quarantine present}       → inert rollback source; never touched except by export/rollback docs
```

MIGRATE algorithm (state B):

1. Acquire `backlog.lock` (the SAME artifact/concurrency protocol REQ-TOSQ-008 keeps —
   concurrent factory lanes serialize here; second process observes state C and skips).
2. Load legacy record fully (missing/malformed handling inherited from current `load()`
   contract: malformed = abort surface, never delete).
3. Create database at `backlog.db` (fresh), apply pragmas + DDL.
4. Single transaction: insert all items (seq derived from `t<N>` ids; any non-conforming
   id aborts to state-B-no-op with structured error), findings in original order;
   write `last_seq = max(persisted, max-present)`; write `schema_version`.
5. Parity verification INSIDE the same lock, BEFORE authority flips: reload via the new
   reader and compare field-for-field against the in-memory source record
   (counts, per-item equality including SpecID null-shape, findings sequence and
   tuples, last_seq).
6. On success: `rename backlog.json → backlog.json.migrated` (quarantine). On ANY
   failure: close + remove the partial `backlog.db` (+ `-wal`/`-shm`), leave JSON
   authoritative, return structured error naming both files (REQ-TOSQ-012).
7. Crash windows: between 4 and 6 leaves state D — db authoritative, quarantine
   completes next open; between 2 and 4 (partial tmp artifacts) — removed via the same
   failure path; both benign and enumerated here because factory lanes can crash
   mid-window.

## 5. Write-path discipline (D5)

The outer cross-process serialization REMAINS the existing `backlog.lock` acquisition
(identical retry window). Inside it, mutations become short SQL transactions instead of
whole-file rewrites; WAL removes the reader/writer contention the old scheme was
indifferent to (reads stayed lock-free there, they stay lock-free here through WAL
snapshot reads).

Justification for KEEPING the file lock (the judgement call flagged in research.md §8):
(a) mechanical identity with the concurrency protocol every existing test and ten-lane
factory operation already exercises; (b) cross-version serialization during upgrade
windows — a downgraded old binary only understands the file lock, so dropping it would
let mixed-binary fleets race; (c) the busy_timeout alone serializes writers but gives
readers of D-state compounds no fence while a downgrade-export executes. Cost: one
advisory lock acquisition per mutation (~µs), already paid today.

Id issuance inside `Mutate`: read `meta.last_seq` for update within the transaction,
increment, insert, commit — plus the post-normalize invariant (max clear of present
ids) ported verbatim. Unique-violation on id maps to the named id-conflict error.

## 6. Error taxonomy (D6)

Port-and-extend the `IsBoardLockHeld` style alongside the existing
`backlog_store_errors_test.go` branches:

| Engine condition | Named domain error | Caller-visible behavior |
|---|---|---|
| busy_timeout expiry / SQLITE_BUSY | `ErrBacklogBusy` | mutation refused, files untouched, message names the root |
| corrupt/unreadable db (header, integrity quick-check) | `ErrBacklogCorrupt` | NEVER deleted/overwritten; operator action documented |
| constraint violation on insert | id-conflict error | aborts the transaction, prior state intact |
| legacy JSON malformed during migration | current malformed-load error shape | abort cutover (REQ-TOSQ-012) |
| open/relocation of directories fails (EXDEV, perms) | fallback-read degradation | serve old layout best-effort, log-grade context |

## 7. Rollback mechanics (D7 — decision recorded per dispatch premise C)

Two-part answer, replacing "config knob OR automatic export":

- AT cutover: automatic one-time quarantine `backlog.json.migrated` (pre-migration
  exact bytes).
- AFTER cutover: deliberate regeneration via an ADDITIVE CLI verb
  `moai todo export-json` writing current live state as valid legacy JSON at the queue
  root. Downgrade procedure = export-json → replace binary with previous release →
  previous binary reads only `backlog.json` and ignores `backlog.db*` entirely
  (verified by construction: its path constant targets the json filename only).
  Documented in docs tree with an artifact-meaning table (`.db`, `-wal`, `-shm`,
  `.migrated`).

Justification against a config knob: two always-live engines = permanent second
implementation of every regression class this swap closes (silent divergence, dual
write paths) for a scenario one additive verb covers. The knob is rejected on
simplicity doctrine, not capability grounds.

## 8. Statusline / hot-read strategy (D8)

Current cost model: whole-file parse of an 118 KB JSON per statusline render
(process-per-render). Post-change: the count query opens the (WAL) database cold and
issues three cheap aggregates (`COUNT(*) … GROUP BY state`). Cold-open with WAL warm
files is expected at low single-digit milliseconds; hard evidence comes from the C-2
benchmark harness (fixture ≥500 items, median+p95 report checked into progress notes),
with the 25 ms ceiling as the actionable gate. Fail-open behavior preserved byte-for-
byte (unreadable ⇒ Available=false ⇒ nothing rendered).

## 9. Testing matrix

| Suite | Covers |
|---|---|
| Migration parity property tests (synthesized fixtures: empty / queued-only / mixed-states-heavy / findings+drops / last_seq>max / malformed-json red) | REQ-TOSQ-011/012/014 |
| Directory-rename matrix (only-old / both-dirs / registry-file census before-after / EXDEV simulation) | REQ-TOSQ-015 |
| Concurrency stress: N goroutine-procs × Add/Mutate under the kept file lock; process-pair spawn variant where testable | REQ-TOSQ-008 |
| Round-trip equivalence: seed JSON → migrate → dump via existing writer shape → field-equality vs source | SC-3 |
| Characterization port: existing `backlog_store_test.go` suite exercised against the SQLite-backed store UNCHANGED wherever behavioral (contract proofs port intact); CLI golden outputs asserted untouched | REQ-TOSQ-007/010 |
| Error-path battery extending `backlog_store_errors_test.go` taxonomy | REQ-TOSQ-006 |
| Pragma configuration assert (journal_mode reads wal, busy_timeout ≥ 5000 at every opened connection) | REQ-TOSQ-003 |
| `GOOS=windows go vet ./internal/...` compile gate + CI windows job verdict | REQ-TOSQ-002 / C-3 |
| Perf micro-benchmarks: mutate latency, statusline cold-open, binary size delta | C-2/C-4 |

Fixture privacy rule (C-5): generators synthesize cards' text domains; production card
texts never enter the repo.

## 10. Known risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | modernc/sqlite lint/tooling weight (generated-code warnings, govulncheck noise) | thin wrapper confined to one file cluster; lint scoping decided in M2 with evidence, never blanket-suppressed without a written reason |
| R2 | `-wal`/`-shm` confuse downgraded binaries or backup tooling | old binaries ignore unknown files by construction; rollback docs enumerate artifacts (D7) |
| R3 | ten-lane fleet mid-migration storm | all transitions behind the kept file lock; crash-window enumeration §4.7 |
| R4 | stale old-directory edits after rename (operator writes to dead path) | both-dirs stale-copy policy documents dominance; old dir left visible precisely so divergence is observable, never silently eaten |
| R5 | unrelate index addressing drifts after DELETE churn in SQL ordering | round-trip suite asserts insertion-order addressing explicitly (§9 row 3 companion assertion) |
| R6 | Windows file-lock interplay (AV scanners holding .db/-wal; LockFileEx semantics across directory rename differ from POSIX) | busy_timeout absorbs transient holds; ErrBacklogBusy surfaces named otherwise; lock-held-rename degradation covered by the dedicated M6 test (plan.md M6) |
