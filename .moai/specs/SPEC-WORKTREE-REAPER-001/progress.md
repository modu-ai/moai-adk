# progress.md — SPEC-WORKTREE-REAPER-001

## §E.1 Plan-phase Audit-Ready Signal

**v0.4.0 (§A.7 fork closed by measurement), 2026-08-24**

- Tier: L. Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
  `research.md`, `progress.md` — the full Tier L set (`research.md` added at
  v0.2.0; its absence was audit finding D16).
- Requirements: **25** (REQ-WR-001 … REQ-WR-025), GEARS notation, no gaps or
  duplicates. IDs 001-017 keep their v0.1.0 numbering; 018-025 are amendment
  additions placed in their owning milestone.
- Acceptance criteria: **28**. Every criterion carries a `Covers:` line and a
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

**Design decisions resolved at plan-phase — all of them.** The §A.7
ignored-content fork was the last open one and is now **closed** by the lead's
measurement (see the block below). No `[NEEDS CLARIFICATION]` markers were ever
used: the fork was settled by measurement rather than by preference, and every
branch of its decision rule terminates in a measured answer or in the
fail-closed sub-rule (`design.md` §A.7, Q1/Q2):

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

**[HARD] THE §A.7 FORK IS CLOSED — resolved to policy P2.** The deciding
measurement was run by the lead from the primary checkout, the one position
neither this session nor the auditor could occupy (the worktree guard refuses
`cd` and `git -C` into sibling trees).

```
worktrees measured                 156        failures (rc≠0)   0
carrying ignored content           156        ← 100%
   …with entries outside .moai/    153
```

**The predicate is measured-and-fatal, not degraded.** At 156/156, "holds
ignored content" has zero discriminating power: it would not merely re-block
M1's ~99 merged trees, it would make every tree in the repository permanently
immortal. Q1 eliminates P1. Q2's input — the intersection of the 5
`.claude/agent-memory/` trees with the ~99 M1-unblocked set — is **unmeasured**,
and the v0.3.1 fail-closed sub-rule resolves it to **P2** identically whether
that intersection is 0 or 5. The rule reached an answer through its own fallback
on an input nobody had when it was written, which is what it was built for.

Do NOT read "153 with entries outside `.moai/`" as 153 irreplaceable trees:
`.claude/settings.local.json` (148), `bin` (64), `.ruff_cache` (19) and
`docs-site/public` (12) are all outside `.moai/` and all regenerable.

**`.claude/agent-memory/` classified irreplaceable — on the REGENERABILITY axis,
not the value axis.** Nothing regenerates agent memory: no build, no runtime, no
session replay. That is a property of the category, true regardless of what any
particular file holds — which matters because grounding the classification on
content value would make it contingent on inspecting all 5 trees and every
future tree. Content was inspected and the value is real and branch-independent
(this card's own audit wrote a cross-cutting lesson on acceptance-criterion
falsifiability), but the classification does not rest on it.

**Cost, as a bound:** P2 preserves at most 5 of 156 trees, ~1GB of 30G. What it
costs M1 specifically is **between 0 and 5 trees** — the unmeasured intersection
again. "Blocks nothing M1 needs" would overstate the measurement.

**P2 is a stopgap** (REQ-WR-025). It preserves trees because worktree-local
agent memory has nowhere else to live; drain-then-dispose is the correct fix and
is a separate card (`spec.md` §G).

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
