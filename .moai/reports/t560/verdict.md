# t560 — hook-layer git-environment leak sweep (GH #1691 follow-up to t516)

Card t560. Worktree `.claude/worktrees/t560`, branch `WT-hook-env-scrub`, base `3ac58b5a1`
(= `origin/develop` at dispatch). Re-scoped from its original text by operator decision after the
pre-dispatch premise check below.

---

## Claim

1. The primary defect of GH #1691 was ALREADY repaired and landed by card t516; t560's original
   text would have rebuilt an existing fix.
2. GH #1691 defect 1 (`core.bare` flips to `true`) **does reproduce**, contrary to t516's record of
   non-reproduction, and only under one specific shape.
3. `internal/hook/quality/gate.go:1065` was a **sibling gap** of the t516 repair — same file, same
   process, same defect, left unscrubbed.
4. Six call sites now scrub the git repository-location environment; five sites in
   `worktree_base_branch.go` deliberately do NOT and are pinned against a future sweep; one site is
   defect-shaped but unreachable and is reported rather than repaired.
5. The touched packages and every consumer package pass.

---

## Evidence

### 1 — t516 already landed the primary repair

```
$ git ls-tree -r --name-only origin/develop -- internal/hook/quality/ | grep git_env
internal/hook/quality/gate_step_git_env_test.go
internal/hook/quality/step_git_env.go

$ git grep -n 'stepEnv()' origin/develop -- '*.go'
origin/develop:internal/hook/quality/gate.go:1169:	cmd.Env = stepEnv()
origin/develop:internal/hook/quality/step_git_env.go:72:func stepEnv() []string {
```

Wired, not a vacuous helper. GH #1691's own comment thread records the operator answering
"Queued as card **t516**", so t560 was a second card on the same issue.

### 2 — defect 1 reproduces; the trigger is narrower than reported

Isolated throwaway repositories under `/tmp`, git 2.50.1 (Apple Git-155), darwin:

```
core.bare BEFORE: false
core.bare AFTER : true
host status rc  : 128        # fatal: this operation must be run in a work tree
```

Per-variable discriminant, same git:

| Inherited variable | host `core.bare` after a child `git init` |
|---|---|
| `GIT_DIR=<host>/.git/worktrees/<wt>` | **true** |
| `GIT_DIR=<host>/.git` | false |
| `GIT_INDEX_FILE=<host>/.git/worktrees/<wt>/index` | false |

Only the per-worktree gitdir fires. `bare = true` lands in `<host>/.git/config`, which every linked
worktree shares — one fixture breaks the whole repository.

Reproduced a second time **through production code**, by mutating away the t516 scrub at
`gate.go:1139`:

```
--- FAIL: TestRunStep_ChildGitInitFromWorktreeDoesNotFlipOuterCoreBare
    a gate step's child flipped the OUTER repository's core.bare: "false" before, "true" after.
    Reinitialized existing Git repository in .../host/.git/worktrees/host-wt/
```

Mutant fully reverted; `git diff -- internal/hook/quality/gate.go` showed the scrub restored, and
the test returned to PASS.

### 3 — the sibling gap, RED before the fix

```
--- FAIL: TestStagedFiles_ReadsGivenRepoNotLeakedGitDir
    stagedFiles(.../target) answered about the LEAKED repository:
    got "host-staged.txt", want "target-staged.txt"
--- FAIL: TestStagedFiles_EmptyGivenRepoIsNotFilledFromLeakedGitDir
    stagedFiles(.../target) reported [host-staged.txt] from the LEAKED repository
```

### 4 — the sweep

All five remaining sites observed RED before their fix. Three were observed directly (unfixed at the
time of the run); two were observed by mutation, because the fix had already been applied to them:

```
--- FAIL: TestGetModifiedGoFilesReadsGivenRepoNotLeakedGitDir
    got [.../target/host.go], want [.../target/target.go]
--- FAIL: TestResolveWorktreeRepoRootReadsGivenRepoNotLeakedGitDir
    got ".../target/internal/deep", want ".../target"
--- FAIL: TestRunGitReadsGivenRootNotLeakedGitDir
    got "host-seed", want "target-seed"
--- FAIL: TestCommitChangedFilesReadsGivenRootNotLeakedGitDir
    got [], want [target.txt]
--- FAIL: TestGitOutputReadsGivenRepoNotLeakedGitDir            (under mutation)
    got "host-seed", want "target-seed"
--- FAIL: TestChangedAtForProjectReadsGivenRepoNotLeakedGitDir  (under mutation)
    got "2001-01-01T00:00:00Z", want "2002-02-02T00:00:00Z"
```

After the fixes, all six PASS.

### 5 — the exception, pinned and shown to fail

`worktree_base_branch.go` is UNCHANGED. Adding a scrub to
`worktreeBaseBranchInPrimaryCheckoutReal` turns its guard red:

```
--- FAIL: TestWorktreeBaseBranchInPrimaryCheckout_ReadsAmbientGitDir
    the discriminant stopped reading the ambient git context:
    GIT_DIR=.../repo/.git and GIT_DIR=.../repo/.git/worktrees/wt both reported primary=false.
```

Mutant reverted; `git diff --stat -- internal/hook/worktree_base_branch.go` is empty.

### 6 — verification

```
go build ./...                                        # clean
gofmt -l internal/                                    # empty
go vet ./internal/hook/... ./internal/gitenv/...       # clean
go test ./internal/hook/... ./internal/gitenv/...      # all ok
go test ./internal/cli/                               # ok 490.380s
go test ./internal/codexadapter/... ./internal/codexwiring/... ./internal/feedback/... \
        ./internal/permission/... ./internal/web/... ./internal/migration/...  # all ok
```

### 7 — host untouched (binding constraint)

```
$ git config core.bare
false
$ grep -c 'bare = true' /Users/goos/MoAI/moai-adk-go/.git/config
0
```

Every `core.bare` reproduction ran in `t.TempDir()` or a `/tmp` throwaway repository.

---

## Baseline-attribution

Every figure above was measured in this run, in this worktree, at base `3ac58b5a1`, on git 2.50.1
(Apple Git-155), darwin. Nothing is carried over from t516's run or from the reporter's.

The classification is a closed set over the 20 real `exec.Command` call expressions in
`internal/hook` (raw list: `.moai/reports/t560/exec-sites-raw.txt`), 1+6+5+1+7 = 20:

| Class | Count | Disposition |
|---|---|---|
| Already repaired by t516 | 1 | `gate.go:1139`, untouched |
| Scrubbed by this card | 6 | `gate.go:1065`, `agentmemory.go:140`, `navigator_detect.go:500`, `session_end.go:255`, `security/guardian.go:246`, `worktree_create.go:145` |
| Deliberately NOT scrubbed | 5 | `worktree_base_branch.go:145,146,162,171,189` — ambient context IS the answer |
| Defect-shaped, unreachable | 1 | `quality/tool_registry.go:143` — see Gaps |
| Out of scope (non-git child) | 7 | tmux ×5, `sg` ×2 |

---

## Gaps — what was NOT observed

- **Reachability of five of the six scrubbed sites is not established.** Only `gate.go:1065` is
  confirmed reachable from a git-hook environment (the pre-commit hook shells out to `moai gate`
  and `stagedFiles` runs in that process). The other five have Claude Code as their parent, which
  does not export `GIT_DIR`. Their fix closes a leak surface; it is **not** a repair of an observed
  incident, and must not be reported as one.
- **`quality/tool_registry.go:143` (`RunTool`) was not repaired.** It is defect-shaped — arbitrary
  child, `cmd.Dir` set, `cmd.Env` nil — but its only callers are `Formatter` and `Linter`, and
  those have zero non-test constructors outside the package (control: `runStep` 8 hits vs
  `quality.NewFormatter|NewLinter` outside the package 0 hits). Repairing dead code produces a
  change no test can verify. Whether that triangle should be deleted is a separate card.
- **The reporter's own environment was not reproduced.** Linux / WSL2, moai-adk 3.1.2. All
  measurements here are darwin / git 2.50.1.
- **`GIT_CONFIG_COUNT` / `KEY_n` / `VALUE_n` injection is untouched.** t516 left it deliberately
  (behaviour variables are kept by design) and it can in principle set `core.bare`. Still open.
- **`exec.Command` outside `internal/hook` was not swept.** The whole tree carries 447 occurrences
  across 86 non-test files; this card scoped to the hook layer, where a git-hook parent is possible.
- **Attribution split.** The A/B/C discriminant and the `core.bare` flip are this lane's
  measurements. The lead did not reproduce them and recorded them under that attribution.

---

## Residual-risk

- **The scrub list is a denylist.** A future git variable that scopes a repository and is not in
  `gitenv.RepoScopingVars` leaks silently. An allowlist would fail closed, but would strip the
  identity and behaviour variables the boundary deliberately keeps.
- **Keeping `GIT_CONFIG_*` is a considered risk, not an oversight.** Those can set arbitrary config
  in a child, `core.bare` included. Dropping them would change what the gate grades.
- **A behavioural change reaches every hook consumer.** Children no longer inherit repository
  location. A consumer that depended on that inheritance would now resolve differently; no such
  consumer was found and all consumer packages pass, but the sweep was by test, not by proof.
- **`internal/gitenv` is now a shared seam.** Its scrub list has six consumers, so a future edit to
  that list changes all of them at once. That is the point — it is what prevented the sibling gap
  from recurring — but it also means the list deserves the review a shared contract gets.
- **The unreachable `RunTool` may become reachable.** If `Formatter`/`Linter` are ever wired into
  the gate, the defect arrives with them, and no test currently guards that site.

---

## Integration windows (2026-09-11, lane-8)

The sections above describe the card at base `3ac58b5a1`. The integration record lives in three
files and is summarised here only as pointers:

- `window/summary.md` — first window. Absorbed local develop `84e5666d9` (merge `2c07f89ff`,
  CHANGELOG union). gitenv, hook/quality, hook/security ok; hook copy parity 2/2; internal/hook
  failed `TestSessionStart_DeferredScanDoesNotBlockReturn` (559ms > 500ms) on a handler and test
  blob-identical to develop. Not merged.
- `rerun/summary.md` — one home-isolated internal/hook run. The target test did not fail; two GLM
  credential tests failed because the prescribed `MOAI_HOME` override hid their `HOME`-seeded env
  file. Lead verdict (ii): merge without a further run; the two GLM failures are excluded as
  isolation by-products.
- `window2/summary.md` — second window. Absorbed local develop `526249cf6` (merge `7a76becca`, no
  conflict). The delta touched no file under `internal/gitenv` or `internal/hook` (only reports and
  three `internal/cli` test files), so tests were not re-run, by lead instruction.
