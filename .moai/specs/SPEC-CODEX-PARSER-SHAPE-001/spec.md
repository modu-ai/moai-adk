---
id: SPEC-CODEX-PARSER-SHAPE-001
title: "codex review parser — output-shape coupling and its silent failure"
version: "0.2.4"
status: completed
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
- 2026-09-26 · v0.2.1 · manager-spec · Repair of the plan-audit iter-1 defects
  D1–D14 recorded in `.moai/reports/t1203/plan-audit.md` (FAIL 0.75). Wording and
  structure only — no new measurement, no candidate chosen. The §A.4 scope is
  re-drawn on all four co-varying axes (project, path, date, CLI version) and
  "differs by project" is demoted to an inference; the candidate ordering is
  counted per failure shape with one test for all four candidates; the
  REQ-CPS-010 question is re-posed as whether the #1718 adversarial outcome is
  acceptable, with no candidate excluded by either answer; the population
  greeting count is corrected to 93 (t1203 §3 as corrected); AC-CPS-011 gains a
  re-executable RED-now, structural property checks, a sanitization check, and
  negative fixtures; the guard classification of AC-CPS-011..014 is stated
  before and after AC-CPS-011 closes. REQ and AC counts are unchanged (15 / 15).
- 2026-09-26 · v0.2.2 · manager-spec · Repair of the plan-audit iter-2 defects
  recorded in `.moai/reports/t1203/plan-audit-iter2.md` (FAIL 0.83). Wording and
  check commands only — no candidate chosen, no requirement changed. AC-CPS-011's
  structural commands move into a verbatim evidence ledger and the table
  transcription rule is deleted (N4-P4); S1 gains a FAIL-statement check and a
  location-link check bound to the declared finding count, with a mutant →
  killing-check table (N5, D4); the "four axes at once" wording in §A.4, §A.6,
  and plan.md is corrected — project and path differ, the #1718 bodies' date
  differs, the population's date range includes 2026-09-21, and the #1718 CLI
  version is not recorded (N1). The v0.2.1 entry above is kept as written; its
  "all four co-varying axes" is superseded by this correction. Also: the
  consequence of a "not acceptable" answer to the REQ-CPS-013 question is stated
  (N2), "single evidence base" becomes "primary evidence base" (N3), and the
  fidelity check gets a named test selector with a non-empty-sweep condition
  (N6). REQ and AC counts are unchanged (15 / 15).
- 2026-09-26 · v0.2.3 · manager-spec · acceptance.md wording fixes reported by
  the run phase (`progress.md` §E.2/§E.3; evidence
  `.moai/reports/t1203/run/n7-mutants.log`, test
  `TestCodex1718_P10RejectsMutants`). P10 is bound to line 1 with a PASS-negative
  companion P10b, and the mutant table gains M3/M4 (N7); mutant scope, regex note,
  and declared-count definition corrected (N8–N10); AC-CPS-013 re-anchored to
  `562126b1f` or a V8-shaped body (a) does not cover; the AC-CPS-011 check 1
  selector's change of meaning after (a) is recorded. Wording only — no
  requirement changed, no candidate decision; REQ and AC counts unchanged
  (15 / 15).
- 2026-09-26 · v0.2.4 · manager-spec · The candidate (b) native disambiguation
  mechanism is authored into the SPEC, per the operator run-resume decision
  recorded in `progress.md` §E.1 (decision taken at worktree `52ab653c8`; the
  record commit is `f1bd21fc4`), resolving the AC-CPS-004 BLOCKED row of
  `progress.md` §E.2. REQ-CPS-005 is amended in place: the native review
  request pins its output format — the (d) family (REQ-CPS-012) applied to the
  native request — so that a body stating `pass` in the pinned form is
  distinguishable from a body carrying no recognized signal, which the native
  fall-through downgrades to `inconclusive`; the target the request reviews is
  unchanged; and the claim that live native codex honours the pin rests on a
  recorded live observation (AC-CPS-016, new). AC-CPS-004 is amended in place
  with two-cell adoption content — its RED-now was re-executed on tree
  `0ff644530` through the committed fixture test and is quoted verbatim in
  acceptance.md — and with the mutant probe that sharpened it. AC-CPS-008
  carries a scope note: its guard fixture bodies state the pinned verdict
  line, so the guard keeps testing REQ-CPS-009 as written. REQ count unchanged
  (15); AC count 15 → 16. No figure is introduced that is absent from the
  evidence files; the one new observation (the RED-now run) is recorded
  verbatim in acceptance.md.

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

The primary evidence base for this SPEC is `.moai/reports/t1053/verdict.md`,
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

**Audit re-probe (v0.2.1).** One field is cited from the plan-audit record
`.moai/reports/t1203/plan-audit.md` § 감사자 재측정 rather than from a verdict
file: the auditor's re-probe of the same bodies on tree `686b75ebb` printed
`gate=""` for ctrlB, a field the t1203 §1 probe did not print. It is used only
where §C says so.

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
   inconclusive" — is therefore **already implemented for adversarial**. As far
   as candidate (b)'s mechanism is concerned, only native is open. That is a
   statement about (b), not about adversarial adequacy: #1718 reports the
   adversarial `inconclusive` itself as the defect (§A.6, §C.1).
3. **Native's `pass` is a declared decision, not an oversight** (verdict.md §E4).
   In native mode, a bullet-less body means codex found nothing to block on. A
   genuinely clean review (V9) and a shape-changed review (V2–V6) are
   indistinguishable to the parser. Candidate (b) therefore cannot be extended to
   native without a prior disambiguation mechanism. (Specified as of v0.2.4:
   REQ-CPS-005's native format pin — the (d) family applied to the native
   request.)
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
  a statement about urgency, not about treatment (verdict.md §A1). This
  observation did not change the candidate space of §C, which the offline
  measurements had narrowed; candidate (d) was added later, in v0.2.0, from the
  #1718 case, not from this observation. It also does not retroactively close
  verdict.md §6's first Gap: that Gap records that the convention was unmeasured
  on the measurement day, which remains permanently true.

  **Scope of that conclusion, re-drawn 2026-09-26 (t1203; v0.2.1).** The
  observation above stands as recorded: one live call with `mode = "native"`
  (review/start), in the moai-adk-go tree, on 2026-09-21, under codex-cli 0.155.1,
  returned the fixture shape (`live-convention-20260921.md` §1, §3). The
  "prospective, not currently active" conclusion holds for exactly that
  combination — **moai-adk-go · native path · 2026-09-21 · codex-cli 0.155.1** —
  and not for codex output in general.

  The #1718 case (§A.6) is a measured counter-observation, and the two
  observations do not hold the project or the path constant. The project
  differs (moai-cowork against moai-adk-go). The path differs as #1718 reports
  it — adversarial (turn/start) for its two bodies, against the native path of
  2026-09-21; for the population sessions the path is a Gap (§A.6). The date
  differs for the two #1718 bodies, whose codex sessions are dated 2026-09-25
  (t1203 §1), but it is not a clean separator for the population: that range,
  2026-09-14 .. 2026-09-25, includes 2026-09-21 (t1203 §3). The codex CLI
  version is recorded for 2026-09-21 (codex-cli 0.155.1) and is not recorded for
  #1718 or the population, so whether it differs is not established. **Which
  axis separates the two observations is unmeasured.** "The project decides the
  shape" is an inference, not a measurement.

  For the moai-cowork bodies the risk is measured as active in one specific
  sense: across 142 codex-gate final bodies not one parser run produced a
  finding (t1203 §3) — measured on the session-log final assistant messages,
  whose byte-identity to the `reviewText` the parser actually received is a Gap
  (§A.6). That the shape-determining variable is the target project's
  instructions is an **inference**, not a measurement (t1203 §4, §5).

- **The moai-adk-go adversarial live convention has never been measured.** The
  only live observation taken in this repository was on the native path
  (`live-convention-20260921.md` §1). Whether this repository's adversarial
  audits already yield `inconclusive`/0 on live bodies is unknown, and nothing in
  this SPEC may be read as saying they do not.

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
**V8 shape (§A.3 fact 4)**. ctrlB is a **derived body**: body2 with one greeting
token removed by the measurer. Codex did not emit it. What it adds over V8 is
that the finding markup it carries is the markup codex actually emitted, where
the V8 body was written by the measurer from scratch; it is an instance of the
same failure shape, not a new one.

**Population (t1203 §3).** The final assistant message of each of 142
`moai-codex-gate` sessions whose working directory was the moai-cowork project
(2026-09-14 .. 2026-09-25), run through the same parser in both modes:

| Path | Outcome distribution | Status |
|---|---|---|
| turn/start | `inconclusive`/0 = 138, `pass`/0 = 3, `fail`/0 = 1 | measured |
| review/start | `pass`/0 = 141, `fail`/0 = 1 | measured |
| findings > 0 | 0 of 284 runs | measured |

Text counts over the same 142 bodies (grep): first line opening with the persona
greeting **93**, line-head verdict label 4, `[P<digit>]` bullets 0, a `fail`
string present 124 (t1203 §3) — measured. The 93 is the first-line count as
corrected in t1203 §3 on 2026-09-26, whose recorded command runs `head -1` on
each body and matches the greeting at line head (the first non-empty line gives
93 as well). An earlier figure of 101 in the same section counts something
else — files that have **any** line opening with the greeting (`grep -l`) — and
is not a first-line count.

**What this establishes, and what it does not.**

- #1718 reproduces on the adversarial path; it is not an isolated incident in
  that project (measured, t1203 §1, §3).
- The two live observations differ: the 2026-09-21 moai-adk-go call was
  same-shape (§A.4), the moai-cowork bodies are not (measured, t1203 §1, §3).
  **What separates them is not measured.** Project and path (native in
  2026-09-21, adversarial as #1718 reports it) differ, and so does the date of
  the two #1718 bodies; the population's date range includes 2026-09-21, and the
  codex CLI version is not recorded for #1718, so neither date nor version is
  established as a separating axis. "The live convention differs by project" is
  therefore an **inference**, confounded with path and not separated from date
  or version (§A.4).
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
be by measured failure-shape coverage — counted per distinct failure shape, a
shape counting for a candidate only where a probe executed that candidate's
remedy on it — applied identically to every candidate (verdict.md §E3, §A2, §A5;
t1203 §1), and the plan shall state that criterion and its unit explicitly
rather than leaving the order to be read as a recommendation.

**REQ-CPS-013** — When the Implementation Kickoff Approval gate is reached, the
plan shall present to the operator the question of §C.1 — whether the #1718
adversarial outcome, `inconclusive` with an empty findings list for a body whose
prose states FAIL, is acceptable — together with what each candidate changes on
the adversarial path, and shall not answer it itself; the operator's answer,
whether REQ-CPS-010 is kept as written or revised, and the reason given shall be
recorded in the progress record before the first run-phase commit of any
selected candidate.

### Behavioural requirements, conditional on the selected candidate

**REQ-CPS-005** — Where candidate (b) is selected, the native review request
shall specify an output format the recognizers of §A.2 accept — a first-line
verdict statement in the form `codexStatedVerdict` reads, and findings as
bracketed-severity bullets — without changing which changes the request asks
codex to review; and the native review path shall distinguish, before any body
is downgraded from `pass`, a body that states `pass` in that pinned form —
codex found nothing to block on, and it remains `pass` — from a body carrying
no recognized signal — its shape was not recognized, or the pin was not
followed — which the native fall-through reports as `inconclusive` rather than
as a silent `pass`.

> **Mechanism note (v0.2.4).** The disambiguation rides on what the request
> asks for, because §E4 (t1053 verdict) records that V9 and V2–V6 are
> byte-indistinguishable to a parser that reads the body alone. The mechanism
> is the (d) family (REQ-CPS-012) applied to the native request: once the
> request names the format, `Verdict: pass` is a recognized signal
> (`codexStatedVerdict` already reads it), so the pinned clean review never
> reaches the fall-through. The input classes and their outcomes:
>
> - **V1 and the (a)-widened shapes** — recognized before and after; unchanged
>   (REQ-CPS-007, AC-CPS-006).
> - **V9, following the pin** — states `Verdict: pass`; recognized; stays
>   `pass`. This is REQ-CPS-009's protected class.
> - **V2–V6-class bodies carrying no recognized signal on native** —
>   downgraded from silent `pass` to `inconclusive`. The review text remains
>   in `Summary` verbatim, so the content is lost from `findings` only, never
>   from the output.
> - **Bodies whose prose states FAIL in a shape no recognizer reads** — the
>   same downgrade. Making the FAIL recognizable is (a)'s work, already landed
>   for the measured shapes.
>
> **Failure mode (stated, not hidden).** A body that ignores the pin —
> including a genuinely clean review — carries no recognized signal and is
> downgraded. The downgrade is therefore only as sound as the live evidence
> behind the pin. The fall-through is keyed on the absence of a recognized
> signal, and on nothing in the prose: no token (the word `fail`, a severity
> word, a greeting) may key it.
>
> **Live burden.** The claim that the live native review honours the pinned
> format shall rest on a live observation recorded verbatim (AC-CPS-016), not
> on a fixture, not on a test compiled from this tree, and not on reading the
> request text — the same exclusions REQ-CPS-012 states for the adversarial
> pin.

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
that carry no absolute user path and no content of the originating project — as
decided by the sanitization check of AC-CPS-011, not by inspection — and each
reduction shall first be shown to reproduce, on the unmodified parser, the
output measured for its raw body (t1203 §1).

### Unwanted behaviour (preserved properties)

**REQ-CPS-008** — The test suite shall not replace an exact-count findings
assertion in `internal/cli/codex_findings_parse_test.go` with an inequality; the
exact count is the measured defence against partial shape drift (§A.5).

**REQ-CPS-009** — A review body carrying no findings because codex found nothing
to block on shall not be reported as `inconclusive` on the native path.

> **Scope note (v0.2.4; not a revision).** With the REQ-CPS-005 mechanism, the
> way a native review states "nothing to block on" is the pinned verdict line,
> and this requirement binds every body that states `pass` in that form. A body
> that ignores the pin states nothing recognizable; (b) downgrades it. That
> downgrade is the mechanism's stated failure mode, not a breach of this
> requirement — and it is why AC-CPS-016's live observation is load-bearing.
> The AC-CPS-008 guard's fixture bodies state the pinned line, so the guard
> keeps testing this requirement as written.

**REQ-CPS-010** — The adversarial path's existing behaviour for an unrecognized
body shall not change; it already returns `inconclusive` (verdict.md §E3), and no
candidate in §C adds work there.

> **Conflict note (v0.2.0, re-stated v0.2.1; not a revision).** REQ-CPS-010's
> text above is unchanged. It has two parts. Its **normative clause** — the
> adversarial path's handling of an unrecognized body shall not change — binds
> only what happens to a body no recognizer accepts. Its **rationale clause** —
> "it already returns `inconclusive`, and no candidate in §C adds work there" —
> is where #1718 bites: the `inconclusive` it relies on is exactly the output
> #1718 reports as the defect (t1203 §4), and "no candidate adds work there" is
> no longer accurate — a measured (a) widening already changed adversarial output
> (adversarial V2: `inconclusive` before, verdict.md §E3; `fail`/2 after the
> numbered-list widening, verdict.md §A5 — "adversarial 도 같은 값"),
> and (d) would change what the adversarial request asks for. None of those
> changes the handling of a body that stays unrecognized. Whether the #1718
> outcome is acceptable, and so whether REQ-CPS-010 is kept or revised, is the
> operator's question at the Kickoff gate (§C.1, REQ-CPS-013); this SPEC does not
> answer it.

**REQ-CPS-011** — This SPEC shall not alter the codex synthesis `next_steps`
field or any behaviour that card t1052 closed.

---

## §C The candidates (presented, not chosen)

The candidates are **not mutually exclusive**; each targets a different measured
failure shape (verdict.md §A2). (a)–(c) come from the t1053 measurements;
(d) was added in v0.2.0 from the #1718 case (t1203 §4) and is entirely inferred.

**How coverage is recorded (v0.2.1; one unit and one test for all four).** The
unit is the **failure shape**, not the body instance: the §E3 V8 body and ctrlB
are two instances of one shape (V8). For every candidate and every shape, the
same two questions are asked — was the shape's failure measured on the
unmodified parser, and was the candidate's **remedy executed** on it by a probe.
A shape counts toward a candidate's coverage only when both hold. Two further
grades are recorded and carry **no weight** in any ordering:

- **Deduced** — the remedy was never executed, but its output on the shape is
  fully determined by fields a probe measured (a predicate evaluated on paper
  over measured values). The same test is applied to all four candidates.
- **Inferred** — neither executed nor determined by measured fields.

| Candidate | Failure shape it targets | Remedy executed on (counts) | Deduced, not executed (no weight) | Inferred, NOT measured (no weight) |
|---|---|---|---|---|
| (a) reduce shape dependence — widen the range of recognized shapes | shapes outside the current recognizers | **V2 and V8**, both modes — a widening of `codexFindingBullet` / `codexFindingLine` to numbered-list markers turned V2 `pass`/0 → `fail`/2 and V8 `fail`/0 → `fail`/1, with V1 and V9 unchanged as regression controls (verdict.md §A5) | none — no other shape's outcome is fixed by a measured field; each needs a recognizer change whose behaviour is not yet written | V3, V4, V5, V6 — each needs a **different** recognizer change and none was exercised; the #1718 shapes — see the #1718 table below |
| (b) detect a shape change and drop to inconclusive | native's silent `pass` | none — the failure shape is measured (native V2–V6 are `pass`/0, verdict.md §E3), but no native remedy has run; on adversarial the mechanism already exists and targets no failure there | none — the native remedy's output depends on a disambiguation mechanism that does not exist (§A.3 fact 3), so no measured field determines it | that a native downgrade is possible without turning V9 into `inconclusive` |
| (c) detect a verdict/findings contradiction | V8 — verdict survives, content empties | none — (c) has never been implemented or run | **V8**, both modes — the predicate `verdict == fail && len(findings) == 0 && GateUnmet == ""` evaluated over measured parser outputs: the §E3 V8 body gives `fail`/0 (verdict.md §E3) and ctrlB gives `fail`/0 (t1203 §1) with `gate=""` (auditor re-probe, §A.1); both are instances of the one V8 shape | V8 instances in shapes nobody has enumerated (the predicate reads the output, not the body — §C trade-offs) |
| (d) pin the output format in the adversarial prompt | adversarial bodies shaped by the target project's instructions (§A.6) | none — the failure shape is measured (t1203 §1, §3), but no remedy ran | none — (d) acts on what codex emits, which no measured field determines | that codex honours a format instruction over a project persona instruction; that the pinned format survives across projects |

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
  v0.2.4: the native disambiguation mechanism is now specified — REQ-CPS-005
  pins the native request's output format, the (d) family applied to the native
  request. Its coverage stays in the "Inferred" column until the remedy runs
  (plan.md §F M4) and the live native observation (AC-CPS-016) exists; the
  mechanism's stated failure mode is the downgrade of a clean review whose
  body ignores the pin. No count in the table above changes: nothing has run.
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

  Scoped that way, (c) would catch V8 — deduced from the predicate; (c) has not
  been executed — **regardless of shape**, including shapes nobody has
  enumerated, because it
  reads the parser's output rather than any recognizer. That is a difference in
  kind from (a), not in coverage count: (a)'s V8 coverage is bounded by the
  shapes its recognizers were widened for, (c)'s is not bounded at all. What (c)
  does not do is say anything about V2–V6, where no verdict signal survives, so
  there is no contradiction to detect.
- **(d)** changes what codex is asked to emit rather than what the parser
  accepts. Entirely inferred: no prompt change was exercised. Three costs. First,
  it changes what the adversarial request asks for. It does not change how a body
  that stays unrecognized is handled, so REQ-CPS-010's normative clause does not
  bind it; it does make that requirement's rationale clause inaccurate (§C.1).
  Second, its effect depends on
  codex obeying a format instruction against the target project's own
  instructions, which is unmeasured (t1203 §4, §5). Third, it does nothing for
  the parser: native's silent `pass` and any body that ignores the instruction
  are exactly as exposed as before. It composes with (c) rather than replacing
  it — a body that follows the pinned verdict line but not the pinned finding
  format is V8 again.

### Coverage against the #1718 real case (t1203 §1, §4)

Rows are candidates; columns are the #1718 inputs of §A.6. The **parser outputs**
quoted in the cells are measured (t1203 §1). What a candidate would do with them
is never measured here — no candidate was executed on these inputs; each cell
says whether its effect is **deduced** (fixed by measured fields) or
**inferred**, using the same grades as the table above.

| Candidate | body1 raw (adversarial) | body2 raw (adversarial) | ctrlB — derived body: body2 minus one greeting token |
|---|---|---|---|
| (a) | inferred — needs **three distinct recognizer changes**, each unexercised: a relaxed line-head anchor for the verdict, a localized verdict label (`판정`), and a findings recognizer for tables and bold severity-word bullets. The line-head anchor is a documented narrowness contract in `codexStatedVerdict`'s comment, so relaxing it is a regression risk of its own | inferred — same three changes, less the localized label | inferred — needs the findings recognizer change only (the verdict is already recognized: `fail`/0 measured) |
| (b) | nothing to add: the current adversarial output is already `inconclusive`/0 (measured), and #1718 reports **that** output as the defect | same (measured `inconclusive`/0) | not applicable — a verdict is recognized (`fail`/0, measured), so no shape-change fall-through occurs |
| (c) | deduced: **catches nothing** — the measured output is `inconclusive`/0, so the predicate's `verdict == fail` term is false; no blocking verdict survives, so there is no contradiction | deduced: **catches nothing** — measured `inconclusive`/0 | deduced: **would catch it** — the measured output `fail`/0 in both modes with `gate=""` satisfies the predicate; (c) itself was not run |
| (d) | inferred only — would act before the body exists, by changing it | inferred only | inferred only |

On the native path the raw bodies yield `pass`/0 (measured, t1203 §1) — the
§A.3 fact 1 silent pass — but which path the #1718 calls took is a Gap
(§A.6); #1718 itself reports the adversarial path.

**Net, stated plainly.** On the raw #1718 bodies no candidate has an executed
remedy, and none has a deduced catch. The only cell in which a candidate is
deduced to catch anything is (c) on ctrlB — a derived body codex did not emit.

### C.1 The REQ-CPS-010 conflict — an operator decision, not resolved here

**Which clause conflicts with what.**

| Part of REQ-CPS-010 | What it binds | Relation to #1718 and the candidates |
|---|---|---|
| Normative clause — "the adversarial path's existing behaviour for an unrecognized body shall not change" | only the handling of a body that no recognizer accepts | no candidate changes that handling (see the table below); AC-CPS-009's check — a body matching no recognized signal returns `inconclusive` — is expected to stay true under each (inferred for (a)'s unwritten widenings, (c), and (d); none was executed) |
| Rationale clause — "it already returns `inconclusive` … and no candidate in §C adds work there" | nothing on its own; it is the stated reason | #1718 reports that `inconclusive` as the defect — its reporter wrote that the structured output cannot be used as an audit gate (t1203 §4); and "no candidate adds work there" is no longer accurate (conflict note under REQ-CPS-010) |

**The question the operator answers (REQ-CPS-013).** Is the #1718 adversarial
outcome — `inconclusive` with an empty findings list, for a body whose prose
states FAIL — acceptable? Per REQ-CPS-003's "presented, not chosen" rule, this
SPEC presents the two answers and chooses neither, and it excludes no candidate
under either answer:

| Answer | Consequence for REQ-CPS-010 and AC-CPS-009 |
|---|---|
| Acceptable — the outcome is a correct adversarial fall-through | REQ-CPS-010 is kept as written; AC-CPS-009 applies unchanged. A body that stays unrecognized keeps returning `inconclusive`; candidates that make more bodies recognizable remain selectable |
| Not acceptable — the outcome is a defect to be removed | REQ-CPS-010 is revised, because what then changes is the handling of an unrecognized body itself; the revised text and a matching revision of AC-CPS-009 are authored at that decision — not pre-written here |

**What each candidate changes on the adversarial path** — stated so the
operator's answer is not read as a candidate choice:

| Candidate | Adversarial-path change | Handling of a body that stays unrecognized | Status |
|---|---|---|---|
| (a) | more bodies become recognized, and their output changes from `inconclusive` to a stated verdict and findings — measured for the numbered-list widening (V2, V8; verdict.md §A5) | unchanged | measured for one widening; inferred for any other |
| (b) | none — the mechanism already exists on adversarial | unchanged | measured (existing behaviour, verdict.md §E3) |
| (c) | a recognized `fail` with no findings and no `GateUnmet` would be reported as self-contradictory | unchanged — an unrecognized body yields `inconclusive`, which the predicate does not match | deduced from the predicate; (c) not executed |
| (d) | the request asks codex for a recognizable format; if codex follows it, bodies become recognizable | unchanged | inferred |

Under "Not acceptable", none of (a)–(d) as described changes the handling of a
body that stays unrecognized (third column), so none of them alone satisfies the
revised requirement; the revision names the change that implements it, which
may be added alongside any candidate. This states a consequence of the answer,
not a preference among candidates: no candidate is excluded under either answer.

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
- Committing a live body recorded under AC-CPS-014. It is kept verbatim only at
  the gitignored evidence path that criterion names; the committed record carries
  its digest and the synthesized counts, not its text.

### Out of Scope — deliberately left unchanged

- The adversarial path's unrecognized-body behaviour (REQ-CPS-010) — pending the
  §C.1 operator question (REQ-CPS-013).
- The convergence layer and the GLM backend.
- The installed moai binary's staleness (`f67d2193f`). It is a note about
  measurement provenance, not a defect this SPEC repairs.
