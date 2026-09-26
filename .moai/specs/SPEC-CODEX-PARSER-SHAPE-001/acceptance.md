# SPEC-CODEX-PARSER-SHAPE-001 — acceptance criteria

> Verification layer. Each criterion is a binary-testable Given-When-Then
> scenario. Measured statements trace to `.moai/reports/t1053/verdict.md` by
> section; no figure appears here that is not in that file.
>
> v0.2.0 (card t1203): criteria for the #1718 real case (§C.1, AC-CPS-011..014)
> and the REQ-CPS-010 decision (AC-CPS-015) trace to
> `.moai/reports/t1203/verdict.md` (tree `df526c9a9`), cited as `t1203 §N`. The
> same rule binds it.

## §A Gating criteria — satisfied BEFORE run-phase entry

These sit ahead of the Implementation Kickoff Approval gate (REQ-CPS-001). The
SPEC does not enter the run phase while either is unsatisfied.

**Both are SATISFIED.** The codex usage limit reset and the live call ran on
2026-09-21; the record carrying all four AC-CPS-002 items is
`.moai/reports/t1053/live-convention-20260921.md`. Result: **same-shape** — the
live body carries `- [P1] <message> — <path>:<line>` bullets and the parser
structured three of them with severities P1/P2/P2, read off that call's
`findings` array. Tree `a5c3f5dc6`, codex-cli 0.155.1 (the same version as the
blocked day, so version is not a variable). Two tolerated sub-shape differences —
a line RANGE whose end is discarded, and an absolute path — are recorded in
spec.md §A.4. The criteria below are retained as written because they govern any
re-measurement.

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
**Then** candidates (a), (b), (c), and (d) are all present with their measured
coverage, no one of them has been adopted as the implementation plan, and where
the plan orders them, it states that the ordering criterion is measured
failure-shape coverage (verdict.md §E3, §A2; t1203 §1) and nothing else.

### AC-CPS-015 — the REQ-CPS-010 question is decided by the operator and recorded

**Given** the plan-phase artifacts present the REQ-CPS-010 conflict (spec.md
§C.1) with its two options — keep REQ-CPS-010 as written, or revise it,
**When** the Implementation Kickoff Approval gate is passed,
**Then** `progress.md` carries a line beginning `- REQ-CPS-010 decision:` naming
the option the operator chose and the reason given, and that line exists before
the first run-phase commit of any candidate that changes adversarial-path
behaviour ((d), or an (a) widening for the #1718 shapes).

A plan-side or agent-side resolution — a line written without the operator's
choice behind it, or a candidate implemented on the adversarial path before the
line exists — fails this criterion.

- **RED-now.** Command
  `grep -c '^- REQ-CPS-010 decision:' .moai/specs/SPEC-CODEX-PARSER-SHAPE-001/progress.md`;
  stdout `0`; exit code `1`; measured on the working tree of this amendment
  (parent `e0eab2c68`), where §E.1 records the decision as pending and names no
  option. Red for the stated reason: no decision has been taken.
- **Green path.** M1 (Kickoff) — the operator's answer is recorded; the same
  command then prints `1` with exit `0`.

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

### AC-CPS-005 — candidate (c): verdict/findings contradiction is reported, and the unmet-gate state is NOT

**Given** candidate (c) is selected,
**When** a synthesized review output would carry a blocking verdict together
with an empty findings list **and an empty `GateUnmet`** — the V8 shape,
measured as `fail` / `findings=0` in both modes (verdict.md §E3),
**Then** that state is reported as self-contradictory rather than emitted as a
clean review, and a consumer reading the output can tell the content was lost.

**Control case (mandatory, REQ-CPS-006a).**
**Given** candidate (c) is selected,
**When** a review output carries `Verdict == "fail"` with an empty findings list
**and a non-empty `GateUnmet`** — the state `applyGateUnmet`
(`internal/cli/mcp_codex.go`) produces when `workflow.audit.gates.codex` is
`required` and the audit returned a fail-open `inconclusive` with
`Findings: []Finding{}`,
**Then** that state is **NOT** reported as self-contradictory.

The control is mandatory rather than advisory because the bare two-term
predicate (`verdict == fail && len(findings) == 0`) matches this legitimate
state exactly: an unmet required gate is a correctly functioning gate, and
flagging it as a parser contradiction would convert a true signal into a false
defect report. Without a test pinning it, an implementation can regress into
that behaviour and every V8 test still passes.

### AC-CPS-006 — candidate (a): widening parses more without changing what parsed

**Given** candidate (a) is selected,
**When** a review body carries findings in a shape the widened recognizers
accept,
**Then** the parsed findings match the fixture's **exact expected count AND
content** — severity, message, file, and line for every finding — **and** the
findings parsed from the V1 shape (`- [P1] …`, measured as `fail` / 2 findings
in both modes — verdict.md §E3) are unchanged in count and content.

**Partial-drift case (mandatory).**
**Given** candidate (a) is selected and a widened fixture carrying N findings,
**When** only some of those findings are rewritten into a shape the widened
recognizers do NOT accept,
**Then** the exact-count assertion fails — it does not pass on the surviving
subset.

[HARD] **This is §A.5's lesson applied to a new criterion, and it was originally
written the wrong way.** The first draft of this AC required only that "that
finding is parsed" — singular, no count. A two-finding body that parses one and
silently drops the other satisfied it, which is precisely the partial-drift
failure spec.md §A.5 records: a non-empty-style assertion passes while content is
lost, and only an exact count fires. The V1 unchanged-count clause does not cover
it, because V1 is the OLD shape and the drift happens in the NEW one. Every AC
this SPEC adds for a parsed-findings shape states an exact count; an author
tempted to write "at least one is parsed" is re-creating the defect this SPEC
exists to close.

## §C.1 #1718 real-case criteria — conditional on a candidate being claimed for #1718

Only the criteria matching the operator's selection apply; the others are
recorded as not-applicable with the selection as the reason. None of them may be
satisfied by the raw #1718 bodies: those carry another project's absolute paths
and content and are never committed (REQ-CPS-014).

**The shared RED-now observation (evidence ledger E-1718).** Every criterion
below cites this ledger entry as its RED-now cell.

```
id:       E-1718
command:  go test -count=1 -v -run '^TestT1203Probe$' ./internal/cli/
tree:     df526c9a9
exit:     0
stdout (verbatim lines, t1203 repro/probe.log; ctrlA lines summarized below):
PROBE body1 turn/start verdict=inconclusive findings=0 note="" findingsJSON=[]
PROBE body1 review/start verdict=pass findings=0 note="" findingsJSON=[]
PROBE body2 turn/start verdict=inconclusive findings=0 note="" findingsJSON=[]
PROBE body2 review/start verdict=pass findings=0 note="" findingsJSON=[]
PROBE ctrlB turn/start verdict=fail findings=0 note="" findingsJSON=[]
PROBE ctrlB review/start verdict=fail findings=0 note="" findingsJSON=[]
positive control: ctrlA turn/start and review/start both verdict=fail findings=1
```

[HARD] **Disposition of E-1718.** The probe that produced it was a temporary test
file deleted after measurement, reading bodies from a session scratch path
(t1203 §1). It therefore cannot be re-executed on the current tree as written, so
per verification-completeness §2.1 the criteria below are **regression-guard, not
release-blocking**, and none is recorded as a pass on the strength of E-1718.
AC-CPS-011 exists to end that state: once the sanitized reductions are committed
and observed red on the pre-implementation tree, that observation replaces E-1718
as the RED-now cell of AC-CPS-012 and AC-CPS-013.

### AC-CPS-011 — the sanitized reductions reproduce the measured raw output

**Given** a candidate is claimed to address #1718,
**When** sanitized reductions of both #1718 shapes are committed as fixtures — S1
(persona greeting opening the verdict line, a localized verdict label, findings in
a markdown table with bold severity words) and S2 (persona greeting followed by a
bold `verdict: fail`, findings as bold severity-word bullets carrying a
`[path:line](<…>)` link) — carrying no absolute user path and no content of the
originating project, plus S2′ (S2 with the greeting prefix removed, the ctrlB
analogue),
**Then**, on the parser **before** any candidate change, S1 and S2 each yield
`inconclusive` with `findings=0` on turn/start and `pass` with `findings=0` on
review/start, and S2′ yields `fail` with `findings=0` on both paths — the outputs
E-1718 recorded for body1, body2, and ctrlB — recorded verbatim with the tree SHA.

A reduction that does not reproduce its raw body's measured output is not a
reduction of that shape, and no candidate claim may rest on it.

- **RED-now.** E-1718 (the raw bodies' output the reductions must match). The
  reductions themselves are not yet measured; their own first observation is the
  green path below.
- **Green path.** M2 first step, before any recognizer or prompt change: the
  fixture test is committed and its pre-change output is recorded; this AC closes
  when that output matches E-1718 row for row.

### AC-CPS-012 — candidate (a) on the #1718 shapes: exact verdict, count, and content

**Given** candidate (a) is selected for the #1718 shapes and AC-CPS-011 holds,
**When** S1, S2, and S2′ are synthesized on turn/start and on review/start,
**Then** each yields verdict `fail` and a findings list whose **exact count and
content** — severity, message, file, and line for every finding — match the
fixture's declared expectation; the V1 and V9 shapes are unchanged in count and
content (AC-CPS-006); and a prose sentence that mentions a verdict without
stating one is still not read as a verdict (the narrowness contract in
`codexStatedVerdict`'s comment).

**Partial-drift case (mandatory).** When one finding of S1 or S2 is rewritten
into a shape the widened recognizers do not accept, the exact-count assertion
fails; it does not pass on the surviving subset (§A.5, AC-CPS-006).

- **RED-now.** E-1718: the raw counterparts of S1/S2 yield `inconclusive`/0 and
  `pass`/0, that of S2′ yields `fail`/0 — red against a `fail`/N expectation
  because neither the greeting-prefixed verdict nor the table / bold-severity
  findings are recognized (t1203 §2).
- **Green path.** M2 under (a): the three recognizer changes of spec.md §C (#1718
  table) turn the fixture test green; each is measured, none inherited from the
  numbered-list result (plan.md §F M1).

### AC-CPS-013 — candidate (c) on the #1718 shapes: what it catches, and what it does not

**Given** candidate (c) is selected and AC-CPS-011 holds,
**When** S1, S2, and S2′ are synthesized with an empty `GateUnmet`,
**Then** S2′ is reported as self-contradictory on both paths (its output is
`fail` with `findings=0`); S1 and S2 are **not** reported as self-contradictory,
because no blocking verdict survives on them; and the run-phase record states that
(c) alone leaves the raw #1718 shapes at `inconclusive`/0 (turn/start) and
`pass`/0 (review/start).

The second half is load-bearing: it keeps (c)'s ctrlB coverage from being
reported as #1718 coverage (plan.md §G anti-pattern 6).

- **RED-now.** E-1718: ctrlB yields `fail`/0 on both paths with nothing marking
  it contradictory — red because (c) does not exist yet.
- **Green path.** M2 under (c): the contradiction report appears on S2′ and only
  there; the AC-CPS-005 control case still holds.

### AC-CPS-014 — candidate (d): the pinned format is observed in live output

**Given** candidate (d) is selected and REQ-CPS-010 has been revised
(AC-CPS-015),
**When** a live adversarial codex review is invoked through the moai MCP path
against a target known to produce findings, under project instructions that
previously produced an unrecognized shape,
**Then** the returned body is recorded verbatim together with the invocation, the
tree, and the codex CLI version, and its synthesized output carries a verdict
other than the unrecognized-body fall-through and a findings list whose exact
count matches the findings stated in that body.

[HARD] **What cannot satisfy this criterion** — the same exclusions as
AC-CPS-001: a fixture, any test compiled from the tree, reading the prompt text,
or an inference from the codex CLI version. (d) acts on what codex emits, so only
codex's own live output can show it worked.

- **RED-now.** The population observation (t1203 §3): of 142 moai-cowork
  codex-gate final bodies, turn/start gave `inconclusive`/0 for 138, and no run
  produced a parsed finding (0 of 284). Recorded in t1203 `repro/pop.log`
  (tree `df526c9a9`, exit 0); same disposition as E-1718 — the probe is not in
  the tree. Which path each session took is a Gap (spec.md §A.6).
- **Green path.** M2 under (d), then one live call recorded as above.

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

**Scope (v0.2.0).** This criterion applies as written only if the operator keeps
REQ-CPS-010 at the AC-CPS-015 decision. If REQ-CPS-010 is revised, this criterion
is re-authored at that decision together with the revised requirement; it is not
pre-written here, and it is not silently dropped.

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
| REQ-CPS-006 (candidate (c) contradiction, `GateUnmet == ""` scoped) | AC-CPS-005 |
| REQ-CPS-006a (unmet gate is NOT a contradiction) | AC-CPS-005 control case |
| REQ-CPS-007 (candidate (a) widening) | AC-CPS-006 |
| REQ-CPS-008 (exact counts stay exact) | AC-CPS-007 |
| REQ-CPS-009 (clean native review stays `pass`) | AC-CPS-008 |
| REQ-CPS-010 (adversarial unchanged) | AC-CPS-009 |
| REQ-CPS-011 (`next_steps` closed) | AC-CPS-010 |
| REQ-CPS-012 (candidate (d) format pin, live-observed) | AC-CPS-014 |
| REQ-CPS-013 (REQ-CPS-010 conflict is the operator's decision) | AC-CPS-015 |
| REQ-CPS-014 (#1718 claims rest on sanitized, fidelity-checked reductions) | AC-CPS-011, AC-CPS-012, AC-CPS-013 |

## §E Edge cases

- **Live call succeeds but returns a body with no findings at all.**
  [HARD] **AC-CPS-001 stays OPEN**, and the observation is retried against a
  target known to produce findings. A body carrying no finding shape does not
  establish that the finding shape is unchanged — it establishes nothing about
  the finding shape at all, which is the exact state the criterion exists to
  detect. The record states that no finding shape was observable and names that
  as the limit of what the observation establishes; it is NOT recorded as
  "same shape", and — the load-bearing half — **it does not close the
  criterion.** Allowing the gate to close here would let the SPEC enter the run
  phase with the convention still unmeasured.
- **Live call is still blocked after the reset.** AC-CPS-001 stays OPEN and the
  SPEC does not enter the run phase. The blockage is recorded with the same
  two-path control used on the measurement day — through moai and through a
  direct `codex exec` call — so that "codex account" and "moai path" are
  distinguishable (verdict.md §E1).
- **The observed live shape is outside the nine measured variants.** The record
  says so. verdict.md §6 already states that the nine variants were enumerated by
  the measurer rather than derived from codex's behaviour; an unenumerated shape
  is expected, not anomalous.
- **A #1718 reduction reproduces only one of the two paths.** AC-CPS-011 stays
  OPEN for that reduction. Its raw body was measured on both paths (t1203 §1),
  and a half-faithful reduction cannot stand in for it.
- **The population is re-measured and the distribution differs.** Record the new
  figures with their tree; do not overwrite the t1203 figures, which remain the
  observation of their day.

## §F Quality gates

- Package tests for `internal/cli` pass, compiled from the tree — not from the
  installed binary (`f67d2193f`, older than this tree; verdict.md §6).
- `go vet ./internal/cli/...` clean.
- The run-phase evidence record names the command it ran and the output it
  observed for every criterion it marks satisfied.

## §G Definition of Done

- [ ] AC-CPS-001 and AC-CPS-002 satisfied and recorded, BEFORE Kickoff approval
- [ ] AC-CPS-003 satisfied — candidates presented, not pre-selected
- [ ] AC-CPS-015 satisfied — the REQ-CPS-010 decision recorded before any
      adversarial-affecting candidate is implemented
- [ ] Selected candidate's criterion (AC-CPS-004 / 005 / 006 / 014) satisfied;
      where #1718 is claimed, AC-CPS-011 and the matching AC-CPS-012 / 013 as
      well; the unselected ones recorded as not-applicable with the selection as
      the reason
- [ ] AC-CPS-007 through AC-CPS-010 satisfied
- [ ] Every claim in the run-phase evidence traces to a command run in this tree
- [ ] Windows path behaviour and live `audit_multi` appear, if at all, as
      explicitly-unmeasured notes — never as claims
