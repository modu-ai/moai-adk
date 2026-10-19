verdict: FAIL
audited_sha: aba63b148319b25ec44080c89071a2b99462a8bd
required_backend: codex fail
convergence_overall: fail

auditor-model: claude-sonnet-5-5

# Sync audit round 4 (delta, last round before the rule C hold) - SPEC-UPDATE-OWNERSHIP-REPAIR-001, card t1865

Rounds 1-3 are preserved byte-identical as sync-audit-r1.md, sync-audit-r2.md, sync-audit-r3.md. This file replaces the round-3 file as the current verdict.

## Overall

FAIL. Both round-3 findings (F-A, F-B) are fixed in the code and their tests pass in my own run. The required codex backend nevertheless returned `fail` with two new P2 findings, and I reproduced BOTH by execution in this run (a scratch test through `-overlay`, nothing written in the tree). One of them (N-A) is a residue of the same check-then-use class as F-A: the repair confines the read to an `os.Root`, but the Root is opened from the project-root path STRING after the boundary check, so replacing the root path in that window makes the Root point outside. The codex failure is substantive; it is not the "no go files to analyze" sandbox reason. A required-backend fail forces FAIL.

## Claim

1. F-A is fixed in the code: an unregistered file is read through `os.OpenRoot(root).Open(rel)`. The two post-check windows the lane targeted (before the path is resolved again, after it is resolved and before it is opened) are closed; the race rows are green and the lane's mutants show each window red without the fix.
2. F-B is fixed in the code: `claimPriorManifestDir` removes the sibling it created (with `os.RemoveAll`) when the marker copy fails after the marker file exists; the new subtest is green. (The removal is broader than needed: see N-B.)
3. NEW N-A [SHOULD-FIX]: the Root is opened from a path string after the boundary check, so a swap of the project root path in that window reads a file outside the project. Reproduced by me (1 outside read in 2079 attempts).
4. NEW N-B [MINOR]: `os.RemoveAll(sibling)` also deletes content a foreign process put into the freshly created sibling. Reproduced by me with a seam that simulates the foreign write.
5. The hard invariants hold at HEAD, the documents match the code, and no new defect was found in the other helpers (class check below).

## Evidence

### Tree and plan

`git rev-parse HEAD` printed `aba63b148319b25ec44080c89071a2b99462a8bd` (equal to the requested SHA). `git rev-parse --show-toplevel` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`. `git status --short` printed nothing (0 lines) at the start and at the end.

`moai verify audit-plan --project-root <toplevel>`: `config_status: ok`, `model: codex`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`; claude gate off, codex gate required, glm gate off.

### Tests I ran (each in its own call through a wrapper that waits for the admission rule and re-checks after 30 seconds)

Admission values at each Go call (load1 / free memory / slot / lane script processes, wall clock UTC). Before the first call I waited because the lane's own `c12b-verify.sh` was running Go commands (processes found by `pgrep`, last Go step finished 17:21:45Z) and the landing slot was held by lane-32 under "landing"/"remeasure" commands until it freed at about 17:18Z.

```
main     BEFORE load=12.59 free=83% slot=held (not landing/remeasure at the check) lane-script-pids=0 17:22:48Z   AFTER load=30.07 17:23:14Z
scratch  BEFORE load=27.94 free=83% slot=free lane-script-pids=0 17:25:23Z   AFTER load=26.08
scratch2 BEFORE load=28.48 free=83% slot=free lane-script-pids=0 17:26:17Z   AFTER load=23.48
backup   BEFORE load=18.74 free=83% slot=free lane-script-pids=0 17:27:07Z   AFTER load=19.48
manifest BEFORE load=13.74 free=83% slot=free lane-script-pids=0 17:27:41Z   AFTER load=13.74
build    BEFORE load=25.81 free=83% slot=free lane-script-pids=0 17:28:15Z   AFTER load=54.32 (others' load, after my build had ended)
winbuild BEFORE load=25.08 free=83% slot=held lane-script-pids=0 17:30:10Z   AFTER load=25.07
vet      BEFORE load=23.08 free=83% slot=held lane-script-pids=0 17:30:47Z   AFTER load=22.28
scratch3 BEFORE load=28.26 free=83% slot=held lane-script-pids=0 17:41:02Z   AFTER load=27.45
```
The slot, when shown as `held` at my checks, was held by lane-22 with the command "t1868 M5 mutant run plus final-HEAD scoped test, vet, build (window 2, board e1b8)" (read at 17:30Z and again after the last call), which contains neither "landing" nor "remeasure"; condition (c) held as written. The `build` and `winbuild` calls were issued in one turn but ran one after another (start times 17:27:41Z/17:28:15Z, 17:30:10Z): one Go call at a time. I did not take a slot and did not use `moai slot acquire`. No failure carried a timeout or a load-dependent cause.

Anchored tests, run from this tree with `TMPDIR=/private/tmp/own-fixture-d1`:
```
go test ./internal/cli/ -run '^(TestPathBoundaryCounterexamples|TestMarkerCopyFailureFailsClosedWithoutResidue|TestUnregisteredFileBelowASymlinkedAncestorIsNotRead|TestUnregisteredSymlinkOutsideProjectIsNotRead|TestCounterpartFifoIsNotOpened|TestNonRegularFilesAreKeptWithoutBeingRead|TestPriorManifestOfAnEarlierRunSurvivesALaterRun)$' -count=1 -v
--- PASS: TestUnregisteredFileBelowASymlinkedAncestorIsNotRead (5.82s)
--- PASS: TestPathBoundaryCounterexamples (2.87s)
--- PASS: TestCounterpartFifoIsNotOpened (1.79s)
--- PASS: TestNonRegularFilesAreKeptWithoutBeingRead (1.56s)
--- PASS: TestUnregisteredSymlinkOutsideProjectIsNotRead (3.01s)
--- PASS: TestPriorManifestOfAnEarlierRunSurvivesALaterRun (0.48s)
--- PASS: TestMarkerCopyFailureFailsClosedWithoutResidue (3.38s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	20.168s
rc=0
```
(7 top-level PASS, 0 FAIL, 0 SKIP; the table printed 46 `TABLE|` rows.) The race rows, verbatim from that run:
```
TABLE|race|target swapped for a link to an outside file between the boundary check and the read|expected: class=unregistered keep=symlink-escape read=none swapped=true|observed: class=unregistered keep=symlink-escape read=none swapped=true
TABLE|race|target swapped for a link to an outside file after the path was resolved, before it is opened|expected: class=unregistered keep=symlink-escape read=none swapped=true|observed: class=unregistered keep=symlink-escape read=none swapped=true
TABLE|race|control: target swapped for a link to ANOTHER INSIDE file is still judged by content|expected: class=unregistered keep=no-provenance-record read=other-bytes swapped=true|observed: class=unregistered keep=no-provenance-record read=other-bytes swapped=true
```
The F-B subtest `TestMarkerCopyFailureFailsClosedWithoutResidue/FailureAfterTheMarkerFileWasCreatedLeavesNoSibling` printed `--- PASS` (0.07s).

```
go test ./internal/cli/update/backup/ -count=1   ->  ok  github.com/modu-ai/moai-adk/internal/cli/update/backup	0.698s   rc=0
go test ./internal/manifest/ -count=1            ->  ok  github.com/modu-ai/moai-adk/internal/manifest	0.437s       rc=0
go build ./...                                   ->  rc=0 (no output)
GOOS=windows go build ./...                      ->  rc=0 (no output)
go vet ./internal/cli/ ./internal/cli/update/backup/  ->  rc=0 (no output; 2 s, build cache)
```

### Scratch reproductions (my own, through `-overlay`; sources in the audit scratchpad `zz_r4_test.go` and `zz_r4b_test.go`; `git status --short` printed nothing afterwards)

Class check over `readUnregisteredInsideRoot` and the classifier (`go test ./internal/cli/ -overlay ... -run '^TestR4AuditClasses$' -count=1 -v`):
```
T1 symlinked project root: class=unregistered keep=no-provenance-record read=content
T2 relative root: class=unregistered keep=no-provenance-record read=content
T3 directory: class=unregistered keep=read-failed read=none
T4 odd names (space, non-ASCII, "..hidden.md"): class=unregistered keep=no-provenance-record read=content
T5 in-project ancestor link: class=unregistered keep=no-provenance-record read=content
T6 absolute in-project final link: class=unregistered keep=no-provenance-record read=content
T7 fds before=5 after=5   (800 verdicts: 400 normal + 400 escaping)
T9 abs in-project swap at resolve-to-open: class=unregistered keep=symlink-escape read=none
--- PASS: TestR4AuditClasses (0.04s)
```
(A first run failed only because my own descriptor-counting helper used `/dev/fd`; the T1-T6 lines were identical. The second run is the one quoted.) T9 is a fail-closed false positive: `os.Root` refuses an ABSOLUTE symlink even when it points inside the root, so a race swap to such a link is kept as `symlink-escape` instead of being judged by content. Race-only, nothing is read, nothing removed.

New-defect reproductions (`go test ./internal/cli/ -overlay ... -run '^TestR4Audit(F1ForeignDataInSibling|F2RootSwapRace)$' -count=1 -v`):
```
zz_r4b_test.go:34: AF1 claim err=true foreign-data-exists=false
--- PASS: TestR4AuditF1ForeignDataInSibling (0.00s)
zz_r4b_test.go:78: AF2 attempts=2079 outside-reads=1
--- PASS: TestR4AuditF2RootSwapRace (0.14s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.256s
```
AF2 flips the project-root path between the real directory and a link to an outside directory in a goroutine and calls `readUnregisteredInsideRoot(root, abs)` in a loop; the loop stops at the first outside read, which came at attempt 2079 (`data == "OUTSIDE"`, `err == nil`). AF1 uses the `priorManifestMarkerCopy` seam to create `user-data/keep.txt` inside the new sibling and then fail; `claimPriorManifestDir` returned an error and the foreign file was gone.

### Hard invariants (this run, HEAD aba63b1483)

```
git diff --quiet cdd72a814 HEAD -- internal/config/protected_zone.go internal/config/userroots.go internal/profile
inv=0
```
`removalEnabled` in non-test code: the only assignment is `internal/cli/migrate_project_assets.go:40:var removalEnabled = false`. `t.Skip|SkipNow|testing.Short` over the card's changed `_test.go` files (`git diff --name-only cdd72a814 HEAD -- '*_test.go' | xargs grep`): only `migrate_ownership_helpers_test.go:1008` and `:1009` (the run-result field `first.Skipped`) and `migrate_ownership_symlink_test.go:8` (a comment); no call. `git ls-files --others --exclude-standard` printed nothing; `gofmt -l internal/cli internal/cli/update/backup` printed nothing. No added code opens a profile root: the added `os.OpenRoot` is on the project root, the counterpart read in `userCounterpartConfirmed` stays behind `unreadableKind`.

### Code read at HEAD (delta `git diff 1d9b06a737 HEAD`: 5 files, 164 insertions, 5 deletions, non-`.moai`)

- `readUnregisteredInsideRoot` (`internal/cli/migrate_ownership.go:399-445`): normalises the root (absolute, EvalSymlinks), `filepath.Abs` + `EvalSymlinks` of the target, `withinRootBoundary`, `filepath.Rel`, then `os.OpenRoot(root)` (line 428) and `r.Open(rel)` (line 434), `io.ReadAll`; both handles closed by `defer`. The Root is opened from the root STRING after the boundary check: this is N-A.
- Error text: the stdlib text is `errPathEscapes = errors.New("path escapes from parent")` (`os/file.go:421` in toolchain go1.26.8, read); the code matches it by `strings.Contains`. Disclosed limit; not re-reported.
- `claimPriorManifestDir` (`update_prior_manifest.go:177-208`): `os.RemoveAll(sibling)` at line 203: N-B.
- Unchanged and re-read: `IsPriorManifestBookkeeping` (root-level names only), `withinRootBoundary` (element-wise), `pathLeavesRoot`, `unreadableKind`, `userCounterpartConfirmed` (kind check before open), `copyRunMarker` (absent or non-regular marker is nothing to copy; CRLF), `writePriorManifest` (removes the temporary file on failure).
- The test seam `readWindowHook` is a package variable replaced by a test only; the race rows run without `t.Parallel`, set it before the goroutine starts and restore it with `t.Cleanup`, so no data race in the rows.

### audit_multi (required backend gate)

Call: `mcp__moai__audit_multi` with `project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865`, `target=baseBranch`, a focus restating the delta, no `gates` argument. Verbatim result members:
- `per_backend_verdicts[0]`: `backend: "codex"`, `gate: "required"`, `verdict: "fail"`
- `overall_verdict: "fail"`; `disagreement_flag: null`; `participant_count: 1`; `residual_risk_note: "required-backend FAIL: codex"`; `fail_open_backends: []`
- `plan_source: "config"`; no `gate_unmet` member in the returned object
- `audit_receipt: "rcpt-bc15a518bfe20c055b8f4bda"`
- `build_commit: "ee981a932"` with `build_lag` reporting the installed binary as an ancestor of tree HEAD `aba63b148`.

The codex failure is two substantive P2 findings, NOT the "no go files to analyze" sandbox reason (codex ran helper-level reproductions). Codex also said its full classification-path reproduction could not run (Go build cache blocked in its sandbox); its evidence is limited to direct helper calls. I reproduced both findings myself (above). This verdict is not final until the orchestrator's convergence check passes: the orchestrator writes `overall_verdict: fail`, `gate_unmet` (absent), `plan_source: config`, and codex/required/fail to `<toplevel>/.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <that path>`. I do not write that file.

## Baseline-attribution

- All test, build, vet and scratch measurements above were made in this run against tree HEAD `aba63b148319b25ec44080c89071a2b99462a8bd`, built and run from this worktree's sources. The installed `moai` CLI was not used as a measuring tool for any code claim.
- Judging build for the MCP/CLI calls (`audit-plan`, `audit_multi`): commit `ee981a932`, a strict ancestor of the tree HEAD (the tool reported `build_lag`). The plan and the codex dispatch therefore come from older code; the codex review read this tree's diff.
- Diff base for the card: `cdd72a814`. Delta base: `1d9b06a737`. The lane's E.2 chunk 12 figures (RED cells, five green runs, two mutants, 57+9+1 regression names, lint "0 issues") were read from progress.md; I re-observed the anchored selectors, the two packages, build, Windows build and vet, but not the mutants, the five-run repetition, the 57-name regression set or golangci-lint.

## RED/arrival cells and test honesty

- Chunk 12 RED (progress.md "시험 먼저"): two F-A race rows red before the fix (first row red for the right reason: `read=other-bytes`; second row red because the window did not exist before the fix, `swapped=false`: a weak RED, disclosed in the record), and the F-B subtest red for the right reason (`[…153045 …153045-02] remain`, `unmanaged=1`). The lane's mutants (plain `os.ReadFile`; `os.Open(filepath.Join(root, rel))` instead of `Root.Open`; `RemoveAll` -> `Remove`) are recorded red for the stated reasons with `cmp` restore checks. I did not re-run these mutants.
- Blind spot of the F-A rows, found by this audit: every race row swaps the FILE; none swaps the PROJECT ROOT path, and there is no seam before `os.OpenRoot`. The window N-A lives in is therefore untested.
- Blind spot of the F-B subtest: it makes the failure come after a marker FILE exists; it does not create a foreign entry in the sibling, so it cannot see N-B.

## Binding run conditions (12: N4-remainder, N14-N24)

Re-read in progress.md E.3 ("Binding run conditions 의 처분"): every item still carries a disposition and evidence pointer; N16 is open and disclosed (now noting that chunk 12 moved the unregistered read to `os.Root` and that the registered path is untouched), N24 is a confirmed limit, N18/N21/N22 are SPEC-wording debts. I did not re-measure any cell of this table in this round (see Gaps). Undisposed conditions: none found by reading.

## Documents versus code

- CHANGELOG entry: `final run code commit dc32a94c9e` equals the last non-test Go change (`git log -3 -- '*.go'` lists `dc32a94c9e`, `714adcc444` (tests), `1003a9edd4`); the new sentence "an unregistered file is read through a root-confined open so a link swapped after the check cannot make the read leave the project" is overstated by N-A (the root path is re-opened by name). Counts (17 REQ, 21 AC) agree with the E.3 matrix (14 release-blocking + 7 regression-guard = 21).
- progress.md E.4: caveat 6 (EvalSymlinks-to-read race on the unregistered read closed by `os.OpenRoot`; FIFO swap and registered read open), caveat 9 (error-text dependence) and caveat 10 (F-C) match the code. Caveat 6 shares the N-A overstatement. `sync_commit_sha: e99a990296dc8d84b48891a9d916b8379b9f4f3b` is consistent with the history (the spec.md `completed` transition is in that commit). The SPEC-body edit set is design.md D-1 by `0a33265120` only, as stated.
- The "Honest gaps" bullet about Go commands run without a slot check is present and accurate for the lane's earlier calls.

## Findings

- **N-A [SHOULD-FIX] [blocking for the required gate]** `internal/cli/migrate_ownership.go:428` (`readUnregisteredInsideRoot`). The Root is created with `os.OpenRoot(root)` from the root path string AFTER `EvalSymlinks`/`withinRootBoundary` has accepted the target. If the project root path (or an ancestor of it) is replaced by a link to an outside directory in that window, the Root pins the OUTSIDE directory and `Root.Open(rel)` reads the same relative name there. Reproduced: `AF2 attempts=2079 outside-reads=1`. Attacker model: a process that can rename entries in the parent directory of the project root while `moai update` runs (strictly narrower than F-A, which needed write access only inside the project). The bytes are hashed for provenance, never removed (removal disabled). Introduced by this card (the F-A repair is incomplete); at base there was no guard at all. Required fix: open the Root once on the verified project root when the environment is built (or at the start of the walk, and re-verify with `os.SameFile`/an fstat of the handle against the resolved root), and reuse that handle for every read; do not re-open the root by name after the check. Add a root-swap row to the race group. Alternatively disclose this window with the attacker model in E.4 and the CHANGELOG and soften the "so a link swapped after the check cannot make the read leave the project" sentence.
- **N-B [MINOR]** `internal/cli/update_prior_manifest.go:203`. `os.RemoveAll(sibling)` removes anything under the new sibling, including content a foreign process created there between `Mkdir` and the cleanup (reproduced with the marker-copy seam: `foreign-data-exists=false`). The lane's comment says the sibling "holds at most the marker file", which is not enforced. Introduced by this card (F-B repair; my round-3 recommendation was to remove the marker file and then the directory, which would not delete foreign content). Likelihood is very low (a foreign process must guess the new name within microseconds). Required fix: remove `sibling/.moai-run` (a single file) and then `os.Remove(sibling)`; if the directory is not empty, keep it and report the cleanup error.
- **N-C [MINOR, optional]** `readUnregisteredInsideRoot`: a link swapped in at the second window that is ABSOLUTE and points inside the project is kept as `symlink-escape` (T9), because `os.Root` rejects absolute links. Fail closed, race-only. No fix needed; mention in E.4 if N-A is documented.
- No other new defect found in `pathLeavesRoot`, `withinRootBoundary`, `IsPriorManifestBookkeeping`, `copyRunMarker`, `unreadableKind`, `userCounterpartConfirmed`, `writePriorManifest`: symlinked project root (T1), relative root (T2), directory where a file was expected (T3, `read-failed` as at base), odd names (T4), in-project ancestor and absolute links (T5, T6), no descriptor growth over 800 verdicts (T7). Windows could not be run (build only, rc=0); Windows separators in `IsPriorManifestBookkeeping` are handled by `filepath.ToSlash`.
- F-C (case-insensitive file system) remains a disclosed limit, not re-reported.

## Score (flat weighted model, must-pass firewall)

| Dimension | Score | Verdict | Basis |
|---|---|---|---|
| Functionality (40%) | 84/100 | PASS (by score) | F-A and F-B fixed and green, 7 anchored tests, 2 packages, builds and vet clean; no AC failure found |
| Security (25%) | 72/100 | FAIL (gate) | N-A reproduced by execution: a root-path swap reads outside the project; required codex backend failed on it |
| Craft (20%) | 80/100 | PASS | gofmt, vet, builds clean; race rows do not cover the root swap; cleanup broader than needed (N-B) |
| Consistency (15%) | 86/100 | PASS | style matches; documents match the code except the "link swapped after the check" overstatement |

Overall FAIL, driven by the required codex backend gate (unmet) and the reproduced N-A, not by a score collapse. Without the gate the code-level picture would be PASS-WITH-DEBT at best, with N-A as the item to fix first.

## Gaps (not observed)

- Not run by me: package-wide `internal/cli` tests, coverage, `-race`, golangci-lint, the 57-name regression set, the lane's five-run repetition and its mutants (the mutants' red/green claims rest on the lane's record).
- Windows: build only (`GOOS=windows go build ./...` rc=0); no execution, so `os.Root` behaviour on Windows (reserved names, junctions, short names) is unobserved.
- Case-insensitive file system behaviour (F-C) and a case-sensitive branch of `TestCaseVariantOnHostFilesystem` were not exercised.
- The 12 binding conditions were re-read, not re-measured cell by cell; the frozen call-site baseline row set was not re-counted.
- The root-swap race was observed once in a scratch loop (2079 attempts to the first hit); the rate is not characterised, and it ran on a loaded machine (load about 28).
- I did not verify the codex "4,782nd attempt" figure; my own run is the evidence for N-A. I reproduced N-B with a seam that simulates the foreign write, not with a second real process creating a `.moai-run` directory.
- The installed `moai` binary lags HEAD (`ee981a932` is an ancestor), so `audit-plan` and the `audit_multi` dispatch are served by older code.
- Slot status was read at my checks; the exact holder at the instant of each admission check between 17:30Z and 17:41Z was read at 17:30Z and 17:41Z only (lane-22, not landing or remeasure).
- No refused tool call occurred; no `.moai/state` file was edited or deleted by me.

## Residual-risk

- The defect class "a read of a project path ends up outside the project" has now been closed for the file component (F1, F-A) but each repair has left the next-outer component (symlinked ancestor, then root path string) to the next round; the registered-file read-before-boundary is still open at base (N16 family). A round 5 after N-A may find another edge in the newest helper.
- All open items fail closed or only hash bytes (removal is disabled: `removalEnabled` false), so no data-loss path was found; the exposure is a provenance read of an outside file by a local attacker who already controls the project root's parent directory.
- N-B can delete foreign data only inside a directory this run created moments earlier.

## Iteration history

- Round 1 (`dfd7c40fb2`): FAIL, codex fail, receipt `rcpt-7770ccadc2f78594b050e777`; F1 BLOCKING, F2/F3 SHOULD-FIX.
- Round 2 (`38c8ca5376`): FAIL, codex fail, receipt `rcpt-aeb48f1d567c915a06f7b90f`; F1-F3 fixed; N1 SHOULD-FIX, N2 MINOR.
- Round 3 (`1d9b06a737`): FAIL, codex fail, receipt `rcpt-665b06f8d214d9a97633b90c`; F-A, F-B SHOULD-FIX, F-C MINOR (tests not run by the auditor).
- Round 4 (this file, HEAD `aba63b1483`): FAIL, codex fail, receipt `rcpt-bc15a518bfe20c055b8f4bda`; F-A and F-B verified fixed with tests run by me; new N-A SHOULD-FIX (reproduced), N-B MINOR (reproduced), N-C MINOR optional.
- Path forward: fix N-A (one Root opened on the verified project root and reused, with a root-swap row) and N-B (remove the marker file, then a non-recursive directory remove), or disclose N-A with its attacker model and soften the two sentences; then a delta re-audit scoped to N-A and N-B. Under the factory's rule C this is the hold point.

AUDIT-VERDICT: FAIL spec=SPEC-UPDATE-OWNERSHIP-REPAIR-001 receipts=rcpt-bc15a518bfe20c055b8f4bda
