# SPEC Review Report: SPEC-DECISION-AUTHORITY-001

Iteration: 1/3 (Tier L ceiling = `harness.yaml` `plan_audit_tier_ceilings.L: 3`)
Auditor: plan-auditor (independent) — tree `.claude/worktrees/t692`, HEAD `6732d1461`,
branch `WT-judgment-authority`, artifacts measured as committed at `6732d1461`
(authoring baseline `62fbd6baf`, re-verified reachable in this tree).
Audited artifact set (Tier L): spec.md, plan.md, acceptance.md, design.md, research.md, progress.md.

**Verdict: FAIL**
**Overall Score: 0.85** (harmonic mean of the four dimensions; the Tier L numeric
threshold 0.85 is met, but the verdict fails on blocking-class defects — see
Defects D1-D4. The threshold is a necessary condition, not a sufficient one.)

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-DA-001..REQ-DA-021: 21 unique ids,
  sequential, no gaps, no duplicates, uniform zero-padding (`grep -o 'REQ-DA-0[0-9][0-9]'
  spec.md | sort -u` = exactly 21; spec.md:183-294). AC-DA-001..AC-DA-019: 19 unique,
  sequential (acceptance.md:21-39). Both within the Tier L ceilings (25/25).
- **[PASS] MP-2 GEARS format compliance** — all 21 REQ entries match GEARS patterns and
  carry pattern annotations: Ubiquitous (REQ-DA-001 spec.md:183, 004, 006, 008, 009, 010,
  014, 017, 018, 019, 020, 021), Where-capability-gate (REQ-DA-002 :186, 003 :190, 007
  :209, 013 :241), While-state-driven, incl. the PASS-equivalent [While][When] compound
  (REQ-DA-005 :200, REQ-DA-012 :236), When-event (REQ-DA-011 :229, 015 :252, 016 :257).
  No informal "should/may" normative text in the requirement layer. MP-2 judged against
  the REQ-XXX layer only; the AC layer is graded under Group 4 per the two-layer scope.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with
  correct types at spec.md:2-16 (`id` SPEC-DECISION-AUTHORITY-001, quoted semver
  `"0.1.0"`, `status: draft` (valid enum), `created`/`updated` 2026-09-13 ISO, `author`,
  `priority: P2`, `phase: "v3.2.0 target"` (release label — not a prohibited stage
  value), path-like `module`, `lifecycle: spec-anchored`, comma-separated `tags`). No
  rejected snake_case aliases. Optional fields well-formed (`tier: L`, `issue_number:
  1683`, `depends_on: [SPEC-JUDGMENT-FIRST-MODE-001]`). Sibling artifacts stateless on
  the status axis (no `status:` in plan/acceptance/design/research; progress.md has no
  frontmatter) per spec-frontmatter-schema.md § Artifact Statelessness.
- **[N/A] MP-4 Section 22 language neutrality** — N/A: the SPEC is single-language-scoped
  (Go config internals + this project's doctrine text); it covers no multi-language
  tooling. Its own REQ-DA-021 (spec.md:291-294) additionally mandates 16-language
  neutrality for all template-side text. Auto-passes per the MP-4 single-language rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — SPEC-ID sweep (`grep -Eo
  'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' spec.md | sort -u`) yields exactly two ids, one being
  this SPEC itself. The only external reference, SPEC-JUDGMENT-FIRST-MODE-001, exists at
  `.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/spec.md` with `status: completed` (verified
  by grep this run) — not in {retired, superseded, archived}, so no reconciliation is
  owed. The t401 §F exclusions are independently carried/re-justified in this SPEC's §F
  (see Exclusion Fidelity below). No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` on spec.md = 0 →
  auto-PASS per D8-4. The only Go change is inside `internal/config` (portable); plan.md
  §E.2 nevertheless carries the GOOS=windows build check.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'` across the whole
  SPEC directory: no matches (exit 1). No unresolved markers in plan.md or research.md.

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | 0.75 | Requirements are otherwise single-interpretation (e.g. REQ-DA-002's "exactly as it does at base commit `62fbd6baf`", REQ-DA-013's "as an enrichment of that gate"); the band is set by D1 — REQ-DA-005 (:200) and REQ-DA-012 (:236) carry a population predicate that contradicts REQ-DA-008/009's label semantics (:215-221), so two core requirements do not have a single consistent reading. |
| Completeness | 0.95 | 1.0 band minus one instrument gap | All sections present: HISTORY :21-25, §A context :27-97, §B solution :98-176, §C requirements :178-294, §D constraints :296-321, §E boundaries :323-368, 4× `### Out of Scope — <topic>` H3 with concrete `-` bullets (:376, :389, :402, :415 — verified against `internal/spec/lint.go` `OutOfScopeRule`, which checks the region + bullet items), §G references :428-436. The 0.05 deduction is D4: plan.md's byte-unchanged PRESERVE claim has no closing verification instrument. |
| Testability | 0.80 | 0.75-1.0 boundary | 16 of 17 release-blocking RED cells fully conform to verification-completeness §2.1 (single-invocation command + verbatim stdout + exit code as own field + document-level SHA pin `62fbd6baf` + green path naming the flipping milestone). Verified by re-execution at this HEAD: P1/P3 reproduce (decision_gate count 0 in types.go/defaults.go, exit 1); P19 reproduces (`:217` exactly 1 hit, both local and template). AC-DA-002's empty-sweep cell correctly discloses `[no tests to run]` and pairs with the §D.2 floor ≥4. Deductions: D2 (P11 non-conforming form) and D3 (AC-DA-019 classification). |
| Traceability | 0.95 | 1.0 band minus indirect rows | §D.3 covers all 21 REQs (verified row-by-row); every AC references valid REQ ids. AC-DA-011's REQ cell is "— (preserve)" with the indirect mapping to REQ-DA-013 documented at §D.4:176-177; REQ-DA-021's coverage is indirect (D7). No orphaned ACs, no uncovered REQs. |

Overall: harmonic mean(0.75, 0.95, 0.80, 0.95) = **0.85**.

## Audit-Focus Results (beyond the canonical checklist)

- **Exclusion fidelity vs SPEC-JFM §F (dispatch focus 2): PASS.** All eight dispatch-named
  exclusions are carried or re-justified: Stage 1/2 split (spec.md:378-381, carried + D1
  re-justification), standalone human gate (:382-383), plan-auditor report contract
  (:384-387 + REQ-DA-014), verdict vocabulary (:391-394 — with the correct layer argument
  that decision labels/actions are not audit vocabulary), PRD template (:395-396), legacy
  EARS / verbatim adoption (:397-400), deny/block (:404-406 + CONST-7 :315-317), mode
  auto-selection (:407-410). The zero-flag rule re-enters legitimately as a
  decision-layer property (REQ-DA-019) — distinct surface from the JFM audit-verdict
  exclusion, and the SPEC draws that boundary explicitly (:392-394).
- **Verification completeness (dispatch focus 3): PASS with exceptions (D2, D3, D4).**
  Swept-count floors present (§D.2) with empty-sweep tokens recorded (P17); mutant-probe
  reasoning visible in research §3 (the false-positive-control incident) and R3's
  distinctive-token anchor choice; RED-for-the-right-reason disclosures present
  (AC-DA-002 "swept 0", AC-DA-004 "nothing unconditioned exists").
- **Coordinate integrity (dispatch focus 4): PASS with one locator error (D6).** Content
  anchors + as-of-SHA locating-aid pattern used throughout (acceptance header :12-14,
  plan R1 :45-46, research R1 :107-109, research §5 :78). One stale locator at
  spec.md:365 — see D6.
- **Operator-decision coherence (dispatch focus 5): PASS.** OD-1 (default `off`) matches
  the proposal's own "선택적으로 적용할 수 있는" scoping (verified at
  issue-1683-proposal.md:41-42) and CONST-1. OD-2's routing to manager-docs verified
  against the catalog: `.moai/config/sections/delegation.yaml` assigns the `project`
  phase to `[Explore, manager-docs]` with `moai-workflow-project` — manager-docs owns
  project-doc scaffolding. OD-3 (defer auditor) is consistent with REQ-DA-014/§E.2 and
  the t401 §A.3 measurement it cites. OD-1/2/3 stay inside the D1-D10 lane scope.
- **Cross-artifact consistency (dispatch focus 6): PASS.** REQ ids, milestone names
  M1-M5, surface paths, label vocabulary (DECIDED/POLICY-COVERED/EVIDENCE-NEEDED/FOUNDER)
  and verdict actions (DECIDE/NEED_ANALYSIS/NEED_EVIDENCE/DEFER) are consistent across
  spec/plan/acceptance/design/research. Mirror-state claims re-measured and exact:
  spec-assembly.md and askuser-protocol.md byte-identical; manager-spec.md,
  clarity-interview.md, interview.yaml differ — matching spec.md §B D10 and research §4.
  CONST-2 precedent verified at `internal/config/types.go` (`RecommendationMode string
  yaml:"recommendation_mode"`), `internal/config/defaults.go:1163` (`"push"`), and
  `interview_recommendation_mode_test.go` exists; plan M1's `loader_interview.go` exists.
  Research §1/§6 line anchors spot-checked true (DP1 at clarity-interview.md:188;
  [HARD] clause at spec-assembly.md:217).

## Defects Found

D1. **REQ-DA-005/REQ-DA-012 population predicate contradicts the four-label routing** —
spec.md:200-203 and :236-239 vs :215-221 (REQ-DA-008/009) and the fixed row shape
:126-127 — **Severity: major — Class: blocking.** Both requirements scope the index to
"one row per decision … **whose authority no document carries**". But `DECIDED` /
`POLICY-COVERED` rows are by definition decisions whose authority a document DOES carry
(REQ-DA-009 requires them to cite a committed authority anchor; the proposal itself:
"DECIDED/POLICY-COVERED는 repository 안에 실제 authority와 정확한 source/section이
존재하는 경우에만 허용"). Under a literal reading of REQ-DA-005, the index can never
contain a DECIDED or POLICY-COVERED row — deleting the mechanical-routing half of the
adopted model that spec.md §A.2 (:58-61) names as the proposal's core move. design.md's
own routing algorithm (:69-79) presumes both populations exist, so the intent is
recoverable but the normative text is self-inconsistent. **Required fix:** reword the
population predicate in REQ-DA-005 and REQ-DA-012 to the entry condition the design
implements — e.g. "every decision surfaced during clarification or assembly that the
operator does not settle in the interview" — with routing per REQ-DA-008/009/011.
(clarity-interview M2 step 3, plan.md:120-121, already carries the correct "or" shape.)

D2. **P11 — the sole RED cell for release-blocking AC-DA-016 — violates the adopted
single-invocation form** — acceptance.md:108-110 — **Severity: major — Class:
blocking.** The cell's command is a pipe (`grep -rn "Detect" <…> | grep -c "Explain"`)
and carries an unexpanded placeholder file list (`<manager-spec.md spec-assembly.md
clarity-interview.md>`), so it is neither single-invocation (§2.1(a): pipes are outside
the form) nor runnable verbatim as read (§2.1(b)). By the rule the SPEC itself adopts in
its header (:10-15), a non-conforming citation takes the undecidable disposition — which
strips release-blocking eligibility from AC-DA-016 as classified in §D.1. The probe is
also a weak discriminator ("Detect" is a common token; counting lines containing both
words is not "the row carries Detect → Explain → Ask and no preferred answer").
**Required fix:** replace P11 with one conforming single-invocation probe on a
distinctive token (e.g. `grep -c "Detect → Explain"
.claude/skills/moai/workflows/plan/clarity-interview.md` and companions, or a single
grep across the three spelled-out paths) with its own verbatim stdout + exit code at the
pinned SHA, and re-key AC-DA-016's RED cell to it.

D3. **AC-DA-019 misclassified release-blocking while its own cell concedes no red is
observable** — acceptance.md:39 (RED: "behavioral RED not statically observable"), :43
(Critical list includes 019) vs design.md:120-122 — **Severity: minor — Class:
blocking.** The off-mode behavioral assertion is green-at-arrival by design (at base the
flow cannot create an index); by the SPEC's own stated principle (design §5, citing
verification-completeness §2.1's undecidable disposition) a criterion whose red cannot
be produced is regression-guard, not release-blocking. The internal inconsistency sits
in the closure-gate severity list. **Required fix:** reclassify AC-DA-019 as
regression-guard (the M5 repro remains its verifying act; §D.4 already documents the
static cells as necessary-not-sufficient), or split out a flippable static component
with a conforming red.

D4. **Byte-unchanged PRESERVE claim for plan-auditor.md has no closing verification
instrument** — plan.md:22 vs acceptance.md:32 (AC-DA-012 greps only `decision-index` = 0)
— **Severity: minor — Class: blocking.** plan.md §A declares
`.claude/agents/moai/plan-auditor.md` "byte-unchanged", but no AC or §D.5 closure gate
verifies byte-identity; AC-DA-012 passes even if the file were edited in any other way.
A stated criterion without an instrument is an unfinished check under the SPEC's own
adopted doctrine. **Required fix:** add a preserve ledger cell to §D.5/§E — e.g. `git
diff --stat 62fbd6baf..HEAD -- .claude/agents/moai/plan-auditor.md` → empty output at
close, pinned to both SHAs.

D5. **AC-DA-017's coupling-grep observation asserted in prose without a ledger entry** —
acceptance.md:37 — **Severity: minor — Class: optional.** "coupling grep: 0
co-occurrences … — 0 at base" names an observation but carries no command/stdout/exit id;
§2.1 requires the carrier to be a table cell or a cited ledger entry. **Fix:** add a P20
ledger entry for the coupling grep at the pinned SHA.

D6. **Stale locator in §E.5** — spec.md:365 cites "spec-assembly.md:212" for the
SPEC-JFM 0.2.2-conditioned `(권장)`-label clause; the clause is at :221 (inside the
:217-223 `[HARD]` block; :212 falls in the render-html verb text) — **Severity: minor —
Class: optional.** Cosmetic under the content-anchored discipline the SPEC adopts (the
clause text is quoted alongside), but the printed locator is wrong. **Fix:** correct to
`:217-223 block` or drop the line number.

D7. **REQ-DA-021's AC coverage is indirect** — acceptance.md:170 maps it to
"AC-DA-018 (neutrality pass in M4) + template-neutrality CI guard" — **Severity: minor —
Class: optional.** Acceptable reuse of an existing binary guard
(`.github/workflows/template-neutrality-check.yaml`); recorded so the closure gate reads
the CI guard as REQ-DA-021's verifying act. No change required.

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Recommendation

**FAIL — 4 blocking findings (D1-D4), all cheap to fix; the re-audit is scoped to this
defect delta (retry-loop contract, iteration 2 of 3).**

Numbered fixes for manager-spec:

1. (D1) Reword REQ-DA-005 (spec.md:200-203) and REQ-DA-012 (:236-239): replace "whose
   authority no document carries" with the entry condition "that the operator does not
   settle during the interview" (or equivalent), so the DECIDED/POLICY-COVERED routing of
   REQ-DA-008/009 is reachable. Keep the escalation direction of REQ-DA-011 unchanged.
2. (D2) Replace P11 (acceptance.md:108-110) with a conforming single-invocation probe —
   spelled-out paths, no pipe, distinctive token, verbatim stdout + exit code at
   `62fbd6baf` — and re-key AC-DA-016's RED cell.
3. (D3) Reclassify AC-DA-019 as regression-guard in §D.1/§D.2 (or give its flippable
   static half a conforming red); update §D.5 closure gate 1 accordingly.
4. (D4) Add the plan-auditor.md byte-identity preserve cell to §D.5 closure gates (or a
   regression-guard AC), pinned `62fbd6baf..close`.
5. (Optional, D5/D6) Add the P20 coupling-grep ledger entry; correct the §E.5 locator to
   the :217-223 block.

Rationale for the verdict: the must-pass firewall is fully green (MP-1..MP-7, evidence
above), and the numeric score 0.85 meets the Tier L threshold — but D1 is a
requirements-layer contradiction in the SPEC's central mechanism (the index population
vs the four-label routing), and D2/D3 leave one release-blocking AC resting on a cell
the SPEC's own binding discipline disqualifies. Neither may enter run-phase unfixed.
Everything else audited clean: exclusion fidelity is complete and correctly re-justified,
the D4 authority-register measurement reproduces exactly (gitignore lines, check-ignore
exit 1, untracked state), the config precedent and mirror-divergence claims re-measure
true, OD-1/2/3 are inside the lane scope and catalog-consistent, and 16 of 17
release-blocking RED cells fully conform to the two-cell discipline.
