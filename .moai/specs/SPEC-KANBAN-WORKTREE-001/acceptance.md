---
id: SPEC-KANBAN-WORKTREE-001
title: "Acceptance criteria — per-card worktree lifecycle with holder liveness and mutual exclusion"
version: "0.3.0"
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, worktree, acceptance, given-when-then, verification"
tier: L
---

## §A. How these criteria are judged

Every criterion below is `Given … When … Then …`, binary, and names the command or the mechanism that decides it. A criterion whose command cannot fail is not a criterion.

### A.1 Command hygiene (binding on every criterion here)

Five rules, each paid for by a prior failure in this repository.

1. **Never read `$?` after a pipe.** `cmd | tail` makes `$?` belong to `tail`. Redirect to a log, read `rc` from the command, then count `^FAIL` across the **whole** log — not a tail, which truncates exactly the region failures live in.
2. **Never iterate an undefined shell array.** Write the literal `for` list.
3. **`spec lint` is invoked per file.** A `*.md` glob is unsatisfiable for a multi-artifact SPEC here: each path is treated as a separate SPEC, so siblings fail with `ParseFailure` or `DuplicateSPECID`.
4. **A table cell `\|` is a literal pipe, not an alternation.** Any criterion whose pattern is transcribed out of a markdown table must be re-authored before it is run, or it passes vacuously.
5. **Do not reimplement an existing guard's regex.** A guard reimplemented without its exemption list is a false-failure machine. Where a mechanical guard exists (neutrality, catalog parity), run **that guard** and treat its exit code as the verdict; a supplementary grep is an early-warning aid, never the authority.

### A.2 Positive controls

Where a criterion asserts an absence, or asserts that a refusal fires, a **positive control** is required: a construction under which the same command reports the opposite, run once and recorded. An absence-check with no demonstrated ability to fire is indistinguishable from a broken command — and so is a refusal that fires unconditionally.

### A.3 Negative controls where a criterion repairs a defect

Several criteria here exist because an earlier version of this SPEC decided something wrongly, and for those a positive control is not enough. A criterion added to repair a defect must be shown to **fail against the defective design** — otherwise it may be satisfied by both, and the repair is unverified. Each such criterion names its negative control explicitly: AC-KW-011 against an age-only release predicate, AC-KW-003 against a prefix-keyed branch match, AC-KW-007 against a branch-name-shaped merge predicate, AC-KW-019 against a containment-based identifier match and against an implementation that silently picks one of several branches, AC-KW-015 against a clearing act that unlinks without re-checking the artifact's identity, and AC-KW-023 against creation serialized by nothing but the underlying worktree command's own refusal.

### A.4 Where a criterion asserts a platform-dependent behaviour

Two rows here are judged on a platform this project does not build on daily, because the defect they cover is platform-asymmetric: on Unix the lock is released when the holding process dies, so an artifact left behind is inert, while on Windows it persists and holds (`spec.md` §A.6.1). Any such row **records the platform it ran on**. A pass recorded without its platform is not evidence — it is compatible with having exercised the half where the defect cannot occur.

---

## §B. Preconditions

**AC-KW-001** (REQ-KW-001, REQ-KW-002) — *Given* the run-phase tree, *when* `test -d internal/kanban && test ! -d internal/factory` is run and the board model's card record is scanned for the holder field, *then* both report present; and *when* the added lines of this SPEC's diff are scanned for the token `factory` (case-insensitive, excluding paths under `.moai/specs/`), *then* zero matches are reported. **Positive control**: on the plan-time tree both gates report the opposite — `internal/kanban` is absent — recorded at plan time.

Additionally, *given* this SPEC's frontmatter, *when* `dependencies:` is read, *then* it does **not** name `SPEC-KANBAN-BOOTSTRAP-001`, and *when* that sibling's own frontmatter is read, *then* it names this SPEC among its own `dependencies:` — the pair establishing that the omission is a refused cycle rather than an oversight, and that the `lead` role is consequently a runtime dependency, discharged by AC-KW-007, AC-KW-011 and AC-KW-022.

And, *given* both this SPEC's frontmatter and `SPEC-KANBAN-BOARD-001`'s, *when* each `dependencies:` line is read, *then* they do **not** both name each other. At v0.3.0 authoring time this observation **fails**: each names the other, a mutual declaration created when the board sibling promoted this SPEC out of `related_specs:` (`spec.md` §A.4.0). The criterion is written as it must eventually hold rather than as it currently reads, and the resolution belongs to the board sibling — the declared edge belongs on the landing dependency, which is this one. A run-phase that finds it still failing surfaces it rather than silently deleting this SPEC's own entry to make the check pass, since that would leave a real prerequisite undeclared.

---

## §C. Creation, naming, and per-card scope

**AC-KW-002** (REQ-KW-003) — *Given* a card whose SPEC identifier is known, *when* its worktree path is derived, *then* the path is the L2 persistent form for that identifier, obtained from the extracted derivation rather than from a new literal, **and its final segment is that SPEC identifier and does not begin with `worker-`**; *given* two different cards, *then* their paths do not collide; and *given* an existing worktree, *when* the card's branch is needed for any decision after creation, *then* the value used is the one the worktree itself reports (`git worktree list --porcelain`, `branch refs/heads/<name>`) and **no** re-derivation of a name occurs on that path — established by scanning the post-creation decision paths for a call to the branch derivation and finding none, with a **positive control**: a deliberately introduced re-derivation is reported by the same scan, run once and recorded. Decided together with a table-driven test over several identifiers, including one that is not SPEC-shaped, which must pass through the derivation unchanged. The `worker-` half is judged additionally as a **behaviour** rather than as a string check: a card worktree created at the derived path, followed by an invocation of the Claude-only launch path's worktree cleanup, leaves that worktree present — with a **positive control** in which a tree deliberately named `worker-<SPEC-ID>` under the same base is removed by the same invocation, run once and recorded, since a survival assertion with no demonstrated ability to fail establishes nothing about the sweep.

**AC-KW-003** (REQ-KW-004, no-op half) — *Given* a card whose worktree already exists at the expected path on a branch whose name carries that card's SPEC identifier, *when* creation is requested a second time, *then* it succeeds, returns no error, and the worktree list is unchanged — the same number of entries, the same path, the same branch. **This criterion is run twice, once with the branch on `feature/<SPEC-ID>` and once on `feat/<SPEC-ID>`, and both must succeed.** The second run is the load-bearing one: `feat/` is the prefix 63 of the repository's branches carry against 3 on `feature/`, and it is precisely the case a prefix-keyed implementation refuses as foreign — while the `feature/` run alone passes against exactly that implementation, because `feature/` is what the fallback synthesis emits. A second creation that errors on either run fails this criterion, because the clean-orphan re-dispatch of AC-KW-012 calls exactly this path.

**AC-KW-004** (REQ-KW-004, refusal half) — *Given* a worktree path that exists on a branch naming a **different** card's SPEC identifier, and separately *given* a branch naming this card with no worktree at the expected path, *when* creation is requested for each, *then* each is refused with its own distinct named error, the two errors are not equal, and in both cases the worktree list, the branch's tip commit, and the tree's contents are byte-unchanged afterwards. **Positive control**: both runs of AC-KW-003 succeed against the same code path, so the refusal is conditional on the identifier mismatch rather than on the prefix and rather than unconditional.

**AC-KW-019** (REQ-KW-003 match rule, REQ-KW-019) — Four rows, and the fourth is the one that keeps the refusal from breaking disposal.

*Given* a card `SPEC-X-001` and a branch `feat/SPEC-X-001-run`, *when* the branch is tested against the card, *then* it matches — the suffixed form being the majority shape in this repository (20 of 35 distinct SPEC-carrying branch segments carry one; `research.md` §C.1), so a rule requiring equality refuses most real SPEC branches.

*Given* the same card and a branch `feat/SPEC-X-0010`, *when* it is tested, *then* it does **not** match. This row is the **negative control** for the containment reading: an implementation testing for the identifier as a substring passes the first row and fails here, so a criterion carrying only the first measures nothing about the rule's boundary.

*Given* two branches both matching the card and no worktree at its path, *when* creation's fourth condition searches for a branch naming the card, *then* it refuses with a named error, surfaces **both** branch names, and modifies nothing — no worktree is created, no branch is re-pointed. An implementation that takes the first match satisfies every other row here.

*Given* the same two branches, *when* the card's pull-request identities are enumerated for the disposal gate, *then* the enumeration accepts both and does **not** refuse. This row is the scope control: without it, an implementation that refuses on multiplicity everywhere passes the third row while making disposal impossible for every card with more than one branch, which measured is the common case.

**AC-KW-005** (REQ-KW-005) — *Given* one card, *when* its `run`, `review`, and `sync` phases each resolve their working directory, *then* all three resolve the same path, and *when* the worktree list is read after all three, *then* exactly one entry exists for that card. A per-session implementation would produce three.

**AC-KW-022** (REQ-KW-022) — *Given* a card and a resolvable `lead`, *when* creation is performed, *then* the actor performing it is the session occupying the `lead` role and no other, the occupancy having been read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` from a session that is not itself the `lead`; and *given* the same card with **no** session occupying that role resolvable, *when* creation is requested, *then* it is refused, the absence is surfaced, and no worktree and no branch are created. The refusal row is the load-bearing one — the permissive row passes against an implementation with no actor rule at all — and it is the same shape AC-KW-007 and AC-KW-011 already require for disposal and release, so a creation path lacking it is the family's one remaining unattributed lifecycle act.

**AC-KW-023** (REQ-KW-023) — *Given* one card, *when* two creation sequences for it are begun from **separate operating-system processes**, *then* the second is serialized behind the first — refused with the card lock's named held-error, or admitted only after the first releases — and *when* both have completed, *then* exactly one worktree exists for the card and the second sequence's outcome is one of REQ-KW-004's four named outcomes rather than an unclassified error. A same-process test does **not** satisfy this criterion, for the reason AC-KW-014 already gives. **Negative control**: with the serialization removed and only the underlying worktree command's own refusal in place, the pair must be able to produce an outcome outside those four — the interval in which the path exists and its branch is not yet reportable — establishing that the criterion distinguishes serialization from the command's own error handling. *When* the creation path is scanned for the card lock's acquisition, *then* it is present.

---

## §D. Branch-guard invariance

**AC-KW-006** (REQ-KW-006) — *Given* the primary checkout's branch and short HEAD read immediately before a card's worktree is created, *when* creation completes and both are read again, *then* both are equal to the pre-creation values. **Positive control**: a deliberate `git switch` in the primary checkout, performed once in a scratch clone, changes the compared value — establishing that the comparison can detect a disturbance. A criterion that instead greps the branch guard's forbidden-pattern list for `worktree add` does **not** satisfy this: it decides a fact about the list, not about the tree.

---

## §E. Disposal and refused removal

**AC-KW-007** (REQ-KW-007) — *Given* a card in `done` for which two pull requests are discovered, one observed `MERGED` and one not, *when* disposal is evaluated, *then* it is refused and the worktree still exists; *given* the same card for which only **one** pull request is discovered and it is observed `MERGED`, *then* disposal is likewise refused — this row being the one that distinguishes a two-identity gate from a branch-shaped predicate, which passes the one-merged case and would dispose a card that has run but not synced; and *given* the same card with **two** discovered and **both** observed `MERGED`, *when* disposal is evaluated by the lead, *then* the worktree is removed. The third row is the **positive control**: without it, an implementation that refuses unconditionally passes. Additionally, *when* the disposal path is scanned for a decision taken from the card's column, *then* none is found; *when* it is scanned for a predicate whose inputs are a branch name and an availability flag, *then* none is found, each identity being verified individually in the form `spec-workflow.md` § Sync to Cleanup prescribes; and *when* disposal is evaluated with no session occupying the `lead` role resolvable — the occupancy being read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006`, from a session that is not itself the `lead` — *then* it is refused, the absence is surfaced, and the worktree still exists.

**AC-KW-017** (REQ-KW-017) — *Given* a card whose two pull requests are both merged and whose pull-request observation is made **unavailable**, *when* disposal is evaluated, *then* no removal occurs **and** a notice is emitted recording that disposal is suspended and worktrees will accumulate unreclaimed. Both halves are required: asserting only "no removal" passes against an implementation that silently does nothing, which is the state this requirement exists to make visible. Additionally, *when* the disposal path is scanned for **any per-branch merge predicate** used as a substitute for the unavailable observation, *then* none is found — the scan covering both a reachability-based merged-branch listing and the strategy-aware `IsBranchMerged` exported by the `WorktreeManager` this SPEC otherwise adopts. Scanning for the listing alone does **not** satisfy this criterion: `IsBranchMerged` is squash-aware, needs no `gh`, and sits in a package already imported on this path, so it is the substitution an implementer is most likely to reach for, and it is the one v0.2.0's premise wrongly implied did not exist. **Positive controls**, three: with the observation available the same card disposes, so the suspension is conditional; a deliberately introduced merged-branch substitution is reported by the scan; and a deliberately introduced `IsBranchMerged` substitution is reported by it as well, each run once and recorded. Two measurements stand behind this criterion and are recorded rather than re-derived — over the last 200 first-parent commits of `origin/main`, 0 are merge commits and 199 of 200 subjects carry a `(#NNNN)` suffix, so a reachability listing cannot open this gate here; and three distinct branches carry one card's SPEC identifier (`research.md` §C.1), so a branch count is not a pull-request count and no branch-subject predicate can supply the gate's at-least-two condition.

**AC-KW-008** (REQ-KW-008) — *Given* a card's worktree holding uncommitted work, *when* removal is attempted, *then* it is refused, the tree and its uncommitted content are byte-unchanged, and the failure is recorded with the tree's path and surfaced; and *when* the implementation is scanned for a forced-removal call or a retry of a refused removal, *then* none is found. **Positive control**: a deliberately introduced forced-removal call is reported by the same scan, run once and recorded.

---

## §F. Stall detection

**AC-KW-009** (REQ-KW-009) — *Given* a card whose last-transition instant is older than the threshold, *when* the age criterion is evaluated, *then* the card is judged a release **candidate**; and *given* the same card with the instant advanced to now and **nothing else changed**, *then* it is not. The second half is load-bearing: an implementation reading any other signal into the age criterion could not change its answer from that mutation alone. Additionally, *when* the card record is inspected, *then* it carries no heartbeat field; *when* the age criterion's own inputs are enumerated, *then* the peer registry is not among them and no tree-cleanliness observation is among them — the registry being read only by the release predicate of AC-KW-011, and then only to resolve a process identity; and *when* the implementation is scanned for a decision that reads an entry's mere existence as evidence a session is alive, *then* none is found. **Positive controls**: a deliberately introduced cleanliness input to the age criterion, and a deliberately introduced existence-implies-alive read, are each reported by the same scans, run once and recorded.

**AC-KW-010** (REQ-KW-010) — *Given* the shipped configuration, *when* the stall threshold is read with no operator override, *then* it is 21600; and *given* a configured value of `0` and separately of a negative value, *when* configuration is loaded, *then* each is rejected with a named error and no coerced value is returned. **Deliberately absent**: any comparison against the coder chain's four-hour bound. `grep -rn '14400' --include='*.go' internal/ pkg/` exits 1 with zero matches, so no constant exists to compare against; the relationship is documentation-anchored to `.claude/skills/moai/workflows/factory.md`, and a criterion asserting it mechanically would be unsatisfiable.

---

## §G. Holder release and orphan classification

**AC-KW-011** (REQ-KW-011) — Four constructions, and the second is the one this criterion exists for.

*Given* a card in `run` whose age criterion has fired **and** whose holder's recorded process is observed absent on the host that record names, *when* release is evaluated, *then* the holder is released, the card's holder is empty, its column is still `run`, and its last-transition instant is unchanged by the release.

*Given* the **post-commit clean window** — a card whose age criterion has fired, whose holder's recorded process is **live**, and whose worktree has just been committed and is therefore clean — *when* release is evaluated, *then* **no release occurs** and the card remains held. This construction is deliberate and is built rather than waited for. It is also its own **negative control**: evaluated against an age-only predicate it must report a release, so a criterion that cannot separate the two predicates is a criterion that measures nothing. An implementation that releases here hands a live session's card to a second session, which is the outcome the whole section prevents.

*Given* a card whose age criterion has fired and whose holder resolves to **no registry entry**, and separately *given* one whose recorded host is **not this one**, *when* release is evaluated, *then* neither is released automatically, and each is surfaced with the card, the holder's identity, and the reason it is unprobeable. An implementation that reads "cannot probe" as "absent" satisfies every other row here while restoring the original defect.

*Given* any of the above, *when* the actor performing the release is observed, *then* it is the session occupying the `lead` role and no other, per `REQ-KB-017`, the occupancy being read through the role declaration of `SPEC-KANBAN-BOOTSTRAP-001` `REQ-KS-006` from a session that is not itself the `lead`; *when* release is attempted with no `lead` resolvable, *then* it is refused and the absence surfaced; and *when* the column enumeration is read, *then* it still carries exactly the six values the board model declares — no held column, no blocked column, no seventh value.

**AC-KW-020** (REQ-KW-020) — *Given* a card the third construction of AC-KW-011 left unreleased because its holder is unprobeable, *when* the force-release operation is invoked by an operator, *then* the holder is released, the card's column is unchanged, the operation reports the card, the holder's identity and the reason the holder was unprobeable, and the orphan classification of AC-KW-012 or AC-KW-013 then applies to the card exactly as it would after an automatic release — so the card is demonstrably no longer terminal.

*Given* a card whose holder's recorded process **can** be probed and is observed **live**, *when* the same operation is invoked, *then* it refuses and modifies nothing. This row is the criterion's whole point: an unconditional force-release is v0.1.0's age-only predicate with a human in the loop, and a human looking at a stuck board is not a liveness oracle. A criterion carrying only the first row passes against exactly that implementation.

Additionally, *when* the automatic release path of AC-KW-011 is scanned for a call to this operation, *then* none is found — with a **positive control**: a deliberately introduced automatic invocation is reported by the same scan, run once and recorded.

**AC-KW-012** (REQ-KW-012, clean half) — *Given* an orphaned worktree with no uncommitted changes, *when* it is classified, *then* it is reported clean, the card is reported re-dispatchable, and the target tree for the re-dispatch is that same path; and *when* creation is invoked for that card as part of the re-dispatch, *then* it succeeds as the no-op of AC-KW-003.

**AC-KW-013** (REQ-KW-012, dirty half) — *Given* an orphaned worktree holding uncommitted changes, *when* it is classified, *then* it is reported dirty, the card is reported **not** dispatchable, the tree's path and the released holder's identity are both recorded on the card and surfaced, and the tree's uncommitted content is byte-unchanged after classification. A test asserting only the clean half fails this criterion: the permissive path passes against an implementation with no gate at all, and the gate is the entire behavior.

**AC-KW-021** (REQ-KW-021) — *Given* the card AC-KW-013 left withheld, *when* the orphan-clear operation is invoked by an operator, *then* the card's recorded orphan-tree path and released-holder identity are both absent afterwards, the card is reported re-dispatchable, and the operation reports what it cleared — so "cleared by a human" is an observable end-state rather than an unstated one.

*And in the same run*, *then* the orphaned tree's uncommitted content is **byte-unchanged**, and its working tree is still reported dirty. This half is not a secondary assertion: an implementation that resolves the card by discarding, resetting, stashing, or committing the work satisfies every end-state assertion above while destroying precisely what the withholding gate exists to protect.

Additionally, *when* the classification path of AC-KW-013 and any scheduled or age-driven path is scanned for a call to this operation, *then* none is found — the clearing being invoked by an operator and never automatically. **Positive control**: a deliberately introduced automatic invocation is reported by the same scan, run once and recorded.

---

## §H. Mutual exclusion on holder mutation

**AC-KW-014** (REQ-KW-013) — *Given* one card, *when* a second holder-mutation attempt is made from a **separate operating-system process** while the first holds the card's lock, *then* the second is refused with the named held-error; and *when* the first releases, *then* a subsequent attempt succeeds. A same-process test does **not** satisfy this criterion — it passes against an in-process fallback, which is the failure the substrate choice exists to prevent. Additionally, *given* the new lock's source files, *when* they are listed, *then* a build-tagged counterpart exists for every platform the reused `internal/spec` lock pattern already supports, and *when* the package body outside those build-tagged files is scanned for a naked platform-specific lock call, *then* none is found — with a positive control.

**AC-KW-015** (REQ-KW-013 fence, REQ-KW-014) — *Given* the assignment path, *when* it is scanned for a condition gating assignment on the working tree being clean, *then* none is found; and *given* a lock artifact left behind by a process that no longer exists, *when* the card's holder is read, *then* the holder is whatever the board record says and is **not** derived from the artifact, and the card is not reported held on the artifact's account. **Positive control**: a deliberately introduced clean-tree gate, and a deliberately introduced holder-from-lock inference, are each reported by the same scans, run once and recorded.

Additionally — and this half is judged on the platform whose implementation lacks stale-lock detection — *given* a lock artifact whose recorded process is observed absent, *when* the clearing operation is invoked, *then* the artifact is removed, the operation reports what it removed, and the card's holder can subsequently be mutated; *given* a lock artifact whose recorded process is **live**, *when* the same operation is invoked, *then* the artifact is **not** removed; and *when* the acquire path is scanned for a call to the clearing operation, *then* none is found, the clearing being explicit rather than a step acquisition takes on its own. And — the row v0.2.0 lacked, and the only one of the four that is concurrent — *given* a lock artifact whose recorded process is observed absent, *when* the artifact is **released and re-acquired by a live process between the clearing operation's inspection and its removal**, *then* the clear **aborts**, reports that it aborted, and the artifact is still present and still held by the live acquirer afterwards. All four rows are required, and the fourth is the one the other three cannot reach: each of them is sequential, so each passes against an implementation that inspects and then unlinks unconditionally — which admits two holders to the critical section the lock exists to hold. **Negative control**: run against a clearing act with no identity re-check, this row must report the unlink, establishing that the criterion separates the hardened implementation from the racy one. **Positive control** for the acquire-path scan: a deliberately introduced automatic clear is reported by it, run once and recorded.

All four rows **record the platform they ran on**, per §A.4. The artifact's persistence is platform-asymmetric — on Unix the lock is released when its holder dies and the file left behind is inert, measurably so (`.moai/state/` currently holds 14 such artifacts, the oldest from 2026-05-30; `research.md` §E.3) — so a pass recorded without its platform is compatible with having exercised only the half where the deadlock cannot occur.

---

## §I. Reuse without an import cycle

**AC-KW-018** (REQ-KW-018) — *Given* the kanban package, *when* `go list -deps` enumerates its transitive imports, *then* the command-surface package `internal/cli` is not among them; and *given* the branch derivation, *when* its home package's own transitive imports are enumerated, *then* neither `internal/cli` nor any package under it is among them, so both consumers reach it without either depending on the other. *And* — the observation v0.2.0 lacked, which is why it selected a package this SPEC excludes — *when* that home package's `doc.go` is read, *then* its declared subject is Git repository operations including branch lifecycle, and *when* the package excluded by `spec.md` §C as the L1 worktree state guard is checked, *then* the derivation is **not** there. A dependency-count check alone does not satisfy this criterion: the previously selected package passes it and is still the wrong home, so cohesion is asserted from the package's documented purpose rather than from its import graph. **Positive control, and the compiler is the instrument**: a deliberately introduced import of `internal/cli` from the kanban package must fail to build with an import-cycle diagnostic, run once and its output recorded. Without it a clean dependency list is indistinguishable from a list queried wrongly — and the cycle is the whole reason the requirement exists. Additionally, *when* the diff is checked against `internal/cli/session_worktree_prmerge.go`, *then* it is unmodified: this SPEC implements its own merge-observation contract rather than widening an existing caller's.

## §J. Mirror, neutrality, and verification

**AC-KW-016** (REQ-KW-015, REQ-KW-016) — *Given* every mirrored pair this SPEC touches, *when* each pair's `diff` is taken before and after the change, *then* the two are equal once the change's own token substitutions are applied — a pair measured byte-identical stays byte-identical and a sanitized pair keeps exactly what its template side strips; *when* `make build` has run, *then* the regenerated `internal/template/catalog.yaml` is committed; *when* the neutrality guards (`internal/template/internal_content_leak_test.go` and the template-neutrality workflow) are run, *then* each exits 0, their exit codes being the verdict rather than any grep authored here; and *when* the **full** test suite is run — not an affected-packages subset — *then* it passes. Pair classification is re-measured at run-phase; the plan-time measurements (`archive.yaml` byte-identical, `handoff.yaml` differs by 3 lines, `state.yaml` by 7) are a starting picture, not a standing fact.

---

## §K. Traceability

Twenty-three criteria against twenty-three requirements. Every requirement maps to at least one criterion, and every criterion to at least one requirement. **Tier L: 23 of 25 criteria against 23 of 25 requirements — within budget, with two of each remaining.** The v0.3.0 additions are `AC-KW-019` … `AC-KW-023`, matching `REQ-KW-019` … `REQ-KW-023`; `AC-KW-001`, `AC-KW-002`, `AC-KW-007`, `AC-KW-011`, `AC-KW-012`, `AC-KW-013`, `AC-KW-015`, `AC-KW-017` and `AC-KW-018` gain rows in place. Folding each addition into an existing criterion was available and refused; the reasoning is recorded in `spec.md` §B, and the shape of the refusal is the same in every case — one verdict over two observations lets the weaker half break while the stronger half passes.

| Requirement | Criteria |
|---|---|
| REQ-KW-001 | AC-KW-001 |
| REQ-KW-002 | AC-KW-001, AC-KW-007, AC-KW-011, AC-KW-022 |
| REQ-KW-003 | AC-KW-002, AC-KW-019 |
| REQ-KW-004 | AC-KW-003, AC-KW-004 |
| REQ-KW-005 | AC-KW-005 |
| REQ-KW-006 | AC-KW-006 |
| REQ-KW-007 | AC-KW-007, AC-KW-017, AC-KW-019 |
| REQ-KW-008 | AC-KW-008 |
| REQ-KW-009 | AC-KW-009 |
| REQ-KW-010 | AC-KW-010 |
| REQ-KW-011 | AC-KW-011 |
| REQ-KW-012 | AC-KW-012, AC-KW-013 |
| REQ-KW-013 | AC-KW-014, AC-KW-015, AC-KW-023 |
| REQ-KW-014 | AC-KW-015 |
| REQ-KW-015 | AC-KW-016 |
| REQ-KW-016 | AC-KW-016 |
| REQ-KW-017 | AC-KW-017 |
| REQ-KW-018 | AC-KW-018 |
| REQ-KW-019 | AC-KW-019 |
| REQ-KW-020 | AC-KW-020 |
| REQ-KW-021 | AC-KW-021 |
| REQ-KW-022 | AC-KW-022 |
| REQ-KW-023 | AC-KW-023 |

### Out of Scope — criteria deliberately not written

- A criterion comparing the stall default against the coder chain's four-hour bound. Unsatisfiable — see AC-KW-010.
- A criterion over the board's columns, WIP admission, or the column↔status table. Those belong to `SPEC-KANBAN-BOARD-001`; AC-KW-011 reads the column enumeration only to establish that this SPEC added nothing to it.
- A criterion over dispatch, roles, bootstrap, or backend selection. Those belong to `SPEC-KANBAN-BOOTSTRAP-001`. AC-KW-007 and AC-KW-011 observe only that the actor is the `lead` and that its absence is refused; they elect nothing.
- A criterion over who may write board state, over write atomicity, or over the board-wide lock. Those are `SPEC-KANBAN-BOARD-001` `REQ-KB-017`, `REQ-KB-018` and `REQ-KB-019`; AC-KW-011 observes the actor because this SPEC's release is subject to that rule, not because it defines it.
- A criterion asserting that tree cleanliness predicts liveness, in any direction. Rejected at v0.2.0 — `spec.md` §A.7.0. AC-KW-011's second construction asserts the opposite: a clean tree and a live process together must produce no release.
- A criterion over the operator surface through which AC-KW-020 and AC-KW-021 are invoked — its subcommand, flags, or prompts. Both criteria judge the operations' gates and observable end-states; the surface belongs to the bootstrap sibling's command layer, and asserting one here would fix a decision this SPEC does not own.
- A criterion asserting that a degraded, `gh`-free disposal path exists. Deliberately absent, and its absence is now argued rather than assumed: `IsBranchMerged` is squash-aware and needs no `gh`, so the v0.2.0 reason for having no such path is false — but it answers a per-branch question and the gate is per-pull-request-identity, so no criterion can be written for a path that cannot express the condition (`spec.md` §A.4.1). AC-KW-017 asserts the substitution's **absence** instead, and now scans for that predicate by name.
