# V3R2 + V3R3 Master Wave Plan

> Created: 2026-04-28
> Status: IN PROGRESS — Wave 0 complete, Wave 1 starting
> Total SPECs: 38 draft (35 V3R2 + 3 V3R3)
> Waves: 11
> Estimated scope: Rules-only (14 SPECs) + Go code (24 SPECs)

## Session Log

| Session | Date | Completed |
|---------|------|-----------|
| 1 | 2026-04-28 | Wave 0: PATTERNS-001 (already done), DESIGN-FOLDER-FIX-001 (implemented), HARNESS-LEARNING-001 (already done) |
| 2 | 2026-04-28 | Wave 1: CON-003 (rules-only), CON-002 (Go: 5-layer safety gate), SPC-001 (Go: EARS hierarchical acceptance). SPC-002 deferred to Wave 2 (RT-001 dep). |

---

## Overview

11개 웨이브로 38개 draft SPEC을 의존성 순서대로 실행. 각 웨이브는 독립 PR로 머지 가능하도록 구성.

### Wave Sizing Rules

- Go code SPEC: 웨이브당 3-5개 (context pressure 고려)
- Rules-only SPEC: 웨이브당 4-8개 (빠른 실행)
- Mixed: Go code 3개 + rules-only 2개 수준

### Type Classification

- **Rules-only**: `.claude/rules/`, `.claude/skills/`, `.claude/agents/` 내 마크다운/스킬 파일만 수정
- **Go code**: `internal/`, `pkg/` 하위 Go 소스 수정 필요
- **Mixed**: 둘 다

---

## Dependency DAG (Critical Path)

```
CON-001 (done) ─┬─→ CON-002 ──→ SPC-001 ──→ SPC-003
                 ├─→ CON-003 ──→ SPC-002 ──→ SPC-004
                 ├─→ RT-005 ──┬─→ RT-001 ──→ RT-002 ──→ RT-003
                 │            └─→ RT-004       RT-006 ──→ RT-007
                 ├─→ ORC-001 ──→ ORC-002 ──→ ORC-003
                 │            └─→ ORC-004 ──→ ORC-005 (→ RT-004)
                 ├─→ HRN-001 ──→ HRN-002 ──→ HRN-003
                 └─→ EXT-004 ──→ MIG-002 ──→ MIG-001
                               → MIG-003 ──↗

WF-001 (done) ──→ WF-002, WF-003, WF-004, WF-005
WF-006 (done) ──→ EXT-002

V3R3 (independent):
HARNESS-001 (done) ──→ HARNESS-LEARNING-001
DEF-007 (done) ──→ PATTERNS-001
(none) ──→ DESIGN-FOLDER-FIX-001
```

---

## Wave 0: V3R3 Completion

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| PATTERNS-001 | Rules-only | P1 | DEF-007 (done) | 6개 reference doc 흡수 (agent-patterns, boundary-verification, skill-ab-testing, team-pattern-cookbook, orchestrator-templates, skill-writing-craft) |
| DESIGN-FOLDER-FIX-001 | Go code | P1 | None | moai update 시 reserved filename collision → warning 격하 |
| HARNESS-LEARNING-001 | Go code | P0 | HARNESS-001 (done), PROJECT-HARNESS-001 (done) | Self-learning dynamic harness (4-tier confidence pipeline) |

**Rationale**: V3R3은 active release이므로 최우선 완료. PATTERNS-001은 이미 템플릿에 NOTICE.md 등 일부 반영됨.

---

## Wave 1: Constitution Foundation

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| CON-002 | Rules-only | P0 | CON-001 (done) | Amendment protocol (5-layer safety gate) |
| CON-003 | Rules-only | P0 | CON-001 (done) | Rule tree consolidation (merge duplicates, relocate, frontmatter migration) |
| SPC-001 | Rules-only | P0 | CON-001 (done), CON-002 | EARS + hierarchical acceptance criteria |
| SPC-002 | Mixed | P0 | CON-001 (done), RT-001 | @MX TAG v2 (hook JSON integration, sidecar index) |

**Rationale**: 모든 후속 웨이브의 기반. CON-002/003과 SPC-001은 rules-only로 빠른 실행. SPC-002는 RT-001 의존이 있으나 rules 부분만 선구현 가능.

**Note**: SPC-002의 RT-001 의존은 Go 코드 부분에만 해당. Rules 파트는 Wave 1에서, Go 파트는 Wave 2에서 분할 구현 가능.

---

## Wave 2: Runtime Core

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| RT-005 | Go code | P0 | CON-001 (done) | Multi-layer settings resolution with provenance tags |
| RT-001 | Go code | P0 | CON-001 (done), RT-005 | Hook JSON-OR-ExitCode dual protocol |
| RT-004 | Go code | P0 | CON-001 (done), RT-005 | Typed session state + phase checkpoint |

**Rationale**: RT 레이어의 기반. RT-005(Settings Resolution)가 최우선 — RT-001과 RT-004가 의존.

---

## Wave 3: Permission & Runtime Extended

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| RT-002 | Go code | P0 | RT-001, RT-005 | Permission stack + bubble mode |
| RT-003 | Go code | P0 | RT-002 | Sandbox execution layer (Bubblewrap/Seatbelt/Docker) |
| RT-006 | Go code | P1 | RT-001, RT-002, RT-004 | Hook handler 27-event coverage |

**Rationale**: Wave 2 산출물 기반. RT-006은 RT-001/002/004 모두에 의존하므로 Wave 3에서 처리.

---

## Wave 4: Agent Management

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| ORC-001 | Go code | P1 | CON-001 (done) | Agent roster consolidation (22 → 17) |
| ORC-002 | Go code | P1 | ORC-001 | Agent common protocol CI lint (moai agent lint) |
| ORC-003 | Rules-only | P1 | ORC-001, ORC-002 | Effort-level calibration matrix |
| ORC-004 | Rules-only | P1 | ORC-001 | Worktree MUST rule for write-heavy profiles |
| ORC-005 | Go code | P1 | ORC-004, RT-004 | Dynamic team generation + mailbox protocol v2 |

**Rationale**: ORC-001이 기반. ORC-003/004는 rules-only. 5개 SPEC이므로 context 제한에 주의. 필요시 ORC-001/002 (Wave 4A) + ORC-003/004/005 (Wave 4B)로 분할.

---

## Wave 5: SPEC Tooling

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| SPC-003 | Go code | P1 | SPC-001 | SPEC linter (moai spec lint) |
| SPC-004 | Go code | P1 | SPC-002 | @MX anchor resolver (query by SPEC ID, fan_in, danger) |

**Rationale**: SPC-001/002(Wave 1)에 의존하는 작은 웨이브. 빠른 실행 가능.

---

## Wave 6: Workflow Optimization

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| WF-002 | Go code | P1 | WF-001 (done) | Commands thin-wrapper enforcement + fat-command extraction |
| WF-003 | Go code | P1 | WF-001 (done), WF-004 | Multi-mode router (--mode flag) |
| WF-004 | Go code | P1 | WF-001 (done) | Agentless fixed-pipeline classification |
| WF-005 | Rules-only | P1 | WF-001 (done) | Language rules vs skills boundary codification |

**Rationale**: WF-001/006 이미 완료. WF-005는 rules-only로 나머지 3개 Go code SPEC과 병행 가능.

---

## Wave 7: Runtime Hardening Completion

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| RT-007 | Go code | P1 | RT-001, RT-006 | Hardcoded path fix + versioned migration |

**Rationale**: 단일 SPEC. Wave 3 완료 후 실행. 작은 규모이므로 Wave 6과 병합 검토 가능.

---

## Wave 8: Harness System

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| HRN-001 | Go code | P0 | CON-001 (done), ORC-003 | Harness routing + harness.yaml Go loader |
| HRN-002 | Rules-only | P0 | HRN-001 | Evaluator memory scope amendment |
| HRN-003 | Go code | P0 | HRN-001, HRN-002, SPC-001 | Hierarchical acceptance scoring (4-dimension × sub-criteria) |

**Rationale**: HRN-001이 harness 시스템의 기반. V3R3 HARNESS-001(meta-harness) 완료 상태에서 V3R2 harness routing 구현.

---

## Wave 9: Extension Layer

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| EXT-001 | Go code | P1 | pattern-library refs | Typed memory taxonomy (4-type enforcement) |
| EXT-002 | Go code | P1 | WF-006 (done), EXT-001 | Output-styles and memdir Go loader |
| EXT-003 | Go code | P1 | WF-001 (done) | Plugin system (design-only scope for v3.0.0) |
| EXT-004 | Go code | P1 | CON-001 (done) | Versioned migration framework |

**Rationale**: EXT-004는 MIG 웨이브의 전제 조건. 4개 Go code SPEC이므로 context 관리 주의.

---

## Wave 10: Migration Tools (Final)

| SPEC | Type | Priority | Depends On | Scope |
|------|------|----------|------------|-------|
| MIG-002 | Go code | P1 | EXT-004 | Hook registration cleanup |
| MIG-003 | Go code | P1 | EXT-004 | Config loader completeness (5 unloaded YAML sections) |
| MIG-001 | Go code | P0 | EXT-004, MIG-002, MIG-003 | v2 to v3 migrator (moai migrate v2-to-v3) |

**Rationale**: MIG-001은 MIG-002/003 + EXT-004에 의존하므로 반드시 마지막. 최종 마이그레이션 도구.

---

## Execution Strategy

### Per-Wave Execution Flow

```
1. /moai run WAVE-N  (또는 개별 SPEC-ID 지정)
2. 각 SPEC을 manager-ddd/tdd로 구현
3. go test ./... 통과 확인
4. PR 생성 (squash merge for feature, merge commit for release)
5. 다음 웨이브 진행
```

### Session Management

- 각 웨이브는 1-2 세션에서 완료 목표
- Context 75% 도달 시 /clear + resume message로 이어서 진행
- 진행 상황을 본 파일의 진행 상태 섹션에 업데이트

### Merge Strategy per Wave

| Wave | Branch Prefix | Merge Strategy |
|------|---------------|----------------|
| 0 | `feat/V3R3-*` | squash |
| 1-9 | `feat/SPEC-V3R2-*` | squash |
| 10 | `feat/SPEC-V3R2-MIG-*` | squash |
| Final release | `release/v3.0.0` | merge commit |

---

## Progress Tracking

### V3R3 Progress

| SPEC | Wave | Status |
|------|------|--------|
| PATTERNS-001 | 0 | 🟢 |
| DESIGN-FOLDER-FIX-001 | 0 | 🟢 |
| HARNESS-LEARNING-001 | 0 | 🟢 |

### V3R2 Progress

| SPEC | Wave | Status |
|------|------|--------|
| CON-002 | 1 | 🟢 |
| CON-003 | 1 | 🟢 |
| SPC-001 | 1 | 🟢 |
| SPC-002 | 1→2 | 🟡 |
| RT-005 | 2 | ⏸️ |
| RT-001 | 2 | ⏸️ |
| RT-004 | 2 | ⏸️ |
| RT-002 | 3 | ⏸️ |
| RT-003 | 3 | ⏸️ |
| RT-006 | 3 | ⏸️ |
| ORC-001 | 4 | ⏸️ |
| ORC-002 | 4 | ⏸️ |
| ORC-003 | 4 | ⏸️ |
| ORC-004 | 4 | ⏸️ |
| ORC-005 | 4 | ⏸️ |
| SPC-003 | 5 | ⏸️ |
| SPC-004 | 5 | ⏸️ |
| WF-002 | 6 | ⏸️ |
| WF-003 | 6 | ⏸️ |
| WF-004 | 6 | ⏸️ |
| WF-005 | 6 | ⏸️ |
| RT-007 | 7 | ⏸️ |
| HRN-001 | 8 | ⏸️ |
| HRN-002 | 8 | ⏸️ |
| HRN-003 | 8 | ⏸️ |
| EXT-001 | 9 | ⏸️ |
| EXT-002 | 9 | ⏸️ |
| EXT-003 | 9 | ⏸️ |
| EXT-004 | 9 | ⏸️ |
| MIG-002 | 10 | ⏸️ |
| MIG-003 | 10 | ⏸️ |
| MIG-001 | 10 | ⏸️ |

---

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Context overflow per wave | 세션 중단 | Rules-only와 Go code 혼합, 5개 이하 유지 |
| Cross-SPEC dependency conflict | 재작업 | DAG 엄격 준수, 이전 웨이브 PR 머지 확인 후 진행 |
| Breaking change 축적 | 마이그레이션 부담 | Wave 10에서 MIG-001로 일괄 처리 |
| ORC-001 (22→17 consolidation) | 기존 워크플로우 영향 | 보수적 접근, deprecated 단계 거쳐 제거 |
| HARNESS-LEARNING-001 | 복잡도 높음 | V3R3 HARNESS-001 기반 위에 구축, incremental approach |

---

## Appendix: Already Completed

### V3R2 Completed (3)
- CON-001: FROZEN/EVOLVABLE zone codification
- WF-001: Skill Consolidation Stage 1 (48→38)
- WF-006: Output Styles Alignment

### V3R3 Completed (8)
- ARCH-007: Token Circuit Breaker
- CMD-CLEANUP-001: Commands Cleanup
- COV-001: Mobile Native Coverage
- DEF-001: ORC Dependency Cycle Resolution
- DEF-007: Convention Compliance Sweep
- DESIGN-PIPELINE-001: Hybrid Design Pipeline
- HARNESS-001: Meta-Harness Skill
- PROJECT-HARNESS-001: Project Harness Activation
