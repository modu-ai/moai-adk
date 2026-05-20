# SPEC-V3R5-HARNESS-AUTONOMY-001 Progress Log

## Plan Phase (initial draft)

- plan_drafted_at: 2026-05-20T03:36:00Z
- plan_status: draft (awaiting plan-auditor iter1)
- plan_auditor_iterations: 0 (target: 2 with PASS at iter2 ≥0.85)
- plan_auditor_target_final_verdict: PASS
- plan_auditor_target_final_score: 0.902 (per issue #1022 body — W3 expected ≥0.902 vs W1 actual 0.96)
- plan_auditor_threshold: 0.85
- baseline_main_HEAD: 7bd23bb69 (post W2 sync, "sync(SPEC-V3R5-CORE-SLIM-001): W2 lifecycle 완료")
- scope_tier: T2 Standard (per orchestrator AskUserQuestion in issue #1022)
- mega_sprint_wave: W3 (W0→W1→max(W2,W3)→W4→v3.5.0; W3 parallel-eligible with W2)
- empirical_ground_truth:
    - brownfield_files_internal_harness: 30+ (mixed PRESERVE/EXTEND per §1.5)
    - brownfield_files_internal_harness_safety: 12 (6 .go + 6 _test.go, all EXTEND)
    - brownfield_internal_hook_pre_tool_loc: 653 (EXTEND with 8 sentinel catalog)
    - new_packages: 4 (capture/, tier/, throttle/, seeds/)
    - new_agents: 1 (.claude/agents/moai/harness-learner.md)
    - estimated_touched_files: ~25
    - estimated_LOC: ~2850 (code + tests)

## Acceptance Targets (draft)

- AC-HRA-001 through AC-HRA-014 (14 ACs, 100% binary verification)
- REQ-HRA-001 through REQ-HRA-038 (38 REQs across 6 EARS types + Integrity/Hybrid)
- EC-HRA-001 through EC-HRA-006 (6 edge cases)
- R-HRA-01 through R-HRA-05 (5 risk mitigations)
- C-HRA-008 (1 constraint AC — subagent boundary grep gate)
- Sentinel catalog: 8 `HARNESS_FROZEN_*` + 6 `HARNESS_LEARNING_*` = **14 total**
- EXCL-HRA-001 through EXCL-HRA-010 (10 exclusions with W4 downstream refs)

## Dependencies (verified at draft time)

- **W1 (Hard, COMPLETE)**: SPEC-V3R5-CONSTITUTION-DUAL-001 — zone-registry 111 entries available at `.claude/rules/moai/core/zone-registry.md`. L1 Frozen Guard reads only per W1 EXCL-001.
- **W2 (Parallel, COMPLETE)**: SPEC-V3R5-CORE-SLIM-001 — no race, no touch points.
- **W4 (Downstream, BLOCKED on W3)**: SPEC-V3R5-PROJECT-MEGA-001 — needs W3 substrate.

## Files Drafted

- [x] spec.md (~720 LOC) — 12-field frontmatter + §1.5 Brownfield Inventory + §1.6 D11 Seed Decision + §1.7 Field Naming Policy + 38 EARS REQ
- [x] plan.md (~520 LOC) — 5-Layer architecture diagram + INVARIANT ordering rationale + M1-M6 milestones + 14-sentinel catalog + 5 risk mitigations + 3 OQ
- [x] acceptance.md (~370 LOC) — 14 binary AC-HRA + 6 EC + 5 Risk Mitigations + 1 Constraint AC (C-HRA-008) + traceability matrix
- [x] spec-compact.md (~110 LOC) — token-efficient summary for context window economy
- [x] progress.md (this file)

## Next Steps

1. **Plan-Auditor iter1**: Submit `plan.md` to plan-auditor agent → expect REVISE verdict ≥0.75 with BLOCKING/SHOULD findings.
2. **Iteration 2 revision**: Address all BLOCKING + P1 SHOULD per W1 recovery pattern (iter1 0.71 → iter2 0.96, delta +0.25).
3. **Plan-Auditor iter2**: Re-submit for PASS verdict ≥0.85 (target: 0.902).
4. **DP2 (Git environment selection)**: Single SPEC, single run-phase PR per W1/W2 precedent.
5. **Phase 3.6 SPEC quality gate**: Lint clean (validate against StatusGitConsistency).
6. **Run Phase (M1-M6)**: Sequential implementation; M3 is largest milestone (brownfield EXTEND `internal/harness/safety/*.go` + `internal/hook/pre_tool.go`).
7. **Sync Phase**: PR + ACs binary-PASS verification + status `draft → implemented → completed`.

## Open Questions for Plan-Auditor

1. **OQ1** — Sentinel count discrepancy: spec.md §7 says 10, plan.md §3 says 14 (expansion after M3/M4 review). Resolution: plan.md §3 is authoritative; spec.md §7 to be updated in iter2.
2. **OQ2** — L2 Canary baseline source for single-project repos: proposal = synthetic snapshots from `.moai/specs/SPEC-*/` last 3 SPECs' acceptance.md PASS state.
3. **OQ3** — `.moai/harness/disabled` sentinel — `.gitignore` default but commit-allowed per per-repo isolation (EXCL-HRA-008).

## Risks Surface

- **R-HRA-04 (Brownfield test regression)** — highest probability. Mitigation: pre-EXTEND baseline `go test ./internal/harness/safety/... -v` capture; no test name removal allowed.
- **R-HRA-03 (Subagent boundary leakage)** — highest impact. Mitigation: C-HRA-008 grep gate in CI.

---

Status: **DRAFT** awaiting plan-auditor review. Issue: #1022. Branch: `claude/issue-1022-20260520-0336`.
