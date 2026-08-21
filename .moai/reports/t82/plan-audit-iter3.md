# SPEC Review Report: SPEC-AGENTS-MD-CANON-001

Iteration: 3/3 (cap reached)
Verdict: **FAIL** — 0.82 against the Tier L 0.85 threshold
Overall Score: **0.82** (harmonic mean)
Signal: **STOP** (score regression 0.83 → 0.82; see § Score-regression note — the cause is audit
coverage, not document deterioration)

Audited tree: `.claude/worktrees/t82`, commit **`51fec2f25`**
("feat(spec-agents-md-canon-001): pin the headroom ratio in the requirement").
`git status --short .moai/specs` → empty at audit start and at audit end. The document did not move
during this audit.

Reasoning context ignored per M1 Context Isolation.

## Provenance correction to the iteration-2 report

The dispatch asked me to correct the iteration-2 tree label. Measured, the correction needed is
narrower than the dispatch states, and points somewhere else:

- **The commit label is already correct in the committed report.** `plan-audit-iter2.md` names
  `e5699d4fc` and carries an explicit "Provenance correction" section recording that its first
  draft said `cd6e12459` and why that was wrong. Nothing further to fix there.
- **The version label is wrong and is still wrong.** That report says "SPEC version `0.3.1`".
  `git show e5699d4fc:…/spec.md | sed -n '4p'` → `version: "0.3.0"`. The document never carried
  `0.3.1`, and it does not carry `0.3.2` now either — see D4 below. The dispatch's own "v0.3.2"
  is likewise a commit-message version, not a document version.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-AMC-001` … `REQ-AMC-018` (18, extracted ordered
  via `grep -o '^\*\*REQ-AMC-[0-9]*'`), sequential, no gaps, no duplicates, uniform 3-digit
  padding. AC side `AC-AMC-001` … `AC-AMC-024` (24), contiguous. The delta's claim "no criterion
  was added — AC count stays 24, REQ 18" is verified, not accepted.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §C),
  never against an AC. `REQ-AMC-013` remains canonical Ubiquitous form: subject
  (`AlwaysLoadedTokenBudget`) + `shall` + response, with the band inside the response clause. The
  added material is a separate explanatory paragraph, not a modality mix inside the clause — the
  same shape `REQ-AMC-014` already uses, so it is not a D12 regression.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types,
  plus `tier: L`. `version: "0.3.0"` is a quoted semver string, so the field is type-valid; its
  *value* being two revisions stale is graded under Completeness (D4), not here.
- **[N/A] MP-4 language neutrality** — the SPEC is scoped to this repository's own instruction
  surface and names no per-language tooling. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the only foreign reference is
  `SPEC-ALWAYS-LOADED-DIET-001`; `grep '^status:'` on its `spec.md` → `completed`, which is not in
  {retired, superseded, archived}. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → 0. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → 0
  matches.

No must-pass failure. The FAIL is driven by the rubric scores and the blocking findings below.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | The band is stated unambiguously (`spec.md:209-212`). Residual: "post-diet" in `AC-AMC-018` is never defined against a checkable state, and the diet magnitude the band now implies is stated nowhere (D3). |
| Completeness | 0.88 | 0.75-1.0 | All sections present; 5 `### Out of Scope —` H3 sub-headings with specific bullets. Deductions: the guard enumeration gap (D2) is unaddressed anywhere in `spec.md`/`design.md`, and version + HISTORY are stale by two revisions (D4). |
| Testability | 0.72 | 0.50-0.75 | The decisive constraint added this iteration has no criterion that reads it — `AC-AMC-019` still resolves the ratio from the constant's own comment (`acceptance.md:138-139`), so the SPEC's own counterexample still passes every criterion (D1). |
| Traceability | 0.85 | 0.75-1.0 | `REQ-AMC-013`'s band clause is uncovered: grepping `acceptance.md` for the band or the ratio returns no criterion binding it. Everything else maps. |

Harmonic mean = 4 / (1/0.85 + 1/0.88 + 1/0.72 + 1/0.85) = 4 / 4.8782 = **0.820**.

---

## Rulings on the four questions put to me

### 1. Does the pinned band close N1? **No — it closes the counterexample in the requirement layer and leaves it passing in the verification layer.**

`REQ-AMC-013` now forbids a ratio outside 13-17 %. Nothing checks it. Re-running the
counterexample verbatim against the current criterion set:

| Step | Value | Criterion | Result |
|---|---|---|---|
| achieved N (AC-AMC-018) | 60,000 | branch + `rev-parse` recorded | PASS |
| declared ratio (comment) | 25 % | AC-AMC-020: "the ratio is stated there" | PASS |
| constant | 75,000 | REQ-AMC-013 ≤ 75,000 | PASS |
| AC-AMC-019 | 75,000 vs 60,000 × 1.25 = 75,000, Δ = 0 | within ±1,000 | PASS |
| band 13-17 % | 25 % is outside | **no criterion reads the band** | not checked |

Every criterion green, ratchet nil — the same all-green run iteration 2 reported. The requirement
is violated, but the verification layer cannot see it, and the verification layer is what a run
phase executes. Grepping `acceptance.md` for the band or for "ratio" returns four hits, all of them
either the M1 compression ratio or `AC-AMC-019`/`AC-AMC-020` deferring to the constant's comment;
none bounds it.

**And the hazard has a second route, which is the one the dispatch's question 1 asks about
directly.** Yes — an actor can still reach a predetermined constant by choosing what counts as the
achieved figure, because this SPEC changes what is *in* the always-loaded surface without changing
the enumeration that measures it:

- `alwaysLoadedSurface()` (`internal/config/token_budget_guard.go`) enumerates rule files without
  `paths:` plus exactly three fixed slots — `CLAUDE.md`, the output style, `MEMORY.md`. Read from
  source, not from the SPEC's description of it.
- `design.md` §4 puts the contract at `CLAUDE.md` → `@AGENTS.md`. An `@`-imported document is
  always-loaded on the Claude side; it is **not** in the enumeration.
- So every `[HARD]` clause M1 moves out of a rule file and into `AGENTS.md` leaves the *measured*
  surface while staying in Claude's real context. Up to `REQ-AMC-004`'s 24,576 B ceiling
  (≈ 6,144 tokens at the guard's `char/4`) can drop out of the achieved figure without one byte
  leaving the always-loaded context.

This is precisely the pattern `spec.md` §D.4 cites the `SPEC-ALWAYS-LOADED-DIET-001` precedent for
avoiding — "a 72 % always-loaded reduction **with no obligation moved off the always-loaded
surface**" — and this SPEC's own design moves obligations off the *counted* surface while §D.4
carries the precedent as though it did not. The ratchet then locks the constant to an undercount,
and future `AGENTS.md` growth is invisible to the token guard by construction.

### 2. Is 13-17 % defensible? **Yes — and it checks out against the source, not just the claim.**

`internal/config/token_budget_guard.go:22-25` (the derivation comment, read directly):
baseline ≈ 64,624 tokens, "약 15% 여유(≈ 74,317)를 클린 상수로 올림". `design.md` §5.1 states the
same ("The original constant carried ~15 % over its baseline"). So the SPEC's claim that 15 %
matches what the original constant carried is supported by the code, not only by the design doc.

The ±2-point band is also the right width, and I can put a number on it: the constant actually
landed at 75,000 against 64,624, which is **16.05 %**, not 15 % — the rounding-to-clean the SPEC
describes. 16.05 % sits inside 13-17 % with 0.95 points to spare, so the band admits the historical
precedent it claims to codify. A ±1-point band would have excluded it. Not arbitrary.

**One consequence the SPEC does not state.** The band and the ≤ 75,000 ceiling together bound the
achieved figure: 1.13 × N ≤ 75,000 ⟹ **N ≤ 66,371**. Above that, no constant satisfies
`REQ-AMC-013` at all. This tree measures **71,212** (`spec.md:49`, SSOT
`.moai/reports/t82/measurement.md:43`) and the integration state that forced the raise measured
**75,282** (`spec.md:229-230`). So the pinned band silently sets a diet target of at least
**4,841 tokens off this tree** — 8,911 off the integration state — and no goal, requirement, or
milestone states a target of that magnitude. A run phase discovers this at M5, after the diet is
already done. Worse: the cheapest way to reach 66,371 is exactly the unmeasured relocation in D2
(≈ 6,144 tokens), so the two findings compose into a route where the ratchet is satisfied without
a real reduction.

### 3. Did the N1 paragraph introduce anything needing testing? **One claim, and it verifies.**

The paragraph is otherwise explanatory. Its single checkable assertion — "15 % matches the
allowance the original constant already carried (`design.md` §5.1)" — I verified against both
`design.md` §5.1 and the Go comment (question 2). No new criterion is needed for the paragraph
itself. What needs a criterion is the band in the clause above it, which is D1.

### 4. Regression over the eleven resolved findings: **none regressed.** Spot re-executed, not re-read.

| # | Iteration-1 finding | Status at `51fec2f25` | Evidence I ran now |
|---|---|---|---|
| D2 | `AC-AMC-010` fails on a correct tree | holds | `git ls-files --full-name ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'` → 0; `CLAUDE.md` analogue → 6 |
| D3 | integration branch undefined | holds | `spec.md:225-231` carries the `release/vX.Y.Z` + merged-sibling discriminator and the two `git` recordings |
| D4 | unreproducible fixture | holds | `git ls-files .moai/reports/t82/probe-fixture.sh` → tracked |
| D5 | line-grep proxy undisclosed | holds | `spec.md` §A.4 / `AC-AMC-003` unchanged |
| D6 | cap-raise rationale contradicted | holds | grep for the retired premise over the SPEC dir → 4 hits, every one inside an explicit correction (`spec.md:365` "That premise is false") |
| D7 | nested-`CLAUDE.md` asymmetry | holds | `design.md:145-152` two-property table intact |
| D8 | `REQ-AMC-006` had no AC | holds | `AC-AMC-012` at `acceptance.md:85` |
| D9 | `AC-AMC-013` names no mechanism | holds | command re-executed: the duplicate-line scan returns 4 against a doubled `CLAUDE.md` and 0 against a single copy. Discriminates in both directions |
| D10 | `REQ-AMC-004` mislabelled | holds | `spec.md:170` reads `(Unwanted)` |
| D11 | `REQ-AMC-006` led with `MAY` | holds | `spec.md:179` "This SPEC **shall** record" |
| D12 | inline rationale in two clauses | holds | `spec.md:192` "Rationale: §D.7"; `spec.md:177` "§D.6" |
| N2 | `design.md:173` cited `AC-AMC-012` | **RESOLVED** | `design.md:173` now reads "AC-AMC-013 tests exactly this"; grepping `AC-AMC-012` over the SPEC dir returns only `acceptance.md:85` (the criterion itself) and one correct `progress.md` mention |

No stagnation: nothing has appeared unchanged across all three iterations.

---

## Defects Found

D1. **N1-residual-a — the band has no criterion** — `.moai/specs/SPEC-AGENTS-MD-CANON-001/acceptance.md`:L138-139 —
`AC-AMC-019` resolves the headroom ratio from "the constant's comment", so `REQ-AMC-013`'s 13-17 %
band is never read by any criterion. The SPEC's own counterexample (achieved 60,000, ratio 25 %,
constant 75,000, Δ 0) still passes every one of the 24 criteria. — Severity: **critical** —
Class: **blocking** — Required fix: amend `AC-AMC-019` in place, no new criterion:
"Given the achieved figure N from AC-AMC-018 and the headroom ratio stated in the constant's
comment, When the ratio is checked against the 13 %-17 % band `REQ-AMC-013` fixes **and**
`AlwaysLoadedTokenBudget` is compared against `N × (1 + ratio)`, Then the ratio is inside the band
and the two figures agree within ±1,000 tokens." (AC budget: stays at 24 of 25.)

D2. **N1-residual-b — the achieved figure omits a file this SPEC makes always-loaded** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md`:L209 (with `design.md`:L162-168) —
`alwaysLoadedSurface()` enumerates rules-without-`paths:` plus `CLAUDE.md`, the output style, and
`MEMORY.md`. `design.md` §4 makes `AGENTS.md` an `@`-import of `CLAUDE.md`, i.e. always-loaded and
unenumerated. Clauses relocated into `AGENTS.md` therefore reduce the achieved figure by up to
≈ 6,144 tokens (`REQ-AMC-004`'s 24,576 B ÷ 4) without leaving the always-loaded context, and the
ratchet blesses that reduction. — Severity: **critical** — Class: **blocking** —
Required fix: one clause on `REQ-AMC-013` (or `REQ-AMC-008`) requiring the achieved figure to be
measured over an enumeration that includes the root `AGENTS.md` and every `@`-imported contract
document, plus one clause on `AC-AMC-017` asserting the enumeration contains it. No new criterion.

D3. **The diet target the band implies is stated nowhere** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md`:L209-212 vs L49 / L229-230 —
`1.13 × N ≤ 75,000` makes `REQ-AMC-013` unsatisfiable for any achieved figure above 66,371 tokens.
This tree measures 71,212 and the integration state measured 75,282, so the band requires a cut of
≥ 4,841 (≥ 8,911 from integration) that §B goal 2 describes only as "reduce … enough". —
Severity: **major** — Class: **blocking** — Required fix: state the derived ceiling on the achieved
figure (≤ 66,371 tokens) in §C.4 alongside the band, so the target is visible before M1 rather than
discovered at M5.

D4. **`version` and HISTORY are two revisions stale** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/spec.md`:L4 and L20-26 — `version: "0.3.0"` and the HISTORY
table's last row (`0.3.0`) have been unchanged across `e5699d4fc` ("v0.3.1" per its message) and
`51fec2f25` ("v0.3.2"). The document does not record either revision, which is what produced the
label error in the iteration-2 report and the dispatch's "v0.3.2". A SPEC carrying a §D.4
"measurement provenance" standing constraint should not misstate its own provenance. —
Severity: **major** — Class: **blocking** — Required fix: bump to `0.3.2` and add the two missing
HISTORY rows (0.3.1 = the `AC-AMC-016` negative-test citation; 0.3.2 = the pinned band).

D5. **"post-diet" is not defined against a checkable state** —
`.moai/specs/SPEC-AGENTS-MD-CANON-001/acceptance.md`:L131 — `AC-AMC-018` records branch name and
`rev-list --count main..HEAD`, which identify *a* tree but do not establish that this SPEC's diet
landed in it. A mid-diet measurement satisfies the criterion verbatim and inflates N; with D1's fix
in place, an inflated N admits a proportionally inflated constant. — Severity: **minor** —
Class: **optional** — Suggested fix: define the measured state as one in which `AC-AMC-021` passes
(every file this SPEC lands is present), or name the milestone whose completion the measurement
follows.

The six optional findings from iteration 2 (§5.2 ordering, the §D.5 numbering gap — confirmed
still present, `spec.md` jumps §D.4 → §D.6 — `probe-fixture.sh` printing rather than asserting,
`AC-AMC-021`'s "make build has been run", `AC-AMC-016`'s prose/GWT mix, the four "(Event-detected)"
labels) are unchanged. I re-checked each and consider none of them blocking. Not routed.

---

## Score-regression note (STOP)

0.83 → 0.82 triggers the STOP clause, so it is emitted. But the orchestrator needs the cause, and
the cause is not deterioration: **the document did not get worse.** N2 was fixed, N1's requirement
layer improved, and all eleven prior resolutions held under re-execution. The aggregate fell
because iteration 3 surfaced two defects (D2, D4) that were present at iteration 2 and were not
detected then — D2 because iteration 2 read the SPEC's description of `alwaysLoadedSurface()`
rather than the function, D4 because iteration 2 took the version from a commit message rather
than the frontmatter. Both are audit-coverage gaps, not regressions of the author's.

Read the STOP as "the third iteration found older defects", not "the SPEC is deteriorating".

## Recommendation

**FAIL.** The smallest change that reaches PASS is four edits, none of which adds a criterion and
none of which touches the closed rulings:

1. `acceptance.md` `AC-AMC-019` — add the band check to the existing When/Then (D1, exact wording
   above). This is the one edit that is genuinely load-bearing: without it the pinned band is
   unenforced and iteration 3's fix changes nothing a run phase executes.
2. `spec.md` `REQ-AMC-013` (or `REQ-AMC-008`) + `acceptance.md` `AC-AMC-017` — require the achieved
   figure's enumeration to include the root `AGENTS.md` and the `@`-imported contract set (D2).
3. `spec.md` §C.4 — state the ≤ 66,371-token ceiling on the achieved figure that the band implies
   (D3).
4. `spec.md` frontmatter + HISTORY — bump to `0.3.2`, add the two missing rows (D4).

With those four, Testability recovers to ≈ 0.90 (the ratchet's decisive input becomes checkable),
Traceability to ≈ 0.95, Completeness to ≈ 0.95, Clarity to ≈ 0.90 — harmonic mean ≈ 0.92, above the
Tier L threshold.

**Iteration cap.** This is iteration 3 of 3. The retry loop must not continue unconditionally; the
orchestrator escalates to the user with three options: (1) PASS-with-debt — accept and carry D1-D4
as documented debt into run phase, which I do **not** recommend, because D1 and D2 both let M5 land
a ratchet that ratchets nothing and neither is detectable after the fact; (2) apply the four edits
above and re-audit scoped to that delta — the cheapest path, and the one the defect list is written
to support; (3) explicit user override to extend past iteration 3.

## Evidence and Gaps

**Commands actually run** (all from `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t82`, cwd
re-anchored per call): `git rev-parse --short HEAD` → `51fec2f25`; `git status --short .moai/specs`
→ empty (start and end); `git show 51fec2f25 -- <specdir>`; `git show e5699d4fc:…/spec.md`;
the REQ/AC extraction greps; the D2 / D6 / D9 re-executions; `git ls-files` on `probe-fixture.sh`;
`sed -n` reads of `token_budget_guard.go:1-38` and `alwaysLoadedSurface()`;
`grep '^status:' .moai/specs/SPEC-ALWAYS-LOADED-DIET-001/spec.md`;
`grep -rn '\[NEEDS CLARIFICATION' plan.md research.md`; `grep -c 'syscall' spec.md`.

**Gaps — what I did NOT observe.**
- I did **not** run the Go test suite (dispatch instruction). So `AC-AMC-016(a)`'s quoted
  `--- PASS` lines and `AC-AMC-018`'s exit-0 claim are carried from iteration 2's execution, not
  re-measured here. The token figures I cite (71,212 / 75,282 / 64,624) are read from `spec.md`,
  `measurement.md`, and the Go comment — I did not re-derive them from a run.
- I did **not** verify that `@`-imported files are excluded from Claude's own accounting in any
  runtime sense; D2 rests on reading `alwaysLoadedSurface()`'s enumeration and `design.md` §4's
  topology, which is sufficient for the claim made (the guard does not count it) and not for any
  claim about Claude's loader.
- `plan.md` and `research.md` were read only where the delta or a must-pass check reached them;
  this was a delta-scoped re-audit, not a fresh Tier L pass over all five artifacts.

**Residual risk.** D2's magnitude (≈ 6,144 tokens) is an upper bound from `REQ-AMC-004`'s ceiling,
not a prediction of what M1 will actually relocate; the real figure is whatever the classification
pass moves. If M1 relocates far less, D2's effect on the ratchet shrinks proportionally — but the
enumeration gap itself remains, and it is permanent once the constant is set against it.
