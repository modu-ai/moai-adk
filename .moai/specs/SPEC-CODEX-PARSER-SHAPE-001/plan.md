# SPEC-CODEX-PARSER-SHAPE-001 — implementation plan

> Plan-phase artifact. Every measured statement traces to
> `.moai/reports/t1053/verdict.md` by section. No figure appears here that is not
> in that file, and nothing here re-runs those measurements.
>
> v0.2.0 (card t1203): the #1718 real case adds a second evidence base,
> `.moai/reports/t1203/verdict.md` (tree `df526c9a9`), cited as `t1203 §N`. The
> same rule binds it.

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

v0.2.0: GitHub #1718 supplied the live different-shape case the t1053
measurements lacked. Two real adversarial bodies from another project
(moai-cowork) carry the verdict behind a persona greeting and the findings as a
table or as bold severity-word bullets; both synthesize as `inconclusive`/0 on
the adversarial path, and across 142 of that project's codex-gate sessions no
run produced a parsed finding (t1203 §1–§3; spec.md §A.6).

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
   on `verdict == fail && len(findings) == 0 && GateUnmet == ""` and so catches
   it regardless of shape. (c) exists because of V8 (verdict.md §A2).
5. **The exact-count tests are the measured defence against partial drift**
   (verdict.md §4, §A3) and look like a tidy-up target. See §G anti-pattern 1.
6. **The #1718 case** (t1203 §1–§4; spec.md §A.6): on the raw bodies no
   candidate has measured coverage; (c) catches only the ctrlB shape (body2 with
   its greeting removed, measured `fail`/0 in both modes). The adversarial
   `inconclusive` that REQ-CPS-010 freezes is what #1718 reports as the defect
   (spec.md §C.1).

## §C Pre-flight — the gate that sits ahead of everything

[HARD] **M0 must close before Implementation Kickoff Approval is asked.**

**M0 is CLOSED (2026-09-21): result SAME-SHAPE.** Record:
`.moai/reports/t1053/live-convention-20260921.md` (tree `a5c3f5dc6`, codex-cli
0.155.1). The convention has not drifted, so the risk is **prospective, not
currently active** — which sets urgency, not treatment (verdict.md §A1), and
leaves the §C candidate space unchanged. **Scope narrowed in v0.2.0:** that
conclusion holds for the 2026-09-21 moai-adk-go measurement only; in the
moai-cowork project the risk is measured as active (t1203 §3, §4; spec.md §A.4). Two tolerated sub-shape differences
(line range with the end discarded; absolute path) are recorded in spec.md §A.4.
The requirement below is retained because it governs any re-measurement.

`AC-CPS-001` requires a live codex call after the usage-limit reset
(2026-09-21 04:21), its body recorded verbatim, and an explicit
same-shape / different-shape verdict against the fixture. Until that record
exists, the SPEC does not enter the run phase (REQ-CPS-001).

What M0 does **not** decide: the prescription. The lead's recorded reasoning
(verdict.md §A1) is that the live comparison sets **urgency**, not **treatment** —
if the convention matches, the future-change exposure remains; if it differs, the
same treatment is simply more urgent. The eighteen offline measurements already
narrowed the treatment space to the three candidates (a)–(c); candidate (d) was
added in v0.2.0 from the #1718 case and is entirely inferred.

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

### M0 — live-convention comparison (pre-run gate; highest change likelihood) — CLOSED 2026-09-21, same-shape

Performed after 2026-09-21 04:21. One live codex review call through the moai MCP
path; the returned body recorded verbatim alongside the invocation, the tree, and
the codex CLI version. Compared against the shape the recognizers accept. Outcome
recorded as same-shape or different-shape, with the observed body as the evidence.

This is the milestone whose result can change what the rest of the SPEC is worth,
which is why it is first and why it gates the Kickoff.

Closes AC-CPS-001, AC-CPS-002.

### M1 — candidate selection and the REQ-CPS-010 decision (operator decisions at the Kickoff gate)

The operator selects one or more of (a), (b), (c), (d). The selection is recorded
in the SPEC's progress record together with the reason given.

[HARD] **The Kickoff now carries a second decision (v0.2.0).** The operator must
also decide whether REQ-CPS-010 is kept as written or revised (spec.md §C.1,
REQ-CPS-013). The two are coupled: keeping REQ-CPS-010 excludes (d) and any part
of (a) that changes adversarial-path behaviour. This plan presents both options
and resolves neither; the choice and its reason are recorded in the progress
record before any adversarial-affecting candidate is implemented.

Candidates ordered below **by measured failure-shape coverage** (verdict.md §E3,
§A2; t1203 §1). **That ordering criterion is coverage, and nothing else** — it is
not a recommendation, not a cost ranking, and not a plan-side preference.

The coverage column is split: **measured** shapes are those a probe actually
exercised, **inferred** shapes are those the mechanism suggests but nothing ran.
An inferred shape carries no weight in the ordering.

| Order | Candidate | Measured coverage | Inferred, NOT measured |
|---|---|---|---|
| 1= | (a) reduce shape dependence | **V2 and V8**, both modes — widening `codexFindingBullet` / `codexFindingLine` to accept numbered-list markers turned native V2 from `pass`/0 into `fail`/2 and V8 from `fail`/0 into `fail`/1, with V1 (`fail`/2) and V9 (`pass`/0) unchanged as regression controls (verdict.md §A5) | V3 bold severity, V4 heading-per-finding, V5 JSON, V6 severity-as-word — **each needs a different recognizer change and none was exercised**; the #1718 raw bodies and ctrlB — **three further distinct changes** (relaxed line-head verdict anchor, localized verdict label, table / bold severity-word findings), none exercised (spec.md §C #1718 table) |
| 1= | (c) verdict/findings contradiction detection | **V8** and **ctrlB**, both modes — each measured as `fail` with `findings=0` (verdict.md §E3; t1203 §1), which is the contradiction itself. On the raw #1718 bodies it catches nothing (measured `inconclusive`/0 — no blocking verdict survives) | — |
| 3= | (b) shape-change detection → inconclusive | native V2–V6 are measured as silent `pass`/0 (verdict.md §E3), so the failure shape is measured; the remedy is not — adversarial is already implemented, and that implemented output is what #1718 reports as the defect (t1203 §4); native is blocked on a prior disambiguation mechanism (verdict.md §E4) | that a native downgrade can be made without turning V9 into `inconclusive` |
| 3= | (d) pin the output format in the adversarial prompt | none — the failure shape is measured (#1718 raw bodies `inconclusive`/0; population, t1203 §1, §3), no remedy was exercised | that codex honours a format instruction over a project persona instruction; that the pinned format holds across projects. Requires the REQ-CPS-010 revision option |

**The ordering changed in v0.2.0, and the change is stated rather than left to be
noticed.** Before the #1718 case, (a)'s measured half (V2 + V8) strictly
contained (c)'s (V8), which put (a) first. Adding ctrlB — a real-body shape,
measured under (c) and not under (a) — ends that containment: (a) and (c) now
cover two measured shapes each, overlapping at V8, and neither contains the
other. On the stated criterion they are **tied** (`1=`); the listing order inside
the tie carries no meaning. (b) and (d) are likewise tied (`3=`): for each, the
failure shape is measured and the remedy is not.

Two things were weighed against this order and neither moves it *on this
criterion*:

- (a)'s inferred half — now including the three #1718 recognizer changes — adds
  nothing, because inferred shapes carry no weight.
- (c)'s shape-independence (§C trade-offs — it catches the contradiction
  regardless of shape, because it keys on
  `verdict == fail && len(findings) == 0 && GateUnmet == ""` rather than on any
  recognizer) is not a coverage count. It is a property of a different kind.

[HARD] **That is a limitation of the criterion, not a verdict on (c).** Coverage
counting cannot express "bounded by an enumerated shape set" versus "not bounded
at all", so the order above under-describes (c) by construction: (a)'s V8
coverage is bounded by the shapes its recognizers were widened for, and a shape
outside that set recreates V8 exactly, while (c)'s is not bounded. An operator
selecting for durability rather than for measured breadth has a sound reason to
take (c) first, and this order is not an argument against that — it is not a
recommendation (see the criterion sentence above).

[HARD] **ctrlB coverage is not #1718 coverage.** (c)'s ctrlB cell is a real body
with its greeting removed; codex did not emit it verbatim. On the raw #1718
bodies, no candidate has measured coverage (spec.md §C #1718 table). An operator
selecting to close #1718 as reported is selecting unexercised work under any
candidate.

[HARD] **A wider (a) is not a measured (a).** Only the numbered-list widening was
run. An operator selecting (a) for V3–V6 or for the #1718 shapes is selecting
unexercised recognizer changes, and the run phase measures each of them rather
than inheriting the V2/V8 result.

Closes AC-CPS-003, AC-CPS-015.

### M2 — implement the selected candidate(s)

Scoped to whichever of (a) / (b) / (c) / (d) the operator selected. The
per-candidate behavioural requirements are REQ-CPS-005 (b), REQ-CPS-006 (c),
REQ-CPS-007 (a), REQ-CPS-012 (d); only the selected ones apply.

Where a selected candidate is claimed to address #1718, M2 first commits the
sanitized reductions of both #1718 shapes and shows each reproduces its raw
body's measured output on the unmodified parser (REQ-CPS-014, AC-CPS-011) —
before any recognizer or prompt change, so the RED observation is taken on a
fixture the tree actually carries.

For (b), the disambiguation mechanism of REQ-CPS-005 is part of M2, not a
prerequisite assumed to exist: without it, extending (b) to native turns a
genuine clean review into `inconclusive` (verdict.md §E4).

Closes the AC matching the selected candidate(s): AC-CPS-004 / AC-CPS-005 /
AC-CPS-006 / AC-CPS-012 / AC-CPS-013 / AC-CPS-014, plus AC-CPS-011 where #1718 is
claimed.

### M3 — preserved-behaviour guards (mechanical; lowest change likelihood)

Regression coverage for the properties no candidate may break: the clean-review
native path stays `pass` (AC-CPS-008), the adversarial path is unchanged
(AC-CPS-009 — applies as written only if the operator keeps REQ-CPS-010), the exact-count assertions remain exact (AC-CPS-007), and
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
6. **Reading ctrlB as #1718.** ctrlB is body2 with its greeting removed; (c)
   catching it says nothing about the raw bodies, on which (c) catches nothing
   (t1203 §1, §4).
7. **Committing the raw #1718 bodies as fixtures.** They carry another project's
   absolute paths and content. Fixtures are sanitized reductions that first prove
   they reproduce the raw bodies' measured output (REQ-CPS-014).

## §H Cross-references

- `.moai/reports/t1053/verdict.md` — the evidence base (§E1–§E4, §4, §5, §6, §7,
  §A1–§A5)
- `.moai/reports/t1203/verdict.md` — the #1718 real case (§1–§5), tree
  `df526c9a9`
- `internal/cli/mcp_codex.go` — `codexStatedVerdict`, `codexScoredVerdict`,
  `codexFindingBullet`, `codexFindingLine`, `codexFindingsOf`,
  `codexVerdictSignalsOf`, `codexUnrecognizedVerdict`
- `internal/cli/codex_findings_parse_test.go` — the exact-count assertions
- SPEC-CODEX-VERDICT-SYNTH-001, SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 — sibling
  SPECs on the same file, consumed unchanged
- card t1052 — the `next_steps` axis, closed
