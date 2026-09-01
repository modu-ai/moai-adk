# Sync Audit — SPEC-STATUS-TRANSITION-VALIDITY-001 (card t376)

**FINAL.** Independent, read-only post-implementation audit.

| | |
|---|---|
| Auditor | `sync-auditor` |
| Tree | `.claude/worktrees/t376`, branch `WT-status-transition-gap` |
| HEAD audited | `ff8a7dcba` (base `a5d963db6`) |
| Measurement binary | `make build` → `strings bin/moai \| grep -c ff8a7dcba` → **4** |
| Overall verdict | **PASS** |
| Blocking findings | **0** |
| Optional findings | **4** (F1, F2, F3, F4) |

Every number below was measured by this auditor in this tree, with the command
cited. Nothing was carried over from `progress.md`; where a run-phase figure is
reproduced it is because the same command produced the same output here.

---

## Dimension Scores

Flat weighted-percentage mode (`harness.yaml` sets no `evaluator_mode: hierarchical`).
Must-pass firewall: Functionality + Security, each independently.

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 93/100 | **PASS** (must-pass) | `go test ./internal/spec/ -run 'TestStatusTransition\|TestIsCanonicalStatusTransition\|TestApplyEraDemotion\|TestDemotionCause\|TestPhaseValueShape' -count=1 -v` → 79 `--- PASS`, 0 `--- FAIL`. Full package: `go test ./internal/spec/... -count=1` → `ok github.com/modu-ai/moai-adk/internal/spec 61.675s`. Corpus: `./bin/moai spec lint --json \| jq length` → `1200`. Four independent mutants killed (below). |
| Security (25%) | 95/100 | **PASS** (must-pass) | `golangci-lint run ./internal/spec/...` → `0 issues.`; `go vet ./internal/spec/...` → rc 0. No new `exec` surface (the rule reuses `getOwnershipTransitionRunner`; the existing `git log` invocation passes the path after `--`). No secret, credential, or network surface. Observation-only verified: `git status --porcelain internal/ .moai/specs/` → empty after a full-corpus run. |
| Craft (20%) | 86/100 | PASS | Test set is bidirectional in one execution with a live control, and survives mutation (below). Deductions: F1 (unmeasured, unrecorded doubling of the per-document git walk) and F3 (a code comment citing a requirement that does not say what is cited). |
| Consistency (15%) | 91/100 | PASS | Module conventions honored (`internal/spec/CLAUDE.md`): table-driven subtests, a closed-sibling no-false-positive case (`closed_sibling_spec_is_not_false_flagged`), observation-only rule, stable `Code()` string, `Line` 1-indexed. Deduction: F2 (the §C limitation statement is incomplete on its composition). |

**Weighted total: 91.2/100.** Both must-pass dimensions clear independently.

---

## Anti-vacuity evidence — mutation probes

The run's own RED evidence (`red-evidence.md`) establishes only that the rule was
absent. It does not establish that the **silent-side** assertions can fail, which
is the half that a vacuous test set hides in. I therefore mutated the shipped rule
and re-ran. Each mutant was applied to `internal/spec/lint_transition.go`, measured,
then reverted; `shasum` before and after is `5ab6e89994500d94a92a229a1d0e01d3828c1c6a`
in both cases and `git status --porcelain internal/` is empty at close.

| Mutant | Change | Result | What it proves |
|---|---|---|---|
| M1 | `isCanonicalStatusTransition` returns `true` unconditionally | **KILLED** — 4 must-fire subtests + `TestStatusTransitionTrailerIndependence` + `TestStatusTransitionFindingGates` + `TestStatusTransitionRuleIsObservationOnly` + 6 `TestIsCanonicalStatusTransition` subtests fail | the fire-side assertions are live end-to-end, not only at the unit seam |
| M2a | drop `{"in-progress","completed"}` from `canonicalStatusTransitions` | **KILLED** — `in_progress_to_completed_single_sync_close_passes` fails, and only it | the **must-stay-silent** assertions are live; a silent case is not silent because nothing runs |
| M2b | remove the `(none)` skip in `Check` | **KILLED** — `none_to_completed_is_skipped`, `none_to_draft_is_skipped`, `TestStatusTransitionCheckOrder/none_skip_precedes_token_check` fail | the 136-record `(none)` skip is pinned, not incidental |
| M3 | set `Advisory: true` at the `StatusTransitionInvalid` emission site | **KILLED** — `TestStatusTransitionFindingGates` fails on both assertions (`finding is advisory` + `HasErrors() is false under --strict`) | REQ-STV-009's "no emission-site Advisory flag" is genuinely pinned; the gating property is a tested fact, not a claim |

M2a/M2b together are the load-bearing result: they are the evidence the RED run
could not produce, and they answer the question the RED note itself raised.

---

## AC Matrix — this auditor's judgment

`=` means my verdict matches the run's. Divergences are called out.

| AC | Run | Mine | Basis |
|---|---|---|---|
| 001 `draft → completed` caught | PASS | = PASS | subtest passes; M1 kills it |
| 002 `completed → draft` caught | PASS | = PASS | subtest passes; M1 kills it |
| 003 trailer-independence | PASS | = PASS | `TestStatusTransitionTrailerIndependence` passes; M1 kills it. The normalize-then-compare shape genuinely tests message identity |
| 004 `implemented → completed` silent | PASS | = PASS | subtest passes; silent side proven live by M2a |
| 005 `draft → in-progress` silent | PASS | = PASS | same |
| 006 `in-progress → implemented` silent | PASS | = PASS | same |
| 007 `completed → in-progress` amendment silent | PASS | = PASS | same |
| 007a single-sync close + `draft → implemented` | PASS | = PASS | **M2a specifically kills the single-sync case** — the strongest silent-side evidence in the set |
| 008 `planned` tolerated | PASS | = PASS | both sides covered; corpus shows no `planned` edge among the 97 |
| 009 no transition / no git silent | PASS | = PASS | three cases incl. the injected-lookup `git_unreachable_is_silent`; no error-severity emission asserted per case |
| 010 live control in same execution | PASS | = PASS | `firedAtLeastOnce` is set inside sequential subtests and checked after the loop; the RED log line 405 shows it firing as designed |
| 011 demotion annotation names its cause | PASS | = PASS **and independently confirmed on the live corpus** | `SPEC-V3R6-LINK-FIX-001` (frontmatter `era: V3R6`, `status: completed`) carries `[terminal lifecycle status — downgraded to warning]` — measured, not asserted. The unit test's "the two annotations must actually DIFFER" cross-check is the correct anti-vacuity guard |
| 012 observation-only | PASS | = PASS | `git status --porcelain internal/ .moai/specs/` empty after a full-corpus run. Diff surface: exactly 5 files, all under `internal/spec/`. `grep` over the diff for `var terminalStatusEnum` / `var eraDemotableCodes` / `StatusGitConsistencyRule` → rc 1 (**no definition touched**). The `lint.go:239` decision keeps both disjuncts and the same OR |
| 013 per-code baseline re-measurement | PASS | = PASS | I re-derived both sides. Baseline `lint-baseline-merged.json` = 1096 and matches the AC's ten listed per-code figures **exactly**. After = 1200. Every pre-existing code Δ = 0 |
| 014 `completed → implemented` reversal caught | PASS | = PASS | subtest passes; M1 kills it; 48 on the corpus |
| 015 unrecognized token fires its own code | PASS | = PASS | 4 token cases; the load-bearing "and no `StatusTransitionInvalid`" half is asserted per case |
| 016 non-overlap, measured | PASS (caveated) | **PASS-WITH-DEBT** | see § Weakness 1 — the AC passes as written and the caveat is correctly stated, but one of the two supporting evidence legs is a mis-citation (F4) and nothing pins the property |
| 017 `(none) → X` skipped | PASS | = PASS | **M2b kills it** |
| 018 the finding actually gates | PASS | = PASS | **M3 kills it** — this is now a tested property rather than a demonstrated one |
| 019 gating population measured and decided | PASS — nothing owed | = PASS | `jq '[.[] \| select(.advisory != true)] \| length'` → **0** overall, **0** per new code, on my own run. Baseline was also 0. Under the AC's terms a decision is owed only on a non-zero count |

**20/20 PASS**, one of them with debt. No AC is marked PASS on reasoning where I
could not reproduce the cited output.

### Per-code table, re-measured here

`./bin/moai spec lint --json | jq -r 'group_by(.code)|map({code:.[0].code,n:length})|sort_by(-.n)|.[]'`

| Code | Baseline | This audit | Δ |
|---|---:|---:|---:|
| CoverageIncomplete | 846 | 846 | 0 |
| MovingRefUnpinned | 114 | 114 | 0 |
| **StatusTransitionInvalid** | — | **97** | +97 |
| LegacyEARSKeyword | 43 | 43 | 0 |
| ModalityMalformed | 25 | 25 | 0 |
| MissingExclusions | 24 | 24 | 0 |
| StatusGitConsistency | 18 | 18 | 0 |
| FrontmatterInvalid | 14 | 14 | 0 |
| **StatusTokenUnrecognized** | — | **7** | +7 |
| InvalidREQID | 6 | 6 | 0 |
| SyncSHASlotFormat | 5 | 5 | 0 |
| OwnershipTransitionInvalid | 1 | 1 | 0 |
| **Total** | **1096** | **1200** | **+104** |

---

## Ruling — Weakness 1: is `AC-STV-016` vacuous?

**Ruling: acceptable and correctly recorded — with one mis-citation and one
unpinned property (F4). Not a reproduction of plan-audit iter-1's D4 defect.**

The distinction that decides it: D4's defect was an AC that **could not fail**
while being reported as if it had been tested. What landed is an AC that **passes
on a measurement whose discriminating power is limited**, reported with that
limitation stated in the run's own words before anyone asked. Those are different
failures, and only the first is the one this card exists to close.

Three checks:

1. **Is the second operand structurally empty (a dead rule)?** No.
   `StatusValueEnumRule` is registered at `internal/spec/lint.go:135` and emits
   `StatusValueInvalid` at `lint.go:1173`. The 0 is a measured-clean corpus, not an
   unwired rule. Had it been unwired, the AC would have been unconditionally
   vacuous and I would be reporting a blocking defect.

2. **Does the named substantive evidence carry the weight?** *Leg one does; leg two
   does not.*
   - **Leg (a) HOLDS, verified independently.** All 7 `StatusTokenUnrecognized`
     documents carry a valid current frontmatter status — measured:
     `completed` ×5 (`SPEC-CONFIG-001`, `SPEC-HOOK-001`, `SPEC-SDD-001`, `SPEC-UI-001`,
     `SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001`), `superseded` (`SPEC-LSP-001`),
     `rejected` (`SPEC-WEB-CONSOLE-CONFIG-DIET-001`). That is a **per-document**
     reason each one cannot appear in the second set, not an appeal to the empty
     operand. It genuinely discharges REQ-STV-015 on this corpus.
   - **Leg (b) does NOT support the claim — F4.** `progress.md` §E.2 cites
     "the paired assertion in `AC-STV-015`'s test cases fails if a token case ever
     also emits `StatusTransitionInvalid`". That pins non-overlap between the two
     **new** codes. `AC-STV-016` and REQ-STV-015 are about non-overlap between
     `StatusTokenUnrecognized` and **`StatusValueInvalid`** — a different pair.
     The citation inflates the apparent evidence base; strike it and leg (a) still
     stands on its own.

3. **Is the property pinned against regression?** No, and this is the residual gap.
   The two rules are **not structurally disjoint**: a SPEC carrying an invalid
   frontmatter status *and* an unrecognized token in its history would appear in
   both sets and break the invariant `AC-STV-016` asserts. No fixture constructs
   that case. So the AC's guarantee is corpus-conditional and will stop holding
   silently — the same shape, one level down, that the card is about.

**Net:** the fix did not reproduce D4's defect. It left a weaker version of the
same hazard one level down, which the run half-saw (it named the empty operand)
and half-missed (it did not notice its second evidence leg was about a different
pair, nor that nothing pins the property). Optional, not blocking: the corpus
evidence is real and the AC as written was satisfied.

---

## Ruling — Weakness 2: the rule "gates nothing"

**Ruling: the card's central claim is honest. §C's framing is inadequate on
composition and permanence — F2. The time-index is needed.**

**On honesty.** The claim under audit is that a silently-passing detector gap is
closed. It is:

- 104 facts that were previously invisible are now emitted, machine-readable, and
  attributable to a commit SHA. `spec.md` §C states plainly that the finding "is
  visible in `--json` output and in the per-code baseline; **it is not a gate**."
  No reader is led to believe CI now blocks.
- The rule's ability to gate is not merely demonstrated — it is **pinned**. Mutant
  M3 (setting the emission-site `Advisory` flag) is killed by
  `TestStatusTransitionFindingGates` on both assertions. "Can gate but currently
  does not" is therefore a verified property, not a reassurance.
- `progress.md` §E.2 explicitly refuses to read the 0 as a clean bill: "the gate is
  live and the corpus is simply all sheltered." That is the correct reading and it
  was volunteered.

**On adequacy — where §C falls short (F2).** §C's "Stated, accepted limitation"
paragraph names exactly one shelter — the terminal frontmatter status — and offers
`draft → completed` as its worked example. I measured the actual composition of the
97:

```
$ jq -r '...select(.code=="StatusTransitionInvalid")|.file' ... | sort -u   → 97 files
$ tr '\n' '\0' < files | xargs -0 grep -h -m1 -E "^status:" | sort | uniq -c
  49 status: completed
  48 status: implemented
```

- **49** end at `completed` — sheltered by the terminal-status disjunct, exactly as
  §C describes.
- **48** (the whole `completed → implemented` reversal class) end at `implemented`,
  which is **not** in `terminalStatusEnum`. Their `Advisory: true` is measured
  (`jq '[.[]|select(.code=="StatusTransitionInvalid" and .advisory==true)]|length'`
  → 97), and since `demoted() = GrandfatheredEra || TerminalStatus` and
  `TerminalStatus` is false for them, they are sheltered **solely by the
  grandfathered-era exemption**. *(Labelled: this last step is an inference from
  the code I read at `lint.go:245` plus the measured advisory flag, not a direct
  observation of the era classifier.)*

That is nearly half the population, sheltered by a mechanism §C never names in the
limitation paragraph. The consequence is not just incompleteness — it changes how
the limitation reads over time. §C presents it as a structural consequence of the
lifecycle (`completed` shelters itself, permanently true). For 48 of 97 the shelter
is instead a **policy exemption currently being narrowed by another lane**.

**On the time-index: yes, it is needed, and it needs to carry the split.** Card
t382 (`SPEC-ERA-H3-NARROWING-001`) reports 22 of 23 era-exempt rows unjustified. If
it lands, some or all of those 48 stop being advisory, and `spec-lint --strict` —
which `.github/workflows/spec-lint.yml:40` runs on every push to `main` and
`develop` — reddens the integration branch. `AC-STV-019` was written precisely to
force a knowing decision "at the point where the number exists"; the number is 0
today only because of a shelter that is being removed. A time-index reading only
"measured at `ff8a7dcba`" would understate this. It should record the **49/48
split** and name the era half as the population t382 puts in play, so the next
reader can size the exposure without re-deriving it.

I accept the lead's statement that the time-index was deferred to after the audit;
its absence is not treated as concealment.

---

## Also examined

### The 97-vs-98 projection miss — VERIFIED, not merely plausible

The run's explanation is a hypothesis until checked. I checked all four legs:

| Leg | Command | Observed |
|---|---|---|
| the history really records the edge | `git log --follow -p --format='COMMIT %H' -- .moai/specs/SPEC-V3R6-LINK-FIX-001/spec.md \| grep -E '^COMMIT \|^[+-]status:'` | `COMMIT 3b793f00d…` / `-status: draft` / `+status: completed` |
| the frontmatter uses the rejected alias | `head -14 …/spec.md` | `spec_id: SPEC-V3R6-LINK-FIX-001` (line 2) — no `id:` key |
| the rule skips it | `grep -c SPEC-V3R6-LINK-FIX-001` over the 97 flagged files | `0` |
| the fact is not lost | `jq` over the corpus for that file | `FrontmatterInvalid  Frontmatter required field missing: id` present |

Arithmetic closes: 49 observed + 1 skipped = the census's 50 `draft → completed`.
The guard doing the skipping is `fm.ID == ""` at `lint_transition.go:114`, and the
behavior is pinned by
`TestStatusTransitionRuleGuards/unreadable_frontmatter_id_defers_to_frontmatter_rule`.
Bonus: this same document is the live-corpus confirmation of `AC-STV-011`.

### The recorded deviation (one rule, two codes) — justification HOLDS

- **Memoization claim: true.** `lookupOwnershipTransitionFromGit`
  (`lint_ownership.go:202`) spawns `git log --follow -p` on every call with no
  result cache. `gitquery_cache.go` exposes only `cachedGitDirAvailable` and
  `cachedMainBranch` — the `git rev-parse` environment probes. So two registered
  rules would indeed have meant an additional walk per document.
- **Precedent claim: true.** `BreakingChangeIDRule.Code()` returns
  `BreakingChangeMissingID` (`lint.go:1124`) while the body also emits `OrphanBCID`
  (`lint.go:1145`).
- **Recorded properly:** yes — named as a deviation, with the plan clause it
  departs from, the reason, the precedent, and why the AC set is unaffected.
- **But the reasoning is anchored to the wrong baseline — see F1.**

### Bidirectionality actually exercised — CONFIRMED

Fire and silence are asserted in **one** execution of
`TestStatusTransitionValidityRule` (24 cases: 4 must-fire, 4 token-fire, 16
must-be-silent), with `firedAtLeastOnce` checked after the loop. Subtests run
sequentially (no `t.Parallel()`), so the flag is set before it is read. M1 kills
the fire side and M2a/M2b kill the silence side — so this is a suite that can
distinguish a working rule from a dead one **in both directions**, which is the
property `AC-STV-010` was written to buy.

### Non-goals respected — CONFIRMED

`git diff --name-only a5d963db6..ff8a7dcba -- internal/` → exactly the 5 files
`progress.md` names, all under `internal/spec/`. A grep of the diff for the
definitions of `terminalStatusEnum` / `eraDemotableCodes` and for
`StatusGitConsistencyRule` returns rc 1 — none touched. The `lint.go:239` demotion
decision was restructured from `demote := A || B` into `cause := {A, B}` with
`demoted() = A || B`; semantically identical, and permitted by `AC-STV-012`'s
"other than the message text required by AC-STV-011". The demotion-message
misattribution fix — in scope — landed and is tested at the unit seam and
end-to-end.

### RED evidence integrity

`red-evidence.md` transcribes the RED runs because `*.log` is gitignored. I
spot-checked the transcription against the surviving `red-m1.log`: `head -4`
matches byte-for-byte, and the `AC-STV-010 live control` line is present verbatim.
The transcription is faithful. Note that the M4 RED is a **compile failure**
(`undefined: demotionCause`), which is a weaker RED than a failing assertion — it
does not establish the demotion assertions can fail. That gap is covered instead by
the explicit `annotations[0] == annotations[1]` cross-check in
`TestApplyEraDemotionNamesItsCause` and by my live-corpus confirmation on
`SPEC-V3R6-LINK-FIX-001`, so I raise no finding.

---

## Findings

Format: `Fn [severity] [blocking|optional] file:line — description — Required fix`.
Per the finding-consumption discipline, **blocking** means correctness or a stated
requirement; everything else is discretionary and is not auto-routed into fixes.

### F1 [MEDIUM] [optional] — the card doubled the per-document `git log` walk, which the plan named as the thing to avoid

`internal/spec/lint_transition.go:118` and `internal/spec/lint_ownership.go:387`
both call `getOwnershipTransitionRunner`, which is `lookupOwnershipTransitionFromGit`
(`lint_ownership.go:192,202`) — unmemoized, spawning `git log --follow -p` per call.
Before this card the corpus lint performed **one** such walk per document; after it,
**two**.

`plan.md:22-24` states: *"Adding a second independent `git log` walk per document
would double the run cost; reusing the existing lookup is the intended shape."*
"Reusing the existing lookup" was implemented as reusing the *function*, not the
*result* — so the doubling the clause exists to prevent happened anyway.

The recorded deviation reasons from the wrong baseline. Its comparison is
"one rule = one walk vs. two rules = two walks"; the actual comparison is
"one rule = **two** walks vs. two rules = **three**". The deviation avoided a third
walk, not a second, and no shape available to it met plan.md's stated goal —
only memoizing the lookup does, and that was neither done nor recorded as debt.

Confidence: high (read directly from the two call sites and the cache's exposed
surface). Magnitude, measured here: `time ./bin/moai spec lint --json > /dev/null`
→ **7m16.9s wall** on this tree. The before/after delta is **UNMEASURED** — I did
not build the base binary — so I state the doubling structurally and the wall-clock
comparison as a Gap.

**Time-indexed risk.** `.github/workflows/spec-lint.yml` uses `actions/checkout@v7`
with **no `fetch-depth`**, i.e. a shallow depth-1 clone, so `git log --follow`
returns almost nothing and the doubling costs little on CI today. Card **t371** —
which `spec.md` §C names — proposes `fetch-depth: 0`. That job carries
`timeout-minutes: 10` (line 28), and my local run of the doubled version took 7m17s.
The two cards are coupled in the direction §C does not record: §C says t371 does not
close this card; it does not say this card makes t371 more expensive.

**Required fix (choose one):** memoize `getOwnershipTransitionRunner` per `Lint()`
run in `gitquery_cache.go` (removes the doubling and the pre-existing single walk's
duplication in one move); **or** record the doubling in `progress.md` §E.2 as
explicit debt with an `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` marker naming the
t371 `fetch-depth: 0` coupling as the upgrade trigger.

### F2 [MEDIUM] [optional] — `spec.md` §C states the gating limitation incompletely, and in a form that does not survive t382

`.moai/specs/SPEC-STATUS-TRANSITION-VALIDITY-001/spec.md` §C, "Out of Scope — the
era / terminal-status demotion path", "Stated, accepted limitation" paragraph.

The paragraph names only the terminal-status shelter and illustrates it with
`draft → completed`. Measured composition of the 97: **49** sheltered by terminal
status, **48** sheltered solely by the grandfathered-era exemption (current
frontmatter status `implemented`, which is not in `terminalStatusEnum`). Nearly half
the population is sheltered by a mechanism the limitation paragraph does not
mention, and that mechanism is a policy exemption t382 is actively narrowing —
not a permanent lifecycle property. As written, §C reads as permanent.

`progress.md` §E.2 gets the mechanism right ("end in `completed` **or** sit on
grandfathered-era directories") but gives no split, so a reader cannot size how
much of the reported 0 is era-dependent.

**Required fix:** amend §C's limitation paragraph to name both shelters, and record
the 49/48 split in `progress.md` §E.2 with a sentence stating that the era-sheltered
48 are the population that begins gating if the era exemption narrows. Fold this
into the time-index the lead intends to add rather than writing a separate note.

### F3 [LOW] [optional] — `lint_transition.go` cites REQ-STV-012 for a claim REQ-STV-012 does not make, and the claim is wrong

`internal/spec/lint_transition.go:36-37`:

```go
// Observation-only: never mutates a SPEC, and reuses the existing git lookup
// rather than adding a second history walk (REQ-STV-012).
```

REQ-STV-012 (`spec.md:315-320`) is the Unwanted/observation-only requirement — it
says the rule shall not mutate any SPEC file and shall not alter `terminalStatusEnum`,
`eraDemotableCodes`, the `StatusGitConsistencyRule` early return, or the `lint.go:239`
demotion decision. It says nothing about history walks. The walk clause belongs to
`plan.md` §B, and per F1 the claim itself is false: a second history walk per
document *is* added.

**Required fix:** drop the "rather than adding a second history walk" clause (or
correct it per F1's outcome) and keep the REQ-STV-012 citation attached only to the
observation-only sentence it actually supports.

### F4 [MEDIUM] [optional] — `AC-STV-016`'s second evidence leg is a mis-citation, and the non-overlap is unpinned

`progress.md` §E.2, "AC-STV-016 — the intersection, and what it does and does not
establish", final paragraph.

Two problems, detailed under § Weakness 1:

1. The cited `AC-STV-015` paired assertion pins non-overlap between
   `StatusTokenUnrecognized` and `StatusTransitionInvalid` — the two **new** codes.
   `AC-STV-016`/REQ-STV-015 concern non-overlap with **`StatusValueInvalid`**. Wrong
   pair; it inflates the evidence base for a claim it does not touch.
2. No test pins the property. The rules are not structurally disjoint: a SPEC with an
   invalid frontmatter status *and* an unrecognized history token lands in both sets
   and breaks the invariant, and no fixture constructs that case. The guarantee is
   corpus-conditional and will lapse silently.

**Required fix:** strike the second leg from §E.2 (leg (a) stands alone and is
sound); add a fixture writing an out-of-enum frontmatter status together with an
out-of-enum history token, asserting whichever outcome REQ-STV-015 intends — either
that only one code fires, or that both are correct and `AC-STV-016` should be
amended to name that bounded set, which the AC's own text already prescribes as the
second remedy.

---

## Gaps — what this audit did NOT observe

- **No before/after wall-clock measurement of the git-walk doubling.** I did not
  build a binary at `a5d963db6`; F1's magnitude is a structural inference plus a
  single post-change timing (7m16.9s). The delta is unquantified.
- **No CI-side measurement.** All numbers here are local, on a machine with other
  lanes active. The CI verdict on `ff8a7dcba` is not read; per project convention
  the full-suite and cross-platform verdict is CI's.
- **No full-suite run.** Scoped to `./internal/spec/...` as instructed.
- **Era classification of the 48 not directly observed.** The "sheltered solely by
  the era exemption" step is an inference from `demoted() = A || B` plus the measured
  `Advisory: true` and non-terminal status, not a direct read of `ClassifyEra`
  output for those 48 documents.
- **Plan-audit optional defects D5, D6, D8, D9 not re-examined.** They were left open
  at plan-audit iter-2 (PASS 0.96) and are outside a sync-phase audit's remit.
- **`OwnershipTransitionRule`'s own behavior not re-audited.** Its verdict count is
  unmoved (1 → 1), which is all this card claims about it.

## Residual risk

- **The 0 gating population is a moving number, not a property.** It rests on the
  era exemption for 48 of 97. The card is correct today and its correctness has a
  dependency the SPEC does not name (F2). A reader who consults §C after t382 lands
  will be reading a limitation statement that no longer describes the system.
- **`AC-STV-016` will lapse without announcing it** (F4). The first SPEC that carries
  both an invalid frontmatter status and an unrecognized history token silently
  violates an invariant the card recorded as measured.
- **The git-walk doubling is latent, not benign** (F1). It is nearly free under the
  current shallow CI checkout and becomes material the moment `fetch-depth: 0` lands
  against a 10-minute job timeout.
- **`run_commit_sha: pending-backfill-final-run-commit`** in `progress.md` §E.3 is
  still a placeholder. That is the sanctioned self-reference workaround, but the
  backfill is outstanding and is a sync-phase obligation.

---

## Verdict

**PASS.** Both must-pass dimensions clear independently; there are no blocking
findings. Weighted score 91.2/100.

The card ships a detector that works, is tested in both directions in a single
execution, survives four independent mutations of its own logic, and reports its
own limitations before being asked — including the one (`AC-STV-016`'s empty
operand) that a less careful run would have banked as a clean measurement. The two
self-declared weaknesses are correctly recorded rather than concealed; my
disagreement with the run is narrow and lands on framing and completeness, not on
correctness.

The four findings are optional. F1 and F2 are the two worth acting on before the
card closes, both because they are cheap and because both are coupled to other
lanes already in flight (t371, t382) — the point at which they stop being cheap is
the point at which those cards land.

---

*Auditor note: this audit ran read-only against the implementation. Four temporary
mutations were applied to `internal/spec/lint_transition.go` as measurement, each
reverted immediately; the file's `shasum` is unchanged
(`5ab6e89994500d94a92a229a1d0e01d3828c1c6a`) and `git status --porcelain internal/`
is empty. Nothing was committed, pushed, or merged. New artifacts written by this
audit: this file and `.moai/reports/t376/lint-sync-audit.json`.*
