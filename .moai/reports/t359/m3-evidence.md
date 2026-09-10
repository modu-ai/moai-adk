# M3 evidence — SPEC-TODO-LANDING-EVIDENCE-001 (card t359)

The recording verb. Verbatim command + output pairs, in the order they were run.

**Tree**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`
**Branch**: `WT-landing-evidence`
**HEAD at M3 entry, re-read by this agent as its first command**: `705838c2a62f5c9a9f73eefa27d9df257a685af1`
(the lead quoted the same value at dispatch; this is an independent re-read, not a carry-over)

---

## 0. Tree anchoring

```
$ pwd && git rev-parse --show-toplevel && git rev-parse HEAD && git branch --show-current && git status --short
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359
705838c2a62f5c9a9f73eefa27d9df257a685af1
WT-landing-evidence
```

Working tree clean at entry. `git rev-list --count origin/develop..HEAD` → `15`.

---

## 1. Pre-edit baseline (so a later red is attributable)

`internal/cli` had no measured state on this card — M1 cited a pre-edit baseline at `903bcc03c`
and M2 correctly named that as a carry-over. This run re-establishes it in this tree at
`705838c2a`.

```
$ go test ./internal/kanban/... -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	138.282s
kanban rc=0

$ go test ./internal/cli/... -count=1 -timeout 600s
cli rc=0
ok  	github.com/modu-ai/moai-adk/internal/cli/update/merge	4.381s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/plan	3.944s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	6.021s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	5.659s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	9.579s
```

Both green before any edit.

---

## 2. RED — the tests written first

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/todo_landed_test.go:104:10: it.Landing undefined (type kanban.BacklogItem has no field or method Landing)
internal/cli/todo_landed_test.go:107:15: it.Landing undefined (type kanban.BacklogItem has no field or method Landing)
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```

A compile-stage RED: neither the store field nor the verb existed.

---

## 3. GREEN, and the proof the tests actually ran

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1 -v | grep -E '^(=== RUN|--- |ok|FAIL|PASS)'
=== RUN   TestTodoLanded_OneRecordUnderTheLock
--- PASS: TestTodoLanded_OneRecordUnderTheLock (0.81s)
=== RUN   TestTodoLanded_StateCheckAndStatesUntouched
--- PASS: TestTodoLanded_StateCheckAndStatesUntouched (0.69s)
=== RUN   TestTodoLanded_WholeQueueUnmoved
--- PASS: TestTodoLanded_WholeQueueUnmoved (0.44s)
=== RUN   TestTodoLanded_ReplaceAndClear
--- PASS: TestTodoLanded_ReplaceAndClear (1.10s)
=== RUN   TestTodoLanded_ClearIsExclusive
--- PASS: TestTodoLanded_ClearIsExclusive (0.43s)
=== RUN   TestTodoLanded_SpecStatusReadNeverInvented
--- PASS: TestTodoLanded_SpecStatusReadNeverInvented (0.94s)
=== RUN   TestTodoLanded_NoSHADerivedFromTheGrep
--- PASS: TestTodoLanded_NoSHADerivedFromTheGrep (0.46s)
=== RUN   TestTodoLanded_SHAValidation
=== RUN   TestTodoLanded_SHAValidation/reachable_accepts_and_stores_the_full_SHA
=== RUN   TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable
=== RUN   TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/non-existent_object
=== RUN   TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/unreachable_commit
=== RUN   TestTodoLanded_SHAValidation/refusing_branches_exit_1,_write_nothing,_and_are_distinguishable/unresolvable_ref
=== RUN   TestTodoLanded_SHAValidation/no_git_is_unrunnable_and_refuses
=== RUN   TestTodoLanded_SHAValidation/the_card_id_reaches_neither_check
--- PASS: TestTodoLanded_SHAValidation (3.82s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	9.440s
```

The `-v` listing is here because a `-run` filter that matches nothing also prints `ok`. Eight top
-level tests and six subtests are named, so the pass is not a zero-match.

---

## 4. Planted mutants

Hashes before any plant:

```
$ shasum internal/cli/todo_landed.go internal/cli/todo_landed_test.go internal/kanban/backlog_store.go internal/kanban/backlog_migrate.go internal/kanban/status_read.go internal/cli/todo.go internal/cli/todo_surface_test.go
8ba42ab00e33fa7d89a1188c3b69ec630f0cfe06  internal/cli/todo_landed.go
f61f00c56b9fd1744d99b068fa94dac486389101  internal/cli/todo_landed_test.go
3b2a1277422cb383fcc3358dbbc00c8ba9045631  internal/kanban/backlog_store.go
95c3d7de9ea54c94ece1548534cf6d74fa7e5a9c  internal/kanban/backlog_migrate.go
f31030193774f7aa3e7578f6e4208929f3e7a6af  internal/kanban/status_read.go
e15abde4d7ed82b2f733d69652e94dfa8822a96d  internal/cli/todo.go
b2a56c4abd442cd10ba87830c1a6e6e7e5ff91f6  internal/cli/todo_surface_test.go
```

### 4.1 AC-TLE-004 — the write also sets `state = 'dropped'`

Mutant: one line added inside the `Mutate` callback,
`rec.Items[i].State = kanban.BacklogStateDropped`.

Hash with mutant: `b35ac61268409e4d0fb5ddae7643192aacbbc88b  internal/cli/todo_landed.go`

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1
--- FAIL: TestTodoLanded_StateCheckAndStatesUntouched (0.70s)
    todo_landed_test.go:322: card t1 state moved "queued" -> "dropped" across a landing record; evidence never transitions a card
    todo_landed_test.go:322: card t2 state moved "picked" -> "dropped" across a landing record; evidence never transitions a card
--- FAIL: TestTodoLanded_WholeQueueUnmoved (0.45s)
    todo_landed_test.go:351: the queue moved across a landing record
        before:
        t1|queued|1|alpha|<nil>
        t2|picked|2|bravo|<nil>
        t3|queued|3|charlie|SPEC-FIXTURE-LANDED-001
        t4|dropped|4|delta|<nil>
        t5|queued|5|echo|<nil>
        after:
        t1|queued|1|alpha|<nil>
        t2|picked|2|bravo|<nil>
        t3|dropped|3|charlie|SPEC-FIXTURE-LANDED-001
        t4|dropped|4|delta|<nil>
        t5|queued|5|echo|<nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	9.592s
FAIL
```

**Which assertion fired**: `todo_landed_test.go:322`, the per-card state comparison, **twice** —
once for `t1` (queued) and once for `t2` (picked). AC-TLE-004 requires BOTH cards' assertions to
fail; both did, and neither fired on a name match.

The DDL assertion at `:311` did NOT fire, correctly: the mutant writes a LEGAL state value rather
than adding a fourth, so the CHECK is genuinely untouched. AC-TLE-004's stated RED is the per-card
assertion, and that is the one that fired.

Post-revert hash: `8ba42ab00e33fa7d89a1188c3b69ec630f0cfe06` — identical to pre-plant.

### 4.2 AC-TLE-008 — the write also moves the card to the head of the queue

Hash with mutant: `54ce4f0c5d7500f90848380fe27bf1fcb25233c1  internal/cli/todo_landed.go`

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1
--- FAIL: TestTodoLanded_WholeQueueUnmoved (0.43s)
    todo_landed_test.go:351: the queue moved across a landing record
        before:
        t1|queued|1|alpha|<nil>
        t2|picked|2|bravo|<nil>
        t3|queued|3|charlie|SPEC-FIXTURE-LANDED-001
        t4|dropped|4|delta|<nil>
        t5|queued|5|echo|<nil>
        after:
        t3|queued|1|charlie|SPEC-FIXTURE-LANDED-001
        t1|queued|2|alpha|<nil>
        t2|picked|3|bravo|<nil>
        t4|dropped|4|delta|<nil>
        t5|queued|5|echo|<nil>
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	9.566s
FAIL
```

**Which assertion fired**: `todo_landed_test.go:351`, the whole-queue ordered-tuple comparison.
The item count stayed 5, so the count conjunct at `:348` did not fire — the reorder alone is what
this criterion detects, which is the behavioural (not name-based) form AC-TLE-008 requires.

Post-revert hash: `8ba42ab00e33fa7d89a1188c3b69ec630f0cfe06` — identical to pre-plant.

### 4.3 AC-TLE-012 — the SHA filled from the landed grep's first match

The mutant had to ALSO claim `operator` provenance. That is a finding in itself: M2's encoder
refuses a SHA with no provenance and refuses any provenance other than `operator`, so the naive
leak is structurally impossible and the only leak that survives encoding is one that **lies**
about where the SHA came from. REQ-TLE-012's structural half is doing work.

Hash with mutant (first observation): `355dad163ddb404d1193bd220d3f0fd09096ae28  internal/cli/todo_landed.go`

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1
--- FAIL: TestTodoLanded_NoSHADerivedFromTheGrep (0.53s)
    todo_landed_test.go:469: stored delivering SHA = "a18ca73", want empty: no --sha was supplied and the machine has no lawful source
    todo_landed_test.go:472: stored provenance = "operator" with no operator assertion
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	10.126s
FAIL
```

**A red is not automatically the red that was wanted.** Two of the criterion's three assertion
groups fired (`:469`, `:472`). The **containment** group — AC-TLE-012's own words, "none of the
three known SHAs appears in the stored value" — did **not**. Cause: the criterion's leak stores
git's `--oneline` abbreviation, which is **7** characters in this fixture repository, and the
clause probed the full SHA and a **9**-character prefix. Against the very leak it names, the
containment clause was **vacuous**.

This is reported rather than quietly repaired, and the repair is narrow: the probe is now 7
characters — git's own floor, so a longer abbreviation still contains it as a prefix. The mutant
was re-observed with the repair in place and the mutant still applied:

```
$ go test ./internal/cli/ -run 'TestTodoLanded_NoSHADerivedFromTheGrep' -count=1
--- FAIL: TestTodoLanded_NoSHADerivedFromTheGrep (0.51s)
    todo_landed_test.go:469: stored delivering SHA = "194be97", want empty: no --sha was supplied and the machine has no lawful source
    todo_landed_test.go:472: stored provenance = "operator" with no operator assertion
    todo_landed_test.go:487: stored value contains the abbreviated #3 matching commit (194be97)
        stored: {"ref":"HEAD","ref_head":"0148f36940d0f9088877470b7ed11882dc1627a0","observed_at":"2026-09-07T23:20:25Z","sha":"194be97","sha_source":"operator"}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.462s
FAIL
```

All three groups now fire, and `:487` names commit **#3** — the NEWEST mention, not a delivering
commit. That is M2's read-side finding reproduced mechanically on the write side: the leak names
whichever commit most recently mentioned the card.

The test-file hash changed across this repair, deliberately and as a declared edit:
`f61f00c56b9fd1744d99b068fa94dac486389101` → `bf9efd229159adb5580a4c6cd6a6770793491975`.
This is a test repair, NOT a surviving mutant — `internal/cli/todo_landed.go` returned to
`8ba42ab00e33fa7d89a1188c3b69ec630f0cfe06`, byte-identical to pre-plant.

### 4.4 F1 — the plan-audit's boundary premise, OBSERVED

The plan-audit's iteration-3 F1 fix was a **stated premise**: AC-TLE-020's attribution-boundary
clause fires only when the fixture ref's history mentions exactly one of the two card ids. The
audit ceiling is spent, so M3's own observation was the only place it could be checked.

The fixture's non-degeneracy is asserted in the test itself, not assumed (`mentions("t1") &&
!mentions("t2")`), and that assertion passed.

Mutant: the card token routed into validation — `validateSuppliedSHA` given the card id, and the
reachability check replaced with the card-token grep predicate REQ-1.10 forbids.

Hash with mutant: `dbd438dbef68ab1c3859dec388c9f9dbb43fb708  internal/cli/todo_landed.go`

```
$ go test ./internal/cli/ -run 'TestTodoLanded_SHAValidation' -count=1
--- FAIL: TestTodoLanded_SHAValidation (3.93s)
    --- FAIL: TestTodoLanded_SHAValidation/the_card_id_reaches_neither_check (1.82s)
        todo_landed_test.go:619: the accepting branch diverged by card id: t1 failed=false (), t2 failed=true (Error: todo landed: the reachability check failed: commit f2ff8736cc2116be5278b39c51e0de5d768ed691 is not reachable from HEAD; nothing was written
            Error: todo landed: the reachability check failed: commit f2ff8736cc2116be5278b39c51e0de5d768ed691 is not reachable from HEAD; nothing was written
            todo landed: the reachability check failed: commit f2ff8736cc2116be5278b39c51e0de5d768ed691 is not reachable from HEAD; nothing was written)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.838s
FAIL
```

**The premise is OBSERVED, not asserted.** `todo_landed_test.go:619` — the ACCEPTING branch —
fired: one `--sha`, two card ids, `t1` accepted and `t2` refused. This is exactly the surfacing
point the audit predicted, and it is what a card-token leak looks like.

The REFUSING branch's comparison at `:626` did not fire, correctly: an unreachable commit is
absent from both cards' greps, so both refuse alike there. The criterion asks for invariance on
both branches, and the leak breaks it on one — which is sufficient for the clause to fail, and is
the behaviour the audit described.

Post-revert hash: `8ba42ab00e33fa7d89a1188c3b69ec630f0cfe06` — identical to pre-plant.

### 4.5 No mutant survived

```
$ grep -n "MUTANT" internal/cli/todo_landed.go internal/cli/todo_landed_test.go internal/kanban/*.go internal/cli/todo.go
MUTANT residue scan rc=1 (1 = none found)
```

```
$ go test ./internal/cli/ -run 'TestTodoLanded' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	9.403s
```

---

## 5. Two guards from OTHER SPECs went red, and both were repaired rather than relaxed

The first full-suite run after implementation:

```
$ go test ./internal/kanban/... -count=1
--- FAIL: TestBacklogArchive_PerItemContractFrozen (0.00s)
    backlog_archive_test.go:50: BacklogItem carries 6 fields ([ID Text AddedAt SpecID State Landing]), want the frozen 5
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	136.790s
FAIL

$ go test ./internal/cli/... -count=1 -timeout 600s
--- FAIL: TestAuditLagUsesBinlagSeam (0.65s)
    mcp_build_identity_test.go:624: sweep: NEW ancestry hit todo_landed.go:216 — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)
    mcp_build_identity_test.go:624: sweep: NEW ancestry hit todo_landed.go:248 — a second comparison outside binlag.Evaluate (REQ-ABI-006 violation)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	462.700s
FAIL
```

### 5.1 `TestBacklogArchive_PerItemContractFrozen` (AC-TDG-004, SPEC-TODO-DESTRUCTIVE-GUARD-001)

A `NumField() != 5` freeze. `design.md` §5 sanctions the addition explicitly ("`BacklogItem` gains
one optional field carrying the record; the five that exist keep their names, types, and JSON
tags"), so the guard was rebuilt in the **declared-addition** shape the sibling verb-surface guard
already uses (`internal/cli/todo_surface_test.go`): the frozen five are pinned by ordered
`(name, type, json tag)`, additions must be declared with their type, and a count check keeps a
removal-plus-addition from cancelling out.

This is strictly **stronger** than the count check it replaces — a reorder, a retype, or a renamed
tag now fails, none of which the old form could see. Bumping `5` to `6` was rejected: it would have
weakened the guard to exactly the blindness AC-TLE-019 was written to close on the schema side.

### 5.2 `TestAuditLagUsesBinlagSeam` (AC-ABI-007, REQ-ABI-006)

An **exact-set** source sweep over `internal/cli` non-test files, matching the raw strings
`merge-base` and `is-ancestor`. It produced two hits:

- `todo_landed.go:248` was a **comment** quoting the flag. Pure text over-match, carrying no
  comparison. The comment was reworded to name the commands indirectly, which removes a phantom
  baseline coordinate that would have churned on every unrelated edit to the file. No content lost.
- `todo_landed.go:216` is the genuine invocation `design.md` §3 mandates for REQ-TLE-020. It is
  declared in the sweep's baseline with a comment stating why it is not a binlag comparison: it
  asks whether an operator-named commit is on an operator-named ref, which is a different question
  from "is the running binary behind its source", and routing it through `binlag.Evaluate` would
  make it answer that other question instead.

Both green after the repair:

```
$ go test ./internal/kanban/ -run 'TestBacklogArchive_PerItemContractFrozen' -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.454s

$ go test ./internal/cli/ -run 'TestAuditLagUsesBinlagSeam|TestTodoLanded|TestTodoSurface|TestTodoCmd_NoAskUserQuestion' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	10.157s
```

---

## 6. Final scoped verification

```
$ go test ./internal/kanban/... -count=1
kanban rc=0
ok  	github.com/modu-ai/moai-adk/internal/kanban	136.644s

$ go test ./internal/cli/... -count=1 -timeout 600s
cli rc=0
ok  	github.com/modu-ai/moai-adk/internal/cli/update/report	5.660s
ok  	github.com/modu-ai/moai-adk/internal/cli/wizard	5.714s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	8.260s

$ go vet ./internal/kanban/... ./internal/cli/...
vet rc=0

$ gofmt -l <the nine M3-touched files>
M3-touched gofmt: rc=0 (no output = all clean)
```

`gofmt -l internal/kanban/ internal/cli/` reports **28** files, none of them M3-touched. That is a
pre-existing condition of the two trees at `705838c2a`, recorded here so the number is not read as
a regression this milestone introduced. The nine touched files were checked individually and are
clean.

Final hashes:

```
aacac18416814ef702df5bc7a1b087c04d1de542  internal/cli/todo_landed.go
bf9efd229159adb5580a4c6cd6a6770793491975  internal/cli/todo_landed_test.go
e15abde4d7ed82b2f733d69652e94dfa8822a96d  internal/cli/todo.go
b2a56c4abd442cd10ba87830c1a6e6e7e5ff91f6  internal/cli/todo_surface_test.go
de4a4c5fbdfee6e9966800391ab061e67bc93466  internal/cli/mcp_build_identity_test.go
3b2a1277422cb383fcc3358dbbc00c8ba9045631  internal/kanban/backlog_store.go
95c3d7de9ea54c94ece1548534cf6d74fa7e5a9c  internal/kanban/backlog_migrate.go
f31030193774f7aa3e7578f6e4208929f3e7a6af  internal/kanban/status_read.go
124afbae0a71b48991baeffda0b705b7508c9490  internal/kanban/backlog_archive_test.go
```

`todo_landed.go` differs from the pre-mutant `8ba42ab0…` ONLY by the §5.2 comment rewording,
applied after the last revert. Every mutant was reverted to `8ba42ab0…` first, which is recorded
per plant above.

---

## 7. The scope disagreement between `plan.md` §F and what M3 requires

**Stated, not resolved silently.** `plan.md` §F M3 lists deliverables as
"`internal/cli/todo_landed.go` + tests; registration on the `todo` command", and assigns the
`backlog_migrate.go` SELECT (`:60`) and INSERT (`:276`) to **M5**. M3 cannot record anything under
that division, for two measured reasons:

1. `BacklogStore.Mutate` is the only locked write path, and it operates on an in-memory
   `BacklogRecord`. With no `Landing` field on `BacklogItem`, the callback has nothing to set.
2. `writeRecord` does `DELETE FROM items` followed by a re-INSERT of an explicit column list
   (`internal/kanban/backlog_migrate.go:262`, pre-edit). A record written by ANY other route —
   raw SQL, a dedicated statement — is therefore **erased by the next `Mutate` of any kind**,
   including `todo done` and `todo edit`. That is not a latent defect M5 could tidy up later: it
   would make M4's render read a column that reliably empties itself.

Resolution taken, and it is the minimum: M3 plumbs the `items` read and write plus the
`BacklogItem` field, and touches **nothing else** of M5's list. Explicitly LEFT to M5:

- the archived-table carry (`readArchive` `:121-123`, `writeArchive` `:196-201`)
- the JSON⇄SQLite migration round trip and the parity comparison (`:585-630`)
- AC-TLE-017's drop-mutant observation

`design.md` §5 already sanctions the field itself ("`BacklogItem` gains one optional field carrying
the record"), so what is in dispute is the milestone boundary, not the design.

---

## 8. Tooling findings (continuing `m2-evidence.md` § Tooling finding)

### 8.1 `go-error-ignored-blank` — new instances, and a rule defect

The rule fired on `internal/cli/todo_landed.go` at four `_, _ = fmt.Fprintf(...)` lines and on
`todo_landed_test.go` at one `ev1, _ = landingOf(...)` reassignment. Both are the over-match M2
recorded: the pattern `$_, $ERR = $FUNC($$$ARGS)` matches ANY node in the first position rather
than only the blank identifier, and it matches `=` but not `:=`.

The test-file instance was cleared by using a fresh `:=` binding, which reads better anyway.

The production instances were NOT reworked. `_, _ = fmt.Fprintf` is this surface's universal
idiom — **35** occurrences across the three neighbouring files:

```
$ grep -c '_, _ = fmt.Fprint' internal/cli/todo.go internal/cli/todo_pr.go internal/cli/todo_drop.go
internal/cli/todo_drop.go:6
internal/cli/todo.go:22
internal/cli/todo_pr.go:7
```

**Rule defect worth fixing separately**: the rule's own message says "Handle it or mark the
intentional ignore with a comment", and its `note` suggests `// nolint:errcheck` — but the rule
body is a bare `pattern:` with no comment exemption implemented
(`.moai/astgrep-rules/go/error-handling.yml`). The escape the message promises does not exist, so
the only ways to satisfy the scanner are to abandon the codebase idiom or to write the file by
another path. This is a defect in the rule, not in the code it flags.

### 8.2 `go-error-not-wrapped` — advisory, over-matching the same way

Four warnings on `return err` after a wrap has already happened upstream, and on
`&exitCodeError{...}` returns which are constructed errors rather than propagated ones. Warnings
only; left alone.

### 8.3 `sql-injection` — did not fire

Recorded as a negative: M1 and M2 both saw this rule fire on quoted SQL in prose. It did not fire
on this file, which quotes `SELECT landing IS NULL FROM items WHERE id = ?` in several places.
Reported so the absence is not later read as the rule having been dodged.

---

## 9. Known losses — deliberately NOT exported, and NOT citable as a verdict basis

Material named here was NOT measured, or was measured and then discarded. None of it may be cited
later as evidence.

1. **The full local test suite was never run.** Only `./internal/kanban/...` and
   `./internal/cli/...`, per `CLAUDE.local.md` §4 and the dispatch. Every other package's state at
   this HEAD is UNMEASURED by this milestone; CI owns that verdict.
2. **Cross-platform**: darwin/arm64 only. No windows or linux build or test was run. The verb
   spawns `git` and compares command-error strings in `gitUnrunnable`; the "no such file or
   directory" / "executable file not found" phrasings are **Go's own** `exec` error text rather
   than the OS's, so they are expected to hold — but that is an ARGUMENT, not a measurement, and
   the windows path is unobserved.
3. **`golangci-lint` was not run.** Only `go vet` and `gofmt`. The DoD names "the project linter";
   that check is unperformed for M3.
4. **The archive probe was discarded.** A throwaway test measured the
   `landed → done → undone` sequence:

   ```
   PROBE RESULT after landed->done->undone: record present=false sha=""
   ```

   The record is silently lost, because `writeArchive` carries no `landing` column. This is a
   **measured** gap, and it sits precisely on M5's declared sites. The probe file was deleted
   rather than kept, because it asserted nothing (a `t.Logf`) and a test that passes either way is
   a vacuous green. M5 should build the real assertion; until it does, the sequence loses data.
5. **`--json` / render behaviour is unmeasured.** M4's surface; the verb writes a column that
   nothing yet displays.
6. **No concurrency measurement beyond the two-goroutine pair** in AC-TLE-007. Higher contention,
   cross-process contention, and lock-timeout behaviour are unobserved.
7. **The evidence-file rendering itself is unverified against a schema.** There is no mechanical
   check that this document's claims match the commands above; a reader must compare them by hand.
