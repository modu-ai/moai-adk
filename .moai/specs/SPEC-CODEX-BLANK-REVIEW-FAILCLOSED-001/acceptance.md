---
id: SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001
title: Acceptance criteria — blank codex review output must not synthesize a pass
version: 0.1.0
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tags: "codex, audit, fail-open, gate, issue-1632, internal/cli, t551"
---

# Acceptance Criteria — SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001

Every criterion below is mechanically checkable by a Go test in package
`internal/cli`.

---

## §A The control matrix — the substance of this repair

Separating "no review happened" from "a review happened and found nothing" IS the
repair. An acceptance suite that observes only the blank case cannot distinguish
a correct repair from one that blocks everything — that is the same
undistinguished-states defect, re-enacted in the test.

[HARD] All three states below MUST be observed in the SAME test run, and MUST
produce three distinguishable signals.

| State | Input | Required verdict | Required distinguishing signal |
|---|---|---|---|
| **A — blank** | review body with no non-whitespace character | MUST NOT be `pass` | Summary names blank review output |
| **B — control: reviewed, zero findings** | real clean review prose | MUST stay `pass` | Summary carries the review prose |
| **C — control: backend unavailable** | codex binary absent / handshake refused | MUST stay `inconclusive` fail-open | Summary names codex unavailability |

A and C are BOTH `inconclusive`; they are separated by their Summary, not by
their verdict. That separation is itself an acceptance criterion (AC-CBR-005).

---

## §B Acceptance criteria

### AC-CBR-001 — blank body does not synthesize a pass (state A)

**Given** a codex turn that completes with a review body containing only
whitespace or newlines
**When** the audit runs through the production `runCodexReviewRPC` path
**Then** the resulting verdict is `inconclusive`, and is NOT `pass`.

Covers REQ-CBR-004. Fixtures: `" "`, `"\n"`, `"\n\t  \n"`, and `""`.

### AC-CBR-002 — real clean review still passes (state B, CONTROL)

**Given** a codex turn that completes with the body
`"The change introduces no blocking issues."`
**When** the audit runs through the same path in the same test
**Then** the verdict is `pass` and the Summary carries that prose verbatim.

Covers REQ-CBR-007. This is the control that proves AC-CBR-001 blocked the right
thing rather than everything.

### AC-CBR-003 — a blank structured review must not shadow a real agent message

**Given** an `exitedReviewMode` item whose `review` is whitespace-only AND an
`agentMessage` item whose `text` is a real clean review
**When** the review-text selector chooses between them
**Then** the agent text is selected and the verdict is `pass` — NOT
`inconclusive`.

Covers REQ-CBR-003 only.

Its discriminating power, stated exactly: this criterion fails when the guard
(`internal/cli/mcp_codex.go:817`) is repaired without the SELECTION site
(`:1253`) — the blank structured review would then be returned by
`bestCodexReviewText`, reach the repaired guard, and produce `inconclusive` where
`pass` is required. `:1253` alone is the load-bearing site here.

It says NOTHING about the collection sites `:1134` / `:1137`. With `:817` and
`:1253` repaired and the collection sites left alone, a whitespace-only review
still passes the `!= ""` test at `:1134` and is stored, but the repaired `:1253`
falls through to the real agent text and the verdict is `pass` — which is what
this criterion requires. The collection sites are covered by AC-CBR-009, not
here.

### AC-CBR-004 — unavailable backend still fails open (state C, CONTROL)

**Given** a codex backend that cannot be reached
**When** the audit runs
**Then** the verdict is `inconclusive` and the pre-existing unavailable-path
behavior is byte-identical to its behavior before this SPEC.

Covers REQ-CBR-006. Establish the "before" side as a characterization test in M1
so this is a measured comparison rather than an assertion.

### AC-CBR-005 — states A and C are distinguishable

**Given** the outputs captured by AC-CBR-001 (state A) and AC-CBR-004 (state C)
**When** their Summary strings are compared
**Then** they differ, the state-A Summary names blank review output, and the
state-C Summary names codex unavailability.

Covers REQ-CBR-005. Assert on the distinguishing substring, not on full-string
equality, so the criterion survives incidental wording changes.

### AC-CBR-006 — no blank body is reported as `fail`

**Given** each blank fixture from AC-CBR-001
**When** the audit runs
**Then** no verdict equals `fail`.

Covers REQ-CBR-008. This pins the §C decision against a later drift to
fail-closed.

### AC-CBR-007 — required gate annotates the blank-output inconclusive

**Given** a project tree whose `workflow.audit.gates.codex` is `required`
**And** a codex turn completing with a whitespace-only body
**When** the audit runs
**Then** the output carries a non-empty `GateUnmet` annotation.

Covers REQ-CBR-009. Exercises `applyGateUnmet`
(`internal/cli/mcp_codex.go:1608`) without modifying it.

### AC-CBR-008 — non-blank unrecognized body is unchanged

**Given** the body `"I looked at the diff."` — present, but matching no known
verdict signal
**When** it is synthesized in native review mode (`review/start`)
**Then** the verdict is `pass`, unchanged from the pre-SPEC behavior recorded in
`.moai/reports/t551/probe-synthesizer-20260908.txt` (`prose-no-signal
verdict="pass"`).

Covers REQ-CBR-007. This is the criterion that stops the repair from widening
into `codexUnrecognizedVerdict`'s documented policy.

### AC-CBR-009 — a later blank item must not clobber an earlier real one

**Given** an `item/completed` stream carrying a REAL `exitedReviewMode.review`
followed by a LATER `exitedReviewMode` item whose `review` is whitespace-only
**When** the turn completes
**Then** the real review text survives and the verdict is `pass`.

**And given** the same stream shape for `agentMessage.text` — a real text
followed by a later blank one
**Then** the real agent text survives.

**And given** the REVERSED ordering for each type — a whitespace-only item
arriving BEFORE a real one of the same type
**Then** the real text is still the value that survives. A first-wins
implementation, which keeps the first non-blank value and ignores later ones,
satisfies the two real-then-blank orderings above while recording the blank and
never replacing it in this one; that mutant violates REQ-CBR-001/REQ-CBR-002 and
no other criterion catches it.

Covers REQ-CBR-001 and REQ-CBR-002.

This is the ONLY behavioral difference the collection sites make. The collection
loop (`internal/cli/mcp_codex.go:1125-1139`) REASSIGNS `reviewText` / `agentText`
on every matching `item/completed`, so today a later whitespace-only item passes
the `!= ""` test at `:1134` / `:1137` and overwrites an earlier real one.
Blank-awareness at those two sites is what makes the real text survive. Without
this criterion, REQ-CBR-001 and REQ-CBR-002 have no coverage at all — AC-CBR-003
does not reach them (see its note).

---

## §C Edge cases

- `""` (exactly empty) — already `inconclusive` before this SPEC
  (`internal/cli/mcp_codex.go:817`); must remain so, and must now share its
  treatment with the whitespace fixtures.
- U+00A0 non-breaking space only — behavior follows the plan §B.2 decision, and
  whichever way it is decided, it is asserted rather than left undefined.
- A body that is whitespace plus a single severity-tagged finding bullet — must
  still reach the synthesizer and produce `fail`; blankness must not swallow real
  findings.
- Both `exitedReviewMode.review` and `agentMessage.text` blank — falls to state
  A.

---

## §D Quality gates

- `go test ./internal/cli/...` passes (no `-run` narrowing in the recorded run).
- `go vet ./internal/cli/...` clean.
- New tests fail against the pre-repair tree (RED established for AC-CBR-001,
  AC-CBR-003, AC-CBR-005, AC-CBR-009) — an absence-shaped criterion is not
  adopted on RED-now alone; the RED must be observed on the unrepaired code.
  AC-CBR-009 belongs on this list precisely because it CAN go red: on the
  pre-repair tree the later blank item overwrites the earlier real one at
  `internal/cli/mcp_codex.go:1134` / `:1137`, so the real text does not survive.

  AC-CBR-003 belongs on this list only if it asserts the SELECTED TEXT, not
  merely the verdict. Its verdict half is already `pass` on the pre-repair tree —
  the blank structured review reaches the synthesizer and the native-mode default
  produces `pass` — so a test asserting `verdict == "pass"` alone is green before
  AND after the repair and proves nothing. To go red pre-repair it MUST assert
  that the value returned by `bestCodexReviewText` is the real agent text: on the
  unrepaired `:1253` that value is the whitespace-only structured review, which
  fails the assertion, and the same green verdict is then reached for the right
  reason rather than by coincidence.
- AC-CBR-002, AC-CBR-004, AC-CBR-006 and AC-CBR-008 are deliberately EXCLUDED
  from the RED list: they are preservation pins that must be green both before
  and after the repair. A pin that goes red pre-repair would mean the repair
  changed behavior it was required to leave alone.
- The pinned case at `internal/cli/codex_review_rpc_test.go:119` is untouched
  (option a) or updated in place with a rationale (option b) — never deleted.

---

## §E Definition of Done

- [ ] AC-CBR-001 .. AC-CBR-009 all PASS, with the command and its verbatim output
      cited.
- [ ] The three-state control matrix (§A) is observed in one run and yields three
      distinguishable signals.
- [ ] The plan §B.1 decision is recorded in `progress.md` §E.2 with its rationale.
- [ ] `spec.md` §E gaps are either closed by measurement or restated as residual
      risk — not silently dropped.
