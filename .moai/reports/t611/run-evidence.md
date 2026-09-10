# t611 run-phase evidence: ResolveBlocker scope

- Card: t611 (Class B, Tier S, cycle_type=tdd)
- Worktree: `.claude/worktrees/agent-a7560aa0a01956710`
- Branch: `WT-blocker-scope`
- Base HEAD: `d060e0d136f173f07353e46a709382d08ae25dd2`
- Code commit: `cbb48c5d27edb4bb8707d2cc5d0cd2af5c2bf70f`

The audit report's evidence (the probe measured on `main` at 2213871af, store.go:485) reached this worker inline from the lane, not as a file. Every measurement below was taken by this worker on this worktree's tree. The report's own numbers were not reused.

Every test run used the same environment scrub as the report, as one invocation: `unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test ...`. Raw logs are in the same directory.

## 1. Setup confirmation

Command: `git rev-parse HEAD; git branch --show-current; git status --short | wc -l`

```
d060e0d136f173f07353e46a709382d08ae25dd2
worktree-agent-a7560aa0a01956710
       0
```

HEAD matched the base the lane measured, `d060e0d13`, so the ancestry check was not needed and not run.

Command: `git branch -m WT-blocker-scope; echo "rename_exit=$?"; git branch --show-current`

```
rename_exit=0
WT-blocker-scope
```

No command that moves the tree (reset, checkout, switch, merge) was run.

## 2. Previous worker's two leads, confirmed by reading source (HEAD d060e0d13)

- `checkBlockerFiles` narrows by filename pattern. Confirmed at store.go:332: `pattern := filepath.Join(fs.stateDir, fmt.Sprintf("blocker-%s-%s-*.json", phase, specID))`
- `RecordBlocker` writes `unknown` into the filename when a field is empty. Confirmed at store.go:432-440: `phaseStr = "unknown"` / `specIDStr = "unknown"`, and only the filename is substituted; the JSON body is `json.MarshalIndent(report, ...)`, so the fields stay empty. The runtime log in §4 confirms this: `blocker-unknown-unknown-20260910-000100.json (spec="" phase="")`.

## 3. Callers

Command: `grep -rn '\.ResolveBlocker(' internal pkg cmd`

```
internal/session/store_test.go:298:	if err := store.ResolveBlocker(PhaseRun, "SPEC-001", resolution); err != nil {
internal/session/store_test.go:736:	err := store.ResolveBlocker(PhaseRun, "SPEC-NONE", "resolution")
internal/session/store_test.go:762:	err := store.ResolveBlocker(PhaseRun, "SPEC-ONLY-RESOLVED", "new resolution")
internal/session/store_test.go:795:	if err := store.ResolveBlocker(PhaseRun, "SPEC-MULTI", "resolved via test"); err != nil {
```

Result: there is no non-test caller in Go source. This grep cannot see shell hooks or reflection, so it does not establish that the defect has no effect in production use.

## 4. Reproduction and pre-edit RED (HEAD d060e0d13 plus the new test file only, store.go unmodified)

Expected outcomes, written down before running:

| Cell | Test | Expected | Observed |
|---|---|---|---|
| (i) single match, others older | `TestResolveBlockerScope_SingleMatchResolved` | GREEN | GREEN |
| (ii) no match, other unresolved blockers exist | `TestResolveBlockerScope_NoMatchResolvesNothing` | RED | RED |
| (iii) a different SPEC (same phase) is newer | `TestResolveBlockerScope_NewerOtherSpecUntouched` | RED | RED |
| (iv) same SPEC, a different phase is newer | `TestResolveBlockerScope_NewerOtherPhaseUntouched` | RED | RED |
| report criterion (probe fixture) | `TestResolveBlockerScope_ReportCriterion` | RED | RED |
| design edge (empty-field record is newer) | `TestResolveBlockerScope_EmptyFieldRecordNotMatched` | RED | RED |

Every observed result matched its expectation, so **the defect reproduces on this tree.** Every cell's fixture holds several blocker files, and every file other than the target is compared byte-for-byte before and after the call.

Command: `unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test -count=1 -v -run 'TestResolveBlockerScope_' ./internal/session/ > .moai/reports/t611/pre-edit-red.log 2>&1; echo "exit=$?"` → `exit=1`

```
=== RUN   TestResolveBlockerScope_SingleMatchResolved
--- PASS: TestResolveBlockerScope_SingleMatchResolved (0.00s)
=== RUN   TestResolveBlockerScope_NoMatchResolvesNothing
    store_resolve_scope_test.go:110: ResolveBlocker with no matching blocker returned nil error
    store_resolve_scope_test.go:112: non-target blocker changed: blocker-plan-SPEC-A-20260910-000200.json (spec="SPEC-A" phase="plan" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NoMatchResolvesNothing (0.00s)
=== RUN   TestResolveBlockerScope_NewerOtherSpecUntouched
    store_resolve_scope_test.go:126: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
    store_resolve_scope_test.go:126: non-target blocker changed: blocker-run-SPEC-B-20260910-000100.json (spec="SPEC-B" phase="run" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NewerOtherSpecUntouched (0.00s)
=== RUN   TestResolveBlockerScope_NewerOtherPhaseUntouched
    store_resolve_scope_test.go:140: non-target blocker changed: blocker-plan-SPEC-A-20260910-000100.json (spec="SPEC-A" phase="plan" resolved=true resolution="approved A")
    store_resolve_scope_test.go:140: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
--- FAIL: TestResolveBlockerScope_NewerOtherPhaseUntouched (0.00s)
=== RUN   TestResolveBlockerScope_ReportCriterion
    store_resolve_scope_test.go:154: non-target blocker changed: blocker-plan-SPEC-B-20260910-000100.json (spec="SPEC-B" phase="plan" resolved=true resolution="approved A")
    store_resolve_scope_test.go:154: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
--- FAIL: TestResolveBlockerScope_ReportCriterion (0.00s)
=== RUN   TestResolveBlockerScope_EmptyFieldRecordNotMatched
    store_resolve_scope_test.go:168: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
    store_resolve_scope_test.go:168: non-target blocker changed: blocker-unknown-unknown-20260910-000100.json (spec="" phase="" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_EmptyFieldRecordNotMatched (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/session	0.358s
FAIL
```

The report criterion reproduced in the same shape as the report: SPEC-B/plan was resolved and SPEC-A/run was left outstanding.

## 5. Design decision: empty-field records

A record whose Phase or SPECID is empty is matched by exact comparison on the JSON fields, so it matches no request carrying non-empty values (pinned by `TestResolveBlockerScope_EmptyFieldRecordNotMatched`). The reason: if an empty field acted as a wildcard, cell (ii)'s rule "resolve nothing on a mismatch" would break exactly for records with no scope, so the original defect would come back through the side door. The `unknown` in the filename is only a placeholder that `RecordBlocker` substitutes when naming the file, so it was not used as a matching key.

Consequence for an existing test: `TestFileSessionStoreResolveBlocker` (store_test.go:274) records a blocker with no Phase or SPECID, then resolves it under `(PhaseRun, "SPEC-001")`, which means it depended on the defective behavior (resolving a blocker outside the requested scope). Right after the fix, that single test fails:

Command: `unset ... && go test -count=1 ./internal/session/... > .moai/reports/t611/post-fix-pkg-before-fixture.log 2>&1; echo "exit=$?"` → `exit=1`

```
WARN: checkpoint for run/SPEC-RESUME-001 is stale (age=1h0m0s); loading anyway (--resume)
--- FAIL: TestFileSessionStoreResolveBlocker (0.00s)
    store_test.go:299: ResolveBlocker() failed: no outstanding blocker found
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/session	6.288s
FAIL
```

That test's fixture now carries `Phase: PhaseRun, SPECID: "SPEC-001"` (4 lines added). Its assertions are unchanged. **This is a modification to an existing test**, and it is disclosed here and in the commit message.

## 6. The fix (store.go, `ResolveBlocker`)

```go
		if blocker.Phase != phase || blocker.SPECID != specID {
			continue
		}
```

The filter runs before the most-recent selection. The glob, the error text (`no outstanding blocker found`), filename generation, `BlockerReport`, and every other function are unchanged.

## 7. Post-fix GREEN (working tree = the code commit cbb48c5d2)

Command: `unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test -count=1 -v ./internal/session/... > .moai/reports/t611/post-restore-green.log 2>&1; echo "exit=$?"` → `exit=0`

Named results read from the log (`grep -n 'ResolveBlocker' .moai/reports/t611/post-restore-green.log`):

```
342:=== RUN   TestResolveBlockerScope_SingleMatchResolved
343:--- PASS: TestResolveBlockerScope_SingleMatchResolved (0.00s)
344:=== RUN   TestResolveBlockerScope_NoMatchResolvesNothing
345:--- PASS: TestResolveBlockerScope_NoMatchResolvesNothing (0.00s)
346:=== RUN   TestResolveBlockerScope_NewerOtherSpecUntouched
347:--- PASS: TestResolveBlockerScope_NewerOtherSpecUntouched (0.00s)
348:=== RUN   TestResolveBlockerScope_NewerOtherPhaseUntouched
349:--- PASS: TestResolveBlockerScope_NewerOtherPhaseUntouched (0.00s)
350:=== RUN   TestResolveBlockerScope_ReportCriterion
351:--- PASS: TestResolveBlockerScope_ReportCriterion (0.00s)
352:=== RUN   TestResolveBlockerScope_EmptyFieldRecordNotMatched
353:--- PASS: TestResolveBlockerScope_EmptyFieldRecordNotMatched (0.00s)
368:=== RUN   TestFileSessionStoreResolveBlocker
369:--- PASS: TestFileSessionStoreResolveBlocker (0.00s)
413:=== RUN   TestResolveBlocker_NoBlockerFound
414:--- PASS: TestResolveBlocker_NoBlockerFound (0.00s)
415:=== RUN   TestResolveBlocker_OnlyResolvedExists
416:--- PASS: TestResolveBlocker_OnlyResolvedExists (0.00s)
417:=== RUN   TestResolveBlocker_MostRecentSelectedAmongMultiple
418:--- PASS: TestResolveBlocker_MostRecentSelectedAmongMultiple (1.10s)
```

Package totals (`grep -c -- '--- FAIL'` / `grep -c -- '--- PASS'` / verdict line):

```
0
216
442:ok  	github.com/modu-ai/moai-adk/internal/session	5.321s
```

`go vet ./internal/session/... > .moai/reports/t611/vet-post-restore.log 2>&1` → `exit=0`, log size 0 bytes (`wc -c`: `0 .moai/reports/t611/vet-post-restore.log`).

`gofmt -l internal/session/store.go internal/session/store_test.go internal/session/store_resolve_scope_test.go` → no output.

Attribution: no code file changed after this run; the only commands in between were `git diff`, `git add`, and `git commit`. `git status --short` after the commit shows only `?? .moai/reports/t611/`, so the tree measured is the tree committed in `cbb48c5d2`.

## 8. Mutant probe: only the specID half removed

Mutation: `if blocker.Phase != phase || blocker.SPECID != specID {` → `if blocker.Phase != phase {`

Prediction written before running: (ii) RED and (iii) RED; **report criterion GREEN**; (i), (iv), and the design edge GREEN. The lead expected the report criterion to go RED, but in the probe fixture the newer blocker is SPEC-B/**plan**, which differs from the request in both phase and SPEC. The phase filter alone still excludes it, so this fixture cannot detect the specID half on its own.

Command: `unset ... && go test -count=1 -v -run 'TestResolveBlockerScope_|ResolveBlocker' ./internal/session/ > .moai/reports/t611/mutant-no-specid.log 2>&1; echo "exit=$?"` → `exit=1`

```
=== RUN   TestResolveBlockerScope_SingleMatchResolved
--- PASS: TestResolveBlockerScope_SingleMatchResolved (0.00s)
=== RUN   TestResolveBlockerScope_NoMatchResolvesNothing
    store_resolve_scope_test.go:110: ResolveBlocker with no matching blocker returned nil error
    store_resolve_scope_test.go:112: non-target blocker changed: blocker-run-SPEC-C-20260910-000100.json (spec="SPEC-C" phase="run" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NoMatchResolvesNothing (0.00s)
=== RUN   TestResolveBlockerScope_NewerOtherSpecUntouched
    store_resolve_scope_test.go:126: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
    store_resolve_scope_test.go:126: non-target blocker changed: blocker-run-SPEC-B-20260910-000100.json (spec="SPEC-B" phase="run" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NewerOtherSpecUntouched (0.00s)
=== RUN   TestResolveBlockerScope_NewerOtherPhaseUntouched
--- PASS: TestResolveBlockerScope_NewerOtherPhaseUntouched (0.00s)
=== RUN   TestResolveBlockerScope_ReportCriterion
--- PASS: TestResolveBlockerScope_ReportCriterion (0.00s)
=== RUN   TestResolveBlockerScope_EmptyFieldRecordNotMatched
--- PASS: TestResolveBlockerScope_EmptyFieldRecordNotMatched (0.00s)
=== RUN   TestFileSessionStoreResolveBlocker
--- PASS: TestFileSessionStoreResolveBlocker (0.00s)
=== RUN   TestResolveBlocker_NoBlockerFound
--- PASS: TestResolveBlocker_NoBlockerFound (0.00s)
=== RUN   TestResolveBlocker_OnlyResolvedExists
--- PASS: TestResolveBlocker_OnlyResolvedExists (0.00s)
=== RUN   TestResolveBlocker_MostRecentSelectedAmongMultiple
--- PASS: TestResolveBlocker_MostRecentSelectedAmongMultiple (1.10s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/session	1.471s
FAIL
```

The prediction matched. (iii) failed for the right reason: `SPEC-B/run` was resolved and the target `SPEC-A/run` stayed unresolved. (ii) resolved `SPEC-C/run` instead, the no-match variant of the same failure. Detection of the specID half therefore rests on (ii) and (iii), not on the report criterion.

After restoring, `git diff` showed only the intended changes (store.go +5/-1, store_test.go +4; `2 files changed, 10 insertions(+), 1 deletion(-)`), and the post-restore run in §7 was GREEN.

## Gaps

- **Not measured**: whether the fix changes the behavior of any non-Go caller (shell hooks, reflection). The §3 grep cannot see those.
- **Not measured**: a mutant removing the phase half. Detection of the phase half is claimed from the pre-edit RED of (iv) only, with no dedicated mutant probe run.
- **Not pinned**: behavior of an empty request `ResolveBlocker("", "", ...)`. Exact comparison means it matches empty-field records, but no test fixes that.
- **Not measured**: the probe's output on the report's tree (main 2213871af). That evidence reached this worker inline from the lane; it was not re-run here.
- **Not absorbed**: the 28 commits between d060e0d13 and local develop d3b7d438d. Measuring whether `internal/session` imports any path they change is the lane's job at integration.
- **Not run**: `internal/cli` and the full suite (`go test ./...`), per instruction.

## Residual risk

- If some external path (such as a hook) relied on scope-free resolution by passing an arbitrary phase or SPEC, it will now get `no outstanding blocker found`. There is no Go caller, but the grep blind spot above remains.
- Blocker files written with empty fields in the past can no longer be released by a non-empty request. Only an empty request can release them.
- Filename timestamps have one-second granularity, so a same-scope record written within the same second can overwrite an earlier one. This predates the card and was not changed.
