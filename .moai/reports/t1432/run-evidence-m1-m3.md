# Run-phase evidence, milestones M1 to M3 (card t1432, SPEC-HARNESS-RETENTION-HARDEN-001)

Platform for every run below: darwin, uid 501, go1.26.8. Linux and Windows runtime: not observed. Windows evidence is `GOOS=windows go build` and `go vet` only. Tests were run with the kanban variables unset in the same invocation. Baseline commit T = `ac40cf3bf` (observed RED).

## M1 — tail-carry and log terminator (REQ-HRH-006, -007, -008; AC-HRH-007, -008, -009)

Tree measured: HEAD `ac40cf3bf` plus the uncommitted M1 edit of `internal/harness/retention.go` (committed as the M1 commit that carries this section).

Claim 1: AC-HRH-007 and AC-HRH-008 now pass and the quiescent pins stay green.

- Command: `go test -count=1 -v -run '^(TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval|TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched)$' ./internal/harness/`
- Output (the deciding lines, exit 0):

```
--- PASS: TestPruneNothingStaleLeavesLogUntouched (0.00s)
--- PASS: TestPruneKeepsUnparsedLinesVerbatim (0.00s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/wrong-field-type (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/truncated-json (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/plain-text (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/json-array (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/padded-text (0.01s)
--- PASS: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.01s)
--- PASS: TestPruneCarriesLateEvents (0.31s)
--- PASS: TestPruneTailPartialLineCarriedAndTerminated (0.31s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.945s
```

Claim 2: the whole package passes except the two tests that are M2-owned RED by design (AC-HRH-001, AC-HRH-002).

- Command: `go test -count=1 -v ./internal/harness/` redirected to a scratch file: exit 1, 271 `--- PASS` lines, exactly two `--- FAIL` lines: `TestPruneStateUnwritableFileReplaced` and `TestPruneStateSymlinkReplacedTargetUntouched` (the two M0 RED tests that M2 flips).
- Command: `go test -count=1 -skip '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced)$' ./internal/harness/`
- Output: `ok  	github.com/modu-ai/moai-adk/internal/harness	1.218s`, exit 0.
- Gap: the instruction to commit M1 "when the whole package passes" cannot be met literally, because the two M0 REDs belong to M2 (`plan.md` M2 step 5). The M1 commit is gated on the package minus those two named tests.

Claim 3: vet, gofmt and Windows compile are clean after M1.

- `go vet ./internal/harness/` exit 0; `gofmt -l internal/harness/` empty; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.

Mutation runs (mutants built from the CURRENT `retention.go` as scratch copies outside the tree, applied with `go test -overlay`; each removed afterwards, `git status --short` shows only `M internal/harness/retention.go`). Command shape: `go test -count=1 -overlay <overlay> -run '^(TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval)$' ./internal/harness/`

| Mutant | Killed by | Exit |
|---|---|---|
| verbatim tail copy with no terminator (the `if tail[len(tail)-1] != '\n'` block removed) | `TestPruneTailPartialLineCarriedAndTerminated` (`replacement log does not end with a newline`; the next append glued onto the fragment) and `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` (`rewritten log does not end with a newline`) | 1 |
| classifying the final unterminated line (the unterminated-line guard in `scanTerminatedLines` removed) | `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` (`stale-final count after first prune = 0, want 1`) | 1 |
| dropping the tail copy (the `appendLogTail` call removed) | `TestPruneCarriesLateEvents` (`late-event count = 0, want 1`), `TestPruneTailPartialLineCarriedAndTerminated` (`partial fragment occurs 0 times, want 1`), `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 1 |

Gaps (M1): whether a concurrent `O_APPEND` write can be seen half-complete was not observed (the boundary rule is tested deterministically). The residual window between the final tail reading and the rename is not closed and not measured. The `-race` run is recorded once after M3.
