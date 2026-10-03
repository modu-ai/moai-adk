auditor-model: claude-sonnet-5-5[1m]

verdict: PASS-WITH-DEBT
audited_sha: 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8

# Sync Audit Report: SPEC-HARNESS-RETENTION-HARDEN-001 (card t1432, Tier M)

Overall Verdict: PASS-WITH-DEBT. No must-pass dimension fails and no finding is classified blocking. One finding (F1) is a SPEC-disclosed concurrency residual that the required codex backend rated P1 and that this audit reproduced; the leader must decide whether to accept it (see "Decision owed to the leader").
Overall Score: 88/100 (flat weighted percentage, default profile, harmonic aggregation 87.9; `evaluator_mode: hierarchical` is not set in `.moai/config/sections/harness.yaml`, and `spec.md` carries no `evaluator_profile`, so the built-in default profile applies).
Cross-model convergence result (`mcp__moai__audit_multi`, quoted in full under Claim 9): overall_verdict `fail`, required-backend FAIL: codex. This is reported, not overridden silently; the reasoning for the audit verdict differing from it is in F1.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 92/100 | PASS | All 14 acceptance criteria reproduced green (`go test -count=1 -v` on the 16 named tests, 16 top-level `--- PASS`, 0 `--- FAIL`); RED independently reproduced against base code for AC-HRH-001/002/007/008(b,c); package `ok` plain, under `-race`, and 375 `--- PASS` lines of which 277 top-level, 0 FAIL |
| Security (25%) | 80/100 | PASS (no Critical/High by this audit's rating) | Ownership check reads the link's own owner with `Lstat`; deletion only of a current-user entry; foreign/undeterminable entries left byte-identical (AC-HRH-004 plus mutants L1, L2, L3, L11 killed). Residual: F1 (check-then-unlink window, reproduced), F4 (O_EXCL create and post-open identity check not pinned) |
| Craft (20%) | 88/100 | PASS | Package coverage 87.1 percent (>= 85); `golangci-lint run ./internal/harness/` `0 issues.` (v2.1.6, the CI-pinned version string); gofmt and vet clean. New-branch gaps: `openStateFile` 72.4 percent, `removeStateEntryIfUnchanged` 66.7 percent; 7 of 25 mutants survived (F2-F5) |
| Consistency (15%) | 92/100 | PASS | Build-tag twins present and `GOOS=windows` build+vet green; comments English with no card id or design history (`grep` of the sentinels, below); `@MX:WARN` tags carry `@MX:REASON`; only nine files changed against base, as scoped. One stale claim in a commit message (F6) |

Must-pass firewall (Functionality, Security): both meet their thresholds independently (default profile: all AC PASS; no Critical/High finding), so the firewall does not force FAIL.

## Findings (structured defect list)

Classification: `blocking` is reserved for correctness or a stated SPEC requirement; none qualifies. Confidence is stated per finding.

- F1 [medium] [optional: accept-as-debt, DECISION OWED TO THE LEADER] `internal/harness/retention.go:265-275` (`removeStateEntryIfUnchanged`) - the identity re-check (`os.Lstat`, `os.SameFile`, mode, mtime) and the unlink (`os.Remove`) are two separate system calls. A concurrent healer that renames a fresh state file over the entry between them has its file removed, which re-opens the two-pruner condition REQ-HRH-005 exists to prevent. Confidence: high (reproduced). SPEC status: disclosed as "the guarantee is narrowed to the gap between the last re-inspection and the removal, which is not closed and **not measured**" (`spec.md` §B D4.c option A, §F "Heal burst"), and the CHANGELOG says "narrows but does not close"; D4.c option C (a separate heal lock) was rejected in the plan as adding an artifact. This audit supplies the measurement the SPEC lacked: in a single-swap probe against the production helper (scratch test via `go test -overlay`, outside the tree), the helper removed the swapped-in fresh entry in 5 of 5235 trials (about 0.1 percent per trial with the swapper deliberately racing the window; trigger in production needs a faulty owned entry plus a simultaneous second hook process). The required codex backend independently rated it P1 and reports a 3 of 3 reproduction through its own overlay (that overlay lived outside this tree and was not available to this audit; the existence of the window was reproduced here by a different probe). Not exploitable by a foreign user to delete data: a foreign-owned swapped-in entry is refused by directory permissions or sticky bit, and the heal never starts on a foreign-owned entry. Required fix (if the leader does not accept the residual): replace the check-then-unlink with an atomic take: `os.Rename(path, <unique temp name in the same directory>)`, `Lstat` the temp name and compare it to the inspected entry, unlink the temp name when it matches, and when it does not match return it with `os.Link(temp, path)` (which fails if the path exists, so it never clobbers) followed by removal of the temp name; or adopt D4.c option C. Either changes `retention.go` only; a delta re-audit scoped to that helper plus a deterministic test is sufficient. If accepted, record the measured 0.1 percent figure and its probe conditions in the residual-risk record.
- F2 [medium] [optional: accept-as-debt] `internal/harness/retention.go:254` (`healStateEntry`) - REQ-HRH-005 wiring is unpinned. Mutant L4 (the heal calls `os.Remove(statePath)` directly instead of `removeStateEntryIfUnchanged`) passes the entire package (`ok`, re-run alone to confirm). AC-HRH-006 exercises the helper in isolation, as `spec.md` §B D4.c specifies ("a deterministic helper-level test"), so the wiring from the heal into the helper is checked by nothing. Confidence: high. Required fix (cheap, uses the existing seam): a test that sets `ret.ownerCheck` to a function that renames a fresh state file over the inspected link and returns true; with the correct wiring the fresh file keeps its identity, with mutant L4 it is removed.
- F3 [low] [optional: accept-as-debt] `internal/harness/retention.go:528` (`appendLogTail`) - mutant U7 (the terminator is added even when the tail already ends in a newline) survives, because `readLogLines` drops blank lines. REQ-HRH-006 says carried bytes stay verbatim and REQ-HRH-007 allows only the missing terminator to be added, so an extra blank line per prune is a modification the suite cannot see (the next rewrite drops blank lines, so the effect is bounded and cosmetic). Confidence: high. Fix: in `TestPruneCarriesLateEvents`, compare the raw file bytes against `<fresh line>\n<late line>\n` instead of the blank-dropping helper.
- F4 [low] [optional: accept-as-debt] `internal/harness/retention.go:203`, `:220-224`, `:272` - defence-in-depth branches are unpinned: mutant U3 (create without `O_EXCL`), U5 (post-open identity check dropped) and U8 (mode and mtime dropped from the identity comparison) all survive; the coverage profile shows these blocks unexecuted: 207-208, 211-212, 223-224, 226-231, 240, 243, 266-270, 276-278. `spec.md` §F states the time-of-check gap "cannot be forced without a seam and is not pinned by a test"; U8 contradicts the D4.c wording ("same file identity, same type, same permission mode, same modification time") without any test noticing. Confidence: high that they are unpinned, medium that they matter. Fix: none required; optionally pin U8 with an entry whose mode changes between inspection and removal.
- F5 [low] [optional: accept-as-debt] `internal/harness/retention.go:229` - mutant U6 (heal on every open error, not only permission errors) survives; this is the limit already recorded in `acceptance.md` AC-HRH-002. Confidence: high. No action.
- F6 [low] [optional: accept-as-debt] commit `5bb35abe8` message - states "the draft to in-progress step was never committed separately" and that "the in-progress and implemented steps ride this one commit", while commit `1b5c3c057` is exactly the separate draft to in-progress commit and `spec.md` history shows both (`git log` below). The text is a leftover of the undone first sync commit (`390c9b505`, per `progress.md` §E.4). The correct account is in `progress.md` §E.4. Fix: reword the message at merge if the leader amends the unpushed branch tip; otherwise leave it, the record is correct elsewhere.
- F7 [info] [accept] Windows runtime is not observed: the twin `retention_owner_windows.go` (`return false`) is verified by `GOOS=windows go build` and `go vet` only. A mutant that makes the Windows twin return true cannot be exercised here; REQ-HRH-004's "every entry is not owned on Windows" is reviewer-read.
- F8 [info] [accept] Warning volume: five attempts on one `Retention` instance with a foreign-owned entry produced five lines (probe), one per attempt, unbounded while the condition persists; disclosed in `spec.md` §F ("Warning volume"). Whether hook stderr reaches a user was not measured.
- F9 [info] [accept] The pre-existing FIFO-at-the-state-path hang in the lock-free pre-check is not repaired and no test uses a FIFO there, as the SPEC records; unchanged by this audit.
- F10 [info] Convergence tool outputs: GLM returned `z.ai returned HTTP 401` (inconclusive, advisory gate); the "claude" backend entry in the convergence result is this audit's own in-session anchor and is not an independent second Claude opinion. No `audit_receipt` was issued (the tree does not set `workflow.audit.gates.codex` explicitly).

### Carried plan-audit debt N1-N6, O4, O8, O9

| Item | Status at HEAD | Became a defect? |
|---|---|---|
| N1 DoD-1 wording, N2 stale "untracked" note (G-7, `plan.md` B5), N3 CI Windows statement | open, documentation-only; SPEC body deliberately not edited (plan-artifact hash) | no |
| N4 AC-HRH-005 (e) skip granularity | not skipped on darwin (`/var` qualified); Linux unobserved | no |
| N5 no end-to-end case for a user-owned link to a root-owned target | still no pruner-level case; the owner-check level is pinned (mutant L1 killed by case (d)) | no |
| N6 mutant copies base-bound | discharged: the run regenerated its mutants, and this audit generated 25 more from the current files | no |
| O4 inode-reuse false red in AC-HRH-002 | open, inferred only; APFS did not reuse in any run here | no |
| O8 late-event substring comparison | narrower than recorded: the committed test compares the late line by equality (`lines[1] != marshalEvent(...)`); the real gap is the byte-level one in F3 | related (F3) |
| O9 phrase naming the missing archive step not pinned | open; the observed message under mutant L9 reads `the pruner never reached its archive step: nothing opened the archive FIFO ... for writing within 10s`, so it names the step | no |

## Claim

1. The tree audited is `5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8`, branch `WT-harness-retention-debt`, toplevel `.moai/worktrees/t1432`, clean before and after the audit.
2. AC-HRH-001..014 are met on this tree, with the qualifications under Gaps.
3. The change stays inside the seven operator-scoped items; `internal/lockfile` and `observer.go` are byte-identical to base; the four pre-lock functions have no commit since the red baseline.
4. Baseline-first ordering holds on the commit graph.
5. 18 of 25 mutants built by this audit are killed; the 7 survivors are listed in F2-F6 (U2 and U11 are equivalent in observable behaviour).
6. The CHANGELOG entry and `progress.md` §E.2-§E.4 are truthful against the tree, with the F6 exception.
7. The Windows-runtime and Linux disclosures are present and true.
8. The disclosure sentences (REQ-HRH-009..011) say what the requirements ask.
9. Cross-model convergence: codex fail (F1), claude anchor pass, glm inconclusive.

## Evidence

All commands ran in this session against the tree above. Kanban variables were scrubbed inside the same invocation as every test run (`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ...`); that prefix is omitted below. Heavy runs held the lease `moai slot acquire --resource go-test-internal-harness --max-duration 20m` and released it afterwards (`slot go-test-internal-harness released`, then `free`).

### E1 Tree identity

```
git rev-parse --show-toplevel   -> /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1432
git branch --show-current       -> WT-harness-retention-debt
git rev-parse --short HEAD      -> 5bb35abe8
git status --short              -> (empty), exit 0 ; empty again after all runs
```

### E2 The named tests (swept count: 16 names, 16 top-level `--- PASS`)

```
go test -count=1 -v -run '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced|TestPruneStateUnreplaceableInReadOnlyDirSkips|TestPruneStateForeignOwnedLeftUntouchedAndWarns|TestOwnerCheckDefault|TestHealDoesNotRemoveAFreshStateFile|TestPruneCarriesLateEvents|TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval|TestPruneStamp_FreshStampNeedsNoLock|TestPruneStamp_StampExistsBeforeTheWork|TestPruneStampShorterOverLongerIsExact|TestPruneStampWriteFailureSkipsPrune|TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds)$' ./internal/harness/
--- PASS: TestPruneStateForeignOwnedLeftUntouchedAndWarns (0.01s)   [symlink, unwritable-file subtests PASS]
--- PASS: TestOwnerCheckDefault (0.00s)
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.00s)
--- PASS: TestPruneNothingStaleLeavesLogUntouched (0.01s)
--- PASS: TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds (0.01s)
--- PASS: TestHealDoesNotRemoveAFreshStateFile (0.01s)
--- PASS: TestPruneStateSymlinkReplacedTargetUntouched (0.01s)
--- PASS: TestPruneStampWriteFailureSkipsPrune (0.01s)
--- PASS: TestPruneStamp_StampExistsBeforeTheWork (0.02s)
--- PASS: TestPruneStamp_FreshStampNeedsNoLock (0.02s)
--- PASS: TestPruneStampShorterOverLongerIsExact (0.02s)
--- PASS: TestPruneKeepsUnparsedLinesVerbatim (0.00s)   [5 subtests PASS]
--- PASS: TestPruneStateUnwritableFileReplaced (0.02s)
--- PASS: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.02s)
--- PASS: TestPruneTailPartialLineCarriedAndTerminated (0.32s)
--- PASS: TestPruneCarriesLateEvents (0.32s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.935s
```

No `--- SKIP` belongs to any of them (darwin, uid 501), including `TestOwnerCheckDefault` step (e), which found `/var`.

### E3 Whole package

```
go test -count=1 ./internal/harness/            -> ok  	github.com/modu-ai/moai-adk/internal/harness	1.520s
go test -race -count=1 ./internal/harness/      -> ok  	github.com/modu-ai/moai-adk/internal/harness	7.480s
go test -count=1 -v ./internal/harness/ > scratch ; exit=0
  grep -c -e '--- PASS'  -> 375 (indented subtests included)
  grep -c -e '^--- PASS' -> 277 (top level; equals the figure in verdict.md 2.4)
  grep -c -e '--- FAIL'  -> 0
  --- SKIP: TestPruneHelperProcess (0.00s)   (pre-existing helper process, as verdict.md records)
go test -race -count=15 -run '<8 new timing-sensitive tests>' ./internal/harness/
                                                -> ok  	github.com/modu-ai/moai-adk/internal/harness	7.385s
go test -count=1 -coverprofile=<scratch> ./internal/harness/
                                                -> ok  	github.com/modu-ai/moai-adk/internal/harness	1.887s	coverage: 87.1% of statements
```

Per-function (from `go tool cover -func`): `healStateEntry 100.0%`, `pruneLocked 100.0%`, `scanTerminatedLines 100.0%`, `writeStamp 100.0%`, `openStateFile 72.4%`, `removeStateEntryIfUnchanged 66.7%`, `appendLogTail 81.2%`, `entryOwnedByCurrentUser 71.4%`, `overwriteWithEvents 50.0%` (its uncovered lines are error returns).

### E4 RED reproduced against base production code

Overlay replacing `retention.go` with `git show 1e2151a38:internal/harness/retention.go`, with `retention_owner_unix.go`, `retention_owner_windows.go`, `retention_owner_test.go` and `retention_stampwrite_test.go` mapped to empty (they need the new symbols); HEAD test files otherwise unchanged:

```
--- FAIL: TestPruneStateSymlinkReplacedTargetUntouched   victim changed: content="2026-10-02T00:00:00Z" ; state path is not a regular file
--- FAIL: TestPruneStateUnwritableFileReplaced           first prune returned retention: prune state open failed: ... permission denied, want nil
--- FAIL: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval   stale-final count after first prune = 0, want 1
--- FAIL: TestPruneTailPartialLineCarriedAndTerminated   partial fragment occurs 0 times, want 1
--- FAIL: TestPruneCarriesLateEvents                     late-event count = 0, want 1
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips  (green at base: the AC-HRH-003 regression guard, including the strengthened errors.Is assertion)
--- PASS: TestPruneNothingStaleLeavesLogUntouched, TestPruneKeepsUnparsedLinesVerbatim, TestPruneStampShorterOverLongerIsExact
```

The five RB criteria are red at base for the stated reasons; the pins are green at base.

### E5 Mutation (25 scratch mutants, generated from the CURRENT files by one script; none was ever in the tree; each run is `go test -count=1 -overlay <json> ./internal/harness/` over the whole package; every output read for `--- FAIL` and checked for build errors)

Listed by the task:

| Mutant | Result | Deciding failing output |
|---|---|---|
| L1 owner read follows the link (`os.Stat`) | killed | `retention_owner_test.go:151: a user-owned link to / must count as owned: the link's own owner decides, not its target's` |
| L2 every symlink treated as owned | killed | `retention_owner_test.go:169: /var is a symbolic link owned by another user and must not count as owned` |
| L3 removal before the ownership check | killed | `state-path entry changed: ... no such file or directory`; `link text changed` |
| L4 heal calls `os.Remove` directly | SURVIVED (`ok ... 1.540s`, re-run alone) | none (F2) |
| L5 tail copy dropped | killed | `late-event count = 0, want 1`; `partial fragment occurs 0 times, want 1` |
| L6 tail copy without terminator | killed | `replacement log does not end with a newline` |
| L7 `Truncate(0)` dropped | killed | `state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"` |
| L8 stamp-write error ignored | killed | `pruneLocked returned nil, want the stamp-write error` |
| L9 prune replaced by a no-op under the M4 test | killed in 12.7 s wall (under the 15 s bound), no panic | `retention_killed_test.go:119: the pruner never reached its archive step: nothing opened the archive FIFO ... for writing within 10s` |
| L10 unconditional removal inside the helper | killed | `removal of a changed entry: removed=true err=<nil>, want false and nil` |
| L11 owner check always true | killed | `/ must not count as owned for a non-root user` |

Unlisted (invented by this audit):

| Mutant | Violates | Result |
|---|---|---|
| U1 tail read before the archive step instead of immediately before the rename | REQ-HRH-006 | killed (`late-event count = 0, want 1`) |
| U2 classified prefix not bounded by the opened file's size | REQ-HRH-006 boundary | survived; equivalent in observable behaviour (bytes consumed define `classifiedEnd`, so no byte is lost or duplicated) |
| U3 replacement created without `O_EXCL` | D4.c / REQ-HRH-005 | survived (F4; time-of-check, unforceable without a seam) |
| U4 warning written to stdout | REQ-HRH-004 | killed (`want exactly one warning line ... got ""`) |
| U5 post-open identity check dropped | D4.c | survived (F4) |
| U6 heal on every open error | REQ-HRH-002 scope | survived (F5; recorded limit) |
| U7 tail terminator added unconditionally | REQ-HRH-006/007 | survived (F3) |
| U8 mode and mtime dropped from the identity comparison | D4.c | survived (F4) |
| U9 owner check applied to links only | REQ-HRH-003/004 | killed (`unwritable-file` subtest: `want an error naming ...`) |
| U10 tail seek off by one | REQ-HRH-006 | killed (`parse log line: invalid character ':'`; `partial fragment occurs 0 times`) |
| U11 real uid instead of effective uid | none here | survived; equivalent (uid equals euid in this environment) |
| U12 `maxStateInspections` 3 to 1 | REQ-HRH-001 | killed (`changed on every inspection; prune skipped, want nil`) |
| U13 tail offset beyond the end of the file | REQ-HRH-006 | killed (first build failed on an unused variable, which is not a kill; regenerated and re-run: `late-event count = 0`) |
| U14 warning line removed | REQ-HRH-004 | killed |

Counts: 11 listed, 10 killed, 1 survived (L4); 14 unlisted, 8 killed, 6 survived (U2 and U11 equivalent; U3, U5, U6, U7, U8 are F2-F5 material). Survivors in total 7 of 25.

### E6 Adversarial runtime probes (scratch test files injected with `-overlay`, outside the tree)

```
PROBE dangling: err=<nil> target_created=false state_regular=true
PROBE link-to-dir: err=<nil> dir_entries=0 state_regular=true
PROBE self-loop: err=<nil> state_regular=true
PROBE created state file mode=-rw-r--r-- (umask-dependent; base created with 0644)
PROBE mode-0000 file: err=<nil> mode_after=-rw-r--r--
PROBE crlf: kept=1 stale=1 classifiedEnd=281 want=281 err=<nil>
PROBE crlf after prune: "...fresh event...\n{\"partial\"\n"
PROBE warn volume for 5 attempts on one Retention instance: 5 lines
PROBE remove-window: trials=5235 helper_removed_the_swapped_in_fresh_entry=5 helper_first=5230 swap_first=0
```

A first attempt at the race probe (hard-link spinner with link-count detection) was unsound and was discarded; the second design (one swap per trial; path absent after a removal means the unlink hit the swapped-in file) produced the figure above. `swap_first=0` shows the swapper never won the start, so the figure is a lower-bound style sample, not a rate for production.

### E7 Boundary, ordering, platform, static checks

```
git diff --quiet 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go   -> exit=0
git diff 1e2151a38 HEAD --stat -- internal cmd   -> 9 files (retention.go, killed test, 6 new files incl. 2 owner files), 980 insertions, 26 deletions
git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go   -> c1cc3fe67 f52dd1b1c b3a469eab 9289a92b6
git merge-base --is-ancestor ac40cf3bf <each of the four>   -> exit 0 x4 ; reversed control (c1cc3fe67 ac40cf3bf) -> exit=1
git log --format=%h -s -L <PruneStaleEntries> -L <readStamp> -L <readStampFile> -L <stampIsFresh> ac40cf3bf..HEAD   -> (empty) exit=0
  positive control, same options, range fe211e9c9~1..fe211e9c9   -> fe211e9c9 exit=0
git show ac40cf3bf --stat   -> red-baseline.md + retention_stampbytes_test.go + retention_statepath_test.go + retention_tail_test.go (701 insertions)
git ls-files .moai/reports/t1432   -> red-baseline.md, verdict.md, run-evidence-m1-m3.md, decision-records.md, plan-audit*.md and the red-now-drafts files (tracked)
GOOS=windows go build ./internal/harness/ ./internal/lockfile/   -> exit 0
GOOS=windows go vet   ./internal/harness/ ./internal/lockfile/   -> exit 0
gofmt -l internal/harness/   -> (empty) exit 0
go vet ./internal/harness/ ./internal/lockfile/   -> exit 0
golangci-lint --version -> golangci-lint has version v2.1.6 built with go1.26.8
golangci-lint run ./internal/harness/   -> 0 issues.
```

### E8 SPEC and document consistency

```
moai version -> moai-adk v3.2.0-rc.26 ; v3.2.0-rc.26   archive/t1401-293-g45600e4ee   built 2026-10-02T14:42:32Z
git merge-base --is-ancestor 45600e4ee HEAD -> exit 0 (strict ancestor; HEAD is not an ancestor of it: exit 1)
git rev-list --count 45600e4ee..HEAD -> 143
git diff --stat 45600e4ee HEAD -- internal/spec -> (empty)
moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001 -> ✓ No findings — all SPEC documents are valid ; exit 0
moai spec audit -> Total SPECs: 1028 / Grandfathered: 266 / Modern-era clean: 759 / Drift findings: 737 ;
                   only line naming this SPEC: [INFO] SPEC-HARNESS-RETENTION-HARDEN-001 (V3R6) — EraAutoDetected
git log --format='%h %s' -- .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/spec.md ->
  5bb35abe8 chore(...): sync-phase artifacts, 3-phase close [t1432]      (in-progress -> completed)
  1b5c3c057 chore(...): record the draft to in-progress transition [t1432]
  c2cc32080, 2d1918534, 69ddcc73a (plan phase)
grep -o -E 'AC-HRH-[0-9]+' acceptance.md | sort -u | wc -l -> 14 ; REQ-HRH in spec.md -> 15 ; CHANGELOG lines naming the SPEC -> 1
```

Disclosure sentinels in `retention.go` (single `grep -c -F` each): `residual window` 2, `no cross-process exclusion` 1, `burst of hook processes` 1, `F5: not reproduced, not measured` 1, `lock waiters block with no timeout` 1, `5 s hook timeout` 1, `appended before the wait` 1, `F6: not reproduced, not measured` 1, the old sentence `events other hooks append in that window are lost` 0, card-id pattern `t1[0-9]{3}` 0. A comment scan for `operator|audit|REQ-|SPEC-|AC-|plan|relayed|card|decision` finds only the three pre-existing `REQ-HL-011` lines. Read against REQ-HRH-010 and -011: the Windows sentence names the in-process mutex, no cross-process exclusion, the burst and the F5 label; the lock-wait sentence names no timeout, no try-lock, the 5 s timeout, `async: true`, the append-before-wait order and the F6 label. `observer.go:90-93` shows the prune call after the write; the template shows `"timeout": 5` with `"async": true` on the observe hooks.

### E9 Cross-model convergence (`mcp__moai__audit_multi`, `project_root` set to this worktree, `target: baseBranch`, `card_id: t1432`)

```
backends: claude (required) pass [source: in_session_anchor] ; codex (required) fail ; glm (advisory) inconclusive "z.ai returned HTTP 401"
overall_verdict: fail ; residual_risk_note: required-backend FAIL: codex; cross-model disagreement: pass=[claude(required)] fail=[codex(required)]
build_lag: installed binary built from 802a72235, an ancestor of HEAD 5bb35abe8 (the running MCP server lags the tree)
codex finding (P1, retention.go:275): TOCTOU after the ownership and identity re-check; reported repro 3/3 through an overlay under /tmp/t1432-review-...
```

## Baseline-attribution

Every row in Evidence was measured in this session on the tree and SHA stated in E1. Tool provenance (`verification-claim-integrity.md` §2.2):

- `go test`, `go vet`, `gofmt`, `go tool cover`: the Go toolchain of this machine (darwin, uid 501) building the tree's own packages; the tests judged the tree directly, so the build-lag concern does not apply to them.
- `moai spec lint` and `moai spec audit`: judged by the installed build `v3.2.0-rc.26` (`45600e4ee`), a strict ancestor of the tree HEAD `5bb35abe8` (143 commits behind); `git diff --stat 45600e4ee HEAD -- internal/spec` printed nothing, so the lint package is unchanged between the two, while the rest of the binary was not compared.
- `golangci-lint`: version string v2.1.6 equals the CI pin; the build itself was not byte-compared with CI's.
- `audit_multi`: the MCP server builds from `802a72235` (also an ancestor of HEAD); its codex and claude backends read the diff of this worktree because `project_root` was passed.
- No figure is carried over from `verdict.md` or `progress.md`; where this report quotes them, it says so (277 top-level PASS reproduced here).

## Gaps (explicitly not observed)

1. **Windows runtime not observed.** Build and vet only; the Windows owner-check twin and the in-process Windows lock were not executed. Confirmed present and true in `verdict.md` §5 and `progress.md` §E.2/§E.3/§E.4.
2. **Linux unobserved.** Every run was darwin arm64, uid 501. First exercised on Linux in CI: the read-only-handle truncate seam, the FIFO tests, `O_EXCL`, the owner stat, the AC-HRH-005 (e) candidate list. Confirmed disclosed.
3. **Root and foreign-owned execution.** Tests that skip at uid 0 were not run as root; the real foreign-owned refusal was reached only through the injected `ownerCheck` (a non-root user cannot create a foreign-owned file) plus the `/var` link case of `TestOwnerCheckDefault`.
4. **F1 rate in production.** The 5-of-5235 figure comes from a probe that races the window on purpose; no production rate was measured, and a heal burst of N real hook processes was not exercised. Codex's own reproduction was not available to this audit.
5. **`moai spec lint` judging build** lags HEAD by 143 commits (above); only `internal/spec` equality was checked.
6. **GLM backend inconclusive (HTTP 401)**; the cross-model result has two voices, one of which is this audit's own anchor.
7. **Whether hook stderr reaches a user** (warning volume) and whether a concurrent `O_APPEND` write can be seen half-complete: not measured (SPEC-disclosed).
8. **Not run by this audit:** the plan-phase controls E-017/E-018 (untagged versus tagged FIFO test under Windows vet); `moai spec audit` and the lint were run once each on the committed tree, not on a pre-edit tree.
9. **Refused commands:** none. No command was refused by the worktree guard in this audit. Two retries were mine: a `grep --include=*.go` glob failed under zsh and was re-run quoted; the first nlink-based race probe was discarded as unsound (E6). Neither changes a measurement cited above.

## Residual-risk

- F1: a faulty owned state-path entry met by two hook processes at the same instant can lose the fresh state file and yield two pruners in one interval (reproduced at the helper level; production frequency unmeasured; consequence bounded to one interval and to the t1425 double-archive hazard).
- Unpinned wiring and defence-in-depth (F2-F4) mean a later refactor could silently drop the conditional-removal call, the `O_EXCL` create or the post-open identity check with a green suite.
- A foreign-owned or Windows state-path entry keeps retention off with one stderr line per prune attempt and no rate limit; the observer discards the error by design.
- The residual late-event window between the final tail reading and the rename remains, as disclosed.
- The pre-existing FIFO hang (F9) remains.
- This audit's sampling is one machine and one filesystem (APFS); behaviour of `rename`, `link` and the read-only-handle seam on Linux filesystems may differ.

## Decision owed to the leader

1. **F1 / convergence FAIL.** The required codex backend returned FAIL on a residual the SPEC discloses and the plan accepted as D4.c option A (a plan-level default, not an operator-relayed verdict: `decision-index.md` carries no row accepting it). This audit rates it Medium (narrow trigger, bounded consequence, no foreign-user data loss, disclosed in four places) and issues PASS-WITH-DEBT. If the leader or operator does not accept the residual, return it for a delta fix scoped to `removeStateEntryIfUnchanged` plus one deterministic test (F1 fix text); the confirming re-audit then needs only that delta. If accepted, add the measured figure to the residual-risk record.
2. **F2-F4** are optional test-strengthening items; none is routed to a fix unless the leader chooses to.
3. **F6** is a commit-message wording issue; amend at merge or leave.

## Iteration history

Iteration 1 (this report). No earlier sync audit exists for this card. Plan audits: iteration 1 FAIL 0.75, iteration 2 FAIL 0.87, iteration 3 PASS 0.92 (`.moai/reports/t1432/plan-audit*.md`, read for context only after the independent pass).
