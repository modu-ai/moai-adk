# SPEC-AGENT-MODEL-ENFORCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-08-08
tier: M
artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M 표준 — 편차 없음)
issue: #1376
open_clarifications: 2 (plan.md §B D4 스텁 배치 위치, D5 감사 로그 형식)
blocking_gate: M1 페이로드 실측 — PreToolUse가 Agent/Task에 발화하는지 미검증

## §E.2 Run-phase Evidence

### M1 — 페이로드 실측 결과 (REQ-AME-002 3문항)

측정 방법: 격리된 임시 프로젝트(`scratchpad/m1probe`)에 PreToolUse `Agent|Task` matcher 블록 +
stdin 덤프 훅을 배선하고, `claude -p`(Claude Code 2.1.226) 헤드리스 세션에서 실제 Agent spawn을
발생시켜 페이로드를 포획. `Bash` matcher 블록을 **양성 대조군**으로 함께 배선해
"훅 미발화"와 "프로브 고장"을 구분함.

| # | 질문 | 관측된 답 | 근거 |
|---|------|-----------|------|
| (a) | **PreToolUse가 Agent/Task에 발화하는가** | **YES** | 포획 레코드 `hook_event_name: "PreToolUse"`, `tool_name: "Agent"`. 양성 대조군(Bash) 3건도 동시 포획되어 프로브 정상 동작 확인 |
| (b) | **`tool_input`이 `subagent_type`을 담는가** | **YES** | `tool_input.subagent_type == "Explore"`. 픽스처: `internal/hook/testdata/agent_pretool_payload.json` |
| (c) | **`tool_input`이 `model` 키를 담을 수 있는가** | **YES** | 2차 프로브에서 model 인자를 명시한 spawn 유도 → `tool_input.model == "haiku"` 실제 포획. 픽스처: `internal/hook/testdata/agent_pretool_payload_with_model.json` |

판정: **M1 게이트 통과**. REQ-AME-003의 재라우팅 분기(PreToolUse 미발화 시 M2 이후 중단)는
발동하지 않음. M2-M4는 관측된 능력 위에 구축됨.

관측된 `tool_input` 키 집합 (model 부재 spawn): `{description, prompt, run_in_background, subagent_type}`
관측된 `tool_input` 키 집합 (model 보유 spawn): `{description, model, prompt, run_in_background, subagent_type}`

부수 관측(범위 밖, 기록만): PreToolUse 페이로드는 최상위에 `effort` 객체(`{"level":"medium"}`)를
담으며, 서브에이전트 내부에서 발생한 도구 호출은 최상위 `agent_id` / `agent_type`을 추가로 담는다.
effort는 §A.4에 따라 본 SPEC의 집행 대상이 아니며, 이 관측은 후속 SPEC의 입력 자료로만 남긴다.

### F4 독립 교차 검증 (PostToolUse Agent 분기 도달 불가)

`.moai/logs/task-metrics.jsonl` 3409행 전수 스캔 결과 UUID 형태 session_id **0건**
(전량 `sess-metrics-*` 등 테스트 픽스처 id). 즉 `logTaskMetrics`는 실세션에서 한 번도
기록한 적이 없으며, spec.md §A.1 F4의 "배선상 도달 불가" 판정이 독립 증거로 재확인됨.

### 순환 임포트 사전 점검 (plan.md §C 6번 — 차단 조건)

```
go list -deps ./internal/template/... | grep 'moai-adk/internal/hook'   → 출력 없음
```
역방향 의존 **부재 확인**. `internal/hook → internal/template` 직접 임포트가 안전하므로
REQ-AME-012의 해석기 직접 호출을 그대로 채택(제3 패키지 승격/인터페이스 주입 불필요).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
