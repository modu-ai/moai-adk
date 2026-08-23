---
id: SPEC-WORKTREE-REAPER-001
title: "Worktree reaper repair: merge-detection no-answer handling, lock-based anchor guard shared by both sweep consumers, and non-WT coverage via worktree clean --stale"
version: "0.2.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "worktree,cleanup,session-anchor,pr-merge,disposal,reaper"
tier: L
---

## §A — History

- **2026-08-24** — plan-phase v0.1.0 authored from card t209 ("worktree reaper"). The card was dispatched as a design request for a new tool; the plan-phase investigation (`.moai/reports/t209/investigation.md`) found a reaper already shipped, already enabled in this repository, and running on every `moai session register` / `moai session list`. The SPEC was reframed as a repair of `prMergeCleanup` plus one coverage decision. Tier L, 17 requirements, 18 criteria.
- **2026-08-24** — plan-phase v0.2.0 amendment after plan-audit iteration 1 returned FAIL 0.55 (all seven must-pass criteria passed; the failure was rubric-driven — Testability 0.30, Traceability 0.55). Five substantive changes, none cosmetic: **(1)** the criterion set was rebuilt so it can fail — the v0.1.0 form `go test -run <NewTest>` expecting "exit 0" was measured to pass on the pre-implementation tree, because Go exits 0 with `[no tests to run]` when `-run` matches nothing, and v0.1.0 additionally asserted the opposite as fact (an unobserved verification claim). **(2)** REQ-WR-018 added for the removal class M1 newly reaches — a branch with zero commits of its own is reported merged by `git branch --merged`, which v0.1.0 never analysed. **(3)** REQ-WR-019 added: the lock-based anchor decision moves to `internal/session` so it reaches **both** consumers, including `cleanStaleWorktrees`, which sweeps every registered tree rather than only `WT-*`. **(4)** the M3 decision was retaken against a complete option set — `moai worktree clean --stale` already enumerates all trees, previews by default, and gates removal behind `--yes`, so the v0.1.0 decision to build a parallel inventory surface was made against an incomplete option set and is reversed (REQ-WR-012/022). **(5)** the M2/M1 ordering caveat was restated: locked anchors are already protected by git's non-forced removal, and the residual is the *unlocked* anchor produced by this package's own materialiser (REQ-WR-020). Requirements 17 → 22, criteria 18 → 23. **A follow-up measurement round then settled the two questions this amendment left open** (`.moai/reports/t209/ec9-measurement.md`): `git branch --merged` was confirmed equivalent to zero-unique-commits in both directions, upholding REQ-WR-018's accept-the-class decision against audit finding D2's prescribed remedy (`design.md` §A.5); and EC-9 was measured to close in the safe direction — git's own removal check counts gitignored files where `git status --porcelain` does not (§A.6). The second result exposed a shared symptom between the ignored-only tree and the confirmed-dead-lock tree, generalising REQ-WR-021 from locks to the whole refusal class and adding REQ-WR-023 for notice attributability. Final counts: **23 requirements, 24 criteria**.

## §B — Problem

### B.1 — The reframing, and the evidence for it

The dispatching card asks for a worktree reaper covering three axes: a
safe-disposal condition, an L1/L2 disposal path, and session-anchor detection.
All three already exist, in `internal/cli/session_worktree_prmerge.go`
(`prMergeCleanup`, SPEC-SESSION-WORKTREE-001 M8). The sweep is enabled in this
repository (`.moai/config/sections/workflow.yaml:124` `auto_cleanup: true`) and
was observed acting during the investigation — a single `moai session list
--json` emitted preserve notices for two dirty trees, one removal, and one
lock-refusal.

So the premise "there is no reaper" is measurably false. What is true is that
the shipped reaper preserves almost everything it looks at, is blind to most of
what it should be guarding against, and cannot see a third of the trees at all.
This SPEC states that reframing explicitly because the dispatching lead's
premise differs from what was measured.

### B.2 — Measurement provenance and observed drift

Every figure below carries its measurement time, because **the population is
being mutated by the very sweep under study** and three independent
measurements disagree:

| Measured at | worktree entries | `WT-*` branches | non-`WT-*` (excl. primary) | `WT-*` merged into `origin/main` |
|---|---|---|---|---|
| investigation, 2026-08-24 (early) | 155 | 111 | 43 | 99 |
| plan-audit, 2026-08-24 (mid) | 154 | 110 | 43 | 99 |
| amendment re-measure, 2026-08-23T19:47:10Z (`date -u`) | 155 | 111 | 43 | 98 |

**Three different values are the expected result, not three errors: the sweep
under study is mutating the population between measurements.** Every row is
correct at its own instant. `prMergeCleanup` removed
worktree t208 (branch `WT-profile-test-isolation`, merged) *between* the first
two measurements — the removal notice appears in the investigation's own
transcript. A survey of a population the sweep is concurrently mutating is
itself evidence that the sweep is live, which is the SPEC's central claim.

The apparent 43-vs-44 discrepancy in the non-`WT-*` count is a boundary
convention, not drift: 155 − 111 = 44 includes the primary checkout (branch
`main`), which is not a disposal candidate. Excluding it gives 43. This SPEC
uses **43 = non-`WT-*` worktrees excluding the primary checkout**.

### B.3 — What is actually broken

**Defect 1 — merge detection cannot distinguish "no answer" from "no".**
`branchMergedForCleanup` takes the `gh` path whenever `gh` is on PATH, and
`ghPRViewStateReal` returns `""` on any gh error. The ordinary, expected case
of a merged PR whose head branch was deleted on the remote produces exactly
that error (`no pull requests found for branch "WT-forge-counts"`), so the
sweep reads `""`, concludes "not merged", and preserves the tree — permanently,
on every future sweep. The `git branch --merged` fallback that would catch it
is unreachable because it is gated on `gh` being *absent*.

The two sources fail in opposite directions, which is what makes the repair
safe: `gh` sees squash merges `git branch --merged` cannot (the documented
reason `gh` is primary); `git branch --merged` sees deleted-branch merges `gh`
cannot. Consulting the second on a *non-answer* — as distinct from a negative
answer — loses neither property.

That argument establishes only that git never **overrides** gh. It says nothing
about what git **adds**, which is the separate question REQ-WR-018 and
`design.md` §A.4 answer.

**Defect 2 — the anchor guard is 1-of-5 blind, and git's lock is what is
actually protecting live sessions.** `LiveAnchoredSessions` reads the session
registry and matches entry `cwd` against the tree. Measured coverage: the git
worktree lock reason names 5 of 5 live anchors (t207, t209, t210, t212, t213,
all confirmed live by `ps`); the registry names 1 (t207). The cause is
documented in `anchor.go`'s own doc comment — a session entering a tree via
`EnterWorktree` keeps its launch-time CWD until the `CwdChanged` hook calls
`RelocateSession`, and that relocation is not happening for 4 of 5 lanes.
During the investigation the sweep selected t212 — a tree with a live session —
for removal, and was stopped only by git refusing to remove a locked tree.

The same blind `LiveAnchoredSessions` backs a **second** consumer with a wider
blast radius: `cleanStaleWorktrees` (`internal/cli/worktree/clean.go:162`),
reached by `moai worktree clean --stale --yes`, which sweeps every registered
worktree rather than only `WT-*` ones.

**Defect 3 — 43 worktrees are outside `prMergeCleanup`'s field of view.** That
sweep skips every branch not prefixed `WT-`, permanently excluding 43 trees
including 7 `worktree-agent-*` left by `Agent(isolation: "worktree")`. The
prefix filter is not wrong as a *default* — it distinguishes a tree the tooling
created from one a human made deliberately. The coverage gap is real but the
remedy is not a new surface: `moai worktree clean --stale` already enumerates
every tree, prints a per-tree keep-reason, previews by default, and requires
`--yes`. See `design.md` §C.

### B.4 — Consequence being paid

155 registered worktrees (L1; L2 is empty), 30G on disk, 155 directory trees
under filesystem-event watch. Reported alongside: `fseventsd` RSS 25.5G, CPU
165%, swap exhausted, event loop stalled 29s.

## §C — Goal

Repair the two defects that make the shipped sweep ineffective and unsafe,
make the anchor repair reach both of its consumers, and close the non-`WT-*`
coverage gap by extending the command that already covers it — without
widening any surface's authority to delete.

## §D — Requirements (GEARS)

23 requirements. IDs 001-017 retain their v0.1.0 numbering so existing
traceability holds; 018-022 are the amendment additions, placed in the
milestone they belong to rather than appended at the end (018-023).

### D.1 — M1: merge-detection repair

- **REQ-WR-001** (Ubiquitous) — The merge-detection seam shall represent three
  distinct outcomes for a branch: *merged*, *not merged*, and *no answer*.
- **REQ-WR-002** (Event-driven) — When the `gh` query yields *no answer* for a
  branch, the merge-detection seam shall consult `git branch --merged
  origin/main` before deciding.
- **REQ-WR-003** (Event-driven) — When neither source yields a determinate
  answer, the sweep shall preserve the worktree and emit a notice naming the
  worktree path and the indeterminate state.
- **REQ-WR-004** (Unwanted) — The merge-detection seam shall not treat a
  determinate negative answer from `gh` (`OPEN`, `CLOSED`, `DRAFT`) as *no
  answer*, and shall not consult the git fallback in that case.
- **REQ-WR-005** (State-driven) — While `gh` is absent from PATH, `git branch
  --merged origin/main` shall be the sole merge source and the squash-blindness
  notice shall be emitted exactly once per sweep invocation.
- **REQ-WR-018** (Event-driven) — When the git fallback reports a branch with
  zero commits of its own as merged, the sweep shall treat it as merged and
  rely on the uncommitted-changes guard as the sole protection for that
  worktree; the sweep shall not add a separate unique-commit predicate.

  *Rationale, stated because this is the removal class M1 newly reaches:* a
  branch listed by `git branch --merged` is by definition an ancestor of the
  base, so it holds no committed work that the base does not already have. The
  shipped `moai worktree clean --stale` predicate reaches the same conclusion
  through the same call (`IsBranchMerged`'s S1 stage is `git branch --merged`)
  and pairs it with a dirty check — which `prMergeCleanup` already performs.
  The class is therefore accepted, not excluded, and its boundary is pinned by
  AC-WR-007 and EC-8 / EC-9.

### D.2 — M2: anchor-guard repair

- **REQ-WR-006** (Ubiquitous) — The anchor decision shall read the git worktree
  lock reason from `git worktree list --porcelain` as its primary anchor
  source.
- **REQ-WR-007** (Event-driven) — When a lock reason names a process id, the
  anchor decision shall probe that process for liveness and shall treat a live
  process as an anchor.
- **REQ-WR-008** (Event-driven) — When a lock line is present but cannot be
  parsed, names no process id, or the liveness probe returns an undetermined
  result, the anchor decision shall report the worktree as anchored.
- **REQ-WR-009** (Ubiquitous) — The anchor decision shall report a worktree as
  anchored when *either* the lock source or the session registry names a live
  anchor.
- **REQ-WR-010** (Unwanted) — The anchor decision shall not remove or replace
  the session-registry source; the registry remains a supplementary input.
- **REQ-WR-011** (Ubiquitous) — The preserve notice for an anchored worktree
  shall name which source detected the anchor.
- **REQ-WR-019** (Ubiquitous) — The anchor decision shall reside in
  `internal/session` alongside `LiveAnchoredSessions`, and both consumers —
  `prMergeCleanup` (`internal/cli`) and `cleanStaleWorktrees`
  (`internal/cli/worktree`) — shall use it.
- **REQ-WR-020** (Event-driven) — When a worktree carries a session-worktree
  branch prefix but no lock, the anchor decision shall report the lock source
  as having no opinion, and the sweep shall depend on the registry source
  alone for that worktree.

  *Residual risk, recorded rather than closed:* `materializeSessionWorktree`
  (`internal/cli/session_worktree.go`) creates `WT-<session>-<subcommand>`
  trees with a plain `git worktree add` and never locks them — `moai` contains
  no lock call anywhere. Such a tree is inside the swept prefix set and is
  invisible to the lock source, so it is guarded only by the registry, which is
  the source measured as 1-of-5 reliable. This SPEC does not close that case;
  it is bounded by `auto_cleanup` being `false` in the distributed default, and
  is named here so the next reader does not mistake M2 for total coverage.
- **REQ-WR-021** (Event-driven) — When a worktree is in a state that `git
  worktree remove` will refuse without `--force` — it is locked, or it holds
  content git counts as modified or untracked that `git status --porcelain`
  omits — the sweep shall detect the condition before attempting removal,
  preserve the worktree, and not attempt a removal it knows will fail. The
  sweep shall not unlock worktrees and shall not pass `--force`.

  *Two causes, one symptom.* Both conditions produce a tree selected for
  removal on every sweep and refused by git every time, with nothing in the
  sweep clearing the condition — so an attempt-and-fail design emits an
  identical `cleanup failed` notice forever, for a tree behaving correctly.
  The locked case is free to pre-detect (the lock line is already in the
  porcelain output). The ignored-content case needs one extra probe
  (`git status --porcelain --ignored`) on a candidate that has already passed
  the merge, dirty, and anchor guards. `design.md` §B.6a carries the decision
  and the two rejected alternatives.
- **REQ-WR-023** (Ubiquitous) — The preserve notice for a refusal-class
  worktree shall name which cause applied, so a locked tree and an
  ignored-content tree are distinguishable in the sweep's output.

### D.3 — M3: non-`WT-*` coverage

- **REQ-WR-012** (Ubiquitous) — `moai worktree clean --stale` shall emit a
  machine-readable report carrying, for every non-protected registered
  worktree, its path, branch, keep-reason, dirty state, merge state, and anchor
  state.
- **REQ-WR-013** (Unwanted) — The reporting path shall not remove any
  worktree.
- **REQ-WR-014** (Where — capability gate) — Where `--stale` runs without an
  explicit apply flag, the command shall preview and shall not remove.
- **REQ-WR-022** (Ubiquitous) — The `--stale` base reference shall default to
  `origin/main`, matching the reference `prMergeCleanup` compares against.

### D.4 — Cross-cutting invariants

- **REQ-WR-015** (State-driven) — While `workflow.worktree.auto_cleanup` is
  false, `prMergeCleanup` shall not enumerate worktrees and shall emit no
  output.
- **REQ-WR-016** (Ubiquitous) — The sweep shall not abort its caller on any
  failure; every failure path shall remain a non-blocking notice.
- **REQ-WR-017** (Unwanted) — The sweep shall not remove a worktree with
  uncommitted or untracked changes, an undetermined merge state, or an
  undetermined anchor state.

## §E — Constraints

- **L2 is empty in this repository.** All registered worktrees are L1
  (`.claude/worktrees/`); `~/.moai/worktrees/MoAI-ADK/.moai-worktree-registry.json`
  is `{}`. `moai worktree done` has nothing to operate on here. The applicable
  disposal path is `git worktree remove` (git refuses a locked tree without
  `--force`, per REQ-WR-021). No requirement or criterion exercises L2.
- **No bulk deletion.** 13 open PRs ride on these trees. `WT-lint-heading` is
  preserved by explicit instruction from card t154. A live session's tree must
  never be removed.
- **`auto_cleanup` ships `false`** and is `true` only in this repository. Any
  behaviour change must be invisible to users with the toggle off.
- **Fail-open on the sweep, fail-closed on the guards.** The sweep never aborts
  its caller; an *undetermined* merge or anchor state preserves the tree.
- **git's own removal check is stricter than the dirty guard.** Measured
  (`.moai/reports/t209/ec9-measurement.md` §Q1): `git status --porcelain` omits
  gitignored files, but non-forced `git worktree remove` counts them and exits
  128. Ignored-only trees are therefore preserved by git even where
  `worktreeIsDirty` reads clean. This SPEC does not change `worktreeIsDirty`
  (it is shared with the M4 session-exit path); it pre-detects the condition in
  the PR-merge path only.
- **Windows liveness is unconditionally `true`** (`anchor_pid_windows.go`), so
  the platform can never assert "dead" and therefore never widens removal
  there. Cross-platform verification is `GOOS=windows go vet` — compilation
  evidence only, not behavioural evidence.
- **Repo conventions.** `t.TempDir()` for isolation; English code, comments,
  godoc; test injection through the existing package-level function-variable
  seams — extended, never replaced.

## §F — Verification sample

`WT-web-live-todo` (worktree `.claude/worktrees/t207`) MUST be judged
NOT-disposable. Measured: unmerged (`git branch --no-merged origin/main`), no
PR, sole copy of 3 unpushed commits, and anchored by `locked claude session
t207 (pid 36912 …)` with `ps` confirming 36912 is a live `claude`.

t207 is also the *one* tree the registry sees, so a criterion satisfied by the
registry alone would still be 4-of-5 blind. The sample is exercised against the
lock-based decision with the registry emptied — `acceptance.md` AC-WR-014 — and
against the real tree in AC-WR-023.

## §G — Out of Scope

### Out of Scope — building a new reaper or a new inventory surface

- A new sweep command, daemon, or scheduled job. The sweep exists and runs on
  every session touch.
- A parallel non-`WT-*` inventory command. `moai worktree clean --stale`
  already enumerates every tree with per-tree keep-reasons; the M3 work extends
  it (`design.md` §C, option O3-d) rather than duplicating it.
- Replacing `prMergeCleanup`'s invocation points. The trigger surface is
  unchanged.

### Out of Scope — bulk disposal of the existing backlog

- Deleting the ~98 already-merged trees as part of this SPEC's execution. The
  repair makes the next sweep able to see them; the mass removal is an operator
  action taken deliberately, after the guards are in place, against an
  enumerated expected set (`acceptance.md` AC-WR-023(b)).
- Any change to `WT-lint-heading` or to trees backing the 13 open PRs.

### Out of Scope — L2 worktree lifecycle

- `moai worktree done`, the L2 registry, and `~/.moai/worktrees/` disposal.
  Measured empty here; no criterion exercises it.

### Out of Scope — unlocking worktrees

- The sweep never calls `git worktree unlock` and never passes `--force` to
  `git worktree remove` (REQ-WR-021). Acquiring the authority to unlock another
  process's lock is a distinct escalation and would need its own SPEC.

### Out of Scope — the registry's relocation defect

- Fixing why `RelocateSession` / the `CwdChanged` hook is not relocating 4 of 5
  lanes. A separate card. This SPEC routes around it by reading the
  authoritative signal, and keeps the registry as a supplementary source so the
  fix composes rather than conflicts.

### Out of Scope — closing the unlocked-anchor case

- Making `materializeSessionWorktree` lock the trees it creates, or adding an
  independent cwd probe for unlocked trees. REQ-WR-020 records this as accepted
  residual risk bounded by the toggle default; eliminating it is follow-on work.

### Out of Scope — filesystem-event and disk-pressure remediation

- `fseventsd` tuning, watch exclusions, or disk-quota policy. Named in §B.4 as
  the consequence being paid, not as a deliverable.
