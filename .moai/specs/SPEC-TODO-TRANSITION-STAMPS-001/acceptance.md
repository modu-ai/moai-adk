# Acceptance — SPEC-TODO-TRANSITION-STAMPS-001

Every criterion is mechanically checkable against a `t.TempDir()` queue db
(schema assertions via `pragma_table_info`, behavior via `todo` verb runs,
output via tab-field greps). Binary-testable wording throughout; no coverage
of subjective qualities.

## §D AC Matrix

| AC | Verifies | Kind | Mechanical check |
|------|-----------|------|------------------|
| AC-TST-001 | REQ-TST-001..003 | schema | `pragma_table_info` on fresh + upgraded db |
| AC-TST-002 | REQ-TST-004/005 | behavior | pick then unpick, read column |
| AC-TST-003 | REQ-TST-006 | behavior | drop then undrop, read column |
| AC-TST-004 | REQ-TST-007 | behavior | done, read archived row |
| AC-TST-005 | REQ-TST-008 | behavior | done --require-landed, read archived row |
| AC-TST-006 | REQ-TST-009 | behavior | done without flag, read archived row |
| AC-TST-007 | REQ-TST-010 | invariant | no bare-verdict write without ref |
| AC-TST-008 | REQ-TST-011 | output | history listing/lookup field greps |
| AC-TST-009 | REQ-TST-012 | output | history live lookup + list --json |
| AC-TST-010 | REQ-TST-013 | invariant | stored verdict carries ref |
| AC-TST-011 | §5 freeze | regression | schema-freeze test updated and green |
| AC-TST-012 | JSON contract | regression | existing fields byte-identical |

## §D.1 Scenarios

**AC-TST-001** — Additive columns exist and converge.
Given a queue db created fresh by the new binary, and a second db created by
the pre-change DDL then opened by the new binary (the upgrade path),
When `pragma_table_info` is read for `items` and `archived_items` on both,
Then both dbs declare `picked_at` and `dropped_at` on `items` and
`archived_at` on `archived_items` as nullable TEXT, both column sets are
identical, and no pre-existing column was dropped, retyped, or reordered
(diff of full column lists, old columns as a prefix in original order).

**AC-TST-002** — picked_at stamps on entry, clears on exit.
Given a queue with one queued card,
When `todo add --pick` (or `todo next`) picks it, Then its `picked_at` is a
non-NULL TEXT timestamp and `state` is `picked`;
When `todo unpick` returns it to queued, Then `picked_at` is NULL;
When it is picked a second time, Then `picked_at` holds the SECOND
transition's time, not the first.

**AC-TST-003** — dropped_at stamps and clears symmetrically.
Given a queued card, When `todo drop` runs, Then `dropped_at` is non-NULL;
When `todo undrop` restores it, Then `dropped_at` is NULL.

**AC-TST-004** — Archive preserves stamps and adds archived_at.
Given a card picked at T1 (picked_at = T1 observed in the live row),
When `todo done <id>` archives it,
Then the archived row carries picked_at = T1 unchanged and a non-NULL
`archived_at`, and the live table no longer holds the row.

**AC-TST-005** — The require-landed verdict is persisted, without a stored SHA.
Given a queue seeded so that the landing query against a controlled ref
answers `landed` (a commit whose subject attributes the card, per the
axis-F predicate),
When `todo done <id> --require-landed` runs,
Then the archived row's verdict record holds verdict `landed`, the answering
ref, and the verdict time — three fields readable from the archived row
without re-running any git command — and the record carries NO SHA field
(query-derived SHAs are outside the evidence store's write authority; the
delivering SHA is re-derived at re-adjudication, AC-TST-010).

**AC-TST-006** — No fabricated verdict without the flag.
Given a card carrying NO recorded landing evidence,
When `todo done <id>` runs without `--require-landed`,
Then the archived row's landing record is NULL (rendered `-` by history),
even though the printed stdout line reads `landing=unknown`.

**AC-TST-007** — No verdict persisted without its ref.
Given any archived row whose landing record holds a query-produced verdict,
When the record is read, Then the answering ref is non-empty; the store
offers no write path that persists a query verdict with an empty ref
(exercised by attempting the write through the store API and asserting the
refusal).

**AC-TST-008** — History exposes the archive time axis.
Given the archive of AC-TST-004/005,
When `todo history <id>` (lookup) and `todo history --limit N` (listing) run,
Then each archived row renders `picked_at`, `dropped_at`, `archived_at`, and
the landing record as appended tab-separated fields, `-` marking each absent
value, existing column order and content unchanged (grep the pre-existing
fields byte-for-byte), and a card with no stamps renders `-` in all three new
fields.

**AC-TST-009** — Live WIP age is computable from the surfaces.
Given a picked live card, When `todo history <id>` runs, Then the live row
renders `picked_at`; When `todo list --json` runs, Then the card's JSON
object carries a non-empty `picked_at` key, and a never-picked card's JSON
object omits the key entirely (`omitempty`).

**AC-TST-010** — Stored verdict cites the axis-F ref; the SHA is re-derivable.
Given AC-TST-005's archived row, When the verdict record is read, Then its
ref equals the ref the printed line named (`ref=<ref>` on stdout), and
re-running the axis-F predicate against that recorded ref reproduces both
the stored verdict and the delivering SHA for the seeded commit (consistency
of storage with the predicate, not a replacement of it — the record stores
the answer's coordinates, the re-run supplies the SHA).

**AC-TST-011** — The freeze test records the decision.
Given the updated `backlog_schema_freeze_test.go`, When
`go test ./internal/kanban/ -run SchemaFreeze` runs, Then it passes with the
three new columns recorded in its expectation set, and the diff shows the
freeze test's change is exactly the column additions.

**AC-TST-012** — Existing JSON disclosure is byte-identical.
Given a fixture queue db (a checked-in `testdata` seed script building a
queue holding a queued, a picked, a dropped, and an archived card) and a
golden `list --json` snapshot captured from the PRE-change struct shape and
committed as a `testdata` file BEFORE the implementation lands (the ordering
rule of verification-claim-integrity §2.3: baseline commit precedes the
change commit),
When the post-change binary renders `todo list --json` against the fixture,
Then every pre-existing JSON key and value is byte-identical to the golden
file and the only diff is the three new `omitempty` keys on the objects that
have them.

## §D.2 Severity

- Critical (blocks close): AC-TST-001, 004, 005, 006, 011, 012.
- Major: AC-TST-002, 003, 007, 008, 010.
- Minor: AC-TST-009 (surface ergonomics).

## §D.3 Traceability

Every REQ-TST-001..013 maps to at least one AC above; REQ-TST-013 is verified
by AC-TST-007 + AC-TST-010 jointly. No AC verifies a requirement absent from
spec.md.

## §D.4 Edge cases

- Drop → done directly: the archived row carries `dropped_at` with
  `picked_at` NULL — drop is queued-only (`todo_drop.go:79-80`), so the two
  stamps never coexist on one row (AC-TST-004 covers the general
  preservation).
- Re-picked card: second stamp overwrites the first (AC-TST-002 third step).
- `--require-landed` refusing (not-landed): the refusal path archives
  nothing; only the answering path persists a verdict (AC-TST-005's seeded
  control is the landed arm; the refused arm asserts the live row survives
  with stamps intact).
- Empty archive (`history` on a fresh queue): unchanged empty-archive line.

## §D.5 Quality gates

- `go vet` + `golangci-lint` on affected packages.
- `go test -cover ./internal/kanban/... ./internal/cli/...` ≥ 85%.
- Fresh-vs-upgraded convergence asserted in-test (AC-TST-001), not by hand.

## §D.6 Definition of Done

All Critical and Major ACs green with command output recorded in progress.md
§E.2; Minor ACs green or explicitly deferred with the lane's acceptance;
schema-freeze test updated as a reviewed decision; template mirror of any
workflow-doc change regenerated (`make build`).

## §D.7 Indirect verification

AC-TST-012's byte-identity check doubles as the characterization test for
pre-existing consumers (scripts parsing `history` and `list --json`): the
regression suite compares against a golden captured from the pre-change
behavior in the same test run.
