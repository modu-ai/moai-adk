# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 4 (cap extended past 3/3 by dispatcher authorisation; annotation-cycle ceiling 6)
Verdict: **FAIL** — one blocking finding, on an aggregate that clears the threshold
Overall Score: **0.87** (harmonic mean; Tier L threshold 0.85, read from
`.claude/rules/moai/workflow/spec-workflow.md:142`)
Signal: **no STOP** — 0.82 → 0.87, no score regression.

Audited tree: `.claude/worktrees/t82` (cwd re-anchored on every call), commit **`d378e5358`**.
Document version **`0.3.3`**, read from `spec.md:4` frontmatter — not from a commit message.
`git status --short .moai/specs` → empty at audit start and at audit end.

Reasoning context ignored per M1 Context Isolation.

---

## Verdict in one paragraph

All four blocking defects from iteration 3, and the optional D5, are **resolved** — not
cosmetically: D1's amended `AC-AMC-019` kills the counterexample outright, and D2's ordering
assertion is co-temporal rather than after-the-fact, which is the stronger of the two shapes the
dispatch asked me to distinguish. The sixteen prior findings all hold under spot re-execution.
The aggregate rises to 0.87, above the Tier L threshold.

It still fails, on one finding, and it is the successor of D3. §C.4 now states the implied ceiling
(N ≤ 66,371) and the arithmetic reaching it is correct. But the two "required cut" figures under it
are computed against a surface that **excludes `AGENTS.md`** — the file `REQ-AMC-013`, one screen
below, insists the achieved figure must be measured over. §D.2 states plainly that the Claude side
**retains** every relocated clause, so the contract layer is net-additive to the always-loaded
surface. The stated cut of 4,841 / 8,911 tokens is therefore understated by the whole size of
`AGENTS.md` — up to 6,144 tokens. §C.4 exists precisely so M1 sizes its work correctly before
starting; as written it will size it wrong, and the error surfaces at M5, which is the failure D3
was raised to prevent.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-AMC-001` … `REQ-AMC-018` (18) and
  `AC-AMC-001` … `AC-AMC-024` (24), both extracted ordered by anchored grep on the bold-prefixed
  identifiers. Sequential, no gaps, no duplicates, uniform 3-digit padding. Counts unchanged by
  this delta — verified, not accepted.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md`
  §C) only, never against an AC. `REQ-AMC-008` and `REQ-AMC-013`, the two clauses this delta
  rewrote, both remain canonical Ubiquitous form: subject + `shall` + response, with the added
  material in separate explanatory paragraphs below the clause rather than mixed into it. The
  Given-When-Then entries in `acceptance.md` are verification-layer and are graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types,
  plus `tier: L`. `version: "0.3.3"` is a quoted semver string **and** its value now matches the
  last HISTORY row (`spec.md:36`), which is what D4 fixed.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own instruction
  surface and names no per-language tooling. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only foreign reference is
  `SPEC-ALWAYS-LOADED-DIET-001`; its `spec.md` frontmatter reads `status: completed`, not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` occurrence count in `spec.md` → 0.
  Auto-PASS.
- **[PASS] MP-7 clarification gate** — recursive scan for `[NEEDS CLARIFICATION` over `plan.md`
  and `research.md` → 0 matches.

No must-pass failure. The FAIL is carried by the blocking finding E1 below.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | The band, the enumeration content, the ordering, and the measured state are each now stated unambiguously (`spec.md:240-262`, `acceptance.md:126-152`). Deduction: §C.4 states a diet target a reader will act on that is wrong by roughly a factor of two (E1) — a stated-but-wrong number misdirects more than an absent one, which was D3's own argument. |
| Completeness | 0.80 | 0.75-1.0 | All sections present; 5 `### Out of Scope —` H3 sub-headings; version + HISTORY current with a provenance rule above the table. Deductions: E1, and no milestone tests the diet against the ratchet target — M1's stop condition checks only that the *contract* fits 24,576 B, never that the *diet* reaches 66,371. |
| Testability | 0.92 | 0.75-1.0 | `AC-AMC-019` now reads the band and the product (`acceptance.md:148-152`); `AC-AMC-017` asserts the enumeration contains `AGENTS.md`; `AC-AMC-018`'s ordering assertion binds the evidence co-temporally. The iteration-2 counterexample no longer passes — re-run below. |
| Traceability | 0.92 | 0.75-1.0 | `REQ-AMC-013`'s band is now covered by `AC-AMC-019`; `REQ-AMC-008`'s enumeration clause by `AC-AMC-017`; `REQ-AMC-017` → `AC-AMC-023`, `REQ-AMC-018` → `AC-AMC-024`. No uncovered REQ found; no AC references a non-existent REQ. |

Harmonic mean = 4 / (1/0.85 + 1/0.80 + 1/0.92 + 1/0.92) = 4 / 4.6004 = **0.870**.

**The aggregate clears 0.85.** I am recording that explicitly rather than tuning a dimension down
to make the verdict follow the number. The FAIL is carried by E1 as a blocking correctness defect,
and the dispatcher can weigh PASS-with-debt against it on those terms — see § Recommendation.

---

## Rulings on the five questions put to me

### 1. Does D1's fix close the counterexample, and did the hazard move again? **Closed. It did not move.**

Re-running the SPEC's own counterexample verbatim against the amended criterion set:

| Step | Value | Criterion | Result |
|---|---|---|---|
| achieved N (`AC-AMC-018`) | 60,000 | branch + revision + post-diet state recorded | PASS |
| declared ratio (constant's comment) | 25 % | `AC-AMC-020`: the ratio is stated there | PASS |
| constant | 75,000 | `REQ-AMC-013` ≤ 75,000 | PASS |
| `AC-AMC-019` half 1 | 25 % vs the 13 %-17 % band | **band read** | **FAIL** |
| `AC-AMC-019` half 2 | 75,000 vs 60,000 × 1.25, Δ = 0 | within ±1,000 | PASS |

The run now fails, and it fails on exactly the half that was missing. The two-sentence rationale
below the criterion (`acceptance.md:150-152`) states correctly why the agreement check alone is
insufficient — the ratio and the constant are chosen by one actor in one edit, so a check that
reads the ratio from that same edit has a free variable.

I then walked the five degrees of freedom the dispatch named, to test whether the hazard relocated
a third time:

| Variable | Bound by | Free? |
|---|---|---|
| the ratio | `REQ-AMC-013` band + `AC-AMC-019` half 1 | no |
| N | `AC-AMC-018` — measured by a test run, not declared | no |
| the enumeration N is measured over | `REQ-AMC-008` + `REQ-AMC-013` ¶2 + `AC-AMC-017` | no |
| the moment of measurement | `AC-AMC-018` post-diet state (`AC-AMC-021` passes) + the ordering assertion | no |
| "the achieved figure" | `REQ-AMC-013` ¶2 defines it against the extended enumeration | no |

One residual slack, sub-blocking: the ±1,000 tolerance loosens the derived ceiling from 66,371 to
67,256 (`(75,000 + 1,000) / 1.13`). §C.4 uses the tight figure. The direction is conservative and
the gap is 885 tokens; noted, not routed.

### 2. Is the D2 fix sufficient at the mechanism level? **Yes — the detection is co-temporal, not after-the-fact.**

This was the sharp version of the question and it turns on one sentence. `AC-AMC-018`'s ordering
assertion requires the enumeration to be confirmed to contain `AGENTS.md` **"in the same run that
produces N, not in a later one"** (`acceptance.md:139-143`). That is a constraint on the *evidence*,
not on the actor's intent: evidence that does not carry both the `always-loaded surface = N tokens`
line and the enumeration assertion from one invocation does not satisfy the criterion, and
"the extension landed afterwards" is explicitly ruled out as a cure.

I checked the one place this could have been hollow. The criterion's command selects tests by the
regex `Budget|AlwaysLoaded`, so co-execution depends on the run-phase author naming the enumeration
test so that selector matches it. That is not a hole: an author who names it outside the selector
produces evidence lacking the assertion, and the criterion fails. The requirement is self-enforcing
on the evidence rather than on a naming convention nobody stated.

The requirement-layer / code-layer split the dispatch flagged is real but correctly handled:
`plan.md` M5 carries the extension as a code change with the ordering restated
(`plan.md:141-149`), adds it as a **fourth fixed slot** keeping the hermetic absent-file-measures-0
treatment, so the pre-`AGENTS.md` baseline is unchanged and the enumeration stays one path
(`REQ-AMC-008`'s no-second-measurement-path constraint holds). Verified against the function:
`alwaysLoadedSurface()` (`internal/config/token_budget_guard.go:107-134`) carries exactly three
fixed slots today — `CLAUDE.md`, the output style, `MEMORY.md` — read from source, not from the
SPEC's description of it. `spec.md` §A.1 (`spec.md:48-51`) now says the same three and names the
extension as owed, so the two statements no longer contradict.

### 3. Does the ordering constraint bind? **Yes — a wrong-order run fails a criterion, not merely prose.**

`REQ-AMC-013` ¶3 (`spec.md:257-266`) orders the extension before *any measurement cited as a
ratchet basis*, and argues the point correctly: measurements taken in the gap "manufacture false
evidence" that stays quotable after the gap closes. The binding half is `AC-AMC-018`'s ordering
assertion (question 2). `plan.md` carries it in both places — M1's sequencing boundary allows
parallel work but pins the citation moment (`plan.md:60-67`), M5 restates it on the code change
(`plan.md:141-143`). An actor who measures first and extends second produces a figure that fails
`AC-AMC-018`. Binds.

### 4. Is §C.4's arithmetic right, and is the implied cut stated honestly? **The arithmetic is right. The cut is not.**

The derivation checks out: `1.13 × N ≤ 75,000` ⟹ N ≤ 66,371.68 ⟹ **66,371** tokens.
71,212 − 66,371 = 4,841 ✓. 75,282 − 66,371 = 8,911 ✓. Both subtractions are correct.

They are subtractions from the wrong surface. See E1 below — this is the finding.

Second half of the question, on reachability: under the corrected figures the integration-state cut
is ≈ 15,055 tokens ≈ 60,220 B, against a rule surface of 202,621 B (`spec.md:45`). That is **29.7 %
of the entire always-loaded rule tree**, and `REQ-AMC-002` restricts what may be removed to
rationale, procedure, worked examples, incident records, and cross-reference tables — no obligation
may move. The precedent SPEC's stub + lazy-companion pattern has achieved cuts of that order on
individual files, so I do not find the goal unreachable. I find it **unargued**: the SPEC states a
target it will miss and never tests the real one. `plan.md` M1's stop condition
(`plan.md:86-91`) fires only if the *contract* overflows 24,576 B; nothing fires if the *diet*
lands above the ratchet ceiling, which is the failure that costs M2-M4 of work.

### 5. Regression over all sixteen prior findings: **none regressed.** Spot re-executed.

| # | Finding | Status | Evidence I ran now |
|---|---|---|---|
| i1-D2 | `AC-AMC-010` fails on a correct tree | holds | tracked-file scan for a root-relative `AGENTS.md` excluding the template mirror → 0; the `CLAUDE.md` analogue → 6 |
| i1-D3 | integration branch undefined | holds | `acceptance.md:126-128` carries the `release/vX.Y.Z` + merged-sibling discriminator and both revision recordings |
| i1-D4 | unreproducible fixture | holds | `probe-fixture.sh` confirmed tracked |
| i1-D5 | line-grep proxy undisclosed | holds | `spec.md` §A.4 / §D.1's "bracket, not a point" paragraph intact |
| i1-D6 | cap-raise rationale contradicted | holds | scan for the retired premise → 1 hit, inside the §D.8 correction |
| i1-D7 | nested-`CLAUDE.md` asymmetry | holds | `design.md:145-152` two-property table intact |
| i1-D8 | `REQ-AMC-006` had no AC | holds | `AC-AMC-012` at `acceptance.md:85` |
| i1-D9 | `AC-AMC-013` names no mechanism | holds | command re-executed against this tree: doubled `CLAUDE.md` → 4 duplicate lines, single copy → 0. Discriminates in both directions |
| i1-D10 | `REQ-AMC-004` mislabelled | holds | reads `(Unwanted)` |
| i1-D11 | `REQ-AMC-006` led with `MAY` | holds | "This SPEC **shall** record" |
| i1-D12 | inline rationale in two clauses | holds | `REQ-AMC-005` → "§D.6"; `REQ-AMC-009` → "Rationale: §D.7" |
| i2-N1 | headroom ratio a free variable | **RESOLVED** | question 1 above — counterexample now fails |
| i2-N2 | `design.md:173` cited `AC-AMC-012` | holds | scan for `AC-AMC-012` over the SPEC dir → 3 hits: the criterion itself, one HISTORY row, one `progress.md` mention. None in `design.md` |
| i3-D1 | the band has no criterion | **RESOLVED** | `acceptance.md:148-152` |
| i3-D2 | achieved figure omits `AGENTS.md` | **RESOLVED** (requirement + verification + plan) | `spec.md:200-206`, `spec.md:250-262`, `acceptance.md:126-131`, `plan.md:141-149` |
| i3-D3 | implied diet target unstated | **RESOLVED with a defect** | ceiling stated (`spec.md:229-239`); the cut figures under it are wrong — E1 |
| i3-D4 | version + HISTORY stale | **RESOLVED** | `version: "0.3.3"` = last HISTORY row; two missing rows added; provenance rule at `spec.md:20-26` |
| i3-D5 (optional) | "post-diet" undefined | **RESOLVED** | `acceptance.md:126-130` defines it as a state where `AC-AMC-021` passes |

No stagnation: nothing has persisted unchanged across iterations 2, 3 and 4.

---

## Defects Found

E1. **§C.4's required-cut figures are computed over a surface that excludes `AGENTS.md`** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md`:L229-239 (against L250-262 and L327-331) —
§C.4 subtracts 66,371 from 71,212 and from 75,282. Both minuends are measurements of the
**unextended** enumeration — `spec.md:45-49`'s table has four rows and `AGENTS.md` is not one of
them, because the file does not exist yet. One screen below, `REQ-AMC-013` ¶2 requires the achieved
figure to be measured over an enumeration that **includes** `AGENTS.md`. The two numbers are not
the same quantity.

The contract layer is **net-additive**, and the SPEC says so itself: §D.2 (`spec.md:327-331`) —
excluding Claude-only clauses from `AGENTS.md` "removes nothing from either harness's binding
surface: **they remain always-loaded on the Claude side**." `REQ-AMC-002` forbids moving any
obligation off the always-loaded rules, and `REQ-AMC-001` requires `AGENTS.md` to carry every
Codex-binding `[HARD]` clause. So the clauses exist in both places by construction, and
`AGENTS.md` joins the measured surface with nothing removed in exchange for it.

Corrected arithmetic, at `REQ-AMC-004`'s ceiling (24,576 B = 6,144 tokens at the guard's `char/4`,
`token_budget_guard.go:41-46`):

| Tree | §C.4 states | Surface + `AGENTS.md` | Actual cut to reach 66,371 | Understatement |
|---|---:|---:|---:|---:|
| this worktree | 4,841 | 284,850 B + 24,576 B = 77,356 tok | **10,985 tok** (43,940 B) | ×2.27 |
| integration state | 8,911 | 301,128 B + 24,576 B = 81,426 tok | **15,055 tok** (60,220 B) | ×1.69 |

§C.4's own closing sentence is "M1 sizes its work against this figure" — so the figure is
load-bearing for the milestone the section was written to size, and an M1 actor who hits 4,841
exactly will discover at M5 that `REQ-AMC-013` is unsatisfiable, after M2-M4 have landed. That is
the discovered-at-M5 failure D3 existed to prevent, reproduced one layer up. —
Severity: **critical** — Class: **blocking** —
Required fix: correct the two figures in §C.4's table to include the contract layer's own
contribution, and state the dependency in one sentence: the achieved figure is measured over an
enumeration containing `AGENTS.md` (`REQ-AMC-013`), the Claude side retains every clause the
contract duplicates (§D.2), so the required cut is `stated cut + |AGENTS.md|` — up to 10,985 /
15,055 tokens at the 24,576 B ceiling. No requirement changes and no criterion is added.

E2. **No milestone tests the diet against the ratchet ceiling** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/plan.md`:L86-91 — M1's stop condition fires only when the
*contract* projection exceeds 24,576 B. Nothing checks that the *diet* will reach N ≤ 66,371, which
is the constraint that decides whether M5 can close at all. With E1 fixed the number is available
at M1; without a check, reaching it is left to whoever notices. — Severity: **major** —
Class: **blocking** — Required fix: extend M1's stop condition with a second arm — project the
post-diet surface including the contract layer against 66,371 and return a blocker naming the
shortfall if it does not reach it, using the same two-lever framing the existing arm uses.

E3. **The ±1,000 tolerance widens the derived ceiling by 885 tokens** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md`:L231-233 — §C.4 derives 66,371 from
`1.13 × N ≤ 75,000`, but `AC-AMC-019` admits agreement within ±1,000, so a constant of
`N × 1.13 − 1,000` passes and the true bound is 67,256. The direction is conservative.
— Severity: **minor** — Class: **optional** — Suggested fix: none required; note the tolerance in
§C.4 if the exact bound is ever quoted as a target.

The six optional findings carried since iteration 2 (§5.2 ordering; the §D.5 numbering gap —
re-confirmed present, the `### D.n` headings run D.1-D.4 then D.6; `probe-fixture.sh` printing
rather than asserting; `AC-AMC-021`'s "make build has been run"; `AC-AMC-016`'s prose/GWT mix; the
four "(Event-detected)" labels) are unchanged. I re-checked each and consider none of them blocking.
Not routed, and nothing has changed that would make me route one now.

---

## Recommendation

**FAIL**, on E1 alone. Every defect this iteration was scoped to is resolved, the sixteen prior
findings hold, and the aggregate clears the Tier L threshold — the failure is one number that is
wrong in the section written to prevent exactly this failure.

**The smallest change reaching PASS is a single edit to `spec.md` §C.4**: replace the two
required-cut figures with figures that include the contract layer, and add the one sentence tying
them to `REQ-AMC-013` ¶2 and §D.2. No requirement changes, no criterion is added, nothing touches a
closed ruling. E2 is a second, cheap edit I would route in the same pass — it converts the
corrected number into something a milestone actually tests — but E1 alone is what the verdict turns
on.

**If the dispatcher prefers PASS-with-debt**, this is a defensible case for it and I will say so
plainly: E1 is a documentation arithmetic error, not a structural one; it is visible to anyone who
reads §C.4 next to `REQ-AMC-013`; and unlike iteration 3's D1 and D2 it is fully detectable after
the fact, because the run-phase measurement at M5 will surface the true figure. The cost of
carrying it is one wasted sizing pass at M1, not a ratchet that ratchets nothing. That is a
materially weaker reason to block than iteration 3 had, and the score reflects it.

---

## Evidence and Gaps

**Commands actually run** (all from `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82`, cwd
re-anchored per call): working-directory confirmation; short-revision read → `d378e5358` (start and
end); porcelain status over `.moai/specs` → empty (start and end); the REQ/AC ordered-extraction
greps; the tracked-file scan for a root-relative `AGENTS.md` excluding the template mirror, and its
`CLAUDE.md` analogue; the tracked-file check on `probe-fixture.sh`; the `AC-AMC-013` duplicate-line
scan against a doubled and a single `CLAUDE.md`; the `AC-AMC-012` scan over the SPEC directory; the
`### Out of Scope —` heading count → 5; the `[NEEDS CLARIFICATION` scan over `plan.md` and
`research.md` → 0; the `syscall` count in `spec.md` → 0; the status read on
`.moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md` → `completed`; ranged reads of
`internal/config/token_budget_guard.go:16-50` and `:107-145` (constant derivation, `estimateTokens`,
`alwaysLoadedSurface`'s three fixed slots); ranged reads of `spec.md`
frontmatter/HISTORY/§A.1/§C.1/§C.2/§C.4/§D.1/§D.2, `acceptance.md:85-165`, `plan.md` M1/M3/M4/M5,
`design.md:140-180`; the Tier-table read in
`.claude/rules/moai/workflow/spec-workflow.md` for the 0.85 Tier L threshold.

**Gaps — what I did NOT observe.**
- I did **not** run the Go test suite (dispatch instruction). `AC-AMC-016(a)`'s `--- PASS` lines
  and `AC-AMC-018`'s exit-0 claim are carried from iteration 2's execution, not re-measured. The
  token figures I cite (71,212 / 75,282 / 64,624) are read from `spec.md`, `measurement.md` and the
  guard's derivation comment; I re-derived 71,212 = 284,850 / 4 and the E1 table arithmetic by hand
  from those, but did not re-run the measurement.
- `|AGENTS.md|` in E1 is taken at `REQ-AMC-004`'s **ceiling** (24,576 B). The file does not exist;
  its real size is M1/M2's deliverable and could be smaller, which shrinks E1's understatement
  proportionally. What does not shrink is that §C.4's figure treats it as **zero**.
- I did **not** verify anything about Claude Code's own loader accounting. E1 rests on the SPEC's
  own §D.2 statement that relocated clauses remain always-loaded on the Claude side, plus
  `REQ-AMC-001`/`REQ-AMC-002` — not on any runtime observation.
- This was a delta-scoped re-audit plus regression, not a fresh Tier L pass. `research.md` was read
  only for MP-7.

**Residual risk if run-phase inherits this as-is.** The document is now enforceable everywhere the
ratchet is concerned — the band, the enumeration, the ordering and the measurement moment are all
bound by criteria that fail when violated. What run-phase inherits is a **sizing** risk, not an
enforcement one: M1 will aim at a target roughly half the real one, and nothing between M1 and M5
will say so. The correction is available at any point before M5 and costs a re-projection; the cost
of not making it is M2-M4 landing against a diet that cannot satisfy `REQ-AMC-013`.
