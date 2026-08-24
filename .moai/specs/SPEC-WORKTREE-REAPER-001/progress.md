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

Card: t209. Branch `WT-worktree-reaper`, worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`. Base `origin/main` =
`cd0cee1b8`; plan-phase HEAD at run entry `ed002d089`.

Milestone order **M2 → M1 → M3**, per the operator-approved Implementation
Kickoff. The rationale is `plan.md` §F's ordering caveat: M1 does not create the
data-loss path, it **amplifies** it (occasionally one tree at a time becomes
roughly 98 in a single sweep), and the unlocked anchor is the residual — so the
guard stands before the amplifier.

### Commits

| Milestone | SHA | Subject |
|---|---|---|
| M2 | `3f8067fb2` | lock-aware anchor guard across all three sweeps |
| M1 | `81d8ae9a9` | three-valued merge detection |
| M3 | `aa14918d7` | `clean --stale` inventory + `origin/main` base |

### Files changed

- **New**: `internal/session/anchor_lock.go` (+ test),
  `internal/cli/session_worktree_prmerge_lock_test.go`,
  `internal/cli/session_worktree_prmerge_merge_test.go`,
  `internal/cli/worktree/clean_json_test.go`.
- **Modified**: `internal/session/anchor_pid_{unix,windows}.go`,
  `internal/cli/session_worktree_prmerge.go`, `internal/cli/worktree/clean.go`,
  four existing test files, `.claude/rules/moai/workflow/worktree-integration.md`
  (+ its template mirror), `.moai/config/sections/workflow.yaml`.

### Two decisions taken during run, both outside what the SPEC anticipated

**1. `auto_cleanup` set to `false` in this repository (M1 landing condition).**
`plan.md` §F requires M1 to land either with the toggle off or with an
enumerated expected-removal set. Measured at run time: this repository had
`workflow.worktree.auto_cleanup: true`
(`.moai/config/sections/workflow.yaml:124`), so the caveat was live rather than
hypothetical. The first option was taken — it is reversible, and it keeps the
first repaired sweep a deliberate act rather than a side effect of `moai session
list`. Re-enabling is an operator decision downstream of AC-WR-023(b).

**2. `isBaseBranch` added (M3).** Changing the `--stale` base default from
`main` to `origin/main` (REQ-WR-022) silently disabled the "checked out on the
base branch" guard, which compared a LOCAL branch name against a
remote-tracking ref literally. A second worktree sitting on `main` would then
have been judged by the merge predicate — which reports `main` as merged into
`origin/main` — and become removable. Caught by the pre-existing
`TestCleanStale_SkipsProtectedAndDetached`, which failed with
`protected/detached/base worktrees were removed: [/wt/on-base]`. The comparison
now also matches the base's trailing segment and errs toward keeping. **This
hazard is not named anywhere in `spec.md`, `design.md`, or `acceptance.md`.**

### One test expectation legitimately moved

`TestCleanStale_KeepsUnmergedWorktree` asserted the keep reason contained
`commits not in main`. With REQ-WR-022's base default the sweep now compares
against — and names — `origin/main`. The assertion was updated to
`commits not in origin/main`: the expectation moved because the SPEC mandated
the behaviour change, not to accommodate the implementation.

### Acceptance battery

All 32 test-shaped criteria run in the exact `acceptance.md` §0 form
(`go test <pkg> -run '^<Test>$' -v -count=1 2>&1 | grep -c '^--- PASS: <Test>'`).
Script: `.moai/state/verify/t209/ac-battery.sh`; verbatim output:
`.moai/state/verify/t209/ac-results.txt`.

**Every criterion returned `1`.** Every criterion whose recorded
pre-implementation baseline was `0` moved to `1`; every criterion recorded at
`1` (the existing M8 regression anchors) stayed at `1` across the seam-signature
change.

### AC-WR-022 — build / vet / cross-compile gate

```
go build ./...                                              → exit 0
go vet ./internal/cli/... ./internal/session/...             → exit 0
GOOS=windows go vet ./internal/cli/... ./internal/session/... → exit 0
MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/...   → ok (62.352s)
```

### Package suites — and an honest account of what the `internal/cli` run shows

```
go test ./internal/session/...      → ok   18.151s
go test ./internal/cli/worktree/... → ok   18.228s
go test ./internal/cli/             → see below
```

Every `internal/cli` sibling package (`harness`, `update/*`, `wizard`, `pr`,
`printer`, `specid`, `taskledger`, `uikit`, `agentlint`, `preference`) reported
`ok`. `internal/cli` itself is the exception and needs stating precisely rather
than summarised:

| Run | Scope | Result |
|---|---|---|
| at `3f8067fb2` (M2) | `internal/cli` alone, `-timeout 600s` | **ok, 383.405s** |
| at `aa14918d7` (M3) | all of `./internal/cli/... ./internal/session/...`, `-timeout 600s` | FAIL 601.548s — `panic: test timed out`, **0** `--- FAIL` lines |
| at `aa14918d7` (M3) | `internal/cli` alone, `-timeout 900s` | FAIL 909.905s — timeout panic, **1** `--- FAIL` |

The single `--- FAIL` is `TestHookWorktreeCreate_EchoesCreatedPath`:

```
hook_worktree_create_test.go:75: RunE returned error: dispatch hook: … 
  add worktree with new branch "worktree-e2e-probe": git worktree: 
  Preparing worktree (new branch 'worktree-e2e-probe') (git killed: context deadline exceeded)
```

Re-run in isolation at the same HEAD:

```
go test ./internal/cli/ -run '^TestHookWorktreeCreate_EchoesCreatedPath$' -v -count=1
→ --- PASS: TestHookWorktreeCreate_EchoesCreatedPath (4.16s)
```

**4.16s passing alone against 19.53s deadline-exceeded under load.** The test
shells out to a real `git worktree add` under a context deadline, in a
repository carrying 148 registered worktrees, on a loaded machine. It exercises
the WorktreeCreate hook — a path this SPEC does not touch (the diff reaches
`session_worktree_prmerge.go`, `worktree/clean.go`, and `internal/session`
only). Classified as **load-induced, not a regression**, on that evidence.

The 383s → >900s spread across two runs of the same package on the same tree is
the same cause: contention, not code. The package's wall-clock is measuring the
machine.

[HARD] A background-wrapper `exit=0` was reported for the 909s run while the log
said `FAIL`. The verdict here is read from the log, not from the wrapper's exit
code.

`GOOS=windows go vet` proves cross-platform **compilation only**. It is not
evidence of Windows behaviour, which matters here because the Windows liveness
probe cannot assert death and returns `(true, true)` unconditionally
(`design.md` §B.5) — deliberately, since reporting otherwise would widen removal
on the one platform with no real probe.

### AC-WR-023(a) — inert-sweep check, executed

```
go build -o ./bin/moai ./cmd/moai                                → exit 0
./bin/moai session list --json > .moai/state/verify/t209/sweep-a.log   → exit 0
grep -c 'PR-merge cleanup' .moai/state/verify/t209/sweep-a.log   → 0
git worktree list --porcelain | grep -c '^worktree '             → 148
git worktree list --porcelain | grep -c 't207'                   → 2
```

The repaired sweep removed nothing and emitted no notice; the registered
worktree population is intact and t207 is still on it. The tree-local binary was
used deliberately — the installed `~/go/bin/moai` predates the change.

**AC-WR-023(b) was NOT executed.** It performs the mass removal `spec.md` §G
declares out of scope and is gated on explicit operator authorisation plus a
pre-enumerated `expected-removals.txt`. Neither exists.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-08-24
milestones_complete: [M2, M1, M3]
commits: [3f8067fb2, 81d8ae9a9, aa14918d7]
criteria_passing: 32/32 test-shaped (AC-WR-001..021, 024, 024b, 025b, 026)
criteria_measurement_only: [AC-WR-025]
criteria_gated: [AC-WR-023b]
evidence_dir: .moai/state/verify/t209/
```

### Gaps — what was explicitly NOT observed

1. **Windows runtime behaviour.** Only compilation was verified. The two-valued
   probe's Windows arm is asserted by construction, not by execution.
2. **The real `clean --stale --json` inventory was not run against this
   repository.** `clean` calls `WorktreeProvider.Prune()` before listing, which
   mutates the shared repository's worktree administrative state; with 148
   registered trees and no operator authorisation, that was left undone. The
   behaviour is covered by unit tests only.
3. **`t207`'s directory was not stat-ed.** The worktree-isolation guard refuses
   `cd` and `git -C` into a sibling tree, so its presence is evidenced from
   `git worktree list` (shared across the repository) rather than from a
   filesystem check. Same positional limit that made AC-WR-025 a lead-run
   measurement.
4. **The agent-memory ∩ M1-unblocked intersection remains unmeasured**, exactly
   as `design.md` §A.7.2 Q2 records. P2 is the answer whether it is 0 or 5, so
   the gap does not change the implementation — but it is still a gap.
5. **The full test suite was not run locally** (repo rule). `go build ./...`,
   `go vet`, and the affected packages were run; CI owns the full-suite verdict.
6. **`internal/cli` has no clean green at the FINAL HEAD.** It is green at
   `3f8067fb2` (M2) and every one of its 32 acceptance criteria is green
   individually at `aa14918d7`, but the whole-package run at the final HEAD hit
   the wall-clock ceiling twice under load. The one `--- FAIL` was reproduced as
   passing in isolation and lies outside this diff, so the evidence points to
   contention — but "the package passes at `aa14918d7`" is an inference from
   parts, not an observation, and is recorded as a gap rather than claimed.
   CI's clean-environment run against the PR head is the stronger evidence and
   is the intended resolution.

### Residual risk — what could still be wrong despite the above

- **P2's allowlist is enumerated from one measurement.** A regenerable path that
  did not appear in those 156 trees preserves a tree it need not. That is the
  fail-closed direction and is intended, but it means the sweep will be more
  conservative than necessary until the list is extended deliberately.
- **The unlocked anchor (REQ-WR-020) is untouched by M2.** A `WT-*` tree
  materialised by this package's own code carries no lock and is guarded by the
  registry alone — the source measured at 1-of-5. Bounded by `auto_cleanup`
  being off, which is now also true in this repository.
- **The check→act race (EC-13) is narrowed, not closed**, and has no protection
  at all for ignored content: there is no second observation before removal.
- **Branch-naming convention change, notified mid-run by the lead.** Card
  worktree branches now carry a bare slug rather than the `WT-` prefix, so
  `prMergeCleanup` — which matches the prefix and only the prefix — will not
  see them. The direction is safe (they are preserved, never swept), and M3's
  `clean --stale --json` is the intended coverage for that population precisely
  because it enumerates by checkout rather than by branch name. No code change
  was made for this; it is recorded so the next reader does not mistake the
  prefix filter for full coverage.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
