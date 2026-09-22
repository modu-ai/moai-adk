# Run-phase Verdict — SPEC-QUEUE-UPGRADE-PROOF-001

Card: `t470` · Tier M · Branch `WT-queue-upgrade-proof` · cycle_type `tdd`

---

## Claim

The composed v3.1.2-to-next queue upgrade — the state-directory relocation
followed by the JSON-to-SQLite conversion, entered through the `moai todo`
command path rather than the `BacklogStore` API — works, and is now guarded by
an automated test that has been shown capable of failing.

Eleven acceptance criteria are asserted PASS: `AC-QUP-001a`, `001b`, `002`,
`003`, `004`, `005`, `006`, the optional `007`, `008`, `009`, `010`. Two of
them carry a stated qualification rather than a bare pass — `AC-QUP-008` (its
run-wide limb is a Gap; see below) and `AC-QUP-009` (baseline substituted, with
the substitution recorded).

No production file was changed. `REQ-QUP-009` holds.

---

## Evidence

### The deliverable

`internal/cli/todo_composed_upgrade_test.go` — one new file, two tests:

- `TestTodoComposedUpgrade_FromLegacyV312Layout` (M1 / G1)
- `TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive` (M2 / optional)

### E1 — the AC matrix run

```
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -v -timeout 600s
=== RUN   TestTodoComposedUpgrade_FromLegacyV312Layout
--- PASS: TestTodoComposedUpgrade_FromLegacyV312Layout (0.59s)
=== RUN   TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive
--- PASS: TestTodoComposedUpgrade_ForwardCompatibleFieldsSurvive (0.45s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	2.475s
```

Recorded verbatim at `.moai/reports/t470/green.log`.

`AC-QUP-005` requires the named test be reported RUN and PASSING, and states
that a zero-match selector is a failure of the criterion. The selector matched
**2** tests, both reported `=== RUN` and `--- PASS`; the swept set is non-empty.

The per-criterion assertions inside those tests:

- **AC-QUP-001a** — decodes `todo list --json` and asserts ids `t2`/`t3`/`t5`,
  states `queued`/`picked`/`dropped`, texts, and the seeded ORDER, plus the
  picked card's `spec_id` (`SPEC-LEGACY-001`).
- **AC-QUP-001b** — the post-upgrade `add` must issue `t8` (seeded `last_seq` 7
  plus one). A mark re-derived from the items would have issued `t6`; the test
  says so in its own failure message.
- **AC-QUP-002** — after the command: `todo/` exists, `todo/backlog.db` exists
  and is non-empty, `todo/companions.json` is byte-identical to the seeded
  sentinel, and `kanban/` no longer exists. The two preconditions are asserted
  BEFORE the command runs: both legacy files present, `todo/` absent.
- **AC-QUP-003** — `todo/backlog.json.migrated` present, bytes identical to the
  seeded F1 fixture, and no `backlog.json` beside it.
- **AC-QUP-004** — `todo/backlog.db` exists and is non-empty.

### E2 — AC-QUP-006 fixture fidelity

```
$ sed -n '43,46p' internal/cli/todo_composed_upgrade_test.go \
    | grep -o '"\(version\|last_seq\|items\|findings\|archived\)"' | sort -u
"items"
"last_seq"
"version"
```

The F1 literal carries the three fields v3.1.2's `BacklogRecord` actually has,
and neither `findings` nor `archived`.

### E3 — AC-QUP-010, RED before GREEN

The mutation `acceptance.md` names was applied to the TEST's fixture only:
`seedLegacyV312Layout(..., preCreateCurrentDir: true)` additionally creates an
EMPTY `<root>/.moai/state/todo/` before the first command. `resolveStateDir`
then takes the stale-copy branch (`internal/kanban/state_dir.go:81-88`) and
returns the current directory unconditionally; the relocation branch
(`:100-103`) is never reached, and migration
(`internal/kanban/backlog_store.go:603-607`, state B) never sees the seeded
document.

The mutation forces the current-name directory to exist, which the AC-QUP-002
precondition asserts must NOT — so the same flag relaxes exactly that one
precondition limb, and only under the mutation, letting RED land on the
post-command symptom the criterion names instead of stopping at the setup.

```
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade_FromLegacyV312Layout' -v -timeout 600s
=== RUN   TestTodoComposedUpgrade_FromLegacyV312Layout
    todo_composed_upgrade_test.go:236: relocation sentinel must be readable under the current name: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout1675990274/002/.moai/state/todo/companions.json: no such file or directory
    todo_composed_upgrade_test.go:236: legacy state dir "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout1675990274/002/.moai/state/kanban" must no longer exist after the upgrade (stat err = <nil>)
    todo_composed_upgrade_test.go:236: quarantined legacy document "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout1675990274/002/.moai/state/todo/backlog.json.migrated" must exist: open /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoComposedUpgrade_FromLegacyV312Layout1675990274/002/.moai/state/todo/backlog.json.migrated: no such file or directory
    todo_composed_upgrade_test.go:249: composed upgrade yielded 0 items, want 3: []
--- FAIL: TestTodoComposedUpgrade_FromLegacyV312Layout (1.26s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.688s
```

Recorded verbatim at `.moai/reports/t470/red-mutation.log`.

`AC-QUP-002` is the NAMED failing criterion and it failed on both of its
load-bearing limbs: the sentinel did not arrive under the current name, and the
legacy directory **still exists** (`stat err = <nil>`). `AC-QUP-003` (no
quarantine) and `AC-QUP-001a` (empty queue) went red alongside as corroboration,
exactly as the criterion predicts.

One limb did NOT go red and is worth recording: `todo/backlog.db` exists and is
non-empty even under the mutation, because the store creates an empty database
at the resolved path. That is precisely why the sentinel limb and the
negative-existence limb are the ones AC-QUP-002 leans on — a database-exists
check alone would have passed vacuously here.

The mutation was then reverted (both flags back to `false`, verified by grep:
lines 221/223 and 294/296 all read `false`) and the same test re-run — the E1
output above is that reverted GREEN.

The mutation explicitly rejected in `acceptance.md` — seeding under the CURRENT
directory name — was not used.

### E4 — AC-QUP-008 isolation

Structural, inside the test: `composedUpgradeFixture` builds a committed temp
git repository via `todoFixture` (`t.TempDir()` + `initGitRepo` +
`t.Setenv("CLAUDE_PROJECT_DIR", …)`) and additionally redirects
`userHomeDirFn` to a second temp directory. `assertQueueRootIsolated` then
asserts, BEFORE any command runs, that the resolved queue root is the temp
fixture root and lies under the OS temp tree. `runTodo`'s own t422 guard fails
the test if the root ever resolves to the live repository.

Observational, on the live file. The primary checkout was derived rather than
hardcoded — `git rev-parse --path-format=absolute --git-common-dir` run as its
own command (the worktree session guard refuses the nested `$(...)` form), its
parent being `/Users/goos/MoAI/moai-adk-go`, so
`QUEUE_DB=/Users/goos/MoAI/moai-adk-go/.moai/state/todo/backlog.db`. The
derivation succeeded and returned an absolute path, so the criterion's
resolution-failure branch did not fire.

Controlled window — a window containing ONLY this card's tests:

```
$ shasum -a 256 "$QUEUE_DB"   # before
ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f
$ stat -f '%m %z' "$QUEUE_DB" # before
1788422246 368640
$ go test ./internal/cli/ -run 'TestTodoComposedUpgrade' -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	3.068s
$ shasum -a 256 "$QUEUE_DB"   # after
ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f
$ stat -f '%m %z' "$QUEUE_DB" # after
1788422246 368640
```

Byte-identical, same mtime, same size. A second, much longer window — the full
`./internal/cli/` package suite, 934.6s of wall time (mtimes `1788423097` →
`1788424038`) — left the same file equally untouched.

`git status --short .moai/state/` was NOT used: `.gitignore:311` ignores that
tree, so it prints nothing whether or not the queue was mutated and cannot fail.

### E5 — quality gate

```
$ go test ./internal/kanban/... -count=1 -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/kanban	173.279s

$ go test ./internal/cli/ -count=1 -timeout 1800s
ok  	github.com/modu-ai/moai-adk/internal/cli	934.629s
exit=0

$ go vet ./internal/cli/ ./internal/kanban/...
(no output; vet_exit=0)

$ gofmt -l internal/cli/todo_composed_upgrade_test.go
(no output)
```

`./internal/cli/` took 934.6s — past the Bash tool's own 600s ceiling, which is
why it was run in the background with `-timeout 1800s` (C-3). Its log is at
`.moai/reports/t470/cli_suite.log`. No full local suite (`go test ./...`) was
run (C-2). No test spawns background load (C-4).

### E6 — AC-QUP-009, no production change

```
$ git diff --stat 6765a75c0..HEAD
 .moai/reports/t470/plan-audit-iter2.md             | 480 +++++++++++++++++++++
 .moai/reports/t470/plan-audit.md                   | 388 +++++++++++++++++
 .../SPEC-QUEUE-UPGRADE-PROOF-001/acceptance.md     | 248 +++++++++++
 .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/plan.md   | 279 ++++++++++++
 .../specs/SPEC-QUEUE-UPGRADE-PROOF-001/progress.md |  60 +++
 .moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/spec.md   | 290 +++++++++++
 6 files changed, 1745 insertions(+)
```

**CORRECTED at sync-audit (card t470, auditor finding F1).** The block above is
a stale measurement and the sentence that followed it was FALSE at the final
HEAD. Both defects are recorded rather than quietly overwritten, because the
falsified sentence was already committed and a later reader would otherwise
believe it.

- **The diff was measured BEFORE the M1 commit**, so it does not contain
  `internal/cli/todo_composed_upgrade_test.go` — the very path the claim
  generalizes over. The claim outran its own evidence.
- **`CHANGELOG.md` is outside the `acceptance.md` allowlist.** It was added by
  the sync commit `305a39bd6`, after this block was written.

Re-measured at final HEAD `077da90b8`:

```
$ git diff --name-only 6765a75c0..HEAD | grep -v '_test\.go$' \
    | grep -v '^\.moai/specs/SPEC-QUEUE-UPGRADE-PROOF-001/' \
    | grep -v '^\.moai/reports/t470/'
CHANGELOG.md
```

15 paths change in total: 1 `_test.go`, 4 SPEC artifacts, 9 under
`.moai/reports/t470/`, and `CHANGELOG.md`. **`CHANGELOG.md` is a sync-phase
deliverable, argued in `progress.md`, and is NOT in the allowlist `AC-QUP-009`
names.** The criterion's own allowlist was written in plan-phase against a
run-phase diff and did not anticipate the sync-phase artifact; the honest
disposition is that the allowlist is incomplete, NOT that the edit was improper.

The REQUIREMENT the criterion serves — `REQ-QUP-009`, no production change — is
verified separately and holds at final HEAD:

```
$ git diff --name-only 6765a75c0..HEAD -- internal/ pkg/ cmd/ | grep -v '_test\.go$' | wc -l
       0
```

Zero non-test files under `internal/`, `pkg/`, or `cmd/`. What was wrong here is
the RECORD, not the work.

### E7 — commits

- `d2675d57b` — M1: the test file + the `draft → in-progress` status transition.
- (M3 commit SHA recorded in the branch log; carries this verdict, the evidence
  logs, and `progress.md` §E.2/§E.3.)

Both staged by explicit pathspec, with `git status --short` re-read immediately
before staging. No `git add -A`, no `git add .`, no `git commit -a`. Nothing
pushed — develop push is the lead's batch operation, and pushing a `WT-` branch
is prohibited.

---

## Baseline-attribution

- Tree: branch `WT-queue-upgrade-proof`, HEAD `ec830400e` at run-phase start,
  `d2675d57b` after M1. SPEC base pin `4e4607abe`; absorbed develop tip
  `6765a75c0`.
- Every command above was run in this worktree
  (`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t470`, confirmed by
  `git rev-parse --show-toplevel`) against this tree, in this run. No figure is
  carried over from another package, tree, or point in time.
- **AC-QUP-009 baseline substitution, and why.** `acceptance.md` names
  `git diff --stat 4e4607abe..HEAD`. That pin is now an ANCESTOR of the develop
  tip this branch absorbed, so the named diff spans 37 commits of unrelated
  develop work and no longer measures what this card authored. The measurement
  was taken against `6765a75c0` instead — itself a pinned SHA, not a moving ref.
  The criterion's intent (no production file changed by this card) is what was
  verified; the substitution changes the baseline, not the assertion.
- Live-queue baseline: derived, not hardcoded, per the E4 derivation above. The
  before-values the dispatch supplied
  (`4b4656fd9026a1300e16cd40b5ba57e7b5a0c4cc00dc830e34f2915b2202feb4`,
  `1788421830 368640`) were re-measured at run-phase start and matched exactly.

---

## Gaps

The G-numbering, complete:

- **G1 — DELIVERED.** The composed-path proof is the deliverable of this card.
- **G2 — ABSORBED into G1, therefore NOT a gap.** G2 asked for the chained case:
  from a real v3.1.2 layout, the rename and the SQLite conversion firing one
  after the other. Fixture F1 makes a single `moai todo` invocation do exactly
  that, so G2's content is carried by `AC-QUP-001a`, `001b`, `002`, `003`, `004`
  and `006` per `plan.md §A`. No separate requirement or criterion was added,
  and the earlier "closes as unstarted" contingency is withdrawn.
- **G3 — cross-PROCESS concurrency. EXCLUDED, not covered.** Two OS processes
  contending on the same queue is not tested here. The common framing of this
  gap is wrong and the correction matters:
  `internal/kanban/backlog_concurrency_test.go:135 TestConcurrencyStress`
  already exists, but it drives in-process goroutines against a single
  `NewBacklogStore(...)` object. The gap is cross-PROCESS, not "no concurrency
  test at all". Excluded because it requires spawning a second process, which
  C-4 forbids. Separate-card candidate.
- **G4 — split-brain guarded by notice rather than prevention. EXCLUDED, not
  covered.** `export-json` re-creates `backlog.json` at the canonical path, and
  while the store prefers the database, any consumer bypassing `BacklogStore`
  and reading the file directly gets a stale answer. Prevention-versus-notice is
  a design decision and belongs to its own card.
- **G5 — no `moai doctor` check for the stale layout. EXCLUDED, not covered.**
  `moai doctor` reports nothing about a project still on the old directory. The
  one situation where the silence bites: when relocation is REFUSED (a
  cross-device mount, an unexpected permission), `resolveStateDir` fails open
  and the user keeps running READ-ONLY on the legacy layout with no diagnostic
  on any surface. Adding the check is a feature; this card's scope is proof.

Beyond the G-numbering, what this run did NOT observe:

- **AC-QUP-008's run-wide window is a Gap, not a clean pass.** Measured across
  the whole run-phase window, the live queue DID change:
  `4b4656fd9026a1300e16cd40b5ba57e7b5a0c4cc00dc830e34f2915b2202feb4` →
  `ecefa722b3b1181a8864e02ff0257301ed368d3f652b439f35b20521004d802f`, mtime
  `1788421830` → `1788422246`, size unchanged. It is attributed to a foreign
  actor on four observations: `current-session-id.txt` names session
  `9318281e-…`, not this one (`4b71a715-…`); a session-start signature (that
  file plus `.moai/state/mcp-server/6845.json` plus `.moai/state/lsel/clusters.json`)
  is stamped at `1788421869`; `.moai/state/todo/backlog.lock` carries the SAME
  mtime as the changed database (`1788422246`), so a lock-taking WRITE reached
  the live queue; and `pgrep` showed three other lanes (t446, t454, t410)
  running `go test ./internal/cli/...` concurrently. The controlled window and
  the 934s full-suite window both left the file byte-identical, which is the
  positive evidence for this card. But the attribution is inference from
  timestamps and process listings, not proof, so it is recorded here rather
  than claimed as a clean PASS.
- The proof exercises the composed path through `todo list --json` and
  `todo add`. Other `moai todo` verbs (`next`, `done`, `drop`, `relate`,
  `export-json`) were not exercised against a legacy layout.
- Relocation REFUSAL (a cross-device rename failure) is not exercised here; the
  existing per-layer suite covers it.
- No `golangci-lint` run was performed — the SPEC's §C quality gate names
  `gofmt`, `go vet`, and the two-package test run, and those are what was run.
- Cross-platform build (`GOOS=windows`) was not run: this card adds a test file
  only, with no build-tag-sensitive code.

---

## Residual-risk

- **The mutation seam is committed code.** `preCreateCurrentDir` and
  `currentDirPreCreated` are permanent parameters, `false` on every committed
  call. That makes the falsifiability record reproducible by flipping two
  literals — but it also means a future editor could flip one and leave it
  flipped, silently disarming the AC-QUP-002 precondition. The parameters are
  documented in place with exactly that warning; no mechanical guard prevents it.
- **A green here does not prove the mechanism handles every legacy shape.** It
  proves the shape v3.1.2 actually writes survives the composition. A project
  whose legacy directory holds something unexpected — a partially-written
  document, a `backlog.json` with an in-flight marker, a directory the process
  cannot rename — takes a branch this proof does not enter.
- **AC-QUP-008's attribution could be wrong.** If some path in this card's own
  test run did touch the live queue in a way both controlled windows happened to
  miss, the foreign-actor attribution would be a false exoneration. The
  structural isolation assertions make this unlikely (the test fails outright if
  the queue root escapes the temp tree), but "unlikely" is the honest word.
- **The `internal/cli` suite ran under contention.** Three other lanes were
  running the same package concurrently. It passed, so contention did not
  produce a false red; a flaky test masked by timing under that load would not
  have been caught either.
- **The two commits are unpushed.** The branch worktree is the only copy of this
  work until the lead's batch merge lands on `origin/develop`. The worktree must
  not be disposed before then.
