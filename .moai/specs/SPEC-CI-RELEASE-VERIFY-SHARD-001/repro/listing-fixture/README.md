# listing-fixture — what `go test -list` enumerates (research.md L-E20) and the paired race-listing evidence

Files carry a `.txt` suffix so no Go tool treats this directory as code. NOT executed at plan time (leader quiet window: no heavy Go execution). Copy each sub-directory to a scratch directory OUTSIDE the repository, dropping the `.txt` suffix, and run one Go call at a time under the resource-entry rules.

## behaviour/ — what the tool lists (L-E20)

`go -C <scratch> test -vet=off -list '.*' ./...` — expected (measured once, plain listing): `TestReal Test BenchmarkB Fuzz FuzzReal Example ExampleLower ExampleBlock ExampleUnordered`; not listed: `Fuzzhelper`, `ExampleAlpha`. This package deliberately contains bare `Test`/`Fuzz`/`Example`; it is NOT the base for the race-pair evidence below (a plain listing of it already trips the guards).

## race-pair/ — the paired evidence for mutant (x) (AC-RVS-002, plan M3)

One clean-base package: `base_test.go` holds `TestReal`, `BenchmarkB`, `Fuzzhelper`, `ExampleAlpha` (no bare `Test`/`Fuzz`/`Example`, no output-bearing example); `race_only_test.go` (`//go:build race`) adds `FuzzRaceOnly`. The name collector under test takes a package directory (or pattern) as input, so it is pointed at this scratch package.

| Run | Listing command | Names listed | Expected guard outcome |
|-----|-----------------|--------------|------------------------|
| 1 | `go -C <scratch> test -vet=off -list '.*' ./...` (plain) | `TestReal`, `BenchmarkB` | cover check PASSES (`TestReal` matches selector 3 once); discovery guard PASSES (no `Fuzz...`/`Example...` name listed) |
| 2 | `go -C <scratch> test -race -vet=off -list '.*' ./...` | adds `FuzzRaceOnly` | discovery guard FAILS naming `FuzzRaceOnly` |

Mutant (x) is satisfied only by BOTH runs on this same package: run 1 passing and run 2 failing. An implementation that drops `-race` yields run 1's result for both and is rejected (it never fails). The two expected outcomes are written down here before any run; they have not been observed.
