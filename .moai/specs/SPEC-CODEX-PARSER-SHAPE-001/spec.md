---
id: SPEC-CODEX-PARSER-SHAPE-001
title: "codex review parser — output-shape coupling and its silent failure"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
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

- **The live output convention is unmeasured** (verdict.md §6). Every codex call
  on the measurement day was blocked by an account usage limit (reset
  2026-09-21 04:21), verified twice — once through the moai MCP path and once by
  a direct `codex exec` call that bypasses moai entirely. This SPEC does not
  claim that the current convention equals the fixture, and it does not claim
  that it differs. §D carries the acceptance criterion that closes this.
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

**REQ-CPS-003** — The plan shall present candidates (a), (b), and (c) of §C as
non-exclusive candidates with their measured failure-shape coverage, and shall
not adopt one as the implementation plan; the selection is the operator's at the
Implementation Kickoff Approval gate.

**REQ-CPS-004** — Where the plan orders the three candidates, the ordering shall
be by measured failure-shape coverage (verdict.md §E3, §A2), and the plan shall
state that criterion explicitly rather than leaving the order to be read as a
recommendation.

### Behavioural requirements, conditional on the selected candidate

**REQ-CPS-005** — Where candidate (b) is selected, the native review path shall
distinguish a body carrying no recognized signal because codex found nothing to
block on from a body carrying no recognized signal because its shape was not
recognized, before any body is downgraded from `pass`.

**REQ-CPS-006** — Where candidate (c) is selected, when a synthesized review
output carries a blocking verdict together with an empty findings list, the
adapter shall report that state as self-contradictory rather than emitting it as
a clean review.

**REQ-CPS-007** — Where candidate (a) is selected, when a review body carries a
finding in a shape the widened recognizers accept, the adapter shall parse that
finding, and the findings parsed from the shapes recognized before the widening
shall be unchanged.

### Unwanted behaviour (preserved properties)

**REQ-CPS-008** — The test suite shall not replace an exact-count findings
assertion in `internal/cli/codex_findings_parse_test.go` with an inequality; the
exact count is the measured defence against partial shape drift (§A.5).

**REQ-CPS-009** — A review body carrying no findings because codex found nothing
to block on shall not be reported as `inconclusive` on the native path.

**REQ-CPS-010** — The adversarial path's existing behaviour for an unrecognized
body shall not change; it already returns `inconclusive` (verdict.md §E3), and no
candidate in §C adds work there.

**REQ-CPS-011** — This SPEC shall not alter the codex synthesis `next_steps`
field or any behaviour that card t1052 closed.

---

## §C The three candidates (presented, not chosen)

The three are **not mutually exclusive**; each targets a different measured
failure shape (verdict.md §A2).

The coverage column is split. **Measured** names a shape a probe actually
exercised; **inferred** names a shape the mechanism suggests where nothing ran.
An inferred shape is not evidence and carries no weight.

| Candidate | Failure shape it targets | Measured coverage | Inferred, NOT measured |
|---|---|---|---|
| (a) reduce shape dependence — widen the range of recognized shapes | shapes outside the current recognizers | **V2 and V8**, both modes — a widening of `codexFindingBullet` / `codexFindingLine` to numbered-list markers turned V2 `pass`/0 → `fail`/2 and V8 `fail`/0 → `fail`/1, with V1 and V9 unchanged as regression controls (verdict.md §A5) | V3, V4, V5, V6 — each needs a **different** recognizer change and none was exercised |
| (b) detect a shape change and drop to inconclusive | native's silent `pass` | the failure shape is measured (native V2–V6 are `pass`/0, verdict.md §E3); already implemented on adversarial | the remedy — that a native downgrade is possible without turning V9 into `inconclusive`; blocked on the §A.3 fact 3 disambiguation |
| (c) detect a verdict/findings contradiction | V8 — verdict survives, content empties | **V8**, both modes — the measured V8 output is `fail` with `findings=0` (verdict.md §E3), which is the contradiction itself | — |

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
- **(c)** keys on `verdict == fail && len(findings) == 0` — a state that is
  self-contradictory whatever the body looks like. It therefore catches V8
  **regardless of shape**, including shapes nobody has enumerated, because it
  reads the parser's output rather than any recognizer. That is a difference in
  kind from (a), not in coverage count: (a)'s V8 coverage is bounded by the
  shapes its recognizers were widened for, (c)'s is not bounded at all. What (c)
  does not do is say anything about V2–V6, where no verdict signal survives, so
  there is no contradiction to detect.

---

## §D Acceptance criteria

Full criteria live in `acceptance.md`. One is load-bearing enough to restate
here:

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

### Out of Scope — deliberately left unchanged

- The adversarial path's unrecognized-body behaviour (REQ-CPS-010).
- The convergence layer and the GLM backend.
- The installed moai binary's staleness (`f67d2193f`). It is a note about
  measurement provenance, not a defect this SPEC repairs.
