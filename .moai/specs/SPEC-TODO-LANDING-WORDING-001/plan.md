# Plan — SPEC-TODO-LANDING-WORDING-001

Plan-phase implementation plan for the five wording amendments to
`SPEC-TODO-LANDING-ATTRIBUTION-001` (card t486). Tier S: wording-only, single milestone pair, no
code.

---

## §A Context

The completed SPEC `SPEC-TODO-LANDING-ATTRIBUTION-001` (v0.4.0, card t472) understates its
measured residual and leaves four further t482 findings unrecorded. This card applies five wording
amendments to its `spec.md` and `plan.md`. All figures are carried from t482's committed evidence
(`.moai/reports/t482/`) — the run phase edits prose, it never measures.

The amendment targets, located (line numbers measured at plan time in this tree, branch
`WT-t472-spec-wording`; prose anchors named so a drifted line number cannot orphan an amendment):

| # | REQ | Target anchor |
|---|-----|---------------|
| 1 | REQ-TLW-001 | spec.md §A.4.2 `[HARD] The residual is a FLOOR, not a total` block (incl. the history line "1 → 19 → 7 → ≥10"); plan.md §D residual risk row ("at least 10" ×2) |
| 2 | REQ-TLW-002 | spec.md §A.4.2 residual discussion (new S1 paragraph) |
| 3 | REQ-TLW-003 | spec.md §A.4.2 (property paragraph appended to the floor block) |
| 4 | REQ-TLW-004 | spec.md §D `### Out of Scope — axis C, landing evidence storage` |
| 5 | REQ-TLW-005 | spec.md §A.4 form table row 3b (`Observed` 76) + the section text carrying the count |

## §B Known Issues

- The target SPEC's prose is dense and citation-heavy; insertions must match its register and
  provenance discipline (every figure travels with its evidence path).
- The target is `completed`; a careless edit that touches frontmatter would constitute a lifecycle
  transition this card is forbidden to perform (REQ-TLW-006, C-4).
- Line numbers in the target drift with each edit; anchor on section headings and quoted phrases,
  not line numbers.

## §C Pre-flight

1. Read the five anchor regions of the target `spec.md` / `plan.md` and confirm each quoted anchor
   phrase still resolves (grep, exit 0).
2. Confirm the target frontmatter still reads `status: completed`.
3. Confirm t482 evidence files exist at the cited paths (`verdict.md`, `residual-evidence.txt`,
   `s1-reconcile.txt`, `form3b-delta.txt`).

## §D Constraints (from spec.md §D)

- Wording only; no forms added; no Go code or templates.
- Figures cited, never re-measured.
- Only the target's `spec.md` and `plan.md` change.
- No residual count as a reduction target.

## §E Self-Verification (run phase)

- `moai spec lint .moai/specs/SPEC-TODO-LANDING-WORDING-001/spec.md` → 0 errors.
- `grep` each amendment's load-bearing token in the amended files (28, S1, 18/14, casing latitude,
  t359 termination, 77) → present; `grep "at least 10"` in target spec.md/plan.md → absent (or
  present only inside quoted history, per REQ-TLW-001's history-line form).
- `git diff --name-only` over the amendment commits → target changes confined to `spec.md` +
  `plan.md` of the target (AC-TLW-006).
- Target frontmatter after amendment: `status: completed` unchanged.

## §F Milestones

Ordered by decision-reversibility (the interpretive wording first, mechanical substitutions last):

- **M1 — Property and floor (REQ-TLW-001 + REQ-TLW-003).** Rewrite the §A.4.2 floor block: 28
  measured, still a floor, classification and evidence path, history line extended; append the
  non-closure property paragraph (three grounds + under-estimation sentence). These two share a
  block and one editing pass. Priority: High.
- **M2 — S1 naming with rejection (REQ-TLW-002).** Insert the S1 paragraph: definition, example,
  both figures (18 / 14) as distinct measures, three-round anonymity, operator rejection with its
  ground. Priority: High.
- **M3 — Termination path (REQ-TLW-004).** Amend the axis C out-of-scope section: transitional
  instrument, t359 the sole termination path, further rounds not the plan. Priority: Medium.
- **M4 — Form 3b count + casing (REQ-TLW-005).** Mechanical: 76 → 77 in the form table row, casing
  latitude sentence with the delta subject and evidence path; verify 347/309/38 untouched.
  Priority: Medium.
- **M5 — plan.md §D row + immutability check (REQ-TLW-001 tail + REQ-TLW-006).** Update the risk
  row's two "at least 10" occurrences; final `git diff --name-only` + frontmatter check.
  Priority: Medium.

## §G Anti-Patterns

- **Re-measuring "just to check".** Any re-count of the residual in this card's run phase violates
  C-2 and repeats the defect the card exists to retire.
- **Writing 28 as a target.** "Reduce the residual to 28" or any shrinking framing is forbidden
  (REQ-TLW-003 [HARD]).
- **Adopting S1 in passing.** The S1 paragraph must not read as a form definition; it names and
  records a rejection.
- **Touching frontmatter.** No status change, no `amendment_of:`, no target version bump framing
  as a reopening.
- **Correcting 76 → 77 and "helpfully" adjusting 309/38.** The residual is invariant under
  REQ-TLW-005.

## §H Cross-References

- `spec.md` §A.2 — the evidence table every figure is carried from.
- Target: `.moai/specs/SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md` (§A.4, §A.4.2, §D) and `plan.md`
  (§D).
- Evidence: `.moai/reports/t482/verdict.md` (§2.4, §2.6, §2.7, §4, §4.1, §5.1, §5.2, §6),
  `residual-evidence.txt`, `s1-reconcile.txt`, `form3b-delta.txt`.
