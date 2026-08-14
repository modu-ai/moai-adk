---
id: SPEC-KANBAN-BOARD-001
title: "Progress — six-column kanban board model with a single-origin board state store"
version: "0.6.2"
status: draft
created: 2026-08-10
updated: 2026-08-14
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, board, progress, evidence"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md` (Tier L set — `design.md` and `research.md` added at v0.2.0 with the promotion).

- Requirements: 25 (`REQ-KB-001` … `REQ-KB-025`) — **at** the Tier L ceiling of 25, with none to spare as of v0.5.0.
- Acceptance criteria: 25 (`AC-KB-001` … `AC-KB-025`) — **at** the Tier L ceiling of 25, with none to spare.
- Milestones: 4 (M0 … M3), ordered by decision-reversibility. M1 grew at v0.2.0 to carry the sole-writer, atomicity, and board-wide-lock requirements ahead of the M2 admission they make sound, and again at v0.3.0 to carry the absent/unreadable split, the recovery bound, and the stale-lock exit. `REQ-KB-020` lands in M2, ahead of the compatibility table it feeds, and `REQ-KB-024` lands beside it at v0.4.0.
- Dependencies: `SPEC-KANBAN-RENAME-001` alone — **landed on `origin/main` as of the v0.5.0 re-measurement** (`git ls-tree -d --name-only origin/main internal/kanban` prints the path; the same query for `internal/factory` prints nothing). M0 still runs the gate and is now expected to pass rather than halt; AC-KB-001 decides it against `origin/main` rather than against the working tree, as does `plan.md` §C command 1. The prerequisite SPEC's own frontmatter still reads `status: in-progress`, which is bookkeeping lag and not a landing signal — the base-branch query is the observation relied on, and it is the one `acceptance.md` §A.1 rule 6 makes authoritative.
- Siblings, both consumed by name and neither declared: `SPEC-KANBAN-WORKTREE-001` (`REQ-KW-003`, read by `REQ-KB-020`) and `SPEC-KANBAN-BOOTSTRAP-001` (`REQ-KS-006`, read by `REQ-KB-017`). Each already declares this SPEC among its own `dependencies:` on a landing need, so a reverse edge would close a cycle; both consumptions are **contract** dependencies discharged by citation, carried in requirement text (`spec.md` §A.4a). The absences are decisions — `plan.md` AP-27 forbids re-promoting either. The scope boundary is stated in `spec.md` §C so an omission is not read as a gap.

### v0.2.0 — plan-audit repair

Promoted Tier M → Tier L. Six defects repaired; the first is the one the promotion exists for.

1. **F1 — the sole-writer rule was lost in the split** and is restored as `REQ-KB-017` … `REQ-KB-019` (`spec.md` §A.7). Measured loss: `grep -rc 'sole writer\|single writer'` over all three sibling SPECs reported 0 in every file, and `grep -c 'atomic'` over `spec.md` reported 0. Consequences closed: nothing stopped a worker writing the board; the WIP-2 bound of `REQ-KB-009` was unenforceable under a card-scoped lock; and `REQ-KB-013` bricked the board permanently on one partial write.
2. **`REQ-KB-013` amended** with a bounded, operator-visible recovery path out of the unknown state, with the refusal preserved unchanged (`AC-KB-020`).
3. **The primary-checkout path rule was wrong in the binding document** — `spec.md` §A.3 and `REQ-KB-005` phrased against the bare `--git-common-dir`, which returns a relative `.git` in the primary checkout. Repaired against `internal/hook/branch_guard.go` lines 178 and 190; `plan.md` already carried the correct form.
4. **`(sync, completed)` added to the compatibility table** — the usually-observable pairing, whose omission would have marked nearly every card inconsistent one column short of `done`.
5. **`status: planned` handled**, and the three out-of-lifecycle terminals stated as producing no card (`AC-KB-021`).
6. **Two criteria repaired**: `AC-KB-001` observed the working tree while claiming to observe the base branch; `AC-KB-012` asserted an independence claim with no falsifiable observation.

### v0.3.0 — plan-audit repair, second pass

Scored 0.825 against the Tier L threshold of 0.85. The six v0.2.0 repairs verified closed with zero regressions; the shortfall was entirely in the layer they exposed. Eight repairs, all delta-scoped — no rewrite, and no requirement withdrawn.

1. **The tree the board reads `status` from was decided by no SPEC**, and the omission broke the normal path: transitions land on the card's branch, whose worktree survives until both pull requests merge, so the primary-side copy reads `draft` while the card sits in `run` — `(run, draft)`, refused by `REQ-KB-008`, for every card. `REQ-KB-020` + `AC-KB-022` (`spec.md` §A.4a). `REQ-KW-003` is consumed as a **contract** dependency in requirement text, the sibling staying in `related_specs:`. It was briefly promoted to `dependencies:` inside this revision; that closed a cycle against the sibling's landing dependency on this SPEC, was found by the sibling's author and recorded in its §A.4.0 with `AC-KW-001` observing it, and is reversed here. `spec.md` §A.4a carries the two-kinds analysis, `spec.md` §C forbids resolving it from the sibling's side, and `plan.md` AP-27 forbids re-promoting.
2. **The board-wide lock re-created the brick `REQ-KB-013` was amended to remove** — the reused family performs no stale-lock detection on its Windows substrate, by its own header. `REQ-KB-023` + `AC-KB-023`, porting `REQ-KW-014`'s shape and hardened against the check-then-unlink race that sibling's audit found. Measured as platform-asymmetric: Unix `flock(2)` releases on exit, so the hazard is invisible on the developer's machine (`research.md` §M).
3. **`plan.md` §C command 1 still ran `test -d`** — the predicate `acceptance.md` §A.1 rule 6 was written against, in the one tree guaranteed to satisfy it falsely. Replaced with `git ls-tree` against `origin/main`; §C and §F swept for others.
4. **An absent state file was owned by nobody**, and nothing created the store's directory. `REQ-KB-021` + `AC-KB-024`, deciding absent as a legitimately empty board and unreadable as unknown, checked against each other.
5. **"Reuse rather than re-derive" mandated something the code shape forbids** — `isPrimaryCheckout` is unexported and returns a boolean, not a path. `REQ-KB-005` takes `REQ-KW-018`'s extraction disposition into `internal/core/git`, target confirmed by reading its `doc.go`.
6. **The sanctioned recovery could produce the empty board §A.6 forbids**, and "bounded" was asserted without a bound. `REQ-KB-022`, folded into `AC-KB-020`.
7. **The board state path collided with `REQ-KR-009`'s session-record path** — one name, two occupants, two resolution rules. Board moved to `.moai/state/kanban-board/`; the sibling is untouched.
8. **Three criteria repaired**: `AC-KB-012` varied a knob this SPEC excludes (reverse direction re-expressed as an absence over the board's own inputs); `AC-KB-017` asked a static enumeration to establish a runtime value (split static/runtime, runtime half reading `REQ-KS-006`'s role declaration); `AC-KB-002` never entered the fallback branch it requires (third resolution forced through the `execCommand` indirection). Line-number citations removed from normative text throughout.

Two measurements contradicted the brief that motivated them and are recorded rather than smoothed: the stale-lock brick is Windows-only, not universal (`research.md` §M); and the borrowed fallback normalization is **sound** as a resolver — what is unsound is inferring anything about it from the existing caller's green suite, since an equality comparison cannot detect a shared offset (`research.md` §L.2).

### v0.4.0 — plan-audit delta repair (D1-D2)

Two defects, both inside requirements v0.3.0 had just added, and both of the same family: a requirement written as though a thing were singular when it is not. One requirement and one criterion were spent of the two and one available; the rest was closed by amending four existing entries in place.

| ID | Defect | Closure |
|---|---|---|
| D1 (critical) | `REQ-KB-020` read a card's `status` "from the card's branch" without saying **which**, and cards keep every branch they ever carried — nothing in this family deletes one (`REQ-KW-004` is scoped to a *mismatched* branch during a creation refusal; `REQ-KW-007` removes the worktree only). Measured, **3 of the 29** SPEC identifiers on branches carry two or more, and `SPEC-NAVIGATOR-SYNC-003` carries `draft`, `in-progress` and `completed` at once. The consequence is the v0.2.0 defect reflected: v0.2.0 read the primary checkout and broke `run`; v0.3.0's text breaks `sync` and `done`, since a disposed card resolving a retained `plan/` branch pairs as `(done, draft)`, outside §A.4's table and refused by `REQ-KB-008`. Separately, `SPEC-KANBAN-WORKTREE-001` §A.2.1 routes `REQ-KW-019`'s multiple-match outcome to this SPEC by name and it never arrived — `grep -rn 'KW-019'` in this directory returned **0** | the source is selected on **worktree liveness**, not on branch existence: the branch the card's worktree reports (`REQ-KW-003`'s observation route, row 1 of `REQ-KW-019`'s cardinality table where exactly one branch is reported by construction), and the primary checkout where no worktree is live. The board never enters the search route, so no tiebreak is required and none is permitted. `REQ-KB-020` amended, `REQ-KB-024` added for the two residuals (a worktree reporting no branch; a `REQ-KW-019` refusal reaching the board), `AC-KB-022` rebuilt on the measured three-branch shape, `AC-KB-025` added. New `spec.md` §A.4b, `design.md` §B.4a, `research.md` §O.2, `plan.md` AP-28 / AP-28a |
| D2 (major) | `REQ-KB-023` required the removal to be conditioned on the artifact still being the one inspected, **deterministically** — an atomicity POSIX does not offer. `unlink(2)`, and the `os.Remove` wrapping it, takes a **path** and resolves the name at call time; the descriptor the identity was read through has no bearing on which file is unlinked. No portable handle-based form exists: `funlinkat(2)` is FreeBSD-only, and the Windows delete-on-close disposition is set by the holder at open time — the very process the operation exists to clean up after | weakened to what is achievable and named as such: re-read the recorded identity **immediately before** the unlink and abort on mismatch, stated in the requirement as a **time-of-check-to-time-of-use mitigation** that narrows the window rather than closing it, with the residual recorded. `AC-KB-023` now decides the mitigation and states explicitly what its passing does *not* establish. Rejected alternatives recorded in `design.md` §C.3a (exclusive-acquire-then-unlink, whose portability cost runs the wrong way — least expressible on the only platform where the defect occurs); `plan.md` AP-29 added for the durable error of performing the re-read and then reasoning as though the artifact were pinned |

Counts re-measured after the edits, in this worktree:

```
$ grep -cE '^\*\*REQ-KB-[0-9]{3}\*\*' spec.md
24
$ grep -cE '^\*\*AC-KB-[0-9]{3}\*\*' acceptance.md
25
```

Against the Tier L ceiling of 25 and 25: **one requirement to spare, criteria at the ceiling.** Both sequences are contiguous with no duplicates (`REQ-KB-001` … `REQ-KB-024`, `AC-KB-001` … `AC-KB-025`). The headroom was two and one; one of each was spent, and `REQ-KB-024` was authored separately rather than folded into `REQ-KB-020` because an implementation can resolve the source correctly on every card that has one and still render a card that has none as `draft` — the zero-value default, which pairs legally with `backlog` and `plan` and dispatches.

**Two premises of the audit brief were contradicted by measurement and are recorded rather than smoothed.**

1. **"Fifteen SPEC IDs carry two or more stage branches."** Measured, **3 of 29**. The brief's command counts `origin/feat/X` and `feat/X` as two branches; fourteen of its fifteen are local-tracking pairs of a *single* branch, and after `sed 's|^origin/||' | sort -u` its own command yields **1**. Applying `REQ-KW-003`'s exact-token rule to the 158 deduplicated names gives 3. The defect is undiminished — the three that remain genuinely disagree, and two of them are how `REQ-KB-020` would have been implemented — but "the normal state of the repository" overstates it, and the prevalence claim is stated at the measured figure in `spec.md` §A.4b and `research.md` §O.2.
2. **"`REQ-KW-004` forbids branch deletion."** It does not. Its prohibition is scoped to adopting, re-pointing, resetting or deleting a **mismatched** tree or branch during a creation refusal. There is no general no-deletion rule anywhere in the sibling — which makes §A.4a's tail case (a branch deleted after merge) legitimate rather than dead text, and it is retained in the amended `REQ-KB-020` as one of the three conditions the fallback covers. The conclusion the brief drew from the wrong premise is nonetheless correct by a different route: branches are *not in practice* deleted, so branch existence is a selector that never flips, which is why the amendment keys on worktree liveness instead.

One further correction, found while editing and outside both defects: `design.md` §B.4 asserted that `SPEC-KANBAN-WORKTREE-001` "is now a `dependencies:` entry", contradicting `spec.md` §A.4a, `plan.md` AP-27 and the frontmatter, all of which record the v0.3.0 promotion as made and **reversed**. The sentence survived the reversal in that file alone and is corrected.

**Artifact versions.** All six move to `0.4.0`; every one was edited. `research.md` gains §O.2 (the branch-multiplicity measurement), `plan.md` M2 and its anti-pattern catalogue were carrying the superseded branch-existence rule and would otherwise have contradicted `REQ-KB-020` directly.

### v0.5.0 — scope refresh and one forced borrow

Not a plan audit. A refresh against what landed after v0.4.0 was authored, plus one decision the scope change forced. Six changes, one requirement spent, no criterion added, retired, or merged.

| # | Change | Ground |
|---|---|---|
| 1 | **`REQ-KB-025` added — the role-declaration carrier is borrowed conditionally.** `spec.md` §A.8 (the borrow and its seam), §D.14 (the verification surface), §C (exclusion narrowed from the whole mechanism to the contract alone), `design.md` §C.1a (the decision and its four rejected alternatives), `plan.md` §E D6 + M1 + AP-30/AP-31, `acceptance.md` `AC-KB-017` carrier half | `SPEC-KANBAN-BOOTSTRAP-001` left the current scope, so `REQ-KS-006` has no landing date and `REQ-KB-017`'s runtime refusal was consuming a contract that will not arrive. A refusal with no readable role does not refuse — it admits every write, which is §A.7's prose-only-rule failure reached by pointing at an absent rule instead of by deleting one. The contract stays the sibling's and is not restated; only the **carrier**, which `REQ-KS-006` fixes not at all by its own terms, is taken — and it is bound by that requirement's own resolvable-from-a-non-`lead`-session clause so the two cannot fork |
| 2 | **The rename prerequisite landed; `plan.md` §B.1 and §C command 1 inverted** | Measured at `origin/main` = `c55c61aa5`: `git ls-tree -d --name-only origin/main internal/kanban` prints the path, the same query for `internal/factory` prints nothing. The gate written to halt now passes (`research.md` §P.1) |
| 3 | **Three stale citations repaired** — the session-record constant is `internal/kanban/record.go`'s `{".moai", "state", "kanban"}`, not `internal/factory/record.go`'s `{".moai", "state", "factory"}`. `spec.md` §A.3(e) + §E, `plan.md` §B.8 / §C-11 / M1 / AP-24, `design.md` §B.1, `acceptance.md` `AC-KB-002` | The rename moved the file and changed the value, which is what turned §A.3(e)'s collision from a **projection** into an **observation** — before the rename the two names did not yet coincide (`research.md` §P.2) |
| 4 | **`.moai/state/kanban/` restated at its current contents** | It exists and holds two per-tree occupants — the session records and `backlog.json`; `kanban-board` is still absent. §A.3(e) measured "no entry named `kanban` or `kanban-board`", now half false. The change **strengthens** the separation: a primary-resolved board would sit among two per-tree stores rather than one (`research.md` §P.3) |
| 5 | **`kanban-dispatch.md` reconciled — six agreements cited, two disagreements reported.** New `spec.md` §A.9, `plan.md` §B.9 + AP-32, §C command 11b | That always-loaded rule states part of this SPEC's model and reaches every session. Where it agrees it is cited rather than restated, so the two cannot drift into two versions of one claim. It **disagrees** on column derivation and on what the backlog is; neither is harmonized (see below) |
| 6 | **The admission mechanism exists.** New `spec.md` §A.10, `plan.md` §B.10; the closure recorded on the sibling that recorded the hole (`SPEC-KANBAN-BOOTSTRAP-001` v0.6.0) | `/moai todo` is the CLI-verb shape that sibling's `plan.md` §B.4 named first of three. Cited, not annexed — closing another SPEC's note by widening one's own scope is how the predecessor reached 59 requirements |

**Two disagreements, one per surface, reported rather than silently harmonized.** (This heading read "Two disagreements with `kanban-dispatch.md`" at v0.5.0 and was wrong about the second; corrected at v0.6.0 — see that entry's finding 4.)

1. **Column derivation.** That rule's § Boundaries holds column position "re-derived from SPEC status after a clear". `REQ-KB-006` forbids exactly that, and §A.2 rejects it on the measured `run`/`review` collision — both read `in-progress`, and the `§E.3` marker that would separate them is written at the *end* of run-phase, after the interval `review` occupies. It is a live contradiction rather than a wording difference: on the day `REQ-KB-006` lands, a shipped always-loaded rule states the opposite. The rule's own boundaries mark the derivation as interim pending the store this SPEC defines, but that framing appears nowhere its reader will see it. Editing a shipped always-loaded rule is outside this SPEC's §C, so it is a reported finding with an owner to be assigned.
2. **What the backlog is.** `backlog` is a **column** here holding cards, and §A.2 contemplates one with no `spec.md` at all; `/moai todo` describes itself as "Not a board" and carries `spec_id: null` until dispatch. The two readings answer differently a question `REQ-KB-004` decides — whether a queued item has a card record. This one **may not be a defect**: a queue upstream of a `backlog` column is coherent, and nothing measured chooses between that and the flat reading. What is missing is a statement, not a decision, and inventing one would be fixing a run-phase question on preference.

Counts re-measured after the edits, in this worktree:

```
$ grep -cE '^\*\*REQ-KB-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KB-[0-9]{3}\*\*' acceptance.md
25
```

Both sequences contiguous with no duplicates (`REQ-KB-001` … `REQ-KB-025`, `AC-KB-001` … `AC-KB-025`; 25 unique ids each). Against the Tier L ceiling of 25 and 25: **both at the ceiling, neither with headroom.** The single spare requirement was spent on `REQ-KB-025`. No criterion was added: its carrier half was authored into `AC-KB-017`, which already reads the declaration in its runtime half, so the two observations share one subject and one fixture — `AC-KB-017` now serves two requirements, joining `AC-KB-001` and `AC-KB-015`.

`REQ-KB-025` was **not** folded into `REQ-KB-017` despite the family's widening precedent and despite the ceiling being tight, and the reason is stated so a later reader does not read it as budget carelessness. The two have different subjects — who may write the board, versus what session-topology datum exists and where it can be read from — and folding them would give `REQ-KB-017` two subjects, the F1 shape this SPEC exists to have repaired. They also fail apart in the direction that matters: an implementation can satisfy `REQ-KB-017`'s runtime half completely on a session-private carrier, which breaks two of `SPEC-KANBAN-WORKTREE-001`'s gates and is invisible to every observation the board itself makes.

**Lint, re-run per file after the edits** (`moai spec lint` invoked per path per `acceptance.md` §A.1 rule 3, never over a glob):

```
spec.md        ✓ No findings
plan.md        ✓ No findings
acceptance.md  ✓ No findings
design.md      ✓ No findings
research.md    ✓ No findings
progress.md    0 error(s), 1 warning(s)  — MissingExclusions, pre-existing
```

The `progress.md` warning is pre-existing and structural: this file has carried no `Out of Scope` section since v0.1.0, the rule is downgraded to a warning on grandfathered-era SPECs, and 0 errors is the gate. It is recorded rather than suppressed with `lint.skip`.

**One premise of the revision brief was contradicted by measurement and is recorded rather than smoothed.**

The brief stated that "the only `factory` occurrences in BOARD are `REQ-KB-001`'s own prohibition text and its verification surfaces — those are correct as they stand; do not 'fix' them." Measured, that is false for roughly half of them. Seven occurrences across four files were **live citations** of `internal/factory/record.go` and its `{".moai", "state", "factory"}` value — a path and a value that no longer exist on `main` — in `spec.md` §A.3(e) and §E, `plan.md` §B.8 / §C command 11 / the M1 persistence bullet / AP-24, `design.md` §B.1, and `acceptance.md` `AC-KB-002`. They are not prohibition text and not gate commands; they are the evidentiary basis of the path-collision decision, and left alone they would have sent a run-phase implementer to a file that is not there. They are repaired.

The brief's underlying instinct was right about the class it named: the `git ls-tree -d --name-only origin/main internal/factory` occurrences in `plan.md` §C, `acceptance.md` `AC-KB-001` and `research.md` §P.1 **are** correct as they stand and are deliberately untouched — a gate asserting the old path's *absence* must name the old path. So is `REQ-KB-001`'s prohibition. What the brief missed is that the same token appears in a third role, and the measurement is what separated them.

Two verbatim transcripts naming the pre-rename path are also preserved unchanged, in `spec.md` §A.4a and `research.md` §N: a recorded command-and-output edited to look current stops being evidence of anything. `research.md` §P records what the same commands report now, beside them rather than over them.

### v0.6.0 — plan-audit delta repair of v0.5.0

The independent plan audit of v0.5.0 returned **FAIL at 0.87** against the Tier L threshold of 0.85 — a FAIL despite a passing score, because the delta's headline claim did not hold as written: `REQ-KB-025` did not make `REQ-KB-017`'s runtime refusal implementable without forking `REQ-KS-006`. Four blocking findings, one optional. All five repaired **in place at zero requirement and zero criterion cost**.

Every finding was independently re-measured here before repair rather than relayed. The commands and outputs are below; all five confirmed.

| # | Finding | Repair |
|---|---|---|
| D1 (major) | **`REQ-KB-025` restated a narrowed contract while declaring it restated none.** It enumerated three properties and dropped `REQ-KS-006`'s fourth clause group — the declaration as *the key on which the lead selects a dispatch target (`REQ-KS-019`) and over which quorum is accounted (`REQ-KS-012`)* — then, in the same sentence, pushed dispatch routing and quorum accounting away as the sibling's. The two surviving directional clauses do not imply the dropped one: "resolvable from a session that is not the `lead`" is workers-read-lead; the dropped clause is lead-reads-workers. A **lead-only carrier** satisfied all three enumerated properties, passed every observation, and could not route — the fork §A.8 exists to prevent, arriving through an incomplete list rather than a second datum, and colliding with `REQ-KS-006`'s own "sole place the declaration contract is defined" | enumeration **deleted entirely**. `REQ-KB-025` now cites the contract whole and by reference and carries no list; the one distinction retained is **key vs mechanism** — this SPEC's carrier carries the key `REQ-KS-019` and `REQ-KS-012` read, and defines no part of the mechanisms. `spec.md` §A.8 rewritten, §D.14 gains the second cross-session direction, `AC-KB-017`'s carrier half gains the lead-reads-workers observation with its own positive control, `design.md` §C.1a gains the lead-only carrier as a named rejected alternative, `plan.md` D6 / M1 / AP-30 / new AP-33 |
| D2 (major) | **The adoption obligation was owned by no requirement on the side that must honour it.** §A.8 asserted in prose that a later-landing `SPEC-KANBAN-BOOTSTRAP-001` adopts this carrier — "one declaration in the system, or the borrow has failed" — with nothing behind it. `REQ-KB-025`'s second branch is structurally vacuous: that sibling declares this SPEC in its `dependencies:`, so it cannot land first and the adopt-if-landed branch is unreachable at this SPEC's run-phase. `AC-KB-017`'s uniqueness scan has the same shape — it runs at M1, before the sibling exists, so it can only ever find the declaration this SPEC wrote. Meanwhile BOOTSTRAP v0.6.0 recorded nothing: `REQ-KS-006` byte-unchanged, still reading "shall fix no carrier for it … is a run-phase decision", with no pointer to `REQ-KB-025`. An implementer reading only BOOTSTRAP found carrier choice explicitly free. **This is the shape §C names — an obligation each document expects the other to hold — for the fourth time in this family, introduced by the delta that cites the first three** | the obligation **landed on the enforceable side**. `REQ-KS-006` widened in place (BOOTSTRAP v0.7.0): where BOARD's run-phase has already established a carrier under `REQ-KB-025`, that implementation adopts it and defines no second declaration. `AC-KS-030` gains a fifth conjunct deciding it, with a positive control. Zero requirement and zero criterion cost to BOOTSTRAP — the same in-place move its own v0.3.0 and v0.5.0 made. **No `dependencies:` edge added**; the cycle §A.4a refuses is refused again. The unreachable branch and the M1 scan are kept against a future re-ordering and are explicitly marked as not-the-guard in `spec.md` §D.14 and `AC-KB-017` |
| D3 (major) | **`AC-KB-002`'s absence scan became undecidable in the rename this delta recorded.** It was decided by scanning "the board package" for a reference to the session-record constant, requiring none — executable while that constant lived in `internal/factory`, not now. v0.5.0 recorded the same-package fact and propagated it to `plan.md` AP-24 and the M1 persistence bullet, but not to the one criterion whose decision procedure it broke — v0.4.0's cross-file-drift shape running the other way | reformulated as a **property** scan: no board-state path is *constructed from* the session record's constant. The file-scoped name scan (board's own files excluding `record.go`) is offered as an equivalent alternative, the run-phase records which form it used, and a positive control is mandatory on either — without it the scan's zero is again indistinguishable from a scan that cannot fire. A distinct sub-package is noted as also restoring decidability and deliberately not mandated, since `REQ-KB-005` binds the constant and not the layout |
| D4 (major) | **Disagreement 2 was misattributed.** The "Not a board" text is `/moai todo`'s, and `kanban-dispatch.md`'s own column table lists `backlog` as a row — so on the single point it was cited as disagreeing on, that rule **agrees**. The internal inconsistency was itself a finding: §A.9's body and this file's body named `/moai todo` correctly, while `spec.md` §E, the v0.5.0 HISTORY entry, `plan.md` §B.9 and this file's section heading attributed it to the rule — precisely the surfaces a reader uses to locate a finding, inside a section whose stated discipline is precise citation | re-attributed in all four, plus a new §A.9 agreement row recording that the rule agrees on `backlog` being a column. **The substantive analysis is unchanged** — queue-upstream versus flat, both coherent, nothing measured choosing, `REQ-KB-004` owning the question. Only the filing moved, and the correction narrows the blast radius: a disagreement with an always-loaded rule reaches every session, one with a queue surface does not |
| D5 (optional — **taken**) | §A.9's agreement row cited the rule as corroborating "only the lead records a transition; a dispatch confers no write authority". It states neither phrase; its nearest support is `sync → done`-scoped | row narrowed to what the rule actually states (the lead reads the sync session's evidence and records the terminal transition itself, for that step only), with the sole-writer property marked **inferred** from the rule's silence rather than cited from its text |

D6 (`SPEC-KANBAN-MULTISESSION-001` has no directory) needed no action and none was taken: both references correctly call it the superseded predecessor and neither is a live edge.

**One finding beyond the brief, found by applying D1's own reasoning and repaired with it.** D1 was raised against `REQ-KB-025`'s enumeration, and the repair deleted it. A grep afterwards found the identical 3-of-4 list surviving in **two further surfaces** — §A.8's own layer table (the "contract" row) and §C's role-declaration exclusion — both of which listed *exists / distinct from the label / not derived / resolvable from a non-`lead` session* and both of which dropped the dispatch-selection and quorum-accounting key exactly as the requirement had:

```
$ grep -c 'distinct from the launch label' spec.md
2                                   # after the REQ-KB-025 repair, before this one
```

Repairing only the requirement would have left the narrowing in the two places a reader goes to *understand* the borrow, which is where a summary does the most damage: an implementer who reads §A.8's table and never opens `REQ-KS-006` builds the lead-only carrier the requirement no longer licenses. Both now carry a whole-contract pointer and no list, and each records why it carries none. The general rule is in `plan.md` AP-33 — cite the contract, never summarize it — and it binds prose surfaces, not only normative ones.

**Re-measurement of the brief, before repair.**

```
$ grep -n '^\*\*REQ-KS-006\*\*' BOOTSTRAP/spec.md      # D1: the fourth clause group
… "that declaration being the key on which the lead selects a dispatch target
   (REQ-KS-019) and over which quorum is accounted (REQ-KS-012)" …   → present, and absent from REQ-KB-025

$ grep -n '^dependencies:' BOOTSTRAP/spec.md            # D2: the branch is unreachable
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
$ grep -rn 'REQ-KB-025' .moai/specs/SPEC-KANBAN-BOOTSTRAP-001/
(no matches)                                            → BOOTSTRAP recorded nothing

$ grep -n '^module:' BOARD/spec.md                      # D3: the scan is undecidable
11:module: internal/kanban
$ grep -c 'stateDirSegments' internal/kanban/record.go
3

$ grep -rin 'not a board' kanban-dispatch.md todo.md    # D4: the attribution
todo.md:94:- **Not a board.** Column position lives with the lead and the SPEC status, not
$ grep -n 'backlog' kanban-dispatch.md | head -3
25:| `backlog` | *none* — a queue | …        → the rule lists backlog as a column: it AGREES

$ grep -in 'write authority\|only the lead' kanban-dispatch.md   # D5: the overclaim
(no matches)
```

**Counts re-measured after the edits:**

```
$ grep -cE '^\*\*REQ-KB-[0-9]{3}\*\*' spec.md
25
$ grep -cE '^\*\*AC-KB-[0-9]{3}\*\*' acceptance.md
25
$ grep -cE '^\*\*REQ-KS-[0-9]{3}\*\*' ../SPEC-KANBAN-BOOTSTRAP-001/spec.md
25
$ grep -cE '^\*\*AC-KS-[0-9]{3}\*\*' ../SPEC-KANBAN-BOOTSTRAP-001/acceptance.md
30
```

BOARD unchanged at **25 / 25**, both at the Tier L ceiling; BOOTSTRAP unchanged at **25 / 30**. Nothing was added, retired, merged, or renumbered on either side — every repair is an amendment in place, which is what made a four-finding FAIL closable without touching a budget that has no headroom left.

**Nothing in the repair brief was contradicted by measurement.** All four blocking findings and the optional fifth were independently re-verified above before any edit, and every one held exactly as stated, including the internal-inconsistency inventory of D4 (four BOARD-side surfaces wrong, the BOOTSTRAP §B.4 closure already correct). This is recorded because the practice runs both ways: the v0.5.0 brief carried a premise that measurement overturned, and this one did not.

### v0.6.1 / v0.6.2 — re-audit PASS, three minors closed, plan-phase final state

The re-audit of v0.6.0 returned **PASS at 0.97**. D1-D5 all CLOSED, no regression in anything previously cleared, counts held, no cycle. The auditor specifically confirmed D2 did not survive as a partial fix: the `REQ-KS-006` widening is a verbatim prefix extension (old text unchanged), normative on the BOOTSTRAP side, and observed by `AC-KS-030`'s fifth conjunct whose positive control fires on nothing else — the finding most likely to be half-closed, closed properly.

Three new minors, all non-blocking and all one-line, closed here at zero requirement and zero criterion cost.

| # | Defect | Closure |
|---|---|---|
| N1 | BOOTSTRAP `acceptance.md` frontmatter `updated: 2026-08-11` while `version:` went 0.5.0 → 0.7.0 and the body changed today; its two siblings both read `2026-08-14` | date corrected to `2026-08-14`; file bumped to **0.7.1** |
| N2 | `AC-KB-017`'s "What this criterion does not decide" paragraph said the observations "assert its **three** properties" while the carrier half now carries four — a leftover of the D1 repair that added the lead-reads-workers direction. **The same count survived in two further surfaces** (see below) | corrected to **four** at all three sites, named individually (workers-read-lead, lead-reads-workers, label-distinctness, uniqueness), each with an in-place note recording the old value. Why the count matters: a criterion that miscounts its own observations invites a run-phase to run one fewer |
| N3 | The narrowed v0.5.0 summary of the anti-fork clause survived in orientation surfaces — `spec.md` §E's `REQ-KS-006` bullet and `plan.md`'s invariant-table row both still framed *resolvable-from-a-non-`lead`-session* as the mechanism, which post-D1 understates it by half. Neither is normative, and `REQ-KB-025` directs the reader to the contract rather than to any list, so the normative layer was protected — but these are the surfaces an implementer skims first, and the hazard is the one `AP-33` names | **pointer form** taken at both, per the auditor's own note that a pointer is more robust against recurrence: §E's bullet now states none of the contract's properties and says so; `plan.md`'s row now reads "must satisfy that contract **whole — read it there, not from this row** (AP-33)". Widening the summaries to name both directions was the alternative and was rejected — a two-of-four list is the same defect as a three-of-four list one item later, and the next clause added to `REQ-KS-006` would re-open it |

**N2's two mirror surfaces, closed at v0.6.2 as part of the same defect rather than as new findings.** The v0.6.1 fix corrected the count at `acceptance.md:165` and left the same claim standing in two places that describe that very criterion, so the documents disagreed on how many properties `AC-KB-017` asserts — N2 exactly, one file over, and the same cross-file-drift shape as D3:

| Surface | Was | Now |
|---|---|---|
| `spec.md` §D.14 — the verification surface describing `AC-KB-017` | "asserts the three properties, not the choice" | "asserts the **four** properties — workers-read-lead, lead-reads-workers, label-distinctness, and uniqueness — not the choice", plus the in-place note |
| `acceptance.md` §K — the Out-of-Scope entry for the carrier | "judges its three contract properties" | "judges its **four** contract properties", named, plus the in-place note |

The second was **not named in the brief**; it surfaced only because the sweep covered the whole artifact set rather than the two files the brief pointed at. Both now name the four properties individually instead of carrying a bare count, which is what makes a future addition visible as a mismatch rather than as a silently-stale number. `design.md` and `research.md` were swept and are clean — their only occurrences are historical narrative about the D1 defect ("v0.5.0 enumerated three properties"), correct as written. So is `plan.md` §E D6. `spec.md` §E and `plan.md`'s invariant row are clean because N3 replaced their summaries with pointers.

```
$ grep -rn 'three contract propert\|asserts the three propert\|its three propert' *.md
acceptance.md:254   # the corrective note's own quotation of the old text — the only match
```

**A third N3 site, found by grep and closed with the other two.** `REQ-KB-017` itself carried a two-of-four recital — *"a datum distinct from its launch label and resolvable by a session that is not the `lead`"* — inside a sentence ending "and shall not restate that contract's content". Same class as N3, same self-contradiction D1 repaired in `REQ-KB-025`, and normative rather than orientation. It predates every finding (v0.3.0) and no finding named it; it is closed here rather than left, because the whole content of N3 is that a partial list is the fork's mechanism and leaving one in a **requirement** while removing them from prose would invert the priority. The clause is replaced by a whole-and-by-name pointer; nothing else in `REQ-KB-017` moves.

```
$ grep -c 'distinct from the launch label' spec.md
0        # after; the only surviving occurrence in the SPEC is progress.md's own grep transcript
```

**Two corrections the lead recorded against its own brief**, preserved here because this family records contradicted premises rather than smoothing them, and because both were the lead's own measurement errors rather than defects in the work: (1) BOOTSTRAP `acceptance.md` was reported untouched and was in fact modified and at v0.7.0 — a held AC count was taken as evidence the file was unedited, and the edit dismissed was the load-bearing half of the D2 closure; (2) three `not a board` + `kanban-dispatch` colocations were reported where there are seven, the grep having required both tokens on one line. The substantive read on (2) — that all of them are recorded measurements, HISTORY entries, or the correction itself, none asserting the wrong attribution — was confirmed correct.

---

## Plan-phase final state

**Audit history**: FAIL 0.87 (four blocking findings) → repair → **PASS 0.97** (three one-line minors) → closed. Dimension scores at the passing audit are the auditor's record; what this file records is that traceability was 1.00, all seven must-pass checks passed, every HISTORY measurement re-verified true, and no cycle was introduced or re-introduced at any point.

**Findings**: D1 (narrowed contract restatement) · D2 (unenforceable adoption obligation) · D3 (undecidable absence scan) · D4 (misattributed disagreement) · D5 (overclaimed corroboration) — all **CLOSED**. N1 (stale `updated:`) · N2 (miscounted properties, **three surfaces**: `acceptance.md` §165 + `spec.md` §D.14 + `acceptance.md` §K) · N3 (surviving narrowed summary, **three sites**: `spec.md` §E + `plan.md` invariant row + `REQ-KB-017`) — all **CLOSED**.

**A pattern worth naming, since it recurred three times in this SPEC.** D3, N2 and N3 are the same shape: a fact was corrected in one surface and left standing in the surfaces that mirror it. D3 was the rename propagated to `plan.md` but not to the criterion it broke; N2 was a count corrected in the criterion but not in the two places describing it; N3 was an enumeration removed from the requirement but not from the prose that summarized it. In every case the first fix was correct and the sweep was too narrow. The mitigation that actually worked was grepping the **whole artifact set** rather than the surfaces a brief pointed at — which is how the second N2 site and the third N3 site were found, neither having been named by any audit or brief.

**Counts — both at the Tier L ceiling, zero headroom:**

```
$ grep -cE '^\*\*REQ-KB-[0-9]{3}\*\*' spec.md                        → 25   (ceiling 25)
$ grep -cE '^\*\*AC-KB-[0-9]{3}\*\*'  acceptance.md                  → 25   (ceiling 25)
$ grep -cE '^\*\*REQ-KS-[0-9]{3}\*\*' ../SPEC-KANBAN-BOOTSTRAP-001/spec.md       → 25
$ grep -cE '^\*\*AC-KS-[0-9]{3}\*\*'  ../SPEC-KANBAN-BOOTSTRAP-001/acceptance.md → 30
```

Tier L is the top tier, so there is no promotion available. **Any future finding requiring a new requirement or a new observation must be escalated to the orchestrator rather than absorbed**, and re-bundling an existing entry to make room is prohibited — that is the F1 defect the v0.2.0 repair closed and would be reintroduced by the act of making room. Every repair from v0.5.0 onward was an amendment in place for exactly this reason.

**Debt a run-phase implementer inherits.** Each item below is recorded deliberately and is not a defect to fix before run-phase:

| Debt | Where it is recorded | What it costs |
|---|---|---|
| `.claude/rules/moai/workflow/kanban-dispatch.md` re-derives column position from SPEC `status` after a `/clear`, which `REQ-KB-006` forbids | `spec.md` §A.9 Disagreement 1, `plan.md` §B.9 + AP-32 | On the day `REQ-KB-006` lands, a shipped **always-loaded** rule contradicts it and reaches every session. AP-32 frames the consequence correctly: it is a reportable finding, never a licence to implement the derivation. Reconciling the rule is outside this SPEC's §C and needs an owner |
| Whether a `/moai todo` queue item occupies the `backlog` **column** or sits upstream of the board | `spec.md` §A.9 Disagreement 2; `SPEC-KANBAN-BOOTSTRAP-001` §C + `plan.md` §B.4 | Undecided by measurement and coherent both ways. `REQ-KB-004` owns the question; both SPECs record it open rather than one of them inventing an answer |
| `AC-KB-023`'s TOCTOU residual — the interval between the pre-removal re-read and the unlink | `spec.md` §A.7(3a), §D.12; `AC-KB-023`'s closing paragraph | Not closable at this layer (`unlink(2)` takes a path). The criterion decides the **mitigation** and states the residual rather than testing for its absence |
| Cross-SPEC declaration uniqueness between BOARD's M1 and BOOTSTRAP's run-phase | `spec.md` §D.14, `AC-KB-017`'s uniqueness note, `AC-KS-030`'s adoption conjunct | BOARD's M1 scan clears BOARD's own implementation only — the sibling has not landed when it runs. The cross-SPEC half is decided on BOOTSTRAP, and the window between the two is real |

**Plan phase is complete.** Artifacts are audit-passed at 0.97 and unmodified since; the SPEC is ready for Implementation Kickoff Approval and run-phase entry.

## §E.2 Run-phase Evidence

### M0 — preflight measurements (plan.md §C, run 2026-08-14 in worktree `~/.moai/worktrees/kanban-board`, base `origin/main` = `a301ef6f8`)

**Prerequisite gate (cmd 1, REQ-KB-002 / AC-KB-001) — PASS.** Decided against `origin/main`, not the working tree:

```
$ git ls-tree -d --name-only origin/main internal/kanban
internal/kanban
$ git ls-tree -d --name-only origin/main internal/factory
(no output)
```

The exact pair AC-KB-001 requires. The prerequisite SPEC's frontmatter still reads `status: in-progress` — bookkeeping lag, per plan.md §B.1; the base-branch query is the deciding observation.

**Single-origin discriminant (cmd 2/2b) — both checkouts, expectations held.** From the worktree: bare `--git-common-dir` already absolute; from the **primary** (`/Users/goos/MoAI/moai-adk-go`): the bare form returns the **relative** `.git` while both absolute forms return the full path — the asymmetry REQ-KB-005 exists against, re-confirmed:

```
# worktree ($W = ~/.moai/worktrees/kanban-board)
$ git -C $W rev-parse --git-common-dir                          → /Users/goos/MoAI/moai-adk-go/.git (absolute)
# primary ($P = /Users/goos/MoAI/moai-adk-go)
$ git -C $P rev-parse --git-common-dir                          → .git              # RELATIVE — never used alone
$ git -C $P rev-parse --path-format=absolute --git-common-dir   → /Users/goos/MoAI/moai-adk-go/.git

# cmd 2b — the branch_guard.go single-probe form (parent of line 2 = board root):
$ git -C $W rev-parse --path-format=absolute --git-dir --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git/worktrees/kanban-board
/Users/goos/MoAI/moai-adk-go/.git
$ git -C $P rev-parse --path-format=absolute --git-dir --git-common-dir
/Users/goos/MoAI/moai-adk-go/.git
/Users/goos/MoAI/moai-adk-go/.git
```

**Recorded primary root for M1's resolution test (REQ-KB-005 positive half):** `/Users/goos/MoAI/moai-adk-go` — the parent of the absolute common git directory. `git version 2.50.1 (Apple Git-155)` ≥ 2.31, so the primary probe path applies; the older-git fallback executes zero times in the wild here and is exercised only through the `execCommand` indirection (cmd 13 confirmed: `var execCommand = exec.Command` at `internal/hook/branch_guard.go:29`; test comment "direct invocation of the fallback is INSUFFICIENT" at `branch_guard_test.go:108`).

**Remaining §C commands — all expectations held:**

| cmd | expectation | measured |
|---|---|---|
| 3 | consumer probe near lines 178/190 | `runGitRevParse(projectDir, "--path-format=absolute", "--git-dir", "--git-common-dir")` at line 178; `--absolute-git-dir` fallback at line 190 ✓ |
| 4 | 8-value enum from schema | `draft, planned, in-progress, implemented, completed, superseded, archived, rejected` ✓ |
| 4b | 17 `planned` / 42 terminal | 17 / 42 ✓ |
| 5 | both gitignore lines | `**/.moai/state/` (line 212) + `.moai/state/` (line 280) ✓ (line drift from plan's 207/275; content identical) |
| 6 | two lock substrates | `internal/spec/lock{,_unix,_windows}.go` + `internal/lockfile/` present ✓ |
| 7 | atomicfile `Replace` | rename-step-only primitive, caller owns temp lifecycle ✓ |
| 8 | sole-writer text present | `spec.md` count = 8 (was 0 at v0.1.0) ✓ |
| 9 | `func isPrimaryCheckout(projectDir string) (bool, error)` | at `branch_guard.go:167`, unexported, boolean ✓ |
| 9b | extraction target | `internal/core/git/doc.go`: "Git repository operations for MoAI-ADK"; `grep internal/hook\|internal/cli` over the package → **0** ✓ |
| 10 | branch-side blob read, no checkout | `git show feat/SPEC-KANBAN-BOARD-001:.moai/specs/SPEC-KANBAN-BOARD-001/spec.md \| grep -m1 '^status:'` → `status: draft` ✓ |
| 11 | collision state | `internal/kanban/record.go:43` `stateDirSegments = []string{".moai", "state", "kanban"}`; worktree `.moai/state/` carries no `kanban`/`kanban-board` (gitignored dir absent in fresh worktree); primary holds both occupants (session records + `backlog.json`) ✓ |
| 11b | three landed surfaces + disagreement text | all three files present; "re-derived from SPEC status after a clear" at `kanban-dispatch.md:97` ✓ |
| 11c | no landed role-declaration carrier | `grep -rln 'declared role\|role declaration' internal/` → **0 files** — REQ-KB-025's **establish** branch applies. `REQ-KS-006` contract read in full from BOOTSTRAP `spec.md` (see M1) ✓ |
| 12 | stale-lock gap | `lock_windows.go` header: "Stale lock detection … post-MVP enhancement … manual `del`"; `lock_unix.go:23` "Close releases the flock atomically"; primary `.moai/state/` carries **14** orphaned zero-length `spec-close-*.lock` artifacts, all inert ✓ |

**Bootstrap-sibling artifact state (measured, affects nothing in this run-phase):** the primary checkout's BOOTSTRAP spec/plan/acceptance carry uncommitted v0.7.x edits — the D2 adoption widening progress.md v0.6.0 records. `grep -c 'REQ-KB-025'` over **origin/main's** BOOTSTRAP `spec.md` → 0, i.e. the widening is uncommitted. The BOARD run-phase consumes only the REQ-KS-006 **contract body**, which the v0.7.0 widening extends as a verbatim prefix (old text unchanged, per the v0.6.1 re-audit) — the unmodified origin/main text is therefore the whole contract for this SPEC's purposes, and the adoption obligation itself is BOOTSTRAP's run-phase concern (debt table, row 4).

### M1 — state store, resolution, sole writer, board-wide lock, bounded recovery (run 2026-08-14, worktree `~/.moai/worktrees/kanban-board`, TDD RED-GREEN-REFACTOR)

**Implemented** (all beneath `internal/`, worktree-local):

| file | content |
|---|---|
| `core/git/checkout.go` (+`checkout_test.go`) | REQ-KB-005 extraction: `ResolveGitDirs`/`IsPrimaryCheckout` returning the resolved PATHS; older-git fallback intact; exported `ExecCommand` indirection preserved so the fallback is forced through the dispatcher |
| `hook/branch_guard.go` (+ both test files) | re-point: `isPrimaryCheckout` keeps its boolean contract and delegates to the extraction; the hook fallback suite swaps `gitcore.ExecCommand` (the indirection moved WITH the logic) |
| `kanban/board.go` (+test) | board root/dir/path via the board's OWN `boardDirSegments = {".moai","state","kanban-board"}` (AP-24: record.go's `stateDirSegments` untouched); minimal Card/BoardState schema with the column as a RECORDED value (REQ-KB-006); absent-vs-unreadable `LoadBoard` (REQ-KB-021/013 detection) |
| `kanban/role.go` (+test) | REQ-KB-025 carrier: `DeclareRole`/`ResolveDeclaredRole` at the board root (single origin), label recorded as a separate datum, both cross-session directions resolvable |
| `kanban/board_store.go` (+test) | THE guarded write entry `WriteBoardState`: lead-role guard (fail-closed) → board-wide lock across the whole read-modify-write → same-directory temp + `atomicfile.Replace` (REQ-KB-017/018/019); minimal WIP-2 `TransitionIntoRun` with named refusal (REQ-KB-009 minimal form, for AC-KB-019) |
| `kanban/board_lock.go` + `_unix`/`_windows` + `lock_alive_*` (+tests) | REQ-KB-019 substrate reusing the `internal/spec/lock.go` pattern (flock / atomic-create; `internal/lockfile` untouched); owner identity recorded IN the artifact; REQ-KB-023 bounded `ClearStaleBoardLock` with pre-unlink re-read, TOCTOU residual stated in the doc comment (AP-29) |
| `kanban/board_recover.go` (+test) | REQ-KB-013/022 bounded recovery: sole writer + board-wide lock; readable board recovered in place (no card moved); unreadable board's raw content preserved in a durable sidecar BEFORE the empty replacement; one invocation, one verdict |

**E1 — AC binary matrix (M1-executable criteria).** Verbatim `-run` outputs below; `ok github.com/modu-ai/moai-adk/internal/kanban` on every run.

| AC | verdict | deciding observation |
|---|---|---|
| AC-KB-002 | PASS | `TestBoardRoot_SingleOrigin` (worktree AND primary-checkout runs, the primary being the load-bearing control) + `TestBoardRoot_FallbackForcedThroughIndirection` (third run, primary probe forced to fail via `gitcore.ExecCommand`, non-fallback calls delegated to real git) — all PASS; property scan + positive control below |
| AC-KB-003 | PASS | absence scan over the five board files: `grep -n 'os\.Getwd\|filepath.Join("."' board*.go role.go` → rc=1 (zero); positive control below |
| AC-KB-004 | PASS | `TestLoadBoard_ColumnRecordedNotDerived` — recorded `review` read as `review`; mutating ONLY the recorded column to `run` changes the answer |
| AC-KB-005 | PASS | scan (frontmatter writes + status literals) → rc=1 zero matches over board files; positive control below; no `factory` token in any authored line (AC-KB-001 standing, checked at commit) |
| AC-KB-006 | PASS | `TestLoadBoard_AbsentVsUnreadable` truncated row → unknown, no board returned; same-size well-formed row reads (conditional-on-content control) |
| AC-KB-017 | PASS | static half: `BoardPath` write sites enumerate to exactly `board_store.go:172` (inside `writeBoardAtomic`, reachable only from the two guarded entries); runtime half: non-lead refused + byte-unchanged, lead succeeds, undeclared refused (fail-closed); carrier half: workers-read-lead + lead-reads-workers + cross-process resolution + label-distinctness all PASS; uniqueness scan below |
| AC-KB-018 | PASS | static dir-equality: `os.CreateTemp(dir, ...)` where `dir == filepath.Dir(target)` (`board_store.go:167`, self-evident in code); dynamic: `TestWriteBoardState_ConcurrentReaderSeesWholeBoards` — 60 concurrent writes, reader never observed a parse failure or a shrinking board; no temp leftovers |
| AC-KB-019 | PASS | `TestBoardMutation_SerializedAcrossProcesses` — two separate OS processes (helper re-exec), two DIFFERENT cards, board holding one: exactly 1 success + 1 WIP refusal, final run count 2; positive control `TestBoardMutation_ConcurrencyPositiveControl` — zero-baseline, both succeed, count 2 |
| AC-KB-020 | PASS | `TestRecoverBoard_ReadPathNeverRepairs` (5 reads, all unknown, bytes unchanged) + `TestRecoverBoard_BoundedRecovery` (verdict `replaced`, sidecar preserves the lost raw content, roles carrier untouched, second invocation yields `recovered` and modifies nothing) + readable-board and non-lead-refused rows |
| AC-KB-023 | PASS | three observations in separate processes: dead owner cleared (report carries the PID), live owner refused (age irrelevant), release-and-reacquire interleaving ahead of the re-read → `ErrBoardLockChangedHands`, artifact survives. Platform recorded in test output: **darwin (Unix substrate)** — flock releases on exit, so observation 1 is trivially satisfiable here; the Windows substrate is the requirement's reason, covered by `GOOS=windows` build. TOCTOU residual stated in `ClearStaleBoardLock`'s doc comment, not claimed closed |
| AC-KB-024 | PASS | ONE table-driven test, both rows: absent → empty board + success; unreadable (truncated, permission-denied) → unknown. The pairing is the criterion; the same-size well-formed row is the conditional control |
| AC-KB-021 (structural) | PASS (M1 half) | absent-row + sole-writer-creates-directory both exercised (`TestWriteBoardState_CreatesStateDirectory`); the full no-recovery-required reading rides AC-KB-024's absent row |
| DEFERRED to M2 | — | AC-KB-007/008/009/010/011/012/013/014/021(full)/022/025 — column enum, full card model, compatibility table, unheld state, status-source selection all structurally wait on M2's model (traceability table assigns them M2) |

**Verbatim outputs (per-AC `-run`, exit 0 each):**

```
$ go test ./internal/kanban/ -run 'TestBoardRoot_SingleOrigin|TestBoardRoot_FallbackForcedThroughIndirection' -count=1 -v | grep -E '^--- '
--- PASS: TestBoardRoot_FallbackForcedThroughIndirection (0.16s)
--- PASS: TestBoardRoot_SingleOrigin (0.41s)
ok  	github.com/modu-ai/moai-adk/internal/kanban	7.344s

$ go test ./internal/kanban/ -run 'TestLoadBoard_AbsentVsUnreadable' -count=1 -v | grep -E '^(--- |    --- )'
--- PASS: TestLoadBoard_AbsentVsUnreadable (0.00s)
    --- PASS: TestLoadBoard_AbsentVsUnreadable/absent_file_is_legitimately_empty_board (0.00s)
    --- PASS: TestLoadBoard_AbsentVsUnreadable/truncated_json_is_unknown (0.00s)
    --- PASS: TestLoadBoard_AbsentVsUnreadable/well_formed_same_size_reads_successfully (0.00s)
    --- PASS: TestLoadBoard_AbsentVsUnreadable/permission_denied_is_unknown (0.00s)

$ go test ./internal/kanban/ -run 'TestWriteBoardState_NonLeadRefused|TestWriteBoardState_LeadSucceeds|TestWriteBoardState_UndeclaredSessionRefused|TestRoleDeclaration_' -count=1 -v | grep -E '^--- '
--- PASS: TestRoleDeclaration_CrossProcessResolution (0.05s)
--- PASS: TestWriteBoardState_UndeclaredSessionRefused (0.00s)
--- PASS: TestRoleDeclaration_LabelDistinct (0.02s)
--- PASS: TestRoleDeclaration_WorkersReadLead (0.03s)
--- PASS: TestWriteBoardState_NonLeadRefused_BoardUnchanged (0.05s)
--- PASS: TestRoleDeclaration_LeadReadsWorkers (0.05s)
--- PASS: TestWriteBoardState_LeadSucceeds (0.07s)

$ go test ./internal/kanban/ -run 'TestBoardMutation_|TestBoardLock_ExcludesAcrossProcesses' -count=1 -v | grep -E '^--- '
--- PASS: TestBoardMutation_SerializedAcrossProcesses (0.01s)
--- PASS: TestBoardMutation_ConcurrencyPositiveControl (0.02s)
--- PASS: TestBoardLock_ExcludesAcrossProcesses (0.04s)

$ go test ./internal/kanban/ -run 'TestClearStaleBoardLock' -count=1 -v | grep -E '^(--- |    board_lock)'
    board_lock_test.go:149: platform: darwin-or-linux — Unix substrate releases flock(2) on exit, so this observation is trivially satisfiable here; the Windows substrate is the reason REQ-KB-023 exists
--- PASS: TestClearStaleBoardLock_DeadOwnerCleared (0.03s)
--- PASS: TestClearStaleBoardLock_LiveOwnerRefused (0.02s)
--- PASS: TestClearStaleBoardLock_ReacquireRaceAborts (0.07s)
--- PASS: TestClearStaleBoardLock_UnreadableArtifact (0.03s)
--- PASS: TestClearStaleBoardLock_NoArtifactAndUnparseable (0.06s)

$ go test ./internal/kanban/ -run 'TestRecoverBoard' -count=1 -v | grep -E '^--- '
--- PASS: TestRecoverBoard_ReadPathNeverRepairs (0.04s)
--- PASS: TestRecoverBoard_AbsentNeedsNoRecovery (0.06s)
--- PASS: TestRecoverBoard_NonLeadRefused (0.09s)
--- PASS: TestRecoverBoard_ReadableBoardRecoveredUnchanged (0.09s)
--- PASS: TestRecoverBoard_BoundedRecovery (0.09s)
```

**Static scans + positive controls (each run once and recorded):**

```
# AC-KB-002 property scan — no board-state path constructed from the record's constant.
# Form chosen: name scan scoped to the board's own files, code lines only
# (board.go's AP-24 COMMENT mentions stateDirSegments; a mention is not a use).
$ grep -n 'stateDirSegments' internal/kanban/board.go board_store.go board_lock.go board_recover.go role.go | grep -vE ':[0-9]+:\s*//'
(no output; rc=1 — zero code-line matches)
$ grep -c 'stateDirSegments' internal/kanban/record.go
3
# positive control: the SAME scan over the constant's real user reports 3 —
# the scan demonstrably fires.

# AC-KB-003 — no tree-relative anchor.
$ grep -n 'os\.Getwd\|filepath.Join("."' internal/kanban/{board,board_store,board_lock,board_recover,role}.go
(no output; rc=1)
# positive control: a deliberate os.Getwd()-anchored board path added in a
# temp file is REPORTED by the same scan (zz_positive_control_tmp.go:10), then removed.

# AC-KB-005 — no frontmatter write, no status literal.
$ grep -nE '\.moai.{0,3}specs|"status"|status:' internal/kanban/{board,board_store,board_lock,board_recover,role}.go
(no output; rc=1)
# positive control: a deliberate "status: in-progress" write is REPORTED by
# the same scan, then removed.

# AC-KB-017 static half — every write site reaches the file through one entry.
$ grep -n 'BoardPath' internal/kanban/{board,board_store,board_recover,role}.go
board.go:99:   os.ReadFile(BoardPath(root))                 # read
board_store.go:172: atomicfile.Replace(tmpName, BoardPath(root))  # THE write (writeBoardAtomic)
board_recover.go:78: os.ReadFile(BoardPath(root))           # read
# writeBoardAtomic is unexported, called only by WriteBoardState and
# RecoverBoard — both guarded entry points.
# positive control: a deliberate direct os.WriteFile(BoardPath(...)) in a
# temp file is REPORTED by the same enumeration, then removed.

# AC-KB-017 carrier uniqueness — exactly one declaration surface.
$ grep -rn 'func DeclareRole\|func ResolveDeclaredRole' internal/ | grep -v _test.go
internal/kanban/role.go:63:func DeclareRole(...)
internal/kanban/role.go:104:func ResolveDeclaredRole(...)
# one surface (the roles/ carrier beneath the board root); no other
# declaration mechanism exists in internal/ (M0 cmd 11c re-verified below).
# Clears THIS SPEC's implementation only; the cross-SPEC half is AC-KS-030's.

# REQ-KB-025 branch re-check (plan §C cmd 11c, re-run before authoring):
$ grep -rln 'declared role\|role declaration' internal/ 2>/dev/null
internal/kanban/board.go internal/kanban/board_store.go internal/kanban/role.go   # all M1-authored
# no pre-existing carrier — the establish branch applied.
```

**E8 — TDD RED evidence (verbatim, captured BEFORE each green):**

```
$ go test ./internal/core/git/ -run 'TestResolveGitDirs|TestIsPrimaryCheckout'   # before checkout.go
internal/core/git/checkout_test.go:39:15: undefined: ResolveGitDirs
internal/core/git/checkout_test.go:127:10: undefined: ExecCommand
FAIL github.com/modu-ai/moai-adk/internal/core/git [build failed]

$ go vet ./internal/kanban/   # before board.go
vet: internal/kanban/board_test.go:308:51: undefined: BoardState

$ go vet ./internal/kanban/   # before role.go
vet: internal/kanban/kanban_helper_test.go:26:16: undefined: ResolveDeclaredRole

$ go vet ./internal/kanban/   # before board_lock.go
vet: internal/kanban/board_lock_test.go:109:15: undefined: AcquireBoardLock

$ go vet ./internal/kanban/   # before board_store.go
vet: internal/kanban/board_lock_cross_test.go:37:10: undefined: ErrWipLimitExceeded

$ go vet ./internal/kanban/   # before board_recover.go
vet: internal/kanban/board_recover_test.go:77:14: undefined: RecoverBoard
```

Mid-green failures also observed and fixed (not smoothed): the worktree-fixture
symlink mismatch (`/var` vs `/private/var`), the two cross-process transitions
bouncing off the non-blocking lock (fixed by a bounded acquisition retry so
both processes reach the WIP decision in turn — the AC-KB-019 outcome), and
the unwritable-dir test being defeated by an earlier MkdirAll.

**E2 — builds.**

```
$ go build ./...            ; echo rc=$?
rc=0
$ GOOS=windows GOARCH=amd64 go build ./... ; echo rc=$?
rc=0
```

**E3 — coverage (this run, this tree, HEAD after M1 commits).**

```
$ go test -cover ./internal/kanban/... ./internal/core/git/...
ok  github.com/modu-ai/moai-adk/internal/kanban    1.047s  coverage: 86.0% of statements
ok  github.com/modu-ai/moai-adk/internal/core/git  78.428s coverage: 86.7% of statements
```

**E4 — subagent boundary.**

```
$ grep -rn 'AskUserQuestion' internal/kanban internal/core/git | grep -v _test.go
(no output; rc=1)
```

**E5 — lint.** Baseline captured BEFORE any edit: `golangci-lint run` → `0 issues.` After M1: `0 issues.` — NEW findings: zero.

**E6 — commits (worktree `feat/SPEC-KANBAN-BOARD-001`).**
- `dc03459bf` — extraction + hook re-point + status flip (draft → in-progress)
- `<this commit>` — board store, carrier, lock, recovery, tests, scans, evidence

Nothing outside the worktree was touched: all file operations used
`/Users/goos/.moai/worktrees/kanban-board` paths; the shared primary
checkout (`/Users/goos/MoAI/moai-adk-go`) was never written, staged, or
committed to.


**Full suite (REQ-KB-016 discipline, run at M1 close).**

```
$ go test ./... -count=1 > /tmp/kb-fullsuite.log 2>&1
rc=1 — two packages FAIL: internal/cli, internal/hook
```

Both failures attribute to ENVIRONMENT contamination, not to this change —
proven by rerun with the contaminating variables removed:

```
$ env -u MOAI_KANBAN_LABEL -u MOAI_KANBAN_ID -u MOAI_KANBAN_SETTINGS_INJECTED \
      -u CLAUDE_CODE_STOP_HOOK_BLOCK_CAP \
      go test ./internal/hook/ ./internal/cli/ -count=1
ok  github.com/modu-ai/moai-adk/internal/cli	212.015s
ok  github.com/modu-ai/moai-adk/internal/hook	17.133s
```

Root cause: this run-phase executes INSIDE a Kanban Mode session, whose
launcher exports `MOAI_KANBAN_LABEL`/`MOAI_KANBAN_ID`/`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`
into the process environment. The hook's session-start kanban notice reads
`MOAI_KANBAN_ID` (so the "AdditionalContext must be empty" negative controls
saw a live "Kanban Mode: joined run tjqim9" notice) and the launcher's
block-cap tests see the pre-set `CLAUDE_CODE_STOP_HOOK_BLOCK_CAP=200`. The
same tests pass once the inherited env is unset — the failures are a property
of WHERE the suite ran, not of WHAT changed (the same environment-dependence
class this repository has recorded before for the local cli suite). Every
other package in the full run was `ok`. No new lint findings; gofmt clean.

**E7 — blockers: none.** Four inherited debts left as recorded (not fixed):
kanban-dispatch.md column-derivation disagreement (AP-32 — reported finding,
rule untouched), the backlog column-vs-queue statement, the v3.1.0-rc binary
lag, and the BOOTSTRAP adoption obligation (sibling's run-phase concern).

_<M1–M3 evidence appended below as milestones complete>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Gate order.** Implementation Kickoff Approval was granted by the operator before this entry (kanban dispatch, run session `tjqim9`, 2026-08-14: "Implementation Kickoff Approval: GRANTED by the operator this turn"). Progression mode: **autonomous** (M0 → M3 without stopping at milestone boundaries). Phase 1 plan-audit re-execution: skipped as eligible — verdict PASS 0.97 ≥ Tier L threshold 0.85, artifacts unmodified since (progress.md § Plan-phase final state); skip rationale recorded here per the single authoritative skip contract.

**Input parameters.**

| signal | value |
|---|---|
| tier | L |
| scope | ~15–25 files: `internal/kanban` (board model, new), `internal/core/git` (extraction), `internal/hook` (re-point), `internal/template/templates` (mirror), plus tests |
| domain count | 2 (Go board/store model; template mirror) — single-domain core, not cross-domain fan-out |
| file language mix | Go + markdown templates |
| concurrency benefit | LOW — coding-heavy; each milestone's files feed the next (M1 store → M2 model on top of it → M3 mirror of what landed) |

**Mode evaluation.**

| mode | selected | rationale |
|---|---|---|
| trivial | no | Tier L, 25 REQ / 25 AC |
| background | no | write-capable implementation work |
| agent-team | no | RETIRED (Mode 3 tombstone) |
| parallel | no | coding-heavy — Anthropic coding-task parallelism caveat; M1→M2→M3 are strictly ordered by decision-reversibility |
| sub-agent | **yes** | single sequential `manager-develop` (cycle_type=tdd), one milestone at a time, per run.md Phase Owners |
| workflow | no | not high-volume mechanical; inter-file dependencies throughout |

**Decision:** `sub-agent`

**Justification.** The milestone order plan.md §F defines is a dependency chain — M1's state store and its resolution are the least reversible decisions and M2's board model is unusable without them, and M3 mirrors only what has landed. Nothing inside a milestone is genuinely parallelizable at coding granularity (the store, the lock, and the recovery all mutate the same package and the same files), so a sequential Mode 5 spawn per milestone with the orchestrator verifying between spawns is both the Anthropic-recommended default for coding work and the shape that keeps each delegation's context small enough to carry the full Section A–E template.

**Delegation surface.** One `manager-develop` spawn per milestone (M1, M2, M3), each with the Tier L 5-section template (context / known issues / pre-flight / constraints / self-verification), each inside the isolated worktree `~/.moai/worktrees/kanban-board` on branch `feat/SPEC-KANBAN-BOARD-001` (worktree created from `origin/main` = `a301ef6f8`; the shared primary checkout is dirty with another session's work and is not touched). Orchestrator work concurrent with a write-capable spawn stays read-only.
