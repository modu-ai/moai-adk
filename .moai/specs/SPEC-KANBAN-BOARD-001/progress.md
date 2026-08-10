---
id: SPEC-KANBAN-BOARD-001
title: "Progress — six-column kanban board model with a single-origin board state store"
version: "0.3.0"
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

- Requirements: 23 (`REQ-KB-001` … `REQ-KB-023`) — under the Tier L ceiling of 25, with two to spare.
- Acceptance criteria: 24 (`AC-KB-001` … `AC-KB-024`) — under the Tier L ceiling of 25, with one to spare.
- Milestones: 4 (M0 … M3), ordered by decision-reversibility. M1 grew at v0.2.0 to carry the sole-writer, atomicity, and board-wide-lock requirements ahead of the M2 admission they make sound, and again at v0.3.0 to carry the absent/unreadable split, the recovery bound, and the stale-lock exit. `REQ-KB-020` lands in M2, ahead of the compatibility table it feeds.
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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
