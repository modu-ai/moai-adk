# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-002 — Progress

**Status**: draft (plan-phase → run-phase)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-16
- plan-auditor verdict: PASS
- plan-auditor score: 0.95 (Tier M threshold 0.80)
- skip-eligible: YES (score ≥ 0.90; governs ONLY Phase 1 re-execution — Implementation Kickoff Approval remains mandatory, obtained 2026-07-16)
- probes: P1 GEARS PASS / P2 root-cause PASS / P3 parent-contract non-weakening PASS / P4 Reproduction-First PASS
- blocking defects: none
- SHOULD-FIX: S1 (acceptance.md HARD-5 number collision, cosmetic) — deferred

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**:
- tier: M
- scope: ~5-10 files (internal/cli/update.go, v2_detection.go, update_clean_install.go, update_preserve_inventory.go, update_cleanup.go + tests)
- domain count: 1 (internal/cli clean-reinstall code path)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, sequential dependency between milestones)

**Mode evaluation**:
- Mode 1 (trivial): not selected — semantic regression repair, not a typo
- Mode 2 (background): not selected — write-capable implementation work
- Mode 3 (agent-team): RETIRED — never selected
- Mode 4 (parallel): not selected — single domain, coding-heavy (Anthropic coding-task parallelism caveat)
- Mode 5 (sub-agent): **selected** — single sequential manager-develop per milestone
- Mode 6 (workflow): not selected — scope < ~30 files, not mechanical-uniform

**Decision**: sub-agent (Mode 5)

**Justification**: Coding-heavy single-domain regression repair. Per Anthropic's coding-task parallelism caveat ("most coding tasks involve fewer truly parallelizable tasks than research"), the sequential sub-agent path is the safe default. Milestones M1→M5 have sequential dependencies (M1 fingerprint fix is the highest-irreversibility data-model change that M2-M5 build on). Tier M Section A-E delegation template applies.

**Implementation Kickoff Approval**: obtained 2026-07-16 (user selected "run-phase 진입 (권장)"). cycle_type=tdd (existing v2_detection_test.go / update_clean_install_test.go baseline).
