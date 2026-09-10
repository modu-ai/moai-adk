# t611 — lane verdict (Go full review F02 / evidence Q1)

card: t611 (Class B, Tier S)
worktree: .claude/worktrees/agent-a7560aa0a01956710
branch: WT-blocker-scope
base: d060e0d136f173f07353e46a709382d08ae25dd2 (runtime isolation base; ancestor of local develop d3b7d438d, internal/session delta 0 bytes — lead ruling (a), no tree move)
commits under verdict: cbb48c5d2 (code), 1e47aa479 (run-phase evidence)
measured by: lane-5 orchestrator, independently of the implementing agent's report

## Claim

1. `ResolveBlocker` released a blocker belonging to a different SPEC or phase. The defect
   reproduced on this tree (agent-measured, pre-edit RED table in run-evidence.md §4) in the
   same shape as the report: resolving SPEC-A/run released SPEC-B/plan and left SPEC-A/run
   outstanding.
2. The fix — skip any blocker whose Phase or SPECID differs from the request before selecting
   the most recent — makes every required cell green. Re-measured here by the lane.
3. **Both halves of the match condition are guarded, by different cells, and both halves are
   now measured rather than inferred.** The specID half is caught by (ii) and (iii)
   (agent-measured). The phase half is caught by (ii) and (iv) (lane-measured below). The
   report-criterion test guards neither half on its own.

## Evidence — independent re-measurement (lane)

Tree state immediately before the run:

```
$ git rev-parse HEAD            → 1e47aa479053a73ba0c88efd81569c3d9e63ddb1
$ git branch --show-current     → WT-blocker-scope
$ git status --porcelain        → (no output)
```

Same environment scrub as the report, output to file, exit code read with no pipe:

```
$ unset MOAI_SESSION_ID MOAI_PROJECT_ROOT MOAI_WORKTREE_ROOT CLAUDE_PROJECT_DIR && go test -count=1 -v ./internal/session/... > t611-lane-green.txt 2>&1
$ echo "TEST_EXIT=$?"           → TEST_EXIT=0
$ unset ... && go vet ./internal/session/... > t611-lane-vet.txt 2>&1
$ echo "VET_EXIT=$?"            → VET_EXIT=0      (log: 0 bytes)
```

Exit 0 alone is not a pass, so the file was read. Every ResolveBlocker-related test ran by name:

```
--- PASS: TestResolveBlockerScope_SingleMatchResolved (0.00s)
--- PASS: TestResolveBlockerScope_NoMatchResolvesNothing (0.00s)
--- PASS: TestResolveBlockerScope_NewerOtherSpecUntouched (0.00s)
--- PASS: TestResolveBlockerScope_NewerOtherPhaseUntouched (0.00s)
--- PASS: TestResolveBlockerScope_ReportCriterion (0.00s)
--- PASS: TestResolveBlockerScope_EmptyFieldRecordNotMatched (0.00s)
--- PASS: TestFileSessionStoreResolveBlocker (0.00s)
--- PASS: TestResolveBlocker_NoBlockerFound (0.00s)
--- PASS: TestResolveBlocker_OnlyResolvedExists (0.00s)
--- PASS: TestResolveBlocker_MostRecentSelectedAmongMultiple (1.10s)
```

Package totals from the same file: `--- FAIL` 0, `--- PASS` 216, verdict line
`ok  github.com/modu-ai/moai-adk/internal/session  5.546s`. These match the agent's reported
totals; they are this run's own figures, not a copy of them.

## Evidence — phase-half mutant (lane)

The agent ran only the specID-half mutant, leaving the phase half guarded by argument alone.
Closed here.

Mutation, shown as applied before running:

```diff
-		if blocker.Phase != phase || blocker.SPECID != specID {
+		if blocker.SPECID != specID { // MUTANT t611 lane: phase half removed
```

Prediction, written before running:

| cell | predicted | why |
|---|---|---|
| (i) single match | GREEN | only one SPEC-A record |
| (ii) no match | **RED** | the fixture's newest record is SPEC-A/**plan**; specID alone matches it |
| (iii) newer other SPEC | GREEN | different SPEC still filtered |
| (iv) newer other phase | **RED** | SPEC-A/plan is newer and now selected |
| report criterion | GREEN | SPEC-B still filtered |
| empty-field edge | GREEN | empty SPECID ≠ SPEC-A |
| four existing tests | GREEN | all use one matching SPEC and phase |

Run: `unset ... && go test -count=1 -v -run 'TestResolveBlockerScope_|ResolveBlocker' ./internal/session/ > t611-mutant-nophase.txt 2>&1` → `TEST_EXIT=1`. Observed, verbatim:

```
--- PASS: TestResolveBlockerScope_SingleMatchResolved (0.00s)
    store_resolve_scope_test.go:110: ResolveBlocker with no matching blocker returned nil error
    store_resolve_scope_test.go:112: non-target blocker changed: blocker-plan-SPEC-A-20260910-000200.json (spec="SPEC-A" phase="plan" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NoMatchResolvesNothing (0.00s)
--- PASS: TestResolveBlockerScope_NewerOtherSpecUntouched (0.00s)
    store_resolve_scope_test.go:140: target blocker-run-SPEC-A-20260910-000000.json not resolved: resolved=false resolution=""
    store_resolve_scope_test.go:140: non-target blocker changed: blocker-plan-SPEC-A-20260910-000100.json (spec="SPEC-A" phase="plan" resolved=true resolution="approved A")
--- FAIL: TestResolveBlockerScope_NewerOtherPhaseUntouched (0.00s)
--- PASS: TestResolveBlockerScope_ReportCriterion (0.00s)
--- PASS: TestResolveBlockerScope_EmptyFieldRecordNotMatched (0.00s)
--- PASS: TestFileSessionStoreResolveBlocker (0.00s)
--- PASS: TestResolveBlocker_NoBlockerFound (0.00s)
--- PASS: TestResolveBlocker_OnlyResolvedExists (0.00s)
--- PASS: TestResolveBlocker_MostRecentSelectedAmongMultiple (1.10s)
FAIL	github.com/modu-ai/moai-adk/internal/session	1.536s
```

Exactly (ii) and (iv) failed, each for the right reason — a wrong file resolved — and not a
compile error. Prediction matched every row.

Restore verified byte-exact:

```
$ git diff --quiet -- internal/session/store.go ; echo "STORE_DIFF_QUIET_EXIT=$?"   → 0
$ git status --porcelain                                                        → (no output)
$ git rev-parse HEAD                                                            → 1e47aa479...
```

The restored tree is byte-identical to the committed tree measured green above, so no
separate post-restore run was needed.

### Guard map

| condition half removed | cells that go RED | measured by |
|---|---|---|
| specID | (ii), (iii) | agent (run-evidence.md §8) |
| phase | (ii), (iv) | lane (above) |

## Evidence — code review (lane, by reading)

- The fix adds one exact-equality `continue` before the most-recent selection; the glob, the
  no-match error text, filename generation, and `BlockerReport` are unchanged.
- The new tests are not vacuous: non-target files are compared with `bytes.Equal`; the
  snapshot asserts the file COUNT, so a filename collision cannot silently merge two fixtures;
  cell (ii) asserts both a non-nil error AND no file changed.
- The existing `TestResolveBlocker_MostRecentSelectedAmongMultiple` stores both records as
  `SPEC-MULTI/run`, so after the fix it still exercises selection among multiple MATCHING
  records rather than passing because nothing matches.

**Off-scope change, disclosed.** `TestFileSessionStoreResolveBlocker`'s fixture gained
`Phase: PhaseRun, SPECID: "SPEC-001"`. The original recorded an unscoped blocker and resolved
it under `(PhaseRun, "SPEC-001")`, so it depended on the defect; it was the only test that
failed immediately after the fix (agent log `post-fix-pkg-before-fixture.log`). Assertions are
unchanged. Reported to the lead for acceptance.

**Correction to the lane's own dispatch.** The dispatch predicted the report-criterion test
would go RED under the specID-half mutant. That was wrong: its newer record is SPEC-B/plan,
different in phase as well, so the phase filter alone excludes it. The agent corrected this
with a pre-run prediction, and the measurement bore the correction out.

## Baseline-attribution

The independent run, the vet run, the mutant run, and the restore check above were all
executed in this run, in this worktree, at HEAD 1e47aa479 with a clean tree (the mutant run
with exactly the one shown mutation applied). Agent-measured results are labelled as such and
cited to run-evidence.md; none of their figures is presented as the lane's measurement.

## Gaps — not observed

- **The 28 commits between d060e0d13 and local develop are not yet absorbed**, and whether
  `internal/session` imports any path they change (`internal/cli/todo*`,
  `internal/cli/update/merge`, `internal/hook/quality`) is not yet measured. Per the lead, this
  is measured with `go list -deps` at absorb time, not left to the merged-tree re-run.
- Callers outside Go source (shell hooks, reflection) — the caller grep cannot see them.
- An empty request `ResolveBlocker("", "", ...)` is not pinned by a test. Under exact match it
  releases only empty-field records. Advisory, lead's call.
- The report's own tree (main 2213871af) was not re-run.
- `internal/cli` and the full suite were not run; the card does not touch them.
- darwin only.

## Residual-risk

- A non-Go path that relied on scope-free release would now receive
  `no outstanding blocker found`.
- Blocker records written with empty Phase/SPECID in the past can no longer be released by a
  non-empty request; only an empty request releases them.
- Filenames carry one-second timestamps, so two same-scope records written within one second
  collide. This predates the card and was not changed.

## Sync decision

**No CHANGELOG entry.** There is no non-test caller in Go source and no user-visible effect has
been established; the card forbids inflating the defect into a user-facing outage. The
function's behaviour change is recorded here and in commit cbb48c5d2. The lead may override.
