---
id: SPEC-CODEX-PARSER-SHAPE-001
title: "codex review parser — output-shape coupling and its silent failure"
version: "0.2.0"
status: draft
created: 2026-09-20
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "codex, review-parser, shape-coupling, silent-pass, verdict-findings-contradiction"
---

# SPEC-CODEX-PARSER-SHAPE-001 — the codex review parser is coupled to the shape of codex's prose

## HISTORY

- 2026-09-20 · v0.1.0 · manager-spec · Initial authoring from card t1053. Every
  measured statement in this SPEC is attributed to `.moai/reports/t1053/verdict.md`
  by section. No measurement is re-derived here and no figure appears here that is
  not in that file.
- 2026-09-26 · v0.2.0 · manager-spec · Amendment from card t1203 (GitHub #1718).
  Adds the #1718 real case — two live adversarial codex bodies whose verdict and
  findings both fall outside the recognizers — as a second evidence base,
  `.moai/reports/t1203/verdict.md` (tree `df526c9a9`). Adds §A.6; narrows the
  scope of the §A.4 "prospective, not currently active" conclusion to the
  measurement it came from; adds the #1718 case and a new candidate (d) to §C;
  records the conflict between REQ-CPS-010 and #1718 as an operator decision
  (§C.1, REQ-CPS-013) without changing REQ-CPS-010's text; widens REQ-CPS-003 /
  REQ-CPS-004 from three candidates to four; adds REQ-CPS-012..014. No figure is
  introduced that is absent from the two verdict files.

---

## §0 Governing principle [HARD]

> **A parser that cannot recognize the shape must not report the absence of
> findings as the absence of problems.**

The subject of this SPEC is not "codex changed its output". It is that the
adapter's recognizers decide, by regular expression alone, whether a review had
content — and that on one of the two modes a failed recognition is currently
indistinguishable, in the emitted output, from a clean review.

---

## §A Background

### A.1 Evidence base

The single evidence base for this SPEC is `.moai/reports/t1053/verdict.md`,
measured against tree `8b55fc8f0` (verdict.md §3). Section references in this
document (`verdict.md §E3`, `§A2`, and so on) point into that file. Nothing in
this SPEC re-runs those measurements, and no number appears here that is not
recorded there.

**Second evidence base (v0.2.0).** The #1718 real case (§A.6) is attributed to
`.moai/reports/t1203/verdict.md`, measured against tree `df526c9a9` on
2026-09-26. References of the form `t1203 §N` point into that file; unprefixed
`verdict.md §…` references continue to mean the t1053 file. The same rule holds:
nothing here re-runs those measurements, and no number appears that is not
recorded in one of the two files.

### A.2 The coupling surface (verdict.md §E2)

Three recognizers in `internal/cli/mcp_codex.go` carry the whole coupling:

| Recognizer | What it matches |
|---|---|
| `codexStatedVerdict` | a `Verdict: fail` line |
| `codexScoredVerdict` | a `FAIL 0.79` score line |
| `codexFindingBullet` / `codexFindingLine` | a `- [P1] …` bullet |

When none matches, `codexUnrecognizedVerdict(method)` decides the value: the
native path (review/start) returns `pass`, the adversarial path (turn/start)
returns `inconclusive`.

### A.3 What the offline measurements established (verdict.md §E3)

Eighteen offline measurements — nine body variants (V1–V9) across the two modes —
establish four facts that this SPEC rests on.

1. **The native silent pass is real.** Under every shape change measured
   (V2 numbered list, V3 bold severity without brackets, V4 heading-per-finding,
   V5 JSON, V6 severity-as-word), native returns `pass` with `findings=0`. The
   gate passes and no signal is left behind.
2. **Adversarial is already loud.** The same bodies return `inconclusive` on the
   adversarial path. Candidate (b) below — "detect the shape changed and drop to
   inconclusive" — is therefore **already implemented for adversarial**. Only
   native is open.
3. **Native's `pass` is a declared decision, not an oversight** (verdict.md §E4).
   In native mode, a bullet-less body means codex found nothing to block on. A
   genuinely clean review (V9) and a shape-changed review (V2–V6) are
   indistinguishable to the parser. Candidate (b) therefore cannot be extended to
   native without a prior disambiguation mechanism.
4. **V8 is the sharpest cell.** A body that changed bullet shape but still
   carries a `Verdict: fail` line yields verdict `fail` with `findings=0` in
   **both** modes. The gate blocks correctly while the review content is silently
   lost, and any consumer reading `findings` sees a clean review. **As the parser
   currently stands, neither (a) nor (b) detects it** — that is the measured §E3
   state. It is NOT a statement about the direction (a): a widening under (a)
   does resolve the measured V8 instance (`fail`/0 → `fail`/1, verdict.md §A5).
   The durable difference is stated in §C: (a) resolves V8 only for the shapes
   its widening covers, so a later shape outside that set recreates V8 exactly,
   whereas (c) keys on the contradiction rather than on any recognizer and is
   therefore shape-independent.

### A.4 What is unmeasured

- **The live output convention WAS unmeasured on the measurement day**
  (verdict.md §6) — every codex call was blocked by an account usage limit,
  verified twice, once through the moai MCP path and once by a direct
  `codex exec` call that bypasses moai entirely.

  **It is now measured: SAME-SHAPE.** After the limit reset the live call ran
  (record: `.moai/reports/t1053/live-convention-20260921.md`; tree `a5c3f5dc6`,
  codex-cli 0.155.1 — the same version as the blocked day, so the CLI version is
  not a variable across the two observations). The live body carries
  `- [P1] <message> — <path>:<line>` bullets and the parser structured three of
  them with severities P1/P2/P2, read off that same call's `findings` array
  rather than inferred.

  **What this changes and what it does not.** The convention has NOT drifted, so
  the risk this SPEC addresses is **prospective, not currently active** — that is
  a statement about urgency, not about treatment (verdict.md §A1). The candidate
  space in §C is unchanged, because it was narrowed by the offline measurements
  and not by this one. It also does not retroactively close verdict.md §6's first
  Gap: that Gap records that the convention was unmeasured on the measurement
  day, which remains permanently true.

  **Scope of that conclusion, narrowed 2026-09-26 (t1203).** The observation above
  stands as recorded: one live call, in the moai-adk-go tree, on 2026-09-21,
  returned the fixture shape. What does not stand is reading its conclusion as a
  statement about codex output in general. The #1718 case (§A.6) is a measured
  counter-observation from a different project (moai-cowork): there, live
  adversarial bodies already fall outside the recognizers, and across 142 of that
  project's codex-gate sessions not one run produced a parsed finding (t1203 §3).
  The "prospective, not currently active" conclusion is therefore scoped to the
  2026-09-21 moai-adk-go measurement; for the moai-cowork project the risk is
  **measured as active** (t1203 §4). That the shape-determining variable is the
  target project's instructions rather than the codex CLI version is an
  **inference**, not a measurement (t1203 §4, §5).

- **Two sub-shape differences from the fixture, both tolerated** — observed in
  the same live call, recorded because they are real and small, not because they
  block anything:
  - The live line reference is a **range** (`acceptance.md:150-153`) where the
    fixture carries a single line. `codexPathLineRef`
    (`internal/cli/mcp_codex.go`) captures only the start line, so the parsed
    finding's `line` is `150` and `-153` is discarded. It parses; the range
    information is lost.
  - The live path is **absolute** where the fixture is repo-relative. It matches,
    and the absolute path lands in `File` as-is.
- **The nine variants were chosen by the measurer**, not derived from what codex
  actually emits (verdict.md §6). A shape outside that enumeration may behave
  differently.
- **Windows path behaviour and live `audit_multi` calls are unmeasured**
  (verdict.md §6). They appear nowhere in this SPEC as claims.
- The installed moai binary is older than this tree (build `f67d2193f`,
  verdict.md §6). The offline measurements used tests compiled from the tree and
  are unaffected; the installed binary is not a measurement source for anything
  in this SPEC.

### A.5 [HARD] The exact-count tests are the measured first line of defence

`internal/cli/codex_findings_parse_test.go` carries assertions of the form
`want 4 parsed findings`. Measurement corrected an earlier claim about which test
fires on shape drift (verdict.md §4, §A3):

- Under **partial** drift (one bullet of four changes shape), the exact-count
  assertions are what fire — three of them. A "`Findings` is non-empty" style
  positive control passes, because three findings remain.
- Under **total** drift, the exact-count assertions fire alongside the non-empty
  control, as four simultaneous failures. No test holds a privileged earlier
  position.

[HARD] These assertions look like brittle hardcoding and are therefore a natural
tidy-up target ("loosen to `>= 1`"). **Loosening them erases the only measured
defence against partial shape drift.** Any change to this SPEC's subject that
relaxes an exact-count assertion into an inequality is a silent regression, and
REQ-CPS-008 forbids it.

### A.6 The #1718 real case — live bodies outside the recognizers (t1203)

Evidence: `.moai/reports/t1203/verdict.md`, tree `df526c9a9`, 2026-09-26. The
status column follows the §C convention: **measured** means a probe actually
exercised it; **inferred** means nothing ran and the statement carries no weight.

GitHub #1718 reports an adversarial `codex_audit` whose codex body said FAIL
being synthesized as `inconclusive` with an empty findings list. The two real
bodies of that report were recovered from codex session logs and matched to the
defect locations the issue cites (t1203 §1). They are kept out of this SPEC —
they carry another project's absolute paths and content — and only the short
shape fragments already quoted in t1203 §2 appear below.

**The two shapes (t1203 §2).**

| # | Shape property | body1 | body2 | Status |
|---|---|---|---|---|
| 1 | Verdict label does not open the line | a persona greeting opens the first line; the label is Korean — `판정은 **FAIL**` | a persona greeting opens the first line — `**verdict: fail**` follows it | measured (text) |
| 2 | Findings are not `- [P1]` bullets | a markdown table — `\| **High** \| [SKILL.md:28](…) \|` | bold severity-word bullets — `- **Medium · [test-cases.yaml:38](<…>) · …` | measured (text) |

`codexStatedVerdict` and `codexScoredVerdict` are line-head anchored and
`codexFindingBullet` / `codexFindingLine` accept only the bracketed-severity
bullet, so neither property is recognized (§A.2).

**Both-mode parser output (t1203 §1; `synthesizeReviewOutput` on tree
`df526c9a9`).**

| Input | turn/start (adversarial) | review/start (native) | Status |
|---|---|---|---|
| body1 (raw) | `inconclusive`, findings 0 | `pass`, findings 0 | measured |
| body2 (raw) | `inconclusive`, findings 0 | `pass`, findings 0 | measured |
| ctrlA — positive control, `Verdict: fail` + one `- [P1] … a/b.go:3` bullet | `fail`, findings 1 | `fail`, findings 1 | measured |
| ctrlB — body2 with only the persona greeting prefix removed | `fail`, findings 0 | `fail`, findings 0 | measured |

ctrlA firing shows the probe is not idle. ctrlB separates the two properties:
removing the greeting restores the verdict, while the findings stay empty — the
**V8 shape (§A.3 fact 4) reproduced from a real codex body**, where before it was
known only from a body the measurer wrote.

**Population (t1203 §3).** The final assistant message of each of 142
`moai-codex-gate` sessions whose working directory was the moai-cowork project
(2026-09-14 .. 2026-09-25), run through the same parser in both modes:

| Path | Outcome distribution | Status |
|---|---|---|
| turn/start | `inconclusive`/0 = 138, `pass`/0 = 3, `fail`/0 = 1 | measured |
| review/start | `pass`/0 = 141, `fail`/0 = 1 | measured |
| findings > 0 | 0 of 284 runs | measured |

Text counts over the same 142 bodies (grep): first line opening with the persona
greeting 101, line-head verdict label 4, `[P<digit>]` bullets 0, a `fail` string
present 124 (t1203 §3) — measured.

**What this establishes, and what it does not.**

- #1718 reproduces on the adversarial path; it is not an isolated incident in
  that project (measured, t1203 §1, §3).
- The live convention differs by project: the 2026-09-21 moai-adk-go call was
  same-shape (§A.4), the moai-cowork bodies are not (measured, t1203 §3).
- The shape is set by the target project's instructions, not by the codex CLI
  version (**inferred, not measured**, t1203 §4, §5).
- The adversarial prompt specifying no output format (the
  `codexUnrecognizedVerdict` comment) is what leaves room for project
  instructions to shape the body (**inferred, not measured**, t1203 §4).

**Gaps carried from t1203 §5 — never to be restated as claims.**

- Which path each population session actually took is unknown: the session logs
  carry no review-mode marker, so the population was run through both paths.
- That the last assistant message is byte-identical to the `reviewText` the parser
  received was not established; body1 and body2 were matched to #1718 by the
  issue's defect locations and FAIL wording only.
- The issue's MCP build `60017eb83` was not measured; the parser of tree
  `df526c9a9` was. Whether the parser changed between the two is unchecked.
- A codex memory-citation block trails both bodies and is carried into `summary`
  verbatim (t1203 §2, side observation); it is recorded, not acted on.

---

## §B Requirements (GEARS)

### Pre-run gating requirement

**REQ-CPS-001** — Where the live codex output convention remains unmeasured, the
SPEC shall not enter the run phase; the comparison described in AC-CPS-001 shall
be performed and its result recorded first.

**REQ-CPS-002** — When the live comparison of AC-CPS-001 is performed, the
recorder shall write the observed body verbatim, together with the invocation
that produced it and the tree it was measured against, into this SPEC's evidence
record, so that a later reader can tell what codex actually emitted rather than
what was assumed.

### Candidate selection

**REQ-CPS-003** — The plan shall present candidates (a), (b), (c), and (d) of §C
as non-exclusive candidates with their measured failure-shape coverage, and shall
not adopt one as the implementation plan; the selection is the operator's at the
Implementation Kickoff Approval gate.

**REQ-CPS-004** — Where the plan orders the candidates, the ordering shall
be by measured failure-shape coverage (verdict.md §E3, §A2), and the plan shall
state that criterion explicitly rather than leaving the order to be read as a
recommendation.

**REQ-CPS-013** — When the Implementation Kickoff Approval gate is reached, the
plan shall present the conflict between REQ-CPS-010 and the #1718 case (§C.1) as
an operator decision with two options — keep REQ-CPS-010 as written, or revise
it — and shall not resolve it itself; the operator's choice and the reason given
shall be recorded in the progress record before any candidate that changes
adversarial-path behaviour is implemented.

### Behavioural requirements, conditional on the selected candidate

**REQ-CPS-005** — Where candidate (b) is selected, the native review path shall
distinguish a body carrying no recognized signal because codex found nothing to
block on from a body carrying no recognized signal because its shape was not
recognized, before any body is downgraded from `pass`.

**REQ-CPS-006** — Where candidate (c) is selected, when a synthesized review
output carries a blocking verdict together with an empty findings list **and no
`GateUnmet` annotation**, the adapter shall report that state as
self-contradictory rather than emitting it as a clean review.

**REQ-CPS-006a** — Where candidate (c) is selected, when a review output carries
a blocking verdict together with an empty findings list **and a non-empty
`GateUnmet` annotation**, the adapter shall not report that state as
self-contradictory; an unmet required gate is a legitimate producer of that
shape, not a parser defect.

**REQ-CPS-007** — Where candidate (a) is selected, when a review body carries a
finding in a shape the widened recognizers accept, the adapter shall parse that
finding, and the findings parsed from the shapes recognized before the widening
shall be unchanged.

**REQ-CPS-012** — Where candidate (d) is selected, the adversarial review request
shall specify an output format for the verdict and the findings that the
recognizers of §A.2 accept, and the claim that live codex output follows that
format shall rest on a live observation recorded verbatim, not on a fixture or on
reading the request text.

**REQ-CPS-014** — When a candidate is claimed to address the #1718 case, the claim
shall be established against sanitized reductions of both #1718 shapes (§A.6)
that carry no absolute user path and no content of the originating project, and
each reduction shall first be shown to reproduce, on the unmodified parser, the
output measured for its raw body (t1203 §1).

### Unwanted behaviour (preserved properties)

**REQ-CPS-008** — The test suite shall not replace an exact-count findings
assertion in `internal/cli/codex_findings_parse_test.go` with an inequality; the
exact count is the measured defence against partial shape drift (§A.5).

**REQ-CPS-009** — A review body carrying no findings because codex found nothing
to block on shall not be reported as `inconclusive` on the native path.

**REQ-CPS-010** — The adversarial path's existing behaviour for an unrecognized
body shall not change; it already returns `inconclusive` (verdict.md §E3), and no
candidate in §C adds work there.

> **Conflict note (v0.2.0, not a revision).** REQ-CPS-010's text above is
> unchanged. Its premise is contradicted by the #1718 case: the `inconclusive`
> it preserves is exactly the output #1718 reports as the defect (t1203 §4), and
> candidate (d) — and (a) when widened for the #1718 shapes — does add work on
> the adversarial path. Whether REQ-CPS-010 is kept or revised is an operator
> decision at the Kickoff gate (§C.1, REQ-CPS-013); this SPEC does not make it.

**REQ-CPS-011** — This SPEC shall not alter the codex synthesis `next_steps`
field or any behaviour that card t1052 closed.

---

## §C The candidates (presented, not chosen)

The candidates are **not mutually exclusive**; each targets a different measured
failure shape (verdict.md §A2). (a)–(c) come from the t1053 measurements;
(d) was added in v0.2.0 from the #1718 case (t1203 §4) and is entirely inferred.

The coverage column is split. **Measured** names a shape a probe actually
exercised; **inferred** names a shape the mechanism suggests where nothing ran.
An inferred shape is not evidence and carries no weight.

| Candidate | Failure shape it targets | Measured coverage | Inferred, NOT measured |
|---|---|---|---|
| (a) reduce shape dependence — widen the range of recognized shapes | shapes outside the current recognizers | **V2 and V8**, both modes — a widening of `codexFindingBullet` / `codexFindingLine` to numbered-list markers turned V2 `pass`/0 → `fail`/2 and V8 `fail`/0 → `fail`/1, with V1 and V9 unchanged as regression controls (verdict.md §A5) | V3, V4, V5, V6 — each needs a **different** recognizer change and none was exercised; the #1718 shapes — see the #1718 table below |
| (b) detect a shape change and drop to inconclusive | native's silent `pass` | the failure shape is measured (native V2–V6 are `pass`/0, verdict.md §E3); already implemented on adversarial | the remedy — that a native downgrade is possible without turning V9 into `inconclusive`; blocked on the §A.3 fact 3 disambiguation |
| (c) detect a verdict/findings contradiction | V8 — verdict survives, content empties | **V8**, both modes — the measured V8 output is `fail` with `findings=0` (verdict.md §E3), which is the contradiction itself; **ctrlB**, both modes — a real codex body with only its greeting removed, measured `fail`/0 (t1203 §1) | — |
| (d) pin the output format in the adversarial prompt | adversarial bodies shaped by the target project's instructions (§A.6) | — (nothing measured; the failure shape is measured, t1203 §1, §3, but no remedy ran) | that codex honours a format instruction over a project persona instruction; that the pinned format survives across projects |

Trade-offs:

- **(a)** is effective on both modes, and one widening of it is now measured
  rather than inferred (verdict.md §A5). Two costs. First, regression risk per
  widened recognizer — which the §A5 probe addressed for that one widening only,
  by holding V1 and V9 unchanged as controls. Second, and durable: **(a) resolves
  V8 only for the shapes its widening covers.** A shape outside that set
  recreates V8 exactly, because the verdict line still parses while the finding
  markers no longer do. Extending (a) to V3–V6 means four further recognizer
  changes, each unexercised, each carrying its own regression risk.
- **(b)** is **already implemented for adversarial** — there is nothing to do
  there — and on native it is not cheap, because V9 and V2–V6 are
  byte-indistinguishable to the parser, so a disambiguation mechanism must come
  first (verdict.md §E4, §A4). The dispatch estimate that (b) is cheaper did not
  separate the two modes; that estimate was corrected (verdict.md §A4).
  v0.2.0: "nothing to do there" describes the mechanism, not its adequacy — the
  adversarial `inconclusive` that (b) already produces is the output #1718
  reports as the defect (t1203 §4; §C.1).
- **(c)** keys on `verdict == fail && len(findings) == 0 && GateUnmet == ""` — a
  state that is self-contradictory whatever the body looks like.

  [HARD] The `GateUnmet == ""` conjunct is load-bearing and was added after the
  bare two-term predicate was found to match a legitimate state. `applyGateUnmet`
  (`internal/cli/mcp_codex.go`) converts an `inconclusive` into
  `Verdict = "fail"` with a non-empty `GateUnmet` when
  `workflow.audit.gates.codex` is `required`, and the inconclusive paths carry an
  empty findings list — so `fail` + `findings=0` is exactly what a correctly
  functioning unmet required gate produces. Without the conjunct, (c) would flag
  that gate as a parser contradiction. AC-CPS-005 fixes this as an explicit
  control case so an implementation cannot regress into it.

  Scoped that way, (c) catches V8
  **regardless of shape**, including shapes nobody has enumerated, because it
  reads the parser's output rather than any recognizer. That is a difference in
  kind from (a), not in coverage count: (a)'s V8 coverage is bounded by the
  shapes its recognizers were widened for, (c)'s is not bounded at all. What (c)
  does not do is say anything about V2–V6, where no verdict signal survives, so
  there is no contradiction to detect.
- **(d)** changes what codex is asked to emit rather than what the parser
  accepts. Entirely inferred: no prompt change was exercised. Three costs. First,
  it **touches the adversarial path, which REQ-CPS-010 currently freezes** — it
  cannot be selected without the §C.1 decision. Second, its effect depends on
  codex obeying a format instruction against the target project's own
  instructions, which is unmeasured (t1203 §4, §5). Third, it does nothing for
  the parser: native's silent `pass` and any body that ignores the instruction
  are exactly as exposed as before. It composes with (c) rather than replacing
  it — a body that follows the pinned verdict line but not the pinned finding
  format is V8 again.

### Coverage against the #1718 real case (t1203 §1, §4)

Rows are candidates; columns are the #1718 inputs of §A.6. **Measured** cells
cite a probe output; everything else is inferred.

| Candidate | body1 raw (adversarial) | body2 raw (adversarial) | ctrlB — body2 minus greeting |
|---|---|---|---|
| (a) | not measured — needs **three distinct recognizer changes**, each unexercised: a relaxed line-head anchor for the verdict, a localized verdict label (`판정`), and a findings recognizer for tables and bold severity-word bullets. The line-head anchor is a documented narrowness contract in `codexStatedVerdict`'s comment, so relaxing it is a regression risk of its own (inferred) | not measured — same three changes, less the localized label | not measured — needs the findings recognizer change only (verdict already recognized: `fail`/0 measured) |
| (b) | nothing to add: the current adversarial output is already `inconclusive`/0 (measured), and #1718 reports **that** output as the defect | same (measured `inconclusive`/0) | not applicable — a verdict is recognized (`fail`/0, measured), so no shape-change fall-through occurs |
| (c) | **catches nothing**: output is `inconclusive`/0 (measured) — no blocking verdict survives, so there is no contradiction | **catches nothing**: `inconclusive`/0 (measured) | **catches it**: `fail`/0 in both modes (measured) is the contradiction itself |
| (d) | inferred only — would act before the body exists, by changing it | inferred only | inferred only |

On the native path the raw bodies yield `pass`/0 (measured, t1203 §1) — the
§A.3 fact 1 silent pass — but which path the #1718 calls took is a Gap
(§A.6); #1718 itself reports the adversarial path.

**Net, stated plainly.** On the raw #1718 bodies no candidate has measured
coverage. The only measured cell in which a candidate catches anything is (c) on
the ctrlB shape — a body codex did not emit verbatim.

### C.1 The REQ-CPS-010 conflict — an operator decision, not resolved here

REQ-CPS-010 freezes the adversarial path's unrecognized-body behaviour on the
premise that it "already returns `inconclusive`" and that "no candidate in §C
adds work there". #1718 reports that `inconclusive` as the defect — its reporter
wrote that the structured output cannot be used as an audit gate (t1203 §4) —
and the premise that no candidate adds work there is no longer true once (d), or
an (a) widened for the #1718 shapes, is on the table.

Per REQ-CPS-003's "presented, not chosen" rule, this SPEC presents the two
options and chooses neither (REQ-CPS-013):

| Option | Consequence |
|---|---|
| Keep REQ-CPS-010 as written | the adversarial path stays frozen; (d) and any adversarial-affecting part of (a) are excluded from selection; for #1718 only (c) remains, which catches the ctrlB shape and not the raw bodies (table above); AC-CPS-009 applies unchanged |
| Revise REQ-CPS-010 | the adversarial path becomes eligible for change; the revised text, and a matching revision of AC-CPS-009, are authored at that decision — not pre-written here |

---

## §D Acceptance criteria

Full criteria live in `acceptance.md`. One is load-bearing enough to restate
here.

**Status: AC-CPS-001 and AC-CPS-002 are SATISFIED** by the 2026-09-21 live call
recorded at `.moai/reports/t1053/live-convention-20260921.md` (result:
same-shape — see §A.4). The clause below is the criterion that record had to
meet; it is retained verbatim because it governs any re-measurement.

[HARD] **AC-CPS-001 (live-convention comparison) sits ahead of the Implementation
Kickoff Approval gate.** It is satisfied only by an observation of live codex
output taken after the usage-limit reset (2026-09-21 04:21) and recorded verbatim.
It **cannot** be satisfied by the eighteen offline measurements of verdict.md §E3,
by reading `internal/cli/mcp_codex.go`, by the fixture
`issue1632ReviewBody`, or by any test compiled from this tree. Those are the
substitute measurements the comparison exists to check (verdict.md §1, §6).

---

## §E Exclusions

### Out of Scope — closed by sibling cards

- **The `next_steps` axis is CLOSED by card t1052.** Do not reopen it and do not
  propose changes to it.
- The verdict-synthesis conservative-adoption rule, owned by
  SPEC-CODEX-VERDICT-SYNTH-001. This SPEC consumes it unchanged.
- The blank-review fail-closed path, owned by
  SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001. This SPEC changes nothing about a body
  that is absent; its subject is a body that is PRESENT and unrecognized.

### Out of Scope — unmeasured, therefore not claimed and not acted on

- **Windows path behaviour.** Unmeasured (verdict.md §6). It may appear in a
  report as an explicitly-unmeasured note; it must never appear as a claim.
- **Live `audit_multi` calls.** Unmeasured (verdict.md §6). Same treatment.
- **What shape codex will actually move to.** The nine variants were enumerated
  by the measurer, not derived from codex's behaviour (verdict.md §6). This SPEC
  does not predict a shape.
- **The cause of the #1718 shape.** That the target project's instructions set
  it is inferred (t1203 §4, §5); it is not claimed and not acted on.

### Out of Scope — the originating project

- Changing the moai-cowork project's instructions or persona. The #1718 case is
  evidence about what the parser receives; the project that produced it is not
  a target of this SPEC.
- Committing the raw #1718 bodies or any population body into the tree. Fixtures
  are sanitized reductions only (REQ-CPS-014).

### Out of Scope — deliberately left unchanged

- The adversarial path's unrecognized-body behaviour (REQ-CPS-010) — pending the
  §C.1 operator decision (REQ-CPS-013).
- The convergence layer and the GLM backend.
- The installed moai binary's staleness (`f67d2193f`). It is a note about
  measurement provenance, not a defect this SPEC repairs.
