# t1864 — errcheck: three unchecked fmt.Fprintf results in factory_run_adopt.go (Class A, run)

Card: t1864 (run tn1p4e, lane-31, P1, main-red express). Tree `.moai/worktrees/t1864`, branch `WT-run-adopt-errcheck`, base `bc5817d75` (main after the t1858 landing).

## Claim
`internal/cli/factory_run_adopt.go` lines 79, 91 and 103 call `fmt.Fprintf(warn, ...)` without handling the return values, so `golangci-lint run --enable-only errcheck ./internal/cli/` exits 1 with 3 issues. Source: t1707's `adoptResumedLeaderRun` warnings. CI Lint runs only on origin/main, so the red has not surfaced yet.

## Evidence (this run, this tree `bc5817d75`)
```
golangci-lint --version  -> golangci-lint has version v2.1.6 built with go1.26.8
golangci-lint run --enable-only errcheck ./internal/cli/   (exit=1)
internal/cli/factory_run_adopt.go:79:15: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:91:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
internal/cli/factory_run_adopt.go:103:14: Error return value of `fmt.Fprintf` is not checked (errcheck)
3 issues:
* errcheck: 3
```
The file's and package's convention for a best-effort write to a warning/notes writer is `_, _ = fmt.Fprintf(w, ...)` (for example `internal/cli/factory.go:916`, `internal/cli/factory_adopt.go:95-100`). The warnings go to a launch-time notes writer; a failed write has nowhere to be reported, so ignoring the result is deliberate and matches those sites.

## Repair design
Only the three call sites change: `fmt.Fprintf(warn, ...)` -> `_, _ = fmt.Fprintf(warn, ...)`. No lint configuration change, no `//nolint`.

## Gaps
- The first attempt to rename the branch in this session was refused by the branch guard before any change: `BRANCH_GUARD_VIOLATION: git branch in primary checkout (use a worktree; ...)`. The session's working directory was still the primary checkout at that moment (the `cd <tree> && git branch -m` form is read against the session directory). It was not routed around: once the session's working directory was this worktree the same rename ran there and succeeded.
- No code change yet at the time of writing; verification and the re-run of the same command are recorded below.
