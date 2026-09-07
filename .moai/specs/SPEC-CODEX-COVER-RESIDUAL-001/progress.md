# SPEC-CODEX-COVER-RESIDUAL-001 — Progress

Card: t519 · Branch: `WT-codex-cover-residual` · Plan-phase base: `bf779ecf2`

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-09-07T08:54:01Z
plan_status: audit-ready
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 10   # REQ-CCR-001..010
acceptance_criteria: 12   # AC-CCR-001..012
needs_clarification: 0
```

### Tier deviation from dispatch

The dispatch named Tier S; this SPEC is classified **M**. Grounds are mechanical and recorded in spec.md §A.1: the requested artifact set is the 3-file Tier M set, and the counts (10 REQ / 12 AC) exceed both Tier S ceilings of 8. The card text allowed "Tier S~M". Consequence: plan-audit PASS threshold is 0.80 and the Section A-E delegation template is required.

### Phase 1 SKIP rationale (no Socratic interview)

The Context-First Discovery interview was skipped. Intent clarity was already 100% at dispatch: the card named both target surfaces with file and line, supplied a measured per-function baseline with its command and verbatim output, enumerated the six test items, named the reusable seams, and pre-identified one vacuous mutant. None of the four ambiguity triggers fired — no unresolved referent, no multi-interpretable verb, no unclear boundary (the out-of-scope set is explicit), and no conflict with existing state. No `[NEEDS CLARIFICATION]` markers were emitted.

One judgment call was resolved without asking, because the card itself supplied the latitude: the Tier S-vs-M inconsistency above, resolved by the governing ceiling rule and surfaced rather than silently absorbed.

### Plan-phase verification performed

Read-only checks run by manager-spec on tree `bf779ecf2`:

| Check | Command | Observation |
|---|---|---|
| SPEC ID format | `[[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS \|\| echo FAIL` | `PASS` |
| ID uniqueness | `ls -d .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001` | `No such file or directory`, exit 1 — no collision |
| Axis 1 baseline | `grep runCodexReviewGate .moai/reports/t519/coverage-baseline-targets.txt` | `…codex_review_gate.go:183:		runCodexReviewGate			0.0%`, exit 0 |
| Axis 2 baseline | `grep mcp_codex.go:701 .moai/reports/t519/coverage-baseline-targets.txt` | `…mcp_codex.go:701:			pid					66.7%`, exit 0 |
| Codex RunE tests absent | `grep -rn 'func TestRunCodexReviewGate' internal/cli/` | no output, exit 1 |
| Control for the above | `grep -rn 'func TestRunMultiReviewGate' internal/cli/` | 3 rows from `multi_review_gate_wiring_test.go`, exit 0 |

The last two rows are a pair. The empty result alone would be consistent with a wrong grep form or a wrong path; the control run in the same form on the sibling gate prints three rows, so the absence is a real absence rather than a broken probe.

Source facts were verified by direct read, not inferred: `runCodexReviewGate` (codex_review_gate.go:183-202), `(codexSessionHandle).pid` (mcp_codex.go:701-706), `codexConnPID` (mcp_codex.go:411-416), `(*fakeCodexConn).pid` (codex_jobs_test.go:35), `newGateCmd` binding to `runMultiReviewGate` (multi_review_gate_wiring_test.go:104), and the one-for-one symmetry claim in that file's header (line 13).

**Not verified by manager-spec:** the coverage baseline itself was measured by the lane, not re-run here; it is cited with that provenance in spec.md §B.1. The 12-of-13 statement ceiling in spec.md §D.1 is a derivation from reading the source, not a measurement, and is labelled as such.

### spec lint

```
$ moai spec lint .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001
```

Result recorded at run time by the lane — see §E.2. Known candidate defect from card t500: `moai spec lint` may emit a `ParseFailure` on an ID-form argument; if it does, the exact command and output are recorded here and the run continues (the defect is the linter's, not the SPEC's).

### Card Cross-Check

| Milestone | Card |
|---|---|
| M1 — runCodexReviewGate wiring tests | t519 |
| M2 — codexSessionHandle.pid nil-guard arm | t519 |
| M3 — close: re-measurement + evidence | t519 |

3 milestones → 1 card (t519). No milestone requires a new card.

## §E.2 Run-phase Evidence

_&lt;pending run-phase&gt;_

## §E.3 Run-phase Audit-Ready Signal

_&lt;pending run-phase&gt;_

## §E.4 Sync-phase Audit-Ready Signal

_&lt;pending sync-phase&gt;_
