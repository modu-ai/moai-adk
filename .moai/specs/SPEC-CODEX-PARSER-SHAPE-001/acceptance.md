# SPEC-CODEX-PARSER-SHAPE-001 — acceptance criteria

> Verification layer. Each criterion is a binary-testable Given-When-Then
> scenario. Measured statements trace to `.moai/reports/t1053/verdict.md` by
> section; no figure appears here that is not in that file.

## §A Gating criteria — satisfied BEFORE run-phase entry

These sit ahead of the Implementation Kickoff Approval gate (REQ-CPS-001). The
SPEC does not enter the run phase while either is unsatisfied.

### AC-CPS-001 — live-convention comparison

**Given** the codex account usage limit has reset (2026-09-21 04:21) and live
codex calls succeed again,
**When** a codex review is invoked live through the moai MCP path against a real
target,
**Then** the returned review body is recorded verbatim, and the record states
explicitly whether that body's finding shape is one the current recognizers
accept (same-shape) or is not (different-shape).

[HARD] **What cannot satisfy this criterion.** It is NOT satisfied by:

- the eighteen offline measurements of verdict.md §E3 — those measure parser
  behaviour on bodies the measurer wrote, not codex output;
- the fixture `issue1632ReviewBody` or any other fixture in the tree;
- any test compiled from this tree;
- reading `internal/cli/mcp_codex.go`;
- an inference from the codex CLI version.

The criterion is satisfied only by an observation of **codex's own live output**,
taken after the reset. A record that cites a substitute measurement in place of
that observation leaves the criterion OPEN.

[HARD] **This criterion does not close verdict.md §6's first Gap.** That Gap
records that the convention was unmeasured on the measurement day; that remains
true permanently. AC-CPS-001 records a new measurement at a new point in time.

### AC-CPS-002 — the comparison is recorded, not summarized

**Given** the AC-CPS-001 observation has been taken,
**When** its result is written into the SPEC's evidence record,
**Then** the record carries all four of: the verbatim body, the exact invocation
that produced it, the tree it was measured against, and the codex CLI version —
so that a later reader can re-derive the same-shape / different-shape verdict
rather than taking it on trust.

A one-line "same as before" with no body recorded leaves AC-CPS-002 OPEN.

## §B Selection criterion

### AC-CPS-003 — candidates presented, not pre-selected

**Given** the plan-phase artifacts are complete,
**When** the Implementation Kickoff Approval gate is reached,
**Then** candidates (a), (b), and (c) are all present with their measured
coverage, no one of them has been adopted as the implementation plan, and where
the plan orders them, it states that the ordering criterion is measured
failure-shape coverage (verdict.md §E3, §A2) and nothing else.

## §C Behavioural criteria — conditional on the selected candidate

Only the criteria matching the operator's selection apply; the others are
recorded as not-applicable with the selection as the reason.

### AC-CPS-004 — candidate (b): native disambiguation before downgrade

**Given** candidate (b) is selected,
**When** the native path receives a body carrying no recognized signal,
**Then** the implementation distinguishes "codex found nothing to block on" (the
V9 shape) from "the shape was not recognized" (the V2–V6 shapes) before any
downgrade from `pass`, and a V9-shaped body still yields `pass`.

A change that downgrades both cases identically fails this criterion — that is
the byte-level indistinguishability recorded in verdict.md §E4, not a
verification detail.

### AC-CPS-005 — candidate (c): verdict/findings contradiction is reported

**Given** candidate (c) is selected,
**When** a synthesized review output would carry a blocking verdict together
with an empty findings list — the V8 shape, measured as `fail` / `findings=0` in
both modes (verdict.md §E3),
**Then** that state is reported as self-contradictory rather than emitted as a
clean review, and a consumer reading the output can tell the content was lost.

### AC-CPS-006 — candidate (a): widening parses more without changing what parsed

**Given** candidate (a) is selected,
**When** a review body carries a finding in a shape the widened recognizers
accept,
**Then** that finding is parsed, **and** the findings parsed from the V1 shape
(`- [P1] …`, measured as `fail` / 2 findings in both modes — verdict.md §E3)
are unchanged in count and content.

## §D Preserved-behaviour criteria — apply whichever candidate is selected

### AC-CPS-007 — exact-count assertions stay exact

**Given** the run phase has completed,
**When** `internal/cli/codex_findings_parse_test.go` is inspected,
**Then** every findings-count assertion that was an exact equality before the
change is still an exact equality — none has been relaxed to an inequality such
as `>= 1`.

Basis: under partial shape drift (one bullet of four changed), the exact-count
assertions are what fire, while a "non-empty" style control passes because three
findings remain (verdict.md §4, §A3). The exact count is the measured defence;
loosening it erases the defence silently.

### AC-CPS-008 — a genuinely clean native review is not made loud

**Given** any candidate has been implemented,
**When** the native path receives a body carrying no findings because codex found
nothing to block on,
**Then** the emitted verdict is `pass` — not `inconclusive`.

### AC-CPS-009 — the adversarial path is unchanged

**Given** any candidate has been implemented,
**When** the adversarial path receives a body matching no recognized signal,
**Then** it returns `inconclusive`, as it did before the change (verdict.md §E3)
— no candidate adds or removes work on that path.

### AC-CPS-010 — the closed axis stays closed

**Given** the run phase has completed,
**When** the diff is inspected,
**Then** no change touches the codex synthesis `next_steps` field or any
behaviour card t1052 closed.

## §D.1 Traceability (REQ → AC)

| REQ | AC |
|---|---|
| REQ-CPS-001 (no run-phase entry while unmeasured) | AC-CPS-001 |
| REQ-CPS-002 (comparison recorded verbatim) | AC-CPS-002 |
| REQ-CPS-003 (candidates presented, not chosen) | AC-CPS-003 |
| REQ-CPS-004 (ordering criterion is coverage) | AC-CPS-003 |
| REQ-CPS-005 (candidate (b) disambiguation) | AC-CPS-004 |
| REQ-CPS-006 (candidate (c) contradiction) | AC-CPS-005 |
| REQ-CPS-007 (candidate (a) widening) | AC-CPS-006 |
| REQ-CPS-008 (exact counts stay exact) | AC-CPS-007 |
| REQ-CPS-009 (clean native review stays `pass`) | AC-CPS-008 |
| REQ-CPS-010 (adversarial unchanged) | AC-CPS-009 |
| REQ-CPS-011 (`next_steps` closed) | AC-CPS-010 |

## §E Edge cases

- **Live call succeeds but returns a body with no findings at all.** AC-CPS-001
  is still satisfiable: the record states that no finding shape was observable in
  that body and names that as the limit of what the observation establishes. It
  is NOT recorded as "same shape".
- **Live call is still blocked after the reset.** AC-CPS-001 stays OPEN and the
  SPEC does not enter the run phase. The blockage is recorded with the same
  two-path control used on the measurement day — through moai and through a
  direct `codex exec` call — so that "codex account" and "moai path" are
  distinguishable (verdict.md §E1).
- **The observed live shape is outside the nine measured variants.** The record
  says so. verdict.md §6 already states that the nine variants were enumerated by
  the measurer rather than derived from codex's behaviour; an unenumerated shape
  is expected, not anomalous.

## §F Quality gates

- Package tests for `internal/cli` pass, compiled from the tree — not from the
  installed binary (`f67d2193f`, older than this tree; verdict.md §6).
- `go vet ./internal/cli/...` clean.
- The run-phase evidence record names the command it ran and the output it
  observed for every criterion it marks satisfied.

## §G Definition of Done

- [ ] AC-CPS-001 and AC-CPS-002 satisfied and recorded, BEFORE Kickoff approval
- [ ] AC-CPS-003 satisfied — candidates presented, not pre-selected
- [ ] Selected candidate's criterion (AC-CPS-004 / 005 / 006) satisfied; the
      unselected ones recorded as not-applicable with the selection as the reason
- [ ] AC-CPS-007 through AC-CPS-010 satisfied
- [ ] Every claim in the run-phase evidence traces to a command run in this tree
- [ ] Windows path behaviour and live `audit_multi` appear, if at all, as
      explicitly-unmeasured notes — never as claims
