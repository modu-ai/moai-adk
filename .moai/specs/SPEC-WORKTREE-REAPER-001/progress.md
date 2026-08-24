# progress.md — SPEC-WORKTREE-REAPER-001

## §E.1 Plan-phase Audit-Ready Signal

**v0.3.1 (post-audit-iteration-3 gate closure), 2026-08-24**

- Tier: L. Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
  `research.md`, `progress.md` — the full Tier L set (`research.md` added at
  v0.2.0; its absence was audit finding D16).
- Requirements: **24** (REQ-WR-001 … REQ-WR-024), GEARS notation, no gaps or
  duplicates. IDs 001-017 keep their v0.1.0 numbering; 018-024 are amendment
  additions placed in their owning milestone.
- Acceptance criteria: **27**. Every criterion carries a `Covers:` line and a
  recorded pre-implementation observation; 24/24 requirements covered.
- Audit trajectory: iter-1 FAIL 0.55 → iter-2 FAIL 0.84 → **iter-3 PASS 0.875**
  (threshold 0.85), zero regressions across all three, D1 re-verified at
  iteration 3 across 29 distinct test names with 0 discrepancies. Reports:
  `.moai/reports/t209/plan-audit{,-iter2,-iter3}.md`. v0.3.1 closes the two
  remaining blocking findings (F1 decision-rule decisiveness, F2 the fail-open
  criterion) plus residue.
- Audit iteration 1 verdict: FAIL 0.55 (Testability 0.30, Traceability 0.55);
  all seven must-pass criteria passed. Report:
  `.moai/reports/t209/plan-audit.md`.

**Criterion falsifiability — the finding that drove the amendment.** The v0.1.0
form `go test -run <NewTest>` expecting "exit 0" passes on the pre-implementation
tree, because Go exits 0 with `[no tests to run]` when `-run` matches nothing.
Verified in both directions on this tree:

```
go test ./internal/cli/ -run '^TestParseWorktreeList_BranchExtraction$' -v -count=1 \
  | grep -c '^--- PASS: TestParseWorktreeList_BranchExtraction'   → 1   (existing test)
go test ./internal/cli/ -run '^TestDoesNotExistAtAll$' -v -count=1 \
  | grep -c '^--- PASS: TestDoesNotExistAtAll'                    → 0   (missing test)
```

All criteria now use the second-form-falsifiable command. `acceptance.md` §0
carries the baseline and the mechanical proof that no new test name exists yet.

**Design decisions resolved at plan-phase** — all except the §A.7
ignored-content fork, which is deliberately open and gated on a measurement (see
the OPEN DECISION block below). No `[NEEDS CLARIFICATION]` markers are used: the
fork is settled by measurement rather than by preference, and every branch of
its decision rule now terminates in a measured answer (`design.md` §A.7, Q1/Q2):

- D1 — merge seam becomes `(string, bool)` (`design.md` §A.1).
- D2 — the zero-unique-commit removal class is **accepted**, bounded by the
  existing dirty guard, not excluded by a redundant predicate (§A.4).
- D3 — the anchor decision lives in `internal/session` and reaches **all three**
  blind consumers: `prMergeCleanup`, `cleanStaleWorktrees`
  (`internal/cli/worktree/clean.go:163`), and the `--merged-only` path
  (`clean.go:95`, the one with no dirty guard of its own) — §B.9.
- D4 — the M3 deliverable is an **extension of `moai worktree clean --stale`**
  (option O3-d), reversing v0.1.0's parallel-inventory decision (§C.4).
- D5 — the liveness probe becomes `(alive, determined)` with a per-platform
  mapping (§B.5).
- D6 — a confirmed-dead lock is **inert**; the sweep never unlocks (§B.6).

**Post-amendment measurements folded in** (`.moai/reports/t209/ec9-measurement.md`):

- **Q2 — `git branch --merged` ≡ zero unique commits, confirmed both
  directions.** Audit finding D2's *finding* stands (v0.1.0 never analysed the
  class); its *prescribed remedy* does not (there is no second predicate in
  `staleKeepReason` to copy). REQ-WR-018's accept-the-class decision is
  unchanged, and the disagreement is recorded as resolved by measurement in
  `design.md` §A.5.
- **Q1 — WITHDRAWN at v0.3.0. The v0.2.0 result does not reproduce.** The
  corrected record (`ec9-measurement.md` **v2**) shows non-forced `git worktree
  remove` **succeeds** on an ignored-only tree — exit 0, content destroyed. The
  v1 refusal was an untracked-file refusal caused by MoAI's statusline writing
  into a fixture that lived inside a live session's worktree. There is no third
  backstop layer; §A.4's accept-the-class decision rests on two, not three.
- **Downstream of the withdrawal:** REQ-WR-021 re-derived over the observed
  behaviour rather than an enumeration (its second limb was empty, and the
  enumeration was not exhaustive — a populated submodule refuses with a clean
  porcelain); REQ-WR-023 widened from "distinguish these two causes" to "every
  preserve notice names its cause"; REQ-WR-024 added for the ignored-content
  policy; AC-WR-024 rebuilt to assert cause-naming rather than string
  inequality; EC-8/9/11 corrected and EC-12/13 added.

**[HARD] OPEN DECISION — gated on a measurement, not deferred by omission.**
`design.md` §A.7 records the ignored-content policy as an unresolved fork with a
**fixed decision rule** and a **named measurement** (AC-WR-025), which is a
[HARD] precondition on M1 (`plan.md` §C.1). The hypothesis: `.moai/state/` is
gitignored (`git check-ignore -v .moai/state/config-cache.json` →
`.gitignore:284`) and MoAI writes into it in every tree a session occupies, so
"holds ignored content" may be true of essentially every worktree that has ever
hosted a session. If so, a preserve-on-ignored policy cancels M1's unblocking of
the ~98 merged trees. Candidate policies P1 (preserve on any ignored content),
P2 (classify regenerable vs irreplaceable by path), P3 (accept the loss
explicitly) and the rule selecting between them are all fixed at plan-phase;
only the measured input is outstanding. The rule is two sequential questions,
**both answered by AC-WR-025's own output** — Q1 decides P1-vs-rest on
prevalence, Q2 decides P2-vs-P3 on whether any observed ignored path lies
outside the regenerable set, with unclassifiable paths counting as
irreplaceable (fail-closed). No branch terminates in a judgement call.

**Third blind consumer named.** REQ-WR-019 now covers all three
`LiveAnchoredSessions` call sites; the `--merged-only` path
(`internal/cli/worktree/clean.go:95`) was unnamed through v0.2.0 and is the most
exposed — its own comment records it has no dirty guard, so the blind anchor
check is its sole protection.

**Recorded residual risks, not closed by this SPEC:**

- The **unlocked anchor** (REQ-WR-020): `materializeSessionWorktree` creates
  `WT-*` trees with no lock, inside the swept prefix set and invisible to the
  lock source. Bounded by `auto_cleanup` defaulting to `false`.
- A **force-pushed `origin/main`** could make an unmerged branch appear merged
  (EC-10); neither sweep fetches. The ordinary stale-ref direction is safe.
- **Submodule-bearing worktrees** (EC-12): a measured member of the refusal
  class outside the pre-detection set. No live instance (`.gitmodules` absent);
  falls through to the fail-open path, where git refuses and nothing is lost.
- **The check→act race** (EC-13): narrowed by the immediately-before-removal
  re-check, not closed. Benign for tracked/untracked content because git
  re-checks at removal time; **no equivalent protection for ignored content**.

**Measurement provenance.** Worktree-population figures drift because the sweep
under study is mutating the population — three measurements, three values, all
correct at their instant. Three different values are the expected result, not
three errors; `spec.md` §B.2 says so in one line above the table, and every
figure is timestamped rather than silently restated.

**M2's value, stated honestly.** M2 is a correctness and legibility gain, not a
rescue: every live anchor here is already locked and `git worktree remove`
refuses a locked tree, so no live session is removable today with or without M2.
The gain is that a protection currently arising from a side effect of git's lock
handling — untested, uncommented, and one `--force` away from vanishing — becomes
an intentional, tested invariant (`design.md` §D).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
