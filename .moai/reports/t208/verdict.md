# t208 — profile test isolation (`TestProfileRename/ok` under `-count>1`)

Card class: B (defect, cause unknown at dispatch). Branch `WT-profile-test-isolation`.

## Claim

1. `TestProfileRename/ok` failed 9 times in 10 under `-count=10` because the profile
   base directory is installed once in `TestMain` and therefore shared by every
   `-count` iteration. Fixed by narrowing the base to a per-test `t.TempDir()`.
2. The `$TMPDIR` accumulation of `moai-cli-profiles-*` directories is a **different
   root** in a **different package** (`internal/cli`, not `internal/web`), caused by
   helper subprocesses that call `os.Exit` and so never reach `TestMain`'s cleanup.
   Not fixed here — reported to the lead (t213 overlap).

## Evidence

### The failure, before the fix

```
$ go test ./internal/web/ -run 'TestProfileRename' -count=10
--- FAIL: TestProfileRename (0.31s)
    --- FAIL: TestProfileRename/ok (0.07s)
        profile_bar_test.go:197: rename status = 400, want 200
...
FAIL	github.com/modu-ai/moai-adk/internal/web	3.417s
FAIL
```

### The mechanism

`internal/web/main_test.go` `TestMain` calls `sandboxProfileBaseDir()`, which
`os.MkdirTemp`s one directory and points `profile.BaseDirOverride` at it for the
whole binary. `TestMain` runs once per binary; `-count=N` runs the suite N times
*inside that binary*. So all N iterations share one base.

`TestProfileRename/ok` seeds `scratch` and renames it to `scratch2`, asserting 200.
Iteration 1 leaves `scratch2` behind. Iteration 2 seeds `scratch`, the handler finds
`scratch2` already present, and refuses with 400 — hence 1 pass and 9 failures, and a
green `-count=1`.

### The fix

`internal/web/profile_bar_test.go`: new `sandboxProfileBaseForTest(t)` sets
`profile.BaseDirOverride` to `t.TempDir()` with `t.Cleanup` restore, installed on the
`TestProfileRename` parent (its subtests are serial; `BaseDirOverride` is a package
global). The package-wide `TestMain` sandbox is unchanged — it protects `$HOME`; this
one gives each iteration a clean slate. The same per-test override pattern already
exists in this package (`profile_crud_test.go`, `profile_traversal_test.go`,
`integration_test.go`).

### After the fix

```
$ go test ./internal/web/ -run 'TestProfileRename' -count=10
ok  	github.com/modu-ai/moai-adk/internal/web	3.801s

$ go test ./internal/web/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/web	12.249s

$ go vet ./internal/web/
(no output)
```

### The `$TMPDIR` leak is a separate root

Counts in `$TMPDIR` (`/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/`):

| prefix | owning package | count |
|---|---|---|
| `moai-cli-profiles-*` | `internal/cli` | 11,960 |
| `moai-web-profiles-*` | `internal/web` | 1 |

`TestProfileRename` lives in `internal/web`, which is not leaking. The leaking helper
is `internal/cli/main_test.go:152`, and the mechanism is not a missing cleanup — the
cleanup is there (`main_test.go:163`, `os.RemoveAll` before `os.Exit`). It is bypassed
by the package's helper-subprocess pattern: `internal/cli` re-execs its own test
binary (`exec.Command(os.Args[0], "-test.run=…")` at `todo_test.go:150,333,662`,
`exitcode_guard_test.go:31`, `launch_session_pid_exec_posix_test.go:55`), each
subprocess runs `TestMain` and creates a base directory, then the helper test body
calls `os.Exit(N)` (`exitcode_guard_test.go:47`, `todo_test.go:395..441`) — which
terminates the process before `m.Run()` returns, so `restoreProfileBaseDir()` never
runs.

Measured directly:

```
before:  11960
$ go test ./internal/cli/ -run 'TestExitCodeGuard|TestHelperProcess' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.631s
after:   11962      # +2 from one invocation
```

Running the full `internal/web` suite twice moved the `moai-web-profiles-*` count not
at all (1 → 1), confirming the two symptoms do not share a root.

## Baseline-attribution

Every figure above was measured in this run, in this tree
(`.claude/worktrees/t208`, branch `WT-profile-test-isolation`, base `cd0cee1b8`), with
the command shown immediately above it.

## Gaps

- No fix for the `internal/cli` leak — out of this card's root, and t213 was warned
  about the overlap. The 11,962 stale directories were left in place rather than
  deleted, so the next owner can measure the same baseline.
- Only `internal/web` was tested. `internal/cli` and `internal/profile` carry
  structurally identical `TestMain` sandboxes and may hold the same
  `-count>1` cross-iteration hazard in their own tests; not surveyed here.
- The full suite was not run locally (repo rule); CI on the PR is the full-suite
  verdict.

## Residual-risk

- The fix assumes `TestProfileRename`'s subtests stay serial. If one later calls
  `t.Parallel()`, a parent-level write to the package-global `BaseDirOverride` becomes
  a race — the same global that made the package-wide sandbox necessary.
- `-count>1` is not exercised by CI, so a future test in this package can reintroduce
  the same cross-iteration coupling without any signal.
