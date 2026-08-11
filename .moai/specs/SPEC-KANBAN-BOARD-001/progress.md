---
id: SPEC-KANBAN-BOARD-001
title: "Progress — six-column kanban board model with a single-origin board state store"
version: "0.4.0"
status: draft
created: 2026-08-10
updated: 2026-08-11
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

- Requirements: 24 (`REQ-KB-001` … `REQ-KB-024`) — under the Tier L ceiling of 25, with one to spare.
- Acceptance criteria: 25 (`AC-KB-001` … `AC-KB-025`) — **at** the Tier L ceiling of 25, with none to spare.
- Milestones: 4 (M0 … M3), ordered by decision-reversibility. M1 grew at v0.2.0 to carry the sole-writer, atomicity, and board-wide-lock requirements ahead of the M2 admission they make sound, and again at v0.3.0 to carry the absent/unreadable split, the recovery bound, and the stale-lock exit. `REQ-KB-020` lands in M2, ahead of the compatibility table it feeds, and `REQ-KB-024` lands beside it at v0.4.0.
- Dependencies: `SPEC-KANBAN-RENAME-001` alone (untracked in this worktree, not on `origin/main`) — M0 gates on it, and AC-KB-001 decides that gate against `origin/main` rather than against the working tree, as of v0.3.0 does `plan.md` §C command 1.
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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
