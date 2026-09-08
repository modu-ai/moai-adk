---
id: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001
title: Blank codex review output must not synthesize a pass
version: 0.1.0
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, audit, fail-open, gate, issue-1632, internal/cli, t551"
issue_number: 1632
tier: M
---

# SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001 — Blank codex review output must not synthesize a pass

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-09-08 | 0.1.0 | Initial draft (card t551, GitHub #1632 blank-output axis). |

---

## §A Context

`codex` is an OPTIONAL audit backend. Its absence is handled by a deliberate
fail-open: the audit returns `inconclusive`, and the session proceeds. That
design is correct and is not under revision here.

What IS under revision is a second, unintended path to a clean verdict. Three
emptiness checks on the review-text path test for EXACT equality with `""`, so a
review body carrying only whitespace or newlines counts as content, passes the
guard, and reaches the verdict synthesizer — where the native review mode's
documented default for an unrecognized body is `pass`. A review that never
produced a verdict is therefore reported as a review that found nothing wrong.

The two states have opposite meanings and must not share a signal:

- **codex could not be reached** — nothing was observed; fail open, say so.
- **codex reviewed and found nothing** — something was observed; `pass` is a
  finding and must be reported as one.

A blank body belongs to the first class and is currently reported as the second.

### Measured evidence (this tree, HEAD `3ac58b5a1`)

Reachability probe driving the real `runCodexReviewRPC` production path against a
stubbed codex connection (`.moai/reports/t551/probe-reachability-20260908.txt`):

| Input review body | verdict | summary |
|---|---|---|
| exactly empty | `inconclusive` | `codex unavailable: codex review produced no verdict text` |
| whitespace-only | `pass` | *(empty)* |
| newline-only | `pass` | *(empty)* |
| real clean review | `pass` | `The change introduces no blocking issues.` |

Rows 2 and 3 are the defect: they are indistinguishable in verdict from row 4,
and they carry an empty summary, so no consumer — human or machine — can tell
that no review happened.

The synthesizer-layer probe (`.moai/reports/t551/probe-synthesizer-20260908.txt`)
confirms the same shape one layer down and shows that the adversarial mode
(`turn/start`) already answers `inconclusive` for a whitespace-only body; only
the native mode (`review/start`) answers `pass`.

### The five implicated sites

| File:line | Expression | Role |
|---|---|---|
| `internal/cli/mcp_codex.go:1134` | `p.Item.Review != ""` | collection — `exitedReviewMode` |
| `internal/cli/mcp_codex.go:1137` | `p.Item.Text != ""` | collection — `agentMessage` |
| `internal/cli/mcp_codex.go:1253` | `review != ""` | selection — `bestCodexReviewText` |
| `internal/cli/mcp_codex.go:817` | `reviewText == ""` | **the guard** — `runTurn` |
| `internal/cli/mcp_codex.go:1409` | `codexUnrecognizedVerdict` | native mode default → `pass` |

The collection and selection sites are load-bearing and are NOT optional extras.
`bestCodexReviewText` (`internal/cli/mcp_codex.go:1252-1257`) prefers the
structured `exitedReviewMode.review` over the free-form `agentMessage.text`
whenever the former is non-empty by exact equality. A blank structured review
therefore SHADOWS a real agent message. Repairing the guard alone would convert
that case — one that has genuine review content available — from `pass` into
`inconclusive`, trading one wrong answer for another. The three sites are one
repair.

### What the resulting `inconclusive` does and does NOT achieve

[HARD] **It does NOT close the gate hole, and this SPEC does not claim it does.**

Traced through `internal/cli/mcp_convergence.go:185-200`: a codex verdict of
`inconclusive` leaves `requiredFails` empty and makes `allPass(required)` false,
so control reaches the `default:` arm → `claudeVerdictOrDefault(verdicts,
overallVerdictPass)` → the CLAUDE anchor's verdict. When claude passes,
`overall_verdict` is `pass` — identical before and after this repair. Making a
`required` codex gate actually BLOCK is #1632 axis 3, a sibling card, and is
explicitly out of scope here (§D).

What the repair does achieve — and this is the correctly-scoped win — is replace
a **fabricated codex pass** with an honest codex "could not tell":

- codex leaves the all-pass arm at `internal/cli/mcp_convergence.go:190`;
- codex is listed in `FailOpenBackends`
  (`internal/cli/mcp_convergence.go:133`, populated at `:280`);
- codex carries `GateUnmet` when the gate is `required`
  (`applyGateUnmet`, `internal/cli/mcp_codex.go:1608`);
- codex's Summary names the blank body as the cause (REQ-CBR-005).

Before the repair, a blank body produced a codex row reading `pass` with an empty
Summary — indistinguishable from a real clean review. After it, the row reads
`inconclusive` and says why. That is a change in what codex REPORTS about itself,
not a change in what the convergence layer decides.

---

## §B Requirements (GEARS)

### Collection layer

**REQ-CBR-001** — When the review-text collector observes an `item/completed`
notification of type `exitedReviewMode` whose `review` field contains no
non-whitespace character, the collector shall treat that field as absent and
shall not record it as the collected review text.

**REQ-CBR-002** — When the review-text collector observes an `item/completed`
notification of type `agentMessage` whose `text` field contains no non-whitespace
character, the collector shall treat that field as absent and shall not record it
as the collected agent text.

### Selection layer

**REQ-CBR-003** — When the review-text selector is offered a structured review
text containing no non-whitespace character, the selector shall fall through to
the agent-message text rather than adopting the blank structured text.

### Guard layer

**REQ-CBR-004** — When a codex turn completes and its collected review text
contains no non-whitespace character, the turn shall return an `inconclusive`
review output and shall not invoke the verdict synthesizer.

**REQ-CBR-005** — The blank-output `inconclusive` summary shall name blank review
output as its cause, in wording distinct from the summary emitted when the codex
backend is unavailable, so that a reader can tell the two states apart.

### Preserved behavior (unwanted-behavior requirements)

**REQ-CBR-006** — While the codex backend is unavailable — binary absent,
handshake refused, request rejected — the audit shall continue to return
`inconclusive` fail-open, and this SPEC shall not alter that path.

**REQ-CBR-007** — The audit shall not change the verdict it derives for a
non-blank body that matches no known verdict signal; the native review mode's
documented `pass` default (`internal/cli/mcp_codex.go:1409-1415`) shall remain in
force for such a body.

**REQ-CBR-008** — The audit shall not report a blank review body as `fail`.

### Gate visibility

**REQ-CBR-009** — Where `workflow.audit.gates.codex` is `required`, when the
audit returns the blank-output `inconclusive`, the audit shall carry the
`GateUnmet` annotation so the unmet gate is visible to a machine consumer.

---

## §C The verdict value for a blank body — decision and argument

**Decision: `inconclusive`. Not `fail`.**

The strict fail-closed reading argues for `fail`: a required gate that produced no
verdict has not been satisfied, and `fail` is the only value that mechanically
blocks. Three considerations outweigh it.

1. **`inconclusive` is already this codebase's answer for the sibling case.** The
   exactly-empty body returns `inconclusive` today at
   `internal/cli/mcp_codex.go:817`. A whitespace-only body differs from an empty
   one by characters that carry no meaning; giving the two different verdicts
   would encode a distinction that does not exist.
2. **`fail` would contradict the optional-backend contract.** codex is optional
   (REQ-CBR-006). A blank body is one symptom of a backend that is present but
   not answering — an installed-but-misbehaving codex, a truncated stream, a
   version whose output shape this server does not yet read. Mapping that to
   `fail` blocks sessions on a backend the project never required.
3. **`inconclusive` is not silent, even though it does not block.** It exits the
   all-pass arm at `internal/cli/mcp_convergence.go:190`, lands codex in
   `FailOpenBackends`, and under a `required` gate picks up the `GateUnmet`
   annotation (`internal/cli/mcp_codex.go:1608`). It does NOT change
   `overall_verdict` when the claude anchor passes — see §A. The gap is
   REPORTED rather than laundered; making it BLOCK is axis 3, a sibling card.
   Reporting honestly is the whole of what this card owes.

The repair is therefore correctly described as *fail-open, but honestly labelled*
— not as fail-closed. The card's shorthand name (`FAILCLOSED`) refers to closing
the blank-output *loophole*, not to adopting a blocking verdict.

---

## §D Exclusions

This SPEC repairs exactly one axis of GitHub issue #1632.
Everything below is out of scope for it.

### Out of Scope — axes of #1632 already closed in this tree

- **Axis 1 — findings parsing.** Already closed by card t234; the `Findings`
  field is filled by `synthesizeReviewOutput`
  (`internal/cli/mcp_codex.go:1432-1435` names t234 explicitly; the per-bullet
  parse is documented at `internal/cli/mcp_codex.go:1466`). Do not reopen.
- **Axis 2 — verdict signal adoption.** Already closed; a rejected request is no
  longer reported identically to a review that passed
  (`internal/cli/mcp_codex.go:1036-1039`), and conservative adoption across
  signals is implemented in `adoptConservativeVerdict`
  (`internal/cli/mcp_codex.go:1386-1394`). Do not reopen.

### Out of Scope — axes of #1632 owned by sibling cards

- **Axis 3 — gates enforcement.** The `required` gate's enforcement semantics
  beyond the existing `GateUnmet` annotation belong to a sibling card. This SPEC
  consumes `applyGateUnmet` as it stands and changes none of it.
- **Axis 4 — `auth_provider` unknown.** Sibling card.
- **Axis 5 — spurious `codex_task` timeout.** Sibling card.

### Out of Scope — behavior deliberately left unchanged

- The mode-keyed policy of `codexUnrecognizedVerdict` for a body that is present
  but unrecognized. Native mode meaning `pass` there is deliberate and documented
  (`internal/cli/mcp_codex.go:1395-1415`). Only ABSENCE is being reclassified.
- The unavailable-backend fail-open path in every one of its forms.
- Any change to the convergence layer's arms
  (`internal/cli/mcp_convergence.go:180-200`).
- Any change to the GLM backend or to `audit_multi` fan-out behavior.

---

## §E Gaps — what this SPEC did NOT observe

- The convergence-layer claim in §A is a **static trace** of
  `internal/cli/mcp_convergence.go:185-200`, read line by line during plan-phase
  audit, NOT a test result. The trace is what §A now states: `inconclusive`
  reaches the `default:` arm and `overall_verdict` follows the claude anchor, so
  the repair does not change `overall_verdict` when claude passes. Run-phase
  should confirm this by execution rather than inherit the reading — but the
  claim in §A is scoped to what the trace supports, and no longer asserts that a
  machine consumer is blocked or that the gate hole is closed.
- No live `codex` binary was exercised. Every measurement in §A comes from a
  stubbed connection driving the production path.
- The real-world frequency of a whitespace-only body from codex-cli was NOT
  measured; the defect is established by reachability, not by field incidence.
