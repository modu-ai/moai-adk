# SPEC-GUARD-LIVENESS-001 — Progress (card t333, surfacing model)

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored on `091966c55` @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`).

- Artifacts: `spec.md`, `plan.md`, `acceptance.md` (Tier M set) + this file.
- Requirements: 13 (Tier M ceiling 16). Acceptance criteria: 13 (ceiling 16). Both under, which is the outcome of the scope reduction.
- **Two baselines.** RED-now cells pin `091966c55`; card t326 citations pin `origin/develop` at `ec15ec2cd`, a diverged tree (diverged, `merge-base --is-ancestor` false). Each t326 citation names its tree inline — reading the baseline for a t326 surface reports a landed feature as absent (`spec.md` §A.10).
- Primary evidence artifact: `.moai/reports/t333/trigger-axis-observation.md` (tracked at `c30f761dd`).
- Every RED-now cell is pinned to `091966c55` and its command was re-run during the split; no cell was carried across without re-measurement.

### Audit history

| Iteration | Verdict | Score | Outcome |
|---|---|---|---|
| 1 | PASS-WITH-DEBT | 0.800 | 7 blocking defects (D1-D7) closed at v0.5.0 |
| 2 | PASS-WITH-DEBT | 0.800 (flat) | N1, N2, N5 closed at v0.6.0; Traceability 0.75→1.00, Completeness 1.00→0.75 |
| 3 | **FAIL + STOP** | 0.667 | Regression clause fired. Operator chose **scope reduction**, the audit's own recommendation. No fourth repair round. |
| 6 (iter 3, **terminal**) | **PASS-WITH-DEBT + STOP** | 0.845 | Threshold cleared but regressed from 0.8625, so STOP. Design converged — seam holds from both sides, 13/13 traceability, no orphan requirement. **The regression's cause is the finding**: iteration 2 declared a sweep it did not run, and three of iter-3's defects were reachable by re-running a command the artifacts already quoted. D12-D19 closed at v1.5.0, each with the command that confirms it. |
| 5 (new cycle, iter 2) | **PASS-WITH-DEBT** | 0.8625 | Monotonic (0.75 → 0.8625), no STOP. **Fit to implement after three localized text repairs**, none touching the design or the milestone map; Traceability and Completeness verified not to have degraded under the prior repair. D9, D10, D11 closed at v1.4.0, plus two further stale twins found by sweeping rather than by audit. |
| 4 (new cycle, iter 1) | **FAIL** | 0.75 | No STOP (0.75 > 0.667). Traceability and Completeness graded excellent; the failure was concentrated in two criticals. All six blocking (D1-D6) plus both optional (D7, D8) closed at v1.3.0. T9's second layer — the "invoked unconditionally, evaluates conditionally" mutant — closed by moving AC-GDL-003's counts to the query layer. |

**The split.** The state model moved to `SPEC-GUARD-STATE-MODEL-001`, **card t347**. This SPEC keeps the surfacing model, which converged: the auditor re-ran both prior N1 mutants and could not revive either. The seam is a consumed contract, not a `depends_on` (`spec.md` §B.1), so **this artifact set is finishable on its own** — nothing in it requires a t347 artifact to exist, and it goes to the Implementation Kickoff Approval gate independently of t347's dispatch.

**§A.9 (instance 7) added at v1.1.0** — a verdict produced, correct, red, and never collected, by the orchestrator running this card. It is the live case for §A.8 and the evidence that discipline without a mechanism is insufficient.

**Iter-3 findings disposition:** T9 closed here (AC-GDL-003, the criterion the audit named most consequential). T3 dissolved structurally by the contract restatement rather than corrected as two sentences. T7, T10, T11 folded into `spec.md` §D.3. N4 taken in AC-GDL-006. T2, T4, T1, T5, T8 travel with the state model as its starting material.

## §E.2 Run-phase Evidence

### M1 — the invocation contract

**Tree.** Every measurement below was taken on `WT-guard-liveness` at parent **`a0a5b84f3`** (the merge of `origin/develop` at `d566ecc75` into this branch), in the worktree `.claude/worktrees/t333`. Pre-implementation cells were measured on that tree clean; post-implementation cells on that tree plus M1's working-tree changes, which the M1 commit carries. `plan.md` §C pins its cells to `091966c55`; all five were **re-measured here** rather than carried, and all five still hold — the baseline moved, the observations did not.

**Deliverable.** `internal/guardliveness/{contract.go,evaluator.go}` (the consumed contract + the refresh), `internal/hook/session_start_guard_liveness.go` (the invocation), and one 14-line insertion at the top of `sessionStartHandler.Handle`. No new session-start handler; no workflow file touched; no file under `internal/template/templates/`.

#### Pre-flight re-measurement (`plan.md` §C, re-run on `a0a5b84f3`)

| Cell | Command | Output |
|---|---|---|
| No advisory or evaluator wiring | `grep -rln "guard-liveness\|guardLiveness" .claude/hooks/ internal/` | no output, `rc=1` |
| No `moai guard liveness` verb | `grep -n '"guard"' internal/cli/*.go \| grep -v _test.go` | `internal/cli/constitution.go:49:\t\tUse:   "guard",` — the unrelated `constitution guard` verb only |
| Rule carries no continued-firing clause | `grep -nE "last.fired\|continued.firing\|stopped firing\|liveness\|stale guard" .claude/rules/moai/development/verification-completeness.md` | no output, `rc=1` |
| Scheduled-workflow baseline | `grep -l '^  schedule:' .github/workflows/* \| wc -l` | `3` |
| Session-start handler baseline | `grep -rn -A2 'EventType() EventType' internal/hook --include='*.go' \| grep -v _test \| grep 'EventSessionStart'` | 4 lines — `auto_update.go`, `session_start_compact.go`, `session_start.go`, `handoff_inject.go`. **4 handlers**, matching AC-GDL-010(b) in figure and definition. `session_start_binary_lag.go` landed with the merge and declares no `EventType()`, so it is not a handler. |

#### RED, before any implementation existed

```
$ go test ./internal/guardliveness/... ./internal/hook/...
github.com/modu-ai/moai-adk/internal/guardliveness: no non-test Go files in .../internal/guardliveness
# github.com/modu-ai/moai-adk/internal/guardliveness [.../guardliveness.test]
internal/guardliveness/contract_test.go:15:35: undefined: Designation
internal/guardliveness/contract_test.go:17:53: undefined: Entry
internal/guardliveness/contract_test.go:25:16: undefined: Result
internal/guardliveness/evaluator_test.go:23:59: undefined: Activation
internal/guardliveness/contract_test.go:15:57: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/guardliveness [build failed]
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
```

A build failure is a weak RED on its own — it establishes that the tests ran before the code, not that they discriminate. The mutant probes below are the strong half, and they were run against the GREEN tree.

#### AC-GDL-001 — the advisory consumes the contract and nothing more

**(a) + (c)** — two vocabularies differing in size and in value names, each carrying a non-clean entry that folds to the clean surface value:

```
$ go test -count=1 -run 'TestPartitionFiresOnExactlyTheNonCleanEntries|TestPartitionDoesNotConsultTheSurfaceFold|TestRenderNamesExactlyTheFiredEntries' -v ./internal/guardliveness/
--- PASS: TestPartitionFiresOnExactlyTheNonCleanEntries/three-value_vocabulary
--- PASS: TestPartitionFiresOnExactlyTheNonCleanEntries/two-value_vocabulary,_no_shared_value_name
--- PASS: TestPartitionDoesNotConsultTheSurfaceFold
--- PASS: TestRenderNamesExactlyTheFiredEntries
```

**Mutant (c) probe — the fold route, written and observed to fail.** `Partition` was temporarily rewritten to partition by `Entry.Surface` instead of the carried designation:

```
--- FAIL: TestPartitionFiresOnExactlyTheNonCleanEntries/three-value_vocabulary
    contract_test.go:87: fired on [subject-3], want [subject-2 subject-3] — an entry folding to the clean surface value is still non-clean
--- FAIL: TestPartitionFiresOnExactlyTheNonCleanEntries/two-value_vocabulary,_no_shared_value_name
    contract_test.go:87: fired on [], want [subject-8] — an entry folding to the clean surface value is still non-clean
--- FAIL: TestPartitionDoesNotConsultTheSurfaceFold
    contract_test.go:111: fired on [] with every surface value collapsed, want [subject-2 subject-3] — the partition read the surface fold
```

The mutant **under-fires** rather than failing loudly, exactly as `spec.md` REQ-GDL-001 clause (iii) predicts. Probe reverted; the reverted tree is what is committed.

**(b)** — the §D.2 seam instrument over the deliverable's own non-test source:

```
$ grep -rnE '\b(OK|STALE|UNKNOWN|UNDECLARED|UNREADABLE|UNRESOLVED|ORPHANED)\b' \
    internal/guardliveness/contract.go internal/guardliveness/evaluator.go \
    internal/hook/session_start_guard_liveness.go \
  | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
(no output; rc=1)
```

#### AC-GDL-003 — the evaluator is invoked on every host activation (T9)

Measured at **both** layers, on N = 5 activations, with two fixtures of opposite diff content. The package-level fixtures are real git repositories carrying an uncommitted modification (to `docs/notes.md` and to `.github/workflows/probe.yml` respectively), so a mutant reading `git diff` has genuine content to filter on.

```
$ go test -count=1 -run 'TestRefreshReachesTheQueryLayerOnEveryActivation|TestOnActivationReturnsWithoutAwaitingTheRefresh|TestUnwiredProducerIsStillReached' ./internal/guardliveness/
ok  	github.com/modu-ai/moai-adk/internal/guardliveness

$ go test -count=1 -run 'TestSessionStart_GuardLiveness' -v ./internal/hook/
--- PASS: TestSessionStart_GuardLivenessRefreshOnEveryActivation (0.58s)
    --- PASS: TestSessionStart_GuardLivenessRefreshOnEveryActivation/activation_touches_no_workflow_file (0.29s)
    --- PASS: TestSessionStart_GuardLivenessRefreshOnEveryActivation/activation_touches_workflow_files (0.28s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	1.166s
```

- **(a)** every activation yielded a refresh handle — 5 of 5, both fixtures, both layers.
- **(b)** every refresh reached the query layer. The count is taken at the producer call (`Producer.Produce`), which is the point at which subject queries are issued, not at the call site.
- **(c)** the count did not move with the fixture's diff content: 5 under both the workflow-touching and the non-workflow-touching fixture, at both layers.

**Mutant 2 probe — "invoked unconditionally, evaluates conditionally", written and observed to fail.** A subject-matter filter was temporarily inserted inside `Evaluator.Refresh`, one frame past the call site:

```
--- FAIL: TestRefreshReachesTheQueryLayerOnEveryActivation/diff_touches_no_workflow_file
    evaluator_test.go:109: query-layer arrivals = 0, want 5 — a filter sits between the activation and the subject queries
--- FAIL: TestSessionStart_GuardLivenessRefreshOnEveryActivation/activation_touches_no_workflow_file
    session_start_guard_liveness_test.go:100: query-layer arrivals = 0, want 5 — the session-start invocation is conditional
```

The mutant is `docs-i18n-check.yml` rebuilt inside the deliverable, and it dies at both layers. Probe reverted.

**Two design decisions that carry this criterion beyond the fixture**, recorded because a test measures the tree it was run on and not the tree a later edit produces:

1. `Activation` carries only `Root`. It has no changed-file list, no diff summary, and no subject-matter hint — there is nothing in the type for a condition to grow from.
2. The invocation sits at the **top** of `Handle`, before the `json.Marshal` early return several hundred lines down. Placed after it, the invocation would have been unconditional only on the activations that got that far.

There is also no early return on an empty root. "Nothing to query" is the producer's judgement; a caller-side skip would be a condition wearing a guard clause's clothes.

#### AC-GDL-004 — no scheduled watcher is introduced

**(a)** the count of workflow files carrying a `schedule:` trigger, unchanged from the measured baseline of 3:

```
$ grep -l '^  schedule:' .github/workflows/* | wc -l
       3
```

**(b)** no job reachable from a `schedule:` trigger references the evaluator — and the stronger statement holds, that nothing under `.github/` references it at all and nothing under `.github/` was touched:

```
$ grep -rniE 'guard.?liveness|guardliveness' .github/
(no output; rc=1)

$ git diff --stat a0a5b84f3 -- .github/
(empty)
```

#### Constraints asserted at merge (`plan.md` §D)

```
$ git status --short -- internal/template/templates/
(empty)
$ git status --short -- .claude/rules/moai/development/verification-completeness.md
(empty)
$ grep -rn -A2 'EventType() EventType' internal/hook --include='*.go' | grep -v _test | grep -c 'EventSessionStart'
4
```

Forge mutation and working-tree writes (REQ-GDL-008): M1's refresh path performs neither — it calls one injected producer and returns. Result persistence does not exist yet; it arrives at M2/M3, outside the working tree, where AC-GDL-008 measures it.

#### Supporting verification

```
$ go build ./...                                → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...      → exit 0
$ go vet ./...                                  → exit 0
$ go test -count=1 ./internal/guardliveness/... → ok   0.699s
$ go test -count=1 ./internal/hook/             → ok  28.438s
$ go test -count=1 -race ./internal/guardliveness/... → ok  1.901s
$ go test -count=1 -cover ./internal/guardliveness/... → coverage: 91.3% of statements
$ golangci-lint run --timeout=5m ./internal/guardliveness/... ./internal/hook/... → 0 issues.
```

The full suite was not run locally: `AGENTS.md` §4 and `CLAUDE.local.md` §4.1 scope local verification to the change and leave the full-suite verdict to CI, which runs it in a clean environment against the pushed head. **That verdict has not been read for this commit** — it is a Gap below, not a pass, and §A.9 of this SPEC is what happens when a deferral to CI is not returned to.

#### Gaps — what M1 did NOT observe

- **CI's verdict on this commit.** Not pushed at M1 (the delegation forbids it), so no full-suite, no cross-platform matrix, no `Graph Freshness`. `origin/develop` additionally carries a standing red (`spec.md` §A.9), so a future reading must separate an inherited red from a new one rather than counting a red row as this change's.
- **AC-GDL-002, 005, 006, 007, 008, 009, 010, 011, 012, 013** — not measured. They belong to M2/M3/M4 and remain RED. In particular the advisory does not yet **arrive** anywhere: `Render` exists and is exercised by tests, but nothing calls it from the session-start block. Nothing in M1 makes a liveness verdict reach an operator.
- **AC-GDL-013's five fixtures.** `Partition` refuses all five contract-violating shapes with a typed error, which is asserted here; what is NOT asserted is that an advisory **names** the violation, which is the criterion's clause (a) and is M2's.
- **The producer does not exist.** `Unwired()` is what every production activation currently reaches. The query layer is reached on schedule and answers "there is no producer" — so the count in production equals the count of activations, but no subject is actually queried by anything on this tree. The subject querying is `SPEC-GUARD-STATE-MODEL-001`'s (card t347) and is out of scope here by construction (`spec.md` §E).
- **Invocation frequency over the deliverable's lifetime.** AC-GDL-003 binds this tree's wiring at merge; nothing measures a later edit that reintroduces a filter (`acceptance.md` §D.7, unchanged).
- **The 250 ms render join bound** is neither declared nor measured — it is AC-GDL-011(c)/(d), M2.

#### Residual risk

- **The composition target is owned by an unclosed SPEC.** M1 does not yet touch the session-start additional-context block, so this risk is not live at M1 — but M2 binds to `appendAdditionalContext` and to the `sessionStartHandler` call-site pattern, both owned by card t326, whose SPEC still reads `status: in-progress`.
- **`Entry.Surface` is carried and never read.** That is deliberate and is what makes the fold mutant expressible, but a future reader may see an unused field and remove it, taking clause (c)'s fixture with it. The field's doc comment says so; nothing enforces it.
- **The unwired producer is a silence with a name.** It returns an error on every activation and M1 renders nothing, so today that error reaches a debug log and no one else. That is honest at M1 (nothing is claimed green) and becomes a defect if M2 ships a render that treats a producer error as an all-clear. AC-GDL-013 is the criterion that must catch it.
- **The invocation is now the first statement in `Handle`.** It is non-blocking in production and inline only under the test flag, so it does not spend the input-lag budget — but it is upstream of every other session-start contributor, which is a position a later panic-introducing change would make expensive. The refresh goroutine recovers from panics; the inline test path does not.

### M2 — the trigger and its arrival

**Tree.** Every measurement below was taken on `WT-guard-liveness` at **`8fa67f647`** (the M1 commit) plus M2's working-tree changes, which the M2 commit carries, in the worktree `.claude/worktrees/t333`. M1's own cells were re-run here rather than carried; all still hold.

**Deliverable.** `internal/guardliveness/{store.go,advisory.go}` (persistence + the rendered advisory), edits to `internal/guardliveness/{contract.go,evaluator.go}` (JSON tags on the consumed shapes; `Option`/`WithSink`/`Run`), `internal/hook/session_start_guard_liveness.go` (the declared join bound, the store seam, the render), and one call added beside the binary-lag contributor in `sessionStartHandler.Handle`. No new session-start handler; no workflow file touched; no file under `internal/template/templates/`; no existing text in `verification-completeness.md` modified.

#### RED, before any M2 implementation existed

```
$ go test ./internal/guardliveness/... ./internal/hook/
# github.com/modu-ai/moai-adk/internal/guardliveness [.../guardliveness.test]
internal/guardliveness/advisory_test.go:12:46: undefined: Snapshot
internal/guardliveness/advisory_test.go:42:12: undefined: Advisory
internal/guardliveness/advisory_test.go:136:31: undefined: contractViolationMarker
internal/guardliveness/advisory_test.go:136:31: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/guardliveness [build failed]
# github.com/modu-ai/moai-adk/internal/hook [.../hook.test]
internal/hook/session_start_guard_liveness_render_test.go:18:58: undefined: guardliveness.Store
internal/hook/session_start_guard_liveness_render_test.go:21:10: undefined: newGuardLivenessStore
internal/hook/session_start_guard_liveness_render_test.go:208:10: undefined: guardLivenessAdvisory
internal/hook/session_start_guard_liveness_render_test.go:220:5: undefined: guardLivenessJoinBound
internal/hook/session_start_guard_liveness_render_test.go:223:5: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
FAIL
```

As at M1, a build failure is the weak half of RED — it establishes that the tests ran before the code, not that they discriminate. The four mutant probes below are the strong half, and each was written against the GREEN tree and observed to fail.

#### AC-GDL-002 — the trigger is the partition, not a list

```
$ go test -count=1 -run 'TestAdvisoryRendersOnAnyNonCleanEntry|TestAdvisoryIsSilentOnAConformingAllCleanResult' -v ./internal/guardliveness/
--- PASS: TestAdvisoryRendersOnAnyNonCleanEntry/one_non-clean_entry_among_clean_ones
--- PASS: TestAdvisoryRendersOnAnyNonCleanEntry/every_entry_non-clean,_no_two_sharing_a_classification
--- PASS: TestAdvisoryIsSilentOnAConformingAllCleanResult
```

(a) both fixtures render, including the one where every entry is non-clean and no two share a classification. (b) the §D.2 seam instrument, re-run below over the deliverable's whole non-test source set including M2's new files, returns rc=1.

#### AC-GDL-005 — the advisory arrives with no operator input

```
$ go test -count=1 -run 'TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput' -v ./internal/hook/
--- PASS: TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput (0.12s)
```

- **(a)** the advisory reaches `HookSpecificOutput.AdditionalContext` from an ordinary `SessionStart` whose input names no guard, no workflow file, and no query.
- **(b)** stated as *inputs consumed*, so it is decidable: a second activation varying every operator-authored field (`Prompt`, `CustomInstructions`, `AgentType`) produces a byte-identical advisory. The render's only input is the root. There is no CLI verb — `moai guard liveness` does not exist and was not added.

#### AC-GDL-006 — the age is derived from the persisted result's own timestamp

```
$ go test -count=1 -run 'TestAdvisoryAgeComesFromThePersistedTimestamp|TestAdvisoryIsSilentWithoutARecordedTimestamp|TestFormatAgeResolution' -v ./internal/guardliveness/
--- PASS: TestAdvisoryAgeComesFromThePersistedTimestamp
--- PASS: TestAdvisoryIsSilentWithoutARecordedTimestamp
--- PASS: TestFormatAgeResolution
```

Two results at T₁ = now−90m and T₂ = now−30m render `1h30m` and `30m`; both non-zero, and the two advisories differ. `Advisory` takes `now` as a parameter rather than calling `time.Now` internally, so the derivation is observable rather than asserted.

**Mutant probe — the constant-offset renderer, written and observed to fail.** `age` was temporarily computed as `formatAge(time.Hour)`:

```
--- FAIL: TestAdvisoryAgeComesFromThePersistedTimestamp
    advisory_test.go:87: two results persisted at distinct times rendered the same advisory — the age is a constant, not a derivation:
        moai guard liveness (measured 1h00m ago) — 2 subject(s) not reporting clean:
          - subject-2
          - subject-3
```

The single-fixture mutant (age taken at the render moment) is killed by the same test's non-zero clauses. Probe reverted.

#### AC-GDL-010 — the advisory joins the existing block, and opens no second surface

**(a)** two halves, one behavioural and one on the deliverable's own wiring:

```
$ go test -count=1 -run 'TestGuardLivenessJoinsThroughTheExistingContributorHelper' -v ./internal/hook/
--- PASS: TestGuardLivenessJoinsThroughTheExistingContributorHelper
```

The call is `appendAdditionalContext(out, guardLivenessAdvisory(guardLivenessRoot))`, inside `sessionStartHandler.Handle`, immediately after the binary-lag contributor's identical call — the pattern the criterion mandates, measured on the composition target's own tree.

**(b)** the handler count is unchanged at the measured baseline of **4**:

```
$ grep -rn -A2 'EventType() EventType' internal/hook --include='*.go' | grep -v _test | grep -c 'EventSessionStart'
4

$ go test -count=1 -run 'TestSessionStartHandlerCountIsUnchanged' -v ./internal/hook/
--- PASS: TestSessionStartHandlerCountIsUnchanged
```

The test scans the package's own non-test source under the criterion's definition (*a type whose `EventType()` returns `EventSessionStart`*) and names the four: `auto_update.go`, `handoff_inject.go`, `session_start.go`, `session_start_compact.go`.

**Mutant probe — the second surface, written and observed to fail.** A file declaring a fifth handler was added:

```
--- FAIL: TestSessionStartHandlerCountIsUnchanged
    session_start_guard_liveness_render_test.go:177: session-start handlers = 5 [auto_update.go handoff_inject.go session_start.go session_start_compact.go zz_mutant_second_surface.go], want 4 — a second advisory surface was opened
```

**This probe earned its keep on its first run and the correction is recorded rather than quietly applied.** The first version of the scanner slid a 3-line window and tested `window[0]`, so a declaration in a file's last two lines was never examined — and the mutant, whose declaration was its final line, PASSED. The instrument was measuring less than it claimed. It now reads the whole file and takes two lines of trailing context from each match, matching the criterion's own `grep -A2`. Mutant file removed.

#### AC-GDL-011 — the render performs no forge query, under a declared bound

```
$ go test -count=1 -run 'TestGuardLivenessRenderIssuesNoForgeQuery|TestGuardLivenessRenderJoinBoundIsDeclaredAndHonoured|TestGuardLivenessRenderIsSilentWithNoPersistedResult' -v ./internal/hook/
--- PASS: TestGuardLivenessRenderIssuesNoForgeQuery
--- PASS: TestGuardLivenessRenderJoinBoundIsDeclaredAndHonoured
--- PASS: TestGuardLivenessRenderIsSilentWithNoPersistedResult
```

- **(a)** with the forge unreachable (a producer returning `context.DeadlineExceeded` on every call), the advisory still renders from the persisted result.
- **(b)** zero forge calls on the render path, counted at the producer — which IS the query layer, per M1's `Producer` seam. The count is taken while exercising the render alone, so it is not confounded by the refresh's own arrivals.
- **(c)** `const guardLivenessJoinBound = 250 * time.Millisecond`, declared in `internal/hook/session_start_guard_liveness.go` beside the join it governs, mirroring `binaryLagJoinBound`'s placement and for the reason that file records. The test asserts both that it is positive and that it is at most 250 ms, so raising it later fails the criterion rather than passing it trivially.
- **(d)** the render completes within the declared bound, measured as elapsed wall-clock around `guardLivenessAdvisory`.

**Mutant probe — the inline forge query, written and observed to fail.** A producer call was inserted into the render's read:

```
--- FAIL: TestGuardLivenessRenderIssuesNoForgeQuery
    session_start_guard_liveness_render_test.go:233: the render path issued 1 forge call(s), want 0 — session start waits on the network
```

Probe reverted.

#### AC-GDL-012 — the refresh is initiated and never awaited

```
$ go test -count=1 -run 'TestRefreshResultIsPersistedForALaterActivation|TestAFailedRefreshDoesNotOverwriteThePersistedResult' -v ./internal/guardliveness/
--- PASS: TestRefreshResultIsPersistedForALaterActivation
--- PASS: TestAFailedRefreshDoesNotOverwriteThePersistedResult
```

The fixture is the one where the two obligations conflict: subject queries held open past the bound. **(a)** the refresh is initiated and `OnActivation` returns inside the bound; **(b)** the render completes inside the bound from the PREVIOUSLY persisted result while the queries are still blocked; **(c)** when the stalled refresh completes, its result is persisted with a later timestamp and is what the next activation reads.

**Mutant probe — initiate, abandon, DISCARD, written and observed to fail.** The persistence branch was disabled:

```
--- FAIL: TestRefreshResultIsPersistedForALaterActivation
    evaluator_persist_test.go:76: persisted TakenAt = ... not after the seeded ... — the abandoned refresh never landed
```

This is the mutant clause (c) exists for: the render stays fast, the entailment looks intact, and the verdict silently stops advancing. Probe reverted.

**One decision recorded because it is a judgement, not a mechanism:** a refresh that FAILED is not persisted. Overwriting the last measured verdict with an empty result would report an all-clear about a set nothing evaluated — §A.0 at the consumer's layer — so the previous verdict stays and its disclosed age keeps growing.

#### AC-GDL-013 — a contract-violating result is reported, never rendered green

```
$ go test -count=1 -run 'TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear' -v ./internal/guardliveness/
--- PASS: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/designation_absent
--- PASS: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/designation_null
--- PASS: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/designation_multi-valued
--- PASS: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/entry_carries_no_classification
--- PASS: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/entry_carries_two_classifications

$ go test -count=1 -run 'TestSessionStart_GuardLivenessReportsAContractViolation' -v ./internal/hook/
--- PASS: TestSessionStart_GuardLivenessReportsAContractViolation
```

All five fixtures render an advisory that **names** the violation — `no clean-value designation`, `designation is null`, `more than one value`, `exactly one classification` (twice) — and none reports an all-clear. M1 got as far as `Partition` returning a typed error for these five shapes; clause (a), the advisory that names them, is closed here, and the hook-level test carries one of them the whole way to the operator's session-start block.

**Mutant probe — the all-clear by omission, written and observed to fail.** `Advisory` was made to return the empty string on a partition error:

```
--- FAIL: TestAdvisoryNamesTheContractViolationAndNeverReportsAllClear/designation_absent
    advisory_test.go:134: silence on a contract-violating result — the partition had no referent and nothing rendered, which is an all-clear by omission
```

All five sub-cases failed identically. Probe reverted.

#### Preserved invariants (`plan.md` §D, re-measured on this tree)

```
$ grep -rnE '\b(OK|STALE|UNKNOWN|UNDECLARED|UNREADABLE|UNRESOLVED|ORPHANED)\b' \
    internal/guardliveness/contract.go internal/guardliveness/evaluator.go \
    internal/guardliveness/store.go internal/guardliveness/advisory.go \
    internal/hook/session_start_guard_liveness.go \
  | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
(no output; rc=1)

$ grep -l '^  schedule:' .github/workflows/* | wc -l
       3

$ grep -rniE 'guard.?liveness|guardliveness' .github/
(no output; rc=1)

$ git diff --stat 8fa67f647 -- .github/
(empty)

$ git status --short -- internal/template/templates/
(empty)
$ git status --short -- .claude/rules/moai/development/verification-completeness.md
(empty)
```

The §D.2 instrument now covers **five** non-test source files — M1's three plus `store.go` and `advisory.go` — and still returns rc=1.

REQ-GDL-008: the persistence lives at `~/.moai/state/guard-liveness/` (resolved through `internal/paths`), outside every evaluated working tree. Asserted three ways: `TestStoreWritesNothingIntoTheEvaluatedRoot` walks the evaluated root before and after a save and requires it unchanged; `TestDefaultStoreResolvesUnderTheMoaiStateTree` locates the written file under a throwaway `MOAI_HOME`; and the render path issues no forge call at all (AC-GDL-011(b)), so there is no mutation to count.

#### M1 non-regression

```
$ go test -count=1 -run 'TestPartitionFiresOnExactlyTheNonCleanEntries|TestPartitionDoesNotConsultTheSurfaceFold|TestRenderNamesExactlyTheFiredEntries|TestPartitionRefusesAContractViolatingResult|TestRefreshReachesTheQueryLayerOnEveryActivation|TestOnActivationReturnsWithoutAwaitingTheRefresh|TestUnwiredProducerIsStillReached' -v ./internal/guardliveness/ | grep -c -E '^( +)?--- PASS'
16

$ go test -count=1 -run 'TestSessionStart_GuardLivenessRefreshOnEveryActivation' -v ./internal/hook/
--- PASS: TestSessionStart_GuardLivenessRefreshOnEveryActivation/activation_touches_no_workflow_file (0.29s)
--- PASS: TestSessionStart_GuardLivenessRefreshOnEveryActivation/activation_touches_workflow_files (0.29s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.068s
```

AC-GDL-001, 003 and 004 hold unchanged. `New` gained variadic options rather than a second parameter precisely so every M1 call site and every M1 test compiles untouched — the non-regression is structural, not a re-verification of edited tests. The invocation is still the first statement in `Handle`, still takes no changed-file input, and still has no branch between the activation and the producer call: M2 added a call at the END of `Handle` and put nothing between the activation and the query layer.

#### Supporting verification

```
$ go build ./...                                       → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...             → exit 0
$ GOOS=windows GOARCH=amd64 go vet ./internal/guardliveness/... ./internal/hook/  → exit 0
$ go vet ./...                                         → exit 0
$ gofmt -l internal/guardliveness internal/hook/session_start_guard_liveness.go \
      internal/hook/session_start_guard_liveness_render_test.go internal/hook/session_start.go
(no output)
$ go test -count=1 ./internal/guardliveness/...        → ok   2.919s
$ go test -count=1 ./internal/hook/                    → ok  30.850s
$ go test -count=1 -race ./internal/guardliveness/...  → ok   1.883s
$ go test -count=1 -cover ./internal/guardliveness/... → coverage: 87.1% of statements
$ golangci-lint run --timeout=6m ./internal/guardliveness/... ./internal/hook/... → 0 issues.
```

Coverage first read 80.2%; the shortfall was in the degradation paths (`DefaultStore`, the `Save`/`Load` error branches, `formatAge`'s resolution boundaries), which are now covered directly rather than by raising the ceiling.

The full suite was not run locally: `AGENTS.md` §4 and `CLAUDE.local.md` §4.1 scope local verification to the change and leave the full-suite verdict to CI. **That verdict has not been read for this commit** — it is a Gap below, not a pass.

#### Gaps — what M2 did NOT observe

- **CI's verdict on this commit.** Not pushed at M2 (the delegation forbids it), so no full suite, no cross-platform matrix, no `Graph Freshness`. `origin/develop` additionally carries a standing red (`spec.md` §A.9), so a future reading must separate an inherited red from a new one rather than counting a red row as this change's.
- **AC-GDL-007, AC-GDL-008, AC-GDL-009** — not measured. Change-leading and the standing count are M3; the doctrine clause is M4. In particular the advisory today re-renders the full non-clean list every session, which is the noise profile REQ-GDL-007 exists to fix and which `spec.md` §A.8 says is how a channel gets filtered.
- **AC-GDL-008's own measurement.** REQ-GDL-008's substance is asserted here at the unit layer (the evaluated root is byte-identical across a save; no forge call on the render path), but the criterion's own `git status --porcelain` across a render, and its mutating-call count at the call layer, belong to M3 and were not run.
- **The producer still does not exist.** `Unwired()` is what every production activation reaches, so on a real tree the store is never written and the render is silent. Every rendering assertion above is against a seeded or stubbed result. Nothing here demonstrates a liveness verdict about an actual workflow, because nothing yet produces one — that is `SPEC-GUARD-STATE-MODEL-001`'s (card t347).
- **The one-activation-stale magnitude.** AC-GDL-012 shows the mechanism; nothing bounds how old a persisted verdict can get on a surface visited rarely (`acceptance.md` §D.7, unchanged).
- **The async render path.** Tests run with `deferredScansAsync=false`, so the timer/goroutine branch of `guardLivenessAdvisory` — the one that actually enforces the bound in production — is not exercised. The bound is asserted as a declared value and as measured elapsed time on the inline path; the abandonment behaviour at the bound is not.

#### Residual risk

- **The composition target is owned by an unclosed SPEC, and M2 is where that binds.** M1 recorded this as not-yet-live; it is live now. The advisory calls `appendAdditionalContext` and sits beside `binaryLagAdvisory` in `Handle`, both owned by card t326, whose SPEC still reads `status: in-progress`.
- **The advisory renders the full standing list.** Deliberate at M2 and a defect at M3: beside an always-red neighbour it inherits the filter §A.8 describes, and REQ-GDL-007 is the mitigation that has not landed.
- **Silence covers two outcomes that a reader cannot tell apart.** No verdict persisted yet, and every subject clean, both render nothing. That is honest — the second is the designed all-clear and the first is not a claim at all — but an operator seeing nothing cannot distinguish "the mechanism is working and content" from "the mechanism has never completed a refresh". A verdict that could not be READ is the third case and does speak.
- **A store that cannot be resolved degrades to silence.** The refresh still runs and still reaches the query layer; only its carriage is lost, and the render reports an absent verdict rather than an all-clear. The degradation is logged at debug level, which `verification-completeness.md` §1.2(c) names as a reachability that nobody observes.
- **The test path and the production path differ at the render.** Under `deferredScansAsync=false` the read is inline; in production it runs behind a timer. The two share `read()`, so the logic is common, but the abandonment path is production-only.

### M3 — legibility and safety

**Tree.** Every measurement below was taken on `WT-guard-liveness` at **`24ecc7e65`** (the M2 commit) plus M3's working-tree changes, which the M3 commit carries, in the worktree `.claude/worktrees/t333`. M1's and M2's invariant cells were re-run here rather than carried; all still hold.

**Deliverable.** `internal/guardliveness/advisory.go` (the `RenderRecord` shape, the change/standing split, the new `Advisory` signature), `internal/guardliveness/store.go` (`SaveRendered` / `LoadRendered` on a second per-root file, and the atomic write factored out of `Save`), `internal/hook/session_start_guard_liveness.go` (the render loads the previous record, diffs against it, and persists the new one). No new session-start handler; no workflow file touched; no file under `internal/template/templates/`; no existing text in `verification-completeness.md` modified.

**One decision worth stating, because the criterion does not force it.** The comparison subject is the previously **RENDERED** state, not the previously persisted result. A verdict nobody was shown is not one a reader can be assumed to know, so a refresh that completes and is never rendered does not silence the entry it found. The record is therefore written by the render, into a file of its own beside the verdict.

#### RED, before any M3 implementation existed

```
$ go test ./internal/guardliveness/... ./internal/hook/
# github.com/modu-ai/moai-adk/internal/guardliveness [.../guardliveness.test]
internal/guardliveness/advisory_change_test.go:50:20: assignment mismatch: 2 variables but Advisory returns 1 value
internal/guardliveness/advisory_change_test.go:50:71: undefined: RenderRecord
internal/guardliveness/advisory_change_test.go:50:87: too many arguments in call to Advisory
	have (Snapshot, unknown type, "time".Time)
	want (Snapshot, "time".Time)
internal/guardliveness/advisory_change_test.go:60:16: assignment mismatch: 2 variables but Advisory returns 1 value
internal/guardliveness/advisory_change_test.go:89:15: assignment mismatch: 2 variables but Advisory returns 1 value
internal/guardliveness/advisory_change_test.go:105:81: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/guardliveness [build failed]
# github.com/modu-ai/moai-adk/internal/hook [.../hook.test]
internal/hook/session_start_guard_liveness_mutation_test.go:132:20: store.LoadRendered undefined (type *guardliveness.Store has no field or method LoadRendered)
FAIL	github.com/modu-ai/moai-adk/internal/hook [build failed]
FAIL
```

As at M1 and M2, a build failure is the weak half of RED — it establishes that the tests ran before the code, not that they discriminate. The four mutant probes below are the strong half, and each was written against the GREEN tree and observed to fail.

#### AC-GDL-007 — the advisory leads with changes and survives a standing non-clean neighbour

```
$ go test -count=1 -run 'TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries|TestAdvisoryFirstRenderTreatsEveryNonCleanEntryAsChanged|TestAdvisoryRendersAStandingCountWhenNothingChanged|TestAdvisoryTreatsAReclassifiedEntryAsChanged|TestAdvisoryRecordsNothingForAContractViolatingResult|TestAdvisoryRecordsNothingWithoutARecordedTimestamp|TestAdvisoryRecordsAnAllCleanResultSoARelapseIsAnnounced' -v ./internal/guardliveness/
--- PASS: TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries (0.00s)
--- PASS: TestAdvisoryFirstRenderTreatsEveryNonCleanEntryAsChanged (0.00s)
--- PASS: TestAdvisoryRendersAStandingCountWhenNothingChanged (0.00s)
--- PASS: TestAdvisoryTreatsAReclassifiedEntryAsChanged (0.00s)
--- PASS: TestAdvisoryRecordsNothingForAContractViolatingResult (0.00s)
--- PASS: TestAdvisoryRecordsNothingWithoutARecordedTimestamp (0.00s)
--- PASS: TestAdvisoryRecordsAnAllCleanResultSoARelapseIsAnnounced (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/guardliveness	0.160s
```

The criterion's own fixture is `TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries`: S = {`subject-2`, `subject-3`} is non-clean in both evaluations and T = `subject-1` newly became non-clean in the second. (a) the first `- ` line names T. (b) neither member of S appears anywhere in the second render's text, and the count `2` does.

**Mutant probe (a) — the renderer the criterion's own mutant note names: print the full standing list with T at the top.** The standing count line was replaced by a loop over `standing`, and the criterion failed:

```
--- FAIL: TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries (0.00s)
    advisory_change_test.go:76: standing subject "subject-2" is re-rendered individually instead of being counted:
        moai guard liveness (measured 5m ago) — 1 subject(s) changed:
          - subject-1
          - subject-2
          - subject-3
--- FAIL: TestAdvisoryTreatsAReclassifiedEntryAsChanged (0.00s)
    advisory_change_test.go:147: the unchanged subject is re-rendered individually:
```

**Mutant probe (b) — the opposite direction: nothing ever counts as changed.** `splitByChange`'s predicate was disabled (`if false && ...`), so every non-clean entry became standing and the advisory collapsed to a count with no leading entry:

```
--- FAIL: TestAdvisoryLeadsWithChangesAndCollapsesStandingEntries (0.00s)
    advisory_change_test.go:56: the first render omits "subject-2", so it never announced what the second must treat as standing:
        moai guard liveness (measured 2h00m ago) — 2 subject(s) still not reporting clean, unchanged since the last session.
--- FAIL: TestAdvisoryFirstRenderTreatsEveryNonCleanEntryAsChanged (0.00s)
    advisory_change_test.go:92: first render omits non-clean subject "subject-2":
```

This mutant is not in the criterion's text, and it is the one a change-leading design is most likely to ship by accident: a renderer so quiet it never announces anything is trivially compact.

#### AC-GDL-008 — the advisory path mutates nothing

**(a) zero mutating calls, counted at the call layer** over the deliverable's own five non-test source files. The count is an AST walk plus an import-set check rather than a text grep, so an aliased selector cannot slip past it:

```
$ go test -count=1 -run 'TestAdvisoryPathIssuesNoMutatingCall|TestRenderRecordRoundTripsThroughTheStore|TestLoadRenderedIsEmptyBeforeAnyRender|TestRenderRecordDoesNotOverwriteThePersistedResult' -v ./internal/guardliveness/
--- PASS: TestAdvisoryPathIssuesNoMutatingCall (0.00s)
--- PASS: TestRenderRecordRoundTripsThroughTheStore (0.00s)
--- PASS: TestLoadRenderedIsEmptyBeforeAnyRender (0.00s)
--- PASS: TestRenderRecordDoesNotOverwriteThePersistedResult (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/guardliveness	0.176s
```

**(b) the working tree is byte-identical across a render**, verified by `git status --porcelain` in a throwaway repository whose porcelain baseline is deliberately NON-empty (one modified tracked file, one untracked file) — an empty baseline compares equal to an empty result even when a write happened:

```
$ go test -count=1 -run 'TestGuardLivenessRenderLeavesTheWorkingTreeByteIdentical|TestGuardLivenessRenderPersistsItsRecordOutsideTheTree' -v ./internal/hook/
--- PASS: TestGuardLivenessRenderLeavesTheWorkingTreeByteIdentical (0.09s)
--- PASS: TestGuardLivenessRenderPersistsItsRecordOutsideTheTree (0.05s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.575s
```

**Mutant probe — the criterion's own: a renderer that writes its cache into the working tree.** One `os.WriteFile` into the evaluated root was added to the render path. Clause (a) passed it — `os.WriteFile` is not a forge call, and the store legitimately uses it outside the tree — and clause (b) caught it, which is exactly the division the criterion's mutant note predicts:

```
--- FAIL: TestGuardLivenessRenderLeavesTheWorkingTreeByteIdentical (0.09s)
    session_start_guard_liveness_mutation_test.go:95: the render changed the working tree.
        before:
         M tracked.txt
        ?? untracked.txt

        after:
         M tracked.txt
        ?? .moai-guard-liveness-cache
        ?? untracked.txt
```

**Mutant probe — the other direction: a forge mutation with no file write.** The same line was replaced with `exec.Command("gh", "issue", "create", ...)`. Clause (b) would have passed it (no file changed); clause (a) caught it at both the import and the call:

```
--- FAIL: TestAdvisoryPathIssuesNoMutatingCall (0.00s)
    mutation_test.go:92: the advisory path carries 2 mutating call site(s), want 0:
        ../hook/session_start_guard_liveness.go: imports os/exec
        ../hook/session_start_guard_liveness.go:156: calls Command
```

**The persistence lives outside the working tree**, which the criterion names as its own consequence rather than an oversight. Both per-root files resolve under `paths.StateDir()` → `~/.moai/state/guard-liveness/`, and `TestGuardLivenessRenderPersistsItsRecordOutsideTheTree` asserts the record directory is not a prefix-child of the evaluated root.

#### Invariants held (re-measured on this tree, not carried)

```
$ grep -rnE '\b(OK|STALE|UNKNOWN|UNDECLARED|UNREADABLE|UNRESOLVED|ORPHANED)\b' \
    internal/guardliveness/contract.go internal/guardliveness/evaluator.go \
    internal/guardliveness/store.go internal/guardliveness/advisory.go \
    internal/hook/session_start_guard_liveness.go \
  | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
(no output; rc=1)

$ grep -l '^  schedule:' .github/workflows/* | wc -l
       3

$ grep -rn -A2 'EventType() EventType' internal/hook --include='*.go' | grep -v _test | grep -c 'EventSessionStart'
4
```

The §D.2 instrument runs over the same five files it did at M2 — M3 added no source file, only edits — and still returns rc=1. The schedule baseline and the session-start handler count are unchanged.

#### Build, tests, coverage, vet

```
$ go build ./...                                          → exit 0
$ go test -count=1 ./internal/guardliveness/... ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/guardliveness	0.770s
ok  	github.com/modu-ai/moai-adk/internal/hook	30.085s

$ go test -count=1 -cover ./internal/guardliveness/...
ok  	github.com/modu-ai/moai-adk/internal/guardliveness	0.938s	coverage: 88.2% of statements

$ go vet ./internal/guardliveness/... ./internal/hook/    → exit 0
$ gofmt -l internal/guardliveness                          → no output
```

Coverage moved 87.1% → 88.2%. The full suite was not run locally: `AGENTS.md` §4 and `CLAUDE.local.md` §4.1 scope local verification to the change and leave the full-suite verdict to CI. **That verdict has not been read for this commit** — it is a Gap below, not a pass.

#### One existing test was amended, and the amendment is stated rather than absorbed

`TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput` (AC-GDL-005, landed at M2) drives `Handle` twice and asserts the two advisories are byte-identical while operator input varies between them. Change-leading breaks the second premise honestly: the second render now differs because it is the *second*, not because of the operator input. The test resets the render record between the two runs, restoring operator input as the only variable — which is the property AC-GDL-005 actually asserts. **The assertion was not weakened**; a confound introduced by M3 was removed:

```
$ go test -count=1 -run TestSessionStart_GuardLivenessAdvisoryArrivesWithNoOperatorInput ./internal/hook/
ok  	github.com/modu-ai/moai-adk/internal/hook
```

#### Gaps — what M3 did NOT observe

- **CI's verdict on this commit.** Not pushed at M3 (the delegation forbids it), so no full suite, no cross-platform matrix, no `Graph Freshness`. `origin/develop` additionally carries a standing red (`spec.md` §A.9), so a future reading must separate an inherited red from a new one rather than counting a red row as this change's.
- **AC-GDL-009** — not measured at M3. The doctrine clause is M4.
- **The producer still does not exist.** `Unwired()` is what every production activation reaches, so on a real tree the store is never written and the render is silent. Every rendering assertion above is against a seeded result. Nothing here demonstrates change-leading over an actual workflow's liveness, because nothing yet produces one — that is `SPEC-GUARD-STATE-MODEL-001`'s (card t347).
- **The async render path.** Tests run with `deferredScansAsync=false`, so the timer/goroutine branch of `guardLivenessAdvisory` is not exercised. The record is written inside that branch too, and the abandonment behaviour at the bound is not measured.
- **Whether a compact standing count is itself skippable.** AC-GDL-007's own bounded claim: this asserts change-leading, not that the advisory gets read. Nothing here measures a reader.
- **`golangci-lint` was not run at M3.** `go vet` and `gofmt` were; the lint gate is CI's.

#### Residual risk

- **A render abandoned at the join bound still records.** The record is written inside `read()`, which runs in a goroutine the timer may abandon. When the bound elapses the operator sees nothing and the entry is nonetheless marked announced, so it appears as a standing count on the next session. The window is a local file read against a 250 ms ceiling, so it is narrow; it is not zero, and its failure direction is silence, which is this card's own subject.
- **A resolved entry is never announced.** An entry returning to the clean value has changed, and the advisory does not say so: REQ-GDL-004 fires only on a non-clean entry, and rendering recoveries would grow the block change-leading exists to shrink. Deliberate, and it means the advisory reports deterioration only.
- **The change-leading trade `spec.md` §D.3 names is now real code.** A long-standing non-clean entry is announced once and thereafter only counted. Quiet is this card's subject, and this is the deliberate quiet it accepts.
- **The record and the verdict can disagree about what exists.** They are separate files with no cross-check: a store where the verdict was written and the record was not — or the reverse, after a partial failure — yields a render that treats entries as new or as standing on the strength of a file that does not describe the same measurement. Each write is atomic individually; the pair is not.
- **`Render` in `contract.go` is now a second renderer with no caller.** It predates `Advisory`, is still exercised by M1's contract tests, and re-renders the full non-clean list. Nothing stops a later caller from reaching for it and reintroducing the noise profile M3 removes.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
