# t2124 progress — PR #1969 Codex P2: location memo fingerprint must track the config of an external gitdir (Class B)

Base: local main 04d8c27a7e. Branch WT-pr-codex-repair. Lane: lane-22. Finding: PR #1969 review comment on internal/homestate/location_memo.go:238 (commit 04d8c27a7e).

## Reproduction (RED, before the fix)

Test `internal/homestate/location_memo_gitdir_config_t2124_test.go` (TestLocationMemo_CoreWorktreeChangedInAnExternalGitDir): a repository whose `.git` is a file naming an external git directory (`git init --separate-git-dir`); after a first `CanonicalProjectRoot(wt)`, `git config core.worktree <other>` is run in it.
Observed on the unfixed tree, verbatim: `CanonicalProjectRoot = ".../001/wt" after core.worktree changed, a fresh lookup answers ".../001/other" (a stale remembered answer ...)` and `fingerprint unchanged after the external git directory's config changed`. Both assertions failed.

## Fix

`stampGitdirTarget` (location_memo.go) recorded only whether the named directory has HEAD and commondir; it now also stamps the directory's `config` (size and mtime, or absence) through the existing `stampFile`. One line plus a comment.

## Verification (this run, card worktree, uncommitted at the time of the run)

- `go test -count=1 -timeout 100s -run 'LocationMemo|CanonicalProjectRoot|ProjectKey|OutsideRepository' ./internal/homestate/` → ok (24.2s); the new test passes.
- `gofmt -l internal/homestate/` → empty; `go vet ./internal/homestate/` → no output.
- `go test -count=1 -run TestRebindRollsBackWhenAWriteIsRefused ./internal/homestate/` → ok (9.3s).

## Gaps

- The full `go test -count=1 ./internal/homestate/` did NOT complete: it hit the default 10-minute limit with a timeout panic while TestRebindRollsBackWhenAWriteIsRefused (factory_run_rebind_faults_test.go:97) was running (FAIL 600.5s). That test passes alone in 9.3s and does not touch the location memo; the cause of the slow whole-package run is not established (it was not run on the base tree either, so it is not claimed as pre-existing). No whole-package PASS is claimed.
- The symlink-ancestor finding on the same file was card t2060 and is already in the tree; the other #1969 findings go to t2129.
- No race detector, no other OS, no golangci-lint. No codex card-review yet.
