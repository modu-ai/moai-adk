# SPEC Review Report: SPEC-WORKTREE-REAPER-001

Iteration: 2/3
Verdict: **FAIL**
Overall Score: **0.84** (0.8375 exact; Tier L PASS threshold = 0.85)

Reasoning context ignored per M1 Context Isolation. The dispatch's
characterisations — including "v0.3.0", "the equivalence claim holds", and "EC-9
closes in the safe direction" — were treated as hypotheses to test, not as
premises. Every judgment below is anchored to an artifact line or to a command
run in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209` (branch
`WT-worktree-reaper`, HEAD `bee2c2640`) on 2026-08-24.

**Score trajectory: iter1 0.55 → iter2 0.84.** No regression; no STOP signal.
The FAIL margin is thin (0.0125) and rests on one critical finding whose fix is
bounded and mechanical. Iteration 3 remains available under the Tier L ceiling.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-WR-[0-9]*' spec.md | sort -u` → `REQ-WR-001 … REQ-WR-023`, 23 unique, zero-padded, no gaps, no duplicates; `grep -c '^- \*\*REQ-WR-' spec.md` → `23`. Within the Tier L ceiling of 25.
- **[PASS] MP-2 GEARS format compliance (requirement layer)** — judged against the 23 `REQ-WR-XXX` entries in `spec.md` §D **only**; the Given/When/Then entries in `acceptance.md` are the verification layer and are graded under Group 4. `grep -o '(Ubiquitous)\|(Event-driven)\|(State-driven)\|(Unwanted)\|(Where…)'` → 23 labels for 23 requirements; `grep -n 'Event-detected' spec.md` → **0 matches** (iteration-1 D17 fixed). Mechanical confirmation: `moai spec lint .moai/specs/SPEC-WORKTREE-REAPER-001/spec.md` → exit 0, `✓ No findings — all SPEC documents are valid`.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:L1-15 carries all 12 canonical fields (`id`, `title`, `version: "0.2.0"`, `status: draft`, `created`/`updated` ISO, `author`, `priority: P1`, `phase: "v3.1.4 target"` — not a prohibited stage name, `module`, `lifecycle: spec-anchored`, `tags`) plus optional `tier: L`. No rejected snake_case alias. `moai spec lint` exit 0.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `module: internal/cli`); no template-bound or 16-language content.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -o 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u` → `SPEC-SESSION-WORKTREE-001`, `SPEC-WORKTREE-REAPER-001`. `grep '^status:' .moai/specs/SPEC-SESSION-WORKTREE-001/spec.md` → `status: completed` (not retired/superseded/archived). No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall' <spec dir>` now returns **2 matches** (v0.2.0 additions: `plan.md:125`, `design.md:221`), so D8 is no longer auto-PASS. Both occurrences carry an explicit cross-platform clause in the same sentence/section: `plan.md:125` reads "wrapping the existing syscall, **with the per-platform mapping in `design.md` §B.5**", and `design.md` §B.5 carries a five-row unix/windows probe-mapping table plus the "windows can never assert dead, so never widens removal" clause; `spec.md` §E carries "Windows liveness is unconditionally `true` (`anchor_pid_windows.go`) … Cross-platform verification is `GOOS=windows go vet`", and AC-WR-022 runs it. Confirmed the referenced files are already build-tag-separated: `head -3 internal/session/anchor_pid_unix.go` → `//go:build !windows`; `anchor_pid_windows.go` → `//go:build windows`. Explicit cross-platform exemption clause present ⇒ PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' <spec dir>` → exactly one line, `progress.md:34`, which is a negation ("no `[NEEDS CLARIFICATION]` markers remain"), not a marker. `plan.md` and `research.md` both exist and carry zero markers.

No must-pass criterion fails. **The FAIL is driven by the rubric scores**, principally by N1 below.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75-1.0 | Requirement text is unusually precise and every v0.2.0 wording defect from iteration 1 is repaired: REQ-WR-005 now states its observable rather than "as it does today" (spec.md:L144-146); `design.md` §B.3 pins the stored lock reason vs the porcelain line in a two-row table; the §B.4 fail-closed table is exhaustive and its first row is "no opinion", not "not anchored". Deductions: **REQ-WR-021 (spec.md:L196-201) states as fact a condition that measurement shows does not exist** — "content git counts as modified or untracked that `git status --porcelain` omits" (N1); `spec.md` §E's constraint "git's own removal check is stricter than the dirty guard" asserts the opposite of what reproduces; §D's preamble is self-contradictory — "018-022 are the amendment additions … (018-023)" (spec.md:L127-129). |
| Completeness | 0.75 | 0.75 | All required sections present; `research.md` added (iteration-1 D16 closed) and it carries real Tier L weight, not padding — the two-sweep comparison table (§A), exact existing test names (§B), the `IsBranchMerged` staged-predicate finding (§C.2), the platform-probe limitation (§D.2), and the Go `-run` toolchain hazard (§E) are each load-bearing for a specific design decision. §G Out of Scope is exemplary: `grep -c '^### Out of Scope — '` → **7** H3 sub-headings, each with specific bullets. Deductions: the refusal-class pre-detection set is incomplete and one of its two named limbs is empty (N1, N2); a **third** consumer of the blind anchor guard is unnamed (N3); the DoD still says "REQ-WR-001…022" for a 23-requirement SPEC (N5). |
| Testability | 0.85 | 0.75-1.0 | **The iteration-1 FAIL driver is genuinely fixed and I verified it mechanically.** The criterion form is now `go test <pkg> -run '^<T>$' -v -count=1 2>&1 \| grep -c '^--- PASS: <T>'`, which fails on a missing test. **I ran 23 of the 23 distinct test names cited across all 24 criteria and every recorded `Pre-impl observed:` value reproduced exactly** — 10 recorded `1` returned `1`, 13 recorded `0` returned `0`, zero discrepancies (table below). Deductions: AC-WR-024 rests on the falsified premise (N1) and its stated load-bearing assertion — string inequality — is weak by construction (N4); AC-WR-022/023 are completion gates rather than discriminators, which the file states honestly rather than disguising. |
| Traceability | 0.95 | 0.75-1.0 | `grep -o 'REQ-WR-[0-9]*' acceptance.md \| sort -u` → **all 23** requirements, `REQ-WR-001 … REQ-WR-023`, up from 2 of 17 at iteration 1. Every criterion carries an explicit `Covers:` line; AC-WR-020 honestly declares it covers a craft constraint rather than a REQ. Deduction: the §F Definition-of-Done checkbox still reads "REQ-WR-001…022" (N5), an off-by-one residue of the growth. |

Aggregate: (0.80 + 0.75 + 0.85 + 0.95) / 4 = **0.8375** < 0.85.

---

## 1. Regression Check — all 19 iteration-1 findings

| # | Disposition | Verdict | Evidence |
|---|---|---|---|
| **D1** AC falsifiability | rebuilt criterion form | **FIXED (verified, 23/23)** | See § Pre-impl sample verification below. Every criterion now uses the `grep -c '^--- PASS: <T>'` form; `acceptance.md` §0 verifies the form in both directions and records the mechanical no-new-test-name proof. |
| **D2** merge fallback widens removal | analysed + accepted (REQ-WR-018, design §A.4/§A.5) | **FIXED by analysis** | Equivalence independently reproduced — see §2. |
| **D3** anchor repair too narrow | REQ-WR-019 moves decision to `internal/session` | **FIXED (partially — see N3)** | Second call site confirmed real: `grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `163:  case len(session.LiveAnchoredSessions(wt.Path, time.Now())) > 0:` inside `cleanStaleWorktrees`. But the same grep also returns line `95`, a **third** consumer the SPEC never names (N3). |
| **D4** M3 incomplete option set | O3-d added and selected | **FIXED** | Options table `design.md` §C.3 now carries O3-a…O3-d with O3-d selected; all three recorded gaps verified real (§4). |
| **D5** AC-WR-011 destructive, unobserving | split into AC-WR-023 (a)/(b) | **FIXED** | (a) is non-destructive with `auto_cleanup` temporarily false, captures output to `.moai/state/verify/t209/`, and builds the tree-local binary (`go build -o ./bin/moai ./cmd/moai`) rather than trusting `~/go/bin/moai`; (b) is gated on an operator-enumerated `expected-removals.txt` with a `comm -13` assertion. |
| **D6** safety inversion mischaracterised | restated (design §D, plan §F) | **FIXED** | `design.md` §D now says "overstated for locked trees… understated for unlocked trees", names M2 as "a correctness and legibility gain, not a rescue", and adds the strongest version of the argument (the protection is an untested side effect of git's lock handling, one `--force` from vanishing). |
| **D7** probe-seam contradiction | `(alive, determined)` seam | **FIXED** | `design.md` §B.5 specifies `func(pid int) (alive bool, determined bool)` with a five-row per-platform mapping; AC-WR-011's fourth case is explicitly justified as reachable only because of it. |
| **D8** traceability implicit | `Covers:` lines added | **FIXED** | 23/23 requirements cited (measured above). |
| **D9** AC count wrong | corrected to 24 | **FIXED** | `grep -c '^### AC-WR-'` → `24`; header sub-counts 7+9+4+1+3 = 24, each range verified against the actual IDs. |
| **D10** non-existent "existing" test names | corrected | **FIXED (verified)** | `TestPRMergeCleanup_GhPresentMergedRemoves` → 1, `…_GhAbsentBranchMergedRemoves` → 1, `…_GhAbsentEmitsBlindnessNoticeOnce` → 1. All three run and pass. |
| **D11** AC-WR-016 overstated the test | scope note added | **FIXED** | AC-WR-021 now says explicitly "the existing `ToggleOffNoOp` asserts *`wtList` was not called* plus empty output — not 'no git seam is invoked' across all seams. The criterion is stated to match what the test actually asserts". |
| **D12** unanchored `-run` | exact anchored name | **FIXED (verified)** | AC-WR-020 names `TestCleanStale_NoAskUserQuestion` anchored; measured `0` on this tree. The 31-guard hazard is recorded: `grep -rn 'func Test.*NoAskUserQuestion' internal/cli/ \| wc -l` → **31**, matching the SPEC's figure exactly. |
| **D13** measurement drift 111/43 | timestamped table + reconciliation | **FIXED** | `spec.md` §B.2 carries three timestamped rows and states the reconciliation ("three values are the expected result, not three errors") plus the 43-vs-44 boundary convention. Re-measured now: 156 entries / 112 `WT-` / 97 merged — a fourth value, which the table's own design accommodates. Not a defect. |
| **D14** dead-lock trees unreapable | REQ-WR-021 + design §B.6 inert policy | **FIXED** | `design.md` §B.6 "the confirmed-dead branch is inert"; AC-WR-012 pins it; §B.4's dead row is retained with an explicit note that REQ-WR-021 stops it implying an impossible removal. |
| **D15** stale `origin/main` | EC-10 added | **FIXED** | EC-10 states the direction argument and the force-push exception, and explicitly declines to claim safety ("recorded as a bounded residual, not a claim of safety"). `research.md` §C.3 confirms neither sweep fetches. |
| **D16** Tier L without research.md | added | **FIXED** | `research.md` present, 197 lines, substantive (see Completeness row). |
| **D17** GEARS label errors | relabelled | **FIXED (verified)** | `grep -n 'Event-detected' spec.md` → 0 matches. |
| **D18** baseline-relative REQ-WR-005 | restated | **FIXED** | Now "shall be the sole merge source and the squash-blindness notice shall be emitted exactly once per sweep invocation". |
| **D19** lock reason vs porcelain line | design §B.3 table | **FIXED** | Two-row table pins stored reason vs rendered line; AC-WR-008 asserts the `locked ` prefix is stripped and that a bare `locked` line maps to an empty reason. |

**19 of 19 dispositioned. 18 fully fixed; D3 fixed for the two named consumers but incomplete for a third (N3). Zero regressed.**

### Pre-impl sample verification (D1) — 23 of 23 test names, 100% reproduction

Every distinct test name appearing in any of the 24 criteria was run with the exact command form the criterion specifies. Not a sample of convenience — this is the complete set.

| Test | Recorded | Observed | |
|---|---|---|---|
| `TestPRMergeCleanup_GhPresentMergedRemoves` | 1 | **1** | ✓ |
| `TestPRMergeCleanup_GhNoAnswerConsultsGitFallback` | 0 | **0** | ✓ |
| `TestPRMergeCleanup_GhOpenSkipsGitFallback` | 0 | **0** | ✓ |
| `TestPRMergeCleanup_UndeterminedMergePreserves` | 0 | *(name absent — grep-confirmed)* | ✓ |
| `TestPRMergeCleanup_GhAbsentBranchMergedRemoves` | 1 | **1** | ✓ |
| `TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce` | 1 | **1** | ✓ |
| `TestPRMergeCleanup_GhPresentSeesSquashMerge` | 1 | **1** | ✓ |
| `TestPRMergeCleanup_ZeroUniqueCommitPreserved` | 0 | **0** | ✓ |
| `TestParseWorktreeList_CapturesLockReason` | 0 | **0** | ✓ |
| `TestLockAnchor_LivePidAnchored` | 0 | **0** | ✓ |
| `TestLockAnchor_DeadPidNotAnchored` | 0 | *(0 — `grep -rn 'func TestLockAnchor_' internal/` → no matches)* | ✓ |
| `TestLockAnchor_IndeterminateIsAnchored` | 0 | *(same grep)* | ✓ |
| `TestLockAnchor_DeadLockRemovalIsInert` | 0 | *(same grep)* | ✓ |
| `TestPRMergeCleanup_AnchoredSessionSkipsRemoval` | 1 | **1** | ✓ |
| `TestAnchorDecision_RegistryOnlyPathStillAnchors` | 0 | **0** | ✓ |
| `TestPRMergeCleanup_T207SamplePreservedByLock` | 0 | **0** | ✓ |
| `TestCleanStale_LockAnchoredWorktreeKept` | 0 | **0** | ✓ |
| `TestCleanStale_KeepsAnchoredWorktree` | 1 | **1** | ✓ |
| `TestPRMergeCleanup_PorcelainFailureRemovesNothing` | 0 | **0** | ✓ |
| `TestPRMergeCleanup_WorktreeListErrorFailOpen` | 1 | **1** | ✓ |
| `TestCleanStale_JSONEmitsAllTreesWithStates` | 0 | **0** | ✓ |
| `TestCleanStale_PreviewsByDefault` | 1 | **1** | ✓ |
| `TestCleanStale_RemovesWithYes` | 1 | **1** | ✓ |
| `TestCleanStale_BaseDefaultsToOriginMain` | 0 | **0** | ✓ |
| `TestCleanStale_NoAskUserQuestion` | 0 | **0** | ✓ |
| `TestPRMergeCleanup_RefusalClassPreDetected` | 0 | **0** | ✓ |

**20 executed as `go test` invocations, 3 (`TestLockAnchor_*` siblings) confirmed absent by `grep -rn 'func TestLockAnchor_\|func TestAnchorDecision_' internal/` returning exit 1 / no matches. Zero discrepancies.** The grep-token baseline also reproduces: `grep -rn 'mergeStateUndetermined' internal/ | wc -l` → **0**, as `acceptance.md` §0 records.

This is the strongest part of the revision. The recorded observations are real, not asserted.

---

## 2. Adjudication — the rejected D2 remedy

**The equivalence claim holds. I verified it myself, in both directions, in a scratch repository built and destroyed under this worktree.**

```
$ git init -q -b main; <one commit incl. .gitignore>
$ git worktree add -q ../wt-a -b WT-ignored
$ git worktree add -q ../wt-b -b WT-fresh
$ git branch --merged main --format='%(refname:short)'
WT-fresh WT-ignored main            # zero-unique-commit branches ARE listed as merged

$ cd ../wt-a; <commit one file>; git rev-list --count main..WT-ignored
1
$ cd ../base; git branch --merged main --format='%(refname:short)'
WT-fresh main                       # the branch with a unique commit DROPPED OUT
```

A branch appears in `git branch --merged <base>` exactly when its tip is an ancestor of base, which is exactly when `git rev-list --count <base>..<branch>` is 0. **There is no second predicate in `staleKeepReason` to copy.** Iteration 1's D2 remedy rested on a misreading of `IsBranchMerged`, whose S1 stage is literally `git branch --merged` (`research.md` §C.2, confirmed against `internal/core/git/worktree.go`).

**Stated plainly, as required: D2's finding is fixed by analysis-and-acceptance, not by exclusion. The prescribed remedy was wrong and the SPEC is right to reject it.** `design.md` §A.5 records the disagreement and its resolution correctly, and the rejection is documented rather than silent — which is the right way to overrule an auditor.

**Is accepting the class *sound*, independently of the redundancy argument?** The SPEC argues three backstop layers (`design.md` §A.4):

1. **Committed work** — cannot be lost; an ancestor of base holds nothing base lacks. **Sound.** Verified above.
2. **Tracked or untracked work** — caught by `worktreeIsDirty` → `git status --porcelain`, which reports untracked entries as `??`. **Sound.** Independently confirmed: an untracked non-ignored file also makes non-forced `git worktree remove` exit 128.
3. **Ignored work** — "git's own removal check is stricter still (§A.6)". **NOT SOUND. This layer does not exist** — see N1.

So the accept-the-class decision survives on two of its three stated layers, and the SPEC's own claim that it is "backstopped at every layer rather than at two" (`design.md` §A.6, closing sentence) is false. The residual exposure is narrow but real: a merged, clean-to-porcelain worktree holding only gitignored content (a local `.env`, a build cache) is removed and that content is destroyed. **The decision is defensible; the stated justification overclaims and must be corrected.**

---

## 3. EC-9 — the measurement does NOT reproduce

**This is the critical finding.** I reproduced the fixture twice, independently, and got the opposite result both times.

```
# fixture: repo with one commit including a committed .gitignore containing "build/",
# a linked worktree created AFTER that commit (so .gitignore is present in it),
# and one ignored file build/artifact.bin inside the worktree.

$ cat ../wt-c/.gitignore
build/                                     # ignore rule IS present in the worktree

$ cd .../wt-c && git status --porcelain | wc -l
0                                          # dirty guard reads CLEAN — as ec9-measurement says
$ git status --porcelain --ignored | wc -l
1                                          # the ignored entry is there

$ cd .../base && git worktree remove ../wt-c
$ echo $?
0                                          # ← REMOVED. exit 0. No refusal.
$ ls .../audit2scratch/
base  wt-a                                 # wt-c is GONE; build/artifact.bin destroyed
```

Run twice (`wt-b`, `wt-c`), same result. `git version 2.50.1 (Apple Git-155)`.

**Contrast — the case that DOES refuse:**

```
$ git worktree add -q ../wt-d -b WT-untracked
$ echo notignored > ../wt-d/loose.txt      # untracked, NOT ignored
$ git worktree remove ../wt-d
fatal: '../wt-d' contains modified or untracked files, use --force to delete it
exit=128
```

That message is byte-identical to the one `ec9-measurement.md` §Q1 quotes. **The recorded measurement almost certainly used a fixture whose file was untracked rather than ignored** — e.g. `.gitignore` committed after the worktree was created, or `git status --porcelain` read in the base repo rather than in the worktree. An untracked non-ignored file is a condition `git status --porcelain` *does* report as `??`, so it is already caught by the existing dirty guard and establishes nothing new.

**The hazard closes in the UNSAFE direction, not the safe one.** `git status --porcelain` omits ignored files AND `git worktree remove` also ignores them — the two agree, and nothing backstops the gap. The shape is live in this repository right now: this very worktree reports `git status --porcelain | wc -l` → **0** and `git status --porcelain --ignored | wc -l` → **5**.

---

## 4. New material introduced by the revision

### REQ-WR-021's pre-detection set — incomplete, and one limb is empty

REQ-WR-021 defines the refusal class as two conditions: locked, **or** "holds content git counts as modified or untracked that `git status --porcelain` omits". Measured:

| Candidate cause | `git status --porcelain` sees it? | Non-forced `git worktree remove` refuses? | Detected by REQ-WR-021's set? |
|---|---|---|---|
| Locked worktree | no | **yes** (exit 128, `cannot remove a locked working tree`) | yes — lock line |
| Untracked non-ignored file | **yes** (`??`) | yes (exit 128) | already the dirty guard's job |
| **Ignored-only content** | no | **NO — removes, exit 0** | probe exists, but the premise is false |
| **Populated submodule** | **no** (0 lines) | **yes** — `fatal: working trees containing submodules cannot be moved or removed`, exit 128 | **NO — missed** |
| Missing worktree directory | n/a | **no** — removes, exit 0 | n/a |

Both edge cases I tested beyond the SPEC's set were measured, not inferred. The submodule case is the *only* verified member of the class the SPEC invented the `--ignored` probe for — clean porcelain, git refuses anyway — and it is the one the SPEC misses. It is also not curable by `--force` ("cannot be moved or removed", flat). It is **not live in this repository** (`.gitmodules` absent), which caps its severity.

Net: with the ignored limb empty and submodules absent here, **the refusal class in this repository reduces to exactly what REQ-WR-021 covered before the v0.2.0 generalisation: locked trees.**

### REQ-WR-023 / AC-WR-024 — the inequality assertion is largely trivial

Two problems, one inherited and one intrinsic:

- **Inherited**: cause (b) as specified cannot occur (§3). A run-phase implementer will build it with seam-injected fakes, producing a green test that asserts nothing true about git.
- **Intrinsic**: REQ-WR-023 requires the notice to "name which cause applied"; AC-WR-024's stated load-bearing assertion is only that the two strings **differ**. Any two distinct literals satisfy it, including `preserved (cause A)` / `preserved (cause B)`, which name nothing. The criterion is strictly weaker than the requirement it covers.

The *motivation* is sound — the SPEC's reasoning that one indistinguishable permanently-recurring symptom for two separately-fixable states is bad is correct, and worth keeping.

### REQ-WR-019 / AC-WR-015 — the second call site is real; a third is unnamed

Verified: `grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` returns **two** lines.

- `163:` inside `cleanStaleWorktrees` — the consumer REQ-WR-019 names. **Real. The placement in `internal/session` does reach it**, and AC-WR-015 exercises it with an empty registry, which is the right fixture.
- `95:` inside the `--merged-only` path — **unnamed by any requirement**, and it is the most exposed of the three. Its own in-code comment says so:

  ```go
  // Anchor guard (t73): --merged-only has no dirty guard of its
  // own, so this is the only protection between the sweep and a
  // live lane's tree.
  ```

`research.md` §F lists `--merged-only` under "surveyed and found NOT to apply", with the reason "removes merged worktrees without the dirty/anchor pairing that `--stale` applies". That reason is an argument for **including** it, not excluding it. The SPEC's own framing — "repairing only the automatic sweep leaves the blind guard on the surface that can remove more" (`research.md` §A) — applies verbatim.

### O3-d — the recorded gaps are real and sufficient

All three verified directly against `internal/cli/worktree/clean.go`:

| Gap claimed | Verified |
|---|---|
| No machine-readable output | ✓ flag set is exactly `--merged-only`, `--stale`, `--yes`, `--base` (L32-35); `grep '"json"'` → 0 hits |
| `--base` defaults to local `main` | ✓ `35: cmd.Flags().String("base", "main", "Base branch for --merged-only and --stale checks")` |
| Same blind anchor source | ✓ `163: case len(session.LiveAnchoredSessions(...)) > 0:` |

Each maps to a distinct requirement (REQ-WR-012 / 022 / 019) and each is independently observable. **The gaps are real and they do justify extension over construction.** The rejection of O3-b is correct and correctly grounded in the Simplicity ladder. This is the strongest design section in the document.

### research.md — carries Tier L weight

Not padding. Six sections, each load-bearing for a specific decision: §A's two-sweep table is the evidence for REQ-WR-019; §B's exact test names are what fixed iteration-1 D10 and D12 (I confirmed the 31-guard figure and every cited name); §C.2's `IsBranchMerged` staged-predicate finding is the evidence that settled D2; §D.2's platform-probe limitation drives the `(alive, determined)` seam; §E's Go `-run` finding is correctly generalised as a repo-wide authoring hazard rather than a t209 quirk; §F is an explicit negative survey (L2 empty, `git worktree prune` 0 prunable). One defect: §D.4a restates the falsified EC-9 claim and must be corrected with the rest of N1.

---

## 5. Growth regressions (17→23 REQ, 18→24 AC)

Checked for restatement, scope creep, and cross-artifact inconsistency.

- **No two requirements restate each other.** The six additions each occupy distinct ground: 018 (accept a removal class), 019 (placement/reach), 020 (record a residual), 021 (pre-detect refusals), 022 (base ref), 023 (notice attributability).
- **Scope creep**: the only growth past "repair the shipped reaper" is the REQ-WR-021 generalisation + REQ-WR-023 + AC-WR-024 + EC-11 cluster — and §3 shows it was driven by a measurement that does not reproduce. This is the one place where growth was not earned.
- **Counts**: REQ 23 ✓, AC 24 ✓, header sub-counts sum correctly ✓, `progress.md` §E.1 states 23/24 ✓. Two textual residues survive (N5).
- **AC↔REQ mapping**: complete in both directions (23/23 covered; every AC names its REQs).
- **Misfiling**: AC-WR-024 is a `prMergeCleanup` criterion but sits physically inside `§C — M3: non-WT-* coverage`, between AC-WR-020 and §D. The header classifies it separately ("refusal-class one"), so the intent is clear, but the placement contradicts the section heading.

---

## Defects Found (structured defect-list)

**N1.** EC9-MEASUREMENT-DOES-NOT-REPRODUCE — `.moai/reports/t209/ec9-measurement.md` §Q1 → `spec.md` §E (L255-260) + `design.md` §A.6, §A.4 reason 3, §B.6a + `research.md` §D.4a + `acceptance.md` EC-9, EC-11, AC-WR-024 + `progress.md` §E.1 "Q1" — The claim "non-forced `git worktree remove` counts gitignored files and refuses, so git's own check is stricter than the dirty guard" is **false**. Measured twice on a faithful fixture: `git status --porcelain` → 0, `git status --porcelain --ignored` → 1, `git worktree remove` → **exit 0, tree deleted, ignored file destroyed**. The quoted `contains modified or untracked files` refusal is reproducible only with an **untracked non-ignored** file — a condition the dirty guard already catches. Consequences: (a) `spec.md` §E's constraint is inverted; (b) `design.md` §A.4's third backstop layer does not exist and §A.6's "backstopped at every layer rather than at two" is false; (c) REQ-WR-021's second limb names an empty condition; (d) EC-9's "Preserved" is wrong — the tree is removed; (e) EC-11 and AC-WR-024 pair a real cause (lock) with a non-existent one; (f) the hazard closes in the **unsafe** direction — a merged clean-to-porcelain tree holding a local `.env` or build cache is silently destroyed, which is a genuine data-loss path the SPEC now believes is closed. — **Evidence**: scratch repo under this worktree, run twice; contrast case exit 128; `git version 2.50.1 (Apple Git-155)`; live shape in-tree — this worktree reports porcelain `0` / `--ignored` `5`. — Severity: **critical** — Class: **blocking** — **Required fix**: correct the measurement record and all six downstream sites. **Keep the `git status --porcelain --ignored` probe — it is still the right mechanism — but invert its rationale**: not "git will refuse anyway, so pre-detect to avoid a permanent failure notice", but "**git will not refuse, and the dirty guard is blind, so ignored content is destroyed unless this SPEC preserves it**". Restate EC-9 as "removed today; ignored content lost", and decide explicitly whether REQ-WR-021 preserves such a tree or accepts the loss. Re-derive REQ-WR-023 / AC-WR-024 / EC-11 from whatever refusal causes survive (§N2).

**N2.** REFUSAL-CLASS-PRE-DETECTION-INCOMPLETE — `spec.md` REQ-WR-021 (L196-211) / `design.md` §B.6a — The pre-detection set is `{lock line, --ignored probe}`. Measured, a **populated submodule** makes non-forced `git worktree remove` exit 128 (`fatal: working trees containing submodules cannot be moved or removed`) while `git status --porcelain` reports 0 lines — the exact "clean porcelain, git refuses" shape REQ-WR-021 exists for, and neither the lock check nor the `--ignored` probe catches it. It is also not curable by `--force`. Two further causes were tested and are **not** refusals (recording them so the next reader need not re-measure): a **missing worktree directory** removes cleanly (exit 0); the primary checkout is excluded by both sweeps already (`protectedWorktreePaths`, and the `WT-` prefix filter). — **Evidence**: scratch repo — submodule added with `-c protocol.file.allow=always`, `submodule update --init` in the worktree, `git status --porcelain | wc -l` → 0, `git worktree remove` → exit 128 with the quoted message; missing-dir case → exit 0; `.gitmodules` absent from this repository. — Severity: **major** — Class: **blocking** — **Required fix**: either add the submodule condition to the pre-detection set with an edge case, or record it in §G Out of Scope with the measured reason that this repository has no submodules — but do not leave the class defined as exhaustive when it is not. Note that after N1's correction the class in this repository reduces to **locked trees alone**, which means the v0.2.0 generalisation should be re-derived rather than patched.

**N3.** THIRD-CONSUMER-OF-THE-BLIND-ANCHOR-GUARD-UNNAMED — `spec.md` REQ-WR-019 (L179-182) / `research.md` §A, §F — REQ-WR-019 names "both consumers", but `internal/cli/worktree/clean.go` contains **two** `LiveAnchoredSessions` call sites: line 163 (`cleanStaleWorktrees`, named) and line 95 (the `--merged-only` path, unnamed). The unnamed one is the most exposed of the three: its own comment records that it has **no dirty guard**, so the blind anchor check is the *sole* protection between that sweep and a live lane's tree. `research.md` §F excludes `--merged-only` for a reason ("removes merged worktrees without the dirty/anchor pairing") that argues for inclusion. Since REQ-WR-019 exports the decision from `internal/session`, wiring the third site is a one-line change. — **Evidence**: `grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `95:`, `163:`; the verbatim in-code comment at L92-94. — Severity: **major** — Class: **blocking** — **Required fix**: either extend REQ-WR-019 to name all three call sites and add a criterion for the `--merged-only` path, or add an explicit `### Out of Scope — the --merged-only sweep` entry stating why the surface with no dirty guard is deliberately left blind.

**N4.** AC-WR-024-ASSERTION-WEAKER-THAN-ITS-REQUIREMENT — `acceptance.md` AC-WR-024 (§C, closing paragraph) — The criterion's declared load-bearing assertion is that the two preserve-notice strings are **unequal**. REQ-WR-023 requires the notice to *name which cause applied*. Inequality is satisfied by any two distinct literals that name nothing, so the criterion is strictly weaker than the requirement it claims to cover — and, compounded by N1, one of its two fixtures models a condition that does not exist. — **Evidence**: `acceptance.md` AC-WR-024 `Covers: REQ-WR-021, REQ-WR-023`; the criterion text asserts only difference. — Severity: **minor** — Class: **blocking** — **Required fix**: assert that each notice contains a cause-specific token (e.g. the locked notice matches `locked`, the other matches its cause word), not merely that the two differ. Rebuild the second fixture once N1 settles which causes are real.

**N5.** COUNT-AND-PLACEMENT-RESIDUES-FROM-THE-GROWTH — `acceptance.md` §F (DoD bullet 5) / `spec.md` §D preamble (L127-129) / `acceptance.md` §C — Three textual residues of the 17→23 growth: (a) the DoD reads "Every requirement **REQ-WR-001…022** named by at least one `Covers:` line" for a 23-requirement SPEC — the checkbox understates the obligation it is meant to enforce (traceability is in fact complete, so this is a stale label, not a gap); (b) `spec.md` §D's preamble contradicts itself — "**018-022** are the amendment additions, placed in the milestone they belong to rather than appended at the end (**018-023**)"; (c) AC-WR-024 sits physically inside `§C — M3: non-WT-* coverage` although it is a `prMergeCleanup` criterion that the file's own header classifies separately. — **Evidence**: `grep -o 'REQ-WR-[0-9]*' acceptance.md | sort -u` → 23 ids (so the DoD is wrong, not the coverage); spec.md:L127-129 verbatim. — Severity: minor — Class: **blocking** (mechanical, and (a) would cause a run-phase self-check to under-verify) — **Required fix**: `001…023`; drop the contradictory parenthetical; move AC-WR-024 to its own §C-bis or into §B beside the other `prMergeCleanup` criteria.

**N6.** VERSION-NOT-BUMPED-FOR-THE-POST-MEASUREMENT-ADDITIONS — `spec.md` frontmatter L4 / §A History — The dispatch describes this revision as **v0.3.0**; the artifact says `version: "0.2.0"`. The measurement round that added REQ-WR-023, AC-WR-024, EC-9/EC-11 and generalised REQ-WR-021 landed *after* the v0.2.0 amendment entry was written, and was folded into that entry in place ("A follow-up measurement round then settled…") rather than cut as a new version. The SPEC is internally self-consistent at 0.2.0, so this is a records defect, not a contradiction — but a reader diffing against a "v0.3.0" reference will not find it. — **Evidence**: `grep -n '^version:' spec.md` → `4:version: "0.2.0"`; `spec.md` §A L20 carries both amendments in one bullet. — Severity: minor — Class: **optional** — **Required fix**: bump to `0.3.0` with its own HISTORY bullet covering the measurement-round changes, or leave at 0.2.0 and correct whatever cites 0.3.0.

**N7.** EC-8-CONSISTENCY-BREAK-ONCE-N1-IS-FIXED — `acceptance.md` §E EC-8 — EC-8 states "Clean ⇒ removable and **nothing is lost**". After N1, that is false for a tree whose only content is ignored: it is clean to the dirty guard, removable by git, and its ignored content *is* lost. Flagged separately from N1 because it is a distinct sentence in a distinct table that a fix confined to §A.6 would miss. — **Evidence**: `acceptance.md` EC-8 row; §3 measurement above. — Severity: minor — Class: **blocking** — **Required fix**: qualify EC-8 to "nothing **committed, tracked, or untracked** is lost; ignored content is not protected — see EC-9".

### Positive findings (recorded, not defects)

- **The falsifiability repair is genuine and complete.** 23 of 23 recorded pre-implementation observations reproduce exactly. This is what iteration 1 asked for and it was delivered without a single overstatement — including the honest `Pre-impl observed: not run` on AC-WR-022/023, where a discriminator does not exist.
- **The mandatory sample is correctly constructed.** t207 verified in-tree: `git worktree list --porcelain` shows `branch refs/heads/WT-web-live-todo` + `locked claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)`; `git branch --no-merged origin/main | grep -c 'WT-web-live-todo'` → 1. AC-WR-014 exercises it **against the lock path specifically, with the registry emptied**, and states why ("t207 is the one tree the registry can see, so a criterion that let the registry answer would still be 4-of-5 blind"). This is exactly right.
- **Overruling the auditor was done correctly.** `design.md` §A.5 states the disagreement, cites the measurement, distinguishes the surviving finding from the rejected remedy, and does not quietly drop either. That is the model for how a SPEC should reject an audit prescription.
- **`design.md` §D is more honest than it needed to be** — it declines to claim M2 rescues live sessions, then supplies the stronger argument (a protection arising from an untested side effect of git's lock handling is one refactor from vanishing). Downgrading one's own deliverable to state it accurately is rare and worth recording.
- **The O3-d reversal is well-grounded** and its three gaps all verified real.

---

## Recommendation

**FAIL — narrowly, at 0.84 against a 0.85 Tier L threshold, and for one reason.**

The revision did what iteration 1 demanded. The criterion set is now falsifiable and I verified every recorded observation. Traceability went from 2-of-17 to 23-of-23. The M3 decision was retaken against a complete option set and the new decision is correct. The D2 disagreement was settled by measurement and the SPEC's position is right. Eighteen of nineteen findings are fully closed and none regressed.

What blocks it is that **one of the two measurements the revision relied on does not reproduce, and it now carries a requirement, a criterion, and two edge cases.** The SPEC's own discipline is what indicts it here: it correctly refused to let an unmeasured claim stand at v0.1.0, and then built new scope on a measurement that was taken against the wrong fixture. A recorded observation that does not reproduce is worse than an open question, because it looks settled.

Fix in this order:

1. **Re-measure EC-9 and correct all six downstream sites** (N1). Keep the `--ignored` probe; invert its justification. The mechanism survives; the premise does not.
2. **Re-derive the refusal class** (N2) from what actually refuses — locked trees here, plus submodules if you choose to cover them — and decide whether REQ-WR-021/023 and AC-WR-024 still earn their place once the second cause is gone. Reverting to the v0.1.0 lock-only scope is a legitimate outcome.
3. **Name the third consumer or exclude it explicitly** (N3). One line of code; one line of SPEC.
4. **Strengthen AC-WR-024 to assert cause naming, not string inequality** (N4).
5. Sweep the residues: DoD `001…023`, the §D preamble contradiction, AC-WR-024's placement (N5), EC-8's "nothing is lost" (N7), and the version label (N6, optional).

Optional-class findings (N6) are surfaced for the orchestrator's discretion and do not contribute to the verdict. The verdict rests on N1, with N2/N3 as supporting blockers.

---

## Evidence appendix — commands run in this audit

```
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209
git branch --show-current / rev-parse --short HEAD  → WT-worktree-reaper / bee2c2640
moai spec lint .../spec.md                          → exit 0, "✓ No findings"
grep -o 'REQ-WR-[0-9]*' spec.md | sort -u           → REQ-WR-001…023 (23, no gaps/dupes)
grep -c '^- \*\*REQ-WR-' spec.md                    → 23
grep -c '^### AC-WR-' acceptance.md                 → 24
grep -o 'REQ-WR-[0-9]*' acceptance.md | sort -u     → all 23
grep -n 'Event-detected' spec.md                    → 0 matches
grep -c '^### Out of Scope — ' spec.md              → 7
grep -rn 'NEEDS CLARIFICATION' <spec dir>           → 1 (progress.md:34, a negation)
grep -rn 'syscall' <spec dir>                       → 2 (plan.md:125, design.md:221)
head -3 internal/session/anchor_pid_{unix,windows}.go → //go:build !windows / //go:build windows
grep '^status:' .../SPEC-SESSION-WORKTREE-001/spec.md → completed
grep -rn 'mergeStateUndetermined' internal/ | wc -l → 0
grep -rn 'func Test.*NoAskUserQuestion' internal/cli/ | wc -l → 31
grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go → 95, 163
grep -n 'String("base"' internal/cli/worktree/clean.go → 35: default "main"
grep '"json"' internal/cli/worktree/clean.go        → 0 hits
grep -rn 'func TestLockAnchor_\|func TestAnchorDecision_' internal/ → no matches (exit 1)

23 × go test <pkg> -run '^<T>$' -v -count=1 | grep -c '^--- PASS: <T>'
                                                    → 23/23 match recorded Pre-impl values

git worktree list --porcelain | grep -c '^worktree ' → 156 ; '^branch refs/heads/WT-' → 112
git branch --merged origin/main | grep -c 'WT-'     → 97
git worktree list --porcelain (t207 entry)          → branch WT-web-live-todo
                                                       locked claude session t207 (pid 36912 …)
git branch --no-merged origin/main | grep -c 'WT-web-live-todo' → 1
git status --porcelain | wc -l (this tree)          → 0
git status --porcelain --ignored | wc -l (this tree) → 5

scratch (built + destroyed under this worktree):
  branch --merged lists zero-unique-commit branches  → WT-fresh WT-ignored main
  branch with 1 unique commit drops out              → rev-list=1, absent from --merged
  ignored-only tree: porcelain 0, --ignored 1
    git worktree remove                              → exit 0, TREE DELETED  ← EC-9 falsified
    (repeated on a second fixture)                   → exit 0, TREE DELETED
  untracked non-ignored file: git worktree remove    → exit 128 "contains modified or
                                                        untracked files"
  populated submodule: porcelain 0
    git worktree remove                              → exit 128 "working trees containing
                                                        submodules cannot be moved or removed"
  missing worktree directory: git worktree remove    → exit 0
  git --version                                      → 2.50.1 (Apple Git-155)
```
