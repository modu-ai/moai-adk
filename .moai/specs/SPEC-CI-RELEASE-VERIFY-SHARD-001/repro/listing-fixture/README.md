# listing-fixture — what `go test -list` enumerates (research.md L-E20) and the paired race-listing evidence

Files carry a `.txt` suffix so no Go tool treats this directory as code. NOT executed at plan time (leader quiet window: no heavy Go execution). Copy each sub-directory to a scratch directory OUTSIDE the repository, dropping the `.txt` suffix, and run one Go call at a time under the resource-entry rules.

## behaviour/ — what the tool lists (L-E20)

`go -C <scratch> test -vet=off -list '.*' ./...` — expected (measured once, plain listing): `TestReal Test BenchmarkB Fuzz FuzzReal Example ExampleLower ExampleBlock ExampleUnordered`; not listed: `Fuzzhelper`, `ExampleAlpha`; `-vet=off` is REQUIRED for this package only because `ExampleAlpha`, `ExampleLower`, `ExampleBlock` and `ExampleUnordered` name identifiers that do not exist in it, which go vet's `tests` analyzer reports ("... refers to unknown identifier ...", observed once: the first scratch run without `-vet=off` ended `FAIL gl [build failed]`). It deliberately contains bare `Test`/`Fuzz`/`Example`, so it is NOT the base for the race-pair evidence below.

## race-pair/ — the paired evidence for mutant (x) (AC-RVS-002, plan M3)

One clean-base package: `base_test.go` holds `TestReal`, `BenchmarkB` and `Fuzzhelper` only (no bare `Test`/`Fuzz`/`Example`, no `Example...` function); `race_only_test.go` (`//go:build race`) adds `FuzzRaceOnly`. The name collector under test takes a package directory (or pattern) as input, so it is pointed at this scratch package. **No `-vet=off`**: the commands are exactly the collector's own, and the package must pass go vet's `tests` analyzer (read, not executed: the analyzer, `cmd/vendor/golang.org/x/tools/go/analysis/passes/tests/tests.go`, checks `Test`/`Benchmark`/`Fuzz`-prefixed functions only with one parameter and no result — `Fuzzhelper` has a result so it is skipped; `FuzzRaceOnly(f *testing.F)` passes `checkTest` and its `f.Fuzz(func(t *testing.T, n int) {})` / `f.Add(1)` calls are well-formed; `Example...` names are checked against package identifiers, which is why none is present).

| Run | Listing command | Exit | Names listed | Expected guard outcome |
|-----|-----------------|------|--------------|------------------------|
| 1 | `go -C <scratch> test -list '.*' ./...` (plain) | 0 | `TestReal`, `BenchmarkB` | cover check PASSES (`TestReal` matches selector 3 once); discovery guard PASSES (no `Fuzz...`/`Example...` name listed) |
| 2 | `go -C <scratch> test -race -list '.*' ./...` | 0 | adds `FuzzRaceOnly` | discovery guard FAILS naming `FuzzRaceOnly` |

Mutant (x) is satisfied only by BOTH runs on this same package, with run 1 exiting 0 and passing and run 2 exiting 0 and failing the guard. A run that ends `FAIL ... [build failed]` or exits non-zero is not evidence for either row. An implementation that drops `-race` yields run 1's result for both and is rejected. The expected outcomes are written down before any run; they have not been observed.
