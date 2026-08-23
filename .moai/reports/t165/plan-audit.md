# SPEC Review Report: SPEC-SVG-QUALITY-ABSORB-001
Iteration: 1/2 (Tier M ceiling)
Verdict: PASS
Overall Score: 0.857 (harmonic mean; Tier M threshold 0.80)

Audited at worktree `.claude/worktrees/t165` @ `5c24fe67f` (branch `WT-svg-quality`),
verified identical at start and end of audit. SPEC files not modified.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency — REQ-1..REQ-8 sequential, no gaps, no duplicates
  (spec.md:91-135; headings `### REQ-1` … `### REQ-8`).
- [PASS] MP-2 EARS/GEARS compliance — **judged against the requirement layer only**
  (all 8 `REQ-XXX` entries in spec.md §3; the Given-When-Then entries in acceptance.md
  are the verification layer and are graded under Testability, not here). All 8 REQs
  match the Ubiquitous form "`<subject>` SHALL [response]" — REQ-1 (spec.md:93
  "`authoring.md` §2 SHALL state…"), REQ-2 (:98), REQ-3 (:103), REQ-4 (:109), REQ-5
  (:114), REQ-6 (:120), REQ-7 (:125), REQ-8 (:131-132). No informal "should/must try"
  requirements; no GWT scenario presented as a REQ.
- [PASS] MP-3 YAML frontmatter validity — `moai spec lint` executed against the
  worktree spec.md: `✓ No findings — all SPEC documents are valid` (exit 0). All 12
  canonical fields present with correct types (spec.md:2-15); no snake_case aliases;
  `tier: M`, `era: V3R6`, `status: draft`, `phase: "v3.2"` all enum/type-valid.
  Note: `priority: medium` is lowercase where the schema doc lists `Medium` — the
  linter (the mechanical SSOT) accepts it.
- [N/A] MP-4 language neutrality — single-skill scoped (`.claude/skills/moai-domain-svg-infographic`,
  markdown + one .mjs linter); not multi-language tooling. Auto-passes.
- [PASS] MP-5 D7 cross-SPEC reconciliation — `grep -Eo 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'`
  over spec/plan/acceptance returns only the SPEC's own ID. No external SPEC
  references, nothing to reconcile. ("B-7/B-8/B-9" in §2 are survey/card classification
  labels, not SPEC-IDs.)
- [PASS] MP-6 D8 cross-platform discipline — `grep -c 'syscall'` = 0 across all three
  artifacts. Auto-pass.
- [PASS] MP-7 clarification gate — `grep -rn '\[NEEDS CLARIFICATION'` over the SPEC dir:
  0 matches. research.md does not exist (Tier M does not require it) → N/A for that
  file; plan.md clean.

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity, resolvable | REQ-7's exception-list host file never named (§2 "Lands in" column covers A-1..A-6 only; spec.md:45-52); AC-7b's Given "the candidate set considered" (acceptance.md:103) is undefined anywhere in the artifacts — card t165 names journey / ER-schema detail / quadrant / timeline, the SPEC dropped them; "ten files" (spec.md:86-87, plan.md:8) not derivable — see D2. All three resolvable from card t165 + the skill's own structure. |
| Completeness | 1.0 | all sections + frontmatter | HISTORY (spec.md:159), WHY (§1), WHAT (§2 with per-item "Lands in" table), REQUIREMENTS §3 (8 entries), ACs in acceptance.md (14 rows, Tier M-correct), four `### Out of Scope — <topic>` H3 headings each with specific bullets (spec.md:58-82). Frontmatter 12/12 + lint clean. Artifact set exactly 3 (spec/plan/acceptance) — Tier M budget holds. |
| Testability | 0.75 | several ACs need minor interpretation | 11 of 14 ACs are crisply binary, and the two trap ACs are genuinely bidirectional (verified — see below). Soft spots: AC-REQ-4 "an entry a reader cannot check against a diagram fails" (acceptance.md:68, judgment phrasing); AC-7b undefined candidate set weakens its Given; plan.md §D hard constraint "no external asset at view time / Google Fonts rejected" (:53-54) has **no AC** — a verbatim lift importing font references would pass all 14 ACs; AC-BUDGET defers the numeric budget to skill-authoring.md (acceptable at SHOULD). |
| Traceability | 1.0 | full bidirectional coverage | 8/8 REQs covered (REQ-1→AC-1; REQ-2→AC-2+BUDGET; REQ-3→3a/3b; REQ-4→AC-4; REQ-5→AC-5; REQ-6→6a/6b; REQ-7→7a/7b/7c; REQ-8→8a/8b). 14/14 ACs reference valid REQs. No orphans either direction. |

**Harmonic mean** = 4 / (1/0.75 + 1/1.0 + 1/0.75 + 1/1.0) = **0.857** ≥ 0.80.

## Claimed-structure verification (caller-requested re-measurements)

1. **REQ 8 / AC 14** — confirmed. acceptance.md §D matrix has 14 rows: 13 MUST-PASS
   + 1 SHOULD (AC-BUDGET). The "thirteen criteria" phrasing (spec.md:88,
   acceptance.md:29) counts MUST-PASSs only — internally consistent, under the
   Tier M ceiling of 16.
2. **A-3 absence measurement re-run (primary checkout)** — the claim as literally
   stated does NOT reproduce: `grep -rniE 'aria|role=|<title>|<desc>'` over
   `SKILL.md references/ scripts/` = **1 hit** (scripts/render.mjs:128,
   "environment **varia**ble" — bare `aria` is a substring of "variable").
   The *semantic* claim is TRUE: anchored rerun `aria-|\brole=|<title>|<desc>` = 0
   hits, and `\baria\b` = 0. See D1.
3. **Template-mirror count** — the mirror tree
   `internal/template/templates/.claude/skills/moai-domain-svg-infographic/` holds
   **6 files** (SKILL.md, 3 references, 2 scripts), not 10. The SPEC's own changed-file
   set ("Lands in" + plan §A) is 4 files (SKILL.md, authoring.md, archetypes.md,
   check-svg.mjs) → 4 + 4 mirrors = 8. "Ten" is reachable only as 4 changed + the
   whole 6-file mirror tree, i.e. counting 2 mirrors (sketch.md, render.mjs) that no
   A-item touches. See D2. Tier M remains justified regardless (13 MUST-PASS > the
   Tier S AC ceiling of 8 — the prescribed tier-up signal).
4. **Trap ACs verified against current file state** —
   - **AC-6b is a genuine deletion detector**: the rule it guards exists today —
     `references/authoring.md` §3.2 "Focal discipline — single accent, signalled by
     colour" (lines 221-237: "Reserve the `accent` token for that focal node…
     Scattering `accent` across several 'important' nodes erases the signal"). A
     restructure that drops §3.2 fails AC-6b's search; AC-6a would not notice. Sound.
   - **AC-3b asserts the failure direction**: "no accessible name → exits non-zero and
     names the missing part; full contract → exits zero" (acceptance.md:56-59). The
     extension target runs today (`node scripts/check-svg.mjs --help` → usage, exit 0;
     exit-code contract "1 = errors, 2 = usage" already fits). An always-pass or
     always-fail checker fails one of the two Givens. Sound.
5. **Bypass-list falsifiability** — a type FAILS to qualify iff no committed sample
   pair exists (REQ-7 :125-127, AC-7a :97-99, AC-7b :103-106, constraint §5 :148-149,
   decision D3). Empty list is an explicitly legitimate outcome (AC-7b :105-106).
   Binary and stated. `.moai/reports/` is NOT gitignored in this repo
   (`git check-ignore` exit 1), so "committed renders" under
   `.moai/reports/t165/samples/` is satisfiable as written.
6. **Card t165 ↔ SPEC scope** — 6 A-items + evidence-gated bypass list match; MIT
   attribution and Template-First carried (REQ-8). Three content-level drifts found
   (D5). Card says "Tier S"; the SPEC documents the tier-up with rationale — not a
   silent drift.
7. **Rejections (39-type catalogue, Google Fonts, HTML/motion/terminal)** — documented
   as prose constraints (§2 Out of Scope, §5, plan §D) but enforced by no AC (D6).

## Defects Found (0 blocking; all non-must-pass)

D1. SVG-A11Y-PREMISE — spec.md:143-144 (§4 Evidence) — the stated measurement
    `aria|role=|<title>|<desc>` → "0 matches" does not reproduce verbatim: the literal
    pattern yields 1 hit (`scripts/render.mjs:128`, "variable" ⊃ "aria"). The anchored
    form yields 0, so the requirement's premise (no accessibility markup) is correct —
    only the cited evidence is wrong. — Severity: minor — Class: optional — Required
    fix: anchor the pattern in §4 (e.g. `aria-|\brole=|<title>|<desc>` → 0 matches).
D2. FILE-COUNT — spec.md:86-87 / plan.md:8-9 — "the scope reaches ten files once
    template mirrors are counted" is not derivable: changed files per the SPEC's own
    tables = 4 (+4 mirrors = 8); the mirror tree = 6 (4+6=10 only by counting
    sketch.md and render.mjs mirrors that no A-item touches). Tier M is independently
    justified (13 MUST-PASS > Tier S AC ceiling 8), so the verdict is unaffected. —
    Severity: minor — Class: optional — Required fix: show the arithmetic or drop the
    count and rest the escalation on the criteria-count prong alone.
D3. LIST-HOST-UNSTATED — spec.md §2 / plan.md M5 — REQ-7's exception list has no
    "Lands in" target: no artifact names the file the list lives in, yet AC-7a requires
    "the entry references them". A verifier must guess (the natural home is SKILL.md's
    mermaid-vs-SVG selection section). — Severity: minor — Class: optional — Required
    fix: name the host file for the exception list in the §2 In-Scope block or REQ-7.
D4. CANDIDATE-SET-UNDEFINED — acceptance.md:103 (AC-7b Given) — "the candidate set
    considered" is never enumerated; card t165 seeds it (journey · ER/db-schema
    detail · quadrant · timeline) and plan M5 should carry those seeds forward instead
    of leaving the candidate derivation to run-phase invention. — Severity: minor —
    Class: optional — Required fix: add the card's candidate list to plan M5 as the
    starting set.
D5. CARD-CONTENT-DRIFT — spec.md:47 / acceptance.md:36-38 — three rule details present
    in card t165 were dropped in the SPEC: (a) A-1's exception clause "no routing
    behind a non-endpoint box **except dash + visible label segment**" (the SPEC and
    AC-REQ-1 state the rule absolutely); (b) A-5's coupling "size preset determines
    the type ramp"; (c) A-3's `role=img` generalized to bare `role`. (a) matters most:
    sibling card t166 builds the executable checker from these rules and an absolute
    rule will false-flag legitimate dashed crossings. — Severity: minor — Class:
    optional — Required fix: carry the A-1 exception (and optionally b/c) into the §2
    table and AC-REQ-1.
D6. UNENFORCED-REJECTION — plan.md:53-54 (§D Constraints, "Hard") — "no external
    asset at view time / Google Fonts rejected" has no acceptance criterion; absorbed
    text that verbatim-imports a font CDN reference passes all 14 ACs (AC-BUDGET and
    the neutrality guard check other things). — Severity: minor — Class: optional —
    Required fix: add one absence AC over the changed files (bounded grep for
    `fonts.googleapis|@import url|<link` in SKILL.md/references → 0).
D7. EVIDENCE-PATH-LOCALITY — plan.md §C.3 / spec.md §4 — the survey
    (`.moai/reports/diagram-design-absorption/`: survey.md 27,310 B +
    diagram-design-absorption-20260822.md 3,086 B + .html) exists only in the primary
    checkout and is NOT git-tracked (`git ls-files` = 0); the execution worktree t165
    has no copy (verified absent). Plan pre-flight 3 ("survey.md readable — the rule
    text is derived from it") therefore fails in the tree the run will execute in,
    unless the lane happens to read the primary. — Severity: major — Class: optional —
    Required fix: pin the absolute path in plan §C.3, or commit/copy the survey into
    the worktree, before run-phase entry.

## Regression Check

Iteration 1 — no prior-iteration defects to check.

## Recommendation

PASS at 0.857. All seven must-pass criteria hold with cited evidence; the defect list
contains no blocking findings — D1-D6 are wording/seed fixes absorbable in a single
small revision or at run-phase M5, and D7 (the only major) is a one-line path fix that
should land before `/moai run` so the M5 sample work and the §C pre-flight resolve in
the execution tree. This PASS does not bypass Implementation Kickoff Approval.
