# EC-9 and the `branch --merged` equivalence — measured

> **v2, 2026-08-24. Section Q1 below REPLACES a wrong result in v1.** The v1
> conclusion ("`git worktree remove` refuses a tree holding only ignored files")
> was an artifact of a contaminated fixture. Iteration-2 plan-audit failed to
> reproduce it, was right, and the contamination is now identified. The v1 text
> is not preserved: it was cited by `spec.md` §E, `design.md` §A.4/§A.6,
> `research.md` §D.4a, EC-9, EC-11, REQ-WR-021 and AC-WR-024, and leaving a
> superseded version in place invites one of those to be re-derived from it.

## Q1 — does `git worktree remove` delete a tree holding only gitignored files?

**Answer: YES. It removes the tree and destroys the ignored content, exit 0.**
The dirty guard is not backstopped by git here; `git status --porcelain` and
`git worktree remove` agree in disregarding ignored files.

Clean fixture — repo with `.gitignore` (`build/`) in the initial commit, a linked
worktree, one ignored file:

```
$ git status --porcelain | wc -l        → 0     # worktreeIsDirty reads CLEAN
$ git status --porcelain --ignored | wc -l → 1
$ git worktree remove ../w ; echo $?
0
$ ls ../w
ls: ../w: No such file or directory             # build/artifact.bin destroyed
```

### Why v1 measured the opposite — and the finding hiding inside the mistake

The v1 fixture was created **inside this live t209 worktree**, and MoAI's own
statusline writes state files into whatever tree the session occupies. Between
v1's `git status` measurement and its `git worktree remove`, two files appeared
in the scratch worktree:

```
$ git status --porcelain --ignored=no -uall     # in the scratch worktree
?? .moai/state/config-cache.json
?? .moai/state/context-usage.json
```

The scratch repo carried no `.gitignore` for `.moai/state/`, so those were
**untracked**, not ignored — and untracked files are exactly what
`git worktree remove` refuses. v1 measured the untracked-file refusal and
attributed it to the ignored file. The two fixtures differ in nothing that was
recorded; they differ in a file written by a process neither fixture mentioned.

Three fixtures, run to isolate it — A (`.gitignore` in the base commit, removed
promptly) exit 0; B and C (scratch tree lived long enough for the statusline to
write) exit 128, with the refusal explained by the `??` lines above, not by the
ignored file. The variable is elapsed time in a live session's tree, not
`.gitignore` placement or branch topology — both of which were tested and
rejected as explanations.

**Two lessons, both worth keeping:**

1. **A measurement taken inside a live session's worktree is not isolated.** The
   session mutates the tree asynchronously. Fixtures for worktree-disposal
   behaviour belong outside any tree a session occupies.
2. **The gap between measuring and acting is a race, and the sweep has that same
   race.** `prMergeCleanup` re-checks `worktreeIsDirty` "immediately before
   removal" (EC-11 in the M8 SPEC) precisely to narrow it, but narrow is not
   closed: a tree can go dirty between the check and the `git worktree remove`.
   Here that race is benign — it makes removal fail rather than succeed — but the
   SPEC should say that is why, rather than leaving it unstated.

### What this changes in the SPEC

REQ-WR-021's `--ignored` probe is **still the right mechanism**; its rationale
inverts. It is not a courtesy that avoids a doomed removal — it is the **only**
thing standing between the sweep and the destruction of ignored content. Every
artifact that describes it as a second layer behind git's own refusal is wrong
and must be restated, and `design.md` §A.4's "three backstop layers" is two.

### An unmeasured consequence that needs measuring before M1 lands

`.moai/state/` **is** gitignored in the real repository
(`git check-ignore -v` → `.gitignore:284:.moai/state/`), and MoAI writes into it
in every tree a session occupies. So the probe's preserve condition — "this tree
holds ignored content" — is plausibly true of **every worktree that has ever
hosted a session**, before Go build output is even considered.

If that is so, the probe preserves nearly the whole population and M1's unblocking
of the 99 merged trees is undone by M2's own guard. I could not measure it: the
worktree-isolation guard refuses `cd` and `git -C` into sibling trees, so I can
observe only this one (`--ignored` → 5 entries). **This is a hypothesis, not a
finding**, and it is the single most important thing to measure before M1 is
implemented — from outside the worktrees, over all 154 trees.

It also reframes the design question. If ignored content is universal, "preserve
on any ignored content" is too blunt, and the SPEC needs a policy that separates
*regenerable* ignored content (`.moai/state/`, build output — safe to destroy)
from *irreplaceable* ignored content (a local `.env`) — or it needs to accept
destroying the former explicitly.

## Q2 — is `git branch --merged` equivalent to zero unique commits?

**Answer: yes, they are the same predicate — unchanged from v1, and independently
reproduced by the iteration-2 audit.** A branch appears in
`git branch --merged <base>` exactly when its tip is an ancestor of base, which is
exactly when `git rev-list --count <base>..<branch>` is 0.

```
$ git rev-list --count main..WT-ec9        → 1
$ git branch --merged main                 → WT-ec9 ABSENT

$ git rev-list --count origin/main..HEAD   → 0        # branch WT-worktree-reaper
$ git branch --merged origin/main | grep -x WT-worktree-reaper
WT-worktree-reaper                                     # PRESENT
```

The audit's D2 *finding* stands — v0.1.0 never analysed the zero-unique-commit
removal class — but its prescribed remedy (copy a "unique-commit check" from
`staleKeepReason`) does not: `IsBranchMerged` is `git branch --merged`, and there
is no second predicate there to copy. Adding `git rev-list --count` would be a
second call answering the question the first already answered.

The genuinely additional guard in `staleKeepReason` is `worktreeHasLocalChanges`,
and `prMergeCleanup` already runs the equivalent (`worktreeIsDirty` →
`git status --porcelain`). Per Q1, that guard is now known to be the last line
rather than one of several.
