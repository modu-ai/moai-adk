# design.md — SPEC-WORKTREE-REAPER-001

Design decisions for the three milestones. Measurements are cited with their
provenance; where two measurements of the same figure disagree, `spec.md` §B.2
carries the reconciliation rather than a silent restatement.

## §A — D1: how merge state gains a third outcome (M1)

### A.1 — The seam signature is the decision

`sessionWorktreeGhPRViewState` returns a bare `string`, and `""` is overloaded:
it means "gh errored", "no PR exists for this branch", and "the JSON was
malformed" — three situations the caller cannot separate, none of which mean
"not merged". Changing that signature is the substance of M1.

| Option | Shape | Assessment |
|---|---|---|
| **O1-a** | Keep `string`, sentinel value (`"UNKNOWN"`) | Cheapest diff, but a sentinel inside a value space that is otherwise gh's own vocabulary invites a future gh state colliding with it. Rejected. |
| **O1-b** | `(state string, ok bool)` | Idiomatic Go, expresses exactly the missing bit, composes with the existing seams by adding one return value. **Selected.** |
| **O1-c** | Typed `mergeState` enum across the whole path | Cleanest model, but rewrites `branchMergedForCleanup`'s contract and every M8 test expectation for a distinction one call site needs. Rejected (Simplicity ladder step 6). |

**Decision D1: `sessionWorktreeGhPRViewState(branch) (string, bool)`** — `ok`
false means gh produced no usable answer. `ghPRViewStateReal` returns
`("", false)` on non-zero exit or unparseable JSON, `(state, true)` otherwise.

### A.2 — The resolution order

```
gh absent            → git branch --merged (sole source, existing notice)   [unchanged]
gh answers MERGED    → merged
gh answers OPEN/…    → not merged      (git NOT consulted — REQ-WR-004)
gh gives no answer   → git branch --merged decides
git query errors     → undetermined → preserve + notice (REQ-WR-003)
```

`gh` is primary because it sees squash merges git cannot; git is consulted on a
non-answer because it sees deleted-branch merges gh cannot. Git is never allowed
to *override* a determinate gh answer — only to fill a hole.

### A.3 — Why `git branch --merged origin/main` is the right second source

Already implemented, already seam-injected (`sessionWorktreeGitBranchMerged`),
already the documented fallback. No new external dependency, no new failure
mode. It is called at most once per sweep — the existing implementation returns
all branches in one invocation — so the extra cost is one `git branch` per
sweep, not one per candidate.

### A.4 — What the git fallback *adds*: the zero-unique-commit class

The §A.2 argument establishes that git never overrides gh. It does not address
what git adds, and the plan-audit was right that v0.1.0 never analysed it.

`git branch --merged <base>` lists every branch whose tip is reachable from
`<base>` — which includes a branch created at the base with **no commits of its
own**. So M1 makes the following newly reachable: a fresh `WT-*` branch, zero
unique commits, clean tree, no PR → gh no-answer → git says merged → removal
candidate.

The class is live on this tree. `WT-worktree-reaper` — the branch authoring this
SPEC — is in it:

```
git rev-list --count origin/main..HEAD   → 0
git branch --merged origin/main          → includes WT-worktree-reaper
```

**Decision: accept the class; do not add a unique-commit predicate.** Three
reasons, in order of weight:

1. **A branch with zero unique commits holds no committed work the base lacks.**
   That is what "ancestor of the base" means. Removing its *worktree directory*
   destroys no commit, and the branch itself is never deleted by either sweep.
2. **The proposed extra guard is the same call** — audit finding D2's remedy
   rests on a misreading, resolved by measurement (§A.5).
3. **The dirty guard is what bounds the class, and it covers untracked files.**
   `worktreeIsDirty` reads `git status --porcelain`, which lists untracked
   entries as `??`. This is why `WT-worktree-reaper` survives today. And where
   that guard is too permissive, git's own removal check is stricter still
   (§A.6).

### A.5 — Audit finding D2: the finding stands, the prescribed remedy does not

Recorded explicitly so a re-auditor can see this was settled by measurement
rather than by assertion. Evidence: `.moai/reports/t209/ec9-measurement.md` §Q2.

**D2's finding stands.** v0.1.0 genuinely never analysed the removal class the
git fallback newly reaches. That gap was real and §A.4 closes it.

**D2's prescribed remedy does not stand.** It directed the SPEC to copy
`staleKeepReason`, described as pairing a merge check *with* a unique-commit
check. Measured, `staleKeepReason` calls
`WorktreeProvider.IsBranchMerged(branch, base)`, whose S1 stage is literally
`execGit(ctx, w.root, "branch", "--merged", base)`
(`internal/core/git/worktree.go`). There is no second predicate to copy: a
branch appears in `git branch --merged <base>` **exactly when** its tip is an
ancestor of base, which is **exactly when**
`git rev-list --count <base>..<branch>` is 0. Measured in both directions:

```
# branch WITH a unique commit → NOT listed as merged
git rev-list --count main..WT-ec9                  → 1
git branch --merged main --format='%(refname:short)' → main only; WT-ec9 absent

# branch with ZERO unique commits → IS listed as merged
git rev-list --count origin/main..HEAD             → 0        (WT-worktree-reaper)
git branch --merged origin/main | grep -x WT-worktree-reaper → WT-worktree-reaper
```

Adding a `git rev-list --count` guard would therefore be a second call answering
the question the first already answered. The genuinely additional guard in
`staleKeepReason` is `worktreeHasLocalChanges`, and `prMergeCleanup` already runs
its equivalent. **REQ-WR-018's accept-the-class decision stands unchanged.**

### A.6 — EC-9 measured: git's removal check is stricter than the dirty guard

v0.1.0 left this open in both directions because it could not be measured. It
now is. Evidence: `.moai/reports/t209/ec9-measurement.md` §Q1.

`git status --porcelain` omits **gitignored** files, so a tree whose only content
is ignored (build output, a local `.env`) reads clean to `worktreeIsDirty`.
Measured, with a committed `.gitignore` and one ignored file present:

```
git status --porcelain | wc -l   → 0        # the dirty guard sees a CLEAN tree
git worktree remove <tree>       → fatal: '<tree>' contains modified or untracked
                                    files, use --force to delete it   (exit 128)
cat <tree>/build/artifact.bin    → precious # the file survived
```

**The hazard closes in the safe direction**: git's own removal check counts
ignored files where `git status --porcelain` does not, so the ignored-only tree
is **preserved** even though this SPEC's dirty guard would have let it through.
EC-9 is asserted rather than left open, and AC-WR-007 verifies it rather than
merely establishing it.

That is the last of the three ways the zero-unique-commit class could have lost
work — committed work (impossible: ancestor of base), tracked-or-untracked work
(the dirty guard), ignored work (git's own check) — so §A.4's accept decision is
now backstopped at every layer rather than at two.

## §B — D2: which signal is authoritative for "a live session is in this tree" (M2)

### B.1 — Measured coverage

| Anchor signal | Live anchors it names |
|---|---|
| git worktree lock reason | **5 of 5** — t207, t209, t210, t212, t213 |
| session registry `cwd` | **1 of 5** — t207 only |

Both PID sets confirmed live: `ps -o pid=,comm= -p 36912 -p 34699 -p 51045 -p
31329 -p 15207` returned five `claude` processes.

### B.2 — Why the lock is authoritative and the registry is not

Claude Code writes the lock at `EnterWorktree` time, **in the same act that
anchors the session**, and releases it at `ExitWorktree`. It cannot drift from
the thing it describes. The registry is separately maintained: an entry's `cwd`
is corrected only when the `CwdChanged` hook calls `RelocateSession`, and that
correction is measurably not running for 4 of 5 lanes. A signal that requires a
second mechanism to stay true is a signal that will be false.

The lock also carries the owning PID inline, so liveness is one probe away.

### B.3 — Lock reason: the stored field, not the porcelain line

The porcelain renders `locked <reason>`; `locked ` is git's own prefix and is
**not** part of the stored reason. A reason-less lock renders as a bare
`locked` line.

| Porcelain line | Stored reason the parser must carry |
|---|---|
| `locked claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)` | `claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)` |
| `locked` (no reason) | `""` — locked with an empty reason ⇒ anchored (§B.4) |

Naming this explicitly removes a guess the run phase would otherwise have to
make. The PID parser extracts the integer following `pid ` in the stored reason
and is deliberately narrow: anything it does not recognise is *not* read as
"unlocked".

### B.4 — Fail-closed direction (REQ-WR-008), stated exhaustively

| Observation | Lock source's verdict |
|---|---|
| No `locked` line for the worktree | **no opinion** — registry still consulted (REQ-WR-020) |
| `locked`, PID parsed, probe says alive | **anchored** |
| `locked`, PID parsed, probe positively confirms dead | not anchored *by this source*; removal is inert (§B.6) |
| `locked`, no PID in the reason | **anchored** |
| `locked`, reason unparseable | **anchored** |
| `locked`, probe undetermined | **anchored** |
| Porcelain parse fails entirely | **anchored** for every worktree in that sweep |

"Confirmed dead" is the only negative the lock source may assert, and only when
the probe positively established death. This inverts the sweep's fail-open
posture deliberately: failing open on the sweep costs a preserved tree, failing
open on the guard costs a live session's shell.

Note the first row is "no opinion", not "not anchored" — the distinction that
makes the union in §B.7 meaningful.

### B.5 — The probe seam must be able to say "I do not know"

REQ-WR-008's undetermined case is unimplementable against the existing probe.
Measured:

- `internal/session/anchor_pid_unix.go` — `isProcessAlive` returns `bool`; any
  non-`EPERM` error collapses to `false`, i.e. is reported as *dead*.
- `internal/session/anchor_pid_windows.go` — returns `true` unconditionally.

**Decision: the anchor decision consumes a probe of shape
`func(pid int) (alive bool, determined bool)`.** Platform mapping:

| Platform | Probe result | Mapping |
|---|---|---|
| unix | `kill(pid,0)` → nil | `(true, true)` |
| unix | `EPERM` | `(true, true)` — exists, other owner |
| unix | `ESRCH` | `(false, true)` — positively dead |
| unix | any other errno | `(false, false)` — **undetermined** ⇒ anchored |
| windows | always | `(true, true)` — cannot assert dead, so never widens removal |

The existing `isProcessAlive` remains for `LiveAnchoredSessions`; the new
two-valued form wraps the same syscall rather than duplicating it. The
`(false, false)` case is what makes AC-WR-011's fourth row reachable in
production and not only through a test-only seam.

### B.6 — Dead-lock policy: inert, never unlocked

`git worktree remove` refuses a locked tree regardless of PID liveness. So a
worktree whose lock names a dead process is **not removable by this sweep**,
and attempting it produces `PR-merge cleanup failed (fatal: cannot remove a
locked working tree …)` on every sweep, forever — the exact notice already
observed for t212 in the investigation transcript.

**Decision: the confirmed-dead branch is inert.** The sweep does not attempt
removal of a locked tree, and does not unlock. Unlocking would mean taking
authority over another process's lock — a distinct escalation, held out in
`spec.md` §G. The table row in §B.4 is retained because it is the correct
*anchor* verdict; REQ-WR-021 is what stops that verdict implying a removal that
cannot happen.

### B.6a — The same symptom now has two causes, and they must stay attributable

The EC-9 measurement (§A.6) produced a second condition with the identical
shape: a tree that is selected for removal on every sweep and refused by git
every time, with nothing in the sweep clearing the condition. Two distinct
causes, one indistinguishable symptom:

| Cause | Detectable before attempting removal? | git's refusal |
|---|---|---|
| Locked tree (dead or live PID) | **Yes** — the lock line is already in the porcelain output the sweep parses | `cannot remove a locked working tree` |
| Ignored-only content | **Not by the current dirty guard** — `git status --porcelain` omits ignored files | `contains modified or untracked files` |

That asymmetry is the whole design question: one cause is free to pre-detect,
the other costs an extra probe.

**Decision: pre-detect both; preserve quietly with a cause-naming notice; never
attempt a removal that is known to fail.** For the ignored-only case the probe is
`git status --porcelain --ignored` on the candidate. Consistency with the
dead-lock decision is the reason: having decided that a permanently recurring
`cleanup failed` notice is the wrong observable for one cause, emitting it
forever for the other would be arbitrary.

Cost is bounded: the extra probe runs only for candidates that already passed the
merge check, the dirty guard, and the anchor guard — a small set in steady state,
and a one-off ~98 in the first repaired sweep.

**Rejected alternative — attempt the removal and let git refuse, with distinct
notice text.** Cheaper by one git call, and the notice text does distinguish the
causes. Rejected because the observable is a `failed` notice on every
`moai session list`, forever, for a tree that is behaving correctly. A permanent
error-shaped message for a non-error state trains readers to ignore the channel
that also carries real failures.

**Rejected alternative — add `--ignored` to `worktreeIsDirty`.** This would make
the shared helper agree with git's own check, which is attractive. Rejected
because `worktreeIsDirty` is shared with the M4 session-exit path, and widening
what it calls "dirty" changes that path's behaviour in a way this SPEC has not
analysed. The probe is therefore added in the PR-merge path only, leaving the
shared helper's semantics untouched (REQ-WR-017 continues to mean what it meant).

REQ-WR-021 is generalised to the refusal class rather than to locks alone, and
REQ-WR-023 requires the preserve notice to name which cause applied — without it
the two conditions are indistinguishable in the output, which is the state this
section exists to prevent.

### B.7 — Union, not replacement

`session.LiveAnchoredSessions` is retained unchanged and its result is unioned
with the lock verdict (REQ-WR-009 / REQ-WR-010). Two reasons: an unlocked
anchored tree is a real shape (§B.8) that only the registry can see, and the
separate card fixing `RelocateSession` will make the registry stronger — a
union composes with that fix, a replacement would have to be undone.

### B.8 — The unlocked anchor is the residual, and M2 does not close it

Measured: `moai` contains no `git worktree lock` call anywhere; the only
`unlock` reference in the codebase is a hint string in `worktree/done.go`.
Locks on the five live trees are written by Claude Code, not by `moai`. But
`materializeSessionWorktree` (`internal/cli/session_worktree.go`) creates
`WT-<session>-<subcommand>` trees with a plain `git worktree add` and never
locks them.

Those trees carry the `WT-` prefix, so they are **inside** the swept set, and
the lock source has no opinion on them. Only the registry can see them — and the
registry is the 1-of-5 source. M2 therefore does not close this case; REQ-WR-020
records it as accepted residual risk, bounded by `auto_cleanup` defaulting to
`false`. v0.1.0 described this as "a shape not measured but not excluded", which
understated it: it is produced by code in the package under repair.

### B.9 — Where the decision lives, and which consumers get it

**Decision: the anchor decision is exported from `internal/session`, beside
`LiveAnchoredSessions`.** v0.1.0 placed it in a new `internal/cli` file wired
only into `prMergeCleanup`, which leaves the blind guard on the surface with the
wider blast radius:

| Consumer | Population swept | Anchor source today |
|---|---|---|
| `prMergeCleanup` (`internal/cli`) | `WT-*` branches only | `LiveAnchoredSessions` (blind) |
| `cleanStaleWorktrees` (`internal/cli/worktree`) | **every registered worktree** (provider is `git worktree list --porcelain`) | `LiveAnchoredSessions` (blind) |

Placing the decision in `internal/session` fixes both call sites with one diff
(REQ-WR-019, AC-WR-015). Both consumers already fetch porcelain output — the
lock map is built from data each sweep already has, with no additional git
invocation.

`parseWorktreeList` must stop discarding the `locked` line: today its switch
recognises `worktree`, `branch`, and `detached` only. Extending `wtEntry` to
carry the stored reason is the smallest change that gets the signal to the
decision.

### B.10 — Where the guard runs

Same position as the dirty guard: immediately before removal, per candidate
(the EC-11 re-check position from SPEC-SESSION-WORKTREE-001).

## §C — D3: what happens to the 43 non-`WT-*` worktrees (M3)

This is the one genuine design decision in the SPEC; M1 and M2 are repairs.
**v0.1.0 decided it against an incomplete option set, and the decision is
reversed here.**

### C.1 — The population is heterogeneous

43 worktrees (excluding the primary checkout) whose branch is not `WT-*`:

| Group | Approx. count | Provenance |
|---|---|---|
| `worktree-agent-*` | 7 | runtime-created by `Agent(isolation: "worktree")` |
| `worktree-t*` | 18 | `EnterWorktree` default naming, pre-rename convention |
| `worktree-<slug>` | ~15 | same, slug-named |
| `feat/…`, `fix/…`, `card-wtiso`, `release/v*` | remainder | created deliberately by a human |

A machine cannot infer disposal intent across that mix: `release/v3.1.3` and a
stale `worktree-agent-*` are the same shape to a prefix filter and opposite
things to an operator.

### C.2 — What already ships (the option v0.1.0 missed)

`moai worktree clean --stale` (`internal/cli/worktree/clean.go`), measured:

- **Enumerates every registered worktree** — its provider is `git worktree list
  --porcelain`, with no `WT-` filter — so all 43 are already in view.
- **Prints a per-tree keep-reason** covering dirty state, merge state, live-session
  anchor, detached HEAD, base-branch checkout, and protected paths.
- **Previews by default**; `--yes` is required to remove.
- **Never deletes branches** — only worktree directories.
- Carries an `@MX:ANCHOR` on the two-part no-work-lost predicate.

The `--yes` gate is verbatim the "forward constraint" REQ-WR-014 proposed to
invent, and `TestCleanStale_PreviewsByDefault` / `TestCleanStale_RemovesWithYes`
already guard it.

v0.1.0's premise that enumerate-and-report is "exactly the work that is not
being done today" was therefore an unverified recommendation premise: the
*capability* exists. What is genuinely unverified is whether operators invoke
it — and that is not something a second command would fix.

### C.3 — Options, complete

| Option | What ships | Trade-off |
|---|---|---|
| **O3-a** — widened opt-in prefix set for `prMergeCleanup` | config key listing extra prefixes the auto-sweep removes | Removes the 7 agent trees automatically, but grants delete authority over a population whose intent cannot be inferred, gated only by a config key set once and forgotten. Rejected. |
| **O3-b** — new report-only inventory command | a second surface listing unswept trees with their states | Duplicates ~80% of `clean --stale`. Two commands answering the same question diverge. **Rejected — this was v0.1.0's selection.** |
| **O3-c** — documented manual procedure only | a doc section, no code | Indistinguishable from the status quo, under which the tail grew to 43. Rejected. |
| **O3-d** — extend `moai worktree clean --stale` | machine-readable output, `origin/main` base, the M2 lock guard | Reuses the shipped enumerate-preview-gate machinery; the only additions are the measured gaps in §C.4. **Selected.** |

### C.4 — Decision D3: extend `clean --stale` (O3-d)

Simplicity ladder step 2 binds: reuse an existing capability before writing new
code. The three measured gaps that extension closes:

| Gap | Measured | Requirement |
|---|---|---|
| No machine-readable output | flag set is `--merged-only`, `--stale`, `--yes`, `--base` — no `--json` | REQ-WR-012 |
| Base defaults to **local** `main` | `cmd.Flags().String("base", "main", …)`; `prMergeCleanup` compares against `origin/main` — the two sweeps can disagree | REQ-WR-022 |
| Same blind anchor source | `clean.go:162` calls `LiveAnchoredSessions` | REQ-WR-019 |

**Tool or documented procedure?** The card asks explicitly. The answer is
**both, split by what each is good at** — but the tool half is now *extension*
rather than construction:

- **The tool** owns the mechanical part: enumerate the trees and resolve each
  one's merge, anchor, and dirty state. `clean --stale` already does this; M3
  makes its output consumable and its base reference correct.
- **The procedure** owns the judgement: deciding a particular `feat/…` tree is
  finished with. A human is irreplaceable here, and being wrong destroys the
  only copy of unpushed work.

Ruling out O3-a remains the load-bearing half: the 43 trees are the population
most likely to hold work existing nowhere else, so it is the worst population to
extend automatic deletion over.

## §D — Milestone ordering, restated against what is actually unguarded

v0.1.0 claimed that landing M1 before M2 "means the backlog sweep runs under the
1-of-5-blind guard", implying live sessions are at risk. That was wrong in both
directions.

**Overstated for locked trees.** `git worktree remove` without `--force` refuses
a locked tree (`fatal: cannot remove a locked working tree …`, exit 128). All
five live anchors in this repository are locked, so **no live session here is
removable today, with or without M2**. M2 does not rescue them from a hazard;
it replaces an accidental protection (git's refusal, observed as a failure
notice) with an intentional one (a preserve decision, with a source-naming
notice). **That is a correctness and legibility gain, not a rescue**, and the
SPEC states it that way deliberately rather than claiming a rescue the
measurement does not support.

The gain is still worth having, for a reason "legibility" undersells: a guard
whose safety comes from a side effect of git's lock handling is one refactor
away from having no safety at all. Nothing in the codebase records that
`git worktree remove`'s refusal is load-bearing for session safety — it is not
asserted by any test, not named in a comment, and would be silently defeated by
a future `--force`, an `unlock`-then-remove convenience, or a switch to direct
directory removal. Making the protection intentional means the invariant is
written down, tested (AC-WR-014, AC-WR-015), and visible to whoever next edits
the sweep.

**Understated for unlocked trees.** The residual is the *unlocked* anchor
(§B.8), produced by this package's own materialiser, inside the swept prefix
set, invisible to the lock source. **M2 does not close it.** Only the registry
sees such a tree, and the registry is the blind source.

**Consequent ordering guidance.** M1 and M2 touch disjoint paths and can land in
either order; the M1-first hazard is not the live-session hazard v0.1.0 named.
What M1-first actually does in a repository with `auto_cleanup: true` is make
~98 merged trees removable in one sweep — a bulk action `spec.md` §G holds out
of scope. So the constraint is about **authorising the backlog removal**, not
about session safety:

> Land M1 with `auto_cleanup` temporarily `false`, or land it together with the
> AC-WR-023(b) enumerated expected-removal set, so the first repaired sweep is a
> deliberate act rather than a side effect of a `moai session list`.

M3 depends on M2 for the anchor column it reports; it does not depend on M1.

## §E — Cross-references

- `.moai/reports/t209/investigation.md` — the measured survey
- `.moai/reports/t209/plan-audit.md` — iteration-1 audit driving the v0.2.0 amendment
- `research.md` — prior-art survey of the two sweeps and their shared dependency
- `internal/cli/session_worktree_prmerge.go` — `prMergeCleanup`, `branchMergedForCleanup`, `parseWorktreeList`, the gh/git seams
- `internal/cli/worktree/clean.go` — `cleanStaleWorktrees`, `staleKeepReason`, the `--stale`/`--yes`/`--base` surface
- `internal/session/anchor.go`, `anchor_pid_unix.go`, `anchor_pid_windows.go` — `LiveAnchoredSessions` and the platform probes
- `internal/core/git/worktree.go` — `IsBranchMerged` (S1 = `git branch --merged`), the porcelain provider
- `internal/cli/session_worktree.go` — `SessionWorktreeBranchPrefix`, `materializeSessionWorktree`, `worktreeIsDirty`
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary — the L1/L2 SSOT
- SPEC-SESSION-WORKTREE-001 — shipped `prMergeCleanup` (M8) and the EC-11 guard position
