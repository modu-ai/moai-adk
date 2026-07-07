---
id: SPEC-INTERNAL-SECURITY-001
title: "internal/ 보안 하드닝 — 진행 상태"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P0
phase: "v3.0.0"
module: "internal/ (web, cli, hook, profile, template)"
lifecycle: spec-anchored
tags: "security, hardening, progress"
tier: M
---

## §E.1 Plan-phase Audit-Ready Signal

- plan-phase 산출물 4종(spec.md + plan.md + acceptance.md + progress.md) 작성 완료 (status: draft).
- 8개 verified finding → REQ-SEC-001..008 변환 완료, 3개 논리 그룹(web / template-update / hook)으로 그룹핑.
- 각 finding의 file:line 앵커 plan-phase 착수 직전 spot-check 재확인.
- 미결 설계 분기 3건(plan.md §E): REQ-SEC-006 배선 vs 제거, REQ-SEC-002 Sec-Fetch-Site vs 토큰, REQ-SEC-005 provenance vs 보수적 백업 — run-phase 진입 전 사용자 확인 권장(blocker 후보).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
