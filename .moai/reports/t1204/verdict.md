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

## Baseline-attribution

- RED measured in this worktree at HEAD `df526c9a9` with only the new test file added (production
  and helpers unmodified).
- GREEN and the lane repro measured in this worktree after the fix, before commit, same session.
- The lane's original repro result (3 subtests FAIL, exit status 3, victim gains
  `remote.upstream.url` + origin) is the lane's measurement, not re-run here on base code; the RED
  test above reproduces the same write shape (`remote.upstream.*`) independently.

## Gaps

- The lane repro command itself was not re-run against base code in this session (positive
  control for that exact command); the RED test is the positive control for the mechanism.
- `GIT_WORK_TREE` alone (without `GIT_DIR`) was not tested separately.
- No `go test ./...` (per repo rule); CI on develop is the full-suite verdict.
- Windows/linux not measured locally.
- The primary checkout's `.git/config` was not touched or inspected (lead owns removal of its stray `upstream`).

## Residual-risk

- ~32 other test files still run `git -C <tempdir>` unscrubbed (listed in the completion report
  as candidates from a text-pattern grep; not verified by reproduction). Any of them can still
  contaminate an outer repo when the suite runs under an inherited `GIT_DIR`.
- `gitenv.Env()` snapshots `os.Environ()` at call time; a caller that later sets GIT_* on the
  returned slice would reintroduce the defect (not the case in the changed code).
