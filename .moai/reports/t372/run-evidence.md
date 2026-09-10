# SPEC-STRESS-INVARIANT-VERDICT-001 — Run-phase Evidence

Card t372 · worktree `.claude/worktrees/t372` · branch `WT-stress-invariant-guard` ·
base `origin/develop` = `b9149857c` · run-phase start HEAD = `3cd1a09f1`.

Every figure below was measured **in this run, against this tree**. The t370 measurements
(`.moai/reports/t370/verdict.md`) are consumed as given ground truth and re-measured nowhere.

---

## 1. Claim

The stress test's verdict criterion has been moved off lock acquisition and onto the four queue
invariants; acquisition latency now has its own constant-coherence guard. Both criteria were
demonstrated capable of turning RED under a planted mutant, and both mutants were reverted with
their restoring GREEN recorded.

## 2. Per-AC matrix

| AC | Status | Verification | Observed |
|---|---|---|---|
| AC-SIV-001 | PASS | `go test -race -count=1 -v -run '…' ./internal/kanban/` | `--- PASS: TestStressAddClassificationToleratesStarvation (3.37s)`; log `2/2 adds starved under a seeded holder, all satisfying IsBoardLockHeld, 0 hard failures` |
| AC-SIV-002 | PASS | source read + `grep -n 'strings.Contains' internal/kanban/backlog_concurrency_test.go` | classification is `classifyStressAdd`, deciding on `err == nil` / `IsBoardLockHeld(err)` only; the sole `strings.Contains` hit is line 343, inside the AC-SIV-005 sub-test asserting the zero-progress **message wording** — not in the classification path |
| AC-SIV-003 | PASS | source read (`backlog_concurrency_test.go:205-231`) | the four assertions compare against `issuedCount := len(issued)`; `wantTotal` no longer exists in the file |
| AC-SIV-004 | PASS | `grep -n 't.Skip' internal/kanban/backlog_concurrency_test.go` | one hit, line 202, inside a **comment**; no `t.Skip` call and no starvation conditional guards any of the four |
| AC-SIV-005 | PASS | `go test -race -count=1 -v -run '…' ./internal/kanban/` | `--- PASS: TestStressZeroProgressFloorFailsTotalStarvation (1.68s)`; log `zero-success outcome rejected ("0 of 1 attempted adds succeeded (starved=1, hard failures=0) — total starvation is a broken lock, not tolerable contention"); 1-success outcome admitted` |
| AC-SIV-014 | PASS | source read + the package `-race` run below | `successes + starved + len(hardFailures) == stressWriters * stressAddsPerWriter`; no wall clock, no fraction, no percentage threshold in the assertion |
| AC-SIV-006 | PASS | `go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` | guard asserts `boardLockWaitBudget >= time.Duration(stressWriters*stressAddsPerWriter) * boardLockCIMutationCost`; `stressWriters` / `stressAddsPerWriter` have one definition (`backlog_concurrency_test.go:24-27`) consumed by `TestConcurrencyStress`; messages claim coherence only (verbatim in §4) |
| AC-SIV-007 | PASS | `sed -n '95,120p' internal/kanban/board_lock_wait_test.go \| grep -nE 'time\.(Now\|Since\|Sleep)\|go func'` → **no output** | floor is `time.Duration(stressWriters*stressAddsPerWriter) * boardLockCIMutationCost`; the two stress constants are not supplied by the budget expression, so the comparison is not a value against itself |
| **AC-SIV-008** | **PASS — binding** | §5 below | new guard RED, old guard GREEN, same run; reverted; restoring GREEN |
| **AC-SIV-009** | **PASS — binding** | §6 below | `TestConcurrencyStress` RED at invariant (d) and, on the second mutant, at (b)+(c); both reverted; restoring GREEN |
| AC-SIV-010 | PASS | the `TestConcurrencyStress` log line in §3 | starved count and back-derived per-mutation cost logged via `t.Logf`; no `t.Error`/`t.Fatal` is gated on either (the only starvation-related fatal is the AC-SIV-005 zero-progress floor) |
| AC-SIV-011 | PASS | §8 below | all four REQ-SIV-014 limits stated |
| AC-SIV-012 | PASS | §7 below | 3 changed paths; `board_store.go` change is comment-only; both mutants reverted before the diff |
| AC-SIV-013 | **OPEN at merge** | — | requires ≥5 post-landing develop heads; deliberately not claimed (§8 clause 2) |

## 3. Baseline — unmutated tree

```
$ go test -race -count=1 -v -run 'TestConcurrencyStress|TestBoardLockWaitBudget|TestStressAddClassificationToleratesStarvation|TestStressZeroProgressFloorFailsTotalStarvation' ./internal/kanban/
=== RUN   TestConcurrencyStress
=== RUN   TestStressAddClassificationToleratesStarvation
=== RUN   TestStressZeroProgressFloorFailsTotalStarvation
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== RUN   TestBoardLockWaitBudgetCoversSerializedMutations
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
    backlog_concurrency_test.go:235: SPEC-STRESS-INVARIANT-VERDICT-001: 8 writers x 6 adds = 48 attempts; 48 succeeded, 0 starved (tolerated), 0 hard failures; 48 distinct ids, 48 stored items, last_seq 48; back-derived per-mutation cost 15.696085ms (elapsed 753.412125ms / 48 successful mutations)
--- PASS: TestConcurrencyStress (0.76s)
--- PASS: TestStressZeroProgressFloorFailsTotalStarvation (1.68s)
--- PASS: TestStressAddClassificationToleratesStarvation (3.37s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	4.887s
```

**Selector match count: 5** (five `=== RUN` lines), non-zero.

This run's own starved count is **0** and its back-derived per-mutation cost is **15.696085ms**
(local darwin `-race`). That is a machine measurement, not a verdict input.

Package suite, unmutated:

```
$ go test -race -count=1 ./internal/kanban/
ok  	github.com/modu-ai/moai-adk/internal/kanban	24.890s
```

```
$ go vet ./internal/kanban/...
(no output — exit 0)
```

```
$ gofmt -l internal/kanban/
(no output)
```

## 4. AC-SIV-006 — the guard's messages, verbatim

Success (`t.Logf`), from the baseline run:

> constant coherence: the lock policy budgets 10 supported writers x 5 headroom = 50 serialized mutations; the stress test serializes 8 x 6 = 48. The per-mutation cost cancels on both sides, so this relation is cost-independent — it states nothing about the wait any real machine needs, and a change to boardLockCIMutationCost (declared 33ms) would not be caught here.

Failure (`t.Fatalf`), from the §5 RED run — reproduced there verbatim. Neither message claims the
budget suffices on any machine, and neither implies the declared 33ms conditions the verdict.

## 5. AC-SIV-008 — latency direction (binding)

### 5.1 Census, taken BEFORE the mutant was planted

```
$ grep -rn 'boardLockHeadroom' internal/kanban/
internal/kanban/board_store.go:109:	// boardLockHeadroom is the stated headroom factor over the product
internal/kanban/board_store.go:113:	// The product boardLockSupportedWriters * boardLockHeadroom = 50 is the
internal/kanban/board_store.go:121:	boardLockHeadroom = 5
internal/kanban/board_store.go:126:	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
internal/kanban/integration_lock_cross_test.go:55:	// × boardLockHeadroom = 10 × 33ms × 5 = 1.65s (board_store.go:96-117).
internal/kanban/board_lock_wait_test.go:27:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:32:			boardLockHeadroom, recomputed)
internal/kanban/board_lock_wait_test.go:38:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:59:	if boardLockHeadroom < 2 {
internal/kanban/board_lock_wait_test.go:60:		t.Errorf("headroom factor %d states no headroom", boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:73://	boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter
internal/kanban/board_lock_wait_test.go:77:// lowering boardLockSupportedWriters or boardLockHeadroom, or raising
internal/kanban/board_lock_wait_test.go:92:// (boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:99:	const policyBudgetedMutations = boardLockSupportedWriters * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:109:			boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,
internal/kanban/board_lock_wait_test.go:118:		boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,

$ grep -rln 'boardLockHeadroom\|boardLockWaitBudget' --include='*_test.go' internal/
internal/kanban/backlog_concurrency_test.go
internal/kanban/board_lock_wait_test.go
internal/kanban/integration_lock_cross_test.go
```

Three test files cover the mutated constant: `board_lock_wait_test.go` (both budget guards +
`TestBacklogLockStuckHolderSurfacesBoundedNamedError`, which uses the budget as a bound);
`integration_lock_cross_test.go` (a doc comment sizing its 500ms release timeout against the
budget); `backlog_concurrency_test.go` (a doc comment only). The census is what makes §5.4's
"only one test went RED" an observation about a bounded set rather than an assumption.

### 5.2 Mutant diff (constant axis — the qualifying shape)

```diff
-	boardLockHeadroom = 5
+	boardLockHeadroom = 4
```

Budget 1.65s → 1.32s, below the 48 × 33ms = 1.584s floor. This is the constant axis. A
**cost-axis** mutant (`boardLockCIMutationCost`) is NOT a qualifying shape and cannot make this
guard fire at all — the term appears on both sides and cancels; none was planted.

### 5.3 RED, verbatim, with the old guard GREEN in the same run

```
$ go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== RUN   TestBoardLockWaitBudgetCoversSerializedMutations
=== PAUSE TestBoardLockWaitBudgetCoversSerializedMutations
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
=== CONT  TestBoardLockWaitBudgetCoversSerializedMutations
    board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 10 supported writers x 4 headroom = 40 serialized mutations, while the stress test serializes 8 x 6 = 48 (1.32s budget < 1.584s floor). Lowering either policy constant, or raising either stress constant past that product, fails this guard. The per-mutation cost cancels on both sides, so the relation is cost-independent and asserts nothing about the wait any real machine needs — the CI -race per-mutation cost observed by t370 was 42-105ms against the declared 33ms.
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.461s
FAIL
```

- **Test that went RED, by name**: `TestBoardLockWaitBudgetCoversSerializedMutations` (the new guard).
- **`-run` selector match count: 2** (two `=== RUN` lines), non-zero.
- **`TestBoardLockWaitBudgetDerivedFromNamedInputs` stayed GREEN** under the same mutant —
  `--- PASS`, same run. This is what attributes the RED to the new guard alone.

### 5.4 Whole-package run under the same mutant

```
$ go test -count=1 ./internal/kanban/
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	15.405s
FAIL
```

Exactly one test in the package went RED — no other member of the §5.1 census fired.

### 5.5 Revert and restoring GREEN

```
$ grep -n 'boardLockHeadroom = ' internal/kanban/board_store.go
113:	// The product boardLockSupportedWriters * boardLockHeadroom = 50 is the
121:	boardLockHeadroom = 5

$ go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== RUN   TestBoardLockWaitBudgetCoversSerializedMutations
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.453s
```

## 6. AC-SIV-009 — invariant direction (binding)

### 6.1 Census, taken BEFORE the mutants were planted

```
$ grep -rln 'LastSeq\|\.Add(' --include='*_test.go' internal/kanban/
internal/kanban/board_lock_cross_test.go
internal/kanban/board_store_test.go
internal/kanban/state_dir_test.go
internal/kanban/backlog_migrate_test.go
internal/kanban/backlog_counts_test.go
internal/kanban/backlog_store_test.go
internal/kanban/backlog_concurrency_test.go
internal/kanban/backlog_archive_test.go
internal/kanban/kanban_helper_test.go
internal/kanban/integration_lock_cross_test.go
internal/kanban/board_lock_wait_test.go
internal/kanban/backlog_store_errors_test.go
```

Twelve test files cover `(*BacklogStore).Add` / `LastSeq` — a broad census, which is precisely why
attribution here rests on two further observations rather than on the RED alone: the mutants are
**scale-gated** so they engage only at the stress test's 48-add fan-out, and the whole-package run
under each mutant (§6.2.3, §6.3.2) shows which tests actually fired.

### 6.2 Mutant 1 — upward `last_seq` advance above the item count

The excluded shapes were not attempted: a **duplicate-id** mutant is rejected by
`id TEXT NOT NULL UNIQUE` (`backlog_sqlite.go:109`) and would trip the hard-failure gate; a
**downward** `last_seq` mutation is erased by `normalizeBacklogRecord`.

#### 6.2.1 Diff (`internal/kanban/backlog_store.go`, inside `Add`'s mutation callback)

```diff
 		rec.Items = append(rec.Items, item)
+		if len(rec.Items) > 40 {
+			rec.LastSeq++ // AC-SIV-009 MUTANT: upward last_seq advance above the item count
+		}
 		pos = 0
```

#### 6.2.2 RED, verbatim

```
$ go test -race -count=1 -v -run TestConcurrencyStress ./internal/kanban/
=== RUN   TestConcurrencyStress
=== PAUSE TestConcurrencyStress
=== CONT  TestConcurrencyStress
    backlog_concurrency_test.go:228: invariant (d) mark consistency: last_seq = 56, want 48 (distinct issued ids)
    backlog_concurrency_test.go:235: SPEC-STRESS-INVARIANT-VERDICT-001: 8 writers x 6 adds = 48 attempts; 48 succeeded, 0 starved (tolerated), 0 hard failures; 48 distinct ids, 48 stored items, last_seq 56; back-derived per-mutation cost 15.44176ms (elapsed 741.2045ms / 48 successful mutations)
--- FAIL: TestConcurrencyStress (0.75s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	1.267s
FAIL
```

- **Test that went RED, by name**: `TestConcurrencyStress`.
- **Which of the four**: invariant **(d) mark consistency**.
- **Assertion message**: `invariant (d) mark consistency: last_seq = 56, want 48 (distinct issued ids)`.
- **Source line**: `internal/kanban/backlog_concurrency_test.go:228`.
- **`-run` selector match count: 1** (one `=== RUN` line), non-zero.
- The RED did **not** come from `store.Load()`, the hard-failure gate, the zero-progress floor, the
  conservation assertion, a storage-layer rejection, a DATA RACE, or a panic — the log line shows
  48 successes, 0 starved, 0 hard failures, and conservation intact.

#### 6.2.3 Whole-package run under mutant 1

```
$ go test -count=1 ./internal/kanban/
--- FAIL: TestConcurrencyStress (0.21s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	15.802s
FAIL
```

Exactly one test of the twelve-file census fired.

### 6.3 Mutant 2 — dropped item after its id was issued (lost update)

Planted after mutant 1 was reverted; both directions strengthen the evidence.

#### 6.3.1 Diff

```diff
 		rec.Items = append(rec.Items, item)
+		if len(rec.Items) == 45 {
+			rec.Items = rec.Items[1:] // AC-SIV-009 MUTANT 2: drop an item after its id was issued (lost update)
+		}
 		pos = 0
```

#### 6.3.2 RED, verbatim

```
$ go test -race -count=1 -v -run TestConcurrencyStress ./internal/kanban/
=== RUN   TestConcurrencyStress
=== PAUSE TestConcurrencyStress
=== CONT  TestConcurrencyStress
    backlog_concurrency_test.go:218: invariant (b) no lost update: id t2 was issued but is not in the queue
    backlog_concurrency_test.go:218: invariant (b) no lost update: id t1 was issued but is not in the queue
    backlog_concurrency_test.go:218: invariant (b) no lost update: id t3 was issued but is not in the queue
    backlog_concurrency_test.go:223: invariant (c) count consistency: stored items = 44, want 47 (distinct issued ids) — 3 updates were lost
    backlog_concurrency_test.go:235: SPEC-STRESS-INVARIANT-VERDICT-001: 8 writers x 6 adds = 48 attempts; 47 succeeded, 1 starved (tolerated), 0 hard failures; 47 distinct ids, 44 stored items, last_seq 47; back-derived per-mutation cost 48.756859ms (elapsed 2.291572417s / 47 successful mutations)
--- FAIL: TestConcurrencyStress (2.33s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	3.066s
FAIL
```

- **Which of the four**: invariants **(b) no lost update** (line 218) and **(c) count consistency**
  (line 223).
- **`-run` selector match count: 1**, non-zero.
- **This run is the strongest single observation in the card.** It happened to produce a genuine
  starved add (`1 starved (tolerated)`), and the invariants fired anyway. Starvation tolerated *and*
  the invariant criterion red, in the same run, is the direct demonstration that the two criteria
  were separated rather than that the rule was switched off. It was not staged — the starvation
  arose from this machine's own contention.

### 6.4 Revert and restoring GREEN

```
$ git diff --stat -- internal/kanban/backlog_store.go internal/kanban/board_lock_unix.go internal/kanban/board_lock_windows.go internal/kanban/board_lock.go
(no output — all four files identical to base)

$ go test -race -count=1 -v -run TestConcurrencyStress ./internal/kanban/
=== RUN   TestConcurrencyStress
=== PAUSE TestConcurrencyStress
=== CONT  TestConcurrencyStress
    backlog_concurrency_test.go:235: SPEC-STRESS-INVARIANT-VERDICT-001: 8 writers x 6 adds = 48 attempts; 48 succeeded, 0 starved (tolerated), 0 hard failures; 48 distinct ids, 48 stored items, last_seq 48; back-derived per-mutation cost 14.81139ms (elapsed 710.94675ms / 48 successful mutations)
--- PASS: TestConcurrencyStress (0.72s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	2.750s
```

## 7. AC-SIV-012 — scope discipline (diff taken AFTER both mutants were reverted)

```
$ git diff --stat HEAD
 internal/kanban/backlog_concurrency_test.go | 329 +++++++++++++++++++++++++---
 internal/kanban/board_lock_wait_test.go     |  58 +++++
 internal/kanban/board_store.go              |   9 +
 3 files changed, 362 insertions(+), 34 deletions(-)
```

Changed paths are exactly the three the SPEC permits.

`board_store.go` is comment-only:

```
$ git diff HEAD -- internal/kanban/board_store.go | grep -E '^[+-]' | grep -v '^[+-][+-]' | grep -vE '^\+\s*//'
(no output — every added line is a comment, no line removed)
```

No **surviving** change to `board_lock_unix.go`, `board_lock_windows.go`, `board_lock.go`, or
`backlog_store.go` (§6.4 diff, empty). `boardLockCIMutationCost`, `boardLockHeadroom`,
`boardLockSupportedWriters`, `boardLockWaitMin/Max/Step`, and `boardLockRetryWait` are unchanged in
value and behaviour.

## 8. Non-claims (REQ-SIV-014, all four stated)

1. **No before/after comparison exists, in any quantity.** t370 never measured the pre-repair firing
   rate, so no improvement-versus-prior-rate claim is available at any magnitude. The strongest
   sentence this card supports is: *the verdict criterion moved to the invariants, and under that
   criterion it is green.*
2. **A single green run can never close this card.** Post-repair there exists one green `Race Test`
   job (`51daada00`) and one run in which `TestConcurrencyStress` was green inside a job reddened by
   a **different** test (`c6aa61346`) — calling the second a green run overstates the source.
   Judgment requires a firing rate across multiple runs (AC-SIV-013, ≥5 post-landing develop heads),
   which remains **OPEN at merge**. This is exactly where card t354 stopped.
3. **A green observation window does not evidence that the invariants still fire.** The invariant
   criterion was red in **0 of the 14** observed runs *before* this change, so a post-landing window
   of greens under the new criterion is fully consistent with the invariants having been switched
   off. It evidences only that no *new* failure mode was introduced. The burden of showing the
   invariants still fire belongs solely to the §6 mutant evidence.
4. **The tolerated error class is wider than "contention" on the CI platform.** The tolerance admits
   whatever `board_lock_unix.go` maps to `ErrBoardLockHeld`, which is **every** `unix.Flock`
   failure — `ENOLCK`, `EINTR`, `EBADF` included — not only `EWOULDBLOCK`; the Windows substrate is
   narrower by contrast, so the tolerance is widest on exactly the platform the CI `Race Test` job
   runs. Separately, `IsBoardLockHeld` is `errors.Is`, which traverses the `errors.Join(mutErr,
   relErr)` that `Mutate` returns, so a future joined error carrying a lock-held branch alongside a
   real defect would be tolerated wholesale. Neither is repaired here; narrowing the sentinel is
   production behaviour and belongs to a follow-up card.

No sentence in this report asserts an improvement versus the pre-repair firing rate, and none claims
the budget is sufficient on any machine.

## 9. Gaps — what was explicitly NOT observed

- **No CI run was read or triggered.** Every figure here is local darwin. Cross-platform behaviour
  (linux, windows) is unobserved in this run.
- **No full-repository suite.** Verification was scoped to `internal/kanban` per the lane-local
  discipline; `go test ./...` was deliberately not run.
- **No `golangci-lint` run.** Only `go vet ./internal/kanban/...` (clean) and `gofmt -l`
  (clean) were executed.
- **Cross-platform build not exercised.** `GOOS=windows go build` / `go vet` were not run; the
  change is test-only plus a comment, but that is an argument, not an observation.
- **The pre-repair firing rate was not measured** and is unrecoverable (§8 clause 1).
- **AC-SIV-013 is unmeasured by construction** — it requires post-landing develop heads that do not
  yet exist.
- **The 48-add fan-out under real starvation was observed once, incidentally** (§6.3.2, 1 starved of
  48). No run was staged to produce starvation at the parent test's scale; the AC-SIV-001/005
  sub-tests produce it deterministically at 1-2 adds instead.
- **The mutants were scale-gated.** Ungated versions of the same shapes were not run, so the
  behaviour of the other eleven census files under an ungated mutation is unobserved.

## 10. Residual risk

- **The tolerated class is wider than contention** (§8 clause 4). A future `ENOLCK`/`EINTR`
  regression, or a joined error carrying a lock-held branch alongside a genuine defect, would be
  absorbed as starvation. This is the single largest weakening the change introduces, and it is
  accepted rather than repaired.
- **The zero-progress floor admits 1 success in 48**, where t370 measured real starvation at 3-7 of
  48 — roughly 40× weaker than observed behaviour. Conservation is the compensating control; the
  floor was deliberately not tightened into a fraction, because a fractional floor is a load sensor
  wearing an accounting label and would recreate the flake on the next slower runner.
- **The new guard's name contains "Covers".** The name is the one prescribed verbatim by `plan.md`
  §F M1 step 2, and the SPEC's no-overclaim constraint (REQ-SIV-009) binds the guard's *messages*,
  which comply. A CI reader nonetheless sees `TestBoardLockWaitBudgetCoversSerializedMutations` on a
  failure line, which reads as a coverage claim the guard does not make. Renaming was not done
  because the name is contract text; flagging it here rather than deviating unilaterally.
- **The constant-coherence margin is 4.2%** (66ms on 1.584s). The guard is a narrow tripwire, not a
  broad guarantee, and it is cost-independent by construction — no per-mutation cost regression of
  any size can make it fire.
- **`integration_lock_cross_test.go` sizes its 500ms release timeout against `boardLockWaitBudget`
  in a comment.** It did not fire under the §5 mutant (budget 1.32s, 500ms = 37.9%), but a deeper
  reduction of the budget could make that timeout marginal without any guard saying so.
- **Local green is not the verdict.** The CI `Race Test` job on `origin/develop` is, and it has not
  run against this tree.
