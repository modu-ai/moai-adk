# t148 — Launcher-stamped session PID (`MOAI_SESSION_PID`)

Card class B (plan skipped; run → sync in lane). Follow-up to t144
(`eeecbe363`, merged as `b317f47c4`), which made the coordination registry
record a live session PID by walking the process ancestry from the hook
subprocess. t148 removes the walk from the launcher path by having the launcher
declare the PID outright.

- Branch: `WT-t148`
- Base: `b317f47c4` (`origin/release/v3.1.1` head — NOT `origin/main`, which
  carries neither `EnvMoaiSessionPID` nor `internal/session/session_pid.go`)
- Evidence: this file

## 1. Claim

1. On POSIX, the launcher stamps `MOAI_SESSION_PID` into the environment it
   `execve(2)`s into, and the stamped value names the process that becomes the
   session — so `resolveSessionPID` resolves at step 1 (env override) with no
   ancestry walk.
2. An inherited stamp from an outer launch is replaced, never duplicated.
3. The Windows path deliberately does NOT stamp, and no hook does either.
4. Nothing else in the launch chain is disturbed.

## 2. Evidence

### Change surface (4 files, 2 modified + 2 new; plus 3 test files)

| File | Change |
|---|---|
| `internal/cli/launch_session_pid.go` | NEW — `withSessionPID(env, pid)`: strips any inherited `MOAI_SESSION_PID=` entry, appends the new one; non-positive pid is a no-op |
| `internal/cli/launch_exec_posix.go` | `syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))` |
| `internal/cli/launch_exec_windows.go` | comment only — records WHY it does not stamp (the session is the spawned child; its PID does not exist before its env is fixed) |
| `internal/cli/launch_session_pid_test.go` | NEW — unit + structural guards |
| `internal/cli/launch_session_pid_exec_posix_test.go` | NEW — exec round-trip (POSIX-only build tag) |
| `internal/session/session_pid_launcher_test.go` | NEW — launcher-path registry cases |

Entry point measured before editing: `internal/cli/launch_exec_posix.go:16`
`syscall.Exec(claudeBin, args, env)` — the single POSIX exec site.
`grep -rn "execOrSpawnClaude" internal/ | grep -v _test.go` reports exactly one
non-definition caller, `internal/cli/launcher.go:798`, the shared path all of
`cc` / `glm` / `cg` / `--spawn` funnel through. `grep -rn "syscall.Exec"`
confirms the only other exec site is `internal/cli/update.go:775` (binary
self-reexec, unrelated).

The stamp is applied to the `env []string` slice only — never via `os.Setenv` —
so the `--continue` fallback path at `launcher.go:762` (`exec.Command`, which
forks rather than replaces) does not inherit a PID that would name its parent.

### Commands run and observed output

```
$ go build ./... && go vet ./internal/cli/... ./internal/session/...
(no output — rc 0)

$ GOOS=linux   go vet ./internal/cli/... ./internal/session/...   → rc=0
$ GOOS=windows go vet ./internal/cli/... ./internal/session/...   → rc=0

$ go test ./internal/session/ -count=1 -timeout 300s
ok  	github.com/modu-ai/moai-adk/internal/session	5.133s

$ go test ./internal/cli/ -count=1 -timeout 540s
ok  	github.com/modu-ai/moai-adk/internal/cli	225.293s

$ golangci-lint run internal/cli/... internal/session/...
0 issues.
```

New tests, verbose:

```
--- PASS: TestExecOrSpawnClaude_PosixBuildTagGate      (pre-existing, still green)
--- PASS: TestWithSessionPID_StampsAndReplaces
      /appends_to_a_clean_environment
      /replaces_an_inherited_value
      /non-positive_pid_is_a_no-op
--- PASS: TestWithSessionPID_ParsesBackAsSelf
--- PASS: TestSessionPIDStamp_PlatformBoundary
--- PASS: TestSessionPIDStamp_NotSetFromHooks
--- PASS: TestExecOrSpawnClaude_StampsLiveSessionPID
--- PASS: TestRegister_RecordsLivePID_LauncherPath      (internal/session)
--- PASS: TestResolveSessionPID_IgnoresDeadLauncherStamp (internal/session)
```

### The load-bearing assertion (claim 1)

`TestExecOrSpawnClaude_StampsLiveSessionPID` is behavioral, not structural. It
spawns exactly one bounded child (30 s context, `-test.run` pinned to the
helper, guarded by `MOAI_TEST_LAUNCH_EXEC_HELPER=1` so a plain run execs
nothing). The child calls the production `execOrSpawnClaude` with `/bin/sh`
standing in for the claude binary; the shell prints `$MOAI_SESSION_PID` and
`$$`. Because `execve(2)` replaces rather than forks, the shell IS the process
that was stamped, so the two numbers must agree — and they do. That is the
card's property (`stamped PID == session PID`, no ancestry consulted) proven on
the real code path, with the registry half covered separately by
`TestRegister_RecordsLivePID_LauncherPath`.

### Falsification (the guards actually bite)

Temporarily neutering the stamp to `withSessionPID(env, 0*os.Getpid())`:

```
--- FAIL: TestExecOrSpawnClaude_StampsLiveSessionPID (0.03s)
--- FAIL: TestSessionPIDStamp_PlatformBoundary (0.00s)
    launch_session_pid_test.go:81: launch_exec_posix.go must stamp the session PID
    via withSessionPID(env, os.Getpid()); after execve(2) this process IS the session
```

The file was restored immediately (`grep -n withSessionPID
internal/cli/launch_exec_posix.go` → line 26, the real form) and the suites
above were re-run green at the final tree.

### The [HARD] hook prohibition

`TestSessionPIDStamp_NotSetFromHooks` walks every `.go` file under
`internal/hook/` and fails on any mention of `MOAI_SESSION_PID` /
`EnvMoaiSessionPID`. A hook subprocess's own PID is dead within milliseconds of
being recorded — the exact defect t144 fixed — so the rule is now mechanical,
not a comment.

## 3. Baseline attribution

Every number above was measured in this worktree
(`.claude/worktrees/t148`, branch `WT-t148`) against the final tree at base
`b317f47c4`. `git log --oneline -1` → `b317f47c4`; `grep -c EnvMoaiSessionPID
internal/config/envkeys.go` → `2` (the t144 constant is present, confirming the
release base rather than `main`).

Base correction on record: `EnterWorktree` created the tree from
`origin/main` (`4100d8767`), which lacks t144. The branch was moved to
`origin/release/v3.1.1` before any commit (`git reset --mixed`, working-tree
edits preserved); `git diff HEAD origin/release/v3.1.1` over everything outside
the two touched packages reported no difference, so the base is the release
head exactly, with no merge commit carried in.

## 4. Gaps (explicitly NOT observed)

- **No live `moai cc` session was launched.** The card's phrasing ("the
  registered PID of a launcher-started session") was verified as its two
  component properties — exec-time stamp equals the running PID
  (`TestExecOrSpawnClaude_StampsLiveSessionPID`), and the registry records the
  stamp without walking (`TestRegister_RecordsLivePID_LauncherPath`) — rather
  than end-to-end. Launching a real session would replace this session's own
  process and, per CLAUDE.local.md §13, `moai cc` / `moai glm` flows mutate real
  settings files and must not be run in the dev project.
- **`kill -0 rc=0` on a real registered PID was not run**; the liveness half is
  asserted through the package's `pidIsAlive` seam in the registry tests.
- **Windows runtime behavior was not executed** — only `GOOS=windows go vet`
  (which does compile the `//go:build windows` file). No Windows host available.
- **Full suite not run locally** (CLAUDE.local.md §4). Only the two affected
  packages were run; the cross-package verdict is CI's.
- `internal/kanban`'s `workers.json` PID stamp was read for context but not
  re-verified; this card did not touch it.

## 5. Residual risk

- **The `--continue` fallback path is unstamped.** `launcher.go:762` runs claude
  as a forked child, so it has no PID to declare and falls back to the ancestry
  walk exactly as before. Correct, but it means the deterministic path covers
  fresh launches only — a resumed session still depends on t144's walk.
- **Windows keeps the walk**, and `internal/session/proc_info_other.go` reports
  no ancestry there, so Windows still lands on the `os.Getpid()` fallback. That
  is unchanged by this card, and is a known residual limit of t144 rather than a
  regression.
- **A stamp inherited across an unusual chain** (a launcher that execs something
  which later forks a second claude) would name the wrong process. Mitigated,
  not eliminated: the resolver rejects a stamp whose PID is not alive
  (`TestResolveSessionPID_IgnoresDeadLauncherStamp`), but a recycled-PID
  collision would slip through. Same exposure the env override already carried
  before this card.
- The exec round-trip test depends on `/bin/sh`; it skips (does not fail) where
  that is absent.

## 6. Verdict

PASS for the stated scope. Not pushed — branch `WT-t148`, base `b317f47c4`.
