# t1204 — inherited GIT_DIR redirects init git-detection into the outer repo

Card: t1204 (no SPEC; bug fix with reproduction) · branch `WT-test-gitconfig-leak` · base `df526c9a9`

## Claim

1. `internal/cli/init_gitdetect.go` `gitRemoteListFunc` / `gitOriginURLFunc` (read shape) and
   `internal/cli/init_gitdetect_test.go` `gitDetectInitRepo` / `gitAddRemote` (write shape) ran
   `git -C <dir>` with the inherited environment; an inherited `GIT_DIR` outranks `-C`, so the
   helpers wrote `remote.upstream.*` into the outer repo and detection answered about it.
2. Setting `cmd.Env = gitenv.Env()` on all four call sites removes both shapes. No new helper;
   the existing `internal/gitenv` package is reused.
3. The regression test `TestDetectGitConfig_IgnoresInheritedGitDir` fails on the base code and
   passes after the fix.
4. (Follow-up to sync-audit F1) A fifth call site existed in the same file:
   `init_gitdetect_test.go:274` in `TestInitGitDetectionFillsConfig` ran
   `git -C projectDir init` unscrubbed. Under an inherited per-worktree `GIT_DIR`
   (`<repo>/.git/worktrees/<name>`) it rewrote the SHARED config to `core.bare=true`. It now calls
   the new fixture helper `gitInitAt`, which carries `cmd.Env = gitenv.Env()`; `gitDetectInitRepo`
   uses the same helper, so every fixture `git init` in the file goes through one scrubbed site.
   After this change no `exec.Command("git"…)` in `init_gitdetect_test.go` or
   `init_gitdetect_env_test.go` runs without `cmd.Env = gitenv.Env()`.
5. The new regression test `TestGitInitAt_IgnoresInheritedWorktreeGitDir` fails with the
   `gitInitAt` scrub removed and passes with it restored.
6. (Follow-up to sync-audit F2) `victimRemotes` now goes through `victimConfig`, which treats
   `git config` exit 1 as "no match" and fails the test on any other error.

## Evidence

RED — new test against unmodified production + helpers:

```
$ go test ./internal/cli/ -run 'TestDetectGitConfig_IgnoresInheritedGitDir' -count=1 -v
=== RUN   TestDetectGitConfig_IgnoresInheritedGitDir
    init_gitdetect_env_test.go:49: read shape: detectGitConfig(prebuilt) = ("manual", "github"), want (personal, gitlab) — answered about the GIT_DIR repo
    init_gitdetect_env_test.go:57: write shape: victim repo gained remotes:
        remote.upstream.url https://gitlab.com/group/proj.git
        remote.upstream.fetch +refs/heads/*:refs/remotes/upstream/*
--- FAIL: TestDetectGitConfig_IgnoresInheritedGitDir (0.74s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.551s
FAIL
```

GREEN — after the four `cmd.Env = gitenv.Env()` assignments (output filtered to RUN/PASS/FAIL lines):

```
$ go test ./internal/cli/ -run 'TestDetectGitConfig' -count=1 -v
=== RUN   TestDetectGitConfig_IgnoresInheritedGitDir
--- PASS: TestDetectGitConfig_IgnoresInheritedGitDir (0.74s)
=== RUN   TestDetectGitConfig
=== RUN   TestDetectGitConfig/non-git_directory_falls_back_to_manual
=== RUN   TestDetectGitConfig/git_repo_with_no_remote_is_manual
=== RUN   TestDetectGitConfig/github_https_origin_is_personal_+_github
=== RUN   TestDetectGitConfig/github_ssh_origin_is_personal_+_github
=== RUN   TestDetectGitConfig/gitlab.com_origin_is_personal_+_gitlab
=== RUN   TestDetectGitConfig/self-hosted_non-github_origin_is_personal_+_gitlab
=== RUN   TestDetectGitConfig/remote_present_but_not_named_origin_keeps_github_default
--- PASS: TestDetectGitConfig (1.69s)
=== RUN   TestDetectGitConfig_RemoteListError
--- PASS: TestDetectGitConfig_RemoteListError (0.00s)
=== RUN   TestDetectGitConfig_OriginURLError
--- PASS: TestDetectGitConfig_OriginURLError (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.188s

$ go test ./internal/gitenv/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/gitenv	0.254s

$ go vet ./internal/cli/ ; echo "vet_exit=$?"
vet_exit=0
$ gofmt -l internal/cli/init_gitdetect.go internal/cli/init_gitdetect_test.go internal/cli/init_gitdetect_env_test.go; echo "gofmt_exit=$?"
gofmt_exit=0
```

Lane repro against a fresh victim repo (scratchpad, `git init -q <scratch>/victim`):

```
$ GIT_DIR=<scratch>/victim/.git go test ./internal/cli/ -run 'TestDetectGitConfig$' -count=1; echo "exit=$?"
ok  	github.com/modu-ai/moai-adk/internal/cli	2.158s
exit=0
$ git config --file <scratch>/victim/.git/config --get-regexp '^remote\.'; echo "exit=$? (1 = no match)"
exit=1 (1 = no match)
```

### Follow-up (sync-audit F1/F2)

RED — `TestGitInitAt_IgnoresInheritedWorktreeGitDir` with the `cmd.Env = gitenv.Env()` line
temporarily removed from `gitInitAt` (the helper `:274` now calls), then restored:

```
$ go test ./internal/cli/ -run 'TestGitInitAt_IgnoresInheritedWorktreeGitDir' -count=1 -v
=== RUN   TestGitInitAt_IgnoresInheritedWorktreeGitDir
    init_gitdetect_env_test.go:104: shared config gained core.bare=true — gitInitAt re-initialized the inherited worktree gitdir
--- FAIL: TestGitInitAt_IgnoresInheritedWorktreeGitDir (0.54s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.284s
FAIL
exit=1
```

(A first draft of the test also set `GIT_WORK_TREE`; with it set the unscrubbed init did NOT flip
`core.bare` and the test passed on the mutant. The final test sets `GIT_DIR` only, which is the
shape a git hook exports.)

GREEN — scrub restored:

```
$ go test ./internal/cli/ -run 'TestDetectGitConfig' -count=1 -v   (--- / ok lines)
--- PASS: TestDetectGitConfig_IgnoresInheritedGitDir (0.72s)
--- PASS: TestDetectGitConfig (1.64s)
    --- PASS: TestDetectGitConfig/non-git_directory_falls_back_to_manual (0.02s)
    --- PASS: TestDetectGitConfig/git_repo_with_no_remote_is_manual (0.18s)
    --- PASS: TestDetectGitConfig/github_https_origin_is_personal_+_github (0.30s)
    --- PASS: TestDetectGitConfig/github_ssh_origin_is_personal_+_github (0.27s)
    --- PASS: TestDetectGitConfig/gitlab.com_origin_is_personal_+_gitlab (0.30s)
    --- PASS: TestDetectGitConfig/self-hosted_non-github_origin_is_personal_+_gitlab (0.29s)
    --- PASS: TestDetectGitConfig/remote_present_but_not_named_origin_keeps_github_default (0.28s)
--- PASS: TestDetectGitConfig_RemoteListError (0.00s)
--- PASS: TestDetectGitConfig_OriginURLError (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	3.116s
g1 exit=0

$ go test ./internal/cli/ -run 'TestInitGitDetectionFillsConfig' -count=1 -v
--- PASS: TestInitGitDetectionFillsConfig (0.43s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.020s
g2 exit=0

$ go test ./internal/cli/ -run 'TestGitInitAt_IgnoresInheritedWorktreeGitDir' -count=1 -v
=== RUN   TestGitInitAt_IgnoresInheritedWorktreeGitDir
--- PASS: TestGitInitAt_IgnoresInheritedWorktreeGitDir (0.57s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.163s
g3 exit=0

$ go vet ./internal/cli/; echo "vet_exit=$?"
vet_exit=0
$ gofmt -l internal/cli/init_gitdetect_test.go internal/cli/init_gitdetect_env_test.go; echo "gofmt_exit=$?"
gofmt_exit=0
```

Victim repro — fresh `git init -q <scratch>/victim1`, `<scratch>/victim2`, and a scratch repo with
a linked worktree (`<scratch>/linked/main` + `worktree add --detach <scratch>/linked/wt`):

```
$ GIT_DIR=<scratch>/victim1/.git go test ./internal/cli/ -run 'TestDetectGitConfig' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	3.171s
exit=0
$ GIT_DIR=<scratch>/victim2/.git go test ./internal/cli/ -run 'TestInitGitDetectionFillsConfig' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.097s
exit=0
$ git config --file <scratch>/victim{1,2}/.git/config --get-regexp '^remote\.'   (each)
victim1 remote_exit=1 (1 = no match)
victim2 remote_exit=1 (1 = no match)
$ git config --file <scratch>/victim{1,2}/.git/config --get core.bare             (each)
false   (victim1)
false   (victim2)

$ GIT_DIR=<scratch>/linked/main/.git/worktrees/wt go test ./internal/cli/ -run 'TestInitGitDetectionFillsConfig|TestDetectGitConfig' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	3.531s
exit=0
$ git config --file <scratch>/linked/main/.git/config --get core.bare
false
$ git config --file <scratch>/linked/main/.git/config --get-regexp '^remote\.'
remote_exit=1 (1 = no match)
$ git -C <scratch>/linked/main status --short
?? .moai/
```

The `?? .moai/` is a separate, production-side leak (see Residual-risk): isolating it,
`TestInitGitDetectionFillsConfig` alone under the same per-worktree `GIT_DIR` writes
`<scratch>/linked/main/.moai/db/main-af1cbc46/project.json` with
`"project_root": "<scratch>/linked/main"`. Producer path: `internal/homestate/paths.go:280`.

## Baseline-attribution

- RED measured in this worktree at HEAD `df526c9a9` with only the new test file added (production
  and helpers unmodified).
- GREEN and the lane repro measured in this worktree after the fix, before commit, same session.
- The lane's original repro result (3 subtests FAIL, exit status 3, victim gains
  `remote.upstream.url` + origin) is the lane's measurement, not re-run here on base code; the RED
  test above reproduces the same write shape (`remote.upstream.*`) independently. (The sync-audit
  E5 later re-ran it on base code and reproduced it.)
- Follow-up RED/GREEN, victim repros, and the file counts below were measured in this worktree on
  top of HEAD `1abdc8205` with the follow-up edits applied (uncommitted at measurement time), same
  session.

## Gaps

- The lane repro command itself was not re-run against base code in this session (positive
  control for that exact command); the RED test is the positive control for the mechanism.
- `GIT_WORK_TREE` alone (without `GIT_DIR`) was not tested separately.
- No `go test ./...` (per repo rule); CI on develop is the full-suite verdict.
- Windows/linux not measured locally.
- The primary checkout's `.git/config` was not touched or inspected (lead owns removal of its stray `upstream`).
- Whether the `.moai/db/.../project.json` leak (Residual-risk) also occurs on base `df526c9a9` was
  not measured; it is production code this card does not change.
- The 134/136 file counts below were not reproduced per file; they are text-pattern counts, not a
  defect count.

## Residual-risk

- Other test files still spawn git without referencing `gitenv`. Text-pattern counts, measured in
  this worktree:

  ```
  $ grep -rln 'exec.Command("git"' internal/cli --include='*_test.go' | wc -l
        53
  $ grep -rln 'exec.Command("git"' --include='*_test.go' internal pkg cmd | wc -l
       136
  $ grep -rln 'exec.Command("git"' --include='*_test.go' internal pkg cmd | xargs grep -L 'gitenv' | wc -l
       134
  ```

  The 134 are candidates, not confirmed defects: a file may isolate git another way (e.g. its own
  env), and the only sibling confirmed by reproduction is `:274`, fixed here. Any unscrubbed one
  can still contaminate an outer repo when the suite runs under an inherited `GIT_DIR`.
- Production `runInit` still leaks under a per-worktree `GIT_DIR`: `TestInitGitDetectionFillsConfig`
  writes `.moai/db/<key>/project.json` into the OUTER repo's work tree
  (`internal/homestate/paths.go:280`, observed above). Out of this card's two-file scope; reported
  for a follow-up card.
- `gitenv.Env()` snapshots `os.Environ()` at call time; a caller that later sets GIT_* on the
  returned slice would reintroduce the defect (not the case in the changed code).
