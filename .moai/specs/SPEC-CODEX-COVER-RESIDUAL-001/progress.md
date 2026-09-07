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

### Plan-audit iteration 1 — fix round

Verdict: **FAIL score=0.775** (Tier M threshold 0.80; blocking 4, advisory 5). Report: `.moai/reports/t519/plan-audit-iter1.md`, audited on tree `faa76fa7e`. All seven must-pass criteria passed; the FAIL was driven by the aggregate score, and three of the four blocking findings were the same defect — a Given clause written from a sibling test's *intent* rather than its *text*, dropping a seam the sibling actually carries.

Every finding was re-verified against the source before editing rather than accepted on the report's word, because the auditor's own Gaps section states it executed no tests and its predictions are control-flow reasoning. The verification is recorded below alongside each fix.

| Finding | Artifact:line changed | What changed | Source verification |
|---|---|---|---|
| F1 | acceptance.md AC-CCR-004 Given + rationale; plan.md §F M1 step 5 | The fixture now swaps **all three** codex seams under one `t.Cleanup`, with the `codexLookPath` line copied verbatim; plan.md carries the block as a code fence. Added the reason: without it the test does a real PATH lookup and its verdict depends on the host. | `var codexLookPath = exec.LookPath` confirmed at mcp_codex.go:368; the three-seam precedent confirmed at codex_review_gate_test.go:152-164; `HandleCodexReviewGate` consults it at step 4 (codex_review_gate.go:78), before the session starts. |
| F2 | acceptance.md AC-CCR-003 Given + new rationale paragraph; ledger row M2; plan.md §F M1 step 4 | Added `withChangeDetector(t, true)` and stated why it must not be dropped as redundant: without it the mutated handler ALLOWs at step 3 on the non-git temp dir and M2 is vacuous. | The precedent pairing confirmed at codex_review_gate_test.go:40-46 (comment "even with changes present…"); the non-git-dir-yields-false assertion confirmed at codex_review_gate_test.go:312-314. |
| F3 | acceptance.md §B matrix row AC-CCR-002; ledger rows M1 and new M1b; plan.md §F M1 step 7 + M1 exit line | Split the row: M1 keeps the `fmt.Fprintf` deletion and is bound to AC-CCR-001 only; new **M1b** replaces codex_review_gate.go:188 with `return err` and is AC-CCR-002's adoption basis. | Confirmed by reading codex_review_gate.go:184-189 — `err` is in scope inside the `if err != nil` block, so the edit compiles; empty stdin does produce a non-nil `err` (`json.Unmarshal` on empty input), so M1b flips AC-CCR-002 as well as AC-CCR-001. |
| F4 | plan.md §A.4 table (rewritten, 12 rows); spec.md §E constraint 3 | Every Location cell corrected and the file count stated: the helpers live in **four** files. `stubCodexRunner` moved to `codex_rpc_error_test.go:29`, a file neither artifact had named. | All 12 locations grep-verified in this worktree: `withChangeDetector` codex_review_gate_test.go:28; `writeWorkflowYAML` :33 / `assertAllowJSON` :157 in multi_review_gate_wiring_test.go; `withCodexRunner` :56, `withCodexLookPath` :63, `fakeCodexSession` :74, `fakeCodexConn` :88, `withCodexSession` :108, `codexSessionScript` :123, `errFakeCodexCrash` :408 in mcp_codex_test.go; `stubCodexRunner` codex_rpc_error_test.go:29; `fakeCodexConnPID` codex_jobs_test.go:31. |

Advisories:

| Advisory | Disposition | Where |
|---|---|---|
| A1 — leave the ≥90.0% threshold alone | **Honoured** (it recommends changing nothing) | no edit; the threshold is untouched |
| A2 — no `t.Parallel()` in the six new tests | **Applied** | plan.md §D constraint 10, with the seam-race reason and the sibling-file evidence |
| A3 — M5a's RED arrives as a panic | **Applied** | acceptance.md AC-CCR-010, third bullet: run M5a in isolation and record the panic trace beside the `--- FAIL:` line |
| A4 — spec.md has no scope heading | **Declined.** The fix would edit spec.md prose, outside the acceptance.md / plan.md surface this round authorises, and scope is already carried precisely by §B.2's Disposition column and §F. Re-raise at iteration 2 if the auditor still finds it material. | no edit |
| A5 — `moai spec lint` fails on a directory argument | **Declined as a SPEC change.** It is a linter defect, already pre-recorded in §E.1 and reproduced by the auditor. Belongs in `/moai feedback`, not this SPEC. | no edit |

One defect neither the audit nor the original pass caught, found while applying the above and fixed in the same round: spec.md REQ-CCR-003 pointed at `§D.1` (the coverage-ceiling derivation) where it meant `§D.2` (the vacuous-mutant record). Corrected at spec.md:92.

Post-fix re-verification on this tree: `moai spec lint .moai/specs/SPEC-CODEX-COVER-RESIDUAL-001/spec.md` → `✓ No findings — all SPEC documents are valid`. Counts unchanged at 10 REQ / 12 AC; ledger grew from 9 rows to 10 with M1b. No production `.go` file touched.

## §E.2 Run-phase Evidence

_&lt;pending run-phase&gt;_

## §E.3 Run-phase Audit-Ready Signal

_&lt;pending run-phase&gt;_

## §E.4 Sync-phase Audit-Ready Signal

_&lt;pending sync-phase&gt;_

- 2026-09-07 plan-audit iter-2 PASS 0.9375 (`.moai/reports/t519/plan-audit-iter2.md`); remaining minor F5 (acceptance.md:60 M1→M1b) and F6 (plan.md:37, spec.md:148 four→five files) fixed by the lane directly (one-word edits; distinct-file count re-verified = 5). Advisory A6 (withCodexSession overwrites codexLookPath — never combine it with a t.Fatal LookPath guard in one test) carried into the run-phase delegation as a constraint.

## §F Phase 4 Mode Selection

Decision: serial (Implementation Kickoff Approval obtained 2026-09-07 via lead; autonomous progression)

Input parameters: tier M · scope 2 files (1 new test file + 1 test-file addition) · domains 1 (Go test code, internal/cli) · language mix 100% Go test · concurrency benefit LOW (coding-heavy, sequential mutant discipline) · Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | not a typo-level change; needs mutant RED/GREEN cycles and evidence capture |
| serial | **yes** | coding-heavy Tier M, single domain, 2 files — Anthropic coding-task parallelism caveat |
| fanout | no | 1 domain, 2 files — below the ≥3 domains / ≥10 files auto-select threshold |
| sweep | no | not mechanical bulk transformation |

Decision: serial

Justification: one manager-develop (cycle_type=tdd) drives M1 → M2 → M3 sequentially; each milestone depends on the previous mutant ledger state and the union gate, so concurrent spawns would race on the same package and the same evidence files. Implementation Kickoff Approval: obtained 2026-09-07 via the lead session (operator decision, autonomous progression). Logged at HEAD c6bf21a72 before the first run-phase spawn.
