---
id: SPEC-CODEX-RESUME-SCOPE-001
document: progress
card: t1201
---

# Progress — SPEC-CODEX-RESUME-SCOPE-001

## §E.1 Plan-phase Audit-Ready Signal

- 산출물: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- 전제 측정: fixture transport 탐침(`.moai/reports/t1201/premise-probe.txt`, 소스 `premise-probe_test.go.txt`) — `resume_last` 가 카드 B의 `thr-6` 을 재개. MCP 서버 해석 경로(pid 82350: cwd `.claude/worktrees/t1169`, `CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go`) — 워크트리들이 레지스트리 하나를 공유.
- 운영자 결정 3건(`plan.md` §B): 혼합안, `thread_id` 는 이 프로젝트 레지스트리 기록분만, 거부는 구조화 JSON 만(`IsError` 없음). 모두 잠정 기본값이며 Implementation Kickoff 에서 운영자가 확정한다.
- fixture 응답 id 측정(`.moai/reports/t1201/fixture-probe.txt`): 기본 fixture 는 `turn/start` 를 `tid-fake` 로, 에코 변형은 요청 id 로 보낸다.
- plan-audit: iter-1 FAIL 0.75(`.moai/reports/t1201/plan-audit.md`) → v0.2.0 에서 D1~D8 수리, 선택 항목 D9~D13 반영. iter-2 재감사 대기.
- 규모: REQ 10개, AC 15개.
- LIVE 호출: plan 단계 0회.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
