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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
