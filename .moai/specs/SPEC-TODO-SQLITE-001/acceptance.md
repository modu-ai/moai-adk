# SPEC-TODO-SQLITE-001 — Acceptance Criteria

Every AC names: WHEN it runs, the INPUT that makes it RED today (against
`WT-todo-sqlite@d29b8942e`), and WHERE the red surfaces. Two-cell discipline per AC:
**RED-now** (fails against current tree as-is) and **green path** (the command that
must pass post-implementation). Commands run from the card worktree root unless stated;
all fixtures are synthesized from the observed schema (spec.md C-5).

## D. AC Matrix

| AC | Verifies | Scenario (GWT core) | RED-now proof | Green evidence |
|----|----------|---------------------|---------------|----------------|
| AC-TOSQ-001 | REQ-TOSQ-011 | **Given** a scratch root seeded with a synthetic legacy `backlog.json` (mixed states incl. drops, ≥1 finding tuple, `last_seq` > max present id) **when** `go test ./internal/kanban -run TestMigrationParity` runs **then** migrated DB matches item-for-item (counts, field equality incl. SpecID-null shape), findings order+tuples, and `last_seq`. | Test name does not exist → suite failure ("no tests to run" surfaces red). | Suite PASS with named subtests for each parity axis. |
| AC-TOSQ-002 | REQ-TOSQ-011, REQ-TOSQ-005 | **Given** the migrated fixture above **when** store `Add("probe")` runs **then** the issued id is `t<last_seq+1>` — never reused, never reset. | Characterization in same suite; red via missing test. | Subtest asserts id continuity across cutover. |
| AC-TOSQ-003 | REQ-TOSQ-014 | **Given** migration completed on the scratch root **when** directory listing runs **then** `backlog.json` absent AND `backlog.json.migrated` exists byte-equal to the seed (sha256 compared before/after by test). | Red via missing test. | Quarantine byte-preservation test PASS. |
| AC-TOSQ-004 | REQ-TOSQ-013 | **Given** db + legacy json both present **when** any adopting open runs **then** db state served; json re-quarantined best-effort without erroring the call. | Red via missing test. | Both-exist precedence test PASS. |
| AC-TOSQ-005 | REQ-TOSQ-006, REQ-TOSQ-012, C-6 | **Given** a MALFORMED legacy json (truncated bytes) and no db **when** open/migrate attempted **then** structured error, NO db file created-or-left, seed file bytes unchanged (sha256 assert). | Red via missing test. | Abort-no-destruct test PASS; mirrors existing malformed-load contract shape. |
| AC-TOSQ-006 | REQ-TOSQ-015 | **Given** only `.moai/state/kanban/` under a scratch root containing backlog.json + N synthetic `<uuid>.json` registry files **when** adopting CLI open (`moai todo list` against injected root or direct store call) runs **then** all N+1 files resolve under `.moai/state/todo/` and the read succeeds. Registry census (file count + one sampled record equality) asserted. | Command errors / lists empty today against that layout name. | Census test PASS. |
| AC-TOSQ-007 | REQ-TOSQ-015 stale-copy | **Given** BOTH directories exist **when** reads/writes run **then** todo-dir wins and kanban dir mtimes/contents remain untouched (dir-hash compare pre/post). | Red via missing test. | Stale-copy inviolability test PASS. |
| AC-TOSQ-008 | REQ-TOSQ-015 fallback READ | **Given** relocation simulated to fail (EXDEV-injected seam / perms) **when** open runs **then** no error surfaces to the verb; queue served FROM old layout best-effort. | Red via missing test. | Fail-open degradation test PASS. |
| AC-TOSQ-009 | REQ-TOSQ-005, REQ-TOSQ-008 | **Given** 8 concurrent mutator processes/goroutines each adding M cards over one seeded root **when** stress suite runs **then** total adds = observed distinct sequential unique ids, zero lost updates (final count == total attempts), zero id collisions. | Stress suite absent → red. | `go test -race ./internal/kanban -run TestConcurrencyStress` PASS with counts logged to evidence path. |
| AC-TOSQ-010 | REQ-TOSQ-007, REQ-TOSQ-010 | **Given** the existing CLI suites and web-seam/convention guards **when** affected-package suites run UNMODIFIED where behavioral (golden outputs, flags, exit codes; internal/web seam-import assertion; bare-join convention walk) **then** all green; a frozen verb-surface comparison table (verb × flags × exit codes) generated pre/post swap shows zero deltas. | Existing suites pass TODAY; they go red the moment a behavioral change sneaks in — the guard direction is inverted here: their continued green IS the criterion, plus a new diff-table test asserting emptiness. | `go test ./internal/cli -run 'TestTodo' ./internal/web ...` PASS + surface-diff test asserts zero-delta. |
| AC-TOSQ-011 | REQ-TOSQ-016, SC-3 | **Given** a migrated-and-mutated scratch root (≥1 add, ≥1 state change post-cutover) **when** `moai todo export-json` runs **then** resulting backlog.json parses through the LEGACY record loader with full field equality vs the live reader, including findings order. | New verb does not exist → command non-zero exit = red. | Round-trip property test + manual rehearsal transcript captured in run notes. |
| AC-TOSQ-012 | REQ-TOSQ-009, C-2 | **Given** a generated ≥500-item db fixture **when** the statusline count benchmark runs **then** median added latency ≤10 ms and p95 ≤25 ms vs pre-change baseline harness (same machine). | Benchmark absent → measurement red (no evidence = fail, not pass-by-silence). | Benchmark output persisted under run-notes evidence with med/p95. |
| AC-TOSQ-013 | REQ-TOSQ-002, C-3 | **When** `GOOS=windows GOARCH=amd64 go vet ./internal/...` and `GOOS=linux GOARCH=amd64 go vet ./internal/...` run at M6 **then** exit 0 (compile evidence only); CI windows job verdict green read back from the push. | Current tree red in the future sense — driver+pragmas don't exist yet; concrete red is trivially that modernc dep is absent. | vet outputs + CI job read-back recorded. |
| AC-TOSQ-014 | C-1 | **When** coverage measured per changed-path packages **then** internal/kanban ≥85%, touched internal/cli paths ≥90%. | Absent new code measures trivially below bar once landed-without-tests — enforced at review of run-phase evidence. | Coverage output lines cited in §E.2. |
| AC-TOSQ-015 | REQ-TOSQ-018, E4 | **When** literal-cleanliness sweep runs **then** zero `state/kanban` occurrences across production Go EXCEPT: intentional old-layout fallback reader(s) (each carrying an inline justification comment naming this SPEC) and historical `.moai/specs/SPEC-KANBAN-*` documents; zero occurrences in `internal/template/templates/**`. Sweep checks BOTH spellings: joined literals and segment joins (`".moai", "state", "kanban"`). | Sweep passes TODAY only because nothing renamed yet — flips meaningful post-M3; enforced as release-blocking then. | grep (rg) output quoted in §E.2 with allowlist enumerated inline. |
| AC-TOSQ-016 | REQ-TOSQ-018, Template-First | **Given** the three template skill docs edited **when** `make build` + template-neutrality workflow checks run locally-equivalent steps **then** embedded bundle carries new path; grep of templates shows `state/todo`; neutrality guard classes stay clean. | Red after edit-before-build (bundle staleness verified by strings-check of built binary). | make build exit 0 + grep clean. |
| AC-TOSQ-017 | REQ-TOSQ-001, REQ-TOSQ-004 | **Given** a migrated scratch root **when** the store unit tests assert physical artifacts and ordering fidelity **then** the database file opens containing the meta rows (`schema_version`, `last_seq`) plus items/findings tables per design.md §2 DDL, AND `Load()` returns items in exactly the legacy array insertion order (a queued card positioned AFTER a picked card stays after it; findings likewise ordered by insertion rowid semantics). | Red via missing test. | Schema-presence + order-fidelity subtests PASS within the TestMigrationParity family. |
| AC-TOSQ-018 | REQ-TOSQ-003 | **Given** any freshly opened store connection **when** `PRAGMA journal_mode` and `PRAGMA busy_timeout` are queried **then** they return `wal` and an integer ≥ 5000 respectively, asserted once per connection-establishment path in the wrapper unit suite. | Red via missing test. | Pragma unit test PASS in the kanban suite. |

### D.1 Severity classification

- Release-blocking (any FAIL blocks merge): AC-TOSQ-001…009, 010(guard-green half),
  011…013, 017, 018, 015(post-M3), 016.
- Quality-bar (FAIL requires written debt note, not auto-block): AC-TOSQ-012 within
  2× budget (hard ceiling breach still blocks), AC-TOSQ-014 borderline within −2 pts.
- Advisory: benchmark curve shape observations recorded without gates.

### D.2 Traceability

REQ ↔ AC mapping complete for REQ-TOSQ-001..018 (all eighteen carry at least one
directly-citing AC: see matrix Verify column; B.1 pair AC-TOSQ-017/018 close the
001/003/004 citations). SC-1..SC-4 covered by the AC set + DoD (D.7).

### D.3 Indirect verification accepted

- Windows BEHAVIORAL verdict arrives via CI job read-back (local environment cannot
  execute windows binaries); local obligation stops at cross-vet compile evidence —
  recorded explicitly because local-green ≠ windows-behavioral-pass.
- Ten-lane process-level contention approximated in-process with `-race` plus a
  subprocess pair variant; fleet-scale validation remains CI/dogfood observation.

### D.4 Closure gates

1. All release-blocking ACs PASS with evidence paths resolvable from progress.md §E.2.
2. Sync-audit tier: thorough (migration machinery + storage replace = high-risk class).
3. Rollback documentation reviewed for accuracy against actual artifact names.

### D.5 Edge cases catalog (beyond primary GWTs)

Empty queue first-run · single-item · all-dropped · findings-only-no-items mutation ·
`last_seq` hand-lowered below max-present · SpecID pointing at dropped card · duplicate
finding tuples in legacy data · non-`t<N>` conforming-id aborts · db file present but
zero-byte (treated corrupt → ErrBacklogCorrupt, never deleted) · `-wal` orphaned after
crash · relocation target dir pre-created by another actor mid-flight · statusline read
during active write txn (WAL snapshot) · worktree session reading primary-root queue
(resolver unchanged semantics).

### D.6 Quality gate criteria

Per-project TRUST-5: Tested (C-1) · Readable (godoc parity with surrounding comment
density; comments en per language.yaml) · Unified (gofmt/golangci clean) · Secured
(no SQL string interpolation of user text anywhere — parameterized statements only,
grep-enforced) · Trackable (conventional commits referencing SPEC id per milestone).

### D.7 Definition of Done

Plan+run+sync closed; artifacts consistent; live-operator rehearsal (SC-3) executed on
a scratch root and its transcript linked from §E.2; develop-integration merge window
completed by lead confirmation; t309 marked absorbed by this card's delivery.
