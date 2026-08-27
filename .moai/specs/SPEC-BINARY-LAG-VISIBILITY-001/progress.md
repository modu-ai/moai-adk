---
id: SPEC-BINARY-LAG-VISIBILITY-001
title: "배포 지연 가시성 — 진행 기록"
version: "0.3.0"
status: draft
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: HIGH
phase: "v3.1.4 target"
module: internal/cli, internal/hook, build
lifecycle: spec-anchored
tags: "deployment-lag, doctor, session-start, version-stamp, observability, fail-open"
tier: M
---

# SPEC-BINARY-LAG-VISIBILITY-001 — 진행 기록

카드: t326 · 워크트리: `.claude/worktrees/t326` · 브랜치: `WT-integration-lock-identity`

## §E.1 Plan-phase Audit-Ready Signal

- Tier: **M** (영향 파일 6건 전수 열거 — `spec.md` §4). Tier S의 `< 5 files` 기준 위반으로 M 채택.
- 산출물: `spec.md` / `plan.md` / `acceptance.md` / `progress.md`, `status: draft`.
- 요구 9 (REQ-BLV-001..009) / 수락 9 (AC-BLV-001..009). 추적표는 `spec.md` §3.1 — 모든 REQ가 ≥1개 AC를 가지며 **엄밀한 1:1은 아니다**(REQ-BLV-002 → AC-BLV-001 + AC-BLV-002). 감사 지적(표현 부정확)을 반영해 「1:1」 표현을 정정했다.
- SPEC ID 정규식 자가 점검: 실행됨, 출력 `PASS`. `moai spec lint --strict spec.md` → `✓ No findings` (`--json` → `[]`).
- **전제 반증 1건**: 카드 지시문의 「Link 1 미커버」가 거짓임을 실측으로 확인하고 범위를 재구성했다(`spec.md` §1.4). 감사 시 이 절을 먼저 읽을 것.
- **선례 인용 정정 1건 (v0.2.0)**: v0.1.0이 `computeDeferredAdvisory`를 **발화** 선례로 인용했으나, 그 권고는 `HookOutput.Data`(`internal/hook/types.go:394`, `json:"-"`)로 들어가 도달하지 않는다. 선례를 「일정」과 「발화」로 분리하고 REQ/AC-BLV-008을 신설했다(`spec.md` §1.7.1). 이 정정을 놓친 구현은 결손을 그대로 재생산한다.
- **clarification 2건 모두 종결** (리드·운영자 결정): (a) 발화 표면 = `additionalContext` 단독, (b) `VERSION` 무변경 + 별도 `BUILD_ID`. `[NEEDS CLARIFICATION]` 잔여 0건.
- **잔여 위험 1건 명시**: `VERSION`이 그대로이므로 `moai version` 제목 줄은 `v3.1.2`로 남고 제목 줄만 읽는 오독 가능성이 남는다(`spec.md` §7.5). 제목 줄 축은 **별도 카드**로 이관 — 이 SPEC에서 넓히지 않는다.

### plan-audit iter-1 결과 (v0.3.0에서 반영 완료)

- 판정: **PASS-WITH-DEBT 0.80** (Tier M 임계 0.80 — 경계선). 보고서: `.moai/reports/t326/plan-audit-iter1.md`, 감사 트리 HEAD `25f7b0fe9`.
- must-pass 7건 전부 통과. 중심 분석(비교 존재 → 자동 호출 0건 → 운반 표면이 `json:"-"`)이 **독립 재측정으로 확인**됨.
- 결함 9건 중 **8건 수리 완료**(blocking 3 + optional 5, 리드가 optional까지 확대 지시). D9는 정보성으로 유지 — t317이 이 트리에 없다는 사실은 경로·스냅샷 일자와 함께 이미 공개돼 있다(`spec.md` §5 각주).
- **핀 판정**: AC-BLV-004의 `343399d2f` 핀은 감사관이 **독립적으로** 「올바른 규율이며 staleness가 아니다」로 판정(보고서 42-43행, 근거 `git merge-base --is-ancestor 343399d2f 25f7b0fe9` → true). 리드 판정과 일치하며 **핀은 그대로 유지**한다. 502로 갱신하지 않는다.
- 감사 미관측 항목(Gaps)은 보고서에 기록돼 있으며, 그중 「비-git 트리 실물 `moai doctor`」는 `plan.md` M4의 좁힌 형태 실측이 닫는다.

_iter-2 재감사 또는 Implementation Kickoff Approval 대기_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
