verdict: FAIL
audited_sha: 38c8ca53767a32552b8aa559df9d1bb5a04fdf5b
required_backend: codex fail
convergence_overall: fail

auditor-model: claude-sonnet-5-5

# Sync audit round 2 (delta), card t1865, SPEC-UPDATE-OWNERSHIP-REPAIR-001 (Tier L, tdd)

Overall: FAIL. The required codex backend returned `fail` again, so PASS and PASS-WITH-DEBT are excluded. This time the codex failure is substantive, not the "no go files to analyze" sandbox failure: codex ran probes and returned two P2 findings. I checked both myself. The first is reproduced by execution (a SHOULD-FIX data-loss edge in the legacy restore filter that the code's own comment says it does not have). The second is confirmed by reading the code and is a MINOR fail-safe residue.

All three round-1 findings (F1 BLOCKING, F2, F3) are fixed in the code at HEAD, and each fix's anchored test passes in my run. The round-1 blocking defect is closed. What keeps the verdict at FAIL is the unmet required gate plus the new restore-filter edge (N1).

Round 1 is preserved as `.moai/reports/t1865/sync-audit-r1.md` (not edited).

HEAD check: `git rev-parse HEAD` printed `38c8ca53767a32552b8aa559df9d1bb5a04fdf5b`, equal to the requested SHA.

Judging builds (named per verification-claim-integrity 2.2):
- My own Go measurements (`go test`, `go build`, `gofmt`) were built and run from this tree at HEAD `38c8ca5376`. The installed `moai` CLI was not used as a measuring tool.
- The `audit-plan` verb and the `audit_multi` call were served by the installed moai build `ee981a932`, which the tool itself reports as an ancestor of HEAD (`build_lag`: "an ancestor of this tree's HEAD 38c8ca537"). It lags, so the plan and the codex dispatch come from older code. The codex review read this tree's diff, and its two cited repros are about code in this tree.

## Claim

1. F1 (unregistered symlink read outside the project) is fixed. `classifyOwnershipCore` now calls `pathLeavesRoot` (EvalSymlinks of the WHOLE path, so a symlinked ancestor directory is covered) before `unreadableKind` and before any read, and keeps the file unread with `symlink-escape`.
2. F2 (same-second overwrite of the earlier `manifest.prior.json`) is fixed: `claimPriorManifestDir` creates a sibling `<runDir>-02..-99` with `Mkdir`, copies the run marker, and fails closed when no name is free. The reader's descending name sort puts the sibling before its base, so age order holds.
3. F3 (FIFO at the user-root counterpart hung the verdict) is fixed: `userCounterpartConfirmed` returns false for a counterpart whose kind is unreadable before opening it.
4. The two review-round-4 fixes are present: a failed prior-manifest rename removes its temporary file; `backup.IsPriorManifestBookkeeping` makes the legacy restore skip the temporary name. The second one has a new edge (N1 below).
5. Hard invariants hold: `internal/config/protected_zone.go`, `internal/config/userroots.go` and `internal/profile` are byte-unchanged against `cdd72a814`; `removalEnabled` is `false` in production code; no `t.Skip`; no stray files; no code added by the card opens a profile root.
6. The sync documents (CHANGELOG, progress.md E.4) state the repair honestly except for two small understatements (N1, N2).

## Evidence

### Commands run and verbatim output

`git rev-parse HEAD`:
```
38c8ca53767a32552b8aa559df9d1bb5a04fdf5b
```

`moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865` (abridged, quoted): `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]`, gates: claude off, codex required, glm off.

Hard invariants:
```
git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
inv=0
```
`removalEnabled` in non-test code (grep over `internal`, `*.go` minus `_test.go`):
```
internal/cli/update_dryrun_preview.go:84:		if removalEnabled && isCommonAssetCleanTarget(f.RelPath) && verdicts.removalEligible(projectRoot, mf, f.RelPath) {
internal/cli/migrate_ownership.go:162:	removalEnabled bool
internal/cli/migrate_ownership.go:410:	if !env.removalEnabled {
internal/cli/migrate_project_assets.go:40:var removalEnabled = false
internal/cli/migrate_project_assets.go:205:		removalEnabled: removalEnabled,
```
The only assignment is `var removalEnabled = false`.

Skips in test files changed since base (`grep -n "t.Skip\|SkipNow\|testing.Short"` over the card's changed `_test.go`): only `migrate_ownership_helpers_test.go:1008-1009` (`first.Skipped`, a run-result field) and `migrate_ownership_symlink_test.go:8` (a comment). No `t.Skip` call. `git ls-files -o --exclude-standard` printed nothing before this report.

`git diff --stat dfd7c40fb2 HEAD -- . ':!.moai'`: 14 files changed, 1104 insertions, 114 deletions.

Format and build (this tree): `gofmt -l internal/cli internal/cli/update/backup` printed nothing; `go build ./...` exit 0 (`build=0`).

Anchored selectors, run sequentially with `TMPDIR=/private/tmp/own-fixture-d1`, load average about 30 at the start:
```
--- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead (3.06s)
--- PASS: TestCounterpartFifoIsNotOpened (2.27s)
--- PASS: TestUnregisteredSymlinkOutsideProjectIsNotRead (1.10s)
--- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun (0.11s)   [6 sub-tests incl. RetentionStillCountsTheSiblingWithItsRun, NoFreeNameFailsClosedAndOverwritesNothing]
--- PASS: TestFailedPriorManifestRenameLeavesNoTemporaryFile (1.72s)
ok  	github.com/modu-ai/moai-adk/internal/cli	9.603s
--- PASS: TestRestoreLegacySkipsPriorManifestTemporaryFile (0.00s)
--- PASS: TestRestoreRetainedWithoutSectionsSkipsPriorManifestTemporaryFile (0.01s)
--- PASS: TestRestoreRetainedWithSectionsSkipsPriorManifestTemporaryFile (0.01s)
--- PASS: TestRestoreFromBackupDirSkipsPriorManifestTemporaryFile (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.279s
```
The round-1 overlay test that reproduced F1 (`TestT1865AuditUnregisteredProfileSymlink`) is no longer in the scratch overlay form; the card's own `TestUnregisteredSymlinkOutsideProjectIsNotRead` replaces it, and I read its subject (a symlink to a regular file outside the project, asserting no bytes read and keep reason `symlink-escape`) from the code it exercises, not by re-running the old overlay.

### Code reading (the fixes)

- `internal/cli/migrate_ownership.go:286-298`: for an unregistered path the order is `pathLeavesRoot` then `unreadableKind` then `classifyUnregistered`. `pathLeavesRoot` (lines 351-360) calls `filepath.EvalSymlinks(abs)` and returns `!withinRootBoundary(rootResolved, resolved)`; a path that does not resolve is not reported and keeps the prior `read-failed` behavior. It never opens the target. A symlinked ancestor directory is covered because EvalSymlinks resolves every component.
- `internal/cli/migrate_project_assets.go:541-548` (`userCounterpartConfirmed`): the counterpart is checked with `unreadableKind(counterpart)` before `os.ReadFile`.
- `internal/cli/update_prior_manifest.go:155-208`: `claimPriorManifestDir` Lstat-checks `manifest.prior.json`; if present it tries `-02..-99` with `os.Mkdir` (an existing name is skipped, any other error is returned) and `copyRunMarker` (Lstat regular check, then read, then `backup.WriteRunMarker`). `readPriorGenerations` (lines 262-310) sorts directory names descending and Lstat-checks the file kind, so a sibling `X-02` is read before `X`, and a non-regular `manifest.prior.json` is skipped unread.
- Retention (`internal/cli/update/backup/backup.go` around `configBackupName`, and the test `RetentionStillCountsTheSiblingWithItsRun`): a marked sibling is grouped with its run by run id; an unmarked directory whose name does not match `^[0-9]{8}_[0-9]{6}$` is "neither counted nor deleted". So a sibling whose marker copy failed is never pruned (N2).
- `readStartManifest` (`update_prior_manifest.go:127`) reads `.moai/manifest.json` with `os.ReadFile`. `internal/manifest/manifest.go:88` reads the same path with `os.ReadFile` at base and in the run's earlier load, so this adds no new path to hang on or follow; not a new defect.
- `migrate_ownership_emptydir.go`: `WalkDir` does not follow symlinks (a symlink is a non-dir entry, so the tree is not a candidate), `ReadDir` entries are filtered with `e.IsDir()` (false for a symlink), and `pruneNestedEmptyDir` re-resolves and boundary-checks. No new read-through-symlink route found.

### The restore filter edge, reproduced (N1)

`IsPriorManifestBookkeeping(relPath)` is `relPath == PriorManifestName || strings.HasPrefix(relPath, PriorManifestTempPrefix)`. `relPath` in `RestoreMoaiConfigLegacy` is the whole path relative to the backup root, with separators. A user file below a directory that merely starts with the temporary prefix therefore matches and is skipped. The code comment at `restore.go:269-271` says "Only the root-level names are meant; a same-named file below the root is user content", which the test only checks for a subdirectory named `notes`.

My reproduction (scratch test placed through `-overlay`, nothing written into the tree; source file `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3485ddff-4832-470a-9a17-ee4dac7ad178/scratchpad/zz_r2_test.go`):
```
go test ./internal/cli/update/backup/ -overlay <ov.json> -run '^TestR2AuditTmpPrefixedDirectoryContent$' -count=1 -v
=== RUN   TestR2AuditTmpPrefixedDirectoryContent
    zz_r2_test.go:16: restored-user-file-below-prefixed-dir err=stat .../.moai/config/manifest.prior.json.tmp-notes/keep.txt: no such file or directory
    zz_r2_test.go:17: user file below a directory named with the temp prefix was not restored
--- FAIL: TestR2AuditTmpPrefixedDirectoryContent (0.01s)
FAIL
```
`git status --short` printed nothing afterward.

### audit_multi (required backend gate), quoted from the call result

Call: `mcp__moai__audit_multi` with `project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, `target=baseBranch`, a focus string restating the delta, no `gates` argument.
- `per_backend_verdicts[0].backend`: `codex`; `gate`: `required`; `verdict`: `fail`
- `overall_verdict`: `fail`
- `disagreement_flag`: `null`; `participant_count`: 1; `residual_risk_note`: `required-backend FAIL: codex`; `fail_open_backends`: `[]`
- `audit_receipt`: `rcpt-aeb48f1d567c915a06f7b90f`
- `plan_source`: `config`; `gate_unmet`: not present in the returned object
- `build_commit`: `ee981a932`, `build_lag` reports the installed binary as an ancestor of HEAD `38c8ca537`.

The codex failure is NOT the "no go files to analyze" sandbox reason. Codex ran probes at HEAD and returned two P2 findings, with the control run "restore function replaced by the `dfd7c40fb2` version" restoring the file:
1. P2, `internal/cli/update/backup/run_marker.go:38`: the temporary-file filter also drops user content below a directory such as `manifest.prior.json.tmp-notes/keep.txt`. Reproduced by me (above). N1.
2. P2, `internal/cli/update_prior_manifest.go:188`: when the run marker cannot be copied to a new sibling, the sibling stays out of the retention census for good (`PruneBackups(root, 0, 1)` gave `runs=1 unmanaged=1 removed=1`, the sibling remained). Confirmed by reading `copyRunMarker` (best effort, return on any failure) and the census rule above. N2.

Per the per-backend gate, no PASS or PASS-WITH-DEBT is possible. This verdict is not final until the orchestrator's check passes: the orchestrator writes the digest members (`overall_verdict`: fail, `gate_unmet`: absent, `plan_source`: config; codex/required/fail) to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`. A `convergence_check.ok: false` is an unmet gate named by backend. I do not write that file.

## Baseline-attribution

- Every observation above was made in this run against tree HEAD `38c8ca53767a32552b8aa559df9d1bb5a04fdf5b`, worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, `TMPDIR=/private/tmp/own-fixture-d1`, Go toolchain built from this tree. Judging build coordinates for the moai tool calls: `ee981a932` (lagging; stated at the top).
- Diff base for the card: `cdd72a814`. Delta base for this audit: `dfd7c40fb2` (`git diff --stat dfd7c40fb2 HEAD -- . ':!.moai'`: 14 files, 1104 insertions, 114 deletions).
- The lane's scoped figures (68 top-level PASS in internal/cli + internal/manifest, backup package ok, new tests 3x green, windows build, vet, lint with 3 untouched errcheck findings in `internal/cli/factory_run_adopt.go`) were read from the brief and progress.md. I did not re-observe them; I re-ran only the nine anchored tests listed, `go build ./...` and `gofmt -l`.
- The SPEC-body claim (design.md D-1 only) was checked with `git log` on the three SPEC files: spec.md changed after plan only by status commits `7d8be68f6c`, `e99a990296`; design.md by `0a33265120`; plan.md, acceptance.md, research.md unchanged since `073b3e1382` (read from the progress.md E.4 statement; the `git log` of spec/acceptance/design lists `0a33265120`, `e99a990296`, `7d8be68f6c`, `073b3e1382` and earlier plan commits, which agrees).

## Binding run conditions (12, from round 1; re-read in E.3)

N4-remainder, N14-N24 each still carry a disposition in progress.md E.3 as in round 1. I did not re-measure each cell in this round (see Gaps). N16 (no profile-root-opening code added; open item disclosed) still holds for the card-added code: F1 closed the one card-added route that read through a user symlink. The registered-file read before the boundary check at base is unchanged and disclosed in E.4 as the N16 family debt.

## RED/arrival cells and test honesty (chunks 9-10, read from progress.md E.2)

- Chunk 9 ancestor-link cell: RED output at progress.md lines ~3433-3441 shows `class="unregistered" keep="no-provenance-record"` where `symlink-escape` is wanted and "the verdict carried 34 bytes read from outside the project": red for the stated reason before green.
- Chunk 10 (B) temp-file cell: RED lines ~3471-3473 (`temporary files remain after the failed rename`) and ~3487-3494 (`manifest.prior.json.tmp-4242 from the backup root was copied into the config tree`): red for the stated reason.
- I did not open the F1, F2, F3 RED cells of the earlier chunks (7-8), and I did not run the mutants. For F1/F3/F2 I rely on the code reading plus the green anchored tests above. See Gaps.
- Vacuity probe of the new restore test: the fixture checks a subdirectory named `notes` only, so it cannot see N1 (a directory that itself starts with the temporary prefix). The mutant "HasPrefix with no separator check" satisfies the test, which is exactly why N1 survived.

## Score (flat weighted model)

| Dimension | Score | Verdict | Basis |
|---|---|---|---|
| Functionality (40%) | 85/100 | FAIL (gate) | F1, F2, F3 fixed and green; N1 restore edge reproduced; N2 residue. The must-pass firewall is not tripped by these, the FAIL comes from the required backend gate |
| Security (25%) | 88/100 | PASS | F1 closed (whole-path EvalSymlinks before any read), F3 closed (no open of a FIFO counterpart), no card-added route reads outside the project or opens a profile root; registered-file read-before-boundary is a disclosed base debt, unchanged |
| Craft (20%) | 80/100 | PASS | gofmt clean, build 0, no skips, nine anchored tests green; N1 test blind spot; package-wide test, coverage, `-race`, Windows tests not run (carried to CI) |
| Consistency (15%) | 86/100 | PASS | style matches; the unregistered path now has a pre-read boundary guard that the registered path lacks (disclosed); docs honest with two small understatements |

Overall: FAIL because the required codex gate returned `fail` (receipt `rcpt-aeb48f1d567c915a06f7b90f`). Without that gate the code-level picture would be PASS-WITH-DEBT at best, with N1 as the only item I would want fixed first.

## Findings

- **N1 [SHOULD-FIX]** `internal/cli/update/backup/run_marker.go:38` (`IsPriorManifestBookkeeping`), used at `internal/cli/update/backup/restore.go:273`. `strings.HasPrefix(relPath, PriorManifestTempPrefix)` has no separator check, so every file below a backup-root directory whose name starts with `manifest.prior.json.tmp-` is silently skipped by the legacy restore. The code comment promises "only the root-level names". Reproduced by execution (above) and by codex. Practical likelihood is very low (a user directory with that exact prefix at `.moai/config` root), but it is a silent data drop in a restore path and contradicts the stated contract. Required fix: make the temporary-name match root-level only (for example `!strings.ContainsAny(relPath, "/\\") && strings.HasPrefix(...)`), and extend `restore_prior_manifest_tmp_test.go` with a fixture directory named with the prefix, test-first.
- **N2 [MINOR, optional]** `internal/cli/update_prior_manifest.go:188-208` (`copyRunMarker`). If the run marker cannot be copied to a new sibling, the sibling is left without a marker, its name fails the config-root name pattern, and retention never counts or prunes it (a small directory that stays). The code comment documents "counted as unmanaged, never removed" as a deliberate fail-safe (a removal is never wrongly made), and the prior-manifest evidence itself is preserved and still read. progress.md E.4 caveat 3 says the marker is copied "so retention still counts one run" without naming this failure residue. Suggested: either name the residue in E.4, or make the sibling use the pattern retention recognises; not required.
- **N3 [MINOR, disclosed]** Registered managed file that is a symlink out of the project: its target bytes are still read for the hash compare before the boundary check (`classifyOwnershipCore` registered branch). Unchanged from the base, disclosed in E.4 as the N16-family debt, and removal is refused as `symlink-escape`. The unregistered branch now has a pre-read guard that the registered branch does not; the lane declined to change the pinned AC-OWR-01 class expectation. Carry as debt.
- **N4 [MINOR]** progress.md E.4 caveat 8 states the declined-finding limit as "a user config file ... literally named `manifest.prior.json`, or starting `manifest.prior.json.tmp-`". The code also drops files below a directory with that prefix (N1), a wider effect than the sentence says. Resolved by fixing N1 or widening the sentence.
- **N5 [MINOR]** AC count: the SPEC-wording debts (REQ-OWR-006/AC-OWR-10, AC-OWR-21 path and case (o), AC-OWR-20 (b), AC-OWR-06 (b), AC-OWR-08 glob, AC-OWR-12 referent) and the 21 vs 22 B12 counter (acceptance.md line 19 cites AC-COLLECTOR-ANCHOR-001) remain disclosed, bodies unedited. No new item.
- Round-1 findings status: F1 fixed (verified); F2 fixed (verified, plus the scheme parses correctly against the reader and the census); F3 fixed (verified); F4 declined and still outside REQ-OWR-008, unchanged; F5/F6 disclosed and unchanged.

## Gaps (not observed)

- Not run by me, and per the lane not run by it either: package-wide `internal/cli` test, package coverage, `-race`. No Windows test execution (build and vet only, per the lane). `GOOS=windows go build`, `go vet` and golangci-lint were not re-run by me; `go build ./...` (host) and `gofmt` were.
- I did not re-run the lane's 68 top-level PASS set. I ran nine anchored tests plus one scratch overlay test.
- I did not open the RED cells of chunks 7-8 (F1/F2/F3 fixes) or run the mutants for them; chunks 9-10 RED lines were read only by the `grep`-selected excerpts above, not the full cell bodies.
- I did not re-measure the 12 binding conditions cell by cell, nor re-count the frozen call-site baseline row set, nor test a case-sensitive file system.
- Time-of-check/time-of-use races (a regular file swapped for a FIFO or symlink between the check and the read) are not covered, as the lane discloses; I did not test them.
- The dry-run no-lock and no-installer changes (`update.go`) were not re-opened in this round; they were accepted in round 1 (REQ-OWR-011, disclosed).
- The installed moai binary lags HEAD (`ee981a932` is an ancestor), so any CLI behaviour of the tree that I did not build from source was not judged.
- No refused tool call occurred in this audit. The `audit-plan` call succeeded; no `.moai/state` file was edited or deleted by me.

## Residual-risk

- The defect class "a read of a project or user-root path reaches a file the card should not open" is now guarded for the unregistered branch and the counterpart read, but still open for the registered-file read at base and for time-of-check races. The AC-OWR-08 source scan plus differential canary only sees fixtures that exist (N16).
- The card-review and sync-audit rounds each found a new edge in the newest code (round 4 review, then N1/N2 here). The two new edges are both in small, best-effort bookkeeping helpers; another pass may find more.
- A user-controlled root-level name (`manifest.prior.json`, `manifest.prior.json.tmp-*`) is now reserved in legacy restores; a directory with that prefix loses its content on restore until N1 is fixed.
- The sibling scheme caps at 99 collisions per second-named directory and fails closed (the run removes nothing) beyond that; this cannot be reached in normal use.

## Iteration history

- Round 1 (HEAD `dfd7c40fb2`): FAIL, codex fail, receipt `rcpt-7770ccadc2f78594b050e777`; F1 BLOCKING, F2/F3 SHOULD-FIX. File kept as `sync-audit-r1.md`.
- Round 2 (this file, HEAD `38c8ca5376`, delta scope F1-F3 plus regression and same-class hunt): FAIL, codex fail, receipt `rcpt-aeb48f1d567c915a06f7b90f`; F1-F3 verified fixed; new N1 SHOULD-FIX, N2 MINOR.
- Verdict path forward: fix N1 test-first (root-level-only temporary-name match, plus a prefix-named-directory fixture); optionally name or remove the N2 residue; re-run the codex gate. Re-audit scope: N1 (and N2 if touched).

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-aeb48f1d567c915a06f7b90f
