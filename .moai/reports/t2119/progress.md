# t2119 progress (#1969 Windows race residuals)

## Claim
This card's first slice fixes the largest single cause that is deterministic from the code: the claim store's directory fsync after a rename failed the write on Windows ("claim store unavailable: sync …: Access is denied", 326 failures in the CI analysis). It is now best effort on Windows only (`access denied` = `fs.ErrPermission`), and stays a hard failure on every other platform and for every other error. The remaining causes are proposed as separate cards below (no Windows runner is available to this lane, so each needs its own reproduction).

## Evidence
- Source of the failure list: `.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md` (t2077 tree), row "Windows: 디렉터리 fsync 거부(326건)", error text `agenticfactory: claim store unavailable: sync C:\…\docs\checkpoint: Access is denied.`
- Code: `internal/agenticfactory/claim/durable.go` `writeAtomic` called `s.sync.syncDir(dir)` and failed the write on any error; `internal/agenticfactory/ledger/snapshot.go` already treats the same Windows refusal as best effort (its comment: "Windows cannot open a directory to sync it").
- RED: `go test ./internal/agenticfactory/claim -run 'TestDirSyncFailureIsFatal|TestSyncDirDefaultPath' -count=1` -> `undefined: dirSyncFailureIsFatal` (build failed).
- GREEN: `go test ./internal/agenticfactory/claim -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/agenticfactory/claim  137.631s`.
- Cross-compile: `GOOS=windows go vet ./internal/agenticfactory/claim` printed nothing. `gofmt -l internal/agenticfactory/claim` printed nothing.

## Baseline-attribution
Tree t2119, base = local main at lease time; commands ran in this run against this tree.

## Gaps
- The Windows behavior itself was NOT run: no Windows host here. The new test pins the decision function with explicit `goos` values and the real default path on this platform; the CI Windows race matrix is the verdict (Residual-risk).
- Other agenticfactory packages named in the analysis (runner 135, e2e 80, identity, monitor/adapt, intake, checkpoint, ctxmodel) were not individually inspected; they are expected to go through the same `writeAtomic` (they carry the `claim store unavailable` text) but that was not traced.
- Only the claim package was run, not the whole suite.
- A worktree-guard refusal of a compound command occurred earlier in the session; commands were re-issued plainly.

## Proposed split (not done in this slice)
1. `/bin/sh` absent on Windows (19): legitimate Windows skip with a reason or `sh` lookup; 
2. POSIX permission bits (46): Windows skip with a reason, or assert on the portable subset;
3. home guard sees Windows TEMP under the real home (54): environment-dependent guard, needs a Windows reproduction;
4. pre-commit e2e (12) and the ~200 "other" path-spelling failures: classify first;
5. epoch-0 refusals (26): likely a follow-on of the fsync failure — recheck after this slice's CI;
6. SQLite URI `invalid uri authority: C:` (474) belongs to t2118 (lane-28), deliberately not touched.

## Residual-risk
If Windows CI still reports the fsync failure after this change, the error is not `fs.ErrPermission` on that platform (for example a different errno) and the tolerance must widen. Tolerating the Windows refusal gives up the directory-fsync durability step there only; the file fsync and the rename remain.
