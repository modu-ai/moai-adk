# SPEC-CLIFIX-CRITICAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §1 + §5 P0. Status: draft. Pending plan-audit.

## §E.2 Run-phase Evidence

### M1 — Reproduction batch (RED confirmed 2026-07-10)

8 failing repro tests written before fixes (REQ-CRIT-001-009). Verbatim output
persisted at `.moai/state/verify/327c3427/m1-red-{cli,harness}.log`.

| Defect | Test | M1 result | Evidence |
|--------|------|-----------|----------|
| a | TestSettingsLocalPreserve_Repro | FAIL | hooks/outputStyle/model wiped by struct RMW |
| b | TestClaimTaskAppend_Repro | FAIL | ledger head overwritten (O_RDWR write at offset 0) |
| c | TestHarnessMutePreserve_Repro | FAIL | agentic_loop/team wiped by minimal-struct YAML marshal |
| d | TestRemoveHarnessBoundary_Repro | FAIL | release-update specialist + skill dir deleted by bare prefix match |
| e | TestUpdateLock_ContendedFailsFast_Repro | PASS (primitive) | lock primitive works; defect = zero prod callers (grep-verified pre-fix) |
| f | TestMigrateRollbackPreexisting_Repro | FAIL | pre-existing precious.txt removed by unconditional rollback |
| g | TestMigrateSymlinkSkip_Repro | FAIL | out-of-tree symlink target copied (os.Stat follows link) |
| h | TestTierPromotionHighWater_Repro | FAIL | before=1 after=2 (duplicate promotion appended) |

### M2-M4 — _<pending implementation>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase completion>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
