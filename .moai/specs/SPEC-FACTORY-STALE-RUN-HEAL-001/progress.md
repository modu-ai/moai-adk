# SPEC-FACTORY-STALE-RUN-HEAL-001 — Progress

## Plan Phase (2026-10-02, card t1345)

- Status: draft — spec.md, plan.md, acceptance.md authored by manager-spec (Tier M, 3 artifacts plus this record and the behavioural probe `probe/hook-probe.sh`).
- Research: read-only; mechanism read at tree `802a72235536958ada5b7cd5876a168e4b8c325f` and observed through the probe at `cda6913d127c959cee93b54254e6c7241f8b2032` (Go identical).
- Scope boundary (leader): Factory Mode only — Kanban Mode is being removed by card t1399; no kanban-only surface is touched or depended on.
- SPEC ID self-check: `SPEC-FACTORY-STALE-RUN-HEAL-001` — regex check executed as Bash, output `PASS`; unique in `.moai/specs/`.

## Plan-audit iteration 1 (FAIL 0.65, `.moai/reports/t1345/plan-audit-iter1.md`, audited `cda6913d1`) and iteration 2 revision (spec/plan/acceptance 0.2.0)

Counts after revision: 16 requirements (Tier M ceiling 16), 16 acceptance criteria (ceiling 16); 9 Release-blocking (each with an attached RED-now cell), 7 High regression-guards (reason per row).

Lane decisions applied as given: Codex option (a); Codex per-card loop out of scope; the hook rebind heals current-vocabulary sessions only; SPEC-STALE-RUN-LABEL-001 not amended; additive read of the retired run's broker dropped; single-measurement healthy path; the existing admission-checked claim path evaluated for lane admission.

| Defect | Disposition | Where |
|---|---|---|
| D1 codex launch line refused | fixed (decided: option a) — verb prints and launches only `moai codex -f lane`; `--lane`/`--run` refused with `--provider codex`; A5, DP9, §C item 6, REQ-SRH-012/016, AC-SRH-001/002 corrected; classifier-acceptance assertion added; RED E1 + E2/E3 (`moai codex -f lane-3` exit 1) | spec §C.6, §D.5, §G; acceptance AC-001/002, E2/E3 |
| D2 hook rebind does not repair the measured legacy incident | fixed — §B and §H state that the rebind heals current-vocabulary sessions only and the legacy lanes (worker-69/72) get the printed command for a new session; the causal sentence is rewritten; no legacy rebinding added (REQ-RNC-009) | spec §B, §H |
| D3 SPEC-STALE-RUN-LABEL-001 reconciliation asserted, not shown | fixed (decided: no amendment) — per-requirement table (REQ-SRL-001..009, AC-SRL-005(b): preserved / superseded / amended) and an explicit supersession clause | spec §D.9, DP7 |
| D4 "final" unbind vs later rebound | fixed — notice state machine (unbound is final for legacy only; not final for current-vocabulary), REQ-SRH-003/004 amended, AC-SRH-009 added (unbind-then-active-run) | spec REQ-SRH-003, §D.8; acceptance AC-009, E5 |
| D5 claim path cannot know "rebound"; budget statement inconsistent; no timing AC | fixed — one resolution + one `ProbeRunStateAt` replaces `ValidateActiveRun` (count unchanged); registration returns the rebound run to the same invocation's claim; Stop does not rebind (stated, with its consequence in §H); budgets named per step; injectable seams; timing/fail-open AC | plan §B, §F M2; spec REQ-SRH-008..010; acceptance AC-014 |
| D6 AC-SRH-009 vacuous | fixed — REQ-SRH-005 states silence for any current-vocabulary session on a not-active run; the AC asserts absence of the degraded string; RED observed (E8); old AC-009 is now AC-012 | spec REQ-SRH-005; acceptance AC-012, E8 |
| D7 ten Release-blocking ACs without RED-now | fixed — Release-blocking ACs 001, 002, 004, 005, 006, 008, 009, 010, 012 carry plan-phase RED cells (verb absence and the behavioural probe, E1-E10); AC-003, 007, 011, 013, 014, 015, 016 reclassified High (regression-guard) with the reason per row; none is Release-blocking with a deferred RED | acceptance matrix, D.1 |
| D8 grep proxies as RED; guard ACs marked Release-blocking | fixed — E3/E4 demoted to hints H1-H3, none gates an AC; the behavioural cells replace them; guard ACs reclassified; the "gate called once outside the loop" mutant added to AC-SRH-015's probe row | acceptance D.1 hints, D.2 |
| D9 notice line under-specified; AC-005 shallow | fixed — exact-line table (vocabulary × run state × active-run count), `unknown` backend maps to `cc`, wrong-run-id mutant with exact-line assertion, current-vocabulary rows moved to M2 ACs (008-011) | spec §D.7; acceptance AC-005, D.2 |
| D10 new notice catalogue unspecified; D7 misdescribed current behaviour | fixed — message catalogue N1-N6 (purpose, surface/locales, protocol tokens, cadence carrier); the current-vocabulary account corrected (it receives the per-prompt degraded string, observed E4/E8) | spec §D.8, §C.2 |
| D11 hook registration bypasses lane admission | decided — `ClaimFactoryLaneWithin` infeasible (the slot is held by the session's own live pid → "already occupied"); declared bypass as a tested exclusion with its consequence for the free-slot view (unchanged: pid-keyed) and for declared capacity (not enforced; parity with cc/glm joins) | spec DP11, REQ-SRH-007; acceptance AC-008 |
| D12 Codex loop and loop plumbing | out-of-scope-with-reason (Codex per-card loop, residual risk named) + plumbing listed (signature gains `explicit`, `leadTarget`; the two call sites) | spec §F; plan §B, M3 |
| D13 additive read of the retired run's broker | fixed (decided: dropped) — a rebound session reads only the rebound run's broker; REQ-SRL-007 sentence 1/2 reconciled in the table | spec REQ-SRH-008, DP5, §D.9 |
| D14 agent-facing executable line invites running it | adopted (cheap) — the notice frames the line as the operator's, to run from a terminal after ending the session | spec REQ-SRH-001, §H |
| D15 load-dependent hook ACs | adopted — injectable budget/resolver/listing seams; the hook ACs do not depend on wall-clock; one elapsed bound two orders above measured cost | plan §B; acceptance preamble, AC-014 |
| D16 AC-015 pathspec, `$(...)`, t1399 overlap, two uncovered tests | adopted — pathspec widened to the kanban files and package with the new builder files excluded; command rewritten without `$(...)` against a pinned SHA; t1399 overlap named; the two literal-assertion tests added to AC-SRH-007 | acceptance AC-016, AC-007; plan §B |
| D17 REQ-002/012 bundle many behaviours | partly adopted — REQ-002 split (grammar vs provider/Codex, REQ-SRH-016); REQ-012 kept whole to stay at the 16-requirement ceiling | spec §D.1, §D.5 |
| D18 A4 unmeasured | adopted (state the consequence) — A4 stays an unmeasured assumption; §H states that an autonomous lane without operator prompts stays unrebound | spec §G A4, §H |

- 2026-10-02 — plan-audit iteration 2 (PASS-WITH-DEBT 0.85) debt F1 repaired, nothing else changed: the seven probe-cited RED cells (AC-SRH-004, 005, 006, 008, 009, 010, 012) are re-pinned to `e48d22fc4a14b1f3105ad3129c1c9f910c4afd2a` (the commit that contains `probe/hook-probe.sh`; Go code equals `802a72235`) and written as literal single-invocation commands; `probe/hook-probe.sh` gained an optional binary argument defaulting to `./bin/moai-t1345` (scenarios unchanged). Evidence ledger `acceptance.md` D.1.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02
plan_audit_iteration: 2 (final, ceiling 2)

- Artifacts: spec.md, plan.md, acceptance.md, progress.md, probe/hook-probe.sh under `.moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/`.
- Frontmatter: 12 canonical fields plus `tier: M`, `card`, `depends_on`, `related_specs`; `status: draft`; version 0.2.0.
- Out of Scope: eight H3 topics including the Kanban Mode and Codex per-card loop exclusions.
- Open items: the High ACs' REDs are acquired at M1/M2/M3 RED (reasons per row); A4 (UserPromptSubmit firing on autonomous lanes) is unmeasured; decisions DP1-DP12 await operator or auditor confirmation at the Kickoff.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
