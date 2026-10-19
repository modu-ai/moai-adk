auditor-model: claude-sonnet-5-5
verdict: FAIL
audited_sha: 4a625be15622eeb610df317807831d51e4257e36
required_backend: codex fail
required_backend_fail: codex
convergence_overall: fail

# Sync audit round 5 (final audit of the tree after the chunk 13 shrink) - SPEC-UPDATE-OWNERSHIP-REPAIR-001, card t1865

The previous content of this file (round 4, HEAD aba63b1483) is preserved as `.moai/reports/t1865/sync-audit-r4.md`.

SPEC: SPEC-UPDATE-OWNERSHIP-REPAIR-001
Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, branch WT-update-ownership-repair, audited HEAD `4a625be15622eeb610df317807831d51e4257e36`, `git status --short` empty.
Diff base: `48bfe97afa4b78cb1c1ff2c98f1a5680b1441309..HEAD` (77 files, 22836 insertions, 167 deletions).
Judging build: the moai MCP server `build_commit` field of the audit_multi result is `7705b365e2` (the server banner says the same). That commit is a valid commit, not an ancestor of HEAD (`git merge-base --is-ancestor 7705b365e2 HEAD` exit 1; merge-base with HEAD is `48bfe97afa`; 144 commits in HEAD not in the build, 129 in the build not in HEAD). It is therefore not a strict ancestor of the tree's HEAD (verification-claim-integrity.md section 2.2), but it also does not contain this card's code: the backend judged the diff text, not a build of this tree.
Overall Verdict: FAIL

Verdict basis: the required codex backend returned `fail` (`overall_verdict: fail`, `residual_risk_note: required-backend FAIL: codex`). Independently of that gate, I read the code and found the same two P2 defects in the shrink surface (F1 and F2 below). The shrink did close every round 4 finding (N-A root path swap, N-B over-broad RemoveAll, N-C absolute in-project link) as designed, and all tests I ran are green.

## audit_multi result digest (for the orchestrator's convergence check)

- overall_verdict: fail
- gate_unmet: not present in the result (no key returned)
- plan_source: config
- build_commit: 7705b365e2
- participant_count: 1, disagreement_flag: null, fail_open_backends: []
- per_backend_verdicts: backend=codex, gate=required, verdict=fail
- audit_receipt: `rcpt-e012674ec01973db90b4bb52`
- Plan read first: `moai verify audit-plan --project-root <toplevel>` returned `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`; claude and glm gates `off`, codex gate `required`. `audit_multi` was called exactly once, with `project_root` set to the toplevel, `target: baseBranch`, no `gates` argument.
- This verdict is not final until the orchestrator writes these members to `<toplevel>/.moai/state/audit-plan-result.json` and `moai verify audit-plan --result-file` passes; `convergence_check.ok: false` is an unmet gate named by backend.

Raw required-backend text (codex, translated from Korean only in the one-line titles below; the file/line cites are verbatim):

- [P2] a regular file swapped for a FIFO right after the kind check makes update wait forever - internal/cli/migrate_ownership.go:427
- [P2] the symlink check really reaches an outside profile root - internal/cli/migrate_ownership.go:414
- [P3] a file name can poison the root-escape error test - internal/cli/migrate_ownership.go:363

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 80/100 | PASS (on the ACs run) | 101 + 100 + 14 test passes, 0 fail, 0 skip on the selectors below; AC-OWR-09..15 not run here (Gaps) |
| Security (25%) | 55/100 | FAIL (must-pass) | F1 (REQ-OWR-005 stat through a project link), F2 (unbounded block on a swapped FIFO) |
| Craft (20%) | 80/100 | UNVERIFIED | `go vet` exit 0, `gofmt -l` empty; coverage, lint and `-race` not run (Gaps) |
| Consistency (15%) | 85/100 | PASS | closed vocabulary test green, existing seams and naming used |

Must-pass firewall: Security (a built-in must-pass dimension) does not meet its threshold, so the overall verdict is FAIL whatever the other scores are. The required codex gate fails independently.

## Findings

- F1 [P2 / BLOCKING] `internal/cli/migrate_ownership.go:414` - `readUnregistered`, when a path component is a symlink, calls `os.Stat(abs)` only to choose between the reasons `read-failed` and `symlink-escape`. `os.Stat` follows the link, so a project link whose target is a profile root file (for example a link into a profile configuration directory's `skills/`) makes the update stat that profile path. REQ-OWR-005 says the update shall not "open, read, enumerate, stat, walk, supply, or write any profile root". The call is outside the AC-OWR-08 source scan (the scan freezes call sites of profile symbols, and this is an `os` call on a project path), and the differential canary has no project symlink into a profile root, so no test sees it. The stat is gratuitous: the SPEC decision is that every link is kept unread; the only thing the stat changes is a reason label. Codex reproduced it: flipping only the existence of the outside target changed the result from `read-failed` to `symlink-escape`. I did not execute that reproduction (Gaps). Required fix: remove the `os.Stat(abs)` call; for a symlink component that cannot be examined inside the Root, keep the file unread with `symlink-escape` (or, if the dangling-link `read-failed` baseline must stay, decide it from `root.Lstat`/`Root` readlink-style metadata that never leaves the root), then add a differential test whose project link points into a fake profile root and asserts that root's metadata is never touched, and extend AC-OWR-08 with that row. Related residual, not new in the shrink: `unreadableKind(abs)` (`migrate_ownership.go:345`) and `os.ReadFile(abs)` (`:320`) follow links of REGISTERED files by path; progress.md already discloses the registered-file read as debt (N16 family), so I rank it MINOR and mention it so F1's fix is not mistaken for the whole link surface.
- F2 [P2 / SHOULD-FIX] `internal/cli/migrate_ownership.go:427` - `root.Open(...)` runs after the kind check; a regular file swapped for a FIFO between `root.Lstat` (`:402`) and `root.Open` makes the open block until a writer appears, so `moai update` hangs even with `removalEnabled=false`. The code comment on `keepNonRegularFile` (`:73`) states the design intent "kept without ever being opened, because reading such a file can block forever", and the race leaves a window where that does not hold. progress.md caveat 6 discloses it ("a regular file swapped for a FIFO between the kind check and the read" is still uncovered), and the race needs a local writer inside the project during the run, so it is a disclosed residual, but it is reproduced as a hang by the required backend. Required fix: open without blocking (for example `root.OpenFile` with `O_NONBLOCK` on unix, via a build-tagged helper, or open then `f.Stat()` and read only a regular file), re-check the kind on the opened handle, and add a `lstat-to-open` race row that swaps in a FIFO (the `readWindowHook` seam already supports it).
- F3 [P3 / MINOR / optional] `internal/cli/migrate_ownership.go:363` - `escapesRoot` matches the substring `path escapes from parent` in `err.Error()`, and a `PathError` message contains the user's file name, so a missing file named `path escapes from parent.md` is reported `symlink-escape` instead of `read-failed`. Both outcomes keep the file unread, so only a reason label and a per-reason tally are wrong. Required fix if taken: compare the unwrapped cause (`errors.Unwrap`/`*fs.PathError.Err`) text instead of the full string. The stdlib-text dependence itself is real (`errPathEscapes` is unexported in `os/file.go:421` of the toolchain in use) but fails closed: a changed text yields `read-failed`, nothing is read, and the race rows in `TestPathBoundaryCounterexamples` expecting `symlink-escape` would turn red, so it is pinned.
- F4 [MINOR / optional, observation] `pruneManagedEmptyDirs` (REQ-OWR-008) is not gated by `priorManifestFailed`, while REQ-OWR-016 says the run "shall remove no file". It removes only empty directories and never a file, so the requirement text holds; recorded so nobody reads AC-OWR-20 as "no filesystem change at all" when the prior-manifest copy failed.

Classification note: F1 is blocking because REQ-OWR-005 is a requirement the SPEC states; F2 is a correctness defect with a disclosed residual; F3 and F4 are optional and must not by themselves convert a PASS into a FAIL. The FAIL verdict is driven by the required-backend gate and by F1 plus the must-pass Security score.

## What I probed on the shrink and found sound

- Same-second claim: `preservePriorManifest` (`update_prior_manifest.go:139-161`) refuses when `manifest.prior.json` already exists in the target directory (Lstat), returns `priorManifestFailed`, creates no sibling directory, runs no `RemoveAll`; `backUpRemovalCandidates` (`migrate_project_assets.go:143-146`) returns the blocking reason before any copy. `TestSecondClaimInTheSameDirectoryFailsClosed` (4 sub-tests) passes. The sibling scheme, `copyRunMarker`, `priorManifestMarkerCopy`, `priorManifestSiblingMax` are gone (grep of the non-test sources finds no use).
- One open per run: `newOwnershipEnv` opens the root once (`migrate_project_assets.go:204`), `defer env.closeRoot()` (`:306`), preview `defer verdicts.close()` (`update_dryrun_preview.go:74`); `TestMigrationOpensTheProjectRootOnce` and `TestPreviewVerdictsCloseTheRootTheyOpened` pass. Root-path swap row (N-A) passes: the read lands on the original project, not the swapped path.
- Per-component `Lstat` through the Root, any symlink component kept unread as `symlink-escape`, dangling or looping link `read-failed`, FIFO/socket/device decided before any open: code at `:377-440`, rows in `TestPathBoundaryCounterexamples` pass.
- Path boundary: `withinRootBoundary` (`migrate_project_assets.go:641-649`) compares path elements (`rel == "."`, `rel != ".."`, no `../` prefix), not string prefixes; a name starting with two dots is inside; the 7 pure rows plus the update-shaped (`.`) rows pass.
- Keep-reason vocabulary: closed set of 14 plus the `evidence-incomplete:<item>` family in `isKnownKeepReason` (`:88-105`); `TestKeepReasonVocabularyIsClosed` and `TestVocabularyGuardControls` pass; the shrink added no reason.
- `removalEnabled`: single non-test assignment `var removalEnabled = false` (`migrate_project_assets.go:44`); `grep -rn 'removalEnabled *= *true' internal/cli --include='*.go'` minus tests prints nothing; `TestRemovalSwitchStaysOffInProduction`, `TestMigrationRemovalGate`, `TestMigrationDefaultKeepsConfirmedProjectCopy`, `TestPreviewOmitsRemovalCandidateWhileSwitchOff` pass.
- Protected files: `git diff 48bfe97afa..HEAD --name-only -- internal/config/protected_zone.go internal/config/userroots.go internal/profile` lists 0 files.
- Dry run: no update lock and no binary self-update under `--dry-run` (`update.go` diff); `TestUpdateDryRun_ZeroMutation` and `TestDryRunChangesNothingAndMatchesVerdict` pass.
- No skipped test: the only `Skip` hits in the card's test files are a comment and an unrelated `first.Skipped` field; the runs show SKIP=0.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim 1: the shrink's guarded behaviour holds on the selectors the lane named

Evidence, command (cwd = the audited tree, `TMPDIR=/private/tmp/lxq/aud/tmp`):

```
timeout 600 go test ./internal/cli/ -run '^(TestSecondClaimInTheSameDirectoryFailsClosed|TestMigrationOpensTheProjectRootOnce|TestPathBoundaryCounterexamples|TestUnregisteredSymlinkOutsideProjectIsNotRead|TestUnregisteredFileBelowASymlinkedAncestorIsNotRead|TestNonRegularFilesAreKeptWithoutBeingRead|TestInterruptedMigrationRecoversAndConverges|TestRepeatedUpdateIsIdempotent|TestPreviewVerdictsCloseTheRootTheyOpened)$' -count=1 -v > /private/tmp/lxq/aud/run1.txt 2>&1
```

Observed: `exit=0`; tail of the output file:

```
    --- PASS: TestSecondClaimInTheSameDirectoryFailsClosed/TheLaterRunRemovesNothingAndSaysWhy (0.79s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	20.090s
PASS=101 FAIL=0 SKIP=0
```

(counts are `grep -c -- '--- PASS|FAIL|SKIP'` over the saved file; the swept set is 101, not empty.)

### Claim 2: the acceptance criteria outside the shrink surface that I judged still pass

Command: `timeout 600 go test ./internal/cli/ ./internal/manifest/ ./internal/cli/update/backup/ -run '^(TestOwnershipVerdictClasses|TestMigrationClassifiesUnregisteredMirrorCopy|TestOwnershipPreserveByteEqualUserFile|TestOwnershipUnregisteredUnknownOriginKeptWithReason|TestOwnershipAttestedManagedOnlyWithFullEvidence|TestPriorGenerationSourceBoundedAndTolerant|TestPriorManifestRetainedInRunBackup|TestNoProfileRootAccessInUpdateSources|TestProfileEnvDoesNotReachProfileRoot|TestProfileScanCatchesMutants|TestDryRunChangesNothingAndMatchesVerdict|TestPreviewOmitsRemovalCandidateWhileSwitchOff|TestUpdateDryRun_ZeroMutation|TestMigrationDefaultKeepsConfirmedProjectCopy|TestMigrationRemovalGate|TestRemovalSwitchStaysOffInProduction|TestBackupFailureRemovesNothingThroughProductionPath|TestPriorManifestAtomicWriteFailures|TestPriorGenerationCorruptIsNoEvidence|TestManifestRecordsDeployedVersion|TestKeepReasonVocabulary.*|TestVocabulary.*)$' -count=1 -v`

Observed: `exit=0`

```
ok  	github.com/modu-ai/moai-adk/internal/cli	19.449s
ok  	github.com/modu-ai/moai-adk/internal/manifest	0.258s
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.313s [no tests to run]
PASS=100 FAIL=0 SKIP=0
```

The `[no tests to run]` for the backup package is the empty-sweep token (that selector names no test of that package); I did not read it as a pass for that package and ran the backup selectors separately (Claim 3). Passing tests seen by name: TestBackupFailureRemovesNothingThroughProductionPath, TestDryRunChangesNothingAndMatchesVerdict, TestKeepReasonVocabularyIsClosed, TestManifestRecordsDeployedVersion, TestMigrationClassifiesUnregisteredMirrorCopy, TestMigrationDefaultKeepsConfirmedProjectCopy, TestMigrationRemovalGate, TestNoProfileRootAccessInUpdateSources, TestOwnershipAttestedManagedOnlyWithFullEvidence, TestOwnershipPreserveByteEqualUserFile, TestOwnershipUnregisteredUnknownOriginKeptWithReason, TestOwnershipVerdictClasses, TestPreviewOmitsRemovalCandidateWhileSwitchOff, TestPriorGenerationCorruptIsNoEvidence, TestPriorGenerationSourceBoundedAndTolerant, TestPriorManifestAtomicWriteFailures, TestPriorManifestRetainedInRunBackup, TestProfileEnvDoesNotReachProfileRoot, TestProfileScanCatchesMutants, TestRemovalSwitchStaysOffInProduction, TestUpdateDryRun_ZeroMutation, TestVocabularyGuardControls.

### Claim 3: the restore walks and the manifest version stamp

Command: `timeout 600 go test ./internal/cli/update/backup/ ./internal/manifest/ -run '^(TestRestore.*PriorManifest.*|TestIsPriorManifestBookkeeping|TestManifestRecordsDeployedVersion)$' -count=1 -v`

Observed: `exit=0`, `ok  github.com/modu-ai/moai-adk/internal/cli/update/backup 0.181s`, `ok  github.com/modu-ai/moai-adk/internal/manifest 0.085s`, `PASS=14 FAIL=0 SKIP=0`.

### Claim 4: static checks

Commands and observations: `gofmt -l internal/cli internal/manifest` printed nothing, `gofmt-exit=0`. `timeout 600 go vet ./internal/cli/ ./internal/manifest/ ./internal/cli/update/backup/` printed nothing, `vet-exit=0`. `git rev-parse HEAD` printed `4a625be15622eeb610df317807831d51e4257e36`; `git status --short` printed nothing.

### Baseline-attribution

All test and static results above were produced in this run against the tree at HEAD `4a625be15622eeb610df317807831d51e4257e36` (clean), using the installed `go` of the toolchain pinned by go.mod (`go1.26.8 darwin/arm64`). The tests are built from the tree; the moai binary was not used to produce any of them. The audit_multi result names `build_commit: 7705b365e2` as the judging build (see the header for its relation to HEAD). Output files: `/private/tmp/lxq/aud/run1.txt`, `run2.txt`, `run3.txt`, `vet.txt` (machine-local scratch, not citation targets; the deciding lines are quoted above).

Admission record for every Go call (read before the call, in this run):

| Go call | reading time (UTC) | load average 1-min | free memory | go-test-heavy slot holder |
|---|---|---|---|---|
| Claim 1 selectors | 2026-10-19T01:31:52Z (after a 30 s stagger from the 01:31:20Z reading) | 14.45 | 81% | lane-35, "t1791 D2 correction" (development, not a landing or remeasure record) |
| Claim 2 selectors | 01:32:49Z | 11.91 | 81% | lane-45, command "(not given)", since 01:32:26Z |
| Claim 3 selectors | 01:33:21Z | 11.69 | 81% | lane-45, command "(not given)" |
| gofmt then go vet | 01:33:57Z | 13.48 | 81% | lane-45, command "(not given)" |

The first reading (01:31:20Z) showed load 18.77. One Go call ran at a time, each under `timeout 600`, output redirected to a file, `TMPDIR` under `/private/tmp/lxq/aud/tmp`, no card id in any temp path, nothing killed, no command refused by a guard.

### Gaps (explicitly NOT observed)

1. The full `internal/cli` package suite, coverage measurement, `-race`, and Windows test execution were not run (landing remeasure's job per the brief). Craft is therefore UNVERIFIED, not PASS.
2. `golangci-lint` was not run; the lane's record of `--new-from-rev=aba63b1483` 0 issues was not re-observed.
3. Acceptance criteria AC-OWR-09 to -15 (commands/agents, case variants, empty directories, remaining list, backup/restore bytes) were not run by me; the shrink did not change those code paths (progress.md section E.2 chunk 13 scope guard), but that is the lane's statement, not my observation.
4. F1 and F2 are established by reading the code lines cited and by the required backend's reproduction; I did not execute either reproduction, because writing a scratch test into the tree is outside my allowed write set and a test outside the package cannot reach the unexported functions. The F3 mechanism was confirmed by reading the code and the `os.PathError` message form, not by execution.
5. Windows behaviour of `os.Root` on junctions and reparse points was not observed.
6. The slot holder's command for three of the four calls read "(not given)"; I could not tell from the slot record whether it was a landing or remeasure session. I treated "(not given)" as not a landing record and recorded it here.
7. The `.moai/lessons-inbox.jsonl` was not consulted for refused tool calls; no tool call was refused in this run.
8. Plan-phase binding run conditions (`Binding run conditions` heading in progress.md) were not re-read item by item in this round; the FAIL verdict does not depend on them.

### Residual-risk

- A local writer inside the project during a run can still hang the update through the FIFO swap (F2), and a project link into a profile root still gets its target stat'ed (F1) until fixed.
- The registered-file read (`os.ReadFile(abs)`, `unreadableKind`) still follows links by path; disclosed debt, bytes never leave memory unless the path is removal-eligible, which needs the boundary check and the switch that is off.
- Reason labels can be wrong for file names containing the stdlib escape text (F3).
- The escape detection by the stdlib error text stays a documented limit; it fails closed.
- Coverage and the whole-package result are unmeasured here; a regression outside the selected tests would not be seen.

## Iteration history

| Round | HEAD | Result | Findings |
|---|---|---|---|
| 1 | dfd7c40fb2 | FAIL (codex) | F1 unregistered symlink read before the boundary check; F2, F3 |
| 2 | 38c8ca5376 | FAIL (codex) | temp-name filter, sibling retention, boundary fixes |
| 3 | 1d9b06a737 | FAIL (codex) | link swapped between check and read; leaked sibling directory |
| 4 | aba63b1483 | FAIL (codex) | N-A root path swap, N-B over-broad RemoveAll, N-C absolute in-project link |
| 5 (this file) | 4a625be156 | FAIL (codex) | N-A, N-B, N-C closed by the shrink; new P2 x2 (stat through a link, FIFO race) and P3 (file name in error text) |

Round 5 is a from-scratch judgment of the final tree, not a delta. Between round 4 and round 5 the shrink commits `b2ad99d068`, `36f3315244`, `7c3abe0e0d`, the progress record `6e28a19a97`, the SPEC wording `701adc3edd` and the docs `4a625be156` landed.

## Recommendations

- Fix F1 first (one call removed, plus the differential row); it is the only finding that touches a stated requirement.
- Fix F2 in the same change (non-blocking open plus a handle-kind re-check, plus a FIFO swap race row).
- F3 and F4 are optional.
- A confirming re-audit after the fix should be scoped to F1, F2 and the lines they touch, with the required codex gate called once.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-e012674ec01973db90b4bb52
