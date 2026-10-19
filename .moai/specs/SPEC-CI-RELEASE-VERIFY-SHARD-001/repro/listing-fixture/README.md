# listing-fixture — what `go test -list` enumerates (research.md L-E20)

Files carry a `.txt` suffix so no Go tool treats this directory as code. NOT executed at plan time (leader quiet window: no heavy Go execution).

Run-phase runbook (one Go call at a time, under the resource-entry rules):

1. Copy to a scratch directory outside the repository: `a_test.go.txt` -> `a_test.go`, `race_only_test.go.txt` -> `race_only_test.go`, `go.mod.txt` -> `go.mod`.
2. `go -C <scratch> test -vet=off -list '.*' ./...` — expected (measured once, L-E20): `TestReal Test BenchmarkB Fuzz FuzzReal Example ExampleLower ExampleBlock ExampleUnordered`; not listed: `Fuzzhelper`, `ExampleAlpha`; `TestRaceOnly` absent.
3. `go -C <scratch> test -race -vet=off -list '.*' ./...` — expected: the same names plus `TestRaceOnly` (the mutant of AC-RVS-002 (x): a plain listing misses it).
