# Plan — SPEC-TODO-TRANSITION-STAMPS-001

## §A Context

Card t1310 (Class C, Tier M, P3): the queue has no time axis beyond
`added_at`, and `done --require-landed` prints its verdict then discards it.
This plan turns the SPEC's four decisions into additive-schema milestones.
Ground truth measured 2026-09-29: live db schema (lane measurement, recorded
in spec.md §2), and this tree's code at `9cc3fdc4d`.

Ordering principle: the data-model decision (column set + row structs) comes
first because everything downstream compiles against it and is the hardest
decision to reverse once archives accumulate; the done-verdict persistence is
the user-facing behavior most likely to need wording iteration; mechanical
test/coverage work sits at the bottom.

## §B Known Issues

- `backlog_schema_freeze_test.go` asserts the frozen schema and WILL fail on
  the first new column — that is its job. The update is a reviewed decision
  (spec.md §5), not a test to silence.
- JSON disclosure tests assert byte-identity of existing fields; the new
  `omitempty` fields must not appear on objects that predate them.
- `history` output is tab-separated and consumed by scripts; adding fields
  must append columns, never reorder existing ones (REQ-TST-011).
- Sibling card t1308 may touch the same DDL region for the state-CHECK
  question. Merge-order note: whichever lands second absorbs and re-runs the
  freeze test; no semantic conflict is expected (different columns).

## §C Pre-flight

- [ ] Re-verify the live-db schema facts in spec.md §2 before M1 (they are
      dated; the db moves).
- [ ] Confirm `landingCarryingTables` / `ensureLandingColumn` shape is
      unchanged from the reading this plan was written against.
- [ ] Confirm t1308's status: if its CHECK rebuild landed, read the new DDL
      region before adding columns beside it.

## §D Constraints

- No destructive migration; idempotence via `pragma_table_info` (spec.md §5).
- The five frozen fields keep names, types, JSON tags; stamps are additive
  `omitempty` pointers following the `Landing`/`CardUUID` precedent.
- English code/comments/godoc; `t.TempDir()` test isolation; affected-package
  tests only locally (`go test -timeout 30m ./internal/kanban/...
  ./internal/cli/...`), full-suite verdict is CI's.

## §E Self-Verification

The run phase records evidence here and in progress.md §E.2/§E.3 per the
lifecycle contract. Minimum: affected-package test output, a fresh-db vs
upgraded-db schema convergence check, and a byte-identity check of existing
JSON disclosure fields.

## §F Milestones

### M1 — Schema and row structs (data model; highest reversal cost)

- `internal/kanban/backlog_sqlite.go`: extend the additive migration to add
  `picked_at TEXT`, `dropped_at TEXT` (items) and `archived_at TEXT`
  (archived_items) — generalize the existing `ensureLandingColumn`
  column-metadata-gated `ADD COLUMN` pattern to the new column set; keep the
  compile-time-constant table/column interpolation discipline.
- `internal/kanban/backlog_store.go`: add `PickedAt *string`,
  `DroppedAt *string` to `BacklogItem` and `ArchivedAt *string` to the
  archive entry, all `json:"...,omitempty"`, NULL-mapped like `Landing`.
- Row scan/persist paths in `internal/kanban/backlog_migrate.go` (the
  `LandingEvidenceValue` write funnel) extended for the three columns.
- Update `backlog_schema_freeze_test.go` as the recorded decision.
- Fresh-db and upgraded-db converge on identical `pragma_table_info` output.

### M2 — Transition stamping (pick / unpick / drop / undrop / archive)

- Stamp on entry into picked (`todo add --pick`, `todo next`) and drop;
  clear on unpick and undrop (spec.md REQ-TST-004..006).
- `ArchiveCard` preserves the stamps into the archive entry and stamps
  `archived_at` at archive time (REQ-TST-007).

### M3 — Done-time verdict persistence

- In the `done` verb (`internal/cli/todo.go`), persist the
  `--require-landed` answer — verdict, answering ref, verdict time — into the
  archived row's landing record through the existing evidence machinery;
  without the flag, persist only pre-existing recorded evidence or NULL
  (REQ-TST-008..010). The printed verdict line stays byte-identical.

### M4 — History and list exposure

- `internal/cli/todo_history.go`: append `picked_at` / `dropped_at` /
  `archived_at` columns to lookup and listing rows, `-` for absent
  (REQ-TST-011..012).
- `todo list --json` carries the stamps as `omitempty`.

### M5 — Tests, coverage, docs (mechanical; lowest reversal cost)

- Table-driven tests per AC in `acceptance.md`; fresh-vs-upgraded
  convergence test; JSON byte-identity regression test.
- Affected-package coverage ≥ 85% (`go test -cover` on internal/kanban,
  internal/cli).
- Documentation touch where the verbs' output is documented
  (`.claude/skills/moai/workflows/todo.md` + its template mirror, per the
  Template-First rule — template first, then `make build`).

## §G Anti-Patterns

- Do NOT catch "duplicate column name" as the idempotence mechanism.
- Do NOT write `""` or `{}` into any new column — typed NULL only.
- Do NOT reorder or reformat existing `history` output fields.
- Do NOT derive a landing verdict anywhere except the query path; storage is
  not a second attribution oracle (REQ-TST-013).

## §H Cross-References

- SPEC-TODO-LANDING-EVIDENCE-001 — the additive-column and evidence-write
  precedent this plan generalizes.
- SPEC-TODO-LANDING-ATTRIBUTION-001 — the axis-F predicate; verdict records
  cite its ref so it can be re-run.
- Card t1308 — sibling schema-convention review (state CHECK); coordinate at
  merge, no scope overlap.
