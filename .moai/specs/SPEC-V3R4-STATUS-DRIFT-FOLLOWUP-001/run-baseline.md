# SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — Wave 1 BASELINE

측정 일시: 2026-05-16
브랜치: feat/SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001
base commit: 602a07c84

## 측정 명령

```bash
moai spec lint --strict 2>&1 | grep "StatusGitConsistency" > /tmp/sdf-baseline.txt
wc -l /tmp/sdf-baseline.txt  # → 65 (64 WARNINGs + 1 blank line)
```

## 총 drift 건수

**64건** (± 계획 64건, 정확히 일치)

## 패턴별 분포

| 패턴 | 설명 | 건수 | affected-list |
|------|------|------|---------------|
| A | completed → implemented | 47 | affected-list-pattern-A.txt |
| B | completed → in-progress | 4 | affected-list-pattern-B.txt |
| C | implemented → in-progress | 6 | affected-list-pattern-C.txt |
| D | superseded → completed | 1 | (inline: SPEC-LSP-001) |
| E | superseded → implemented | 1 | (inline: SPEC-V3R3-HARNESS-001) |
| F | archived → implemented | 1 | (inline: SPEC-I18N-001-ARCHIVED) |
| G | archived → in-progress | 1 | (inline: SPEC-V3R3-WEB-001) |
| H | self-drift cleanup chain | 3 | affected-list-pattern-H.txt |
| (self) | 이 SPEC 자체 draft → planned | 1 | sync-phase 위임 (OQ4) |
| **합계** | | **64** | |

> 참고: 계획서 패턴 A는 50건으로 예상했으나 실제 47건. 합계는 정확히 64건.

## 패턴별 SPEC 목록

### Pattern A (47건) — completed → implemented

SPEC-AGENCY-ABSORB-001, SPEC-AGENT-002, SPEC-CC2122-HOOK-001, SPEC-CC2122-STATUSLINE-001,
SPEC-CC297-001, SPEC-CICD-001, SPEC-DB-SYNC-HARDEN-001, SPEC-DB-SYNC-RELOC-001,
SPEC-DESIGN-001, SPEC-DOCS-SB-REMOVE-001, SPEC-GLM-001, SPEC-HOOK-008, SPEC-HOOK-009,
SPEC-KARPATHY-001, SPEC-PSR-001, SPEC-QUALITY-001, SPEC-REFACTOR-001, SPEC-SKILL-002,
SPEC-SKILL-GATE-001, SPEC-SLE-001, SPEC-SLV3-001, SPEC-SRS-001, SPEC-SRS-002, SPEC-SRS-003,
SPEC-STATUS-AUTO-001, SPEC-STATUSLINE-002, SPEC-TEAM-001, SPEC-UI-003, SPEC-UPDATE-002,
SPEC-UTIL-003, SPEC-V3R2-ORC-001, SPEC-V3R2-ORC-005, SPEC-V3R2-RT-004, SPEC-V3R2-RT-005,
SPEC-V3R2-SPC-004, SPEC-V3R2-WF-005, SPEC-V3R2-WF-006, SPEC-V3R3-ARCH-007,
SPEC-V3R3-BRAIN-001, SPEC-V3R3-CMD-CLEANUP-001, SPEC-V3R3-COV-001, SPEC-V3R3-DEF-001,
SPEC-V3R3-DEF-007, SPEC-V3R3-DESIGN-PIPELINE-001, SPEC-V3R4-STATUS-LIFECYCLE-001,
SPEC-WF-AUDIT-GATE-001, SPEC-WORKTREE-002

### Pattern B (4건) — completed → in-progress

SPEC-CORE-001, SPEC-LOOP-001, SPEC-V3R4-CATALOG-001, SPEC-V3R4-HARNESS-003

### Pattern C (6건) — implemented → in-progress

SPEC-UTIL-001, SPEC-V3R2-CON-001, SPEC-V3R2-CON-002, SPEC-V3R2-CON-003,
SPEC-V3R2-RT-001, SPEC-V3R2-SPC-003

### Pattern D (1건) — superseded → completed (terminal-state exemption 대상)

SPEC-LSP-001

### Pattern E (1건) — superseded → implemented (terminal-state exemption 대상)

SPEC-V3R3-HARNESS-001

### Pattern F (1건) — archived → implemented (terminal-state exemption 대상)

SPEC-I18N-001-ARCHIVED

### Pattern G (1건) — archived → in-progress (terminal-state exemption 대상)

SPEC-V3R3-WEB-001

### Pattern H (3건) — self-drift cleanup chain (sync-phase 위임)

SPEC-V3R4-LINT-SKIP-CLEANUP-001, SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001, SPEC-V3R4-SPECLINT-DEBT-001

### 이 SPEC 자체 (1건) — sync-phase 위임

SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 (draft → planned)

## Wave 실행 계획

- Wave 1 (현재): BASELINE 측정 + 카테고리화 ✅
- Wave 2: Pattern A 47건 bulk downgrade (completed → implemented)
- Wave 3: Pattern B(4건) + C(6건) 개별 검증 후 일부 downgrade
- Wave 4: Pattern D/E/F/G 4건 — terminalStatusEnum detector 예외 처리 (TDD)
- Wave 5: Pattern H 3건 + 이 SPEC 자체 — sync-phase 위임 문서화
- Wave 6: Closeout (full suite 검증 + progress.md 업데이트)

## 목표 상태

Wave 4 이후 `moai spec lint --strict 2>&1 | grep -c StatusGitConsistency` = **0**
