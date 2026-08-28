# SPEC-MOVING-REF-GUARD-001 M4 — corpus triage record

Card t342 · milestone M4 · **classification only, no remediation**.

`plan.md` §F M4 marks this milestone [HARD] no-edit, and `plan.md` §G names bulk-pinning the corpus
as this card's dominant failure mode. This record exists precisely so the classification survives
without the edit: **no `<!-- moving-ref-ok: -->` marker was added anywhere, no SHA was pinned
anywhere, and no SPEC artifact outside this SPEC's own directory was touched.** Every remedy below
is a recommendation addressed to the finding's owning SPEC, not an action taken here.

## §1. How the findings were produced

```
$ go run ./cmd/moai spec lint --json > <out>
rc=0
```

Run in worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342`, branch `WT-moving-ref-guard`,
at HEAD `1ed8a9997`, working tree clean at the time of the run. The rule under test is
`internal/spec/lint_movingref.go` as landed by M3; no implementation file was modified for this
milestone.

Code tally over the whole `.moai/specs` catalogue:

| Code | Findings |
|---|---|
| **MovingRefUnpinned** | **107** |
| MissingExclusions | 24 |
| StatusGitConsistency | 18 |
| FrontmatterInvalid | 14 |
| LegacyEARSKeyword | 7 |
| OwnershipTransitionInvalid | 1 |

All 107 `MovingRefUnpinned` findings are distinct `file:line` pairs — no duplicates, despite
`Check` being invoked once per `SPECDoc` and scanning the whole SPEC directory each time. None is
the REQ-MRG-003 incomplete-marker variant (0 of 107): no `moving-ref-ok` marker exists anywhere in
the corpus yet, which is expected — M2 defined the surface and nothing has used it.

## §2. Finding count vs the §B.3 prevalence figure of 42

`progress.md` §E.1 states that the plan-phase 42 is a **grep prevalence measurement, not rule
output**, that the two may differ, and that no criterion assumes they agree. They differ. The
difference is fully attributed below; neither figure was adjusted to meet the other.

| Population | Count |
|---|---|
| Rule findings, whole catalogue | **107** |
| Rule findings, this SPEC's own directory excluded | **97** |
| §B.3 filter-4 grep, re-run at HEAD `1ed8a9997`, own directory excluded | **42** |

The gap is **97 − 42 = +55**, and it decomposes exactly:

- **+56 from the REQ-MRG-006 divergence-figure branch.** 61 of the 97 external findings carry
  `rev-list --count --left-right`; 56 of those do NOT match the §B.3 claim-marker alternation and
  therefore fire only via the divergence branch, which filter 3 never contained. This branch was
  added at M3 as a SHOULD-tier variant and kept (progress.md §E.2 M3 Claim 6). Measured:

  ```
  $ cut -f2 <external findings> | grep -ciE 'byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로'
  41
  $ cut -f2 <external findings> | grep -cE 'rev-list\s+--count\s+--left-right'
  61
  ```
  41 + 61 − 97 = 5 lines satisfy both branches.

- **−1 from the REQ-MRG-008 frozen-baseline exclusion**, which filter 4 does not have. The claim
  branch yields 41 against grep's 42; set-differencing the two populations names the single line:

  ```
  $ comm -23 <grep 42 locations> <rule 97 locations>
  .moai/specs/SPEC-DESIGN-MOAIWEBV2-002/acceptance.md:14
  ```

  That line reads `` `$BASELINE_SHA` below is the pre-flight-recorded origin/main SHA … `` — the R2
  remedy in use, correctly exempted by the rule and correctly counted by the blunter grep. This is
  the rule being *more* accurate than the prevalence measurement, not less.

41 + 56 = 97. The §B.3 figure of 42 remains correct for what it measured; it is reproduced verbatim
at this HEAD (`42`), and the M3 record's re-measurement of it stands unchanged.

## §3. Classification tally

The predicate applied is `spec.md` §D.1 (Tests 1-4) with `verification-claim-integrity.md` §2.1 as
its doctrine home. Classification and remedy are kept as **two separate steps** per §D.1 — the class
does not name the remedy.

| Class | Whole catalogue | External (own SPEC excluded) | This SPEC's own directory |
|---|---|---|---|
| anchor | **48** | 48 | 0 |
| subject / S1 | **56** | 48 | 8 |
| subject / S2 | **2** | **0** | 2 |
| already-frozen | **1** | 1 | 0 |
| unclassifiable | **0** | 0 | 0 |
| **total** | **107** | **97** | **10** |

### §3.1 The S2 count corroborates §B.7

`spec.md` §B.7 measured R4's reachable scope in the live corpus as **empty** on two independent
probes, and predicted that its future occupants would be created by the doctrine shipping alongside
it — "every R4 remediation M1 recommends produces exactly one".

Re-running both §B.7 probes against the rule's 97 external findings (rather than the grep 42):

```
$ probe 1 — grep -ciE 'run `git|by running|measure at entry|reference (value|reading)|at read time'
1
$ probe 2 — grep -ciE 'reference|as of|at entry|re-measure|remeasure|재측정'
1
```

The single hit is `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/plan.md:116`, which reads
"the orchestrator MUST run `git fetch …` and verify `0 N` or `0 0`" — a directive, with no demoted
dated reference value, so it is a probe false positive rather than an R4-form line. **External R4
occupancy is 0**, unchanged from §B.7.

Both S2 findings are inside this SPEC's own directory (findings 27 and 32), and both are the R4 form
this SPEC's own doctrine created — exactly the prediction. They are flagged because the REQ-MRG-010
lint exclusion is **deferred by operator decision (Q0 option C)**, which `lint_movingref.go` records
as a known consequence.

### §3.2 The three §B.3-named classes all reproduce under the rule

| §B.3 claim | Reproduced? |
|---|---|
| anchor, unpinned — `AC-TST-010`, `AC-AFS-012`, `AC-MSD-010`, `AC-WC-009`, `AC-CFS-010`, `AC-NS2-005` | yes — findings 97/100, 55-57, 34/35, 107, 3, 36/37 |
| subject, correctly unpinned — `AC-COORD-016`, `REQ-LB-006` | yes, and **both are flagged** — findings 82 and 53 |
| already remediated by freezing — `SPEC-DESIGN-MOAIWEBV2-002` | partly: `acceptance.md:14` is correctly exempted, `plan.md:36` is flagged (finding 7) |

The second row is the expected L2/L4 outcome, not a defect: the detector reads shape, never subject.
Both lines are correct as written and **must not be pinned**; the marker is what they are for.

### §3.3 One detector observation, offered as an input to this SPEC's own review

Finding 7 (`SPEC-DESIGN-MOAIWEBV2-002/plan.md:36`) is the R2 remedy being *prescribed* —
"Record `BASELINE_SHA=$(git rev-parse origin/main)` BEFORE the first run-phase commit". It is
flagged because `frozenBaselinePattern` matches a variable **reference** (`$BASELINE_SHA`) and not
the **assignment** form that creates it. The sibling line that uses the variable
(`acceptance.md:14`) is correctly exempted, so the SPEC is not mis-served in substance — but the
line that teaches R2 is flagged while the line that applies it is not.

This is recorded as an observation, not acted on: changing the pattern is a Go source change, which
M4 does not make. Whether it is worth a follow-up belongs to the operator and to this SPEC's owner.

## §4. AC-MRG-010 — both decider forms, verbatim

Measured at HEAD `1ed8a9997`, `BASELINE_SHA = d566ecc7511e1954e3aeb1dff3a60afa5be1089b` cited as a
literal and never re-resolved from `origin/develop`.

**Form A — the criterion as literally written in `acceptance.md:161`:**

```
$ git diff --name-only "$BASELINE_SHA"..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/spec.md
.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/progress.md
.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/spec.md
```

**NON-EMPTY — 4 paths.** Attribution:

```
$ git log --format='%h %s' --no-merges "$BASELINE_SHA"..HEAD -- <those two directories>
f9c827217 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): correct two overstated claims in the sync signal (t322)
28f16d030 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): backfill sync_commit_sha (t322)
bc66c30b7 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): sync-phase close (3-phase plan->run->sync) (t322)
ed68889e3 docs(SPEC-SYNC-STRATEGY-KEY-001): terminal transition implemented -> completed (t303)
```

All four commits are cards t322 and t303, landed on mainline and absorbed into this branch by its
two merge commits. Each was confirmed an ancestor of `origin/develop` (re-fetched immediately before
the judgment; `origin/develop` = `77b2bcae6` at that moment):

```
$ git merge-base --is-ancestor f9c827217 origin/develop   # rc=0
$ git merge-base --is-ancestor 28f16d030 origin/develop   # rc=0
$ git merge-base --is-ancestor bc66c30b7 origin/develop   # rc=0
$ git merge-base --is-ancestor ed68889e3 origin/develop   # rc=0
```

**None was authored on `WT-moving-ref-guard`.** This is an attributed observation, not a failure of
M4 and not a thing to be made green.

**Form B — narrowed to what this branch adds beyond mainline** (merge-base with develop at dispatch,
`5e194bba27c146d8c2157d92b4a3fb3995919ff0`):

```
$ git diff --name-only 5e194bba27c146d8c2157d92b4a3fb3995919ff0..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
(empty)
```

**Second command of the pair, both forms:**

```
$ git status --short -- .moai/specs
(empty)
```

### §4.1 Which form should be the decider — a view, not a decision

Form B should be the decider, and the reason is that the criterion's stated purpose is to catch
**this card** bulk-pinning the corpus. Its mutation is "pin a single occurrence in another SPEC
directory **and commit it**" — an act of this branch. Form A cannot separate that act from any
mainline commit absorbed by a routine merge, so on a branch that merges mainline even once it is
non-empty for reasons the card did not cause, and the criterion stops discriminating: a real
bulk-pin and a clean merge produce the same red. Form B measures exactly the branch's own additions,
turns red on the stated mutation, and stays green on absorption. The two-surface pairing with
`git status --short` is unaffected — it still catches the working-tree half, and Form B still
catches the committed half that `git status` is blind to.

The counter-argument deserves stating: `BASELINE_SHA` is frozen at pre-flight precisely so this
SPEC's own PRESERVE evidence does not rest on a moving ref (`plan.md` §C step 4), and swapping it
for a merge-base makes the comparator a computed value rather than a recorded one. That is a real
cost. It is answerable by recording the merge-base as a second frozen literal at pre-flight, exactly
as `BASELINE_SHA` is — which keeps the anti-moving-ref discipline intact while restoring the
criterion's power to discriminate.

**This is an input to the orchestrator's decision, not a decision taken here.** `acceptance.md` body
content belongs to manager-spec; nothing in it was edited.

## §5. Per-finding classification (107)

Shape codes used in the rows below, each a shorthand for one path through §D.1:

| Shape | Class | Remedy | What it names |
|---|---|---|---|
| `REC` | anchor | R1 | a recorded measurement result |
| `CRIT` | anchor | R2 | a PRESERVE criterion whose value is not knowable at authoring time |
| `DIR` | subject / S1 | R3 | a directive: command plus its pass condition, executed at read time |
| `DOC` | subject / S1 | R3 | the command is the subject matter (doctrine text, code comment, verbatim-preservation assertion) |
| `SELF` | subject / S1 | R3 | this SPEC's own fixture text or a verbatim citation of a grounded instance |
| `R4APP` | subject / S2 | R4 (already applied) | already in R4 form; flagged only because the REQ-MRG-010 exclusion is deferred |
| `FRZ` | already-frozen | none needed | the line IS the R2 freeze instruction |

The `DIR` / `REC` boundary is the judgment most likely to be disputed, and it is stated rather than
hidden: a line reading "run X → expect `0 0`" is a directive whose expected value is a pass
condition and cannot rot (S1); a line reading "X → `0 0` (synced)" records what was observed and its
address must be fixed (anchor). Where a line does both — an evidence table row that is also the
criterion definition — the reading taken here is the one its surrounding section implies. See §6 for
the specific rows where a second reader could reasonably differ.

<!-- Quoted lines below are truncated to 190 characters; the full line is at the cited file:line. -->

**1. `.moai/specs/SPEC-AGENT-PARALLEL-OPT-001/plan.md:92`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| 6 | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | 병렬 세션 레이스 부재 확인 |
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**2. `.moai/specs/SPEC-CC2178-TEAM-API-ALIGN-001/progress.md:19`** — **anchor** · remedy **R1** · shape `REC`

```
- Git baseline: synced (`git rev-list --count --left-right origin/main...HEAD` → `0  0`); HEAD at run-phase entry verified clean.
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**3. `.moai/specs/SPEC-CI-FLAKY-STABILIZE-001/acceptance.md:103`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-CFS-010 | plan §G AP-3 | `git diff --stat origin/main...HEAD` | M1 재현 테스트 파일이 최종 diff에 **부재**. 단 AC-CFS-009의 소스 보존이 선행 충족되어 있어야 함 | SHOULD |
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**4. `.moai/specs/SPEC-CLI-TUX-V3-003/progress.md:174`** — **anchor** · remedy **R1** · shape `REC`

```
Orchestrator Trust-but-verify (2026-07-14, fresh batch; evidence `/tmp/moai-verify-tux3/`): `go test -count=1 -cover ./internal/cli/update/deploy/` → `coverage: 97.5% of statements` (exit 0)
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**5. `.moai/specs/SPEC-CLI-WIZARD-RESTRUCTURE-001/progress.md:424`** — **anchor** · remedy **R1** · shape `REC`

```
  (`git rev-list --count --left-right origin/main...HEAD` → `0 21`, i.e.
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**6. `.moai/specs/SPEC-DESIGN-MOAIWEBV2-002/plan.md:35`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- [ ] `git rev-list --count --left-right origin/main...HEAD` → `0 0` (or local-ahead only) before run-phase spawn
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**7. `.moai/specs/SPEC-DESIGN-MOAIWEBV2-002/plan.md:36`** — **already-frozen** · remedy **none needed** · shape `FRZ`

```
- [ ] Record `BASELINE_SHA=$(git rev-parse origin/main)` BEFORE the first run-phase commit/push — the committed-diff comparator for AC-MWA-007a / AC-MWA-013 (per-milestone pushes must not va
```

The line IS the R2 remedy: it instructs the reader to capture a frozen baseline before the first run-phase commit. It is flagged because the rule's frozen-baseline exclusion matches a variable REFERENCE, not the assignment form that creates the variable.

**8. `.moai/specs/SPEC-DESIGN-MOAIWEBV2-002/progress.md:58`** — **anchor** · remedy **R2** · shape `CRIT`

```
- `git diff --name-only origin/main..HEAD -- docs-site/` = empty, working tree clean (AC-MWA-007a)
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**9. `.moai/specs/SPEC-DOCSITE-E2E-001/progress.md:22`** — **anchor** · remedy **R1** · shape `REC`

```
- **P2 pre-spawn sync**: `git fetch origin main` OK; `git rev-list --count --left-right origin/main...HEAD` → `0 1` (local ahead by 1 = plan commit; proceed). `moai session list --json --fil
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**10. `.moai/specs/SPEC-EVIDENCE-CLAIM-INVARIANT-001/progress.md:97`** — **anchor** · remedy **R1** · shape `REC`

```
- Pre-spawn sync: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` (synced).
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**11. `.moai/specs/SPEC-EVIDENCE-CLAIM-INVARIANT-001/progress.md:138`** — **anchor** · remedy **R1** · shape `REC`

```
  push_precondition: "git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0"
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**12. `.moai/specs/SPEC-FACTORY-MODE-001/progress.md:75`** — **anchor** · remedy **R1** · shape `REC`

```
- `git rev-list --count --left-right origin/main...HEAD` → `0	0`
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**13. `.moai/specs/SPEC-GOAL-DOCS-RETIRE-001/progress.md:126`** — **anchor** · remedy **R1** · shape `REC`

```
| Pre-edit base — `git show origin/main:.moai/docs/autonomous-workflow-strategy.md`, exclusion form | `25` (recorded baseline unchanged) |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**14. `.moai/specs/SPEC-GOAL-DOCS-RETIRE-001/spec.md:28`** — **anchor** · remedy **R1** · shape `REC`

```
| 1.3.0 | 2026-07-27 | `AC-GDR-007`'s content-pin detector corrected to exclude the mandated sentinel line, closing an off-by-one internal contradiction found during run-phase N5 verificatio
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**15. `.moai/specs/SPEC-GOAL-SURFACE-UNIFY-001/progress.md:285`** — **anchor** · remedy **R1** · shape `REC`

```
l44_pre_commit_fetch: "git fetch origin main → rev-list --count --left-right origin/main...HEAD = 2 9 (feature branch behind by 2, ahead by 9); the 2 origin/main commits touch 0 of the 8 fil
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**16. `.moai/specs/SPEC-INTERNAL-ARCH-001/plan.md:30`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- [ ] pre-spawn sync check: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` 또는 `0 N`만 허용
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**17. `.moai/specs/SPEC-INTERNAL-TEST-004/plan.md:67`** — **subject/S1** · remedy **R3** · shape `DIR`

```
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 (or 0 N if local ahead)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**18. `.moai/specs/SPEC-INTERNAL-TEST-004/research.md:15`** — **anchor** · remedy **R1** · shape `REC`

```
| `origin/main...HEAD` | `0 0` (synced) | `git rev-list --count --left-right origin/main...HEAD` |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**19. `.moai/specs/SPEC-KANBAN-WORKTREE-001/research.md:207`** — **subject/S1** · remedy **R3** · shape `DOC`

```
// (gh absent): branch appears in `git branch --merged origin/main`. The
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**20. `.moai/specs/SPEC-KANBAN-WORKTREE-001/spec.md:255`** — **subject/S1** · remedy **R3** · shape `DOC`

```
// (gh absent): branch appears in `git branch --merged origin/main`. The
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**21. `.moai/specs/SPEC-LEAD-DEBOTTLENECK-001/progress.md:121`** — **anchor** · remedy **R1** · shape `REC`

```
l44_pre_commit_fetch: "git fetch origin main; git rev-list --count --left-right origin/main...HEAD -> 0 5 (local ahead only, clean)"
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**22. `.moai/specs/SPEC-LSEL-DRAIN-STALL-001/progress.md:86`** — **subject/S1** · remedy **R3** · shape `DIR`

```
**E4 PRESERVE:** 관측은 M1 커밋 직후 three-dot(`git diff --stat origin/main...HEAD`)로 완료 보고에 기록한다(관측 전 여기에 기입하지 않는다 — VCI §2).
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**23. `.moai/specs/SPEC-LSPMCP-RETIRE-001/plan.md:15`** — **anchor** · remedy **R1** · shape `REC`

```
  (`git rev-list --count --left-right origin/main...HEAD` = `0 0`, no race).
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**24. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:38`** — **subject/S1** · remedy **R3** · shape `SELF`

```
`` | AC-X | `git diff --name-only origin/main -- internal/` | empty (unchanged) | ``,
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**25. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:77`** — **subject/S1** · remedy **R3** · shape `SELF`

```
**Given** a fixture using `git diff --stat origin/main...HEAD -- internal/` to decide "unchanged",
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**26. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:101`** — **subject/S1** · remedy **R3** · shape `SELF`

```
`` `git rev-list --count --left-right origin/main...HEAD` → 0 0 `` with no SHA and no date,
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**27. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:218`** — **subject/S2** · remedy **R4 (already applied)** · shape `R4APP`

```
- verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty)
```

Test 4 — the line asserts the state of a moving thing and instructs the reader to re-measure at read time. R4 is ALREADY APPLIED (command first, value demoted to a dated reference); it is flagged only because the REQ-MRG-010 lint exclusion is deferred (Q0 option C).

**28. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:262`** — **subject/S1** · remedy **R3** · shape `SELF`

```
**CM-1 (positional bypass).** `git diff --name-only origin/main -- internal/ (unchanged)` — mentions
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**29. `.moai/specs/SPEC-MOVING-REF-GUARD-001/acceptance.md:268`** — **subject/S1** · remedy **R3** · shape `SELF`

```
`git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` (no
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**30. `.moai/specs/SPEC-MOVING-REF-GUARD-001/progress.md:113`** — **subject/S1** · remedy **R3** · shape `SELF`

```
  `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → expect `0 0`
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**31. `.moai/specs/SPEC-MOVING-REF-GUARD-001/progress.md:115`** — **subject/S1** · remedy **R3** · shape `SELF`

```
  `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | 병렬 세션 레이스 부재 확인
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**32. `.moai/specs/SPEC-MOVING-REF-GUARD-001/progress.md:158`** — **subject/S2** · remedy **R4 (already applied)** · shape `R4APP`

```
`` - verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty) ``.
```

Test 4 — the line asserts the state of a moving thing and instructs the reader to re-measure at read time. R4 is ALREADY APPLIED (command first, value demoted to a dated reference); it is flagged only because the REQ-MRG-010 lint exclusion is deferred (Q0 option C).

**33. `.moai/specs/SPEC-MOVING-REF-GUARD-001/spec.md:127`** — **subject/S1** · remedy **R3** · shape `SELF`

```
  `git rev-list --count --left-right origin/main...HEAD` is preserved verbatim inside a document.
```

Test 1 — the ref is quoted subject matter: this SPEC's own fixture text, or a verbatim citation of a grounded instance. Test 4 — nothing is measured at read time, so S1.

**34. `.moai/specs/SPEC-MX-SCANNER-DOCS-001/progress.md:65`** — **anchor** · remedy **R2** · shape `CRIT`

```
- `internal/` source tree untouched (AC-MSD-010) — `git diff --name-only origin/main...HEAD` filtered for `internal/` is empty.
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**35. `.moai/specs/SPEC-MX-SCANNER-DOCS-001/progress.md:81`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-MSD-010 | PASS | `internal/` untouched — `git diff --name-only origin/main...HEAD \| grep internal/` empty |
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**36. `.moai/specs/SPEC-NAVIGATOR-SYNC-002/progress.md:411`** — **anchor** · remedy **R2** · shape `CRIT`

```
preserve_list_post_run_count: 2   # internal/navigator/sync/, internal/mx/ byte-unchanged (REQ-NS2-005) — verified via `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigato
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**37. `.moai/specs/SPEC-NAVIGATOR-SYNC-002/progress.md:449`** — **anchor** · remedy **R2** · shape `CRIT`

```
  consumer_only_m0_mx_byte_unchanged: PASS   # git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/sync|mx)/' → exit 1 (0 matches); TestConsumerOnly_M0AndMxByteUnchanged
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**38. `.moai/specs/SPEC-NAVIGATOR-SYNC-005/progress.md:76`** — **anchor** · remedy **R2** · shape `CRIT`

```
- AC-NS5-005a (PRESERVE) — `git diff --name-only origin/main...HEAD | grep -E '^internal/(navigator/(sync|detect|route|tiers)|hook/navigator_detect|mx)/'` → grep exit 1 (no matches = PASS — 
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**39. `.moai/specs/SPEC-PRECOMMIT-PRESERVE-001/progress.md:308`** — **anchor** · remedy **R1** · shape `REC`

```
  hook_body_vs_origin_main: "git diff --stat origin/main HEAD -- internal/template/templates/.git_hooks/pre-commit -> empty (equivalent to cmp rc 0; the literal cmp pipeline was refused by t
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**40. `.moai/specs/SPEC-PROJECT-NAVIGATOR-004/progress.md:24`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-004 (SessionStart hook unchanged) | PASS | `git diff origin/main..HEAD -- .claude/hooks/moai/handle-session-start-navigator.sh internal/template/templates/.claude/hooks/moai/handle-sess
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**41. `.moai/specs/SPEC-SESSION-WORKTREE-001/plan.md:163`** — **subject/S1** · remedy **R3** · shape `DOC`

```
     - **Fallback (gh absent):** `git branch --merged origin/main` → branch in merged list? (squash-merge blind — documented).
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**42. `.moai/specs/SPEC-STAMP-REACHABILITY-001/plan.md:23`** — **subject/S1** · remedy **R3** · shape `DIR`

```
2. `git merge-base --is-ancestor $(jq -r '.commit_sha' .moai/project/codemaps/provenance.json) origin/main` still rc=0 before merging (guard-green baseline preserved through the delivery).
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**43. `.moai/specs/SPEC-STAMP-REACHABILITY-001/progress.md:26`** — **anchor** · remedy **R1** · shape `REC`

```
5. Divergence: `git rev-list --count --left-right origin/main...HEAD` after `git fetch origin main` → `0 6` (local ahead only; clean).
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**44. `.moai/specs/SPEC-STEERING-ALIGN-LOCAL-DIET-001/progress.md:71`** — **anchor** · remedy **R1** · shape `REC`

```
l44_pre_commit_fetch: "git fetch origin main → rev-list --count --left-right origin/main...HEAD → 0 0 (synced, clean pre-spawn)"
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**45. `.moai/specs/SPEC-TEMPLATE-DATE-NEUTRALITY-002/research.md:365`** — **subject/S1** · remedy **R3** · shape `DOC`

```
| **bounded to `$(git merge-base origin/main HEAD)..HEAD`** | empty | `NOT-EVALUABLE` — correct |
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**46. `.moai/specs/SPEC-UPDATE-LEGACY-SKILL-LIST-001/progress.md:10`** — **anchor** · remedy **R1** · shape `REC`

```
- `git rev-list --count --left-right origin/main...HEAD` → `0	0` (synced, no divergence).
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**47. `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/strategy-wave4.md:162`** — **anchor** · remedy **R1** · shape `REC`

```
- `git ls-tree origin/main .github/workflows/` 결과 `llm-panel.yml` 파일 부재 확인 (2026-05-08)
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**48. `.moai/specs/SPEC-V3R3-CLI-TUI-001/progress.md:90`** — **anchor** · remedy **R1** · shape `REC`

```
- Hash match basis: `git rev-list --left-right --count origin/main...HEAD = 0/0` → SPEC files (`spec.md` / `plan.md` / `acceptance.md`) unchanged since M1 entry; recompute would yield same h
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**49. `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/acceptance.md:113`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**Then** Phase D completes with `git status --porcelain` returning empty and `git rev-parse main == git rev-parse origin/main`
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**50. `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/acceptance.md:204`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**Expected behavior**: Phase A precondition check (`git rev-parse --abbrev-ref HEAD == main && git status --porcelain == empty && git rev-parse main == git rev-parse origin/main`) catches th
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**51. `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/plan.md:202`** — **subject/S1** · remedy **R3** · shape `DOC`

```
   > Post-condition: `git status --porcelain` returns empty AND `git rev-parse main` == `git rev-parse origin/main`.
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**52. `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/spec-compact.md:32`** — **subject/S1** · remedy **R3** · shape `DOC`

```
Formalize the **Late-branch workflow pattern**: SPEC commits accumulate directly on `main`; the `feat/SPEC-*` branch is created ONLY at PR creation time via `git switch -c`; PR squash-merges
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**53. `.moai/specs/SPEC-V3R5-LATE-BRANCH-001/spec.md:168`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**REQ-LB-006**: `spec-workflow.md` Step 4 (Cleanup) shall document `git reset --hard origin/main` as the canonical Late-branch closure step after squash merge, including the prerequisite `gi
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**54. `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/acceptance.md:366`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**Then**: diff output non-empty (e.g., 2 lines for the literal change). AC-AFS-012 FAILS. Manager reverts via `git restore --source=origin/main -- internal/hook/pre_tool.go`. Re-verify: AC-A
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**55. `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/plan.md:322`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- AC-AFS-012 verification (`git diff origin/main -- <9 files>` empty) 의무
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**56. `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/progress.md:126`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-AFS-012 | PASS | `git diff origin/main -- <9 out-of-scope files> \| wc -l` | 0 (byte-identical) + 42 preserved occurrences |
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**57. `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/progress.md:192`** — **anchor** · remedy **R2** · shape `CRIT`

```
- 9 out-of-scope Go files (REQ-AFS-010) — `git diff origin/main -- <9 files> | wc -l` returned 0, 42 occurrences preserved (AC-AFS-012)
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**58. `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/spec.md:356`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| R-AFS-006 (NEW iter 2): 9 out-of-scope Go files 실수로 수정 시 cross-SPEC tension | M | H | AC-AFS-012 (git diff origin/main check) 의무. M2 phase에서 grep filtering: `grep -rln '\.claude/agents/moa
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**59. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/acceptance.md:72`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| L44 HARD pre-spawn fetch | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | `0 0` pre-spawn AND `0 0` post-push |
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**60. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md:134`** — **subject/S1** · remedy **R3** · shape `DIR`

```
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**61. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md:338`** — **subject/S1** · remedy **R3** · shape `DIR`

```
   git fetch origin main && git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**62. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/plan.md:358`** — **subject/S1** · remedy **R3** · shape `DIR`

```
   git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 post-push
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**63. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md:90`** — **anchor** · remedy **R1** · shape `REC`

```
| L44 HARD pre-spawn fetch (pre-M1) | `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | `0 0` | `0 0` (verified by orchestrator pre-spawn batch per Section A)
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**64. `.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/progress.md:212`** — **subject/S1** · remedy **R3** · shape `DIR`

```
  2) git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0 (L44 HARD pre-spawn race recheck; verify no other-session activity since this Mx-chore)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**65. `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/acceptance.md:352`** — **subject/S1** · remedy **R3** · shape `DIR`

```
**Verification**: B12-flavored sentinel — `git fetch origin && git rev-list --count --left-right origin/main...HEAD` returns `0 0` after M3 commit.
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**66. `.moai/specs/SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001/plan.md:415`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| Sibling SPEC (COORD-001) 동시 main push race | pre-spawn fetch `git rev-list --count --left-right origin/main...HEAD` 매 M commit 직전 + `0 0` or `0 N` 확인 |
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**67. `.moai/specs/SPEC-V3R6-ASKUSER-DECISION-MEMORY-001/progress.md:215`** — **subject/S1** · remedy **R3** · shape `DIR`

```
E6. Branch HEAD + Push: 신규 commit SHA (본 커밋). push 후 `git rev-list --count --left-right origin/main...HEAD` → `0 0` 예상.
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**68. `.moai/specs/SPEC-V3R6-ASKUSER-DECISION-MEMORY-001/progress.md:281`** — **subject/S1** · remedy **R3** · shape `DIR`

```
**E6. Branch HEAD + Push**: 신규 commit SHA (본 커밋 — M5 preference recovery toggle + sensitive-domain gate + correction loop). push 후 `git rev-list --count --left-right origin/main...HEAD` → `0
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**69. `.moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/progress.md:103`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**Decision**: manager-develop spawn prompt used the Tier S minimal form (~750 tokens covering goal + deliverables + constraints + self-verification) per `.claude/rules/moai/development/manag
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**70. `.moai/specs/SPEC-V3R6-CI-FLAKY-STABILIZE-001/progress.md:24`** — **anchor** · remedy **R1** · shape `REC`

```
| Multi-session race | `git fetch && git rev-list --count --left-right origin/main...HEAD` | `0 5` (local ahead, clean) |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**71. `.moai/specs/SPEC-V3R6-DOCS-RC2-README-001/plan.md:235`** — **subject/S1** · remedy **R3** · shape `DIR`

```
**Mitigation**: The orchestrator MUST run the Pre-Spawn Sync Check (`agent-common-protocol.md` § Pre-Spawn Sync Check) before any M1..M5 delegation. `git fetch origin main && git rev-list --
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**72. `.moai/specs/SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001/progress.md:27`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-DFS-008 | REQ-DFS-008 | preserve | `-run Preserve\|Namespace\|CleanInstall` suites | all green; `git diff --name-only origin/main` excludes update.go / update_clean_install.go / applier
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**73. `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/plan.md:224`** — **subject/S1** · remedy **R3** · shape `DIR`

```
2. `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → MUST be `0 0` clean (Pre-Spawn Sync Check per agent-common-protocol.md)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**74. `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/plan.md:232`** — **subject/S1** · remedy **R3** · shape `DIR`

```
2. `git rev-list --count --left-right origin/main...HEAD` — `0 0` post-push
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**75. `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/progress.md:104`** — **anchor** · remedy **R1** · shape `REC`

```
1. **Pre-spawn fetch**: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` (clean, no divergence)
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**76. `.moai/specs/SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001/progress.md:136`** — **anchor** · remedy **R1** · shape `REC`

```
| V2 | Push divergence | `git rev-list --count --left-right origin/main...HEAD` | `0 0` (clean, no divergence; verified pre-each-commit + post-each-push per L44 HARD discipline) |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**77. `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001/acceptance.md:128`** — **subject/S1** · remedy **R3** · shape `DIR`

```
**Mitigation**: Scopes are disjoint (verified at plan-phase). Pre-spawn `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` MUST return `N 0` or `0 0` before run-
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**78. `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001/progress.md:47`** — **anchor** · remedy **R1** · shape `REC`

```
| Pre-spawn git sync check | PASS | `git rev-list --count --left-right origin/main...HEAD` → `0 0` |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**79. `.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/plan.md:116`** — **subject/S1** · remedy **R3** · shape `DIR`

```
Before spawning manager-develop for run-phase, the orchestrator MUST run `git fetch origin && git rev-list --count --left-right origin/main...HEAD` and verify `0 N` or `0 0` (local clean or 
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**80. `.moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/progress.md:24`** — **anchor** · remedy **R1** · shape `REC`

```
| Multi-session race | `git fetch && git rev-list --count --left-right origin/main...HEAD` | `0 0` (clean) |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**81. `.moai/specs/SPEC-V3R6-MAIN-RED-REMEDIATION-001/progress.md:25`** — **anchor** · remedy **R1** · shape `REC`

```
| Multi-session race | `git fetch && git rev-list --count --left-right origin/main...HEAD` | `0 0` (clean) |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**82. `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md:296`** — **subject/S1** · remedy **R3** · shape `DOC`

```
| AC-COORD-016 | PASS | All 3 sub-verifications return 1 (≥1): `^git fetch origin main` + `^git rev-list --count --left-right origin/main...HEAD` preserved verbatim; awk additive assertion f
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**83. `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md:364`** — **anchor** · remedy **R1** · shape `REC`

```
l44_pre_commit_fetch: 0 0  # pre-commit `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → 0 N where N = M1-M5 bundled commit count (clean ahead, no race)
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**84. `.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/spec.md:34`** — **anchor** · remedy **R1** · shape `REC`

```
**Race signal**: Session C의 pre-spawn `git fetch origin main && git rev-list --count --left-right origin/main...HEAD`은 `0 0` (clean ahead)을 반환했다. 이유는 Session B의 sync commits가 이미 Session C의 l
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**85. `.moai/specs/SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/plan.md:63`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → expect `0 0` (pre-spawn fetch obligation per `agent-common-protocol.md` § Pre-Spawn Sync Check)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**86. `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/plan.md:137`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| R7 | parallel session 이 `internal/session/` production 파일 동시 수정 | Low | pre-spawn fetch `0 0` 의무 (L44 HARD) + run-phase entry 시점 `git rev-list --count --left-right origin/main...HEAD` 재확인.
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**87. `.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/progress.md:287`** — **anchor** · remedy **R1** · shape `REC`

```
| L44 HARD pre-commit fetch | PASS | `git rev-list --count --left-right origin/main...HEAD` = `0 0` at sync-phase entry |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**88. `.moai/specs/SPEC-V3R6-SKILL-GEARS-ALIGN-001/progress.md:38`** — **anchor** · remedy **R1** · shape `REC`

```
Pre-spawn fetch: `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0` clean.
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**89. `.moai/specs/SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001/plan.md:76`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| Concurrent edits from parallel session (L9) | Pre-spawn fetch + post-edit `git rev-list --count --left-right origin/main...HEAD` = `0 0` check before commit. |
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**90. `.moai/specs/SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001/plan.md:30`** — **anchor** · remedy **R1** · shape `REC`

```
- [x] **Race check**: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` (clean sync, no concurrent-session divergence).
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**91. `.moai/specs/SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001/spec.md:51`** — **anchor** · remedy **R1** · shape `REC`

```
| origin/main race | `0 0` (synced) | `git rev-list --count --left-right origin/main...HEAD` |
```

Test 1 — substituting the SHA the ref resolved to leaves the sentence saying what it meant, and still will when read later: the line records what was measured at that address. Test 3's different answer next week is not the point of the claim.

**92. `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001/plan.md:88`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| Parallel session race condition during commit (L9 pattern) | Low | Pre-spawn fetch + `git rev-list --count --left-right origin/main...HEAD` = `0 0` check before commit per `.claude/rules/m
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**93. `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/plan.md:141`** — **subject/S1** · remedy **R3** · shape `DIR`

```
git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**94. `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/plan.md:302`** — **subject/S1** · remedy **R3** · shape `DIR`

```
   git fetch origin main && git rev-list --count --left-right origin/main...HEAD  # expect: 0 0
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**95. `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/plan.md:322`** — **subject/S1** · remedy **R3** · shape `DIR`

```
   git rev-list --count --left-right origin/main...HEAD  # expect: 0 0 post-push
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**96. `.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/progress.md:168`** — **subject/S1** · remedy **R3** · shape `DIR`

```
  2) git fetch origin main && git rev-list --count --left-right origin/main...HEAD → 0 0 (L44 HARD pre-Sprint-8)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**97. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/acceptance.md:35`** — **anchor** · remedy **R2** · shape `CRIT`

```
| 10 | AC-TST-010 | MUST | No predecessor SPEC body modified — `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-00
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**98. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/acceptance.md:37`** — **subject/S1** · remedy **R3** · shape `DIR`

```
| 12 | AC-TST-012 | MUST | Every milestone commit's pre-commit and post-push `git rev-list --count --left-right origin/main...HEAD` returns `0 0` (or documents L52 race absorption with verba
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**99. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/plan.md:239`** — **subject/S1** · remedy **R3** · shape `DIR`

```
3. Pre-commit `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` returns `0 0` (L44 HARD).
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**100. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/progress.md:83`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-TST-010 | PASS | `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**101. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/progress.md:86`** — **anchor** · remedy **R2** · shape `CRIT`

```
| AC-TST-013 | PASS | `git diff origin/main -- internal/ \| grep -E "^\+[ \t]*(t\.Skip\|testing\.Short)"` | empty (no skip/short added); no failing test deleted |
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**102. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/spec.md:113`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- **HARD-4**: Every milestone commit shall pass pre-commit and post-push `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` returning `0 0` (or document L52 race
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**103. `.moai/specs/SPEC-V3R6-TEST-REFACTOR-001/spec.md:133`** — **anchor** · remedy **R2** · shape `CRIT`

```
8. Predecessor SPEC bodies remain unmodified (verified via `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

**104. `.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/progress.md:328`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → expect `0 0` (clean baseline)
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**105. `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/acceptance.md:401`** — **subject/S1** · remedy **R3** · shape `DOC`

```
**Expected behavior**: Per `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check, the orchestrator MUST have executed pre-spawn `git fetch + git rev-list --count --left-r
```

Test 1 — substituting a SHA changes the sentence, because the command is itself the subject matter (doctrine text, a code comment, a verbatim-preservation assertion, or a form comparison). Test 4 — nothing is measured at read time, so S1.

**106. `.moai/specs/SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001/plan.md:261`** — **subject/S1** · remedy **R3** · shape `DIR`

```
- **Anti-pattern (cohort cross-attribution)**: Cross-commit contamination from parallel Sprint 10 cohort sessions. Pre-spawn fetch `git fetch origin && git rev-list --count --left-right orig
```

Test 3 — a re-run is expected to differ and that variance IS the point (the check exists to detect drift), so SUBJECT. Test 4 — the line asserts no current state; it instructs a reader to measure and states the pass condition. Nothing in it rots, so S1.

**107. `.moai/specs/SPEC-WEBCONF-SIMPLIFY-001/acceptance.md:203`** — **anchor** · remedy **R2** · shape `CRIT`

```
- **AC-WC-009 (doctrine preserved)**: `git diff --name-only origin/main...HEAD -- .claude/rules/moai/workflow/mx-tag-protocol.md .claude/rules/moai/workflow/context-window-management.md .cla
```

Test 1 — the claim is a PRESERVE criterion decided against mainline, and substituting a fixed baseline preserves its meaning. Test 3 — a re-run must give the same answer, so the value must be fixed; it is not knowable when the criterion is written, hence R2 rather than R1.

## §6. Gaps — what was explicitly NOT observed

- **M5 (template mirror) was not started.** AC-MRG-012 is entirely unobserved by this milestone.
- **REQ-MRG-010 / AC-MRG-013 remain unimplemented and untested** (Q0 option C). CM-1 and CM-2 were
  not run. Findings 27 and 32 are the deferral's live consequence and no claim is made here about
  whether an exclusion would classify them correctly.
- **Every classification in §5 is one reader's judgment.** This is limit L2 acting on the triage
  rather than on the detector: the rule reads shape and never subject, so the predicate was applied
  by hand, by the same actor that ran the rule, with no independent adjudication. Limit L7 sharpens
  it for the anchor rows specifically — the ANCHOR branch of the predicate has zero adjudicated
  instances, so the 48 anchor verdicts here should be weighted less confidently than the subject
  ones, and this milestone adds 48 author-classified anchors to the seven §F L7 already counts.
- **Rows where a second reader could reasonably reach a different class**, named rather than
  smoothed over. None was forced into a bucket to avoid saying so:
  - **11** (`SPEC-EVIDENCE-CLAIM-INVARIANT-001/progress.md:138`) — a `push_precondition:` field in a
    `verification:` block. Read as a recorded check (anchor) because its sibling fields record
    results; read as a precondition definition it is `DIR`/S1.
  - **80, 81, 91** — risk/verification table rows of the form `| … | command | \`0 0\` (clean) |`.
    Read as recorded results; a reader who treats the third column as the pass condition gets S1.
  - **65, 98, 103** — acceptance criteria phrased as a standing obligation ("every milestone commit
    returns `0 0`", "returns empty after the 4-phase close"). 65 and 98 were read as directives,
    103 as an anchor criterion, on the grounds that 103 names a fixed terminal state and the other
    two recur per commit. That line is defensible but not forced.
  - **22, 42, 67, 68** — statements of what will be measured or is expected ("→ `0 0` 예상",
    "still rc=0 before merging"), read as directives because they assert no value that rots.
- **No claim is made that the rule's 107 is the complete set of moving-ref defects.** Limits L1
  (refs without an `origin/` token), L3 (line-scoped detection), and L5 (no carrier outside
  `.moai/specs`) are unchanged and unmeasured here. In particular, `.moai/reports/**` — where the
  grounded instance that motivated R4 lives — was not scanned.
- **The rule was not run against any tree other than this worktree at HEAD `1ed8a9997`.** The
  figures are not claimed to hold on `origin/develop` or on any other branch.
- **No Go source file was modified, no test was run, and no build was performed** for this
  milestone. `go run ./cmd/moai spec lint` compiled the tree as a side effect (rc 0), which is the
  only compilation evidence this milestone produces.

## §7. Residual risk

- **The classification could be internally consistent and still wrong.** The shape buckets were
  derived from the corpus by the same reader who then applied them, so a misreading of §D.1 would
  reproduce itself uniformly across 107 rows rather than showing up as an inconsistency. The
  external S2 count of 0 is the one figure with an independent check against it (§B.7's two probes,
  measured before this milestone existed), and it agrees.
- **The `DIR` bucket is large (35 external rows) and carries the most doctrinal weight.** If
  directives are in fact S2 rather than S1 — that is, if "a reader will act on it" is read to
  include executing a procedure step — then 35 external findings change remedy from R3 to R4, and
  §B.7's "empty scope" conclusion changes with them. The reading taken here is the narrower one
  (§D.1 Test 4 asks whether *the claim asserts the current state of a moving thing*, which a
  directive does not), and it is the reading that keeps §B.7's measurement true. A reviewer who
  disagrees should expect the R4 scope question to reopen.
- **107 findings is a large enough wall to provoke exactly the response this SPEC exists to
  prevent.** Severity is `warning` and the message names four remedies, but 56 of the 107 are
  subject-class lines that are correct as written and will keep firing until someone writes 56
  markers. The incentive to write one blanket suppression instead is real, and nothing in the
  mechanism prevents it — only the reason field's cost does (§D.4), and that price is enforced by
  review rather than by the detector (L6's second residual).
- **The finding-count decomposition in §2 rests on grep re-derivations of the rule's own branches,**
  not on instrumented rule output. The rule does not report which conjunct fired, so 41 / 61 / 56
  were reconstructed by applying the rule's published patterns to the flagged lines. A divergence
  between the reconstruction and the rule's actual internal path would not be visible here.
