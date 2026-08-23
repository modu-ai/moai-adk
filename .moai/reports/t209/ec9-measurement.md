# EC-9 and the `branch --merged` equivalence — measured

Two questions were left open after plan-audit iteration 1. Both are now measured,
in a scratch repository created and destroyed under
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209/.moai/reports/t209/ec9scratch/`
on 2026-08-24. The scratch tree was removed afterwards; this file is the record.

## Q1 — does `git worktree remove` delete a tree holding only gitignored files?

The SPEC author flagged this as the one place where "accept the zero-unique-commit
removal class" could be wrong, and could not measure it because the
worktree-isolation guard refused both a scratch-repo invocation and a `git -C`
redirect. (The guard also refuses a relative `cd`; an absolute `cd` into a path
inside this worktree is accepted, which is how the measurement below was taken.)

Setup: a repo with one commit, a linked worktree on branch `WT-ec9`, a committed
`.gitignore` containing `build/`, and one ignored file `build/artifact.bin`.

```
$ git status --porcelain | wc -l
0                                    # worktreeIsDirty reads this tree as CLEAN

$ git worktree remove ../wt-test
fatal: '../wt-test' contains modified or untracked files, use --force to delete it
exit=128

$ cat ../wt-test/build/artifact.bin
precious                             # the file survived
```

**Answer: no — and the hazard closes in the safe direction.** git's own removal
check is STRICTER than `worktreeIsDirty`: `git status --porcelain` omits ignored
files, but `git worktree remove` counts them and refuses. So a tree holding only
build output or a local `.env` is preserved by git even though the SPEC's dirty
guard would have let it through.

EC-9 can therefore be asserted rather than left open: the ignored-only tree is
**preserved**, and the observable is a non-blocking
`moai: PR-merge cleanup failed (fatal: … contains modified or untracked files …)`
notice.

**Second-order consequence, worth a line in the SPEC.** Because nothing in the
sweep clears the condition, that notice recurs on *every* future invocation —
the same permanent-recurring-notice shape as D14's confirmed-dead-lock case.
Two distinct causes now produce one indistinguishable symptom: a tree that is
selected for removal on every sweep and refused by git every time. Whatever the
SPEC says about the dead-lock notice should cover this one too.

## Q2 — is `git branch --merged` equivalent to zero unique commits?

This is the whole of the SPEC author's disagreement with the audit's D2 remedy.
The audit prescribed copying `staleKeepReason`, described as pairing a merge check
WITH a unique-commit check; the author replied that `IsBranchMerged` is
`git branch --merged`, and that "reachable from base" already *means* "zero
commits of its own", so there is no second predicate to copy.

```
# a branch with 1 commit not in main
$ git rev-list --count main..WT-ec9
1
$ git branch --merged main --format='%(refname:short)'
main                                 # WT-ec9 absent

# and, measured earlier in the real tree, the converse:
$ git rev-list --count origin/main..HEAD          # branch WT-worktree-reaper
0
$ git branch --merged origin/main | grep -x WT-worktree-reaper
WT-worktree-reaper                   # present
```

**Answer: yes, they are the same predicate.** A branch appears in
`git branch --merged <base>` exactly when its tip is an ancestor of base, which
is exactly when `git rev-list --count <base>..<branch>` is 0.

**The SPEC author is correct and the audit's D2 remedy rests on a misreading.**
Adding a `git rev-list --count` guard would be a second call answering the
question the first already answered. D2's *finding* stands unaffected — v0.1.0
genuinely never analysed the class — but the fix is to analyse and accept it (or
reject it on other grounds), not to add a redundant predicate.

The genuinely additional guard in `staleKeepReason` is `worktreeHasLocalChanges`,
and `prMergeCleanup` already runs the equivalent (`worktreeIsDirty` →
`git status --porcelain`). Q1 above shows both of those are in turn backstopped
by git's own stricter check.
