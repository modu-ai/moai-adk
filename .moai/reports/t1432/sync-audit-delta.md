auditor-model: claude-sonnet-5-5[1m]

verdict: PASS-WITH-DEBT
audited_sha: 030f09cbb8d4aae83b4778a784a8e6737e978da2

# Sync Audit, Delta: SPEC-HARNESS-RETENTION-HARDEN-001 (card t1432)

Scope: the delta `5bb35abe8..030f09cbb` over the full audit `.moai/reports/t1432/sync-audit.md` (PASS-WITH-DEBT 88 at `5bb35abe8`). The full audit is not re-done; its findings F2 and F3 are re-checked against the new tests, and its F1 is re-checked against the new residual-risk record.

Overall: PASS-WITH-DEBT. No blocking finding. F2 (heal wiring) and F3 (raw tail bytes) are closed for the exact mutants they named (L4, U7), each killed by its new test and by that test alone. Four new same-path mutants were tried; three survive (N1, N2, N4), each a narrower sibling of the closed finding. The residual-risk record is accurate on the window, its measurement and the repair analysis, but its "consequence, bounded" paragraph understates the consequence the SPEC itself names (D1) and carries one reasoned-not-observed claim stated as fact (D2). No re-score of the four dimensions is issued: nothing outside the two test files changed, and the 88 stands for every dimension this delta did not touch; the Craft evidence improved by two killed mutants.

## Dimension note (delta only)

| Dimension | Delta effect | Evidence |
|-----------|--------------|----------|
| Functionality | unchanged (production code byte-identical) | `git diff --quiet 5bb35abe8 HEAD -- retention.go retention_owner_unix.go retention_owner_windows.go` exit 0; `git diff --stat 5bb35abe8 HEAD -- internal cmd` names only the two test files |
| Security | unchanged; F1 stays a recorded residual | see D1, D2 |
| Craft | improved: F2 and F3 mutants now killed | E2, E4 |
| Consistency | unchanged; gofmt, vet, lint, Windows build and vet clean | E6 |

## Findings (structured defect list)

None is blocking. Confidence is stated per finding.

- D1 [medium] [optional: correct the record before it is relied on] `.moai/reports/t1432/residual-risk-removal-window.md:18` - the "Consequence, bounded" paragraph says the loss of the winner's fresh file "does not destroy user data: it can let a second pruner run once in that interval". `spec.md` §B D4.c names the consequence as two pruners, "the t1425 hazard" (line 131), and says an unconditional removal "would re-open concurrent appends to one archive file" (line 136); the full audit's own residual-risk line says "bounded to one interval and to the t1425 double-archive hazard" (`sync-audit.md:261`). The record's wording drops the archive consequence and states a no-data-damage conclusion that nothing in this card measured. Confidence: high that the wording is narrower than the SPEC and the audit; not measured here whether concurrent pruners lose or duplicate bytes. Required fix: reword the sentence to name the SPEC-stated consequence (a second pruner in the interval, duplicate archive entries, concurrent appends to one archive file) and mark the bound as reasoned, not observed.
- D2 [low] [optional] `.moai/reports/t1432/residual-risk-removal-window.md:18` - "a foreign-owned swapped-in entry cannot be removed by this path (... directory permissions or the sticky bit refuse it)" is stated as fact. It was inherited from the full audit's F1 text, which also stated it without observation, and it cannot be observed here (a non-root user cannot create a foreign-owned file; the card's own tests reach foreign ownership only through the injected `ownerCheck`). It is also not universal: a user who owns the directory may unlink any entry in it, sticky bit or not. Confidence: medium. Required fix: say "refused where the directory is not owned by the current user (permission and sticky-bit rules); not observed".
- D3 [low] [optional] `.moai/reports/t1432/residual-risk-removal-window.md:11` and `CHANGELOG.md` (the one changed sentence) - the figure "5 of 5235, about 0.1 percent" is true of that probe design only. This audit's independent probe (E5) measured 235 of 20000 (about 1.2 percent) with a differently started swapper, a factor of about twelve. Both race the window on purpose and neither is a production rate, which the record already says; the order of magnitude, however, is probe-dependent. Confidence: high. Suggested fix: state "reproduced twice, 0.1 and 1.2 percent per trial under two deliberately racing probes" or drop the single figure; `spec.md` D4.c and §F still say "not measured" (the SPEC body is frozen by design to keep the plan-artifact hash), and the record does not say its figure supersedes that wording.
- D4 [low] [optional] `internal/harness/retention.go:272` - mutant N1 (the identity comparison in `removeStateEntryIfUnchanged` reduced to modification time only) survives the whole package (`ok ... 2.888s`). It violates D4.c ("same file identity, same type, same permission mode, same modification time") and REQ-HRH-005. Both new tests and the older helper test use a fresh file whose mtime differs from the link's, so a coarse-timestamp file system is the only place the mutant would show. Confidence: high. Required fix (cheap): in `TestHealDoesNotRemoveAFreshStateFile`, `os.Chtimes` the fresh file to the inspected entry's modification time before the removal call.
- D5 [low] [optional] `internal/harness/retention.go:254` - mutant N4 (the "file" arm of `healStateEntry` calls `os.Remove` directly; the symbolic-link arm keeps the conditional removal) survives (`ok ... 2.792s`). The new heal test drives only the symbolic-link arm, so "pins the wiring from `healStateEntry` into `removeStateEntryIfUnchanged`" (`run-evidence-f2-f3.md` Claim 1) holds for one of the two callers. Confidence: high. Required fix: a file-arm variant (an unwritable regular file at the state path, root skip, the same swap interposition in `ownerCheck`).
- D6 [low] [optional] `internal/harness/retention.go:525-527` (`appendLogTail`) - mutant N2 (a newline written when the tail is empty) survives (`ok ... 1.976s`). The new tail test covers only a non-empty, newline-terminated tail; the empty-tail case, the common production path with no late events, has no raw-byte pin. The effect is a blank line per prune, dropped at the next rewrite, so it is bounded and cosmetic like F3. Confidence: high. Required fix: compare the replacement log's raw bytes in a prune with no late event.
- D7 [info] `internal/harness/retention_tail_test.go:111-123` - the new tail test is timing-reliant like its siblings: `blockedPruner` settles 300 ms, and nothing asserts the pruner had read the log before the late append. On a machine slow enough that the pruner reads after the append, the late event is classified instead of carried and the test could pass without exercising the tail path. Not reproduced (background load is prohibited here); it passed 15 of 15 under `-race`. The heal test has no such dependency: the interposition runs in the same goroutine and a `freshInfo == nil` guard fails the test if the owner check never ran.
- D8 [info] `internal/harness/retention.go:196-197` - the `@MX:REASON` says "a concurrent healer's fresh state file must survive" with no window caveat; the window is disclosed in `spec.md`, the CHANGELOG and the record, not at the code site. REQ-HRH-009..011 do not require a code-site disclosure for it. No action required.
- D9 [info] The `audit_multi` result of the full audit (codex FAIL, required backend) was not re-run and is not cleared by this delta; see "F1 representation".

## Claim

1. The tree audited is `030f09cbb8d4aae83b4778a784a8e6737e978da2`, branch `WT-harness-retention-debt`, toplevel `.moai/worktrees/t1432`, clean before and after (the scratch mutants and the probe lived only in the session scratchpad).
2. Production code is unchanged since `5bb35abe8`; `internal/lockfile` and `observer.go` are unchanged since the base `1e2151a38`; the SPEC directory is unchanged since `5bb35abe8`.
3. The two new tests pass, are race-clean over 15 runs, and the package stays green.
4. L4 is killed by the heal test alone and U7 by the tail test alone, for the stated reasons, and neither test is vacuous or parallel-unsafe.
5. Three of four new mutants survive (D4, D5, D6); one is killed (N3).
6. The residual-risk record is truthful on the window, the measurement and the repair analysis, with the exceptions D1, D2, D3.
7. The status transition edges are intact.

## Evidence

All commands ran in this session. Test runs used one compound invocation `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ...` (prefix omitted below) under the lease `moai slot acquire --resource go-test-internal-harness --max-duration 20m`, released afterwards (`slot go-test-internal-harness released`, then `free`).

### E1 Tree identity and production-unchanged checks

```
git rev-parse --show-toplevel -> /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432
git branch --show-current     -> WT-harness-retention-debt
git rev-parse --short HEAD    -> 030f09cbb   (full 030f09cbb8d4aae83b4778a784a8e6737e978da2)
git status --short            -> (empty); empty again after every run
git diff 5bb35abe8 HEAD --stat ->
 .../reports/t1432/residual-risk-removal-window.md  |  32 +++
 .../reports/t1432/run-evidence-f2-f3.md          |  91 +++++++
 .moai/reports/t1432/second-review.jsonl            |   1 +
 .moai/reports/t1432/sync-audit.md                  | 276 +++++++++++++++++++++
 CHANGELOG.md                                       |   2 +-
 internal/harness/retention_owner_test.go           |  64 +++
 internal/harness/retention_tail_test.go            |  43 +++
 7 files changed, 508 insertions(+), 1 deletion(-)
git diff --quiet 5bb35abe8 HEAD -- internal/harness/retention.go internal/harness/retention_owner_unix.go internal/harness/retention_owner_windows.go -> exit=0
git diff --quiet 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go -> exit=0
git diff --quiet 5bb35abe8 HEAD -- .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001 -> exit=0
git diff --stat 5bb35abe8 HEAD -- internal cmd -> retention_owner_test.go 64 +, retention_tail_test.go 43 +, 2 files changed, 107 insertions(+)
```

### E2 The two new tests and the package

```
go test -count=1 -v -run '^(TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal|TestPruneTailAlreadyTerminatedGetsNoExtraNewline)$' ./internal/harness/
--- PASS: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.01s)
--- PASS: TestPruneTailAlreadyTerminatedGetsNoExtraNewline (0.33s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.941s

go test -race -count=15 -run '<the same two names>' ./internal/harness/   -> exit=0
ok  	github.com/modu-ai/moai-adk/internal/harness	6.589s

go test -count=1 ./internal/harness/                                       -> exit=0
ok  	github.com/modu-ai/moai-adk/internal/harness	2.173s
```

(The first run was piped through `tail -n 20`; the output is 10 lines, so nothing was cut, but its exit status was not captured. The other two were redirected to a file with `exit=` echoed.)

### E3 Reading the two tests

- Heal test (`retention_owner_test.go:180-235`): the log is stale and a link sits at the state path, so the prune reaches `healStateEntry`. The stand-in `ownerCheck` renames a fresh, still-fresh-stamped file over the inspected link inside the heal and returns true. Correct wiring leaves that file in place (same identity, same bytes) and the pruner adopts its fresh stamp and skips the prune (log unchanged). A `freshInfo == nil` fatal guards against vacuity. `t.Parallel()` is safe: the swap is a field on this test's own `Retention`, no package-level state is touched, and no stderr is written because the check returns true. The two-step kill reason holds on a file system that reuses inode numbers too: the test also compares the fresh file's bytes and the log, which a removal-and-recreate changes.
- Tail test (`retention_tail_test.go:97-135`): a raw-byte comparison of the replacement log against `rest + late`, with a fatal precondition that the late bytes already end in a newline. `t.Parallel()` is safe (own `t.TempDir`, no globals).

### E4 Mutation, current `retention.go` copied to the session scratchpad outside the tree, edited there, applied with `go test -count=1 -v -overlay <json> ./internal/harness/` over the whole package; the tree never held a mutant

| Mutant | Change | Result | Deciding output |
|---|---|---|---|
| L4 | `healStateEntry` calls `os.Remove(statePath)` | killed, only the heal test fails | `retention_owner_test.go:227: the fresh state file was removed or replaced by the heal: err=<nil>`; `--- FAIL: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.03s)`; `FAIL github.com/modu-ai/moai-adk/internal/harness 1.521s` |
| U7 | tail terminator appended unconditionally | killed, only the tail test fails | `retention_tail_test.go:133: pruned log bytes = "...late-event...}\n\n", want "...late-event...}\n" (the kept line plus the late bytes, no extra blank line)`; `--- FAIL: TestPruneTailAlreadyTerminatedGetsNoExtraNewline (0.32s)` |
| N1 | identity comparison reduced to `ModTime` equality only (violates D4.c) | SURVIVED | `ok  	github.com/modu-ai/moai-adk/internal/harness	2.888s` |
| N2 | newline written when the tail is empty (violates REQ-HRH-006, -007 spirit) | SURVIVED | `ok  	github.com/modu-ai/moai-adk/internal/harness	1.976s` |
| N3 | heal re-inspects with a fresh `Lstat` after the owner check and passes that to the helper (violates "the entry the pruner inspected") | killed, only the heal test fails | `retention_owner_test.go:227: ...removed or replaced by the heal`; `--- FAIL: TestPruneHealKeepsAFreshStateFileSwappedInDuringTheHeal (0.02s)` |
| N4 | the "file" arm of the heal calls `os.Remove` directly, the symbolic-link arm keeps the helper | SURVIVED | `ok  	github.com/modu-ai/moai-adk/internal/harness	2.792s` |

L4 and U7 each fail exactly one test, the one aimed at them, and for the stated reason; the older helper-level test `TestHealDoesNotRemoveAFreshStateFile` does not fail under L4, which is the gap F2 named. The scratch copies were built from the committed `retention.go` of this tree (`030f09cbb`, byte-identical to `5bb35abe8`). The lane's own scratch files in the shared scratchpad (`retention_L4.go`, `retention_U7.go`) were not used.

### E5 Residual-risk claims, independent reproduction

```
scratch test via -overlay (a new file mapped into the package; not in the tree), one swap per trial against the production helper, 20000 trials:
zz_scratch_probe_test.go:51: PROBE trials=20000 swapped_in_file_removed=235 helper_first=19742 swap_first=23
--- PASS: TestScratchProbeRemovalWindow (27.11s)
ok  	github.com/modu-ai/moai-adk/internal/harness	27.894s
```

The check-then-remove window exists and the helper removes the swapped-in file (235 of 20000 here, against 5 of 5235 in the full audit's probe; the designs differ, see D3).

```
macOS ln on a symbolic link (shell `ln`, which calls link(2); Go's os.Link was not exercised):
ls -li ...:
3474017454 -rw-r--r--@   2 goos  wheel     5 Oct  3 10:30 hard
3474017455 lrwxr-xr-x@   1 goos  wheel    10 Oct  3 10:30 sym -> target.txt
3474017454 -rw-r--r--@   2 goos  wheel     5 Oct  3 10:30 target.txt
```

`ln sym hard` produced a link to the target file, not to the symbolic link (same inode as `target.txt`, link count 2). That supports the record's point that a put-back with `link` behaves differently on a symbolic-link entry across platforms (darwin follows; Linux does not, which this audit did not observe).

Reading against the code and the SPEC: `removeStateEntryIfUnchanged` (`retention.go:264-282`) is `Lstat`, identity/mode/mtime compare, then `os.Remove`, exactly as the record describes. The three-party failure of the rename-aside repair (a third process creating a file while the path is empty, the put-back then failing as "already exists") follows from the code and is reasoning, not a measurement; the record labels the whole repair analysis "an analysis from reading the code, not a measurement of the alternative". REQ-HRH-005 (`spec.md:161`) reads "remove that entry only if it is still the entry the pruner inspected, and otherwise shall re-inspect without removing it", which a rename-aside moves the entry before verifying; the wording conflict is real as stated, though "remove" versus "move" leaves it arguable.

CHANGELOG sentence (the only changed line, `git diff 5bb35abe8 HEAD -- CHANGELOG.md`): adds "an audit probe that raced the window on purpose removed the fresh file in 5 of 5235 trials, recorded in `.moai/reports/t1432/residual-risk-removal-window.md`". It matches the record and the full audit's E6 figure, keeps "narrows but does not close", and contradicts nothing in `spec.md` D4.c or §F (the SPEC still says "not measured"; see D3).

### E6 Regression checks

```
golangci-lint --version -> golangci-lint has version v2.1.6 built with go1.26.8 ...   (CI pin v2.1.6, version string equal; the build itself was not byte-compared)
golangci-lint run ./internal/harness/ -> 0 issues.   lint exit=0
go vet ./internal/harness/ ./internal/lockfile/   -> vet exit=0
gofmt -l internal/harness/                        -> (empty) gofmt exit=0
GOOS=windows go build ./internal/harness/ ./internal/lockfile/ -> win build exit=0
GOOS=windows go vet   ./internal/harness/ ./internal/lockfile/ -> win vet exit=0
moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001  -> ✓ No findings — all SPEC documents are valid ; exit=0
moai version -> moai-adk v3.2.0-rc.26 ; v3.2.0-rc.26   archive/t1401-293-g45600e4ee   built 2026-10-02T14:42:32Z
git merge-base --is-ancestor 45600e4ee HEAD -> exit 0 ; reverse (HEAD an ancestor of it) -> exit 1 ; git rev-list --count 45600e4ee..HEAD -> 146
git diff --stat 45600e4ee HEAD -- internal/spec | wc -l -> 0
git log --format='%h %s' -- .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/spec.md ->
5bb35abe8 chore(SPEC-HARNESS-RETENTION-HARDEN-001): sync-phase artifacts, 3-phase close [t1432]
1b5c3c057 chore(SPEC-HARNESS-RETENTION-HARDEN-001): record the draft to in-progress transition [t1432]
c2cc32080 docs(...): revise to 0.3.0 for plan-audit iteration 3 [t1432]
2d1918534 docs(...): revise to 0.2.0 for plan-audit iteration 2 [t1432]
69ddcc73a feat(...): plan-phase artifacts (tier M, 3 artifacts) [t1432]
git log --format='%h %s' 5bb35abe8..HEAD -> 030f09cbb, 9d9bc1a0b, 5e2d58a5b (no spec.md change in the delta)
```

## Baseline-attribution

Every row above was measured in this session on the tree `030f09cbb` (E1). Tool provenance (`verification-claim-integrity.md` §2.2): `go test`, `go vet`, `gofmt` and the overlay builds are the Go toolchain compiling the tree's own files, so the build-lag concern does not apply. `moai spec lint` was judged by the installed `v3.2.0-rc.26` (`45600e4ee`), a strict ancestor of HEAD (146 commits behind); `internal/spec` is byte-identical between the two (empty `git diff --stat`), the rest of the binary was not compared. `golangci-lint` v2.1.6 equals the CI pin by version string only. No figure in this report is carried over from `sync-audit.md`, `run-evidence-f2-f3.md` or the residual-risk record; where those are quoted, the quote is marked.

## F1 representation

The convergence FAIL (codex, required backend, P1) from the full audit is now represented as: a SPEC-disclosed residual (D4.c option A), reproduced at the helper level by the full audit and again here, an attempted-repair analysis recorded with its reasons, and a leader ruling to record it if the repair is impossible within the SPEC wording (relayed in the dispatch, not an operator answer; the record itself says "relayed by the leader session, not an operator answer"). The record is complete on the mechanism, the measurement conditions, the codex 3-of-3 attribution, the platform limit and the follow-up path (a new card amending D4.c toward option C). It is not complete on the consequence (D1). The `audit_multi` fan-out was not re-run here, so its recorded `fail` stands unchanged; this audit does not claim it cleared. The leader's decision is not re-litigated; the repair analysis was read for soundness only, and its main point (the rename-aside repair needs a third party to fail and does not close the window) follows from the code.

## Gaps (explicitly not observed)

1. Linux and Windows runtime were not observed. Every run was darwin arm64, uid 501. The Windows twin is build-and-vet only, and both new tests carry `//go:build !windows`.
2. Root and foreign-owned execution: the real foreign-owned refusal was not reached (a non-root user cannot create a foreign-owned file); D2 is therefore unobserved.
3. The new tests were not run under load. The 300 ms settle in `blockedPruner` is the timing dependency D7 names; background load was not spawned.
4. Production frequency of the F1 window: both probes (5 of 5235 in the full audit, 235 of 20000 here) race it on purpose; no production rate exists, and no burst of real hook processes was run.
5. Go's `os.Link` was not exercised on a symbolic link; only the shell `ln` was, on darwin.
6. Whether concurrent pruners lose or duplicate archive or log bytes (D1) was reasoned from the code and the SPEC text, not measured.
7. `run-evidence-f2-f3.md`: every claim was independently re-observed (both passes, both kills with the same failure text and line numbers `:227` and `:133`, race 15, whole package, static checks) except two it cites as its own context: the lane's slot lease and its scratch mutants, which were not available to this audit (equivalent mutants were rebuilt here). Its Gaps and Residual-risk sections (Linux inode reuse; a refactor moving the owner check; the 300 ms settle) are consistent with what was observed; the inode-reuse worry is partly covered by the content assertions (E3).
8. The `audit_multi` fan-out was not re-run (D9).
9. Refused commands: none. Two Bash calls ran as a `;` chain of literal commands and were not refused. The first test run was piped through `tail`, a cosmetic choice noted in E2.

## Residual-risk

- D1, D2 and D3 mean the residual-risk record, as written, can be read as a smaller risk than the SPEC itself states. Correct before the record is cited as the basis for accepting the residual.
- Three narrower siblings of the closed findings survive (N1, N2, N4); none is reachable by a requirement-violating change that the suite would catch.
- The heal check-then-remove window is unchanged by this delta; the new test pins the conditional removal, not an atomic one.
- The tail test can pass without exercising the tail path on a machine slow enough that the pruner reads after the append (D7).
- Single machine, single file system (APFS).

## Iteration history

Iteration 1: `sync-audit.md`, PASS-WITH-DEBT 88 at `5bb35abe8`. Iteration 2 (this report): delta audit at `030f09cbb`, PASS-WITH-DEBT, no blocking finding, scoped to the F2 and F3 tests and the F1 record.
