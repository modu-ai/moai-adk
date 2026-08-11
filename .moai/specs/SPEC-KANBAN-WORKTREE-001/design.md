---
id: SPEC-KANBAN-WORKTREE-001
title: "Design — per-card worktree lifecycle with holder liveness and mutual exclusion"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, design, decisions, rejected-alternatives, liveness, pr-gate, import-direction"
tier: L
---

## §A. What this file is

The decisions this SPEC rests on, each with the alternative that was rejected and the **measured** reason it was rejected. It is a Tier L artifact added at v0.2.0 with the promotion; `spec.md` carries the requirements those decisions produce, and `research.md` carries the raw measurements both cite.

A decision recorded without its rejected alternative is not a decision — it is a preference, and the next reader re-litigates it for free. Every section below therefore names what was *not* chosen, and why the choice was forced rather than merely preferred.

One further discipline is specific to this SPEC, and §B is why it exists. Two of the decisions below **replace an earlier decision of this same SPEC**, made at v0.1.0 and found broken by two independent plan audits. Where that is the case the superseded decision is recorded alongside the rejected alternatives rather than quietly overwritten — a design that was wrong once, in a way that passed review, is the most likely thing for a later reader to re-propose.

---

## §B. Liveness: what makes it safe to release a card

### B.1 Positive evidence of death, not an absent transition

**Decided.** A card's holder is released only when **both** hold: the card's last-transition instant has aged past the configured threshold, **and** the holder's recorded process has been positively observed absent on the host that record names. The age criterion is necessary and never sufficient.

**Superseded — release on the age criterion alone** (v0.1.0's actual shipped predicate). It rested on a safety net that is open exactly when it is needed. The v0.1.0 argument ran: a wrongly-released card whose session is in fact alive has a dirty working tree, and a dirty tree stops the re-dispatch. Its own §A.6 records the counterexample one sentence earlier — a live session's tree is **clean** for the moment after each commit.

The composition is what makes this a design defect rather than a small hole. The threshold is measured against **column transitions**, and a long `run` phase produces none; so the population of cards the detector examines is enriched in exactly the healthy long-runners whose ordinary commits produce the clean window. The failure is not rare — it is *biased toward firing*.

**Rejected — asserting the window is short.** A race whose probability is low is still a race, and this one is skewed the wrong way, per the paragraph above.

**Rejected — leaning on the card lock to close it.** This is the most plausible wrong answer, because a lock is already in the design and looks like exclusion. `REQ-KW-013` serializes the *holders of the lock* — a read-decide-write over a holder field. It says nothing about who is **occupying the tree**. Once a release has committed, the second session takes the lock legitimately, in sequence, and walks into an occupied worktree. Exclusion over a critical section is not exclusion over a working directory.

**Rejected — confirming cleanliness across a bounded interval.** The softened form: sample the tree repeatedly, release only if it stays clean. It narrows the window without closing it. A session in a long test phase after a commit presents an unchanging clean tree for as long as the phase lasts, so *any* finite confirmation window still admits the false positive — and the design would have bought a polling loop in exchange for a smaller probability rather than a different answer.

**Rejected — a heartbeat field on the card.** Rejected at v0.1.0 and again here, for a reason that survives the repair: a heartbeat's write path dies together with the session it reports on, so a stopped heartbeat is ambiguous between a dead session and a stalled writer, while an absent process is ambiguous about nothing. It would additionally duplicate the `LastHeartbeat` the peer registry already records (`research.md` §F) and add a board write path that `SPEC-KANBAN-BOARD-001` `REQ-KB-017` reserves to the `lead`.

**What the dirty-tree gate becomes.** It is kept and demoted. It no longer carries the argument; it catches the residue — an unprobeable holder released by hand, a probe that answered wrongly, a session whose process died while a child of it still writes. Defence in depth behind a predicate that no longer needs it.

### B.2 The registry is consulted asymmetrically, and only for one thing

**Decided.** The peer registry (`internal/session/registry.go`) resolves a holder's session identifier to the recorded process identity and host that B.1's probe then examines. That is its whole job here.

**Rejected — reading an entry's existence as evidence of life.** The registry holds dead-PID entries; the repository's own doctrine already treats it as an unreliable emptiness signal. That defeats the *positive* reading.

**Corrected — v0.1.0's "not a liveness signal in either direction".** Over-broad. Dead-PID entries defeat only the positive reading; a recorded process that the operating system reports absent is evidence the session is gone. The negative reading is exactly what B.1 needs, so the exclusion is narrowed to the half that is actually true rather than kept whole for tidiness.

**Rejected — `LastHeartbeat` as the decider.** Same objection as the card heartbeat: a stopped heartbeat is compatible with two worlds, an absent process with one. Corroboration only.

### B.3 The residual, named rather than absorbed

A holder that cannot be probed — no registry entry resolves it, or the recorded host is not this one — produces **no automatic release**. The card is surfaced with the holder's identity and the reason, and the release becomes an explicit human act.

**Rejected — treating "cannot probe" as "absent".** It is the original defect returning through a side door: absence of evidence read as evidence of death. It would satisfy every other criterion in the SPEC while restoring the exact hole B.1 closes.

The cost is real and is not hidden: a holder on another host is unreachable by this mechanism, so a card held by a session on a second machine is never released automatically at all. That is a smaller and louder failure than releasing it wrongly, and it is the direction this design takes everywhere it has the choice.

### B.4 An unescapable state is not a cost, it is a defect — and v0.2.0 shipped two

**Corrected at v0.3.0.** §B.3 above described the unprobeable holder as an accepted cost and stopped there, and §A.7.1's dirty-orphan gate said re-dispatch waits "until a human clears it". Neither named an operation, an actor, or an end-state. That is the difference between an accepted cost and a permanent one: a card in a multi-machine topology is unprobeable by construction, so the "cost" was every such card, forever.

**Decided.** Each state gets a bounded, operator-visible operation with a defined observable end-state, modelled on §E.2's clearing act rather than invented — the same shape `SPEC-KANBAN-BOARD-001` `REQ-KB-013` uses to make its own unknown state "escapable rather than terminal".

- **Force-release** (`REQ-KW-020`), gated on the holder actually being unprobeable, and **refusing** where the probe succeeds and reports the process live. End-state identical to an automatic release.
- **Orphan-clear** (`REQ-KW-021`), whose end-state is that the card's recorded orphan path and holder identity are gone and the card is re-dispatchable, and which does not touch the tree.

**Rejected — one recovery verb covering both.** Tidier, and it collapses two questions with different gates: *is this holder gone?* is answerable by a probe; *has this work been dealt with?* is answerable by nobody but the operator. A single verb must take the weaker gate, and the weaker gate is none.

**Rejected — an automatic timeout on either.** It clears on inference, which §B.1 and §E.2 both already refuse. A card stuck loudly beats a card released quietly.

**Rejected — an ungated force-release.** It is §B.1's superseded predicate with a human substituted for the age criterion, and a human looking at a stuck board is not a liveness oracle. The refusal row is what makes it an escape rather than a footgun, and it is the row a test is most likely to omit.

**Rejected — an orphan-clear that resolves the card by discarding the tree's work.** It satisfies every end-state assertion while destroying the thing the withholding gate exists to protect. The operation records the operator's judgement; the operator does the dealing-with, with ordinary tools.

---

## §C. Identity: how a card's branch is recognized

### C.1 Observe the branch, do not re-derive it

**Decided.** The worktree **path** is the card's identity — deterministic from the SPEC identifier under the existing L2 scheme. The **branch** is an observed attribute: after creation, every decision (idempotency matching, orphan classification, disposal) reads the branch the worktree reports, and recognizes it as the card's by **the SPEC identifier it carries**, never by its prefix.

**Superseded — keying on the synthesized name** (v0.1.0). `resolveSpecBranch` returns `"feature/" + name`; measured, the repository runs **64** `feat/` branches against **3** `feature/` (`research.md` §C). A gate keyed on the synthesized form matches almost nothing — which is how a disposal gate comes to be written, shipped, and never once fire. The defect's shape is worth stating: it fails **silently and permanently**, and every test written with the synthesized name passes.

**Rejected — swapping the literal `feature/` for `feat/`.** The tempting one-character fix. It trades one brittle synthesis for another and inverts the failure rather than removing it: the 3 surviving `feature/` branches become the ones that never dispose. Worse, it leaves the design still asserting that a *prefix* identifies a card's branch, when a prefix is not what identifies it. `feat/SPEC-X` and `feature/SPEC-X` are both card X's branch; `feat/SPEC-Y` is not.

**Kept, narrowly — the helper as a fallback synthesis at creation only.** A name must be produced when none exists, and there is no observation to make yet. The helper remains the producer, explicitly labelled as a fallback, with its prefix divergence recorded rather than assumed away.

### C.2 Four creation outcomes, and two distinct refusals

**Decided.** Path-and-branch both matching → no-op success. Path exists on a foreign branch → a named error. Branch exists without a path → a *different* named error. Neither refusal deletes, resets, or re-points anything.

**Rejected — adopting an existing tree that looks close enough.** The predecessor was silent here, and silence's most convenient reading — reuse whatever is there — is precisely the reading that lets a card's session start writing into a tree belonging to something else.

**Rejected — collapsing the two refusals into one error.** They call for different operator actions: a path-on-a-foreign-branch is a stale or misplaced tree; a branch-without-a-path is prior work a human left behind. One error makes the wrong remedy plausible.

**Why idempotency is load-bearing rather than a nicety.** §B's recovery path re-dispatches a clean orphan card *into that same tree*, which calls creation again. If the second call errors instead of no-opping, recovery breaks on the mechanism meant to enable it.

### C.3 "Carries the identifier" needed a degree, not just a kind

**Decided at v0.3.0.** A branch names card X **iff** the segment after its type prefix begins with X's SPEC identifier and the next character is either absent or a hyphen. Where a **single** branch must be resolved this way and more than one qualifies, nothing is picked: the system refuses and surfaces all of them (`REQ-KW-019`).

**Superseded — "recognize it by the SPEC identifier it carries", unqualified** (v0.2.0). It replaced a wrong predicate with the right *kind* of predicate and left the degree unstated, which is a genuine improvement with two open ends. Measured (`research.md` §C.1): three distinct branch names carry `SPEC-CODEX-PHASE2-001`, so the phrase is not a function; and nothing in it distinguished a containment test from a token test.

**Rejected — equality on the whole segment.** The clean-looking answer, and it refuses the majority of the repository's real SPEC branches: 20 of 35 distinct SPEC-carrying segments are phase-suffixed (`-run`, `-wave5`, `-m0-close`), and those *are* the card's branches.

**Rejected — containment.** The lazy answer. It admits `SPEC-X-0010` and anything else embedding the identifier. Note the hazard is against arbitrary branch text rather than a second card: the identifier grammar ends on exactly three digits, so `SPEC-X-0010` cannot itself be a SPEC identifier.

**Rejected — picking a match by a tiebreak** (prefix precedence, most recent, the one whose worktree exists). Every tiebreak produces a confident answer with no evidence behind it, and this predicate now has a consumer outside this SPEC — `SPEC-KANBAN-BOARD-001` `REQ-KB-020` reads a card's `status` from whatever it resolves — so a wrong pick becomes the board's belief about the card.

**Scoped, not global.** The refusal binds only the single-resolution use. Enumerating a card's branches for pull-request discovery legitimately returns several, and refusing there would break disposal for most cards. Getting the scope wrong in either direction is a defect, which is why `AC-KW-019` carries a row for each.

**The residual, named.** The hyphen boundary admits a hypothetical `SPEC-X-001-EXTRA-002`, itself a valid identifier, as card `SPEC-X-001`'s branch. Measured over the 31 identifiers currently on branches, none is a hyphen-delimited prefix of another, so this is structural rather than present; `plan.md` §C.13 re-measures it, and the refusal above is what catches it if it arises.

### C.4 The path's final segment may not begin with `worker-`

**Decided.** The card worktree's base name is the card's SPEC identifier, and never a `worker-`-prefixed name.

**The measurement, and it inverts the alarm that prompted it.** `cleanupMoaiWorktrees` runs unconditionally on every `moai cc` launch and removes worktrees under either `.claude/worktrees/` or a directory beneath `~/.moai/worktrees/` — and the second base is this SPEC's L2 home, which reads as a threat to every card tree. It is not: the `worker-` prefix filter is applied **before** the base-path loop, so it gates both bases equally, and directories under `~/.moai/worktrees/` are enumerated only as containers to scan. A tree named for its SPEC identifier is skipped. Removal is additionally non-force, so even a matching dirty tree is kept and reported as kept.

So the constraint is a prohibition this SPEC must honour rather than a hazard it must defend against. **Rejected — adopting the `worker-` convention**, which is what a reader arriving from the team-worktree code reaches for. It would place every card's tree inside the sweep radius of a command operators run routinely, and for a clean tree the loss would be silent.

`SPEC-KANBAN-BOOTSTRAP-001` records the same constraint from its side and assigns it here, so the two documents agree on the owner.

---

## §D. Disposal: what evidence opens the gate

### D.1 Keyed on pull-request identities, discovered rather than synthesized

**Decided.** The gate opens only when **at least two** pull requests were discovered for the card — enumerated from the repository by the card's SPEC identifier and the branch its worktree reports — and **every** discovered one is observed `MERGED`, each verified individually in the `gh pr view <PR> --json state` form `.claude/rules/moai/workflow/spec-workflow.md:437` already prescribes for the same act.

**Superseded — the named branch predicate** (v0.1.0). `branchMergedForCleanup` takes one branch name plus a boolean and returns one bool (`research.md` §D). The gate requires **two distinct pull-request identities** to be merged, and two identities are not recoverable from one branch name. The requirement was therefore unimplementable as written — not awkward, unimplementable.

**Why "at least two, and every one".** Requiring two is what makes the sync pull request load-bearing: a card with one merged pull request has run but not synced. Requiring *every* discovered one avoids classifying which is which by title or branch convention, and fails in the safe direction — an unrelated open follow-up holds the tree rather than releasing it.

**Rejected — inferring merge from the card's column.** It would make the board's own record the evidence for a fact the board does not own; a board wrong about a merge would then authorize deleting a tree holding the only copy of something. A card reaching `done` is a board fact; a merged pull request is a repository fact, and the tree's disposability depends on the second.

**Rejected — the lead's convenience.** The executor is the lead session, and no worker removes its own tree. A session removing the tree it is running inside is removing the ground under itself, and a session that has just finished its phase is the one least able to see whether the *other* pull request has merged. The lead is the only actor that observes both.

### D.1.1 Creation gets the same actor, and separately gets serialized

**Decided at v0.3.0.** The `lead` creates a card's worktree, refusing where no session occupying the role resolves — the same shape D.1 gives disposal and §B gives release. And, separately, the creation sequence runs beneath the card's existing lock.

**Superseded — "the system shall create"** (v0.2.0), which named no actor for the one lifecycle act that lacked one. That is not merely an inconsistency of prose: the reading that fills the gap most naturally is *the dispatched worker creates its own tree*, and that reading changes the concurrency, because two workers can then call `git worktree add` for one path with nothing serializing them. §E's lock was scoped to holder mutation, the four-outcome table describes sequential observations, and neither sibling claims the decision — `SPEC-KANBAN-BOOTSTRAP-001` §C returns creation, naming and per-card scope here by requirement id. Owned by nobody is this family's F1 failure shape.

**Rejected — naming the actor and treating the race as closed.** The most likely half-fix. The `lead` is a role occupied by a session, but the command surface is per-invocation — §E already records that two `lead`-role invocations are two operating-system processes — so concurrency survives the actor rule intact.

**Rejected — relying on `git worktree add`'s own refusal plus §C.2's idempotency.** Nearly right, which is the danger. A serialized race does land on either git's refusal or the no-op success of the matching case. What neither covers is the interval in which the directory exists and the branch is not yet reportable: the second caller observes a state none of the four outcomes describes, and the four outcomes are the whole of the creation contract.

**Rejected — a board-wide lock for creation.** Wrong granularity. Creating a card's worktree is a per-card read-decide-create with no board-wide invariant at stake; `REQ-KB-019`'s scope exists for invariants a per-card lock cannot express, such as the WIP bound, and borrowing it here would serialize every card's creation against every other's for nothing.

The two are `REQ-KW-022` and `REQ-KW-023`, kept apart because an implementation can satisfy either alone and the failure modes do not overlap.

### D.2 Where the observation is unavailable, nothing disposes — and the system says so

**Decided.** No disposal at all, plus a once-per-invocation notice recording that disposal is suspended and worktrees will accumulate unreclaimed.

**Rejected — falling back to `git branch --merged`.** This is the fallback that already exists in the neighbouring code, which is exactly what makes it dangerous to copy. Its own source comment records it as **squash-merge blind** — squash-merged branches are not listed. Measured, this repository squash-merges: over the last 200 first-parent commits of `origin/main`, every one has exactly **one** parent and **199 of 200** carry a `(#N)` subject (`research.md` §B). So the fallback's consequence here is not a degradation but an absolute — with the observation unavailable, disposal never happens. Not less often; never.

Adopting it would therefore be **worse than refusing**, and that is the whole reason this is a decision rather than an inherited default: a check that structurally cannot say yes is indistinguishable, from the outside, from a card that simply has not merged yet. Refusing is at least legible.

**Rejected — `IsBranchMerged`, and this rejection replaces the argument v0.2.0 actually made.** The paragraphs above were, at v0.2.0, the *whole* case: no non-`gh` path exists because a merged-branch listing cannot see squashes. That generalization is false, and it is falsified by a predicate this design already adopts. Measured (`research.md` §D.4), `internal/core/git`'s `IsBranchMerged` is documented as reporting merge "irrespective of the merge strategy that placed them there", through an ordered OR whose **S4 is a dedicated squash-merge probe** conjoined with a state check; its package contains **zero** `gh` invocations; it is exported on the `WorktreeManager` interface this design adopts in §C; and it is live, gating `moai worktree clean --stale`. A squash-aware, `gh`-free merge predicate was inside the adopted mechanism the entire time, and the section deciding merge detection never named it.

It is rejected anyway, on a ground that survives measurement: **it is per-branch, and the gate is per-pull-request-identity.** The gate needs *at least two pull requests discovered* — the clause that makes the sync pull request load-bearing — and a branch predicate cannot supply a count of pull requests. Branch count is not pull-request count in either direction: measured, three branches carry one card's identifier (§C.3); a branch can exist with no pull request; two pushes to one branch are one pull request. And `IsBranchMerged` answers a *content* question — does this branch's work survive in the base — so a card that ran and never synced is indistinguishable from one that did both on a single branch.

**Why this rejection is recorded at length rather than corrected quietly.** The conclusion did not move; the argument did, and the argument is what a later reader re-uses. A design that reaches the right outcome from a false premise is fragile in a specific way: nothing will ever contradict it, and the next person to check the premise will conclude the outcome was wrong. Recording the falsification is what keeps the decision re-derivable.

**Rejected in advance — re-expressing the gate's at-least-two condition over branches**, which would make a degraded path available. The measurement above is the reason: a branch count as a proxy for a pull-request count is wrong in the permissive direction, and the permissive direction deletes trees.

**The cost, accepted rather than hidden.** Worktrees accumulate and are never reclaimed for as long as the outage lasts. That is why the refusal is *surfaced* rather than logged, and the precedent is local: the same file already emits a once-per-invocation on-entry blindness notice, with a test asserting the notice documents the blindness.

**Rejected — preserving quietly.** Preservation is the right direction (the alternative to preserving is deleting a tree whose merge state was never established), so it is kept — but silence turns an unbounded accumulation into an invisible one.

---

## §E. Exclusion: the lock, and the artifact it can leave behind

### E.1 The `internal/spec` pattern, not `internal/lockfile`

**Decided.** The card-scoped advisory lock reuses the `internal/spec/lock.go` family's pattern — `flock(2)` on Unix, atomic create (`O_CREATE|O_EXCL`) on Windows, platform bodies in build-tagged files — keyed per card rather than per SPEC close. Non-blocking: a session that cannot take the lock is told so immediately.

**Rejected — `internal/lockfile`,** the package literally named after locking. Its Windows path is a `map[string]*sync.Mutex`, and its own comment states that concurrent writes across **different** OS processes are not protected, and forbids upgrading it to `LockFileEx` (`research.md` §E). Kanban sessions are distinct OS processes, so it would hold nothing in production while passing every same-process test. The failure would be silent, on the one platform nobody in this project builds on daily.

That property propagates into the verification: `AC-KW-014` demands a **separate-process** contention test, because a same-process test measures the harness rather than the lock.

**Rejected — a third locking mechanism.** Two exist; one is right. A third multiplies the platform matrix for no capability.

**Rejected — blocking acquisition.** It buys nothing and costs a session parked indefinitely behind a lock a dead process left on Windows — which §E.2 is about.

### E.2 The Windows artifact, and the deadlock it would otherwise ship

**Decided.** Each lock artifact records the identity of the process that created it. A bounded clearing operation removes such an artifact **only when** that recorded process is positively observed absent — the same probe as §B.1 — **and** the artifact is still the one that was inspected. The clearing is an explicit, operator-visible act that reports what it removed, or that it aborted. The acquire path never calls it.

**The problem is specific, has an unpleasant shape, and is confined to one platform.** The reused substrate performs **no stale-lock detection on Windows**; its header comment records cleanup as manual (`del …`) and stale detection as post-MVP (`research.md` §E.2). On Unix it is a non-issue, and measurably so rather than by assumption: `internal/spec/lock_unix.go` holds `flock(2)` on an open descriptor with its own comment recording that "close releases the flock atomically", and the release path never unlinks — so `.moai/state/` currently holds **14** zero-length `spec-close-*.lock` artifacts, the oldest from 2026-05-30, every one inert (`research.md` §E.3). On Windows the artifact persists *and* holds, and is indistinguishable from a live hold — and §B's recovery path is **itself a holder change**, which takes the card's lock. Recovery is blocked by exactly the artifact it exists to clear. The card is stuck until a human deletes a file by hand.

**Stating this asymmetry is part of the decision, not colour.** A hazard invisible on the machine a design is written on and permanent on the machine it runs on is one nobody will report; the acceptance criteria that judge it therefore record the platform they ran on, and a green macOS run is not evidence about this at all.

### E.2.1 The clearing act as first written had a check-then-unlink race

**Superseded — inspect, probe, unlink** (v0.2.0). Three steps, two gaps, and in either gap the lock can be **legitimately released by its owner and re-acquired by a live process**. The clear then unlinks a valid lock, and two holders are inside the critical section the lock exists to hold. Nothing in the requirement, in §E.2 as written, or in the three rows of `AC-KW-015` conditioned the removal on the artifact still being the one inspected — and none of those rows was concurrent, so the defect was not merely unguarded but unobservable.

The window is not narrow in the way it looks. The operation is invoked precisely when a card has been stuck, which is when an operator is most likely to be restarting sessions — so a re-acquisition landing between the probe and the unlink is the ordinary consequence of the situation that motivated the invocation, not a coincidence.

**Decided.** The removal is conditioned on the inspected artifact's identity still holding — a recorded-identity re-read under the same handle, or a content or inode match. A mismatch **aborts and reports**; it does not retry and does not fall through.

**Rejected — treating the window as too small to matter.** The same argument §B.1 already refuses for the post-commit clean window, and refused here for the same reason plus one: this window is *correlated* with the operation's own use.

**Rejected — holding a lock over the clearing act.** Circular. The artifact being cleared is the lock.

`SPEC-KANBAN-BOARD-001` `REQ-KB-023` carries the identical condition at board scope, having taken this requirement's shape and closed this hole in it before this revision reached it. The scopes differ — board-wide there, per-card here — and the condition does not.

**Rejected — inheriting the limitation silently.** It ships that deadlock into the platform this project exercises least, which is the same failure mode §E.1 refuses for the in-process fallback.

**Rejected — automatic stale detection alone (a timeout, or age).** It clears on inference, and inference about a live holder is precisely what §B.1 replaces with positive evidence.

**Rejected — a manual clearing act alone.** It leaves the card stuck until somebody notices, which is the state being repaired.

The escape is therefore the **conjunction**: recorded identity, probe-gated clearing, explicit invocation. The residual is named — where the recorded process cannot be probed at all (a foreign host), the artifact is not clearable by this mechanism and manual deletion remains the last resort. That is a narrower stuck state than the one being repaired, not its elimination.

### E.3 What the lock is not

Lock state is never a liveness signal, in either direction: the holder is never inferred from a lock's presence, and a surviving artifact is never read as a live holder. A lock is a statement about a critical section, not about a session.

The converse **is** in scope and is easy to mistake for a contradiction: §E.2 probes an artifact's recorded process in order to **clear the artifact**. That is a statement about the artifact, never about the card's holder — which stays governed by §B.1 alone.

---

## §F. Reuse: extraction versus export

**Decided.** Branch derivation is **extracted** into `internal/core/git`, and both consumers call it there. Merge observation is **not** extracted; this SPEC owns and implements that contract itself.

**Superseded — `internal/worktree`** (v0.2.0), chosen on the strength of being "the existing dependency-free leaf". That property is true and was the wrong criterion. Measured, `internal/worktree/doc.go:1-5` declares the package as **working tree state guard primitives** — Snapshot, Divergence, SuspectFlag, the L1 mechanism `spec.md` §C names explicitly "so an implementer does not wire it by mistake". v0.2.0 selected, as the home for an L2 naming concern, the implementation of the mechanism this SPEC warns implementers away from, and cited the line of that same `doc.go` one below the sentence that decides it. A second supporting claim in the same breath — "the leaf both consumers already reach" — was false: `internal/kanban` does not exist, so one consumer reached nothing.

**The replacement was chosen by reading `doc.go`, which is the discipline whose absence produced the previous answer.** Three measurements (`research.md` §D.5): `internal/core/git`'s own `doc.go` declares its subject as Git repository operations implementing "BranchManager: branch lifecycle", and a SPEC-identifier-to-branch-name derivation *is* branch naming; `go list -deps` reports `internal/foundation` as its only internal dependency, so it imports neither consumer and cannot import kanban; and `internal/cli/worktree` already imports it, so the consumer holding the symbol today needs no new dependency. It is additionally where `WorktreeManager` and `IsBranchMerged` already live, so the extraction consolidates rather than scatters.

**Rejected — keeping `internal/worktree` on the grounds that it compiles and the move is cheap.** Cohesion is what is being weighed, not compilation. Placing this SPEC's code inside the package this SPEC declares is not its mechanism makes §C's exclusion unreadable to the next implementer.

**Rejected — a new package for the derivation.** A third worktree-adjacent package to reason about, for one function with an existing home.

**The constraint that forces the split treatment.** Both named symbols are unexported, and they fail differently. `internal/cli` is the command surface: `cmd/moai/main.go` imports it, and the kanban command lives there, so the dependency must run `internal/cli` → `internal/kanban`. A kanban package importing `internal/cli` to reach `branchMergedForCleanup` closes that loop into an import cycle and the compiler refuses it (`research.md` §D.2).

**Rejected — exporting `resolveSpecBranch` in place and importing `internal/cli/worktree`.** It compiles today; `internal/cli/worktree` does not import `internal/cli`, so no cycle exists in that direction. It is still the wrong edge, and the measurement makes it a firmer one to avoid than it first appears: `internal/cli/worktree` is a **live production dependency of the command surface**, imported by `internal/cli/root.go:14` and `internal/cli/inventory.go:17` — both `package cli`, neither a test (`research.md` §D.3). So the export would point a domain package at something the command surface actively depends on, and it would make a future edge from `internal/cli/worktree` into kanban a cycle — buying nothing over an extraction into a package whose declared subject the behaviour belongs to and which one consumer already imports.

**Rejected — lifting `branchMergedForCleanup` into a shared package and widening its signature.** The superficially symmetric move, and it fails for a reason that is not about the import graph at all: the signature cannot express the two-identity condition (§D.1), so extracting it relocates a helper that still cannot answer the question. Widening it there would change an existing caller's contract — `session_worktree_prmerge.go`'s semantics are branch-keyed and deliberately squash-blind, serving its own requirement — which is a modification of a surface this SPEC declares out of scope.

What is reused from that file is the **form** (`gh pr view <PR> --json state`, per `spec-workflow.md:437`) and the **precedent** (the once-per-invocation blindness notice), not the function.

**Why this is a requirement (`REQ-KW-018`) rather than a plan note.** The cycle is discovered by the compiler, which means it is discovered *after* someone has written the import. Recording the constraint is what stops the next reader re-proposing the export and rediscovering it the expensive way.

---

## §G. What this SPEC defers, and to whom

This section exists because of a measured failure in this SPEC family. The predecessor split produced a hole with a boundary drawn around each side: two SPECs each disclaimed the same decision toward the other, and neither claimed it (`SPEC-KANBAN-BOARD-001` `research.md` §I). A deferral is therefore stated **positively** — naming the owner and the requirement id — rather than as an absence.

- **Who may write board state.** A card's holder is board state. `SPEC-KANBAN-BOARD-001` `REQ-KB-017` makes the session occupying the `lead` role its **sole writer**, enforced rather than documented. Every holder mutation this SPEC describes is consequently a `lead` act. This SPEC grants no write path of its own, and its release path is subject to that rule rather than an exception to it.
- **How that write reaches disk.** `REQ-KB-018` (same-directory temporary file, atomic rename). This SPEC states no atomicity rule.
- **Board-wide exclusion.** `REQ-KB-019` holds a board-wide lock across a whole board mutation, and supersedes card scope *for board mutations* while explicitly preserving this SPEC's card-scoped lock for holder assignment. The two are not redundant: a WIP bound cannot be enforced by two mutations each holding only their own card's lock, and a per-card read-modify-write does not need board scope. Different granularities, different invariants.
- **The `lead` role itself** — its definition, election, launch, and the topology it sits in. `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004`. This SPEC consumes the role name to say who acts, and refuses at runtime when no `lead` resolves.
- **How a session's role occupancy is read at runtime.** `REQ-KS-006`, which since that SPEC's v0.3.0 requires each session to declare the role it occupies and requires the declaration **resolvable from a session that is not the `lead`** — a clause written for this SPEC's three gates, all of which resolve the lead from outside it. v0.2.0 deferred this to `REQ-KS-004`, which elects the role without defining how anyone reads it, so the deferral pointed at a requirement that did not carry the obligation; corrected at v0.3.0. The declaration's **carrier** is left open there and is not fixed here.
- **The operator surface for the escapes of §B.4.** This design fixes the operations, their gates, and their observable end-states; the subcommand, flag, or prompt through which an operator reaches them belongs to the bootstrap sibling's command layer.
- **Dispatch.** `REQ-KW-012` decides *whether* a card may be re-dispatched; who sends the dispatch and by what message is the bootstrap sibling's.

**A note on the dependency edge, which is deliberately asymmetric.** The `lead` is a **runtime** dependency, resolved where the creation, disposal and release paths run, and it is *not* a `dependencies:` frontmatter entry. `SPEC-KANBAN-BOOTSTRAP-001` already lists this SPEC among its own dependencies, so the reverse declaration would state a cycle and imply a landing order the siblings do not agree on. The in-family convention for this shape — `related_specs:`, citation by requirement id, discharge at runtime — is the one `SPEC-KANBAN-BOARD-001` already uses for the same role.

**And a note on the edge where that convention is currently broken, by the other document.** Measured at v0.3.0, this SPEC and `SPEC-KANBAN-BOARD-001` each name the other in `dependencies:` — the cycle the paragraph above forbids, stated twice. The asymmetry that resolves it is a difference in *kind* rather than a preference: this SPEC's need is a **landing** dependency (the holder field must exist before there is anything to release), while the board's need is a **contract** dependency (`REQ-KB-020` consumes `REQ-KW-003`'s identification rule, readable from this document with no code landed, exactly as this SPEC consumes `REQ-KS-006`). The declared edge belongs on the landing dependency. This frontmatter is deliberately unchanged and the finding is surfaced instead — a SPEC that unilaterally drops a dependency to break a cycle its sibling created leaves the family with an undeclared prerequisite and no record of why (`spec.md` §A.4.0).

---

## §H. Out of Scope

### Out of Scope — decisions this file does not make

- The board state store, the card record, the column enumeration, the WIP limit, and the unheld state's definition. All `SPEC-KANBAN-BOARD-001`; §G names each by requirement id.
- Who elects or hands over the `lead` role, the dispatch transport, bootstrap, quorum, and backend selection. All `SPEC-KANBAN-BOOTSTRAP-001`.
- The internal design of `internal/spec/lock.go` and `internal/core/git`'s `WorktreeManager`. §E.1 and §C decide to reuse them, not how they work.
- The L2 persistent-worktree path scheme, lifetime, and disposal contract. Adopted whole from existing doctrine; no second worktree system is designed here.

### Out of Scope — alternatives rejected rather than deferred

- A heartbeat field, a per-session liveness ping, and any form of cleanliness-as-liveness including bounded-interval sampling (§B.1). Their absence is a decision, not a gap.
- A third locking mechanism, and any "upgrade" of `internal/lockfile`'s Windows path (§E.1).
- A widened `branchMergedForCleanup`, and **any per-branch merge predicate** substituted for the unavailable pull-request observation — the reachability listing and the squash-aware `IsBranchMerged` alike (§D.2, §F).
- Re-expressing the disposal gate's at-least-two condition over branches, which would make a degraded path available (§D.2).
- An automatic timeout on either operator escape, and a single recovery verb covering both (§B.4).
- Placing the branch derivation in the L1 worktree state-guard package (§F), and selecting any extraction target by its dependency count rather than its documented subject.
- A forced or retrying removal path. A refused removal has already told the truth: something is still in there.

---

## §I. Cross-references

- `spec.md` §A.2 … §A.9 — the context these decisions compress, and §B, the requirements they produce.
- `research.md` §B … §F — the commands and observed outputs every measurement above cites.
- `plan.md` §E (settled decisions and their costs) and §G (AP-1 … AP-23) — the same decisions as an execution order, and the route back into each rejected alternative.
- `acceptance.md` AC-KW-002 … AC-KW-023 — the criteria that decide them, notably AC-KW-011 (the post-commit clean-window construction and its negative control), AC-KW-015 (the concurrent re-acquisition row), AC-KW-018 (the compiler as the import-cycle instrument), AC-KW-019 (the superstring negative control and the enumeration scope row), and AC-KW-020 (the live-holder refusal that keeps the escape from being a footgun).
- `SPEC-KANBAN-BOARD-001` `design.md` §C — the sibling's ownership decisions this file defers to in §G.
