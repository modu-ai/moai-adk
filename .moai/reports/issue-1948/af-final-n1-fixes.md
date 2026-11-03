# Final audit of a1ec370195 — N1 repair (lane-22)

Commit graph (verification-claim-integrity.md §2.3): RED `69d3a830fd` → GREEN `b1ed0f42f2`. Base: `a1ec370195`.
Tool-provenance (§2.2): every `go test` below was compiled from this worktree at the commit named beside it.

## Claim
- N1 (BLOCKING): the failure tail of `moai af land --gate` is built as: drop a cut first line when other lines follow →
  remove escape sequences and control characters whole → join a secret-named marker that ends its line to the next
  line → `core.Redact` → cut to the last 2048 bytes. A marker or value split by a control character or an escape
  sequence is masked; the value on the line after a `Cookie:` line is masked; a value whose marker fell off the kept
  64 KiB is not shown.
- `mergeq.TailWriter.Truncated()` added (true once the start of the output was dropped).

## Evidence
RED at `69d3a830fd` (the old `af_land.go` of `a1ec370195` plus the committed test; anchored `-v`, scratch
`n1_red_corrected.out`):

```
--- FAIL: TestAFLand_ASecretSplitByAControlOrAnEscapeIsStillMasked (5.21s)
    --- FAIL: .../value_split_by_CR (0.34s)
    --- FAIL: .../value_split_by_FF (0.43s)
    --- FAIL: .../marker_split_by_BEL (0.36s)
    --- FAIL: .../marker_split_by_NUL (0.36s)
    --- FAIL: .../marker_coloured_equals (0.33s)
    --- FAIL: .../marker_coloured_colon (0.47s)
    --- FAIL: .../marker_coloured_only_reset_before_sep (0.37s)
    --- FAIL: .../authorization_split (0.31s)
    --- FAIL: .../pem_header_split (0.29s)
--- FAIL: TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt (0.37s)
--- FAIL: TestAFLand_AValueWhoseMarkerFellOffTheKeptTailIsNotShown (0.38s)
--- PASS: TestAFLand_AValueStraddlingTheCutAndControlsBeyondItStayMasked (0.50s)
```

(5 matrix cases already masked before the repair: plain assignment, value split by BEL, value split by an escape
sequence, whole coloured value, URL credential. 14 cases = 9 FAIL + 5 PASS.)

GREEN at `b1ed0f42f2` (same commands, scratch `n1_green.out`, plus the 37th-review control-character test):

```
--- PASS: TestAFLand_ControlCharactersInAFailedGatesOutputAreNotKept (0.31s)
--- PASS: TestAFLand_ASecretSplitByAControlOrAnEscapeIsStillMasked (5.14s)   [14 sub-cases PASS]
--- PASS: TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt (0.49s)
--- PASS: TestAFLand_AValueWhoseMarkerFellOffTheKeptTailIsNotShown (0.43s)
--- PASS: TestAFLand_AValueStraddlingTheCutAndControlsBeyondItStayMasked (0.38s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.662s
```

Also at the final tree: `go test -count=1 -v -run '^(TestAF|TestCardFlow)' ./internal/cli/` → `ok 20.345s`, no `--- FAIL`;
`go test -count=1 -race -p 2 ./internal/agenticfactory/mergeq/` → `ok 23.700s`;
`TestTailWriter_TruncatedSaysWhetherTheStartWasDropped` PASS; `go build ./...`, `go vet ./internal/cli/ ./internal/agenticfactory/...`,
`env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...` and `gofmt -l` on the touched files print nothing.

Mutation controls (disable, observe FAIL, restore): without the marker join `TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt` fails
("the value on the line after a Cookie line survives"); without the fragment drop `TestAFLand_AValueWhoseMarkerFellOffTheKeptTailIsNotShown` fails.

## Baseline-attribution
Real ledger, real git repositories, the real `moai af land` verb, built from this worktree. The auditor's reproduction file was
copied in temporarily to read its baseline (11 FAIL: 9 matrix + Cookie + lost marker) and removed before any commit.

## Corrections made along the way (named, not hidden)
- The first RED commit's CR and FF cases read the whole reason, which quotes the gate command (`B=super`), so they
  failed for a reason that is not output. The test now reads only the part after `exit status 1: `. The RED commit was amended
  (not pushed) and RED was re-observed with the old code and the corrected test (output above). The auditor's own
  `value_split_by_CR/FF` verdicts used the same whole-reason read.
- A first version of the fragment drop removed the only line of a 9 MB newline-less output and broke
  `TestAFLand_ALoudFailingGateKeepsItsLastLine`; the drop now applies only when other lines follow the fragment.

## Behavior changes (named)
(a) A marker that ends a line is joined to the next line in the recorded failure tail, so the line break between them is gone.
(b) The first kept line of a truncated tail is not shown when other lines follow it. (c) An escape sequence is removed whole,
where only its ESC byte went before.

## Gaps
- Not measured: a single-line value longer than 64 KiB with no newline after it and a shape no `core.Redact` rule knows
  (the fragment is then the gate's last line and stays). Residual risk below.
- The editor's diagnostics showed `undefined: core.StageMerging`, `c.Epoch`, `st.ReapLocks` for `af_land.go` during the work;
  `go build ./...`, `go vet` and the windows cross-build print nothing, so they are not compiler errors of this tree (editor snapshot).
- The security hint on the new test (`hardcoded-secret`) is the fixture value `hunter2`, a test input.
- Full `internal/cli` and `-race` over `internal/agenticfactory/...` not re-run here (the leader's single full run follows the new SHA).
- N2 (lane-18) is not merged at this point; this file is extended after it is.

## Residual-risk
A secret value that is longer than the kept 64 KiB, sits on one line with nothing after it, and has a shape no mask rule
recognises, is still shown in its last 2048 bytes: its marker is gone and nothing in the text says it is a secret.
`core.Redact` is lane-18's and is unchanged.

---

# Continuation 2026-11-01 (lane-2): verdicts A·B·C and the N2 merge

All commands ran in
`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, branch `WT-agentic-factory-integration`, start HEAD b1ed0f42f2, dirty 0.
This section supersedes the line "N2 (lane-18) is not merged at this point" above.

## Claim / Evidence / Baseline
| Claim | Command (this run, this tree) | Observed |
|---|---|---|
| A RED before fix | `go test -count=1 -run '^TestCleanReportsArchive_MovesOnlyEligibleCandidates$' -v ./internal/cli/` | `clean_reports_archive_test.go:121: no new YYYY-MM shard created under archive/` `--- FAIL` |
| A GREEN after fix (commit 72c02a4d9d) | same command | `--- PASS: TestCleanReportsArchive_MovesOnlyEligibleCandidates` |
| B RED before fix | `go test -count=1 -run '^TestAutoPickMirrorParity$' -v ./internal/cli/` | `live and mirror differ on 0 lines, want exactly 1` |
| B GREEN after fix (commit 29db2ca756) | same command | `--- PASS: TestAutoPickMirrorParity` incl. subtest `factory-dispatch.md_differs_by_exactly_the_live-only_sentence` |
| C RED before fix | `go test -count=1 -run '^TestLaneChoiceMergeOffer$/^choosing_m' -v ./internal/cli/` | `spy entered 0 times, want 1` (the chooser never opened: the stamp-dead run was merged automatically) |
| C GREEN after fix (commit dc387ee2ed) | `go test -count=1 -run '^TestLaneChoiceMergeOffer$' -v ./internal/cli/` | `--- PASS: TestLaneChoiceMergeOffer`, subtest `choosing_m<n>_moves_the_cards,_retires_the_source,_and_joins_the_live_run` PASS (fixture: pid-0 run + dead role='leader' peer, BasisPeer; assertions unchanged) |
| N2 merged once | `git merge-tree --write-tree HEAD 2e5e2ee2a68add54f9485fb8e554186e48bdf259` then `git merge --no-ff 2e5e2ee2a6` | tree efe562bb8c, exit 0, no conflict; merge commit 9be13361fb (full 9be13361fbb10836da3d688ac547110472b6928b) |
| Template and budget tests, run on commit B before the merge | `go test -count=1 ./internal/template/` ; `go test -count=1 -run 'Budget\|RoleCore\|RuleFile\|Mirror' ./internal/cli/` | `ok ... 219.623s` ; `ok ... 13.732s` |
| Post-merge checks | `go build ./...` ; `go vet ./internal/cli/ ./internal/agenticfactory/...` ; `env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...` ; `gofmt -l internal/cli internal/agenticfactory` | build exit=0 ; vet exit=0 ; winbuild exit=0 ; gofmt prints nothing |
| Post-merge anchors | `go test -count=1 -v -run '^(TestCleanReportsArchive_MovesOnlyEligibleCandidates\|TestLaneChoiceMergeOffer\|TestAutoPickMirrorParity\|TestAFLand_ASecretSplitByAControlOrAnEscapeIsStillMasked\|TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt)$' ./internal/cli/` | exit=0, five `--- PASS` lines, `ok ... 4.695s` |
| mergeq | `go test -count=1 -race -p 2 ./internal/agenticfactory/mergeq/` | `ok ... 17.630s`, exit=0 |

## Out of scope and left as found
- `clean_reports_archive_test.go` line 154 (`archive/2026-06` stat) is not changed, as instructed.
- N3 (`TestA1Done_WithdrawnLanded_ReclaimedStrands`) and N4..N10, F8..F12 are not touched.

## Gaps (this section)
- Refused commands: the worktree-isolation guard refused (1) a `sed` whose file argument was a shell variable, (2) a compound
  `gofmt && go test | grep` line, (3) a heredoc `python3` edit. None ran. The same edits were redone as a plain `sed` with a literal
  path (verified by `git diff`), and as two `Edit` tool calls (verified by gofmt and the test run). No check rests on a refused command.
- A zsh `PIPESTATUS` read printed an empty exit code for the first build/vet/windows run; those three were re-run without a pipe and
  report exit=0 above. The first run's empty value is not counted as a pass.
- Not run in this section: `golangci-lint run`, the full `internal/cli` suite, `-race` over `internal/agenticfactory/ledger/...` and
  `.../core/...` after the merge (lane-1 reports its own run; the leader's archive run of 120a3cf4b0 is the leader's evidence, not observed here).
- No separate mutation check beyond the observed RED-then-GREEN of each fix (the old fixture / old template line is the failing control).
- The editor diagnostics list (unusedparams, writestring, scannererr) concerns files this work did not edit.

## Open item (NOT closed)
N1 residual risk stays open. Leader measurement on the merged tree 5b8719910b: `TestA1Secret_MarkerLostToTheTailWriter` FAIL
(`stdout has value=true ledger has value=true len(tail)=2048`): a 70 KB one-line value whose marker was cut keeps its last 2048 bytes in
stdout and the ledger. It is not repaired here (hiding a fragment collides with `TestAFLand_ALoudFailingGateKeepsItsLastLine`) and is
carried as an open risk in the AF completion report. Not observed by lane-2 in this run.
