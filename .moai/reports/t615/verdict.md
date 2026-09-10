# t615 — DiffLines allocates a full DP table for identical inputs

Card t615 (Go full review F07, P2 performance, Tier S, Class B). Worktree
`.claude/worktrees/t615`, branch `WT-diff-equal-fastpath`, base `develop`
`d3b7d438d`.

Source report: `go-full-review-20260910-01a089e2` §F07. Its evidence file
numbers findings differently from the report body (this one is item 2 of
`performance-findings.md`), so it was located by its target file,
`internal/merge/differ.go`. The report is an untracked file in the primary
checkout; its benchmark was copied into this worktree byte-exact (`cmp` exit 0)
and is quoted below so the measurement stays reproducible after the copy was
removed.

## Claim

1. **The defect reproduced on `develop`** with the report's own instrument and
   flags: identical 2,000-line inputs allocated 32,833,540 B/op in 2002
   allocations; 2,001 lines allocated nothing.
2. **Fixed** by returning `nil` before the LCS table is built when the inputs
   are equal (`slices.Equal`). After the fix, the same instrument measured
   **0 B/op, 0 allocs/op at every size**: 100, 1000, 2000, and 2001.
3. **The return contract is unchanged.** On the unmodified code, both existing
   paths already returned `nil` for identical inputs. The new contract test
   passed on the old code and passes on the fixed code.
4. **Non-identical inputs behave as before.** Six boundary cells (one-line
   change, insertion, deletion at 2000 and 2001 lines) passed on the old code
   and pass on the fixed code.
5. **The tests catch the two relevant regressions.** A mutant returning
   `[]Edit{}` and a mutant widening the fast path to a length comparison each
   turned the targeted tests red, in the pattern predicted before they ran.
6. **A latent defect in an existing test helper was found and not fixed.**
   `reconstructFromGreedyEdits` (`differ_perf_test.go`) inserts at index 0
   whenever a script has no deletes, whatever the `NewLine` value.
7. The threshold (`diffLinesThreshold = 2000`) is unchanged, as the card
   requires.

## Evidence

### The aborted delegation

The implementation was first delegated to manager-develop. That agent **did
not return**: it stopped on an API session limit (HTTP 429). No completion
report exists, so nothing it did is counted as done. Before continuing, the
lane checked what the agent had left. `differ.go` was unmodified (`git diff`
produced 0 lines). The test file existed. Only the agent's own old-code runs
had been recorded, and no mutant or after-measurement had been run. The lane
finished the work directly, because re-delegating on the same model was likely
to hit the same limit. Every result below was re-run by the lane rather than
taken from the agent's files.

### The instrument

The report's benchmark, byte-exact:

```go
func BenchmarkAuditDiffLinesIdentical(b *testing.B) {
 for _, n := range []int{100,1000,2000,2001} {
  b.Run(fmt.Sprint(n),func(b *testing.B) {
   lines := make([]string,n)
   for i := range lines {lines[i]=fmt.Sprintf("line %d",i)}
   b.ReportAllocs()
   b.ResetTimer()
   for i:=0;i<b.N;i++ {if edits:=DiffLines(lines,lines); len(edits)!=0 {b.Fatalf("unexpected edits: %d",len(edits))}}
  })
 }
}
```

The before and after runs used the report's exact invocation:

```
unset GOFLAGS MOAI_KANBAN … && GOMAXPROCS=2 timeout 120s go test -p=2 ./internal/merge \
  -run '^$' -bench '^BenchmarkAuditDiffLinesIdentical$' -benchmem -benchtime=200ms -count=2 \
  > <file> 2>&1; echo "EXIT=$?"
```

### Before and after, same instrument and flags

`bench-before-develop-reportflags.txt` (EXIT=0) and
`bench-after-develop-reportflags.txt` (BENCH-AFTER-EXIT=0), verbatim cells:

| Identical size | Before | After |
|---|---|---|
| 2000 | `32833540 B/op  2002 allocs/op` (×2) | `0 B/op  0 allocs/op` (×2) |
| 1000 | `8224836` / `8224770 B/op  1002 allocs/op` | `0 B/op  0 allocs/op` (×2) |
| 100 | `93184 B/op  102 allocs/op` (×2) | `0 B/op  0 allocs/op` (×2) |
| 2001 (control) | `0 B/op  0 allocs/op` (×2) | `0 B/op  0 allocs/op` (×2) |

The report's run on `main` under the same flags measured
`32833536 B/op  2002 allocs/op` for 2000. The allocation count matches exactly.
The 2002 allocations are the outer `dp` slice plus 2001 rows. The 4-byte
difference is runtime noise.

**B/op and allocs/op are the primary evidence.** ns/op is secondary, and its
conditions differ between the two runs, so no speed-up ratio is claimed. For
the 2000 cell it moved from 8,384,301–8,755,363 ns/op at load average 21.61 to
2,813–3,105 ns/op at load average 9.05. Both runs used
`go version go1.26.4 darwin/arm64` on an Apple M4 Max with 16 CPUs.

### The change

`differ-fix.diff`, 36 lines, three hunks: `"slices"` added to the imports, one
godoc sentence, and the fast path:

```go
func DiffLines(a, b []string) []Edit {
	// Identical inputs have no edits. Both paths below also return nil for
	// equal inputs, so returning here keeps the result unchanged while skipping
	// the O(m*n) table.
	if slices.Equal(a, b) {
		return nil
	}
```

`gofmt -l` on both changed Go files printed nothing (`gofmt-final.txt`,
0 bytes).

### The return contract

On the unmodified code, the DP path starts `var edits []Edit`, appends nothing
for equal lines, reverses zero elements, and returns `nil`. `diffLinesGreedy`
trims the whole input as a common prefix and also returns its unappended
`nil`. The three production callers (`UnifiedDiff`, `computeLineChanges`, and
`mergeLineBased` through the latter) only check `len` or `range`, but the
contract must stay byte-identical so that a future `edits == nil` check does
not silently diverge.

The pre-existing `TestDiffLines_IdenticalInput` asserts only `len(edits) != 0`,
which passes for both `nil` and `[]Edit{}`. Mutant A below demonstrates the gap.

### The tests

The new file is `internal/merge/differ_equal_fastpath_test.go`:

- `TestDiffLines_IdenticalReturnsNil` checks identical inputs at `nil,nil`,
  `[]string{},[]string{}`, 3, 2000, and 2001 lines and asserts `edits == nil`.
- `TestDiffLines_IdenticalAllocatesNothing` uses `testing.AllocsPerRun` on
  identical 2000-line input and asserts 0. Its control is the same size with
  one line different, asserting `> 0`. The test deliberately omits
  `t.Parallel()`, because `AllocsPerRun` reads process-wide counters.
- `TestDiffLines_SingleEditAtThresholdBoundary` covers a one-line change, an
  insertion, and a deletion at 2000 (DP) and 2001 (greedy) lines, and asserts
  the exact op list, indices, and text.

### A latent helper defect, found on the old code

The first version of the boundary test also called
`reconstructFromGreedyEdits`. On the **unmodified** code
(`lane-rerun-old-code.txt`), one cell failed:

```
differ_equal_fastpath_test.go:119: greedy edits do not reconstruct b (result len=2001, b len=2001)
--- FAIL: TestDiffLines_SingleEditAtThresholdBoundary/one_insertion_greedy_2001
```

In the same cell, the op-list, `NewLine`, and `NewText` assertions reported no
error, so `DiffLines` had produced the correct script: one insert at line
1000. The failure came from the helper. When a script has no deletes, it sets
`delStart = 0` (`differ_perf_test.go:28-31`) and splices the insertions in at
index 0, whatever their `NewLine`. The length still matches, but the order is
wrong. The only existing caller, `empty_a`, inserts at 0 anyway, which is why
the defect never showed.

The new test no longer uses the helper. For single-edit scripts the op-list,
index, and text assertions determine the script exactly, so nothing is lost.
The helper itself is **not modified**: it belongs to a pre-existing file
outside this card's scope. It is reported here as a latent trap for any
future mid-file pure-insertion case.

### Before the fix: the old contract established and RED confirmed

`lane-before-fix.txt` (BEFORE-FIX-EXIT=1). `git diff --quiet -- internal/merge/differ.go`
exited 0 (unmodified) for this run.

```
--- FAIL: TestDiffLines_IdenticalAllocatesNothing
    differ_equal_fastpath_test.go:50: identical 2000-line inputs: expected 0 allocs per run, got 2002
--- PASS: TestDiffLines_IdenticalReturnsNil          (5/5 subtests PASS)
--- PASS: TestDiffLines_SingleEditAtThresholdBoundary  (6/6 subtests PASS)
```

The allocation guard is the only failure. The contract and boundary tests
pass on the old code, so they can serve as before/after guards.

### After the fix: all green, and the tests genuinely ran

`lane-after-fix.txt` (AFTER-FIX-EXIT=0):
`--- PASS: TestDiffLines_IdenticalAllocatesNothing`,
`TestDiffLines_IdenticalReturnsNil` 5/5, and
`TestDiffLines_SingleEditAtThresholdBoundary` 6/6. The file was read rather
than relying on the exit code, because a `-run` selector that matches nothing
also exits 0.

### Mutants: the tests can fail

Each mutant was applied to `differ.go`, the targeted tests were run, and the
file was restored. The expected pattern was fixed **before** each run.

**Mutant A**: the fast path returns `[]Edit{}`. The prediction was that all 5
contract subtests fail on the assertion and none fail to build.
`mutantA.txt` (MUTANT-A-EXIT=1):

```
differ_equal_fastpath_test.go:28: expected nil edits for identical input, got non-nil slice (len=0)
--- FAIL: TestDiffLines_IdenticalReturnsNil/{empty_empty,three_lines,over_threshold_2001,nil_nil,threshold_2000}
```

All 5 failed on the assertion, with no build failure. `len=0` is the point:
the pre-existing length-only test would have passed this mutant.

**Mutant B**: the fast path is widened to a length comparison. It was written
as `slices.Equal(a[:0], b[:0]) && len(a) == len(b)` so that `slices` stays
imported. A bare `len(a) == len(b)` would have left the import unused, and the
test would have failed at compile time, which says nothing about the
assertions. Predicted: the control fails, both `one_line_change` cells fail
(equal lengths), and the four insertion and deletion cells pass (different
lengths). `mutantB.txt` (MUTANT-B-EXIT=1):

```
differ_equal_fastpath_test.go:57: control (2000 lines, one line differs): expected allocs > 0, got 0
--- FAIL: TestDiffLines_IdenticalAllocatesNothing
differ_equal_fastpath_test.go:95: expected non-nil edits for non-identical input
--- FAIL: .../one_line_change_dp_2000
--- FAIL: .../one_line_change_greedy_2001
--- PASS: .../one_deletion_greedy_2001  one_insertion_greedy_2001  one_deletion_dp_2000  one_insertion_dp_2000
```

This matches the prediction cell for cell, with no build failure.

**Restore**: sha256 of the fixed `differ.go` recorded before mutating
(`differ-fixed.sha256`) and after restoring:

```
3f0407512a5239780191477d8828f8ae8b0006c7075a279e1ec0f91d72b1eef2  (baseline)
3f0407512a5239780191477d8828f8ae8b0006c7075a279e1ec0f91d72b1eef2  (restored)
```

### Final tree verification: touched package only

The benchmark copy was deleted first; `test -e` then exited 1.

```
go vet ./internal/merge/                    VET-EXIT=0    vet-final.txt 0 bytes
golangci-lint run ./internal/merge/...      LINT-EXIT=0   lint-final.txt: "0 issues."
go test ./internal/merge/ -count=1 -v       TEST-EXIT=0   test-final.txt
```

`test-final.txt` was counted directly rather than through a pipe: `--- FAIL`
lines 0, `no test files` lines 0, `ok  	github.com/modu-ai/moai-adk/internal/merge	0.458s`
at line 548, 83 top-level `--- PASS` lines, and a PASS line for each of the
three new tests.

A Unicode format-character (`Cf`) count was run on every Go file written
through a tool payload, with a control string holding one such character
(counted 1): `differ.go` 0, `differ_equal_fastpath_test.go` 0.

## Integration window

The window was acquired as lane-3 (`moai integration acquire`, exit 0, no
settings-drift detection). Local `develop` was `c352330d3`, and the merge base
`d3b7d438d` was re-confirmed inside the window. Every measurement in this section
was taken at the absorb commit `4b87a7244`.

### Absorb judgment

**Dependency set of `internal/merge`** (`go list -deps -test`, non-standard
packages only; `deps-internal-merge.txt`): `gopkg.in/yaml.v3`, which is external
and pinned by `go.mod`/`go.sum`, and `internal/merge` itself. The package has no
module-internal dependency and no embed pattern.

That is a claim of absence, so the observer got a positive control first. The
same template run on `./internal/cli/update/merge` listed 118 non-standard
packages, one of which was `github.com/modu-ai/moai-adk/internal/merge`
(`deps-observer-control.txt`). The observer does report module-internal
dependencies, so the short list reflects a real absence rather than an observer
that stayed silent.

**Delta paths** (from diffing the merge base against `c352330d3`, name-only;
`absorb-delta-paths.txt`): 121 paths. Matches against the judgment set: 0 under
the `internal/merge/` prefix, 0 for `go.mod`, 0 for `go.sum`. Positive controls
on the same file: 24 under `internal/`, 95 under `.moai/`, 2 under `.claude/`.
Those add up to 121, so every line is accounted for.

**Second observer: content hashes**, taken at the merge base and at `develop`.
All three are the same on both sides: `internal/merge` `80b758a1`, `go.mod`
`18a58bbe`, `go.sum` `a59f8dab`.

The two observers agree that the absorbed delta touches nothing `internal/merge`
depends on, which would have permitted carrying the earlier results forward.
The package was re-measured directly anyway. It is small and needs no test
slot, and a direct run also covers the build of the test binary, which a hash
comparison cannot.

### Absorb and merged-tree re-measurement

Absorb commit `4b87a7244` (parents `05946ea8c` and `c352330d3`): an `ort` merge,
no conflict, 121 files changed, matching the delta. The `internal/merge` tree is
`17a1a7e6` both at `05946ea8c` and at the absorb commit.

| Check | Result |
|---|---|
| `go test ./internal/merge/ -count=1 -v` | exit 0; 0 `--- FAIL` lines, 0 `no test files` lines, `ok` line present, 83 top-level `--- PASS`, all three new tests PASS (`merged-tree-test.txt`) |
| `go vet ./internal/merge/` | exit 0, 0 bytes of output |
| `golangci-lint run ./internal/merge/...` | exit 0, `0 issues.` |
| `gofmt -l` on both changed Go files | 0 bytes of output |

The benchmark was not re-run. `internal/merge` is byte-identical across the
absorb, so the allocation result cannot have changed.

## Baseline-attribution

Everything above was measured in this run, in this worktree, at
`HEAD = d3b7d438d` plus the working-tree changes listed. The before and after
benchmarks share one instrument, the report's file byte-exact, and one
invocation. No number is carried from the report as a measurement of this
tree; the report's figures appear only as a comparison. Results recorded by
the aborted agent were not used as evidence. Each one was re-run by the lane
first.

## Gaps

- This file does not record the tree identity of the final `develop` merge,
  because the file is committed before that merge happens. The lane's
  completion report carries it instead. The merged-tree re-measurement was run
  on the absorb commit; see Integration window.
- Real-world invocation frequency of `DiffLines` on identical inputs is not
  measured; the source report states the same gap.
- ns/op was measured under different load averages before and after (21.61 vs
  9.05), so no speed-up ratio is claimed.
- No end-to-end `moai update` run exercised the merge paths that call
  `DiffLines`; the evidence is package-level.
- Cross-platform builds were not run. CI owns that verdict.
- `reconstructFromGreedyEdits` is left unfixed (see above).

## Residual-risk

- The fast path adds a `slices.Equal` scan in front of every call. For
  non-identical inputs of equal length it runs until the first differing line,
  so inputs that differ only near the end pay a near-full linear scan before
  the DP path. That cost is O(n) against the O(m·n) table that follows, so it
  is negligible, but it was not measured separately.
- Identical inputs above 2000 lines no longer reach `diffLinesGreedy`. Both
  paths returned `nil` there, and the `over_threshold_2001` contract cell
  covers it, but the greedy path's own identical-input handling is now
  exercised only by `TestDiffLinesGreedyProperty/identical_large` and no
  longer through `DiffLines`.
