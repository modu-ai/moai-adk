---
id: SPEC-MEMORY-FOLD-RENAME-RACE-001
title: "Memory fold write path — cross-process store lock closing the last-check-to-rename lost-update window"
version: "0.1.1"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: GOOS행님
priority: P1
phase: "v3.2.0 target"
module: "internal/cli (memory fold)"
lifecycle: spec-anchored
tags: "memory-fold, concurrency, data-loss, cross-process-lock, rename-window, card-t1568"
tier: M
related_specs: [SPEC-MEMORY-FOLD-BUDGET-001]
---

## HISTORY

- 2026-10-09 — v0.1.1 — manager-spec — Base absorb re-pin (lane decision D-1, `.moai/reports/t1568/lane-decisions.md`): the card branch fast-forwarded `81786284e` → `2aab5f797` (155 commits; 3 added lines in `internal/cli/memory_fold.go` — 2 imports, 1 in `foldClosedCardMemory`). Every line citation re-measured by grep on the new base (+2 below the import block, +3 after the `foldClosedCardMemory` hunk; `os.Rename` now `:649`); `memory_fold_test.go` and `internal/sessionmsg/lock_*.go` verified unchanged by `git diff 81786284e 2aab5f797 --stat`. First baseline run on the old base classified a timeout artifact (`panic: test timed out after 4m0s`, zero `--- FAIL` lines) and is superseded by a 900s re-run on the new base (result in progress.md §E.1 and plan-evidence.md). No requirement, criterion, or scope text changed; REQ count 7, AC count 8.
- 2026-10-09 — v0.1.0 — manager-spec — Plan-phase artifacts authored for card t1568 (Tier M: spec.md + plan.md + acceptance.md; plus design.md — the lane-directed mechanism-rationale record — and progress.md + decision-index.md, the latter because `interview.decision_gate` is `on`). Defect: the fold apply path's last byte comparison and `os.Rename` are not serialized against concurrent writers, so a writer publishing inside that window is destroyed while the fold returns success (five independent observations recorded on the card). Plan mandate honored: RED reproduction first (plan.md M1), then a cross-process lock design. Tree baseline: HEAD `81786284e`. `status: draft`.

## Prior-Art Review

| Prior SPEC | Status | What it did | Relationship to this SPEC | Verdict |
|---|---|---|---|---|
| `SPEC-MEMORY-FOLD-BUDGET-001` (v0.4.0, Tier M) | completed | The fold command itself: plan/preview/apply, the REQ-MFB-004 write ordering (byte comparison immediately before each rename, effective-state re-check in the pre-rename position), the card-close wiring, the fold test seam. | **Constraint-bearing.** This SPEC hardens the write path that SPEC created. Its REQ-MFB-004 ordering — "bytes judged at the last observable moment" — is preserved exactly: the lock is added AROUND the existing check sequence (acquire before the first check, release after the last rename), never instead of it. The MFB fold-family tests remain the regression surface (card: "t1502 fold 계열·MEMORY.md 인덱스 회귀 포함"). | new (changes the behavior of a completed SPEC's code; that SPEC's body is not amended) |
| `SPEC-MEMORY-INDEX-FOLD-001` (v1.0.0, Tier S) | completed | One-time manual repair of dropped index lines. | No code overlap. Its discipline — MEMORY.md edits never lose a link — is the same value this SPEC enforces mechanically for concurrent edits. | new |

## 1. Background and measured premises

### 1.1 What the card asks

Card t1568 (P1, data-loss class, five independent observations): at `internal/cli/memory_fold.go:649` the fold write path runs its final byte-equality check and then `os.Rename(tmpName, target)` with nothing serializing the two steps against other writers. A concurrent writer that publishes its own update to `MEMORY.md` or to the archive index inside that window is destroyed by the rename while the fold returns success. The card mandates a RED reproduction first, then a lock / atomic-replace design, and names the t1502 fold family and the MEMORY.md index regression as the surfaces that must stay green.

The five observations recorded on the card (lane-12 r13 reproduction, lane-8 ledger #27, lane-2 r6, lane-9 t1507-gate, and the prior-generation mis-base group row ②) are motivation; no requirement below depends on re-measuring them. The defect itself is re-measured in this tree (P1, P3 below).

### 1.2 Measured premises (this tree, HEAD `2aab5f797`, 2026-10-09; first measured on the pre-absorb base `81786284e` and re-pinned by grep after the base absorb)

| Ref | Premise | Evidence (this tree) |
|---|---|---|
| P1 | The window exists exactly where the card says, and no existing seam can inject into it. | `internal/cli/memory_fold.go` — the final byte comparisons run at `:635-642` (`checkFoldUnchanged` on the target, then on the guard file), the "last observable moment" order probe fires at `:643-645`, the close-path abandonment check at `:646-648`, and `os.Rename` at `:649`. The existing test seam `mutateDuringWrite` fires at `:614-616` — BEFORE those checks — so anything it injects is caught by them (its own comment at `:75-78` says so). No hook fires after `:645`. |
| P2 | One write chokepoint: serializing `applyFold` serializes every fold write. | `applyFold` (`:414`) is the only non-test caller of `atomicWriteFoldFile`; both write surfaces go through it — the archive append at `:445` (guard nil) and the MEMORY.md rewrite at `:490` (guard set). Both entry points reach writes only through `applyFold`: the verb (`RunE`, `:172`) and the card-close step (`foldOnDoneStep`, `:1175`). |
| P3 | Both write surfaces carry the identical unprotected check→rename tail. | `atomicWriteFoldFile`'s tail (`:635-649`) is shared; the guard parameter only adds one more comparison, not protection. The defect class therefore covers the archive index append as well as the MEMORY.md rewrite. |
| P4 | The repository already owns the advisory-lock pattern this SPEC reuses. | `internal/sessionmsg/lock_unix.go` — `unix.Flock(fd, unix.LOCK_EX\|LOCK_NB)` on a `O_CREAT\|O_RDWR\|O_CLOEXEC` descriptor, with the `unacquiredFD = -1` sentinel lesson recorded; `internal/sessionmsg/lock_windows.go` — `LockFileEx` with `LOCKFILE_EXCLUSIVE_LOCK\|LOCKFILE_FAIL_IMMEDIATELY` as the Windows parity. `golang.org/x/sys v0.48.0` is already in `go.mod`. The pattern's origin (`internal/session` registry lock) is frozen — stated in sessionmsg's own header comment — so the reuse is copy-follow, not import. |
| P5 | Concurrency claims here are judged by repeated runs with the race detector, never one green run. | Repository test doctrine (concurrency tests judged by `-count` repetition with `-race`; recorded as a repo lesson after t1448). The acceptance criteria encode `-race -count=10` for every race-window criterion. |
| P6 | The fix must protect cooperating writers; it cannot protect writers that ignore the lock. | A POSIX rename is unconditional — there is no compare-and-swap rename. The only mechanism that closes the window is mutual exclusion that every fold honors. Writers outside that convention (a human editor) keep today's narrowing (the byte comparison), which is a best-effort limit, stated as such in §1.3 and Out of Scope. |
| P7 (card-reported, not re-measured) | Five independent loss observations across lane sessions. | Card text, dated 2026-10-09. Motivation only; no requirement or acceptance criterion asserts them. |

### 1.3 The loss invariant this SPEC establishes

The contract is REQ-MRR-004: while a fold apply runs on a store, a cooperating writer's published content survives whatever the fold does — the fold either aborts without writing (the writer landed before a check) or completes with the writer's content still present (the writer was serialized after the fold's renames by the lock). What the invariant deliberately does NOT claim: protection for writers that never acquire the lock. That limit is honest and is restated in Out of Scope; today the cooperating set is "every fold apply", and the lock convention is published so other memory writers can adopt it later.

## 2. Requirements (GEARS)

- REQ-MRR-001 — **The fold apply path shall hold an exclusive per-store cross-process lock from before its first pre-write check until its final rename has completed, on every fold application of a store — the `moai memory fold` verb and the card-close fold step alike.** (Ubiquitous. The span is the whole `applyFold` — premise P2: one chokepoint.)
- REQ-MRR-002 — **The store lock shall be one lock per store directory with a deterministic name derived from that store, guarding both write surfaces of a fold: the archive index append and the MEMORY.md rewrite.** (Ubiquitous.)
- REQ-MRR-003 — **When another holder keeps the fold from acquiring the store lock within a bounded wait, the requesting fold shall refuse cleanly: it shall write nothing, exit non-zero, and name the contended store in its error — it shall never block without a bound.** (Event-driven.)
- REQ-MRR-004 — **When a cooperating writer publishes content to MEMORY.md or to the archive index at any point during a fold apply on that store, the fold shall preserve the published content in the final store state — by serializing against it under the lock, or by aborting without a write — and shall never return success while the published content has been destroyed.** (Event-driven. This is the data-loss invariant; its RED proof is acceptance.md AC-MRR-001/002.)
- REQ-MRR-005 — **The presence of the store lock file shall change no observable fold or doctor output: no new doctor finding, no entry in the fold preview or the unlinked-archive listing, and no byte change to any store file.** (Ubiquitous — the lock file must be invisible to the store's own tooling.)
- REQ-MRR-006 — **While the card-close fold step waits on or holds the store lock, the step shall honor its existing abandonment bound exactly as before: a step whose bound expires shall begin no write and report at most one stderr line.** (State-driven — the MFB wiring's bounded-step semantics are unchanged by the lock.)
- REQ-MRR-007 — **The lock mechanism shall work on POSIX and on Windows with the same acquire, refuse-on-contention, and release semantics, and shall add no new module dependency.** (Ubiquitous — CI builds both GOOS; `go.mod` stays unchanged. The mechanism itself is design.md's subject, not a requirement here.)

## 3. Non-functional constraints

- C-1 — No third-party dependency: the mechanism reuses the repository's own advisory-lock pattern (P4) and the already-present `golang.org/x/sys`.
- C-2 — Uncontended behavior is byte-identical: fold preview text, `--json` plan, applied file bytes, and every existing fold test's expectations are unchanged.
- C-3 — The REQ-MFB-004 check ordering is frozen: the lock wraps the sequence; it does not reorder, remove, or weaken any existing check (the `orderProbe` regression test stays green untouched).
- C-4 — TRUST 5 gates run as usual; the affected test family (the fold family, premise P2's callers) is the local verification scope, with CI as the full-suite judge.

## 4. Out of Scope

### Out of Scope — non-cooperating writers
- A writer that never acquires the lock (a human editor, a foreign tool) is NOT protected by this SPEC; it keeps today's byte-comparison narrowing. Full protection there is impossible without the writer's cooperation (premise P6) and is not claimed by any requirement.

### Out of Scope — other memory-writing verbs adopting the lock
- `moai memory drain`, `moai memory diet`, and `moai memory archive` do not acquire the store lock in this SPEC (decision-index Q5, default applied: preserve current behavior). The lock convention is published for a follow-up adoption; nothing here changes those verbs.

### Out of Scope — the same-store two-fold scenario of card t1595
- Card t1595 item 8 (concurrent fold on one store; archive-before-snapshot overwrite) remains owned by t1595. This SPEC's whole-apply lock serializes fold-vs-fold on one store as a mechanical side effect of REQ-MRR-001's span — recorded as `related_specs` context, not as duplicated scope, and t1595's acceptance stays t1595's.

### Out of Scope — index content and budget behavior
- No line classification, dedupe, entry-link guard, or byte-budget behavior changes. The fold's outputs are frozen by C-2; this SPEC touches only who may write when.
