# SPEC Review Report: SPEC-AUDIT-PARTICIPANT-COUNT-001

Iteration: 2/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (Tier M PASS threshold 0.80; iter 1 was FAIL 0.75 — monotonic increase, no STOP signal)

Reasoning context ignored per M1 Context Isolation. Audited from the artifacts, the iter-1
report (defect delta + regression scope), and the subject code only.

## Tree / HEAD attribution

Worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t284`, branch
`WT-audit-participant-count`, HEAD `8c1d911df` (`git rev-parse HEAD`, re-read this run).
HEAD is a develop integration merge (`Merge branch 'WT-init-harness-prompt' into develop`);
the SPEC artifacts (v0.2.0) and the t284 reports are **untracked working-tree files**
(`git status --porcelain` → `?? .moai/specs/SPEC-AUDIT-PARTICIPANT-COUNT-001/`,
`?? .moai/reports/t284/`) sitting on top of that HEAD.

**Measurement-tree validity.** The probe evidence (`.moai/reports/t284/participant-count-probe.log`,
the §A.2 premise probe) was measured at iter-1 time on tree `64bba61aa`. The six
convergence-relevant sources (`mcp_convergence.go`, `mcp_convergence_test.go`,
`multi_review_gate.go`, `multi_review_gate_test.go`, `codex_verdict_divergence_test.go`,
`mcp_audit_multi.go`) are **byte-identical between `64bba61aa` and `8c1d911df`**
(`git diff --stat 64bba61aa..HEAD -- <six files>` → empty, rc=0), so every participant-count
figure in the artifacts transfers to the tree under audit. Spot re-derivations from the test
inputs themselves (below) corroborate the log independently of trusting it.

**Audit mode.** `grep -rn audit_model .moai/config/` → no match: the project config requests
no cross-backend second opinion, so this is a Claude-only audit (same path as iter 1). No
`mcp__moai__*` audit backend was consulted.

## Required verification re-measured (dispatch items 1-3)

1. **Site count.** `grep -rn DisagreementFlag --include='*_test.go' internal/cli` → **12 sites
   across 3 files**, at exactly the coordinates plan.md §B lists:
   `mcp_convergence_test.go` `:75,:94,:113,:131,:163,:180,:201,:457` (8),
   `multi_review_gate_test.go` `:65,:80,:95` (3), `codex_verdict_divergence_test.go:70` (1).
   Matches plan.md §B inventory row and §E M2 ("The 12 test sites of §B").

2. **All 8 probe rows vs. the (a)-semantics expectations, every surface.**

   | Probe row | Measured count / flag | Expected under (a) | Surfaces stating it | Match |
   |---|---|---|---|---|
   | AllRequiredPass | 3 / false | 3 / non-nil false | plan §E M2, AC-APC-005 | ✓ |
   | AllRequiredFail | 2 / false | 2 / non-nil false | plan §E M2, AC-APC-005 | ✓ |
   | RequiredFailWithInconclusive | 1 / false | **1 / null** (moved) | AC-APC-002 coverage, AC-APC-005 exclusion note, plan §E M2, plan §H risk row | ✓ consistent on every surface |
   | RequiredSplit | 3 / true | 3 / non-nil true | plan §E M2, AC-APC-005, spec §C row 4 | ✓ |
   | AdvisoryOnlyConflict | 3 / true | 3 / non-nil true | plan §E M2, AC-APC-005 | ✓ |
   | DisagreementAdvisoryNotBlock | 2 / true | 2 / non-nil true | plan §E M2, AC-APC-005 | ✓ |
   | NoRequiredBackends_VacuousPass | 2 / false | 2 / non-nil false | plan §E M2, AC-APC-005, acceptance §B edge | ✓ |
   | SurfacesSignalDivergence | 1 / true | **1 / non-nil true** (carve-out) | AC-APC-008, spec §C row 5, plan §E M2 twelfth-site routing, §D witness table | ✓ |

   Independent arithmetic spot-checks from the test inputs (not from the log):
   `TestConverge_AllRequiredPass_Case1` (`mcp_convergence_test.go:66-70`, three required
   passes → 3), `TestConverge_AllRequiredFail_Case2` (`:86-89`, two required fails → 2),
   `TestConverge_RequiredFailWithInconclusive_Case2` (`:105-108`, claude required=fail +
   codex required=inconclusive → 1 under REQ-APC-002), `TestConverge_RequiredSplit_Case3`
   (`:122-126`, three required → 3), `TestConverge_SurfacesSignalDivergence_WithoutBlocking`
   (`codex_verdict_divergence_test.go:56-59`, claude required=pass + codex
   required=inconclusive-with-note → 1). All match.

3. **MP-1..MP-7** — see Must-Pass Results.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-APC-001..005`, sequential, no gaps, no
  duplicates, consistent padding (`grep -o 'REQ-APC-[0-9]*' spec.md | sort | uniq -c` →
  001×1, 002×1, 003×4 (cross-references), 004×1, 005×1).
- **[PASS] MP-2 GEARS format compliance** — judged against the `REQ-XXX` requirement layer in
  spec.md §B. 001/002 ubiquitous, 003 state-driven + unwanted with a compound `unless`-clause
  (PASS-equivalent per the GEARS compound rule), 004 state-driven, 005 unwanted. The
  Given/When/Then bodies in `acceptance.md` are the verification layer, graded under Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct
  types (spec.md:1-15): `version: "0.2.0"` (quoted semver), `status: draft`, ISO dates,
  `priority: P1`, `phase: "v3.1.5 target"` (a release target, not a prohibited stage name),
  `lifecycle: spec-anchored`, `tags` comma string, plus `tier: M`. No rejected snake_case
  alias. **Version consistency**: spec.md / plan.md / acceptance.md all `0.2.0`; HISTORY
  carries 0.1.0 and 0.2.0 rows. Artifact statelessness holds (plan/acceptance carry no
  `status:`).
- **[N/A] MP-4 language neutrality** — single-language SPEC scoped to `internal/cli` (Go).
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — full extraction
  (`grep -rhoE 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` over the SPEC dir) finds only
  `SPEC-CODEX-VERDICT-SYNTH-001` and `SPEC-AUDIT-MULTI-MODEL-001` (plus self-references);
  both resolve and both read `status: completed` — neither retired/superseded/archived. No
  BLOCKING finding, no missing-SPEC SHOULD finding.
- **[N/A→PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` over the three
  artifacts → `0 0 0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC
  directory → no match (rc=1). `research.md` does not exist (Tier M does not require it).

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 | Requirements are individually precise — REQ-APC-003's carve-out scopes its `false`-forbidden clause "without exception" correctly to `false`, and its `null` obligation via the unless-clause. Deducted for D7: AC-APC-002/003's Given clauses ("any input whose participant_count is 0 or 1") over-universalize what REQ-APC-003 excepts, leaving two conflicting instructions a reasonable engineer resolves only by consulting the coverage lists and §D. Minor additional: D8 (spec §C's "(existing case #3)" provenance pointer). |
| Completeness | 1.0 | 1.0 | HISTORY / §A context / §B requirements / §C contract / §D traceability / §E exclusions (four `### Out of Scope — <topic>` H3s, each with specific bullets) all present and substantive; frontmatter complete; site inventory complete (12/3, re-measured); the D2 decision recorded with its rejected alternative; mutation-witness record complete for all 8 criteria; §G's 12 documentation rows re-measured exact at the current tree. |
| Testability | 0.75 | 0.75 | Criteria are binary and aggressively anti-vacuous (AC-APC-002 names the vacuous `!= nil && *ptr` idiom; AC-APC-003 requires member-presence; AC-APC-004's vacuous byte-inequality clause is gone, its rationale now states why). Deducted for D7: under the literal universal reading AC-APC-002/003 and AC-APC-008 are jointly unsatisfiable — measurable, but only with a scope interpretation the criteria text itself does not state. |
| Traceability | 0.90 | 0.75→1.0 | Every REQ has ≥1 AC and every AC traces to an existing REQ (001→001/002; 002/003/008→003; 004→AC-004+001; 005→AC-005; 006/007→005); no orphans either direction; §D marks non-witnesses rather than padding. Deducted because AC-APC-002/003 assert more than their traced REQ-APC-003 permits (the carve-out exception) — a criterion-vs-requirement content mismatch. |

Aggregate (arithmetic mean): (0.75 + 1.0 + 0.75 + 0.90) / 4 = **0.85** ≥ 0.80.

## Defects Found

**D7. AC-APC-002 and AC-APC-003 universalize over every 0-or-1 input, contradicting
AC-APC-008 and REQ-APC-003's carve-out — introduced by the repair** —
`acceptance.md` AC-APC-002 (:42-46), AC-APC-003 (:64-68), AC-APC-005 (:116-118),
§D (:211-213); `spec.md` §C (:145); `internal/cli/codex_verdict_divergence_test.go:56-59`
— Severity: **major** — Class: **blocking**

AC-APC-002: "**Given any input whose participant_count is 0 or 1** … **then** the returned
`DisagreementFlag` **is nil**". AC-APC-003: same universal Given, "**that member's value is
JSON `null`**". But the carve-out input (`SurfacesSignalDivergence`: claude required=pass +
codex required=inconclusive carrying a synthesis note; measured count 1) must, per
AC-APC-008 and REQ-APC-003, produce a **non-nil `true`** — "the one below-2 case where
`null` is forbidden". Under the literal universal reading, no implementation satisfies all
eight criteria: the chosen option-(a) design fails AC-APC-002/003 on the divergence input;
an option-(b) implementation fails AC-APC-008.

The over-scope is not just inferable — the document asserts it three times: AC-APC-002's
Given ("any input"), AC-APC-005's exclusion note ("AC-APC-002 and AC-APC-003 cover it (**both
bind every 0-or-1 input**…)"), and §D's independence analysis ("an option-(b) implementation
(unconditional `null` below 2) **satisfies AC-APC-002 and AC-APC-003**" — which is true only
if they bind every 0-or-1 input, the very scope the chosen design violates on one of them).

**Why major, not critical (the iter-1 D1 distinction).** In iter-1's D1 the colliding case sat
*inside* AC-APC-005's own enumerated must-pass list — unavoidable by any reading. Here the
collision input is absent from AC-APC-002's stated coverage list (§A.2 rows 1/3, the DQ-2
refusal, `RequiredFailWithInconclusive`), and §D's worked examples demonstrate the intended
three-witness coexistence, so the run-phase owner following the document's own construction
writes tests that all pass. The defect is that the *normative* Given/Then text forbids what
the design does — a future auditor or maintainer applying AC-APC-002 literally (e.g. a
sync-phase check over "every 0-or-1 input") flags the correct implementation red.

**Required fix** (one clause, no expectation value changes; owner manager-spec, may be applied
before or at Implementation Kickoff): scope the two Given clauses — AC-APC-002: "any input
whose `participant_count` is 0 or 1 **and which produced no intra-backend synthesis
divergence**", AC-APC-003: same qualifier — and align AC-APC-005's note ("both bind every
0-or-1 input **that produced no intra-backend divergence; AC-APC-008 owns the divergence
input**"). §D's option-(b) sentence needs no change once the Given clauses carry the
qualifier.

---

**D8. spec.md §C's "(existing case #3)" cites a numbering the document never introduces** —
`spec.md` §C table row 4 — Severity: **minor** — Class: **optional**

The parenthetical resolves only against the predecessor corpus's Case #N numbering (visible
in `TestConverge_RequiredSplit_Case3_AC_AMM_008` and `mcp_convergence.go:153` "Case #2 / #3"),
which §C does not name. The row's own values (`2+`, `true`) are unambiguous and correct. A
one-word fix ("existing derivation Case #3 of SPEC-AUDIT-MULTI-MODEL-001's policy table") if
touched anyway; not worth its own revision round.

## Regression Check (iter-1 defects D1-D6)

- **D1 (AC-APC-005 contradicts REQ-APC-003 + AC-APC-002 on RequiredFailWithInconclusive) —
  RESOLVED.** AC-APC-005 now enumerates exactly the six measured ≥2 cases (AllRequiredPass 3,
  AllRequiredFail 2, NoRequiredBackends_VacuousPass 2 as `false`; RequiredSplit 3,
  AdvisoryOnlyConflict 3, DisagreementAdvisoryNotBlock 2 as `true` — all counts re-verified
  against the probe and the test inputs), and its exclusion paragraph names
  `RequiredFailWithInconclusive` (count 1) as owned by AC-APC-002/003 with the false→null
  move stated as deliberate. plan.md §E M2 states the same partition with the same counts —
  both surfaces fixed together as iter 1 required.
- **D2 (REQ-APC-003 suppresses a genuinely-true intra-backend divergence; case absent from
  every inventory) — RESOLVED.** The interaction is decided and recorded: REQ-APC-003 carries
  the carve-out (divergence-observed sub-2 input keeps `true`); spec.md §B carries the
  option-(a) rationale (three numbered reasons) with option (b) recorded as the **rejected
  alternative** and its rejection reasons; §E's exclusion is restated so it now covers what
  the SPEC actually does ("Its observable effect on `disagreement_flag` is preserved, not
  changed… The intended observable consequence of this SPEC for that pass is therefore *nil*");
  the case now appears in plan.md §B (inventory row), §E M2 (twelfth-site routing through the
  carve-out), AC-APC-008 (dedicated criterion), spec.md §C (fifth row), and the §D
  mutation-witness table (witness). progress.md §E.1 records the decision provenance
  ("decided by the orchestrator and surfaced at the kickoff gate"). Iter-1's recording
  requirement is fully met.
- **D3 (§A.3 byte-identical premise false; AC-APC-004 leading clause vacuous) — RESOLVED.**
  §A.3 now states the documents are **not** byte-identical (447 vs 235 bytes, matching the
  probe log line) and that the four *derived* fields are what is identical; it also correctly
  narrows the claim against case 1 vs case 2 (distinguishable via `fail_open_backends`).
  AC-APC-004 asserts only the two derived positions, with the rationale explicitly stating
  byte inequality already holds at HEAD and would assert nothing.
- **D4 (11-site undercount; third file omitted) — RESOLVED.** §B now reads 3 files / 12 sites
  with exact line numbers; re-measured grep returns 12 across exactly those files and lines;
  M2 says 12 and routes the twelfth site through the D2 carve-out rather than the mechanical
  pass.
- **D5 (§A.4 "lines 1-38" off by one) — RESOLVED.** §A.4 cites "lines 1-39";
  `@MX:SPEC: SPEC-AUDIT-MULTI-MODEL-001` is line 39, `package cli` is line 40 (re-read this
  run).
- **D6 (§A.1 pass-3 coordinate stops at :194) — RESOLVED.** §A.1 row 3 now cites `:188-197`
  and names the inner coordinates (`:194` call, `:195-197` flag assignment) — all re-verified
  against source.

No stagnation: all six defects moved to resolved with evidence.

## Verified-clean (probed this run, no defect found)

Recorded so their absence from the defect list is not silence:

- **Every other cited coordinate resolves at the CURRENT tree** (which matters — the develop
  absorb changed `mcp_server.go` by 2 lines): `converge` at `mcp_convergence.go:140` ✓;
  required-set split `:168-170` ✓; advisory-conflict `:171-185` ✓; pass 3 `:188-197` ✓;
  DQ-2 refusal block `:526-533` ✓; gate-`off` `continue` `:571-573` ✓; `pbv()` at
  `mcp_convergence_test.go:30` ✓; gate-off test `:419` ✓;
  `TestConverge_SurfacesSignalDivergence_WithoutBlocking` at `codex_verdict_divergence_test.go:55`
  with the `:70` assertion ✓.
- **multi_review_gate.go claims**: the only result-field reads are `OverallVerdict` (`:83`)
  and `ResidualRiskNote` (`:97`, inside `blockReason` at `:96`); `loadConvergenceResult` at
  `:110` does a plain `json.Unmarshal` (`:119-121`) so a recorded `false` decodes into a
  non-nil `*bool` exactly as plan.md §D and AC-APC-007 claim; the gate's only BLOCK path is
  `overall_verdict == fail` (`:83-88`) ✓.
- **mcp_server.go claims**: the two `WithOutputSchema[ReviewOutput]()` sites sit at `:259`
  and `:400` (other tools); the `audit_multi` registration block with no output schema is at
  `:405-425` (comment at `:406`) ✓ — coordinates survived the develop absorb.
- **plan.md §G occurrence counts: all 12 rows re-measured exact at HEAD `8c1d911df`** —
  multi-model-audit 1/1/1/1, autonomous-loops ko=0 with en/ja/zh=3 (substantiating spec.md
  §E's recorded pre-existing 4-locale parity gap, which stays correctly out of scope),
  SKILL.md 5+5, review.md 1+1, template mirrors included ✓.
- **§A.1's "0 hits" claim** — `grep -ci participant` over the three convergence surfaces →
  `0 0 0` ✓.
- **AC-APC-001's eight-row table** is arithmetically consistent with REQ-APC-002's definition
  (rows g and h pin the two boundaries; row c matches §A.2 case 1; row f matches case 2) ✓.
- **Mutation-witness record**: 8 rows for 8 criteria, witnesses 002/003/008 matching
  plan.md §F and spec.md §D; the two independence worked examples (`omitempty`;
  option-(b)) are present ✓.
- **Sibling fixes beyond D1-D6 are sweep, not scope creep**: the §C fifth row (required by
  D2's decision), the version bumps + HISTORY row (required by the repair convention), the
  per-case count annotations in M2/AC-APC-005 (the D1 fix itself), plan.md §D's carve-out
  mention and §H's "two movements are deliberate" row (necessary consequences), and the
  §D witness-table extension to 8 rows (required by AC-APC-008). Nothing exceeds the three
  card items; §E still forbids the obvious creep. Tier M budget: 5/16 REQ, 8/16 AC ✓.
- **progress.md §E.1** is audit-ready with counts matching the artifacts (5 REQ, 8 AC,
  card t284) and carries the repair provenance ✓.
- **REQ-APC-004 interplay**: at 2+ participants the three-pass derivation (including any
  divergence note) produces the boolean unchanged — no conflict with the carve-out, which is
  scoped below 2 only ✓.

## Recommendation

**PASS-WITH-DEBT at 0.85** against the Tier M threshold of 0.80, iteration 2 of 2. All
must-pass criteria hold; all six iter-1 defects are resolved with evidence; the repair's
arithmetic is correct on every surface I re-measured. One repair-introduced defect (D7)
remains, and it is the debt:

1. **Before Implementation Kickoff (or at it)**: re-delegate the one-clause D7 scoping fix to
   manager-spec — AC-APC-002 Given, AC-APC-003 Given, and AC-APC-005's "bind every 0-or-1
   input" sentence. No expectation value changes; no design decision reopens; no iter 3 is
   needed (and none is available — Tier M ceiling is 2).
2. **D8** may be folded into the same edit or left; optional.
3. **Skip-eligibility note**: this verdict is PASS-WITH-DEBT, not PASS, and the D7 fix will
   change `acceptance.md` anyway — both independently mean the Phase 1 Plan Audit Gate's skip
   conditions do not hold; expect (and accept) a hash-invalidated cache. The Kickoff human
   gate is of course unaffected either way.
4. If the orchestrator prefers a strictly clean PASS over carrying the debt, the fix above is
   the entire distance — it does not warrant the FAIL escalation paths (PASS-with-debt scope
   reduction / user override) for a wording-scoping clause whose intended reading the
   document already demonstrates in its own §D.
