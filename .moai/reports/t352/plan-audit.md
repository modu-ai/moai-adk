# SPEC Review Report: SPEC-TEMPDIR-CLEANUP-RACE-001

Card: t352 · Lane worktree: `.claude/worktrees/t352` · Branch `WT-tempdir-cleanup-race` · HEAD `77b2bcae6`
SPEC version audited: **0.1.1**
Iteration: **2** (delta re-audit over the iteration-1 defect list; iteration-1 record preserved in the Appendix)
Verdict: **PASS**
Overall Score: **0.93** — Tier S threshold **0.75** (`spec-workflow.md:140`)
Score movement: **0.88 → 0.93, monotonic increase.** No STOP escalation (the LEAN score-regression clause fires only on a decrease).

Reasoning context ignored per M1 Context Isolation. This is a scoped re-audit: the enumerated
D1-D8 delta plus the four suspicion probes the coordinator named, plus a fresh read of the
corrected evidence base. Not a from-scratch full audit; the iteration-1 must-pass findings are
re-confirmed where the repair touched them and carried forward otherwise.

---

## Verdict at a glance

| Defect (iter-1) | Severity then | Status now |
|---|---|---|
| D1 caller inventory omits `internal/cli/deps.go:221` | major/blocking | **CLOSED** |
| D2 AC-TCR-002b records no base SHA | major/blocking | **CLOSED** |
| D3 1-in-5 CI figure misattributed | major/blocking | **CLOSED** |
| D4 fixture rationale names the wrong quantity | minor/blocking | **CLOSED** |
| D5 Tier S budget overrun | minor/blocking | **CLOSED as recorded** (operator disposition; justification judged honest — see below) |
| D6 CI runtime headroom unstated | minor/optional | **CLOSED** (with one unattributed clause — D9) |
| D7 "every durable write" universal, one writer named | minor/optional | **CLOSED** |
| D8 REQ-TCR-002 compound | minor/optional | **PARTIALLY CLOSED** — compounding reduced, not eliminated |
| — new | — | **D9** unattributed superlative in the new §D constraint (minor) |
| — new | — | **D10** AC-TCR-002b(ii) is vacuous-green on a path typo (minor) |

Zero must-pass failures. No blocking-class defect remains open; D8/D9/D10 are minor.

---

## Must-Pass Results (re-confirmed on 0.1.1)

- **[PASS] MP-1 REQ number consistency** — REQ-TCR-001…006 (`spec.md:106,110,114,118,122,126`),
  sequential, no gap, no duplicate, uniform padding. The new REQ-TCR-006 extends the run without
  breaking it. Requirement count 6 ≤ Tier S ceiling 8 (`spec-workflow.md:148`).
- **[PASS] MP-2 GEARS compliance** — requirement layer only. REQ-TCR-006 (`spec.md:126-128`) is a
  `shall not` unwanted-pattern entry, consistent with the other five; the split of the old
  REQ-TCR-002 left both halves pattern-conformant (`:110-112` unwanted, `:126-128` unwanted).
  Given-When-Then entries in `acceptance.md` §D.1 remain verification-layer and are graded under
  Group 4, not here.
- **[PASS] MP-3 frontmatter** — 12 canonical fields intact; `version: "0.1.1"` quoted semver, HISTORY
  gained the matching 0.1.1 row (`spec.md:28`). `updated:` remains `2026-08-28`, which is correct —
  the repair landed the same day.
- **[N/A] MP-4** — single-language SPEC.
- **[PASS] MP-5 D7** — reference set unchanged (three related SPECs, statuses `completed` /
  `implemented` / `in-progress`); `mcp__moai__spec_audit` (project_root = this worktree) returns one
  INFO `EraAutoDetected` row and no drift finding.
- **[PASS] MP-6 D8** — `grep -c syscall` over spec/plan/acceptance → `0, 0, 0`.
- **[PASS] MP-7** — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC directory and
  `.moai/reports/t352/` → exit 1, no match.

---

## Per-defect judgment

### D1 — **CLOSED**, and the mechanism was verified rather than accepted

The repair is present at all four claimed loci: `spec.md:74-86` (§A.3 retitled "two callers, one of
them production", each caller its own bullet, the non-variadic constructor named at `:84-86`),
`spec.md:126-128` (REQ-TCR-006), `spec.md:173-178` (the §D "The seam is variadic" constraint),
`plan.md:62-71` (§D.3 first bullet), `acceptance.md:62-76` (AC-TCR-002b clause (ii)).

Verified independently, not taken from the repair's own description:

- `internal/hook/session_start.go:41` — `func NewSessionStartHandler(cfg ConfigProvider) Handler`.
  Non-variadic, one parameter. The SPEC's premise is true.
- `internal/cli/deps.go:221` — `deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))`.
  Quoted accurately, including the surrounding `Register(...)` call.
- **The prescribed shape actually works.** `NewSessionStartHandler(cfg ConfigProvider, opts ...Option)`
  leaves `deps.go:221` (one argument) and `binary_lag_test.go:57` (`nil`) compiling unchanged, and
  admits the `(nil, <option>)` form `acceptance.md:42` presumes. There is no name collision to trip
  over: a grep for `type Option` and `func With` over non-test `internal/hook` sources returns
  nothing, so both `Option` and `WithSynchronousDeferredScans` are free identifiers in the package.

**The call shape I flagged at iter-1 `acceptance.md:33` is fixed, not relocated.** At iter-1 the
two-argument form was a presumption with nothing behind it; at 0.1.1 it is the *consequence* of a
constraint stated in three places and checked two ways — `spec.md:173-178` mandates the variadic
shape, `plan.md:140-143` (M1 step 1) carries it into the work order, AC-TCR-002b(ii) asserts
`deps.go` is absent from the file list, and AC-TCR-006/007 compile and run the package containing
it. `plan.md:182-184` adds the matching anti-pattern ("the fix is to change the seam's shape, not to
update `deps.go:221`"), which closes the obvious wrong turn.

### D2 — **CLOSED**

`acceptance.md:78-96` ("Base-SHA attribution (mandatory)") records `77b2bcae6` as the base in force
at plan time, states *why* the three-dot form is load-bearing, obliges the run phase to record the
merge-base next to its reading, and names the mid-card-merge case that moves the base. Propagated to
`plan.md:33-34` (pre-flight), `plan.md:165-167` (§F), `progress.md:16-17` (base recorded with the
observation that develop's tip had already moved to `c6aa61346`), and `acceptance.md:5-7` +
`:197-198`. The attribution gap is closed at every surface that could have carried it forward.

Re-measured now, so the recorded value is not merely repeated from the repair: the merge-base of
`origin/develop` and HEAD resolves to `77b2bcae6…`, while develop's own tip has moved again since
iteration 1. The fork point held across three different develop tips (`947f5cffb` at iter-1,
`c6aa61346` at repair time, and today's) — which is the empirical case for the three-dot form the
sub-section makes, now observed three times rather than argued.

### D3 — **CLOSED**, and correctly attributed at line level

`spec.md:88` retitles §A.4 to "…and what is not known about its frequency"; `:95-100` states "The CI
frequency of this observation is unknown", cites `t322/verdict.md:282` for observation 2's single
appearance, and explicitly returns the 1-in-5 figure to `TestGitDiffNameCount_Predicate` at `:281`.
Both citations verified verbatim in the source report: line 281 is the `TestGitDiffNameCount_Predicate`
row ("did not recur (now 1 of 5 post-landing runs)"), line 282 the
`TestBinaryLag_OneSeamServesBothSurfaces` row ("appeared", count 1). `spec.md:216-220` (§F) carries
the same split; `acceptance.md:212-214` (§D.5) adds "one appearance is not a rate — so its cost is
unquantified, not low"; `plan.md:187-189` makes reporting observation 2 as "low-frequency" an
anti-pattern. The withdrawn inference is not merely deleted anywhere — the evidence base records it
as a dated Correction under Claim, so the retraction is auditable.

### D4 — **CLOSED**

`acceptance.md:36-41` now reads "reliably finishes **after `Handle` returns** — that, not the 250 ms
join bound, is the operative condition", cites `Handle=59.59ms` at `nFiles=8000` and the two lateness
figures (43 ms / 223 ms), and keeps the 8000-file padding. The stated rationale now matches the
measurement it rests on. `plan.md:178-179`'s empty-fixture anti-pattern is unchanged and still
correct.

### D5 — **CLOSED as recorded.** Operator disposition accepted; the justification is honest.

I do not re-litigate the disposition and do not propose folding an AC. What I judge is whether
`plan.md:91-132` (§D.4) earns what it claims.

- **The scope claim is TRUE as written.** "well under 300 LOC and under 5 files" (`plan.md:101`).
  The plan's own work order (`plan.md:140-153`) touches `internal/hook/session_start.go`,
  `internal/cli/binary_lag_test.go`, one new guard file in `internal/cli`, and the probe file — four
  files, against the Tier S bound of "< 5 files" (`spec-workflow.md:140`). LOC: a variadic parameter,
  one struct field, one branch condition, plus a guard test with a padding helper — tens of lines, not
  hundreds. Tier M's stated band (300-1000 LOC, 5-15 files) is indeed false of this work, so the
  refusal to tier up is not a dodge. *One caveat, not a defect:* the count reaches four only if the
  `Option` type lands in `session_start.go` rather than a new `option.go`; a fifth file would put the
  work exactly on the boundary rather than "under" it. `plan.md:140-143` implies the former.
- **The threshold-honesty argument is verified.** "would silently raise the plan-auditor threshold
  from 0.75 to 0.80" (`plan.md:104`) is exactly what `spec-workflow.md:140-141` prescribes. The SPEC
  is arguing *against* the tier that would give it the easier budget and *for* the tier with the
  stricter scope claim — an argument that costs its author something, which is the mark of an honest
  one rather than a rationalisation.
- **§D.4's per-criterion table earns each of the nine**, it does not merely list them. Each row names
  a distinct coverage the others do not provide, and the two rows most vulnerable to "padding" are the
  ones argued hardest: AC-TCR-004a is defended as the only RED observation (`plan.md:107-112`, and
  again at `acceptance.md:123-124` "it is not foldable into AC-TCR-004b"), and AC-TCR-003 is defended
  *by its own weakness* ("disclaims itself as non-evidence of the fix, which is precisely why it is
  not a substitute for 004"). I checked the table for a row that only restates its AC's title: none
  does. AC-TCR-002a is the weakest (a one-line grep), and the table says so in those words rather than
  inflating it.
- **The separate-file justification is a real argument** (`plan.md:128-132`): the 2-file prescription
  assumes an inlinable AC set, this one is not, and compressing to fit would cost content. Recorded as
  a deviation, not hidden.

Verdict on D5: the overrun is declared at three surfaces (`spec.md:199-205`, `acceptance.md:9-12`,
`plan.md:91-132`), the scope claim underpinning "stay S" is true, and no criterion was deleted to
make a count fit. This is the outcome the operator instructed, discharged honestly.

### D6 — **CLOSED** (see D9 for one clause inside it)

`spec.md:184-190` adds the CI-runtime-headroom constraint, cites `.github/workflows/ci.yml:238`
(verified: the race-suite step runs the whole module with no `-timeout` flag, hence the 10-minute
per-package default), bounds the expected addition with the reproduction's own 1.61 s
four-configuration wall time, and — the part that matters — makes the cost *measured* rather than
argued: `acceptance.md:166-168` requires AC-TCR-007 to record the package wall-clock, and
`plan.md:168-170` repeats the obligation in §F.

### D7 — **CLOSED**

`spec.md:59-65` states that the guard matches REQ-TCR-001's "every durable write" scope *by
construction* (whole-entry-set comparison), and cites the `binaryLagAdvisory` exemption to
`internal/binlag/binlag.go:101,111` — re-verified: those two lines invoke `git … rev-parse HEAD` and
`git … merge-base --is-ancestor`, and no write call exists in the package. `acceptance.md:50-52`
adds the matching scope note to AC-TCR-001, and `plan.md:81-89` reframes the exemption as "cited
rather than asserted" plus belt-and-braces ("a writer this plan wrongly exempted would still be
caught"). That last clause is the right shape: it does not rest the requirement on the exemption
being correct.

### D8 — **PARTIALLY CLOSED.** The split reduced compounding; it did not eliminate it.

The coordinator's suspicion is well founded and I confirm it. Before: one requirement, four claims,
three subjects. After: REQ-TCR-002 (`spec.md:110-112`) carries two claims over one subject (the scan
remains asynchronous; remains joined with the existing bound) — genuinely de-compounded. But
REQ-TCR-006 (`spec.md:126-128`) carries two claims over two different subjects: a *variable's default*
(`deferredScansAsync` unchanged) and a *call site's immunity* (`deps.go:221` needs no modification).
Those are different obligations verified by different mechanisms — AC-TCR-002a (a grep) and
AC-TCR-002b(ii) (a file-list check) respectively, as `acceptance.md:192` itself records by mapping
REQ-TCR-006 to both.

So compounding fell from 4-claims/3-subjects to a worst case of 2-claims/2-subjects, and every claim
is now individually traced. It did not vanish; part of it moved into the new requirement created to
absorb D1. That is a real reduction, not a shuffle — but "split" overstates it.

Severity: **minor, optional class.** Both clauses are traced and decidable; splitting REQ-TCR-006
into 007 would add a seventh requirement (still within the ceiling of 8) for a presentational gain.
Recorded so the next reader is not told the compounding is gone.

---

## New findings

**D9 — minor / blocking-class — `spec.md:185-186`: an unattributed superlative.**
"…and `internal/cli` is already the slowest package in that run." Nothing in the evidence base
measures per-package durations, the reproduction explicitly records that `internal/cli`'s full
package was never run (`reproduction.md` Gaps; `spec.md:233-234`), and no test-timing output is
cited anywhere. The claim may well be true, but it is asserted, and this SPEC holds itself to
attribution discipline everywhere else in the same section — the sentence immediately after it
sources its 1.61 s figure. An unobserved claim inside a document whose §A.4 and §F were just repaired
*for* an unobserved claim is a consistency wobble worth one word.
Required fix: soften to "`internal/cli` is among the slower packages in that run (not measured here)",
or attribute it — AC-TCR-007's newly-required wall-clock will supply the number at run phase.

**D10 — minor / optional — `acceptance.md:73` + `:76`: clause (ii) is vacuous-green on a path typo.**
The `--name-only` diff restricted to `internal/cli/deps.go` exits 0 with empty output when `deps.go`
is unchanged — correct — but it *also* exits 0 with empty output if the pathspec is misspelled, if
the file is renamed, or if it moves package. The criterion cannot distinguish "the production call
site was untouched" from "I asked about a path that does not exist". This SPEC elsewhere legislates
against exactly this failure mode (AC-TCR-004a exists to exclude the vacuous direction;
`plan.md:180-181` forbids grep-as-evidence), which is why it is worth naming here rather than letting
it pass as a small thing.
Required fix: pair (ii) with a positive control in the same recording — assert that the same
`--name-only` diff restricted to `internal/cli/binary_lag_test.go` is **non-empty** (that file must
change per M1 step 4), so an empty result for `deps.go` is known to mean "unchanged" rather than
"unmatched". One extra line in the same AC; no new criterion, no budget movement.

---

## Coordinator's four suspicion probes — findings

1. **Did REQ-TCR-006 just move the compounding?** Partly yes — see D8. Reduction is real
   (4/3 → 2/2 worst case) and every clause is traced, but the new requirement is itself two-clause
   across two subjects. Reported rather than smoothed over.
2. **Is anything added merely padding?** No — with one clause excepted (D9). I checked each addition
   against "does this change what a run-phase reader must do?": the variadic constraint (`spec.md:173-178`)
   changes the implementation shape; the base-SHA sub-section (`acceptance.md:78-96`) adds a mandatory
   recording; the CI-headroom constraint adds a recorded measurement (`acceptance.md:166-168`); §A.2's
   new paragraph (`spec.md:59-65`) supplies the scope argument the guard rests on; §D.4 is the
   operator-instructed record. The two anti-patterns added at `plan.md:182-184` and `:187-189` each
   name a concrete wrong turn. The artifacts grew, but under a budget whose overrun is about *criteria
   count and file count*, not prose length — and no criterion was added, so the overrun did not widen.
   D9's superlative is the one clause that adds an assertion without adding an obligation.
3. **Is AC-TCR-002b still one decidable criterion?** Yes, conjunctively — it passes iff (i) the
   `session_start.go` diff touches neither the async-branch body nor the join-bound declaration and
   (ii) `deps.go` is absent from the file list. Both sub-checks are binary and the pass condition is
   stated for each (`acceptance.md:74-76`). The Base-SHA sub-section is a *recording obligation*, not a
   third verdict — it produces no pass/fail of its own — so the id carries two checks and one
   record, not "three checks wearing one id". Acceptable. Its real weakness is (ii)'s vacuity, D10 —
   not its multiplicity.
4. **Operator [HARD] wording and Gap integrity.** `spec.md:144` still reads verbatim: "Consequence,
   stated plainly: landing this work leaves observation 1's flake in `develop`." Substance intact, and
   §C's surrounding bullets still refuse the one-class reading (`:142-143`). **No Gap was upgraded into
   a solved claim by the repair.** Re-read of the corrected `reproduction.md`: both corrections are
   marked as dated Corrections rather than silent rewrites (the withdrawn 1-in-5 inference under
   Claim; the caller-inventory scope under Evidence), the Gaps list *gained* a bullet
   ("Observation 2's CI frequency is unknown, not low") rather than losing any, and every Gap that
   bears on the plan is carried into `spec.md` §F — including the ones that argue against this card's
   own urgency. `spec.md:235-238` correctly restates the caller-inventory risk as "one test, one
   production", matching the corrected evidence.
   *One evidence-base hygiene note, outside the SPEC:* `reproduction.md` § Residual risk still says
   "The inventory above covers `_test.go` callers only", which its own Evidence-section correction now
   supersedes. The SPEC does not inherit the stale wording, so this is a note for the lane's record,
   not a SPEC defect.

---

## Category Scores (0.1.1)

| Dimension | iter-1 | iter-2 | Band | Evidence |
|-----------|--------|--------|------|----------|
| Clarity | 0.88 | **0.92** | 0.75-1.0 | §A.3 now names both callers with their roles (`spec.md:74-86`); §A.4 separates mechanism from frequency (`:88-100`); REQ-TCR-002 de-compounded (`:110-112`). Deduction: REQ-TCR-006 still two-subject (D8). |
| Completeness | 0.90 | **0.95** | 0.75-1.0 | Two constraints added (`spec.md:173-178`, `:184-190`), the budget overrun recorded at three surfaces, base-SHA attribution added. Deduction: still no titled WHAT/Scope section — scope is assembled from §A.2 + §C + §D. |
| Testability | 0.90 | **0.92** | 0.75-1.0 | AC-TCR-001's rationale now matches its measurement (`acceptance.md:36-41`); AC-TCR-002b gained a mechanical clause and a mandatory base record; AC-TCR-007 records wall-clock. Deduction: AC-TCR-002b(ii) vacuous-green (D10); (i) remains a recorded human reading. |
| Traceability | 1.00 | **1.00** | 1.0 | `acceptance.md:185-193` maps all six REQs and every AC; REQ-TCR-006 → AC-TCR-002a + AC-TCR-002b verified against the criteria bodies. No orphan, no uncovered REQ. |

Aggregate **0.93** ≥ 0.75. Movement 0.88 → 0.93 is **monotonic upward**; no LEAN STOP condition.

---

## Recommendation

**PASS.** Six of the eight iteration-1 defects are closed outright, D5 is discharged as the operator
instructed with a justification I judge honest on its own terms (the "stay Tier S" scope claim is
true of the plan as written, and §D.4's table earns each of the nine criteria rather than listing
them), and D8 is partially closed — the compounding genuinely fell, but part of it relocated into
REQ-TCR-006, and the repair should not be described as having eliminated it.

Two new minor findings, neither blocking the run phase:

1. **D9** (`spec.md:185-186`) — attribute or soften "already the slowest package in that run".
2. **D10** (`acceptance.md:73`) — give AC-TCR-002b(ii) a positive control so an empty file list cannot
   be produced by a mispointed pathspec.

Both are one-line edits and may be folded into the run-phase commit rather than gating kickoff. No
must-pass criterion fails; the SPEC is fit to enter Implementation Kickoff Approval. If D9/D10 are
applied, bump `version:` to `0.1.2` with a HISTORY row; if they are carried into the run phase
instead, record that decision in `progress.md` §E.2 so the deferral is visible rather than lost.

---

# Appendix — iteration 1 record (verbatim, headings demoted one level)

Preserved as history per the coordinator's instruction. Verdict then: PASS 0.88, defects D1-D8.

## SPEC Review Report: SPEC-TEMPDIR-CLEANUP-RACE-001

Card: t352 · Lane worktree: `.claude/worktrees/t352` · Branch `WT-tempdir-cleanup-race` · HEAD `77b2bcae6`
Iteration: 1/1 (Tier S ceiling = 1)
Verdict: **PASS**
Overall Score: **0.88** (Tier S threshold 0.75 — `spec-workflow.md:140`)

Reasoning context ignored per M1 Context Isolation. Every judgment below rests on the four SPEC
artifacts, `.moai/reports/t352/reproduction.md`, `.moai/reports/t322/verdict.md`, and the source
files re-read on HEAD `77b2bcae6` in this worktree.

Cross-model audit not run: no `audit_model` key exists under `.moai/config/` (grep → no match), and
every artifact under audit is untracked (`git status --short` → three `??` rows), so a diff-fed
backend would return `inconclusive` with nothing measured. Claude-only audit, stated rather than
implied.

---

### Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-TCR-001…005, sequential, no gap, no duplicate,
  uniform 3-digit padding (`spec.md:82,85,91,95,99`).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md` §B) only; the Given-When-Then entries in `acceptance.md` §D.1 are verification-layer and
  are graded under Group 4. REQ-TCR-001 Where-pattern (`spec.md:82`), REQ-TCR-002 unwanted/`shall
  not` (`:85`), REQ-TCR-003 event-driven `When` (`:91`), REQ-TCR-004 ubiquitous (`:95`), REQ-TCR-005
  state-driven `While` (`:99`). No IF/THEN legacy form present.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-14`): `id` matches the **enforced** pattern `internal/spec/lint.go:715`
  `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (the stricter single-segment regex in the schema doc's Field
  Reference table is not the implementation and is not this SPEC's defect); `version: "0.1.0"`
  quoted; `status: draft` in the 8-value enum; `created`/`updated` ISO; `priority: P1`; `phase:
  "v3.2.0"` is a release target, not a prohibited lifecycle-stage token; `lifecycle:
  spec-anchored`; `tags` comma-separated. No rejected snake_case alias.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go: `internal/hook`, `internal/cli`).
  Auto-pass.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — three referenced SPECs, all present, none
  retired/superseded/archived: `SPEC-CLI-TEST-CWD-ISOLATION-001` → `completed`,
  `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` → `implemented`, `SPEC-BINARY-LAG-VISIBILITY-001` →
  `in-progress`. `mcp__moai__spec_audit` (project_root = this worktree) returned one INFO row
  (`EraAutoDetected`, H-3) and no drift finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` over spec/plan/acceptance → `0,
  0, 0`. Auto-PASS (D8-4).
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC directory
  and `.moai/reports/t352/` → exit 1, no match. `research.md` absent by Tier S design.

---

### Category Scores

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.88 | 0.75-1.0 | Every requirement single-interpretation; deductions: REQ-TCR-002 is compound across two subjects (`spec.md:85-89`), and AC-TCR-001's fixture criterion names the wrong quantity (D4). |
| Completeness | 0.90 | 0.75-1.0 | HISTORY `spec.md:23`, WHY `§A:31`, REQUIREMENTS `§B:80`, AC `acceptance.md:9`, four `### Out of Scope — <topic>` H3 blocks with specific bullets (`spec.md:110,119,126,133`), risks `§F:160`. Deduction: no titled WHAT/Scope section (scope is inferable from §A.2 + §C + §D only), and the Tier S artifact set is exceeded (D3). |
| Testability | 0.90 | 0.75-1.0 | Every AC names its deciding command and its passing output (`acceptance.md:39-40,47-48,67-68,80-83,93-94,104-106,113-115,125-126`); no weasel words; AC-TCR-004a demands a verbatim recorded FAIL. Deduction: AC-TCR-002b is a recorded human reading, and its base SHA is not recorded (D2). |
| Traceability | 1.00 | 1.0 | `acceptance.md:141-150` maps every REQ to ≥1 AC and every AC to a REQ or to §D constraints; no orphan, no uncovered REQ. Verified against `spec.md` §B by enumeration. |

Aggregate 0.88 ≥ 0.75 (Tier S). No must-pass failure. Verdict PASS.

---

### Candidate defect the lane raised — AC-TCR-002b: **largely refuted, one residual defect**

Verbatim AC text (`acceptance.md:50-59`):

> ### AC-TCR-002b — the async branch and the join bound are unchanged
>
> - **Given** the diff of this card against its base,
> - **When** `internal/hook/session_start.go` is inspected,
> - **Then** the diff adds an option and a branch condition only: the body of the async branch
>   (`spawnDeferredAdvisoryScans` + the `select` on `joinTimer`) and the value of
>   `deferredScanJoinBound` are unmodified.
> - **Command:** `git diff origin/develop...HEAD -- internal/hook/session_start.go`
> - **Passing output:** no hunk touching the `select { case advisory := <-advisoryCh: ... }` block or
>   the `deferredScanJoinBound` declaration. Judged by reading the diff, and the reading is recorded.

The suspicion assumes the two-dot semantics. The AC uses the **three-dot** form, which diffs from
`merge-base(origin/develop, HEAD)` — the branch's fork point — not from the moving tip. Measured now:

```
$ git merge-base origin/develop HEAD
77b2bcae6a3cea087e9f8e7b102eb7320575a582
$ git rev-parse --short origin/develop
947f5cffb
```

`origin/develop` has advanced to `947f5cffb`, three commits past the evidence base, and the
merge-base is nevertheless still `77b2bcae6` — exactly the SHA the evidence base cites. It stays
there as this card commits, because develop's new commits are not in HEAD's ancestry. The premise
"measures a different baseline on every run" does **not** hold for `...`. (Today the two forms even
agree: `git diff --stat origin/develop..HEAD -- internal/hook/session_start.go` is empty, so
develop's three new commits did not touch that file — the two-dot hazard is latent here, not active.)

Two residual weaknesses survive, and one is a real defect (D2 below):

1. The base the AC actually judged against is never recorded. `acceptance.md:5` obliges recording
   the **HEAD** SHA only. A reading of a diff whose base is unattributed is an unattributed claim
   (`verification-claim-integrity.md` §2) — and the base here is a *derived* value, so it cannot be
   reconstructed later once refs move.
2. Merging `develop` into the lane branch mid-card moves the merge-base forward. That is the
   semantically correct baseline for "this card's diff", but it silently changes what the AC
   compared, with nothing in the record to show it moved.

Judgment: **the lane misread the operator (`...` vs `..`); the AC is materially safer than reported.
The defect that remains is the missing base-SHA attribution, not a moving reference.**

---

### Audit emphasis findings

1. **Is the RED evidence real? — YES, the AC forces the observation.** `acceptance.md:80-83` requires
   the mutation to be applied, the command run, and "*a `FAIL` line and exit status 1, with the
   verbatim failure text recorded in `progress.md` §E.2 together with the mutation's diff*" as this
   AC's own passing output. An assertion that the mutation "would" fail cannot satisfy it: the
   passing artifact IS the observed failure text plus the mutant diff. `acceptance.md:95-96` pairs
   004a/004b and states why neither alone suffices; `plan.md:86-88` repeats the obligation; `plan.md:111`
   forbids grep-for-the-fixed-form as a substitute. No defect.
2. **Is AC-TCR-001's fixture strong enough? — YES on padding, with an imprecision in how the
   criterion is phrased (D4).** `acceptance.md:28-31` pads "to at least the larger of" the measured
   rows, i.e. 8000 files (reproduction: 8000 → 223 ms late). `spec.md:185-187` and `plan.md:109-110`
   both state *why*: the `nFiles=0` row flipped between bases, so an unpadded fixture would itself be
   a flake. The requirement to pad is therefore stated and justified, not assumed.
3. **Does AC-TCR-003 overclaim? — NO.** `acceptance.md:69-70`: "*50 local iterations passed on the
   pre-fix tree too. This AC is a non-regression check, **not** the evidence that the defect is
   fixed — AC-TCR-004 is.*" Exactly the required disclaimer.
4. **Is the mechanism rung justified? — YES.** `plan.md:42-47` evaluates four rungs and states a
   falsifiable rejection reason plus the *cost of taking it* for each of the three rejected ones
   (rung 3's rejection — a stat cannot close a window where the directory still exists while
   `RemoveAll` walks it — is technically correct). `plan.md:49-55` (§D.2) names the tension
   explicitly: "*Rung 1 adds production API surface for what is presently a test-only need. That is a
   real cost…*" and bounds the argument ("*If the option were introducing a new code path, the
   balance would go the other way*"). Not waved away.
5. **Scope integrity — YES, the operator's [HARD] wording is present.** `spec.md:110-117`, in
   particular `:117`: "*Consequence, stated plainly: landing this work leaves observation 1's flake
   in `develop`.*" `spec.md:115-116` refuses the one-class reading. No requirement or AC touches
   `internal/graph`; `acceptance.md:165` restates the non-establishment; `plan.md:113-114` makes a
   diff touching `internal/graph` an anti-pattern. Requirement satisfied.
6. **Cross-platform claim — NO overclaim.** `acceptance.md:116-117`: "*`go vet` proves compilation
   only, not behaviour … Behaviour on those platforms is CI's verdict.*" Correct reading.

---

### Source-claim verification (every factual citation in §A re-read on `77b2bcae6`)

| SPEC claim | Verified |
|---|---|
| `session_start.go:240` async branch on `deferredScansAsyncEnabled()` | TRUE (`:240` `if deferredScansAsyncEnabled() {`) |
| `:608-611` advisory sent first, `runMXColdStartScan` after | TRUE (`:608 resultCh <- advisory`, `:609-611` scan) |
| `:1606-1613` writes `<projectDir>/.moai/state` | TRUE (`:1606 stateDir := filepath.Join(projectDir, ".moai", "state")`, `:1613 mgr.Write`) |
| `deferredScanJoinBound` = 250 ms | TRUE (`session_start.go:1418`) |
| `var deferredScansAsync = true` | TRUE (`session_start.go:1460`) |
| `internal/hook/main_test.go:45-50` flips it to false for that binary | TRUE (`TestMain` at `:45`, flip at `:47`, `goleak.VerifyTestMain` at `:50`) |
| `internal/cli/binary_lag_test.go:57` passes a `t.TempDir()` as `ProjectDir` | TRUE (`:56` TempDir, `:57` `hook.NewSessionStartHandler(nil).Handle`) |
| `internal/cli/main_test.go:213-245` residue guard, different locus | TRUE (`TestMain` at `:213`, watches `entryWD/.moai`, `RESIDUE GUARD FAIL` at `:232`) |
| `session_start.go:196-215` documents the input-lag rationale | TRUE (`:196-215` is that comment block) |
| plan §D.3: `binaryLagAdvisory` is read-only | TRUE, and this audit verified the premise rather than accepting it: `internal/binlag/binlag.go:101,111` are `git rev-parse HEAD` / `git merge-base --is-ancestor`; no write call exists in the package |
| §A.3 "the **single** cross-package caller" | **FALSE as phrased** — see D1 |

---

### Defects Found

**D1 — `spec.md:65-67` — "the single cross-package caller" omits a production caller.**
`internal/cli/deps.go:221` is `deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))`
— a non-test cross-package call site. The parenthetical grep in the same sentence discloses the
`--include='*_test.go'` scope, which mitigates the claim but does not repair it, and no artifact
(`spec.md` §A.3/§D, `plan.md` §E, `acceptance.md`) names `deps.go:221` anywhere. This matters
because the chosen rung changes the signature of the exported constructor that production line
calls: `internal/hook/session_start.go:41` is `func NewSessionStartHandler(cfg ConfigProvider)
Handler` — non-variadic today, while `acceptance.md:33` presumes `NewSessionStartHandler(nil,
<option>)`. A variadic addition keeps `deps.go:221` compiling, but that is a constraint no
requirement states.
Required fix: correct §A.3 to "the single cross-package **test** caller", name
`internal/cli/deps.go:221` as the production caller, and add a one-line constraint to §D that the
existing production call site must keep compiling unchanged (already covered mechanically by
AC-TCR-006 + AC-TCR-007, which need only be cited).
Severity: **major** — Class: **blocking** — Label: SHOULD-FIX

**D2 — `acceptance.md:5` + `acceptance.md:57` — AC-TCR-002b records no base SHA.**
The preamble obliges recording the HEAD SHA only; the AC's baseline is a *derived* value
(`merge-base(origin/develop, HEAD)`, measured today as `77b2bcae6`) that cannot be reconstructed
after refs move. The lane's stronger claim (a branch name measuring a different baseline each run)
is refuted above — the three-dot form pins the fork point — but the attribution gap is real
(`verification-claim-integrity.md` §2).
Required fix: add to AC-TCR-002b's passing output "*record `git merge-base origin/develop HEAD` next
to the reading*", and state that a mid-card merge of `develop` moves that base and must be noted.
Severity: **major** — Class: **blocking** — Label: SHOULD-FIX

**D3 — `spec.md:166` (§F) — the carried "1 failure in 5 runs" figure belongs to the excluded
observation.** `.moai/reports/t322/verdict.md:281` attributes "*now 1 of 5 post-landing runs*" to
`TestGitDiffNameCount_Predicate` — observation 1, which `spec.md` §C explicitly excludes. The same
report's line 281 row for `TestBinaryLag_OneSeamServesBothSurfaces` reads count `1`, "**appeared**"
— one CI observation, no frequency. `spec.md:165-166` carries the 1-in-5 figure into a SPEC scoped
to observation 2 without naming which test it describes, so the default reading attributes another
test's frequency to this card's target. §A.4's "*CI fails intermittently*" (`spec.md:75-76`) rests on
the same conflation: for observation 2 the CI record is a single failure, and "intermittent" is an
inference from the mechanism, not a measured frequency.
Required fix: name the test in §F ("*the 1-in-5 figure describes `TestGitDiffNameCount_Predicate`,
observation 1; observation 2 was observed in CI once*") and soften §A.4 accordingly.
Severity: **major** — Class: **blocking** — Label: SHOULD-FIX

**D4 — `acceptance.md:29` — the fixture criterion names a quantity the evidence did not measure.**
"*padded with enough files that `runMXColdStartScan`'s `ScanDir` measurably exceeds the 250 ms join
bound*". The reproduction measured lateness relative to **`Handle`'s return**, which occurred at
~60 ms (`Handle=59.59ms` at `nFiles=8000`) because the advisory arrives long before the 250 ms
bound — the bound is never reached in these runs. The operative condition is "the scan finishes
after `Handle` returns", not "the scan exceeds 250 ms". The padding number the AC then fixes (8000)
is correct and well-justified, so the guard is not weakened; the stated rationale is.
Required fix: restate the Given as "padded so the scan reliably finishes after `Handle` returns —
measured 223 ms late at 8000 files".
Severity: **minor** — Class: **blocking** (a stated rationale that contradicts the cited
measurement) — Label: SHOULD-FIX

**D5 — `acceptance.md:13-21` + artifact set — Tier S budget exceeded on two axes.**
Nine acceptance criteria (001, 002a, 002b, 003, 004a, 004b, 005, 006, 007) against the Tier S
ceiling of 8 (`spec-workflow.md:148`), and a separate `acceptance.md` where Tier S prescribes
"*2 files: spec.md + plan.md (AC inline in spec.md §3)*" (`spec-workflow.md:140`). The rule states
that exceeding a ceiling "*is a signal to tier up or to split the SPEC, not to relax the budget*".
Required fix: either fold a pair (002a+002b, or 004a+004b, each already a single concern with two
cells) into one AC with two recorded observations, or reclassify `tier: M` — whose 3-file artifact
set this SPEC already matches. Do not simply delete a criterion.
Severity: **minor** — Class: **blocking** (a criterion this doctrine actually states) — Label:
SHOULD-FIX

**D6 — `spec.md:143-144` (§D) / `plan.md` §E — the guard's added CI cost is unstated.**
The guard creates an 8000-file fixture inside `internal/cli`, a package whose local single run is
already measured in the hundreds of seconds, and CI runs `go test -race -count=1 ./...` with **no
`-timeout` flag** (`.github/workflows/ci.yml:238`), i.e. the 10-minute per-package default. The
added cost is probably small (one fixture, `-count=1` in CI), but the SPEC asserts only that "*the
change is expected to be small*" in LOC terms and says nothing about runtime headroom in the package
the guard lands in.
Required fix: add one §D constraint — the guard's fixture build plus `Handle` must add well under
the package's remaining default-timeout headroom, and AC-TCR-007's wall-clock is recorded in
`progress.md` §E.2 so the cost is measured rather than assumed.
Severity: **minor** — Class: **optional** — Label: MINOR

**D7 — `spec.md:82` (REQ-TCR-001) — "every durable write" is universal; only one writer is named.**
The requirement quantifies over all durable writes into `ProjectDir`, while §A.2 establishes exactly
one (`runMXColdStartScan`) and `plan.md:68-72` exempts the second known post-`Handle` goroutine
(`binaryLagAdvisory`) on a read-only premise. This audit verified that premise true
(`internal/binlag/binlag.go:101,111`), and AC-TCR-001's whole-directory entry-set comparison
empirically covers any writer, so the gap is closed in practice — but the SPEC never says that the
entry-set comparison is what generalizes the requirement, and the plan's exemption is asserted
rather than cited.
Required fix: one sentence in §A.2 or plan §D.3 — "the guard compares the directory's whole entry
set, so it covers any writer, not only the MX index; `binaryLagAdvisory`'s goroutine performs only
`git rev-parse` / `git merge-base` (`internal/binlag/binlag.go:101,111`) and writes nothing."
Severity: **minor** — Class: **optional** — Label: MINOR

**D8 — `spec.md:85-89` (REQ-TCR-002) — compound requirement across two subjects.**
One requirement carries four testable claims and shifts subject mid-sentence ("the session-start
handler" → "deferred advisory scanning" → "the production default of the existing execution-mode
seam"). GEARS-conformant as an unwanted-pattern entry, and fully traced (002a/002b), so this is a
splittability preference, not a correctness defect — recorded so it is not mistaken for an oversight.
Severity: **minor** — Class: **optional** — Label: MINOR

---

### Regression Check

N/A — iteration 1 of a Tier S ceiling of 1. No prior plan-audit report exists for this SPEC
(`.moai/reports/plan-audit/SPEC-TEMPDIR-CLEANUP-RACE-001-*` absent).

---

### Recommendation

PASS. The must-pass firewall is clean on all seven criteria, the aggregate 0.88 clears the Tier S
threshold of 0.75, and each of the six audit-emphasis probes the operator named came back supported
by quotable text — the RED observation is forced rather than asserted, the fixture is padded with a
stated reason, AC-TCR-003 disclaims itself as non-regression, the mechanism ladder states the cost
of each rejected rung and the tension of the chosen one, the observation-1 consequence is stated in
the operator's required plain form, and `go vet` is explicitly held to compilation only.

Five defects are blocking-class and all five are wording repairs on existing artifacts — none
requires re-deciding the design:

1. **D1** — repair `spec.md:65-67` and name `internal/cli/deps.go:221`.
2. **D2** — make AC-TCR-002b record `git merge-base origin/develop HEAD` alongside its reading.
3. **D3** — attribute the 1-in-5 figure to `TestGitDiffNameCount_Predicate` in `spec.md:166` and
   soften `spec.md:75-76`.
4. **D4** — restate `acceptance.md:29`'s Given in terms of "after `Handle` returns", not the 250 ms
   bound.
5. **D5** — resolve the Tier S budget overrun by folding one AC pair or by reclassifying `tier: M`.

D6-D8 are optional and left to the orchestrator's discretion. Bump `version:` to `0.1.1` and
`updated:` when the repairs land.
