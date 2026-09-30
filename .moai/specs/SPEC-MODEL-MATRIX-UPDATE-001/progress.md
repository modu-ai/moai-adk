---
spec: SPEC-MODEL-MATRIX-UPDATE-001
tier: M
created: 2026-09-30
author: manager-spec
---

# progress.md — SPEC-MODEL-MATRIX-UPDATE-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-ID: SPEC-MODEL-MATRIX-UPDATE-001 (Bash regex 검증 PASS, 2026-09-30 — `[[ "SPEC-MODEL-MATRIX-UPDATE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS` → PASS; 기존 카탈로그에 동일 ID 부재, MODEL-MATRIX-* 선행 4종은 전부 superseded로 동명이족 아님)
- Tier: M — 근거: config(defaults/closed_sets/audit_models) + cli(mcp_claude/mcp_codex) + settings(schema_sections) + template(workflow.yaml/llm.yaml/settings.json.tmpl 불변) + web(i18n) + 룰 미러 + 테스트 8+파일. 파일 수 12-15, 마일스톤 5. 단일 도메인 결정(모델 핀 값)이 지배하는 Tier M.
- 아티팩트: spec.md / plan.md / acceptance.md / progress.md 4종 (Tier M 집합; research.md 불요 — 검증표가 plan.md §A에 귀속됨)
- 이슈 소지: [NEEDS CLARIFICATION] 3건(plan.md [NC-1][NC-2][NC-3]) — 모두 권장안 + 기본 진행 동반, 블로킹 아님. 리드 SPEC 검수 때 확정.
- 운영자 지시 대체 기록: REQ-AMP-005(Audit.Codex Go 기본 EMPTY 중립성)를 REQ-MMU-001로 대체 — spec.md §C.1.
- plan-audit iter-1: **FAIL 0.875** (`.moai/reports/t1368/plan-audit.md` — MP-7 미해결 마커 게이트 + D2-D6). Wave-1 반영 완료(2026-09-30): D2 AC 19→15 병합·재번호(커버리지 손실 없음), D3 보장-RED 테스트 2파일 AC 편입(10파일), D4 i18n 4 로케일 + 감사자 미발견 동급 effort 설명 표면(spec.md §B REQ-MMU-002에 직접 관측 근거 기록), D5 무효 주석 2표면, D6 llm.yaml 환상 앵커 제거(verify-only 재분류), D8 related_specs 산문 이관. **D1(NC 결정 기록)은 wave-2 — 운영자 응답 대기.**
- 검증 예산: go build + grep/sed만 (go test ./... 금지 — plan.md §D).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop가 마일스톤 착지 증거를 기록>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop가 종료 시 기입>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs가 sync 커밋 착지 시 기입>_
