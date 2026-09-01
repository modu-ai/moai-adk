# Run-phase evidence — SPEC-VACUOUS-FLOOR-GUARD-001 (card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, base `3f03d9c36`,
pre-commit HEAD `226bdd0dc`. Companion records: `repair-direction.md` (M1), `census.md` (M2),
`mutants.md` (M3), `negative-evidence.md` (M4).

## Per-AC PASS/FAIL matrix

| AC | Verdict | Verification command | Observed |
|---|---|---|---|
| AC-VFG-001 | **PASS** | `grep -n 'boardLockWaitBudget <' …`; `grep -c 'floor :=' …`; `go vet ./internal/kanban/...` | 1 match at line 122 (t372's guard), baseline 2; `floor :=` count 1 at line 120, baseline 2; vet exit 0 |
| AC-VFG-002 | **PASS** | `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/` | `--- PASS`, 1 `=== RUN` (non-zero selector); four assertions at lines 28 / 65 / 74 / 79 |
| AC-VFG-003 | **PASS** | M1 plant → whole-package run → revert → re-run | RED with `per-mutation cost 20ms is below the CI-class observation of 33ms`; post-revert `ok` |
| AC-VFG-004 | **PASS** | M2 plant → whole-package run → revert → re-run | RED with `headroom factor 1 states no headroom`; post-revert `ok` |
| AC-VFG-005 | **PASS** | M3 plant → whole-package run → revert → re-run | RED with `supported writers = 8, want 10 (…)`; post-revert `ok` |
| AC-VFG-006 | **PASS** | M4 form A (`1650ms`) then form B (`1400ms`) → revert | Form A GREEN (the stated gap, observed); form B RED with `is not the product of its named inputs`; post-revert `ok` |
| AC-VFG-007 | **PASS** | Branch present + M2 planted, scoped `-v` run | `headroom factor 1 states no headroom` PRESENT; branch message `< headroom floor` ABSENT at a 330ms budget |
| AC-VFG-008 | **PASS (with a stated criterion-wording qualification)** | `git diff --stat`; `git diff` per file; `grep -rn 'go test' .moai/reports/t378/`; `./bin/moai spec lint --strict` | 1 file / 27+ 7-; `board_store.go` diff empty; every recorded invocation compliant; `0 error(s), 1096 warning(s)` |

8 of 8 PASS. Detail for the AC-VFG-008 qualification is below.

## AC-VFG-001 — the dead branch is gone, nothing of its shape replaced it

Command: `grep -n 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go`
Observed (exactly one match; baseline before the edit was **2**):

```
122:	if boardLockWaitBudget < floor {
```

Command: `grep -n 'floor :=' internal/kanban/board_lock_wait_test.go`
Observed (exactly one match; baseline before the edit was **2**):

```
120:	floor := time.Duration(serialized) * boardLockCIMutationCost
```

Both survivors are lines 120 and 122 — t372's `TestBoardLockWaitBudgetCoversSerializedMutations`,
whose floor is `serialized * cost` (the stress constants), a term the budget expression does not
supply. That is the file's one legitimate floor comparison and it is untouched.

The pre-repair baselines were recorded BEFORE the edit, so "1 remaining" is a measured delta
rather than an absolute:

```
boardLockWaitBudget < : 2
floor := : 2
```

Command: `go vet ./internal/kanban/...` — exit 0, no output.
Command: `gofmt -l internal/kanban/` — no output.

No new inequality was introduced (REQ-VFG-003): the diff below shows the removed block replaced by
comment lines only.

## AC-VFG-002 — the four load-bearing assertions survive, GREEN

Command: `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`
Observed (exit 0):

```
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.452s
```

**Selector match count: 1** (`=== RUN` present). This check is not decoration: a zero-match
selector also prints `ok`, so the `=== RUN` line is what separates a real pass from a vacuous one.

Command: `grep -n 'boardLockWaitBudget != recomputed\|boardLockSupportedWriters != 10\|boardLockCIMutationCost < 33\|boardLockHeadroom < 2' internal/kanban/board_lock_wait_test.go`
Observed, at post-repair line numbers:

```
28:	if boardLockWaitBudget != recomputed {
65:	if boardLockSupportedWriters != 10 {
74:	if boardLockCIMutationCost < 33*time.Millisecond {
79:	if boardLockHeadroom < 2 {
```

All four retained, unweakened — same operators, same thresholds as at base.

Whole-package, post-repair:
Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	15.830s` (exit 0).

## AC-VFG-003 .. AC-VFG-006 — mutant evidence

Full censuses, pre-plant predictions, mutant diffs, verbatim RED outputs, reverts, and post-revert
GREENs are in `mutants.md`. Summary of the named-assertion attribution:

| AC | Mutant | Named assertion message (verbatim) | Line | Other REDs in the same run |
|---|---|---|---|---|
| 003 | cost 33ms→20ms | `per-mutation cost 20ms is below the CI-class observation of 33ms` | 55 | none |
| 004 | headroom 5→1 | `headroom factor 1 states no headroom` | 60 | 3, each named and attributed |
| 005 | writers 10→8 | `supported writers = 8, want 10 (Factory mode's ten lanes against one queue)` | 46 | 1 (t372's guard), predicted |
| 006 | budget→1400ms | `budget 1.4s is not the product of its named inputs (…)` | 29 | 1 (t372's guard), predicted |

Attribution rests on the verbatim message in every row, never on the failure count (REQ-VFG-005).

Two things `mutants.md` records that are worth surfacing here:

- **A prediction that missed.** The census predicted `TestConcurrencyStress` might starve under
  M2's 330ms budget. It did not redden; two other budget-consuming tests did
  (`TestIntegrationLockAcquire_SerializedAcrossProcesses`, `TestBacklogLock_TimeoutNamesLockPath`),
  both inside the predicted category. The prediction was hedged ("MAY"), so this is not a
  contradiction — it is recorded because writing predictions first is only useful if a miss stays
  visible.
- **AC-VFG-006's stated gap is observed, not assumed.** The `1650 * time.Millisecond` form
  genuinely replaced the derivation with a bare literal — the regression REQ-BLB-001 exists to
  catch — and BOTH guards passed (2 `=== RUN`, both `--- PASS`). The equality compares values, not
  syntax. That is a real limit of the retained assertion, recorded rather than hidden.

## AC-VFG-007 — the deletion's negative evidence

Full record in `negative-evidence.md`. The observation was taken with the branch STILL PRESENT in
its landed form (pre-edit), so no reinstated copy could differ from the original.

Command: `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`
with M2 (`boardLockHeadroom` 5→1) planted. Complete output (exit 1):

```
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
    board_lock_wait_test.go:60: headroom factor 1 states no headroom
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.234s
```

PRESENT: `headroom factor 1 states no headroom`. ABSENT: any line from `board_lock_wait_test.go:40`
and any occurrence of `< headroom floor`. The output above is complete, not filtered — an absence
claim read off a truncated capture would establish nothing.

Budget at that moment: `10 x 33ms x 1 = 330ms`, half the 660ms floor the four retained assertions
compose. A floor that meant anything would have fired.

**M2 was required and is not interchangeable.** M1 leaves the budget at 1000ms and M3 at 1320ms,
both above 660ms, so a genuine floor would have been silent under either and the observation would
have established nothing. This is recorded in `census.md` as arithmetic written before the runs.

## AC-VFG-008 — the record is written and the scope held

Command: `git diff --stat`
Observed:

```
 internal/kanban/board_lock_wait_test.go | 34 ++++++++++++++++++++++++++-------
 1 file changed, 27 insertions(+), 7 deletions(-)
```

Exactly one source file changed.

Command: `git diff -- internal/kanban/board_store.go`
Observed: empty (no output). No constant was changed; every mutant was reverted.

Command: `git diff -- internal/kanban/board_lock_wait_test.go`
Observed — a single hunk, whose header names the enclosing function
`TestBoardLockWaitBudgetDerivedFromNamedInputs`:

```
@@ -32,13 +32,33 @@ func TestBoardLockWaitBudgetDerivedFromNamedInputs(t *testing.T) {
-	// The inequality REQ-BLB-002 states: at least headroom x per-mutation
-	// cost x supported contender count.
-	floor := time.Duration(boardLockSupportedWriters) *
-		boardLockCIMutationCost * boardLockHeadroom
-	if boardLockWaitBudget < floor {
-		t.Errorf("budget %v < headroom floor %v", boardLockWaitBudget, floor)
-	}
+	// Where REQ-BLB-002's floor is enforced: INPUT-WISE, by the equality
+	// above plus the three assertions below — never by a comparison against
+	// boardLockWaitBudget.
+	// … (the composed-floor arithmetic, the reinstatement warning, and the
+	//    pointer to t372's guard as the file's one legitimate floor)
```

The hunk spans original lines 32-44. t372's `TestBoardLockWaitBudgetCoversSerializedMutations`
begins at original line 95 and appears in no hunk — it is byte-identical, as AC-SIV-013's open
observation window requires.

**The rewritten comment** discharges REQ-VFG-004 on both counts. It names the enforcement site —
the equality plus the three input assertions, composing `budget >= 10 * 33ms * 2 = 660ms`, each
conjunct falsifiable by one constant — and it states why no floor-versus-budget comparison appears
in the function: the budget IS that three-constant product, the equality above is a `t.Fatalf` hard
stop, so any floor built from those terms evaluates false on every assignment. It names the removal
and its SPEC, so reinstating one reads as a regression rather than an improvement, and it points at
t372's guard as the file's one legitimate floor comparison.

Command: `./bin/moai spec lint --strict` (the TREE binary, built this run — not the PATH binary)
Observed, final line: `0 error(s), 1096 warning(s)`
Findings naming this SPEC: `grep -c 'SPEC-VACUOUS-FLOOR-GUARD-001' <lint output>` → `0`.

Binary identity, because the PATH binary is stale (v3.1.2, predating t299's `SyncSHASlotFormat`
rule) and reports a different rule set:
Command: `strings bin/moai | grep -c 226bdd0dc` → `4` (non-zero: the binary carries this HEAD).
`make build` modified no tracked file — `git status --short` after the build showed only this
card's own test-file edit and its untracked evidence files.

### The verification-load check, and its one honest qualification

Command: `grep -rn 'go test' .moai/reports/t378/`
Observed — full output, 15 matching lines:

```
.moai/reports/t378/negative-evidence.md:47:`go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`
.moai/reports/t378/negative-evidence.md:95:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:25:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:45:Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:62:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:101:Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:118:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:139:Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:159:`go test -timeout 600s -count=1 -v -run 'TestBoardLockWaitBudgetDerivedFromNamedInputs|TestBoardLockWaitBudgetCoversSerializedMutations' ./internal/kanban/`
.moai/reports/t378/mutants.md:194:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/mutants.md:223:Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
.moai/reports/t378/plan-audit.md:18:  … REQ-007 (L165) Event-driven "**When** any verification for this SPEC is run, it shall be a single serial `go test` invocation". …
.moai/reports/t378/plan-audit.md:71:  … (the plan-auditor's prose stating the prohibition) …
.moai/reports/t378/plan-audit.md:72:  … (the plan-auditor's prose proposing the AC wording) …
.moai/reports/t378/census.md:9:Command: `go test -timeout 600s -count=1 ./internal/kanban/`
```

**12 of the 15 lines are recorded invocations. Every one is
`go test -timeout 600s -count=1 [-v -run …] ./internal/kanban/`** — scoped to the single package,
serial, no race detector, no full-suite target, no backgrounding.

**The qualification, stated rather than glossed:** AC-VFG-008 asks for "zero occurrences of
`./...`, zero of `-race`". Taken as a literal token count over the grep output, that is NOT
satisfied — both tokens appear, on `plan-audit.md` lines 71-72, inside the plan-auditor's own prose
*stating the prohibition* and proposing this very criterion's wording. Verified precisely:

Command: `grep -rn -- '-race' .moai/reports/t378/` → 2 matches, both `plan-audit.md:71,72`.
Command: `grep -rn 'go test \./\.\.\.' .moai/reports/t378/` → 2 matches, both `plan-audit.md:71,72`.
Command: `grep -rn 'go test.*&\s*$' .moai/reports/t378/` → no output (zero backgrounded invocations).

So the substantive requirement REQ-VFG-007 states — every verification is a single serial
package-scoped invocation without the race detector, without the full suite, without a background
process — holds without exception, and that is the claim recorded as PASS. The literal
zero-occurrence phrasing is defeated only by a document that describes the prohibition, which is a
criterion-wording artifact and not a load violation. Recording it as a clean zero would have been
the more comfortable report and the false one.

This file inherits the same property: the paragraph above necessarily contains both tokens.

## Load discipline actually practised

Every observation was one serial invocation, run to completion before the next was issued. No
`&`, no `nohup`, no background execution, no generated load, no full-suite run. Mutants were
planted one at a time and reverted before the next, with `git diff --stat -- internal/kanban/board_store.go`
confirming clean after each. Total: 12 test invocations across the whole card.

## Gaps — what was explicitly NOT observed

1. **Cross-platform build.** Not run. This card changes one `_test.go` file with no build tags and
   no platform-conditional code, so a Windows/Linux build result would add nothing; it was not run,
   and is therefore recorded as a gap rather than claimed.
2. **Full-suite verdict.** Deliberately not run locally (REQ-VFG-007). CI on `origin/develop` owns
   it. Nothing here establishes that packages outside `internal/kanban` are unaffected — though the
   changed file is a test file in one package, so the blast radius is bounded by construction.
3. **Coverage percentage.** Not measured. The change removes an unreachable branch and adds
   comments; no new executable line was introduced.
4. **`golangci-lint`.** Not run. `gofmt -l` and `go vet` were.
5. **Unreachability on every assignment.** Established by the static argument only (spec.md §A.1).
   AC-VFG-007 corroborates it on ONE assignment of the three constants.
6. **Sibling vacuous guards elsewhere in the repository.** Not swept — excluded by the SPEC as an
   unverified defect claim. Nothing here says whether others exist.
7. **`integration_lock_cross_test.go:54-55`** carries a comment restating the budget arithmetic.
   Observed during the census, not touched, and not verified for accuracy.

## Residual risk

- **The comment can rot.** It states `660ms` and `10 * 33ms * 2` inline. If a future card retunes a
  constant, the comment's arithmetic goes stale while the four assertions keep working. The
  assertions are the enforcement; the comment is the record. Mitigated only by the assertions
  themselves being mutation-proven, not by anything in the comment.
- **AC-VFG-006's form-A gap is now a documented, unclosed hole.** A bare literal numerically equal
  to the product still passes the derivation equality. This card observed and recorded that; it
  did not fix it, and fixing it would need a different kind of assertion (syntactic or
  build-time), which is out of scope.
- **The composed 660ms floor is weaker than the landed 1650ms budget by design.** That is not a
  loosening — the 1650 was never enforced by the deleted branch — but a reader comparing the two
  numbers without reading `negative-evidence.md` may mistake it for one. The comment and that
  record both address it; a reviewer who reads neither could still raise it.
- **t372's guard was not exercised for its own sake.** It reddened under M2, M3, and M4-B, which
  is incidental evidence that it works, but this card ran no observation designed to test it and
  claims nothing about its health.
