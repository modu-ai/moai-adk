---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Progress"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/rules/moai/core, internal/template/templates"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: `spec.md` / `plan.md` / `acceptance.md` / `research.md` / `progress.md` 생성 완료 (초기 `status: draft`).
- SPEC ID 사전 자가검사: `decomposition: SPEC ✓ | AGENT ✓ | PARALLEL ✓ | OPT ✓ | 001 ✓ → PASS` (실행 Bash 출력 `PASS`).
- Ground truth: `spec.md` §F 전량 본 세션 실측. 브리프 대비 정정 2건은 `research.md` §E, 배포 전제 정정 3건은 `spec.md` §F.8 / `research.md` §H 기록.
- Tier: L (`spec.md` §E 근거).
- 규모: 54 REQ (Group 1~6) / 55 AC.

### 결정 사항 (v0.2.0 — 전량 RESOLVED)

| 결정 | 값 | 반영 위치 |
|---|---|---|
| D1 `.js` 배포 | **템플릿 미러(배포) 채택** | Group 6 REQ-APO-069..073, `plan.md` M2, frontmatter `partially_supersedes` |
| D2 sync P12 형태 | **read-only drafter + 단일 적용자 확정**, disjoint-writer 불채택 | REQ-APO-024 / 024b, `plan.md` M3, `spec.md` §C |
| D3 SPEC-ID 마커 | **선행 grep 게이트 후 결정** (지금 결정하지 않음) | `plan.md` §B.3 + M4 작업 0, AC-APO-043 |

미해소 clarification 마커 **0건**. Implementation Kickoff Approval 진행 가능.

### 미결 관측 항목 (run-phase에서 기록)

- **D3 게이트 출력** — M4 착수 시 `grep -rn "decomposition\|segment match trace" internal/ .github/ .claude/` 를 실행하고 verbatim 출력을 §E.2에 기록해야 한다. 현재 미실행이며, 실행 전에는 마커 제거 여부를 단정하지 않는다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
