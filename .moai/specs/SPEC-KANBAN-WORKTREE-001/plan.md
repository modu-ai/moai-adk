---
id: SPEC-KANBAN-WORKTREE-001
title: "Implementation plan — per-card worktree lifecycle with holder liveness and mutual exclusion"
version: "0.3.0"
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, plan, milestones, lock, orphan"
tier: L
---

## §A. Context

`spec.md` is the authority for scope, requirements, and exclusions. This plan says how the work is sequenced and what must be re-measured before any edit.

Milestones are ordered by **decision-reversibility**: the choices most likely to change under review come first (which lock substrate, what the card's lock and holder mutation look like, how a refusal is surfaced), and the mechanical work — mirror, neutrality, catalog — sits at the bottom.

At v0.2.0 the release predicate of `REQ-KW-011` is the most reversible decision in the SPEC — it replaces the one two audits found broken — but it stays in M3 rather than moving to the front, because it is not independently landable: it consumes the lock's process-identity recording from M1 and the creation idempotency from M2, and reordering it ahead of either would land a predicate with nothing to release into. What moves instead is its **test**, which is built first within M3 and confirmed against the old predicate before the new one exists.

At v0.3.0 the same reasoning places the five additions. The **branch match rule** (`REQ-KW-003`, `REQ-KW-019`) is the most reversible of them — it fixes a predicate with an external consumer in `SPEC-KANBAN-BOARD-001` `REQ-KB-020`, so a change to it changes another SPEC's behavior — and it lands in M2 with the creation logic that first depends on it, not later. The **creation actor and its serialization** (`REQ-KW-022`, `REQ-KW-023`) also land in M2, and the serialization is the reason M1's lock must already exist. The **lock-clearing race condition** (`REQ-KW-014`) lands with the lock in M1, because a clearing act shipped racy and hardened afterwards has a window in every build between the two. The **two operator escapes** (`REQ-KW-020`, `REQ-KW-021`) land in M3 with the states they escape, and last among the five, since neither is reachable until the state it resolves can be produced.

## §B. Known issues and unlanded ground

- **Both prerequisites are unlanded.** `internal/kanban` does not exist in this worktree (`ls internal/kanban` → *No such file or directory*), and `SPEC-KANBAN-BOARD-001` is itself a `draft` authored in this same worktree and not on `origin/main`. M0 gates on both; nothing in M1 may proceed against a package this SPEC would otherwise have to create itself.
- **The four-hour bound has no constant.** Measured: `grep -rn '14400' --include='*.go' internal/ pkg/` exits 1 with zero matches; the only occurrence is prose at `.claude/skills/moai/workflows/factory.md:96`. `spec.md` §A.5 records the consequence. Do not restore a constant-comparison criterion during run-phase.
- **The two existing lock packages are not interchangeable.** `internal/lockfile`'s Windows path is an in-process mutex by deliberate decision (its own comment forbids upgrading it). Only `internal/spec/lock*.go` is cross-process on both platforms. Choosing the wrong one produces a race that reports success.
- **The reused lock has no stale detection on Windows.** `internal/spec/lock_windows.go`'s header comment records it as a post-MVP enhancement, with manual `del` as the documented workaround. Because this SPEC's recovery path is itself a holder change requiring the lock, inheriting that limitation ships a stuck-card deadlock. `REQ-KW-014` adds the escape; `spec.md` §A.6.1 argues it.
- **The branch prefix `resolveSpecBranch` synthesizes is the minority form.** Measured at plan time: 63 `feat/` against 3 `feature/` — 64 against 3 when re-run at promotion time, the count drifting while the ratio holds (`research.md` §C). Anything keyed on the synthesized name matches almost nothing. `REQ-KW-003` observes the reported branch and recognizes a card's branch by the SPEC identifier it carries; do not restore a prefix-keyed match at run-phase.
- **The named merge helper cannot express the gate's condition, and is unreachable anyway.** `branchMergedForCleanup` takes one branch name and returns one bool (two pull-request identities are not recoverable from it), and it is unexported in `internal/cli`, which the kanban package cannot import without a cycle. `REQ-KW-018` records the constraint; `REQ-KW-007` implements the contract here instead.
- **The `gh`-absent path never disposes in this repository.** The reachability fallback is squash-merge blind by its own comment, and `origin/main`'s last 200 first-parent commits carry 0 merge commits against 199 `(#N)` subjects. `REQ-KW-017` refuses the substitution and requires the notice.
- **A squash-aware, `gh`-free merge predicate exists, and it is still not the answer.** `IsBranchMerged` (`internal/core/git/worktree.go:233`, interface at `types.go:194`) reports merge irrespective of strategy — its S4 signal is dedicated to squash detection — and its package contains no `gh` invocation. v0.2.0 refused every degraded path on the ground that no such predicate existed, which was false. It remains unadopted for the gate on a different ground: it is per-branch and the gate is per-pull-request-identity (`spec.md` §A.4.1). Do **not** wire it into disposal at run-phase; `AC-KW-017` scans for it by name.
- **"Carries the SPEC identifier" is now a stated match rule, and it was not one at v0.2.0.** Exact token, bounded by end-of-segment or a hyphen. Measured, three distinct branches carry one card's identifier and 20 of 35 SPEC-carrying branch segments are phase-suffixed (`research.md` §C.1), so an equality rule refuses most real branches and a containment rule admits superstrings. Where a **single** branch must be resolved and more than one matches, `REQ-KW-019` refuses — and that refusal must not be applied to the disposal enumeration, where multiplicity is normal.
- **The extraction target changed, and the v0.2.0 one is excluded by this SPEC's own §C.** `internal/worktree` is the L1 worktree state guard by its own `doc.go`. The target is `internal/core/git` (`spec.md` §A.9.1). Read a candidate package's `doc.go` before placing anything in it; selecting on a dependency count is what produced the previous answer.
- **`worker-` is an unusable prefix for a card worktree's base name.** `cleanupMoaiWorktrees` (`internal/cli/launcher.go:481`), called unconditionally by `applyCCMode` (`:227`) on every `moai cc`, removes worktrees so named beneath either `.claude/worktrees/` or any directory under `~/.moai/worktrees/` — the second being this SPEC's L2 base. Measured, the filter gates both bases and the removal is non-force, so a tree named for its SPEC identifier survives; `REQ-KW-003` requires it stay that way.
- **The clearing operation of `REQ-KW-014` was racy as written at v0.2.0.** Inspect-then-unlink admits a re-acquisition in between. The removal is now conditioned on the artifact still being the one inspected. Note the defect is **Windows-only**: `internal/spec/lock_unix.go` releases its `flock` by closing the descriptor, so the 14 stale `spec-close-*.lock` files currently in `.moai/state/` are inert (`research.md` §E.3). A green run on macOS establishes nothing here.
- **A declared dependency cycle exists with `SPEC-KANBAN-BOARD-001` and is not repaired by this SPEC.** Each frontmatter names the other. The resolution is in the board sibling — this SPEC's need is a landing dependency, its need of this SPEC is a contract dependency dischargeable by citation (`spec.md` §A.4.0). Surface it; do **not** resolve it by deleting this SPEC's entry.

## §C. Pre-flight (M0 — run these before any edit)

Run as one batch. Each command's own exit status is read directly; none is read through a pipe.

1. `test -d internal/kanban && test ! -d internal/factory` — the rename gate. Currently **fails**; that is the expected plan-time state.
2. `grep -n 'holder\|last.transition' .moai/specs/SPEC-KANBAN-BOARD-001/spec.md` — the board-model gate. The card record (REQ-KB-004) must carry the holder field before this SPEC's release path has anything to release.
3. `git -C <primary> branch --show-current; git -C <primary> rev-parse --short HEAD` — the branch-guard baseline of `spec.md` §A.8. Recorded at plan time as `main` / `b59a8ba7d`; **re-read at run-phase**, never carried forward from this line.
4. `git rev-parse --git-dir; git rev-parse --git-common-dir` — confirms the current tree is a worktree and yields the primary root the board state resolves under.
5. `grep -rn '14400' --include='*.go' internal/ pkg/` — expect exit 1. Re-measure; if it now hits, `spec.md` §A.5 and its criterion are both reopened.
6. `ls internal/spec/lock*.go internal/lockfile/*.go` — confirms both substrates still exist and still carry their `_windows.go` counterparts.
7. Mirror-pair classification for every pair this SPEC will touch — re-measured, never trusted from §D.
8. `git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' | grep -oE '^(feat|feature|fix|chore|docs|spec)/' | sort | uniq -c | sort -rn` — the branch-prefix census. Recorded at plan time as 63 `feat/` against 3 `feature/`. Re-measure; the ratio is the ground for `REQ-KW-003`'s observe-don't-synthesize rule, and if it has inverted that rule is reopened (the rule itself is prefix-independent, so it survives either way — what changes is the argument's weight).
9. `sed -n '174,180p' internal/cli/session_worktree_prmerge.go` — the merge helper's signature and its squash-blindness comment, the ground for `REQ-KW-007` and `REQ-KW-017`. Re-read rather than trusted from this document.
10. `go list -deps ./internal/cli/worktree | grep moai-adk` and `go list -deps ./internal/worktree | grep moai-adk` — the import-direction baseline of `REQ-KW-018`. Recorded at plan time: `internal/cli/worktree` reaches `foundation`, `core/git`, `tui`, `tui/internal`, `worktree`; `internal/worktree` reaches **nothing** internal. Read this as establishing the *import direction* only — that an extraction is available and an export in place is the wrong edge. It does **not** select the target; a dependency count is a fact about a package's dependencies, not its purpose, and reading it as a selection is what put the derivation in the L1 state guard at v0.2.0. §C.14 selects the target.
11. `git rev-list --first-parent -200 origin/main`, classifying each commit as merge or not and counting `(#N)` subjects — the squash-merge census behind `REQ-KW-017`. Recorded at plan time as 0 merge commits and 199 of 200 carrying `(#N)`.
12. `grep -rniE 'clean tree|clean-tree|working tree is clean|tree is clean|concurrent assignment' .moai/state/kanban-source/SPEC-KANBAN-MULTISESSION-001/` — the predecessor check. Recorded at plan time as 2 hits, both orphan classification, **no assignment fence**. This is the measurement that corrected v0.1.0's replacement framing; re-run it before repeating any claim about what the predecessor carried.
13. The branch-identifier multiplicity and prefix-freedom census, three commands, grounding `REQ-KW-003`'s match rule and `REQ-KW-019`'s refusal. Recorded at v0.3.0 authoring time (`research.md` §C.1): 3 distinct branch names carry `SPEC-CODEX-PHASE2-001`; 20 of 35 distinct SPEC-carrying branch segments are phase-suffixed; and **no** SPEC identifier appearing on a branch is a hyphen-delimited prefix of another. The third is the one to re-run with care — where it now hits, the residual named in `spec.md` §A.2.1 has become present rather than structural, and `REQ-KW-019`'s refusal is the only thing standing between it and a wrong branch.
14. `sed -n '1,6p' internal/worktree/doc.go` and `sed -n '1,10p' internal/core/git/doc.go`, plus `go list -deps ./internal/core/git | grep moai-adk` — the extraction-target check of `REQ-KW-018`. Recorded: the first declares working-tree **state guard** primitives (the L1 mechanism `spec.md` §C excludes); the second declares Git repository operations including branch lifecycle; the third reports `internal/foundation` as the only internal dependency. Read the `doc.go`, not the package name — selecting on the name is the defect `spec.md` §A.9.1 repairs.
15. `sed -n '478,500p' internal/cli/launcher.go` and `grep -n 'cleanupMoaiWorktrees' internal/cli/launcher.go` — the `worker-` prefix constraint of `REQ-KW-003`. Confirm two properties rather than one: the prefix filter is applied **before** the base-path loop (so it gates `~/.moai/worktrees/` as well as `.claude/worktrees/`), and the removal is non-force. Where either has changed, `spec.md` §A.2.2's conclusion that a SPEC-identifier-named tree survives is reopened.
16. `grep -n '^dependencies:' .moai/specs/SPEC-KANBAN-{WORKTREE,BOARD}-001/spec.md` — the cycle check of `spec.md` §A.4.0. Recorded at v0.3.0 authoring time as **failing**: each names the other. Surface it; the repair belongs to the board sibling and is not performed here.

### C.1 Two shell conventions this plan inherits

1. **Never read `$?` after a pipe.** `cmd | tail` makes `$?` belong to `tail`. Redirect to a log, read `rc` from the command itself, then count failures across the **whole** log.
2. **`moai spec lint` is invoked per file.** A `*.md` glob is unsatisfiable for a multi-artifact SPEC here — each path is treated as a separate SPEC, so siblings fail with `DuplicateSPECID`.

## §D. Constraints

| Constraint | Surface in this SPEC | Handling |
|---|---|---|
| Template-First, then `make build`, then commit the regenerated `internal/template/catalog.yaml` | **Live.** The stall-threshold configuration key lands in a `.moai/config/sections/` file, and that directory is mirrored under `internal/template/templates/`. | REQ-KW-015; M4 |
| Mirror pairs preserve their measured delta | **Live**, same surface. Measured at plan time: `archive.yaml` **byte-identical**; `handoff.yaml` **differs (3 lines)**; `state.yaml` **differs (7 lines)**. Classification is time-varying — re-measure at run-phase. A sanitized pair becoming byte-identical is a failure, not a convergence. | REQ-KW-015; M4 |
| `CLAUDE.local.md` §25 — no SPEC ID, REQ/AC token, internal date, or commit SHA under `internal/template/templates/` | **Live**, same surface. The mechanical guards are the verdict; a directed grep is an early-warning aid only. | REQ-KW-015; M4 |
| Full test suite, not an affected-packages subset | **Live.** | REQ-KW-016; M4 |
| Post-rename identifiers; no occurrence of `factory` | **Live** across every milestone. | REQ-KW-001 |
| New environment-variable names are constants in `internal/config/envkeys.go` | **No surface.** The stall threshold is configuration under `.moai/config/sections/`, not an environment variable, and this SPEC introduces no new environment variable. Recorded here so the absence reads as a scope boundary rather than an oversight. | — |
| `CLAUDE.local.md` §14 — no naked platform syscall in a package body | **Live.** The lock's platform calls live in build-tagged files, exactly as the reused `internal/spec` pattern already arranges them. | REQ-KW-013; M1 |
| Primary-checkout branch guard | **Live.** Creation is measured against it (REQ-KW-006). Note the guard's mechanical enforcer ships **disabled by default**; the invariance check here does not depend on it being on. | REQ-KW-006; M2 |

## §E. Settled decisions and what each one costs

- **Reuse the `internal/spec` lock pattern, not `internal/lockfile`.** Costs a second consumer of a pattern rather than a shared package; buys cross-process correctness on Windows. The alternative reports success while serializing nothing across processes.
- **Non-blocking contention, not a blocking wait.** A session that cannot take a card's lock is told so immediately. Costs a caller-side decision about what to do next; buys the absence of a session parked indefinitely on a lock a dead process left behind on Windows.
- **The stall criterion is a configuration check.** Costs the mechanical link to the four-hour bound; buys a criterion that can actually be satisfied. The weakness is named in `spec.md` §A.5, not hidden.
- **Creation is idempotent on the matching case and refuses on both mismatches.** Costs two error paths where one would have compiled; buys a recovery path (§A.7) that does not break on its own enabling mechanism.
- **The lead disposes; no worker removes its own tree.** Costs a round trip; buys an actor that can see both pull requests. Since `SPEC-KANBAN-BOARD-001` v0.2.0 the `lead` is also the sole writer of board state (`REQ-KB-017`), so holder release is a `lead` act for a second, independent reason.
- **Release requires positive evidence of death, not merely an absent transition.** Costs a probe, a registry lookup, and a whole class of cards that cannot be released automatically at all (unprobeable holders, foreign hosts); buys the closure of the post-commit clean window that defeated v0.1.0's argument (`spec.md` §A.7.0). The rejected cheaper option — sampling cleanliness across a bounded interval — narrows that window without closing it, and adds a polling loop for the privilege.
- **The branch is observed after creation, synthesized only to create.** Costs a `git worktree list --porcelain` read on paths that previously computed a string; buys a gate that can actually fire in a repository whose branches were 63 `feat/` to 3 `feature/` at plan time (64 to 3 at promotion time — §C.8 re-measures).
- **The disposal gate is keyed on pull-request identities, discovered rather than synthesized.** Costs a `gh` dependency on the disposal path and a hard stop when it is absent; buys a gate that expresses the condition the SPEC actually requires. The existing branch-keyed predicate cannot express it at any cost.
- **Where the pull-request observation is unavailable, nothing disposes and the system says so.** Costs unbounded worktree accumulation for as long as the outage lasts; buys the absence of a check that appears to run and structurally cannot pass.
- **The merge-observation contract is implemented here rather than extracted.** Costs a second implementation of a superficially similar idea; buys leaving `session_worktree_prmerge.go`'s deliberately squash-blind, branch-keyed semantics intact for its own caller.
- **A branch is matched by exact token, bounded by end-of-segment or a hyphen.** Costs a stated rule where v0.2.0 had a phrase, and admits one structural collision (a SPEC identifier that is a hyphen-delimited prefix of another — measured absent, §C.13); buys a predicate that recognizes the 20-of-35 phase-suffixed branches this repository actually has and refuses superstrings. Equality would refuse most real branches; containment would admit anything embedding the identifier.
- **Where a single branch must be resolved and several match, nothing is picked.** Costs an operator interruption in a case that is legitimate for enumeration; buys the absence of a confidently wrong branch selection — which now propagates outside this SPEC, since `REQ-KB-020` reads a card's `status` from whatever this predicate resolves.
- **The lead creates, and creation is serialized under the card lock.** Costs a round trip and a lock acquisition on a path that previously had neither; buys the third lifecycle act having the same actor and the same refusal as the other two, and closes the interval in which a partially created tree is observable. Relying on `git worktree add`'s own refusal was the cheaper option and leaves that interval open.
- **Each terminal state gets a bounded operator-visible escape, and each escape is gated.** Costs two operations and an operator surface; buys a card held by a session on another machine, or holding a dirty orphan tree, ceasing to be permanently stuck. The force-release's gate — refuse where the holder is probeable and live — is what keeps it from being v0.1.0's predicate with a human attached.
- **The clearing act re-checks the artifact's identity before unlinking.** Costs one more observation inside an operation that already probes; buys the absence of a window in which a legitimately re-acquired lock is unlinked and two holders enter the critical section. The cost is trivial and the alternative is a silent correctness hole on the one platform the deadlock exists on.

## §F. Milestones

### M0 — Preflight and prerequisite gate

Run §C 1-16. On a failing gate 1 or 2, halt and surface — do not perform the rename, and do not author the board model here. Gates 8-16 are measurements rather than gates: each re-establishes a fact this SPEC's requirements were repaired against, and a changed measurement reopens the section it grounds rather than halting the milestone. Gate 16 is expected to fail at run-phase entry and is surfaced rather than repaired here.

### M1 — Holder mutation under a lock (highest reversibility)

The substrate choice and the shape of the critical section are the decisions most likely to be challenged, so they land first and alone.

- Add a per-card advisory lock in `internal/kanban`, following the `internal/spec/lock*.go` pattern: an exported acquire returning a named held-error on contention, a `Release`, and build-tagged `_unix.go` / `_windows.go` implementations. No naked platform call in the package body.
- Wrap the holder read → decide → holder write sequence for a card in that lock. The board state file it reads and writes is the single-origin one (REQ-KB-005), resolved from the primary checkout.
- Encode the boundary: no code path infers a holder from lock state, and no code path treats a surviving lock artifact as a live holder (REQ-KW-014).
- Record the creating process's identity in each lock artifact, and add the bounded clearing operation gated on that process being observed absent — explicit, reporting what it removed, never called from the acquire path (REQ-KW-014, `spec.md` §A.6.1). This lands with the lock rather than later: without it the Windows path ships a stuck card, and M3's recovery is the thing that gets stuck.
- Condition that removal on the artifact still being the one inspected, and abort-with-report on a mismatch. This lands in the **same** change as the clearing operation, not as a follow-up hardening: an inspect-then-unlink act shipped first has a two-holder window in every build until the condition arrives, and the window is widest exactly when the operation is being used. Build the concurrent re-acquisition row of `AC-KW-015` alongside it, and confirm it reports the unlink against the unconditioned form.
- Scope the lock so it can cover the creation sequence M2 will wrap in it (REQ-KW-023). The lock is per card and its critical section is the caller's to choose; nothing platform-specific changes.
- The holder write is a `lead` act (`REQ-KB-017`) and reaches disk through the board's atomic write (`REQ-KB-018`). This milestone adds neither; it takes the card lock around the read-decide-write and defers the rest to the board sibling.
- Do **not** add any clean-tree condition to assignment. The absence is deliberate and is checked as an absence.

### M2 — Creation, naming, per-card scope, and branch-guard invariance

- Extract the branch derivation into `internal/core/git` — whose `doc.go` declares branch lifecycle among its subjects, which imports neither consumer, and which `internal/cli/worktree` already imports — rather than exporting `resolveSpecBranch` in place, and **not** into `internal/worktree`, which is the L1 state guard `spec.md` §C excludes (REQ-KW-018, §A.9.1). Leave `internal/cli/session_worktree_prmerge.go` untouched.
- Derive the path from the card's SPEC identifier on the existing L2 scheme, its final segment being that identifier and never beginning with `worker-` (REQ-KW-003, `spec.md` §A.2.2). Synthesize a branch name **only** when creating one; everywhere afterwards read the branch the worktree reports.
- Implement the match rule as an exact token bounded by end-of-segment or a hyphen (REQ-KW-003). Write the superstring row of `AC-KW-019` before the rule and confirm it fails against a containment test — it is the row that separates the two.
- Implement the four conditions of `spec.md` §A.3 with two distinct named errors, modifying nothing on either refusal. Exercise the no-op case on both `feat/` and `feature/`; a prefix-keyed implementation passes the second and fails the first. On the fourth condition, where a branch is **searched for** by identifier, refuse on more than one match and surface all of them (REQ-KW-019) — and confirm the disposal enumeration is not caught by that refusal.
- Make the `lead` the creation actor, refusing and surfacing where no session occupying the role is resolvable, reading occupancy through `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` (REQ-KW-022), and wrap the observe-decide-create sequence in M1's card lock (REQ-KW-023). The separate-process contention test is the one that decides the second; a same-process test measures the harness.
- Read the primary checkout's branch and HEAD before and after creation and assert equality (REQ-KW-006).

### M3 — Release predicate, holder release, orphan classification, disposal

- The **age criterion** reads the card's last-transition instant and nothing else, and produces a *candidate* rather than a decision (REQ-KW-009). The threshold ships at 21600 in the configuration section; a non-positive value is refused by name.
- The **release predicate** conjoins the age criterion with a positive observation that the holder's recorded process is absent, resolved through the peer registry's `SessionID` → `PID` / `Host` (REQ-KW-011). An unprobeable holder — no entry, or a foreign host — is surfaced and **not** released automatically. Build the post-commit clean-window construction as a test before the predicate, and confirm it reports a release against an age-only predicate: a criterion that cannot separate the two measures nothing.
- Release leaves the column untouched, introduces no column, and is performed by the `lead` (`REQ-KB-017`); with no `lead` resolvable it refuses and surfaces.
- Classify the orphaned tree; clean → re-dispatchable into the same tree (which is why M2's idempotency must already be in place); dirty → record path and released holder, surface, withhold. This gate is defence in depth now, not the safety argument — do not let it back into the release predicate.
- Add the two operator escapes last within this milestone, each reachable only once the state it resolves can be produced. The **force-release** (REQ-KW-020) releases an unprobeable holder, reports the card, the holder and the reason, and **refuses where the holder is probeable and observed live** — build that refusal row first; without it the operation is the age-only predicate with a human attached. The **orphan-clear** (REQ-KW-021) removes the recorded orphan path and holder identity and makes the card re-dispatchable, and touches nothing in the tree: assert the uncommitted content byte-unchanged in the same test, since an implementation that resolves the card by discarding the work passes every end-state assertion.
- Neither escape is called from any automatic path. Check that as an absence, with a positive control.
- Disposal: lead-only, keyed on the card's **discovered pull-request identities**, opening only when at least two were discovered and every one is observed `MERGED`; never on the card reaching `done`, and never through a branch-name predicate (REQ-KW-007). Where the observation is unavailable: no disposal, and a once-per-invocation notice; no merged-branch substitution (REQ-KW-017). Refused removal records, surfaces, stops — no force, no destructive retry.

### M4 — Template mirror, neutrality, catalog, and verification

- Mirror the configuration section template-source-first, run `make build`, commit the regenerated `catalog.yaml`, and preserve each touched pair's re-measured delta.
- Run the neutrality guards and read their exit codes as the verdict.
- Run the full test suite.

## §G. Anti-patterns

- **AP-1 — Reusing `internal/lockfile` because it is the package literally named after locking.** Its Windows path is in-process only, by a decision its own comment defends. A cross-session race would survive, silently, on the one platform nobody here builds on daily.
- **AP-2 — A same-process contention test.** It passes against the wrong substrate, which is precisely the outcome the substrate choice exists to prevent.
- **AP-3 — Restoring the constant comparison for 21600 vs 14400.** Measured zero-hit; the criterion would be unsatisfiable. If a reviewer asks for it, the answer is the grep, not a constant invented to satisfy the question.
- **AP-4 — Collapsing the two creation-mismatch errors into one.** They call for different operator actions, and a single error makes the wrong one plausible.
- **AP-5 — Adopting an existing tree because it looks close enough.** The only safe adoption is path-and-branch both matching; everything else refuses.
- **AP-6 — Treating an unheld card as an error.** It is a legal steady state (REQ-KB-011). Code that "repairs" it will fight the WIP-2 board.
- **AP-7 — Escalating a refused removal to `--force`.** The refusal already told the truth.
- **AP-8 — Testing narrowly.** A prior run-phase here missed a cross-cutting template guard exactly this way.
- **AP-9 — Writing a column, a status, or a board field this SPEC does not own.** Release touches the holder. Nothing else — and it touches it as the `lead`, per `REQ-KB-017`.
- **AP-10 — Releasing on the age criterion alone.** It is the v0.1.0 predicate and it is the defect: a healthy long-running card that has just committed satisfies it, presents a clean tree, and is re-dispatched into a live session's worktree. The age criterion is necessary and never sufficient (`spec.md` §A.7.0).
- **AP-11 — Reading "cannot probe" as "absent".** No registry entry and a foreign host both mean *unknown*, and unknown does not release. Treating them as death restores AP-10 through a side door while passing every other criterion.
- **AP-12 — Restoring tree cleanliness as a liveness signal**, including in the softened form of "sample it a few times first". Any finite window still admits a session sitting clean in a long post-commit phase. It stays defence in depth behind the re-dispatch, never a release condition.
- **AP-13 — Keying anything on the synthesized branch name.** 63 `feat/` against 3 `feature/` at plan time, 64 against 3 at promotion time (§C.8 re-measures); a prefix-keyed gate silently never fires at either count. Swapping the literal to `feat/` is the same bug with the 3 and the majority figure exchanged.
- **AP-14 — Deciding disposal with a branch-name predicate.** Two pull-request identities are not recoverable from one branch name, so such a predicate cannot express the gate — it will pass the one-PR-merged case and dispose a card that ran but never synced.
- **AP-15 — Falling back to `git branch --merged` when `gh` is absent.** It cannot list squash-merged branches, this repository squash-merges, and adopting it reports that a check ran which structurally cannot pass. Refuse and notify instead.
- **AP-16 — Exporting `branchMergedForCleanup` to call it from kanban.** The compiler will refuse: the command surface imports kanban, so the reverse edge is a cycle. It is also the wrong helper regardless (AP-14). If a reviewer proposes it, the answer is `go list -deps`, not a workaround.
- **AP-17 — Shipping the reused lock's Windows path without the clearing operation.** A dead holder's artifact blocks the holder change that recovers the card, and the recovery is a holder change. The deadlock is invisible on the platform this project builds on daily, which is precisely why it must land in M1. Note the asymmetry when reasoning about it: on Unix the lock dies with its holder and the file left behind is inert — 14 such files sit in `.moai/state/` right now, none blocking anything — so a green macOS run says nothing about this.
- **AP-18 — Substituting `IsBranchMerged` for the unavailable pull-request observation.** It is squash-aware, `gh`-free, exported, and in a package already on this path, which makes it the most attractive wrong answer available and the one v0.2.0's own premise implied did not exist. It answers whether a branch's content survives in the base; the gate asks how many pull requests the card merged. `AC-KW-017` scans for it by name.
- **AP-19 — Matching a card's branch by containment.** `strings.Contains(branch, specID)` passes every suffixed-branch row and admits `SPEC-X-0010` and anything else embedding the identifier. The rule is an exact token bounded by end-of-segment or a hyphen, and the superstring row is what tells the two apart.
- **AP-20 — Picking one branch when several match.** Silently taking the first is the same class of error as synthesizing a name: confidently wrong, and wrong invisibly — and it now propagates into `SPEC-KANBAN-BOARD-001` `REQ-KB-020`, which reads a card's `status` from whichever branch this resolves. The inverse is equally an anti-pattern: applying the refusal to the disposal enumeration, where several matches are the ordinary case, breaks disposal for most cards.
- **AP-21 — Letting a worker session create its own worktree, or naming an actor and calling the race closed.** Two separate mistakes. The first leaves creation as the one lifecycle act with no actor and no refusal. The second is subtler: the command surface is per-invocation, so two `lead`-role invocations are two processes, and `git worktree add`'s own refusal does not cover the interval in which the path exists and its branch is not yet reportable. Both the actor and the lock are required.
- **AP-22 — An ungated force-release, or an orphan-clear that touches the tree.** The first is v0.1.0's age-only predicate with a human in the loop; it must refuse where the holder is probeable and live. The second destroys the work the withholding gate exists to protect while satisfying every end-state assertion — the operation records that the tree was dealt with, and never deals with it.
- **AP-23 — Resolving the frontmatter cycle by deleting this SPEC's `dependencies:` entry.** It would make the check pass and leave a real landing prerequisite undeclared. The board sibling's entry is the contract-dependency one; surface the finding.

## §H. Out of Scope

### Out of Scope — carried from spec.md §C

- The board store, the card record, the columns, the WIP limit, and the unheld state's definition — `SPEC-KANBAN-BOARD-001`.
- Preflight beyond REQ-KW-002's two gates, roles, bootstrap, configuration surfacing, quorum, dispatch, backend selection — `SPEC-KANBAN-BOOTSTRAP-001`.
- A third lock mechanism, a second worktree system, a forced-removal path, a heartbeat field.

### Out of Scope — plan-phase deliverables

- No branch, no commit, and no push at plan phase. This plan authors artifacts only.
- No edit to either existing lock package. They are read and reused, not modified.
- No edit to `internal/cli/session_worktree_prmerge.go`. Its branch-keyed, squash-blind semantics serve its own caller's requirement; this SPEC implements its own contract (REQ-KW-018).
- No definition of the `lead` role, no election, no handover. It is resolved at runtime and refused when absent (REQ-KW-002, REQ-KW-007, REQ-KW-011).
- No board-state write path of this SPEC's own, no atomicity rule, and no board-wide lock. `REQ-KB-017`, `REQ-KB-018` and `REQ-KB-019` own those.
- No operator surface for the escapes of `REQ-KW-020` and `REQ-KW-021` — no subcommand, flag, or prompt design. The operations, their gates, and their observable end-states are required here; the surface is the bootstrap sibling's command layer.
- No edit to either sibling's frontmatter, including the mutual `dependencies:` declaration §C.16 measures. It is surfaced, not repaired.
- No edit to `internal/cli/launcher.go`. Its `worker-` sweep is read and accommodated by `REQ-KW-003`'s naming rule, not changed.

## §I. Cross-references

`spec.md` §E carries the full list. The six this plan leans on directly: `internal/spec/lock*.go` (the reused pattern, and its Unix/Windows asymmetry), `internal/cli/worktree/shared.go` (`resolveSpecBranch`), `internal/core/git` (the extraction target, and the home of `WorktreeManager` and `IsBranchMerged`), `internal/cli/session_worktree_prmerge.go` (merge-state observation, left untouched), `internal/cli/launcher.go` (the `worker-` sweep), and `.claude/rules/moai/workflow/worktree-integration.md` (the L2 scheme).
