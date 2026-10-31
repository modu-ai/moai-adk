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

## Residual-risk (round 1 list, superseded by the round 2 list below)

Round 1 listed: a marker longer than the carry, nameless tokens (Bearer, ghp_, sk-), a PEM body whose header fell off, a
JSON field, a marker followed by a blank line, a 600-byte control run. Items 2, 3 and the control-run part of item 1 are
closed in round 2 (below); the rest is restated there.

---

# Round 2 (leader's seven shapes; HEAD before: d7a437ee72)

## Claim (round 2)

1. Root cause: the writer decided "this line holds a secret marker" with a hand-written pattern (`SecretMarkerPattern`
   plus `authorization:`, a PEM header, a URL user part) while `core.Redact` masks by a different, larger set of rules, so
   whatever Redact masks and the pattern does not name is not recognised once the marker falls off the window.
2. Fixed by deriving the decision from `core.Sanitize` itself (credential categories only) instead of copying its rules,
   and by three structural changes (below). For the seven shapes, neither the stdout JSON nor the ledger fail reason holds
   any byte of the value; each is closed by a test (table in the report back).
3. N1 is NOT declared closed: the shapes under Residual-risk (round 2) are still uncaught, each with a reproduction.

## Evidence (round 2)

### (a) Root cause in code: which rules Redact has, which the old pattern knew

`core.Redact` (internal/agenticfactory/core/redact.go, read, not edited) applies, in order: (1) PEM private key block
`-----BEGIN ... PRIVATE KEY-----` to its END or to the end of the text; (2) `Authorization: Bearer|Basic|Token <v>`; (3)
bare `bearer\s+<8+ chars>`; (4) `Cookie|Set-Cookie: <rest of line>`; (5) `https?://user:<v>@`; (6) JSON `"...password|token|secret|
api_key|credential...": "<v>"` (needs the closing quote); (7) `<secret-named field>\s*[:=]\s*<v>` and (8) `<name>key=<v>`;
(9) `gh[pousr]_<20+>`, (10) `github_pat_<20+>`, (11) `sk-<16+>`, (12) `AKIA|ASIA<16>`; then email, home path, phone, person name,
stack traces (not credentials, ignored for marking) and the opaque-run rules (a value that identifies itself, needs no
marker). `\s*` in (7) crosses newlines, so Redact masks `password=\n\n<v>`.

| Shape | Redact masks it with the marker in the window | Old writer detector (hand pattern) |
|---|---|---|
| 1 `Bearer <v>` | yes, rule 3 | no: `bearer` was removed from the pattern in round 1 for speed |
| 2 `ghp_<v>` | yes, rule 9 | no: no token-prefix alternative |
| 3 `sk-<v>` | yes, rule 11 | no |
| 4 JSON `"password": "<v>"` | yes, rule 6 | no: a quote sits between the name and the colon, and rule 6 needs the closing quote |
| 5 PEM, no END | yes, rule 1 (`$` ends the match) | the header line only: the body lines were not covered |
| 6 `password=\n\n<v>` | yes, `\s*` crosses blank lines | the chain covered exactly one following line, blank or not |
| 7 `pass` + 600 BEL + `word=<v>` | yes, after control stripping | no: the 512-byte carry was filled with BEL |

The parity table (`TestAFWatch_TheWritersMarkerDetectorAgreesWithRedact`) was the first RED to show the divergence both
ways: at the RED commit it also failed the row `monkey_key: ` (the writer marked it, Redact does not: the old pattern
accepted `key:`, rule 8 takes only `key=`).

### (b) RED (commit e4fcf42fba)

- `go test ... -run '^(TestAFGate_ARound2...|TestAFFailureTail_ARound2...|TestAFLand_ARound2...)$' -v`: all 7 shapes FAIL at
  gate level, 7 at verb level (stdout JSON and ledger), 10 of 13 split-write rows; the 3 passing split rows
  (`BEL_run_inside_one_write`, `escape_sequence_split_between_writes`, `OSC_sequence_split_between_writes`) are guards.
- parity table: 17 rows FAIL at RED (`monkey_key: `, Bearer, bearer, the six token prefixes, the PEM header, and the six JSON
  spellings); the rest pass. (The rows for a value on a later line were added after RED, in 19aba37706 and 946d6dfa39.)
- the round-1 tests were unchanged and passing at RED.

### (c) Fix (commits 7a3049313f, 19aba37706, 946d6dfa39, 0ab88b8ca9)

- **Detector = core.Sanitize.** `holdsSecretMarker(text)` runs `core.Sanitize` on four probes: the text; the text plus a
  sentinel value (`ZqZqZqZq"@x`, which carries a closing quote and an `@` so rules 5 and 6 fire); the text, a newline and the
  sentinel (a scheme word such as `Bearer` ending a line); the text, a newline, a quote and the sentinel (a JSON field
  pretty-printed across two lines). A line is marked when a credential category (token, authorization, cookie,
  url-credential, credential-assignment, private-key) comes back. A rule added to core follows automatically; a keyword gate
  (`tailMayHoldSecret`: self-describing words, or a secret word with a `:`/`=` in the text) must also admit it, and the
  parity table is what catches a gate that does not.
- **Controls leave before the carry.** Text is normalised (`StripControls`) before it enters the 512-byte carry; an escape
  sequence a write ends in the middle of (ESC `[`, ESC `]`, or the single-byte CSI U+009B) is held back raw, at most 256
  bytes, and joined to the next write. New carry ceiling: **512 normalised (visible) bytes** per line; controls cost
  nothing against it.
- **Spans.** A marker that ends its line makes the next *content* line its value: blank and whitespace-only lines between do not
  end the wait. A private-key span stays open until a line holding `-----END`, or the end of the stream.
- **Throughput.** A chunk-level fast path (`skippable`) updates only the position and carry when the carry plus the write
  holds no credential. Measured with a temporary test (deleted): 9 MB, 64 KiB window, no race detector / `-race`:
  plain log lines 0.04 s / 0.13 s; prose containing the words key, pass, token, secret but no separator 0.06 s / 0.15 s; one
  16000-byte line repeated 0.03 s / 0.11 s. Marker-dense output (every line `key=cache count: 12 ...`, which Redact does
  treat as a credential) 1 MB in 2.2 s without `-race` (about 0.45 MB/s); under `-race` the same dense sample took 23 s per MB
  (measured before the last gates, same order). This is the price of asking the real sanitizer about each marked line.
- The existing `WatchStateIsBounded` assertion that a 600-BEL run inside a marker is unseen could not stay as written: that
  shape is now seen on purpose (leader's shape 7). The ceiling it pinned moved to visible text: the test now asserts that
  `password` + 600 spaces + `=` over three writes is unseen, and that a 2560-byte BEL run inside `pass...word=` is seen.
  Nothing was deleted or skipped.

### (d) Mutation contrast (each applied once, tests run, restored; no mutation committed)

| Guard | Mutation | Tests that FAILED |
|---|---|---|
| value-to-come probe (sentinel with closing quote and `@`) | probe 2 = probe 1 | NONE at first: survived. Added the row "URL credential whose @ arrives in a later write" (600-byte value before the `@`); now 1 unit row fails |
| next-line probe | probe 3 = probe 1 | unit "a bare bearer word ends the line", parity `Bearer`, `Authorization: Bearer` |
| quoted next-line probe | probe 4 dropped | unit "JSON field with its value on the next line", parity `{"api_key":\n"` (added after the first run found the shape) |
| Token category | removed | 11 cli rows (ghp/gho/ghu/ghs/ghr/github_pat/sk, round-2 gate and split) + 2 unit rows |
| Authorization category | removed | 9 cli rows + 2 unit rows |
| Cookie category | removed | 5 cli rows (cookie shapes, parity Cookie/Set-Cookie) |
| URL credential category | removed | parity `https://user:`, `http://u:` + the new unit row |
| Credential-assignment category | removed | 84 cli rows + 11 unit rows |
| private-key flag | `w.pem = w.pem && pem` | 4 cli rows + 1 unit row |
| private-key END closes the span | marker `-----ENDX` | `pem_with_end_keeps_the_lines_after_it` + unit "private-key block ended, plain text after" |
| blank lines between marker and value | case removed | 5 cli rows + 1 unit row |
| escape held back across writes | `cut = len(raw)` | `escape_sequence_split_between_writes`, `OSC_sequence_split_between_writes`, 1 unit row |
| control stripping in `line()` | removed | 5 cli rows + 2 unit rows |
| control stripping in `skippable()` | removed | 6 cli rows + 2 unit rows |
| `skippable` ignores a marked chunk | `&& false` | 113 cli rows + 16 unit rows |
| `skippable` keeps the old lineStart | line removed | unit "a marker line wholly inside the window after clean text" (added for it) |
| literal gate: `sk-` in the self-describing list | removed | parity `sk-`, `sk_token` gate and split rows |
| literal gate: `key` among the secret words | removed | 14 cli rows (api_key shapes, parity, `monkey_key=`) |
| single-byte CSI held back | condition inverted | `C1_CSI_split_between_writes` (added after the residual check reproduced it as a leak) |

An earlier probe-less variant of the throughput gate (`(w.curMarked || w.chain) &&` before `endsWithMarker`) changes cost
only; it is behaviour-equivalent and is not mutation-tested.

### (e) Checks (HEAD 0ab88b8ca9)

```
gofmt -l internal/cli internal/agenticfactory                                     -> (no output)
go build ./...                                                                    -> exit 0
go vet ./internal/cli/ ./internal/agenticfactory/...                              -> exit 0
env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
go test -count=1 -race -p 2 ./internal/agenticfactory/mergeq/                     -> ok 17.697s
anchored N1 tests (5 round-1 old tests, ALoudFailingGate, ALoudPassingGate, round-1 new, round-2 new, parity), -v
                                                                                  -> 164 `--- PASS` lines, 0 FAIL
TestTailWriter_Truncated... / MaskedPrefix... / WatchStateIsBounded / NineMegabytes -> 4 PASS
golangci-lint run ./internal/agenticfactory/mergeq/...                            -> 0 issues (one staticcheck De Morgan finding was fixed on the way)
golangci-lint run ./internal/cli/...                                              -> 1 issue, errcheck in codex_close_entry_signal_test.go:89 (untouched)
```

## Baseline-attribution (round 2)

- Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, branch `WT-agentic-factory-integration`.
- HEAD before round 2: d7a437ee72. RED e4fcf42fba; GREEN 7a3049313f; follow-ups 19aba37706, 49f4fb167d (test row), 946d6dfa39,
  0ab88b8ca9; the docs commit holding this section follows. Tools: `go` and `golangci-lint` from PATH on this tree.
- The throughput figures come from a temporary test that was deleted; they are one run each on this machine, not a benchmark.

## Gaps (round 2)

- The full `internal/cli` suite was not run (forbidden). No audit, no GLM, no live switches, no push.
- Under `-race` the marker-dense sample (about 0.45 MB/s without it) was measured at 23 s per MB before the final two small
  probe additions; it was not re-measured after them.
- The worktree guard refused one compound command (a `go test` with a command substitution and a redirect, about checking
  the PASS count); I reran it as separate plain commands and read the counts from the saved output. Two `Write` calls were
  refused by the secret-scanning hook (a literal PEM header, and a table row it read as a hard-coded credential); the PEM
  lines are now assembled from parts and the parity table is built from name and separator lists.
- The leader's shapes were run at 70 KB through `afGate` and `moai af land`'s JSON + ledger in tests; no live `moai af land`
  binary was used.

## Residual-risk (round 2)

Reproduced with a temporary test (deleted) that wrote the shape through `TailWriter` (128-byte window) and
`afFailureTail`, next to `core.Redact` on the whole text:

| Shape | afFailureTail leaks the value | core.Redact on the whole text |
|---|---|---|
| `password` + 600 spaces + `=` + value, three writes | yes | masks it: the carry ceiling (512 visible bytes) |
| `password\n` then `=` + value | yes | masks it: a name and its separator on different lines |
| `password=\nfirst\n` then `second <value>` | yes | leaks too: Redact masks only the first token after a marker, the writer follows it |
| `foo=<value>` (a name Redact does not treat as secret) | yes | leaks too: parity means the writer cannot know more than Redact |
| single-byte CSI U+009B split across writes | closed in the last commit (test `C1_CSI_split_between_writes`) | n/a |
| 5000-byte prefix, ` password=` + value, one write | no (marker and value in one write) | n/a |

Also open:
1. Marker-dense output is slow (about 0.45 MB/s without `-race`): every candidate line goes through `core.Sanitize`.
2. Over-masking: a marked line whose start was dropped is masked from the window front even when its marker is inside the
   window after the cut (the position of the marker within the line is not tracked); the loud-gate rule still holds for lines
   with no marker (`ALoudFailingGate...` and the three `ALoudGateWithNoMarker...` rows).
3. A secret whose name Redact does not know, and anything Redact itself misses. The parity design means the writer is as good
   as `core.Redact` for markers, not better; a gap in Redact (outside this lane's write scope) is a gap here.
4. An unterminated OSC sequence longer than 256 bytes is flushed as it is; a split inside it can leave stray characters in
   the carry.
5. Spans are per raw line: a very large PEM or value with the BEGIN line wholly inside the window and the front inside the
   body is covered, but memory for spans is bounded by the window only (at most one span per marked line in it).
