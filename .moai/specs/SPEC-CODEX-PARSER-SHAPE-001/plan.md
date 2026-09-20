# SPEC-CODEX-PARSER-SHAPE-001 — implementation plan

> Plan-phase artifact. Every measured statement traces to
> `.moai/reports/t1053/verdict.md` by section. No figure appears here that is not
> in that file, and nothing here re-runs those measurements.

## §A Context

The subject is `internal/cli/mcp_codex.go` — the codex review parser. Three
recognizers (`codexStatedVerdict`, `codexScoredVerdict`,
`codexFindingBullet` / `codexFindingLine`) carry the whole coupling to the shape
of codex's prose, and `codexUnrecognizedVerdict(method)` decides what happens
when none matches (verdict.md §E2).

The card is measurement-first: the offline measurement is complete (verdict.md
§E3, eighteen measurements), and the one measurement that would decide urgency —
whether the live convention still matches the fixture — was blocked all day by a
codex account usage limit (verdict.md §E1, §6).

## §B Known issues carried into the plan

1. **Native's silent pass** (verdict.md §E3): V2–V6 all return `pass` with
   `findings=0` on the native path. The gate passes and nothing is recorded.
2. **Adversarial already loud** (verdict.md §E3): the same bodies return
   `inconclusive`, so candidate (b) has no work there.
3. **The disambiguation problem** (verdict.md §E4): on native, a clean review
   (V9) and a shape-changed review (V2–V6) are indistinguishable to the parser,
   so (b) cannot be extended to native as-is.
4. **V8** (verdict.md §E3): verdict survives, findings empty, both modes. **As
   the parser currently stands, neither (a) nor (b) detects it** — the measured
   §E3 state, and not a claim about the direction (a): a widening under (a) does
   resolve the measured V8 instance (`fail`/0 → `fail`/1, verdict.md §A5). The
   durable difference is that (a) resolves V8 only for the shapes its widening
   covers — a later shape outside that set recreates V8 exactly — while (c) keys
   on `verdict == fail && len(findings) == 0` and so catches it regardless of
   shape. (c) exists because of V8 (verdict.md §A2).
5. **The exact-count tests are the measured defence against partial drift**
   (verdict.md §4, §A3) and look like a tidy-up target. See §G anti-pattern 1.

## §C Pre-flight — the gate that sits ahead of everything

[HARD] **M0 must close before Implementation Kickoff Approval is asked.**

`AC-CPS-001` requires a live codex call after the usage-limit reset
(2026-09-21 04:21), its body recorded verbatim, and an explicit
same-shape / different-shape verdict against the fixture. Until that record
exists, the SPEC does not enter the run phase (REQ-CPS-001).

What M0 does **not** decide: the prescription. The lead's recorded reasoning
(verdict.md §A1) is that the live comparison sets **urgency**, not **treatment** —
if the convention matches, the future-change exposure remains; if it differs, the
same treatment is simply more urgent. The eighteen offline measurements already
narrowed the treatment space to the three candidates.

[HARD] M0 also does not close verdict.md §6's first Gap. That Gap records that
the convention was unmeasured on the measurement day, and that fact is permanent.
M0 records a NEW measurement at a NEW point in time.

## §D Constraints

- The run-phase candidate choice is the operator's, made at the Implementation
  Kickoff Approval gate. This plan does not pre-select one (REQ-CPS-003).
- `next_steps` is closed by card t1052 and is not touched (REQ-CPS-011).
- Windows path behaviour and live `audit_multi` calls are unmeasured; they may be
  recorded as explicitly-unmeasured notes and never as claims.
- The installed moai binary (`f67d2193f`) is older than this tree and is not a
  measurement source. Run-phase verification compiles from the tree.

## §E Self-verification

Before the plan is presented as complete:

- [ ] Every measured statement in spec.md, plan.md, and acceptance.md cites a
      verdict.md section.
- [ ] No figure appears that is absent from verdict.md.
- [ ] AC-CPS-001 cannot be read as satisfied by the offline measurements.
- [ ] The exact-count preservation constraint is stated in spec.md §A.5 and
      fixed by an acceptance criterion, not only in prose.
- [ ] No candidate is adopted as the implementation plan.

## §F Milestones

Ordered by decision-reversibility: the decisions most likely to change come
first, the mechanical work last.

### M0 — live-convention comparison (pre-run gate; highest change likelihood)

Performed after 2026-09-21 04:21. One live codex review call through the moai MCP
path; the returned body recorded verbatim alongside the invocation, the tree, and
the codex CLI version. Compared against the shape the recognizers accept. Outcome
recorded as same-shape or different-shape, with the observed body as the evidence.

This is the milestone whose result can change what the rest of the SPEC is worth,
which is why it is first and why it gates the Kickoff.

Closes AC-CPS-001, AC-CPS-002.

### M1 — candidate selection (operator decision at the Kickoff gate)

The operator selects one or more of (a), (b), (c). The selection is recorded in
the SPEC's progress record together with the reason given.

Candidates ordered below **by measured failure-shape coverage** (verdict.md §E3,
§A2). **That ordering criterion is coverage, and nothing else** — it is not a
recommendation, not a cost ranking, and not a plan-side preference.

The coverage column is split: **measured** shapes are those a probe actually
exercised, **inferred** shapes are those the mechanism suggests but nothing ran.
An inferred shape carries no weight in the ordering.

| Order | Candidate | Measured coverage | Inferred, NOT measured |
|---|---|---|---|
| 1 | (a) reduce shape dependence | **V2 and V8**, both modes — widening `codexFindingBullet` / `codexFindingLine` to accept numbered-list markers turned native V2 from `pass`/0 into `fail`/2 and V8 from `fail`/0 into `fail`/1, with V1 (`fail`/2) and V9 (`pass`/0) unchanged as regression controls (verdict.md §A5) | V3 bold severity, V4 heading-per-finding, V5 JSON, V6 severity-as-word — **each needs a different recognizer change and none was exercised** |
| 2 | (c) verdict/findings contradiction detection | **V8**, both modes — the measured V8 output is `fail` with `findings=0` (verdict.md §E3), which is the contradiction itself; the only candidate that catches it | — |
| 3 | (b) shape-change detection → inconclusive | native V2–V6 are measured as silent `pass`/0 (verdict.md §E3), so the failure shape is measured; the remedy is not — adversarial is already implemented and native is blocked on a prior disambiguation mechanism (verdict.md §E4) | that a native downgrade can be made without turning V9 into `inconclusive` |

**The ordering is unchanged, and this is a stated conclusion rather than a
default.** Two things were weighed against it and neither moves it *on this
criterion*:

- Removing (a)'s inferred half does not move (a): its measured half (V2 + V8)
  strictly contains (c)'s measured coverage (V8).
- (c)'s shape-independence (§C trade-offs — it catches the contradiction
  regardless of shape, because it keys on `verdict == fail && len(findings) == 0`
  rather than on any recognizer) does not move (c) up either, because the stated
  criterion is **measured failure-shape coverage** and shape-independence is not
  a coverage count. It is a property of a different kind.

[HARD] **That is a limitation of the criterion, not a verdict on (c).** Coverage
counting cannot express "bounded by an enumerated shape set" versus "not bounded
at all", so the order above under-describes (c) by construction: (a)'s V8
coverage is bounded by the shapes its recognizers were widened for, and a shape
outside that set recreates V8 exactly, while (c)'s is not bounded. An operator
selecting for durability rather than for measured breadth has a sound reason to
take (c) first, and this order is not an argument against that — it is not a
recommendation (see the criterion sentence above).

(b) stays last because its remedy — unlike (a)'s and (c)'s — is the one nothing
has yet exercised on the native path.

[HARD] **A wider (a) is not a measured (a).** Only the numbered-list widening was
run. An operator selecting (a) for V3–V6 is selecting four unexercised recognizer
changes, and the run phase measures each of them rather than inheriting the V2/V8
result.

Closes AC-CPS-003.

### M2 — implement the selected candidate(s)

Scoped to whichever of (a) / (b) / (c) the operator selected. The per-candidate
behavioural requirements are REQ-CPS-005 (b), REQ-CPS-006 (c), REQ-CPS-007 (a);
only the selected ones apply.

For (b), the disambiguation mechanism of REQ-CPS-005 is part of M2, not a
prerequisite assumed to exist: without it, extending (b) to native turns a
genuine clean review into `inconclusive` (verdict.md §E4).

Closes the AC matching the selected candidate(s): AC-CPS-004 / AC-CPS-005 /
AC-CPS-006.

### M3 — preserved-behaviour guards (mechanical; lowest change likelihood)

Regression coverage for the properties no candidate may break: the clean-review
native path stays `pass` (AC-CPS-008), the adversarial path is unchanged
(AC-CPS-009), the exact-count assertions remain exact (AC-CPS-007), and
`next_steps` is untouched (AC-CPS-010).

## §G Anti-patterns

1. **Loosening an exact-count assertion.** `want 4 parsed findings` in
   `internal/cli/codex_findings_parse_test.go` reads as brittle hardcoding and is
   a natural cleanup target. Measurement says the opposite: under partial drift
   (one bullet of four changed), the exact count is what fires, while a
   "non-empty" style control passes because three findings remain (verdict.md §4).
   Relaxing it to `>= 1` erases the defence and the erasure is silent.
2. **Reading the offline measurements as the live comparison.** The eighteen
   measurements describe parser behaviour on **bodies the measurer wrote**, not
   what codex emits (verdict.md §6). AC-CPS-001 exists precisely to separate the
   two, and is written so it cannot be closed by them.
3. **Treating (b) as the cheap option.** The dispatch estimate that (b) is
   cheaper did not separate the two modes; it was corrected (verdict.md §A4).
   Adversarial has nothing to do; native needs a disambiguation mechanism first.
4. **Checking the verdict only.** V8 passes any check that reads the verdict
   alone: the verdict is `fail` and correct, while `findings` is empty and wrong
   (verdict.md §E3, §7).
5. **Citing the installed binary.** It predates this tree (`f67d2193f`). Compile
   from the tree.

## §H Cross-references

- `.moai/reports/t1053/verdict.md` — the evidence base (§E1–§E4, §4, §5, §6, §7,
  §A1–§A5)
- `internal/cli/mcp_codex.go` — `codexStatedVerdict`, `codexScoredVerdict`,
  `codexFindingBullet`, `codexFindingLine`, `codexFindingsOf`,
  `codexVerdictSignalsOf`, `codexUnrecognizedVerdict`
- `internal/cli/codex_findings_parse_test.go` — the exact-count assertions
- SPEC-CODEX-VERDICT-SYNTH-001, SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 — sibling
  SPECs on the same file, consumed unchanged
- card t1052 — the `next_steps` axis, closed
