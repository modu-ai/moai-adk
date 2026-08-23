# progress.md — SPEC-WORKTREE-REAPER-001

## §E.1 Plan-phase Audit-Ready Signal

**v0.2.0 (post-audit amendment), 2026-08-24**

- Tier: L. Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `design.md`,
  `research.md`, `progress.md` — the full Tier L set (`research.md` added at
  v0.2.0; its absence was audit finding D16).
- Requirements: **23** (REQ-WR-001 … REQ-WR-023), GEARS notation, no gaps or
  duplicates. IDs 001-017 keep their v0.1.0 numbering; 018-023 are amendment
  additions placed in their owning milestone.
- Acceptance criteria: **24**. Every criterion carries a `Covers:` line and a
  recorded pre-implementation observation.
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

**Design decisions resolved at plan-phase** (no `[NEEDS CLARIFICATION]` markers
remain):

- D1 — merge seam becomes `(string, bool)` (`design.md` §A.1).
- D2 — the zero-unique-commit removal class is **accepted**, bounded by the
  existing dirty guard, not excluded by a redundant predicate (§A.4).
- D3 — the anchor decision lives in `internal/session` and reaches **both**
  consumers, `prMergeCleanup` and `cleanStaleWorktrees` (§B.9).
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
- **Q1 — EC-9 closes in the safe direction and is now asserted.** `git status
  --porcelain` returns 0 for an ignored-only tree, but non-forced `git worktree
  remove` exits 128 and the file survives: git's own check is stricter than the
  dirty guard (`design.md` §A.6).
- **Downstream of Q1:** the ignored-only tree and the locked tree produce one
  indistinguishable permanently-recurring symptom. REQ-WR-021 was generalised
  from locks to the whole refusal class with pre-detection, REQ-WR-023 added for
  notice attributability, AC-WR-024 and EC-11 added, and two alternatives
  recorded as rejected (attempt-and-fail with distinct text; widening the shared
  `worktreeIsDirty`) — `design.md` §B.6a.

**Recorded residual risks, not closed by this SPEC:**

- The **unlocked anchor** (REQ-WR-020): `materializeSessionWorktree` creates
  `WT-*` trees with no lock, inside the swept prefix set and invisible to the
  lock source. Bounded by `auto_cleanup` defaulting to `false`.
- A **force-pushed `origin/main`** could make an unmerged branch appear merged
  (EC-10); neither sweep fetches. The ordinary stale-ref direction is safe.

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
