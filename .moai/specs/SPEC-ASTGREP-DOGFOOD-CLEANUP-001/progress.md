---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup — Progress"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P3
phase: "maintainer-tooling"
module: ".moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, dogfood, cleanup, tooling, local"
---

# 진행 기록 — SPEC-ASTGREP-DOGFOOD-CLEANUP-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물 4종 저작 완료 (spec.md + plan.md + acceptance.md + progress.md), 전부 `status: draft`.
- 실측 앵커 재검증 완료 (2026-07-08, manager-spec): 로컬 트리 41 tracked + 1 untracked, sg 0.40.5 설치,
  §2.2 "메시지 혼재 ko/en" 실측 교정(데모 stub=영어 / go·security=한국어), go·security 오늘 수정 +
  credentials.yml untracked in-flight 레이스 신호.
- Tier: **M (standard)**. 스코프 결정: **curated-baseline alignment** (16-lang 전면 저작 아님).
- GEARS 요구사항 10개(REQ-ADC-001..010), AC 12개(AC-ADC-001..010, sub-ID 003a/003b/009a/009b), Given-When-Then 3.
- Out of Scope 4개 H3 sub-heading + `-` bullet 확보 (OutOfScopeRule 충족).
- 커밋 미수행 (오케스트레이터 소관 — hyper-active shared checkout).

## §E.2 Run-phase Evidence

_<pending run-phase — manager-develop 소관>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
