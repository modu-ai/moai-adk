---
id: SPEC-CODEMAPS-REFRESH-001
title: "progress.md"
version: "0.1.1"
created: 2026-09-02
updated: 2026-09-02
author: manager-spec
phase: "v3.2.0 target"
module: ".moai/project/codemaps"
tier: M
---

# progress.md — SPEC-CODEMAPS-REFRESH-001

## §E.1 Plan-phase Audit-Ready Signal

- Tier: M (근거: 다중 마일스톤 구조 + 정확성 검증의 독립 AC 열거 필요 + 스탬프 고아화 회피 절차 규율. LOC는 소규모이나 검증 깊이와 실패모드 면에서 S 상한 초과 → plan-audit skip 임계 0.80 적용)
- Artifact set: spec.md / plan.md / acceptance.md (Tier M 3종) + tracking용 progress.md
- GEARS 준수: REQ-CMR-001~008 (Ubiquitous / While / When / Where / Unwanted 패턴 사용, IF/THEN 없음)
- 기준선 실측(2026-09-02, 워크트리 t432 @ ad272be20): `moai graph check` codemaps value=60 threshold=40 verdict=stale (contribution 13 described-worthy files vs first parent f3e11e113) · mx-index/edges absent(신규 worktree 예상 상태) · `go list ./...` 137 · 팬텀 6 디렉터리 부재 · `graph stamp --commit` 플래그 존재(`internal/cli/graph_stamp.go:68`)
- Out of Scope: t304 소관 팬텀 수정 / 임계값 재보정 / Go 코드·신규 툴링 — `### Out of Scope` H3 3건으로 명시
- NEEDS CLARIFICATION 마커: 0건 (카드 의도·범위·기준선 모두 실측 확정)
- plan-audit iter-1: FAIL 0.81 (D1 BLOCKING) → 수리 완료(D1~D6 전부 적용, 2026-09-02) — iter-2 delta 재감사 대기

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
