# N1 — a secret value longer than the kept gate output window (issue-1948, AF-22)

## Claim

1. Hypothesis confirmed in code: `mergeq.TailWriter` dropped the start of a gate's output before any masking could see
   it, so a `password=` marker outside the kept 64 KiB left the value's last bytes (cut to 2048) in the stdout JSON and
   in the ledger fail reason.
2. Repaired: the writer now watches the stream in bounded state for lines a secret marker introduces and reports how
   many leading kept bytes are such a value without its marker (`MaskedPrefix`). `afFailureTail` replaces them with
   `[a secret value was masked: its marker is outside the captured output]`. A line with no marker is unchanged, so a
   loud failing gate still shows its last line.
3. N1 is NOT declared closed: the shapes listed under Residual-risk are still uncaught.

## Evidence

### Every path gate output can leave the process (read from callers)

`afGate` (internal/cli/af_land.go) is the only producer; `TailWriter.Bytes` is read only by `afFailureTail`.

| # | Path | Where |
|---|---|---|
| 1 | error string `gate "<redacted gate>": <exit>: <tail>` returned from the `mergeq.Verify` func | af_land.go `afGate` |
| 2 | `Merger.Land` wraps it as `gates failed: ...` and `fail()` writes it to `merge_queue.fail_reason` (ledger) | mergeq.go:194-196, 363-369; ledger/queue.go:332 |
| 3 | the same string as `Result.Reason`, printed in the stdout JSON line `reason` | af_land.go:241-244 |
| 4 | `notes` row (`kind=decision`) that appends `failReason` when a landing row of an older card epoch is settled as failed | ledger/queue.go:341-347 (same string, same sanitised tail) |
| 5 | the timeout wrapper `%w (the gate ran past ...)` adds only the error, not output | af_land.go:99-101 |

Nothing else reads `Bytes()`; no log or event writer takes the tail. All of 2-4 carry the string built by path 1, so one
repair point (`afFailureTail`) covers them. The gate command text is quoted redacted (`afRedactedGate`), unchanged.

### Reproduction from a repository test (RED, commit 807fe22d94)

`go test -count=1 -p 2 ./internal/cli/ -run '^(TestAFGate_A...|TestAFFailureTail_A...|TestAFLand_ALongValue...)$' -v` before the fix:

- new secret cases FAIL: `a piece of the value survives (2048 bytes of tail)`, 17 table rows + 8 split-write rows + 3 verb
  rows (stdout JSON and ledger both hold the value, `len=2048`); 17 rows also lacked the "masked" notice (one,
  `other_lines_follow`, failed only on the notice: the old fragment-drop rule already hid the value there).
- guards PASS before the change: `window_inside_by_one`, `window_exactly_full`, `window_value_line_untruncated`, the 3
  `ALoudGateWithNoMarker...` rows, `no_marker_keeps_the_last_line`.
- old tests PASS at RED: `TestAFLand_ASecretSplit...`, `..._ACookieLine...`, `..._AValueWhoseMarkerFellOff...`,
  `..._AValueStraddling...`, `..._ALoudFailingGateKeepsItsLastLine`, `TestTailWriter_TruncatedSaysWhetherTheStartWasDropped`.

### After the fix (HEAD abb2feda48)

```
go test -count=1 -p 2 ./internal/cli/ -run '^(<the 5 old tests + ALoudPassingGateStillLands + the 4 new tests>)$' -v
  -> 52 `--- PASS` lines, 0 FAIL, ok  github.com/modu-ai/moai-adk/internal/cli  7.234s
go test -count=1 ./internal/agenticfactory/mergeq/ -run '^(TestTailWriter_Truncated...|..._MaskedPrefix...|..._WatchStateIsBounded)$' -v -> 3 PASS
gofmt -l internal/cli internal/agenticfactory                    -> (no output)
go build ./...                                                    -> exit 0
go vet ./internal/cli/ ./internal/agenticfactory/...             -> exit 0
env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
go test -count=1 -race -p 2 ./internal/agenticfactory/mergeq/     -> ok 17.648s
golangci-lint run ./internal/agenticfactory/mergeq/...            -> 0 issues
golangci-lint run ./internal/cli/...                              -> 1 issue, errcheck in codex_close_entry_signal_test.go:89 (not a file touched here)
```

New table sweep: `TestAFGate_ALongValue...` has 22 rows (13 marker shapes, 8 around the 64 KiB edge for the 9-byte marker
`password=`, 1 long-output-follows); `AMarkerSplitAcrossWrites` 9 split-write rows + 1 no-marker row; verb test 3 rows.

### Design

`TailWriter` additionally keeps: a carry of the last `tailCarry` = 512 raw bytes of the line being written (a marker
that two writes, or a control or escape sequence, split is rejoined and normalised with `StripControls`), and one span
`[line start, line end)` per marked line still in the window (a span runs over the next line when the marker ends its
line). `MaskedPrefix` = kept bytes from the window front to the end of the first span that starts before the front.
Memory is bounded by the window (spans for lines inside it, at most one per 3 bytes) plus the 512 carry. The reason
text, including the notice, never exceeds `afGateTail` = 2048. `StripControls` and the marker core pattern were moved into
`mergeq` (`SecretMarkerPattern`, `StripControls`) and `af_land.go` now builds `afMarkerAtLineEnd` and `afStripControls`
from them (one source). Throughput: a marker prefilter (`:`, `=`, `-`) and a printable-ASCII fast path keep a 9 MB
marker-free stream at 0.07 s under `-race`; marker-bearing text goes through the regexp until a line's first marker.

### Mutation contrast (each applied once, test run, restored; no mutation committed)

| Guard | Mutation | New tests that FAILED |
|---|---|---|
| span starts before front | `s.start < front` -> `>` | 8 unit rows + 27 cli subtests |
| span end ahead of front | remove `front < s.end` | none: it is implied by `prune` (equivalent mutant), so the guard was deleted (follow-up commit) |
| `prune` | never prune | `MaskedPrefixCovers...` ("-30, want 0") and `WatchStateIsBounded` (3000 spans) |
| chain over the value line | marker ending a line closes its span | 2 unit rows, 4 table rows, 2 split-write rows, 1 verb row |
| carry between writes | carry not kept | 2 unit rows, 4 split-write rows |
| control/escape normalisation | match on raw text | 1 unit row, `marker_split_by_BEL`, `marker_coloured`, split-write BEL row |
| marker prefilter | drop `=` from the byte set | 8 unit rows, 15 cli subtests |
| notice replaces the prefix | `lost` forced to 0 | 30 cli subtests |
| 2048 cap with the notice | cap ignores the notice length | first run: none (body was empty in every row); added `long_output_follows` + a length assertion, re-run: FAILS (`tail is 2119 bytes`) |

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, branch `WT-agentic-factory-integration`.
- HEAD before: 3868473f47. RED 807fe22d94, GREEN f96e514369, follow-up abb2feda48; the docs commit holding this file follows.
- Tools judging the tree: `go` and `golangci-lint` from PATH, run on this tree; no `moai` binary was used as a measurement.
- Window arithmetic: `afGateKeep` = 65536, marker `password=` = 9 bytes.

## Gaps

- The full `internal/cli` suite was not run (forbidden here); only the anchored tests above, plus `go build ./...`.
- `golangci-lint run ./internal/cli/...` ran for the whole package; its single finding is in an untouched file.
- No independent audit, no GLM, no live switches, no push.
- The worktree guard did not refuse any command I issued; one compound heredoc-and-sed command was refused on first try and
  redone with the Edit tool (no measurement was substituted).

## Residual-risk

1. A marker whose raw form spans more than 512 bytes across writes (unit test `WatchStateIsBounded` pins this: 600
   control bytes inside `pass...word=` split over three writes are not recognised).
2. A value with no recognisable name: `ghp_...`, `sk-...`, an unnamed AWS secret, or `Bearer <token>` with no header, whose
   start was dropped. Such tokens self-identify only while their prefix is in the window.
3. A PEM private key body whose `-----BEGIN` header is outside the window (only the header line is a marker span; the
   body lines are separate lines).
4. An unterminated OSC escape sequence that swallows newlines; spans are per raw line.
5. Over-masking: a marked line whose start was dropped is masked from the window front even when its marker lies inside
   the window after the cut (the marker position within the line is not tracked); the line then reads as the notice.
6. A marker such as `x-api-token : value` with the separator on the next line, or a value two lines after a
   marker-ending line (only the next line is covered).
7. Marker-bearing text is scanned with a regular expression until a line's first marker (about 7 MB/s measured on
   marker-free `x` text without the prefilter); log-like output with many `:` or `=` is slower than before this change.
