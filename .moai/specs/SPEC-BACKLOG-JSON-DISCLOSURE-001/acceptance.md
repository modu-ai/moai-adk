# Acceptance — SPEC-BACKLOG-JSON-DISCLOSURE-001

Every criterion below is binary and machine-checkable. Where a criterion asserts
a repair, it names the condition under which it must have been **red before the
repair** — a criterion that would pass against the shipped tree asserts nothing.

Fixture convention: acceptance work uses an isolated queue root
(`t.TempDir()`, or a scratch project resolved through the home fallback). No
criterion may read, mutate, or measure the live primary-checkout queue at
`/Users/goos/MoAI/moai-adk-go/.moai/state/todo/`.

## §D Acceptance criteria

### D.1 Disclosure (REQ-BJD-001..006)

**AC-BJD-001** — the State D fact is reported.
Given a queue directory containing both `backlog.db` and `backlog.json`,
When the store-identity inspector is called on that queue path,
Then it reports the SQLite store as the answering store **and** reports the
presence of the non-authoritative `backlog.json` as a distinct field.

**AC-BJD-002** — the disclosure reaches the operator.
Given the same layout,
When a `moai todo` read command is invoked against it,
Then exactly one disclosure line is written naming the SQLite store as the store
that answered and naming `backlog.json` as not authoritative.

**AC-BJD-003** — stdout is unpolluted (REQ-BJD-004).
Given the same layout and invocation,
When stdout and stderr are captured separately,
Then the disclosure text appears on stderr and **zero** bytes of it appear on
stdout, and stdout is byte-identical to the same command's stdout against a
layout with no `backlog.json`.

**AC-BJD-004** — no disclosure when there is nothing to disclose (REQ-BJD-003).
Given a queue directory containing `backlog.db` and no `backlog.json`,
When the same command is invoked,
Then no disclosure line is emitted on either stream.

**AC-BJD-005** — the file is untouched (REQ-BJD-005).
Given a queue directory containing both artifacts, with the `backlog.json`
content and mtime captured beforehand,
When the inspector and the disclosing command have both run,
Then `backlog.json` still exists, its bytes are unchanged (sha256 equal), and its
mtime is unchanged.

**AC-BJD-006** — the disclosure path is read-only (REQ-BJD-005).
Given a queue directory whose `backlog.db` is made read-only for the duration,
When the inspector is called,
Then it returns its report without error, and no queue lock file is created or
modified by the inspector's own execution.

**AC-BJD-007** — the mechanism was extended, not duplicated (REQ-BJD-006).
Given the post-repair tree,
When `internal/kanban/` is searched for store-identity reporting,
Then `InspectBacklogArchiveVouch` in `backlog_archive_vouch.go` is the only
function reporting which backlog store answers reads, and the new fact is a field
on its existing return type rather than a second inspector.

### D.2 The foreman queue watch (REQ-BJD-007..010)

**AC-BJD-008** — the R1 red, made into a regression. **[required]**
Given an isolated fixture queue whose authoritative store is `backlog.db` and
which has **no** `backlog.json` (the migrated-project layout of
`.moai/reports/t395/r1-repro.md`),
When the queue-watch script is taken **verbatim from
`.claude/skills/moai-kanban-foreman/SKILL.md`** (only its queue path rebound to
the fixture), armed, and the queue is mutated inside the watch window,
Then a change event is observed within the window.

Falsifiability condition, which the check MUST itself demonstrate: reverting the
script's watch target to `.moai/state/todo/backlog.json` makes this criterion
**fail**. A check that passes against both targets does not satisfy AC-BJD-008.

**AC-BJD-009** — Case B: a stale `backlog.json` present.
Given the same fixture, plus a `backlog.json` written beside the database and
never modified again,
When the queue is mutated inside the watch window,
Then a change event is observed.

Recorded provenance: Case B was **reasoned but not observed** in
`.moai/reports/t395/r1-repro.md` (its Gaps section) — only Case A (file absent)
was run. This criterion is what closes that gap, and it is asserted here
deliberately rather than inherited.

**AC-BJD-010** — WAL deferral does not hide a mutation (REQ-BJD-010).
Given a fixture queue mutated through the normal `moai todo` write path,
When the mutation is committed but SQLite has not yet checkpointed into the main
database file,
Then the watch still emits a change event within the window.

If the repaired watch cannot satisfy this against a WAL-deferred write, the
run-phase records the measured checkpoint behaviour and the criterion is met by a
watch target that covers the deferral — not by widening the window until it
happens to pass.

**AC-BJD-011** — both copies of the script agree.
Given the post-repair tree,
When the queue-watch block is extracted from `.claude/skills/moai-kanban-foreman/SKILL.md`
and from `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md`,
Then the two watch targets are identical, and neither is `backlog.json`.

### D.3 The reader instructions (REQ-BJD-011..013)

**AC-BJD-012** — the false assertions are gone.
Given the post-repair tree,
When `.claude/skills/moai/SKILL.md` and `.claude/skills/moai/workflows/todo.md`
are searched for `state/todo/backlog.json` and for
`~/.moai/todo/<project-key>/backlog.json`,
Then zero matches remain that assert either path is where queue state lives.

A residual mention is permitted only where it is explicitly labelled as an export
artifact or a legacy pre-migration layout; the check distinguishes the two by
requiring any surviving occurrence to sit on a line that also names the database
or the word `export`.

**AC-BJD-013** — the corrected assertions name the database.
Given the post-repair tree,
When the state-location sentence of each of the three sites in `spec.md` §A.2 is
read,
Then each names `.moai/state/todo/backlog.db` (or the home-fallback form ending
in `backlog.db`).

**AC-BJD-014** — no new contradiction with the correct document (REQ-BJD-013).
Given the post-repair tree,
When the corrected sentences are compared against `.moai/docs/todo-queue-storage.md`,
Then no corrected sentence contradicts it, and
`.moai/docs/todo-queue-storage.md` itself is unmodified by this SPEC (`git diff
--exit-code` clean for that path across the SPEC's commits).

### D.4 The distributed copy (REQ-BJD-014)

**AC-BJD-015** — the mirror is complete.
Given the post-repair tree,
When `grep -rn 'state/todo/backlog\.json' internal/template/templates/` is run,
Then the only surviving match is the `export-json` line of
`internal/template/templates/.moai/docs/todo-queue-storage.md`, which is correct
and out of scope.

**AC-BJD-016** — the binary carries the corrected templates.
Given the post-repair tree,
When `make build` has been run and the resulting binary's embedded template FS is
queried for the four mirrored files,
Then each embedded copy matches its `internal/template/templates/` source
byte-for-byte.

## §D.5 Severity

| Criterion | Severity | Rationale |
|---|---|---|
| AC-BJD-008 | BLOCKING | The card's only observed runtime defect. Without it the repair is unverified. |
| AC-BJD-002, 003, 005 | BLOCKING | The disclosure is the durable defence (§A.3); silent, stdout-polluting, or file-mutating variants defeat it. |
| AC-BJD-012, 015 | BLOCKING | The false instruction is what produced the stale read, locally and in the distributed copy. |
| AC-BJD-009, 010 | SHOULD-FIX | Closes a recorded Gap and a recorded Residual-risk respectively; a measured negative result is an acceptable outcome if recorded. |
| AC-BJD-001, 004, 006, 007, 011, 013, 014, 016 | SHOULD-FIX | Shape, parity, and non-regression guards. |

## §D.6 Traceability

| REQ | Criteria |
|---|---|
| REQ-BJD-001 | AC-BJD-001 |
| REQ-BJD-002 | AC-BJD-002 |
| REQ-BJD-003 | AC-BJD-004 |
| REQ-BJD-004 | AC-BJD-003 |
| REQ-BJD-005 | AC-BJD-005, AC-BJD-006 |
| REQ-BJD-006 | AC-BJD-007 |
| REQ-BJD-007 | AC-BJD-008, AC-BJD-011 |
| REQ-BJD-008 | AC-BJD-008, AC-BJD-009 |
| REQ-BJD-009 | AC-BJD-008 (falsifiability condition), AC-BJD-009 |
| REQ-BJD-010 | AC-BJD-010 |
| REQ-BJD-011 | AC-BJD-013 |
| REQ-BJD-012 | AC-BJD-012 |
| REQ-BJD-013 | AC-BJD-014 |
| REQ-BJD-014 | AC-BJD-011, AC-BJD-015, AC-BJD-016 |

## §D.7 Definition of Done

- Every BLOCKING criterion passes, with the command and its verbatim output cited.
- AC-BJD-008's falsifiability condition is demonstrated, not asserted — the
  run-phase evidence shows the check red against the shipped watch target.
- Package-scoped tests pass for every package touched
  (`go test ./internal/kanban/... ./internal/cli/...`); the full-suite verdict is
  CI's, not the lane's.
- `go vet` clean on the touched packages; `make build` run and committed.
- Any criterion that could not be met is recorded in `progress.md` §E.2 as a Gap
  with the measurement that established it — never dropped.
