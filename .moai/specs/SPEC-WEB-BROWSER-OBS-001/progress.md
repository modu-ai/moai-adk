# Progress — SPEC-WEB-BROWSER-OBS-001

카드 t1081 · 트리 WT-cdp-observation @ cd99336bf · 측정 전용 SPEC (tier M)

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit>_ — plan-phase 산출물 4종(spec.md / plan.md / acceptance.md / progress.md) 작성 완료. plan-auditor 실행 시 `plan_status: audit-ready` + `plan_complete_at` 기록 예정.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

_<pending orchestrator — run 단계 진입 전 기록>_

---

## Run-phase 참고 메모 (plan-phase 작성)

- MCP 서버 project root 는 spawn-frozen 이므로, run 중 `mcp__moai__*` 호출은 전부 `project_root` = 본 워크트리의 `git rev-parse --show-toplevel` 절대 경로를 명시 전달한다(spec.md HARD-7).
- 판정 기록 거처: `.moai/reports/t1081/verdict.md`. 프로브·캡처도 같은 디렉터(무추적).
- 레인 의무: 완료 보고에 카드 id · 브랜치와 HEAD · 로컬 병합 SHA · 증거 경로 — push 는 리드 일괄(CLAUDE.local.md §4.1).
