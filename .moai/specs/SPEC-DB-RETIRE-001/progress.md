---
id: SPEC-DB-RETIRE-001
title: "DB 문서화 서브시스템 전면 제거 — 진행 기록"
version: "0.1.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
---

# SPEC-DB-RETIRE-001 — Progress (§E)

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 생성 + plan-auditor D1-D5 반영 완료.
- 22 REQ ↔ 24 AC 매핑(커버리지 22/22). 코드 위치 콘텐츠-앵커 실측 완료(DeprecatedPaths·audit_loader·PhaseDB 추가).
- clarification RESOLVED(db.yaml 능동 삭제=REQ-DBR-022). 미해소 clarification 마커(콜론-정규형) **0건**.
- SPEC ID pre-write self-check PASS. Out of Scope 5항목 명시(MCP backend-db / deployment migration / settings testdata db.yaml / `.moai/project/db/` preserve-root / PhaseDB).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소유>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소유>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소유>_
