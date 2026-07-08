---
id: SPEC-INTERNAL-PERF-001
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
---

# SPEC-INTERNAL-PERF-001 Progress

## §E.1 Plan-phase Audit-Ready Signal

plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 작성 완료, `status: draft`. plan-audit 및 Implementation Kickoff Approval 대기.

plan_complete_at: 2026-07-08
plan_status: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §J Plan-phase Debt

> plan-auditor Phase 0.5 audit MINOR 결함 3건(D2/D4/D5)의 run/sync-phase 가시화 기록. 본 섹션은 기록 전용(plan-phase debt registry)이며 본SPEC의 plan-phase 범위에서 수정하지 않는다. `§E` namespace는 era.go 파서가 lifecycle 구조로 예약하므로 별도 letter(§J)를 사용한다.

### D2 (MINOR) — REQ-PERF-001-A cache key에 query-kind 누락 가능성

- **현상**: REQ-PERF-001-A가 캐시 키를 `(BaseDir, SpecID)`로 명시하나 query-kind(질의 종류)는 생략. plan §F M1 방향과 §H R1 risk는 query-kind 포함을 요구.
- **영향**: REQ-level WHAT이 silent → 두 규칙이 서로 다른 git 질의 결과를 동일 키로 오염 공유할 수 있는 cross-rule cache-contamination 해석 간극.
- **bounding**: plan §H R1(캐시 키에 query-kind 포함 의무) + REQ-PERF-001-C(doc/code parity AC)가 run-phase 구현을 종속. REQ 본문 수정 없이도 run-phase에서 R1이 구현 방향을 보정.
- **선택적 run-phase 액션**(필수 아님): REQ-PERF-001-A 본문을 query-kind 포함으로 강화하거나, cross-rule isolation AC sub-criterion을 acceptance.md에 추가.
- **결정**: 사용자 판단으로 REQ-PERF-001-A 본문 미수정(plan-phase debt 기록만). run-phase에서 R1 + REQ-PERF-001-C가 기능적 격리를 보장.

### D4 (MINOR) — AC가 strict GEARS 대신 Given/When/Then test-scenario 형식

- **현상**: acceptance.md의 AC들이 strict GEARS pattern 대신 Given/When/Then test-scenario 형식으로 작성됨.
- **분류**: mislabel 아님 — acceptance.md 헤더가 이들을 sub-criteria로 명시. GEARS 무게는 spec.md §B의 REQ가 담당.
- **판정**: borderline PASS — action 불필요, audit trail 목적 기록.
- **결정**: plan-phase 수정 없음.

### D5 (MINOR) — plan.md §A.1 strategies.go line-number drift

- **현상**: plan.md §A.1이 `internal/merge/strategies.go` `computeLineChanges`를 L107-108로 인용하나 실제 def는 L222 (~115줄 offset).
- **분류**: content-token match는 유효(`computeLineChanges` 함수명 자체는 정확). plan §A.1 footnote가 line-drift asymmetry를 self-acknowledge함.
- **bounding**: run-phase M5 착수 시 content-token 앵커 우선 재검증 의무(plan §A.1 footnote + CLAUDE.local.md line-number drift 교훈). line-number는 보조 신호일 뿐.
- **결정**: plan-phase 수정 없음. run-phase에서 content-token 기준 재검증 예정.
