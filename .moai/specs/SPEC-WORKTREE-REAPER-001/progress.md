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
| at `aa14918d7` (M3) | `internal/cli` alone, uncontended, `-timeout 1800s` | **ok, 768.952s** — 0 `--- FAIL`, 0 timeout panics, `exit=0` |

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

A fourth run, once the machine was uncontended, closed the question directly:
**`ok  github.com/modu-ai/moai-adk/internal/cli  768.952s`**, exit 0, zero
`--- FAIL` lines and zero timeout panics — the same tree, the same HEAD, the
same command, differing only in load. `TestHookWorktreeCreate_EchoesCreatedPath`
passed in that run.

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
6. ~~**`internal/cli` has no clean green at the FINAL HEAD.**~~ **CLOSED by
   measurement.** Recorded here rather than deleted, because the closure is the
   point: the claim was withheld while it rested on an inference from parts (32
   criteria green individually, plus one `--- FAIL` reproduced as passing in
   isolation), and released only once the whole package was observed green at
   `aa14918d7` in an uncontended run — `ok … 768.952s`, exit 0, zero `--- FAIL`,
   zero timeout panics. CI's clean-environment run against the PR head remains
   the stronger evidence and is still the binding verdict.

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

```yaml
sync_status: audit-ready
sync_complete_at: 2026-08-24
sync_commit_sha: 3b64581101fb345f222545ac29b87e927b5b07f9
changelog_entry_position: "CHANGELOG.md [Unreleased] — Added 1, Changed 2, Fixed 4"
docs_locales_synced: [en, ko, ja, zh]   # 8 files: cli-reference/worktree.md + worktree/guide.md per locale
frontmatter_status_transitions:
  spec.md: draft -> completed
  plan.md: n/a (no frontmatter block)
  acceptance.md: n/a (no frontmatter block)
  progress.md: n/a (no frontmatter block)
b12_self_test_a: "grep -c 'WORKTREE-REAPER' CHANGELOG.md -> 0 (pre-emission, no duplicate)"
b12_self_test_b: "grep -oE 'AC-WR-[0-9]+[a-z]?' acceptance.md | sort -u | wc -l -> 28 (matches the 28 declared)"
b12_self_test_c: "ls of every path named in the CHANGELOG entry -> all present"
```

### Claim

The sync phase closed SPEC-WORKTREE-REAPER-001: the `[Unreleased]` CHANGELOG
section carries the three milestones' user-visible surface, the four-locale
`moai worktree` CLI reference was corrected where M3 made it false, and
`spec.md` moved to `status: completed`. The tree builds, vets clean, and the two
packages this SPEC changed pass.

### Evidence

Commands run in this tree, at `c738b6e53`, with their verbatim output. Exit
codes are read from the command itself, never from the tail of a pipeline.

```
$ go build ./... > .moai/state/verify/t209/sync-build.log 2>&1; echo "build_exit=$?"
build_exit=0
$ wc -c .moai/state/verify/t209/sync-build.log
       0 .moai/state/verify/t209/sync-build.log

$ go vet ./... > .moai/state/verify/t209/sync-vet.log 2>&1; echo "vet_exit=$?"
vet_exit=0
$ wc -c .moai/state/verify/t209/sync-vet.log
       0 .moai/state/verify/t209/sync-vet.log

$ go test -count=1 ./internal/session/... ./internal/cli/worktree/... \
    > .moai/state/verify/t209/sync-tests.log 2>&1; echo "exit=$?"
exit=0
$ cat .moai/state/verify/t209/sync-tests.log
ok  	github.com/modu-ai/moai-adk/internal/session	31.405s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	21.518s
```

B12 CHANGELOG-emission self-tests, run before the append:

```
$ grep -c 'WORKTREE-REAPER' CHANGELOG.md
0
$ grep -oE 'AC-WR-[0-9]+[a-z]?' .moai/specs/SPEC-WORKTREE-REAPER-001/acceptance.md | sort -u | wc -l
      28
$ ls internal/session/anchor_lock.go internal/cli/worktree/clean.go internal/cli/session_worktree_prmerge.go
internal/cli/session_worktree_prmerge.go
internal/cli/worktree/clean.go
internal/session/anchor_lock.go
```

The AC count is 28, not the 26 a naive `AC-([A-Z0-9]+-)*[0-9]+` sweep returns:
that pattern truncates the two sub-ID criteria (`AC-WR-024b`, `AC-WR-025b`) onto
their parents. 28 matches `acceptance.md`'s own opening line and `§E.1`.

Cause tokens claimed in the CHANGELOG were read out of the source rather than
recalled:

```
$ grep -o 'cause=[a-z-]*' internal/cli/session_worktree_prmerge.go | sort -u
cause=anchored-by-
cause=dirty
cause=dirty-check-failed
cause=ignored-check-failed
cause=ignored-content
cause=refusal-class
cause=undetermined-merge
```

`cause=anchored-by-` is a format prefix (`cause=anchored-by-%s`) completed at
runtime by the anchor source, `lock` or `registry` — both spellings named in the
CHANGELOG exist only through that verb.

docs-site finding and the edit it forced:

```
$ grep -n -e '--base' -e '--json' docs-site/content/en/cli-reference/worktree.md
45:| `--base <branch>` | Base branch (default: `main`) |
87:| `--base <branch>` | Base branch used to judge `--merged-only` and `--stale` (default: `main`) |
```

Line 87 belonged to `worktree clean` and became false at M3; line 45 belongs to
`worktree sync`, whose base default is still the local `main`
(`internal/cli/worktree/sync.go:28`) and which was therefore left untouched.

A second page carried the same false default. Sweeping the `worktree/` guide,
FAQ, and examples pages for the changed flags found `worktree/guide.md`
documenting the clean `--base` default as `main` in all four locales as well
(en:270, ko:417, ja:263, zh:259 pre-edit); the FAQ and examples pages name the
flags only in command lines, which stay correct. Eight files were therefore
corrected, not four: in each, the clean `--base` default now reads
`origin/main`, a `--json` entry was added beside `--yes`, and the
`cli-reference` usage line carries `[--json]`.

```
$ git diff --stat -- docs-site
 docs-site/content/en/cli-reference/worktree.md | 5 +++--
 docs-site/content/en/worktree/guide.md         | 3 ++-
 docs-site/content/ja/cli-reference/worktree.md | 5 +++--
 docs-site/content/ja/worktree/guide.md         | 3 ++-
 docs-site/content/ko/cli-reference/worktree.md | 5 +++--
 docs-site/content/ko/worktree/guide.md         | 3 ++-
 docs-site/content/zh/cli-reference/worktree.md | 5 +++--
 docs-site/content/zh/worktree/guide.md         | 3 ++-
 8 files changed, 20 insertions(+), 12 deletions(-)
```

The post-edit sweep confirms no clean `--base` row still reads `main`: all eight
matches for the `--merged-only` + `--stale` base row now carry `origin/main`.

### Baseline-attribution

Every figure above was measured in this run, in the worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`, on branch
`WT-worktree-reaper`, at `git rev-parse --short HEAD` = `c738b6e53` — the run
phase's last commit, which is the tree the sync commit is built on. The build
and vet logs are zero bytes, which is what their exit-0 claim rests on; the test
log is quoted whole. Nothing here is carried over from another package, another
tree, or another point in time.

One figure IS carried over, and is labelled as such: `internal/cli`'s package
suite was **not** re-run in this phase. Its green is the prior attributed
measurement recorded in `§E.2` — `ok  github.com/modu-ai/moai-adk/internal/cli
768.952s`, exit 0, zero `--- FAIL`, zero timeout panics, observed at
`aa14918d7` in an uncontended run. The sync commit touches no Go source, so that
attribution still describes the code under test; it does not describe this tree
at this instant, and is cited as prior evidence rather than as a fresh
measurement.

### Gaps — what was explicitly NOT observed

1. ~~**`sync_commit_sha` is a placeholder.**~~ **CLOSED by backfill.** The sync
   commit `3b64581101fb345f222545ac29b87e927b5b07f9` could not name its own
   hash, so the field carried `pending-backfill-t209` until this follow-up
   commit wrote the observed value (the SHA-placeholder backfill exemption in
   `spec-frontmatter-schema.md`). Era classification now reads a real SHA.
2. **The full test suite was not run locally** — repo rule (`CLAUDE.local.md`
   §4/§6). `go build ./...`, `go vet ./...`, and the two packages this SPEC
   changed were run; CI owns the full-suite verdict against the PR head.
3. **`internal/cli` was not re-run in this phase** (see the attribution note
   above). The sync commit changes no Go file, so nothing in it can move that
   package — but the package was not observed at the sync HEAD.
4. **The docs-site build was not run.** The edits are table rows, list items,
   and one fenced usage line per page; no shortcode, front matter, or menu entry
   was touched. Hugo was not invoked to confirm it.
5. **The `--json` inventory was still not run against this repository.** `§E.3`
   gap 2 stands unchanged: `clean` prunes before listing, which mutates shared
   worktree administrative state, and no operator authorisation exists for that.
   The behaviour remains covered by unit tests only.
6. **Only `spec.md` carries a frontmatter block.** `plan.md`, `acceptance.md`,
   and `progress.md` open with an H1 and have no YAML block, so the close
   transition was applied to `spec.md` alone. No frontmatter was synthesised into
   the other three — inserting one would be a body modification, which sync
   phase does not own.
7. **`status` moved `draft → completed` in one step.** The run phase never
   performed `draft → in-progress`, so the intermediate states were never
   recorded in the artifact; the git history carries them
   (`3f8067fb2` / `81d8ae9a9` / `aa14918d7`) but the frontmatter does not.

### Residual risk — what could still be wrong despite the above

- **The CHANGELOG describes behaviour, and behaviour was verified by unit tests,
  not by a live sweep.** The `--json` inventory, the ignored-content guard, and
  the refusal-class pre-detection are each asserted by tests over injected
  seams; none was observed acting on this repository's 148 registered trees.
- **The docs correction was driven by a flag-token grep, not by reading the
  pages.** `--base`, `--merged-only`, `--stale`, and `--json` were swept across
  the four locales' `cli-reference/` and `worktree/` trees, and the second false
  page surfaced only because of that sweep — the first pass had stopped at the
  CLI reference. A prose sentence describing base-branch behaviour without
  naming a flag would still have been missed by the same method.
- **`auto_cleanup: false` is repo-local and deliberately absent from the
  CHANGELOG.** A reader of the release notes learns the sweep was repaired
  without learning that this repository has it switched off; that is correct for
  a distributed changelog and wrong for anyone reasoning about this tree from
  the changelog alone.
- **The `§E.2` carry-over is the one attribution a later reader could misread.**
  It is labelled prior evidence here, but a summary that quotes "internal/cli
  green" without its `aa14918d7` anchor would convert a correct citation into an
  unattributed claim.

## §E.5 Post-audit repair

Four blocking findings from the sync-phase audit (`.moai/reports/t209/sync-audit.md`,
FAIL 0.778, Functionality 0.72) were repaired test-first. F5–F12 are optional and
were NOT touched.

### Claim

- **F1 (Critical) closed.** The gh-no-answer merge fallback can now recognise a
  branch checked out in a linked worktree — the entire population the sweep
  evaluates. The parser strips git's `+` marker as well as `*`, and skips the
  `(HEAD detached at …)` / `(no branch)` lines. It is bound by a parser-level
  battery plus one decision-level criterion that routes the seam through the
  real parser, closing the test-seam blind spot that let the defect ship green.
- **F2 (High) closed.** `worktreeLockStates` returns `(map, error)`; both
  `clean` sweeps fail closed on an unreadable porcelain, matching
  `prMergeCleanup` / AC-WR-016. `--merged-only` removes nothing and emits
  `cause=lock-source-unreadable`; `--stale` keeps every tree with that cause;
  `--stale --json` reports `anchored: "undetermined"` rather than `"no"`.
- **F3 (High) closed.** The ignored-content decision — allowlist included —
  moved to `internal/session` and now covers `clean --stale`, so a merged,
  clean, unanchored tree holding only `.claude/agent-memory/` is preserved
  instead of destroyed. `--stale --json` gains a fourth predicate, `ignored`.
  The rule doc and its template mirror were updated to match.
- **F4 (Medium) closed.** `isBaseBranch` has a table test plus a sweep-level
  case; a mutation check proves the test binds the trailing-segment guard.

Four commits, each naming `t209`:

```
2045acc92 fix(worktree): share the ignored-content guard with clean --stale (t209)
90fc07845 fix(worktree): fail closed when the lock source is unreadable (t209)
dcdab7220 test(worktree): bind the trailing-segment base guard (t209)
69e415311 fix(cli): strip git's linked-worktree marker in the merge fallback (t209)
```

### Evidence

Every fix was driven RED → GREEN. Verbatim, in order.

**F1 — RED** (`go test -count=1 -run 'TestParseGitBranchMergedOutput|TestBranchMergedForCleanup_LinkedWorktreeMarker' ./internal/cli/`):

```
--- FAIL: TestParseGitBranchMergedOutput (0.00s)
    --- FAIL: TestParseGitBranchMergedOutput/linked-worktree_marker_is_stripped (0.00s)
        session_worktree_branchmerged_test.go:58: parseGitBranchMergedOutput("+ WT-agent-toml-dual\n") = []string{"+ WT-agent-toml-dual"}, want []string{"WT-agent-toml-dual"}
    --- FAIL: TestParseGitBranchMergedOutput/mixed_markers_in_one_listing (0.00s)
        session_worktree_branchmerged_test.go:58: parseGitBranchMergedOutput("* main\n+ WT-worktree-reaper\n  WT-docs-redesign\n") = []string{"main", "+ WT-worktree-reaper", "WT-docs-redesign"}, want []string{"main", "WT-worktree-reaper", "WT-docs-redesign"}
    --- FAIL: TestParseGitBranchMergedOutput/detached_entries_are_not_branch_names (0.00s)
        session_worktree_branchmerged_test.go:58: parseGitBranchMergedOutput("* (HEAD detached at 301841e0f)\n+ (no branch)\n  WT-real\n") = []string{"(HEAD detached at 301841e0f)", "+ (no branch)", "WT-real"}, want []string{"WT-real"}
--- FAIL: TestBranchMergedForCleanup_LinkedWorktreeMarker (0.00s)
    session_worktree_branchmerged_test.go:77: a branch checked out in a linked worktree must be recognised as merged; got 0
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.854s
```

**F1 — GREEN** (same command, after the parser fix):

```
ok  	github.com/modu-ai/moai-adk/internal/cli	1.541s
```

**F1 — real-data replay.** The audit's measurement reproduced in this tree, then
re-run against the corrected logic:

```
$ git branch --merged origin/main > /tmp/t209_merged.txt; wc -l < /tmp/t209_merged.txt
     149
$ grep -c '^+' /tmp/t209_merged.txt
119
old parser entries: 149
old parser still '+' prefixed: 119
new parser entries: 149
new parser still '+' prefixed: 0
sample new: WT-agent-toml-dual
```

**F2 — RED** (`go test -count=1 -run 'UnreadableLockSource|JSONReportsUndeterminedAnchor' ./internal/cli/worktree/`):

```
--- FAIL: TestCleanStale_UnreadableLockSourceRemovesNothing (0.00s)
    clean_lock_unreadable_test.go:57: an unreadable lock source must remove nothing; removed [/wt/lock-unreadable]
--- FAIL: TestCleanMergedOnly_UnreadableLockSourceRemovesNothing (0.00s)
    clean_lock_unreadable_test.go:75: an unreadable lock source must remove nothing; removed [/wt/merged-lock-unreadable]
--- FAIL: TestCleanStale_JSONReportsUndeterminedAnchor (0.00s)
    clean_lock_unreadable_test.go:104: an unreadable lock source must report anchored="undetermined", got "no"
    clean_lock_unreadable_test.go:107: keep_reason must name the cause, got ""
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.553s
```

The RED output is itself the harm statement: the fixture's tree was removed.

**F2 — GREEN** (same command, `-v`):

```
=== RUN   TestCleanStale_UnreadableLockSourceRemovesNothing
--- PASS: TestCleanStale_UnreadableLockSourceRemovesNothing (0.00s)
=== RUN   TestCleanMergedOnly_UnreadableLockSourceRemovesNothing
--- PASS: TestCleanMergedOnly_UnreadableLockSourceRemovesNothing (0.00s)
=== RUN   TestCleanStale_JSONReportsUndeterminedAnchor
--- PASS: TestCleanStale_JSONReportsUndeterminedAnchor (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.327s
```

**F3 — RED** (`go test -count=1 -run 'IrreplaceableIgnoredContent|RemovesWhenIgnoredContentIsRegenerable|UnreadableIgnoredStatusPreserves|JSONReportsIgnoredPredicate' ./internal/cli/worktree/`):

```
--- FAIL: TestCleanStale_KeepsIrreplaceableIgnoredContent (0.00s)
    clean_ignored_content_test.go:65: a tree holding irreplaceable ignored content was removed: [/wt/memory]
--- FAIL: TestCleanStale_UnreadableIgnoredStatusPreserves (0.00s)
    clean_ignored_content_test.go:117: an unreadable ignored status must preserve; removed [/wt/unreadable]
--- FAIL: TestCleanStale_JSONReportsIgnoredPredicate (0.00s)
    clean_ignored_content_test.go:142: a tree holding irreplaceable ignored content must report ignored="yes", got "not-checked"
    clean_ignored_content_test.go:145: the record must carry a keep_reason
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.659s
```

The fourth case in that battery — `RemovesWhenIgnoredContentIsRegenerable` —
passed on the pre-repair tree, and is deliberate: it is the anti-immortality
control, and it must still pass after the guard exists.

**F3 — GREEN** (same command):

```
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.566s
```

Shared predicate, bound in its new home
(`go test -count=1 -run 'TestIrreplaceableIgnoredEntries|TestIsRegenerableIgnoredPath' ./internal/session/`):

```
ok  	github.com/modu-ai/moai-adk/internal/session	0.721s
```

Rule-doc mirror parity after the four-predicate edit:

```
$ diff -q .claude/rules/moai/workflow/worktree-integration.md internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md && echo MIRROR_IDENTICAL
MIRROR_IDENTICAL
```

**F4 — the test passed on first run**, because the predicate was already correct
(the audit confirmed `isBaseBranch("feature/main","origin/main")` → `false`); the
finding was that nothing bound it. A mutation check supplies the RED evidence —
with the trailing-segment branch deleted from `isBaseBranch`:

```
--- FAIL: TestIsBaseBranch (0.00s)
    --- FAIL: TestIsBaseBranch/local_main_against_the_remote-tracking_base (0.00s)
        clean_base_branch_test.go:42: isBaseBranch("main", "origin/main") = false, want true
    --- FAIL: TestIsBaseBranch/a_different_remote's_main_still_matches_on_the_trailing_segment (0.00s)
        clean_base_branch_test.go:42: isBaseBranch("main", "upstream/main") = false, want true
--- FAIL: TestCleanStale_KeepsWorktreeOnBaseBranch (0.00s)
    clean_base_branch_test.go:64: a worktree on the base branch was removed: [/wt/second-main]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.509s
```

**F4 — GREEN** (source restored, same command):

```
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.324s
```

**Verification gate**, run at `2045acc92`:

```
$ go build ./... > .moai/state/verify/t209/f-repair-build.log 2>&1; echo "build_exit=$?"
build_exit=0
$ wc -c .moai/state/verify/t209/f-repair-build.log
       0 .moai/state/verify/t209/f-repair-build.log
$ go vet ./... > .moai/state/verify/t209/f-repair-vet.log 2>&1; echo "vet_exit=$?"
vet_exit=0
$ wc -c .moai/state/verify/t209/f-repair-vet.log
       0 .moai/state/verify/t209/f-repair-vet.log
$ GOOS=windows go vet ./... > .moai/state/verify/t209/f-repair-vet-windows.log 2>&1; echo "windows_vet_exit=$?"
windows_vet_exit=0
$ wc -c .moai/state/verify/t209/f-repair-vet-windows.log
       0 .moai/state/verify/t209/f-repair-vet-windows.log
$ go test -count=1 ./internal/session/... ./internal/cli/worktree/... > .moai/state/verify/t209/f-repair-pkgtests.log 2>&1; echo "pkg_exit=$?"
pkg_exit=0
$ cat .moai/state/verify/t209/f-repair-pkgtests.log
ok  	github.com/modu-ai/moai-adk/internal/session	9.846s
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	7.399s
$ go test -count=1 -run 'TestParseGitBranchMergedOutput|TestBranchMergedForCleanup_LinkedWorktreeMarker|TestPRMergeCleanup' ./internal/cli/ > .moai/state/verify/t209/f-repair-cli-targeted.log 2>&1; echo "cli_targeted_exit=$?"
cli_targeted_exit=0
$ cat .moai/state/verify/t209/f-repair-cli-targeted.log
ok  	github.com/modu-ai/moai-adk/internal/cli	2.147s
```

Template neutrality / leak guard over the mirrored rule doc:

```
$ go test -count=1 -run 'Leak|Neutral|Mirror|Parity' ./internal/template/
ok  	github.com/modu-ai/moai-adk/internal/template	7.730s
```

### Baseline-attribution

Every figure above was measured in this run, in the worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209`, on branch
`WT-worktree-reaper`. The RED runs were measured at the tree state immediately
before each fix (starting from `301841e0f`, the HEAD the audit judged); the
GREEN runs and the verification gate were measured at `2045acc92`, the last
repair commit. The build and both vet logs are zero bytes, which is what their
exit-0 claims rest on; the test logs are quoted whole. The `git branch --merged
origin/main` replay was run against this repository's live worktree population
at this instant, and reproduces the audit's own 149 / 119 figures exactly.
Nothing here is carried over from another package, another tree, or another
point in time.

### Gaps — what was explicitly NOT observed

1. **The `internal/cli` package suite was not run in full.** The repo rule
   forbids a local full suite, and the package alone runs ~13 minutes. Targeted
   runs of the tests this repair touches (`TestParseGitBranchMergedOutput`,
   `TestBranchMergedForCleanup_LinkedWorktreeMarker`, `TestPRMergeCleanup*`,
   plus a broader `-run 'Worktree|Anchor|Ignored|Merged'` sweep) were run and
   are green. CI owns the full-suite verdict against the PR head; the branch is
   unpushed, so that verdict is outstanding.
2. **The 28-criterion AC battery was not re-executed** — same reason, and the
   same gap the audit itself recorded. F1's fix changes an implementation the
   AC battery only reached through a swapped seam, so no criterion's expected
   value moved; that is a reading, not a measurement.
3. **`clean --stale` was never executed against a real repository.** All four
   repairs are bound by fixtures with stubbed git. Running the real command
   would prune shared worktree records, which the dispatch prohibits.
4. **`golangci-lint` was not run** — outside the authorised command set. Only
   `go vet` was observed.
5. **Windows behaviour is unobserved.** `GOOS=windows go vet ./...` exits 0,
   which proves compilation and nothing about runtime.
6. **F5–F12 were not addressed.** They are optional and explicitly out of
   scope, including F5 (`.moai/state` allowlisting covers this SPEC's own
   audit-evidence store) — which means the allowlist this repair SHARED still
   carries the classification the audit flagged, now in one place rather than
   two.
7. **One unrelated flake was observed and not root-caused.** In a contended
   run, `TestHookWorktreeCreate_EchoesCreatedPath` failed with `git worktree:
   (git killed: context deadline exceeded)` while the `internal/session` and
   `internal/cli/worktree` suites were still finishing. It passed in isolation
   (3.195s) and passed on re-run of the same targeted set (12.953s, against the
   65.752s contended run). It creates a real worktree in a repository holding
   150+ of them; nothing in this repair touches worktree creation.

### Residual risk — what could still be wrong despite the above

- **F1's end-to-end path is still unmeasured.** The parser is now correct
  against real stdout and the decision-level test binds the fallback, but no
  test drives `gitBranchMergedReal` against a live `git branch --merged` — the
  seam that made the original defect invisible is narrowed, not removed.
- **The shared ignored-content decision now has two callers and one allowlist.**
  That is the point of the lift, and it also means an over-broad entry (F5)
  now misclassifies for both sweeps at once rather than one.
- **`clean --stale`'s check→act race is unchanged.** Classification of every
  tree completes before any removal begins, so tree N's ignored-content and
  anchor state were both read before trees 1..N-1 were removed. The new guard
  inherits that window rather than closing it.
- **The rule doc now describes a guard whose blast radius is unmeasured here.**
  How many trees in this repository hold only regenerable ignored content — and
  are therefore still removable — was not counted; the sibling-worktree
  inspection that would count it is refused by the isolation guard.

## §E.6 Post-audit repair, iteration 2 (finding N1)

### Claim

`moai worktree clean --merged-only` — the third removal path, and the one the
shared ignored-content decision's own doc comment already claimed to cover —
now consults that decision. A merged, unanchored worktree holding irreplaceable
gitignored content is preserved on all three paths, not two.

### Evidence

RED, at `96a26b0f8` with the new tests and no guard:

```
--- FAIL: TestCleanMergedOnly_KeepsIrreplaceableIgnoredContent (0.00s)
    clean_ignored_content_test.go:168: a tree holding irreplaceable ignored content was removed: [/wt/merged-memory]
--- FAIL: TestCleanMergedOnly_UnreadableIgnoredStatusPreserves (0.00s)
    clean_ignored_content_test.go:199: an unreadable ignored status must preserve; removed [/wt/merged-unreadable]
FAIL	github.com/modu-ai/moai-adk/internal/cli/worktree	0.839s
```

The third case of the battery — `RemovesWhenIgnoredContentIsRegenerable` —
passed in the RED run by design. It is the anti-immortality control: 156 of 156
trees in this repository hold ignored content (`design.md` §A.7), so a guard
that preserved on the blunt predicate would make the whole population immortal.
It must pass before and after.

GREEN, after factoring the predicate into `ignoredContentVerdict` and calling it
from `cleanMergedWorktrees`:

```
$ go test -count=1 -run 'TestCleanMergedOnly|TestCleanStale' ./internal/cli/worktree/
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	0.470s

$ go build ./...                          → exit 0 (no output)
$ go vet ./...                            → exit 0 (no output)
$ gofmt -l internal/cli/worktree/         → no output
$ go test -count=1 ./internal/cli/worktree/... ./internal/session/...
ok  	github.com/modu-ai/moai-adk/internal/cli/worktree	5.589s
ok  	github.com/modu-ai/moai-adk/internal/session	8.484s
```

One pre-existing test moved as a consequence, and the movement is recorded
rather than tidied away. `TestRunClean_MergedOnly` drives fixture paths that do
not exist on disk, so the new guard read an unrunnable `git status --porcelain
--ignored` and failed closed — the correct behaviour, observed as a failure:

```
subcommands_test.go:681: output should contain 'Removing merged worktree', got
  "  Keeping /repo-feature [feature]: cause=ignored-check-failed; could not read ignored content: exit status 128\n"
```

The fixture now stubs the ignored read clean, because that test asserts the
removal path; the fail-closed branch is bound by
`TestCleanMergedOnly_UnreadableIgnoredStatusPreserves` instead. Production
behaviour was not relaxed to make a test pass.

### Baseline-attribution

Measured in this tree, branch `WT-worktree-reaper`, RED at `96a26b0f8` and GREEN
at the commit this section accompanies. The audit input is
`.moai/reports/t209/sync-audit-iter2.md` (FAIL 0.844, threshold 0.85; Security
0.82 tripping the must-pass firewall on N1 alone).

The doc comment at `internal/session/ignored_content.go:50` — "three removal
paths consult it" — was false when written and is true now. That is the whole of
N1: the gap predated this SPEC, the *claim* did not.

### Gaps — what was explicitly NOT observed

1. **`internal/cli` was not re-run**, and this change does not touch it. Its
   green is the attributed prior measurement from §E.2 (`ok … 768.952s`, exit 0,
   at `aa14918d7`).
2. **The full suite was not run locally** — repo rule. CI owns that verdict
   against the PR head, which is still outstanding: the branch is unpushed.
3. **No sweep was run against a real repository.** All three paths remain bound
   by stubbed-git fixtures; `clean` prunes shared worktree administrative state,
   so running it is not authorised here.
4. **`golangci-lint` was not run** — outside the authorised command set.
5. **The audit's optional findings N2 and F5–F12 are untouched**, deliberately.
   N2 (`TrimLeft(…, "*+ ")` corrupting a legal branch name beginning with `+`,
   and the `(`-skip dropping a legal `(paren` name) has zero measured exposure
   here but is a real narrowing that `git for-each-ref` would remove.

### Residual risk

- **The check→act race is unchanged and now spans one more predicate.** Every
  path classifies before it removes, so the ignored-content read on the merged
  sweep inherits the same window the other two have.
- **The guard's blast radius on the merged path is uncounted**, for the same
  positional reason recorded in §E.5.
- **One evaluation now decides three paths.** That is the point of the share,
  and it is also the concentration: an error in `IrreplaceableIgnoredEntries` is
  now wrong in three places at once rather than one.
