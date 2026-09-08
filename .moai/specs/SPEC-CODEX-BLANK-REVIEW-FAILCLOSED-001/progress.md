# Progress — SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001

Card: t551 · Branch: `WT-audit-fail-open` · Plan-phase HEAD: `3ac58b5a1`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M).
- SPEC ID regex self-check: executed as Bash, output `PASS`.
- ID uniqueness: no existing SPEC directory matches `BLANK` / `FAILCLOSED` / `FAIL-OPEN`.
- Every file:line citation in the artifacts was re-measured against this tree at
  HEAD `3ac58b5a1` before being written. Two citations supplied by the dispatch
  were corrected by measurement: the `runTurn` guard is at
  `internal/cli/mcp_codex.go:817` (not `:818`), and the pinned synthesizer case
  is at `internal/cli/codex_review_rpc_test.go:119` (not `:120`).
- `moai spec lint --strict .moai/specs/SPEC-CODEX-BLANK-REVIEW-FAILCLOSED-001/spec.md`
  → `✓ No findings — all SPEC documents are valid`, exit 0.
  The invocation covers the whole SPEC directory, not spec.md alone: an earlier
  run of the same command reported findings in `plan.md:5` and `acceptance.md:5`,
  which is the catch-all control showing the linter reaches all three artifacts.
  Two authoring defects were found and repaired by that run:
  (1) `tags:` was authored as a YAML sequence, but `internal/spec/lint.go:511`
      declares `Tags string` — corrected to a quoted CSV;
  (2) `status:` was present in `plan.md` and `acceptance.md`, which
      `ArtifactStatusFieldForbidden` rejects — lifecycle state lives in `spec.md`
      alone. Removed from both.
- Open decision carried into run-phase: `plan.md` §B.1 (option a vs option b for
  the pinned synthesizer test). Recommendation recorded; not settled.

### Plan-audit repair pass (FAIL 0.81 → repaired)

Six blocking findings addressed in one edit pass. Two were FALSE STATEMENTS I
authored; both were independently re-traced against the code in this tree before
rewriting, not accepted on report alone.

| Finding | Repair |
|---|---|
| D2 — AC-CBR-003's discriminating power overstated | Restated: `:1253` ALONE is load-bearing; it says nothing about `:1134`/`:1137`. Covers REQ-CBR-003 only. |
| D3 — REQ-CBR-001/002 had zero coverage | New AC-CBR-009 pins the OVERWRITE case (later blank clobbers earlier real, loop `:1125-1139`). |
| D4 — headline overstated the effect | §A and §C.3 rewritten: `inconclusive` does NOT close the gate hole (`default:` arm → claude anchor). The scoped win is an honest codex row, not a blocked gate. §E reconciled. |
| D5/D7 — old AC-CBR-009 vacuous-capable + mandated the helper | Demoted out of the AC set into `plan.md` §E as a checklist item; helper is now explicitly RECOMMENDED, and no AC mandates implementation shape. |
| D6 — orphan AC naming no REQ | Resolved by the demotion; the AC-CBR-009 slot now covers REQ-CBR-001/002. |
| D8/D9/D10 | `acceptance.md` §D→§A (2 sites); test range `:113-127`→`:114-126`; auditor report citations corrected to `:817`, `:1409`, `:276`/`:1446`, `:291`, `:1608`. |

Also corrected: the §D RED list now includes AC-CBR-009 and states why
AC-CBR-002/004/006/008 are excluded (preservation pins must be green both sides).

REQ↔AC mapping checked directly: REQ-CBR-009 IS referenced, by AC-CBR-007
(`acceptance.md:123`). The relayed `CoverageIncomplete` finding was an artifact
of a wrong lint invocation form and is not a real defect.

Process defect recorded (not mine to undo): the lint-repair rewrites of
2026-09-08 13:44:27-13:44:40 landed inside an audit window opened 13:36.

### Plan-audit iteration 2 (PASS 0.93) — final plan-phase edit

Tier M's 2-iteration audit ceiling is spent; these landed without re-audit.

| Finding | Repair |
|---|---|
| F1 (blocking) — AC-CBR-009 admitted a first-wins mutant | Added one `**And given**` clause pinning the BLANK-FIRST ordering for both item types. Criterion not restructured. |
| F2 — AC-CBR-003's RED reason unstated | §D now names what must be asserted beyond the verdict: the value returned by `bestCodexReviewText`, since the verdict half is already `pass` pre-repair and would be a vacuous green. |
| F5 | `codexFindingsOf` citation `:1465`→`:1473` (re-measured; `:1465` is the first line of its doc comment, not the func). |
| F6 | `acceptance.md:112`→`:123` (re-measured, not offset). |

F3 and F4 were classed optional and are deliberately NOT addressed — the AC set
is closed.

**Residual risk (F4, unpinned).** No criterion pins the PREFERENCE ordering in
`bestCodexReviewText` (`internal/cli/mcp_codex.go:1252-1257`) — that a non-blank
structured `exitedReviewMode.review` wins over a non-blank `agentMessage.text`.
An implementation that reversed that preference would satisfy every criterion in
this SPEC, because each one exercises a case where exactly one of the two is
non-blank. The preference is pre-existing behavior this SPEC does not modify, so
the exposure is a silent regression during run-phase editing of `:1253`, not a
defect introduced here. Run-phase should keep the preference intact when making
that site blank-aware.
- Unobserved items are recorded in `spec.md` §E Gaps, not asserted in the body.

**Plan-audit closed.** Iteration 1 FAIL 0.81 → iteration 2 PASS 0.93 (Tier M
threshold 0.80, delta +0.12, so the score-regression STOP clause did not fire).
Both verdicts are recorded at `.moai/reports/t551/plan-audit-verdict.md`
(iteration 2 appended at line 292, iteration 1 preserved).

## §F Phase 4 Mode Selection

**Decision: serial**

Input parameters — tier: M · scope: 1 implementation file
(`internal/cli/mcp_codex.go`) plus test files · domain count: 1 (Go source,
`internal/cli`) · file language mix: 100% Go · concurrency benefit: LOW.

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial — four coupled sites, a RED-first contract, and a nine-criterion acceptance suite. |
| `serial` | **yes** | Coding-heavy single-package work; one sequential `manager-develop` spawn per milestone. |
| `fanout` | no | Single domain, single file. Fan-out would add reconciliation cost with nothing to parallelize, and Anthropic's coding-task parallelism caveat points at `serial` for coding work regardless. |
| `sweep` | no | Far below the ~30-file mechanical threshold, and the transform is not uniform — four sites with different surrounding logic. |

Justification: every criterion for the alternatives fails on the same fact —
this is one file in one package, and the four edit sites are coupled (the guard's
correctness depends on the selection site, which depends on the collection
sites). Splitting coupled edits across concurrent agents would create exactly the
write race the concurrency safeguard forbids, for no wall-clock gain.

Implementation Kickoff Approval: PASSED (operator, this session). Progression
mode: autonomous — report once at completion with the evidence gathered.

## §E.2 Run-phase Evidence

Run-phase HEAD at entry: `3ac58b5a1` · branch `WT-audit-fail-open` · card t551.

### M1 — decisions settled and the characterization baseline

**plan.md §B.1 decision: OPTION (a).** The repair lands at the guard
(`internal/cli/mcp_codex.go:817`), the two collection sites (`:1134`, `:1137`),
and the selection site (`:1253`). `synthesizeReviewOutput` and
`codexUnrecognizedVerdict` are NOT touched, so the pinned case at
`internal/cli/codex_review_rpc_test.go:119` (`"": "pass"`) stays byte-identical
and untouched.

Rationale, stated as the plan states it: the defect is REACHABILITY, not
synthesis. The synthesizer's answer for an unrecognized body is a documented
mode-keyed policy (REQ-CBR-007) that this card preserves; the bug is that an
ABSENT body is routed to it at all. Repairing the routing removes the blank
body from the synthesizer's input set without moving the synthesizer's
contract. Option (b) was rejected because it edits a deliberately-placed record
and widens the blast radius to every caller of a function this card has not
surveyed, for a case the guard already makes unreachable.

**plan.md §B.2 decision: the helper IS adopted.** `codexReviewTextIsBlank`
(`internal/cli/mcp_codex.go`) is a one-line wrapper over
`strings.TrimSpace(s) == ""`, called at all four sites. The plan marks it
RECOMMENDED-not-mandated and notes it earns its keep only through the
miss-one-site argument — that argument is exactly this defect's shape (three
sites written with the same wrong test, a fourth added later), so the wrapper
is adopted. No acceptance criterion mandates it; it remains an implementation
choice.

**U+00A0 decision: IN SCOPE, and asserted.** `strings.TrimSpace` cuts on
`unicode.IsSpace`, which includes U+00A0, so a body of non-breaking spaces
alone is blank. `acceptance.md` §C requires the behavior be asserted either
way; `TestCodexBlankReview_BlankDiscriminator` pins it rather than leaving it
implicit.

**Characterization baseline (M1, measured BEFORE any repair).**
`internal/cli/codex_blank_review_characterization_test.go` pins the pre-repair
verdict + Summary of the review-text path per fixture class, and pins the
unavailable-backend path as the permanent AC-CBR-004 control. Command and
verbatim output:

```
$ go test -run 'TestCharacterize_UnavailableBackend|TestCharacterize_ReviewTextPathPreRepair' -v ./internal/cli/
=== RUN   TestCharacterize_UnavailableBackend
--- PASS: TestCharacterize_UnavailableBackend (0.00s)
=== RUN   TestCharacterize_ReviewTextPathPreRepair
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/space-only (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/newline-only (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/mixed-whitespace (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/exactly-empty (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/real-clean-review (0.00s)
    --- PASS: TestCharacterize_ReviewTextPathPreRepair/finding-bullet (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.932s
```

Baseline-attribution: this run, this tree, HEAD `3ac58b5a1`. The three
whitespace rows pin `verdict="pass" summary=""` — the defect, observed here as
a green characterization rather than asserted from the plan-phase probe.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
