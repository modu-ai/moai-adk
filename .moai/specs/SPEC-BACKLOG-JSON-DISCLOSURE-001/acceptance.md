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
Given the same layout, **whose `backlog.db` has its archive tables present**,
When a `moai todo` **read** command is invoked against it,
Then exactly one **line introduced by this SPEC** is written, naming the SQLite
store as the store that answered and naming `backlog.json` as not authoritative.

Two clauses in that wording are load-bearing, not hedging:

- **"read"** matches REQ-BJD-002 exactly, and the breadth behind it is settled:
  the operator decided the read surface **in full** on 2026-09-02 (`spec.md`
  REQ-BJD-002, `plan.md` §D.1). This criterion is satisfied by any verb in that
  surface; the concrete verb list is enumerated from the code at run-phase, so
  this criterion deliberately does not name one.
- **"introduced by this SPEC"** and the archive-tables Given together fix what is
  counted. `internal/cli/todo_history.go:99-107` already writes a REQ-TAQ-013
  store-identity line to stderr on two branches, so a fixture whose archive
  tables are missing would produce two lines and make a bare "exactly one"
  ambiguous. The Given removes the collision; the qualifier removes the ambiguity
  even if it recurs on a surface not yet identified.

**AC-BJD-003** — stdout is unpolluted (REQ-BJD-004).
Given the same layout and invocation,
When stdout and stderr are captured separately,
Then the disclosure text appears on stderr and **zero** bytes of it appear on
stdout, and stdout is byte-identical to the same command's stdout against a
layout with no `backlog.json`.

**AC-BJD-004** — no disclosure when there is nothing to disclose (REQ-BJD-003).
Given a queue directory containing `backlog.db` and no `backlog.json`, **whose
`backlog.db` has its archive tables present**,
When the same command is invoked,
Then **no line introduced by this SPEC** is emitted on either stream.

Both qualifiers are inherited from AC-BJD-002 deliberately. This criterion
restates its Given rather than referencing it, so the archive-tables condition
does not carry across on its own — and without it the existing REQ-TAQ-013 line
at `internal/cli/todo_history.go:99-107` fires on an archive-less fixture,
reintroducing on the negative side exactly the ambiguity AC-BJD-002 removed on
the positive side.

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
When the two commands below are run from the repository root,
Then both produce their stated result.

```sh
# 1. Exactly one backlog-store inspector exists, and it is the existing one.
grep -c 'func Inspect.*Backlog.*\(Store\|Vouch\)' internal/kanban/*.go
#    → the summed count across all files is 1, and the sole non-zero file is
#      internal/kanban/backlog_archive_vouch.go

# 2. The store-name constants are defined in that one file only.
grep -rn 'BacklogStore\(SQLite\|LegacyJSON\|None\) *=' internal/kanban/ | grep -v _test.go
#    → every hit is in internal/kanban/backlog_archive_vouch.go
```

And: the new fact is a **field on `BacklogArchiveVouch`**, verified by
`grep -n 'type BacklogArchiveVouch struct' -A 8 internal/kanban/backlog_archive_vouch.go`
showing it inside that struct — not a second return type and not a second
inspector.

A criterion that only said "is searched for store-identity reporting" would have
delegated a universal negative to human judgement, which the preamble above
forbids. These commands are what make it binary.

### D.2 The foreman queue watch (REQ-BJD-007..010)

**AC-BJD-008** — the R1 red, made into a regression. **[required]**
Given an isolated fixture queue whose authoritative store is `backlog.db` and
which has **no** `backlog.json` (the migrated-project layout of
`.moai/reports/t395/r1-repro.md`),
When the queue-watch script is taken **verbatim from
`.claude/skills/moai-kanban-foreman/SKILL.md`**, rebound only so that **the queue
directory its watch target or targets resolve against** is the fixture's, armed,
and the queue is mutated inside the watch window,
Then a change event is observed within the window.

The rebinding is stated against the **directory**, not against a single `f=`
variable, because AC-BJD-010 explicitly permits a repair whose watch covers more
than one path (`backlog.db` plus `backlog.db-wal`). A rebinding instruction
naming one variable would become undefined the moment that repair is chosen, and
the two criteria would then invalidate each other.

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

The premise is real, not hypothetical: `internal/kanban/backlog_sqlite.go:229`
sets `_pragma journal_mode(WAL)` on every queue connection.

Given a fixture queue in which a mutation has been **committed and not yet
checkpointed**, established by a technique the run-phase names, and **evidenced
before the watch is armed** by both of:

- `backlog.db-wal` exists and its size is greater than zero, **and**
- `cksum backlog.db` is byte-identical to its value captured before the mutation,

When the watch is armed and the window elapses,
Then a change event is observed within the window.

**Establishing the Given is part of the criterion, not a preliminary.** If the
two evidence values above are not both observed, the run has not built the state
this criterion is about, and the result is recorded as a **Gap** — never as a
pass. This closes a vacuity path distinct from the window-widening one below:
`.moai/reports/t395/r1-repro.md` recorded `backlog.db`'s cksum moving
immediately on a mutation (`1125950968` → `1188524958`), so a naive run that just
calls `moai todo add` and then watches may measure an already-checkpointed state
and pass **without the Given ever holding**.

Note on technique, for the run-phase to resolve by measurement rather than
assumption: `wal_autocheckpoint` is a **per-connection** setting, and no code in
`internal/kanban` sets it, so setting it from a side connection does not restrain
the connection `moai todo` opens. Two candidate techniques, neither yet measured
here: hold an open read snapshot on the database from a second connection for the
duration (a checkpoint cannot reclaim WAL past a live reader's mark), or perform
the mutation on a directly-opened connection that sets `wal_autocheckpoint=0` and
stays open. Whichever is used, the two evidence values above are what decide
whether it worked — the technique is not credited on its reasoning.

The two ways this criterion may **not** be satisfied: by widening the observation
window until it happens to pass, or by measuring an already-checkpointed state.
If the repaired watch genuinely cannot see a WAL-deferred write, the criterion is
met by a watch target that covers the deferral, and the measured checkpoint
behaviour is recorded either way.

**AC-BJD-011** — both copies of the script agree.
Given the post-repair tree,
When the queue-watch block is extracted from `.claude/skills/moai-kanban-foreman/SKILL.md`
and from `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md`,
Then the two watch targets are identical, and neither is `backlog.json`.

### D.3 The reader instructions (REQ-BJD-011..013)

**AC-BJD-012** — the false assertions are gone.
Given the post-repair tree,
When both commands below are run from the repository root,
Then each produces its stated result.

```sh
# 1. The primary-checkout assertions (SKILL.md:170, todo.md:17, foreman :95).
grep -rn 'state/todo/backlog\.json' .claude/skills/
#    → zero matches

# 2. The home-fallback assertion (todo.md:21). Pattern 1 cannot see this one.
grep -rn 'moai/todo/<project-key>/backlog\.json' .claude/skills/
#    → zero matches
```

Both are required, for the same reason AC-BJD-015 requires two: the four defect
sites do not share a path shape, and a single pattern silently omits the fourth.

**Stated result, precisely**: zero matches, **or** every surviving match sits on a
line that also contains `backlog.db` or `export`. The alternative exists because
a correct repair may legitimately keep a labelled mention — "a `backlog.json`
here is an export, not the queue" is true and useful — and a bare zero-match rule
would forbid the very sentence that prevents the next stale read. Mechanically:

```sh
grep -rn -e 'state/todo/backlog\.json' -e 'moai/todo/<project-key>/backlog\.json' \
  .claude/skills/ | grep -v -e 'backlog\.db' -e 'export'
#    → zero matches
```

What must not survive is the **assertion** that the JSON is where queue state
lives; a mention that names it as an export or a legacy layout is not that.

**AC-BJD-013** — the corrected assertions name the database.
Given the post-repair tree,
When the state-location sentence of each of the three **R2 assertion** sites is
read — `.claude/skills/moai/SKILL.md:170`,
`.claude/skills/moai/workflows/todo.md:17`, `.claude/skills/moai/workflows/todo.md:21`
(the fourth site, foreman `SKILL.md:95`, is a watch target rather than a
state-location sentence and is covered by AC-BJD-008 and AC-BJD-011) —
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

This is the only BLOCKING completeness check over the template mirror, so it
**enumerates the sites it must cover** rather than trusting one pattern to find
them. The four defect sites do not share a path shape: three are
`state/todo/backlog.json`, and the fourth is the home fallback
`~/.moai/todo/<project-key>/backlog.json`, which **does not match** the first
pattern. A single-regex form of this criterion passes while the template's
`workflows/todo.md:21` stays unrepaired — measured, not supposed.

Given the post-repair tree,
When both commands below are run from the repository root,
Then each produces its stated result.

```sh
# 1. The three same-shape sites are gone; only the export control survives.
grep -rn 'state/todo/backlog\.json' internal/template/templates/
#    → exactly one match, and it is
#      internal/template/templates/.moai/docs/todo-queue-storage.md:55
#      (the export-json line — correct, out of scope, not edited)

# 2. The home-fallback site is gone. Pattern 1 cannot see this one.
grep -rn 'moai/todo/<project-key>/backlog\.json' internal/template/templates/
#    → zero matches
```

Both commands are required. A run that reports only the first has checked three
of the four sites and has not established completeness.

**AC-BJD-016** — the binary carries the corrected templates.
Given the post-repair tree,
When `make build` has been run and the resulting binary's embedded template FS is
queried for the **three** mirrored files —
`.claude/skills/moai/SKILL.md`, `.claude/skills/moai/workflows/todo.md`,
`.claude/skills/moai-kanban-foreman/SKILL.md` —
Then each embedded copy matches its `internal/template/templates/` source
byte-for-byte.

Three files, four sites: `workflows/todo.md` carries two of the four (`:17`
primary-checkout, `:21` home-fallback). The unit is named explicitly here because
a check written against "the four mirrored files" names a set that does not
exist.

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
