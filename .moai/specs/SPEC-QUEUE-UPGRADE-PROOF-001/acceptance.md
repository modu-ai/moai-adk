# Acceptance Criteria — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Tier M

Every criterion below is binary-testable and independently checkable by a
command. Sub-IDs carrying a trailing letter (`AC-QUP-001a` / `001b`) pair
sub-criteria within one logical criterion.

## §A AC matrix

| AC | Requirement | Mandatory | Summary |
|---|---|---|---|
| AC-QUP-001a | REQ-QUP-001 | yes | Cards and states survive the composed upgrade |
| AC-QUP-001b | REQ-QUP-001 | yes | `last_seq` survives the composed upgrade |
| AC-QUP-002 | REQ-QUP-002 | yes | Legacy directory is relocated to the current name |
| AC-QUP-003 | REQ-QUP-003 | yes | Legacy document survives under the `.migrated` name |
| AC-QUP-004 | REQ-QUP-004 | yes | SQLite artifact exists after the upgrade |
| AC-QUP-005 | REQ-QUP-005 | yes | The proof runs as an automated test in the suite |
| AC-QUP-006 | REQ-QUP-006 | yes | The F1 fixture carries the v3.1.2 record shape only |
| AC-QUP-007 | REQ-QUP-007 | no (optional) | `findings` and `archived` survive from the legacy directory |
| AC-QUP-008 | REQ-QUP-008 | yes | The proof never touches the live queue |
| AC-QUP-009 | REQ-QUP-009 | yes | No production behavior changed |
| AC-QUP-010 | REQ-QUP-010 | yes | RED was established before GREEN was claimed |

## §B Scenarios

### AC-QUP-001a — cards and states survive

Given a temporary project root that is a git repository and holds only the F1
layout — a legacy state directory `<root>/.moai/state/kanban/` containing
`backlog.json` with a `{version, last_seq, items}` record whose items span the
`queued`, `picked`, and `dropped` states —
When the first `moai todo` command runs against that root,
Then the listed queue contains exactly the seeded items, each with the state it
was seeded with, in the seeded order.

### AC-QUP-001b — `last_seq` survives

Given the same fixture, seeded with a `last_seq` strictly greater than the
highest item id,
When the first `moai todo` command runs and a new card is then added,
Then the new card's id is `last_seq + 1`, demonstrating the high-water mark
crossed the composition rather than being re-derived from the items.

### AC-QUP-002 — directory relocated

Given the same fixture, plus a sentinel registry file
`<root>/.moai/state/kanban/companions.json` seeded INSIDE the legacy directory
(`state_dir.go:13-16` records that the relocation moves the directory, not a
file list, so the registry files ride along by construction),
And given the preconditions asserted BEFORE any command runs — that
`<root>/.moai/state/kanban/backlog.json` and the sentinel both exist, and that
`<root>/.moai/state/todo/` does NOT exist —
When the first `moai todo` command completes,
Then `<root>/.moai/state/todo/` exists and holds the queue — observed as
`<root>/.moai/state/todo/backlog.db` existing and non-empty, the same
observation `AC-QUP-004` states, so no third weaker check is invented for it —
the sentinel is
readable at `<root>/.moai/state/todo/companions.json` with its seeded bytes,
and `<root>/.moai/state/kanban/` no longer exists.

The two preconditions are load-bearing, not ceremony. The final limb is a
negative-existence assertion, which passes vacuously when the legacy directory
was never created; asserting its presence first is what makes that limb
falsifiable. The sentinel supplies the matching POSITIVE evidence — the
directory did not merely disappear, its contents arrived at the new name.

### AC-QUP-003 — legacy document quarantined, not destroyed

Given the same fixture,
When the first `moai todo` command completes,
Then a file named `backlog.json.migrated` exists in
`<root>/.moai/state/todo/`, its bytes are identical to the seeded fixture, and
no file named `backlog.json` remains beside it.

### AC-QUP-004 — SQLite artifact present

Given the same fixture,
When the first `moai todo` command completes,
Then `<root>/.moai/state/todo/backlog.db` exists and is non-empty.

### AC-QUP-005 — the proof is automated

Given the delivered branch,
When `go test ./internal/cli/ -run <the new test name> -v` is run,
Then the named test is reported as run and passing — a zero-match selector is a
failure of this criterion, not a pass.

### AC-QUP-006 — fixture fidelity

Given the delivered test source,
When the F1 fixture literal is inspected,
Then it contains the keys `version`, `last_seq`, and `items` and no others —
specifically it contains neither `findings` nor `archived`, because
`v3.1.2`'s `BacklogRecord` has no such fields.

### AC-QUP-007 — forward-compatible fields survive (OPTIONAL)

Given a second fixture placed in the LEGACY directory carrying `findings` and
`archived` — a state reachable only for someone who ran a development build
before the release, never for a `v3.1.2` user —
When the first `moai todo` command runs against it,
Then both collections are present and equal in the migrated store.

This criterion is optional. Its absence does not fail the card; its presence
must carry the reachability caveat in a comment so a reader does not mistake it
for a user-facing scenario.

### AC-QUP-008 — isolation

Given the delivered test source,
When it is inspected and run,
Then every path it touches is rooted under a `t.TempDir()`, the resolved queue
root is asserted to be inside that temp directory before any command runs, and
the PRIMARY CHECKOUT's live queue is unmodified — verified by observing the
FILE, not git's view of it.

The queue at risk is the primary checkout's, NOT the worktree's, and that is by
design: `internal/kanban/todo_root.go:95-99` resolves `primaryCheckoutRoot` as
`filepath.Dir(dirs.CommonDir)`, so from any linked worktree the root the
production code returns is the primary checkout. A worktree-relative
`.moai/state/todo/backlog.db` names a file that does not exist there, and
comparing that absent file before and after would assert nothing.

The criterion therefore DERIVES the path the way the production code does,
rather than hardcoding an absolute path that is machine-specific and would rot:

```bash
QUEUE_DB="$(dirname "$(git rev-parse --path-format=absolute --git-common-dir)")/.moai/state/todo/backlog.db"
```

`--path-format=absolute` is required, not decorative: the bare
`--git-common-dir` prints an absolute path from a linked worktree but the
relative `.git` from the primary checkout, so omitting it makes the derivation
CWD-dependent — the very defect this limb is fixing.

A run-phase implementer verifying this criterion BY HAND from inside a worktree
session must issue the `git rev-parse` as its own command and substitute the
result, rather than nesting it in `$(...)` as written above: the worktree
session guard refuses the compound form as too complex to verify, and that
refusal looks like a broken criterion when it is not. The Go test is unaffected
— it never passes through that guard.

**Resolution failure is a FAILURE of this criterion, never a pass.** If the
`git rev-parse` exits non-zero, or the resolved `QUEUE_DB` does not begin with
`/`, the criterion is reported FAILED naming the resolution error. It must not
fall through to the absent-before → absent-after branch, which would silently
convert a broken derivation into a green.

With `QUEUE_DB` resolved, the observation is a before/after comparison recorded
in the verdict: capture `shasum -a 256 "$QUEUE_DB"` together with its mtime
(`stat -f '%m %z' "$QUEUE_DB"` on darwin) before the test run, and assert both
are identical afterwards. When the file does not exist before the run — the
resolution itself having succeeded — the assertion is that it still does not
exist afterwards.

`git status --short .moai/state/` is NOT a valid check here and must not be
used: `.gitignore:311` ignores `.moai/state/`, confirmed with
`git check-ignore -v .moai/state/todo/backlog.db`, so that command prints
nothing whether or not the live queue was mutated — it cannot fail, and a check
that cannot fail asserts nothing. The other two limbs of this criterion are
unchanged.

### AC-QUP-009 — no production change

Given the delivered branch,
When `git diff --stat 4e4607abe..HEAD` is read (the SPEC base pinned in
`plan.md:L3`; a moving ref would not be a baseline),
Then every changed path is either a `_test.go` file, a file under
`.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/`, or `.moai/reports/t470/`. No
non-test file under `internal/` is changed.

### AC-QUP-010 — RED before GREEN

Given the composed-path test with its fixture unchanged — the legacy directory
`<root>/.moai/state/kanban/` created and holding the seeded `backlog.json` and
the `companions.json` sentinel —
When the mutation is applied: **additionally pre-create an EMPTY
`<root>/.moai/state/todo/` directory before the first `moai todo` command
runs**, so both directories exist at resolution time,
Then `AC-QUP-002` fails, with the observable symptom that
`<root>/.moai/state/kanban/` STILL EXISTS after the command completes and no
`companions.json` appears under `<root>/.moai/state/todo/`.

The mutation is applied to the TEST's fixture only. No production file is
touched, so `REQ-QUP-009` holds while RED is being established.

Why this mutation severs the composed path (traced against the tree at base
`4e4607abe`, not assumed):

- `internal/kanban/state_dir.go:81-88` — `resolveStateDir` returns the CURRENT
  directory unconditionally the moment it exists, and the comment states the
  stale-copy policy explicitly: the legacy directory is then "left exactly where
  it is". The relocation branch at `state_dir.go:100-103` is never reached.
- `internal/kanban/backlog_store.go:603-607` — migration runs only in state B
  (`!dbExists && jsonExists`) at the RESOLVED queue path. With the resolved path
  now `todo/backlog.json`, which does not exist, neither `migrateLegacyBacklog`
  nor `quarantineLegacyBacklog` runs against the seeded document at all.

So the mutation makes legacy-DIRECTORY resolution the thing under test:
everything downstream of the relocation is bypassed, and the seeded queue is
never reached. `AC-QUP-001a` and `AC-QUP-003` go RED alongside `AC-QUP-002`
(an empty listed queue; no `backlog.json.migrated`), which is corroboration —
`AC-QUP-002` is the named criterion.

Why it cannot pass vacuously: `AC-QUP-002`'s preconditions assert the legacy
directory and its contents EXIST before the command runs. The negative-existence
limb therefore has something to be false about, and the sentinel limb is a
positive assertion that the mutation makes unsatisfiable.

Explicitly rejected as the mutation: seeding the fixture under the CURRENT
directory name instead of the legacy one. That mutation produces GREEN on every
criterion and demonstrates nothing —
`internal/kanban/backlog_migrate.go:434` keys `migrateLegacyBacklog` on the
RESOLVED queue path, never on the directory's name, so a `backlog.json` in
`todo/` migrates and quarantines exactly as one in `kanban/` does, and
`AC-QUP-002`'s negative limb passes vacuously because the legacy directory was
never created. It is recorded here so it is not re-proposed.

A test that has only ever been observed passing has not been shown to assert
anything. This criterion is what separates the proof from a vacuous green.

## §C Quality gate

- `go test ./internal/kanban/... ./internal/cli/` passes. Run with an explicit
  timeout at or above 600s: `./internal/cli/` alone runs past 600s on the
  development machine, so a default-timeout failure there measures the timeout,
  not the code.
- `gofmt -l` reports nothing for the changed files.
- `go vet ./internal/cli/ ./internal/kanban/...` is clean.
- No full local suite (`go test ./...`) is run.
- No background load is spawned by any test.

## §D Definition of Done

- [ ] AC-QUP-001a, 001b, 002, 003, 004, 005, 006, 008, 009, 010 all pass
- [ ] AC-QUP-007 either passes or is explicitly recorded as not attempted
- [ ] `.moai/reports/t470/verdict.md` exists on the branch and carries: Claim,
      Evidence (command + verbatim output), Baseline-attribution (tree SHA and
      the commands run in this run), Gaps, Residual-risk
- [ ] The Gaps section states the complete G-numbering: G2 ABSORBED into G1
      (covered by AC-QUP-001a/001b/002/003/004/006 per `plan.md §A`, therefore
      NOT a gap), and G3 (cross-process concurrency), G4 (split-brain guarded by
      notice), and G5 (`moai doctor` check) excluded and named as not covered
- [ ] The `AC-QUP-010` mutation's RED output is recorded verbatim in the verdict,
      alongside the GREEN run of the same test with the mutation reverted
- [ ] Quality gate above passes with its elapsed time recorded
- [ ] No production file changed (AC-QUP-009)
