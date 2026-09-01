# Acceptance — SPEC-TODO-ARCHIVE-QUERY-001

Every criterion below is binary and names the command that decides it. Commands
run from the repository root of the tree under test.

Fixture convention: `FIXTURE` is a `t.TempDir()` project root holding a queue
seeded through the public CLI — cards added with `moai todo add`, closed with
`moai todo done`. A test writes storage rows by hand only where the criterion
must construct a shape the CLI **cannot** produce, and every such criterion is
named here. There are exactly two, both of which reconstruct the artefact of an
*older* binary that the current CLI no longer creates:

- **AC-TAQ-004** removes a row from `items` without lowering `meta.last_seq` —
  the shape a pre-archive `done` left behind. The current CLI archives instead
  of deleting, so it cannot produce this queue.
- **AC-TAQ-013** drops `archived_items` and `archived_findings` — the shape a
  pre-archive database has. The DDL recreates both on every open
  (`internal/kanban/backlog_sqlite.go:92-94`), so the CLI cannot produce a
  database without them either.

The exception is deliberately narrow, and the reason for the narrowness is that
a hand-written row is a claim about what the storage layer would have written.
Where the CLI can produce the shape, using the CLI is what keeps the fixture
honest; where it cannot, the surgery is named and its target shape is cited to
the code that once produced it.

---

## §D AC Matrix

| AC | Requirement | Decided by |
|---|---|---|
| AC-TAQ-001 | REQ-TAQ-001 | `TestTodoHistoryReportsLiveCard` |
| AC-TAQ-002 | REQ-TAQ-002 | `TestTodoHistoryReportsArchivedCard` |
| AC-TAQ-003 | REQ-TAQ-003 | `TestTodoHistoryReportsAbsentCard` |
| AC-TAQ-004 | REQ-TAQ-004 | `TestTodoHistoryDisclosesPreArchiveQueue` |
| AC-TAQ-005 | REQ-TAQ-005 | `TestTodoHistoryNormalizesBareOrdinal` |
| AC-TAQ-006 | REQ-TAQ-006 | `TestTodoHistoryListsNewestFirst` |
| AC-TAQ-007 | REQ-TAQ-007 | `TestTodoHistoryLimitBound` |
| AC-TAQ-008 | REQ-TAQ-008 | `TestTodoHistoryStatesWithheldCount` |
| AC-TAQ-009 | REQ-TAQ-009 | `TestTodoHistoryEmptyArchiveIsExplicit` |
| AC-TAQ-010 | REQ-TAQ-010 | `TestTodoHistoryLeavesStorageByteIdentical` |
| AC-TAQ-011 | REQ-TAQ-011 | `TestLiveReadersUnchangedByHistoryVerb` |
| AC-TAQ-012 | REQ-TAQ-012 | `TestTodoHistoryAddsNoSchemaChange` |
| AC-TAQ-013 | REQ-TAQ-013 | `TestTodoHistoryDegradesWithoutArchiveTables` |
| AC-TAQ-014 | REQ-TAQ-014 | `TestTodoHistoryNeverPrompts` |
| AC-TAQ-015 | REQ-TAQ-015 | `TestTodoSkillDocumentsHistoryVerb` + neutrality grep |

---

## §D.1 Lookup

### AC-TAQ-001 — a live card reports `live` and its state

**Given** a `FIXTURE` queue holding card `t1` in state `queued`,
**when** `moai todo history t1` runs,
**then** stdout is exactly one line whose fields are `t1`, `live`, `queued`, and
the card's text, and the exit code is 0.

The test repeats the same assertion for a `picked` card and for a `dropped` card,
so the three live states are proven distinct rather than collapsed to `live`.

```
go test ./internal/cli/ -run TestTodoHistoryReportsLiveCard -count=1
```

### AC-TAQ-002 — an archived card reports `archived`

**Given** a `FIXTURE` queue where `t1` was added and then closed with
`moai todo done t1`,
**when** `moai todo history t1` runs,
**then** stdout is exactly one line whose fields are `t1`, `archived`, the state
`t1` held at archive time, and its text; exit code 0.

```
go test ./internal/cli/ -run TestTodoHistoryReportsArchivedCard -count=1
```

### AC-TAQ-003 — an unknown id reports `absent` and exits 0

**Given** a `FIXTURE` queue holding at least one card and no card `t9999`,
**when** `moai todo history t9999` runs,
**then** stdout names `t9999` and `absent`, and the exit code is 0.

The test additionally asserts that this output is NOT byte-equal to what
`moai todo why t9999` prints, so the conflation recorded at
`internal/cli/todo_why.go:34-37` is proven closed rather than reproduced.

```
go test ./internal/cli/ -run TestTodoHistoryReportsAbsentCard -count=1
```

### AC-TAQ-004 — `absent` is qualified for an id that was issued, and only for one

**Given** a `FIXTURE` queue whose `meta.last_seq` is 5 and which holds `t3` in
neither its live rows nor its archive (seeded by adding five cards through the
CLI and removing `t3`'s row from `items` without lowering `last_seq` — the shape
a pre-archive `done` leaves),
**when** `moai todo history t3` runs,
**then** stderr states that `t3` lies at or below the queue's issued-id mark and
so may have been issued and destroyed, stdout still carries only the single
`absent` line, and the exit code is 0.

**And given** the same queue after one `moai todo done` on a different card,
**when** the same `history t3` lookup runs, **then** that stderr note is **still
present**. This clause is the regression guard for the defect the first draft
carried: a condition written as *"the archive is empty"* goes silent here while
`t3` stays destroyed, so an implementation that re-derives the disclosure from
archive emptiness fails this AC.

**And given** `moai todo history t9999` on the same queue (`9999 > last_seq`),
**when** it runs, **then** stdout reports `absent` and stderr carries **no** such
note — a never-issued id is not dressed up as a destroyed one.

```
go test ./internal/cli/ -run TestTodoHistoryDisclosesPreArchiveQueue -count=1
```

### AC-TAQ-005 — bare ordinals normalize

**Given** a `FIXTURE` queue holding `t1`,
**when** `moai todo history 1` runs,
**then** its stdout is byte-identical to `moai todo history t1`.

```
go test ./internal/cli/ -run TestTodoHistoryNormalizesBareOrdinal -count=1
```

## §D.2 Listing

### AC-TAQ-006 — newest first

**Given** a `FIXTURE` queue where `t1`, `t2`, `t3` were archived in that order,
**when** `moai todo history` runs with no id,
**then** stdout's three lines name `t3`, `t2`, `t1` in that order.

```
go test ./internal/cli/ -run TestTodoHistoryListsNewestFirst -count=1
```

### AC-TAQ-007 — the bound holds and is adjustable

**Given** a `FIXTURE` queue holding 25 archived entries,
**when** `moai todo history` runs with no flag, **then** stdout carries exactly 20
lines; **when** `--limit 5` is passed, exactly 5; **when** `--limit 0` is passed,
exactly 25.

```
go test ./internal/cli/ -run TestTodoHistoryLimitBound -count=1
```

### AC-TAQ-008 — truncation is stated

**Given** the same 25-entry queue,
**when** `moai todo history --limit 5` runs,
**then** stderr states that 20 entries were withheld, and stdout carries no such
note (a machine reading stdout is unaffected).

```
go test ./internal/cli/ -run TestTodoHistoryStatesWithheldCount -count=1
```

### AC-TAQ-009 — an empty archive says so

**Given** a `FIXTURE` queue with an empty archive,
**when** `moai todo history` runs with no id,
**then** stdout carries one explicit empty-archive line — not zero bytes — and the
exit code is 0.

```
go test ./internal/cli/ -run TestTodoHistoryEmptyArchiveIsExplicit -count=1
```

## §D.3 Invariants

### AC-TAQ-010 — the verb writes nothing

**Given** a `FIXTURE` queue, **and** the SHA-256 of every file under the queue's
state directory recorded before the call,
**when** `moai todo history` and `moai todo history <id>` each run,
**then** every recorded SHA-256 is unchanged, and no lock artifact remains.

```
go test ./internal/cli/ -run TestTodoHistoryLeavesStorageByteIdentical -count=1
```

### AC-TAQ-011 — live readers are byte-identical to a golden this change cannot have produced

Three clauses. Clauses 1 and 3 are asserted by the test and run in CI; clause 2
is a Definition-of-Done gate the implementer runs locally and cites.

What clause 1 establishes, stated exactly: **the goldens were added by a commit
whose tree did not contain the verb, and their bytes have not moved since.** It
does *not* establish that those bytes were produced by running that tree's
binary — clause 2 is what closes that, and clause 2 is a manual gate. The
previous wording ("a golden captured before this change") named no mechanism at
all, so an implementer who built the change and then captured the goldens
recorded the post-change output, including a broken one, and the guard for
REQ-TDG-007 went green while the invariant was broken. That path is now
mechanically closed; the residuals that survive are named below rather than
implied away.

**Clause 1 — capture provenance and integrity.** The goldens live under
`internal/cli/testdata/golden/live-readers/`. Their introducing commit `C` must
be the *only* commit that added them, a proper ancestor of `HEAD`, and its tree
must not contain the verb; and the goldens' current bytes must still be `C`'s:

```
GOLDENS=internal/cli/testdata/golden/live-readers/
test "$(git log --diff-filter=A --format=%H -- "$GOLDENS" | wc -l | tr -d '[:space:]')" -eq 1
C=$(git log --diff-filter=A --format=%H -- "$GOLDENS" | tail -1)
test -n "$C"
test "$C" != "$(git rev-parse HEAD)"
git merge-base --is-ancestor "$C" HEAD
! git cat-file -e "$C:internal/cli/todo_history.go" 2>/dev/null
! git grep -q "newTodoHistoryCmd" "$C" -- internal/cli/
git diff --exit-code "$C" -- "$GOLDENS"
```

Every command is load-bearing and each has a reachable failing input. The
**measured** rows below were reproduced in a throwaway repository at plan phase
(`goldens` commit → `verb` commit → tampering commit, plus a separate
single-commit repository standing in for a squash); the **reasoned** rows state
an expectation that was not executed, and the run phase is obliged to observe
each of them RED once rather than inherit them as established.

| Command | Failing input that reaches it | Status | Result |
|---|---|---|---|
| `test … -eq 1` | the goldens are added across two commits, so `tail -1` would silently take the oldest | **measured** | a second adding commit makes the count `2`; `-eq 1` fails. Also `0` at a tree with no goldens |
| `test -n "$C"` | no golden commit exists at all | **measured** | `$C` empty at `2c18091d1` (the goldens directory does not exist) |
| `test "$C" !=` | goldens committed in the same commit as the verb | **measured** | in the single-commit repository, `C == HEAD` |
| `merge-base --is-ancestor` | `C` unreachable (shallow checkout) | *reasoned* | expected to fail — the safe direction, since unverifiable provenance must not read as verified |
| `! git cat-file -e` | goldens captured from a tree in which the verb already exists | **measured** | `git cat-file -e "$C:<verb>"` exits `0` where `C`'s tree holds the verb, so the `!` inverts to fail |
| `! git grep -q` | same input as the row above | *reasoned* | expected to fail alongside `cat-file`; not separately executed |
| `git diff --exit-code` | **the modify path**: land M0 honestly, break the default read in M1, then edit the goldens to match | **measured** | `--diff-filter=A` does **not** move `C` for a modifying commit — `C` was byte-identical before and after a commit that rewrote the golden bytes, so every other command above still passed, and this one exited `1` |

The last row is why that command exists. Without it, clause 1 pins the commit
that *added* the goldens while saying nothing about what the goldens now contain,
so the self-certification this AC exists to prevent is deferred by one commit
rather than eliminated. Clause 2 does not close it either, being a manual gate.

**Clause 1 is scoped to the pre-integration branch** — the card branch
`WT-todo-done-history` and its pull request, which is where M0-before-M1 is a
property that can still be violated. It is not an invariant of the integration
branch, and it must not be asserted there. Under a squash the two milestones
collapse into one commit whose tree contains the verb, so `C` becomes `HEAD` and
`! cat-file` inverts to fail — permanently, on every later run (measured: a
single commit holding both goldens and verb resolves `C == HEAD` and
`git cat-file -e "$C:<verb>"` exits `0`). A guard that is red forever after
integration is a guard that gets deleted. `plan.md` §D carries the merge-method
dependency this creates.

**Clause 2 — regeneration (Definition of Done, run locally, output cited).**
Clause 1 proves `C`'s *tree* lacked the verb; it does not prove the golden bytes
came from that tree rather than from a dirty working directory. Clause 2 closes
that: build the binary from `C` in a throwaway worktree, run it against the same
`FIXTURE`, and diff each stream against the committed golden.

```
C=$(git log --diff-filter=A --format=%H -- internal/cli/testdata/golden/live-readers/ | tail -1)
TMP=$(mktemp -d); git worktree add --detach "$TMP/base" "$C"
(cd "$TMP/base" && go build -o "$TMP/moai-base" ./cmd/moai)
# run the six reads against FIXTURE with "$TMP/moai-base"; diff each against the golden
git worktree remove --force "$TMP/base"
```

**Clause 3 — the comparison itself.** Given a `FIXTURE` queue holding both live
and archived cards, when `list`, `list --json`, `next`, `why <id>`, `analyze` and
the state counts each run under the post-change binary, then each output is
byte-identical to the corresponding golden. The test fails on any difference, so
REQ-TDG-007's invisibility cannot be relaxed silently.

```
go test ./internal/cli/ -run TestLiveReadersUnchangedByHistoryVerb -count=1
```

**Residual, stated rather than discovered later — clause 3 reds on develop
drift.** The goldens are byte-comparisons captured from an ancestor of a moving
`develop`. Any change landing on `develop` that alters what `list`, `list --json`,
`next`, `why` or `analyze` print — a change with nothing to do with this SPEC —
turns clause 3 red for that reason. That is in one sense the point: the clause
cannot distinguish *this* change growing the default read from *another* change
growing it, and the conservative direction is to fail. But it means clause 3 is
not a guard the run phase can leave unattended across a long-lived branch. Two
consequences the run phase owns: rebase before relying on a green clause 3, and
when it reds, diff the golden against the current `develop` output **before**
regenerating — a regeneration performed without reading the diff converts an
unrelated default-read growth into a silently accepted one, which is the same
failure this AC exists to catch, arriving from the other side. Regenerating the
goldens also re-opens clause 1: the regenerating commit is a *modify*, so `C`
does not move and `git diff --exit-code "$C"` reds until the goldens are
re-captured as a fresh M0-shaped commit on the rebased branch.

### AC-TAQ-012 — no schema change

**Given** a `FIXTURE` database opened by the built binary,
**when** its table list and the `items.state` `CHECK` clause are read from
`sqlite_master`, **and** `schema_version` is read from `meta`,
**then** the table set is exactly `{meta, items, findings, archived_items,
archived_findings}` plus the existing indexes, the `CHECK` clause still admits
exactly `queued`, `picked`, `dropped`, and `schema_version` is `"1"`.

```
go test ./internal/kanban/ -run TestTodoHistoryAddsNoSchemaChange -count=1
```

### AC-TAQ-013 — a pre-archive database degrades rather than fails

**Given** a `FIXTURE` database whose `archived_items` and `archived_findings`
tables have been dropped (the shape `internal/kanban/backlog_archive_test.go:98`
already constructs),
**when** `moai todo history` and `moai todo history <live-id>` each run,
**then** the live lookup still answers, stderr names the store that answered and
states that no archive is available, and the exit code is 0.

**And given** a `FIXTURE` project holding only a legacy `backlog.json` and no
`backlog.db` — the shape `LoadPure` serves at
`internal/kanban/backlog_store.go:551-561` @ `2c18091d1` — whose JSON carries no
`archived` key (the shape a pre-archive binary writes; measured in the operator's
own queue, see `spec.md` §D),
**when** `moai todo history <unknown-id>` runs,
**then** stderr names the legacy JSON store as the one that answered and states
that no archive is available, so the `absent` on stdout is not read as
authoritative. Failing input reachable by construction: an implementation that
inspects only the archive tables reports a bare `absent` here and fails.

```
go test ./internal/cli/ -run TestTodoHistoryDegradesWithoutArchiveTables -count=1
```

### AC-TAQ-014 — nothing prompts

**Given** the run-phase source for the verb, **located by its symbol rather than
by an assumed filename**, **when** it is scanned for a prompting call, **then**
the source is found and no match is in it:

```
SRC=$(git grep -l "newTodoHistoryCmd" -- 'internal/cli/*.go' ':!internal/cli/*_test.go')
test -n "$SRC"
for f in $SRC; do test -f "$f"; done
! grep -nE "AskUserQuestion|survey\.|promptui|bufio\.NewReader\(os\.Stdin\)" $SRC
```

The first three commands are the repair for a defect in the previous wording: it
grepped a hard-coded `internal/cli/todo_history.go`, and `grep` on a missing path
exits 2, which the leading `!` inverted into a PASS. If the run phase registers
the verb inside `todo.go` instead of creating a new file — a choice this SPEC
does not forbid — the old form reported PASS with no subject. `test -n "$SRC"`
now fails in that case, and the scan follows the verb wherever it lands.

Because no existing `todo` verb prompts, the negated grep is near-baseline-vacuous
on its own. The run phase must therefore demonstrate it RED once: plant
`bufio.NewReader(os.Stdin)` in the verb's source, observe the command exit
non-zero, remove it, observe exit zero. Both observations are cited in
`progress.md` §E.2.

```
go test ./internal/cli/ -run TestTodoHistoryNeverPrompts -count=1
```

## §D.4 Documentation

### AC-TAQ-015 — documented in both surfaces, mirror stays neutral

**Given** the run-phase change,
**when** both skill documents are read,
**then** each carries a row for `moai todo history`:

```
grep -c "moai todo history" .claude/skills/moai/workflows/todo.md
grep -c "moai todo history" internal/template/templates/.claude/skills/moai/workflows/todo.md
```

both ≥ 1, **and** the mirror carries no internal content:

```
! grep -nE "SPEC-[A-Z0-9-]+-[0-9]{3}|REQ-[A-Z]+-[0-9]{3}|20[0-9]{2}-[0-9]{2}-[0-9]{2}|\b[0-9a-f]{9,40}\b" internal/template/templates/.claude/skills/moai/workflows/todo.md
```

```
go test ./internal/cli/ -run TestTodoSkillDocumentsHistoryVerb -count=1
```

---

## §D.5 Severity

| AC | Severity | If it fails |
|---|---|---|
| AC-TAQ-011 | BLOCKING | REQ-TDG-007 is broken; the queue's cheapest read grew |
| AC-TAQ-012 | BLOCKING | a field queue would need a rebuild (REQ-TDG-005) |
| AC-TAQ-010 | BLOCKING | a read verb mutates storage |
| AC-TAQ-001..003, 006 | BLOCKING | the surface does not answer the question it exists for |
| AC-TAQ-004, 008, 013 | SHOULD-FIX | the answer is right but its limits are unstated |
| AC-TAQ-005, 007, 009, 014, 015 | SHOULD-FIX | ergonomics and documentation |

## §D.6 Traceability

Every REQ-TAQ-NNN in `spec.md` §B has exactly one AC row above; the matrix at the
head of this file is the mapping, and a REQ with no row is a plan-phase defect.

## §D.7 Definition of Done

- Every AC above passes on the run-phase tree, cited by the command that decided it.
- AC-TAQ-011 clause 2 (golden regeneration from commit `C`) run locally with its
  diff output cited, and AC-TAQ-014's RED/GREEN plant-and-remove pair observed —
  both recorded in `progress.md` §E.2.
- `go vet ./internal/cli/... ./internal/kanban/...` clean.
- `golangci-lint run` shows no new finding attributable to the change.
- `make build` run, so the template mirror is embedded (Template-First).
- CI green on the pushed head — the local run is an early signal, not the verdict.
