# SPEC Review Report: SPEC-DECISION-AUTHORITY-001

Iteration: 2/3 (Tier L ceiling = `harness.yaml` `plan_audit_tier_ceilings.L: 3`)
Auditor: plan-auditor (independent) — tree `.claude/worktrees/t692`, HEAD `4c2ec29f1`,
branch `WT-judgment-authority`. Scope per the retry-loop contract: the enumerated defect
delta from `plan-audit-iter1.md` (D1-D7) plus a regression check over that delta — not a
from-scratch re-audit. Fix commit under review: `4c2ec29f1` (one commit on top of
`6732d1461`, the tree measured at iteration 1; diff touches exactly 4 files —
spec.md, acceptance.md, design.md, and the committed iter1 report — and no others).

**Verdict: PASS**
**Overall Score: 0.95** (harmonic mean; Tier L threshold 0.85 met with margin)

---

## Regression Check (Iteration 2)

Defects from iteration 1:

- **D1 (blocking, major) — RESOLVED**: REQ-DA-005 and REQ-DA-012 population predicates
  rewritten (spec.md:203-208, :241-246) — the index population is now keyed on
  settlement status ("the operator does not settle in the interview") and explicitly
  spans all four label classes, with the routing tied to REQ-DA-008 through REQ-DA-011.
  §B D5 prose aligned (:144-147). The contradiction that made DECIDED/POLICY-COVERED
  rows unreachable is closed: an implementer reading §C alone now produces all four
  classes. Verified against REQ-DA-009/010/011 — no new conflict with the closed
  authority register or the escalation rule (see Advisory A1 for the residual seam).
- **D2 (blocking, major) — RESOLVED**: P11 rebuilt (acceptance.md:109-117) as a
  single-invocation `grep -c "Detect → Explain"` over three expanded paths — no pipe,
  no placeholder, distinctive token, verbatim stdout + exit code, pinned `6732d1461`.
  Re-executed at HEAD `4c2ec29f1`: reproduces exactly (0/0/0, exit 1). AC-DA-016's
  release-blocking status restored (§D.1 Critical list now 001..010, 013..018).
- **D3 (blocking, minor) — RESOLVED**: AC-DA-019 reclassified regression-guard in all
  four surfaces — §D row (acceptance.md:42), §D.1 (Critical excludes 019; Guard list
  includes it with the undecidable-disposition rationale), §D.5 closure gate 2 (M5
  repro explicitly named as the completing act), design.md §5 (:116-125, consistent
  restatement). No surface still calls it release-blocking.
- **D4 (blocking, minor) — RESOLVED**: P21 ledger entry added
  (acceptance.md:158-166): `git diff --stat 62fbd6baf..HEAD --
  .claude/agents/moai/plan-auditor.md` → empty output, exit 0. Re-executed at HEAD in
  both the HEAD form and the fully pinned form (`62fbd6baf..4c2ec29f1`): both empty,
  exit 0 — byte-identity holds through the fix commit itself. Wired into the AC-DA-012
  row (acceptance.md:32) and §D.5 closure gate 2. The stated PRESERVE criterion now has
  a closing instrument.
- **D5 (optional) — RESOLVED**: P20 ledger entry added (acceptance.md:151-156) —
  `grep -cE "decision_gate.*recommendation_mode|recommendation_mode.*decision_gate"
  internal/config/types.go` → 0, exit 1; re-executed at HEAD, reproduces. AC-DA-017's
  RED cell now cites "P3, P20" — the coupling observation has a ledger carrier.
- **D6 (optional) — RESOLVED**: §E.5 locator corrected (spec.md:372-375) to the
  `:217-223` `[HARD]` block, label clause at `:221` — matches the measurement from
  iteration 1 exactly, and carries the content-anchor discipline note.
- **D7 (optional) — UNCHANGED (accepted)**: REQ-DA-021's AC coverage remains indirect
  (AC-DA-018 + the template-neutrality CI guard). Accepted reuse; advisory only.

Delta regression sweep: the fix commit's diff contains only the reviewed hunks; no
REQ/AC ids added or removed (21/19 re-counted by grep); frontmatter stays valid
(`version: "0.1.1"` quoted, HISTORY 0.1.1 row records the fix pass); sibling artifacts
remain stateless (their `status:` grep hits are body-prose mentions of the statelessness
rule, frontmatter blocks unchanged); progress.md §E.1 unchanged (honest pre-audit
value). All iter1 probe subjects (internal/config, `.claude/` doctrine files) are
byte-untouched by the fix commit, so the remaining ledger entries (P1-P10, P12-P19)
remain valid at the new HEAD without re-measurement.

## Must-Pass Results (re-confirmed on the delta)

- MP-1 PASS — 21 REQ / 19 AC ids, sequential, unchanged (grep re-count).
- MP-2 PASS — the rewritten REQ-DA-005/012 retain GEARS shape (While-state-driven with
  the PASS-equivalent [While][When] compound; no informal normative text introduced).
- MP-3 PASS — 12 canonical fields intact at the new version; no aliases.
- MP-4 N/A — unchanged (single-language-scoped).
- MP-5 PASS — SPEC-ID reference set unchanged (sibling completed).
- MP-6 PASS — `syscall` still absent (fix commit touched no relevant text).
- MP-7 PASS — no [NEEDS CLARIFICATION] markers introduced.

## Category Scores

| Dimension | Score | Change | Evidence |
|-----------|-------|--------|----------|
| Clarity | 0.90 | 0.75 → 0.90 | The population predicate is now single-interpretation and the four-class reachability is explicit (spec.md:203-208, :241-246). Residual: one resolvable seam — REQ-DA-012's inline-settled sub-clause presumes "its authority anchor" exists; the anchorless case is forced consistent by REQ-DA-011 + the §A.4 meta-rule (escalate, never downgrade), but the document leaves that composition implicit (Advisory A1). |
| Completeness | 1.00 | 0.95 → 1.00 | D4's missing preserve instrument landed (P21 + closure gate 2). All sections, 4 Out-of-Scope H3s, and all frontmatter fields verified. |
| Testability | 0.95 | 0.80 → 0.95 | All 16 release-blocking RED cells now conform to §2.1's four-element form; the three new/rebuilt probes (P11, P20, P21) re-executed and reproduce at HEAD. Reserve: P20 is a single-file carrier of the coupling claim (completed by M1's independence test per its own green-path note), and §D.5 gate 2's stale "Both guard ACs" wording (Advisory A2). |
| Traceability | 0.95 | unchanged | §D.3 mapping intact; AC-DA-011 documented-indirect (§D.4); REQ-DA-021 indirect (Advisory A3). |

Overall: harmonic mean(0.90, 1.00, 0.95, 0.95) = **0.95**.

## Defects Found

No blocking defects remain. Advisory items (recorded, no fix required for PASS):

A1. REQ-DA-012's inline-settled sub-clause (spec.md:245-246) says a settled decision "is
recorded as a `DECIDED` or `POLICY-COVERED` row citing its authority anchor" — for an
inline settlement with no closed-register anchor, REQ-DA-009 disqualifies those labels
and REQ-DA-011 forces `FOUNDER`; the composite is consistent but implicit, and §B D5's
prose covers only the unsettled case. One clarifying sentence would remove the seam.
Severity: minor — Class: optional.

A2. §D.5 closure gate 2 (acceptance.md:203-205) still opens "Both guard ACs" while now
enumerating three (P19, P7/P21, AC-DA-019's repro). Stale count word. Severity: trivial
— Class: optional.

A3. P21's command uses the moving ref `HEAD`; acceptable because the cell mandates
recording both resolved SHAs at each measurement (the moving ref is the claim's subject
— identity through close — per the R2 freeze pattern), but the closing measurement MUST
record the resolved close SHA alongside `62fbd6baf`, as §D.1's guard language already
requires. Severity: minor — Class: optional.

## Recommendation

**PASS — score 0.95, threshold 0.85 met.** All four blocking findings from iteration 1
are closed with substance (verified by diff reading AND by re-executing every new or
rebuilt probe at HEAD `4c2ec29f1`); both optional findings D5/D6 were also taken. The
three advisory items (A1-A3) are left to the orchestrator's discretion — none blocks
run-phase entry. The skip-eligibility inputs for the run-gate now hold on this verdict:
verdict PASS, score ≥ 0.85, and the artifact hash is current as of `4c2ec29f1` (any
further artifact edit re-opens Phase 1 per the hash-unchanged condition).
