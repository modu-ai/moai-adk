# SPEC-CODEX-PARSER-SHAPE-001 — implementation plan

> Plan-phase artifact. Every measured statement traces to
> `.moai/reports/t1053/verdict.md` by section. No figure appears here that is not
> in that file, and nothing here re-runs those measurements.
>
> v0.2.0 (card t1203): the #1718 real case adds a second evidence base,
> `.moai/reports/t1203/verdict.md` (tree `df526c9a9`), cited as `t1203 §N`. The
> same rule binds it.
>
> v0.2.1 (card t1203): repair of plan-audit iter-1 defects D1–D14
> (`.moai/reports/t1203/plan-audit.md`). Wording and structure only.

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
   `inconclusive`, so candidate (b) has no mechanism work there. Whether that
   `inconclusive` is an acceptable outcome is a separate question — #1718 reports
   it as the defect — and it is the operator's (spec.md §C.1).
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
   candidate has an executed remedy; (c) is deduced — not executed — to catch
   only ctrlB, a derived body (body2 with one greeting token removed, measured
   `fail`/0 in both modes). The adversarial `inconclusive` that REQ-CPS-010's
   rationale clause relies on is what #1718 reports as the defect (spec.md §C.1).

## §C Pre-flight — the gate that sits ahead of everything

[HARD] **M0 must close before Implementation Kickoff Approval is asked.**

**M0 is CLOSED (2026-09-21): result SAME-SHAPE.** Record:
`.moai/reports/t1053/live-convention-20260921.md` (tree `a5c3f5dc6`, codex-cli
0.155.1). The convention had not drifted, so the risk was **prospective, not
currently active** — which sets urgency, not treatment (verdict.md §A1); that
observation did not change the §C candidate space (candidate (d) was added
later, in v0.2.0, from the #1718 case). **Scope, re-drawn in v0.2.1:** the
conclusion holds for **moai-adk-go · native path (`mode = "native"`) · 2026-09-21
· codex-cli 0.155.1** only (`live-convention-20260921.md` §1, §3). The #1718
observation differs on project, path, date, and CLI version at once, so no single
axis is measured as the separating one; "differs by project" is an inference.
For the moai-cowork bodies the risk is measured as active in the sense that no
parser run produced a finding (t1203 §3), measured on session-log final bodies
whose identity to the parser's `reviewText` is a Gap (spec.md §A.6). The
moai-adk-go adversarial live convention has never been measured (spec.md §A.4).
Two tolerated sub-shape differences (line range with the end discarded; absolute
path) are recorded in spec.md §A.4.
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
      section of one of the evidence files — `.moai/reports/t1053/verdict.md`,
      `.moai/reports/t1053/live-convention-20260921.md`, or
      `.moai/reports/t1203/verdict.md` (the single auditor re-probe field is
      cited to `.moai/reports/t1203/plan-audit.md`, spec.md §A.1).
- [ ] No figure appears that is absent from those files.
- [ ] Every inference is labelled as one, including "the project sets the shape".
- [ ] The candidate ordering names its unit (failure shape) and applies one test
      to all four candidates.
- [ ] AC-CPS-011 carries a sanitization check command and a forbidden-token list.
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

### M1 — candidate selection and the REQ-CPS-010 question (operator decisions at the Kickoff gate)

The operator selects one or more of (a), (b), (c), (d). The selection is recorded
in the SPEC's progress record together with the reason given.

[HARD] **The Kickoff carries a second, separate question (v0.2.0; re-posed
v0.2.1).** The operator also answers whether the #1718 adversarial outcome —
`inconclusive` with an empty findings list for a body whose prose states FAIL —
is acceptable, which decides whether REQ-CPS-010 is kept as written or revised
(spec.md §C.1, REQ-CPS-013). The two questions are **not coupled**: REQ-CPS-010's
normative clause binds only the handling of a body that stays unrecognized, no
candidate changes that handling, and so neither answer excludes a candidate.
spec.md §C.1 states what each candidate changes on the adversarial path. This
plan presents the question and answers it neither way; the answer and its reason
are recorded in the progress record before the first run-phase commit of any
selected candidate.

**The ordering criterion (REQ-CPS-004), stated once and applied to all four.**

- **Unit: the failure shape**, not the body instance. The §E3 V8 body and ctrlB
  are two instances of one shape, V8, and count once.
- **Test, identical for every candidate:** a shape counts toward a candidate
  only when (1) the shape's failure was measured on the unmodified parser and
  (2) a probe **executed that candidate's remedy** on the shape and observed the
  outcome.
- Shapes whose outcome under the candidate is only **deduced** (fixed on paper
  by measured fields, remedy never run) or only **inferred** are shown for the
  operator's information and **carry no weight** in the order.
- The order is the count, highest first; equal counts are tied, and the listing
  order inside a tie carries no meaning.

**That ordering criterion is coverage, and nothing else** — it is not a
recommendation, not a cost ranking, and not a plan-side preference
(verdict.md §E3, §A2, §A5; t1203 §1).

| Order | Candidate | Shapes with the remedy executed (count) | Deduced, not executed (no weight) | Inferred, NOT measured (no weight) |
|---|---|---|---|---|
| 1 | (a) reduce shape dependence | **2 — V2 and V8**, both modes: widening `codexFindingBullet` / `codexFindingLine` to accept numbered-list markers turned native V2 from `pass`/0 into `fail`/2 and V8 from `fail`/0 into `fail`/1, with V1 (`fail`/2) and V9 (`pass`/0) unchanged as regression controls (verdict.md §A5) | none — no other shape's outcome is fixed by a measured field | V3 bold severity, V4 heading-per-finding, V5 JSON, V6 severity-as-word — **each needs a different recognizer change and none was exercised**; the #1718 raw bodies and ctrlB — **three further distinct changes** (relaxed line-head verdict anchor, localized verdict label, table / bold severity-word findings), none exercised (spec.md §C #1718 table) |
| 2= | (b) shape-change detection → inconclusive | **0** — native V2–V6 are measured as silent `pass`/0 (verdict.md §E3), but no native remedy has run; on adversarial the mechanism already exists, and its output is what #1718 reports as the defect (t1203 §4) | none — the native remedy's output depends on a disambiguation mechanism that does not exist (verdict.md §E4), so no measured field fixes it | that a native downgrade can be made without turning V9 into `inconclusive` |
| 2= | (c) verdict/findings contradiction detection | **0** — (c) has never been implemented or run | **V8**, both modes: the predicate `verdict == fail && len(findings) == 0 && GateUnmet == ""` evaluated over measured outputs — the §E3 V8 body (`fail`/0, verdict.md §E3) and ctrlB (`fail`/0, t1203 §1; `gate=""`, spec.md §A.1). On the raw #1718 bodies the measured output is `inconclusive`/0, so the predicate is false there | V8 instances in shapes nobody has enumerated |
| 2= | (d) pin the output format in the adversarial prompt | **0** — the failure shape is measured (#1718 raw bodies `inconclusive`/0; population, t1203 §1, §3), no remedy was exercised | none — (d) acts on what codex emits, which no measured field fixes | that codex honours a format instruction over a project persona instruction; that the pinned format holds across projects |

**The ordering changed twice, and both changes are stated rather than left to be
noticed.** v0.1.0 placed (a) first on the containment of (c)'s measured V8 in
(a)'s V2 + V8. v0.2.0 tied (a) and (c) at `1=` by counting ctrlB as a second
shape for (c); that counted a body instance, not a shape, and counted (c)'s
deduced outcome as if (c) had run — two departures from the stated criterion
that the plan-audit recorded (`.moai/reports/t1203/plan-audit.md` D2). Under the
criterion above, applied to all four, (a) is first on 2 shapes and (b), (c), (d)
are tied at 0.

What the order does not show, stated without weight in either direction:

- (a)'s inferred shapes — including the three #1718 recognizer changes — are
  not counted, because inferred shapes carry no weight.
- (c)'s deduced V8 outcome, and the property that its predicate reads the
  parser's output rather than any recognizer (spec.md §C trade-offs), are not
  coverage counts. Counting cannot express "bounded by an enumerated shape set"
  versus "not bounded". An operator who weighs durability rather than executed
  breadth may reach a different conclusion than this order; the order does not
  address durability, in either direction.

[HARD] **ctrlB coverage is not #1718 coverage.** ctrlB is a derived body — body2
with one greeting token removed by the measurer; codex did not emit it. On the
raw #1718 bodies no candidate has an executed remedy or a deduced catch
(spec.md §C #1718 table). An operator selecting to close #1718 as reported is
selecting unexercised work under any candidate.

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
sanitized reductions of both #1718 shapes, plus the S2′ control and the two
negative fixtures, and shows each reduction reproduces its raw body's measured
output on the unmodified parser (REQ-CPS-014, AC-CPS-011) — before any
recognizer or prompt change, so the RED observation is taken on a fixture the
tree actually carries. That observation becomes the RED-now of AC-CPS-012 and
AC-CPS-013 and returns them to release-blocking (acceptance.md §C.1).

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
   For (b), adversarial has no mechanism to add — it already exists; whether its
   `inconclusive` output is acceptable is not established (#1718 reports it as
   the defect; spec.md §C.1). Native needs a disambiguation mechanism first.
4. **Checking the verdict only.** V8 passes any check that reads the verdict
   alone: the verdict is `fail` and correct, while `findings` is empty and wrong
   (verdict.md §E3, §7).
5. **Citing the installed binary.** It predates this tree (`f67d2193f`). Compile
   from the tree.
6. **Reading ctrlB as #1718.** ctrlB is a derived body — body2 with one greeting
   token removed; (c) being deduced to catch it says nothing about the raw
   bodies, on which (c) is deduced to catch nothing (t1203 §1, §4).
7. **Committing the raw #1718 bodies as fixtures.** They carry another project's
   absolute paths and content. Fixtures are sanitized reductions that first prove
   they reproduce the raw bodies' measured output and pass the sanitization check
   (REQ-CPS-014, AC-CPS-011).
8. **Reading "differs by project" as measured.** The 2026-09-21 and #1718
   observations differ on project, path, date, and CLI version at once; which
   axis separates them is an inference (spec.md §A.4, §A.6).

## §H Cross-references

- `.moai/reports/t1053/verdict.md` — the evidence base (§E1–§E4, §4, §5, §6, §7,
  §A1–§A5)
- `.moai/reports/t1203/verdict.md` — the #1718 real case (§1–§5), tree
  `df526c9a9`
- `.moai/reports/t1053/live-convention-20260921.md` — the 2026-09-21 live
  observation (`mode = "native"`, codex-cli 0.155.1)
- `.moai/reports/t1203/plan-audit.md` — plan-audit iter-1 (defects D1–D14
  repaired in v0.2.1; source of the single `gate=""` re-probe field)
- `internal/cli/mcp_codex.go` — `codexStatedVerdict`, `codexScoredVerdict`,
  `codexFindingBullet`, `codexFindingLine`, `codexFindingsOf`,
  `codexVerdictSignalsOf`, `codexUnrecognizedVerdict`
- `internal/cli/codex_findings_parse_test.go` — the exact-count assertions
- SPEC-CODEX-VERDICT-SYNTH-001, SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 — sibling
  SPECs on the same file, consumed unchanged
- card t1052 — the `next_steps` axis, closed
