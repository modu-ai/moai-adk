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
3. **The dirty guard is what bounds the class, and it covers tracked and
   untracked files.** `worktreeIsDirty` reads `git status --porcelain`, which
   lists untracked entries as `??`. This is why `WT-worktree-reaper` survives
   today. For that class git independently re-checks at removal time and
   refuses, so the guard has a backstop. **For gitignored content it does not**
   — §A.6, corrected.

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

### A.6 — EC-9 re-measured: git does NOT backstop the dirty guard for ignored content

**This section states a result that reverses what v0.2.0 of this SPEC claimed.**
v0.2.0 recorded that non-forced `git worktree remove` refuses a tree holding only
gitignored files, and built a requirement, a criterion and two edge cases on it.
Iteration-2 plan-audit failed to reproduce it and was right. Corrected record:
`.moai/reports/t209/ec9-measurement.md` **v2** (v1 is superseded in place and
must not be re-derived from).

```
git status --porcelain | wc -l              → 0     # the dirty guard sees CLEAN
git status --porcelain --ignored | wc -l    → 1
git worktree remove <tree> ; echo $?        → 0     # REMOVED
ls <tree>                                   → No such file or directory
```

**`git status --porcelain` and `git worktree remove` agree in disregarding
ignored files.** There is no third backstop layer. §A.4's accept-the-class
decision is backstopped at **two** layers, not three: committed work (impossible
— an ancestor of base holds none) and tracked-or-untracked work (the dirty
guard, itself backstopped by git's own check). Ignored content has **no**
protection at all today.

**Why v1 measured the opposite, and why the reason is itself a finding.** The v1
fixture lived inside this live t209 worktree, and MoAI's statusline writes state
files into whatever tree a session occupies. Between v1's `git status` and its
`git worktree remove`, `.moai/state/config-cache.json` and
`.moai/state/context-usage.json` appeared in the scratch worktree — **untracked**
there, because the scratch repo carried no `.gitignore` for them. v1 measured an
untracked-file refusal and credited it to the ignored file. `.gitignore`
placement and branch topology were both tested and rejected as explanations; the
variable is elapsed time inside a live session's tree.

Two lessons this SPEC adopts:

1. **A measurement taken inside a live session's worktree is not isolated.** Any
   remaining worktree-disposal measurement — including AC-WR-025's — is run from
   outside every tree a session occupies.
2. **The gap between checking and acting is a race, and the sweep has it too**
   (§B.11).

**Scope correction, so the hazard is not overstated.** Destruction of ignored
content is a **pre-existing** property of the shipped sweep, not one M1
introduces: `prMergeCleanup` already removes merged, porcelain-clean trees today
— that is how worktree t208 was removed during the investigation. What M1 changes
is the *rate*, by making ~98 currently-preserved trees removable at once. That
distinction sets the priority (it is an M1 precondition because M1 amplifies it),
not the blame.

### A.7 — UNRESOLVED: the ignored-content policy, and the measurement that decides it

**This is the one decision this SPEC deliberately does not make at plan-phase.**
It is recorded here as an open fork with a fixed decision rule and a named
measurement, rather than guessed — twice now, a policy built on an unmeasured
premise about ignored files has had to be withdrawn.

**The hypothesis that could invalidate M1.** `.moai/state/` is gitignored in this
repository, verified:

```
git check-ignore -v .moai/state/config-cache.json
→ .gitignore:284:.moai/state/    .moai/state/config-cache.json
```

MoAI writes into that directory in every tree a session occupies — this worktree
carries `!! .moai/state/config-cache.json` and `!! .moai/state/context-usage.json`
right now (`git status --porcelain --ignored` → 7 entries, 5 of them `!!`). So
"this tree holds ignored content" is plausibly true of **every worktree that has
ever hosted a session**, before Go build output is even considered.

If that holds, a policy of "preserve on any ignored content" preserves nearly the
whole population, and **M2's guard undoes M1's unblocking of the ~98 merged
trees** — the SPEC would ship a repair that cancels itself.

**This is a hypothesis, not a finding.** It could not be measured from here: the
worktree-isolation guard refuses `cd` and `git -C` into sibling trees, so only
this tree is observable.

**The deciding measurement (AC-WR-025), run from outside any worktree:**

```
git status --porcelain --ignored   # per tree, over all registered worktrees
```
recording, for each tree that M1 would newly unblock (merged, porcelain-clean,
unanchored), whether it carries any `!!` entry — and if so, whether every such
entry lies under a regenerable path.

**The candidate policies:**

- **P1 — preserve on any ignored content.** Simplest and safest; cost is bounded
  only if the condition is rare.
- **P2 — classify by path.** Treat an enumerated set of regenerable paths as
  destroyable and preserve on anything else, so a local `.env` is protected
  while a statusline cache is not. Cost: an allowlist to maintain.
- **P3 — accept the loss explicitly.** Keep today's behaviour, document that
  ignored content in a merged clean worktree is destroyed, and name it in the
  removal notice. Cost: honest and matching git's own default, but a local
  `.env` is unrecoverable.

**The decision rule, fixed now — two sequential questions, both answered by
AC-WR-025's own output.** No branch of this rule terminates in a judgement call.

**Q1 — would P1 preserve more than half the trees M1 unblocks?**

| Answer | Policy |
|---|---|
| **No** | **P1.** The condition is rare enough that blanket preservation costs little and M1 still delivers. Stop here. |
| **Yes** | P1 would cancel M1 — it is too blunt. Proceed to Q2. |

**Q2 — among the M1-unblocked trees, is there at least one carrying ignored
content OUTSIDE the regenerable set?**

| Answer | Policy | Why this is decisive, not a preference |
|---|---|---|
| **No — every ignored entry is regenerable** | **P3.** | P2 and P3 would behave **identically** on the measured population: P2 classifies every observed entry as regenerable and destroys it, which is what P3 does. P2's allowlist would then protect nothing that exists, so it is unearned complexity — Simplicity ladder step 1, "does this need to be built at all?". |
| **Yes — at least one tree carries irreplaceable ignored content** | **P2.** | There is a measured population that P3 destroys irrecoverably and P2 preserves exactly. The allowlist earns its complexity by protecting a set that was observed, not imagined. |

**Why Q2 is a measurement and not a preference.** AC-WR-025 already collects the
discriminating datum ("and if so, whether every such entry lies under a
regenerable path"); until v0.3.1 no rule consumed it, which left the P2/P3 branch
terminating in "choose between P2 and P3" — a preference wearing a procedure's
clothes, and precisely the failure this section exists to avoid. The rule above
consumes it.

**Two sub-rules that keep Q2 from smuggling a judgement back in:**

1. **The regenerable set is enumerated FROM the measurement, not invented ahead
   of it.** AC-WR-025 records the actual ignored paths observed across the
   unblocked trees. The regenerable set is drawn from that concrete list
   (`.moai/state/`, build-output directories), so the classification is applied
   to paths in hand rather than to a hypothetical allowlist.
2. **Unclassifiable ⇒ irreplaceable.** Any observed ignored path the operator
   cannot confidently classify counts as irreplaceable, which routes Q2 to
   **P2**. This is fail-closed, consistent with every other guard in this SPEC
   (`design.md` §B.4), and it means uncertainty selects the preserving policy
   rather than being resolved by taste.

[HARD] **This rule lands before the fork is closed, not merely before M1.** A
gate whose final branch is a judgement call shows green while deciding nothing.

**What is NOT deferred:** the fork, the candidate policies, the rule that selects
between them, and the measurement that feeds it are all fixed here. Only the
measured input is outstanding. A run-phase implementer therefore has a procedure,
not a judgement call.

**Consequence for the probe.** `git status --porcelain --ignored` is the
implementation of P1 and the classifier for P2; under **P3 it is not needed at
all**. So this SPEC does not assert the probe is required — its necessity is
contingent on a fork that is open. (v2 of the measurement record calls the probe
"still the right mechanism"; that is true under P1 and P2 and false under P3, and
is flagged in the return report rather than adopted silently.)

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

### B.6a — The refusal class, re-derived over what actually refuses

v0.2.0 defined this class as `{locked, ignored-content}` on the strength of the
v1 EC-9 measurement. That measurement is withdrawn (§A.6), and with it the second
member. The class is re-derived here from what was actually observed to refuse.

| Candidate condition | `git status --porcelain` sees it | Non-forced `git worktree remove` | Live here? |
|---|---|---|---|
| Locked worktree | no | **refuses**, exit 128 | **yes** — 5 trees |
| Untracked non-ignored file | **yes** (`??`) | refuses, exit 128 | yes — already the dirty guard's job |
| Ignored-only content | no | **removes, exit 0** | yes — and it is *not* a refusal (§A.7) |
| Populated submodule | no (0 lines) | **refuses**, exit 128, not curable by `--force` | **no** — `.gitmodules` absent |
| Missing worktree directory | n/a | removes, exit 0 | n/a |

**In this repository the refusal class reduces to locked trees** — exactly the
v0.1.0 scope, before the generalisation the withdrawn measurement motivated.
Reverting outright was considered and rejected for one reason: the class is
demonstrably **not closed** (the submodule row was found by an auditor testing
beyond the SPEC's set, and found a real member the SPEC had missed).

**Decision: define REQ-WR-021 over the observable, not over an enumeration.** A
requirement that lists members claims completeness it cannot have; the next
unlisted member then falls silently outside it, which is exactly what happened.
So: *when non-forced removal refuses, preserve and name the cause; where the
condition is already visible in data the sweep holds, pre-detect it.*

- **Pre-detection set: `{lock line}`.** Free — the lock line is already in the
  porcelain output the sweep parses. It is stated as non-exhaustive.
- **Everything else falls through to fail-open.** git refuses, the sweep emits a
  non-blocking cause-naming notice, nothing is lost. Correct, merely noisier.
- **Submodules: out of scope with a measured reason** (`spec.md` §G) rather than
  in the pre-detection set — no live instance, and the fall-through path already
  handles them safely.

**What survives from the v0.2.0 reasoning.** The argument that a permanently
recurring, cause-ambiguous notice is a bad observable was sound and is kept — it
is now REQ-WR-023, widened from "distinguish these two causes" to "every preserve
notice names its cause", which is useful regardless of how many refusal causes
exist. The rejected alternatives recorded at v0.2.0 (attempt-and-fail with
distinct text; widening the shared `worktreeIsDirty`) are retained below because
both were rejected on grounds the correction does not touch.

**Retained rejection — widening `worktreeIsDirty` to `--ignored`.** Still
rejected, and now for a sharper reason: it is shared with the M4 session-exit
path, and under §A.7 it would silently impose policy P1 on that path too, before
the measurement that decides whether P1 is even viable.

**Retained rejection — attempt-and-fail with distinct notice text.** For the
locked case, pre-detection remains preferable: the condition never clears, so
attempting produces an error-shaped notice forever for a correctly-behaving tree.

### B.11 — The check→act race, and why its benignity is class-dependent

`prMergeCleanup` re-reads `worktreeIsDirty` immediately before removal (the EC-11
position inherited from SPEC-SESSION-WORKTREE-001) precisely to narrow the window
between observing a tree and acting on it. **Narrow is not closed.**

The v1 EC-9 fixture is a live demonstration: MoAI's statusline wrote two
untracked files into the tree *between* the fixture's `git status` and its `git
worktree remove`. A sweep candidate can go dirty the same way.

**Why the race is benign — and the limit of that.** For tracked or untracked
content it is benign because git performs its own check at removal time: a tree
that goes dirty after the guard passes is refused by git, so the race turns a
would-be removal into a preserved tree plus a notice. The failure direction is
safe.

That protection is **class-specific**. Ignored content gets no second
observation, because neither the guard nor git looks at it (§A.6) — so for that
class there is no race protection, and there is nothing to make benign. Which is
another way of stating why §A.7's fork has to be settled rather than assumed.

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

| Consumer | Population swept | Other guards | Anchor source today |
|---|---|---|---|
| `prMergeCleanup` (`internal/cli`) | `WT-*` branches only | dirty guard | `LiveAnchoredSessions` (blind) |
| `cleanStaleWorktrees` (`clean.go:163`) | **every registered worktree** | dirty guard + merge check | `LiveAnchoredSessions` (blind) |
| `--merged-only` path (`clean.go:95`) | **every registered worktree** | **none** — no dirty guard | `LiveAnchoredSessions` (blind) |

**Three, not two.** v0.2.0 named the first two. Verified:
`grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `95:`, `163:`.
The third is the most exposed of all: its own in-code comment records that
"`--merged-only` has no dirty guard of its own, so this is the only protection
between the sweep and a live lane's tree" — so there the blind guard is not one
layer among several, it is the entire protection. v0.2.0's `research.md` §F had
excluded `--merged-only` on the reason that it lacks the dirty/anchor pairing,
which on reflection is an argument for including it: the SPEC's own framing —
"repairing only the automatic sweep leaves the blind guard on the surface that
can remove more" — applies verbatim.

Placing the decision in `internal/session` fixes all three call sites with one
diff (REQ-WR-019, AC-WR-015, AC-WR-026). Both consumers already fetch porcelain output — the
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
| Same blind anchor source | `clean.go:163` calls `LiveAnchoredSessions` | REQ-WR-019 |

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
