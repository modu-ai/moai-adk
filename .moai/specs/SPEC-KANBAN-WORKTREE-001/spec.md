---
id: SPEC-KANBAN-WORKTREE-001
title: "Per-card worktree lifecycle with holder liveness and mutual exclusion"
version: "0.3.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, lifecycle, orphan, stall-detection, holder, flock, branch-guard, liveness, pr-gate, import-direction, branch-token-match, operator-escape, creation-actor"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001]
related_specs: [SPEC-KANBAN-BOOTSTRAP-001, SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## HISTORY

- **v0.3.0** (2026-08-11) — Second plan-audit repair. An independent audit scored v0.2.0 at 0.79 against the Tier L threshold of 0.85, verified all seven v0.2.0 repairs and both follow-up corrections as genuinely closed with no regressions, and found every remaining defect delta-closable. Seven are repaired; five add requirements and two are amendments in place.

  **(1) `REQ-KW-017`'s rejection premise was falsified by a predicate this SPEC already adopts.** v0.2.0 refused every non-`gh` merge path on the ground that a merged-branch listing cannot report squash-merged branches. That is true of the bare `git branch --merged` the neighbouring fallback uses — and **false** of `IsBranchMerged`, which this SPEC names twice in its own `WorktreeManager` adoption and which is documented as reporting merge "irrespective of the merge strategy that placed them there", via an ordered OR whose **S4 is dedicated to squash-merge detection**, with **zero** `gh` usage in its package (`research.md` §D.4). The requirement is re-grounded on the argument that actually holds — a **per-branch** predicate cannot answer a **per-pull-request-identity** question — and the no-disposal outcome is re-decided on that ground rather than restored (§A.4.1).

  **(2) "the SPEC identifier the branch carries" was multi-valued and superstring-vulnerable.** Measured, three distinct branch names carry `SPEC-CODEX-PHASE2-001`, and **20 of 35** SPEC-carrying branch segments carry a suffix (`research.md` §C.1). v0.2.0 defined no match rule, so both idempotency matching and pull-request enumeration were ambiguous — and the predicate now has an **external consumer**, `SPEC-KANBAN-BOARD-001` `REQ-KB-020` reading a card's `status` from the branch it resolves. `REQ-KW-003` gains an exact-token match bounded by end-of-segment or `-`; new **`REQ-KW-019`** decides the multiple-match outcome (§A.2.1).

  **(3) Two terminal states had no escape and no owner.** `REQ-KW-011`'s unprobeable holder and `REQ-KW-012`'s dirty orphan were both left to "an explicit act" and "until a human clears it", with no operation required, no actor, and no observable end-state. New **`REQ-KW-020`** and **`REQ-KW-021`** give each a bounded operator-visible operation, mirroring the shape `REQ-KW-014` already carries (§A.7.2).

  **(4) The lock-clearing act had a check-then-unlink race.** Between reading the artifact's recorded process and unlinking it, the lock can be released by its owner and **re-acquired by a live process**; the clear then unlinks a valid lock. `REQ-KW-014` gains an identity-still-holds condition, and the argument is restated as **platform-asymmetric** rather than universal — measured, the Unix substrate releases its `flock` by closing the descriptor, so a killed holder there leaves a file and no lock, and `.moai/state/` currently holds **14** such inert artifacts, the oldest from 2026-05-30 (`research.md` §E.3) (§A.6.1).

  **(5) The extraction target was the package §C excludes.** `REQ-KW-018` sent branch derivation to `internal/worktree` on the strength of it being "the existing dependency-free leaf" — measured, that package's own `doc.go` declares it the **working tree state guard**, the L1 mechanism §C names so an implementer does not wire it by mistake. The supporting claim "the leaf both consumers already reach" was also false: `internal/kanban` does not exist. The target becomes `internal/core/git`, chosen after reading its `doc.go` rather than its name (§A.9).

  **(6) Worktree creation had no named actor and no serialization.** Disposal and release are lead-only with an explicit refusal; creation was "the system shall create". New **`REQ-KW-022`** names the lead and **`REQ-KW-023`** serializes creation beneath the card's existing lock (§A.3.1).

  **(7) `worker-` is unavailable as a worktree name prefix.** Not from the audit — found while repairing the bootstrap sibling, and independently recorded there. `cleanupMoaiWorktrees` runs unconditionally on every `moai cc` and removes worktrees whose base name begins with `worker-`. Measured, the prefix filter gates **both** scanned bases, so an L2 card worktree named for its SPEC identifier survives; the constraint is to keep it that way, and `REQ-KW-003` now carries it (§A.2.2).

  Two consumptions from siblings that reached v0.3.0 immediately before this revision. `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` now carries the **role declaration**, resolvable **by a session that is not the `lead`** — the clause exists for `REQ-KW-007` and `REQ-KW-011`, which both resolve the lead from outside it and previously deferred to `REQ-KS-004`, which does not carry it. Both now cite `REQ-KS-006` by name. And `SPEC-KANBAN-BOARD-001` `REQ-KB-023` hardened its own clearing act against repair (4) above; the same hardening is taken here at card scope.

  **A cross-SPEC defect is recorded rather than repaired, because its repair belongs to the other document.** `SPEC-KANBAN-BOARD-001` v0.3.0 promoted this SPEC from `related_specs:` to `dependencies:`, while this SPEC's frontmatter already names it — so the two now declare a **mutual dependency**, which is exactly the cycle §A.4 argues must never be declared. The asymmetry that resolves it is measured and stated in §A.4: this SPEC's need is a **landing** dependency (the holder field must exist before there is anything to release), while the board's need is a **contract** dependency (`REQ-KB-020` consumes `REQ-KW-003`'s identification *rule*, which is readable from this document without any code having landed). The declared edge therefore belongs here and the reverse belongs in `related_specs:` there. This frontmatter is left unchanged; the sibling's is the one to correct.

  **Budget.** 23 requirements and 23 criteria, against the Tier L ceiling of 25 and 25. Five requirements and five criteria are added; repairs (1), (4), (5) and (7) are amendments in place. Folding was available for each addition and refused for the reason this family has already paid for once — `SPEC-KANBAN-BOARD-001` v0.2.0 records a predecessor requirement that bundled a storage mechanism with an ownership rule, and the split deleted the rule along with the mechanism it had rejected.

- **v0.2.0** (2026-08-10) — Plan-audit repair. Two independent audits scored this SPEC at 0.70 and 0.78, both finding the structure sound and the defects local. Seven are repaired, and two of them change what a requirement demands.

  **(1) §A.6 and §A.7 contradicted each other, load-bearingly.** §A.6 argued a live session's tree is dirty; §A.7 rested the entire false-positive safety argument on that dirtiness — while §A.6 itself recorded that the tree is *clean* for the moment after each commit. A release firing in that window hands a live session's card to a second session, which is the outcome the section exists to prevent, and the card lock does not close it (a lock serializes lock holders, not tree occupants). **Requirement change**: cleanliness is no longer the liveness discriminant. `REQ-KW-009` and `REQ-KW-011` now require **positive evidence of death** — the age criterion is necessary and never sufficient — and the dirty-tree gate is demoted from the basis of the argument to defence in depth (§A.5, §A.7).

  **(2) The branch prefix was wrong, so the disposal gate could never fire.** Measured: `resolveSpecBranch` (`internal/cli/worktree/shared.go:32`) synthesizes `feature/<SPEC-ID>`, while the repository's branches run 63 `feat/` against 3 `feature/` as measured at v0.2.0 authoring time (re-measured at 64 against 3 at promotion time — the count drifts, the ratio does not; `research.md` §C). A gate keyed on the synthesized name matches almost nothing. `REQ-KW-003` now **observes** the branch a worktree reports and recognizes a card's branch by the SPEC identifier it carries rather than by its prefix (§A.2, §A.3).

  **(3) The named merge predicate could not express the required condition.** `branchMergedForCleanup` takes a branch name and a boolean and returns one bool; two pull-request identities are not recoverable from one branch name. **Requirement change**: `REQ-KW-007` now keys the gate on the card's **pull-request identities**, in the form `spec-workflow.md:437` already establishes (§A.4).

  **(4) The two merge-detection paths were described as equivalent and are not.** The `gh`-absent fallback is squash-merge blind by its own source comment, and this repository squash-merges — measured, 0 merge commits and 199 `(#N)`-carrying subjects across the last 200 first-parent commits on `origin/main`. New `REQ-KW-017` decides the blind case explicitly and names its cost (§A.4).

  **(5) Both named symbols are unexported, and one cannot be exported without a cycle.** New `REQ-KW-018` records the import-direction constraint and resolves each symbol differently (§A.9).

  **(6) A lost Windows lock left a card permanently stuck.** `internal/spec/lock_windows.go` has no stale-lock detection, and the recovery path is itself a holder change requiring the lock it needs cleared. `REQ-KW-014` gains an escape (§A.6).

  **(7) The `lead` role was a runtime dependency declared nowhere.** `REQ-KW-002` gains the gate and `REQ-KW-007` / `REQ-KW-011` the runtime refusal; the reason it is **not** a `dependencies:` entry is measured in §A.4.

  Two corrections to this document's own claims ride along. **v0.1.0's §A.6 asserted the predecessor carried a clean-tree assignment fence that this SPEC rejects. That is false** — `grep -rniE 'clean tree|clean-tree|working tree is clean|tree is clean|concurrent assignment'` over `SPEC-KANBAN-MULTISESSION-001` reports only orphan-classification hits, none an assignment fence. The lock of `REQ-KW-013` is a **new addition**, not a replacement, and §A.6 is rewritten accordingly. And **§A.5's "the peer registry is not a liveness signal in either direction" was over-broad**: its dead-PID entries defeat the *positive* reading only. Repair (1) consumes its *negative* reading, and the asymmetry is now stated rather than the whole registry excluded.

  Finally, `SPEC-KANBAN-BOARD-001` was promoted to Tier L and gained `REQ-KB-017` (the `lead` is the sole writer of board state), `REQ-KB-018` (write atomicity) and `REQ-KB-019` (a board-wide lock superseding card scope for board mutations, while preserving `REQ-KW-013` for holder assignment). Every place this SPEC discussed who may write board state now defers to `REQ-KB-017` (§A.6).

  **And this SPEC is promoted from Tier M to Tier L, in the same revision and for the same reason its sibling was.** Repairs (4) and (5) added `REQ-KW-017` and `REQ-KW-018`, taking the SPEC to 18 requirements and 18 criteria against a Tier M ceiling of 16 and 16. The governing rule — `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — reads an overage as a signal to tier up or split, never to relax the budget, so the tier is raised. **Folding the two additions back into `REQ-KW-007` and `REQ-KW-003` was considered and rejected**: bundling separable concerns into one requirement is precisely the failure this pass repaired in the sibling, where the predecessor's `REQ-KM-044` carried a storage mechanism and an ownership rule together and the split deleted the rule along with the mechanism it had rejected. At Tier L the SPEC sits at 18 of 25 and 18 of 25. No requirement or criterion is renumbered, retitled, or reworded by the promotion — it changes the tier metadata, replaces the now-stale over-budget disclosures in §B, `acceptance.md` §K and `progress.md` §E.1 with the Tier L position, and adds the two artifacts Tier L requires: `design.md` (the decisions, each with its rejected alternative) and `research.md` (the commands and observed outputs those decisions rest on). Tier L additionally raises the plan-auditor PASS threshold from 0.80 to 0.85.

- **v0.1.0** (2026-08-10) — Initial plan-phase authoring. Second of three SPECs split out of the superseded `SPEC-KANBAN-MULTISESSION-001` (59 requirements, 61 criteria), which failed a plan audit at 0.87 for structural reasons. This SPEC carries the per-card worktree lifecycle and holder liveness only. The predecessor's §A.10 (stall and death), §A.11 (worktree lifecycle), §B.10, §B.11 and §D.6 are its primary material. Five decisions change: the clean-tree assignment fence is **rejected** and replaced by an advisory file lock over an existing substrate (§A.6) — *corrected at v0.2.0: the predecessor carried no such fence, so the lock is an addition rather than a replacement*; the stall threshold's verification is **re-anchored** from a constant relationship to a configuration check, because the constant it was to be compared against does not exist (§A.5); the branch name and creation idempotency, both **unspecified** in the predecessor, are settled (§A.2, §A.3); and the disposal executor, **unnamed** in the predecessor, is named along with how merge state is observed (§A.4).

---

## §A. Context

### A.1 What this SPEC is, and what the other two are

The kanban system has three separable concerns. Splitting them is the point of this SPEC's existence: the predecessor tried to carry all three and became unauditable.

| SPEC | Owns |
|---|---|
| `SPEC-KANBAN-BOARD-001` | the six columns, the card record, the board state store, the column↔status check, the WIP limit, and the unheld state |
| **this one** | worktree creation and naming, per-card sharing, branch-guard invariance, disposal gating, refused-removal handling, orphan classification, stall detection, holder release, and mutual exclusion on holder mutation |
| `SPEC-KANBAN-BOOTSTRAP-001` | preflight, session roles and topology, bootstrap and the entry switch, configuration surfacing, quorum, the dispatch protocol, backend selection, and the coder-session internal chain |

This SPEC defines *where the work happens* and *when a card stops being held*. It defines neither *what a card is* nor *who dispatches it*. The boundary is restated as an exclusion (§C) so an auditor reading this SPEC alone does not read the omissions as gaps.

Two consumption relationships are load-bearing and are cited by requirement id rather than restated:

- The **card record** — including the holder field and the last-transition instant this SPEC's detector reads — is `SPEC-KANBAN-BOARD-001` REQ-KB-004. This SPEC adds no field to it.
- The **unheld state** this SPEC's holder release resolves to is that SPEC's REQ-KB-011: an unheld card in `run` is a legal steady state, not an error and not a stall. That is why no held or blocked column appears here.
- The **single-origin rule** — the board state lives beneath the primary checkout's `.moai/state/kanban/`, resolved through `git rev-parse --git-common-dir` — is its REQ-KB-005. Every holder read and write this SPEC performs is against that one file, from whichever worktree the session runs in.

Every identifier below is written in its **post-`SPEC-KANBAN-RENAME-001`** form. That rename is a `dependencies:` entry and a hard gate (REQ-KW-002).

### A.2 One worktree per card, at a path and a branch that two sessions compute alike

**Decided: the kanban system creates the worktrees itself**, rather than leaving provisioning to the operator's `moai cc -w` entry flag. This is deliberately the larger option, and its cost — a whole lifecycle rather than one line of bootstrap guidance — is carried here rather than absorbed into a happy path.

It is built **on** the repository's existing worktree doctrine, not beside it. `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary already defines the **L2 persistent SPEC worktree**: its path scheme `~/.moai/worktrees/<project>/<SPEC>/`, its lifetime, and its disposal contract. All three are adopted unchanged. No second worktree system is introduced, and the existing `WorktreeManager` of `internal/core/git` (`Add`, `List`, `Remove`, `Prune`, `DeleteBranch`, `IsBranchMerged`) is the mechanism rather than a new one.

**Scope is per card, not per session.** The card's `run`, `review` and `sync` sessions all work in the same tree. This is not a new rule: it is the existing SPEC-to-worktree mapping, in which run and sync reuse one worktree.

**The branch name was unspecified in the predecessor. v0.1.0 settled it on a synthesis, and the synthesis is wrong for this repository.** The predecessor wrote the creation command as `git worktree add -b <branch> <path> origin/main` without ever saying what `<branch>` was, so two sessions computing it independently had no guarantee of agreeing. v0.1.0 answered with `resolveSpecBranch` (`internal/cli/worktree/shared.go:32`), which returns `"feature/" + name` for a SPEC-shaped input. Measured now:

```
$ git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' \
    | grep -oE '^(feat|feature|fix|chore|docs|spec)/' | sort | uniq -c | sort -rn
  63 feat/
  30 docs/
  23 chore/
  18 fix/
   3 feature/
   2 spec/
```

**Sixty-three against three.** This is a real divergence in the repository, not a preference: the helper synthesizes the form that 3 branches use while 63 use the other. Anything keyed on the synthesized name matches almost nothing, which is how a disposal gate comes to be written, shipped, and never once fire.

The census above and the figures drawn from it throughout this document are **that command's output at v0.2.0 authoring time**, not a standing fact. Re-run at promotion time it reports **64** `feat/` against the same 3 `feature/` — one branch created in between (`research.md` §C). The figures here are deliberately left at their own measurement moment rather than chased, because the argument rests on the ratio's order of magnitude and the count is expected to drift; `plan.md` §C.8 re-measures it at preflight, and the rule `REQ-KW-003` adopts is prefix-independent and survives either number.

**The repair is to stop synthesizing wherever a name can be observed instead.** Two surfaces, treated differently, because only one of them has anything to observe:

- **At creation** a name must be produced, and the helper remains the producer — but as a **fallback synthesis**, explicitly labelled, not as the identity of the card's branch. Its failure mode is the measurement above: a tree created under this SPEC lands on `feature/`, against a repository convention of `feat/`, and any later consumer that re-derives the name rather than reading it inherits that mismatch.
- **Everywhere after creation** — idempotency matching (§A.3), orphan classification (§A.7), disposal (§A.4) — the branch is the one **the worktree actually reports** (`git worktree list --porcelain` emits `branch refs/heads/<name>` per entry). Nothing re-synthesizes it.

**Swapping the literal `feature/` for `feat/` was considered and rejected.** It trades one brittle synthesis for another and inverts the failure rather than removing it: the 3 surviving `feature/` branches would become the ones that never dispose. Worse, it would leave the design still asserting a prefix, when a prefix is not what identifies a card's branch.

**What identifies a card's branch is the SPEC identifier it carries**, not its prefix. `feat/SPEC-KANBAN-BOARD-001` and `feature/SPEC-KANBAN-BOARD-001` are both that card's branch; `feat/SPEC-OTHER-001` is not. That predicate is prefix-independent, survives a future convention change, and is the one §A.3 and §A.4 are written against.

Two concurrent `run` cards are two different SPEC identifiers, so neither the path nor the branch can collide. The **path** — deterministic from the SPEC identifier under the L2 scheme — is the card's identity throughout; the branch is an observed attribute of the tree found there.

### A.2.1 "Carries the identifier" is a match rule, and v0.2.0 did not state one

v0.2.0 replaced a wrong predicate with a right *kind* of predicate and left its degree unspecified. That is a real improvement and an incomplete one: the prefix form was wrong but **single-valued**, while "carries the SPEC identifier" is under-specified in two directions at once.

**Multi-valued.** Measured now:

```
$ git for-each-ref --format='%(refname:short)' | sed 's|^origin/||' \
    | grep -oE 'SPEC-[A-Z0-9-]+' | sort | uniq -c | sort -rn | head -5
   5 SPEC-CODEX-PHASE2-001-
   3 SPEC-TOKEN-001-
   3 SPEC-NAVIGATOR-SYNC-003
   2 SPEC-V3R5-STATUSLINE-FMC-001
   2 SPEC-V3R5-ATOMIC-WRITE-001
```

Those 5 occurrences resolve to **three distinct branch names** — `feat/SPEC-CODEX-PHASE2-001-run`, `docs/SPEC-CODEX-PHASE2-001-m0-close`, `docs/SPEC-CODEX-PHASE2-001-fork-resolution` — the other two being `origin/` mirrors of the same names (`research.md` §C.1). One SPEC identifier, three branches, and a card's phases landing under different type prefixes. So "the branch naming this card" is not a function: §A.3's fourth row asks whether such a branch exists, and §A.4's enumeration asks for all of them, and v0.2.0 gave neither an answer for the case where the count is greater than one.

**Superstring-vulnerable.** Nothing in v0.2.0 distinguished a *containment* test from a *token* test. Under containment, `SPEC-X-0010` matches `SPEC-X-001`, and so does anything else that merely embeds the string.

**Decided: an exact-token match, bounded by end-of-segment or a hyphen.** Strip the branch's `<type>/` prefix; the remaining segment names card X **iff** it begins with X's SPEC identifier and the next character is either nothing or `-`.

Three consequences, each deliberate:

- `feat/SPEC-X-001` and `feature/SPEC-X-001` both match. This is the prefix-independence §A.2 argues for.
- `feat/SPEC-X-001-run` matches, and **must** — measured, 20 of the 35 distinct SPEC-carrying branch segments in this repository carry exactly such a suffix (`-run`, `-wave5`, `-m0-close`, `-sync-sha-backfill`), and they are the card's own phase branches (`research.md` §C.1). A rule requiring equality would refuse the majority of the repository's real SPEC branches.
- `feat/SPEC-X-0010` does **not** match `SPEC-X-001`: the character after `001` is `0`, which is neither the end of the segment nor a hyphen. Note that `SPEC-X-0010` cannot itself be a valid SPEC identifier — the identifier grammar ends on exactly three digits — so this is a guard against arbitrary branch text rather than against a second card.

**The residual is named rather than glossed.** The hyphen boundary admits one collision the rule cannot see: a hypothetical `SPEC-X-001-EXTRA-002` is itself a valid SPEC identifier *and* matches card `SPEC-X-001`. Measured over the 31 SPEC identifiers currently appearing on branches, **no identifier is a hyphen-delimited prefix of another** (`research.md` §C.1), so the collision is structural rather than present. Where it does arise the card would find a branch that is not its own; §A.3's refusals and `REQ-KW-019`'s ambiguity refusal both fail in the safe direction, and `plan.md` §C.13 re-measures the prefix-freedom at preflight.

**Where more than one branch matches, nothing is picked.** This is `REQ-KW-019`, and it is scoped rather than global, because multiplicity is correct in one of the three places the predicate is used:

| Use | Expected cardinality | Outcome on more than one |
|---|---|---|
| Reading the branch an existing worktree reports (§A.3 rows 2-3, §A.7, §A.4) | exactly one, by construction — a worktree reports one branch | not reachable |
| Finding a branch by identifier where none is checked out (§A.3 row 4) | at most one is resolvable | **refuse and surface both**, modifying nothing |
| Enumerating a card's branches to discover its pull requests (§A.4) | one or more, legitimately | accept all — this is the normal case |

Picking one of several would be the same class of error as synthesizing a name: it produces an answer that is confidently wrong, and it is wrong silently. The refusal is scoped to the single-resolution use so that scoping it wrongly — refusing during enumeration — would break disposal for every card that has more than one branch, which is most of them.

**This predicate now has a consumer outside this SPEC.** `SPEC-KANBAN-BOARD-001` `REQ-KB-020` reads a card's frontmatter `status` from the card's branch, identifying that branch by this rule and explicitly declining to re-derive a name. An ambiguous resolution there is not a local inconvenience — it selects which branch's `status` the board believes — so the refusal is a property of the rule rather than of this SPEC's callers.

### A.2.2 The worktree's base name may not begin with `worker-`

Not an audit finding. It surfaced while repairing `SPEC-KANBAN-BOOTSTRAP-001`, which records the same constraint from its side and assigns it here.

Measured, `internal/cli/launcher.go`: `applyCCMode` calls `cleanupMoaiWorktrees` unconditionally on every `moai cc` launch (`:227`), and that function (`:481`) scans `git worktree list --porcelain`, keeps only entries whose **base name begins with `worker-`**, and removes each one that also lies beneath either `.claude/worktrees/` or a directory under `~/.moai/worktrees/`.

The second base is this SPEC's: the L2 scheme places a card's tree at `~/.moai/worktrees/<project>/<SPEC>/`. So the question is whether a running `moai cc` would destroy live card worktrees, and the measured answer is **no** — but for a reason that must be recorded, because it is a naming property this SPEC controls rather than a protection the launcher offers:

- The `worker-` filter is applied **before** the base-path loop, so it gates both bases equally. Directories under `~/.moai/worktrees/` are enumerated only as *containers to scan*, never removed themselves.
- A card worktree whose base name is its SPEC identifier therefore never matches the filter and is skipped entirely.
- Removal is additionally **non-force** (`removeWorktree` omits `--force` by a documented decision), so even a matching tree holding uncommitted work is kept and reported as kept.

The constraint is consequently a prohibition rather than a discovery: the card worktree's base name is the card's SPEC identifier and shall not begin with `worker-`. Adopting the `worker-` convention — which is what a reader coming from the team-worktree code would reach for — would place every card's tree inside the sweep radius of a command an operator runs routinely, and the loss would be silent for a clean tree. `REQ-KW-003` carries it.

### A.3 Creation is idempotent, and a mismatch is never silently adopted

The predecessor said nothing about what happens when the path or the branch already exists. That silence is the dangerous kind: the most convenient reading — reuse whatever is there — is exactly the reading that lets a card's session start writing into a tree belonging to something else.

Four conditions, four distinct outcomes:

| Condition | Outcome |
|---|---|
| Neither the path nor a branch naming this card exists | Create both. The ordinary path. |
| The path exists **and** the branch it reports names this card's SPEC identifier | **No-op success.** The tree is the card's, so a second creation call for the same card is satisfied by what is already there. |
| The path exists but the branch it reports does **not** name this card's SPEC identifier | **Named error, nothing modified.** The tree is not this card's. |
| A branch naming this card exists but the path does not | **A distinct named error, nothing modified.** A branch with no worktree is a state a human created, and adopting it silently would attach a card to a history it never chose. |

The matching condition is written against the **reported** branch and against the SPEC identifier it carries, per §A.2 — not against the synthesized name. Under the prefix-keyed form of v0.1.0, a tree a human created at the card's path on `feat/SPEC-…` fell into the third row and was refused as foreign, when it is in fact exactly the card's tree.

The two error conditions are distinguished rather than collapsed into one, because an operator resolving them does different things: a path-on-wrong-branch is a stale or misplaced tree, and a branch-without-path is prior work someone left behind. Neither outcome deletes, resets, or re-points anything — the refusal is the whole behavior.

Idempotency is what makes the lifecycle safe under the re-dispatch this SPEC's own recovery path produces. A card whose holder was released and whose tree is clean is re-dispatched *into that same tree* (§A.7); the re-dispatch calls creation again, and the second call has to be a no-op rather than an error, or recovery breaks on the mechanism meant to enable it.

The matching condition reads the branch by the exact-token rule of §A.2.1, and row 4 — a branch naming the card with no path — is the one place a *search* by identifier occurs, so it is the row `REQ-KW-019`'s ambiguity refusal binds to.

### A.3.1 Creation has an actor, and it is the same one the other two lifecycle acts have

**v0.2.0 named an actor for disposal and for release, and left creation to "the system".** That asymmetry is not cosmetic. Disposal is lead-only with an explicit refusal when no `lead` resolves (§A.4); release is lead-only for a second, independent reason — `SPEC-KANBAN-BOARD-001` `REQ-KB-017` makes the `lead` the sole writer of board state (§A.6). Creation was left unattributed, and the reading that fills the gap most naturally is the one that changes the concurrency: **a dispatched worker creating its own tree**.

Under that reading two sessions can call `git worktree add` for one path at the same time, and nothing in this SPEC serializes them. `REQ-KW-013`'s lock is scoped to holder mutation, §A.3's four-row table is a table of *sequential* observations, and neither sibling claims the decision — `SPEC-KANBAN-BOOTSTRAP-001` §C returns worktree creation, naming, and per-card scope to this SPEC by requirement id. So the gap was owned by nobody, which is the family's F1 failure shape.

**Decided: the `lead` is the creation actor**, consistent with disposal and release, and refusing in the same way when no session occupying the role is resolvable. The role is read through the declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, which is required resolvable from a session that is not the `lead` precisely so gates like this one can evaluate it. This is `REQ-KW-022`.

**Decided separately: creation is serialized beneath the card's own lock.** Naming the actor does not by itself close the race, and stating it as though it did is the mistake worth avoiding here. The `lead` role is occupied by a session, but the command surface is per-invocation — §A.6 already records that two `lead`-role invocations are two operating-system processes — so two concurrent invocations remain possible. `git worktree add` is not atomic, and its own refusal is not a substitute: a partially-created tree can be observed by the second caller between the directory appearing and the branch being reported, which is exactly the observation §A.3's rows are decided on. The card's existing advisory lock (`REQ-KW-013`) is widened in scope to cover the creation read-decide-create sequence as well as the holder read-decide-write; no new lock and no new mechanism. This is `REQ-KW-023`.

**Two alternatives were considered and rejected.**

- **Relying on `git worktree add`'s own refusal plus §A.3's idempotency.** It is nearly right, and that is the danger: a serialized race does resolve to either git's refusal or the no-op success of §A.3 row 2. What it does not cover is the interval during which the tree exists and its branch is not yet reportable, where the second caller reads a state none of the four rows describes. The lock is already present and already per-card, so the cheaper answer costs a correctness argument that has a hole in it.
- **A board-wide lock for creation.** Rejected on granularity: creating a card's worktree is a per-card read-decide-create and needs no board-wide invariant. `REQ-KB-019`'s board-wide scope exists for invariants a per-card lock cannot express, such as the WIP bound; borrowing it here would serialize every card's creation against every other's for nothing.

### A.4 Disposal has a named executor, and merge state is observed rather than assumed

**The gate is unchanged from the predecessor**: a card's worktree is disposed of only after **both** the run pull request and the sync pull request have merged. A card reaching `done` is not sufficient on its own — `done` is a board fact, and a merged pull request is a repository fact, and the second is the one the tree's disposability depends on.

**The executor was unnamed in the predecessor, and is named here: the lead session.** It evaluates the gate and it runs the removal; no worker session removes its own tree. The reason is not tidiness. A session removing the tree it is running inside is removing the ground under itself, and a session that has just finished its phase is the one least able to see whether the *other* pull request has merged yet. The lead is the only actor that observes both.

> **Boundary.** The `lead` role's definition, its election, its launch, and the topology it sits in belong to `SPEC-KANBAN-BOOTSTRAP-001` (`REQ-KS-004`). **How a session at runtime is known to occupy that role belongs to `REQ-KS-006`**, which since that SPEC's v0.3.0 requires each session to declare the role it occupies and requires that declaration **resolvable by a session that is not the `lead`** — a clause added for this SPEC, because both gates that read the role (the disposal refusal here and the release of §A.7) are evaluated from outside the lead. v0.2.0 deferred runtime occupancy to `REQ-KS-004`, which elects the role without defining how anyone reads it; that deferral is corrected here and in §A.7. This SPEC consumes the role name to say who acts; it defines no role, no session, and no carrier for the declaration.

**The `lead` is a runtime dependency, and v0.1.0 declared it nowhere.** Neither the frontmatter nor the preflight named it, so an implementation with no resolvable `lead` had no stated behavior at all — and the plausible default, letting whichever session noticed do the removal, is precisely what naming an executor was meant to forbid. `REQ-KW-002` gains the gate; `REQ-KW-007` and `REQ-KW-011` gain the runtime refusal.

It is **not** added to `dependencies:`, and the reason is measured rather than stylistic: `SPEC-KANBAN-BOOTSTRAP-001` already carries `dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]`, so declaring the reverse edge here would state a dependency cycle — and the landing order it implies (bootstrap before worktree) is the opposite of the one the siblings agree on. The in-family convention for exactly this shape is the one `SPEC-KANBAN-BOARD-001` uses for the same role: `SPEC-KANBAN-BOOTSTRAP-001` sits in `related_specs:` (where it already is), the role is cited by requirement id, and the dependency is discharged at **runtime resolution** rather than at landing.

#### A.4.0 A declared cycle now exists with the board sibling, and this document is not where it is repaired

The rule above was applied to the bootstrap edge and, at v0.3.0, is violated on the board edge — by the other document. Measured:

```
$ grep -n '^dependencies:' .moai/specs/SPEC-KANBAN-{WORKTREE,BOARD}-001/spec.md
SPEC-KANBAN-WORKTREE-001/spec.md:15:dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001]
SPEC-KANBAN-BOARD-001/spec.md:15:dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-WORKTREE-001]
```

Each names the other. That is the cycle this section forbids, stated twice, and it is new: the board sibling promoted this SPEC from `related_specs:` to `dependencies:` at its own v0.3.0, because its new `REQ-KB-020` consumes `REQ-KW-003` as the rule by which it identifies the branch a card's `status` is read from.

**Both consumptions are real, so the resolution is not to delete one — it is to notice they are of different kinds.**

- This SPEC's need of the board is a **landing** dependency. `REQ-KW-002` gates on the card record's holder field *existing*; until it does, the release path of `REQ-KW-011` has nothing to release and cannot be implemented at all. No amount of reading the board's document substitutes for the field.
- The board's need of this SPEC is a **contract** dependency. `REQ-KB-020` consumes an identification *rule* — observe the branch, match the SPEC identifier it carries, do not re-derive a name — which is fully readable from §A.2 and §A.2.1 of this document with no code of this SPEC's having landed. It is discharged by citation, exactly as this SPEC discharges the `lead`.

The declared edge therefore belongs on the landing dependency, which is the one already in this frontmatter, and the board's new entry is the one that should sit in `related_specs:` with `REQ-KW-003` cited by id. **This frontmatter is deliberately left unchanged**, and the correction is recorded here rather than performed, for the same reason the bootstrap edge was resolved by prose: a SPEC that unilaterally drops a dependency to break a cycle its sibling created leaves the family with an undeclared prerequisite and no record of why. The finding is surfaced to the orchestrator instead.

**Merge state is observed, never inferred. But the named helper cannot observe the condition this gate requires.** Measured at `internal/cli/session_worktree_prmerge.go:174-179`:

```go
// branchMergedForCleanup decides whether a branch is a cleanup candidate per
// REQ-SW-023. Primary path (gh available): state == "MERGED". Fallback path
// (gh absent): branch appears in `git branch --merged origin/main`. The
// fallback is squash-merge blind — squash-merged branches are NOT listed, so
// the worktree is preserved (documented via the on-entry blindness notice).
func branchMergedForCleanup(branch string, ghAvailable bool) bool {
```

It takes a branch **name** plus an availability flag and returns one bool. The gate above requires **two distinct pull-request identities** to be merged, and two identities are not recoverable from one branch name — so the two-PR condition is not expressible through this signature. v0.1.0 named this helper as the mechanism anyway, which made the requirement unimplementable as written.

**The gate is therefore keyed on pull-request identities, in the form the repository already establishes.** `.claude/rules/moai/workflow/spec-workflow.md:437` states the same pre-condition for the same act:

```
- Pre-condition: BOTH run PR AND sync PR are in MERGED state (verify via `gh pr view <PR>`)
```

`<PR>` there is a number, not a branch. The identities are **discovered from the repository, never synthesized**: the pull requests belonging to a card are enumerated by the card's SPEC identifier — over every branch matching the exact-token rule of §A.2.1, of which there are legitimately several, this being the one use where multiplicity is accepted rather than refused — and by the branch its worktree reports (`gh pr list`), and each discovered identity is then verified individually (`gh pr view <number> --json state`). The gate opens only when **at least two** pull requests were discovered for the card and **every** discovered one reports `MERGED`. Requiring two is what makes the sync pull request load-bearing — a card that has produced one merged pull request has run but not synced. Requiring *every* discovered one avoids classifying which is which by title or branch convention, and fails in the safe direction: an unrelated open follow-up holds the tree rather than releasing it.

Inferring merge from the card's column would make the board's own record the evidence for a fact the board does not own, and a board that is wrong about a merge would then authorize deleting a tree that still holds the only copy of something.

### A.4.1 Where the pull-request observation is unavailable, nothing disposes — and v0.2.0 argued that on a premise this SPEC's own adopted mechanism falsifies

**The outcome survives the repair. The argument does not, and it is the argument that was load-bearing.**

v0.2.0 refused every non-`gh` path on one ground: a merged-branch listing does not report squash-merged branches, and this repository squash-merges. The squash census is re-measured and stands — over the last 200 first-parent commits of `origin/main`, **0** merge commits and **199** of 200 subjects carrying GitHub's `(#N)` suffix (`research.md` §B). And the blindness claim is true of the thing v0.2.0 was looking at: `branchMergedForCleanup`'s `gh`-absent fallback is `git branch --merged`, whose own source comment names it squash-merge blind.

**It is false of a predicate this SPEC names twice in its own adoption.** §A.2 adopts the existing `WorktreeManager` "as the mechanism rather than a new one" and enumerates `IsBranchMerged` among its methods; §E cites the same list. Measured (`research.md` §D.4), `internal/core/git/worktree.go` documents that method as reporting merge

> irrespective of the merge strategy that placed them there

via an ordered OR over five signals whose **S4 is dedicated to squash-merge detection** — a synthetic-commit `git cherry` probe — conjoined with a state check, and it needs no `gh` at all: the whole package contains **zero** `gh` invocations. It is exported on the `WorktreeManager` interface, it is live (`internal/cli/worktree/clean.go` gates `moai worktree clean --stale` on it), and its own documentation calls it "a safety-critical predicate… a false positive destroys unmerged user work with no undo", carrying eight guards for the no-false-positive argument. A squash-aware, `gh`-free merge predicate was therefore sitting inside the mechanism this SPEC had already adopted, and the section that decided merge detection never mentioned it.

That is a defect of method rather than of outcome, and it is the more damaging kind: the conclusion happened to be right, so nothing would ever have contradicted it.

**Re-grounded: a per-branch predicate cannot answer a per-pull-request-identity question.** This is the argument §A.4 already establishes for `branchMergedForCleanup`, and it is strategy-independent — it holds however well a branch predicate handles squashes. The gate requires **at least two pull requests discovered** and **every one merged**, and the first half is what makes the sync pull request load-bearing. A merge predicate over branches cannot supply it:

- **Branch count is not pull-request count, in either direction.** Measured, three distinct branches carry `SPEC-CODEX-PHASE2-001` (§A.2.1); a branch can exist with no pull request ever opened on it; and successive pushes to one branch are one pull request, not several. So "two branches merged" neither implies nor is implied by "two pull requests merged".
- **`IsBranchMerged` answers a content question, not an identity one.** It reports whether a branch's changes survive in `base`'s tree. A card whose run work landed and whose sync pull request was never opened presents exactly the same answer as one that ran and synced on a single branch.

So the degraded path is re-examined on the new ground and still **not** adopted, and the reason is now falsifiable rather than inherited: it fails on arity, not on blindness. Were the gate's "at least two" clause ever re-expressed over branches, a degraded path would become available — and it is refused in advance, because the measurement above shows branch multiplicity per card running to three, so a branch-count proxy for a pull-request count would be wrong in the permissive direction.

The decision and its cost, restated:

- **Where the pull-request observation is unavailable, no disposal occurs, and the system says so.** Preserving is the safe direction — the alternative is deleting a tree whose merge state was never established.
- **No per-branch merge predicate is substituted for it** — neither the reachability-only `git branch --merged` listing nor the strategy-aware `IsBranchMerged`. The first is refused because it would *report that a merge check ran* while being structurally incapable of opening the gate in a squash-merging repository, which is worse than refusing. The second is refused because it answers a different question, and answering a different question confidently is the failure this revision exists to repair.
- **The cost is that worktrees accumulate and are never reclaimed**, silently, for as long as the observation is unavailable. That is why the refusal is surfaced rather than logged: the existing precedent in the same file is a once-per-invocation **on-entry blindness notice** (`session_worktree_prmerge.go:138`, with its test at `session_worktree_prmerge_test.go:227-250` asserting the notice documents the blindness), and this SPEC follows it.

This is `REQ-KW-017`.

> A note on this SPEC's own guard, recorded because the failure was inherited rather than invented. The falsified premise entered v0.2.0 from the prompt that commissioned that revision, and the document adopted it without measuring the mechanism it had itself already adopted two sections earlier. Every premise arriving with a repair instruction is measured before it is written down; §A.2.2 and §A.4.0 in this revision are two where the measurement changed the answer.

### A.5 The stall threshold keeps its default, loses its constant, and stops being sufficient on its own

**Decided: the card's last-transition instant ageing past a configured threshold makes a session a *candidate* for release — and nothing more.** No heartbeat field is added to the card; the field the detector reads is `SPEC-KANBAN-BOARD-001` REQ-KB-004's last-transition instant. What changed at v0.2.0 is the word *candidate*: at v0.1.0 the age criterion **decided** release on its own, and §A.7 shows why it cannot (the safety net it leaned on is open exactly when it is needed). Release now requires a second, positive condition, and §A.7 carries it.

**The peer registry's exclusion was over-broad, and is narrowed.** v0.1.0 wrote that the registry "is not a liveness signal in either direction". Its dead-PID entries defeat the **positive** reading — an entry existing does not establish that the session is alive, and that half stands. They do not defeat the **negative** reading: a recorded process that the operating system reports absent is evidence that the session is gone. The registry is therefore consulted, asymmetrically and for one purpose only — to resolve a holder's session identifier to the process identity that can be probed. `internal/session/registry.go:86-95` carries exactly the fields that makes possible (`SessionID`, `PID`, `Host`, `LastHeartbeat`); the repository already treats `kill -0`-style probing as the way a stale entry is distinguished from a live one, and this SPEC adds a consumer of that judgement rather than a mechanism.

`LastHeartbeat` is deliberately **not** the decider. A stopped heartbeat is compatible with two different worlds — a dead session and a live session whose heartbeat writer stalled — while an absent process is compatible with one. Corroboration only.

**The default is 6 hours (21600 seconds), and the number is not arbitrary.** The goal preset driving a coder session's internal chain is armed `--max-duration 14400` — four hours — so a card is *entitled* to run that long without a column transition, and any threshold at or below 14400s would fire on a perfectly healthy card at its own bound. 21600s leaves two hours of margin above that entitlement for the run-phase tail, while still catching a session that died overnight. An operator running longer cards raises it; a non-positive configured value is rejected with a named error rather than silently coerced.

**What changed is how that rationale is checked.** The predecessor's criterion required verifying that 21600 exceeds the preset's own 14400-second bound — that is, comparing two constants. Measured at authoring time:

```
$ grep -rn '14400' --include='*.go' internal/ pkg/     → exit 1, zero matches
$ grep -rn '14400' .claude/skills/moai/workflows/factory.md
  factory.md:96: … arms with `--max-turns 0 --max-duration 14400` — infinite turns, a four-hour wall clock …
```

**There is no second constant.** The four-hour bound exists only as a sentence in a markdown workflow file; nothing in Go declares it. A criterion asserting a relationship between two constants is therefore unachievable, and restating it would ship an acceptance criterion that no implementation can satisfy — the specific failure the predecessor's audit was meant to catch.

The replacement keeps the default and the reasoning, and moves the *check* to what can actually be decided: the shipped default equals 21600, and a non-positive configured value is refused by name rather than coerced. The relationship to the four-hour bound is stated in prose, cited to `factory.md`, and is **documentation-anchored, not constant-anchored**. That is a real weakness and it is named rather than glossed: should the preset's bound later change in prose, nothing mechanical will notice that 21600 no longer clears it. Making it mechanical would require the upstream bound to become a constant first, which is not this SPEC's to do.

### A.6 Mutual exclusion is a file lock, over a substrate that already exists

**This lock is a new addition. v0.1.0 said it replaced a predecessor fence, and that claim is false.** Measured over the whole predecessor:

```
$ grep -rniE 'clean tree|clean-tree|working tree is clean|tree is clean|concurrent assignment' \
    .moai/state/kanban-source/SPEC-KANBAN-MULTISESSION-001/
  → 2 hits, both orphan classification (plan.md:201, acceptance.md:159); no assignment fence
```

`SPEC-KANBAN-MULTISESSION-001` carried **no** exclusion mechanism for concurrent assignment at all — no clean-tree fence, no lock, nothing. So there is no inherited mechanism to reject and none to replace: everything below is added, and the gap it closes was open in the predecessor rather than mis-closed. The distinction matters for review, because "we replaced a weak fence with a strong lock" invites a comparison, while "there was no exclusion and now there is" invites the question of what else the split inherited as absent.

**The clean-tree fence is nevertheless rejected as a design**, because it is the shape a reader reaches for first and it does not work: a card's tree is dirty for most of a TDD cycle and clean for the instant after each commit, so such a fence would admit a second assignment exactly when a healthy session has just committed. That property is not merely a reason to prefer a lock — it is the same window that defeats the v0.1.0 liveness argument, and §A.7 is where it is actually closed.

**What is added is an advisory file lock held across the holder-mutation critical section** — read the card's holder, decide, write the holder — so that two sessions cannot interleave a read and a write over the same card.

**Who performs that write is not this SPEC's to decide, and since `SPEC-KANBAN-BOARD-001` v0.2.0 it is decided.** A card's holder is board state, and `REQ-KB-017` makes the session occupying the `lead` role its **sole writer** — enforced, not documented. Every holder mutation this SPEC describes is therefore a `lead` act; no worker session writes a holder, and this SPEC neither grants nor implies such a path. `REQ-KB-018` governs how that write reaches disk (same-directory temporary file, atomic rename), and this SPEC adds nothing to it.

`REQ-KB-019` places a **board-wide** lock across every board mutation's whole read-modify-write, and states that for a board mutation the board-wide scope **supersedes** card scope — while explicitly preserving this SPEC's card-scoped lock for holder assignment. The two are not redundant, and the boundary is worth stating precisely: the board-wide lock exists because a WIP bound cannot be enforced by two mutations each holding only their own card's lock; the card-scoped lock here exists because holder assignment is a per-card read-modify-write that must not interleave with another operation on **that** card. Cross-process is still required of it, because the `lead` is not guaranteed to be one long-lived process — a per-invocation command surface makes two `lead`-role invocations two operating-system processes.

**No new locking mechanism is designed, because two already exist here.** They are not interchangeable, and the difference decides which one is reused:

| Substrate | Unix | Windows | Cross-process? |
|---|---|---|---|
| `internal/lockfile/lockfile_{unix,windows}.go` | `syscall.Flock(LOCK_EX)`, blocking | **in-process `sync.Mutex` keyed by path** | **No** on Windows — the file's own comment records this as a deliberately preserved limitation ("do NOT silently upgrade this to LockFileEx"), acceptable there because its callers are tmux team workflows that are macOS/Linux-only |
| `internal/spec/lock.go` + `lock_{unix,windows}.go` | `unix.Flock(LOCK_EX\|LOCK_NB)`, returning a named held-error on contention | atomic create (`O_CREATE\|O_EXCL`) with a small transient-retry budget | **Yes**, on both |

Kanban sessions are separate operating-system processes. A Windows fallback that serializes only within one process would leave the exact race this requirement exists to close, while reporting success — so the reuse target is the **`internal/spec` per-scope lock pattern**, keyed per card rather than per SPEC-close. Its non-blocking contention behavior is also the right shape: a session that cannot take the lock should be told so immediately, not parked.

**The Windows counterpart is mandatory, and this is why it is stated rather than assumed.** Both existing substrates already ship a `_windows.go` file. A Unix-only `syscall.Flock` call added to a new package would not be a missing nicety — it would *regress cross-platform support this repository currently has*, and it would do so silently, because the build simply fails to compile on a platform nobody in this project builds on daily.

**What the lock does not cover, stated so it is not mistaken for coverage.** A lock is a statement about a critical section, not about a session. It says nothing about a session that died while holding a card: on Unix the kernel releases the flock at process death and the card's holder field stays populated regardless; on Windows the atomic-create lock file *persists* after the process is gone. In neither case does lock state answer "is this holder alive?". That question has a separate mechanism — the release predicate of §A.7 — and the two are never substituted for one another: the holder is never inferred from lock state, and the lock is never treated as a liveness signal.

### A.6.1 A lost Windows lock leaves the card stuck, and the recovery path needs the lock it must clear

The Windows asymmetry is not merely a note about durability. Measured, the header comment of `internal/spec/lock_windows.go`:

```
// Windows lacks fcntl-style advisory flock; we use atomic-create-file (O_CREATE|O_EXCL)
// per design.md §D.2 fallback. Stale lock detection (PID + timestamp embedded) is a
// post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring
// manual `del .moai/state/spec-close-*.lock`.
```

There is **no stale-lock detection on Windows**. A session that dies holding a card's lock leaves the lock file on disk, and the file is indistinguishable from a live hold.

**The defect is platform-asymmetric, and v0.2.0 argued it as though it were universal.** Measured (`research.md` §E.3), `internal/spec/lock_unix.go` holds `flock(2)` on an open descriptor and its own comment records that "close releases the flock atomically" — so the kernel releases the lock when a holder dies, and the release path never unlinks the file. The consequence is visible on disk right now: the primary checkout's `.moai/state/` holds **14** zero-length `spec-close-*.lock` artifacts, the oldest dated 2026-05-30, **every one of them inert**. A Unix developer therefore sees the artifact accumulate and never sees it block anything, which is precisely why this hazard is easy to ship: it is invisible on the machine it is written on and permanent on the machine it runs on. The `AC-KW-015` rows that judge it record the platform they ran on for the same reason.

On Windows the artifact persists *and* holds, and it is a deadlock with a specific and unpleasant shape:

> The recovery path of §A.7 is itself a **holder change**, and a holder change takes the card's lock. So recovery is blocked by exactly the artifact it exists to clear. The card is stuck — not until a timeout, not until the next dispatch, but until a human deletes a file by hand.

Inheriting the substrate's limitation silently would ship that deadlock into the one platform nobody in this project exercises daily, which is the same failure mode §A.6 already argues against for the in-process fallback.

**Decided: the lock gains an escape, and the escape is explicit rather than automatic.** The precedent the substrate's own comment names is stale detection via an embedded process identity and timestamp; the alternative is an operator-visible clearing act. The escape adopted here is the **conjunction**, because each alone has a failure this design already refuses elsewhere:

- Automatic stale detection alone would clear an artifact on inference, and inference about a live holder is exactly what §A.7 replaces with positive evidence.
- A manual clearing act alone leaves the card stuck until somebody notices, which is the state being repaired.

So: the artifact carries the identity of the process that created it; a lock whose recorded process is **positively observed absent** — the same probe §A.7 uses, not a timeout — is clearable; and the clearing is a bounded, operator-visible act that reports what it removed, never a silent step the acquire path takes on its own. `REQ-KW-014` carries it.

**The clearing act as v0.2.0 stated it has a check-then-unlink race, and the race admits two writers.** The operation reads the artifact's recorded process, probes it absent, then removes the artifact — three steps with two gaps. In either gap the lock can be **legitimately released by its owner and re-acquired by a live process**, and nothing in the sequence notices: the clear then unlinks a valid lock, and the holder that acquired it plus the next acquirer are both inside the critical section the lock exists to hold. Nothing in v0.2.0's `REQ-KW-014`, in `design.md` §E.2, or in `AC-KW-015` conditioned the removal on the artifact still being the one inspected, and none of `AC-KW-015`'s three rows was concurrent — so the defect was not merely unguarded, it was unobservable.

The window is not narrow in the way it first looks. The clearing act is invoked precisely when a card has been stuck, which is when an operator is most likely to be restarting sessions — so a re-acquisition landing between the probe and the unlink is not a pathological coincidence, it is the ordinary consequence of the situation that motivated the invocation.

**Decided: the removal is conditioned on the inspected artifact's identity still holding.** A recorded-identity re-read under the same handle, or a content or inode match against what was inspected — the mechanism is a run-phase choice, the condition is not. A mismatch **aborts the clear** and reports that it aborted; it does not retry, and it does not fall through to an unconditional removal. `SPEC-KANBAN-BOARD-001` `REQ-KB-023` carries the same hardening at board scope, having taken this requirement's shape and closed this hole in it; the scopes differ — board-wide there, per-card here — and the condition is identical.

The residual is named rather than glossed: where the recorded process cannot be probed at all — a foreign host — the artifact is not clearable by this mechanism and the manual deletion the substrate's comment describes remains the last resort. That is a narrower stuck state than the one being repaired, not its elimination.

### A.7 Death mid-card, and the tree it leaves behind

#### A.7.0 The contradiction v0.1.0 shipped, and what closes it

v0.1.0 held two claims that cannot both ground the same safety argument:

- **§A.6**: a live session's tree is *dirty* for most of a TDD cycle — this is why the clean-tree fence fails.
- **§A.6, one sentence later**: the tree is *clean* for the instant after each commit.
- **§A.7**: "a wrongly-released card whose session is in fact alive has a dirty tree, and a dirty tree stops the re-dispatch" — the entire justification for releasing on a timestamp alone.

The third claim is false during the window the second describes. Compose them and the failure is concrete: a healthy session whose card has not changed column for longer than the threshold — which is an *ordinary* state, since the threshold is measured against column transitions and a long `run` phase produces none — commits, and for that moment its tree is clean. If the release fires there, the card is released **and** classified clean **and** re-dispatched into the tree a live session is working in. Two sessions, one card: the exact outcome §A.7 exists to prevent.

**The card lock does not close it.** `REQ-KW-013` serializes the *holders of the lock* — the read-decide-write of a holder field. It says nothing about who is **occupying the tree**. Once the release has committed and the card has been handed to a second session, that second session takes the lock legitimately, in sequence, and walks into an occupied worktree. A lock is exclusion over a critical section, not over a working directory.

**Asserting the window is short would not close it either**, and is the repair this section refuses. A race whose probability is low is still a race, and this one is *biased toward firing*: the detector's threshold is compared against a quantity a long-running healthy card leaves untouched, so the population of cards it examines is enriched in exactly the healthy long-runners whose commits produce the clean window.

**Decided: tree cleanliness is not the liveness discriminant. Release requires positive evidence of death.** The predicate becomes a conjunction, and the age criterion is the cheap necessary half rather than the deciding one:

1. The card's last-transition instant has aged past the configured threshold (§A.5), **and**
2. The holder's process is **positively observed absent** — the holder's session identifier is resolved through the peer registry to a recorded process identity and host, and, where the host is this one, that process is probed and reports absent.

Both, or no release. A live session that has just committed satisfies (1) and fails (2), because it is a running process, so the window closes by construction rather than by timing.

**Where the holder cannot be probed** — no registry entry, or a recorded host that is not this one — **no automatic release occurs.** The card is surfaced to the operator as unprobeable, with the holder's identity and the reason, and the release becomes an explicit act somebody performs. Silently releasing an unprobeable holder would restore the original defect through the back door: absence of evidence read as evidence of death. What that explicit act *is* — who performs it, through what operation, and what state the card is in afterwards — was left unstated at v0.2.0 and is settled in §A.7.2.

**Two alternatives were considered and rejected.**

- **Confirming the discriminant across a bounded interval** — sampling the tree's cleanliness repeatedly and releasing only if it holds throughout. Rejected: it narrows the window without closing it. A session in a long test phase after a commit presents an unchanging clean tree for as long as the phase lasts, so *any* finite confirmation window still admits the false positive, and the design would have bought a polling loop in exchange for a smaller probability rather than a different answer.
- **A heartbeat field on the card.** Rejected, as at v0.1.0 and for the same reason: the write path dies with the session it is meant to report on, so a stopped heartbeat is ambiguous between a dead session and a stalled writer. It would additionally duplicate `internal/session/registry.go`'s existing `LastHeartbeat` while adding a board write path that `REQ-KB-017` reserves to the `lead`.

The dirty-tree gate below is **kept**, and is now what it should always have been: defence in depth behind a predicate that no longer needs it. It catches the residue — an unprobeable holder released by hand, a probe that answered wrongly — rather than carrying the argument.

#### A.7.1 The release itself

**Decided: when a session dies holding a card, the card's holder is released and its column does not change.** No new column is introduced, and no held or blocked column is invented — the state already exists, because an unheld card in `run` is the same steady state a WIP-2 board with one coder session produces (`SPEC-KANBAN-BOARD-001` §A.5, REQ-KB-011). One field, two causes.

**The half-finished working tree is the part that must not be silent.** A dead session leaves its worktree behind, and that tree may hold uncommitted work. A re-dispatch therefore never writes into an unexamined tree:

- **Clean orphan tree** (no uncommitted changes) — the card is re-dispatchable immediately, into that same tree, because the tree is the card's by construction (§A.2). This is the path §A.3's creation idempotency exists to serve.
- **Dirty orphan tree** (uncommitted changes present) — the card is unheld but **not dispatchable**. The orphaned tree's path and the released holder's identity are recorded and surfaced to the operator, and re-dispatch waits for a human. Overwriting someone's half-finished work to keep the board moving is the one outcome this design refuses.

**That human gate is defence in depth, and it is no longer load-bearing.** v0.1.0 argued it was what made a timestamp-only detector safe; §A.7.0 shows why that argument fails at the post-commit clean window, and the repair is in the release predicate rather than here. What the gate still buys is real but bounded: where a release does happen wrongly — an unprobeable holder released by hand, a probe that answered wrongly, a session whose process died while a child of it still writes — a dirty tree stops the re-dispatch before two sessions work one card, and does so without destroying anything. It is the second line, not the argument.

**Cleanup failure is a first-class path, not an error tail.** Removal can be refused because the tree is dirty (`internal/core/git` already returns a named dirty-worktree error rather than proceeding) or because a process still holds it. When that happens, the failure and the tree's path are recorded, surfaced, and that is where it stops. There is no escalation to a forced removal and no destructive retry. A refused removal has already told the truth: something is still in there.

#### A.7.2 The two states that had no way out

v0.2.0 left two cards stranded with no specified path back:

- An **unprobeable holder** (§A.7.0): release is "an explicit act somebody performs" — no operation required, no actor, no end-state.
- A **dirty orphan tree** (§A.7.1): re-dispatch is withheld "until a human clears it" — and "clears it" is undefined, so an implementation could satisfy the sentence by requiring the operator to hand-edit board state, or by requiring nothing at all.

Neither is a small omission. A card held by a session on a second machine cannot be probed by construction (§A.7.0's residual, restated in `design.md` §B.3 as accepted cost), so the unprobeable state is not an edge case — it is the *normal* state of every card in a multi-machine topology. Leaving it terminal makes the accepted cost unbounded rather than merely real.

**The family already knows the shape and applies it twice.** `SPEC-KANBAN-BOARD-001` `REQ-KB-013` mandates a bounded recovery "invoked as an explicit operator-visible act" so that its unknown state "is escapable rather than terminal", and this SPEC's own `REQ-KW-014` mandates exactly that for the lock artifact. The two states above got the same prose without the operation — which is how a design ends up sounding like it has an escape while having none.

**Decided: each gets a bounded, operator-visible operation with a defined observable end-state**, modelled on `REQ-KW-014` rather than invented.

**A force-release, for the unprobeable holder** (`REQ-KW-020`). It is gated on the holder actually being unprobeable: where the probe succeeds and reports the process **live**, the operation refuses. That gate is the whole difference between an escape and a footgun — an unconditional force-release is the age-only predicate of v0.1.0 with a human in the loop, and a human looking at a stuck board is not a reliable liveness oracle. The operation reports the card, the holder's identity, and the reason the holder was unprobeable; its end-state is the same one the automatic release produces — holder empty, column unchanged — after which the orphan classification of `REQ-KW-012` runs as it would have. It is invoked by the operator and never by the release path.

**An orphan-clear, for the dirty tree** (`REQ-KW-021`). Its end-state is defined so that "cleared" is observable rather than a matter of opinion: the recorded orphan-tree path and released-holder identity are removed from the card, and the card becomes re-dispatchable. What the operation explicitly does **not** do is touch the tree — it neither deletes, resets, stashes, nor commits the uncommitted work, because that work is the reason the gate exists. It records the operator's judgement that the tree has been dealt with; the dealing-with is the operator's, performed with ordinary tools. Reporting what it cleared is part of the operation, on `REQ-KW-014`'s precedent.

**Rejected — a single "recover this card" operation covering both.** Superficially tidier, and it collapses two different questions into one verb. The unprobeable-holder case asks *is this holder gone?* and is gated on a probe; the dirty-orphan case asks *has this work been dealt with?* and is gated on nothing a machine can check. One operation would have to be gated on the weaker of the two, and the weaker one is "nothing".

**Rejected — an automatic timeout on either.** Both would clear on inference, which §A.6.1 and §A.7.0 already refuse for the lock artifact and for the release predicate respectively. A card stuck loudly is the direction this design takes everywhere it has the choice.

### A.8 The branch guard is respected, and the respect is measured

`.claude/rules/moai/workflow/main-checkout-branch-guard.md` forbids branch-state mutation in the primary checkout — switch, checkout, branch, `reset --hard`, stash, rebase, merge. Automatic creation touches none of them: `git worktree add` leaves the primary checkout's `HEAD` exactly where it was, which is precisely why that doctrine's own § Procedure prescribes worktree creation as *the* way to work on another branch. The guard additionally exempts worktree paths from its deny.

The predecessor's technique for establishing this is carried verbatim, because the alternative is a subtly empty check: the primary checkout's branch and HEAD are read **before** creation and again **after**, and must be equal. Reasoning instead from the guard's forbidden-pattern list would establish only that `git worktree add` is absent from that list — which is a claim about the list, not about the tree.

Baseline recorded at authoring time, from inside this worktree:

```
$ git rev-parse --git-dir         → /Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban
$ git rev-parse --git-common-dir  → /Users/goos/MoAI/moai-adk-go/.git
$ git -C <primary> branch --show-current → main
$ git -C <primary> rev-parse --short HEAD → b59a8ba7d
```

One boundary, stated so an implementer does not wire the wrong thing: `.claude/rules/moai/workflow/worktree-state-guard.md` governs **L1** ephemeral subagent isolation, is dormant by default, and is **not** this mechanism. Its snapshot and verify primitives are no part of this lifecycle.

### A.9 Both reused symbols are unexported, and one cannot be exported at all

v0.1.0 named two existing symbols as the mechanisms this SPEC reuses. Neither is callable as written, and they fail differently — so one resolution does not cover both. Measured:

| Symbol | Location | Package | Exported? |
|---|---|---|---|
| `resolveSpecBranch` | `internal/cli/worktree/shared.go:32` | `worktree` (under `internal/cli/`) | no |
| `branchMergedForCleanup` | `internal/cli/session_worktree_prmerge.go:179` | `cli` | no |

**The second is the one that cannot be repaired by exporting it.** `internal/cli` is the command surface — `cmd/moai/main.go` imports it, and the kanban command this SPEC's work is reached through will live there, so the dependency must run `internal/cli` → `internal/kanban`. A `internal/kanban` that imported `internal/cli` to reach the helper would close that loop into an import cycle, and the compiler would refuse it. A requirement assuming the call is available is unimplementable as written, not merely awkward.

**The first is not blocked by a cycle.** Measured, `internal/cli/worktree`'s own transitive internal dependencies are `internal/foundation`, `internal/core/git`, `internal/tui`, `internal/tui/internal` and `internal/worktree` — it does not import `internal/cli`. That is the whole of what the conclusion needs, and the conclusion stands: exporting `resolveSpecBranch` would compile.

**A supporting claim made at v0.2.0 was wrong, and correcting it makes the edge worse rather than better.** That revision added that "nothing outside test files imports it from `internal/cli` either". Measured, that is false — `internal/cli` imports `internal/cli/worktree` from **production** code, in two places:

```
$ grep -n 'cli/worktree' internal/cli/root.go internal/cli/inventory.go
internal/cli/inventory.go:17:  wtroot "github.com/modu-ai/moai-adk/internal/cli/worktree"
internal/cli/root.go:14:      "github.com/modu-ai/moai-adk/internal/cli/worktree"
```

Both files declare `package cli`; neither is a test. Nothing about the cycle analysis moves — the edge that would cycle (`internal/cli/worktree` → `internal/cli`) is still absent in both directions of that measurement, so exporting still compiles. What changes is the weight of the *rejection* below: `internal/cli/worktree` is not merely a package that happens to sit beneath the command surface, it is a live production dependency **of** it. Pointing a domain package at it is therefore a firmer edge to avoid than v0.2.0's prose claimed, not a softer one.

**Resolved differently, because the constraints differ.**

- **Branch derivation → extract into a package both can import.** The target is **`internal/core/git`**. The SPEC-shape test and the derivation move there, and both `internal/cli/worktree` and `internal/kanban` call it. *Rejected alternative*: exporting the symbol in place and importing `internal/cli/worktree` from a domain package. It compiles today, but it points a domain package at a command-surface package and makes a future edge from `internal/cli/worktree` into kanban a cycle — buying nothing over the extraction.

#### A.9.1 The v0.2.0 target was the package §C excludes, selected on its name

**`internal/worktree` is the wrong home, and the way it was chosen is the finding.** v0.2.0 characterized it solely as "the existing dependency-free leaf" — a true and irrelevant property — and cited `doc.go:7`, which is the sentence recording that `internal/cli/worktree` consumes it. The citation is accurate. It is also one line below the sentence that decides the question. Measured, `internal/worktree/doc.go:1-5`:

```
// Package worktree provides working tree state guard primitives for the MoAI
// orchestrator. It captures Snapshots of working tree state, computes Divergence
// between pre/post states, logs divergences to .moai/reports/worktree-guard/,
// and writes SuspectFlags when an Agent(isolation: "worktree") response shows
// an empty worktreePath.
```

That is the **L1 worktree state guard** — and §C excludes it by name: "L1 ephemeral subagent isolation and the worktree state guard's snapshot and verify primitives… a different mechanism entirely; it is named here only so an implementer does not wire it by mistake." So v0.2.0 selected, as the home for an L2 naming concern, the implementation of the very mechanism this SPEC warns implementers away from — and no section noticed the collision.

A second supporting claim was false in the same paragraph: "the leaf both consumers already reach". Measured, `ls internal/kanban` reports no such directory; the package does not exist, so one of the two consumers reached nothing.

**The target is `internal/core/git`, and it was chosen by reading its `doc.go` rather than its name** — the discipline whose absence produced the previous answer. Three measurements decide it (`research.md` §D.5):

- Its `doc.go` declares the package's subject as "Git repository operations", implementing "BranchManager: branch lifecycle and conflict detection". A SPEC-identifier-to-branch-name derivation is branch naming, which is that subject rather than adjacent to it.
- `go list -deps ./internal/core/git` reports exactly one internal dependency besides itself, `internal/foundation`. It imports neither consumer, and it cannot import `internal/kanban`, so no edge this SPEC adds can cycle.
- `internal/cli/worktree` **already** imports it (`research.md` §D.3), so the consumer that holds the symbol today reaches the new home without a new dependency.

It is additionally where this SPEC's other reused mechanism already lives: `WorktreeManager` and `IsBranchMerged` are in the same package (§A.2, §A.4.1), so the extraction consolidates rather than scatters.

*Rejected — keeping `internal/worktree` on the grounds that it is a leaf and the move is cheap.* Cohesion is the cost being weighed, not compilation. Placing branch naming inside the L1 state guard would make §C's exclusion unreadable to the next implementer, who would find this SPEC's own code in the package this SPEC says is not its mechanism.

*Rejected — a new package created for the derivation.* It would be a third worktree-adjacent package to reason about, for one function that has an existing home.
- **Merge observation → this SPEC owns and implements the contract itself.** There is no extraction worth doing, and the reason is §A.4 rather than the import graph: `branchMergedForCleanup`'s signature cannot express the two-identity condition, so extracting it would relocate a helper that still fails to answer the question. What is reused is the *form* — `gh pr view <PR> --json state`, per `spec-workflow.md:437` — not the function. *Rejected alternative*: lifting the helper into a shared package and widening its signature there. Rejected because it changes an existing caller's contract (`session_worktree_prmerge.go`, whose own `REQ-SW-023` semantics are branch-keyed and deliberately squash-blind) to serve this SPEC's different condition, which is a modification of a surface §H declares out of scope.

The constraint itself is recorded as `REQ-KW-018` so that a later reader does not re-propose exporting `branchMergedForCleanup` and rediscover the cycle in the compiler.

---

## §B. Requirements (GEARS)

> Requirement count: 23 (`REQ-KW-001` … `REQ-KW-023`). **Tier L: 23 of 25 requirements, 23 of 25 acceptance criteria — within budget, with two of each remaining.** The v0.3.0 additions are `REQ-KW-019` (the multiple-match refusal), `REQ-KW-020` and `REQ-KW-021` (the two operator escapes), and `REQ-KW-022` and `REQ-KW-023` (the creation actor and its serialization); `REQ-KW-003`, `REQ-KW-014`, `REQ-KW-017` and `REQ-KW-018` are amended in place. Nothing is renumbered, retitled, or reworded except where a repair required it.
>
> Folding was available for every one of the five and refused each time, on the precedent this family has already paid for: `SPEC-KANBAN-BOARD-001` v0.2.0 records a predecessor requirement that bundled a storage mechanism with an ownership rule, and the split then deleted **both** when only one was rejected. The specific temptations, recorded so the next reviewer does not re-propose them for free — `REQ-KW-019` into `REQ-KW-003` (a match *rule* and an *ambiguity outcome* are separable: an implementation can define the first and silently pick under the second); `REQ-KW-020` and `REQ-KW-021` into one recovery verb (§A.7.2 rejects it — the two are gated on different things, and one verb must take the weaker gate); `REQ-KW-023` into `REQ-KW-022` (naming an actor does not serialize that actor's concurrent invocations, and §A.3.1 shows the interval it leaves open). Where the ceiling had bitten, the finding would have been reported rather than a twenty-sixth requirement written.
>
> The Tier L artifact set is unchanged: `design.md` and `research.md` accompany this file, both revised at v0.3.0.

### B.1 Preconditions

**REQ-KW-001** — The implementation shall write every renamed identifier in its post-`SPEC-KANBAN-RENAME-001` form, and shall introduce no occurrence of `factory` in any identifier, path, environment variable, sentinel, preset name, or prose it authors.

**REQ-KW-002** — The implementer shall verify at preflight that both landing prerequisites have landed on the base branch — the rename, and the board model this SPEC consumes — and **when** either the renamed package or the card record's holder field is found absent, the implementer shall halt and surface the absence rather than proceeding against an unlanded prerequisite or supplying the missing part itself; and the implementer shall additionally record the `lead` role, defined by `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004`, as a **runtime** dependency of the disposal and release paths, resolved at the point those paths run rather than gated at landing, because the bootstrap sibling declares this SPEC among its own `dependencies:` and the reverse declaration would state a cycle.

### B.2 Creation, naming, and per-card scope

**REQ-KW-003** — The system shall create each card's worktree at a path derived deterministically from the card's SPEC identifier using the existing L2 persistent-worktree path scheme, that path being the card's identity throughout, and that path's final segment shall be the card's SPEC identifier and shall **not** begin with `worker-`, because the Claude-only launch path removes every worktree so named beneath either scanned base on every invocation; **where** a branch name must be produced because none yet exists, the system shall obtain it from the repository's existing SPEC-to-branch derivation as a **fallback synthesis** and shall record that the synthesized prefix diverges from the repository's dominant branch convention; and every decision taken about an existing worktree's branch — matching, classification, and disposal — shall read the branch that worktree reports rather than re-deriving a name, shall recognize a branch as the card's **when and only when** the branch name's segment following its type prefix begins with that card's SPEC identifier and the character immediately after it is either absent or a hyphen, so that a phase-suffixed branch is recognized and a branch merely embedding the identifier as a longer token is not, and shall introduce no worktree mechanism and no branch-naming scheme parallel to the ones the repository already defines.

**REQ-KW-019** — **Where** the system must resolve a single branch for a card by searching on the identifier rather than reading the branch an existing worktree reports, and more than one branch satisfies the match rule of REQ-KW-003, the system shall refuse with a named error, shall surface every matching branch, and shall modify nothing — it shall not select one, and shall not prefer one by prefix, recency, or any other tiebreak; and this refusal shall bind only that single-resolution use, the enumeration of a card's branches for pull-request discovery under REQ-KW-007 accepting more than one match as its ordinary case.

**REQ-KW-004** — **When** creation is requested for a card whose worktree path already exists and whose reported branch names that card's SPEC identifier, the system shall succeed as a no-op; **when** the path exists and its reported branch does not name that identifier, and **when** a branch naming that identifier exists without the path, the system shall refuse each condition with its own distinct named error, modifying nothing; and in no case shall the system adopt, re-point, reset, or delete a mismatched tree or branch.

**REQ-KW-022** — The session occupying the `lead` role shall be the sole actor that creates a card's worktree, no dispatched worker session creating its own, and **when** no session occupying that role is resolvable — the occupancy being read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, which that requirement makes resolvable from a session that is not the `lead` — the system shall refuse the creation and surface the absence rather than performing it as another session.

**REQ-KW-023** — The system shall serialize a card's creation sequence — the observation of the path and of the branch, the decision among the four outcomes of REQ-KW-004, and the creation itself — beneath the card-scoped lock of REQ-KW-013, whose scope shall accordingly cover creation as well as holder mutation, so that two concurrent invocations occupying the `lead` role cannot both create for one card; and the system shall not rely on the underlying worktree command's own refusal in place of that serialization, since a partially created tree is observable between the path appearing and the branch becoming reportable, which is a state none of REQ-KW-004's four outcomes describes.

**REQ-KW-005** — Exactly one worktree shall exist per card, and the card's `run`, `review`, and `sync` sessions shall all work within it, so that a worktree is never created per session.

### B.3 Branch-guard invariance

**REQ-KW-006** — Worktree creation shall leave the primary checkout's branch and HEAD unchanged, shall perform none of the branch-state mutations the primary-checkout branch guard forbids, and this invariance shall be established by comparing the primary checkout's branch and HEAD before and after creation rather than by reasoning from the guard's forbidden-pattern list.

### B.4 Disposal and refused removal

**REQ-KW-007** — The lead session shall be the sole actor that evaluates a card's disposal gate and performs the removal, and **when** no session occupying the `lead` role is resolvable — the occupancy being read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, which that requirement makes resolvable from a session that is not the `lead` — the system shall refuse the disposal and surface the absence rather than performing it as another session; the gate shall be keyed on the card's **pull-request identities**, which the system shall discover from the repository — enumerated over every branch satisfying REQ-KW-003's match rule and over the branch its worktree reports — and shall verify individually in the form `spec-workflow.md` § Sync to Cleanup already prescribes, opening only **when** at least two pull requests were discovered for the card and every discovered one is observed merged; the gate shall not be decided by any predicate whose subject is a branch, since a pull-request count is not recoverable from a branch set in either direction; and a card reaching `done` shall not by itself trigger disposal, nor shall merge state be inferred from the card's column.

**REQ-KW-017** — **Where** the pull-request observation of REQ-KW-007 is unavailable, the system shall perform no disposal and shall emit a notice once per invocation recording that disposal is suspended and that worktrees will therefore accumulate unreclaimed; and the system shall substitute **no per-branch merge predicate** for the unavailable observation — neither a reachability-based merged-branch listing, which additionally does not report squash-merged branches in a repository that squash-merges and would therefore report that a merge check ran while being structurally incapable of opening the gate, nor a strategy-aware merge predicate that does report them, which answers whether a branch's changes survive in the base rather than how many pull requests the card produced and merged, and so cannot supply the at-least-two condition that makes the sync pull request load-bearing.

**REQ-KW-008** — A worktree holding uncommitted work shall not be removed, and **when** a removal is refused — because the tree is dirty or because a process holds it — the system shall record the failure with the tree's path, surface it to the operator, and stop; it shall not escalate to a forced removal and shall not retry the removal destructively.

### B.5 Stall detection

**REQ-KW-009** — The system shall judge a session a **candidate** for release solely by the age of the card's last-transition instant, that age being necessary and never sufficient, and shall add no heartbeat field to the card record; it shall not treat the presence of a peer-registry entry as evidence that a session is alive, and shall consult that registry for one purpose only — resolving a holder's session identifier to the recorded process identity and host that the release predicate of REQ-KW-011 probes; and it shall not judge liveness by whether a card's working tree is clean.

**REQ-KW-010** — The stall threshold shall be configurable, shall ship a default of 21600 seconds, and **when** a configured value is non-positive the system shall reject it with a named error rather than coercing it; the default's relationship to the coder chain's four-hour wall-clock bound shall be recorded in prose citing that bound's documented location, because no constant declaring it exists to compare against.

### B.6 Holder release and orphan classification

**REQ-KW-011** — The system shall release a card's holder only **when** both conditions hold — the age criterion of REQ-KW-009 has fired, **and** the holder's recorded process has been positively observed absent on the host that record names — and **where** the holder cannot be so probed, whether because no registry entry resolves it or because the recorded host is not this one, the system shall perform no automatic release but shall surface the card, the holder's identity, and the reason it is unprobeable, the escape from that state being the operation REQ-KW-020 requires; the release, being a write of board state, shall be performed by the session occupying the `lead` role per `SPEC-KANBAN-BOARD-001` `REQ-KB-017` and by no other, that occupancy being read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`; and a release shall leave the card's column unchanged, introducing no additional column and no held or blocked column for this case.

**REQ-KW-020** — The system shall provide a bounded force-release operation that releases the holder of a card REQ-KW-011 left unprobeable, that operation being invoked explicitly by an operator and never by the release path, reporting the card, the holder's identity, and the reason the holder was unprobeable, and leaving the card in the same observable state an automatic release produces — holder empty, column unchanged, with the orphan classification of REQ-KW-012 then applying; and **when** the holder's recorded process can be probed and is observed **live**, the operation shall refuse and modify nothing, so that it cannot be used to release a running session's card.

**REQ-KW-012** — **When** a card's holder is released, the system shall classify the orphaned worktree; **where** it is clean the card shall be immediately re-dispatchable into that same tree; and **where** it holds uncommitted work the system shall record the tree path and the released holder's identity, surface both to the operator, and withhold re-dispatch until the operation REQ-KW-021 requires is invoked, so that half-finished work is never silently overwritten.

**REQ-KW-021** — The system shall provide a bounded orphan-clear operation that resolves a card REQ-KW-012 withheld, that operation being invoked explicitly by an operator and never automatically or on the age of anything, reporting what it cleared, and having as its observable end-state that the card's recorded orphan-tree path and released-holder identity are removed and the card is again re-dispatchable; and the operation shall not delete, reset, stash, commit, or otherwise alter the orphaned tree's uncommitted content, recording the operator's judgement that the tree has been dealt with rather than dealing with it.

### B.7 Mutual exclusion on holder mutation

**REQ-KW-013** — The system shall serialize holder mutation for a card — the read of the holder, the decision, and the write — and, per REQ-KW-023, that card's creation sequence, beneath an advisory file lock scoped to that card, shall obtain that lock by reusing the repository's existing cross-process per-scope lock pattern rather than adding a third locking mechanism, shall ship a platform counterpart for every platform the reused pattern already supports, and shall not gate concurrent assignment on the working tree being clean; this card scope shall stand alongside, and not in place of, the board-wide lock that `SPEC-KANBAN-BOARD-001` `REQ-KB-019` holds across a board mutation's whole read-modify-write.

**REQ-KW-014** — The system shall not treat lock state as a liveness signal: it shall not infer a card's holder from the presence or absence of a lock, and **when** a lock artifact outlives the process that created it the system shall not read it as evidence of a live holder, holder release remaining governed by the predicate of REQ-KW-011; and because the reused pattern's Windows implementation performs no stale-lock detection — an asymmetry, the Unix implementation releasing its lock when the holding process dies and leaving only an inert file, so that an artifact left by a dead holder blocks the very holder change that recovers the card on one platform and on one platform only — the system shall record in each lock artifact the identity of the process that created it and shall provide a bounded clearing operation that removes such an artifact **only when** that recorded process is positively observed absent by the probe of REQ-KW-011, the clearing being an explicit operator-visible act reporting what it removed and never a step the acquire path takes on its own; and the removal shall additionally be conditioned on the artifact still being the same artifact that was inspected, so that a lock legitimately released and **re-acquired by a live process** between the inspection and the removal aborts the clear and is reported as aborted rather than being unlinked, which would admit two holders to the critical section the lock exists to hold.

### B.8 Reuse without an import cycle

**REQ-KW-018** — The system shall not reach a reused behaviour by importing the command-surface package `internal/cli` from the kanban package, because the command surface imports the kanban package and the reverse edge would be an import cycle; **where** a reused behaviour lives in an unexported symbol there, the system shall either extract that behaviour into a package that imports neither and whose documented subject the behaviour belongs to — the branch derivation being so extracted, into the repository-operations package whose declared subject is branch lifecycle, which imports neither consumer and which the command-surface worktree package already imports, and **not** into the worktree state-guard package excluded by §C, whose leaf status is a property of its dependencies rather than of its purpose — or shall implement the behaviour as a contract this SPEC owns, which is the resolution for merge observation, whose existing predicate cannot express the two-identity condition of REQ-KW-007 and is therefore not extracted; and the system shall modify neither existing caller's contract in the course of either resolution.

### B.9 Mirror, neutrality, and verification

**REQ-KW-015** — The implementer shall edit template source under `internal/template/templates/` before its local counterpart, shall run `make build`, and shall commit the regenerated `internal/template/catalog.yaml`; **while** applying a change to a mirrored pair, shall preserve that pair's measured relationship, so that a pair measured byte-identical remains byte-identical and a sanitized pair retains exactly the content its template side strips; and no file authored or modified under `internal/template/templates/` shall contain a SPEC identifier, a REQ or AC token, an internal date, or a commit SHA.

**REQ-KW-016** — The verification shall run the full test suite rather than an affected-packages subset, because a prior run-phase in this repository missed a cross-cutting template guard by testing narrowly.

---

## §C. Exclusions

### Out of Scope — the board sibling

- The board state store, its single-origin resolution rule, the card record's shape, and the column enumeration. All belong to `SPEC-KANBAN-BOARD-001` and are consumed here by requirement id (REQ-KB-004, REQ-KB-005, REQ-KB-011), never redefined.
- **Who is permitted to write board state, how that write reaches disk, and the exclusion covering a whole board mutation.** `REQ-KB-017` (the `lead` as sole writer), `REQ-KB-018` (same-directory temporary file plus atomic rename) and `REQ-KB-019` (the board-wide lock, which supersedes card scope for board mutations while preserving REQ-KW-013 for holder assignment) are that SPEC's. This SPEC's holder mutations are `lead` acts **because** of REQ-KB-017; it grants no write path of its own and states no atomicity rule.
- The WIP limit, admission into `run`, and the column↔status compatibility check. This SPEC releases a holder; it never moves a card between columns and never decides whether a column will admit one.
- The definition of the unheld state. This SPEC's release resolves *to* that state; the state itself is REQ-KB-011.

### Out of Scope — the bootstrap sibling

- Preflight beyond the two prerequisite gates of REQ-KW-002, session roles and topology, the bootstrap surface and the entry switch, configuration surfacing and its `moai init` question, the quorum bound, the dispatch protocol, backend selection, and the coder-session internal chain. All belong to `SPEC-KANBAN-BOOTSTRAP-001`.
- The definition, election, and launch of the `lead` role (`REQ-KS-004`), **and the carrier of the role declaration through which a non-lead session reads who occupies it** (`REQ-KS-006`, which fixes the contract and explicitly leaves the carrier a run-phase decision). REQ-KW-007, REQ-KW-011 and REQ-KW-022 name the role as the actor and consume the declaration; none defines either.
- The operator interface through which the escapes of REQ-KW-020 and REQ-KW-021 are invoked. This SPEC requires the operations, their gates, and their observable end-states; the surface they are reached through — a subcommand, a flag, a prompt — belongs with the bootstrap sibling's command surface.
- The dispatch that follows a re-dispatchable card. REQ-KW-012 decides *whether* a card may be re-dispatched; who sends the dispatch, and by what message, is the bootstrap sibling's.

### Out of Scope — worktree mechanisms this SPEC does not build

- A second worktree system beside the existing L2 persistent-worktree scheme. The path scheme, lifetime, and disposal contract are adopted whole (REQ-KW-003).
- L1 ephemeral subagent isolation and the worktree state guard's snapshot and verify primitives. That guard is dormant by default and is a different mechanism entirely; it is named here only so an implementer does not wire it by mistake. **This exclusion binds the extraction of REQ-KW-018 as well**: the state guard's implementation package is not the home for this SPEC's branch derivation, and v0.2.0's selection of it — on the strength of its dependency count rather than its documented purpose — is corrected in §A.9.1.
- A third locking mechanism. REQ-KW-013 reuses an existing cross-process pattern; neither existing lock package is rewritten, and the in-process Windows fallback of the other one is not "upgraded". The clearing operation of REQ-KW-014 is a consumer of the reused pattern's artifact, not an amendment to that package.
- A forced or retrying removal path. Its absence is the decision of REQ-KW-008, not a deferred feature.
- **A widened `branchMergedForCleanup`.** Its branch-keyed, deliberately squash-blind semantics serve an existing caller with its own requirement; REQ-KW-018 implements this SPEC's condition rather than modifying that contract (§A.9).
- **Substituting any per-branch merge predicate where the pull-request observation is unavailable.** Excluded by REQ-KW-017 rather than deferred, and the exclusion covers **two** candidates rather than one. A reachability-based merged-branch listing is refused because it cannot open the gate in a squash-merging repository while appearing to check. A strategy-aware predicate that *does* recognize squash merges — `IsBranchMerged`, which this SPEC adopts as part of the same `WorktreeManager` and which needs no `gh` — is refused for a different and stronger reason: it answers whether a branch's content survives in the base, which is not the number of pull requests a card produced (§A.4.1). v0.2.0 refused the first and never considered the second, on a premise the second falsifies.
- **Re-expressing the gate's at-least-two condition over branches rather than pull requests**, which would make a degraded path available. Refused in advance: measured, one card carries three branches (§A.2.1), so a branch count is not a pull-request count and the proxy would be wrong in the permissive direction — the direction that deletes trees.

### Out of Scope — liveness mechanisms considered and rejected

- A heartbeat field on the card, and any per-session liveness ping. Rejected: a heartbeat's write path fails together with the session it reports on, so a stopped heartbeat is ambiguous between a dead session and a stalled writer — and it would duplicate the `LastHeartbeat` the peer registry already records while adding a board write path REQ-KB-017 reserves to the `lead`.
- **Reading a peer-registry entry's existence as evidence that a session is alive.** Rejected on the measured ground that the registry holds dead-PID entries. Its *negative* reading is **not** excluded and is consumed by REQ-KW-011 — this narrows v0.1.0's "not a signal in either direction", which was over-broad.
- Treating a lock as a liveness signal, in either direction (REQ-KW-014). Note the converse is now in scope: a lock artifact's recorded process is probed in order to **clear** the artifact, which is a statement about the artifact and never about the card's holder.
- Automatic stale-lock clearing on a timeout or on age alone. Rejected: it would clear on inference, which is what REQ-KW-011 replaces with positive evidence.
- **Judging liveness from tree cleanliness, in any form** — an assignment fence, a release condition, or a confirmation sampled across a bounded interval. Rejected at v0.2.0 on the ground argued in §A.7.0: the tree is clean for the moment after each commit, and a finite confirmation window does not close that hole for a session in a long post-commit phase. v0.1.0's claim that this fence was *inherited from the predecessor and replaced* is false and is corrected in HISTORY.

### Out of Scope — the four-hour bound upstream

- Promoting the coder chain's four-hour wall-clock bound to a constant so the stall default could be checked against it mechanically. That bound lives in a workflow document belonging to another SPEC's surface; REQ-KW-010's documentation anchoring is the consequence, and the resulting weakness is named in §A.5 rather than hidden.

---

## §D. Verification surfaces

### D.1 Creation idempotency needs its mismatch rows (REQ-KW-004)

Asserting only the no-op row leaves the requirement's whole point untested: an implementation that adopts *any* existing tree passes a no-op-only test perfectly. All four conditions of §A.3 are therefore exercised, and each refusal is checked twice over — that the named error is the one for that condition, and that the tree and branch are byte-unchanged afterwards. A refusal that fires while having already re-pointed something is not a refusal.

### D.2 The lock is checked for contention, not for existence (REQ-KW-013)

That a lock is *taken* is trivially observable and proves nothing. The criterion is a contention test: a second holder-mutation attempt against the same card, from a **separate process**, must be refused while the first holds the lock, and must succeed once it is released. A same-process test would pass against the in-process Windows fallback this SPEC deliberately does not reuse, which is exactly the failure the substrate choice exists to avoid.

The cross-platform half is a file-existence claim over the new package — a platform counterpart exists for every platform the reused pattern already supports — paired with the reused package's own counterpart as the reference. The absence of a naked platform-specific lock call outside a build-tagged file is checked as an absence, with a positive control.

### D.3 The stall threshold is a configuration check, not a constant comparison (REQ-KW-010)

Decided by three observations, none of which requires a second constant: the shipped default reads 21600; a non-positive configured value is refused by its name rather than coerced to anything; and the measured zero-hit grep of §A.5 is recorded as the reason the fourth, tempting observation — a comparison against 14400 — is absent. A criterion asserting that comparison would be unsatisfiable, and shipping it would be worse than shipping nothing.

### D.4 Branch-guard invariance is measured, never reasoned (REQ-KW-006)

The primary checkout's branch and HEAD are read before creation and again after, and must be equal. The baseline of §A.8 is the recorded starting point. A criterion that instead greps the guard's pattern list for `worktree add` is checking the list; this one checks the tree.

### D.5 Orphan classification is checked on the dirty row (REQ-KW-012)

The clean row is the permissive one and passes against an implementation with no gate at all. The load-bearing row is the dirty one: the card must be unheld *and* not dispatchable, the tree path and released holder must both be recorded, and the tree's uncommitted content must be byte-unchanged after the classification runs. A classification that reads a dirty tree and leaves it altered has already done the thing the gate exists to prevent.

### D.6 Disposal refusal must be conditional, and keyed on identities (REQ-KW-007, REQ-KW-017)

A gate that refuses unconditionally passes a naive refusal test. The criterion therefore pairs the refusal with a **positive control**: with both pull requests observed merged, the same disposal succeeds. Only the pair establishes that the gate reads merge state rather than simply declining.

Two further observations, each closing a way the gate could pass while being wrong. First, the gate must be shown to distinguish **one merged pull request from two** — a card with a single merged pull request is refused — because a predicate that collapses the two identities into one branch-shaped question is exactly the defect §A.4 repairs, and it passes a one-PR-merged test perfectly. Second, the suspended path is checked as a **behaviour**, not as a comment: with the pull-request observation unavailable, no removal occurs *and* the notice is emitted. Asserting only "no removal" would pass against an implementation that silently does nothing, which is the state REQ-KW-017 exists to make visible.

The substitution scan is checked against **both** candidates, and that is a v0.3.0 change rather than a restatement. Scanning only for a merged-branch listing would pass against an implementation that reached instead for the squash-aware `IsBranchMerged` — which is exported, `gh`-free, and sits in a package this SPEC already adopts, so it is the substitution a competent implementer is *most* likely to make. The scan therefore covers any per-branch merge predicate on the disposal path, and its positive control introduces the strategy-aware one specifically.

### D.7 The release predicate is checked on the window that defeated v0.1.0 (REQ-KW-009, REQ-KW-011)

The criterion is not "an old card is released". That passes against the v0.1.0 predicate, which is the thing being repaired.

The load-bearing construction is the **post-commit clean window** of §A.7.0, built deliberately: a card whose last-transition instant is older than the threshold, whose holder's process is **live**, and whose worktree has just been committed and is therefore clean. Under the repaired predicate no release occurs. Under the v0.1.0 predicate the card is released, classified clean, and re-dispatched — so this construction is also its own **negative control**: run against age-alone, it must report the release, establishing that the criterion can tell the two predicates apart. A criterion that cannot distinguish the repaired predicate from the defective one measures nothing.

The unprobeable arm is separate and equally load-bearing: a holder whose registry entry is absent, and a holder whose recorded host is not this one, must each produce **no** automatic release and a surfaced reason. An implementation that treats "cannot probe" as "absent" satisfies every other criterion here while reintroducing the original defect.

### D.8 Branch identification is checked on the prefix the helper does not produce (REQ-KW-003, REQ-KW-004)

The measurement of §A.2 is what makes the row selection non-obvious: a criterion exercising only `feature/`-prefixed branches passes against a prefix-keyed implementation, because that is the prefix the synthesis emits. The criterion therefore presents a worktree on `feat/<SPEC-ID>` — the form 63 of the repository's branches actually used when §A.2 measured them, 64 when re-measured at promotion time — and requires it recognized as the card's, while `feat/<other SPEC-ID>` is refused. Recognition by SPEC identifier passes both; recognition by prefix fails the first.

### D.9 The import-direction constraint is checked by the compiler, not by a reviewer (REQ-KW-018)

The claim is structural, so the check is structural: the kanban package's transitive import set must not contain the command-surface package. `go list -deps` answers it directly and its output is the verdict. The **positive control** is the compiler itself — a deliberately introduced import of the command-surface package from the kanban package must fail to build, run once and recorded — because a dependency list that has never been shown to contain the forbidden edge is indistinguishable from one queried wrongly.

### D.10 Mirror delta preservation and neutrality (REQ-KW-015)

For each mirrored pair this SPEC touches, the `diff` taken before the change must equal the `diff` taken after, once the change's own token substitutions are applied. Pair classification is **time-varying** and is re-measured at run-phase rather than trusted from this document; three pairs measured at authoring time are recorded in `plan.md` §C as the starting picture, not as a standing fact.

`internal/template/internal_content_leak_test.go` and `.github/workflows/template-neutrality-check.yaml` are the mechanical authority for neutrality, and their exit codes are the verdict. This SPEC adds one directed check but does not reimplement the guard's regex — a hand-rolled reimplementation without the guard's exemption list is a false-failure machine.

### D.11 The match rule is checked at both boundaries and on the multiplicity it must refuse (REQ-KW-003, REQ-KW-019)

Three rows, and each excludes a different wrong implementation. A **suffixed** branch must be recognized, or the rule refuses the majority of this repository's real SPEC branches (20 of 35 segments carry one). A **superstring** branch must not be, or the rule is a containment test wearing a token test's description. And where **two** branches match a single-resolution search, the refusal must fire — a criterion exercising only the one-match case passes against an implementation that silently takes the first.

The refusal's **scope** needs its own observation, and it is the row most likely to be omitted: the enumeration path of REQ-KW-007 must accept multiple matches. Without it, an implementation that refuses globally satisfies every ambiguity row while breaking disposal for every card with more than one branch — which, measured, is common.

### D.12 The two escapes are checked on their gates, not on their happy paths (REQ-KW-020, REQ-KW-021)

An escape that always fires is not a gate, and both of these are dangerous when ungated. The force-release is therefore checked on its **refusal**: with the holder probeable and observed live, it must decline and modify nothing. That row is the one distinguishing a bounded escape from a manual restatement of the age-only predicate v0.1.0 shipped.

The orphan-clear is checked on what it must **not** do: after invocation the tree's uncommitted content is byte-unchanged. An implementation that resolves the card by discarding the work satisfies every end-state assertion while destroying the thing the gate exists to protect.

### D.13 The clearing act is checked concurrently, because sequentially it cannot fail (REQ-KW-014)

The three sequential rows of `AC-KW-015` — dead process clears, live process does not, acquire path never calls it — all pass against the racy implementation, because none of them re-acquires between the inspection and the removal. The load-bearing row constructs exactly that: the artifact is released and re-acquired by a live process after the inspection, and the clear must abort rather than unlink. The row records the platform it ran on, because the artifact's persistence differs by platform (§A.6.1) and a Unix-only pass would establish nothing about the deadlock this escape exists for.

### D.14 Creation is checked for its actor and, separately, for its serialization (REQ-KW-022, REQ-KW-023)

Two observations, deliberately not one, because an implementation can satisfy either alone. The actor check runs creation with no `lead` resolvable and requires a refusal — the same shape disposal and release already carry. The serialization check runs two creations for one card from **separate operating-system processes** and requires that one is serialized behind the other, on the same reasoning as `AC-KW-014`: a same-process test measures the harness. Neither substitutes for the other — naming an actor does not serialize that actor's concurrent invocations.

---

## §E. Cross-references

- `SPEC-KANBAN-RENAME-001` — the prerequisite rename. A `dependencies:` entry and a blocking gate (REQ-KW-002).
- `SPEC-KANBAN-BOARD-001` — the prerequisite board model. REQ-KB-004 (the card record, including the holder field and the last-transition instant), REQ-KB-005 (single-origin state), REQ-KB-011 (the unheld state), REQ-KB-017 (the `lead` as sole writer of board state), REQ-KB-018 (write atomicity), and REQ-KB-019 (the board-wide lock, which supersedes card scope for board mutations while preserving REQ-KW-013 for holder assignment) are consumed here and defined there.
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-004` — the topology requirement defining and electing the `lead` role this SPEC names as the creation actor (REQ-KW-022), the disposal actor (REQ-KW-007) and, per REQ-KB-017, the writer of every holder release (REQ-KW-011). It sits in `related_specs:` rather than `dependencies:` because that SPEC already declares this one among its own dependencies (§A.4).
- `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` — the **role declaration**, and the requirement that makes it resolvable by a session that is **not** the `lead`. That clause exists for this SPEC's three gates, each of which resolves the lead from outside it; v0.2.0 deferred runtime occupancy to `REQ-KS-004`, which elects the role without defining how it is read, and the deferral is corrected at v0.3.0. The declaration's carrier is deliberately unfixed there and is not fixed here.
- `.claude/rules/moai/workflow/spec-workflow.md` § Sync to Cleanup — the established form of the two-pull-request pre-condition REQ-KW-007 adopts, stated there against pull-request numbers rather than a branch name.
- `internal/session/registry.go` — the peer registry whose `SessionID` / `PID` / `Host` fields resolve a holder to the process identity REQ-KW-011 probes; consumed for its negative reading only (§A.5).
- `internal/core/git` — the extraction target REQ-KW-018 names for branch derivation at v0.3.0. Its `doc.go` declares the package's subject as Git repository operations including branch lifecycle; `go list -deps` reports `internal/foundation` as its only internal dependency, so it imports neither consumer; and `internal/cli/worktree` already imports it (§A.9.1). It is also where `WorktreeManager` and `IsBranchMerged` live, so the extraction consolidates rather than scatters.
- `internal/worktree` — the **L1 worktree state guard**, and the v0.2.0 extraction target this revision rejects. Its `doc.go:1-5` declares it as working-tree state-guard primitives (Snapshot, Divergence, SuspectFlag), which is the mechanism §C excludes by name; its dependency-free leaf status is true and was the wrong criterion (§A.9.1).
- `internal/cli/launcher.go` — `applyCCMode` (`:227`) and `cleanupMoaiWorktrees` (`:481`), the measured reason REQ-KW-003 forbids a `worker-` prefix on a card worktree's base name. The prefix filter gates both scanned bases and the removal is non-force, so an L2 card tree named for its SPEC identifier survives a `moai cc` launch (§A.2.2).
- `SPEC-KANBAN-MULTISESSION-001` — the superseded 59-requirement predecessor. Its §A.10, §A.11, §B.10, §B.11 and §D.6 are this SPEC's primary material; five of its decisions are changed and each change is argued at the section named in HISTORY.
- `.claude/rules/moai/workflow/worktree-integration.md` § Terminology Glossary — the L2 persistent-worktree path scheme, lifetime, and disposal contract adopted whole by REQ-KW-003.
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — the doctrine REQ-KW-006 is measured against, and whose § Procedure prescribes worktree creation as the sanctioned way to work on another branch.
- `.claude/rules/moai/workflow/worktree-state-guard.md` — L1 ephemeral subagent isolation, dormant by default. Named to record that it is **not** this mechanism.
- `internal/core/git/worktree.go` and `types.go` — the existing `WorktreeManager` (`Add`, `List`, `Remove`, `Prune`, `DeleteBranch`, `IsBranchMerged`) and its named dirty-worktree error. `IsBranchMerged` (interface `types.go:194`, implementation `worktree.go:233` under a doc comment beginning at `:208`) is **squash-aware and `gh`-free**, and is the predicate whose existence falsifies v0.2.0's rejection premise for the degraded disposal path. It is not adopted for the gate, on the arity argument of §A.4.1 rather than on blindness.
- `internal/cli/worktree/shared.go:32` — `resolveSpecBranch`, the existing SPEC-to-branch derivation. Unexported, and it synthesizes the `feature/` prefix that 3 of the repository's branches carry against 63 on `feat/` at v0.2.0 authoring time, 64 at promotion time (§A.2; `research.md` §C). REQ-KW-003 keeps it only as a fallback synthesis; REQ-KW-018 extracts the derivation rather than exporting it in place.
- `internal/cli/session_worktree_prmerge.go:174-179` — `branchMergedForCleanup`, unexported and in the command-surface package, whose signature cannot carry two pull-request identities and whose own comment records the `gh`-absent fallback as squash-merge blind. It is the measured reason REQ-KW-007 keys on identities and REQ-KW-017 refuses the substitution. Its once-per-invocation blindness notice (line 138, tested at `session_worktree_prmerge_test.go:227-250`) is the precedent REQ-KW-017 follows.
- `internal/spec/lock.go`, `lock_unix.go`, `lock_windows.go` — the cross-process per-scope lock pattern REQ-KW-013 reuses (`AcquireSpecCloseLock`, `ErrSpecCloseLockHeld`, `IsLockHeldError` are exported, so no cycle arises). `lock_windows.go`'s header comment records the absent stale-lock detection REQ-KW-014 escapes (§A.6.1).
- `internal/lockfile/lockfile_unix.go`, `lockfile_windows.go` — the other advisory-lock package, whose Windows in-process fallback is the measured reason it is **not** the reuse target.
- `.claude/skills/moai/workflows/factory.md` — the documented location of the four-hour wall-clock bound REQ-KW-010's default is reasoned against, and the reason that reasoning is documentation-anchored.
- `CLAUDE.local.md` §2 (Template-First), §14 (env constants and the no-naked-syscall rule), §25 (Template Internal-Content Isolation).
