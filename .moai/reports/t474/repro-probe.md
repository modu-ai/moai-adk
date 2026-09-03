# Card t474 — Reproduction Probe Record

Card: t474 · branch `WT-git-status-fixture` @ `25a3212a9` (origin/develop tip) · tree: this worktree, no local commits yet.
All measurements below are this run, this worktree, unless noted.

## Claim

The two CI-red tests (`TestStatusAheadBehindFromHeader`, `TestStatusBranchHeaderShapes`, both `internal/core/git`) are GREEN locally because this machine's system gitconfig pins `init.defaultBranch=main`; on CI (no such config) the fixture's bare remote `git init -q --bare` (no `-b`, `status_branch_test.go:99` and `:235`) leaves the remote's HEAD symref at git's compiled default (`master`), which no pushed ref satisfies, so every `git clone` of it degrades to an unborn-HEAD clone — and that single mechanism produces all observed CI failure signatures. One root, two tests.

## Evidence — chain links, each measured

**L0. Local green (selector sweep confirmed, not vacuous):**
```
$ go test ./internal/core/git/ -run 'TestStatusAheadBehind|TestStatusBranchHeaderShapes' -count=1 -v
=== RUN   TestStatusBranchHeaderShapes
--- PASS: TestStatusBranchHeaderShapes (2.59s)
=== RUN   TestStatusAheadBehindFromHeader
--- PASS: TestStatusAheadBehindFromHeader (1.82s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/core/git	4.825s
```

**L1. `init.defaultBranch` origin on this machine (why local is green):**
```
$ git config --show-origin --get init.defaultBranch
file:/Applications/Xcode.app/Contents/Developer/usr/share/git-core/gitconfig	main
```

**L2. Bare-init HEAD follows that config:**
```
$ git init -q --bare /tmp/t474-bare-probe/remote.git && cat .../remote.git/HEAD
ref: refs/heads/main
```
(On CI, absent any such config, git's compiled default applies → `ref: refs/heads/master`.)

**L3. Identity hypothesis eliminated — `gitFixture` injects committer identity via env**
(`status_branch_test.go:22-27`: `GIT_AUTHOR_NAME=t`, `GIT_COMMITTER_NAME=t`, …), so missing CI identity cannot be the cause.

**L4. Mechanism probe — CI condition simulated by pointing the bare remote's HEAD at a nonexistent ref** (worktree-internal gitignored probe dir, since removed):

```
$ printf 'ref: refs/heads/master\n' > remote.git/HEAD
$ git init -q -b main repo; commit one; remote add origin ../remote.git; push -q -u origin main  → PUSHED
$ git clone -q .moai/cache/t474probe/remote.git clone
warning: remote HEAD refers to nonexistent ref, unable to checkout
$ ls clone            → (empty working tree)
$ cat clone/.git/HEAD → ref: refs/heads/master
$ git -C clone status --porcelain --branch | head -1
## No commits yet on master...origin/master [gone]
$ git -C clone rev-parse HEAD~1
fatal: ambiguous argument 'HEAD~1': unknown revision or path not in the working tree.
```
← byte-identical to CI `status_branch_test.go:128 fatal: ambiguous argument 'HEAD~1'`.

**L5. Why the clone's `push -q` did NOT abort the fixture on CI (exit 0):** the unborn clone carries tracking for `master` (`origin/master [gone]`), so `push.default=simple` finds a name-matching upstream and pushes — creating `master` on the remote while `main` stays untouched:
```
$ printf 'three\n' > clone/third.txt && git -C clone add third.txt && git -C clone commit -qm three
$ git -C clone push -q; echo "push-exit=$?"   → push-exit=0
$ ls remote.git/refs/heads/ → main  master      (master NEW, main unchanged)
```
→ the repo's later `fetch` never updates `origin/main` → repo sees ahead 1 / behind 0 = CI `:122 diverged header = "## main...origin/main [ahead 1]"` and `:269 Ahead=1 Behind=0, want 1/1`, exactly.

**L6. Sibling sweep (same unpinned shape, `internal/core/git`):**
```
$ grep -rn '"init"' internal/core/git/*_test.go
helpers_test.go:16,46        runGit ... "init", "-b", "main"          (pinned)
helpers_test.go:42           runGit ... "init", "--bare", "-b", "main" (pinned — IN-REPO PRECEDENT for the repair)
status_branch_test.go:54,79,229   "init", "-q", "-b", "main"          (pinned)
status_branch_test.go:99     "init", "-q", "--bare", remote           (UNPINNED — red site, TestStatusBranchHeaderShapes)
status_branch_test.go:235    "init", "-q", "--bare", remote           (UNPINNED — red site, TestStatusAheadBehindFromHeader)
status_branch_test.go:315    "init", "-q", "--bare", barePath         (UNPINNED shape, benign today — TestNewRepositoryErrorTaxonomy only asserts NewRepository rejects the bare repo; nothing clones from it, HEAD value cannot affect its assertions)
worktree_squash_merge_test.go:454  "init", "-b", "main"               (pinned)
```

**L7. RED-now attempt — worktree-scoped config forcing (NEGATIVE result, recorded honestly):**
```
$ git config extensions.worktreeConfig true && git config --worktree init.defaultBranch master
$ go test ./internal/core/git/ -run 'TestStatusAheadBehind|TestStatusBranchHeaderShapes' -count=1
ok  	github.com/modu-ai/moai-adk/internal/core/git	4.521s    ← still GREEN
```
`git init <path>` initializes the new repository in its own context and does not consult the enclosing
repository's worktree config for `init.defaultBranch` (observed). Config was reverted immediately after
(`--worktree --unset` + extension unset; verified empty). Remaining local-RED forms were rejected:
system/global config edits are machine-wide and would collide with parallel lanes' concurrent test runs
under this repo's multi-worktree regime; env-var injection (`GIT_CONFIG_GLOBAL`, …) is refused by the
session's worktree guard. **The RED-now evidence therefore rests on L4/L5 (exact mechanism + byte-identical
fatal signature, locally measured) plus the live dual-job CI red (relayed via the lead's dispatch —
`Test (ubuntu-latest)` + `Race Test`, deterministic, all four signatures matching the mechanism's
predictions).** Full-confidence bidirectional confirmation for the fix is the CI verdict on the merged
tree — the deciding command is the CI run, not a local number.

## Baseline-attribution

All commands this run; tree `25a3212a9` (branch `WT-git-status-fixture`, worktree `.claude/worktrees/t474`, zero local commits at probe time). The L4/L5 probe ran in `.moai/cache/t474probe/` (gitignored, removed after capture). The CI signatures cited (`:122`, `:128`, `:269`) are relayed from the lead's dispatch quoting the live `Test (ubuntu-latest)` + `Race Test` failures on develop — not re-measured here (no CI access from the lane); their match to the locally-measured mechanism is what L4/L5 establish.

## Gaps

- The runner's ambient git config (`push.default`, `init.defaultBranch`) was not inspected directly — L5's push.default=simple reasoning is the git default plus the observed `[gone]` tracking state; signature match across all four CI lines is the supporting evidence, not direct runner inspection.
- The full CI log beyond the lead's 3-line excerpt was not consulted (the `warning: remote HEAD refers to nonexistent ref` line, if present in the runner log, would directly confirm the unborn-clone path on CI).
- L7's worktree-scoped RED probe result is recorded below (this file was written before running it; the result lands as an addendum).

## Residual-risk

If GitHub's runner image sets `init.defaultBranch=main` system-wide in the future, the defect would go latent again (tests pass) while remaining config-dependent — the fix (pinning `-b main`) removes that dependence regardless, which is why the pin, not a CI-config workaround, is the durable repair.
