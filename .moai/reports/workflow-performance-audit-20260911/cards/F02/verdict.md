# F02 verdict

## Claim

F02 is fixed: the verification snapshot key now changes when an already-present,
non-ignored untracked file changes content. Ignored files remain excluded by the
Git file-set query.

## Evidence

Implementation changed `internal/verify/key.go` to enumerate
`git ls-files --others --exclude-standard -z`, read each listed file, and bind
its path and bytes into the digest. A disappearing file returns an error so the
caller re-executes instead of caching a partial key. The regression test in
`internal/verify/key_test.go` rewrites one untracked file and asserts distinct
keys.

Command:

```text
go test ./internal/verify -run '^TestSnapshotKeyUntrackedContentChanges$' -count=1 -v
=== RUN   TestSnapshotKeyUntrackedContentChanges
=== PAUSE TestSnapshotKeyUntrackedContentChanges
=== CONT  TestSnapshotKeyUntrackedContentChanges
--- PASS: TestSnapshotKeyUntrackedContentChanges (0.71s)
PASS
ok  github.com/modu-ai/moai-adk/internal/verify 0.943s
```

Command:

```text
go test ./internal/verify -count=1
ok  github.com/modu-ai/moai-adk/internal/verify 2.567s
```

## Baseline-attribution

Both commands ran in the F02 worktree `WT-workflow-audit-f02`, at the local
`develop` baseline `4c99d973e` before its card commit.

## Gaps

The full repository test suite and cross-platform matrix were not run for this
card; the change is scoped to `internal/verify`.

## Residual-risk

Key computation reads untracked files after listing them. A concurrent create,
delete, or rewrite can still cause an error or capture a later byte sequence;
the caller's documented fail-open re-execution path must remain in force.
